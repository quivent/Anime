package reel

import (
	"fmt"
	"strconv"

	"github.com/joshkornreich/anime/internal/theme"
)

// HandleOutput handles the output directory subcommand
// Usage: anime reel output ./my_videos | anime reel output show
func HandleOutput(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Output Directory")
		printKeyValue("Current", cfg.OutDir)

		printSubSection("Usage")
		fmt.Printf("      %s anime reel output ./videos%s   %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Set output directory"))
		fmt.Printf("      %s anime reel output show%s      %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Print path only"))
		printStatus("info", "Videos saved to: result/<outdir>/")
		return
	}

	if args[0] == "show" {
		fmt.Println(cfg.OutDir)
		return
	}

	cfg.OutDir = args[0]
	cfg.Save()
	printStatus("success", "Output directory set to "+cfg.OutDir)
}

// HandleModel handles the model selection subcommand
// Usage: anime reel model t2v-14b | anime reel model i2v-14b | anime reel model list
func HandleModel(cfg *ReelConfig, args []string) {
	models := map[string]string{
		"t2v-14b-540p": "Skywork/SkyReels-V2-T2V-14B-540P",
		"t2v-14b-720p": "Skywork/SkyReels-V2-T2V-14B-720P",
		"i2v-1.3b":     "Skywork/SkyReels-V2-I2V-1.3B-540P",
		"i2v-14b-540p": "Skywork/SkyReels-V2-I2V-14B-540P",
		"i2v-14b-720p": "Skywork/SkyReels-V2-I2V-14B-720P",
	}

	if len(args) == 0 || args[0] == "list" {
		printSection("Model Selection")
		printKeyValue("Current", cfg.ModelID)

		printSubSection("Text-to-Video")
		printPreset("t2v-14b-540p", "14B model, 540P")
		printPreset("t2v-14b-720p", "14B model, 720P")

		printSubSection("Image-to-Video")
		printPreset("i2v-1.3b", "1.3B model, 540P (fast)")
		printPreset("i2v-14b-540p", "14B model, 540P")
		printPreset("i2v-14b-720p", "14B model, 720P")
		return
	}

	modelKey := args[0]
	if fullID, ok := models[modelKey]; ok {
		cfg.ModelID = fullID
		cfg.Save()
		printStatus("success", "Model set to "+fullID)
	} else {
		// Allow full model ID
		cfg.ModelID = modelKey
		cfg.Save()
		printStatus("success", "Model set to "+modelKey)
	}
}

// HandleImage handles the image input subcommand (for i2v)
// Usage: anime reel image ./input.png | anime reel image clear
func HandleImage(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Image Input (for Image-to-Video)")
		if cfg.Image == "" {
			printKeyValue("Current", "none (text-to-video mode)")
		} else {
			printKeyValue("Current", cfg.Image)
		}

		printSubSection("Usage")
		fmt.Printf("      %s anime reel image ./photo.png%s  %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Set input image"))
		fmt.Printf("      %s anime reel image clear%s        %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Clear (use text-to-video)"))

		printSubSection("Compatible Models (I2V)")
		printPreset("i2v-1.3b", "Fast, 540P")
		printPreset("i2v-14b-540p", "Quality, 540P")
		printPreset("i2v-14b-720p", "Quality, 720P")
		return
	}

	if args[0] == "clear" || args[0] == "none" {
		cfg.Image = ""
		cfg.Save()
		printStatus("success", "Image cleared (text-to-video mode)")
		return
	}

	cfg.Image = args[0]
	cfg.Save()
	printStatus("success", "Image set to "+cfg.Image)
	printStatus("info", "Using image-to-video mode")
}

// HandleUSP handles the multi-GPU USP subcommand
// Usage: anime reel usp on | anime reel usp off
func HandleUSP(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("USP (Multi-GPU Parallelism)")
		printKeyValue("Status", fmt.Sprintf("%v", cfg.UseUSP))

		printSubSection("Usage")
		fmt.Printf("      %s anime reel usp on%s   %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Enable multi-GPU"))
		fmt.Printf("      %s anime reel usp off%s  %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Disable (single GPU)"))
		return
	}

	switch args[0] {
	case "on", "enable", "true", "1":
		cfg.UseUSP = true
		cfg.Save()
		printStatus("success", "USP enabled (multi-GPU parallelism)")
	case "off", "disable", "false", "0":
		cfg.UseUSP = false
		cfg.Save()
		printStatus("success", "USP disabled (single GPU)")
	default:
		printSuggestion("Invalid value: "+args[0], []string{
			"Use: on, off",
		})
	}
}

// HandleOffload handles CPU offloading subcommand
// Usage: anime reel offload on | anime reel offload off
func HandleOffload(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("CPU Offloading")
		printKeyValue("Status", fmt.Sprintf("%v", cfg.Offload))

		printSubSection("Options")
		printPreset("on", "Enable CPU offloading (saves GPU memory)")
		printPreset("off", "Disable (faster, needs more VRAM)")
		printStatus("info", "Use when GPU memory is limited")
		return
	}

	switch args[0] {
	case "on", "enable", "true", "1":
		cfg.Offload = true
		cfg.Save()
		printStatus("success", "CPU offloading enabled")
	case "off", "disable", "false", "0":
		cfg.Offload = false
		cfg.Save()
		printStatus("success", "CPU offloading disabled")
	default:
		printSuggestion("Invalid value: "+args[0], []string{
			"Use: on, off",
		})
	}
}

// HandleTeaCache handles TeaCache optimization subcommand
// Usage: anime reel teacache on | anime reel teacache off | anime reel teacache 0.3
func HandleTeaCache(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("TeaCache Optimization")
		printKeyValue("Status", fmt.Sprintf("%v", cfg.TeaCache))
		printKeyValue("Threshold", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))

		printSubSection("Threshold Guide")
		printPreset("0.2", "Conservative (higher quality)")
		printPreset("0.3", "Balanced (default)")
		printPreset("0.5", "Aggressive (faster)")
		return
	}

	switch args[0] {
	case "on", "enable", "true", "1":
		cfg.TeaCache = true
		cfg.Save()
		printStatus("success", fmt.Sprintf("TeaCache enabled (threshold=%.2f)", cfg.TeaCacheThresh))
	case "off", "disable", "false", "0":
		cfg.TeaCache = false
		cfg.Save()
		printStatus("success", "TeaCache disabled")
	default:
		// Try to parse as threshold
		f, err := strconv.ParseFloat(args[0], 64)
		if err != nil || f < 0 || f > 1 {
			printSuggestion("Invalid value: "+args[0], []string{
				"Use: on, off, or a threshold (0.0-1.0)",
				"Example: anime reel teacache 0.3",
			})
			return
		}
		cfg.TeaCache = true
		cfg.TeaCacheThresh = f
		cfg.Save()
		printStatus("success", fmt.Sprintf("TeaCache enabled with threshold %.2f", f))
	}
}

// HandleScript handles script selection subcommand
// Usage: anime reel script df | anime reel script sequential
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
		printSection("Generation Script")
		printKeyValue("Current", cfg.Script)

		printSubSection("Available")
		printPreset("default", "generate_video.py (standard)")
		printPreset("df", "generate_video_df.py (diffusion forcing)")
		printPreset("sequential", "generate_video_sequential.py (continuity)")
		return
	}

	scriptKey := args[0]
	if script, ok := scripts[scriptKey]; ok {
		cfg.Script = script
		cfg.Save()
		printStatus("success", "Script set to "+script)
	} else {
		printSuggestion("Unknown script: "+scriptKey, []string{
			"Available: default, df, sequential",
		})
	}
}

// HandleFPS handles FPS subcommand
// Usage: anime reel fps 24 | anime reel fps 30
func HandleFPS(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Frames Per Second")
		printKeyValue("Current", fmt.Sprintf("%d", cfg.FPS))

		printSubSection("Common Values")
		printPreset("24", "Cinema standard (default)")
		printPreset("30", "Web/streaming standard")
		printPreset("60", "Smooth motion")
		printStatus("info", "Higher FPS = shorter video duration for same frame count")
		return
	}

	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 || n > 60 {
		printSuggestion("Invalid FPS: "+args[0], []string{
			"Use a number between 1 and 60",
			"Common values: 24, 30",
		})
		return
	}

	cfg.FPS = n
	cfg.Save()
	printStatus("success", fmt.Sprintf("FPS set to %d", cfg.FPS))
}
