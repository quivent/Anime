package cmd

import (
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/errors"
	"github.com/joshkornreich/anime/internal/hf"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/logger"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/muesli/termenv"
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

	// Logging flags
	logLevel string
	logFile  string
	debug    bool

	// Color flag — drives lipgloss color profile selection in initColor.
	colorMode string

	// Global SSH security flags
	SSHInsecure               bool
	SSHStrictHostKeyChecking  bool
	SSHNonInteractive         bool

	rootCmd = &cobra.Command{
		Use:   "anime",
		Short: "Lambda GH200 deployment and management tool",
		Long: `anime - A beautiful CLI for managing Lambda Labs GH200 instances.

Configure servers, select installation modules, and deploy with ease.
No more shell scripts!`,
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

	// Count embedded assets
	agentCount := countEmbeddedAgents()
	hasSource := HasEmbeddedSource()
	sourceSize := GetEmbeddedSourceSize()

	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ ANIME ⚡"))
	fmt.Println()

	// Tagline
	fmt.Println(theme.InfoStyle.Render("  Self-Contained AI Development & Deployment System"))
	fmt.Println()

	// Section: What's embedded
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.GlowStyle.Render("  📦 EMBEDDED IN THIS BINARY"))
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Agents
	fmt.Printf("  %s %s\n",
		theme.SuccessStyle.Render(fmt.Sprintf("🤖 %d Claude Code Agents", agentCount)),
		theme.DimTextStyle.Render("— architect, developer, researcher, planner..."))
	fmt.Printf("     %s\n", theme.DimTextStyle.Render("Push to any machine: ")+theme.HighlightStyle.Render("anime claude agents push"))
	fmt.Println()

	// Source
	if hasSource {
		sizeMB := float64(sourceSize) / (1024 * 1024)
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(fmt.Sprintf("📜 Full Source Code (%.1fMB)", sizeMB)),
			theme.DimTextStyle.Render("— self-updating, self-deploying"))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render("Extract anywhere: ")+theme.HighlightStyle.Render("anime extract --embedded"))
	} else {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  📜 Source: not embedded (use 'make build' to include)"))
	}
	fmt.Println()

	// Tooling
	fmt.Printf("  %s %s\n",
		theme.SuccessStyle.Render("🎬 Sky + Reel"),
		theme.DimTextStyle.Render("— video generation & processing pipelines"))
	fmt.Printf("  %s %s\n",
		theme.SuccessStyle.Render("🔧 100+ Packages"),
		theme.DimTextStyle.Render("— ComfyUI, Ollama, CUDA, Python envs..."))
	fmt.Println()

	fmt.Printf("  %s %s\n",
		theme.InfoStyle.Render("Browse everything:"),
		theme.HighlightStyle.Render("anime contents"))
	fmt.Println()

	// Section: Capabilities
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.GlowStyle.Render("  ⚡ CAPABILITIES"))
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	capabilities := []struct {
		icon    string
		name    string
		desc    string
		cmd     string
	}{
		{"🤖", "Claude Code Integration", "65 specialized AI agents for any task", "anime claude agents push"},
		{"☁️ ", "Lambda Cloud", "Launch, SSH, manage GH200 GPU instances", "anime lambda list"},
		{"📦", "Package Management", "Install AI/ML stack with dependencies", "anime packages"},
		{"🎨", "AI Models", "Browse & install LLMs, SD, video models", "anime models"},
		{"🎬", "Video Generation", "Reel workflows for AI video pipelines", "anime reel"},
		{"🔄", "Source Sync", "Rsync-based code deployment", "anime source push"},
		{"🚀", "Remote Deploy", "Push CLI + agents to any server", "anime push"},
	}

	for _, cap := range capabilities {
		fmt.Printf("  %s %s\n", cap.icon, theme.SuccessStyle.Render(cap.name))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(cap.desc))
		fmt.Printf("     %s\n", theme.HighlightStyle.Render(cap.cmd))
		fmt.Println()
	}

	// Section: Quick Start
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.GlowStyle.Render("  🚀 QUICK START"))
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	quickStart := []struct {
		cmd  string
		desc string
	}{
		{"anime claude agents push", "Deploy AI agents to Claude Code"},
		{"anime interactive", "TUI package selector"},
		{"anime extract --embedded", "Extract source code anywhere"},
		{"anime push <server>", "Deploy CLI to remote server"},
		{"anime contents", "Browse everything embedded"},
		{"anime packages", "List installable packages"},
		{"anime models", "Browse AI models catalog"},
		{"anime tree", "Full command tree"},
	}

	for _, qs := range quickStart {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(qs.cmd))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(qs.desc))
	}
	fmt.Println()

	// Footer
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────────────"))
	fmt.Printf("  %s %s    %s %s\n",
		theme.DimTextStyle.Render("Version:"),
		theme.InfoStyle.Render(Version),
		theme.DimTextStyle.Render("Help:"),
		theme.HighlightStyle.Render("anime <command> --help"))
	fmt.Println()
}

// countEmbeddedAgents returns the number of agents embedded in the binary
func countEmbeddedAgents() int {
	entries, err := embeddedAgents.ReadDir("embedded/agents")
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			count++
		}
	}
	return count
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
		// Check if it's a DetailedError
		var detailedErr *errors.DetailedError
		if stderrors.As(err, &detailedErr) {
			showDetailedError(detailedErr)
			return err
		}

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
	// Initialize logger and color profile on startup, in that order.
	cobra.OnInitialize(initLogger, initColor)

	// Always set HF_TOKEN from embedded value if not already set
	if os.Getenv("HF_TOKEN") == "" {
		os.Setenv("HF_TOKEN", hf.GetToken())
	}

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

	// Add logging flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Log level (debug|info|warn|error)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path (default: stderr)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging (shorthand for --log-level=debug)")

	// Color control: by default we auto-detect (TTY → colors, pipe → no
	// colors). --color=always force-enables; --color=never strips. Also
	// respects the standard env vars NO_COLOR / FORCE_COLOR / CLICOLOR_FORCE.
	rootCmd.PersistentFlags().StringVar(&colorMode, "color", "auto", "Color output (auto|always|never). Also: FORCE_COLOR=1 / NO_COLOR")

	// Add SSH security flags
	rootCmd.PersistentFlags().BoolVar(&SSHInsecure, "insecure", false, "Disable SSH host key verification (INSECURE - not recommended)")
	rootCmd.PersistentFlags().BoolVar(&SSHStrictHostKeyChecking, "strict-host-key-checking", true, "Enable strict SSH host key checking (default: true)")
	rootCmd.PersistentFlags().BoolVar(&SSHNonInteractive, "non-interactive", false, "Non-interactive mode - fail on unknown host keys instead of prompting")

	// Keep old commands for backwards compatibility
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	// Note: addCmd, genCmd, installNewCmd, modulesCmd, removeCmd, and sequenceCmd are registered in their respective files
}

// initColor decides the lipgloss color profile based on (in priority order):
//
//  1. --color=always|never (explicit user choice)
//  2. NO_COLOR env (any value → strip)
//  3. FORCE_COLOR / CLICOLOR_FORCE env (any value → force)
//  4. Auto: lipgloss/termenv default — colors when stdout is a TTY, none otherwise
//
// The default lipgloss behavior (auto) is what most users expect, but it's
// easy to land in a non-TTY situation accidentally (screen, tmux without
// xterm-256color, captured output) and lose colors entirely. Setting
// FORCE_COLOR=1 or `--color=always` rescues those cases.
func initColor() {
	mode := strings.ToLower(strings.TrimSpace(colorMode))
	if mode == "" {
		mode = "auto"
	}
	// Subcommands with DisableFlagParsing=true (e.g., anime wan render) skip
	// ALL flag parsing including persistent flags, so the cobra-bound
	// `colorMode` stays at "auto" no matter what the user typed. Fall back
	// to scanning os.Args directly so --color works there too.
	if mode == "auto" {
		for i, a := range os.Args {
			switch {
			case strings.HasPrefix(a, "--color="):
				mode = strings.ToLower(strings.TrimPrefix(a, "--color="))
			case a == "--color" && i+1 < len(os.Args):
				mode = strings.ToLower(os.Args[i+1])
			}
		}
	}
	if mode == "auto" {
		if os.Getenv("NO_COLOR") != "" {
			mode = "never"
		} else if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			mode = "always"
		}
	}
	switch mode {
	case "always", "force", "1", "true", "yes":
		lipgloss.SetColorProfile(termenv.TrueColor)
		// Propagate to child processes (wan.py, npm, screen-launched ComfyUI)
		// that have their own color logic. Set both env vars so anything
		// following either convention picks it up.
		if os.Getenv("FORCE_COLOR") == "" {
			os.Setenv("FORCE_COLOR", "1")
		}
		os.Unsetenv("NO_COLOR") // an explicit FORCE wins over a stale NO_COLOR
	case "never", "off", "0", "false", "no":
		lipgloss.SetColorProfile(termenv.Ascii)
		if os.Getenv("NO_COLOR") == "" {
			os.Setenv("NO_COLOR", "1")
		}
		os.Unsetenv("FORCE_COLOR")
	}
	// "auto" with no overriding env: leave the lipgloss default + child env alone.
}

// stripRootFlags removes flags that belong to rootCmd from `args` so they
// don't leak into downstream argparse-based tools (most notably wan.py via
// the DisableFlagParsing=true subcommands). Applied effects (like --color)
// have already been picked up by initColor's os.Args scan, so it's safe to
// drop them here.
func stripRootFlags(args []string) []string {
	known := map[string]bool{
		"--color": true, "--debug": true,
		"--log-level": true, "--log-file": true,
		"--insecure": true, "--strict-host-key-checking": true, "--non-interactive": true,
	}
	// Bool flags don't consume the next arg as a value.
	bools := map[string]bool{
		"--debug": true, "--insecure": true,
		"--strict-host-key-checking": true, "--non-interactive": true,
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if eq := strings.IndexByte(a, '='); eq > 0 {
			if known[a[:eq]] {
				continue
			}
		}
		if known[a] {
			if !bools[a] && i+1 < len(args) {
				i++ // consume value
			}
			continue
		}
		out = append(out, a)
	}
	return out
}

// initLogger initializes the global logger based on flags
func initLogger() {
	// If debug flag is set, override log level
	if debug {
		logLevel = "debug"
	}

	// Parse log level
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelError
	}

	// Initialize logger
	if err := logger.Init(level, logFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Log initialization
	logger.Debug("Logger initialized",
		"level", logLevel,
		"output", func() string {
			if logFile != "" {
				return logFile
			}
			return "stderr"
		}(),
		"version", Version,
		"commit", Commit,
	)
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
		{"Getting Started", "🚀", []string{"walkthrough", "docs", "wizard", "tree", "contents", "help"}},
		{"Claude Code", "🤖", []string{"claude", "key"}},
		{"Lambda Cloud", "☁️ ", []string{"lambda", "add", "list", "remove", "ssh", "status"}},
		{"Package Management", "📦", []string{"packages", "install", "interactive", "library", "cpm", "explore"}},
		{"Models & AI", "🧠", []string{"models", "ollama", "llm", "query", "prompt"}},
		{"Services", "⚡", []string{"start", "comfyui", "ui", "notebook", "run", "serve"}},
		{"Deployment", "🎯", []string{"deploy", "push", "source", "bootstrap", "ship", "rsync"}},
		{"Source Code", "📜", []string{"extract", "build", "develop", "home"}},
		{"Generation", "🎬", []string{"generate", "animate", "upscale", "reel", "sky"}},
		{"Workflows", "🎨", []string{"browse-workflows", "collection", "suggest", "dashboard", "studio"}},
		{"Git & GitHub", "🔀", []string{"git", "github", "gh"}},
		{"Configuration", "⚙️", []string{"config", "modules", "set-modules", "set", "db"}},
		{"Diagnostics", "🏥", []string{"doctor", "jobs", "metrics", "logs", "billing", "validate"}},
		{"Updates", "📋", []string{"update", "updates", "resume", "rollback"}},
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
	fmt.Printf("  %s %s\n", "🔧", theme.SuccessStyle.Render("Utilities"))
	otherCmds := []string{"aliases", "purge", "reboot", "dns", "enhance", "completion"}
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

// showDetailedError displays a beautifully formatted DetailedError
func showDetailedError(err *errors.DetailedError) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Error"))
	fmt.Println()

	// Show what failed
	if err.Operation != "" {
		fmt.Printf("  %s %s\n",
			theme.WarningStyle.Render("Operation:"),
			theme.HighlightStyle.Render("Failed to "+err.Operation))
		fmt.Println()
	}

	// Show the underlying error
	if err.Cause != nil {
		fmt.Printf("  %s\n", theme.ErrorStyle.Render("Reason:"))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(err.Cause.Error()))
		fmt.Println()
	}

	// Show context if available
	if len(err.Context) > 0 {
		fmt.Printf("  %s\n", theme.InfoStyle.Render("Context:"))
		for _, ctx := range err.Context {
			fmt.Printf("    %s %s\n",
				theme.SuccessStyle.Render("•"),
				theme.DimTextStyle.Render(ctx))
		}
		fmt.Println()
	}

	// Show suggestions prominently
	if len(err.Suggestions) > 0 {
		fmt.Println(theme.GlowStyle.Render("  💡 How to Fix This:"))
		fmt.Println()
		for i, suggestion := range err.Suggestions {
			// Highlight the first 2 suggestions more prominently
			if i < 2 {
				fmt.Printf("    %s %s\n",
					theme.SuccessStyle.Render("→"),
					theme.HighlightStyle.Render(suggestion))
			} else {
				fmt.Printf("    %s %s\n",
					theme.DimTextStyle.Render("•"),
					theme.DimTextStyle.Render(suggestion))
			}
		}
		fmt.Println()
	}

	// Footer
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Need more help? Run 'anime help' or check the documentation"))
	fmt.Println()
}
