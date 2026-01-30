// Package cli provides CLI management functionality for the anime CLI
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// RegistryDir is the directory where CLI registry data is stored
	RegistryDir = ".anime/cli"
	// RegistryFile is the name of the registry file
	RegistryFile = "registry.json"
	// SourceDir is the subdirectory where CLI sources are stored
	SourceDir = "src"
	// BinDir is the subdirectory where built/registered binaries are stored
	BinDir = "bin"
)

// CLIType represents the type of CLI registration
type CLIType string

const (
	// TypeSource indicates the CLI was added from source
	TypeSource CLIType = "source"
	// TypeBinary indicates a prebuilt binary was registered
	TypeBinary CLIType = "binary"
	// TypeRemote indicates the CLI was pulled from a remote source
	TypeRemote CLIType = "remote"
)

// CLI represents a registered CLI tool
type CLI struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        CLIType   `json:"type"`
	SourcePath  string    `json:"source_path,omitempty"`   // Path to source directory
	BinaryPath  string    `json:"binary_path,omitempty"`   // Path to binary
	RemoteURL   string    `json:"remote_url,omitempty"`    // Remote URL (GitHub, etc.)
	Version     string    `json:"version,omitempty"`       // Version string
	Language    string    `json:"language,omitempty"`      // go, rust, python, etc.
	Built       bool      `json:"built"`                   // Whether binary has been built
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Registry manages the collection of registered CLIs
type Registry struct {
	CLIs    map[string]*CLI `json:"clis"`
	path    string
	homeDir string
}

// GetRegistryPath returns the full path to the registry directory
func GetRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, RegistryDir), nil
}

// GetSourcePath returns the path to a CLI's source directory
func GetSourcePath(name string) (string, error) {
	regPath, err := GetRegistryPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(regPath, SourceDir, name), nil
}

// GetBinPath returns the path to the binaries directory
func GetBinPath() (string, error) {
	regPath, err := GetRegistryPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(regPath, BinDir), nil
}

// GetCLIBinaryPath returns the path to a specific CLI's binary
func GetCLIBinaryPath(name string) (string, error) {
	binPath, err := GetBinPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(binPath, name), nil
}

// LoadRegistry loads the CLI registry from disk
func LoadRegistry() (*Registry, error) {
	regPath, err := GetRegistryPath()
	if err != nil {
		return nil, err
	}

	home, _ := os.UserHomeDir()
	r := &Registry{
		CLIs:    make(map[string]*CLI),
		path:    regPath,
		homeDir: home,
	}

	registryFile := filepath.Join(regPath, RegistryFile)
	data, err := os.ReadFile(registryFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory structure
			if err := os.MkdirAll(filepath.Join(regPath, SourceDir), 0755); err != nil {
				return nil, fmt.Errorf("failed to create source directory: %w", err)
			}
			if err := os.MkdirAll(filepath.Join(regPath, BinDir), 0755); err != nil {
				return nil, fmt.Errorf("failed to create bin directory: %w", err)
			}
			return r, nil
		}
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	if err := json.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return r, nil
}

// Save persists the registry to disk
func (r *Registry) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(r.path, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	registryFile := filepath.Join(r.path, RegistryFile)
	if err := os.WriteFile(registryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// Add adds a new CLI to the registry
func (r *Registry) Add(cli *CLI) error {
	if cli.Name == "" {
		return fmt.Errorf("CLI name cannot be empty")
	}

	if _, exists := r.CLIs[cli.Name]; exists {
		return fmt.Errorf("CLI '%s' already exists", cli.Name)
	}

	cli.AddedAt = time.Now()
	cli.UpdatedAt = time.Now()
	r.CLIs[cli.Name] = cli

	return r.Save()
}

// Update updates an existing CLI in the registry
func (r *Registry) Update(cli *CLI) error {
	if cli.Name == "" {
		return fmt.Errorf("CLI name cannot be empty")
	}

	if _, exists := r.CLIs[cli.Name]; !exists {
		return fmt.Errorf("CLI '%s' does not exist", cli.Name)
	}

	cli.UpdatedAt = time.Now()
	r.CLIs[cli.Name] = cli

	return r.Save()
}

// Remove removes a CLI from the registry
func (r *Registry) Remove(name string) error {
	if _, exists := r.CLIs[name]; !exists {
		return fmt.Errorf("CLI '%s' does not exist", name)
	}

	delete(r.CLIs, name)
	return r.Save()
}

// Get retrieves a CLI by name
func (r *Registry) Get(name string) (*CLI, bool) {
	cli, exists := r.CLIs[name]
	return cli, exists
}

// List returns all registered CLIs
func (r *Registry) List() []*CLI {
	clis := make([]*CLI, 0, len(r.CLIs))
	for _, cli := range r.CLIs {
		clis = append(clis, cli)
	}
	return clis
}

// Exists checks if a CLI exists in the registry
func (r *Registry) Exists(name string) bool {
	_, exists := r.CLIs[name]
	return exists
}
