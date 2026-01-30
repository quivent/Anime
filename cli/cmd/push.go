package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	pushArch           string
	pushIncludeSource  bool
	pushRunConfig      bool
	pushFast           bool
	pushSkipClaude     bool
	pushBinaryOnly     bool
	pushIncremental    bool
	pushInstallAliases bool
	pushDryRun         bool
	pushDelete         bool
	pushExclude        []string
)

var pushCmd = &cobra.Command{
	Use:   "push [source] [server] [dest]",
	Short: "Push anime CLI or files to remote server",
	Long: `Push anime CLI binary or files/folders to a remote server.

MODES:
  anime push [server]                    Push anime CLI to server
  anime push <source> <server>           Push files to server:~/
  anime push <source> <server> <dest>    Push files to server:<dest>

SERVER FORMATS:
  - user@IP          (e.g., ubuntu@192.168.1.100)
  - IP               (defaults to ubuntu@IP)
  - alias            (from anime config or .ssh/config)
  - (default: lambda if no server specified)

EXAMPLES:
  # Push anime CLI
  anime push                                  # Push CLI to 'lambda'
  anime push captain                          # Push CLI to captain
  anime push captain --config                 # Push CLI and run config

  # Push files/folders
  anime push ./myapp captain                  # Push folder to ~/myapp
  anime push ./data captain ~/backup          # Push to ~/backup
  anime push . captain ~/project              # Push current dir

CLI PUSH OPTIONS:
  --binary, -b      Binary only via scp (fastest)
  --incremental, -i Rsync delta transfer
  --fast, -f        Skip source, Claude assets, verification
  --aliases, -a     Install shell aliases
  --config, -c      Run config on server after push
  --dry-run, -n     Preview without changes

FILE PUSH OPTIONS:
  --exclude, -e     Exclude pattern (can use multiple: -e build -e dist)
  --delete, -d      Delete extraneous files from dest (mirror)
  --dry-run, -n     Preview without changes

DEFAULT EXCLUDES:
  .git, node_modules, __pycache__, .venv, venv, .DS_Store,
  target (Rust), dist, build, .next, .nuxt
`,
	Args: cobra.MaximumNArgs(3),
	RunE: runPush,
}

func init() {
	// CLI push flags
	pushCmd.Flags().StringVar(&pushArch, "arch", "", "Target architecture (amd64 or arm64, auto-detected if not specified)")
	pushCmd.Flags().BoolVar(&pushIncludeSource, "source", true, "Include source code in the package")
	pushCmd.Flags().BoolVarP(&pushRunConfig, "config", "c", false, "Run config on server after push")
	pushCmd.Flags().BoolVarP(&pushFast, "fast", "f", false, "Fast mode: skip source, Claude assets, and verification")
	pushCmd.Flags().BoolVar(&pushSkipClaude, "no-claude", false, "Skip deploying Claude commands/agents")
	pushCmd.Flags().BoolVarP(&pushBinaryOnly, "binary", "b", false, "Binary only: just push the binary, force overwrite (fastest)")
	pushCmd.Flags().BoolVarP(&pushIncremental, "incremental", "i", false, "Incremental: rsync delta transfer, only sync changed bytes")
	pushCmd.Flags().BoolVarP(&pushInstallAliases, "aliases", "a", false, "Install shell aliases (codec, coder, etc.) and source bashrc")
	// Shared flags
	pushCmd.Flags().BoolVarP(&pushDryRun, "dry-run", "n", false, "Preview what would be done without making changes")
	// File push flags
	pushCmd.Flags().BoolVarP(&pushDelete, "delete", "d", false, "Delete extraneous files from dest (mirror mode)")
	pushCmd.Flags().StringArrayVarP(&pushExclude, "exclude", "e", nil, "Exclude pattern (can use multiple times)")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	// Determine push mode based on argument count
	// 0-1 args: CLI push
	// 2 args: file push to ~/
	// 3 args: file push to specified dest
	if len(args) >= 2 {
		return runFilePush(args)
	}

	// CLI push mode
	server := "lambda"
	if len(args) > 0 {
		server = args[0]
	}

	// Parse server argument
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	// Dry-run mode: show what would happen
	if pushDryRun {
		return runPushDryRun(target)
	}

	// Binary-only mode: fastest push, just the binary
	if pushBinaryOnly {
		return runBinaryOnlyPush(target)
	}

	// Incremental mode: rsync delta transfer
	if pushIncremental {
		return runIncrementalPush(target, server)
	}

	// Apply fast mode settings
	if pushFast {
		pushIncludeSource = false
		pushSkipClaude = true
	}

	if pushFast {
		fmt.Println(theme.InfoStyle.Render("⚡ Fast push to remote server"))
	} else {
		fmt.Println(theme.InfoStyle.Render("🚀 Pushing anime to remote server"))
	}
	fmt.Println()

	// Run arch detection + connection test + build in parallel
	type buildResult struct {
		binaryPath string
		version    string
		buildTime  string
		sourceDir  string
		err        error
	}

	buildChan := make(chan buildResult, 1)
	archChan := make(chan archResult, 1)
	connChan := make(chan error, 1)

	// Start all three operations in parallel
	fmt.Print(theme.DimTextStyle.Render("▶ Preparing (build + connect)... "))

	// Detect architecture (if not specified)
	go func() {
		if pushArch != "" {
			archChan <- archResult{arch: pushArch, err: nil}
			return
		}
		detectedArch, err := detectRemoteArchitecture(target)
		if err != nil {
			archChan <- archResult{arch: "amd64", err: err}
		} else {
			archChan <- archResult{arch: detectedArch, err: nil}
		}
	}()

	// Test connection
	go func() {
		connChan <- testConnection(target)
	}()

	// Wait for arch detection first (needed for build)
	archRes := <-archChan
	pushArch = archRes.arch

	// Start build after we have arch
	go func() {
		binaryPath, version, buildTime, sourceDir, err := buildLinuxBinary()
		buildChan <- buildResult{binaryPath, version, buildTime, sourceDir, err}
	}()

	// Wait for connection test
	connErr := <-connChan
	if connErr != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showConnectionSuggestions(target, connErr)
		return fmt.Errorf("connection test failed")
	}

	// Wait for build
	buildRes := <-buildChan
	if buildRes.err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showBuildSuggestions(buildRes.err)
		return fmt.Errorf("build failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	binaryPath := buildRes.binaryPath
	version := buildRes.version
	buildTime := buildRes.buildTime
	sourceDir := buildRes.sourceDir
	defer os.Remove(binaryPath) // Clean up temp binary

	// Show build info
	fmt.Println()
	fmt.Printf("  Target:  %s\n", theme.HighlightStyle.Render(target))
	fmt.Printf("  Version: %s\n", theme.HighlightStyle.Render(version))
	fmt.Printf("  Built:   %s\n", theme.HighlightStyle.Render(strings.ReplaceAll(buildTime, "_", " ")))
	fmt.Printf("  Arch:    %s\n", theme.HighlightStyle.Render("linux/"+pushArch))
	if pushIncludeSource {
		fmt.Printf("  Source:  %s\n", theme.SuccessStyle.Render("included"))
	} else {
		fmt.Printf("  Source:  %s\n", theme.DimTextStyle.Render("skipped"))
	}
	fmt.Println()

	// Step 3: Create tar.gz package
	fmt.Print(theme.DimTextStyle.Render("▶ Creating package... "))
	tarPath, err := createPackage(binaryPath, sourceDir)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showPackagingSuggestions(err)
		return fmt.Errorf("packaging failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))
	defer os.Remove(tarPath) // Clean up temp tar

	// Step 4: Rsync to remote server
	fmt.Print(theme.DimTextStyle.Render("▶ Syncing to server... "))
	if err := rsyncToServer(tarPath, target); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showRsyncSuggestions(target, err)
		return fmt.Errorf("rsync failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Step 5: Extract on remote server
	fmt.Print(theme.DimTextStyle.Render("▶ Extracting on server... "))
	if err := extractOnServer(target); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showExtractionSuggestions(target, err)
		return fmt.Errorf("extraction failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Step 6: Configure PATH on server
	fmt.Print(theme.DimTextStyle.Render("▶ Configuring PATH... "))
	if err := addToPathOnServer(target); err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓"))
	}

	// Step 6.5: Push embedded Claude commands and agents to ~/.claude/
	if !pushSkipClaude {
		fmt.Print(theme.DimTextStyle.Render("▶ Deploying Claude commands... "))
		if err := pushClaudeAssetsOnServer(target); err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
			fmt.Println()
		} else {
			fmt.Println(theme.SuccessStyle.Render("✓"))
		}
	}

	// Step 6.6: Install shell aliases if requested
	if pushInstallAliases {
		fmt.Print(theme.DimTextStyle.Render("▶ Installing shell aliases... "))
		if err := installShellAliasesOnServer(target); err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
			fmt.Println()
		} else {
			fmt.Println(theme.SuccessStyle.Render("✓"))
		}
	}

	// Step 7: Verify version on server (skip in fast mode)
	if !pushFast {
		fmt.Print(theme.DimTextStyle.Render("▶ Verifying version... "))
		if err := verifyServerVersion(target, version); err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
			fmt.Println()
		} else {
			fmt.Println(theme.SuccessStyle.Render("✓"))
		}
	}

	// Step 8: Auto-configure server in anime config
	if err := autoConfigureServer(server, target); err != nil {
		// Don't fail the push if config save fails, just warn
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠ Warning: Could not save server to config: " + err.Error()))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Push complete!"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📍 Installed to: ") + theme.HighlightStyle.Render("~/.local/bin/anime"))
	fmt.Println(theme.InfoStyle.Render("📍 PATH configured: ") + theme.HighlightStyle.Render("~/.local/bin added to shell"))
	fmt.Println(theme.InfoStyle.Render("📍 Claude commands: ") + theme.HighlightStyle.Render("~/.claude/commands/"))
	fmt.Println(theme.InfoStyle.Render("📍 Claude agents: ") + theme.HighlightStyle.Render("~/.claude/agents/"))
	fmt.Println()

	// Run config on server if requested
	if pushRunConfig {
		fmt.Println(theme.InfoStyle.Render("🔧 Running config on server..."))
		fmt.Println()

		if err := runSSHInteractive(target, "anime config"); err != nil {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("⚠️  Remote config had an issue"))
			fmt.Println(theme.DimTextStyle.Render("  You can run it manually: ") + theme.HighlightStyle.Render("ssh "+target+" anime config"))
			fmt.Println()
		}
	} else {
		fmt.Println(theme.DimTextStyle.Render("  Run on server: ") + theme.HighlightStyle.Render("ssh "+target+" anime"))
		fmt.Println(theme.DimTextStyle.Render("  Or SSH in:     ") + theme.HighlightStyle.Render("ssh "+target))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  💡 Configure server: ") + theme.HighlightStyle.Render("anime push "+server+" --config"))
	}
	fmt.Println()

	return nil
}

func parseServerTarget(server string) (string, error) {
	// If it contains @, use as-is
	if strings.Contains(server, "@") {
		return server, nil
	}

	// If it looks like an IP address or hostname, prepend ubuntu@
	if strings.Contains(server, ".") || strings.Contains(server, ":") {
		return "ubuntu@" + server, nil
	}

	// First, check anime config (servers and aliases)
	cfg, err := config.Load()
	if err == nil {
		// Check servers by name first
		if srv, err := cfg.GetServer(server); err == nil && srv != nil {
			return srv.User + "@" + srv.Host, nil
		}
		// Then check aliases
		if target := cfg.GetAlias(server); target != "" {
			// Recursively resolve the alias target
			return parseServerTarget(target)
		}
	}

	// Otherwise, try to resolve as SSH alias
	// First check if ssh -G works to resolve the alias
	sshCmd := exec.Command("ssh", "-G", server)
	output, err := sshCmd.CombinedOutput()
	if err == nil {
		// Parse output to get hostname and user
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
		if hostname != "" && user != "" {
			return user + "@" + hostname, nil
		}
		if hostname != "" {
			return "ubuntu@" + hostname, nil
		}
	}

	// If SSH alias resolution fails, assume it's a hostname and use ubuntu@
	return "ubuntu@" + server, nil
}

func buildLinuxBinary() (binaryPath, version, buildTime, sourceDir string, err error) {
	// Find the CLI source directory
	sourceDir, err = findSourceDir()
	if err != nil {
		return "", "", "", "", fmt.Errorf("could not find source directory: %w", err)
	}

	// Get version info (version is bumped by Makefile during local build)
	version = getGitVersionFromDir(sourceDir)
	buildTime = time.Now().UTC().Format("2006-01-02_15:04:05")
	commit := getGitCommitFromDir(sourceDir)

	// Build flags - include BuildDir so future pushes know where source lives
	ldflags := fmt.Sprintf("-X github.com/joshkornreich/anime/cmd.Version=%s -X github.com/joshkornreich/anime/cmd.BuildTime=%s -X github.com/joshkornreich/anime/cmd.Commit=%s -X github.com/joshkornreich/anime/cmd.BuildDir=%s",
		version, buildTime, commit, sourceDir)

	// Output path
	binaryPath = filepath.Join(os.TempDir(), "anime-linux-"+pushArch)

	// Build command - run from the source directory
	buildCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", binaryPath, ".")
	buildCmd.Dir = sourceDir
	buildCmd.Env = append(os.Environ(),
		"GOOS=linux",
		"GOARCH="+pushArch,
	)

	output, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		return "", "", "", "", fmt.Errorf("%w: %s", buildErr, string(output))
	}

	return binaryPath, version, buildTime, sourceDir, nil
}

// findSourceDir locates the anime CLI source directory
func findSourceDir() (string, error) {
	home, _ := os.UserHomeDir()

	// 1. Check if BuildDir was set at compile time (most reliable)
	if BuildDir != "" {
		if _, err := os.Stat(filepath.Join(BuildDir, "go.mod")); err == nil {
			return BuildDir, nil
		}
	}

	// 2. Check current directory (for development)
	if _, err := os.Stat("go.mod"); err == nil {
		if _, err := os.Stat("main.go"); err == nil {
			cwd, _ := os.Getwd()
			return cwd, nil
		}
	}

	// 3. Check common source locations
	sourcePaths := []string{
		filepath.Join(home, "anime", "cli"),          // Remote push location
		filepath.Join(home, ".anime", "cli", "src"),  // Legacy location
		filepath.Join(home, ".anime", "cli"),         // Alternative
		filepath.Join(home, "github", "anime", "cli"), // GitHub clone location
	}

	for _, srcDir := range sourcePaths {
		if _, err := os.Stat(filepath.Join(srcDir, "go.mod")); err == nil {
			return srcDir, nil
		}
	}

	// 4. Try to find source relative to the running binary
	execPath, err := os.Executable()
	if err == nil {
		binDir := filepath.Dir(execPath)

		// Check if we're in a dev environment (binary alongside source)
		parentDir := filepath.Dir(binDir)
		if _, err := os.Stat(filepath.Join(parentDir, "go.mod")); err == nil {
			return parentDir, nil
		}

		// Check grandparent (e.g., ~/anime/cli/bin/anime -> ~/anime/cli)
		grandparentDir := filepath.Dir(parentDir)
		if _, err := os.Stat(filepath.Join(grandparentDir, "go.mod")); err == nil {
			return grandparentDir, nil
		}
	}

	return "", fmt.Errorf("source not found - run from source directory or ensure ~/anime/cli exists")
}

// getGitVersionFromDir reads version from a specific directory
func getGitVersionFromDir(dir string) string {
	versionBytes, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(versionBytes))
}

// getGitCommitFromDir gets git commit from a specific directory
func getGitCommitFromDir(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func createPackage(binaryPath, sourceDir string) (string, error) {
	// Create tar.gz file
	tarPath := filepath.Join(os.TempDir(), "anime-package.tar.gz")

	outFile, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add binary to tar
	if err := addFileToTar(tarWriter, binaryPath, "anime"); err != nil {
		return "", fmt.Errorf("failed to add binary: %w", err)
	}

	// Add source code if requested
	if pushIncludeSource {
		if err := addSourceToTar(tarWriter, sourceDir); err != nil {
			return "", fmt.Errorf("failed to add source: %w", err)
		}
	}

	// Add embedded files if any exist
	if err := addEmbeddedFilesToTar(tarWriter); err != nil {
		// Just log the error but don't fail the push if no embedded files exist
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to add embedded files: %w", err)
		}
	}

	return tarPath, nil
}

func addFileToTar(tw *tar.Writer, srcPath, destName string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    destName,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func addSourceToTar(tw *tar.Writer, sourceDir string) error {
	// Add main source directories
	items := []string{"cmd", "internal", "main.go", "go.mod", "go.sum", "VERSION"}

	for _, item := range items {
		srcPath := filepath.Join(sourceDir, item)
		if err := addDirToTar(tw, srcPath, "src/"+item, sourceDir); err != nil {
			// Skip if directory doesn't exist
			if !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func addDirToTar(tw *tar.Writer, srcPath, destPrefix, baseDir string) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git and other hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip hidden files and build artifacts
		if strings.HasPrefix(info.Name(), ".") || info.Name() == "build" {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update name to include destination prefix
		// Use relative path from the source item, not from baseDir
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			// Single file (not directory), just use the destPrefix
			header.Name = destPrefix
		} else {
			header.Name = filepath.Join(destPrefix, relPath)
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}

func addEmbeddedFilesToTar(tw *tar.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	embeddedDir := filepath.Join(home, ".anime", "embedded")
	manifestPath := filepath.Join(embeddedDir, "manifest.json")

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return err // No embedded files
	}

	// Add the manifest file
	if err := addFileToTar(tw, manifestPath, "embedded/manifest.json"); err != nil {
		return err
	}

	// Read manifest to get all embedded files
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest struct {
		Files map[string]struct {
			StoredName string `json:"stored_name"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// Add each embedded tar.gz file
	for _, file := range manifest.Files {
		tarGzPath := filepath.Join(embeddedDir, file.StoredName)
		if err := addFileToTar(tw, tarGzPath, "embedded/"+file.StoredName); err != nil {
			return err
		}
	}

	return nil
}

func rsyncToServer(localPath, target string) error {
	// Rsync command - use compression and quiet mode for speed
	args := []string{"-az", "--compress-level=9"}
	if !pushFast {
		args = append(args, "--progress")
	}

	// Add SSH options with embedded key if available
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		defer cleanup()
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", keyPath)
		args = append(args, "-e", sshCmd)
	}

	args = append(args, localPath, target+":~/")

	rsyncCmd := exec.Command("rsync", args...)
	if !pushFast {
		rsyncCmd.Stdout = os.Stdout
		rsyncCmd.Stderr = os.Stderr
	}

	return rsyncCmd.Run()
}

func extractOnServer(target string) error {
	// SSH command to extract tar.gz on server and install to ~/.local/bin
	// Also extracts embedded files to ~/.anime/embedded/ and source to ~/.anime/cli/src
	extractCmd := `
		cd ~/ && \
		rm -rf /tmp/anime-extract 2>/dev/null || true && \
		mkdir -p /tmp/anime-extract && \
		tar -xzf anime-package.tar.gz -C /tmp/anime-extract && \
		mkdir -p ~/.local/bin && \
		mv -f /tmp/anime-extract/anime ~/.local/bin/anime && \
		chmod +x ~/.local/bin/anime && \
		if [ -d /tmp/anime-extract/src ]; then \
			mkdir -p ~/anime && \
			rm -rf ~/anime/cli 2>/dev/null || true && \
			mv /tmp/anime-extract/src ~/anime/cli; \
		fi && \
		if [ -d /tmp/anime-extract/embedded ]; then \
			mkdir -p ~/.anime/embedded && \
			cp -r /tmp/anime-extract/embedded/* ~/.anime/embedded/ 2>/dev/null || true; \
		fi && \
		rm -rf /tmp/anime-extract && \
		rm -f anime-package.tar.gz && \
		echo "Installed to ~/.local/bin/anime" && \
		if [ -d ~/anime/cli ]; then echo "Source at ~/anime/cli"; fi
	`

	args := buildSSHArgs(target, extractCmd)
	sshCmd := exec.Command("ssh", args...)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}

func addToPathOnServer(target string) error {
	// Script to add ~/.local/bin to PATH - uses both user configs AND system-wide /etc/profile.d/
	pathScript := `
		# Function to add PATH if not already present
		add_to_path() {
			local rc_file="$1"
			local path_line='export PATH="$HOME/.local/bin:$PATH"'

			# Create file if it doesn't exist
			touch "$rc_file"

			# Check if PATH is already configured
			if ! grep -q "\.local/bin" "$rc_file" 2>/dev/null; then
				echo "" >> "$rc_file"
				echo "# Added by anime push" >> "$rc_file"
				echo "$path_line" >> "$rc_file"
				echo "Added to $rc_file"
			else
				echo "Already configured in $rc_file"
			fi
		}

		# Add to .bashrc if it exists or if bash is the shell
		if [ -f "$HOME/.bashrc" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
			add_to_path "$HOME/.bashrc"
		fi

		# Add to .zshrc if it exists or if zsh is the shell
		if [ -f "$HOME/.zshrc" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
			add_to_path "$HOME/.zshrc"
		fi

		# Add to .profile as well (works for dash and other POSIX shells)
		add_to_path "$HOME/.profile"

		# CRITICAL: Also add system-wide PATH config for all users/shells
		# This ensures it works even if shell configs aren't sourced properly
		if [ -w /etc/profile.d ] 2>/dev/null || sudo -n true 2>/dev/null; then
			sudo tee /etc/profile.d/anime-path.sh > /dev/null <<'EOF'
# Added by anime push - ensures ~/.local/bin is in PATH for all users
if [ -d "$HOME/.local/bin" ]; then
    case ":$PATH:" in
        *":$HOME/.local/bin:"*) ;;
        *) export PATH="$HOME/.local/bin:$PATH" ;;
    esac
fi
EOF
			sudo chmod +x /etc/profile.d/anime-path.sh
			echo "Added system-wide PATH config to /etc/profile.d/anime-path.sh"
		fi

		# Verify PATH is accessible in current session
		export PATH="$HOME/.local/bin:$PATH"
		if command -v anime >/dev/null 2>&1; then
			echo "✓ anime command is now available"
		else
			echo "⚠ Warning: anime installed but not yet in PATH (may need to restart shell)"
		fi
	`

	args := buildSSHArgs(target, pathScript)
	sshCmd := exec.Command("ssh", args...)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}

func getGitVersion() string {
	// Read version from VERSION file
	versionBytes, err := os.ReadFile("VERSION")
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(versionBytes))
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// testConnection tests SSH connection to the target
func testConnection(target string) error {
	_, err := execSSHPooled(target, "echo ok")
	return err
}

// detectRemoteArchitecture detects the architecture of the remote server (uses pooled connection)
func detectRemoteArchitecture(target string) (string, error) {
	// Run uname -m on the remote server to get architecture
	output, err := execSSHPooled(target, "uname -m")
	if err != nil {
		return "", fmt.Errorf("failed to detect remote architecture: %w", err)
	}

	arch := strings.TrimSpace(output)

	// Convert uname -m output to GOARCH values
	switch arch {
	case "x86_64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	case "armv7l", "armv6l":
		return "arm", nil
	case "i386", "i686":
		return "386", nil
	default:
		return "", fmt.Errorf("unknown architecture: %s", arch)
	}
}

// showConnectionSuggestions shows helpful suggestions when connection fails
func showConnectionSuggestions(target string, err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Connection failed"))
	fmt.Println()
	fmt.Printf("  Target: %s\n", theme.HighlightStyle.Render(target))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Suggestions:"))
	fmt.Println()

	errStr := err.Error()

	if strings.Contains(errStr, "Permission denied") || strings.Contains(errStr, "publickey") {
		fmt.Println(theme.DimTextStyle.Render("  • Check your SSH key:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh-add -l"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Add your SSH key if needed:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh-add ~/.ssh/id_rsa"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Test SSH connection manually:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
	} else if strings.Contains(errStr, "Connection refused") {
		fmt.Println(theme.DimTextStyle.Render("  • Server may be down or SSH service not running"))
		fmt.Println(theme.DimTextStyle.Render("  • Check if server is accessible:"))
		fmt.Println(theme.HighlightStyle.Render("    ping " + strings.Split(target, "@")[len(strings.Split(target, "@"))-1]))
	} else if strings.Contains(errStr, "No route to host") || strings.Contains(errStr, "Network unreachable") {
		fmt.Println(theme.DimTextStyle.Render("  • Check your network connection"))
		fmt.Println(theme.DimTextStyle.Render("  • Verify the IP address is correct"))
		fmt.Println(theme.DimTextStyle.Render("  • Check VPN if required"))
	} else if strings.Contains(errStr, "Could not resolve hostname") {
		fmt.Println(theme.DimTextStyle.Render("  • Check if the alias is configured:"))
		fmt.Println(theme.HighlightStyle.Render("    anime set --list"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Create alias if needed:"))
		fmt.Println(theme.HighlightStyle.Render("    anime set lambda 209.20.159.132"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  • Test SSH connection:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Check server is running and accessible"))
		fmt.Println(theme.DimTextStyle.Render("  • Verify SSH key permissions"))
	}
	fmt.Println()
}

// showBuildSuggestions shows helpful suggestions when build fails
func showBuildSuggestions(err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Build failed"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Suggestions:"))
	fmt.Println()

	errStr := err.Error()

	if strings.Contains(errStr, "cannot find package") || strings.Contains(errStr, "no required module") {
		fmt.Println(theme.DimTextStyle.Render("  • Update dependencies:"))
		fmt.Println(theme.HighlightStyle.Render("    go mod tidy"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Download missing packages:"))
		fmt.Println(theme.HighlightStyle.Render("    go mod download"))
	} else if strings.Contains(errStr, "undefined:") {
		fmt.Println(theme.DimTextStyle.Render("  • Code may have compilation errors"))
		fmt.Println(theme.DimTextStyle.Render("  • Try building locally first:"))
		fmt.Println(theme.HighlightStyle.Render("    go build"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  • Check Go installation:"))
		fmt.Println(theme.HighlightStyle.Render("    go version"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Try building locally:"))
		fmt.Println(theme.HighlightStyle.Render("    go build"))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Error details:"))
	fmt.Println(theme.DimTextStyle.Render("  " + err.Error()))
	fmt.Println()
}

// showPackagingSuggestions shows helpful suggestions when packaging fails
func showPackagingSuggestions(err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Packaging failed"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Suggestions:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  • Check disk space:"))
	fmt.Println(theme.HighlightStyle.Render("    df -h"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  • Check temp directory permissions"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Error details:"))
	fmt.Println(theme.DimTextStyle.Render("  " + err.Error()))
	fmt.Println()
}

// showRsyncSuggestions shows helpful suggestions when rsync fails
func showRsyncSuggestions(target string, err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Rsync failed"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Suggestions:"))
	fmt.Println()

	errStr := err.Error()

	if strings.Contains(errStr, "rsync: command not found") || strings.Contains(errStr, "executable file not found") {
		fmt.Println(theme.DimTextStyle.Render("  • Install rsync:"))
		if strings.Contains(runtime.GOOS, "darwin") {
			fmt.Println(theme.HighlightStyle.Render("    brew install rsync"))
		} else {
			fmt.Println(theme.HighlightStyle.Render("    sudo apt-get install rsync"))
		}
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Or use scp instead (manual workaround):"))
		fmt.Println(theme.HighlightStyle.Render("    scp anime-package.tar.gz " + target + ":~/"))
	} else if strings.Contains(errStr, "Permission denied") {
		fmt.Println(theme.DimTextStyle.Render("  • Check SSH connection:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Verify write permissions on remote server"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  • Test SSH connection:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Check network connectivity"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Error details:"))
		fmt.Println(theme.DimTextStyle.Render("  " + err.Error()))
	}
	fmt.Println()
}

// showExtractionSuggestions shows helpful suggestions when extraction fails
func showExtractionSuggestions(target string, err error) {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Extraction failed"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Suggestions:"))
	fmt.Println()

	errStr := err.Error()

	if strings.Contains(errStr, "tar: command not found") {
		fmt.Println(theme.DimTextStyle.Render("  • tar is not installed on the server"))
		fmt.Println(theme.DimTextStyle.Render("  • SSH in and install it:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println(theme.HighlightStyle.Render("    sudo apt-get install tar"))
	} else if strings.Contains(errStr, "No space left") {
		fmt.Println(theme.DimTextStyle.Render("  • Server is out of disk space"))
		fmt.Println(theme.DimTextStyle.Render("  • Check disk usage:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target + " df -h"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  • Check if file was transferred:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target + " ls -lh anime-package.tar.gz"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  • Try extracting manually:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println(theme.HighlightStyle.Render("    tar -xzf anime-package.tar.gz"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Error details:"))
		fmt.Println(theme.DimTextStyle.Render("  " + err.Error()))
	}
	fmt.Println()
}

// runPushDryRun shows what would be done without making changes
func runPushDryRun(target string) error {
	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("DRY RUN - No changes will be made"))
	fmt.Println()

	// Parse target for display
	user := "ubuntu"
	host := target
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		user = parts[0]
		host = parts[1]
	}

	// Detect architecture
	arch := pushArch
	if arch == "" {
		detectedArch, err := detectRemoteArchitecture(target)
		if err != nil {
			arch = "amd64 (default, could not detect)"
		} else {
			arch = detectedArch + " (detected)"
		}
	}

	// Get current version info
	version := Version
	if version == "" {
		version = "dev"
	}

	fmt.Println(theme.InfoStyle.Render("Push Configuration:"))
	fmt.Println()
	PrintKeyValue("Target", fmt.Sprintf("%s@%s", user, host))
	PrintKeyValue("Architecture", "linux/"+arch)
	PrintKeyValue("Version", version)
	PrintKeyValue("Include Source", fmt.Sprintf("%t", pushIncludeSource))
	PrintKeyValue("Fast Mode", fmt.Sprintf("%t", pushFast))
	PrintKeyValue("Skip Claude", fmt.Sprintf("%t", pushSkipClaude))
	PrintKeyValue("Install Aliases", fmt.Sprintf("%t", pushInstallAliases))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Actions that would be performed:"))
	fmt.Println()

	step := 1
	fmt.Printf("  %d. Build binary for linux/%s\n", step, strings.TrimSuffix(arch, " (detected)"))
	step++

	if pushIncludeSource && !pushFast {
		fmt.Printf("  %d. Package source code\n", step)
		step++
	}

	fmt.Printf("  %d. Create package archive\n", step)
	step++

	fmt.Printf("  %d. Upload to %s@%s:~/.local/bin/anime\n", step, user, host)
	step++

	if !pushSkipClaude && !pushFast {
		fmt.Printf("  %d. Deploy Claude commands and agents\n", step)
		step++
	}

	if pushInstallAliases {
		fmt.Printf("  %d. Install shell aliases\n", step)
		step++
	}

	if pushRunConfig {
		fmt.Printf("  %d. Run interactive config on server\n", step)
		step++
	}

	fmt.Printf("  %d. Verify installation\n", step)
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render("Run without --dry-run to execute"))
	return nil
}

// runBinaryOnlyPush does the fastest possible push: just the binary, force overwrite
func runBinaryOnlyPush(target string) error {
	fmt.Println(theme.InfoStyle.Render("⚡ Binary-only push (fastest)"))
	fmt.Println()

	// Step 1: Detect architecture
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[1/4]"), theme.InfoStyle.Render("Detecting architecture... "))
	if pushArch == "" {
		detectedArch, err := detectRemoteArchitecture(target)
		if err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠ (defaulting to amd64)"))
			pushArch = "amd64"
		} else {
			pushArch = detectedArch
			fmt.Println(theme.SuccessStyle.Render("✓ " + pushArch))
		}
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ " + pushArch))
	}

	// Step 2: Build binary
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("[2/4]"), theme.InfoStyle.Render("Building binary..."))
	binaryPath, version, buildTime, _, err := buildLinuxBinary()
	if err != nil {
		fmt.Printf("        %s\n", theme.ErrorStyle.Render("✗ Build failed"))
		return fmt.Errorf("build failed: %w", err)
	}

	// Get binary size
	binaryInfo, _ := os.Stat(binaryPath)
	binarySize := "unknown"
	if binaryInfo != nil {
		sizeMB := float64(binaryInfo.Size()) / (1024 * 1024)
		binarySize = fmt.Sprintf("%.1fMB", sizeMB)
	}

	fmt.Printf("        %s v%s (%s)\n", theme.SuccessStyle.Render("✓"), version, binarySize)
	defer os.Remove(binaryPath)

	// Step 3: Push directly with scp, show progress
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("[3/4]"), theme.InfoStyle.Render("Pushing binary..."))

	// First ensure ~/.local/bin exists and remove old binary (uses pooled connection)
	if _, err := execSSHPooled(target, "mkdir -p ~/.local/bin && rm -f ~/.local/bin/anime"); err != nil {
		fmt.Printf("        %s\n", theme.ErrorStyle.Render("✗ Failed to prepare remote"))
		return fmt.Errorf("failed to prepare remote: %w", err)
	}

	// Use scp with progress to copy the binary directly
	scpArgs := buildSCPArgs(binaryPath, target+":~/.local/bin/anime")
	scpCmd := exec.Command("scp", scpArgs...)
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr
	if err := scpCmd.Run(); err != nil {
		fmt.Printf("        %s\n", theme.ErrorStyle.Render("✗ SCP failed"))
		return fmt.Errorf("scp failed: %w", err)
	}

	// Make executable (uses pooled connection)
	if _, err := execSSHPooled(target, "chmod +x ~/.local/bin/anime"); err != nil {
		fmt.Printf("        %s\n", theme.ErrorStyle.Render("✗ chmod failed"))
		return fmt.Errorf("chmod failed: %w", err)
	}
	fmt.Printf("        %s\n", theme.SuccessStyle.Render("✓ Transferred"))

	// Step 4: Verify (quick, uses pooled connection)
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[4/4]"), theme.InfoStyle.Render("Verifying... "))
	output, err := execSSHPooled(target, "~/.local/bin/anime --version 2>&1 | head -1")
	if err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠ (could not verify)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓"))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Binary pushed!"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.HighlightStyle.Render(version))
	fmt.Printf("  %s   %s\n", theme.DimTextStyle.Render("Size:"), theme.InfoStyle.Render(binarySize))
	fmt.Printf("  %s   %s\n", theme.DimTextStyle.Render("Arch:"), theme.InfoStyle.Render("linux/"+pushArch))
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Built:"), theme.DimTextStyle.Render(strings.ReplaceAll(buildTime, "_", " ")))
	if len(output) > 0 {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Remote:"), theme.DimTextStyle.Render(strings.TrimSpace(output)))
	}
	fmt.Println()

	return nil
}

// runIncrementalPush uses rsync delta transfer for efficient incremental updates
func runIncrementalPush(target, serverArg string) error {
	fmt.Println(theme.InfoStyle.Render("🔄 Incremental push (delta transfer)"))
	fmt.Println()

	// Find source directory first
	sourceDir, err := findSourceDir()
	if err != nil {
		return fmt.Errorf("could not find source directory: %w", err)
	}

	// Step 1: Detect architecture + test connection in parallel
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[1/5]"), theme.InfoStyle.Render("Connecting... "))

	archChan := make(chan archResult, 1)
	connChan := make(chan error, 1)

	go func() {
		if pushArch != "" {
			archChan <- archResult{arch: pushArch, err: nil}
			return
		}
		detectedArch, err := detectRemoteArchitecture(target)
		if err != nil {
			archChan <- archResult{arch: "amd64", err: err}
		} else {
			archChan <- archResult{arch: detectedArch, err: nil}
		}
	}()

	go func() {
		connChan <- testConnection(target)
	}()

	// Wait for both
	archRes := <-archChan
	pushArch = archRes.arch

	connErr := <-connChan
	if connErr != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showConnectionSuggestions(target, connErr)
		return fmt.Errorf("connection test failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓ " + pushArch))

	// Step 2: Check if rebuild is needed
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[2/5]"), theme.InfoStyle.Render("Checking for changes... "))

	localHash, err := computeSourceHash(sourceDir)
	if err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠ (will rebuild)"))
		localHash = ""
	}

	// Check remote hash
	remoteHash := getRemoteSourceHash(target)
	needsRebuild := localHash == "" || localHash != remoteHash

	if needsRebuild {
		fmt.Println(theme.InfoStyle.Render("changes detected"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ no source changes"))
	}

	// Step 3: Build binary (if needed or always for safety)
	var binaryPath, version string
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[3/5]"), theme.InfoStyle.Render("Building... "))

	if needsRebuild {
		binaryPath, version, _, _, err = buildLinuxBinary()
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("✗"))
			showBuildSuggestions(err)
			return fmt.Errorf("build failed")
		}
		defer os.Remove(binaryPath)

		// Get binary size
		binaryInfo, _ := os.Stat(binaryPath)
		binarySize := "unknown"
		if binaryInfo != nil {
			sizeMB := float64(binaryInfo.Size()) / (1024 * 1024)
			binarySize = fmt.Sprintf("%.1fMB", sizeMB)
		}
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ v%s (%s)", version, binarySize)))
	} else {
		// Even without source changes, we need version info
		version = getGitVersionFromDir(sourceDir)
		fmt.Println(theme.DimTextStyle.Render("skipped (no changes)"))
	}

	// Step 4: Rsync with delta transfer
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("[4/5]"), theme.InfoStyle.Render("Syncing (delta transfer)..."))

	syncStats := &incrementalSyncStats{}

	// Sync binary if we rebuilt it
	if needsRebuild && binaryPath != "" {
		fmt.Printf("        %s", theme.DimTextStyle.Render("Binary: "))
		transferred, err := rsyncFileIncremental(binaryPath, target, "~/.local/bin/anime")
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("✗"))
			return fmt.Errorf("failed to sync binary: %w", err)
		}
		syncStats.binaryBytes = transferred
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ %s", formatBytes(transferred))))

		// Make executable
		chmodArgs := buildSSHArgs(target, "chmod +x ~/.local/bin/anime")
		chmodCmd := exec.Command("ssh", chmodArgs...)
		chmodCmd.Run()
	}

	// Sync source if requested
	if pushIncludeSource {
		fmt.Printf("        %s", theme.DimTextStyle.Render("Source: "))
		transferred, err := rsyncDirIncremental(sourceDir, target, "~/anime/cli")
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("✗"))
			return fmt.Errorf("failed to sync source: %w", err)
		}
		syncStats.sourceBytes = transferred
		if transferred > 0 {
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ %s", formatBytes(transferred))))
		} else {
			fmt.Println(theme.DimTextStyle.Render("✓ up to date"))
		}
	}

	// Sync embedded files if they exist
	home, _ := os.UserHomeDir()
	embeddedDir := filepath.Join(home, ".anime", "embedded")
	if _, err := os.Stat(embeddedDir); err == nil {
		fmt.Printf("        %s", theme.DimTextStyle.Render("Embedded: "))
		transferred, err := rsyncDirIncremental(embeddedDir, target, "~/.anime/embedded")
		if err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
		} else if transferred > 0 {
			syncStats.embeddedBytes = transferred
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ %s", formatBytes(transferred))))
		} else {
			fmt.Println(theme.DimTextStyle.Render("✓ up to date"))
		}
	}

	// Store the hash on remote for next incremental check
	if localHash != "" {
		storeRemoteSourceHash(target, localHash)
	}

	// Step 5: Deploy Claude assets and verify
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[5/5]"), theme.InfoStyle.Render("Finalizing... "))

	// Configure PATH
	addToPathOnServer(target)

	// Push Claude assets if not skipped
	if !pushSkipClaude {
		pushClaudeAssetsOnServer(target)
	}

	// Install shell aliases if requested
	if pushInstallAliases {
		installShellAliasesOnServer(target)
	}

	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Summary
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Incremental push complete!"))
	fmt.Println()

	totalTransferred := syncStats.binaryBytes + syncStats.sourceBytes + syncStats.embeddedBytes
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.HighlightStyle.Render(version))
	fmt.Printf("  %s    %s\n", theme.DimTextStyle.Render("Arch:"), theme.InfoStyle.Render("linux/"+pushArch))
	fmt.Printf("  %s   %s\n", theme.DimTextStyle.Render("Synced:"), theme.InfoStyle.Render(formatBytes(totalTransferred)))

	if syncStats.binaryBytes > 0 {
		fmt.Printf("           %s\n", theme.DimTextStyle.Render(fmt.Sprintf("binary: %s", formatBytes(syncStats.binaryBytes))))
	}
	if syncStats.sourceBytes > 0 {
		fmt.Printf("           %s\n", theme.DimTextStyle.Render(fmt.Sprintf("source: %s", formatBytes(syncStats.sourceBytes))))
	}
	if syncStats.embeddedBytes > 0 {
		fmt.Printf("           %s\n", theme.DimTextStyle.Render(fmt.Sprintf("embedded: %s", formatBytes(syncStats.embeddedBytes))))
	}

	fmt.Println()

	// Auto-configure server
	autoConfigureServer(serverArg, target)

	// Run config on server if requested
	if pushRunConfig {
		fmt.Println(theme.InfoStyle.Render("🔧 Running config on server..."))
		fmt.Println()

		runSSHInteractive(target, "anime config")
	}

	return nil
}

// incrementalSyncStats tracks bytes transferred during incremental sync
type incrementalSyncStats struct {
	binaryBytes   int64
	sourceBytes   int64
	embeddedBytes int64
}

// archResult holds architecture detection results (moved to package level for reuse)
type archResult struct {
	arch string
	err  error
}

// computeSourceHash computes a hash of the source directory for change detection
func computeSourceHash(sourceDir string) (string, error) {
	h := sha256.New()

	// Hash key source files
	patterns := []string{"*.go", "go.mod", "go.sum", "VERSION"}
	var files []string

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(sourceDir, pattern))
		files = append(files, matches...)

		// Also check subdirectories
		for _, subdir := range []string{"cmd", "internal"} {
			subMatches, _ := filepath.Glob(filepath.Join(sourceDir, subdir, "**", pattern))
			files = append(files, subMatches...)

			// Direct children too
			directMatches, _ := filepath.Glob(filepath.Join(sourceDir, subdir, pattern))
			files = append(files, directMatches...)
		}
	}

	// Walk cmd and internal directories for all .go files
	for _, subdir := range []string{"cmd", "internal"} {
		filepath.Walk(filepath.Join(sourceDir, subdir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				files = append(files, path)
			}
			return nil
		})
	}

	// Deduplicate and sort for consistent hashing
	seen := make(map[string]bool)
	var uniqueFiles []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			uniqueFiles = append(uniqueFiles, f)
		}
	}
	sort.Strings(uniqueFiles)

	// Hash each file's path and content
	for _, file := range uniqueFiles {
		// Include relative path in hash
		relPath, _ := filepath.Rel(sourceDir, file)
		h.Write([]byte(relPath))

		// Include file content
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		h.Write(content)
	}

	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// getRemoteSourceHash retrieves the stored source hash from the remote server
func getRemoteSourceHash(target string) string {
	args := buildSSHArgs(target, "cat ~/.anime/.source-hash 2>/dev/null || echo ''")
	cmd := exec.Command("ssh", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// storeRemoteSourceHash saves the source hash on the remote server
func storeRemoteSourceHash(target, hash string) {
	args := buildSSHArgs(target, fmt.Sprintf("mkdir -p ~/.anime && echo '%s' > ~/.anime/.source-hash", hash))
	cmd := exec.Command("ssh", args...)
	cmd.Run()
}

// rsyncFileIncremental rsyncs a single file with delta transfer, returns bytes transferred
func rsyncFileIncremental(localPath, target, remotePath string) (int64, error) {
	// Ensure remote directory exists
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "~" && remoteDir != "." {
		prepArgs := buildSSHArgs(target, fmt.Sprintf("mkdir -p %s", remoteDir))
		prepCmd := exec.Command("ssh", prepArgs...)
		prepCmd.Run()
	}

	// Use rsync with checksum-based delta transfer
	// --checksum: use checksum for determining changes (more accurate than mtime)
	// --partial: keep partially transferred files
	// --inplace: update files in place (better for delta)
	// --stats: show transfer statistics
	args := []string{
		"-az",
		"--checksum",
		"--partial",
		"--inplace",
		"--stats",
	}

	// Add SSH options with embedded key if available
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		defer cleanup()
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", keyPath)
		args = append(args, "-e", sshCmd)
	}

	args = append(args, localPath, fmt.Sprintf("%s:%s", target, remotePath))

	rsyncCmd := exec.Command("rsync", args...)
	output, err := rsyncCmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", err, string(output))
	}

	// Parse transferred bytes from rsync stats
	return parseRsyncTransferred(string(output)), nil
}

// rsyncDirIncremental rsyncs a directory with delta transfer, returns bytes transferred
func rsyncDirIncremental(localDir, target, remoteDir string) (int64, error) {
	// Ensure remote directory exists
	prepArgs := buildSSHArgs(target, fmt.Sprintf("mkdir -p %s", remoteDir))
	prepCmd := exec.Command("ssh", prepArgs...)
	prepCmd.Run()

	// Build exclude patterns
	excludes := []string{
		"--exclude=.git",
		"--exclude=.DS_Store",
		"--exclude=*.test",
		"--exclude=build/",
		"--exclude=.idea/",
		"--exclude=.vscode/",
	}

	// Use rsync with delta transfer
	args := []string{
		"-az",
		"--checksum",
		"--partial",
		"--delete", // Remove files on remote that don't exist locally
		"--stats",
	}

	// Add SSH options with embedded key if available
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		defer cleanup()
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", keyPath)
		args = append(args, "-e", sshCmd)
	}

	args = append(args, excludes...)
	args = append(args, localDir+"/", fmt.Sprintf("%s:%s/", target, remoteDir))

	rsyncCmd := exec.Command("rsync", args...)
	output, err := rsyncCmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", err, string(output))
	}

	return parseRsyncTransferred(string(output)), nil
}

// parseRsyncTransferred parses the "Total transferred file size" from rsync --stats output
func parseRsyncTransferred(output string) int64 {
	// Look for "Total transferred file size: X bytes" or similar
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Total transferred file size") {
			// Extract the number
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				numStr := strings.TrimSpace(parts[1])
				numStr = strings.ReplaceAll(numStr, ",", "")
				numStr = strings.Split(numStr, " ")[0]
				var n int64
				fmt.Sscanf(numStr, "%d", &n)
				return n
			}
		}
		// Also check for "sent X bytes" pattern
		if strings.HasPrefix(line, "sent ") && strings.Contains(line, " bytes") {
			var sent int64
			fmt.Sscanf(line, "sent %d bytes", &sent)
			if sent > 0 {
				return sent
			}
		}
	}
	return 0
}

// installShellAliasesOnServer installs anime shell aliases and sources bashrc/zshrc
func installShellAliasesOnServer(target string) error {
	// Install shell aliases using anime aliases install, then source the config
	installScript := `
		export PATH="$HOME/.local/bin:$PATH"

		# Run anime aliases install
		if anime aliases install 2>&1 | grep -qE "(installed|updated)"; then
			echo "Aliases installed"
		else
			echo "Aliases may already be installed"
		fi

		# Source the appropriate shell config
		if [ -f "$HOME/.zshrc" ]; then
			source "$HOME/.zshrc" 2>/dev/null || true
			echo "Sourced .zshrc"
		elif [ -f "$HOME/.bashrc" ]; then
			source "$HOME/.bashrc" 2>/dev/null || true
			echo "Sourced .bashrc"
		fi
	`

	args := buildSSHArgs(target, installScript)
	sshCmd := exec.Command("ssh", args...)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}

// pushClaudeAssetsOnServer runs anime claude commands push and agents push on the remote server
// This extracts the embedded Claude Code commands and agents to ~/.claude/
func pushClaudeAssetsOnServer(target string) error {
	// Run anime claude commands push and agents push on the remote server
	// The binary already has the commands/agents embedded via go:embed
	pushScript := `
		export PATH="$HOME/.local/bin:$PATH"

		# Push embedded commands to ~/.claude/commands/
		if anime claude commands push 2>&1 | grep -q "Pushed"; then
			echo "Commands deployed"
		else
			echo "No commands to deploy or already up to date"
		fi

		# Push embedded agents to ~/.claude/agents/
		if anime claude agents push 2>&1 | grep -q "Pushed"; then
			echo "Agents deployed"
		else
			echo "No agents to deploy or already up to date"
		fi
	`

	args := buildSSHArgs(target, pushScript)
	sshCmd := exec.Command("ssh", args...)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}

// verifyServerVersion checks that the version on the server matches the expected version
func verifyServerVersion(target, expectedVersion string) error {
	// Run anime --version on the server
	// Use PATH and explicitly call anime from ~/.local/bin
	args := buildSSHArgs(target, "PATH=$HOME/.local/bin:$PATH anime --version 2>&1 | grep 'Version:' | awk '{print $2}'")
	versionCmd := exec.Command("ssh", args...)
	output, err := versionCmd.Output()
	if err != nil {
		return fmt.Errorf("could not get version from server")
	}

	serverVersion := strings.TrimSpace(string(output))

	// Compare versions
	if serverVersion != expectedVersion {
		return fmt.Errorf("version mismatch - expected %s, got %s on server", expectedVersion, serverVersion)
	}

	return nil
}
// autoConfigureServer saves the pushed server to the anime config
func autoConfigureServer(serverArg, target string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse target to extract user and host
	parts := strings.Split(target, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", target)
	}
	user := parts[0]
	host := parts[1]

	// Check if this server already exists in config (by host)
	var existingServer *config.Server
	for i := range cfg.Servers {
		if cfg.Servers[i].Host == host {
			existingServer = &cfg.Servers[i]
			break
		}
	}

	// If using an existing alias, just update the alias mapping
	if serverArg != host && !strings.Contains(serverArg, "@") && !strings.Contains(serverArg, ".") {
		// This is an alias name, not an IP/hostname
		cfg.SetAlias(serverArg, target)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save alias: %w", err)
		}
		fmt.Print(theme.DimTextStyle.Render("▶ Saved alias... "))
		fmt.Println(theme.SuccessStyle.Render("✓"))
		return nil
	}

	// If server exists, update it
	if existingServer != nil {
		existingServer.User = user
		existingServer.Host = host
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to update server: %w", err)
		}
		fmt.Print(theme.DimTextStyle.Render("▶ Updated server config... "))
		fmt.Println(theme.SuccessStyle.Render("✓"))
		return nil
	}

	// Otherwise, create a new server entry
	serverName := host
	if strings.Contains(host, ".") {
		// If it's an IP, use "server-{last-octet}" or similar
		octets := strings.Split(host, ".")
		if len(octets) == 4 {
			serverName = "server-" + octets[3]
		}
	}

	newServer := config.Server{
		Name:        serverName,
		Host:        host,
		User:        user,
		SSHKey:      "", // Could be detected from SSH config
		CostPerHour: 0,  // User can configure later
		Modules:     []string{},
	}

	cfg.AddServer(newServer)

	// Also create an alias if serverArg was provided
	if serverArg != "" && serverArg != host {
		cfg.SetAlias(serverArg, target)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save server: %w", err)
	}

	fmt.Print(theme.DimTextStyle.Render("▶ Saved server config... "))
	fmt.Println(theme.SuccessStyle.Render("✓"))

	return nil
}

// buildSSHArgs builds SSH command arguments, using embedded key if available
func buildSSHArgs(target string, command string) []string {
	args := []string{"-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking=accept-new"}

	// Try to use embedded SSH key
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		args = append(args, "-i", keyPath)
		// Note: cleanup will be called when the process exits or we need to manually clean up
		// For short-lived commands this is fine; the temp file will be cleaned up
		_ = cleanup // We can't defer here, but the file is in temp and will be cleaned up
	}

	args = append(args, target, command)
	return args
}

// buildRsyncArgs builds rsync command arguments, using embedded key if available
func buildRsyncArgs(extraArgs []string, source, dest string) []string {
	args := extraArgs

	// Try to use embedded SSH key
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", keyPath)
		args = append(args, "-e", sshCmd)
		_ = cleanup
	}

	args = append(args, source, dest)
	return args
}

// buildSCPArgs builds SCP command arguments, using embedded key if available
func buildSCPArgs(source, dest string) []string {
	args := []string{"-o", "Compression=yes", "-o", "StrictHostKeyChecking=accept-new"}

	// Try to use embedded SSH key
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		args = append(args, "-i", keyPath)
		_ = cleanup
	}

	args = append(args, source, dest)
	return args
}

// execSSH executes an SSH command using the embedded key if available
func execSSH(target, command string) ([]byte, error) {
	args := buildSSHArgs(target, command)
	cmd := exec.Command("ssh", args...)
	return cmd.CombinedOutput()
}

// parseTarget splits a target like "ubuntu@192.168.1.1" into user and host
func parseTarget(target string) (user, host string) {
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		return parts[0], parts[1]
	}
	return "ubuntu", target
}

// getPooledClient gets or creates a pooled SSH connection for the target
func getPooledClient(target string) (*ssh.Client, error) {
	user, host := parseTarget(target)

	// Get embedded key path if available
	var keyPath string
	if kp, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		keyPath = kp
		_ = cleanup // Temp file cleaned up on process exit
	}

	return ssh.GetPool().GetWithOptions(host, user, keyPath, ssh.ClientOptions{
		StrictHostKeyChecking: false, // Accept new keys like -o StrictHostKeyChecking=accept-new
		Interactive:           false,
	})
}

// execSSHPooled executes a command via SSH (uses simple exec, not Go SSH library)
func execSSHPooled(target, command string) (string, error) {
	args := buildSSHArgs(target, command)
	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runSSHInteractive executes an interactive SSH command with TTY allocation
func runSSHInteractive(target, command string) error {
	args := []string{"-t"} // Allocate pseudo-TTY for interactive use
	args = append(args, buildSSHArgs(target, command)...)
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runFilePush pushes files/folders to a remote server
// args[0] = source, args[1] = server, args[2] = dest (optional, defaults to ~/)
func runFilePush(args []string) error {
	source := args[0]
	server := args[1]
	dest := "~/"
	if len(args) >= 3 {
		dest = args[2]
	}

	// Check source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source not found: %s", source)
	}

	// Load config and resolve server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve server to user@host
	target, err := resolveFilePushTarget(cfg, server)
	if err != nil {
		return err
	}

	// If dest is just ~/ and source is a directory, append source dirname
	if dest == "~/" && sourceInfo.IsDir() {
		dest = "~/" + filepath.Base(source)
	}

	fmt.Println(theme.InfoStyle.Render("📦 Pushing files to remote server"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Source:"), theme.InfoStyle.Render(source))
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target+":"+dest))
	fmt.Println()

	if pushDryRun {
		fmt.Println(theme.WarningStyle.Render("DRY RUN - no changes will be made"))
		fmt.Println()
	}

	// Build rsync command
	rsyncArgs := []string{"-av", "--progress", "-z", "--stats"}

	if pushDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}

	if pushDelete {
		rsyncArgs = append(rsyncArgs, "--delete")
	}

	// Add user-specified excludes
	for _, exc := range pushExclude {
		rsyncArgs = append(rsyncArgs, "--exclude", exc)
	}

	// Add common excludes
	rsyncArgs = append(rsyncArgs,
		"--exclude", ".git",
		"--exclude", "node_modules",
		"--exclude", "__pycache__",
		"--exclude", "*.pyc",
		"--exclude", ".venv",
		"--exclude", "venv",
		"--exclude", ".DS_Store",
		"--exclude", "target",      // Rust/Cargo build output
		"--exclude", "dist",        // Common frontend build output
		"--exclude", "build",       // Generic build output
		"--exclude", ".next",       // Next.js build output
		"--exclude", ".nuxt",       // Nuxt.js build output
	)

	// Add SSH options with embedded key if available
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		defer cleanup()
		sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=accept-new", keyPath)
		rsyncArgs = append(rsyncArgs, "-e", sshCmd)
	}

	// Add trailing slash to source if it's a directory to sync contents
	sourcePath := source
	if sourceInfo.IsDir() && !strings.HasSuffix(source, "/") {
		sourcePath = source + "/"
	}

	rsyncArgs = append(rsyncArgs, sourcePath, target+":"+dest)

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdin = os.Stdin
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Get actual size of source for summary
	sourceSize, fileCount := getSourceStats(source)

	fmt.Println()
	if pushDryRun {
		fmt.Println(theme.WarningStyle.Render("Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✨ Push complete!"))
		fmt.Println()
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Source:"), theme.InfoStyle.Render(source))
		fmt.Printf("  %s    %s\n", theme.DimTextStyle.Render("Dest:"), theme.HighlightStyle.Render(target+":"+dest))
		fmt.Printf("  %s   %s (%d files)\n", theme.DimTextStyle.Render("Size:"), theme.InfoStyle.Render(formatBytes(sourceSize)), fileCount)
	}
	fmt.Println()

	return nil
}

// getSourceStats returns the total size and file count of a source path
func getSourceStats(source string) (int64, int) {
	var totalSize int64
	var fileCount int

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common excluded directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "__pycache__" ||
				name == ".venv" || name == "venv" || name == "target" ||
				name == "dist" || name == "build" || name == ".next" || name == ".nuxt" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		totalSize += info.Size()
		fileCount++
		return nil
	})

	return totalSize, fileCount
}

// resolveFilePushTarget resolves a server name to user@host for file push
func resolveFilePushTarget(cfg *config.Config, server string) (string, error) {
	// If already has @, use as-is
	if strings.Contains(server, "@") {
		return server, nil
	}

	// Check servers by name
	if srv, err := cfg.GetServer(server); err == nil && srv != nil {
		return srv.User + "@" + srv.Host, nil
	}

	// Check aliases
	if alias := cfg.GetAlias(server); alias != "" {
		if strings.Contains(alias, "@") {
			return alias, nil
		}
		return "ubuntu@" + alias, nil
	}

	// Try SSH config resolution
	if resolved, err := trySSHConfigResolve(server); err == nil {
		return resolved, nil
	}

	// If it looks like an IP/hostname, add ubuntu@
	if strings.Contains(server, ".") {
		return "ubuntu@" + server, nil
	}

	return "", fmt.Errorf("could not resolve server: %s (use 'anime add' to configure)", server)
}
