package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
}

type ChatResponse struct {
	Model   string      `json:"model"`
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
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

	commandsList := strings.Join(meta.AvailableCommands, ", ")
	filesList := strings.Join(meta.CurrentDirFiles, ", ")
	prompt := fmt.Sprintf(`
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
Request: %s
`, meta.OS, meta.Shell, commandsList, filesList, query)

	reqBody := ChatRequest{
		Model: p.model,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant that generates structured shell commands in JSON. Do not output anything other than JSON."},
			{Role: "user", Content: prompt},
		},
		Stream: false,
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

	return &cmd, nil
}
