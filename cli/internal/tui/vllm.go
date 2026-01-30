package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/theme"
)

// ============================================================================
// VLLM SETUP TUI
// ============================================================================

// Screen states
type vllmScreen int

const (
	vllmScreenWelcome vllmScreen = iota
	vllmScreenInstall
	vllmScreenModelSelect
	vllmScreenGPUConfig
	vllmScreenAdvanced
	vllmScreenSummary
)

// Model tabs
type modelTab int

const (
	tabCurated modelTab = iota
	tabHuggingFace
	tabCustom
)

// Pre-configured models with metadata
type CuratedModel struct {
	Shortcut    string
	HuggingFace string
	Size        string
	Category    string
	Description string
	MinGPUs     int
	MinVRAM     int // GB per GPU
}

var curatedModels = []CuratedModel{
	{"llama-70b", "meta-llama/Llama-3.3-70B-Instruct", "70B", "Llama", "Latest Llama 3.3, best quality", 4, 20},
	{"qwen-72b", "Qwen/Qwen2.5-72B-Instruct", "72B", "Qwen", "Powerful multilingual model", 4, 20},
	{"qwen-32b", "Qwen/Qwen2.5-32B-Instruct", "32B", "Qwen", "Balanced size and performance", 2, 20},
	{"qwen-14b", "Qwen/Qwen2.5-14B-Instruct", "14B", "Qwen", "Mid-size with great quality", 1, 24},
	{"qwen-7b", "Qwen/Qwen2.5-7B-Instruct", "7B", "Qwen", "Lightweight but capable", 1, 16},
	{"deepseek-r1", "deepseek-ai/DeepSeek-R1", "671B", "DeepSeek", "Reasoning-focused MoE model", 8, 80},
	{"deepseek-67b", "deepseek-ai/deepseek-llm-67b-chat", "67B", "DeepSeek", "Strong general purpose", 4, 20},
	{"mixtral", "mistralai/Mixtral-8x7B-Instruct-v0.1", "8x7B", "Mistral", "MoE architecture, fast inference", 2, 24},
	{"mistral-7b", "mistralai/Mistral-7B-Instruct-v0.3", "7B", "Mistral", "Compact and efficient", 1, 16},
	{"codellama", "codellama/CodeLlama-34b-Instruct-hf", "34B", "Code", "Specialized for coding", 2, 20},
	{"phi-4", "microsoft/phi-4", "14B", "Microsoft", "Small but mighty", 1, 16},
}

// HuggingFace model from API
type HFModel struct {
	ID        string `json:"id"`
	Downloads int    `json:"downloads"`
	Likes     int    `json:"likes"`
}

// System info for vLLM setup
type VLLMSystemInfo struct {
	VLLMInstalled    bool
	VLLMVersion      string
	PythonVersion    string
	CUDAVersion      string
	PyTorchInstalled bool
	PyTorchVersion   string
	PyTorchCUDA      bool
	GPUCount         int
	GPUNames         []string
	GPUMemory        []int // GB per GPU
	HasHFToken       bool
}

// Installation step
type InstallStep struct {
	Name    string
	Status  string // pending, running, done, failed
	Message string
}

// VLLMSetupModel is the main TUI model
type VLLMSetupModel struct {
	screen vllmScreen
	cursor int

	// System detection
	systemInfo     VLLMSystemInfo
	checkingSystem bool

	// Installation
	installSteps   []InstallStep
	installCursor  int
	installing     bool
	installOutput  []string

	// Model selection
	modelTab       modelTab
	curatedCursor  int
	hfModels       []HFModel
	hfCursor       int
	hfLoading      bool
	hfSearchInput  textinput.Model
	customInput    textinput.Model
	selectedModel  string
	selectedModelID string

	// GPU config
	tensorParallel int
	gpuMemory      float64
	gpuMemCursor   int // For slider

	// Advanced options
	quantization   string
	dtype          string
	maxModelLen    int
	swapSpace      int
	enableLora     bool
	loraRank       int
	advancedCursor int
	advancedInputs []textinput.Model
	advancedFocus  int

	// Server options
	port       int
	background bool

	// UI state
	width      int
	height     int
	spinner    spinner.Model
	starting   bool
	startError error
	quitting   bool

	// Result (returned to caller)
	Confirmed bool
}

// Result struct for returning configuration
type VLLMSetupResult struct {
	Model          string
	TensorParallel int
	GPUMemory      float64
	Quantization   string
	DType          string
	MaxModelLen    int
	SwapSpace      int
	EnableLora     bool
	LoraRank       int
	Port           int
	Background     bool
}

// Styles
var (
	vllmTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.SakuraPink).
		MarginBottom(1)

	vllmSelectedStyle = lipgloss.NewStyle().
		Foreground(theme.MintGreen).
		Bold(true)

	vllmUnselectedStyle = lipgloss.NewStyle().
		Foreground(theme.TextSecondary)

	vllmDimStyle = lipgloss.NewStyle().
		Foreground(theme.TextDim)

	vllmInfoStyle = lipgloss.NewStyle().
		Foreground(theme.ElectricBlue)

	vllmSuccessStyle = lipgloss.NewStyle().
		Foreground(theme.MintGreen)

	vllmErrorStyle = lipgloss.NewStyle().
		Foreground(theme.ActionRed)

	vllmWarningStyle = lipgloss.NewStyle().
		Foreground(theme.WarningYellow)

	vllmBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.NeonPurple).
		Padding(1, 2)

	vllmTabActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.SakuraPink).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.SakuraPink)

	vllmTabInactiveStyle = lipgloss.NewStyle().
		Foreground(theme.TextDim)

	vllmSliderFillStyle = lipgloss.NewStyle().
		Foreground(theme.SakuraPink)

	vllmSliderEmptyStyle = lipgloss.NewStyle().
		Foreground(theme.TextDim)
)

// NewVLLMSetupModel creates a new TUI model
func NewVLLMSetupModel() VLLMSetupModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.SakuraPink)

	// Create text inputs
	hfSearch := textinput.New()
	hfSearch.Placeholder = "Search HuggingFace models..."
	hfSearch.CharLimit = 100

	customInput := textinput.New()
	customInput.Placeholder = "e.g., meta-llama/Llama-3.3-70B-Instruct"
	customInput.CharLimit = 200

	// Advanced inputs
	advInputs := make([]textinput.Model, 3)
	advInputs[0] = textinput.New()
	advInputs[0].Placeholder = "8192"
	advInputs[0].CharLimit = 10

	advInputs[1] = textinput.New()
	advInputs[1].Placeholder = "0"
	advInputs[1].CharLimit = 5

	advInputs[2] = textinput.New()
	advInputs[2].Placeholder = "64"
	advInputs[2].CharLimit = 5

	return VLLMSetupModel{
		screen:         vllmScreenWelcome,
		checkingSystem: true,
		tensorParallel: 0, // Auto
		gpuMemory:      0.90,
		quantization:   "none",
		dtype:          "auto",
		maxModelLen:    0,
		swapSpace:      0,
		enableLora:     false,
		loraRank:       64,
		port:           8000,
		background:     false,
		spinner:        s,
		hfSearchInput:  hfSearch,
		customInput:    customInput,
		advancedInputs: advInputs,
		installSteps: []InstallStep{
			{Name: "Check Python version", Status: "pending"},
			{Name: "Check pip", Status: "pending"},
			{Name: "Check/Install PyTorch", Status: "pending"},
			{Name: "Install vLLM", Status: "pending"},
			{Name: "Verify installation", Status: "pending"},
		},
	}
}

// Messages
type vllmSysInfoMsg VLLMSystemInfo
type hfModelsMsg []HFModel
type installStepMsg struct {
	index   int
	status  string
	message string
}
type installDoneMsg struct{ err error }

func (m VLLMSetupModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkSystemInfo,
	)
}

func (m VLLMSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.screen == vllmScreenWelcome {
				m.quitting = true
				return m, tea.Quit
			}
		case "esc":
			return m.handleEsc()
		}

		// Screen-specific handling
		switch m.screen {
		case vllmScreenWelcome:
			return m.updateWelcome(msg)
		case vllmScreenInstall:
			return m.updateInstall(msg)
		case vllmScreenModelSelect:
			return m.updateModelSelect(msg)
		case vllmScreenGPUConfig:
			return m.updateGPUConfig(msg)
		case vllmScreenAdvanced:
			return m.updateAdvanced(msg)
		case vllmScreenSummary:
			return m.updateSummary(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case vllmSysInfoMsg:
		m.systemInfo = VLLMSystemInfo(msg)
		m.checkingSystem = false
		// Set default tensor parallel based on GPU count
		if m.systemInfo.GPUCount > 1 {
			m.tensorParallel = m.systemInfo.GPUCount
		} else {
			m.tensorParallel = 1
		}

	case hfModelsMsg:
		m.hfModels = msg
		m.hfLoading = false

	case installStepMsg:
		if msg.index < len(m.installSteps) {
			m.installSteps[msg.index].Status = msg.status
			m.installSteps[msg.index].Message = msg.message
		}

	case installDoneMsg:
		m.installing = false
		if msg.err == nil {
			// Refresh system info and proceed
			return m, checkSystemInfo
		}
	}

	// Update text inputs if focused
	if m.screen == vllmScreenModelSelect {
		if m.modelTab == tabHuggingFace {
			var cmd tea.Cmd
			m.hfSearchInput, cmd = m.hfSearchInput.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.modelTab == tabCustom {
			var cmd tea.Cmd
			m.customInput, cmd = m.customInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.screen == vllmScreenAdvanced && m.advancedFocus >= 0 {
		var cmd tea.Cmd
		m.advancedInputs[m.advancedFocus], cmd = m.advancedInputs[m.advancedFocus].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m VLLMSetupModel) handleEsc() (tea.Model, tea.Cmd) {
	switch m.screen {
	case vllmScreenInstall:
		m.screen = vllmScreenWelcome
	case vllmScreenModelSelect:
		if m.systemInfo.VLLMInstalled {
			m.screen = vllmScreenWelcome
		}
	case vllmScreenGPUConfig:
		m.screen = vllmScreenModelSelect
	case vllmScreenAdvanced:
		m.screen = vllmScreenGPUConfig
	case vllmScreenSummary:
		m.screen = vllmScreenAdvanced
	}
	m.cursor = 0
	return m, nil
}

// ============================================================================
// WELCOME SCREEN
// ============================================================================

func (m VLLMSetupModel) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		maxCursor := 1
		if !m.systemInfo.VLLMInstalled {
			maxCursor = 2
		}
		if m.cursor < maxCursor {
			m.cursor++
		}
	case "enter":
		if m.checkingSystem {
			return m, nil
		}
		if m.systemInfo.VLLMInstalled {
			// Options: Continue, Reinstall
			if m.cursor == 0 {
				m.screen = vllmScreenModelSelect
				m.cursor = 0
			} else {
				m.screen = vllmScreenInstall
				m.cursor = 0
				return m, m.startInstallation()
			}
		} else {
			// Options: Install, Skip (if possible), Quit
			if m.cursor == 0 {
				m.screen = vllmScreenInstall
				m.cursor = 0
				return m, m.startInstallation()
			} else if m.cursor == 1 {
				// Skip - just proceed (will likely fail)
				m.screen = vllmScreenModelSelect
				m.cursor = 0
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m VLLMSetupModel) viewWelcome() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("vLLM Setup Wizard"))
	s.WriteString("\n\n")

	if m.checkingSystem {
		s.WriteString(m.spinner.View() + " Checking system...\n")
		return s.String()
	}

	// System info box
	info := m.systemInfo
	infoBox := strings.Builder{}
	infoBox.WriteString(vllmInfoStyle.Render("System Information") + "\n\n")

	// vLLM status
	if info.VLLMInstalled {
		infoBox.WriteString(vllmSuccessStyle.Render("  vLLM: ") + info.VLLMVersion + "\n")
	} else {
		infoBox.WriteString(vllmErrorStyle.Render("  vLLM: ") + "Not installed\n")
	}

	// Python
	if info.PythonVersion != "" {
		infoBox.WriteString(vllmSuccessStyle.Render("  Python: ") + info.PythonVersion + "\n")
	} else {
		infoBox.WriteString(vllmErrorStyle.Render("  Python: ") + "Not found\n")
	}

	// PyTorch
	if info.PyTorchInstalled {
		cudaStatus := ""
		if info.PyTorchCUDA {
			cudaStatus = " (CUDA enabled)"
		} else {
			cudaStatus = vllmWarningStyle.Render(" (no CUDA)")
		}
		infoBox.WriteString(vllmSuccessStyle.Render("  PyTorch: ") + info.PyTorchVersion + cudaStatus + "\n")
	} else {
		infoBox.WriteString(vllmErrorStyle.Render("  PyTorch: ") + "Not installed (required for vLLM)\n")
	}

	// CUDA
	if info.CUDAVersion != "" {
		infoBox.WriteString(vllmSuccessStyle.Render("  CUDA: ") + info.CUDAVersion + "\n")
	} else {
		infoBox.WriteString(vllmWarningStyle.Render("  CUDA: ") + "Not detected\n")
	}

	// GPUs
	if info.GPUCount > 0 {
		infoBox.WriteString(vllmSuccessStyle.Render(fmt.Sprintf("  GPUs: %d detected", info.GPUCount)) + "\n")
		for i, name := range info.GPUNames {
			mem := ""
			if i < len(info.GPUMemory) {
				mem = fmt.Sprintf(" (%dGB)", info.GPUMemory[i])
			}
			infoBox.WriteString(vllmDimStyle.Render(fmt.Sprintf("    [%d] %s%s", i, name, mem)) + "\n")
		}
	} else {
		infoBox.WriteString(vllmWarningStyle.Render("  GPUs: None detected") + "\n")
	}

	// HF Token
	if info.HasHFToken {
		infoBox.WriteString(vllmSuccessStyle.Render("  HuggingFace: ") + "Authenticated\n")
	} else {
		infoBox.WriteString(vllmWarningStyle.Render("  HuggingFace: ") + "No token (some models may be unavailable)\n")
	}

	s.WriteString(vllmBoxStyle.Render(infoBox.String()))
	s.WriteString("\n\n")

	// Menu options
	var options []string
	if info.VLLMInstalled {
		options = []string{"Continue to model selection", "Reinstall vLLM"}
	} else {
		options = []string{"Install vLLM", "Skip installation (advanced)", "Quit"}
	}

	for i, opt := range options {
		cursor := "  "
		style := vllmUnselectedStyle
		if i == m.cursor {
			cursor = "> "
			style = vllmSelectedStyle
		}
		s.WriteString(cursor + style.Render(opt) + "\n")
	}

	s.WriteString("\n" + vllmDimStyle.Render("Press Enter to select, q to quit"))

	return s.String()
}

// ============================================================================
// INSTALLATION SCREEN
// ============================================================================

func (m VLLMSetupModel) updateInstall(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if !m.installing {
			// Installation done, proceed
			if m.systemInfo.VLLMInstalled {
				m.screen = vllmScreenModelSelect
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m VLLMSetupModel) startInstallation() tea.Cmd {
	m.installing = true
	return func() tea.Msg {
		// Step 1: Check Python
		tea.Println(installStepMsg{index: 0, status: "running", message: ""})
		cmd := exec.Command("python3", "--version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return installStepMsg{index: 0, status: "failed", message: "Python 3 not found. Run: anime install python"}
		}
		tea.Println(installStepMsg{index: 0, status: "done", message: strings.TrimSpace(string(output))})

		// Step 2: Check pip
		tea.Println(installStepMsg{index: 1, status: "running", message: ""})
		cmd = exec.Command("pip3", "--version")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return installStepMsg{index: 1, status: "failed", message: "pip3 not found. Run: anime install python"}
		}
		tea.Println(installStepMsg{index: 1, status: "done", message: strings.TrimSpace(string(output))})

		// Step 3: Check/Install PyTorch with CUDA
		tea.Println(installStepMsg{index: 2, status: "running", message: ""})
		cmd = exec.Command("python3", "-c", "import torch; assert torch.cuda.is_available(), 'CUDA not available'")
		_, err = cmd.CombinedOutput()
		if err != nil {
			// PyTorch not installed or CUDA not available - install it
			tea.Println(installStepMsg{index: 2, status: "running", message: "Installing PyTorch with CUDA..."})
			cmd = exec.Command("pip3", "install", "torch", "torchvision", "torchaudio",
				"--index-url", "https://download.pytorch.org/whl/cu126")
			output, err = cmd.CombinedOutput()
			if err != nil {
				return installStepMsg{index: 2, status: "failed", message: "Failed to install PyTorch. Run: anime install pytorch"}
			}
			// Verify PyTorch installation
			cmd = exec.Command("python3", "-c", "import torch; print(f'PyTorch {torch.__version__}')")
			output, err = cmd.CombinedOutput()
			if err != nil {
				return installStepMsg{index: 2, status: "failed", message: "PyTorch installed but import failed"}
			}
			tea.Println(installStepMsg{index: 2, status: "done", message: strings.TrimSpace(string(output)) + " (newly installed)"})
		} else {
			// PyTorch already installed with CUDA
			cmd = exec.Command("python3", "-c", "import torch; print(f'PyTorch {torch.__version__} (CUDA {torch.version.cuda})')")
			output, _ = cmd.CombinedOutput()
			tea.Println(installStepMsg{index: 2, status: "done", message: strings.TrimSpace(string(output))})
		}

		// Step 4: Install vLLM
		tea.Println(installStepMsg{index: 3, status: "running", message: ""})
		cmd = exec.Command("pip3", "install", "vllm")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return installStepMsg{index: 3, status: "failed", message: string(output)}
		}
		tea.Println(installStepMsg{index: 3, status: "done", message: "vLLM installed"})

		// Step 5: Verify installation
		tea.Println(installStepMsg{index: 4, status: "running", message: ""})
		cmd = exec.Command("python3", "-c", "import vllm; print(vllm.__version__)")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return installStepMsg{index: 4, status: "failed", message: string(output)}
		}
		tea.Println(installStepMsg{index: 4, status: "done", message: strings.TrimSpace(string(output))})

		return installDoneMsg{err: nil}
	}
}

func (m VLLMSetupModel) viewInstall() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("Installing vLLM"))
	s.WriteString("\n\n")

	for _, step := range m.installSteps {
		var icon string
		var style lipgloss.Style
		switch step.Status {
		case "pending":
			icon = "  "
			style = vllmDimStyle
		case "running":
			icon = m.spinner.View()
			style = vllmInfoStyle
		case "done":
			icon = vllmSuccessStyle.Render("")
			style = vllmSuccessStyle
		case "failed":
			icon = vllmErrorStyle.Render("")
			style = vllmErrorStyle
		}

		s.WriteString(fmt.Sprintf("%s %s", icon, style.Render(step.Name)))
		if step.Message != "" {
			s.WriteString(vllmDimStyle.Render(" - " + step.Message))
		}
		s.WriteString("\n")
	}

	if !m.installing {
		s.WriteString("\n" + vllmDimStyle.Render("Press Enter to continue, Esc to go back"))
	}

	return s.String()
}

// ============================================================================
// MODEL SELECTION SCREEN
// ============================================================================

func (m VLLMSetupModel) updateModelSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// Cycle tabs
		m.modelTab = (m.modelTab + 1) % 3
		if m.modelTab == tabHuggingFace && len(m.hfModels) == 0 && !m.hfLoading {
			return m, fetchHFModels("")
		}
		if m.modelTab == tabHuggingFace {
			m.hfSearchInput.Focus()
		} else if m.modelTab == tabCustom {
			m.customInput.Focus()
		} else {
			m.hfSearchInput.Blur()
			m.customInput.Blur()
		}
		m.cursor = 0
		return m, nil
	case "up", "k":
		if m.modelTab == tabCurated && m.curatedCursor > 0 {
			m.curatedCursor--
		} else if m.modelTab == tabHuggingFace && m.hfCursor > 0 {
			m.hfCursor--
		}
	case "down", "j":
		if m.modelTab == tabCurated && m.curatedCursor < len(curatedModels)-1 {
			m.curatedCursor++
		} else if m.modelTab == tabHuggingFace && m.hfCursor < len(m.hfModels)-1 {
			m.hfCursor++
		}
	case "enter":
		switch m.modelTab {
		case tabCurated:
			model := curatedModels[m.curatedCursor]
			m.selectedModel = model.Shortcut
			m.selectedModelID = model.HuggingFace
			m.screen = vllmScreenGPUConfig
			m.cursor = 0
		case tabHuggingFace:
			if len(m.hfModels) > 0 {
				m.selectedModel = m.hfModels[m.hfCursor].ID
				m.selectedModelID = m.hfModels[m.hfCursor].ID
				m.screen = vllmScreenGPUConfig
				m.cursor = 0
			}
		case tabCustom:
			if m.customInput.Value() != "" {
				m.selectedModel = m.customInput.Value()
				m.selectedModelID = m.customInput.Value()
				m.screen = vllmScreenGPUConfig
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m VLLMSetupModel) viewModelSelect() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("Select Model"))
	s.WriteString("\n\n")

	// Tabs
	tabs := []string{"Curated", "HuggingFace", "Custom"}
	tabLine := ""
	for i, tab := range tabs {
		style := vllmTabInactiveStyle
		if modelTab(i) == m.modelTab {
			style = vllmTabActiveStyle
		}
		tabLine += style.Render(" "+tab+" ") + "  "
	}
	s.WriteString(tabLine + "\n\n")

	switch m.modelTab {
	case tabCurated:
		s.WriteString(m.viewCuratedModels())
	case tabHuggingFace:
		s.WriteString(m.viewHFModels())
	case tabCustom:
		s.WriteString(m.viewCustomModel())
	}

	s.WriteString("\n" + vllmDimStyle.Render("Tab: switch tabs | Enter: select | Esc: back"))

	return s.String()
}

func (m VLLMSetupModel) viewCuratedModels() string {
	var s strings.Builder

	// Group by category
	categories := make(map[string][]CuratedModel)
	for _, model := range curatedModels {
		categories[model.Category] = append(categories[model.Category], model)
	}

	// Sort categories
	catOrder := []string{"Llama", "Qwen", "DeepSeek", "Mistral", "Code", "Microsoft"}

	idx := 0
	for _, cat := range catOrder {
		models, ok := categories[cat]
		if !ok {
			continue
		}

		s.WriteString(vllmInfoStyle.Render(cat) + "\n")
		for _, model := range models {
			cursor := "  "
			style := vllmUnselectedStyle
			if idx == m.curatedCursor {
				cursor = "> "
				style = vllmSelectedStyle
			}

			line := fmt.Sprintf("%s%-12s %s", cursor, model.Shortcut, vllmDimStyle.Render(model.Size))
			s.WriteString(style.Render(line))

			if idx == m.curatedCursor {
				// Show description for selected
				s.WriteString("\n    " + vllmDimStyle.Render(model.Description))
				s.WriteString(fmt.Sprintf("\n    " + vllmDimStyle.Render("Min: %d GPU(s), %dGB VRAM each"), model.MinGPUs, model.MinVRAM))
			}
			s.WriteString("\n")
			idx++
		}
		s.WriteString("\n")
	}

	return s.String()
}

func (m VLLMSetupModel) viewHFModels() string {
	var s strings.Builder

	s.WriteString(m.hfSearchInput.View() + "\n\n")

	if m.hfLoading {
		s.WriteString(m.spinner.View() + " Loading models from HuggingFace...\n")
		return s.String()
	}

	if len(m.hfModels) == 0 {
		s.WriteString(vllmDimStyle.Render("No models loaded. Press Enter to search.") + "\n")
		return s.String()
	}

	maxShow := 10
	for i, model := range m.hfModels {
		if i >= maxShow {
			s.WriteString(vllmDimStyle.Render(fmt.Sprintf("  ... and %d more", len(m.hfModels)-maxShow)) + "\n")
			break
		}

		cursor := "  "
		style := vllmUnselectedStyle
		if i == m.hfCursor {
			cursor = "> "
			style = vllmSelectedStyle
		}

		downloads := formatNumber(model.Downloads)
		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, style.Render(model.ID), vllmDimStyle.Render(downloads+" downloads")))
	}

	return s.String()
}

func (m VLLMSetupModel) viewCustomModel() string {
	var s strings.Builder

	s.WriteString("Enter HuggingFace model ID:\n\n")
	s.WriteString(m.customInput.View() + "\n\n")
	s.WriteString(vllmDimStyle.Render("Example: meta-llama/Llama-3.3-70B-Instruct") + "\n")

	return s.String()
}

// ============================================================================
// GPU CONFIGURATION SCREEN
// ============================================================================

func (m VLLMSetupModel) updateGPUConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 2 {
			m.cursor++
		}
	case "left", "h":
		switch m.cursor {
		case 0: // Tensor parallel
			if m.tensorParallel > 1 {
				m.tensorParallel--
			}
		case 1: // GPU memory
			if m.gpuMemory > 0.50 {
				m.gpuMemory -= 0.05
			}
		}
	case "right", "l":
		switch m.cursor {
		case 0: // Tensor parallel
			if m.tensorParallel < m.systemInfo.GPUCount || m.tensorParallel < 8 {
				m.tensorParallel++
			}
		case 1: // GPU memory
			if m.gpuMemory < 0.99 {
				m.gpuMemory += 0.05
			}
		}
	case "enter":
		if m.cursor == 2 {
			m.screen = vllmScreenAdvanced
			m.cursor = 0
		}
	}
	return m, nil
}

func (m VLLMSetupModel) viewGPUConfig() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("GPU Configuration"))
	s.WriteString("\n\n")

	// Selected model
	s.WriteString(vllmInfoStyle.Render("Model: ") + m.selectedModel + "\n")
	s.WriteString(vllmDimStyle.Render(m.selectedModelID) + "\n\n")

	// Tensor parallelism
	cursor := "  "
	style := vllmUnselectedStyle
	if m.cursor == 0 {
		cursor = "> "
		style = vllmSelectedStyle
	}
	tpText := fmt.Sprintf("%d GPU(s)", m.tensorParallel)
	if m.tensorParallel == 0 {
		tpText = "Auto"
	}
	s.WriteString(fmt.Sprintf("%sTensor Parallelism: %s\n", cursor, style.Render(tpText)))
	if m.cursor == 0 {
		s.WriteString("    " + vllmDimStyle.Render("Use left/right arrows to adjust") + "\n")
		s.WriteString("    " + m.renderTPSlider() + "\n")
	}
	s.WriteString("\n")

	// GPU memory utilization
	cursor = "  "
	style = vllmUnselectedStyle
	if m.cursor == 1 {
		cursor = "> "
		style = vllmSelectedStyle
	}
	memText := fmt.Sprintf("%.0f%%", m.gpuMemory*100)
	s.WriteString(fmt.Sprintf("%sGPU Memory Usage: %s\n", cursor, style.Render(memText)))
	if m.cursor == 1 {
		s.WriteString("    " + vllmDimStyle.Render("Use left/right arrows to adjust (50%-99%)") + "\n")
		s.WriteString("    " + m.renderMemSlider() + "\n")
	}
	s.WriteString("\n")

	// Continue button
	cursor = "  "
	style = vllmUnselectedStyle
	if m.cursor == 2 {
		cursor = "> "
		style = vllmSelectedStyle
	}
	s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render("Continue to advanced options")))

	s.WriteString("\n" + vllmDimStyle.Render("Enter: select | Esc: back"))

	return s.String()
}

func (m VLLMSetupModel) renderTPSlider() string {
	maxGPU := m.systemInfo.GPUCount
	if maxGPU < 8 {
		maxGPU = 8
	}

	var slider strings.Builder
	slider.WriteString("[")
	for i := 1; i <= maxGPU; i++ {
		if i == m.tensorParallel {
			slider.WriteString(vllmSliderFillStyle.Render(""))
		} else if i <= m.systemInfo.GPUCount {
			slider.WriteString(vllmSliderEmptyStyle.Render(""))
		} else {
			slider.WriteString(vllmDimStyle.Render(""))
		}
	}
	slider.WriteString("]")
	return slider.String()
}

func (m VLLMSetupModel) renderMemSlider() string {
	width := 20
	filled := int((m.gpuMemory - 0.50) / 0.50 * float64(width))

	var slider strings.Builder
	slider.WriteString("[")
	for i := 0; i < width; i++ {
		if i < filled {
			slider.WriteString(vllmSliderFillStyle.Render(""))
		} else {
			slider.WriteString(vllmSliderEmptyStyle.Render(""))
		}
	}
	slider.WriteString("]")
	return slider.String()
}

// ============================================================================
// ADVANCED OPTIONS SCREEN
// ============================================================================

func (m VLLMSetupModel) updateAdvanced(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := []string{"none", "awq", "gptq", "fp8"}
	dtypes := []string{"auto", "float16", "bfloat16", "float32"}

	switch msg.String() {
	case "up", "k":
		if m.advancedCursor > 0 {
			m.advancedCursor--
			m.advancedFocus = -1
		}
	case "down", "j":
		if m.advancedCursor < 7 {
			m.advancedCursor++
			m.advancedFocus = -1
		}
	case "left", "h":
		switch m.advancedCursor {
		case 0: // Quantization
			idx := indexOf(options, m.quantization)
			if idx > 0 {
				m.quantization = options[idx-1]
			}
		case 1: // DType
			idx := indexOf(dtypes, m.dtype)
			if idx > 0 {
				m.dtype = dtypes[idx-1]
			}
		}
	case "right", "l":
		switch m.advancedCursor {
		case 0: // Quantization
			idx := indexOf(options, m.quantization)
			if idx < len(options)-1 {
				m.quantization = options[idx+1]
			}
		case 1: // DType
			idx := indexOf(dtypes, m.dtype)
			if idx < len(dtypes)-1 {
				m.dtype = dtypes[idx+1]
			}
		}
	case "tab":
		// Focus text input if on max len, swap, or lora rank
		switch m.advancedCursor {
		case 2:
			m.advancedFocus = 0
			m.advancedInputs[0].Focus()
		case 3:
			m.advancedFocus = 1
			m.advancedInputs[1].Focus()
		case 5:
			m.advancedFocus = 2
			m.advancedInputs[2].Focus()
		}
	case " ":
		switch m.advancedCursor {
		case 4: // LoRA toggle
			m.enableLora = !m.enableLora
		case 6: // Background toggle
			m.background = !m.background
		}
	case "enter":
		if m.advancedCursor == 7 {
			// Parse inputs
			if v, err := strconv.Atoi(m.advancedInputs[0].Value()); err == nil {
				m.maxModelLen = v
			}
			if v, err := strconv.Atoi(m.advancedInputs[1].Value()); err == nil {
				m.swapSpace = v
			}
			if v, err := strconv.Atoi(m.advancedInputs[2].Value()); err == nil {
				m.loraRank = v
			}
			m.screen = vllmScreenSummary
			m.cursor = 0
		}
	}
	return m, nil
}

func (m VLLMSetupModel) viewAdvanced() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("Advanced Options"))
	s.WriteString("\n\n")

	options := []struct {
		name  string
		value string
		hint  string
	}{
		{"Quantization", m.quantization, "none, awq, gptq, fp8"},
		{"Data Type", m.dtype, "auto, float16, bfloat16, float32"},
		{"Max Context Length", m.advancedInputs[0].View(), "0 = model default"},
		{"Swap Space (GB)", m.advancedInputs[1].View(), "CPU KV cache offload"},
		{"Enable LoRA", boolToCheck(m.enableLora), "Space to toggle"},
		{"LoRA Rank", m.advancedInputs[2].View(), "Max adapter rank"},
		{"Run in Background", boolToCheck(m.background), "Space to toggle"},
		{"Continue to summary", "", ""},
	}

	for i, opt := range options {
		cursor := "  "
		style := vllmUnselectedStyle
		if i == m.advancedCursor {
			cursor = "> "
			style = vllmSelectedStyle
		}

		if opt.value != "" {
			s.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, style.Render(opt.name), opt.value))
		} else {
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(opt.name)))
		}

		if i == m.advancedCursor && opt.hint != "" {
			s.WriteString("    " + vllmDimStyle.Render(opt.hint) + "\n")
		}
	}

	s.WriteString("\n" + vllmDimStyle.Render("Enter: continue | Space: toggle | Left/Right: change | Esc: back"))

	return s.String()
}

// ============================================================================
// SUMMARY SCREEN
// ============================================================================

func (m VLLMSetupModel) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 2 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0: // Start
			m.Confirmed = true
			return m, tea.Quit
		case 1: // Edit
			m.screen = vllmScreenModelSelect
			m.cursor = 0
		case 2: // Cancel
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m VLLMSetupModel) viewSummary() string {
	var s strings.Builder

	s.WriteString(vllmTitleStyle.Render("Review Configuration"))
	s.WriteString("\n\n")

	// Configuration summary
	config := []struct {
		label string
		value string
	}{
		{"Model", m.selectedModel},
		{"HuggingFace ID", m.selectedModelID},
		{"Tensor Parallel", fmt.Sprintf("%d GPU(s)", m.tensorParallel)},
		{"GPU Memory", fmt.Sprintf("%.0f%%", m.gpuMemory*100)},
		{"Quantization", m.quantization},
		{"Data Type", m.dtype},
		{"Max Context", formatIntOrDefault(m.maxModelLen, "default")},
		{"Swap Space", formatIntOrDefault(m.swapSpace, "disabled") + " GB"},
		{"LoRA", boolToEnabled(m.enableLora)},
		{"Background", boolToYesNo(m.background)},
		{"Port", fmt.Sprintf("%d", m.port)},
	}

	boxContent := strings.Builder{}
	for _, c := range config {
		boxContent.WriteString(fmt.Sprintf("%s: %s\n", vllmInfoStyle.Render(c.label), c.value))
	}
	s.WriteString(vllmBoxStyle.Render(boxContent.String()))
	s.WriteString("\n\n")

	// Actions
	actions := []string{"Start vLLM Server", "Edit Configuration", "Cancel"}
	for i, action := range actions {
		cursor := "  "
		style := vllmUnselectedStyle
		if i == m.cursor {
			cursor = "> "
			style = vllmSelectedStyle
		}
		if i == 0 {
			style = vllmSuccessStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(action)))
	}

	s.WriteString("\n" + vllmDimStyle.Render("Enter: select | Esc: back"))

	return s.String()
}

// ============================================================================
// VIEW ROUTER
// ============================================================================

func (m VLLMSetupModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case vllmScreenWelcome:
		return m.viewWelcome()
	case vllmScreenInstall:
		return m.viewInstall()
	case vllmScreenModelSelect:
		return m.viewModelSelect()
	case vllmScreenGPUConfig:
		return m.viewGPUConfig()
	case vllmScreenAdvanced:
		return m.viewAdvanced()
	case vllmScreenSummary:
		return m.viewSummary()
	default:
		return ""
	}
}

// GetResult returns the configuration result
func (m VLLMSetupModel) GetResult() VLLMSetupResult {
	return VLLMSetupResult{
		Model:          m.selectedModelID,
		TensorParallel: m.tensorParallel,
		GPUMemory:      m.gpuMemory,
		Quantization:   m.quantization,
		DType:          m.dtype,
		MaxModelLen:    m.maxModelLen,
		SwapSpace:      m.swapSpace,
		EnableLora:     m.enableLora,
		LoraRank:       m.loraRank,
		Port:           m.port,
		Background:     m.background,
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func checkSystemInfo() tea.Msg {
	info := VLLMSystemInfo{}

	// Check vLLM
	if out, err := exec.Command("vllm", "--version").CombinedOutput(); err == nil {
		info.VLLMInstalled = true
		info.VLLMVersion = strings.TrimSpace(string(out))
	} else if out, err := exec.Command("python3", "-c", "import vllm; print(vllm.__version__)").CombinedOutput(); err == nil {
		info.VLLMInstalled = true
		info.VLLMVersion = strings.TrimSpace(string(out))
	}

	// Check Python
	if out, err := exec.Command("python3", "--version").CombinedOutput(); err == nil {
		info.PythonVersion = strings.TrimSpace(strings.TrimPrefix(string(out), "Python "))
	}

	// Check PyTorch
	if out, err := exec.Command("python3", "-c", "import torch; print(torch.__version__)").CombinedOutput(); err == nil {
		info.PyTorchInstalled = true
		info.PyTorchVersion = strings.TrimSpace(string(out))
		// Check if CUDA is available in PyTorch
		if _, err := exec.Command("python3", "-c", "import torch; assert torch.cuda.is_available()").CombinedOutput(); err == nil {
			info.PyTorchCUDA = true
		}
	}

	// Check CUDA
	if out, err := exec.Command("nvcc", "--version").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "release") {
				parts := strings.Split(line, "release ")
				if len(parts) > 1 {
					info.CUDAVersion = strings.Split(parts[1], ",")[0]
				}
			}
		}
	}

	// Check GPUs (using centralized, cached GPU detection)
	gpuInfo := gpu.GetSystemInfo()
	info.GPUCount = gpuInfo.Count
	for _, g := range gpuInfo.GPUs {
		info.GPUNames = append(info.GPUNames, g.Name)
		info.GPUMemory = append(info.GPUMemory, g.VRAM)
	}

	// Check HF token
	if os.Getenv("HF_TOKEN") != "" || os.Getenv("HUGGING_FACE_HUB_TOKEN") != "" {
		info.HasHFToken = true
	} else if home, err := os.UserHomeDir(); err == nil {
		tokenPath := home + "/.cache/huggingface/token"
		if _, err := os.Stat(tokenPath); err == nil {
			info.HasHFToken = true
		}
	}

	return vllmSysInfoMsg(info)
}

func fetchHFModels(query string) tea.Cmd {
	return func() tea.Msg {
		url := "https://huggingface.co/api/models?filter=text-generation&sort=downloads&limit=50"
		if query != "" {
			url += "&search=" + query
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return hfModelsMsg{}
		}
		defer resp.Body.Close()

		var models []HFModel
		if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
			return hfModelsMsg{}
		}

		// Sort by downloads
		sort.Slice(models, func(i, j int) bool {
			return models[i].Downloads > models[j].Downloads
		})

		return hfModelsMsg(models)
	}
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return 0
}

func boolToCheck(b bool) string {
	if b {
		return vllmSuccessStyle.Render("[x]")
	}
	return "[ ]"
}

func boolToEnabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func formatIntOrDefault(v int, def string) string {
	if v == 0 {
		return def
	}
	return fmt.Sprintf("%d", v)
}

// RunVLLMSetup launches the TUI and returns the result
func RunVLLMSetup() (*VLLMSetupResult, error) {
	model := NewVLLMSetupModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := finalModel.(VLLMSetupModel)
	if !m.Confirmed {
		return nil, nil // User cancelled
	}

	result := m.GetResult()
	return &result, nil
}
