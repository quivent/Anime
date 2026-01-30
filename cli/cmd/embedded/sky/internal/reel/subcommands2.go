package reel

import (
    "fmt"
    "strconv"

    "github.com/sky-cli/sky/ui"
)

// HandleOutput handles the output directory subcommand
// Usage: sky reel output ./my_videos | sky reel output show
func HandleOutput(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Output Directory")
        ui.PrintKeyValue("Current", cfg.OutDir)

        ui.PrintSubSection("Usage")
        fmt.Printf("    %s sky reel output ./videos%s   %s\n", ui.Muted("$"), ui.Reset, "Set output directory")
        fmt.Printf("    %s sky reel output show%s      %s\n", ui.Muted("$"), ui.Reset, "Print path only")
        ui.PrintStatus("info", "Videos saved to: result/<outdir>/")
        return
    }

    if args[0] == "show" {
        fmt.Println(cfg.OutDir)
        return
    }

    cfg.OutDir = args[0]
    cfg.Save()
    ui.PrintStatus("success", "Output directory set to "+cfg.OutDir)
}

// HandleModel handles the model selection subcommand
// Usage: sky reel model t2v-14b | sky reel model i2v-14b | sky reel model list
func HandleModel(cfg *ReelConfig, args []string) {
    models := map[string]string{
        "t2v-14b-540p": "Skywork/SkyReels-V2-T2V-14B-540P",
        "t2v-14b-720p": "Skywork/SkyReels-V2-T2V-14B-720P",
        "i2v-1.3b":     "Skywork/SkyReels-V2-I2V-1.3B-540P",
        "i2v-14b-540p": "Skywork/SkyReels-V2-I2V-14B-540P",
        "i2v-14b-720p": "Skywork/SkyReels-V2-I2V-14B-720P",
    }

    if len(args) == 0 || args[0] == "list" {
        ui.PrintSection("Model Selection")
        ui.PrintKeyValue("Current", cfg.ModelID)

        ui.PrintSubSection("Text-to-Video")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "t2v-14b-540p", ui.Reset, "14B model, 540P")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "t2v-14b-720p", ui.Reset, "14B model, 720P")

        ui.PrintSubSection("Image-to-Video")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-1.3b", ui.Reset, "1.3B model, 540P (fast)")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-14b-540p", ui.Reset, "14B model, 540P")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-14b-720p", ui.Reset, "14B model, 720P")
        return
    }

    modelKey := args[0]
    if fullID, ok := models[modelKey]; ok {
        cfg.ModelID = fullID
        cfg.Save()
        ui.PrintStatus("success", "Model set to "+fullID)
    } else {
        // Allow full model ID
        cfg.ModelID = modelKey
        cfg.Save()
        ui.PrintStatus("success", "Model set to "+modelKey)
    }
}

// HandleImage handles the image input subcommand (for i2v)
// Usage: sky reel image ./input.png | sky reel image clear
func HandleImage(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Image Input (for Image-to-Video)")
        if cfg.Image == "" {
            ui.PrintKeyValue("Current", "none (text-to-video mode)")
        } else {
            ui.PrintKeyValue("Current", cfg.Image)
        }

        ui.PrintSubSection("Usage")
        fmt.Printf("    %s sky reel image ./photo.png%s  %s\n", ui.Muted("$"), ui.Reset, "Set input image")
        fmt.Printf("    %s sky reel image clear%s        %s\n", ui.Muted("$"), ui.Reset, "Clear (use text-to-video)")

        ui.PrintSubSection("Compatible Models (I2V)")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-1.3b", ui.Reset, "Fast, 540P")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-14b-540p", ui.Reset, "Quality, 540P")
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, "i2v-14b-720p", ui.Reset, "Quality, 720P")
        return
    }

    if args[0] == "clear" || args[0] == "none" {
        cfg.Image = ""
        cfg.Save()
        ui.PrintStatus("success", "Image cleared (text-to-video mode)")
        return
    }

    cfg.Image = args[0]
    cfg.Save()
    ui.PrintStatus("success", "Image set to "+cfg.Image)
    ui.PrintStatus("info", "Using image-to-video mode")
}

// HandleUSP handles the multi-GPU USP subcommand
// Usage: sky reel usp on | sky reel usp off
func HandleUSP(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("USP (Multi-GPU Parallelism)")
        ui.PrintKeyValue("Status", fmt.Sprintf("%v", cfg.UseUSP))

        ui.PrintSubSection("Usage")
        fmt.Printf("    %s sky reel usp on%s   %s\n", ui.Muted("$"), ui.Reset, "Enable multi-GPU")
        fmt.Printf("    %s sky reel usp off%s  %s\n", ui.Muted("$"), ui.Reset, "Disable (single GPU)")
        return
    }

    switch args[0] {
    case "on", "enable", "true", "1":
        cfg.UseUSP = true
        cfg.Save()
        ui.PrintStatus("success", "USP enabled (multi-GPU parallelism)")
    case "off", "disable", "false", "0":
        cfg.UseUSP = false
        cfg.Save()
        ui.PrintStatus("success", "USP disabled (single GPU)")
    default:
        ui.PrintSuggestion("Invalid value: "+args[0], []string{
            "Use: on, off",
        })
    }
}

// HandleOffload handles CPU offloading subcommand
// Usage: sky reel offload on | sky reel offload off
func HandleOffload(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("CPU Offloading")
        ui.PrintKeyValue("Status", fmt.Sprintf("%v", cfg.Offload))

        ui.PrintSubSection("Options")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "on", ui.Reset, "Enable CPU offloading (saves GPU memory)")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "off", ui.Reset, "Disable (faster, needs more VRAM)")
        ui.PrintStatus("info", "Use when GPU memory is limited")
        return
    }

    switch args[0] {
    case "on", "enable", "true", "1":
        cfg.Offload = true
        cfg.Save()
        ui.PrintStatus("success", "CPU offloading enabled")
    case "off", "disable", "false", "0":
        cfg.Offload = false
        cfg.Save()
        ui.PrintStatus("success", "CPU offloading disabled")
    default:
        ui.PrintSuggestion("Invalid value: "+args[0], []string{
            "Use: on, off",
        })
    }
}

// HandleTeaCache handles TeaCache optimization subcommand
// Usage: sky reel teacache on | sky reel teacache off | sky reel teacache 0.3
func HandleTeaCache(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("TeaCache Optimization")
        ui.PrintKeyValue("Status", fmt.Sprintf("%v", cfg.TeaCache))
        ui.PrintKeyValue("Threshold", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))

        ui.PrintSubSection("Threshold Guide")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "0.2", ui.Reset, "Conservative (higher quality)")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "0.3", ui.Reset, "Balanced (default)")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "0.5", ui.Reset, "Aggressive (faster)")
        return
    }

    switch args[0] {
    case "on", "enable", "true", "1":
        cfg.TeaCache = true
        cfg.Save()
        ui.PrintStatus("success", fmt.Sprintf("TeaCache enabled (threshold=%.2f)", cfg.TeaCacheThresh))
    case "off", "disable", "false", "0":
        cfg.TeaCache = false
        cfg.Save()
        ui.PrintStatus("success", "TeaCache disabled")
    default:
        // Try to parse as threshold
        f, err := strconv.ParseFloat(args[0], 64)
        if err != nil || f < 0 || f > 1 {
            ui.PrintSuggestion("Invalid value: "+args[0], []string{
                "Use: on, off, or a threshold (0.0-1.0)",
                "Example: sky reel teacache 0.3",
            })
            return
        }
        cfg.TeaCache = true
        cfg.TeaCacheThresh = f
        cfg.Save()
        ui.PrintStatus("success", fmt.Sprintf("TeaCache enabled with threshold %.2f", f))
    }
}

// HandleScript handles script selection subcommand
// Usage: sky reel script df | sky reel script sequential
func HandleScript(cfg *ReelConfig, args []string) {
    scripts := map[string]string{
        "default":    "generate_video.py",
        "standard":   "generate_video.py",
        "df":         "generate_video_df.py",
        "forcing":    "generate_video_df.py",
        "seq":        "generate_video_sequential.py",
        "sequential": "generate_video_sequential.py",
    }

    if len(args) == 0 || args[0] == "list" {
        ui.PrintSection("Generation Script")
        ui.PrintKeyValue("Current", cfg.Script)

        ui.PrintSubSection("Available")
        fmt.Printf("    %s%-12s%s %s\n", ui.BrightCyan, "default", ui.Reset, "generate_video.py (standard)")
        fmt.Printf("    %s%-12s%s %s\n", ui.BrightCyan, "df", ui.Reset, "generate_video_df.py (diffusion forcing)")
        fmt.Printf("    %s%-12s%s %s\n", ui.BrightCyan, "sequential", ui.Reset, "generate_video_sequential.py (continuity)")
        return
    }

    scriptKey := args[0]
    if script, ok := scripts[scriptKey]; ok {
        cfg.Script = script
        cfg.Save()
        ui.PrintStatus("success", "Script set to "+script)
    } else {
        ui.PrintSuggestion("Unknown script: "+scriptKey, []string{
            "Available: default, df, sequential",
        })
    }
}

// HandleFPS handles FPS subcommand
// Usage: sky reel fps 24 | sky reel fps 30
func HandleFPS(cfg *ReelConfig, args []string) {
    if len(args) == 0 {
        ui.PrintSection("Frames Per Second")
        ui.PrintKeyValue("Current", fmt.Sprintf("%d", cfg.FPS))

        ui.PrintSubSection("Common Values")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "24", ui.Reset, "Cinema standard (default)")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "30", ui.Reset, "Web/streaming standard")
        fmt.Printf("    %s%-6s%s %s\n", ui.BrightCyan, "60", ui.Reset, "Smooth motion")
        ui.PrintStatus("info", "Higher FPS = shorter video duration for same frame count")
        return
    }

    n, err := strconv.Atoi(args[0])
    if err != nil || n < 1 || n > 60 {
        ui.PrintSuggestion("Invalid FPS: "+args[0], []string{
            "Use a number between 1 and 60",
            "Common values: 24, 30",
        })
        return
    }

    cfg.FPS = n
    cfg.Save()
    ui.PrintStatus("success", fmt.Sprintf("FPS set to %d", cfg.FPS))
}
