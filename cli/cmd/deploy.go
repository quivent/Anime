package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [server-name]",
	Short: "Deploy and install modules on a Lambda server",
	Long:  `Connect to a configured Lambda server and install the selected modules.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime deploy <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime deploy lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime deploy my-server"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  anime list            # List configured servers"))
			fmt.Println(theme.DimTextStyle.Render("  anime add             # Add a new server"))
			fmt.Println(theme.DimTextStyle.Render("  anime modules <name>  # Configure modules first"))
			fmt.Println()
			return fmt.Errorf("deploy requires a server name")
		}
		return nil
	},
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
