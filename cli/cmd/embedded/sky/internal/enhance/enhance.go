package enhance

import (
    "fmt"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Enhancement represents a possible enhancement
type Enhancement struct {
    Name        string
    Category    string
    Impact      string
    Difficulty  string
    Description string
    Action      string
    Applied     bool
}

// Enhance handles enhancement suggestions
type Enhance struct {
    Config *config.Config
}

// New creates a new Enhance instance
func New(cfg *config.Config) *Enhance {
    return &Enhance{Config: cfg}
}

// GetEnhancements returns available enhancements based on config
func (e *Enhance) GetEnhancements() []Enhancement {
    enhancements := []Enhancement{}

    // Check FP8
    if !e.Config.Optimization.FP8Quantization {
        enhancements = append(enhancements, Enhancement{
            Name:        "Enable FP8 Quantization",
            Category:    "Memory",
            Impact:      "High",
            Difficulty:  "Easy",
            Description: "Use FP8 precision to reduce memory usage by ~50%",
            Action:      "sky config set optimization.fp8_quantization true",
            Applied:     false,
        })
    }

    // Check TeaCache
    if !e.Config.Optimization.TeaCacheEnabled {
        enhancements = append(enhancements, Enhancement{
            Name:        "Enable TeaCache",
            Category:    "Speed",
            Impact:      "Medium",
            Difficulty:  "Easy",
            Description: "Token-level caching for ~20-30% speedup",
            Action:      "sky config set optimization.teacache_enabled true",
            Applied:     false,
        })
    } else if e.Config.Optimization.TeaCacheThresh > 0.35 {
        enhancements = append(enhancements, Enhancement{
            Name:        "Optimize TeaCache Threshold",
            Category:    "Quality",
            Impact:      "Low",
            Difficulty:  "Easy",
            Description: "Lower threshold for better quality (current: %.2f)",
            Action:      "sky config set optimization.teacache_thresh 0.3",
            Applied:     false,
        })
    }

    // Check model compilation
    if !e.Config.Optimization.CompileModel {
        enhancements = append(enhancements, Enhancement{
            Name:        "Enable Model Compilation",
            Category:    "Speed",
            Impact:      "Medium",
            Difficulty:  "Easy",
            Description: "Use torch.compile for ~10-15% speedup after warmup",
            Action:      "sky config set optimization.compile_model true",
            Applied:     false,
        })
    }

    // Check ar_step for continuity
    if e.Config.Diffusion.ARStep == 0 {
        enhancements = append(enhancements, Enhancement{
            Name:        "Enable Diffusion Forcing",
            Category:    "Quality",
            Impact:      "High",
            Difficulty:  "Easy",
            Description: "Use ar_step > 0 for better video continuity",
            Action:      "sky config set diffusion.ar_step 5",
            Applied:     false,
        })
    }

    // Check parallelism optimization
    if e.Config.Hardware.GPUCount == 4 && e.Config.Parallelism.ContextParallel != 4 {
        enhancements = append(enhancements, Enhancement{
            Name:        "Optimize Parallelism",
            Category:    "Speed",
            Impact:      "High",
            Difficulty:  "Medium",
            Description: "Use CP4 for maximum throughput with 4 GPUs",
            Action:      "sky config set parallelism.context_parallel 4",
            Applied:     false,
        })
    }

    // Check model variant
    if e.Config.Hardware.GPUMemoryGB >= 80 && e.Config.Model.Variant != "SkyReels-V2-DF-14B-540P" {
        enhancements = append(enhancements, Enhancement{
            Name:        "Use Larger Model",
            Category:    "Quality",
            Impact:      "High",
            Difficulty:  "Medium",
            Description: "H100 80GB can run 14B model for best quality",
            Action:      "sky config set model.variant SkyReels-V2-DF-14B-540P",
            Applied:     false,
        })
    }

    // Check attention
    if e.Config.Model.Attention != "flash_attention_2" {
        enhancements = append(enhancements, Enhancement{
            Name:        "Use Flash Attention 2",
            Category:    "Speed",
            Impact:      "Medium",
            Difficulty:  "Medium",
            Description: "Optimized attention for faster inference",
            Action:      "sky config set model.attention flash_attention_2",
            Applied:     false,
        })
    }

    // Check inference steps
    if e.Config.Diffusion.NumInferSteps > 35 {
        enhancements = append(enhancements, Enhancement{
            Name:        "Reduce Inference Steps",
            Category:    "Speed",
            Impact:      "Medium",
            Difficulty:  "Easy",
            Description: "30 steps often sufficient for good quality",
            Action:      "sky config set diffusion.num_inference_steps 30",
            Applied:     false,
        })
    }

    return enhancements
}

// Print prints enhancement suggestions
func (e *Enhance) Print() {
    ui.PrintHeader("Enhancement Suggestions")

    enhancements := e.GetEnhancements()

    if len(enhancements) == 0 {
        ui.PrintSection("Status")
        ui.PrintStatus("success", "Configuration is fully optimized!")
        fmt.Println()
        ui.PrintList("All recommended optimizations are already applied")
        ui.PrintList("Run 'sky benchmark' to verify performance")
        return
    }

    ui.PrintSection("Available Enhancements")
    fmt.Printf("  Found %s potential improvements\n\n", ui.Highlight(fmt.Sprintf("%d", len(enhancements))))

    // Group by category
    categories := map[string][]Enhancement{
        "Speed":   {},
        "Memory":  {},
        "Quality": {},
    }

    for _, enh := range enhancements {
        categories[enh.Category] = append(categories[enh.Category], enh)
    }

    for category, items := range categories {
        if len(items) == 0 {
            continue
        }

        ui.PrintSubSection(category + " Enhancements")
        for i, enh := range items {
            impactColor := ui.BrightGreen
            if enh.Impact == "Medium" {
                impactColor = ui.BrightYellow
            } else if enh.Impact == "Low" {
                impactColor = ui.BrightBlack
            }

            fmt.Printf("    %d. %s %s[%s]%s\n", i+1, ui.Title(enh.Name), impactColor, enh.Impact, ui.Reset)
            fmt.Printf("       %s\n", ui.Muted(enh.Description))
            fmt.Printf("       %s %s\n\n", ui.Key("Apply:"), ui.Value(enh.Action))
        }
    }

    ui.PrintSection("Quick Apply")
    fmt.Printf("  Apply all enhancements: %s\n", ui.Value("sky enhance apply all"))
    fmt.Printf("  Apply by category:      %s\n", ui.Value("sky enhance apply speed"))
}

// PrintApply simulates applying enhancements
func (e *Enhance) PrintApply(target string) {
    ui.PrintHeader("Applying Enhancements")

    enhancements := e.GetEnhancements()

    var toApply []Enhancement
    switch target {
    case "all":
        toApply = enhancements
    case "speed":
        for _, enh := range enhancements {
            if enh.Category == "Speed" {
                toApply = append(toApply, enh)
            }
        }
    case "memory":
        for _, enh := range enhancements {
            if enh.Category == "Memory" {
                toApply = append(toApply, enh)
            }
        }
    case "quality":
        for _, enh := range enhancements {
            if enh.Category == "Quality" {
                toApply = append(toApply, enh)
            }
        }
    default:
        ui.PrintSuggestion("Unknown target: "+target, []string{
            "Use: all, speed, memory, or quality",
            "Example: sky enhance apply speed",
        })
        return
    }

    if len(toApply) == 0 {
        ui.PrintStatus("info", "No enhancements to apply for: "+target)
        return
    }

    ui.PrintSection("Changes to Apply")
    for _, enh := range toApply {
        ui.PrintStatus("pending", enh.Name)
    }

    fmt.Println()
    ui.PrintSuggestion("Dry run mode - no changes made", []string{
        "To apply changes, run the individual commands shown above",
        "Or edit ~/.sky/config.json directly",
        "Run 'sky status' after applying changes",
    })
}

// PrintProfile prints and applies optimization profiles
func (e *Enhance) PrintProfile(profile string) {
    ui.PrintHeader("Optimization Profiles")

    profiles := map[string]struct {
        Name        string
        Description string
        Settings    map[string]string
    }{
        "speed": {
            Name:        "Maximum Speed",
            Description: "Optimize for fastest generation",
            Settings: map[string]string{
                "diffusion.ar_step":              "0",
                "optimization.teacache_enabled":  "true",
                "optimization.teacache_thresh":   "0.4",
                "optimization.compile_model":     "true",
                "diffusion.num_inference_steps":  "25",
            },
        },
        "quality": {
            Name:        "Maximum Quality",
            Description: "Optimize for best output quality",
            Settings: map[string]string{
                "diffusion.ar_step":              "10",
                "optimization.teacache_enabled":  "false",
                "diffusion.num_inference_steps":  "50",
                "diffusion.guidance_scale":       "7.5",
            },
        },
        "balanced": {
            Name:        "Balanced",
            Description: "Good balance of speed and quality",
            Settings: map[string]string{
                "diffusion.ar_step":              "5",
                "optimization.teacache_enabled":  "true",
                "optimization.teacache_thresh":   "0.3",
                "optimization.compile_model":     "true",
                "diffusion.num_inference_steps":  "30",
            },
        },
        "memory": {
            Name:        "Memory Efficient",
            Description: "Minimize memory usage",
            Settings: map[string]string{
                "optimization.fp8_quantization":  "true",
                "optimization.offload_enabled":   "true",
                "generation.base_num_frames":     "49",
            },
        },
    }

    if profile == "" {
        // List all profiles
        ui.PrintSection("Available Profiles")
        for key, p := range profiles {
            fmt.Printf("\n  %s %s\n", ui.Highlight(key), ui.Muted("- "+p.Name))
            fmt.Printf("    %s\n", p.Description)
        }
        fmt.Printf("\n  Apply profile: %s\n", ui.Value("sky enhance profile <name>"))
        return
    }

    p, exists := profiles[profile]
    if !exists {
        ui.PrintSuggestion("Unknown profile: "+profile, []string{
            "Available profiles: speed, quality, balanced, memory",
            "Run 'sky enhance profile' to see all profiles",
        })
        return
    }

    ui.PrintSection("Profile: " + p.Name)
    fmt.Printf("  %s\n\n", p.Description)

    ui.PrintSubSection("Settings")
    for key, val := range p.Settings {
        fmt.Printf("    %s = %s\n", ui.Key(key), ui.Value(val))
    }

    fmt.Println()
    ui.PrintSuggestion("Dry run mode - no changes made", []string{
        "To apply this profile, run the settings commands manually",
        fmt.Sprintf("Example: sky config set %s", p.Settings["diffusion.ar_step"]),
    })
}
