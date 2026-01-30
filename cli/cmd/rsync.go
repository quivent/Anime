package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	rsyncDelete   bool
	rsyncDryRun   bool
	rsyncExclude  []string
	rsyncCompress bool
	rsyncVerbose  bool
)

var rsyncCmd = &cobra.Command{
	Use:   "rsync <source> <dest>",
	Short: "Rsync files with server alias support",
	Long: `Rsync files to/from servers using aliases.

Server aliases are automatically resolved. Use server:path syntax.

EXAMPLES:
  anime rsync ./folder lambda:~/backup     # Upload folder to lambda
  anime rsync lambda:~/data ./local        # Download from lambda
  anime rsync -n ./src alice:~/project     # Dry run
  anime rsync --delete ./www lambda:/var   # Sync with delete
  anime rsync -e node_modules ./app srv:~  # Exclude pattern

FLAGS:
  -n, --dry-run     Preview without changes
  -d, --delete      Delete extraneous files from dest
  -e, --exclude     Exclude pattern (can use multiple)
  -z, --compress    Compress during transfer
  -v, --verbose     Verbose output`,
	Args: cobra.ExactArgs(2),
	RunE: runRsync,
}

func init() {
	rsyncCmd.Flags().BoolVarP(&rsyncDryRun, "dry-run", "n", false, "Preview without changes")
	rsyncCmd.Flags().BoolVarP(&rsyncDelete, "delete", "d", false, "Delete extraneous files from dest")
	rsyncCmd.Flags().StringArrayVarP(&rsyncExclude, "exclude", "e", nil, "Exclude pattern")
	rsyncCmd.Flags().BoolVarP(&rsyncCompress, "compress", "z", true, "Compress during transfer")
	rsyncCmd.Flags().BoolVarP(&rsyncVerbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(rsyncCmd)
}

func runRsync(cmd *cobra.Command, args []string) error {
	source := args[0]
	dest := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve aliases in source and dest
	source, err = resolveRsyncPath(cfg, source)
	if err != nil {
		return err
	}
	dest, err = resolveRsyncPath(cfg, dest)
	if err != nil {
		return err
	}

	// Build rsync command
	rsyncArgs := []string{"-av", "--progress"}

	if rsyncCompress {
		rsyncArgs = append(rsyncArgs, "-z")
	}

	if rsyncDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
		fmt.Println(theme.WarningStyle.Render("DRY RUN - no changes will be made"))
		fmt.Println()
	}

	if rsyncDelete {
		rsyncArgs = append(rsyncArgs, "--delete")
	}

	if rsyncVerbose {
		rsyncArgs = append(rsyncArgs, "-v")
	}

	// Add excludes
	for _, exc := range rsyncExclude {
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
	)

	rsyncArgs = append(rsyncArgs, source, dest)

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("From:"), theme.InfoStyle.Render(source))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(dest))
	fmt.Println()

	rsyncExec := exec.Command("rsync", rsyncArgs...)
	rsyncExec.Stdin = os.Stdin
	rsyncExec.Stdout = os.Stdout
	rsyncExec.Stderr = os.Stderr

	if err := rsyncExec.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Println()
	if rsyncDryRun {
		fmt.Println(theme.WarningStyle.Render("Dry run complete"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("Rsync complete"))
	}

	return nil
}

// resolveRsyncPath resolves server aliases in rsync paths
// Handles: server:path, user@server:path, or local paths
func resolveRsyncPath(cfg *config.Config, path string) (string, error) {
	// Check if it contains a colon (remote path)
	if !strings.Contains(path, ":") {
		// Local path, return as-is
		return path, nil
	}

	// Split on first colon
	parts := strings.SplitN(path, ":", 2)
	if len(parts) != 2 {
		return path, nil
	}

	serverPart := parts[0]
	pathPart := parts[1]

	// If already has @, it's a full user@host
	if strings.Contains(serverPart, "@") {
		return path, nil
	}

	// Try to resolve alias
	if alias := cfg.GetAlias(serverPart); alias != "" {
		if !strings.Contains(alias, "@") {
			alias = "ubuntu@" + alias
		}
		return alias + ":" + pathPart, nil
	}

	// Try server config
	if server, err := cfg.GetServer(serverPart); err == nil {
		return fmt.Sprintf("%s@%s:%s", server.User, server.Host, pathPart), nil
	}

	// Check SSH config by trying to resolve
	if resolved, err := trySSHConfigResolve(serverPart); err == nil {
		return resolved + ":" + pathPart, nil
	}

	// If it looks like an IP/hostname, add ubuntu@
	if strings.Contains(serverPart, ".") {
		return "ubuntu@" + serverPart + ":" + pathPart, nil
	}

	return "", fmt.Errorf("could not resolve server: %s (use 'anime set %s <ip>' to configure)", serverPart, serverPart)
}

// trySSHConfigResolve tries to resolve a host through SSH config
func trySSHConfigResolve(host string) (string, error) {
	sshCmd := exec.Command("ssh", "-G", host)
	output, err := sshCmd.Output()
	if err != nil {
		return "", err
	}

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

	if hostname != "" && hostname != host {
		if user != "" {
			return user + "@" + hostname, nil
		}
		return "ubuntu@" + hostname, nil
	}

	return "", fmt.Errorf("not found in SSH config")
}
