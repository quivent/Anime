package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var (
	userCreateServer string
	userSudoServer   string
	userShellServer  string
	userSudoRevoke   bool
)

var userCreateCmd = &cobra.Command{
	Use:   "create <username>",
	Short: "Interactive wizard to create a system user",
	Long: `Create a new system user with an interactive wizard.

The wizard walks through:
  1. Username
  2. Shell (bash, zsh, fish, sh)
  3. Sudo access (yes/no)
  4. SSH key setup (generate or paste)
  5. Groups (docker, www-data, etc.)

Can run locally or on a remote server.

Examples:
  anime user create deploy
  anime user create app --server wings
  anime user create dev -s lambda`,
	Args: cobra.ExactArgs(1),
	RunE: runUserCreate,
}

var userSudoCmd = &cobra.Command{
	Use:   "sudo <username>",
	Short: "Grant or revoke sudo access for a user",
	Long: `Grant sudo privileges to a user, or revoke with --revoke.

Examples:
  anime user sudo deploy                # Grant sudo
  anime user sudo deploy --revoke       # Revoke sudo
  anime user sudo deploy -s wings       # On remote server`,
	Args: cobra.ExactArgs(1),
	RunE: runUserSudo,
}

var userShellCmd = &cobra.Command{
	Use:   "shell <username> [shell]",
	Short: "Change a user's login shell",
	Long: `Change the login shell for a user.

If no shell is specified, prompts interactively.

Examples:
  anime user shell deploy bash
  anime user shell deploy zsh
  anime user shell deploy fish
  anime user shell app --server wings`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUserShell,
}

func init() {
	userCreateCmd.Flags().StringVarP(&userCreateServer, "server", "s", "", "Run on remote server")
	userSudoCmd.Flags().StringVarP(&userSudoServer, "server", "s", "", "Run on remote server")
	userSudoCmd.Flags().BoolVar(&userSudoRevoke, "revoke", false, "Revoke sudo access instead of granting")
	userShellCmd.Flags().StringVarP(&userShellServer, "server", "s", "", "Run on remote server")

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userSudoCmd)
	userCmd.AddCommand(userShellCmd)
}

// runSystemCmd executes a command locally or on a remote server
func runSystemCmd(serverName, script string) (string, error) {
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

func runUserCreate(cmd *cobra.Command, args []string) error {
	username := args[0]
	if err := validate.Username(username); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)

	target := "locally"
	if userCreateServer != "" {
		target = "on " + userCreateServer
	}

	fmt.Println(theme.RenderBanner("👤 CREATE USER 👤"))
	fmt.Println()
	fmt.Printf("  Creating user %s %s\n\n",
		theme.HighlightStyle.Render(username),
		theme.DimTextStyle.Render(target))

	// Step 1: Shell
	fmt.Println(theme.InfoStyle.Render("  1. Login shell"))
	fmt.Println()
	fmt.Printf("    %s bash  %s\n", theme.HighlightStyle.Render("1"), theme.DimTextStyle.Render("(default)"))
	fmt.Printf("    %s zsh\n", theme.HighlightStyle.Render("2"))
	fmt.Printf("    %s fish\n", theme.HighlightStyle.Render("3"))
	fmt.Printf("    %s sh\n", theme.HighlightStyle.Render("4"))
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	shellChoice, _ := reader.ReadString('\n')
	shellChoice = strings.TrimSpace(shellChoice)

	shellPath := "/bin/bash"
	shellName := "bash"
	switch shellChoice {
	case "2":
		shellPath = "/bin/zsh"
		shellName = "zsh"
	case "3":
		shellPath = "/usr/bin/fish"
		shellName = "fish"
	case "4":
		shellPath = "/bin/sh"
		shellName = "sh"
	}
	fmt.Printf("  %s Shell: %s\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess), shellName)

	// Step 2: Sudo
	fmt.Println(theme.InfoStyle.Render("  2. Sudo access"))
	fmt.Print("  Grant sudo? [Y/n]: ")
	sudoChoice, _ := reader.ReadString('\n')
	sudoChoice = strings.TrimSpace(strings.ToLower(sudoChoice))
	grantSudo := sudoChoice != "n" && sudoChoice != "no"
	if grantSudo {
		fmt.Printf("  %s Sudo: yes\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	} else {
		fmt.Printf("  %s Sudo: no\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}

	// Step 3: SSH key
	fmt.Println(theme.InfoStyle.Render("  3. SSH key setup"))
	fmt.Println()
	fmt.Printf("    %s Generate new ed25519 key  %s\n", theme.HighlightStyle.Render("1"), theme.DimTextStyle.Render("(default)"))
	fmt.Printf("    %s Paste an existing public key\n", theme.HighlightStyle.Render("2"))
	fmt.Printf("    %s Skip SSH key setup\n", theme.HighlightStyle.Render("3"))
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	sshChoice, _ := reader.ReadString('\n')
	sshChoice = strings.TrimSpace(sshChoice)

	var sshKeyAction string // "generate", "paste", "skip"
	var pastedKey string
	switch sshChoice {
	case "2":
		sshKeyAction = "paste"
		fmt.Print("  Paste public key: ")
		pastedKey, _ = reader.ReadString('\n')
		pastedKey = strings.TrimSpace(pastedKey)
		fmt.Printf("  %s SSH key will be added\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	case "3":
		sshKeyAction = "skip"
		fmt.Printf("  %s Skipping SSH key\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	default:
		sshKeyAction = "generate"
		fmt.Printf("  %s Will generate ed25519 key\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}

	// Step 4: Extra groups
	fmt.Println(theme.InfoStyle.Render("  4. Extra groups (comma-separated, empty for none)"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Common: docker, www-data, adm"))
	fmt.Print("  Groups []: ")
	groupsInput, _ := reader.ReadString('\n')
	groupsInput = strings.TrimSpace(groupsInput)
	var groups []string
	if groupsInput != "" {
		for _, g := range strings.Split(groupsInput, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				if err := validate.GroupName(g); err != nil {
					return err
				}
				groups = append(groups, g)
			}
		}
	}
	if len(groups) > 0 {
		fmt.Printf("  %s Groups: %s\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.Join(groups, ", "))
	} else {
		fmt.Printf("  %s No extra groups\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}

	// Confirmation
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("  Summary"))
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("    User:   %s\n", theme.HighlightStyle.Render(username))
	fmt.Printf("    Shell:  %s\n", shellName)
	fmt.Printf("    Sudo:   %v\n", grantSudo)
	fmt.Printf("    SSH:    %s\n", sshKeyAction)
	if len(groups) > 0 {
		fmt.Printf("    Groups: %s\n", strings.Join(groups, ", "))
	}
	fmt.Printf("    Target: %s\n", target)
	fmt.Println()
	fmt.Print("  Create user? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm == "n" || confirm == "no" {
		fmt.Println("  Cancelled")
		return nil
	}

	// Build and execute the script
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Creating user..."))

	script := fmt.Sprintf(`#!/bin/bash
set -e

# Create user
if id "%s" &>/dev/null; then
    echo "User %s already exists"
else
    sudo useradd -m -s %s "%s"
    echo "Created user %s with shell %s"
fi
`, username, username, shellPath, username, username, shellPath)

	// Sudo
	if grantSudo {
		script += fmt.Sprintf(`
# Grant sudo
sudo usermod -aG sudo "%s" 2>/dev/null || sudo usermod -aG wheel "%s" 2>/dev/null || true
echo "Granted sudo to %s"
`, username, username, username)
	}

	// Groups
	for _, g := range groups {
		script += fmt.Sprintf(`sudo usermod -aG "%s" "%s" 2>/dev/null || echo "Group %s not found, skipping"
`, g, username, g)
	}

	// SSH key
	switch sshKeyAction {
	case "generate":
		script += fmt.Sprintf(`
# Generate SSH key
sudo mkdir -p /home/%s/.ssh
sudo ssh-keygen -t ed25519 -f /home/%s/.ssh/id_ed25519 -N "" -C "%s@$(hostname)" <<<y 2>/dev/null || true
sudo chown -R %s:%s /home/%s/.ssh
sudo chmod 700 /home/%s/.ssh
sudo chmod 600 /home/%s/.ssh/id_ed25519
sudo chmod 644 /home/%s/.ssh/id_ed25519.pub
echo "SSH key generated:"
sudo cat /home/%s/.ssh/id_ed25519.pub
`, username, username, username, username, username, username, username, username, username, username)
	case "paste":
		script += fmt.Sprintf(`
# Add SSH public key
sudo mkdir -p /home/%s/.ssh
echo "%s" | sudo tee /home/%s/.ssh/authorized_keys > /dev/null
sudo chown -R %s:%s /home/%s/.ssh
sudo chmod 700 /home/%s/.ssh
sudo chmod 600 /home/%s/.ssh/authorized_keys
echo "SSH public key added to authorized_keys"
`, username, pastedKey, username, username, username, username, username, username)
	}

	output, err := runSystemCmd(userCreateServer, script)
	if output != "" {
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			if line != "" {
				fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
			}
		}
	}
	if err != nil {
		fmt.Printf("  %s %s\n", theme.ErrorStyle.Render(theme.SymbolError+" Failed:"), theme.DimTextStyle.Render(err.Error()))
		return err
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ User %s created!", username)))
	fmt.Println()
	return nil
}

func runUserSudo(cmd *cobra.Command, args []string) error {
	username := args[0]
	if err := validate.Username(username); err != nil {
		return err
	}

	var script string
	var action string
	if userSudoRevoke {
		action = "Revoking sudo from"
		script = fmt.Sprintf(`sudo deluser "%s" sudo 2>/dev/null || sudo gpasswd -d "%s" wheel 2>/dev/null || true
echo "Revoked sudo from %s"`, username, username, username)
	} else {
		action = "Granting sudo to"
		script = fmt.Sprintf(`sudo usermod -aG sudo "%s" 2>/dev/null || sudo usermod -aG wheel "%s" 2>/dev/null || true
echo "Granted sudo to %s"`, username, username, username)
	}

	fmt.Printf("  %s %s %s...\n",
		theme.SymbolLoading, action,
		theme.HighlightStyle.Render(username))

	output, err := runSystemCmd(userSudoServer, script)
	if output != "" {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(output)))
	}
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	fmt.Printf("  %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess+" Done"))
	return nil
}

func runUserShell(cmd *cobra.Command, args []string) error {
	username := args[0]
	if err := validate.Username(username); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)

	var shellPath string
	if len(args) > 1 {
		shellPath = resolveShellPath(args[1])
	} else {
		// Interactive
		fmt.Println()
		fmt.Printf("  Change shell for %s\n\n", theme.HighlightStyle.Render(username))
		fmt.Printf("    %s bash\n", theme.HighlightStyle.Render("1"))
		fmt.Printf("    %s zsh\n", theme.HighlightStyle.Render("2"))
		fmt.Printf("    %s fish\n", theme.HighlightStyle.Render("3"))
		fmt.Printf("    %s sh\n", theme.HighlightStyle.Render("4"))
		fmt.Println()
		fmt.Print("  Choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1", "bash":
			shellPath = "/bin/bash"
		case "2", "zsh":
			shellPath = "/bin/zsh"
		case "3", "fish":
			shellPath = "/usr/bin/fish"
		case "4", "sh":
			shellPath = "/bin/sh"
		default:
			return fmt.Errorf("invalid choice: %s", choice)
		}
	}

	fmt.Printf("  %s Changing shell for %s to %s...\n",
		theme.SymbolLoading,
		theme.HighlightStyle.Render(username),
		theme.DimTextStyle.Render(shellPath))

	script := fmt.Sprintf(`sudo chsh -s "%s" "%s" && echo "Shell changed to %s for %s"`,
		shellPath, username, shellPath, username)

	output, err := runSystemCmd(userShellServer, script)
	if output != "" {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(output)))
	}
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	fmt.Printf("  %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess+" Done"))
	return nil
}

func resolveShellPath(name string) string {
	switch strings.ToLower(name) {
	case "bash":
		return "/bin/bash"
	case "zsh":
		return "/bin/zsh"
	case "fish":
		return "/usr/bin/fish"
	case "sh":
		return "/bin/sh"
	default:
		return name // assume it's a full path
	}
}
