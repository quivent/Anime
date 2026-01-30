package reel

import (
    "fmt"

    "github.com/sky-cli/sky/ui"
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
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Run 'sky reel help' to see available subcommands",
            "Common: prompt, frames, resolution, run",
        })
    }
}

// PrintHelp prints the reel help
func PrintHelp(cfg *ReelConfig) {
    ui.PrintHeader("SkyReel Video Generation")

    fmt.Printf("  %s\n\n", ui.Muted("Subcommand-based interface for SkyReels-V2"))

    ui.PrintSection("Usage")
    fmt.Printf("  %s <subcommand> [value]\n\n", ui.Value("sky reel"))

    ui.PrintSection("Configuration Subcommands")
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "prompt", ui.Reset, ui.Muted("Set generation prompt"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "frames", ui.Reset, ui.Muted("Set frame count (97, 4s, etc)"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "resolution", ui.Reset, ui.Muted("Set resolution (540p, 720p)"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "steps", ui.Reset, ui.Muted("Set inference steps"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "guidance", ui.Reset, ui.Muted("Set CFG guidance scale"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "seed", ui.Reset, ui.Muted("Set random seed"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "output", ui.Reset, ui.Muted("Set output directory"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "model", ui.Reset, ui.Muted("Select model variant"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "image", ui.Reset, ui.Muted("Set input image (for i2v)"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "fps", ui.Reset, ui.Muted("Set frames per second"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "script", ui.Reset, ui.Muted("Select generation script"))

    ui.PrintSection("Optimization Subcommands")
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "usp", ui.Reset, ui.Muted("Toggle multi-GPU parallelism"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "offload", ui.Reset, ui.Muted("Toggle CPU offloading"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "teacache", ui.Reset, ui.Muted("Toggle/configure TeaCache"))

    ui.PrintSection("Execution Subcommands")
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "run", ui.Reset, ui.Muted("Execute generation"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "dry", ui.Reset, ui.Muted("Preview without executing"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "show", ui.Reset, ui.Muted("Show current configuration"))
    fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "reset", ui.Reset, ui.Muted("Reset to defaults"))

    ui.PrintSection("Examples")
    fmt.Printf("    %s sky reel prompt \"A serene lake at sunset\"%s\n", ui.Muted("$"), ui.Reset)
    fmt.Printf("    %s sky reel frames 4s%s\n", ui.Muted("$"), ui.Reset)
    fmt.Printf("    %s sky reel resolution 720p%s\n", ui.Muted("$"), ui.Reset)
    fmt.Printf("    %s sky reel usp on%s\n", ui.Muted("$"), ui.Reset)
    fmt.Printf("    %s sky reel teacache 0.3%s\n", ui.Muted("$"), ui.Reset)
    fmt.Printf("    %s sky reel run%s\n", ui.Muted("$"), ui.Reset)

    ui.PrintSection("Quick Generate")
    fmt.Printf("    %s sky reel prompt \"Ocean waves\" && sky reel run%s\n", ui.Muted("$"), ui.Reset)

    // Show current config summary
    ui.PrintSection("Current Configuration")
    if cfg.Prompt != "" {
        ui.PrintKeyValue("Prompt", truncate(cfg.Prompt, 40))
    } else {
        ui.PrintKeyValue("Prompt", ui.Muted("(not set)"))
    }
    ui.PrintKeyValue("Frames", fmt.Sprintf("%d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
    ui.PrintKeyValue("Resolution", cfg.Resolution)

    fmt.Println()
}

// PrintConfig prints the current configuration
func PrintConfig(cfg *ReelConfig) {
    ui.PrintHeader("Current Reel Configuration")

    ui.PrintSection("Content")
    if cfg.Prompt != "" {
        ui.PrintKeyValue("Prompt", cfg.Prompt)
    } else {
        ui.PrintKeyValue("Prompt", ui.Muted("(not set)"))
    }
    if cfg.Image != "" {
        ui.PrintKeyValue("Image", cfg.Image)
    }

    ui.PrintSection("Generation")
    ui.PrintKeyValue("Script", cfg.Script)
    ui.PrintKeyValue("Model", cfg.ModelID)
    ui.PrintKeyValue("Resolution", cfg.Resolution)
    ui.PrintKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
    ui.PrintKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))
    ui.PrintKeyValue("Duration", fmt.Sprintf("%.1f seconds", float64(cfg.NumFrames)/float64(cfg.FPS)))

    ui.PrintSection("Parameters")
    ui.PrintKeyValue("Inference Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
    ui.PrintKeyValue("Guidance Scale", fmt.Sprintf("%.1f", cfg.GuidanceScale))
    ui.PrintKeyValue("Shift", fmt.Sprintf("%.1f", cfg.Shift))
    if cfg.Seed >= 0 {
        ui.PrintKeyValue("Seed", fmt.Sprintf("%d", cfg.Seed))
    } else {
        ui.PrintKeyValue("Seed", "random")
    }

    ui.PrintSection("Optimization")
    ui.PrintKeyValue("USP (Multi-GPU)", fmt.Sprintf("%v", cfg.UseUSP))
    ui.PrintKeyValue("Offload", fmt.Sprintf("%v", cfg.Offload))
    ui.PrintKeyValue("TeaCache", fmt.Sprintf("%v", cfg.TeaCache))
    if cfg.TeaCache {
        ui.PrintKeyValue("TeaCache Thresh", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))
    }

    ui.PrintSection("Output")
    ui.PrintKeyValue("Directory", cfg.OutDir)

    fmt.Println()
}

// ResetConfig resets configuration to defaults
func ResetConfig() {
    cfg := DefaultConfig()
    cfg.Save()
    ui.PrintStatus("success", "Configuration reset to defaults")
}
