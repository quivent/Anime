package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "Interactive package selection with beautiful TUI",
	Aliases: []string{"i", "tui"},
	Run:     runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

type packageItem struct {
	pkg      *installer.Package
	selected bool
}

type model struct {
	packages     []*installer.Package
	selected     map[string]bool
	cursor       int
	width        int
	height       int
	quitting     bool
	confirmed    bool
	categoryView bool
}

var (
	// Color scheme
	sakuraPink   = lipgloss.Color("205")
	electricBlue = lipgloss.Color("39")
	neonPurple   = lipgloss.Color("141")
	mintGreen    = lipgloss.Color("86")
	dimGray      = lipgloss.Color("243")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(sakuraPink).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimGray).
			Padding(0, 1)

	categoryHeaderStyle = lipgloss.NewStyle().
				Foreground(electricBlue).
				Bold(true).
				Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(mintGreen).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(sakuraPink).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sakuraPink).
			Padding(1, 2).
			Margin(1, 0)

	summaryStyle = lipgloss.NewStyle().
			Foreground(mintGreen).
			Bold(true).
			Padding(1, 2).
			MarginTop(1)
)

func initialModel() model {
	packagesMap := installer.GetPackages()

	// Convert map to slice
	packages := make([]*installer.Package, 0, len(packagesMap))
	for _, pkg := range packagesMap {
		packages = append(packages, pkg)
	}

	return model{
		packages:     packages,
		selected:     make(map[string]bool),
		cursor:       0,
		categoryView: true,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.packages)-1 {
				m.cursor++
			}

		case " ":
			if m.cursor < len(m.packages) {
				pkg := m.packages[m.cursor]
				m.selected[pkg.ID] = !m.selected[pkg.ID]
			}

		case "enter":
			m.confirmed = true
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	title := "⚡ ANIME PACKAGE INSTALLER ⚡"
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	// Instructions
	help := "↑/↓: Navigate  |  SPACE: Toggle  |  ENTER: Install  |  Q: Quit"
	s.WriteString(helpStyle.Render(help))
	s.WriteString("\n\n")

	// Group packages by category
	categories := make(map[string][]*installer.Package)
	categoryOrder := []string{"Foundation", "ML Framework", "LLM Runtime", "Models", "Video Generation", "Application"}

	for _, pkg := range m.packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Render packages by category
	currentIdx := 0
	for _, category := range categoryOrder {
		pkgs := categories[category]
		if len(pkgs) == 0 {
			continue
		}

		// Category header with emoji
		emoji := "📦"
		switch category {
		case "Foundation":
			emoji = "🏗️"
		case "ML Framework":
			emoji = "🤖"
		case "LLM Runtime":
			emoji = "🔮"
		case "Models":
			emoji = "⭐"
		case "Video Generation":
			emoji = "🎬"
		case "Application":
			emoji = "🎯"
		}

		s.WriteString(categoryHeaderStyle.Render(fmt.Sprintf("%s %s", emoji, category)))
		s.WriteString("\n")

		// Render packages in this category
		for _, pkg := range pkgs {
			checkbox := "☐"
			style := dimStyle
			cursor := "  "

			if m.selected[pkg.ID] {
				checkbox = "☑"
				style = selectedStyle
			}

			if currentIdx == m.cursor {
				cursor = "▶ "
				style = cursorStyle
			}

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, pkg.Name)
			s.WriteString(style.Render(line))

			// Add description on same line
			desc := fmt.Sprintf(" - %s", pkg.Description)
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			s.WriteString(dimStyle.Render(desc))
			s.WriteString("\n")

			// Show size and time for cursor position
			if currentIdx == m.cursor {
				details := fmt.Sprintf("    ⏱️  %s  |  💾 %s", pkg.EstimatedTime, pkg.Size)
				s.WriteString(dimStyle.Render(details))
				s.WriteString("\n")
			}

			currentIdx++
		}
		s.WriteString("\n")
	}

	// Summary of selected packages
	if len(m.selected) > 0 {
		selectedIDs := make([]string, 0, len(m.selected))
		for id := range m.selected {
			selectedIDs = append(selectedIDs, id)
		}

		resolved, err := installer.ResolveDependencies(selectedIDs)
		var summary string
		if err == nil {
			totalMinutes := 0
			for _, pkg := range resolved {
				totalMinutes += int(pkg.EstimatedTime.Minutes())
			}
			timeStr := fmt.Sprintf("%dh %dm", totalMinutes/60, totalMinutes%60)
			if totalMinutes < 60 {
				timeStr = fmt.Sprintf("%dm", totalMinutes)
			}

			summary = fmt.Sprintf("✓ Selected: %d package(s) → %d with dependencies  |  ⏱️  %s",
				len(m.selected), len(resolved), timeStr)
		} else {
			summary = fmt.Sprintf("✓ Selected: %d package(s)", len(m.selected))
		}

		s.WriteString("\n")
		s.WriteString(summaryStyle.Render(summary))
		s.WriteString("\n")
	}

	return s.String()
}

func runInteractive(cmd *cobra.Command, args []string) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	finalModel := m.(model)

	if !finalModel.confirmed || len(finalModel.selected) == 0 {
		fmt.Println("Installation cancelled")
		return
	}

	selectedIDs := make([]string, 0, len(finalModel.selected))
	for id := range finalModel.selected {
		selectedIDs = append(selectedIDs, id)
	}

	// Now actually install the selected packages
	// Use the same logic as runInstallNew but skip confirmation since they already confirmed in TUI
	resolved, err := installer.ResolveDependencies(selectedIDs)
	if err != nil {
		fmt.Println("Error resolving dependencies: " + err.Error())
		return
	}

	// If remote flag is set, use remote install, otherwise local
	if installRemote {
		runRemoteInstall(resolved)
	} else {
		runLocalInstall(resolved)
	}
}
