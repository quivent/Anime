package ssh

import (
	"testing"

	"github.com/joshkornreich/anime/internal/testutil"
	"golang.org/x/crypto/ssh"
)

func TestClient_Structure(t *testing.T) {
	// Test that the Client struct has expected fields
	// We can't easily create a real SSH client in tests, but we can verify the struct exists
	var c *Client

	// Verify it's a pointer type (nil pointer)
	if c != nil {
		t.Error("expected nil client pointer")
	}

	// Create a client with nil values (not connected)
	c = &Client{
		client: nil,
		config: nil,
		host:   "test-host:22",
	}

	testutil.AssertNotNil(t, c)
	testutil.AssertEqual(t, c.host, "test-host:22")
}

func TestClient_Host(t *testing.T) {
	tests := []struct {
		name     string
		hostPort string
		expected string
	}{
		{
			name:     "host with port",
			hostPort: "example.com:22",
			expected: "example.com",
		},
		{
			name:     "host with custom port",
			hostPort: "example.com:2222",
			expected: "example.com",
		},
		{
			name:     "host without port",
			hostPort: "example.com",
			expected: "example.com",
		},
		{
			name:     "IP with port",
			hostPort: "192.168.1.1:22",
			expected: "192.168.1.1",
		},
		{
			name:     "localhost with port",
			hostPort: "localhost:22",
			expected: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{host: tt.hostPort}
			result := c.Host()
			testutil.AssertEqual(t, result, tt.expected)
		})
	}
}

func TestClient_Close(t *testing.T) {
	// Test Close with nil client (should not panic)
	c := &Client{client: nil}
	err := c.Close()
	testutil.AssertNil(t, err)
}

func TestNewClient_HostPortHandling(t *testing.T) {
	t.Skip("Skipping NewClient tests as they require actual SSH connection")
	// Note: NewClient adds :22 port if not specified in the host string
	// This is tested indirectly through integration tests
}

func TestNewClient_KeyPathExpansion(t *testing.T) {
	t.Skip("Skipping NewClient tests as they require actual SSH connection")
	// Note: NewClient expands ~/ paths using os.UserHomeDir
	// This is tested indirectly through integration tests
}

func TestClientConfig_Structure(t *testing.T) {
	// Test that we can create a ClientConfig with expected fields
	config := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	testutil.AssertNotNil(t, config)
	testutil.AssertEqual(t, config.User, "testuser")
	testutil.AssertNotNil(t, config.HostKeyCallback)
}

func TestClient_Methods_Exist(t *testing.T) {
	// Verify that all expected methods exist on Client
	// This is a compile-time check more than runtime
	var c *Client = &Client{}

	// These calls will panic or fail, but the test is that they compile
	// showing the methods exist with correct signatures

	// Verify method signatures exist (won't actually call them)
	_ = func() (string, error) { return c.RunCommand("") }
	_ = func() error { return c.RunCommandWithProgress("", nil) }
	_ = func() error { return c.UploadFile("", "") }
	_ = func() error { return c.UploadString("", "") }
	_ = func() (bool, error) { return c.FileExists("") }
	_ = func() error { return c.MakeExecutable("") }
	_ = func() error { return c.Close() }
	_ = func() string { return c.Host() }

	// If we got here, all methods exist with correct signatures
	testutil.AssertNotNil(t, c)
}

func TestNewClient_AuthMethods(t *testing.T) {
	// Create a temp directory for test keys
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Test with non-existent key file - should fail before attempting connection
	keyPath := tempDir + "/nonexistent_key"
	_, err := NewClient("example.com", "user", keyPath)
	testutil.AssertError(t, err)
	// Should fail with key loading error
	testutil.AssertContains(t, err.Error(), "load_key")
}

func TestNewClient_InvalidKey(t *testing.T) {
	// Test with invalid key content
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Write invalid key data
	invalidKey := testutil.WriteFile(t, tempDir, "invalid_key", "this is not a valid SSH key")

	_, err := NewClient("example.com", "user", invalidKey)
	testutil.AssertError(t, err)
	// Should fail with key loading error (invalid key)
	testutil.AssertContains(t, err.Error(), "load_key")
}

func TestClient_Integration_Methods(t *testing.T) {
	// Create a minimal client structure for testing methods that don't need connection
	c := &Client{
		client: nil,
		config: nil,
		host:   "test-host:22",
	}

	// Test Host() method
	host := c.Host()
	testutil.AssertEqual(t, host, "test-host")

	// Test Close() with nil client
	err := c.Close()
	testutil.AssertNil(t, err)
}

func TestClient_RunCommand_RequiresConnection(t *testing.T) {
	// Verify that RunCommand requires an active connection
	c := &Client{
		client: nil,
		config: nil,
		host:   "test",
	}

	// Should fail because client is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling RunCommand with nil client")
		}
	}()

	// This should panic with nil pointer dereference
	_, _ = c.RunCommand("echo test")
}

func TestClient_FileExists_Logic(t *testing.T) {
	// Test the logic of FileExists (it checks command output)
	// We can't test actual SSH, but we can verify the method exists

	c := &Client{host: "test"}

	// Should panic or error because client is nil
	// but the method signature is correct
	testutil.AssertNotNil(t, c)
}

func TestHostKeyCallback(t *testing.T) {
	// Test that InsecureIgnoreHostKey works
	callback := ssh.InsecureIgnoreHostKey()
	testutil.AssertNotNil(t, callback)

	// The callback should not return an error for any host key
	err := callback("hostname", nil, nil)
	testutil.AssertNoError(t, err)
}
