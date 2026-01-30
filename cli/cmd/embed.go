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
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// EmbeddedFile represents a file that has been embedded
type EmbeddedFile struct {
	OriginalPath string    `json:"original_path"`
	StoredName   string    `json:"stored_name"`
	Size         int64     `json:"size"`
	Hash         string    `json:"hash"`
	EmbeddedAt   time.Time `json:"embedded_at"`
	IsDirectory  bool      `json:"is_directory"`
}

// EmbeddedManifest tracks all embedded files
type EmbeddedManifest struct {
	Files map[string]EmbeddedFile `json:"files"`
}

var embedCmd = &cobra.Command{
	Use:   "embed FILEPATH",
	Short: "Embed a file or directory for deployment",
	Long: `Embed a file or directory into anime's embedded filesystem.

The embedded files are stored in ~/.anime/embedded/ and will be automatically
included when you push to a remote server. You can extract them on the remote
server using 'anime extract --embedded FILEPATH'.

Examples:
  anime embed config.yaml              # Embed a single file
  anime embed ./models                 # Embed a directory
  anime embed ~/datasets/training.csv  # Embed file from home directory

The embedded files are tracked in a manifest and can be listed with:
  anime extract --list
`,
	Args: cobra.ExactArgs(1),
	RunE: runEmbed,
}

func init() {
	rootCmd.AddCommand(embedCmd)
}

func runEmbed(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]

	// Expand home directory if needed
	if sourcePath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		sourcePath = filepath.Join(home, sourcePath[1:])
	}

	// Make path absolute
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	// Check if source exists
	sourceInfo, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(theme.ErrorStyle.Render("❌ File or directory not found"))
			fmt.Println()
			fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(absSourcePath))
			fmt.Println()
			return fmt.Errorf("source not found: %s", absSourcePath)
		}
		return fmt.Errorf("failed to access source: %w", err)
	}

	fmt.Println(theme.InfoStyle.Render("📦 Embedding file for deployment"))
	fmt.Println()
	fmt.Printf("  Source: %s\n", theme.HighlightStyle.Render(absSourcePath))
	if sourceInfo.IsDir() {
		fmt.Printf("  Type:   %s\n", theme.InfoStyle.Render("directory"))
	} else {
		fmt.Printf("  Type:   %s\n", theme.InfoStyle.Render("file"))
		fmt.Printf("  Size:   %s\n", theme.DimTextStyle.Render(formatBytesSize(sourceInfo.Size())))
	}
	fmt.Println()

	// Create embedded directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	embeddedDir := filepath.Join(home, ".anime", "embedded")
	if err := os.MkdirAll(embeddedDir, 0755); err != nil {
		return fmt.Errorf("failed to create embedded directory: %w", err)
	}

	// Generate unique name for the embedded file
	hash := sha256.New()
	hash.Write([]byte(absSourcePath))
	hash.Write([]byte(time.Now().String()))
	hashStr := hex.EncodeToString(hash.Sum(nil))[:16]

	storedName := fmt.Sprintf("%s-%s.tar.gz", filepath.Base(absSourcePath), hashStr)
	tarPath := filepath.Join(embeddedDir, storedName)

	// Create tar.gz archive
	fmt.Print(theme.DimTextStyle.Render("▶ Creating archive... "))
	totalSize, fileHash, err := createTarGz(absSourcePath, tarPath)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to create archive: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Update manifest
	fmt.Print(theme.DimTextStyle.Render("▶ Updating manifest... "))
	if err := updateManifest(absSourcePath, storedName, totalSize, fileHash, sourceInfo.IsDir()); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ File embedded successfully!"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The file will be automatically included when you push to a server:"))
	fmt.Println(theme.HighlightStyle.Render("    anime push"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Extract on remote server with:"))
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("    anime extract --embedded %s", filepath.Base(absSourcePath))))
	fmt.Println()

	return nil
}

func createTarGz(sourcePath, tarPath string) (int64, string, error) {
	// Create tar.gz file
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return 0, "", err
	}
	defer tarFile.Close()

	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	var totalSize int64
	hash := sha256.New()

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return 0, "", err
	}

	if sourceInfo.IsDir() {
		// Add directory
		err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip hidden files and directories
			if strings.HasPrefix(info.Name(), ".") && path != sourcePath {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			// Calculate relative path
			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.Join(filepath.Base(sourcePath), relPath)

			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// Copy file content
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			n, err := io.Copy(io.MultiWriter(tarWriter, hash), file)
			totalSize += n
			return err
		})
	} else {
		// Add single file
		header, err := tar.FileInfoHeader(sourceInfo, "")
		if err != nil {
			return 0, "", err
		}
		header.Name = filepath.Base(sourcePath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return 0, "", err
		}

		file, err := os.Open(sourcePath)
		if err != nil {
			return 0, "", err
		}
		defer file.Close()

		totalSize, err = io.Copy(io.MultiWriter(tarWriter, hash), file)
		if err != nil {
			return 0, "", err
		}
	}

	return totalSize, hex.EncodeToString(hash.Sum(nil)), err
}

func updateManifest(originalPath, storedName string, size int64, hash string, isDir bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(home, ".anime", "embedded", "manifest.json")

	// Load existing manifest or create new one
	manifest := &EmbeddedManifest{
		Files: make(map[string]EmbeddedFile),
	}

	if data, err := os.ReadFile(manifestPath); err == nil {
		if err := json.Unmarshal(data, manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}
	}

	// Add or update entry
	manifest.Files[filepath.Base(originalPath)] = EmbeddedFile{
		OriginalPath: originalPath,
		StoredName:   storedName,
		Size:         size,
		Hash:         hash,
		EmbeddedAt:   time.Now(),
		IsDirectory:  isDir,
	}

	// Save manifest
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

func formatBytesSize(bytes int64) string {
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
