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
	Run     CommandText `yaml:"run"`
	Version CommandText `yaml:"version"`
}

// LoadCommands loads command text from a YAML file
func LoadCommands(path string) (*CommandsConfig, error) {
	var config CommandsConfig
	if err := LoadFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadCommandsFromBytes loads command text from byte data
func LoadCommandsFromBytes(data []byte) (*CommandsConfig, error) {
	var config CommandsConfig
	if err := Parse(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
