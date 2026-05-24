package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var firewallServer string

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Manage ufw firewall rules",
	Run:   runFirewallHelp,
}

var firewallStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show firewall status and rules",
	RunE:  runFirewallStatus,
}

var firewallOpenCmd = &cobra.Command{
	Use:   "open <port>",
	Short: "Open a port (allow incoming)",
	Long: `Examples:
  anime firewall open 3000
  anime firewall open 8080 -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runFirewallOpen,
}

var firewallCloseCmd = &cobra.Command{
	Use:   "close <port>",
	Short: "Close a port (deny incoming)",
	Args:  cobra.ExactArgs(1),
	RunE:  runFirewallClose,
}

var firewallAllowCmd = &cobra.Command{
	Use:   "allow <service>",
	Short: "Allow a named service (e.g. Nginx Full, OpenSSH)",
	Args:  cobra.ExactArgs(1),
	RunE:  runFirewallAllow,
}

func init() {
	for _, c := range []*cobra.Command{firewallStatusCmd, firewallOpenCmd, firewallCloseCmd, firewallAllowCmd} {
		c.Flags().StringVarP(&firewallServer, "server", "s", "", "Remote server")
	}
	firewallCmd.AddCommand(firewallStatusCmd)
	firewallCmd.AddCommand(firewallOpenCmd)
	firewallCmd.AddCommand(firewallCloseCmd)
	firewallCmd.AddCommand(firewallAllowCmd)
	rootCmd.AddCommand(firewallCmd)
}

func runFirewallHelp(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("🔥 FIREWALL 🔥"))
	fmt.Println()
	cmds := []struct{ c, d string }{
		{"anime firewall status", "Show firewall rules"},
		{"anime firewall open <port>", "Open a port"},
		{"anime firewall close <port>", "Close a port"},
		{"anime firewall allow <service>", "Allow a service (e.g. 'Nginx Full')"},
	}
	for _, c := range cmds {
		fmt.Printf("  %s\n    %s\n\n", theme.HighlightStyle.Render(c.c), theme.DimTextStyle.Render(c.d))
	}
	fmt.Println(theme.DimTextStyle.Render("  All commands support --server/-s for remote"))
	fmt.Println()
}

func runFirewallStatus(cmd *cobra.Command, args []string) error {
	output, err := runOnServer(firewallServer, "sudo ufw status verbose 2>&1")
	if err != nil && output == "" {
		return fmt.Errorf("failed: %w", err)
	}
	fmt.Println()
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
	}
	fmt.Println()
	return nil
}

func runFirewallOpen(cmd *cobra.Command, args []string) error {
	port := args[0]
	if err := validate.Port(port); err != nil {
		return err
	}
	script := fmt.Sprintf("sudo ufw allow %s && echo 'Opened port %s'", port, port)
	output, err := runOnServer(firewallServer, script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}

func runFirewallClose(cmd *cobra.Command, args []string) error {
	port := args[0]
	if err := validate.Port(port); err != nil {
		return err
	}
	script := fmt.Sprintf("sudo ufw deny %s && echo 'Closed port %s'", port, port)
	output, err := runOnServer(firewallServer, script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}

func runFirewallAllow(cmd *cobra.Command, args []string) error {
	service := args[0]
	if err := validate.ShellSafe(service); err != nil {
		return fmt.Errorf("invalid service name: %w", err)
	}
	script := fmt.Sprintf(`sudo ufw allow '%s' && echo "Allowed %s"`, service, service)
	output, err := runOnServer(firewallServer, script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}
