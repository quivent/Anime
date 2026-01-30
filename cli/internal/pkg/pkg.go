// Package pkg provides package management functionality
package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	DefaultServer  = "alice"
	PackagesPath   = "~/cpm/packages"
	PackageFile    = "cpm.json"
	InstalledFile  = ".cpm-installed.json"
	LocalModules   = "cpm_modules"
	GlobalModules  = ".cpm/packages"
)

// Package represents package metadata
type Package struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	License     string   `json:"license,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	Scripts     Scripts  `json:"scripts,omitempty"`
}

// Scripts for lifecycle hooks
type Scripts struct {
	Install     string `json:"install,omitempty"`
	PostInstall string `json:"postinstall,omitempty"`
	Build       string `json:"build,omitempty"`
	Test        string `json:"test,omitempty"`
}

// InstalledPackages tracks installed packages
type InstalledPackages struct {
	Packages map[string]InstalledPackage `json:"packages"`
}

// InstalledPackage represents an installed package
type InstalledPackage struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	Path        string `json:"path"`
	Global      bool   `json:"global"`
}

// Config holds package manager configuration
type Config struct {
	Server  string
	DryRun  bool
	Force   bool
	Global  bool
	KeyPath string
	Cleanup func()
}

// VersionInfo holds version information
type VersionInfo struct {
	Version   string
	IsLatest  bool
	Published string
}

// LoadPackageFile reads cpm.json from current directory
func LoadPackageFile() (*Package, error) {
	return LoadPackageFileFrom(".")
}

// LoadPackageFileFrom reads cpm.json from specified path
func LoadPackageFileFrom(path string) (*Package, error) {
	filePath := filepath.Join(path, PackageFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("no %s found - create one first", PackageFile)
	}

	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", PackageFile, err)
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("%s missing required field: name", PackageFile)
	}
	if pkg.Version == "" {
		return nil, fmt.Errorf("%s missing required field: version", PackageFile)
	}

	return &pkg, nil
}

// SavePackageFile writes a package file
func SavePackageFile(pkg *Package, path string) error {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(path, PackageFile)
	return os.WriteFile(filePath, data, 0644)
}

// InitPackage creates a new cpm.json
func InitPackage(name, version, description string) (*Package, error) {
	if name == "" {
		// Use current directory name
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		name = filepath.Base(cwd)
	}

	if version == "" {
		version = "1.0.0"
	}

	pkg := &Package{
		Name:        name,
		Version:     version,
		Description: description,
		License:     "MIT",
	}

	if err := SavePackageFile(pkg, "."); err != nil {
		return nil, err
	}

	return pkg, nil
}

// LoadInstalledFile reads the installed packages tracking file
func LoadInstalledFile(path string) (*InstalledPackages, error) {
	filePath := filepath.Join(path, InstalledFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &InstalledPackages{Packages: make(map[string]InstalledPackage)}, nil
	}

	var installed InstalledPackages
	if err := json.Unmarshal(data, &installed); err != nil {
		return &InstalledPackages{Packages: make(map[string]InstalledPackage)}, nil
	}

	if installed.Packages == nil {
		installed.Packages = make(map[string]InstalledPackage)
	}

	return &installed, nil
}

// SaveInstalledFile writes the installed packages tracking file
func SaveInstalledFile(path string, installed *InstalledPackages) error {
	filePath := filepath.Join(path, InstalledFile)
	data, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// GetInstallPath returns the installation path for packages
func GetInstallPath(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, GlobalModules), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, LocalModules), nil
}

// ParsePackageSpec parses package@version into name and version
func ParsePackageSpec(spec string) (name, version string) {
	parts := strings.SplitN(spec, "@", 2)
	name = parts[0]
	if len(parts) == 2 {
		version = parts[1]
	}
	return
}

// IsValidVersion validates semver-like versions
func IsValidVersion(v string) bool {
	match, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[\w.]+)?$`, v)
	return match
}

// Publish publishes a package to the registry
func Publish(target string, cfg *Config) (*Package, error) {
	pkg, err := LoadPackageFile()
	if err != nil {
		return nil, err
	}

	packagePath := filepath.Join(PackagesPath, pkg.Name, pkg.Version)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if version already exists
	checkCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("test -d %s && echo 'exists'", packagePath),
	)
	checkOutput, _ := checkCmd.CombinedOutput()
	if strings.TrimSpace(string(checkOutput)) == "exists" && !cfg.Force {
		return nil, fmt.Errorf("version %s already published - use --force or republish", pkg.Version)
	}

	if !cfg.DryRun {
		// Create remote directory
		mkdirCmd := exec.Command("ssh",
			"-i", cfg.KeyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("mkdir -p %s", packagePath),
		)
		if output, err := mkdirCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to create remote directory: %w\n%s", err, string(output))
		}
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cfg.DryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+packagePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return nil, fmt.Errorf("rsync failed: %w", err)
	}

	// Update latest symlink
	if !cfg.DryRun {
		latestPath := filepath.Join(PackagesPath, pkg.Name, "latest")
		symlinkCmd := exec.Command("ssh",
			"-i", cfg.KeyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("rm -f %s && ln -s %s %s", latestPath, pkg.Version, latestPath),
		)
		symlinkCmd.Run()
	}

	return pkg, nil
}

// Install installs a package from the registry
func Install(target, pkgSpec string, cfg *Config) (*InstalledPackage, error) {
	pkgName, pkgVersion := ParsePackageSpec(pkgSpec)
	if pkgVersion == "" {
		pkgVersion = "latest"
	}

	packagePath := filepath.Join(PackagesPath, pkgName, pkgVersion)

	installPath, err := GetInstallPath(cfg.Global)
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(installPath, pkgName)

	// Check if package exists
	checkCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("test -d %s && echo 'exists'", packagePath),
	)
	checkOutput, _ := checkCmd.CombinedOutput()
	if strings.TrimSpace(string(checkOutput)) != "exists" {
		return nil, fmt.Errorf("package not found: %s@%s", pkgName, pkgVersion)
	}

	// Check if already installed
	if _, err := os.Stat(localPath); err == nil && !cfg.Force {
		return nil, fmt.Errorf("package already installed at %s (use --force to overwrite)", localPath)
	}

	// Create install directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create install directory: %w", err)
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	rsyncArgs = append(rsyncArgs, getRsyncExcludes()...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+packagePath+"/", localPath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		os.RemoveAll(localPath)
		return nil, fmt.Errorf("install failed: %w", err)
	}

	// Read actual version from installed package
	actualVersion := pkgVersion
	if pkgVersion == "latest" {
		installedPkg, err := LoadPackageFileFrom(localPath)
		if err == nil {
			actualVersion = installedPkg.Version
		}
	}

	// Track installation
	installed, _ := LoadInstalledFile(installPath)
	pkg := InstalledPackage{
		Version:     actualVersion,
		InstalledAt: time.Now().Format(time.RFC3339),
		Path:        localPath,
		Global:      cfg.Global,
	}
	installed.Packages[pkgName] = pkg
	SaveInstalledFile(installPath, installed)

	return &pkg, nil
}

// Uninstall removes an installed package
func Uninstall(pkgName string, cfg *Config) error {
	installPath, err := GetInstallPath(cfg.Global)
	if err != nil {
		return err
	}

	localPath := filepath.Join(installPath, pkgName)

	// Check if installed
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("package not installed: %s", pkgName)
	}

	// Remove package
	if err := os.RemoveAll(localPath); err != nil {
		return fmt.Errorf("failed to uninstall: %w", err)
	}

	// Update tracking
	installed, _ := LoadInstalledFile(installPath)
	delete(installed.Packages, pkgName)
	SaveInstalledFile(installPath, installed)

	return nil
}

// Search searches for packages
func Search(target, query string, cfg *Config) ([]Package, error) {
	// List all packages
	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -1 %s 2>/dev/null || echo ''", PackagesPath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	queryLower := strings.ToLower(query)

	var matches []Package
	for _, pkgName := range packages {
		if pkgName == "" || !strings.Contains(strings.ToLower(pkgName), queryLower) {
			continue
		}

		// Get package info
		infoCmd := exec.Command("ssh",
			"-i", cfg.KeyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("cat %s/%s/latest/%s 2>/dev/null || echo '{}'", PackagesPath, pkgName, PackageFile),
		)
		infoOutput, _ := infoCmd.CombinedOutput()

		var pkg Package
		json.Unmarshal(infoOutput, &pkg)
		if pkg.Name == "" {
			pkg.Name = pkgName
		}
		matches = append(matches, pkg)
	}

	return matches, nil
}

// GetInfo gets package information
func GetInfo(target, pkgSpec string, cfg *Config) (*Package, error) {
	pkgName, pkgVersion := ParsePackageSpec(pkgSpec)
	if pkgVersion == "" {
		pkgVersion = "latest"
	}

	packagePath := filepath.Join(PackagesPath, pkgName, pkgVersion)

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("cat %s/%s 2>/dev/null || echo 'NOT_FOUND'", packagePath, PackageFile),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "NOT_FOUND" {
		return nil, fmt.Errorf("package not found: %s@%s", pkgName, pkgVersion)
	}

	var pkg Package
	if err := json.Unmarshal(output, &pkg); err != nil {
		return nil, fmt.Errorf("invalid package metadata: %w", err)
	}

	return &pkg, nil
}

// GetVersions lists all versions of a package
func GetVersions(target, pkgName string, cfg *Config) ([]VersionInfo, error) {
	packagePath := filepath.Join(PackagesPath, pkgName)

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -1 %s 2>/dev/null | grep -v latest || echo 'NOT_FOUND'", packagePath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "NOT_FOUND" || outputStr == "" {
		return nil, fmt.Errorf("package not found: %s", pkgName)
	}

	versions := strings.Split(outputStr, "\n")
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	var result []VersionInfo
	for i, v := range versions {
		result = append(result, VersionInfo{
			Version:  v,
			IsLatest: i == 0,
		})
	}

	return result, nil
}

// ListInstalled lists installed packages
func ListInstalled(global bool) (*InstalledPackages, error) {
	installPath, err := GetInstallPath(global)
	if err != nil {
		return nil, err
	}
	return LoadInstalledFile(installPath)
}

// Helper functions

func getRsyncExcludes() []string {
	return []string{
		"--exclude", ".git",
		"--exclude", "node_modules",
		"--exclude", "cpm_modules",
		"--exclude", "__pycache__",
		"--exclude", "*.pyc",
		"--exclude", ".env",
		"--exclude", "venv",
		"--exclude", ".venv",
		"--exclude", ".cpm-link",
		"--exclude", InstalledFile,
	}
}
