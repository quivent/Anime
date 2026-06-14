// Package term is qwentize's terminal styling layer: the Aurum palette
// (molten gold over obsidian) rendered with stdlib-only ANSI, downsampled to
// whatever the terminal can do, and degrading to plain text when piped.
//
// Cheat-sheet (public API):
//
//	Enabled() / SetEnabled(on) / Level()           capability
//	Color{R,G,B}; palette vars Gold..InkFaint       color
//	c.S(s) c.Seq(); Reset; SeqBold/Dim/Italic/Under text seqs
//	Bold/Dim/Italic/Under(s); Gradient(s, stops...)  inline styling
//	Section/Rule/KV/KVw/Ok/Fail/Warn/Info/Step       text primitives
//	Badge(text,c) Hex()                              chips + glyph
//	VisLen(s) Pad(s,w)                               ANSI-aware width + align
//	Node + NewNode/Child/Add/With/Render             honeycomb tree
//	Table + NewTable/Row/Right/Render                aligned table
package term

import (
	"fmt"
	"os"
	"strings"
)

// Reset is the SGR reset; callers concat it freely (it stays a no-op visually
// even when styling is off — see note: when disabled, Seq()/SeqBold()/... give
// "" so the surrounding Reset is harmless).
const Reset = "\033[0m"

// Color carries truecolor RGB; rendering downsamples to the detected Level.
type Color struct{ R, G, B uint8 }

// Aurum palette — truecolor anchors shared with the web design system.
var (
	Gold       = Color{0xD9, 0xB4, 0x5A}
	GoldBright = Color{0xF6, 0xDF, 0x9A}
	GoldDeep   = Color{0xA6, 0x80, 0x2F}
	Ember      = Color{0x6E, 0x53, 0x1A}

	Cyan     = Color{0x41, 0xE0, 0xD0}
	CyanDeep = Color{0x1F, 0x9F, 0xB0}

	Jade     = Color{0x4A, 0xDE, 0x80}
	JadeDeep = Color{0x1F, 0xA4, 0x63}

	Loss     = Color{0xFF, 0x5C, 0x5C}
	LossDeep = Color{0xB2, 0x30, 0x30}

	Ink       = Color{0xD8, 0xD5, 0xCC}
	InkBright = Color{0xF5, 0xF3, 0xEC}
	InkMuted  = Color{0x9A, 0x95, 0x8B}
	InkFaint  = Color{0x63, 0x5F, 0x58}
)

var (
	enabled bool // ANSI styling on?
	level   int  // 0 none, 1 16-color, 2 256-color, 3 truecolor
)

func init() { detect() }

// detect probes stdout once: TTY + NO_COLOR for enabled, COLORTERM/TERM for level.
func detect() {
	enabled = isTTY() && os.Getenv("NO_COLOR") == ""
	if !enabled {
		level = 0
		return
	}
	switch {
	case hasAny(os.Getenv("COLORTERM"), "truecolor", "24bit"):
		level = 3
	case strings.Contains(os.Getenv("TERM"), "256color"):
		level = 2
	default:
		level = 1
	}
}

// isTTY reports whether stdout is a character device (a terminal).
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func hasAny(s string, subs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// Enabled reports whether ANSI styling should be emitted.
func Enabled() bool { return enabled }

// SetEnabled overrides detection (e.g. --no-color forcing off, or forcing on).
func SetEnabled(on bool) {
	enabled = on
	if !on {
		level = 0
		return
	}
	if level == 0 {
		// re-derive a sensible level for the forced-on case
		switch {
		case hasAny(os.Getenv("COLORTERM"), "truecolor", "24bit"):
			level = 3
		case strings.Contains(os.Getenv("TERM"), "256color"):
			level = 2
		default:
			level = 1
		}
	}
}

// Level: 0 none, 1 16-color, 2 256-color, 3 truecolor.
func Level() int {
	if !enabled {
		return 0
	}
	return level
}

// Seq is the raw SGR foreground-open sequence for c ("" if disabled) — for
// interop with existing fmt.Printf code that concatenates color + Reset.
func (c Color) Seq() string {
	switch Level() {
	case 3:
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
	case 2:
		return fmt.Sprintf("\033[38;5;%dm", c.xterm256())
	case 1:
		return fmt.Sprintf("\033[%dm", c.basic())
	default:
		return ""
	}
}

// S paints s in c then resets; returns s unchanged when styling is off.
func (c Color) S(s string) string {
	seq := c.Seq()
	if seq == "" {
		return s
	}
	return seq + s + Reset
}

// xterm256 maps RGB to the nearest xterm-256 palette index.
func (c Color) xterm256() int {
	// grayscale ramp (232-255) when channels are close together
	if c.R == c.G && c.G == c.B {
		if c.R < 8 {
			return 16
		}
		if c.R > 248 {
			return 231
		}
		return 232 + int((int(c.R)-8)*24/247)
	}
	q := func(v uint8) int {
		switch {
		case v < 48:
			return 0
		case v < 115:
			return 1
		default:
			return int((v - 35) / 40)
		}
	}
	return 16 + 36*q(c.R) + 6*q(c.G) + q(c.B)
}

// basic maps a palette Color to a sensible basic 16-color SGR foreground.
func (c Color) basic() int {
	switch c {
	case Gold, GoldBright, GoldDeep, Ember:
		return 33 // yellow
	case Cyan, CyanDeep:
		return 36
	case Jade, JadeDeep:
		return 32
	case Loss, LossDeep:
		return 31
	case InkBright:
		return 97 // bright white
	case Ink:
		return 37
	case InkMuted, InkFaint:
		return 90 // bright black / grey
	default:
		return 39 // default fg
	}
}

// style-attribute sequences (empty when disabled).
func seqAttr(code string) string {
	if !Enabled() {
		return ""
	}
	return "\033[" + code + "m"
}

func SeqBold() string   { return seqAttr("1") }
func SeqDim() string    { return seqAttr("2") }
func SeqItalic() string { return seqAttr("3") }
func SeqUnder() string  { return seqAttr("4") }

func attr(code, s string) string {
	if !Enabled() {
		return s
	}
	return "\033[" + code + "m" + s + Reset
}

func Bold(s string) string   { return attr("1", s) }
func Dim(s string) string    { return attr("2", s) }
func Italic(s string) string { return attr("3", s) }
func Under(s string) string  { return attr("4", s) }

// Gradient paints s char-by-char across a ramp (default GoldBright->Gold->GoldDeep
// when no stops given). Below truecolor it falls back to a flat Gold paint.
func Gradient(s string, stops ...Color) string {
	if len(stops) == 0 {
		stops = []Color{GoldBright, Gold, GoldDeep}
	}
	if Level() < 3 {
		return stops[0].S(s)
	}
	rs := []rune(s)
	n := len(rs)
	if n == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range rs {
		// leave whitespace uncolored so trailing resets stay clean
		if r == ' ' || r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		b.WriteString(rampAt(stops, t).Seq())
		b.WriteRune(r)
	}
	b.WriteString(Reset)
	return b.String()
}

// rampAt samples a multi-stop color ramp at t in [0,1].
func rampAt(stops []Color, t float64) Color {
	if len(stops) == 1 || t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	seg := t * float64(len(stops)-1)
	i := int(seg)
	f := seg - float64(i)
	a, c := stops[i], stops[i+1]
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*f) }
	return Color{lerp(a.R, c.R), lerp(a.G, c.G), lerp(a.B, c.B)}
}

// visibleLen returns the display width of s ignoring ANSI escape sequences.
func visibleLen(s string) int {
	return len([]rune(stripANSI(s)))
}

// VisLen is the display width of s, ignoring ANSI escape sequences — the one
// width measure the whole CLI shares for column alignment.
func VisLen(s string) int { return visibleLen(s) }

// Pad right-pads s with spaces to w visible columns (ANSI-aware), so colored
// cells align the same as plain ones.
func Pad(s string, w int) string {
	if d := w - visibleLen(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// stripANSI removes CSI ...m escape sequences for width measurement.
func stripANSI(s string) string {
	if !strings.Contains(s, "\033[") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !(s[j] >= '@' && s[j] <= '~') {
				j++
			}
			if j < len(s) {
				j++ // consume final byte
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
