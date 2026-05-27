// Package theme is the anime CLI design layer — Aurum palette (molten gold over
// obsidian) rendered via the stdlib-only term package. No Charmbracelet, no
// external dependencies.
//
// Backward-compatible surface: all existing theme.XxxStyle.Render(s) call sites
// continue to work unchanged; the underlying rendering now uses term.Color.
package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	t "github.com/joshkornreich/anime/internal/term"
)

// ─── Aurum lipgloss.Color aliases ────────────────────────────────────────────
// For files that still call lipgloss.NewStyle().Foreground(theme.XxxColor).
// Old anime names are remapped to the nearest Aurum palette entry.
// These can be removed once tui/vllm.go and other callers are rewritten.

var (
	// Gold tier (was pink/purple accent)
	SakuraPink    = lipgloss.Color("#D9B45A") // → Gold
	NeonPurple    = lipgloss.Color("#D9B45A") // → Gold
	SunsetOrange  = lipgloss.Color("#A6802F") // → GoldDeep
	LavenderMist  = lipgloss.Color("#F6DF9A") // → GoldBright
	WarningYellow = lipgloss.Color("#D9B45A") // → Gold

	// Structural (was cyan/blue)
	ElectricBlue = lipgloss.Color("#41E0D0") // → Cyan

	// Success/growth (was green)
	MintGreen = lipgloss.Color("#4ADE80") // → Jade

	// Error (was red)
	ActionRed = lipgloss.Color("#FF5C5C") // → Loss

	// Ink (was grey)
	TextPrimary   = lipgloss.Color("#D8D5CC") // → Ink
	TextSecondary = lipgloss.Color("#9A958B") // → InkMuted
	TextDim       = lipgloss.Color("#635F58") // → InkFaint

	// Backgrounds (kept; only used for Foreground calls so fine as stubs)
	BgDark    = lipgloss.Color("#0A0B10") // → bg-abyss
	BgAccent  = lipgloss.Color("#171A22") // → bg-raised
)

// ─── Style adapter ───────────────────────────────────────────────────────────
// Style wraps a render function so existing .Render(s) call sites work unchanged.

type Style struct{ fn func(string) string }

// Render applies styling to the concatenation of strs (variadic to match the
// lipgloss.Style.Render signature used in existing call sites).
func (s Style) Render(strs ...string) string { return s.fn(strings.Join(strs, "")) }

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	// SuccessStyle: jade — accepted, running, online.
	SuccessStyle = Style{func(s string) string { return t.Bold(t.Jade.S(s)) }}

	// ErrorStyle: crimson — rejected, dead, offline.
	ErrorStyle = Style{func(s string) string { return t.Bold(t.Loss.S(s)) }}

	// WarningStyle: gold — caution, admin badge, partial.
	WarningStyle = Style{func(s string) string { return t.Bold(t.Gold.S(s)) }}

	// InfoStyle: cyan — structure, links, status lines.
	InfoStyle = Style{func(s string) string { return t.Cyan.S(s) }}

	// HighlightStyle: bright gold bold — names, IDs, emphasis.
	HighlightStyle = Style{func(s string) string { return t.Bold(t.Gold.S(s)) }}

	// DimTextStyle: faint ink — timestamps, secondary values.
	DimTextStyle = Style{func(s string) string { return t.Dim(s) }}

	// GlowStyle: gold gradient — section headings, labels of value.
	GlowStyle = Style{func(s string) string { return t.Gradient(s) }}

	// Semantic aliases
	TitleStyle         = GlowStyle
	HeaderStyle        = InfoStyle
	SubHeaderStyle     = HighlightStyle
	PrimaryTextStyle   = Style{func(s string) string { return t.Ink.S(s) }}
	SecondaryTextStyle = Style{func(s string) string { return t.InkMuted.S(s) }}
)

// ─── Symbols — pre-colored ───────────────────────────────────────────────────
// Used as: fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("..."))

var (
	SymbolSuccess = t.Jade.S("✓")
	SymbolError   = t.Loss.S("✗")
	SymbolWarning = t.Gold.S("!")
	SymbolInfo    = t.Gold.S("⬢")
	SymbolLoading = t.Dim("◌")
	SymbolStar    = t.Gold.S("◆")
	SymbolBolt    = t.Cyan.S("›")
	SymbolShield  = t.Gold.S("◈")

	// Extended symbols (Aurum-mapped from old anime emoji names)
	SymbolBuild      = t.Gold.S("⚒")
	SymbolDeploy     = t.Cyan.S("→")
	SymbolConfig     = t.Gold.S("⚙")
	SymbolSparkle    = t.Gold.S("◆")
	SymbolDownload   = t.Cyan.S("↓")
	SymbolInstall    = t.Jade.S("↑")
	SymbolPackage    = t.Gold.S("◈")
	SymbolModule     = t.Gold.S("◆")
	SymbolComponent  = t.Cyan.S("›")
	SymbolTree       = t.Dim("│")
	SymbolSakura     = t.Gold.S("◆")
	SymbolHeart      = t.Jade.S("♥")
	SymbolSword      = t.Cyan.S("×")
	SymbolMagic      = t.Gold.S("◈")

	// Tree branch glyphs (plain strings — used in fmt output as structural chars)
	SymbolBranch     = "├─"
	SymbolLastBranch = "└─"
	SymbolPipe       = "│"
	SymbolSpace      = "   "
)

// ─── term passthrough ────────────────────────────────────────────────────────
// Call these instead of the old fmt.Printf + SymbolXxx pattern.
// They delegate directly to the term package — same output as qwentize.

func Ok(s string)      { t.Ok(s) }
func Fail(s string)    { t.Fail(s) }
func Warn(s string)    { t.Warn(s) }
func Info(s string)    { t.Info(s) }
func Section(s string) { t.Section(s) }
func Rule()            { t.Rule() }

// KV prints an aligned "  label │ value" row.
func KV(label, value string) { t.KV(label, value) }

// Step prints a "[n/total] → action…" progress line.
func Step(n, total int, action string) { t.Step(n, total, action) }

// NewTable creates an Aurum-styled aligned table.
func NewTable(headers ...string) *t.Table { return t.NewTable(headers...) }

// Gold / Cyan / Jade / Loss / Dim / Bold — inline color helpers for call sites
// that need to colorize a substring rather than print a full line.
var (
	Gold = t.Gold
	Cyan = t.Cyan
	Jade = t.Jade
	Loss = t.Loss
)

func Dim(s string) string  { return t.Dim(s) }
func Bold(s string) string { return t.Bold(s) }

// ─── Category helpers ────────────────────────────────────────────────────────

// CategoryStyle renders a bold gold section heading. Replaces old lipgloss category styles.
func CategoryStyle(text string) string { return t.Bold(t.Gold.S(text)) }

// GetCategoryStyle returns a Style for the given category name.
// All map to Aurum gold; the semantic distinction is handled by color context elsewhere.
func GetCategoryStyle(_ string) Style { return HighlightStyle }

// RenderProgressBar renders a gold/dim progress bar.
func RenderProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)
	empty := width - filled
	bar := strings.Repeat("█", filled) + t.Dim(strings.Repeat("░", empty))
	return t.Gold.S(bar) + t.Dim(fmt.Sprintf(" %d%%", int(percent*100)))
}

// ─── Banner ──────────────────────────────────────────────────────────────────

// RenderBanner renders a gold-gradient double-rule box with a centered title.
// Width is fixed at 62 columns (matches existing CLI output width).
func RenderBanner(text string) string {
	const width = 62
	const inner = width - 2 // excluding left + right ║

	textLen := len([]rune(text))
	left := (inner - textLen) / 2
	right := inner - textLen - left
	if left < 0 {
		left = 0
	}
	if right < 0 {
		right = 0
	}

	sp := func(n int) string { return strings.Repeat(" ", n) }

	top := t.Gradient("╔" + strings.Repeat("═", inner) + "╗")
	blank := t.Gradient("║") + sp(inner) + t.Gradient("║")
	title := t.Gradient("║") + sp(left) + t.Gradient(text) + sp(right) + t.Gradient("║")
	bot := t.Gradient("╚" + strings.Repeat("═", inner) + "╝")

	return "\n" + top + "\n" + blank + "\n" + title + "\n" + blank + "\n" + bot + "\n"
}
