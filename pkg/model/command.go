package model

// CommandStep represents a single step in a command pipeline
type CommandStep struct {
	Tool string   `json:"tool"`
	Args []string `json:"args"`
	// Op is the operator connecting this step to the next (e.g., "|", "&&", ";", "||", ">", ">>")
	Op string `json:"op,omitempty"`
}

// CommandResult represents the full generated command pipeline
type CommandResult struct {
	Steps       []CommandStep `json:"steps"`
	Explanation string        `json:"explanation"`
	Dangerous   bool          `json:"dangerous"`
}
