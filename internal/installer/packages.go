package installer

import "time"

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
			Description:  "Essential tools, NVIDIA drivers, CUDA 12.4, Node.js, Docker",
			Dependencies: []string{},
			EstimatedTime: 15 * time.Minute,
			Category:     "Foundation",
			Size:         "~4GB",
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
			Dependencies: []string{"core", "python"},
			EstimatedTime: 10 * time.Minute,
			Category:     "ML Framework",
			Size:         "~8GB",
		},
		"ollama": {
			ID:           "ollama",
			Name:         "Ollama LLM Server",
			Description:  "Ollama service with systemd integration",
			Dependencies: []string{"core"},
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
			Dependencies: []string{"core", "python", "pytorch"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Application",
			Size:         "~5GB",
		},
		"claude": {
			ID:           "claude",
			Name:         "Claude Code",
			Description:  "Official Anthropic Claude CLI",
			Dependencies: []string{"core"},
			EstimatedTime: 2 * time.Minute,
			Category:     "Application",
			Size:         "~100MB",
		},

		// Video Generation Models
		"mochi": {
			ID:           "mochi",
			Name:         "Mochi-1",
			Description:  "Open source video generation model, 10B params",
			Dependencies: []string{"core", "python", "pytorch"},
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
			Dependencies: []string{"core", "python", "pytorch"},
			EstimatedTime: 25 * time.Minute,
			Category:     "Video Generation",
			Size:         "~14GB",
		},
		"opensora": {
			ID:           "opensora",
			Name:         "Open-Sora 2.0",
			Description:  "High-quality video generation model",
			Dependencies: []string{"core", "python", "pytorch"},
			EstimatedTime: 30 * time.Minute,
			Category:     "Video Generation",
			Size:         "~16GB",
		},
		"ltxvideo": {
			ID:           "ltxvideo",
			Name:         "LTXVideo",
			Description:  "Fast video generation with latent transformers",
			Dependencies: []string{"core", "python", "pytorch"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Video Generation",
			Size:         "~7GB",
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
		if seen[id] {
			return nil
		}

		pkg, exists := packages[id]
		if !exists {
			return &PackageNotFoundError{ID: id}
		}

		// Resolve dependencies first
		for _, depID := range pkg.Dependencies {
			if err := resolve(depID); err != nil {
				return err
			}
		}

		seen[id] = true
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
	script, exists := Scripts[packageID]
	return script, exists
}

// PackageNotFoundError is returned when a package doesn't exist
type PackageNotFoundError struct {
	ID string
}

func (e *PackageNotFoundError) Error() string {
	return "package not found: " + e.ID
}
