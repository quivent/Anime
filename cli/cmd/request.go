package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request anime capabilities on machines",
	Long:  `Parent command for requesting anime capabilities (access, retrieval, etc.)`,
	Run:   runRequestHelp,
}

func runRequestHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("REQUEST"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Request anime capabilities on machines"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime request access", "Request SSH access to a machine"},
		{"anime request access <server>", "Request access to a specific server"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Examples"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime request access"))
	fmt.Println(theme.DimTextStyle.Render("    Add anime's public key to local authorized_keys"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime request access lambda"))
	fmt.Println(theme.DimTextStyle.Render("    Request access to lambda server (uses alias)"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime request access ubuntu@10.0.0.5"))
	fmt.Println(theme.DimTextStyle.Render("    Request access via user@host"))
	fmt.Println()
}

var requestAccessCmd = &cobra.Command{
	Use:   "access [server]",
	Short: "Request access to a machine for anime",
	Long: `Deposits anime's embedded public key into authorized_keys.

When run without arguments, adds the key locally.
When given a server (alias or user@host), uses password authentication to deposit the key.

Examples:
  anime request access                  # Add key locally
  anime request access alice            # Request access to alice (uses alias)
  anime request access ubuntu@10.0.0.5  # Request access via user@host`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRequestAccess,
}

func init() {
	requestCmd.AddCommand(requestAccessCmd)
	rootCmd.AddCommand(requestCmd)
}

func runRequestAccess(cmd *cobra.Command, args []string) error {
	pubKey := ssh.GetEmbeddedPublicKeyString()

	if len(args) == 1 {
		return requestAccessRemote(args[0], pubKey)
	}
	return requestAccessLocal(pubKey)
}

func requestAccessLocal(pubKey string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔑 Adding anime public key..."))
	fmt.Println()

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(home, ".ssh")
	authKeysPath := filepath.Join(sshDir, "authorized_keys")

	// Create .ssh directory if it doesn't exist
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Check if key already exists
	if exists, err := keyExistsInFile(authKeysPath, pubKey); err == nil && exists {
		fmt.Println(theme.SuccessStyle.Render("  ✓ Key already present"))
		fmt.Println()
		return nil
	}

	// Append the key
	f, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer f.Close()

	// Add newline before key if file is not empty
	info, _ := f.Stat()
	if info.Size() > 0 {
		f.WriteString("\n")
	}
	f.WriteString(pubKey + "\n")

	fmt.Println(theme.SuccessStyle.Render("  ✓ Key added"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  anime can now SSH to this machine"))
	fmt.Println()

	return nil
}

func requestAccessRemote(server, pubKey string) error {
	// Resolve alias to user@host
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	target, err := resolveSSHTarget(cfg, server)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔑 Requesting access..."))
	fmt.Println()
	if server != target {
		fmt.Printf("  %s %s (%s)\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(server), theme.DimTextStyle.Render(target))
	} else {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
	}
	fmt.Println()

	// Deposit the key using ssh-copy-id style approach with interactive password
	depositCmd := fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo 'Key deposited'", pubKey)

	fmt.Println(theme.DimTextStyle.Render("  Enter password when prompted:"))
	fmt.Println()

	sshCmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "PreferredAuthentications=password,keyboard-interactive",
		target,
		depositCmd,
	)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("failed to deposit key: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  ✓ Access granted!"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Connect with: anime ssh "+server))
	fmt.Println()

	return nil
}

func keyExistsInFile(path, key string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(key) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// RequestAccessForServer prompts user to request access and does so if they confirm
func RequestAccessForServer(server string) error {
	fmt.Println()
	fmt.Printf("  %s\n", theme.WarningStyle.Render("⚠ No access to "+server))
	fmt.Println()
	fmt.Printf("  %s", theme.InfoStyle.Render("Request access now? [y/N] "))

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		pubKey := ssh.GetEmbeddedPublicKeyString()
		return requestAccessRemote(server, pubKey)
	}

	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Run manually: anime request access "+server))
	fmt.Println()
	return fmt.Errorf("access denied to %s", server)
}
