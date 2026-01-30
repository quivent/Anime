package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var contentsCmd = &cobra.Command{
	Use:     "contents",
	Aliases: []string{"embedded", "toc"},
	Short:   "Show everything embedded in the CLI",
	Long:    "Interactive TUI to browse all content embedded in the anime CLI binary.",
	Run:     runContentsTUI,
}

func init() {
	rootCmd.AddCommand(contentsCmd)
}

// Tab definitions
var tabNames = []string{"Claude", "Models", "Packages", "Tooling", "Source"}

// Styles for contents TUI
var (
	cTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#888888"))

	cActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FF00")).
			Bold(true)

	cTabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#444444"))

	cItemStyle = lipgloss.NewStyle().
			Width(24).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FFFFFF"))

	cSelectedItemStyle = lipgloss.NewStyle().
				Width(24).
				Padding(0, 1).
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FFFF")).
				Bold(true)

	cDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	cTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")).
			Bold(true).
			MarginBottom(1)

	cHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			MarginTop(1)
)

type contentsModel struct {
	activeTab   int
	cursor      int
	width       int
	height      int
	tabContents map[int][]string
	cols        int
}

func newContentsModel() contentsModel {
	m := contentsModel{
		activeTab:   0,
		cursor:      0,
		cols:        3,
		tabContents: make(map[int][]string),
	}

	// Load content for each tab
	m.tabContents[0] = getClaudeContent()
	m.tabContents[1] = getModelsContent()
	m.tabContents[2] = getPackagesContent()
	m.tabContents[3] = getToolingContent()
	m.tabContents[4] = getSourceContent()

	return m
}

func (m contentsModel) Init() tea.Cmd {
	return nil
}

func (m contentsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(tabNames)
			m.cursor = 0

		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
			m.cursor = 0

		case "down", "j":
			items := m.tabContents[m.activeTab]
			if m.cursor+m.cols < len(items) {
				m.cursor += m.cols
			}

		case "up", "k":
			if m.cursor-m.cols >= 0 {
				m.cursor -= m.cols
			}

		case "ctrl+right", "L":
			items := m.tabContents[m.activeTab]
			if m.cursor+1 < len(items) && (m.cursor+1)%m.cols != 0 {
				m.cursor++
			}

		case "ctrl+left", "H":
			if m.cursor > 0 && m.cursor%m.cols != 0 {
				m.cursor--
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust columns based on width
		if m.width > 100 {
			m.cols = 4
		} else if m.width > 75 {
			m.cols = 3
		} else {
			m.cols = 2
		}
	}

	return m, nil
}

func (m contentsModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString("\n")
	s.WriteString(cTitleStyle.Render("  📦 ANIME CLI - EMBEDDED CONTENTS"))
	s.WriteString("\n\n")

	// Tab bar
	var tabs []string
	for i, name := range tabNames {
		count := len(m.tabContents[i])
		label := fmt.Sprintf("%s (%d)", name, count)
		if i == m.activeTab {
			tabs = append(tabs, cActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, cTabStyle.Render(label))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	s.WriteString(cTabBarStyle.Render(tabBar))
	s.WriteString("\n\n")

	// Content grid
	items := m.tabContents[m.activeTab]
	if len(items) == 0 {
		s.WriteString(cDimStyle.Render("  (empty)"))
	} else {
		// Calculate visible rows
		maxRows := (m.height - 12) / 2
		if maxRows < 3 {
			maxRows = 3
		}

		startRow := m.cursor / m.cols
		if startRow > maxRows/2 {
			startRow = startRow - maxRows/2
		} else {
			startRow = 0
		}

		for row := startRow; row < startRow+maxRows && row*m.cols < len(items); row++ {
			var rowItems []string
			for col := 0; col < m.cols; col++ {
				idx := row*m.cols + col
				if idx >= len(items) {
					rowItems = append(rowItems, cItemStyle.Render(""))
					continue
				}

				item := items[idx]
				if idx == m.cursor {
					rowItems = append(rowItems, cSelectedItemStyle.Render(truncate(item, 22)))
				} else {
					rowItems = append(rowItems, cItemStyle.Render(truncate(item, 22)))
				}
			}
			s.WriteString("  ")
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
			s.WriteString("\n")
		}
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(cHelpStyle.Render("  ←/→ tabs • ↑/↓ navigate • q quit"))
	s.WriteString("\n")

	return s.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func runContentsTUI(cmd *cobra.Command, args []string) {
	m := newContentsModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

// Content loaders

func getClaudeContent() []string {
	var items []string

	// Agents
	entries, err := embeddedAgents.ReadDir("embedded/agents")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				name := strings.TrimSuffix(e.Name(), ".md")
				items = append(items, "🤖 "+name)
			}
		}
	}

	// Commands
	entries, err = embeddedCommands.ReadDir("embedded/commands")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				name := strings.TrimSuffix(e.Name(), ".md")
				items = append(items, "📝 "+name)
			}
		}
	}

	return items
}

func getModelsContent() []string {
	// From the models catalog in models.go
	models := []string{
		"🧠 Llama 3.3 70B",
		"🧠 Llama 3.3 8B",
		"🧠 Mistral 7B",
		"🧠 Mixtral 8x7B",
		"🧠 Qwen3 235B",
		"🧠 DeepSeek V3",
		"🧠 Phi-3.5 Mini",
		"🎨 SDXL",
		"🎨 SD 1.5",
		"🎨 Flux.1 Dev",
		"🎨 Flux.1 Schnell",
		"🎨 SD 3.5 Large",
		"🎨 Playground v2.5",
		"🎬 Stable Video",
		"🎬 AnimateDiff",
		"🎬 Mochi-1",
		"🎬 CogVideoX",
		"🎬 HunyuanVideo",
		"🎬 Wan2.2",
		"🔧 Real-ESRGAN",
		"🔧 GFPGAN",
		"🔧 RIFE",
		"🎛️ ControlNet",
		"🎛️ IP-Adapter",
		"🎛️ InstantID",
	}
	return models
}

func getPackagesContent() []string {
	pkgs := installer.GetPackages()
	var items []string
	for _, pkg := range pkgs {
		icon := "📦"
		switch pkg.Category {
		case "ai", "llm":
			icon = "🤖"
		case "comfyui":
			icon = "🎨"
		case "development":
			icon = "💻"
		case "system":
			icon = "⚙️"
		}
		items = append(items, icon+" "+pkg.ID)
	}
	return items
}

func getToolingContent() []string {
	return []string{
		"🎬 Sky",
		"🎬 Reel",
		"📡 SSH Manager",
		"☁️  Lambda Cloud",
		"📦 Package Manager",
		"🔄 Source Sync",
		"🏥 Doctor",
		"📊 Metrics",
		"🎯 Wizard",
		"📋 Updates",
	}
}

func getSourceContent() []string {
	if !HasEmbeddedSource() {
		return []string{"(source not embedded)"}
	}

	data, err := SourceTarball.ReadFile("anime-src.tar.gz")
	if err != nil {
		return nil
	}

	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Get top-level directories and key files
	seen := make(map[string]bool)
	var items []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		parts := strings.Split(header.Name, string(filepath.Separator))
		if len(parts) < 2 {
			continue
		}

		// Show first two levels
		var key string
		if len(parts) >= 2 {
			key = filepath.Join(parts[0], parts[1])
		}

		if !seen[key] {
			seen[key] = true
			icon := "📁"
			if header.Typeflag == tar.TypeReg {
				icon = "📄"
				if strings.HasSuffix(key, ".go") {
					icon = "🔷"
				} else if strings.HasSuffix(key, ".md") {
					icon = "📝"
				}
			}
			items = append(items, icon+" "+key)
		}

		if len(items) > 80 {
			break
		}
	}

	return items
}
