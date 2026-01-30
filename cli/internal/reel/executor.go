package reel

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/joshkornreich/anime/internal/theme"
)

// Execute runs the configured generation
func Execute(cfg *ReelConfig) {
	// Check SkyReels directory
	if _, err := os.Stat(SkyReelsDir); os.IsNotExist(err) {
		printSuggestion("SkyReels-V2 not found", []string{
			"Expected location: " + SkyReelsDir,
			"Clone: git clone https://github.com/SkyworkAI/SkyReels-V2.git",
		})
		return
	}

	// Validate prompt
	if cfg.Prompt == "" {
		printSuggestion("No prompt specified", []string{
			"Set prompt: anime reel prompt \"Your prompt here\"",
		})
		return
	}

	scriptPath := SkyReelsDir + "/" + cfg.Script

	// Check script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		printSuggestion("Script not found: "+cfg.Script, []string{
			"Available: generate_video.py, generate_video_df.py, generate_video_sequential.py",
		})
		return
	}

	// Find python
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		pythonPath, err = exec.LookPath("python")
		if err != nil {
			printSuggestion("Python not found", []string{
				"Install Python 3 or activate your virtual environment",
			})
			return
		}
	}

	// Build args
	args := cfg.ToArgs()

	// Print configuration
	printHeader("SkyReel Generation")

	printSection("Configuration")
	printKeyValue("Script", cfg.Script)
	printKeyValue("Prompt", truncate(cfg.Prompt, 50))
	printKeyValue("Frames", fmt.Sprintf("%d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
	printKeyValue("Resolution", cfg.Resolution)
	printKeyValue("Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
	printKeyValue("Guidance", fmt.Sprintf("%.1f", cfg.GuidanceScale))

	if cfg.UseUSP {
		printKeyValue("Parallelism", "USP enabled (multi-GPU)")
	}
	if cfg.TeaCache {
		printKeyValue("TeaCache", fmt.Sprintf("enabled (thresh=%.2f)", cfg.TeaCacheThresh))
	}
	if cfg.Offload {
		printKeyValue("Offload", "enabled")
	}

	printSection("Executing")
	printKeyValue("Command", "python3 "+cfg.Script)
	fmt.Println()

	// Change to SkyReels directory
	if err := os.Chdir(SkyReelsDir); err != nil {
		printStatus("error", "Failed to change directory: "+err.Error())
		return
	}

	// Build full command
	cmdArgs := append([]string{pythonPath, scriptPath}, args...)
	env := os.Environ()

	// Execute, replacing current process
	if err := syscall.Exec(pythonPath, cmdArgs, env); err != nil {
		printStatus("error", "Execution failed: "+err.Error())
	}
}

// ExecuteDry shows what would be executed without running
func ExecuteDry(cfg *ReelConfig) {
	printHeader("SkyReel Generation (Dry Run)")

	printSection("Configuration")
	printKeyValue("Script", cfg.Script)
	printKeyValue("Prompt", cfg.Prompt)
	printKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
	printKeyValue("Resolution", cfg.Resolution)
	printKeyValue("Model", cfg.ModelID)
	printKeyValue("Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
	printKeyValue("Guidance", fmt.Sprintf("%.1f", cfg.GuidanceScale))
	printKeyValue("Shift", fmt.Sprintf("%.1f", cfg.Shift))
	printKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))
	printKeyValue("Seed", fmt.Sprintf("%d", cfg.Seed))
	printKeyValue("Output", cfg.OutDir)

	printSection("Optimizations")
	printKeyValue("USP (multi-GPU)", fmt.Sprintf("%v", cfg.UseUSP))
	printKeyValue("Offload", fmt.Sprintf("%v", cfg.Offload))
	printKeyValue("TeaCache", fmt.Sprintf("%v", cfg.TeaCache))
	if cfg.TeaCache {
		printKeyValue("TeaCache Thresh", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))
	}

	printSection("Command")
	args := cfg.ToArgs()
	fmt.Printf("  python3 %s \\\n", cfg.Script)
	for i, arg := range args {
		if i < len(args)-1 {
			fmt.Printf("    %s \\\n", arg)
		} else {
			fmt.Printf("    %s\n", arg)
		}
	}

	fmt.Println()
	printStatus("info", "Dry run - no execution. Use 'anime reel run' to generate.")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// UI Helper functions using anime theme
func printHeader(text string) {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎬 " + text))
	fmt.Println()
}

func printSection(text string) {
	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  " + text))
}

func printSubSection(text string) {
	fmt.Println()
	fmt.Println(theme.SubHeaderStyle.Render("    " + text))
}

func printKeyValue(key, value string) {
	fmt.Printf("    %s  %s\n",
		theme.HighlightStyle.Render(fmt.Sprintf("%-14s", key)),
		theme.PrimaryTextStyle.Render(value))
}

func printStatus(status, message string) {
	var symbol, style string
	switch status {
	case "success":
		symbol = theme.SymbolSuccess
		style = theme.SuccessStyle.Render(message)
	case "error":
		symbol = theme.SymbolError
		style = theme.ErrorStyle.Render(message)
	case "warning":
		symbol = theme.SymbolWarning
		style = theme.WarningStyle.Render(message)
	case "info":
		symbol = theme.SymbolInfo
		style = theme.InfoStyle.Render(message)
	case "pending":
		symbol = theme.SymbolLoading
		style = theme.DimTextStyle.Render(message)
	default:
		symbol = "  "
		style = message
	}
	fmt.Printf("  %s %s\n", symbol, style)
}

func printSuggestion(problem string, suggestions []string) {
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render(problem))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("    Suggestions:"))
	for _, s := range suggestions {
		fmt.Printf("      %s %s\n", theme.SuccessStyle.Render("->"), theme.PrimaryTextStyle.Render(s))
	}
	fmt.Println()
}

func printList(item string) {
	fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("->"), theme.PrimaryTextStyle.Render(item))
}
