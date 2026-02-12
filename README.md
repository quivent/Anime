# anime

A toolkit for provisioning and managing AI workloads on Lambda Labs GH200 GPU instances. Includes a Go CLI for package installation and model management, and a Tauri desktop app for server monitoring.

> **Status:** Work in progress. The installer scripts and package definitions are complete. The CLI and desktop app are partially implemented.

## Components

### anime-cli

Go CLI built with [Cobra](https://github.com/spf13/cobra) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**What works:**
- 30+ installable packages with dependency resolution
- Embedded bash install scripts for each package
- Model catalog browser (interactive TUI and CLI modes)
- Local and remote model file scanning

**Commands:**
```bash
anime models              # Interactive TUI model browser
anime models --catalog    # Print full model catalog
anime models --local      # Scan local filesystem for model files
anime models list         # List all installable models
anime install <package>   # Install a package (resolves dependencies)
anime packages            # Show available packages
```

### anime-desktop

[Tauri 2.0](https://tauri.app/) desktop app with a Rust backend and React/TypeScript frontend.

**Backend (Rust):**
- Lambda Labs API client
- SSH connection management
- Real-time server monitoring

**Frontend (React):**
- Lambda instance dashboard
- Server monitoring view

## Installable Packages

### Infrastructure

| Package | Description | Size |
|---------|-------------|------|
| `core` | Build tools, git, curl, Python 3 | ~500MB |
| `nvidia` | NVIDIA drivers + CUDA 12.4 | ~4GB |
| `docker` | Docker container platform | ~500MB |
| `python` | Python 3.11+, numpy, scipy, pandas | ~500MB |
| `pytorch` | PyTorch, transformers, diffusers | ~8GB |
| `ollama` | Ollama LLM server with systemd | ~200MB |
| `nodejs` | Node.js 20.x LTS | ~100MB |
| `claude` | Anthropic Claude Code CLI | ~100MB |
| `comfyui` | ComfyUI with Manager | ~5GB |

### LLM Models (via Ollama)

| Package | Model | Size |
|---------|-------|------|
| `llama-3.3-70b` | Llama 3.3 70B | ~40GB |
| `llama-3.3-8b` | Llama 3.3 8B | ~5GB |
| `mistral` | Mistral 7B | ~4GB |
| `mixtral` | Mixtral 8x7B | ~26GB |
| `qwen-2.5-72b` | Qwen 2.5 72B | ~42GB |
| `qwen-2.5-14b` | Qwen 2.5 14B | ~8GB |
| `qwen-2.5-7b` | Qwen 2.5 7B | ~4GB |
| `deepseek-coder-33b` | DeepSeek Coder 33B | ~18GB |
| `deepseek-v3` | DeepSeek V3 (671B MoE) | ~250GB |
| `phi-3.5` | Phi-3.5 Mini 3.8B | ~2GB |

Model bundles are also available: `models-small`, `models-medium`, `models-large`.

### Image Generation (for ComfyUI)

| Package | Model | Size |
|---------|-------|------|
| `sdxl` | Stable Diffusion XL | ~7GB |
| `sd15` | Stable Diffusion 1.5 | ~4GB |
| `flux-dev` | Flux.1 Dev | ~12GB |
| `flux-schnell` | Flux.1 Schnell | ~12GB |

### Video Generation

| Package | Model | Size |
|---------|-------|------|
| `mochi` | Mochi-1 (10B) | ~12GB |
| `svd` | Stable Video Diffusion | ~8GB |
| `animatediff` | AnimateDiff | ~4GB |
| `cogvideo` | CogVideoX-5B | ~14GB |
| `opensora` | Open-Sora 2.0 | ~16GB |
| `ltxvideo` | LTXVideo | ~7GB |
| `wan2` | Wan2.2 | ~10GB |
| `comfyui-wan2` | Wan2 ComfyUI wrapper | ~100MB |

## Project Structure

```
anime/
├── anime-cli/
│   ├── cmd/
│   │   └── models.go           # Model browser + install commands
│   └── internal/
│       └── installer/
│           ├── packages.go     # Package definitions + dependency resolution
│           └── scripts.go      # Embedded bash install scripts
├── anime-desktop/
│   ├── src/                    # React frontend
│   │   ├── App.tsx
│   │   └── components/
│   └── src-tauri/              # Rust backend
│       └── src/
│           ├── lambda/         # Lambda Labs API client
│           └── server/         # SSH + server monitoring
├── .gitignore
└── README.md
```

## Dependencies

**CLI (Go):**
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/spf13/cobra` — CLI framework
- `golang.org/x/crypto/ssh` — SSH client

**Desktop (Rust/TypeScript):**
- Tauri 2.0, reqwest, ssh2, rusqlite
- React, TypeScript

## License

MIT
