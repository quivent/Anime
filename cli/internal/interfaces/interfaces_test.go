package interfaces_test

import (
	"errors"
	"testing"

	"github.com/joshkornreich/anime/internal/interfaces"
)

// Mock implementations for testing

type MockSSHClient struct {
	CommandOutput string
	CommandError  error
	CloseCalled   bool
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
	return m.CommandOutput, m.CommandError
}

func (m *MockSSHClient) RunCommandWithProgress(cmd string, progress chan<- string) error {
	if m.CommandError != nil {
		return m.CommandError
	}
	progress <- m.CommandOutput
	return nil
}

func (m *MockSSHClient) UploadString(content, path string) error {
	return m.CommandError
}

func (m *MockSSHClient) MakeExecutable(path string) error {
	return m.CommandError
}

func (m *MockSSHClient) Close() error {
	m.CloseCalled = true
	return m.CommandError
}

type MockInstaller struct {
	InstallError    error
	ProgressChannel chan interfaces.ProgressUpdate
	ParallelEnabled bool
	JobCount        int
}

func (m *MockInstaller) Install(modules []string) error {
	if m.ProgressChannel != nil {
		for _, mod := range modules {
			m.ProgressChannel <- interfaces.ProgressUpdate{
				Module: mod,
				Status: "Complete",
				Done:   true,
			}
		}
		close(m.ProgressChannel)
	}
	return m.InstallError
}

func (m *MockInstaller) GetProgressChannel() <-chan interfaces.ProgressUpdate {
	if m.ProgressChannel == nil {
		m.ProgressChannel = make(chan interfaces.ProgressUpdate, 10)
	}
	return m.ProgressChannel
}

func (m *MockInstaller) SetParallel(parallel bool) {
	m.ParallelEnabled = parallel
}

func (m *MockInstaller) SetJobs(jobs int) {
	m.JobCount = jobs
}

func (m *MockInstaller) TestConnection() error {
	return m.InstallError
}

type MockSourceController struct {
	PushError   error
	PullError   error
	StatusError error
	LinkPath    string
}

func (m *MockSourceController) Push() error {
	return m.PushError
}

func (m *MockSourceController) Pull() error {
	return m.PullError
}

func (m *MockSourceController) Status() error {
	return m.StatusError
}

func (m *MockSourceController) Link(remotePath string) error {
	m.LinkPath = remotePath
	return nil
}

func (m *MockSourceController) Unlink() error {
	m.LinkPath = ""
	return nil
}

type MockPackageManager struct {
	InstallError   error
	UninstallError error
	PublishError   error
	SearchError    error
	ListError      error
}

func (m *MockPackageManager) Install(name string, global bool, force bool) error {
	return m.InstallError
}

func (m *MockPackageManager) Uninstall(name string, global bool) error {
	return m.UninstallError
}

func (m *MockPackageManager) Publish() error {
	return m.PublishError
}

func (m *MockPackageManager) Search(query string) error {
	return m.SearchError
}

func (m *MockPackageManager) List(global bool) error {
	return m.ListError
}

// Tests demonstrating interface usage

func TestSSHClientInterface(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		mock := &MockSSHClient{
			CommandOutput: "test output",
		}

		output, err := mock.RunCommand("echo test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if output != "test output" {
			t.Errorf("expected 'test output', got '%s'", output)
		}
	})

	t.Run("command execution with error", func(t *testing.T) {
		mock := &MockSSHClient{
			CommandError: errors.New("connection failed"),
		}

		_, err := mock.RunCommand("echo test")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("close connection", func(t *testing.T) {
		mock := &MockSSHClient{}

		err := mock.Close()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !mock.CloseCalled {
			t.Error("expected Close to be called")
		}
	})
}

func TestInstallerInterface(t *testing.T) {
	t.Run("install modules", func(t *testing.T) {
		progressChan := make(chan interfaces.ProgressUpdate, 10)
		mock := &MockInstaller{
			ProgressChannel: progressChan,
		}

		modules := []string{"module1", "module2"}
		err := mock.Install(modules)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify progress updates
		count := 0
		for update := range mock.GetProgressChannel() {
			if !update.Done {
				t.Error("expected Done to be true")
			}
			count++
		}

		if count != len(modules) {
			t.Errorf("expected %d progress updates, got %d", len(modules), count)
		}
	})

	t.Run("set parallel mode", func(t *testing.T) {
		mock := &MockInstaller{}

		mock.SetParallel(true)
		if !mock.ParallelEnabled {
			t.Error("expected ParallelEnabled to be true")
		}

		mock.SetJobs(4)
		if mock.JobCount != 4 {
			t.Errorf("expected JobCount to be 4, got %d", mock.JobCount)
		}
	})
}

func TestSourceControllerInterface(t *testing.T) {
	t.Run("link and unlink", func(t *testing.T) {
		mock := &MockSourceController{}

		err := mock.Link("remote/path")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if mock.LinkPath != "remote/path" {
			t.Errorf("expected LinkPath to be 'remote/path', got '%s'", mock.LinkPath)
		}

		err = mock.Unlink()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if mock.LinkPath != "" {
			t.Errorf("expected LinkPath to be empty, got '%s'", mock.LinkPath)
		}
	})

	t.Run("push and pull", func(t *testing.T) {
		mock := &MockSourceController{}

		if err := mock.Push(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if err := mock.Pull(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestPackageManagerInterface(t *testing.T) {
	t.Run("install package", func(t *testing.T) {
		mock := &MockPackageManager{}

		err := mock.Install("package@1.0.0", false, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("publish package", func(t *testing.T) {
		mock := &MockPackageManager{}

		err := mock.Publish()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("search packages", func(t *testing.T) {
		mock := &MockPackageManager{}

		err := mock.Search("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Example of using interfaces for dependency injection

type DeploymentService struct {
	ssh       interfaces.SSHClient
	installer interfaces.Installer
}

func NewDeploymentService(ssh interfaces.SSHClient, inst interfaces.Installer) *DeploymentService {
	return &DeploymentService{
		ssh:       ssh,
		installer: inst,
	}
}

func (d *DeploymentService) Deploy(modules []string) error {
	// Test connection
	if _, err := d.ssh.RunCommand("echo test"); err != nil {
		return err
	}

	// Install modules
	return d.installer.Install(modules)
}

func TestDeploymentService(t *testing.T) {
	mockSSH := &MockSSHClient{
		CommandOutput: "test",
	}
	progressChan := make(chan interfaces.ProgressUpdate, 10)
	mockInstaller := &MockInstaller{
		ProgressChannel: progressChan,
	}

	service := NewDeploymentService(mockSSH, mockInstaller)

	err := service.Deploy([]string{"module1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
