package system

import (
	"fmt"
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

// GetFileContext returns a list of visible files and directories in the given path.
// It limits the result to 50 items and ignores hidden files or common ignored dirs.
func GetFileContext(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
	}

	var files []string
	ignored := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"dist":         true,
		"build":        true,
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if ignored[name] {
			continue
		}

		if entry.IsDir() {
			files = append(files, name+"/")
		} else {
			files = append(files, name)
		}

		if len(files) >= 50 {
			break
		}
	}
	return files, nil
}
