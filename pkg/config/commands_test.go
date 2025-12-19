package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCommands(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "commands_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test commands YAML file
	commandsYAML := `
root:
  use: "testcmd"
  short: "Test command short"
  long: "Test command long description"
get:
  use: "get [url]"
  short: "Get short"
  long: "Get long description"
version:
  use: "version"
  short: "Version short"
  long: "Version long"
`
	path := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(path, []byte(commandsYAML), 0644); err != nil {
		t.Fatalf("Failed to write commands.yaml: %v", err)
	}

	config, err := LoadCommands(path)
	if err != nil {
		t.Fatalf("LoadCommands() error = %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"root.use", config.Root.Use, "testcmd"},
		{"root.short", config.Root.Short, "Test command short"},
		{"root.long", config.Root.Long, "Test command long description"},
		{"get.use", config.Get.Use, "get [url]"},
		{"get.short", config.Get.Short, "Get short"},
		{"version.use", config.Version.Use, "version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestLoadCommandsFromBytes(t *testing.T) {
	data := []byte(`
root:
  use: "app"
  short: "App short"
  long: "App long"
get:
  use: "get"
  short: "Get short"
  long: "Get long"
version:
  use: "ver"
  short: "Ver short"
  long: "Ver long"
`)

	config, err := LoadCommandsFromBytes(data)
	if err != nil {
		t.Fatalf("LoadCommandsFromBytes() error = %v", err)
	}

	if config.Root.Use != "app" {
		t.Errorf("root.use = %v, want %v", config.Root.Use, "app")
	}

	if config.Get.Short != "Get short" {
		t.Errorf("get.short = %v, want %v", config.Get.Short, "Get short")
	}
}

func TestLoadCommandsError(t *testing.T) {
	_, err := LoadCommands("/nonexistent/path/commands.yaml")
	if err == nil {
		t.Error("LoadCommands() expected error for non-existent file")
	}
}

func TestLoadCommandsFromBytesError(t *testing.T) {
	invalidYAML := []byte("key: [unclosed")
	_, err := LoadCommandsFromBytes(invalidYAML)
	if err == nil {
		t.Error("LoadCommandsFromBytes() expected error for invalid YAML")
	}
}

func TestCommandTextMultiline(t *testing.T) {
	data := []byte(`
root:
  use: "test"
  short: "Short"
  long: |
    This is a multiline
    long description
    with multiple lines
get:
  use: "get"
  short: "Get"
  long: "Single line"
version:
  use: "version"
  short: "Version"
  long: "Version long"
`)

	config, err := LoadCommandsFromBytes(data)
	if err != nil {
		t.Fatalf("LoadCommandsFromBytes() error = %v", err)
	}

	if !strings.Contains(config.Root.Long, "multiline") {
		t.Error("Expected multiline long description to contain 'multiline'")
	}

	if !strings.Contains(config.Root.Long, "multiple lines") {
		t.Error("Expected multiline long description to contain 'multiple lines'")
	}
}
