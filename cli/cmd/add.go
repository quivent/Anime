package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/spf13/cobra"
)

var (
	serverName string
	serverHost string
	serverUser string
	sshKey     string
	costPerHour float64
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new Lambda server",
	Long:  `Add a new Lambda server with simple flags. No TUI required.`,
	Example: `  anime add --name lambda-1 --host 192.168.1.100 --user ubuntu --key ~/.ssh/lambda.pem --cost 20
  anime add -n my-server -H 10.0.0.5 -u ubuntu -k ~/.ssh/id_rsa -c 18.50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if serverName == "" || serverHost == "" {
			return fmt.Errorf("--name and --host are required")
		}

		// Set defaults
		if serverUser == "" {
			serverUser = "ubuntu"
		}
		if sshKey == "" {
			sshKey = findDefaultSSHKey()
		}
		if costPerHour == 0 {
			costPerHour = 20.0
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		server := config.Server{
			Name:        serverName,
			Host:        serverHost,
			User:        serverUser,
			SSHKey:      sshKey,
			CostPerHour: costPerHour,
			Modules:     []string{},
		}

		// Validate server before adding
		if err := server.Validate(); err != nil {
			fmt.Println()
			fmt.Println("Server validation failed:")
			if valErr, ok := err.(*config.ValidationError); ok {
				for _, e := range valErr.Errors {
					fmt.Printf("  - %s\n", e)
				}
			} else {
				fmt.Printf("  - %s\n", err.Error())
			}
			fmt.Println()
			return fmt.Errorf("invalid server configuration")
		}

		cfg.AddServer(server)

		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("✓ Added server '%s'\n", serverName)
		fmt.Printf("  Host: %s@%s\n", serverUser, serverHost)
		fmt.Printf("  SSH Key: %s\n", sshKey)
		fmt.Printf("  Cost: $%.2f/hr\n\n", costPerHour)
		fmt.Println("Next steps:")
		fmt.Printf("  anime push %s                     # Push anime CLI to server\n", serverName)
		fmt.Printf("  anime modules %s                  # Select modules interactively\n", serverName)
		fmt.Printf("  anime set-modules %s core pytorch # Or set modules via CLI\n", serverName)
		fmt.Printf("  anime deploy %s                   # Deploy when ready\n", serverName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&serverName, "name", "n", "", "Server name (required)")
	addCmd.Flags().StringVarP(&serverHost, "host", "H", "", "Server IP or hostname (required)")
	addCmd.Flags().StringVarP(&serverUser, "user", "u", "ubuntu", "SSH user (default: ubuntu)")
	addCmd.Flags().StringVarP(&sshKey, "key", "k", "", "SSH private key path (auto-detects ed25519/rsa)")
	addCmd.Flags().Float64VarP(&costPerHour, "cost", "c", 20.0, "Cost per hour in USD")
}

// findDefaultSSHKey finds the first available SSH key
func findDefaultSSHKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.ssh/id_ed25519"
	}

	// Check in order of preference: ed25519, rsa, ecdsa
	keys := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}

	for _, key := range keys {
		if _, err := os.Stat(key); err == nil {
			return key
		}
	}

	// Default to ed25519 path even if not found
	return "~/.ssh/id_ed25519"
}
