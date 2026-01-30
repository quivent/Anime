package ui

import (
    "fmt"
    "strings"
)

// PrintHeader prints a styled header
func PrintHeader(title string) {
    width := 60
    line := strings.Repeat(BoxHoriz, width)
    fmt.Printf("\n%s%s%s\n", BrightCyan, line, Reset)
    padding := (width - len(title)) / 2
    fmt.Printf("%s%s%s%s%s\n", BrightCyan, BoxVert, Reset, strings.Repeat(" ", padding-1)+Title(title)+strings.Repeat(" ", width-padding-len(title)-1), BrightCyan+BoxVert+Reset)
    fmt.Printf("%s%s%s\n\n", BrightCyan, line, Reset)
}

// PrintSection prints a section header
func PrintSection(title string) {
    fmt.Printf("\n%s %s\n", Cyan+Triangle+Reset, Title(title))
    fmt.Printf("%s%s%s\n", Dim, strings.Repeat(BoxHoriz, 50), Reset)
}

// PrintSubSection prints a subsection header
func PrintSubSection(title string) {
    fmt.Printf("\n  %s %s\n", Yellow+Bullet+Reset, Subtitle(title))
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
    fmt.Printf("    %s: %s\n", Key(key), Value(value))
}

// PrintKeyValueIndent prints an indented key-value pair
func PrintKeyValueIndent(indent int, key, value string) {
    fmt.Printf("%s%s: %s\n", strings.Repeat(" ", indent), Key(key), Value(value))
}

// PrintList prints a bulleted list item
func PrintList(item string) {
    fmt.Printf("    %s %s\n", Muted(Bullet), item)
}

// PrintListIndent prints an indented list item
func PrintListIndent(indent int, item string) {
    fmt.Printf("%s%s %s\n", strings.Repeat(" ", indent), Muted(Bullet), item)
}

// PrintStatus prints a status line with icon
func PrintStatus(status string, message string) {
    var icon, color string
    switch status {
    case "success", "done", "complete", "completed":
        icon = Checkmark
        color = BrightGreen
    case "error", "failed", "fail":
        icon = Cross
        color = BrightRed
    case "warning", "warn":
        icon = "!"
        color = BrightYellow
    case "info":
        icon = "i"
        color = BrightBlue
    case "pending", "waiting":
        icon = Circle
        color = BrightBlack
    case "running", "in_progress":
        icon = FilledCircle
        color = BrightCyan
    default:
        icon = Bullet
        color = White
    }
    fmt.Printf("  %s%s%s %s\n", color, icon, Reset, message)
}

// PrintProgressBar prints a progress bar
func PrintProgressBar(current, total int, width int) {
    if width <= 0 {
        width = 40
    }
    percent := float64(current) / float64(total)
    filled := int(percent * float64(width))
    empty := width - filled

    bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
    fmt.Printf("  [%s%s%s] %s%.1f%%%s\n", BrightCyan, bar, Reset, BrightWhite, percent*100, Reset)
}

// PrintProgressBarInline prints a progress bar on a single line (for updates)
func PrintProgressBarInline(label string, current, total int, width int) {
    if width <= 0 {
        width = 30
    }
    percent := float64(current) / float64(total)
    filled := int(percent * float64(width))
    empty := width - filled

    bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
    fmt.Printf("\r  %s [%s%s%s] %s%.0f%%%s  ", label, BrightCyan, bar, Reset, BrightWhite, percent*100, Reset)
}

// PrintSpinner prints a spinner character for the given frame
func PrintSpinner(frame int, message string) {
    spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    spinner := spinners[frame%len(spinners)]
    fmt.Printf("\r  %s%s%s %s", BrightCyan, spinner, Reset, message)
}

// ClearLine clears the current line
func ClearLine() {
    fmt.Print("\r\033[K")
}

// PrintTable prints a simple table
func PrintTable(headers []string, rows [][]string) {
    // Calculate column widths
    widths := make([]int, len(headers))
    for i, h := range headers {
        widths[i] = len(h)
    }
    for _, row := range rows {
        for i, cell := range row {
            if i < len(widths) && len(cell) > widths[i] {
                widths[i] = len(cell)
            }
        }
    }

    // Print header
    fmt.Print("  ")
    for i, h := range headers {
        fmt.Printf("%s%-*s%s  ", Bold+Cyan, widths[i], h, Reset)
    }
    fmt.Println()

    // Print separator
    fmt.Print("  ")
    for _, w := range widths {
        fmt.Print(strings.Repeat(BoxHoriz, w) + "  ")
    }
    fmt.Println()

    // Print rows
    for _, row := range rows {
        fmt.Print("  ")
        for i, cell := range row {
            if i < len(widths) {
                fmt.Printf("%-*s  ", widths[i], cell)
            }
        }
        fmt.Println()
    }
}

// PrintBox prints text in a box
func PrintBox(title string, content []string) {
    maxLen := len(title)
    for _, line := range content {
        if len(line) > maxLen {
            maxLen = len(line)
        }
    }
    width := maxLen + 4

    // Top border
    fmt.Printf("  %s%s%s%s%s\n", Cyan, BoxTopLeft, strings.Repeat(BoxHoriz, width), BoxTopRight, Reset)

    // Title
    if title != "" {
        padding := (width - len(title)) / 2
        fmt.Printf("  %s%s%s%s%s%s%s\n", Cyan, BoxVert, Reset, strings.Repeat(" ", padding)+Bold+title+Reset+strings.Repeat(" ", width-padding-len(title)), Cyan, BoxVert, Reset)
        fmt.Printf("  %s%s%s%s%s\n", Cyan, BoxVertRight, strings.Repeat(BoxHoriz, width), BoxVertLeft, Reset)
    }

    // Content
    for _, line := range content {
        fmt.Printf("  %s%s%s  %-*s  %s%s%s\n", Cyan, BoxVert, Reset, maxLen, line, Cyan, BoxVert, Reset)
    }

    // Bottom border
    fmt.Printf("  %s%s%s%s%s\n", Cyan, BoxBotLeft, strings.Repeat(BoxHoriz, width), BoxBotRight, Reset)
}

// PrintSuggestion prints a suggestion instead of an error
func PrintSuggestion(context string, suggestions []string) {
    fmt.Printf("\n  %s%s Unable to proceed: %s%s\n", BrightYellow, "⚡", context, Reset)
    fmt.Printf("  %s\n", Muted(strings.Repeat(BoxHoriz, 40)))
    fmt.Printf("  %s Suggestions:%s\n", Bold, Reset)
    for i, s := range suggestions {
        fmt.Printf("    %s%d.%s %s\n", BrightCyan, i+1, Reset, s)
    }
    fmt.Println()
}

// PrintDiagram prints an ASCII diagram with colors
func PrintDiagram(lines []string) {
    for _, line := range lines {
        // Color box characters
        colored := line
        colored = strings.ReplaceAll(colored, "┌", Cyan+"┌"+Reset)
        colored = strings.ReplaceAll(colored, "┐", Cyan+"┐"+Reset)
        colored = strings.ReplaceAll(colored, "└", Cyan+"└"+Reset)
        colored = strings.ReplaceAll(colored, "┘", Cyan+"┘"+Reset)
        colored = strings.ReplaceAll(colored, "│", Cyan+"│"+Reset)
        colored = strings.ReplaceAll(colored, "─", Cyan+"─"+Reset)
        colored = strings.ReplaceAll(colored, "├", Cyan+"├"+Reset)
        colored = strings.ReplaceAll(colored, "┤", Cyan+"┤"+Reset)
        colored = strings.ReplaceAll(colored, "┬", Cyan+"┬"+Reset)
        colored = strings.ReplaceAll(colored, "┴", Cyan+"┴"+Reset)
        colored = strings.ReplaceAll(colored, "◄", BrightGreen+"◄"+Reset)
        colored = strings.ReplaceAll(colored, "►", BrightGreen+"►"+Reset)
        colored = strings.ReplaceAll(colored, "▼", BrightGreen+"▼"+Reset)
        colored = strings.ReplaceAll(colored, "▲", BrightGreen+"▲"+Reset)
        fmt.Println(colored)
    }
}
