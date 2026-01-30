package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// Config represents the main configuration
type Config struct {
    Hardware    HardwareConfig    `json:"hardware"`
    Model       ModelConfig       `json:"model"`
    Parallelism ParallelismConfig `json:"parallelism"`
    Generation  GenerationConfig  `json:"generation"`
    Diffusion   DiffusionConfig   `json:"diffusion"`
    Optimization OptimizationConfig `json:"optimization"`
    Paths       PathsConfig       `json:"paths"`
}

// HardwareConfig represents hardware configuration
type HardwareConfig struct {
    GPUCount       int     `json:"gpu_count"`
    GPUModel       string  `json:"gpu_model"`
    GPUMemoryGB    int     `json:"gpu_memory_gb"`
    NVLinkEnabled  bool    `json:"nvlink_enabled"`
    NVLinkBandwidth int    `json:"nvlink_bandwidth_gbs"`
    SystemRAMGB    int     `json:"system_ram_gb"`
    StorageGB      int     `json:"storage_gb"`
}

// ModelConfig represents model configuration
type ModelConfig struct {
    Variant       string `json:"variant"`
    Parameters    string `json:"parameters"`
    Precision     string `json:"precision"`
    TextEncoder   string `json:"text_encoder"`
    Attention     string `json:"attention"`
}

// ParallelismConfig represents parallelism settings
type ParallelismConfig struct {
    Strategy        string `json:"strategy"`
    ContextParallel int    `json:"context_parallel"`
    CFGParallel     int    `json:"cfg_parallel"`
    VAEParallel     bool   `json:"vae_parallel"`
}

// GenerationConfig represents generation settings
type GenerationConfig struct {
    Resolution     string `json:"resolution"`
    Width          int    `json:"width"`
    Height         int    `json:"height"`
    FPS            int    `json:"fps"`
    MaxFrames      int    `json:"max_frames"`
    BaseNumFrames  int    `json:"base_num_frames"`
}

// DiffusionConfig represents diffusion forcing settings
type DiffusionConfig struct {
    ARStep         int     `json:"ar_step"`
    NumInferSteps  int     `json:"num_inference_steps"`
    GuidanceScale  float64 `json:"guidance_scale"`
    ShiftScale     float64 `json:"shift_scale"`
}

// OptimizationConfig represents optimization settings
type OptimizationConfig struct {
    TeaCacheEnabled   bool    `json:"teacache_enabled"`
    TeaCacheThresh    float64 `json:"teacache_thresh"`
    CompileModel      bool    `json:"compile_model"`
    FP8Quantization   bool    `json:"fp8_quantization"`
    OffloadEnabled    bool    `json:"offload_enabled"`
}

// PathsConfig represents file paths
type PathsConfig struct {
    ModelPath      string `json:"model_path"`
    OutputPath     string `json:"output_path"`
    CachePath      string `json:"cache_path"`
    ConfigPath     string `json:"config_path"`
}

// DefaultConfig returns the default configuration for 4xH100
func DefaultConfig() *Config {
    return &Config{
        Hardware: HardwareConfig{
            GPUCount:        4,
            GPUModel:        "H100-SXM5",
            GPUMemoryGB:     80,
            NVLinkEnabled:   true,
            NVLinkBandwidth: 900,
            SystemRAMGB:     512,
            StorageGB:       2000,
        },
        Model: ModelConfig{
            Variant:     "SkyReels-V2-DF-14B",
            Parameters:  "14B",
            Precision:   "fp8",
            TextEncoder: "t5-xxl",
            Attention:   "flash_attention_2",
        },
        Parallelism: ParallelismConfig{
            Strategy:        "xdit_usp",
            ContextParallel: 4,
            CFGParallel:     1,
            VAEParallel:     true,
        },
        Generation: GenerationConfig{
            Resolution:    "540p",
            Width:         960,
            Height:        544,
            FPS:           24,
            MaxFrames:     289,
            BaseNumFrames: 97,
        },
        Diffusion: DiffusionConfig{
            ARStep:        5,
            NumInferSteps: 30,
            GuidanceScale: 6.0,
            ShiftScale:    8.0,
        },
        Optimization: OptimizationConfig{
            TeaCacheEnabled: true,
            TeaCacheThresh:  0.3,
            CompileModel:    true,
            FP8Quantization: true,
            OffloadEnabled:  false,
        },
        Paths: PathsConfig{
            ModelPath:  "/models/skyreels",
            OutputPath: "/output",
            CachePath:  "/cache",
            ConfigPath: "~/.sky/config.json",
        },
    }
}

// ConfigPath returns the default config file path
func ConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".sky", "config.json")
}

// Load loads configuration from file
func Load() (*Config, error) {
    path := ConfigPath()

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // Return default config if file doesn't exist
            return DefaultConfig(), nil
        }
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
    path := ConfigPath()

    // Ensure directory exists
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}

// GetMemoryBudget calculates memory budget per GPU
func (c *Config) GetMemoryBudget() MemoryBudget {
    totalMem := c.Hardware.GPUMemoryGB

    var modelMem int
    switch c.Model.Precision {
    case "fp8":
        modelMem = 14 // 14B model in FP8
    case "fp16", "bf16":
        modelMem = 28
    default:
        modelMem = 56 // fp32
    }

    textEncoderMem := 6 // T5-XXL
    vaeMem := 4
    contextMem := 15 // Per-GPU context partition

    used := modelMem + textEncoderMem + vaeMem + contextMem
    headroom := totalMem - used

    return MemoryBudget{
        TotalGB:        totalMem,
        ModelGB:        modelMem,
        TextEncoderGB:  textEncoderMem,
        VAEGB:          vaeMem,
        ContextGB:      contextMem,
        HeadroomGB:     headroom,
    }
}

// MemoryBudget represents memory allocation per GPU
type MemoryBudget struct {
    TotalGB       int
    ModelGB       int
    TextEncoderGB int
    VAEGB         int
    ContextGB     int
    HeadroomGB    int
}
