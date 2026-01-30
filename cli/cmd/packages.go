package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	packagesCmd = &cobra.Command{
		Use:   "packages [package-id]",
		Short: "List available installation packages",
		Long:  "Display all available packages with dependencies, descriptions, and estimated installation times. Optionally specify a package ID to show details for just that package.",
		Args:  cobra.MaximumNArgs(1),
		Run:   runPackages,
	}

	packageCmd = &cobra.Command{
		Use:   "package <package-id>",
		Short: "Show details for a specific package",
		Long:  "Display detailed information and installation status for a specific package.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				fmt.Println()
				fmt.Println(theme.ErrorStyle.Render("❌ Package ID required"))
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
				fmt.Println(theme.HighlightStyle.Render("  anime package <package-id>"))
				fmt.Println()
				fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
				fmt.Println(theme.DimTextStyle.Render("  anime package core"))
				fmt.Println(theme.DimTextStyle.Render("  anime package pytorch"))
				fmt.Println(theme.DimTextStyle.Render("  anime package ollama"))
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
				fmt.Println(theme.DimTextStyle.Render("  anime packages        # List all packages"))
				fmt.Println(theme.DimTextStyle.Render("  anime packages status # Check installation status"))
				fmt.Println()
				return fmt.Errorf("package requires a package ID")
			}
			return nil
		},
		Run: runPackage,
	}

	packagesStatusCmd = &cobra.Command{
		Use:   "status [package-id]",
		Short: "Show installation status of packages",
		Long:  "Display packages with their installation status on the lambda server. Optionally specify a package ID to check only that package.",
		Run:   runPackagesStatus,
	}

	packagesModelsCmd = &cobra.Command{
		Use:   "models",
		Short: "List only AI model packages (image, video, enhancement, controlnet)",
		Long:  "Display only AI model packages for image generation, video generation, enhancement, and controlnet adapters.",
		Run:   runPackagesModels,
	}

	showTree bool
)

func init() {
	rootCmd.AddCommand(packagesCmd)
	rootCmd.AddCommand(packageCmd)
	packagesCmd.AddCommand(packagesStatusCmd)
	packagesCmd.AddCommand(packagesModelsCmd)
	packagesCmd.Flags().BoolVarP(&showTree, "tree", "t", false, "Show dependency tree")
}

func runPackages(cmd *cobra.Command, args []string) {
	packages := installer.GetPackages()

	// If package-id specified, show just that package
	if len(args) > 0 {
		showPackageDetails(args[0], packages)
		return
	}

	// Try to check installation status on lambda server
	installedPackages := checkInstalledPackages()

	if showTree {
		displayTree(packages, installedPackages)
		return
	}

	displayList(packages, installedPackages)
}

func runPackagesModels(cmd *cobra.Command, args []string) {
	packages := installer.GetPackages()
	packages = filterModelPackages(packages)

	// Try to check installation status on lambda server
	installedPackages := checkInstalledPackages()

	displayModelsList(packages, installedPackages)
}

// filterModelPackages returns only AI model packages
func filterModelPackages(packages map[string]*installer.Package) map[string]*installer.Package {
	modelCategories := map[string]bool{
		"Image Generation":  true,
		"Video Generation":  true,
		"Image Enhancement": true,
		"Video Enhancement": true,
		"ControlNet":        true,
		"Models":            true,
	}

	filtered := make(map[string]*installer.Package)
	for id, pkg := range packages {
		if modelCategories[pkg.Category] {
			filtered[id] = pkg
		}
	}
	return filtered
}

// displayModelsList displays a compact list of model packages
func displayModelsList(packages map[string]*installer.Package, installed map[string]bool) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 AI MODEL PACKAGES 🎨"))
	fmt.Println()

	// Group by category
	categories := make(map[string][]*installer.Package)
	categoryOrder := []string{"Image Generation", "Video Generation", "Image Enhancement", "Video Enhancement", "ControlNet", "Models"}

	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort within categories
	for cat := range categories {
		sort.Slice(categories[cat], func(i, j int) bool {
			return categories[cat][i].Name < categories[cat][j].Name
		})
	}

	// Category emoji mapping
	categoryEmojis := map[string]string{
		"Image Generation":  "🎨",
		"Video Generation":  "🎬",
		"Image Enhancement": "🔧",
		"Video Enhancement": "📹",
		"ControlNet":        "🎛️",
		"Models":            "⭐",
	}

	totalModels := 0
	installedCount := 0

	for _, category := range categoryOrder {
		pkgs := categories[category]
		if len(pkgs) == 0 {
			continue
		}

		emoji := categoryEmojis[category]
		if emoji == "" {
			emoji = "📦"
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("%s  %s (%d models)", emoji, category, len(pkgs))))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, pkg := range pkgs {
			totalModels++
			status := theme.DimTextStyle.Render("◯")
			if installed[pkg.ID] {
				status = theme.SuccessStyle.Render("✓")
				installedCount++
			}

			fmt.Printf("  %s %s %s\n",
				status,
				theme.HighlightStyle.Render(pkg.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", pkg.Size)))
			fmt.Printf("      %s\n", theme.DimTextStyle.Render("anime install "+pkg.ID))
		}
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("📦 Total: %d models (%d installed)", totalModels, installedCount)))
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
}

func runPackage(cmd *cobra.Command, args []string) {
	packages := installer.GetPackages()
	showPackageDetails(args[0], packages)
}

// checkInstalledPackages checks which packages are installed on lambda server
func checkInstalledPackages() map[string]bool {
	installed := make(map[string]bool)

	// First, try checking locally (if we're running on the server itself)
	if isRunningOnServer() {
		packages := installer.GetPackages()
		for pkgID := range packages {
			if checkPackageInstalledLocal(pkgID) {
				installed[pkgID] = true
			}
		}
		return installed
	}

	// Not running locally, try to connect to remote lambda server
	cfg, err := config.Load()
	if err != nil {
		return installed
	}

	// Get lambda server
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	if lambdaTarget == "" {
		return installed
	}

	// Parse target
	var user, host string
	if strings.Contains(lambdaTarget, "@") {
		parts := strings.SplitN(lambdaTarget, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = lambdaTarget
	}

	// Connect and check
	client, err := ssh.NewClient(host, user, "")
	if err != nil {
		// Return empty but don't log error - commands will handle display
		return installed
	}
	defer client.Close()

	// Check each package
	packages := installer.GetPackages()
	for pkgID := range packages {
		if checkPackageInstalled(client, pkgID) {
			installed[pkgID] = true
		}
	}

	return installed
}

func showPackageDetails(packageID string, packages map[string]*installer.Package) {
	pkg, exists := packages[packageID]
	if !exists {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Package '%s' not found", packageID)))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Available packages:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages"))
		fmt.Println()
		return
	}

	// Check if installed
	isInstalled := false
	installedPackages := checkInstalledPackages()
	if installedPackages[packageID] {
		isInstalled = true
	}

	displaySinglePackageStatus(pkg, isInstalled)
}

func displayList(packages map[string]*installer.Package, installedPackages map[string]bool) {
	// Anime-style banner
	fmt.Println(theme.RenderBanner("⚡ ANIME PACKAGE REGISTRY ⚡"))
	fmt.Println()

	// Show connection status
	if len(installedPackages) > 0 {
		installedCount := len(installedPackages)
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render("✓ Connected to Lambda:"),
			theme.HighlightStyle.Render(fmt.Sprintf("%d/%d packages installed", installedCount, len(packages))))
		fmt.Println()
	}

	// Group by category
	categories := make(map[string][]*installer.Package)
	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort categories
	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	for _, cat := range catNames {
		// Use category-specific styling
		catStyle := theme.GetCategoryStyle(cat)
		fmt.Println(catStyle.Render(theme.SymbolSakura + " " + cat))

		pkgs := categories[cat]
		sort.Slice(pkgs, func(i, j int) bool {
			return pkgs[i].ID < pkgs[j].ID
		})

		for _, pkg := range pkgs {
			// Installation status badge
			statusBadge := ""
			if len(installedPackages) > 0 {
				if installedPackages[pkg.ID] {
					statusBadge = " " + theme.SuccessStyle.Render("✓ INSTALLED")
				} else {
					statusBadge = " " + theme.DimTextStyle.Render("◯ not installed")
				}
			}

			// Package name with sparkle and status
			fmt.Printf("  %s %s%s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(pkg.Name),
				statusBadge)

			// Description
			fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(pkg.Description))

			// Metadata with anime symbols
			meta := fmt.Sprintf("%s %s  %s %s  %s %s",
				theme.SymbolBolt, pkg.ID,
				"⏱️", pkg.EstimatedTime,
				"💾", pkg.Size)
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(meta))

			// Dependencies with special styling
			if len(pkg.Dependencies) > 0 {
				deps := strings.Join(pkg.Dependencies, " → ")
				fmt.Printf("    %s\n",
					theme.WarningStyle.Render("⚡ Requires: "+deps))
			}
			fmt.Println()
		}
	}

	// Next steps
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 What to do next:"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime interactive"))
	fmt.Println(theme.DimTextStyle.Render("    Launch interactive package selector TUI"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <package-id>"))
	fmt.Println(theme.DimTextStyle.Render("    Install a specific package"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages --tree"))
	fmt.Println(theme.DimTextStyle.Render("    View dependency tree visualization"))
	fmt.Println()
	if len(installedPackages) == 0 {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime config"))
		fmt.Println(theme.DimTextStyle.Render("    Configure Lambda server to see installation status"))
		fmt.Println()
	}
}

func displayTree(packages map[string]*installer.Package, installedPackages map[string]bool) {
	// Anime-style tree banner
	fmt.Println(theme.RenderBanner("🌸 ANIME PACKAGE TREE 🌸"))
	fmt.Println()

	// Show connection status
	if len(installedPackages) > 0 {
		installedCount := len(installedPackages)
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render("✓ Connected to Lambda:"),
			theme.HighlightStyle.Render(fmt.Sprintf("%d/%d packages installed", installedCount, len(packages))))
		fmt.Println()
	}

	// Category order and emojis
	categoryOrder := []string{
		"Foundation", "GPU", "Runtime", "ML Framework", "LLM Runtime", "Containers",
		"LLM", "Models", "Image Generation", "Video Generation",
		"Image Enhancement", "Video Enhancement", "ControlNet",
		"Application", "ComfyUI Node",
	}

	categoryEmojis := map[string]string{
		"Foundation":        "🏗️",
		"GPU":               "🎮",
		"Runtime":           "⚙️",
		"ML Framework":      "🤖",
		"LLM Runtime":       "🔮",
		"Containers":        "📦",
		"LLM":               "💬",
		"Models":            "⭐",
		"Image Generation":  "🎨",
		"Video Generation":  "🎬",
		"Image Enhancement": "✨",
		"Video Enhancement": "📹",
		"ControlNet":        "🎛️",
		"Application":       "🎯",
		"ComfyUI Node":      "🔌",
	}

	// Group packages by category
	categories := make(map[string][]*installer.Package)
	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort packages within each category
	for cat := range categories {
		sort.Slice(categories[cat], func(i, j int) bool {
			return categories[cat][i].Name < categories[cat][j].Name
		})
	}

	// Print tree by category
	for catIdx, category := range categoryOrder {
		pkgs := categories[category]
		if len(pkgs) == 0 {
			continue
		}

		emoji := categoryEmojis[category]
		if emoji == "" {
			emoji = "📦"
		}

		// Category branch
		catMarker := "├──"
		if catIdx == len(categoryOrder)-1 || !hasMoreCategories(categoryOrder[catIdx+1:], categories) {
			catMarker = "└──"
		}

		fmt.Printf("%s %s %s (%d)\n",
			theme.InfoStyle.Render(catMarker),
			emoji,
			theme.HighlightStyle.Render(category),
			len(pkgs))

		// Packages in category
		for i, pkg := range pkgs {
			// Installation status
			statusBadge := theme.DimTextStyle.Render("◯")
			if installedPackages[pkg.ID] {
				statusBadge = theme.SuccessStyle.Render("✓")
			}

			// Package branch styling
			prefix := "│   "
			if catIdx == len(categoryOrder)-1 || !hasMoreCategories(categoryOrder[catIdx+1:], categories) {
				prefix = "    "
			}

			pkgMarker := "├──"
			if i == len(pkgs)-1 {
				pkgMarker = "└──"
			}

			// Name style based on installation
			nameStyle := theme.DimTextStyle
			if installedPackages[pkg.ID] {
				nameStyle = theme.SuccessStyle
			}

			fmt.Printf("%s%s %s %s %s\n",
				prefix,
				theme.DimTextStyle.Render(pkgMarker),
				statusBadge,
				nameStyle.Render(pkg.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("[%s]", pkg.ID)))

			// Show size and dependencies for uninstalled packages
			if !installedPackages[pkg.ID] {
				innerPrefix := prefix
				if i == len(pkgs)-1 {
					innerPrefix += "    "
				} else {
					innerPrefix += "│   "
				}
				fmt.Printf("%s%s %s\n",
					innerPrefix,
					theme.DimTextStyle.Render("└─"),
					theme.DimTextStyle.Render(fmt.Sprintf("💾 %s  ⏱️ %s", pkg.Size, pkg.EstimatedTime)))
			}
		}
		fmt.Println()
	}

	// Footer legend
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Tree Legend"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  🏗️ Foundation    🤖 ML Framework   🔮 LLM Runtime    💬 LLMs"))
	fmt.Println(theme.DimTextStyle.Render("  🎨 Image Gen     🎬 Video Gen      ✨ Enhancement    🎛️ ControlNet"))
	fmt.Println(theme.DimTextStyle.Render("  🎯 Application   🔌 ComfyUI Node   ⭐ Model Packs"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  ✓ = Installed   ◯ = Not installed"))
	fmt.Println()

	// Next steps
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 What to do next:"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime library"))
	fmt.Println(theme.DimTextStyle.Render("    Launch tabbed package browser TUI"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime interactive"))
	fmt.Println(theme.DimTextStyle.Render("    Launch interactive package selector"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <package-id>"))
	fmt.Println(theme.DimTextStyle.Render("    Install a specific package"))
	fmt.Println()
}

// hasMoreCategories checks if any remaining categories have packages
func hasMoreCategories(remaining []string, categories map[string][]*installer.Package) bool {
	for _, cat := range remaining {
		if len(categories[cat]) > 0 {
			return true
		}
	}
	return false
}

// checkPackageInstalled checks if a package is installed on the remote server
// isRunningOnServer checks if we're running on a Lambda/GPU server (not remote client)
func isRunningOnServer() bool {
	// Check for NVIDIA GPU (primary indicator of Lambda server)
	cmd := exec.Command("command", "-v", "nvidia-smi")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Check if we're in a typical Lambda environment
	home := os.Getenv("HOME")
	if home == "" {
		return false
	}

	// Check for typical Lambda directories
	comfyUIPath := filepath.Join(home, "ComfyUI")
	videoModelsPath := filepath.Join(home, "video-models")

	if _, err := os.Stat(comfyUIPath); err == nil {
		return true
	}
	if _, err := os.Stat(videoModelsPath); err == nil {
		return true
	}

	return false
}

// checkPackageInstalledLocal checks if a package is installed locally (without SSH)
func checkPackageInstalledLocal(packageID string) bool {
	// Define check commands for each package type
	checkCommands := map[string]string{
		"core":          "command -v nvidia-smi && command -v nvcc",
		"python":        "command -v python3 && python3 -c 'import numpy' 2>/dev/null",
		"pytorch":       "python3 -c 'import torch' 2>/dev/null",
		"ollama":        "command -v ollama",
		"vllm":          "python3 -c 'import vllm' 2>/dev/null",
		"models-small":  "ollama list 2>/dev/null | grep -q 'llama3.3:8b'",
		"models-medium": "ollama list 2>/dev/null | grep -q 'qwen3:14b'",
		"models-large":  "ollama list 2>/dev/null | grep -q 'llama3.3:70b'",
		"nodejs":        "command -v node",
		"go":            "command -v go",
		"claude":        "command -v claude",
		"comfyui":       "test -d ~/ComfyUI",
		"mochi":         "test -d ~/video-models/mochi-1",
		"svd":           "test -d ~/ComfyUI/models/checkpoints && ls ~/ComfyUI/models/checkpoints/*svd* 2>/dev/null",
		"animatediff":   "test -d ~/ComfyUI/custom_nodes/ComfyUI-AnimateDiff-Evolved",
		"cogvideo":      "test -d ~/video-models/cogvideo",
		"opensora":      "test -d ~/video-models/open-sora",
		"ltxvideo":      "test -d ~/video-models/ltxvideo",
		"wan2":          "test -d ~/video-models/wan2",
	}

	checkCmd, exists := checkCommands[packageID]
	if !exists {
		return false
	}

	// Run command locally using sh
	cmd := exec.Command("sh", "-c", checkCmd)
	return cmd.Run() == nil
}

func checkPackageInstalled(client *ssh.Client, packageID string) bool {
	// Define check commands for each package type
	checkCommands := map[string]string{
		"core":          "command -v nvidia-smi && command -v nvcc",
		"python":        "command -v python3 && python3 -c 'import numpy' 2>/dev/null",
		"pytorch":       "python3 -c 'import torch' 2>/dev/null",
		"ollama":        "command -v ollama",
		"vllm":          "python3 -c 'import vllm' 2>/dev/null",
		"models-small":  "ollama list 2>/dev/null | grep -q 'llama3.3:8b'",
		"models-medium": "ollama list 2>/dev/null | grep -q 'qwen3:14b'",
		"models-large":  "ollama list 2>/dev/null | grep -q 'llama3.3:70b'",
		"nodejs":        "command -v node",
		"go":            "command -v go",
		"claude":        "command -v claude",
		"comfyui":       "test -d ~/ComfyUI",
		"mochi":         "test -d ~/video-models/mochi-1",
		"svd":           "test -d ~/ComfyUI/models/checkpoints && ls ~/ComfyUI/models/checkpoints/*svd* 2>/dev/null",
		"animatediff":   "test -d ~/ComfyUI/custom_nodes/ComfyUI-AnimateDiff-Evolved",
		"cogvideo":      "test -d ~/video-models/cogvideo",
		"opensora":      "test -d ~/video-models/open-sora",
		"ltxvideo":      "test -d ~/video-models/ltxvideo",
		"wan2":          "test -d ~/video-models/wan2",
	}

	checkCmd, exists := checkCommands[packageID]
	if !exists {
		return false
	}

	_, err := client.RunCommand(checkCmd)
	return err == nil
}

func displaySinglePackageStatus(pkg *installer.Package, isInstalled bool) {
	fmt.Println()

	// Status header
	if isInstalled {
		fmt.Println(theme.RenderBanner("✓ " + strings.ToUpper(pkg.Name) + " - INSTALLED"))
	} else {
		fmt.Println(theme.RenderBanner("◯ " + strings.ToUpper(pkg.Name) + " - NOT INSTALLED"))
	}

	fmt.Println()

	// Package details
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Package Information"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  ID:          %s\n", theme.HighlightStyle.Render(pkg.ID))
	fmt.Printf("  Name:        %s\n", theme.HighlightStyle.Render(pkg.Name))
	fmt.Printf("  Category:    %s\n", theme.HighlightStyle.Render(pkg.Category))
	fmt.Printf("  Description: %s\n", theme.SecondaryTextStyle.Render(pkg.Description))
	fmt.Println()

	// Installation status
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Status"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if isInstalled {
		fmt.Printf("  Status:      %s\n", theme.SuccessStyle.Render("✓ INSTALLED"))
	} else {
		fmt.Printf("  Status:      %s\n", theme.WarningStyle.Render("◯ NOT INSTALLED"))
		fmt.Printf("  Install:     %s\n", theme.HighlightStyle.Render("anime install "+pkg.ID))
	}
	fmt.Println()

	// Dependencies
	if len(pkg.Dependencies) > 0 {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("🔗 Dependencies"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		allPackages := installer.GetPackages()
		for _, depID := range pkg.Dependencies {
			if depPkg, exists := allPackages[depID]; exists {
				fmt.Printf("  • %s %s\n",
					theme.HighlightStyle.Render(depPkg.Name),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s)", depID)))
			} else {
				fmt.Printf("  • %s\n", theme.HighlightStyle.Render(depID))
			}
		}
		fmt.Println()
	}

	// Installation details
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("⏱️  Installation Details"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  Estimated Time: %s\n", theme.HighlightStyle.Render(pkg.EstimatedTime.String()))
	fmt.Printf("  Disk Space:     %s\n", theme.HighlightStyle.Render(pkg.Size))
	fmt.Println()
}

func runPackagesStatus(cmd *cobra.Command, args []string) {
	packages := installer.GetPackages()

	// If a specific package is requested
	if len(args) > 0 {
		packageID := args[0]
		pkg, exists := packages[packageID]
		if !exists {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Package '%s' not found", packageID)))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Available packages:"))
			var ids []string
			for id := range packages {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				fmt.Printf("  • %s\n", theme.HighlightStyle.Render(id))
			}
			fmt.Println()
			return
		}

		// Check installation status
		installedPackages := checkInstalledPackages()
		if len(installedPackages) == 0 {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("⚠️  Lambda server not configured or unreachable"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Configure your Lambda server:"))
			fmt.Println(theme.HighlightStyle.Render("  $ anime set lambda <server-ip>"))
			fmt.Println()
			return
		}

		// Display single package status
		displaySinglePackageStatus(pkg, installedPackages[packageID])
		return
	}

	// Display all packages
	fmt.Println()
	fmt.Println(theme.RenderBanner("📊 PACKAGE STATUS 📊"))
	fmt.Println()

	// Check installation status
	installedPackages := checkInstalledPackages()

	if len(installedPackages) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠️  Lambda server not configured or unreachable"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Configure your Lambda server:"))
		fmt.Println(theme.HighlightStyle.Render("  $ anime set lambda <server-ip>"))
		fmt.Println()
		return
	}

	// Group packages by category
	categoryOrder := []string{"Foundation", "ML Framework", "LLM Runtime", "Models", "Video Generation", "Application"}
	categories := make(map[string][]*installer.Package)

	for _, pkg := range packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}

	// Sort packages within each category by ID
	for _, pkgs := range categories {
		sort.Slice(pkgs, func(i, j int) bool {
			return pkgs[i].ID < pkgs[j].ID
		})
	}

	// Count installed vs total
	totalInstalled := 0
	totalPackages := len(packages)
	for pkgID := range packages {
		if installedPackages[pkgID] {
			totalInstalled++
		}
	}

	// Display summary
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Total Packages:   %d", totalPackages)))
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ Installed:      %d", totalInstalled)))
	fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  ◯ Not Installed:  %d", totalPackages-totalInstalled)))
	fmt.Println()

	// Display packages by category
	for _, category := range categoryOrder {
		pkgs := categories[category]
		if len(pkgs) == 0 {
			continue
		}

		// Category header with emoji
		emoji := "📦"
		switch category {
		case "Foundation":
			emoji = "🏗️"
		case "ML Framework":
			emoji = "🤖"
		case "LLM Runtime":
			emoji = "🔮"
		case "Models":
			emoji = "⭐"
		case "Video Generation":
			emoji = "🎬"
		case "Application":
			emoji = "🎯"
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Printf("%s %s\n", emoji, theme.InfoStyle.Render(category))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

		for _, pkg := range pkgs {
			isInstalled := installedPackages[pkg.ID]

			// Status badge and styling
			var statusBadge, statusText, pkgName string

			if isInstalled {
				statusBadge = theme.SuccessStyle.Render("✓")
				statusText = theme.SuccessStyle.Render("INSTALLED")
				pkgName = theme.HighlightStyle.Render(pkg.Name)
			} else {
				statusBadge = theme.DimTextStyle.Render("◯")
				statusText = theme.WarningStyle.Render("NOT INSTALLED")
				pkgName = theme.DimTextStyle.Render(pkg.Name)
			}

			// Package line
			fmt.Printf("  %s %s %s\n",
				statusBadge,
				pkgName,
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", pkg.ID)))
			fmt.Printf("    %s\n", statusText)

			// Show description and metadata if not installed
			if !isInstalled {
				fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(pkg.Description))
				fmt.Printf("    %s  |  %s\n",
					theme.DimTextStyle.Render("⏱️  "+pkg.EstimatedTime.String()),
					theme.DimTextStyle.Render("💾 "+pkg.Size))
			}
			fmt.Println()
		}
	}

	// Show helpful commands
	if totalInstalled < totalPackages {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("✨ Quick Actions"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime lambda defaults"))
		fmt.Println(theme.DimTextStyle.Render("    View recommended packages for Lambda GPU"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime interactive"))
		fmt.Println(theme.DimTextStyle.Render("    Select packages with interactive TUI"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime lambda install <package>"))
		fmt.Println(theme.DimTextStyle.Render("    Install specific packages"))
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.SuccessStyle.Render("  ✨ All packages installed! Your system is fully set up! ✨"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
	}
}
