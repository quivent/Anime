package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get workflow and package suggestions based on your setup",
	Long: `Analyze your installed packages and assets to suggest:
  • Available workflows you can run
  • Complementary packages to install
  • Optimal usage patterns for your current setup

This command examines your Lambda server's installed packages and local collections
to provide personalized recommendations.`,
	Run: runSuggest,
}

func init() {
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("💡 NEXT STEPS 💡"))
	fmt.Println()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to load config: " + err.Error()))
		return
	}

	// Get installed packages
	installedPackages := checkInstalledPackages()
	allPackages := installer.GetPackages()

	// Get collections
	collections := cfg.ListCollections()

	// Analyze setup
	analysis := analyzeSetup(installedPackages, allPackages, collections)

	// Get actual actionable suggestions
	suggestions := getActionableSuggestions(analysis, installedPackages, allPackages, collections)

	// Display suggestions
	displayActionableSuggestions(suggestions)
}

// SetupAnalysis contains the analysis results
type SetupAnalysis struct {
	hasCore          bool
	hasPython        bool
	hasPyTorch       bool
	hasOllama        bool
	hasComfyUI       bool
	hasClaude        bool
	hasSmallModels   bool
	hasMediumModels  bool
	hasLargeModels   bool
	videoModels      []string
	hasCollections   bool
	imageCollections int
	videoCollections int
	mixedCollections int
	totalCollections int
	capabilities     []string
}

func analyzeSetup(installed map[string]bool, allPackages map[string]*installer.Package, collections []config.Collection) *SetupAnalysis {
	analysis := &SetupAnalysis{
		videoModels: []string{},
		capabilities: []string{},
	}

	// Analyze installed packages
	analysis.hasCore = installed["core"]
	analysis.hasPython = installed["python"]
	analysis.hasPyTorch = installed["pytorch"]
	analysis.hasOllama = installed["ollama"]
	analysis.hasComfyUI = installed["comfyui"]
	analysis.hasClaude = installed["claude"]
	analysis.hasSmallModels = installed["models-small"]
	analysis.hasMediumModels = installed["models-medium"]
	analysis.hasLargeModels = installed["models-large"]

	// Check video models
	videoModelIDs := []string{"mochi", "svd", "animatediff", "cogvideo", "opensora", "ltxvideo", "wan2"}
	for _, modelID := range videoModelIDs {
		if installed[modelID] {
			analysis.videoModels = append(analysis.videoModels, modelID)
		}
	}

	// Analyze collections
	analysis.totalCollections = len(collections)
	analysis.hasCollections = len(collections) > 0
	for _, col := range collections {
		switch col.Type {
		case "image":
			analysis.imageCollections++
		case "video":
			analysis.videoCollections++
		case "mixed":
			analysis.mixedCollections++
		}
	}

	// Determine capabilities
	if analysis.hasOllama && (analysis.hasSmallModels || analysis.hasMediumModels || analysis.hasLargeModels) {
		analysis.capabilities = append(analysis.capabilities, "Text Generation")
		analysis.capabilities = append(analysis.capabilities, "Code Generation")
		analysis.capabilities = append(analysis.capabilities, "Chat/Assistant")
	}
	if analysis.hasComfyUI {
		analysis.capabilities = append(analysis.capabilities, "Image Generation")
		analysis.capabilities = append(analysis.capabilities, "Image Editing")
	}
	if len(analysis.videoModels) > 0 {
		analysis.capabilities = append(analysis.capabilities, "Video Generation")
	}
	if analysis.hasPyTorch {
		analysis.capabilities = append(analysis.capabilities, "Model Training")
		analysis.capabilities = append(analysis.capabilities, "Fine-tuning")
	}
	if analysis.hasClaude {
		analysis.capabilities = append(analysis.capabilities, "AI-Assisted Coding")
	}

	return analysis
}

func displayAnalysisSummary(analysis *SetupAnalysis, installed map[string]bool, allPackages map[string]*installer.Package) {
	installedCount := len(installed)
	totalCount := len(allPackages)
	completionPct := 0
	if totalCount > 0 {
		completionPct = (installedCount * 100) / totalCount
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Setup Overview"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Check if we can detect packages
	if installedCount == 0 {
		// Check if lambda is configured
		cfg, _ := config.Load()
		lambdaConfigured := false
		if cfg != nil {
			lambdaTarget := cfg.GetAlias("lambda")
			if lambdaTarget != "" {
				lambdaConfigured = true
			} else if _, err := cfg.GetServer("lambda"); err == nil {
				lambdaConfigured = true
			}
		}

		if !lambdaConfigured {
			fmt.Printf("  Packages Installed:  %s\n",
				theme.WarningStyle.Render("Unable to detect - lambda server not configured"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  💡 Configure lambda server to detect installed packages:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime set lambda <server-ip>"))
		} else {
			fmt.Printf("  Packages Installed:  %s\n",
				theme.WarningStyle.Render("Unable to detect - connection failed"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  💡 Check SSH connection to lambda server"))
			fmt.Println(theme.DimTextStyle.Render("     Make sure you can: ssh ubuntu@<server-ip>"))
		}
		fmt.Println()
	} else {
		fmt.Printf("  Packages Installed:  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d/%d (%d%%)", installedCount, totalCount, completionPct)))
	}

	fmt.Printf("  Asset Collections:   %s\n",
		theme.HighlightStyle.Render(fmt.Sprintf("%d", analysis.totalCollections)))

	if analysis.totalCollections > 0 {
		fmt.Printf("    %s %d image  %s %d video  %s %d mixed\n",
			theme.SymbolSparkle, analysis.imageCollections,
			theme.SymbolSparkle, analysis.videoCollections,
			theme.SymbolSparkle, analysis.mixedCollections)
	}
	fmt.Println()

	if len(analysis.capabilities) > 0 {
		fmt.Println(theme.InfoStyle.Render("  🎯 Available Capabilities:"))
		for _, cap := range analysis.capabilities {
			fmt.Printf("    %s %s\n", theme.SymbolBolt, theme.SuccessStyle.Render(cap))
		}
		fmt.Println()
	} else {
		fmt.Println(theme.WarningStyle.Render("  ⚠️  No capabilities detected - install packages to get started"))
		fmt.Println()
	}
}

func displayPackageBreakdown(installed map[string]bool, allPackages map[string]*installer.Package) {
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Package Status"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Group packages by category
	categories := map[string][]string{
		"Foundation":      {},
		"ML Framework":    {},
		"LLM Runtime":     {},
		"Models":          {},
		"Application":     {},
		"Video Generation": {},
	}

	// Sort packages into categories
	for id, pkg := range allPackages {
		categories[pkg.Category] = append(categories[pkg.Category], id)
	}

	// Display each category
	categoryOrder := []string{"Foundation", "ML Framework", "LLM Runtime", "Models", "Application", "Video Generation"}

	for _, category := range categoryOrder {
		pkgIDs := categories[category]
		if len(pkgIDs) == 0 {
			continue
		}

		// Sort packages alphabetically
		sort.Strings(pkgIDs)

		fmt.Printf("  %s\n", theme.GetCategoryStyle(category).Render(category))
		for _, id := range pkgIDs {
			pkg := allPackages[id]
			if installed[id] {
				fmt.Printf("    %s %s\n",
					theme.SuccessStyle.Render("✓"),
					theme.DimTextStyle.Render(pkg.Name))
			} else {
				fmt.Printf("    %s %s  %s\n",
					theme.WarningStyle.Render("○"),
					theme.PrimaryTextStyle.Render(pkg.Name),
					theme.DimTextStyle.Render("→ anime install "+id))
			}
		}
		fmt.Println()
	}

	// Add prominent install all suggestion if nothing is installed
	if len(installed) == 0 {
		fmt.Println(theme.InfoStyle.Render("  💡 Quick Start:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime install core"))
		fmt.Println(theme.DimTextStyle.Render("    Start with the foundation package, then add others as needed"))
		fmt.Println()
	}
}

func displayWorkflowSuggestions(analysis *SetupAnalysis) {
	workflows := getWorkflowSuggestions(analysis)

	if len(workflows) == 0 {
		return
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Suggested Workflows"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, wf := range workflows {
		fmt.Printf("  %s %s\n", wf.emoji, theme.HighlightStyle.Render(wf.name))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(wf.description))
		if wf.command != "" {
			fmt.Printf("    %s %s\n", theme.SymbolBolt, theme.InfoStyle.Render(wf.command))
		}
		if wf.requirements != "" {
			fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render("Requires: "+wf.requirements))
		}
		fmt.Println()
	}
}

func displayPackageSuggestions(analysis *SetupAnalysis, installed map[string]bool, allPackages map[string]*installer.Package) {
	suggestions := getPackageSuggestions(analysis, installed, allPackages)

	if len(suggestions) == 0 {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("🎉 Package Suggestions"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("  ✓ You have a complete setup! All complementary packages are installed."))
		fmt.Println()
		return
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Recommended Packages"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, sugg := range suggestions {
		pkg := allPackages[sugg.packageID]
		if pkg == nil {
			continue
		}

		fmt.Printf("  %s %s\n", sugg.emoji, theme.HighlightStyle.Render(pkg.Name))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(pkg.Description))
		fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(sugg.reason))
		fmt.Printf("    %s  |  %s  |  %s\n",
			theme.DimTextStyle.Render("⏱️  "+pkg.EstimatedTime.String()),
			theme.DimTextStyle.Render("💾 "+pkg.Size),
			theme.InfoStyle.Render("anime install "+pkg.ID))
		fmt.Println()
	}
}

func displayOptimizationTips(analysis *SetupAnalysis) {
	tips := getOptimizationTips(analysis)

	if len(tips) == 0 {
		return
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 Optimization Tips"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, tip := range tips {
		fmt.Printf("  %s %s\n", theme.SymbolSparkle, theme.HighlightStyle.Render(tip.title))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(tip.description))
		if tip.command != "" {
			fmt.Printf("    %s %s\n", theme.SymbolBolt, theme.InfoStyle.Render(tip.command))
		}
		fmt.Println()
	}
}

type ActionableSuggestion struct {
	priority    int
	title       string
	description string
	command     string
	why         string
}

func getActionableSuggestions(analysis *SetupAnalysis, installed map[string]bool, allPackages map[string]*installer.Package, collections []config.Collection) []ActionableSuggestion {
	suggestions := []ActionableSuggestion{}

	// Priority 1: No packages at all
	if len(installed) == 0 {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    1,
			title:       "Install core foundation",
			description: "You have nothing installed yet. Start with the core system.",
			command:     "anime install core",
			why:         "Required for all other packages",
		})
		return suggestions
	}

	// Priority 2: Have core but no PyTorch
	if analysis.hasCore && !analysis.hasPyTorch {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    2,
			title:       "Install PyTorch",
			description: "Core is installed. Add PyTorch for AI/ML workloads.",
			command:     "anime install pytorch",
			why:         "Required for ComfyUI, video models, and AI workflows",
		})
	}

	// Priority 3: Have PyTorch but no ComfyUI
	if analysis.hasPyTorch && !analysis.hasComfyUI {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    3,
			title:       "Install ComfyUI",
			description: "PyTorch is ready. Add ComfyUI for image generation.",
			command:     "anime install comfyui",
			why:         "Most popular tool for AI image workflows",
		})
	}

	// Priority 4: Have ComfyUI but no image collections
	if analysis.hasComfyUI && analysis.imageCollections == 0 {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    4,
			title:       "Add an image collection",
			description: "ComfyUI is installed but you have no images to process.",
			command:     "anime collection add photos ~/Pictures",
			why:         "Collections let you batch process images with ComfyUI",
		})
	}

	// Priority 5: Have collections - show what you can DO with them
	if len(collections) > 0 {
		collectionName := collections[0].Name
		collectionType := collections[0].Type

		var action, description, why string
		if collectionType == "image" || collectionType == "mixed" {
			action = fmt.Sprintf("anime collection info %s", collectionName)
			description = fmt.Sprintf("You have '%s' collection. See what you can DO with it.", collectionName)
			why = "Generate videos, upscale, apply styles, batch process, etc."
		} else if collectionType == "video" {
			action = fmt.Sprintf("anime collection info %s", collectionName)
			description = fmt.Sprintf("You have '%s' video collection. See processing options.", collectionName)
			why = "Interpolate FPS, extract frames, upscale, etc."
		} else {
			action = fmt.Sprintf("anime collection info %s", collectionName)
			description = fmt.Sprintf("You have %d collection(s). See what you can do with them.", len(collections))
			why = "Process, enhance, convert, share, and more"
		}

		suggestions = append(suggestions, ActionableSuggestion{
			priority:    5,
			title:       "Process your collection",
			description: description,
			command:     action,
			why:         why,
		})
	}

	// Priority 6: Have everything but no Ollama for LLMs
	if analysis.hasPyTorch && !analysis.hasOllama {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    6,
			title:       "Install Ollama for local LLMs",
			description: "Run language models locally on your server.",
			command:     "anime install ollama && anime install models-small",
			why:         "Chat, code generation, and AI assistance",
		})
	}

	// Priority 7: Have Ollama but no models
	if analysis.hasOllama && !analysis.hasSmallModels && !analysis.hasMediumModels && !analysis.hasLargeModels {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    7,
			title:       "Download LLM models",
			description: "Ollama is installed but you have no models.",
			command:     "anime install models-small",
			why:         "Start with small 7-8B models (fastest)",
		})
	}

	// Priority 8: Have everything, suggest video models
	if analysis.hasComfyUI && len(analysis.videoModels) == 0 {
		suggestions = append(suggestions, ActionableSuggestion{
			priority:    8,
			title:       "Try video generation",
			description: "You have image generation working. Add video models.",
			command:     "anime install mochi",
			why:         "Mochi-1: Open source video generation (10B params)",
		})
	}

	// Sort by priority
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].priority < suggestions[j].priority
	})

	// Return top 3 most relevant
	if len(suggestions) > 3 {
		return suggestions[:3]
	}
	return suggestions
}

func displayActionableSuggestions(suggestions []ActionableSuggestion) {
	fmt.Println(theme.InfoStyle.Render("  📋 AVAILABLE COMMANDS"))
	fmt.Println()

	// System & Monitoring
	fmt.Println(theme.HighlightStyle.Render("  System & Monitoring:"))
	fmt.Printf("    %s - Check server status and installed packages\n", theme.InfoStyle.Render("anime status"))
	fmt.Printf("    %s - View GPU metrics and cost tracking\n", theme.InfoStyle.Render("anime metrics"))
	fmt.Printf("    %s - Browse all available packages\n", theme.InfoStyle.Render("anime packages"))
	fmt.Printf("    %s - Interactive monitoring dashboard\n", theme.InfoStyle.Render("anime workstation"))
	fmt.Println()

	// Package Management
	fmt.Println(theme.HighlightStyle.Render("  Package Management:"))
	fmt.Printf("    %s - Install a package\n", theme.InfoStyle.Render("anime install <package-id>"))
	fmt.Printf("    %s - Install multiple packages in parallel\n", theme.InfoStyle.Render("anime parallelize"))
	fmt.Printf("    %s - Interactive setup wizard\n", theme.InfoStyle.Render("anime wizard"))
	fmt.Println()

	// Collections
	fmt.Println(theme.HighlightStyle.Render("  Collections:"))
	fmt.Printf("    %s - List all collections\n", theme.InfoStyle.Render("anime collection list"))
	fmt.Printf("    %s - Create new collection\n", theme.InfoStyle.Render("anime collection create <name> <path>"))
	fmt.Printf("    %s - Collection info and workflows\n", theme.InfoStyle.Render("anime collection info <name>"))
	fmt.Printf("    %s - Animate collection (image→video)\n", theme.InfoStyle.Render("anime collection animate <name>"))
	fmt.Printf("    %s - Upscale collection\n", theme.InfoStyle.Render("anime collection upscale <name>"))
	fmt.Printf("    %s - Browse collection visually\n", theme.InfoStyle.Render("anime browse <name>"))
	fmt.Println()

	// LLM & AI
	fmt.Println(theme.HighlightStyle.Render("  LLM & AI:"))
	fmt.Printf("    %s - List available models\n", theme.InfoStyle.Render("anime ollama list"))
	fmt.Printf("    %s - Chat with a model\n", theme.InfoStyle.Render("anime ollama run <model>"))
	fmt.Printf("    %s - Run model in background\n", theme.InfoStyle.Render("anime ollama start <model>"))
	fmt.Printf("    %s - AI-powered LLM chat interface\n", theme.InfoStyle.Render("anime llm <prompt>"))
	fmt.Println()

	// Video Generation
	fmt.Println(theme.HighlightStyle.Render("  Video Generation:"))
	fmt.Printf("    %s - View available video models\n", theme.InfoStyle.Render("anime models"))
	fmt.Printf("    %s - Run video generation workflow\n", theme.InfoStyle.Render("anime workflow <name>"))
	fmt.Printf("    %s - Browse workflows visually\n", theme.InfoStyle.Render("anime explore"))
	fmt.Println()

	// Development
	fmt.Println(theme.HighlightStyle.Render("  Development:"))
	fmt.Printf("    %s - SSH into server\n", theme.InfoStyle.Render("anime ssh"))
	fmt.Printf("    %s - Deploy anime to server\n", theme.InfoStyle.Render("anime push"))
	fmt.Printf("    %s - Configure server settings\n", theme.InfoStyle.Render("anime config"))
	fmt.Printf("    %s - Run diagnostic checks\n", theme.InfoStyle.Render("anime doctor"))
	fmt.Println()

	// Priority suggestions
	if len(suggestions) > 0 {
		fmt.Println(theme.HighlightStyle.Render("  🎯 Recommended Next Steps:"))
		for i, sugg := range suggestions {
			if i >= 3 {
				break
			}
			fmt.Printf("    %d. %s - %s\n", i+1, theme.InfoStyle.Render(sugg.command), theme.DimTextStyle.Render(sugg.description))
		}
		fmt.Println()
	}

	fmt.Println(theme.DimTextStyle.Render("  💡 Tip: Use ") + theme.HighlightStyle.Render("anime <command> --help") + theme.DimTextStyle.Render(" for detailed help"))
	fmt.Println()
}

// SuggestedWorkflow represents a suggested workflow
type SuggestedWorkflow struct {
	name         string
	description  string
	emoji        string
	command      string
	requirements string
}

func getWorkflowSuggestions(analysis *SetupAnalysis) []SuggestedWorkflow {
	workflows := []SuggestedWorkflow{}

	// Text generation workflows
	if analysis.hasOllama && analysis.hasSmallModels {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Interactive Chat with Small Models",
			description: "Fast, efficient chat with 7-8B parameter models",
			emoji:       "💬",
			command:     "ollama run llama3.3:8b",
			requirements: "Ollama + Small Models",
		})
	}

	if analysis.hasOllama && analysis.hasMediumModels {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Advanced Reasoning with Medium Models",
			description: "Balanced performance with 14-34B parameter models",
			emoji:       "🧠",
			command:     "ollama run qwen3:14b",
			requirements: "Ollama + Medium Models",
		})
	}

	if analysis.hasOllama && analysis.hasLargeModels {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Expert-Level Inference with Large Models",
			description: "Best quality responses with 70B+ parameter models",
			emoji:       "🚀",
			command:     "ollama run llama3.3:70b",
			requirements: "Ollama + Large Models",
		})
	}

	// ComfyUI workflows
	if analysis.hasComfyUI {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Image Generation with ComfyUI",
			description: "Create images using Stable Diffusion workflows",
			emoji:       "🎨",
			command:     "cd ~/ComfyUI && python main.py",
			requirements: "ComfyUI",
		})
	}

	// Video generation workflows
	if suggestContains(analysis.videoModels, "svd") && analysis.hasCollections && analysis.imageCollections > 0 {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Image-to-Video with Stable Video Diffusion",
			description: "Convert your image collections to videos",
			emoji:       "🎬",
			command:     "anime collection list  # then use with SVD in ComfyUI",
			requirements: "SVD + Image Collections",
		})
	}

	if suggestContains(analysis.videoModels, "mochi") {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Text-to-Video with Mochi-1",
			description: "Generate videos from text prompts (10B param model)",
			emoji:       "🎥",
			command:     "cd ~/video-models/mochi-1 && python generate.py",
			requirements: "Mochi-1",
		})
	}

	if suggestContains(analysis.videoModels, "animatediff") && analysis.hasCollections {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Animate Images with AnimateDiff",
			description: "Add motion to static images from your collections",
			emoji:       "✨",
			command:     "Use AnimateDiff workflows in ComfyUI",
			requirements: "AnimateDiff + Collections",
		})
	}

	if suggestContains(analysis.videoModels, "cogvideo") {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "High-Quality Text-to-Video with CogVideoX",
			description: "Generate cinematic videos from text descriptions",
			emoji:       "🎞️",
			command:     "cd ~/video-models/cogvideo && python inference.py",
			requirements: "CogVideoX-5B",
		})
	}

	// Development workflows
	if analysis.hasClaude {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "AI-Assisted Development with Claude Code",
			description: "Code with AI assistance directly in your terminal",
			emoji:       "👨‍💻",
			command:     "claude",
			requirements: "Claude Code CLI",
		})
	}

	if analysis.hasPyTorch && analysis.hasOllama {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Model Fine-Tuning Pipeline",
			description: "Fine-tune LLMs on custom datasets",
			emoji:       "🔧",
			command:     "Use PyTorch with transformers library",
			requirements: "PyTorch + Ollama",
		})
	}

	// Collection-based workflows
	if analysis.hasCollections && !analysis.hasComfyUI && !hasVideoModels(analysis) {
		workflows = append(workflows, SuggestedWorkflow{
			name:        "Asset Collection Processing",
			description: "Install ComfyUI or video models to process your collections",
			emoji:       "📁",
			requirements: "Install image/video generation packages",
		})
	}

	return workflows
}

// PackageSuggestion represents a suggested package to install
type PackageSuggestion struct {
	packageID string
	reason    string
	emoji     string
	priority  int // higher = more important
}

func getPackageSuggestions(analysis *SetupAnalysis, installed map[string]bool, allPackages map[string]*installer.Package) []PackageSuggestion {
	suggestions := []PackageSuggestion{}

	// Foundation packages
	if !analysis.hasCore {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "core",
			reason:    "Essential foundation - required for all other packages",
			emoji:     "🏗️",
			priority:  100,
		})
	}

	if analysis.hasCore && !analysis.hasPython {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "python",
			reason:    "Required for AI/ML workloads",
			emoji:     "🐍",
			priority:  90,
		})
	}

	// LLM suggestions
	if analysis.hasCore && !analysis.hasOllama && !analysis.hasPyTorch {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "ollama",
			reason:    "Start running LLMs locally with minimal setup",
			emoji:     "🔮",
			priority:  80,
		})
	}

	if analysis.hasOllama && !analysis.hasSmallModels && !analysis.hasMediumModels && !analysis.hasLargeModels {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "models-small",
			reason:    "Fast inference with 7-8B models - great starting point",
			emoji:     "⭐",
			priority:  85,
		})
	}

	if analysis.hasOllama && analysis.hasSmallModels && !analysis.hasMediumModels {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "models-medium",
			reason:    "Upgrade to better reasoning with 14-34B models",
			emoji:     "⭐",
			priority:  60,
		})
	}

	if analysis.hasOllama && analysis.hasMediumModels && !analysis.hasLargeModels {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "models-large",
			reason:    "Best quality inference with 70B+ models",
			emoji:     "⭐",
			priority:  40,
		})
	}

	// ML Framework suggestions
	if analysis.hasPython && !analysis.hasPyTorch && len(analysis.videoModels) == 0 {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "pytorch",
			reason:    "Unlock video generation and model training capabilities",
			emoji:     "🤖",
			priority:  75,
		})
	}

	// Creative tools suggestions
	if analysis.hasPyTorch && !analysis.hasComfyUI && len(analysis.videoModels) == 0 {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "comfyui",
			reason:    "Start creating images and videos with Stable Diffusion",
			emoji:     "🎨",
			priority:  70,
		})
	}

	// Video model suggestions based on what's already installed
	if analysis.hasComfyUI && !suggestContains(analysis.videoModels, "svd") && !installed["svd"] {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "svd",
			reason:    "Convert images to videos in ComfyUI",
			emoji:     "🎬",
			priority:  65,
		})
	}

	if analysis.hasComfyUI && !suggestContains(analysis.videoModels, "animatediff") && !installed["animatediff"] {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "animatediff",
			reason:    "Animate your images with motion modules",
			emoji:     "✨",
			priority:  55,
		})
	}

	if analysis.hasPyTorch && !suggestContains(analysis.videoModels, "mochi") && !installed["mochi"] {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "mochi",
			reason:    "Fast text-to-video generation (10B params)",
			emoji:     "🎥",
			priority:  60,
		})
	}

	if analysis.hasPyTorch && suggestContains(analysis.videoModels, "mochi") && !suggestContains(analysis.videoModels, "cogvideo") && !installed["cogvideo"] {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "cogvideo",
			reason:    "Higher quality video generation alternative",
			emoji:     "🎞️",
			priority:  50,
		})
	}

	// Developer tools
	if analysis.hasCore && !analysis.hasClaude {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "claude",
			reason:    "AI-powered coding assistant in your terminal",
			emoji:     "👨‍💻",
			priority:  45,
		})
	}

	// Collection-based suggestions
	if analysis.hasCollections && analysis.imageCollections > 0 && !analysis.hasComfyUI {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "comfyui",
			reason:    "Process your image collections with AI workflows",
			emoji:     "🎨",
			priority:  75,
		})
	}

	if analysis.hasCollections && analysis.imageCollections > 0 && analysis.hasComfyUI && !suggestContains(analysis.videoModels, "svd") {
		suggestions = append(suggestions, PackageSuggestion{
			packageID: "svd",
			reason:    "Turn your image collections into videos",
			emoji:     "🎬",
			priority:  70,
		})
	}

	// Sort by priority (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].priority > suggestions[j].priority
	})

	// Limit to top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

// OptimizationTip represents a tip for better usage
type OptimizationTip struct {
	title       string
	description string
	command     string
}

func getOptimizationTips(analysis *SetupAnalysis) []OptimizationTip {
	tips := []OptimizationTip{}

	// No collections but have image/video tools
	if !analysis.hasCollections && (analysis.hasComfyUI || len(analysis.videoModels) > 0) {
		tips = append(tips, OptimizationTip{
			title:       "Create Asset Collections",
			description: "Organize your images/videos into collections for easier workflow management",
			command:     "anime collection create <name> <path>",
		})
	}

	// Has collections but no processing tools
	if analysis.hasCollections && !analysis.hasComfyUI && len(analysis.videoModels) == 0 {
		tips = append(tips, OptimizationTip{
			title:       "Install Processing Tools",
			description: "Your collections are ready - install ComfyUI or video models to process them",
			command:     "anime install comfyui",
		})
	}

	// Has Ollama but no models
	if analysis.hasOllama && !analysis.hasSmallModels && !analysis.hasMediumModels && !analysis.hasLargeModels {
		tips = append(tips, OptimizationTip{
			title:       "Install LLM Models",
			description: "Ollama is installed but you need models to run inference",
			command:     "anime install models-small",
		})
	}

	// Has basic setup, suggest parallel installation
	if analysis.hasCore && !analysis.hasOllama && !analysis.hasPyTorch {
		tips = append(tips, OptimizationTip{
			title:       "Use Parallel Installation",
			description: "Install multiple packages simultaneously for faster setup",
			command:     "anime parallelize",
		})
	}

	// Has video models, suggest exploring more
	if len(analysis.videoModels) >= 1 && len(analysis.videoModels) < 3 {
		tips = append(tips, OptimizationTip{
			title:       "Explore More Video Models",
			description: "Try different video generation models for varied output styles",
			command:     "anime packages | grep 'Video Generation'",
		})
	}

	// Setup wizard suggestion for incomplete setups
	if !analysis.hasCore || (!analysis.hasOllama && !analysis.hasPyTorch) {
		tips = append(tips, OptimizationTip{
			title:       "Use the Setup Wizard",
			description: "Interactive wizard helps you configure the perfect setup for your use case",
			command:     "anime wizard",
		})
	}

	// Monitoring suggestion if they have a production setup
	if analysis.hasOllama && (analysis.hasMediumModels || analysis.hasLargeModels) {
		tips = append(tips, OptimizationTip{
			title:       "Monitor Your Server",
			description: "Track GPU usage, costs, and performance metrics in real-time",
			command:     "anime metrics",
		})
	}

	return tips
}

// Helper functions

func suggestContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasVideoModels(analysis *SetupAnalysis) bool {
	return len(analysis.videoModels) > 0
}

// Reuse checkPackageInstalled from packages.go
func getInstalledPackagesForSuggest() map[string]bool {
	installed := make(map[string]bool)

	cfg, err := config.Load()
	if err != nil {
		return installed
	}

	// Get lambda server
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	if lambdaTarget == "" {
		return installed
	}

	// Parse target
	var user, host string
	if strings.Contains(lambdaTarget, "@") {
		parts := strings.SplitN(lambdaTarget, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = lambdaTarget
	}

	// Connect and check
	client, err := ssh.NewClient(host, user, "")
	if err != nil {
		return installed
	}
	defer client.Close()

	// Check each package
	packages := installer.GetPackages()
	for pkgID := range packages {
		if checkPackageInstalled(client, pkgID) {
			installed[pkgID] = true
		}
	}

	return installed
}
