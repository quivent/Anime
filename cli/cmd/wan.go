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
and output URL is recorded in SQLite at ~/.anime/wan-pipeline.db.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printWanHelp()
		return nil
	},
}

// subcommand metadata — short descriptions that actually tell you what the command does.
var wanSubDescs = map[string]string{
	"render":  "Submit a text-to-video render",
	"history": "List recent renders",
	"show":    "Show full details of a render",
	"resume":  "Re-render with the same seed (deterministic replay)",
	"vary":    "Generate variations (same prompt, fresh seeds)",
	"rate":    "Rate a render 1-5 stars",
	"presets": "Show available render presets",
	"models":  "Show installed Wan models",
	"stats":   "Pipeline statistics dashboard",
}

func init() {
	rootCmd.AddCommand(wanCmd)

	// Each subcommand is a thin pass-through to the embedded Python script.
	for _, sub := range []string{
		"render", "history", "show", "resume", "vary", "rate",
		"presets", "models", "stats",
	} {
		sub := sub
		desc := wanSubDescs[sub]
		c := &cobra.Command{
			Use:                sub,
			Short:              desc,
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
		Short: "Interactive terminal UI for browsing and queueing renders",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWanTUI()
		},
	})
}

// printWanHelp renders a grouped, scannable help screen.
// The goal: a new user reads this once and knows exactly what to do.
func printWanHelp() {
	pink := "\033[38;5;213m"
	cyan := "\033[38;5;51m"
	dim := "\033[2m"
	bold := "\033[1m"
	green := "\033[38;5;42m"
	yellow := "\033[38;5;220m"
	reset := "\033[0m"

	fmt.Println()
	fmt.Printf("  %s%swan%s  %sStateful Wan 2.2 render pipeline%s\n", bold, pink, reset, dim, reset)
	fmt.Printf("  %sEvery render is recorded in ~/.anime/wan-pipeline.db%s\n", dim, reset)
	fmt.Println()

	// -- Creation --
	fmt.Printf("  %s%sCreate%s\n", bold, cyan, reset)
	fmt.Printf("    %srender%s  %s<prompt>%s     Submit a text-to-video render\n", bold, reset, dim, reset)
	fmt.Printf("    %svary%s    %s<id>%s         Generate variations (same prompt, new seeds)\n", bold, reset, dim, reset)
	fmt.Printf("    %sresume%s  %s<id>%s         Re-render with the exact same seed\n", bold, reset, dim, reset)
	fmt.Println()

	// -- Review --
	fmt.Printf("  %s%sReview%s\n", bold, cyan, reset)
	fmt.Printf("    %shistory%s              List recent renders\n", bold, reset)
	fmt.Printf("    %sshow%s    %s<id>%s         Full details of a single render\n", bold, reset, dim, reset)
	fmt.Printf("    %sstats%s                Pipeline statistics dashboard\n", bold, reset)
	fmt.Printf("    %srate%s    %s<id> <1-5>%s   Rate a render\n", bold, reset, dim, reset)
	fmt.Println()

	// -- Setup --
	fmt.Printf("  %s%sSetup%s\n", bold, cyan, reset)
	fmt.Printf("    %sstudio%s               Launch the Comfort web studio (browser UI)\n", bold, reset)
	fmt.Printf("    %smodels%s               Show installed Wan models\n", bold, reset)
	fmt.Printf("    %spresets%s              Show available render presets\n", bold, reset)
	fmt.Printf("    %stui%s                  Interactive terminal UI\n", bold, reset)
	fmt.Println()

	// -- Quick start examples --
	fmt.Printf("  %s%sQuick start%s\n", bold, yellow, reset)
	fmt.Printf("    %s$%s anime wan render %s\"a cat dancing in the rain\"%s\n", green, reset, dim, reset)
	fmt.Printf("    %s$%s anime wan history\n", green, reset)
	fmt.Printf("    %s$%s anime wan vary 1 %s-n 3%s\n", green, reset, dim, reset)
	fmt.Printf("    %s$%s anime wan studio\n", green, reset)
	fmt.Println()

	// -- First-time setup --
	fmt.Printf("  %sFirst time? Run:%s  anime install wan\n", dim, reset)
	fmt.Println()
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
		fmt.Println(theme.ErrorStyle.Render("✗ Failed to extract wan.py: " + err.Error()))
		fmt.Println(theme.DimTextStyle.Render("  Check disk space and permissions on ~/.anime/"))
		return err
	}
	py := findPython()

	// Verify Python is actually reachable before launching — a missing interpreter
	// produces a confusing "exec: not found" error otherwise.
	if !filepath.IsAbs(py) {
		if _, lookErr := exec.LookPath(py); lookErr != nil {
			fmt.Println(theme.ErrorStyle.Render("✗ Python not found on PATH"))
			fmt.Println(theme.DimTextStyle.Render("  The wan pipeline needs Python 3.8+."))
			fmt.Println(theme.DimTextStyle.Render("  Install it:  anime install wan"))
			fmt.Println(theme.DimTextStyle.Render("  Or manually: brew install python3  /  apt install python3"))
			return fmt.Errorf("python not found")
		}
	}

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
		// Generic execution failure — give the user something to work with.
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("✗ wan pipeline exited with an error"))
		fmt.Println(theme.DimTextStyle.Render("  Python: " + py))
		fmt.Println(theme.DimTextStyle.Render("  Script: " + scriptPath))
		fmt.Println(theme.DimTextStyle.Render("  If ComfyUI is not running: anime comfyui start"))
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

