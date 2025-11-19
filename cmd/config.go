package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Lambda servers and installation modules",
	Long:  `Interactive TUI for configuring Lambda GH200 servers, selecting installation modules, and managing API keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := tui.NewConfigModel()
		if err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		return nil
	},
}
