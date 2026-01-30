package cmd

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	configRemote bool
	configServer string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Lambda servers and installation modules",
	Long:  `Interactive TUI for configuring Lambda GH200 servers, selecting installation modules, and managing API keys.

Run locally:
  anime config

Run on remote server:
  anime config --remote --server lambda
  anime config -r -s 192.168.1.100`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVarP(&configRemote, "remote", "r", false, "Run config on remote server")
	configCmd.Flags().StringVarP(&configServer, "server", "s", "lambda", "Server to configure (default: lambda)")
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Run on remote server
	if configRemote {
		return runRemoteConfig()
	}

	// Run locally
	m, err := tui.NewConfigModel()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

func runRemoteConfig() error {
	// Parse server target
	target, err := parseServerTarget(configServer)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔧 Running config on remote server"))
	fmt.Println()
	fmt.Printf("  Target: %s\n", theme.HighlightStyle.Render(target))
	fmt.Println()

	// Check if anime is installed on remote server
	fmt.Print(theme.DimTextStyle.Render("▶ Checking anime installation... "))
	checkCmd := exec.Command("ssh", target, "which anime")
	if err := checkCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  anime not found on server"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Push anime first:"))
		fmt.Printf("   %s\n", theme.HighlightStyle.Render("anime push "+configServer))
		fmt.Println()
		return fmt.Errorf("anime not installed on server")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Run config on remote server with SSH forwarding for interactive terminal
	fmt.Print(theme.DimTextStyle.Render("▶ Starting remote config... "))
	fmt.Println()
	fmt.Println()

	sshCmd := exec.Command("ssh", "-t", target, "anime config")
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Remote config failed"))
		fmt.Println()
		return fmt.Errorf("failed to run remote config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Remote config complete!"))
	fmt.Println()

	return nil
}
