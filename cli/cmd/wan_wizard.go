package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/joshkornreich/anime/internal/gpu"
)

// ─── wizard screens ───

type wizScreen int

const (
	wizWelcome wizScreen = iota
	wizLevel
	wizConfirm
	wizInstalling
	wizDone
)

// ─── styles ───

var (
	wizTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).MarginBottom(1)
	wizSubStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	wizDimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wizGoodStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	wizWarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	wizErrStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	wizBorderStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("213")).Padding(1, 2)
	wizSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Background(lipgloss.Color("236")).Padding(0, 1)
	wizNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1)
)

// ─── level option ───

type wizLevelOption struct {
	id       string
	name     string
	size     string
	time     string
	vram     string
	models   string
	preset   string
}

var wizLevels = []wizLevelOption{
	{
		id:     "minimal",
		name:   "Minimal",
		size:   "~20 GB",
		time:   "~5 min",
		vram:   "12+ GB",
		models: "5B TI2V only",
		preset: "ti2v-5b",
	},
	{
		id:     "standard",
		name:   "Standard",
		size:   "~35 GB",
		time:   "~10 min",
		vram:   "24+ GB",
		models: "14B T2V dual-expert + 4-step LoRAs",
		preset: "t2v-14b-dual-fast",
	},
	{
		id:     "full",
		name:   "Full",
		size:   "~85 GB",
		time:   "~22 min",
		vram:   "48+ GB",
		models: "Everything: T2V + I2V dual + 5B + LoRAs",
		preset: "t2v-14b-dual-maxq",
	},
}

// ─── model ───

type wizModel struct {
	screen       wizScreen
	width        int
	height       int
	gpuInfo      *gpu.SystemInfo
	vram         int
	levelIdx     int // cursor in wizLevels
	recIdx       int // recommended index
	spinner      spinner.Model
	phaseIdx     int
	phases       []wizPhaseState
	installErr   error
	launchStudio bool
}

type wizPhaseState struct {
	name   string
	status string // "pending", "running", "done", "failed"
	detail string
}

// bubbletea messages
type wizPhaseStartMsg struct{ idx int }
type wizPhaseDoneMsg struct {
	idx    int
	detail string
}
type wizPhaseFailMsg struct {
	idx int
	err error
}
type wizAllDoneMsg struct{}

func init() {
	wizardCmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive guided setup for the Wan video generation stack",
		Long: `Walk through the complete Wan stack setup step by step.

The wizard detects your GPU, recommends an install level, explains what
will be downloaded, and runs the setup with live progress. At the end,
it offers to launch the Comfort studio.

This is the recommended way to set up Wan for the first time.`,
		RunE: runWanWizard,
	}
	wanCmd.AddCommand(wizardCmd)
}

func runWanWizard(cmd *cobra.Command, args []string) error {
	g := gpu.GetSystemInfo()
	vram := g.TotalVRAM

	// Pick recommended level
	recIdx := 0
	switch {
	case vram >= 48:
		recIdx = 2
	case vram >= 24:
		recIdx = 1
	default:
		recIdx = 0
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = wizSubStyle

	m := &wizModel{
		screen:   wizWelcome,
		gpuInfo:  g,
		vram:     vram,
		levelIdx: recIdx,
		recIdx:   recIdx,
		spinner:  sp,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}

	// After TUI exits, if user chose to launch studio, do it
	if rm, ok := result.(*wizModel); ok && rm.launchStudio {
		level := wizLevels[rm.levelIdx].id
		fmt.Println()
		fmt.Printf("  Launching studio (level=%s)...\n\n", level)
		return runWanStudio(&cobra.Command{}, []string{"--" + level, "--yes"})
	}

	return nil
}

func (m *wizModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *wizModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch m.screen {
		case wizWelcome:
			switch msg.String() {
			case "enter", " ":
				m.screen = wizLevel
			case "q", "ctrl+c":
				return m, tea.Quit
			}

		case wizLevel:
			switch msg.String() {
			case "up", "k":
				if m.levelIdx > 0 {
					m.levelIdx--
				}
			case "down", "j":
				if m.levelIdx < len(wizLevels)-1 {
					m.levelIdx++
				}
			case "enter", " ":
				m.screen = wizConfirm
			case "esc":
				m.screen = wizWelcome
			case "q", "ctrl+c":
				return m, tea.Quit
			}

		case wizConfirm:
			switch msg.String() {
			case "enter", "y":
				m.screen = wizInstalling
				m.phases = buildWizPhases(wizLevels[m.levelIdx].id)
				cmds = append(cmds, m.runPhase(0))
			case "esc", "backspace":
				m.screen = wizLevel
			case "q", "ctrl+c":
				return m, tea.Quit
			}

		case wizInstalling:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			}

		case wizDone:
			switch msg.String() {
			case "enter", "s":
				m.launchStudio = true
				return m, tea.Quit
			case "q", "ctrl+c", "esc":
				return m, tea.Quit
			}
		}

	case wizPhaseStartMsg:
		if msg.idx < len(m.phases) {
			m.phases[msg.idx].status = "running"
			m.phaseIdx = msg.idx
		}

	case wizPhaseDoneMsg:
		if msg.idx < len(m.phases) {
			m.phases[msg.idx].status = "done"
			m.phases[msg.idx].detail = msg.detail
			// Start next phase
			next := msg.idx + 1
			if next < len(m.phases) {
				cmds = append(cmds, m.runPhase(next))
			} else {
				m.screen = wizDone
			}
		}

	case wizPhaseFailMsg:
		if msg.idx < len(m.phases) {
			m.phases[msg.idx].status = "failed"
			m.phases[msg.idx].detail = msg.err.Error()
			m.installErr = msg.err
			m.screen = wizDone
		}

	case wizAllDoneMsg:
		m.screen = wizDone
	}

	// Spinner
	if m.screen == wizInstalling {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *wizModel) View() string {
	switch m.screen {
	case wizWelcome:
		return m.viewWelcome()
	case wizLevel:
		return m.viewLevel()
	case wizConfirm:
		return m.viewConfirm()
	case wizInstalling:
		return m.viewInstalling()
	case wizDone:
		return m.viewDone()
	}
	return ""
}

func (m *wizModel) viewWelcome() string {
	var s strings.Builder

	s.WriteString(wizTitleStyle.Render("🌀 Wan Setup Wizard"))
	s.WriteString("\n\n")

	s.WriteString(wizSubStyle.Render("  System Detection"))
	s.WriteString("\n\n")

	if m.gpuInfo.Available && len(m.gpuInfo.GPUs) > 0 {
		name := m.gpuInfo.GPUs[0].Name
		s.WriteString(fmt.Sprintf("  %s  GPU         %s\n", wizGoodStyle.Render("✓"), name))
		s.WriteString(fmt.Sprintf("  %s  VRAM        %d GB\n", wizGoodStyle.Render("✓"), m.vram))
		if m.gpuInfo.DriverVersion != "" {
			drv := m.gpuInfo.DriverVersion
			if m.gpuInfo.CUDAVersion != "" {
				drv += " (CUDA " + m.gpuInfo.CUDAVersion + ")"
			}
			s.WriteString(fmt.Sprintf("  %s  Driver      %s\n", wizGoodStyle.Render("✓"), drv))
		}
	} else {
		s.WriteString(fmt.Sprintf("  %s  GPU         not detected\n", wizWarnStyle.Render("⚠")))
		s.WriteString(wizDimStyle.Render("     Setup will continue — you can install locally and render on a remote GPU.\n"))
	}

	// Check existing installation
	home, _ := os.UserHomeDir()
	comfyExists := exists(filepath.Join(home, "ComfyUI", "main.py"))
	comfortExists := exists(filepath.Join(home, "Comfort", "comfort-ui", "dist", "index.html"))

	s.WriteString("\n")
	s.WriteString(wizSubStyle.Render("  Current State"))
	s.WriteString("\n\n")

	if comfyExists {
		s.WriteString(fmt.Sprintf("  %s  ComfyUI     installed\n", wizGoodStyle.Render("✓")))
	} else {
		s.WriteString(fmt.Sprintf("  %s  ComfyUI     not installed\n", wizDimStyle.Render("·")))
	}
	if comfortExists {
		s.WriteString(fmt.Sprintf("  %s  Comfort UI  installed\n", wizGoodStyle.Render("✓")))
	} else {
		s.WriteString(fmt.Sprintf("  %s  Comfort UI  not installed\n", wizDimStyle.Render("·")))
	}

	s.WriteString("\n\n")
	s.WriteString(wizDimStyle.Render("  Press enter to continue • q to quit"))

	return wizBorderStyle.Render(s.String())
}

func (m *wizModel) viewLevel() string {
	var s strings.Builder

	s.WriteString(wizTitleStyle.Render("🌀 Choose Install Level"))
	s.WriteString("\n\n")

	for i, lvl := range wizLevels {
		cursor := "  "
		style := wizNormalStyle
		if i == m.levelIdx {
			cursor = "▸ "
			style = wizSelectedStyle
		}

		rec := ""
		if i == m.recIdx {
			rec = wizGoodStyle.Render(" ← recommended")
		}

		s.WriteString(cursor)
		s.WriteString(style.Render(fmt.Sprintf("%-10s", lvl.name)))
		s.WriteString(rec)
		s.WriteString("\n")

		// Detail row for selected
		if i == m.levelIdx {
			s.WriteString(wizDimStyle.Render(fmt.Sprintf("      VRAM: %s  •  Download: %s  •  Time: %s\n", lvl.vram, lvl.size, lvl.time)))
			s.WriteString(wizDimStyle.Render(fmt.Sprintf("      Models: %s\n", lvl.models)))
			s.WriteString(wizDimStyle.Render(fmt.Sprintf("      Default preset: %s\n", lvl.preset)))
		}
		s.WriteString("\n")
	}

	s.WriteString(wizDimStyle.Render("  ↑/↓ select • enter confirm • esc back • q quit"))

	return wizBorderStyle.Render(s.String())
}

func (m *wizModel) viewConfirm() string {
	var s strings.Builder

	lvl := wizLevels[m.levelIdx]

	s.WriteString(wizTitleStyle.Render("🌀 Confirm Setup"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("  Install level:  %s\n", wizSubStyle.Render(lvl.name)))
	s.WriteString(fmt.Sprintf("  Download size:  %s\n", lvl.size))
	s.WriteString(fmt.Sprintf("  Est. time:      %s\n", lvl.time))
	s.WriteString(fmt.Sprintf("  Models:         %s\n", lvl.models))
	s.WriteString(fmt.Sprintf("  Render preset:  %s\n", lvl.preset))

	if m.gpuInfo.Available {
		s.WriteString(fmt.Sprintf("  GPU:            %s (%d GB)\n", m.gpuInfo.GPUs[0].Name, m.vram))
	}

	s.WriteString("\n")
	s.WriteString(wizSubStyle.Render("  Install steps:"))
	s.WriteString("\n\n")

	phases := buildWizPhases(lvl.id)
	for i, ph := range phases {
		s.WriteString(fmt.Sprintf("    %d. %s\n", i+1, ph.name))
	}

	s.WriteString("\n")
	s.WriteString(wizDimStyle.Render("  enter to start • esc to go back • q to quit"))

	return wizBorderStyle.Render(s.String())
}

func (m *wizModel) viewInstalling() string {
	var s strings.Builder

	s.WriteString(wizTitleStyle.Render("🌀 Installing"))
	s.WriteString("\n\n")

	for _, ph := range m.phases {
		var icon string
		switch ph.status {
		case "done":
			icon = wizGoodStyle.Render("✓")
		case "running":
			icon = m.spinner.View()
		case "failed":
			icon = wizErrStyle.Render("✗")
		default:
			icon = wizDimStyle.Render("·")
		}

		name := wizDimStyle.Render(ph.name)
		if ph.status == "running" {
			name = wizSubStyle.Render(ph.name)
		} else if ph.status == "done" {
			name = wizGoodStyle.Render(ph.name)
		} else if ph.status == "failed" {
			name = wizErrStyle.Render(ph.name)
		}

		s.WriteString(fmt.Sprintf("  %s  %s", icon, name))
		if ph.detail != "" && ph.status != "pending" {
			s.WriteString(wizDimStyle.Render("  " + ph.detail))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(wizDimStyle.Render("  Installing... ctrl+c to abort"))

	return wizBorderStyle.Render(s.String())
}

func (m *wizModel) viewDone() string {
	var s strings.Builder

	if m.installErr != nil {
		s.WriteString(wizErrStyle.Render("✗ Setup failed"))
		s.WriteString("\n\n")
		s.WriteString(wizErrStyle.Render("  " + m.installErr.Error()))
		s.WriteString("\n\n")

		// Show phase status
		for _, ph := range m.phases {
			icon := wizDimStyle.Render("·")
			switch ph.status {
			case "done":
				icon = wizGoodStyle.Render("✓")
			case "failed":
				icon = wizErrStyle.Render("✗")
			}
			s.WriteString(fmt.Sprintf("  %s  %s\n", icon, ph.name))
		}

		s.WriteString("\n")
		s.WriteString(wizDimStyle.Render("  Try: anime wan fix"))
		s.WriteString("\n")
		s.WriteString(wizDimStyle.Render("  q to quit"))
	} else {
		s.WriteString(wizGoodStyle.Render("✓ Setup complete!"))
		s.WriteString("\n\n")

		for _, ph := range m.phases {
			s.WriteString(fmt.Sprintf("  %s  %s\n", wizGoodStyle.Render("✓"), ph.name))
		}

		s.WriteString("\n")
		s.WriteString(wizSubStyle.Render("  Ready to render."))
		s.WriteString("\n\n")
		s.WriteString(wizDimStyle.Render("  s / enter  Launch Comfort studio"))
		s.WriteString("\n")
		s.WriteString(wizDimStyle.Render("  q          Quit"))
	}

	return wizBorderStyle.Render(s.String())
}

// ─── phase execution ───

func buildWizPhases(level string) []wizPhaseState {
	phases := wanStudioPhases(level)
	var out []wizPhaseState
	for _, ph := range phases {
		if ph.id == "" {
			// Skip the "server running" pseudo-phase — wizard doesn't start the server
			continue
		}
		out = append(out, wizPhaseState{
			name:   ph.name,
			status: "pending",
		})
	}
	return out
}

func (m *wizModel) runPhase(idx int) tea.Cmd {
	level := wizLevels[m.levelIdx].id
	phases := wanStudioPhases(level)

	// Find the real phase matching this wizard index (skip id=="" phases)
	var realPhases []phase
	for _, ph := range phases {
		if ph.id != "" {
			realPhases = append(realPhases, ph)
		}
	}

	if idx >= len(realPhases) {
		return func() tea.Msg { return wizAllDoneMsg{} }
	}

	ph := realPhases[idx]

	return tea.Batch(
		func() tea.Msg { return wizPhaseStartMsg{idx: idx} },
		func() tea.Msg {
			// Check if already satisfied
			if ok, detail := ph.check(); ok {
				return wizPhaseDoneMsg{idx: idx, detail: "already installed — " + detail}
			}

			// Set WAN_INSTALL_LEVEL for wanmodels
			os.Setenv("WAN_INSTALL_LEVEL", level)

			var err error
			if ph.custom != nil {
				err = ph.custom(&setupOpts{yes: true, installLevel: level})
			} else if ph.id != "" {
				err = runInstallScript(ph.id)
			}

			if err != nil {
				return wizPhaseFailMsg{idx: idx, err: err}
			}

			// Verify
			if ok, detail := ph.check(); !ok {
				return wizPhaseFailMsg{idx: idx, err: fmt.Errorf("check failed after install: %s", detail)}
			}

			// Record snapshot
			saveWanInstallSnapshot(ph.id, ph.name, nil)

			return wizPhaseDoneMsg{idx: idx, detail: "installed"}
		},
	)
}
