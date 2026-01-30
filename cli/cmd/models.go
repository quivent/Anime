package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// ============================================================================
// MODEL DATA STRUCTURES
// ============================================================================

type ModelFile struct {
	Name      string
	Path      string
	SizeMB    float64
	ModelType string // LLM, Image/Video, Audio, etc.
	Category  string // Specific category within the type
	FileType  string // File extension
}

type InstallableModel struct {
	ID            string
	Name          string
	Description   string
	Type          string   // LLM, Image, Video, Enhancement, ControlNet
	Category      string   // Subcategory
	Size          string
	VRAM          string   // Minimum VRAM required
	HuggingFaceID string   // HuggingFace repo ID
	OllamaID      string   // Ollama model name (for LLMs)
	UseCases      []string
	Tags          []string
}

// GPU Architecture for model architect
type GPUConfig struct {
	Name       string
	VRAM       int    // GB
	Count      int
	TensorCore bool
	Generation string // Ampere, Ada, Hopper, etc.
}

// ============================================================================
// COMMANDS
// ============================================================================

var (
	modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "AI model management - list, search, install, and architect",
		Long: `Comprehensive AI model management system.

Commands:
  list       List all available models by category
  search     Search models by name, type, or use case
  dashboard  Interactive TUI to browse all models
  install    Install a specific model
  architect  Design optimal model architecture for your GPU setup

Examples:
  anime models list                    # List all models
  anime models list --type llm         # List only LLMs
  anime models search "video"          # Search for video models
  anime models dashboard               # Open interactive browser
  anime models install llama-3.3-70b   # Install a model
  anime models architect --gpus 2      # Design for 2 GPUs`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default: show help
			cmd.Help()
		},
	}

	modelsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available models",
		Long: `List all installable AI models organized by category.

Filter options:
  --type     Filter by model type (llm, image, video, enhancement, controlnet)
  --size     Filter by size (small <5GB, medium 5-20GB, large >20GB)
  --local    Show locally installed models instead`,
		Run: runModelsList,
	}

	modelsSearchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "Search for models",
		Long: `Search for models by name, description, use cases, or tags.

Examples:
  anime models search llama           # Search for llama models
  anime models search "code gen"      # Search for code generation
  anime models search video --type video`,
		Args: cobra.MinimumNArgs(1),
		Run:  runModelsSearch,
	}

	modelsDashboardCmd = &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"d", "browse", "tui"},
		Short:   "Interactive TUI to browse all models",
		Long: `Open an interactive terminal UI to browse all available models.

Features:
  - Browse by model type (LLM, Image, Video, etc.)
  - View detailed model information
  - Filter and search
  - One-click installation`,
		Run: runModelsDashboard,
	}

	modelsInstallCmd = &cobra.Command{
		Use:   "install [model-id]",
		Short: "Install a specific model",
		Long: `Install an AI model by its ID.

Examples:
  anime models install llama-3.3-70b
  anime models install flux-dev
  anime models install real-esrgan

Use 'anime models list' to see available model IDs.`,
		Args: cobra.ExactArgs(1),
		Run:  runModelsInstall,
	}

	modelsArchitectCmd = &cobra.Command{
		Use:   "architect",
		Short: "Design optimal model architecture for your GPU setup",
		Long: `Analyze your GPU configuration and recommend optimal model architectures.

This command will:
  1. Detect your GPU(s) or accept manual configuration
  2. Analyze available VRAM and compute capabilities
  3. Recommend models that fit your hardware
  4. Suggest optimal configurations for multi-GPU setups
  5. Propose different architecture options (speed vs quality)

Examples:
  anime models architect                    # Auto-detect GPUs
  anime models architect --gpus 2           # Specify GPU count
  anime models architect --vram 24          # Specify total VRAM
  anime models architect --propose 3        # Show 3 architecture proposals`,
		Run: runModelsArchitect,
	}

	// Flags
	modelsListType   string
	modelsListSize   string
	modelsListLocal  bool
	modelsSearchType string
	architectGPUs    int
	architectVRAM    int
	architectPropose int
)

func init() {
	rootCmd.AddCommand(modelsCmd)

	// Add subcommands
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsSearchCmd)
	modelsCmd.AddCommand(modelsDashboardCmd)
	modelsCmd.AddCommand(modelsInstallCmd)
	modelsCmd.AddCommand(modelsArchitectCmd)

	// List flags
	modelsListCmd.Flags().StringVarP(&modelsListType, "type", "t", "", "Filter by type (llm, image, video, enhancement, controlnet)")
	modelsListCmd.Flags().StringVarP(&modelsListSize, "size", "s", "", "Filter by size (small, medium, large)")
	modelsListCmd.Flags().BoolVarP(&modelsListLocal, "local", "l", false, "Show locally installed models")

	// Search flags
	modelsSearchCmd.Flags().StringVarP(&modelsSearchType, "type", "t", "", "Filter by type")

	// Architect flags
	modelsArchitectCmd.Flags().IntVarP(&architectGPUs, "gpus", "g", 0, "Number of GPUs (0 = auto-detect)")
	modelsArchitectCmd.Flags().IntVarP(&architectVRAM, "vram", "v", 0, "Total VRAM in GB (0 = auto-detect)")
	modelsArchitectCmd.Flags().IntVarP(&architectPropose, "propose", "p", 3, "Number of architecture proposals")
}

// ============================================================================
// LIST COMMAND
// ============================================================================

func runModelsList(cmd *cobra.Command, args []string) {
	if modelsListLocal {
		runLocalModelsList()
		return
	}

	models := getInstallableModels()

	// Filter by type
	if modelsListType != "" {
		filtered := []InstallableModel{}
		typeMap := map[string]string{
			"llm":         "LLM",
			"image":       "Image Generation",
			"video":       "Video Generation",
			"enhancement": "Enhancement",
			"controlnet":  "ControlNet",
		}
		targetType := typeMap[strings.ToLower(modelsListType)]
		for _, m := range models {
			if m.Type == targetType {
				filtered = append(filtered, m)
			}
		}
		models = filtered
	}

	// Filter by size
	if modelsListSize != "" {
		filtered := []InstallableModel{}
		for _, m := range models {
			sizeGB := parseSizeGB(m.Size)
			switch strings.ToLower(modelsListSize) {
			case "small":
				if sizeGB < 5 {
					filtered = append(filtered, m)
				}
			case "medium":
				if sizeGB >= 5 && sizeGB <= 20 {
					filtered = append(filtered, m)
				}
			case "large":
				if sizeGB > 20 {
					filtered = append(filtered, m)
				}
			}
		}
		models = filtered
	}

	// Group by type
	grouped := make(map[string][]InstallableModel)
	for _, m := range models {
		grouped[m.Type] = append(grouped[m.Type], m)
	}

	// Display
	typeOrder := []string{"LLM", "Image Generation", "Video Generation", "Enhancement", "ControlNet"}
	typeEmojis := map[string]string{
		"LLM":              "🤖",
		"Image Generation": "🎨",
		"Video Generation": "🎬",
		"Enhancement":      "🔧",
		"ControlNet":       "🎛️",
	}

	totalModels := 0
	for _, modelType := range typeOrder {
		models := grouped[modelType]
		if len(models) == 0 {
			continue
		}

		// Sort by name
		sort.Slice(models, func(i, j int) bool {
			return models[i].Name < models[j].Name
		})

		emoji := typeEmojis[modelType]
		fmt.Printf("\n%s %s\n", emoji, theme.HighlightStyle.Render(fmt.Sprintf("%s (%d)", modelType, len(models))))
		fmt.Println(theme.DimTextStyle.Render("─────────────────────────────────────────────────────────────────────"))

		// Compact table: ID | Name | Size | VRAM
		for _, model := range models {
			fmt.Printf("  %-22s %-8s %s\n",
				theme.InfoStyle.Render(model.ID),
				theme.DimTextStyle.Render(model.Size),
				theme.SecondaryTextStyle.Render(model.Category))
		}

		totalModels += len(models)
	}

	fmt.Println()
	fmt.Printf("%s  Install: %s\n",
		theme.SuccessStyle.Render(fmt.Sprintf("Total: %d models", totalModels)),
		theme.DimTextStyle.Render("anime models install <id>"))
	fmt.Println()
}

func runLocalModelsList() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("INSTALLED MODELS"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Scanning for installed models..."))
	fmt.Println()

	// Scan common model locations
	var models []ModelFile
	searchPaths := []string{
		filepath.Join(homeDir(), "ComfyUI", "models"),
		filepath.Join(homeDir(), ".cache", "huggingface"),
		filepath.Join(homeDir(), ".ollama", "models"),
		filepath.Join(homeDir(), "models"),
	}

	for _, searchPath := range searchPaths {
		for _, ext := range []string{"*.safetensors", "*.ckpt", "*.pth", "*.bin", "*.gguf"} {
			cmd := exec.Command("find", searchPath, "-type", "f", "-name", ext)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				model := parseModelFile(line)
				if model != nil {
					models = append(models, *model)
				}
			}
		}
	}

	if len(models) == 0 {
		fmt.Println(theme.WarningStyle.Render("No models found locally"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Install models with: anime models install <model-id>"))
		fmt.Println()
		return
	}

	// Group and display
	types := groupModelsByType(models)
	displayModels(types)
}

// ============================================================================
// SEARCH COMMAND
// ============================================================================

func runModelsSearch(cmd *cobra.Command, args []string) {
	query := strings.ToLower(strings.Join(args, " "))
	models := getInstallableModels()

	var matches []InstallableModel
	for _, m := range models {
		// Search in name, description, use cases, tags
		searchText := strings.ToLower(m.Name + " " + m.Description + " " + strings.Join(m.UseCases, " ") + " " + strings.Join(m.Tags, " "))
		if strings.Contains(searchText, query) {
			if modelsSearchType == "" || strings.EqualFold(m.Type, modelsSearchType) {
				matches = append(matches, m)
			}
		}
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("SEARCH RESULTS"))
	fmt.Println()
	fmt.Printf("  Query: %s\n", theme.HighlightStyle.Render(query))
	fmt.Printf("  Found: %s\n", theme.SuccessStyle.Render(fmt.Sprintf("%d models", len(matches))))
	fmt.Println()

	if len(matches) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No models found matching your query"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Try a different search term or browse with: anime models dashboard"))
		fmt.Println()
		return
	}

	for _, model := range matches {
		typeEmoji := map[string]string{
			"LLM":              "🤖",
			"Image Generation": "🎨",
			"Video Generation": "🎬",
			"Enhancement":      "🔧",
			"ControlNet":       "🎛️",
		}[model.Type]

		fmt.Printf("  %s %s %s\n",
			typeEmoji,
			theme.HighlightStyle.Render(model.Name),
			theme.DimTextStyle.Render(fmt.Sprintf("[%s / %s]", model.Type, model.Category)))
		fmt.Printf("    💾 %-10s  🎯 %s\n",
			theme.InfoStyle.Render(model.Size),
			theme.DimTextStyle.Render(model.VRAM))
		fmt.Printf("    📝 %s\n",
			theme.SecondaryTextStyle.Render(model.Description))
		fmt.Printf("    ⚡ %s\n",
			theme.DimTextStyle.Render("anime models install "+model.ID))
		fmt.Println()
	}
}

// ============================================================================
// DASHBOARD COMMAND - TUI
// ============================================================================

type modelsDashboardModel struct {
	models       []InstallableModel
	types        []string
	currentType  int
	currentModel int
	width        int
	height       int
	installing   bool
	selectedID   string
}

func newModelsDashboardModel() modelsDashboardModel {
	models := getInstallableModels()
	types := []string{"LLM", "Image Generation", "Video Generation", "Enhancement", "ControlNet"}

	return modelsDashboardModel{
		models: models,
		types:  types,
	}
}

func (m modelsDashboardModel) Init() tea.Cmd {
	return nil
}

func (m modelsDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "right", "l":
			m.currentType = (m.currentType + 1) % len(m.types)
			m.currentModel = 0

		case "shift+tab", "left", "h":
			m.currentType = (m.currentType - 1 + len(m.types)) % len(m.types)
			m.currentModel = 0

		case "up", "k":
			modelsInType := m.getModelsForCurrentType()
			if m.currentModel > 0 {
				m.currentModel--
			} else {
				m.currentModel = len(modelsInType) - 1
			}

		case "down", "j":
			modelsInType := m.getModelsForCurrentType()
			if m.currentModel < len(modelsInType)-1 {
				m.currentModel++
			} else {
				m.currentModel = 0
			}

		case "enter":
			modelsInType := m.getModelsForCurrentType()
			if m.currentModel < len(modelsInType) {
				m.selectedID = modelsInType[m.currentModel].ID
				m.installing = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m modelsDashboardModel) getModelsForCurrentType() []InstallableModel {
	var result []InstallableModel
	targetType := m.types[m.currentType]
	for _, model := range m.models {
		if model.Type == targetType {
			result = append(result, model)
		}
	}
	return result
}

func (m modelsDashboardModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF69B4")).
		Padding(0, 1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#6272A4"))

	activeTabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#FF69B4")).
		Bold(true)

	modelNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50FA7B"))

	selectedModelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#50FA7B"))

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))

	// Title
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🎨 AI MODEL BROWSER"))
	s.WriteString("\n\n")

	// Type tabs
	typeEmojis := map[string]string{
		"LLM":              "🤖",
		"Image Generation": "🎨",
		"Video Generation": "🎬",
		"Enhancement":      "🔧",
		"ControlNet":       "🎛️",
	}

	var tabs []string
	for i, t := range m.types {
		emoji := typeEmojis[t]
		shortName := strings.Split(t, " ")[0]
		label := fmt.Sprintf("%s %s", emoji, shortName)
		if i == m.currentType {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, tabStyle.Render(label))
		}
	}
	s.WriteString("  ")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	s.WriteString("\n\n")

	// Current type header
	currentTypeName := m.types[m.currentType]
	modelsInType := m.getModelsForCurrentType()
	s.WriteString(fmt.Sprintf("  %s %s (%d models)\n\n",
		typeEmojis[currentTypeName],
		currentTypeName,
		len(modelsInType)))

	// Model list (left side) and details (right side)
	maxHeight := m.height - 12
	if maxHeight < 5 {
		maxHeight = 5
	}

	// Calculate scroll
	startIdx := 0
	if m.currentModel > maxHeight-3 {
		startIdx = m.currentModel - maxHeight + 3
	}

	// Model list
	for i := startIdx; i < len(modelsInType) && i < startIdx+maxHeight; i++ {
		model := modelsInType[i]
		cursor := "  "
		style := modelNameStyle

		if i == m.currentModel {
			cursor = "▶ "
			style = selectedModelStyle
		}

		line := fmt.Sprintf("%s%s", cursor, model.Name)
		s.WriteString(style.Render(line))
		s.WriteString(detailStyle.Render(fmt.Sprintf(" %s", model.Size)))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Selected model details
	if m.currentModel < len(modelsInType) {
		model := modelsInType[m.currentModel]
		s.WriteString("  ────────────────────────────────────────────────────────────\n")
		s.WriteString(fmt.Sprintf("  📦 %s\n", modelNameStyle.Render(model.Name)))
		s.WriteString(fmt.Sprintf("  💾 Size: %s  |  🎯 VRAM: %s\n", model.Size, model.VRAM))
		s.WriteString(fmt.Sprintf("  📁 Category: %s\n", model.Category))
		s.WriteString(fmt.Sprintf("  📝 %s\n", descStyle.Render(model.Description)))
		if len(model.UseCases) > 0 {
			s.WriteString("  🎯 Use Cases:\n")
			for _, uc := range model.UseCases {
				s.WriteString(fmt.Sprintf("     • %s\n", uc))
			}
		}
		s.WriteString(fmt.Sprintf("  ⚡ Install: anime models install %s\n", model.ID))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  ←/→: Switch type  |  ↑/↓: Navigate  |  Enter: Install  |  q: Quit"))
	s.WriteString("\n")

	return s.String()
}

func runModelsDashboard(cmd *cobra.Command, args []string) {
	m := newModelsDashboardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check if user selected a model to install
	dm := finalModel.(modelsDashboardModel)
	if dm.installing && dm.selectedID != "" {
		fmt.Printf("\nInstalling %s...\n", dm.selectedID)
		installModel(dm.selectedID)
	}
}

// ============================================================================
// INSTALL COMMAND
// ============================================================================

func runModelsInstall(cmd *cobra.Command, args []string) {
	modelID := args[0]
	installModel(modelID)
}

func installModel(modelID string) {
	models := getInstallableModels()

	var model *InstallableModel
	for _, m := range models {
		if m.ID == modelID {
			model = &m
			break
		}
	}

	if model == nil {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("Model '%s' not found", modelID)))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Use 'anime models list' to see available models"))
		return
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("INSTALLING MODEL"))
	fmt.Println()
	fmt.Printf("  📦 %s\n", theme.HighlightStyle.Render(model.Name))
	fmt.Printf("  💾 Size: %s\n", model.Size)
	fmt.Printf("  🎯 VRAM Required: %s\n", model.VRAM)
	fmt.Println()

	// Determine installation method
	if model.OllamaID != "" {
		// Install via Ollama
		fmt.Println(theme.InfoStyle.Render("Installing via Ollama..."))
		cmd := exec.Command("ollama", "pull", model.OllamaID)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("Installation failed: %v", err)))
			return
		}
	} else if model.HuggingFaceID != "" {
		// Install via HuggingFace
		fmt.Println(theme.InfoStyle.Render("Downloading from HuggingFace..."))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(model.HuggingFaceID))
		fmt.Println()

		// Use huggingface-cli or wget
		destDir := filepath.Join(homeDir(), "ComfyUI", "models")
		switch model.Type {
		case "LLM":
			destDir = filepath.Join(homeDir(), ".cache", "huggingface", "hub")
		case "Image Generation":
			destDir = filepath.Join(homeDir(), "ComfyUI", "models", "checkpoints")
		case "Video Generation":
			destDir = filepath.Join(homeDir(), "ComfyUI", "models", "checkpoints")
		case "Enhancement":
			destDir = filepath.Join(homeDir(), "ComfyUI", "models", "upscale_models")
		case "ControlNet":
			destDir = filepath.Join(homeDir(), "ComfyUI", "models", "controlnet")
		}

		// Create directory
		os.MkdirAll(destDir, 0755)

		// Try huggingface-cli first
		cmd := exec.Command("huggingface-cli", "download", model.HuggingFaceID, "--local-dir", destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(theme.WarningStyle.Render("huggingface-cli not found, trying alternative..."))
			// Alternative: git clone
			cloneCmd := exec.Command("git", "clone", fmt.Sprintf("https://huggingface.co/%s", model.HuggingFaceID), filepath.Join(destDir, model.ID))
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			if err := cloneCmd.Run(); err != nil {
				fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("Installation failed: %v", err)))
				return
			}
		}
	} else {
		fmt.Println(theme.WarningStyle.Render("No installation method available for this model"))
		fmt.Println(theme.DimTextStyle.Render("Please install manually"))
		return
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Installation complete!"))
	fmt.Println()
}

// ============================================================================
// ARCHITECT COMMAND
// ============================================================================

func runModelsArchitect(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("MODEL ARCHITECT"))
	fmt.Println()

	// Detect or use provided GPU config
	var gpus []GPUConfig

	if architectGPUs == 0 && architectVRAM == 0 {
		// Auto-detect
		fmt.Println(theme.InfoStyle.Render("🔍 Detecting GPU configuration..."))
		fmt.Println()
		gpus = detectGPUs()
	} else {
		// Manual config
		vram := architectVRAM
		if vram == 0 {
			vram = 24 // Default assumption
		}
		count := architectGPUs
		if count == 0 {
			count = 1
		}

		for i := 0; i < count; i++ {
			gpus = append(gpus, GPUConfig{
				Name:       fmt.Sprintf("GPU %d", i),
				VRAM:       vram / count,
				Count:      1,
				TensorCore: true,
				Generation: "Unknown",
			})
		}
	}

	// Calculate total VRAM
	totalVRAM := 0
	for _, gpu := range gpus {
		totalVRAM += gpu.VRAM * gpu.Count
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.HighlightStyle.Render("  🖥️  GPU Configuration"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, gpu := range gpus {
		fmt.Printf("  • %s: %d GB VRAM (%s)\n", gpu.Name, gpu.VRAM, gpu.Generation)
	}
	fmt.Printf("\n  💾 Total VRAM: %d GB\n", totalVRAM)
	fmt.Println()

	// Generate architecture proposals
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.HighlightStyle.Render("  🏗️  Architecture Proposals"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	proposals := generateArchitectureProposals(totalVRAM, len(gpus), architectPropose)

	for i, proposal := range proposals {
		fmt.Printf("  %s %s\n",
			theme.SymbolSparkle,
			theme.HighlightStyle.Render(fmt.Sprintf("Proposal %d: %s", i+1, proposal.Name)))
		fmt.Printf("    📋 %s\n", theme.SecondaryTextStyle.Render(proposal.Description))
		fmt.Println()

		fmt.Println(theme.InfoStyle.Render("    Recommended Models:"))
		for _, model := range proposal.Models {
			fmt.Printf("      • %s (%s)\n",
				theme.HighlightStyle.Render(model.Name),
				theme.DimTextStyle.Render(model.Size))
		}
		fmt.Println()

		fmt.Printf("    ⚡ %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Estimated VRAM: %s", proposal.EstimatedVRAM)))
		fmt.Printf("    🎯 %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Use Case: %s", proposal.UseCase)))
		fmt.Println()
	}

	// Installation commands
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.HighlightStyle.Render("  📦 Quick Install Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for i, proposal := range proposals {
		fmt.Printf("  # Proposal %d: %s\n", i+1, proposal.Name)
		for _, model := range proposal.Models {
			fmt.Printf("  anime models install %s\n", model.ID)
		}
		fmt.Println()
	}
}

type ArchitectureProposal struct {
	Name          string
	Description   string
	Models        []InstallableModel
	EstimatedVRAM string
	UseCase       string
}

func detectGPUs() []GPUConfig {
	// Use centralized GPU detection (cached)
	sysInfo := gpu.GetSystemInfo()

	if !sysInfo.Available {
		fmt.Println(theme.WarningStyle.Render("  No NVIDIA GPU detected or nvidia-smi not available"))
		fmt.Println(theme.DimTextStyle.Render("  Using default configuration: 1x 24GB GPU"))
		fmt.Println()
		return []GPUConfig{{
			Name:       "Default GPU",
			VRAM:       24,
			Count:      1,
			TensorCore: true,
			Generation: "Unknown",
		}}
	}

	var gpus []GPUConfig
	for _, g := range sysInfo.GPUs {
		gpus = append(gpus, GPUConfig{
			Name:       g.Name,
			VRAM:       g.VRAM,
			Count:      1,
			TensorCore: g.TensorCore,
			Generation: g.Generation,
		})
	}

	return gpus
}

func generateArchitectureProposals(totalVRAM, gpuCount, numProposals int) []ArchitectureProposal {
	models := getInstallableModels()
	var proposals []ArchitectureProposal

	// Proposal 1: Speed-Optimized (smaller, faster models)
	speedModels := []InstallableModel{}
	speedVRAM := 0
	for _, m := range models {
		sizeGB := parseSizeGB(m.Size)
		if sizeGB <= 8 && speedVRAM+int(sizeGB) <= totalVRAM/2 {
			speedModels = append(speedModels, m)
			speedVRAM += int(sizeGB)
			if len(speedModels) >= 4 {
				break
			}
		}
	}
	proposals = append(proposals, ArchitectureProposal{
		Name:          "Speed-Optimized",
		Description:   "Fast inference with smaller, efficient models",
		Models:        speedModels,
		EstimatedVRAM: fmt.Sprintf("~%d GB", speedVRAM),
		UseCase:       "Rapid prototyping, real-time applications",
	})

	// Proposal 2: Quality-Optimized (larger, better models)
	qualityModels := []InstallableModel{}
	qualityVRAM := 0
	// Sort by size descending
	sortedModels := make([]InstallableModel, len(models))
	copy(sortedModels, models)
	sort.Slice(sortedModels, func(i, j int) bool {
		return parseSizeGB(sortedModels[i].Size) > parseSizeGB(sortedModels[j].Size)
	})

	for _, m := range sortedModels {
		sizeGB := parseSizeGB(m.Size)
		if qualityVRAM+int(sizeGB) <= totalVRAM && sizeGB >= 10 {
			qualityModels = append(qualityModels, m)
			qualityVRAM += int(sizeGB)
			if len(qualityModels) >= 3 {
				break
			}
		}
	}
	proposals = append(proposals, ArchitectureProposal{
		Name:          "Quality-Optimized",
		Description:   "Maximum quality with flagship models",
		Models:        qualityModels,
		EstimatedVRAM: fmt.Sprintf("~%d GB", qualityVRAM),
		UseCase:       "Production work, professional output",
	})

	// Proposal 3: Balanced (mix of speed and quality)
	balancedModels := []InstallableModel{}
	balancedVRAM := 0
	typesSeen := make(map[string]bool)

	for _, m := range models {
		sizeGB := parseSizeGB(m.Size)
		if !typesSeen[m.Type] && balancedVRAM+int(sizeGB) <= totalVRAM*3/4 {
			balancedModels = append(balancedModels, m)
			balancedVRAM += int(sizeGB)
			typesSeen[m.Type] = true
			if len(balancedModels) >= 5 {
				break
			}
		}
	}
	proposals = append(proposals, ArchitectureProposal{
		Name:          "Balanced Pipeline",
		Description:   "Full creative pipeline with diverse model types",
		Models:        balancedModels,
		EstimatedVRAM: fmt.Sprintf("~%d GB", balancedVRAM),
		UseCase:       "Complete workflow: text, image, video, enhancement",
	})

	// Limit to requested number
	if numProposals < len(proposals) {
		proposals = proposals[:numProposals]
	}

	return proposals
}

func parseSizeGB(size string) float64 {
	size = strings.ToLower(strings.TrimSpace(size))
	size = strings.ReplaceAll(size, "~", "")
	size = strings.ReplaceAll(size, "gb", "")
	size = strings.ReplaceAll(size, "mb", "")

	val, _ := strconv.ParseFloat(strings.TrimSpace(size), 64)
	if strings.Contains(strings.ToLower(size), "mb") || val > 100 {
		return val / 1024
	}
	return val
}

// ============================================================================
// MODEL CATALOG - Uses centralized model registry
// ============================================================================

func getInstallableModels() []InstallableModel {
	// Get all models from the centralized registry
	allSpecs := GetAllModels()

	var models []InstallableModel
	for _, spec := range allSpecs {
		// Map registry type to display type
		displayType := spec.Type
		switch spec.Type {
		case TypeLLM, TypeCoding, TypeMultimodal:
			displayType = "LLM"
		case TypeImage:
			displayType = "Image Generation"
		case TypeVideo:
			displayType = "Video Generation"
		case TypeEnhance:
			displayType = "Enhancement"
		case TypeControl:
			displayType = "ControlNet"
		}

		models = append(models, InstallableModel{
			ID:            spec.ID,
			Name:          spec.Name,
			Description:   spec.Description,
			Type:          displayType,
			Category:      spec.Category,
			Size:          spec.Size,
			VRAM:          spec.VRAM,
			HuggingFaceID: spec.HuggingFaceID,
			OllamaID:      spec.OllamaID,
			UseCases:      spec.UseCases,
			Tags:          spec.Tags,
		})
	}

	return models
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func scanLocalModels() ([]ModelFile, error) {
	var models []ModelFile

	searchPaths := []string{
		filepath.Join(homeDir(), "ComfyUI", "models"),
		filepath.Join(homeDir(), ".cache", "huggingface"),
		filepath.Join(homeDir(), "models"),
	}

	for _, searchPath := range searchPaths {
		for _, ext := range []string{"*.safetensors", "*.ckpt", "*.pth", "*.bin"} {
			cmd := exec.Command("find", searchPath, "-type", "f", "-name", ext)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}

				model := parseModelFile(line)
				if model != nil {
					models = append(models, *model)
				}
			}
		}
	}

	return models, nil
}

func scanRemoteModels(client *ssh.Client) ([]ModelFile, error) {
	var models []ModelFile

	searchCmd := `find ~ -type f \( -name "*.safetensors" -o -name "*.ckpt" -o -name "*.pth" -o -name "*.bin" \) 2>/dev/null`
	output, err := client.RunCommand(searchCmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		sizeCmd := fmt.Sprintf("du -m %s 2>/dev/null | awk '{print $1}'", line)
		sizeOutput, _ := client.RunCommand(sizeCmd)
		sizeMB := 0.0
		if sizeStr := strings.TrimSpace(sizeOutput); sizeStr != "" {
			if val, err := strconv.ParseFloat(sizeStr, 64); err == nil {
				sizeMB = val
			}
		}

		modelType, category := categorizeModel(line)
		model := &ModelFile{
			Name:      filepath.Base(line),
			Path:      line,
			SizeMB:    sizeMB,
			ModelType: modelType,
			Category:  category,
			FileType:  filepath.Ext(line),
		}

		models = append(models, *model)
	}

	return models, nil
}

func parseModelFile(path string) *ModelFile {
	sizeCmd := exec.Command("du", "-m", path)
	sizeOutput, err := sizeCmd.Output()
	sizeMB := 0.0
	if err == nil {
		fields := strings.Fields(string(sizeOutput))
		if len(fields) > 0 {
			if val, err := strconv.ParseFloat(fields[0], 64); err == nil {
				sizeMB = val
			}
		}
	}

	modelType, category := categorizeModel(path)
	return &ModelFile{
		Name:      filepath.Base(path),
		Path:      path,
		SizeMB:    sizeMB,
		ModelType: modelType,
		Category:  category,
		FileType:  filepath.Ext(path),
	}
}

func categorizeModel(path string) (string, string) {
	lowerPath := strings.ToLower(path)
	lowerName := strings.ToLower(filepath.Base(path))

	// Detect LLMs
	if strings.Contains(lowerPath, "ollama") || strings.Contains(lowerPath, ".ollama") {
		return "LLM", detectLLMCategory(lowerPath)
	}

	llmPatterns := []string{
		"llama", "mistral", "phi", "qwen", "vicuna", "alpaca",
		"gpt", "falcon", "bloom", "mpt", "starcoder", "codellama",
		"gemma", "mixtral", "yi-", "yi_", "deepseek", "orca", "wizard",
	}
	for _, pattern := range llmPatterns {
		if strings.Contains(lowerPath, pattern) || strings.Contains(lowerName, pattern) {
			return "LLM", detectLLMCategory(lowerPath)
		}
	}

	if strings.Contains(lowerPath, "transformers") ||
		strings.Contains(lowerPath, "language") ||
		strings.Contains(lowerPath, "/llm") {
		return "LLM", detectLLMCategory(lowerPath)
	}

	// Image/Video Generation
	if strings.Contains(lowerPath, "/checkpoints") || strings.Contains(lowerPath, "/checkpoint") {
		return "Image/Video", "Checkpoint"
	}
	if strings.Contains(lowerPath, "/loras") || strings.Contains(lowerPath, "/lora") {
		return "Image/Video", "LoRA"
	}
	if strings.Contains(lowerPath, "/vae") {
		return "Image/Video", "VAE"
	}
	if strings.Contains(lowerPath, "/controlnet") {
		return "Image/Video", "ControlNet"
	}
	if strings.Contains(lowerPath, "/embeddings") || strings.Contains(lowerPath, "/textual") {
		return "Image/Video", "Embedding"
	}
	if strings.Contains(lowerPath, "/upscale") {
		return "Image/Video", "Upscaler"
	}
	if strings.Contains(lowerPath, "/clip") {
		return "Image/Video", "CLIP"
	}
	if strings.Contains(lowerPath, "/unet") {
		return "Image/Video", "UNet"
	}

	// Check by filename patterns
	if strings.Contains(lowerName, "sdxl") {
		return "Image/Video", "SDXL"
	}
	if strings.Contains(lowerName, "sd15") || strings.Contains(lowerName, "sd_v15") {
		return "Image/Video", "SD 1.5"
	}
	if strings.Contains(lowerName, "flux") {
		return "Image/Video", "Flux"
	}
	if strings.Contains(lowerName, "stable") && strings.Contains(lowerName, "diffusion") {
		return "Image/Video", "Stable Diffusion"
	}
	if strings.Contains(lowerPath, "comfyui") {
		return "Image/Video", "ComfyUI"
	}

	// Audio models
	if strings.Contains(lowerPath, "/audio") ||
		strings.Contains(lowerName, "whisper") ||
		strings.Contains(lowerName, "bark") ||
		strings.Contains(lowerName, "audioldm") {
		return "Audio", "Audio Model"
	}

	return "Other", "Uncategorized"
}

func detectLLMCategory(pathOrName string) string {
	lower := strings.ToLower(pathOrName)

	if strings.Contains(lower, "7b") {
		return "7B"
	}
	if strings.Contains(lower, "13b") {
		return "13B"
	}
	if strings.Contains(lower, "34b") || strings.Contains(lower, "33b") {
		return "30B+"
	}
	if strings.Contains(lower, "70b") {
		return "70B"
	}
	if strings.Contains(lower, "180b") || strings.Contains(lower, "176b") {
		return "180B+"
	}

	if strings.Contains(lower, "llama") {
		return "Llama"
	}
	if strings.Contains(lower, "mistral") {
		return "Mistral"
	}
	if strings.Contains(lower, "phi") {
		return "Phi"
	}
	if strings.Contains(lower, "qwen") {
		return "Qwen"
	}
	if strings.Contains(lower, "gemma") {
		return "Gemma"
	}
	if strings.Contains(lower, "yi-") || strings.Contains(lower, "yi_") {
		return "Yi"
	}

	return "Language Model"
}

func groupModelsByType(models []ModelFile) map[string]map[string][]ModelFile {
	types := make(map[string]map[string][]ModelFile)

	for _, model := range models {
		if types[model.ModelType] == nil {
			types[model.ModelType] = make(map[string][]ModelFile)
		}
		types[model.ModelType][model.Category] = append(types[model.ModelType][model.Category], model)
	}

	for modelType := range types {
		for category := range types[modelType] {
			sort.Slice(types[modelType][category], func(i, j int) bool {
				return types[modelType][category][i].SizeMB > types[modelType][category][j].SizeMB
			})
		}
	}

	return types
}

func displayModels(types map[string]map[string][]ModelFile) {
	typeOrder := []string{"LLM", "Image/Video", "Audio", "Other"}
	var sortedTypes []string
	for _, t := range typeOrder {
		if _, exists := types[t]; exists {
			sortedTypes = append(sortedTypes, t)
		}
	}

	totalModels := 0
	totalSizeGB := 0.0
	totalCategories := 0

	typeEmojis := map[string]string{
		"LLM":         "🤖",
		"Image/Video": "🎨",
		"Audio":       "🎵",
		"Other":       "📦",
	}

	for _, modelType := range sortedTypes {
		categories := types[modelType]
		totalCategories += len(categories)

		typeModels := 0
		typeSizeGB := 0.0
		for _, models := range categories {
			typeModels += len(models)
			for _, m := range models {
				typeSizeGB += m.SizeMB / 1024.0
			}
		}
		totalModels += typeModels
		totalSizeGB += typeSizeGB

		emoji := typeEmojis[modelType]
		fmt.Println(theme.SuccessStyle.Render("╔══════════════════════════════════════════════════════════════════════════╗"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("║  %s  %s", emoji, modelType)))
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("║  %d models • %.1f GB • %d categories", typeModels, typeSizeGB, len(categories))))
		fmt.Println(theme.SuccessStyle.Render("╚══════════════════════════════════════════════════════════════════════════╝"))
		fmt.Println()

		var categoryNames []string
		for name := range categories {
			categoryNames = append(categoryNames, name)
		}
		sort.Strings(categoryNames)

		for _, category := range categoryNames {
			models := categories[category]

			categorySize := 0.0
			for _, m := range models {
				categorySize += m.SizeMB
			}

			fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  📁 %s (%d models, %.1f GB)", category, len(models), categorySize/1024.0)))
			fmt.Println(theme.DimTextStyle.Render("  ────────────────────────────────────────────────────────────────"))
			fmt.Println()

			for _, model := range models {
				sizeStr := formatModelSize(model.SizeMB)
				fmt.Printf("    %s %s\n",
					theme.HighlightStyle.Render(model.Name),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s)", sizeStr)))
				fmt.Printf("      %s\n", theme.DimTextStyle.Render(model.Path))
				fmt.Println()
			}
		}

		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s  %s  %s\n",
		theme.SuccessStyle.Render(fmt.Sprintf("📦 Total: %d models", totalModels)),
		theme.SuccessStyle.Render(fmt.Sprintf("💾 %.2f GB", totalSizeGB)),
		theme.DimTextStyle.Render(fmt.Sprintf("(%d types, %d categories)", len(sortedTypes), totalCategories)))
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
}

func formatModelSize(sizeMB float64) string {
	if sizeMB < 1024 {
		return fmt.Sprintf("%.0f MB", sizeMB)
	}
	return fmt.Sprintf("%.1f GB", sizeMB/1024.0)
}

func homeDir() string {
	cmd := exec.Command("sh", "-c", "echo $HOME")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "~"
}

// Keep this for remote scanning compatibility
func runModels(cmd *cobra.Command, args []string) error {
	// This is now handled by subcommands
	return nil
}
