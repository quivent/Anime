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

var skyBinaryPath = filepath.Join(os.Getenv("HOME"), "amber", "sky", "sky")

var skyCmd = &cobra.Command{
	Use:   "sky [command] [args]",
	Short: "SkyReels system analysis and management",
	Long: `SkyReels System Analysis & Management - Architecture analysis, GPU status, and setup procedures.

This command wraps the sky CLI for comprehensive system management.

Commands:
  analysis      View complete architecture analysis
  architecture  Display architecture diagrams and config
  procedures    Setup procedures and installation steps
  status        Check system status and readiness
  benchmark     Run performance benchmarks
  variants      List and configure model variants
  metrics       Display current system metrics
  enhance       Get optimization suggestions
  sequence      Implementation sequence protocols
  config        View and modify configuration
  init          Initialize configuration file
  doctor        Check if models are loaded on GPUs
  reload        Quick reload models with defaults
  load          Interactive wizard to load models

Examples:
  anime sky status              # Check system status
  anime sky analysis            # Full architecture analysis
  anime sky procedures show 3   # Show procedure #3
  anime sky benchmark estimate  # Performance estimates
  anime sky load                # Interactive model loading`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printSkyHelp()
			return
		}
		runSkyBinary(args)
	},
	DisableFlagParsing: true,
}

func printSkyHelp() {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🛸 SkyReels System Management"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  High-throughput video generation analysis for multi-GPU systems"))
	fmt.Println()

	fmt.Println(theme.HeaderStyle.Render("  Commands"))
	fmt.Println()

	commands := []struct {
		name string
		desc string
	}{
		{"load", "Interactive wizard to load models on GPUs"},
		{"doctor", "Check if models are loaded on GPUs"},
		{"reload", "Quick reload models with defaults"},
		{"status", "Check system status and readiness"},
		{"analysis", "View complete architecture analysis"},
		{"architecture", "Display architecture diagrams"},
		{"procedures", "Setup procedures and installation steps"},
		{"sequence", "Implementation sequence protocols"},
		{"benchmark", "Run performance benchmarks"},
		{"variants", "List and configure model variants"},
		{"metrics", "Display current system metrics"},
		{"enhance", "Get optimization suggestions"},
		{"config", "View and modify configuration"},
		{"init", "Initialize configuration file"},
	}

	for _, c := range commands {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-14s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Examples"))
	fmt.Println()
	fmt.Printf("    %s anime sky status%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime sky analysis explain memory%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime sky procedures show 3%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime sky benchmark estimate%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Println()
}

func runSkyBinary(args []string) {
	// Check if binary exists
	if _, err := os.Stat(skyBinaryPath); os.IsNotExist(err) {
		fmt.Println()
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Sky binary not found"))
		fmt.Println()
		fmt.Printf("    Expected: %s\n", theme.DimTextStyle.Render(skyBinaryPath))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("    Build it with:"))
		fmt.Printf("      %s cd ~/amber/sky && go build -o sky%s\n", theme.DimTextStyle.Render("$"), "")
		fmt.Println()
		return
	}

	// Execute the sky binary, replacing current process
	cmdArgs := append([]string{skyBinaryPath}, args...)
	env := os.Environ()

	if err := syscall.Exec(skyBinaryPath, cmdArgs, env); err != nil {
		// Fallback to exec.Command if syscall.Exec fails (e.g., on some platforms)
		cmd := exec.Command(skyBinaryPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

func init() {
	rootCmd.AddCommand(skyCmd)
}
