package benchmark

import (
    "fmt"
    "time"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// BenchmarkResult represents benchmark results
type BenchmarkResult struct {
    Name           string
    Duration       time.Duration
    FramesGenerated int
    FPS            float64
    MemoryPeakGB   float64
    GPUUtilization []float64
    Success        bool
    Error          string
}

// Benchmark handles benchmarking
type Benchmark struct {
    Config *config.Config
}

// New creates a new Benchmark instance
func New(cfg *config.Config) *Benchmark {
    return &Benchmark{Config: cfg}
}

// PrintInfo prints benchmark information and estimates
func (b *Benchmark) PrintInfo() {
    ui.PrintHeader("Benchmark Information")

    ui.PrintSection("Available Benchmarks")

    headers := []string{"Name", "Frames", "Resolution", "ar_step", "Est. Time"}
    rows := [][]string{
        {"quick", "25", "544×960", "0", "~20-30s"},
        {"standard", "97", "544×960", "5", "~45-60s"},
        {"extended", "289", "544×960", "5", "~90-120s"},
        {"hd", "97", "720×1280", "5", "~80-100s"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("How to Run")
    fmt.Printf("  %s Run benchmark:      %s\n", ui.Muted("$"), ui.Value("sky benchmark run <name>"))
    fmt.Printf("  %s Run all benchmarks: %s\n", ui.Muted("$"), ui.Value("sky benchmark run all"))
    fmt.Printf("  %s Custom benchmark:   %s\n", ui.Muted("$"), ui.Value("sky benchmark run --frames 50 --ar-step 3"))

    ui.PrintSection("Expected Results (4×H100)")
    ui.PrintSubSection("Throughput Estimates")
    ui.PrintList("Quick (25 frames): ~0.8-1.2 frames/sec")
    ui.PrintList("Standard (97 frames): ~1.5-2.0 frames/sec")
    ui.PrintList("With TeaCache: +20-30% improvement")

    ui.PrintSubSection("Memory Usage")
    ui.PrintList("14B FP8: ~39GB peak per GPU")
    ui.PrintList("14B FP16: ~55GB peak per GPU")
    ui.PrintList("1.3B FP8: ~18GB peak per GPU")
}

// PrintEstimate prints estimated performance
func (b *Benchmark) PrintEstimate() {
    ui.PrintHeader("Performance Estimates")

    ui.PrintSection("Configuration")
    ui.PrintKeyValue("Model", b.Config.Model.Variant)
    ui.PrintKeyValue("Precision", b.Config.Model.Precision)
    ui.PrintKeyValue("GPUs", fmt.Sprintf("%d× %s", b.Config.Hardware.GPUCount, b.Config.Hardware.GPUModel))
    ui.PrintKeyValue("Context Parallel", fmt.Sprintf("%d", b.Config.Parallelism.ContextParallel))
    ui.PrintKeyValue("ar_step", fmt.Sprintf("%d", b.Config.Diffusion.ARStep))
    ui.PrintKeyValue("TeaCache", fmt.Sprintf("%v", b.Config.Optimization.TeaCacheEnabled))

    // Calculate estimates
    baseTime := 80.0 // Single 4090 baseline for 97 frames

    // GPU scaling factor (H100 vs 4090)
    gpuFactor := 0.6 // H100 is roughly 1.5-2x faster per GPU

    // Multi-GPU scaling
    multiGPUFactor := 0.42 // 4x GPUs give ~58% reduction (42% of original time)

    // ar_step factor
    arStepFactor := 1.0
    if b.Config.Diffusion.ARStep > 0 {
        arStepFactor = 1.0 + float64(b.Config.Diffusion.ARStep)*0.08
    }

    // TeaCache factor
    teaCacheFactor := 1.0
    if b.Config.Optimization.TeaCacheEnabled {
        teaCacheFactor = 0.75 // 25% speedup
    }

    // FP8 factor (already in baseline for memory, slight speed gain)
    fp8Factor := 1.0
    if b.Config.Optimization.FP8Quantization {
        fp8Factor = 0.95
    }

    // Calculate final estimate
    estimatedTime := baseTime * gpuFactor * multiGPUFactor * arStepFactor * teaCacheFactor * fp8Factor
    framesPerSec := float64(b.Config.Generation.BaseNumFrames) / estimatedTime

    ui.PrintSection("Estimates for 97 Frames (4s video)")
    ui.PrintKeyValue("Generation Time", fmt.Sprintf("%.1f - %.1f seconds", estimatedTime*0.85, estimatedTime*1.15))
    ui.PrintKeyValue("Throughput", fmt.Sprintf("%.2f frames/sec", framesPerSec))
    ui.PrintKeyValue("Videos/Hour", fmt.Sprintf("%.0f (4s clips)", 3600/estimatedTime))

    ui.PrintSection("Breakdown")
    headers := []string{"Factor", "Impact", "Value"}
    rows := [][]string{
        {"Base (4090)", "Reference", fmt.Sprintf("%.0fs", baseTime)},
        {"H100 Performance", fmt.Sprintf("%.0f%%", (1-gpuFactor)*100), fmt.Sprintf("×%.2f", gpuFactor)},
        {"4-GPU Parallel", fmt.Sprintf("%.0f%%", (1-multiGPUFactor)*100), fmt.Sprintf("×%.2f", multiGPUFactor)},
        {"ar_step=" + fmt.Sprintf("%d", b.Config.Diffusion.ARStep), fmt.Sprintf("+%.0f%%", (arStepFactor-1)*100), fmt.Sprintf("×%.2f", arStepFactor)},
        {"TeaCache", fmt.Sprintf("-%.0f%%", (1-teaCacheFactor)*100), fmt.Sprintf("×%.2f", teaCacheFactor)},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("Memory Estimate")
    budget := b.Config.GetMemoryBudget()
    ui.PrintProgressBar(budget.TotalGB-budget.HeadroomGB, budget.TotalGB, 50)
    ui.PrintKeyValue("Peak Usage", fmt.Sprintf("~%dGB per GPU", budget.TotalGB-budget.HeadroomGB))
    ui.PrintKeyValue("Available Headroom", ui.Success(fmt.Sprintf("~%dGB", budget.HeadroomGB)))

    ui.PrintSection("Scaling Projections")
    headers = []string{"Video Length", "Frames", "Est. Time", "Throughput"}
    rows = [][]string{
        {"2 seconds", "49", fmt.Sprintf("%.0fs", estimatedTime*0.5), fmt.Sprintf("%.2f fps", framesPerSec*1.1)},
        {"4 seconds", "97", fmt.Sprintf("%.0fs", estimatedTime), fmt.Sprintf("%.2f fps", framesPerSec)},
        {"8 seconds", "193", fmt.Sprintf("%.0fs", estimatedTime*2.1), fmt.Sprintf("%.2f fps", framesPerSec*0.95)},
        {"12 seconds", "289", fmt.Sprintf("%.0fs", estimatedTime*3.2), fmt.Sprintf("%.2f fps", framesPerSec*0.9)},
    }
    ui.PrintTable(headers, rows)
}

// PrintSimulated prints simulated benchmark results
func (b *Benchmark) PrintSimulated(benchmarkName string) {
    ui.PrintHeader(fmt.Sprintf("Benchmark: %s (Simulated)", benchmarkName))

    ui.PrintSuggestion("Actual benchmark not run", []string{
        "This shows expected results based on configuration",
        "To run actual benchmark, ensure SkyReel is installed",
        "Run 'sky status' to check setup completion",
    })

    var frames int
    var arStep int
    var resolution string

    switch benchmarkName {
    case "quick":
        frames = 25
        arStep = 0
        resolution = "544×960"
    case "standard":
        frames = 97
        arStep = 5
        resolution = "544×960"
    case "extended":
        frames = 289
        arStep = 5
        resolution = "544×960"
    case "hd":
        frames = 97
        arStep = 5
        resolution = "720×1280"
    default:
        frames = 97
        arStep = b.Config.Diffusion.ARStep
        resolution = fmt.Sprintf("%d×%d", b.Config.Generation.Height, b.Config.Generation.Width)
    }

    // Simulate results
    baseTimePerFrame := 0.5 // seconds per frame on 4×H100
    if arStep > 0 {
        baseTimePerFrame *= 1.4
    }
    if resolution == "720×1280" {
        baseTimePerFrame *= 1.5
    }
    if b.Config.Optimization.TeaCacheEnabled {
        baseTimePerFrame *= 0.75
    }

    totalTime := baseTimePerFrame * float64(frames)
    fps := float64(frames) / totalTime

    ui.PrintSection("Simulated Results")
    ui.PrintKeyValue("Frames", fmt.Sprintf("%d", frames))
    ui.PrintKeyValue("Resolution", resolution)
    ui.PrintKeyValue("ar_step", fmt.Sprintf("%d", arStep))
    ui.PrintKeyValue("Total Time", fmt.Sprintf("%.1f seconds", totalTime))
    ui.PrintKeyValue("Throughput", fmt.Sprintf("%.2f frames/sec", fps))

    ui.PrintSection("GPU Utilization (Simulated)")
    for i := 0; i < b.Config.Hardware.GPUCount; i++ {
        util := 85 + i*2 // Simulated utilization
        bar := ""
        for j := 0; j < util/5; j++ {
            bar += "█"
        }
        for j := util / 5; j < 20; j++ {
            bar += "░"
        }
        fmt.Printf("  GPU %d: [%s%s%s] %d%%\n", i, ui.BrightGreen, bar, ui.Reset, util)
    }

    ui.PrintSection("Memory Usage (Simulated)")
    used := b.Config.GetMemoryBudget().TotalGB - b.Config.GetMemoryBudget().HeadroomGB
    peak := used + 5 // Simulate some overhead
    ui.PrintKeyValue("Baseline", fmt.Sprintf("%dGB", used))
    ui.PrintKeyValue("Peak", fmt.Sprintf("%dGB", peak))
    ui.PrintKeyValue("Headroom Used", fmt.Sprintf("%dGB of %dGB", 5, b.Config.GetMemoryBudget().HeadroomGB))
}

// PrintComparison prints benchmark comparison
func (b *Benchmark) PrintComparison() {
    ui.PrintHeader("Benchmark Comparison")

    ui.PrintSection("Configuration Variants")

    headers := []string{"Config", "ar_step", "TeaCache", "Est. Time", "Quality"}
    rows := [][]string{
        {"Fast", "0", "0.5", "~25s", "Good"},
        {"Balanced", "5", "0.3", "~45s", "Better"},
        {"Quality", "5", "0.2", "~55s", "Best"},
        {"Max Quality", "10", "off", "~80s", "Maximum"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("Hardware Scaling")
    headers = []string{"GPUs", "Time", "Efficiency", "Cost/Video"}
    rows = [][]string{
        {"1× H100", "~150s", "100%", "$$$"},
        {"2× H100", "~85s", "88%", "$$"},
        {"4× H100", "~45s", "83%", "$"},
        {"8× H100", "~28s", "67%", "$$"},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("Model Variants")
    headers = []string{"Model", "VRAM", "Speed", "Quality"}
    rows = [][]string{
        {"1.3B FP8", "~15GB", "Fast", "Good"},
        {"5B FP8", "~22GB", "Medium", "Better"},
        {"14B FP8", "~39GB", "Slower", "Best"},
        {"14B FP16", "~55GB", "Slower", "Best+"},
    }
    ui.PrintTable(headers, rows)
}
