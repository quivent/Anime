package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

type pipeline struct {
	Name    string
	Cmd     string
	Desc    string
}

type pipelineGroup struct {
	Name      string
	Color     lipgloss.Color
	Pipelines []pipeline
}

var pipelineGroups = []pipelineGroup{
	{
		Name:  "Video Generation",
		Color: theme.SakuraPink,
		Pipelines: []pipeline{
			{"Wan T2V", "anime wan generate", "Text-to-video with Wan2.1"},
			{"Wan I2V", "anime wan generate --image input.png", "Image-to-video"},
			{"Reel", "anime reel", "Multi-shot video pipeline"},
			{"Upscale", "anime upscale", "Video super-resolution"},
		},
	},
	{
		Name:  "Image & Animation",
		Color: theme.ElectricBlue,
		Pipelines: []pipeline{
			{"Animate", "anime animate", "Image-to-animation pipeline"},
			{"Collection", "anime collection transform", "Batch transform images"},
			{"Enhance", "anime enhance", "AI upscale + restore"},
			{"Sequence", "anime sequence", "Frame sequence generation"},
		},
	},
	{
		Name:  "Model Serving",
		Color: theme.NeonPurple,
		Pipelines: []pipeline{
			{"vLLM", "anime vllm start", "LLM inference server"},
			{"ComfyUI", "anime comfyui start", "Node-based generation"},
			{"Ollama", "anime ollama", "Local model runner"},
			{"Serve", "anime serve", "API server with auth"},
		},
	},
	{
		Name:  "Deployment",
		Color: theme.MintGreen,
		Pipelines: []pipeline{
			{"Push", "anime push <server>", "Deploy CLI + assets to remote"},
			{"Bootstrap", "anime bootstrap <server>", "Full server setup"},
			{"Install", "anime install", "Install models + tools"},
			{"Config", "anime config", "Configure server modules"},
		},
	},
	{
		Name:  "Development",
		Color: theme.SunsetOrange,
		Pipelines: []pipeline{
			{"Clone", "anime clone anime", "Clone this project"},
			{"Source Extract", "anime source extract", "Extract embedded source"},
			{"Pkg Publish", "anime pkg publish", "Publish a package"},
			{"Embed Token", "anime embed token <slot>", "Bake auth tokens"},
		},
	},
}

var pipelinesCmd = &cobra.Command{
	Use:   "pipelines",
	Short: "Show common pipelines and workflows",
	Long:  `Display available pipelines grouped by category with usage examples.`,
	Run:   runPipelines,
}

func init() {
	rootCmd.AddCommand(pipelinesCmd)
}

func runPipelines(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ PIPELINES ⚡"))
	fmt.Println()

	for _, group := range pipelineGroups {
		groupStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(group.Color)
		cmdStyle := lipgloss.NewStyle().
			Foreground(group.Color)
		descStyle := lipgloss.NewStyle().
			Foreground(theme.TextSecondary)

		fmt.Printf("  %s\n", groupStyle.Render("┌─ "+group.Name))

		for i, p := range group.Pipelines {
			connector := "├"
			if i == len(group.Pipelines)-1 {
				connector = "└"
			}
			fmt.Printf("  %s %s  %s\n",
				groupStyle.Render(connector+"─"),
				cmdStyle.Render(p.Cmd),
				descStyle.Render("— "+p.Desc),
			)
		}
		fmt.Println()
	}
}
