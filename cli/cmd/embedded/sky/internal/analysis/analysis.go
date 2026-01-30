package analysis

import (
    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Analysis contains all analysis data
type Analysis struct {
    Config *config.Config
}

// New creates a new Analysis instance
func New(cfg *config.Config) *Analysis {
    return &Analysis{Config: cfg}
}

// PrintFull prints the complete analysis
func (a *Analysis) PrintFull() {
    ui.PrintHeader("SkyReel 4×H100 Architecture Analysis")

    a.PrintOverview()
    a.PrintHardware()
    a.PrintModel()
    a.PrintParallelism()
    a.PrintMemory()
    a.PrintContinuity()
    a.PrintThroughput()
}

// PrintOverview prints the overview section
func (a *Analysis) PrintOverview() {
    ui.PrintSection("Overview")

    ui.PrintKeyValue("Configuration", a.Config.Hardware.GPUModel+" × "+string(rune('0'+a.Config.Hardware.GPUCount)))
    ui.PrintKeyValue("Model", a.Config.Model.Variant)
    ui.PrintKeyValue("Strategy", a.Config.Parallelism.Strategy)
    ui.PrintKeyValue("Target", "Sequential video generation with continuity")

    ui.PrintSubSection("Key Findings")
    ui.PrintList("SkyReel uses Context/CFG Parallel, NOT traditional Tensor Parallelism")
    ui.PrintList("Model weights are REPLICATED across GPUs (not sharded)")
    ui.PrintList("Continuity via AutoRegressive Diffusion Forcing (ar_step parameter)")
    ui.PrintList("~41GB headroom per GPU with FP8 quantization")
}

// PrintHardware prints hardware analysis
func (a *Analysis) PrintHardware() {
    ui.PrintSection("Hardware Configuration")

    headers := []string{"Spec", "Value", "Notes"}
    rows := [][]string{
        {"GPU Model", a.Config.Hardware.GPUModel, "Hopper architecture"},
        {"GPU Count", "4", "Optimal for xDiT USP"},
        {"VRAM/GPU", "80GB HBM3", "3.9 TB/s bandwidth"},
        {"NVLink", "900 GB/s", "4th gen, full mesh"},
        {"System RAM", "512GB", "For model offload if needed"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSubSection("H100 Advantages")
    ui.PrintList("Native FP8 support → 50% memory reduction")
    ui.PrintList("HBM3 bandwidth → Reduced memory bottlenecks")
    ui.PrintList("NVLink mesh → Fast inter-GPU communication for context parallel")
    ui.PrintList("Transformer Engine → Automatic mixed precision")
}

// PrintModel prints model analysis
func (a *Analysis) PrintModel() {
    ui.PrintSection("Model Architecture")

    ui.PrintKeyValue("Architecture", "Diffusion Transformer (DiT)")
    ui.PrintKeyValue("Foundation", "Based on HunyuanVideo")
    ui.PrintKeyValue("Text Encoder", "T5-XXL")
    ui.PrintKeyValue("Latent Space", "3D VAE (spatiotemporal)")

    ui.PrintSubSection("Model Variants")
    headers := []string{"Variant", "Params", "VRAM", "Quality"}
    rows := [][]string{
        {"SkyReels-V2-DF-1.3B", "1.3B", "~15GB", "Good"},
        {"SkyReels-V2-DF-5B", "5B", "~25GB", "Better"},
        {"SkyReels-V2-DF-14B", "14B", "~51GB", "Best"},
    }
    ui.PrintTable(headers, rows)
}

// PrintParallelism prints parallelism strategy analysis
func (a *Analysis) PrintParallelism() {
    ui.PrintSection("Parallelism Strategy")

    ui.PrintSubSection("xDiT USP (NOT Traditional TP)")
    ui.PrintList("Context Parallel: Distributes frames across GPUs")
    ui.PrintList("CFG Parallel: Parallelizes conditional/unconditional branches")
    ui.PrintList("VAE Parallel: Distributes encoding/decoding")

    ui.PrintSubSection("Current Configuration")
    ui.PrintKeyValue("Context Parallel", "4 (each GPU handles 1/4 of frames)")
    ui.PrintKeyValue("CFG Parallel", "1 (single branch per context group)")
    ui.PrintKeyValue("VAE Parallel", "Enabled")

    ui.PrintSubSection("Key Difference from TP")
    content := []string{
        "Traditional TP: Model weights SHARDED across GPUs",
        "SkyReel xDiT: Model weights REPLICATED on each GPU",
        "",
        "Implication: More headroom per GPU (~41GB vs ~20GB)",
        "Trade-off: Higher aggregate memory usage",
    }
    ui.PrintBox("Parallelism Comparison", content)
}

// PrintMemory prints memory analysis
func (a *Analysis) PrintMemory() {
    ui.PrintSection("Memory Budget (Per GPU)")

    budget := a.Config.GetMemoryBudget()

    headers := []string{"Component", "Size", "Notes"}
    rows := [][]string{
        {"Model Weights (FP8)", "~14GB", "Full model replicated"},
        {"T5-XXL Encoder", "~6GB", "Text embedding"},
        {"3D VAE", "~4GB", "Encode/decode"},
        {"Context Partition", "~15GB", "1/4 of sequence activations"},
        {"Headroom", "~41GB", "TeaCache, iteration, overflow"},
    }
    ui.PrintTable(headers, rows)

    // Progress bar showing usage
    used := budget.TotalGB - budget.HeadroomGB
    ui.PrintSubSection("Memory Utilization")
    ui.PrintProgressBar(used, budget.TotalGB, 50)
    ui.PrintKeyValue("Used", ui.Value(string(rune('0'+used/10))+string(rune('0'+used%10))+"GB"))
    ui.PrintKeyValue("Headroom", ui.Success(string(rune('0'+budget.HeadroomGB/10))+string(rune('0'+budget.HeadroomGB%10))+"GB"))
}

// PrintContinuity prints continuity/diffusion forcing analysis
func (a *Analysis) PrintContinuity() {
    ui.PrintSection("Continuity System")

    ui.PrintSubSection("AutoRegressive Diffusion Forcing")
    ui.PrintKeyValue("Mechanism", "Staggered noise levels across frame blocks")
    ui.PrintKeyValue("ar_step", "5 (Block i lags Block i-1 by 5 timesteps)")

    ui.PrintSubSection("Modes")
    headers := []string{"Mode", "ar_step", "Speed", "Continuity"}
    rows := [][]string{
        {"Synchronous", "0", "Faster", "Basic"},
        {"Asynchronous", "5+", "Slower", "Superior"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSubSection("How It Works")
    ui.PrintList("Later blocks condition on cleaner earlier blocks")
    ui.PrintList("Enables infinite-length video generation")
    ui.PrintList("Better motion consistency across segments")
    ui.PrintList("No explicit frame overlap buffer needed")
}

// PrintThroughput prints throughput estimates
func (a *Analysis) PrintThroughput() {
    ui.PrintSection("Throughput Estimates")

    ui.PrintSubSection("Reference Benchmarks")
    ui.PrintList("Single RTX 4090: ~80 sec for 4s 544p video")
    ui.PrintList("4× RTX 4090 (xDiT): ~58% latency reduction")

    ui.PrintSubSection("4×H100 Projections")
    headers := []string{"Scenario", "Duration", "Time Est."}
    rows := [][]string{
        {"4s segment (ar_step=0)", "97 frames", "~25-35s"},
        {"4s segment (ar_step=5)", "97 frames", "~40-55s"},
        {"12s segment (ar_step=5)", "289 frames", "~90-120s"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSubSection("Optimization Impact")
    ui.PrintList("TeaCache: ~20-30% speedup")
    ui.PrintList("Model compilation: ~10-15% speedup")
    ui.PrintList("FP8 quantization: ~40% memory reduction")
}

// PrintExplanation prints detailed explanation for a topic
func (a *Analysis) PrintExplanation(topic string) {
    switch topic {
    case "parallelism":
        a.explainParallelism()
    case "diffusion-forcing":
        a.explainDiffusionForcing()
    case "memory":
        a.explainMemory()
    case "teacache":
        a.explainTeaCache()
    case "nvlink":
        a.explainNVLink()
    default:
        ui.PrintSuggestion("Unknown topic: "+topic, []string{
            "Try: parallelism, diffusion-forcing, memory, teacache, nvlink",
            "Run 'sky analysis explain' to see all topics",
        })
    }
}

func (a *Analysis) explainParallelism() {
    ui.PrintHeader("Parallelism Strategy Explained")

    ui.PrintSection("Why NOT Traditional Tensor Parallelism?")
    content := []string{
        "Tensor Parallelism (TP) splits model weights across GPUs.",
        "Each GPU holds 1/N of the model and must communicate",
        "activations at every layer boundary.",
        "",
        "For video DiT models, this creates issues:",
        "• High communication overhead for 3D attention",
        "• Sequence length varies with video duration",
        "• CFG requires duplicate computation",
    }
    ui.PrintBox("TP Limitations", content)

    ui.PrintSection("Context Parallel Approach")
    content = []string{
        "Instead of splitting WEIGHTS, split the SEQUENCE:",
        "",
        "GPU 0: Frames 0-24   (full model)",
        "GPU 1: Frames 25-49  (full model)",
        "GPU 2: Frames 50-74  (full model)",
        "GPU 3: Frames 75-97  (full model)",
        "",
        "Sync only at attention boundaries, not every layer.",
    }
    ui.PrintBox("Context Parallel", content)

    ui.PrintSection("CFG Parallel")
    content = []string{
        "Classifier-Free Guidance requires two forward passes:",
        "• Conditional (with text prompt)",
        "• Unconditional (without text)",
        "",
        "CFG Parallel runs these on separate GPU groups:",
        "GPUs 0-1: Conditional branch",
        "GPUs 2-3: Unconditional branch",
    }
    ui.PrintBox("CFG Parallel", content)
}

func (a *Analysis) explainDiffusionForcing() {
    ui.PrintHeader("Diffusion Forcing Explained")

    ui.PrintSection("The Problem: Video Continuity")
    ui.PrintList("Standard diffusion denoises ALL frames at same noise level")
    ui.PrintList("When extending video, new frames don't \"know\" about old ones")
    ui.PrintList("Results in discontinuities, sudden scene changes")

    ui.PrintSection("The Solution: Staggered Noise")
    content := []string{
        "Diffusion Forcing staggers noise levels:",
        "",
        "Timestep T:   Block0=T   Block1=T   Block2=T   Block3=T",
        "Timestep T-5: Block0=T-5 Block1=T   Block2=T   Block3=T",
        "Timestep T-10: Block0=T-10 Block1=T-5 Block2=T  Block3=T",
        "...",
        "",
        "Earlier blocks denoise FASTER than later blocks.",
        "Later blocks can attend to cleaner earlier content.",
    }
    ui.PrintBox("Staggered Denoising", content)

    ui.PrintSection("ar_step Parameter")
    ui.PrintKeyValue("ar_step=0", "Synchronous mode (all same noise level)")
    ui.PrintKeyValue("ar_step=5", "Async mode (5 timestep stagger)")
    ui.PrintKeyValue("Higher ar_step", "Better continuity, slower inference")
}

func (a *Analysis) explainMemory() {
    ui.PrintHeader("Memory Management Explained")

    ui.PrintSection("Why 41GB Headroom?")
    ui.PrintList("Model replicated (not sharded) → fixed ~14GB FP8")
    ui.PrintList("Context parallel → only 1/4 activations per GPU")
    ui.PrintList("H100 has 80GB → 80 - 14 - 6 - 4 - 15 = 41GB")

    ui.PrintSection("Headroom Usage")
    headers := []string{"Purpose", "Size", "Why"}
    rows := [][]string{
        {"TeaCache", "~10GB", "Token-level activation caching"},
        {"Diffusion states", "~15GB", "ar_step intermediate results"},
        {"Iteration buffer", "~10GB", "Quality refinement passes"},
        {"Overflow", "~6GB", "Safety margin"},
    }
    ui.PrintTable(headers, rows)
}

func (a *Analysis) explainTeaCache() {
    ui.PrintHeader("TeaCache Explained")

    ui.PrintSection("What is TeaCache?")
    ui.PrintList("Token-level caching for diffusion transformers")
    ui.PrintList("Identifies and reuses redundant attention computations")
    ui.PrintList("Configurable threshold trades quality for speed")

    ui.PrintSection("Configuration")
    ui.PrintKeyValue("teacache_thresh=0.2", "Conservative, higher quality")
    ui.PrintKeyValue("teacache_thresh=0.3", "Balanced (recommended)")
    ui.PrintKeyValue("teacache_thresh=0.5", "Aggressive, faster but lower quality")
}

func (a *Analysis) explainNVLink() {
    ui.PrintHeader("NVLink Explained")

    ui.PrintSection("H100 NVLink Specs")
    ui.PrintKeyValue("Generation", "4th gen NVLink")
    ui.PrintKeyValue("Bandwidth", "900 GB/s bidirectional")
    ui.PrintKeyValue("Topology", "Full mesh (all-to-all)")

    ui.PrintSection("Why It Matters for SkyReel")
    ui.PrintList("Context parallel requires sync at attention boundaries")
    ui.PrintList("Each GPU needs to exchange attention keys/values")
    ui.PrintList("900 GB/s ensures communication isn't the bottleneck")
    ui.PrintList("PCIe would be 7× slower → would negate parallel gains")
}
