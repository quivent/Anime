package ui

import "fmt"

// ANSI color codes
const (
    Reset      = "\033[0m"
    Bold       = "\033[1m"
    Dim        = "\033[2m"
    Italic     = "\033[3m"
    Underline  = "\033[4m"

    // Colors
    Black      = "\033[30m"
    Red        = "\033[31m"
    Green      = "\033[32m"
    Yellow     = "\033[33m"
    Blue       = "\033[34m"
    Magenta    = "\033[35m"
    Cyan       = "\033[36m"
    White      = "\033[37m"

    // Bright colors
    BrightBlack   = "\033[90m"
    BrightRed     = "\033[91m"
    BrightGreen   = "\033[92m"
    BrightYellow  = "\033[93m"
    BrightBlue    = "\033[94m"
    BrightMagenta = "\033[95m"
    BrightCyan    = "\033[96m"
    BrightWhite   = "\033[97m"

    // Background colors
    BgBlack   = "\033[40m"
    BgRed     = "\033[41m"
    BgGreen   = "\033[42m"
    BgYellow  = "\033[43m"
    BgBlue    = "\033[44m"
    BgMagenta = "\033[45m"
    BgCyan    = "\033[46m"
    BgWhite   = "\033[47m"
)

// Styled text helpers
func Title(s string) string {
    return fmt.Sprintf("%s%s%s%s", Bold, BrightCyan, s, Reset)
}

func Subtitle(s string) string {
    return fmt.Sprintf("%s%s%s", Cyan, s, Reset)
}

func Success(s string) string {
    return fmt.Sprintf("%s%s%s", BrightGreen, s, Reset)
}

func Warning(s string) string {
    return fmt.Sprintf("%s%s%s", BrightYellow, s, Reset)
}

func Error(s string) string {
    return fmt.Sprintf("%s%s%s", BrightRed, s, Reset)
}

func Info(s string) string {
    return fmt.Sprintf("%s%s%s", BrightBlue, s, Reset)
}

func Muted(s string) string {
    return fmt.Sprintf("%s%s%s", BrightBlack, s, Reset)
}

func Highlight(s string) string {
    return fmt.Sprintf("%s%s%s", BrightMagenta, s, Reset)
}

func Value(s string) string {
    return fmt.Sprintf("%s%s%s", BrightWhite, s, Reset)
}

func Key(s string) string {
    return fmt.Sprintf("%s%s%s", Yellow, s, Reset)
}

// Box drawing characters
const (
    BoxHoriz     = "─"
    BoxVert      = "│"
    BoxTopLeft   = "┌"
    BoxTopRight  = "┐"
    BoxBotLeft   = "└"
    BoxBotRight  = "┘"
    BoxVertRight = "├"
    BoxVertLeft  = "┤"
    BoxHorizDown = "┬"
    BoxHorizUp   = "┴"
    BoxCross     = "┼"

    // Double line
    DBoxHoriz    = "═"
    DBoxVert     = "║"
    DBoxTopLeft  = "╔"
    DBoxTopRight = "╗"
    DBoxBotLeft  = "╚"
    DBoxBotRight = "╝"

    // Arrows
    ArrowRight   = "→"
    ArrowLeft    = "←"
    ArrowUp      = "↑"
    ArrowDown    = "↓"
    ArrowBidir   = "↔"

    // Symbols
    Checkmark    = "✓"
    Cross        = "✗"
    Bullet       = "•"
    Diamond      = "◆"
    Circle       = "○"
    FilledCircle = "●"
    Square       = "■"
    Triangle     = "▶"
)
