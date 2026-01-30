package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InstallationStatus represents the current status of an installation
type InstallationStatus string

const (
	StatusInProgress InstallationStatus = "in_progress"
	StatusCompleted  InstallationStatus = "completed"
	StatusFailed     InstallationStatus = "failed"
	StatusCancelled  InstallationStatus = "cancelled"
)

// ModuleState represents the state of a single module
type ModuleState struct {
	Status      string    `json:"status"`       // Starting, Installing, Complete, Failed
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Error       string    `json:"error,omitempty"`
	LastOutput  string    `json:"last_output,omitempty"`
}

// InstallationState tracks the state of an installation session
type InstallationState struct {
	ServerName        string                    `json:"server_name"`
	StartTime         time.Time                 `json:"start_time"`
	LastUpdate        time.Time                 `json:"last_update"`
	Modules           []string                  `json:"modules"`            // All modules to install (in order)
	CompletedModules  map[string]ModuleState    `json:"completed_modules"`  // Successfully completed modules
	FailedModules     map[string]ModuleState    `json:"failed_modules"`     // Failed modules
	InProgressModule  string                    `json:"in_progress_module,omitempty"` // Currently installing module
	Status            InstallationStatus        `json:"status"`
	Parallel          bool                      `json:"parallel,omitempty"`
	Jobs              int                       `json:"jobs,omitempty"`
}

// NewInstallationState creates a new installation state
func NewInstallationState(serverName string, modules []string, parallel bool, jobs int) *InstallationState {
	return &InstallationState{
		ServerName:       serverName,
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
		Modules:          modules,
		CompletedModules: make(map[string]ModuleState),
		FailedModules:    make(map[string]ModuleState),
		Status:           StatusInProgress,
		Parallel:         parallel,
		Jobs:             jobs,
	}
}

// GetStatePath returns the path to the state file for a server
func GetStatePath(serverName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(home, ".config", "anime", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create state directory: %w", err)
	}

	return filepath.Join(stateDir, fmt.Sprintf("%s.json", serverName)), nil
}

// Save persists the state to disk
func (s *InstallationState) Save() error {
	s.LastUpdate = time.Now()

	statePath, err := GetStatePath(s.ServerName)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads an existing state from disk
func LoadState(serverName string) (*InstallationState, error) {
	statePath, err := GetStatePath(serverName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file exists
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state InstallationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// ClearState removes the state file for a server
func ClearState(serverName string) error {
	statePath, err := GetStatePath(serverName)
	if err != nil {
		return err
	}

	if err := os.Remove(statePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already doesn't exist
		}
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	return nil
}

// MarkModuleStarted marks a module as started
func (s *InstallationState) MarkModuleStarted(moduleID string) {
	s.InProgressModule = moduleID
	s.LastUpdate = time.Now()
}

// MarkModuleCompleted marks a module as completed
func (s *InstallationState) MarkModuleCompleted(moduleID string) {
	if s.CompletedModules == nil {
		s.CompletedModules = make(map[string]ModuleState)
	}

	s.CompletedModules[moduleID] = ModuleState{
		Status:    "Complete",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	s.InProgressModule = ""
	s.LastUpdate = time.Now()
}

// MarkModuleFailed marks a module as failed
func (s *InstallationState) MarkModuleFailed(moduleID string, err error) {
	if s.FailedModules == nil {
		s.FailedModules = make(map[string]ModuleState)
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	s.FailedModules[moduleID] = ModuleState{
		Status:    "Failed",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Error:     errMsg,
	}
	s.InProgressModule = ""
	s.LastUpdate = time.Now()
}

// MarkCompleted marks the entire installation as completed
func (s *InstallationState) MarkCompleted() {
	s.Status = StatusCompleted
	s.InProgressModule = ""
	s.LastUpdate = time.Now()
}

// MarkFailed marks the entire installation as failed
func (s *InstallationState) MarkFailed() {
	s.Status = StatusFailed
	s.InProgressModule = ""
	s.LastUpdate = time.Now()
}

// MarkCancelled marks the entire installation as cancelled
func (s *InstallationState) MarkCancelled() {
	s.Status = StatusCancelled
	s.InProgressModule = ""
	s.LastUpdate = time.Now()
}

// GetPendingModules returns modules that haven't been completed yet
func (s *InstallationState) GetPendingModules() []string {
	pending := []string{}
	for _, modID := range s.Modules {
		if _, completed := s.CompletedModules[modID]; !completed {
			pending = append(pending, modID)
		}
	}
	return pending
}

// GetCompletedCount returns the number of completed modules
func (s *InstallationState) GetCompletedCount() int {
	return len(s.CompletedModules)
}

// GetFailedCount returns the number of failed modules
func (s *InstallationState) GetFailedCount() int {
	return len(s.FailedModules)
}

// GetTotalCount returns the total number of modules
func (s *InstallationState) GetTotalCount() int {
	return len(s.Modules)
}

// GetProgress returns the installation progress as a percentage
func (s *InstallationState) GetProgress() float64 {
	if len(s.Modules) == 0 {
		return 0
	}
	return float64(len(s.CompletedModules)) / float64(len(s.Modules)) * 100
}

// IsModuleCompleted checks if a module is completed
func (s *InstallationState) IsModuleCompleted(moduleID string) bool {
	_, ok := s.CompletedModules[moduleID]
	return ok
}

// IsModuleFailed checks if a module has failed
func (s *InstallationState) IsModuleFailed(moduleID string) bool {
	_, ok := s.FailedModules[moduleID]
	return ok
}

// IsModulePending checks if a module is pending
func (s *InstallationState) IsModulePending(moduleID string) bool {
	return !s.IsModuleCompleted(moduleID) && !s.IsModuleFailed(moduleID)
}

// GetElapsedTime returns the elapsed time since installation started
func (s *InstallationState) GetElapsedTime() time.Duration {
	return time.Since(s.StartTime)
}

// GetTimeSinceLastUpdate returns the time since the last update
func (s *InstallationState) GetTimeSinceLastUpdate() time.Duration {
	return time.Since(s.LastUpdate)
}

// IsStale checks if the state is stale (no update in over 1 hour)
func (s *InstallationState) IsStale() bool {
	return s.GetTimeSinceLastUpdate() > time.Hour
}

// CanResume checks if the installation can be resumed
func (s *InstallationState) CanResume() bool {
	return s.Status == StatusInProgress && len(s.GetPendingModules()) > 0
}
