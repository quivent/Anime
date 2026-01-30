package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
)

// Workflow TUI Styles
var (
	wfTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF69B4")).
			MarginBottom(1)

	wfTableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF69B4")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1)

	wfTableRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			Padding(0, 1)

	wfTableRowSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#444444")).
				Bold(true).
				Padding(0, 1)

	wfTableRowActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00")).
				Padding(0, 1)

	wfOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	wfOptionSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00BFFF")).
				Bold(true)

	wfOptionActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00")).
				Bold(true)

	wfHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)
)

// Screen states
type workflowView int

const (
	viewTable workflowView = iota
	viewCreate
	viewEdit
)

// Column widths
const (
	colName   = 20
	colServer = 12
	colModels = 8
	colGPUs   = 12
	colStatus = 10
)

// WorkflowModel is the TUI model
type WorkflowModel struct {
	config    *config.Config
	view      workflowView
	workflows []config.WorkflowProfile

	// Table cursor
	cursor int
	scroll int

	// Edit state
	editing      *config.WorkflowProfile
	editField    int // Which field is being edited
	editOption   int // Which option within field is selected
	isNewWorkflow bool

	// Terminal size
	width  int
	height int

	quitting bool
}

// Edit fields
const (
	fieldName = iota
	fieldServer
	fieldModels
	fieldGPU
	fieldOptimizations
	fieldAutoLoad
	fieldSave
	fieldCount
)

// Server options
var serverOptions = []struct {
	value config.LLMServerType
	label string
	desc  string
}{
	{config.ServerOllama, "Ollama", "Local, easy setup"},
	{config.ServerVLLM, "vLLM", "High throughput"},
	{config.ServerTensorRT, "TensorRT-LLM", "NVIDIA optimized"},
	{config.ServerLlamaCpp, "llama.cpp", "CPU/GPU hybrid"},
	{config.ServerExllamaV2, "ExLlamaV2", "Fast quantized"},
}

// GPU presets
var gpuPresets = []struct {
	label   string
	gpus    int
	gpuType string
	vram    int
}{
	{"None", 0, "", 0},
	{"1x H100 (80GB)", 1, "H100", 80},
	{"2x H100 (160GB)", 2, "H100", 80},
	{"4x H100 (320GB)", 4, "H100", 80},
	{"8x H100 (640GB)", 8, "H100", 80},
	{"1x GH200 (96GB)", 1, "GH200", 96},
	{"8x B200 (1.5TB)", 8, "B200", 192},
	{"1x A100 (80GB)", 1, "A100", 80},
	{"4x A100 (320GB)", 4, "A100", 80},
	{"8x A100 (640GB)", 8, "A100", 80},
}

// Workflow name presets
var namePresets = []string{
	"default",
	"inference",
	"training",
	"dev",
	"production",
	"substrate",
	"embedding",
	"coding",
	"reasoning",
}

// NewWorkflowModel creates a new workflow TUI
func NewWorkflowModel() (WorkflowModel, error) {
	cfg, err := config.Load()
	if err != nil {
		return WorkflowModel{}, err
	}

	return WorkflowModel{
		config:    cfg,
		view:      viewTable,
		workflows: cfg.Workflows,
	}, nil
}

// NewWorkflowModelWithCreate starts in create mode
func NewWorkflowModelWithCreate(name string) (WorkflowModel, error) {
	m, err := NewWorkflowModel()
	if err != nil {
		return m, err
	}

	m.view = viewCreate
	m.isNewWorkflow = true
	m.editing = &config.WorkflowProfile{
		Name:   name,
		Server: config.ServerOllama,
	}
	m.editField = fieldName
	m.editOption = 0

	// Find name in presets
	for i, n := range namePresets {
		if n == name {
			m.editOption = i
			break
		}
	}

	return m, nil
}

func (m WorkflowModel) Init() tea.Cmd {
	return nil
}

func (m WorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.view == viewTable {
				m.quitting = true
				return m, tea.Quit
			}
		case "esc":
			if m.view == viewEdit || m.view == viewCreate {
				m.view = viewTable
				m.editing = nil
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.view {
	case viewTable:
		return m.updateTable(msg)
	case viewCreate, viewEdit:
		return m.updateEdit(msg)
	}

	return m, nil
}

func (m WorkflowModel) updateTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxIdx := len(m.workflows) // +1 for "New" row handled by bound
			if m.cursor < maxIdx {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor == len(m.workflows) {
				// Create new
				m.view = viewCreate
				m.isNewWorkflow = true
				m.editing = &config.WorkflowProfile{
					Name:   namePresets[0],
					Server: config.ServerOllama,
				}
				m.editField = fieldName
				m.editOption = 0
			} else if m.cursor < len(m.workflows) {
				// Edit existing
				w := m.workflows[m.cursor]
				m.editing = &w
				m.view = viewEdit
				m.isNewWorkflow = false
				m.editField = fieldServer
				m.editOption = m.serverToIndex(w.Server)
			}
		case "u":
			// Use/activate
			if m.cursor < len(m.workflows) {
				m.config.ActiveWorkflow = m.workflows[m.cursor].Name
				m.config.Save()
			}
		case "d":
			// Delete
			if m.cursor < len(m.workflows) {
				name := m.workflows[m.cursor].Name
				m.config.DeleteWorkflow(name)
				m.config.Save()
				m.workflows = m.config.Workflows
				if m.cursor >= len(m.workflows) && m.cursor > 0 {
					m.cursor--
				}
			}
		case "s":
			// Start workflow
			if m.cursor < len(m.workflows) {
				m.config.ActiveWorkflow = m.workflows[m.cursor].Name
				m.config.Save()
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m WorkflowModel) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.editField > 0 {
				m.editField--
				m.editOption = 0
				m.syncOptionFromEditing()
			}
		case "down", "j":
			if m.editField < fieldCount-1 {
				m.editField++
				m.editOption = 0
				m.syncOptionFromEditing()
			}
		case "left", "h":
			if m.editOption > 0 {
				m.editOption--
				m.applyOption()
			}
		case "right", "l":
			m.editOption++
			m.clampOption()
			m.applyOption()
		case "tab":
			m.editField = (m.editField + 1) % fieldCount
			m.editOption = 0
			m.syncOptionFromEditing()
		case "enter", " ":
			if m.editField == fieldSave {
				m.saveWorkflow()
				m.view = viewTable
				m.workflows = m.config.Workflows
				return m, nil
			}
			// For other fields, cycle through options
			m.editOption++
			m.clampOption()
			m.applyOption()
		case "ctrl+s":
			m.saveWorkflow()
			m.view = viewTable
			m.workflows = m.config.Workflows
			return m, nil
		}
	}
	return m, nil
}

func (m *WorkflowModel) syncOptionFromEditing() {
	if m.editing == nil {
		return
	}
	switch m.editField {
	case fieldName:
		for i, n := range namePresets {
			if n == m.editing.Name {
				m.editOption = i
				return
			}
		}
	case fieldServer:
		m.editOption = m.serverToIndex(m.editing.Server)
	case fieldGPU:
		for i, p := range gpuPresets {
			if p.gpus == m.editing.GPUConfig.TotalGPUs && p.gpuType == m.editing.GPUConfig.GPUType {
				m.editOption = i
				return
			}
		}
	case fieldAutoLoad:
		if m.editing.AutoLoad {
			m.editOption = 1
		} else {
			m.editOption = 0
		}
	}
}

func (m *WorkflowModel) clampOption() {
	max := 0
	switch m.editField {
	case fieldName:
		max = len(namePresets) - 1
	case fieldServer:
		max = len(serverOptions) - 1
	case fieldModels:
		max = len(m.getLLMModels()) - 1
	case fieldGPU:
		max = len(gpuPresets) - 1
	case fieldOptimizations:
		max = 5 // 6 options (0-5)
	case fieldAutoLoad:
		max = 1
	}
	if m.editOption > max {
		m.editOption = 0
	}
	if m.editOption < 0 {
		m.editOption = max
	}
}

func (m *WorkflowModel) applyOption() {
	if m.editing == nil {
		return
	}
	switch m.editField {
	case fieldName:
		if m.editOption < len(namePresets) {
			m.editing.Name = namePresets[m.editOption]
		}
	case fieldServer:
		if m.editOption < len(serverOptions) {
			m.editing.Server = serverOptions[m.editOption].value
		}
	case fieldModels:
		models := m.getLLMModels()
		if m.editOption < len(models) {
			modelID := models[m.editOption].ID
			// Toggle model selection
			found := false
			for i, mod := range m.editing.Models {
				if mod.ID == modelID {
					m.editing.Models = append(m.editing.Models[:i], m.editing.Models[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				m.editing.Models = append(m.editing.Models, config.ModelDeployment{
					ID:      modelID,
					Enabled: true,
				})
			}
		}
	case fieldGPU:
		if m.editOption < len(gpuPresets) {
			p := gpuPresets[m.editOption]
			m.editing.GPUConfig.TotalGPUs = p.gpus
			m.editing.GPUConfig.GPUType = p.gpuType
			m.editing.GPUConfig.GPUMemoryGB = p.vram
		}
	case fieldOptimizations:
		switch m.editOption {
		case 0:
			m.editing.Optimizations.FlashAttention = !m.editing.Optimizations.FlashAttention
		case 1:
			m.editing.Optimizations.PagedAttention = !m.editing.Optimizations.PagedAttention
		case 2:
			m.editing.Optimizations.SpeculativeDecoding = !m.editing.Optimizations.SpeculativeDecoding
		case 3:
			m.editing.Optimizations.ContinuousBatching = !m.editing.Optimizations.ContinuousBatching
		case 4:
			m.editing.Optimizations.ChunkedPrefill = !m.editing.Optimizations.ChunkedPrefill
		case 5:
			m.editing.Optimizations.PrefixCaching = !m.editing.Optimizations.PrefixCaching
		}
	case fieldAutoLoad:
		m.editing.AutoLoad = m.editOption == 1
	}
}

func (m *WorkflowModel) saveWorkflow() {
	if m.editing == nil {
		return
	}

	if m.isNewWorkflow {
		m.config.AddWorkflow(*m.editing)
	} else {
		// Update existing
		for i, w := range m.config.Workflows {
			if w.Name == m.editing.Name {
				m.config.Workflows[i] = *m.editing
				break
			}
		}
	}
	m.config.Save()
}

func (m WorkflowModel) serverToIndex(s config.LLMServerType) int {
	for i, opt := range serverOptions {
		if opt.value == s {
			return i
		}
	}
	return 0
}

func (m WorkflowModel) getLLMModels() []config.Module {
	var models []config.Module
	for _, mod := range config.AvailableModules {
		if strings.HasPrefix(mod.Category, "LLM") {
			models = append(models, mod)
		}
	}
	return models
}

func (m WorkflowModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case viewTable:
		return m.viewTable()
	case viewCreate, viewEdit:
		return m.viewEdit()
	}
	return ""
}

func (m WorkflowModel) viewTable() string {
	var b strings.Builder

	b.WriteString(wfTitleStyle.Render("⚡ Workflow Manager"))
	b.WriteString("\n\n")

	// Table header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		colName, "NAME",
		colServer, "SERVER",
		colModels, "MODELS",
		colGPUs, "GPUs",
		colStatus, "STATUS")
	b.WriteString(wfTableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Separator
	b.WriteString(strings.Repeat("─", colName+colServer+colModels+colGPUs+colStatus+4))
	b.WriteString("\n")

	// Rows
	for i, w := range m.workflows {
		name := truncate(w.Name, colName-1)
		server := truncate(string(w.Server), colServer-1)
		models := fmt.Sprintf("%d", len(w.Models))
		gpus := "—"
		if w.GPUConfig.TotalGPUs > 0 {
			gpus = fmt.Sprintf("%dx %s", w.GPUConfig.TotalGPUs, w.GPUConfig.GPUType)
		}
		status := "ready"
		if w.Name == m.config.ActiveWorkflow {
			status = "★ active"
		}

		row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
			colName, name,
			colServer, server,
			colModels, models,
			colGPUs, truncate(gpus, colGPUs-1),
			colStatus, status)

		if i == m.cursor {
			b.WriteString(wfTableRowSelectedStyle.Render(row))
		} else if w.Name == m.config.ActiveWorkflow {
			b.WriteString(wfTableRowActiveStyle.Render(row))
		} else {
			b.WriteString(wfTableRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// New workflow row
	newRow := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		colName, "+ New Workflow",
		colServer, "",
		colModels, "",
		colGPUs, "",
		colStatus, "")
	if m.cursor == len(m.workflows) {
		b.WriteString(wfTableRowSelectedStyle.Render(newRow))
	} else {
		b.WriteString(wfOptionStyle.Render(newRow))
	}
	b.WriteString("\n")

	// Help
	b.WriteString("\n")
	b.WriteString(wfHelpStyle.Render("↑↓ navigate  enter edit  u activate  s start  d delete  q quit"))

	return b.String()
}

func (m WorkflowModel) viewEdit() string {
	var b strings.Builder

	title := "Edit Workflow"
	if m.isNewWorkflow {
		title = "Create Workflow"
	}
	if m.editing != nil {
		title = fmt.Sprintf("%s: %s", title, m.editing.Name)
	}
	b.WriteString(wfTitleStyle.Render(title))
	b.WriteString("\n\n")

	// Field rows
	fields := []struct {
		label   string
		options func() string
	}{
		{"Name", m.renderNameOptions},
		{"Server", m.renderServerOptions},
		{"Models", m.renderModelOptions},
		{"GPU", m.renderGPUOptions},
		{"Optimizations", m.renderOptOptions},
		{"Auto-load", m.renderAutoLoadOptions},
		{"", m.renderSaveButton},
	}

	for i, f := range fields {
		if f.label != "" {
			label := f.label
			if i == m.editField {
				label = wfOptionSelectedStyle.Render("▶ " + label)
			} else {
				label = wfOptionStyle.Render("  " + label)
			}
			b.WriteString(fmt.Sprintf("%-16s", label))
		}
		b.WriteString(f.options())
		b.WriteString("\n")
		if i == fieldModels || i == fieldOptimizations {
			b.WriteString("\n") // Extra space after multi-line sections
		}
	}

	// Help
	b.WriteString("\n")
	b.WriteString(wfHelpStyle.Render("↑↓ field  ←→/space cycle  enter select  ctrl+s save  esc back"))

	return b.String()
}

func (m WorkflowModel) renderNameOptions() string {
	var parts []string
	for i, name := range namePresets {
		style := wfOptionStyle
		if m.editing != nil && m.editing.Name == name {
			style = wfOptionActiveStyle
		}
		if m.editField == fieldName && i == m.editOption {
			style = wfOptionSelectedStyle
		}
		parts = append(parts, style.Render(name))
	}
	return strings.Join(parts, "  ")
}

func (m WorkflowModel) renderServerOptions() string {
	var parts []string
	for i, opt := range serverOptions {
		style := wfOptionStyle
		if m.editing != nil && m.editing.Server == opt.value {
			style = wfOptionActiveStyle
		}
		if m.editField == fieldServer && i == m.editOption {
			style = wfOptionSelectedStyle
		}
		parts = append(parts, style.Render(opt.label))
	}
	return strings.Join(parts, "  ")
}

func (m WorkflowModel) renderModelOptions() string {
	var b strings.Builder
	models := m.getLLMModels()

	// Show in grid format
	perRow := 3
	for i, mod := range models {
		if i > 0 && i%perRow == 0 {
			b.WriteString("\n                ") // Align with label
		}

		// Check if selected
		isSelected := false
		if m.editing != nil {
			for _, em := range m.editing.Models {
				if em.ID == mod.ID {
					isSelected = true
					break
				}
			}
		}

		style := wfOptionStyle
		prefix := "[ ]"
		if isSelected {
			style = wfOptionActiveStyle
			prefix = "[✓]"
		}
		if m.editField == fieldModels && i == m.editOption {
			style = wfOptionSelectedStyle
		}

		name := truncate(mod.Name, 18)
		b.WriteString(style.Render(fmt.Sprintf("%s %-18s ", prefix, name)))
	}
	return b.String()
}

func (m WorkflowModel) renderGPUOptions() string {
	var parts []string
	for i, p := range gpuPresets {
		style := wfOptionStyle
		if m.editing != nil && m.editing.GPUConfig.TotalGPUs == p.gpus && m.editing.GPUConfig.GPUType == p.gpuType {
			style = wfOptionActiveStyle
		}
		if m.editField == fieldGPU && i == m.editOption {
			style = wfOptionSelectedStyle
		}
		parts = append(parts, style.Render(p.label))
	}

	// Show in rows of 3
	var rows []string
	for i := 0; i < len(parts); i += 3 {
		end := i + 3
		if end > len(parts) {
			end = len(parts)
		}
		row := strings.Join(parts[i:end], "  ")
		if i > 0 {
			row = "                " + row // Align with label
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func (m WorkflowModel) renderOptOptions() string {
	opts := []struct {
		label   string
		enabled bool
	}{
		{"Flash", m.editing != nil && m.editing.Optimizations.FlashAttention},
		{"Paged", m.editing != nil && m.editing.Optimizations.PagedAttention},
		{"Speculative", m.editing != nil && m.editing.Optimizations.SpeculativeDecoding},
		{"ContBatch", m.editing != nil && m.editing.Optimizations.ContinuousBatching},
		{"Chunked", m.editing != nil && m.editing.Optimizations.ChunkedPrefill},
		{"Prefix", m.editing != nil && m.editing.Optimizations.PrefixCaching},
	}

	var parts []string
	for i, opt := range opts {
		style := wfOptionStyle
		prefix := "[ ]"
		if opt.enabled {
			style = wfOptionActiveStyle
			prefix = "[✓]"
		}
		if m.editField == fieldOptimizations && i == m.editOption {
			style = wfOptionSelectedStyle
		}
		parts = append(parts, style.Render(fmt.Sprintf("%s %s", prefix, opt.label)))
	}

	// Two rows
	row1 := strings.Join(parts[:3], "  ")
	row2 := "                " + strings.Join(parts[3:], "  ")
	return row1 + "\n" + row2
}

func (m WorkflowModel) renderAutoLoadOptions() string {
	opts := []string{"No", "Yes"}
	var parts []string
	for i, opt := range opts {
		style := wfOptionStyle
		isActive := (m.editing != nil && m.editing.AutoLoad && i == 1) ||
			(m.editing != nil && !m.editing.AutoLoad && i == 0)
		if isActive {
			style = wfOptionActiveStyle
		}
		if m.editField == fieldAutoLoad && i == m.editOption {
			style = wfOptionSelectedStyle
		}
		parts = append(parts, style.Render(opt))
	}
	return strings.Join(parts, "  ")
}

func (m WorkflowModel) renderSaveButton() string {
	style := wfOptionStyle
	if m.editField == fieldSave {
		style = wfOptionSelectedStyle
	}
	return "\n" + style.Render("                [ 💾 Save Workflow ]")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
