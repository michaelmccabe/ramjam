package runner

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCaptureHeaderWithRegex(t *testing.T) {
	// Mock server
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

	// Create a temporary YAML file
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
	if err := r.RunPaths([]string{tmpFile.Name()}); err != nil {
		t.Fatalf("RunPaths failed: %v", err)
	}
}
