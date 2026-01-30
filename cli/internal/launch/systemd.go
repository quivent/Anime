package launch

import (
	"bytes"
	"fmt"
	"text/template"
)

// SystemdConfig holds parameters for systemd unit generation
type SystemdConfig struct {
	Name        string
	Description string
	ExecStart   string
	WorkingDir  string
	User        string
	Port        int
	Environment map[string]string
}

const systemdTemplate = `[Unit]
Description={{.Description}}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{.ExecStart}}
WorkingDirectory={{.WorkingDir}}
User={{.User}}
Restart=always
RestartSec=3
Environment="PORT={{.Port}}"
{{- range $k, $v := .Environment}}
Environment="{{$k}}={{$v}}"
{{- end}}
StandardOutput=journal
StandardError=journal
SyslogIdentifier={{.Name}}

[Install]
WantedBy=multi-user.target
`

// GenerateSystemdUnit returns the .service file content
func GenerateSystemdUnit(cfg SystemdConfig) (string, error) {
	tmpl, err := template.New("systemd").Parse(systemdTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse systemd template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute systemd template: %w", err)
	}

	return buf.String(), nil
}

// ServiceName returns the standardized systemd service name for an app
func ServiceName(appName string) string {
	return "anime-" + appName
}

// InstallSystemdUnit writes, reloads, enables and starts the service
func InstallSystemdUnit(serviceName, content, password string, runner CommandRunner) error {
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	// Write unit file
	writeCmd := fmt.Sprintf("cat > %s << 'SYSTEMDEOF'\n%s\nSYSTEMDEOF", unitPath, content)
	if _, err := runner.RunSudo(writeCmd, password); err != nil {
		return fmt.Errorf("failed to write systemd unit: %w", err)
	}

	// Reload daemon
	if _, err := runner.RunSudo("systemctl daemon-reload", password); err != nil {
		return fmt.Errorf("daemon-reload failed: %w", err)
	}

	// Enable and start
	if _, err := runner.RunSudo(fmt.Sprintf("systemctl enable --now %s", serviceName), password); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}
