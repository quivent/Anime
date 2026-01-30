package metrics

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Metrics handles metrics display
type Metrics struct {
    Config *config.Config
}

// New creates a new Metrics instance
func New(cfg *config.Config) *Metrics {
    return &Metrics{Config: cfg}
}

// GPUMetrics represents GPU metrics
type GPUMetrics struct {
    Index       int
    Name        string
    Temperature int
    Utilization int
    MemoryUsed  int
    MemoryTotal int
    PowerDraw   int
    PowerLimit  int
}

// Print prints current metrics
func (m *Metrics) Print() {
    ui.PrintHeader("Current Metrics")

    gpuMetrics := m.getGPUMetrics()

    if len(gpuMetrics) == 0 {
        ui.PrintSuggestion("Unable to get GPU metrics", []string{
            "Ensure nvidia-smi is available",
            "Check that NVIDIA drivers are installed",
            "Run 'sky status' to check system status",
        })
        return
    }

    ui.PrintSection("GPU Status")

    for _, gpu := range gpuMetrics {
        fmt.Printf("\n  %s%s GPU %d: %s%s\n", ui.Bold, ui.BrightCyan, gpu.Index, gpu.Name, ui.Reset)

        // Temperature
        tempColor := ui.BrightGreen
        if gpu.Temperature > 70 {
            tempColor = ui.BrightYellow
        }
        if gpu.Temperature > 85 {
            tempColor = ui.BrightRed
        }
        fmt.Printf("    Temperature:  %s%d°C%s\n", tempColor, gpu.Temperature, ui.Reset)

        // Utilization
        utilBar := m.makeBar(gpu.Utilization, 100, 20)
        fmt.Printf("    Utilization:  [%s] %d%%\n", utilBar, gpu.Utilization)

        // Memory
        memPercent := (gpu.MemoryUsed * 100) / gpu.MemoryTotal
        memBar := m.makeBar(memPercent, 100, 20)
        fmt.Printf("    Memory:       [%s] %d/%dMB (%d%%)\n", memBar, gpu.MemoryUsed, gpu.MemoryTotal, memPercent)

        // Power
        powerPercent := (gpu.PowerDraw * 100) / gpu.PowerLimit
        powerBar := m.makeBar(powerPercent, 100, 20)
        fmt.Printf("    Power:        [%s] %dW/%dW\n", powerBar, gpu.PowerDraw, gpu.PowerLimit)
    }

    ui.PrintSection("Aggregate Metrics")

    totalMem := 0
    usedMem := 0
    avgUtil := 0
    avgTemp := 0
    totalPower := 0

    for _, gpu := range gpuMetrics {
        totalMem += gpu.MemoryTotal
        usedMem += gpu.MemoryUsed
        avgUtil += gpu.Utilization
        avgTemp += gpu.Temperature
        totalPower += gpu.PowerDraw
    }

    if len(gpuMetrics) > 0 {
        avgUtil /= len(gpuMetrics)
        avgTemp /= len(gpuMetrics)
    }

    ui.PrintKeyValue("Total VRAM", fmt.Sprintf("%d MB (%.1f GB)", totalMem, float64(totalMem)/1024))
    ui.PrintKeyValue("Used VRAM", fmt.Sprintf("%d MB (%.1f GB)", usedMem, float64(usedMem)/1024))
    ui.PrintKeyValue("Avg Utilization", fmt.Sprintf("%d%%", avgUtil))
    ui.PrintKeyValue("Avg Temperature", fmt.Sprintf("%d°C", avgTemp))
    ui.PrintKeyValue("Total Power", fmt.Sprintf("%dW", totalPower))
}

// PrintPerformance prints performance metrics based on config
func (m *Metrics) PrintPerformance() {
    ui.PrintHeader("Performance Metrics")

    ui.PrintSection("Configuration Impact")

    // Calculate theoretical performance
    basePerf := 100.0

    // Precision impact
    precisionFactor := 1.0
    switch m.Config.Model.Precision {
    case "fp8":
        precisionFactor = 1.4 // FP8 is faster
    case "fp16", "bf16":
        precisionFactor = 1.0
    case "fp32":
        precisionFactor = 0.5
    }

    // Parallelism impact
    parallelFactor := float64(m.Config.Parallelism.ContextParallel) * 0.85 // 85% scaling efficiency

    // Optimization impact
    optFactor := 1.0
    if m.Config.Optimization.TeaCacheEnabled {
        optFactor += 0.25
    }
    if m.Config.Optimization.CompileModel {
        optFactor += 0.1
    }

    // ar_step impact (negative - more steps = slower)
    arFactor := 1.0 - float64(m.Config.Diffusion.ARStep)*0.05

    totalPerf := basePerf * precisionFactor * parallelFactor * optFactor * arFactor

    headers := []string{"Factor", "Setting", "Impact"}
    rows := [][]string{
        {"Precision", m.Config.Model.Precision, fmt.Sprintf("×%.2f", precisionFactor)},
        {"Parallelism", fmt.Sprintf("CP%d", m.Config.Parallelism.ContextParallel), fmt.Sprintf("×%.2f", parallelFactor)},
        {"Optimizations", fmt.Sprintf("TeaCache=%v, Compile=%v", m.Config.Optimization.TeaCacheEnabled, m.Config.Optimization.CompileModel), fmt.Sprintf("×%.2f", optFactor)},
        {"ar_step", fmt.Sprintf("%d", m.Config.Diffusion.ARStep), fmt.Sprintf("×%.2f", arFactor)},
    }
    ui.PrintTable(headers, rows)

    ui.PrintSection("Estimated Performance")
    ui.PrintKeyValue("Relative Score", fmt.Sprintf("%.0f%%", totalPerf))
    ui.PrintKeyValue("Frames/sec (est)", fmt.Sprintf("%.2f", totalPerf/50)) // Rough estimate

    // Memory efficiency
    budget := m.Config.GetMemoryBudget()
    memEfficiency := float64(budget.HeadroomGB) / float64(budget.TotalGB) * 100

    ui.PrintSection("Memory Efficiency")
    ui.PrintKeyValue("Headroom", fmt.Sprintf("%dGB (%.0f%%)", budget.HeadroomGB, memEfficiency))
    ui.PrintProgressBar(budget.TotalGB-budget.HeadroomGB, budget.TotalGB, 50)

    if memEfficiency > 40 {
        ui.PrintStatus("success", "Excellent headroom for iteration and caching")
    } else if memEfficiency > 25 {
        ui.PrintStatus("info", "Good headroom for normal operation")
    } else {
        ui.PrintStatus("warning", "Limited headroom - consider FP8 or smaller model")
    }
}

func (m *Metrics) getGPUMetrics() []GPUMetrics {
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=index,name,temperature.gpu,utilization.gpu,memory.used,memory.total,power.draw,power.limit",
        "--format=csv,noheader,nounits")

    out, err := cmd.Output()
    if err != nil {
        return nil
    }

    var metrics []GPUMetrics
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")

    for _, line := range lines {
        parts := strings.Split(line, ", ")
        if len(parts) < 8 {
            continue
        }

        idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
        temp, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
        util, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
        memUsed, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
        memTotal, _ := strconv.Atoi(strings.TrimSpace(parts[5]))
        powerDraw, _ := strconv.ParseFloat(strings.TrimSpace(parts[6]), 64)
        powerLimit, _ := strconv.ParseFloat(strings.TrimSpace(parts[7]), 64)

        metrics = append(metrics, GPUMetrics{
            Index:       idx,
            Name:        strings.TrimSpace(parts[1]),
            Temperature: temp,
            Utilization: util,
            MemoryUsed:  memUsed,
            MemoryTotal: memTotal,
            PowerDraw:   int(powerDraw),
            PowerLimit:  int(powerLimit),
        })
    }

    return metrics
}

func (m *Metrics) makeBar(value, max, width int) string {
    if max == 0 {
        max = 1
    }
    filled := (value * width) / max
    if filled > width {
        filled = width
    }

    bar := ""
    color := ui.BrightGreen
    if value > 70 {
        color = ui.BrightYellow
    }
    if value > 90 {
        color = ui.BrightRed
    }

    for i := 0; i < width; i++ {
        if i < filled {
            bar += color + "█" + ui.Reset
        } else {
            bar += ui.Muted("░")
        }
    }
    return bar
}
