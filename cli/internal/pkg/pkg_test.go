package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		input           string
		expectedName    string
		expectedVersion string
	}{
		{"mypackage", "mypackage", ""},
		{"mypackage@1.0.0", "mypackage", "1.0.0"},
		{"package@latest", "package", "latest"},
		{"pkg@1.0.0-beta.1", "pkg", "1.0.0-beta.1"},
		{"my-package@2.0.0", "my-package", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version := ParsePackageSpec(tt.input)
			if name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, name)
			}
			if version != tt.expectedVersion {
				t.Errorf("Expected version %s, got %s", tt.expectedVersion, version)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.1.0", true},
		{"10.20.30", true},
		{"1.0.0-alpha", true},
		{"1.0.0-beta.1", true},
		{"1.0.0-rc.1", true},
		{"1.0", false},
		{"1", false},
		{"v1.0.0", false},
		{"latest", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := IsValidVersion(tt.version)
			if result != tt.valid {
				t.Errorf("IsValidVersion(%s) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestLoadPackageFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with no file
	_, err = LoadPackageFile()
	if err == nil {
		t.Error("Expected error with no package file")
	}

	// Test with valid file
	pkg := Package{
		Name:        "testpkg",
		Version:     "1.0.0",
		Description: "Test package",
		Author:      "Test Author",
		License:     "MIT",
	}
	data, _ := json.MarshalIndent(pkg, "", "  ")
	os.WriteFile(PackageFile, data, 0644)

	result, err := LoadPackageFile()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", result.Version)
	}

	// Test with missing name
	badPkg := `{"version": "1.0.0"}`
	os.WriteFile(PackageFile, []byte(badPkg), 0644)
	_, err = LoadPackageFile()
	if err == nil {
		t.Error("Expected error with missing name")
	}

	// Test with missing version
	badPkg = `{"name": "testpkg"}`
	os.WriteFile(PackageFile, []byte(badPkg), 0644)
	_, err = LoadPackageFile()
	if err == nil {
		t.Error("Expected error with missing version")
	}
}

func TestLoadPackageFileFrom(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with no file
	_, err = LoadPackageFileFrom(tmpDir)
	if err == nil {
		t.Error("Expected error with no package file")
	}

	// Create valid package file
	pkg := Package{Name: "testpkg", Version: "2.0.0"}
	data, _ := json.Marshal(pkg)
	os.WriteFile(filepath.Join(tmpDir, PackageFile), data, 0644)

	result, err := LoadPackageFileFrom(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Name != "testpkg" {
		t.Errorf("Expected name 'testpkg', got %s", result.Name)
	}
	if result.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", result.Version)
	}
}

func TestSavePackageFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pkg := &Package{
		Name:        "savepkg",
		Version:     "3.0.0",
		Description: "Saved package",
		Author:      "Test",
		Keywords:    []string{"test", "package"},
		License:     "Apache-2.0",
	}

	err = SavePackageFile(pkg, tmpDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file was created correctly
	data, err := os.ReadFile(filepath.Join(tmpDir, PackageFile))
	if err != nil {
		t.Errorf("Failed to read package file: %v", err)
	}

	var result Package
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to parse package file: %v", err)
	}

	if result.Name != "savepkg" {
		t.Errorf("Expected name 'savepkg', got %s", result.Name)
	}
	if len(result.Keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(result.Keywords))
	}
}

func TestInitPackage(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with custom name
	pkg, err := InitPackage("mypkg", "0.0.1", "My description")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pkg.Name != "mypkg" {
		t.Errorf("Expected name 'mypkg', got %s", pkg.Name)
	}
	if pkg.Version != "0.0.1" {
		t.Errorf("Expected version '0.0.1', got %s", pkg.Version)
	}
	if pkg.License != "MIT" {
		t.Errorf("Expected default license 'MIT', got %s", pkg.License)
	}

	// Verify file was created
	if _, err := os.Stat(PackageFile); os.IsNotExist(err) {
		t.Error("Package file was not created")
	}
}

func TestInitPackageDefaults(t *testing.T) {
	// Create temp directory with specific name
	tmpDir, err := os.MkdirTemp("", "myproject")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with empty name (should use directory name)
	pkg, err := InitPackage("", "", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Name should be based on directory
	if pkg.Name == "" {
		t.Error("Expected name to be set from directory")
	}

	// Version should default to 1.0.0
	if pkg.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got %s", pkg.Version)
	}
}

func TestLoadInstalledFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with no file (should return empty struct)
	installed, err := LoadInstalledFile(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if installed.Packages == nil {
		t.Error("Packages map should be initialized")
	}
	if len(installed.Packages) != 0 {
		t.Errorf("Expected empty packages, got %d", len(installed.Packages))
	}

	// Create installed file
	data := `{
		"packages": {
			"pkg1": {"version": "1.0.0", "installed_at": "2024-01-01", "path": "/tmp/pkg1", "global": false},
			"pkg2": {"version": "2.0.0", "installed_at": "2024-01-02", "path": "/tmp/pkg2", "global": true}
		}
	}`
	os.WriteFile(filepath.Join(tmpDir, InstalledFile), []byte(data), 0644)

	installed, err = LoadInstalledFile(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(installed.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(installed.Packages))
	}
	if installed.Packages["pkg1"].Version != "1.0.0" {
		t.Errorf("Expected pkg1 version '1.0.0', got %s", installed.Packages["pkg1"].Version)
	}
	if !installed.Packages["pkg2"].Global {
		t.Error("Expected pkg2 to be global")
	}
}

func TestSaveInstalledFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	installed := &InstalledPackages{
		Packages: map[string]InstalledPackage{
			"mypkg": {
				Version:     "1.0.0",
				InstalledAt: "2024-01-01T12:00:00Z",
				Path:        "/tmp/mypkg",
				Global:      false,
			},
		},
	}

	err = SaveInstalledFile(tmpDir, installed)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(tmpDir, InstalledFile))
	if err != nil {
		t.Errorf("Failed to read installed file: %v", err)
	}

	var result InstalledPackages
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to parse installed file: %v", err)
	}

	if len(result.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(result.Packages))
	}
}

func TestGetInstallPath(t *testing.T) {
	// Test local path
	localPath, err := GetInstallPath(false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !filepath.IsAbs(localPath) {
		t.Error("Expected absolute path")
	}
	if filepath.Base(localPath) != LocalModules {
		t.Errorf("Expected path ending in %s, got %s", LocalModules, filepath.Base(localPath))
	}

	// Test global path
	globalPath, err := GetInstallPath(true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !filepath.IsAbs(globalPath) {
		t.Error("Expected absolute path")
	}
	// Global path should contain .cpm/packages
	if filepath.Base(globalPath) != "packages" {
		t.Errorf("Expected path ending in 'packages', got %s", filepath.Base(globalPath))
	}
}

func TestConstants(t *testing.T) {
	if DefaultServer != "alice" {
		t.Errorf("Expected DefaultServer 'alice', got %s", DefaultServer)
	}
	if PackagesPath != "~/cpm/packages" {
		t.Errorf("Expected PackagesPath '~/cpm/packages', got %s", PackagesPath)
	}
	if PackageFile != "cpm.json" {
		t.Errorf("Expected PackageFile 'cpm.json', got %s", PackageFile)
	}
	if InstalledFile != ".cpm-installed.json" {
		t.Errorf("Expected InstalledFile '.cpm-installed.json', got %s", InstalledFile)
	}
	if LocalModules != "cpm_modules" {
		t.Errorf("Expected LocalModules 'cpm_modules', got %s", LocalModules)
	}
}

func TestPackageJSON(t *testing.T) {
	pkg := Package{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test package",
		Author:      "Test Author",
		Keywords:    []string{"test", "demo"},
		License:     "MIT",
		Repository:  "https://github.com/test/test",
		Scripts: Scripts{
			Install:     "npm install",
			PostInstall: "npm run build",
			Build:       "go build",
			Test:        "go test ./...",
		},
	}

	// Test marshaling
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	// Test unmarshaling
	var result Package
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	if result.Name != pkg.Name {
		t.Error("Name mismatch")
	}
	if result.Scripts.Build != "go build" {
		t.Error("Scripts.Build mismatch")
	}
	if len(result.Keywords) != 2 {
		t.Error("Keywords mismatch")
	}
}

func TestVersionInfo(t *testing.T) {
	info := VersionInfo{
		Version:   "1.0.0",
		IsLatest:  true,
		Published: "2024-01-01",
	}

	if info.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
	if !info.IsLatest {
		t.Error("IsLatest should be true")
	}
}

func TestConfig(t *testing.T) {
	cleanupCalled := false
	cfg := &Config{
		Server:  "testserver",
		DryRun:  true,
		Force:   false,
		Global:  true,
		KeyPath: "/tmp/key",
		Cleanup: func() { cleanupCalled = true },
	}

	if cfg.Server != "testserver" {
		t.Error("Server mismatch")
	}
	if !cfg.DryRun {
		t.Error("DryRun should be true")
	}
	if cfg.Force {
		t.Error("Force should be false")
	}
	if !cfg.Global {
		t.Error("Global should be true")
	}

	cfg.Cleanup()
	if !cleanupCalled {
		t.Error("Cleanup was not called")
	}
}

func TestListInstalled(t *testing.T) {
	// Create temp directories for local and global
	localDir, err := os.MkdirTemp("", "pkg_local_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(localDir)

	// Change to local dir (for local packages test)
	oldWd, _ := os.Getwd()
	os.Chdir(localDir)
	defer os.Chdir(oldWd)

	// Create cpm_modules directory
	os.MkdirAll(LocalModules, 0755)

	// Test local (should be empty)
	installed, err := ListInstalled(false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if installed.Packages == nil {
		t.Error("Packages should be initialized")
	}
}

func TestGetRsyncExcludes(t *testing.T) {
	excludes := getRsyncExcludes()

	// Should have pairs of --exclude and pattern
	if len(excludes)%2 != 0 {
		t.Errorf("Expected even number of exclude args, got %d", len(excludes))
	}

	// Check for common excludes
	expectedExcludes := []string{".git", "node_modules", "__pycache__", ".env"}
	for _, expected := range expectedExcludes {
		found := false
		for i := 0; i < len(excludes); i += 2 {
			if excludes[i] == "--exclude" && excludes[i+1] == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected exclude for %s not found", expected)
		}
	}
}
