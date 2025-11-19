package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/spf13/cobra"
)

var modulesCmd = &cobra.Command{
	Use:   "modules [server-name]",
	Short: "Select modules for a server (simple interactive picker)",
	Long:  `Opens a SIMPLE module selector - just numbers, no complex navigation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		server, err := cfg.GetServer(serverName)
		if err != nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		m := newSimpleModulePicker(server, cfg)
		p := tea.NewProgram(m)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running module picker: %w", err)
		}

		return nil
	},
}

var setModulesCmd = &cobra.Command{
	Use:   "set-modules [server-name] [module-ids...]",
	Short: "Set modules via CLI (no TUI)",
	Long:  `Set modules for a server using CLI flags. No interactive UI.`,
	Example: `  anime set-modules lambda-1 core pytorch ollama
  anime set-modules my-server core pytorch ollama models-small claude`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		moduleIDs := args[1:]

		server, err := cfg.GetServer(serverName)
		if err != nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		// Validate modules
		validModules := make(map[string]bool)
		for _, mod := range config.AvailableModules {
			validModules[mod.ID] = true
		}

		for _, id := range moduleIDs {
			if !validModules[id] {
				return fmt.Errorf("invalid module: %s (see 'anime list-modules')", id)
			}
		}

		server.Modules = moduleIDs
		cfg.UpdateServer(serverName, *server)

		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("✓ Updated modules for '%s'\n\n", serverName)

		// Show what will be installed
		cost := config.EstimateCost(moduleIDs, server.CostPerHour)
		modules := config.GetModulesByID(moduleIDs)

		totalTime := 0
		for _, mod := range modules {
			totalTime += mod.TimeMinutes
			fmt.Printf("  • %s (%dm)\n", mod.Name, mod.TimeMinutes)
		}

		fmt.Printf("\nTotal time: %dm\n", totalTime)
		fmt.Printf("Estimated cost: $%.2f @ $%.2f/hr\n\n", cost, server.CostPerHour)
		fmt.Printf("Deploy with: anime deploy %s\n", serverName)

		return nil
	},
}

var listModulesCmd = &cobra.Command{
	Use:   "list-modules",
	Short: "List all available modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Available modules:\n")

		for _, mod := range config.AvailableModules {
			fmt.Printf("  %s\n", mod.ID)
			fmt.Printf("    Name: %s\n", mod.Name)
			fmt.Printf("    Time: %d minutes\n", mod.TimeMinutes)
			fmt.Printf("    Description: %s\n", mod.Description)
			if len(mod.Dependencies) > 0 {
				fmt.Printf("    Dependencies: %s\n", strings.Join(mod.Dependencies, ", "))
			}
			fmt.Println()
		}

		fmt.Println("Quick install combos:")
		fmt.Println("  Minimal:    core pytorch")
		fmt.Println("  Standard:   core pytorch ollama models-small")
		fmt.Println("  Full:       core pytorch ollama models-large comfyui claude")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  anime set-modules SERVER core pytorch ollama")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modulesCmd)
	rootCmd.AddCommand(setModulesCmd)
	rootCmd.AddCommand(listModulesCmd)
}

// Simple module picker - just type numbers!
type simpleModulePicker struct {
	server   *config.Server
	cfg      *config.Config
	selected map[string]bool
	input    string
	done     bool
}

func newSimpleModulePicker(server *config.Server, cfg *config.Config) simpleModulePicker {
	selected := make(map[string]bool)
	for _, id := range server.Modules {
		selected[id] = true
	}

	return simpleModulePicker{
		server:   server,
		cfg:      cfg,
		selected: selected,
		input:    "",
	}
}

func (m simpleModulePicker) Init() tea.Cmd {
	return nil
}

func (m simpleModulePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "enter":
			// Parse input
			numbers := strings.Fields(m.input)
			m.selected = make(map[string]bool)

			for _, numStr := range numbers {
				var idx int
				fmt.Sscanf(numStr, "%d", &idx)
				if idx > 0 && idx <= len(config.AvailableModules) {
					m.selected[config.AvailableModules[idx-1].ID] = true
				}
			}

			// Save
			var modules []string
			for id := range m.selected {
				modules = append(modules, id)
			}

			m.server.Modules = modules
			m.cfg.UpdateServer(m.server.Name, *m.server)
			m.cfg.Save()

			m.done = true
			return m, tea.Quit

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

func (m simpleModulePicker) View() string {
	if m.done {
		return "✓ Modules updated!\n"
	}

	var s strings.Builder

	s.WriteString(fmt.Sprintf("Select modules for: %s\n\n", m.server.Name))
	s.WriteString("Available modules:\n\n")

	for i, mod := range config.AvailableModules {
		check := " "
		if m.selected[mod.ID] {
			check = "✓"
		}
		s.WriteString(fmt.Sprintf("  [%s] %d. %s (%dm) - %s\n",
			check, i+1, mod.Name, mod.TimeMinutes, mod.Description))
	}

	s.WriteString("\n")

	// Calculate cost
	var selectedIDs []string
	for id := range m.selected {
		selectedIDs = append(selectedIDs, id)
	}
	cost := config.EstimateCost(selectedIDs, m.server.CostPerHour)
	s.WriteString(fmt.Sprintf("Estimated cost: $%.2f\n\n", cost))

	s.WriteString("Type numbers separated by spaces (e.g., '1 2 3 8'), then press Enter:\n")
	s.WriteString("> " + m.input + "\n\n")
	s.WriteString("Press 'q' to cancel\n")

	return s.String()
}
