package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

func init() {
	fixCmd := &cobra.Command{
		Use:   "fix [phase...]",
		Short: "Diagnose and repair broken Wan setup phases",
		Long: `Check each Wan setup phase and offer to reinstall broken ones.

Without arguments, checks all phases and repairs any that fail their health
check. With arguments, only checks and repairs the named phases.

Unlike 'anime wan studio' which skips satisfied phases, fix will optionally
tear down a phase before reinstalling it — useful when a phase is partially
installed or corrupted.

Flags:
  --force     Force reinstall even if phase check passes
  --check     Only diagnose, don't repair
  --yes       Don't prompt for confirmation

Examples:
  anime wan fix                    # diagnose all, repair broken
  anime wan fix wantorch           # fix just the torch/sageattention phase
  anime wan fix --force wannodes   # tear down + reinstall custom nodes
  anime wan fix --check            # just show what's broken`,
		DisableFlagParsing: true,
		RunE:               runWanFix,
	}
	wanCmd.AddCommand(fixCmd)
}

func runWanFix(cmd *cobra.Command, args []string) error {
	var (
		force   bool
		check   bool
		yes     bool
		targets []string
	)

	for _, a := range args {
		switch a {
		case "--force":
			force = true
		case "--check":
			check = true
		case "--yes", "-y":
			yes = true
		case "-h", "--help":
			return cmd.Help()
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag: %s", a)
			}
			targets = append(targets, a)
		}
	}

	// Auto-detect install level from what's on disk
	level := detectInstalledLevel()

	phases := wanStudioPhases(level)

	// Filter to requested targets if any
	if len(targets) > 0 {
		phaseMap := make(map[string]phase)
		for _, ph := range phases {
			if ph.id != "" {
				phaseMap[ph.id] = ph
			}
		}
		var filtered []phase
		for _, t := range targets {
			ph, ok := phaseMap[t]
			if !ok {
				valid := make([]string, 0)
				for _, p := range phases {
					if p.id != "" {
						valid = append(valid, p.id)
					}
				}
				return fmt.Errorf("unknown phase %q (valid: %s)", t, strings.Join(valid, ", "))
			}
			filtered = append(filtered, ph)
		}
		phases = filtered
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🔧 Wan fix — diagnosing setup phases"))
	fmt.Println()

	type diagnosis struct {
		ph     phase
		ok     bool
		detail string
	}

	var results []diagnosis
	var broken []diagnosis

	for _, ph := range phases {
		if ph.id == "" {
			continue // skip the "server running" pseudo-phase
		}
		ok, detail := ph.check()
		d := diagnosis{ph: ph, ok: ok, detail: detail}
		results = append(results, d)

		label := theme.HighlightStyle.Render(fmt.Sprintf("%-32s", ph.name))
		if ok && !force {
			fmt.Printf("  %s %s  %s\n", theme.SymbolSuccess, label,
				theme.DimTextStyle.Render(detail))
		} else if ok && force {
			fmt.Printf("  %s %s  %s\n", theme.WarningStyle.Render("↻"), label,
				theme.DimTextStyle.Render(detail+" (--force: will reinstall)"))
			broken = append(broken, d)
		} else {
			fmt.Printf("  %s %s  %s\n", theme.SymbolWarning, label,
				theme.WarningStyle.Render(detail))
			broken = append(broken, d)
		}
	}

	fmt.Println()

	if len(broken) == 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ All phases healthy — nothing to fix"))
		return nil
	}

	if check {
		fmt.Printf(theme.WarningStyle.Render("  %d phase(s) need repair\n"), len(broken))
		fmt.Println(theme.DimTextStyle.Render("  Run without --check to fix them"))
		return nil
	}

	// Confirm
	if !yes {
		fmt.Printf(theme.WarningStyle.Render("  %d phase(s) will be repaired"), len(broken))
		if force {
			fmt.Print(theme.WarningStyle.Render(" (--force: includes tear-down)"))
		}
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("  Continue? [y/N] "))
		var ans string
		fmt.Scanln(&ans)
		if !strings.EqualFold(strings.TrimSpace(ans), "y") &&
			!strings.EqualFold(strings.TrimSpace(ans), "yes") {
			return fmt.Errorf("aborted")
		}
	}

	fmt.Println()

	// If --force, tear down before reinstalling
	if force {
		resetPhases := wanResetPhases()
		resetMap := make(map[string]wanResetPhase)
		for _, rp := range resetPhases {
			resetMap[rp.id] = rp
		}

		home, _ := os.UserHomeDir()
		for _, d := range broken {
			if rp, ok := resetMap[d.ph.id]; ok {
				fmt.Printf("  %s Tearing down %s...\n", theme.InfoStyle.Render("→"), d.ph.name)
				if rp.pre != nil {
					rp.pre()
				}
				for _, p := range rp.paths(home) {
					os.RemoveAll(p)
				}
			}
		}
		fmt.Println()
	}

	// Reinstall broken phases
	for _, d := range broken {
		fmt.Printf("  %s Reinstalling %s...\n", theme.InfoStyle.Render("→"), d.ph.name)
		fmt.Println()

		var err error
		if d.ph.custom != nil {
			err = d.ph.custom(&setupOpts{yes: true})
		} else if d.ph.id != "" {
			err = runInstallScript(d.ph.id)
		}

		if err != nil {
			fmt.Printf("  %s %s failed: %v\n", theme.ErrorStyle.Render("✗"), d.ph.name, err)
			fmt.Println(theme.DimTextStyle.Render("    Remaining phases skipped."))
			return fmt.Errorf("fix failed at phase %q: %w", d.ph.id, err)
		}

		// Verify
		if ok, detail := d.ph.check(); !ok {
			fmt.Printf("  %s %s reinstalled but check still fails: %s\n",
				theme.ErrorStyle.Render("✗"), d.ph.name, detail)
			return fmt.Errorf("phase %q completed but check still fails: %s", d.ph.id, detail)
		}

		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("✓"),
			theme.SuccessStyle.Render(d.ph.name+" — fixed"))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("✓ All broken phases repaired"))
	fmt.Println()
	return nil
}

// detectInstalledLevel determines the install level from what models are on disk.
func detectInstalledLevel() string {
	home, _ := os.UserHomeDir()
	j := func(parts ...string) string { return filepath.Join(append([]string{home}, parts...)...) }

	// Check for full level markers
	fullMarkers := modelsRequiredForLevel("full")
	allFull := true
	for _, m := range fullMarkers {
		if !exists(j("ComfyUI", m.rel)) {
			allFull = false
			break
		}
	}
	if allFull {
		return "full"
	}

	// Check for standard level markers
	stdMarkers := modelsRequiredForLevel("standard")
	allStd := true
	for _, m := range stdMarkers {
		if !exists(j("ComfyUI", m.rel)) {
			allStd = false
			break
		}
	}
	if allStd {
		return "standard"
	}

	// Fall back to recommended based on VRAM
	return recommendedLevel(gpu.GetTotalVRAM())
}
