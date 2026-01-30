package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
)

// Panel types
type panelType int

const (
	panelGPU panelType = iota
	panelOllama
	panelSoftware
	panelCollections
	panelWorkflows
	panelTasks
	panelSystem
)

// Key bindings
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left panel"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right panel"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next panel"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev panel"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Workstation-specific styles using centralized theme
var (
	workstationTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.BrightPink).
			Background(theme.BgBlack).
			Padding(0, 1)

	// wsSelectedItemStyle uses BrightPink for consistency
	wsSelectedItemStyle = lipgloss.NewStyle().
				Foreground(theme.BrightPink).
				Bold(true).
				PaddingLeft(2)

	wsItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)
)

// Model
type workstationModel struct {
	activePanel   panelType
	panels        []panelType
	width         int
	height        int
	viewport      viewport.Model
	scrollPos     map[panelType]int
	gpuData       *GPUData
	ollamaModels  []OllamaModel
	software      []Software
	collections   []config.Collection
	workflows     []Workflow
	tasks         []Task
	systemInfo    *SystemInfo
	lastUpdate    time.Time
	showHelp      bool
	cfg           *config.Config
}

// Data structures
type GPUData struct {
	GPUs []GPU
}

type GPU struct {
	ID          int
	Name        string
	Temperature int
	PowerUsage  int
	PowerLimit  int
	MemoryUsed  int
	MemoryTotal int
	Utilization int
	Processes   []GPUProcess
}

type GPUProcess struct {
	PID     int
	Name    string
	Memory  int
	Command string
}

type OllamaModel struct {
	Name         string
	Size         string
	Modified     string
	Quantization string
}

type Software struct {
	Name        string
	Version     string
	Status      string
	Description string
}

type Workflow struct {
	Name        string
	Status      string
	LastRun     string
	Description string
}

type Task struct {
	Name     string
	Progress int
	Status   string
	ETA      string
}

type SystemInfo struct {
	Hostname    string
	OS          string
	Arch        string
	CPUUsage    float64
	MemoryUsed  int
	MemoryTotal int
	DiskUsed    int
	DiskTotal   int
	Uptime      string
}

// Messages
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Initialize
func initialWorkstationModel() workstationModel {
	cfg, _ := config.Load()

	vp := viewport.New(78, 20)

	m := workstationModel{
		activePanel: panelGPU,
		panels: []panelType{
			panelGPU,
			panelOllama,
			panelSoftware,
			panelCollections,
			panelWorkflows,
			panelTasks,
			panelSystem,
		},
		viewport:  vp,
		scrollPos: make(map[panelType]int),
		cfg:       cfg,
	}

	m.refreshData()
	return m
}

func (m *workstationModel) refreshData() {
	m.gpuData = fetchGPUData()
	m.ollamaModels = fetchOllamaModels()
	m.software = fetchSoftware()
	if m.cfg != nil {
		m.collections = m.cfg.Collections
	}
	m.workflows = fetchWorkflows()
	m.tasks = fetchTasks()
	m.systemInfo = fetchSystemInfo()
	m.lastUpdate = time.Now()
}

func (m workstationModel) Init() tea.Cmd {
	return tickCmd()
}

func (m workstationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, keys.Refresh):
			m.refreshData()
			return m, nil

		case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Right):
			m.activePanel = m.nextPanel()
			return m, nil

		case key.Matches(msg, keys.ShiftTab), key.Matches(msg, keys.Left):
			m.activePanel = m.prevPanel()
			return m, nil

		case key.Matches(msg, keys.Up):
			if m.scrollPos[m.activePanel] > 0 {
				m.scrollPos[m.activePanel]--
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			m.scrollPos[m.activePanel]++
			return m, nil
		}

	case tickMsg:
		m.refreshData()
		return m, tickCmd()
	}

	return m, nil
}

func (m workstationModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Title
	title := workstationTitleStyle.Render("  ANIME WORKSTATION MONITOR  ")
	lastUpdate := theme.DimTextStyle.Render(fmt.Sprintf("Last update: %s", m.lastUpdate.Format("15:04:05")))
	titleBar := lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", lastUpdate)
	s.WriteString(titleBar + "\n\n")

	// Main content - two columns
	leftCol := m.renderLeftColumn()
	rightCol := m.renderRightColumn()

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol)
	s.WriteString(content + "\n")

	// Help bar
	if m.showHelp {
		s.WriteString("\n" + m.renderHelp())
	} else {
		helpBar := theme.HelpStyle.Render("↑↓ scroll • ←→ switch panel • tab next • r refresh • ? help • q quit")
		s.WriteString("\n" + helpBar)
	}

	return s.String()
}

func (m workstationModel) renderLeftColumn() string {
	var panels []string

	// GPU Panel
	gpuContent := m.renderGPUPanel()
	if m.activePanel == panelGPU {
		panels = append(panels, theme.ActivePanelStyle.Render(gpuContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(gpuContent))
	}

	// Ollama Panel
	ollamaContent := m.renderOllamaPanel()
	if m.activePanel == panelOllama {
		panels = append(panels, theme.ActivePanelStyle.Render(ollamaContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(ollamaContent))
	}

	// Software Panel
	softwareContent := m.renderSoftwarePanel()
	if m.activePanel == panelSoftware {
		panels = append(panels, theme.ActivePanelStyle.Render(softwareContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(softwareContent))
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

func (m workstationModel) renderRightColumn() string {
	var panels []string

	// Collections Panel
	collectionsContent := m.renderCollectionsPanel()
	if m.activePanel == panelCollections {
		panels = append(panels, theme.ActivePanelStyle.Render(collectionsContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(collectionsContent))
	}

	// Workflows Panel
	workflowsContent := m.renderWorkflowsPanel()
	if m.activePanel == panelWorkflows {
		panels = append(panels, theme.ActivePanelStyle.Render(workflowsContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(workflowsContent))
	}

	// Tasks Panel
	tasksContent := m.renderTasksPanel()
	if m.activePanel == panelTasks {
		panels = append(panels, theme.ActivePanelStyle.Render(tasksContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(tasksContent))
	}

	// System Panel
	systemContent := m.renderSystemPanel()
	if m.activePanel == panelSystem {
		panels = append(panels, theme.ActivePanelStyle.Render(systemContent))
	} else {
		panels = append(panels, theme.InactivePanelStyle.Render(systemContent))
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

func (m workstationModel) renderGPUPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("GPU METRICS") + "\n")

	if m.gpuData == nil || len(m.gpuData.GPUs) == 0 {
		s.WriteString(theme.DimTextStyle.Render("No GPU detected"))
		return s.String()
	}

	for i, gpu := range m.gpuData.GPUs {
		if i > 0 {
			s.WriteString("\n")
		}

		// GPU name
		s.WriteString(theme.LabelStyle.Render(fmt.Sprintf("GPU %d: ", gpu.ID)) + theme.ValueStyle.Render(gpu.Name) + "\n")

		// Temperature
		tempColor := theme.SuccessStyle
		if gpu.Temperature > 80 {
			tempColor = theme.ErrorStyle
		} else if gpu.Temperature > 70 {
			tempColor = theme.WarningStyle
		}
		s.WriteString(theme.LabelStyle.Render("  Temp: ") + tempColor.Render(fmt.Sprintf("%d°C", gpu.Temperature)) + "\n")

		// Power
		s.WriteString(theme.LabelStyle.Render("  Power: ") + theme.ValueStyle.Render(fmt.Sprintf("%dW / %dW", gpu.PowerUsage, gpu.PowerLimit)) + "\n")

		// Memory
		memPercent := float64(gpu.MemoryUsed) / float64(gpu.MemoryTotal) * 100
		memBar := renderProgressBar(int(memPercent), 20)
		s.WriteString(theme.LabelStyle.Render("  Memory: ") + memBar + theme.ValueStyle.Render(fmt.Sprintf(" %dMB / %dMB", gpu.MemoryUsed, gpu.MemoryTotal)) + "\n")

		// Utilization
		utilBar := renderProgressBar(gpu.Utilization, 20)
		s.WriteString(theme.LabelStyle.Render("  Util: ") + utilBar + theme.ValueStyle.Render(fmt.Sprintf(" %d%%", gpu.Utilization)) + "\n")

		// Processes
		if len(gpu.Processes) > 0 {
			s.WriteString(theme.LabelStyle.Render("  Processes: ") + theme.DimTextStyle.Render(fmt.Sprintf("(%d)", len(gpu.Processes))) + "\n")
			for j, proc := range gpu.Processes {
				if j >= 2 { // Show max 2 processes
					s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    ... and %d more", len(gpu.Processes)-2)) + "\n")
					break
				}
				s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    %s (%dMB)", proc.Name, proc.Memory)) + "\n")
			}
		}
	}

	return s.String()
}

func (m workstationModel) renderOllamaPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("OLLAMA MODELS") + "\n")

	if len(m.ollamaModels) == 0 {
		s.WriteString(theme.DimTextStyle.Render("No models found"))
		return s.String()
	}

	startIdx := m.scrollPos[panelOllama]
	endIdx := startIdx + 5
	if endIdx > len(m.ollamaModels) {
		endIdx = len(m.ollamaModels)
	}
	if startIdx >= len(m.ollamaModels) {
		startIdx = 0
		m.scrollPos[panelOllama] = 0
	}

	for i := startIdx; i < endIdx; i++ {
		model := m.ollamaModels[i]
		prefix := "  "
		if m.activePanel == panelOllama {
			prefix = wsSelectedItemStyle.Render("▸ ")
		}

		s.WriteString(prefix + theme.LabelStyle.Render(model.Name) + "\n")
		s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    %s • %s", model.Size, model.Quantization)) + "\n")
	}

	if len(m.ollamaModels) > 5 {
		s.WriteString("\n" + theme.DimTextStyle.Render(fmt.Sprintf("    Showing %d-%d of %d", startIdx+1, endIdx, len(m.ollamaModels))))
	}

	return s.String()
}

func (m workstationModel) renderSoftwarePanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("INSTALLED SOFTWARE") + "\n")

	if len(m.software) == 0 {
		s.WriteString(theme.DimTextStyle.Render("Scanning..."))
		return s.String()
	}

	startIdx := m.scrollPos[panelSoftware]
	endIdx := startIdx + 5
	if endIdx > len(m.software) {
		endIdx = len(m.software)
	}
	if startIdx >= len(m.software) {
		startIdx = 0
		m.scrollPos[panelSoftware] = 0
	}

	for i := startIdx; i < endIdx; i++ {
		sw := m.software[i]
		style := theme.SuccessStyle
		if sw.Status == "inactive" {
			style = theme.DimTextStyle
		} else if sw.Status == "error" {
			style = theme.ErrorStyle
		}

		prefix := "  "
		if m.activePanel == panelSoftware {
			prefix = wsSelectedItemStyle.Render("▸ ")
		}

		s.WriteString(prefix + theme.LabelStyle.Render(sw.Name) + " " + theme.DimTextStyle.Render(sw.Version) + " " + style.Render("●") + "\n")
	}

	if len(m.software) > 5 {
		s.WriteString("\n" + theme.DimTextStyle.Render(fmt.Sprintf("    Showing %d-%d of %d", startIdx+1, endIdx, len(m.software))))
	}

	return s.String()
}

func (m workstationModel) renderCollectionsPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("ASSET COLLECTIONS") + "\n")

	if len(m.collections) == 0 {
		s.WriteString(theme.DimTextStyle.Render("No collections configured"))
		return s.String()
	}

	for i, coll := range m.collections {
		prefix := "  "
		if m.activePanel == panelCollections {
			prefix = wsSelectedItemStyle.Render("▸ ")
		}

		s.WriteString(prefix + theme.LabelStyle.Render(coll.Name) + "\n")
		s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    %s", coll.Path)) + "\n")

		if i < len(m.collections)-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m workstationModel) renderWorkflowsPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("WORKFLOWS") + "\n")

	if len(m.workflows) == 0 {
		s.WriteString(theme.DimTextStyle.Render("No active workflows"))
		return s.String()
	}

	for i, wf := range m.workflows {
		style := theme.SuccessStyle
		if wf.Status == "failed" {
			style = theme.ErrorStyle
		} else if wf.Status == "running" {
			style = theme.WarningStyle
		}

		prefix := "  "
		if m.activePanel == panelWorkflows {
			prefix = wsSelectedItemStyle.Render("▸ ")
		}

		s.WriteString(prefix + theme.LabelStyle.Render(wf.Name) + " " + style.Render(wf.Status) + "\n")
		s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    Last: %s", wf.LastRun)) + "\n")

		if i < len(m.workflows)-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m workstationModel) renderTasksPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("TASKS IN PROGRESS") + "\n")

	if len(m.tasks) == 0 {
		s.WriteString(theme.DimTextStyle.Render("No active tasks"))
		return s.String()
	}

	for i, task := range m.tasks {
		style := theme.WarningStyle
		if task.Status == "completed" {
			style = theme.SuccessStyle
		} else if task.Status == "failed" {
			style = theme.ErrorStyle
		}
		_ = style // Mark as used

		prefix := "  "
		if m.activePanel == panelTasks {
			prefix = wsSelectedItemStyle.Render("▸ ")
		}

		s.WriteString(prefix + theme.LabelStyle.Render(task.Name) + "\n")

		progressBar := renderProgressBar(task.Progress, 30)
		s.WriteString("    " + progressBar + " " + theme.ValueStyle.Render(fmt.Sprintf("%d%%", task.Progress)) + "\n")

		if task.ETA != "" {
			s.WriteString(theme.DimTextStyle.Render(fmt.Sprintf("    ETA: %s", task.ETA)) + "\n")
		}

		if i < len(m.tasks)-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m workstationModel) renderSystemPanel() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("SYSTEM INFO") + "\n")

	if m.systemInfo == nil {
		s.WriteString(theme.DimTextStyle.Render("Loading..."))
		return s.String()
	}

	sys := m.systemInfo

	s.WriteString(theme.LabelStyle.Render("Host: ") + theme.ValueStyle.Render(sys.Hostname) + "\n")
	s.WriteString(theme.LabelStyle.Render("OS: ") + theme.ValueStyle.Render(fmt.Sprintf("%s %s", sys.OS, sys.Arch)) + "\n")
	s.WriteString(theme.LabelStyle.Render("Uptime: ") + theme.ValueStyle.Render(sys.Uptime) + "\n\n")

	// CPU
	cpuBar := renderProgressBar(int(sys.CPUUsage), 20)
	s.WriteString(theme.LabelStyle.Render("CPU: ") + cpuBar + theme.ValueStyle.Render(fmt.Sprintf(" %.1f%%", sys.CPUUsage)) + "\n")

	// Memory
	memPercent := float64(sys.MemoryUsed) / float64(sys.MemoryTotal) * 100
	memBar := renderProgressBar(int(memPercent), 20)
	s.WriteString(theme.LabelStyle.Render("RAM: ") + memBar + theme.ValueStyle.Render(fmt.Sprintf(" %d/%d GB", sys.MemoryUsed, sys.MemoryTotal)) + "\n")

	// Disk
	diskPercent := float64(sys.DiskUsed) / float64(sys.DiskTotal) * 100
	diskBar := renderProgressBar(int(diskPercent), 20)
	s.WriteString(theme.LabelStyle.Render("Disk: ") + diskBar + theme.ValueStyle.Render(fmt.Sprintf(" %d/%d GB", sys.DiskUsed, sys.DiskTotal)) + "\n")

	return s.String()
}

func (m workstationModel) renderHelp() string {
	var s strings.Builder
	s.WriteString(theme.HeaderStyle.Render("KEYBOARD SHORTCUTS") + "\n\n")
	s.WriteString(theme.LabelStyle.Render("  ↑/↓ or k/j") + "    Scroll within panel\n")
	s.WriteString(theme.LabelStyle.Render("  ←/→ or h/l") + "    Switch between panels\n")
	s.WriteString(theme.LabelStyle.Render("  Tab") + "           Next panel\n")
	s.WriteString(theme.LabelStyle.Render("  Shift+Tab") + "     Previous panel\n")
	s.WriteString(theme.LabelStyle.Render("  r") + "             Refresh data\n")
	s.WriteString(theme.LabelStyle.Render("  ?") + "             Toggle help\n")
	s.WriteString(theme.LabelStyle.Render("  q or Ctrl+C") + "   Quit\n")
	return theme.HelpStyle.Render(s.String())
}

func (m workstationModel) nextPanel() panelType {
	for i, p := range m.panels {
		if p == m.activePanel {
			if i+1 < len(m.panels) {
				return m.panels[i+1]
			}
			return m.panels[0]
		}
	}
	return m.activePanel
}

func (m workstationModel) prevPanel() panelType {
	for i, p := range m.panels {
		if p == m.activePanel {
			if i-1 >= 0 {
				return m.panels[i-1]
			}
			return m.panels[len(m.panels)-1]
		}
	}
	return m.activePanel
}

// Helper functions
func renderProgressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * float64(percent) / 100.0)
	empty := width - filled

	var bar strings.Builder
	bar.WriteString("[")

	// Color based on percentage
	var barStyle lipgloss.Style
	if percent < 60 {
		barStyle = theme.SuccessStyle
	} else if percent < 85 {
		barStyle = theme.WarningStyle
	} else {
		barStyle = theme.ErrorStyle
	}

	bar.WriteString(barStyle.Render(strings.Repeat("█", filled)))
	bar.WriteString(theme.DimTextStyle.Render(strings.Repeat("░", empty)))
	bar.WriteString("]")

	return bar.String()
}

// Data fetching functions
func fetchGPUData() *GPUData {
	// Try nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,temperature.gpu,power.draw,power.limit,memory.used,memory.total,utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return &GPUData{GPUs: []GPU{}}
	}

	var gpus []GPU
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) < 8 {
			continue
		}

		gpu := GPU{
			ID:          parseInt(fields[0]),
			Name:        fields[1],
			Temperature: parseInt(fields[2]),
			PowerUsage:  parseInt(fields[3]),
			PowerLimit:  parseInt(fields[4]),
			MemoryUsed:  parseInt(fields[5]),
			MemoryTotal: parseInt(fields[6]),
			Utilization: parseInt(fields[7]),
			Processes:   fetchGPUProcesses(parseInt(fields[0])),
		}
		gpus = append(gpus, gpu)
	}

	return &GPUData{GPUs: gpus}
}

func fetchGPUProcesses(gpuID int) []GPUProcess {
	cmd := exec.Command("nvidia-smi", "--query-compute-apps=pid,name,used_memory", "--format=csv,noheader,nounits", "-i", fmt.Sprintf("%d", gpuID))
	output, err := cmd.Output()
	if err != nil {
		return []GPUProcess{}
	}

	var processes []GPUProcess
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) < 3 {
			continue
		}

		processes = append(processes, GPUProcess{
			PID:    parseInt(fields[0]),
			Name:   fields[1],
			Memory: parseInt(fields[2]),
		})
	}

	return processes
}

func fetchOllamaModels() []OllamaModel {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return []OllamaModel{}
	}

	var models []OllamaModel
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Skip header
	for i, line := range lines {
		if i == 0 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		model := OllamaModel{
			Name:     fields[0],
			Size:     fields[1],
			Modified: strings.Join(fields[2:], " "),
		}

		// Parse quantization from name
		if strings.Contains(fields[0], "q4") {
			model.Quantization = "Q4"
		} else if strings.Contains(fields[0], "q8") {
			model.Quantization = "Q8"
		} else {
			model.Quantization = "Full"
		}

		models = append(models, model)
	}

	return models
}

func fetchSoftware() []Software {
	software := []Software{}

	// Check common ML/AI software
	checks := map[string][]string{
		"Python":     {"python3", "--version"},
		"PyTorch":    {"python3", "-c", "import torch; print(torch.__version__)"},
		"ComfyUI":    {"which", "comfyui"},
		"Blender":    {"blender", "--version"},
		"FFmpeg":     {"ffmpeg", "-version"},
		"Git":        {"git", "--version"},
		"Docker":     {"docker", "--version"},
		"CUDA":       {"nvcc", "--version"},
		"Node.js":    {"node", "--version"},
		"Ollama":     {"ollama", "--version"},
	}

	for name, cmdArgs := range checks {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.Output()

		sw := Software{
			Name: name,
		}

		if err != nil {
			sw.Status = "inactive"
			sw.Version = "not installed"
		} else {
			sw.Status = "active"
			// Extract version from output
			versionStr := strings.TrimSpace(string(output))
			lines := strings.Split(versionStr, "\n")
			if len(lines) > 0 {
				sw.Version = strings.TrimSpace(lines[0])
				// Clean up version string
				if len(sw.Version) > 40 {
					sw.Version = sw.Version[:40] + "..."
				}
			}
		}

		software = append(software, sw)
	}

	return software
}

func fetchWorkflows() []Workflow {
	// This would integrate with your actual workflow system
	// For now, return empty or mock data
	return []Workflow{}
}

func fetchTasks() []Task {
	// This would integrate with your actual task system
	// For now, return empty or mock data
	return []Task{}
}

func fetchSystemInfo() *SystemInfo {
	hostname, _ := os.Hostname()

	info := &SystemInfo{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	// CPU usage
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", "ps -A -o %cpu | awk '{s+=$1} END {print s}'")
		output, err := cmd.Output()
		if err == nil {
			info.CPUUsage = parseFloat(strings.TrimSpace(string(output)))
		}
	}

	// Memory info (Linux/macOS)
	if runtime.GOOS == "linux" {
		cmd := exec.Command("sh", "-c", "free -g | grep Mem | awk '{print $3,$2}'")
		output, err := cmd.Output()
		if err == nil {
			fields := strings.Fields(string(output))
			if len(fields) >= 2 {
				info.MemoryUsed = parseInt(fields[0])
				info.MemoryTotal = parseInt(fields[1])
			}
		}
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", "vm_stat | grep 'Pages active' | awk '{print $3}' | sed 's/\\.//'")
		output, _ := cmd.Output()
		pageSize := 4096
		activePages := parseInt(strings.TrimSpace(string(output)))
		info.MemoryUsed = (activePages * pageSize) / (1024 * 1024 * 1024)

		cmd = exec.Command("sh", "-c", "sysctl -n hw.memsize")
		output, _ = cmd.Output()
		totalBytes := parseInt(strings.TrimSpace(string(output)))
		info.MemoryTotal = totalBytes / (1024 * 1024 * 1024)
	}

	// Disk info
	cmd := exec.Command("df", "-BG", "/")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 4 {
				info.DiskUsed = parseInt(strings.TrimSuffix(fields[2], "G"))
				info.DiskTotal = parseInt(strings.TrimSuffix(fields[1], "G"))
			}
		}
	}

	// Uptime
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.Command("uptime", "-p")
		output, err := cmd.Output()
		if err == nil {
			info.Uptime = strings.TrimSpace(string(output))
		}
	}

	return info
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

// RunWorkstation starts the workstation TUI
func RunWorkstation() error {
	m := initialWorkstationModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
