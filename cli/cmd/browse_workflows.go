package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var browseWorkflowsCmd = &cobra.Command{
	Use:   "browse-workflows",
	Short: "Browse AI workflows and pipelines",
	Long:  "Explore different AI workflows: inference, animation, image generation, training, and more",
	Run:   runWorkflows,
}

func init() {
	rootCmd.AddCommand(browseWorkflowsCmd)
}

type Workflow struct {
	Name        string
	Category    string
	Emoji       string
	Description string
	Stack       []string
	Command     string
}

func runWorkflows(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ AI WORKFLOWS ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🎨 Complete AI pipelines from idea to output"))
	fmt.Println()

	workflows := []Workflow{
		{
			Name:        "Text-to-Video Generation",
			Category:    "Animation",
			Emoji:       "🎬",
			Description: "Generate videos from text prompts using state-of-the-art models",
			Stack:       []string{"Claude Code", "ComfyUI", "Mochi/CogVideoX/LTXVideo"},
			Command:     "anime install comfyui mochi ltxvideo",
		},
		{
			Name:        "Image-to-Video Animation",
			Category:    "Animation",
			Emoji:       "📹",
			Description: "Animate static images with motion using AnimateDiff or SVD",
			Stack:       []string{"ComfyUI", "AnimateDiff", "Stable Video Diffusion"},
			Command:     "anime install comfyui animatediff svd",
		},
		{
			Name:        "Character Generation",
			Category:    "Image Generation",
			Emoji:       "👤",
			Description: "Create consistent characters using FLUX or Stable Diffusion",
			Stack:       []string{"ComfyUI", "FLUX", "LoRA Training"},
			Command:     "anime install comfyui pytorch",
		},
		{
			Name:        "Image Enhancement & Upscaling",
			Category:    "Image Enhancement",
			Emoji:       "✨",
			Description: "Upscale and enhance images with AI",
			Stack:       []string{"ComfyUI", "ESRGAN", "Real-ESRGAN"},
			Command:     "anime install comfyui",
		},
		{
			Name:        "LLM Inference & Chat",
			Category:    "Inference",
			Emoji:       "💬",
			Description: "Run local LLMs for chat, code generation, and reasoning",
			Stack:       []string{"Ollama", "Llama 3.3 70B", "Qwen3", "DeepSeek"},
			Command:     "anime install ollama models-large",
		},
		{
			Name:        "Code Generation Assistant",
			Category:    "Development",
			Emoji:       "💻",
			Description: "AI-powered coding with Claude Code and local models",
			Stack:       []string{"Claude Code", "Ollama", "DeepSeek Coder"},
			Command:     "anime install claude ollama models-medium",
		},
		{
			Name:        "Model Fine-tuning",
			Category:    "Training",
			Emoji:       "🔧",
			Description: "Fine-tune models on custom datasets",
			Stack:       []string{"PyTorch", "Transformers", "Accelerate", "LoRA"},
			Command:     "anime install pytorch",
		},
		{
			Name:        "Research & Experimentation",
			Category:    "Research",
			Emoji:       "🔬",
			Description: "Full ML research stack with notebooks and visualization",
			Stack:       []string{"Python", "PyTorch", "JupyterLab", "Weights & Biases"},
			Command:     "anime install python pytorch",
		},
		{
			Name:        "ComfyUI Web Server",
			Category:    "Setup",
			Emoji:       "🌐",
			Description: "Launch ComfyUI web interface for image/video generation",
			Stack:       []string{"ComfyUI", "Video Models", "Stable Diffusion"},
			Command:     "anime lambda launch comfyui",
		},
		{
			Name:        "Claude → ComfyUI → Video Pipeline",
			Category:    "Streaming",
			Emoji:       "🎥",
			Description: "End-to-end: Claude generates prompts → ComfyUI creates frames → Video model animates",
			Stack:       []string{"Claude Code", "ComfyUI", "CogVideoX", "Streaming Server"},
			Command:     "anime workflows stream video",
		},
	}

	// Group by category
	categories := make(map[string][]Workflow)
	for _, w := range workflows {
		categories[w.Category] = append(categories[w.Category], w)
	}

	// Display workflows by category
	categoryOrder := []string{"Animation", "Image Generation", "Image Enhancement", "Inference", "Development", "Training", "Research", "Setup", "Streaming"}

	for _, cat := range categoryOrder {
		ws, exists := categories[cat]
		if !exists {
			continue
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("  %s\n", theme.InfoStyle.Render(cat))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, w := range ws {
			fmt.Printf("  %s %s\n",
				w.Emoji,
				theme.HighlightStyle.Render(w.Name))
			fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(w.Description))
			fmt.Println()
			fmt.Printf("    %s\n", theme.DimTextStyle.Render("Stack: "+fmt.Sprint(w.Stack)))
			fmt.Printf("    %s\n", theme.GlowStyle.Render("$ "+w.Command))
			fmt.Println()
		}
	}

	// Quick reference
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Quick Reference"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime packages"),
		theme.DimTextStyle.Render("View all packages"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime lambda defaults"),
		theme.DimTextStyle.Render("See recommended starter pack"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime lambda launch <config>"),
		theme.DimTextStyle.Render("Launch predefined configurations"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime serve <path>"),
		theme.DimTextStyle.Render("Serve content on public IP"))
	fmt.Println()
}
