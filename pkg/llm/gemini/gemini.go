package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
	"google.golang.org/genai"
)

// GeminiProvider implements the llm.Provider interface
type GeminiProvider struct {
	client *genai.Client
	model  string
}

func init() {
	llm.RegisterProvider("gemini", NewGeminiProvider)
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(cfg llm.ProviderConfig) (llm.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required for Gemini")
	}
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash" // Default model
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

// GenerateCommand generates a command using Gemini
func (p *GeminiProvider) GenerateCommand(ctx context.Context, query string, meta llm.SystemMetadata) (*model.CommandResult, error) {
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

	resp, err := p.client.Models.GenerateContent(ctx, p.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates received")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			sb.WriteString(part.Text)
		}
	}

	result := strings.TrimSpace(sb.String())
	// Strip markdown code blocks
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	var cmd model.CommandResult
	if err := json.Unmarshal([]byte(result), &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w. raw: %s", err, result)
	}

	return &cmd, nil
}
