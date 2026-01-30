package installer

import (
	"fmt"
	"path/filepath"
)

// ModulePaths defines installation paths for each module
type ModulePaths struct {
	// Binaries installed to system PATH
	Binaries []string
	// Python packages installed via pip
	PythonPackages []string
	// Directories created during installation
	Directories []string
	// Config files created/modified
	ConfigFiles []string
	// Systemd services installed
	SystemdServices []string
}

// GetModulePaths returns expected installation paths for a module
func GetModulePaths(modID string) *ModulePaths {
	paths := &ModulePaths{}

	switch modID {
	case "core":
		paths.Binaries = []string{
			"/usr/bin/git",
			"/usr/bin/curl",
			"/usr/bin/wget",
			"/usr/bin/aria2c",
			"/usr/bin/vim",
			"/usr/bin/htop",
			"/usr/bin/tmux",
			"/usr/bin/cmake",
			"/usr/bin/pkg-config",
			"/usr/bin/gcc",
			"/usr/bin/g++",
			"/usr/bin/make",
		}
		paths.Directories = []string{
			"/usr/include/openssl",
			"/usr/lib/x86_64-linux-gnu/libssl.so",
		}

	case "python":
		paths.PythonPackages = []string{
			"numpy",
			"scipy",
			"pandas",
			"matplotlib",
			"pillow",
		}

	case "pytorch":
		paths.PythonPackages = []string{
			"torch",
			"torchvision",
			"torchaudio",
			"transformers",
			"diffusers",
			"accelerate",
			"safetensors",
			"xformers",
			"bitsandbytes",
			"opencv-python",
		}

	case "nvidia":
		paths.Binaries = []string{
			"/usr/bin/nvidia-smi",
			"/usr/bin/nvcc",
		}
		paths.Directories = []string{
			"/usr/local/cuda",
			"/usr/lib/nvidia-*",
		}
		paths.ConfigFiles = []string{
			"/etc/modprobe.d/blacklist-nouveau.conf",
		}

	case "docker":
		paths.Binaries = []string{
			"/usr/bin/docker",
			"/usr/bin/docker-compose",
		}
		paths.SystemdServices = []string{
			"docker.service",
		}
		paths.ConfigFiles = []string{
			"/etc/docker/daemon.json",
		}

	case "ollama":
		paths.Binaries = []string{
			"/usr/local/bin/ollama",
		}
		paths.SystemdServices = []string{
			"ollama.service",
		}
		paths.Directories = []string{
			"/usr/share/ollama",
			"/var/lib/ollama",
		}

	case "vllm":
		paths.PythonPackages = []string{
			"vllm",
		}

	case "flash-attn":
		paths.PythonPackages = []string{
			"flash-attn",
		}

	case "comfyui":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI"),
			filepath.Join(homeDir, "ComfyUI/models"),
			filepath.Join(homeDir, "ComfyUI/custom_nodes"),
			filepath.Join(homeDir, "ComfyUI/output"),
			filepath.Join(homeDir, "ComfyUI/input"),
		}
		paths.Binaries = []string{
			"/usr/local/bin/comfy",
		}

	case "nodejs":
		paths.Binaries = []string{
			"/usr/bin/node",
			"/usr/bin/npm",
			"/usr/bin/npx",
		}

	case "go":
		paths.Binaries = []string{
			"/usr/local/go/bin/go",
		}
		paths.Directories = []string{
			"/usr/local/go",
		}

	case "claude":
		paths.Binaries = []string{
			"/usr/local/bin/claude",
		}

	case "gh":
		paths.Binaries = []string{
			"/usr/bin/gh",
		}

	case "make":
		paths.Binaries = []string{
			"/usr/bin/make",
			"/usr/bin/autoconf",
			"/usr/bin/automake",
			"/usr/bin/libtool",
		}

	case "comfy-cli":
		paths.Binaries = []string{
			"/usr/local/bin/comfy",
		}
		paths.PythonPackages = []string{
			"comfy-cli",
		}

	// LLM models (via Ollama)
	case "llama-3.3-70b", "llama-3.3-8b", "llama-3.2-1b", "llama-3.2-3b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "mistral", "mistral-7b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "mixtral", "mixtral-8x7b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "qwen3-235b", "qwen3-32b", "qwen3-30b", "qwen3-14b", "qwen3-8b", "qwen3-4b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "qwen3-30b-a3b", "qwen3-coder-30b-a3b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "qwq-32b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "deepseek-coder-33b", "deepseek-v3", "deepseek-r1-8b", "deepseek-r1-70b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "deepseek-r1-1.5b", "deepseek-r1-7b", "deepseek-r1-14b", "deepseek-r1-32b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "deepseek-r1-671b", "deepseek-v3-671b", "qwen3-235b-a22b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "llama4-maverick", "llama4-scout":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "phi-3.5", "phi-4":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "gemma3-4b", "gemma3-12b", "gemma3-27b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	case "command-r-7b":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	// Image generation models (via ComfyUI)
	case "sdxl", "sd15", "flux-dev", "flux-schnell", "flux2":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
			filepath.Join(homeDir, "ComfyUI/models/clip"),
			filepath.Join(homeDir, "ComfyUI/models/vae"),
		}

	case "sd3.5-large", "sd3.5-large-turbo", "sd3.5-medium":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
		}

	case "sdxl-turbo", "sdxl-lightning":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
		}

	case "playground-v2.5", "pixart-sigma", "kandinsky-3", "kolors":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
		}

	// Video generation models
	case "svd", "svd-xt", "animatediff", "cogvideo", "wan2", "ltxvideo":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
			filepath.Join(homeDir, "ComfyUI/models/vae"),
		}

	case "mochi", "cogvideox-1.5", "cogvideox-i2v", "hunyuan-video", "pyramid-flow":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/diffusion_models"),
		}

	case "opensora":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "Open-Sora"),
		}

	case "i2v-adapter":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/controlnet"),
		}

	case "comfyui-wan2":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/custom_nodes/ComfyUI-Wan2"),
		}

	// Image enhancement/upscaling
	case "real-esrgan", "gfpgan", "aurasr", "supir":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/upscale_models"),
		}

	// Video enhancement
	case "rife", "film":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/vfi"),
		}

	// Inpainting models
	case "sd-inpainting", "sdxl-inpainting":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/checkpoints"),
		}

	// ControlNet & Adapters
	case "controlnet-canny", "controlnet-depth", "controlnet-openpose":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/controlnet"),
		}

	case "ip-adapter", "ip-adapter-faceid", "instantid":
		homeDir := "/root"
		paths.Directories = []string{
			filepath.Join(homeDir, "ComfyUI/models/ipadapter"),
		}

	// Model bundles
	case "models-small", "models-medium", "models-large":
		paths.Directories = []string{
			"/var/lib/ollama/models",
		}

	default:
		// Unknown module - return empty paths
		return paths
	}

	return paths
}

// GetAllPaths returns all paths that would be affected by a module
func (mp *ModulePaths) GetAllPaths() []string {
	var allPaths []string
	allPaths = append(allPaths, mp.Binaries...)
	allPaths = append(allPaths, mp.Directories...)
	allPaths = append(allPaths, mp.ConfigFiles...)
	return allPaths
}

// FormatPathsList returns a formatted string listing all paths
func (mp *ModulePaths) FormatPathsList() string {
	result := ""
	if len(mp.Binaries) > 0 {
		result += fmt.Sprintf("  Binaries (%d):\n", len(mp.Binaries))
		for _, p := range mp.Binaries {
			result += fmt.Sprintf("    - %s\n", p)
		}
	}
	if len(mp.PythonPackages) > 0 {
		result += fmt.Sprintf("  Python Packages (%d):\n", len(mp.PythonPackages))
		for _, p := range mp.PythonPackages {
			result += fmt.Sprintf("    - %s\n", p)
		}
	}
	if len(mp.Directories) > 0 {
		result += fmt.Sprintf("  Directories (%d):\n", len(mp.Directories))
		for _, p := range mp.Directories {
			result += fmt.Sprintf("    - %s\n", p)
		}
	}
	if len(mp.ConfigFiles) > 0 {
		result += fmt.Sprintf("  Config Files (%d):\n", len(mp.ConfigFiles))
		for _, p := range mp.ConfigFiles {
			result += fmt.Sprintf("    - %s\n", p)
		}
	}
	if len(mp.SystemdServices) > 0 {
		result += fmt.Sprintf("  Systemd Services (%d):\n", len(mp.SystemdServices))
		for _, p := range mp.SystemdServices {
			result += fmt.Sprintf("    - %s\n", p)
		}
	}
	return result
}
