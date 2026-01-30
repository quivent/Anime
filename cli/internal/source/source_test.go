package source

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetRsyncExcludes(t *testing.T) {
	excludes := GetRsyncExcludes()

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

func TestGetLinkedPath(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "source_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with no link file
	path := GetLinkedPath()
	if path != "" {
		t.Errorf("Expected empty path with no link file, got %s", path)
	}

	// Test with plain text link file
	os.WriteFile(LinkFile, []byte("test/path"), 0644)
	path = GetLinkedPath()
	if path != "test/path" {
		t.Errorf("Expected 'test/path', got %s", path)
	}

	// Test with JSON link file
	info := LinkInfo{RemotePath: "json/path", Server: "myserver"}
	data, _ := json.Marshal(info)
	os.WriteFile(LinkFile, data, 0644)
	path = GetLinkedPath()
	if path != "json/path" {
		t.Errorf("Expected 'json/path', got %s", path)
	}
}

func TestGetLinkInfo(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "source_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with no link file
	_, err = GetLinkInfo()
	if err == nil {
		t.Error("Expected error with no link file")
	}

	// Test with JSON link file
	info := LinkInfo{RemotePath: "test/path", Server: "myserver", LinkedAt: "2024-01-01"}
	data, _ := json.Marshal(info)
	os.WriteFile(LinkFile, data, 0644)

	result, err := GetLinkInfo()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.RemotePath != "test/path" {
		t.Errorf("Expected RemotePath 'test/path', got %s", result.RemotePath)
	}
	if result.Server != "myserver" {
		t.Errorf("Expected Server 'myserver', got %s", result.Server)
	}
}

func TestSaveLink(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "source_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Save link
	err = SaveLink("my/remote/path", "testserver")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(LinkFile)
	if err != nil {
		t.Errorf("Failed to read link file: %v", err)
	}

	var info LinkInfo
	if err := json.Unmarshal(data, &info); err != nil {
		t.Errorf("Failed to parse link file: %v", err)
	}

	if info.RemotePath != "my/remote/path" {
		t.Errorf("Expected RemotePath 'my/remote/path', got %s", info.RemotePath)
	}
	if info.Server != "testserver" {
		t.Errorf("Expected Server 'testserver', got %s", info.Server)
	}
	if info.LinkedAt == "" {
		t.Error("Expected LinkedAt to be set")
	}
}

func TestFilterRsyncOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty output",
			input:    "",
			expected: 0,
		},
		{
			name:     "only summary lines",
			input:    "sending incremental file list\ntotal size is 100",
			expected: 0,
		},
		{
			name:     "files only",
			input:    "file1.txt\nfile2.txt\nfile3.txt",
			expected: 3,
		},
		{
			name:     "mixed with directories",
			input:    "file1.txt\ndir/\nfile2.txt",
			expected: 2, // directories (ending with /) are excluded
		},
		{
			name:     "with summary and files",
			input:    "sending incremental file list\nfile1.txt\nfile2.txt\ntotal size is 100",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRsyncOutput(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected %d files, got %d: %v", tt.expected, len(result), result)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	if DefaultServer != "alice" {
		t.Errorf("Expected DefaultServer 'alice', got %s", DefaultServer)
	}
	if BasePath != "~/cpm/anime" {
		t.Errorf("Expected BasePath '~/cpm/anime', got %s", BasePath)
	}
	if LinkFile != ".cpm-link" {
		t.Errorf("Expected LinkFile '.cpm-link', got %s", LinkFile)
	}
	if HistoryFile != ".cpm-history" {
		t.Errorf("Expected HistoryFile '.cpm-history', got %s", HistoryFile)
	}
}

func TestHistoryEntry(t *testing.T) {
	entry := HistoryEntry{
		Timestamp: "2024-01-01T12:00:00Z",
		Hostname:  "testhost",
		Action:    "push",
	}

	if entry.Timestamp != "2024-01-01T12:00:00Z" {
		t.Error("Timestamp mismatch")
	}
	if entry.Hostname != "testhost" {
		t.Error("Hostname mismatch")
	}
	if entry.Action != "push" {
		t.Error("Action mismatch")
	}
}

func TestStatus(t *testing.T) {
	status := &Status{
		ToPush:   []string{"file1.txt", "file2.txt"},
		ToPull:   []string{"file3.txt"},
		InSync:   false,
		LinkedTo: "test/path",
	}

	if len(status.ToPush) != 2 {
		t.Errorf("Expected 2 ToPush, got %d", len(status.ToPush))
	}
	if len(status.ToPull) != 1 {
		t.Errorf("Expected 1 ToPull, got %d", len(status.ToPull))
	}
	if status.InSync {
		t.Error("Expected InSync to be false")
	}
	if status.LinkedTo != "test/path" {
		t.Errorf("Expected LinkedTo 'test/path', got %s", status.LinkedTo)
	}
}

func TestConfig(t *testing.T) {
	cleanupCalled := false
	cfg := &Config{
		Server:  "testserver",
		DryRun:  true,
		Force:   false,
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

	cfg.Cleanup()
	if !cleanupCalled {
		t.Error("Cleanup was not called")
	}
}

// Integration test helpers
func setupTestDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "source_integration_test")
	if err != nil {
		t.Fatal(err)
	}

	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("content3"), 0644)

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func TestLinkInfoJSON(t *testing.T) {
	info := LinkInfo{
		RemotePath: "org/project",
		Server:     "alice",
		LinkedAt:   "2024-01-01T12:00:00Z",
	}

	// Test JSON marshaling
	data, err := json.Marshal(info)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	// Test JSON unmarshaling
	var parsed LinkInfo
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	if parsed.RemotePath != info.RemotePath {
		t.Errorf("RemotePath mismatch: %s vs %s", parsed.RemotePath, info.RemotePath)
	}
	if parsed.Server != info.Server {
		t.Errorf("Server mismatch: %s vs %s", parsed.Server, info.Server)
	}
}
