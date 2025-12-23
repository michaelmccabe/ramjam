package runner

import (
	"bytes"
	"encoding/json"
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
		Method  string                 `yaml:"method"`
		URL     string                 `yaml:"url"`
		Headers map[string]string      `yaml:"headers"`
		Body    map[string]interface{} `yaml:"body"`
	}

	StepExpect struct {
		Status        int           `yaml:"status"`
		JSONPathMatch []JSONPathVal `yaml:"json_path_match"`
	}

	JSONPathVal struct {
		Path  string      `yaml:"path"`
		Value interface{} `yaml:"value"`
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

	WorkflowError struct {
		Errors []error
	}
)

func (e *StepError) Error() string {
	return fmt.Sprintf("step %q in %s failed: %v", e.Step, e.File, e.Err)
}

func (e *WorkflowError) Error() string {
	return fmt.Sprintf("%d steps failed", len(e.Errors))
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
		err  error
	}
	results := make(chan result, len(files))

	for _, f := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			logs, err := r.runFile(f)
			results <- result{logs, err}
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
		if res.err != nil {
			if we, ok := res.err.(*WorkflowError); ok {
				errs = append(errs, we.Errors...)
			} else {
				errs = append(errs, res.err)
			}
		}
	}

	if len(errs) > 0 {
		return &WorkflowError{Errors: errs}
	}
	return nil
}

func (r *Runner) collectFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("unable to access %s: %w", path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %s: %w", path, err)
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

func (r *Runner) runFile(path string) ([]string, error) {
	var logs []string
	prefix := filepath.Base(path)
	log := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs = append(logs, fmt.Sprintf("[%s] %s", prefix, msg))
	}

	log("Running workflow file: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return logs, fmt.Errorf("read %s: %w", path, err)
	}
	var spec InstructionsFile
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return logs, fmt.Errorf("parse %s: %w", path, err)
	}

	if spec.Metadata.Name != "" {
		prefix = spec.Metadata.Name
	}

	vars := map[string]string{
		"base_url": spec.Config.BaseURL,
	}

	var errs []error
	for _, step := range spec.Workflow {
		if err := r.executeStep(step, vars, log); err != nil {
			errs = append(errs, &StepError{
				File:        path,
				Step:        step.Step,
				Description: step.Description,
				Err:         err,
			})
		}
	}

	if len(errs) > 0 {
		return logs, &WorkflowError{Errors: errs}
	}
	return logs, nil
}

func (r *Runner) executeStep(step Step, vars map[string]string, log func(string, ...interface{})) error {
	if r.verbose {
		log("Executing step: %s", step.Step)
	}

	method := strings.ToUpper(strings.TrimSpace(step.Request.Method))
	if method == "" {
		method = http.MethodGet
	}

	url := applyVars(step.Request.URL, vars)
	if !strings.HasPrefix(url, "http") && vars["base_url"] != "" {
		url = strings.TrimSuffix(vars["base_url"], "/") + "/" + strings.TrimPrefix(url, "/")
	}

	bodyReader := io.Reader(nil)
	if len(step.Request.Body) > 0 {
		body := applyVarsToInterface(step.Request.Body, vars)
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "ramjam-cli")
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range step.Request.Headers {
		req.Header.Set(k, applyVars(v, vars))
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if r.verbose {
		log("Received status: %d", resp.StatusCode)
	}

	if step.Expect.Status != 0 && resp.StatusCode != step.Expect.Status {
		return fmt.Errorf("expected status %d, got %d", step.Expect.Status, resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var jsonObj interface{}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &jsonObj); err != nil {
			return fmt.Errorf("parse response json: %w", err)
		}
	}

	for _, matcher := range step.Expect.JSONPathMatch {
		actual, err := evalJSONPath(jsonObj, matcher.Path)
		if err != nil {
			return fmt.Errorf("jsonpath %s: %w", matcher.Path, err)
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
			if err != nil {
				return fmt.Errorf("capture json_path %s: %w", cap.JSONPath, err)
			}
		} else if cap.Header != "" {
			headerVal := resp.Header.Get(cap.Header)
			if cap.Regex != "" {
				re, err := regexp.Compile(cap.Regex)
				if err != nil {
					return fmt.Errorf("invalid regex %s: %w", cap.Regex, err)
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
