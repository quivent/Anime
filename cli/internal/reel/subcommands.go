package reel

import (
	"fmt"
	"strconv"

	"github.com/joshkornreich/anime/internal/theme"
)

// HandlePrompt handles the prompt subcommand
// Usage: anime reel prompt set "text" | anime reel prompt show | anime reel prompt clear
func HandlePrompt(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		// Show current prompt
		printSection("Prompt")
		if cfg.Prompt == "" {
			printKeyValue("Current", "(not set)")
		} else {
			printKeyValue("Current", cfg.Prompt)
		}

		printSubSection("Usage")
		fmt.Printf("      %s anime reel prompt \"Your text\"%s   %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Set prompt directly"))
		fmt.Printf("      %s anime reel prompt set \"...\"%s     %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Set prompt explicitly"))
		fmt.Printf("      %s anime reel prompt show%s          %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Print prompt only"))
		fmt.Printf("      %s anime reel prompt clear%s         %s\n",
			theme.DimTextStyle.Render("$"), "", theme.DimTextStyle.Render("Clear prompt"))
		return
	}

	switch args[0] {
	case "set":
		if len(args) < 2 {
			printSuggestion("Prompt text required", []string{
				"anime reel prompt set \"Your prompt here\"",
			})
			return
		}
		cfg.Prompt = args[1]
		cfg.Save()
		printStatus("success", "Prompt set")
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(cfg.Prompt))

	case "show":
		if cfg.Prompt == "" {
			printStatus("pending", "No prompt set")
		} else {
			fmt.Printf("%s\n", cfg.Prompt)
		}

	case "clear":
		cfg.Prompt = ""
		cfg.Save()
		printStatus("success", "Prompt cleared")

	default:
		// Treat as the prompt itself
		cfg.Prompt = args[0]
		cfg.Save()
		printStatus("success", "Prompt set")
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(cfg.Prompt))
	}
}

// HandleFrames handles the frames subcommand
// Usage: anime reel frames 97 | anime reel frames 2s | anime reel frames 4s
func HandleFrames(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Frame Configuration")
		printKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
		printKeyValue("Duration", fmt.Sprintf("%.1f seconds", float64(cfg.NumFrames)/float64(cfg.FPS)))
		printKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))

		printSubSection("Presets")
		printPreset("2s", "49 frames")
		printPreset("4s", "97 frames (default)")
		printPreset("8s", "193 frames")
		printPreset("12s", "289 frames")
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
			printSuggestion("Invalid frame count: "+input, []string{
				"Use a number: anime reel frames 97",
				"Or a preset: anime reel frames 4s",
			})
			return
		}
		cfg.NumFrames = n
	}

	cfg.Save()
	printStatus("success", fmt.Sprintf("Frames set to %d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
}

// HandleResolution handles the resolution subcommand
// Usage: anime reel resolution 540p | anime reel resolution 720p
func HandleResolution(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Resolution")
		printKeyValue("Current", cfg.Resolution)

		printSubSection("Available")
		printPreset("540p", "544x960 (faster)")
		printPreset("720p", "720x1280 (higher quality)")
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
		printSuggestion("Invalid resolution: "+res, []string{
			"Available: 540p, 720p",
		})
		return
	}

	cfg.Save()
	printStatus("success", "Resolution set to "+cfg.Resolution)
}

// HandleSteps handles the inference steps subcommand
// Usage: anime reel steps 30 | anime reel steps fast | anime reel steps quality
func HandleSteps(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Inference Steps")
		printKeyValue("Current", fmt.Sprintf("%d", cfg.InferenceSteps))

		printSubSection("Presets")
		printPreset("fast", "20 steps (faster, lower quality)")
		printPreset("default", "30 steps (balanced)")
		printPreset("quality", "50 steps (slower, higher quality)")
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
			printSuggestion("Invalid steps: "+input, []string{
				"Use a number: anime reel steps 30",
				"Or a preset: fast, default, quality",
			})
			return
		}
		cfg.InferenceSteps = n
	}

	cfg.Save()
	printStatus("success", fmt.Sprintf("Inference steps set to %d", cfg.InferenceSteps))
}

// HandleGuidance handles the guidance scale subcommand
// Usage: anime reel guidance 6.0 | anime reel guidance low | anime reel guidance high
func HandleGuidance(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Guidance Scale (CFG)")
		printKeyValue("Current", fmt.Sprintf("%.1f", cfg.GuidanceScale))

		printSubSection("Presets")
		printPreset("low", "3.0 (more creative)")
		printPreset("default", "6.0 (balanced)")
		printPreset("high", "9.0 (stronger prompt adherence)")
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
			printSuggestion("Invalid guidance: "+input, []string{
				"Use a number: anime reel guidance 6.0",
				"Or a preset: low, default, high",
			})
			return
		}
		cfg.GuidanceScale = f
	}

	cfg.Save()
	printStatus("success", fmt.Sprintf("Guidance scale set to %.1f", cfg.GuidanceScale))
}

// HandleSeed handles the seed subcommand
// Usage: anime reel seed 42 | anime reel seed random
func HandleSeed(cfg *ReelConfig, args []string) {
	if len(args) == 0 {
		printSection("Random Seed")
		if cfg.Seed < 0 {
			printKeyValue("Current", "random (not fixed)")
		} else {
			printKeyValue("Current", fmt.Sprintf("%d", cfg.Seed))
		}

		printSubSection("Options")
		printPreset("<number>", "Fixed seed for reproducibility")
		printPreset("random", "Clear seed (use random each time)")
		printStatus("info", "Fixed seed required when using USP (multi-GPU)")
		return
	}

	input := args[0]
	switch input {
	case "random", "none", "clear":
		cfg.Seed = -1
		cfg.Save()
		printStatus("success", "Seed cleared (will use random)")
	default:
		n, err := strconv.Atoi(input)
		if err != nil || n < 0 {
			printSuggestion("Invalid seed: "+input, []string{
				"Use a positive number: anime reel seed 42",
				"Or 'random' to clear: anime reel seed random",
			})
			return
		}
		cfg.Seed = n
		cfg.Save()
		printStatus("success", fmt.Sprintf("Seed set to %d", cfg.Seed))
	}
}

// Helper to print preset options
func printPreset(name, desc string) {
	fmt.Printf("      %s  %s\n",
		theme.InfoStyle.Render(fmt.Sprintf("%-10s", name)),
		theme.DimTextStyle.Render(desc))
}
