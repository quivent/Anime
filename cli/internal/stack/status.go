package stack

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/launch"
)

// ServiceStatus represents the status of a single service
type ServiceStatus struct {
	Name   string
	Status string // "running", "stopped", "failed", "unknown"
	Port   int
	Domain string
}

// GetStackStatus returns the status of all services in a stack
func GetStackStatus(stackName string, runner launch.CommandRunner) (map[string]string, error) {
	statuses := make(map[string]string)

	// List all systemd services matching the stack name pattern
	cmd := fmt.Sprintf("systemctl list-units --type=service --no-legend 'anime-%s-*' 2>/dev/null | awk '{print $1, $4}'", stackName)
	out, err := runner.Run(cmd)
	if err != nil {
		// If no services found, return empty map
		return statuses, nil
	}

	// Parse output: each line is "service-name status"
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Extract service name from full systemd unit name
			// e.g., "anime-mystack-api.service running" -> "api"
			unitName := parts[0]
			status := parts[1]

			// Remove "anime-{stackName}-" prefix and ".service" suffix
			prefix := "anime-" + stackName + "-"
			serviceName := strings.TrimPrefix(unitName, prefix)
			serviceName = strings.TrimSuffix(serviceName, ".service")

			statuses[serviceName] = status
		}
	}

	return statuses, nil
}

// GetServiceStatus returns the status of a specific service in a stack
func GetServiceStatus(stackName, serviceName string, runner launch.CommandRunner) (string, error) {
	fullServiceName := launch.ServiceName(stackName + "-" + serviceName)
	return launch.GetServiceStatus(fullServiceName, runner)
}

// GetDetailedStackStatus returns detailed status information for all services
func GetDetailedStackStatus(cfg *StackConfig, runner launch.CommandRunner) ([]ServiceStatus, error) {
	var statuses []ServiceStatus

	for name, svc := range cfg.Services {
		status := ServiceStatus{
			Name:   name,
			Port:   svc.Port,
			Domain: svc.Domain,
		}

		// Handle database services differently
		if svc.Type == "postgres" {
			// Check if postgres is running
			out, err := runner.Run("systemctl is-active postgresql 2>/dev/null || echo stopped")
			if err == nil {
				status.Status = strings.TrimSpace(out)
			} else {
				status.Status = "unknown"
			}
		} else {
			// Get systemd service status
			fullServiceName := launch.ServiceName(cfg.Name + "-" + name)
			svcStatus, err := launch.GetServiceStatus(fullServiceName, runner)
			if err != nil {
				status.Status = "unknown"
			} else {
				status.Status = svcStatus
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// StopStack stops all services in a stack
func StopStack(stackName string, sudoPassword string, runner launch.CommandRunner) error {
	// Get all running services for this stack
	statuses, err := GetStackStatus(stackName, runner)
	if err != nil {
		return err
	}

	var errors []string
	for serviceName := range statuses {
		fullServiceName := launch.ServiceName(stackName + "-" + serviceName)
		if err := launch.StopService(fullServiceName, sudoPassword, runner); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some services: %s", strings.Join(errors, "; "))
	}

	return nil
}

// RestartStack restarts all services in a stack
func RestartStack(stackName string, sudoPassword string, runner launch.CommandRunner) error {
	// Get all services for this stack
	statuses, err := GetStackStatus(stackName, runner)
	if err != nil {
		return err
	}

	var errors []string
	for serviceName := range statuses {
		fullServiceName := launch.ServiceName(stackName + "-" + serviceName)
		cmd := fmt.Sprintf("systemctl restart %s", fullServiceName)
		if _, err := runner.RunSudo(cmd, sudoPassword); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to restart some services: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetStackLogs returns recent logs for all services in a stack
func GetStackLogs(stackName string, lines int, runner launch.CommandRunner) (map[string]string, error) {
	logs := make(map[string]string)

	// Get all services for this stack
	statuses, err := GetStackStatus(stackName, runner)
	if err != nil {
		return nil, err
	}

	for serviceName := range statuses {
		fullServiceName := launch.ServiceName(stackName + "-" + serviceName)
		svcLogs, err := launch.GetServiceLogs(fullServiceName, lines, runner)
		if err == nil {
			logs[serviceName] = svcLogs
		}
	}

	return logs, nil
}
