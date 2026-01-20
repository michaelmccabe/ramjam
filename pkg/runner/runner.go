package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
		Method     string                 `yaml:"method"`
		URL        string                 `yaml:"url"`
		Headers    map[string]string      `yaml:"headers"`
		Body       map[string]interface{} `yaml:"body,omitempty"`
		BodyFile   string                 `yaml:"body_file,omitempty"`
		Params     map[string]string      `yaml:"params"`
		bodyData   map[string]interface{} // resolved body data
		bodySource string                 // tracks source for debugging
	}

	StepExpect struct {
		Status        int                 `yaml:"status"`
		JSONPathMatch []JSONPathVal       `yaml:"json_path_match"`
		Headers       []HeaderExpectation `yaml:"headers"`
	}

	JSONPathVal struct {
		Path  string      `yaml:"path"`
		Value interface{} `yaml:"value"`
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

type Runner struct {
	client  *http.Client
	verbose bool
}

func New(timeout time.Duration, verbose bool) *Runner {
	return &Runner{
		client:  &http.Client{Timeout: timeout},
		verbose: verbose,
	}
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
	type result struct {
		logs []string
		errs []error
	}
	results := make(chan result, len(files))

	for _, f := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			logs, errs := r.runFile(f)
			results <- result{logs: logs, errs: errs}
		}(f)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []error
	for res := range results {
		for _, l := range res.logs {
			fmt.Println(l)
		}
		if len(res.errs) > 0 {
			errs = append(errs, res.errs...)
		}
	}

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

func (r *Runner) runFile(path string) ([]string, []error) {
	var logs []string
	prefix := filepath.Base(path)
	log := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs = append(logs, fmt.Sprintf("[%s] %s", prefix, msg))
	}

	log("Running workflow file: %s", path)

	data, err := os.ReadFile(path)
	if err := e.Wrapf(err, "read %s", path); err != nil {
		return logs, []error{err}
	}
	var spec InstructionsFile
	if err := e.Wrapf(yaml.Unmarshal(data, &spec), "parse %s", path); err != nil {
		return logs, []error{err}
	}

	if spec.Metadata.Name != "" {
		prefix = spec.Metadata.Name
	}

	vars := map[string]string{
		"base_url": spec.Config.BaseURL,
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
				Err:         fmt.Errorf("resolve body file: %w", err),
			})
			continue
		}

		if err := r.executeStep(step, vars, log); err != nil {
			errs = append(errs, &StepError{
				File:        path,
				Step:        step.Step,
				Description: step.Description,
				Err:         err,
			})
		}
	}

	return logs, errs
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

func (r *Runner) executeStep(step Step, vars map[string]string, log func(string, ...interface{})) error {
	if r.verbose {
		log("Executing step: %s", step.Step)
	}

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

	url := requestURL
	if !strings.HasPrefix(url, "http") && vars["base_url"] != "" {
		url = strings.TrimSuffix(vars["base_url"], "/") + "/" + strings.TrimPrefix(url, "/")
	}

	bodyReader := io.Reader(nil)
	if len(step.Request.bodyData) > 0 {
		body := applyVarsToInterface(step.Request.bodyData, vars)
		payload, err := json.Marshal(body)
		if err := e.Wrap(err, "marshal body"); err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
		if r.verbose && step.Request.bodySource != "" {
			log("Using body from: %s", step.Request.bodySource)
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err := e.Wrap(err, "build request"); err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ramjam-cli")
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
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
	if err := e.Wrap(err, "request"); err != nil {
		return err
	}
	defer resp.Body.Close()

	if r.verbose {
		log("Received status: %d", resp.StatusCode)
	}

	if step.Expect.Status != 0 && resp.StatusCode != step.Expect.Status {
		return fmt.Errorf("expected status %d, got %d", step.Expect.Status, resp.StatusCode)
	}

	for _, headerExpect := range step.Expect.Headers {
		name := strings.TrimSpace(headerExpect.Name)
		if name == "" {
			return fmt.Errorf("header expectation must specify a name")
		}
		if headerExpect.Value == "" && headerExpect.Contains == "" {
			return fmt.Errorf("header expectation for %s must specify value or contains", name)
		}
		actual := resp.Header.Get(name)
		if headerExpect.Value != "" {
			expected := applyVars(headerExpect.Value, vars)
			if r.verbose {
				log("Asserting header %s == %s", name, expected)
			}
			if actual != expected {
				return fmt.Errorf("expected header %s to equal %q, got %q", name, expected, actual)
			}
		}
		if headerExpect.Contains != "" {
			expected := applyVars(headerExpect.Contains, vars)
			if r.verbose {
				log("Asserting header %s contains %s", name, expected)
			}
			if !strings.Contains(actual, expected) {
				return fmt.Errorf("expected header %s to contain %q, got %q", name, expected, actual)
			}
		}
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err := e.Wrap(err, "read body"); err != nil {
		return err
	}

	var jsonObj interface{}
	if len(rawBody) > 0 {
		if err := e.Wrap(json.Unmarshal(rawBody, &jsonObj), "parse response json"); err != nil {
			return err
		}
	}

	for _, matcher := range step.Expect.JSONPathMatch {
		actual, err := evalJSONPath(jsonObj, matcher.Path)
		if err := e.Wrapf(err, "jsonpath %s", matcher.Path); err != nil {
			return err
		}
		expected := applyVars(fmt.Sprint(matcher.Value), vars)
		if r.verbose {
			log("Asserting %s == %s", matcher.Path, expected)
		}
		if fmt.Sprint(actual) != expected {
			return fmt.Errorf("jsonpath %s expected %q, got %q", matcher.Path, expected, actual)
		}
	}

	for _, cap := range step.Capture {
		var val interface{}
		var err error

		if cap.JSONPath != "" {
			val, err = evalJSONPath(jsonObj, cap.JSONPath)
			if err := e.Wrapf(err, "capture json_path %s", cap.JSONPath); err != nil {
				return err
			}
		} else if cap.Header != "" {
			headerVal := resp.Header.Get(cap.Header)
			if cap.Regex != "" {
				re, err := regexp.Compile(cap.Regex)
				if err := e.Wrapf(err, "invalid regex %s", cap.Regex); err != nil {
					return err
				}
				matches := re.FindStringSubmatch(headerVal)
				if len(matches) > 1 {
					val = matches[1]
				} else if len(matches) > 0 {
					val = matches[0]
				} else {
					return fmt.Errorf("regex %s did not match header %s value %q", cap.Regex, cap.Header, headerVal)
				}
			} else {
				val = headerVal
			}
		} else {
			return fmt.Errorf("capture must specify json_path or header")
		}

		if r.verbose {
			log("Captured %s => %s", cap.As, fmt.Sprint(val))
		}
		vars[cap.As] = fmt.Sprint(val)
	}

	if step.Output.Print != "" {
		msg := applyVars(step.Output.Print, vars)
		log("%s", msg)
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

	// Handle filter of form $[?(@.field==value)].rest (value may be quoted or bare)
	if m := regexp.MustCompile(`^\$\[\?\(@\.([A-Za-z0-9_\-]+)==['"]?([^'"]+)['"]?\)\](?:\.(.*))?$`).FindStringSubmatch(p); m != nil {
		field, val, rest := m[1], m[2], m[3]
		arr, ok := obj.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array for filter %s", path)
		}
		var matches []interface{}
		for _, el := range arr {
			if mp, ok := el.(map[string]interface{}); ok {
				if fmt.Sprint(mp[field]) == val {
					matches = append(matches, el)
				}
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no match for filter %s", path)
		}
		selected := matches[0]
		if rest != "" {
			return evalJSONPath(selected, rest)
		}
		return matches, nil
	}

	// Handle index of form $[0].rest
	if m := regexp.MustCompile(`^\$\[([0-9]+)\](?:\.(.*))?$`).FindStringSubmatch(p); m != nil {
		idx, _ := strconv.Atoi(m[1])
		arr, ok := obj.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array for index %s", path)
		}
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("index out of range for %s", path)
		}
		selected := arr[idx]
		if rest := m[2]; rest != "" {
			return evalJSONPath(selected, rest)
		}
		return selected, nil
	}

	// Trim leading $ or $.
	p = strings.TrimPrefix(strings.TrimPrefix(p, "$."), "$")
	segments := strings.Split(p, ".")
	cur := obj
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		name := seg
		idx := -1
		if strings.Contains(seg, "[") && strings.HasSuffix(seg, "]") {
			parts := strings.SplitN(seg, "[", 2)
			name = parts[0]
			idStr := strings.TrimSuffix(parts[1], "]")
			if idStr != "" {
				parsed, err := strconv.Atoi(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid index in segment %s", seg)
				}
				idx = parsed
			}
		}
		if name != "" {
			m, ok := cur.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object for segment %s", name)
			}
			cur = m[name]
		}
		if idx >= 0 {
			arr, ok := cur.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array for segment %s", seg)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("index out of range for segment %s", seg)
			}
			cur = arr[idx]
		}
	}
	return cur, nil
}
