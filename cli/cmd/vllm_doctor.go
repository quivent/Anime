package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/joshkornreich/anime/internal/hf"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// ============================================================================
// DOCTOR COMMAND
// ============================================================================

var (
	doctorFix     bool
	doctorVerbose bool
	doctorJSON    bool
)

var vllmDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and fix vLLM installation issues",
	Long: `Comprehensive diagnostic tool for vLLM installation.

Checks:
  - Python version and environment
  - NumPy version and compatibility
  - PyTorch installation and CUDA support
  - vLLM installation status
  - GPU detection and memory
  - CUDA toolkit and drivers
  - Dependency conflicts
  - HuggingFace authentication
  - System resources

Use --fix to automatically resolve detected issues.

Examples:
  anime vllm doctor              # Run diagnostics
  anime vllm doctor --fix        # Diagnose and fix issues
  anime vllm doctor --verbose    # Show detailed output
  anime vllm doctor --json       # Output as JSON`,
	RunE: runVLLMDoctor,
}

func init() {
	vllmDoctorCmd.Flags().BoolVarP(&doctorFix, "fix", "f", false, "Automatically fix detected issues")
	vllmDoctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Show detailed diagnostic output")
	vllmDoctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	vllmCmd.AddCommand(vllmDoctorCmd)
}

// ============================================================================
// DIAGNOSTIC TYPES
// ============================================================================

type DiagnosticResult struct {
	Name        string   `json:"name"`
	Status      string   `json:"status"` // pass, warn, fail, skip
	Message     string   `json:"message"`
	Details     []string `json:"details,omitempty"`
	FixCommand  string   `json:"fix_command,omitempty"`
	FixFunction func() error `json:"-"`
}

type DoctorReport struct {
	Timestamp    string              `json:"timestamp"`
	Platform     string              `json:"platform"`
	Architecture string              `json:"architecture"`
	Results      []DiagnosticResult  `json:"results"`
	Summary      DiagnosticSummary   `json:"summary"`
}

type DiagnosticSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Warned  int `json:"warned"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// ============================================================================
// MAIN DOCTOR FUNCTION
// ============================================================================

func runVLLMDoctor(cmd *cobra.Command, args []string) error {
	if !doctorJSON {
		fmt.Println()
		fmt.Println(theme.RenderBanner("VLLM DOCTOR"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Running comprehensive diagnostics..."))
		fmt.Println()
	}

	report := DoctorReport{
		Timestamp:    "now",
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
		Results:      []DiagnosticResult{},
	}

	// Run all diagnostics
	diagnostics := []func() DiagnosticResult{
		checkPythonVersion,
		checkPythonEnvironment,
		checkPipVersion,
		checkNumPyInstallation,
		checkNumPyVersion,
		checkNumPyBLAS,
		checkPyTorchInstallation,
		checkPyTorchCUDA,
		checkVLLMInstallation,
		checkVLLMVersion,
		checkCUDAToolkit,
		checkNVIDIADriver,
		checkGPUDetection,
		checkGPUMemory,
		checkDependencyConflicts,
		checkTransformersVersion,
		checkTokenizersVersion,
		checkHuggingFaceAuth,
		checkDiskSpace,
		checkSystemMemory,
		checkFlashAttention,
		checkTriton,
	}

	for _, diag := range diagnostics {
		result := diag()
		report.Results = append(report.Results, result)

		if !doctorJSON {
			printDiagnosticResult(result)
		}

		// Update summary
		switch result.Status {
		case "pass":
			report.Summary.Passed++
		case "warn":
			report.Summary.Warned++
		case "fail":
			report.Summary.Failed++
		case "skip":
			report.Summary.Skipped++
		}
		report.Summary.Total++
	}

	// Output JSON if requested
	if doctorJSON {
		jsonOutput, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	// Print summary
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.HighlightStyle.Render("  DIAGNOSTIC SUMMARY"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s %d\n", theme.SuccessStyle.Render("Passed:"), report.Summary.Passed)
	fmt.Printf("  %s %d\n", theme.WarningStyle.Render("Warnings:"), report.Summary.Warned)
	fmt.Printf("  %s %d\n", theme.ErrorStyle.Render("Failed:"), report.Summary.Failed)
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Skipped:"), report.Summary.Skipped)
	fmt.Println()

	// Collect fixable issues
	var fixableIssues []DiagnosticResult
	for _, result := range report.Results {
		if (result.Status == "fail" || result.Status == "warn") && result.FixCommand != "" {
			fixableIssues = append(fixableIssues, result)
		}
	}

	if len(fixableIssues) > 0 {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.HighlightStyle.Render("  RECOMMENDED FIXES"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for i, issue := range fixableIssues {
			fmt.Printf("  %d. %s\n", i+1, theme.WarningStyle.Render(issue.Name))
			fmt.Printf("     %s\n", theme.DimTextStyle.Render(issue.Message))
			fmt.Printf("     %s %s\n", theme.InfoStyle.Render("Fix:"), theme.HighlightStyle.Render(issue.FixCommand))
			fmt.Println()
		}

		if doctorFix {
			fmt.Println(theme.InfoStyle.Render("Applying fixes..."))
			fmt.Println()
			applyFixes(fixableIssues)
		} else {
			fmt.Println(theme.InfoStyle.Render("Run with --fix to automatically apply these fixes:"))
			fmt.Println(theme.HighlightStyle.Render("  anime vllm doctor --fix"))
			fmt.Println()
		}
	} else if report.Summary.Failed == 0 {
		fmt.Println(theme.SuccessStyle.Render("All checks passed! vLLM should work correctly."))
		fmt.Println()
	}

	return nil
}

func printDiagnosticResult(result DiagnosticResult) {
	var statusIcon string
	var styledIcon string

	switch result.Status {
	case "pass":
		statusIcon = "✓"
		styledIcon = theme.SuccessStyle.Render(statusIcon)
	case "warn":
		statusIcon = "⚠"
		styledIcon = theme.WarningStyle.Render(statusIcon)
	case "fail":
		statusIcon = "✗"
		styledIcon = theme.ErrorStyle.Render(statusIcon)
	case "skip":
		statusIcon = "○"
		styledIcon = theme.DimTextStyle.Render(statusIcon)
	}

	fmt.Printf("  %s %s\n", styledIcon, result.Name)
	if result.Message != "" && (result.Status != "pass" || doctorVerbose) {
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(result.Message))
	}

	if doctorVerbose && len(result.Details) > 0 {
		for _, detail := range result.Details {
			fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("•"), detail)
		}
	}
}

// ============================================================================
// PYTHON DIAGNOSTICS
// ============================================================================

func checkPythonVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Python Version",
	}

	output, err := runCommand("python3", "--version")
	if err != nil {
		output, err = runCommand("python", "--version")
	}

	if err != nil {
		result.Status = "fail"
		result.Message = "Python not found"
		result.FixCommand = "Install Python 3.9-3.12"
		return result
	}

	version := strings.TrimPrefix(strings.TrimSpace(output), "Python ")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Could not parse version: %s", version)
		return result
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])

	result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))

	if major != 3 {
		result.Status = "fail"
		result.Message = fmt.Sprintf("Python 3 required, found Python %d", major)
		result.FixCommand = "Install Python 3.9-3.12"
		return result
	}

	if minor < 9 {
		result.Status = "fail"
		result.Message = fmt.Sprintf("Python 3.9+ required, found 3.%d", minor)
		result.FixCommand = "Install Python 3.9-3.12"
		return result
	}

	if minor > 12 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Python 3.%d may have compatibility issues, 3.11 recommended", minor)
		return result
	}

	if minor == 11 {
		result.Status = "pass"
		result.Message = fmt.Sprintf("Python %s (optimal)", version)
	} else {
		result.Status = "pass"
		result.Message = fmt.Sprintf("Python %s", version)
	}

	return result
}

func checkPythonEnvironment() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Python Environment",
	}

	// Check if in virtual environment or conda
	venv := os.Getenv("VIRTUAL_ENV")
	conda := os.Getenv("CONDA_DEFAULT_ENV")
	condaPrefix := os.Getenv("CONDA_PREFIX")

	if venv != "" {
		result.Status = "pass"
		result.Message = fmt.Sprintf("Virtual environment: %s", filepath.Base(venv))
		result.Details = append(result.Details, fmt.Sprintf("Path: %s", venv))
		return result
	}

	if conda != "" || condaPrefix != "" {
		envName := conda
		if envName == "" {
			envName = filepath.Base(condaPrefix)
		}
		result.Status = "pass"
		result.Message = fmt.Sprintf("Conda environment: %s", envName)
		result.Details = append(result.Details, fmt.Sprintf("Path: %s", condaPrefix))
		return result
	}

	result.Status = "warn"
	result.Message = "No virtual environment detected"
	result.Details = append(result.Details, "Using system Python may cause conflicts")
	result.FixCommand = "conda create -n vllm python=3.11 -y && conda activate vllm"

	return result
}

func checkPipVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Pip Version",
	}

	output, err := runCommand("pip", "--version")
	if err != nil {
		result.Status = "fail"
		result.Message = "pip not found"
		result.FixCommand = "python -m ensurepip --upgrade"
		return result
	}

	// Extract version: "pip 24.0 from /path..."
	re := regexp.MustCompile(`pip (\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 3 {
		result.Status = "warn"
		result.Message = "Could not parse pip version"
		return result
	}

	major, _ := strconv.Atoi(matches[1])
	result.Details = append(result.Details, fmt.Sprintf("Version: %s.%s", matches[1], matches[2]))

	if major < 21 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("pip %d is outdated, recommend 23+", major)
		result.FixCommand = "pip install --upgrade pip"
		return result
	}

	result.Status = "pass"
	result.Message = fmt.Sprintf("pip %s.%s", matches[1], matches[2])
	return result
}

// ============================================================================
// NUMPY DIAGNOSTICS
// ============================================================================

func checkNumPyInstallation() DiagnosticResult {
	result := DiagnosticResult{
		Name: "NumPy Installation",
	}

	output, err := runPythonCode(`
import sys
try:
    import numpy as np
    print(f"OK:{np.__version__}:{np.__file__}")
except ImportError as e:
    print(f"MISSING:{e}")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "fail"
		result.Message = "Could not check NumPy"
		result.Details = append(result.Details, err.Error())
		return result
	}

	output = strings.TrimSpace(output)

	if strings.HasPrefix(output, "MISSING:") {
		result.Status = "fail"
		result.Message = "NumPy not installed"
		result.FixCommand = "pip install 'numpy<2.0'"
		return result
	}

	if strings.HasPrefix(output, "ERROR:") {
		errMsg := strings.TrimPrefix(output, "ERROR:")
		result.Status = "fail"
		result.Message = fmt.Sprintf("NumPy import error: %s", errMsg)

		if strings.Contains(errMsg, "multiarray") {
			result.FixCommand = "pip uninstall numpy -y && pip install 'numpy<2.0' --force-reinstall --no-cache-dir"
		} else {
			result.FixCommand = "pip install 'numpy<2.0' --force-reinstall --no-cache-dir"
		}
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		parts := strings.Split(strings.TrimPrefix(output, "OK:"), ":")
		if len(parts) >= 1 {
			result.Status = "pass"
			result.Message = fmt.Sprintf("NumPy %s installed", parts[0])
			if len(parts) >= 2 {
				result.Details = append(result.Details, fmt.Sprintf("Location: %s", parts[1]))
			}
		}
		return result
	}

	result.Status = "warn"
	result.Message = fmt.Sprintf("Unexpected output: %s", output)
	return result
}

func checkNumPyVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "NumPy Version Compatibility",
	}

	output, err := runPythonCode(`
import numpy as np
version = np.__version__
parts = version.split('.')
major = int(parts[0])
minor = int(parts[1]) if len(parts) > 1 else 0
print(f"{major}:{minor}:{version}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "NumPy not available"
		return result
	}

	parts := strings.Split(strings.TrimSpace(output), ":")
	if len(parts) < 3 {
		result.Status = "warn"
		result.Message = "Could not parse NumPy version"
		return result
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	version := parts[2]

	result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))

	// Note: vLLM 0.13.0+ works with NumPy 2.x, only older vLLM versions required NumPy <2.0
	if major >= 2 {
		result.Status = "pass"
		result.Message = fmt.Sprintf("NumPy %s (vLLM 0.13.0+ compatible)", version)
		return result
	}

	if major == 1 && minor < 21 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("NumPy %s is old, recommend 1.26.x", version)
		result.FixCommand = "pip install 'numpy>=1.26,<2.0'"
		return result
	}

	if major == 1 && minor >= 26 {
		result.Status = "pass"
		result.Message = fmt.Sprintf("NumPy %s (optimal)", version)
	} else {
		result.Status = "pass"
		result.Message = fmt.Sprintf("NumPy %s", version)
	}

	return result
}

func checkNumPyBLAS() DiagnosticResult {
	result := DiagnosticResult{
		Name: "NumPy BLAS Configuration",
	}

	output, err := runPythonCode(`
import numpy as np
try:
    config = np.show_config(mode='dicts')
    if config:
        blas = config.get('Build Dependencies', {}).get('blas', {})
        if blas:
            print(f"OK:{blas.get('name', 'unknown')}")
        else:
            # Try older numpy config
            import numpy.distutils.system_info as sysinfo
            blas_info = sysinfo.get_info('blas')
            if blas_info:
                print(f"OK:{blas_info.get('libraries', ['unknown'])[0]}")
            else:
                print("OK:unknown")
    else:
        print("OK:unknown")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "Could not check BLAS"
		return result
	}

	output = strings.TrimSpace(output)

	if strings.HasPrefix(output, "ERROR:") {
		result.Status = "warn"
		result.Message = "Could not determine BLAS library"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		blas := strings.TrimPrefix(output, "OK:")
		result.Details = append(result.Details, fmt.Sprintf("BLAS: %s", blas))

		if strings.Contains(strings.ToLower(blas), "openblas") ||
		   strings.Contains(strings.ToLower(blas), "mkl") ||
		   strings.Contains(strings.ToLower(blas), "accelerate") {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Using %s", blas)
		} else if blas == "unknown" {
			result.Status = "warn"
			result.Message = "BLAS library not detected"
			result.Details = append(result.Details, "Performance may be suboptimal")
			if runtime.GOOS == "linux" {
				result.FixCommand = "conda install openblas -c conda-forge"
			}
		} else {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Using %s", blas)
		}
		return result
	}

	result.Status = "warn"
	result.Message = "Could not determine BLAS configuration"
	return result
}

// ============================================================================
// PYTORCH DIAGNOSTICS
// ============================================================================

func checkPyTorchInstallation() DiagnosticResult {
	result := DiagnosticResult{
		Name: "PyTorch Installation",
	}

	output, err := runPythonCode(`
try:
    import torch
    print(f"OK:{torch.__version__}")
except ImportError:
    print("MISSING")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "fail"
		result.Message = "Could not check PyTorch"
		return result
	}

	output = strings.TrimSpace(output)

	if output == "MISSING" {
		result.Status = "fail"
		result.Message = "PyTorch not installed"
		result.FixCommand = "pip install torch"
		return result
	}

	if strings.HasPrefix(output, "ERROR:") {
		result.Status = "fail"
		result.Message = fmt.Sprintf("PyTorch error: %s", strings.TrimPrefix(output, "ERROR:"))
		result.FixCommand = "pip install torch --force-reinstall"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("PyTorch %s", version)
		result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))
	}

	return result
}

func checkPyTorchCUDA() DiagnosticResult {
	result := DiagnosticResult{
		Name: "PyTorch CUDA Support",
	}

	output, err := runPythonCode(`
try:
    import torch
    cuda_available = torch.cuda.is_available()
    cuda_version = torch.version.cuda if torch.version.cuda else "N/A"
    device_count = torch.cuda.device_count() if cuda_available else 0

    if cuda_available:
        devices = []
        for i in range(device_count):
            name = torch.cuda.get_device_name(i)
            mem = torch.cuda.get_device_properties(i).total_memory / (1024**3)
            devices.append(f"{name}:{mem:.1f}GB")
        print(f"OK:{cuda_version}:{device_count}:{';'.join(devices)}")
    else:
        print(f"NOCUDA:{cuda_version}")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "PyTorch not available"
		return result
	}

	output = strings.TrimSpace(output)

	if strings.HasPrefix(output, "ERROR:") {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Error checking CUDA: %s", strings.TrimPrefix(output, "ERROR:"))
		return result
	}

	if strings.HasPrefix(output, "NOCUDA:") {
		cudaVer := strings.TrimPrefix(output, "NOCUDA:")
		result.Status = "fail"
		result.Message = "CUDA not available in PyTorch"
		result.Details = append(result.Details, fmt.Sprintf("PyTorch CUDA version: %s", cudaVer))
		result.Details = append(result.Details, "PyTorch may be CPU-only build")
		result.FixCommand = "pip install torch --index-url https://download.pytorch.org/whl/cu121"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		parts := strings.Split(strings.TrimPrefix(output, "OK:"), ":")
		if len(parts) >= 3 {
			cudaVer := parts[0]
			deviceCount := parts[1]
			devices := parts[2]

			result.Status = "pass"
			result.Message = fmt.Sprintf("CUDA %s with %s GPU(s)", cudaVer, deviceCount)
			result.Details = append(result.Details, fmt.Sprintf("CUDA Version: %s", cudaVer))
			result.Details = append(result.Details, fmt.Sprintf("GPU Count: %s", deviceCount))

			if devices != "" {
				for _, dev := range strings.Split(devices, ";") {
					result.Details = append(result.Details, fmt.Sprintf("Device: %s", dev))
				}
			}
		}
	}

	return result
}

// ============================================================================
// VLLM DIAGNOSTICS
// ============================================================================

func checkVLLMInstallation() DiagnosticResult {
	result := DiagnosticResult{
		Name: "vLLM Installation",
	}

	output, err := runPythonCode(`
try:
    import vllm
    print(f"OK:{vllm.__version__}")
except ImportError as e:
    print(f"MISSING:{e}")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "fail"
		result.Message = "Could not check vLLM"
		result.Details = append(result.Details, err.Error())
		return result
	}

	output = strings.TrimSpace(output)

	if strings.HasPrefix(output, "MISSING:") {
		errMsg := strings.TrimPrefix(output, "MISSING:")
		result.Status = "fail"
		result.Message = "vLLM not installed"
		result.Details = append(result.Details, errMsg)
		result.FixCommand = "pip install vllm"
		return result
	}

	if strings.HasPrefix(output, "ERROR:") {
		errMsg := strings.TrimPrefix(output, "ERROR:")
		result.Status = "fail"
		result.Message = fmt.Sprintf("vLLM import error: %s", errMsg)

		// Diagnose specific errors
		if strings.Contains(errMsg, "numpy") {
			result.FixCommand = "pip install 'numpy<2.0' --force-reinstall && pip install vllm --force-reinstall"
			result.Details = append(result.Details, "NumPy compatibility issue detected")
		} else if strings.Contains(errMsg, "torch") {
			result.FixCommand = "pip install torch && pip install vllm --force-reinstall"
			result.Details = append(result.Details, "PyTorch issue detected")
		} else {
			result.FixCommand = "pip install vllm --force-reinstall --no-cache-dir"
		}
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("vLLM %s", version)
		result.Details = append(result.Details, fmt.Sprintf("Version: %s", version))
	}

	return result
}

func checkVLLMVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "vLLM Version",
	}

	output, err := runPythonCode(`
try:
    import vllm
    version = vllm.__version__
    parts = version.split('.')
    major = int(parts[0])
    minor = int(parts[1]) if len(parts) > 1 else 0
    print(f"{major}:{minor}:{version}")
except:
    print("SKIP")
`)

	if err != nil || strings.TrimSpace(output) == "SKIP" {
		result.Status = "skip"
		result.Message = "vLLM not available"
		return result
	}

	parts := strings.Split(strings.TrimSpace(output), ":")
	if len(parts) < 3 {
		result.Status = "warn"
		result.Message = "Could not parse vLLM version"
		return result
	}

	minor, _ := strconv.Atoi(parts[1])
	version := parts[2]

	// Check for known issues with specific versions
	if minor < 4 {
		result.Status = "warn"
		result.Message = fmt.Sprintf("vLLM %s is outdated, recommend 0.6+", version)
		result.FixCommand = "pip install --upgrade vllm"
		return result
	}

	result.Status = "pass"
	result.Message = fmt.Sprintf("vLLM %s", version)
	return result
}

// ============================================================================
// CUDA/GPU DIAGNOSTICS
// ============================================================================

func checkCUDAToolkit() DiagnosticResult {
	result := DiagnosticResult{
		Name: "CUDA Toolkit",
	}

	output, err := runCommand("nvcc", "--version")
	if err != nil {
		// Try nvidia-smi as fallback
		smiOutput, smiErr := runCommand("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
		if smiErr != nil {
			result.Status = "warn"
			result.Message = "CUDA toolkit not found (nvcc not in PATH)"
			result.Details = append(result.Details, "This may be OK if using PyTorch's bundled CUDA")
			return result
		}
		result.Status = "warn"
		result.Message = "nvcc not found, but NVIDIA driver present"
		result.Details = append(result.Details, fmt.Sprintf("Driver: %s", strings.TrimSpace(smiOutput)))
		return result
	}

	// Parse CUDA version from nvcc output
	re := regexp.MustCompile(`release (\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) >= 2 {
		cudaVersion := matches[1]
		result.Details = append(result.Details, fmt.Sprintf("Version: %s", cudaVersion))

		major, _ := strconv.ParseFloat(cudaVersion, 64)
		if major < 11.8 {
			result.Status = "warn"
			result.Message = fmt.Sprintf("CUDA %s may have compatibility issues, 12.x recommended", cudaVersion)
		} else {
			result.Status = "pass"
			result.Message = fmt.Sprintf("CUDA %s", cudaVersion)
		}
	} else {
		result.Status = "pass"
		result.Message = "CUDA toolkit found"
	}

	return result
}

func checkNVIDIADriver() DiagnosticResult {
	result := DiagnosticResult{
		Name: "NVIDIA Driver",
	}

	output, err := runCommand("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	if err != nil {
		result.Status = "fail"
		result.Message = "NVIDIA driver not found"
		result.Details = append(result.Details, "nvidia-smi command failed")
		result.FixCommand = "Install NVIDIA drivers for your GPU"
		return result
	}

	driverVersion := strings.TrimSpace(strings.Split(output, "\n")[0])
	result.Details = append(result.Details, fmt.Sprintf("Version: %s", driverVersion))

	// Parse driver version
	parts := strings.Split(driverVersion, ".")
	if len(parts) >= 1 {
		major, _ := strconv.Atoi(parts[0])
		if major < 525 {
			result.Status = "warn"
			result.Message = fmt.Sprintf("Driver %s is old, recommend 535+", driverVersion)
			result.FixCommand = "Update NVIDIA drivers"
		} else {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Driver %s", driverVersion)
		}
	} else {
		result.Status = "pass"
		result.Message = fmt.Sprintf("Driver %s", driverVersion)
	}

	return result
}

func checkGPUDetection() DiagnosticResult {
	result := DiagnosticResult{
		Name: "GPU Detection",
	}

	output, err := runCommand("nvidia-smi", "--query-gpu=name,compute_cap", "--format=csv,noheader")
	if err != nil {
		result.Status = "fail"
		result.Message = "No NVIDIA GPUs detected"
		return result
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	gpuCount := len(lines)

	result.Details = append(result.Details, fmt.Sprintf("GPU Count: %d", gpuCount))

	var hasCompatibleGPU bool
	for i, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			computeCap := strings.TrimSpace(parts[1])
			result.Details = append(result.Details, fmt.Sprintf("GPU %d: %s (SM %s)", i, name, computeCap))

			// Check compute capability (vLLM needs SM 7.0+)
			capParts := strings.Split(computeCap, ".")
			if len(capParts) >= 1 {
				major, _ := strconv.Atoi(capParts[0])
				if major >= 7 {
					hasCompatibleGPU = true
				}
			}
		}
	}

	if !hasCompatibleGPU {
		result.Status = "fail"
		result.Message = "No compatible GPU found (need SM 7.0+)"
		result.Details = append(result.Details, "vLLM requires NVIDIA GPU with compute capability 7.0+")
		result.Details = append(result.Details, "Volta (V100), Turing (RTX 20xx), Ampere (A100, RTX 30xx), Ada (RTX 40xx), Hopper (H100)")
		return result
	}

	result.Status = "pass"
	result.Message = fmt.Sprintf("%d compatible GPU(s) found", gpuCount)
	return result
}

func checkGPUMemory() DiagnosticResult {
	result := DiagnosticResult{
		Name: "GPU Memory",
	}

	output, err := runCommand("nvidia-smi", "--query-gpu=index,name,memory.total,memory.used,memory.free", "--format=csv,noheader,nounits")
	if err != nil {
		result.Status = "skip"
		result.Message = "Could not query GPU memory"
		return result
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var totalMem, freeMem float64
	var hasLowMemory bool

	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) >= 5 {
			idx := strings.TrimSpace(parts[0])
			name := strings.TrimSpace(parts[1])
			total, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			used, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			free, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)

			totalMem += total
			freeMem += free

			usedPct := (used / total) * 100
			result.Details = append(result.Details,
				fmt.Sprintf("GPU %s (%s): %.0f/%.0f MB used (%.1f%%), %.0f MB free",
					idx, name, used, total, usedPct, free))

			if free < 4000 { // Less than 4GB free
				hasLowMemory = true
			}
		}
	}

	if hasLowMemory {
		result.Status = "warn"
		result.Message = fmt.Sprintf("Low GPU memory: %.1f GB free of %.1f GB total", freeMem/1024, totalMem/1024)
		result.Details = append(result.Details, "Consider stopping other GPU processes")
		result.FixCommand = "nvidia-smi --query-compute-apps=pid,name,used_memory --format=csv"
	} else {
		result.Status = "pass"
		result.Message = fmt.Sprintf("%.1f GB free of %.1f GB total", freeMem/1024, totalMem/1024)
	}

	return result
}

// ============================================================================
// DEPENDENCY DIAGNOSTICS
// ============================================================================

func checkDependencyConflicts() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Dependency Conflicts",
	}

	output, err := runCommand("pip", "check")
	if err != nil {
		// pip check returns non-zero if there are conflicts
		lines := strings.Split(strings.TrimSpace(output), "\n")

		var conflicts []string
		for _, line := range lines {
			if strings.Contains(line, "requires") || strings.Contains(line, "incompatible") {
				conflicts = append(conflicts, line)
			}
		}

		if len(conflicts) > 0 {
			result.Status = "warn"
			result.Message = fmt.Sprintf("%d dependency conflict(s) found", len(conflicts))
			for _, conflict := range conflicts {
				result.Details = append(result.Details, conflict)
			}
			// Note: Use --no-deps to avoid vllm pulling in PyPI torch which is CPU-only on aarch64
		// User's torch+cu128 is correct, just upgrade vllm without its deps
		result.FixCommand = "pip install --upgrade --no-deps vllm && pip install --upgrade transformers"
			return result
		}
	}

	result.Status = "pass"
	result.Message = "No dependency conflicts"
	return result
}

func checkTransformersVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Transformers Library",
	}

	output, err := runPythonCode(`
try:
    import transformers
    print(f"OK:{transformers.__version__}")
except ImportError:
    print("MISSING")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "Could not check transformers"
		return result
	}

	output = strings.TrimSpace(output)

	if output == "MISSING" {
		result.Status = "fail"
		result.Message = "transformers not installed"
		result.FixCommand = "pip install transformers"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("transformers %s", version)
	}

	return result
}

func checkTokenizersVersion() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Tokenizers Library",
	}

	output, err := runPythonCode(`
try:
    import tokenizers
    print(f"OK:{tokenizers.__version__}")
except ImportError:
    print("MISSING")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "Could not check tokenizers"
		return result
	}

	output = strings.TrimSpace(output)

	if output == "MISSING" {
		result.Status = "warn"
		result.Message = "tokenizers not installed"
		result.FixCommand = "pip install tokenizers"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("tokenizers %s", version)
	}

	return result
}

func checkFlashAttention() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Flash Attention",
	}

	output, err := runPythonCode(`
try:
    import flash_attn
    print(f"OK:{flash_attn.__version__}")
except ImportError:
    print("MISSING")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "Could not check flash-attn"
		return result
	}

	output = strings.TrimSpace(output)

	if output == "MISSING" {
		result.Status = "warn"
		result.Message = "flash-attn not installed (optional but recommended)"
		result.Details = append(result.Details, "Flash Attention improves performance significantly")
		result.FixCommand = "pip install flash-attn --no-build-isolation"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("flash-attn %s", version)
	}

	return result
}

func checkTriton() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Triton",
	}

	output, err := runPythonCode(`
try:
    import triton
    print(f"OK:{triton.__version__}")
except ImportError:
    print("MISSING")
except Exception as e:
    print(f"ERROR:{e}")
`)

	if err != nil {
		result.Status = "skip"
		result.Message = "Could not check triton"
		return result
	}

	output = strings.TrimSpace(output)

	if output == "MISSING" {
		result.Status = "warn"
		result.Message = "triton not installed"
		result.Details = append(result.Details, "Triton is required for some vLLM operations")
		result.FixCommand = "pip install triton"
		return result
	}

	if strings.HasPrefix(output, "OK:") {
		version := strings.TrimPrefix(output, "OK:")
		result.Status = "pass"
		result.Message = fmt.Sprintf("triton %s", version)
	}

	return result
}

// ============================================================================
// SYSTEM DIAGNOSTICS
// ============================================================================

func checkHuggingFaceAuth() DiagnosticResult {
	result := DiagnosticResult{
		Name: "HuggingFace Authentication",
	}

	// Check if token is available
	token := hf.GetToken()
	if token == "" {
		// Check environment
		token = os.Getenv("HF_TOKEN")
	}
	if token == "" {
		// Check HF CLI cache
		homeDir, _ := os.UserHomeDir()
		tokenPath := filepath.Join(homeDir, ".cache", "huggingface", "token")
		if data, err := os.ReadFile(tokenPath); err == nil {
			token = strings.TrimSpace(string(data))
		}
	}

	if token == "" {
		result.Status = "warn"
		result.Message = "No HuggingFace token found"
		result.Details = append(result.Details, "Some models (Llama, etc.) require authentication")
		result.FixCommand = "huggingface-cli login"
		return result
	}

	// Verify token works
	result.Status = "pass"
	result.Message = "HuggingFace token available"
	result.Details = append(result.Details, fmt.Sprintf("Token: %s...%s", token[:4], token[len(token)-4:]))

	return result
}

func checkDiskSpace() DiagnosticResult {
	result := DiagnosticResult{
		Name: "Disk Space",
	}

	homeDir, _ := os.UserHomeDir()
	cachePath := filepath.Join(homeDir, ".cache", "huggingface")

	// Get disk usage
	output, err := runCommand("df", "-BG", cachePath)
	if err != nil {
		// Try without -BG for macOS
		output, err = runCommand("df", "-g", homeDir)
		if err != nil {
			result.Status = "skip"
			result.Message = "Could not check disk space"
			return result
		}
	}

	lines := strings.Split(output, "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			available := strings.TrimSuffix(fields[3], "G")
			availGB, _ := strconv.ParseFloat(available, 64)

			result.Details = append(result.Details, fmt.Sprintf("Available: %.0f GB", availGB))
			result.Details = append(result.Details, fmt.Sprintf("Cache path: %s", cachePath))

			if availGB < 50 {
				result.Status = "warn"
				result.Message = fmt.Sprintf("Low disk space: %.0f GB available", availGB)
				result.Details = append(result.Details, "Large models like Llama-70B need 100+ GB")
			} else if availGB < 100 {
				result.Status = "warn"
				result.Message = fmt.Sprintf("%.0f GB available (may be tight for large models)", availGB)
			} else {
				result.Status = "pass"
				result.Message = fmt.Sprintf("%.0f GB available", availGB)
			}
		}
	}

	return result
}

func checkSystemMemory() DiagnosticResult {
	result := DiagnosticResult{
		Name: "System Memory",
	}

	var totalMem, freeMem float64

	if runtime.GOOS == "darwin" {
		// macOS
		output, err := runCommand("sysctl", "-n", "hw.memsize")
		if err == nil {
			bytes, _ := strconv.ParseFloat(strings.TrimSpace(output), 64)
			totalMem = bytes / (1024 * 1024 * 1024)
		}
		// Get free memory from vm_stat
		output, _ = runCommand("vm_stat")
		// This is complex to parse, so we'll estimate
		freeMem = totalMem * 0.3 // Rough estimate
	} else {
		// Linux
		output, err := runCommand("free", "-g")
		if err == nil {
			lines := strings.Split(output, "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 7 {
					totalMem, _ = strconv.ParseFloat(fields[1], 64)
					freeMem, _ = strconv.ParseFloat(fields[6], 64) // "available" column
				}
			}
		}
	}

	if totalMem > 0 {
		result.Details = append(result.Details, fmt.Sprintf("Total: %.0f GB", totalMem))
		result.Details = append(result.Details, fmt.Sprintf("Available: ~%.0f GB", freeMem))

		if totalMem < 32 {
			result.Status = "warn"
			result.Message = fmt.Sprintf("%.0f GB RAM (32+ recommended)", totalMem)
		} else {
			result.Status = "pass"
			result.Message = fmt.Sprintf("%.0f GB RAM", totalMem)
		}
	} else {
		result.Status = "skip"
		result.Message = "Could not determine system memory"
	}

	return result
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runPythonCode(code string) (string, error) {
	cmd := exec.Command("python3", "-c", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try python instead of python3
		cmd = exec.Command("python", "-c", code)
		output, err = cmd.CombinedOutput()
	}
	return string(output), err
}

func applyFixes(issues []DiagnosticResult) {
	for i, issue := range issues {
		if issue.FixCommand == "" {
			continue
		}

		fmt.Printf("\n%s Fix %d: %s\n", theme.InfoStyle.Render("→"), i+1, issue.Name)
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(issue.FixCommand))
		fmt.Println()

		// Parse and execute the fix command
		if issue.FixFunction != nil {
			err := issue.FixFunction()
			if err != nil {
				fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("Failed:"), err)
			} else {
				fmt.Printf("  %s\n", theme.SuccessStyle.Render("Done"))
			}
			continue
		}

		// Execute shell command
		var cmd *exec.Cmd
		if strings.Contains(issue.FixCommand, "&&") || strings.Contains(issue.FixCommand, "|") {
			cmd = exec.Command("sh", "-c", issue.FixCommand)
		} else {
			parts := strings.Fields(issue.FixCommand)
			if len(parts) == 0 {
				continue
			}
			cmd = exec.Command(parts[0], parts[1:]...)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Ask for confirmation on potentially destructive commands
		if strings.Contains(issue.FixCommand, "uninstall") || strings.Contains(issue.FixCommand, "force") {
			fmt.Print(theme.WarningStyle.Render("  Execute? [y/N]: "))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println(theme.DimTextStyle.Render("  Skipped"))
				continue
			}
		}

		err := cmd.Run()
		if err != nil {
			fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("Failed:"), err)
		} else {
			fmt.Printf("  %s\n", theme.SuccessStyle.Render("Done"))
		}
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Re-run 'anime vllm doctor' to verify fixes"))
	fmt.Println()
}
