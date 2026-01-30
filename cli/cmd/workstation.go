package cmd

import (
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var workstationCmd = &cobra.Command{
	Use:   "workstation",
	Short: "Launch the workstation monitoring TUI",
	Long: `Launch a comprehensive Terminal User Interface (TUI) to monitor your workstation.

The workstation TUI provides real-time monitoring of:
  • GPU usage and metrics (temperature, power, memory, utilization)
  • Ollama models (installed models with size and quantization)
  • Installed software (Python, PyTorch, CUDA, Docker, etc.)
  • Asset collections (configured in anime.yaml)
  • Workflows and active tasks
  • System resources (CPU, RAM, disk, uptime)

KEYBOARD SHORTCUTS:
  ↑/↓ or k/j      Scroll within panel
  ←/→ or h/l      Switch between panels
  Tab             Next panel
  Shift+Tab       Previous panel
  r               Refresh data
  ?               Toggle help
  q or Ctrl+C     Quit

EXAMPLE:
  anime workstation       # Launch the monitoring TUI
`,
	RunE: runWorkstation,
}

func init() {
	rootCmd.AddCommand(workstationCmd)
}

func runWorkstation(cmd *cobra.Command, args []string) error {
	return tui.RunWorkstation()
}
