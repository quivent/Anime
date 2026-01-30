package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
)

// modelTitleStyle uses BrightMagenta with black background
var modelTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(theme.BrightMagenta).
	Background(theme.BgBlack).
	Padding(0, 1)

type ModelItem struct {
	Name     string
	Size     string
	Desc     string
	UseCases []string
	Category string
	Type     string
}

func (i ModelItem) Title() string       { return i.Name }
func (i ModelItem) Description() string { return i.Size + " • " + i.Category }
func (i ModelItem) FilterValue() string { return i.Name }

type ModelsModel struct {
	list         list.Model
	models       []ModelItem
	selectedIdx  int
	width        int
	height       int
	showingHelp  bool
	categoryView bool
}

func NewModelsModel() ModelsModel {
	models := getModelCatalog()

	items := make([]list.Item, len(models))
	for i, m := range models {
		items[i] = m
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "🎨 AI Model Catalog"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = modelTitleStyle

	return ModelsModel{
		list:   l,
		models: models,
	}
}

func (m ModelsModel) Init() tea.Cmd {
	return nil
}

func (m ModelsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if i, ok := m.list.SelectedItem().(ModelItem); ok {
				m.selectedIdx = m.list.Index()
				// Could trigger installation here
				_ = i
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("?"))):
			m.showingHelp = !m.showingHelp
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ModelsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	// Show list
	s.WriteString(m.list.View())
	s.WriteString("\n\n")

	// Show detailed view of selected model
	if i, ok := m.list.SelectedItem().(ModelItem); ok {
		s.WriteString(theme.HeaderStyle.Render(fmt.Sprintf("📦 %s [%s]", i.Type, i.Category)))
		s.WriteString("\n\n")

		s.WriteString(theme.HighlightStyle.Render("  " + i.Name))
		s.WriteString("\n")

		detailStyle := lipgloss.NewStyle().Foreground(theme.LightGray).MarginLeft(2)
		s.WriteString(detailStyle.Render("💾 " + i.Size))
		s.WriteString("\n")

		s.WriteString(theme.DescriptionStyle.Render("📝 " + i.Desc))
		s.WriteString("\n\n")

		if len(i.UseCases) > 0 {
			s.WriteString(theme.DescriptionStyle.Render("🎯 Use Cases:"))
			s.WriteString("\n")
			useCaseStyle := lipgloss.NewStyle().Foreground(theme.MediumGray).MarginLeft(4)
			for _, uc := range i.UseCases {
				s.WriteString(useCaseStyle.Render("• " + uc))
				s.WriteString("\n")
			}
		}
		s.WriteString("\n")
	}

	// Help text
	if m.showingHelp {
		s.WriteString(theme.HelpStyle.Render("? - toggle help • ↑/↓ - navigate • enter - select • q - quit"))
	} else {
		s.WriteString(theme.HelpStyle.Render("? for help • q to quit"))
	}

	return s.String()
}

func getModelCatalog() []ModelItem {
	return []ModelItem{
		// LLMs
		{
			Name:     "Llama 3.3 70B",
			Size:     "~40GB",
			Desc:     "Meta's latest open-source flagship model with exceptional reasoning and coding capabilities",
			UseCases: []string{"Code generation", "Complex reasoning", "Creative writing", "Research assistance"},
			Category: "General Purpose",
			Type:     "🤖 Large Language Model",
		},
		{
			Name:     "Llama 3.3 8B",
			Size:     "~5GB",
			Desc:     "Efficient smaller version of Llama 3.3, great balance of performance and speed",
			UseCases: []string{"Fast inference", "Chatbots", "Text summarization", "Simple coding tasks"},
			Category: "Efficient",
			Type:     "🤖 Large Language Model",
		},
		{
			Name:     "Mistral 7B",
			Size:     "~4GB",
			Desc:     "High-performance 7B model outperforming many larger models, excellent for coding",
			UseCases: []string{"Code completion", "Technical writing", "Quick Q&A", "Function generation"},
			Category: "Coding",
			Type:     "🤖 Large Language Model",
		},
		{
			Name:        "Mixtral 8x7B",
			Size:        "~26GB",
			Desc:"Mixture of Experts model with 47B parameters, runs efficiently via sparse activation",
			UseCases:    []string{"Multi-task processing", "Complex instructions", "Long context", "Specialized domains"},
			Category:    "Multi-Task",
			Type:        "🤖 Large Language Model",
		},
		{
			Name:        "Qwen3 235B MoE",
			Size:        "~142GB",
			Desc:"Flagship Qwen3 MoE model with advanced reasoning and coding capabilities",
			UseCases:    []string{"Multilingual tasks", "Mathematics", "Complex reasoning", "Advanced coding"},
			Category:    "Multilingual",
			Type:        "🤖 Large Language Model",
		},
		{
			Name:        "Qwen3 14B",
			Size:        "~9GB",
			Desc:"Mid-size Qwen3 model with excellent multilingual performance",
			UseCases:    []string{"Bilingual applications", "Translation", "Cross-language tasks"},
			Category:    "Bilingual",
			Type:        "🤖 Large Language Model",
		},
		{
			Name:        "DeepSeek Coder 33B",
			Size:        "~18GB",
			Desc:"Specialized coding model trained on 2T+ tokens of code and text",
			UseCases:    []string{"Code generation", "Bug fixing", "Code review", "Documentation"},
			Category:    "Code Specialist",
			Type:        "🤖 Large Language Model",
		},
		{
			Name:        "DeepSeek V3",
			Size:        "~250GB",
			Desc:"Latest frontier model with 671B parameters using MoE architecture",
			UseCases:    []string{"Research", "Complex reasoning", "Frontier capabilities", "Benchmarks"},
			Category:    "Frontier",
			Type:        "🤖 Large Language Model",
		},
		{
			Name:        "Phi-3.5 Mini (3.8B)",
			Size:        "~2GB",
			Desc:"Microsoft's compact model with strong reasoning despite small size",
			UseCases:    []string{"Edge deployment", "Mobile apps", "Resource-constrained environments"},
			Category:    "Compact",
			Type:        "🤖 Large Language Model",
		},

		// Image Generation
		{
			Name:        "Stable Diffusion XL (SDXL)",
			Size:        "~7GB",
			Desc:"Latest Stable Diffusion with improved image quality and composition",
			UseCases:    []string{"High-quality images", "Concept art", "Product visualization", "Marketing materials"},
			Category:    "General Purpose",
			Type:        "🎨 Image Generation",
		},
		{
			Name:        "Stable Diffusion 1.5",
			Size:        "~4GB",
			Desc:"Widely-used base model with huge ecosystem of fine-tunes and LoRAs",
			UseCases:    []string{"Art generation", "Style transfer", "Photo editing", "Custom training"},
			Category:    "Classic",
			Type:        "🎨 Image Generation",
		},
		{
			Name:        "Flux.1 Dev",
			Size:        "~12GB",
			Desc:"Black Forest Labs' new model with exceptional prompt following and quality",
			UseCases:    []string{"Photorealism", "Typography", "Complex compositions", "Professional work"},
			Category:    "Professional",
			Type:        "🎨 Image Generation",
		},
		{
			Name:        "Flux.1 Schnell",
			Size:        "~12GB",
			Desc:"Fast version of Flux optimized for speed while maintaining quality",
			UseCases:    []string{"Rapid prototyping", "Iterative design", "Real-time generation"},
			Category:    "Fast",
			Type:        "🎨 Image Generation",
		},

		// Video Generation
		{
			Name:        "Stable Video Diffusion",
			Size:        "~10GB",
			Desc:"Stability AI's image-to-video model for smooth animations",
			UseCases:    []string{"Product demos", "Animation", "Visual effects", "Motion design"},
			Category:    "Image-to-Video",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "AnimateDiff",
			Size:        "~4GB",
			Desc:"Motion module for Stable Diffusion, animates still images",
			UseCases:    []string{"Character animation", "Scene transitions", "Loop creation"},
			Category:    "Animation",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "Mochi-1",
			Size:        "~12GB",
			Desc:"Open source video generation model with 10B parameters",
			UseCases:    []string{"Text-to-video", "Creative videos", "Experimental content"},
			Category:    "Text-to-Video",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "CogVideoX-5B",
			Size:        "~14GB",
			Desc:"Open source text-to-video with strong temporal consistency",
			UseCases:    []string{"Video creation", "Content production", "Social media"},
			Category:    "Content Creation",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "Open-Sora 2.0",
			Size:        "~16GB",
			Desc:"High-quality video generation with advanced architectures",
			UseCases:    []string{"High-res video", "Professional content", "Research"},
			Category:    "High Quality",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "LTXVideo",
			Size:        "~7GB",
			Desc:"Fast video generation using latent transformers",
			UseCases:    []string{"Quick videos", "Rapid iteration", "Previews"},
			Category:    "Fast Generation",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "Wan2.2",
			Size:        "~10GB",
			Desc:"State-of-the-art image-to-video with exceptional quality",
			UseCases:    []string{"Professional videos", "Cinematic effects", "High-quality output"},
			Category:    "Professional",
			Type:        "🎬 Video Generation",
		},
		{
			Name:        "Topaz Video AI",
			Size:        "~5GB",
			Desc:"Commercial video enhancement and upscaling tool using AI",
			UseCases:    []string{"Video upscaling", "Frame interpolation", "Denoising", "Quality enhancement"},
			Category:    "Enhancement",
			Type:        "🎬 Video Generation",
		},
	}
}
