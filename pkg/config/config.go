package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LLMConfig holds configuration for a specific LLM provider
type LLMConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url,omitempty"`
	Model   string `yaml:"model,omitempty"`
}

// Config holds the application configuration
type Config struct {
	CurrentProvider string               `yaml:"current_provider"`
	Providers       map[string]LLMConfig `yaml:"providers"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		CurrentProvider: "gemini",
		Providers:       make(map[string]LLMConfig),
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cmdfy", "config.yaml"), nil
}

// LoadConfig loads the configuration from disk or returns default if not found
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]LLMConfig)
	}

	return cfg, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil { // Secure permissions
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
