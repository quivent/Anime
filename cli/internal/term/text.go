package term

import (
	"fmt"
	"strings"
)

// ruleWidth is the canonical hairline width.
const ruleWidth = 56

// hairline returns a dim divider of n columns.
func hairline(n int) string { return Dim(strings.Repeat("─", n)) }

// Section prints a dim hairline rule, then "» Title" (gold », bold cyan title).
func Section(title string) {
	fmt.Printf("\n%s\n", hairline(ruleWidth))
	fmt.Printf("%s %s\n", Gold.S("»"), Bold(Cyan.S(title)))
}

// Rule prints a single dim hairline divider (~56 cols).
func Rule() { fmt.Println(hairline(ruleWidth)) }

// KV prints an aligned "  label │ value" row (dim label, cyan bar).
func KV(label, value string) { KVw(label, value, 13) }

// KVw is KV with an explicit label width.
func KVw(label, value string, w int) {
	pad := w - visibleLen(label)
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("  %s%s %s %s\n", Dim(label), strings.Repeat(" ", pad), Cyan.S("│"), value)
}

// Ok prints "  ✓ s" with a jade check.
func Ok(s string) { fmt.Printf("  %s %s\n", Jade.S("✓"), s) }

// Fail prints "  ✗ s" with a loss cross.
func Fail(s string) { fmt.Printf("  %s %s\n", Loss.S("✗"), s) }

// Warn prints "  ! s" with a gold bang.
func Warn(s string) { fmt.Printf("  %s %s\n", Gold.S("!"), s) }

// Info prints "  ⬢ s" with a gold hex bullet.
func Info(s string) { fmt.Printf("  %s %s\n", Gold.S("⬢"), s) }

// Step prints a "[n/total] → action…" progress line (gold counter, cyan arrow).
func Step(n, total int, action string) {
	counter := Gold.Seq() + "[" + Bold(fmt.Sprintf("%d", n)) + Gold.Seq() + fmt.Sprintf("/%d]", total) + reset()
	fmt.Printf("%s %s %s…\n", counter, Cyan.S("→"), action)
}

// reset returns Reset only when styling is on (keeps disabled output clean).
func reset() string {
	if Enabled() {
		return Reset
	}
	return ""
}

// Badge returns a compact pill chip "▏text▕" colored c.
func Badge(text string, c Color) string {
	return c.S("▏" + text + "▕")
}

// Hex returns the gold "⬢" bullet glyph (already colored).
func Hex() string { return Gold.S("⬢") }
