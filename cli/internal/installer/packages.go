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
		"flash-attn": {
			ID:            "flash-attn",
			Name:          "Flash Attention",
			Description:   "Optimized attention implementation for faster transformer inference (compiles from source)",
			Dependencies:  []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 12 * time.Minute,
			Category:      "ML Framework",
			Size:          "~500MB",
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
		"vllm": {
			ID:           "vllm",
			Name:         "vLLM Inference Engine",
			Description:  "High-performance LLM inference with PagedAttention, optimized for throughput",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 8 * time.Minute,
			Category:     "LLM Runtime",
			Size:         "~2GB",
		},
		"models-small": {
			ID:           "models-small",
			Name:         "Small Models (1-8B)",
			Description:  "Llama 3.2 1B/3B, Gemma3 4B, Mistral, Llama 3.3 8B, Qwen3 8B",
			Dependencies: []string{"ollama"},
			EstimatedTime: 25 * time.Minute,
			Category:     "Models",
			Size:         "~22GB",
		},
		"models-medium": {
			ID:           "models-medium",
			Name:         "Medium Models (8-34B)",
			Description:  "DeepSeek-R1 8B, Phi-4, Gemma3 12B, Qwen3 14B/32B, Qwen3-Coder, Mixtral, DeepSeek Coder",
			Dependencies: []string{"ollama"},
			EstimatedTime: 60 * time.Minute,
			Category:     "Models",
			Size:         "~95GB",
		},
		"models-large": {
			ID:           "models-large",
			Name:         "Large Models (27B-235B)",
			Description:  "Gemma3 27B, DeepSeek-R1 70B, Llama 3.3 70B, Qwen3 235B MoE",
			Dependencies: []string{"ollama"},
			EstimatedTime: 90 * time.Minute,
			Category:     "Models",
			Size:         "~220GB",
		},
		"comfy-cli": {
			ID:            "comfy-cli",
			Name:          "ComfyUI CLI",
			Description:   "Official ComfyUI command-line management tool (comfy-cli)",
			Dependencies:  []string{"core", "python"},
			EstimatedTime: 2 * time.Minute,
			Category:      "Application",
			Size:          "~50MB",
		},
		"comfyui": {
			ID:           "comfyui",
			Name:         "ComfyUI",
			Description:  "ComfyUI in isolated venv at ~/ComfyUI/venv with host-aware torch + ComfyUI-Manager",
			Dependencies: []string{"core", "python", "nvidia"},
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
		"go": {
			ID:           "go",
			Name:         "Go",
			Description:  "Go programming language compiler and tools",
			Dependencies: []string{},
			EstimatedTime: 2 * time.Minute,
			Category:     "Runtime",
			Size:         "~150MB",
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
		"gh": {
			ID:            "gh",
			Name:          "GitHub CLI",
			Description:   "Official GitHub CLI for PRs, issues, repos, and GitHub actions",
			Dependencies:  []string{},
			EstimatedTime: 1 * time.Minute,
			Category:      "Application",
			Size:          "~50MB",
		},
		"make": {
			ID:            "make",
			Name:          "Make & Build Tools",
			Description:   "GNU Make, autotools, cmake, and essential build utilities",
			Dependencies:  []string{},
			EstimatedTime: 2 * time.Minute,
			Category:      "Foundation",
			Size:          "~100MB",
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
		"flux2": {
			ID:           "flux2",
			Name:         "Flux 2 (FP8)",
			Description:  "Next-generation video model from Black Forest Labs with superior motion and coherence (FP8 quantized)",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Video Generation",
			Size:         "~8GB",
		},
		"cogvideox-1.5": {
			ID:           "cogvideox-1.5",
			Name:         "CogVideoX 1.5 5B",
			Description:  "Upgraded CogVideoX supporting 10-second videos at higher resolutions",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 30 * time.Minute,
			Category:     "Video Generation",
			Size:         "~18GB",
		},
		"cogvideox-i2v": {
			ID:           "cogvideox-i2v",
			Name:         "CogVideoX 1.5 I2V",
			Description:  "Image-to-video variant of CogVideoX 1.5 with any resolution support",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 30 * time.Minute,
			Category:     "Video Generation",
			Size:         "~18GB",
		},
		"hunyuan-video": {
			ID:           "hunyuan-video",
			Name:         "HunyuanVideo",
			Description:  "Tencent's open-source text-to-video diffusion transformer model",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 35 * time.Minute,
			Category:     "Video Generation",
			Size:         "~20GB",
		},
		"pyramid-flow": {
			ID:           "pyramid-flow",
			Name:         "Pyramid Flow",
			Description:  "Efficient video generation using pyramidal flow matching (768p, up to 10s)",
			Dependencies: []string{"core", "python", "pytorch", "nvidia"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Video Generation",
			Size:         "~12GB",
		},
		"svd-xt": {
			ID:           "svd-xt",
			Name:         "SVD-XT 1.1",
			Description:  "Extended Stable Video Diffusion with improved temporal consistency",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Video Generation",
			Size:         "~10GB",
		},
		"i2v-adapter": {
			ID:           "i2v-adapter",
			Name:         "I2V-Adapter",
			Description:  "General image-to-video adapter for diffusion models (SIGGRAPH 2024)",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Video Generation",
			Size:         "~4GB",
		},

		// Text-to-Image Models
		"sd3.5-large": {
			ID:           "sd3.5-large",
			Name:         "Stable Diffusion 3.5 Large",
			Description:  "8B parameter flagship SD3 model with exceptional quality at 1MP resolution",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 25 * time.Minute,
			Category:     "Image Generation",
			Size:         "~16GB",
		},
		"sd3.5-large-turbo": {
			ID:           "sd3.5-large-turbo",
			Name:         "SD 3.5 Large Turbo",
			Description:  "Distilled SD3.5 Large generating high-quality images in 4 steps",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 25 * time.Minute,
			Category:     "Image Generation",
			Size:         "~16GB",
		},
		"sd3.5-medium": {
			ID:           "sd3.5-medium",
			Name:         "Stable Diffusion 3.5 Medium",
			Description:  "2.6B parameter SD3 with MMDiT-X architecture, consumer GPU friendly",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},
		"sdxl-turbo": {
			ID:           "sdxl-turbo",
			Name:         "SDXL Turbo",
			Description:  "Real-time SDXL generating photorealistic images in a single step",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},
		"sdxl-lightning": {
			ID:           "sdxl-lightning",
			Name:         "SDXL Lightning",
			Description:  "ByteDance's lightning-fast SDXL generating 1024px images in few steps",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},
		"playground-v2.5": {
			ID:           "playground-v2.5",
			Name:         "Playground v2.5",
			Description:  "State-of-the-art aesthetic model outperforming SDXL and DALL-E 3",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},
		"pixart-sigma": {
			ID:           "pixart-sigma",
			Name:         "PixArt-Σ",
			Description:  "Efficient DiT-based text-to-image with 4K support and improved text rendering",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~8GB",
		},
		"kandinsky-3": {
			ID:           "kandinsky-3",
			Name:         "Kandinsky 3",
			Description:  "Sber AI's multilingual text-to-image model with strong Russian support",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~8GB",
		},
		"kolors": {
			ID:           "kolors",
			Name:         "Kolors",
			Description:  "KWAI's bilingual Chinese-English text-to-image model",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~8GB",
		},

		// Image Upscaling/Enhancement
		"real-esrgan": {
			ID:           "real-esrgan",
			Name:         "Real-ESRGAN",
			Description:  "Practical 4x image/video upscaling with artifact removal",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Image Enhancement",
			Size:         "~200MB",
		},
		"gfpgan": {
			ID:           "gfpgan",
			Name:         "GFPGAN",
			Description:  "Practical face restoration algorithm for real-world images",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Image Enhancement",
			Size:         "~350MB",
		},
		"aurasr": {
			ID:           "aurasr",
			Name:         "AuraSR",
			Description:  "GigaGAN-based open-source 4x image upscaler from Fal.ai",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Image Enhancement",
			Size:         "~500MB",
		},
		"supir": {
			ID:           "supir",
			Name:         "SUPIR",
			Description:  "Photo-realistic image restoration using SDXL with text-guided enhancement",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 20 * time.Minute,
			Category:     "Image Enhancement",
			Size:         "~12GB",
		},

		// Video Enhancement/Interpolation
		"rife": {
			ID:           "rife",
			Name:         "RIFE",
			Description:  "Real-time intermediate flow estimation for video frame interpolation",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Video Enhancement",
			Size:         "~200MB",
		},
		"film": {
			ID:           "film",
			Name:         "FILM",
			Description:  "Google's frame interpolation model for large motion between frames",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "Video Enhancement",
			Size:         "~400MB",
		},

		// Inpainting Models
		"sd-inpainting": {
			ID:           "sd-inpainting",
			Name:         "SD 1.5 Inpainting",
			Description:  "Stable Diffusion 1.5 fine-tuned for image inpainting and outpainting",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "Image Generation",
			Size:         "~4GB",
		},
		"sdxl-inpainting": {
			ID:           "sdxl-inpainting",
			Name:         "SDXL Inpainting",
			Description:  "SDXL fine-tuned for high-resolution inpainting and outpainting",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 15 * time.Minute,
			Category:     "Image Generation",
			Size:         "~7GB",
		},

		// ControlNet & Adapters
		"controlnet-canny": {
			ID:           "controlnet-canny",
			Name:         "ControlNet Canny",
			Description:  "Edge detection based image conditioning for Stable Diffusion",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ControlNet",
			Size:         "~1.5GB",
		},
		"controlnet-depth": {
			ID:           "controlnet-depth",
			Name:         "ControlNet Depth",
			Description:  "Depth map conditioning for 3D-aware image generation",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ControlNet",
			Size:         "~1.5GB",
		},
		"controlnet-openpose": {
			ID:           "controlnet-openpose",
			Name:         "ControlNet OpenPose",
			Description:  "Human pose estimation conditioning for character generation",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ControlNet",
			Size:         "~1.5GB",
		},
		"ip-adapter": {
			ID:           "ip-adapter",
			Name:         "IP-Adapter",
			Description:  "Tencent's image prompt adapter for style and content transfer",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ControlNet",
			Size:         "~100MB",
		},
		"ip-adapter-faceid": {
			ID:           "ip-adapter-faceid",
			Name:         "IP-Adapter FaceID",
			Description:  "Face-specific IP-Adapter for identity-preserving generation",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:     "ControlNet",
			Size:         "~200MB",
		},
		"instantid": {
			ID:           "instantid",
			Name:         "InstantID",
			Description:  "Zero-shot identity-preserving generation with single reference image",
			Dependencies: []string{"comfyui"},
			EstimatedTime: 10 * time.Minute,
			Category:     "ControlNet",
			Size:         "~2GB",
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

		// =========================================================================
		// GH200 + Wan 2.2 full stack — captures the complete tuned setup
		// =========================================================================
		"wantorch": {
			ID:            "wantorch",
			Name:          "PyTorch cu130 (GH200/ARM64)",
			Description:   "PyTorch nightly cu130 + sage attention + hf_transfer; unlocks comfy_kitchen.cuda backend on GH200/H200",
			Dependencies:  []string{"core", "python", "comfyui"},
			EstimatedTime: 5 * time.Minute,
			Category:      "ML Framework",
			Size:          "~3.5GB",
		},
		"wannodes": {
			ID:            "wannodes",
			Name:          "ComfyUI Wan custom-node stack (Kijai)",
			Description:   "kijai/ComfyUI-WanVideoWrapper + KJNodes + ComfyUI-Manager — Wan-specific nodes with TeaCache, sage support, dual-expert helpers",
			Dependencies:  []string{"comfyui"},
			EstimatedTime: 3 * time.Minute,
			Category:      "ComfyUI Node",
			Size:          "~200MB",
		},
		"wanmodels": {
			ID:            "wanmodels",
			Name:          "Wan 2.2 — full model set",
			Description:   "Wan 2.2 T2V+I2V dual-expert 14B fp8, TI2V-5B fp16, lightx2v 4-step LoRAs, umt5_xxl encoder, wan_2.1 + wan2.2 VAEs",
			Dependencies:  []string{"comfyui"},
			EstimatedTime: 8 * time.Minute,
			Category:      "Video Generation",
			Size:          "~85GB",
		},
		"comfort": {
			ID:            "comfort",
			Name:          "Comfort — Wan T2V Atelier UI",
			Description:   "Polished single-page web studio for Wan 2.2 T2V (clones quivent/comfort, npm ci + build). Launch with: anime wan studio",
			Dependencies:  []string{"nodejs"},
			EstimatedTime: 4 * time.Minute,
			Category:      "Application",
			Size:          "~400MB",
		},
		"wan": {
			ID:            "wan",
			Name:          "GH200 + Wan 2.2 — full setup (meta)",
			Description:   "Complete tuned GH200 video-gen stack: cu130 torch, sage attn, Kijai Wan nodes, full Wan 2.2 model set, Comfort studio UI, no-LoRA max-quality workflow JSON",
			Dependencies:  []string{"comfyui", "wantorch", "wannodes", "wanmodels", "comfort"},
			EstimatedTime: 22 * time.Minute,
			Category:      "Bundle",
			Size:          "~91GB",
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
		"qwen3-235b": {
			ID:           "qwen3-235b",
			Name:         "Qwen3 235B MoE",
			Description:  "Flagship Qwen3 MoE model (235B total, 22B activated) with advanced reasoning and coding",
			Dependencies: []string{"ollama"},
			EstimatedTime: 60 * time.Minute,
			Category:     "LLM",
			Size:         "~142GB",
		},
		"qwen3-32b": {
			ID:           "qwen3-32b",
			Name:         "Qwen3 32B",
			Description:  "Large dense Qwen3 model with strong multilingual and reasoning capabilities",
			Dependencies: []string{"ollama"},
			EstimatedTime: 10 * time.Minute,
			Category:     "LLM",
			Size:         "~20GB",
		},
		"qwen3-30b": {
			ID:           "qwen3-30b",
			Name:         "Qwen3 30B MoE",
			Description:  "MoE model (30B total, 3B activated) with fast inference and strong capabilities",
			Dependencies: []string{"ollama"},
			EstimatedTime: 10 * time.Minute,
			Category:     "LLM",
			Size:         "~19GB",
		},
		"qwen3-14b": {
			ID:           "qwen3-14b",
			Name:         "Qwen3 14B",
			Description:  "Mid-size Qwen3 model with excellent multilingual performance and reasoning",
			Dependencies: []string{"ollama"},
			EstimatedTime: 8 * time.Minute,
			Category:     "LLM",
			Size:         "~9GB",
		},
		"qwen3-8b": {
			ID:           "qwen3-8b",
			Name:         "Qwen3 8B",
			Description:  "Compact yet powerful Qwen3 model with strong multilingual support",
			Dependencies: []string{"ollama"},
			EstimatedTime: 5 * time.Minute,
			Category:     "LLM",
			Size:         "~5GB",
		},
		"qwen3-4b": {
			ID:           "qwen3-4b",
			Name:         "Qwen3 4B",
			Description:  "Efficient small model with 256K context window, great for edge devices",
			Dependencies: []string{"ollama"},
			EstimatedTime: 3 * time.Minute,
			Category:     "LLM",
			Size:         "~2.5GB",
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
		"phi-4": {
			ID:           "phi-4",
			Name:         "Phi-4 (14B)",
			Description:  "Microsoft's 14B reasoning model that rivals much larger models on complex tasks",
			Dependencies: []string{"ollama"},
			EstimatedTime: 8 * time.Minute,
			Category:     "LLM",
			Size:         "~9GB",
		},
		"deepseek-r1-8b": {
			ID:           "deepseek-r1-8b",
			Name:         "DeepSeek-R1 8B",
			Description:  "Latest reasoning model with outstanding performance in math, programming, and logic",
			Dependencies: []string{"ollama"},
			EstimatedTime: 5 * time.Minute,
			Category:     "LLM",
			Size:         "~5GB",
		},
		"deepseek-r1-70b": {
			ID:           "deepseek-r1-70b",
			Name:         "DeepSeek-R1 70B",
			Description:  "Large reasoning model approaching O3/Gemini 2.5 Pro level performance",
			Dependencies: []string{"ollama"},
			EstimatedTime: 30 * time.Minute,
			Category:     "LLM",
			Size:         "~43GB",
		},
		"gemma3-4b": {
			ID:           "gemma3-4b",
			Name:         "Gemma3 4B",
			Description:  "Google's multimodal model with vision capabilities, 128K context, 140+ languages",
			Dependencies: []string{"ollama"},
			EstimatedTime: 3 * time.Minute,
			Category:     "LLM",
			Size:         "~3GB",
		},
		"gemma3-12b": {
			ID:           "gemma3-12b",
			Name:         "Gemma3 12B",
			Description:  "Mid-size multimodal Gemma3 with vision, strong multilingual performance",
			Dependencies: []string{"ollama"},
			EstimatedTime: 7 * time.Minute,
			Category:     "LLM",
			Size:         "~8GB",
		},
		"gemma3-27b": {
			ID:           "gemma3-27b",
			Name:         "Gemma3 27B",
			Description:  "Largest Gemma3 with vision capabilities, runs on single GPU",
			Dependencies: []string{"ollama"},
			EstimatedTime: 12 * time.Minute,
			Category:     "LLM",
			Size:         "~17GB",
		},
		"llama-3.2-1b": {
			ID:           "llama-3.2-1b",
			Name:         "Llama 3.2 1B",
			Description:  "Ultra-compact model for edge devices, personal assistants, low-resource environments",
			Dependencies: []string{"ollama"},
			EstimatedTime: 1 * time.Minute,
			Category:     "LLM",
			Size:         "~1GB",
		},
		"llama-3.2-3b": {
			ID:           "llama-3.2-3b",
			Name:         "Llama 3.2 3B",
			Description:  "Small efficient model for summarization, instructions, tool use, 128K context",
			Dependencies: []string{"ollama"},
			EstimatedTime: 2 * time.Minute,
			Category:     "LLM",
			Size:         "~2GB",
		},
		"qwen3-coder-30b": {
			ID:           "qwen3-coder-30b",
			Name:         "Qwen3-Coder 30B MoE",
			Description:  "Most agentic code model in Qwen series (30B total, 3.3B activated), 256K context",
			Dependencies: []string{"ollama"},
			EstimatedTime: 10 * time.Minute,
			Category:     "LLM",
			Size:         "~19GB",
		},
		"command-r-7b": {
			ID:           "command-r-7b",
			Name:         "Command-R 7B",
			Description:  "Cohere's efficient model optimized for RAG, multilingual, long context",
			Dependencies: []string{"ollama"},
			EstimatedTime: 4 * time.Minute,
			Category:     "LLM",
			Size:         "~4GB",
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
