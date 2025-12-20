package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCommands(t *testing.T) {
	// Load the actual commands.yaml file from resources
	path := filepath.Join("..", "..", "resources", "commands.yaml")

	config, err := LoadCommands(path)
	if err != nil {
		t.Fatalf("LoadCommands() error = %v", err)
	}

	// Validate root command
	if config.Root.Use != "ramjam" {
		t.Errorf("Root.Use = %v, want ramjam", config.Root.Use)
	}
	if config.Root.Short == "" {
		t.Error("Root.Short should not be empty")
	}

	// Validate run command
	if config.Run.Use == "" {
		t.Error("Run.Use should not be empty")
	}
	if config.Run.Short == "" {
		t.Error("Run.Short should not be empty")
	}

	// Validate version command
	if config.Version.Use != "version" {
		t.Errorf("Version.Use = %v, want version", config.Version.Use)
	}
	if config.Version.Short == "" {
		t.Error("Version.Short should not be empty")
	}
}

func TestLoadCommandsFromBytes(t *testing.T) {
	data := []byte(`
root:
  use: "app"
  short: "App short"
  long: "App long"
run:
  use: "run [file]"
  short: "Run short"
  long: "Run long"
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

	if config.Version.Short != "Ver short" {
		t.Errorf("version.short = %v, want %v", config.Version.Short, "Ver short")
	}

	if config.Run.Use != "run [file]" {
		t.Errorf("run.use = %v, want %v", config.Run.Use, "run [file]")
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
run:
  use: "run"
  short: "Run"
  long: "Run long"
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

func TestCommandsConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid config",
			yaml: `
root:
  use: "ramjam"
  short: "Test CLI"
  long: "Test CLI tool"
run:
  use: "run [file]"
  short: "Execute workflow"
  long: "Execute a YAML workflow file"
version:
  use: "version"
  short: "Print version"
  long: "Print the version number"
`,
			wantErr: false,
		},
		{
			name: "missing root command",
			yaml: `
run:
  use: "run [file]"
  short: "Execute workflow"
version:
  use: "version"
  short: "Print version"
`,
			wantErr: false, // YAML will just have empty Root
		},
		{
			name:    "empty config",
			yaml:    ``,
			wantErr: false, // Empty YAML is valid, just all fields will be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.yaml)
			_, err := LoadCommandsFromBytes(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadCommandsFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
