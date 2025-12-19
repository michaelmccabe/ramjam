package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid YAML",
			input:   []byte("key: value\nlist:\n  - item1\n  - item2"),
			wantErr: false,
		},
		{
			name:    "empty YAML",
			input:   []byte(""),
			wantErr: false,
		},
		{
			name:    "invalid YAML",
			input:   []byte("key: [unclosed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target map[string]interface{}
			err := Parse(tt.input, &target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid YAML file
	validYAML := []byte("name: test\nversion: 1.0")
	validPath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validPath, validYAML, 0644); err != nil {
		t.Fatalf("Failed to write valid.yaml: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantName string
	}{
		{
			name:     "valid file",
			path:     validPath,
			wantErr:  false,
			wantName: "test",
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target struct {
				Name    string `yaml:"name"`
				Version string `yaml:"version"`
			}
			err := LoadFile(tt.path, &target)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && target.Name != tt.wantName {
				t.Errorf("LoadFile() name = %v, want %v", target.Name, tt.wantName)
			}
		})
	}
}

func TestLoader(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test YAML file
	testYAML := []byte("message: hello")
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), testYAML, 0644); err != nil {
		t.Fatalf("Failed to write test.yaml: %v", err)
	}

	loader := NewLoader(tmpDir)

	var target struct {
		Message string `yaml:"message"`
	}

	if err := loader.Load("test.yaml", &target); err != nil {
		t.Errorf("Loader.Load() error = %v", err)
		return
	}

	if target.Message != "hello" {
		t.Errorf("Loader.Load() message = %v, want %v", target.Message, "hello")
	}
}

func TestLoadBytes(t *testing.T) {
	data := []byte("items:\n  - one\n  - two\n  - three")

	var target struct {
		Items []string `yaml:"items"`
	}

	if err := LoadBytes(data, &target); err != nil {
		t.Errorf("LoadBytes() error = %v", err)
		return
	}

	if len(target.Items) != 3 {
		t.Errorf("LoadBytes() items length = %v, want %v", len(target.Items), 3)
	}

	expected := []string{"one", "two", "three"}
	for i, item := range target.Items {
		if item != expected[i] {
			t.Errorf("LoadBytes() item[%d] = %v, want %v", i, item, expected[i])
		}
	}
}
