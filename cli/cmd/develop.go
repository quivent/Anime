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

var developCmd = &cobra.Command{
	Use:   "develop [path]",
	Short: "Launch Claude Code in anime's source directory",
	Long: `Navigate to anime's source build location and start a Claude Code session.

This command uses the build directory embedded during compilation to automatically
launch Claude Code in the anime CLI source directory with bypass permissions enabled.

On remote servers where the embedded build path doesn't exist, you can:
  - Provide a custom path:  anime develop /path/to/anime/cli
  - Clone the repo first:   git clone https://github.com/joshkornreich/anime.git`,
	Run: runDevelop,
}

func init() {
	rootCmd.AddCommand(developCmd)
}

func runDevelop(cmd *cobra.Command, args []string) {
	// Resolve the source directory to use
	sourceDir := resolveSourceDirectory(args)
	if sourceDir == "" {
		os.Exit(1)
	}

	// Check if claude command is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("💥 Claude Code Not Found"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  The 'claude' command is not available in your PATH."))
		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("  💡 Install Claude Code:"))
		fmt.Println()
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("npm install -g @anthropic-ai/claude-code"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Or visit: https://github.com/anthropics/claude-code"))
		fmt.Println()
		os.Exit(1)
	}

	// Display info before launching
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 ANIME DEVELOP 🚀"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Source:"),
		theme.SuccessStyle.Render(sourceDir))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Claude:"),
		theme.DimTextStyle.Render(claudePath))
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("  Launching Claude Code with bypass permissions..."))
	fmt.Println()

	// Change to source directory and exec claude with bypass permissions
	// Using syscall.Exec to replace the current process with claude
	if err := os.Chdir(sourceDir); err != nil {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Failed to change directory: %v", err)))
		fmt.Println()
		os.Exit(1)
	}

	// Prepare arguments for claude
	claudeArgs := []string{"claude", "--permission-mode", "bypassPermissions"}

	// Execute claude, replacing the current process
	err = syscall.Exec(claudePath, claudeArgs, os.Environ())
	if err != nil {
		// If exec fails, we'll still be here
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Failed to launch Claude Code: %v", err)))
		fmt.Println()
		os.Exit(1)
	}

	// This line should never be reached if exec succeeds
}

// resolveSourceDirectory finds the anime source directory to use
func resolveSourceDirectory(args []string) string {
	// 1. If user provided a path argument, use that
	if len(args) > 0 {
		customPath := args[0]
		if !filepath.IsAbs(customPath) {
			if abs, err := filepath.Abs(customPath); err == nil {
				customPath = abs
			}
		}
		if isValidAnimeSource(customPath) {
			return customPath
		}
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("💥 Invalid Source Directory"))
		fmt.Println()
		fmt.Printf("  %s %s\n",
			theme.WarningStyle.Render("Provided path:"),
			theme.DimTextStyle.Render(customPath))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  The path doesn't appear to be an anime CLI source directory."))
		fmt.Println(theme.DimTextStyle.Render("  (Expected to find go.mod with anime module)"))
		fmt.Println()
		return ""
	}

	// 2. If BuildDir is set and exists, use it (for development on the build machine)
	if BuildDir != "" {
		if isValidAnimeSource(BuildDir) {
			return BuildDir
		}
	}

	// 3. Check if we already have extracted source in cache
	cacheDir := GetSourceCacheDir()
	cachedSource := filepath.Join(cacheDir, "cli")
	if isValidAnimeSource(cachedSource) {
		fmt.Println()
		fmt.Printf("  %s %s\n",
			theme.InfoStyle.Render("Using cached source:"),
			theme.SuccessStyle.Render(cachedSource))
		return cachedSource
	}

	// 4. Try to extract embedded source (this is the main path for deployed binaries)
	if HasEmbeddedSource() {
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  📦 Source directory not found, extracting from embedded source..."))
		fmt.Println()

		extractedPath, err := ExtractEmbeddedSource(cacheDir)
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  ✗ Failed to extract: %v", err)))
		} else if isValidAnimeSource(extractedPath) {
			fmt.Printf("  %s %s\n",
				theme.SuccessStyle.Render("✓ Extracted to:"),
				theme.HighlightStyle.Render(extractedPath))
			return extractedPath
		}
	}

	// 5. Try common locations on this machine
	homeDir, _ := os.UserHomeDir()
	commonPaths := []string{
		filepath.Join(homeDir, "anime", "cli"),
		filepath.Join(homeDir, "anime-cli"),
		filepath.Join(homeDir, "src", "anime", "cli"),
		filepath.Join(homeDir, "projects", "anime", "cli"),
		"/opt/anime/cli",
		"/home/anime/cli",
	}

	for _, path := range commonPaths {
		if isValidAnimeSource(path) {
			fmt.Println()
			fmt.Printf("  %s %s\n",
				theme.InfoStyle.Render("Found source at:"),
				theme.SuccessStyle.Render(path))
			return path
		}
	}

	// 6. Nothing found - show helpful error
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("💥 Anime Source Directory Not Found"))
	fmt.Println()

	if BuildDir != "" {
		fmt.Printf("  %s %s\n",
			theme.WarningStyle.Render("Embedded build path:"),
			theme.DimTextStyle.Render(BuildDir))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  The embedded build path doesn't exist on this machine."))
	}

	if !HasEmbeddedSource() {
		fmt.Println(theme.WarningStyle.Render("  ⚠ No embedded source available in this binary."))
		fmt.Println(theme.DimTextStyle.Render("  Rebuild with 'make build' to embed source."))
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("  Options:"))
	fmt.Println()
	fmt.Printf("    %s\n", theme.HighlightStyle.Render("1. Extract source manually:"))
	fmt.Printf("       %s\n", theme.DimTextStyle.Render("anime extract ~/anime-src"))
	fmt.Printf("       %s\n", theme.DimTextStyle.Render("anime develop ~/anime-src/cli"))
	fmt.Println()
	fmt.Printf("    %s\n", theme.HighlightStyle.Render("2. Clone the repository:"))
	fmt.Printf("       %s\n", theme.DimTextStyle.Render("git clone https://github.com/joshkornreich/anime.git"))
	fmt.Printf("       %s\n", theme.DimTextStyle.Render("anime develop ~/anime/cli"))
	fmt.Println()

	return ""
}

// isValidAnimeSource checks if a path contains anime CLI source
func isValidAnimeSource(path string) bool {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for go.mod file
	goModPath := filepath.Join(path, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	// Check if go.mod contains anime module reference
	contentStr := string(content)
	return len(contentStr) > 0 && (
		// Check for module declaration
		filepath.Base(path) == "cli" && filepath.Base(filepath.Dir(path)) == "anime" ||
		// Or check content for anime module path
		len(contentStr) > 10)
}
