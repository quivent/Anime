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

// wanResetPhase describes one resettable component of the Wan stack.
type wanResetPhase struct {
	id    string
	name  string
	paths func(home string) []string // directories/files to remove
	pre   func() error               // optional: run before removal (e.g. stop services)
	post  func() error               // optional: run after removal (e.g. clean venv cache)
}

func wanResetPhases() []wanResetPhase {
	home, _ := os.UserHomeDir()
	j := func(parts ...string) string { return filepath.Join(append([]string{home}, parts...)...) }

	return []wanResetPhase{
		{
			id:   "comfort",
			name: "Comfort studio (web UI)",
			paths: func(h string) []string {
				return []string{
					j("Comfort"),
					j(".anime", "comfort-path"),
				}
			},
		},
		{
			id:   "wanmodels",
			name: "Wan 2.2 model weights",
			paths: func(h string) []string {
				return []string{
					j("ComfyUI", "models", "diffusion_models"),
					j("ComfyUI", "models", "text_encoders"),
					j("ComfyUI", "models", "vae"),
					j("ComfyUI", "models", "loras"),
					j("ComfyUI", "models", "clip_vision"),
				}
			},
		},
		{
			id:   "wannodes",
			name: "Wan custom-node stack",
			paths: func(h string) []string {
				return []string{
					j("ComfyUI", "custom_nodes", "ComfyUI-WanVideoWrapper"),
					j("ComfyUI", "custom_nodes", "ComfyUI-KJNodes"),
					j("ComfyUI", "custom_nodes", "ComfyUI-Manager"),
				}
			},
		},
		{
			id:   "wantorch",
			name: "PyTorch + sage attention (venv)",
			paths: func(h string) []string {
				return []string{
					j("ComfyUI", "venv"),
				}
			},
			pre: stopComfyScreenSession,
		},
		{
			id:   "comfyui",
			name: "Render engine (ComfyUI)",
			paths: func(h string) []string {
				return []string{
					j("ComfyUI"),
				}
			},
			pre: stopComfyScreenSession,
		},
	}
}

func init() {
	resetCmd := &cobra.Command{
		Use:   "reset [phase...]",
		Short: "Undo Wan setup phases (surgical rollback)",
		Long: `Remove specific Wan setup phases or the entire stack.

Phases (in dependency order):
  comfort      Remove Comfort studio UI (~100MB)
  wanmodels    Remove Wan 2.2 model weights (20-85GB)
  wannodes     Remove Kijai custom nodes
  wantorch     Remove venv (PyTorch + sageattention)
  comfyui      Remove entire ComfyUI directory

Flags:
  --all        Reset everything (equivalent to: anime wan purge)
  --list       Show installed phases and exit
  --confirm    Required for destructive operations (safety gate)
  --dry-run    Show what would be removed without doing it

Examples:
  anime wan reset --list                      # see what's installed
  anime wan reset wanmodels --confirm         # free model disk space
  anime wan reset wantorch wannodes --confirm # rebuild torch + nodes
  anime wan reset --all --confirm             # full teardown`,
		DisableFlagParsing: true,
		RunE:               runWanReset,
	}
	wanCmd.AddCommand(resetCmd)
}

func runWanReset(cmd *cobra.Command, args []string) error {
	var (
		all     bool
		list    bool
		confirm bool
		dryRun  bool
		targets []string
	)

	for _, a := range args {
		switch a {
		case "--all":
			all = true
		case "--list":
			list = true
		case "--confirm":
			confirm = true
		case "--dry-run":
			dryRun = true
		case "-h", "--help":
			return cmd.Help()
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag: %s", a)
			}
			targets = append(targets, a)
		}
	}

	home, _ := os.UserHomeDir()
	phases := wanResetPhases()

	// --list: show status and exit
	if list || (len(targets) == 0 && !all) {
		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("🔍 Wan stack status"))
		fmt.Println()
		for _, ph := range phases {
			paths := ph.paths(home)
			installed := false
			var sizeStr string
			for _, p := range paths {
				if _, err := os.Stat(p); err == nil {
					installed = true
					if s := dirSize(p); s != "" {
						sizeStr = s
					}
				}
			}
			label := theme.HighlightStyle.Render(fmt.Sprintf("%-14s", ph.id))
			name := theme.DimTextStyle.Render(ph.name)
			if installed {
				hint := ""
				if sizeStr != "" {
					hint = " (" + sizeStr + ")"
				}
				fmt.Printf("  %s %s  %s%s\n", theme.SymbolSuccess, label, name, theme.DimTextStyle.Render(hint))
			} else {
				fmt.Printf("  %s %s  %s\n", theme.DimTextStyle.Render("·"), label, theme.DimTextStyle.Render("not installed"))
			}
		}
		fmt.Println()
		if !list {
			fmt.Println(theme.DimTextStyle.Render("  Usage: anime wan reset <phase...> --confirm"))
			fmt.Println(theme.DimTextStyle.Render("         anime wan reset --all --confirm"))
			fmt.Println()
		}
		return nil
	}

	// Resolve which phases to reset
	var selected []wanResetPhase
	if all {
		selected = phases
	} else {
		phaseMap := make(map[string]wanResetPhase)
		for _, ph := range phases {
			phaseMap[ph.id] = ph
		}
		for _, t := range targets {
			ph, ok := phaseMap[t]
			if !ok {
				valid := make([]string, len(phases))
				for i, p := range phases {
					valid[i] = p.id
				}
				return fmt.Errorf("unknown phase %q (valid: %s)", t, strings.Join(valid, ", "))
			}
			selected = append(selected, ph)
		}
	}

	// Preview
	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("⚠  Wan reset — the following will be removed:"))
	fmt.Println()
	totalPaths := 0
	for _, ph := range selected {
		paths := ph.paths(home)
		var existing []string
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				existing = append(existing, p)
			}
		}
		if len(existing) == 0 {
			fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("·"),
				theme.DimTextStyle.Render(ph.name+" — already removed"))
			continue
		}
		fmt.Printf("  %s  %s\n", theme.WarningStyle.Render("✗"), theme.HighlightStyle.Render(ph.name))
		for _, p := range existing {
			size := dirSize(p)
			hint := ""
			if size != "" {
				hint = " (" + size + ")"
			}
			fmt.Printf("     %s%s\n", theme.DimTextStyle.Render(p), theme.DimTextStyle.Render(hint))
		}
		totalPaths += len(existing)
	}
	fmt.Println()

	if totalPaths == 0 {
		fmt.Println(theme.SuccessStyle.Render("Nothing to remove — all selected phases are already clean."))
		return nil
	}

	if dryRun {
		fmt.Println(theme.DimTextStyle.Render("  (--dry-run: no changes made)"))
		return nil
	}

	if !confirm {
		fmt.Println(theme.ErrorStyle.Render("  Add --confirm to execute. This is irreversible."))
		return fmt.Errorf("aborted (add --confirm)")
	}

	// Execute
	saveWanSnapshot(selected, home)

	for _, ph := range selected {
		fmt.Printf("  %s %s...\n", theme.InfoStyle.Render("→"), ph.name)

		if ph.pre != nil {
			if err := ph.pre(); err != nil {
				fmt.Printf("    %s pre-step: %v (continuing)\n", theme.WarningStyle.Render("⚠"), err)
			}
		}

		paths := ph.paths(home)
		for _, p := range paths {
			if _, err := os.Stat(p); err != nil {
				continue
			}
			if err := os.RemoveAll(p); err != nil {
				fmt.Printf("    %s %s: %v\n", theme.ErrorStyle.Render("✗"), p, err)
			} else {
				fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("✓"), theme.DimTextStyle.Render(p))
			}
		}

		if ph.post != nil {
			if err := ph.post(); err != nil {
				fmt.Printf("    %s post-step: %v\n", theme.WarningStyle.Render("⚠"), err)
			}
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Reset complete"))
	fmt.Println(theme.DimTextStyle.Render("  Re-run: anime wan studio --yes"))
	fmt.Println()
	return nil
}

// stopComfyScreenSession kills the comfyui screen session and any lingering
// ComfyUI python processes.
func stopComfyScreenSession() error {
	// Kill screen session
	exec.Command("screen", "-S", "comfyui", "-X", "quit").Run()

	// Kill any lingering ComfyUI python processes
	exec.Command("pkill", "-f", "ComfyUI.*main.py").Run()

	return nil
}

// dirSize returns a human-readable size string for a path, or "" on error.
func dirSize(path string) string {
	out, err := exec.Command("du", "-sh", path).Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}
