package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Port     *int     `yaml:"port"`
	Backends []string `yaml:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %s", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Error parsing config file: %s", err)
	}

	if len(config.Backends) == 0 {
		return nil, fmt.Errorf("No backends found in config file. Please enter at least one.")
	}

	if config.Port == nil {
		return nil, fmt.Errorf("No port found in config file. Please enter it.")
	}

	port := *config.Port
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("Invalid port number %d. Port can be between 1 and 65535", config.Port)
	}

	return &config, nil
}
