package installer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/ssh"
)

// ModuleSnapshot captures the state before module installation
type ModuleSnapshot struct {
	ModuleID        string            `json:"module_id"`
	Timestamp       time.Time         `json:"timestamp"`
	InstalledPaths  []string          `json:"installed_paths"`
	PreInstallState map[string]bool   `json:"pre_install_state"`
	PythonPackages  []string          `json:"python_packages"`
	SystemdServices []string          `json:"systemd_services"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// RollbackState tracks all snapshots for a deployment session
type RollbackState struct {
	SessionID string            `json:"session_id"`
	ServerName string           `json:"server_name"`
	StartTime time.Time         `json:"start_time"`
	Snapshots []*ModuleSnapshot `json:"snapshots"`
}

// TakeSnapshot captures the current state before installing a module
func (i *Installer) TakeSnapshot(modID string) (*ModuleSnapshot, error) {
	snapshot := &ModuleSnapshot{
		ModuleID:        modID,
		Timestamp:       time.Now(),
		PreInstallState: make(map[string]bool),
		Metadata:        make(map[string]string),
	}

	// Get expected paths for this module
	paths := GetModulePaths(modID)

	// Check which paths already exist
	allPaths := paths.GetAllPaths()
	for _, path := range allPaths {
		exists, err := i.pathExists(path)
		if err != nil {
			// Log warning but continue - don't fail snapshot on path check errors
			i.sendProgress(modID, "Snapshot", fmt.Sprintf("Warning: couldn't check path %s: %v", path, err), nil, false)
			continue
		}
		snapshot.PreInstallState[path] = exists
		if !exists {
			// Track paths that don't exist yet - these will be removed on rollback
			snapshot.InstalledPaths = append(snapshot.InstalledPaths, path)
		}
	}

	// Track python packages that don't exist yet
	for _, pkg := range paths.PythonPackages {
		exists, err := i.pythonPackageExists(pkg)
		if err != nil {
			i.sendProgress(modID, "Snapshot", fmt.Sprintf("Warning: couldn't check package %s: %v", pkg, err), nil, false)
			continue
		}
		if !exists {
			snapshot.PythonPackages = append(snapshot.PythonPackages, pkg)
		}
	}

	// Track systemd services
	snapshot.SystemdServices = paths.SystemdServices

	i.sendProgress(modID, "Snapshot", fmt.Sprintf("Captured snapshot: %d paths, %d packages",
		len(snapshot.InstalledPaths), len(snapshot.PythonPackages)), nil, false)

	return snapshot, nil
}

// Rollback removes installed components based on snapshots
// Processes snapshots in reverse order (LIFO) to respect dependencies
func (i *Installer) Rollback(snapshots []*ModuleSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	var errors []string

	i.sendProgress("", "Rollback", fmt.Sprintf("Starting rollback of %d modules", len(snapshots)), nil, false)

	// Process in reverse order (last installed first)
	for idx := len(snapshots) - 1; idx >= 0; idx-- {
		snapshot := snapshots[idx]

		i.sendProgress(snapshot.ModuleID, "Rollback", fmt.Sprintf("Rolling back %s", snapshot.ModuleID), nil, false)

		// Uninstall Python packages
		if len(snapshot.PythonPackages) > 0 {
			if err := i.uninstallPythonPackages(snapshot.ModuleID, snapshot.PythonPackages); err != nil {
				errors = append(errors, fmt.Sprintf("failed to uninstall packages for %s: %v", snapshot.ModuleID, err))
				// Continue with rollback even if package uninstall fails
			}
		}

		// Stop and disable systemd services
		if len(snapshot.SystemdServices) > 0 {
			if err := i.stopSystemdServices(snapshot.ModuleID, snapshot.SystemdServices); err != nil {
				errors = append(errors, fmt.Sprintf("failed to stop services for %s: %v", snapshot.ModuleID, err))
				// Continue - services may not be running
			}
		}

		// Remove files and directories that didn't exist before installation
		if len(snapshot.InstalledPaths) > 0 {
			if err := i.removePaths(snapshot.ModuleID, snapshot.InstalledPaths, snapshot.PreInstallState); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove paths for %s: %v", snapshot.ModuleID, err))
				// Continue with rollback
			}
		}

		i.sendProgress(snapshot.ModuleID, "Rollback", fmt.Sprintf("Rolled back %s", snapshot.ModuleID), nil, true)
	}

	if len(errors) > 0 {
		i.sendProgress("", "Rollback", "Rollback completed with errors", nil, false)
		return fmt.Errorf("rollback errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	i.sendProgress("", "Rollback", "Rollback completed successfully", nil, true)
	return nil
}

// pathExists checks if a path exists on the remote server
func (i *Installer) pathExists(path string) (bool, error) {
	// Handle glob patterns in paths (e.g., /usr/lib/nvidia-*)
	if strings.Contains(path, "*") {
		// Use ls to check for glob matches
		output, err := i.client.RunCommand(fmt.Sprintf("ls -d %s 2>/dev/null | wc -l", path))
		if err != nil {
			// Command failed - assume doesn't exist
			return false, nil
		}
		count := strings.TrimSpace(output)
		return count != "0", nil
	}

	// Regular path check
	output, err := i.client.RunCommand(fmt.Sprintf("test -e %s && echo exists || echo missing", path))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "exists", nil
}

// pythonPackageExists checks if a Python package is installed
func (i *Installer) pythonPackageExists(pkg string) (bool, error) {
	output, err := i.client.RunCommand(fmt.Sprintf("pip3 show %s >/dev/null 2>&1 && echo exists || echo missing", pkg))
	if err != nil {
		// Command execution error - return error
		return false, err
	}
	return strings.TrimSpace(output) == "exists", nil
}

// uninstallPythonPackages removes Python packages
func (i *Installer) uninstallPythonPackages(modID string, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	i.sendProgress(modID, "Rollback", fmt.Sprintf("Uninstalling %d Python packages", len(packages)), nil, false)

	// Uninstall packages in batch
	pkgList := strings.Join(packages, " ")
	cmd := fmt.Sprintf("pip3 uninstall -y %s 2>&1 || true", pkgList)

	output, err := i.client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to uninstall packages: %w\nOutput: %s", err, output)
	}

	i.sendProgress(modID, "Rollback", fmt.Sprintf("Uninstalled packages: %s", strings.Join(packages, ", ")), nil, false)
	return nil
}

// stopSystemdServices stops and disables systemd services
func (i *Installer) stopSystemdServices(modID string, services []string) error {
	if len(services) == 0 {
		return nil
	}

	i.sendProgress(modID, "Rollback", fmt.Sprintf("Stopping %d services", len(services)), nil, false)

	var errors []string
	for _, service := range services {
		// Stop the service
		cmd := fmt.Sprintf("sudo systemctl stop %s 2>&1 || true", service)
		if _, err := i.client.RunCommand(cmd); err != nil {
			errors = append(errors, fmt.Sprintf("failed to stop %s: %v", service, err))
			continue
		}

		// Disable the service
		cmd = fmt.Sprintf("sudo systemctl disable %s 2>&1 || true", service)
		if _, err := i.client.RunCommand(cmd); err != nil {
			errors = append(errors, fmt.Sprintf("failed to disable %s: %v", service, err))
			continue
		}

		i.sendProgress(modID, "Rollback", fmt.Sprintf("Stopped service: %s", service), nil, false)
	}

	if len(errors) > 0 {
		return fmt.Errorf("service errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// removePaths removes files and directories that were created during installation
func (i *Installer) removePaths(modID string, paths []string, preInstallState map[string]bool) error {
	if len(paths) == 0 {
		return nil
	}

	i.sendProgress(modID, "Rollback", fmt.Sprintf("Removing %d paths", len(paths)), nil, false)

	var errors []string
	for _, path := range paths {
		// Double-check: only remove if it didn't exist before
		if preInstallState[path] {
			i.sendProgress(modID, "Rollback", fmt.Sprintf("Skipping %s (existed before installation)", path), nil, false)
			continue
		}

		// Safety check: don't remove critical system paths
		if isCriticalPath(path) {
			i.sendProgress(modID, "Rollback", fmt.Sprintf("Skipping critical path: %s", path), nil, false)
			continue
		}

		// Check if path exists before trying to remove
		exists, err := i.pathExists(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to check %s: %v", path, err))
			continue
		}

		if !exists {
			// Already removed or never existed
			continue
		}

		// Remove the path
		var cmd string
		if strings.Contains(path, "*") {
			// Glob pattern - use rm with glob
			cmd = fmt.Sprintf("sudo rm -rf %s 2>&1 || true", path)
		} else {
			cmd = fmt.Sprintf("sudo rm -rf %s 2>&1 || true", path)
		}

		output, err := i.client.RunCommand(cmd)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to remove %s: %v\nOutput: %s", path, err, output))
			continue
		}

		i.sendProgress(modID, "Rollback", fmt.Sprintf("Removed: %s", path), nil, false)
	}

	if len(errors) > 0 {
		return fmt.Errorf("path removal errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// isCriticalPath returns true for paths that should never be removed
func isCriticalPath(path string) bool {
	criticalPaths := []string{
		"/",
		"/bin",
		"/sbin",
		"/usr",
		"/usr/bin",
		"/usr/sbin",
		"/usr/lib",
		"/lib",
		"/lib64",
		"/etc",
		"/home",
		"/root",
		"/var",
		"/boot",
		"/dev",
		"/proc",
		"/sys",
	}

	for _, critical := range criticalPaths {
		if path == critical {
			return true
		}
	}

	return false
}

// SaveRollbackState saves the rollback state to a file on the remote server
func (i *Installer) SaveRollbackState(serverName string, snapshots []*ModuleSnapshot) error {
	state := &RollbackState{
		SessionID:  fmt.Sprintf("%s-%d", serverName, time.Now().Unix()),
		ServerName: serverName,
		StartTime:  time.Now(),
		Snapshots:  snapshots,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rollback state: %w", err)
	}

	// Save to remote server
	remotePath := fmt.Sprintf("/tmp/anime-rollback-%s.json", state.SessionID)
	if err := i.client.UploadString(string(data), remotePath); err != nil {
		return fmt.Errorf("failed to save rollback state: %w", err)
	}

	i.sendProgress("", "Rollback", fmt.Sprintf("Saved rollback state to %s", remotePath), nil, false)
	return nil
}

// LoadRollbackState loads a rollback state from the remote server
func LoadRollbackState(client *ssh.Client, sessionID string) (*RollbackState, error) {
	remotePath := fmt.Sprintf("/tmp/anime-rollback-%s.json", sessionID)

	output, err := client.RunCommand(fmt.Sprintf("cat %s 2>/dev/null || echo '{}'", remotePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read rollback state: %w", err)
	}

	var state RollbackState
	if err := json.Unmarshal([]byte(output), &state); err != nil {
		return nil, fmt.Errorf("failed to parse rollback state: %w", err)
	}

	return &state, nil
}

// ListRollbackStates lists all available rollback states on the remote server
func ListRollbackStates(client *ssh.Client) ([]*RollbackState, error) {
	output, err := client.RunCommand("ls -1 /tmp/anime-rollback-*.json 2>/dev/null || true")
	if err != nil {
		return nil, fmt.Errorf("failed to list rollback states: %w", err)
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	var states []*RollbackState

	for _, file := range files {
		if file == "" {
			continue
		}

		// Extract session ID from filename
		base := filepath.Base(file)
		sessionID := strings.TrimPrefix(base, "anime-rollback-")
		sessionID = strings.TrimSuffix(sessionID, ".json")

		state, err := LoadRollbackState(client, sessionID)
		if err != nil {
			// Skip invalid state files
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

// DeleteRollbackState removes a rollback state file from the remote server
func DeleteRollbackState(client *ssh.Client, sessionID string) error {
	remotePath := fmt.Sprintf("/tmp/anime-rollback-%s.json", sessionID)
	_, err := client.RunCommand(fmt.Sprintf("rm -f %s", remotePath))
	return err
}

// GetSnapshotSummary returns a human-readable summary of a snapshot
func (s *ModuleSnapshot) GetSnapshotSummary() string {
	summary := fmt.Sprintf("Module: %s\n", s.ModuleID)
	summary += fmt.Sprintf("Timestamp: %s\n", s.Timestamp.Format(time.RFC3339))
	summary += fmt.Sprintf("Paths to remove: %d\n", len(s.InstalledPaths))
	summary += fmt.Sprintf("Packages to uninstall: %d\n", len(s.PythonPackages))
	summary += fmt.Sprintf("Services to stop: %d\n", len(s.SystemdServices))
	return summary
}

// GetRollbackPreview returns a preview of what will be removed during rollback
func (i *Installer) GetRollbackPreview(snapshots []*ModuleSnapshot) string {
	var preview strings.Builder

	preview.WriteString(fmt.Sprintf("Rollback Preview (%d modules):\n\n", len(snapshots)))

	for idx := len(snapshots) - 1; idx >= 0; idx-- {
		snapshot := snapshots[idx]
		preview.WriteString(fmt.Sprintf("Module: %s (installed at %s)\n",
			snapshot.ModuleID, snapshot.Timestamp.Format("2006-01-02 15:04:05")))

		if len(snapshot.PythonPackages) > 0 {
			preview.WriteString(fmt.Sprintf("  Python packages to uninstall: %s\n",
				strings.Join(snapshot.PythonPackages, ", ")))
		}

		if len(snapshot.SystemdServices) > 0 {
			preview.WriteString(fmt.Sprintf("  Services to stop: %s\n",
				strings.Join(snapshot.SystemdServices, ", ")))
		}

		if len(snapshot.InstalledPaths) > 0 {
			preview.WriteString(fmt.Sprintf("  Paths to remove (%d):\n", len(snapshot.InstalledPaths)))
			for _, path := range snapshot.InstalledPaths {
				preview.WriteString(fmt.Sprintf("    - %s\n", path))
			}
		}

		preview.WriteString("\n")
	}

	return preview.String()
}
