package load

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sky-cli/sky/internal/config"
	"github.com/sky-cli/sky/ui"
)

const (
	SkyReelsDir         = "/home/ubuntu/SkyReels-V2"
	MinVRAMForModelLoad = 20000 // 20GB per GPU
)

// LoadConfig represents the configuration for model loading
type LoadConfig struct {
	GPUCount        int
	Parallelism     string // "tp" (tensor parallel) or "cp" (context parallel)
	ParallelDegree  int    // 1, 2, 4, 8
	Model           string // model variant
	Precision       string // fp16, bf16, fp8
	OffloadEnabled  bool
	TeaCacheEnabled bool
}

// Loader handles model loading wizard
type Loader struct {
	Config    *config.Config
	GPUCount  int
	GPUMemory int // MB per GPU
}

// GPUInfo holds detected GPU information
type GPUInfo struct {
	Index      int
	Name       string
	MemoryMB   int
	MemoryUsed int
}

// New creates a new Loader instance
func New(cfg *config.Config) *Loader {
	return &Loader{Config: cfg}
}

// RunWizard runs the interactive load wizard
func (l *Loader) RunWizard() {
	ui.PrintHeader("Model Loading Wizard")

	// Step 1: Detect GPUs
	gpus := l.detectGPUs()
	if len(gpus) == 0 {
		ui.PrintStatus("error", "No GPUs detected")
		ui.PrintSuggestion("GPU detection failed", []string{
			"Ensure NVIDIA drivers are installed",
			"Run 'nvidia-smi' to verify GPU access",
		})
		return
	}

	l.GPUCount = len(gpus)
	if len(gpus) > 0 {
		l.GPUMemory = gpus[0].MemoryMB
	}

	l.printGPUStatus(gpus)

	// Check if models already loaded
	loadedCount := 0
	for _, gpu := range gpus {
		if gpu.MemoryUsed >= MinVRAMForModelLoad {
			loadedCount++
		}
	}

	if loadedCount == len(gpus) {
		ui.PrintStatus("success", "Models already loaded on all GPUs")
		fmt.Println()
		if !l.confirm("Do you want to reload anyway?") {
			return
		}
	}

	// Step 2: Select parallelism strategy
	ui.PrintSection("Parallelism Strategy")
	loadCfg := l.selectParallelism()

	// Step 3: Select model variant
	ui.PrintSection("Model Selection")
	loadCfg.Model = l.selectModel()

	// Step 4: Select precision
	ui.PrintSection("Precision")
	loadCfg.Precision = l.selectPrecision()

	// Step 5: Optimizations
	ui.PrintSection("Optimizations")
	loadCfg.TeaCacheEnabled = l.selectTeaCache()
	loadCfg.OffloadEnabled = l.selectOffload()

	// Step 6: Summary and confirm
	l.printSummary(loadCfg)

	if !l.confirm("Proceed with model loading?") {
		ui.PrintStatus("info", "Cancelled")
		return
	}

	// Step 7: Execute loading
	l.executeLoad(loadCfg)
}

func (l *Loader) detectGPUs() []GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,memory.total,memory.used",
		"--format=csv,noheader,nounits")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) < 4 {
			continue
		}

		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		memTotal, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		memUsed, _ := strconv.Atoi(strings.TrimSpace(parts[3]))

		gpus = append(gpus, GPUInfo{
			Index:      idx,
			Name:       strings.TrimSpace(parts[1]),
			MemoryMB:   memTotal,
			MemoryUsed: memUsed,
		})
	}

	return gpus
}

func (l *Loader) printGPUStatus(gpus []GPUInfo) {
	ui.PrintSection(fmt.Sprintf("Detected GPUs (%d)", len(gpus)))

	for _, gpu := range gpus {
		status := ui.Muted("empty")
		if gpu.MemoryUsed >= MinVRAMForModelLoad {
			status = ui.Success("loaded")
		} else if gpu.MemoryUsed > 1000 {
			status = ui.Warning("partial")
		}

		memGB := float64(gpu.MemoryMB) / 1024
		usedGB := float64(gpu.MemoryUsed) / 1024

		fmt.Printf("  %sGPU %d%s: %s  [%.0fGB / %.0fGB] %s\n",
			ui.BrightCyan, gpu.Index, ui.Reset,
			gpu.Name, usedGB, memGB, status)
	}
	fmt.Println()
}

func (l *Loader) selectParallelism() LoadConfig {
	cfg := LoadConfig{GPUCount: l.GPUCount}

	// Show options based on GPU count
	fmt.Println("  Choose how to distribute the model across GPUs:")
	fmt.Println()

	options := []struct {
		key         string
		name        string
		desc        string
		parallelism string
		degree      int
		recommended bool
	}{}

	switch l.GPUCount {
	case 1:
		options = append(options,
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"1", "Single GPU", "Full model on one GPU (requires 80GB+)", "none", 1, true},
		)
	case 2:
		options = append(options,
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"1", "TP2", "Tensor Parallel across 2 GPUs", "tp", 2, false},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"2", "CP2", "Context Parallel across 2 GPUs (recommended)", "cp", 2, true},
		)
	case 4:
		options = append(options,
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"1", "TP4", "Tensor Parallel across 4 GPUs", "tp", 4, false},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"2", "CP4", "Context Parallel across 4 GPUs (recommended)", "cp", 4, true},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"3", "TP2×CP2", "Hybrid: TP2 with CP2", "hybrid", 4, false},
		)
	case 8:
		options = append(options,
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"1", "TP8", "Tensor Parallel across 8 GPUs", "tp", 8, false},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"2", "CP8", "Context Parallel across 8 GPUs", "cp", 8, false},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"3", "TP4×CP2", "Hybrid: TP4 with CP2 (recommended)", "hybrid", 8, true},
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"4", "TP2×CP4", "Hybrid: TP2 with CP4", "hybrid", 8, false},
		)
	default:
		// Fallback for other GPU counts
		options = append(options,
			struct {
				key         string
				name        string
				desc        string
				parallelism string
				degree      int
				recommended bool
			}{"1", fmt.Sprintf("CP%d", l.GPUCount), fmt.Sprintf("Context Parallel across %d GPUs", l.GPUCount), "cp", l.GPUCount, true},
		)
	}

	// Print options
	for _, opt := range options {
		rec := ""
		if opt.recommended {
			rec = ui.Success(" [recommended]")
		}
		fmt.Printf("    %s[%s]%s %s%s\n", ui.BrightCyan, opt.key, ui.Reset, opt.name, rec)
		fmt.Printf("         %s\n", ui.Muted(opt.desc))
	}

	// Explain the difference
	fmt.Println()
	fmt.Printf("  %s\n", ui.Muted("TP = splits model layers across GPUs (lower latency)"))
	fmt.Printf("  %s\n", ui.Muted("CP = splits sequence across GPUs (better for long videos)"))
	fmt.Println()

	// Get selection
	choice := l.prompt("Select parallelism", "2")

	// Find selected option
	for _, opt := range options {
		if opt.key == choice {
			cfg.Parallelism = opt.parallelism
			cfg.ParallelDegree = opt.degree
			ui.PrintStatus("success", fmt.Sprintf("Selected: %s", opt.name))
			return cfg
		}
	}

	// Default to first recommended
	for _, opt := range options {
		if opt.recommended {
			cfg.Parallelism = opt.parallelism
			cfg.ParallelDegree = opt.degree
			ui.PrintStatus("success", fmt.Sprintf("Selected: %s (default)", opt.name))
			return cfg
		}
	}

	return cfg
}

func (l *Loader) selectModel() string {
	fmt.Println("  Available models:")
	fmt.Println()

	models := []struct {
		key         string
		name        string
		size        string
		recommended bool
	}{
		{"1", "SkyReels-V2-DF-14B-540P", "14B params, 540P optimized", true},
		{"2", "SkyReels-V2-DF-14B-720P", "14B params, 720P optimized", false},
		{"3", "SkyReels-V2-DF-1.3B-540P", "1.3B params, fast inference", false},
	}

	for _, m := range models {
		rec := ""
		if m.recommended {
			rec = ui.Success(" [recommended]")
		}
		fmt.Printf("    %s[%s]%s %s%s\n", ui.BrightCyan, m.key, ui.Reset, m.name, rec)
		fmt.Printf("         %s\n", ui.Muted(m.size))
	}
	fmt.Println()

	choice := l.prompt("Select model", "1")

	for _, m := range models {
		if m.key == choice {
			ui.PrintStatus("success", fmt.Sprintf("Selected: %s", m.name))
			return m.name
		}
	}

	ui.PrintStatus("success", "Selected: SkyReels-V2-DF-14B-540P (default)")
	return "SkyReels-V2-DF-14B-540P"
}

func (l *Loader) selectPrecision() string {
	fmt.Println("  Available precisions:")
	fmt.Println()

	options := []struct {
		key         string
		name        string
		desc        string
		recommended bool
	}{
		{"1", "bf16", "BFloat16 - Best quality, standard memory", true},
		{"2", "fp16", "Float16 - Good quality, standard memory", false},
		{"3", "fp8", "Float8 - Faster, ~40% less memory, slight quality loss", false},
	}

	for _, opt := range options {
		rec := ""
		if opt.recommended {
			rec = ui.Success(" [recommended]")
		}
		fmt.Printf("    %s[%s]%s %s%s\n", ui.BrightCyan, opt.key, ui.Reset, opt.name, rec)
		fmt.Printf("         %s\n", ui.Muted(opt.desc))
	}
	fmt.Println()

	choice := l.prompt("Select precision", "1")

	for _, opt := range options {
		if opt.key == choice {
			ui.PrintStatus("success", fmt.Sprintf("Selected: %s", opt.name))
			return opt.name
		}
	}

	ui.PrintStatus("success", "Selected: bf16 (default)")
	return "bf16"
}

func (l *Loader) selectTeaCache() bool {
	fmt.Println("  TeaCache accelerates inference by caching activations.")
	fmt.Printf("  %s\n", ui.Muted("Provides ~20-30% speedup with minimal quality impact"))
	fmt.Println()

	return l.confirm("Enable TeaCache?")
}

func (l *Loader) selectOffload() bool {
	// Only suggest offload if memory might be tight
	totalMem := l.GPUCount * l.GPUMemory
	if totalMem >= 320000 { // 320GB total (4×80GB)
		fmt.Printf("  %s\n", ui.Muted("CPU offloading not needed with your GPU memory"))
		return false
	}

	fmt.Println("  CPU Offloading moves some weights to system RAM.")
	fmt.Printf("  %s\n", ui.Muted("Reduces GPU memory but increases latency"))
	fmt.Println()

	return l.confirm("Enable CPU offloading?")
}

func (l *Loader) printSummary(cfg LoadConfig) {
	ui.PrintSection("Configuration Summary")

	fmt.Println()
	ui.PrintKeyValue("GPUs", fmt.Sprintf("%d", cfg.GPUCount))
	ui.PrintKeyValue("Parallelism", fmt.Sprintf("%s (degree=%d)", strings.ToUpper(cfg.Parallelism), cfg.ParallelDegree))
	ui.PrintKeyValue("Model", cfg.Model)
	ui.PrintKeyValue("Precision", cfg.Precision)
	ui.PrintKeyValue("TeaCache", fmt.Sprintf("%v", cfg.TeaCacheEnabled))
	ui.PrintKeyValue("CPU Offload", fmt.Sprintf("%v", cfg.OffloadEnabled))

	// Estimate memory usage
	baseMemPerGPU := 28000 // ~28GB for 14B model with CP4
	if strings.Contains(cfg.Model, "1.3B") {
		baseMemPerGPU = 5000
	}
	if cfg.Precision == "fp8" {
		baseMemPerGPU = int(float64(baseMemPerGPU) * 0.6)
	}

	estMemPerGPU := baseMemPerGPU / cfg.ParallelDegree * cfg.GPUCount
	fmt.Println()
	ui.PrintKeyValue("Est. Memory/GPU", fmt.Sprintf("~%dGB", estMemPerGPU/1024))
	fmt.Println()
}

func (l *Loader) executeLoad(cfg LoadConfig) {
	ui.PrintSection("Loading Model")

	// Check SkyReels exists
	if _, err := os.Stat(SkyReelsDir); os.IsNotExist(err) {
		ui.PrintStatus("error", "SkyReels-V2 not found at "+SkyReelsDir)
		ui.PrintSuggestion("SkyReels not installed", []string{
			"Clone: git clone git@github.com:SkyworkAI/SkyReels-V2.git",
			"Or update SkyReelsDir path in config",
		})
		return
	}

	// Build the warmup command based on config
	// For multi-GPU, use torchrun with generate_video_sequential.py
	// For single GPU, use python3 with generate_video_df.py

	var cmdName string
	var args []string

	if cfg.GPUCount > 1 {
		// Multi-GPU: use torchrun with sequential script
		script := filepath.Join(SkyReelsDir, "generate_video_sequential.py")
		configFile := filepath.Join(SkyReelsDir, "skyreels_v2_infer/sequential_pipeline/config.yaml")
		cmdName = "torchrun"
		args = []string{
			fmt.Sprintf("--nproc_per_node=%d", cfg.GPUCount),
			"--master_port=29500",
			script,
			"--prompt", "warmup test",
			"--num_frames", "17", // Minimal frames for warmup
			"--resolution", "540P",
			"--output_dir", "/tmp/sky_warmup",
			"--config", configFile,
		}
	} else {
		// Single GPU: use regular python
		script := filepath.Join(SkyReelsDir, "generate_video_df.py")
		cmdName = "python3"
		args = []string{
			script,
			"--prompt", "warmup test",
			"--num_frames", "9",
			"--resolution", "540P",
			"--inference_steps", "1",
			"--outdir", "/tmp/sky_warmup",
		}

		// Single GPU options
		if cfg.TeaCacheEnabled {
			args = append(args, "--teacache")
		}
		if cfg.OffloadEnabled {
			args = append(args, "--offload")
		}
	}

	// Add seed for reproducibility (required for USP mode)
	args = append(args, "--seed", "42")

	fmt.Printf("  %sCommand: %s %s%s\n", ui.Muted(""), ui.Muted(cmdName), ui.Muted(strings.Join(args, " ")), ui.Reset)
	fmt.Println()

	// Execute with progress tracking
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = SkyReelsDir

	// Create pipes for output
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start the command
	if err := cmd.Start(); err != nil {
		ui.PrintStatus("error", "Failed to start: "+err.Error())
		return
	}

	// Progress tracking
	done := make(chan bool)
	var loadErr error
	var stderrOutput strings.Builder

	// Start spinner and GPU monitoring in parallel
	go func() {
		frame := 0
		stages := []string{
			"Initializing...",
			"Loading model weights...",
			"Setting up CUDA...",
			"Distributing across GPUs...",
			"Warming up...",
		}
		stage := 0
		stageTime := time.Now()

		for {
			select {
			case <-done:
				ui.ClearLine()
				return
			default:
				// Update stage every 30 seconds
				if time.Since(stageTime) > 30*time.Second && stage < len(stages)-1 {
					stage++
					stageTime = time.Now()
				}

				ui.PrintSpinner(frame, stages[stage])
				frame++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Monitor GPU memory in parallel
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Quick GPU check in background
				gpus := l.detectGPUs()
				loadedCount := 0
				for _, gpu := range gpus {
					if gpu.MemoryUsed >= MinVRAMForModelLoad {
						loadedCount++
					}
				}
				// We could update a shared state here if needed
			}
		}
	}()

	// Consume stdout, capture stderr for error reporting
	go io.Copy(io.Discard, stdout)
	go io.Copy(&stderrOutput, stderr)

	// Wait for command
	loadErr = cmd.Wait()
	close(done)

	fmt.Println()

	if loadErr != nil {
		ui.PrintStatus("error", "Model loading failed: "+loadErr.Error())

		// Show captured error output
		errStr := stderrOutput.String()
		if errStr != "" {
			fmt.Println()
			ui.PrintSection("Error Output")
			// Show last 20 lines of error
			lines := strings.Split(strings.TrimSpace(errStr), "\n")
			start := 0
			if len(lines) > 20 {
				start = len(lines) - 20
				fmt.Printf("  %s\n", ui.Muted("... (truncated, showing last 20 lines)"))
			}
			for _, line := range lines[start:] {
				fmt.Printf("  %s\n", ui.Muted(line))
			}
		}

		fmt.Println()
		ui.PrintSuggestion("Loading failed", []string{
			"Check if Python environment is activated",
			"Verify model files are downloaded",
			"Run 'sky doctor' for diagnostics",
		})
		return
	}

	// Verify loading with progress
	ui.PrintSpinner(0, "Verifying GPU memory...")
	time.Sleep(500 * time.Millisecond)
	ui.ClearLine()

	gpus := l.detectGPUs()
	loadedCount := 0
	for _, gpu := range gpus {
		if gpu.MemoryUsed >= MinVRAMForModelLoad {
			loadedCount++
		}
	}

	if loadedCount > 0 {
		ui.PrintStatus("success", fmt.Sprintf("Models loaded on %d/%d GPUs", loadedCount, len(gpus)))
	} else {
		ui.PrintStatus("warning", "Model may not have loaded - check GPU memory")
	}

	fmt.Println()
	l.printGPUStatus(gpus)

	// Cleanup
	os.RemoveAll("/tmp/sky_warmup")
}

func (l *Loader) prompt(question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("  %s: ", question)
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

func (l *Loader) confirm(question string) bool {
	fmt.Printf("  %s [Y/n]: ", question)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "" || input == "y" || input == "yes"
}

// QuickLoad loads models with default/recommended settings (non-interactive)
func (l *Loader) QuickLoad() {
	ui.PrintHeader("Quick Model Load")

	gpus := l.detectGPUs()
	if len(gpus) == 0 {
		ui.PrintStatus("error", "No GPUs detected")
		return
	}

	l.GPUCount = len(gpus)
	l.printGPUStatus(gpus)

	// Check if already loaded
	loadedCount := 0
	for _, gpu := range gpus {
		if gpu.MemoryUsed >= MinVRAMForModelLoad {
			loadedCount++
		}
	}

	if loadedCount == len(gpus) {
		ui.PrintStatus("success", "Models already loaded on all GPUs")
		return
	}

	// Use recommended defaults
	cfg := LoadConfig{
		GPUCount:        l.GPUCount,
		Model:           "SkyReels-V2-DF-14B-540P",
		Precision:       "bf16",
		TeaCacheEnabled: true,
		OffloadEnabled:  false,
	}

	// Set parallelism based on GPU count
	switch l.GPUCount {
	case 1:
		cfg.Parallelism = "none"
		cfg.ParallelDegree = 1
	case 2:
		cfg.Parallelism = "cp"
		cfg.ParallelDegree = 2
	case 4:
		cfg.Parallelism = "cp"
		cfg.ParallelDegree = 4
	case 8:
		cfg.Parallelism = "hybrid"
		cfg.ParallelDegree = 8
	default:
		cfg.Parallelism = "cp"
		cfg.ParallelDegree = l.GPUCount
	}

	ui.PrintSection("Using Recommended Settings")
	ui.PrintKeyValue("Parallelism", fmt.Sprintf("%s%d", strings.ToUpper(cfg.Parallelism), cfg.ParallelDegree))
	ui.PrintKeyValue("Precision", cfg.Precision)
	ui.PrintKeyValue("TeaCache", "enabled")
	fmt.Println()

	l.executeLoad(cfg)
}
