package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var treeDepth int
var treeCompact bool

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show anime CLI command tree",
	Long:  "Display a complete tree visualization of all available anime commands, subcommands, and flags",
	Run:   runTree,
}

func init() {
	treeCmd.Flags().IntVarP(&treeDepth, "depth", "d", 0, "Maximum depth to display (0 = unlimited)")
	treeCmd.Flags().BoolVarP(&treeCompact, "compact", "c", false, "Compact view without descriptions")
	rootCmd.AddCommand(treeCmd)
}

func runTree(cmd *cobra.Command, args []string) {
	theme.RenderBanner("ANIME COMMAND TREE")
	fmt.Println()

	// Count total commands
	total := countCommands(rootCmd)
	fmt.Printf("  %s %d commands\n\n", theme.DimTextStyle.Render("Total:"), total)

	// Print the tree starting from root
	printCommandTree(rootCmd, "", true, 0)

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))
	fmt.Println(theme.InfoStyle.Render("  💡 Tips:"))
	fmt.Println(theme.DimTextStyle.Render("     anime tree --compact     Compact view"))
	fmt.Println(theme.DimTextStyle.Render("     anime tree --depth 2     Limit depth"))
	fmt.Println(theme.DimTextStyle.Render("     anime <cmd> --help       Detailed help"))
	fmt.Println()
}

func countCommands(cmd *cobra.Command) int {
	count := 1
	for _, sub := range cmd.Commands() {
		if !sub.Hidden {
			count += countCommands(sub)
		}
	}
	return count
}

func printCommandTree(cmd *cobra.Command, _ string, isRoot bool, depth int) {
	if treeDepth > 0 && depth > treeDepth {
		return
	}

	// Get visible subcommands
	subs := getVisibleCommands(cmd)

	if isRoot {
		// Print root command
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime"))
		if !treeCompact {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(cmd.Short))
		}
		fmt.Println()

		// Print subcommands
		for i, sub := range subs {
			isLast := i == len(subs)-1
			printSubcommand(sub, "  ", isLast, depth+1)
		}
	}
}

func printSubcommand(cmd *cobra.Command, prefix string, isLast bool, depth int) {
	if treeDepth > 0 && depth > treeDepth {
		return
	}

	// Determine branch characters
	branch := "├── "
	childPrefix := "│   "
	if isLast {
		branch = "└── "
		childPrefix = "    "
	}

	// Command name with styling
	cmdName := cmd.Name()
	if len(cmd.Commands()) > 0 {
		// Has subcommands - show as group
		cmdName = theme.InfoStyle.Render(cmdName)
	} else {
		cmdName = theme.HighlightStyle.Render(cmdName)
	}

	// Print command
	if treeCompact {
		fmt.Printf("%s%s%s\n", prefix, theme.DimTextStyle.Render(branch), cmdName)
	} else {
		desc := cmd.Short
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		fmt.Printf("%s%s%s %s\n", prefix, theme.DimTextStyle.Render(branch), cmdName, theme.DimTextStyle.Render("- "+desc))
	}

	// Get visible subcommands
	subs := getVisibleCommands(cmd)

	// Print subcommands
	for i, sub := range subs {
		subIsLast := i == len(subs)-1
		printSubcommand(sub, prefix+childPrefix, subIsLast, depth+1)
	}
}

func getVisibleCommands(cmd *cobra.Command) []*cobra.Command {
	var visible []*cobra.Command
	for _, sub := range cmd.Commands() {
		if !sub.Hidden && sub.Name() != "help" && sub.Name() != "completion" {
			visible = append(visible, sub)
		}
	}
	// Sort alphabetically
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Name() < visible[j].Name()
	})
	return visible
}

// TreeStats returns statistics about the command tree
func TreeStats() (commands, subcommands, flags int) {
	return walkCommandTree(rootCmd)
}

func walkCommandTree(cmd *cobra.Command) (commands, subcommands, flags int) {
	commands = 1
	flags = cmd.Flags().NFlag()

	for _, sub := range cmd.Commands() {
		if !sub.Hidden {
			subcommands++
			c, s, f := walkCommandTree(sub)
			commands += c
			subcommands += s
			flags += f
		}
	}
	return
}

// PrintCommandPath prints the full path to a command
func PrintCommandPath(cmd *cobra.Command) string {
	var parts []string
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() != "" {
			parts = append([]string{c.Name()}, parts...)
		}
	}
	return strings.Join(parts, " ")
}
