package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRunCmdRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c == runCmd {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("run command should be registered with root")
	}
}

func TestRunCmdNoArgs(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	defer rootCmd.SetArgs(nil)

	rootCmd.SetArgs([]string{"run"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestRunCmdInvalidPath(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	defer rootCmd.SetArgs(nil)

	rootCmd.SetArgs([]string{"run", "/nonexistent/path/file.yaml"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestRunCmdHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("expected Accept header application/json, got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("X-Run-Test") != "happy-path" {
			t.Fatalf("expected X-Run-Test header happy-path, got %s", r.Header.Get("X-Run-Test"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"ok"}`))
	}))
	defer srv.Close()

	yamlContent := fmt.Sprintf(`
metadata:
  name: "Test"
config:
  base_url: "%s"
workflow:
- step: "get-message"
  request:
    method: "GET"
    url: "${base_url}/"
    headers:
      Accept: "application/json"
      X-Run-Test: "happy-path"
  expect:
    status: 200
    json_path_match:
    - path: "message"
      value: "ok"
`, srv.URL)

	tmpFile, err := os.CreateTemp("", "run_cmd_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}
	tmpFile.Close()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)
	defer rootCmd.SetArgs(nil)

	rootCmd.SetArgs([]string{"run", tmpFile.Name()})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}
}
