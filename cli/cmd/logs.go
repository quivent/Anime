package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View and manage anime CLI logs",
	Long: `View and manage various anime CLI logs.

Available subcommand categories:
  development - Track development cycles, features, and changes`,
	Run: showLogsHelp,
}

func showLogsHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📜 ANIME LOGS 📜"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Log Management & Development Tracking"))
	fmt.Println()

	fmt.Println(theme.GlowStyle.Render("🌸 Available Log Categories:"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HeaderStyle.Render("development"))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render("Track development cycles, features, and changes"))
	fmt.Println()

	subcommands := []struct {
		cmd  string
		desc string
	}{
		{"anime logs development list", "List all development cycles"},
		{"anime logs development last", "Show last development activity"},
		{"anime logs development features", "List developed features"},
		{"anime logs development changelist", "View change history"},
	}

	fmt.Println(theme.GlowStyle.Render("✨ Quick Commands:"))
	fmt.Println()

	for _, sc := range subcommands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(sc.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(sc.desc))
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
