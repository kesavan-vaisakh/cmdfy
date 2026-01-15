package system

import (
	"os"
	"sort"
	"strings"
)

// GetAvailableCommands returns a sorted list of unique executable commands found in PATH
func GetAvailableCommands() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	commandsMap := make(map[string]bool)

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip unreadable directories
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Check if executable (simple check)
			if info.Mode()&0111 != 0 {
				commandsMap[entry.Name()] = true
			}
		}
	}

	commands := make([]string, 0, len(commandsMap))
	for cmd := range commandsMap {
		commands = append(commands, cmd)
	}

	sort.Strings(commands)
	return commands, nil
}
