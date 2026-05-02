package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var ghServer string

var ghCmd = &cobra.Command{
	Use:   "gh",
	Short: "GitHub CLI wrapper",
	Long: `GitHub CLI wrapper with remote server support.

Run gh commands locally or on remote servers.

Examples:
  anime gh login                    # Login locally
  anime gh login lambda             # Login on lambda server
  anime gh status                   # Check local auth status
  anime gh status lambda            # Check auth status on lambda`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var ghLoginCmd = &cobra.Command{
	Use:   "login [server]",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using gh auth login.

Examples:
  anime gh login                    # Login locally
  anime gh login lambda             # Login on lambda server (interactive)
  anime gh login lambda --web       # Login on lambda using web flow`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGhLogin,
}

var ghStatusCmd = &cobra.Command{
	Use:   "status [server]",
	Short: "Check GitHub authentication status",
	Long: `Check GitHub authentication status.

Examples:
  anime gh status                   # Check local status
  anime gh status lambda            # Check status on lambda server`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGhStatus,
}

var ghRunCmd = &cobra.Command{
	Use:   "run [server] -- <gh command>",
	Short: "Run arbitrary gh command",
	Long: `Run any gh command locally or on a remote server.

Examples:
  anime gh run -- repo list         # List repos locally
  anime gh run lambda -- repo list  # List repos on lambda
  anime gh run -- pr list           # List PRs locally`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runGhRun,
}

func init() {
	rootCmd.AddCommand(ghCmd)
	ghCmd.AddCommand(ghLoginCmd)
	ghCmd.AddCommand(ghStatusCmd)
	ghCmd.AddCommand(ghRunCmd)
}

func runGhLogin(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("GITHUB LOGIN"))
	fmt.Println()

	// Check if running on remote server
	if len(args) > 0 {
		server := args[0]
		return runGhOnServer(server, []string{"auth", "login"}, true)
	}

	// Local fluid login: install gh if missing → web auth → ensure SSH
	// key → upload key → verify. After this, `git clone git@github.com:...`
	// just works for any repo the account has access to.
	step := func(n int, label string) {
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("[%d/5] %s", n, label)))
	}

	// 1. Install gh CLI if missing (apt on Linux, brew on macOS).
	step(1, "Ensure gh CLI is installed")
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Println(theme.DimTextStyle.Render("  gh not on PATH — installing..."))
		if err := installGhCLI(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  ✗ auto-install failed: " + err.Error()))
			return showGhInstallInstructions()
		}
	}
	fmt.Println(theme.SuccessStyle.Render("  ✓ gh installed"))
	fmt.Println()

	// 2. Web auth flow (works on headless cloud boxes — gh prints the URL +
	//    one-time code; user pastes them into their laptop browser).
	step(2, "Authenticate with GitHub (web flow)")
	if exec.Command("gh", "auth", "status").Run() == nil {
		who, _ := exec.Command("gh", "api", "user", "--jq", ".login").Output()
		fmt.Println(theme.SuccessStyle.Render("  ✓ already authenticated as " + strings.TrimSpace(string(who))))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  Opening device-code flow — copy the code into the URL gh prints."))
		fmt.Println()
		login := exec.Command("gh", "auth", "login",
			"--hostname", "github.com",
			"--git-protocol", "ssh",
			"--web",
			"--scopes", "admin:public_key,repo,read:org")
		login.Stdin, login.Stdout, login.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := login.Run(); err != nil {
			return fmt.Errorf("gh auth login failed: %w", err)
		}
	}
	fmt.Println()

	// 3. Make sure an SSH key exists at ~/.ssh/id_ed25519 (generate if not).
	step(3, "Ensure SSH key exists")
	home, _ := os.UserHomeDir()
	keyPath := filepath.Join(home, ".ssh", "id_ed25519")
	pubPath := keyPath + ".pub"
	if _, err := os.Stat(pubPath); err != nil {
		if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0o700); err != nil {
			return fmt.Errorf("create ~/.ssh: %w", err)
		}
		hostname, _ := os.Hostname()
		fmt.Println(theme.DimTextStyle.Render("  Generating new ed25519 key at " + keyPath + " ..."))
		gen := exec.Command("ssh-keygen", "-t", "ed25519",
			"-f", keyPath,
			"-N", "", // empty passphrase so headless usage works
			"-C", "anime-cli@"+hostname)
		gen.Stdout, gen.Stderr = os.Stdout, os.Stderr
		if err := gen.Run(); err != nil {
			return fmt.Errorf("ssh-keygen failed: %w", err)
		}
	}
	fmt.Println(theme.SuccessStyle.Render("  ✓ key at " + pubPath))
	fmt.Println()

	// 4. Upload the public key to GitHub so SSH clones work everywhere.
	//    `gh ssh-key add` fails if the same key already exists; treat that
	//    as success because it means the work's done.
	step(4, "Upload public key to GitHub")
	hostnameOut, _ := os.Hostname()
	keyTitle := "anime-cli (" + hostnameOut + ")"
	addCmd := exec.Command("gh", "ssh-key", "add", pubPath, "--title", keyTitle)
	addOut, addErr := addCmd.CombinedOutput()
	switch {
	case addErr == nil:
		fmt.Println(theme.SuccessStyle.Render("  ✓ key uploaded as: " + keyTitle))
	case strings.Contains(string(addOut), "key is already in use") ||
		strings.Contains(string(addOut), "already added"):
		fmt.Println(theme.SuccessStyle.Render("  ✓ key already on github.com (skipping)"))
	default:
		fmt.Println(theme.WarningStyle.Render("  ⚠  could not upload key: " + strings.TrimSpace(string(addOut))))
		fmt.Println(theme.DimTextStyle.Render("     This is non-fatal — the gh CLI still works for clones via HTTPS."))
	}
	fmt.Println()

	// 5. Verify SSH access.
	step(5, "Verify SSH access to github.com")
	sshTest := exec.Command("ssh", "-T", "-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking=accept-new",
		"git@github.com")
	out, _ := sshTest.CombinedOutput()
	if strings.Contains(string(out), "successfully authenticated") {
		fmt.Println(theme.SuccessStyle.Render("  ✓ " + strings.TrimSpace(string(out))))
	} else {
		fmt.Println(theme.WarningStyle.Render("  ⚠  ssh -T git@github.com did not authenticate yet"))
		fmt.Println(theme.DimTextStyle.Render("     (key propagation can take a few seconds — try again or use HTTPS via gh)"))
	}
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("✓ GitHub login complete"))
	fmt.Println(theme.DimTextStyle.Render("  You can now run: anime wan studio --yes"))
	fmt.Println()
	return nil
}

// installGhCLI installs the GitHub CLI from the official APT repo on Ubuntu/
// Debian, or via Homebrew on macOS. Idempotent — bails immediately if `gh`
// is already on PATH.
func installGhCLI() error {
	if _, err := exec.LookPath("gh"); err == nil {
		return nil
	}
	// macOS: Homebrew is the canonical path.
	if _, err := exec.LookPath("brew"); err == nil {
		fmt.Println(theme.DimTextStyle.Render("  Detected Homebrew — running: brew install gh"))
		c := exec.Command("brew", "install", "gh")
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		return c.Run()
	}
	// Linux: APT via the official keyring.
	if _, err := exec.LookPath("apt-get"); err != nil {
		return fmt.Errorf("no supported package manager (need brew or apt-get)")
	}
	script := `set -e
SUDO=""
[ "$(id -u)" -ne 0 ] && SUDO=sudo
$SUDO mkdir -p /etc/apt/keyrings
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | $SUDO dd of=/etc/apt/keyrings/githubcli-archive-keyring.gpg status=none
$SUDO chmod a+r /etc/apt/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | $SUDO tee /etc/apt/sources.list.d/github-cli.list >/dev/null
$SUDO DEBIAN_FRONTEND=noninteractive apt-get update -y
$SUDO DEBIAN_FRONTEND=noninteractive apt-get install -y gh
`
	c := exec.Command("bash", "-c", script)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

func runGhStatus(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("GITHUB STATUS"))
	fmt.Println()

	// Check if running on remote server
	if len(args) > 0 {
		server := args[0]
		return runGhOnServer(server, []string{"auth", "status"}, false)
	}

	// Local status
	if _, err := exec.LookPath("gh"); err != nil {
		return showGhInstallInstructions()
	}

	ghCmd := exec.Command("gh", "auth", "status")
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr

	if err := ghCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("Not authenticated with GitHub"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Run: ") + theme.HighlightStyle.Render("anime gh login"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	return nil
}

func runGhRun(cmd *cobra.Command, args []string) error {
	// Parse args - look for -- separator or check if first arg is a server
	var server string
	var ghArgs []string

	// Find -- separator
	dashIdx := -1
	for i, arg := range args {
		if arg == "--" {
			dashIdx = i
			break
		}
	}

	if dashIdx >= 0 {
		// Has -- separator
		if dashIdx > 0 {
			// First arg before -- might be server
			potentialServer := args[0]
			if isKnownServer(potentialServer) {
				server = potentialServer
			} else {
				// Not a server, include in gh args
				ghArgs = append(ghArgs, args[:dashIdx]...)
			}
		}
		ghArgs = append(ghArgs, args[dashIdx+1:]...)
	} else {
		// No -- separator, check if first arg is a server
		if len(args) > 0 && isKnownServer(args[0]) {
			server = args[0]
			ghArgs = args[1:]
		} else {
			ghArgs = args
		}
	}

	if len(ghArgs) == 0 {
		return fmt.Errorf("no gh command specified")
	}

	if server != "" {
		return runGhOnServer(server, ghArgs, false)
	}

	// Local execution
	if _, err := exec.LookPath("gh"); err != nil {
		return showGhInstallInstructions()
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stdin = os.Stdin
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr

	return ghCmd.Run()
}

func runGhOnServer(server string, ghArgs []string, interactive bool) error {
	target, err := resolveGhServerTarget(server)
	if err != nil {
		return err
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(target))
	fmt.Println()

	// Build the remote command
	ghCommand := "gh " + strings.Join(ghArgs, " ")

	// For login, we need special handling
	if len(ghArgs) > 0 && ghArgs[0] == "auth" && len(ghArgs) > 1 && ghArgs[1] == "login" {
		// Check if gh is installed on remote
		checkCmd := exec.Command("ssh", target, "which gh")
		if err := checkCmd.Run(); err != nil {
			fmt.Println(theme.WarningStyle.Render("gh CLI not installed on server"))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("Install with:"))
			fmt.Println(theme.HighlightStyle.Render("  ssh " + target))
			fmt.Println(theme.HighlightStyle.Render("  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg"))
			fmt.Println(theme.HighlightStyle.Render("  echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null"))
			fmt.Println(theme.HighlightStyle.Render("  sudo apt update && sudo apt install gh"))
			fmt.Println()
			return nil
		}
	}

	var sshCmd *exec.Cmd
	if interactive {
		sshCmd = exec.Command("ssh", "-t", target, ghCommand)
	} else {
		sshCmd = exec.Command("ssh", target, ghCommand)
	}

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		// Don't return error for status check failures
		if len(ghArgs) > 0 && ghArgs[0] == "auth" && len(ghArgs) > 1 && ghArgs[1] == "status" {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("Not authenticated on " + server))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("Run: ") + theme.HighlightStyle.Render("anime gh login "+server))
			fmt.Println()
			return nil
		}
		return err
	}

	return nil
}

func resolveGhServerTarget(server string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Check aliases
	if target := cfg.GetAlias(server); target != "" {
		return target, nil
	}

	// Check if it looks like user@host or has dots (IP/hostname)
	if strings.Contains(server, "@") {
		return server, nil
	}
	if strings.Contains(server, ".") {
		return "ubuntu@" + server, nil
	}

	// Try SSH config
	sshCmd := exec.Command("ssh", "-G", server)
	output, err := sshCmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		var hostname, user string
		for _, line := range lines {
			if strings.HasPrefix(line, "hostname ") {
				hostname = strings.TrimPrefix(line, "hostname ")
			}
			if strings.HasPrefix(line, "user ") {
				user = strings.TrimPrefix(line, "user ")
			}
		}
		if hostname != "" {
			if user != "" {
				return user + "@" + hostname, nil
			}
			return "ubuntu@" + hostname, nil
		}
	}

	return "", fmt.Errorf("unknown server: %s (use 'anime set %s <ip>' to configure)", server, server)
}

func isKnownServer(name string) bool {
	// Check if it's in config aliases
	cfg, err := config.Load()
	if err == nil {
		if cfg.GetAlias(name) != "" {
			return true
		}
	}

	// Check SSH config
	sshCmd := exec.Command("ssh", "-G", name)
	if output, err := sshCmd.Output(); err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "hostname ") {
				hostname := strings.TrimPrefix(line, "hostname ")
				// If hostname differs from input, it's a valid SSH alias
				if hostname != name {
					return true
				}
			}
		}
	}

	return false
}

func showGhInstallInstructions() error {
	fmt.Println(theme.ErrorStyle.Render("GitHub CLI (gh) is not installed"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Install instructions:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  macOS:"))
	fmt.Println(theme.HighlightStyle.Render("    brew install gh"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Ubuntu/Debian:"))
	fmt.Println(theme.HighlightStyle.Render("    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg"))
	fmt.Println(theme.HighlightStyle.Render("    echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null"))
	fmt.Println(theme.HighlightStyle.Render("    sudo apt update && sudo apt install gh"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  More info: ") + theme.HighlightStyle.Render("https://cli.github.com/"))
	fmt.Println()

	return fmt.Errorf("gh not installed")
}
