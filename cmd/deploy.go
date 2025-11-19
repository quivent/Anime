package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [server-name]",
	Short: "Deploy and install modules on a Lambda server",
	Long:  `Connect to a configured Lambda server and install the selected modules.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		server, err := cfg.GetServer(serverName)
		if err != nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		if len(server.Modules) == 0 {
			return fmt.Errorf("no modules configured for server %s", serverName)
		}

		m := tui.NewInstallModel(server)
		p := tea.NewProgram(m)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running installation: %w", err)
		}

		return nil
	},
}
