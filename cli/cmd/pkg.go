package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/pkg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// Flags for pkg commands
var (
	pkgDryRun bool
	pkgForce  bool
	pkgServer string
	pkgGlobal bool
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Package manager - publish, install, and manage packages",
	Long: `Package Manager provides package publishing and dependency management.

Packages are stored at alice:~/cpm/packages with versioning support.

COMMANDS:
  init        Create a new cpm.json file
  publish     Publish package to registry
  republish   Update published version in place
  install     Install a package
  uninstall   Remove installed package
  search      Search for packages
  info        Show package information
  versions    List available versions
  update      Update installed packages
  list        List installed packages

PACKAGE FILE (cpm.json):
  {
    "name": "mypackage",
    "version": "1.0.0",
    "description": "My awesome package",
    "author": "Your Name",
    "license": "MIT"
  }

EXAMPLES:
  anime pkg init                       # Create cpm.json
  anime pkg publish                    # Publish current directory
  anime pkg install mypackage          # Install latest version
  anime pkg install mypackage@1.0.0    # Install specific version
  anime pkg install -g mypackage       # Install globally
  anime pkg search utils               # Search packages

FLAGS:
  --server, -s   Override default server
  --dry-run, -n  Preview without changes
  --global, -g   Use global packages
  --force, -f    Force overwrite`,
	Run: showPkgDashboard,
}

func showPkgDashboard(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("PACKAGE MANAGER"))
	fmt.Println()

	// Check for cpm.json
	pkgInfo, err := pkg.LoadPackageFile()
	if err == nil {
		fmt.Printf("  %s %s@%s\n",
			theme.DimTextStyle.Render("Package:"),
			theme.HighlightStyle.Render(pkgInfo.Name),
			theme.InfoStyle.Render(pkgInfo.Version))
		if pkgInfo.Description != "" {
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Description:"), pkgInfo.Description)
		}
		fmt.Println()
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No cpm.json found (use 'anime pkg init')"))
		fmt.Println()
	}

	// Show quick actions
	fmt.Println(theme.InfoStyle.Render("Quick Actions:"))
	fmt.Println()

	actions := []struct {
		cmd  string
		desc string
	}{
		{"anime pkg init", "Create new cpm.json"},
		{"anime pkg publish", "Publish to registry"},
		{"anime pkg install <pkg>", "Install a package"},
		{"anime pkg search <query>", "Search packages"},
		{"anime pkg list", "List installed packages"},
	}

	for _, a := range actions {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(a.cmd),
			theme.DimTextStyle.Render("- "+a.desc))
	}
	fmt.Println()

	// Show installed packages count
	localInstalled, _ := pkg.ListInstalled(false)
	globalInstalled, _ := pkg.ListInstalled(true)
	localCount := 0
	globalCount := 0
	if localInstalled != nil {
		localCount = len(localInstalled.Packages)
	}
	if globalInstalled != nil {
		globalCount = len(globalInstalled.Packages)
	}

	if localCount > 0 || globalCount > 0 {
		fmt.Printf("  %s %d local, %d global\n",
			theme.DimTextStyle.Render("Installed:"),
			localCount, globalCount)
		fmt.Println()
	}
}

var pkgInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new cpm.json file",
	Long: `Initialize a new package by creating cpm.json.

If name is not provided, uses the current directory name.

Examples:
  anime pkg init                       # Use directory name
  anime pkg init mypackage             # Specify name`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgInit,
}

var pkgPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish package to registry",
	Long: `Publish the current directory as a versioned package.

Requires a cpm.json file with name and version.

Examples:
  anime pkg publish                    # Publish current version
  anime pkg publish -n                 # Dry run`,
	RunE: runPkgPublish,
}

var pkgRepublishCmd = &cobra.Command{
	Use:   "republish",
	Short: "Update published version in place",
	Long: `Re-publish the current version, replacing existing files.

Useful for fixing issues without bumping version.`,
	RunE: runPkgRepublish,
}

var pkgInstallCmd = &cobra.Command{
	Use:   "install <package[@version]>",
	Short: "Install a package",
	Long: `Install a package from the registry.

Packages go to ./cpm_modules by default, or ~/.cpm/packages with --global.

Examples:
  anime pkg install mypackage          # Install latest
  anime pkg install mypackage@1.0.0    # Install specific version
  anime pkg install -g mypackage       # Install globally
  anime pkg install -f mypackage       # Force reinstall`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgInstall,
}

var pkgUninstallCmd = &cobra.Command{
	Use:   "uninstall <package>",
	Short: "Remove installed package",
	Args:  cobra.ExactArgs(1),
	RunE:  runPkgUninstall,
}

var pkgSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for packages",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPkgSearch,
}

var pkgInfoCmd = &cobra.Command{
	Use:   "info <package[@version]>",
	Short: "Show package information",
	Args:  cobra.ExactArgs(1),
	RunE:  runPkgInfo,
}

var pkgVersionsCmd = &cobra.Command{
	Use:   "versions <package>",
	Short: "List available versions",
	Args:  cobra.ExactArgs(1),
	RunE:  runPkgVersions,
}

var pkgUpdateCmd = &cobra.Command{
	Use:   "update [package]",
	Short: "Update installed packages",
	Long: `Update installed packages to latest versions.

If no package specified, updates all.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgUpdate,
}

var pkgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	RunE:  runPkgList,
}

func init() {
	// Global flags
	pkgCmd.PersistentFlags().StringVarP(&pkgServer, "server", "s", "", "Override default server")
	pkgCmd.PersistentFlags().BoolVarP(&pkgDryRun, "dry-run", "n", false, "Preview without changes")

	// Install/uninstall flags
	pkgInstallCmd.Flags().BoolVarP(&pkgGlobal, "global", "g", false, "Install globally")
	pkgInstallCmd.Flags().BoolVarP(&pkgForce, "force", "f", false, "Force overwrite")
	pkgUninstallCmd.Flags().BoolVarP(&pkgGlobal, "global", "g", false, "Uninstall global package")
	pkgUpdateCmd.Flags().BoolVarP(&pkgGlobal, "global", "g", false, "Update global packages")
	pkgListCmd.Flags().BoolVarP(&pkgGlobal, "global", "g", false, "List global packages")

	// Publish flags
	pkgPublishCmd.Flags().BoolVarP(&pkgForce, "force", "f", false, "Force overwrite existing version")

	// Add subcommands
	pkgCmd.AddCommand(pkgInitCmd)
	pkgCmd.AddCommand(pkgPublishCmd)
	pkgCmd.AddCommand(pkgRepublishCmd)
	pkgCmd.AddCommand(pkgInstallCmd)
	pkgCmd.AddCommand(pkgUninstallCmd)
	pkgCmd.AddCommand(pkgSearchCmd)
	pkgCmd.AddCommand(pkgInfoCmd)
	pkgCmd.AddCommand(pkgVersionsCmd)
	pkgCmd.AddCommand(pkgUpdateCmd)
	pkgCmd.AddCommand(pkgListCmd)

	rootCmd.AddCommand(pkgCmd)
}

func getPkgConfig() (*pkg.Config, error) {
	keyPath, cleanup, err := writeEmbeddedKeyToTemp()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SSH key: %w", err)
	}

	return &pkg.Config{
		Server:  getPkgServer(),
		DryRun:  pkgDryRun,
		Force:   pkgForce,
		Global:  pkgGlobal,
		KeyPath: keyPath,
		Cleanup: cleanup,
	}, nil
}

func getPkgServer() string {
	if pkgServer != "" {
		return pkgServer
	}
	return pkg.DefaultServer
}

func getPkgTarget() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return resolveSSHTarget(cfg, getPkgServer())
}

func runPkgInit(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("INIT"))
	fmt.Println()

	pkgInfo, err := pkg.InitPackage(name, "", "")
	if err != nil {
		return err
	}

	fmt.Printf("  Created %s\n", theme.HighlightStyle.Render("cpm.json"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), pkgInfo.Name)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), pkgInfo.Version)
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Edit cpm.json to add description, author, etc."))
	fmt.Println()

	return nil
}

func runPkgPublish(cmd *cobra.Command, args []string) error {
	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	if pkgDryRun {
		fmt.Println(theme.WarningStyle.Render("PUBLISH (DRY RUN)"))
	} else {
		fmt.Println(theme.InfoStyle.Render("PUBLISH"))
	}
	fmt.Println()

	pkgInfo, err := pkg.Publish(target, cfg)
	if err != nil {
		return err
	}

	fmt.Println()
	if pkgDryRun {
		fmt.Println(theme.WarningStyle.Render("  Dry run complete"))
	} else {
		fmt.Printf("  %s %s@%s\n",
			theme.SuccessStyle.Render("Published"),
			theme.HighlightStyle.Render(pkgInfo.Name),
			theme.InfoStyle.Render(pkgInfo.Version))
	}
	fmt.Println()

	return nil
}

func runPkgRepublish(cmd *cobra.Command, args []string) error {
	// Republish is publish with force
	pkgForce = true
	return runPkgPublish(cmd, args)
}

func runPkgInstall(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]

	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	pkgName, pkgVersion := pkg.ParsePackageSpec(pkgSpec)
	if pkgVersion == "" {
		pkgVersion = "latest"
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("INSTALL"))
	fmt.Println()
	fmt.Printf("  %s %s@%s\n",
		theme.DimTextStyle.Render("Package:"),
		theme.HighlightStyle.Render(pkgName),
		theme.InfoStyle.Render(pkgVersion))

	if pkgGlobal {
		fmt.Printf("  %s global\n", theme.DimTextStyle.Render("Scope:"))
	}
	fmt.Println()

	installed, err := pkg.Install(target, pkgSpec, cfg)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  %s %s@%s\n",
		theme.SuccessStyle.Render("Installed"),
		theme.HighlightStyle.Render(pkgName),
		theme.InfoStyle.Render(installed.Version))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Location:"), installed.Path)
	fmt.Println()

	return nil
}

func runPkgUninstall(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("UNINSTALL"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkgName))
	fmt.Println()

	if err := pkg.Uninstall(pkgName, cfg); err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render("  Uninstalled"))
	fmt.Println()

	return nil
}

func runPkgSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Printf("  %s \"%s\"\n", theme.DimTextStyle.Render("Searching for:"), theme.InfoStyle.Render(query))
	fmt.Println()

	matches, err := pkg.Search(target, query, cfg)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No packages found"))
	} else {
		fmt.Printf("  %s\n\n", theme.DimTextStyle.Render(fmt.Sprintf("Found %d package(s):", len(matches))))
		for _, p := range matches {
			fmt.Printf("  %s", theme.HighlightStyle.Render(p.Name))
			if p.Version != "" {
				fmt.Printf("@%s", theme.InfoStyle.Render(p.Version))
			}
			if p.Description != "" {
				fmt.Printf(" - %s", theme.DimTextStyle.Render(p.Description))
			}
			fmt.Println()
		}
	}
	fmt.Println()

	return nil
}

func runPkgInfo(cmd *cobra.Command, args []string) error {
	pkgSpec := args[0]

	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("PACKAGE INFO"))
	fmt.Println()

	pkgInfo, err := pkg.GetInfo(target, pkgSpec, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(pkgInfo.Name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Version:"), theme.InfoStyle.Render(pkgInfo.Version))
	if pkgInfo.Description != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Description:"), pkgInfo.Description)
	}
	if pkgInfo.Author != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Author:"), pkgInfo.Author)
	}
	if pkgInfo.License != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("License:"), pkgInfo.License)
	}
	if len(pkgInfo.Keywords) > 0 {
		fmt.Printf("  %s %v\n", theme.DimTextStyle.Render("Keywords:"), pkgInfo.Keywords)
	}
	if pkgInfo.Repository != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repository:"), pkgInfo.Repository)
	}
	fmt.Println()

	return nil
}

func runPkgVersions(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package:"), theme.HighlightStyle.Render(pkgName))
	fmt.Println()

	versions, err := pkg.GetVersions(target, pkgName, cfg)
	if err != nil {
		return err
	}

	fmt.Println(theme.DimTextStyle.Render("  Available versions:"))
	for _, v := range versions {
		if v.IsLatest {
			fmt.Printf("    %s %s\n", theme.HighlightStyle.Render(v.Version), theme.SuccessStyle.Render("(latest)"))
		} else {
			fmt.Printf("    %s\n", v.Version)
		}
	}
	fmt.Println()

	return nil
}

func runPkgUpdate(cmd *cobra.Command, args []string) error {
	target, err := getPkgTarget()
	if err != nil {
		return err
	}

	cfg, err := getPkgConfig()
	if err != nil {
		return err
	}
	defer cfg.Cleanup()

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("UPDATE"))
	fmt.Println()

	installed, err := pkg.ListInstalled(pkgGlobal)
	if err != nil {
		return err
	}

	if len(installed.Packages) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No packages installed"))
		fmt.Println()
		return nil
	}

	// Filter if specific package requested
	packagesToUpdate := installed.Packages
	if len(args) == 1 {
		pkgName := args[0]
		if _, ok := installed.Packages[pkgName]; !ok {
			return fmt.Errorf("package not installed: %s", pkgName)
		}
		packagesToUpdate = map[string]pkg.InstalledPackage{pkgName: installed.Packages[pkgName]}
	}

	updated := 0
	for pkgName, installedPkg := range packagesToUpdate {
		// Get latest version
		pkgInfo, err := pkg.GetInfo(target, pkgName, cfg)
		if err != nil || pkgInfo.Version == "" || pkgInfo.Version == installedPkg.Version {
			fmt.Printf("  %s %s - %s\n",
				theme.DimTextStyle.Render("o"),
				pkgName,
				theme.DimTextStyle.Render("up to date"))
			continue
		}

		fmt.Printf("  %s %s %s -> %s\n",
			theme.InfoStyle.Render("^"),
			theme.HighlightStyle.Render(pkgName),
			theme.DimTextStyle.Render(installedPkg.Version),
			theme.SuccessStyle.Render(pkgInfo.Version))

		// Reinstall
		pkgForce = true
		if _, err := pkg.Install(target, pkgName, cfg); err != nil {
			fmt.Printf("    %s\n", theme.ErrorStyle.Render(err.Error()))
		} else {
			updated++
		}
	}

	fmt.Println()
	if updated > 0 {
		fmt.Printf("  %s %d package(s)\n", theme.SuccessStyle.Render("Updated"), updated)
	} else {
		fmt.Println(theme.SuccessStyle.Render("  All packages up to date"))
	}
	fmt.Println()

	return nil
}

func runPkgList(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("INSTALLED PACKAGES"))
	fmt.Println()

	installed, err := pkg.ListInstalled(pkgGlobal)
	if err != nil {
		return err
	}

	if len(installed.Packages) == 0 {
		scope := "local"
		if pkgGlobal {
			scope = "global"
		}
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("No %s packages installed", scope)))
	} else {
		installPath, _ := pkg.GetInstallPath(pkgGlobal)
		fmt.Printf("  %s %s\n\n", theme.DimTextStyle.Render("Location:"), installPath)

		for name, p := range installed.Packages {
			fmt.Printf("  %s@%s\n",
				theme.HighlightStyle.Render(name),
				theme.InfoStyle.Render(p.Version))
		}
	}
	fmt.Println()

	// Also show other scope count
	otherInstalled, _ := pkg.ListInstalled(!pkgGlobal)
	if otherInstalled != nil && len(otherInstalled.Packages) > 0 {
		scope := "global"
		if pkgGlobal {
			scope = "local"
		}
		otherPath, _ := pkg.GetInstallPath(!pkgGlobal)
		fmt.Printf("  %s %d %s packages in %s\n",
			theme.DimTextStyle.Render("Also:"),
			len(otherInstalled.Packages),
			scope,
			filepath.Base(otherPath))
		fmt.Println()
	}

	return nil
}
