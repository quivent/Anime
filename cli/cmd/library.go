package cmd

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var libraryCmd = &cobra.Command{
	Use:     "library",
	Short:   "Tabbed package browser with beautiful TUI",
	Aliases: []string{"lib", "browse"},
	Long: `Browse all available packages organized by category tabs.

Navigate tabs with Tab/Shift+Tab or ←/→
Navigate packages with ↑/↓ or j/k
Toggle selection with Space
Press Enter to install selected packages
Press Q to quit`,
	Run: runLibrary,
}

func init() {
	rootCmd.AddCommand(libraryCmd)
}

// Tab definitions for the library
type libraryTab struct {
	name       string
	emoji      string
	categories []string
}

var libraryTabs = []libraryTab{
	{name: "Models", emoji: "🎨", categories: []string{"Image Generation", "Video Generation"}},
	{name: "LLMs", emoji: "💬", categories: []string{"LLM", "LLM Runtime"}},
	{name: "Enhancement", emoji: "✨", categories: []string{"Image Enhancement", "Video Enhancement", "ControlNet"}},
	{name: "ComfyUI", emoji: "🔌", categories: []string{"ComfyUI Node", "Application"}},
	{name: "Foundation", emoji: "🏗️", categories: []string{"Foundation", "GPU", "ML Framework"}},
	{name: "Runtime", emoji: "⚙️", categories: []string{"Runtime", "Containers"}},
	{name: "Bundles", emoji: "⭐", categories: []string{"Models"}},
}

type libraryModel struct {
	tabs         []libraryTab
	activeTab    int
	packages     map[string][]*installer.Package // packages by tab
	selected     map[string]bool
	installed    map[string]bool
	cursor       int
	width        int
	height       int
	quitting     bool
	confirmed    bool
	allPackages  map[string]*installer.Package
}

// Styles for library TUI
var (
	libTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236")).
			Bold(true).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("235")).
				Padding(0, 2)

	tabBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			MarginBottom(1)

	libSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	libCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	libDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	libInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("34")).
				Bold(true)

	libCategoryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)

	libSummaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(1, 0)

	libHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Padding(0, 1)
)

func initialLibraryModel() libraryModel {
	allPkgs := installer.GetPackages()
	installed := checkInstalledPackages()

	// Group packages by tab
	tabPackages := make(map[string][]*installer.Package)

	for _, tab := range libraryTabs {
		var pkgs []*installer.Package
		for _, pkg := range allPkgs {
			for _, cat := range tab.categories {
				if pkg.Category == cat {
					pkgs = append(pkgs, pkg)
					break
				}
			}
		}
		// Sort packages alphabetically
		sort.Slice(pkgs, func(i, j int) bool {
			return pkgs[i].Name < pkgs[j].Name
		})
		tabPackages[tab.name] = pkgs
	}

	return libraryModel{
		tabs:        libraryTabs,
		activeTab:   0,
		packages:    tabPackages,
		selected:    make(map[string]bool),
		installed:   installed,
		cursor:      0,
		allPackages: allPkgs,
	}
}

func (m libraryModel) Init() tea.Cmd {
	return nil
}

func (m libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			m.cursor = 0

		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			m.cursor = 0

		case "up", "k":
			pkgs := m.packages[m.tabs[m.activeTab].name]
			if m.cursor > 0 {
				m.cursor--
			} else if len(pkgs) > 0 {
				m.cursor = len(pkgs) - 1
			}

		case "down", "j":
			pkgs := m.packages[m.tabs[m.activeTab].name]
			if m.cursor < len(pkgs)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case " ":
			pkgs := m.packages[m.tabs[m.activeTab].name]
			if m.cursor < len(pkgs) {
				pkg := pkgs[m.cursor]
				m.selected[pkg.ID] = !m.selected[pkg.ID]
			}

		case "enter":
			if len(m.selected) > 0 {
				m.confirmed = true
				m.quitting = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m libraryModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	title := "⚡ ANIME PACKAGE LIBRARY ⚡"
	s.WriteString(libTitleStyle.Render(title))
	s.WriteString("\n\n")

	// Tab bar
	var tabs []string
	for i, tab := range m.tabs {
		tabCount := len(m.packages[tab.name])
		tabLabel := fmt.Sprintf("%s %s (%d)", tab.emoji, tab.name, tabCount)
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(tabLabel))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(tabLabel))
		}
	}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	s.WriteString("\n\n")

	// Current tab packages
	currentTab := m.tabs[m.activeTab]
	pkgs := m.packages[currentTab.name]

	if len(pkgs) == 0 {
		s.WriteString(libDimStyle.Render("  No packages in this category"))
		s.WriteString("\n")
	} else {
		// Group by subcategory
		subcats := make(map[string][]*installer.Package)
		for _, pkg := range pkgs {
			subcats[pkg.Category] = append(subcats[pkg.Category], pkg)
		}

		// Calculate visible area (leave room for header, tabs, help)
		maxVisible := 20
		if m.height > 0 {
			maxVisible = m.height - 12
		}
		if maxVisible < 5 {
			maxVisible = 5
		}

		currentIdx := 0
		linesRendered := 0

		for _, cat := range currentTab.categories {
			catPkgs := subcats[cat]
			if len(catPkgs) == 0 {
				continue
			}

			// Category header
			if linesRendered < maxVisible {
				emoji := getLibraryCategoryEmoji(cat)
				s.WriteString(libCategoryStyle.Render(fmt.Sprintf("  %s %s", emoji, cat)))
				s.WriteString("\n")
				linesRendered++
			}

			// Packages in this subcategory
			for _, pkg := range catPkgs {
				if linesRendered >= maxVisible {
					s.WriteString(libDimStyle.Render(fmt.Sprintf("  ... and %d more", len(pkgs)-currentIdx)))
					s.WriteString("\n")
					break
				}

				checkbox := "☐"
				style := libDimStyle
				cursor := "  "

				if m.selected[pkg.ID] {
					checkbox = "☑"
					style = libSelectedStyle
				}

				if m.installed[pkg.ID] {
					style = libInstalledStyle
					checkbox = "✓"
				}

				if currentIdx == m.cursor {
					cursor = "▶ "
					if !m.installed[pkg.ID] {
						style = libCursorStyle
					}
				}

				// Package line
				line := fmt.Sprintf("%s%s %s", cursor, checkbox, pkg.Name)
				s.WriteString(style.Render(line))

				// Size badge
				s.WriteString(libDimStyle.Render(fmt.Sprintf(" [%s]", pkg.Size)))
				s.WriteString("\n")
				linesRendered++

				// Show description for cursor position
				if currentIdx == m.cursor && linesRendered < maxVisible {
					desc := pkg.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					s.WriteString(libDimStyle.Render(fmt.Sprintf("      %s", desc)))
					s.WriteString("\n")
					linesRendered++

					// Show install command
					if !m.installed[pkg.ID] {
						s.WriteString(libDimStyle.Render(fmt.Sprintf("      $ anime install %s", pkg.ID)))
						s.WriteString("\n")
						linesRendered++
					}
				}

				currentIdx++
			}
			s.WriteString("\n")
			linesRendered++
		}
	}

	// Selection summary
	if len(m.selected) > 0 {
		selectedIDs := make([]string, 0, len(m.selected))
		for id, selected := range m.selected {
			if selected {
				selectedIDs = append(selectedIDs, id)
			}
		}

		if len(selectedIDs) > 0 {
			resolved, err := installer.ResolveDependencies(selectedIDs)
			var summary string
			if err == nil {
				summary = fmt.Sprintf("✓ Selected: %d package(s) → %d with dependencies",
					len(selectedIDs), len(resolved))
			} else {
				summary = fmt.Sprintf("✓ Selected: %d package(s)", len(selectedIDs))
			}
			s.WriteString("\n")
			s.WriteString(libSummaryStyle.Render(summary))
			s.WriteString("\n")
		}
	}

	// Help bar
	s.WriteString("\n")
	help := "Tab/←/→: Switch tabs  |  ↑/↓: Navigate  |  Space: Select  |  Enter: Install  |  Q: Quit"
	s.WriteString(libHelpStyle.Render(help))
	s.WriteString("\n")

	return s.String()
}

func getLibraryCategoryEmoji(cat string) string {
	emojis := map[string]string{
		"Image Generation":  "🎨",
		"Video Generation":  "🎬",
		"Image Enhancement": "✨",
		"Video Enhancement": "📹",
		"ControlNet":        "🎛️",
		"LLM":               "💬",
		"LLM Runtime":       "🔮",
		"Foundation":        "🏗️",
		"GPU":               "🎮",
		"ML Framework":      "🤖",
		"Runtime":           "⚙️",
		"Containers":        "📦",
		"ComfyUI Node":      "🔌",
		"Application":       "🎯",
		"Models":            "⭐",
	}
	if emoji, ok := emojis[cat]; ok {
		return emoji
	}
	return "📦"
}

func runLibrary(cmd *cobra.Command, args []string) {
	p := tea.NewProgram(initialLibraryModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	finalModel := m.(libraryModel)

	if !finalModel.confirmed || len(finalModel.selected) == 0 {
		fmt.Println("No packages selected")
		return
	}

	// Collect selected package IDs
	selectedIDs := make([]string, 0, len(finalModel.selected))
	for id, selected := range finalModel.selected {
		if selected {
			selectedIDs = append(selectedIDs, id)
		}
	}

	// Resolve dependencies
	resolved, err := installer.ResolveDependencies(selectedIDs)
	if err != nil {
		fmt.Println("Error resolving dependencies: " + err.Error())
		return
	}

	// Install packages
	if installRemote {
		runRemoteInstall(resolved)
	} else {
		runLocalInstall(resolved)
	}
}
