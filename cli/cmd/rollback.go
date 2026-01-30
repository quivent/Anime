package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	confirmRollback bool
	rollbackSession string
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <server-name> [modules...]",
	Short: "Rollback installed modules on a server",
	Long: `Rollback previously installed modules to restore the system to a clean state.

This command can either:
  1. Rollback specific modules by name
  2. Rollback an entire installation session using --session flag
  3. List available rollback sessions with --list flag

Examples:
  anime rollback lambda-1 pytorch comfyui --confirm    # Rollback specific modules
  anime rollback lambda-1 --session <id> --confirm     # Rollback entire session
  anime rollback lambda-1 --list                       # List available sessions

IMPORTANT: Rollback removes files, packages, and services. Always use --confirm flag.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("server name required")
		}
		return nil
	},
	RunE: runRollback,
}

func init() {
	rollbackCmd.Flags().BoolVar(&confirmRollback, "confirm", false, "Confirm rollback operation (required for safety)")
	rollbackCmd.Flags().StringVar(&rollbackSession, "session", "", "Rollback entire installation session by ID")
	rollbackCmd.Flags().Bool("list", false, "List available rollback sessions")
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	serverName := args[0]
	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server %s not found", serverName)
	}

	// Create SSH client
	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serverName, err)
	}
	defer client.Close()

	// Create installer for rollback operations
	inst := installer.New(client)
	inst.SetServerName(serverName)

	// Handle --list flag
	listFlag, _ := cmd.Flags().GetBool("list")
	if listFlag {
		return listRollbackSessions(client, serverName)
	}

	// Determine rollback mode
	var snapshots []*installer.ModuleSnapshot

	if rollbackSession != "" {
		// Rollback by session ID
		state, err := installer.LoadRollbackState(client, rollbackSession)
		if err != nil {
			return fmt.Errorf("failed to load rollback session: %w", err)
		}
		snapshots = state.Snapshots

		fmt.Println()
		fmt.Println(theme.HeaderStyle.Render(fmt.Sprintf("Rollback Session: %s", state.SessionID)))
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Server: %s", state.ServerName)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Created: %s", state.StartTime.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Modules: %d", len(state.Snapshots))))
		fmt.Println()

	} else if len(args) > 1 {
		// Rollback specific modules
		moduleIDs := args[1:]
		fmt.Println()
		fmt.Println(theme.HeaderStyle.Render("Manual Rollback"))
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Server: %s", serverName)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Modules: %s", strings.Join(moduleIDs, ", "))))
		fmt.Println()

		// Create snapshots for manual rollback
		for _, modID := range moduleIDs {
			// Get module paths
			paths := installer.GetModulePaths(modID)

			// Create a snapshot for rollback
			// Note: We mark everything for removal since this is a manual rollback
			snapshot := &installer.ModuleSnapshot{
				ModuleID:        modID,
				InstalledPaths:  paths.GetAllPaths(),
				PreInstallState: make(map[string]bool),
				PythonPackages:  paths.PythonPackages,
				SystemdServices: paths.SystemdServices,
			}

			// Mark all paths as installed (didn't exist before) for removal
			for _, path := range snapshot.InstalledPaths {
				snapshot.PreInstallState[path] = false
			}

			snapshots = append(snapshots, snapshot)
		}

	} else {
		return fmt.Errorf("specify modules to rollback or use --session flag")
	}

	// Show preview
	preview := inst.GetRollbackPreview(snapshots)
	fmt.Println(theme.InfoStyle.Render("Rollback Preview:"))
	fmt.Println(preview)

	// Require confirmation
	if !confirmRollback {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("Rollback requires --confirm flag for safety"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("To proceed, run:"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("  anime rollback %s %s --confirm",
			serverName, strings.Join(args[1:], " "))))
		fmt.Println()
		return fmt.Errorf("rollback not confirmed")
	}

	// Execute rollback
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Starting rollback..."))
	fmt.Println()

	// Monitor progress
	progressChan := inst.GetProgressChannel()
	go func() {
		for update := range progressChan {
			if update.Error != nil {
				fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("[%s] %s: %v", update.Module, update.Status, update.Error)))
			} else if update.Output != "" {
				fmt.Println(theme.DimTextStyle.Render(update.Output))
			} else {
				fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("[%s] %s", update.Module, update.Status)))
			}
		}
	}()

	if err := inst.Rollback(snapshots); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("Rollback failed: %v", err)))
		fmt.Println()
		return err
	}

	// Delete rollback state if rolling back a session
	if rollbackSession != "" {
		if err := installer.DeleteRollbackState(client, rollbackSession); err != nil {
			fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("Warning: failed to delete rollback state: %v", err)))
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Rollback completed successfully"))
	fmt.Println()

	return nil
}

func listRollbackSessions(client *ssh.Client, serverName string) error {
	states, err := installer.ListRollbackStates(client)
	if err != nil {
		return fmt.Errorf("failed to list rollback sessions: %w", err)
	}

	if len(states) == 0 {
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("No rollback sessions found"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render(fmt.Sprintf("Rollback Sessions for %s", serverName)))
	fmt.Println()

	for _, state := range states {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("Session ID: %s", state.SessionID)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Server: %s", state.ServerName)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Created: %s", state.StartTime.Format("2006-01-02 15:04:05"))))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Modules: %d", len(state.Snapshots))))

		if len(state.Snapshots) > 0 {
			moduleNames := make([]string, 0, len(state.Snapshots))
			for _, snap := range state.Snapshots {
				moduleNames = append(moduleNames, snap.ModuleID)
			}
			fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  -> %s", strings.Join(moduleNames, ", "))))
		}

		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("To rollback a session:"))
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("  anime rollback %s --session <session-id> --confirm", serverName)))
	fmt.Println()

	return nil
}
