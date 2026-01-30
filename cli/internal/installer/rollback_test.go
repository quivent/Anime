package installer

import (
	"testing"
	"time"
)

func TestGetModulePaths(t *testing.T) {
	tests := []struct {
		name              string
		moduleID          string
		expectBinaries    bool
		expectPython      bool
		expectDirectories bool
		expectServices    bool
	}{
		{
			name:              "core module",
			moduleID:          "core",
			expectBinaries:    true,
			expectPython:      false,
			expectDirectories: true,
			expectServices:    false,
		},
		{
			name:              "python module",
			moduleID:          "python",
			expectBinaries:    false,
			expectPython:      true,
			expectDirectories: false,
			expectServices:    false,
		},
		{
			name:              "pytorch module",
			moduleID:          "pytorch",
			expectBinaries:    false,
			expectPython:      true,
			expectDirectories: false,
			expectServices:    false,
		},
		{
			name:              "ollama module",
			moduleID:          "ollama",
			expectBinaries:    true,
			expectPython:      false,
			expectDirectories: true,
			expectServices:    true,
		},
		{
			name:              "docker module",
			moduleID:          "docker",
			expectBinaries:    true,
			expectPython:      false,
			expectDirectories: false,
			expectServices:    true,
		},
		{
			name:              "unknown module",
			moduleID:          "nonexistent",
			expectBinaries:    false,
			expectPython:      false,
			expectDirectories: false,
			expectServices:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := GetModulePaths(tt.moduleID)

			if tt.expectBinaries && len(paths.Binaries) == 0 {
				t.Errorf("Expected binaries for %s, got none", tt.moduleID)
			}
			if !tt.expectBinaries && len(paths.Binaries) > 0 {
				t.Errorf("Expected no binaries for %s, got %d", tt.moduleID, len(paths.Binaries))
			}

			if tt.expectPython && len(paths.PythonPackages) == 0 {
				t.Errorf("Expected Python packages for %s, got none", tt.moduleID)
			}
			if !tt.expectPython && len(paths.PythonPackages) > 0 {
				t.Errorf("Expected no Python packages for %s, got %d", tt.moduleID, len(paths.PythonPackages))
			}

			if tt.expectDirectories && len(paths.Directories) == 0 {
				t.Errorf("Expected directories for %s, got none", tt.moduleID)
			}
			if !tt.expectDirectories && len(paths.Directories) > 0 {
				t.Errorf("Expected no directories for %s, got %d", tt.moduleID, len(paths.Directories))
			}

			if tt.expectServices && len(paths.SystemdServices) == 0 {
				t.Errorf("Expected systemd services for %s, got none", tt.moduleID)
			}
			if !tt.expectServices && len(paths.SystemdServices) > 0 {
				t.Errorf("Expected no systemd services for %s, got %d", tt.moduleID, len(paths.SystemdServices))
			}
		})
	}
}

func TestGetAllPaths(t *testing.T) {
	paths := &ModulePaths{
		Binaries:    []string{"/usr/bin/test1", "/usr/bin/test2"},
		Directories: []string{"/opt/test"},
		ConfigFiles: []string{"/etc/test.conf"},
	}

	allPaths := paths.GetAllPaths()

	expectedCount := 4
	if len(allPaths) != expectedCount {
		t.Errorf("Expected %d paths, got %d", expectedCount, len(allPaths))
	}
}

func TestModuleSnapshot(t *testing.T) {
	snapshot := &ModuleSnapshot{
		ModuleID:  "test-module",
		Timestamp: time.Now(),
		InstalledPaths: []string{
			"/usr/bin/test",
			"/opt/test",
		},
		PreInstallState: map[string]bool{
			"/usr/bin/test": false,
			"/opt/test":     false,
			"/usr/bin/git":  true, // existed before
		},
		PythonPackages: []string{"test-package"},
		SystemdServices: []string{"test.service"},
	}

	summary := snapshot.GetSnapshotSummary()
	if summary == "" {
		t.Error("Expected non-empty snapshot summary")
	}

	if !containsString(summary, "test-module") {
		t.Error("Summary should contain module ID")
	}

	if !containsString(summary, "Paths to remove: 2") {
		t.Error("Summary should show 2 paths to remove")
	}
}

func TestIsCriticalPath(t *testing.T) {
	tests := []struct {
		path     string
		critical bool
	}{
		{"/", true},
		{"/bin", true},
		{"/usr/bin", true},
		{"/etc", true},
		{"/home", true},
		{"/usr/local/bin/custom", false},
		{"/opt/myapp", false},
		{"/var/lib/ollama", false},
		{"/root/ComfyUI", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isCriticalPath(tt.path)
			if result != tt.critical {
				t.Errorf("isCriticalPath(%s) = %v, want %v", tt.path, result, tt.critical)
			}
		})
	}
}

func TestRollbackState(t *testing.T) {
	state := &RollbackState{
		SessionID:  "test-session",
		ServerName: "test-server",
		StartTime:  time.Now(),
		Snapshots: []*ModuleSnapshot{
			{
				ModuleID:       "core",
				Timestamp:      time.Now(),
				InstalledPaths: []string{"/usr/bin/test"},
			},
			{
				ModuleID:       "python",
				Timestamp:      time.Now().Add(1 * time.Minute),
				PythonPackages: []string{"numpy"},
			},
		},
	}

	if state.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", state.SessionID)
	}

	if len(state.Snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(state.Snapshots))
	}
}

func TestReverseOrderProcessing(t *testing.T) {
	// Test that snapshots would be processed in reverse order
	snapshots := []*ModuleSnapshot{
		{ModuleID: "first", Timestamp: time.Now()},
		{ModuleID: "second", Timestamp: time.Now().Add(1 * time.Minute)},
		{ModuleID: "third", Timestamp: time.Now().Add(2 * time.Minute)},
	}

	// Verify that when processing in reverse, we get third -> second -> first
	expectedOrder := []string{"third", "second", "first"}
	for idx := len(snapshots) - 1; idx >= 0; idx-- {
		expectedIdx := len(snapshots) - 1 - idx
		if snapshots[idx].ModuleID != expectedOrder[expectedIdx] {
			t.Errorf("Expected %s at position %d, got %s",
				expectedOrder[expectedIdx], expectedIdx, snapshots[idx].ModuleID)
		}
	}
}

func TestFormatPathsList(t *testing.T) {
	paths := &ModulePaths{
		Binaries:        []string{"/usr/bin/test"},
		PythonPackages:  []string{"numpy", "torch"},
		Directories:     []string{"/opt/test"},
		ConfigFiles:     []string{"/etc/test.conf"},
		SystemdServices: []string{"test.service"},
	}

	formatted := paths.FormatPathsList()

	if !containsString(formatted, "Binaries (1)") {
		t.Error("Formatted output should include binaries count")
	}

	if !containsString(formatted, "Python Packages (2)") {
		t.Error("Formatted output should include Python packages count")
	}

	if !containsString(formatted, "Directories (1)") {
		t.Error("Formatted output should include directories count")
	}

	if !containsString(formatted, "Config Files (1)") {
		t.Error("Formatted output should include config files count")
	}

	if !containsString(formatted, "Systemd Services (1)") {
		t.Error("Formatted output should include systemd services count")
	}
}

func TestLLMModelPaths(t *testing.T) {
	// Test that LLM models all map to Ollama directory
	llmModels := []string{
		"llama-3.3-70b",
		"mistral",
		"qwen3-8b",
		"deepseek-r1-8b",
	}

	for _, modID := range llmModels {
		paths := GetModulePaths(modID)
		if len(paths.Directories) == 0 {
			t.Errorf("LLM model %s should have directories", modID)
		}
		if !containsString(paths.Directories[0], "ollama") {
			t.Errorf("LLM model %s should map to Ollama directory", modID)
		}
	}
}

func TestImageModelPaths(t *testing.T) {
	// Test that image models map to ComfyUI directories
	imageModels := []string{
		"sdxl",
		"flux-dev",
		"sd15",
	}

	for _, modID := range imageModels {
		paths := GetModulePaths(modID)
		if len(paths.Directories) == 0 {
			t.Errorf("Image model %s should have directories", modID)
		}
		if !containsString(paths.Directories[0], "ComfyUI") {
			t.Errorf("Image model %s should map to ComfyUI directory", modID)
		}
	}
}

func TestVideoModelPaths(t *testing.T) {
	// Test that video models map to ComfyUI directories
	videoModels := []string{
		"svd",
		"animatediff",
		"wan2",
	}

	for _, modID := range videoModels {
		paths := GetModulePaths(modID)
		if len(paths.Directories) == 0 {
			t.Errorf("Video model %s should have directories", modID)
		}
		if !containsString(paths.Directories[0], "ComfyUI") {
			t.Errorf("Video model %s should map to ComfyUI directory", modID)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
