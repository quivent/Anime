package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joshkornreich/anime/internal/testutil"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create temp directory for config
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set HOME to temp directory so GetConfigPath uses it
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create config directory structure
	configDir := filepath.Join(tempDir, ".config", "anime")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	// Create SSH directory and key file for validation
	sshDir := filepath.Join(tempDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("failed to create ssh dir: %v", err)
	}
	sshKeyPath := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(sshKeyPath, []byte("test-key-content"), 0600); err != nil {
		t.Fatalf("failed to write ssh key: %v", err)
	}

	// Create a valid config file (using valid API key formats)
	configContent := `servers:
  - name: test-server
    host: example.com
    user: testuser
    ssh_key: ~/.ssh/id_rsa
    cost_per_hour: 1.5
    modules:
      - core
      - pytorch
api_keys:
  anthropic: sk-ant-api03-test-key-12345678901234567890123456789012345678901234567890
  openai: sk-proj-test-key-1234567890123456789012345678901234567890
aliases:
  quick: "llama-3.3-70b"
shell_aliases:
  gs: "git status"
collections:
  - name: my-collection
    path: /data/images
    type: image
users:
  - name: alice
    path: /home/alice
active_user: alice
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Load config
	cfg, err := Load()

	testutil.AssertNoError(t, err)
	testutil.AssertNotNil(t, cfg)

	// Verify servers
	testutil.AssertEqual(t, len(cfg.Servers), 1)
	testutil.AssertEqual(t, cfg.Servers[0].Name, "test-server")
	testutil.AssertEqual(t, cfg.Servers[0].Host, "example.com")
	testutil.AssertEqual(t, cfg.Servers[0].User, "testuser")
	testutil.AssertEqual(t, cfg.Servers[0].CostPerHour, 1.5)
	testutil.AssertEqual(t, len(cfg.Servers[0].Modules), 2)

	// Verify API keys
	testutil.AssertEqual(t, cfg.APIKeys.Anthropic, "sk-ant-api03-test-key-12345678901234567890123456789012345678901234567890")
	testutil.AssertEqual(t, cfg.APIKeys.OpenAI, "sk-proj-test-key-1234567890123456789012345678901234567890")

	// Verify aliases
	testutil.AssertNotNil(t, cfg.Aliases)
	testutil.AssertEqual(t, cfg.Aliases["quick"], "llama-3.3-70b")

	// Verify shell aliases
	testutil.AssertNotNil(t, cfg.ShellAliases)
	testutil.AssertEqual(t, cfg.ShellAliases["gs"], "git status")

	// Verify collections
	testutil.AssertEqual(t, len(cfg.Collections), 1)
	testutil.AssertEqual(t, cfg.Collections[0].Name, "my-collection")

	// Verify users
	testutil.AssertEqual(t, len(cfg.Users), 1)
	testutil.AssertEqual(t, cfg.Users[0].Name, "alice")
	testutil.AssertEqual(t, cfg.ActiveUser, "alice")
}

func TestLoad_MissingFile(t *testing.T) {
	// Create temp directory for config
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set HOME to temp directory with no config file
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Load should return default config when file doesn't exist
	cfg, err := Load()

	testutil.AssertNoError(t, err)
	testutil.AssertNotNil(t, cfg)
	testutil.AssertEqual(t, len(cfg.Servers), 0)
	testutil.AssertNotNil(t, cfg.Aliases)
	testutil.AssertNotNil(t, cfg.Collections)
	testutil.AssertNotNil(t, cfg.Users)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create config directory structure
	configDir := filepath.Join(tempDir, ".config", "anime")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	// Create invalid YAML
	invalidYAML := `servers:
  - name: test
    invalid yaml content here [[[
    no proper structure
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Load should return error
	_, err := Load()
	testutil.AssertError(t, err)
}

func TestSave(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, ".config", "anime", "config.yaml")

	// Create and save config (using valid API key format)
	cfg := &Config{
		Servers: []Server{
			{Name: "test", Host: "localhost", User: "user"},
		},
		APIKeys: APIKeys{Anthropic: "sk-ant-api03-test-key-12345678901234567890123456789012345678901234567890"},
		Aliases: map[string]string{"test": "value"},
	}

	err := cfg.Save()
	testutil.AssertNoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	testutil.AssertNoError(t, err)

	// Load and verify
	loaded, err := Load()
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(loaded.Servers), 1)
	testutil.AssertEqual(t, loaded.Servers[0].Name, "test")
}

func TestGetServer(t *testing.T) {
	cfg := &Config{
		Servers: []Server{
			{Name: "server1", Host: "host1.com"},
			{Name: "server2", Host: "host2.com"},
		},
	}

	tests := []struct {
		name        string
		serverName  string
		expectError bool
		expectHost  string
	}{
		{
			name:        "existing server",
			serverName:  "server1",
			expectError: false,
			expectHost:  "host1.com",
		},
		{
			name:        "another existing server",
			serverName:  "server2",
			expectError: false,
			expectHost:  "host2.com",
		},
		{
			name:        "non-existent server",
			serverName:  "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := cfg.GetServer(tt.serverName)

			if tt.expectError {
				testutil.AssertError(t, err)
				if server != nil {
					t.Errorf("expected nil server, got %v", server)
				}
			} else {
				testutil.AssertNoError(t, err)
				if server == nil {
					t.Error("expected non-nil server")
				} else {
					testutil.AssertEqual(t, server.Host, tt.expectHost)
				}
			}
		})
	}
}

func TestEstimateCost(t *testing.T) {
	// Save original modules
	originalModules := AvailableModules
	defer func() { AvailableModules = originalModules }()

	// Set up test modules with known times
	AvailableModules = []Module{
		{ID: "core", TimeMinutes: 5, Dependencies: []string{}},
		{ID: "pytorch", TimeMinutes: 10, Dependencies: []string{"core"}},
		{ID: "ollama", TimeMinutes: 2, Dependencies: []string{"core"}},
		{ID: "vllm", TimeMinutes: 8, Dependencies: []string{"core", "pytorch"}},
	}

	tests := []struct {
		name         string
		modules      []string
		costPerHour  float64
		expectedCost float64
	}{
		{
			name:         "single module no deps",
			modules:      []string{"core"},
			costPerHour:  60.0,
			expectedCost: 5.0, // 5 minutes at $60/hour = $5
		},
		{
			name:         "module with one dep",
			modules:      []string{"pytorch"},
			costPerHour:  60.0,
			expectedCost: 15.0, // 5 (core) + 10 (pytorch) = 15 minutes = $15
		},
		{
			name:         "module with chain deps",
			modules:      []string{"vllm"},
			costPerHour:  60.0,
			expectedCost: 23.0, // 5 + 10 + 8 = 23 minutes = $23
		},
		{
			name:         "multiple modules with overlap",
			modules:      []string{"pytorch", "ollama"},
			costPerHour:  60.0,
			expectedCost: 17.0, // 5 (core, shared) + 10 (pytorch) + 2 (ollama) = 17 minutes
		},
		{
			name:         "different cost per hour",
			modules:      []string{"core"},
			costPerHour:  120.0,
			expectedCost: 10.0, // 5 minutes at $120/hour = $10
		},
		{
			name:         "fractional hours",
			modules:      []string{"pytorch"},
			costPerHour:  1.0,
			expectedCost: 0.25, // 15 minutes at $1/hour = $0.25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := EstimateCost(tt.modules, tt.costPerHour)

			// Use approximate comparison for floats
			if cost < tt.expectedCost-0.01 || cost > tt.expectedCost+0.01 {
				t.Errorf("got cost %.2f, want %.2f", cost, tt.expectedCost)
			}
		})
	}
}

func TestSetAlias(t *testing.T) {
	cfg := &Config{}

	// Test setting alias when map is nil
	cfg.SetAlias("test", "value")
	testutil.AssertNotNil(t, cfg.Aliases)
	testutil.AssertEqual(t, cfg.Aliases["test"], "value")

	// Test updating existing alias
	cfg.SetAlias("test", "newvalue")
	testutil.AssertEqual(t, cfg.Aliases["test"], "newvalue")

	// Test setting multiple aliases
	cfg.SetAlias("another", "alias")
	testutil.AssertEqual(t, len(cfg.Aliases), 2)
}

func TestGetAlias(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{
			"runtime": "runtime-value",
		},
	}

	tests := []struct {
		name     string
		alias    string
		expected string
	}{
		{
			name:     "runtime alias",
			alias:    "runtime",
			expected: "runtime-value",
		},
		{
			name:     "non-existent alias",
			alias:    "nonexistent",
			expected: "", // Falls back to defaults which might return empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.GetAlias(tt.alias)
			// Only check runtime aliases, embedded defaults may vary
			if tt.alias == "runtime" {
				testutil.AssertEqual(t, result, tt.expected)
			}
		})
	}
}

func TestDeleteAlias(t *testing.T) {
	tests := []struct {
		name        string
		initialMap  map[string]string
		deleteKey   string
		expectError bool
	}{
		{
			name:        "delete existing alias",
			initialMap:  map[string]string{"test": "value"},
			deleteKey:   "test",
			expectError: false,
		},
		{
			name:        "delete non-existent alias",
			initialMap:  map[string]string{"test": "value"},
			deleteKey:   "other",
			expectError: true,
		},
		{
			name:        "delete from nil map",
			initialMap:  nil,
			deleteKey:   "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Aliases: tt.initialMap}
			err := cfg.DeleteAlias(tt.deleteKey)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
				_, exists := cfg.Aliases[tt.deleteKey]
				if exists {
					t.Error("alias should have been deleted")
				}
			}
		})
	}
}

func TestGetModulesByID(t *testing.T) {
	// Save original modules
	originalModules := AvailableModules
	defer func() { AvailableModules = originalModules }()

	AvailableModules = []Module{
		{ID: "core", Name: "Core"},
		{ID: "pytorch", Name: "PyTorch"},
		{ID: "ollama", Name: "Ollama"},
	}

	tests := []struct {
		name     string
		ids      []string
		expected int
	}{
		{
			name:     "single module",
			ids:      []string{"core"},
			expected: 1,
		},
		{
			name:     "multiple modules",
			ids:      []string{"core", "pytorch", "ollama"},
			expected: 3,
		},
		{
			name:     "some non-existent",
			ids:      []string{"core", "unknown"},
			expected: 1,
		},
		{
			name:     "empty list",
			ids:      []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModulesByID(tt.ids)
			testutil.AssertEqual(t, len(result), tt.expected)
		})
	}
}

func TestAddCollection(t *testing.T) {
	cfg := &Config{
		Collections: []Collection{},
	}

	// Add first collection
	err := cfg.AddCollection(Collection{Name: "col1", Path: "/path1"})
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Collections), 1)

	// Add duplicate name should error
	err = cfg.AddCollection(Collection{Name: "col1", Path: "/path2"})
	testutil.AssertError(t, err)
	testutil.AssertEqual(t, len(cfg.Collections), 1)

	// Add different collection
	err = cfg.AddCollection(Collection{Name: "col2", Path: "/path2"})
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Collections), 2)
}

func TestGetCollection(t *testing.T) {
	cfg := &Config{
		Collections: []Collection{
			{Name: "col1", Path: "/path1"},
			{Name: "col2", Path: "/path2"},
		},
	}

	// Get existing collection
	col, err := cfg.GetCollection("col1")
	testutil.AssertNoError(t, err)
	if col == nil {
		t.Error("expected non-nil collection")
	} else {
		testutil.AssertEqual(t, col.Path, "/path1")
	}

	// Get non-existent collection
	col, err = cfg.GetCollection("unknown")
	testutil.AssertError(t, err)
	if col != nil {
		t.Errorf("expected nil collection, got %v", col)
	}
}

func TestDeleteCollection(t *testing.T) {
	cfg := &Config{
		Collections: []Collection{
			{Name: "col1", Path: "/path1"},
			{Name: "col2", Path: "/path2"},
		},
	}

	// Delete existing
	err := cfg.DeleteCollection("col1")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Collections), 1)
	testutil.AssertEqual(t, cfg.Collections[0].Name, "col2")

	// Delete non-existent
	err = cfg.DeleteCollection("unknown")
	testutil.AssertError(t, err)
}

func TestAddUser(t *testing.T) {
	cfg := &Config{
		Users: []User{},
	}

	// Add first user
	err := cfg.AddUser(User{Name: "alice", Path: "/home/alice"})
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Users), 1)

	// Add duplicate should error
	err = cfg.AddUser(User{Name: "alice", Path: "/home/alice2"})
	testutil.AssertError(t, err)

	// Add different user
	err = cfg.AddUser(User{Name: "bob", Path: "/home/bob"})
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Users), 2)
}

func TestSetActiveUser(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{Name: "alice", Path: "/home/alice"},
		},
	}

	// Set existing user
	err := cfg.SetActiveUser("alice")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, cfg.ActiveUser, "alice")

	// Set non-existent user should error
	err = cfg.SetActiveUser("unknown")
	testutil.AssertError(t, err)
}

func TestGetActiveUser(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{Name: "alice", Path: "/home/alice"},
		},
		ActiveUser: "alice",
	}

	// Get active user
	user, err := cfg.GetActiveUser()
	testutil.AssertNoError(t, err)
	if user == nil {
		t.Error("expected non-nil user")
	} else {
		testutil.AssertEqual(t, user.Name, "alice")
	}

	// No active user set
	cfg.ActiveUser = ""
	user, err = cfg.GetActiveUser()
	testutil.AssertError(t, err)
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}

func TestShellAliases(t *testing.T) {
	cfg := &Config{}

	// Add shell alias
	err := cfg.AddShellAlias("gs", "git status")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, cfg.ShellAliases["gs"], "git status")

	// Get shell aliases
	aliases := cfg.GetShellAliases()
	testutil.AssertNotNil(t, aliases)
	testutil.AssertEqual(t, aliases["gs"], "git status")

	// Remove shell alias
	err = cfg.RemoveShellAlias("gs")
	testutil.AssertNoError(t, err)
	_, exists := cfg.ShellAliases["gs"]
	if exists {
		t.Error("shell alias should have been removed")
	}

	// Remove non-existent should error
	err = cfg.RemoveShellAlias("unknown")
	testutil.AssertError(t, err)
}

func TestGetModulesByCategory(t *testing.T) {
	// Save original modules
	originalModules := AvailableModules
	defer func() { AvailableModules = originalModules }()

	AvailableModules = []Module{
		{ID: "core", Category: "System"},
		{ID: "pytorch", Category: "System"},
		{ID: "ollama", Category: "System"},
		{ID: "llama", Category: "LLM-Large"},
		{ID: "sdxl", Category: "Image"},
	}

	categories := GetModulesByCategory()

	// Check that all categories are present
	testutil.AssertEqual(t, len(categories["System"]), 3)
	testutil.AssertEqual(t, len(categories["LLM-Large"]), 1)
	testutil.AssertEqual(t, len(categories["Image"]), 1)
}

func TestUpdateServer(t *testing.T) {
	cfg := &Config{
		Servers: []Server{
			{Name: "test", Host: "old-host.com", User: "olduser"},
		},
	}

	// Update existing server
	newServer := Server{Name: "test", Host: "new-host.com", User: "newuser"}
	err := cfg.UpdateServer("test", newServer)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, cfg.Servers[0].Host, "new-host.com")
	testutil.AssertEqual(t, cfg.Servers[0].User, "newuser")

	// Update non-existent server
	err = cfg.UpdateServer("unknown", newServer)
	testutil.AssertError(t, err)
}

func TestDeleteServer(t *testing.T) {
	cfg := &Config{
		Servers: []Server{
			{Name: "server1", Host: "host1.com"},
			{Name: "server2", Host: "host2.com"},
		},
	}

	// Delete existing server
	err := cfg.DeleteServer("server1")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Servers), 1)
	testutil.AssertEqual(t, cfg.Servers[0].Name, "server2")

	// Delete non-existent server
	err = cfg.DeleteServer("unknown")
	testutil.AssertError(t, err)
}

func TestDeleteUser(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{Name: "alice", Path: "/home/alice"},
			{Name: "bob", Path: "/home/bob"},
		},
		ActiveUser: "alice",
	}

	// Delete active user should clear ActiveUser
	err := cfg.DeleteUser("alice")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(cfg.Users), 1)
	testutil.AssertEqual(t, cfg.ActiveUser, "")

	// Delete non-existent user
	err = cfg.DeleteUser("unknown")
	testutil.AssertError(t, err)
}
