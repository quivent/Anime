package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// ============================================================================
// USER MANAGEMENT COMMANDS
// Manages Linux users on remote servers + local user profiles
// ============================================================================

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users on remote servers",
	Long: `Create, list, update, and remove Linux users on remote servers.

Commands:
  add      Interactive wizard to create a new user on a server
  list     List all users on a server
  remove   Remove a user from a server
  update   Update user settings (password, SSH keys, sudo)

Local Profile Commands:
  profile  Manage local user profiles for quick navigation

Examples:
  anime user add lambda              # Start wizard for 'lambda' server
  anime user list lambda             # List users on lambda
  anime user remove lambda john      # Remove user 'john' from lambda
  anime user update lambda john      # Update settings for 'john'`,
	Run: runUserHelp,
}

// ============================================================================
// SERVER USER COMMANDS
// ============================================================================

var userAddCmd = &cobra.Command{
	Use:   "add [server]",
	Short: "Add a new user to a server (interactive wizard)",
	Long: `Interactive wizard to create a new Linux user on a remote server.

The wizard will guide you through:
  - Username
  - Password (or auto-generate)
  - SSH public keys
  - Sudo access
  - Shell preference
  - Home directory creation

Examples:
  anime user add              # Use default server
  anime user add lambda       # Create user on lambda server`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserAdd,
}

var userListCmd = &cobra.Command{
	Use:     "list [server]",
	Aliases: []string{"ls"},
	Short:   "List users on a server",
	Long: `List all users on a remote server.

Shows system users (UID >= 1000) with their details:
  - Username
  - UID/GID
  - Home directory
  - Shell
  - Sudo access

Examples:
  anime user list              # Use default server
  anime user list lambda       # List users on lambda`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserList,
}

var userRemoveCmd = &cobra.Command{
	Use:   "remove [server] <username>",
	Short: "Remove a user from a server",
	Long: `Remove a Linux user from a remote server.

By default, the user's home directory is preserved.
Use --delete-home to remove the home directory.

Examples:
  anime user remove lambda john              # Remove john, keep home
  anime user remove lambda john --delete-home # Remove john and home dir
  anime user remove john                     # Remove from default server`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUserRemove,
}

var userUpdateCmd = &cobra.Command{
	Use:   "update [server] <username>",
	Short: "Update user settings",
	Long: `Update settings for an existing user on a remote server.

You can update:
  - Password
  - SSH public keys
  - Sudo access
  - Shell

Examples:
  anime user update lambda john              # Interactive update
  anime user update lambda john --password   # Change password only
  anime user update lambda john --sudo       # Grant sudo access
  anime user update lambda john --no-sudo    # Revoke sudo access`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUserUpdate,
}

// ============================================================================
// LOCAL PROFILE COMMANDS (for directory navigation)
// ============================================================================

var userProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage local user profiles for quick navigation",
	Long:  "Create, list, and manage local user profiles for quick directory navigation",
	Run:   runUserProfileHelp,
}

var userProfileAddCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Add a local user profile",
	Long: `Add a new local user profile with their home directory.

The user's home directory will be auto-detected from the system or can be specified.

Examples:
  anime user profile add josh
  anime user profile add maria --path /custom/path`,
	Args: cobra.ExactArgs(1),
	RunE: runUserProfileAdd,
}

var userProfileRemoveCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove a local user profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserProfileRemove,
}

var userProfileListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List local user profiles",
	Run:     runUserProfileList,
}

var userProfileSetCmd = &cobra.Command{
	Use:   "set <username>",
	Short: "Set the active local user profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserProfileSet,
}

var userProfileCdCmd = &cobra.Command{
	Use:   "cd [username]",
	Short: "Print cd command for user directory",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUserProfileCd,
}

// Alias command for quick listing
var usersCmd = &cobra.Command{
	Use:   "users [server]",
	Short: "List users on a server (alias for 'user list')",
	Long:  "List all users on a remote server.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUserList,
}

var (
	deleteHome     bool
	updatePassword bool
	updateSudo     bool
	updateNoSudo   bool
	profilePath    string
	confirmRemove  bool
)

func init() {
	// Server user flags
	userRemoveCmd.Flags().BoolVar(&deleteHome, "delete-home", false, "Delete user's home directory")
	userUpdateCmd.Flags().BoolVar(&updatePassword, "password", false, "Update password only")
	userUpdateCmd.Flags().BoolVar(&updateSudo, "sudo", false, "Grant sudo access")
	userUpdateCmd.Flags().BoolVar(&updateNoSudo, "no-sudo", false, "Revoke sudo access")

	// Local profile flags
	userProfileAddCmd.Flags().StringVarP(&profilePath, "path", "p", "", "Custom path for user profile")
	userProfileRemoveCmd.Flags().BoolVar(&confirmRemove, "confirm", false, "Confirm removal")

	// Add server user subcommands
	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userRemoveCmd)
	userCmd.AddCommand(userUpdateCmd)

	// Add local profile subcommands
	userProfileCmd.AddCommand(userProfileAddCmd)
	userProfileCmd.AddCommand(userProfileRemoveCmd)
	userProfileCmd.AddCommand(userProfileListCmd)
	userProfileCmd.AddCommand(userProfileSetCmd)
	userProfileCmd.AddCommand(userProfileCdCmd)
	userCmd.AddCommand(userProfileCmd)

	// Add to root
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(usersCmd)
}

func runUserHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("👤 USER MANAGEMENT 👤"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Manage Linux users on remote servers"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🖥️  Server User Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	serverCommands := []struct {
		cmd  string
		desc string
	}{
		{"anime user add [server]", "Interactive wizard to create a new user"},
		{"anime user list [server]", "List all users on a server"},
		{"anime user remove [server] <user>", "Remove a user from a server"},
		{"anime user update [server] <user>", "Update user settings"},
		{"anime users [server]", "List users (shortcut)"},
	}

	for _, c := range serverCommands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📁 Local Profile Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	profileCommands := []struct {
		cmd  string
		desc string
	}{
		{"anime user profile add <name>", "Add a local user profile"},
		{"anime user profile list", "List local profiles"},
		{"anime user profile set <name>", "Set active profile"},
		{"$(anime user profile cd)", "Navigate to profile directory"},
	}

	for _, c := range profileCommands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}
}

// ============================================================================
// SERVER USER: ADD WIZARD
// ============================================================================

type serverUserConfig struct {
	username   string
	fullName   string
	password   string
	sshKeys    []string
	sudo       bool
	shell      string
	createHome bool
}

func runUserAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Determine target server
	serverName := "lambda" // default
	if len(args) > 0 {
		serverName = args[0]
	}

	// Load config and get server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server not found: %s\nUse 'anime add' to add a server first", serverName)
	}

	// Welcome banner
	fmt.Println()
	fmt.Println(theme.RenderBanner("👤 ADD USER WIZARD 👤"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🖥️  Target server: %s (%s@%s)", serverName, server.User, server.Host)))
	fmt.Println()

	userCfg := &serverUserConfig{
		createHome: true,
		shell:      "/bin/bash",
	}

	// Step 1: Username
	printUserStep(1, "Username")
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Enter the username for the new Linux user."))
	fmt.Println(theme.DimTextStyle.Render("  Must start with a letter, lowercase, no spaces."))
	fmt.Println()

	for {
		fmt.Print(theme.HighlightStyle.Render("Username ▶ "))
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		if username == "" {
			fmt.Println(theme.ErrorStyle.Render("  ✗ Username cannot be empty"))
			continue
		}

		if !isValidUsername(username) {
			fmt.Println(theme.ErrorStyle.Render("  ✗ Invalid username. Use lowercase letters, numbers, and underscores."))
			continue
		}

		userCfg.username = username
		break
	}
	fmt.Println()

	// Step 2: Full Name (optional)
	printUserStep(2, "Full Name (optional)")
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Enter the user's full name (for GECOS field)."))
	fmt.Println(theme.DimTextStyle.Render("  Press Enter to skip."))
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Full name ▶ "))
	fullName, _ := reader.ReadString('\n')
	userCfg.fullName = strings.TrimSpace(fullName)
	fmt.Println()

	// Step 3: Password
	printUserStep(3, "Password")
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Generate random password"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Enter password manually"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("No password (SSH key only)"))
	fmt.Println()

	choice := promptUserChoice(reader, "Choose option", []string{"1", "2", "3"})
	switch choice {
	case "1":
		userCfg.password = generateRandomPassword(16)
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("  ✓ Generated password: ") + theme.HighlightStyle.Render(userCfg.password))
		fmt.Println(theme.WarningStyle.Render("  ⚠️  Save this password! It won't be shown again."))
	case "2":
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("Password ▶ "))
		password, _ := reader.ReadString('\n')
		userCfg.password = strings.TrimSpace(password)
	case "3":
		userCfg.password = ""
		fmt.Println(theme.InfoStyle.Render("  → User will authenticate via SSH key only"))
	}
	fmt.Println()

	// Step 4: SSH Keys
	printUserStep(4, "SSH Public Keys")
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Add SSH public keys for passwordless authentication."))
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Paste public key(s)"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Read from file"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Use my local public key (~/.ssh/id_*.pub)"))
	fmt.Println(theme.HighlightStyle.Render("  4") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Skip (no SSH keys)"))
	fmt.Println()

	choice = promptUserChoice(reader, "Choose option", []string{"1", "2", "3", "4"})
	switch choice {
	case "1":
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Paste public key(s), one per line. Enter empty line when done:"))
		for {
			fmt.Print(theme.HighlightStyle.Render("Key ▶ "))
			key, _ := reader.ReadString('\n')
			key = strings.TrimSpace(key)
			if key == "" {
				break
			}
			if strings.HasPrefix(key, "ssh-") || strings.HasPrefix(key, "ecdsa-") {
				userCfg.sshKeys = append(userCfg.sshKeys, key)
				fmt.Println(theme.SuccessStyle.Render("  ✓ Key added"))
			} else {
				fmt.Println(theme.ErrorStyle.Render("  ✗ Invalid key format. Should start with ssh-rsa, ssh-ed25519, etc."))
			}
		}
	case "2":
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("Path to public key file ▶ "))
		keyPath, _ := reader.ReadString('\n')
		keyPath = strings.TrimSpace(keyPath)
		keyPath = expandPath(keyPath)

		if content, err := os.ReadFile(keyPath); err == nil {
			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "ssh-") || strings.HasPrefix(line, "ecdsa-") {
					userCfg.sshKeys = append(userCfg.sshKeys, line)
				}
			}
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Loaded %d key(s) from file", len(userCfg.sshKeys))))
		} else {
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  ✗ Failed to read file: %s", err)))
		}
	case "3":
		// Try to find local public keys
		home, _ := os.UserHomeDir()
		keyFiles := []string{
			filepath.Join(home, ".ssh", "id_ed25519.pub"),
			filepath.Join(home, ".ssh", "id_rsa.pub"),
			filepath.Join(home, ".ssh", "id_ecdsa.pub"),
		}
		for _, kf := range keyFiles {
			if content, err := os.ReadFile(kf); err == nil {
				key := strings.TrimSpace(string(content))
				if strings.HasPrefix(key, "ssh-") || strings.HasPrefix(key, "ecdsa-") {
					userCfg.sshKeys = append(userCfg.sshKeys, key)
					fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Found: %s", kf)))
				}
			}
		}
		if len(userCfg.sshKeys) == 0 {
			fmt.Println(theme.WarningStyle.Render("  ⚠️  No local public keys found"))
		}
	case "4":
		fmt.Println(theme.InfoStyle.Render("  → No SSH keys will be added"))
	}
	fmt.Println()

	// Step 5: Sudo Access
	printUserStep(5, "Sudo Access")
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Grant the user sudo (administrator) privileges?"))
	fmt.Println()

	userCfg.sudo = promptUserYesNo(reader, "Grant sudo access", false)
	if userCfg.sudo {
		fmt.Println(theme.SuccessStyle.Render("  ✓ User will have sudo privileges"))
	} else {
		fmt.Println(theme.InfoStyle.Render("  → User will be a regular user"))
	}
	fmt.Println()

	// Step 6: Shell
	printUserStep(6, "Shell Preference")
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/bash") + theme.DimTextStyle.Render(" (default)"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/zsh"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/sh"))
	fmt.Println()

	choice = promptUserChoice(reader, "Choose shell (or Enter for bash)", []string{"1", "2", "3", ""})
	switch choice {
	case "2":
		userCfg.shell = "/bin/zsh"
	case "3":
		userCfg.shell = "/bin/sh"
	default:
		userCfg.shell = "/bin/bash"
	}
	fmt.Println()

	// Step 7: Home Directory
	printUserStep(7, "Home Directory")
	fmt.Println()
	userCfg.createHome = promptUserYesNo(reader, "Create home directory", true)
	fmt.Println()

	// Summary
	fmt.Println()
	fmt.Println(theme.RenderBanner("📋 USER SUMMARY"))
	fmt.Println()

	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Username:"), theme.HighlightStyle.Render(userCfg.username))
	if userCfg.fullName != "" {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Full Name:"), theme.DimTextStyle.Render(userCfg.fullName))
	}
	if userCfg.password != "" {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Password:"), theme.DimTextStyle.Render("****"))
	} else {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Password:"), theme.DimTextStyle.Render("(none - SSH key only)"))
	}
	fmt.Printf("  %s %d key(s)\n", theme.InfoStyle.Render("SSH Keys:"), len(userCfg.sshKeys))
	fmt.Printf("  %s %v\n", theme.InfoStyle.Render("Sudo:"), userCfg.sudo)
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Shell:"), theme.DimTextStyle.Render(userCfg.shell))
	fmt.Printf("  %s %v\n", theme.InfoStyle.Render("Home Dir:"), userCfg.createHome)
	fmt.Println()

	// Confirm
	if !promptUserYesNo(reader, "Create user on "+serverName, true) {
		fmt.Println(theme.WarningStyle.Render("User creation cancelled"))
		return nil
	}

	// Execute
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Creating user..."))
	fmt.Println()

	return executeCreateUser(server, userCfg)
}

func executeCreateUser(server *config.Server, userCfg *serverUserConfig) error {
	// Connect to server
	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Build useradd command
	useraddArgs := []string{"sudo", "useradd"}

	if userCfg.createHome {
		useraddArgs = append(useraddArgs, "-m")
	}

	useraddArgs = append(useraddArgs, "-s", userCfg.shell)

	if userCfg.fullName != "" {
		useraddArgs = append(useraddArgs, "-c", fmt.Sprintf("'%s'", userCfg.fullName))
	}

	useraddArgs = append(useraddArgs, userCfg.username)

	// Create user
	cmd := strings.Join(useraddArgs, " ")
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("→"), cmd)

	output, err := client.RunCommand(cmd)
	if err != nil {
		// Check if user already exists
		if strings.Contains(output, "already exists") {
			fmt.Println(theme.WarningStyle.Render("  ⚠️  User already exists"))
		} else {
			return fmt.Errorf("failed to create user: %s", output)
		}
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ User created"))
	}

	// Set password if provided
	if userCfg.password != "" {
		// Use chpasswd to set password
		cmd := fmt.Sprintf("echo '%s:%s' | sudo chpasswd", userCfg.username, userCfg.password)
		fmt.Printf("  %s Setting password...\n", theme.DimTextStyle.Render("→"))

		_, err := client.RunCommand(cmd)
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("  ✗ Failed to set password"))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  ✓ Password set"))
		}
	}

	// Add SSH keys
	if len(userCfg.sshKeys) > 0 {
		fmt.Printf("  %s Adding SSH keys...\n", theme.DimTextStyle.Render("→"))

		// Determine home directory
		homeDir := fmt.Sprintf("/home/%s", userCfg.username)

		// Create .ssh directory
		mkdirCmd := fmt.Sprintf("sudo mkdir -p %s/.ssh && sudo chmod 700 %s/.ssh", homeDir, homeDir)
		client.RunCommand(mkdirCmd)

		// Write authorized_keys
		keysContent := strings.Join(userCfg.sshKeys, "\n")
		writeCmd := fmt.Sprintf("echo '%s' | sudo tee %s/.ssh/authorized_keys > /dev/null", keysContent, homeDir)
		client.RunCommand(writeCmd)

		// Set permissions
		chmodCmd := fmt.Sprintf("sudo chmod 600 %s/.ssh/authorized_keys && sudo chown -R %s:%s %s/.ssh",
			homeDir, userCfg.username, userCfg.username, homeDir)
		client.RunCommand(chmodCmd)

		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Added %d SSH key(s)", len(userCfg.sshKeys))))
	}

	// Add to sudo group if requested
	if userCfg.sudo {
		fmt.Printf("  %s Granting sudo access...\n", theme.DimTextStyle.Render("→"))

		cmd := fmt.Sprintf("sudo usermod -aG sudo %s", userCfg.username)
		_, err := client.RunCommand(cmd)
		if err != nil {
			// Try wheel group (CentOS/RHEL)
			cmd = fmt.Sprintf("sudo usermod -aG wheel %s", userCfg.username)
			client.RunCommand(cmd)
		}
		fmt.Println(theme.SuccessStyle.Render("  ✓ Sudo access granted"))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ User created successfully!"))
	fmt.Println()

	if len(userCfg.sshKeys) > 0 {
		fmt.Println(theme.InfoStyle.Render("💡 Connect with:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("ssh %s@%s", userCfg.username, server.Host)))
		fmt.Println()
	}

	return nil
}

// ============================================================================
// SERVER USER: LIST
// ============================================================================

func runUserList(cmd *cobra.Command, args []string) error {
	// Determine target server
	serverName := "lambda" // default
	if len(args) > 0 {
		serverName = args[0]
	}

	// Load config and get server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server not found: %s", serverName)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("👥 USERS ON " + strings.ToUpper(serverName) + " 👥"))
	fmt.Println()

	// Connect to server
	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Get users with UID >= 1000 (regular users, not system users)
	output, err := client.RunCommand("awk -F: '$3 >= 1000 && $3 < 65534 {print $1\":\"$3\":\"$4\":\"$6\":\"$7}' /etc/passwd")
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Get sudo/wheel group members
	sudoOutput, _ := client.RunCommand("getent group sudo wheel 2>/dev/null | cut -d: -f4")
	sudoUsers := make(map[string]bool)
	for _, user := range strings.Split(sudoOutput, ",") {
		sudoUsers[strings.TrimSpace(user)] = true
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println(theme.WarningStyle.Render("  No regular users found"))
		fmt.Println()
		return nil
	}

	fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Found %d user(s)", len(lines))))
	fmt.Println()

	// Header
	fmt.Printf("  %-15s %-8s %-25s %-15s %s\n",
		theme.InfoStyle.Render("USERNAME"),
		theme.InfoStyle.Render("UID"),
		theme.InfoStyle.Render("HOME"),
		theme.InfoStyle.Render("SHELL"),
		theme.InfoStyle.Render("SUDO"))
	fmt.Println(theme.DimTextStyle.Render("  " + strings.Repeat("─", 75)))

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 5 {
			continue
		}

		username := parts[0]
		uid := parts[1]
		home := parts[3]
		shell := filepath.Base(parts[4])

		sudoStatus := theme.DimTextStyle.Render("no")
		if sudoUsers[username] {
			sudoStatus = theme.SuccessStyle.Render("yes")
		}

		fmt.Printf("  %-15s %-8s %-25s %-15s %s\n",
			theme.HighlightStyle.Render(username),
			theme.DimTextStyle.Render(uid),
			theme.DimTextStyle.Render(home),
			theme.DimTextStyle.Render(shell),
			sudoStatus)
	}

	fmt.Println()
	return nil
}

// ============================================================================
// SERVER USER: REMOVE
// ============================================================================

func runUserRemove(cmd *cobra.Command, args []string) error {
	var serverName, username string

	// Parse arguments - could be "user remove <server> <user>" or "user remove <user>"
	if len(args) == 1 {
		serverName = "lambda"
		username = args[0]
	} else {
		serverName = args[0]
		username = args[1]
	}

	// Load config and get server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server not found: %s", serverName)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🗑️  REMOVE USER 🗑️"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Server:"), theme.HighlightStyle.Render(serverName))
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("User:"), theme.HighlightStyle.Render(username))
	fmt.Printf("  %s %v\n", theme.InfoStyle.Render("Delete home:"), deleteHome)
	fmt.Println()

	// Confirm
	reader := bufio.NewReader(os.Stdin)
	if !promptUserYesNo(reader, fmt.Sprintf("Remove user '%s' from %s", username, serverName), false) {
		fmt.Println(theme.WarningStyle.Render("Operation cancelled"))
		return nil
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Removing user..."))
	fmt.Println()

	// Connect to server
	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Build userdel command
	userdelCmd := "sudo userdel"
	if deleteHome {
		userdelCmd += " -r"
	}
	userdelCmd += " " + username

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("→"), userdelCmd)

	output, err := client.RunCommand(userdelCmd)
	if err != nil {
		if strings.Contains(output, "does not exist") {
			return fmt.Errorf("user '%s' does not exist", username)
		}
		return fmt.Errorf("failed to remove user: %s", output)
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ User removed"))
	fmt.Println()

	return nil
}

// ============================================================================
// SERVER USER: UPDATE
// ============================================================================

func runUserUpdate(cmd *cobra.Command, args []string) error {
	var serverName, username string

	// Parse arguments
	if len(args) == 1 {
		serverName = "lambda"
		username = args[0]
	} else {
		serverName = args[0]
		username = args[1]
	}

	// Load config and get server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server not found: %s", serverName)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("⚙️  UPDATE USER ⚙️"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Server:"), theme.HighlightStyle.Render(serverName))
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("User:"), theme.HighlightStyle.Render(username))
	fmt.Println()

	// Connect to server
	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Check if user exists
	output, err := client.RunCommand(fmt.Sprintf("id %s", username))
	if err != nil {
		return fmt.Errorf("user '%s' does not exist", username)
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Current:"), theme.DimTextStyle.Render(strings.TrimSpace(output)))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Handle specific flags
	if updatePassword {
		return updateUserPassword(client, username, reader)
	}

	if updateSudo {
		return grantSudoAccess(client, username)
	}

	if updateNoSudo {
		return revokeSudoAccess(client, username)
	}

	// Interactive update menu
	fmt.Println(theme.InfoStyle.Render("What would you like to update?"))
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Change password"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Add SSH key"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Grant sudo access"))
	fmt.Println(theme.HighlightStyle.Render("  4") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Revoke sudo access"))
	fmt.Println(theme.HighlightStyle.Render("  5") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Change shell"))
	fmt.Println(theme.HighlightStyle.Render("  6") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Cancel"))
	fmt.Println()

	choice := promptUserChoice(reader, "Choose option", []string{"1", "2", "3", "4", "5", "6"})

	switch choice {
	case "1":
		return updateUserPassword(client, username, reader)
	case "2":
		return addUserSSHKey(client, username, reader)
	case "3":
		return grantSudoAccess(client, username)
	case "4":
		return revokeSudoAccess(client, username)
	case "5":
		return changeUserShell(client, username, reader)
	default:
		fmt.Println(theme.WarningStyle.Render("Operation cancelled"))
	}

	return nil
}

func updateUserPassword(client *ssh.Client, username string, reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Generate random password"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Enter password manually"))
	fmt.Println()

	choice := promptUserChoice(reader, "Choose option", []string{"1", "2"})

	var password string
	if choice == "1" {
		password = generateRandomPassword(16)
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("  Generated password: ") + theme.HighlightStyle.Render(password))
		fmt.Println(theme.WarningStyle.Render("  ⚠️  Save this password!"))
	} else {
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("New password ▶ "))
		pwd, _ := reader.ReadString('\n')
		password = strings.TrimSpace(pwd)
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Setting password..."))

	cmd := fmt.Sprintf("echo '%s:%s' | sudo chpasswd", username, password)
	_, err := client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to set password")
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Password updated"))
	fmt.Println()

	return nil
}

func addUserSSHKey(client *ssh.Client, username string, reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Paste the SSH public key:"))
	fmt.Print(theme.HighlightStyle.Render("Key ▶ "))
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	if !strings.HasPrefix(key, "ssh-") && !strings.HasPrefix(key, "ecdsa-") {
		return fmt.Errorf("invalid key format")
	}

	homeDir := fmt.Sprintf("/home/%s", username)
	cmd := fmt.Sprintf("echo '%s' | sudo tee -a %s/.ssh/authorized_keys > /dev/null", key, homeDir)
	_, err := client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to add SSH key")
	}

	fmt.Println(theme.SuccessStyle.Render("✓ SSH key added"))
	fmt.Println()

	return nil
}

func grantSudoAccess(client *ssh.Client, username string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Granting sudo access..."))

	cmd := fmt.Sprintf("sudo usermod -aG sudo %s", username)
	_, err := client.RunCommand(cmd)
	if err != nil {
		// Try wheel group
		cmd = fmt.Sprintf("sudo usermod -aG wheel %s", username)
		client.RunCommand(cmd)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Sudo access granted"))
	fmt.Println()

	return nil
}

func revokeSudoAccess(client *ssh.Client, username string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Revoking sudo access..."))

	cmd := fmt.Sprintf("sudo deluser %s sudo 2>/dev/null; sudo gpasswd -d %s wheel 2>/dev/null", username, username)
	client.RunCommand(cmd)

	fmt.Println(theme.SuccessStyle.Render("✓ Sudo access revoked"))
	fmt.Println()

	return nil
}

func changeUserShell(client *ssh.Client, username string, reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/bash"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/zsh"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("/bin/sh"))
	fmt.Println()

	choice := promptUserChoice(reader, "Choose shell", []string{"1", "2", "3"})

	var shell string
	switch choice {
	case "2":
		shell = "/bin/zsh"
	case "3":
		shell = "/bin/sh"
	default:
		shell = "/bin/bash"
	}

	cmd := fmt.Sprintf("sudo chsh -s %s %s", shell, username)
	_, err := client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to change shell")
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Shell changed to %s", shell)))
	fmt.Println()

	return nil
}

// ============================================================================
// LOCAL PROFILE COMMANDS
// ============================================================================

func runUserProfileHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📁 LOCAL PROFILES 📁"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Manage local user profiles for quick directory navigation"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime user profile add <name>", "Add a local user profile"},
		{"anime user profile remove <name> --confirm", "Remove a profile"},
		{"anime user profile list", "List all profiles"},
		{"anime user profile set <name>", "Set active profile"},
		{"$(anime user profile cd)", "Navigate to profile directory"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}
}

func runUserProfileAdd(cmd *cobra.Command, args []string) error {
	username := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine path
	var path string
	if profilePath != "" {
		path = profilePath
	} else {
		// Auto-detect
		if _, err := os.Stat("/Users"); err == nil {
			path = filepath.Join("/Users", username)
		} else {
			path = filepath.Join("/home", username)
		}
	}

	path = expandPath(path)
	absPath, _ := filepath.Abs(path)

	user := config.User{
		Name: username,
		Path: absPath,
	}

	if err := cfg.AddUser(user); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Profile added!"))
	fmt.Printf("  Name: %s\n", theme.HighlightStyle.Render(username))
	fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(absPath))
	fmt.Println()

	return nil
}

func runUserProfileRemove(cmd *cobra.Command, args []string) error {
	username := args[0]

	if !confirmRemove {
		fmt.Println(theme.ErrorStyle.Render("Use --confirm to remove profile"))
		return fmt.Errorf("confirmation required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteUser(username); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Profile removed"))
	return nil
}

func runUserProfileList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to load config: " + err.Error()))
		return
	}

	users := cfg.ListUsers()

	fmt.Println()
	fmt.Println(theme.RenderBanner("📁 LOCAL PROFILES 📁"))
	fmt.Println()

	if len(users) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No profiles found"))
		fmt.Println()
		fmt.Printf("  Add one: %s\n", theme.HighlightStyle.Render("anime user profile add <name>"))
		fmt.Println()
		return
	}

	// Sort by name
	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	for _, user := range users {
		marker := "  "
		if cfg.ActiveUser == user.Name {
			marker = "★ "
		}
		fmt.Printf("%s%s  →  %s\n",
			marker,
			theme.HighlightStyle.Render(user.Name),
			theme.DimTextStyle.Render(user.Path))
	}

	fmt.Println()
}

func runUserProfileSet(cmd *cobra.Command, args []string) error {
	username := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetActiveUser(username); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Active profile set to: " + username))
	return nil
}

func runUserProfileCd(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var user *config.User

	if len(args) > 0 {
		user, err = cfg.GetUser(args[0])
	} else {
		user, err = cfg.GetActiveUser()
	}

	if err != nil {
		return err
	}

	fmt.Printf("cd %s", shellQuote(user.Path))
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func printUserStep(num int, title string) {
	fmt.Println(theme.CategoryStyle(fmt.Sprintf("╔══ Step %d: %s", num, title)))
}

func promptUserChoice(reader *bufio.Reader, prompt string, validChoices []string) string {
	for {
		fmt.Print(theme.HighlightStyle.Render(prompt + " ▶ "))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		for _, valid := range validChoices {
			if response == valid {
				return response
			}
		}

		fmt.Println(theme.ErrorStyle.Render("  ✗ Invalid choice. Please try again."))
	}
}

func promptUserYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	fmt.Print(theme.HighlightStyle.Render(fmt.Sprintf("%s (%s) ▶ ", prompt, defaultStr)))
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func isValidUsername(username string) bool {
	if len(username) == 0 || len(username) > 32 {
		return false
	}

	if username[0] < 'a' || username[0] > 'z' {
		return false
	}

	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}

	return true
}

func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	password := base64.URLEncoding.EncodeToString(bytes)[:length]
	password = strings.ReplaceAll(password, "-", "!")
	password = strings.ReplaceAll(password, "_", "@")
	return password
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

