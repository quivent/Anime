package cmd

import (
	"fmt"
	"os"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	homeRaw    bool
	homeShell  bool
	homeVerify bool
)

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Navigate to the anime source directory",
	Long: `Navigate to the source directory where anime was built.

Since a subprocess cannot change the parent shell's directory, you need to use this command
with a shell function or wrapper.

Add this to your ~/.zshrc or ~/.bashrc:

    # Function to cd to anime source directory
    anime-home() {
        local dir=$(anime home --raw)
        if [ -n "$dir" ] && [ -d "$dir" ]; then
            cd "$dir"
        else
            echo "anime: source directory not found or not set"
            return 1
        fi
    }

Then use: anime-home

Or use inline:
    cd $(anime home --raw)
`,
	RunE: runHome,
}

func runHome(cmd *cobra.Command, args []string) error {
	// Check if BuildDir is set
	if BuildDir == "" {
		if homeRaw {
			return nil // Just return nothing for raw mode
		}

		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("⚠️  Source Directory Not Set"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  The source directory was not embedded during build."))
		fmt.Println(theme.DimTextStyle.Render("  This usually happens when using 'go install' directly."))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Rebuild using:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("make build"))
		fmt.Println()
		return fmt.Errorf("source directory not available")
	}

	// Verify directory exists if requested
	if homeVerify {
		if _, err := os.Stat(BuildDir); os.IsNotExist(err) {
			if homeRaw {
				return nil // Just return nothing for raw mode
			}

			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("⚠️  Source Directory Not Found"))
			fmt.Println()
			fmt.Printf("  %s: %s\n",
				theme.DimTextStyle.Render("Expected location"),
				theme.HighlightStyle.Render(BuildDir))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("  The directory may have been moved or deleted."))
			fmt.Println()
			return fmt.Errorf("source directory does not exist: %s", BuildDir)
		}
	}

	// Raw mode - just print the path for scripting
	if homeRaw {
		fmt.Println(BuildDir)
		return nil
	}

	// Shell mode - print cd command for eval
	if homeShell {
		fmt.Printf("cd %q\n", BuildDir)
		return nil
	}

	// Pretty display mode (default)
	fmt.Println()
	fmt.Println(theme.RenderBanner("🏠 ANIME HOME"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Source Directory:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(BuildDir))
	fmt.Println()

	// Show how to use it
	fmt.Println(theme.GlowStyle.Render("💡 How to Navigate:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Option 1: Direct cd"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render("cd $(anime home --raw)"))
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render("  Option 2: Add shell function to ~/.zshrc or ~/.bashrc"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("    anime-home() {"))
	fmt.Println(theme.DimTextStyle.Render("        local dir=$(anime home --raw)"))
	fmt.Println(theme.DimTextStyle.Render("        if [ -n \"$dir\" ] && [ -d \"$dir\" ]; then"))
	fmt.Println(theme.DimTextStyle.Render("            cd \"$dir\""))
	fmt.Println(theme.DimTextStyle.Render("        else"))
	fmt.Println(theme.DimTextStyle.Render("            echo \"anime: source directory not found\""))
	fmt.Println(theme.DimTextStyle.Render("            return 1"))
	fmt.Println(theme.DimTextStyle.Render("        fi"))
	fmt.Println(theme.DimTextStyle.Render("    }"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("    Then use: anime-home"))
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render("  Option 3: Eval mode (for advanced users)"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render("eval $(anime home --shell)"))
	fmt.Println()

	return nil
}

func init() {
	rootCmd.AddCommand(homeCmd)

	// Add flags
	homeCmd.Flags().BoolVar(&homeRaw, "raw", false, "Output only the path (for scripting)")
	homeCmd.Flags().BoolVar(&homeShell, "shell", false, "Output cd command for eval")
	homeCmd.Flags().BoolVar(&homeVerify, "verify", false, "Verify directory exists before outputting")
}
