package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

const cpmDefaultServer = "alice"
const cpmBasePath = "~/cpm/anime"
const cpmPackagesPath = "~/cpm/packages"
const cpmLinkFile = ".cpm-link"
const cpmHistoryFile = ".cpm-history"
const cpmPackageFile = "cpm.json"
const cpmInstalledFile = ".cpm-installed.json"

// Flags
var (
	cpmDryRun  bool
	cpmForce   bool
	cpmServer  string
	cpmGlobal  bool
	cpmVersion string
)

// CpmPackage represents package metadata
type CpmPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	License     string   `json:"license,omitempty"`
	Repository  string   `json:"repository,omitempty"`
}

// CpmInstalled tracks installed packages
type CpmInstalled struct {
	Packages map[string]InstalledPackage `json:"packages"`
}

// InstalledPackage represents an installed package
type InstalledPackage struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	Path        string `json:"path"`
	Global      bool   `json:"global"`
}

var cpmCmd = &cobra.Command{
	Use:   "cpm",
	Short: "Code Push Manager - source control and package manager",
	Long: `CPM (Code Push Manager) provides rsync-based source control and package management.

SOURCE CONTROL:
  Code is synced to alice:~/cpm/anime by default.

PACKAGE MANAGEMENT:
  Packages are stored at alice:~/cpm/packages with versioning support.

Global Flags:
  --server, -s    Override default server (default: alice)
  --dry-run, -n   Preview what would be transferred without doing it

Source Control Examples:
  anime cpm push                        # Push current dir to alice
  anime cpm pull org/project            # Pull from alice into current dir
  anime cpm clone org/project           # Clone from alice into ./project
  anime cpm status                      # Show sync status
  anime cpm sync org/project            # Bidirectional sync

Package Management Examples:
  anime cpm publish                     # Publish current dir as package
  anime cpm install mypackage           # Install a package
  anime cpm install mypackage@1.0.0     # Install specific version
  anime cpm search utils                # Search for packages
  anime cpm info mypackage              # Show package info
  anime cpm versions mypackage          # List available versions
  anime cpm update                      # Update all installed packages`,
}

var cpmPushCmd = &cobra.Command{
	Use:   "push [path]",
	Short: "Push code to remote server",
	Long: `Push current directory to remote server.

If the current directory is linked (via 'cpm link'), the path argument is optional.

Flags:
  --dry-run, -n   Preview what would be transferred

Examples:
  anime cpm push                        # Push to linked path or ~/cpm/anime/
  anime cpm push myproject              # Push to ~/cpm/anime/myproject
  anime cpm push org/project            # Push to ~/cpm/anime/org/project
  anime cpm push -n org/project         # Dry run - preview only`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmPush,
}

var cpmPullCmd = &cobra.Command{
	Use:   "pull [path]",
	Short: "Pull code from remote server",
	Long: `Pull from remote server to current directory.

If the current directory is linked (via 'cpm link'), the path argument is optional.

Flags:
  --dry-run, -n   Preview what would be transferred

Examples:
  anime cpm pull                        # Pull from linked path or ~/cpm/anime/
  anime cpm pull myproject              # Pull from ~/cpm/anime/myproject
  anime cpm pull org/project            # Pull from ~/cpm/anime/org/project
  anime cpm pull -n org/project         # Dry run - preview only`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmPull,
}

var cpmListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List repos or packages on remote server",
	Long: `List contents of remote server (flat view).

Examples:
  anime cpm list                        # List ~/cpm/anime/
  anime cpm list org                    # List ~/cpm/anime/org/
  anime cpm list --packages             # List all packages`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmList,
}

var cpmCloneCmd = &cobra.Command{
	Use:   "clone <path>",
	Short: "Clone a repo from remote into a new folder",
	Long: `Clone from remote server into a new folder in current directory.

Unlike pull which copies contents into the current directory,
clone creates a new folder with the repo name.

Flags:
  --force, -f     Overwrite existing destination folder
  --dry-run, -n   Preview what would be transferred

Examples:
  anime cpm clone myproject             # Clone ~/cpm/anime/myproject into ./myproject
  anime cpm clone org/project           # Clone ~/cpm/anime/org/project into ./project
  anime cpm clone -f org/project        # Force overwrite if ./project exists
  anime cpm clone -n org/project        # Dry run - preview only`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmClone,
}

var cpmStatusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show sync status between local and remote",
	Long: `Compare local directory with remote and show what's different.

Shows files that are:
  - Only local (would be pushed)
  - Only remote (would be pulled)
  - Modified (different between local and remote)

Examples:
  anime cpm status                      # Status for linked path
  anime cpm status org/project          # Status for specific path`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmStatus,
}

var cpmDiffCmd = &cobra.Command{
	Use:   "diff [path]",
	Short: "Show file differences between local and remote",
	Long: `Show detailed file differences between local and remote.

Uses rsync to compare and show what would change.

Examples:
  anime cpm diff                        # Diff for linked path
  anime cpm diff org/project            # Diff for specific path`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmDiff,
}

var cpmDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a repo from remote server",
	Long: `Delete a repository from the remote server.

WARNING: This is destructive and cannot be undone!

Examples:
  anime cpm delete myproject            # Delete ~/cpm/anime/myproject
  anime cpm delete org/project          # Delete ~/cpm/anime/org/project`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmDelete,
}

var cpmRenameCmd = &cobra.Command{
	Use:   "rename <old-path> <new-path>",
	Short: "Rename or move a repo on remote server",
	Long: `Rename or move a repository on the remote server.

Examples:
  anime cpm rename oldname newname      # Rename repo
  anime cpm rename proj org/proj        # Move repo to org folder`,
	Args: cobra.ExactArgs(2),
	RunE: runCpmRename,
}

var cpmTreeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Show tree view of remote repos",
	Long: `Show a recursive tree view of remote repositories.

Examples:
  anime cpm tree                        # Tree of ~/cpm/anime/
  anime cpm tree org                    # Tree of ~/cpm/anime/org/`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmTree,
}

var cpmInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize current directory as a cpm repo and push",
	Long: `Initialize the current directory as a CPM repository.

This will:
1. Link the current directory to the remote path
2. Push all contents to the remote server

Examples:
  anime cpm init myproject              # Init as ~/cpm/anime/myproject
  anime cpm init org/project            # Init as ~/cpm/anime/org/project`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmInit,
}

var cpmSyncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Bidirectional sync between local and remote",
	Long: `Perform bidirectional synchronization.

This will:
1. Pull newer files from remote
2. Push newer local files to remote

Uses file modification times to determine which version is newer.

Flags:
  --dry-run, -n   Preview what would be transferred

Examples:
  anime cpm sync                        # Sync linked path
  anime cpm sync org/project            # Sync specific path
  anime cpm sync -n org/project         # Dry run - preview only`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmSync,
}

var cpmHistoryCmd = &cobra.Command{
	Use:   "history <path>",
	Short: "Show push history for a repo",
	Long: `Show when a repository was last pushed.

Displays timestamps of recent push operations.

Examples:
  anime cpm history myproject           # History for myproject
  anime cpm history org/project         # History for org/project`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmHistory,
}

var cpmLinkCmd = &cobra.Command{
	Use:   "link <path>",
	Short: "Link current directory to a remote path",
	Long: `Link the current directory to a remote CPM path.

Once linked, you can run push/pull/sync without specifying the path.

Examples:
  anime cpm link myproject              # Link to ~/cpm/anime/myproject
  anime cpm link org/project            # Link to ~/cpm/anime/org/project

After linking:
  anime cpm push                        # Pushes to linked path
  anime cpm pull                        # Pulls from linked path`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmLink,
}

// Package Management Commands

var cpmPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish current directory as a package",
	Long: `Publish the current directory as a versioned package.

Requires a cpm.json file with at least name and version fields.

Example cpm.json:
  {
    "name": "mypackage",
    "version": "1.0.0",
    "description": "My awesome package",
    "author": "Your Name"
  }

Flags:
  --dry-run, -n   Preview what would be published

Examples:
  anime cpm publish                     # Publish current directory`,
	RunE: runCpmPublish,
}

var cpmRepublishCmd = &cobra.Command{
	Use:   "republish",
	Short: "Re-publish package with same version (update in place)",
	Long: `Re-publish the current package version, replacing existing files.

This is useful for fixing issues without bumping the version number.
Requires a cpm.json file.

Flags:
  --dry-run, -n   Preview what would be published

Examples:
  anime cpm republish                   # Update current version in place`,
	RunE: runCpmRepublish,
}

var cpmInstallCmd = &cobra.Command{
	Use:   "install <package[@version]>",
	Short: "Install a package",
	Long: `Install a package from the package registry.

Packages are installed to ./cpm_modules by default,
or to ~/.cpm/packages with --global.

Flags:
  --global, -g    Install globally to ~/.cpm/packages
  --force, -f     Overwrite existing installation

Examples:
  anime cpm install mypackage           # Install latest version
  anime cpm install mypackage@1.0.0     # Install specific version
  anime cpm install -g mypackage        # Install globally`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmInstall,
}

var cpmUninstallCmd = &cobra.Command{
	Use:   "uninstall <package>",
	Short: "Uninstall a package",
	Long: `Remove an installed package.

Examples:
  anime cpm uninstall mypackage         # Uninstall local package
  anime cpm uninstall -g mypackage      # Uninstall global package`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmUninstall,
}

var cpmSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for packages",
	Long: `Search for packages by name or keyword.

Examples:
  anime cpm search utils                # Search for 'utils'
  anime cpm search json parser          # Search for 'json parser'`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCpmSearch,
}

var cpmInfoCmd = &cobra.Command{
	Use:   "info <package>",
	Short: "Show package information",
	Long: `Display detailed information about a package.

Examples:
  anime cpm info mypackage              # Show info for mypackage
  anime cpm info mypackage@1.0.0        # Show info for specific version`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmInfo,
}

var cpmVersionsCmd = &cobra.Command{
	Use:   "versions <package>",
	Short: "List available versions of a package",
	Long: `Show all published versions of a package.

Examples:
  anime cpm versions mypackage          # List all versions`,
	Args: cobra.ExactArgs(1),
	RunE: runCpmVersions,
}

var cpmUpdateCmd = &cobra.Command{
	Use:   "update [package]",
	Short: "Update installed packages",
	Long: `Update installed packages to their latest versions.

If no package is specified, updates all installed packages.

Examples:
  anime cpm update                      # Update all packages
  anime cpm update mypackage            # Update specific package
  anime cpm update -g                   # Update global packages`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCpmUpdate,
}

var cpmListPackagesFlag bool

func init() {
	// Add global flags to cpm command
	cpmCmd.PersistentFlags().StringVarP(&cpmServer, "server", "s", "", "Override default server (default: alice)")
	cpmCmd.PersistentFlags().BoolVarP(&cpmDryRun, "dry-run", "n", false, "Preview what would be transferred")

	// Add force flag to clone and install
	cpmCloneCmd.Flags().BoolVarP(&cpmForce, "force", "f", false, "Overwrite existing destination folder")
	cpmInstallCmd.Flags().BoolVarP(&cpmForce, "force", "f", false, "Overwrite existing installation")

	// Add global flag to install/uninstall/update
	cpmInstallCmd.Flags().BoolVarP(&cpmGlobal, "global", "g", false, "Install globally to ~/.cpm/packages")
	cpmUninstallCmd.Flags().BoolVarP(&cpmGlobal, "global", "g", false, "Uninstall global package")
	cpmUpdateCmd.Flags().BoolVarP(&cpmGlobal, "global", "g", false, "Update global packages")

	// Add packages flag to list
	cpmListCmd.Flags().BoolVarP(&cpmListPackagesFlag, "packages", "p", false, "List packages instead of repos")

	// Add all subcommands - Source Control
	cpmCmd.AddCommand(cpmPushCmd)
	cpmCmd.AddCommand(cpmPullCmd)
	cpmCmd.AddCommand(cpmCloneCmd)
	cpmCmd.AddCommand(cpmStatusCmd)
	cpmCmd.AddCommand(cpmDiffCmd)
	cpmCmd.AddCommand(cpmDeleteCmd)
	cpmCmd.AddCommand(cpmRenameCmd)
	cpmCmd.AddCommand(cpmTreeCmd)
	cpmCmd.AddCommand(cpmInitCmd)
	cpmCmd.AddCommand(cpmSyncCmd)
	cpmCmd.AddCommand(cpmHistoryCmd)
	cpmCmd.AddCommand(cpmLinkCmd)
	cpmCmd.AddCommand(cpmListCmd)

	// Add all subcommands - Package Management
	cpmCmd.AddCommand(cpmPublishCmd)
	cpmCmd.AddCommand(cpmRepublishCmd)
	cpmCmd.AddCommand(cpmInstallCmd)
	cpmCmd.AddCommand(cpmUninstallCmd)
	cpmCmd.AddCommand(cpmSearchCmd)
	cpmCmd.AddCommand(cpmInfoCmd)
	cpmCmd.AddCommand(cpmVersionsCmd)
	cpmCmd.AddCommand(cpmUpdateCmd)

	rootCmd.AddCommand(cpmCmd)
}

func getEffectiveServer() string {
	if cpmServer != "" {
		return cpmServer
	}
	return cpmDefaultServer
}

func getCpmTarget() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	return resolveSSHTarget(cfg, getEffectiveServer())
}

// getLinkedPath reads the linked remote path from .cpm-link file
func getLinkedPath() string {
	data, err := os.ReadFile(cpmLinkFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// resolveRemotePath returns the remote path, checking link file if no arg provided
func resolveRemotePath(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return getLinkedPath()
}

// isAuthError checks if an SSH error is an authentication/permission error
func isAuthError(output string, err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error() + " " + output
	return strings.Contains(errStr, "Permission denied") ||
		strings.Contains(errStr, "publickey") ||
		strings.Contains(errStr, "No more authentication methods") ||
		strings.Contains(errStr, "Host key verification failed")
}

// getRsyncExcludes returns common exclude arguments for rsync
func getRsyncExcludes() []string {
	return []string{
		"--exclude", ".git",
		"--exclude", "node_modules",
		"--exclude", "cpm_modules",
		"--exclude", "__pycache__",
		"--exclude", "*.pyc",
		"--exclude", ".env",
		"--exclude", "venv",
		"--exclude", ".venv",
		"--exclude", cpmLinkFile,
		"--exclude", cpmInstalledFile,
	}
}

// recordHistory appends a push record to the remote history file
func recordHistory(target, remotePath, keyPath string) error {
	historyPath := filepath.Join(remotePath, cpmHistoryFile)
	timestamp := time.Now().Format(time.RFC3339)
	hostname, _ := os.Hostname()
	entry := fmt.Sprintf("%s|%s|push", timestamp, hostname)

	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("echo '%s' >> %s", entry, historyPath),
	)
	return sshCmd.Run()
}

// loadPackageFile reads cpm.json from current directory
func loadPackageFile() (*CpmPackage, error) {
	data, err := os.ReadFile(cpmPackageFile)
	if err != nil {
		return nil, fmt.Errorf("no %s found - run 'cpm init-package' first or create one manually", cpmPackageFile)
	}

	var pkg CpmPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", cpmPackageFile, err)
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("%s missing required field: name", cpmPackageFile)
	}
	if pkg.Version == "" {
		return nil, fmt.Errorf("%s missing required field: version", cpmPackageFile)
	}

	return &pkg, nil
}

// loadInstalledFile reads the installed packages tracking file
func loadInstalledFile(path string) (*CpmInstalled, error) {
	filePath := filepath.Join(path, cpmInstalledFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &CpmInstalled{Packages: make(map[string]InstalledPackage)}, nil
	}

	var installed CpmInstalled
	if err := json.Unmarshal(data, &installed); err != nil {
		return &CpmInstalled{Packages: make(map[string]InstalledPackage)}, nil
	}

	if installed.Packages == nil {
		installed.Packages = make(map[string]InstalledPackage)
	}

	return &installed, nil
}

// saveInstalledFile writes the installed packages tracking file
func saveInstalledFile(path string, installed *CpmInstalled) error {
	filePath := filepath.Join(path, cpmInstalledFile)
	data, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// getInstallPath returns the installation path for packages
func getInstallPath(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cpm", "packages"), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "cpm_modules"), nil
}

// parsePackageSpec parses package@version into name and version
func parsePackageSpec(spec string) (name, version string) {
	parts := strings.SplitN(spec, "@", 2)
	name = parts[0]
	if len(parts) == 2 {
		version = parts[1]
	}
	return
}

func runCpmPush(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePath(args)

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := cpmBasePath
	if remotePath != "" {
		fullRemotePath = filepath.Join(cpmBasePath, remotePath)
	}

	server := getEffectiveServer()

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("📤 Push (DRY RUN)..."))
	} else {
		fmt.Println(theme.InfoStyle.Render("📤 Pushing..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("To:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	if !cpmDryRun {
		// Create remote directory
		mkdirCmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("mkdir -p %s", fullRemotePath),
		)
		if output, err := mkdirCmd.CombinedOutput(); err != nil {
			if isAuthError(string(output), err) {
				return RequestAccessForServer(server)
			}
			return fmt.Errorf("failed to create remote directory: %w\n%s", err, string(output))
		}
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cpmDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Record history (only if not dry run)
	if !cpmDryRun {
		recordHistory(target, fullRemotePath, keyPath)
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ Push complete"))
	}
	fmt.Println()

	return nil
}

func runCpmPull(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePath(args)

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := cpmBasePath
	if remotePath != "" {
		fullRemotePath = filepath.Join(cpmBasePath, remotePath)
	}

	server := getEffectiveServer()

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("📥 Pull (DRY RUN)..."))
	} else {
		fmt.Println(theme.InfoStyle.Render("📥 Pulling..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Check for local modifications that might conflict
	entries, _ := os.ReadDir(cwd)
	if len(entries) > 0 && !cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ⚠ Local directory not empty, files may be overwritten"))
		fmt.Println()
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cpmDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	output, err := rsyncCmd.CombinedOutput()
	fmt.Print(string(output))

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		if strings.Contains(err.Error(), "exit status") {
			return fmt.Errorf("pull failed - remote path may not exist: %s:%s", server, fullRemotePath)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ Pull complete"))
	}
	fmt.Println()

	return nil
}

func runCpmClone(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	folderName := filepath.Base(remotePath)
	server := getEffectiveServer()

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("📦 Clone (DRY RUN)..."))
	} else {
		fmt.Println(theme.InfoStyle.Render("📦 Cloning..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Printf("  %s ./%s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(folderName))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	localPath := filepath.Join(cwd, folderName)

	// Check if destination already exists
	if _, err := os.Stat(localPath); err == nil {
		if cpmForce {
			fmt.Println(theme.WarningStyle.Render("  ⚠ Overwriting existing folder"))
			fmt.Println()
			if !cpmDryRun {
				os.RemoveAll(localPath)
			}
		} else {
			return fmt.Errorf("destination folder already exists: %s (use --force to overwrite)", folderName)
		}
	}

	if !cpmDryRun {
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return fmt.Errorf("failed to create destination folder: %w", err)
		}
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cpmDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", localPath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	output, err := rsyncCmd.CombinedOutput()
	fmt.Print(string(output))

	if err != nil {
		if !cpmDryRun {
			os.RemoveAll(localPath)
		}
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		if strings.Contains(err.Error(), "exit status") {
			return fmt.Errorf("clone failed - remote path may not exist: %s:%s", server, fullRemotePath)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Create link file in cloned directory
	if !cpmDryRun {
		linkPath := filepath.Join(localPath, cpmLinkFile)
		os.WriteFile(linkPath, []byte(remotePath), 0644)
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ Clone complete"))
	}
	fmt.Println()

	return nil
}

func runCpmStatus(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePath(args)
	if remotePath == "" {
		return fmt.Errorf("no path specified and no linked path found (use 'cpm link' first)")
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📊 Status..."))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)

	// Check what would be pushed (local -> remote)
	fmt.Println(theme.DimTextStyle.Render("  Files that would be pushed (local → remote):"))
	pushArgs := []string{"-avzn", "--out-format", "    %n"}
	pushArgs = append(pushArgs, getRsyncExcludes()...)
	pushArgs = append(pushArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")
	pushCmd := exec.Command("rsync", pushArgs...)
	pushOutput, _ := pushCmd.CombinedOutput()
	pushLines := filterRsyncOutput(string(pushOutput))
	if len(pushLines) == 0 {
		fmt.Println(theme.DimTextStyle.Render("    (none)"))
	} else {
		for _, line := range pushLines {
			fmt.Println(theme.InfoStyle.Render("    → " + line))
		}
	}
	fmt.Println()

	// Check what would be pulled (remote -> local)
	fmt.Println(theme.DimTextStyle.Render("  Files that would be pulled (remote → local):"))
	pullArgs := []string{"-avzn", "--out-format", "    %n"}
	pullArgs = append(pullArgs, getRsyncExcludes()...)
	pullArgs = append(pullArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")
	pullCmd := exec.Command("rsync", pullArgs...)
	pullOutput, _ := pullCmd.CombinedOutput()
	pullLines := filterRsyncOutput(string(pullOutput))
	if len(pullLines) == 0 {
		fmt.Println(theme.DimTextStyle.Render("    (none)"))
	} else {
		for _, line := range pullLines {
			fmt.Println(theme.WarningStyle.Render("    ← " + line))
		}
	}
	fmt.Println()

	if len(pushLines) == 0 && len(pullLines) == 0 {
		fmt.Println(theme.SuccessStyle.Render("  ✓ In sync"))
	} else {
		fmt.Printf("  %s %d to push, %d to pull\n",
			theme.WarningStyle.Render("⚠"),
			len(pushLines), len(pullLines))
	}
	fmt.Println()

	return nil
}

// filterRsyncOutput filters rsync dry-run output to only show actual file names
func filterRsyncOutput(output string) []string {
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines, summary lines, and directory-only entries
		if line == "" || strings.HasPrefix(line, "sending") || strings.HasPrefix(line, "receiving") ||
			strings.HasPrefix(line, "total") || strings.HasPrefix(line, "sent") ||
			strings.HasSuffix(line, "/") || line == "." || line == "./" {
			continue
		}
		files = append(files, line)
	}
	return files
}

func runCpmDiff(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePath(args)
	if remotePath == "" {
		return fmt.Errorf("no path specified and no linked path found (use 'cpm link' first)")
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Diff..."))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)

	// Use rsync with itemize-changes to show detailed diff
	rsyncArgs := []string{"-avzin", "--itemize-changes"}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	output, err := rsyncCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	// Parse and display the output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	hasChanges := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ">") {
			// File would be sent
			fmt.Printf("  %s %s\n", theme.InfoStyle.Render("→ PUSH"), strings.TrimPrefix(line, ">f+++++++++"))
			hasChanges = true
		} else if strings.HasPrefix(line, "<") {
			// File would be received (need reverse check)
			hasChanges = true
		} else if strings.HasPrefix(line, "*deleting") {
			fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("✗ DELETE"), strings.TrimPrefix(line, "*deleting   "))
			hasChanges = true
		}
	}

	if !hasChanges {
		fmt.Println(theme.SuccessStyle.Render("  ✓ No differences"))
	}
	fmt.Println()

	return nil
}

func runCpmDelete(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("🗑️  Delete..."))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(server), theme.ErrorStyle.Render(fullRemotePath))
	fmt.Println()

	// Confirm deletion
	fmt.Print(theme.WarningStyle.Render("  ⚠ This will permanently delete the remote repo. Type 'yes' to confirm: "))
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Cancelled"))
		fmt.Println()
		return nil
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("rm -rf %s", fullRemotePath),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		return fmt.Errorf("delete failed: %w\n%s", err, string(output))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  ✓ Deleted"))
	fmt.Println()

	return nil
}

func runCpmRename(cmd *cobra.Command, args []string) error {
	oldPath := args[0]
	newPath := args[1]

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullOldPath := filepath.Join(cpmBasePath, oldPath)
	fullNewPath := filepath.Join(cpmBasePath, newPath)
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📝 Rename..."))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullOldPath))
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("To:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullNewPath))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Create parent directory if needed, then move
	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("mkdir -p $(dirname %s) && mv %s %s", fullNewPath, fullOldPath, fullNewPath),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		return fmt.Errorf("rename failed: %w\n%s", err, string(output))
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ Renamed"))
	fmt.Println()

	return nil
}

func runCpmTree(cmd *cobra.Command, args []string) error {
	remotePath := ""
	if len(args) == 1 {
		remotePath = args[0]
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := cpmBasePath
	if remotePath != "" {
		fullRemotePath = filepath.Join(cpmBasePath, remotePath)
	}
	server := getEffectiveServer()

	fmt.Println()
	fmt.Printf("  %s:%s\n", theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Try tree command, fall back to find if not available
	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("tree -L 3 %s 2>/dev/null || find %s -maxdepth 3 -type d 2>/dev/null | head -50", fullRemotePath, fullRemotePath),
	)
	output, err := sshCmd.CombinedOutput()
	fmt.Print(string(output))

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	fmt.Println()

	return nil
}

func runCpmInit(cmd *cobra.Command, args []string) error {
	remotePath := args[0]
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Initializing..."))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.InfoStyle.Render(remotePath))
	fmt.Printf("  %s %s:~/cpm/anime/%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(server), remotePath)
	fmt.Println()

	// First, create the link
	if err := os.WriteFile(cpmLinkFile, []byte(remotePath), 0644); err != nil {
		return fmt.Errorf("failed to create link file: %w", err)
	}
	fmt.Println(theme.DimTextStyle.Render("  ✓ Created .cpm-link"))

	// Then push
	fmt.Println()
	return runCpmPush(cmd, []string{remotePath})
}

func runCpmSync(cmd *cobra.Command, args []string) error {
	remotePath := resolveRemotePath(args)
	if remotePath == "" {
		return fmt.Errorf("no path specified and no linked path found (use 'cpm link' first)")
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	server := getEffectiveServer()

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("🔄 Sync (DRY RUN)..."))
	} else {
		fmt.Println(theme.InfoStyle.Render("🔄 Syncing..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)

	// Step 1: Pull newer files from remote (--update flag keeps newer local files)
	fmt.Println(theme.DimTextStyle.Render("  Step 1: Pulling newer remote files..."))
	pullArgs := []string{"-avz", "--progress", "--update"}
	if cpmDryRun {
		pullArgs = append(pullArgs, "--dry-run")
	}
	pullArgs = append(pullArgs, getRsyncExcludes()...)
	pullArgs = append(pullArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")

	pullCmd := exec.Command("rsync", pullArgs...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("sync pull failed: %w", err)
	}

	fmt.Println()

	// Step 2: Push newer local files to remote
	fmt.Println(theme.DimTextStyle.Render("  Step 2: Pushing newer local files..."))
	pushArgs := []string{"-avz", "--progress", "--update"}
	if cpmDryRun {
		pushArgs = append(pushArgs, "--dry-run")
	}
	pushArgs = append(pushArgs, getRsyncExcludes()...)
	pushArgs = append(pushArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")

	pushCmd := exec.Command("rsync", pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("sync push failed: %w", err)
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ Sync complete"))
	}
	fmt.Println()

	return nil
}

func runCpmHistory(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	fullRemotePath := filepath.Join(cpmBasePath, remotePath)
	historyPath := filepath.Join(fullRemotePath, cpmHistoryFile)
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📜 History..."))
	fmt.Println()
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Repo:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("cat %s 2>/dev/null | tail -20 || echo '(no history)'", historyPath),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	// Parse and display history
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "(no history)" {
		fmt.Println(theme.DimTextStyle.Render("  No push history found"))
	} else {
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				timestamp, host, action := parts[0], parts[1], parts[2]
				t, err := time.Parse(time.RFC3339, timestamp)
				if err == nil {
					timestamp = t.Format("2006-01-02 15:04:05")
				}
				fmt.Printf("  %s  %s from %s\n",
					theme.DimTextStyle.Render(timestamp),
					theme.InfoStyle.Render(action),
					theme.HighlightStyle.Render(host))
			}
		}
	}
	fmt.Println()

	return nil
}

func runCpmLink(cmd *cobra.Command, args []string) error {
	remotePath := args[0]
	server := getEffectiveServer()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔗 Linking..."))
	fmt.Println()
	fmt.Printf("  %s %s:~/cpm/anime/%s\n", theme.DimTextStyle.Render("Remote:"), theme.HighlightStyle.Render(server), remotePath)
	fmt.Println()

	// Write link file
	if err := os.WriteFile(cpmLinkFile, []byte(remotePath), 0644); err != nil {
		return fmt.Errorf("failed to create link file: %w", err)
	}

	// Also add to .gitignore if it exists and doesn't already contain it
	if _, err := os.Stat(".gitignore"); err == nil {
		gitignore, _ := os.ReadFile(".gitignore")
		if !strings.Contains(string(gitignore), cpmLinkFile) {
			f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString("\n" + cpmLinkFile + "\n")
				f.Close()
				fmt.Println(theme.DimTextStyle.Render("  Added .cpm-link to .gitignore"))
			}
		}
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ Linked"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  You can now use 'cpm push', 'cpm pull', 'cpm sync' without arguments"))
	fmt.Println()

	return nil
}

func runCpmList(cmd *cobra.Command, args []string) error {
	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	var fullRemotePath string

	if cpmListPackagesFlag {
		// List packages
		fullRemotePath = cpmPackagesPath
	} else {
		// List repos
		remotePath := ""
		if len(args) == 1 {
			remotePath = args[0]
		}
		fullRemotePath = cpmBasePath
		if remotePath != "" {
			fullRemotePath = filepath.Join(cpmBasePath, remotePath)
		}
	}

	fmt.Println()
	fmt.Printf("  %s:%s\n", theme.HighlightStyle.Render(server), theme.InfoStyle.Render(fullRemotePath))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -la %s/ 2>/dev/null || echo '(not found)'", fullRemotePath),
	)
	output, err := sshCmd.CombinedOutput()
	fmt.Print(string(output))

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		return fmt.Errorf("failed to list: %w", err)
	}

	fmt.Println()

	return nil
}

// Package Management Functions

func runCpmPublish(cmd *cobra.Command, args []string) error {
	pkg, err := loadPackageFile()
	if err != nil {
		return err
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	packagePath := filepath.Join(cpmPackagesPath, pkg.Name, pkg.Version)

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("📦 Publish (DRY RUN)..."))
	} else {
		fmt.Println(theme.InfoStyle.Render("📦 Publishing..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkg.Name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.InfoStyle.Render(pkg.Version))
	if pkg.Description != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Description:"), pkg.Description)
	}
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("To:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(packagePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Check if version already exists
	checkCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("test -d %s && echo 'exists'", packagePath),
	)
	checkOutput, _ := checkCmd.CombinedOutput()
	if strings.TrimSpace(string(checkOutput)) == "exists" {
		return fmt.Errorf("version %s already published - use 'cpm republish' to update or bump version", pkg.Version)
	}

	if !cpmDryRun {
		// Create remote directory
		mkdirCmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("mkdir -p %s", packagePath),
		)
		if output, err := mkdirCmd.CombinedOutput(); err != nil {
			if isAuthError(string(output), err) {
				return RequestAccessForServer(server)
			}
			return fmt.Errorf("failed to create remote directory: %w\n%s", err, string(output))
		}
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cpmDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+packagePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Update latest symlink
	if !cpmDryRun {
		latestPath := filepath.Join(cpmPackagesPath, pkg.Name, "latest")
		symlinkCmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("rm -f %s && ln -s %s %s", latestPath, pkg.Version, latestPath),
		)
		symlinkCmd.Run() // Ignore errors for symlink
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Published %s@%s", pkg.Name, pkg.Version)))
	}
	fmt.Println()

	return nil
}

func runCpmRepublish(cmd *cobra.Command, args []string) error {
	pkg, err := loadPackageFile()
	if err != nil {
		return err
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	packagePath := filepath.Join(cpmPackagesPath, pkg.Name, pkg.Version)

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("📦 Republish (DRY RUN)..."))
	} else {
		fmt.Println(theme.WarningStyle.Render("📦 Republishing (updating in place)..."))
	}
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkg.Name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.InfoStyle.Render(pkg.Version))
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("To:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(packagePath))
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	if !cpmDryRun {
		// Create remote directory (in case it doesn't exist)
		mkdirCmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("mkdir -p %s", packagePath),
		)
		if output, err := mkdirCmd.CombinedOutput(); err != nil {
			if isAuthError(string(output), err) {
				return RequestAccessForServer(server)
			}
			return fmt.Errorf("failed to create remote directory: %w\n%s", err, string(output))
		}
	}

	// Rsync with delete to ensure clean update
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress", "--delete"}
	if cpmDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+packagePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Println()
	if cpmDryRun {
		fmt.Println(theme.WarningStyle.Render("  ✓ Dry run complete (no changes made)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Republished %s@%s", pkg.Name, pkg.Version)))
	}
	fmt.Println()

	return nil
}

func runCpmInstall(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]
	pkgName, pkgVersion := parsePackageSpec(pkgSpec)

	if pkgVersion == "" {
		pkgVersion = "latest"
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	packagePath := filepath.Join(cpmPackagesPath, pkgName, pkgVersion)

	installPath, err := getInstallPath(cpmGlobal)
	if err != nil {
		return err
	}

	localPath := filepath.Join(installPath, pkgName)

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📥 Installing..."))
	fmt.Println()
	fmt.Printf("  %s %s@%s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkgName), theme.InfoStyle.Render(pkgVersion))
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("From:"), theme.HighlightStyle.Render(server), theme.InfoStyle.Render(packagePath))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(localPath))
	if cpmGlobal {
		fmt.Printf("  %s global\n", theme.DimTextStyle.Render("Scope:"))
	}
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Check if package exists
	checkCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("test -d %s && echo 'exists'", packagePath),
	)
	checkOutput, _ := checkCmd.CombinedOutput()
	if strings.TrimSpace(string(checkOutput)) != "exists" {
		return fmt.Errorf("package not found: %s@%s", pkgName, pkgVersion)
	}

	// Check if already installed
	if _, err := os.Stat(localPath); err == nil && !cpmForce {
		return fmt.Errorf("package already installed at %s (use --force to overwrite)", localPath)
	}

	// Create install directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", keyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+packagePath+"/", localPath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	output, err := rsyncCmd.CombinedOutput()
	fmt.Print(string(output))

	if err != nil {
		os.RemoveAll(localPath)
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
		return fmt.Errorf("install failed: %w", err)
	}

	// Read actual version from installed package
	actualVersion := pkgVersion
	if pkgVersion == "latest" {
		installedPkg, err := loadPackageFileFrom(localPath)
		if err == nil {
			actualVersion = installedPkg.Version
		}
	}

	// Track installation
	installed, _ := loadInstalledFile(installPath)
	installed.Packages[pkgName] = InstalledPackage{
		Version:     actualVersion,
		InstalledAt: time.Now().Format(time.RFC3339),
		Path:        localPath,
		Global:      cpmGlobal,
	}
	saveInstalledFile(installPath, installed)

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Installed %s@%s", pkgName, actualVersion)))
	fmt.Println()

	return nil
}

func loadPackageFileFrom(path string) (*CpmPackage, error) {
	data, err := os.ReadFile(filepath.Join(path, cpmPackageFile))
	if err != nil {
		return nil, err
	}

	var pkg CpmPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func runCpmUninstall(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	installPath, err := getInstallPath(cpmGlobal)
	if err != nil {
		return err
	}

	localPath := filepath.Join(installPath, pkgName)

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🗑️  Uninstalling..."))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkgName))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("From:"), theme.InfoStyle.Render(localPath))
	fmt.Println()

	// Check if installed
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("package not installed: %s", pkgName)
	}

	// Remove package
	if err := os.RemoveAll(localPath); err != nil {
		return fmt.Errorf("failed to uninstall: %w", err)
	}

	// Update tracking
	installed, _ := loadInstalledFile(installPath)
	delete(installed.Packages, pkgName)
	saveInstalledFile(installPath, installed)

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Uninstalled %s", pkgName)))
	fmt.Println()

	return nil
}

func runCpmSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()

	fmt.Println()
	fmt.Printf("  %s \"%s\"\n", theme.DimTextStyle.Render("Searching for:"), theme.InfoStyle.Render(query))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// List all packages
	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -1 %s 2>/dev/null || echo ''", cpmPackagesPath),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Filter by query
	queryLower := strings.ToLower(query)
	var matches []string
	for _, pkg := range packages {
		if pkg != "" && strings.Contains(strings.ToLower(pkg), queryLower) {
			matches = append(matches, pkg)
		}
	}

	if len(matches) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No packages found"))
	} else {
		fmt.Printf("  %s\n\n", theme.DimTextStyle.Render(fmt.Sprintf("Found %d package(s):", len(matches))))
		for _, pkg := range matches {
			// Try to get latest version info
			infoCmd := exec.Command("ssh",
				"-i", keyPath,
				"-o", "IdentitiesOnly=yes",
				"-o", "StrictHostKeyChecking=accept-new",
				target,
				fmt.Sprintf("cat %s/%s/latest/%s 2>/dev/null || echo '{}'", cpmPackagesPath, pkg, cpmPackageFile),
			)
			infoOutput, _ := infoCmd.CombinedOutput()

			var pkgInfo CpmPackage
			json.Unmarshal(infoOutput, &pkgInfo)

			fmt.Printf("  %s", theme.HighlightStyle.Render(pkg))
			if pkgInfo.Version != "" {
				fmt.Printf("@%s", theme.InfoStyle.Render(pkgInfo.Version))
			}
			if pkgInfo.Description != "" {
				fmt.Printf(" - %s", theme.DimTextStyle.Render(pkgInfo.Description))
			}
			fmt.Println()
		}
	}
	fmt.Println()

	return nil
}

func runCpmInfo(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]
	pkgName, pkgVersion := parsePackageSpec(pkgSpec)

	if pkgVersion == "" {
		pkgVersion = "latest"
	}

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	packagePath := filepath.Join(cpmPackagesPath, pkgName, pkgVersion)

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Package Info..."))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// Get package.json
	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("cat %s/%s 2>/dev/null || echo 'NOT_FOUND'", packagePath, cpmPackageFile),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "NOT_FOUND" {
		return fmt.Errorf("package not found: %s@%s", pkgName, pkgVersion)
	}

	var pkg CpmPackage
	if err := json.Unmarshal(output, &pkg); err != nil {
		return fmt.Errorf("invalid package metadata: %w", err)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(pkg.Name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.InfoStyle.Render(pkg.Version))
	if pkg.Description != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Description:"), pkg.Description)
	}
	if pkg.Author != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Author:"), pkg.Author)
	}
	if pkg.License != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("License:"), pkg.License)
	}
	if len(pkg.Keywords) > 0 {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Keywords:"), strings.Join(pkg.Keywords, ", "))
	}
	if pkg.Repository != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repository:"), pkg.Repository)
	}
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Location:"), theme.HighlightStyle.Render(server), packagePath)
	fmt.Println()

	return nil
}

func runCpmVersions(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	server := getEffectiveServer()
	packagePath := filepath.Join(cpmPackagesPath, pkgName)

	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkgName))
	fmt.Println()

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	// List versions
	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -1 %s 2>/dev/null | grep -v latest || echo 'NOT_FOUND'", packagePath),
	)
	output, err := sshCmd.CombinedOutput()

	if err != nil {
		if isAuthError(string(output), err) {
			return RequestAccessForServer(server)
		}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "NOT_FOUND" || outputStr == "" {
		return fmt.Errorf("package not found: %s", pkgName)
	}

	versions := strings.Split(outputStr, "\n")
	// Sort versions (simple string sort, could be improved with semver)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	fmt.Println(theme.DimTextStyle.Render("  Available versions:"))
	for i, v := range versions {
		if i == 0 {
			fmt.Printf("    %s %s\n", theme.HighlightStyle.Render(v), theme.SuccessStyle.Render("(latest)"))
		} else {
			fmt.Printf("    %s\n", v)
		}
	}
	fmt.Println()

	return nil
}

func runCpmUpdate(cmd *cobra.Command, args []string) error {
	installPath, err := getInstallPath(cpmGlobal)
	if err != nil {
		return err
	}

	installed, err := loadInstalledFile(installPath)
	if err != nil {
		return err
	}

	if len(installed.Packages) == 0 {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  No packages installed"))
		fmt.Println()
		return nil
	}

	// Filter to specific package if provided
	packagesToUpdate := installed.Packages
	if len(args) == 1 {
		pkgName := args[0]
		if _, ok := installed.Packages[pkgName]; !ok {
			return fmt.Errorf("package not installed: %s", pkgName)
		}
		packagesToUpdate = map[string]InstalledPackage{pkgName: installed.Packages[pkgName]}
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔄 Updating packages..."))
	fmt.Println()

	target, err := getCpmTarget()
	if err != nil {
		return err
	}

	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return fmt.Errorf("failed to prepare SSH key: %w", err)
	}
	defer cleanup()

	updated := 0
	for pkgName, pkg := range packagesToUpdate {
		// Get latest version
		latestPath := filepath.Join(cpmPackagesPath, pkgName, "latest", cpmPackageFile)
		sshCmd := exec.Command("ssh",
			"-i", keyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("cat %s 2>/dev/null || echo '{}'", latestPath),
		)
		output, _ := sshCmd.CombinedOutput()

		var latestPkg CpmPackage
		json.Unmarshal(output, &latestPkg)

		if latestPkg.Version == "" || latestPkg.Version == pkg.Version {
			fmt.Printf("  %s %s - %s\n", theme.DimTextStyle.Render("○"), pkgName, theme.DimTextStyle.Render("up to date"))
			continue
		}

		fmt.Printf("  %s %s %s → %s\n",
			theme.InfoStyle.Render("↑"),
			theme.HighlightStyle.Render(pkgName),
			theme.DimTextStyle.Render(pkg.Version),
			theme.SuccessStyle.Render(latestPkg.Version))

		// Reinstall with latest
		cpmForce = true
		if err := runCpmInstall(cmd, []string{pkgName}); err != nil {
			fmt.Printf("    %s\n", theme.ErrorStyle.Render(err.Error()))
		} else {
			updated++
		}
	}

	fmt.Println()
	if updated > 0 {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Updated %d package(s)", updated)))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ All packages up to date"))
	}
	fmt.Println()

	return nil
}

// CpmLinkInfo represents the link configuration
type CpmLinkInfo struct {
	RemotePath string `json:"remote_path"`
	Server     string `json:"server,omitempty"`
}

// For future: could store more metadata in JSON format
func readLinkInfo() (*CpmLinkInfo, error) {
	data, err := os.ReadFile(cpmLinkFile)
	if err != nil {
		return nil, err
	}

	// Try JSON first
	var info CpmLinkInfo
	if err := json.Unmarshal(data, &info); err != nil {
		// Fall back to plain text (just the path)
		info.RemotePath = strings.TrimSpace(string(data))
	}

	return &info, nil
}

// Validation helper for semver-like versions
func isValidVersion(v string) bool {
	// Simple version validation - allows x.y.z, x.y.z-beta, etc.
	match, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[\w.]+)?$`, v)
	return match
}
