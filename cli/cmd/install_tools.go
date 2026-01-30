package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var buildToolCmd = &cobra.Command{
	Use:   "build [tool]",
	Short: "Build and install embedded tools (reels, sky)",
	Long: `Build and install embedded tools globally to ~/.local/bin

Available tools:
  reels  - SkyReels comprehensive CLI (Wan2.1 video generation)
  sky    - SkyReels system analysis and management

Note: The 'reel' command is already built into anime as 'anime reel'

Examples:
  anime build reels  # Build and install reels CLI
  anime build sky    # Build and install sky CLI
  anime build all    # Build and install all tools`,
	ValidArgs: []string{"reels", "sky", "all"},
	Args:      cobra.ExactValidArgs(1),
	RunE:      runBuildTool,
}

func runBuildTool(cmd *cobra.Command, args []string) error {
	tool := args[0]

	if tool == "all" {
		tools := []string{"reels", "sky"}
		for _, t := range tools {
			if err := installTool(t); err != nil {
				return err
			}
		}
		return nil
	}

	// reel is built-in to anime, not a separate tool
	if tool == "reel" {
		return fmt.Errorf("'reel' is built into anime - use 'anime reel' instead of building it separately")
	}

	return installTool(tool)
}

func installTool(tool string) error {
	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render(fmt.Sprintf("📦 Installing %s", tool)))
	fmt.Println()

	// Get the embedded source directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find the anime repo root (where cmd/ directory exists)
	animeRoot := findAnimeRoot(cwd)
	if animeRoot == "" {
		return fmt.Errorf("could not find anime repository root")
	}

	sourceDir := filepath.Join(animeRoot, "cmd", "embedded", tool)

	// Check if source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("tool %s not found in embedded sources", tool)
	}

	fmt.Printf("  %s Source: %s\n", theme.SymbolInfo, theme.DimTextStyle.Render(sourceDir))

	// No special dependency setup needed for sky or reels

	// Determine the appropriate source directory based on tool structure
	buildDir := sourceDir
	if tool == "reels" {
		// reels has a cli subdirectory
		cliDir := filepath.Join(sourceDir, "cli")
		if _, err := os.Stat(cliDir); err == nil {
			buildDir = cliDir
		}
	}

	// Create ~/.local/bin if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	outputPath := filepath.Join(binDir, tool)

	fmt.Printf("  %s Building...\n", theme.SymbolInfo)

	// Build the tool
	buildCmd := exec.Command("go", "build", "-o", outputPath)
	buildCmd.Dir = buildDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build %s: %w", tool, err)
	}

	// Make executable
	if err := os.Chmod(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	fmt.Println()
	fmt.Printf("  %s Installed to: %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render(outputPath))
	fmt.Println()

	// Check if ~/.local/bin is in PATH
	pathEnv := os.Getenv("PATH")
	if !containsPath(filepath.SplitList(pathEnv), binDir) {
		fmt.Println(theme.WarningStyle.Render("  ⚠️  ~/.local/bin is not in your PATH"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Add this to your ~/.zshrc or ~/.bashrc:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(`export PATH="$HOME/.local/bin:$PATH"`))
		fmt.Println()
	}

	return nil
}

func findAnimeRoot(startDir string) string {
	dir := startDir
	for {
		cmdDir := filepath.Join(dir, "cmd")
		if _, err := os.Stat(cmdDir); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}
	return ""
}

func containsPath(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(buildToolCmd)
}
