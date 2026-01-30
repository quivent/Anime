package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewInstallationState(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, true, 4)

	if state.ServerName != "test-server" {
		t.Errorf("Expected ServerName to be 'test-server', got '%s'", state.ServerName)
	}

	if len(state.Modules) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(state.Modules))
	}

	if state.Status != StatusInProgress {
		t.Errorf("Expected status to be 'in_progress', got '%s'", state.Status)
	}

	if !state.Parallel {
		t.Error("Expected Parallel to be true")
	}

	if state.Jobs != 4 {
		t.Errorf("Expected Jobs to be 4, got %d", state.Jobs)
	}

	if state.CompletedModules == nil {
		t.Error("Expected CompletedModules to be initialized")
	}

	if state.FailedModules == nil {
		t.Error("Expected FailedModules to be initialized")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	// Create temporary state directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Mark some modules as completed
	state.MarkModuleCompleted("core")
	state.MarkModuleCompleted("pytorch")

	// Save state
	err := state.Save()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load state
	loadedState, err := LoadState("test-server")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if loadedState == nil {
		t.Fatal("Expected loaded state to not be nil")
	}

	if loadedState.ServerName != state.ServerName {
		t.Errorf("ServerName mismatch: expected '%s', got '%s'", state.ServerName, loadedState.ServerName)
	}

	if len(loadedState.Modules) != len(state.Modules) {
		t.Errorf("Modules count mismatch: expected %d, got %d", len(state.Modules), len(loadedState.Modules))
	}

	if len(loadedState.CompletedModules) != 2 {
		t.Errorf("Expected 2 completed modules, got %d", len(loadedState.CompletedModules))
	}

	if loadedState.Status != StatusInProgress {
		t.Errorf("Expected status to be 'in_progress', got '%s'", loadedState.Status)
	}
}

func TestLoadStateNonExistent(t *testing.T) {
	// Create temporary state directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Try to load non-existent state
	state, err := LoadState("non-existent-server")
	if err != nil {
		t.Errorf("Expected no error for non-existent state, got: %v", err)
	}

	if state != nil {
		t.Error("Expected nil state for non-existent server")
	}
}

func TestClearState(t *testing.T) {
	// Create temporary state directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	modules := []string{"core", "pytorch"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Save state
	err := state.Save()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify state file exists
	statePath, _ := GetStatePath("test-server")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file should exist after save")
	}

	// Clear state
	err = ClearState("test-server")
	if err != nil {
		t.Fatalf("Failed to clear state: %v", err)
	}

	// Verify state file is gone
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("State file should not exist after clear")
	}
}

func TestClearStateNonExistent(t *testing.T) {
	// Create temporary state directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Try to clear non-existent state
	err := ClearState("non-existent-server")
	if err != nil {
		t.Errorf("Expected no error clearing non-existent state, got: %v", err)
	}
}

func TestMarkModuleCompleted(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkModuleCompleted("core")

	if len(state.CompletedModules) != 1 {
		t.Errorf("Expected 1 completed module, got %d", len(state.CompletedModules))
	}

	if _, ok := state.CompletedModules["core"]; !ok {
		t.Error("Expected 'core' to be in completed modules")
	}

	if state.InProgressModule != "" {
		t.Errorf("Expected InProgressModule to be empty, got '%s'", state.InProgressModule)
	}
}

func TestMarkModuleFailed(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	testErr := os.ErrNotExist
	state.MarkModuleFailed("pytorch", testErr)

	if len(state.FailedModules) != 1 {
		t.Errorf("Expected 1 failed module, got %d", len(state.FailedModules))
	}

	failedModule, ok := state.FailedModules["pytorch"]
	if !ok {
		t.Fatal("Expected 'pytorch' to be in failed modules")
	}

	if failedModule.Error == "" {
		t.Error("Expected error message to be set")
	}

	if state.InProgressModule != "" {
		t.Errorf("Expected InProgressModule to be empty, got '%s'", state.InProgressModule)
	}
}

func TestMarkModuleStarted(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkModuleStarted("pytorch")

	if state.InProgressModule != "pytorch" {
		t.Errorf("Expected InProgressModule to be 'pytorch', got '%s'", state.InProgressModule)
	}
}

func TestGetPendingModules(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama", "comfyui"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Mark some as completed
	state.MarkModuleCompleted("core")
	state.MarkModuleCompleted("pytorch")

	pending := state.GetPendingModules()

	if len(pending) != 2 {
		t.Errorf("Expected 2 pending modules, got %d", len(pending))
	}

	expectedPending := map[string]bool{"ollama": true, "comfyui": true}
	for _, modID := range pending {
		if !expectedPending[modID] {
			t.Errorf("Unexpected pending module: %s", modID)
		}
	}
}

func TestGetProgress(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama", "comfyui"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Initially 0% progress
	if progress := state.GetProgress(); progress != 0 {
		t.Errorf("Expected 0%% progress, got %.2f%%", progress)
	}

	// Mark 2 out of 4 as completed (50%)
	state.MarkModuleCompleted("core")
	state.MarkModuleCompleted("pytorch")

	if progress := state.GetProgress(); progress != 50.0 {
		t.Errorf("Expected 50%% progress, got %.2f%%", progress)
	}

	// Mark all as completed (100%)
	state.MarkModuleCompleted("ollama")
	state.MarkModuleCompleted("comfyui")

	if progress := state.GetProgress(); progress != 100.0 {
		t.Errorf("Expected 100%% progress, got %.2f%%", progress)
	}
}

func TestGetCounts(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama", "comfyui"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkModuleCompleted("core")
	state.MarkModuleCompleted("pytorch")
	state.MarkModuleFailed("ollama", nil)

	if count := state.GetCompletedCount(); count != 2 {
		t.Errorf("Expected 2 completed modules, got %d", count)
	}

	if count := state.GetFailedCount(); count != 1 {
		t.Errorf("Expected 1 failed module, got %d", count)
	}

	if count := state.GetTotalCount(); count != 4 {
		t.Errorf("Expected 4 total modules, got %d", count)
	}
}

func TestIsModuleStatus(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkModuleCompleted("core")
	state.MarkModuleFailed("pytorch", nil)

	if !state.IsModuleCompleted("core") {
		t.Error("Expected 'core' to be completed")
	}

	if state.IsModuleCompleted("pytorch") {
		t.Error("Expected 'pytorch' to not be completed")
	}

	if !state.IsModuleFailed("pytorch") {
		t.Error("Expected 'pytorch' to be failed")
	}

	if state.IsModuleFailed("core") {
		t.Error("Expected 'core' to not be failed")
	}

	if !state.IsModulePending("ollama") {
		t.Error("Expected 'ollama' to be pending")
	}

	if state.IsModulePending("core") {
		t.Error("Expected 'core' to not be pending")
	}
}

func TestMarkInstallationComplete(t *testing.T) {
	modules := []string{"core", "pytorch"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkCompleted()

	if state.Status != StatusCompleted {
		t.Errorf("Expected status to be 'completed', got '%s'", state.Status)
	}

	if state.InProgressModule != "" {
		t.Errorf("Expected InProgressModule to be empty, got '%s'", state.InProgressModule)
	}
}

func TestMarkInstallationFailed(t *testing.T) {
	modules := []string{"core", "pytorch"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkFailed()

	if state.Status != StatusFailed {
		t.Errorf("Expected status to be 'failed', got '%s'", state.Status)
	}
}

func TestMarkInstallationCancelled(t *testing.T) {
	modules := []string{"core", "pytorch"}
	state := NewInstallationState("test-server", modules, false, 0)

	state.MarkCancelled()

	if state.Status != StatusCancelled {
		t.Errorf("Expected status to be 'cancelled', got '%s'", state.Status)
	}
}

func TestIsStale(t *testing.T) {
	modules := []string{"core", "pytorch"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Fresh state should not be stale
	if state.IsStale() {
		t.Error("Fresh state should not be stale")
	}

	// Manipulate LastUpdate to be over 1 hour ago
	state.LastUpdate = time.Now().Add(-2 * time.Hour)

	if !state.IsStale() {
		t.Error("State should be stale after 2 hours")
	}
}

func TestCanResume(t *testing.T) {
	modules := []string{"core", "pytorch", "ollama"}
	state := NewInstallationState("test-server", modules, false, 0)

	// Fresh state with pending modules should be resumable
	if !state.CanResume() {
		t.Error("State with pending modules should be resumable")
	}

	// Mark all as completed
	state.MarkModuleCompleted("core")
	state.MarkModuleCompleted("pytorch")
	state.MarkModuleCompleted("ollama")

	// No pending modules, should not be resumable
	if state.CanResume() {
		t.Error("State with no pending modules should not be resumable")
	}

	// Reset and mark as completed
	state = NewInstallationState("test-server", modules, false, 0)
	state.MarkCompleted()

	// Completed state should not be resumable even with pending modules
	if state.CanResume() {
		t.Error("Completed state should not be resumable")
	}
}

func TestGetStatePath(t *testing.T) {
	// Create temporary home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	path, err := GetStatePath("test-server")
	if err != nil {
		t.Fatalf("GetStatePath failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, ".config", "anime", "state", "test-server.json")
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}

	// Verify directory was created
	stateDir := filepath.Dir(path)
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		t.Error("State directory should be created")
	}
}

func TestStatePersistenceAcrossUpdates(t *testing.T) {
	// Create temporary state directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	modules := []string{"core", "pytorch", "ollama", "comfyui"}
	state := NewInstallationState("test-server", modules, true, 8)

	// Simulate installation progress
	state.MarkModuleStarted("core")
	state.Save()

	state.MarkModuleCompleted("core")
	state.Save()

	state.MarkModuleStarted("pytorch")
	state.Save()

	state.MarkModuleCompleted("pytorch")
	state.Save()

	state.MarkModuleStarted("ollama")
	state.Save()

	state.MarkModuleFailed("ollama", os.ErrPermission)
	state.Save()

	// Load and verify
	loadedState, err := LoadState("test-server")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if loadedState.GetCompletedCount() != 2 {
		t.Errorf("Expected 2 completed modules, got %d", loadedState.GetCompletedCount())
	}

	if loadedState.GetFailedCount() != 1 {
		t.Errorf("Expected 1 failed module, got %d", loadedState.GetFailedCount())
	}

	if !loadedState.Parallel {
		t.Error("Expected Parallel to be preserved")
	}

	if loadedState.Jobs != 8 {
		t.Errorf("Expected Jobs to be 8, got %d", loadedState.Jobs)
	}

	pending := loadedState.GetPendingModules()
	// Should have 2 pending modules: ollama (failed, needs retry) and comfyui (not started)
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending modules (ollama failed + comfyui not started), got %d: %v", len(pending), pending)
	}

	// Verify both ollama and comfyui are in pending
	pendingMap := make(map[string]bool)
	for _, modID := range pending {
		pendingMap[modID] = true
	}
	if !pendingMap["ollama"] {
		t.Error("Expected 'ollama' (failed) to be in pending modules")
	}
	if !pendingMap["comfyui"] {
		t.Error("Expected 'comfyui' (not started) to be in pending modules")
	}
}
