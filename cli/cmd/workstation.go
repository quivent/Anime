package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var workstationCmd = &cobra.Command{
	Use:   "workstation",
	Short: "Launch the workstation monitoring TUI",
	Long: `Launch a comprehensive Terminal User Interface (TUI) to monitor your workstation.

The workstation TUI provides real-time monitoring of:
  • GPU usage and metrics
  • Ollama models
  • Installed software and packages
  • Asset collections
  • Workflows and tasks
  • Training libraries
  • System resources

Navigate using arrow keys, tab to switch panels, and 'q' to quit.`,
	RunE: runWorkstation,
}

func init() {
	rootCmd.AddCommand(workstationCmd)
}

func runWorkstation(cmd *cobra.Command, args []string) error {
	// TODO: Implement workstation TUI
	fmt.Println("Workstation TUI coming soon!")
	return nil
}
