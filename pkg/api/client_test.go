package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com", 30*time.Second)

	if client.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %v, want %v", client.BaseURL, "https://api.example.com")
	}

	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", client.HTTPClient.Timeout, 30*time.Second)
	}

	if client.Verbose != false {
		t.Errorf("Verbose = %v, want %v", client.Verbose, false)
	}
}

func TestSetVerbose(t *testing.T) {
	client := NewClient("https://api.example.com", 30*time.Second)

	client.SetVerbose(true)
	if client.Verbose != true {
		t.Errorf("Verbose = %v, want %v after SetVerbose(true)", client.Verbose, true)
	}

	client.SetVerbose(false)
	if client.Verbose != false {
		t.Errorf("Verbose = %v, want %v after SetVerbose(false)", client.Verbose, false)
	}
}

func TestGet(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("Expected path /users, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	resp, err := client.Get("/users")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestGetWithVerbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	client.SetVerbose(true)

	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("Get() with verbose error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestGetError(t *testing.T) {
	// Use an invalid URL to trigger an error
	client := NewClient("http://invalid.invalid.invalid", 1*time.Second)

	_, err := client.Get("/test")
	if err == nil {
		t.Error("Get() expected error for invalid URL")
	}
}

func TestGetDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL, 5*time.Second)
			resp, err := client.Get("/test")
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.statusCode)
			}
		})
	}
}
