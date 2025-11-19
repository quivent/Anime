# Lambda GH200 Modular Setup

Fast, modular installation for expensive GPU clusters. Install only what you need!

## Quick Start

### Method 1: Copy/Paste (Fastest)

```bash
# On your local machine - copy the script
cat setup-core.sh | pbcopy

# SSH into Lambda machine
ssh ubuntu@YOUR_LAMBDA_IP

# Paste and run
bash
# (paste here with Cmd+V)
```

### Method 2: SCP Upload

```bash
# Upload all scripts
scp setup-*.sh ubuntu@YOUR_LAMBDA_IP:~/

# SSH in
ssh ubuntu@YOUR_LAMBDA_IP

# Run scripts
bash setup-core.sh
```

### Method 3: Single Command Download

```bash
ssh ubuntu@YOUR_LAMBDA_IP
wget https://your-server.com/setup-core.sh
bash setup-core.sh
```

## Installation Modules

### 1. Core Setup (Required, ~5 min)

```bash
bash setup-core.sh
```

Installs:
- ✓ CUDA 12.4 + NVIDIA drivers
- ✓ Python 3 + pip
- ✓ Node.js 20 + npm
- ✓ Docker
- ✓ Build tools (gcc, cmake, etc.)

**Cost: ~$2-3 on expensive cluster**

### 2. PyTorch + AI Libraries (~2-3 min)

```bash
bash setup-pytorch.sh
```

Installs:
- PyTorch with CUDA 12.4
- Transformers, Diffusers, Accelerate
- xformers, bitsandbytes
- NumPy, Pandas, OpenCV, Pillow

**Cost: ~$1-2**

### 3. Ollama (No Models, ~1 min)

```bash
bash setup-ollama.sh
```

Installs Ollama server only, no model downloads.

**Cost: ~$0.50**

### 4. Models (Interactive, 2-40 min depending on selection)

```bash
bash setup-models.sh
```

**Interactive menu** lets you choose:

**Quick picks:**
- `recommended` - Llama 3.3 70B + Qwen 2.5 7B + Mistral (15-20 min)
- `all-small` - All 7B models (8-10 min) **← BEST for expensive clusters**
- `all-medium` - 14-34B models (20-30 min)
- `all-large` - 70B+ models (40-60 min)

**Or pick specific models:**
```bash
bash setup-models.sh llama3.3:8b mistral:latest  # Just small models, 5 min
bash setup-models.sh qwen2.5:7b                  # Single model, 2-3 min
```

**Models available:**
- `llama3.3:70b` (40GB, ~20 min) - Best quality
- `llama3.3:8b` (4.7GB, ~3 min) - Fast, good quality
- `qwen2.5:72b` (41GB, ~20 min) - Excellent coding
- `qwen2.5:14b` (9GB, ~7 min) - Great coding, smaller
- `qwen2.5:7b` (4.7GB, ~3 min) - Fast, versatile
- `mistral:latest` (4.1GB, ~2 min) - Very fast
- `mixtral:8x7b` (26GB, ~12 min) - MoE, efficient
- `codellama:70b` (39GB, ~20 min) - Code specialist
- `deepseek-coder:33b` (19GB, ~10 min) - Excellent coder

**Cost: $5-20 depending on selection**

### 5. ComfyUI (~2 min)

```bash
bash setup-comfyui.sh
```

Installs ComfyUI + Manager plugin.

**Cost: ~$1**

### 6. Claude Code (~30 sec)

```bash
bash setup-claude.sh
```

**Cost: ~$0.25**

### 7. All Optional (Except Models, ~7 min)

```bash
bash setup-all-optional.sh
```

Installs PyTorch + Ollama + ComfyUI + Claude Code (but no models).

**Cost: ~$3-5**

## Recommended Workflows

### Budget Workflow ($3-5 total, ~10 min)
```bash
bash setup-core.sh              # 5 min
bash setup-pytorch.sh           # 2 min
bash setup-ollama.sh            # 1 min
bash setup-models.sh            # Choose 1-2 small models, 3-5 min
```

### Standard Workflow ($8-12 total, ~20 min)
```bash
bash setup-core.sh              # 5 min
bash setup-all-optional.sh      # 7 min
bash setup-models.sh            # Choose "recommended", 8 min
```

### Full Workflow ($20-30 total, ~40 min)
```bash
bash setup-core.sh              # 5 min
bash setup-all-optional.sh      # 7 min
bash setup-models.sh            # Choose "all-large", 30+ min
```

### Just Testing PyTorch ($2-3 total, ~7 min)
```bash
bash setup-core.sh              # 5 min
bash setup-pytorch.sh           # 2 min
# Test and terminate instance
```

## Verification

```bash
# Quick check
nvidia-smi
python3 -c "import torch; print(torch.cuda.is_available())"
ollama list
docker --version
node --version
```

## Starting Services

```bash
# ComfyUI
cd ~/ComfyUI && python3 main.py --listen 0.0.0.0 --port 8188

# Or use the helper script
./start-comfyui.sh

# Ollama is auto-started as a service
ollama run llama3.3:8b "Hello!"
```

## Time & Cost Estimates

At $20/hour:

| Task | Time | Cost |
|------|------|------|
| Core setup | 5 min | $1.67 |
| PyTorch | 2 min | $0.67 |
| Ollama | 1 min | $0.33 |
| Small model (7B) | 3 min | $1.00 |
| Medium model (34B) | 10 min | $3.33 |
| Large model (70B) | 20 min | $6.67 |
| ComfyUI | 2 min | $0.67 |
| Claude Code | 30 sec | $0.17 |

**Total for minimal setup: ~$3-5 (8 minutes)**
**Total for full setup: ~$25-30 (60 minutes)**

## Tips for Expensive Clusters

1. **Run core setup first**, verify it works, then add components
2. **Skip large models** unless you specifically need them
3. **Use small models** (7B) for testing - they're fast and cheap
4. **Test PyTorch/CUDA** immediately after setup-pytorch.sh
5. **Download models last** - they're the slowest part
6. **Consider downloading models later** from a cheaper instance
7. **Snapshot the instance** after setup to reuse later

## Troubleshooting

### Script fails
```bash
# Check what's installed
ls -la ~/setup-*.sh
# Rerun specific component
bash setup-pytorch.sh
```

### Ollama not starting
```bash
sudo systemctl status ollama
sudo journalctl -u ollama -f
```

### Out of space
```bash
df -h
# Clean up
docker system prune -a
pip cache purge
```

### Need to reboot (NVIDIA drivers)
```bash
sudo reboot
# Wait 1 minute, SSH back in
nvidia-smi
```

## File Reference

- `setup-core.sh` - Base system, CUDA, Python, Node, Docker
- `setup-pytorch.sh` - PyTorch + AI libraries
- `setup-ollama.sh` - Ollama server only
- `setup-models.sh` - Interactive model downloader
- `setup-comfyui.sh` - ComfyUI installation
- `setup-claude.sh` - Claude Code CLI
- `setup-all-optional.sh` - All except models
- `setup-gh200.sh` - Original monolithic script (all-in-one)

## Single-Line Copy-Paste Scripts

For fastest deployment, here are one-liners you can paste:

```bash
# Core only
curl -sSL YOUR_URL/setup-core.sh | bash

# Core + PyTorch (minimal AI setup)
curl -sSL YOUR_URL/setup-core.sh | bash && curl -sSL YOUR_URL/setup-pytorch.sh | bash

# Everything except models
curl -sSL YOUR_URL/setup-core.sh | bash && curl -sSL YOUR_URL/setup-all-optional.sh | bash
```
