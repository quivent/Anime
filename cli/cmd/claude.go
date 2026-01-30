package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

//go:embed all:embedded/agents
var embeddedAgents embed.FS

//go:embed all:embedded/commands
var embeddedCommands embed.FS

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude Code agents and commands",
	Long:  "Manage embedded Claude Code agents and slash commands for portable CLI deployment",
	Run:   runClaudeHelp,
}

// Agents subcommand
var claudeAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage embedded Claude Code agents",
	Long:  "List, pull, and push Claude Code agents to/from the binary",
	Run:   runClaudeAgentsHelp,
}

var claudeAgentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List embedded agents",
	Long:  "Display all agents currently embedded in the binary",
	Run:   runClaudeAgentsList,
}

var claudeAgentsPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull agents from ~/.claude/agents into embedded directory",
	Long: `Pull agents from ~/.claude/agents/ and copy them to the embedded directory.

After running this command, you must rebuild the binary to embed the agents:
  go build -o anime .

The agents will then be embedded in the binary and can be pushed to new machines.`,
	RunE: runClaudeAgentsPull,
}

var claudeAgentsPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push embedded agents to ~/.claude/agents",
	Long: `Push all embedded agents to ~/.claude/agents/ on this machine.

This is useful for setting up Claude Code on a new machine with your agents.`,
	RunE: runClaudeAgentsPush,
}

// Commands subcommand
var claudeCommandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Manage embedded Claude Code slash commands",
	Long:  "List, pull, and push Claude Code slash commands to/from the binary",
	Run:   runClaudeCommandsHelp,
}

var claudeCommandsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List embedded commands",
	Long:  "Display all slash commands currently embedded in the binary",
	Run:   runClaudeCommandsList,
}

var claudeCommandsPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull commands from ~/.claude/commands into embedded directory",
	Long: `Pull commands from ~/.claude/commands/ and copy them to the embedded directory.

After running this command, you must rebuild the binary to embed the commands:
  go build -o anime .

The commands will then be embedded in the binary and can be pushed to new machines.`,
	RunE: runClaudeCommandsPull,
}

var claudeCommandsPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push embedded commands to ~/.claude/commands",
	Long: `Push all embedded commands to ~/.claude/commands/ on this machine.

This is useful for setting up Claude Code on a new machine with your slash commands.`,
	RunE: runClaudeCommandsPush,
}

func init() {
	// Register agents subcommands
	claudeAgentsCmd.AddCommand(claudeAgentsListCmd)
	claudeAgentsCmd.AddCommand(claudeAgentsPullCmd)
	claudeAgentsCmd.AddCommand(claudeAgentsPushCmd)

	// Register commands subcommands
	claudeCommandsCmd.AddCommand(claudeCommandsListCmd)
	claudeCommandsCmd.AddCommand(claudeCommandsPullCmd)
	claudeCommandsCmd.AddCommand(claudeCommandsPushCmd)

	// Register to claude command
	claudeCmd.AddCommand(claudeAgentsCmd)
	claudeCmd.AddCommand(claudeCommandsCmd)

	// Register to root
	rootCmd.AddCommand(claudeCmd)
}

func runClaudeHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CLAUDE CODE MANAGER"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Manage Claude Code agents and slash commands embedded in this binary"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime claude agents list", "List embedded agents"},
		{"anime claude agents pull", "Pull agents from ~/.claude/agents to embed"},
		{"anime claude agents push", "Push embedded agents to ~/.claude/agents"},
		{"anime claude commands list", "List embedded slash commands"},
		{"anime claude commands pull", "Pull commands from ~/.claude/commands to embed"},
		{"anime claude commands push", "Push embedded commands to ~/.claude/commands"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}
}

func runClaudeAgentsHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CLAUDE AGENTS"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Manage Claude Code agents embedded in this binary"))
	fmt.Println()

	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("list"), theme.DimTextStyle.Render("List embedded agents"))
	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("pull"), theme.DimTextStyle.Render("Pull from ~/.claude/agents to embedded dir (requires rebuild)"))
	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("push"), theme.DimTextStyle.Render("Push embedded agents to ~/.claude/agents"))
	fmt.Println()
}

func runClaudeCommandsHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CLAUDE COMMANDS"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Manage Claude Code slash commands embedded in this binary"))
	fmt.Println()

	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("list"), theme.DimTextStyle.Render("List embedded commands"))
	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("pull"), theme.DimTextStyle.Render("Pull from ~/.claude/commands to embedded dir (requires rebuild)"))
	fmt.Printf("  %s - %s\n", theme.HighlightStyle.Render("push"), theme.DimTextStyle.Render("Push embedded commands to ~/.claude/commands"))
	fmt.Println()
}

func runClaudeAgentsList(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("EMBEDDED AGENTS"))
	fmt.Println()

	entries, err := listClaudeEmbeddedFiles(embeddedAgents, "embedded/agents")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to list embedded agents: " + err.Error()))
		return
	}

	if len(entries) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No agents embedded"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  To embed agents:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime claude agents pull"))
		fmt.Println(theme.DimTextStyle.Render("    Then rebuild: go build -o anime ."))
		fmt.Println()
		return
	}

	fmt.Printf("  %s embedded agents:\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(entries))))
	fmt.Println()

	for _, entry := range entries {
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(entry))
	}
	fmt.Println()
}

func runClaudeCommandsList(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("EMBEDDED COMMANDS"))
	fmt.Println()

	entries, err := listClaudeEmbeddedFiles(embeddedCommands, "embedded/commands")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to list embedded commands: " + err.Error()))
		return
	}

	if len(entries) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No commands embedded"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  To embed commands:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime claude commands pull"))
		fmt.Println(theme.DimTextStyle.Render("    Then rebuild: go build -o anime ."))
		fmt.Println()
		return
	}

	fmt.Printf("  %s embedded commands:\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(entries))))
	fmt.Println()

	for _, entry := range entries {
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(entry))
	}
	fmt.Println()
}

func runClaudeAgentsPull(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sourceDir := filepath.Join(home, ".claude", "agents")
	destDir, err := getEmbeddedSourceDir("agents")
	if err != nil {
		return err
	}

	return pullToEmbedded(sourceDir, destDir, "agents")
}

func runClaudeAgentsPush(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	destDir := filepath.Join(home, ".claude", "agents")
	return pushFromEmbedded(embeddedAgents, "embedded/agents", destDir, "agents")
}

func runClaudeCommandsPull(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sourceDir := filepath.Join(home, ".claude", "commands")
	destDir, err := getEmbeddedSourceDir("commands")
	if err != nil {
		return err
	}

	return pullToEmbedded(sourceDir, destDir, "commands")
}

func runClaudeCommandsPush(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	destDir := filepath.Join(home, ".claude", "commands")
	return pushFromEmbedded(embeddedCommands, "embedded/commands", destDir, "commands")
}

// Helper functions

func listClaudeEmbeddedFiles(fsys embed.FS, root string) ([]string, error) {
	var files []string

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Skip .gitkeep files
		if d.Name() == ".gitkeep" {
			return nil
		}
		// Get relative path from root
		relPath := strings.TrimPrefix(path, root+"/")
		files = append(files, relPath)
		return nil
	})

	return files, err
}

func getEmbeddedSourceDir(subdir string) (string, error) {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Go up to find the cli directory (assuming standard structure)
	// Try several potential locations
	possiblePaths := []string{
		filepath.Join(filepath.Dir(execPath), "cmd", "embedded", subdir),
		filepath.Join(filepath.Dir(execPath), "embedded", subdir),
		filepath.Join(filepath.Dir(execPath), "..", "cmd", "embedded", subdir),
		filepath.Join(filepath.Dir(execPath), "..", "embedded", subdir),
	}

	// Also try from current working directory
	cwd, err := os.Getwd()
	if err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(cwd, "cmd", "embedded", subdir),
			filepath.Join(cwd, "embedded", subdir),
			filepath.Join(cwd, "..", "cmd", "embedded", subdir),
		)
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find cmd/embedded/%s directory - make sure you're in the CLI source directory", subdir)
}

func pullToEmbedded(sourceDir, destDir, itemType string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Pulling %s from %s", itemType, sourceDir)))
	fmt.Println()

	// Check source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  Source directory does not exist: %s", sourceDir)))
		fmt.Println()
		return nil
	}

	// Clear existing files in dest (except .gitkeep)
	entries, err := os.ReadDir(destDir)
	if err != nil {
		return fmt.Errorf("failed to read destination directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}
		path := filepath.Join(destDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to clean destination: %w", err)
		}
	}

	// Copy files from source to dest
	var count int
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
			return err
		}

		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("+"), theme.DimTextStyle.Render(relPath))
		count++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	fmt.Println()
	if count == 0 {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  No %s found to pull", itemType)))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  Pulled %d %s", count, itemType)))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Now rebuild to embed:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("go build -o anime ."))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Or use make:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("make build"))
	}
	fmt.Println()

	return nil
}

func pushFromEmbedded(fsys embed.FS, root, destDir, itemType string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Pushing embedded %s to %s", itemType, destDir)))
	fmt.Println()

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	var count int
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .gitkeep files
		if d.Name() == ".gitkeep" {
			return nil
		}

		relPath := strings.TrimPrefix(path, root+"/")
		if relPath == "" || relPath == root {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read from embedded FS
		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Write file
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}

		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("+"), theme.DimTextStyle.Render(relPath))
		count++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to push files: %w", err)
	}

	fmt.Println()
	if count == 0 {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  No %s embedded to push", itemType)))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  To embed " + itemType + " first:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime claude %s pull", itemType)))
		fmt.Println(theme.DimTextStyle.Render("    Then rebuild: go build -o anime ."))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  Pushed %d %s to %s", count, itemType, destDir)))
	}
	fmt.Println()

	return nil
}
