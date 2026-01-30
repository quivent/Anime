package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var guideCmd = &cobra.Command{
	Use:     "guide",
	Short:   "Show comprehensive usage guide",
	Aliases: []string{"help-guide", "tutorial"},
	Run:     runGuide,
}

func init() {
	rootCmd.AddCommand(guideCmd)
}

func runGuide(cmd *cobra.Command, args []string) {
	guide := `
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║                    🌸 ANIME - USAGE GUIDE 🌸                        ║
║              Lambda GPU Management & AI Installation                ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📖 TABLE OF CONTENTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  1. Quick Start
  2. Server Configuration
  3. Installing Packages
  4. Parallel Installation
  5. Lambda Server Management
  6. Push & Deploy
  7. Interactive Selection
  8. Video Generation Models
  9. Tips & Tricks


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 1. QUICK START
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Get up and running in minutes!

  ▸ Step 1: Configure your Lambda server
    $ anime set lambda 209.20.159.132

  ▸ Step 2: Push anime to the server
    $ anime push

  ▸ Step 3: Install packages
    $ anime lambda install python pytorch ollama

That's it! Your Lambda GPU is now ready for AI development.


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚙️  2. SERVER CONFIGURATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Managing server aliases and configuration.

  ▸ Create an alias
    $ anime set lambda <server-ip>
    $ anime set lambda ubuntu@192.168.1.100

  ▸ List all aliases
    $ anime set --list

  ▸ Delete an alias
    $ anime set --delete lambda

  ▸ Use different server
    $ anime set production 10.0.0.5
    $ anime push production


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📦 3. INSTALLING PACKAGES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Browse and install AI packages on your Lambda server.

  ▸ Browse all packages
    $ anime packages

  ▸ View dependency tree
    $ anime packages --tree

  ▸ Install single package
    $ anime lambda install python

  ▸ Install multiple packages
    $ anime lambda install pytorch ollama models-small

  ▸ Interactive selection
    $ anime interactive

  💡 All packages show installation status with ✓ or ◯


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚡ 4. PARALLEL INSTALLATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Speed up installations with automatic parallelization!

  ▸ Default: Parallel enabled
    $ anime lambda install pytorch ollama comfyui
    → Installs independent packages concurrently
    → Uses all CPU cores for compilation
    → 2-3x faster than sequential

  ▸ Disable parallel (sequential)
    $ anime lambda install python --parallel=false

  ▸ Limit CPU cores
    $ anime lambda install pytorch --jobs 8

  ▸ How it works:
    • Analyzes dependency graph
    • Installs up to 3 modules concurrently
    • Auto-detects all CPU cores with nproc
    • Respects dependencies (pytorch after core)


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🖥️  5. LAMBDA SERVER MANAGEMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Commands specific to Lambda GPU servers.

  ▸ Install on lambda server
    $ anime lambda install <packages>

  ▸ View default packages
    $ anime lambda defaults
    → Shows recommended starter pack
    → Displays installation status
    → Estimates time and cost

  ▸ Default starter pack includes:
    • core        - CUDA, Python, Docker, Node.js
    • python      - AI libraries (numpy, scipy, pandas)
    • pytorch     - PyTorch + transformers
    • ollama      - LLM runtime server
    • models-small- 7-8B models (Llama, Mistral, Qwen)
    • claude      - Claude Code CLI


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 6. PUSH & DEPLOY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Deploy the anime binary to remote servers.

  ▸ Push to lambda (default)
    $ anime push

  ▸ Push to specific server
    $ anime push 192.168.1.100
    $ anime push user@10.0.0.5

  ▸ Include source code
    $ anime push --source

  ▸ Different architecture
    $ anime push --arch arm64

  ▸ What it does:
    1. Tests SSH connection
    2. Builds Linux binary
    3. Creates tar.gz package
    4. Rsyncs to server
    5. Extracts to ~/.local/bin/anime
    6. Sets executable permissions

  💡 Binary is installed to ~/.local/bin on the server


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎨 7. INTERACTIVE SELECTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Beautiful TUI for package selection.

  ▸ Launch interactive mode
    $ anime interactive

  ▸ Navigation:
    • ↑/↓ or k/j  - Move cursor
    • Space       - Toggle selection
    • Enter       - Confirm and install
    • q           - Quit

  ▸ Features:
    • Category-based layout
    • Installation status (✓ or ◯)
    • Live time estimates
    • Dependency resolution
    • Beautiful anime-themed colors


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎬 8. VIDEO GENERATION MODELS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

State-of-the-art video generation models available.

  ┌─────────────────┬──────────────────┬──────┬─────────────────┐
  │ Model           │ Type             │ Size │ Best For        │
  ├─────────────────┼──────────────────┼──────┼─────────────────┤
  │ Wan2.2          │ Image-to-video   │ 10GB │ I2V conversion  │
  │ Mochi-1         │ Text/Image-video │ 12GB │ Open source     │
  │ SVD             │ ComfyUI plugin   │  8GB │ SD-based video  │
  │ AnimateDiff     │ ComfyUI plugin   │  4GB │ Image animation │
  │ CogVideoX-5B    │ Text-to-video    │ 14GB │ Chinese model   │
  │ Open-Sora 2.0   │ Text-to-video    │ 16GB │ Long videos     │
  │ LTXVideo        │ Text-to-video    │  7GB │ Fast generation │
  └─────────────────┴──────────────────┴──────┴─────────────────┘

  ▸ Install single model
    $ anime lambda install wan2

  ▸ Install multiple models (parallel!)
    $ anime lambda install wan2 mochi ltxvideo
    → Installs all 3 in parallel
    → ~20min instead of ~55min sequential


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 9. TIPS & TRICKS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Pro tips for maximum efficiency!

  ▸ Check what's installed
    All commands show ✓ or ◯ status automatically:
    • anime packages
    • anime packages --tree
    • anime interactive
    • anime lambda defaults

  ▸ Speed up downloads
    Video model scripts use:
    • aria2c for multi-connection downloads
    • huggingface-cli with --max-workers 8
    • Parallel pip installs where possible

  ▸ Optimize compilation
    Parallel mode automatically sets:
    • MAKEFLAGS=-j$(nproc)
    • CMAKE_BUILD_PARALLEL_LEVEL=$(nproc)
    • Uses all available CPU cores

  ▸ SSH config aliases work
    $ anime push my-ssh-alias
    → Resolves SSH config automatically

  ▸ Helpful suggestions on errors
    Every failure shows:
    • What went wrong
    • Suggested fixes
    • Example commands
    • No more cryptic errors!

  ▸ Cost tracking
    Installation TUI shows:
    • Elapsed time
    • Estimated cost
    • Progress per module

  ▸ Resume installations
    Scripts check if already installed:
    • Skips completed modules
    • Safe to re-run
    • No wasted time


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📚 ADDITIONAL RESOURCES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Command help
    $ anime <command> --help

  ▸ Full command tree
    $ anime tree

  ▸ Package categories
    • Foundation     - Core system, Python, CUDA
    • ML Framework   - PyTorch, AI libraries
    • LLM Runtime    - Ollama server
    • Models         - Pre-trained LLMs
    • Video Gen      - Video generation models
    • Application    - ComfyUI, Claude Code


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🌸 EXAMPLE WORKFLOWS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🎯 Complete Setup from Scratch
  ───────────────────────────────
  $ anime set lambda 209.20.159.132
  $ anime push
  $ anime lambda install core python pytorch ollama models-small

  🎬 Video Generation Setup
  ──────────────────────────
  $ anime lambda install comfyui wan2 mochi svd
  → Parallel installation in ~25 minutes

  🤖 LLM Development Setup
  ─────────────────────────
  $ anime lambda install ollama models-medium claude
  → Ollama + 14-34B models + Claude Code

  🔍 Explore Before Installing
  ──────────────────────────────
  $ anime packages --tree
  $ anime interactive
  → See what's available and installed


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

                    ✨ Happy coding with anime! ✨

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`

	// Print the guide with styling
	lines := strings.Split(guide, "\n")
	for _, line := range lines {
		// Apply different styles based on line content
		if strings.HasPrefix(line, "━━") {
			fmt.Println(theme.InfoStyle.Render(line))
		} else if strings.HasPrefix(line, "╔") || strings.HasPrefix(line, "╚") || strings.HasPrefix(line, "║") {
			fmt.Println(theme.GlowStyle.Render(line))
		} else if strings.Contains(line, "▸") {
			// Command examples
			parts := strings.SplitN(line, "$", 2)
			if len(parts) == 2 {
				fmt.Print(theme.DimTextStyle.Render(parts[0]))
				fmt.Println(theme.HighlightStyle.Render("$" + parts[1]))
			} else {
				fmt.Println(theme.InfoStyle.Render(line))
			}
		} else if strings.HasPrefix(line, "  •") || strings.HasPrefix(line, "  ▸") {
			fmt.Println(theme.SuccessStyle.Render(line))
		} else if strings.Contains(line, "→") {
			fmt.Println(theme.WarningStyle.Render(line))
		} else if strings.Contains(line, "💡") || strings.Contains(line, "⚡") ||
		          strings.Contains(line, "🎯") || strings.Contains(line, "🎬") ||
		          strings.Contains(line, "🤖") || strings.Contains(line, "🔍") {
			fmt.Println(theme.InfoStyle.Render(line))
		} else if strings.HasPrefix(line, "  ┌") || strings.HasPrefix(line, "  ├") ||
		          strings.HasPrefix(line, "  │") || strings.HasPrefix(line, "  └") {
			fmt.Println(theme.DimTextStyle.Render(line))
		} else {
			fmt.Println(line)
		}
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  For command-specific help: anime <command> --help"))
	fmt.Println(theme.DimTextStyle.Render("  View full command tree: anime tree"))
	fmt.Println()
}
