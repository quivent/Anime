package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var (
	tunnelLocalPort string
	tunnelServer    string
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel <remote-port> -s <server>",
	Short: "SSH tunnel to access remote ports locally",
	Long: `Create an SSH tunnel to forward a remote port to localhost.

Examples:
  anime tunnel 5432 -s wings              # postgres on localhost:5432
  anime tunnel 6379 -s wings              # redis on localhost:6379
  anime tunnel 8080 -s wings -l 9090      # remote 8080 → local 9090
  anime tunnel 3000 -s wings              # app on localhost:3000`,
	Args: cobra.ExactArgs(1),
	RunE: runTunnel,
}

func init() {
	tunnelCmd.Flags().StringVarP(&tunnelServer, "server", "s", "", "Remote server (required)")
	tunnelCmd.Flags().StringVarP(&tunnelLocalPort, "local", "l", "", "Local port (defaults to same as remote)")
	tunnelCmd.MarkFlagRequired("server")
	rootCmd.AddCommand(tunnelCmd)
}

func runTunnel(cmd *cobra.Command, args []string) error {
	remotePort := args[0]
	if err := validate.Port(remotePort); err != nil {
		return err
	}

	localPort := remotePort
	if tunnelLocalPort != "" {
		if err := validate.Port(tunnelLocalPort); err != nil {
			return err
		}
		localPort = tunnelLocalPort
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var user, host, sshKey string
	target := cfg.GetAlias(tunnelServer)
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
		server, err := cfg.GetServer(tunnelServer)
		if err != nil {
			return fmt.Errorf("server not found: %s", tunnelServer)
		}
		user = server.User
		host = server.Host
		sshKey = server.SSHKey
	}

	sshTarget := fmt.Sprintf("%s@%s", user, host)
	tunnelSpec := fmt.Sprintf("%s:localhost:%s", localPort, remotePort)

	sshArgs := []string{"-N", "-L", tunnelSpec, sshTarget}
	if sshKey != "" {
		sshArgs = append([]string{"-i", sshKey}, sshArgs...)
	}

	fmt.Println(theme.RenderBanner("🚇 TUNNEL 🚇"))
	fmt.Println()
	fmt.Printf("  %s localhost:%s → %s:%s\n",
		theme.SuccessStyle.Render("→"),
		theme.HighlightStyle.Render(localPort),
		theme.InfoStyle.Render(host),
		theme.HighlightStyle.Render(remotePort))
	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Press Ctrl+C to close tunnel"))
	fmt.Println()

	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	// Handle Ctrl+C gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	if err := sshCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tunnel: %w", err)
	}

	go func() {
		<-sig
		fmt.Printf("\n  %s Tunnel closed\n", theme.InfoStyle.Render("→"))
		sshCmd.Process.Kill()
	}()

	return sshCmd.Wait()
}
