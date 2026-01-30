package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/theme"
)

// SourceTarball contains the anime CLI source code as a tarball.
// This is populated at build time by embedding anime-src.tar.gz.
// The tarball is created by the Makefile before building.
//
//go:embed anime-src.tar.gz
var SourceTarball embed.FS

// HasEmbeddedSource returns true if source code is embedded in the binary
func HasEmbeddedSource() bool {
	_, err := SourceTarball.ReadFile("anime-src.tar.gz")
	return err == nil
}

// GetEmbeddedSourceSize returns the size of the embedded source tarball
func GetEmbeddedSourceSize() int64 {
	data, err := SourceTarball.ReadFile("anime-src.tar.gz")
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// ExtractEmbeddedSource extracts the embedded source tarball to the specified directory.
// Returns the path to the extracted cli directory.
func ExtractEmbeddedSource(destDir string) (string, error) {
	// Read the embedded tarball
	data, err := SourceTarball.ReadFile("anime-src.tar.gz")
	if err != nil {
		return "", fmt.Errorf("no embedded source available: %w", err)
	}

	sizeMB := len(data) / (1024 * 1024)
	fmt.Printf("  %s Extracting embedded source (%d MB)...\n",
		theme.InfoStyle.Render("📦"),
		sizeMB)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create a gzip reader
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	fileCount := 0
	var cliDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar read error: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		// Track the cli directory
		if cliDir == "" && filepath.Base(header.Name) == "go.mod" {
			cliDir = filepath.Dir(target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("failed to create file %s: %w", target, err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", fmt.Errorf("failed to write file %s: %w", target, err)
			}
			f.Close()
			fileCount++
		}
	}

	fmt.Printf("  %s Extracted %d files\n",
		theme.SuccessStyle.Render("✓"),
		fileCount)

	// Return the directory containing go.mod
	if cliDir != "" {
		return cliDir, nil
	}

	// Fallback: look for common structures
	possiblePaths := []string{
		filepath.Join(destDir, "cli"),
		filepath.Join(destDir, "anime", "cli"),
		destDir,
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("could not locate go.mod in extracted source")
}

// GetSourceCacheDir returns the cache directory for extracted source
func GetSourceCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "/tmp"
	}
	return filepath.Join(cacheDir, "anime-cli-source")
}
