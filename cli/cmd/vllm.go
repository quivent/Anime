package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/hf"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

// ============================================================================
// VLLM CONFIGURATION
// ============================================================================

type VLLMConfig struct {
	Model              string  `json:"model"`
	Host               string  `json:"host"`
	Port               int     `json:"port"`
	TensorParallelSize int     `json:"tensor_parallel_size"`
	GPUMemoryUtil      float64 `json:"gpu_memory_utilization"`
	MaxModelLen        int     `json:"max_model_len"`
	QuantMethod        string  `json:"quantization"`
	DType              string  `json:"dtype"`
	SwapSpace          int     `json:"swap_space"`
	EnforceEager       bool    `json:"enforce_eager"`
	TrustRemoteCode    bool    `json:"trust_remote_code"`
	PID                int     `json:"pid,omitempty"`
	StartedAt          string  `json:"started_at,omitempty"`
}

// Default models are now managed centrally in model_registry.go
// Lazy-loaded on first access to avoid startup cost for non-vllm commands
var (
	vllmModelShortcutsOnce  sync.Once
	vllmModelShortcutsCache map[string]string
)

func getVLLMModelShortcuts() map[string]string {
	vllmModelShortcutsOnce.Do(func() {
		vllmModelShortcutsCache = GetModelShortcuts()
	})
	return vllmModelShortcutsCache
}

// ============================================================================
// COMMAND DEFINITIONS
// ============================================================================

var (
	// Global flags
	vllmServer string
	vllmLocal  bool

	// Start flags
	vllmModel          string
	vllmPort           int
	vllmTPSize         int
	vllmGPUMemUtil     float64
	vllmMaxModelLen    int
	vllmQuantization   string
	vllmDType          string
	vllmSwapSpace      int
	vllmEnforceEager   bool
	vllmTrustRemote    bool
	vllmBackground     bool
	vllmPreloadToRAM   bool
	vllmEnableLoRA     bool
	vllmMaxLoRARank    int

	// Status flags
	vllmStatusJSON bool
)

var vllmCmd = &cobra.Command{
	Use:   "vllm",
	Short: "Manage vLLM inference server",
	Long: `Comprehensive vLLM server management.

Commands:
  start      Start vLLM server with a model
  stop       Stop running vLLM server
  restart    Restart vLLM with same or different model
  status     Check server status and loaded models
  load       Load or switch to a different model
  models     List available models for vLLM
  logs       View vLLM server logs

Examples:
  anime vllm start llama-70b              # Start with Llama 3.3 70B
  anime vllm start --model qwen-72b       # Start with Qwen 72B
  anime vllm status                       # Check what's running
  anime vllm load mistral-7b              # Switch to different model
  anime vllm stop                         # Stop the server`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var vllmStartCmd = &cobra.Command{
	Use:   "start [model]",
	Short: "Start vLLM server with a model",
	Long: `Start vLLM inference server with specified model.

Model shortcuts (FP16 - full precision):
  llama-70b    → meta-llama/Llama-3.3-70B-Instruct
  qwen-72b     → Qwen/Qwen2.5-72B-Instruct
  deepseek-r1  → deepseek-ai/DeepSeek-R1-Distill-Llama-70B
  mistral-7b   → mistralai/Mistral-7B-Instruct-v0.3
  mixtral      → mistralai/Mixtral-8x7B-Instruct-v0.1

Quantized models (4-bit AWQ - faster, less VRAM):
  llama-70b-awq    → hugging-quants/Meta-Llama-3.1-70B-Instruct-AWQ-INT4
  qwen-72b-awq     → Qwen/Qwen2.5-72B-Instruct-AWQ
  deepseek-r1-awq  → cognitivecomputations/DeepSeek-R1-Distill-Llama-70B-AWQ

Or use full HuggingFace model ID directly.

Memory Management:
  --gpu-mem       GPU memory utilization (0.0-1.0, default 0.90)
  --preload-ram   Keep model weights in unified RAM (GH200)
  --swap          Swap space in GB for KV cache offloading
  --dtype         Data type: auto, float16, bfloat16, float32
  --quant         Quantization: awq, gptq, fp8 (auto-detected from model name)

Examples:
  anime vllm start llama-70b              # FP16, auto-detect GPUs
  anime vllm start llama-70b-awq          # AWQ 4-bit (half the VRAM)
  anime vllm start llama-70b --quant fp8  # FP8 quantization
  anime vllm start qwen-72b --tp 4        # 4-GPU tensor parallel
  anime vllm start llama-8b -b            # Run in background`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVLLMStart,
}

var vllmStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop vLLM server",
	Long: `Stop the running vLLM server.

This will gracefully terminate the vLLM process and free GPU memory.

Examples:
  anime vllm stop              # Stop local vLLM
  anime vllm stop -s lambda    # Stop vLLM on Lambda server`,
	RunE: runVLLMStop,
}

var vllmRestartCmd = &cobra.Command{
	Use:   "restart [model]",
	Short: "Restart vLLM with same or different model",
	Long: `Restart vLLM server, optionally with a different model.

If no model is specified, restarts with the same configuration.
If a model is specified, stops and starts with the new model.

Examples:
  anime vllm restart                  # Restart with same model
  anime vllm restart llama-8b         # Restart with different model
  anime vllm restart --tp 8           # Restart with new config`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVLLMRestart,
}

var vllmStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check vLLM server status and loaded models",
	Long: `Display comprehensive vLLM server status.

Shows:
  - Server running state
  - Loaded model information
  - GPU memory usage
  - Request statistics
  - Configuration details

Examples:
  anime vllm status           # Check status
  anime vllm status --json    # Output as JSON`,
	RunE: runVLLMStatus,
}

var vllmLoadCmd = &cobra.Command{
	Use:   "load [model]",
	Short: "Load or switch to a different model",
	Long: `Load a new model into vLLM.

Note: vLLM currently requires restart to change models.
This command will stop the server and restart with the new model.

Examples:
  anime vllm load llama-8b           # Switch to Llama 8B
  anime vllm load qwen-72b --tp 4    # Load Qwen with 4 GPUs`,
	Args: cobra.ExactArgs(1),
	RunE: runVLLMLoad,
}

var vllmModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models for vLLM",
	Long: `List models that can be loaded into vLLM.

Shows:
  - Model shortcuts and their full IDs
  - Currently loaded model (if any)
  - Recommended GPU configurations

Examples:
  anime vllm models`,
	RunE: runVLLMModels,
}

var vllmLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View vLLM server logs",
	Long: `View vLLM server output and logs.

Examples:
  anime vllm logs            # Show recent logs
  anime vllm logs -f         # Follow logs in real-time`,
	RunE: runVLLMLogs,
}

var vllmSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive TUI wizard for vLLM setup",
	Long: `Launch an interactive TUI wizard for vLLM setup.

The wizard guides you through:
  - vLLM installation (if not installed)
  - Model selection from curated list or HuggingFace
  - GPU configuration (tensor parallelism, memory)
  - Advanced options (quantization, LoRA, etc.)

Examples:
  anime vllm setup           # Launch setup wizard`,
	RunE: runVLLMSetup,
}

func init() {
	rootCmd.AddCommand(vllmCmd)

	// Global flags
	vllmCmd.PersistentFlags().StringVarP(&vllmServer, "server", "s", "", "Remote server to use")
	vllmCmd.PersistentFlags().BoolVarP(&vllmLocal, "local", "l", false, "Force local execution")

	// Start command flags
	vllmStartCmd.Flags().StringVarP(&vllmModel, "model", "m", "", "Model to load (shortcut or HuggingFace ID)")
	vllmStartCmd.Flags().IntVarP(&vllmPort, "port", "p", 8000, "Server port")
	vllmStartCmd.Flags().IntVar(&vllmTPSize, "tp", 0, "Tensor parallel size (0 = auto-detect GPUs)")
	vllmStartCmd.Flags().Float64Var(&vllmGPUMemUtil, "gpu-mem", 0.90, "GPU memory utilization (0.0-1.0)")
	vllmStartCmd.Flags().IntVar(&vllmMaxModelLen, "max-len", 0, "Maximum model context length (0 = model default)")
	vllmStartCmd.Flags().StringVarP(&vllmQuantization, "quant", "q", "", "Quantization method: awq, gptq, squeezellm, fp8")
	vllmStartCmd.Flags().StringVar(&vllmDType, "dtype", "auto", "Data type: auto, float16, bfloat16, float32")
	vllmStartCmd.Flags().IntVar(&vllmSwapSpace, "swap", 0, "CPU swap space in GB for KV cache")
	vllmStartCmd.Flags().BoolVar(&vllmEnforceEager, "eager", false, "Disable CUDA graphs for debugging")
	vllmStartCmd.Flags().BoolVar(&vllmTrustRemote, "trust-remote", true, "Trust remote code from HuggingFace")
	vllmStartCmd.Flags().BoolVarP(&vllmBackground, "background", "b", false, "Run in background")
	vllmStartCmd.Flags().BoolVar(&vllmPreloadToRAM, "preload-ram", false, "Preload model to unified RAM (Apple/GH200)")
	vllmStartCmd.Flags().BoolVar(&vllmEnableLoRA, "enable-lora", false, "Enable LoRA adapter support")
	vllmStartCmd.Flags().IntVar(&vllmMaxLoRARank, "max-lora-rank", 64, "Maximum LoRA rank")

	// Status command flags
	vllmStatusCmd.Flags().BoolVar(&vllmStatusJSON, "json", false, "Output as JSON")

	// Restart inherits start flags
	vllmRestartCmd.Flags().StringVarP(&vllmModel, "model", "m", "", "Model to load")
	vllmRestartCmd.Flags().IntVar(&vllmTPSize, "tp", 0, "Tensor parallel size")
	vllmRestartCmd.Flags().Float64Var(&vllmGPUMemUtil, "gpu-mem", 0.90, "GPU memory utilization")

	// Load inherits some start flags
	vllmLoadCmd.Flags().IntVar(&vllmTPSize, "tp", 0, "Tensor parallel size")
	vllmLoadCmd.Flags().Float64Var(&vllmGPUMemUtil, "gpu-mem", 0.90, "GPU memory utilization")

	// Add subcommands
	vllmCmd.AddCommand(vllmStartCmd)
	vllmCmd.AddCommand(vllmStopCmd)
	vllmCmd.AddCommand(vllmRestartCmd)
	vllmCmd.AddCommand(vllmStatusCmd)
	vllmCmd.AddCommand(vllmLoadCmd)
	vllmCmd.AddCommand(vllmModelsCmd)
	vllmCmd.AddCommand(vllmLogsCmd)
	vllmCmd.AddCommand(vllmSetupCmd)
}

// ============================================================================
// START COMMAND
// ============================================================================

func runVLLMStart(cmd *cobra.Command, args []string) error {
	// Determine model
	var model string
	if len(args) > 0 {
		model = args[0]
	} else if vllmModel != "" {
		model = vllmModel
	} else {
		// Interactive model selection
		return showVLLMModelSelection()
	}

	// Resolve model shortcut
	if fullID, ok := getVLLMModelShortcuts()[model]; ok {
		model = fullID
	}

	// Auto-detect quantization from model name if not explicitly set
	if vllmQuantization == "" {
		vllmQuantization = detectQuantizationMethod(model)
	}

	// Check if we should run remotely
	if vllmServer != "" && !vllmLocal {
		return runVLLMStartRemote(model)
	}

	return runVLLMStartLocal(model)
}

// detectQuantizationMethod auto-detects quantization from model name
func detectQuantizationMethod(model string) string {
	modelLower := strings.ToLower(model)
	switch {
	case strings.Contains(modelLower, "-awq") || strings.Contains(modelLower, "_awq"):
		return "awq"
	case strings.Contains(modelLower, "-gptq") || strings.Contains(modelLower, "_gptq"):
		return "gptq"
	case strings.Contains(modelLower, "-fp8") || strings.Contains(modelLower, "_fp8"):
		return "fp8"
	case strings.Contains(modelLower, "-squeezellm"):
		return "squeezellm"
	default:
		return ""
	}
}

func runVLLMStartLocal(model string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("VLLM SERVER"))
	fmt.Println()

	// ========================================================================
	// PHASE 1: Pre-flight Checks
	// ========================================================================
	fmt.Println(theme.InfoStyle.Render("[1/5] Pre-flight checks"))

	// Check if vLLM is already running
	if isVLLMRunning() {
		fmt.Printf("  %s vLLM is already running\n", theme.WarningStyle.Render("⚠"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Options:"))
		fmt.Println(theme.DimTextStyle.Render("  anime vllm stop       # Stop current server"))
		fmt.Println(theme.DimTextStyle.Render("  anime vllm restart    # Restart with new model"))
		fmt.Println(theme.DimTextStyle.Render("  anime vllm status     # Check current status"))
		fmt.Println()
		return nil
	}
	fmt.Printf("  %s No existing server running\n", theme.SuccessStyle.Render("✓"))

	// Check GPU availability
	gpuCount := vllmTPSize
	if gpuCount == 0 {
		gpuCount = detectVLLMGPUCount()
	}
	if gpuCount == 0 {
		fmt.Printf("  %s No GPU detected\n", theme.ErrorStyle.Render("✗"))
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("ERROR: No GPU found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Possible causes:"))
		fmt.Println(theme.DimTextStyle.Render("  • NVIDIA drivers not installed"))
		fmt.Println(theme.DimTextStyle.Render("  • CUDA not properly configured"))
		fmt.Println(theme.DimTextStyle.Render("  • GPU not visible to the system"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Try:"))
		fmt.Println(theme.HighlightStyle.Render("  nvidia-smi           # Check if GPU is visible"))
		fmt.Println(theme.HighlightStyle.Render("  anime install nvidia # Install NVIDIA drivers"))
		fmt.Println()
		return fmt.Errorf("no GPU detected")
	}
	fmt.Printf("  %s Found %d GPU(s)\n", theme.SuccessStyle.Render("✓"), gpuCount)

	// Check disk space (rough estimate: 140GB for 70B model)
	modelSize := estimateModelSize(model)
	freeSpace := getFreeDiskSpaceGB()
	if freeSpace > 0 && freeSpace < modelSize {
		fmt.Printf("  %s Low disk space: %.0fGB free, need ~%.0fGB\n", theme.WarningStyle.Render("⚠"), freeSpace, modelSize)
	} else if freeSpace > 0 {
		fmt.Printf("  %s Disk space OK (%.0fGB free)\n", theme.SuccessStyle.Render("✓"), freeSpace)
	}

	// Check HuggingFace auth for gated models
	if isGatedModel(model) {
		if hf.GetToken() == "" {
			fmt.Printf("  %s HuggingFace token required for gated model\n", theme.ErrorStyle.Render("✗"))
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("ERROR: This model requires HuggingFace authentication"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Fix:"))
			fmt.Println(theme.HighlightStyle.Render("  1. Get token from https://huggingface.co/settings/tokens"))
			fmt.Println(theme.HighlightStyle.Render("  2. Accept model license at https://huggingface.co/" + model))
			fmt.Println(theme.HighlightStyle.Render("  3. Run: export HF_TOKEN=your_token"))
			fmt.Println(theme.HighlightStyle.Render("     Or:  echo 'your_token' > ~/.cache/huggingface/token"))
			fmt.Println()
			return fmt.Errorf("HuggingFace authentication required")
		}
		fmt.Printf("  %s HuggingFace token configured\n", theme.SuccessStyle.Render("✓"))
	}
	fmt.Println()

	// ========================================================================
	// PHASE 2: Dependencies
	// ========================================================================
	fmt.Println(theme.InfoStyle.Render("[2/5] Checking dependencies"))

	if !isVLLMInstalled() {
		fmt.Printf("  %s vLLM not installed\n", theme.WarningStyle.Render("⚠"))
		fmt.Println()

		// Prompt for auto-install
		fmt.Println(theme.InfoStyle.Render("vLLM is required but not installed."))
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("Install vLLM now? [Y/n]: "))

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "" || response == "y" || response == "yes" {
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Installing vLLM (this may take a few minutes)..."))
			fmt.Println()

			// Run the installer
			if err := runVLLMInstall(); err != nil {
				fmt.Printf("  %s Installation failed: %v\n", theme.ErrorStyle.Render("✗"), err)
				fmt.Println()
				fmt.Println(theme.InfoStyle.Render("Try manual installation:"))
				fmt.Println(theme.HighlightStyle.Render("  anime install vllm"))
				fmt.Println(theme.HighlightStyle.Render("  anime vllm doctor --fix"))
				fmt.Println()
				return fmt.Errorf("vLLM installation failed")
			}
			fmt.Printf("  %s vLLM installed successfully\n", theme.SuccessStyle.Render("✓"))
		} else {
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Install manually with:"))
			fmt.Println(theme.HighlightStyle.Render("  anime install vllm"))
			fmt.Println()
			return fmt.Errorf("vLLM not installed")
		}
	} else {
		fmt.Printf("  %s vLLM installed\n", theme.SuccessStyle.Render("✓"))
	}

	// Quick health check - run doctor if issues detected
	if !quickVLLMHealthCheck() {
		fmt.Printf("  %s Dependency issues detected, running auto-fix...\n", theme.WarningStyle.Render("⚠"))
		if err := runVLLMDoctorFix(); err != nil {
			fmt.Printf("  %s Auto-fix failed: %v\n", theme.ErrorStyle.Render("✗"), err)
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Run manual diagnostics:"))
			fmt.Println(theme.HighlightStyle.Render("  anime vllm doctor --fix"))
			fmt.Println()
			return fmt.Errorf("dependency issues could not be resolved")
		}
		fmt.Printf("  %s Dependencies fixed\n", theme.SuccessStyle.Render("✓"))
	} else {
		fmt.Printf("  %s Dependencies OK\n", theme.SuccessStyle.Render("✓"))
	}
	fmt.Println()

	// ========================================================================
	// PHASE 3: Configuration
	// ========================================================================
	fmt.Println(theme.InfoStyle.Render("[3/5] Configuring server"))

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(model))
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Port:"), vllmPort)
	fmt.Printf("  %s %d GPU(s)\n", theme.DimTextStyle.Render("Tensor Parallel:"), gpuCount)
	fmt.Printf("  %s %.0f%%\n", theme.DimTextStyle.Render("GPU Memory:"), vllmGPUMemUtil*100)
	if vllmQuantization != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Quantization:"), theme.HighlightStyle.Render(vllmQuantization))
	}
	if vllmPreloadToRAM {
		fmt.Printf("  %s Enabled\n", theme.DimTextStyle.Render("Unified RAM:"))
	}
	fmt.Println()

	// ========================================================================
	// PHASE 4: Model Download (if needed)
	// ========================================================================
	fmt.Println(theme.InfoStyle.Render("[4/5] Preparing model"))

	if isModelCached(model) {
		fmt.Printf("  %s Model already cached\n", theme.SuccessStyle.Render("✓"))
	} else {
		fmt.Printf("  %s Model will be downloaded on first start\n", theme.InfoStyle.Render("ℹ"))
		if modelSize > 0 {
			fmt.Printf("  %s Estimated size: ~%.0fGB\n", theme.DimTextStyle.Render("  "), modelSize)
		}
	}
	fmt.Println()

	// ========================================================================
	// PHASE 5: Start Server
	// ========================================================================
	fmt.Println(theme.InfoStyle.Render("[5/5] Starting server"))

	// Build vLLM command
	vllmArgs := buildVLLMArgs(model, gpuCount)

	if vllmBackground {
		return startVLLMBackground(model, vllmArgs)
	}

	return startVLLMForeground(model, vllmArgs)
}

// estimateModelSize returns estimated size in GB for a model
func estimateModelSize(model string) float64 {
	model = strings.ToLower(model)
	switch {
	case strings.Contains(model, "70b"):
		return 140
	case strings.Contains(model, "72b"):
		return 145
	case strings.Contains(model, "32b") || strings.Contains(model, "33b") || strings.Contains(model, "34b"):
		return 70
	case strings.Contains(model, "13b") || strings.Contains(model, "14b"):
		return 28
	case strings.Contains(model, "7b") || strings.Contains(model, "8b"):
		return 16
	default:
		return 50 // Default estimate
	}
}

// getFreeDiskSpaceGB returns free disk space in GB, or 0 if unknown
func getFreeDiskSpaceGB() float64 {
	cmd := exec.Command("df", "-BG", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0
	}
	// Field 3 is available space, e.g., "500G"
	avail := strings.TrimSuffix(fields[3], "G")
	if val, err := strconv.ParseFloat(avail, 64); err == nil {
		return val
	}
	return 0
}

// isGatedModel returns true if the model requires HuggingFace authentication
func isGatedModel(model string) bool {
	gatedPrefixes := []string{
		"meta-llama/",
		"mistralai/Mistral",
		"google/gemma",
	}
	model = strings.ToLower(model)
	for _, prefix := range gatedPrefixes {
		if strings.Contains(model, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// isModelCached checks if model weights are already downloaded
func isModelCached(model string) bool {
	home, _ := os.UserHomeDir()
	// Check common cache locations
	cachePaths := []string{
		filepath.Join(home, ".cache", "huggingface", "hub"),
		filepath.Join(home, ".cache", "huggingface", "transformers"),
	}

	// Normalize model name for cache lookup
	modelDir := strings.ReplaceAll(model, "/", "--")

	for _, cachePath := range cachePaths {
		entries, err := os.ReadDir(cachePath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if strings.Contains(entry.Name(), modelDir) {
				return true
			}
		}
	}
	return false
}

// runVLLMInstall runs the vLLM installation
func runVLLMInstall() error {
	cmd := exec.Command("anime", "install", "vllm", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// quickVLLMHealthCheck does a fast check if vLLM is likely to work
func quickVLLMHealthCheck() bool {
	// Check if we can import vllm in Python
	cmd := exec.Command("python3", "-c", "import vllm; import torch; print(torch.cuda.is_available())")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "True"
}

// runVLLMDoctorFix runs the vLLM doctor with --fix flag
func runVLLMDoctorFix() error {
	cmd := exec.Command("anime", "vllm", "doctor", "--fix")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildVLLMArgs(model string, gpuCount int) []string {
	args := []string{
		"serve", model,
		"--host", "0.0.0.0",
		"--port", strconv.Itoa(vllmPort),
		"--tensor-parallel-size", strconv.Itoa(gpuCount),
		"--gpu-memory-utilization", fmt.Sprintf("%.2f", vllmGPUMemUtil),
	}

	if vllmMaxModelLen > 0 {
		args = append(args, "--max-model-len", strconv.Itoa(vllmMaxModelLen))
	}

	if vllmQuantization != "" {
		args = append(args, "--quantization", vllmQuantization)
	}

	if vllmDType != "auto" {
		args = append(args, "--dtype", vllmDType)
	}

	if vllmSwapSpace > 0 {
		args = append(args, "--swap-space", strconv.Itoa(vllmSwapSpace))
	}

	if vllmEnforceEager {
		args = append(args, "--enforce-eager")
	}

	if vllmTrustRemote {
		args = append(args, "--trust-remote-code")
	}

	if vllmEnableLoRA {
		args = append(args, "--enable-lora")
		args = append(args, "--max-lora-rank", strconv.Itoa(vllmMaxLoRARank))
	}

	// Unified RAM preload settings (for Apple Silicon / GH200)
	if vllmPreloadToRAM {
		// These env vars will be set before running
		args = append(args, "--disable-log-requests")
	}

	return args
}

func startVLLMForeground(model string, args []string) error {
	cmd := exec.Command("vllm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment with HuggingFace token for model downloads
	cmd.Env = append(os.Environ(), "HF_TOKEN="+hf.GetToken())

	// Add unified RAM settings if requested
	if vllmPreloadToRAM {
		cmd.Env = append(cmd.Env,
			"VLLM_CPU_KVCACHE_SPACE=4",
			"PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True",
		)
	}

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Stopping vLLM..."))
		cmd.Process.Signal(syscall.SIGTERM)
	}()

	fmt.Println(theme.SuccessStyle.Render("vLLM server starting..."))
	fmt.Println(theme.DimTextStyle.Render("Press Ctrl+C to stop"))
	fmt.Println()

	return cmd.Run()
}

func startVLLMBackground(model string, args []string) error {
	// Create log file
	logPath := filepath.Join(os.TempDir(), "vllm.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd := exec.Command("vllm", args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Set environment with HuggingFace token for model downloads
	cmd.Env = append(os.Environ(), "HF_TOKEN="+hf.GetToken())

	// Add unified RAM settings if requested
	if vllmPreloadToRAM {
		cmd.Env = append(cmd.Env,
			"VLLM_CPU_KVCACHE_SPACE=4",
			"PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True",
		)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start vLLM: %w", err)
	}

	// Save config for status/restart
	config := VLLMConfig{
		Model:              model,
		Host:               "0.0.0.0",
		Port:               vllmPort,
		TensorParallelSize: vllmTPSize,
		GPUMemoryUtil:      vllmGPUMemUtil,
		MaxModelLen:        vllmMaxModelLen,
		QuantMethod:        vllmQuantization,
		DType:              vllmDType,
		SwapSpace:          vllmSwapSpace,
		EnforceEager:       vllmEnforceEager,
		TrustRemoteCode:    vllmTrustRemote,
		PID:                cmd.Process.Pid,
		StartedAt:          time.Now().Format(time.RFC3339),
	}
	saveVLLMConfig(config)

	// Save PID
	pidPath := filepath.Join(os.TempDir(), "vllm.pid")
	os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)

	fmt.Println(theme.SuccessStyle.Render("vLLM server started in background"))
	fmt.Println()
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("PID:"), cmd.Process.Pid)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Logs:"), logPath)
	fmt.Printf("  %s http://localhost:%d\n", theme.DimTextStyle.Render("URL:"), vllmPort)
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Commands:"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm status    # Check status"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm logs      # View logs"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm stop      # Stop server"))
	fmt.Println()

	// Wait for server to be ready
	fmt.Println(theme.InfoStyle.Render("Waiting for server to be ready..."))
	for i := 0; i < 60; i++ {
		if checkVLLMHealth(vllmPort) {
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("Server is ready!"))
			fmt.Println()
			return nil
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}

	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("Server is still starting. Check logs for progress."))
	fmt.Println()

	return nil
}

func runVLLMStartRemote(model string) error {
	target, host, user, keyPath, err := getServerTarget(vllmServer)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("VLLM SERVER (REMOTE)"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(target))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(model))
	fmt.Println()

	client, err := ssh.NewClient(host, user, keyPath)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Check if vLLM is already running
	output, _ := client.RunCommand("pgrep -f 'vllm serve'")
	if strings.TrimSpace(output) != "" {
		fmt.Println(theme.WarningStyle.Render("vLLM is already running on remote server!"))
		fmt.Println()
		return nil
	}

	// Detect GPU count on remote
	gpuOutput, _ := client.RunCommand("nvidia-smi --list-gpus | wc -l")
	gpuCount := 1
	if n, err := strconv.Atoi(strings.TrimSpace(gpuOutput)); err == nil && n > 0 {
		gpuCount = n
	}
	if vllmTPSize > 0 {
		gpuCount = vllmTPSize
	}

	// Build start command with HuggingFace token
	startCmd := fmt.Sprintf(
		"HF_TOKEN=%s nohup vllm serve %s --host 0.0.0.0 --port %d --tensor-parallel-size %d --gpu-memory-utilization %.2f --trust-remote-code > ~/vllm.log 2>&1 & echo $! > ~/vllm.pid",
		hf.GetToken(), model, vllmPort, gpuCount, vllmGPUMemUtil,
	)

	fmt.Println(theme.InfoStyle.Render("Starting vLLM on remote server..."))

	_, err = client.RunCommand(startCmd)
	if err != nil {
		return fmt.Errorf("failed to start vLLM: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("vLLM started on remote server"))
	fmt.Println()

	// Setup port forwarding
	return setupPortForwarding(host, strconv.Itoa(vllmPort), "vLLM", fmt.Sprintf("http://localhost:%d", vllmPort))
}

// ============================================================================
// STOP COMMAND
// ============================================================================

func runVLLMStop(cmd *cobra.Command, args []string) error {
	if vllmServer != "" && !vllmLocal {
		return runVLLMStopRemote()
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("STOPPING VLLM"))
	fmt.Println()

	// Find vLLM process
	pidPath := filepath.Join(os.TempDir(), "vllm.pid")
	pidBytes, err := os.ReadFile(pidPath)
	if err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if pid > 0 {
			fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Stopping PID:"), pid)
			exec.Command("kill", "-TERM", strconv.Itoa(pid)).Run()
			time.Sleep(2 * time.Second)
			exec.Command("kill", "-KILL", strconv.Itoa(pid)).Run()
			os.Remove(pidPath)
		}
	}

	// Also kill any vllm processes
	exec.Command("pkill", "-f", "vllm serve").Run()

	// Clean up config
	configPath := filepath.Join(os.TempDir(), "vllm.config.json")
	os.Remove(configPath)

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("vLLM server stopped"))
	fmt.Println()

	return nil
}

func runVLLMStopRemote() error {
	target, host, user, keyPath, err := getServerTarget(vllmServer)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("STOPPING VLLM (REMOTE)"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(target))
	fmt.Println()

	client, err := ssh.NewClient(host, user, keyPath)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Kill vLLM process
	client.RunCommand("pkill -f 'vllm serve'")
	client.RunCommand("rm -f ~/vllm.pid")

	fmt.Println(theme.SuccessStyle.Render("vLLM stopped on remote server"))
	fmt.Println()

	return nil
}

// ============================================================================
// RESTART COMMAND
// ============================================================================

func runVLLMRestart(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("RESTARTING VLLM"))
	fmt.Println()

	// Load existing config
	config := loadVLLMConfig()

	// Stop current server
	runVLLMStop(cmd, nil)

	// Wait for cleanup
	time.Sleep(2 * time.Second)

	// Determine model for restart
	var model string
	if len(args) > 0 {
		model = args[0]
	} else if vllmModel != "" {
		model = vllmModel
	} else if config.Model != "" {
		model = config.Model
	} else {
		return fmt.Errorf("no model specified and no previous configuration found")
	}

	// Restore config values if not overridden
	if vllmTPSize == 0 && config.TensorParallelSize > 0 {
		vllmTPSize = config.TensorParallelSize
	}

	// Start with new/same model
	vllmBackground = true // Always background on restart
	return runVLLMStart(cmd, []string{model})
}

// ============================================================================
// STATUS COMMAND
// ============================================================================

func runVLLMStatus(cmd *cobra.Command, args []string) error {
	if vllmServer != "" && !vllmLocal {
		return runVLLMStatusRemote()
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("VLLM STATUS"))
	fmt.Println()

	// Check if running
	running := isVLLMRunning()

	if !running {
		fmt.Println(theme.WarningStyle.Render("vLLM is not running"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Start with:"))
		fmt.Println(theme.HighlightStyle.Render("  anime vllm start llama-70b"))
		fmt.Println()
		return nil
	}

	// Get loaded model info from API
	modelsResp, err := getVLLMModels(vllmPort)
	if err != nil {
		fmt.Println(theme.WarningStyle.Render("Server running but not responding yet..."))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render("Server Status: Running"))
	fmt.Println()

	// Display loaded models
	fmt.Println(theme.InfoStyle.Render("Loaded Models:"))
	for _, model := range modelsResp {
		fmt.Printf("  %s %s\n", theme.SymbolSparkle, theme.HighlightStyle.Render(model))
	}
	fmt.Println()

	// Load saved config for details
	config := loadVLLMConfig()
	if config.Model != "" {
		fmt.Println(theme.InfoStyle.Render("Configuration:"))
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Port:"), config.Port)
		fmt.Printf("  %s %d GPU(s)\n", theme.DimTextStyle.Render("Tensor Parallel:"), config.TensorParallelSize)
		fmt.Printf("  %s %.0f%%\n", theme.DimTextStyle.Render("GPU Memory:"), config.GPUMemoryUtil*100)
		if config.StartedAt != "" {
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Started:"), config.StartedAt)
		}
		fmt.Println()
	}

	// Get GPU memory usage
	fmt.Println(theme.InfoStyle.Render("GPU Memory Usage:"))
	gpuInfo := getGPUMemoryUsage()
	for _, info := range gpuInfo {
		fmt.Printf("  %s\n", info)
	}
	fmt.Println()

	// Show endpoint
	fmt.Println(theme.InfoStyle.Render("API Endpoint:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("http://localhost:%d/v1/chat/completions", vllmPort)))
	fmt.Println()

	return nil
}

func runVLLMStatusRemote() error {
	target, host, user, keyPath, err := getServerTarget(vllmServer)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("VLLM STATUS (REMOTE)"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(target))
	fmt.Println()

	client, err := ssh.NewClient(host, user, keyPath)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Check if running
	output, _ := client.RunCommand("pgrep -f 'vllm serve'")
	if strings.TrimSpace(output) == "" {
		fmt.Println(theme.WarningStyle.Render("vLLM is not running on remote server"))
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render("Server Status: Running"))
	fmt.Println()

	// Get GPU info
	gpuOutput, _ := client.RunCommand("nvidia-smi --query-gpu=name,memory.used,memory.total --format=csv,noheader")
	if gpuOutput != "" {
		fmt.Println(theme.InfoStyle.Render("GPU Memory Usage:"))
		for _, line := range strings.Split(strings.TrimSpace(gpuOutput), "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	return nil
}

// ============================================================================
// LOAD COMMAND
// ============================================================================

func runVLLMLoad(cmd *cobra.Command, args []string) error {
	model := args[0]

	// Resolve shortcut
	if fullID, ok := getVLLMModelShortcuts()[model]; ok {
		model = fullID
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("LOADING MODEL"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(model))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("Note: vLLM requires restart to change models"))
	fmt.Println(theme.DimTextStyle.Render("This will stop the current server and start with the new model"))
	fmt.Println()

	// Use restart with the new model
	vllmBackground = true
	return runVLLMRestart(cmd, []string{model})
}

// ============================================================================
// MODELS COMMAND
// ============================================================================

func runVLLMModels(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("VLLM MODELS"))
	fmt.Println()

	// Check current status
	var currentModel string
	if isVLLMRunning() {
		models, _ := getVLLMModels(vllmPort)
		if len(models) > 0 {
			currentModel = models[0]
		}
	}

	// Group models by category
	categories := map[string][]struct {
		shortcut string
		fullID   string
	}{
		"Llama": {
			{"llama-70b", getVLLMModelShortcuts()["llama-70b"]},
			{"llama-8b", getVLLMModelShortcuts()["llama-8b"]},
		},
		"Qwen": {
			{"qwen-72b", getVLLMModelShortcuts()["qwen-72b"]},
			{"qwen-32b", getVLLMModelShortcuts()["qwen-32b"]},
			{"qwen-14b", getVLLMModelShortcuts()["qwen-14b"]},
			{"qwen-7b", getVLLMModelShortcuts()["qwen-7b"]},
		},
		"Mistral": {
			{"mistral-7b", getVLLMModelShortcuts()["mistral-7b"]},
			{"mixtral", getVLLMModelShortcuts()["mixtral"]},
		},
		"DeepSeek": {
			{"deepseek-r1", getVLLMModelShortcuts()["deepseek-r1"]},
			{"deepseek-67b", getVLLMModelShortcuts()["deepseek-67b"]},
		},
		"Other": {
			{"codellama", getVLLMModelShortcuts()["codellama"]},
			{"phi-4", getVLLMModelShortcuts()["phi-4"]},
		},
	}

	for category, models := range categories {
		fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("  %s:", category)))
		for _, m := range models {
			status := ""
			if m.fullID == currentModel {
				status = theme.SuccessStyle.Render(" (loaded)")
			}
			fmt.Printf("    %s %-15s → %s%s\n",
				theme.SymbolSparkle,
				theme.InfoStyle.Render(m.shortcut),
				theme.DimTextStyle.Render(m.fullID),
				status)
		}
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("Usage:"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm start llama-70b"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm start meta-llama/Llama-3.3-70B-Instruct"))
	fmt.Println()

	return nil
}

// ============================================================================
// LOGS COMMAND
// ============================================================================

func runVLLMLogs(cmd *cobra.Command, args []string) error {
	if vllmServer != "" && !vllmLocal {
		return runVLLMLogsRemote()
	}

	logPath := filepath.Join(os.TempDir(), "vllm.log")

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println(theme.WarningStyle.Render("No vLLM logs found"))
		fmt.Println(theme.DimTextStyle.Render("Start vLLM in background to generate logs:"))
		fmt.Println(theme.HighlightStyle.Render("  anime vllm start llama-70b -b"))
		return nil
	}

	// Tail the log file
	tailCmd := exec.Command("tail", "-f", "-n", "100", logPath)
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Showing logs from %s", logPath)))
	fmt.Println(theme.DimTextStyle.Render("Press Ctrl+C to stop"))
	fmt.Println()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		tailCmd.Process.Kill()
	}()

	return tailCmd.Run()
}

func runVLLMLogsRemote() error {
	target, host, user, keyPath, err := getServerTarget(vllmServer)
	if err != nil {
		return err
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Showing logs from %s", target)))
	fmt.Println(theme.DimTextStyle.Render("Press Ctrl+C to stop"))
	fmt.Println()

	// SSH tail command
	sshCmd := exec.Command("ssh")
	if keyPath != "" {
		sshCmd.Args = append(sshCmd.Args, "-i", keyPath)
	}
	sshCmd.Args = append(sshCmd.Args, fmt.Sprintf("%s@%s", user, host), "tail", "-f", "~/vllm.log")

	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

// ============================================================================
// SETUP COMMAND (TUI WIZARD)
// ============================================================================

func runVLLMSetup(cmd *cobra.Command, args []string) error {
	result, err := tui.RunVLLMSetup()
	if err != nil {
		return fmt.Errorf("setup wizard error: %w", err)
	}

	if result == nil {
		// User cancelled
		fmt.Println(theme.DimTextStyle.Render("Setup cancelled"))
		return nil
	}

	// Apply the configuration and start vLLM
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Starting vLLM with selected configuration..."))
	fmt.Println()

	// Set the global variables from TUI result
	vllmTPSize = result.TensorParallel
	vllmGPUMemUtil = result.GPUMemory
	vllmPort = result.Port
	vllmBackground = result.Background
	vllmMaxModelLen = result.MaxModelLen
	vllmSwapSpace = result.SwapSpace
	vllmEnableLoRA = result.EnableLora
	vllmMaxLoRARank = result.LoraRank

	if result.Quantization != "none" {
		vllmQuantization = result.Quantization
	}
	if result.DType != "auto" {
		vllmDType = result.DType
	}

	// Call the start command with the model
	return runVLLMStart(cmd, []string{result.Model})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func isVLLMRunning() bool {
	cmd := exec.Command("pgrep", "-f", "vllm serve")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) != ""
}

func isVLLMInstalled() bool {
	cmd := exec.Command("which", "vllm")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) != ""
}

func detectVLLMGPUCount() int {
	// Use centralized GPU detection (cached)
	count := gpu.GetCount()
	if count == 0 {
		return 1
	}
	return count
}

func checkVLLMHealth(port int) bool {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getVLLMModels(port int) ([]string, error) {
	url := fmt.Sprintf("http://localhost:%d/v1/models", port)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

func getGPUMemoryUsage() []string {
	// Use centralized GPU module (not cached - memory usage changes frequently)
	usage := gpu.GetMemoryUsage()
	if len(usage) == 0 {
		return []string{"Unable to get GPU info"}
	}

	var results []string
	for _, mu := range usage {
		pct := float64(0)
		if mu.TotalMiB > 0 {
			pct = float64(mu.UsedMiB) / float64(mu.TotalMiB) * 100
		}
		results = append(results, fmt.Sprintf("GPU %d (%s): %d / %d MB (%.1f%%)",
			mu.Index, mu.Name, mu.UsedMiB, mu.TotalMiB, pct))
	}
	return results
}

func saveVLLMConfig(config VLLMConfig) error {
	configPath := filepath.Join(os.TempDir(), "vllm.config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func loadVLLMConfig() VLLMConfig {
	configPath := filepath.Join(os.TempDir(), "vllm.config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return VLLMConfig{}
	}

	var config VLLMConfig
	json.Unmarshal(data, &config)
	return config
}

func showVLLMModelSelection() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("SELECT MODEL"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Available models:"))
	fmt.Println()

	models := []struct {
		shortcut string
		name     string
		vram     string
	}{
		{"llama-70b", "Llama 3.3 70B (Best quality)", "48GB+"},
		{"qwen-72b", "Qwen 2.5 72B", "48GB+"},
		{"qwen-32b", "Qwen 2.5 32B", "24GB"},
		{"mistral-7b", "Mistral 7B", "8GB"},
		{"mixtral", "Mixtral 8x7B", "32GB"},
		{"deepseek-r1", "DeepSeek R1 (Reasoning)", "48GB+"},
		{"phi-4", "Phi-4 14B", "16GB"},
	}

	for i, m := range models {
		fmt.Printf("  %d. %s %s\n",
			i+1,
			theme.HighlightStyle.Render(m.shortcut),
			theme.DimTextStyle.Render(fmt.Sprintf("- %s (VRAM: %s)", m.name, m.vram)))
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Usage:"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm start llama-70b"))
	fmt.Println(theme.DimTextStyle.Render("  anime vllm start qwen-32b --tp 2"))
	fmt.Println()

	// Interactive prompt
	fmt.Print(theme.InfoStyle.Render("Enter model name or number: "))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	// Check if number
	if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= len(models) {
		input = models[num-1].shortcut
	}

	vllmBackground = true
	return runVLLMStart(nil, []string{input})
}
