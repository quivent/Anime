package cmd

import (
	"testing"
)

func TestGetSourceServer(t *testing.T) {
	// Save original value
	original := srcServer
	defer func() { srcServer = original }()

	// Test with empty flag (should return default)
	srcServer = ""
	server := getSourceServer()
	if server != "alice" {
		t.Errorf("Expected default server 'alice', got %s", server)
	}

	// Test with custom flag
	srcServer = "bob"
	server = getSourceServer()
	if server != "bob" {
		t.Errorf("Expected server 'bob', got %s", server)
	}
}

func TestResolveRemotePathSource(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "with arg",
			args:     []string{"my/path"},
			expected: "my/path",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "", // Will check linked path, which returns empty in tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRemotePathSource(tt.args)
			if tt.args != nil && len(tt.args) > 0 {
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestSourceCommandsExist(t *testing.T) {
	// Verify source command is registered
	if sourceCmd == nil {
		t.Error("sourceCmd should not be nil")
	}

	// Check subcommands exist
	subcommands := []string{
		"push", "pull", "clone", "status", "sync",
		"link", "init", "list", "tree", "history",
		"rename", "delete",
	}

	for _, name := range subcommands {
		found := false
		for _, cmd := range sourceCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Subcommand '%s' not found", name)
		}
	}
}

func TestSourceFlags(t *testing.T) {
	// Check persistent flags
	serverFlag := sourceCmd.PersistentFlags().Lookup("server")
	if serverFlag == nil {
		t.Error("Expected --server flag")
	}

	dryRunFlag := sourceCmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected --dry-run flag")
	}

	// Check clone-specific flags
	forceFlag := srcCloneCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag on clone command")
	}
}
