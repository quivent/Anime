package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/devlog"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	devLogShowAll     bool
	devLogCategory    string
	devLogLimit       int
	devLogCycleID     string
	devLogStartCycle  bool
	devLogCompleteCycle bool
	devLogAddFeature  bool
	devLogAddChange   bool
	devLogFeatureName string
	devLogFeatureDesc string
	devLogChangeType  string
	devLogChangeDesc  string
	devLogChangeFiles string
	devLogImpact      string
)

var logsDevCmd = &cobra.Command{
	Use:   "development",
	Short: "Development cycle tracking and history",
	Long: `Track and view development cycles, features, and changes.

The development log helps document the evolution of the anime CLI,
tracking each development cycle, features added, and code changes.

Management Commands:
  --start --name "Name" --desc "Description"    Start a new development cycle
  --complete                                     Complete the active cycle
  --add-feature --name "Name" --desc "Desc"     Add a new feature
  --add-change --type add --desc "Desc"         Add a new change`,
	Aliases: []string{"dev"},
	RunE:    runDevLogCommand,
}

func runDevLogCommand(cmd *cobra.Command, args []string) error {
	log, err := devlog.Load()
	if err != nil {
		return fmt.Errorf("failed to load development log: %w", err)
	}

	// Handle management commands
	if devLogStartCycle {
		return handleStartCycle(log)
	}
	if devLogCompleteCycle {
		return handleCompleteCycle(log)
	}
	if devLogAddFeature {
		return handleAddFeature(log)
	}
	if devLogAddChange {
		return handleAddChange(log)
	}

	// Default: show overview
	showDevLogOverview(cmd, args)
	return nil
}

func handleStartCycle(log *devlog.DevLog) error {
	if devLogFeatureName == "" {
		return fmt.Errorf("--name is required to start a cycle")
	}

	// Check for existing active cycle
	if active := log.GetActiveCycle(); active != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  There is already an active cycle:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(active.Name))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Complete it first with: anime logs dev --complete"))
		return nil
	}

	cycle := log.AddCycle(devLogFeatureName, devLogFeatureDesc)
	if err := log.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 NEW CYCLE STARTED 🚀"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Name:"),
		theme.SuccessStyle.Render(cycle.Name))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("ID:"),
		theme.DimTextStyle.Render(cycle.ID))
	if cycle.Description != "" {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Description:"),
			theme.DimTextStyle.Render(cycle.Description))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  All new features and changes will be tracked in this cycle."))
	fmt.Println()
	return nil
}

func handleCompleteCycle(log *devlog.DevLog) error {
	active := log.GetActiveCycle()
	if active == nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  No active cycle to complete."))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Start a new cycle with: anime logs dev --start --name \"Name\""))
		return nil
	}

	if err := log.CompleteCycle(active.ID); err != nil {
		return fmt.Errorf("failed to complete cycle: %w", err)
	}
	if err := log.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("✅ CYCLE COMPLETED ✅"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Name:"),
		theme.SuccessStyle.Render(active.Name))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Features:"),
		theme.HighlightStyle.Render(fmt.Sprintf("%d", len(active.Features))))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Changes:"),
		theme.HighlightStyle.Render(fmt.Sprintf("%d", len(active.Changes))))
	fmt.Println()
	return nil
}

func handleAddFeature(log *devlog.DevLog) error {
	if devLogFeatureName == "" {
		return fmt.Errorf("--name is required to add a feature")
	}

	var files []string
	if devLogChangeFiles != "" {
		files = strings.Split(devLogChangeFiles, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}
	}

	category := devLogCategory
	if category == "" {
		category = "general"
	}

	feature := log.AddFeature(devLogFeatureName, devLogFeatureDesc, category, nil, files)
	if err := log.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Feature added!"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Name:"),
		theme.HighlightStyle.Render(feature.Name))
	if feature.Description != "" {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Description:"),
			theme.DimTextStyle.Render(feature.Description))
	}
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Category:"),
		theme.DimTextStyle.Render(feature.Category))
	if feature.CycleID != "" {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Cycle:"),
			theme.DimTextStyle.Render("attached to active cycle"))
	}
	fmt.Println()
	return nil
}

func handleAddChange(log *devlog.DevLog) error {
	if devLogChangeType == "" {
		return fmt.Errorf("--type is required (add, modify, remove, fix, refactor)")
	}
	if devLogFeatureDesc == "" {
		return fmt.Errorf("--desc is required for change description")
	}

	validTypes := map[string]bool{
		"add": true, "modify": true, "remove": true, "fix": true, "refactor": true,
	}
	if !validTypes[devLogChangeType] {
		return fmt.Errorf("invalid change type: %s (use: add, modify, remove, fix, refactor)", devLogChangeType)
	}

	var files []string
	if devLogChangeFiles != "" {
		files = strings.Split(devLogChangeFiles, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}
	}

	change := log.AddChange(devLogChangeType, devLogFeatureDesc, files, devLogImpact)
	if err := log.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	icon := getChangeTypeIcon(change.Type)
	fmt.Println()
	fmt.Printf("%s %s\n", icon, theme.SuccessStyle.Render("Change recorded!"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Type:"),
		getChangeTypeStyle(change.Type).Render(change.Type))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Description:"),
		theme.DimTextStyle.Render(change.Description))
	if len(change.Files) > 0 {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Files:"),
			theme.DimTextStyle.Render(strings.Join(change.Files, ", ")))
	}
	if change.Impact != "" {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Impact:"),
			getImpactStyle(change.Impact).Render(change.Impact))
	}
	if change.CycleID != "" {
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Cycle:"),
			theme.DimTextStyle.Render("attached to active cycle"))
	}
	fmt.Println()
	return nil
}

func showDevLogOverview(cmd *cobra.Command, args []string) {
	log, err := devlog.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to load development log: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🔮 DEVELOPMENT LOG 🔮"))
	fmt.Println()

	// Show stats
	stats := log.GetStats()
	fmt.Println(theme.GlowStyle.Render("📊 Overview:"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Development Cycles:"),
		theme.SuccessStyle.Render(fmt.Sprintf("%d total (%d active, %d completed)",
			stats["total_cycles"], stats["active_cycles"], stats["completed_cycles"])))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Features Developed:"),
		theme.SuccessStyle.Render(fmt.Sprintf("%d", stats["total_features"])))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("Total Changes:"),
		theme.SuccessStyle.Render(fmt.Sprintf("%d", stats["total_changes"])))
	fmt.Println()

	// Show active cycle if any
	if active := log.GetActiveCycle(); active != nil {
		fmt.Println(theme.GlowStyle.Render("🌸 Active Cycle:"))
		fmt.Println()
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Name:"),
			theme.SuccessStyle.Render(active.Name))
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Started:"),
			theme.DimTextStyle.Render(formatTimeAgo(active.StartTime)))
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Description:"),
			theme.DimTextStyle.Render(active.Description))
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Features:"),
			theme.HighlightStyle.Render(fmt.Sprintf("%d", len(active.Features))))
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Changes:"),
			theme.HighlightStyle.Render(fmt.Sprintf("%d", len(active.Changes))))
		fmt.Println()
	}

	// Show recent changes
	if changes := log.GetRecentChanges(3); len(changes) > 0 {
		fmt.Println(theme.GlowStyle.Render("⚡ Recent Changes:"))
		fmt.Println()
		for _, c := range changes {
			typeIcon := getChangeTypeIcon(c.Type)
			fmt.Printf("  %s %s\n",
				typeIcon,
				theme.InfoStyle.Render(c.Description))
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(formatTimeAgo(c.Timestamp)))
		}
		fmt.Println()
	}

	// Show commands
	fmt.Println(theme.GlowStyle.Render("✨ Commands:"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime logs dev list"),
		theme.DimTextStyle.Render("- List all development cycles"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime logs dev last"),
		theme.DimTextStyle.Render("- Show last development activity"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime logs dev features"),
		theme.DimTextStyle.Render("- List developed features"))
	fmt.Printf("  %s  %s\n",
		theme.HighlightStyle.Render("anime logs dev changelist"),
		theme.DimTextStyle.Render("- View change history"))
	fmt.Println()
}

// List command - show all development cycles
var logsDevListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all development cycles",
	Long: `Display all development cycles with their status, timeline, and summary.

Use --all to show detailed information including features and changes for each cycle.`,
	Aliases: []string{"ls"},
	Run:     runDevLogList,
}

func runDevLogList(cmd *cobra.Command, args []string) {
	log, err := devlog.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to load development log: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("📋 DEVELOPMENT CYCLES 📋"))
	fmt.Println()

	cycles := log.ListCycles()
	if len(cycles) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No development cycles recorded yet."))
		fmt.Println()
		fmt.Printf("  Start a new cycle: %s\n",
			theme.HighlightStyle.Render("anime logs dev --start --name \"Cycle Name\""))
		fmt.Println()
		return
	}

	for i, cycle := range cycles {
		// Apply limit if set
		if devLogLimit > 0 && i >= devLogLimit {
			break
		}

		// Status indicator
		statusIcon := "✅"
		statusStyle := theme.SuccessStyle
		if cycle.Status == "active" {
			statusIcon = "🔄"
			statusStyle = theme.InfoStyle
		} else if cycle.Status == "abandoned" {
			statusIcon = "❌"
			statusStyle = theme.ErrorStyle
		}

		fmt.Printf("  %s %s\n",
			statusIcon,
			theme.HeaderStyle.Render(cycle.Name))
		fmt.Printf("    %s  %s\n",
			theme.DimTextStyle.Render("ID:"),
			theme.DimTextStyle.Render(cycle.ID))
		fmt.Printf("    %s  %s\n",
			theme.DimTextStyle.Render("Status:"),
			statusStyle.Render(cycle.Status))
		fmt.Printf("    %s  %s\n",
			theme.DimTextStyle.Render("Started:"),
			theme.InfoStyle.Render(cycle.StartTime.Format("2006-01-02 15:04")))

		if !cycle.EndTime.IsZero() {
			duration := cycle.EndTime.Sub(cycle.StartTime)
			fmt.Printf("    %s  %s (%s)\n",
				theme.DimTextStyle.Render("Completed:"),
				theme.InfoStyle.Render(cycle.EndTime.Format("2006-01-02 15:04")),
				theme.DimTextStyle.Render(formatCycleDuration(duration)))
		}

		fmt.Printf("    %s  %s\n",
			theme.DimTextStyle.Render("Description:"),
			theme.DimTextStyle.Render(cycle.Description))

		// Show summary counts
		fmt.Printf("    %s  %s features, %s changes\n",
			theme.DimTextStyle.Render("Summary:"),
			theme.HighlightStyle.Render(fmt.Sprintf("%d", len(cycle.Features))),
			theme.HighlightStyle.Render(fmt.Sprintf("%d", len(cycle.Changes))))

		// Show details if --all flag
		if devLogShowAll {
			if len(cycle.Features) > 0 {
				features := log.GetFeaturesByIDs(cycle.Features)
				fmt.Println(theme.InfoStyle.Render("    Features:"))
				for _, f := range features {
					fmt.Printf("      %s %s\n",
						theme.SuccessStyle.Render("→"),
						theme.DimTextStyle.Render(f.Name))
				}
			}
			if len(cycle.Changes) > 0 {
				changes := log.GetChangesByIDs(cycle.Changes)
				fmt.Println(theme.InfoStyle.Render("    Changes:"))
				for _, c := range changes {
					icon := getChangeTypeIcon(c.Type)
					fmt.Printf("      %s %s\n",
						icon,
						theme.DimTextStyle.Render(c.Description))
				}
			}
		}
		fmt.Println()
	}
}

// Last command - show last development activity
var logsDevLastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show last development activity",
	Long: `Display the most recent development activity including:
- Last active or completed cycle
- Most recent changes
- Recently added features`,
	Run: runDevLogLast,
}

func runDevLogLast(cmd *cobra.Command, args []string) {
	log, err := devlog.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to load development log: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("⏱️ LAST DEVELOPMENT ⏱️"))
	fmt.Println()

	// Show active cycle or last completed
	if active := log.GetActiveCycle(); active != nil {
		fmt.Println(theme.GlowStyle.Render("🌸 Current Active Cycle:"))
		fmt.Println()
		showCycleDetail(log, active)
	} else if last := log.GetLastCycle(); last != nil {
		fmt.Println(theme.GlowStyle.Render("📦 Last Completed Cycle:"))
		fmt.Println()
		showCycleDetail(log, last)
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No development cycles recorded yet."))
		fmt.Println()
	}

	// Show last change
	if change := log.GetLastChange(); change != nil {
		fmt.Println(theme.GlowStyle.Render("⚡ Most Recent Change:"))
		fmt.Println()
		showChangeDetail(change)
	}

	// Show recently added features (last 3)
	features := log.ListFeatures()
	if len(features) > 0 {
		fmt.Println(theme.GlowStyle.Render("✨ Recently Added Features:"))
		fmt.Println()
		limit := 3
		if len(features) < limit {
			limit = len(features)
		}
		for _, f := range features[:limit] {
			showFeatureSummary(&f)
		}
	}
}

// Features command - list developed features
var logsDevFeaturesCmd = &cobra.Command{
	Use:   "features",
	Short: "List developed features",
	Long: `Display all features that have been developed and documented.

Use --category to filter by feature category (e.g., command, utility, ui).
Use --all to show detailed information including files and commands.`,
	Aliases: []string{"feat"},
	Run:     runDevLogFeatures,
}

func runDevLogFeatures(cmd *cobra.Command, args []string) {
	log, err := devlog.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to load development log: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("✨ DEVELOPED FEATURES ✨"))
	fmt.Println()

	var features []devlog.Feature
	if devLogCategory != "" {
		features = log.ListFeaturesByCategory(devLogCategory)
		fmt.Printf("  %s %s\n\n",
			theme.InfoStyle.Render("Category:"),
			theme.HighlightStyle.Render(devLogCategory))
	} else {
		features = log.ListFeatures()
	}

	if len(features) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No features recorded yet."))
		fmt.Println()
		fmt.Printf("  Add a feature: %s\n",
			theme.HighlightStyle.Render("anime logs dev --add-feature --name \"Feature Name\" --desc \"Description\""))
		fmt.Println()
		return
	}

	// Group by category if not filtered
	if devLogCategory == "" {
		categories := make(map[string][]devlog.Feature)
		for _, f := range features {
			cat := f.Category
			if cat == "" {
				cat = "uncategorized"
			}
			categories[cat] = append(categories[cat], f)
		}

		for cat, catFeatures := range categories {
			fmt.Printf("  %s %s\n",
				getCategoryIcon(cat),
				theme.HeaderStyle.Render(strings.Title(cat)))
			fmt.Println()
			for i, f := range catFeatures {
				if devLogLimit > 0 && i >= devLogLimit {
					break
				}
				if devLogShowAll {
					showFeatureDetail(&f)
				} else {
					showFeatureSummary(&f)
				}
			}
		}
	} else {
		for i, f := range features {
			if devLogLimit > 0 && i >= devLogLimit {
				break
			}
			if devLogShowAll {
				showFeatureDetail(&f)
			} else {
				showFeatureSummary(&f)
			}
		}
	}
}

// Changelist command - view change history
var logsDevChangelistCmd = &cobra.Command{
	Use:   "changelist",
	Short: "View change history",
	Long: `Display the history of code changes.

Use --cycle to filter changes by a specific development cycle.
Use --limit to restrict the number of changes shown.
Use --all to show detailed information including affected files.`,
	Aliases: []string{"changes", "cl"},
	Run:     runDevLogChangelist,
}

func runDevLogChangelist(cmd *cobra.Command, args []string) {
	log, err := devlog.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to load development log: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("📝 CHANGE HISTORY 📝"))
	fmt.Println()

	var changes []devlog.Change
	if devLogCycleID != "" {
		changes = log.ListChangesByCycle(devLogCycleID)
		cycle, err := log.GetCycle(devLogCycleID)
		if err == nil {
			fmt.Printf("  %s %s\n\n",
				theme.InfoStyle.Render("Cycle:"),
				theme.HighlightStyle.Render(cycle.Name))
		}
	} else {
		changes = log.ListChanges()
	}

	if len(changes) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No changes recorded yet."))
		fmt.Println()
		fmt.Printf("  Add a change: %s\n",
			theme.HighlightStyle.Render("anime logs dev --add-change --type add --desc \"Description\""))
		fmt.Println()
		return
	}

	// Group by date
	dateGroups := make(map[string][]devlog.Change)
	for _, c := range changes {
		date := c.Timestamp.Format("2006-01-02")
		dateGroups[date] = append(dateGroups[date], c)
	}

	// Get unique sorted dates
	var dates []string
	dateSet := make(map[string]bool)
	for _, c := range changes {
		date := c.Timestamp.Format("2006-01-02")
		if !dateSet[date] {
			dateSet[date] = true
			dates = append(dates, date)
		}
	}

	changeCount := 0
	for _, date := range dates {
		dayChanges := dateGroups[date]

		// Parse and format the date nicely
		t, _ := time.Parse("2006-01-02", date)
		dayLabel := t.Format("Monday, January 2, 2006")

		fmt.Printf("  %s %s\n",
			theme.SymbolSakura,
			theme.HeaderStyle.Render(dayLabel))
		fmt.Println()

		for _, c := range dayChanges {
			if devLogLimit > 0 && changeCount >= devLogLimit {
				return
			}
			changeCount++

			if devLogShowAll {
				showChangeDetail(&c)
			} else {
				showChangeSummary(&c)
			}
		}
		fmt.Println()
	}
}

// Helper functions

func showCycleDetail(log *devlog.DevLog, cycle *devlog.DevelopmentCycle) {
	statusIcon := "✅"
	if cycle.Status == "active" {
		statusIcon = "🔄"
	} else if cycle.Status == "abandoned" {
		statusIcon = "❌"
	}

	fmt.Printf("  %s %s\n",
		statusIcon,
		theme.HeaderStyle.Render(cycle.Name))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render(cycle.Description))
	fmt.Println()
	fmt.Printf("    %s  %s\n",
		theme.InfoStyle.Render("Started:"),
		theme.DimTextStyle.Render(cycle.StartTime.Format("2006-01-02 15:04")))
	if !cycle.EndTime.IsZero() {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render("Completed:"),
			theme.DimTextStyle.Render(cycle.EndTime.Format("2006-01-02 15:04")))
	}
	fmt.Printf("    %s  %d features, %d changes\n",
		theme.InfoStyle.Render("Progress:"),
		len(cycle.Features), len(cycle.Changes))
	fmt.Println()

	// Show features
	if len(cycle.Features) > 0 {
		features := log.GetFeaturesByIDs(cycle.Features)
		fmt.Println(theme.InfoStyle.Render("    Features in this cycle:"))
		for _, f := range features {
			fmt.Printf("      %s %s - %s\n",
				theme.SuccessStyle.Render("✨"),
				theme.HighlightStyle.Render(f.Name),
				theme.DimTextStyle.Render(f.Description))
		}
		fmt.Println()
	}

	// Show recent changes
	if len(cycle.Changes) > 0 {
		changes := log.GetChangesByIDs(cycle.Changes)
		fmt.Println(theme.InfoStyle.Render("    Recent changes:"))
		limit := 5
		if len(changes) < limit {
			limit = len(changes)
		}
		for _, c := range changes[:limit] {
			icon := getChangeTypeIcon(c.Type)
			fmt.Printf("      %s %s\n",
				icon,
				theme.DimTextStyle.Render(c.Description))
		}
		if len(changes) > limit {
			fmt.Printf("      %s\n",
				theme.DimTextStyle.Render(fmt.Sprintf("... and %d more", len(changes)-limit)))
		}
		fmt.Println()
	}
}

func showFeatureSummary(f *devlog.Feature) {
	fmt.Printf("    %s %s\n",
		theme.SuccessStyle.Render("✨"),
		theme.HighlightStyle.Render(f.Name))
	fmt.Printf("      %s\n", theme.DimTextStyle.Render(f.Description))
	fmt.Printf("      %s\n", theme.DimTextStyle.Render(formatTimeAgo(f.AddedAt)))
	fmt.Println()
}

func showFeatureDetail(f *devlog.Feature) {
	fmt.Printf("    %s %s\n",
		theme.SuccessStyle.Render("✨"),
		theme.HeaderStyle.Render(f.Name))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("ID:"),
		theme.DimTextStyle.Render(f.ID))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("Description:"),
		theme.DimTextStyle.Render(f.Description))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("Category:"),
		theme.HighlightStyle.Render(f.Category))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("Added:"),
		theme.DimTextStyle.Render(f.AddedAt.Format("2006-01-02 15:04")))

	if len(f.Commands) > 0 {
		fmt.Printf("      %s  %s\n",
			theme.InfoStyle.Render("Commands:"),
			theme.HighlightStyle.Render(strings.Join(f.Commands, ", ")))
	}
	if len(f.Files) > 0 {
		fmt.Println(theme.InfoStyle.Render("      Files:"))
		for _, file := range f.Files {
			fmt.Printf("        %s %s\n",
				theme.DimTextStyle.Render("→"),
				theme.DimTextStyle.Render(file))
		}
	}
	fmt.Println()
}

func showChangeSummary(c *devlog.Change) {
	icon := getChangeTypeIcon(c.Type)
	fmt.Printf("    %s %s\n",
		icon,
		theme.InfoStyle.Render(c.Description))
	fmt.Printf("      %s\n",
		theme.DimTextStyle.Render(c.Timestamp.Format("15:04")))
}

func showChangeDetail(c *devlog.Change) {
	icon := getChangeTypeIcon(c.Type)
	fmt.Printf("    %s %s\n",
		icon,
		theme.HeaderStyle.Render(c.Description))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("ID:"),
		theme.DimTextStyle.Render(c.ID))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("Type:"),
		getChangeTypeStyle(c.Type).Render(c.Type))
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render("Time:"),
		theme.DimTextStyle.Render(c.Timestamp.Format("2006-01-02 15:04:05")))
	if c.Impact != "" {
		fmt.Printf("      %s  %s\n",
			theme.InfoStyle.Render("Impact:"),
			getImpactStyle(c.Impact).Render(c.Impact))
	}
	if len(c.Files) > 0 {
		fmt.Println(theme.InfoStyle.Render("      Files:"))
		for _, file := range c.Files {
			fmt.Printf("        %s %s\n",
				theme.DimTextStyle.Render("→"),
				theme.DimTextStyle.Render(file))
		}
	}
	fmt.Println()
}

func getChangeTypeIcon(changeType string) string {
	switch changeType {
	case "add":
		return theme.SuccessStyle.Render("✚")
	case "modify":
		return theme.InfoStyle.Render("✎")
	case "remove":
		return theme.ErrorStyle.Render("✖")
	case "fix":
		return theme.WarningStyle.Render("🔧")
	case "refactor":
		return theme.HighlightStyle.Render("♻")
	default:
		return theme.DimTextStyle.Render("•")
	}
}

func getChangeTypeStyle(changeType string) lipgloss.Style {
	switch changeType {
	case "add":
		return theme.SuccessStyle
	case "modify":
		return theme.InfoStyle
	case "remove":
		return theme.ErrorStyle
	case "fix":
		return theme.WarningStyle
	case "refactor":
		return theme.HighlightStyle
	default:
		return theme.DimTextStyle
	}
}

func getImpactStyle(impact string) lipgloss.Style {
	switch impact {
	case "major":
		return theme.ErrorStyle
	case "minor":
		return theme.WarningStyle
	case "patch":
		return theme.SuccessStyle
	default:
		return theme.DimTextStyle
	}
}

func getCategoryIcon(category string) string {
	switch category {
	case "command":
		return "⚡"
	case "utility":
		return "🔧"
	case "ui":
		return "🎨"
	case "config":
		return "⚙️"
	case "api":
		return "🔌"
	case "docs":
		return "📚"
	default:
		return "📦"
	}
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else {
		return t.Format("Jan 2, 2006")
	}
}

func formatCycleDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	} else {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%d days", days)
	}
}

func init() {
	logsCmd.AddCommand(logsDevCmd)

	// Add subcommands
	logsDevCmd.AddCommand(logsDevListCmd)
	logsDevCmd.AddCommand(logsDevLastCmd)
	logsDevCmd.AddCommand(logsDevFeaturesCmd)
	logsDevCmd.AddCommand(logsDevChangelistCmd)

	// List command flags
	logsDevListCmd.Flags().BoolVarP(&devLogShowAll, "all", "a", false, "Show detailed information")
	logsDevListCmd.Flags().IntVarP(&devLogLimit, "limit", "n", 0, "Limit number of results")

	// Features command flags
	logsDevFeaturesCmd.Flags().BoolVarP(&devLogShowAll, "all", "a", false, "Show detailed information")
	logsDevFeaturesCmd.Flags().StringVarP(&devLogCategory, "category", "c", "", "Filter by category")
	logsDevFeaturesCmd.Flags().IntVarP(&devLogLimit, "limit", "n", 0, "Limit number of results")

	// Changelist command flags
	logsDevChangelistCmd.Flags().BoolVarP(&devLogShowAll, "all", "a", false, "Show detailed information")
	logsDevChangelistCmd.Flags().StringVar(&devLogCycleID, "cycle", "", "Filter by cycle ID")
	logsDevChangelistCmd.Flags().IntVarP(&devLogLimit, "limit", "n", 0, "Limit number of results")

	// Management flags on parent command
	logsDevCmd.Flags().BoolVar(&devLogStartCycle, "start", false, "Start a new development cycle")
	logsDevCmd.Flags().BoolVar(&devLogCompleteCycle, "complete", false, "Complete the active cycle")
	logsDevCmd.Flags().BoolVar(&devLogAddFeature, "add-feature", false, "Add a new feature")
	logsDevCmd.Flags().BoolVar(&devLogAddChange, "add-change", false, "Add a new change")
	logsDevCmd.Flags().StringVar(&devLogFeatureName, "name", "", "Name for cycle or feature")
	logsDevCmd.Flags().StringVar(&devLogFeatureDesc, "desc", "", "Description")
	logsDevCmd.Flags().StringVarP(&devLogCategory, "category", "c", "", "Feature category")
	logsDevCmd.Flags().StringVar(&devLogChangeType, "type", "", "Change type (add, modify, remove, fix, refactor)")
	logsDevCmd.Flags().StringVar(&devLogChangeFiles, "files", "", "Comma-separated list of affected files")
	logsDevCmd.Flags().StringVar(&devLogImpact, "impact", "", "Change impact (major, minor, patch)")
}
