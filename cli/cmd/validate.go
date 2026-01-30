package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate anime configuration",
	Long: `Validates the anime configuration file (~/.config/anime/config.yaml) for errors.

Checks include:
  - Server names are unique and valid
  - Hostnames/IPs are valid
  - SSH keys exist
  - Module references are valid
  - No circular dependencies
  - API keys have valid format
  - Collections and users are properly configured`,
	Example: `  anime validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config (which automatically validates)
		cfg, err := config.Load()
		if err != nil {
			// Check if it's a validation error
			if valErr, ok := err.(*config.ValidationError); ok {
				fmt.Println("Configuration validation FAILED")
				fmt.Println()
				fmt.Println(valErr.Error())
				fmt.Println()
				return fmt.Errorf("configuration has %d error(s)", len(valErr.Errors))
			}
			// Some other error (like file not found, parsing error, etc.)
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Configuration is valid
		fmt.Println("Configuration validation PASSED")
		fmt.Println()

		// Show summary
		fmt.Println("Summary:")
		fmt.Printf("  Servers:      %d\n", len(cfg.Servers))
		fmt.Printf("  Collections:  %d\n", len(cfg.Collections))
		fmt.Printf("  Users:        %d\n", len(cfg.Users))

		// Count API keys configured
		apiKeyCount := 0
		if cfg.APIKeys.Anthropic != "" {
			apiKeyCount++
		}
		if cfg.APIKeys.OpenAI != "" {
			apiKeyCount++
		}
		if cfg.APIKeys.HuggingFace != "" {
			apiKeyCount++
		}
		if cfg.APIKeys.LambdaLabs != "" {
			apiKeyCount++
		}
		fmt.Printf("  API Keys:     %d configured\n", apiKeyCount)

		if apiKeyCount > 0 {
			fmt.Print("    (")
			keys := []string{}
			if cfg.APIKeys.Anthropic != "" {
				keys = append(keys, "Anthropic")
			}
			if cfg.APIKeys.OpenAI != "" {
				keys = append(keys, "OpenAI")
			}
			if cfg.APIKeys.HuggingFace != "" {
				keys = append(keys, "HuggingFace")
			}
			if cfg.APIKeys.LambdaLabs != "" {
				keys = append(keys, "Lambda Labs")
			}
			for i, key := range keys {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(key)
			}
			fmt.Println(")")
		}

		if len(cfg.Servers) > 0 {
			fmt.Println()
			fmt.Println("Servers:")
			for _, server := range cfg.Servers {
				fmt.Printf("  - %s (%s@%s) - %d modules\n",
					server.Name, server.User, server.Host, len(server.Modules))
			}
		}

		if cfg.ActiveUser != "" {
			fmt.Printf("\nActive User:  %s\n", cfg.ActiveUser)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
