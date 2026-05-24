package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	pushArch          string
	pushIncludeSource bool
	pushRunConfig     bool
)

var pushCmd = &cobra.Command{
	Use:   "push [server]",
	Short: "Build and push anime binary to remote server",
	Long: `Build the anime binary with optional source code and rsync to a remote server.

Server formats:
  - user@IP          (e.g., ubuntu@192.168.1.100)
  - IP               (defaults to ubuntu@IP)
  - alias            (from anime config or .ssh/config)
  - (default: lambda if no server specified)

Examples:
  anime set lambda 209.20.159.132             # Create alias first
  anime push                                  # Defaults to 'lambda' alias
  anime push lambda                           # Use anime alias
  anime push lambda --config                  # Push and run config on server
  anime push 192.168.1.100                    # Uses ubuntu@192.168.1.100
  anime push user@10.0.0.5                    # Uses specified user
  anime push my-lambda-server                 # Uses SSH config alias

The binary is built for Linux by default (configurable with --arch).
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPush,
}

func init() {
	pushCmd.Flags().StringVar(&pushArch, "arch", "", "Target architecture (amd64 or arm64, auto-detected if not specified)")
	pushCmd.Flags().BoolVar(&pushIncludeSource, "source", true, "Include source code in the package")
	pushCmd.Flags().BoolVarP(&pushRunConfig, "config", "c", false, "Run config on server after push")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	// Default to "lambda" if no server specified
	server := "lambda"
	if len(args) > 0 {
		server = args[0]
	}

	// Parse server argument
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	fmt.Println(theme.InfoStyle.Render("🚀 Pushing anime to remote server"))
	fmt.Println()

	// Auto-detect architecture if not specified
	if pushArch == "" {
		fmt.Print(theme.DimTextStyle.Render("▶ Detecting remote architecture... "))
		detectedArch, err := detectRemoteArchitecture(target)
		if err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			fmt.Println(theme.DimTextStyle.Render("  Could not detect architecture, defaulting to amd64"))
			pushArch = "amd64"
		} else {
			pushArch = detectedArch
			fmt.Println(theme.SuccessStyle.Render("✓ " + pushArch))
		}
	}

	// Step 1: Build binary for Linux (do this first - fail fast if build breaks)
	fmt.Print(theme.DimTextStyle.Render("▶ Building binary... "))
	binaryPath, version, buildTime, sourceDir, err := buildLinuxBinary()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showBuildSuggestions(err)
		return fmt.Errorf("build failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))
	defer os.Remove(binaryPath) // Clean up temp binary

	// Show build info
	fmt.Println()
	fmt.Printf("  Target:  %s\n", theme.HighlightStyle.Render(target))
	fmt.Printf("  Version: %s\n", theme.HighlightStyle.Render(version))
	fmt.Printf("  Built:   %s\n", theme.HighlightStyle.Render(strings.ReplaceAll(buildTime, "_", " ")))
	fmt.Printf("  Arch:    %s\n", theme.HighlightStyle.Render("linux/"+pushArch))
	if pushIncludeSource {
		fmt.Printf("  Source:  %s\n", theme.SuccessStyle.Render("included"))
	}
	fmt.Println()

	// Step 2: Test SSH connection
	fmt.Print(theme.DimTextStyle.Render("▶ Testing connection... "))
	if err := testConnection(target); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showConnectionSuggestions(target, err)
		return fmt.Errorf("connection test failed")
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

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
	fmt.Print(theme.DimTextStyle.Render("▶ Deploying Claude commands... "))
	if err := pushClaudeAssetsOnServer(target); err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓"))
	}

	// Step 7: Verify version on server
	fmt.Print(theme.DimTextStyle.Render("▶ Verifying version... "))
	if err := verifyServerVersion(target, version); err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  Warning: " + err.Error()))
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓"))
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

		sshCmd := exec.Command("ssh", "-t", target, "anime config")
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
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

	// First, check anime config aliases
	cfg, err := config.Load()
	if err == nil {
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
	// Rsync command
	rsyncCmd := exec.Command("rsync", "-avz", "--progress", localPath, target+":~/")
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

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

	sshCmd := exec.Command("ssh", target, extractCmd)
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

	sshCmd := exec.Command("ssh", target, pathScript)
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
	testCmd := exec.Command("ssh", "-o", "ConnectTimeout=5", target, "echo ok")
	output, err := testCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// detectRemoteArchitecture detects the architecture of the remote server
func detectRemoteArchitecture(target string) (string, error) {
	// Run uname -m on the remote server to get architecture
	archCmd := exec.Command("ssh", "-o", "ConnectTimeout=5", target, "uname -m")
	output, err := archCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to detect remote architecture: %w", err)
	}

	arch := strings.TrimSpace(string(output))

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

	sshCmd := exec.Command("ssh", target, pushScript)
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
	versionCmd := exec.Command("ssh", target, "PATH=$HOME/.local/bin:$PATH anime --version 2>&1 | grep 'Version:' | awk '{print $2}'")
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
