package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// Version info - set via ldflags during build
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
	BuildDir  = "" // Source directory for self-updates

	versionFlag     bool
	verifyEmbedFlag bool

	rootCmd = &cobra.Command{
		Use:   "anime",
		Short: "Lambda GH200 deployment and management tool",
		Long: `anime - A beautiful CLI for managing Lambda Labs GH200 instances.

Configure servers, select installation modules, and deploy with ease.
No more shell scripts!`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ensurePipConstraint()
		},
		Run:                showAnimeWelcome,
		DisableSuggestions: false, // We'll handle suggestions ourselves with better formatting
		SuggestionsMinimumDistance: 2,
	}
)

func showAnimeWelcome(cmd *cobra.Command, args []string) {
	// Verify embedding (for build verification) - exits with status code
	if verifyEmbedFlag {
		verifyEmbedding()
		return
	}

	// Show version if flag is set
	if versionFlag {
		showVersion()
		return
	}

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
		theme.HighlightStyle.Render("anime prompt \"<natural language>\""),
		theme.DimTextStyle.Render("- Use AI to run commands (NEW!)"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime walkthrough"),
		theme.DimTextStyle.Render("- Interactive tutorial"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime query <model> \"prompt\""),
		theme.DimTextStyle.Render("- Query Ollama models"))
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
		{"🤖", "LLM Operations", "query (query Ollama models)"},
		{"🏥", "Diagnostics", "doctor (analyze installation failures)"},
		{"🖥️ ", "Server Management", "add, list, remove, status"},
		{"⚙️ ", "Configuration", "config, modules, set-modules"},
		{"🎨", "Models & Resources", "models (list downloaded AI models)"},
		{"🚀", "Deployment", "deploy, gen, sequence, templates"},
		{"📜", "Development Logs", "logs dev list, features, changelist"},
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

func showVersion() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ ANIME ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Version Information:"))
	fmt.Println()

	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Version:"),
		theme.SuccessStyle.Render(Version))

	// Show build timestamp
	buildTimeDisplay := BuildTime
	if BuildTime != "unknown" {
		// Replace underscores with spaces for better readability
		buildTimeDisplay = strings.ReplaceAll(BuildTime, "_", " ")
	}
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Built:"),
		theme.InfoStyle.Render(buildTimeDisplay))

	// Calculate and show time since build
	if BuildTime != "unknown" {
		timeSince := calculateTimeSinceBuild(BuildTime)
		if timeSince != "" {
			fmt.Printf("  %s  %s\n",
				theme.HighlightStyle.Render("Age:"),
				theme.DimTextStyle.Render(timeSince))
		}
	}

	// Show commit hash
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Commit:"),
		theme.DimTextStyle.Render(Commit))

	// Show build directory if available
	if BuildDir != "" {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render("Source:"),
			theme.DimTextStyle.Render(BuildDir))
	}

	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Lambda GH200 Deployment & Management System"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("https://github.com/joshkornreich/anime"))
	fmt.Println()
}

// verifyEmbedding checks that source code is embedded and BuildDir is set.
// Used by 'make build' to verify the binary is correctly built for distribution.
// Exits with 0 on success, 1 on failure.
func verifyEmbedding() {
	failed := false

	// Check 1: Source code must be embedded
	if !HasEmbeddedSource() {
		fmt.Println("FAIL: No embedded source code found")
		failed = true
	} else {
		size := GetEmbeddedSourceSize()
		fmt.Printf("PASS: Source code embedded (%d bytes)\n", size)
	}

	// Check 2: BuildDir must be set via ldflags
	if BuildDir == "" {
		fmt.Println("FAIL: BuildDir not set (ldflags missing)")
		failed = true
	} else {
		fmt.Printf("PASS: BuildDir set (%s)\n", BuildDir)
	}

	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}

func calculateTimeSinceBuild(buildTime string) string {
	// BuildTime format: "2025-11-21_06:04:01"
	// Parse it
	t, err := time.Parse("2006-01-02_15:04:05", buildTime)
	if err != nil {
		return ""
	}

	duration := time.Since(t)

	// Format human-readable duration
	if duration < time.Minute {
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// showCommandError displays a beautiful error message when a command fails or is not found
func showCommandError(cmd *cobra.Command, args []string, err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("💥 Command Error"))
	fmt.Println()

	// Check if it's an unknown command error
	errMsg := err.Error()
	isUnknownCmd := strings.Contains(errMsg, "unknown command") || strings.Contains(errMsg, "invalid command")

	if isUnknownCmd && len(args) > 0 {
		cmdName := args[0]

		// Check if the command matches a package name (case-insensitive)
		allPackages := installer.GetPackages()
		normalizedCmdName := strings.ToLower(cmdName)
		if pkg, exists := allPackages[normalizedCmdName]; exists {
			// This is a package name, not a command - suggest installation
			fmt.Printf("  %s %s\n",
				theme.InfoStyle.Render("'"+cmdName+"' is a package, not a command."),
				theme.DimTextStyle.Render("It may not be installed yet."))
			fmt.Println()

			fmt.Println(theme.SuccessStyle.Render("  📦 Package: " + pkg.Name))
			fmt.Println(theme.DimTextStyle.Render("  " + pkg.Description))
			fmt.Println()

			fmt.Println(theme.GlowStyle.Render("  💡 To install this package:"))
			fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime install "+pkg.ID))
			fmt.Println()

			fmt.Println(theme.DimTextStyle.Render("  Or view all packages:"))
			fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime packages"))
			fmt.Println()
			return
		}

		// Try to interpret as natural language if Ollama is available
		if isOllamaAvailable() {
			fmt.Printf("  %s %s\n",
				theme.InfoStyle.Render("Unknown command:"),
				theme.HighlightStyle.Render(cmdName))
			fmt.Println()
			fmt.Println(theme.GlowStyle.Render("  💡 Did you mean to use natural language?"))
			fmt.Println()
			fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime prompt %s", strings.Join(args, " "))))
			fmt.Println(theme.DimTextStyle.Render("    Let AI interpret your command"))
			fmt.Println()
			return
		}

		fmt.Printf("  %s %s\n",
			theme.WarningStyle.Render("Unknown command:"),
			theme.HighlightStyle.Render(cmdName))
		fmt.Println()

		// Get suggestions from Cobra and deduplicate
		suggestions := cmd.SuggestionsFor(cmdName)
		uniqueSuggestions := make([]string, 0)
		seen := make(map[string]bool)
		for _, suggestion := range suggestions {
			if !seen[suggestion] {
				seen[suggestion] = true
				uniqueSuggestions = append(uniqueSuggestions, suggestion)
			}
		}

		if len(uniqueSuggestions) > 0 {
			fmt.Println(theme.InfoStyle.Render("  💫 Did you mean one of these?"))
			fmt.Println()
			for _, suggestion := range uniqueSuggestions {
				fmt.Printf("    %s %s\n",
					theme.SuccessStyle.Render("→"),
					theme.HighlightStyle.Render("anime "+suggestion))
			}
			fmt.Println()
		}
	} else {
		fmt.Printf("  %s\n", theme.WarningStyle.Render(errMsg))
		fmt.Println()
	}

	// Show helpful commands
	fmt.Println(theme.GlowStyle.Render("  🌸 Quick Help:"))
	fmt.Println()

	helpCommands := []struct {
		cmd  string
		desc string
	}{
		{"anime tree", "View all available commands"},
		{"anime packages", "Browse available packages"},
		{"anime interactive", "Interactive package selector"},
		{"anime help", "Show detailed help"},
	}

	for _, hc := range helpCommands {
		fmt.Printf("    %s  %s\n",
			theme.HighlightStyle.Render(hc.cmd),
			theme.DimTextStyle.Render("- "+hc.desc))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Run 'anime <command> --help' for command details"))
	fmt.Println()
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		// Get the args that were attempted from os.Args
		// Skip the first arg (program name)
		args := []string{}
		if len(os.Args) > 1 {
			args = os.Args[1:]
		}
		showCommandError(rootCmd, args, err)
		// We've already displayed the error beautifully, so return it
		// but main.go should only check for error to determine exit code
		return err
	}
	return nil
}

func isOllamaAvailable() bool {
	resp, err := http.Get("http://localhost:11434")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func init() {
	// Configure error handling
	rootCmd.SilenceErrors = true // We'll handle errors ourselves
	rootCmd.SilenceUsage = true  // Don't show usage on errors

	// Set custom help function
	rootCmd.SetHelpFunc(showStylizedHelp)

	// Add version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")

	// Add hidden verify-embed flag for build verification
	rootCmd.Flags().BoolVar(&verifyEmbedFlag, "verify-embed", false, "Verify source embedding (for build verification)")
	rootCmd.Flags().MarkHidden("verify-embed")

	// Keep old commands for backwards compatibility
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	// Note: addCmd, genCmd, installNewCmd, modulesCmd, removeCmd, and sequenceCmd are registered in their respective files
}

// showStylizedHelp displays a beautiful styled help output
func showStylizedHelp(cmd *cobra.Command, args []string) {
	fmt.Println()

	// If it's the root command, show the welcome screen style help
	if cmd == rootCmd {
		showRootHelp()
		return
	}

	// For subcommands, show stylized subcommand help
	showSubcommandHelp(cmd)
}

func showRootHelp() {
	// Banner
	fmt.Println(theme.RenderBanner("⚡ ANIME CLI ⚡"))
	fmt.Println()

	// Description
	fmt.Printf("  %s\n", theme.InfoStyle.Render("Lambda GH200 Deployment & Management System"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Configure • Deploy • Manage with style"))
	fmt.Println()

	// Usage
	fmt.Printf("  %s\n", theme.GlowStyle.Render("Usage"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime [command] [flags]"))
	fmt.Println()

	// Group commands by category
	commandGroups := []struct {
		name     string
		emoji    string
		commands []string
	}{
		{"Getting Started", "🚀", []string{"walkthrough", "guide", "wizard", "tree", "help"}},
		{"Package Management", "📦", []string{"packages", "install", "interactive", "library", "explore"}},
		{"Models & AI", "🤖", []string{"models", "ollama", "llm", "query", "prompt"}},
		{"Server Management", "🖥️", []string{"add", "list", "remove", "ssh", "status", "reboot"}},
		{"Services", "⚡", []string{"start", "comfyui", "ui", "notebook", "run"}},
		{"Deployment", "🎯", []string{"deploy", "bootstrap", "push", "ship"}},
		{"Configuration", "⚙️", []string{"config", "modules", "set-modules", "set"}},
		{"Generation", "🎬", []string{"generate", "animate", "upscale", "reel"}},
		{"Workflows", "🎨", []string{"browse-workflows", "collection", "suggest"}},
		{"Diagnostics", "🏥", []string{"doctor", "jobs", "metrics", "logs", "billing"}},
		{"Updates", "📋", []string{"update", "updates"}},
	}

	fmt.Printf("  %s\n", theme.GlowStyle.Render("Commands"))
	fmt.Println()

	for _, group := range commandGroups {
		fmt.Printf("  %s %s\n", group.emoji, theme.SuccessStyle.Render(group.name))

		// Find and display commands in this group
		for _, cmdName := range group.commands {
			for _, subCmd := range rootCmd.Commands() {
				if subCmd.Name() == cmdName && !subCmd.Hidden {
					fmt.Printf("    %s  %s\n",
						theme.HighlightStyle.Render(fmt.Sprintf("%-18s", subCmd.Name())),
						theme.DimTextStyle.Render(subCmd.Short))
					break
				}
			}
		}
		fmt.Println()
	}

	// Show a few more important commands not in groups
	fmt.Printf("  %s %s\n", "📜", theme.SuccessStyle.Render("Other Commands"))
	otherCmds := []string{"aliases", "purge", "home", "develop", "completion"}
	for _, cmdName := range otherCmds {
		for _, subCmd := range rootCmd.Commands() {
			if subCmd.Name() == cmdName && !subCmd.Hidden {
				fmt.Printf("    %s  %s\n",
					theme.HighlightStyle.Render(fmt.Sprintf("%-18s", subCmd.Name())),
					theme.DimTextStyle.Render(subCmd.Short))
				break
			}
		}
	}
	fmt.Println()

	// Flags
	fmt.Printf("  %s\n", theme.GlowStyle.Render("Flags"))
	fmt.Printf("    %s  %s\n",
		theme.HighlightStyle.Render("-h, --help"),
		theme.DimTextStyle.Render("Show help for anime"))
	fmt.Printf("    %s  %s\n",
		theme.HighlightStyle.Render("-v, --version"),
		theme.DimTextStyle.Render("Show version information"))
	fmt.Println()

	// Quick tips
	fmt.Printf("  %s\n", theme.GlowStyle.Render("Quick Tips"))
	fmt.Printf("    %s %s\n", theme.SymbolStar, theme.DimTextStyle.Render("Run 'anime tree' to see all commands in a tree view"))
	fmt.Printf("    %s %s\n", theme.SymbolStar, theme.DimTextStyle.Render("Run 'anime <command> --help' for command details"))
	fmt.Printf("    %s %s\n", theme.SymbolStar, theme.DimTextStyle.Render("Run 'anime updates' to see what's new"))
	fmt.Println()
}

func showSubcommandHelp(cmd *cobra.Command) {
	// Command title
	cmdPath := cmd.CommandPath()
	fmt.Printf("  %s %s\n",
		theme.GlowStyle.Render("📌"),
		theme.TitleStyle.Render(strings.ToUpper(cmdPath)))
	fmt.Println()

	// Description
	if cmd.Long != "" {
		fmt.Printf("  %s\n", theme.InfoStyle.Render(cmd.Long))
	} else if cmd.Short != "" {
		fmt.Printf("  %s\n", theme.InfoStyle.Render(cmd.Short))
	}
	fmt.Println()

	// Usage
	fmt.Printf("  %s\n", theme.GlowStyle.Render("Usage"))
	if cmd.Runnable() {
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(cmd.UseLine()))
	}
	if cmd.HasAvailableSubCommands() {
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(cmd.CommandPath()+" [command]"))
	}
	fmt.Println()

	// Aliases
	if len(cmd.Aliases) > 0 {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Aliases"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(strings.Join(cmd.Aliases, ", ")))
		fmt.Println()
	}

	// Subcommands
	if cmd.HasAvailableSubCommands() {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Available Commands"))
		for _, subCmd := range cmd.Commands() {
			if !subCmd.Hidden {
				fmt.Printf("    %s  %s\n",
					theme.HighlightStyle.Render(fmt.Sprintf("%-18s", subCmd.Name())),
					theme.DimTextStyle.Render(subCmd.Short))
			}
		}
		fmt.Println()
	}

	// Flags
	if cmd.HasAvailableLocalFlags() {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Flags"))
		cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			shorthand := ""
			if flag.Shorthand != "" {
				shorthand = "-" + flag.Shorthand + ", "
			}
			flagStr := fmt.Sprintf("%s--%s", shorthand, flag.Name)
			if flag.Value.Type() != "bool" {
				flagStr += " " + flag.Value.Type()
			}
			fmt.Printf("    %s\n", theme.HighlightStyle.Render(flagStr))
			fmt.Printf("        %s", theme.DimTextStyle.Render(flag.Usage))
			if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "[]" {
				fmt.Printf(" %s", theme.DimTextStyle.Render("(default: "+flag.DefValue+")"))
			}
			fmt.Println()
		})
		fmt.Println()
	}

	// Global flags
	if cmd.HasAvailableInheritedFlags() {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Global Flags"))
		cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
			shorthand := ""
			if flag.Shorthand != "" {
				shorthand = "-" + flag.Shorthand + ", "
			}
			flagStr := fmt.Sprintf("%s--%s", shorthand, flag.Name)
			fmt.Printf("    %s  %s\n",
				theme.HighlightStyle.Render(flagStr),
				theme.DimTextStyle.Render(flag.Usage))
		})
		fmt.Println()
	}

	// Examples (if any)
	if cmd.Example != "" {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Examples"))
		lines := strings.Split(cmd.Example, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("    %s\n", theme.HighlightStyle.Render(line))
			}
		}
		fmt.Println()
	}

	// Footer
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Use \"anime [command] --help\" for more information about a command."))
	fmt.Println()
}
