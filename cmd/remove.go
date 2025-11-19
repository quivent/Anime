package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [server-name]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a server",
	Args:    cobra.ExactArgs(1),
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
