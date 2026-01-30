package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/cli"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var buildInstall bool

var buildCmd = &cobra.Command{
	Use:   "build <cli-name>",
	Short: "Build a registered CLI from source",
	Long: `Build a registered CLI from its source.

This is a shortcut for 'anime cli build <name>'.

Supported languages:
  - Go: runs 'go build'
  - Rust: runs 'cargo build --release'
  - Python: creates a wrapper script

Examples:
  anime build seed
  anime build myapp
  anime build --install seed   # Build and install to PATH`,
	Args: cobra.ExactArgs(1),
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().BoolVarP(&buildInstall, "install", "i", false, "Install binary to ~/.local/bin after build")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("BUILD CLI"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	c, exists := manager.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found (use 'anime cli list' to see registered CLIs)", name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Language:"), theme.InfoStyle.Render(c.Language))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Building..."))
	fmt.Println()

	if err := manager.Build(name); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Build complete"))
	fmt.Println()

	// Install if requested
	if buildInstall {
		if err := installCLI(name); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Failed to install: " + err.Error()))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  Installed to ~/.local/bin"))
		}
		fmt.Println()
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Run:"), theme.HighlightStyle.Render("anime cli "+name))
	fmt.Println()

	return nil
}

func installCLI(name string) error {
	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	c, exists := manager.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	if !c.Built || c.BinaryPath == "" {
		return fmt.Errorf("CLI '%s' has not been built", name)
	}

	// Copy to ~/.local/bin
	return cli.InstallToPath(name, c.BinaryPath)
}
