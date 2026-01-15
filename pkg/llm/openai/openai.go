package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the llm.Provider interface
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func init() {
	// Register both "openai" and "chatGPT" to be user-friendly
	llm.RegisterProvider("openai", NewOpenAIProvider)
	llm.RegisterProvider("chatGPT", NewOpenAIProvider)
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg llm.ProviderConfig) (llm.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required for OpenAI")
	}
	model := cfg.Model
	if model == "" {
		model = openai.GPT3Dot5Turbo // Default model, cheap and fast
	}

	config := openai.DefaultConfig(cfg.APIKey)
	client := openai.NewClientWithConfig(config)

	return &OpenAIProvider{
		client: client,
		model:  model,
	}, nil
}

// GenerateCommand generates a command using OpenAI
func (p *OpenAIProvider) GenerateCommand(ctx context.Context, query string, meta llm.SystemMetadata) (*model.CommandResult, error) {
	commandsList := strings.Join(meta.AvailableCommands, ", ")
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
Request: %s
`, meta.OS, meta.Shell, commandsList, query)

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that generates structured shell commands in JSON.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// Note: We could use ResponseFormat: { Type: "json_object" } for newer models,
			// but to keep it simple and compatible with older/cheaper ones, we prompt instruction.
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received")
	}

	result := strings.TrimSpace(resp.Choices[0].Message.Content)
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
