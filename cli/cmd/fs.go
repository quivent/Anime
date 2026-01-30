package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/vfs"
	"github.com/spf13/cobra"
)

// Flags
var (
	fsRecursive bool
	fsForce     bool
	fsAll       bool
	fsLong      bool
	fsParents   bool
	fsDepth     int
)

// fsCmd is the parent command for VFS operations
var fsCmd = &cobra.Command{
	Use:     "fs",
	Aliases: []string{"vfs"},
	Short:   "Embedded virtual filesystem",
	Long: `Virtual filesystem embedded directly in the anime binary.

The VFS is a self-contained filesystem that lives INSIDE the binary.
All data persists when you copy the binary to another server.

STORAGE:
  - Data is stored in-memory during runtime
  - Use 'anime fs save' to persist changes INTO the binary itself
  - The binary becomes a portable capsule with all your files

COMMANDS:
  ls, mkdir, touch, cat, rm, cp, mv, tree, find, grep, pwd, cd
  import, export - transfer files between VFS and real filesystem
  save - persist filesystem INTO the binary (self-modifying)
  stats - show filesystem statistics

EXAMPLES:
  anime ls                      List root directory
  anime mkdir /projects         Create a directory
  anime touch /hello.txt        Create an empty file
  anime cat /hello.txt          View file contents
  anime cp -r /real/path .      Import from real filesystem
  anime fs save                 Save VFS into binary`,
	Run: func(cmd *cobra.Command, args []string) {
		showVFSDashboard()
	},
}

func showVFSDashboard() {
	fs := vfs.Get()

	fmt.Println()
	fmt.Println(theme.RenderBanner("EMBEDDED FILESYSTEM"))
	fmt.Println()

	stats := fs.Stats()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("CWD:"), theme.HighlightStyle.Render(stats["cwd"].(string)))
	fmt.Printf("  %s %d files, %d directories\n",
		theme.DimTextStyle.Render("Contents:"),
		stats["files"].(int),
		stats["dirs"].(int))
	fmt.Printf("  %s %s\n",
		theme.DimTextStyle.Render("Size:"),
		vfsFormatBytes(stats["total_size"].(int64)))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Quick Commands:"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime ls", "List directory contents"},
		{"anime mkdir DIR", "Create directory"},
		{"anime touch FILE", "Create/update file"},
		{"anime cat FILE", "View file contents"},
		{"anime tree", "Show tree view"},
		{"anime fs save", "Persist to binary"},
	}

	for _, c := range commands {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-20s", c.cmd)),
			theme.DimTextStyle.Render(c.desc))
	}
	fmt.Println()
}

// lsCmd lists directory contents
var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List VFS directory contents",
	Long: `List contents of the embedded virtual filesystem.

Examples:
  anime ls           List current directory
  anime ls /         List root directory
  anime ls -l        Long format with details
  anime ls -a        Show all (currently same as ls)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLs,
}

func runLs(cmd *cobra.Command, args []string) error {
	fs := vfs.Get()

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	entries, err := fs.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ls: %s: %w", path, err)
	}

	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if fsLong {
			// Long format
			typeChar := "-"
			if entry.Type == vfs.TypeDir {
				typeChar = "d"
			} else if entry.Type == vfs.TypeSymlink {
				typeChar = "l"
			}

			modeStr := fmt.Sprintf("%s%o", typeChar, entry.Mode)
			sizeStr := vfsFormatBytes(entry.Size)
			timeStr := entry.ModTime.Format("Jan 02 15:04")

			name := entry.Name
			if entry.Type == vfs.TypeDir {
				name = theme.InfoStyle.Render(name + "/")
			}

			fmt.Printf("%s %8s %s %s\n", modeStr, sizeStr, timeStr, name)
		} else {
			// Simple format
			if entry.Type == vfs.TypeDir {
				fmt.Printf("%s/\n", theme.InfoStyle.Render(entry.Name))
			} else {
				fmt.Println(entry.Name)
			}
		}
	}

	return nil
}

// mkdirCmd creates directories
var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "Create VFS directory",
	Long: `Create a directory in the embedded filesystem.

Examples:
  anime mkdir projects           Create 'projects' in current dir
  anime mkdir /data/logs         Create nested path (with -p)
  anime mkdir -p /a/b/c          Create parent directories`,
	Args: cobra.ExactArgs(1),
	RunE: runMkdir,
}

func runMkdir(cmd *cobra.Command, args []string) error {
	fs := vfs.Get()
	path := args[0]

	var err error
	if fsParents {
		err = fs.MkdirAll(path)
	} else {
		err = fs.Mkdir(path)
	}
	if err != nil {
		return err
	}
	return vfs.AutoSave()
}

// touchCmd creates/updates files
var touchCmd = &cobra.Command{
	Use:   "touch <file>",
	Short: "Create empty file or update timestamp",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vfs.Get().Touch(args[0]); err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// catCmd displays file contents
var catCmd = &cobra.Command{
	Use:   "cat <file>",
	Short: "Display VFS file contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := vfs.Get().ReadFile(args[0])
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	},
}

// rmCmd removes files/directories
var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Remove VFS file or directory",
	Long: `Remove a file or directory from the embedded filesystem.

Examples:
  anime rm file.txt         Remove a file
  anime rm -r directory     Remove directory recursively
  anime rm -rf /temp        Force remove everything`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()
		path := args[0]

		var err error
		if fsRecursive {
			err = fs.RemoveAll(path)
		} else {
			err = fs.Remove(path)
		}
		if err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// cpCmd copies files
var cpCmd = &cobra.Command{
	Use:   "cp <source> <dest>",
	Short: "Copy files in VFS or import from real filesystem",
	Long: `Copy files within VFS or import from real filesystem.

If source is a real filesystem path (starts with ./ or / and exists on disk),
it will be imported into VFS. Otherwise, copy within VFS.

Examples:
  anime cp file.txt copy.txt        Copy within VFS
  anime cp -r dir1 dir2             Copy directory
  anime cp ./real-file.txt /vfs/    Import real file to VFS
  anime cp -r ./project /code       Import directory to VFS`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()
		src, dst := args[0], args[1]

		var err error
		// Check if source is a real filesystem path
		if isRealPath(src) {
			info, statErr := os.Stat(src)
			if statErr == nil {
				// Import from real filesystem
				if info.IsDir() {
					if !fsRecursive {
						return fmt.Errorf("cp: %s is a directory (use -r)", src)
					}
					err = fs.ImportDir(src, dst)
				} else {
					err = fs.ImportFile(src, dst)
				}
				if err != nil {
					return err
				}
				return vfs.AutoSave()
			}
		}

		// Copy within VFS
		if fsRecursive {
			err = fs.CopyAll(src, dst)
		} else {
			err = fs.Copy(src, dst)
		}
		if err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// mvCmd moves/renames files
var mvCmd = &cobra.Command{
	Use:   "mv <source> <dest>",
	Short: "Move/rename VFS file or directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vfs.Get().Rename(args[0], args[1]); err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// pwdCmd prints working directory
var pwdCmd = &cobra.Command{
	Use:   "pwd",
	Short: "Print VFS working directory",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(vfs.Get().Cwd())
	},
}

// cdCmd changes directory
var cdCmd = &cobra.Command{
	Use:   "cd [path]",
	Short: "Change VFS working directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}
		if err := vfs.Get().Cd(path); err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// vfsTreeCmd shows tree view (named differently to avoid conflict with cmd/tree.go)
var vfsTreeCmd = &cobra.Command{
	Use:     "vtree [path]",
	Aliases: []string{"fstree"},
	Short:   "Display VFS tree structure",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fs := vfs.Get()
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}

		depth := fsDepth
		if depth <= 0 {
			depth = 10
		}

		fmt.Print(fs.Tree(path, depth))
	},
}

// findCmd searches for files
var findCmd = &cobra.Command{
	Use:   "find <pattern>",
	Short: "Find files matching pattern in VFS",
	Long: `Find files matching a glob pattern.

Examples:
  anime find "*.txt"         Find all .txt files
  anime find "config*"       Find files starting with 'config'`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		matches := vfs.Get().Find("/", args[0])
		for _, m := range matches {
			fmt.Println(m)
		}
	},
}

// grepCmd searches file contents
var grepCmd = &cobra.Command{
	Use:   "grep <pattern> [path]",
	Short: "Search file contents in VFS",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		fs := vfs.Get()
		pattern := args[0]
		path := "/"
		if len(args) > 1 {
			path = args[1]
		}

		results := fs.Grep(path, pattern)
		for file, matches := range results {
			for _, match := range matches {
				fmt.Printf("%s:%s\n", theme.HighlightStyle.Render(file), match)
			}
		}
	},
}

// duCmd shows disk usage
var duCmd = &cobra.Command{
	Use:   "du [path]",
	Short: "Show VFS disk usage",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fs := vfs.Get()
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}

		size := fs.DiskUsage(path)
		fmt.Printf("%s\t%s\n", formatBytes(size), path)
	},
}

// viCmd opens file in editor
var viCmd = &cobra.Command{
	Use:   "vi <file>",
	Short: "Edit VFS file with vi/vim",
	Long: `Edit a file in the embedded filesystem using vi.

The file is exported to a temp location, opened in vi,
and imported back after editing.`,
	Args: cobra.ExactArgs(1),
	RunE: runViEdit,
}

func runViEdit(cmd *cobra.Command, args []string) error {
	fs := vfs.Get()
	vfsPath := args[0]

	// Ensure parent directory exists
	absPath := fs.AbsPath(vfsPath)

	// Create file if doesn't exist
	if !fs.Exists(absPath) {
		if err := fs.Touch(absPath); err != nil {
			return err
		}
	}

	// Export to temp file
	tmpFile, err := os.CreateTemp("", "anime-vfs-*.txt")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Export current content
	content, _ := fs.ReadFile(absPath)
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return err
	}

	// Find editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Run editor
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Import back
	newContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}

	if err := fs.WriteFile(absPath, newContent); err != nil {
		return err
	}
	return vfs.AutoSave()
}

// echoCmd writes to a file
var echoCmd = &cobra.Command{
	Use:   "echo <text> [> file]",
	Short: "Write text to VFS file",
	Long: `Write text to a file in the embedded filesystem.

Examples:
  anime echo "Hello World" > /hello.txt
  anime echo "More text" >> /hello.txt    (append)`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()

		// Parse arguments for redirection
		text := ""
		var targetFile string
		append := false

		for i, arg := range args {
			if arg == ">" && i+1 < len(args) {
				targetFile = args[i+1]
				break
			} else if arg == ">>" && i+1 < len(args) {
				targetFile = args[i+1]
				append = true
				break
			} else if strings.HasPrefix(arg, ">") {
				// Handle ">file" without space
				if strings.HasPrefix(arg, ">>") {
					targetFile = strings.TrimPrefix(arg, ">>")
					append = true
				} else {
					targetFile = strings.TrimPrefix(arg, ">")
				}
				break
			} else {
				if text != "" {
					text += " "
				}
				text += arg
			}
		}

		text += "\n"

		if targetFile == "" {
			fmt.Print(text)
			return nil
		}

		var err error
		if append {
			err = fs.AppendFile(targetFile, []byte(text))
		} else {
			err = fs.WriteFile(targetFile, []byte(text))
		}
		if err != nil {
			return err
		}
		return vfs.AutoSave()
	},
}

// fsSaveCmd persists VFS to binary
var fsSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save VFS state into the binary",
	Long: `Persist the virtual filesystem INTO the anime binary itself.

This rewrites the binary with the current VFS state appended.
When you copy the binary to another server, all your files come with it.

WARNING: This modifies the anime binary!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()

		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("Saving VFS to binary..."))

		if err := fs.SaveToSelf(); err != nil {
			return fmt.Errorf("save failed: %w", err)
		}

		stats := fs.Stats()
		fmt.Printf("  %s %d files, %d dirs, %s\n",
			theme.SuccessStyle.Render("Saved:"),
			stats["files"].(int),
			stats["dirs"].(int),
			vfsFormatBytes(stats["total_size"].(int64)))
		fmt.Println()

		return nil
	},
}

// fsStatsCmd shows filesystem stats
var fsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show VFS statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats := vfs.Get().Stats()

		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("VFS Statistics:"))
		fmt.Println()
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Working Dir:"), theme.HighlightStyle.Render(stats["cwd"].(string)))
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Files:"), stats["files"].(int))
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Directories:"), stats["dirs"].(int))
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Total Size:"), vfsFormatBytes(stats["total_size"].(int64)))
		fmt.Println()
	},
}

// fsResetCmd clears the filesystem
var fsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset VFS to empty state",
	Long:  `Clear all files and directories from the virtual filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !fsForce {
			fmt.Print("Reset VFS? All data will be lost. [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		vfs.Get().Reset()
		fmt.Println(theme.SuccessStyle.Render("VFS reset to empty state"))
		return vfs.AutoSave()
	},
}

// fsExportCmd exports VFS to real filesystem
var fsExportCmd = &cobra.Command{
	Use:   "export <vfs-path> <real-path>",
	Short: "Export VFS content to real filesystem",
	Long: `Export files from the embedded VFS to the real filesystem.

Examples:
  anime fs export /config ./config-backup
  anime fs export / ./vfs-dump`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()
		vfsPath, realPath := args[0], args[1]

		if fs.IsDir(vfsPath) {
			return fs.ExportDir(vfsPath, realPath)
		}
		return fs.ExportFile(vfsPath, realPath)
	},
}

// fsImportCmd imports from real filesystem
var fsImportCmd = &cobra.Command{
	Use:   "import <real-path> <vfs-path>",
	Short: "Import real filesystem content to VFS",
	Long: `Import files from the real filesystem into the embedded VFS.

Examples:
  anime fs import ./config /config
  anime fs import ~/projects /projects`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := vfs.Get()
		realPath, vfsPath := args[0], args[1]

		info, err := os.Stat(realPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := fs.ImportDir(realPath, vfsPath); err != nil {
				return err
			}
		} else {
			if err := fs.ImportFile(realPath, vfsPath); err != nil {
				return err
			}
		}
		return vfs.AutoSave()
	},
}

// Helper functions

func isRealPath(path string) bool {
	// Check if path exists on real filesystem
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		_, err := os.Stat(path)
		return err == nil
	}
	// Absolute paths starting with / could be real or VFS
	// Check if it exists on real FS
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") {
		_, err := os.Stat(path)
		return err == nil
	}
	return false
}

func vfsFormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func init() {
	// ls flags
	lsCmd.Flags().BoolVarP(&fsLong, "long", "l", false, "Long format")
	lsCmd.Flags().BoolVarP(&fsAll, "all", "a", false, "Show all files")

	// mkdir flags
	mkdirCmd.Flags().BoolVarP(&fsParents, "parents", "p", false, "Create parent directories")

	// rm flags
	rmCmd.Flags().BoolVarP(&fsRecursive, "recursive", "r", false, "Remove recursively")
	rmCmd.Flags().BoolVarP(&fsForce, "force", "f", false, "Force removal")

	// cp flags
	cpCmd.Flags().BoolVarP(&fsRecursive, "recursive", "r", false, "Copy recursively")

	// tree flags
	vfsTreeCmd.Flags().IntVarP(&fsDepth, "depth", "d", 10, "Maximum depth")

	// reset flags
	fsResetCmd.Flags().BoolVarP(&fsForce, "force", "f", false, "Skip confirmation")

	// Add subcommands to fs
	fsCmd.AddCommand(fsSaveCmd)
	fsCmd.AddCommand(fsStatsCmd)
	fsCmd.AddCommand(fsResetCmd)
	fsCmd.AddCommand(fsExportCmd)
	fsCmd.AddCommand(fsImportCmd)

	// Add all commands to root
	rootCmd.AddCommand(fsCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(touchCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(pwdCmd)
	rootCmd.AddCommand(cdCmd)
	rootCmd.AddCommand(vfsTreeCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(duCmd)
	rootCmd.AddCommand(viCmd)
	rootCmd.AddCommand(echoCmd)
}
