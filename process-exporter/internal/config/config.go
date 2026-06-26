package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Process represents a process to be monitored
type Process struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
}

// Config holds the configuration for the application
type Config struct {
	Processes []Process `yaml:"processes"`
}

// LoadConfig reads the configuration from a YAML file
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}
