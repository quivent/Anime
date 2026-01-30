package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
)

type InstallModel struct {
	server      *config.Server
	modules     []string
	client      *ssh.Client
	installer   *installer.Installer
	progress    map[string]string
	currentStep string
	output      []string
	err         error
	done        bool
	startTime   time.Time
	systemInfo  map[string]string
}

type progressMsg installer.ProgressUpdate
type systemInfoMsg map[string]string
type errorMsg error

func NewInstallModel(server *config.Server) InstallModel {
	return InstallModel{
		server:    server,
		modules:   server.Modules,
		progress:  make(map[string]string),
		output:    []string{},
		startTime: time.Now(),
	}
}

func (m InstallModel) Init() tea.Cmd {
	return tea.Batch(
		m.connect,
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return t
		}),
	)
}

func (m InstallModel) connect() tea.Msg {
	client, err := ssh.NewClient(m.server.Host, m.server.User, m.server.SSHKey)
	if err != nil {
		return errorMsg(err)
	}

	inst := installer.New(client)

	// Test connection
	if err := inst.TestConnection(); err != nil {
		return errorMsg(fmt.Errorf("connection test failed: %w", err))
	}

	// Get system info
	info, err := inst.GetSystemInfo()
	if err != nil {
		return errorMsg(err)
	}

	m.client = client
	m.installer = inst

	// Start installation in background
	go func() {
		inst.Install(m.modules)
	}()

	return systemInfoMsg(info)
}

func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			if m.client != nil {
				m.client.Close()
			}
			return m, tea.Quit
		}

	case systemInfoMsg:
		m.systemInfo = msg
		return m, m.waitForProgress()

	case progressMsg:
		if msg.Module != "" {
			m.progress[msg.Module] = msg.Status
		}
		if msg.Output != "" {
			m.output = append(m.output, msg.Output)
			if len(m.output) > 20 {
				m.output = m.output[len(m.output)-20:]
			}
		}
		if msg.Error != nil {
			m.err = msg.Error
		}
		if msg.Done && msg.Module == "" {
			m.done = true
			return m, tea.Quit
		}
		m.currentStep = msg.Status
		return m, m.waitForProgress()

	case errorMsg:
		m.err = msg
		m.done = true
		return m, tea.Quit

	case time.Time:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return t
		})
	}

	return m, nil
}

func (m InstallModel) waitForProgress() tea.Cmd {
	if m.installer == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case update, ok := <-m.installer.GetProgressChannel():
			if !ok {
				return nil
			}
			return progressMsg(update)
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
}

func (m InstallModel) View() string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render("🎌 anime - Installing on " + m.server.Name))
	s.WriteString("\n\n")

	// System info
	if m.systemInfo != nil {
		s.WriteString(theme.InfoStyle.Render("System Information:"))
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("  OS: %s\n", m.systemInfo["os"]))
		s.WriteString(fmt.Sprintf("  GPU: %s\n", m.systemInfo["gpu"]))
		s.WriteString(fmt.Sprintf("  Free Disk: %s | Free RAM: %s\n", m.systemInfo["disk_free"], m.systemInfo["mem_free"]))
		s.WriteString("\n")
	}

	// Module progress
	s.WriteString(theme.InfoStyle.Render("Installation Progress:"))
	s.WriteString("\n\n")

	for _, mod := range config.AvailableModules {
		if status, ok := m.progress[mod.ID]; ok {
			icon := "⏳"
			style := theme.ProgressStyle
			switch status {
			case "Complete":
				icon = "✓"
				style = theme.CompleteStyle
			case "Failed":
				icon = "✗"
				style = theme.FailedStyle
			case "Starting", "Installing":
				icon = "▶"
			}
			s.WriteString(style.Render(fmt.Sprintf("  %s %s - %s\n", icon, mod.Name, status)))
		}
	}

	s.WriteString("\n")

	// Recent output
	if len(m.output) > 0 {
		s.WriteString(theme.HelpStyle.Render("Recent output:"))
		s.WriteString("\n")
		for _, line := range m.output {
			if len(line) > 100 {
				line = line[:100] + "..."
			}
			s.WriteString(theme.HelpStyle.Render("  " + line))
			if !strings.HasSuffix(line, "\n") {
				s.WriteString("\n")
			}
		}
	}

	// Stats
	elapsed := time.Since(m.startTime)
	cost := (elapsed.Minutes() / 60.0) * m.server.CostPerHour
	s.WriteString("\n")
	s.WriteString(theme.InfoStyle.Render(fmt.Sprintf("Elapsed: %s | Cost: $%.2f",
		elapsed.Round(time.Second), cost)))

	// Error
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	// Done
	if m.done {
		s.WriteString("\n\n")
		if m.err == nil {
			s.WriteString(theme.CompleteStyle.Render("✓ Installation complete!"))
		}
		s.WriteString(theme.HelpStyle.Render("\n\nPress q to exit"))
	} else {
		s.WriteString(theme.HelpStyle.Render("\n\nPress Ctrl+C to cancel"))
	}

	return s.String()
}
