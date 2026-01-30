package hardware

import (
    "os/exec"
    "strconv"
    "strings"
)

// GPUInfo represents GPU hardware information
type GPUInfo struct {
    Index       int
    Name        string
    MemoryMB    int
    DriverVer   string
    CUDAVer     string
    NVLinkVer   int
}

// SystemInfo represents system hardware information
type SystemInfo struct {
    GPUs         []GPUInfo
    TotalGPUs    int
    TotalMemoryGB int
    HasNVLink    bool
    NVLinkBW     int
    Platform     string
}

// Detect detects system hardware
func Detect() *SystemInfo {
    info := &SystemInfo{
        GPUs:      []GPUInfo{},
        Platform:  "linux",
    }

    // Detect GPUs
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=index,name,memory.total,driver_version",
        "--format=csv,noheader,nounits")

    out, err := cmd.Output()
    if err != nil {
        return info
    }

    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    for _, line := range lines {
        parts := strings.Split(line, ", ")
        if len(parts) < 4 {
            continue
        }

        idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
        mem, _ := strconv.Atoi(strings.TrimSpace(parts[2]))

        gpu := GPUInfo{
            Index:     idx,
            Name:      strings.TrimSpace(parts[1]),
            MemoryMB:  mem,
            DriverVer: strings.TrimSpace(parts[3]),
        }
        info.GPUs = append(info.GPUs, gpu)
        info.TotalMemoryGB += mem / 1024
    }

    info.TotalGPUs = len(info.GPUs)

    // Check NVLink
    topoCmd := exec.Command("nvidia-smi", "topo", "-m")
    topoOut, err := topoCmd.Output()
    if err == nil {
        topoStr := string(topoOut)
        if strings.Contains(topoStr, "NV") {
            info.HasNVLink = true
            if strings.Contains(topoStr, "NV18") {
                info.NVLinkBW = 900 // H100 full mesh
            } else if strings.Contains(topoStr, "NV12") {
                info.NVLinkBW = 600 // A100
            } else {
                info.NVLinkBW = 300 // Older NVLink
            }
        }
    }

    return info
}

// IsH100 checks if GPUs are H100
func (s *SystemInfo) IsH100() bool {
    for _, gpu := range s.GPUs {
        if strings.Contains(strings.ToLower(gpu.Name), "h100") {
            return true
        }
    }
    return false
}

// IsA100 checks if GPUs are A100
func (s *SystemInfo) IsA100() bool {
    for _, gpu := range s.GPUs {
        if strings.Contains(strings.ToLower(gpu.Name), "a100") {
            return true
        }
    }
    return false
}

// GetRecommendedConfig returns recommended config based on hardware
func (s *SystemInfo) GetRecommendedConfig() map[string]interface{} {
    config := map[string]interface{}{
        "gpu_count":        s.TotalGPUs,
        "context_parallel": s.TotalGPUs,
        "nvlink_enabled":   s.HasNVLink,
    }

    // Model recommendation based on memory
    if s.TotalGPUs > 0 {
        memPerGPU := s.TotalMemoryGB / s.TotalGPUs
        if memPerGPU >= 80 {
            config["model"] = "SkyReels-V2-DF-14B-540P"
            config["precision"] = "fp8"
        } else if memPerGPU >= 48 {
            config["model"] = "SkyReels-V2-DF-14B-540P"
            config["precision"] = "fp8"
            config["offload"] = true
        } else if memPerGPU >= 24 {
            config["model"] = "SkyReels-V2-DF-1.3B-540P"
            config["precision"] = "fp16"
        } else {
            config["model"] = "SkyReels-V2-DF-1.3B-540P"
            config["precision"] = "fp8"
            config["offload"] = true
        }
    }

    return config
}
