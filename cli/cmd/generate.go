package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

import "runtime"

// getReelsBinaryPath returns the appropriate binary path based on platform
func getReelsBinaryPath() string {
	home := os.Getenv("HOME")
	base := filepath.Join(home, "amber", "reels")

	// On macOS, try platform-specific binary for local testing
	if runtime.GOOS == "darwin" {
		macPath := filepath.Join(base, "reels-mac")
		if _, err := os.Stat(macPath); err == nil {
			return macPath
		}
	}

	// Linux production or fallback
	return filepath.Join(base, "reels")
}

var reelsBinaryPath = getReelsBinaryPath()

var generateCmd = &cobra.Command{
	Use:   "generate [command] [args]",
	Short: "Generate videos with Wan2.1/SkyReels",
	Long: `Video Generation CLI - Full-featured video generation with presets, wizards, and GPU optimization.

This command wraps the reels CLI for comprehensive video generation capabilities.

Commands:
  generate      Generate videos from text prompts
  status        Show system, GPU, and environment status
  wizard        Interactive wizards for common operations
  benchmark     Run performance benchmarks
  config        Configuration management
  estimate      Estimate generation time and resources
  models        List and manage models
  optimizations View optimization options
  setup         Setup and installation helpers
  stacks        Manage generation stacks

Generation Subcommands:
  generate presets       List available presets
  generate init-config   Create a config file
  generate example-config Show example configuration

Examples:
  # Basic generation
  anime generate -p "A sunset over the ocean" -o sunset.mp4

  # High quality with custom settings
  anime generate -p "City skyline" -o city.mp4 -r 1080p -s 75 --duration 10

  # Use a preset
  anime generate -p "Nature scene" -o nature.mp4 --preset high-quality

  # Preview settings without generating
  anime generate -p "Test prompt" -o test.mp4 --dry-run

  # Interactive wizard
  anime generate wizard generate`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printGenerateHelp()
			return
		}
		runReelsBinary(args)
	},
	DisableFlagParsing: true,
}

func printGenerateHelp() {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎥 Video Generation"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Wan2.1/SkyReels video generation with presets and wizards"))
	fmt.Println()

	fmt.Println(theme.HeaderStyle.Render("  Commands"))
	fmt.Println()

	commands := []struct {
		name string
		desc string
	}{
		{"status", "Show system, GPU, and environment status"},
		{"wizard", "Interactive wizards for common operations"},
		{"benchmark", "Run performance benchmarks"},
		{"config", "Configuration management"},
		{"estimate", "Estimate generation time and resources"},
		{"models", "List and manage models"},
		{"optimizations", "View optimization options"},
		{"setup", "Setup and installation helpers"},
		{"stacks", "Manage generation stacks"},
	}

	for _, c := range commands {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-16s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Generation Flags"))
	fmt.Println()

	flags := []struct {
		flag string
		desc string
	}{
		{"-p, --prompt", "Text prompt for video generation"},
		{"-o, --output", "Output file path (required)"},
		{"-d, --duration", "Video duration in seconds (1-60)"},
		{"-r, --resolution", "Resolution (480p, 720p, 1080p, 1440p, 4k)"},
		{"-s, --steps", "Number of diffusion steps (1-200)"},
		{"-g, --guidance", "Guidance scale (1.0-20.0)"},
		{"--seed", "Random seed (-1 for random)"},
		{"-m, --model", "Model to use"},
		{"--preset", "Use preset (quick, standard, high-quality)"},
		{"--dry-run", "Preview without generating"},
		{"-w, --watch", "Live progress monitoring"},
	}

	for _, f := range flags {
		fmt.Printf("    %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-18s", f.flag)),
			theme.DimTextStyle.Render(f.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Examples"))
	fmt.Println()
	fmt.Printf("    %s anime generate -p \"Ocean waves\" -o waves.mp4%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime generate -p \"City night\" -o city.mp4 -r 1080p -s 75%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime generate -p \"Nature\" -o nature.mp4 --preset high-quality%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime generate wizard generate%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime generate status%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Println()

	fmt.Println(theme.HeaderStyle.Render("  Presets"))
	fmt.Println()

	presets := []struct {
		name string
		desc string
	}{
		{"quick", "Fast preview (720p, 25 steps)"},
		{"standard", "Balanced quality (1080p, 50 steps)"},
		{"high-quality", "High quality (1080p, 75 steps)"},
		{"maximum", "Maximum quality (1440p, 100 steps)"},
	}

	for _, p := range presets {
		fmt.Printf("    %s  %s\n",
			theme.SuccessStyle.Render(fmt.Sprintf("%-14s", p.name)),
			theme.DimTextStyle.Render(p.desc))
	}

	fmt.Println()
}

func runReelsBinary(args []string) {
	// Check if binary exists
	if _, err := os.Stat(reelsBinaryPath); os.IsNotExist(err) {
		fmt.Println()
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Reels binary not found"))
		fmt.Println()
		fmt.Printf("    Expected: %s\n", theme.DimTextStyle.Render(reelsBinaryPath))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("    Build it with:"))
		fmt.Printf("      %s cd ~/amber/reels/cli && go build -o ../reels%s\n", theme.DimTextStyle.Render("$"), "")
		fmt.Println()
		return
	}

	// Execute the reels binary, replacing current process
	cmdArgs := append([]string{reelsBinaryPath}, args...)
	env := os.Environ()

	if err := syscall.Exec(reelsBinaryPath, cmdArgs, env); err != nil {
		// Fallback to exec.Command if syscall.Exec fails
		cmd := exec.Command(reelsBinaryPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
