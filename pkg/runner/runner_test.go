package runner

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSimpleGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("expected /users, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Simple GET"
config:
  base_url: "%s"
workflow:
- step: "get-users"
  request:
    method: "GET"
    url: "/users"
  expect:
    status: 200
`, srv.URL)

	runTest(t, yamlContent)
}

func TestVariableSubstitution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config" {
			w.Write([]byte(`{"id": "123", "role": "admin"}`))
			return
		}
		if r.URL.Path == "/users/123" {
			if r.Header.Get("X-Request-ID") != "req-123" {
				t.Errorf("expected header req-123, got %s", r.Header.Get("X-Request-ID"))
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"123"`) || !strings.Contains(string(body), `"role":"admin"`) {
				t.Errorf("body mismatch: %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Variable Substitution"
config:
  base_url: "%s"
workflow:
- step: "get-id"
  request:
    method: "GET"
    url: "/config"
  expect:
    status: 200
  capture:
  - json_path: "id"
    as: "user_id"
  - json_path: "role"
    as: "user_role"

- step: "use-vars"
  request:
    method: "POST"
    url: "/users/${user_id}"
    headers:
      X-Request-ID: "req-${user_id}"
    body:
      id: "${user_id}"
      role: "${user_role}"
  expect:
    status: 200
`, srv.URL)

	runTest(t, yamlContent)
}

func TestJsonPathMatching(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/data" {
			w.Write([]byte(`{
				"user": {
					"name": "Alice",
					"age": 30,
					"tags": ["admin", "editor"]
				}
			}`))
			return
		}
		if r.URL.Path == "/list" {
			w.Write([]byte(`[
				{"id": 1, "title": "Hello"},
				{"id": 2, "title": "World"}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "JSONPath Matching"
config:
  base_url: "%s"
workflow:
- step: "check-json"
  request:
    method: "GET"
    url: "/data"
  expect:
    status: 200
    json_path_match:
    - path: "user.name"
      value: "Alice"
    - path: "user.age"
      value: "30"
    - path: "user.tags[0]"
      value: "admin"

- step: "check-filter"
  request:
    method: "GET"
    url: "/list"
  expect:
    status: 200
    json_path_match:
    - path: "$[?(@.id==2)].title"
      value: "World"
`, srv.URL)

	runTest(t, yamlContent)
}

func TestExpectHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", "520")
		w.Header().Set("Cache-Control", "max-age=3600, public")
		payload := `{"status": "ok"}`
		if pad := 520 - len(payload); pad > 0 {
			payload += strings.Repeat(" ", pad)
		}
		w.Write([]byte(payload))
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Header Expect"
config:
  base_url: "%s"
workflow:
- step: "header-check"
  request:
    method: "GET"
    url: "/users"
  expect:
    status: 200
    headers:
    - name: "Content-Type"
      contains: "application/json"
    - name: "Content-Length"
      value: "520"
  capture:
  - header: "Cache-Control"
    regex: "max-age=([0-9]+)"
    as: "cache_max_age"
  output:
    print: "Cache max-age is ${cache_max_age}"
`, srv.URL)

	runTest(t, yamlContent)
}

func TestCaptureHeaderWithRegex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.Header().Set("Authorization", "Bearer my-secret-token")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"logged_in"}`))
			return
		}
		if r.URL.Path == "/verify" {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "my-secret-token") {
				t.Errorf("expected body to contain 'my-secret-token', got '%s'", string(body))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"verified"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Header Capture Test"
config:
  base_url: "%s"
workflow:
- step: "login"
  request:
    method: "POST"
    url: "/login"
  capture:
  - header: "Authorization"
    regex: "Bearer (.*)"
    as: "jwt"
- step: "verify-token"
  request:
    method: "POST"
    url: "/verify"
    body:
      token: "${jwt}"
  expect:
    status: 200
`, srv.URL)

	runTest(t, yamlContent)
}

func TestExpectStatusFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Status Failure"
config:
  base_url: "%s"
workflow:
- step: "fail-status"
  request:
    method: "GET"
    url: "/"
  expect:
    status: 200
`, srv.URL)

	err := runTestError(t, yamlContent)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "expected status 200, got 500") {
		t.Errorf("did not find expected error message 'expected status 200, got 500'. Got: %v", err)
	}
}

func TestExpectJsonPathFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "error"}`))
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "JSONPath Failure"
config:
  base_url: "%s"
workflow:
- step: "fail-json"
  request:
    method: "GET"
    url: "/"
  expect:
    status: 200
    json_path_match:
    - path: "status"
      value: "success"
`, srv.URL)

	err := runTestError(t, yamlContent)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), `expected "success", got "error"`) {
		t.Errorf("did not find expected error message 'expected \"success\", got \"error\"'. Got: %v", err)
	}
}

func TestDirectoryExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Create a temp dir
	tmpDir, err := os.MkdirTemp("", "ramjam_test_dir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two yaml files
	file1 := filepath.Join(tmpDir, "test1.yaml")
	content1 := fmt.Sprintf(`
metadata:
  name: "Test 1"
config:
  base_url: "%s"
workflow:
- step: "step1"
  request:
    url: "/1"
`, srv.URL)
	os.WriteFile(file1, []byte(content1), 0644)

	file2 := filepath.Join(tmpDir, "test2.yaml")
	content2 := fmt.Sprintf(`
metadata:
  name: "Test 2"
config:
  base_url: "%s"
workflow:
- step: "step2"
  request:
    url: "/2"
`, srv.URL)
	os.WriteFile(file2, []byte(content2), 0644)

	r := New(10*time.Second, false)
	if err := r.RunPaths([]string{tmpDir}); err != nil {
		t.Fatalf("RunPaths failed: %v", err)
	}
}

func TestContinueOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Continue On Failure"
config:
  base_url: "%s"
workflow:
- step: "fail-step"
  request:
    url: "/fail"
  expect:
    status: 200
- step: "success-step"
  request:
    url: "/success"
  expect:
    status: 200
`, srv.URL)

	err := runTestError(t, yamlContent)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify the error message
	if !strings.Contains(err.Error(), "expected status 200, got 500") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Verify we have exactly 1 error if possible (errors.Join returns an interface{ Unwrap() []error })
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		errs := joined.Unwrap()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	}
}

func TestBodyFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		// Verify the body was loaded from the JSON file
		if !strings.Contains(bodyStr, `"title":"Test Post"`) {
			t.Errorf("expected title in body, got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, `"priority":"high"`) {
			t.Errorf("expected priority in body, got: %s", bodyStr)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123, "title": "Test Post", "priority": "high"}`))
	}))
	defer srv.Close()

	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "ramjam_bodyfile_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the JSON body file
	bodyJSON := `{
  "title": "Test Post",
  "body": "This is a test post",
  "userId": 1,
  "priority": "high"
}`
	bodyFilePath := filepath.Join(tmpDir, "test-body.json")
	if err := os.WriteFile(bodyFilePath, []byte(bodyJSON), 0644); err != nil {
		t.Fatalf("failed to write body file: %v", err)
	}

	// Create the YAML test file
	yamlContent := fmt.Sprintf(`
metadata:
  name: "Body File Test"
config:
  base_url: "%s"
workflow:
- step: "post-with-file"
  description: "POST with body from external JSON file"
  request:
    method: "POST"
    url: "/posts"
    body_file: "test-body.json"
  expect:
    status: 201
    json_path_match:
    - path: "title"
      value: "Test Post"
    - path: "priority"
      value: "high"
`, srv.URL)

	yamlFilePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlFilePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}

	// Run the test
	r := New(10*time.Second, true)
	if err := r.RunPaths([]string{yamlFilePath}); err != nil {
		t.Fatalf("RunPaths failed: %v", err)
	}
}

func TestBodyFileWithVariables(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		// Verify variables were substituted in the body loaded from file
		if !strings.Contains(bodyStr, `"userId":"42"`) {
			t.Errorf("expected userId to be 42, got: %s", bodyStr)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 999}`))
	}))
	defer srv.Close()

	tmpDir, err := os.MkdirTemp("", "ramjam_bodyfile_vars_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create JSON file with variable placeholder
	bodyJSON := `{
  "userId": "${user_id}",
  "action": "create"
}`
	bodyFilePath := filepath.Join(tmpDir, "body.json")
	if err := os.WriteFile(bodyFilePath, []byte(bodyJSON), 0644); err != nil {
		t.Fatalf("failed to write body file: %v", err)
	}

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Body File Variables Test"
config:
  base_url: "%s"
workflow:
- step: "capture-id"
  request:
    method: "GET"
    url: "/user"
  expect:
    status: 200
  capture:
  - json_path: "id"
    as: "user_id"

- step: "post-with-vars"
  request:
    method: "POST"
    url: "/action"
    body_file: "body.json"
  expect:
    status: 201
`, srv.URL)

	// Need to handle the capture step
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.Write([]byte(`{"id": "42"}`))
			return
		}
		if r.URL.Path == "/action" {
			body, _ := io.ReadAll(r.Body)
			bodyStr := string(body)
			if !strings.Contains(bodyStr, `"userId":"42"`) {
				t.Errorf("expected userId to be 42, got: %s", bodyStr)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 999}`))
			return
		}
	}))
	defer testSrv.Close()

	yamlContent = fmt.Sprintf(`
metadata:
  name: "Body File Variables Test"
config:
  base_url: "%s"
workflow:
- step: "capture-id"
  request:
    method: "GET"
    url: "/user"
  expect:
    status: 200
  capture:
  - json_path: "id"
    as: "user_id"

- step: "post-with-vars"
  request:
    method: "POST"
    url: "/action"
    body_file: "body.json"
  expect:
    status: 201
`, testSrv.URL)

	yamlFilePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlFilePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}

	r := New(10*time.Second, true)
	if err := r.RunPaths([]string{yamlFilePath}); err != nil {
		t.Fatalf("RunPaths failed: %v", err)
	}
}

// Helper to run a test from YAML content string
func runTest(t *testing.T, yamlContent string) {
	if err := runTestError(t, yamlContent); err != nil {
		t.Fatalf("RunPaths failed: %v", err)
	}
}

func runTestError(t *testing.T, yamlContent string) error {
	tmpFile, err := os.CreateTemp("", "runner_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}
	tmpFile.Close()

	r := New(10*time.Second, true)
	return r.RunPaths([]string{tmpFile.Name()})
}
