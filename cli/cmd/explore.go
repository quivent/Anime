package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	exploreDetailedFlag bool
	exploreJSONFlag     bool
)

var exploreCmd = &cobra.Command{
	Use:   "explore [server-name]",
	Short: "Discover models and packages not tracked in anime",
	Long: `Scan the remote server for installed models and packages that aren't
managed by anime. This includes Ollama models, ComfyUI checkpoints, LoRAs,
and other AI models installed manually.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime explore <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime explore lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime explore my-server --detailed"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 What it does:"))
			fmt.Println(theme.DimTextStyle.Render("  Scans server for Ollama models, ComfyUI checkpoints, LoRAs, etc."))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  anime list     # List configured servers"))
			fmt.Println(theme.DimTextStyle.Render("  anime packages # List installed packages"))
			fmt.Println()
			return fmt.Errorf("explore requires a server name")
		}
		return nil
	},
	RunE: runExplore,
}

func init() {
	exploreCmd.Flags().BoolVarP(&exploreDetailedFlag, "detailed", "d", false, "Show detailed information including file paths")
	exploreCmd.Flags().BoolVar(&exploreJSONFlag, "json", false, "Output in JSON format")
	rootCmd.AddCommand(exploreCmd)
}

type ModelInfo struct {
	Name     string
	Size     string
	Path     string
	Category string
	Modified string
}

var (
	exploreCategoryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF69B4")).
		Bold(true).
		MarginTop(1)

	exploreModelNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	exploreSizeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))

	explorePathStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	exploreSummaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true).
		MarginTop(1)
)

func runExplore(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	serverName := args[0]
	server, err := cfg.GetServer(serverName)
	if err != nil {
		return fmt.Errorf("server %s not found", serverName)
	}

	fmt.Printf("🔍 Exploring %s for untracked models...\n\n", server.Name)

	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Collect all models
	allModels := make(map[string][]ModelInfo)

	// 1. Check Ollama models
	ollamaModels, err := discoverOllamaModels(client)
	if err == nil && len(ollamaModels) > 0 {
		allModels["Ollama Models"] = ollamaModels
	}

	// 2. Check ComfyUI checkpoints
	comfyCheckpoints, err := discoverComfyUIModels(client, "checkpoints", "~/ComfyUI/models/checkpoints")
	if err == nil && len(comfyCheckpoints) > 0 {
		allModels["ComfyUI Checkpoints"] = comfyCheckpoints
	}

	// 3. Check ComfyUI LoRAs
	comfyLoras, err := discoverComfyUIModels(client, "loras", "~/ComfyUI/models/loras")
	if err == nil && len(comfyLoras) > 0 {
		allModels["ComfyUI LoRAs"] = comfyLoras
	}

	// 4. Check ComfyUI VAEs
	comfyVAEs, err := discoverComfyUIModels(client, "vae", "~/ComfyUI/models/vae")
	if err == nil && len(comfyVAEs) > 0 {
		allModels["ComfyUI VAEs"] = comfyVAEs
	}

	// 5. Check ComfyUI CLIP models
	comfyCLIP, err := discoverComfyUIModels(client, "clip", "~/ComfyUI/models/clip")
	if err == nil && len(comfyCLIP) > 0 {
		allModels["ComfyUI CLIP"] = comfyCLIP
	}

	// 6. Check ComfyUI Upscalers
	comfyUpscale, err := discoverComfyUIModels(client, "upscale_models", "~/ComfyUI/models/upscale_models")
	if err == nil && len(comfyUpscale) > 0 {
		allModels["ComfyUI Upscalers"] = comfyUpscale
	}

	// 7. Check HuggingFace cache
	hfModels, err := discoverHuggingFaceModels(client)
	if err == nil && len(hfModels) > 0 {
		allModels["HuggingFace Cache"] = hfModels
	}

	// 8. Check standalone model directories
	standaloneModels, err := discoverStandaloneModels(client)
	if err == nil && len(standaloneModels) > 0 {
		allModels["Standalone Models"] = standaloneModels
	}

	// Display results
	if len(allModels) == 0 {
		fmt.Println("No untracked models found.")
		return nil
	}

	// Calculate totals
	totalModels := 0

	for _, models := range allModels {
		totalModels += len(models)
	}

	// Print by category
	categories := make([]string, 0, len(allModels))
	for category := range allModels {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	for _, category := range categories {
		models := allModels[category]
		emoji := getCategoryEmoji(category)

		fmt.Println(exploreCategoryStyle.Render(fmt.Sprintf("%s %s (%d)", emoji, category, len(models))))

		for _, model := range models {
			if exploreDetailedFlag {
				fmt.Printf("  %s\n", exploreModelNameStyle.Render(model.Name))
				fmt.Printf("    Size: %s\n", exploreSizeStyle.Render(model.Size))
				if model.Modified != "" {
					fmt.Printf("    Modified: %s\n", explorePathStyle.Render(model.Modified))
				}
				fmt.Printf("    Path: %s\n", explorePathStyle.Render(model.Path))
			} else {
				fmt.Printf("  %s %s\n",
					exploreModelNameStyle.Render(model.Name),
					exploreSizeStyle.Render(fmt.Sprintf("(%s)", model.Size)))
			}
		}
		fmt.Println()
	}

	// Print summary
	fmt.Println(exploreSummaryStyle.Render(fmt.Sprintf("📊 Total: %d models found across %d categories", totalModels, len(allModels))))
	fmt.Println()

	// Next steps
	fmt.Println(exploreSummaryStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(exploreCategoryStyle.Render("💡 What to do next:"))
	fmt.Println(exploreSummaryStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", exploreModelNameStyle.Render("anime packages status"))
	fmt.Println(explorePathStyle.Render("    Check installation status of all packages"))
	fmt.Println()
	fmt.Printf("  %s\n", exploreModelNameStyle.Render("anime status "+serverName))
	fmt.Println(explorePathStyle.Render("    View detailed server information"))
	fmt.Println()
	fmt.Printf("  %s\n", exploreModelNameStyle.Render("anime workstation"))
	fmt.Println(explorePathStyle.Render("    Launch interactive monitoring dashboard"))
	fmt.Println()
	fmt.Printf("  %s\n", exploreModelNameStyle.Render("anime collection push <name>"))
	fmt.Println(explorePathStyle.Render("    Push local assets to this server"))
	fmt.Println()

	return nil
}

func discoverOllamaModels(client *ssh.Client) ([]ModelInfo, error) {
	output, err := client.RunCommand("ollama list 2>/dev/null")
	if err != nil || output == "" {
		return nil, fmt.Errorf("ollama not found or no models")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= 1 {
		return nil, fmt.Errorf("no models")
	}

	models := make([]ModelInfo, 0)
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			models = append(models, ModelInfo{
				Name:     fields[0],
				Size:     fields[len(fields)-2],
				Modified: fields[len(fields)-1],
				Category: "ollama",
			})
		}
	}

	return models, nil
}

func discoverComfyUIModels(client *ssh.Client, category, path string) ([]ModelInfo, error) {
	// Check if directory exists
	checkCmd := fmt.Sprintf("[ -d %s ] && echo exists", path)
	output, err := client.RunCommand(checkCmd)
	if err != nil || !strings.Contains(output, "exists") {
		return nil, fmt.Errorf("directory not found")
	}

	// List model files (safetensors, ckpt, pt, pth, bin)
	listCmd := fmt.Sprintf(`find %s -maxdepth 2 -type f \( -name "*.safetensors" -o -name "*.ckpt" -o -name "*.pt" -o -name "*.pth" -o -name "*.bin" \) -exec sh -c 'echo "{}|$(du -h "{}" | cut -f1)|$(stat -c %%y "{}" 2>/dev/null || stat -f %%Sm "{}")"' \; 2>/dev/null`, path)

	output, err = client.RunCommand(listCmd)
	if err != nil || output == "" {
		return nil, fmt.Errorf("no models found")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	models := make([]ModelInfo, 0)

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			fullPath := parts[0]
			size := strings.TrimSpace(parts[1])
			modified := ""
			if len(parts) >= 3 {
				modified = strings.TrimSpace(parts[2])
			}

			// Extract just the filename
			pathParts := strings.Split(fullPath, "/")
			name := pathParts[len(pathParts)-1]

			models = append(models, ModelInfo{
				Name:     name,
				Size:     size,
				Path:     fullPath,
				Modified: modified,
				Category: category,
			})
		}
	}

	return models, nil
}

func discoverHuggingFaceModels(client *ssh.Client) ([]ModelInfo, error) {
	// Check HuggingFace cache directory
	hfCache := "~/.cache/huggingface/hub"
	checkCmd := fmt.Sprintf("[ -d %s ] && echo exists", hfCache)
	output, err := client.RunCommand(checkCmd)
	if err != nil || !strings.Contains(output, "exists") {
		return nil, fmt.Errorf("hf cache not found")
	}

	// List model directories
	listCmd := fmt.Sprintf(`find %s -maxdepth 1 -type d -name "models--*" -exec sh -c 'echo "{}|$(du -sh "{}" | cut -f1)"' \; 2>/dev/null`, hfCache)
	output, err = client.RunCommand(listCmd)
	if err != nil || output == "" {
		return nil, fmt.Errorf("no models found")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	models := make([]ModelInfo, 0)

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			fullPath := parts[0]
			size := strings.TrimSpace(parts[1])

			// Extract model name from path (models--org--name -> org/name)
			pathParts := strings.Split(fullPath, "/")
			dirName := pathParts[len(pathParts)-1]
			if strings.HasPrefix(dirName, "models--") {
				modelName := strings.Replace(dirName, "models--", "", 1)
				modelName = strings.Replace(modelName, "--", "/", 1)

				models = append(models, ModelInfo{
					Name:     modelName,
					Size:     size,
					Path:     fullPath,
					Category: "huggingface",
				})
			}
		}
	}

	return models, nil
}

func discoverStandaloneModels(client *ssh.Client) ([]ModelInfo, error) {
	// Check common model directories
	modelDirs := []string{
		"~/models",
		"~/Downloads",
		"~/Mochi-1-preview",
		"~/cogvideox",
		"~/Open-Sora",
		"~/ltx-video",
	}

	models := make([]ModelInfo, 0)

	for _, dir := range modelDirs {
		// Check if directory exists
		checkCmd := fmt.Sprintf("[ -d %s ] && echo exists", dir)
		output, err := client.RunCommand(checkCmd)
		if err != nil || !strings.Contains(output, "exists") {
			continue
		}

		// Find model files
		listCmd := fmt.Sprintf(`find %s -maxdepth 2 -type f \( -name "*.safetensors" -o -name "*.ckpt" -o -name "*.pt" -o -name "*.pth" -o -name "*.bin" \) -exec sh -c 'echo "{}|$(du -h "{}" | cut -f1)"' \; 2>/dev/null`, dir)
		output, err = client.RunCommand(listCmd)
		if err != nil || output == "" {
			continue
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				fullPath := parts[0]
				size := strings.TrimSpace(parts[1])

				// Extract filename
				pathParts := strings.Split(fullPath, "/")
				name := pathParts[len(pathParts)-1]

				models = append(models, ModelInfo{
					Name:     name,
					Size:     size,
					Path:     fullPath,
					Category: "standalone",
				})
			}
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}

	return models, nil
}

func getCategoryEmoji(category string) string {
	switch category {
	case "Ollama Models":
		return "🦙"
	case "ComfyUI Checkpoints":
		return "🎨"
	case "ComfyUI LoRAs":
		return "🎭"
	case "ComfyUI VAEs":
		return "🔧"
	case "ComfyUI CLIP":
		return "📎"
	case "ComfyUI Upscalers":
		return "🔍"
	case "HuggingFace Cache":
		return "🤗"
	case "Standalone Models":
		return "📦"
	default:
		return "📁"
	}
}
