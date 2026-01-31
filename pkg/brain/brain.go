package brain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BrainEntry represents a single record in the brain
type BrainEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Query       string    `json:"query"`
	Context     string    `json:"context,omitempty"` // Captured stdin / error logs
	Command     string    `json:"command"`
	Explanation string    `json:"explanation"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
}

// Brain handles the persistence and retrieval of command history
type Brain struct {
	filePath string
}

// NewBrain creates a new Brain instance
func NewBrain() (*Brain, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	brainDir := filepath.Join(home, ".cmdfy")
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	return &Brain{
		filePath: filepath.Join(brainDir, "brain.jsonl"),
	}, nil
}

// Record saves a selected command to the brain
func (b *Brain) Record(entry BrainEntry) error {
	entry.Timestamp = time.Now()

	// Validate essential fields
	if entry.Query == "" || entry.Command == "" {
		return fmt.Errorf("query and command are required")
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	f, err := os.OpenFile(b.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open brain file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("failed to write to brain file: %w", err)
	}

	return nil
}

// GetExamples retrieves relevant examples from the brain
// For now, it returns the most recent entries.
// In the future, we can implement semantic search or fuzzy matching.
func (b *Brain) GetExamples(query string, limit int) ([]BrainEntry, error) {
	f, err := os.Open(b.filePath)
	if os.IsNotExist(err) {
		return []BrainEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open brain file: %w", err)
	}
	defer f.Close()

	var entries []BrainEntry
	scanner := bufio.NewScanner(f)

	// Read all lines (inefficient for large files, but fine for < 10MB)
	// Improved: Read simply all for now, optimize later.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry BrainEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip malformed lines, don't crash
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading brain file: %w", err)
	}

	// Reverse to get most recent first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	// Filter or rank?
	// For simplicity in this phase, we just return the most recent `limit` entries.
	// We can add simple keyword matching later.
	if len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}
