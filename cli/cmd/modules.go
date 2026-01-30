package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var modulesCmd = &cobra.Command{
	Use:   "modules [server-name]",
	Short: "Select modules for a server (simple interactive picker)",
	Long:  `Opens a SIMPLE module selector - just numbers, no complex navigation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// If no server name provided, show helpful message
		if len(args) == 0 {
			showModulesHelp(cfg)
			return nil
		}

		serverName := args[0]
		server, err := cfg.GetServer(serverName)
		if err != nil {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server not found: " + serverName))
			fmt.Println()
			if len(cfg.Servers) > 0 {
				fmt.Println(theme.InfoStyle.Render("📋 Available servers:"))
				for _, s := range cfg.Servers {
					fmt.Println(theme.DimTextStyle.Render("  • " + s.Name))
				}
				fmt.Println()
			}
			fmt.Println(theme.InfoStyle.Render("💡 Options:"))
			fmt.Println(theme.DimTextStyle.Render("  anime add <server-name>    # Add a new server"))
			fmt.Println(theme.DimTextStyle.Render("  anime list                 # List all servers"))
			fmt.Println()
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
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required arguments"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime set-modules <server-name> <module-ids...>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime set-modules lambda-1 core pytorch ollama"))
			fmt.Println(theme.DimTextStyle.Render("  anime set-modules my-server core pytorch ollama models-small claude"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  anime list-modules        # List all available modules"))
			fmt.Println(theme.DimTextStyle.Render("  anime modules <server>    # Interactive module picker"))
			fmt.Println()
			return fmt.Errorf("set-modules requires server name and at least one module ID")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		moduleIDs := args[1:]

		server, err := cfg.GetServer(serverName)
		if err != nil {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server not found: " + serverName))
			fmt.Println()
			if len(cfg.Servers) > 0 {
				fmt.Println(theme.InfoStyle.Render("📋 Available servers:"))
				for _, s := range cfg.Servers {
					fmt.Println(theme.DimTextStyle.Render("  • " + s.Name))
				}
				fmt.Println()
			}
			fmt.Println(theme.InfoStyle.Render("💡 Options:"))
			fmt.Println(theme.DimTextStyle.Render("  anime add <server-name>    # Add a new server"))
			fmt.Println(theme.DimTextStyle.Render("  anime list                 # List all servers"))
			fmt.Println()
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
		fmt.Println()
		fmt.Println(theme.RenderBanner("AVAILABLE MODULES"))
		fmt.Println()

		categories := config.GetModulesByCategory()
		catEmojis := map[string]string{
			"System":       "⚙️",
			"LLM-Frontier": "🏆",
			"LLM-Large":    "🦣",
			"LLM-Medium":   "🤖",
			"LLM-Small":    "🚀",
			"Image":        "🎨",
			"Video":        "🎬",
			"Tools":        "🛠️",
		}
		catNames := map[string]string{
			"System":       "System",
			"LLM-Frontier": "LLM - Frontier (Multi-GPU/B200)",
			"LLM-Large":    "LLM - Large (70B+)",
			"LLM-Medium":   "LLM - Medium (14-34B)",
			"LLM-Small":    "LLM - Small (≤8B)",
			"Image":        "Image Generation",
			"Video":        "Video Generation",
			"Tools":        "Tools",
		}
		catOrder := []string{"System", "LLM-Frontier", "LLM-Large", "LLM-Medium", "LLM-Small", "Image", "Video", "Tools"}

		for _, cat := range catOrder {
			mods, ok := categories[cat]
			if !ok || len(mods) == 0 {
				continue
			}

			emoji := catEmojis[cat]
			displayName := catNames[cat]
			if displayName == "" {
				displayName = cat
			}
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("━━━ %s %s (%d) ━━━", emoji, displayName, len(mods))))
			fmt.Println()

			for _, mod := range mods {
				fmt.Printf("  %s %s\n",
					theme.HighlightStyle.Render(mod.ID),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s, ~%dm)", mod.Size, mod.TimeMinutes)))
				fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(mod.Description))
				if len(mod.Dependencies) > 0 {
					fmt.Printf("    %s\n", theme.DimTextStyle.Render("requires: "+strings.Join(mod.Dependencies, ", ")))
				}
				// Show cluster requirements for frontier models
				if mod.Cluster != nil {
					fmt.Printf("    %s\n", theme.WarningStyle.Render(fmt.Sprintf("⚡ %d-%d GPUs (%dGB+ VRAM each), %s",
						mod.Cluster.MinGPUs, mod.Cluster.RecommendedGPUs, mod.Cluster.MinVRAMPerGPU, mod.Cluster.Parallelism)))
				}
				fmt.Println()
			}
		}

		// Show B200 cluster guide after frontier section
		if _, hasFrontier := categories["LLM-Frontier"]; hasFrontier {
			fmt.Println(theme.SuccessStyle.Render("━━━ 📋 B200 CLUSTER CONFIGURATION GUIDE ━━━"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  Recommended Setup for MoE Models:"))
			fmt.Println(theme.DimTextStyle.Render("    • TP4EP2: 4-way tensor parallelism + 2-way expert parallelism"))
			fmt.Println(theme.DimTextStyle.Render("    • Total VRAM: 8×B200 = 1.44TB (180GB each)"))
			fmt.Println(theme.DimTextStyle.Render("    • NVLink: Required for efficient expert routing"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  Loading Best Practices:"))
			fmt.Println(theme.DimTextStyle.Render("    • Use FP8 quantization for 30% memory reduction"))
			fmt.Println(theme.DimTextStyle.Render("    • Set --max-model-len for context (32K-64K recommended)"))
			fmt.Println(theme.DimTextStyle.Render("    • Enable --enable-chunked-prefill for long contexts"))
			fmt.Println(theme.DimTextStyle.Render("    • Use --trust-remote-code for DeepSeek models"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  Crash Prevention:"))
			fmt.Println(theme.DimTextStyle.Render("    • Monitor GPU memory with nvidia-smi -l 1"))
			fmt.Println(theme.DimTextStyle.Render("    • Start with smaller context, scale up gradually"))
			fmt.Println(theme.DimTextStyle.Render("    • Use TensorRT-LLM for production (368 tok/s on 8×B200)"))
			fmt.Println(theme.DimTextStyle.Render("    • vLLM: Good for development, easier setup"))
			fmt.Println()
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("💡 Usage:"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime modules <server>"))
		fmt.Println(theme.DimTextStyle.Render("    Interactive module selector (recommended)"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime set-modules <server> core pytorch ollama mistral-7b"))
		fmt.Println(theme.DimTextStyle.Render("    Set modules via command line"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Quick install combos:"))
		fmt.Println(theme.DimTextStyle.Render("    Minimal:    core pytorch claude"))
		fmt.Println(theme.DimTextStyle.Render("    LLM Dev:    core ollama mistral-7b llama-3.3-8b"))
		fmt.Println(theme.DimTextStyle.Render("    Image Gen:  core pytorch comfyui sdxl flux-dev"))
		fmt.Println(theme.DimTextStyle.Render("    Video Gen:  core pytorch comfyui svd wan2"))
		fmt.Println(theme.DimTextStyle.Render("    Full Stack: core pytorch ollama mistral-7b comfyui sdxl claude"))
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modulesCmd)
	rootCmd.AddCommand(setModulesCmd)
	rootCmd.AddCommand(listModulesCmd)
}

func showModulesHelp(cfg *config.Config) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚙️  MODULES ⚙️"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Configure installation modules for your servers"))
	fmt.Println()

	// List available servers
	if len(cfg.Servers) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠️  No servers configured"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Add a server first:"))
		fmt.Println(theme.HighlightStyle.Render("  anime add <server-name>"))
		fmt.Println()
		return
	}

	fmt.Println(theme.SuccessStyle.Render("📋 Configured Servers:"))
	fmt.Println()
	for _, server := range cfg.Servers {
		moduleCount := len(server.Modules)
		moduleText := "no modules"
		if moduleCount > 0 {
			moduleText = fmt.Sprintf("%d modules", moduleCount)
		}
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(server.Name),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s)", moduleText)))
	}
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("💡 Usage:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Interactive picker:"))
	fmt.Println(theme.HighlightStyle.Render("    anime modules <server-name>"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Set via command line:"))
	fmt.Println(theme.HighlightStyle.Render("    anime set-modules <server-name> core pytorch ollama"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  List all available modules:"))
	fmt.Println(theme.HighlightStyle.Render("    anime list-modules"))
	fmt.Println()

	if len(cfg.Servers) > 0 {
		fmt.Println(theme.InfoStyle.Render("✨ Quick start:"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("  anime modules %s", cfg.Servers[0].Name)))
		fmt.Println()
	}
}

// Category order for display
var categoryOrder = []string{"System", "LLM-Frontier", "LLM-Large", "LLM-Medium", "LLM-Small", "Image", "Video", "Tools"}

// Category emojis and display names
var categoryEmojis = map[string]string{
	"System":       "⚙️",
	"LLM-Frontier": "🏆",
	"LLM-Large":    "🦣",
	"LLM-Medium":   "🤖",
	"LLM-Small":    "🚀",
	"Image":        "🎨",
	"Video":        "🎬",
	"Tools":        "🛠️",
}

// Category display names
var categoryNames = map[string]string{
	"System":       "System",
	"LLM-Frontier": "LLM - Frontier (Multi-GPU/B200 Required)",
	"LLM-Large":    "LLM - Large (70B+, 2+ GPUs)",
	"LLM-Medium":   "LLM - Medium (14-34B, Single GPU)",
	"LLM-Small":    "LLM - Small (≤8B, Consumer GPU)",
	"Image":        "Image Generation",
	"Video":        "Video Generation",
	"Tools":        "Tools",
}

// Simple module picker with categories
type simpleModulePicker struct {
	server       *config.Server
	cfg          *config.Config
	selected     map[string]bool
	cursor       int
	showDetails  bool
	detailIdx    int
	done         bool
	width        int
	height       int
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
		cursor:   0,
	}
}

func (m simpleModulePicker) Init() tea.Cmd {
	return nil
}

func (m simpleModulePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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
			if m.cursor < len(config.AvailableModules)-1 {
				m.cursor++
			}

		case " ", "x":
			// Toggle selection
			mod := config.AvailableModules[m.cursor]
			if m.selected[mod.ID] {
				delete(m.selected, mod.ID)
			} else {
				m.selected[mod.ID] = true
				// Auto-select dependencies
				for _, dep := range mod.Dependencies {
					m.selected[dep] = true
				}
			}

		case "enter":
			// Save and exit
			var modules []string
			for id := range m.selected {
				modules = append(modules, id)
			}

			m.server.Modules = modules
			m.cfg.UpdateServer(m.server.Name, *m.server)
			m.cfg.Save()

			m.done = true
			return m, tea.Quit

		case "?", "i":
			// Toggle details view
			m.showDetails = !m.showDetails
			m.detailIdx = m.cursor

		case "a":
			// Select all in current category
			currentMod := config.AvailableModules[m.cursor]
			for _, mod := range config.AvailableModules {
				if mod.Category == currentMod.Category {
					m.selected[mod.ID] = true
					for _, dep := range mod.Dependencies {
						m.selected[dep] = true
					}
				}
			}

		case "n":
			// Deselect all in current category
			currentMod := config.AvailableModules[m.cursor]
			for _, mod := range config.AvailableModules {
				if mod.Category == currentMod.Category {
					delete(m.selected, mod.ID)
				}
			}
		}
	}

	return m, nil
}

func (m simpleModulePicker) View() string {
	if m.done {
		var modules []string
		for id := range m.selected {
			modules = append(modules, id)
		}
		return fmt.Sprintf("✓ Modules updated! (%d selected)\n", len(modules))
	}

	var s strings.Builder

	// Header
	s.WriteString(theme.RenderBanner(fmt.Sprintf("SELECT MODULES: %s", strings.ToUpper(m.server.Name))))
	s.WriteString("\n\n")

	// Group modules by category
	categories := config.GetModulesByCategory()

	idx := 0
	for _, cat := range categoryOrder {
		mods, ok := categories[cat]
		if !ok || len(mods) == 0 {
			continue
		}

		// Category header
		emoji := categoryEmojis[cat]
		displayName := categoryNames[cat]
		if displayName == "" {
			displayName = cat
		}
		s.WriteString(theme.SuccessStyle.Render(fmt.Sprintf("━━━ %s %s ━━━", emoji, displayName)))
		s.WriteString("\n")

		for _, mod := range mods {
			// Find actual index in AvailableModules
			actualIdx := -1
			for i, m := range config.AvailableModules {
				if m.ID == mod.ID {
					actualIdx = i
					break
				}
			}

			// Cursor and selection
			cursor := "  "
			if actualIdx == m.cursor {
				cursor = theme.HighlightStyle.Render("▶ ")
			}

			check := "[ ]"
			if m.selected[mod.ID] {
				check = theme.SuccessStyle.Render("[✓]")
			}

			// Module line
			name := mod.Name
			if actualIdx == m.cursor {
				name = theme.HighlightStyle.Render(mod.Name)
			}

			sizeStr := theme.DimTextStyle.Render(fmt.Sprintf("(%s)", mod.Size))
			descStr := theme.DimTextStyle.Render(mod.Description)

			s.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, check, name, sizeStr))
			s.WriteString(fmt.Sprintf("       %s\n", descStr))

			idx++
		}
		s.WriteString("\n")
	}

	// Summary
	s.WriteString(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	s.WriteString("\n")

	// Calculate totals
	var selectedIDs []string
	totalTime := 0
	for id := range m.selected {
		selectedIDs = append(selectedIDs, id)
		for _, mod := range config.AvailableModules {
			if mod.ID == id {
				totalTime += mod.TimeMinutes
				break
			}
		}
	}
	cost := config.EstimateCost(selectedIDs, m.server.CostPerHour)

	s.WriteString(fmt.Sprintf("  Selected: %s  |  Time: %s  |  Cost: %s\n",
		theme.HighlightStyle.Render(fmt.Sprintf("%d modules", len(selectedIDs))),
		theme.InfoStyle.Render(fmt.Sprintf("~%dm", totalTime)),
		theme.SuccessStyle.Render(fmt.Sprintf("$%.2f", cost))))

	s.WriteString(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	s.WriteString("\n\n")

	// Help
	s.WriteString(theme.DimTextStyle.Render("  ↑/↓ navigate  |  space/x toggle  |  a select category  |  n deselect category"))
	s.WriteString("\n")
	s.WriteString(theme.DimTextStyle.Render("  enter save    |  q cancel        |  ? details"))
	s.WriteString("\n")

	return s.String()
}
