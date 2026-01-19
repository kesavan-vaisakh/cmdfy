package model

import (
	"encoding/json"
	"testing"
)

func TestCommandResult_JSON(t *testing.T) {
	jsonStr := `
	{
		"steps": [
			{
				"tool": "git",
				"args": ["status"],
				"op": "&&"
			},
			{
				"tool": "ls",
				"args": ["-la"]
			}
		],
		"explanation": "Check status and list files",
		"dangerous": false,
		"metrics": {
			"latency": "1.2s",
			"token_count": 150
		}
	}`

	var result CommandResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(result.Steps))
	}

	if result.Steps[0].Tool != "git" {
		t.Errorf("Expected tool 'git', got '%s'", result.Steps[0].Tool)
	}

	if result.Metrics.Latency != "1.2s" {
		t.Errorf("Expected latency '1.2s', got '%s'", result.Metrics.Latency)
	}

	if result.Metrics.TokenCount != 150 {
		t.Errorf("Expected token count 150, got %d", result.Metrics.TokenCount)
	}
}
