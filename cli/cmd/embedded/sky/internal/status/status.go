package status

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// StatusItem represents a status check
type StatusItem struct {
    Name     string
    Status   string // "done", "partial", "pending", "error"
    Message  string
    Details  []string
}

// Status handles status checking
type Status struct {
    Config *config.Config
}

// New creates a new Status instance
func New(cfg *config.Config) *Status {
    return &Status{Config: cfg}
}

// Check performs all status checks (fast mode by default)
func (s *Status) Check() []StatusItem {
    return s.CheckWithOptions(false)
}

// CheckFull performs all status checks including slow Python dependency checks
func (s *Status) CheckFull() []StatusItem {
    return s.CheckWithOptions(true)
}

// CheckWithOptions performs status checks with optional slow checks
func (s *Status) CheckWithOptions(includeSlow bool) []StatusItem {
    // Fast checks (run sequentially)
    items := []StatusItem{
        s.checkGPU(),
        s.checkGPULoad(),
        s.checkNVLink(),
        s.checkCUDA(),
        s.checkPython(),
    }

    if includeSlow {
        // Slow Python checks (~8-10s) - only run with --full flag
        pythonResult := s.checkPythonEnvFast()
        items = append(items, pythonResult.pytorch, pythonResult.deps)
    }

    // More fast checks
    items = append(items,
        s.checkModel(),
        s.checkConfig(),
    )

    return items
}

// pythonCheckResult holds results from combined Python check
type pythonCheckResult struct {
    pytorch StatusItem
    deps    StatusItem
}

// checkPythonEnvFast runs a single Python command to check PyTorch and all dependencies
func (s *Status) checkPythonEnvFast() pythonCheckResult {
    // Single Python script that checks everything
    script := `
import json
result = {"pytorch": None, "pytorch_cuda": False, "deps": {}}
try:
    import torch
    result["pytorch"] = torch.__version__
    result["pytorch_cuda"] = torch.cuda.is_available()
except:
    pass
for dep in ["xfuser", "flash_attn", "transformers", "diffusers"]:
    try:
        __import__(dep)
        result["deps"][dep] = True
    except:
        result["deps"][dep] = False
print(json.dumps(result))
`
    cmd := exec.Command("python3", "-c", script)
    out, err := cmd.Output()

    // Default results if Python fails
    pytorchItem := StatusItem{
        Name:    "PyTorch",
        Status:  "pending",
        Message: "PyTorch not installed",
    }
    depsItem := StatusItem{
        Name:    "Dependencies",
        Status:  "pending",
        Message: "Unable to check",
    }

    if err != nil {
        return pythonCheckResult{pytorch: pytorchItem, deps: depsItem}
    }

    // Parse JSON result
    var result struct {
        Pytorch     *string         `json:"pytorch"`
        PytorchCuda bool            `json:"pytorch_cuda"`
        Deps        map[string]bool `json:"deps"`
    }

    if err := json.Unmarshal(out, &result); err != nil {
        return pythonCheckResult{pytorch: pytorchItem, deps: depsItem}
    }

    // Build PyTorch result
    if result.Pytorch != nil {
        if result.PytorchCuda {
            pytorchItem = StatusItem{
                Name:    "PyTorch",
                Status:  "done",
                Message: *result.Pytorch + " (CUDA enabled)",
            }
        } else {
            pytorchItem = StatusItem{
                Name:    "PyTorch",
                Status:  "partial",
                Message: *result.Pytorch + " (no CUDA)",
                Details: []string{"Reinstall with CUDA support"},
            }
        }
    }

    // Build dependencies result
    installed := 0
    missing := []string{}
    for dep, ok := range result.Deps {
        if ok {
            installed++
        } else {
            missing = append(missing, dep)
        }
    }

    total := len(result.Deps)
    if installed == total {
        depsItem = StatusItem{
            Name:    "Dependencies",
            Status:  "done",
            Message: fmt.Sprintf("%d/%d installed", installed, total),
        }
    } else if installed > 0 {
        depsItem = StatusItem{
            Name:    "Dependencies",
            Status:  "partial",
            Message: fmt.Sprintf("%d/%d installed", installed, total),
            Details: []string{"Missing: " + strings.Join(missing, ", ")},
        }
    } else {
        depsItem = StatusItem{
            Name:    "Dependencies",
            Status:  "pending",
            Message: "No dependencies installed",
            Details: missing,
        }
    }

    return pythonCheckResult{pytorch: pytorchItem, deps: depsItem}
}

// Print prints the status report (fast mode - skips slow Python dependency checks)
func (s *Status) Print() {
    s.printWithOptions(false)
}

// PrintFull prints the full status report including slow Python dependency checks
func (s *Status) PrintFull() {
    s.printWithOptions(true)
}

func (s *Status) printWithOptions(includeSlow bool) {
    ui.PrintHeader("System Status")

    items := s.CheckWithOptions(includeSlow)

    if !includeSlow {
        fmt.Printf("  %s\n\n", ui.Muted("(fast mode - run 'sky status full' for complete check)"))
    }

    // Count statuses
    done, partial, pending, errored := 0, 0, 0, 0
    for _, item := range items {
        switch item.Status {
        case "done", "complete", "completed":
            done++
        case "partial":
            partial++
        case "pending":
            pending++
        case "error", "failed":
            errored++
        }
    }

    // Summary
    ui.PrintSection("Summary")
    total := len(items)
    ui.PrintProgressBar(done, total, 40)
    fmt.Printf("  %s/%d checks passed\n", ui.Success(fmt.Sprintf("%d", done)), total)
    if partial > 0 {
        fmt.Printf("  %s partial\n", ui.Warning(fmt.Sprintf("%d", partial)))
    }
    if pending > 0 {
        fmt.Printf("  %s pending\n", ui.Muted(fmt.Sprintf("%d", pending)))
    }
    if errored > 0 {
        fmt.Printf("  %s errors\n", ui.Error(fmt.Sprintf("%d", errored)))
    }

    // Details
    ui.PrintSection("Details")
    for _, item := range items {
        ui.PrintStatus(item.Status, fmt.Sprintf("%s: %s", item.Name, item.Message))
        for _, detail := range item.Details {
            fmt.Printf("      %s %s\n", ui.Muted("└"), ui.Muted(detail))
        }
    }

    // Suggestions
    s.printSuggestions(items)
}

func (s *Status) printSuggestions(items []StatusItem) {
    suggestions := []string{}

    for _, item := range items {
        switch item.Status {
        case "pending":
            switch item.Name {
            case "GPU Detection":
                suggestions = append(suggestions, "Install NVIDIA drivers and run 'nvidia-smi'")
            case "Model Files":
                suggestions = append(suggestions, "Run 'sky procedures show 3' for model download steps")
            case "Dependencies":
                suggestions = append(suggestions, "Run 'sky procedures show 4' for dependency installation")
            case "Configuration":
                suggestions = append(suggestions, "Run 'sky init' to create configuration")
            }
        case "partial":
            switch item.Name {
            case "GPU Detection":
                suggestions = append(suggestions, fmt.Sprintf("Only %s detected, expected %d GPUs", item.Message, s.Config.Hardware.GPUCount))
            case "NVLink":
                suggestions = append(suggestions, "Check NVLink cables and run 'nvidia-smi topo -m'")
            case "Model Loading":
                suggestions = append(suggestions, "Run 'sky reload' to fully load models on all GPUs")
            }
        case "error":
            switch item.Name {
            case "Model Loading":
                suggestions = append(suggestions, "Run 'sky reload' to load models onto GPUs")
            default:
                suggestions = append(suggestions, fmt.Sprintf("Fix %s: %s", item.Name, item.Message))
            }
        }
    }

    if len(suggestions) > 0 {
        ui.PrintSection("Recommendations")
        for i, sug := range suggestions {
            fmt.Printf("  %s%d.%s %s\n", ui.BrightCyan, i+1, ui.Reset, sug)
        }
    } else {
        ui.PrintSection("Next Steps")
        fmt.Printf("  %s All checks passed! Ready for video generation.\n", ui.Success(ui.Checkmark))
        fmt.Printf("  %s Run 'sky benchmark' to test throughput\n", ui.Info("→"))
    }
}

func (s *Status) checkGPU() StatusItem {
    out, err := exec.Command("nvidia-smi", "-L").Output()
    if err != nil {
        return StatusItem{
            Name:    "GPU Detection",
            Status:  "error",
            Message: "nvidia-smi not found or failed",
            Details: []string{"Install NVIDIA drivers"},
        }
    }

    lines := strings.Split(string(out), "\n")
    gpuCount := 0
    gpuNames := []string{}
    for _, line := range lines {
        if strings.Contains(line, "GPU ") {
            gpuCount++
            gpuNames = append(gpuNames, strings.TrimSpace(line))
        }
    }

    if gpuCount == 0 {
        return StatusItem{
            Name:    "GPU Detection",
            Status:  "pending",
            Message: "No GPUs detected",
        }
    }

    if gpuCount < s.Config.Hardware.GPUCount {
        return StatusItem{
            Name:    "GPU Detection",
            Status:  "partial",
            Message: fmt.Sprintf("%d/%d GPUs detected", gpuCount, s.Config.Hardware.GPUCount),
            Details: gpuNames,
        }
    }

    return StatusItem{
        Name:    "GPU Detection",
        Status:  "done",
        Message: fmt.Sprintf("%d GPUs detected", gpuCount),
        Details: gpuNames[:min(2, len(gpuNames))], // Show first 2
    }
}

// Minimum VRAM to consider model loaded (20GB per GPU for 14B model)
const MinVRAMForModelLoaded = 20000

func (s *Status) checkGPULoad() StatusItem {
    gpuLoads := s.getGPULoad()

    if len(gpuLoads) == 0 {
        return StatusItem{
            Name:    "Model Loading",
            Status:  "pending",
            Message: "Unable to check GPU memory",
        }
    }

    loadedCount := 0
    totalUsed := 0
    totalMem := 0

    for _, gpu := range gpuLoads {
        if gpu.MemoryUsed >= MinVRAMForModelLoaded {
            loadedCount++
        }
        totalUsed += gpu.MemoryUsed
        totalMem += gpu.MemoryTotal
    }

    usedGB := float64(totalUsed) / 1024
    totalGB := float64(totalMem) / 1024

    if loadedCount == 0 {
        return StatusItem{
            Name:   "Model Loading",
            Status: "error",
            Message: fmt.Sprintf("NOT loaded (%.0f/%.0f GB)", usedGB, totalGB),
            Details: []string{"Run 'sky reload' to load models"},
        }
    }

    if loadedCount < len(gpuLoads) {
        return StatusItem{
            Name:   "Model Loading",
            Status: "partial",
            Message: fmt.Sprintf("%d/%d GPUs loaded (%.0f/%.0f GB)", loadedCount, len(gpuLoads), usedGB, totalGB),
        }
    }

    return StatusItem{
        Name:   "Model Loading",
        Status: "done",
        Message: fmt.Sprintf("Loaded on %d GPUs (%.0f/%.0f GB)", loadedCount, usedGB, totalGB),
    }
}

func (s *Status) checkNVLink() StatusItem {
    out, err := exec.Command("nvidia-smi", "topo", "-m").Output()
    if err != nil {
        return StatusItem{
            Name:    "NVLink",
            Status:  "pending",
            Message: "Unable to check topology",
        }
    }

    output := string(out)
    nvlinkCount := strings.Count(output, "NV")

    if nvlinkCount == 0 {
        return StatusItem{
            Name:    "NVLink",
            Status:  "pending",
            Message: "No NVLink connections detected",
            Details: []string{"GPUs may be using PCIe only"},
        }
    }

    // Check for NV18 (H100 full NVLink)
    if strings.Contains(output, "NV18") {
        return StatusItem{
            Name:    "NVLink",
            Status:  "done",
            Message: "Full NVLink mesh (NV18)",
            Details: []string{"900 GB/s bidirectional"},
        }
    }

    return StatusItem{
        Name:    "NVLink",
        Status:  "partial",
        Message: "NVLink detected (partial)",
        Details: []string{"May not be full mesh"},
    }
}

func (s *Status) checkCUDA() StatusItem {
    out, err := exec.Command("nvcc", "--version").Output()
    if err != nil {
        // Try checking with nvidia-smi
        out2, err2 := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader").Output()
        if err2 != nil {
            return StatusItem{
                Name:    "CUDA",
                Status:  "pending",
                Message: "CUDA toolkit not found",
                Details: []string{"Install CUDA 12.x"},
            }
        }
        return StatusItem{
            Name:    "CUDA",
            Status:  "partial",
            Message: "Driver only: " + strings.TrimSpace(string(out2)),
            Details: []string{"Install CUDA toolkit for nvcc"},
        }
    }

    version := ""
    for _, line := range strings.Split(string(out), "\n") {
        if strings.Contains(line, "release") {
            parts := strings.Split(line, "release ")
            if len(parts) > 1 {
                version = strings.Split(parts[1], ",")[0]
            }
        }
    }

    return StatusItem{
        Name:    "CUDA",
        Status:  "done",
        Message: "CUDA " + version,
    }
}

func (s *Status) checkPython() StatusItem {
    out, err := exec.Command("python3", "--version").Output()
    if err != nil {
        return StatusItem{
            Name:    "Python",
            Status:  "pending",
            Message: "Python 3 not found",
        }
    }

    version := strings.TrimSpace(strings.TrimPrefix(string(out), "Python "))

    // Check if in venv
    venv := os.Getenv("VIRTUAL_ENV")
    details := []string{}
    if venv != "" {
        details = append(details, "venv: "+filepath.Base(venv))
    }

    return StatusItem{
        Name:    "Python",
        Status:  "done",
        Message: version,
        Details: details,
    }
}

func (s *Status) checkPyTorch() StatusItem {
    cmd := exec.Command("python3", "-c", "import torch; print(torch.__version__, torch.cuda.is_available())")
    out, err := cmd.Output()
    if err != nil {
        return StatusItem{
            Name:    "PyTorch",
            Status:  "pending",
            Message: "PyTorch not installed",
            Details: []string{"pip install torch"},
        }
    }

    parts := strings.Fields(strings.TrimSpace(string(out)))
    if len(parts) < 2 {
        return StatusItem{
            Name:    "PyTorch",
            Status:  "partial",
            Message: "Unable to parse version",
        }
    }

    version := parts[0]
    cudaAvail := parts[1] == "True"

    if !cudaAvail {
        return StatusItem{
            Name:    "PyTorch",
            Status:  "partial",
            Message: version + " (no CUDA)",
            Details: []string{"Reinstall with CUDA support"},
        }
    }

    return StatusItem{
        Name:    "PyTorch",
        Status:  "done",
        Message: version + " (CUDA enabled)",
    }
}

func (s *Status) checkModel() StatusItem {
    modelPath := s.Config.Paths.ModelPath

    // Check if path exists
    if _, err := os.Stat(modelPath); os.IsNotExist(err) {
        return StatusItem{
            Name:    "Model Files",
            Status:  "pending",
            Message: "Model directory not found",
            Details: []string{modelPath},
        }
    }

    // Check for model files
    patterns := []string{"*.safetensors", "*.bin", "config.json"}
    found := 0
    for _, pattern := range patterns {
        matches, _ := filepath.Glob(filepath.Join(modelPath, pattern))
        if len(matches) > 0 {
            found++
        }
    }

    if found == 0 {
        return StatusItem{
            Name:    "Model Files",
            Status:  "pending",
            Message: "No model files found",
            Details: []string{"Download model from HuggingFace"},
        }
    }

    if found < len(patterns) {
        return StatusItem{
            Name:    "Model Files",
            Status:  "partial",
            Message: "Some model files present",
            Details: []string{modelPath},
        }
    }

    return StatusItem{
        Name:    "Model Files",
        Status:  "done",
        Message: s.Config.Model.Variant,
        Details: []string{modelPath},
    }
}

func (s *Status) checkDependencies() StatusItem {
    deps := []string{"xfuser", "flash_attn", "transformers", "diffusers"}
    installed := []string{}
    missing := []string{}

    for _, dep := range deps {
        cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", dep))
        if err := cmd.Run(); err != nil {
            missing = append(missing, dep)
        } else {
            installed = append(installed, dep)
        }
    }

    if len(missing) == len(deps) {
        return StatusItem{
            Name:    "Dependencies",
            Status:  "pending",
            Message: "No dependencies installed",
            Details: missing,
        }
    }

    if len(missing) > 0 {
        return StatusItem{
            Name:    "Dependencies",
            Status:  "partial",
            Message: fmt.Sprintf("%d/%d installed", len(installed), len(deps)),
            Details: []string{"Missing: " + strings.Join(missing, ", ")},
        }
    }

    return StatusItem{
        Name:    "Dependencies",
        Status:  "done",
        Message: fmt.Sprintf("%d/%d installed", len(installed), len(deps)),
    }
}

func (s *Status) checkConfig() StatusItem {
    configPath := config.ConfigPath()

    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return StatusItem{
            Name:    "Configuration",
            Status:  "pending",
            Message: "No config file",
            Details: []string{"Run 'sky init' to create"},
        }
    }

    return StatusItem{
        Name:    "Configuration",
        Status:  "done",
        Message: "Config loaded",
        Details: []string{configPath},
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// GPULoadInfo represents GPU load information
type GPULoadInfo struct {
    Index       int
    Name        string
    MemoryUsed  int
    MemoryTotal int
    Utilization int
}

// PrintLoad prints GPU load information (memory and utilization)
func (s *Status) PrintLoad() {
    ui.PrintHeader("GPU Load")

    gpuLoads := s.getGPULoad()

    if len(gpuLoads) == 0 {
        ui.PrintSuggestion("Unable to get GPU load", []string{
            "Ensure nvidia-smi is available",
            "Check that NVIDIA drivers are installed",
        })
        return
    }

    ui.PrintSection("Memory & Utilization")

    totalUsed := 0
    totalMem := 0
    avgUtil := 0

    for _, gpu := range gpuLoads {
        memPercent := 0
        if gpu.MemoryTotal > 0 {
            memPercent = (gpu.MemoryUsed * 100) / gpu.MemoryTotal
        }

        // Color based on memory usage
        memColor := ui.BrightGreen
        if memPercent > 70 {
            memColor = ui.BrightYellow
        }
        if memPercent > 90 {
            memColor = ui.BrightRed
        }

        // Color based on utilization
        utilColor := ui.BrightGreen
        if gpu.Utilization > 70 {
            utilColor = ui.BrightYellow
        }
        if gpu.Utilization > 90 {
            utilColor = ui.BrightRed
        }

        // Memory bar
        memBar := s.makeLoadBar(memPercent, 100, 15)

        // Utilization bar
        utilBar := s.makeLoadBar(gpu.Utilization, 100, 15)

        fmt.Printf("  %sGPU %d%s %s\n", ui.Bold, gpu.Index, ui.Reset, ui.Muted(gpu.Name))
        fmt.Printf("    Memory:      [%s] %s%5d%s / %5d MB %s(%d%%)%s\n",
            memBar, memColor, gpu.MemoryUsed, ui.Reset, gpu.MemoryTotal, ui.Muted(""), memPercent, ui.Reset)
        fmt.Printf("    Utilization: [%s] %s%3d%%%s\n",
            utilBar, utilColor, gpu.Utilization, ui.Reset)
        fmt.Println()

        totalUsed += gpu.MemoryUsed
        totalMem += gpu.MemoryTotal
        avgUtil += gpu.Utilization
    }

    if len(gpuLoads) > 0 {
        avgUtil /= len(gpuLoads)
    }

    ui.PrintSection("Summary")
    totalPercent := 0
    if totalMem > 0 {
        totalPercent = (totalUsed * 100) / totalMem
    }
    ui.PrintKeyValue("Total Memory", fmt.Sprintf("%d / %d MB (%d%%)", totalUsed, totalMem, totalPercent))
    ui.PrintKeyValue("Avg Utilization", fmt.Sprintf("%d%%", avgUtil))

    // Model loaded indicator
    if totalUsed > 1000 {
        ui.PrintStatus("success", "Models appear to be loaded")
    } else {
        ui.PrintStatus("info", "Low memory usage - models may not be loaded")
    }
}

// PrintGPUs prints a summary of detected GPUs
func (s *Status) PrintGPUs() {
    ui.PrintHeader("GPU Information")

    out, err := exec.Command("nvidia-smi",
        "--query-gpu=index,name,memory.total,driver_version,pci.bus_id",
        "--format=csv,noheader,nounits").Output()

    if err != nil {
        ui.PrintSuggestion("Unable to detect GPUs", []string{
            "Ensure nvidia-smi is available",
            "Check that NVIDIA drivers are installed",
        })
        return
    }

    lines := strings.Split(strings.TrimSpace(string(out)), "\n")

    ui.PrintSection(fmt.Sprintf("Detected GPUs (%d)", len(lines)))

    for _, line := range lines {
        parts := strings.Split(line, ", ")
        if len(parts) < 5 {
            continue
        }

        idx := strings.TrimSpace(parts[0])
        name := strings.TrimSpace(parts[1])
        memTotal := strings.TrimSpace(parts[2])
        driver := strings.TrimSpace(parts[3])
        pci := strings.TrimSpace(parts[4])

        fmt.Printf("  %sGPU %s%s: %s\n", ui.BrightCyan, idx, ui.Reset, name)
        fmt.Printf("    %s Memory:  %s MB\n", ui.Muted("├"), memTotal)
        fmt.Printf("    %s Driver:  %s\n", ui.Muted("├"), driver)
        fmt.Printf("    %s PCI:     %s\n", ui.Muted("└"), pci)
        fmt.Println()
    }

    // NVLink status
    topoOut, err := exec.Command("nvidia-smi", "topo", "-m").Output()
    if err == nil {
        output := string(topoOut)
        if strings.Contains(output, "NV") {
            nvCount := strings.Count(output, "NV")
            ui.PrintStatus("success", fmt.Sprintf("NVLink detected (%d connections)", nvCount))
        } else {
            ui.PrintStatus("info", "No NVLink connections (PCIe only)")
        }
    }
}

func (s *Status) getGPULoad() []GPULoadInfo {
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=index,name,memory.used,memory.total,utilization.gpu",
        "--format=csv,noheader,nounits")

    out, err := cmd.Output()
    if err != nil {
        return nil
    }

    var loads []GPULoadInfo
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")

    for _, line := range lines {
        parts := strings.Split(line, ", ")
        if len(parts) < 5 {
            continue
        }

        idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
        memUsed, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
        memTotal, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
        util, _ := strconv.Atoi(strings.TrimSpace(parts[4]))

        loads = append(loads, GPULoadInfo{
            Index:       idx,
            Name:        strings.TrimSpace(parts[1]),
            MemoryUsed:  memUsed,
            MemoryTotal: memTotal,
            Utilization: util,
        })
    }

    return loads
}

func (s *Status) makeLoadBar(value, max, width int) string {
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
