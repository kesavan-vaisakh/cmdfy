package anthropic

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

const (
	defaultBaseURL = "https://api.anthropic.com/v1/messages"
	defaultModel   = "claude-3-5-sonnet-latest"
	apiVersion     = "2023-06-01"
)

type AnthropicProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewAnthropicProvider creates a new instance of AnthropicProvider
func NewAnthropicProvider(config llm.ProviderConfig) (llm.Provider, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	modelName := config.Model
	if modelName == "" {
		modelName = defaultModel
	}

	return &AnthropicProvider{
		apiKey:  config.APIKey,
		model:   modelName,
		baseURL: baseURL,
		client:  &http.Client{},
	}, nil
}

func init() {
	llm.RegisterProvider("anthropic", NewAnthropicProvider)
	llm.RegisterProvider("claude", NewAnthropicProvider)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessagesRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens"`
}

type MessagesResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) GenerateCommand(ctx context.Context, query string, meta llm.SystemMetadata) (*model.CommandResult, error) {
	commandsList := strings.Join(meta.AvailableCommands, ", ")
	filesList := strings.Join(meta.CurrentDirFiles, ", ")

	systemPrompt := fmt.Sprintf(`
You are a command line expert. 
Your task is to translate the following natural language request into a shell command or a pipeline of commands.
Respond ONLY with a valid JSON object matching this schema:
{
  "steps": [
    {
      "tool": "string (the primary command, e.g. git, grep)",
      "args": ["string", "arguments"],
      "op": "string (operator to connect to next step: | (pipe), && (and), ; (seq), || (or), > (redirect), >> (append). Empty for last step.)"
    }
  ],
  "explanation": "string (brief explanation of the entire pipeline)",
  "dangerous": boolean (true if ANY step modifies files significantly, deletes data, or has destructive side effects)
}

Operating System: %s
Shell: %s
Available Tools: %s
Current Directory Files: %s
`, meta.OS, meta.Shell, commandsList, filesList)

	userMessage := Message{
		Role:    "user",
		Content: query,
	}

	reqBody := MessagesRequest{
		Model:     p.model,
		Messages:  []Message{userMessage},
		System:    systemPrompt,
		MaxTokens: 1024,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	startTime := time.Now()

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var response MessagesResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	latency := time.Since(startTime)

	if response.Error != nil {
		return nil, fmt.Errorf("anthropic api error: %s - %s", response.Error.Type, response.Error.Message)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from anthropic")
	}

	text := response.Content[0].Text

	// Sanitize output (remove markdown blocks if present)
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result model.CommandResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w\nResponse was: %s", err, text)
	}

	result.Metrics = model.Metrics{
		Latency:    latency.Round(time.Millisecond).String(),
		TokenCount: response.Usage.InputTokens + response.Usage.OutputTokens,
	}

	return &result, nil
}
