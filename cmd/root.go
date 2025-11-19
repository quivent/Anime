package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anime",
	Short: "Lambda GH200 deployment and management tool",
	Long: `anime - A beautiful CLI for managing Lambda Labs GH200 instances.

Configure servers, select installation modules, and deploy with ease.
No more shell scripts!`,
	Run: showAnimeWelcome,
}

func showAnimeWelcome(cmd *cobra.Command, args []string) {
	// Anime-style welcome screen
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ ANIME ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Lambda GH200 Deployment & Management System"))
	fmt.Println(theme.DimTextStyle.Render("  Configure • Deploy • Manage with style"))
	fmt.Println()

	// Quick actions
	fmt.Println(theme.GlowStyle.Render("🌸 Quick Actions:"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime tree"),
		theme.DimTextStyle.Render("- View all commands"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime packages"),
		theme.DimTextStyle.Render("- Browse available packages"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime interactive"),
		theme.DimTextStyle.Render("- Interactive package selector"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime install <id>"),
		theme.DimTextStyle.Render("- Install packages"))
	fmt.Println()

	// Categories
	fmt.Println(theme.InfoStyle.Render("✨ Command Categories:"))
	fmt.Println()

	categories := []struct {
		emoji string
		name  string
		desc  string
	}{
		{"📦", "Package Management", "install, packages, interactive"},
		{"🖥️ ", "Server Management", "add, list, remove, status"},
		{"⚙️ ", "Configuration", "config, modules, set-modules"},
		{"🚀", "Deployment", "deploy, gen, sequence, templates"},
		{"🎯", "Navigation", "tree, help, completion"},
	}

	for _, cat := range categories {
		fmt.Printf("  %s %s  %s\n",
			cat.emoji,
			theme.SuccessStyle.Render(cat.name),
			theme.DimTextStyle.Render("→ "+cat.desc))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Run 'anime tree' for full command tree"))
	fmt.Println(theme.DimTextStyle.Render("  Run 'anime <command> --help' for command details"))
	fmt.Println()
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Keep old commands for backwards compatibility
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(modulesCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(sequenceCmd)
}
