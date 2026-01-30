package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [server-name]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a server",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime remove <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime remove lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime rm my-server"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  anime list  # List all servers"))
			fmt.Println()
			return fmt.Errorf("remove requires a server name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]

		if err := cfg.DeleteServer(serverName); err != nil {
			return err
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("✓ Removed server '%s'\n", serverName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
