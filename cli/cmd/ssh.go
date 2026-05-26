package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh [server]",
	Short: "SSH into a server",
	Long: `SSH into a configured server or lambda instance.

Resolves aliases and server configs automatically.

Examples:
  anime ssh                    # SSH to lambda (default)
  anime ssh lambda             # SSH to lambda server
  anime ssh production         # SSH to production server
  anime ssh ubuntu@10.0.0.5    # SSH to specific host`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSSH,
}

func init() {
	rootCmd.AddCommand(sshCmd)
}

func runSSH(cmd *cobra.Command, args []string) error {
	// Default to lambda
	target := "lambda"
	if len(args) > 0 {
		target = args[0]
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target
	resolvedTarget, err := resolveSSHTarget(cfg, target)
	if err != nil {
		return err
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🔗 Connecting to %s...", resolvedTarget)))
	fmt.Println()

	// Execute SSH command with all identity keys
	sshCmd := sshCommand(resolvedTarget)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

func resolveSSHTarget(cfg *config.Config, target string) (string, error) {
	// Check if it's an alias
	if alias := cfg.GetAlias(target); alias != "" {
		// If alias doesn't have user@, add ubuntu@
		if !strings.Contains(alias, "@") {
			return "ubuntu@" + alias, nil
		}
		return alias, nil
	}

	// Check if it's a server config
	if server, err := cfg.GetServer(target); err == nil {
		return fmt.Sprintf("%s@%s", server.User, server.Host), nil
	}

	// Check if it looks like user@host
	if strings.Contains(target, "@") {
		return target, nil
	}

	// Check if it looks like an IP/hostname
	if strings.Contains(target, ".") {
		return "ubuntu@" + target, nil
	}

	// Check if it resolves as an SSH config alias
	if isSSHConfigAlias(target) {
		return target, nil
	}

	return "", fmt.Errorf("could not resolve target: %s", target)
}

// writeEmbeddedKeyToTemp writes the embedded SSH key to a temp file and returns the path
func writeEmbeddedKeyToTemp() (string, func(), error) {
	keyData := ssh.GetEmbeddedPrivateKey()
	if len(keyData) == 0 {
		return "", nil, fmt.Errorf("no embedded SSH key available")
	}

	tmpFile, err := os.CreateTemp("", "anime-ssh-key-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	if _, err := tmpFile.Write(keyData); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to write key: %w", err)
	}

	if err := tmpFile.Chmod(0600); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to set key permissions: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to close key file: %w", err)
	}

	return tmpFile.Name(), cleanup, nil
}
