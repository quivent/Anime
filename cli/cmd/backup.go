package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var (
	backupServer string
	backupOutput string
)

var backupCmd = &cobra.Command{
	Use:   "backup <path>",
	Short: "Create a timestamped backup archive",
	Long: `Create a compressed tar backup of a directory.
On remote servers, the backup is created remotely then downloaded.

Examples:
  anime backup ~/myapp                            # Local backup
  anime backup /home/ubuntu/api -s wings          # Backup from remote
  anime backup ~/data -s wings -o ./backups/      # Download to specific dir`,
	Args: cobra.ExactArgs(1),
	RunE: runBackup,
}

func init() {
	backupCmd.Flags().StringVarP(&backupServer, "server", "s", "", "Remote server")
	backupCmd.Flags().StringVarP(&backupOutput, "output", "o", ".", "Local directory to save backup")
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	remotePath := args[0]
	if err := validate.ShellSafe(remotePath); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102-150405")
	baseName := filepath.Base(remotePath)
	if baseName == "." || baseName == "/" {
		baseName = "backup"
	}
	archiveName := fmt.Sprintf("%s-%s.tar.gz", baseName, timestamp)

	fmt.Println(theme.RenderBanner("💾 BACKUP 💾"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.HighlightStyle.Render(remotePath))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Archive:"), theme.HighlightStyle.Render(archiveName))

	if backupServer == "" {
		// Local backup
		if err := os.MkdirAll(backupOutput, 0755); err != nil {
			return fmt.Errorf("cannot create output dir: %w", err)
		}
		outPath := filepath.Join(backupOutput, archiveName)
		fmt.Printf("  %s Creating archive...\n", theme.SymbolLoading)

		tarCmd := exec.Command("tar", "-czf", outPath, "-C", filepath.Dir(remotePath), baseName)
		tarCmd.Stderr = os.Stderr
		start := time.Now()
		if err := tarCmd.Run(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
		elapsed := time.Since(start)

		info, _ := os.Stat(outPath)
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("  %s %s (%.1f MB) in %s\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess),
			outPath, sizeMB, elapsed.Round(time.Second))
		return nil
	}

	// Remote backup: create archive on server, then download
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var user, host, sshKey string
	target := cfg.GetAlias(backupServer)
	if target != "" {
		if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			user = parts[0]
			host = parts[1]
		} else {
			user = "ubuntu"
			host = target
		}
	} else {
		server, err := cfg.GetServer(backupServer)
		if err != nil {
			return fmt.Errorf("server not found: %s", backupServer)
		}
		user = server.User
		host = server.Host
		sshKey = server.SSHKey
	}

	sshTarget := fmt.Sprintf("%s@%s", user, host)
	remoteTar := fmt.Sprintf("/tmp/%s", archiveName)

	// Create archive on remote
	fmt.Printf("  %s Creating archive on %s...\n", theme.SymbolLoading, backupServer)
	client, err := ssh.NewClient(host, user, sshKey)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	createScript := fmt.Sprintf(`tar -czf "%s" -C "$(dirname "%s")" "$(basename "%s")" && ls -lh "%s" | awk '{print $5}'`,
		remoteTar, remotePath, remotePath, remoteTar)
	output, err := client.RunCommand(createScript)
	if err != nil {
		return fmt.Errorf("remote backup failed: %s", output)
	}
	fmt.Printf("  %s Archive created (%s)\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))

	// Download
	if err := os.MkdirAll(backupOutput, 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}
	localPath := filepath.Join(backupOutput, archiveName)
	fmt.Printf("  %s Downloading...\n", theme.SymbolLoading)

	scpArgs := []string{}
	if sshKey != "" {
		scpArgs = append(scpArgs, "-i", sshKey)
	}
	scpArgs = append(scpArgs, fmt.Sprintf("%s:%s", sshTarget, remoteTar), localPath)
	scpCmd := exec.Command("scp", scpArgs...)
	scpCmd.Stderr = os.Stderr
	if err := scpCmd.Run(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Cleanup remote
	client.RunCommand(fmt.Sprintf(`rm -f "%s"`, remoteTar))

	info, _ := os.Stat(localPath)
	sizeMB := float64(info.Size()) / (1024 * 1024)
	fmt.Printf("  %s %s (%.1f MB)\n",
		theme.SuccessStyle.Render(theme.SymbolSuccess),
		localPath, sizeMB)
	fmt.Println()
	return nil
}
