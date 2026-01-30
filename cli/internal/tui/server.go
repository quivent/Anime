package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/ssh"
)

// ServerMonitorModel represents the GPU monitoring TUI
type ServerMonitorModel struct {
	host        string
	user        string
	isLocal     bool
	sshClient   *ssh.Client
	gpus        []GPUStats
	systemStats *SystemStats
	width       int
	height      int
	lastUpdate  time.Time
	err         error
	quitting    bool
	refreshRate time.Duration
}

// GPUStats holds detailed GPU statistics
type GPUStats struct {
	Index         int
	Name          string
	Temperature   int
	FanSpeed      int
	PowerUsage    int
	PowerLimit    int
	MemoryUsed    int
	MemoryTotal   int
	MemoryPercent float64
	Utilization   int // GPU compute utilization
	MemoryUtil    int // Memory controller utilization
	Processes     []GPUProcessInfo
}

// GPUProcessInfo holds process information using a GPU
type GPUProcessInfo struct {
	PID         int
	ProcessName string
	MemoryUsed  int
}

// SystemStats holds system-wide statistics
type SystemStats struct {
	Hostname      string
	Uptime        string
	DriverVersion string
	CUDAVersion   string
}

// Styles for server monitor
var (
	smTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6AC1")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 2).
			MarginBottom(1)

	smGPUCardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginRight(1).
			MarginBottom(1)

	smLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	smValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	smGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	smWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	smCritStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	smDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	smHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF"))

	smStatusBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Padding(0, 1)
)

// Key bindings for server monitor
type smKeyMap struct {
	Refresh key.Binding
	Quit    key.Binding
}

var smKeys = smKeyMap{
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
}

// Messages
type smTickMsg time.Time
type smDataMsg struct {
	gpus   []GPUStats
	system *SystemStats
	err    error
}

func smTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return smTickMsg(t)
	})
}

// NewServerMonitorModel creates a new server monitor model
func NewServerMonitorModel(host, user string, isLocal bool) (*ServerMonitorModel, error) {
	m := &ServerMonitorModel{
		host:        host,
		user:        user,
		isLocal:     isLocal,
		refreshRate: 2 * time.Second,
	}

	if !isLocal && host != "" {
		// Create SSH client for remote monitoring
		client, err := ssh.NewClientWithOptions(host, user, "", ssh.ClientOptions{
			StrictHostKeyChecking: true,
			Interactive:           true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s@%s: %w", user, host, err)
		}
		m.sshClient = client
	}

	return m, nil
}

func (m ServerMonitorModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		smTickCmd(m.refreshRate),
	)
}

func (m ServerMonitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, smKeys.Quit):
			m.quitting = true
			if m.sshClient != nil {
				m.sshClient.Close()
			}
			return m, tea.Quit

		case key.Matches(msg, smKeys.Refresh):
			return m, m.fetchData()
		}

	case smTickMsg:
		return m, tea.Batch(
			m.fetchData(),
			smTickCmd(m.refreshRate),
		)

	case smDataMsg:
		m.gpus = msg.gpus
		m.systemStats = msg.system
		m.err = msg.err
		m.lastUpdate = time.Now()
		return m, nil
	}

	return m, nil
}

func (m ServerMonitorModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Title bar
	title := "GPU MONITOR"
	if m.isLocal {
		title += " (Local)"
	} else {
		title += fmt.Sprintf(" (%s@%s)", m.user, m.host)
	}
	s.WriteString(smTitleStyle.Render(title) + "\n\n")

	// Error display
	if m.err != nil {
		s.WriteString(smCritStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	// System info bar
	if m.systemStats != nil {
		sysInfo := m.renderSystemInfo()
		s.WriteString(sysInfo + "\n")
	}

	// GPU cards - side by side
	if len(m.gpus) > 0 {
		gpuCards := m.renderGPUCards()
		s.WriteString(gpuCards)
	} else {
		s.WriteString(smDimStyle.Render("No GPUs detected or nvidia-smi not available\n"))
	}

	// Status bar
	statusBar := m.renderStatusBar()
	s.WriteString("\n" + statusBar)

	return s.String()
}

func (m ServerMonitorModel) renderSystemInfo() string {
	if m.systemStats == nil {
		return ""
	}

	sys := m.systemStats
	var parts []string

	if sys.Hostname != "" {
		parts = append(parts, smLabelStyle.Render("Host: ")+smValueStyle.Render(sys.Hostname))
	}
	if sys.DriverVersion != "" {
		parts = append(parts, smLabelStyle.Render("Driver: ")+smValueStyle.Render(sys.DriverVersion))
	}
	if sys.CUDAVersion != "" {
		parts = append(parts, smLabelStyle.Render("CUDA: ")+smValueStyle.Render(sys.CUDAVersion))
	}
	if sys.Uptime != "" {
		parts = append(parts, smLabelStyle.Render("Uptime: ")+smValueStyle.Render(sys.Uptime))
	}

	return smDimStyle.Render("─────────────────────────────────────────────────────────────────────────────────────") + "\n" +
		strings.Join(parts, "  |  ") + "\n" +
		smDimStyle.Render("─────────────────────────────────────────────────────────────────────────────────────") + "\n\n"
}

func (m ServerMonitorModel) renderGPUCards() string {
	if len(m.gpus) == 0 {
		return ""
	}

	// Calculate card width based on terminal width and number of GPUs
	numGPUs := len(m.gpus)
	cardWidth := 38 // minimum card width

	// Adjust card width based on available space
	if m.width > 0 {
		availableWidth := m.width - 4 // margins
		maxCardWidth := availableWidth/numGPUs - 2
		if maxCardWidth > cardWidth {
			cardWidth = maxCardWidth
		}
		if cardWidth > 50 {
			cardWidth = 50 // max card width
		}
	}

	var cards []string
	for _, gpu := range m.gpus {
		card := m.renderGPUCard(gpu, cardWidth)
		cards = append(cards, card)
	}

	// Join cards horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (m ServerMonitorModel) renderGPUCard(gpu GPUStats, width int) string {
	var content strings.Builder

	// GPU header
	header := fmt.Sprintf("GPU %d: %s", gpu.Index, gpu.Name)
	if len(header) > width-4 {
		header = header[:width-7] + "..."
	}
	content.WriteString(smHeaderStyle.Render(header) + "\n\n")

	// Temperature
	tempStyle := smGoodStyle
	if gpu.Temperature >= 80 {
		tempStyle = smCritStyle
	} else if gpu.Temperature >= 70 {
		tempStyle = smWarnStyle
	}
	tempBar := renderSmTempBar(gpu.Temperature, 100, width-18)
	content.WriteString(smLabelStyle.Render("Temp:    ") + tempBar + " " + tempStyle.Render(fmt.Sprintf("%d°C", gpu.Temperature)) + "\n")

	// Power
	powerPercent := 0
	if gpu.PowerLimit > 0 {
		powerPercent = (gpu.PowerUsage * 100) / gpu.PowerLimit
	}
	powerBar := renderSmUsageBar(powerPercent, width-18)
	content.WriteString(smLabelStyle.Render("Power:   ") + powerBar + " " + smValueStyle.Render(fmt.Sprintf("%dW/%dW", gpu.PowerUsage, gpu.PowerLimit)) + "\n")

	// Memory
	memBar := renderSmUsageBar(int(gpu.MemoryPercent), width-18)
	memStyle := smGoodStyle
	if gpu.MemoryPercent >= 90 {
		memStyle = smCritStyle
	} else if gpu.MemoryPercent >= 75 {
		memStyle = smWarnStyle
	}
	content.WriteString(smLabelStyle.Render("Memory:  ") + memBar + " " + memStyle.Render(fmt.Sprintf("%d/%dMB", gpu.MemoryUsed, gpu.MemoryTotal)) + "\n")

	// GPU Utilization
	utilBar := renderSmUsageBar(gpu.Utilization, width-18)
	utilStyle := smGoodStyle
	if gpu.Utilization >= 90 {
		utilStyle = smWarnStyle
	}
	content.WriteString(smLabelStyle.Render("GPU:     ") + utilBar + " " + utilStyle.Render(fmt.Sprintf("%d%%", gpu.Utilization)) + "\n")

	// Memory Controller Utilization (if available)
	if gpu.MemoryUtil > 0 {
		memUtilBar := renderSmUsageBar(gpu.MemoryUtil, width-18)
		content.WriteString(smLabelStyle.Render("Mem Ctrl:") + memUtilBar + " " + smValueStyle.Render(fmt.Sprintf("%d%%", gpu.MemoryUtil)) + "\n")
	}

	// Fan Speed (if available)
	if gpu.FanSpeed > 0 {
		fanBar := renderSmUsageBar(gpu.FanSpeed, width-18)
		content.WriteString(smLabelStyle.Render("Fan:     ") + fanBar + " " + smValueStyle.Render(fmt.Sprintf("%d%%", gpu.FanSpeed)) + "\n")
	}

	// Processes
	if len(gpu.Processes) > 0 {
		content.WriteString("\n" + smDimStyle.Render("Processes:") + "\n")
		for i, proc := range gpu.Processes {
			if i >= 3 {
				content.WriteString(smDimStyle.Render(fmt.Sprintf("  ... +%d more", len(gpu.Processes)-3)) + "\n")
				break
			}
			procName := proc.ProcessName
			if len(procName) > 15 {
				procName = procName[:12] + "..."
			}
			content.WriteString(smDimStyle.Render(fmt.Sprintf("  %s (%dMB)", procName, proc.MemoryUsed)) + "\n")
		}
	}

	// Apply card style with calculated width
	cardStyle := smGPUCardStyle.Width(width)
	return cardStyle.Render(content.String())
}

func renderSmUsageBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if width < 10 {
		width = 10
	}

	barWidth := width - 2 // account for brackets
	filled := (barWidth * percent) / 100
	empty := barWidth - filled

	// Color based on percentage
	var barStyle lipgloss.Style
	if percent < 60 {
		barStyle = smGoodStyle
	} else if percent < 85 {
		barStyle = smWarnStyle
	} else {
		barStyle = smCritStyle
	}

	bar := "[" + barStyle.Render(strings.Repeat("█", filled)) + smDimStyle.Render(strings.Repeat("░", empty)) + "]"
	return bar
}

func renderSmTempBar(temp int, maxTemp int, width int) string {
	if width < 10 {
		width = 10
	}

	percent := (temp * 100) / maxTemp
	if percent > 100 {
		percent = 100
	}

	barWidth := width - 2
	filled := (barWidth * percent) / 100
	empty := barWidth - filled

	// Color based on temperature
	var barStyle lipgloss.Style
	if temp < 60 {
		barStyle = smGoodStyle
	} else if temp < 75 {
		barStyle = smWarnStyle
	} else {
		barStyle = smCritStyle
	}

	bar := "[" + barStyle.Render(strings.Repeat("▓", filled)) + smDimStyle.Render(strings.Repeat("░", empty)) + "]"
	return bar
}

func (m ServerMonitorModel) renderStatusBar() string {
	lastUpdate := "Never"
	if !m.lastUpdate.IsZero() {
		lastUpdate = m.lastUpdate.Format("15:04:05")
	}

	status := fmt.Sprintf("Last update: %s  |  Refresh: %s  |  r: refresh  |  q: quit",
		lastUpdate,
		m.refreshRate.String())

	return smStatusBarStyle.Render(status)
}

func (m ServerMonitorModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		var gpus []GPUStats
		var system *SystemStats
		var err error

		if m.isLocal {
			gpus, system, err = fetchLocalGPUData()
		} else if m.sshClient != nil {
			gpus, system, err = fetchRemoteGPUData(m.sshClient)
		}

		return smDataMsg{
			gpus:   gpus,
			system: system,
			err:    err,
		}
	}
}

// RunServerMonitor starts the server monitor TUI
func RunServerMonitor(host, user string, isLocal bool) error {
	model, err := NewServerMonitorModel(host, user, isLocal)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Command runner type
type commandRunner func(cmd string) (string, error)

// Data fetching functions
func fetchLocalGPUData() ([]GPUStats, *SystemStats, error) {
	runner := func(cmd string) (string, error) {
		command := exec.Command("sh", "-c", cmd)
		output, err := command.Output()
		return string(output), err
	}
	return fetchGPUDataFromCommand(runner)
}

func fetchRemoteGPUData(client *ssh.Client) ([]GPUStats, *SystemStats, error) {
	runner := func(cmd string) (string, error) {
		return client.RunCommand(cmd)
	}
	return fetchGPUDataFromCommand(runner)
}

func fetchGPUDataFromCommand(run commandRunner) ([]GPUStats, *SystemStats, error) {
	// Query GPU information
	gpuQuery := "nvidia-smi --query-gpu=index,name,temperature.gpu,fan.speed,power.draw,power.limit,memory.used,memory.total,utilization.gpu,utilization.memory --format=csv,noheader,nounits 2>/dev/null"
	gpuOutput, err := run(gpuQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("nvidia-smi not available: %w", err)
	}

	var gpus []GPUStats
	lines := strings.Split(strings.TrimSpace(gpuOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) < 10 {
			continue
		}

		gpu := GPUStats{
			Index:       parseSmIntSafe(fields[0]),
			Name:        strings.TrimSpace(fields[1]),
			Temperature: parseSmIntSafe(fields[2]),
			FanSpeed:    parseSmIntSafe(fields[3]),
			PowerUsage:  parseSmIntSafe(fields[4]),
			PowerLimit:  parseSmIntSafe(fields[5]),
			MemoryUsed:  parseSmIntSafe(fields[6]),
			MemoryTotal: parseSmIntSafe(fields[7]),
			Utilization: parseSmIntSafe(fields[8]),
			MemoryUtil:  parseSmIntSafe(fields[9]),
		}

		if gpu.MemoryTotal > 0 {
			gpu.MemoryPercent = float64(gpu.MemoryUsed) / float64(gpu.MemoryTotal) * 100
		}

		// Get processes for this GPU
		procQuery := fmt.Sprintf("nvidia-smi --query-compute-apps=pid,name,used_memory --format=csv,noheader,nounits -i %d 2>/dev/null", gpu.Index)
		procOutput, _ := run(procQuery)
		if procOutput != "" {
			procLines := strings.Split(strings.TrimSpace(procOutput), "\n")
			for _, procLine := range procLines {
				if procLine == "" {
					continue
				}
				procFields := strings.Split(procLine, ", ")
				if len(procFields) >= 3 {
					gpu.Processes = append(gpu.Processes, GPUProcessInfo{
						PID:         parseSmIntSafe(procFields[0]),
						ProcessName: strings.TrimSpace(procFields[1]),
						MemoryUsed:  parseSmIntSafe(procFields[2]),
					})
				}
			}
		}

		gpus = append(gpus, gpu)
	}

	// Get system info
	system := &SystemStats{}

	// Hostname
	hostname, _ := run("hostname")
	system.Hostname = strings.TrimSpace(hostname)

	// Driver and CUDA version
	versionOutput, _ := run("nvidia-smi --query-gpu=driver_version --format=csv,noheader 2>/dev/null | head -1")
	system.DriverVersion = strings.TrimSpace(versionOutput)

	cudaOutput, _ := run("nvidia-smi | grep 'CUDA Version' | awk '{print $9}' 2>/dev/null")
	system.CUDAVersion = strings.TrimSpace(cudaOutput)

	// Uptime
	uptimeOutput, _ := run("uptime -p 2>/dev/null || uptime")
	system.Uptime = strings.TrimSpace(uptimeOutput)

	return gpus, system, nil
}

func parseSmIntSafe(s string) int {
	s = strings.TrimSpace(s)
	// Handle N/A or [N/A] values
	if s == "N/A" || s == "[N/A]" || s == "" {
		return 0
	}
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
