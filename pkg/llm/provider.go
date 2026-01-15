package llm

import (
	"context"
	"fmt"

	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
)

// SystemMetadata contains context about the user's system to aid generation
type SystemMetadata struct {
	OS                string
	Shell             string
	AvailableCommands []string
	CurrentDirFiles   []string
}

// Provider defines the interface for an LLM provider
type Provider interface {
	// GenerateCommand generates a shell command based on the query and system metadata
	GenerateCommand(ctx context.Context, query string, meta SystemMetadata) (*model.CommandResult, error)
}

// ProviderConfig holds configuration for creating a provider
type ProviderConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Factory is a function that creates a new Provider instance
type Factory func(cfg ProviderConfig) (Provider, error)

var providers = make(map[string]Factory)

// RegisterProvider registers a provider factory
func RegisterProvider(name string, factory Factory) {
	providers[name] = factory
}

// GetProvider returns a new instance of the requested provider
func GetProvider(name string, cfg ProviderConfig) (Provider, error) {
	factory, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return factory(cfg)
}
