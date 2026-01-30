// Package gpu provides cached GPU detection and information retrieval.
// All GPU queries are performed once and cached for the lifetime of the process.
package gpu

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Info contains information about a single GPU
type Info struct {
	Index      int
	Name       string
	VRAM       int    // Total VRAM in GB
	VRAMMiB    int    // Total VRAM in MiB (more precise)
	Generation string // Ada Lovelace, Hopper, Ampere, etc.
	TensorCore bool
}

// MemoryUsage contains current GPU memory usage
type MemoryUsage struct {
	Index    int
	Name     string
	UsedMiB  int
	TotalMiB int
	UsedGB   int
	TotalGB  int
}

// SystemInfo contains overall GPU system information
type SystemInfo struct {
	GPUs          []Info
	Count         int
	TotalVRAM     int // Total VRAM in GB across all GPUs
	DriverVersion string
	CUDAVersion   string
	Available     bool
}

var (
	systemOnce sync.Once
	systemInfo *SystemInfo

	memoryOnce  sync.Once
	memoryUsage []MemoryUsage
)

// GetSystemInfo returns cached GPU system information.
// This function is safe to call concurrently and will only
// execute nvidia-smi once per process lifetime.
func GetSystemInfo() *SystemInfo {
	systemOnce.Do(func() {
		systemInfo = detectGPUs()
	})
	return systemInfo
}

// GetCount returns the number of GPUs available (cached)
func GetCount() int {
	return GetSystemInfo().Count
}

// GetGPUs returns information about all GPUs (cached)
func GetGPUs() []Info {
	return GetSystemInfo().GPUs
}

// GetTotalVRAM returns total VRAM in GB across all GPUs (cached)
func GetTotalVRAM() int {
	return GetSystemInfo().TotalVRAM
}

// IsAvailable returns true if NVIDIA GPUs are detected (cached)
func IsAvailable() bool {
	return GetSystemInfo().Available
}

// GetMemoryUsage returns current memory usage for all GPUs.
// Unlike other functions, this refreshes on each call as memory
// usage changes frequently.
func GetMemoryUsage() []MemoryUsage {
	return fetchMemoryUsage()
}

// GetMemoryUsageCached returns cached memory usage (only fetched once).
// Use GetMemoryUsage() if you need current values.
func GetMemoryUsageCached() []MemoryUsage {
	memoryOnce.Do(func() {
		memoryUsage = fetchMemoryUsage()
	})
	return memoryUsage
}

// RefreshSystemInfo forces a refresh of GPU information.
// Use sparingly as this spawns a subprocess.
func RefreshSystemInfo() *SystemInfo {
	systemOnce = sync.Once{} // Reset the once
	return GetSystemInfo()
}

// detectGPUs performs the actual GPU detection via nvidia-smi
func detectGPUs() *SystemInfo {
	info := &SystemInfo{
		GPUs:      make([]Info, 0),
		Available: false,
	}

	// Query GPU information
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return info
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) >= 3 {
			idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			name := strings.TrimSpace(parts[1])
			vramMiB, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
			vramGB := vramMiB / 1024

			gpu := Info{
				Index:      idx,
				Name:       name,
				VRAM:       vramGB,
				VRAMMiB:    vramMiB,
				Generation: detectGeneration(name),
				TensorCore: hasTensorCores(name),
			}
			info.GPUs = append(info.GPUs, gpu)
			info.TotalVRAM += vramGB
		}
	}

	info.Count = len(info.GPUs)
	info.Available = info.Count > 0

	// Get driver version
	if driverOutput, err := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(driverOutput)), "\n")
		if len(lines) > 0 {
			info.DriverVersion = strings.TrimSpace(lines[0])
		}
	}

	// Get CUDA version from nvidia-smi output
	if cudaOutput, err := exec.Command("bash", "-c", "nvidia-smi | grep 'CUDA Version' | awk '{print $9}'").Output(); err == nil {
		info.CUDAVersion = strings.TrimSpace(string(cudaOutput))
	}

	return info
}

// fetchMemoryUsage gets current GPU memory usage
func fetchMemoryUsage() []MemoryUsage {
	var usage []MemoryUsage

	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.used,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return usage
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) >= 4 {
			idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			name := strings.TrimSpace(parts[1])
			usedMiB, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
			totalMiB, _ := strconv.Atoi(strings.TrimSpace(parts[3]))

			usage = append(usage, MemoryUsage{
				Index:    idx,
				Name:     name,
				UsedMiB:  usedMiB,
				TotalMiB: totalMiB,
				UsedGB:   usedMiB / 1024,
				TotalGB:  totalMiB / 1024,
			})
		}
	}

	return usage
}

// detectGeneration determines GPU generation from name
func detectGeneration(name string) string {
	nameLower := strings.ToLower(name)

	// Blackwell (B100, B200)
	if strings.Contains(nameLower, "b100") || strings.Contains(nameLower, "b200") {
		return "Blackwell"
	}
	// Hopper (H100, H200, GH200)
	if strings.Contains(nameLower, "h100") || strings.Contains(nameLower, "h200") || strings.Contains(nameLower, "gh200") {
		return "Hopper"
	}
	// Ada Lovelace (RTX 40xx, L40, L4)
	if strings.Contains(nameLower, "4090") || strings.Contains(nameLower, "4080") ||
		strings.Contains(nameLower, "4070") || strings.Contains(nameLower, "4060") ||
		strings.Contains(nameLower, "l40") || strings.Contains(nameLower, "l4") ||
		strings.Contains(nameLower, "ada") {
		return "Ada Lovelace"
	}
	// Ampere (RTX 30xx, A100, A10, A40)
	if strings.Contains(nameLower, "3090") || strings.Contains(nameLower, "3080") ||
		strings.Contains(nameLower, "3070") || strings.Contains(nameLower, "3060") ||
		strings.Contains(nameLower, "a100") || strings.Contains(nameLower, "a10") ||
		strings.Contains(nameLower, "a40") || strings.Contains(nameLower, "a30") {
		return "Ampere"
	}
	// Turing (RTX 20xx, T4)
	if strings.Contains(nameLower, "2080") || strings.Contains(nameLower, "2070") ||
		strings.Contains(nameLower, "2060") || strings.Contains(nameLower, "t4") {
		return "Turing"
	}
	// Volta (V100)
	if strings.Contains(nameLower, "v100") {
		return "Volta"
	}

	return "Unknown"
}

// hasTensorCores determines if GPU has tensor cores
func hasTensorCores(name string) bool {
	nameLower := strings.ToLower(name)
	// All datacenter GPUs and RTX cards have tensor cores
	return strings.Contains(nameLower, "h100") || strings.Contains(nameLower, "h200") ||
		strings.Contains(nameLower, "a100") || strings.Contains(nameLower, "a10") ||
		strings.Contains(nameLower, "l40") || strings.Contains(nameLower, "l4") ||
		strings.Contains(nameLower, "v100") || strings.Contains(nameLower, "t4") ||
		strings.Contains(nameLower, "rtx") || strings.Contains(nameLower, "gh200") ||
		strings.Contains(nameLower, "b100") || strings.Contains(nameLower, "b200")
}

// FormatMemoryUsage returns a formatted string for memory usage display
func FormatMemoryUsage(mu MemoryUsage) string {
	percent := 0
	if mu.TotalMiB > 0 {
		percent = (mu.UsedMiB * 100) / mu.TotalMiB
	}
	return strings.TrimSpace(strings.Join([]string{
		"GPU " + strconv.Itoa(mu.Index) + ": " + mu.Name,
		strconv.Itoa(mu.UsedMiB) + "/" + strconv.Itoa(mu.TotalMiB) + " MiB",
		"(" + strconv.Itoa(percent) + "%)",
	}, " "))
}

// DefaultGPUConfig returns a default GPU configuration when no GPUs are detected
func DefaultGPUConfig() Info {
	return Info{
		Index:      0,
		Name:       "Default GPU",
		VRAM:       24,
		VRAMMiB:    24576,
		Generation: "Unknown",
		TensorCore: true,
	}
}
