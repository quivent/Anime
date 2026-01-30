package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	extractEmbedded bool
	extractList     bool
)

var extractCmd = &cobra.Command{
	Use:     "extract [destination]",
	Aliases: []string{"src"},
	Short:   "Extract embedded source code or files to a directory",
	Long: `Extract the anime-cli source code or embedded files that were bundled with the binary.

The source code is embedded directly in the binary during build.
Embedded files are stored in ~/.anime/embedded/ when using 'anime embed'.

Examples:
  anime extract                          # Extract source to ./anime-src/
  anime src                              # Same as extract (alias)
  anime extract ~/projects/anime         # Extract source to specific directory
  anime extract --list                   # List all embedded files
  anime extract --embedded config.yaml   # Extract embedded file
  anime extract --embedded models .      # Extract embedded directory to current dir
`,
	Args: cobra.MaximumNArgs(2),
	RunE: runExtract,
}

func init() {
	extractCmd.Flags().BoolVar(&extractEmbedded, "embedded", false, "Extract an embedded file instead of source code")
	extractCmd.Flags().BoolVar(&extractList, "list", false, "List all embedded files")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	// Handle --list flag
	if extractList {
		return listEmbeddedFiles()
	}

	// Handle --embedded flag
	if extractEmbedded {
		if len(args) == 0 {
			return fmt.Errorf("please specify which embedded file to extract")
		}
		fileName := args[0]
		destination := "."
		if len(args) > 1 {
			destination = args[1]
		}
		return extractEmbeddedFile(fileName, destination)
	}

	// Default: extract source code from embedded tarball
	destination := "./anime-src"
	if len(args) > 0 {
		destination = args[0]
	}

	// Expand home directory if needed
	if len(destination) > 0 && destination[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		destination = filepath.Join(home, destination[1:])
	}

	// Make destination absolute
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Check if source is embedded in the binary
	if !HasEmbeddedSource() {
		fmt.Println(theme.ErrorStyle.Render("❌ No embedded source code found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 This binary was not built with embedded source code."))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  To build with embedded source, use:"))
		fmt.Println(theme.HighlightStyle.Render("    make build"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Or clone the repository:"))
		fmt.Println(theme.HighlightStyle.Render("    git clone https://github.com/joshkornreich/anime.git"))
		fmt.Println()
		return fmt.Errorf("no embedded source code available")
	}

	// Show extraction info
	sizeMB := GetEmbeddedSourceSize() / (1024 * 1024)
	fmt.Println(theme.InfoStyle.Render("📦 Extracting embedded source code"))
	fmt.Println()
	fmt.Printf("  Size: %s\n", theme.DimTextStyle.Render(fmt.Sprintf("%d MB", sizeMB)))
	fmt.Printf("  To:   %s\n", theme.HighlightStyle.Render(absDestination))
	if BuildDir != "" {
		fmt.Printf("  Built from: %s\n", theme.DimTextStyle.Render(BuildDir))
	}
	fmt.Println()

	// Extract the embedded source
	extractedPath, err := ExtractEmbeddedSource(absDestination)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ Extraction failed"))
		return fmt.Errorf("failed to extract source: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Source code extracted!"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  You can now build the binary:"))
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("    cd %s", extractedPath)))
	fmt.Println(theme.HighlightStyle.Render("    make build"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Or develop with Claude:"))
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("    anime develop %s", extractedPath)))
	fmt.Println()

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file
func copyFile(src, dst string, mode os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}

// listEmbeddedFiles lists all embedded files
func listEmbeddedFiles() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	manifestPath := filepath.Join(home, ".anime", "embedded", "manifest.json")

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Println(theme.InfoStyle.Render("📦 No embedded files found"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  To embed a file, use:"))
		fmt.Println(theme.HighlightStyle.Render("    anime embed FILEPATH"))
		fmt.Println()
		return nil
	}

	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest EmbeddedManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	if len(manifest.Files) == 0 {
		fmt.Println(theme.InfoStyle.Render("📦 No embedded files found"))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.InfoStyle.Render("📦 Embedded Files"))
	fmt.Println()

	for name, file := range manifest.Files {
		fileType := "file"
		if file.IsDirectory {
			fileType = "directory"
		}
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(name))
		fmt.Printf("    Type:     %s\n", theme.DimTextStyle.Render(fileType))
		fmt.Printf("    Size:     %s\n", theme.DimTextStyle.Render(formatSizeForExtract(file.Size)))
		fmt.Printf("    Embedded: %s\n", theme.DimTextStyle.Render(file.EmbeddedAt.Format("2006-01-02 15:04:05")))
		fmt.Println()
	}

	fmt.Println(theme.DimTextStyle.Render("  To extract a file, use:"))
	fmt.Println(theme.HighlightStyle.Render("    anime extract --embedded FILENAME [destination]"))
	fmt.Println()

	return nil
}

// extractEmbeddedFile extracts a specific embedded file
func extractEmbeddedFile(fileName, destination string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	manifestPath := filepath.Join(home, ".anime", "embedded", "manifest.json")

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Println(theme.ErrorStyle.Render("❌ No embedded files found"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  To embed a file, use:"))
		fmt.Println(theme.HighlightStyle.Render("    anime embed FILEPATH"))
		fmt.Println()
		return fmt.Errorf("no embedded files found")
	}

	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest EmbeddedManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Find file in manifest
	embeddedFile, found := manifest.Files[fileName]
	if !found {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ File '%s' not found in embedded files", fileName)))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Available files:"))
		for name := range manifest.Files {
			fmt.Printf("    - %s\n", theme.HighlightStyle.Render(name))
		}
		fmt.Println()
		return fmt.Errorf("file not found: %s", fileName)
	}

	// Expand home directory if needed
	if len(destination) > 0 && destination[0] == '~' {
		destination = filepath.Join(home, destination[1:])
	}

	// Make destination absolute
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Path to the tar.gz file
	tarPath := filepath.Join(home, ".anime", "embedded", embeddedFile.StoredName)

	fmt.Println(theme.InfoStyle.Render("📦 Extracting embedded file"))
	fmt.Println()
	fmt.Printf("  File: %s\n", theme.HighlightStyle.Render(fileName))
	if embeddedFile.IsDirectory {
		fmt.Printf("  Type: %s\n", theme.InfoStyle.Render("directory"))
	} else {
		fmt.Printf("  Type: %s\n", theme.InfoStyle.Render("file"))
	}
	fmt.Printf("  To:   %s\n", theme.DimTextStyle.Render(absDestination))
	fmt.Println()

	// Create destination directory
	if err := os.MkdirAll(absDestination, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract tar.gz
	fmt.Print(theme.DimTextStyle.Render("▶ Extracting archive... "))
	if err := extractTarGz(tarPath, absDestination); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ File extracted successfully!"))
	fmt.Println()

	return nil
}

// extractTarGz extracts a tar.gz archive
func extractTarGz(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
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
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatSizeForExtract(bytes int64) string {
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
