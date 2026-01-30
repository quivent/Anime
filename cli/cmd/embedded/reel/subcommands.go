package reel

import (
    "fmt"
    "strconv"

    "github.com/sky-cli/sky/ui"
)

// HandlePrompt handles the prompt subcommand
// Usage: sky reel prompt set "text" | sky reel prompt show | sky reel prompt clear
func HandlePrompt(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        // Show current prompt
        ui.PrintSection("Prompt")
        if cfg.Prompt == "" {
            ui.PrintKeyValue("Current", "(not set)")
        } else {
            ui.PrintKeyValue("Current", cfg.Prompt)
        }

        ui.PrintSubSection("Usage")
        fmt.Printf("    %s sky reel prompt \"Your text\"%s   %s\n", ui.Muted("$"), ui.Reset, "Set prompt directly")
        fmt.Printf("    %s sky reel prompt set \"...\"%s     %s\n", ui.Muted("$"), ui.Reset, "Set prompt explicitly")
        fmt.Printf("    %s sky reel prompt show%s          %s\n", ui.Muted("$"), ui.Reset, "Print prompt only")
        fmt.Printf("    %s sky reel prompt clear%s         %s\n", ui.Muted("$"), ui.Reset, "Clear prompt")
        return
    }

    switch args[0] {
    case "set":
        if len(args) < 2 {
            ui.PrintSuggestion("Prompt text required", []string{
                "sky reel prompt set \"Your prompt here\"",
            })
            return
        }
        cfg.Prompt = args[1]
        cfg.Save()
        ui.PrintStatus("success", "Prompt set")
        fmt.Printf("  %s\n", ui.Value(cfg.Prompt))

    case "show":
        if cfg.Prompt == "" {
            ui.PrintStatus("pending", "No prompt set")
        } else {
            fmt.Printf("%s\n", cfg.Prompt)
        }

    case "clear":
        cfg.Prompt = ""
        cfg.Save()
        ui.PrintStatus("success", "Prompt cleared")

    default:
        // Treat as the prompt itself
        cfg.Prompt = args[0]
        cfg.Save()
        ui.PrintStatus("success", "Prompt set")
        fmt.Printf("  %s\n", ui.Value(cfg.Prompt))
    }
}

// HandleFrames handles the frames subcommand
// Usage: sky reel frames 97 | sky reel frames 2s | sky reel frames 4s
func HandleFrames(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Frame Configuration")
        ui.PrintKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
        ui.PrintKeyValue("Duration", fmt.Sprintf("%.1f seconds", float64(cfg.NumFrames)/float64(cfg.FPS)))
        ui.PrintKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))

        ui.PrintSubSection("Presets")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "2s", ui.Reset, "49 frames")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "4s", ui.Reset, "97 frames (default)")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "8s", ui.Reset, "193 frames")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "12s", ui.Reset, "289 frames")
        return
    }

    input := args[0]

    // Check for duration presets
    switch input {
    case "2s":
        cfg.NumFrames = 49
    case "4s":
        cfg.NumFrames = 97
    case "8s":
        cfg.NumFrames = 193
    case "12s":
        cfg.NumFrames = 289
    default:
        // Try to parse as number
        n, err := strconv.Atoi(input)
        if err != nil {
            ui.PrintSuggestion("Invalid frame count: "+input, []string{
                "Use a number: sky reel frames 97",
                "Or a preset: sky reel frames 4s",
            })
            return
        }
        cfg.NumFrames = n
    }

    cfg.Save()
    ui.PrintStatus("success", fmt.Sprintf("Frames set to %d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
}

// HandleResolution handles the resolution subcommand
// Usage: sky reel resolution 540p | sky reel resolution 720p
func HandleResolution(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Resolution")
        ui.PrintKeyValue("Current", cfg.Resolution)

        ui.PrintSubSection("Available")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "540p", ui.Reset, "544×960 (faster)")
        fmt.Printf("    %s%-8s%s %s\n", ui.BrightCyan, "720p", ui.Reset, "720×1280 (higher quality)")
        return
    }

    res := args[0]
    switch res {
    case "540p", "540P":
        cfg.Resolution = "540P"
        cfg.ModelID = "Skywork/SkyReels-V2-T2V-14B-540P"
    case "720p", "720P":
        cfg.Resolution = "720P"
        cfg.ModelID = "Skywork/SkyReels-V2-T2V-14B-720P"
    default:
        ui.PrintSuggestion("Invalid resolution: "+res, []string{
            "Available: 540p, 720p",
        })
        return
    }

    cfg.Save()
    ui.PrintStatus("success", "Resolution set to "+cfg.Resolution)
}

// HandleSteps handles the inference steps subcommand
// Usage: sky reel steps 30 | sky reel steps fast | sky reel steps quality
func HandleSteps(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Inference Steps")
        ui.PrintKeyValue("Current", fmt.Sprintf("%d", cfg.InferenceSteps))

        ui.PrintSubSection("Presets")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "fast", ui.Reset, "20 steps (faster, lower quality)")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "default", ui.Reset, "30 steps (balanced)")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "quality", ui.Reset, "50 steps (slower, higher quality)")
        return
    }

    input := args[0]
    switch input {
    case "fast":
        cfg.InferenceSteps = 20
    case "default", "balanced":
        cfg.InferenceSteps = 30
    case "quality", "high":
        cfg.InferenceSteps = 50
    default:
        n, err := strconv.Atoi(input)
        if err != nil || n < 1 {
            ui.PrintSuggestion("Invalid steps: "+input, []string{
                "Use a number: sky reel steps 30",
                "Or a preset: fast, default, quality",
            })
            return
        }
        cfg.InferenceSteps = n
    }

    cfg.Save()
    ui.PrintStatus("success", fmt.Sprintf("Inference steps set to %d", cfg.InferenceSteps))
}

// HandleGuidance handles the guidance scale subcommand
// Usage: sky reel guidance 6.0 | sky reel guidance low | sky reel guidance high
func HandleGuidance(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Guidance Scale (CFG)")
        ui.PrintKeyValue("Current", fmt.Sprintf("%.1f", cfg.GuidanceScale))

        ui.PrintSubSection("Presets")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "low", ui.Reset, "3.0 (more creative)")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "default", ui.Reset, "6.0 (balanced)")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "high", ui.Reset, "9.0 (stronger prompt adherence)")
        return
    }

    input := args[0]
    switch input {
    case "low":
        cfg.GuidanceScale = 3.0
    case "default", "balanced":
        cfg.GuidanceScale = 6.0
    case "high":
        cfg.GuidanceScale = 9.0
    default:
        f, err := strconv.ParseFloat(input, 64)
        if err != nil || f < 0 {
            ui.PrintSuggestion("Invalid guidance: "+input, []string{
                "Use a number: sky reel guidance 6.0",
                "Or a preset: low, default, high",
            })
            return
        }
        cfg.GuidanceScale = f
    }

    cfg.Save()
    ui.PrintStatus("success", fmt.Sprintf("Guidance scale set to %.1f", cfg.GuidanceScale))
}

// HandleSeed handles the seed subcommand
// Usage: sky reel seed 42 | sky reel seed random
func HandleSeed(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Random Seed")
        if cfg.Seed < 0 {
            ui.PrintKeyValue("Current", "random (not fixed)")
        } else {
            ui.PrintKeyValue("Current", fmt.Sprintf("%d", cfg.Seed))
        }

        ui.PrintSubSection("Options")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "<number>", ui.Reset, "Fixed seed for reproducibility")
        fmt.Printf("    %s%-10s%s %s\n", ui.BrightCyan, "random", ui.Reset, "Clear seed (use random each time)")
        ui.PrintStatus("info", "Fixed seed required when using USP (multi-GPU)")
        return
    }

    input := args[0]
    switch input {
    case "random", "none", "clear":
        cfg.Seed = -1
        cfg.Save()
        ui.PrintStatus("success", "Seed cleared (will use random)")
    default:
        n, err := strconv.Atoi(input)
        if err != nil || n < 0 {
            ui.PrintSuggestion("Invalid seed: "+input, []string{
                "Use a positive number: sky reel seed 42",
                "Or 'random' to clear: sky reel seed random",
            })
            return
        }
        cfg.Seed = n
        cfg.Save()
        ui.PrintStatus("success", fmt.Sprintf("Seed set to %d", cfg.Seed))
    }
}
