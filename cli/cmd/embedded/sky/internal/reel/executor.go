package reel

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"

    "github.com/sky-cli/sky/ui"
)

// Execute runs the configured generation
func Execute(cfg *ReelConfig) {
    // Check SkyReels directory
    if _, err := os.Stat(SkyReelsDir); os.IsNotExist(err) {
        ui.PrintSuggestion("SkyReels-V2 not found", []string{
            "Expected location: " + SkyReelsDir,
            "Clone: git clone git@github.com:SkyworkAI/SkyReels-V2.git",
        })
        return
    }

    // Validate prompt
    if cfg.Prompt == "" {
        ui.PrintSuggestion("No prompt specified", []string{
            "Set prompt: sky reel prompt set \"Your prompt here\"",
            "Or run: sky reel prompt interactive",
        })
        return
    }

    scriptPath := SkyReelsDir + "/" + cfg.Script

    // Check script exists
    if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
        ui.PrintSuggestion("Script not found: "+cfg.Script, []string{
            "Available: generate_video.py, generate_video_df.py, generate_video_sequential.py",
        })
        return
    }

    // Find python
    pythonPath, err := exec.LookPath("python3")
    if err != nil {
        pythonPath, err = exec.LookPath("python")
        if err != nil {
            ui.PrintSuggestion("Python not found", []string{
                "Install Python 3 or activate your virtual environment",
            })
            return
        }
    }

    // Build args
    args := cfg.ToArgs()

    // Print configuration
    ui.PrintHeader("SkyReel Generation")

    ui.PrintSection("Configuration")
    ui.PrintKeyValue("Script", cfg.Script)
    ui.PrintKeyValue("Prompt", truncate(cfg.Prompt, 50))
    ui.PrintKeyValue("Frames", fmt.Sprintf("%d (~%.1fs)", cfg.NumFrames, float64(cfg.NumFrames)/float64(cfg.FPS)))
    ui.PrintKeyValue("Resolution", cfg.Resolution)
    ui.PrintKeyValue("Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
    ui.PrintKeyValue("Guidance", fmt.Sprintf("%.1f", cfg.GuidanceScale))

    if cfg.UseUSP {
        ui.PrintKeyValue("Parallelism", "USP enabled (multi-GPU)")
    }
    if cfg.TeaCache {
        ui.PrintKeyValue("TeaCache", fmt.Sprintf("enabled (thresh=%.2f)", cfg.TeaCacheThresh))
    }
    if cfg.Offload {
        ui.PrintKeyValue("Offload", "enabled")
    }

    ui.PrintSection("Executing")
    ui.PrintKeyValue("Command", "python3 "+cfg.Script)
    fmt.Println()

    // Change to SkyReels directory
    if err := os.Chdir(SkyReelsDir); err != nil {
        ui.PrintStatus("error", "Failed to change directory: "+err.Error())
        return
    }

    // Build full command
    cmdArgs := append([]string{pythonPath, scriptPath}, args...)
    env := os.Environ()

    // Execute, replacing current process
    if err := syscall.Exec(pythonPath, cmdArgs, env); err != nil {
        ui.PrintStatus("error", "Execution failed: "+err.Error())
    }
}

// ExecuteDry shows what would be executed without running
func ExecuteDry(cfg *ReelConfig) {
    ui.PrintHeader("SkyReel Generation (Dry Run)")

    ui.PrintSection("Configuration")
    ui.PrintKeyValue("Script", cfg.Script)
    ui.PrintKeyValue("Prompt", cfg.Prompt)
    ui.PrintKeyValue("Frames", fmt.Sprintf("%d", cfg.NumFrames))
    ui.PrintKeyValue("Resolution", cfg.Resolution)
    ui.PrintKeyValue("Model", cfg.ModelID)
    ui.PrintKeyValue("Steps", fmt.Sprintf("%d", cfg.InferenceSteps))
    ui.PrintKeyValue("Guidance", fmt.Sprintf("%.1f", cfg.GuidanceScale))
    ui.PrintKeyValue("Shift", fmt.Sprintf("%.1f", cfg.Shift))
    ui.PrintKeyValue("FPS", fmt.Sprintf("%d", cfg.FPS))
    ui.PrintKeyValue("Seed", fmt.Sprintf("%d", cfg.Seed))
    ui.PrintKeyValue("Output", cfg.OutDir)

    ui.PrintSection("Optimizations")
    ui.PrintKeyValue("USP (multi-GPU)", fmt.Sprintf("%v", cfg.UseUSP))
    ui.PrintKeyValue("Offload", fmt.Sprintf("%v", cfg.Offload))
    ui.PrintKeyValue("TeaCache", fmt.Sprintf("%v", cfg.TeaCache))
    if cfg.TeaCache {
        ui.PrintKeyValue("TeaCache Thresh", fmt.Sprintf("%.2f", cfg.TeaCacheThresh))
    }

    ui.PrintSection("Command")
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
    ui.PrintStatus("info", "Dry run - no execution. Use 'sky reel run' to generate.")
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}
