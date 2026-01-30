package reel

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/theme"
)

// Handle is the main entry point for the reel command
func Handle(args []string) {
	cfg := LoadSession()

	if len(args) == 0 {
		PrintHelp(cfg)
		return
	}

	switch args[0] {
	// Core subcommands
	case "prompt":
		HandlePrompt(cfg, args[1:])
	case "frames":
		HandleFrames(cfg, args[1:])
	case "resolution", "res":
		HandleResolution(cfg, args[1:])
	case "steps":
		HandleSteps(cfg, args[1:])
	case "guidance", "cfg":
		HandleGuidance(cfg, args[1:])
	case "seed":
		HandleSeed(cfg, args[1:])
	case "output", "out", "outdir":
		HandleOutput(cfg, args[1:])
	case "model":
		HandleModel(cfg, args[1:])
	case "image", "img":
		HandleImage(cfg, args[1:])
	case "fps":
		HandleFPS(cfg, args[1:])

	// Optimization subcommands
	case "usp", "parallel":
		HandleUSP(cfg, args[1:])
	case "offload":
		HandleOffload(cfg, args[1:])
	case "teacache", "cache":
		HandleTeaCache(cfg, args[1:])
	case "script":
		HandleScript(cfg, args[1:])

	// Execution subcommands
	case "run", "generate", "gen":
		Execute(cfg)
	case "dry", "dry-run", "preview":
		ExecuteDry(cfg)
	case "show", "config":
		PrintConfig(cfg)
	case "reset":
		ResetConfig()

	// Help
	case "help", "-h", "--help":
		PrintHelp(cfg)

	default:
		printSuggestion("Unknown subcommand: "+args[0], []string{
			"Run 'anime reel help' to see available subcommands",
			"Common: prompt, frames, resolution, run",
		})
	}
}

// PrintHelp prints the reel help
func PrintHelp(cfg *ReelConfig) {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎬 SkyReel Video Generation"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Subcommand-based interface for SkyReels-V2"))
	fmt.Println()

	printSection("Usage")
	fmt.Printf("  %s <subcommand> [value]\n", theme.HighlightStyle.Render("anime reel"))
	fmt.Println()

	printSection("Configuration Subcommands")
	printSubcommand("prompt", "Set generation prompt")
	printSubcommand("frames", "Set frame count (97, 4s, etc)")
	printSubcommand("resolution", "Set resolution (540p, 720p)")
	printSubcommand("steps", "Set inference steps")
	printSubcommand("guidance", "Set CFG guidance scale")
	printSubcommand("seed", "Set random seed")
	printSubcommand("output", "Set output directory")
	printSubcommand("model", "Select model variant")
	printSubcommand("image", "Set input image (for i2v)")
	printSubcommand("fps", "Set frames per second")
	printSubcommand("script", "Select generation script")

	printSection("Optimization Subcommands")
	printSubcommand("usp", "Toggle multi-GPU parallelism")
	printSubcommand("offload", "Toggle CPU offloading")
	printSubcommand("teacache", "Toggle/configure TeaCache")

	printSection("Execution Subcommands")
	printSubcommand("run", "Execute generation")
	printSubcommand("dry", "Preview without executing")
	printSubcommand("show", "Show current configuration")
	printSubcommand("reset", "Reset to defaults")

	printSection("Examples")
	fmt.Printf("    %s anime reel prompt \"A serene lake at sunset\"%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime reel frames 4s%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime reel resolution 720p%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime reel usp on%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime reel teacache 0.3%s\n",
		theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime reel run%s\n",
		theme.DimTextStyle.Render("$"), "")

	printSection("Quick Generate")
	fmt.Printf("    %s anime reel prompt \"Ocean waves\" && anime reel run%s\n",
		theme.DimTextStyle.Render("$"), "")

	// Show current config summary
	printSection("Current Configuration")
	if cfg.Prompt != "" {
		printKeyValue("Prompt", truncate(cfg.Prompt, 40))
	} else {
		printKeyValue("Prompt", theme.DimTextStyle.Render("(not set)"))
	}
	printKeyValue("Frames", fmt.Sprintf("%d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
	printKeyValue("Resolution", cfg.Resolution)

	fmt.Println()
}

// PrintConfig prints the current configuration
func PrintConfig(cfg *ReelConfig) {
	printHeader("Current Reel Configuration")

	printSection("Content")
	if cfg.Prompt != "" {
		printKeyValue("Prompt", cfg.Prompt)
	} else {
		printKeyValue("Prompt", theme.DimTextStyle.Render("(not set)"))
	}
	if cfg.Image != "" {
		printKeyValue("Image", cfg.Image)
	}

	printSection("Generation")
	printKeyValue("Script", cfg.Script)
	printKeyValue("Model", cfg.ModelID)
	printKeyValue("Resolution", cfg.Resolution)
	printKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
	printKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))
	printKeyValue("Duration", fmt.Sprintf("%.1f seconds", float64(cfg.NumFrames)/float64(cfg.FPS)))

	printSection("Parameters")
	printKeyValue("Inference Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
	printKeyValue("Guidance Scale", fmt.Sprintf("%.1f", cfg.GuidanceScale))
	printKeyValue("Shift", fmt.Sprintf("%.1f", cfg.Shift))
	if cfg.Seed >= 0 {
		printKeyValue("Seed", fmt.Sprintf("%d", cfg.Seed))
	} else {
		printKeyValue("Seed", "random")
	}

	printSection("Optimization")
	printKeyValue("USP (Multi-GPU)", fmt.Sprintf("%v", cfg.UseUSP))
	printKeyValue("Offload", fmt.Sprintf("%v", cfg.Offload))
	printKeyValue("TeaCache", fmt.Sprintf("%v", cfg.TeaCache))
	if cfg.TeaCache {
		printKeyValue("TeaCache Thresh", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))
	}

	printSection("Output")
	printKeyValue("Directory", cfg.OutDir)

	fmt.Println()
}

// ResetConfig resets configuration to defaults
func ResetConfig() {
	cfg := DefaultConfig()
	cfg.Save()
	printStatus("success", "Configuration reset to defaults")
}

func printSubcommand(name, desc string) {
	fmt.Printf("    %s  %s\n",
		theme.InfoStyle.Render(fmt.Sprintf("%-14s", name)),
		theme.DimTextStyle.Render(desc))
}
