// Package source provides rsync-based source control functionality
package source

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultServer = "llamah"
	BasePath      = "~"
	LinkFile      = ".cpm-link"
	HistoryFile   = ".cpm-history"
)

// Config holds source control configuration
type Config struct {
	Server         string
	DryRun         bool
	Force          bool
	IncludeGit     bool   // If true, include .git directory in sync (git-clone equivalent)
	KeyPath        string
	Cleanup        func()
	AbsolutePath   bool   // If true, remotePath is an absolute path (e.g., ~/foo or /foo)
	BasePathOverride string // If set, overrides the default BasePath constant
}

// GetBasePath returns the effective base path for this config
func (c *Config) GetBasePath() string {
	if c.BasePathOverride != "" {
		return c.BasePathOverride
	}
	return BasePath
}

// LinkInfo represents the link configuration
type LinkInfo struct {
	RemotePath string `json:"remote_path"`
	Server     string `json:"server,omitempty"`
	LinkedAt   string `json:"linked_at,omitempty"`
}

// Status represents the sync status
type Status struct {
	ToPush   []string
	ToPull   []string
	InSync   bool
	LinkedTo string
}

// HistoryEntry represents a push/pull event
type HistoryEntry struct {
	Timestamp string
	Hostname  string
	Action    string
}

// GetRsyncExcludes returns common exclude arguments for rsync
func GetRsyncExcludes() []string {
	return []string{
		"--exclude", ".git",
		"--exclude", "node_modules",
		"--exclude", "cpm_modules",
		"--exclude", "__pycache__",
		"--exclude", "*.pyc",
		"--exclude", ".env",
		"--exclude", "venv",
		"--exclude", ".venv",
		"--exclude", LinkFile,
		"--exclude", ".cpm-installed.json",
	}
}

// GetRsyncExcludesFor returns the appropriate excludes based on config
func GetRsyncExcludesFor(cfg *Config) []string {
	if cfg.IncludeGit {
		return GetRsyncExcludesKeepGit()
	}
	return GetRsyncExcludes()
}

// GetRsyncExcludesKeepGit returns exclude arguments that preserve .git
func GetRsyncExcludesKeepGit() []string {
	return []string{
		"--exclude", "node_modules",
		"--exclude", "cpm_modules",
		"--exclude", "__pycache__",
		"--exclude", "*.pyc",
		"--exclude", ".env",
		"--exclude", "venv",
		"--exclude", ".venv",
		"--exclude", LinkFile,
		"--exclude", ".cpm-installed.json",
	}
}

// GetLinkedPath reads the linked remote path from .cpm-link file
func GetLinkedPath() string {
	data, err := os.ReadFile(LinkFile)
	if err != nil {
		return ""
	}

	// Try JSON first
	var info LinkInfo
	if err := json.Unmarshal(data, &info); err == nil && info.RemotePath != "" {
		return info.RemotePath
	}

	// Fall back to plain text
	return strings.TrimSpace(string(data))
}

// GetLinkInfo reads detailed link info
func GetLinkInfo() (*LinkInfo, error) {
	data, err := os.ReadFile(LinkFile)
	if err != nil {
		return nil, err
	}

	var info LinkInfo
	if err := json.Unmarshal(data, &info); err != nil {
		// Fall back to plain text
		info.RemotePath = strings.TrimSpace(string(data))
	}

	return &info, nil
}

// SaveLink saves a link configuration
func SaveLink(remotePath, server string) error {
	info := LinkInfo{
		RemotePath: remotePath,
		Server:     server,
		LinkedAt:   time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(LinkFile, data, 0644)
}

// Push syncs local directory to remote
func Push(target, remotePath string, cfg *Config) error {
	var fullRemotePath string
	if cfg.AbsolutePath {
		// Use path as-is (absolute path like ~/foo or /foo)
		fullRemotePath = remotePath
	} else if remotePath != "" {
		fullRemotePath = filepath.Join(cfg.GetBasePath(), remotePath)
	} else {
		fullRemotePath = cfg.GetBasePath()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !cfg.DryRun {
		// Create remote directory
		mkdirCmd := exec.Command("ssh",
			"-i", cfg.KeyPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			target,
			fmt.Sprintf("mkdir -p %s", fullRemotePath),
		)
		if output, err := mkdirCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create remote directory: %w\n%s", err, string(output))
		}
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cfg.DryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, GetRsyncExcludesFor(cfg)...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Record history (skip for absolute paths as they may not have history support)
	if !cfg.DryRun && !cfg.AbsolutePath {
		recordHistory(target, fullRemotePath, cfg.KeyPath, "push")
	}

	return nil
}

// Pull syncs remote directory to local
func Pull(target, remotePath string, cfg *Config) error {
	var fullRemotePath string
	if cfg.AbsolutePath {
		// Use path as-is (absolute path like ~/foo or /foo)
		fullRemotePath = remotePath
	} else if remotePath != "" {
		fullRemotePath = filepath.Join(cfg.GetBasePath(), remotePath)
	} else {
		fullRemotePath = cfg.GetBasePath()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cfg.DryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, GetRsyncExcludesFor(cfg)...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	return nil
}

// Clone clones a remote repo into a new folder
func Clone(target, remotePath string, cfg *Config) error {
	fullRemotePath := filepath.Join(cfg.GetBasePath(), remotePath)
	folderName := filepath.Base(remotePath)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	localPath := filepath.Join(cwd, folderName)

	// Check if destination exists
	if _, err := os.Stat(localPath); err == nil {
		if cfg.Force {
			if !cfg.DryRun {
				os.RemoveAll(localPath)
			}
		} else {
			return fmt.Errorf("destination folder already exists: %s (use --force to overwrite)", folderName)
		}
	}

	if !cfg.DryRun {
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return fmt.Errorf("failed to create destination folder: %w", err)
		}
	}

	// Rsync
	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)
	rsyncArgs := []string{"-avz", "--progress"}
	if cfg.DryRun {
		rsyncArgs = append(rsyncArgs, "--dry-run")
	}
	rsyncArgs = append(rsyncArgs, GetRsyncExcludesFor(cfg)...)
	rsyncArgs = append(rsyncArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", localPath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		if !cfg.DryRun {
			os.RemoveAll(localPath)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	// Create link file in cloned directory
	if !cfg.DryRun {
		linkPath := filepath.Join(localPath, LinkFile)
		os.WriteFile(linkPath, []byte(remotePath), 0644)
	}

	return nil
}

// GetStatus compares local and remote directories
func GetStatus(target, remotePath string, cfg *Config) (*Status, error) {
	fullRemotePath := filepath.Join(cfg.GetBasePath(), remotePath)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)

	status := &Status{
		LinkedTo: remotePath,
	}

	// Check what would be pushed
	pushArgs := []string{"-avzn", "--out-format", "%n"}
	pushArgs = append(pushArgs, GetRsyncExcludesFor(cfg)...)
	pushArgs = append(pushArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")
	pushCmd := exec.Command("rsync", pushArgs...)
	pushOutput, _ := pushCmd.CombinedOutput()
	status.ToPush = filterRsyncOutput(string(pushOutput))

	// Check what would be pulled
	pullArgs := []string{"-avzn", "--out-format", "%n"}
	pullArgs = append(pullArgs, GetRsyncExcludesFor(cfg)...)
	pullArgs = append(pullArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")
	pullCmd := exec.Command("rsync", pullArgs...)
	pullOutput, _ := pullCmd.CombinedOutput()
	status.ToPull = filterRsyncOutput(string(pullOutput))

	status.InSync = len(status.ToPush) == 0 && len(status.ToPull) == 0

	return status, nil
}

// Sync performs bidirectional sync
func Sync(target, remotePath string, cfg *Config) error {
	fullRemotePath := filepath.Join(cfg.GetBasePath(), remotePath)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	rsyncSSH := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", cfg.KeyPath)

	// Step 1: Pull newer files from remote
	pullArgs := []string{"-avz", "--progress", "--update"}
	if cfg.DryRun {
		pullArgs = append(pullArgs, "--dry-run")
	}
	pullArgs = append(pullArgs, GetRsyncExcludesFor(cfg)...)
	pullArgs = append(pullArgs, "-e", rsyncSSH, target+":"+fullRemotePath+"/", cwd+"/")

	pullCmd := exec.Command("rsync", pullArgs...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("sync pull failed: %w", err)
	}

	// Step 2: Push newer local files to remote
	pushArgs := []string{"-avz", "--progress", "--update"}
	if cfg.DryRun {
		pushArgs = append(pushArgs, "--dry-run")
	}
	pushArgs = append(pushArgs, GetRsyncExcludesFor(cfg)...)
	pushArgs = append(pushArgs, "-e", rsyncSSH, cwd+"/", target+":"+fullRemotePath+"/")

	pushCmd := exec.Command("rsync", pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("sync push failed: %w", err)
	}

	return nil
}

// ListRepos lists repositories on remote
func ListRepos(target, path string, cfg *Config) ([]string, error) {
	fullPath := cfg.GetBasePath()
	if path != "" {
		fullPath = filepath.Join(cfg.GetBasePath(), path)
	}

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("ls -1 %s/ 2>/dev/null || echo ''", fullPath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			repos = append(repos, line)
		}
	}

	return repos, nil
}

// Delete removes a repo from remote
func Delete(target, remotePath string, cfg *Config) error {
	fullRemotePath := filepath.Join(cfg.GetBasePath(), remotePath)

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("rm -rf %s", fullRemotePath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("delete failed: %w\n%s", err, string(output))
	}

	return nil
}

// Rename moves a repo on remote
func Rename(target, oldPath, newPath string, cfg *Config) error {
	fullOldPath := filepath.Join(cfg.GetBasePath(), oldPath)
	fullNewPath := filepath.Join(cfg.GetBasePath(), newPath)

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("mkdir -p $(dirname %s) && mv %s %s", fullNewPath, fullOldPath, fullNewPath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rename failed: %w\n%s", err, string(output))
	}

	return nil
}

// GetHistory retrieves push history for a repo
func GetHistory(target, remotePath string, cfg *Config) ([]HistoryEntry, error) {
	fullRemotePath := filepath.Join(cfg.GetBasePath(), remotePath)
	historyPath := filepath.Join(fullRemotePath, HistoryFile)

	sshCmd := exec.Command("ssh",
		"-i", cfg.KeyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("cat %s 2>/dev/null | tail -20 || echo ''", historyPath),
	)
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			entries = append(entries, HistoryEntry{
				Timestamp: parts[0],
				Hostname:  parts[1],
				Action:    parts[2],
			})
		}
	}

	return entries, nil
}

// Helper functions

func recordHistory(target, remotePath, keyPath, action string) error {
	historyPath := filepath.Join(remotePath, HistoryFile)
	timestamp := time.Now().Format(time.RFC3339)
	hostname, _ := os.Hostname()
	entry := fmt.Sprintf("%s|%s|%s", timestamp, hostname, action)

	sshCmd := exec.Command("ssh",
		"-i", keyPath,
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		target,
		fmt.Sprintf("echo '%s' >> %s", entry, historyPath),
	)
	return sshCmd.Run()
}

func filterRsyncOutput(output string) []string {
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "sending") || strings.HasPrefix(line, "receiving") ||
			strings.HasPrefix(line, "total") || strings.HasPrefix(line, "sent") ||
			strings.HasSuffix(line, "/") || line == "." || line == "./" {
			continue
		}
		files = append(files, line)
	}
	return files
}
