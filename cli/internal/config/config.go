package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/defaults"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers      []Server          `yaml:"servers"`
	APIKeys      APIKeys           `yaml:"api_keys"`
	Aliases      map[string]string `yaml:"aliases,omitempty"`
	ShellAliases map[string]string `yaml:"shell_aliases,omitempty"`
	Collections  []Collection      `yaml:"collections,omitempty"`
	Users        []User            `yaml:"users,omitempty"`
	ActiveUser   string            `yaml:"active_user,omitempty"`
	Workflows      []WorkflowProfile `yaml:"workflows,omitempty"`
	ActiveWorkflow string            `yaml:"active_workflow,omitempty"`
}

// LLMServerType represents a supported LLM inference server
type LLMServerType string

const (
	ServerOllama    LLMServerType = "ollama"
	ServerVLLM      LLMServerType = "vllm"
	ServerTensorRT  LLMServerType = "tensorrt-llm"
	ServerLlamaCpp  LLMServerType = "llama.cpp"
	ServerExllamaV2 LLMServerType = "exllamav2"
)

// WorkflowProfile defines an LLM serving workflow configuration
type WorkflowProfile struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description,omitempty"`
	Server        LLMServerType     `yaml:"server,omitempty"`
	ServerType    LLMServerType     `yaml:"server_type,omitempty"`
	Model         string            `yaml:"model,omitempty"`
	Port          int               `yaml:"port,omitempty"`
	GPULayers     int               `yaml:"gpu_layers,omitempty"`
	Context       int               `yaml:"context,omitempty"`
	GPUConfig     GPUConfig         `yaml:"gpu_config,omitempty"`
	AutoLoad      bool              `yaml:"auto_load,omitempty"`
	Models        []ModelDeployment `yaml:"models,omitempty"`
	Optimizations Optimizations     `yaml:"optimizations,omitempty"`
	Tags          []string          `yaml:"tags,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	PreCommands   []string          `yaml:"pre_commands,omitempty"`
	PostCommands  []string          `yaml:"post_commands,omitempty"`
}

// GPUConfig holds GPU resource settings for a workflow
type GPUConfig struct {
	TotalGPUs  int    `yaml:"total_gpus,omitempty"`
	GPUType    string `yaml:"gpu_type,omitempty"`
	GPUMemoryGB int   `yaml:"gpu_memory_gb,omitempty"`
}

// ModelDeployment represents a model to be deployed in a workflow
type ModelDeployment struct {
	ID      string `yaml:"id"`
	Enabled bool   `yaml:"enabled,omitempty"`
	GPUs    int    `yaml:"gpus,omitempty"`
}

// Optimizations holds inference optimization flags
type Optimizations struct {
	FlashAttention      bool   `yaml:"flash_attention,omitempty"`
	PagedAttention      bool   `yaml:"paged_attention,omitempty"`
	SpeculativeDecoding bool   `yaml:"speculative_decoding,omitempty"`
	ContinuousBatching  bool   `yaml:"continuous_batching,omitempty"`
	ChunkedPrefill      bool   `yaml:"chunked_prefill,omitempty"`
	PrefixCaching       bool   `yaml:"prefix_caching,omitempty"`
	DraftModel          string `yaml:"draft_model,omitempty"`
}

// AddWorkflow adds a workflow profile to the config
func (c *Config) AddWorkflow(w WorkflowProfile) {
	c.Workflows = append(c.Workflows, w)
}

// DeleteWorkflow removes a workflow by name
func (c *Config) DeleteWorkflow(name string) error {
	for i, w := range c.Workflows {
		if w.Name == name {
			c.Workflows = append(c.Workflows[:i], c.Workflows[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("workflow %q not found", name)
}

// ListWorkflows returns all workflow profiles
func (c *Config) ListWorkflows() []WorkflowProfile {
	return c.Workflows
}

// GetWorkflow returns a workflow by name
func (c *Config) GetWorkflow(name string) (*WorkflowProfile, error) {
	for i := range c.Workflows {
		if c.Workflows[i].Name == name {
			return &c.Workflows[i], nil
		}
	}
	return nil, fmt.Errorf("workflow %q not found", name)
}

// SetActiveWorkflow sets the active workflow by name
func (c *Config) SetActiveWorkflow(name string) error {
	if name == "" {
		c.ActiveWorkflow = ""
		return nil
	}
	for _, w := range c.Workflows {
		if w.Name == name {
			c.ActiveWorkflow = name
			return nil
		}
	}
	return fmt.Errorf("workflow %q not found", name)
}

// GetActiveWorkflow returns the active workflow profile
func (c *Config) GetActiveWorkflow() (*WorkflowProfile, error) {
	if c.ActiveWorkflow == "" {
		return nil, fmt.Errorf("no active workflow")
	}
	return c.GetWorkflow(c.ActiveWorkflow)
}

// CloneWorkflow duplicates a workflow with a new name
func (c *Config) CloneWorkflow(srcName, newName string) error {
	src, err := c.GetWorkflow(srcName)
	if err != nil {
		return err
	}
	clone := *src
	clone.Name = newName
	c.Workflows = append(c.Workflows, clone)
	return nil
}

type Server struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	SSHKey      string `yaml:"ssh_key"`
	CostPerHour float64 `yaml:"cost_per_hour"`
	Modules     []string `yaml:"modules,omitempty"`
}

type APIKeys struct {
	Anthropic   string `yaml:"anthropic,omitempty"`
	OpenAI      string `yaml:"openai,omitempty"`
	HuggingFace string `yaml:"huggingface,omitempty"`
	LambdaLabs  string `yaml:"lambda_labs,omitempty"`
}

type Collection struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Type        string   `yaml:"type"`         // image, video, mixed
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

type User struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"` // Home directory path for the user
}

// ClusterConfig holds GPU cluster requirements for large models
type ClusterConfig struct {
	MinGPUs         int    // Minimum number of GPUs required
	RecommendedGPUs int    // Recommended number of GPUs
	MinVRAMPerGPU   int    // Minimum VRAM per GPU in GB
	TotalVRAM       int    // Total VRAM needed in GB
	Parallelism     string // Recommended parallelism config (e.g., "TP4EP2")
	InferenceEngine string // Recommended engine (vllm, tensorrt-llm, ollama)
	Notes           string // Additional loading/configuration notes
}

type Module struct {
	ID           string
	Name         string
	Description  string
	TimeMinutes  int
	Dependencies []string
	Script       string
	Category     string        // System, LLM, Image, Video, Tools
	Size         string        // Human readable size like "~4GB"
	Cluster      *ClusterConfig // Optional cluster requirements for frontier models
}

var AvailableModules = []Module{
	// ═══════════════════════════════════════════════════════════════════════
	// SYSTEM MODULES
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:          "core",
		Name:        "Core System",
		Description: "CUDA, Python, Node.js, Docker - required base",
		TimeMinutes: 5,
		Script:      "core",
		Category:    "System",
		Size:        "~2GB",
	},
	{
		ID:           "pytorch",
		Name:         "PyTorch + AI Libraries",
		Description:  "PyTorch, Transformers, Diffusers, xformers",
		TimeMinutes:  2,
		Dependencies: []string{"core"},
		Script:       "pytorch",
		Category:     "System",
		Size:         "~8GB",
	},
	{
		ID:           "ollama",
		Name:         "Ollama Server",
		Description:  "Ollama LLM server (required for LLM models)",
		TimeMinutes:  1,
		Dependencies: []string{"core"},
		Script:       "ollama",
		Category:     "System",
		Size:         "~500MB",
	},
	{
		ID:           "vllm",
		Name:         "vLLM Inference Engine",
		Description:  "High-performance LLM inference with PagedAttention",
		TimeMinutes:  8,
		Dependencies: []string{"core", "pytorch"},
		Script:       "vllm",
		Category:     "System",
		Size:         "~2GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// LLM MODELS - FRONTIER CLASS (Best Quality, Multi-GPU Required)
	// Ordered by benchmark performance - highest quality first
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "deepseek-r1-671b",
		Name:         "DeepSeek-R1 671B",
		Description:  "SOTA reasoning, rivals O3/Gemini 2.5 Pro. MoE: 671B total, 37B active",
		TimeMinutes:  90,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-671b",
		Category:     "LLM-Frontier",
		Size:         "~400GB",
		Cluster: &ClusterConfig{
			MinGPUs:         4,
			RecommendedGPUs: 8,
			MinVRAMPerGPU:   80,
			TotalVRAM:       480,
			Parallelism:     "TP4EP2 (4-way tensor, 2-way expert parallelism)",
			InferenceEngine: "vllm or tensorrt-llm",
			Notes:           "Use --max-model-len 32768 for stability. Wide-EP recommended for B200 NVL72. Enable FP8 quantization for 30% memory reduction.",
		},
	},
	{
		ID:           "deepseek-v3-671b",
		Name:         "DeepSeek-V3 671B",
		Description:  "Latest V3 architecture. MoE: 671B total + 14B MTP, 37B active per token",
		TimeMinutes:  90,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-v3-671b",
		Category:     "LLM-Frontier",
		Size:         "~400GB",
		Cluster: &ClusterConfig{
			MinGPUs:         4,
			RecommendedGPUs: 8,
			MinVRAMPerGPU:   80,
			TotalVRAM:       480,
			Parallelism:     "TP4EP2 (4-way tensor, 2-way expert parallelism)",
			InferenceEngine: "vllm or tensorrt-llm",
			Notes:           "MTP head adds 14B params. Use --trust-remote-code. Achieves 368 tok/s with TensorRT-LLM on 8×B200.",
		},
	},
	{
		ID:           "qwen3-235b-a22b",
		Name:         "Qwen3 235B-A22B",
		Description:  "Flagship Qwen3 MoE. 235B total, 22B active. 119 languages, 36T training tokens",
		TimeMinutes:  60,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-235b-a22b",
		Category:     "LLM-Frontier",
		Size:         "~142GB",
		Cluster: &ClusterConfig{
			MinGPUs:         2,
			RecommendedGPUs: 4,
			MinVRAMPerGPU:   80,
			TotalVRAM:       192,
			Parallelism:     "TP2 or TP4 (tensor parallelism only)",
			InferenceEngine: "vllm or ollama",
			Notes:           "Smaller MoE than DeepSeek. Fits on 2×H100/A100. Use BF16 for best quality, INT8 for speed.",
		},
	},
	{
		ID:           "llama4-maverick",
		Name:         "Llama 4 Maverick",
		Description:  "Meta's multimodal MoE. 400B total, 17B active, 128 experts, 512K context",
		TimeMinutes:  75,
		Dependencies: []string{"ollama"},
		Script:       "model-llama4-maverick",
		Category:     "LLM-Frontier",
		Size:         "~240GB",
		Cluster: &ClusterConfig{
			MinGPUs:         4,
			RecommendedGPUs: 8,
			MinVRAMPerGPU:   80,
			TotalVRAM:       320,
			Parallelism:     "TP4EP2 (handles 128 experts efficiently)",
			InferenceEngine: "vllm",
			Notes:           "128 experts benefit from Wide Expert Parallelism on large clusters. Use --max-model-len 65536 for 64K context, scale up for 512K.",
		},
	},
	{
		ID:           "llama4-scout",
		Name:         "Llama 4 Scout",
		Description:  "Meta's efficient MoE. 109B total, 17B active, 10M context! Fits single H100",
		TimeMinutes:  45,
		Dependencies: []string{"ollama"},
		Script:       "model-llama4-scout",
		Category:     "LLM-Frontier",
		Size:         "~65GB",
		Cluster: &ClusterConfig{
			MinGPUs:         1,
			RecommendedGPUs: 2,
			MinVRAMPerGPU:   80,
			TotalVRAM:       80,
			Parallelism:     "TP1 or TP2 (efficient MoE)",
			InferenceEngine: "vllm or ollama",
			Notes:           "Most efficient frontier model! Fits single H100/A100 80GB. 10M context requires chunked attention. Use --enable-chunked-prefill.",
		},
	},

	// ═══════════════════════════════════════════════════════════════════════
	// LLM MODELS - LARGE CLASS (70B+, typically needs 2+ GPUs or quantized)
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "deepseek-r1-70b",
		Name:         "DeepSeek-R1 70B",
		Description:  "Distilled R1, near frontier performance. Strong math/code/logic",
		TimeMinutes:  35,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-70b",
		Category:     "LLM-Large",
		Size:         "~43GB",
	},
	{
		ID:           "llama-3.3-70b",
		Name:         "Llama 3.3 70B",
		Description:  "Meta's flagship dense model. Exceptional reasoning & coding",
		TimeMinutes:  30,
		Dependencies: []string{"ollama"},
		Script:       "model-llama-3.3-70b",
		Category:     "LLM-Large",
		Size:         "~40GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// LLM MODELS - MEDIUM CLASS (14-34B, single GPU friendly)
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "qwq-32b",
		Name:         "QwQ 32B",
		Description:  "Qwen's reasoning specialist. Rivals DeepSeek-R1 & o1-mini at 32B!",
		TimeMinutes:  18,
		Dependencies: []string{"ollama"},
		Script:       "model-qwq-32b",
		Category:     "LLM-Medium",
		Size:         "~20GB",
	},
	{
		ID:           "deepseek-r1-32b",
		Name:         "DeepSeek-R1 32B",
		Description:  "Distilled R1, performs like o1-mini. Excellent reasoning",
		TimeMinutes:  16,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-32b",
		Category:     "LLM-Medium",
		Size:         "~20GB",
	},
	{
		ID:           "qwen3-32b",
		Name:         "Qwen3 32B",
		Description:  "Large dense Qwen3. Strong reasoning, coding, multilingual",
		TimeMinutes:  15,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-32b",
		Category:     "LLM-Medium",
		Size:         "~20GB",
	},
	{
		ID:           "qwen3-30b-a3b",
		Name:         "Qwen3 30B-A3B MoE",
		Description:  "Small MoE, beats QwQ-32B with only 3B active! Ultra efficient",
		TimeMinutes:  14,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-30b-a3b",
		Category:     "LLM-Medium",
		Size:         "~19GB",
	},
	{
		ID:           "qwen3-coder-30b-a3b",
		Name:         "Qwen3-Coder 30B-A3B",
		Description:  "Most agentic code model. MoE 30B/3.3B active, 256K context",
		TimeMinutes:  14,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-coder-30b-a3b",
		Category:     "LLM-Medium",
		Size:         "~19GB",
	},
	{
		ID:           "gemma3-27b",
		Name:         "Gemma3 27B",
		Description:  "Google's largest Gemma3. Vision capable, single GPU",
		TimeMinutes:  14,
		Dependencies: []string{"ollama"},
		Script:       "model-gemma3-27b",
		Category:     "LLM-Medium",
		Size:         "~17GB",
	},
	{
		ID:           "mixtral-8x7b",
		Name:         "Mixtral 8x7B",
		Description:  "Mistral's MoE. 47B total, ~13B active. Multi-task efficient",
		TimeMinutes:  18,
		Dependencies: []string{"ollama"},
		Script:       "model-mixtral-8x7b",
		Category:     "LLM-Medium",
		Size:         "~26GB",
	},
	{
		ID:           "deepseek-coder-33b",
		Name:         "DeepSeek Coder 33B",
		Description:  "Code specialist. 2T+ training tokens of code",
		TimeMinutes:  15,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-coder-33b",
		Category:     "LLM-Medium",
		Size:         "~18GB",
	},
	{
		ID:           "deepseek-r1-14b",
		Name:         "DeepSeek-R1 14B",
		Description:  "Distilled R1 on Qwen2.5. Great reasoning at modest size",
		TimeMinutes:  10,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-14b",
		Category:     "LLM-Medium",
		Size:         "~9GB",
	},
	{
		ID:           "qwen3-14b",
		Name:         "Qwen3 14B",
		Description:  "Dense Qwen3. Excellent bilingual Chinese/English",
		TimeMinutes:  8,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-14b",
		Category:     "LLM-Medium",
		Size:         "~9GB",
	},
	{
		ID:           "phi-4",
		Name:         "Phi-4 14B",
		Description:  "Microsoft's reasoning model. Punches way above weight class",
		TimeMinutes:  6,
		Dependencies: []string{"ollama"},
		Script:       "model-phi-4",
		Category:     "LLM-Medium",
		Size:         "~9GB",
	},
	{
		ID:           "gemma3-12b",
		Name:         "Gemma3 12B",
		Description:  "Google's multimodal. Vision + 140 languages",
		TimeMinutes:  6,
		Dependencies: []string{"ollama"},
		Script:       "model-gemma3-12b",
		Category:     "LLM-Medium",
		Size:         "~8GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// LLM MODELS - SMALL CLASS (≤8B, consumer GPU friendly)
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "deepseek-r1-8b",
		Name:         "DeepSeek-R1 8B",
		Description:  "Distilled R1 on Llama3. Complex reasoning on consumer GPU",
		TimeMinutes:  4,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-8b",
		Category:     "LLM-Small",
		Size:         "~5GB",
	},
	{
		ID:           "llama-3.3-8b",
		Name:         "Llama 3.3 8B",
		Description:  "Meta's efficient workhorse. Great all-rounder",
		TimeMinutes:  4,
		Dependencies: []string{"ollama"},
		Script:       "model-llama-3.3-8b",
		Category:     "LLM-Small",
		Size:         "~5GB",
	},
	{
		ID:           "qwen3-8b",
		Name:         "Qwen3 8B",
		Description:  "Strong multilingual. 119 languages, thinking mode",
		TimeMinutes:  4,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-8b",
		Category:     "LLM-Small",
		Size:         "~5GB",
	},
	{
		ID:           "mistral-7b",
		Name:         "Mistral 7B",
		Description:  "Fast & capable. Excellent coding, technical tasks",
		TimeMinutes:  3,
		Dependencies: []string{"ollama"},
		Script:       "model-mistral-7b",
		Category:     "LLM-Small",
		Size:         "~4GB",
	},
	{
		ID:           "deepseek-r1-7b",
		Name:         "DeepSeek-R1 7B",
		Description:  "Distilled R1 on Qwen2.5. Reasoning at 7B",
		TimeMinutes:  3,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-7b",
		Category:     "LLM-Small",
		Size:         "~4GB",
	},
	{
		ID:           "qwen3-4b",
		Name:         "Qwen3 4B",
		Description:  "Compact powerhouse. 256K context, edge-device friendly",
		TimeMinutes:  2,
		Dependencies: []string{"ollama"},
		Script:       "model-qwen3-4b",
		Category:     "LLM-Small",
		Size:         "~2.5GB",
	},
	{
		ID:           "llama-3.2-3b",
		Name:         "Llama 3.2 3B",
		Description:  "Tiny but capable. 128K context, tool use",
		TimeMinutes:  2,
		Dependencies: []string{"ollama"},
		Script:       "model-llama-3.2-3b",
		Category:     "LLM-Small",
		Size:         "~2GB",
	},
	{
		ID:           "deepseek-r1-1.5b",
		Name:         "DeepSeek-R1 1.5B",
		Description:  "Smallest R1 distill. Reasoning on 4GB VRAM",
		TimeMinutes:  1,
		Dependencies: []string{"ollama"},
		Script:       "model-deepseek-r1-1.5b",
		Category:     "LLM-Small",
		Size:         "~1GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// IMAGE GENERATION MODELS
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "sdxl",
		Name:         "Stable Diffusion XL",
		Description:  "High-quality images, great composition",
		TimeMinutes:  8,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-sdxl",
		Category:     "Image",
		Size:         "~7GB",
	},
	{
		ID:           "sd15",
		Name:         "Stable Diffusion 1.5",
		Description:  "Classic model, huge LoRA ecosystem",
		TimeMinutes:  5,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-sd15",
		Category:     "Image",
		Size:         "~4GB",
	},
	{
		ID:           "flux-dev",
		Name:         "Flux.1 Dev",
		Description:  "Exceptional prompt following & photorealism",
		TimeMinutes:  12,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-flux-dev",
		Category:     "Image",
		Size:         "~12GB",
	},
	{
		ID:           "flux-schnell",
		Name:         "Flux.1 Schnell",
		Description:  "Fast Flux variant, rapid iteration",
		TimeMinutes:  12,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-flux-schnell",
		Category:     "Image",
		Size:         "~12GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// VIDEO GENERATION MODELS
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "svd",
		Name:         "Stable Video Diffusion",
		Description:  "Image-to-video, smooth animations",
		TimeMinutes:  10,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-svd",
		Category:     "Video",
		Size:         "~10GB",
	},
	{
		ID:           "animatediff",
		Name:         "AnimateDiff",
		Description:  "Motion module, animates SD images",
		TimeMinutes:  6,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-animatediff",
		Category:     "Video",
		Size:         "~4GB",
	},
	{
		ID:           "cogvideo",
		Name:         "CogVideoX-5B",
		Description:  "Text-to-video with temporal consistency",
		TimeMinutes:  18,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-cogvideo",
		Category:     "Video",
		Size:         "~14GB",
	},
	{
		ID:           "wan2",
		Name:         "Wan2.2",
		Description:  "State-of-the-art image-to-video quality",
		TimeMinutes:  12,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-wan2",
		Category:     "Video",
		Size:         "~10GB",
	},
	{
		ID:           "ltxvideo",
		Name:         "LTXVideo",
		Description:  "Fast video generation, quick previews",
		TimeMinutes:  8,
		Dependencies: []string{"pytorch", "comfyui"},
		Script:       "model-ltxvideo",
		Category:     "Video",
		Size:         "~7GB",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// TOOLS
	// ═══════════════════════════════════════════════════════════════════════
	{
		ID:           "comfyui",
		Name:         "ComfyUI",
		Description:  "Node-based image/video generation UI",
		TimeMinutes:  2,
		Dependencies: []string{"pytorch"},
		Script:       "comfyui",
		Category:     "Tools",
		Size:         "~500MB",
	},
	{
		ID:           "claude",
		Name:         "Claude Code CLI",
		Description:  "Anthropic Claude Code CLI assistant",
		TimeMinutes:  1,
		Dependencies: []string{"core"},
		Script:       "claude",
		Category:     "Tools",
		Size:         "~100MB",
	},
}

// GetModulesByCategory returns modules grouped by category
func GetModulesByCategory() map[string][]Module {
	categories := make(map[string][]Module)
	for _, mod := range AvailableModules {
		categories[mod.Category] = append(categories[mod.Category], mod)
	}
	return categories
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anime", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Servers:     []Server{},
			APIKeys:     APIKeys{},
			Aliases:     make(map[string]string),
			Collections: []Collection{},
			Users:       []User{},
			ActiveUser:  "",
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize aliases map if nil
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	// Initialize shell aliases map if nil
	if cfg.ShellAliases == nil {
		cfg.ShellAliases = make(map[string]string)
	}

	// Initialize collections slice if nil
	if cfg.Collections == nil {
		cfg.Collections = []Collection{}
	}

	// Initialize users slice if nil
	if cfg.Users == nil {
		cfg.Users = []User{}
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) AddServer(server Server) {
	c.Servers = append(c.Servers, server)
}

func (c *Config) GetServer(name string) (*Server, error) {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i], nil
		}
	}
	return nil, fmt.Errorf("server %s not found", name)
}

func (c *Config) UpdateServer(name string, server Server) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers[i] = server
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

func (c *Config) DeleteServer(name string) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

func EstimateCost(modules []string, costPerHour float64) float64 {
	totalMinutes := 0
	moduleMap := make(map[string]bool)

	// Add dependencies
	var addDeps func(string)
	addDeps = func(id string) {
		if moduleMap[id] {
			return
		}
		moduleMap[id] = true

		for _, mod := range AvailableModules {
			if mod.ID == id {
				for _, dep := range mod.Dependencies {
					addDeps(dep)
				}
				break
			}
		}
	}

	for _, id := range modules {
		addDeps(id)
	}

	// Calculate total time
	for _, mod := range AvailableModules {
		if moduleMap[mod.ID] {
			totalMinutes += mod.TimeMinutes
		}
	}

	return (float64(totalMinutes) / 60.0) * costPerHour
}

func GetModulesByID(ids []string) []Module {
	var result []Module
	for _, id := range ids {
		for _, mod := range AvailableModules {
			if mod.ID == id {
				result = append(result, mod)
				break
			}
		}
	}
	return result
}

// SetAlias adds or updates an alias
func (c *Config) SetAlias(alias, target string) {
	if c.Aliases == nil {
		c.Aliases = make(map[string]string)
	}
	c.Aliases[alias] = target
}

// GetAlias returns the target for an alias, checking runtime config first, then embedded defaults
func (c *Config) GetAlias(alias string) string {
	// First check runtime config
	if c.Aliases != nil {
		if target, ok := c.Aliases[alias]; ok {
			return target
		}
	}
	// Fall back to embedded defaults
	return defaults.GetAlias(alias)
}

// DeleteAlias removes an alias
func (c *Config) DeleteAlias(alias string) error {
	if c.Aliases == nil {
		return fmt.Errorf("alias %s not found", alias)
	}
	if _, exists := c.Aliases[alias]; !exists {
		return fmt.Errorf("alias %s not found", alias)
	}
	delete(c.Aliases, alias)
	return nil
}

// ListAliases returns all aliases (merged: embedded defaults + runtime config)
func (c *Config) ListAliases() map[string]string {
	result := make(map[string]string)

	// Start with embedded defaults
	for k, v := range defaults.ListAliases() {
		result[k] = v
	}

	// Override with runtime config (takes precedence)
	if c.Aliases != nil {
		for k, v := range c.Aliases {
			result[k] = v
		}
	}

	return result
}

// ListEmbeddedAliases returns only the embedded default aliases
func (c *Config) ListEmbeddedAliases() map[string]string {
	return defaults.ListAliases()
}

// IsEmbeddedAlias returns true if the alias is an embedded default
func (c *Config) IsEmbeddedAlias(alias string) bool {
	return defaults.GetAlias(alias) != ""
}

// AddCollection adds a new collection
func (c *Config) AddCollection(collection Collection) error {
	// Check if collection with same name exists
	for _, col := range c.Collections {
		if col.Name == collection.Name {
			return fmt.Errorf("collection %s already exists", collection.Name)
		}
	}
	c.Collections = append(c.Collections, collection)
	return nil
}

// GetCollection returns a collection by name
func (c *Config) GetCollection(name string) (*Collection, error) {
	for i := range c.Collections {
		if c.Collections[i].Name == name {
			return &c.Collections[i], nil
		}
	}
	return nil, fmt.Errorf("collection %s not found", name)
}

// DeleteCollection removes a collection
func (c *Config) DeleteCollection(name string) error {
	for i := range c.Collections {
		if c.Collections[i].Name == name {
			c.Collections = append(c.Collections[:i], c.Collections[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("collection %s not found", name)
}

// ListCollections returns all collections
func (c *Config) ListCollections() []Collection {
	return c.Collections
}

// AddUser adds a new user
func (c *Config) AddUser(user User) error {
	// Check if user with same name exists
	for _, u := range c.Users {
		if u.Name == user.Name {
			return fmt.Errorf("user %s already exists", user.Name)
		}
	}
	c.Users = append(c.Users, user)
	return nil
}

// GetUser returns a user by name
func (c *Config) GetUser(name string) (*User, error) {
	for i := range c.Users {
		if c.Users[i].Name == name {
			return &c.Users[i], nil
		}
	}
	return nil, fmt.Errorf("user %s not found", name)
}

// DeleteUser removes a user
func (c *Config) DeleteUser(name string) error {
	for i := range c.Users {
		if c.Users[i].Name == name {
			c.Users = append(c.Users[:i], c.Users[i+1:]...)
			// Clear active user if it was the deleted user
			if c.ActiveUser == name {
				c.ActiveUser = ""
			}
			return nil
		}
	}
	return fmt.Errorf("user %s not found", name)
}

// ListUsers returns all users
func (c *Config) ListUsers() []User {
	return c.Users
}

// SetActiveUser sets the active user
func (c *Config) SetActiveUser(name string) error {
	// Verify user exists
	_, err := c.GetUser(name)
	if err != nil {
		return err
	}
	c.ActiveUser = name
	return nil
}

// GetActiveUser returns the active user
func (c *Config) GetActiveUser() (*User, error) {
	if c.ActiveUser == "" {
		return nil, fmt.Errorf("no active user set")
	}
	return c.GetUser(c.ActiveUser)
}

// AddShellAlias adds or updates a shell alias
func (c *Config) AddShellAlias(name, command string) error {
	if c.ShellAliases == nil {
		c.ShellAliases = make(map[string]string)
	}
	c.ShellAliases[name] = command
	return nil
}

// RemoveShellAlias removes a shell alias
func (c *Config) RemoveShellAlias(name string) error {
	if c.ShellAliases == nil {
		return fmt.Errorf("shell alias %s not found", name)
	}
	if _, exists := c.ShellAliases[name]; !exists {
		return fmt.Errorf("shell alias %s not found", name)
	}
	delete(c.ShellAliases, name)
	return nil
}

// GetShellAliases returns all shell aliases
func (c *Config) GetShellAliases() map[string]string {
	if c.ShellAliases == nil {
		return make(map[string]string)
	}
	return c.ShellAliases
}
