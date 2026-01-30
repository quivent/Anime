package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	allUpdates       bool
	limitCount       int
	showCommits      bool
	showCommandsOnly bool
)

// Release represents a version release with its changes
type Release struct {
	Version      string
	Date         string
	Codename     string
	Highlights   []string
	Commands     []CommandItem
	Features     []ChangeItem
	Fixes        []ChangeItem
	Improvements []ChangeItem
	Breaking     []ChangeItem
}

// ChangeItem represents a single change entry
type ChangeItem struct {
	Description string
	Category    string
}

// CommandItem represents a new command added in a release
type CommandItem struct {
	Command     string
	Description string
}

// Curated changelog - add new releases at the top
var changelog = []Release{
	{
		Version:  "1.0.242",
		Date:     "December 2025",
		Codename: "Sakura Storm",
		Highlights: []string{
			"Massive model library expansion with 30+ AI models",
			"New package filtering with `anime packages models`",
			"Enhanced TUI with better category organization",
		},
		Commands: []CommandItem{
			{Command: "anime packages models", Description: "List only AI model packages (image, video, enhancement)"},
			{Command: "anime updates", Description: "Show release notes and version history"},
			{Command: "anime models", Description: "List all models or show model catalog"},
			{Command: "anime models --catalog", Description: "Show detailed model catalog with use cases"},
			{Command: "anime models --installable", Description: "Show models available for installation"},
		},
		Features: []ChangeItem{
			{Description: "Added Flux 2 FP8 quantized model for efficient video generation", Category: "Models"},
			{Description: "Added CogVideoX 1.5 5B for 10-second video generation", Category: "Models"},
			{Description: "Added HunyuanVideo for high-quality T2V generation", Category: "Models"},
			{Description: "Added Pyramid Flow for efficient video synthesis", Category: "Models"},
			{Description: "Added SVD-XT for extended image-to-video", Category: "Models"},
			{Description: "Added I2V-Adapter for image-to-video adaptation", Category: "Models"},
			{Description: "Added SD 3.5 Large, Turbo, and Medium variants", Category: "Models"},
			{Description: "Added SDXL Turbo and Lightning for fast generation", Category: "Models"},
			{Description: "Added Playground v2.5 aesthetic model", Category: "Models"},
			{Description: "Added PixArt-Sigma for 4K resolution support", Category: "Models"},
			{Description: "Added Kandinsky 3 and Kolors image models", Category: "Models"},
			{Description: "Added Real-ESRGAN, GFPGAN, AuraSR, SUPIR enhancers", Category: "Enhancement"},
			{Description: "Added RIFE and FILM video interpolation", Category: "Enhancement"},
			{Description: "Added ControlNet adapters (Canny, Depth, OpenPose)", Category: "ControlNet"},
			{Description: "Added IP-Adapter, IP-Adapter FaceID, InstantID", Category: "ControlNet"},
		},
		Improvements: []ChangeItem{
			{Description: "Enhanced TUI category organization with new categories", Category: "TUI"},
			{Description: "Better emoji mappings for Image/Video Enhancement", Category: "TUI"},
			{Description: "Improved model catalog display with use cases", Category: "Display"},
		},
	},
	{
		Version:  "1.0.240",
		Date:     "November 2025",
		Codename: "Neon Genesis",
		Highlights: []string{
			"Interactive TUI package installer",
			"Remote workstation deployment",
			"SSH-based installation support",
		},
		Commands: []CommandItem{
			{Command: "anime tui", Description: "Interactive TUI package selector"},
			{Command: "anime interactive", Description: "Alias for anime tui"},
			{Command: "anime workstation", Description: "Launch workstation monitoring TUI"},
			{Command: "anime packages", Description: "List available installation packages"},
			{Command: "anime install --remote", Description: "Install packages on remote server via SSH"},
		},
		Features: []ChangeItem{
			{Description: "Interactive TUI package selector with Bubbletea", Category: "TUI"},
			{Description: "Remote installation via SSH to Lambda/RunPod", Category: "Remote"},
			{Description: "Package dependency resolution system", Category: "Core"},
			{Description: "Real-time installation progress tracking", Category: "UX"},
			{Description: "Workstation status monitoring", Category: "Remote"},
		},
		Improvements: []ChangeItem{
			{Description: "Unified package management across local/remote", Category: "Core"},
			{Description: "Better error handling for SSH connections", Category: "Remote"},
		},
	},
	{
		Version:  "1.0.235",
		Date:     "October 2025",
		Codename: "Cherry Blossom",
		Highlights: []string{
			"ComfyUI workflow management",
			"Model installation automation",
			"Lambda Labs integration",
		},
		Commands: []CommandItem{
			{Command: "anime comfyui", Description: "Manage ComfyUI server"},
			{Command: "anime ui", Description: "Quick shortcut for anime start comfyui"},
			{Command: "anime browse-workflows", Description: "Browse AI workflows and pipelines"},
			{Command: "anime lambda", Description: "Lambda server operations"},
			{Command: "anime ollama", Description: "Run ollama commands"},
			{Command: "anime llm", Description: "Quick shortcut for anime start ollama"},
			{Command: "anime query", Description: "Query a running Ollama model"},
		},
		Features: []ChangeItem{
			{Description: "ComfyUI automatic installation and configuration", Category: "ComfyUI"},
			{Description: "Workflow browsing and installation", Category: "ComfyUI"},
			{Description: "Lambda Labs API integration", Category: "Cloud"},
			{Description: "GPU instance provisioning", Category: "Cloud"},
			{Description: "Ollama LLM integration", Category: "LLM"},
		},
		Fixes: []ChangeItem{
			{Description: "Fixed CUDA detection on various GPU architectures", Category: "Core"},
			{Description: "Resolved path issues on Windows subsystem", Category: "Platform"},
		},
	},
	{
		Version:  "1.0.230",
		Date:     "September 2025",
		Codename: "First Light",
		Highlights: []string{
			"Initial public release",
			"Core CLI framework",
			"Foundation packages",
		},
		Commands: []CommandItem{
			{Command: "anime install <package>", Description: "Install packages on Lambda server"},
			{Command: "anime add <server>", Description: "Add a new Lambda server"},
			{Command: "anime ssh <server>", Description: "SSH into a server"},
			{Command: "anime status", Description: "Check status of a Lambda server"},
			{Command: "anime start", Description: "Start services (comfyui, ollama, jupyter)"},
			{Command: "anime deploy", Description: "Deploy and install modules on server"},
			{Command: "anime doctor", Description: "Diagnose installation failures"},
		},
		Features: []ChangeItem{
			{Description: "Core CLI framework with Cobra", Category: "Core"},
			{Description: "Package installation system", Category: "Core"},
			{Description: "Python/PyTorch environment setup", Category: "Foundation"},
			{Description: "NVIDIA driver detection", Category: "Foundation"},
			{Description: "Basic model downloads", Category: "Models"},
		},
	},
}

var updatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Show release notes and version history",
	Long: `Display a beautifully formatted changelog with release notes.

Shows:
  - Version history with release dates
  - New features and improvements
  - Bug fixes and breaking changes
  - Release highlights and codenames

Use --commits to see git commit history instead.`,
	RunE: runUpdates,
}

func init() {
	rootCmd.AddCommand(updatesCmd)
	updatesCmd.Flags().BoolVarP(&allUpdates, "all", "a", false, "Show all version history")
	updatesCmd.Flags().IntVarP(&limitCount, "limit", "l", 3, "Number of versions to show")
	updatesCmd.Flags().BoolVarP(&showCommits, "commits", "c", false, "Show git commits instead of changelog")
	updatesCmd.Flags().BoolVar(&showCommandsOnly, "commands", false, "Show only new commands from all releases")
}

// Styles for the changelog
var (
	versionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.SakuraPink).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	versionTitleStyle = lipgloss.NewStyle().
				Foreground(theme.SakuraPink).
				Bold(true)

	codenameStyle = lipgloss.NewStyle().
			Foreground(theme.NeonPurple).
			Italic(true)

	dateStyle = lipgloss.NewStyle().
			Foreground(theme.TextDim)

	highlightBoxStyle = lipgloss.NewStyle().
				Foreground(theme.ElectricBlue).
				PaddingLeft(2)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(theme.MintGreen).
				Bold(true).
				MarginTop(1)

	featureStyle = lipgloss.NewStyle().
			Foreground(theme.TextPrimary).
			PaddingLeft(4)

	categoryTagStyle = lipgloss.NewStyle().
				Foreground(theme.NeonPurple).
				Background(lipgloss.Color("#2D2D3D")).
				Padding(0, 1)

	fixStyle = lipgloss.NewStyle().
			Foreground(theme.SunsetOrange).
			PaddingLeft(4)

	breakingStyle = lipgloss.NewStyle().
			Foreground(theme.ActionRed).
			Bold(true).
			PaddingLeft(4)

	dividerStyle = lipgloss.NewStyle().
			Foreground(theme.TextDim)
)

func runUpdates(cmd *cobra.Command, args []string) error {
	// If --commits flag, show git commit history
	if showCommits {
		return runGitUpdates()
	}

	// If --commands flag, show only commands
	if showCommandsOnly {
		return runCommandsList()
	}

	// Show curated changelog
	return runChangelog()
}

func runCommandsList() error {
	fmt.Println()

	// Header banner
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.SakuraPink).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(1, 4).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.NeonPurple).
		Align(lipgloss.Center)

	fmt.Println(bannerStyle.Render("ANIME CLI - ALL COMMANDS"))
	fmt.Println()

	// Collect all commands from all releases
	allCommands := make(map[string][]CommandItem)
	for _, release := range changelog {
		if len(release.Commands) > 0 {
			allCommands[release.Version] = release.Commands
		}
	}

	// Display commands by version
	for _, release := range changelog {
		if len(release.Commands) == 0 {
			continue
		}

		fmt.Printf("  %s  %s\n",
			versionTitleStyle.Render("v"+release.Version),
			codenameStyle.Render("\""+release.Codename+"\""))
		fmt.Println()

		for _, cmd := range release.Commands {
			fmt.Printf("    %s %s\n", "💻", commandStyle.Render(cmd.Command))
			fmt.Printf("       %s\n", commandDescStyle.Render(cmd.Description))
		}
		fmt.Println()
		fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 60)))
		fmt.Println()
	}

	// Quick reference section
	fmt.Printf("  %s\n", theme.GlowStyle.Render("Quick Reference - Most Used Commands"))
	fmt.Println()

	quickRef := []CommandItem{
		{Command: "anime help", Description: "Show all commands organized by category"},
		{Command: "anime tree", Description: "View complete command tree"},
		{Command: "anime models", Description: "List downloaded models"},
		{Command: "anime models --catalog", Description: "Show model catalog with use cases"},
		{Command: "anime packages", Description: "List all available packages"},
		{Command: "anime packages models", Description: "List only AI model packages"},
		{Command: "anime tui", Description: "Interactive package selector"},
		{Command: "anime install <package>", Description: "Install a package"},
	}

	for _, cmd := range quickRef {
		fmt.Printf("    %s %s\n", "⭐", commandStyle.Render(cmd.Command))
		fmt.Printf("       %s\n", commandDescStyle.Render(cmd.Description))
	}
	fmt.Println()

	return nil
}

func runChangelog() error {
	fmt.Println()

	// Header banner
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.SakuraPink).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(1, 4).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.NeonPurple).
		Align(lipgloss.Center)

	fmt.Println(bannerStyle.Render("ANIME CLI - RELEASE NOTES"))
	fmt.Println()

	// Current version info
	currentStyle := lipgloss.NewStyle().
		Foreground(theme.ElectricBlue).
		Bold(true)

	fmt.Printf("  %s %s\n", currentStyle.Render("Current Version:"), theme.HighlightStyle.Render(Version))
	fmt.Println()

	// Determine how many versions to show
	showCount := limitCount
	if allUpdates {
		showCount = len(changelog)
	}
	if showCount > len(changelog) {
		showCount = len(changelog)
	}

	// Render each release
	for i := 0; i < showCount; i++ {
		release := changelog[i]
		renderRelease(release, i == 0)

		if i < showCount-1 {
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 60)))
		}
	}

	fmt.Println()

	// Footer
	if !allUpdates && len(changelog) > limitCount {
		remainingCount := len(changelog) - limitCount
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("... and %d more releases", remainingCount)))
		fmt.Printf("  %s\n", theme.InfoStyle.Render("Run 'anime updates --all' to see full history"))
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("View git commits:"), theme.HighlightStyle.Render("anime updates --commits"))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Update to latest:"), theme.HighlightStyle.Render("anime update"))
	fmt.Println()

	return nil
}

func renderRelease(release Release, isLatest bool) {
	// Version header
	versionStr := fmt.Sprintf("v%s", release.Version)
	if isLatest {
		versionStr += "  (latest)"
	}

	fmt.Println()
	fmt.Printf("  %s  %s  %s\n",
		versionTitleStyle.Render(versionStr),
		codenameStyle.Render("\""+release.Codename+"\""),
		dateStyle.Render(release.Date))
	fmt.Println()

	// Highlights
	if len(release.Highlights) > 0 {
		fmt.Printf("  %s\n", theme.GlowStyle.Render("Highlights"))
		for _, h := range release.Highlights {
			fmt.Printf("    %s %s\n", theme.SymbolStar, highlightBoxStyle.Render(h))
		}
		fmt.Println()
	}

	// Commands
	if len(release.Commands) > 0 {
		fmt.Printf("  %s\n", sectionHeaderStyle.Render("New Commands"))
		renderCommands(release.Commands)
	}

	// Features
	if len(release.Features) > 0 {
		fmt.Printf("  %s\n", sectionHeaderStyle.Render("New Features"))
		renderChangeItems(release.Features, "✨", featureStyle)
	}

	// Improvements
	if len(release.Improvements) > 0 {
		fmt.Printf("  %s\n", sectionHeaderStyle.Render("Improvements"))
		renderChangeItems(release.Improvements, "⚡", featureStyle)
	}

	// Fixes
	if len(release.Fixes) > 0 {
		fmt.Printf("  %s\n", sectionHeaderStyle.Render("Bug Fixes"))
		renderChangeItems(release.Fixes, "🔧", fixStyle)
	}

	// Breaking changes
	if len(release.Breaking) > 0 {
		fmt.Printf("  %s\n", sectionHeaderStyle.Render("Breaking Changes"))
		renderChangeItems(release.Breaking, "⚠️", breakingStyle)
	}
}

func renderChangeItems(items []ChangeItem, icon string, style lipgloss.Style) {
	// Group by category
	categories := make(map[string][]ChangeItem)
	categoryOrder := []string{}

	for _, item := range items {
		if _, exists := categories[item.Category]; !exists {
			categoryOrder = append(categoryOrder, item.Category)
		}
		categories[item.Category] = append(categories[item.Category], item)
	}

	for _, cat := range categoryOrder {
		catItems := categories[cat]
		for _, item := range catItems {
			tag := categoryTagStyle.Render(item.Category)
			fmt.Printf("    %s %s %s\n", icon, tag, style.Render(item.Description))
		}
	}
	fmt.Println()
}

// Style for command display
var (
	commandStyle = lipgloss.NewStyle().
			Foreground(theme.SakuraPink).
			Bold(true)

	commandDescStyle = lipgloss.NewStyle().
				Foreground(theme.TextSecondary)
)

func renderCommands(commands []CommandItem) {
	for _, cmd := range commands {
		fmt.Printf("    %s %s\n", "💻", commandStyle.Render(cmd.Command))
		fmt.Printf("       %s\n", commandDescStyle.Render(cmd.Description))
	}
	fmt.Println()
}

// runGitUpdates shows git commit history (original functionality)
func runGitUpdates() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("GIT COMMIT HISTORY"))
	fmt.Println()

	// Check if build directory is embedded
	if BuildDir == "" {
		fmt.Println(theme.WarningStyle.Render("Build directory not embedded in binary"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("This binary was not built with self-update support."))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("To see updates, check the git repository:"))
		fmt.Println(theme.HighlightStyle.Render("  https://github.com/joshkornreich/anime"))
		fmt.Println()
		return nil
	}

	// Check if build directory exists
	if _, err := os.Stat(BuildDir); os.IsNotExist(err) {
		fmt.Println(theme.ErrorStyle.Render("Source directory not found"))
		fmt.Println()
		fmt.Printf("  Expected: %s\n", theme.HighlightStyle.Render(BuildDir))
		fmt.Println()
		return fmt.Errorf("source directory not found: %s", BuildDir)
	}

	// Check if it's a git repository
	gitCheckCmd := exec.Command("git", "-C", BuildDir, "rev-parse", "--git-dir")
	if err := gitCheckCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("Not a git repository"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("The source directory is not a git repository."))
		fmt.Println()
		return fmt.Errorf("not a git repository: %s", BuildDir)
	}

	// Show current version
	fmt.Println(theme.GlowStyle.Render("Current Version"))
	fmt.Printf("  Version:   %s\n", theme.HighlightStyle.Render(Version))
	fmt.Printf("  Commit:    %s\n", theme.DimTextStyle.Render(Commit))
	buildTimeDisplay := BuildTime
	if BuildTime != "unknown" {
		buildTimeDisplay = strings.ReplaceAll(BuildTime, "_", " ")
	}
	fmt.Printf("  Built:     %s\n", theme.DimTextStyle.Render(buildTimeDisplay))
	fmt.Println()

	// Fetch latest changes (quietly)
	fmt.Println(theme.GlowStyle.Render("Checking for updates..."))
	fetchCmd := exec.Command("git", "-C", BuildDir, "fetch", "origin", "--quiet")
	if err := fetchCmd.Run(); err != nil {
		fmt.Println(theme.WarningStyle.Render("Could not fetch latest changes"))
		fmt.Println(theme.DimTextStyle.Render("  Showing local commits only"))
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("Fetched latest updates"))
		fmt.Println()
	}

	// Get current commit
	currentCommit := Commit
	if currentCommit == "unknown" || currentCommit == "" {
		getCurrentCmd := exec.Command("git", "-C", BuildDir, "rev-parse", "HEAD")
		if output, err := getCurrentCmd.Output(); err == nil {
			currentCommit = strings.TrimSpace(string(output))[:7]
		}
	} else {
		if len(currentCommit) > 7 {
			currentCommit = currentCommit[:7]
		}
	}

	// Determine log range
	limit := 20
	if limitCount > 0 {
		limit = limitCount
	}

	logCmd := exec.Command("git", "-C", BuildDir, "log",
		fmt.Sprintf("-n%d", limit),
		"--pretty=format:%h|%s|%an|%ar", "--no-merges")

	output, err := logCmd.Output()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to get commit history"))
		fmt.Println()
		return fmt.Errorf("git log failed: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")

	if len(commits) == 0 || (len(commits) == 1 && commits[0] == "") {
		fmt.Println(theme.SuccessStyle.Render("No commits found"))
		fmt.Println()
		return nil
	}

	// Display commits
	fmt.Println(theme.GlowStyle.Render("Recent Commits:"))
	fmt.Println()

	for i, commit := range commits {
		if commit == "" {
			continue
		}

		parts := strings.SplitN(commit, "|", 4)
		if len(parts) != 4 {
			continue
		}

		hash := parts[0]
		message := parts[1]
		author := parts[2]
		timeAgo := parts[3]

		// Color code based on commit message prefix
		messageStyle := theme.InfoStyle
		icon := "•"

		msgLower := strings.ToLower(message)
		if strings.HasPrefix(msgLower, "feat") || strings.HasPrefix(msgLower, "add") {
			messageStyle = theme.SuccessStyle
			icon = "✨"
		} else if strings.HasPrefix(msgLower, "fix") {
			messageStyle = theme.WarningStyle
			icon = "🔧"
		} else if strings.HasPrefix(msgLower, "refactor") {
			icon = "♻️"
		} else if strings.HasPrefix(msgLower, "docs") {
			icon = "📚"
		} else if strings.HasPrefix(msgLower, "test") {
			icon = "🧪"
		} else if strings.HasPrefix(msgLower, "perf") {
			icon = "⚡"
		} else if strings.HasPrefix(msgLower, "chore") {
			icon = "🔨"
		} else if strings.HasPrefix(msgLower, "style") {
			icon = "🎨"
		}

		fmt.Printf("  %s %s %s\n",
			icon,
			theme.DimTextStyle.Render(hash),
			messageStyle.Render(message))

		fmt.Printf("    %s\n",
			theme.DimTextStyle.Render(fmt.Sprintf("by %s, %s", author, timeAgo)))

		if i < len(commits)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  " + strings.Repeat("─", 55)))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("  To install updates, run:"))
	fmt.Printf("   %s\n", theme.HighlightStyle.Render("anime update"))
	fmt.Println()

	return nil
}
