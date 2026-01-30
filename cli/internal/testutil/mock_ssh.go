package testutil

import (
	"fmt"
	"strings"
	"sync"
)

// MockSSHClient is a mock implementation of SSH client for testing
type MockSSHClient struct {
	mu              sync.Mutex
	commands        []string
	outputs         map[string]string
	errors          map[string]error
	files           map[string]string
	uploadedFiles   map[string]string
	executables     map[string]bool
	closed          bool
	host            string
	progressOutputs map[string][]string
}

// NewMockSSHClient creates a new mock SSH client
func NewMockSSHClient(host string) *MockSSHClient {
	return &MockSSHClient{
		commands:        []string{},
		outputs:         make(map[string]string),
		errors:          make(map[string]error),
		files:           make(map[string]string),
		uploadedFiles:   make(map[string]string),
		executables:     make(map[string]bool),
		host:            host,
		progressOutputs: make(map[string][]string),
	}
}

// SetCommandOutput sets the output for a specific command
func (m *MockSSHClient) SetCommandOutput(cmd, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputs[cmd] = output
}

// SetCommandError sets an error for a specific command
func (m *MockSSHClient) SetCommandError(cmd string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[cmd] = err
}

// SetFileExists marks a file as existing
func (m *MockSSHClient) SetFileExists(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = "exists"
}

// SetProgressOutput sets progressive output lines for a command
func (m *MockSSHClient) SetProgressOutput(cmd string, lines []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progressOutputs[cmd] = lines
}

// RunCommand simulates running a command
func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, cmd)

	// Check for errors first
	if err, ok := m.errors[cmd]; ok {
		return "", err
	}

	// Return configured output
	if output, ok := m.outputs[cmd]; ok {
		return output, nil
	}

	// Default responses for common commands
	if strings.HasPrefix(cmd, "test -f") {
		path := strings.TrimPrefix(cmd, "test -f ")
		path = strings.TrimSuffix(path, " && echo exists")
		path = strings.TrimSpace(path)
		if _, exists := m.files[path]; exists {
			return "exists", nil
		}
		return "", fmt.Errorf("file not found")
	}

	if cmd == "echo 'Connection successful'" {
		return "Connection successful", nil
	}

	if cmd == "nvidia-smi --query-gpu=name --format=csv,noheader" {
		return "NVIDIA GH200 120GB", nil
	}

	if cmd == "nproc" {
		return "96", nil
	}

	if strings.HasPrefix(cmd, "chmod +x") {
		path := strings.TrimPrefix(cmd, "chmod +x ")
		path = strings.TrimSpace(path)
		m.executables[path] = true
		return "", nil
	}

	if strings.HasPrefix(cmd, "rm -f") {
		return "", nil
	}

	return "", nil
}

// RunCommandWithProgress simulates running a command with progress updates
func (m *MockSSHClient) RunCommandWithProgress(cmd string, progress chan<- string) error {
	m.mu.Lock()
	m.commands = append(m.commands, cmd)

	// Get progress outputs
	lines := m.progressOutputs[cmd]
	err := m.errors[cmd]
	m.mu.Unlock()

	// Send progress updates
	for _, line := range lines {
		progress <- line + "\n"
	}

	close(progress)
	return err
}

// UploadString simulates uploading a string as a file
func (m *MockSSHClient) UploadString(content, remotePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors["upload:"+remotePath]; ok {
		return err
	}

	m.uploadedFiles[remotePath] = content
	return nil
}

// UploadFile simulates uploading a file
func (m *MockSSHClient) UploadFile(localPath, remotePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors["upload:"+remotePath]; ok {
		return err
	}

	m.uploadedFiles[remotePath] = fmt.Sprintf("uploaded from %s", localPath)
	return nil
}

// FileExists simulates checking if a file exists
func (m *MockSSHClient) FileExists(path string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors["exists:"+path]; ok {
		return false, err
	}

	_, exists := m.files[path]
	return exists, nil
}

// MakeExecutable simulates making a file executable
func (m *MockSSHClient) MakeExecutable(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors["chmod:"+path]; ok {
		return err
	}

	m.executables[path] = true
	return nil
}

// Close simulates closing the connection
func (m *MockSSHClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// Host returns the mock hostname
func (m *MockSSHClient) Host() string {
	return m.host
}

// GetCommands returns all executed commands (for testing)
func (m *MockSSHClient) GetCommands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.commands...)
}

// GetUploadedFile returns the content of an uploaded file (for testing)
func (m *MockSSHClient) GetUploadedFile(path string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	content, ok := m.uploadedFiles[path]
	return content, ok
}

// IsExecutable checks if a file was marked as executable (for testing)
func (m *MockSSHClient) IsExecutable(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.executables[path]
}

// IsClosed checks if the client was closed (for testing)
func (m *MockSSHClient) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}
