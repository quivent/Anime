package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// WanSnapshot records a setup or reset event for auditability.
type WanSnapshot struct {
	Timestamp time.Time        `json:"timestamp"`
	Action    string           `json:"action"` // "install", "reset", "purge", "fix"
	Phases    []WanPhaseRecord `json:"phases"`
}

// WanPhaseRecord records one phase's state at snapshot time.
type WanPhaseRecord struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Paths  []string `json:"paths"`
	Status string   `json:"status"` // "installed", "removed", "failed", "skipped"
}

// WanSnapshotLog is the on-disk format: an append-only list of snapshots.
type WanSnapshotLog struct {
	Snapshots []WanSnapshot `json:"snapshots"`
}

func wanSnapshotPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".anime", "wan-snapshots.json")
}

// saveWanSnapshot records a snapshot of the given phases to ~/.anime/wan-snapshots.json.
func saveWanSnapshot(phases []wanResetPhase, home string) {
	snap := WanSnapshot{
		Timestamp: time.Now(),
		Action:    "reset",
	}
	for _, ph := range phases {
		paths := ph.paths(home)
		var existPaths []string
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				existPaths = append(existPaths, p)
			}
		}
		status := "removed"
		if len(existPaths) == 0 {
			status = "skipped"
		}
		snap.Phases = append(snap.Phases, WanPhaseRecord{
			ID:     ph.id,
			Name:   ph.name,
			Paths:  existPaths,
			Status: status,
		})
	}
	appendSnapshot(snap)
}

// saveWanInstallSnapshot records a successful install phase.
func saveWanInstallSnapshot(phaseID, phaseName string, paths []string) {
	snap := WanSnapshot{
		Timestamp: time.Now(),
		Action:    "install",
		Phases: []WanPhaseRecord{
			{
				ID:     phaseID,
				Name:   phaseName,
				Paths:  paths,
				Status: "installed",
			},
		},
	}
	appendSnapshot(snap)
}

func appendSnapshot(snap WanSnapshot) {
	path := wanSnapshotPath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)

	var log WanSnapshotLog
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &log)
	}

	log.Snapshots = append(log.Snapshots, snap)

	// Keep last 100 snapshots
	if len(log.Snapshots) > 100 {
		log.Snapshots = log.Snapshots[len(log.Snapshots)-100:]
	}

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0o644)
}

func init() {
	snapshotCmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Show setup/reset history",
		Long:  `Display the log of Wan setup, reset, fix, and purge operations.`,
		RunE:  runWanSnapshots,
	}
	wanCmd.AddCommand(snapshotCmd)
}

func runWanSnapshots(cmd *cobra.Command, args []string) error {
	path := wanSnapshotPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(theme.DimTextStyle.Render("No snapshot history yet."))
		return nil
	}

	var log WanSnapshotLog
	if err := json.Unmarshal(data, &log); err != nil {
		return fmt.Errorf("corrupt snapshot log: %w", err)
	}

	if len(log.Snapshots) == 0 {
		fmt.Println(theme.DimTextStyle.Render("No snapshot history yet."))
		return nil
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("📋 Wan setup history"))
	fmt.Println()

	// Show last 20
	start := 0
	if len(log.Snapshots) > 20 {
		start = len(log.Snapshots) - 20
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  (showing last 20 of %d)", len(log.Snapshots))))
		fmt.Println()
	}

	for i := start; i < len(log.Snapshots); i++ {
		s := log.Snapshots[i]
		actionStyle := theme.InfoStyle
		actionIcon := "·"
		switch s.Action {
		case "install":
			actionIcon = "+"
			actionStyle = theme.SuccessStyle
		case "reset":
			actionIcon = "↺"
			actionStyle = theme.WarningStyle
		case "purge":
			actionIcon = "☢"
			actionStyle = theme.ErrorStyle
		case "fix":
			actionIcon = "🔧"
			actionStyle = theme.InfoStyle
		}

		ts := s.Timestamp.Format("2006-01-02 15:04:05")
		var phaseNames []string
		for _, p := range s.Phases {
			phaseNames = append(phaseNames, p.ID)
		}
		fmt.Printf("  %s %s  %s  %s\n",
			actionStyle.Render(actionIcon),
			theme.DimTextStyle.Render(ts),
			actionStyle.Render(s.Action),
			theme.DimTextStyle.Render(joinShort(phaseNames)))
	}

	fmt.Println()
	return nil
}

func joinShort(names []string) string {
	if len(names) <= 3 {
		return joinNames(names)
	}
	return joinNames(names[:3]) + fmt.Sprintf(" +%d more", len(names)-3)
}

func joinNames(names []string) string {
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
