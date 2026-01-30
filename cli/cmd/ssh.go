package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

	// First, test if we can connect without password
	testCmd := exec.Command("ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", resolvedTarget, "echo ok")
	if err := testCmd.Run(); err != nil {
		// Connection failed - likely no key. Offer to copy it.
		fmt.Println(theme.WarningStyle.Render("⚠ SSH key not authorized on server"))
		fmt.Println()
		fmt.Print(theme.InfoStyle.Render("Copy SSH key to server? [Y/n] "))

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "" || response == "y" || response == "yes" {
			if err := copySSHKey(resolvedTarget); err != nil {
				fmt.Println(theme.WarningStyle.Render("⚠ Could not auto-copy key: " + err.Error()))
				fmt.Println(theme.DimTextStyle.Render("  You may need to enter your password manually"))
				fmt.Println()
			} else {
				fmt.Println(theme.SuccessStyle.Render("✓ SSH key copied successfully"))
				fmt.Println()
			}
		}
	}

	// Execute SSH command
	sshExec := exec.Command("ssh", resolvedTarget)
	sshExec.Stdin = os.Stdin
	sshExec.Stdout = os.Stdout
	sshExec.Stderr = os.Stderr

	return sshExec.Run()
}

// copySSHKey copies the user's public key to the remote server
func copySSHKey(target string) error {
	// Find public key
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Check for public keys in order of preference
	pubKeys := []string{
		home + "/.ssh/id_ed25519.pub",
		home + "/.ssh/id_rsa.pub",
		home + "/.ssh/id_ecdsa.pub",
	}

	var pubKeyPath string
	for _, p := range pubKeys {
		if _, err := os.Stat(p); err == nil {
			pubKeyPath = p
			break
		}
	}

	if pubKeyPath == "" {
		return fmt.Errorf("no SSH public key found")
	}

	// Read public key
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return err
	}

	// Copy to server using ssh-copy-id style command
	fmt.Println(theme.DimTextStyle.Render("  Copying " + pubKeyPath + "..."))

	copyCmd := exec.Command("ssh", target,
		fmt.Sprintf(`mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`, strings.TrimSpace(string(pubKey))))
	copyCmd.Stdin = os.Stdin
	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr

	return copyCmd.Run()
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
