package cmd

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

//go:embed capsules
var embeddedCapsules embed.FS

var capsulesCmd = &cobra.Command{
	Use:   "capsules",
	Short: "Manage embedded CLI packages",
	Long: `Manage CLI packages that are embedded in the anime binary.

Capsules are other CLI projects (like producer, coverage) that can be
embedded into the anime binary and installed on remote servers.

The workflow:
  1. anime capsules add <name> <path>   # Register a CLI project
  2. make build                          # Build with capsules embedded
  3. anime deploy <server>               # Deploy to server
  4. anime capsules install              # Install capsules on server

Examples:
  anime capsules add producer ~/CLIs/producer
  anime capsules add coverage ~/CLIs/coverage --build "go install"
  anime capsules list
  anime capsules remove producer
  anime capsules install                 # Install all on server
  anime capsules install producer        # Install specific capsule
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var capsulesAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Add a CLI project as a capsule",
	Long: `Register a CLI project to be embedded as a capsule.

The path should point to a git repository containing a CLI project.
On build, the .git directory will be embedded into the anime binary.

Examples:
  anime capsules add producer ~/CLIs/producer
  anime capsules add coverage ~/CLIs/coverage
  anime capsules add mycli /path/to/mycli --build "cargo build --release"
`,
	Args: cobra.ExactArgs(2),
	RunE: runCapsulesAdd,
}

var capsulesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered capsule",
	Args:  cobra.ExactArgs(1),
	RunE:  runCapsulesRemove,
}

var capsulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered capsules",
	RunE:  runCapsulesList,
}

var capsulesInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install capsules on the current machine",
	Long: `Extract and install embedded capsules.

Without arguments, installs all embedded capsules.
With a name argument, installs only that capsule.

This command:
  1. Extracts the capsule's .git directory
  2. Checks out the working tree
  3. Runs the build command (default: make install)

Examples:
  anime capsules install              # Install all capsules
  anime capsules install producer     # Install only producer
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCapsulesInstall,
}

var capsulesPackCmd = &cobra.Command{
	Use:    "pack",
	Short:  "Pack capsules for embedding (used by Makefile)",
	Hidden: true,
	RunE:   runCapsulesPack,
}

var capsulesDeployCmd = &cobra.Command{
	Use:   "deploy [server]",
	Short: "Deploy and install capsules on remote server",
	Long: `Transfer capsule archives to a remote server and install them.

This is a quick way to deploy capsules without redeploying the entire anime binary.
It transfers the tar.gz files from cmd/capsules/ and runs 'anime capsules install' remotely.

Server formats (same as anime deploy):
  - user@IP          (e.g., ubuntu@192.168.1.100)
  - IP               (defaults to ubuntu@IP)
  - alias            (from anime config or .ssh/config)
  - (default: lambda if no server specified)

Examples:
  anime capsules deploy                    # Deploy to default server (lambda)
  anime capsules deploy my-server          # Deploy to specific server
  anime capsules deploy ubuntu@10.0.0.1    # Deploy with explicit user
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCapsulesDeploy,
}

var (
	capsuleBuildCmd string
	capsuleBinary   string
)

func init() {
	capsulesAddCmd.Flags().StringVar(&capsuleBuildCmd, "build", "", "Build command (default: make install)")
	capsulesAddCmd.Flags().StringVar(&capsuleBinary, "binary", "", "Binary name if different from capsule name")

	capsulesCmd.AddCommand(capsulesAddCmd)
	capsulesCmd.AddCommand(capsulesRemoveCmd)
	capsulesCmd.AddCommand(capsulesListCmd)
	capsulesCmd.AddCommand(capsulesInstallCmd)
	capsulesCmd.AddCommand(capsulesPackCmd)
	capsulesCmd.AddCommand(capsulesDeployCmd)

	rootCmd.AddCommand(capsulesCmd)
}

func runCapsulesAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	path := args[1]

	// Expand path
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	// Verify path exists and is a git repo
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("%s is not a git repository (no .git directory)", absPath)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	capsule := config.Capsule{
		Name:     name,
		Path:     path,
		BuildCmd: capsuleBuildCmd,
		Binary:   capsuleBinary,
	}

	if err := cfg.AddCapsule(capsule); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s Added capsule '%s' from %s\n", theme.SuccessStyle.Render("✓"), name, absPath)
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Run 'make build' to embed capsules into the binary"))

	return nil
}

func runCapsulesRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteCapsule(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s Removed capsule '%s'\n", theme.SuccessStyle.Render("✓"), name)

	return nil
}

func runCapsulesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	capsules := cfg.ListCapsules()

	if len(capsules) == 0 {
		fmt.Println(theme.DimTextStyle.Render("No capsules registered"))
		fmt.Println()
		fmt.Println("Add a capsule with:")
		fmt.Println("  anime capsules add <name> <path>")
		return nil
	}

	fmt.Println(theme.InfoStyle.Render("Registered Capsules"))
	fmt.Println()

	for _, cap := range capsules {
		expandedPath := cap.GetExpandedPath()
		exists := "✓"
		if _, err := os.Stat(filepath.Join(expandedPath, ".git")); os.IsNotExist(err) {
			exists = "✗"
		}

		fmt.Printf("  %s %s\n", exists, theme.PrimaryTextStyle.Render(cap.Name))
		fmt.Printf("    Path:  %s\n", cap.Path)
		fmt.Printf("    Build: %s\n", cap.GetBuildCommand())
		if cap.Binary != "" {
			fmt.Printf("    Binary: %s\n", cap.Binary)
		}
		fmt.Println()
	}

	// Check for embedded capsules
	embedded := listEmbeddedCapsules()
	if len(embedded) > 0 {
		fmt.Println(theme.InfoStyle.Render("Embedded Capsules"))
		fmt.Println()
		for _, name := range embedded {
			fmt.Printf("  ✓ %s\n", name)
		}
		fmt.Println()
	}

	return nil
}

func runCapsulesInstall(cmd *cobra.Command, args []string) error {
	var capsuleName string
	if len(args) > 0 {
		capsuleName = args[0]
	}

	embedded := listEmbeddedCapsules()
	if len(embedded) == 0 {
		fmt.Println(theme.WarningStyle.Render("No capsules embedded in this binary"))
		fmt.Println()
		fmt.Println("To embed capsules:")
		fmt.Println("  1. anime capsules add <name> <path>")
		fmt.Println("  2. make build")
		return nil
	}

	// Load config to get build commands
	cfg, _ := config.Load()

	homeDir, _ := os.UserHomeDir()
	capsulesDir := filepath.Join(homeDir, ".anime", "capsules")

	for _, name := range embedded {
		if capsuleName != "" && name != capsuleName {
			continue
		}

		fmt.Printf("%s Installing capsule '%s'...\n", theme.InfoStyle.Render("▶"), name)

		// Extract capsule
		destDir := filepath.Join(capsulesDir, name)
		if err := extractCapsule(name, destDir); err != nil {
			fmt.Printf("  %s Failed to extract: %v\n", theme.ErrorStyle.Render("✗"), err)
			continue
		}

		// Checkout working tree from .git
		gitDir := filepath.Join(destDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// It's a bare-ish extraction, need to checkout
			checkoutCmd := exec.Command("git", "-C", destDir, "checkout", "-f")
			checkoutCmd.Stdout = os.Stdout
			checkoutCmd.Stderr = os.Stderr
			if err := checkoutCmd.Run(); err != nil {
				fmt.Printf("  %s Failed to checkout: %v\n", theme.ErrorStyle.Render("✗"), err)
				continue
			}
		}

		// Get build command
		buildCmd := "make install"
		if cfg != nil {
			if cap, err := cfg.GetCapsule(name); err == nil {
				buildCmd = cap.GetBuildCommand()
			}
		}

		// Run build with proper PATH (include Go locations)
		fmt.Printf("  Running: %s\n", buildCmd)
		parts := strings.Fields(buildCmd)
		build := exec.Command(parts[0], parts[1:]...)
		build.Dir = destDir
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		// Ensure Go is in PATH - add common Go locations
		build.Env = append(os.Environ(),
			"PATH="+os.Getenv("PATH")+":/usr/local/go/bin:/home/ubuntu/go/bin:"+os.Getenv("HOME")+"/go/bin",
		)
		if err := build.Run(); err != nil {
			fmt.Printf("  %s Build failed: %v\n", theme.ErrorStyle.Render("✗"), err)
			continue
		}

		fmt.Printf("  %s Installed '%s'\n", theme.SuccessStyle.Render("✓"), name)
	}

	return nil
}

func runCapsulesDeploy(cmd *cobra.Command, args []string) error {
	// Get default server from config, or fall back to "lambda"
	cfg, _ := config.Load()
	server := cfg.GetDefaultServer()
	if len(args) > 0 {
		server = args[0]
	}

	// Parse server argument (reuse from deploy.go)
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	fmt.Println(theme.InfoStyle.Render("⚡ Deploying capsules to " + target))
	fmt.Println()

	// Find capsule archives in cmd/capsules/
	sourceDir, err := findSourceDir()
	if err != nil {
		return fmt.Errorf("could not find source directory: %w", err)
	}
	capsulesDir := filepath.Join(sourceDir, "cmd", "capsules")

	entries, err := os.ReadDir(capsulesDir)
	if err != nil {
		return fmt.Errorf("failed to read capsules directory: %w", err)
	}

	var capsuleFiles []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") {
			capsuleFiles = append(capsuleFiles, filepath.Join(capsulesDir, entry.Name()))
		}
	}

	if len(capsuleFiles) == 0 {
		fmt.Println(theme.WarningStyle.Render("No capsule archives found in cmd/capsules/"))
		fmt.Println()
		fmt.Println("Pack capsules first with:")
		fmt.Println("  anime capsules pack")
		return nil
	}

	// Step 1: Create remote capsules directory
	fmt.Printf("  %s %s", theme.DimTextStyle.Render("[1/3]"), theme.InfoStyle.Render("Preparing remote... "))
	sshArgs := []string{"-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=accept-new"}
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		sshArgs = append(sshArgs, "-i", keyPath)
		defer cleanup()
	}
	prepCmd := exec.Command("ssh", append(sshArgs, target, "mkdir -p ~/.anime/capsules-staging")...)
	if err := prepCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to prepare remote: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Step 2: Transfer capsule archives
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("[2/3]"), theme.InfoStyle.Render("Transferring capsules..."))
	scpArgs := []string{"-o", "Compression=no", "-o", "StrictHostKeyChecking=accept-new"}
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		scpArgs = append(scpArgs, "-i", keyPath)
		defer cleanup()
	}

	for _, capsuleFile := range capsuleFiles {
		name := filepath.Base(capsuleFile)
		capsuleName := strings.TrimSuffix(name, ".tar.gz")
		info, _ := os.Stat(capsuleFile)
		size := formatSize(info.Size())

		fmt.Printf("      %s (%s)... ", capsuleName, size)

		args := append(scpArgs, capsuleFile, target+":~/.anime/capsules-staging/")
		scpCmd := exec.Command("scp", args...)
		if err := scpCmd.Run(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("✗"))
			return fmt.Errorf("failed to transfer %s: %w", capsuleName, err)
		}
		fmt.Println(theme.SuccessStyle.Render("✓"))
	}

	// Step 3: Extract and install capsules on remote
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("[3/3]"), theme.InfoStyle.Render("Installing capsules..."))

	// Build install script that extracts and builds each capsule
	installScript := `
cd ~/.anime/capsules-staging
for archive in *.tar.gz; do
    [ -f "$archive" ] || continue
    name="${archive%.tar.gz}"
    echo "  Installing $name..."

    # Create capsule directory
    mkdir -p ~/.anime/capsules/"$name"

    # Extract archive
    tar -xzf "$archive" -C ~/.anime/capsules/"$name"

    # Checkout working tree if it's a git repo
    if [ -d ~/.anime/capsules/"$name"/.git ]; then
        cd ~/.anime/capsules/"$name"
        git checkout -f 2>/dev/null || true
        cd ~/.anime/capsules-staging
    fi

    # Run build
    cd ~/.anime/capsules/"$name"
    export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
    if [ -f Makefile ]; then
        make install 2>&1 || echo "  Warning: build failed for $name"
    fi
    cd ~/.anime/capsules-staging
done
rm -rf ~/.anime/capsules-staging
echo "Done"
`

	installCmd := exec.Command("ssh", append(sshArgs, target, "bash -s")...)
	installCmd.Stdin = strings.NewReader(installScript)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ Installation failed"))
		return fmt.Errorf("failed to install capsules: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Capsules deployed!"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
	fmt.Printf("  %s %d capsule(s)\n", theme.DimTextStyle.Render("Installed:"), len(capsuleFiles))

	return nil
}

func runCapsulesPack(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	capsules := cfg.ListCapsules()
	if len(capsules) == 0 {
		fmt.Println("No capsules to pack")
		return nil
	}

	// Create capsules directory
	capsulesDir := filepath.Join("cmd", "capsules")
	if err := os.MkdirAll(capsulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create capsules directory: %w", err)
	}

	for _, cap := range capsules {
		srcPath := cap.GetExpandedPath()
		gitDir := filepath.Join(srcPath, ".git")

		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Printf("⚠ Skipping %s: .git not found at %s\n", cap.Name, srcPath)
			continue
		}

		tarPath := filepath.Join(capsulesDir, cap.Name+".tar.gz")
		fmt.Printf("Packing %s -> %s\n", cap.Name, tarPath)

		if err := packCapsule(srcPath, tarPath); err != nil {
			fmt.Printf("⚠ Failed to pack %s: %v\n", cap.Name, err)
			continue
		}

		// Get size
		info, _ := os.Stat(tarPath)
		fmt.Printf("✓ Packed %s (%s)\n", cap.Name, formatSize(info.Size()))
	}

	return nil
}

func packCapsule(srcPath, tarPath string) error {
	file, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Walk the source directory
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		// Skip certain directories
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "__pycache__" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			if name == "build" || name == "dist" || name == "target" {
				return filepath.SkipDir
			}
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

func listEmbeddedCapsules() []string {
	var names []string

	entries, err := embeddedCapsules.ReadDir("capsules")
	if err != nil {
		return names
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".tar.gz") {
			names = append(names, strings.TrimSuffix(name, ".tar.gz"))
		}
	}

	return names
}

func extractCapsule(name, destDir string) error {
	// Read embedded tarball
	data, err := embeddedCapsules.ReadFile(fmt.Sprintf("capsules/%s.tar.gz", name))
	if err != nil {
		return fmt.Errorf("capsule not found: %w", err)
	}

	// Remove existing directory to avoid permission issues
	if _, err := os.Stat(destDir); err == nil {
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Extract tarball
	gzr, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			os.Remove(target) // Remove if exists
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
