package theme

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Anime-inspired color palette
var (
	// Primary colors - vibrant anime aesthetic
	SakuraPink    = lipgloss.Color("#FF69B4") // Hot pink
	ElectricBlue  = lipgloss.Color("#00D9FF") // Bright cyan
	NeonPurple    = lipgloss.Color("#BD93F9") // Purple
	MintGreen     = lipgloss.Color("#50FA7B") // Green
	SunsetOrange  = lipgloss.Color("#FFB86C") // Orange
	LavenderMist  = lipgloss.Color("#D4BFFF") // Light purple

	// Accent colors
	ActionRed     = lipgloss.Color("#FF5555") // Red
	WarningYellow = lipgloss.Color("#F1FA8C") // Yellow

	// Neutral colors
	TextPrimary   = lipgloss.Color("#F8F8F2") // Off-white
	TextSecondary = lipgloss.Color("#B4B4B4") // Gray
	TextDim       = lipgloss.Color("#6272A4") // Dim blue-gray

	// Background colors
	BgDark        = lipgloss.Color("#282A36") // Dark bg
	BgAccent      = lipgloss.Color("#44475A") // Accent bg
)

// Common styles with anime aesthetic
var (
	// Headers - bold and colorful
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SakuraPink).
			MarginBottom(1).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ElectricBlue).
			MarginTop(1)

	SubHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(NeonPurple)

	// Content styles
	PrimaryTextStyle = lipgloss.NewStyle().
				Foreground(TextPrimary)

	SecondaryTextStyle = lipgloss.NewStyle().
				Foreground(TextSecondary)

	DimTextStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			Italic(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(MintGreen).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ActionRed).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningYellow).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ElectricBlue)

	// Special effects
	GlowStyle = lipgloss.NewStyle().
			Foreground(SakuraPink).
			Bold(true).
			Underline(true)

	HighlightStyle = lipgloss.NewStyle().
			Background(NeonPurple).
			Foreground(BgDark).
			Bold(true).
			Padding(0, 1)

	// Borders and boxes
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(NeonPurple).
			Padding(0, 1)

	AccentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(SakuraPink).
			Padding(1, 2)

	// Category styles - different color for each
	CategoryFoundation = lipgloss.NewStyle().
				Bold(true).
				Foreground(ElectricBlue).
				MarginTop(1)

	CategoryMLFramework = lipgloss.NewStyle().
				Bold(true).
				Foreground(NeonPurple).
				MarginTop(1)

	CategoryLLMRuntime = lipgloss.NewStyle().
				Bold(true).
				Foreground(SakuraPink).
				MarginTop(1)

	CategoryModels = lipgloss.NewStyle().
			Bold(true).
			Foreground(SunsetOrange).
			MarginTop(1)

	CategoryApplication = lipgloss.NewStyle().
				Bold(true).
				Foreground(MintGreen).
				MarginTop(1)
)

// Anime-themed emojis and symbols
var (
	// Package/module symbols
	SymbolPackage   = "📦"
	SymbolModule    = "🎴"
	SymbolComponent = "⚡"

	// Status symbols
	SymbolSuccess   = "✨"
	SymbolError     = "💥"
	SymbolWarning   = "⚠️"
	SymbolInfo      = "💫"
	SymbolLoading   = "🌸"

	// Action symbols
	SymbolInstall   = "🚀"
	SymbolDownload  = "⬇️"
	SymbolBuild     = "🔨"
	SymbolDeploy    = "🎯"
	SymbolConfig    = "⚙️"

	// Tree symbols
	SymbolTree      = "🌸"
	SymbolBranch    = "├──"
	SymbolLastBranch = "└──"
	SymbolPipe      = "│"
	SymbolSpace     = "   "

	// Anime-themed
	SymbolSakura    = "🌸"
	SymbolStar      = "⭐"
	SymbolSparkle   = "✨"
	SymbolBolt      = "⚡"
	SymbolHeart     = "💖"
	SymbolShield    = "🛡️"
	SymbolSword     = "⚔️"
	SymbolMagic     = "🔮"
)

// Get category style by name
func GetCategoryStyle(category string) lipgloss.Style {
	switch category {
	case "Foundation":
		return CategoryFoundation
	case "ML Framework":
		return CategoryMLFramework
	case "LLM Runtime":
		return CategoryLLMRuntime
	case "Models":
		return CategoryModels
	case "Application":
		return CategoryApplication
	default:
		return HeaderStyle
	}
}

// Render a stylized title banner
func RenderBanner(text string) string {
	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(SakuraPink).
		Background(BgDark).
		Padding(1, 2).
		Margin(1, 0).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(NeonPurple).
		Align(lipgloss.Center).
		Width(60)

	return banner.Render(text)
}

// Render progress bar with anime aesthetic
func RenderProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	barStyle := lipgloss.NewStyle().Foreground(SakuraPink)

	return fmt.Sprintf("%s %s %d%%",
		barStyle.Render(bar),
		SymbolSakura,
		int(percent*100))
}
