package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/source"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// Flags for source commands
var (
	srcDryRun bool
	srcForce  bool
	srcServer string
	srcNoGit  bool
)

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Source control - sync code with remote servers",
	Long: `Source Control provides rsync-based code synchronization.

Code is synced to llamah:~/ by default.

COMMANDS:
  push        Push local changes to remote
  pull        Pull remote changes to local
  clone       Clone a remote repo into new folder
  status      Show sync status
  sync        Bidirectional sync
  link        Link current directory to remote path
  init        Initialize and push new repo
  list        List remote repositories
  tree        Tree view of remote repos
  history     Show push/pull history
  rename      Rename/move remote repo
  delete      Delete remote repo

EXAMPLES:
  anime source push                    # Push to linked remote
  anime source pull org/project        # Pull specific repo
  anime source clone org/project       # Clone into ./project
  anime source status                  # Show what's different
  anime source sync                    # Bidirectional sync

FLAGS:
  --server, -s   Override default server
  --dry-run, -n  Preview without changes
  --no-git       Exclude .git directory from sync`,
	Run: showSourceDashboard,
}

func showSourceDashboard(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("SOURCE CONTROL"))
	fmt.Println()

	// Check for linked repo
	linkedPath := source.GetLinkedPath()
	if linkedPath != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Linked to:"), theme.HighlightStyle.Render(linkedPath))
		fmt.Println()
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No linked repository (use 'anime source link <path>')"))
		fmt.Println()
	}

	// Show quick actions
	fmt.Println(theme.InfoStyle.Render("Quick Actions:"))
	fmt.Println()

	actions := []struct {
		cmd  string
		desc string
	}{
		{"anime source push", "Push local changes to remote"},
		{"anime source pull", "Pull remote changes to local"},
		{"anime source status", "Compare local and remote"},
		{"anime source sync", "Bidirectional sync"},
		{"anime source list", "List remote repositories"},
	}

	for _, a := range actions {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(a.cmd),
			theme.DimTextStyle.Render("- "+a.desc))
	}
	fmt.Println()

	// Show server info
	server := srcServer
	if server == "" {
		server = source.DefaultServer
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.InfoStyle.Render(server))
	fmt.Println()
}

var srcPushCmd = &cobra.Command{
	Use:   "push [server] [path]",
	Short: "Push local changes to remote",
	Long: `Push current directory to remote server.

If linked, the path argument is optional.

Examples:
  anime source push                    # Push to linked path (default server)
  anime source push myproject          # Push to ~/myproject
  anime source push lambda             # Push linked path to lambda server
  anime source push lambda myproject   # Push myproject to lambda
  anime source push lambda:~/code      # Push to absolute path on lambda
  anime source push -n myproject       # Dry run`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSourcePush,
}

var srcPullCmd = &cobra.Command{
	Use:   "pull [server] [path]",
	Short: "Pull remote changes to local",
	Long: `Pull from remote server to current directory.

Examples:
  anime source pull                    # Pull from linked path (default server)
  anime source pull org/project        # Pull specific repo from default server
  anime source pull lambda             # Pull linked path from lambda server
  anime source pull lambda myproject   # Pull myproject from lambda
  anime source pull lambda:~/code      # Pull absolute path from lambda`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSourcePull,
}

var srcCloneCmd = &cobra.Command{
	Use:   "clone <path>",
	Short: "Clone remote repo into new folder",
	Long: `Clone from remote into a new folder.

Examples:
  anime source clone myproject         # Creates ./myproject (no .git)
  anime source clone org/project       # Creates ./project (no .git)
  anime source clone --git myproject   # Clone with .git (like git clone)
  anime source clone -f org/project    # Force overwrite`,
	Args: cobra.ExactArgs(1),
	RunE: runSourceClone,
}

var srcStatusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show sync status",
	Long: `Compare local directory with remote.

Shows files that would be pushed or pulled.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSourceStatus,
}

var srcSyncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Bidirectional sync",
	Long: `Sync in both directions based on modification time.

Newer files win in both directions.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSourceSync,
}

var srcLinkCmd = &cobra.Command{
	Use:   "link <path>",
	Short: "Link current directory to remote path",
	Long: `Link current directory to a remote path.

After linking, push/pull/sync work without arguments.

Examples:
  anime source link myproject          # Link to ~/myproject
  anime source link org/project        # Link to ~/org/project`,
	Args: cobra.ExactArgs(1),
	RunE: runSourceLink,
}

var srcInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize and push new repo",
	Long: `Initialize current directory as repo and push.

This links and pushes in one step.`,
	Args: cobra.ExactArgs(1),
	RunE: runSourceInit,
}

var srcListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List remote repositories",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSourceList,
}

var srcTreeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Tree view of remote repos",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSourceTree,
}

var srcHistoryCmd = &cobra.Command{
	Use:   "history <path>",
	Short: "Show push/pull history",
	Args:  cobra.ExactArgs(1),
	RunE:  runSourceHistory,
}

var srcRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename/move remote repo",
	Args:  cobra.ExactArgs(2),
	RunE:  runSourceRename,
}

var srcDefaultCmd = &cobra.Command{
	Use:   "default [server:path]",
	Short: "Show or set source default server and path",
	Long: `Show or set the default source server and base path.

Examples:
  anime source default                       # Show current default
  anime source default llamah:~/             # Set default to llamah:~/
  anime source default alice:~/cpm/anime     # Set default to alice:~/cpm/anime`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSourceDefault,
}

var srcDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete remote repo",
	Args:  cobra.ExactArgs(1),
	RunE:  runSourceDelete,
}

func init() {
	// Global flags
	sourceCmd.PersistentFlags().StringVarP(&srcServer, "server", "s", "", "Override default server")
	sourceCmd.PersistentFlags().BoolVarP(&srcDryRun, "dry-run", "n", false, "Preview without changes")

	// Clone-specific flags
	srcCloneCmd.Flags().BoolVarP(&srcForce, "force", "f", false, "Force overwrite")

	// --no-git flag on push, pull, clone (default includes .git)
	srcPushCmd.Flags().BoolVar(&srcNoGit, "no-git", false, "Exclude .git directory")
	srcPullCmd.Flags().BoolVar(&srcNoGit, "no-git", false, "Exclude .git directory")
	srcCloneCmd.Flags().BoolVar(&srcNoGit, "no-git", false, "Exclude .git directory")

	// Add subcommands
	sourceCmd.AddCommand(srcPushCmd)
	sourceCmd.AddCommand(srcPullCmd)
	sourceCmd.AddCommand(srcCloneCmd)
	sourceCmd.AddCommand(srcStatusCmd)
	sourceCmd.AddCommand(srcSyncCmd)
	sourceCmd.AddCommand(srcLinkCmd)
	sourceCmd.AddCommand(srcInitCmd)
	sourceCmd.AddCommand(srcListCmd)
	sourceCmd.AddCommand(srcTreeCmd)
	sourceCmd.AddCommand(srcHistoryCmd)
	sourceCmd.AddCommand(srcRenameCmd)
	sourceCmd.AddCommand(srcDeleteCmd)
	sourceCmd.AddCommand(srcDefaultCmd)

	rootCmd.AddCommand(sourceCmd)
}

func getSourceConfig() (*source.Config, error) {
	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SSH key: %w", err)
	}

	return &source.Config{
		Server:           getSourceServer(),
		DryRun:           srcDryRun,
		Force:            srcForce,
		IncludeGit:       !srcNoGit,
		KeyPath:          keyPath,
		Cleanup:          cleanup,
		BasePathOverride: getSourceBasePath(),
	}, nil
}

func getSourceServer() string {
	if srcServer != "" {
		return srcServer
	}
	if cfg, err := config.Load(); err == nil {
		if s := cfg.GetSourceServer(); s != "" {
			return s
		}
	}
	return source.DefaultServer
}

func getSourceBasePath() string {
	if cfg, err := config.Load(); err == nil {
		if p := cfg.GetSourceBasePath(); p != "" {
			return p
		}
	}
	return source.BasePath
}

func getSourceTarget() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return resolveSSHTarget(cfg, getSourceServer())
}

func resolveRemotePathSource(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return source.GetLinkedPath()
}

// parseServerAndPath parses arguments to extract server and path
// Supports:
//   - "lambda" (server only, if it's a known alias)
//   - "lambda:~/path" (server:path syntax)
//   - "lambda path" (server + path as separate args)
//   - "path" (path only, use default server)
// Returns: server (or empty for default), path, isAbsolutePath
func parseServerAndPath(args []string) (server string, remotePath string, isAbsolute bool) {
	if len(args) == 0 {
		return "", source.GetLinkedPath(), false
	}

	firstArg := args[0]

	// Check for server:path syntax
	if strings.Contains(firstArg, ":") {
		parts := strings.SplitN(firstArg, ":", 2)
		server = parts[0]
		remotePath = parts[1]
		// Check if path starts with ~ or / (absolute)
		isAbsolute = strings.HasPrefix(remotePath, "~") || strings.HasPrefix(remotePath, "/")
		return server, remotePath, isAbsolute
	}

	// Check if first arg is a known server alias
	cfg, err := config.Load()
	if err == nil {
		if target := cfg.GetAlias(firstArg); target != "" {
			// It's a server alias
			server = firstArg
			if len(args) > 1 {
				remotePath = args[1]
				isAbsolute = strings.HasPrefix(remotePath, "~") || strings.HasPrefix(remotePath, "/")
			} else {
				remotePath = source.GetLinkedPath()
			}
			return server, remotePath, isAbsolute
		}
	}

	// Not a server alias, treat as path
	return "", firstArg, false
}

// getSourceTargetForServer resolves SSH target for a specific server
func getSourceTargetForServer(server string) (string, error) {
	if server == "" {
		server = getSourceServer()
	}
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return resolveSSHTarget(cfg, server)
}

func runSourcePush(cmd *cobra.Command, args []string) error {
	// Parse server and path from args
	server, remotePath, isAbsolute := parseServerAndPath(args)

	// Use -s flag if provided, otherwise use parsed server
	if srcServer != "" {
		server = srcServer
	}

	target, err := getSourceTargetForServer(server)
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("PUSH (DRY RUN)"))
	} else {
		fmt.Println(theme.InfoStyle.Render("PUSH"))
	}
	fmt.Println()

	// Determine display path and actual path
	displayServer := server
	if displayServer == "" {
		displayServer = source.DefaultServer
	}

	var fullPath string
	if isAbsolute {
		fullPath = remotePath
	} else if remotePath != "" {
		fullPath = filepath.Join(getSourceBasePath(), remotePath)
	} else {
		fullPath = getSourceBasePath()
	}
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("To:"), theme.HighlightStyle.Render(displayServer), theme.InfoStyle.Render(fullPath))
	fmt.Println()

	// Set absolute flag in config for Push to use
	cfg.AbsolutePath = isAbsolute

	if err := source.Push(target, remotePath, cfg); err != nil {
		return err
	}

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("  Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Push complete"))
	}
	fmt.Println()

	return nil
}

func runSourcePull(cmd *cobra.Command, args []string) error {
	// Parse server and path from args
	server, remotePath, isAbsolute := parseServerAndPath(args)

	// Use -s flag if provided, otherwise use parsed server
	if srcServer != "" {
		server = srcServer
	}

	target, err := getSourceTargetForServer(server)
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("PULL (DRY RUN)"))
	} else {
		fmt.Println(theme.InfoStyle.Render("PULL"))
	}
	fmt.Println()

	// Determine display path and actual path
	displayServer := server
	if displayServer == "" {
		displayServer = source.DefaultServer
	}

	var fullPath string
	if isAbsolute {
		fullPath = remotePath
	} else if remotePath != "" {
		fullPath = filepath.Join(getSourceBasePath(), remotePath)
	} else {
		fullPath = getSourceBasePath()
	}
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(displayServer), theme.InfoStyle.Render(fullPath))
	fmt.Println()

	// Set absolute flag in config for Pull to use
	cfg.AbsolutePath = isAbsolute

	if err := source.Pull(target, remotePath, cfg); err != nil {
		return err
	}

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("  Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Pull complete"))
	}
	fmt.Println()

	return nil
}

func runSourceClone(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	label := "CLONE"
	if srcNoGit {
		label = "CLONE (no .git)"
	}
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render(label + " (DRY RUN)"))
	} else {
		fmt.Println(theme.InfoStyle.Render(label))
	}
	fmt.Println()

	fullPath := filepath.Join(getSourceBasePath(), remotePath)
	folderName := filepath.Base(remotePath)
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(getSourceServer()), theme.InfoStyle.Render(fullPath))
	fmt.Printf("  %s ./%s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(folderName))
	fmt.Println()

	if err := source.Clone(target, remotePath, cfg); err != nil {
		return err
	}

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("  Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Clone complete"))
	}
	fmt.Println()

	return nil
}

func runSourceStatus(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePathSource(args)
	if remotePath == "" {
		return fmt.Errorf("no path specified and no linked path found")
	}

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("STATUS"))
	fmt.Println()

	status, err := source.GetStatus(target, remotePath, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repo:"), theme.HighlightStyle.Render(remotePath))
	fmt.Println()

	if len(status.ToPush) > 0 {
		fmt.Println(theme.InfoStyle.Render("  To push:"))
		for _, f := range status.ToPush {
			fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("->"), f)
		}
		fmt.Println()
	}

	if len(status.ToPull) > 0 {
		fmt.Println(theme.WarningStyle.Render("  To pull:"))
		for _, f := range status.ToPull {
			fmt.Printf("    %s %s\n", theme.WarningStyle.Render("<-"), f)
		}
		fmt.Println()
	}

	if status.InSync {
		fmt.Println(theme.SuccessStyle.Render("  In sync"))
	} else {
		fmt.Printf("  %d to push, %d to pull\n", len(status.ToPush), len(status.ToPull))
	}
	fmt.Println()

	return nil
}

func runSourceSync(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePathSource(args)
	if remotePath == "" {
		return fmt.Errorf("no path specified and no linked path found")
	}

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("SYNC (DRY RUN)"))
	} else {
		fmt.Println(theme.InfoStyle.Render("SYNC"))
	}
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repo:"), theme.HighlightStyle.Render(remotePath))
	fmt.Println()

	if err := source.Sync(target, remotePath, cfg); err != nil {
		return err
	}

	fmt.Println()
	if srcDryRun {
		fmt.Println(theme.WarningStyle.Render("  Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Sync complete"))
	}
	fmt.Println()

	return nil
}

func runSourceLink(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("LINK"))
	fmt.Println()
	fmt.Printf("  %s %s:~/%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(getSourceServer()), remotePath)
	fmt.Println()

	if err := source.SaveLink(remotePath, getSourceServer()); err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}

	// Add to .gitignore if exists
	if _, err := os.Stat(".gitignore"); err == nil {
		gitignore, _ := os.ReadFile(".gitignore")
		if !strings.Contains(string(gitignore), source.LinkFile) {
			f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString("\n" + source.LinkFile + "\n")
				f.Close()
				fmt.Println(theme.DimTextStyle.Render("  Added to .gitignore"))
			}
		}
	}

	fmt.Println(theme.SuccessStyle.Render("  Linked"))
	fmt.Println()

	return nil
}

func runSourceInit(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("INIT"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.InfoStyle.Render(remotePath))
	fmt.Println()

	// Create link
	if err := source.SaveLink(remotePath, getSourceServer()); err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}
	fmt.Println(theme.DimTextStyle.Render("  Created link file"))

	// Push
	return runSourcePush(cmd, []string{remotePath})
}

func runSourceList(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) == 1 {
		path = args[0]
	}

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fullPath := getSourceBasePath()
	if path != "" {
		fullPath = filepath.Join(getSourceBasePath(), path)
	}
	fmt.Printf("  %s:%s\n", theme.HighlightStyle.Render(getSourceServer()), theme.InfoStyle.Render(fullPath))
	fmt.Println()

	repos, err := source.ListRepos(target, path, cfg)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  (empty)"))
	} else {
		for _, repo := range repos {
			fmt.Printf("  %s\n", repo)
		}
	}
	fmt.Println()

	return nil
}

func runSourceTree(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) == 1 {
		path = args[0]
	}

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fullPath := getSourceBasePath()
	if path != "" {
		fullPath = filepath.Join(getSourceBasePath(), path)
	}
	fmt.Printf("  %s:%s\n", theme.HighlightStyle.Render(getSourceServer()), theme.InfoStyle.Render(fullPath))
	fmt.Println()

	// Use SSH to run tree command
	treeCmd := fmt.Sprintf("tree -L 3 %s 2>/dev/null || find %s -maxdepth 3 -type d 2>/dev/null | head -50", fullPath, fullPath)

	sshArgs := []string{
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		treeCmd,
	}

	sshCmd := newSSHCommand(sshArgs...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Run()

	fmt.Println()
	return nil
}

func runSourceHistory(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("HISTORY"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repo:"), theme.HighlightStyle.Render(remotePath))
	fmt.Println()

	entries, err := source.GetHistory(target, remotePath, cfg)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No history"))
	} else {
		for _, e := range entries {
			t, err := time.Parse(time.RFC3339, e.Timestamp)
			timestamp := e.Timestamp
			if err == nil {
				timestamp = t.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("  %s  %s from %s\n",
				theme.DimTextStyle.Render(timestamp),
				theme.InfoStyle.Render(e.Action),
				theme.HighlightStyle.Render(e.Hostname))
		}
	}
	fmt.Println()

	return nil
}

func runSourceRename(cmd *cobra.Command, args []string) error {
	oldPath := args[0]
	newPath := args[1]

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("RENAME"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("From:"), theme.InfoStyle.Render(oldPath))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(newPath))
	fmt.Println()

	if err := source.Rename(target, oldPath, newPath, cfg); err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render("  Renamed"))
	fmt.Println()

	return nil
}

func runSourceDefault(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// No args: show current default
	if len(args) == 0 {
		server := cfg.GetSourceServer()
		basePath := cfg.GetSourceBasePath()
		if server == "" {
			server = source.DefaultServer
		}
		if basePath == "" {
			basePath = getSourceBasePath()
		}
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("SOURCE DEFAULT"))
		fmt.Println()
		fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Default:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(basePath))
		fmt.Println()
		return nil
	}

	// Parse server:path
	arg := args[0]
	if !strings.Contains(arg, ":") {
		return fmt.Errorf("expected format server:path (e.g., llamah:~/)")
	}
	parts := strings.SplitN(arg, ":", 2)
	server := parts[0]
	basePath := parts[1]

	// Normalize: strip trailing slash unless it's just "~/"
	if basePath != "~/" && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimRight(basePath, "/")
	}
	// "~/" -> "~"
	if basePath == "~/" {
		basePath = "~"
	}

	cfg.SetSourceDefault(server, basePath)
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("SOURCE DEFAULT SET"))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Default:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(basePath))
	fmt.Println()

	return nil
}

func runSourceDelete(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getSourceTarget()
	if err != nil {
		return err
	}

	cfg, err := getSourceConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("DELETE"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Target:"), theme.ErrorStyle.Render(remotePath))
	fmt.Println()

	// Confirm
	fmt.Print(theme.WarningStyle.Render("  Type 'yes' to confirm: "))
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println(theme.DimTextStyle.Render("  Cancelled"))
		fmt.Println()
		return nil
	}

	if err := source.Delete(target, remotePath, cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Deleted"))
	fmt.Println()

	return nil
}

// Helper to create SSH command
func newSSHCommand(args ...string) *exec.Cmd {
	return exec.Command("ssh", args...)
}
