package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
)

// ============================================================================
// DURATION FORMATTING
// ============================================================================

func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}

// ============================================================================
// OUTPUT HELPERS - Consistent formatting across all commands
// ============================================================================

// PrintSuccess prints a success message with checkmark
func PrintSuccess(msg string) {
	fmt.Println(theme.SuccessStyle.Render("✓ " + msg))
}

// PrintError prints an error message with X mark
func PrintError(msg string) {
	fmt.Println(theme.ErrorStyle.Render("✗ " + msg))
}

// PrintWarning prints a warning message with warning symbol
func PrintWarning(msg string) {
	fmt.Println(theme.WarningStyle.Render("⚠ " + msg))
}

// PrintInfo prints an info message
func PrintInfo(msg string) {
	fmt.Println(theme.InfoStyle.Render("ℹ " + msg))
}

// PrintStep prints a step indicator like [1/5] Message
func PrintStep(current, total int, msg string) {
	fmt.Printf("%s %s\n",
		theme.DimTextStyle.Render(fmt.Sprintf("[%d/%d]", current, total)),
		msg)
}

// PrintHeader prints a section header
func PrintHeader(title string) {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(title))
}

// PrintKeyValue prints a key-value pair with consistent formatting
func PrintKeyValue(key, value string) {
	fmt.Printf("  %s %s\n",
		theme.DimTextStyle.Render(key+":"),
		theme.HighlightStyle.Render(value))
}

// PrintBullet prints a bulleted item
func PrintBullet(msg string) {
	fmt.Printf("  • %s\n", msg)
}

// PrintIndent prints an indented message
func PrintIndent(msg string) {
	fmt.Printf("    %s\n", msg)
}

// PrintDivider prints a horizontal divider line
func PrintDivider() {
	fmt.Println(theme.DimTextStyle.Render(strings.Repeat("─", 60)))
}

// PrintNewline prints an empty line for spacing
func PrintNewline() {
	fmt.Println()
}

// ============================================================================
// ERROR HELPERS - Actionable error messages
// ============================================================================

// PrintErrorWithHint prints an error with a suggested action
func PrintErrorWithHint(err string, hint string) {
	PrintError(err)
	if hint != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Try:"), hint)
	}
}

// PrintErrorWithContext prints an error with context information
func PrintErrorWithContext(operation, target string, err error) {
	PrintError(fmt.Sprintf("%s failed for %s", operation, target))
	if err != nil {
		PrintIndent(theme.DimTextStyle.Render(err.Error()))
	}
}

// ============================================================================
// CONFIRMATION HELPERS
// ============================================================================

// ConfirmAction asks user for confirmation, returns true if confirmed
func ConfirmAction(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// ============================================================================
// DRY RUN HELPERS
// ============================================================================

// IsDryRun checks if dry-run mode is enabled via flag or env
func IsDryRun() bool {
	// Check environment variable
	if os.Getenv("ANIME_DRY_RUN") == "1" || os.Getenv("ANIME_DRY_RUN") == "true" {
		return true
	}
	return false
}

// PrintDryRun prints a message indicating this is a dry run
func PrintDryRun(action string) {
	fmt.Printf("%s %s\n",
		theme.WarningStyle.Render("[DRY-RUN]"),
		action)
}

// ============================================================================
// PROGRESS HELPERS
// ============================================================================

// PrintProgress prints a progress indicator
func PrintProgress(current, total int, item string) {
	percentage := float64(current) / float64(total) * 100
	fmt.Printf("\r  [%3.0f%%] %s", percentage, item)
	if current == total {
		fmt.Println()
	}
}

// PrintSpinner characters for animated spinner (use with goroutine)
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
