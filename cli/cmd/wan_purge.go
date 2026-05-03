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

func init() {
	purgeWanCmd := &cobra.Command{
		Use:   "purge",
		Short: "Remove the entire Wan stack (nuclear reset to pre-install state)",
		Long: `Completely remove all Wan-related files and return the system to
the state it was in before 'anime wan studio' was ever run.

Removes:
  ~/ComfyUI/           Render engine, venv, custom nodes, models
  ~/Comfort/           Studio web UI
  ~/.anime/            Pipeline database, logs, snapshots, config
  Screen sessions      comfyui screen session and orphan processes

This is the nuclear option. Use 'anime wan reset <phase>' for surgical removal.

Examples:
  anime wan purge --confirm          # full teardown
  anime wan purge --dry-run          # preview what would be removed
  anime wan purge --keep-db --confirm  # keep render history`,
		DisableFlagParsing: true,
		RunE:               runWanPurge,
	}
	wanCmd.AddCommand(purgeWanCmd)
}

func runWanPurge(cmd *cobra.Command, args []string) error {
	var (
		confirm bool
		dryRun  bool
		keepDB  bool
	)
	for _, a := range args {
		switch a {
		case "--confirm":
			confirm = true
		case "--dry-run":
			dryRun = true
		case "--keep-db":
			keepDB = true
		case "-h", "--help":
			return cmd.Help()
		default:
			return fmt.Errorf("unknown flag: %s", a)
		}
	}

	home, _ := os.UserHomeDir()

	// Catalog everything that will be removed
	type target struct {
		path string
		desc string
	}

	targets := []target{
		{filepath.Join(home, "ComfyUI"), "Render engine + venv + models + custom nodes"},
		{filepath.Join(home, "Comfort"), "Comfort studio UI"},
	}

	animeDir := filepath.Join(home, ".anime")
	if keepDB {
		// Remove individual files but keep the db
		targets = append(targets,
			target{filepath.Join(animeDir, "comfyui.log"), "Engine log"},
			target{filepath.Join(animeDir, "wan-snapshots.json"), "Setup snapshots"},
			target{filepath.Join(animeDir, "comfort-path"), "Comfort path config"},
		)
	} else {
		targets = append(targets,
			target{filepath.Join(animeDir, "comfyui.log"), "Engine log"},
			target{filepath.Join(animeDir, "wan-pipeline.db"), "Render history database"},
			target{filepath.Join(animeDir, "wan-snapshots.json"), "Setup snapshots"},
			target{filepath.Join(animeDir, "comfort-path"), "Comfort path config"},
		)
	}

	// Preview
	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("☢  Wan purge — FULL TEARDOWN"))
	fmt.Println()

	totalSize := ""
	var existingTargets []target
	for _, t := range targets {
		if _, err := os.Stat(t.path); err != nil {
			continue
		}
		existingTargets = append(existingTargets, t)
		size := dirSize(t.path)
		hint := ""
		if size != "" {
			hint = " (" + size + ")"
		}
		fmt.Printf("  %s  %s%s\n", theme.WarningStyle.Render("✗"),
			theme.HighlightStyle.Render(t.desc),
			theme.DimTextStyle.Render(hint))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(t.path))
	}

	// Check for ComfyUI to get total size
	comfyDir := filepath.Join(home, "ComfyUI")
	if _, err := os.Stat(comfyDir); err == nil {
		totalSize = dirSize(comfyDir)
	}

	// Screen sessions
	screenOut, _ := exec.Command("screen", "-ls").CombinedOutput()
	hasComfyScreen := strings.Contains(string(screenOut), "comfyui")
	if hasComfyScreen {
		fmt.Printf("  %s  %s\n", theme.WarningStyle.Render("✗"),
			theme.HighlightStyle.Render("comfyui screen session (will be killed)"))
	}

	fmt.Println()

	if len(existingTargets) == 0 && !hasComfyScreen {
		fmt.Println(theme.SuccessStyle.Render("Nothing to purge — system is already clean."))
		return nil
	}

	if totalSize != "" {
		fmt.Printf("  %s\n", theme.WarningStyle.Render("Total disk to free: ~"+totalSize))
		fmt.Println()
	}

	if keepDB {
		fmt.Println(theme.DimTextStyle.Render("  (--keep-db: render history will be preserved)"))
		fmt.Println()
	}

	if dryRun {
		fmt.Println(theme.DimTextStyle.Render("  (--dry-run: no changes made)"))
		return nil
	}

	if !confirm {
		fmt.Println(theme.ErrorStyle.Render("  This is irreversible. Add --confirm to execute."))
		return fmt.Errorf("aborted (add --confirm)")
	}

	// Save a snapshot before purging
	saveWanSnapshot(wanResetPhases(), home)

	// Step 1: Stop services
	fmt.Println(theme.InfoStyle.Render("→ Stopping services..."))
	stopComfyScreenSession()
	fmt.Println(theme.SuccessStyle.Render("  ✓ Screen sessions and processes cleaned"))

	// Step 2: Remove everything
	fmt.Println(theme.InfoStyle.Render("→ Removing files..."))
	var errors []string
	for _, t := range existingTargets {
		if err := os.RemoveAll(t.path); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", t.path, err))
			fmt.Printf("  %s %s: %v\n", theme.ErrorStyle.Render("✗"), t.desc, err)
		} else {
			fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("✓"), t.desc)
		}
	}

	fmt.Println()

	if len(errors) > 0 {
		fmt.Println(theme.WarningStyle.Render("Purge completed with errors:"))
		for _, e := range errors {
			fmt.Printf("  • %s\n", theme.DimTextStyle.Render(e))
		}
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ Purge complete — system returned to pre-Wan state"))
	}

	fmt.Println(theme.DimTextStyle.Render("  To set up again: anime wan studio --yes"))
	fmt.Println()
	return nil
}
