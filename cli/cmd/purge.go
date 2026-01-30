package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	purgeAll    bool
	purgePip    bool
	purgeLogs   bool
	purgeTemp   bool
	purgeModels bool
	purgeJobs   bool
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Clean up cache, logs, and temporary files",
	Long: `Remove cached files, logs, and temporary data to free up space.

Options:
  --pip       Clean pip cache and broken packages
  --logs      Remove old log files
  --temp      Clean temporary files
  --models    Remove downloaded model cache (HuggingFace, etc)
  --jobs      Clean up stale PID files and job data
  --all       Clean everything

Examples:
  anime purge --pip              # Clean pip cache
  anime purge --logs             # Remove old logs
  anime purge --all              # Clean everything
  anime purge --pip --logs       # Clean multiple`,
	RunE: runPurge,
}

func init() {
	purgeCmd.Flags().BoolVar(&purgeAll, "all", false, "Clean everything")
	purgeCmd.Flags().BoolVar(&purgePip, "pip", false, "Clean pip cache and broken packages")
	purgeCmd.Flags().BoolVar(&purgeLogs, "logs", false, "Remove old log files")
	purgeCmd.Flags().BoolVar(&purgeTemp, "temp", false, "Clean temporary files")
	purgeCmd.Flags().BoolVar(&purgeModels, "models", false, "Remove model cache (HuggingFace)")
	purgeCmd.Flags().BoolVar(&purgeJobs, "jobs", false, "Clean stale PID files and job data")
	rootCmd.AddCommand(purgeCmd)
}

func runPurge(cmd *cobra.Command, args []string) error {
	// If no flags, show help
	if !purgeAll && !purgePip && !purgeLogs && !purgeTemp && !purgeModels && !purgeJobs {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ No purge options specified"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
		fmt.Println(theme.HighlightStyle.Render("  anime purge --pip              # Clean pip cache"))
		fmt.Println(theme.HighlightStyle.Render("  anime purge --logs             # Remove old logs"))
		fmt.Println(theme.HighlightStyle.Render("  anime purge --all              # Clean everything"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Available options:"))
		fmt.Println(theme.DimTextStyle.Render("  --pip       Clean pip cache and broken packages"))
		fmt.Println(theme.DimTextStyle.Render("  --logs      Remove old log files"))
		fmt.Println(theme.DimTextStyle.Render("  --temp      Clean temporary files"))
		fmt.Println(theme.DimTextStyle.Render("  --models    Remove model cache (HuggingFace)"))
		fmt.Println(theme.DimTextStyle.Render("  --jobs      Clean stale PID files"))
		fmt.Println(theme.DimTextStyle.Render("  --all       Clean everything"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🧹 PURGE"))
	fmt.Println()

	var errors []string
	var successes []string

	// Clean pip cache
	if purgeAll || purgePip {
		fmt.Println(theme.InfoStyle.Render("🗑️  Cleaning pip cache..."))
		if err := cleanPipCache(); err != nil {
			errors = append(errors, fmt.Sprintf("pip cache: %v", err))
			fmt.Println(theme.ErrorStyle.Render("  ❌ Failed"))
		} else {
			successes = append(successes, "pip cache")
			fmt.Println(theme.SuccessStyle.Render("  ✓ Cleaned pip cache"))
		}
		fmt.Println()
	}

	// Clean logs
	if purgeAll || purgeLogs {
		fmt.Println(theme.InfoStyle.Render("📝 Cleaning logs..."))
		if count, err := cleanLogs(); err != nil {
			errors = append(errors, fmt.Sprintf("logs: %v", err))
			fmt.Println(theme.ErrorStyle.Render("  ❌ Failed"))
		} else {
			successes = append(successes, fmt.Sprintf("logs (%d files)", count))
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Removed %d log files", count)))
		}
		fmt.Println()
	}

	// Clean temp files
	if purgeAll || purgeTemp {
		fmt.Println(theme.InfoStyle.Render("🗂️  Cleaning temporary files..."))
		if count, err := cleanTemp(); err != nil {
			errors = append(errors, fmt.Sprintf("temp: %v", err))
			fmt.Println(theme.ErrorStyle.Render("  ❌ Failed"))
		} else {
			successes = append(successes, fmt.Sprintf("temp files (%d files)", count))
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Removed %d temp files", count)))
		}
		fmt.Println()
	}

	// Clean model cache
	if purgeAll || purgeModels {
		fmt.Println(theme.InfoStyle.Render("🤖 Cleaning model cache..."))
		if size, err := cleanModelCache(); err != nil {
			errors = append(errors, fmt.Sprintf("models: %v", err))
			fmt.Println(theme.ErrorStyle.Render("  ❌ Failed"))
		} else {
			successes = append(successes, fmt.Sprintf("model cache (%s)", size))
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Freed %s", size)))
		}
		fmt.Println()
	}

	// Clean stale jobs
	if purgeAll || purgeJobs {
		fmt.Println(theme.InfoStyle.Render("⚙️  Cleaning stale jobs..."))
		if count, err := cleanStaleJobs(); err != nil {
			errors = append(errors, fmt.Sprintf("jobs: %v", err))
			fmt.Println(theme.ErrorStyle.Render("  ❌ Failed"))
		} else {
			successes = append(successes, fmt.Sprintf("stale jobs (%d files)", count))
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Cleaned %d stale job files", count)))
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	if len(successes) > 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ Successfully cleaned:"))
		for _, s := range successes {
			fmt.Printf("  • %s\n", theme.InfoStyle.Render(s))
		}
	}
	if len(errors) > 0 {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to clean:"))
		for _, e := range errors {
			fmt.Printf("  • %s\n", theme.DimTextStyle.Render(e))
		}
	}
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	return nil
}

func cleanPipCache() error {
	// Clean pip cache
	cmd := exec.Command("pip3", "cache", "purge")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try alternative method
		homeDir := os.Getenv("HOME")
		cacheDir := filepath.Join(homeDir, ".cache", "pip")
		if err := os.RemoveAll(cacheDir); err != nil {
			return err
		}
	}

	// Also clean invalid distribution warnings
	cmd = exec.Command("bash", "-c", `find ~/.local/lib/python*/site-packages -name '-*' -type d -exec rm -rf {} + 2>/dev/null || true`)
	cmd.Run()

	_ = output // Suppress unused warning
	return nil
}

func cleanLogs() (int, error) {
	count := 0
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return 0, fmt.Errorf("HOME not set")
	}

	patterns := []string{
		filepath.Join(homeDir, "*.log"),
		filepath.Join(homeDir, "workflow-*.log"),
		filepath.Join(homeDir, "animation-*.log"),
		filepath.Join(homeDir, "comfyui.log"),
		filepath.Join(homeDir, "ollama.log"),
		filepath.Join(homeDir, "jupyter.log"),
		"/tmp/comfyui.log",
		"/tmp/*.log",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, file := range matches {
			if err := os.Remove(file); err == nil {
				count++
			}
		}
	}

	return count, nil
}

func cleanTemp() (int, error) {
	count := 0
	patterns := []string{
		"/tmp/anime-*",
		"/tmp/*-requirements-filtered.txt",
		"/tmp/pip-*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, file := range matches {
			if err := os.RemoveAll(file); err == nil {
				count++
			}
		}
	}

	return count, nil
}

func cleanModelCache() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "0B", fmt.Errorf("HOME not set")
	}

	cachePath := filepath.Join(homeDir, ".cache", "huggingface")

	// Get size before deletion
	cmd := exec.Command("du", "-sh", cachePath)
	output, _ := cmd.Output()
	size := strings.Fields(string(output))
	sizeStr := "unknown"
	if len(size) > 0 {
		sizeStr = size[0]
	}

	// Remove cache
	if err := os.RemoveAll(cachePath); err != nil {
		return sizeStr, err
	}

	return sizeStr, nil
}

func cleanStaleJobs() (int, error) {
	count := 0
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return 0, fmt.Errorf("HOME not set")
	}

	patterns := []string{
		filepath.Join(homeDir, "workflow-*.pid"),
		filepath.Join(homeDir, "animation-*.pid"),
		filepath.Join(homeDir, "serve.pid"),
		filepath.Join(homeDir, "comfyui.pid"),
		filepath.Join(homeDir, "ollama.pid"),
		filepath.Join(homeDir, "jupyter.pid"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, pidFile := range matches {
			// Check if process is still running
			data, err := os.ReadFile(pidFile)
			if err != nil {
				continue
			}
			pid := strings.TrimSpace(string(data))

			// Check if PID exists
			cmd := exec.Command("ps", "-p", pid)
			if err := cmd.Run(); err != nil {
				// Process not running, remove PID file
				os.Remove(pidFile)
				count++
			}
		}
	}

	return count, nil
}
