package model

// CommandStep represents a single step in a command pipeline
type CommandStep struct {
	Tool string   `json:"tool"`
	Args []string `json:"args"`
	// Op is the operator connecting this step to the next (e.g., "|", "&&", ";", "||", ">", ">>")
	Op string `json:"op,omitempty"`
}

// Metrics holds performance and cost metrics for the generation
type Metrics struct {
	Latency      string `json:"latency"` // e.g., "1.2s"
	TokenCount   int    `json:"token_count,omitempty"`
	CostEstimate string `json:"cost_estimate,omitempty"` // Approximation if possible
}

// CommandResult represents the full generated command pipeline
type CommandResult struct {
	Steps       []CommandStep `json:"steps"`
	Explanation string        `json:"explanation"`
	Dangerous   bool          `json:"dangerous"`
	Metrics     Metrics       `json:"metrics,omitempty"`
}
