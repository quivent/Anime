package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

	// Local login
	fmt.Println(theme.InfoStyle.Render("Authenticating with GitHub..."))
	fmt.Println()

	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return showGhInstallInstructions()
	}

	ghCmd := exec.Command("gh", "auth", "login")
	ghCmd.Stdin = os.Stdin
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr

	if err := ghCmd.Run(); err != nil {
		return fmt.Errorf("gh auth login failed: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ GitHub authentication complete"))
	fmt.Println()

	return nil
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
