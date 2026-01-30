package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	shipVerbose       bool
	shipKeepTar       bool
	shipClearArtifacts bool
	shipNoClear       bool
)

var shipCmd = &cobra.Command{
	Use:   "ship <source> [destination]",
	Short: "Tar, rsync, and untar files to a remote destination",
	Long: `Ship files or directories to a remote server efficiently.

This command:
1. Creates a compressed tar archive of the source
2. Transfers it via rsync (with compression and progress)
3. Unpacks it on the remote server
4. Cleans up temporary files

The destination format can be:
  - servername:/path          (uses anime's server configs)
  - user@host:/path           (direct SSH target)
  - host:/path                (assumes ubuntu@ user)
  - (default: lambda:/home/ubuntu/ if no destination specified)

Examples:
  anime ship ./myapp                                  # Ships to lambda:/home/ubuntu/
  anime ship ./myapp lambda:/home/ubuntu/apps/
  anime ship config.yaml production:/etc/myapp/
  anime ship build/ ubuntu@192.168.1.10:/var/www/
  anime ship --keep-tar ./data myserver:/backup/
  anime ship --clear ./dist production:/var/www/      # Clear destination first
  anime ship --no-clear ./app lambda:/apps/           # Skip clear prompt`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runShip,
}

func init() {
	rootCmd.AddCommand(shipCmd)
	shipCmd.Flags().BoolVarP(&shipVerbose, "verbose", "v", false, "Show verbose output")
	shipCmd.Flags().BoolVarP(&shipKeepTar, "keep-tar", "k", false, "Keep the tar file after shipping")
	shipCmd.Flags().BoolVarP(&shipClearArtifacts, "clear", "c", false, "Clear destination before shipping")
	shipCmd.Flags().BoolVar(&shipNoClear, "no-clear", false, "Skip clear destination prompt")
}

func runShip(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Default to "lambda:/home/ubuntu/" if no destination specified
	destination := "lambda:/home/ubuntu/"
	if len(args) > 1 {
		destination = args[1]
	}

	// Validate source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source does not exist: %s", source)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🚢 ANIME SHIP 🚢"))
	fmt.Println()

	// Parse destination
	parts := strings.SplitN(destination, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("destination must be in format 'host:/path' or 'server:/path'")
	}

	hostPart := parts[0]
	remotePath := parts[1]

	// Load config to resolve target
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve the host part (could be server name, alias, or user@host)
	sshTarget, err := resolveSSHTarget(cfg, hostPart)
	if err != nil {
		return fmt.Errorf("could not resolve destination host: %w", err)
	}

	// Create tar file
	timestamp := time.Now().Format("20060102-150405")
	baseName := filepath.Base(source)
	if baseName == "." || baseName == "/" {
		baseName = "archive"
	}
	tarName := fmt.Sprintf("%s-%s.tar.gz", baseName, timestamp)
	tarPath := filepath.Join(os.TempDir(), tarName)

	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Source:"),
		theme.HighlightStyle.Render(source))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Target:"),
		theme.SuccessStyle.Render(sshTarget+":"+remotePath))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Archive:"),
		theme.DimTextStyle.Render(tarName))
	fmt.Println()

	// Step 0: Handle clearing artifacts at destination
	if err := handleClearArtifacts(sshTarget, remotePath, baseName, sourceInfo.IsDir()); err != nil {
		return err
	}

	// Step 1: Create tar with progress
	fmt.Println(theme.GlowStyle.Render("  📦 Creating archive..."))

	// Count files for progress
	fileCount, totalSize := countFilesAndSize(source)
	fmt.Printf("  %s %d files (%.2f MB)\n",
		theme.DimTextStyle.Render("Found:"),
		fileCount,
		float64(totalSize)/(1024*1024))

	if err := createTarWithProgress(source, tarPath, sourceInfo.IsDir(), fileCount); err != nil {
		return fmt.Errorf("failed to create tar: %w", err)
	}

	// Get tar size for display
	tarInfo, _ := os.Stat(tarPath)
	tarSize := float64(tarInfo.Size()) / (1024 * 1024) // Convert to MB
	fmt.Printf("\r  %s Archived %.2f MB (compressed)\n", theme.SuccessStyle.Render("✓"), tarSize)
	fmt.Println()

	// Clean up tar on exit unless --keep-tar is specified
	if !shipKeepTar {
		defer func() {
			os.Remove(tarPath)
			if shipVerbose {
				fmt.Println(theme.DimTextStyle.Render("  Cleaned up local tar file"))
			}
		}()
	}

	// Step 2: Rsync to remote with progress
	fmt.Println(theme.GlowStyle.Render("  🚀 Transferring via rsync..."))
	remoteHost := strings.Split(sshTarget, "@")[1]
	remoteTarPath := fmt.Sprintf("/tmp/%s", tarName)

	if err := rsyncFileWithProgress(tarPath, sshTarget, remoteTarPath, tarInfo.Size()); err != nil {
		return fmt.Errorf("failed to rsync: %w", err)
	}
	fmt.Printf("  %s Transferred %.2f MB to %s\n", theme.SuccessStyle.Render("✓"), tarSize, remoteHost)
	fmt.Println()

	// Step 3: Untar on remote
	fmt.Println(theme.GlowStyle.Render("  📂 Unpacking on remote..."))

	// Determine the tar extraction command based on source type
	var untarCmd string
	if sourceInfo.IsDir() {
		// For directories, extract directly to destination
		untarCmd = fmt.Sprintf("mkdir -p %s && tar -xzf %s -C %s", remotePath, remoteTarPath, remotePath)
	} else {
		// For files, ensure parent directory exists and extract
		untarCmd = fmt.Sprintf("mkdir -p %s && tar -xzf %s -C %s", remotePath, remoteTarPath, remotePath)
	}

	sshCmd := exec.Command("ssh", sshTarget, untarCmd)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("failed to untar on remote: %w", err)
	}
	fmt.Printf("  %s Unpacked to %s\n", theme.SuccessStyle.Render("✓"), remotePath)
	fmt.Println()

	// Step 4: Clean up remote tar
	fmt.Println(theme.GlowStyle.Render("  🧹 Cleaning up..."))
	cleanupCmd := exec.Command("ssh", sshTarget, fmt.Sprintf("rm -f %s", remoteTarPath))
	if err := cleanupCmd.Run(); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: Failed to remove remote tar file"))
	} else {
		fmt.Printf("  %s Removed remote tar file\n", theme.SuccessStyle.Render("✓"))
	}
	fmt.Println()

	// Success summary
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.GlowStyle.Render("  ✨ Ship Complete!"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s shipped to %s:%s\n",
		theme.HighlightStyle.Render(baseName),
		theme.InfoStyle.Render(remoteHost),
		theme.SuccessStyle.Render(remotePath))
	fmt.Println()

	if shipKeepTar {
		fmt.Printf("  %s %s\n",
			theme.InfoStyle.Render("Local tar saved:"),
			theme.DimTextStyle.Render(tarPath))
		fmt.Println()
	}

	return nil
}

// handleClearArtifacts prompts user and clears destination if requested
func handleClearArtifacts(sshTarget, remotePath, baseName string, isDir bool) error {
	// Skip if --no-clear flag is set
	if shipNoClear {
		return nil
	}

	// Determine what path would be cleared
	clearPath := remotePath
	if isDir {
		clearPath = filepath.Join(remotePath, baseName)
	}

	// If --clear flag is set, clear without prompting
	if shipClearArtifacts {
		return clearRemoteArtifacts(sshTarget, clearPath)
	}

	// Otherwise, prompt the user
	fmt.Printf("  %s Clear existing artifacts at %s? [y/N]: ",
		theme.WarningStyle.Render("⚠"),
		theme.HighlightStyle.Render(clearPath))

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil // Don't fail on read error, just skip clearing
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		return clearRemoteArtifacts(sshTarget, clearPath)
	}

	fmt.Println(theme.DimTextStyle.Render("  Skipping artifact cleanup"))
	fmt.Println()
	return nil
}

// clearRemoteArtifacts removes existing files at the destination
func clearRemoteArtifacts(sshTarget, remotePath string) error {
	fmt.Print(theme.GlowStyle.Render("  🗑️  Clearing remote artifacts..."))

	// First check if path exists
	checkCmd := exec.Command("ssh", sshTarget, fmt.Sprintf("test -e %s && echo exists || echo notfound", remotePath))
	output, err := checkCmd.Output()
	if err != nil {
		fmt.Println(theme.WarningStyle.Render(" skipped (could not check)"))
		fmt.Println()
		return nil
	}

	if strings.TrimSpace(string(output)) == "notfound" {
		fmt.Println(theme.DimTextStyle.Render(" nothing to clear"))
		fmt.Println()
		return nil
	}

	// Clear the path
	clearCmd := exec.Command("ssh", sshTarget, fmt.Sprintf("rm -rf %s", remotePath))
	if err := clearCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render(" failed"))
		return fmt.Errorf("failed to clear remote artifacts: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render(" ✓"))
	fmt.Println()
	return nil
}

// countFilesAndSize counts files and total size for progress reporting
func countFilesAndSize(source string) (int, int64) {
	var count int
	var totalSize int64

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
			totalSize += info.Size()
		}
		return nil
	})

	return count, totalSize
}

// createTarWithProgress creates a tar archive with progress indication
func createTarWithProgress(source, tarPath string, isDir bool, totalFiles int) error {
	var tarCmd *exec.Cmd

	parentDir := filepath.Dir(source)
	baseName := filepath.Base(source)

	if isDir {
		tarCmd = exec.Command("tar", "-cvzf", tarPath, "-C", parentDir, baseName)
	} else {
		tarCmd = exec.Command("tar", "-cvzf", tarPath, "-C", parentDir, baseName)
	}

	// Capture stderr for progress (tar -v outputs to stderr)
	stderr, err := tarCmd.StderrPipe()
	if err != nil {
		// Fallback to simple tar without progress
		return createTarSimple(source, tarPath, isDir)
	}

	if err := tarCmd.Start(); err != nil {
		return err
	}

	// Track progress
	var processed int
	var mu sync.Mutex
	scanner := bufio.NewScanner(stderr)

	go func() {
		for scanner.Scan() {
			mu.Lock()
			processed++
			if totalFiles > 0 {
				percent := (processed * 100) / totalFiles
				fmt.Printf("\r  %s %d/%d files (%d%%)",
					theme.DimTextStyle.Render("Progress:"),
					processed, totalFiles, percent)
			}
			mu.Unlock()
		}
	}()

	err = tarCmd.Wait()
	fmt.Print("\r") // Clear progress line
	return err
}

// createTarSimple creates tar without progress (fallback)
func createTarSimple(source, tarPath string, isDir bool) error {
	var tarCmd *exec.Cmd

	if isDir {
		parentDir := filepath.Dir(source)
		baseName := filepath.Base(source)
		tarCmd = exec.Command("tar", "-czf", tarPath, "-C", parentDir, baseName)
	} else {
		parentDir := filepath.Dir(source)
		fileName := filepath.Base(source)
		tarCmd = exec.Command("tar", "-czf", tarPath, "-C", parentDir, fileName)
	}

	if shipVerbose {
		tarCmd.Stdout = os.Stdout
		tarCmd.Stderr = os.Stderr
	}

	return tarCmd.Run()
}

// rsyncFileWithProgress transfers file via rsync with progress display
func rsyncFileWithProgress(localPath, sshTarget, remotePath string, fileSize int64) error {
	args := []string{
		"-avz",           // archive, verbose, compress
		"--progress",     // show progress
		"--human-readable",
		localPath,
		fmt.Sprintf("%s:%s", sshTarget, remotePath),
	}

	rsyncCmd := exec.Command("rsync", args...)

	// Create pipes for stdout to parse progress
	stdout, err := rsyncCmd.StdoutPipe()
	if err != nil {
		// Fallback to simple rsync
		return rsyncFileSimple(localPath, sshTarget, remotePath)
	}
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Start(); err != nil {
		return err
	}

	// Parse rsync output for progress
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		// rsync progress lines contain percentage like "  1,234,567 100%   10.5MB/s"
		if strings.Contains(line, "%") {
			// Extract and display the progress
			fmt.Printf("\r  %s %s", theme.DimTextStyle.Render("Transfer:"), line)
		}
	}

	err = rsyncCmd.Wait()
	fmt.Print("\r                                                                              \r") // Clear progress line
	return err
}

// rsyncFileSimple is a fallback without progress parsing
func rsyncFileSimple(localPath, sshTarget, remotePath string) error {
	args := []string{
		"-avz",
		"--progress",
		localPath,
		fmt.Sprintf("%s:%s", sshTarget, remotePath),
	}

	rsyncCmd := exec.Command("rsync", args...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	return rsyncCmd.Run()
}
