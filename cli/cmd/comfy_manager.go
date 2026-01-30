package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// ComfyUI Manager - manages custom nodes, models, and extensions

var comfyManagerCmd = &cobra.Command{
	Use:     "manager",
	Aliases: []string{"mgr", "m"},
	Short:   "ComfyUI Manager - install nodes, models, and extensions",
	Long: `ComfyUI Manager provides comprehensive management of:
  - Custom Nodes: Install, remove, update, and list custom nodes
  - Models: Download and manage models (checkpoints, LoRAs, VAEs, etc.)
  - Workflows: Browse and install community workflows
  - Extensions: Manage ComfyUI extensions

Similar to ComfyUI-Manager but from the command line.`,
	RunE: runComfyManager,
}

var comfyNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage custom nodes",
	Long:  "Install, remove, update, and list ComfyUI custom nodes",
	RunE:  runComfyNodes,
}

var comfyModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage ComfyUI models",
	Long:  "Download and manage models for ComfyUI (checkpoints, LoRAs, VAEs, ControlNets, etc.)",
	RunE:  runComfyModels,
}

var comfyWorkflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Browse and install workflows",
	Long:  "Browse, download, and manage ComfyUI workflows",
	RunE:  runComfyWorkflows,
}

var comfyExtensionsCmd = &cobra.Command{
	Use:     "extensions",
	Aliases: []string{"ext"},
	Short:   "Manage extensions",
	Long:    "Manage ComfyUI extensions and plugins",
	RunE:    runComfyExtensions,
}

// Subcommands for nodes
var comfyNodesInstallCmd = &cobra.Command{
	Use:   "install <node-name>",
	Short: "Install a custom node",
	Args:  cobra.ExactArgs(1),
	RunE:  runComfyNodesInstall,
}

var comfyNodesRemoveCmd = &cobra.Command{
	Use:     "remove <node-name>",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove a custom node",
	Args:    cobra.ExactArgs(1),
	RunE:    runComfyNodesRemove,
}

var comfyNodesUpdateCmd = &cobra.Command{
	Use:   "update [node-name]",
	Short: "Update custom nodes (all or specific)",
	RunE:  runComfyNodesUpdate,
}

var comfyNodesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed custom nodes",
	RunE:    runComfyNodesList,
}

var comfyNodesCatalogCmd = &cobra.Command{
	Use:     "catalog",
	Aliases: []string{"browse", "available"},
	Short:   "Browse available custom nodes",
	RunE:    runComfyNodesCatalog,
}

// Subcommands for models
var comfyModelsDownloadCmd = &cobra.Command{
	Use:     "download <model-id>",
	Aliases: []string{"dl", "get"},
	Short:   "Download a model",
	Args:    cobra.ExactArgs(1),
	RunE:    runComfyModelsDownload,
}

var comfyModelsListCmd = &cobra.Command{
	Use:     "list [type]",
	Aliases: []string{"ls"},
	Short:   "List installed models (optionally by type)",
	RunE:    runComfyModelsList,
}

var comfyModelsCatalogCmd = &cobra.Command{
	Use:     "catalog",
	Aliases: []string{"browse"},
	Short:   "Browse available models for download",
	RunE:    runComfyModelsCatalog,
}

func init() {
	comfyuiCmd.AddCommand(comfyManagerCmd)

	// Add subcommands to manager
	comfyManagerCmd.AddCommand(comfyNodesCmd)
	comfyManagerCmd.AddCommand(comfyModelsCmd)
	comfyManagerCmd.AddCommand(comfyWorkflowsCmd)
	comfyManagerCmd.AddCommand(comfyExtensionsCmd)

	// Add subcommands to nodes
	comfyNodesCmd.AddCommand(comfyNodesInstallCmd)
	comfyNodesCmd.AddCommand(comfyNodesRemoveCmd)
	comfyNodesCmd.AddCommand(comfyNodesUpdateCmd)
	comfyNodesCmd.AddCommand(comfyNodesListCmd)
	comfyNodesCmd.AddCommand(comfyNodesCatalogCmd)

	// Add subcommands to models
	comfyModelsCmd.AddCommand(comfyModelsDownloadCmd)
	comfyModelsCmd.AddCommand(comfyModelsListCmd)
	comfyModelsCmd.AddCommand(comfyModelsCatalogCmd)
}

// CustomNode represents a ComfyUI custom node
type CustomNode struct {
	Name        string
	Repo        string
	Description string
	Category    string
	Author      string
	Stars       string
}

// ComfyModel represents a model for ComfyUI
type ComfyModel struct {
	ID          string
	Name        string
	Type        string // checkpoint, lora, vae, controlnet, clip, unet, embedding, upscaler
	Source      string // HuggingFace repo or direct URL
	Size        string
	Description string
}

// Popular custom nodes catalog
var customNodesCatalog = []CustomNode{
	// Essential Nodes
	{Name: "ComfyUI-Manager", Repo: "ltdrdata/ComfyUI-Manager", Description: "GUI manager for custom nodes", Category: "Essential", Author: "ltdrdata", Stars: "4.5k"},
	{Name: "ComfyUI-Impact-Pack", Repo: "ltdrdata/ComfyUI-Impact-Pack", Description: "Essential nodes for face detection, SAM, upscaling", Category: "Essential", Author: "ltdrdata", Stars: "2.1k"},
	{Name: "ComfyUI-Inspire-Pack", Repo: "ltdrdata/ComfyUI-Inspire-Pack", Description: "Regional prompting, wildcards, and utilities", Category: "Essential", Author: "ltdrdata", Stars: "1.2k"},
	{Name: "rgthree-comfy", Repo: "rgthree/rgthree-comfy", Description: "Power lora loader, bookmark, context nodes", Category: "Essential", Author: "rgthree", Stars: "1.1k"},
	{Name: "was-node-suite-comfyui", Repo: "WASasquatch/was-node-suite-comfyui", Description: "Huge collection of utility nodes", Category: "Essential", Author: "WASasquatch", Stars: "1.5k"},

	// ControlNet & IP-Adapter
	{Name: "ComfyUI-Advanced-ControlNet", Repo: "Kosinkadink/ComfyUI-Advanced-ControlNet", Description: "Advanced ControlNet features and scheduling", Category: "ControlNet", Author: "Kosinkadink", Stars: "1.3k"},
	{Name: "ComfyUI_IPAdapter_plus", Repo: "cubiq/ComfyUI_IPAdapter_plus", Description: "IP-Adapter integration with advanced features", Category: "ControlNet", Author: "cubiq", Stars: "2.8k"},
	{Name: "ComfyUI_InstantID", Repo: "cubiq/ComfyUI_InstantID", Description: "InstantID integration for identity-preserving generation", Category: "ControlNet", Author: "cubiq", Stars: "1.1k"},
	{Name: "comfyui_controlnet_aux", Repo: "Fannovel16/comfyui_controlnet_aux", Description: "ControlNet preprocessors (OpenPose, Depth, Canny, etc.)", Category: "ControlNet", Author: "Fannovel16", Stars: "1.8k"},

	// Video Generation
	{Name: "ComfyUI-VideoHelperSuite", Repo: "Kosinkadink/ComfyUI-VideoHelperSuite", Description: "Video loading, combining, and export nodes", Category: "Video", Author: "Kosinkadink", Stars: "1.4k"},
	{Name: "ComfyUI-AnimateDiff-Evolved", Repo: "Kosinkadink/ComfyUI-AnimateDiff-Evolved", Description: "Advanced AnimateDiff integration", Category: "Video", Author: "Kosinkadink", Stars: "2.2k"},
	{Name: "ComfyUI-Frame-Interpolation", Repo: "Fannovel16/ComfyUI-Frame-Interpolation", Description: "RIFE and FILM frame interpolation", Category: "Video", Author: "Fannovel16", Stars: "800"},
	{Name: "ComfyUI-CogVideoX-Wrapper", Repo: "kijai/ComfyUI-CogVideoXWrapper", Description: "CogVideoX text-to-video wrapper", Category: "Video", Author: "kijai", Stars: "600"},
	{Name: "ComfyUI-HunyuanVideo-Wrapper", Repo: "kijai/ComfyUI-HunyuanVideoWrapper", Description: "HunyuanVideo integration", Category: "Video", Author: "kijai", Stars: "400"},

	// Image Enhancement
	{Name: "ComfyUI_UltimateSDUpscale", Repo: "ssitu/ComfyUI_UltimateSDUpscale", Description: "Tiled upscaling with Stable Diffusion", Category: "Upscale", Author: "ssitu", Stars: "900"},
	{Name: "ComfyUI-SUPIR", Repo: "kijai/ComfyUI-SUPIR", Description: "SUPIR photo-realistic restoration", Category: "Upscale", Author: "kijai", Stars: "500"},
	{Name: "ComfyUI_essentials", Repo: "cubiq/ComfyUI_essentials", Description: "Essential image processing nodes", Category: "Upscale", Author: "cubiq", Stars: "1.0k"},

	// Flux Support
	{Name: "ComfyUI-GGUF", Repo: "city96/ComfyUI-GGUF", Description: "GGUF quantized model support (Flux, SD3)", Category: "Flux", Author: "city96", Stars: "1.2k"},
	{Name: "ComfyUI-Florence2", Repo: "kijai/ComfyUI-Florence2", Description: "Florence-2 vision model integration", Category: "Flux", Author: "kijai", Stars: "400"},
	{Name: "x-flux-comfyui", Repo: "XLabs-AI/x-flux-comfyui", Description: "X-Labs Flux ControlNet and IP-Adapter", Category: "Flux", Author: "XLabs-AI", Stars: "800"},

	// LLM & Prompting
	{Name: "ComfyUI-Custom-Scripts", Repo: "pythongosssss/ComfyUI-Custom-Scripts", Description: "Autocomplete, image feed, and workflow tools", Category: "Utility", Author: "pythongosssss", Stars: "1.6k"},
	{Name: "ComfyUI-KJNodes", Repo: "kijai/ComfyUI-KJNodes", Description: "Utility nodes for various tasks", Category: "Utility", Author: "kijai", Stars: "700"},
	{Name: "ComfyUI-Easy-Use", Repo: "yolain/ComfyUI-Easy-Use", Description: "Simplified workflow nodes", Category: "Utility", Author: "yolain", Stars: "900"},

	// Face & Portrait
	{Name: "ComfyUI-FaceAnalysis", Repo: "cubiq/ComfyUI_FaceAnalysis", Description: "Face detection and analysis nodes", Category: "Face", Author: "cubiq", Stars: "400"},
	{Name: "ComfyUI-ReActor", Repo: "Gourieff/ComfyUI-ReActor", Description: "Face swap node based on ReActor", Category: "Face", Author: "Gourieff", Stars: "1.0k"},
	{Name: "ComfyUI-BiRefNet", Repo: "viperyl/ComfyUI-BiRefNet", Description: "High-quality background removal", Category: "Face", Author: "viperyl", Stars: "300"},

	// Advanced Sampling
	{Name: "ComfyUI_smZNodes", Repo: "shiimizu/ComfyUI_smZNodes", Description: "A1111-style prompts and CLIP text encoding", Category: "Sampling", Author: "shiimizu", Stars: "400"},
	{Name: "ComfyUI-sampler-lcm-alternative", Repo: "jojkaart/ComfyUI-sampler-lcm-alternative", Description: "Alternative LCM sampler", Category: "Sampling", Author: "jojkaart", Stars: "200"},
}

// Models catalog organized by type
var comfyModelsCatalogData = map[string][]ComfyModel{
	"checkpoint": {
		{ID: "sd15", Name: "SD 1.5", Type: "checkpoint", Source: "runwayml/stable-diffusion-v1-5", Size: "~4GB", Description: "Stable Diffusion 1.5 base model"},
		{ID: "sdxl", Name: "SDXL 1.0", Type: "checkpoint", Source: "stabilityai/stable-diffusion-xl-base-1.0", Size: "~7GB", Description: "Stable Diffusion XL base"},
		{ID: "sd3-medium", Name: "SD 3 Medium", Type: "checkpoint", Source: "stabilityai/stable-diffusion-3-medium", Size: "~4GB", Description: "SD3 Medium with MMDiT"},
		{ID: "sd35-large", Name: "SD 3.5 Large", Type: "checkpoint", Source: "stabilityai/stable-diffusion-3.5-large", Size: "~16GB", Description: "SD3.5 8B parameter flagship"},
		{ID: "flux-dev", Name: "Flux.1 Dev", Type: "checkpoint", Source: "black-forest-labs/FLUX.1-dev", Size: "~24GB", Description: "Flux development model"},
		{ID: "flux-schnell", Name: "Flux.1 Schnell", Type: "checkpoint", Source: "black-forest-labs/FLUX.1-schnell", Size: "~24GB", Description: "Fast Flux model"},
		{ID: "flux-dev-fp8", Name: "Flux.1 Dev FP8", Type: "checkpoint", Source: "Kijai/flux-fp8", Size: "~12GB", Description: "FP8 quantized Flux Dev"},
		{ID: "playground-v25", Name: "Playground v2.5", Type: "checkpoint", Source: "playgroundai/playground-v2.5-1024px-aesthetic", Size: "~7GB", Description: "Aesthetic focused model"},
	},
	"lora": {
		{ID: "lcm-lora-sdxl", Name: "LCM LoRA SDXL", Type: "lora", Source: "latent-consistency/lcm-lora-sdxl", Size: "~400MB", Description: "Fast generation LoRA for SDXL"},
		{ID: "lcm-lora-sd15", Name: "LCM LoRA SD1.5", Type: "lora", Source: "latent-consistency/lcm-lora-sd15", Size: "~130MB", Description: "Fast generation LoRA for SD1.5"},
		{ID: "hyper-sdxl", Name: "Hyper-SDXL LoRA", Type: "lora", Source: "ByteDance/Hyper-SD", Size: "~800MB", Description: "1-step generation LoRA"},
		{ID: "detail-tweaker", Name: "Detail Tweaker XL", Type: "lora", Source: "civitai", Size: "~150MB", Description: "Adjust detail levels"},
	},
	"vae": {
		{ID: "sdxl-vae", Name: "SDXL VAE", Type: "vae", Source: "stabilityai/sdxl-vae", Size: "~335MB", Description: "Official SDXL VAE"},
		{ID: "sd-vae-ft-mse", Name: "SD VAE ft-mse", Type: "vae", Source: "stabilityai/sd-vae-ft-mse", Size: "~335MB", Description: "Fine-tuned SD VAE with better faces"},
		{ID: "flux-vae", Name: "Flux VAE", Type: "vae", Source: "black-forest-labs/FLUX.1-dev", Size: "~335MB", Description: "Flux VAE (ae.safetensors)"},
	},
	"controlnet": {
		{ID: "cn-canny-sdxl", Name: "ControlNet Canny SDXL", Type: "controlnet", Source: "diffusers/controlnet-canny-sdxl-1.0", Size: "~2.5GB", Description: "Canny edge detection for SDXL"},
		{ID: "cn-depth-sdxl", Name: "ControlNet Depth SDXL", Type: "controlnet", Source: "diffusers/controlnet-depth-sdxl-1.0", Size: "~2.5GB", Description: "Depth conditioning for SDXL"},
		{ID: "cn-openpose-sd15", Name: "ControlNet OpenPose", Type: "controlnet", Source: "lllyasviel/control_v11p_sd15_openpose", Size: "~1.5GB", Description: "Pose conditioning for SD1.5"},
		{ID: "cn-tile-sd15", Name: "ControlNet Tile", Type: "controlnet", Source: "lllyasviel/control_v11f1e_sd15_tile", Size: "~1.5GB", Description: "Tile-based upscaling control"},
		{ID: "cn-inpaint-sd15", Name: "ControlNet Inpaint", Type: "controlnet", Source: "lllyasviel/control_v11p_sd15_inpaint", Size: "~1.5GB", Description: "Inpainting guidance"},
	},
	"clip": {
		{ID: "clip-vit-large", Name: "CLIP ViT-L/14", Type: "clip", Source: "openai/clip-vit-large-patch14", Size: "~900MB", Description: "OpenAI CLIP for SD1.5"},
		{ID: "clip-g", Name: "CLIP-G", Type: "clip", Source: "stabilityai/stable-diffusion-xl-base-1.0", Size: "~3.5GB", Description: "Large CLIP for SDXL"},
		{ID: "t5xxl", Name: "T5-XXL", Type: "clip", Source: "google/t5-v1_1-xxl", Size: "~9GB", Description: "T5 encoder for Flux/SD3"},
		{ID: "t5xxl-fp8", Name: "T5-XXL FP8", Type: "clip", Source: "comfyanonymous/flux_text_encoders", Size: "~5GB", Description: "FP8 quantized T5 for Flux"},
	},
	"upscaler": {
		{ID: "realesrgan-x4", Name: "Real-ESRGAN x4", Type: "upscaler", Source: "ai-forever/Real-ESRGAN", Size: "~64MB", Description: "4x upscaling with Real-ESRGAN"},
		{ID: "4x-ultrasharp", Name: "4x-UltraSharp", Type: "upscaler", Source: "civitai", Size: "~67MB", Description: "Sharp 4x upscaler"},
		{ID: "4x-nmkd-siax", Name: "4x-NMKD-Siax", Type: "upscaler", Source: "civitai", Size: "~67MB", Description: "Balanced 4x upscaler"},
	},
	"embedding": {
		{ID: "easynegative", Name: "EasyNegative", Type: "embedding", Source: "civitai", Size: "~25KB", Description: "Negative embedding for quality"},
		{ID: "badhandv4", Name: "BadHand v4", Type: "embedding", Source: "civitai", Size: "~25KB", Description: "Fix bad hands"},
		{ID: "bad-artist", Name: "Bad Artist", Type: "embedding", Source: "civitai", Size: "~25KB", Description: "Negative for bad art"},
	},
	"video": {
		{ID: "svd", Name: "Stable Video Diffusion", Type: "video", Source: "stabilityai/stable-video-diffusion-img2vid", Size: "~10GB", Description: "Image to video generation"},
		{ID: "svd-xt", Name: "SVD-XT", Type: "video", Source: "stabilityai/stable-video-diffusion-img2vid-xt", Size: "~10GB", Description: "Extended SVD with longer videos"},
		{ID: "animatediff-v3", Name: "AnimateDiff v3", Type: "video", Source: "guoyww/animatediff-motion-adapter-v1-5-3", Size: "~1.5GB", Description: "Animation motion module"},
		{ID: "cogvideox-5b", Name: "CogVideoX-5B", Type: "video", Source: "THUDM/CogVideoX-5b", Size: "~14GB", Description: "Text to video model"},
		{ID: "hunyuan-video", Name: "HunyuanVideo", Type: "video", Source: "tencent/HunyuanVideo", Size: "~20GB", Description: "High-quality T2V"},
	},
	"ipadapter": {
		{ID: "ip-adapter-sdxl", Name: "IP-Adapter SDXL", Type: "ipadapter", Source: "h94/IP-Adapter", Size: "~700MB", Description: "Image prompt adapter for SDXL"},
		{ID: "ip-adapter-plus-sdxl", Name: "IP-Adapter Plus SDXL", Type: "ipadapter", Source: "h94/IP-Adapter", Size: "~850MB", Description: "Enhanced IP-Adapter"},
		{ID: "ip-adapter-faceid", Name: "IP-Adapter FaceID", Type: "ipadapter", Source: "h94/IP-Adapter-FaceID", Size: "~200MB", Description: "Face-specific adapter"},
		{ID: "instantid", Name: "InstantID", Type: "ipadapter", Source: "InstantX/InstantID", Size: "~2GB", Description: "Zero-shot identity preservation"},
	},
}

func runComfyManager(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("COMFYUI MANAGER"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Comprehensive ComfyUI package management from the command line"))
	fmt.Println()

	// Show overview
	fmt.Println(theme.GlowStyle.Render("Available Commands"))
	fmt.Println()

	commands := []struct {
		cmd   string
		desc  string
		emoji string
	}{
		{"anime comfyui manager nodes", "Manage custom nodes (install, remove, update)", "🔌"},
		{"anime comfyui manager models", "Manage models (checkpoints, LoRAs, VAEs)", "📦"},
		{"anime comfyui manager workflows", "Browse and install workflows", "🎨"},
		{"anime comfyui manager extensions", "Manage extensions", "🧩"},
	}

	for _, c := range commands {
		fmt.Printf("  %s %s\n", c.emoji, theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	// Quick stats
	homeDir, _ := os.UserHomeDir()
	comfyPath := filepath.Join(homeDir, "ComfyUI")

	fmt.Println(theme.GlowStyle.Render("Installation Status"))
	fmt.Println()

	if _, err := os.Stat(comfyPath); os.IsNotExist(err) {
		fmt.Printf("  %s ComfyUI: %s\n", "❌", theme.WarningStyle.Render("Not installed"))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render("Install with: anime install comfyui"))
	} else {
		fmt.Printf("  %s ComfyUI: %s\n", "✓", theme.SuccessStyle.Render("Installed"))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(comfyPath))

		// Count custom nodes
		customNodesPath := filepath.Join(comfyPath, "custom_nodes")
		if entries, err := os.ReadDir(customNodesPath); err == nil {
			nodeCount := 0
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "__pycache__" {
					nodeCount++
				}
			}
			fmt.Printf("  %s Custom Nodes: %s\n", theme.SymbolSparkle, theme.InfoStyle.Render(fmt.Sprintf("%d installed", nodeCount)))
		}

		// Count models
		modelsPath := filepath.Join(comfyPath, "models")
		if _, err := os.Stat(modelsPath); err == nil {
			fmt.Printf("  %s Models: %s\n", theme.SymbolSparkle, theme.InfoStyle.Render(modelsPath))
		}
	}
	fmt.Println()

	return nil
}

func runComfyNodes(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔌 CUSTOM NODES"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Manage ComfyUI custom nodes"))
	fmt.Println()

	fmt.Println(theme.GlowStyle.Render("Commands"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime comfyui manager nodes list", "List installed custom nodes"},
		{"anime comfyui manager nodes catalog", "Browse available nodes to install"},
		{"anime comfyui manager nodes install <name>", "Install a custom node"},
		{"anime comfyui manager nodes remove <name>", "Remove a custom node"},
		{"anime comfyui manager nodes update [name]", "Update nodes (all or specific)"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	return nil
}

func runComfyNodesList(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔌 INSTALLED CUSTOM NODES"))
	fmt.Println()

	homeDir, _ := os.UserHomeDir()
	customNodesPath := filepath.Join(homeDir, "ComfyUI", "custom_nodes")

	if _, err := os.Stat(customNodesPath); os.IsNotExist(err) {
		fmt.Println(theme.WarningStyle.Render("ComfyUI not found or custom_nodes directory missing"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Install ComfyUI: anime install comfyui"))
		return nil
	}

	entries, err := os.ReadDir(customNodesPath)
	if err != nil {
		return fmt.Errorf("failed to read custom_nodes: %w", err)
	}

	nodes := []string{}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "__pycache__" {
			nodes = append(nodes, e.Name())
		}
	}

	if len(nodes) == 0 {
		fmt.Println(theme.WarningStyle.Render("No custom nodes installed"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Browse available: anime comfyui manager nodes catalog"))
		return nil
	}

	sort.Strings(nodes)

	fmt.Printf("  Found %s custom nodes:\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(nodes))))
	fmt.Println()

	for _, node := range nodes {
		nodePath := filepath.Join(customNodesPath, node)

		// Check if it's a git repo
		gitPath := filepath.Join(nodePath, ".git")
		isGit := false
		if _, err := os.Stat(gitPath); err == nil {
			isGit = true
		}

		gitBadge := ""
		if isGit {
			gitBadge = theme.DimTextStyle.Render(" (git)")
		}

		fmt.Printf("  %s %s%s\n", "✓", theme.SuccessStyle.Render(node), gitBadge)
	}
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Location: %s", customNodesPath)))
	fmt.Println()

	return nil
}

func runComfyNodesCatalog(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔌 CUSTOM NODES CATALOG"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Popular custom nodes for ComfyUI"))
	fmt.Println(theme.DimTextStyle.Render("Install with: anime comfyui manager nodes install <name>"))
	fmt.Println()

	// Group by category
	categories := make(map[string][]CustomNode)
	categoryOrder := []string{"Essential", "ControlNet", "Video", "Upscale", "Flux", "Utility", "Face", "Sampling"}

	for _, node := range customNodesCatalog {
		categories[node.Category] = append(categories[node.Category], node)
	}

	categoryEmojis := map[string]string{
		"Essential":  "⭐",
		"ControlNet": "🎛️",
		"Video":      "🎬",
		"Upscale":    "🔧",
		"Flux":       "⚡",
		"Utility":    "🔨",
		"Face":       "👤",
		"Sampling":   "🎲",
	}

	for _, cat := range categoryOrder {
		nodes := categories[cat]
		if len(nodes) == 0 {
			continue
		}

		emoji := categoryEmojis[cat]
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("%s %s (%d nodes)\n", emoji, theme.HighlightStyle.Render(cat), len(nodes))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, node := range nodes {
			fmt.Printf("  %s %s %s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(node.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("⭐%s", node.Stars)))
			fmt.Printf("     %s\n", theme.SecondaryTextStyle.Render(node.Description))
			fmt.Printf("     %s %s\n",
				theme.DimTextStyle.Render("by"),
				theme.InfoStyle.Render(node.Author))
			fmt.Printf("     %s\n", theme.DimTextStyle.Render(fmt.Sprintf("anime comfyui manager nodes install %s", strings.ToLower(strings.ReplaceAll(node.Name, " ", "-")))))
			fmt.Println()
		}
	}

	return nil
}

func runComfyNodesInstall(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	fmt.Println()
	fmt.Printf("%s Installing custom node: %s\n", theme.SymbolSparkle, theme.HighlightStyle.Render(nodeName))
	fmt.Println()

	// Find in catalog
	var foundNode *CustomNode
	for _, node := range customNodesCatalog {
		normalizedName := strings.ToLower(strings.ReplaceAll(node.Name, " ", "-"))
		if strings.EqualFold(normalizedName, nodeName) || strings.EqualFold(node.Name, nodeName) {
			foundNode = &node
			break
		}
	}

	if foundNode == nil {
		fmt.Println(theme.WarningStyle.Render("Node not found in catalog"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("You can install from a git URL directly:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("cd ~/ComfyUI/custom_nodes && git clone <repo-url>"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Or browse available nodes:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui manager nodes catalog"))
		return nil
	}

	homeDir, _ := os.UserHomeDir()
	customNodesPath := filepath.Join(homeDir, "ComfyUI", "custom_nodes")

	if _, err := os.Stat(customNodesPath); os.IsNotExist(err) {
		return fmt.Errorf("ComfyUI custom_nodes directory not found. Install ComfyUI first: anime install comfyui")
	}

	repoURL := fmt.Sprintf("https://github.com/%s.git", foundNode.Repo)
	targetPath := filepath.Join(customNodesPath, foundNode.Name)

	// Check if already installed
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Println(theme.WarningStyle.Render("Node already installed"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("To update: anime comfyui manager nodes update " + nodeName))
		return nil
	}

	fmt.Printf("  Cloning from: %s\n", theme.DimTextStyle.Render(repoURL))
	fmt.Println()

	gitCmd := exec.Command("git", "clone", repoURL, targetPath)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	// Check for requirements.txt
	reqPath := filepath.Join(targetPath, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		fmt.Println()
		fmt.Print(theme.DimTextStyle.Render("  Installing Python dependencies... "))
		pipCmd := exec.Command("pip", "install", "-r", reqPath)
		pipCmd.Dir = targetPath
		if err := pipCmd.Run(); err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			fmt.Println(theme.DimTextStyle.Render("  Some dependencies may need manual installation"))
		} else {
			fmt.Println(theme.SuccessStyle.Render("✓"))
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ %s installed successfully!", foundNode.Name)))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Restart ComfyUI to load the new node:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui stop && anime comfyui start"))
	fmt.Println()

	return nil
}

func runComfyNodesRemove(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	homeDir, _ := os.UserHomeDir()
	customNodesPath := filepath.Join(homeDir, "ComfyUI", "custom_nodes")

	// Find the node directory
	entries, _ := os.ReadDir(customNodesPath)
	var targetPath string

	for _, e := range entries {
		if e.IsDir() {
			normalizedEntry := strings.ToLower(strings.ReplaceAll(e.Name(), " ", "-"))
			normalizedInput := strings.ToLower(strings.ReplaceAll(nodeName, " ", "-"))
			if normalizedEntry == normalizedInput || strings.EqualFold(e.Name(), nodeName) {
				targetPath = filepath.Join(customNodesPath, e.Name())
				break
			}
		}
	}

	if targetPath == "" {
		fmt.Println(theme.WarningStyle.Render("Node not found"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("List installed nodes: anime comfyui manager nodes list"))
		return nil
	}

	fmt.Println()
	fmt.Printf("%s Removing: %s\n", "❌", theme.HighlightStyle.Render(filepath.Base(targetPath)))
	fmt.Println()

	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Node removed successfully!"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Restart ComfyUI to apply changes"))
	fmt.Println()

	return nil
}

func runComfyNodesUpdate(cmd *cobra.Command, args []string) error {
	homeDir, _ := os.UserHomeDir()
	customNodesPath := filepath.Join(homeDir, "ComfyUI", "custom_nodes")

	fmt.Println()
	fmt.Println(theme.RenderBanner("🔄 UPDATING CUSTOM NODES"))
	fmt.Println()

	if len(args) > 0 {
		// Update specific node
		nodeName := args[0]
		entries, _ := os.ReadDir(customNodesPath)

		for _, e := range entries {
			if e.IsDir() {
				normalizedEntry := strings.ToLower(strings.ReplaceAll(e.Name(), " ", "-"))
				normalizedInput := strings.ToLower(strings.ReplaceAll(nodeName, " ", "-"))
				if normalizedEntry == normalizedInput || strings.EqualFold(e.Name(), nodeName) {
					return updateSingleNode(filepath.Join(customNodesPath, e.Name()), e.Name())
				}
			}
		}
		return fmt.Errorf("node not found: %s", nodeName)
	}

	// Update all nodes
	entries, err := os.ReadDir(customNodesPath)
	if err != nil {
		return fmt.Errorf("failed to read custom_nodes: %w", err)
	}

	updated := 0
	failed := 0

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || e.Name() == "__pycache__" {
			continue
		}

		nodePath := filepath.Join(customNodesPath, e.Name())
		gitPath := filepath.Join(nodePath, ".git")

		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			continue // Not a git repo
		}

		fmt.Printf("  %s %s... ", "→", e.Name())

		pullCmd := exec.Command("git", "pull", "--ff-only")
		pullCmd.Dir = nodePath
		output, err := pullCmd.CombinedOutput()

		if err != nil {
			fmt.Println(theme.WarningStyle.Render("⚠"))
			failed++
		} else if strings.Contains(string(output), "Already up to date") {
			fmt.Println(theme.DimTextStyle.Render("up to date"))
		} else {
			fmt.Println(theme.SuccessStyle.Render("✓ updated"))
			updated++
		}
	}

	fmt.Println()
	if updated > 0 {
		fmt.Printf("  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("✓ Updated %d nodes", updated)))
	}
	if failed > 0 {
		fmt.Printf("  %s\n", theme.WarningStyle.Render(fmt.Sprintf("⚠ %d nodes had issues", failed)))
	}
	if updated == 0 && failed == 0 {
		fmt.Println(theme.InfoStyle.Render("  All nodes are up to date"))
	}
	fmt.Println()

	if updated > 0 {
		fmt.Println(theme.DimTextStyle.Render("Restart ComfyUI to apply updates:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui stop && anime comfyui start"))
		fmt.Println()
	}

	return nil
}

func updateSingleNode(nodePath, nodeName string) error {
	gitPath := filepath.Join(nodePath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("%s is not a git repository", nodeName)
	}

	fmt.Printf("  Updating %s... ", theme.HighlightStyle.Render(nodeName))

	pullCmd := exec.Command("git", "pull", "--ff-only")
	pullCmd.Dir = nodePath
	output, err := pullCmd.CombinedOutput()

	if err != nil {
		fmt.Println(theme.WarningStyle.Render("⚠"))
		return fmt.Errorf("git pull failed: %s", string(output))
	}

	if strings.Contains(string(output), "Already up to date") {
		fmt.Println(theme.DimTextStyle.Render("already up to date"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ updated"))
	}

	fmt.Println()
	return nil
}

func runComfyModels(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📦 COMFYUI MODELS"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Manage models for ComfyUI"))
	fmt.Println()

	fmt.Println(theme.GlowStyle.Render("Commands"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime comfyui manager models list [type]", "List installed models (optionally by type)"},
		{"anime comfyui manager models catalog", "Browse available models to download"},
		{"anime comfyui manager models download <id>", "Download a model"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.GlowStyle.Render("Model Types"))
	fmt.Println()

	types := []struct {
		t     string
		desc  string
		emoji string
	}{
		{"checkpoint", "Base models (SD, SDXL, Flux)", "🎨"},
		{"lora", "LoRA adapters for style/concept", "🎭"},
		{"vae", "Variational autoencoders", "🔄"},
		{"controlnet", "ControlNet models", "🎛️"},
		{"clip", "Text encoders", "📝"},
		{"upscaler", "Upscaling models", "🔍"},
		{"embedding", "Textual inversions", "💬"},
		{"video", "Video generation models", "🎬"},
		{"ipadapter", "IP-Adapter models", "🖼️"},
	}

	for _, t := range types {
		fmt.Printf("  %s %s - %s\n", t.emoji, theme.HighlightStyle.Render(t.t), theme.DimTextStyle.Render(t.desc))
	}
	fmt.Println()

	return nil
}

func runComfyModelsList(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📦 INSTALLED MODELS"))
	fmt.Println()

	homeDir, _ := os.UserHomeDir()
	modelsPath := filepath.Join(homeDir, "ComfyUI", "models")

	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		fmt.Println(theme.WarningStyle.Render("ComfyUI models directory not found"))
		return nil
	}

	modelDirs := map[string]string{
		"checkpoints":    "🎨 Checkpoints",
		"loras":          "🎭 LoRAs",
		"vae":            "🔄 VAE",
		"controlnet":     "🎛️ ControlNet",
		"clip":           "📝 CLIP",
		"upscale_models": "🔍 Upscalers",
		"embeddings":     "💬 Embeddings",
		"animatediff":    "🎬 AnimateDiff",
		"ipadapter":      "🖼️ IP-Adapter",
	}

	filterType := ""
	if len(args) > 0 {
		filterType = strings.ToLower(args[0])
	}

	totalModels := 0
	totalSize := int64(0)

	for dir, label := range modelDirs {
		if filterType != "" && !strings.Contains(strings.ToLower(dir), filterType) {
			continue
		}

		dirPath := filepath.Join(modelsPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		files, err := findModelFiles(dirPath)
		if err != nil || len(files) == 0 {
			continue
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("%s (%d)\n", theme.HighlightStyle.Render(label), len(files))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, f := range files {
			totalModels++
			totalSize += f.size

			sizeStr := formatSizeBytes(f.size)
			fmt.Printf("  %s %s %s\n",
				"✓",
				theme.SuccessStyle.Render(f.name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", sizeStr)))
		}
		fmt.Println()
	}

	if totalModels == 0 {
		fmt.Println(theme.WarningStyle.Render("No models found"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Browse available: anime comfyui manager models catalog"))
	} else {
		fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(fmt.Sprintf("📦 Total: %d models", totalModels)),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s)", formatSizeBytes(totalSize))))
		fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	}
	fmt.Println()

	return nil
}

type modelFile struct {
	name string
	size int64
}

func findModelFiles(dir string) ([]modelFile, error) {
	var files []modelFile
	extensions := []string{".safetensors", ".ckpt", ".pth", ".bin", ".pt"}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(info.Name()), ext) {
				files = append(files, modelFile{
					name: info.Name(),
					size: info.Size(),
				})
				break
			}
		}
		return nil
	})

	return files, err
}

func formatSizeBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func runComfyModelsCatalog(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📦 MODELS CATALOG"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Popular models for ComfyUI"))
	fmt.Println(theme.DimTextStyle.Render("Download with: anime comfyui manager models download <id>"))
	fmt.Println()

	typeEmojis := map[string]string{
		"checkpoint":  "🎨",
		"lora":        "🎭",
		"vae":         "🔄",
		"controlnet":  "🎛️",
		"clip":        "📝",
		"upscaler":    "🔍",
		"embedding":   "💬",
		"video":       "🎬",
		"ipadapter":   "🖼️",
	}

	typeOrder := []string{"checkpoint", "lora", "vae", "controlnet", "clip", "upscaler", "embedding", "video", "ipadapter"}

	for _, modelType := range typeOrder {
		models := comfyModelsCatalogData[modelType]
		if len(models) == 0 {
			continue
		}

		emoji := typeEmojis[modelType]
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("%s %s (%d available)\n", emoji, theme.HighlightStyle.Render(strings.Title(modelType)), len(models))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, model := range models {
			fmt.Printf("  %s %s %s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(model.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", model.Size)))
			fmt.Printf("     %s\n", theme.SecondaryTextStyle.Render(model.Description))
			fmt.Printf("     %s\n", theme.DimTextStyle.Render(fmt.Sprintf("anime comfyui manager models download %s", model.ID)))
			fmt.Println()
		}
	}

	return nil
}

func runComfyModelsDownload(cmd *cobra.Command, args []string) error {
	modelID := args[0]

	fmt.Println()
	fmt.Printf("%s Downloading model: %s\n", theme.SymbolSparkle, theme.HighlightStyle.Render(modelID))
	fmt.Println()

	// Find model in catalog
	var foundModel *ComfyModel
	for _, models := range comfyModelsCatalogData {
		for _, model := range models {
			if strings.EqualFold(model.ID, modelID) {
				foundModel = &model
				break
			}
		}
		if foundModel != nil {
			break
		}
	}

	if foundModel == nil {
		fmt.Println(theme.WarningStyle.Render("Model not found in catalog"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Browse available: anime comfyui manager models catalog"))
		return nil
	}

	homeDir, _ := os.UserHomeDir()

	// Determine target directory
	targetDirMap := map[string]string{
		"checkpoint":  "checkpoints",
		"lora":        "loras",
		"vae":         "vae",
		"controlnet":  "controlnet",
		"clip":        "clip",
		"upscaler":    "upscale_models",
		"embedding":   "embeddings",
		"video":       "checkpoints",
		"ipadapter":   "ipadapter",
	}

	targetDir := filepath.Join(homeDir, "ComfyUI", "models", targetDirMap[foundModel.Type])

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("  Model: %s\n", theme.InfoStyle.Render(foundModel.Name))
	fmt.Printf("  Size: %s\n", theme.InfoStyle.Render(foundModel.Size))
	fmt.Printf("  Source: %s\n", theme.DimTextStyle.Render(foundModel.Source))
	fmt.Printf("  Target: %s\n", theme.DimTextStyle.Render(targetDir))
	fmt.Println()

	// Use huggingface-cli to download
	if strings.Contains(foundModel.Source, "/") && !strings.HasPrefix(foundModel.Source, "http") {
		fmt.Println(theme.InfoStyle.Render("Downloading from HuggingFace..."))
		fmt.Println()

		dlCmd := exec.Command("huggingface-cli", "download", foundModel.Source, "--local-dir", targetDir)
		dlCmd.Stdout = os.Stdout
		dlCmd.Stderr = os.Stderr

		if err := dlCmd.Run(); err != nil {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("huggingface-cli failed. Try manually:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("huggingface-cli download %s --local-dir %s", foundModel.Source, targetDir)))
			return nil
		}
	} else {
		fmt.Println(theme.WarningStyle.Render("Manual download required"))
		fmt.Println()
		fmt.Printf("  Source: %s\n", theme.HighlightStyle.Render(foundModel.Source))
		fmt.Printf("  Target: %s\n", theme.DimTextStyle.Render(targetDir))
		return nil
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ %s downloaded successfully!", foundModel.Name)))
	fmt.Println()

	return nil
}

func runComfyWorkflows(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 COMFYUI WORKFLOWS"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Browse and manage ComfyUI workflows"))
	fmt.Println()

	fmt.Println(theme.GlowStyle.Render("Workflow Sources"))
	fmt.Println()

	sources := []struct {
		name  string
		url   string
		desc  string
		emoji string
	}{
		{"OpenArt", "https://openart.ai/workflows", "Large collection of community workflows", "🌐"},
		{"ComfyWorkflows", "https://comfyworkflows.com", "Curated workflow collection", "📚"},
		{"Civitai", "https://civitai.com/models?types=Workflow", "Workflows on Civitai", "🎭"},
		{"GitHub", "https://github.com/topics/comfyui-workflow", "Open source workflows", "💻"},
	}

	for _, s := range sources {
		fmt.Printf("  %s %s\n", s.emoji, theme.HighlightStyle.Render(s.name))
		fmt.Printf("     %s\n", theme.SecondaryTextStyle.Render(s.desc))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(s.url))
		fmt.Println()
	}

	fmt.Println(theme.GlowStyle.Render("Local Workflows"))
	fmt.Println()

	homeDir, _ := os.UserHomeDir()
	workflowsPath := filepath.Join(homeDir, "ComfyUI", "user", "default", "workflows")

	if entries, err := os.ReadDir(workflowsPath); err == nil {
		jsonCount := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				jsonCount++
			}
		}
		if jsonCount > 0 {
			fmt.Printf("  %s Found %s local workflows\n", "✓", theme.HighlightStyle.Render(fmt.Sprintf("%d", jsonCount)))
			fmt.Printf("     %s\n", theme.DimTextStyle.Render(workflowsPath))
		} else {
			fmt.Printf("  %s No local workflows found\n", "❌")
		}
	} else {
		fmt.Printf("  %s Workflows directory not found\n", "❌")
	}
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render("Tip: Save workflows in ComfyUI using the 'Save' button"))
	fmt.Println(theme.DimTextStyle.Render("     or copy .json files to the workflows directory"))
	fmt.Println()

	return nil
}

func runComfyExtensions(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🧩 COMFYUI EXTENSIONS"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Extensions enhance ComfyUI functionality"))
	fmt.Println()

	fmt.Println(theme.GlowStyle.Render("Core Extensions"))
	fmt.Println()

	extensions := []struct {
		name  string
		desc  string
		emoji string
	}{
		{"ComfyUI-Manager", "The essential manager for custom nodes", "⭐"},
		{"Custom Scripts", "Autocomplete, image feed, workflow tools", "📜"},
		{"Impact Pack", "Face detection, SAM, upscaling nodes", "💥"},
		{"Inspire Pack", "Regional prompting, wildcards, utilities", "✨"},
	}

	for _, ext := range extensions {
		fmt.Printf("  %s %s\n", ext.emoji, theme.HighlightStyle.Render(ext.name))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(ext.desc))
		fmt.Println()
	}

	fmt.Println(theme.DimTextStyle.Render("Extensions are managed as custom nodes."))
	fmt.Println(theme.DimTextStyle.Render("Use: anime comfyui manager nodes catalog"))
	fmt.Println()

	return nil
}
