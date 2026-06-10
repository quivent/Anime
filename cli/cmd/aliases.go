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

	isZsh := strings.Contains(os.Getenv("SHELL"), "zsh")
	shellConfigFile := filepath.Join(home, ".bashrc")
	if isZsh {
		shellConfigFile = filepath.Join(home, ".zshrc")
	}

	// Create the rc file if it doesn't exist (fresh boxes, minimal images)
	createdRc := false
	if _, err := os.Stat(shellConfigFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(shellConfigFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", shellConfigFile, err)
		}
		createdRc = true
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

	// Build alias block (sorted so reinstalls don't reshuffle the file)
	var aliasLines []string
	for name, command := range allAliases {
		aliasLines = append(aliasLines, fmt.Sprintf("alias %s='%s'", name, command))
	}
	sort.Strings(aliasLines)

	aliasBlock := "# >>> ANIME ALIASES START >>>\n" +
		strings.Join(aliasLines, "\n") + "\n" +
		"# <<< ANIME ALIASES END <<<\n"

	const startMarker = "# >>> ANIME ALIASES START >>>"
	const endMarker = "# <<< ANIME ALIASES END <<<"

	// Read current config
	content, err := os.ReadFile(shellConfigFile)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Remove old block if present
	if startIdx := strings.Index(contentStr, startMarker); startIdx != -1 {
		if endIdx := strings.Index(contentStr[startIdx:], endMarker); endIdx != -1 {
			end := startIdx + endIdx + len(endMarker)
			// Trim trailing newline after end marker
			if end < len(contentStr) && contentStr[end] == '\n' {
				end++
			}
			contentStr = contentStr[:startIdx] + contentStr[end:]
		}
	}

	// Also remove legacy block if present
	if strings.Contains(contentStr, "# Anime CLI Aliases - Generated by anime") {
		// Remove legacy format — from first divider line to closing divider
		lines := strings.Split(contentStr, "\n")
		var cleaned []string
		skip := false
		for _, line := range lines {
			if strings.Contains(line, "# Anime CLI Aliases - Generated by anime") {
				skip = true
				continue
			}
			if skip && strings.HasPrefix(line, "# ────") {
				skip = false
				continue
			}
			if strings.HasPrefix(line, "# ────") && !skip {
				// Leading divider before the marker — remove it
				continue
			}
			if !skip {
				cleaned = append(cleaned, line)
			}
		}
		contentStr = strings.Join(cleaned, "\n")
	}

	// Write updated content + new block
	if len(contentStr) > 0 && !strings.HasSuffix(contentStr, "\n") {
		contentStr += "\n"
	}
	if err := os.WriteFile(shellConfigFile, []byte(contentStr+aliasBlock), 0644); err != nil {
		return err
	}

	// On bash, login shells (SSH sessions) never read .bashrc on their own —
	// make sure the profile chain reaches it, or the aliases silently vanish.
	var wiredProfile string
	if !isZsh {
		wiredProfile, err = ensureLoginShellSourcesBashrc(home)
		if err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: could not wire login shell profile: " + err.Error()))
		}
	}

	fmt.Println(theme.SuccessStyle.Render("Aliases installed!"))
	fmt.Println()
	if createdRc {
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Created %s", shellConfigFile)))
	}
	if wiredProfile != "" {
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Added 'source ~/.bashrc' to %s so login shells load it", wiredProfile)))
	}
	fmt.Println(theme.InfoStyle.Render("Installed aliases:"))
	for name, command := range allAliases {
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(name), theme.DimTextStyle.Render(command))
	}
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Open a new terminal or run:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("source %s", shellConfigFile)))
	fmt.Println()

	return nil
}

// bash login shells read only the first existing of ~/.bash_profile,
// ~/.bash_login, ~/.profile — never ~/.bashrc. If that file doesn't source
// .bashrc, installed aliases won't appear in SSH sessions. Returns the file
// modified/created, or "" if the chain already reaches .bashrc.
func ensureLoginShellSourcesBashrc(home string) (string, error) {
	const sourceLine = "[ -f ~/.bashrc ] && . ~/.bashrc # added by anime aliases install"
	for _, name := range []string{".bash_profile", ".bash_login", ".profile"} {
		path := filepath.Join(home, name)
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if strings.Contains(string(content), ".bashrc") {
			return "", nil
		}
		text := string(content)
		if len(text) > 0 && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if err := os.WriteFile(path, []byte(text+sourceLine+"\n"), 0644); err != nil {
			return "", err
		}
		return path, nil
	}
	path := filepath.Join(home, ".bash_profile")
	if err := os.WriteFile(path, []byte(sourceLine+"\n"), 0644); err != nil {
		return "", err
	}
	return path, nil
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

	return strings.Contains(string(content), "# >>> ANIME ALIASES START >>>")
}
