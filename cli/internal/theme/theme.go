package theme

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Anime-inspired color palette
//
// This package provides a centralized theme for the anime CLI, ensuring
// consistent styling across all commands and TUI components.
var (
	// Primary colors - vibrant anime aesthetic
	SakuraPink    = lipgloss.Color("#FF69B4") // Hot pink - primary accent
	ElectricBlue  = lipgloss.Color("#00D9FF") // Bright cyan - headers
	NeonPurple    = lipgloss.Color("#BD93F9") // Purple - highlights
	MintGreen     = lipgloss.Color("#50FA7B") // Green - success
	SunsetOrange  = lipgloss.Color("#FFB86C") // Orange - warnings
	LavenderMist  = lipgloss.Color("#D4BFFF") // Light purple - subtle accent

	// Accent colors
	ActionRed     = lipgloss.Color("#FF5555") // Red - errors
	WarningYellow = lipgloss.Color("#F1FA8C") // Yellow - warnings
	GoldYellow    = lipgloss.Color("#FFD700") // Gold - special highlights

	// Neutral colors
	TextPrimary   = lipgloss.Color("#F8F8F2") // Off-white - primary text
	TextSecondary = lipgloss.Color("#B4B4B4") // Gray - secondary text
	TextDim       = lipgloss.Color("#6272A4") // Dim blue-gray - muted text
	TextDimGray   = lipgloss.Color("#888888") // Dim gray - unselected items
	TextDarkGray  = lipgloss.Color("#444444") // Dark gray - very dim
	TextHelp      = lipgloss.Color("#666666") // Help text gray
	TextMuted     = lipgloss.Color("#626262") // Muted text

	// Additional colors for specialized use
	BrightGreen   = lipgloss.Color("#00FF00") // Bright green - selection
	BrightCyan    = lipgloss.Color("#00BFFF") // Bright cyan - info
	BrightRed     = lipgloss.Color("#FF0000") // Bright red - critical errors
	BrightMagenta = lipgloss.Color("#FF00FF") // Bright magenta - special titles
	BrightPink    = lipgloss.Color("#FF6AC1") // Bright pink - alternate accent
	PurpleAccent  = lipgloss.Color("#7D56F4") // Purple accent - labels
	WhiteText     = lipgloss.Color("#FFFFFF") // Pure white - important values
	LightGray     = lipgloss.Color("#AAAAAA") // Light gray - details
	MediumGray    = lipgloss.Color("#CCCCCC") // Medium gray - descriptions
	DarkGray      = lipgloss.Color("#3a3a3a") // Dark gray - inactive borders

	// Background colors
	BgDark        = lipgloss.Color("#282A36") // Dark bg - main background
	BgAccent      = lipgloss.Color("#44475A") // Accent bg - highlighted areas
	BgBlack       = lipgloss.Color("#1a1a1a") // Black bg - high contrast
)

// Common styles with anime aesthetic
//
// These pre-configured styles should be used throughout the CLI for consistency.
// Import this package and use theme.TitleStyle.Render("text") instead of creating
// new lipgloss.NewStyle() instances in individual files.
var (
	// ==================== Title and Header Styles ====================
	// TitleStyle is the main title style for screens and commands
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SakuraPink).
			MarginBottom(1).
			Padding(0, 1)

	// SubtitleStyle is for secondary titles
	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(NeonPurple).
			MarginBottom(1)

	// HeaderStyle is for section headers
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ElectricBlue).
			MarginTop(1)

	// SubHeaderStyle is for subsection headers
	SubHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(NeonPurple)

	// ==================== Text Content Styles ====================
	// PrimaryTextStyle is for main content text
	PrimaryTextStyle = lipgloss.NewStyle().
				Foreground(TextPrimary)

	// SecondaryTextStyle is for supporting text
	SecondaryTextStyle = lipgloss.NewStyle().
				Foreground(TextSecondary)

	// DimTextStyle is for muted/less important text
	DimTextStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			Italic(true)

	// HelpStyle is for help text and instructions
	HelpStyle = lipgloss.NewStyle().
			Foreground(TextHelp).
			MarginTop(1)

	// ==================== Status and Feedback Styles ====================
	// SuccessStyle indicates successful operations
	SuccessStyle = lipgloss.NewStyle().
			Foreground(MintGreen).
			Bold(true)

	// ErrorStyle indicates errors and failures
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ActionRed).
			Bold(true)

	// WarningStyle indicates warnings and cautions
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningYellow).
			Bold(true)

	// InfoStyle is for informational messages
	InfoStyle = lipgloss.NewStyle().
			Foreground(ElectricBlue)

	// ==================== Interactive UI Styles ====================
	// SelectedStyle is for currently selected items in lists
	SelectedStyle = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	// UnselectedStyle is for non-selected items in lists
	UnselectedStyle = lipgloss.NewStyle().
			Foreground(TextDimGray)

	// CursorStyle is for cursor indicators
	CursorStyle = lipgloss.NewStyle().
			Foreground(SakuraPink).
			Bold(true)

	// CheckedStyle is for checked checkbox items
	CheckedStyle = lipgloss.NewStyle().
			Foreground(BrightGreen)

	// UncheckedStyle is for unchecked checkbox items
	UncheckedStyle = lipgloss.NewStyle().
			Foreground(TextDarkGray)

	// ==================== Special Effects ====================
	// GlowStyle creates a glowing/highlighted effect
	GlowStyle = lipgloss.NewStyle().
			Foreground(SakuraPink).
			Bold(true).
			Underline(true)

	// HighlightStyle emphasizes important text
	HighlightStyle = lipgloss.NewStyle().
			Foreground(NeonPurple).
			Bold(true)

	// ==================== Box and Border Styles ====================
	// BoxStyle is a standard rounded box
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(NeonPurple).
			Padding(0, 1)

	// AccentBoxStyle is a more prominent box with double borders
	AccentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(SakuraPink).
			Padding(1, 2)

	// ActivePanelStyle is for active/focused panels
	ActivePanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(BrightPink).
				Padding(1, 2).
				Width(58)

	// InactivePanelStyle is for inactive/unfocused panels
	InactivePanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(DarkGray).
				Padding(1, 2).
				Width(58)

	// DetailBoxStyle is for detailed information displays
	DetailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(NeonPurple).
			Padding(1, 2).
			MarginTop(1)

	// ==================== Progress and Status Indicators ====================
	// ProgressStyle is for in-progress indicators
	ProgressStyle = lipgloss.NewStyle().
			Foreground(BrightGreen)

	// CompleteStyle is for completed items
	CompleteStyle = lipgloss.NewStyle().
			Foreground(BrightCyan)

	// FailedStyle is for failed items
	FailedStyle = lipgloss.NewStyle().
			Foreground(BrightRed)

	// SpinnerStyle is for loading spinners
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(SakuraPink)

	// ==================== Specialized Component Styles ====================
	// LabelStyle is for form labels and field names
	LabelStyle = lipgloss.NewStyle().
			Foreground(PurpleAccent).
			Bold(true)

	// ValueStyle is for displaying values
	ValueStyle = lipgloss.NewStyle().
			Foreground(WhiteText)

	// DescriptionStyle is for descriptions and explanations
	DescriptionStyle = lipgloss.NewStyle().
			Foreground(MediumGray).
			MarginLeft(2).
			MarginTop(1)

	// ExampleStyle is for code examples and syntax
	ExampleStyle = lipgloss.NewStyle().
			Foreground(WarningYellow)

	// PathStyle is for file paths
	PathStyle = lipgloss.NewStyle().
			Foreground(TextDimGray)

	// SummaryStyle is for summary information
	SummaryStyle = lipgloss.NewStyle().
			Foreground(MintGreen).
			Bold(true).
			Padding(1, 2).
			MarginTop(1)

	// ==================== Table Styles ====================
	// TableHeaderStyle is for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ElectricBlue).
				MarginBottom(1)

	// TableRowStyle is for table rows
	TableRowStyle = lipgloss.NewStyle().
			Foreground(TextPrimary)

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

	CategoryVideoGen = lipgloss.NewStyle().
				Bold(true).
				Foreground(LavenderMist).
				MarginTop(1)
)

// CategoryStyle returns a styled category header
// Use this for consistent category rendering across the application
func CategoryStyle(text string) string {
	return SubHeaderStyle.Render(text)
}

// BannerStyle is a reusable banner style for important announcements
var BannerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(SakuraPink).
	Background(BgDark).
	Padding(1, 2).
	Margin(1, 0).
	Border(lipgloss.DoubleBorder()).
	BorderForeground(NeonPurple).
	Align(lipgloss.Center).
	Width(60)

// ProgressBarStyle is for progress bar visualization
var ProgressBarStyle = lipgloss.NewStyle().Foreground(SakuraPink)

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

// RenderBanner renders a stylized title banner
func RenderBanner(text string) string {
	return BannerStyle.Render(text)
}

// RenderProgressBar renders a progress bar with anime aesthetic
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

	return fmt.Sprintf("%s %s %d%%",
		ProgressBarStyle.Render(bar),
		SymbolSakura,
		int(percent*100))
}
