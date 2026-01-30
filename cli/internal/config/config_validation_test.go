package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a temporary SSH key file
func createTempSSHKey(t *testing.T) string {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	if err := os.WriteFile(keyPath, []byte("fake ssh key"), 0600); err != nil {
		t.Fatalf("failed to create temp SSH key: %v", err)
	}
	return keyPath
}

func TestServerValidate(t *testing.T) {
	validKey := createTempSSHKey(t)

	tests := []struct {
		name        string
		server      Server
		wantErr     bool
		errContains string
	}{
		{
			name: "valid server",
			server: Server{
				Name:        "test-server",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
				Modules:     []string{"core", "pytorch"},
			},
			wantErr: false,
		},
		{
			name: "valid server with hostname",
			server: Server{
				Name:        "prod",
				Host:        "example.com",
				User:        "admin",
				SSHKey:      validKey,
				CostPerHour: 20.0,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			server: Server{
				Name:        "",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "name with spaces",
			server: Server{
				Name:        "test server",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "cannot contain spaces",
		},
		{
			name: "missing host",
			server: Server{
				Name:        "test",
				Host:        "",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "host is required",
		},
		{
			name: "invalid host",
			server: Server{
				Name:        "test",
				Host:        "not a valid host!!!",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "not a valid hostname or IP",
		},
		{
			name: "missing user",
			server: Server{
				Name:        "test",
				Host:        "192.168.1.1",
				User:        "",
				SSHKey:      validKey,
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "user is required",
		},
		{
			name: "non-existent SSH key",
			server: Server{
				Name:        "test",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      "/nonexistent/key/path",
				CostPerHour: 10.5,
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "negative cost",
			server: Server{
				Name:        "test",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: -5.0,
			},
			wantErr:     true,
			errContains: "cannot be negative",
		},
		{
			name: "invalid module ID",
			server: Server{
				Name:        "test",
				Host:        "192.168.1.1",
				User:        "ubuntu",
				SSHKey:      validKey,
				CostPerHour: 10.5,
				Modules:     []string{"core", "nonexistent-module"},
			},
			wantErr:     true,
			errContains: "invalid module ID",
		},
		{
			name: "multiple validation errors",
			server: Server{
				Name:        "",
				Host:        "",
				User:        "",
				SSHKey:      "/nonexistent/key",
				CostPerHour: -10.0,
			},
			wantErr:     true,
			errContains: "validation errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.server.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	validKey := createTempSSHKey(t)

	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: Config{
				Servers: []Server{
					{
						Name:        "server1",
						Host:        "192.168.1.1",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
					},
				},
				APIKeys: APIKeys{
					Anthropic: "sk-ant-" + strings.Repeat("a", 40),
				},
				Collections: []Collection{
					{
						Name: "images",
						Path: "/tmp/images",
						Type: "image",
					},
				},
				Users: []User{
					{
						Name: "alice",
						Path: "/home/alice",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate server names",
			config: Config{
				Servers: []Server{
					{
						Name:        "server1",
						Host:        "192.168.1.1",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
					},
					{
						Name:        "server1",
						Host:        "192.168.1.2",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
					},
				},
			},
			wantErr:     true,
			errContains: "duplicate server name",
		},
		{
			name: "invalid server in list",
			config: Config{
				Servers: []Server{
					{
						Name:        "",
						Host:        "192.168.1.1",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
					},
				},
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "invalid API key format - Anthropic",
			config: Config{
				APIKeys: APIKeys{
					Anthropic: "invalid-key",
				},
			},
			wantErr:     true,
			errContains: "should start with 'sk-ant-'",
		},
		{
			name: "invalid API key format - OpenAI",
			config: Config{
				APIKeys: APIKeys{
					OpenAI: "invalid-key",
				},
			},
			wantErr:     true,
			errContains: "should start with 'sk-'",
		},
		{
			name: "API key too short",
			config: Config{
				APIKeys: APIKeys{
					Anthropic: "sk-ant-short",
				},
			},
			wantErr:     true,
			errContains: "too short",
		},
		{
			name: "duplicate collection names",
			config: Config{
				Collections: []Collection{
					{Name: "images", Path: "/tmp/1", Type: "image"},
					{Name: "images", Path: "/tmp/2", Type: "video"},
				},
			},
			wantErr:     true,
			errContains: "duplicate collection name",
		},
		{
			name: "collection missing name",
			config: Config{
				Collections: []Collection{
					{Name: "", Path: "/tmp/images", Type: "image"},
				},
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "collection missing path",
			config: Config{
				Collections: []Collection{
					{Name: "images", Path: "", Type: "image"},
				},
			},
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name: "collection invalid type",
			config: Config{
				Collections: []Collection{
					{Name: "images", Path: "/tmp/images", Type: "invalid"},
				},
			},
			wantErr:     true,
			errContains: "invalid type",
		},
		{
			name: "duplicate user names",
			config: Config{
				Users: []User{
					{Name: "alice", Path: "/home/alice"},
					{Name: "alice", Path: "/home/alice2"},
				},
			},
			wantErr:     true,
			errContains: "duplicate user name",
		},
		{
			name: "user missing name",
			config: Config{
				Users: []User{
					{Name: "", Path: "/home/user"},
				},
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "user missing path",
			config: Config{
				Users: []User{
					{Name: "alice", Path: ""},
				},
			},
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name: "active user does not exist",
			config: Config{
				Users: []User{
					{Name: "alice", Path: "/home/alice"},
				},
				ActiveUser: "bob",
			},
			wantErr:     true,
			errContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []string
		expected string
	}{
		{
			name:     "no errors",
			errors:   []string{},
			expected: "",
		},
		{
			name:     "single error",
			errors:   []string{"error 1"},
			expected: "error 1",
		},
		{
			name:     "multiple errors",
			errors:   []string{"error 1", "error 2", "error 3"},
			expected: "3 validation errors:\n  - error 1\n  - error 2\n  - error 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{Errors: tt.errors}
			got := ve.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidationErrorAdd(t *testing.T) {
	ve := &ValidationError{}
	if ve.HasErrors() {
		t.Error("expected no errors initially")
	}

	ve.Add("error 1")
	if !ve.HasErrors() {
		t.Error("expected errors after Add()")
	}
	if len(ve.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(ve.Errors))
	}

	ve.Add("error 2")
	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(ve.Errors))
	}
}

func TestIsValidHostOrIP(t *testing.T) {
	tests := []struct {
		host  string
		valid bool
	}{
		// Valid IPs
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},

		// Valid hostnames
		{"example.com", true},
		{"sub.example.com", true},
		{"my-server", true},
		{"server123", true},
		{"a.b.c.d.e", true},

		// Invalid
		{"", false},
		{"not a host!!!", false},
		{"server with spaces", false},
		{"-invalid", false},
		{"invalid-", false},
		{"192.168.1.999", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isValidHostOrIP(tt.host)
			if got != tt.valid {
				t.Errorf("isValidHostOrIP(%q) = %v, want %v", tt.host, got, tt.valid)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~/.ssh/id_rsa", filepath.Join(home, ".ssh/id_rsa")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestCheckCircularDependency(t *testing.T) {
	// Create a module map with circular dependency
	moduleMap := map[string]*Module{
		"a": {ID: "a", Dependencies: []string{"b"}},
		"b": {ID: "b", Dependencies: []string{"c"}},
		"c": {ID: "c", Dependencies: []string{"a"}}, // Creates cycle: a->b->c->a
	}

	err := checkCircularDependency("a", moduleMap, []string{})
	if err == nil {
		t.Error("expected circular dependency error, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error = %v, want error containing 'circular dependency'", err)
	}

	// Test non-circular dependency
	moduleMapValid := map[string]*Module{
		"a": {ID: "a", Dependencies: []string{"b"}},
		"b": {ID: "b", Dependencies: []string{"c"}},
		"c": {ID: "c", Dependencies: []string{}},
	}

	err = checkCircularDependency("a", moduleMapValid, []string{})
	if err != nil {
		t.Errorf("unexpected error for valid dependencies: %v", err)
	}
}

func TestModuleDependencyValidation(t *testing.T) {
	validKey := createTempSSHKey(t)

	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid module dependencies",
			config: Config{
				Servers: []Server{
					{
						Name:        "server1",
						Host:        "192.168.1.1",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
						Modules:     []string{"pytorch"}, // pytorch depends on core
					},
				},
			},
			wantErr: false,
		},
		{
			name: "all real modules are valid",
			config: Config{
				Servers: []Server{
					{
						Name:        "server1",
						Host:        "192.168.1.1",
						User:        "ubuntu",
						SSHKey:      validKey,
						CostPerHour: 10.0,
						Modules:     []string{"core", "pytorch", "ollama", "sdxl"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateModuleDependencies()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
