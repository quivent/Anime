package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	packagesCmd = &cobra.Command{
		Use:   "packages",
		Short: "List available installation packages",
		Long:  "Display all available packages with dependencies, descriptions, and estimated installation times",
		Run:   runPackages,
	}

	showTree bool
)

func init() {
	rootCmd.AddCommand(packagesCmd)
	packagesCmd.Flags().BoolVarP(&showTree, "tree", "t", false, "Show dependency tree")
}

func runPackages(cmd *cobra.Command, args []string) {
	packages := installer.GetPackages()

	if showTree {
		displayTree(packages)
		return
	}

	displayList(packages)
}

func displayList(packages map[string]*installer.Package) {
	// Anime-style banner
	fmt.Println(theme.RenderBanner("⚡ ANIME PACKAGE REGISTRY ⚡"))
	fmt.Println()

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
			// Package name with sparkle
			fmt.Printf("  %s %s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(pkg.Name))

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

	// Footer with helpful hints
	fmt.Println(theme.InfoStyle.Render("\n✨ Commands:"))
	fmt.Println(theme.DimTextStyle.Render("  • anime packages --tree  - View dependency tree"))
	fmt.Println(theme.DimTextStyle.Render("  • anime install <id>     - Install package"))
	fmt.Println(theme.DimTextStyle.Render("  • anime interactive      - Interactive selection"))
}

func displayTree(packages map[string]*installer.Package) {
	// Anime-style tree banner
	fmt.Println(theme.RenderBanner("🌸 DEPENDENCY SAKURA TREE 🌸"))
	fmt.Println()

	// Find root packages (no dependencies)
	var roots []*installer.Package
	for _, pkg := range packages {
		if len(pkg.Dependencies) == 0 {
			roots = append(roots, pkg)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID < roots[j].ID
	})

	printed := make(map[string]bool)

	var printTree func(*installer.Package, string, bool, int)
	printTree = func(pkg *installer.Package, prefix string, isLast bool, depth int) {
		// Different colors for different depths
		var nameStyle func(string) string
		switch depth % 4 {
		case 0:
			nameStyle = func(s string) string { return theme.SuccessStyle.Render(s) }
		case 1:
			nameStyle = func(s string) string { return theme.InfoStyle.Render(s) }
		case 2:
			nameStyle = func(s string) string { return theme.WarningStyle.Render(s) }
		case 3:
			nameStyle = func(s string) string { return theme.GlowStyle.Render(s) }
		}

		if printed[pkg.ID] {
			marker := theme.SymbolBranch
			if isLast {
				marker = theme.SymbolLastBranch
			}
			fmt.Printf("%s%s %s %s\n",
				prefix,
				theme.DimTextStyle.Render(marker),
				nameStyle(pkg.Name),
				theme.DimTextStyle.Render("(shown above)"))
			return
		}

		marker := theme.SymbolBranch
		extension := theme.SymbolPipe + "  "
		if isLast {
			marker = theme.SymbolLastBranch
			extension = theme.SymbolSpace + " "
		}

		// Package with emoji based on category
		emoji := theme.SymbolSakura
		switch pkg.Category {
		case "Foundation":
			emoji = "🏗️"
		case "ML Framework":
			emoji = "🤖"
		case "LLM Runtime":
			emoji = "🔮"
		case "Models":
			emoji = "⭐"
		case "Application":
			emoji = "🎯"
		}

		fmt.Printf("%s%s %s %s %s\n",
			prefix,
			theme.InfoStyle.Render(marker),
			emoji,
			nameStyle(pkg.Name),
			theme.DimTextStyle.Render(fmt.Sprintf("[%s]", pkg.ID)))

		printed[pkg.ID] = true

		// Find packages that depend on this one
		var dependents []*installer.Package
		for _, p := range packages {
			for _, dep := range p.Dependencies {
				if dep == pkg.ID {
					dependents = append(dependents, p)
					break
				}
			}
		}

		sort.Slice(dependents, func(i, j int) bool {
			return dependents[i].ID < dependents[j].ID
		})

		for i, dep := range dependents {
			printTree(dep, prefix+extension, i == len(dependents)-1, depth+1)
		}
	}

	for i, root := range roots {
		printTree(root, "", i == len(roots)-1, 0)
		fmt.Println()
	}

	// Footer
	fmt.Println(theme.InfoStyle.Render("✨ Tree Legend:"))
	fmt.Println(theme.DimTextStyle.Render("  🏗️  Foundation   🤖 ML Framework   🔮 LLM Runtime"))
	fmt.Println(theme.DimTextStyle.Render("  ⭐ Models       🎯 Application"))
}
