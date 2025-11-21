package installer

import (
	"strings"
	"time"
)

// Package represents an installable component
type Package struct {
	ID           string
	Name         string
	Description  string
	Dependencies []string
	EstimatedTime time.Duration
	Category     string
	Size         string
}

// GetPackages returns all available packages
func GetPackages() map[string]*Package {
	return map[string]*Package{
		"core": {
			ID:           "core",
			Name:         "Core System",
			Description:  "Essential build tools: gcc, git, curl, wget, python3, pkg-config",
			Dependencies: []string{},
			EstimatedTime: 3 * time.Minute,
			Category:     "Foundation",
			Size:         "~500MB",
		},
		"nvidia": {
			ID:           "nvidia",
			Name:         "NVIDIA Drivers & CUDA",
			Description:  "NVIDIA GPU drivers and CUDA 12.4 toolkit for GPU acceleration",
			Dependencies: []string{},
			EstimatedTime: 15 * time.Minute,
			Category:     "GPU",
			Size:         "~4GB",
		},
		"docker": {
			ID:           "docker",
			Name:         "Docker",
			Description:  "Docker container platform for running containerized applications",
			Dependencies: []string{},
			EstimatedTime: 5 * time.Minute,
			Category:     "Containers",
			Size:         "~500MB",
		},
		"python": {
			ID:           "python",
			Name:         "Python & AI Libs",
			Description:  "Python 3.11+, pip, venv, numpy, scipy, pandas",
			Dependencies: []string{"core"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Foundation",
			Size:         "~500MB",
		},
		"pytorch": {
			ID:           "pytorch",
			Name:         "PyTorch Stack",
			Description:  "PyTorch, torchvision, transformers, diffusers, accelerate",
			Dependencies: []string{"core", "python", "nvidia"},
			EstimatedTime: 10 * time.Minute,
			Category:     "ML Framework",
			Size:         "~8GB",
		},
		"ollama": {
			ID:           "ollama",
			Name:         "Ollama LLM Server",
			Description:  "Ollama LLM server with systemd integration (CPU or GPU)",
			Dependencies: []string{},
			EstimatedTime: 3 * time.Minute,
			Category:     "LLM Runtime",
			Size:         "~200MB",
		},
		"models-small": {
			ID:           "models-small",
			Name:         "Small Models (7-8B)",
			Description:  "Mistral, Llama 3.3 8B, Qwen 2.5 7B",
			Dependencies: []string{"ollama"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Models",
			Size:         "~15GB",
		},
		"models-medium": {
			ID:           "models-medium",
			Name:         "Medium Models (14-34B)",
			Description:  "Qwen 2.5 14B, Mixtral 8x7B, DeepSeek Coder 33B",
			Dependencies: []string{"ollama"},
			EstimatedTime: 45 * time.Minute,
			Category:     "Models",
			Size:         "~40GB",
		},
		"models-large": {
			ID:           "models-large",
			Name:         "Large Models (70B+)",
			Description:  "Llama 3.3 70B, Qwen 2.5 72B, DeepSeek V3",
			Dependencies: []string{"ollama"},
			EstimatedTime: 90 * time.Minute,
			Category:     "Models",
			Size:         "~80GB",
		},
		"comfyui": {
			ID:           "comfyui",
			Name:         "ComfyUI",
			Description:  "ComfyUI with manager and custom nodes",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Application",
			Size:         "~5GB",
		},
		"nodejs": {
			ID:           "nodejs",
			Name:         "Node.js & npm",
			Description:  "Node.js runtime, npm, and common JS tools",
			Dependencies: []string{},
			EstimatedTime: 3 * time.Minute,
			Category:     "Runtime",
			Size:         "~100MB",
		},
		"claude": {
			ID:           "claude",
			Name:         "Claude Code",
			Description:  "Official Anthropic Claude CLI",
			Dependencies: []string{"nodejs"},
			EstimatedTime: 2 * time.Minute,
			Category:     "Application",
			Size:         "~100MB",
		},

		// Video Generation Models
		"mochi": {
			ID:           "mochi",
			Name:         "Mochi-1",
			Description:  "Open source video generation model, 10B params",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Video Generation",
			Size:         "~12GB",
		},
		"svd": {
			ID:           "svd",
			Name:         "Stable Video Diffusion",
			Description:  "Stability AI's video diffusion model for ComfyUI",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Video Generation",
			Size:         "~8GB",
		},
		"animatediff": {
			ID:           "animatediff",
			Name:         "AnimateDiff",
			Description:  "Motion module for Stable Diffusion, animates images",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Video Generation",
			Size:         "~4GB",
		},
		"cogvideo": {
			ID:           "cogvideo",
			Name:         "CogVideoX-5B",
			Description:  "Open source text-to-video model",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 25 * time.Minute,
			Category:     "Video Generation",
			Size:         "~14GB",
		},
		"opensora": {
			ID:           "opensora",
			Name:         "Open-Sora 2.0",
			Description:  "High-quality video generation model",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 30 * time.Minute,
			Category:     "Video Generation",
			Size:         "~16GB",
		},
		"ltxvideo": {
			ID:           "ltxvideo",
			Name:         "LTXVideo",
			Description:  "Fast video generation with latent transformers",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Video Generation",
			Size:         "~7GB",
		},
		"wan2": {
			ID:           "wan2",
			Name:         "Wan2.2",
			Description:  "State-of-the-art image-to-video generation model",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Video Generation",
			Size:         "~10GB",
		},
		"comfyui-wan2": {
			ID:           "comfyui-wan2",
			Name:         "ComfyUI Wan2 Wrapper",
			Description:  "Wan2.2 custom node for ComfyUI workflows",
			Dependencies: []string{"comfyui", "wan2"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ComfyUI Node",
			Size:         "~100MB",
		},

		// Individual LLM Models (via Ollama)
		"llama-3.3-70b": {
			ID:           "llama-3.3-70b",
			Name:         "Llama 3.3 70B",
			Description:  "Meta's latest open-source flagship model with exceptional reasoning and coding capabilities",
			Dependencies: []string{"ollama"},
			EstimatedTime: 30 * time.Minute,
			Category:     "LLM",
			Size:         "~40GB",
		},
		"llama-3.3-8b": {
			ID:           "llama-3.3-8b",
			Name:         "Llama 3.3 8B",
			Description:  "Efficient smaller version of Llama 3.3, great balance of performance and speed",
			Dependencies: []string{"ollama"},
			EstimatedTime: 5 * time.Minute,
			Category:     "LLM",
			Size:         "~5GB",
		},
		"mistral": {
			ID:           "mistral",
			Name:         "Mistral 7B",
			Description:  "High-performance 7B model outperforming many larger models, excellent for coding",
			Dependencies: []string{"ollama"},
			EstimatedTime: 4 * time.Minute,
			Category:     "LLM",
			Size:         "~4GB",
		},
		"mixtral": {
			ID:           "mixtral",
			Name:         "Mixtral 8x7B",
			Description:  "Mixture of Experts model with 47B parameters, runs efficiently via sparse activation",
			Dependencies: []string{"ollama"},
			EstimatedTime: 20 * time.Minute,
			Category:     "LLM",
			Size:         "~26GB",
		},
		"qwen-2.5-72b": {
			ID:           "qwen-2.5-72b",
			Name:         "Qwen 2.5 72B",
			Description:  "Alibaba's top model with strong multilingual and math capabilities",
			Dependencies: []string{"ollama"},
			EstimatedTime: 35 * time.Minute,
			Category:     "LLM",
			Size:         "~42GB",
		},
		"qwen-2.5-14b": {
			ID:           "qwen-2.5-14b",
			Name:         "Qwen 2.5 14B",
			Description:  "Mid-size Qwen model with excellent Chinese-English bilingual performance",
			Dependencies: []string{"ollama"},
			EstimatedTime: 8 * time.Minute,
			Category:     "LLM",
			Size:         "~8GB",
		},
		"qwen-2.5-7b": {
			ID:           "qwen-2.5-7b",
			Name:         "Qwen 2.5 7B",
			Description:  "Compact Qwen model with strong multilingual support",
			Dependencies: []string{"ollama"},
			EstimatedTime: 5 * time.Minute,
			Category:     "LLM",
			Size:         "~4GB",
		},
		"deepseek-coder-33b": {
			ID:           "deepseek-coder-33b",
			Name:         "DeepSeek Coder 33B",
			Description:  "Specialized coding model trained on 2T+ tokens of code and text",
			Dependencies: []string{"ollama"},
			EstimatedTime: 15 * time.Minute,
			Category:     "LLM",
			Size:         "~18GB",
		},
		"deepseek-v3": {
			ID:           "deepseek-v3",
			Name:         "DeepSeek V3",
			Description:  "Latest frontier model with 671B parameters using MoE architecture",
			Dependencies: []string{"ollama"},
			EstimatedTime: 120 * time.Minute,
			Category:     "LLM",
			Size:         "~250GB",
		},
		"phi-3.5": {
			ID:           "phi-3.5",
			Name:         "Phi-3.5 Mini (3.8B)",
			Description:  "Microsoft's compact model with strong reasoning despite small size",
			Dependencies: []string{"ollama"},
			EstimatedTime: 2 * time.Minute,
			Category:     "LLM",
			Size:         "~2GB",
		},

		// Individual Image Generation Models (for ComfyUI)
		"sdxl": {
			ID:           "sdxl",
			Name:         "Stable Diffusion XL",
			Description:  "Latest Stable Diffusion with improved image quality and composition",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},
		"sd15": {
			ID:           "sd15",
			Name:         "Stable Diffusion 1.5",
			Description:  "Widely-used base model with huge ecosystem of fine-tunes and LoRAs",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 6 * time.Minute,
			Category:     "Image Generation",
			Size:         "~4GB",
		},
		"flux-dev": {
			ID:           "flux-dev",
			Name:         "Flux.1 Dev",
			Description:  "Black Forest Labs' new model with exceptional prompt following and quality",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~12GB",
		},
		"flux-schnell": {
			ID:           "flux-schnell",
			Name:         "Flux.1 Schnell",
			Description:  "Fast version of Flux optimized for speed while maintaining quality",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~12GB",
		},
	}
}

// ResolveDependencies returns packages in installation order
func ResolveDependencies(packageIDs []string) ([]*Package, error) {
	packages := GetPackages()
	resolved := make([]*Package, 0)
	seen := make(map[string]bool)

	var resolve func(string) error
	resolve = func(id string) error {
		// Normalize to lowercase for case-insensitive matching
		normalizedID := strings.ToLower(id)

		if seen[normalizedID] {
			return nil
		}

		pkg, exists := packages[normalizedID]
		if !exists {
			return &PackageNotFoundError{ID: id}
		}

		// Resolve dependencies first
		for _, depID := range pkg.Dependencies {
			if err := resolve(depID); err != nil {
				return err
			}
		}

		seen[normalizedID] = true
		resolved = append(resolved, pkg)
		return nil
	}

	for _, id := range packageIDs {
		if err := resolve(id); err != nil {
			return nil, err
		}
	}

	return resolved, nil
}

// GetScript returns the installation script for a package
func GetScript(packageID string) (string, bool) {
	// Normalize to lowercase for case-insensitive matching
	normalizedID := strings.ToLower(packageID)
	script, exists := Scripts[normalizedID]
	return script, exists
}

// PackageNotFoundError is returned when a package doesn't exist
type PackageNotFoundError struct {
	ID string
}

func (e *PackageNotFoundError) Error() string {
	return "package not found: " + e.ID
}
