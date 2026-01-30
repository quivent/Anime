package cmd

import (
	"fmt"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/state"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var installStatusCmd = &cobra.Command{
	Use:   "install-status [server-name]",
	Short: "Show installation status for a server",
	Long: `Display the current installation status for a server, including:
- Overall progress and completion percentage
- List of completed modules
- List of pending modules
- List of failed modules (if any)
- Estimated time remaining
- Installation start time and duration`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime install-status <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime install-status lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime install-status my-server"))
			fmt.Println()
			return fmt.Errorf("install-status requires a server name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		_, err = cfg.GetServer(serverName)
		if err != nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		// Load installation state
		installState, err := state.LoadState(serverName)
		if err != nil {
			return fmt.Errorf("failed to load installation state: %w", err)
		}

		if installState == nil {
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("ℹ️  No installation state found"))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("No installation has been started for this server."))
			fmt.Println(theme.DimTextStyle.Render("Use 'anime deploy " + serverName + "' to start an installation."))
			fmt.Println()
			return nil
		}

		// Display status header
		fmt.Println()
		fmt.Println(theme.TitleStyle.Render("📊 Installation Status: " + serverName))
		fmt.Println()

		// Overall status
		statusIcon := "⏳"
		statusText := string(installState.Status)
		statusStyle := theme.InfoStyle

		switch installState.Status {
		case state.StatusCompleted:
			statusIcon = "✅"
			statusStyle = theme.SuccessStyle
		case state.StatusFailed:
			statusIcon = "❌"
			statusStyle = theme.ErrorStyle
		case state.StatusCancelled:
			statusIcon = "🚫"
			statusStyle = theme.WarningStyle
		case state.StatusInProgress:
			statusIcon = "⏳"
			statusStyle = theme.InfoStyle
		}

		fmt.Println(statusStyle.Render(fmt.Sprintf("%s Status: %s", statusIcon, statusText)))
		fmt.Println()

		// Installation details
		fmt.Println(theme.InfoStyle.Render("Installation Details:"))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Started:       %s", installState.StartTime.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Last Update:   %s", installState.LastUpdate.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Elapsed Time:  %s", installState.GetElapsedTime().Round(time.Second))))

		if installState.Status == state.StatusInProgress {
			timeSinceUpdate := installState.GetTimeSinceLastUpdate()
			fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Idle Time:     %s", timeSinceUpdate.Round(time.Second))))
			if installState.IsStale() {
				fmt.Println(theme.WarningStyle.Render("  ⚠️  Installation appears stale (no update in over 1 hour)"))
			}
		}

		if installState.Parallel {
			fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Mode:          Parallel (jobs: %d)", installState.Jobs)))
		} else {
			fmt.Println(theme.DimTextStyle.Render("  Mode:          Sequential"))
		}

		fmt.Println()

		// Progress
		fmt.Println(theme.InfoStyle.Render("Progress:"))
		progress := installState.GetProgress()
		completed := installState.GetCompletedCount()
		total := installState.GetTotalCount()
		failed := installState.GetFailedCount()

		// Progress bar
		barWidth := 40
		filledWidth := int(progress / 100.0 * float64(barWidth))
		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filledWidth {
				bar += "█"
			} else {
				bar += "░"
			}
		}

		progressColor := theme.SuccessStyle
		if progress < 50 {
			progressColor = theme.WarningStyle
		}
		if progress >= 100 {
			progressColor = theme.SuccessStyle
		}

		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  [%s] %.1f%%", progressColor.Render(bar), progress)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Completed: %d/%d modules", completed, total)))

		if failed > 0 {
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Failed:    %d modules", failed)))
		}

		// Estimate time remaining
		if installState.Status == state.StatusInProgress && completed > 0 {
			pending := len(installState.GetPendingModules())
			if pending > 0 {
				elapsed := installState.GetElapsedTime()
				avgTimePerModule := elapsed / time.Duration(completed)
				estimatedRemaining := avgTimePerModule * time.Duration(pending)
				fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Estimated Remaining: ~%s", estimatedRemaining.Round(time.Minute))))
			}
		}

		fmt.Println()

		// Completed modules
		if len(installState.CompletedModules) > 0 {
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✅ Completed Modules (%d):", len(installState.CompletedModules))))
			for modID := range installState.CompletedModules {
				// Find module name
				moduleName := modID
				for _, mod := range config.AvailableModules {
					if mod.ID == modID {
						moduleName = mod.Name
						break
					}
				}
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  ✓ %s (%s)", moduleName, modID)))
			}
			fmt.Println()
		}

		// Failed modules
		if len(installState.FailedModules) > 0 {
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Failed Modules (%d):", len(installState.FailedModules))))
			for modID, modState := range installState.FailedModules {
				// Find module name
				moduleName := modID
				for _, mod := range config.AvailableModules {
					if mod.ID == modID {
						moduleName = mod.Name
						break
					}
				}
				errMsg := modState.Error
				if errMsg == "" {
					errMsg = "unknown error"
				}
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  ✗ %s (%s)", moduleName, modID)))
				fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("    Error: %s", errMsg)))
			}
			fmt.Println()
		}

		// Pending modules
		pending := installState.GetPendingModules()
		if len(pending) > 0 {
			fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("⏳ Pending Modules (%d):", len(pending))))

			// Show current in-progress module if any
			if installState.InProgressModule != "" {
				moduleName := installState.InProgressModule
				for _, mod := range config.AvailableModules {
					if mod.ID == installState.InProgressModule {
						moduleName = mod.Name
						break
					}
				}
				fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("  ▶ %s (%s) - IN PROGRESS", moduleName, installState.InProgressModule)))
			}

			// Show other pending modules
			for _, modID := range pending {
				if modID != installState.InProgressModule {
					// Find module name
					moduleName := modID
					estimatedTime := ""
					for _, mod := range config.AvailableModules {
						if mod.ID == modID {
							moduleName = mod.Name
							estimatedTime = fmt.Sprintf("~%d min", mod.TimeMinutes)
							break
						}
					}
					fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  ○ %s (%s) %s", moduleName, modID, estimatedTime)))
				}
			}
			fmt.Println()
		}

		// Show actions
		fmt.Println(theme.InfoStyle.Render("💡 Available Actions:"))
		if installState.CanResume() {
			fmt.Println(theme.DimTextStyle.Render("  anime resume " + serverName + "           # Resume installation"))
			fmt.Println(theme.DimTextStyle.Render("  anime resume " + serverName + " --force   # Restart from scratch"))
		} else if installState.Status == state.StatusCompleted {
			fmt.Println(theme.DimTextStyle.Render("  Installation complete! No action needed."))
		} else if installState.Status == state.StatusFailed {
			fmt.Println(theme.DimTextStyle.Render("  anime resume " + serverName + " --force   # Restart from scratch"))
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installStatusCmd)
}
