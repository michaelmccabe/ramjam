package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCmd(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.Header.Get("User-Agent") != "ramjam-cli" {
			t.Errorf("Expected User-Agent 'ramjam-cli', got %s", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "hello"}`))
	}))
	defer server.Close()

	// Capture output
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	// Set args and execute
	rootCmd.SetArgs([]string{"get", server.URL})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestGetCmdVerbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	rootCmd.SetArgs([]string{"get", "--verbose", server.URL})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() with verbose error = %v", err)
	}
}

func TestGetCmdWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	rootCmd.SetArgs([]string{"get", "--timeout", "10", server.URL})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() with timeout error = %v", err)
	}
}

func TestGetCmdNoArgs(t *testing.T) {
	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)

	rootCmd.SetArgs([]string{"get"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when no URL provided")
	}
}

func TestGetCmdInvalidURL(t *testing.T) {
	rootCmd.SetArgs([]string{"get", "not-a-valid-url"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestGetCmdFlags(t *testing.T) {
	// Test that the timeout flag exists and has correct default
	flag := getCmd.Flags().Lookup("timeout")
	if flag == nil {
		t.Fatal("timeout flag not found")
	}

	if flag.DefValue != "30" {
		t.Errorf("timeout default value = %v, want %v", flag.DefValue, "30")
	}

	if flag.Shorthand != "t" {
		t.Errorf("timeout shorthand = %v, want %v", flag.Shorthand, "t")
	}
}

func TestGetCmdUsage(t *testing.T) {
	if getCmd.Use != "get [url]" {
		t.Errorf("Use = %v, want %v", getCmd.Use, "get [url]")
	}

	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if getCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}
