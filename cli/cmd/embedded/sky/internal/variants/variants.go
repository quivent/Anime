package variants

import (
    "fmt"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Variants handles variant listing and configuration
type Variants struct {
    Config *config.Config
}

// New creates a new Variants instance
func New(cfg *config.Config) *Variants {
    return &Variants{Config: cfg}
}

// PrintList prints all available variants
func (v *Variants) PrintList() {
    ui.PrintHeader("Available Model Variants")

    variants := config.GetVariants()

    ui.PrintSection("Models")
    headers := []string{"Variant", "Params", "Resolution", "VRAM", "Speed"}
    rows := make([][]string, len(variants))
    for i, variant := range variants {
        current := ""
        if variant.Name == v.Config.Model.Variant {
            current = " " + ui.Success("←")
        }
        rows[i] = []string{
            variant.Name + current,
            variant.Parameters,
            variant.Resolution,
            variant.VRAM,
            variant.Speed,
        }
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("Parallelism Strategies")
    strategies := config.GetParallelismStrategies()
    headers = []string{"Strategy", "Context", "CFG", "Best For"}
    rows = make([][]string, len(strategies))
    for i, s := range strategies {
        current := ""
        if s.ContextParallel == v.Config.Parallelism.ContextParallel &&
            s.CFGParallel == v.Config.Parallelism.CFGParallel {
            current = " " + ui.Success("←")
        }
        rows[i] = []string{
            s.Name + current,
            fmt.Sprintf("%d", s.ContextParallel),
            fmt.Sprintf("%d", s.CFGParallel),
            s.BestFor,
        }
    }
    ui.PrintTable(headers, rows)

    fmt.Println()
    ui.PrintStatus("info", "Current model: "+ui.Value(v.Config.Model.Variant))
    ui.PrintStatus("info", "Current strategy: "+ui.Value(fmt.Sprintf("CP%d", v.Config.Parallelism.ContextParallel)))
}

// PrintDetails prints details for a specific variant
func (v *Variants) PrintDetails(name string) {
    variants := config.GetVariants()

    var found *config.ModelVariant
    for _, variant := range variants {
        if variant.Name == name || variant.Parameters == name {
            found = &variant
            break
        }
    }

    if found == nil {
        ui.PrintSuggestion("Variant not found: "+name, []string{
            "Run 'sky variants' to see all available variants",
            "Try: SkyReels-V2-DF-14B, SkyReels-V2-DF-1.3B, etc.",
        })
        return
    }

    ui.PrintHeader("Variant Details: " + found.Name)

    ui.PrintSection("Specifications")
    ui.PrintKeyValue("Parameters", found.Parameters)
    ui.PrintKeyValue("Resolution", found.Resolution)
    ui.PrintKeyValue("VRAM Required", found.VRAM)
    ui.PrintKeyValue("Speed", found.Speed)
    ui.PrintKeyValue("Quality", found.Quality)

    ui.PrintSection("Description")
    fmt.Printf("  %s\n", found.Description)

    ui.PrintSection("Compatibility with Current Hardware")
    gpuMem := v.Config.Hardware.GPUMemoryGB

    // Parse VRAM requirement
    var requiredMem int
    fmt.Sscanf(found.VRAM, "~%dGB", &requiredMem)

    if requiredMem <= gpuMem {
        ui.PrintStatus("success", fmt.Sprintf("Compatible with %s (%dGB)", v.Config.Hardware.GPUModel, gpuMem))
        headroom := gpuMem - requiredMem
        ui.PrintKeyValue("Available Headroom", fmt.Sprintf("~%dGB", headroom))
    } else {
        ui.PrintStatus("warning", fmt.Sprintf("May require offloading on %s (%dGB)", v.Config.Hardware.GPUModel, gpuMem))
        ui.PrintList("Consider using FP8 quantization")
        ui.PrintList("Or use a smaller model variant")
    }

    ui.PrintSection("To Use This Variant")
    fmt.Printf("  %s sky config set model.variant %s\n", ui.Muted("$"), found.Name)
}

// PrintRecommendation prints recommended variant for hardware
func (v *Variants) PrintRecommendation() {
    ui.PrintHeader("Recommended Configuration")

    gpuCount := v.Config.Hardware.GPUCount
    gpuMem := v.Config.Hardware.GPUMemoryGB

    ui.PrintSection("Your Hardware")
    ui.PrintKeyValue("GPUs", fmt.Sprintf("%d× %s", gpuCount, v.Config.Hardware.GPUModel))
    ui.PrintKeyValue("Memory/GPU", fmt.Sprintf("%dGB", gpuMem))
    ui.PrintKeyValue("Total VRAM", fmt.Sprintf("%dGB", gpuCount*gpuMem))

    ui.PrintSection("Recommended Model")

    if gpuMem >= 80 {
        ui.PrintStatus("success", "SkyReels-V2-DF-14B (FP8)")
        ui.PrintList("Maximum quality with ample headroom")
        ui.PrintList("~41GB headroom for TeaCache and iteration")
    } else if gpuMem >= 48 {
        ui.PrintStatus("success", "SkyReels-V2-DF-14B (FP8) with offloading")
        ui.PrintList("Or: SkyReels-V2-DF-5B (FP16)")
    } else if gpuMem >= 24 {
        ui.PrintStatus("success", "SkyReels-V2-DF-1.3B")
        ui.PrintList("Best for consumer GPUs")
    } else {
        ui.PrintStatus("warning", "Limited options")
        ui.PrintList("Consider cloud GPU or upgrade")
    }

    ui.PrintSection("Recommended Parallelism")

    switch gpuCount {
    case 1:
        ui.PrintStatus("info", "Single GPU - No parallelism needed")
    case 2:
        ui.PrintStatus("info", "CP2 - 2-way context parallel")
        ui.PrintList("Or CFG2 for shorter videos")
    case 4:
        ui.PrintStatus("success", "CP4 - 4-way context parallel")
        ui.PrintList("Best throughput for long sequences")
        ui.PrintList("Alternative: CP2+CFG2 for balanced approach")
    case 8:
        ui.PrintStatus("success", "CP4+CFG2 or CP8")
        ui.PrintList("Maximum parallelism")
    default:
        ui.PrintStatus("info", fmt.Sprintf("Custom: Consider CP%d", gpuCount))
    }

    ui.PrintSection("Recommended Optimizations")
    ui.PrintList("FP8 quantization: Enabled (H100 native support)")
    ui.PrintList("TeaCache: Enabled (thresh=0.3)")
    ui.PrintList("Model compilation: Enabled")
    ui.PrintList("ar_step: 5 (for continuity)")
}
