package doctor

import (
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
	// Minimum VRAM (MB) to consider a model loaded per GPU
	// 14B model uses ~28-35GB per GPU with CP4
	MinVRAMForModelLoaded = 20000 // 20GB minimum per GPU

	// SkyReels directory
	SkyReelsDir = "/home/ubuntu/SkyReels-V2"
)

// Doctor handles system health checks
type Doctor struct {
	Config *config.Config
}

// GPUStatus represents the status of a single GPU
type GPUStatus struct {
	Index       int
	Name        string
	MemoryUsed  int
	MemoryTotal int
	Utilization int
	Temperature int
	ModelLoaded bool
}

// DiagnosticResult represents the result of a diagnostic check
type DiagnosticResult struct {
	Name    string
	Status  string // "ok", "warning", "error"
	Message string
	Details []string
}

// New creates a new Doctor instance
func New(cfg *config.Config) *Doctor {
	return &Doctor{Config: cfg}
}

// Run performs all diagnostic checks
func (d *Doctor) Run() {
	ui.PrintHeader("System Diagnostics")

	results := []DiagnosticResult{}

	// Check GPUs and model loading
	gpuResult, gpuStatuses := d.checkGPUMemory()
	results = append(results, gpuResult)

	// Check if models are loaded
	modelResult := d.checkModelsLoaded(gpuStatuses)
	results = append(results, modelResult)

	// Check Python environment
	results = append(results, d.checkPythonEnv())

	// Check SkyReels installation
	results = append(results, d.checkSkyReelsInstall())

	// Check disk space
	results = append(results, d.checkDiskSpace())

	// Print results
	ui.PrintSection("Diagnostic Results")

	okCount := 0
	warnCount := 0
	errCount := 0

	for _, r := range results {
		switch r.Status {
		case "ok":
			okCount++
			ui.PrintStatus("success", fmt.Sprintf("%s: %s", r.Name, r.Message))
		case "warning":
			warnCount++
			ui.PrintStatus("pending", fmt.Sprintf("%s: %s", r.Name, r.Message))
		case "error":
			errCount++
			ui.PrintStatus("error", fmt.Sprintf("%s: %s", r.Name, r.Message))
		}
		for _, detail := range r.Details {
			fmt.Printf("      %s %s\n", ui.Muted("└"), ui.Muted(detail))
		}
	}

	// Summary
	ui.PrintSection("Summary")
	if errCount > 0 {
		fmt.Printf("  %s %d errors, %d warnings, %d ok\n",
			ui.Error("✗"), errCount, warnCount, okCount)
	} else if warnCount > 0 {
		fmt.Printf("  %s %d warnings, %d ok\n",
			ui.Warning("!"), warnCount, okCount)
	} else {
		fmt.Printf("  %s All %d checks passed\n",
			ui.Success("✓"), okCount)
	}

	// Print GPU memory table
	if len(gpuStatuses) > 0 {
		d.printGPUTable(gpuStatuses)
	}

	// Recommendations
	d.printRecommendations(results, gpuStatuses)
}

func (d *Doctor) checkGPUMemory() (DiagnosticResult, []GPUStatus) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,memory.used,memory.total,utilization.gpu,temperature.gpu",
		"--format=csv,noheader,nounits")

	out, err := cmd.Output()
	if err != nil {
		return DiagnosticResult{
			Name:    "GPU Detection",
			Status:  "error",
			Message: "nvidia-smi failed",
			Details: []string{"Check NVIDIA drivers are installed"},
		}, nil
	}

	var statuses []GPUStatus
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) < 6 {
			continue
		}

		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		memUsed, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		memTotal, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
		util, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
		temp, _ := strconv.Atoi(strings.TrimSpace(parts[5]))

		modelLoaded := memUsed >= MinVRAMForModelLoaded

		statuses = append(statuses, GPUStatus{
			Index:       idx,
			Name:        strings.TrimSpace(parts[1]),
			MemoryUsed:  memUsed,
			MemoryTotal: memTotal,
			Utilization: util,
			Temperature: temp,
			ModelLoaded: modelLoaded,
		})
	}

	if len(statuses) == 0 {
		return DiagnosticResult{
			Name:    "GPU Detection",
			Status:  "error",
			Message: "No GPUs detected",
		}, nil
	}

	return DiagnosticResult{
		Name:    "GPU Detection",
		Status:  "ok",
		Message: fmt.Sprintf("%d GPUs detected", len(statuses)),
	}, statuses
}

func (d *Doctor) checkModelsLoaded(gpuStatuses []GPUStatus) DiagnosticResult {
	if len(gpuStatuses) == 0 {
		return DiagnosticResult{
			Name:    "Model Loading",
			Status:  "error",
			Message: "Cannot check - no GPUs",
		}
	}

	loadedCount := 0
	totalVRAM := 0
	usedVRAM := 0

	for _, gpu := range gpuStatuses {
		if gpu.ModelLoaded {
			loadedCount++
		}
		totalVRAM += gpu.MemoryTotal
		usedVRAM += gpu.MemoryUsed
	}

	usedGB := float64(usedVRAM) / 1024
	totalGB := float64(totalVRAM) / 1024

	if loadedCount == 0 {
		return DiagnosticResult{
			Name:   "Model Loading",
			Status: "error",
			Message: fmt.Sprintf("Models NOT loaded (%.1f / %.1f GB used)", usedGB, totalGB),
			Details: []string{
				"Run 'sky reload' to load models",
				fmt.Sprintf("Expected: >%d MB per GPU for 14B model", MinVRAMForModelLoaded),
			},
		}
	}

	if loadedCount < len(gpuStatuses) {
		return DiagnosticResult{
			Name:   "Model Loading",
			Status: "warning",
			Message: fmt.Sprintf("Partial: %d/%d GPUs loaded (%.1f / %.1f GB)",
				loadedCount, len(gpuStatuses), usedGB, totalGB),
			Details: []string{"Some GPUs may not have model shards"},
		}
	}

	return DiagnosticResult{
		Name:   "Model Loading",
		Status: "ok",
		Message: fmt.Sprintf("Models loaded on all %d GPUs (%.1f / %.1f GB)",
			loadedCount, usedGB, totalGB),
	}
}

func (d *Doctor) checkPythonEnv() DiagnosticResult {
	// Check Python
	out, err := exec.Command("python3", "--version").Output()
	if err != nil {
		return DiagnosticResult{
			Name:    "Python Environment",
			Status:  "error",
			Message: "Python 3 not found",
		}
	}

	version := strings.TrimSpace(strings.TrimPrefix(string(out), "Python "))

	// Check if in venv
	venv := os.Getenv("VIRTUAL_ENV")
	if venv == "" {
		return DiagnosticResult{
			Name:    "Python Environment",
			Status:  "warning",
			Message: "Python " + version + " (no venv active)",
			Details: []string{"Consider using a virtual environment"},
		}
	}

	// Check PyTorch
	cmd := exec.Command("python3", "-c", "import torch; print(torch.cuda.is_available())")
	torchOut, err := cmd.Output()
	if err != nil {
		return DiagnosticResult{
			Name:    "Python Environment",
			Status:  "warning",
			Message: "Python " + version + " (PyTorch not installed)",
			Details: []string{"venv: " + filepath.Base(venv)},
		}
	}

	cudaAvail := strings.TrimSpace(string(torchOut)) == "True"
	if !cudaAvail {
		return DiagnosticResult{
			Name:    "Python Environment",
			Status:  "warning",
			Message: "Python " + version + " (PyTorch without CUDA)",
			Details: []string{"venv: " + filepath.Base(venv)},
		}
	}

	return DiagnosticResult{
		Name:    "Python Environment",
		Status:  "ok",
		Message: "Python " + version + " with PyTorch+CUDA",
		Details: []string{"venv: " + filepath.Base(venv)},
	}
}

func (d *Doctor) checkSkyReelsInstall() DiagnosticResult {
	// Check SkyReels directory
	if _, err := os.Stat(SkyReelsDir); os.IsNotExist(err) {
		return DiagnosticResult{
			Name:    "SkyReels Installation",
			Status:  "error",
			Message: "SkyReels-V2 not found",
			Details: []string{"Expected at: " + SkyReelsDir},
		}
	}

	// Check for key files
	requiredFiles := []string{
		"generate_video_df.py",
		"generate_video.py",
	}

	missingFiles := []string{}
	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(SkyReelsDir, f)); os.IsNotExist(err) {
			missingFiles = append(missingFiles, f)
		}
	}

	if len(missingFiles) > 0 {
		return DiagnosticResult{
			Name:    "SkyReels Installation",
			Status:  "warning",
			Message: "Missing generation scripts",
			Details: missingFiles,
		}
	}

	return DiagnosticResult{
		Name:    "SkyReels Installation",
		Status:  "ok",
		Message: "SkyReels-V2 installed",
		Details: []string{SkyReelsDir},
	}
}

func (d *Doctor) checkDiskSpace() DiagnosticResult {
	// Check HuggingFace cache
	home, _ := os.UserHomeDir()
	cachePath := filepath.Join(home, ".cache", "huggingface")

	// Simple disk space check using df
	cmd := exec.Command("df", "-h", cachePath)
	out, err := cmd.Output()
	if err != nil {
		return DiagnosticResult{
			Name:    "Disk Space",
			Status:  "warning",
			Message: "Unable to check disk space",
		}
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return DiagnosticResult{
			Name:    "Disk Space",
			Status:  "warning",
			Message: "Unable to parse disk space",
		}
	}

	// Parse available space (usually 4th column)
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return DiagnosticResult{
			Name:    "Disk Space",
			Status:  "warning",
			Message: "Unable to parse disk space",
		}
	}

	avail := fields[3]
	usePct := ""
	if len(fields) >= 5 {
		usePct = fields[4]
	}

	return DiagnosticResult{
		Name:    "Disk Space",
		Status:  "ok",
		Message: fmt.Sprintf("%s available (%s used)", avail, usePct),
		Details: []string{"HuggingFace cache: " + cachePath},
	}
}

func (d *Doctor) printGPUTable(gpuStatuses []GPUStatus) {
	ui.PrintSection("GPU Memory Status")

	fmt.Printf("  %s%-6s %-12s %12s %12s %6s %s%s\n",
		ui.Bold, "GPU", "Status", "Used", "Total", "Temp", "Model", ui.Reset)
	fmt.Printf("  %s\n", ui.Muted(strings.Repeat("─", 65)))

	for _, gpu := range gpuStatuses {
		status := ui.Error("✗ Empty")
		if gpu.ModelLoaded {
			status = ui.Success("✓ Loaded")
		} else if gpu.MemoryUsed > 1000 {
			status = ui.Warning("? Partial")
		}

		usedGB := float64(gpu.MemoryUsed) / 1024
		totalGB := float64(gpu.MemoryTotal) / 1024
		pct := (gpu.MemoryUsed * 100) / gpu.MemoryTotal

		// Memory bar
		barWidth := 15
		filled := (pct * barWidth) / 100
		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filled {
				if pct > 80 {
					bar += ui.BrightGreen + "█" + ui.Reset
				} else if pct > 50 {
					bar += ui.BrightYellow + "█" + ui.Reset
				} else {
					bar += ui.BrightRed + "█" + ui.Reset
				}
			} else {
				bar += ui.Muted("░")
			}
		}

		tempColor := ui.BrightGreen
		if gpu.Temperature > 70 {
			tempColor = ui.BrightYellow
		}
		if gpu.Temperature > 85 {
			tempColor = ui.BrightRed
		}

		fmt.Printf("  %-6d %-12s %6.1f GB / %5.1f GB  %s%3d°C%s  [%s]\n",
			gpu.Index, status, usedGB, totalGB,
			tempColor, gpu.Temperature, ui.Reset, bar)
	}
}

func (d *Doctor) printRecommendations(results []DiagnosticResult, gpuStatuses []GPUStatus) {
	recommendations := []string{}

	// Check if models need loading
	modelsLoaded := true
	for _, gpu := range gpuStatuses {
		if !gpu.ModelLoaded {
			modelsLoaded = false
			break
		}
	}

	if !modelsLoaded && len(gpuStatuses) > 0 {
		recommendations = append(recommendations,
			"Run 'sky reload' to load models onto GPUs")
	}

	for _, r := range results {
		if r.Status == "error" {
			switch r.Name {
			case "SkyReels Installation":
				recommendations = append(recommendations,
					"Clone SkyReels: git clone git@github.com:SkyworkAI/SkyReels-V2.git")
			case "Python Environment":
				recommendations = append(recommendations,
					"Activate your Python environment: source venv/bin/activate")
			}
		}
	}

	if len(recommendations) > 0 {
		ui.PrintSection("Recommendations")
		for i, rec := range recommendations {
			fmt.Printf("  %s%d.%s %s\n", ui.BrightCyan, i+1, ui.Reset, rec)
		}
	} else {
		ui.PrintSection("Status")
		fmt.Printf("  %s System ready for video generation!\n", ui.Success("✓"))
		fmt.Printf("  %s Run 'reel generate quick \"prompt\"' to test\n", ui.Info("→"))
	}
}

// Reload warms up the model by running a minimal generation
func (d *Doctor) Reload() {
	ui.PrintHeader("Reloading Models")

	// Check current state
	_, gpuStatuses := d.checkGPUMemory()

	loadedCount := 0
	for _, gpu := range gpuStatuses {
		if gpu.ModelLoaded {
			loadedCount++
		}
	}

	if loadedCount == len(gpuStatuses) && loadedCount > 0 {
		ui.PrintStatus("info", "Models already loaded on all GPUs")
		d.printGPUTable(gpuStatuses)
		return
	}

	ui.PrintSection("Warming Up")
	fmt.Println("  This will run a minimal generation to load model weights...")
	fmt.Println()

	// Check SkyReels exists
	if _, err := os.Stat(SkyReelsDir); os.IsNotExist(err) {
		ui.PrintStatus("error", "SkyReels-V2 not found at "+SkyReelsDir)
		return
	}

	// Detect GPU count for appropriate warmup command
	gpuCount := len(gpuStatuses)

	var cmdName string
	var args []string

	if gpuCount > 1 {
		// Multi-GPU: use torchrun with sequential script
		script := filepath.Join(SkyReelsDir, "generate_video_sequential.py")
		configFile := filepath.Join(SkyReelsDir, "skyreels_v2_infer/sequential_pipeline/config.yaml")
		cmdName = "torchrun"
		args = []string{
			fmt.Sprintf("--nproc_per_node=%d", gpuCount),
			"--master_port=29500",
			script,
			"--prompt", "warmup test",
			"--num_frames", "17",
			"--resolution", "540P",
			"--output_dir", "/tmp/sky_warmup",
			"--config", configFile,
			"--seed", "42",
		}
	} else {
		// Single GPU: use regular python
		script := filepath.Join(SkyReelsDir, "generate_video_df.py")
		cmdName = "python3"
		args = []string{
			script,
			"--prompt", "test warmup",
			"--num_frames", "9",
			"--resolution", "540P",
			"--inference_steps", "1",
			"--outdir", "/tmp/sky_warmup",
			"--seed", "42",
		}
	}

	fmt.Printf("  %sCommand: %s %s%s\n", ui.Muted(""), ui.Muted(cmdName), ui.Muted(strings.Join(args, " ")), ui.Reset)
	fmt.Println()

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

	// Progress tracking with spinner
	done := make(chan bool)

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

	// Consume stdout, capture stderr for error reporting
	var stderrOutput strings.Builder
	go io.Copy(io.Discard, stdout)
	go io.Copy(&stderrOutput, stderr)

	// Wait for command
	err := cmd.Wait()
	close(done)

	fmt.Println()

	if err != nil {
		ui.PrintStatus("error", "Warmup failed: "+err.Error())

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
		return
	}

	// Verify loading
	ui.PrintSpinner(0, "Verifying GPU memory...")
	time.Sleep(500 * time.Millisecond)
	ui.ClearLine()

	_, newStatuses := d.checkGPUMemory()
	newLoadedCount := 0
	for _, gpu := range newStatuses {
		if gpu.ModelLoaded {
			newLoadedCount++
		}
	}

	if newLoadedCount > loadedCount {
		ui.PrintStatus("success", fmt.Sprintf("Models loaded on %d GPUs", newLoadedCount))
	} else {
		ui.PrintStatus("warning", "Model loading may have failed - check GPU memory")
	}

	fmt.Println()
	d.printGPUTable(newStatuses)

	// Cleanup warmup output
	os.RemoveAll("/tmp/sky_warmup")
}
