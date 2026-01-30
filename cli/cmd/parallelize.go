package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var parallelizeCmd = &cobra.Command{
	Use:   "parallelize",
	Short: "Interactive package selector for parallel installation",
	Long: `Select packages to install in parallel with an interactive interface.

Packages are grouped by independence - packages in the same group can be
installed simultaneously for maximum speed.

This is useful for setting up a new Lambda instance quickly by installing
multiple independent packages at the same time.`,
	Run: runParallelize,
}

func init() {
	rootCmd.AddCommand(parallelizeCmd)
}

type parallelModel struct {
	packages       map[string]*installer.Package
	groups         []packageGroup
	selectedGroups map[int]bool
	cursor         int
	done           bool
	selections     []string
}

type packageGroup struct {
	name     string
	packages []string
	desc     string
}

func initialParallelModel() parallelModel {
	packages := installer.GetPackages()

	// Define groups of packages that can be installed in parallel
	groups := []packageGroup{
		{
			name:     "🏗️  Foundation Layer",
			packages: []string{"core"},
			desc:     "Essential system setup (required first)",
		},
		{
			name:     "🤖 ML & Python Stack",
			packages: []string{"python", "pytorch"},
			desc:     "Can install in parallel after core",
		},
		{
			name:     "🔮 LLM Runtime",
			packages: []string{"ollama"},
			desc:     "Can install in parallel with ML stack",
		},
		{
			name:     "⭐ LLM Models",
			packages: []string{"models-small", "models-medium", "models-large"},
			desc:     "Install after Ollama (can't parallelize between models)",
		},
		{
			name:     "🎨 Creative Tools",
			packages: []string{"comfyui", "claude"},
			desc:     "Can install in parallel after their dependencies",
		},
		{
			name:     "🎬 Video Generation Models",
			packages: []string{"mochi", "cogvideo", "opensora", "ltxvideo"},
			desc:     "Install after ComfyUI (can parallelize these)",
		},
		{
			name:     "🎥 Video Animation Tools",
			packages: []string{"svd", "animatediff"},
			desc:     "Install after ComfyUI (can parallelize these)",
		},
	}

	return parallelModel{
		packages:       packages,
		groups:         groups,
		selectedGroups: make(map[int]bool),
		cursor:         0,
		done:           false,
	}
}

func (m parallelModel) Init() tea.Cmd {
	return nil
}

func (m parallelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.groups)-1 {
				m.cursor++
			}

		case " ":
			// Toggle selection
			m.selectedGroups[m.cursor] = !m.selectedGroups[m.cursor]

		case "a":
			// Select all
			for i := range m.groups {
				m.selectedGroups[i] = true
			}

		case "n":
			// Select none
			m.selectedGroups = make(map[int]bool)

		case "enter":
			// Build selection list
			m.selections = []string{}
			for i, group := range m.groups {
				if m.selectedGroups[i] {
					m.selections = append(m.selections, group.packages...)
				}
			}
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m parallelModel) View() string {
	if m.done {
		if len(m.selections) == 0 {
			return theme.WarningStyle.Render("\nNo packages selected. Exiting.\n")
		}

		s := "\n" + theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n"
		s += theme.InfoStyle.Render("📦 Selected Packages for Parallel Installation") + "\n"
		s += theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n\n"

		for i, group := range m.groups {
			if m.selectedGroups[i] {
				s += "  " + theme.HighlightStyle.Render(group.name) + "\n"
				for _, pkgID := range group.packages {
					if pkg, exists := m.packages[pkgID]; exists {
						s += "    • " + theme.DimTextStyle.Render(pkg.Name+" ("+pkg.ID+")") + "\n"
					}
				}
				s += "\n"
			}
		}

		s += theme.InfoStyle.Render("✨ Installation Command:") + "\n"
		installCmd := "anime install " + fmt.Sprint(m.selections)
		s += "  " + theme.HighlightStyle.Render(installCmd) + "\n\n"

		s += theme.DimTextStyle.Render("💡 Tip: Packages within each group can be installed in parallel") + "\n"
		s += theme.DimTextStyle.Render("   Run: anime install --parallel <packages>") + "\n\n"

		return s
	}

	s := "\n" + theme.RenderBanner("⚡ PARALLEL INSTALLATION ⚡") + "\n\n"
	s += theme.InfoStyle.Render("Select package groups to install in parallel") + "\n\n"

	// Instructions
	s += theme.DimTextStyle.Render("  ↑/↓: Navigate  Space: Select  A: Select All  N: None  Enter: Continue  Q: Quit") + "\n\n"

	s += theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n\n"

	// Display groups
	for i, group := range m.groups {
		cursor := " "
		if m.cursor == i {
			cursor = theme.HighlightStyle.Render("▶")
		}

		checkbox := "☐"
		if m.selectedGroups[i] {
			checkbox = theme.SuccessStyle.Render("☑")
		}

		groupStyle := theme.SecondaryTextStyle
		if m.cursor == i {
			groupStyle = theme.HighlightStyle
		}

		s += fmt.Sprintf("  %s %s %s\n", cursor, checkbox, groupStyle.Render(group.name))
		s += "    " + theme.DimTextStyle.Render(group.desc) + "\n"

		// Show packages in group
		pkgList := ""
		for j, pkgID := range group.packages {
			if pkg, exists := m.packages[pkgID]; exists {
				if j > 0 {
					pkgList += ", "
				}
				pkgList += pkg.Name
			}
		}
		s += "    " + theme.DimTextStyle.Render("Packages: "+pkgList) + "\n\n"
	}

	s += theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n"

	// Selected count
	selectedCount := 0
	for range m.selectedGroups {
		selectedCount++
	}
	s += theme.InfoStyle.Render(fmt.Sprintf("\n📊 %d groups selected", selectedCount)) + "\n"

	return s
}

func runParallelize(cmd *cobra.Command, args []string) {
	m := initialParallelModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Error: "+err.Error()))
		return
	}

	if fm, ok := finalModel.(parallelModel); ok && len(fm.selections) > 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ Selection complete!"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Next steps:"))
		fmt.Println()
		fmt.Printf("  1. Copy the command above\n")
		fmt.Printf("  2. Run it on your Lambda server\n")
		fmt.Printf("  3. Or use: %s\n", theme.HighlightStyle.Render("anime lambda install <packages>"))
		fmt.Println()
	}
}
