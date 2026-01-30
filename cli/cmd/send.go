package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	sendDryRun   bool
	sendDelete   bool
	sendVerbose  bool
	sendExclude  []string
	sendNoCompress bool
)

var sendCmd = &cobra.Command{
	Use:   "send <file_or_folder> [server] [remote_path]",
	Short: "Send file or folder to remote server via rsync",
	Long: `Send a file or folder to a remote server.

Unlike 'anime push' which uses git, this command uses rsync to transfer
ALL files including uncommitted changes and untracked files.

Optimizations included:
  - Compression during transfer (-z)
  - Delta transfers (only changed parts)
  - Preserves permissions and timestamps
  - Excludes common junk (.git, node_modules, __pycache__, etc.)

Server formats:
  - alias            (from anime config or .ssh/config)
  - user@IP          (e.g., ubuntu@192.168.1.100)
  - IP               (defaults to ubuntu@IP)

Examples:
  anime send .                            # Send cwd to default server:~/anime
  anime send ./src lambda                 # Send ./src to lambda:~/anime
  anime send myfile.txt alice             # Send file to alice:~/anime
  anime send . alice ~/project            # Send cwd to alice:~/project
  anime send --dry-run . lambda           # Preview what would be sent
  anime send --delete . lambda            # Delete extraneous files on remote
  anime send -e "*.log" ./data server     # Exclude additional patterns
`,
	Args: cobra.RangeArgs(1, 3),
	RunE: runSend,
}

func init() {
	sendCmd.Flags().BoolVarP(&sendDryRun, "dry-run", "n", false, "Preview without making changes")
	sendCmd.Flags().BoolVarP(&sendDelete, "delete", "d", false, "Delete extraneous files from destination")
	sendCmd.Flags().BoolVarP(&sendVerbose, "verbose", "v", false, "Verbose output")
	sendCmd.Flags().StringArrayVarP(&sendExclude, "exclude", "e", nil, "Additional exclude patterns")
	sendCmd.Flags().BoolVar(&sendNoCompress, "no-compress", false, "Disable compression")

	rootCmd.AddCommand(sendCmd)
}

func runSend(cmd *cobra.Command, args []string) error {
	// First argument is required: file or folder to send
	source := args[0]

	// Verify source exists
	sourceInfo, err := os.Stat(source)
	if os.IsNotExist(err) {
		return fmt.Errorf("source does not exist: %s", source)
	}
	if err != nil {
		return fmt.Errorf("cannot access source: %w", err)
	}

	// Get default server from config
	cfg, _ := config.Load()
	server := cfg.GetDefaultServer()
	remotePath := "~/anime"

	if len(args) > 1 {
		server = args[1]
	}
	if len(args) > 2 {
		remotePath = args[2]
	}

	// Parse server argument
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	isFile := !sourceInfo.IsDir()

	sourceType := "directory"
	if isFile {
		sourceType = "file"
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📤 Sending %s to remote server", sourceType)))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.InfoStyle.Render(source))
	fmt.Printf("  %s %s:%s\n", theme.DimTextStyle.Render("Dest:"), theme.InfoStyle.Render(target), remotePath)
	fmt.Println()

	if sendDryRun {
		fmt.Println(theme.WarningStyle.Render("DRY RUN - no changes will be made"))
		fmt.Println()
	}

	// Build rsync command with optimizations
	rsyncArgs := []string{
		"-av",          // Archive mode + verbose
		"--progress",   // Show progress
		"--human-readable",
	}

	// Compression (enabled by default)
	if !sendNoCompress {
		rsyncArgs = append(rsyncArgs, "-z")
	}

	// Dry run
	if sendDryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}

	// Delete extraneous files
	if sendDelete {
		rsyncArgs = append(rsyncArgs, "--delete")
	}

	// Extra verbose
	if sendVerbose {
		rsyncArgs = append(rsyncArgs, "-v")
	}

	// Standard excludes - common junk and large build artifacts
	standardExcludes := []string{
		// Version control
		".git",
		".svn",
		".hg",
		// Dependencies
		"node_modules",
		".venv",
		"venv",
		"vendor",
		// Python
		"__pycache__",
		"*.pyc",
		"*.pyo",
		".pytest_cache",
		".mypy_cache",
		"*.egg-info",
		// Build outputs
		"build",
		"dist",
		"target",
		"out",
		"bin",
		"*.o",
		"*.a",
		"*.so",
		"*.dylib",
		// Archives
		"*.tar.gz",
		"*.tar",
		"*.zip",
		"*.tgz",
		// Large model/data files
		"*.pt",
		"*.pth",
		"*.ckpt",
		"*.safetensors",
		"*.bin",
		"*.onnx",
		"*.gguf",
		"*.ggml",
		"*.h5",
		"*.hdf5",
		"*.pkl",
		"*.pickle",
		"models",
		"checkpoints",
		"weights",
		// Media
		"*.mp4",
		"*.mov",
		"*.avi",
		"*.mkv",
		"*.wav",
		"*.mp3",
		// IDE/Editor
		".DS_Store",
		".idea",
		".vscode",
		"*.swp",
		"*.swo",
		// Test/Coverage
		"coverage.*",
		"*.coverage",
		".coverage",
		"htmlcov",
		// Logs
		"*.log",
		"logs",
	}

	for _, exc := range standardExcludes {
		rsyncArgs = append(rsyncArgs, "--exclude", exc)
	}

	// User excludes
	for _, exc := range sendExclude {
		rsyncArgs = append(rsyncArgs, "--exclude", exc)
	}

	// Source and destination
	// For directories: trailing slash copies contents into remote path
	// For files: no trailing slash, file is copied to remote path
	if isFile {
		rsyncArgs = append(rsyncArgs, source, fmt.Sprintf("%s:%s", target, remotePath))
	} else {
		rsyncArgs = append(rsyncArgs, source+"/", fmt.Sprintf("%s:%s", target, remotePath))
	}

	// Show command
	fmt.Printf("%s rsync %v\n", theme.DimTextStyle.Render("$"), rsyncArgs)
	fmt.Println()

	// Create remote directory first
	mkdirCmd := exec.Command("ssh", target, fmt.Sprintf("mkdir -p %s", remotePath))
	mkdirCmd.Stderr = os.Stderr
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Execute rsync
	rsyncExec := exec.Command("rsync", rsyncArgs...)
	rsyncExec.Stdin = os.Stdin
	rsyncExec.Stdout = os.Stdout
	rsyncExec.Stderr = os.Stderr

	if err := rsyncExec.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Println()
	if sendDryRun {
		fmt.Println(theme.WarningStyle.Render("Dry run complete - no changes made"))
	} else {
		fmt.Printf("%s %s sent to %s:%s\n", theme.SuccessStyle.Render("✓"), sourceType, target, remotePath)
	}

	return nil
}
