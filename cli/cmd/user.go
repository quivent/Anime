package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user profiles",
	Long:  "Create, list, and manage user profiles for quick navigation",
	Run:   runUserHelp,
}

var userAddCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Add a new user",
	Long: `Add a new user profile with their home directory.

The user's home directory will be auto-detected from the system or can be specified.

Examples:
  anime user add josh
  anime user add maria`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing username"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime user add <username>"))
			fmt.Println()
			return fmt.Errorf("requires username")
		}
		return nil
	},
	RunE: runUserAdd,
}

var userRemoveCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove a user",
	Long: `Remove a user profile from the configuration.

This does not delete any files, only removes the user from the configuration.

Examples:
  anime user remove josh --confirm`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing username"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime user remove <username> --confirm"))
			fmt.Println()
			return fmt.Errorf("requires username")
		}
		return nil
	},
	RunE: runUserRemove,
}

var userListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all users",
	Long:    "Display all configured user profiles",
	Run:     runUserList,
}

var userSetCmd = &cobra.Command{
	Use:   "set <username>",
	Short: "Set the active user",
	Long: `Set the active user profile for quick navigation.

Once set, you can use 'anime user cd' to quickly navigate to the active user's directory.

Examples:
  anime user set josh
  anime user cd`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing username"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime user set <username>"))
			fmt.Println()
			return fmt.Errorf("requires username")
		}
		return nil
	},
	RunE: runUserSet,
}

var userCdCmd = &cobra.Command{
	Use:   "cd [username]",
	Short: "Print cd command for user directory",
	Long: `Output a cd command to navigate to a user's directory.

If no username is provided, uses the currently active user.

Examples:
  anime user cd josh
  anime user cd          # Uses active user
  $(anime user cd)       # Execute the cd command`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserCd,
}

// Create "users" command as an alias to "user list"
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users (alias for 'user list')",
	Long:  "Display all configured user profiles",
	Run:   runUserList,
}

var (
	confirmRemove bool
	userPath      string
)

func init() {
	// Add flags
	userRemoveCmd.Flags().BoolVar(&confirmRemove, "confirm", false, "Confirm user removal")
	userAddCmd.Flags().StringVarP(&userPath, "path", "p", "", "User home directory path")

	// Add subcommands to user command
	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userRemoveCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userSetCmd)
	userCmd.AddCommand(userCdCmd)

	// Add commands to root
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(usersCmd)
}

func runUserHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("👤 USER MANAGEMENT 👤"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Manage user profiles for quick navigation"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📋 Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime user add <username>", "Add a new user profile"},
		{"anime user remove <username> --confirm", "Remove a user profile"},
		{"anime user list", "List all users (or: anime users)"},
		{"anime user set <username>", "Set the active user"},
		{"anime user cd [username]", "Print cd command to user directory"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Example Workflow"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime user add josh"))
	fmt.Println(theme.DimTextStyle.Render("    Add user 'josh'"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime user set josh"))
	fmt.Println(theme.DimTextStyle.Render("    Set josh as active user"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ $(anime user cd)"))
	fmt.Println(theme.DimTextStyle.Render("    Navigate to josh's home directory"))
	fmt.Println()
}

func runUserAdd(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine user path
	var path string
	if userPath != "" {
		// Use provided path
		path = userPath
	} else {
		// Auto-detect from system
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Check if we're on macOS (Users directory) or Linux (home directory)
		if _, err := os.Stat("/Users"); err == nil {
			// macOS
			path = filepath.Join("/Users", username)
		} else {
			// Linux/Unix
			path = filepath.Join("/home", username)
		}

		// If the detected path doesn't exist, check if user wants to use current user's home
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Try using the current home directory's parent + username
			parentDir := filepath.Dir(homeDir)
			candidatePath := filepath.Join(parentDir, username)
			if _, err := os.Stat(candidatePath); err == nil {
				path = candidatePath
			} else {
				// Default to the original guess
				fmt.Println()
				fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠️  Warning: Directory %s does not exist", path)))
				fmt.Println(theme.DimTextStyle.Render("    You can specify a custom path with --path flag"))
				fmt.Println()
			}
		}
	}

	// Expand path
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create user
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

	// Success message
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ User added successfully!"))
	fmt.Println()
	fmt.Printf("  Name:  %s\n", theme.HighlightStyle.Render(username))
	fmt.Printf("  Path:  %s\n", theme.DimTextStyle.Render(absPath))
	fmt.Println()

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Println(theme.DimTextStyle.Render("  Note: Directory does not exist yet"))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("💡 Next steps:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$ anime user set %s", username)))
	fmt.Println(theme.DimTextStyle.Render("    Set as active user"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$ $(anime user cd %s)", username)))
	fmt.Println(theme.DimTextStyle.Render("    Navigate to user directory"))
	fmt.Println()

	return nil
}

func runUserRemove(cmd *cobra.Command, args []string) error {
	username := args[0]

	if !confirmRemove {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Confirmation required"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("To remove a user, you must confirm with --confirm flag:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime user remove %s --confirm", username)))
		fmt.Println()
		return fmt.Errorf("confirmation required")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get user info before deletion for display
	user, err := cfg.GetUser(username)
	if err != nil {
		return err
	}

	// Delete user
	if err := cfg.DeleteUser(username); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Success message
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ User removed successfully!"))
	fmt.Println()
	fmt.Printf("  Name:  %s\n", theme.HighlightStyle.Render(username))
	fmt.Printf("  Path:  %s\n", theme.DimTextStyle.Render(user.Path))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Note: User files were not deleted"))
	fmt.Println()

	return nil
}

func runUserList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to load config: " + err.Error()))
		return
	}

	users := cfg.ListUsers()

	fmt.Println()
	fmt.Println(theme.RenderBanner("👥 USERS 👥"))
	fmt.Println()

	if len(users) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No users found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Add your first user:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime user add <username>"))
		fmt.Println()
		return
	}

	fmt.Printf("  Total users: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(users))))
	if cfg.ActiveUser != "" {
		fmt.Printf("  Active user: %s\n", theme.SuccessStyle.Render(cfg.ActiveUser))
	}
	fmt.Println()

	// Sort users by name for consistent output
	sortedUsers := make([]config.User, len(users))
	copy(sortedUsers, users)
	sort.Slice(sortedUsers, func(i, j int) bool {
		return sortedUsers[i].Name < sortedUsers[j].Name
	})

	// Find max name length for alignment
	maxLen := 0
	for _, user := range sortedUsers {
		if len(user.Name) > maxLen {
			maxLen = len(user.Name)
		}
	}

	// Print each user
	for _, user := range sortedUsers {
		padding := strings.Repeat(" ", maxLen-len(user.Name))
		isActive := cfg.ActiveUser == user.Name

		// Check if directory exists
		exists := false
		if _, err := os.Stat(user.Path); err == nil {
			exists = true
		}

		nameDisplay := user.Name
		if isActive {
			nameDisplay = user.Name + " ★"
		}

		statusIcon := "📁"
		if !exists {
			statusIcon = "❓"
		}

		fmt.Printf("  %s %s%s  →  %s\n",
			statusIcon,
			theme.HighlightStyle.Render(nameDisplay),
			padding,
			theme.DimTextStyle.Render(user.Path))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Total: %d user(s)", len(users))))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("💡 Quick actions:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime user set <username>"))
	fmt.Println(theme.DimTextStyle.Render("    Set active user"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$(anime user cd [username])"))
	fmt.Println(theme.DimTextStyle.Render("    Navigate to user directory"))
	fmt.Println()
}

func runUserSet(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set active user
	if err := cfg.SetActiveUser(username); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Get user info for display
	user, _ := cfg.GetUser(username)

	// Success message
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Active user set!"))
	fmt.Println()
	fmt.Printf("  User:  %s\n", theme.HighlightStyle.Render(username))
	fmt.Printf("  Path:  %s\n", theme.DimTextStyle.Render(user.Path))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Quick navigation:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$(anime user cd)"))
	fmt.Println(theme.DimTextStyle.Render("    Navigate to active user's directory"))
	fmt.Println()

	return nil
}

func runUserCd(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var user *config.User

	if len(args) > 0 {
		// Username provided
		username := args[0]
		user, err = cfg.GetUser(username)
		if err != nil {
			return err
		}
	} else {
		// No username provided, use active user
		user, err = cfg.GetActiveUser()
		if err != nil {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ No active user set"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Set an active user first:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime user set <username>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Or specify a user:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("$(anime user cd <username>)"))
			fmt.Println()
			return err
		}
	}

	// Check if directory exists
	if _, err := os.Stat(user.Path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s\n", theme.WarningStyle.Render(fmt.Sprintf("⚠️  Warning: Directory does not exist: %s", user.Path)))
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Output the cd command
	// This allows the user to execute: $(anime user cd)
	fmt.Printf("cd %s", shellQuote(user.Path))

	return nil
}
