package architecture

import (
    "fmt"
    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Architecture handles architecture visualization
type Architecture struct {
    Config *config.Config
}

// New creates a new Architecture instance
func New(cfg *config.Config) *Architecture {
    return &Architecture{Config: cfg}
}

// PrintLinear prints linear/text-based architecture
func (a *Architecture) PrintLinear() {
    ui.PrintHeader("SkyReel Architecture (Linear View)")

    ui.PrintSection("Pipeline Flow")
    fmt.Println()
    fmt.Printf("  %s\n", ui.Title("Text Input"))
    fmt.Printf("      %s\n", ui.Muted("│"))
    fmt.Printf("      %s\n", ui.Muted("▼"))
    fmt.Printf("  %s %s\n", ui.Highlight("[T5-XXL]"), ui.Muted("─── Text Encoder (CPU/GPU offload)"))
    fmt.Printf("      %s\n", ui.Muted("│"))
    fmt.Printf("      %s %s\n", ui.Muted("▼"), ui.Muted("text embeddings"))
    fmt.Printf("  %s\n", ui.Highlight("[Diffusion Transformer]"))
    fmt.Printf("      %s %s\n", ui.Muted("│"), ui.Info("← xDiT Context Parallel (4 GPUs)"))
    fmt.Printf("      %s %s\n", ui.Muted("│"), ui.Info("← Diffusion Forcing (ar_step)"))
    fmt.Printf("      %s\n", ui.Muted("▼"))
    fmt.Printf("  %s %s\n", ui.Highlight("[3D VAE Decoder]"), ui.Muted("─── Latent → Pixels"))
    fmt.Printf("      %s\n", ui.Muted("│"))
    fmt.Printf("      %s\n", ui.Muted("▼"))
    fmt.Printf("  %s\n", ui.Success("Video Output"))
    fmt.Println()

    ui.PrintSection("GPU Distribution")
    fmt.Println()
    fmt.Printf("  %s\n", ui.Key("Context Parallel Mode (CP4):"))
    fmt.Printf("    GPU 0: Frames [0-%d]     %s\n", a.Config.Generation.BaseNumFrames/4, ui.Muted("+ full model weights"))
    fmt.Printf("    GPU 1: Frames [%d-%d]   %s\n", a.Config.Generation.BaseNumFrames/4+1, a.Config.Generation.BaseNumFrames/2, ui.Muted("+ full model weights"))
    fmt.Printf("    GPU 2: Frames [%d-%d]   %s\n", a.Config.Generation.BaseNumFrames/2+1, 3*a.Config.Generation.BaseNumFrames/4, ui.Muted("+ full model weights"))
    fmt.Printf("    GPU 3: Frames [%d-%d]   %s\n", 3*a.Config.Generation.BaseNumFrames/4+1, a.Config.Generation.BaseNumFrames, ui.Muted("+ full model weights"))
    fmt.Println()

    ui.PrintSection("Memory Layout (Per GPU)")
    budget := a.Config.GetMemoryBudget()
    fmt.Println()
    fmt.Printf("  ┌─────────────────────────────────────┐\n")
    fmt.Printf("  │ %s            │ %s\n", ui.Key("Model Weights (FP8)"), ui.Value("~14GB"))
    fmt.Printf("  ├─────────────────────────────────────┤\n")
    fmt.Printf("  │ %s              │ %s\n", ui.Key("T5-XXL Encoder"), ui.Value("~6GB"))
    fmt.Printf("  ├─────────────────────────────────────┤\n")
    fmt.Printf("  │ %s          │ %s\n", ui.Key("Context Activations"), ui.Value("~15GB"))
    fmt.Printf("  ├─────────────────────────────────────┤\n")
    fmt.Printf("  │ %s                     │ %s\n", ui.Key("3D VAE"), ui.Value("~4GB"))
    fmt.Printf("  ├─────────────────────────────────────┤\n")
    fmt.Printf("  │ %s                  │ %s\n", ui.Success("HEADROOM"), ui.Success(fmt.Sprintf("~%dGB", budget.HeadroomGB)))
    fmt.Printf("  │ %s                          │\n", ui.Muted("├─ TeaCache"))
    fmt.Printf("  │ %s                          │\n", ui.Muted("├─ Diffusion states"))
    fmt.Printf("  │ %s                          │\n", ui.Muted("└─ Iteration buffer"))
    fmt.Printf("  └─────────────────────────────────────┘\n")
    fmt.Println()
}

// PrintDiagram prints ASCII diagram architecture
func (a *Architecture) PrintDiagram() {
    ui.PrintHeader("SkyReel Architecture (Diagram View)")

    fmt.Println()
    lines := []string{
        "┌─────────────────────────────────────────────────────────────────────────────┐",
        "│                         ORCHESTRATION LAYER                                  │",
        "│  ┌─────────────────┐  ┌──────────────────┐  ┌─────────────────────────────┐ │",
        "│  │  Request Queue  │  │  Diffusion       │  │  Quality Controller         │ │",
        "│  │  (Prompt length │  │  Forcing State   │  │  (ar_step tuning,           │ │",
        "│  │   bucketing)    │  │  (ar_step mgmt)  │  │   TeaCache thresh)          │ │",
        "│  └────────┬────────┘  └────────┬─────────┘  └──────────────┬──────────────┘ │",
        "└───────────┼────────────────────┼───────────────────────────┼────────────────┘",
        "            │                    │                           │                 ",
        "            ▼                    ▼                           ▼                 ",
        "┌─────────────────────────────────────────────────────────────────────────────┐",
        "│              INFERENCE ENGINE (xDiT USP - Custom Parallelism)               │",
        "│                                                                              │",
        "│  ┌──────────────────────────────────────────────────────────────────────┐   │",
        "│  │                    NVLink Mesh (900 GB/s SXM5)                        │   │",
        "│  │                                                                       │   │",
        "│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐      │   │",
        "│  │  │   H100 #0  │  │   H100 #1  │  │   H100 #2  │  │   H100 #3  │      │   │",
        "│  │  │   80GB     │◄─►│   80GB     │◄─►│   80GB     │◄─►│   80GB     │      │   │",
        "│  │  │            │  │            │  │            │  │            │      │   │",
        "│  │  │ Context    │  │ Context    │  │ Context    │  │ Context    │      │   │",
        "│  │  │ Parallel   │  │ Parallel   │  │ Parallel   │  │ Parallel   │      │   │",
        "│  │  │ (frames    │  │ (frames    │  │ (frames    │  │ (frames    │      │   │",
        "│  │  │  0-24)     │  │  25-48)    │  │  49-72)    │  │  73-97)    │      │   │",
        "│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘      │   │",
        "│  │                                                                       │   │",
        "│  └──────────────────────────────────────────────────────────────────────┘   │",
        "│                                                                              │",
        "│  Memory per GPU (14B model, FP8):                                           │",
        "│  ├── Model Weights:     ~14GB (replicated)                                  │",
        "│  ├── T5-XXL:            ~6GB                                                │",
        "│  ├── Context partition: ~15GB                                               │",
        "│  ├── 3D VAE:            ~4GB                                                │",
        "│  └── Headroom:          ~41GB                                               │",
        "│                                                                              │",
        "└─────────────────────────────────────────────────────────────────────────────┘",
    }
    ui.PrintDiagram(lines)

    fmt.Println()
    ui.PrintSection("Diffusion Forcing Flow")
    fmt.Println()

    dfLines := []string{
        "┌─────────────────────────────────────────────────────────────────┐",
        "│           AUTOREGRESSIVE DIFFUSION FORCING                      │",
        "│                                                                 │",
        "│   ar_step = 5 (Asynchronous):                                  │",
        "│   ┌─────────────────────────────────────────────────────────┐  │",
        "│   │  Block 0   Block 1   Block 2   Block 3                  │  │",
        "│   │    t=T                                                  │  │",
        "│   │    t=T-5     t=T                                        │  │",
        "│   │    t=T-10    t=T-5     t=T                              │  │",
        "│   │    t=T-15    t=T-10    t=T-5     t=T                    │  │",
        "│   │     ↓         ↓         ↓         ↓                     │  │",
        "│   │    t=0       t=5       t=10      t=15   (staggered)     │  │",
        "│   └─────────────────────────────────────────────────────────┘  │",
        "│   → Later blocks condition on cleaner earlier blocks           │",
        "│   → Enables infinite-length generation with continuity         │",
        "│                                                                 │",
        "└─────────────────────────────────────────────────────────────────┘",
    }
    ui.PrintDiagram(dfLines)
}

// PrintCurrent prints current architecture based on config
func (a *Architecture) PrintCurrent() {
    ui.PrintHeader("Current Architecture Configuration")

    ui.PrintSection("Hardware")
    ui.PrintKeyValue("GPUs", fmt.Sprintf("%d× %s", a.Config.Hardware.GPUCount, a.Config.Hardware.GPUModel))
    ui.PrintKeyValue("Memory/GPU", fmt.Sprintf("%dGB", a.Config.Hardware.GPUMemoryGB))
    ui.PrintKeyValue("NVLink", fmt.Sprintf("%d GB/s", a.Config.Hardware.NVLinkBandwidth))

    ui.PrintSection("Model")
    ui.PrintKeyValue("Variant", a.Config.Model.Variant)
    ui.PrintKeyValue("Precision", a.Config.Model.Precision)
    ui.PrintKeyValue("Text Encoder", a.Config.Model.TextEncoder)
    ui.PrintKeyValue("Attention", a.Config.Model.Attention)

    ui.PrintSection("Parallelism")
    ui.PrintKeyValue("Strategy", a.Config.Parallelism.Strategy)
    ui.PrintKeyValue("Context Parallel", fmt.Sprintf("%d", a.Config.Parallelism.ContextParallel))
    ui.PrintKeyValue("CFG Parallel", fmt.Sprintf("%d", a.Config.Parallelism.CFGParallel))
    ui.PrintKeyValue("VAE Parallel", fmt.Sprintf("%v", a.Config.Parallelism.VAEParallel))

    ui.PrintSection("Generation")
    ui.PrintKeyValue("Resolution", a.Config.Generation.Resolution)
    ui.PrintKeyValue("Dimensions", fmt.Sprintf("%d×%d", a.Config.Generation.Width, a.Config.Generation.Height))
    ui.PrintKeyValue("FPS", fmt.Sprintf("%d", a.Config.Generation.FPS))
    ui.PrintKeyValue("Max Frames", fmt.Sprintf("%d (~%.1fs)", a.Config.Generation.MaxFrames, float64(a.Config.Generation.MaxFrames)/float64(a.Config.Generation.FPS)))

    ui.PrintSection("Diffusion")
    ui.PrintKeyValue("ar_step", fmt.Sprintf("%d", a.Config.Diffusion.ARStep))
    ui.PrintKeyValue("Inference Steps", fmt.Sprintf("%d", a.Config.Diffusion.NumInferSteps))
    ui.PrintKeyValue("Guidance Scale", fmt.Sprintf("%.1f", a.Config.Diffusion.GuidanceScale))

    ui.PrintSection("Optimization")
    ui.PrintKeyValue("TeaCache", fmt.Sprintf("%v (thresh=%.2f)", a.Config.Optimization.TeaCacheEnabled, a.Config.Optimization.TeaCacheThresh))
    ui.PrintKeyValue("Model Compile", fmt.Sprintf("%v", a.Config.Optimization.CompileModel))
    ui.PrintKeyValue("FP8 Quantization", fmt.Sprintf("%v", a.Config.Optimization.FP8Quantization))

    ui.PrintSection("Memory Budget")
    budget := a.Config.GetMemoryBudget()
    ui.PrintProgressBar(budget.TotalGB-budget.HeadroomGB, budget.TotalGB, 50)
    ui.PrintKeyValue("Used", fmt.Sprintf("%dGB", budget.TotalGB-budget.HeadroomGB))
    ui.PrintKeyValue("Headroom", ui.Success(fmt.Sprintf("%dGB", budget.HeadroomGB)))
}
