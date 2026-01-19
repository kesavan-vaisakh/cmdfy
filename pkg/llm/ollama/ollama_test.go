package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
)

func TestOllamaProvider_GenerateCommand(t *testing.T) {
	// Mock response from Ollama
	mockResponse := ChatResponse{
		Model: "llama3",
		Message: ChatMessage{
			Role: "assistant",
			Content: `{
				"steps": [
					{"tool": "echo", "args": ["hello"]}
				],
				"explanation": "Say hello",
				"dangerous": false
			}`,
		},
		Done:      true,
		EvalCount: 42,
	}

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path /api/chat, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Simulate latency
		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	cfg := llm.ProviderConfig{
		BaseURL: ts.URL,
		Model:   "llama3",
	}

	provider, err := NewOllamaProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	meta := llm.SystemMetadata{
		OS:    "linux",
		Shell: "bash",
	}

	result, err := provider.GenerateCommand(context.Background(), "say hello", meta)
	if err != nil {
		t.Fatalf("GenerateCommand failed: %v", err)
	}

	if len(result.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(result.Steps))
	}

	if result.Steps[0].Tool != "echo" {
		t.Errorf("Expected tool 'echo', got '%s'", result.Steps[0].Tool)
	}

	// Verify Metrics
	if result.Metrics.TokenCount != 42 {
		t.Errorf("Expected 42 tokens, got %d", result.Metrics.TokenCount)
	}

	if result.Metrics.Latency == "" {
		t.Error("Expected latency to be recorded")
	}
}
