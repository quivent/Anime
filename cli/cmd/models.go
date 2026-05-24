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
				sizeStr := formatSizeMB(model.SizeMB)
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

func formatSizeMB(sizeMB float64) string {
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
					Name:        "Qwen3 235B MoE",
					Size:        "~42GB",
					Description: "Alibaba's top model with strong multilingual and math capabilities",
					UseCases:    []string{"Multilingual tasks", "Mathematics", "Science", "International content"},
					Category:    "Multilingual",
				},
				{
					Name:        "Qwen3 14B",
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
					Name:        "Flux 2 (FP8)",
					Size:        "~8GB",
					Description: "Next-generation video model from Black Forest Labs with superior motion and coherence (FP8 quantized)",
					UseCases:    []string{"High-quality video generation", "Complex motion", "Long-form content", "Professional production"},
					Category:    "Next Generation",
				},
				{
					Name:        "CogVideoX 1.5 5B",
					Size:        "~18GB",
					Description: "Upgraded CogVideoX supporting 10-second videos at higher resolutions with exceptional temporal consistency",
					UseCases:    []string{"Long-form video", "High-resolution output", "Content production", "Social media"},
					Category:    "Text-to-Video",
				},
				{
					Name:        "CogVideoX 1.5 I2V",
					Size:        "~18GB",
					Description: "Image-to-video variant of CogVideoX 1.5 with any resolution support",
					UseCases:    []string{"Photo animation", "Image-to-video", "Character motion", "Scene animation"},
					Category:    "Image-to-Video",
				},
				{
					Name:        "HunyuanVideo",
					Size:        "~20GB",
					Description: "Tencent's open-source text-to-video diffusion transformer with exceptional motion quality",
					UseCases:    []string{"Text-to-video", "Complex motion", "Cinematic effects", "Research"},
					Category:    "Text-to-Video",
				},
				{
					Name:        "Pyramid Flow",
					Size:        "~12GB",
					Description: "Efficient video generation using pyramidal flow matching (768p, up to 10s)",
					UseCases:    []string{"Efficient generation", "Medium-length videos", "Quick iteration"},
					Category:    "Efficient",
				},
				{
					Name:        "SVD-XT 1.1",
					Size:        "~10GB",
					Description: "Extended Stable Video Diffusion with improved temporal consistency and longer generation",
					UseCases:    []string{"Extended animations", "Product demos", "Visual effects", "Motion design"},
					Category:    "Image-to-Video",
				},
				{
					Name:        "I2V-Adapter",
					Size:        "~4GB",
					Description: "General image-to-video adapter for diffusion models (SIGGRAPH 2024)",
					UseCases:    []string{"Photo animation", "Style transfer", "Any diffusion model", "Research"},
					Category:    "Adapter",
				},
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
			},
		},
		{
			Category: "🔧 Enhancement Models",
			Emoji:    "🔧",
			Models: []ModelCatalogEntry{
				{
					Name:        "Real-ESRGAN",
					Size:        "~200MB",
					Description: "Practical 4x image/video upscaling with artifact removal",
					UseCases:    []string{"Image upscaling", "Video upscaling", "Old photo restoration"},
					Category:    "Upscaler",
				},
				{
					Name:        "GFPGAN",
					Size:        "~350MB",
					Description: "Practical face restoration algorithm for real-world images",
					UseCases:    []string{"Face restoration", "Old photo repair", "Portrait enhancement"},
					Category:    "Face Restoration",
				},
				{
					Name:        "AuraSR",
					Size:        "~500MB",
					Description: "GigaGAN-based open-source 4x image upscaler from Fal.ai",
					UseCases:    []string{"Fast upscaling", "GAN-based enhancement", "Sharp results"},
					Category:    "Upscaler",
				},
				{
					Name:        "SUPIR",
					Size:        "~12GB",
					Description: "Photo-realistic image restoration using SDXL with text-guided enhancement",
					UseCases:    []string{"Photo restoration", "Detail enhancement", "Quality improvement"},
					Category:    "Restoration",
				},
				{
					Name:        "RIFE",
					Size:        "~200MB",
					Description: "Real-time intermediate flow estimation for video frame interpolation",
					UseCases:    []string{"Frame interpolation", "Slow motion", "FPS increase"},
					Category:    "Interpolation",
				},
				{
					Name:        "FILM",
					Size:        "~400MB",
					Description: "Google's frame interpolation model for large motion between frames",
					UseCases:    []string{"Smooth slow-mo", "Large motion handling", "High-quality interpolation"},
					Category:    "Interpolation",
				},
			},
		},
		{
			Category: "🎛️ ControlNet & Adapters",
			Emoji:    "🎛️",
			Models: []ModelCatalogEntry{
				{
					Name:        "ControlNet Canny",
					Size:        "~1.5GB",
					Description: "Edge detection based image conditioning for Stable Diffusion",
					UseCases:    []string{"Edge-guided generation", "Structural control", "Lineart coloring"},
					Category:    "Edge Detection",
				},
				{
					Name:        "ControlNet Depth",
					Size:        "~1.5GB",
					Description: "Depth map conditioning for 3D-aware image generation",
					UseCases:    []string{"3D-aware generation", "Depth control", "Scene composition"},
					Category:    "Depth",
				},
				{
					Name:        "ControlNet OpenPose",
					Size:        "~1.5GB",
					Description: "Human pose estimation conditioning for character generation",
					UseCases:    []string{"Pose control", "Character generation", "Action scenes"},
					Category:    "Pose",
				},
				{
					Name:        "IP-Adapter",
					Size:        "~100MB",
					Description: "Tencent's image prompt adapter for style and content transfer",
					UseCases:    []string{"Style transfer", "Content reference", "Character consistency"},
					Category:    "Style",
				},
				{
					Name:        "IP-Adapter FaceID",
					Size:        "~200MB",
					Description: "Face-specific IP-Adapter for identity-preserving generation",
					UseCases:    []string{"Face consistency", "Portrait generation", "Character sheets"},
					Category:    "Face",
				},
				{
					Name:        "InstantID",
					Size:        "~2GB",
					Description: "Zero-shot identity-preserving generation with single reference image",
					UseCases:    []string{"Identity preservation", "Character consistency", "Portrait variation"},
					Category:    "Identity",
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
		{ID: "qwen3-235b", Name: "Qwen3 235B MoE", Description: "Flagship Qwen3 MoE model (235B total, 22B activated) with advanced reasoning and coding", Category: "LLM", Size: "~142GB", EstimatedTime: "60m"},
		{ID: "qwen3-32b", Name: "Qwen3 32B", Description: "Large dense Qwen3 model with strong multilingual and reasoning capabilities", Category: "LLM", Size: "~20GB", EstimatedTime: "10m"},
		{ID: "qwen3-30b", Name: "Qwen3 30B MoE", Description: "MoE model (30B total, 3B activated) with fast inference and strong capabilities", Category: "LLM", Size: "~19GB", EstimatedTime: "10m"},
		{ID: "qwen3-14b", Name: "Qwen3 14B", Description: "Mid-size Qwen3 model with excellent multilingual performance and reasoning", Category: "LLM", Size: "~9GB", EstimatedTime: "8m"},
		{ID: "qwen3-8b", Name: "Qwen3 8B", Description: "Compact yet powerful Qwen3 model with strong multilingual support", Category: "LLM", Size: "~5GB", EstimatedTime: "5m"},
		{ID: "qwen3-4b", Name: "Qwen3 4B", Description: "Efficient small model with 256K context window, great for edge devices", Category: "LLM", Size: "~2.5GB", EstimatedTime: "3m"},
		{ID: "deepseek-coder-33b", Name: "DeepSeek Coder 33B", Description: "Specialized coding model trained on 2T+ tokens of code and text", Category: "LLM", Size: "~18GB", EstimatedTime: "15m"},
		{ID: "deepseek-v3", Name: "DeepSeek V3", Description: "Latest frontier model with 671B parameters using MoE architecture", Category: "LLM", Size: "~250GB", EstimatedTime: "2h"},
		{ID: "phi-3.5", Name: "Phi-3.5 Mini (3.8B)", Description: "Microsoft's compact model with strong reasoning despite small size", Category: "LLM", Size: "~2GB", EstimatedTime: "2m"},
		{ID: "phi-4", Name: "Phi-4 (14B)", Description: "Microsoft's 14B reasoning model that rivals much larger models on complex tasks", Category: "LLM", Size: "~9GB", EstimatedTime: "8m"},
		{ID: "deepseek-r1-8b", Name: "DeepSeek-R1 8B", Description: "Latest reasoning model with outstanding performance in math, programming, and logic", Category: "LLM", Size: "~5GB", EstimatedTime: "5m"},
		{ID: "deepseek-r1-70b", Name: "DeepSeek-R1 70B", Description: "Large reasoning model approaching O3/Gemini 2.5 Pro level performance", Category: "LLM", Size: "~43GB", EstimatedTime: "30m"},
		{ID: "gemma3-4b", Name: "Gemma3 4B", Description: "Google's multimodal model with vision capabilities, 128K context, 140+ languages", Category: "LLM", Size: "~3GB", EstimatedTime: "3m"},
		{ID: "gemma3-12b", Name: "Gemma3 12B", Description: "Mid-size multimodal Gemma3 with vision, strong multilingual performance", Category: "LLM", Size: "~8GB", EstimatedTime: "7m"},
		{ID: "gemma3-27b", Name: "Gemma3 27B", Description: "Largest Gemma3 with vision capabilities, runs on single GPU", Category: "LLM", Size: "~17GB", EstimatedTime: "12m"},
		{ID: "llama-3.2-1b", Name: "Llama 3.2 1B", Description: "Ultra-compact model for edge devices, personal assistants, low-resource environments", Category: "LLM", Size: "~1GB", EstimatedTime: "1m"},
		{ID: "llama-3.2-3b", Name: "Llama 3.2 3B", Description: "Small efficient model for summarization, instructions, tool use, 128K context", Category: "LLM", Size: "~2GB", EstimatedTime: "2m"},
		{ID: "qwen3-coder-30b", Name: "Qwen3-Coder 30B MoE", Description: "Most agentic code model in Qwen series (30B total, 3.3B activated), 256K context", Category: "LLM", Size: "~19GB", EstimatedTime: "10m"},
		{ID: "command-r-7b", Name: "Command-R 7B", Description: "Cohere's efficient model optimized for RAG, multilingual, long context", Category: "LLM", Size: "~4GB", EstimatedTime: "4m"},

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
		{ID: "flux2", Name: "Flux 2 (FP8)", Description: "Next-generation video model from Black Forest Labs with superior motion and coherence (FP8 quantized)", Category: "Video Generation", Size: "~8GB", EstimatedTime: "15m"},
		{ID: "cogvideox-1.5", Name: "CogVideoX 1.5 5B", Description: "Upgraded CogVideoX supporting 10-second videos at higher resolutions", Category: "Video Generation", Size: "~18GB", EstimatedTime: "30m"},
		{ID: "cogvideox-i2v", Name: "CogVideoX 1.5 I2V", Description: "Image-to-video variant of CogVideoX 1.5 with any resolution support", Category: "Video Generation", Size: "~18GB", EstimatedTime: "30m"},
		{ID: "hunyuan-video", Name: "HunyuanVideo", Description: "Tencent's open-source text-to-video diffusion transformer model", Category: "Video Generation", Size: "~20GB", EstimatedTime: "35m"},
		{ID: "pyramid-flow", Name: "Pyramid Flow", Description: "Efficient video generation using pyramidal flow matching (768p, up to 10s)", Category: "Video Generation", Size: "~12GB", EstimatedTime: "20m"},
		{ID: "svd-xt", Name: "SVD-XT 1.1", Description: "Extended Stable Video Diffusion with improved temporal consistency", Category: "Video Generation", Size: "~10GB", EstimatedTime: "20m"},
		{ID: "i2v-adapter", Name: "I2V-Adapter", Description: "General image-to-video adapter for diffusion models (SIGGRAPH 2024)", Category: "Video Generation", Size: "~4GB", EstimatedTime: "10m"},

		// Additional Image Generation Models
		{ID: "sd3.5-large", Name: "Stable Diffusion 3.5 Large", Description: "8B parameter flagship SD3 model with exceptional quality at 1MP resolution", Category: "Image Generation", Size: "~16GB", EstimatedTime: "25m"},
		{ID: "sd3.5-large-turbo", Name: "SD 3.5 Large Turbo", Description: "Distilled SD3.5 Large generating high-quality images in 4 steps", Category: "Image Generation", Size: "~16GB", EstimatedTime: "25m"},
		{ID: "sd3.5-medium", Name: "Stable Diffusion 3.5 Medium", Description: "2.6B parameter SD3 with MMDiT-X architecture, consumer GPU friendly", Category: "Image Generation", Size: "~7GB", EstimatedTime: "15m"},
		{ID: "sdxl-turbo", Name: "SDXL Turbo", Description: "Real-time SDXL generating photorealistic images in a single step", Category: "Image Generation", Size: "~7GB", EstimatedTime: "10m"},
		{ID: "sdxl-lightning", Name: "SDXL Lightning", Description: "ByteDance's lightning-fast SDXL generating 1024px images in few steps", Category: "Image Generation", Size: "~7GB", EstimatedTime: "10m"},
		{ID: "playground-v2.5", Name: "Playground v2.5", Description: "State-of-the-art aesthetic model outperforming SDXL and DALL-E 3", Category: "Image Generation", Size: "~7GB", EstimatedTime: "15m"},
		{ID: "pixart-sigma", Name: "PixArt-Σ", Description: "Efficient DiT-based text-to-image with 4K support and improved text rendering", Category: "Image Generation", Size: "~8GB", EstimatedTime: "15m"},
		{ID: "kandinsky-3", Name: "Kandinsky 3", Description: "Sber AI's multilingual text-to-image model with strong Russian support", Category: "Image Generation", Size: "~8GB", EstimatedTime: "15m"},
		{ID: "kolors", Name: "Kolors", Description: "KWAI's bilingual Chinese-English text-to-image model", Category: "Image Generation", Size: "~8GB", EstimatedTime: "15m"},
		{ID: "sd-inpainting", Name: "SD 1.5 Inpainting", Description: "Stable Diffusion 1.5 fine-tuned for image inpainting and outpainting", Category: "Image Generation", Size: "~4GB", EstimatedTime: "10m"},
		{ID: "sdxl-inpainting", Name: "SDXL Inpainting", Description: "SDXL fine-tuned for high-resolution inpainting and outpainting", Category: "Image Generation", Size: "~7GB", EstimatedTime: "15m"},

		// Image/Video Enhancement
		{ID: "real-esrgan", Name: "Real-ESRGAN", Description: "Practical 4x image/video upscaling with artifact removal", Category: "Enhancement", Size: "~200MB", EstimatedTime: "5m"},
		{ID: "gfpgan", Name: "GFPGAN", Description: "Practical face restoration algorithm for real-world images", Category: "Enhancement", Size: "~350MB", EstimatedTime: "5m"},
		{ID: "aurasr", Name: "AuraSR", Description: "GigaGAN-based open-source 4x image upscaler from Fal.ai", Category: "Enhancement", Size: "~500MB", EstimatedTime: "5m"},
		{ID: "supir", Name: "SUPIR", Description: "Photo-realistic image restoration using SDXL with text-guided enhancement", Category: "Enhancement", Size: "~12GB", EstimatedTime: "20m"},
		{ID: "rife", Name: "RIFE", Description: "Real-time intermediate flow estimation for video frame interpolation", Category: "Enhancement", Size: "~200MB", EstimatedTime: "5m"},
		{ID: "film", Name: "FILM", Description: "Google's frame interpolation model for large motion between frames", Category: "Enhancement", Size: "~400MB", EstimatedTime: "5m"},

		// ControlNet & Adapters
		{ID: "controlnet-canny", Name: "ControlNet Canny", Description: "Edge detection based image conditioning for Stable Diffusion", Category: "ControlNet", Size: "~1.5GB", EstimatedTime: "5m"},
		{ID: "controlnet-depth", Name: "ControlNet Depth", Description: "Depth map conditioning for 3D-aware image generation", Category: "ControlNet", Size: "~1.5GB", EstimatedTime: "5m"},
		{ID: "controlnet-openpose", Name: "ControlNet OpenPose", Description: "Human pose estimation conditioning for character generation", Category: "ControlNet", Size: "~1.5GB", EstimatedTime: "5m"},
		{ID: "ip-adapter", Name: "IP-Adapter", Description: "Tencent's image prompt adapter for style and content transfer", Category: "ControlNet", Size: "~100MB", EstimatedTime: "5m"},
		{ID: "ip-adapter-faceid", Name: "IP-Adapter FaceID", Description: "Face-specific IP-Adapter for identity-preserving generation", Category: "ControlNet", Size: "~200MB", EstimatedTime: "5m"},
		{ID: "instantid", Name: "InstantID", Description: "Zero-shot identity-preserving generation with single reference image", Category: "ControlNet", Size: "~2GB", EstimatedTime: "10m"},
	}

	// Group by category
	categories := make(map[string][]Package)
	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort category names
	categoryOrder := []string{"LLM", "Image Generation", "Video Generation", "Enhancement", "ControlNet"}

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
		"Enhancement":      "🔧",
		"ControlNet":       "🎛️",
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
