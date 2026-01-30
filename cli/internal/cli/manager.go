package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles CLI operations
type Manager struct {
	Registry *Registry
}

// NewManager creates a new CLI manager
func NewManager() (*Manager, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	return &Manager{
		Registry: registry,
	}, nil
}

// AddFromSource adds a CLI from local source directory
func (m *Manager) AddFromSource(name, sourcePath string, opts AddOptions) (*CLI, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify source exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source path does not exist: %s", absPath)
	}

	// Detect language
	lang := detectLanguage(absPath)

	// Copy source to registry
	destPath, err := GetSourcePath(name)
	if err != nil {
		return nil, err
	}

	if err := copyDir(absPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy source: %w", err)
	}

	cli := &CLI{
		Name:        name,
		Description: opts.Description,
		Type:        TypeSource,
		SourcePath:  destPath,
		Language:    lang,
		Built:       false,
	}

	if err := m.Registry.Add(cli); err != nil {
		// Cleanup on failure
		os.RemoveAll(destPath)
		return nil, err
	}

	return cli, nil
}

// RegisterBinary registers a prebuilt binary
func (m *Manager) RegisterBinary(name, binaryPath string, opts AddOptions) (*CLI, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify binary exists and is executable
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("binary does not exist: %s", absPath)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a binary: %s", absPath)
	}

	// Copy binary to bin directory
	binPath, err := GetBinPath()
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(binPath, name)
	if err := copyFile(absPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to make binary executable: %w", err)
	}

	cli := &CLI{
		Name:        name,
		Description: opts.Description,
		Type:        TypeBinary,
		BinaryPath:  destPath,
		Built:       true,
	}

	if err := m.Registry.Add(cli); err != nil {
		os.Remove(destPath)
		return nil, err
	}

	return cli, nil
}

// PullFromRemote pulls CLI source from a remote URL (GitHub, etc.)
func (m *Manager) PullFromRemote(name, remoteURL string, opts AddOptions) (*CLI, error) {
	destPath, err := GetSourcePath(name)
	if err != nil {
		return nil, err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Clone repository
	cmd := exec.Command("git", "clone", "--depth", "1", remoteURL, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// Detect language
	lang := detectLanguage(destPath)

	cli := &CLI{
		Name:        name,
		Description: opts.Description,
		Type:        TypeRemote,
		SourcePath:  destPath,
		RemoteURL:   remoteURL,
		Language:    lang,
		Built:       false,
	}

	if err := m.Registry.Add(cli); err != nil {
		os.RemoveAll(destPath)
		return nil, err
	}

	return cli, nil
}

// Build builds a CLI from source
func (m *Manager) Build(name string) error {
	cli, exists := m.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	if cli.Type == TypeBinary {
		return fmt.Errorf("CLI '%s' is a prebuilt binary, cannot build", name)
	}

	if cli.SourcePath == "" {
		return fmt.Errorf("CLI '%s' has no source path", name)
	}

	binPath, err := GetCLIBinaryPath(name)
	if err != nil {
		return err
	}

	// Ensure bin directory exists
	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Build based on language
	switch cli.Language {
	case "go":
		if err := m.buildGo(cli.SourcePath, binPath); err != nil {
			return err
		}
	case "rust":
		if err := m.buildRust(cli.SourcePath, binPath); err != nil {
			return err
		}
	case "python":
		// For Python, we create a wrapper script
		if err := m.createPythonWrapper(cli.SourcePath, binPath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported language: %s", cli.Language)
	}

	// Update registry
	cli.BinaryPath = binPath
	cli.Built = true
	return m.Registry.Update(cli)
}

// Execute runs a registered CLI
func (m *Manager) Execute(name string, args []string) error {
	cli, exists := m.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	var binaryPath string

	if cli.Built && cli.BinaryPath != "" {
		binaryPath = cli.BinaryPath
	} else if cli.Type == TypeBinary {
		binaryPath = cli.BinaryPath
	} else {
		// Try to find binary in standard location
		binPath, err := GetCLIBinaryPath(name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return fmt.Errorf("CLI '%s' has not been built (use 'anime cli build %s')", name, name)
		}
		binaryPath = binPath
	}

	// Execute
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Remove removes a CLI and its associated files
func (m *Manager) Remove(name string) error {
	cli, exists := m.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	// Remove source directory if exists
	if cli.SourcePath != "" {
		os.RemoveAll(cli.SourcePath)
	}

	// Remove binary if exists
	if cli.BinaryPath != "" {
		os.Remove(cli.BinaryPath)
	}

	// Also try standard binary location
	binPath, err := GetCLIBinaryPath(name)
	if err == nil {
		os.Remove(binPath)
	}

	return m.Registry.Remove(name)
}

// Update updates a CLI from its remote source
func (m *Manager) Update(name string) error {
	cli, exists := m.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	if cli.Type != TypeRemote || cli.RemoteURL == "" {
		return fmt.Errorf("CLI '%s' was not pulled from a remote source", name)
	}

	if cli.SourcePath == "" {
		return fmt.Errorf("CLI '%s' has no source path", name)
	}

	// Pull latest changes
	cmd := exec.Command("git", "-C", cli.SourcePath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return m.Registry.Update(cli)
}

// AddOptions contains options for adding a CLI
type AddOptions struct {
	Description string
	Version     string
}

// Build helpers

func (m *Manager) buildGo(sourcePath, outputPath string) error {
	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = sourcePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) buildRust(sourcePath, outputPath string) error {
	// Build in release mode
	cmd := exec.Command("cargo", "build", "--release")
	cmd.Dir = sourcePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Find the built binary in target/release
	releaseDir := filepath.Join(sourcePath, "target", "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		return fmt.Errorf("failed to read release directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Check if executable
		if info.Mode()&0111 != 0 {
			return copyFile(filepath.Join(releaseDir, entry.Name()), outputPath)
		}
	}

	return fmt.Errorf("no executable found in target/release")
}

func (m *Manager) createPythonWrapper(sourcePath, outputPath string) error {
	// Find main entry point
	mainFile := ""
	for _, candidate := range []string{"main.py", "__main__.py", "cli.py", "app.py"} {
		if _, err := os.Stat(filepath.Join(sourcePath, candidate)); err == nil {
			mainFile = candidate
			break
		}
	}

	if mainFile == "" {
		return fmt.Errorf("no Python entry point found (main.py, __main__.py, cli.py, or app.py)")
	}

	// Create wrapper script
	wrapper := fmt.Sprintf(`#!/usr/bin/env python3
import sys
import os

# Add source to path
sys.path.insert(0, %q)
os.chdir(%q)

# Run main
exec(open(%q).read())
`, sourcePath, sourcePath, filepath.Join(sourcePath, mainFile))

	if err := os.WriteFile(outputPath, []byte(wrapper), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper: %w", err)
	}

	return nil
}

// Utility functions

func detectLanguage(path string) string {
	// Check for language-specific files
	checks := []struct {
		file string
		lang string
	}{
		{"go.mod", "go"},
		{"go.sum", "go"},
		{"Cargo.toml", "rust"},
		{"requirements.txt", "python"},
		{"setup.py", "python"},
		{"pyproject.toml", "python"},
		{"package.json", "node"},
		{"Makefile", "make"},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(path, check.file)); err == nil {
			return check.lang
		}
	}

	// Check for main files
	if hasGoFiles(path) {
		return "go"
	}

	return "unknown"
}

func hasGoFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func copyDir(src, dst string) error {
	// Remove destination if exists
	os.RemoveAll(dst)

	// Use rsync for efficiency
	cmd := exec.Command("rsync", "-av", "--exclude", ".git", src+"/", dst+"/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback to cp
		return copyDirRecursive(src, dst)
	}
	return nil
}

func copyDirRecursive(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, info.Mode())
}

// InstallToPath installs a CLI binary to ~/.local/bin
func InstallToPath(name, binaryPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	localBin := filepath.Join(home, ".local", "bin")

	// Ensure directory exists
	if err := os.MkdirAll(localBin, 0755); err != nil {
		return fmt.Errorf("failed to create ~/.local/bin: %w", err)
	}

	destPath := filepath.Join(localBin, name)
	if err := copyFile(binaryPath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}
