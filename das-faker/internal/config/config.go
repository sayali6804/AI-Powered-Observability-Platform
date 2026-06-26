package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	BaseDir            string              `yaml:"baseDir"`
	FileGenerationRate int                 `yaml:"fileGenerationRate"` // Files per minute
	ProcessingDelayMin time.Duration       `yaml:"processingDelayMin"` // Minimum delay before processing
	ProcessingDelayMax time.Duration       `yaml:"processingDelayMax"` // Maximum delay before processing
	FailureRate        float64             `yaml:"failureRate"`        // Percentage of files that fail (0-1)
	Types              []string            `yaml:"types"`              // email, chat, voice
	Sources            map[string][]string `yaml:"sources"`            // Mapping of type to possible sources
	FilesPerBatch      int                 `yaml:"filesPerBatch"`      // Number of files to create in each batch
	MaxConcurrentFiles int                 `yaml:"maxConcurrentFiles"` // Maximum number of files to process concurrently
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		BaseDir:            "c:/mnt/data/das",
		FileGenerationRate: 60,
		ProcessingDelayMin: 5 * time.Second,
		ProcessingDelayMax: 15 * time.Second,
		FailureRate:        0.02, // 2% failure rate
		Types:              []string{"email", "chat", "voice"},
		Sources: map[string][]string{
			"email": {"journal", "non-journal"},
			"chat":  {"bloomberg"},
			"voice": {"bloomberg"},
		},
		FilesPerBatch:      10,
		MaxConcurrentFiles: 100,
	}
}

// Load loads configuration from a file
func Load(path string) (Config, error) {
	config := DefaultConfig()

	// If the file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Save default config for future use
		if err := SaveConfig(config, path); err != nil {
			return config, err
		}
		return config, nil
	}

	// Read and parse the file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	return config, err
}

// SaveConfig saves configuration to a file
func SaveConfig(config Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
