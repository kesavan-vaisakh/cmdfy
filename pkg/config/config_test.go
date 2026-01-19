package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CurrentProvider != "gemini" {
		t.Errorf("Expected default provider 'gemini', got '%s'", cfg.CurrentProvider)
	}
	if cfg.Providers == nil {
		t.Error("Expected Providers map to be initialized")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "cmdfy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock GetConfigPath by manually writing to a file in tmpDir
	// Since LoadConfig calls GetConfigPath which looks at HOME, we can't easily mock that without changing the code structure
	// or mocking os.UserHomeDir (which isn't easy in Go).
	// Instead, let's test the Unmarshal logic or if we can make LoadConfig take a path.
	// But `LoadConfig` is hardcoded to `GetConfigPath`.

	// Recommendation: Refactor LoadConfig to use a path passed in, or set HOME env var.
	// We'll set HOME env var for this test.

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	cfg := &Config{
		CurrentProvider: "ollama",
		Providers: map[string]LLMConfig{
			"ollama": {
				BaseURL: "http://test-url",
				Model:   "llama3-test",
			},
		},
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedCfg.CurrentProvider != "ollama" {
		t.Errorf("Expected loaded provider 'ollama', got '%s'", loadedCfg.CurrentProvider)
	}

	if loadedCfg.Providers["ollama"].BaseURL != "http://test-url" {
		t.Errorf("Expected base URL 'http://test-url', got '%s'", loadedCfg.Providers["ollama"].BaseURL)
	}
}
