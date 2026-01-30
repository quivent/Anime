package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install anime CLI",
	Long:  `Build and install the anime CLI to your system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Building anime...")
		fmt.Println("Run: go install")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Servers) == 0 {
			fmt.Println("No servers configured. Run 'anime config' to add one.")
			return nil
		}

		fmt.Println("Configured servers:")
		fmt.Println()
		for _, server := range cfg.Servers {
			fmt.Printf("  %s (%s@%s)\n", server.Name, server.User, server.Host)
			fmt.Printf("    Cost: $%.2f/hr\n", server.CostPerHour)
			fmt.Printf("    Modules: %d configured\n", len(server.Modules))
			if len(server.Modules) > 0 {
				cost := config.EstimateCost(server.Modules, server.CostPerHour)
				fmt.Printf("    Estimated deployment: $%.2f\n", cost)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
