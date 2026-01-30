package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var enhanceCmd = &cobra.Command{
	Use:   "enhance [path]",
	Short: "Analyze codebase and propose 3-7 improvements",
	Long: `Launch Claude Code with an enhancement prompt that analyzes the codebase
and proposes 3-7 concrete improvements.

This command is similar to 'develop' but automatically starts Claude Code with
a prompt that performs a comprehensive analysis of the codebase and suggests
actionable enhancements.

On remote servers where the embedded build path doesn't exist, you can:
  - Provide a custom path:  anime enhance /path/to/project
  - Use current directory:  anime enhance .`,
	Run: runEnhance,
}

func init() {
	rootCmd.AddCommand(enhanceCmd)
}

const enhancePrompt = `Perform a comprehensive analysis of this codebase and propose 3-7 concrete improvements.

## Analysis Requirements

1. **Codebase Understanding**: Thoroughly explore the project structure, architecture, dependencies, and key components.

2. **Quality Assessment**: Evaluate code quality, patterns, potential issues, and areas for improvement across:
   - Code organization and architecture
   - Performance bottlenecks
   - Security considerations
   - Developer experience
   - Testing coverage
   - Documentation gaps
   - Technical debt

3. **Propose 3-7 Improvements**: Based on your analysis, propose exactly 3-7 specific, actionable improvements. Each proposal should include:
   - **Title**: A clear, concise name for the improvement
   - **Problem**: What issue or gap does this address?
   - **Solution**: Concrete steps to implement the improvement
   - **Impact**: Expected benefits (high/medium/low) and affected areas
   - **Effort**: Estimated complexity (small/medium/large)

## Output Format

Present your findings as:

### Codebase Overview
Brief summary of the project, its purpose, and tech stack.

### Proposed Improvements

For each improvement (3-7 total):

#### [N]. [Title]
**Problem**: [Description of the issue]
**Solution**: [Concrete implementation steps]
**Impact**: [High/Medium/Low] - [What improves]
**Effort**: [Small/Medium/Large]

### Priority Recommendation
Rank the improvements by suggested implementation order with brief justification.

---

Start by exploring the codebase structure, then provide your analysis and proposals.`

func runEnhance(cmd *cobra.Command, args []string) {
	// Resolve the source directory to use
	sourceDir := resolveEnhanceDirectory(args)
	if sourceDir == "" {
		os.Exit(1)
	}

	// Check if claude command is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("💥 Claude Code Not Found"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  The 'claude' command is not available in your PATH."))
		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("  💡 Install Claude Code:"))
		fmt.Println()
		fmt.Printf("    %s\n", theme.HighlightStyle.Render("npm install -g @anthropic-ai/claude-code"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Or visit: https://github.com/anthropics/claude-code"))
		fmt.Println()
		os.Exit(1)
	}

	// Display info before launching
	fmt.Println()
	fmt.Println(theme.RenderBanner("✨ ANIME ENHANCE ✨"))
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Target:"),
		theme.SuccessStyle.Render(sourceDir))
	fmt.Printf("  %s  %s\n",
		theme.InfoStyle.Render("Claude:"),
		theme.DimTextStyle.Render(claudePath))
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("  Launching Claude Code with enhancement analysis prompt..."))
	fmt.Println()

	// Change to source directory
	if err := os.Chdir(sourceDir); err != nil {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Failed to change directory: %v", err)))
		fmt.Println()
		os.Exit(1)
	}

	// Prepare arguments for claude with the enhancement prompt
	claudeArgs := []string{"claude", "--permission-mode", "bypassPermissions", enhancePrompt}

	// Execute claude, replacing the current process
	err = syscall.Exec(claudePath, claudeArgs, os.Environ())
	if err != nil {
		// If exec fails, we'll still be here
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Failed to launch Claude Code: %v", err)))
		fmt.Println()
		os.Exit(1)
	}
}

// resolveEnhanceDirectory finds the directory to analyze
func resolveEnhanceDirectory(args []string) string {
	// 1. If user provided a path argument, use that
	if len(args) > 0 {
		customPath := args[0]
		if customPath == "." {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println()
				fmt.Println(theme.ErrorStyle.Render("💥 Failed to get current directory"))
				fmt.Println()
				return ""
			}
			return cwd
		}

		// Check if directory exists
		info, err := os.Stat(customPath)
		if err != nil || !info.IsDir() {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("💥 Invalid Directory"))
			fmt.Println()
			fmt.Printf("  %s %s\n",
				theme.WarningStyle.Render("Provided path:"),
				theme.DimTextStyle.Render(customPath))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  The path doesn't exist or is not a directory."))
			fmt.Println()
			return ""
		}
		return customPath
	}

	// 2. Default to anime source directory (same as develop)
	sourceDir := resolveSourceDirectory(nil)
	if sourceDir != "" {
		return sourceDir
	}

	// 3. Fall back to current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("💥 Failed to get current directory"))
		fmt.Println()
		return ""
	}

	fmt.Println()
	fmt.Printf("  %s %s\n",
		theme.InfoStyle.Render("Using current directory:"),
		theme.SuccessStyle.Render(cwd))
	return cwd
}
