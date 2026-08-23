package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AsaiYusuke/jsonpath/v2"
	e "github.com/michaelmccabe/ramjam/pkg/errors"
	"gopkg.in/yaml.v3"
)

type (
	InstructionsFile struct {
		Metadata struct {
			Name        string `yaml:"name"`
			Author      string `yaml:"author"`
			Description string `yaml:"description"`
		} `yaml:"metadata"`
		Config struct {
			BaseURL string `yaml:"base_url"`
		} `yaml:"config"`
		Workflow []Step `yaml:"workflow"`
	}

	Step struct {
		Step        string      `yaml:"step"`
		Description string      `yaml:"description"`
		Request     StepRequest `yaml:"request"`
		Expect      StepExpect  `yaml:"expect"`
		Capture     []Capture   `yaml:"capture"`
		Output      Output      `yaml:"output"`
	}

	StepRequest struct {
		Method      string                 `yaml:"method"`
		URL         string                 `yaml:"url"`
		ContentType string                 `yaml:"content_type,omitempty"`
		Headers     map[string]string      `yaml:"headers"`
		Body        map[string]interface{} `yaml:"body,omitempty"`
		BodyFile    string                 `yaml:"body_file,omitempty"`
		Params      map[string]string      `yaml:"params"`
		bodyData    map[string]interface{} // resolved body data
		bodySource  string                 // tracks source for debugging
	}

	StepExpect struct {
		Status        int                 `yaml:"status"`
		JSONPathMatch []JSONPathVal       `yaml:"json_path_match"`
		Headers       []HeaderExpectation `yaml:"headers"`
	}

	JSONPathVal struct {
		Path     string      `yaml:"path"`
		Operator string      `yaml:"operator,omitempty"`
		Value    interface{} `yaml:"value"`
	}

	HeaderExpectation struct {
		Name     string `yaml:"name"`
		Value    string `yaml:"value,omitempty"`
		Contains string `yaml:"contains,omitempty"`
	}

	Capture struct {
		JSONPath string `yaml:"json_path,omitempty"`
		Header   string `yaml:"header,omitempty"`
		Regex    string `yaml:"regex,omitempty"`
		As       string `yaml:"as"`
	}

	Output struct {
		Print string `yaml:"print"`
	}

	StepError struct {
		File        string
		Step        string
		Description string
		Err         error
	}
)

func (e *StepError) Error() string {
	return fmt.Sprintf("step %q in %s failed: %v", e.Step, e.File, e.Err)
}

func (e *StepError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	Step string
	Err  error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type NetworkError struct {
	URL string
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network request failed for %s: %v", e.URL, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

type ParsingError struct {
	Target string
	Err    error
}

func (e *ParsingError) Error() string {
	return fmt.Sprintf("parse error on %s: %v", e.Target, e.Err)
}

func (e *ParsingError) Unwrap() error {
	return e.Err
}

type ResolutionError struct {
	Target string
	Err    error
}

func (e *ResolutionError) Error() string {
	return fmt.Sprintf("resolution error on %s: %v", e.Target, e.Err)
}

func (e *ResolutionError) Unwrap() error {
	return e.Err
}

type GlobalConfig struct {
	Defaults struct {
		BaseURL string            `yaml:"base_url"`
		Timeout string            `yaml:"timeout"`
		Headers map[string]string `yaml:"headers"`
	} `yaml:"defaults"`
}

type Runner struct {
	client       *http.Client
	verbose      bool
	logger       *slog.Logger
	cliVars      map[string]string
	globalConfig GlobalConfig
}

func New(timeout time.Duration, verbose bool) *Runner {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	r := &Runner{
		client:  &http.Client{Timeout: timeout},
		verbose: verbose,
		logger:  logger,
	}
	r.LoadGlobalConfig()
	return r
}

func (r *Runner) LoadGlobalConfig() {
	var data []byte
	var err error

	if _, errStat := os.Stat(".ramjam.yaml"); errStat == nil {
		data, err = os.ReadFile(".ramjam.yaml")
	} else if _, errStat := os.Stat(".ramjam.yml"); errStat == nil {
		data, err = os.ReadFile(".ramjam.yml")
	} else {
		home, errHome := os.UserHomeDir()
		if errHome == nil {
			homeConfig := filepath.Join(home, ".ramjam.yaml")
			if _, errStat := os.Stat(homeConfig); errStat == nil {
				data, err = os.ReadFile(homeConfig)
			} else {
				homeConfigYml := filepath.Join(home, ".ramjam.yml")
				if _, errStat := os.Stat(homeConfigYml); errStat == nil {
					data, err = os.ReadFile(homeConfigYml)
				}
			}
		}
	}

	if err != nil || len(data) == 0 {
		return
	}

	var config GlobalConfig
	if errUnmarshal := yaml.Unmarshal(data, &config); errUnmarshal == nil {
		r.globalConfig = config
		if config.Defaults.Timeout != "" {
			d, errDuration := time.ParseDuration(config.Defaults.Timeout)
			if errDuration == nil {
				r.client.Timeout = d
			}
		}
	}
}

func (r *Runner) SetVars(vars map[string]string) {
	r.cliVars = vars
}

func (r *Runner) RunPaths(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided")
	}

	var files []string
	for _, p := range paths {
		fs, err := r.collectFiles(p)
		if err != nil {
			return err
		}
		files = append(files, fs...)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, f := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			fileErrs := r.runFile(f)
			if len(fileErrs) > 0 {
				mu.Lock()
				errs = append(errs, fileErrs...)
				mu.Unlock()
			}
		}(f)
	}

	wg.Wait()

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (r *Runner) collectFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err := e.Wrapf(err, "unable to access %s", path); err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err := e.Wrapf(err, "unable to read dir %s", path); err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			files = append(files, filepath.Join(path, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func (r *Runner) runFile(path string) []error {
	prefix := filepath.Base(path)

	data, err := os.ReadFile(path)
	if err != nil {
		r.logger.Error("Failed to read file", "file", path, "error", err)
		return []error{&ParsingError{Target: path, Err: err}}
	}
	var spec InstructionsFile
	if err := yaml.Unmarshal(data, &spec); err != nil {
		r.logger.Error("Failed to parse YAML", "file", path, "error", err)
		return []error{&ParsingError{Target: path, Err: err}}
	}

	if spec.Metadata.Name != "" {
		prefix = spec.Metadata.Name
	}

	logger := r.logger.With("workflow", prefix)
	logger.Info("Running workflow file", "path", path)

	baseURL := spec.Config.BaseURL
	if baseURL == "" && r.globalConfig.Defaults.BaseURL != "" {
		baseURL = r.globalConfig.Defaults.BaseURL
	}
	vars := map[string]string{
		"base_url": baseURL,
		"_dir":     filepath.Dir(path),
	}
	for k, v := range r.cliVars {
		vars[k] = v
	}

	// Resolve body files relative to the YAML file's directory
	baseDir := filepath.Dir(path)

	var errs []error
	for _, step := range spec.Workflow {
		// Resolve body from file if specified
		if err := r.resolveBodyFile(&step, baseDir); err != nil {
			errs = append(errs, &StepError{
				File:        path,
				Step:        step.Step,
				Description: step.Description,
				Err:         &ResolutionError{Target: step.Request.BodyFile, Err: err},
			})
			continue
		}

		if err := r.executeStep(step, vars, logger); err != nil {
			errs = append(errs, &StepError{
				File:        path,
				Step:        step.Step,
				Description: step.Description,
				Err:         err,
			})
		}
	}

	return errs
}

func (r *Runner) resolveBodyFile(step *Step, baseDir string) error {
	// If no body_file specified, use inline body
	if step.Request.BodyFile == "" {
		if len(step.Request.Body) > 0 {
			step.Request.bodyData = step.Request.Body
			step.Request.bodySource = "inline"
		}
		return nil
	}

	// Resolve the file path relative to the YAML file
	bodyPath := step.Request.BodyFile
	if !filepath.IsAbs(bodyPath) {
		bodyPath = filepath.Join(baseDir, bodyPath)
	}

	// Read the JSON file
	data, err := os.ReadFile(bodyPath)
	if err := e.Wrapf(err, "read body file %s", step.Request.BodyFile); err != nil {
		return err
	}

	// Parse the JSON
	var bodyData map[string]interface{}
	if err := e.Wrapf(json.Unmarshal(data, &bodyData), "parse body file %s", step.Request.BodyFile); err != nil {
		return err
	}

	step.Request.bodyData = bodyData
	step.Request.bodySource = step.Request.BodyFile
	return nil
}

func (r *Runner) executeStep(step Step, vars map[string]string, log *slog.Logger) error {
	log.Debug("Executing step", "step", step.Step)

	method := strings.ToUpper(strings.TrimSpace(step.Request.Method))
	if method == "" {
		method = http.MethodGet
	}

	requestURL := applyVars(step.Request.URL, vars)
	if len(step.Request.Params) > 0 {
		if idx := strings.Index(requestURL, "?"); idx >= 0 {
			requestURL = requestURL[:idx]
		}
	}

	reqURL := requestURL
	if !strings.HasPrefix(reqURL, "http") && vars["base_url"] != "" {
		reqURL = strings.TrimSuffix(vars["base_url"], "/") + "/" + strings.TrimPrefix(reqURL, "/")
	}

	contentType := strings.TrimSpace(step.Request.ContentType)
	if contentType == "" {
		contentType = "application/json"
	}
	actualContentType := contentType

	bodyReader := io.Reader(nil)
	if len(step.Request.bodyData) > 0 {
		body := applyVarsToInterface(step.Request.bodyData, vars)
		bodyMap, isMap := body.(map[string]interface{})

		if isMap && strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			formValues := url.Values{}
			for k, v := range bodyMap {
				formValues.Set(k, fmt.Sprint(v))
			}
			bodyReader = strings.NewReader(formValues.Encode())
		} else if isMap && strings.HasPrefix(contentType, "multipart/form-data") {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			for k, v := range bodyMap {
				valStr := fmt.Sprint(v)
				if strings.HasPrefix(valStr, "@") {
					filePath := strings.TrimPrefix(valStr, "@")
					if !filepath.IsAbs(filePath) {
						filePath = filepath.Join(vars["_dir"], filePath)
					}
					file, err := os.Open(filePath)
					if err == nil {
						part, errPart := writer.CreateFormFile(k, filepath.Base(filePath))
						if errPart == nil {
							io.Copy(part, file)
						}
						file.Close()
						continue
					}
				}
				writer.WriteField(k, valStr)
			}
			writer.Close()
			bodyReader = &buf
			actualContentType = writer.FormDataContentType()
		} else {
			payload, err := json.Marshal(body)
			if err != nil {
				return &ParsingError{Target: "request body", Err: err}
			}
			bodyReader = bytes.NewReader(payload)
		}
		log.Debug("Using body from", "source", step.Request.bodySource)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return &ResolutionError{Target: "request URL", Err: err}
	}
	req.Header.Set("User-Agent", "ramjam-cli")
	if bodyReader != nil {
		req.Header.Set("Content-Type", actualContentType)
	}

	for k, v := range r.globalConfig.Defaults.Headers {
		req.Header.Set(k, applyVars(v, vars))
	}

	for k, v := range step.Request.Headers {
		req.Header.Set(k, applyVars(v, vars))
	}

	if len(step.Request.Params) > 0 {
		query := req.URL.Query()
		for key, value := range step.Request.Params {
			query.Set(key, applyVars(value, vars))
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return &NetworkError{URL: reqURL, Err: err}
	}
	defer resp.Body.Close()

	log.Debug("Received status", "status", resp.StatusCode)

	if step.Expect.Status != 0 && resp.StatusCode != step.Expect.Status {
		return &ValidationError{
			Step: step.Step,
			Err:  fmt.Errorf("expected status %d, got %d", step.Expect.Status, resp.StatusCode),
		}
	}

	for _, headerExpect := range step.Expect.Headers {
		name := strings.TrimSpace(headerExpect.Name)
		if name == "" {
			return &ValidationError{
				Step: step.Step,
				Err:  fmt.Errorf("header expectation must specify a name"),
			}
		}
		if headerExpect.Value == "" && headerExpect.Contains == "" {
			return &ValidationError{
				Step: step.Step,
				Err:  fmt.Errorf("header expectation for %s must specify value or contains", name),
			}
		}
		actual := resp.Header.Get(name)
		if headerExpect.Value != "" {
			expected := applyVars(headerExpect.Value, vars)
			log.Debug("Asserting header equals", "header", name, "expected", expected)
			if actual != expected {
				return &ValidationError{
					Step: step.Step,
					Err:  fmt.Errorf("expected header %s to equal %q, got %q", name, expected, actual),
				}
			}
		}
		if headerExpect.Contains != "" {
			expected := applyVars(headerExpect.Contains, vars)
			log.Debug("Asserting header contains", "header", name, "contains", expected)
			if !strings.Contains(actual, expected) {
				return &ValidationError{
					Step: step.Step,
					Err:  fmt.Errorf("expected header %s to contain %q, got %q", name, expected, actual),
				}
			}
		}
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ParsingError{Target: "response body read", Err: err}
	}

	var jsonObj interface{}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &jsonObj); err != nil {
			return &ParsingError{Target: "response body JSON", Err: err}
		}
	}

	for _, matcher := range step.Expect.JSONPathMatch {
		actual, err := evalJSONPath(jsonObj, matcher.Path)
		if err != nil {
			return &ValidationError{
				Step: step.Step,
				Err:  fmt.Errorf("jsonpath %s evaluation failed: %w", matcher.Path, err),
			}
		}

		op := strings.ToLower(strings.TrimSpace(matcher.Operator))
		if op == "" {
			op = "eq"
		}

		expected := applyVars(fmt.Sprint(matcher.Value), vars)
		log.Debug("Asserting JSONPath", "path", matcher.Path, "op", op, "expected", expected)

		passed, err := evaluateMatch(actual, op, expected)
		if err != nil {
			return &ValidationError{
				Step: step.Step,
				Err:  fmt.Errorf("jsonpath %s assertion failed: %w", matcher.Path, err),
			}
		}

		if !passed {
			if op == "eq" {
				return &ValidationError{
					Step: step.Step,
					Err:  fmt.Errorf("jsonpath %s expected %q, got %q", matcher.Path, expected, actual),
				}
			}
			return &ValidationError{
				Step: step.Step,
				Err:  fmt.Errorf("jsonpath %s expected operator %q value %q, got %q", matcher.Path, op, expected, actual),
			}
		}
	}

	for _, cap := range step.Capture {
		var val interface{}
		var err error

		if cap.JSONPath != "" {
			val, err = evalJSONPath(jsonObj, cap.JSONPath)
			if err != nil {
				return &ValidationError{
					Step: step.Step,
					Err:  fmt.Errorf("capture json_path %s: %w", cap.JSONPath, err),
				}
			}
		} else if cap.Header != "" {
			headerVal := resp.Header.Get(cap.Header)
			if cap.Regex != "" {
				re, err := regexp.Compile(cap.Regex)
				if err != nil {
					return &ResolutionError{Target: "capture regex", Err: err}
				}
				matches := re.FindStringSubmatch(headerVal)
				if len(matches) > 1 {
					val = matches[1]
				} else if len(matches) > 0 {
					val = matches[0]
				} else {
					return &ValidationError{
						Step: step.Step,
						Err:  fmt.Errorf("regex %s did not match header %s value %q", cap.Regex, cap.Header, headerVal),
					}
				}
			} else {
				val = headerVal
			}
		} else {
			return &ResolutionError{
				Target: "capture rule",
				Err:    fmt.Errorf("capture must specify json_path or header"),
			}
		}

		log.Debug("Captured variable", "name", cap.As, "value", val)
		vars[cap.As] = fmt.Sprint(val)
	}

	if step.Output.Print != "" {
		msg := applyVars(step.Output.Print, vars)
		log.Info(msg)
	}

	return nil
}

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func applyVars(input string, vars map[string]string) string {
	return varPattern.ReplaceAllStringFunc(input, func(m string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		if v, ok := vars[key]; ok {
			return v
		}
		return m
	})
}

func applyVarsToInterface(val interface{}, vars map[string]string) interface{} {
	switch v := val.(type) {
	case string:
		return applyVars(v, vars)
	case []interface{}:
		for i := range v {
			v[i] = applyVarsToInterface(v[i], vars)
		}
		return v
	case map[string]interface{}:
		for k := range v {
			v[k] = applyVarsToInterface(v[k], vars)
		}
		return v
	default:
		return v
	}
}

func evalJSONPath(obj interface{}, path string) (interface{}, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return nil, fmt.Errorf("empty path")
	}

	// Normalize path: prepend $. if missing
	if !strings.HasPrefix(p, "$") {
		if strings.HasPrefix(p, "[") {
			p = "$" + p
		} else {
			p = "$." + p
		}
	}

	// Retrieve values using github.com/AsaiYusuke/jsonpath/v2
	res, err := jsonpath.Retrieve(p, obj)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("no match found for path %q", path)
	}

	// For compatibility with single value assertions, if there is exactly one match, 
	// return that single element. Otherwise return the whole slice.
	if len(res) == 1 {
		return res[0], nil
	}

	return res, nil
}

func evaluateMatch(actual interface{}, operator string, expectedStr string) (bool, error) {
	actualStr := fmt.Sprint(actual)

	switch operator {
	case "eq":
		return actualStr == expectedStr, nil
	case "ne":
		return actualStr != expectedStr, nil
	case "contains":
		switch act := actual.(type) {
		case []interface{}:
			for _, item := range act {
				if fmt.Sprint(item) == expectedStr {
					return true, nil
				}
			}
			return false, nil
		case []string:
			for _, item := range act {
				if item == expectedStr {
					return true, nil
				}
			}
			return false, nil
		default:
			return strings.Contains(actualStr, expectedStr), nil
		}
	case "gt", "gte", "lt", "lte":
		actualFloat, errAct := strconv.ParseFloat(actualStr, 64)
		expectedFloat, errExp := strconv.ParseFloat(expectedStr, 64)
		if errAct != nil || errExp != nil {
			switch operator {
			case "gt":
				return actualStr > expectedStr, nil
			case "gte":
				return actualStr >= expectedStr, nil
			case "lt":
				return actualStr < expectedStr, nil
			case "lte":
				return actualStr <= expectedStr, nil
			}
		}
		switch operator {
		case "gt":
			return actualFloat > expectedFloat, nil
		case "gte":
			return actualFloat >= expectedFloat, nil
		case "lt":
			return actualFloat < expectedFloat, nil
		case "lte":
			return actualFloat <= expectedFloat, nil
		}
	}

	return false, fmt.Errorf("unsupported operator %q", operator)
}

