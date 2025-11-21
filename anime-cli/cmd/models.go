package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "List all downloaded models or show model catalog",
		Long: `List all AI models downloaded on the server or show comprehensive model catalog.

Scans for common model formats:
  • .safetensors (Stable Diffusion, LoRA, etc.)
  • .ckpt (Checkpoints)
  • .pth (PyTorch models)
  • .bin (Binary models)

Shows model locations, sizes, and categories.

Use --catalog to see all available models with descriptions and use cases.`,
		RunE: runModels,
	}

	modelsLocal   bool
	modelsCatalog bool
)

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.Flags().BoolVarP(&modelsLocal, "local", "l", false, "Scan local filesystem instead of remote server")
	modelsCmd.Flags().BoolVarP(&modelsCatalog, "catalog", "c", false, "Show comprehensive catalog of available models")
}

type ModelFile struct {
	Name      string
	Path      string
	SizeMB    float64
	ModelType string // LLM, Image/Video, Audio, etc.
	Category  string // Specific category within the type
	FileType  string // File extension
}

func runModels(cmd *cobra.Command, args []string) error {
	// Check if user wants to list installable models
	if len(args) > 0 && args[0] == "list" {
		return showInstallableModels()
	}

	// Default: Launch TUI for catalog
	if !modelsLocal && !modelsCatalog {
		// Launch interactive TUI
		m := tui.NewModelsModel()
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 AI MODELS 🎨"))
	fmt.Println()

	// Show catalog if requested
	if modelsCatalog {
		return showModelCatalog()
	}

	var models []ModelFile
	var err error

	if modelsLocal {
		fmt.Println(theme.InfoStyle.Render("Scanning local filesystem..."))
		fmt.Println()
		models, err = scanLocalModels()
	} else {
		// Load config and connect to remote server
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		lambdaTarget := cfg.GetAlias("lambda")
		if lambdaTarget == "" {
			if server, err := cfg.GetServer("lambda"); err == nil {
				lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
			}
		}

		if lambdaTarget == "" {
			fmt.Println(theme.WarningStyle.Render("⚠️  No remote Lambda server configured"))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("Options:"))
			fmt.Println(theme.HighlightStyle.Render("  anime models --local"))
			fmt.Println(theme.DimTextStyle.Render("  or configure a server: anime config"))
			fmt.Println()
			return nil
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

		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Scanning remote server: %s...", host)))
		fmt.Println()

		// Connect to server
		sshClient, err := ssh.NewClient(host, user, "")
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer sshClient.Close()

		models, err = scanRemoteModels(sshClient)
	}

	if err != nil {
		return fmt.Errorf("failed to scan models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println(theme.WarningStyle.Render("No models found"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Models are typically stored in:"))
		fmt.Println(theme.DimTextStyle.Render("  ~/ComfyUI/models/"))
		fmt.Println(theme.DimTextStyle.Render("  ~/.cache/huggingface/"))
		fmt.Println()
		return nil
	}

	// Group models by type
	types := groupModelsByType(models)

	// Display results
	displayModels(types)

	return nil
}

func scanLocalModels() ([]ModelFile, error) {
	var models []ModelFile

	// Common model locations
	searchPaths := []string{
		filepath.Join(homeDir(), "ComfyUI", "models"),
		filepath.Join(homeDir(), ".cache", "huggingface"),
		filepath.Join(homeDir(), "models"),
	}

	for _, searchPath := range searchPaths {
		// Find model files
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

	// Search for model files
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

		// Get file size
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
	// Get file size
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

	// Detect LLMs first
	// Check for Ollama models
	if strings.Contains(lowerPath, "ollama") || strings.Contains(lowerPath, ".ollama") {
		return "LLM", detectLLMCategory(lowerPath)
	}

	// Check for common LLM patterns in both path and filename
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

	// Check for transformer/language model directories
	if strings.Contains(lowerPath, "transformers") ||
		strings.Contains(lowerPath, "language") ||
		strings.Contains(lowerPath, "/llm") {
		return "LLM", detectLLMCategory(lowerPath)
	}

	// Image/Video Generation Models
	// Check by directory structure
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

	// Check by filename patterns for image models
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

	// Check for ComfyUI models
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

	// Try to detect model size and type
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

	// Detect by model family
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
	// Two-level grouping: ModelType -> Category -> Models
	types := make(map[string]map[string][]ModelFile)

	for _, model := range models {
		if types[model.ModelType] == nil {
			types[model.ModelType] = make(map[string][]ModelFile)
		}
		types[model.ModelType][model.Category] = append(types[model.ModelType][model.Category], model)
	}

	// Sort models within each category by size (largest first)
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
	// Get sorted type names (LLM first, then Image/Video, then Audio, then Other)
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

	// Model type emoji mapping
	typeEmojis := map[string]string{
		"LLM":         "🤖",
		"Image/Video": "🎨",
		"Audio":       "🎵",
		"Other":       "📦",
	}

	for _, modelType := range sortedTypes {
		categories := types[modelType]
		totalCategories += len(categories)

		// Calculate type totals
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

		// Type header
		emoji := typeEmojis[modelType]
		fmt.Println(theme.SuccessStyle.Render("╔══════════════════════════════════════════════════════════════════════════╗"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("║  %s  %s", emoji, modelType)))
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("║  %d models • %.1f GB • %d categories", typeModels, typeSizeGB, len(categories))))
		fmt.Println(theme.SuccessStyle.Render("╚══════════════════════════════════════════════════════════════════════════╝"))
		fmt.Println()

		// Get sorted category names within this type
		var categoryNames []string
		for name := range categories {
			categoryNames = append(categoryNames, name)
		}
		sort.Strings(categoryNames)

		// Display each category
		for _, category := range categoryNames {
			models := categories[category]

			categorySize := 0.0
			for _, m := range models {
				categorySize += m.SizeMB
			}

			// Category header
			fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  📁 %s (%d models, %.1f GB)", category, len(models), categorySize/1024.0)))
			fmt.Println(theme.DimTextStyle.Render("  ────────────────────────────────────────────────────────────────"))
			fmt.Println()

			// List models
			for _, model := range models {
				sizeStr := formatSize(model.SizeMB)
				fmt.Printf("    %s %s\n",
					theme.HighlightStyle.Render(model.Name),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s)", sizeStr)))
				fmt.Printf("      %s\n", theme.DimTextStyle.Render(model.Path))
				fmt.Println()
			}
		}

		fmt.Println()
	}

	// Summary
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s  %s  %s\n",
		theme.SuccessStyle.Render(fmt.Sprintf("📦 Total: %d models", totalModels)),
		theme.SuccessStyle.Render(fmt.Sprintf("💾 %.2f GB", totalSizeGB)),
		theme.DimTextStyle.Render(fmt.Sprintf("(%d types, %d categories)", len(sortedTypes), totalCategories)))
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
}

func formatSize(sizeMB float64) string {
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

type ModelCatalogEntry struct {
	Name        string
	Size        string
	Description string
	UseCases    []string
	Category    string
}

func showModelCatalog() error {
	catalog := []struct {
		Category string
		Emoji    string
		Models   []ModelCatalogEntry
	}{
		{
			Category: "🤖 Large Language Models (LLMs)",
			Emoji:    "🤖",
			Models: []ModelCatalogEntry{
				{
					Name:        "Llama 3.3 70B",
					Size:        "~40GB",
					Description: "Meta's latest open-source flagship model with exceptional reasoning and coding capabilities",
					UseCases:    []string{"Code generation", "Complex reasoning", "Creative writing", "Research assistance"},
					Category:    "General Purpose",
				},
				{
					Name:        "Llama 3.3 8B",
					Size:        "~5GB",
					Description: "Efficient smaller version of Llama 3.3, great balance of performance and speed",
					UseCases:    []string{"Fast inference", "Chatbots", "Text summarization", "Simple coding tasks"},
					Category:    "Efficient",
				},
				{
					Name:        "Mistral 7B",
					Size:        "~4GB",
					Description: "High-performance 7B model outperforming many larger models, excellent for coding",
					UseCases:    []string{"Code completion", "Technical writing", "Quick Q&A", "Function generation"},
					Category:    "Coding",
				},
				{
					Name:        "Mixtral 8x7B",
					Size:        "~26GB",
					Description: "Mixture of Experts model with 47B parameters, runs efficiently via sparse activation",
					UseCases:    []string{"Multi-task processing", "Complex instructions", "Long context", "Specialized domains"},
					Category:    "Multi-Task",
				},
				{
					Name:        "Qwen 2.5 72B",
					Size:        "~42GB",
					Description: "Alibaba's top model with strong multilingual and math capabilities",
					UseCases:    []string{"Multilingual tasks", "Mathematics", "Science", "International content"},
					Category:    "Multilingual",
				},
				{
					Name:        "Qwen 2.5 14B",
					Size:        "~8GB",
					Description: "Mid-size Qwen model with excellent Chinese-English bilingual performance",
					UseCases:    []string{"Bilingual applications", "Translation", "Cross-language tasks"},
					Category:    "Bilingual",
				},
				{
					Name:        "DeepSeek Coder 33B",
					Size:        "~18GB",
					Description: "Specialized coding model trained on 2T+ tokens of code and text",
					UseCases:    []string{"Code generation", "Bug fixing", "Code review", "Documentation"},
					Category:    "Code Specialist",
				},
				{
					Name:        "DeepSeek V3",
					Size:        "~250GB",
					Description: "Latest frontier model with 671B parameters using MoE architecture",
					UseCases:    []string{"Research", "Complex reasoning", "Frontier capabilities", "Benchmarks"},
					Category:    "Frontier",
				},
				{
					Name:        "Phi-3.5 Mini (3.8B)",
					Size:        "~2GB",
					Description: "Microsoft's compact model with strong reasoning despite small size",
					UseCases:    []string{"Edge deployment", "Mobile apps", "Resource-constrained environments"},
					Category:    "Compact",
				},
			},
		},
		{
			Category: "🎨 Image Generation Models",
			Emoji:    "🎨",
			Models: []ModelCatalogEntry{
				{
					Name:        "Stable Diffusion XL (SDXL)",
					Size:        "~7GB",
					Description: "Latest Stable Diffusion with improved image quality and composition",
					UseCases:    []string{"High-quality images", "Concept art", "Product visualization", "Marketing materials"},
					Category:    "General Purpose",
				},
				{
					Name:        "Stable Diffusion 1.5",
					Size:        "~4GB",
					Description: "Widely-used base model with huge ecosystem of fine-tunes and LoRAs",
					UseCases:    []string{"Art generation", "Style transfer", "Photo editing", "Custom training"},
					Category:    "Classic",
				},
				{
					Name:        "Flux.1 Dev",
					Size:        "~12GB",
					Description: "Black Forest Labs' new model with exceptional prompt following and quality",
					UseCases:    []string{"Photorealism", "Typography", "Complex compositions", "Professional work"},
					Category:    "Professional",
				},
				{
					Name:        "Flux.1 Schnell",
					Size:        "~12GB",
					Description: "Fast version of Flux optimized for speed while maintaining quality",
					UseCases:    []string{"Rapid prototyping", "Iterative design", "Real-time generation"},
					Category:    "Fast",
				},
			},
		},
		{
			Category: "🎬 Video Generation Models",
			Emoji:    "🎬",
			Models: []ModelCatalogEntry{
				{
					Name:        "Stable Video Diffusion",
					Size:        "~10GB",
					Description: "Stability AI's image-to-video model for smooth animations",
					UseCases:    []string{"Product demos", "Animation", "Visual effects", "Motion design"},
					Category:    "Image-to-Video",
				},
				{
					Name:        "AnimateDiff",
					Size:        "~4GB",
					Description: "Motion module for Stable Diffusion, animates still images",
					UseCases:    []string{"Character animation", "Scene transitions", "Loop creation"},
					Category:    "Animation",
				},
				{
					Name:        "Mochi-1",
					Size:        "~12GB",
					Description: "Open source video generation model with 10B parameters",
					UseCases:    []string{"Text-to-video", "Creative videos", "Experimental content"},
					Category:    "Text-to-Video",
				},
				{
					Name:        "CogVideoX-5B",
					Size:        "~14GB",
					Description: "Open source text-to-video with strong temporal consistency",
					UseCases:    []string{"Video creation", "Content production", "Social media"},
					Category:    "Content Creation",
				},
				{
					Name:        "Open-Sora 2.0",
					Size:        "~16GB",
					Description: "High-quality video generation with advanced architectures",
					UseCases:    []string{"High-res video", "Professional content", "Research"},
					Category:    "High Quality",
				},
				{
					Name:        "LTXVideo",
					Size:        "~7GB",
					Description: "Fast video generation using latent transformers",
					UseCases:    []string{"Quick videos", "Rapid iteration", "Previews"},
					Category:    "Fast Generation",
				},
				{
					Name:        "Wan2.2",
					Size:        "~10GB",
					Description: "State-of-the-art image-to-video with exceptional quality",
					UseCases:    []string{"Professional videos", "Cinematic effects", "High-quality output"},
					Category:    "Professional",
				},
				{
					Name:        "Topaz Video AI",
					Size:        "~5GB",
					Description: "Commercial video enhancement and upscaling tool using AI",
					UseCases:    []string{"Video upscaling", "Frame interpolation", "Denoising", "Quality enhancement"},
					Category:    "Enhancement",
				},
			},
		},
	}

	fmt.Println(theme.InfoStyle.Render("📚 Comprehensive AI Model Catalog"))
	fmt.Println(theme.DimTextStyle.Render("Popular models across different categories with descriptions and use cases"))
	fmt.Println()

	for _, section := range catalog {
		// Section header
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("%s  %s", section.Emoji, section.Category)))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, model := range section.Models {
			// Model name and category
			fmt.Printf("  %s %s %s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(model.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("[%s]", model.Category)))

			// Size
			fmt.Printf("    💾 %s\n", theme.InfoStyle.Render(model.Size))

			// Description
			fmt.Printf("    📝 %s\n", theme.SecondaryTextStyle.Render(model.Description))

			// Use cases
			if len(model.UseCases) > 0 {
				fmt.Printf("    🎯 %s\n", theme.DimTextStyle.Render("Use Cases:"))
				for _, useCase := range model.UseCases {
					fmt.Printf("       • %s\n", theme.DimTextStyle.Render(useCase))
				}
			}

			fmt.Println()
		}
	}

	// Installation instructions
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📥 Installation"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("  Install Ollama LLMs:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("ollama pull llama3.3:70b"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("ollama pull mistral"))
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("  Install via anime packages:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install models-small"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  For downloaded models, run: anime models"))
	fmt.Println()

	return nil
}

func showInstallableModels() error {
	// Import from installer package
	type Package struct {
		ID           string
		Name         string
		Description  string
		Category     string
		Size         string
		EstimatedTime string
	}

	packages := []Package{
		// LLM Models
		{ID: "llama-3.3-70b", Name: "Llama 3.3 70B", Description: "Meta's latest open-source flagship model with exceptional reasoning and coding capabilities", Category: "LLM", Size: "~40GB", EstimatedTime: "30m"},
		{ID: "llama-3.3-8b", Name: "Llama 3.3 8B", Description: "Efficient smaller version of Llama 3.3, great balance of performance and speed", Category: "LLM", Size: "~5GB", EstimatedTime: "5m"},
		{ID: "mistral", Name: "Mistral 7B", Description: "High-performance 7B model outperforming many larger models, excellent for coding", Category: "LLM", Size: "~4GB", EstimatedTime: "4m"},
		{ID: "mixtral", Name: "Mixtral 8x7B", Description: "Mixture of Experts model with 47B parameters, runs efficiently via sparse activation", Category: "LLM", Size: "~26GB", EstimatedTime: "20m"},
		{ID: "qwen-2.5-72b", Name: "Qwen 2.5 72B", Description: "Alibaba's top model with strong multilingual and math capabilities", Category: "LLM", Size: "~42GB", EstimatedTime: "35m"},
		{ID: "qwen-2.5-14b", Name: "Qwen 2.5 14B", Description: "Mid-size Qwen model with excellent Chinese-English bilingual performance", Category: "LLM", Size: "~8GB", EstimatedTime: "8m"},
		{ID: "qwen-2.5-7b", Name: "Qwen 2.5 7B", Description: "Compact Qwen model with strong multilingual support", Category: "LLM", Size: "~4GB", EstimatedTime: "5m"},
		{ID: "deepseek-coder-33b", Name: "DeepSeek Coder 33B", Description: "Specialized coding model trained on 2T+ tokens of code and text", Category: "LLM", Size: "~18GB", EstimatedTime: "15m"},
		{ID: "deepseek-v3", Name: "DeepSeek V3", Description: "Latest frontier model with 671B parameters using MoE architecture", Category: "LLM", Size: "~250GB", EstimatedTime: "2h"},
		{ID: "phi-3.5", Name: "Phi-3.5 Mini (3.8B)", Description: "Microsoft's compact model with strong reasoning despite small size", Category: "LLM", Size: "~2GB", EstimatedTime: "2m"},

		// Image Generation Models
		{ID: "sdxl", Name: "Stable Diffusion XL", Description: "Latest Stable Diffusion with improved image quality and composition", Category: "Image Generation", Size: "~7GB", EstimatedTime: "10m"},
		{ID: "sd15", Name: "Stable Diffusion 1.5", Description: "Widely-used base model with huge ecosystem of fine-tunes and LoRAs", Category: "Image Generation", Size: "~4GB", EstimatedTime: "6m"},
		{ID: "flux-dev", Name: "Flux.1 Dev", Description: "Black Forest Labs' new model with exceptional prompt following and quality", Category: "Image Generation", Size: "~12GB", EstimatedTime: "15m"},
		{ID: "flux-schnell", Name: "Flux.1 Schnell", Description: "Fast version of Flux optimized for speed while maintaining quality", Category: "Image Generation", Size: "~12GB", EstimatedTime: "15m"},

		// Video Generation Models
		{ID: "mochi", Name: "Mochi-1", Description: "Open source video generation model, 10B params", Category: "Video Generation", Size: "~12GB", EstimatedTime: "20m"},
		{ID: "svd", Name: "Stable Video Diffusion", Description: "Stability AI's video diffusion model for ComfyUI", Category: "Video Generation", Size: "~8GB", EstimatedTime: "15m"},
		{ID: "animatediff", Name: "AnimateDiff", Description: "Motion module for Stable Diffusion, animates images", Category: "Video Generation", Size: "~4GB", EstimatedTime: "10m"},
		{ID: "cogvideo", Name: "CogVideoX-5B", Description: "Open source text-to-video model", Category: "Video Generation", Size: "~14GB", EstimatedTime: "25m"},
		{ID: "opensora", Name: "Open-Sora 2.0", Description: "High-quality video generation model", Category: "Video Generation", Size: "~16GB", EstimatedTime: "30m"},
		{ID: "ltxvideo", Name: "LTXVideo", Description: "Fast video generation with latent transformers", Category: "Video Generation", Size: "~7GB", EstimatedTime: "15m"},
		{ID: "wan2", Name: "Wan2.2", Description: "State-of-the-art image-to-video generation model", Category: "Video Generation", Size: "~10GB", EstimatedTime: "20m"},
	}

	// Group by category
	categories := make(map[string][]Package)
	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort category names
	categoryOrder := []string{"LLM", "Image Generation", "Video Generation"}

	fmt.Println()
	fmt.Println(theme.RenderBanner("📦 INSTALLABLE AI MODELS 📦"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Install any model with: anime install <model-id>"))
	fmt.Println()

	// Category emoji mapping
	categoryEmojis := map[string]string{
		"LLM":              "🤖",
		"Image Generation": "🎨",
		"Video Generation": "🎬",
	}

	totalModels := 0
	for _, category := range categoryOrder {
		models := categories[category]
		if len(models) == 0 {
			continue
		}

		// Sort models by name
		sort.Slice(models, func(i, j int) bool {
			return models[i].Name < models[j].Name
		})

		emoji := categoryEmojis[category]
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("%s  %s (%d models)", emoji, category, len(models))))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, model := range models {
			fmt.Printf("  %s %s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(model.Name))
			fmt.Printf("    💾 %s  ⏱️  %s\n",
				theme.InfoStyle.Render(model.Size),
				theme.DimTextStyle.Render(model.EstimatedTime))
			fmt.Printf("    📝 %s\n",
				theme.SecondaryTextStyle.Render(model.Description))
			fmt.Printf("    ⚡ %s\n",
				theme.DimTextStyle.Render("anime install "+model.ID))
			fmt.Println()
		}

		totalModels += len(models)
	}

	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("📦 Total: %d installable models", totalModels)))
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	return nil
}
