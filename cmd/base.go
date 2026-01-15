package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kesavan-vaisakh/cmdfy/pkg/config"
	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/gemini" // Register Gemini provider
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/ollama" // Register Ollama provider
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/openai" // Register OpenAI provider
	"github.com/kesavan-vaisakh/cmdfy/pkg/system"
	"github.com/spf13/cobra"
)

var (
	executeFlag  bool
	providerFlag string
)

var rootCmd = &cobra.Command{
	Use:   "cmdfy [query]",
	Short: "Cmdfy is a AI-enabled tool to generate commands",
	Long:  `Cmdfy translates natural language into shell commands using LLMs.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		// Load Config
		cfg, err := config.LoadConfig()
		if err != nil {
			// If config fails, we can't proceed unless maybe we want to allow env vars only?
			// For now, let's treat config load error as fatal but try to handle empty config gracefully if possible.
			// The LoadConfig implementation returns default if not found, so error is real failure.
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Determine provider
		providerName := cfg.CurrentProvider
		if providerFlag != "" {
			providerName = providerFlag
		}

		if providerName == "" {
			fmt.Println("No provider configured. Please run 'cmdfy config set --provider <name> --key <key>' or use --provider flag.")
			os.Exit(1)
		}

		providerConfig, ok := cfg.Providers[providerName]
		if !ok && providerFlag == "" {
			// If implicit default, but missing in map? Should not happen with well-formed config but possible manually edited.
			// If flag provided but not in config, we might want to check env vars or error.
			// For now, error out if not found.
			fmt.Printf("Provider '%s' not configured.\n", providerName)
			os.Exit(1)
		}

		// Allow overriding key from env var if empty in config?
		// Let's keep it simple: take from config.
		apiKey := providerConfig.APIKey
		// Fallback to env var if empty
		if apiKey == "" {
			envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(providerName))
			apiKey = os.Getenv(envKey)
		}

		if apiKey == "" && providerName != "ollama" {
			fmt.Printf("No API key found for provider '%s'. Please set it with 'cmdfy config set' or %s_API_KEY env var.\n", providerName, strings.ToUpper(providerName))
			os.Exit(1)
		}

		// Get Provider
		llmConfig := llm.ProviderConfig{
			APIKey:  apiKey,
			Model:   providerConfig.Model,
			BaseURL: providerConfig.BaseURL,
		}
		llmProvider, err := llm.GetProvider(providerName, llmConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing provider: %v\n", err)
			os.Exit(1)
		}

		// Context
		availableCommands, err := system.GetAvailableCommands()
		if err != nil {
			// On error, just log warning and proceed without commands?
			// Or maybe the list is just empty.
			// For now, let's just proceed with empty list to avoid blocking user.
			// potentially log to debug if we had a logger.
		}

		meta := llm.SystemMetadata{
			OS:                runtime.GOOS,
			Shell:             os.Getenv("SHELL"),
			AvailableCommands: availableCommands,
		}
		if meta.Shell == "" {
			// Fallback for Windows or unknown
			if runtime.GOOS == "windows" {
				meta.Shell = "powershell"
			} else {
				meta.Shell = "/bin/bash"
			}
		}

		// Generate
		spinner := "Generating command..."
		fmt.Fprintln(os.Stderr, spinner)
		// In a real app we'd use a nice spinner lib, print to stderr to not pollute stdout.

		result, err := llmProvider.GenerateCommand(context.Background(), query, meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating command: %v\n", err)
			os.Exit(1)
		}

		// Construct full command string
		var fullCmdBuilder strings.Builder
		for i, step := range result.Steps {
			// Quote arguments if they contain spaces
			var args []string
			for _, arg := range step.Args {
				if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") && !strings.HasPrefix(arg, "'") {
					args = append(args, fmt.Sprintf("\"%s\"", arg))
				} else {
					args = append(args, arg)
				}
			}

			fullCmdBuilder.WriteString(fmt.Sprintf("%s %s", step.Tool, strings.Join(args, " ")))

			if step.Op != "" {
				fullCmdBuilder.WriteString(fmt.Sprintf(" %s ", step.Op))
			} else if i < len(result.Steps)-1 {
				// Default to && if no op specified but more steps exist (fallback safety)
				// But ideally LLM sets the Op. Let's assume LLM is correct or strict.
				// If Op is missing, maybe it implies a separate line?
				// For now, let's just append nothing and assume valid syntax.
			}
		}
		fullCmdStr := fullCmdBuilder.String()

		if executeFlag {
			if result.Dangerous {
				fmt.Printf("[WARNING] This command is marked as dangerous: %s\n", result.Explanation)
				fmt.Print("Are you sure you want to execute it? [y/N]: ")
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" {
					fmt.Println("Aborted.")
					os.Exit(0)
				}
			}

			fmt.Printf("Executing: %s\n", fullCmdStr)

			var execCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				execCmd = exec.Command("cmd", "/C", fullCmdStr)
			} else {
				execCmd = exec.Command(meta.Shell, "-c", fullCmdStr)
			}

			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Pretty print
			fmt.Printf("\nCOMMAND: %s\n", fullCmdStr)
			fmt.Printf("\nEXPLANATION: %s\n", result.Explanation)
			if result.Dangerous {
				fmt.Printf("\n[DANGEROUS]: Yes\n")
			}
			fmt.Println()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&executeFlag, "execute", "y", false, "Execute the generated command immediately")
	rootCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Override configured provider")
}
