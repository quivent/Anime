package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/cli"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// Flags for CLI commands
var (
	cliDescription string
	cliForce       bool
)

var cliCmd = &cobra.Command{
	Use:   "cli [command]",
	Short: "Manage external CLI tools",
	Long: `CLI Management - register, build, and run external CLI tools.

Pull CLI sources from GitHub, add local CLI projects, or register
prebuilt binaries. Then run them as 'anime cli <name>'.

COMMANDS:
  add         Add a local CLI source directory
  pull        Pull CLI source from GitHub/remote
  register    Register a prebuilt binary
  build       Build a CLI from source
  remove      Remove a registered CLI
  list        List all registered CLIs
  update      Update a CLI from its remote source
  run         Run a registered CLI (or just 'anime cli <name>')

EXAMPLES:
  anime cli pull seed https://github.com/user/seed
  anime cli build seed
  anime cli seed --help                   # Run seed CLI
  anime cli add myapp ./path/to/source
  anime cli register mytool /usr/bin/mytool`,
	Run: func(cmd *cobra.Command, args []string) {
		// If a CLI name is provided, try to run it
		if len(args) > 0 {
			// Check if it's a registered CLI
			manager, err := cli.NewManager()
			if err != nil {
				fmt.Println(theme.ErrorStyle.Render("Failed to load CLI registry: " + err.Error()))
				os.Exit(1)
			}

			if manager.Registry.Exists(args[0]) {
				if err := manager.Execute(args[0], args[1:]); err != nil {
					os.Exit(1)
				}
				return
			}
		}

		// Show dashboard
		showCLIDashboard(cmd, args)
	},
	DisableFlagParsing: true,
}

func showCLIDashboard(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CLI MANAGER"))
	fmt.Println()

	// Load registry
	manager, err := cli.NewManager()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Failed to load registry: " + err.Error()))
		fmt.Println()
		return
	}

	clis := manager.Registry.List()
	if len(clis) > 0 {
		fmt.Println(theme.InfoStyle.Render("Registered CLIs:"))
		fmt.Println()

		// Sort by name
		sort.Slice(clis, func(i, j int) bool {
			return clis[i].Name < clis[j].Name
		})

		for _, c := range clis {
			status := theme.DimTextStyle.Render("(not built)")
			if c.Built {
				status = theme.SuccessStyle.Render("(ready)")
			}

			typeStr := ""
			switch c.Type {
			case cli.TypeSource:
				typeStr = theme.DimTextStyle.Render("[source]")
			case cli.TypeBinary:
				typeStr = theme.DimTextStyle.Render("[binary]")
			case cli.TypeRemote:
				typeStr = theme.DimTextStyle.Render("[remote]")
			}

			fmt.Printf("  %s %s %s\n",
				theme.HighlightStyle.Render(c.Name),
				status,
				typeStr)

			if c.Description != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.Description))
			}
		}
		fmt.Println()
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No CLIs registered"))
		fmt.Println()
	}

	// Show quick actions
	fmt.Println(theme.InfoStyle.Render("Quick Actions:"))
	fmt.Println()

	actions := []struct {
		cmd  string
		desc string
	}{
		{"anime cli pull <name> <url>", "Pull CLI source from GitHub"},
		{"anime cli add <name> <path>", "Add local CLI source"},
		{"anime cli register <name> <path>", "Register prebuilt binary"},
		{"anime cli build <name>", "Build CLI from source"},
		{"anime cli <name>", "Run a registered CLI"},
		{"anime cli list", "List all registered CLIs"},
	}

	for _, a := range actions {
		fmt.Printf("  %s\n    %s\n",
			theme.HighlightStyle.Render(a.cmd),
			theme.DimTextStyle.Render(a.desc))
	}
	fmt.Println()
}

var cliAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Add a local CLI source directory",
	Long: `Add a local CLI source directory to the registry.

The source will be copied to ~/.anime/cli/src/<name>.
Use 'anime cli build <name>' to build it.

Supported languages: Go, Rust, Python

Examples:
  anime cli add myapp ./myapp
  anime cli add --desc "My cool app" myapp /path/to/myapp`,
	Args: cobra.ExactArgs(2),
	RunE: runCLIAdd,
}

var cliPullCmd = &cobra.Command{
	Use:   "pull <name> <url>",
	Short: "Pull CLI source from GitHub/remote",
	Long: `Pull CLI source from a Git repository.

The source will be cloned to ~/.anime/cli/src/<name>.
Use 'anime cli build <name>' to build it.

Examples:
  anime cli pull seed https://github.com/user/seed
  anime cli pull --desc "Seed CLI" seed git@github.com:user/seed.git`,
	Args: cobra.ExactArgs(2),
	RunE: runCLIPull,
}

var cliRegisterCmd = &cobra.Command{
	Use:   "register <name> <binary-path>",
	Short: "Register a prebuilt binary",
	Long: `Register a prebuilt binary as a CLI.

The binary will be copied to ~/.anime/cli/bin/<name>.

Examples:
  anime cli register mytool /usr/local/bin/mytool
  anime cli register --desc "My tool" mytool ~/bin/mytool`,
	Args: cobra.ExactArgs(2),
	RunE: runCLIRegister,
}

var cliBuildCmd = &cobra.Command{
	Use:   "build <name>",
	Short: "Build a CLI from source",
	Long: `Build a registered CLI from its source.

Supported languages:
  - Go: runs 'go build'
  - Rust: runs 'cargo build --release'
  - Python: creates a wrapper script

Examples:
  anime cli build seed
  anime cli build myapp`,
	Args: cobra.ExactArgs(1),
	RunE: runCLIBuild,
}

var cliRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered CLI",
	Long: `Remove a CLI from the registry.

This will delete the source and binary files.

Examples:
  anime cli remove seed
  anime cli remove -f myapp   # Force without confirmation`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"rm", "delete"},
	RunE:    runCLIRemove,
}

var cliListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered CLIs",
	Long: `List all registered CLIs with their status.

Shows name, type, language, and build status.`,
	Aliases: []string{"ls"},
	Run:     runCLIList,
}

var cliUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a CLI from its remote source",
	Long: `Update a CLI that was pulled from a remote source.

Runs 'git pull' in the source directory.

Examples:
  anime cli update seed`,
	Args: cobra.ExactArgs(1),
	RunE: runCLIUpdate,
}

var cliRunCmd = &cobra.Command{
	Use:   "run <name> [args...]",
	Short: "Run a registered CLI",
	Long: `Run a registered CLI by name.

You can also run CLIs directly as 'anime cli <name>'.

Examples:
  anime cli run seed --help
  anime cli seed --help       # Shorthand`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runCLIRun,
}

var cliInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a CLI",
	Long: `Show detailed information about a registered CLI.

Examples:
  anime cli info seed`,
	Args: cobra.ExactArgs(1),
	RunE: runCLIInfo,
}

func init() {
	// Add flags
	cliAddCmd.Flags().StringVarP(&cliDescription, "desc", "d", "", "Description for the CLI")
	cliPullCmd.Flags().StringVarP(&cliDescription, "desc", "d", "", "Description for the CLI")
	cliRegisterCmd.Flags().StringVarP(&cliDescription, "desc", "d", "", "Description for the CLI")
	cliRemoveCmd.Flags().BoolVarP(&cliForce, "force", "f", false, "Force removal without confirmation")

	// Add subcommands
	cliCmd.AddCommand(cliAddCmd)
	cliCmd.AddCommand(cliPullCmd)
	cliCmd.AddCommand(cliRegisterCmd)
	cliCmd.AddCommand(cliBuildCmd)
	cliCmd.AddCommand(cliRemoveCmd)
	cliCmd.AddCommand(cliListCmd)
	cliCmd.AddCommand(cliUpdateCmd)
	cliCmd.AddCommand(cliRunCmd)
	cliCmd.AddCommand(cliInfoCmd)

	rootCmd.AddCommand(cliCmd)
}

func runCLIAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	path := args[1]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("ADD CLI SOURCE"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	// Check if already exists
	if manager.Registry.Exists(name) {
		return fmt.Errorf("CLI '%s' already exists (use 'anime cli remove %s' first)", name, name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.InfoStyle.Render(path))
	fmt.Println()

	opts := cli.AddOptions{
		Description: cliDescription,
	}

	c, err := manager.AddFromSource(name, path, opts)
	if err != nil {
		return err
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Language:"), theme.InfoStyle.Render(c.Language))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  CLI added successfully"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Next:"), theme.HighlightStyle.Render("anime cli build "+name))
	fmt.Println()

	return nil
}

func runCLIPull(cmd *cobra.Command, args []string) error {
	name := args[0]
	url := args[1]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("PULL CLI SOURCE"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	// Check if already exists
	if manager.Registry.Exists(name) {
		return fmt.Errorf("CLI '%s' already exists (use 'anime cli remove %s' first)", name, name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("URL:"), theme.InfoStyle.Render(url))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Cloning repository..."))
	fmt.Println()

	opts := cli.AddOptions{
		Description: cliDescription,
	}

	c, err := manager.PullFromRemote(name, url, opts)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Language:"), theme.InfoStyle.Render(c.Language))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  CLI pulled successfully"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Next:"), theme.HighlightStyle.Render("anime cli build "+name))
	fmt.Println()

	return nil
}

func runCLIRegister(cmd *cobra.Command, args []string) error {
	name := args[0]
	binaryPath := args[1]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("REGISTER BINARY"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	// Check if already exists
	if manager.Registry.Exists(name) {
		return fmt.Errorf("CLI '%s' already exists (use 'anime cli remove %s' first)", name, name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Binary:"), theme.InfoStyle.Render(binaryPath))
	fmt.Println()

	opts := cli.AddOptions{
		Description: cliDescription,
	}

	_, err = manager.RegisterBinary(name, binaryPath, opts)
	if err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render("  Binary registered successfully"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Run:"), theme.HighlightStyle.Render("anime cli "+name))
	fmt.Println()

	return nil
}

func runCLIBuild(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("CLI '%s' not found", name)
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
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Run:"), theme.HighlightStyle.Render("anime cli "+name))
	fmt.Println()

	return nil
}

func runCLIRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("REMOVE CLI"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	c, exists := manager.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	if c.SourcePath != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.InfoStyle.Render(c.SourcePath))
	}
	if c.BinaryPath != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Binary:"), theme.InfoStyle.Render(c.BinaryPath))
	}
	fmt.Println()

	// Confirm unless forced
	if !cliForce {
		fmt.Print(theme.WarningStyle.Render("  Type 'yes' to confirm: "))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println(theme.DimTextStyle.Render("  Cancelled"))
			fmt.Println()
			return nil
		}
	}

	if err := manager.Remove(name); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  CLI removed"))
	fmt.Println()

	return nil
}

func runCLIList(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("REGISTERED CLIs"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Failed to load registry: " + err.Error()))
		fmt.Println()
		return
	}

	clis := manager.Registry.List()
	if len(clis) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No CLIs registered"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Add one with:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime cli pull <name> <url>"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime cli add <name> <path>"))
		fmt.Println()
		return
	}

	// Sort by name
	sort.Slice(clis, func(i, j int) bool {
		return clis[i].Name < clis[j].Name
	})

	// Print table header
	fmt.Printf("  %-15s %-10s %-10s %-8s %s\n",
		theme.DimTextStyle.Render("NAME"),
		theme.DimTextStyle.Render("TYPE"),
		theme.DimTextStyle.Render("LANGUAGE"),
		theme.DimTextStyle.Render("STATUS"),
		theme.DimTextStyle.Render("DESCRIPTION"))
	fmt.Println()

	for _, c := range clis {
		status := "not built"
		statusStyle := theme.WarningStyle
		if c.Built {
			status = "ready"
			statusStyle = theme.SuccessStyle
		}

		typeStr := string(c.Type)
		lang := c.Language
		if lang == "" {
			lang = "-"
		}

		desc := c.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		fmt.Printf("  %-15s %-10s %-10s %-8s %s\n",
			theme.HighlightStyle.Render(c.Name),
			typeStr,
			lang,
			statusStyle.Render(status),
			theme.DimTextStyle.Render(desc))
	}
	fmt.Println()
}

func runCLIUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("UPDATE CLI"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	c, exists := manager.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(name))
	if c.RemoteURL != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Remote:"), theme.InfoStyle.Render(c.RemoteURL))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Pulling updates..."))
	fmt.Println()

	if err := manager.Update(name); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Update complete"))
	fmt.Println()

	// Suggest rebuilding
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Rebuild with:"), theme.HighlightStyle.Render("anime cli build "+name))
	fmt.Println()

	return nil
}

func runCLIRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("CLI name required")
	}

	name := args[0]
	cliArgs := args[1:]

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	return manager.Execute(name, cliArgs)
}

func runCLIInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("CLI INFO"))
	fmt.Println()

	manager, err := cli.NewManager()
	if err != nil {
		return err
	}

	c, exists := manager.Registry.Get(name)
	if !exists {
		return fmt.Errorf("CLI '%s' not found", name)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(c.Name))

	if c.Description != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Description:"), c.Description)
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Type:"), string(c.Type))

	if c.Language != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Language:"), c.Language)
	}

	if c.RemoteURL != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Remote URL:"), c.RemoteURL)
	}

	if c.SourcePath != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), c.SourcePath)
	}

	if c.BinaryPath != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Binary:"), c.BinaryPath)
	}

	status := "Not built"
	if c.Built {
		status = "Ready"
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Status:"), status)

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Added:"), c.AddedAt.Format(time.RFC3339))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Updated:"), c.UpdatedAt.Format(time.RFC3339))

	fmt.Println()

	// Show usage hints
	if c.Built {
		fmt.Println(theme.DimTextStyle.Render("  Usage:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime cli "+name+" [args...]"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  Build with:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime cli build "+name))
	}
	fmt.Println()

	return nil
}

// Helper function for handling dynamic CLI names
func handleCLIExecution(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Check for known subcommands
	knownCommands := map[string]bool{
		"add": true, "pull": true, "register": true, "build": true,
		"remove": true, "rm": true, "delete": true, "list": true,
		"ls": true, "update": true, "run": true, "info": true, "help": true,
	}

	firstArg := strings.ToLower(args[0])
	if knownCommands[firstArg] || strings.HasPrefix(firstArg, "-") {
		return false
	}

	// Try to execute as CLI name
	manager, err := cli.NewManager()
	if err != nil {
		return false
	}

	if manager.Registry.Exists(args[0]) {
		if err := manager.Execute(args[0], args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, theme.ErrorStyle.Render("Error: "+err.Error()))
			os.Exit(1)
		}
		return true
	}

	return false
}
