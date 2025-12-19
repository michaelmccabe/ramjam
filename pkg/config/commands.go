package config

// CommandText holds the text fields for a cobra command
type CommandText struct {
	Use   string `yaml:"use"`
	Short string `yaml:"short"`
	Long  string `yaml:"long"`
}

// CommandsConfig holds all command text definitions
type CommandsConfig struct {
	Root    CommandText `yaml:"root"`
	Get     CommandText `yaml:"get"`
	Version CommandText `yaml:"version"`
}

// LoadCommands loads command text from the given YAML file path
func LoadCommands(path string) (*CommandsConfig, error) {
	var config CommandsConfig
	if err := LoadFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadCommandsFromBytes loads command text from YAML bytes (useful for embedded resources)
func LoadCommandsFromBytes(data []byte) (*CommandsConfig, error) {
	var config CommandsConfig
	if err := Parse(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
