package cmd

import (
	"testing"
)

func TestGetPkgServer(t *testing.T) {
	// Save original value
	original := pkgServer
	defer func() { pkgServer = original }()

	// Test with empty flag (should return default)
	pkgServer = ""
	server := getPkgServer()
	if server != "alice" {
		t.Errorf("Expected default server 'alice', got %s", server)
	}

	// Test with custom flag
	pkgServer = "charlie"
	server = getPkgServer()
	if server != "charlie" {
		t.Errorf("Expected server 'charlie', got %s", server)
	}
}

func TestPkgCommandsExist(t *testing.T) {
	// Verify pkg command is registered
	if pkgCmd == nil {
		t.Error("pkgCmd should not be nil")
	}

	// Check subcommands exist
	subcommands := []string{
		"init", "publish", "republish", "install", "uninstall",
		"search", "info", "versions", "update", "list",
	}

	for _, name := range subcommands {
		found := false
		for _, cmd := range pkgCmd.Commands() {
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

func TestPkgFlags(t *testing.T) {
	// Check persistent flags
	serverFlag := pkgCmd.PersistentFlags().Lookup("server")
	if serverFlag == nil {
		t.Error("Expected --server flag")
	}

	dryRunFlag := pkgCmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected --dry-run flag")
	}

	// Check install-specific flags
	globalFlag := pkgInstallCmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Error("Expected --global flag on install command")
	}

	forceFlag := pkgInstallCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag on install command")
	}
}

func TestPkgCommandDescriptions(t *testing.T) {
	// Verify commands have descriptions
	if pkgCmd.Short == "" {
		t.Error("pkgCmd should have a short description")
	}
	if pkgCmd.Long == "" {
		t.Error("pkgCmd should have a long description")
	}

	for _, cmd := range pkgCmd.Commands() {
		if cmd.Short == "" {
			t.Errorf("Command '%s' should have a short description", cmd.Name())
		}
	}
}
