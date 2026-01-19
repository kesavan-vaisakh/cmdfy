package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"sync"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kesavan-vaisakh/cmdfy/app/tui"
	"github.com/kesavan-vaisakh/cmdfy/pkg/config"
	"github.com/kesavan-vaisakh/cmdfy/pkg/llm"
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/anthropic" // Register Anthropic provider
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/gemini"    // Register Gemini provider
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/ollama"    // Register Ollama provider
	_ "github.com/kesavan-vaisakh/cmdfy/pkg/llm/openai"    // Register OpenAI provider
	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
	"github.com/kesavan-vaisakh/cmdfy/pkg/system"
)

var (
	executeFlag   bool
	configFile    string
	providerFlag  string
	clipboardFlag bool
	directoryFlag string
	compareFlag   bool
)

var rootCmd = &cobra.Command{
	Use:   "cmdfy [query]",
	Short: "Cmdfy is a AI-enabled tool to generate commands",
	Long:  `Cmdfy translates natural language into shell commands using LLMs.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		if clipboardFlag {
			content, err := clipboard.ReadAll()
			if err == nil && content != "" {
				query = fmt.Sprintf("%s\n\nContext from Clipboard:\n%s", query, content)
				fmt.Println("📋 Added clipboard content to context.")
			} else if err != nil {
				fmt.Printf("⚠️  Failed to read clipboard: %v\n", err)
			}
		}

		// Load Config
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Context
		// Gather system metadata
		commands, _ := system.GetAvailableCommands()
		files, _ := system.GetFileContext(directoryFlag)

		meta := llm.SystemMetadata{
			OS:                runtime.GOOS,
			Shell:             os.Getenv("SHELL"),
			AvailableCommands: commands,
			CurrentDirFiles:   files,
		}
		if meta.Shell == "" {
			if runtime.GOOS == "windows" {
				meta.Shell = "powershell"
			} else {
				meta.Shell = "/bin/bash"
			}
		}

		if compareFlag {
			runComparison(query, meta, cfg)
			return
		}

		// Single Provider Flow

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
			fmt.Printf("Provider '%s' not configured.\n", providerName)
			os.Exit(1)
		}

		apiKey := providerConfig.APIKey
		if apiKey == "" {
			envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(providerName))
			apiKey = os.Getenv(envKey)
		}

		if apiKey == "" && providerName != "ollama" {
			fmt.Printf("No API key found for provider '%s'. Please set it with 'cmdfy config set' or %s_API_KEY env var.\n", providerName, strings.ToUpper(providerName))
			os.Exit(1)
		}

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

		// Generate
		spinner := "Generating command..."
		fmt.Fprintln(os.Stderr, spinner)

		result, err := llmProvider.GenerateCommand(context.Background(), query, meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating command: %v\n", err)
			os.Exit(1)
		}

		printAndExecute(result, meta)
	},
}

func runComparison(query string, meta llm.SystemMetadata, cfg *config.Config) {
	fmt.Println("Running benchmark across configured providers...")

	var wg sync.WaitGroup
	resultsChan := make(chan tui.ProviderResult, len(cfg.Providers))

	for name, pCfg := range cfg.Providers {
		// Skip if no API key and not local (simplification)
		apiKey := pCfg.APIKey
		if apiKey == "" {
			envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(name))
			apiKey = os.Getenv(envKey)
		}
		if apiKey == "" && name != "ollama" {
			continue // Skip unconfigured providers
		}

		wg.Add(1)
		go func(pName string, pConfig config.LLMConfig, key string) {
			defer wg.Done()

			llmConfig := llm.ProviderConfig{
				APIKey:  key,
				Model:   pConfig.Model,
				BaseURL: pConfig.BaseURL,
			}
			provider, err := llm.GetProvider(pName, llmConfig)
			if err != nil {
				resultsChan <- tui.ProviderResult{Name: pName, Error: err}
				return
			}

			// Timeout for benchmark
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			res, err := provider.GenerateCommand(ctx, query, meta)
			resultsChan <- tui.ProviderResult{Name: pName, Result: res, Error: err}

		}(name, pCfg, apiKey)
	}

	wg.Wait()
	close(resultsChan)

	var results []tui.ProviderResult
	for r := range resultsChan {
		results = append(results, r)
	}

	if len(results) == 0 {
		fmt.Println("No valid providers found to benchmark.")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.InitialModel(results))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(tui.Model); ok && finalModel.Choice != nil {
		// User selected a result
		fmt.Printf("\n🏆 Selected result from %s\n", strings.ToUpper(finalModel.Choice.Name))
		printAndExecute(finalModel.Choice.Result, meta)
	}
}

func printAndExecute(result *model.CommandResult, meta llm.SystemMetadata) {
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
			// Default to &&
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
		if result.Metrics.Latency != "" {
			fmt.Printf("\nMETRICS: %s", result.Metrics.Latency)
			if result.Metrics.TokenCount > 0 {
				fmt.Printf(", %d tokens", result.Metrics.TokenCount)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&executeFlag, "execute", "y", false, "Execute the generated command immediately")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.cmdfy/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "", "Override the LLM provider (e.g., gemini, ollama)")
	rootCmd.PersistentFlags().BoolVarP(&clipboardFlag, "clipboard", "c", false, "Include clipboard content as context")
	rootCmd.PersistentFlags().StringVarP(&directoryFlag, "directory", "d", ".", "Target directory for context scanning")
	rootCmd.PersistentFlags().BoolVar(&compareFlag, "compare", false, "Benchmark all configured providers")

	rootCmd.AddCommand(configCmd)
}
