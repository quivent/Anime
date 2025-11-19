package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show anime CLI command tree",
	Long:  "Display a beautiful tree visualization of all available anime commands",
	Run:   runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
}

func runTree(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("🌸 ANIME COMMAND TREE 🌸"))
	fmt.Println()

	commands := []struct {
		name string
		desc string
		subs []struct {
			name string
			desc string
		}
	}{
		{
			name: "📦 Package Management",
			desc: "Install and manage software packages",
			subs: []struct{ name, desc string }{
				{"packages", "List all available packages"},
				{"packages --tree", "Show dependency tree"},
				{"install <id>", "Install packages with dependencies"},
				{"install --dry-run", "Preview installation plan"},
				{"install --phased", "Install with confirmations"},
				{"install --remote -s <server>", "Remote installation via SSH"},
				{"interactive", "Interactive package selection TUI"},
			},
		},
		{
			name: "🖥️  Server Management",
			desc: "Configure and manage Lambda servers",
			subs: []struct{ name, desc string }{
				{"add <name> <host>", "Add new server"},
				{"list", "List all servers"},
				{"remove <name>", "Remove server"},
				{"status", "Check server status"},
			},
		},
		{
			name: "⚙️  Configuration",
			desc: "System and deployment configuration",
			subs: []struct{ name, desc string }{
				{"config", "Open config TUI"},
				{"modules", "Select modules"},
				{"set-modules", "Set modules via CLI"},
				{"list-modules", "List available modules"},
			},
		},
		{
			name: "🚀 Deployment",
			desc: "Deploy and execute on servers",
			subs: []struct{ name, desc string }{
				{"deploy", "Deploy to server"},
				{"gen", "Generate bash commands"},
				{"sequence", "Show command sequence"},
				{"templates", "Show command templates"},
			},
		},
		{
			name: "🎯 Navigation",
			desc: "Help and utilities",
			subs: []struct{ name, desc string }{
				{"tree", "Show this command tree"},
				{"help", "Get help on commands"},
				{"completion", "Generate shell completion"},
			},
		},
	}

	for i, group := range commands {
		isLast := i == len(commands)-1

		// Group header
		fmt.Println(theme.GlowStyle.Render(group.name))
		fmt.Println(theme.DimTextStyle.Render("  " + group.desc))
		fmt.Println()

		// Commands in group
		for j, subcmd := range group.subs {
			marker := theme.SymbolBranch
			pipe := theme.SymbolPipe
			if j == len(group.subs)-1 {
				marker = theme.SymbolLastBranch
				if isLast {
					pipe = " "
				}
			}

			fmt.Printf("%s %s %s\n",
				theme.InfoStyle.Render(marker),
				theme.HighlightStyle.Render("anime "+subcmd.name),
				theme.DimTextStyle.Render("- "+subcmd.desc))

			if j < len(group.subs)-1 || !isLast {
				fmt.Println(theme.DimTextStyle.Render(pipe))
			}
		}

		if !isLast {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("✨ Quick Start:"))
	fmt.Println(theme.DimTextStyle.Render("  1. anime packages          - Browse available packages"))
	fmt.Println(theme.DimTextStyle.Render("  2. anime interactive       - Select packages interactively"))
	fmt.Println(theme.DimTextStyle.Render("  3. anime install <id>      - Install selected packages"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("For detailed help: anime <command> --help"))
}
