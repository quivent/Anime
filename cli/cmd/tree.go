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
				{"packages status", "Show installation status"},
				{"packages --tree", "Show dependency tree"},
				{"install <id>", "Install packages with dependencies"},
				{"install --dry-run", "Preview installation plan"},
				{"install --phased", "Install with confirmations"},
				{"install --remote -s <server>", "Remote installation via SSH"},
				{"interactive", "Interactive package selection TUI"},
				{"parallelize", "Smart parallel installation planning"},
			},
		},
		{
			name: "💡 Recommendations",
			desc: "Get personalized workflow and package suggestions",
			subs: []struct{ name, desc string }{
				{"suggest", "Analyze setup and suggest workflows/packages"},
				{"wizard", "Interactive setup wizard for node configuration"},
				{"guide", "Comprehensive usage documentation"},
			},
		},
		{
			name: "🎬 Assets & Workflows",
			desc: "Manage collections and run AI workflows",
			subs: []struct{ name, desc string }{
				{"collection create <name> <path>", "Create asset collection"},
				{"collection list", "List all collections"},
				{"collection info <name>", "Show collection details"},
				{"workflow", "Browse and run workflows"},
				{"workstation", "Interactive workstation dashboard"},
			},
		},
		{
			name: "☁️  Lambda Cloud",
			desc: "Manage Lambda Labs GPU instances",
			subs: []struct{ name, desc string }{
				{"lambda list", "List Lambda instances"},
				{"lambda launch", "Launch new instance"},
				{"lambda ssh <instance>", "SSH into instance"},
				{"lambda terminate <instance>", "Terminate instance"},
				{"lambda defaults", "Show default packages"},
				{"metrics", "Real-time GPU metrics and cost tracking"},
			},
		},
		{
			name: "🖥️  Server Management",
			desc: "Configure and manage remote servers",
			subs: []struct{ name, desc string }{
				{"add <name> <host>", "Add new server"},
				{"list", "List all servers"},
				{"remove <name>", "Remove server"},
				{"status", "Check server status"},
				{"set lambda <server>", "Set default Lambda server"},
				{"explore <server>", "Discover untracked models on server"},
				{"push", "Deploy CLI to remote servers"},
			},
		},
		{
			name: "⚙️  Configuration",
			desc: "System and deployment configuration",
			subs: []struct{ name, desc string }{
				{"config", "Open config TUI"},
				{"modules", "Select modules"},
				{"set-modules", "Set modules via CLI"},
				{"bootstrap", "Bootstrap configuration"},
			},
		},
		{
			name: "🚀 Deployment",
			desc: "Deploy and execute on servers",
			subs: []struct{ name, desc string }{
				{"deploy", "Deploy to server"},
				{"ship <source> <dest>", "Tar, rsync, and untar files to remote"},
				{"gen", "Generate bash commands"},
				{"sequence", "Show command sequence"},
			},
		},
		{
			name: "🛠️  Development",
			desc: "Developer tools and source access",
			subs: []struct{ name, desc string }{
				{"home", "Navigate to anime source directory"},
				{"develop", "Launch Claude Code in anime's source directory"},
			},
		},
		{
			name: "🎯 Help & Utilities",
			desc: "Navigation and assistance",
			subs: []struct{ name, desc string }{
				{"tree", "Show this command tree"},
				{"help [command]", "Get help on commands"},
				{"version", "Show version information"},
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
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 What to do next:"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages"))
	fmt.Println(theme.DimTextStyle.Render("    Browse all available packages"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime interactive"))
	fmt.Println(theme.DimTextStyle.Render("    Launch interactive package selector"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime workstation"))
	fmt.Println(theme.DimTextStyle.Render("    Monitor GPU, models, and system resources"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime <command> --help"))
	fmt.Println(theme.DimTextStyle.Render("    Get detailed help for any command"))
	fmt.Println()
}
