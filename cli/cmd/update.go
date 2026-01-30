package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update anime to the latest version",
	Long: `Update anime to the latest version by pulling from git and rebuilding.

This command will:
  • Pull the latest changes from git
  • Rebuild the binary
  • Replace the current binary with the new version

Note: This only works if anime was built from source.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔄 ANIME UPDATE 🔄"))
	fmt.Println()

	// Check if build directory is embedded
	if BuildDir == "" {
		fmt.Println(theme.WarningStyle.Render("⚠️  Build directory not embedded in binary"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("This binary was not built with self-update support."))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("To enable self-updates, rebuild with:"))
		fmt.Println(theme.HighlightStyle.Render("  make build"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Or manually update by running:"))
		fmt.Println(theme.HighlightStyle.Render("  cd /path/to/anime-cli && git pull && make build"))
		fmt.Println()
		return nil
	}

	// Check if build directory exists
	if _, err := os.Stat(BuildDir); os.IsNotExist(err) {
		fmt.Println(theme.ErrorStyle.Render("❌ Source directory not found"))
		fmt.Println()
		fmt.Printf("  Expected: %s\n", theme.HighlightStyle.Render(BuildDir))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("The source directory may have been moved or deleted."))
		fmt.Println()
		return fmt.Errorf("source directory not found: %s", BuildDir)
	}

	fmt.Println(theme.InfoStyle.Render("📂 Source directory: ") + theme.HighlightStyle.Render(BuildDir))
	fmt.Println()

	// Check if it's a git repository
	gitCheckCmd := exec.Command("git", "-C", BuildDir, "rev-parse", "--git-dir")
	if err := gitCheckCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Not a git repository"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("The source directory is not a git repository."))
		fmt.Println(theme.DimTextStyle.Render("Manual update required."))
		fmt.Println()
		return fmt.Errorf("not a git repository: %s", BuildDir)
	}

	// Step 1: Check current version
	fmt.Println(theme.GlowStyle.Render("🔍 Current version"))
	fmt.Printf("  Version:   %s\n", theme.HighlightStyle.Render(Version))
	fmt.Printf("  Commit:    %s\n", theme.DimTextStyle.Render(Commit))
	fmt.Printf("  Built:     %s\n", theme.DimTextStyle.Render(BuildTime))
	fmt.Println()

	// Step 2: Fetch latest changes
	fmt.Println(theme.GlowStyle.Render("📡 Fetching latest changes..."))
	fetchCmd := exec.Command("git", "-C", BuildDir, "fetch", "origin")
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to fetch changes"))
		fmt.Println()
		return fmt.Errorf("git fetch failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓ Fetched successfully"))
	fmt.Println()

	// Check if there are updates
	fmt.Println(theme.GlowStyle.Render("🔎 Checking for updates..."))
	statusCmd := exec.Command("git", "-C", BuildDir, "status", "-uno")
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to check git status"))
		fmt.Println()
		return fmt.Errorf("git status failed: %w", err)
	}

	statusStr := string(statusOutput)
	if strings.Contains(statusStr, "Your branch is up to date") {
		fmt.Println(theme.SuccessStyle.Render("✓ Already up to date!"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("You're running the latest version."))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.InfoStyle.Render("📦 Updates available"))
	fmt.Println()

	// Step 3: Pull latest changes
	fmt.Println(theme.GlowStyle.Render("⬇️  Pulling latest changes..."))
	pullCmd := exec.Command("git", "-C", BuildDir, "pull", "origin", "main")
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to pull changes"))
		fmt.Println()
		return fmt.Errorf("git pull failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓ Pulled successfully"))
	fmt.Println()

	// Step 4: Rebuild binary
	fmt.Println(theme.GlowStyle.Render("🔨 Rebuilding binary..."))

	// Determine the build command
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = BuildDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Env = os.Environ()

	if err := buildCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Build failed"))
		fmt.Println()
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓ Built successfully"))
	fmt.Println()

	// Step 5: Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to get current executable path"))
		fmt.Println()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to resolve executable path"))
		fmt.Println()
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Step 6: Replace binary
	fmt.Println(theme.GlowStyle.Render("🔄 Replacing binary..."))

	// Determine new binary path
	newBinaryName := "anime"
	if runtime.GOOS == "windows" {
		newBinaryName = "anime.exe"
	}
	newBinary := filepath.Join(BuildDir, "build", newBinaryName)

	// Check if new binary exists
	if _, err := os.Stat(newBinary); os.IsNotExist(err) {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ New binary not found"))
		fmt.Printf("  Expected: %s\n", theme.HighlightStyle.Render(newBinary))
		fmt.Println()
		return fmt.Errorf("new binary not found: %s", newBinary)
	}

	// Create backup of current binary
	backupPath := currentExe + ".backup"
	if err := updateCopyFile(currentExe, backupPath, 0755); err != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  Failed to create backup"))
		fmt.Println(theme.DimTextStyle.Render("Continuing without backup..."))
		fmt.Println()
	} else {
		fmt.Println(theme.DimTextStyle.Render("  Created backup: " + backupPath))
	}

	// Copy new binary to current location
	if err := updateCopyFile(newBinary, currentExe, 0755); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to replace binary"))
		fmt.Println()

		// Try to restore backup
		if _, statErr := os.Stat(backupPath); statErr == nil {
			fmt.Println(theme.InfoStyle.Render("Attempting to restore backup..."))
			if restoreErr := updateCopyFile(backupPath, currentExe, 0755); restoreErr != nil {
				fmt.Println(theme.ErrorStyle.Render("❌ Failed to restore backup!"))
				fmt.Printf("  Backup location: %s\n", theme.HighlightStyle.Render(backupPath))
			} else {
				fmt.Println(theme.SuccessStyle.Render("✓ Backup restored"))
			}
		}
		fmt.Println()
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Make executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("⚠️  Failed to set executable permissions"))
			fmt.Println()
		}
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Binary replaced successfully"))
	fmt.Println()

	// Remove backup on success
	if _, err := os.Stat(backupPath); err == nil {
		os.Remove(backupPath)
	}

	// Step 7: Success message
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.SuccessStyle.Render("🎉 Update completed successfully!"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Run 'anime --version' to see the new version."))
	fmt.Println()

	return nil
}

// Helper functions
func updateContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func updateCopyFile(src, dst string, perm os.FileMode) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceData, perm)
}
