package cmd

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/state"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	resumeForce bool
)

var resumeCmd = &cobra.Command{
	Use:   "resume [server-name]",
	Short: "Resume an interrupted installation",
	Long: `Resume an interrupted installation from where it left off.

This command will load the saved installation state and continue installing
any remaining modules. Use --force to restart the installation from scratch.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime resume <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime resume lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime resume my-server --force"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Options:"))
			fmt.Println(theme.DimTextStyle.Render("  --force    Restart installation from scratch"))
			fmt.Println()
			return fmt.Errorf("resume requires a server name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		server, err := cfg.GetServer(serverName)
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
			fmt.Println(theme.ErrorStyle.Render("❌ No installation state found"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("No previous installation found for this server."))
			fmt.Println(theme.DimTextStyle.Render("Use 'anime deploy " + serverName + "' to start a new installation."))
			fmt.Println()
			return nil
		}

		// Handle --force flag
		if resumeForce {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("🔄 Restarting installation from scratch"))
			fmt.Println()

			if err := state.ClearState(serverName); err != nil {
				return fmt.Errorf("failed to clear state: %w", err)
			}

			m := tui.NewInstallModel(server)
			p := tea.NewProgram(m)

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running installation: %w", err)
			}

			return nil
		}

		// Check if installation can be resumed
		if !installState.CanResume() {
			fmt.Println()
			if installState.Status == state.StatusCompleted {
				fmt.Println(theme.SuccessStyle.Render("✅ Installation already completed"))
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("Installation Details:"))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Started:   %s", installState.StartTime.Format("2006-01-02 15:04:05"))))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Completed: %s", installState.LastUpdate.Format("2006-01-02 15:04:05"))))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Duration:  %s", installState.GetElapsedTime().Round(time.Second))))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Modules:   %d/%d completed", installState.GetCompletedCount(), installState.GetTotalCount())))
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("💡 Tip: Use --force to restart the installation"))
				fmt.Println()
			} else if installState.Status == state.StatusFailed {
				fmt.Println(theme.ErrorStyle.Render("❌ Previous installation failed"))
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("Installation Details:"))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Started: %s", installState.StartTime.Format("2006-01-02 15:04:05"))))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Failed:  %s", installState.LastUpdate.Format("2006-01-02 15:04:05"))))
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Modules: %d/%d completed, %d failed",
					installState.GetCompletedCount(), installState.GetTotalCount(), installState.GetFailedCount())))
				fmt.Println()
				if len(installState.FailedModules) > 0 {
					fmt.Println(theme.ErrorStyle.Render("Failed Modules:"))
					for modID, modState := range installState.FailedModules {
						errMsg := modState.Error
						if errMsg == "" {
							errMsg = "unknown error"
						}
						fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  - %s: %s", modID, errMsg)))
					}
					fmt.Println()
				}
				fmt.Println(theme.InfoStyle.Render("💡 Tip: Use --force to restart the installation"))
				fmt.Println()
			} else {
				fmt.Println(theme.WarningStyle.Render("⚠️  Installation cannot be resumed"))
				fmt.Println()
				fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Status: %s", installState.Status)))
				fmt.Println(theme.DimTextStyle.Render("Use --force to restart the installation"))
				fmt.Println()
			}
			return nil
		}

		// Display resume information
		pending := installState.GetPendingModules()
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("🔄 Resuming Installation"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Installation Progress:"))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Server:        %s", serverName)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Started:       %s", installState.StartTime.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Last Update:   %s", installState.LastUpdate.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Elapsed:       %s", installState.GetElapsedTime().Round(time.Second))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Progress:      %d/%d modules (%.1f%%)",
			installState.GetCompletedCount(), installState.GetTotalCount(), installState.GetProgress())))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Remaining:     %d modules", len(pending))))

		if installState.GetFailedCount() > 0 {
			fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  Failed:        %d modules", installState.GetFailedCount())))
		}

		fmt.Println()

		// Show pending modules
		if len(pending) > 0 && len(pending) <= 10 {
			fmt.Println(theme.InfoStyle.Render("Pending Modules:"))
			for _, modID := range pending {
				// Find module name
				for _, mod := range config.AvailableModules {
					if mod.ID == modID {
						fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  - %s (%s)", mod.Name, modID)))
						break
					}
				}
			}
			fmt.Println()
		}

		// Start installation
		m := tui.NewInstallModel(server)
		p := tea.NewProgram(m)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running installation: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
	resumeCmd.Flags().BoolVar(&resumeForce, "force", false, "Restart installation from scratch")
}
