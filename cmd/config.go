package cmd

import (
	"fmt"
	"os"

	"github.com/kesavan-vaisakh/cmdfy/pkg/config"
	"github.com/spf13/cobra"
)

var (
	configProvider string
	configKey      string
	configURL      string
	configModel    string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage configuration for cmdfy, including LLM providers and API keys.`,
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration",
	Long:  `Set configuration values for a specific provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if configProvider == "" {
			fmt.Println("Error: --provider is required")
			os.Exit(1)
		}

		// Update or create provider config
		providerConfig, ok := cfg.Providers[configProvider]
		if !ok {
			providerConfig = config.LLMConfig{}
		}

		if configKey != "" {
			providerConfig.APIKey = configKey
		}
		if configURL != "" {
			providerConfig.BaseURL = configURL
		}
		if configModel != "" {
			providerConfig.Model = configModel
		}

		cfg.Providers[configProvider] = providerConfig
		cfg.CurrentProvider = configProvider

		if err := config.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration updated. Current provider: %s\n", configProvider)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setCmd)

	setCmd.Flags().StringVarP(&configProvider, "provider", "p", "", "LLM provider name (e.g., gemini)")
	setCmd.Flags().StringVarP(&configKey, "key", "k", "", "API key")
	setCmd.Flags().StringVarP(&configURL, "url", "u", "", "Base URL (optional)")
	setCmd.Flags().StringVarP(&configModel, "model", "m", "", "Model name (optional)")

	setCmd.MarkFlagRequired("provider")
}
