package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var serviceServer string

var serviceCmd = &cobra.Command{
	Use:   "service <action> <name>",
	Short: "Manage systemd services",
	Long: `Start, stop, restart, enable, disable, or check status of systemd services.

Actions: start, stop, restart, enable, disable, status

Examples:
  anime service status nginx
  anime service restart nginx -s wings
  anime service enable postgresql -s wings
  anime service stop ollama`,
	Args: cobra.ExactArgs(2),
	RunE: runServiceCmd,
}

func init() {
	serviceCmd.Flags().StringVarP(&serviceServer, "server", "s", "", "Remote server")
	rootCmd.AddCommand(serviceCmd)
}

func runServiceCmd(cmd *cobra.Command, args []string) error {
	action := args[0]
	name := args[1]

	if err := validate.ShellSafe(name); err != nil {
		return fmt.Errorf("invalid service name: %w", err)
	}

	validActions := map[string]string{
		"start":   "sudo systemctl start %s && echo 'Started %s'",
		"stop":    "sudo systemctl stop %s && echo 'Stopped %s'",
		"restart": "sudo systemctl restart %s && echo 'Restarted %s'",
		"enable":  "sudo systemctl enable %s && echo 'Enabled %s'",
		"disable": "sudo systemctl disable %s && echo 'Disabled %s'",
		"status":  "systemctl status %s --no-pager 2>&1",
	}

	tmpl, ok := validActions[action]
	if !ok {
		return fmt.Errorf("unknown action %q (try: start, stop, restart, enable, disable, status)", action)
	}

	script := fmt.Sprintf(tmpl, name, name)

	fmt.Printf("  %s %s %s...\n", theme.SymbolLoading, action, theme.HighlightStyle.Render(name))

	output, err := runOnServer(serviceServer, script)
	if output != "" {
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
	}
	if err != nil && action != "status" {
		return fmt.Errorf("%s failed: %w", action, err)
	}
	if action != "status" {
		fmt.Printf("  %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess+" Done"))
	}
	return nil
}
