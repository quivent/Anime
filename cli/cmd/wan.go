package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

//go:embed embedded/wan-pipeline/wan.py
var wanScriptFS embed.FS

const wanScriptPath = "embedded/wan-pipeline/wan.py"

var wanCmd = &cobra.Command{
	Use:   "wan",
	Short: "Wan 2.2 stateful render pipeline (with memory)",
	Long: `Wan 2.2 stateful render pipeline. Every render's prompt, seed, params,
and output URL is recorded in SQLite at ~/.anime/wan-pipeline.db.

Subcommands:
  anime wan render "..."          Submit a render with the default preset
                                  (add --explicit to drop NSFW gating)
  anime wan history               List past renders
  anime wan show <id>             Detail of a single render
  anime wan resume <id>           Re-render with the same seed (deterministic)
  anime wan vary <id> [-n N]      Same prompt, fresh seeds (variations)
  anime wan rate <id> 1-5         Rate a render
  anime wan presets               Show available render presets
  anime wan models                Show installed Wan models
  anime wan stats                 Pipeline-wide stats
  anime wan tui                   Interactive Bubble Tea TUI
  anime wan studio                Open the Comfort web studio (browser UI)

Setup:
  anime install wan               Full Wan 2.2 stack (driver-aware torch +
                                  sage attn, 14B+5B model set, Comfort UI).
                                  Runs on any CUDA GPU with ≥16GB VRAM.
  anime install comfort           Just the studio UI (clones quivent/comfort).`,
}

func init() {
	rootCmd.AddCommand(wanCmd)

	// Each subcommand is a thin pass-through to the embedded Python script.
	for _, sub := range []string{
		"render", "history", "show", "resume", "vary", "rate",
		"presets", "models", "stats",
	} {
		sub := sub
		c := &cobra.Command{
			Use:                sub,
			Short:              "anime wan " + sub,
			DisableFlagParsing: true, // pass flags straight to wan.py
			RunE: func(cmd *cobra.Command, args []string) error {
				return runWanPython(append([]string{sub}, args...))
			},
		}
		wanCmd.AddCommand(c)
	}

	// Native Go TUI subcommand
	wanCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Bubble Tea TUI for browsing renders + queueing new ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWanTUI()
		},
	})
}

// extractWanScript writes the embedded wan.py to ~/.anime/wan-pipeline/wan.py
// (idempotent — only writes if missing or content differs).
func extractWanScript() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dst := filepath.Join(home, ".anime", "wan-pipeline", "wan.py")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	embedded, err := fs.ReadFile(wanScriptFS, wanScriptPath)
	if err != nil {
		return "", fmt.Errorf("reading embedded wan.py: %w", err)
	}
	if existing, err := os.ReadFile(dst); err == nil && string(existing) == string(embedded) {
		return dst, nil
	}
	if err := os.WriteFile(dst, embedded, 0o755); err != nil {
		return "", fmt.Errorf("writing %s: %w", dst, err)
	}
	return dst, nil
}

// findPython picks the best Python: prefer ComfyUI venv, then python3, then python.
func findPython() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "ComfyUI", "venv", "bin", "python"),
		"python3",
		"python",
	}
	for _, p := range candidates {
		if filepath.IsAbs(p) {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		} else if path, err := exec.LookPath(p); err == nil {
			return path
		}
	}
	return "python3"
}

func runWanPython(args []string) error {
	scriptPath, err := extractWanScript()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ failed to extract wan.py: " + err.Error()))
		return err
	}
	py := findPython()
	full := append([]string{scriptPath}, args...)
	c := exec.Command(py, full...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}
	return nil
}

// ---- light command-output helper for the TUI to reach into the Python CLI ----

// runWanCapture executes wan.py with given args and returns combined stdout.
func runWanCapture(args ...string) (string, error) {
	scriptPath, err := extractWanScript()
	if err != nil {
		return "", err
	}
	py := findPython()
	full := append([]string{scriptPath}, args...)
	c := exec.Command(py, full...)
	c.Env = os.Environ()
	out, err := c.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}

