package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var studioCmd = &cobra.Command{
	Use:     "studio",
	Aliases: []string{"dashboard", "hub"},
	Short:   "Interactive command center for all anime capabilities",
	Long:    "A comprehensive TUI dashboard showing everything anime can do.",
	Run:     runStudioTUI,
}

func init() {
	rootCmd.AddCommand(studioCmd)
}

// Studio tabs - organized by function
var studioTabs = []string{
	"Claude",
	"Packages",
	"Models",
	"Video",
	"Image",
	"ComfyUI",
	"Inference",
	"Deploy",
	"System",
}

// Styles for studio TUI
var (
	sTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#666666"))

	sActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FF00FF")).
			Bold(true)

	sTabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#333333")).
			MarginBottom(1)

	sItemStyle = lipgloss.NewStyle().
			Width(28).
			Height(3).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444"))

	sSelectedItemStyle = lipgloss.NewStyle().
				Width(28).
				Height(3).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")).
				Foreground(lipgloss.Color("#00FF00"))

	sTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")).
			Bold(true)

	sHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	sCmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))

	sDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
)

type studioItem struct {
	name string
	cmd  string
	desc string
}

type studioModel struct {
	activeTab   int
	cursor      int
	width       int
	height      int
	cols        int
	tabContents map[int][]studioItem
}

func newStudioModel() studioModel {
	m := studioModel{
		activeTab:   0,
		cursor:      0,
		cols:        3,
		tabContents: make(map[int][]studioItem),
	}

	m.tabContents[0] = getStudioClaudeItems()
	m.tabContents[1] = getStudioPackagesItems()
	m.tabContents[2] = getStudioModelsItems()
	m.tabContents[3] = getStudioVideoItems()
	m.tabContents[4] = getStudioImageItems()
	m.tabContents[5] = getStudioComfyUIItems()
	m.tabContents[6] = getStudioInferenceItems()
	m.tabContents[7] = getStudioDeployItems()
	m.tabContents[8] = getStudioSystemItems()

	return m
}

func (m studioModel) Init() tea.Cmd {
	return nil
}

func (m studioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "L":
			m.activeTab = (m.activeTab + 1) % len(studioTabs)
			m.cursor = 0

		case "shift+tab", "H":
			m.activeTab = (m.activeTab - 1 + len(studioTabs)) % len(studioTabs)
			m.cursor = 0

		case "right", "l":
			items := m.tabContents[m.activeTab]
			if m.cursor+1 < len(items) {
				m.cursor++
			}

		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			items := m.tabContents[m.activeTab]
			if m.cursor+m.cols < len(items) {
				m.cursor += m.cols
			}

		case "up", "k":
			if m.cursor-m.cols >= 0 {
				m.cursor -= m.cols
			}

		case "enter":
			// Could execute command here
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width > 120 {
			m.cols = 4
		} else if m.width > 90 {
			m.cols = 3
		} else {
			m.cols = 2
		}
	}

	return m, nil
}

func (m studioModel) View() string {
	var s strings.Builder

	// Header
	s.WriteString("\n")
	s.WriteString(sTitleStyle.Render("  ⚡ ANIME STUDIO ⚡"))
	s.WriteString("  ")
	s.WriteString(sDescStyle.Render("Command Center"))
	s.WriteString("\n\n")

	// Tab bar
	var tabs []string
	for i, name := range studioTabs {
		if i == m.activeTab {
			tabs = append(tabs, sActiveTabStyle.Render(name))
		} else {
			tabs = append(tabs, sTabStyle.Render(name))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	s.WriteString(sTabBarStyle.Render(tabBar))
	s.WriteString("\n")

	// Content grid
	items := m.tabContents[m.activeTab]
	if len(items) == 0 {
		s.WriteString(sDescStyle.Render("  (no items)"))
	} else {
		maxRows := (m.height - 10) / 4
		if maxRows < 2 {
			maxRows = 2
		}

		startRow := (m.cursor / m.cols) - maxRows/2
		if startRow < 0 {
			startRow = 0
		}

		for row := startRow; row < startRow+maxRows && row*m.cols < len(items); row++ {
			var rowItems []string
			for col := 0; col < m.cols; col++ {
				idx := row*m.cols + col
				if idx >= len(items) {
					break
				}

				item := items[idx]
				content := fmt.Sprintf("%s\n%s",
					truncateStudio(item.name, 24),
					sDescStyle.Render(truncateStudio(item.desc, 24)))

				if idx == m.cursor {
					rowItems = append(rowItems, sSelectedItemStyle.Render(content))
				} else {
					rowItems = append(rowItems, sItemStyle.Render(content))
				}
			}
			s.WriteString("  ")
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
			s.WriteString("\n")
		}
	}

	// Selected item command
	if m.cursor < len(items) {
		item := items[m.cursor]
		s.WriteString("\n")
		s.WriteString("  ")
		s.WriteString(sDescStyle.Render("Run: "))
		s.WriteString(sCmdStyle.Render(item.cmd))
		s.WriteString("\n")
	}

	// Help
	s.WriteString("\n")
	s.WriteString(sHelpStyle.Render("  Tab switch sections • ←↑↓→ navigate • q quit"))
	s.WriteString("\n")

	return s.String()
}

func truncateStudio(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func runStudioTUI(cmd *cobra.Command, args []string) {
	m := newStudioModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

// Content for each tab

func getStudioClaudeItems() []studioItem {
	return []studioItem{
		{"🤖 Push Agents", "anime claude agents push", "Deploy agents to Claude"},
		{"📋 List Agents", "anime claude agents list", "Show embedded agents"},
		{"⬇️  Pull Agents", "anime claude agents pull", "Pull from ~/.claude"},
		{"📝 Push Commands", "anime claude commands push", "Deploy slash commands"},
		{"📋 List Commands", "anime claude commands list", "Show embedded cmds"},
		{"⬇️  Pull Commands", "anime claude commands pull", "Pull from ~/.claude"},
	}
}

func getStudioPackagesItems() []studioItem {
	return []studioItem{
		{"📦 Browse All", "anime packages", "List all packages"},
		{"🎯 Interactive", "anime interactive", "TUI package selector"},
		{"⚡ Install", "anime install <pkg>", "Install a package"},
		{"📊 Status", "anime packages status", "Installation status"},
		{"🌳 Tree View", "anime packages --tree", "Dependency tree"},
		{"🔄 Parallelize", "anime parallelize", "Parallel install plan"},
		{"📚 Library", "anime library", "Browse model library"},
		{"🔍 Explore", "anime explore <server>", "Find models on server"},
	}
}

func getStudioModelsItems() []studioItem {
	return []studioItem{
		{"🎨 Model Catalog", "anime models", "Browse all AI models"},
		{"🧠 LLMs", "anime models --type llm", "Language models"},
		{"🎨 Image Gen", "anime models --type image", "Image generation"},
		{"🎬 Video Gen", "anime models --type video", "Video generation"},
		{"🔧 Upscalers", "anime models --type upscale", "Enhancement models"},
		{"🎛️ ControlNet", "anime models --type control", "Control adapters"},
	}
}

func getStudioVideoItems() []studioItem {
	return []studioItem{
		{"🎬 Reel", "anime reel", "Video pipeline"},
		{"▶️  Generate", "anime reel generate", "Generate video"},
		{"🔄 Interpolate", "anime reel interpolate", "Frame interpolation"},
		{"⬆️  Upscale", "anime reel upscale", "Video upscaling"},
		{"✂️  Sequence", "anime sequence", "Sequence operations"},
		{"🎞️ Animate", "anime animate", "Animation tools"},
	}
}

func getStudioImageItems() []studioItem {
	return []studioItem{
		{"🎨 Generate", "anime generate", "Image generation"},
		{"⬆️  Upscale", "anime upscale", "Image upscaling"},
		{"🖼️ ComfyUI", "anime comfyui", "ComfyUI interface"},
		{"🔧 Workflows", "anime browse-workflows", "Browse workflows"},
		{"📁 Collections", "anime collection", "Asset collections"},
	}
}

func getStudioComfyUIItems() []studioItem {
	return []studioItem{
		// Core Management
		{"🎨 Install ComfyUI", "anime install comfyui", "Install ComfyUI"},
		{"🔧 Install comfy-cli", "anime install comfy-cli", "Install comfy CLI"},
		{"🚀 Launch", "comfy launch", "Start ComfyUI server"},
		{"🌐 Launch Browser", "comfy launch --browser", "Start with browser"},
		{"⚙️  Status", "comfy env show", "Show environment"},

		// Node Management
		{"📦 List Nodes", "comfy node list", "List custom nodes"},
		{"🔍 Search Nodes", "comfy node search", "Search nodes"},
		{"📥 Install Node", "comfy node install", "Install custom node"},
		{"🔄 Update Nodes", "comfy node update all", "Update all nodes"},
		{"🗑️ Remove Node", "comfy node uninstall", "Remove custom node"},

		// Model Management
		{"📋 List Models", "comfy model list", "List all models"},
		{"🔍 Search Models", "comfy model search", "Search models"},
		{"📥 Download Model", "comfy model download", "Download model"},
		{"🗑️ Remove Model", "comfy model remove", "Remove model"},

		// Workflows
		{"📋 List Workflows", "comfy workflow list", "List workflows"},
		{"▶️  Run Workflow", "comfy workflow run", "Run a workflow"},

		// Environment
		{"🔧 Env Install", "comfy env install", "Install env"},
		{"🔄 Env Update", "comfy env update", "Update Python deps"},
		{"📦 Env Packages", "comfy env pip list", "List pip packages"},
	}
}

func getStudioInferenceItems() []studioItem {
	return []studioItem{
		{"🤖 Ollama", "anime ollama", "Ollama management"},
		{"💬 Query", "anime query <model>", "Query LLM"},
		{"🗣️ Prompt", "anime prompt", "Natural language cmd"},
		{"🧠 LLM Status", "anime llm", "LLM service status"},
		{"📡 Start Ollama", "anime start ollama", "Start Ollama server"},
	}
}

func getStudioDeployItems() []studioItem {
	return []studioItem{
		{"🚀 Push CLI", "anime push <server>", "Deploy CLI to server"},
		{"📤 Source Push", "anime source push", "Push code to remote"},
		{"📥 Source Pull", "anime source pull", "Pull code from remote"},
		{"🔗 Source Link", "anime source link", "Link to remote repo"},
		{"📦 Extract", "anime extract --embedded", "Extract source code"},
		{"🚢 Ship", "anime ship <src> <dst>", "Tar + rsync + untar"},
		{"☁️  Lambda", "anime lambda", "Lambda Cloud mgmt"},
		{"🖥️  Add Server", "anime add <name> <host>", "Add SSH server"},
		{"📋 List Servers", "anime list", "Show all servers"},
	}
}

func getStudioSystemItems() []studioItem {
	return []studioItem{
		{"🏥 Doctor", "anime doctor", "Diagnose issues"},
		{"📊 Metrics", "anime metrics", "GPU metrics"},
		{"📋 Status", "anime status", "Server status"},
		{"💰 Billing", "anime billing", "Usage & costs"},
		{"📜 Logs", "anime logs", "View logs"},
		{"⚙️  Config", "anime config", "Configuration"},
		{"🔄 Update", "anime update", "Update CLI"},
		{"📋 Updates", "anime updates", "What's new"},
		{"🎯 Wizard", "anime wizard", "Setup wizard"},
		{"🌳 Tree", "anime tree", "Command tree"},
		{"📦 Contents", "anime contents", "Embedded contents"},
	}
}
