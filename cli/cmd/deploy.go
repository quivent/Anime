package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [server]",
	Short: "Deploy binary and source to a server",
	Long: `Deploy the anime binary and CLI source folder (with .git) to a remote server.

Server formats:
  - user@IP          (e.g., ubuntu@192.168.1.100)
  - IP               (defaults to ubuntu@IP)
  - alias            (from anime set)
  - (default: lambda if no server specified)

Examples:
  anime deploy blackwell              # Deploy to aliased server
  anime deploy ubuntu@10.0.0.1        # Deploy with explicit user
  anime deploy                        # Deploy to default (lambda)
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Get server argument or default
	cfg, _ := config.Load()
	server := "lambda"
	if cfg != nil {
		if def := cfg.GetDefaultServer(); def != "" {
			server = def
		}
	}
	if len(args) > 0 {
		server = args[0]
	}

	// Resolve alias to target
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🚀 Deploying to " + target))
	fmt.Println()

	// Find source directory (cli/) and its parent (anime/ with .git)
	sourceDir, err := findSourceDir()
	if err != nil {
		return fmt.Errorf("could not find source directory: %w", err)
	}
	// Get parent directory which contains .git
	repoDir := filepath.Dir(sourceDir)

	// Step 1: Build binary for Linux
	fmt.Printf("  %s Building for linux/amd64... ", theme.InfoStyle.Render("[1/3]"))
	binaryPath, err := buildDeployBinary(sourceDir)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Step 2: Push binary
	fmt.Printf("  %s Pushing binary... ", theme.InfoStyle.Render("[2/3]"))
	if err := deployBinary(target, binaryPath); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to push binary: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Step 3: Push repo folder with .git
	fmt.Printf("  %s Syncing repo (with .git)... ", theme.InfoStyle.Render("[3/3]"))
	if err := deploySource(target, repoDir); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to sync source: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Deploy complete"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Target:"), theme.InfoStyle.Render(target))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Binary:"), theme.DimTextStyle.Render("~/.local/bin/anime"))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.DimTextStyle.Render("~/anime (with .git)"))
	fmt.Println()

	return nil
}

func buildDeployBinary(sourceDir string) (string, error) {
	buildDir := filepath.Join(sourceDir, "build")
	os.MkdirAll(buildDir, 0755)

	binaryPath := filepath.Join(buildDir, "anime-linux-amd64")

	// Read version
	version := "dev"
	if data, err := os.ReadFile(filepath.Join(sourceDir, "VERSION")); err == nil {
		version = strings.TrimSpace(string(data))
	}

	buildCmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf("-X github.com/joshkornreich/anime/cmd.Version=%s", version),
		"-o", binaryPath,
		".")
	buildCmd.Dir = sourceDir
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

	if output, err := buildCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w: %s", err, string(output))
	}

	return binaryPath, nil
}

func deployBinary(target, binaryPath string) error {
	// Create remote directory and remove old binary
	sshArgs := []string{"-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=accept-new"}
	prepCmd := exec.Command("ssh", append(sshArgs, target, "mkdir -p ~/.local/bin && rm -f ~/.local/bin/anime")...)
	if err := prepCmd.Run(); err != nil {
		return fmt.Errorf("failed to prepare remote: %w", err)
	}

	// Get file size for progress display
	fileInfo, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %w", err)
	}
	fileSize := fileInfo.Size()
	fileSizeMB := float64(fileSize) / (1024 * 1024)

	// Use rsync with progress for better feedback
	// Note: macOS ships with rsync 2.6.9 which doesn't support --info=progress2
	sshOpts := "ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10"
	rsyncCmd := exec.Command("rsync", "-avz", "--progress", "-e", sshOpts,
		binaryPath, target+":~/.local/bin/anime")

	// Capture stderr for error messages
	var stderrBuf strings.Builder
	rsyncCmd.Stderr = &stderrBuf

	// Create pipes for stdout
	stdout, err := rsyncCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := rsyncCmd.Start(); err != nil {
		return fmt.Errorf("failed to start rsync: %w", err)
	}

	// Read and display progress
	buf := make([]byte, 256)
	lastPercent := -1
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			line := string(buf[:n])
			// Parse progress percentage from rsync output
			// Format: "  1,234,567 100%   12.34MB/s    0:00:01"
			if percent := parseRsyncProgress(line); percent >= 0 && percent != lastPercent {
				lastPercent = percent
				bar := renderProgressBar(percent, 30)
				fmt.Printf("\r    %s %.1fMB %3d%%", bar, fileSizeMB, percent)
			}
		}
		if err != nil {
			break
		}
	}

	if err := rsyncCmd.Wait(); err != nil {
		fmt.Println() // newline after progress
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg != "" {
			return fmt.Errorf("rsync failed: %s", errMsg)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Printf("\r    %s %.1fMB 100%%\n", renderProgressBar(100, 30), fileSizeMB)

	// Make executable
	chmodCmd := exec.Command("ssh", append(sshArgs, target, "chmod +x ~/.local/bin/anime")...)
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	return nil
}

// parseRsyncProgress extracts percentage from rsync --info=progress2 output
func parseRsyncProgress(line string) int {
	// Look for percentage pattern like "42%" or "100%"
	for i := 0; i < len(line)-1; i++ {
		if line[i] == '%' {
			// Find start of number
			j := i - 1
			for j >= 0 && (line[j] >= '0' && line[j] <= '9') {
				j--
			}
			if j < i-1 {
				numStr := line[j+1 : i]
				var percent int
				fmt.Sscanf(numStr, "%d", &percent)
				return percent
			}
		}
	}
	return -1
}

// renderProgressBar creates a visual progress bar
func renderProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return theme.SuccessStyle.Render("[") + theme.HighlightStyle.Render(bar) + theme.SuccessStyle.Render("]")
}

func deploySource(target, repoDir string) error {
	sshArgs := []string{"-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=accept-new"}

	// Create remote directory and remove extras (keep only cli/ and .git/)
	prepCmd := exec.Command("ssh", append(sshArgs, target,
		"mkdir -p ~/anime && cd ~/anime && find . -maxdepth 1 ! -name . ! -name cli ! -name .git -exec rm -rf {} +")...)
	if err := prepCmd.Run(); err != nil {
		return fmt.Errorf("failed to prepare remote directory: %w", err)
	}

	sshOpts := "ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10"

	// Sync cli/ folder
	rsyncCli := exec.Command("rsync", "-az", "--delete", "-e", sshOpts,
		repoDir+"/cli/", target+":~/anime/cli/")
	if output, err := rsyncCli.CombinedOutput(); err != nil {
		return fmt.Errorf("rsync cli failed: %w: %s", err, string(output))
	}

	// Sync .git/ folder
	rsyncGit := exec.Command("rsync", "-az", "--delete", "-e", sshOpts,
		repoDir+"/.git/", target+":~/anime/.git/")
	if output, err := rsyncGit.CombinedOutput(); err != nil {
		return fmt.Errorf("rsync .git failed: %w: %s", err, string(output))
	}

	return nil
}
