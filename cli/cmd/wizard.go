package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var wizardCmd = &cobra.Command{
	Use:     "wizard",
	Short:   "Interactive setup wizard for configuring your node",
	Aliases: []string{"w", "setup", "configure"},
	Run:     runWizard,
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}

type wizardConfig struct {
	nodeType            string // inference, training, art, development
	installCore         bool
	installPython       bool
	installPyTorch      bool
	installOllama       bool
	installVLLM         bool
	llmRuntime          string // "ollama" or "vllm"
	ollamaService       bool
	installClaude       bool
	llmModels           []string // small, medium, large
	installComfyUI      bool
	videoModels         []string // mochi, svd, animatediff, cogvideo, opensora, ltxvideo
	clusterMode         bool
	selectedPackages    []string
}

func runWizard(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	config := &wizardConfig{}

	// Welcome banner
	fmt.Println(theme.RenderBanner("✨ ANIME SETUP WIZARD ✨"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🌸 Welcome to the ANIME configuration wizard!"))
	fmt.Println(theme.DimTextStyle.Render("   I'll help you set up your Lambda GH200 node perfectly."))
	fmt.Println()

	// Step 1: Node Purpose
	printStep(1, "What's your primary use case?")
	fmt.Println()
	fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("🤖 Inference") + theme.DimTextStyle.Render(" - Running LLMs and models for production"))
	fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("🎓 Training") + theme.DimTextStyle.Render(" - Fine-tuning and training models"))
	fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("🎨 Art & Video") + theme.DimTextStyle.Render(" - Creating images, videos, and animations"))
	fmt.Println(theme.HighlightStyle.Render("  4") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("💻 Development") + theme.DimTextStyle.Render(" - Coding, testing, and experimentation"))
	fmt.Println(theme.HighlightStyle.Render("  5") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("🚀 Everything") + theme.DimTextStyle.Render(" - Full production ML/AI workstation"))
	fmt.Println()

	choice := promptChoice(reader, "Choose your path", []string{"1", "2", "3", "4", "5"})
	switch choice {
	case "1":
		config.nodeType = "inference"
		setupInferenceNode(config)
	case "2":
		config.nodeType = "training"
		setupTrainingNode(config)
	case "3":
		config.nodeType = "art"
		setupArtNode(config)
	case "4":
		config.nodeType = "development"
		setupDevelopmentNode(config)
	case "5":
		config.nodeType = "everything"
		setupEverythingNode(config)
	}

	fmt.Println()

	// Step 2: LLM Runtime Selection
	if config.installOllama || config.installVLLM {
		printStep(2, "Choose your LLM inference runtime")
		fmt.Println()
		fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("🔮 Ollama") + theme.DimTextStyle.Render(" - Easy-to-use LLM server with built-in model management"))
		fmt.Println(theme.DimTextStyle.Render("       ✓ Simple setup, automatic model downloads"))
		fmt.Println(theme.DimTextStyle.Render("       ✓ Great for development and quick deployment"))
		fmt.Println()
		fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("⚡ vLLM") + theme.DimTextStyle.Render(" - High-performance inference engine with PagedAttention"))
		fmt.Println(theme.DimTextStyle.Render("       ✓ Optimized for throughput and batch processing"))
		fmt.Println(theme.DimTextStyle.Render("       ✓ Best for production and high-scale deployments"))
		fmt.Println()

		if config.nodeType == "inference" || config.nodeType == "everything" {
			fmt.Println(theme.SuccessStyle.Render("  💡 Recommended for " + config.nodeType + ": vLLM (high performance)"))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  💡 Recommended for " + config.nodeType + ": Ollama (easy setup)"))
		}
		fmt.Println()

		choice := promptChoice(reader, "Select LLM runtime", []string{"1", "2"})
		if choice == "1" {
			config.llmRuntime = "ollama"
			config.installOllama = true
			config.installVLLM = false
		} else {
			config.llmRuntime = "vllm"
			config.installVLLM = true
			config.installOllama = false
		}
		fmt.Println()
	}

	// Step 3: LLM Models
	if config.installOllama {
		printStep(3, "Which LLM models do you want?")
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  You can select multiple by entering numbers separated by spaces (e.g., 1 2)"))
		fmt.Println()
		fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("⭐ Small Models (7-8B)") + theme.DimTextStyle.Render(" - ~15GB, fast inference"))
		fmt.Println(theme.DimTextStyle.Render("       Mistral, Llama 3.3 8B, Qwen3 8B"))
		fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("⭐ Medium Models (14-34B)") + theme.DimTextStyle.Render(" - ~45GB, balanced"))
		fmt.Println(theme.DimTextStyle.Render("       Qwen3 14B, Qwen3 32B, Mixtral 8x7B, DeepSeek Coder 33B"))
		fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("⭐ Large Models (70B+)") + theme.DimTextStyle.Render(" - ~250GB, best quality"))
		fmt.Println(theme.DimTextStyle.Render("       Llama 3.3 70B, Qwen3 235B MoE, DeepSeek V3"))
		fmt.Println()

		recommended := getRecommendedLLMModels(config.nodeType)
		fmt.Println(theme.SuccessStyle.Render("  💡 Recommended: " + strings.Join(recommended, ", ")))
		fmt.Println()

		selections := promptMultiChoice(reader, "Select models (or press Enter for recommended)")
		if len(selections) == 0 {
			config.llmModels = recommended
		} else {
			for _, sel := range selections {
				switch sel {
				case "1":
					config.llmModels = append(config.llmModels, "models-small")
				case "2":
					config.llmModels = append(config.llmModels, "models-medium")
				case "3":
					config.llmModels = append(config.llmModels, "models-large")
				}
			}
		}
		fmt.Println()
	}

	// Step 4: Video/Image Models
	if config.installComfyUI || config.installPyTorch {
		printStep(4, "Which video/image generation models?")
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Select multiple by entering numbers separated by spaces"))
		fmt.Println()

		if config.installComfyUI {
			fmt.Println(theme.CategoryStyle("🎬 ComfyUI Models"))
			fmt.Println(theme.HighlightStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Stable Video Diffusion") + theme.DimTextStyle.Render(" - ~8GB, img2vid"))
			fmt.Println(theme.HighlightStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("AnimateDiff") + theme.DimTextStyle.Render(" - ~4GB, animate images"))
			fmt.Println()
		}

		if config.installPyTorch {
			fmt.Println(theme.CategoryStyle("🎬 Standalone Models"))
			fmt.Println(theme.HighlightStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Mochi-1") + theme.DimTextStyle.Render(" - ~12GB, 10B params video gen"))
			fmt.Println(theme.HighlightStyle.Render("  4") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("CogVideoX-5B") + theme.DimTextStyle.Render(" - ~14GB, text-to-video"))
			fmt.Println(theme.HighlightStyle.Render("  5") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("Open-Sora 2.0") + theme.DimTextStyle.Render(" - ~16GB, high quality"))
			fmt.Println(theme.HighlightStyle.Render("  6") + theme.DimTextStyle.Render(" → ") + theme.InfoStyle.Render("LTXVideo") + theme.DimTextStyle.Render(" - ~7GB, fast generation"))
			fmt.Println()
		}

		recommended := getRecommendedVideoModels(config.nodeType, config.installComfyUI)
		if len(recommended) > 0 {
			fmt.Println(theme.SuccessStyle.Render("  💡 Recommended: " + strings.Join(recommended, ", ")))
			fmt.Println()
		}

		selections := promptMultiChoice(reader, "Select models (or press Enter for recommended)")
		if len(selections) == 0 {
			config.videoModels = getRecommendedVideoModelIDs(config.nodeType, config.installComfyUI)
		} else {
			for _, sel := range selections {
				switch sel {
				case "1":
					if config.installComfyUI {
						config.videoModels = append(config.videoModels, "svd")
					}
				case "2":
					if config.installComfyUI {
						config.videoModels = append(config.videoModels, "animatediff")
					}
				case "3":
					if config.installPyTorch {
						config.videoModels = append(config.videoModels, "mochi")
					}
				case "4":
					if config.installPyTorch {
						config.videoModels = append(config.videoModels, "cogvideo")
					}
				case "5":
					if config.installPyTorch {
						config.videoModels = append(config.videoModels, "opensora")
					}
				case "6":
					if config.installPyTorch {
						config.videoModels = append(config.videoModels, "ltxvideo")
					}
				}
			}
		}
		fmt.Println()
	}

	// Step 5: Ollama Service Configuration
	if config.installOllama {
		printStep(5, "Ollama Service Configuration")
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Should Ollama run as a persistent systemd service?"))
		fmt.Println(theme.DimTextStyle.Render("  This allows it to start automatically and run continuously."))
		fmt.Println()

		if promptYesNo(reader, "Enable Ollama systemd service", config.nodeType == "inference" || config.nodeType == "everything") {
			config.ollamaService = true
			fmt.Println(theme.SuccessStyle.Render("  ✓ Ollama will run on 0.0.0.0:11434 (accessible from network)"))
		} else {
			fmt.Println(theme.InfoStyle.Render("  → Ollama will be available via 'ollama serve' command"))
		}
		fmt.Println()
	}

	// Step 6: Cluster Configuration
	printStep(6, "Cluster & Parallelization")
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Are you setting up multiple nodes for distributed workloads?"))
	fmt.Println()

	if promptYesNo(reader, "Configure for cluster mode", false) {
		config.clusterMode = true
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("  ✓ Cluster mode enabled"))
		fmt.Println(theme.DimTextStyle.Render("    → Ollama will bind to 0.0.0.0 for network access"))
		fmt.Println(theme.DimTextStyle.Render("    → Consider using `anime server add` to register other nodes"))
		fmt.Println(theme.DimTextStyle.Render("    → Use `anime install --remote -s <server>` for multi-node deployment"))
	} else {
		fmt.Println(theme.InfoStyle.Render("  → Single node configuration"))
	}
	fmt.Println()

	// Step 7: Development Tools
	if config.nodeType == "development" || config.nodeType == "everything" {
		printStep(7, "Development Tools")
		fmt.Println()

		if promptYesNo(reader, "Install Claude Code CLI", true) {
			config.installClaude = true
		}
		fmt.Println()
	}

	// Build package list
	config.selectedPackages = buildPackageList(config)

	// Show summary
	fmt.Println()
	fmt.Println(theme.RenderBanner("📋 INSTALLATION SUMMARY"))
	fmt.Println()

	resolved, err := installer.ResolveDependencies(config.selectedPackages)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error: " + err.Error()))
		os.Exit(1)
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🎯 Node Type: %s", strings.Title(config.nodeType))))
	if config.llmRuntime != "" {
		runtimeName := strings.Title(config.llmRuntime)
		runtimeEmoji := "🔮"
		if config.llmRuntime == "vllm" {
			runtimeEmoji = "⚡"
		}
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("%s LLM Runtime: %s", runtimeEmoji, runtimeName)))
	}
	fmt.Println()

	var totalTime int

	fmt.Println(theme.CategoryStyle("📦 Packages to Install:"))
	for i, pkg := range resolved {
		marker := theme.SymbolBranch
		if i == len(resolved)-1 {
			marker = theme.SymbolLastBranch
		}

		fmt.Printf("%s %s %s\n",
			theme.DimTextStyle.Render(marker),
			theme.HighlightStyle.Render(pkg.Name),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s, %s)", pkg.EstimatedTime, pkg.Size)))

		totalTime += int(pkg.EstimatedTime.Minutes())
	}

	fmt.Println()
	timeStr := fmt.Sprintf("%dh %dm", totalTime/60, totalTime%60)
	if totalTime < 60 {
		timeStr = fmt.Sprintf("%dm", totalTime)
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("📊 Total: %d packages  |  ⏱️  ~%s", len(resolved), timeStr)))
	fmt.Println()

	if config.ollamaService {
		fmt.Println(theme.InfoStyle.Render("⚙️  Ollama systemd service will be enabled"))
	}
	if config.clusterMode {
		fmt.Println(theme.InfoStyle.Render("🌐 Cluster mode configuration applied"))
	}
	fmt.Println()

	// Confirm and install
	if !promptYesNo(reader, "Proceed with installation", true) {
		fmt.Println(theme.WarningStyle.Render("Installation cancelled"))
		return
	}

	// Execute installation
	fmt.Println()
	runLocalInstall(resolved)
}

func setupInferenceNode(config *wizardConfig) {
	config.installCore = true
	config.installPython = true
	config.installPyTorch = true
	config.installOllama = true  // Will be chosen in step 2
	config.installVLLM = true     // Will be chosen in step 2
	config.ollamaService = true
	config.installClaude = false
	config.installComfyUI = false
}

func setupTrainingNode(config *wizardConfig) {
	config.installCore = true
	config.installPython = true
	config.installPyTorch = true
	config.installOllama = false
	config.installVLLM = false
	config.installClaude = true
	config.installComfyUI = false
}

func setupArtNode(config *wizardConfig) {
	config.installCore = true
	config.installPython = true
	config.installPyTorch = true
	config.installOllama = true  // Will be chosen in step 2
	config.installVLLM = true     // Will be chosen in step 2
	config.ollamaService = false
	config.installClaude = false
	config.installComfyUI = true
}

func setupDevelopmentNode(config *wizardConfig) {
	config.installCore = true
	config.installPython = true
	config.installPyTorch = false
	config.installOllama = true  // Will be chosen in step 2
	config.installVLLM = true     // Will be chosen in step 2
	config.ollamaService = false
	config.installClaude = true
	config.installComfyUI = false
}

func setupEverythingNode(config *wizardConfig) {
	config.installCore = true
	config.installPython = true
	config.installPyTorch = true
	config.installOllama = true  // Will be chosen in step 2
	config.installVLLM = true     // Will be chosen in step 2
	config.ollamaService = true
	config.installClaude = true
	config.installComfyUI = true
}

func getRecommendedLLMModels(nodeType string) []string {
	switch nodeType {
	case "inference":
		return []string{"Small", "Medium"}
	case "training":
		return []string{"Small"}
	case "art":
		return []string{"Small"}
	case "development":
		return []string{"Small"}
	case "everything":
		return []string{"Small", "Medium", "Large"}
	}
	return []string{"Small"}
}

func getRecommendedVideoModels(nodeType string, hasComfyUI bool) []string {
	switch nodeType {
	case "art":
		if hasComfyUI {
			return []string{"SVD", "AnimateDiff", "Mochi-1"}
		}
		return []string{"Mochi-1", "LTXVideo"}
	case "everything":
		if hasComfyUI {
			return []string{"SVD", "AnimateDiff", "Mochi-1", "CogVideoX"}
		}
		return []string{"Mochi-1", "CogVideoX"}
	}
	return []string{}
}

func getRecommendedVideoModelIDs(nodeType string, hasComfyUI bool) []string {
	switch nodeType {
	case "art":
		if hasComfyUI {
			return []string{"svd", "animatediff", "mochi"}
		}
		return []string{"mochi", "ltxvideo"}
	case "everything":
		if hasComfyUI {
			return []string{"svd", "animatediff", "mochi", "cogvideo"}
		}
		return []string{"mochi", "cogvideo"}
	}
	return []string{}
}

func buildPackageList(config *wizardConfig) []string {
	packages := []string{}

	if config.installCore {
		packages = append(packages, "core")
	}
	if config.installPython {
		packages = append(packages, "python")
	}
	if config.installPyTorch {
		packages = append(packages, "pytorch")
	}
	if config.installOllama {
		packages = append(packages, "ollama")
	}
	if config.installVLLM {
		packages = append(packages, "vllm")
	}
	if config.installComfyUI {
		packages = append(packages, "comfyui")
	}
	if config.installClaude {
		packages = append(packages, "claude")
	}

	packages = append(packages, config.llmModels...)
	packages = append(packages, config.videoModels...)

	return packages
}

func printStep(num int, title string) {
	fmt.Println(theme.CategoryStyle(fmt.Sprintf("╔══ Step %d: %s", num, title)))
}

func promptChoice(reader *bufio.Reader, prompt string, validChoices []string) string {
	for {
		fmt.Print(theme.HighlightStyle.Render(prompt+" ▶ "))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		for _, valid := range validChoices {
			if response == valid {
				return response
			}
		}

		fmt.Println(theme.ErrorStyle.Render("  ✗ Invalid choice. Please try again."))
	}
}

func promptMultiChoice(reader *bufio.Reader, prompt string) []string {
	fmt.Print(theme.HighlightStyle.Render(prompt+" ▶ "))
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" {
		return []string{}
	}

	return strings.Fields(response)
}

func promptYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	fmt.Print(theme.HighlightStyle.Render(fmt.Sprintf("%s (%s) ▶ ", prompt, defaultStr)))
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}
