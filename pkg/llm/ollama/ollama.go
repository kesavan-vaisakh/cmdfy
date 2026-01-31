package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"time"

	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
)

type OllamaProvider struct {
	baseURL string
	model   string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Format   string        `json:"format,omitempty"`
}

type ChatResponse struct {
	Model      string      `json:"model"`
	Message    ChatMessage `json:"message"`
	Done       bool        `json:"done"`
	EvalCount  int         `json:"eval_count"`
	TotalQueue int         `json:"total_duration"` // nanoseconds
}

func init() {
	llm.RegisterProvider("ollama", NewOllamaProvider)
}

func NewOllamaProvider(cfg llm.ProviderConfig) (llm.Provider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   cfg.Model,
	}, nil
}

// GenerateCommand generates a command using Ollama
func (p *OllamaProvider) GenerateCommand(ctx context.Context, query string, meta llm.SystemMetadata) (*model.CommandResult, error) {
	if p.model == "" {
		p.model = "llama3" // Default
	}
	// Limit available commands to avoid overwhelming the context
	maxCommands := 50
	truncated := false
	if len(meta.AvailableCommands) > maxCommands {
		meta.AvailableCommands = meta.AvailableCommands[:maxCommands]
		truncated = true
	}
	commandsList := strings.Join(meta.AvailableCommands, ", ")

	availableToolsSuffix := ""
	if truncated {
		availableToolsSuffix = " (and others)"
	}

	previousErrorSection := ""
	if meta.PreviousError != "" {
		previousErrorSection = fmt.Sprintf("\n\nTHE USER IS TRYING TO FIX A COMMAND THAT FAILED.\nError output:\n%s\n\nAnalyze this error and generate a fixed command.", meta.PreviousError)
	}

	examplesSection := ""
	if len(meta.FewShotExamples) > 0 {
		examplesSection = "\n\nReference - Here are similar commands the user has used before:\n"
		for _, ex := range meta.FewShotExamples {
			examplesSection += fmt.Sprintf("- Query: %s\n  Command: %s\n  Origin: %s\n", ex.Query, ex.Command, ex.Provider)
		}
	}

	prompt := fmt.Sprintf(`
You are a command line expert.
Your task is to translate the following natural language request into a shell command or a pipeline of commands.
You MUST return a JSON object with strictly these fields: "steps", "explanation", "dangerous".
Do NOT list files or answer the question directly. Generate the command to do it.

Schema:
{
  "steps": [
    {
      "tool": "string",
      "args": ["string"],
      "op": "string"
    }
  ],
  "explanation": "string",
  "dangerous": boolean
}

Operating System: %s
Shell: %s
Available Tools: %s%s%s%s
Request: %s
`, meta.OS, meta.Shell, commandsList, availableToolsSuffix, examplesSection, previousErrorSection, query)

	reqBody := ChatRequest{
		Model: p.model,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant that generates structured shell commands in JSON. Use the schema provided."},
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Format: "json",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", strings.TrimRight(p.baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	latency := time.Since(startTime)

	result := strings.TrimSpace(chatResp.Message.Content)
	// Strip markdown code blocks
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	// Sometimes Ollama might still be chatty, finding the first '{' and last '}' might be safer
	start := strings.Index(result, "{")
	end := strings.LastIndex(result, "}")
	if start != -1 && end != -1 && end > start {
		result = result[start : end+1]
	}

	var cmd model.CommandResult
	if err := json.Unmarshal([]byte(result), &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w. raw: %s", err, result)
	}

	cmd.Metrics = model.Metrics{
		Latency:    latency.Round(time.Millisecond).String(),
		TokenCount: chatResp.EvalCount,
	}

	return &cmd, nil
}
