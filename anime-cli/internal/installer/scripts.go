package installer

// Embedded installation scripts
var Scripts = map[string]string{
	"core": `#!/bin/bash
set -e

# Wait for dpkg lock
wait_for_dpkg() {
    local max_wait=300  # 5 minutes max
    local waited=0
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        if [ $waited -ge $max_wait ]; then
            echo "Timeout waiting for package manager lock"
            return 1
        fi
        echo "Waiting for other package managers to finish... ($waited/$max_wait s)"
        sleep 5
        waited=$((waited + 5))
    done
    return 0
}

# Fix any broken packages before starting
fix_broken_packages() {
    echo "==> Checking for broken packages..."
    if ! sudo dpkg --audit >/dev/null 2>&1; then
        echo "==> Fixing broken packages..."
        wait_for_dpkg || return 1
        sudo dpkg --configure -a || true
        wait_for_dpkg || return 1
        sudo apt-get install -f -y || true
    fi
}

# Run cleanup before starting
fix_broken_packages

echo "==> Installing Core System (Essential build tools only)"
wait_for_dpkg
sudo apt update
# Removed: apt upgrade -y (too aggressive, causes interrupted installations)
sudo apt install -y build-essential git curl wget aria2 vim htop tmux cmake pkg-config \
    libssl-dev libffi-dev python3 python3-pip python3-venv python3-dev

echo "==> Core system installed successfully"
echo "==> Note: NVIDIA drivers, Docker, and Node.js are separate packages"
echo "==> Install them separately if needed: anime install nvidia docker nodejs"
`,

	"python": `#!/bin/bash
set -e
echo "==> Setting up Python environment"

# Only upgrade pip if needed (check version)
CURRENT_PIP=$(pip3 --version 2>/dev/null | awk '{print $2}' || echo "0")
MAJOR_VERSION=$(echo $CURRENT_PIP | cut -d. -f1)

if [ "$MAJOR_VERSION" -lt 23 ]; then
    echo "==> Upgrading pip from $CURRENT_PIP to latest"
    pip3 install --upgrade pip setuptools wheel
else
    echo "==> pip $CURRENT_PIP is already recent, skipping upgrade"
fi

pip3 install --upgrade-strategy only-if-needed numpy scipy pandas matplotlib pillow
echo "==> Python environment ready"
python3 --version
pip3 --version
`,

	"pytorch": `#!/bin/bash
set -e
echo "==> Installing PyTorch and AI libraries"

# Quick check if PyTorch is installed (without slow CUDA check)
if python3 -c "import torch" 2>/dev/null; then
    TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)" 2>/dev/null || echo "unknown")
    echo "==> PyTorch $TORCH_VERSION already installed"
    echo "==> Installing/updating AI libraries only..."
    # Install ALL packages in one command to avoid dependency conflicts
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers accelerate safetensors xformers bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python
else
    echo "==> Installing PyTorch with CUDA 12.6 support..."
    # Install PyTorch with CUDA 12.6 support (latest stable, compatible with modern packages)
    pip3 install --upgrade-strategy only-if-needed torch torchvision torchaudio xformers --index-url https://download.pytorch.org/whl/cu126
    # Install ALL AI libraries in one command to avoid dependency conflicts
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python
fi

echo "==> PyTorch installed successfully"
# Quick CUDA check without full initialization
python3 -c "import torch; print(f'PyTorch {torch.__version__} | CUDA available: {torch.cuda.is_available()}')" 2>/dev/null || echo "PyTorch installed"
`,

	"ollama": `#!/bin/bash
set -e
echo "==> Installing Ollama"
if command -v ollama &> /dev/null; then
    echo "Ollama already installed"
    exit 0
fi

curl -fsSL https://ollama.com/install.sh | sh

# Create systemd service
sudo tee /etc/systemd/system/ollama.service > /dev/null <<EOF
[Unit]
Description=Ollama Service
After=network-online.target

[Service]
ExecStart=/usr/local/bin/ollama serve
User=$USER
Group=$USER
Restart=always
RestartSec=3
Environment="OLLAMA_HOST=0.0.0.0:11434"

[Install]
WantedBy=default.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now ollama
sleep 5
echo "==> Ollama installed successfully"
`,

	"models-small": `#!/bin/bash
set -e
echo "==> Downloading small models (7B)"
ollama pull mistral:latest &
ollama pull llama3.3:8b &
ollama pull qwen2.5:7b &
wait
echo "==> Small models downloaded"
ollama list
`,

	"models-medium": `#!/bin/bash
set -e
echo "==> Downloading medium models (14-34B)"
ollama pull qwen2.5:14b &
ollama pull mixtral:8x7b &
ollama pull deepseek-coder:33b &
wait
echo "==> Medium models downloaded"
ollama list
`,

	"models-large": `#!/bin/bash
set -e
echo "==> Downloading large models (70B+)"
ollama pull llama3.3:70b &
ollama pull qwen2.5:72b &
wait
echo "==> Large models downloaded"
ollama list
`,

	"comfyui": `#!/bin/bash
set -e
echo "==> Installing ComfyUI"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

COMFYUI_DIR="$HOME/ComfyUI"
if [ -d "$COMFYUI_DIR" ]; then
    echo "ComfyUI already exists"
    exit 0
fi

echo "Cloning ComfyUI..."
git clone https://github.com/comfyanonymous/ComfyUI.git "$COMFYUI_DIR"

echo "Installing ComfyUI dependencies (excluding torch/torchvision to preserve CUDA setup)..."
# Filter out torch/torchvision/torchaudio from requirements to avoid reinstalling and breaking CUDA
grep -v "^torch" "$COMFYUI_DIR/requirements.txt" > /tmp/comfyui-requirements-filtered.txt || true

# Install filtered requirements
if [ -s /tmp/comfyui-requirements-filtered.txt ]; then
    pip3 install -r /tmp/comfyui-requirements-filtered.txt --upgrade-strategy only-if-needed
fi
rm -f /tmp/comfyui-requirements-filtered.txt

echo "Installing ComfyUI Manager..."
git clone https://github.com/ltdrdata/ComfyUI-Manager.git "$COMFYUI_DIR/custom_nodes/ComfyUI-Manager"

echo "==> ComfyUI installed successfully"
echo "==> PyTorch and CUDA installation preserved"
`,

	"nvidia": `#!/bin/bash
set -e

# Wait for dpkg lock
wait_for_dpkg() {
    local max_wait=300  # 5 minutes max
    local waited=0
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        if [ $waited -ge $max_wait ]; then
            echo "Timeout waiting for package manager lock"
            return 1
        fi
        echo "Waiting for other package managers to finish... ($waited/$max_wait s)"
        sleep 5
        waited=$((waited + 5))
    done
    return 0
}

echo "==> Installing NVIDIA Drivers and CUDA"
if command -v nvidia-smi &> /dev/null; then
    echo "NVIDIA drivers already installed"
    nvidia-smi
    exit 0
fi

echo "==> Downloading CUDA keyring..."
wget -q https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/arm64/cuda-keyring_1.1-1_all.deb -O /tmp/cuda-keyring.deb
wait_for_dpkg
sudo dpkg -i /tmp/cuda-keyring.deb
wait_for_dpkg
sudo apt update
wait_for_dpkg

echo "==> Installing CUDA toolkit and NVIDIA drivers..."
sudo apt install -y cuda-toolkit-12-4 nvidia-driver-550

echo "==> NVIDIA drivers installed successfully"
nvidia-smi || echo "Reboot required for NVIDIA drivers to load"
`,

	"docker": `#!/bin/bash
set -e

# Wait for dpkg lock
wait_for_dpkg() {
    local max_wait=300  # 5 minutes max
    local waited=0
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        if [ $waited -ge $max_wait ]; then
            echo "Timeout waiting for package manager lock"
            return 1
        fi
        echo "Waiting for other package managers to finish... ($waited/$max_wait s)"
        sleep 5
        waited=$((waited + 5))
    done
    return 0
}

echo "==> Installing Docker"
if command -v docker &> /dev/null; then
    echo "Docker already installed"
    docker --version
    exit 0
fi

wait_for_dpkg
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

echo "==> Docker installed successfully"
docker --version
echo "==> Note: Log out and back in for docker group to take effect"
`,

	"nodejs": `#!/bin/bash
set -e

# Wait for dpkg lock
wait_for_dpkg() {
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        echo "Waiting for other package managers to finish..."
        sleep 5
    done
}

echo "==> Installing Node.js and npm"
if command -v node &> /dev/null; then
    echo "Node.js $(node --version) already installed"
    exit 0
fi

echo "==> Installing Node.js 20.x LTS"
wait_for_dpkg
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
wait_for_dpkg
sudo apt install -y nodejs
sudo npm install -g yarn pnpm typescript

echo "==> Node.js installed successfully"
node --version
npm --version
`,

	"claude": `#!/bin/bash
set -e
echo "==> Installing Claude Code CLI"
if command -v claude-code &> /dev/null; then
    echo "Claude Code already installed"
    exit 0
fi
sudo npm install -g @anthropic-ai/claude-code

# Source shell config to make claude-code available immediately
if [ -f "$HOME/.zshrc" ]; then
    echo "==> Sourcing .zshrc..."
    export PATH="$PATH:/usr/local/bin:$HOME/.npm-global/bin"
    source "$HOME/.zshrc" 2>/dev/null || true
elif [ -f "$HOME/.bashrc" ]; then
    echo "==> Sourcing .bashrc..."
    export PATH="$PATH:/usr/local/bin:$HOME/.npm-global/bin"
    source "$HOME/.bashrc" 2>/dev/null || true
fi

echo "==> Claude Code installed successfully"
echo "==> Verifying installation..."
which claude-code || echo "Note: You may need to restart your shell for claude-code to be in PATH"
`,

	// Video Generation Models
	"mochi": `#!/bin/bash
set -e
echo "==> Installing Mochi-1 Video Generation"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

mkdir -p ~/video-models
cd ~/video-models
if [ -d "mochi-1" ]; then
    echo "Mochi-1 already installed"
    exit 0
fi

# Check which dependencies need to be installed
echo "==> Checking dependencies..."
INSTALL_DEPS=""
python3 -c "import diffusers" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS diffusers"
python3 -c "import transformers" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS transformers"
python3 -c "import accelerate" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS accelerate"
python3 -c "import einops" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS einops"

if [ -n "$INSTALL_DEPS" ]; then
    echo "==> Installing missing dependencies:$INSTALL_DEPS"
    pip3 install --upgrade-strategy only-if-needed $INSTALL_DEPS
else
    echo "==> All dependencies already installed"
fi

git clone https://github.com/genmoai/mochi mochi-1
cd mochi-1
if [ -f "requirements.txt" ]; then
    echo "==> Installing additional requirements (excluding torch/cuda to avoid conflicts)..."
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/mochi-requirements-filtered.txt || true
    if [ -s /tmp/mochi-requirements-filtered.txt ]; then
        pip3 install -r /tmp/mochi-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi
echo "==> Downloading Mochi-1 model weights with parallel downloads..."
huggingface-cli download genmo/mochi-1-preview --local-dir ./weights --max-workers 8
echo "==> Mochi-1 installed successfully"
`,

	"svd": `#!/bin/bash
set -e
echo "==> Installing Stable Video Diffusion for ComfyUI"
COMFY_DIR="$HOME/ComfyUI"
if [ ! -d "$COMFY_DIR" ]; then
    echo "Error: ComfyUI not found. Install comfyui first."
    exit 1
fi
cd "$COMFY_DIR/custom_nodes"
if [ -d "ComfyUI-VideoHelperSuite" ]; then
    echo "SVD already installed"
    exit 0
fi
git clone https://github.com/Kosinkadink/ComfyUI-VideoHelperSuite.git
cd "$COMFY_DIR/models/checkpoints"
echo "==> Downloading SVD model with multi-connection download..."
# Use aria2c if available for faster downloads, fallback to wget
if command -v aria2c &> /dev/null; then
    aria2c -x 16 -s 16 https://huggingface.co/stabilityai/stable-video-diffusion-img2vid-xt/resolve/main/svd_xt.safetensors
else
    wget -c https://huggingface.co/stabilityai/stable-video-diffusion-img2vid-xt/resolve/main/svd_xt.safetensors
fi
echo "==> SVD installed successfully"
`,

	"animatediff": `#!/bin/bash
set -e
echo "==> Installing AnimateDiff for ComfyUI"
COMFY_DIR="$HOME/ComfyUI"
if [ ! -d "$COMFY_DIR" ]; then
    echo "Error: ComfyUI not found. Install comfyui first."
    exit 1
fi
cd "$COMFY_DIR/custom_nodes"
if [ -d "ComfyUI-AnimateDiff-Evolved" ]; then
    echo "AnimateDiff already installed"
    exit 0
fi
git clone https://github.com/Kosinkadink/ComfyUI-AnimateDiff-Evolved.git
mkdir -p "$COMFY_DIR/models/animatediff_models"
cd "$COMFY_DIR/models/animatediff_models"
echo "==> Downloading AnimateDiff motion module with multi-connection download..."
if command -v aria2c &> /dev/null; then
    aria2c -x 16 -s 16 https://huggingface.co/guoyww/animatediff/resolve/main/mm_sd_v15_v2.ckpt
else
    wget -c https://huggingface.co/guoyww/animatediff/resolve/main/mm_sd_v15_v2.ckpt
fi
echo "==> AnimateDiff installed successfully"
`,

	"cogvideo": `#!/bin/bash
set -e
echo "==> Installing CogVideoX-5B"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

mkdir -p ~/video-models
cd ~/video-models
if [ -d "cogvideo" ]; then
    echo "CogVideoX already installed"
    exit 0
fi

# Install only missing dependencies (PyTorch already installed)
pip3 install --upgrade-strategy only-if-needed diffusers transformers accelerate

git clone https://github.com/THUDM/CogVideo cogvideo
cd cogvideo
if [ -f "requirements.txt" ]; then
    echo "==> Installing additional requirements (excluding torch/cuda to avoid conflicts)..."
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/cogvideo-requirements-filtered.txt || true
    if [ -s /tmp/cogvideo-requirements-filtered.txt ]; then
        pip3 install -r /tmp/cogvideo-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi
echo "==> Downloading CogVideoX-5B model with parallel downloads..."
huggingface-cli download THUDM/CogVideoX-5b --local-dir ./weights --max-workers 8
echo "==> CogVideoX installed successfully"
`,

	"opensora": `#!/bin/bash
set -e
echo "==> Installing Open-Sora 2.0"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

mkdir -p ~/video-models
cd ~/video-models
if [ -d "open-sora" ]; then
    echo "Open-Sora already installed"
    exit 0
fi

git clone https://github.com/hpcaitech/Open-Sora open-sora
cd open-sora
pip3 install -e . --upgrade-strategy only-if-needed
echo "==> Downloading Open-Sora models with parallel downloads..."
huggingface-cli download hpcai-tech/OpenSora-STDiT-v3 --local-dir ./pretrained_models --max-workers 8
echo "==> Open-Sora installed successfully"
`,

	"ltxvideo": `#!/bin/bash
set -e
echo "==> Installing LTXVideo"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

mkdir -p ~/video-models
cd ~/video-models
if [ -d "ltxvideo" ]; then
    echo "LTXVideo already installed"
    exit 0
fi

# Install only missing dependencies (PyTorch already installed)
pip3 install --upgrade-strategy only-if-needed diffusers transformers accelerate

git clone https://github.com/Lightricks/LTX-Video ltxvideo
cd ltxvideo
if [ -f "requirements.txt" ]; then
    echo "==> Installing additional requirements (excluding torch/cuda to avoid conflicts)..."
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/ltxvideo-requirements-filtered.txt || true
    if [ -s /tmp/ltxvideo-requirements-filtered.txt ]; then
        pip3 install -r /tmp/ltxvideo-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi
echo "==> Downloading LTXVideo model with parallel downloads..."
huggingface-cli download Lightricks/LTX-Video --local-dir ./checkpoints --max-workers 8
echo "==> LTXVideo installed successfully"
`,

	"wan2": `#!/bin/bash
set -e
echo "==> Installing Wan2.2 (Image-to-Video)"

# Verify PyTorch is installed
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

mkdir -p ~/video-models
cd ~/video-models
if [ -d "wan2" ]; then
    echo "Wan2.2 already installed"
    exit 0
fi

# Check which dependencies need to be installed
echo "==> Checking dependencies..."
INSTALL_DEPS=""
python3 -c "import diffusers" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS diffusers"
python3 -c "import transformers" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS transformers"
python3 -c "import accelerate" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS accelerate"
python3 -c "import einops" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS einops"
python3 -c "import imageio" 2>/dev/null || INSTALL_DEPS="$INSTALL_DEPS imageio imageio-ffmpeg"

if [ -n "$INSTALL_DEPS" ]; then
    echo "==> Installing missing dependencies:$INSTALL_DEPS"
    pip3 install --upgrade-strategy only-if-needed $INSTALL_DEPS
else
    echo "==> All dependencies already installed"
fi

# Clone Wan2 repository (public repo)
git clone https://github.com/alibaba/Wan.git wan2
cd wan2

# Install additional requirements, but skip torch/cuda packages to avoid reinstalls
if [ -f "requirements.txt" ]; then
    echo "==> Installing additional requirements (excluding torch/cuda to avoid conflicts)..."
    # Filter out torch and nvidia packages that are already installed
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/wan2-requirements-filtered.txt || true
    if [ -s /tmp/wan2-requirements-filtered.txt ]; then
        pip3 install -r /tmp/wan2-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi

echo "==> Downloading Wan2.2 model weights with parallel downloads..."
# Download model from HuggingFace with parallel workers
huggingface-cli download Alibaba-PAI/wan2.2 --local-dir ./checkpoints --max-workers 8

echo "==> Wan2.2 installed successfully"
echo "Model location: ~/video-models/wan2"
echo "Usage: See https://github.com/Wan-Video/Wan2.2 for inference examples"
`,

	"comfyui-wan2": `#!/bin/bash
set -e
echo "==> Installing ComfyUI Wan2 Wrapper"

# Verify ComfyUI is installed
if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

# Verify Wan2 is installed
if [ ! -d "$HOME/video-models/wan2" ]; then
    echo "Error: Wan2.2 not found. Please install 'wan2' package first: anime install wan2"
    exit 1
fi

cd ~/ComfyUI/custom_nodes

# Check if already installed
if [ -d "ComfyUI-WanWrapper" ]; then
    echo "ComfyUI Wan2 Wrapper already installed"
    exit 0
fi

echo "==> Cloning ComfyUI-WanWrapper..."
git clone https://github.com/kijai/ComfyUI-WanWrapper

cd ComfyUI-WanWrapper

# Install requirements if available
if [ -f "requirements.txt" ]; then
    echo "==> Installing wrapper requirements (excluding torch/cuda to avoid conflicts)..."
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/comfyui-wan2-requirements-filtered.txt || true
    if [ -s /tmp/comfyui-wan2-requirements-filtered.txt ]; then
        pip3 install -r /tmp/comfyui-wan2-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi

# Create symlink to Wan2 model if needed
if [ ! -L "models" ] && [ -d "$HOME/video-models/wan2" ]; then
    echo "==> Linking Wan2 models to ComfyUI..."
    ln -s "$HOME/video-models/wan2/checkpoints" models
fi

echo "==> ComfyUI Wan2 Wrapper installed successfully"
echo "Location: ~/ComfyUI/custom_nodes/ComfyUI-WanWrapper"
echo "Restart ComfyUI to load the new nodes"
`,

	// Individual LLM Models (via Ollama)
	"llama-3.3-70b": `#!/bin/bash
set -e
echo "==> Downloading Llama 3.3 70B via Ollama"
ollama pull llama3.3:70b
echo "==> Llama 3.3 70B installed successfully"
ollama list
`,

	"llama-3.3-8b": `#!/bin/bash
set -e
echo "==> Downloading Llama 3.3 8B via Ollama"
ollama pull llama3.3:8b
echo "==> Llama 3.3 8B installed successfully"
ollama list
`,

	"mistral": `#!/bin/bash
set -e
echo "==> Downloading Mistral 7B via Ollama"
ollama pull mistral:latest
echo "==> Mistral 7B installed successfully"
ollama list
`,

	"mixtral": `#!/bin/bash
set -e
echo "==> Downloading Mixtral 8x7B via Ollama"
ollama pull mixtral:8x7b
echo "==> Mixtral 8x7B installed successfully"
ollama list
`,

	"qwen-2.5-72b": `#!/bin/bash
set -e
echo "==> Downloading Qwen 2.5 72B via Ollama"
ollama pull qwen2.5:72b
echo "==> Qwen 2.5 72B installed successfully"
ollama list
`,

	"qwen-2.5-14b": `#!/bin/bash
set -e
echo "==> Downloading Qwen 2.5 14B via Ollama"
ollama pull qwen2.5:14b
echo "==> Qwen 2.5 14B installed successfully"
ollama list
`,

	"qwen-2.5-7b": `#!/bin/bash
set -e
echo "==> Downloading Qwen 2.5 7B via Ollama"
ollama pull qwen2.5:7b
echo "==> Qwen 2.5 7B installed successfully"
ollama list
`,

	"deepseek-coder-33b": `#!/bin/bash
set -e
echo "==> Downloading DeepSeek Coder 33B via Ollama"
ollama pull deepseek-coder:33b
echo "==> DeepSeek Coder 33B installed successfully"
ollama list
`,

	"deepseek-v3": `#!/bin/bash
set -e
echo "==> Downloading DeepSeek V3 via Ollama"
ollama pull deepseek-v3
echo "==> DeepSeek V3 installed successfully"
ollama list
`,

	"phi-3.5": `#!/bin/bash
set -e
echo "==> Downloading Phi-3.5 Mini via Ollama"
ollama pull phi3.5:latest
echo "==> Phi-3.5 Mini installed successfully"
ollama list
`,

	// Individual Image Generation Models (for ComfyUI)
	"sdxl": `#!/bin/bash
set -e
echo "==> Installing Stable Diffusion XL for ComfyUI"

# Verify ComfyUI is installed
if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/checkpoints"
cd "$HOME/ComfyUI/models/checkpoints"

# Check if already installed
if [ -f "sd_xl_base_1.0.safetensors" ]; then
    echo "SDXL already installed"
    exit 0
fi

echo "==> Downloading SDXL base model with multi-connection download..."
if command -v aria2c &> /dev/null; then
    aria2c -x 16 -s 16 https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0/resolve/main/sd_xl_base_1.0.safetensors
else
    wget -c https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0/resolve/main/sd_xl_base_1.0.safetensors
fi

echo "==> SDXL installed successfully"
echo "Model location: ~/ComfyUI/models/checkpoints/sd_xl_base_1.0.safetensors"
`,

	"sd15": `#!/bin/bash
set -e
echo "==> Installing Stable Diffusion 1.5 for ComfyUI"

# Verify ComfyUI is installed
if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/checkpoints"
cd "$HOME/ComfyUI/models/checkpoints"

# Check if already installed
if [ -f "v1-5-pruned-emaonly.safetensors" ]; then
    echo "SD 1.5 already installed"
    exit 0
fi

echo "==> Downloading SD 1.5 model with multi-connection download..."
if command -v aria2c &> /dev/null; then
    aria2c -x 16 -s 16 https://huggingface.co/runwayml/stable-diffusion-v1-5/resolve/main/v1-5-pruned-emaonly.safetensors
else
    wget -c https://huggingface.co/runwayml/stable-diffusion-v1-5/resolve/main/v1-5-pruned-emaonly.safetensors
fi

echo "==> SD 1.5 installed successfully"
echo "Model location: ~/ComfyUI/models/checkpoints/v1-5-pruned-emaonly.safetensors"
`,

	"flux-dev": `#!/bin/bash
set -e
echo "==> Installing Flux.1 Dev for ComfyUI"

# Verify ComfyUI is installed
if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/unet"
cd "$HOME/ComfyUI/models/unet"

# Check if already installed
if [ -f "flux1-dev.safetensors" ]; then
    echo "Flux.1 Dev already installed"
    exit 0
fi

echo "==> Downloading Flux.1 Dev model with parallel downloads..."
huggingface-cli download black-forest-labs/FLUX.1-dev flux1-dev.safetensors --local-dir . --max-workers 8

echo "==> Flux.1 Dev installed successfully"
echo "Model location: ~/ComfyUI/models/unet/flux1-dev.safetensors"
`,

	"flux-schnell": `#!/bin/bash
set -e
echo "==> Installing Flux.1 Schnell for ComfyUI"

# Verify ComfyUI is installed
if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/unet"
cd "$HOME/ComfyUI/models/unet"

# Check if already installed
if [ -f "flux1-schnell.safetensors" ]; then
    echo "Flux.1 Schnell already installed"
    exit 0
fi

echo "==> Downloading Flux.1 Schnell model with parallel downloads..."
huggingface-cli download black-forest-labs/FLUX.1-schnell flux1-schnell.safetensors --local-dir . --max-workers 8

echo "==> Flux.1 Schnell installed successfully"
echo "Model location: ~/ComfyUI/models/unet/flux1-schnell.safetensors"
`,
}

