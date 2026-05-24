package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var healthServer string

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Server health check — disk, memory, CPU, uptime, services",
	Long: `Quick health snapshot of a server.

Examples:
  anime health                  # Check local machine
  anime health -s wings         # Check remote server`,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().StringVarP(&healthServer, "server", "s", "", "Remote server to check")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	fmt.Println(theme.RenderBanner("💊 HEALTH CHECK 💊"))
	fmt.Println()

	script := `#!/bin/bash
echo "=== UPTIME ==="
uptime

echo ""
echo "=== MEMORY ==="
free -h 2>/dev/null || vm_stat 2>/dev/null

echo ""
echo "=== DISK ==="
df -h / /home 2>/dev/null | head -5

echo ""
echo "=== CPU ==="
nproc 2>/dev/null && echo "cores"
cat /proc/loadavg 2>/dev/null || sysctl -n vm.loadavg 2>/dev/null

echo ""
echo "=== SERVICES ==="
for svc in nginx sshd docker ollama postgresql mysql redis; do
    if systemctl is-active "$svc" &>/dev/null; then
        echo "  ✓ $svc"
    elif systemctl is-enabled "$svc" &>/dev/null; then
        echo "  ✗ $svc (stopped)"
    fi
done 2>/dev/null

echo ""
echo "=== LISTENING PORTS ==="
ss -tlnp 2>/dev/null | grep LISTEN | awk '{print $4}' | sed 's/.*://' | sort -un | head -15 || \
    netstat -tlnp 2>/dev/null | grep LISTEN | awk '{print $4}' | sed 's/.*://' | sort -un | head -15
`

	output, err := runOnServer(healthServer, script)
	if err != nil && output == "" {
		return fmt.Errorf("health check failed: %w", err)
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(line, "===") {
			fmt.Printf("  %s\n", theme.InfoStyle.Render(line))
		} else if strings.HasPrefix(line, "  ✓") {
			fmt.Printf("  %s\n", theme.SuccessStyle.Render(line))
		} else if strings.HasPrefix(line, "  ✗") {
			fmt.Printf("  %s\n", theme.WarningStyle.Render(line))
		} else {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
	}
	fmt.Println()
	return nil
}

// runOnServer runs a script locally or on a remote server.
// Shared helper for health, service, firewall, backup, etc.
func runOnServer(serverName, script string) (string, error) {
	if serverName == "" {
		cmd := exec.Command("bash", "-c", script)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	var user, host, sshKey string
	target := cfg.GetAlias(serverName)
	if target != "" {
		if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			user = parts[0]
			host = parts[1]
		} else {
			user = "ubuntu"
			host = target
		}
	} else {
		server, err := cfg.GetServer(serverName)
		if err != nil {
			return "", fmt.Errorf("server not found: %s", serverName)
		}
		user = server.User
		host = server.Host
		sshKey = server.SSHKey
	}

	client, err := ssh.NewClient(host, user, sshKey)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	return client.RunCommand(script)
}
