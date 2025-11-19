package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive package selection with beautiful TUI",
	Aliases: []string{"i", "tui"},
	Run:   runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

type item struct {
	pkg      *installer.Package
	selected bool
}

func (i item) FilterValue() string { return i.pkg.Name }
func (i item) Title() string {
	checkbox := "☐"
	if i.selected {
		checkbox = "☑"
	}
	return fmt.Sprintf("%s %s", checkbox, i.pkg.Name)
}
func (i item) Description() string {
	return fmt.Sprintf("%s • %s • %s",
		i.pkg.Category,
		i.pkg.EstimatedTime,
		i.pkg.Size)
}

type model struct {
	list     list.Model
	items    []item
	selected map[string]bool
	quitting bool
	confirmed bool
}

func initialModel() model {
	packages := installer.GetPackages()

	items := make([]item, 0, len(packages))
	for _, pkg := range packages {
		items = append(items, item{pkg: pkg})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		BorderLeftForeground(lipgloss.Color("205"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("243")).
		BorderLeftForeground(lipgloss.Color("205"))

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	l := list.New(listItems, delegate, 80, 20)
	l.Title = "📦 Select Packages to Install"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("space"),
				key.WithHelp("space", "toggle"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "confirm"),
			),
		}
	}

	return model{
		list:     l,
		items:    items,
		selected: make(map[string]bool),
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

		case " ":
			// Toggle selection
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected[i.pkg.ID] = !m.selected[i.pkg.ID]
				// Update the item's selected state
				idx := m.list.Index()
				m.items[idx].selected = m.selected[i.pkg.ID]
				m.list.SetItem(idx, m.items[idx])
			}
			return m, nil

		case "enter":
			m.confirmed = true
			m.quitting = true
			return m, tea.Quit

		case "tea.WindowSizeMsg":
			h, v := lipgloss.Size(m.list.View())
			m.list.SetSize(h, v)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var status string
	if len(m.selected) > 0 {
		selectedPkgs := make([]string, 0, len(m.selected))
		for id := range m.selected {
			selectedPkgs = append(selectedPkgs, id)
		}

		totalTime, totalSize := calculateTotals(selectedPkgs)

		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			MarginTop(1).
			MarginLeft(2)

		status = statusStyle.Render(fmt.Sprintf(
			"\n✓ Selected: %d packages • Est. time: %s • Total size: ~%s",
			len(m.selected),
			totalTime,
			totalSize,
		))
	}

	return m.list.View() + status
}

func calculateTotals(packageIDs []string) (string, string) {
	resolved, err := installer.ResolveDependencies(packageIDs)
	if err != nil {
		return "unknown", "unknown"
	}

	var totalMinutes int
	var totalGB float64

	for _, pkg := range resolved {
		totalMinutes += int(pkg.EstimatedTime.Minutes())

		// Parse size
		sizeStr := strings.TrimPrefix(pkg.Size, "~")
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
		var size float64
		if strings.HasSuffix(pkg.Size, "GB") {
			fmt.Sscanf(sizeStr, "%f", &size)
			totalGB += size
		} else if strings.HasSuffix(pkg.Size, "MB") {
			fmt.Sscanf(sizeStr, "%f", &size)
			totalGB += size / 1000
		}
	}

	timeStr := fmt.Sprintf("%dh %dm", totalMinutes/60, totalMinutes%60)
	if totalMinutes < 60 {
		timeStr = fmt.Sprintf("%dm", totalMinutes)
	}

	return timeStr, fmt.Sprintf("%.1fGB", totalGB)
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
