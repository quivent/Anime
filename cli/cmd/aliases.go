package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Manage shell aliases for anime commands",
	Long:  "Create, list, and manage shell aliases for anime commands",
	Run:   runAliasesHelp,
}

var aliasesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all anime shell aliases",
	Long:  "Display all registered shell aliases and their commands",
	Run:   runAliasesList,
}

var aliasesAddCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a new shell alias",
	Long: `Add a new shell alias for an anime command.

Examples:
  anime aliases add q "anime query"
  anime aliases add gen "anime generate"
  anime aliases add up "anime upscale"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("Missing required arguments"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime aliases add <name> <command>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Examples:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime aliases add q \"anime query\""))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime aliases add gen \"anime generate\""))
			fmt.Println()
			return fmt.Errorf("requires alias name and command")
		}
		return nil
	},
	RunE: runAliasesAdd,
}

var aliasesRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a shell alias",
	Long:    "Remove a shell alias from the configuration",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime aliases remove <name>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Example:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime aliases remove codec"))
			fmt.Println()
			return fmt.Errorf("requires alias name")
		}
		return nil
	},
	RunE: runAliasesRemove,
}

var aliasesInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install aliases to shell config (.zshrc/.bashrc)",
	Long: `Install all registered aliases to your shell configuration file.

This will add aliases to your .zshrc (or .bashrc) so they're available
in new terminal sessions.`,
	RunE: runAliasesInstall,
}

func init() {
	aliasesCmd.AddCommand(aliasesListCmd)
	aliasesCmd.AddCommand(aliasesAddCmd)
	aliasesCmd.AddCommand(aliasesRemoveCmd)
	aliasesCmd.AddCommand(aliasesInstallCmd)
	rootCmd.AddCommand(aliasesCmd)
}

// Default aliases that come with anime
var defaultAliases = map[string]string{
	"code":  "claude --permission-mode bypassPermissions",
	"codec": "claude --permission-mode bypassPermissions --continue",
	"coder": "claude --permission-mode bypassPermissions --resume",
}

func runAliasesHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("SHELL ALIASES"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Manage shell aliases for anime commands"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime aliases list", "List all registered aliases"},
		{"anime aliases add <name> <command>", "Add a new alias"},
		{"anime aliases remove <name>", "Remove an alias"},
		{"anime aliases install", "Install aliases to shell config"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Quick Start"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases list"))
	fmt.Println(theme.DimTextStyle.Render("    See all available aliases"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases install"))
	fmt.Println(theme.DimTextStyle.Render("    Install aliases to your shell"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases add q \"anime query\""))
	fmt.Println(theme.DimTextStyle.Render("    Add a custom alias"))
	fmt.Println()
}

func runAliasesList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to load config: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("SHELL ALIASES"))
	fmt.Println()

	// Get shell aliases from config
	shellAliases := cfg.GetShellAliases()

	// Merge with defaults (defaults can be overridden)
	allAliases := make(map[string]string)
	for name, command := range defaultAliases {
		allAliases[name] = command
	}
	for name, command := range shellAliases {
		allAliases[name] = command
	}

	if len(allAliases) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No aliases configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Add your first alias:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases add q \"anime query\""))
		fmt.Println()
		return
	}

	fmt.Printf("  Total aliases: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(allAliases))))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Registered Aliases"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for name, command := range allAliases {
		isDefault := false
		if _, ok := defaultAliases[name]; ok {
			if shellAliases[name] == "" || shellAliases[name] == defaultAliases[name] {
				isDefault = true
			}
		}

		if isDefault {
			fmt.Printf("  %s → %s %s\n",
				theme.HighlightStyle.Render(name),
				theme.DimTextStyle.Render(command),
				theme.DimTextStyle.Render("(default)"))
		} else {
			fmt.Printf("  %s → %s\n",
				theme.HighlightStyle.Render(name),
				theme.DimTextStyle.Render(command))
		}
	}

	fmt.Println()

	// Check if installed in shell
	installed := checkAliasesInstalled()
	if installed {
		fmt.Println(theme.SuccessStyle.Render("  Status: Installed in shell config"))
	} else {
		fmt.Println(theme.WarningStyle.Render("  Status: Not installed in shell config"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  To install:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases install"))
	}
	fmt.Println()
}

func runAliasesAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	command := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add the shell alias
	if err := cfg.AddShellAlias(name, command); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Alias added successfully!"))
	fmt.Println()
	fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(name), theme.DimTextStyle.Render(command))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("To activate in your shell:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime aliases install"))
	fmt.Println()

	return nil
}

func runAliasesRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.RemoveShellAlias(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("Alias '%s' removed", name)))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Note: Run 'anime aliases install' to update your shell config"))
	fmt.Println()

	return nil
}

func runAliasesInstall(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("INSTALL ALIASES"))
	fmt.Println()

	// Determine shell config file
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shellConfigFile := filepath.Join(home, ".bashrc")
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "zsh") {
		shellConfigFile = filepath.Join(home, ".zshrc")
	}

	// Check if file exists
	_, err = os.Stat(shellConfigFile)
	if err != nil {
		return fmt.Errorf("shell config file not found: %s", shellConfigFile)
	}

	// Load config to get all aliases
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get shell aliases from config
	shellAliases := cfg.GetShellAliases()

	// Merge with defaults
	allAliases := make(map[string]string)
	for name, command := range defaultAliases {
		allAliases[name] = command
	}
	for name, command := range shellAliases {
		allAliases[name] = command
	}

	// Build alias script
	var aliasLines []string
	for name, command := range allAliases {
		aliasLines = append(aliasLines, fmt.Sprintf("alias %s=\"%s\"", name, command))
	}

	aliasScript := fmt.Sprintf(`
# ─────────────────────────────────────────────────────────────
# Anime CLI Aliases - Generated by anime
# ─────────────────────────────────────────────────────────────
%s
# ─────────────────────────────────────────────────────────────
`, strings.Join(aliasLines, "\n"))

	// Read current config
	content, err := os.ReadFile(shellConfigFile)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Check if aliases already exist - if so, replace them
	if strings.Contains(contentStr, "# Anime CLI Aliases - Generated by anime") {
		// Remove existing block and add new one
		startMarker := "# ─────────────────────────────────────────────────────────────\n# Anime CLI Aliases - Generated by anime"
		endMarker := "# ─────────────────────────────────────────────────────────────\n"

		startIdx := strings.Index(contentStr, startMarker)
		if startIdx != -1 {
			// Find the end marker after the start
			afterStart := contentStr[startIdx+len(startMarker):]
			endIdx := strings.Index(afterStart, endMarker)
			if endIdx != -1 {
				// Remove the old block (including the trailing end marker)
				beforeBlock := contentStr[:startIdx]
				afterBlock := afterStart[endIdx+len(endMarker):]
				contentStr = beforeBlock + afterBlock
			}
		}

		// Write updated content
		if err := os.WriteFile(shellConfigFile, []byte(contentStr+aliasScript), 0644); err != nil {
			return err
		}

		fmt.Println(theme.SuccessStyle.Render("Aliases updated!"))
	} else {
		// Append aliases
		f, err := os.OpenFile(shellConfigFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(aliasScript); err != nil {
			return err
		}

		fmt.Println(theme.SuccessStyle.Render("Aliases installed!"))
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Installed aliases:"))
	for name, command := range allAliases {
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(name), theme.DimTextStyle.Render(command))
	}
	fmt.Println()

	// Automatically reload shell config by exec'ing into a new shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}

	fmt.Println(theme.SuccessStyle.Render("Reloading shell to activate aliases..."))
	fmt.Println()

	// Exec replaces the current process with a new shell
	err = syscall.Exec(shell, []string{shell}, os.Environ())
	if err != nil {
		// If exec fails, fall back to manual instructions
		fmt.Println(theme.WarningStyle.Render("Could not auto-reload shell"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("To activate aliases manually:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("source %s", shellConfigFile)))
		fmt.Println()
	}

	return nil
}

func checkAliasesInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	shellConfigFile := filepath.Join(home, ".bashrc")
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "zsh") {
		shellConfigFile = filepath.Join(home, ".zshrc")
	}

	content, err := os.ReadFile(shellConfigFile)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "# Anime CLI Aliases - Generated by anime")
}
