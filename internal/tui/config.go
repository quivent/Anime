package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF69B4")).
		MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	unselectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	checkedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	uncheckedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF")).
		MarginTop(1)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginTop(1)
)

type screen int

const (
	screenMenu screen = iota
	screenServerList
	screenServerEdit
	screenModuleSelect
	screenAPIKeys
)

type ConfigModel struct {
	config       *config.Config
	screen       screen
	cursor       int
	err          error
	quitting     bool

	// Server editing
	inputs       []textinput.Model
	focusedInput int
	editingServer *config.Server
	editingIndex  int

	// Module selection
	selectedModules map[string]bool
	serverForModules string

	// API keys
	apiInputs []textinput.Model
	apiCursor int
}

func NewConfigModel() (ConfigModel, error) {
	cfg, err := config.Load()
	if err != nil {
		return ConfigModel{}, err
	}

	return ConfigModel{
		config:          cfg,
		screen:          screenMenu,
		selectedModules: make(map[string]bool),
	}, nil
}

func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen == screenMenu {
				m.quitting = true
				return m, tea.Quit
			}
			m.screen = screenMenu
			m.cursor = 0
			return m, nil

		case "esc":
			if m.screen != screenMenu {
				m.screen = screenMenu
				m.cursor = 0
			}
			return m, nil
		}

		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenServerList:
			return m.updateServerList(msg)
		case screenServerEdit:
			return m.updateServerEdit(msg)
		case screenModuleSelect:
			return m.updateModuleSelect(msg)
		case screenAPIKeys:
			return m.updateAPIKeys(msg)
		}
	}

	return m, nil
}

func (m ConfigModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 3 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0: // Manage Servers
			m.screen = screenServerList
			m.cursor = 0
		case 1: // Configure Modules
			if len(m.config.Servers) == 0 {
				m.err = fmt.Errorf("add a server first")
			} else {
				m.screen = screenServerList
				m.cursor = 0
				// We'll select modules after selecting server
			}
		case 2: // API Keys
			m.screen = screenAPIKeys
			m.apiInputs = m.createAPIInputs()
			m.apiCursor = 0
		case 3: // Save & Exit
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m ConfigModel) updateServerList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.config.Servers) {
			m.cursor++
		}
	case "enter":
		if m.cursor == len(m.config.Servers) {
			// Add new server
			m.screen = screenServerEdit
			m.inputs = m.createServerInputs(nil)
			m.focusedInput = 0
			m.editingIndex = -1
		} else {
			// Edit or configure modules for existing server
			m.screen = screenModuleSelect
			m.serverForModules = m.config.Servers[m.cursor].Name
			m.selectedModules = make(map[string]bool)
			for _, modID := range m.config.Servers[m.cursor].Modules {
				m.selectedModules[modID] = true
			}
			m.cursor = 0
		}
	case "d":
		if m.cursor < len(m.config.Servers) {
			m.config.DeleteServer(m.config.Servers[m.cursor].Name)
			if m.cursor >= len(m.config.Servers) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m ConfigModel) updateServerEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "tab", "down":
		m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
		return m, m.inputs[m.focusedInput].Focus()
	case "shift+tab", "up":
		m.focusedInput--
		if m.focusedInput < 0 {
			m.focusedInput = len(m.inputs) - 1
		}
		return m, m.inputs[m.focusedInput].Focus()
	case "enter":
		// Save server
		server := config.Server{
			Name:        m.inputs[0].Value(),
			Host:        m.inputs[1].Value(),
			User:        m.inputs[2].Value(),
			SSHKey:      m.inputs[3].Value(),
			CostPerHour: 20.0, // Default, parse from input[4] if needed
		}

		if m.editingIndex == -1 {
			m.config.AddServer(server)
		} else {
			m.config.UpdateServer(m.config.Servers[m.editingIndex].Name, server)
		}

		m.screen = screenServerList
		m.cursor = 0
		return m, nil
	}

	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

func (m ConfigModel) updateModuleSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(config.AvailableModules)-1 {
			m.cursor++
		}
	case " ":
		modID := config.AvailableModules[m.cursor].ID
		m.selectedModules[modID] = !m.selectedModules[modID]
	case "enter":
		// Save module selection
		var modules []string
		for id, selected := range m.selectedModules {
			if selected {
				modules = append(modules, id)
			}
		}

		server, err := m.config.GetServer(m.serverForModules)
		if err == nil {
			server.Modules = modules
			m.config.UpdateServer(m.serverForModules, *server)
		}

		m.screen = screenServerList
		m.cursor = 0
	}
	return m, nil
}

func (m ConfigModel) updateAPIKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "tab", "down":
		m.apiCursor = (m.apiCursor + 1) % len(m.apiInputs)
		return m, m.apiInputs[m.apiCursor].Focus()
	case "shift+tab", "up":
		m.apiCursor--
		if m.apiCursor < 0 {
			m.apiCursor = len(m.apiInputs) - 1
		}
		return m, m.apiInputs[m.apiCursor].Focus()
	case "enter":
		m.config.APIKeys.Anthropic = m.apiInputs[0].Value()
		m.config.APIKeys.OpenAI = m.apiInputs[1].Value()
		m.config.APIKeys.HuggingFace = m.apiInputs[2].Value()
		m.config.APIKeys.LambdaLabs = m.apiInputs[3].Value()

		m.screen = screenMenu
		m.cursor = 0
		return m, nil
	}

	m.apiInputs[m.apiCursor], cmd = m.apiInputs[m.apiCursor].Update(msg)
	return m, cmd
}

func (m ConfigModel) View() string {
	if m.quitting {
		return "Configuration saved! 👋\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("🎌 anime - Lambda Configuration"))
	s.WriteString("\n\n")

	switch m.screen {
	case screenMenu:
		s.WriteString(m.viewMenu())
	case screenServerList:
		s.WriteString(m.viewServerList())
	case screenServerEdit:
		s.WriteString(m.viewServerEdit())
	case screenModuleSelect:
		s.WriteString(m.viewModuleSelect())
	case screenAPIKeys:
		s.WriteString(m.viewAPIKeys())
	}

	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		m.err = nil
	}

	return s.String()
}

func (m ConfigModel) viewMenu() string {
	choices := []string{
		"Manage Servers",
		"Configure Modules",
		"API Keys",
		"Save & Exit",
	}

	var s strings.Builder
	for i, choice := range choices {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
			s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s\n", cursor, choice)))
		} else {
			s.WriteString(unselectedStyle.Render(fmt.Sprintf("%s %s\n", cursor, choice)))
		}
	}

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • q: quit"))
	return s.String()
}

func (m ConfigModel) viewServerList() string {
	var s strings.Builder
	s.WriteString("Servers:\n\n")

	for i, server := range m.config.Servers {
		cursor := " "
		moduleCount := len(server.Modules)
		cost := config.EstimateCost(server.Modules, server.CostPerHour)

		line := fmt.Sprintf("%s %s (%s@%s) - %d modules - est. $%.2f",
			cursor, server.Name, server.User, server.Host, moduleCount, cost)

		if m.cursor == i {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(unselectedStyle.Render(line) + "\n")
		}
	}

	cursor := " "
	if m.cursor == len(m.config.Servers) {
		cursor = "▶"
	}
	s.WriteString(fmt.Sprintf("\n%s Add new server\n", cursor))

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • d: delete • esc: back"))
	return s.String()
}

func (m ConfigModel) viewServerEdit() string {
	var s strings.Builder
	s.WriteString("Edit Server:\n\n")

	labels := []string{"Name:", "Host:", "User:", "SSH Key:", "Cost/hr:"}
	for i, input := range m.inputs {
		s.WriteString(labels[i] + "\n")
		s.WriteString(input.View() + "\n\n")
	}

	s.WriteString(helpStyle.Render("tab: next field • enter: save • esc: cancel"))
	return s.String()
}

func (m ConfigModel) viewModuleSelect() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Select modules for %s:\n\n", m.serverForModules))

	var selectedIDs []string
	for id, selected := range m.selectedModules {
		if selected {
			selectedIDs = append(selectedIDs, id)
		}
	}

	totalCost := config.EstimateCost(selectedIDs, 20.0)

	for i, mod := range config.AvailableModules {
		cursor := " "
		check := "☐"
		if m.selectedModules[mod.ID] {
			check = checkedStyle.Render("☑")
		} else {
			check = uncheckedStyle.Render(check)
		}

		if m.cursor == i {
			cursor = "▶"
		}

		line := fmt.Sprintf("%s %s %s (%dm) - %s",
			cursor, check, mod.Name, mod.TimeMinutes, mod.Description)

		if m.cursor == i {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(unselectedStyle.Render(line) + "\n")
		}
	}

	s.WriteString(infoStyle.Render(fmt.Sprintf("\nEstimated cost: $%.2f @ $20/hr", totalCost)))
	s.WriteString(helpStyle.Render("\n\n↑/↓: navigate • space: toggle • enter: save • esc: cancel"))
	return s.String()
}

func (m ConfigModel) viewAPIKeys() string {
	var s strings.Builder
	s.WriteString("API Keys:\n\n")

	labels := []string{"Anthropic:", "OpenAI:", "HuggingFace:", "Lambda Labs:"}
	for i, input := range m.apiInputs {
		s.WriteString(labels[i] + "\n")
		s.WriteString(input.View() + "\n\n")
	}

	s.WriteString(helpStyle.Render("tab: next field • enter: save • esc: cancel"))
	return s.String()
}

func (m ConfigModel) createServerInputs(server *config.Server) []textinput.Model {
	inputs := make([]textinput.Model, 5)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "lambda-gh200-1"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "192.168.1.100"
	inputs[1].CharLimit = 100
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "ubuntu"
	inputs[2].CharLimit = 50
	inputs[2].Width = 50

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "~/.ssh/lambda_key.pem"
	inputs[3].CharLimit = 200
	inputs[3].Width = 50

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "20.0"
	inputs[4].CharLimit = 10
	inputs[4].Width = 20

	if server != nil {
		inputs[0].SetValue(server.Name)
		inputs[1].SetValue(server.Host)
		inputs[2].SetValue(server.User)
		inputs[3].SetValue(server.SSHKey)
		inputs[4].SetValue(fmt.Sprintf("%.2f", server.CostPerHour))
	}

	return inputs
}

func (m ConfigModel) createAPIInputs() []textinput.Model {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sk-ant-..."
	inputs[0].EchoMode = textinput.EchoPassword
	inputs[0].EchoCharacter = '•'
	inputs[0].Focus()
	inputs[0].CharLimit = 200
	inputs[0].Width = 50
	inputs[0].SetValue(m.config.APIKeys.Anthropic)

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "sk-..."
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	inputs[1].CharLimit = 200
	inputs[1].Width = 50
	inputs[1].SetValue(m.config.APIKeys.OpenAI)

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "hf_..."
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'
	inputs[2].CharLimit = 200
	inputs[2].Width = 50
	inputs[2].SetValue(m.config.APIKeys.HuggingFace)

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "lambda_..."
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].CharLimit = 200
	inputs[3].Width = 50
	inputs[3].SetValue(m.config.APIKeys.LambdaLabs)

	return inputs
}
