package config

import (
	"os"
	"path/filepath"

	e "github.com/michaelmccabe/ramjam/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing YAML configuration files
type Loader struct {
	basePath string
}

// NewLoader creates a new Loader with the specified base path for resources
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// Load reads a YAML file and unmarshals it into the provided target
func (l *Loader) Load(filename string, target interface{}) error {
	path := filepath.Join(l.basePath, filename)
	return LoadFile(path, target)
}

// LoadFile reads a YAML file from the given path and unmarshals it into the target
func LoadFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err := e.Wrapf(err, "failed to read file %s", path); err != nil {
		return err
	}

	return Parse(data, target)
}

// Parse parses YAML data and unmarshals it into the target
func Parse(data []byte, target interface{}) error {
	return e.Wrap(yaml.Unmarshal(data, target), "failed to parse YAML")
}

// LoadBytes is a convenience function to load YAML from embedded bytes
func LoadBytes(data []byte, target interface{}) error {
	return Parse(data, target)
}
