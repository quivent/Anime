package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

// ============================================================================
// DASHBOARD COMMAND
// ============================================================================

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"d", "dash", "explore", "browser", "tui"},
	Short:   "Full-scale TUI to explore all anime features",
	Long: `A comprehensive TUI interface with tabs to explore every feature,
option, and configuration of anime CLI.

Features:
  - Packages: Browse all packages grouped by category sub-tabs
  - Models: Browse AI models with details and HuggingFace links
  - Servers: Manage remote servers with pull/push controls
  - Files: File manager with copy/cut/paste/delete operations
  - Source: Source control operations (push, pull, sync)
  - Config: View and edit configuration settings
  - Aliases: View and manage registered aliases
  - Contents: Browse embedded agents, commands, and assets
  - Commands: Browse all CLI commands with descriptions
  - Reel: Video generation pipeline configuration`,
	Run: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

// ============================================================================
// STYLES
// ============================================================================

var (
	// Color palette
	dPink     = lipgloss.Color("#FF69B4")
	dCyan     = lipgloss.Color("#00D9FF")
	dPurple   = lipgloss.Color("#BD93F9")
	dGreen    = lipgloss.Color("#50FA7B")
	dOrange   = lipgloss.Color("#FFB86C")
	dRed      = lipgloss.Color("#FF5555")
	dYellow   = lipgloss.Color("#F1FA8C")
	dWhite    = lipgloss.Color("#F8F8F2")
	dGray     = lipgloss.Color("#6272A4")
	dDarkGray = lipgloss.Color("#44475A")
	dDark     = lipgloss.Color("#282A36")

	// Tab styles
	dTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(dGray)

	dActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(dDark).
			Background(dPink).
			Bold(true)

	dTabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(dDarkGray).
			MarginBottom(1)

	// Header styles
	dTitleStyle = lipgloss.NewStyle().
			Foreground(dPink).
			Bold(true).
			Padding(0, 1)

	dSubtitleStyle = lipgloss.NewStyle().
			Foreground(dGray).
			Italic(true)

	// Content styles
	dCategoryStyle = lipgloss.NewStyle().
			Foreground(dCyan).
			Bold(true).
			MarginTop(1)

	dItemStyle = lipgloss.NewStyle().
			Foreground(dWhite).
			PaddingLeft(2)

	dSelectedStyle = lipgloss.NewStyle().
			Foreground(dGreen).
			Bold(true).
			PaddingLeft(2)

	dDescStyle = lipgloss.NewStyle().
			Foreground(dGray).
			PaddingLeft(4)

	dDetailStyle = lipgloss.NewStyle().
			Foreground(dPurple)

	dValueStyle = lipgloss.NewStyle().
			Foreground(dOrange)

	dHelpStyle = lipgloss.NewStyle().
			Foreground(dGray).
			MarginTop(1)

	// Box styles
	dBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dPurple).
			Padding(0, 1)

	dSelectedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(dGreen).
				Padding(0, 1)

	// Status styles
	dSuccessStyle = lipgloss.NewStyle().Foreground(dGreen).Bold(true)
	dErrorStyle   = lipgloss.NewStyle().Foreground(dRed).Bold(true)
	dWarningStyle = lipgloss.NewStyle().Foreground(dYellow).Bold(true)
	dInfoStyle    = lipgloss.NewStyle().Foreground(dCyan)
)

// ============================================================================
// TAB DEFINITIONS
// ============================================================================

type tabID int

const (
	tabPackages tabID = iota
	tabModels
	tabServers
	tabFiles
	tabSource
	tabConfig
	tabKeys
	tabAliases
	tabContents
	tabCommands
	tabReel
)

var dashTabNames = []string{
	"Packages",
	"Models",
	"Servers",
	"Files",
	"Source",
	"Config",
	"Keys",
	"Aliases",
	"Contents",
	"Commands",
	"Reel",
}

// ============================================================================
// MAIN DASHBOARD MODEL
// ============================================================================

type dashboardModel struct {
	// Navigation
	activeTab tabID
	width     int
	height    int

	// Sub-models for each tab (lazy-loaded pointers)
	packagesModel  *packagesTabModel
	modelsModel    *modelsTabModel
	serversModel   *serversTabModel
	filesModel     *filesTabModel
	sourceModel    *sourceTabModel
	configModel    *configTabModel
	keysModel      *keysTabModel
	aliasesModel   *aliasesTabModel
	contentsModel  *contentsTabModel
	commandsModel  *commandsTabModel
	reelModel      *reelTabModel

	// State
	ready    bool
	err      error
	message  string
	cfg      *config.Config
}

func newDashboardModel() dashboardModel {
	cfg, _ := config.Load()

	// Only initialize the first tab - others are lazy-loaded on first access
	packagesTab := newPackagesTabModel()
	return dashboardModel{
		activeTab:     tabPackages,
		cfg:           cfg,
		packagesModel: &packagesTab,
		// Other tabs are nil and will be initialized on first access
	}
}

// Lazy initialization helpers for each tab
func (m *dashboardModel) getPackagesModel() *packagesTabModel {
	if m.packagesModel == nil {
		tab := newPackagesTabModel()
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.packagesModel = &tab
	}
	return m.packagesModel
}

func (m *dashboardModel) getModelsModel() *modelsTabModel {
	if m.modelsModel == nil {
		tab := newModelsTabModel()
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.modelsModel = &tab
	}
	return m.modelsModel
}

func (m *dashboardModel) getServersModel() *serversTabModel {
	if m.serversModel == nil {
		tab := newServersTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.serversModel = &tab
	}
	return m.serversModel
}

func (m *dashboardModel) getFilesModel() *filesTabModel {
	if m.filesModel == nil {
		tab := newFilesTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.filesModel = &tab
	}
	return m.filesModel
}

func (m *dashboardModel) getSourceModel() *sourceTabModel {
	if m.sourceModel == nil {
		tab := newSourceTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.sourceModel = &tab
	}
	return m.sourceModel
}

func (m *dashboardModel) getConfigModel() *configTabModel {
	if m.configModel == nil {
		tab := newConfigTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.configModel = &tab
	}
	return m.configModel
}

func (m *dashboardModel) getKeysModel() *keysTabModel {
	if m.keysModel == nil {
		tab := newKeysTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.keysModel = &tab
	}
	return m.keysModel
}

func (m *dashboardModel) getAliasesModel() *aliasesTabModel {
	if m.aliasesModel == nil {
		tab := newAliasesTabModel(m.cfg)
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.aliasesModel = &tab
	}
	return m.aliasesModel
}

func (m *dashboardModel) getContentsModel() *contentsTabModel {
	if m.contentsModel == nil {
		tab := newContentsTabModel()
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.contentsModel = &tab
	}
	return m.contentsModel
}

func (m *dashboardModel) getCommandsModel() *commandsTabModel {
	if m.commandsModel == nil {
		tab := newCommandsTabModel()
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.commandsModel = &tab
	}
	return m.commandsModel
}

func (m *dashboardModel) getReelModel() *reelTabModel {
	if m.reelModel == nil {
		tab := newReelTabModel()
		tab.width = m.width - 4
		tab.height = m.height - 8
		m.reelModel = &tab
	}
	return m.reelModel
}

func (m dashboardModel) Init() tea.Cmd {
	if m.packagesModel != nil {
		return tea.Batch(
			tea.EnterAltScreen,
			m.packagesModel.Init(),
		)
	}
	return tea.EnterAltScreen
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update only initialized sub-models with new size (lazy loading)
		contentHeight := m.height - 8 // Account for header and footer
		contentWidth := m.width - 4

		if m.packagesModel != nil {
			m.packagesModel.width = contentWidth
			m.packagesModel.height = contentHeight
		}
		if m.modelsModel != nil {
			m.modelsModel.width = contentWidth
			m.modelsModel.height = contentHeight
			if m.modelsModel.viewport.Width == 0 {
				m.modelsModel.viewport = viewport.New(contentWidth, contentHeight-4)
			} else {
				m.modelsModel.viewport.Width = contentWidth
				m.modelsModel.viewport.Height = contentHeight - 4
			}
		}
		if m.serversModel != nil {
			m.serversModel.width = contentWidth
			m.serversModel.height = contentHeight
		}
		if m.filesModel != nil {
			m.filesModel.width = contentWidth
			m.filesModel.height = contentHeight
		}
		if m.sourceModel != nil {
			m.sourceModel.width = contentWidth
			m.sourceModel.height = contentHeight
		}
		if m.configModel != nil {
			m.configModel.width = contentWidth
			m.configModel.height = contentHeight
		}
		if m.keysModel != nil {
			m.keysModel.width = contentWidth
			m.keysModel.height = contentHeight
		}
		if m.aliasesModel != nil {
			m.aliasesModel.width = contentWidth
			m.aliasesModel.height = contentHeight
		}
		if m.contentsModel != nil {
			m.contentsModel.width = contentWidth
			m.contentsModel.height = contentHeight
		}
		if m.commandsModel != nil {
			m.commandsModel.width = contentWidth
			m.commandsModel.height = contentHeight
		}
		if m.reelModel != nil {
			m.reelModel.width = contentWidth
			m.reelModel.height = contentHeight
		}

	case tea.KeyMsg:
		// Global key handlers
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "L":
			m.activeTab = (m.activeTab + 1) % tabID(len(dashTabNames))
			m.message = ""
			return m, nil

		case "shift+tab", "H":
			m.activeTab = (m.activeTab - 1 + tabID(len(dashTabNames))) % tabID(len(dashTabNames))
			m.message = ""
			return m, nil

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < len(dashTabNames) {
				m.activeTab = tabID(idx)
			}
			return m, nil
		case "0":
			if len(dashTabNames) >= 10 {
				m.activeTab = tabID(9) // 10th tab (0-indexed)
			}
			return m, nil
		}
	}

	// Route update to active tab (lazy initialization via getters)
	var cmd tea.Cmd
	switch m.activeTab {
	case tabPackages:
		tab := m.getPackagesModel()
		*tab, cmd = tab.Update(msg)
		// Check if user pressed Enter to install
		if tab.installing {
			return m, tea.Quit
		}
	case tabModels:
		tab := m.getModelsModel()
		*tab, cmd = tab.Update(msg)
	case tabServers:
		tab := m.getServersModel()
		*tab, cmd = tab.Update(msg)
	case tabFiles:
		tab := m.getFilesModel()
		*tab, cmd = tab.Update(msg)
	case tabSource:
		tab := m.getSourceModel()
		*tab, cmd = tab.Update(msg)
	case tabConfig:
		tab := m.getConfigModel()
		*tab, cmd = tab.Update(msg)
	case tabKeys:
		tab := m.getKeysModel()
		*tab, cmd = tab.Update(msg)
	case tabAliases:
		tab := m.getAliasesModel()
		*tab, cmd = tab.Update(msg)
	case tabContents:
		tab := m.getContentsModel()
		*tab, cmd = tab.Update(msg)
	case tabCommands:
		tab := m.getCommandsModel()
		*tab, cmd = tab.Update(msg)
	case tabReel:
		tab := m.getReelModel()
		*tab, cmd = tab.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m dashboardModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var s strings.Builder

	// Header
	s.WriteString("\n")
	s.WriteString(dTitleStyle.Render("  ANIME DASHBOARD"))
	s.WriteString("  ")
	s.WriteString(dSubtitleStyle.Render("Full-Scale TUI Explorer"))
	s.WriteString("\n\n")

	// Tab bar
	var tabs []string
	for i, name := range dashTabNames {
		if tabID(i) == m.activeTab {
			tabs = append(tabs, dActiveTabStyle.Render(fmt.Sprintf("%d:%s", i+1, name)))
		} else {
			tabs = append(tabs, dTabStyle.Render(fmt.Sprintf("%d:%s", i+1, name)))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	s.WriteString(dTabBarStyle.Render(tabBar))
	s.WriteString("\n")

	// Content area - delegate to active tab (lazy initialization via getters)
	switch m.activeTab {
	case tabPackages:
		s.WriteString(m.getPackagesModel().View())
	case tabModels:
		s.WriteString(m.getModelsModel().View())
	case tabServers:
		s.WriteString(m.getServersModel().View())
	case tabFiles:
		s.WriteString(m.getFilesModel().View())
	case tabSource:
		s.WriteString(m.getSourceModel().View())
	case tabConfig:
		s.WriteString(m.getConfigModel().View())
	case tabKeys:
		s.WriteString(m.getKeysModel().View())
	case tabAliases:
		s.WriteString(m.getAliasesModel().View())
	case tabContents:
		s.WriteString(m.getContentsModel().View())
	case tabCommands:
		s.WriteString(m.getCommandsModel().View())
	case tabReel:
		s.WriteString(m.getReelModel().View())
	}

	// Message area
	if m.message != "" {
		s.WriteString("\n")
		s.WriteString(dInfoStyle.Render("  " + m.message))
	}

	// Footer help
	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  Tab/1-0: Switch tabs | Arrows/hjkl: Navigate | Enter: Select | q: Quit"))
	s.WriteString("\n")

	return s.String()
}

// ============================================================================
// PACKAGES TAB
// ============================================================================

type packagesTabModel struct {
	packages      map[string][]*installer.Package // Grouped by category
	categories    []string
	currentCat    int
	currentPkg    int
	width         int
	height        int
	showDetails   bool
	selected      map[string]bool
	installing    bool // Set to true when user presses Enter to install
}

func newPackagesTabModel() packagesTabModel {
	pkgs := installer.GetPackages()

	// Group by category
	grouped := make(map[string][]*installer.Package)
	for _, pkg := range pkgs {
		grouped[pkg.Category] = append(grouped[pkg.Category], pkg)
	}

	// Sort categories
	categories := []string{}
	catOrder := []string{
		"Foundation", "GPU", "Orchestration", "Runtime", "ML Framework",
		"LLM Runtime", "Models", "LLM", "Image Generation", "Video Generation",
		"Image Enhancement", "Video Enhancement", "ControlNet", "Application", "ComfyUI Node",
	}

	seen := make(map[string]bool)
	for _, cat := range catOrder {
		if _, ok := grouped[cat]; ok && !seen[cat] {
			categories = append(categories, cat)
			seen[cat] = true
		}
	}
	// Add any remaining categories
	for cat := range grouped {
		if !seen[cat] {
			categories = append(categories, cat)
		}
	}

	// Sort packages within each category
	for _, pkgList := range grouped {
		sort.Slice(pkgList, func(i, j int) bool {
			return pkgList[i].Name < pkgList[j].Name
		})
	}

	return packagesTabModel{
		packages:   grouped,
		categories: categories,
		selected:   make(map[string]bool),
	}
}

func (m packagesTabModel) Init() tea.Cmd {
	return nil
}

func (m packagesTabModel) Update(msg tea.Msg) (packagesTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentPkg > 0 {
				m.currentPkg--
			}
		case "down", "j":
			pkgs := m.packages[m.categories[m.currentCat]]
			if m.currentPkg < len(pkgs)-1 {
				m.currentPkg++
			}
		case "left", "h":
			if m.currentCat > 0 {
				m.currentCat--
				m.currentPkg = 0
			}
		case "right", "l":
			if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentPkg = 0
			}
		case " ":
			// Toggle selection on current package
			pkgs := m.packages[m.categories[m.currentCat]]
			if m.currentPkg < len(pkgs) {
				pkg := pkgs[m.currentPkg]
				m.selected[pkg.ID] = !m.selected[pkg.ID]
			}
		case "enter":
			// If packages selected, install them. Otherwise toggle current package.
			if len(m.selected) > 0 {
				m.installing = true
			} else {
				// Toggle selection on current package
				pkgs := m.packages[m.categories[m.currentCat]]
				if m.currentPkg < len(pkgs) {
					pkg := pkgs[m.currentPkg]
					m.selected[pkg.ID] = !m.selected[pkg.ID]
				}
			}
		case "d":
			m.showDetails = !m.showDetails
		case "a":
			// Select all in current category
			pkgs := m.packages[m.categories[m.currentCat]]
			for _, pkg := range pkgs {
				m.selected[pkg.ID] = true
			}
		case "A":
			// Deselect all
			m.selected = make(map[string]bool)
		}
	}
	return m, nil
}

func (m packagesTabModel) View() string {
	var s strings.Builder

	// Category sub-tabs bar
	var catTabs []string
	for i, cat := range m.categories {
		emoji := getDashCategoryEmoji(cat)
		count := len(m.packages[cat])
		label := fmt.Sprintf("%s%d", emoji, count)
		if i == m.currentCat {
			catTabs = append(catTabs, dActiveTabStyle.Render(label))
		} else {
			catTabs = append(catTabs, dTabStyle.Render(label))
		}
	}
	s.WriteString("  ")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, catTabs...))
	s.WriteString("\n")

	// Current category name
	currentCatName := m.categories[m.currentCat]
	s.WriteString(dCategoryStyle.Render(fmt.Sprintf("  %s %s", getDashCategoryEmoji(currentCatName), currentCatName)))
	s.WriteString(dSubtitleStyle.Render(fmt.Sprintf(" (%d packages)", len(m.packages[currentCatName]))))
	s.WriteString("\n\n")

	// Show only packages in current category
	pkgs := m.packages[currentCatName]
	maxHeight := m.height - 12
	if maxHeight < 5 {
		maxHeight = 5
	}

	// Calculate scroll offset
	startIdx := 0
	if m.currentPkg > maxHeight-3 {
		startIdx = m.currentPkg - maxHeight + 3
	}

	for i := startIdx; i < len(pkgs) && i < startIdx+maxHeight; i++ {
		pkg := pkgs[i]

		// Selection indicator
		checkbox := "[ ]"
		if m.selected[pkg.ID] {
			checkbox = dSuccessStyle.Render("[x]")
		}

		// Package line
		if i == m.currentPkg {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s %s", checkbox, pkg.Name)))
			s.WriteString(dDetailStyle.Render(fmt.Sprintf(" (%s)", pkg.Size)))
			s.WriteString("\n")

			// Show description when selected
			s.WriteString(dDescStyle.Render(fmt.Sprintf("      %s", pkg.Description)))
			s.WriteString("\n")

			if m.showDetails {
				if len(pkg.Dependencies) > 0 {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Deps: %s", strings.Join(pkg.Dependencies, ", "))))
					s.WriteString("\n")
				}
				s.WriteString(dDescStyle.Render(fmt.Sprintf("      Time: %s", pkg.EstimatedTime)))
				s.WriteString("\n")
			}
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s %s", checkbox, pkg.Name)))
			s.WriteString(dSubtitleStyle.Render(fmt.Sprintf(" (%s)", pkg.Size)))
			s.WriteString("\n")
		}
	}

	// Selected count
	if len(m.selected) > 0 {
		s.WriteString("\n")
		s.WriteString(dSuccessStyle.Render(fmt.Sprintf("  Selected: %d packages", len(m.selected))))
	}

	// Help
	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  Space: Toggle | d: Details | Enter: Install | Arrows: Navigate"))

	return s.String()
}

func getDashCategoryEmoji(cat string) string {
	switch cat {
	case "Foundation":
		return "🏗️"
	case "GPU":
		return "🎮"
	case "Orchestration":
		return "🎭"
	case "Runtime":
		return "⚡"
	case "ML Framework":
		return "🤖"
	case "LLM Runtime":
		return "🔮"
	case "Models", "LLM":
		return "🧠"
	case "Image Generation":
		return "🎨"
	case "Video Generation":
		return "🎬"
	case "Image Enhancement":
		return "✨"
	case "Video Enhancement":
		return "📹"
	case "ControlNet":
		return "🎛️"
	case "Application":
		return "🎯"
	case "ComfyUI Node":
		return "🔌"
	default:
		return "📦"
	}
}

// ============================================================================
// MODELS TAB - with sub-tabs by category
// ============================================================================

type modelInfo struct {
	Name        string
	Size        string
	Description string
	Category    string
	Type        string
	UseCases    []string
	HFLink      string
}

// Model categories for sub-tabs
var modelCategories = []string{
	"Frontier",
	"Large",
	"Medium",
	"Small",
	"Coding",
	"Multimodal",
	"Image",
	"Video",
	"Enhance",
	"Control",
}

func getModelCategoryEmoji(cat string) string {
	switch cat {
	case "Frontier":
		return "🌟"
	case "Large":
		return "🦁"
	case "Medium":
		return "🔥"
	case "Small":
		return "⚡"
	case "Coding":
		return "💻"
	case "Multimodal":
		return "👁️"
	case "Image":
		return "🎨"
	case "Video":
		return "🎬"
	case "Enhance":
		return "✨"
	case "Control":
		return "🎛️"
	default:
		return "📦"
	}
}

type modelsTabModel struct {
	models      map[string][]modelInfo // Grouped by category
	categories  []string
	currentCat  int
	currentIdx  int
	width       int
	height      int
	viewport    viewport.Model
	showDetails bool
}

func newModelsTabModel() modelsTabModel {
	models := getModelInfoList()

	// Group by Type (which maps to our categories)
	grouped := make(map[string][]modelInfo)
	for _, m := range models {
		grouped[m.Type] = append(grouped[m.Type], m)
	}

	return modelsTabModel{
		models:     grouped,
		categories: modelCategories,
		viewport:   viewport.New(80, 20),
	}
}

func (m modelsTabModel) Init() tea.Cmd {
	return nil
}

func (m modelsTabModel) Update(msg tea.Msg) (modelsTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			models := m.models[m.categories[m.currentCat]]
			if m.currentIdx < len(models)-1 {
				m.currentIdx++
			}
		case "left", "h":
			if m.currentCat > 0 {
				m.currentCat--
				m.currentIdx = 0
			}
		case "right", "l":
			if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentIdx = 0
			}
		case "d":
			m.showDetails = !m.showDetails
		case "o", "enter":
			// Open HuggingFace link
			models := m.models[m.categories[m.currentCat]]
			if m.currentIdx < len(models) && models[m.currentIdx].HFLink != "" {
				exec.Command("open", models[m.currentIdx].HFLink).Start()
			}
		}
	}
	return m, nil
}

func (m modelsTabModel) View() string {
	var s strings.Builder

	// Category sub-tabs bar (like packages)
	var catTabs []string
	for i, cat := range m.categories {
		emoji := getModelCategoryEmoji(cat)
		count := len(m.models[cat])
		label := fmt.Sprintf("%s%d", emoji, count)
		if i == m.currentCat {
			catTabs = append(catTabs, dActiveTabStyle.Render(label))
		} else {
			catTabs = append(catTabs, dTabStyle.Render(label))
		}
	}
	s.WriteString("  ")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, catTabs...))
	s.WriteString("\n")

	// Current category name
	currentCatName := m.categories[m.currentCat]
	s.WriteString(dCategoryStyle.Render(fmt.Sprintf("  %s %s", getModelCategoryEmoji(currentCatName), currentCatName)))
	s.WriteString(dSubtitleStyle.Render(fmt.Sprintf(" (%d models)", len(m.models[currentCatName]))))
	s.WriteString("\n\n")

	// Show only models in current category
	models := m.models[currentCatName]
	maxHeight := m.height - 12
	if maxHeight < 5 {
		maxHeight = 5
	}

	// Calculate scroll offset
	startIdx := 0
	if m.currentIdx > maxHeight-3 {
		startIdx = m.currentIdx - maxHeight + 3
	}

	for i := startIdx; i < len(models) && i < startIdx+maxHeight; i++ {
		model := models[i]

		if i == m.currentIdx {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", model.Name)))
			s.WriteString(dDetailStyle.Render(fmt.Sprintf(" [%s]", model.Size)))
			s.WriteString("\n")

			// Show description when selected
			s.WriteString(dDescStyle.Render(fmt.Sprintf("      %s", model.Description)))
			s.WriteString("\n")

			if m.showDetails {
				if model.Category != "" {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Subcategory: %s", model.Category)))
					s.WriteString("\n")
				}
				if len(model.UseCases) > 0 {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Uses: %s", strings.Join(model.UseCases, ", "))))
					s.WriteString("\n")
				}
				if model.HFLink != "" {
					s.WriteString(dInfoStyle.Render(fmt.Sprintf("      HF: %s", model.HFLink)))
					s.WriteString("\n")
				}
			}
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", model.Name)))
			s.WriteString(dSubtitleStyle.Render(fmt.Sprintf(" [%s]", model.Size)))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  ←→: Categories | ↑↓: Models | d: Details | o: Open HuggingFace"))

	return s.String()
}

func getModelInfoList() []modelInfo {
	return []modelInfo{
		// =====================================================================
		// FRONTIER (100B+, Multi-GPU Required)
		// =====================================================================
		{Name: "DeepSeek-R1 671B", Size: "~400GB", Description: "Full DeepSeek-R1, state-of-the-art reasoning. Requires 8x H100", Type: "Frontier", Category: "Reasoning", UseCases: []string{"Research", "Complex math", "PhD-level problems"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-R1"},
		{Name: "DeepSeek-V3", Size: "~400GB", Description: "671B MoE, 37B active. Near-GPT-4 performance", Type: "Frontier", Category: "General", UseCases: []string{"General AI", "Code", "Multilingual"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-V3"},
		{Name: "Qwen2.5 72B", Size: "~145GB", Description: "Alibaba's flagship, exceptional coding & math", Type: "Frontier", Category: "General", UseCases: []string{"Coding", "Math", "Reasoning"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-72B-Instruct"},
		{Name: "Llama 3.1 405B", Size: "~230GB", Description: "Meta's largest, frontier-class open model", Type: "Frontier", Category: "General", UseCases: []string{"Research", "Enterprise", "Complex tasks"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.1-405B-Instruct"},
		{Name: "Mixtral 8x22B", Size: "~260GB", Description: "176B MoE, 39B active. Strong multilingual", Type: "Frontier", Category: "MoE", UseCases: []string{"Multilingual", "Long context", "Coding"}, HFLink: "https://huggingface.co/mistralai/Mixtral-8x22B-Instruct-v0.1"},
		{Name: "DBRX 132B", Size: "~260GB", Description: "Databricks MoE, 36B active. Strong coding", Type: "Frontier", Category: "MoE", UseCases: []string{"Enterprise", "Coding", "Data analysis"}},
		{Name: "Falcon 180B", Size: "~360GB", Description: "TII's largest, strong benchmark performance", Type: "Frontier", Category: "General", UseCases: []string{"Research", "Multilingual"}},
		{Name: "Command R+", Size: "~200GB", Description: "Cohere's 104B, optimized for RAG & enterprise", Type: "Frontier", Category: "Enterprise", UseCases: []string{"RAG", "Enterprise", "Tool use"}, HFLink: "https://huggingface.co/CohereForAI/c4ai-command-r-plus"},

		// =====================================================================
		// LARGE (70B class, 2-4 GPUs)
		// =====================================================================
		{Name: "Llama 3.3 70B", Size: "~40GB", Description: "Meta's latest 70B, exceptional reasoning & coding", Type: "Large", Category: "General", UseCases: []string{"Code generation", "Complex reasoning", "Research"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.3-70B-Instruct"},
		{Name: "Llama 3.1 70B", Size: "~40GB", Description: "Previous gen, still excellent. 128K context", Type: "Large", Category: "General", UseCases: []string{"Long context", "General tasks"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.1-70B-Instruct"},
		{Name: "DeepSeek-R1-Distill-Llama-70B", Size: "~43GB", Description: "R1 distilled to Llama 70B. Near-frontier reasoning", Type: "Large", Category: "Reasoning", UseCases: []string{"Math", "Code", "Logic"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-R1-Distill-Llama-70B"},
		{Name: "DeepSeek-R1-Distill-Qwen-70B", Size: "~43GB", Description: "R1 distilled to Qwen 70B. Strong math", Type: "Large", Category: "Reasoning", UseCases: []string{"Math", "Reasoning"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-R1-Distill-Qwen-70B"},
		{Name: "Qwen2.5-72B-Instruct", Size: "~45GB", Description: "Alibaba 72B, SOTA coding & multilingual", Type: "Large", Category: "General", UseCases: []string{"Coding", "Math", "Chinese"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-72B-Instruct"},
		{Name: "CodeLlama 70B", Size: "~40GB", Description: "Meta's code-specialized 70B", Type: "Large", Category: "Coding", UseCases: []string{"Code generation", "Code review"}, HFLink: "https://huggingface.co/codellama/CodeLlama-70b-Instruct-hf"},
		{Name: "Nemotron-70B", Size: "~40GB", Description: "NVIDIA's 70B, strong instruction following", Type: "Large", Category: "General", UseCases: []string{"Instruction following", "Chat"}},

		// =====================================================================
		// MEDIUM (14-70B, 1-2 GPUs)
		// =====================================================================
		{Name: "QwQ 32B Preview", Size: "~20GB", Description: "Qwen reasoning model, o1-style thinking", Type: "Medium", Category: "Reasoning", UseCases: []string{"Step-by-step reasoning", "Math", "Logic"}, HFLink: "https://huggingface.co/Qwen/QwQ-32B-Preview"},
		{Name: "Qwen2.5 32B", Size: "~20GB", Description: "Dense 32B, excellent balance of capability", Type: "Medium", Category: "General", UseCases: []string{"General tasks", "Coding"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-32B-Instruct"},
		{Name: "DeepSeek-R1-Distill-Qwen-32B", Size: "~20GB", Description: "R1 distilled to Qwen 32B. Strong reasoning", Type: "Medium", Category: "Reasoning", UseCases: []string{"Math", "Logic"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-R1-Distill-Qwen-32B"},
		{Name: "Mixtral 8x7B", Size: "~26GB", Description: "47B MoE, ~13B active. Fast & capable", Type: "Medium", Category: "MoE", UseCases: []string{"Multi-task", "Long context"}, HFLink: "https://huggingface.co/mistralai/Mixtral-8x7B-Instruct-v0.1"},
		{Name: "Gemma2 27B", Size: "~17GB", Description: "Google's latest, excellent instruction following", Type: "Medium", Category: "General", UseCases: []string{"Chat", "Reasoning"}, HFLink: "https://huggingface.co/google/gemma-2-27b-it"},
		{Name: "Yi-1.5 34B Chat", Size: "~20GB", Description: "01.AI bilingual, strong Chinese & English", Type: "Medium", Category: "Bilingual", UseCases: []string{"Chinese", "English"}, HFLink: "https://huggingface.co/01-ai/Yi-1.5-34B-Chat"},
		{Name: "Command R 35B", Size: "~21GB", Description: "Cohere's RAG-optimized model", Type: "Medium", Category: "RAG", UseCases: []string{"RAG", "Tool use", "Enterprise"}, HFLink: "https://huggingface.co/CohereForAI/c4ai-command-r-v01"},
		{Name: "Phi-4 14B", Size: "~9GB", Description: "Microsoft's reasoning champion, punches above weight", Type: "Medium", Category: "Reasoning", UseCases: []string{"Reasoning", "Math", "Edge"}, HFLink: "https://huggingface.co/microsoft/phi-4"},
		{Name: "Mistral Small 22B", Size: "~14GB", Description: "Mistral's efficient 22B, strong coding", Type: "Medium", Category: "General", UseCases: []string{"Coding", "General tasks"}, HFLink: "https://huggingface.co/mistralai/Mistral-Small-Instruct-2409"},
		{Name: "InternLM2.5 20B", Size: "~12GB", Description: "Shanghai AI Lab, strong Chinese & math", Type: "Medium", Category: "Bilingual", UseCases: []string{"Math", "Chinese", "Reasoning"}},

		// =====================================================================
		// SMALL (1-14B, Single Consumer GPU)
		// =====================================================================
		{Name: "Llama 3.2 3B", Size: "~2GB", Description: "Meta's tiny powerhouse, runs anywhere", Type: "Small", Category: "Tiny", UseCases: []string{"Edge", "Mobile", "Fast chat"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.2-3B-Instruct"},
		{Name: "Llama 3.2 1B", Size: "~1GB", Description: "Ultra-lightweight, on-device AI", Type: "Small", Category: "Tiny", UseCases: []string{"Mobile", "IoT", "Embedded"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.2-1B-Instruct"},
		{Name: "Llama 3.1 8B", Size: "~5GB", Description: "Excellent small model, 128K context", Type: "Small", Category: "General", UseCases: []string{"General tasks", "Long docs"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.1-8B-Instruct"},
		{Name: "DeepSeek-R1-Distill-Llama-8B", Size: "~5GB", Description: "R1 reasoning in 8B package", Type: "Small", Category: "Reasoning", UseCases: []string{"Math", "Logic", "Consumer GPU"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-R1-Distill-Llama-8B"},
		{Name: "DeepSeek-R1-Distill-Qwen-7B", Size: "~4.5GB", Description: "R1 distilled to Qwen 7B", Type: "Small", Category: "Reasoning", UseCases: []string{"Reasoning", "Math"}},
		{Name: "DeepSeek-R1-Distill-Qwen-1.5B", Size: "~1GB", Description: "Tiny R1 distillation, runs anywhere", Type: "Small", Category: "Tiny", UseCases: []string{"Mobile", "Edge", "Fast"}},
		{Name: "Qwen2.5 7B", Size: "~4.5GB", Description: "Strong 7B, excellent multilingual", Type: "Small", Category: "General", UseCases: []string{"Multilingual", "Coding"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct"},
		{Name: "Qwen2.5 3B", Size: "~2GB", Description: "Tiny but capable Qwen", Type: "Small", Category: "Tiny", UseCases: []string{"Edge", "Mobile"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct"},
		{Name: "Qwen2.5 1.5B", Size: "~1GB", Description: "Ultra-lightweight Qwen", Type: "Small", Category: "Tiny", UseCases: []string{"Edge", "IoT"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct"},
		{Name: "Qwen2.5 0.5B", Size: "~350MB", Description: "Smallest Qwen, still useful", Type: "Small", Category: "Tiny", UseCases: []string{"Embedded", "Testing"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct"},
		{Name: "Mistral 7B v0.3", Size: "~4GB", Description: "Mistral's workhorse, excellent coding", Type: "Small", Category: "General", UseCases: []string{"Coding", "Chat"}, HFLink: "https://huggingface.co/mistralai/Mistral-7B-Instruct-v0.3"},
		{Name: "Gemma2 9B", Size: "~6GB", Description: "Google's efficient 9B, strong reasoning", Type: "Small", Category: "General", UseCases: []string{"Reasoning", "Chat"}, HFLink: "https://huggingface.co/google/gemma-2-9b-it"},
		{Name: "Gemma2 2B", Size: "~1.5GB", Description: "Tiny Gemma, great for edge", Type: "Small", Category: "Tiny", UseCases: []string{"Edge", "Mobile"}, HFLink: "https://huggingface.co/google/gemma-2-2b-it"},
		{Name: "Phi-3.5 Mini 3.8B", Size: "~2.5GB", Description: "Microsoft's tiny reasoning model", Type: "Small", Category: "Reasoning", UseCases: []string{"Reasoning", "Edge"}, HFLink: "https://huggingface.co/microsoft/Phi-3.5-mini-instruct"},
		{Name: "TinyLlama 1.1B", Size: "~700MB", Description: "Compact Llama architecture", Type: "Small", Category: "Tiny", UseCases: []string{"Embedded", "Testing"}},
		{Name: "SmolLM2 1.7B", Size: "~1GB", Description: "HuggingFace's efficient tiny model", Type: "Small", Category: "Tiny", UseCases: []string{"Edge", "Mobile"}, HFLink: "https://huggingface.co/HuggingFaceTB/SmolLM2-1.7B-Instruct"},
		{Name: "Yi-1.5 6B", Size: "~4GB", Description: "01.AI's efficient bilingual", Type: "Small", Category: "Bilingual", UseCases: []string{"Chinese", "English"}},
		{Name: "Zephyr 7B", Size: "~4GB", Description: "HuggingFace's DPO-tuned Mistral", Type: "Small", Category: "Chat", UseCases: []string{"Chat", "Helpfulness"}, HFLink: "https://huggingface.co/HuggingFaceH4/zephyr-7b-beta"},

		// =====================================================================
		// CODING SPECIALISTS
		// =====================================================================
		{Name: "Qwen2.5-Coder 32B", Size: "~20GB", Description: "SOTA open-source coding model", Type: "Coding", Category: "Large", UseCases: []string{"Code generation", "Debugging", "Review"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct"},
		{Name: "Qwen2.5-Coder 14B", Size: "~9GB", Description: "Excellent coding in smaller package", Type: "Coding", Category: "Medium", UseCases: []string{"Code completion", "Refactoring"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-Coder-14B-Instruct"},
		{Name: "Qwen2.5-Coder 7B", Size: "~4.5GB", Description: "Efficient coding specialist", Type: "Coding", Category: "Small", UseCases: []string{"Code completion", "Scripts"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-Coder-7B-Instruct"},
		{Name: "Qwen2.5-Coder 3B", Size: "~2GB", Description: "Tiny but capable coder", Type: "Coding", Category: "Tiny", UseCases: []string{"Autocomplete", "Simple tasks"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-Coder-3B-Instruct"},
		{Name: "Qwen2.5-Coder 1.5B", Size: "~1GB", Description: "Ultra-lightweight coder", Type: "Coding", Category: "Tiny", UseCases: []string{"Editor plugins", "Fast completion"}, HFLink: "https://huggingface.co/Qwen/Qwen2.5-Coder-1.5B-Instruct"},
		{Name: "DeepSeek-Coder-V2 236B", Size: "~140GB", Description: "Frontier-class coding MoE", Type: "Coding", Category: "Frontier", UseCases: []string{"Complex projects", "Architecture"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-Coder-V2-Instruct"},
		{Name: "DeepSeek-Coder-V2 16B", Size: "~10GB", Description: "Efficient coding MoE, 2.4B active", Type: "Coding", Category: "Medium", UseCases: []string{"Code generation", "Review"}, HFLink: "https://huggingface.co/deepseek-ai/DeepSeek-Coder-V2-Lite-Instruct"},
		{Name: "DeepSeek-Coder 33B", Size: "~20GB", Description: "Strong coding, fill-in-middle support", Type: "Coding", Category: "Large", UseCases: []string{"Code generation", "FIM"}, HFLink: "https://huggingface.co/deepseek-ai/deepseek-coder-33b-instruct"},
		{Name: "DeepSeek-Coder 6.7B", Size: "~4GB", Description: "Efficient coding model", Type: "Coding", Category: "Small", UseCases: []string{"Code completion", "Scripts"}},
		{Name: "CodeLlama 34B", Size: "~20GB", Description: "Meta's code specialist", Type: "Coding", Category: "Large", UseCases: []string{"Code generation", "Infilling"}, HFLink: "https://huggingface.co/codellama/CodeLlama-34b-Instruct-hf"},
		{Name: "CodeLlama 13B", Size: "~8GB", Description: "Balanced CodeLlama", Type: "Coding", Category: "Medium", UseCases: []string{"Code completion"}},
		{Name: "CodeLlama 7B", Size: "~4GB", Description: "Efficient CodeLlama", Type: "Coding", Category: "Small", UseCases: []string{"Fast completion"}},
		{Name: "StarCoder2 15B", Size: "~9GB", Description: "BigCode's latest, 600+ languages", Type: "Coding", Category: "Medium", UseCases: []string{"Multi-language", "Completion"}, HFLink: "https://huggingface.co/bigcode/starcoder2-15b"},
		{Name: "StarCoder2 7B", Size: "~4.5GB", Description: "Efficient multi-language coder", Type: "Coding", Category: "Small", UseCases: []string{"Completion", "Scripts"}, HFLink: "https://huggingface.co/bigcode/starcoder2-7b"},
		{Name: "StarCoder2 3B", Size: "~2GB", Description: "Tiny StarCoder", Type: "Coding", Category: "Tiny", UseCases: []string{"Autocomplete"}},
		{Name: "Codestral 22B", Size: "~14GB", Description: "Mistral's coding model, 32K context", Type: "Coding", Category: "Medium", UseCases: []string{"Code generation", "Long context"}, HFLink: "https://huggingface.co/mistralai/Codestral-22B-v0.1"},

		// =====================================================================
		// MULTIMODAL (Vision-Language)
		// =====================================================================
		{Name: "Llama 3.2 90B Vision", Size: "~55GB", Description: "Meta's largest vision model", Type: "Multimodal", Category: "Large", UseCases: []string{"Image understanding", "Visual QA"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.2-90B-Vision-Instruct"},
		{Name: "Llama 3.2 11B Vision", Size: "~7GB", Description: "Efficient vision-language model", Type: "Multimodal", Category: "Medium", UseCases: []string{"Image analysis", "OCR"}, HFLink: "https://huggingface.co/meta-llama/Llama-3.2-11B-Vision-Instruct"},
		{Name: "Qwen2-VL 72B", Size: "~45GB", Description: "SOTA vision-language, video support", Type: "Multimodal", Category: "Large", UseCases: []string{"Image", "Video", "Documents"}, HFLink: "https://huggingface.co/Qwen/Qwen2-VL-72B-Instruct"},
		{Name: "Qwen2-VL 7B", Size: "~4.5GB", Description: "Efficient multimodal Qwen", Type: "Multimodal", Category: "Small", UseCases: []string{"Image understanding"}, HFLink: "https://huggingface.co/Qwen/Qwen2-VL-7B-Instruct"},
		{Name: "Qwen2-VL 2B", Size: "~1.5GB", Description: "Tiny vision model", Type: "Multimodal", Category: "Tiny", UseCases: []string{"Edge vision"}, HFLink: "https://huggingface.co/Qwen/Qwen2-VL-2B-Instruct"},
		{Name: "InternVL2 76B", Size: "~45GB", Description: "Shanghai AI Lab's SOTA VLM", Type: "Multimodal", Category: "Large", UseCases: []string{"Vision", "OCR", "Charts"}, HFLink: "https://huggingface.co/OpenGVLab/InternVL2-76B"},
		{Name: "InternVL2 26B", Size: "~16GB", Description: "Excellent vision-language", Type: "Multimodal", Category: "Medium", UseCases: []string{"Image analysis"}, HFLink: "https://huggingface.co/OpenGVLab/InternVL2-26B"},
		{Name: "InternVL2 8B", Size: "~5GB", Description: "Efficient InternVL", Type: "Multimodal", Category: "Small", UseCases: []string{"General vision"}, HFLink: "https://huggingface.co/OpenGVLab/InternVL2-8B"},
		{Name: "LLaVA-1.6 34B", Size: "~20GB", Description: "Strong visual instruction following", Type: "Multimodal", Category: "Large", UseCases: []string{"Visual QA", "Description"}, HFLink: "https://huggingface.co/llava-hf/llava-v1.6-34b-hf"},
		{Name: "LLaVA-1.6 13B", Size: "~8GB", Description: "Balanced LLaVA", Type: "Multimodal", Category: "Medium", UseCases: []string{"Image chat"}},
		{Name: "LLaVA-1.6 7B", Size: "~4GB", Description: "Efficient LLaVA", Type: "Multimodal", Category: "Small", UseCases: []string{"Basic vision"}},
		{Name: "Pixtral 12B", Size: "~7GB", Description: "Mistral's vision model", Type: "Multimodal", Category: "Medium", UseCases: []string{"Image understanding"}, HFLink: "https://huggingface.co/mistralai/Pixtral-12B-2409"},
		{Name: "PaliGemma 3B", Size: "~2GB", Description: "Google's tiny vision-language", Type: "Multimodal", Category: "Tiny", UseCases: []string{"Edge vision"}, HFLink: "https://huggingface.co/google/paligemma-3b-mix-224"},
		{Name: "MiniCPM-V 2.6", Size: "~5GB", Description: "Efficient multimodal, strong OCR", Type: "Multimodal", Category: "Small", UseCases: []string{"OCR", "Documents"}, HFLink: "https://huggingface.co/openbmb/MiniCPM-V-2_6"},
		{Name: "CogVLM2 19B", Size: "~12GB", Description: "Zhipu's vision-language model", Type: "Multimodal", Category: "Medium", UseCases: []string{"Visual grounding"}, HFLink: "https://huggingface.co/THUDM/cogvlm2-llama3-chat-19B"},
		{Name: "Molmo 7B", Size: "~4GB", Description: "AI2's multimodal model", Type: "Multimodal", Category: "Small", UseCases: []string{"Image understanding"}, HFLink: "https://huggingface.co/allenai/Molmo-7B-D-0924"},

		// =====================================================================
		// IMAGE GENERATION
		// =====================================================================
		{Name: "FLUX.1 Dev", Size: "~12GB", Description: "Black Forest Labs flagship, exceptional quality", Type: "Image", Category: "Professional", UseCases: []string{"Photorealism", "Typography"}, HFLink: "https://huggingface.co/black-forest-labs/FLUX.1-dev"},
		{Name: "FLUX.1 Schnell", Size: "~12GB", Description: "Fast 4-step generation", Type: "Image", Category: "Fast", UseCases: []string{"Rapid prototyping", "Preview"}, HFLink: "https://huggingface.co/black-forest-labs/FLUX.1-schnell"},
		{Name: "SD 3.5 Large", Size: "~16GB", Description: "Stability AI's 8B flagship", Type: "Image", Category: "Professional", UseCases: []string{"Production", "Quality"}, HFLink: "https://huggingface.co/stabilityai/stable-diffusion-3.5-large"},
		{Name: "SD 3.5 Large Turbo", Size: "~16GB", Description: "Fast SD3.5, fewer steps", Type: "Image", Category: "Fast", UseCases: []string{"Quick iteration"}, HFLink: "https://huggingface.co/stabilityai/stable-diffusion-3.5-large-turbo"},
		{Name: "SD 3.5 Medium", Size: "~10GB", Description: "Balanced quality/speed", Type: "Image", Category: "Balanced", UseCases: []string{"General use"}, HFLink: "https://huggingface.co/stabilityai/stable-diffusion-3.5-medium"},
		{Name: "Stable Diffusion XL", Size: "~7GB", Description: "Industry standard, huge ecosystem", Type: "Image", Category: "Standard", UseCases: []string{"General", "LoRAs"}, HFLink: "https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0"},
		{Name: "SDXL Turbo", Size: "~7GB", Description: "1-4 step SDXL generation", Type: "Image", Category: "Fast", UseCases: []string{"Real-time", "Preview"}, HFLink: "https://huggingface.co/stabilityai/sdxl-turbo"},
		{Name: "Stable Diffusion 1.5", Size: "~4GB", Description: "Classic, massive LoRA library", Type: "Image", Category: "Classic", UseCases: []string{"Art", "LoRAs"}, HFLink: "https://huggingface.co/runwayml/stable-diffusion-v1-5"},
		{Name: "Playground v2.5", Size: "~7GB", Description: "Best aesthetic quality", Type: "Image", Category: "Aesthetic", UseCases: []string{"Art", "Beauty"}},
		{Name: "Kolors", Size: "~9GB", Description: "Kuaishou's bilingual model", Type: "Image", Category: "Bilingual", UseCases: []string{"Chinese prompts", "Art"}, HFLink: "https://huggingface.co/Kwai-Kolors/Kolors"},
		{Name: "Hunyuan-DiT", Size: "~8GB", Description: "Tencent's bilingual model", Type: "Image", Category: "Bilingual", UseCases: []string{"Chinese", "Quality"}, HFLink: "https://huggingface.co/Tencent-Hunyuan/HunyuanDiT"},
		{Name: "PixArt-Sigma", Size: "~3GB", Description: "Efficient 4K generation", Type: "Image", Category: "Efficient", UseCases: []string{"High-res", "Fast"}, HFLink: "https://huggingface.co/PixArt-alpha/PixArt-Sigma-XL-2-1024-MS"},
		{Name: "Kandinsky 3.0", Size: "~6GB", Description: "Sber's latest, good quality", Type: "Image", Category: "Alternative", UseCases: []string{"Art", "Creativity"}, HFLink: "https://huggingface.co/kandinsky-community/kandinsky-3"},
		{Name: "Würstchen", Size: "~2GB", Description: "Extremely efficient, small", Type: "Image", Category: "Efficient", UseCases: []string{"Fast", "Low VRAM"}, HFLink: "https://huggingface.co/warp-ai/wuerstchen"},

		// =====================================================================
		// VIDEO GENERATION
		// =====================================================================
		{Name: "Wan2.1 14B", Size: "~15GB", Description: "State-of-the-art image-to-video", Type: "Video", Category: "I2V", UseCases: []string{"Professional video", "Animation"}, HFLink: "https://huggingface.co/Wan-AI/Wan2.1-I2V-14B-480P"},
		{Name: "Wan2.1 1.3B", Size: "~3GB", Description: "Efficient Wan video model", Type: "Video", Category: "Efficient", UseCases: []string{"Quick video", "Preview"}},
		{Name: "CogVideoX-5B", Size: "~14GB", Description: "Zhipu's text-to-video", Type: "Video", Category: "T2V", UseCases: []string{"Text to video", "Creativity"}, HFLink: "https://huggingface.co/THUDM/CogVideoX-5b"},
		{Name: "CogVideoX-2B", Size: "~6GB", Description: "Smaller CogVideoX", Type: "Video", Category: "Efficient", UseCases: []string{"Fast T2V"}},
		{Name: "Mochi-1 Preview", Size: "~12GB", Description: "Genmo's 10B video model", Type: "Video", Category: "T2V", UseCases: []string{"Creative videos"}, HFLink: "https://huggingface.co/genmo/mochi-1-preview"},
		{Name: "HunyuanVideo", Size: "~20GB", Description: "Tencent's open video model", Type: "Video", Category: "T2V", UseCases: []string{"High quality", "Research"}, HFLink: "https://huggingface.co/tencent/HunyuanVideo"},
		{Name: "LTX-Video", Size: "~7GB", Description: "Lightricks's fast video", Type: "Video", Category: "Fast", UseCases: []string{"Quick generation"}, HFLink: "https://huggingface.co/Lightricks/LTX-Video"},
		{Name: "Stable Video Diffusion", Size: "~10GB", Description: "Stability's img2vid", Type: "Video", Category: "I2V", UseCases: []string{"Animation", "Product"}, HFLink: "https://huggingface.co/stabilityai/stable-video-diffusion-img2vid-xt"},
		{Name: "AnimateDiff v3", Size: "~4GB", Description: "Motion module for SD", Type: "Video", Category: "Animation", UseCases: []string{"Character animation", "Loops"}},
		{Name: "AnimateDiff Lightning", Size: "~4GB", Description: "Fast AnimateDiff", Type: "Video", Category: "Fast", UseCases: []string{"Quick animation"}},
		{Name: "Open-Sora 1.2", Size: "~15GB", Description: "Open-source Sora-like", Type: "Video", Category: "T2V", UseCases: []string{"Long videos", "Research"}, HFLink: "https://huggingface.co/hpcai-tech/Open-Sora"},
		{Name: "VideoCrafter2", Size: "~8GB", Description: "Text and image to video", Type: "Video", Category: "Multi", UseCases: []string{"T2V", "I2V"}},

		// =====================================================================
		// ENHANCEMENT (Upscaling, Restoration, Interpolation)
		// =====================================================================
		{Name: "Real-ESRGAN x4", Size: "~200MB", Description: "Best general upscaler", Type: "Enhance", Category: "Upscale", UseCases: []string{"4x upscaling", "Photos"}},
		{Name: "Real-ESRGAN x2", Size: "~200MB", Description: "2x version, faster", Type: "Enhance", Category: "Upscale", UseCases: []string{"2x upscaling"}},
		{Name: "Real-ESRGAN Anime", Size: "~200MB", Description: "Optimized for anime/art", Type: "Enhance", Category: "Upscale", UseCases: []string{"Anime", "Art upscaling"}},
		{Name: "SwinIR", Size: "~100MB", Description: "Transformer-based upscaler", Type: "Enhance", Category: "Upscale", UseCases: []string{"Quality upscaling"}},
		{Name: "HAT", Size: "~150MB", Description: "Hybrid attention upscaler", Type: "Enhance", Category: "Upscale", UseCases: []string{"High quality"}},
		{Name: "GFPGAN 1.4", Size: "~350MB", Description: "Face restoration", Type: "Enhance", Category: "Face", UseCases: []string{"Face restoration", "Old photos"}},
		{Name: "CodeFormer", Size: "~400MB", Description: "Advanced face restoration", Type: "Enhance", Category: "Face", UseCases: []string{"Face enhancement", "Damage repair"}},
		{Name: "RIFE 4.6", Size: "~200MB", Description: "Real-time frame interpolation", Type: "Enhance", Category: "Interpolation", UseCases: []string{"Slow motion", "FPS boost"}},
		{Name: "RIFE NCNN", Size: "~150MB", Description: "RIFE for Vulkan/NCNN", Type: "Enhance", Category: "Interpolation", UseCases: []string{"AMD/Intel GPUs"}},
		{Name: "FILM", Size: "~300MB", Description: "Google's frame interpolation", Type: "Enhance", Category: "Interpolation", UseCases: []string{"Smooth interpolation"}},
		{Name: "Video2X", Size: "~500MB", Description: "Video upscaling suite", Type: "Enhance", Category: "Video Upscale", UseCases: []string{"Video upscaling"}},

		// =====================================================================
		// CONTROLNET & ADAPTERS
		// =====================================================================
		{Name: "ControlNet Canny", Size: "~1.5GB", Description: "Edge detection control", Type: "Control", Category: "Edge", UseCases: []string{"Line art", "Edges"}},
		{Name: "ControlNet Depth", Size: "~1.5GB", Description: "Depth map control", Type: "Control", Category: "Depth", UseCases: []string{"3D composition", "Depth"}},
		{Name: "ControlNet OpenPose", Size: "~1.5GB", Description: "Human pose control", Type: "Control", Category: "Pose", UseCases: []string{"Character poses", "Bodies"}},
		{Name: "ControlNet Scribble", Size: "~1.5GB", Description: "Sketch/scribble control", Type: "Control", Category: "Sketch", UseCases: []string{"Rough sketches", "Doodles"}},
		{Name: "ControlNet Lineart", Size: "~1.5GB", Description: "Clean line art control", Type: "Control", Category: "Lines", UseCases: []string{"Line drawings", "Comics"}},
		{Name: "ControlNet Softedge", Size: "~1.5GB", Description: "Soft edge detection", Type: "Control", Category: "Edge", UseCases: []string{"Soft edges", "Painting"}},
		{Name: "ControlNet Normal", Size: "~1.5GB", Description: "Normal map control", Type: "Control", Category: "3D", UseCases: []string{"3D textures", "Surfaces"}},
		{Name: "ControlNet Seg", Size: "~1.5GB", Description: "Segmentation control", Type: "Control", Category: "Segmentation", UseCases: []string{"Region control", "Areas"}},
		{Name: "ControlNet Tile", Size: "~1.5GB", Description: "Tile/upscale control", Type: "Control", Category: "Upscale", UseCases: []string{"Detail enhancement", "Upscaling"}},
		{Name: "ControlNet Inpaint", Size: "~1.5GB", Description: "Inpainting control", Type: "Control", Category: "Inpaint", UseCases: []string{"Inpainting", "Editing"}},
		{Name: "IP-Adapter", Size: "~100MB", Description: "Image prompt adapter", Type: "Control", Category: "Adapter", UseCases: []string{"Style transfer", "References"}, HFLink: "https://huggingface.co/h94/IP-Adapter"},
		{Name: "IP-Adapter FaceID", Size: "~200MB", Description: "Face ID preservation", Type: "Control", Category: "Face", UseCases: []string{"Face consistency", "Portraits"}},
		{Name: "IP-Adapter Plus", Size: "~150MB", Description: "Enhanced IP-Adapter", Type: "Control", Category: "Adapter", UseCases: []string{"Better style transfer"}},
		{Name: "InstantID", Size: "~2GB", Description: "Zero-shot face ID", Type: "Control", Category: "Face", UseCases: []string{"Identity preservation", "Portraits"}, HFLink: "https://huggingface.co/InstantX/InstantID"},
		{Name: "PhotoMaker", Size: "~1GB", Description: "Personalized generation", Type: "Control", Category: "Face", UseCases: []string{"Custom faces", "Characters"}},
		{Name: "ControlNet SDXL", Size: "~2.5GB", Description: "SDXL-specific controls", Type: "Control", Category: "SDXL", UseCases: []string{"SDXL guidance"}},
		{Name: "T2I-Adapter", Size: "~500MB", Description: "Lightweight adapters", Type: "Control", Category: "Adapter", UseCases: []string{"Efficient control", "Low VRAM"}},
	}
}

// ============================================================================
// SERVERS TAB
// ============================================================================

type serversTabModel struct {
	servers     []config.Server
	currentIdx  int
	width       int
	height      int
	cfg         *config.Config
	showDetails bool
	action      string // "", "ssh", "push", "pull"
}

func newServersTabModel(cfg *config.Config) serversTabModel {
	servers := []config.Server{}
	if cfg != nil {
		servers = cfg.Servers
	}
	return serversTabModel{
		servers: servers,
		cfg:     cfg,
	}
}

func (m serversTabModel) Init() tea.Cmd {
	return nil
}

func (m serversTabModel) Update(msg tea.Msg) (serversTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if m.currentIdx < len(m.servers)-1 {
				m.currentIdx++
			}
		case "d":
			m.showDetails = !m.showDetails
		case "s":
			m.action = "ssh"
		case "p":
			m.action = "push"
		case "P":
			m.action = "pull"
		case "enter":
			if len(m.servers) > 0 && m.currentIdx < len(m.servers) {
				// SSH into the selected server
				server := m.servers[m.currentIdx]
				m.action = fmt.Sprintf("Connecting to %s...", server.Name)
				return m, tea.Sequence(
					tea.ExitAltScreen,
					func() tea.Msg {
						sshTarget := fmt.Sprintf("%s@%s", server.User, server.Host)
						sshCmd := exec.Command("ssh", sshTarget)
						sshCmd.Stdin = os.Stdin
						sshCmd.Stdout = os.Stdout
						sshCmd.Stderr = os.Stderr
						sshCmd.Run()
						return nil
					},
					tea.EnterAltScreen,
				)
			}
		}
	}
	return m, nil
}

func (m serversTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  🖥️  Configured Servers"))
	s.WriteString("\n\n")

	if len(m.servers) == 0 {
		s.WriteString(dDescStyle.Render("    No servers configured. Use 'anime add <name> <host>' to add servers."))
		s.WriteString("\n")
	} else {
		for i, server := range m.servers {
			if i == m.currentIdx {
				s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", server.Name)))
				s.WriteString(dDetailStyle.Render(fmt.Sprintf(" @ %s", server.Host)))
				s.WriteString("\n")

				if m.showDetails {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      User: %s", server.User)))
					s.WriteString("\n")
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      SSH Key: %s", server.SSHKey)))
					s.WriteString("\n")
					if server.CostPerHour > 0 {
						s.WriteString(dDescStyle.Render(fmt.Sprintf("      Cost: $%.2f/hr", server.CostPerHour)))
						s.WriteString("\n")
					}
					if len(server.Modules) > 0 {
						s.WriteString(dDescStyle.Render(fmt.Sprintf("      Modules: %s", strings.Join(server.Modules, ", "))))
						s.WriteString("\n")
					}
				}
			} else {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", server.Name)))
				s.WriteString(dSubtitleStyle.Render(fmt.Sprintf(" @ %s", server.Host)))
				s.WriteString("\n")
			}
		}
	}

	if m.action != "" {
		s.WriteString("\n")
		s.WriteString(dInfoStyle.Render(fmt.Sprintf("  Action: %s", m.action)))
	}

	s.WriteString("\n\n")
	s.WriteString(dCategoryStyle.Render("  Quick Actions"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    s - SSH to server"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    p - Push CLI to server"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    P - Pull files from server"))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  d: Toggle details | s/p/P: Actions | Enter: Execute"))

	return s.String()
}

// ============================================================================
// FILES TAB
// ============================================================================

type fileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
	Mode    string
}

type filesTabModel struct {
	// Navigation
	currentPath string
	entries     []fileEntry
	currentIdx  int
	width       int
	height      int
	cfg         *config.Config

	// Selection & operations
	selected    map[string]bool
	clipboard   []string
	clipboardOp string // "copy" or "cut"
	showHidden  bool

	// Server selection
	servers       []string
	currentServer int // 0 = local, 1+ = remote servers
	isLoading     bool
	errorMsg      string

	// View options
	showDetails bool
	sortBy      string // "name", "size", "date"
	sortAsc     bool
}

func newFilesTabModel(cfg *config.Config) filesTabModel {
	servers := []string{"local"}
	if cfg != nil {
		for _, s := range cfg.Servers {
			servers = append(servers, s.Name)
		}
	}

	m := filesTabModel{
		currentPath:   getHomeDir(),
		selected:      make(map[string]bool),
		servers:       servers,
		currentServer: 0,
		showHidden:    false,
		sortBy:        "name",
		sortAsc:       true,
		cfg:           cfg,
	}

	m.loadDirectory()
	return m
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}

func (m *filesTabModel) loadDirectory() {
	m.entries = []fileEntry{}
	m.errorMsg = ""

	if m.currentServer == 0 {
		// Local filesystem
		entries, err := os.ReadDir(m.currentPath)
		if err != nil {
			m.errorMsg = err.Error()
			return
		}

		for _, e := range entries {
			if !m.showHidden && strings.HasPrefix(e.Name(), ".") {
				continue
			}

			info, err := e.Info()
			if err != nil {
				continue
			}

			entry := fileEntry{
				Name:    e.Name(),
				Path:    filepath.Join(m.currentPath, e.Name()),
				IsDir:   e.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format("Jan 02 15:04"),
				Mode:    info.Mode().String(),
			}
			m.entries = append(m.entries, entry)
		}

		// Sort entries (directories first, then by name)
		sort.Slice(m.entries, func(i, j int) bool {
			if m.entries[i].IsDir != m.entries[j].IsDir {
				return m.entries[i].IsDir
			}
			return strings.ToLower(m.entries[i].Name) < strings.ToLower(m.entries[j].Name)
		})
	} else {
		// Remote server - would use SSH
		m.entries = []fileEntry{
			{Name: "(Remote browsing - use 's' to SSH)", IsDir: false},
		}
	}
}

func (m filesTabModel) Init() tea.Cmd {
	return nil
}

func (m filesTabModel) Update(msg tea.Msg) (filesTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if m.currentIdx < len(m.entries)-1 {
				m.currentIdx++
			}
		case "enter", "l":
			if m.currentIdx < len(m.entries) {
				entry := m.entries[m.currentIdx]
				if entry.IsDir {
					m.currentPath = entry.Path
					m.currentIdx = 0
					m.loadDirectory()
				}
			}
		case "backspace", "h":
			// Go up one directory
			parent := filepath.Dir(m.currentPath)
			if parent != m.currentPath {
				m.currentPath = parent
				m.currentIdx = 0
				m.loadDirectory()
			}
		case " ":
			// Toggle selection
			if m.currentIdx < len(m.entries) {
				path := m.entries[m.currentIdx].Path
				m.selected[path] = !m.selected[path]
				if !m.selected[path] {
					delete(m.selected, path)
				}
			}
		case "a":
			// Select all
			for _, entry := range m.entries {
				m.selected[entry.Path] = true
			}
		case "A":
			// Deselect all
			m.selected = make(map[string]bool)
		case "c":
			// Copy to clipboard
			m.clipboard = []string{}
			for path := range m.selected {
				m.clipboard = append(m.clipboard, path)
			}
			if len(m.clipboard) == 0 && m.currentIdx < len(m.entries) {
				m.clipboard = []string{m.entries[m.currentIdx].Path}
			}
			m.clipboardOp = "copy"
		case "x":
			// Cut to clipboard
			m.clipboard = []string{}
			for path := range m.selected {
				m.clipboard = append(m.clipboard, path)
			}
			if len(m.clipboard) == 0 && m.currentIdx < len(m.entries) {
				m.clipboard = []string{m.entries[m.currentIdx].Path}
			}
			m.clipboardOp = "cut"
		case "v":
			// Paste from clipboard
			if len(m.clipboard) > 0 {
				for _, src := range m.clipboard {
					dst := filepath.Join(m.currentPath, filepath.Base(src))
					if m.clipboardOp == "copy" {
						dashCopyFileOrDir(src, dst)
					} else if m.clipboardOp == "cut" {
						os.Rename(src, dst)
					}
				}
				if m.clipboardOp == "cut" {
					m.clipboard = []string{}
					m.clipboardOp = ""
				}
				m.selected = make(map[string]bool)
				m.loadDirectory()
			}
		case "D":
			// Delete selected
			for path := range m.selected {
				os.RemoveAll(path)
			}
			m.selected = make(map[string]bool)
			m.loadDirectory()
			if m.currentIdx >= len(m.entries) && len(m.entries) > 0 {
				m.currentIdx = len(m.entries) - 1
			}
		case "n":
			// New directory (would need input mode)
		case "r":
			// Refresh
			m.loadDirectory()
		case ".":
			// Toggle hidden files
			m.showHidden = !m.showHidden
			m.loadDirectory()
		case "d":
			// Toggle details
			m.showDetails = !m.showDetails
		case "~":
			// Go home
			m.currentPath = getHomeDir()
			m.currentIdx = 0
			m.loadDirectory()
		case "S":
			// Switch server
			m.currentServer = (m.currentServer + 1) % len(m.servers)
			m.currentIdx = 0
			m.loadDirectory()
		case "g":
			// Go to top
			m.currentIdx = 0
		case "G":
			// Go to bottom
			if len(m.entries) > 0 {
				m.currentIdx = len(m.entries) - 1
			}
		}
	}
	return m, nil
}

func dashCopyFileOrDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return dashCopyDir(src, dst)
	}
	return dashCopyFile(src, dst)
}

func dashCopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func dashCopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := dashCopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := dashCopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m filesTabModel) View() string {
	var s strings.Builder

	// Server indicator
	serverName := m.servers[m.currentServer]
	serverStyle := dSuccessStyle
	if m.currentServer > 0 {
		serverStyle = dInfoStyle
	}
	s.WriteString("  ")
	s.WriteString(serverStyle.Render(fmt.Sprintf("[%s]", serverName)))
	s.WriteString(" ")
	s.WriteString(dCategoryStyle.Render(m.currentPath))
	s.WriteString("\n\n")

	if m.errorMsg != "" {
		s.WriteString(dErrorStyle.Render(fmt.Sprintf("  Error: %s", m.errorMsg)))
		s.WriteString("\n")
	}

	maxHeight := m.height - 12
	if maxHeight < 5 {
		maxHeight = 5
	}

	// Calculate scroll offset
	startIdx := 0
	if m.currentIdx > maxHeight-3 {
		startIdx = m.currentIdx - maxHeight + 3
	}

	for i := startIdx; i < len(m.entries) && i < startIdx+maxHeight; i++ {
		entry := m.entries[i]

		// Selection indicator
		sel := "  "
		if m.selected[entry.Path] {
			sel = dSuccessStyle.Render("* ")
		}

		// Icon
		icon := "📄"
		if entry.IsDir {
			icon = "📁"
		} else if strings.HasSuffix(entry.Name, ".go") {
			icon = "🔷"
		} else if strings.HasSuffix(entry.Name, ".py") {
			icon = "🐍"
		} else if strings.HasSuffix(entry.Name, ".js") || strings.HasSuffix(entry.Name, ".ts") {
			icon = "🟨"
		} else if strings.HasSuffix(entry.Name, ".md") {
			icon = "📝"
		} else if strings.HasSuffix(entry.Name, ".json") || strings.HasSuffix(entry.Name, ".yaml") || strings.HasSuffix(entry.Name, ".yml") {
			icon = "⚙️"
		} else if strings.HasSuffix(entry.Name, ".sh") {
			icon = "🔧"
		} else if isImageFile(entry.Name) {
			icon = "🖼️"
		} else if isVideoFile(entry.Name) {
			icon = "🎬"
		} else if isArchiveFile(entry.Name) {
			icon = "📦"
		}

		// Size formatting
		sizeStr := ""
		if !entry.IsDir && m.showDetails {
			sizeStr = dashFormatSize(entry.Size)
		}

		if i == m.currentIdx {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("%s> %s %s", sel, icon, entry.Name)))
			if m.showDetails {
				s.WriteString(dDetailStyle.Render(fmt.Sprintf("  %s  %s", sizeStr, entry.ModTime)))
			}
			s.WriteString("\n")
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("%s  %s %s", sel, icon, entry.Name)))
			if m.showDetails && sizeStr != "" {
				s.WriteString(dSubtitleStyle.Render(fmt.Sprintf("  %s", sizeStr)))
			}
			s.WriteString("\n")
		}
	}

	// Status bar
	s.WriteString("\n")
	statusParts := []string{
		fmt.Sprintf("%d items", len(m.entries)),
	}
	if len(m.selected) > 0 {
		statusParts = append(statusParts, dSuccessStyle.Render(fmt.Sprintf("%d selected", len(m.selected))))
	}
	if len(m.clipboard) > 0 {
		statusParts = append(statusParts, dInfoStyle.Render(fmt.Sprintf("%d in clipboard (%s)", len(m.clipboard), m.clipboardOp)))
	}
	s.WriteString(dDescStyle.Render("  " + strings.Join(statusParts, " | ")))

	// Help
	s.WriteString("\n\n")
	s.WriteString(dHelpStyle.Render("  Enter: Open | Space: Select | c/x/v: Copy/Cut/Paste | D: Delete | .: Hidden | S: Server | d: Details"))

	return s.String()
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" || ext == ".svg"
}

func isVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".mkv" || ext == ".webm"
}

func isArchiveFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".zip" || ext == ".tar" || ext == ".gz" || ext == ".tgz" || ext == ".rar" || ext == ".7z"
}

func dashFormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ============================================================================
// SOURCE TAB
// ============================================================================

type sourceTabModel struct {
	repos       []sourceRepo
	currentIdx  int
	width       int
	height      int
	cfg         *config.Config
	showDetails bool
}

type sourceRepo struct {
	Name   string
	Path   string
	Remote string
	Status string
}

func newSourceTabModel(cfg *config.Config) sourceTabModel {
	// Mock repos - would come from actual source tracking
	repos := []sourceRepo{
		{Name: "anime-cli", Path: "/Users/joshkornreich/anime/cli", Remote: "alice:/home/ubuntu/anime/cli", Status: "synced"},
		{Name: "comfyui", Path: "/opt/comfyui", Remote: "alice:/opt/comfyui", Status: "ahead"},
	}
	return sourceTabModel{
		repos: repos,
		cfg:   cfg,
	}
}

func (m sourceTabModel) Init() tea.Cmd {
	return nil
}

func (m sourceTabModel) Update(msg tea.Msg) (sourceTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if m.currentIdx < len(m.repos)-1 {
				m.currentIdx++
			}
		case "d":
			m.showDetails = !m.showDetails
		}
	}
	return m, nil
}

func (m sourceTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  📁 Source Repositories"))
	s.WriteString("\n\n")

	if len(m.repos) == 0 {
		s.WriteString(dDescStyle.Render("    No source repositories linked. Use 'anime source link' to add."))
		s.WriteString("\n")
	} else {
		for i, repo := range m.repos {
			statusIcon := "✓"
			statusStyle := dSuccessStyle
			if repo.Status == "ahead" {
				statusIcon = "↑"
				statusStyle = dWarningStyle
			} else if repo.Status == "behind" {
				statusIcon = "↓"
				statusStyle = dInfoStyle
			}

			if i == m.currentIdx {
				s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", repo.Name)))
				s.WriteString("  ")
				s.WriteString(statusStyle.Render(statusIcon))
				s.WriteString("\n")

				if m.showDetails {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Local:  %s", repo.Path)))
					s.WriteString("\n")
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Remote: %s", repo.Remote)))
					s.WriteString("\n")
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      Status: %s", repo.Status)))
					s.WriteString("\n")
				}
			} else {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", repo.Name)))
				s.WriteString("  ")
				s.WriteString(statusStyle.Render(statusIcon))
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(dCategoryStyle.Render("  Source Commands"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime source push    - Push changes to remote"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime source pull    - Pull changes from remote"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime source sync    - Bidirectional sync"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime source status  - Show sync status"))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  d: Toggle details | Arrows: Navigate"))

	return s.String()
}

// ============================================================================
// CONFIG TAB
// ============================================================================

type configTabModel struct {
	sections    []string
	currentSec  int
	cfg         *config.Config
	width       int
	height      int
	editing     bool
	input       textinput.Model
}

func newConfigTabModel(cfg *config.Config) configTabModel {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."

	return configTabModel{
		sections: []string{"API Keys", "Server Defaults", "Shell Aliases", "Collections", "Users"},
		cfg:      cfg,
		input:    ti,
	}
}

func (m configTabModel) Init() tea.Cmd {
	return nil
}

func (m configTabModel) Update(msg tea.Msg) (configTabModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		m.input, cmd = m.input.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				m.editing = false
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentSec > 0 {
				m.currentSec--
			}
		case "down", "j":
			if m.currentSec < len(m.sections)-1 {
				m.currentSec++
			}
		case "e":
			m.editing = true
			m.input.Focus()
		}
	}
	return m, nil
}

func (m configTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  ⚙️  Configuration"))
	s.WriteString("\n\n")

	for i, sec := range m.sections {
		if i == m.currentSec {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", sec)))
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", sec)))
		}
		s.WriteString("\n")
	}

	// Show section details
	s.WriteString("\n")
	switch m.sections[m.currentSec] {
	case "API Keys":
		s.WriteString(dCategoryStyle.Render("  API Keys"))
		s.WriteString("\n")
		if m.cfg != nil {
			if m.cfg.APIKeys.Anthropic != "" {
				s.WriteString(dItemStyle.Render("    Anthropic: "))
				s.WriteString(dValueStyle.Render("configured"))
			} else {
				s.WriteString(dItemStyle.Render("    Anthropic: "))
				s.WriteString(dSubtitleStyle.Render("not set"))
			}
			s.WriteString("\n")
			if m.cfg.APIKeys.OpenAI != "" {
				s.WriteString(dItemStyle.Render("    OpenAI: "))
				s.WriteString(dValueStyle.Render("configured"))
			} else {
				s.WriteString(dItemStyle.Render("    OpenAI: "))
				s.WriteString(dSubtitleStyle.Render("not set"))
			}
			s.WriteString("\n")
			if m.cfg.APIKeys.HuggingFace != "" {
				s.WriteString(dItemStyle.Render("    HuggingFace: "))
				s.WriteString(dValueStyle.Render("configured"))
			} else {
				s.WriteString(dItemStyle.Render("    HuggingFace: "))
				s.WriteString(dSubtitleStyle.Render("not set"))
			}
			s.WriteString("\n")
			if m.cfg.APIKeys.LambdaLabs != "" {
				s.WriteString(dItemStyle.Render("    Lambda Labs: "))
				s.WriteString(dValueStyle.Render("configured"))
			} else {
				s.WriteString(dItemStyle.Render("    Lambda Labs: "))
				s.WriteString(dSubtitleStyle.Render("not set"))
			}
			s.WriteString("\n")
			s.WriteString("\n")
			s.WriteString(dSubtitleStyle.Render("    (See Keys tab for details)"))
			s.WriteString("\n")
		}

	case "Server Defaults":
		s.WriteString(dCategoryStyle.Render("  Server Defaults"))
		s.WriteString("\n")
		s.WriteString(dItemStyle.Render("    Default server: alice"))
		s.WriteString("\n")
		s.WriteString(dItemStyle.Render("    Default user: ubuntu"))
		s.WriteString("\n")

	case "Shell Aliases":
		s.WriteString(dCategoryStyle.Render("  Shell Aliases"))
		s.WriteString("\n")
		if m.cfg != nil && len(m.cfg.ShellAliases) > 0 {
			for name, cmd := range m.cfg.ShellAliases {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s = %s", name, cmd)))
				s.WriteString("\n")
			}
		} else {
			s.WriteString(dSubtitleStyle.Render("    No shell aliases configured"))
			s.WriteString("\n")
		}

	case "Collections":
		s.WriteString(dCategoryStyle.Render("  Collections"))
		s.WriteString("\n")
		if m.cfg != nil && len(m.cfg.Collections) > 0 {
			for _, col := range m.cfg.Collections {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s (%s): %s", col.Name, col.Type, col.Path)))
				s.WriteString("\n")
			}
		} else {
			s.WriteString(dSubtitleStyle.Render("    No collections configured"))
			s.WriteString("\n")
		}

	case "Users":
		s.WriteString(dCategoryStyle.Render("  Users"))
		s.WriteString("\n")
		if m.cfg != nil {
			if m.cfg.ActiveUser != "" {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    Active: %s", m.cfg.ActiveUser)))
				s.WriteString("\n")
			}
			for _, user := range m.cfg.Users {
				active := ""
				if user.Name == m.cfg.ActiveUser {
					active = " (active)"
				}
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s%s: %s", user.Name, active, user.Path)))
				s.WriteString("\n")
			}
			if len(m.cfg.Users) == 0 {
				s.WriteString(dSubtitleStyle.Render("    No users configured"))
				s.WriteString("\n")
			}
		}
	}

	if m.editing {
		s.WriteString("\n")
		s.WriteString(dInfoStyle.Render("  Editing: "))
		s.WriteString(m.input.View())
	}

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  e: Edit | Arrows: Navigate"))

	return s.String()
}

// ============================================================================
// KEYS TAB
// ============================================================================

type apiKeyInfo struct {
	name     string
	envVar   string
	getValue func(*config.Config) string
	setValue func(*config.Config, string)
	desc     string
}

var apiKeyProviders = []apiKeyInfo{
	{
		name:   "Anthropic",
		envVar: "ANTHROPIC_API_KEY",
		desc:   "Claude API access",
		getValue: func(c *config.Config) string {
			if c == nil {
				return ""
			}
			return c.APIKeys.Anthropic
		},
		setValue: func(c *config.Config, v string) {
			c.APIKeys.Anthropic = v
		},
	},
	{
		name:   "OpenAI",
		envVar: "OPENAI_API_KEY",
		desc:   "GPT models access",
		getValue: func(c *config.Config) string {
			if c == nil {
				return ""
			}
			return c.APIKeys.OpenAI
		},
		setValue: func(c *config.Config, v string) {
			c.APIKeys.OpenAI = v
		},
	},
	{
		name:   "HuggingFace",
		envVar: "HF_TOKEN",
		desc:   "Model downloads & inference",
		getValue: func(c *config.Config) string {
			if c == nil {
				return ""
			}
			return c.APIKeys.HuggingFace
		},
		setValue: func(c *config.Config, v string) {
			c.APIKeys.HuggingFace = v
		},
	},
	{
		name:   "Lambda Labs",
		envVar: "LAMBDA_API_KEY",
		desc:   "GPU cloud instances",
		getValue: func(c *config.Config) string {
			if c == nil {
				return ""
			}
			return c.APIKeys.LambdaLabs
		},
		setValue: func(c *config.Config, v string) {
			c.APIKeys.LambdaLabs = v
		},
	},
}

type keysTabModel struct {
	currentIdx int
	width      int
	height     int
	cfg        *config.Config
	editing    bool
	input      textinput.Model
}

func newKeysTabModel(cfg *config.Config) keysTabModel {
	ti := textinput.New()
	ti.Placeholder = "Enter API key..."
	ti.CharLimit = 256
	ti.Width = 60

	return keysTabModel{
		cfg:   cfg,
		input: ti,
	}
}

func (m keysTabModel) Init() tea.Cmd {
	return nil
}

func (m keysTabModel) Update(msg tea.Msg) (keysTabModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// Save the key
				if m.cfg != nil && m.currentIdx < len(apiKeyProviders) {
					apiKeyProviders[m.currentIdx].setValue(m.cfg, m.input.Value())
					m.cfg.Save()
				}
				m.editing = false
				m.input.Blur()
				return m, nil
			case "esc":
				m.editing = false
				m.input.Blur()
				return m, nil
			}
		}
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.currentIdx < len(apiKeyProviders)-1 {
				m.currentIdx++
			}
		case "k", "up":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "e", "enter":
			// Edit selected key
			m.editing = true
			if m.cfg != nil && m.currentIdx < len(apiKeyProviders) {
				m.input.SetValue(apiKeyProviders[m.currentIdx].getValue(m.cfg))
			}
			m.input.Focus()
		case "d":
			// Delete/clear selected key
			if m.cfg != nil && m.currentIdx < len(apiKeyProviders) {
				apiKeyProviders[m.currentIdx].setValue(m.cfg, "")
				m.cfg.Save()
			}
		}
	}
	return m, nil
}

func (m keysTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  🔑  API Keys"))
	s.WriteString("\n\n")

	for i, key := range apiKeyProviders {
		// Indicator
		indicator := "  "
		if i == m.currentIdx {
			indicator = "> "
		}

		// Get value and mask it
		value := key.getValue(m.cfg)
		status := ""
		maskedValue := ""

		if value != "" {
			status = dSuccessStyle.Render("configured")
			// Show masked key: first 4 chars + ... + last 4 chars
			if len(value) > 12 {
				maskedValue = value[:4] + "..." + value[len(value)-4:]
			} else if len(value) > 4 {
				maskedValue = value[:2] + "..." + value[len(value)-2:]
			} else {
				maskedValue = "****"
			}
		} else {
			status = dSubtitleStyle.Render("not set")
			// Check environment variable
			if envVal := os.Getenv(key.envVar); envVal != "" {
				status = dInfoStyle.Render("from env")
				if len(envVal) > 12 {
					maskedValue = envVal[:4] + "..." + envVal[len(envVal)-4:]
				} else {
					maskedValue = "****"
				}
			}
		}

		// Format line
		if i == m.currentIdx {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("%s%-12s", indicator, key.name)))
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("%s%-12s", indicator, key.name)))
		}
		s.WriteString("  ")
		s.WriteString(status)
		if maskedValue != "" {
			s.WriteString("  ")
			s.WriteString(dValueStyle.Render(maskedValue))
		}
		s.WriteString("\n")
		s.WriteString(dDescStyle.Render(fmt.Sprintf("    %s  (%s)", key.desc, key.envVar)))
		s.WriteString("\n")
	}

	// Edit input
	if m.editing {
		s.WriteString("\n")
		s.WriteString(dInfoStyle.Render("  Enter key: "))
		s.WriteString(m.input.View())
		s.WriteString("\n")
		s.WriteString(dHelpStyle.Render("  Enter: Save | Esc: Cancel"))
	} else {
		s.WriteString("\n")
		s.WriteString(dHelpStyle.Render("  e/Enter: Edit | d: Delete | Arrows: Navigate"))
	}

	s.WriteString("\n\n")
	s.WriteString(dSubtitleStyle.Render("  Keys are saved to ~/.config/anime/config.yaml"))
	s.WriteString("\n")
	s.WriteString(dSubtitleStyle.Render("  Environment variables take precedence if config is not set"))

	return s.String()
}

// ============================================================================
// ALIASES TAB
// ============================================================================

type aliasesTabModel struct {
	aliases     map[string]string
	aliasNames  []string
	currentIdx  int
	width       int
	height      int
	cfg         *config.Config
	showEmbedded bool
}

func newAliasesTabModel(cfg *config.Config) aliasesTabModel {
	var aliases map[string]string
	if cfg != nil {
		aliases = cfg.ListAliases()
	}
	if aliases == nil {
		aliases = make(map[string]string)
	}

	// Sort alias names
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	return aliasesTabModel{
		aliases:    aliases,
		aliasNames: names,
		cfg:        cfg,
	}
}

func (m aliasesTabModel) Init() tea.Cmd {
	return nil
}

func (m aliasesTabModel) Update(msg tea.Msg) (aliasesTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if m.currentIdx < len(m.aliasNames)-1 {
				m.currentIdx++
			}
		case "e":
			m.showEmbedded = !m.showEmbedded
		case "enter":
			// SSH to the selected alias
			if m.currentIdx < len(m.aliasNames) {
				alias := m.aliasNames[m.currentIdx]
				target := m.aliases[alias]
				if target != "" {
					return m, tea.Sequence(
						tea.ExitAltScreen,
						func() tea.Msg {
							sshTarget := target
							if !strings.Contains(target, "@") {
								sshTarget = "ubuntu@" + target
							}
							sshCmd := exec.Command("ssh", sshTarget)
							sshCmd.Stdin = os.Stdin
							sshCmd.Stdout = os.Stdout
							sshCmd.Stderr = os.Stderr
							sshCmd.Run()
							return nil
						},
						tea.EnterAltScreen,
					)
				}
			}
		}
	}
	return m, nil
}

func (m aliasesTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  🔗 Registered Aliases"))
	s.WriteString("\n\n")

	if len(m.aliasNames) == 0 {
		s.WriteString(dDescStyle.Render("    No aliases registered."))
		s.WriteString("\n")
	} else {
		maxHeight := m.height - 10
		if maxHeight < 5 {
			maxHeight = 5
		}

		// Calculate scroll offset
		startIdx := 0
		if m.currentIdx > maxHeight-3 {
			startIdx = m.currentIdx - maxHeight + 3
		}

		for i := startIdx; i < len(m.aliasNames) && i < startIdx+maxHeight; i++ {
			name := m.aliasNames[i]
			value := m.aliases[name]

			// Check if embedded
			isEmbedded := m.cfg != nil && m.cfg.IsEmbeddedAlias(name)
			embeddedTag := ""
			if isEmbedded {
				embeddedTag = dSubtitleStyle.Render(" [embedded]")
			}

			if i == m.currentIdx {
				s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", name)))
				s.WriteString(embeddedTag)
				s.WriteString("\n")
				s.WriteString(dDescStyle.Render(fmt.Sprintf("      -> %s", value)))
				s.WriteString("\n")
			} else {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", name)))
				s.WriteString(embeddedTag)
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(dCategoryStyle.Render("  Alias Commands"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime alias <name> <target>  - Create alias"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime aliases list           - List all aliases"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime aliases install        - Install to shell"))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  e: Toggle embedded | Arrows: Navigate"))

	return s.String()
}

// ============================================================================
// CONTENTS TAB
// ============================================================================

type contentsTabModel struct {
	categories  []string
	contents    map[string][]contentItem
	currentCat  int
	currentIdx  int
	width       int
	height      int
	showDetails bool
}

type contentItem struct {
	Name        string
	Description string
	Type        string
}

func newContentsTabModel() contentsTabModel {
	categories := []string{"Agents", "Commands", "Reel Configs", "Sky Procedures"}

	contents := map[string][]contentItem{
		"Agents": {
			{Name: "architect", Description: "System architecture design specialist", Type: "agent"},
			{Name: "developer", Description: "Comprehensive software development", Type: "agent"},
			{Name: "researcher", Description: "Research and investigation specialist", Type: "agent"},
			{Name: "planner", Description: "Structured implementation planning", Type: "agent"},
			{Name: "designer", Description: "Visual design and UX strategy", Type: "agent"},
			{Name: "analyst", Description: "Multi-domain analysis specialist", Type: "agent"},
			{Name: "inspector", Description: "Quality audit and compliance", Type: "agent"},
			{Name: "optimizer", Description: "Code optimization and performance", Type: "agent"},
			{Name: "refactorer", Description: "Systematic code refactoring", Type: "agent"},
			{Name: "automator", Description: "Workflow and process automation", Type: "agent"},
		},
		"Commands": {
			{Name: "/protocol", Description: "Execute 8-phase MORCHESTRATED protocol", Type: "command"},
			{Name: "/orchestrate", Description: "Development orchestration", Type: "command"},
			{Name: "/gaps", Description: "Detect implementation gaps", Type: "command"},
			{Name: "/resolve", Description: "Autonomous gap resolution", Type: "command"},
			{Name: "/agents", Description: "Manage agent system", Type: "command"},
			{Name: "/status", Description: "Monitor progress", Type: "command"},
			{Name: "/simple", Description: "Force simplest solution", Type: "command"},
			{Name: "/evidence", Description: "Show work and provide evidence", Type: "command"},
		},
		"Reel Configs": {
			{Name: "default", Description: "Default video generation config", Type: "config"},
			{Name: "high-quality", Description: "High quality settings", Type: "config"},
			{Name: "fast", Description: "Fast generation settings", Type: "config"},
		},
		"Sky Procedures": {
			{Name: "doctor", Description: "System diagnosis", Type: "procedure"},
			{Name: "load", Description: "System load analysis", Type: "procedure"},
			{Name: "procedures", Description: "Available procedures", Type: "procedure"},
		},
	}

	return contentsTabModel{
		categories: categories,
		contents:   contents,
	}
}

func (m contentsTabModel) Init() tea.Cmd {
	return nil
}

func (m contentsTabModel) Update(msg tea.Msg) (contentsTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			} else if m.currentCat > 0 {
				m.currentCat--
				items := m.contents[m.categories[m.currentCat]]
				m.currentIdx = len(items) - 1
			}
		case "down", "j":
			items := m.contents[m.categories[m.currentCat]]
			if m.currentIdx < len(items)-1 {
				m.currentIdx++
			} else if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentIdx = 0
			}
		case "left", "h":
			if m.currentCat > 0 {
				m.currentCat--
				m.currentIdx = 0
			}
		case "right", "l":
			if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentIdx = 0
			}
		case "d":
			m.showDetails = !m.showDetails
		}
	}
	return m, nil
}

func (m contentsTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  📦 Embedded Contents"))
	s.WriteString("\n\n")

	maxHeight := m.height - 8
	if maxHeight < 5 {
		maxHeight = 5
	}

	lineCount := 0

	for catIdx, cat := range m.categories {
		if lineCount >= maxHeight {
			break
		}

		items := m.contents[cat]
		emoji := "📁"
		switch cat {
		case "Agents":
			emoji = "🤖"
		case "Commands":
			emoji = "📋"
		case "Reel Configs":
			emoji = "🎬"
		case "Sky Procedures":
			emoji = "⚡"
		}

		if catIdx == m.currentCat {
			s.WriteString(dCategoryStyle.Render(fmt.Sprintf("  %s %s (%d)", emoji, cat, len(items))))
		} else {
			s.WriteString(dSubtitleStyle.Render(fmt.Sprintf("  %s %s (%d)", emoji, cat, len(items))))
		}
		s.WriteString("\n")
		lineCount++

		for itemIdx, item := range items {
			if lineCount >= maxHeight {
				break
			}

			if catIdx == m.currentCat && itemIdx == m.currentIdx {
				s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s", item.Name)))
				s.WriteString("\n")
				lineCount++

				if m.showDetails {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      %s", item.Description)))
					s.WriteString("\n")
					lineCount++
				}
			} else {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s", item.Name)))
				s.WriteString("\n")
				lineCount++
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(dCategoryStyle.Render("  Content Commands"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime claude agents push     - Push agents to ~/.claude"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime claude commands push   - Push commands to ~/.claude"))
	s.WriteString("\n")
	s.WriteString(dItemStyle.Render("    anime contents              - List all embedded content"))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  d: Toggle details | Arrows: Navigate"))

	return s.String()
}

// ============================================================================
// COMMANDS TAB
// ============================================================================

type commandInfo struct {
	Name        string
	Description string
	Category    string
	Subcommands []string
}

type commandsTabModel struct {
	commands    []commandInfo
	categories  []string
	catCommands map[string][]commandInfo
	currentCat  int
	currentIdx  int
	width       int
	height      int
	showDetails bool
}

func newCommandsTabModel() commandsTabModel {
	commands := getCommandList()

	// Group by category
	catCommands := make(map[string][]commandInfo)
	for _, cmd := range commands {
		catCommands[cmd.Category] = append(catCommands[cmd.Category], cmd)
	}

	categories := []string{
		"Server Management",
		"Package Management",
		"Model Management",
		"Content & Collections",
		"Video & Image",
		"Source Control",
		"Claude Integration",
		"Configuration",
		"System",
	}

	return commandsTabModel{
		commands:    commands,
		categories:  categories,
		catCommands: catCommands,
	}
}

func (m commandsTabModel) Init() tea.Cmd {
	return nil
}

func (m commandsTabModel) Update(msg tea.Msg) (commandsTabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			} else if m.currentCat > 0 {
				m.currentCat--
				cmds := m.catCommands[m.categories[m.currentCat]]
				m.currentIdx = len(cmds) - 1
			}
		case "down", "j":
			cmds := m.catCommands[m.categories[m.currentCat]]
			if m.currentIdx < len(cmds)-1 {
				m.currentIdx++
			} else if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentIdx = 0
			}
		case "left", "h":
			if m.currentCat > 0 {
				m.currentCat--
				m.currentIdx = 0
			}
		case "right", "l":
			if m.currentCat < len(m.categories)-1 {
				m.currentCat++
				m.currentIdx = 0
			}
		case "d":
			m.showDetails = !m.showDetails
		case "enter":
			// Show help for selected command
			cmds := m.catCommands[m.categories[m.currentCat]]
			if m.currentIdx < len(cmds) {
				cmd := cmds[m.currentIdx]
				return m, tea.Sequence(
					tea.ExitAltScreen,
					func() tea.Msg {
						helpCmd := exec.Command("anime", cmd.Name, "--help")
						helpCmd.Stdin = os.Stdin
						helpCmd.Stdout = os.Stdout
						helpCmd.Stderr = os.Stderr
						helpCmd.Run()
						fmt.Println("\nPress Enter to return to dashboard...")
						fmt.Scanln()
						return nil
					},
					tea.EnterAltScreen,
				)
			}
		}
	}
	return m, nil
}

func (m commandsTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  📋 CLI Commands"))
	s.WriteString("\n\n")

	maxHeight := m.height - 6
	if maxHeight < 5 {
		maxHeight = 5
	}

	lineCount := 0

	for catIdx, cat := range m.categories {
		if lineCount >= maxHeight {
			break
		}

		cmds := m.catCommands[cat]
		if len(cmds) == 0 {
			continue
		}

		if catIdx == m.currentCat {
			s.WriteString(dCategoryStyle.Render(fmt.Sprintf("  %s (%d)", cat, len(cmds))))
		} else {
			s.WriteString(dSubtitleStyle.Render(fmt.Sprintf("  %s (%d)", cat, len(cmds))))
		}
		s.WriteString("\n")
		lineCount++

		for cmdIdx, cmd := range cmds {
			if lineCount >= maxHeight {
				break
			}

			if catIdx == m.currentCat && cmdIdx == m.currentIdx {
				s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > anime %s", cmd.Name)))
				s.WriteString("\n")
				lineCount++

				if m.showDetails {
					s.WriteString(dDescStyle.Render(fmt.Sprintf("      %s", cmd.Description)))
					s.WriteString("\n")
					lineCount++

					if len(cmd.Subcommands) > 0 {
						s.WriteString(dDescStyle.Render(fmt.Sprintf("      Subcommands: %s", strings.Join(cmd.Subcommands, ", "))))
						s.WriteString("\n")
						lineCount++
					}
				}
			} else {
				s.WriteString(dItemStyle.Render(fmt.Sprintf("    anime %s", cmd.Name)))
				s.WriteString("\n")
				lineCount++
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  d: Toggle details | anime tree: Show all commands"))

	return s.String()
}

func getCommandList() []commandInfo {
	return []commandInfo{
		// Server Management
		{Name: "lambda", Description: "Lambda server operations", Category: "Server Management", Subcommands: []string{"install", "launch", "defaults"}},
		{Name: "add", Description: "Add new server", Category: "Server Management"},
		{Name: "ssh", Description: "SSH to remote server", Category: "Server Management"},
		{Name: "status", Description: "System status", Category: "Server Management"},
		{Name: "list", Description: "List servers", Category: "Server Management"},
		{Name: "push", Description: "Deploy CLI to server", Category: "Server Management"},

		// Package Management
		{Name: "packages", Description: "List all packages", Category: "Package Management"},
		{Name: "install", Description: "Install packages", Category: "Package Management"},
		{Name: "interactive", Description: "TUI package selector", Category: "Package Management"},
		{Name: "library", Description: "Package library browser", Category: "Package Management"},
		{Name: "explore", Description: "Find models on server", Category: "Package Management"},

		// Model Management
		{Name: "models", Description: "Browse AI models", Category: "Model Management"},
		{Name: "ollama", Description: "Ollama management", Category: "Model Management"},
		{Name: "llm", Description: "LLM service", Category: "Model Management"},
		{Name: "query", Description: "Query LLM", Category: "Model Management"},

		// Content & Collections
		{Name: "collection", Description: "Asset collections", Category: "Content & Collections", Subcommands: []string{"create", "list", "delete", "push", "pull"}},
		{Name: "contents", Description: "Browse embedded content", Category: "Content & Collections"},

		// Video & Image
		{Name: "reel", Description: "Video generation", Category: "Video & Image", Subcommands: []string{"prompt", "frames", "resolution", "run"}},
		{Name: "generate", Description: "Generate content", Category: "Video & Image"},
		{Name: "upscale", Description: "Upscale images/videos", Category: "Video & Image"},
		{Name: "animate", Description: "Animation tools", Category: "Video & Image"},

		// Source Control
		{Name: "source", Description: "Source control", Category: "Source Control", Subcommands: []string{"push", "pull", "sync", "link", "status"}},
		{Name: "cpm", Description: "Code Push Manager", Category: "Source Control", Subcommands: []string{"push", "pull", "publish"}},

		// Claude Integration
		{Name: "claude", Description: "Claude Code management", Category: "Claude Integration", Subcommands: []string{"agents", "commands"}},
		{Name: "prompt", Description: "Natural language query", Category: "Claude Integration"},

		// Configuration
		{Name: "config", Description: "Configuration", Category: "Configuration"},
		{Name: "wizard", Description: "Setup wizard", Category: "Configuration"},
		{Name: "aliases", Description: "Manage aliases", Category: "Configuration", Subcommands: []string{"list", "add", "remove", "install"}},

		// System
		{Name: "doctor", Description: "Diagnose issues", Category: "System"},
		{Name: "tree", Description: "Command tree", Category: "System"},
		{Name: "updates", Description: "What's new", Category: "System"},
		{Name: "extract", Description: "Extract source", Category: "System"},
	}
}

// ============================================================================
// REEL TAB
// ============================================================================

type reelTabModel struct {
	settings    []reelSetting
	currentIdx  int
	width       int
	height      int
	editing     bool
	input       textinput.Model
}

type reelSetting struct {
	Name        string
	Value       string
	Description string
	Type        string // "string", "int", "choice"
	Choices     []string
}

func newReelTabModel() reelTabModel {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."

	return reelTabModel{
		settings: []reelSetting{
			{Name: "prompt", Value: "", Description: "Text prompt for video generation", Type: "string"},
			{Name: "frames", Value: "97", Description: "Number of frames (or use --seconds)", Type: "int"},
			{Name: "resolution", Value: "540p", Description: "Video resolution", Type: "choice", Choices: []string{"540p", "720p", "1080p"}},
			{Name: "steps", Value: "30", Description: "Inference steps", Type: "int"},
			{Name: "guidance", Value: "5.0", Description: "CFG guidance scale", Type: "string"},
			{Name: "seed", Value: "-1", Description: "Random seed (-1 for random)", Type: "int"},
			{Name: "output", Value: "./output", Description: "Output directory", Type: "string"},
			{Name: "model", Value: "wan2", Description: "Video model", Type: "choice", Choices: []string{"wan2", "mochi", "cogvideo", "ltxvideo", "svd"}},
			{Name: "image", Value: "", Description: "Input image for img2vid", Type: "string"},
			{Name: "fps", Value: "24", Description: "Output FPS", Type: "int"},
		},
		input: ti,
	}
}

func (m reelTabModel) Init() tea.Cmd {
	return nil
}

func (m reelTabModel) Update(msg tea.Msg) (reelTabModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		m.input, cmd = m.input.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.settings[m.currentIdx].Value = m.input.Value()
				m.editing = false
				m.input.Reset()
			case "esc":
				m.editing = false
				m.input.Reset()
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if m.currentIdx < len(m.settings)-1 {
				m.currentIdx++
			}
		case "enter", "e":
			setting := m.settings[m.currentIdx]
			if setting.Type == "choice" {
				// Cycle through choices
				for i, choice := range setting.Choices {
					if choice == setting.Value {
						m.settings[m.currentIdx].Value = setting.Choices[(i+1)%len(setting.Choices)]
						break
					}
				}
			} else {
				m.editing = true
				m.input.SetValue(setting.Value)
				m.input.Focus()
			}
		case "r":
			// Run generation
		case "c":
			// Copy command
		}
	}
	return m, nil
}

func (m reelTabModel) View() string {
	var s strings.Builder

	s.WriteString(dCategoryStyle.Render("  🎬 Reel Video Generation"))
	s.WriteString("\n\n")

	for i, setting := range m.settings {
		if i == m.currentIdx {
			s.WriteString(dSelectedStyle.Render(fmt.Sprintf("  > %s:", setting.Name)))
			s.WriteString(" ")
			if m.editing {
				s.WriteString(m.input.View())
			} else {
				if setting.Value == "" {
					s.WriteString(dSubtitleStyle.Render("(not set)"))
				} else {
					s.WriteString(dValueStyle.Render(setting.Value))
				}
			}
			s.WriteString("\n")
			s.WriteString(dDescStyle.Render(fmt.Sprintf("      %s", setting.Description)))
			if setting.Type == "choice" {
				s.WriteString(dDescStyle.Render(fmt.Sprintf(" [%s]", strings.Join(setting.Choices, ", "))))
			}
			s.WriteString("\n")
		} else {
			s.WriteString(dItemStyle.Render(fmt.Sprintf("    %s:", setting.Name)))
			s.WriteString(" ")
			if setting.Value == "" {
				s.WriteString(dSubtitleStyle.Render("(not set)"))
			} else {
				s.WriteString(dValueStyle.Render(setting.Value))
			}
			s.WriteString("\n")
		}
	}

	// Generate command preview
	s.WriteString("\n")
	s.WriteString(dCategoryStyle.Render("  Generated Command"))
	s.WriteString("\n")
	s.WriteString(dInfoStyle.Render(m.generateCommand()))
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(dHelpStyle.Render("  Enter/e: Edit | r: Run | c: Copy command"))

	return s.String()
}

func (m reelTabModel) generateCommand() string {
	cmd := "    anime reel"
	for _, setting := range m.settings {
		if setting.Value != "" && setting.Value != "-1" {
			switch setting.Name {
			case "prompt":
				cmd += fmt.Sprintf(" --prompt \"%s\"", setting.Value)
			case "frames":
				cmd += fmt.Sprintf(" --frames %s", setting.Value)
			case "resolution":
				cmd += fmt.Sprintf(" --resolution %s", setting.Value)
			case "steps":
				cmd += fmt.Sprintf(" --steps %s", setting.Value)
			case "guidance":
				cmd += fmt.Sprintf(" --guidance %s", setting.Value)
			case "seed":
				if setting.Value != "-1" {
					cmd += fmt.Sprintf(" --seed %s", setting.Value)
				}
			case "output":
				cmd += fmt.Sprintf(" --output %s", setting.Value)
			case "model":
				cmd += fmt.Sprintf(" --model %s", setting.Value)
			case "image":
				if setting.Value != "" {
					cmd += fmt.Sprintf(" --image %s", setting.Value)
				}
			case "fps":
				cmd += fmt.Sprintf(" --fps %s", setting.Value)
			}
		}
	}
	cmd += " run"
	return cmd
}

// ============================================================================
// KEY BINDINGS
// ============================================================================

type dashKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Select   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Quit     key.Binding
	Help     key.Binding
}

var dashKeys = dashKeyMap{
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
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "select"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// ============================================================================
// RUN DASHBOARD
// ============================================================================

func runDashboard(cmd *cobra.Command, args []string) {
	m := newDashboardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running dashboard: %v\n", err)
		return
	}

	// Check if user pressed Enter to install packages
	dm := finalModel.(dashboardModel)
	if dm.packagesModel != nil && dm.packagesModel.installing && len(dm.packagesModel.selected) > 0 {
		selectedIDs := make([]string, 0, len(dm.packagesModel.selected))
		for id, isSelected := range dm.packagesModel.selected {
			if isSelected {
				selectedIDs = append(selectedIDs, id)
			}
		}

		// Resolve dependencies and install
		resolved, err := installer.ResolveDependencies(selectedIDs)
		if err != nil {
			fmt.Println("Error resolving dependencies: " + err.Error())
			return
		}

		fmt.Printf("\nInstalling %d package(s)...\n", len(resolved))

		// If remote flag is set, use remote install, otherwise local
		if installRemote {
			runRemoteInstall(resolved)
		} else {
			runLocalInstall(resolved)
		}
	}
}

// ============================================================================
// LIST ITEM INTERFACE
// ============================================================================

type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }
