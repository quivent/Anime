package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var doctorSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Test SSH connectivity to all configured servers",
	Long:  `Tests SSH connectivity, key auth, and latency to every configured server.`,
	RunE:  runDoctorSSH,
}

func init() {
	doctorCmd.AddCommand(doctorSSHCmd)
}

func runDoctorSSH(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println(theme.InfoStyle.Render("🏥 SSH Health Check"))
	fmt.Println()

	servers := cfg.Servers
	if len(servers) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No servers configured"))
		return nil
	}

	okCount := 0
	failCount := 0

	for _, srv := range servers {
		target := srv.User + "@" + srv.Host
		if srv.User == "" {
			target = "ubuntu@" + srv.Host
		}

		start := time.Now()
		testArgs := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			"-o", "BatchMode=yes",
		}

		// Try embedded key first
		if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
			testArgs = append(testArgs, "-i", keyPath)
			defer cleanup()
		}

		testArgs = append(testArgs, target, "echo ok")
		testCmd := exec.Command("ssh", testArgs...)
		output, err := testCmd.CombinedOutput()
		elapsed := time.Since(start)

		if err == nil && strings.TrimSpace(string(output)) == "ok" {
			fmt.Printf("  %s  %-12s  %s  %s\n",
				theme.SuccessStyle.Render("✓"),
				theme.HighlightStyle.Render(srv.Name),
				theme.DimTextStyle.Render(srv.Host),
				theme.DimTextStyle.Render(fmt.Sprintf("%dms", elapsed.Milliseconds())))
			okCount++
		} else {
			reason := "unreachable"
			outStr := string(output)
			if strings.Contains(outStr, "Permission denied") {
				reason = "auth failed"
			} else if strings.Contains(outStr, "Connection refused") {
				reason = "refused"
			} else if strings.Contains(outStr, "timed out") {
				reason = "timeout"
			}
			fmt.Printf("  %s  %-12s  %s  %s\n",
				theme.ErrorStyle.Render("✗"),
				theme.HighlightStyle.Render(srv.Name),
				theme.DimTextStyle.Render(srv.Host),
				theme.WarningStyle.Render(reason))
			failCount++
		}
	}

	fmt.Println()
	fmt.Printf("  %d ok, %d failed\n", okCount, failCount)

	if failCount > 0 {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Fix: anime sync"))
	}

	return nil
}
