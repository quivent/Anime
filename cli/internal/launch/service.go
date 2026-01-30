package launch

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/ssh"
)

// CommandRunner abstracts local vs remote command execution
type CommandRunner interface {
	Run(cmd string) (string, error)
	RunSudo(cmd, password string) (string, error)
	User() string
}

// LocalRunner executes commands on the local machine
type LocalRunner struct {
	user string
}

// NewLocalRunner creates a runner for local execution
func NewLocalRunner() *LocalRunner {
	user := "root"
	if out, err := exec.Command("whoami").Output(); err == nil {
		user = strings.TrimSpace(string(out))
	}
	return &LocalRunner{user: user}
}

func (r *LocalRunner) Run(cmd string) (string, error) {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *LocalRunner) RunSudo(cmd, password string) (string, error) {
	sudoCmd := fmt.Sprintf("echo '%s' | sudo -S bash -c '%s'", password, strings.ReplaceAll(cmd, "'", "'\\''"))
	return r.Run(sudoCmd)
}

func (r *LocalRunner) User() string {
	return r.user
}

// RemoteRunner executes commands via SSH
type RemoteRunner struct {
	client *ssh.Client
	user   string
}

// NewRemoteRunner creates a runner for remote execution
func NewRemoteRunner(client *ssh.Client, user string) *RemoteRunner {
	return &RemoteRunner{client: client, user: user}
}

func (r *RemoteRunner) Run(cmd string) (string, error) {
	return r.client.RunCommand(cmd)
}

func (r *RemoteRunner) RunSudo(cmd, password string) (string, error) {
	sudoCmd := fmt.Sprintf("echo '%s' | sudo -S bash -c '%s'", password, strings.ReplaceAll(cmd, "'", "'\\''"))
	return r.client.RunCommand(sudoCmd)
}

func (r *RemoteRunner) User() string {
	return r.user
}

// GetServiceStatus returns the systemd service status
func GetServiceStatus(serviceName string, runner CommandRunner) (string, error) {
	out, err := runner.Run(fmt.Sprintf("systemctl is-active %s 2>/dev/null || echo stopped", serviceName))
	if err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(out), nil
}

// StopService stops a systemd service
func StopService(serviceName, password string, runner CommandRunner) error {
	_, err := runner.RunSudo(fmt.Sprintf("systemctl stop %s", serviceName), password)
	return err
}

// GetServiceLogs returns recent journal logs for a service
func GetServiceLogs(serviceName string, lines int, runner CommandRunner) (string, error) {
	return runner.Run(fmt.Sprintf("journalctl -u %s -n %d --no-pager 2>/dev/null || echo 'No logs available'", serviceName, lines))
}
