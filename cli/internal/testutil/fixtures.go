package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempConfigDir creates a temporary config directory for testing
func TempConfigDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "anime")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create temp config dir: %v", err)
	}

	return configDir
}

// TempConfigFile creates a temporary config file with the given content
func TempConfigFile(t *testing.T, content string) string {
	t.Helper()

	configDir := TempConfigDir(t)
	configPath := filepath.Join(configDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	return configPath
}

// SetTestConfigPath sets the HOME env var to use temp config dir
func SetTestConfigPath(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)

	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	return tmpDir
}

// SampleYAMLConfig returns a sample YAML config for testing
func SampleYAMLConfig() string {
	return `servers:
  - name: test-server
    host: 192.168.1.100
    user: ubuntu
    ssh_key: ~/.ssh/id_rsa
    cost_per_hour: 3.5
    modules:
      - core
      - pytorch
api_keys:
  anthropic: test-key-123
  openai: test-key-456
aliases:
  dev: test-server
shell_aliases:
  ll: ls -lah
collections:
  - name: test-collection
    path: /data/test
    type: image
    description: Test collection
    tags:
      - test
      - sample
users:
  - name: testuser
    path: /home/testuser
active_user: testuser
`
}

// MinimalYAMLConfig returns a minimal YAML config for testing
func MinimalYAMLConfig() string {
	return `servers: []
api_keys: {}
aliases: {}
collections: []
users: []
`
}
