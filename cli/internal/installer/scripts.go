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
    libssl-dev libffi-dev python3 python3-pip python3-venv python3-dev \
    jq unzip ripgrep fd-find rsync sqlite3 net-tools dnsutils \
    fail2ban ufw nginx certbot python3-certbot-nginx

echo "==> Enabling fail2ban..."
sudo systemctl enable --now fail2ban

echo "==> Configuring ufw (SSH allowed)..."
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw --force enable

echo "==> Core system installed successfully"
echo "==> Includes: build tools, jq, ripgrep, rsync, sqlite3, nginx, certbot, fail2ban, ufw"
`,

	"python": `#!/bin/bash
set -e
echo "==> Setting up Python environment"

# Use --break-system-packages on systems with PEP 668 (externally managed)
PIP_EXTRA=""
if pip3 install --help 2>&1 | grep -q "break-system-packages"; then
    PIP_EXTRA="--break-system-packages"
fi

# Only upgrade pip if needed (check version)
CURRENT_PIP=$(pip3 --version 2>/dev/null | awk '{print $2}' || echo "0")
MAJOR_VERSION=$(echo $CURRENT_PIP | cut -d. -f1)

if [ "$MAJOR_VERSION" -lt 23 ]; then
    echo "==> Upgrading pip from $CURRENT_PIP to latest"
    pip3 install --upgrade pip setuptools wheel $PIP_EXTRA
else
    echo "==> pip $CURRENT_PIP is already recent, skipping upgrade"
fi

pip3 install --upgrade-strategy only-if-needed $PIP_EXTRA numpy scipy pandas matplotlib pillow
echo "==> Python environment ready"
python3 --version
pip3 --version
`,

	"pytorch": `#!/bin/bash
set -e
echo "==> Installing PyTorch and AI libraries"

if python3 -c "import torch" 2>/dev/null; then
    TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)" 2>/dev/null || echo "unknown")

    # Refuse to touch a torch install whose CUDA is already broken — adding more
    # packages on top will only mask the root cause.
    if ! python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
        echo "==> ERROR: PyTorch $TORCH_VERSION is installed but CUDA is unavailable."
        echo "==> A prior pip install likely replaced torch with a wheel built for the wrong CUDA."
        echo "==> Recovery:"
        echo "==>   pip uninstall -y torch torchvision torchaudio xformers"
        echo "==>   anime install pytorch"
        exit 1
    fi

    echo "==> PyTorch $TORCH_VERSION already installed (CUDA OK)"

    # Pin torch via PIP_CONSTRAINT so transitive deps (xformers, bitsandbytes, etc.)
    # cannot silently upgrade it onto a wheel built for the wrong CUDA.
    CONSTRAINT_FILE="$HOME/.config/anime/torch-constraints.txt"
    mkdir -p "$(dirname "$CONSTRAINT_FILE")"
    {
        echo "torch==$TORCH_VERSION"
        echo "torchvision"
        echo "torchaudio"
    } > "$CONSTRAINT_FILE"
    export PIP_CONSTRAINT="$CONSTRAINT_FILE"
    echo "==> Pinned torch via $CONSTRAINT_FILE"

    # xformers declares torch>=N which will trip the constraint. --no-deps bypasses
    # its torch requirement and reuses the working torch already present.
    pip3 install --upgrade-strategy only-if-needed --no-deps xformers
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python

    # Verify nothing snuck past the constraint.
    if ! python3 -c "import torch; assert torch.cuda.is_available(), 'replaced'" 2>/dev/null; then
        echo "==> FATAL: torch was replaced despite PIP_CONSTRAINT. CUDA broken."
        echo "==> Recovery: pip uninstall -y torch torchvision torchaudio xformers && anime install pytorch"
        exit 1
    fi
else
    echo "==> Installing PyTorch with CUDA 12.6 support..."
    pip3 install --upgrade-strategy only-if-needed torch torchvision torchaudio xformers --index-url https://download.pytorch.org/whl/cu126
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python
fi

echo "==> PyTorch installed successfully"
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

	"vllm": `#!/bin/bash
set -e
echo "==> Installing vLLM Inference Engine"

if ! python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
    echo "Error: PyTorch with CUDA not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

if python3 -c "import vllm" 2>/dev/null; then
    VLLM_VERSION=$(python3 -c "import vllm; print(vllm.__version__)" 2>/dev/null || echo "unknown")
    echo "==> vLLM $VLLM_VERSION already installed"
    exit 0
fi

# Pin torch BEFORE vllm install. vllm declares torch as a hard dep, and without
# this guard pip will silently pull a torch wheel built for a different CUDA
# version (e.g. cu130 wheels on a cu128 driver -> CUDA unavailable).
TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)")
TORCH_CUDA=$(python3 -c "import torch; print(torch.version.cuda)")
ARCH=$(uname -m)
CONSTRAINT_FILE="$HOME/.config/anime/torch-constraints.txt"
mkdir -p "$(dirname "$CONSTRAINT_FILE")"
{
    echo "torch==$TORCH_VERSION"
    echo "torchvision"
    echo "torchaudio"
    echo "numpy<2"
} > "$CONSTRAINT_FILE"
export PIP_CONSTRAINT="$CONSTRAINT_FILE"
echo "==> Pinned torch==$TORCH_VERSION (CUDA $TORCH_CUDA) via $CONSTRAINT_FILE"

# aarch64 + CUDA 12 path:
# - All prebuilt aarch64 vllm wheels from vllm 0.20+ are cu13; they fail on
#   driver <580 with "libcudart.so.13: cannot open shared object file".
# - vllm 0.10.1.1 is the last release whose source-build can target cu12 cleanly
#   (pins torch==2.7.1 which matches the cu128 PyTorch wheel index).
# - GH200 is sm_90 (Hopper); TORCH_CUDA_ARCH_LIST="9.0" keeps the build small
#   and ensures the resulting .so is loadable on this card.
NEEDS_SOURCE_BUILD=false
VLLM_PIN=""
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    TORCH_CUDA_MAJOR=$(echo "$TORCH_CUDA" | cut -d. -f1)
    if [ "$TORCH_CUDA_MAJOR" = "12" ]; then
        NEEDS_SOURCE_BUILD=true
        VLLM_PIN="==0.10.1.1"
        echo "==> aarch64 + CUDA 12 detected — source-building vllm 0.10.1.1 (last cu12-compatible release)"

        # vllm 0.10.1.1 needs torch==2.7.1. If current torch is older/different,
        # install the cu128 wheel to user-site (shadows system torch).
        TORCH_MAJOR_MINOR=$(echo "$TORCH_VERSION" | cut -d. -f1,2)
        if [ "$TORCH_MAJOR_MINOR" != "2.7" ]; then
            echo "==> Installing cu128 torch==2.7.1 for vllm 0.10.x compatibility..."
            pip3 install --upgrade-strategy only-if-needed \
                --index-url https://download.pytorch.org/whl/cu128 \
                torch==2.7.1 torchvision==0.22.1 torchaudio==2.7.1
            # Refresh constraint file to the new torch version we just pinned.
            {
                echo "torch==2.7.1+cu128"
                echo "torchvision"
                echo "torchaudio"
                echo "numpy<2"
            } > "$CONSTRAINT_FILE"
            TORCH_VERSION="2.7.1"
        fi
    fi
fi

echo "==> Installing vLLM build prerequisites..."
pip3 install --upgrade-strategy only-if-needed \
    'numpy<2' pybind11 setuptools setuptools_scm wheel cmake ninja \
    'huggingface_hub[cli]' hf_transfer

if [ "$NEEDS_SOURCE_BUILD" = true ]; then
    # --no-binary=vllm: refuse the cu13 prebuilt wheel; build from sdist
    # --no-build-isolation: use the EXISTING torch instead of pip downloading
    #     a different torch version (likely cu13 wheel) for the build
    # --no-deps: vllm's METADATA pins one torch version per release;
    #     manual runtime-dep install below avoids the version conflict
    # TORCH_CUDA_ARCH_LIST=9.0a (not "9.0"): enables Hopper-only WGMMA, TMA, FA3,
    # Marlin-FP8 kernels — worth +31% prefill throughput per agent research
    # (docs/research/SOTA_GH200_INFERENCE_2026.md, Agent 9). Plain "9.0" silently
    # disables these via __CUDA_ARCH_FEAT_SM90_ALL__ guards → FA2 fallback.
    # VLLM_FA_CMAKE_GPU_ARCHES=90a-real pairs with above to actually build FA3.
    echo "==> Building vLLM 0.10.1.1 from source for sm_90a (Hopper WGMMA/TMA/FA3) — ~25 min on GH200..."
    SCCACHE_LAUNCHER=""
    if command -v sccache >/dev/null 2>&1; then
        SCCACHE_LAUNCHER="CMAKE_CUDA_COMPILER_LAUNCHER=sccache CMAKE_CXX_COMPILER_LAUNCHER=sccache"
        echo "==> sccache detected — cache hits cut rebuild time ~90%"
    fi
    env $SCCACHE_LAUNCHER \
        TORCH_CUDA_ARCH_LIST="9.0a" \
        VLLM_FA_CMAKE_GPU_ARCHES="90a-real" \
        CUDA_HOME=/usr/lib/cuda \
        MAX_JOBS=32 NVCC_THREADS=4 \
        pip3 install --no-binary=vllm --no-build-isolation --no-deps "vllm${VLLM_PIN}"
else
    echo "==> Installing vLLM prebuilt wheel..."
    pip3 install --no-deps "vllm${VLLM_PIN}"
fi

echo "==> Installing vLLM runtime dependencies (torch left alone)..."
pip3 install --upgrade-strategy only-if-needed \
    'transformers>=4.40' 'tokenizers>=0.19' sentencepiece 'accelerate>=0.26' \
    fastapi 'uvicorn[standard]' 'pydantic>=2.0' \
    prometheus-client py-cpuinfo msgspec gguf \
    aiohttp openai pyzmq cloudpickle \
    blake3 cbor2 cachetools diskcache ijson lark numba \
    opencv-python-headless outlines_core partial-json-parser \
    pybase64 python-json-logger setproctitle tiktoken \
    watchfiles tqdm regex pillow protobuf psutil pyyaml \
    fastsafetensors lm-format-enforcer xgrammar mistral_common \
    openai-harmony compressed-tensors flashinfer-python \
    apache-tvm-ffi prometheus-fastapi-instrumentator

# Verify torch survived.
if ! python3 -c "import torch; assert torch.cuda.is_available(), 'replaced'" 2>/dev/null; then
    echo "==> FATAL: torch was replaced despite PIP_CONSTRAINT. CUDA broken."
    echo "==> Recovery:"
    echo "==>   rm -rf ~/.local/lib/python3.10/site-packages/{torch*,torchvision*,torchaudio*,nvidia*,triton*,cuda_*,xformers*,vllm*}"
    echo "==>   anime install pytorch && anime install vllm"
    exit 1
fi

# Verify vllm._C loads (catches the libcudart.so.13 trap before runtime).
echo "==> Verifying vLLM C extension links against the installed CUDA..."
if ! python3 -c "import vllm._C" 2>/dev/null; then
    echo "==> FATAL: vllm._C failed to load. Likely a CUDA runtime mismatch."
    echo "==> Driver supports CUDA $TORCH_CUDA; vllm wheel may target a different CUDA."
    python3 -c "import vllm._C" 2>&1 | tail -5
    exit 1
fi

echo "==> Verifying vLLM installation..."
python3 -c "import vllm, torch; print(f'vLLM {vllm.__version__} | torch {torch.__version__} | CUDA {torch.cuda.is_available()}')"

echo "==> vLLM installed successfully"
echo ""
echo "Usage examples:"
echo "  1. Python API: from vllm import LLM, SamplingParams"
echo "  2. OpenAI-compatible server: python3 -m vllm.entrypoints.openai.api_server --model <model-name>"
echo "  3. Offline inference: vllm serve <model-name>"
echo ""
echo "Documentation: https://docs.vllm.ai/"
`,

	"models-small": `#!/bin/bash
set -e
echo "==> Downloading small models (1-8B)"
ollama pull llama3.2:1b &
ollama pull llama3.2:3b &
ollama pull gemma3:4b &
ollama pull mistral:latest &
ollama pull llama3.3:8b &
ollama pull qwen3:8b &
wait
echo "==> Small models downloaded"
ollama list
`,

	"models-medium": `#!/bin/bash
set -e
echo "==> Downloading medium models (8-34B)"
ollama pull deepseek-r1:8b &
ollama pull phi4:latest &
ollama pull gemma3:12b &
ollama pull qwen3:14b &
ollama pull qwen3-coder:30b &
ollama pull qwen3:32b &
ollama pull mixtral:8x7b &
ollama pull deepseek-coder:33b &
wait
echo "==> Medium models downloaded"
ollama list
`,

	"models-large": `#!/bin/bash
set -e
echo "==> Downloading large models (70B+)"
ollama pull gemma3:27b &
ollama pull deepseek-r1:70b &
ollama pull llama3.3:70b &
ollama pull qwen3:235b &
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

	"go": `#!/bin/bash
set -e

echo "==> Installing Go"
if command -v go &> /dev/null; then
    echo "Go $(go version) already installed"
    exit 0
fi

# Get latest stable Go version (or use a specific version)
GO_VERSION="1.23.5"
GO_ARCH="amd64"

# Detect architecture
if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
    GO_ARCH="arm64"
fi

echo "==> Downloading Go $GO_VERSION for $GO_ARCH"
wget -q https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz -O /tmp/go.tar.gz

echo "==> Installing Go to /usr/local"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

# Add Go to PATH if not already present
if ! grep -q "/usr/local/go/bin" ~/.profile 2>/dev/null; then
    echo "==> Adding Go to PATH in ~/.profile"
    echo "" >> ~/.profile
    echo "# Go" >> ~/.profile
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.profile
fi

if ! grep -q "/usr/local/go/bin" ~/.bashrc 2>/dev/null; then
    echo "==> Adding Go to PATH in ~/.bashrc"
    echo "" >> ~/.bashrc
    echo "# Go" >> ~/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
fi

# Also add to current session
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

echo "==> Go installed successfully"
/usr/local/go/bin/go version

echo ""
echo "Note: Restart your shell or run 'source ~/.profile' to update PATH"
echo "Go workspace: ~/go"
`,

	"claude": `#!/bin/bash
set -e
echo "==> Installing Claude Code CLI"

# Check if already installed
if command -v claude &> /dev/null; then
    echo "Claude Code already installed: $(which claude)"
    exit 0
fi

# Find npm — check PATH, homebrew, nvm, common locations
NPM=""
for candidate in \
    npm \
    /opt/homebrew/bin/npm \
    /usr/local/bin/npm \
    ; do
    if command -v "$candidate" &> /dev/null || [ -x "$candidate" ]; then
        NPM="$candidate"
        break
    fi
done

# Try nvm
if [ -z "$NPM" ]; then
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    command -v npm &> /dev/null && NPM="npm"
fi

# Try globbing nvm paths
if [ -z "$NPM" ]; then
    NPM=$(ls -1 $HOME/.nvm/versions/node/*/bin/npm $HOME/.local/share/nvm/*/bin/npm 2>/dev/null | tail -1)
fi

if [ -z "$NPM" ]; then
    echo "Error: npm not found. Install nodejs first: anime install nodejs"
    exit 1
fi

echo "==> Using npm at: $NPM"

# Try without sudo first, fall back to sudo
$NPM install -g @anthropic-ai/claude-code 2>/dev/null || sudo $NPM install -g @anthropic-ai/claude-code

echo "==> Claude Code installed successfully"
command -v claude && echo "==> Verified: $(which claude)" || echo "Note: Restart your shell to pick up claude in PATH"
`,

	"rust": `#!/bin/bash
set -e
echo "==> Installing Rust Toolchain"
if command -v rustc &> /dev/null; then
    echo "Rust $(rustc --version) already installed"
    exit 0
fi

curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable

# Source cargo env for current session
source "$HOME/.cargo/env"

# Common tools
echo "==> Installing common Rust tools..."
cargo install sccache cargo-watch cargo-edit

echo "==> Rust installed successfully"
rustc --version
cargo --version
echo ""
echo "Note: Restart your shell or run 'source ~/.cargo/env' to update PATH"
`,

	"gh": `#!/bin/bash
set -e
echo "==> Installing GitHub CLI"
if command -v gh &> /dev/null; then
    echo "GitHub CLI $(gh --version | head -1) already installed"
    exit 0
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    if [[ "$ID" == "ubuntu" || "$ID" == "debian" ]]; then
        (type -p wget >/dev/null || (sudo apt update && sudo apt-get install wget -y)) \
        && sudo mkdir -p -m 755 /etc/apt/keyrings \
        && out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        && cat $out | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
        && sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
        && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
        && sudo apt update \
        && sudo apt install gh -y
    fi
elif [[ "$(uname)" == "Darwin" ]]; then
    brew install gh
fi

echo "==> GitHub CLI installed successfully"
gh --version
`,

	"uv": `#!/bin/bash
set -e
echo "==> Installing uv (fast Python package manager)"
if command -v uv &> /dev/null; then
    echo "uv $(uv --version) already installed"
    exit 0
fi

curl -LsSf https://astral.sh/uv/install.sh | sh

echo "==> uv installed successfully"
echo "Note: Restart your shell or run 'source $HOME/.local/bin/env' to update PATH"
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

	"qwen3-235b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 235B MoE via Ollama"
ollama pull qwen3:235b
echo "==> Qwen3 235B installed successfully"
ollama list
`,

	"qwen3-32b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 32B via Ollama"
ollama pull qwen3:32b
echo "==> Qwen3 32B installed successfully"
ollama list
`,

	"qwen3-30b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 30B MoE via Ollama"
ollama pull qwen3:30b
echo "==> Qwen3 30B installed successfully"
ollama list
`,

	"qwen3-14b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 14B via Ollama"
ollama pull qwen3:14b
echo "==> Qwen3 14B installed successfully"
ollama list
`,

	"qwen3-8b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 8B via Ollama"
ollama pull qwen3:8b
echo "==> Qwen3 8B installed successfully"
ollama list
`,

	"qwen3-4b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3 4B via Ollama"
ollama pull qwen3:4b
echo "==> Qwen3 4B installed successfully"
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

	"phi-4": `#!/bin/bash
set -e
echo "==> Downloading Phi-4 14B via Ollama"
ollama pull phi4:latest
echo "==> Phi-4 14B installed successfully"
ollama list
`,

	"deepseek-r1-8b": `#!/bin/bash
set -e
echo "==> Downloading DeepSeek-R1 8B via Ollama"
ollama pull deepseek-r1:8b
echo "==> DeepSeek-R1 8B installed successfully"
ollama list
`,

	"deepseek-r1-70b": `#!/bin/bash
set -e
echo "==> Downloading DeepSeek-R1 70B via Ollama"
ollama pull deepseek-r1:70b
echo "==> DeepSeek-R1 70B installed successfully"
ollama list
`,

	"gemma3-4b": `#!/bin/bash
set -e
echo "==> Downloading Gemma3 4B via Ollama"
ollama pull gemma3:4b
echo "==> Gemma3 4B installed successfully"
ollama list
`,

	"gemma3-12b": `#!/bin/bash
set -e
echo "==> Downloading Gemma3 12B via Ollama"
ollama pull gemma3:12b
echo "==> Gemma3 12B installed successfully"
ollama list
`,

	"gemma3-27b": `#!/bin/bash
set -e
echo "==> Downloading Gemma3 27B via Ollama"
ollama pull gemma3:27b
echo "==> Gemma3 27B installed successfully"
ollama list
`,

	"llama-3.2-1b": `#!/bin/bash
set -e
echo "==> Downloading Llama 3.2 1B via Ollama"
ollama pull llama3.2:1b
echo "==> Llama 3.2 1B installed successfully"
ollama list
`,

	"llama-3.2-3b": `#!/bin/bash
set -e
echo "==> Downloading Llama 3.2 3B via Ollama"
ollama pull llama3.2:3b
echo "==> Llama 3.2 3B installed successfully"
ollama list
`,

	"qwen3-coder-30b": `#!/bin/bash
set -e
echo "==> Downloading Qwen3-Coder 30B MoE via Ollama"
ollama pull qwen3-coder:30b
echo "==> Qwen3-Coder 30B installed successfully"
ollama list
`,

	"command-r-7b": `#!/bin/bash
set -e
echo "==> Downloading Command-R 7B via Ollama"
ollama pull command-r:7b
echo "==> Command-R 7B installed successfully"
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
	"flux2": `#!/bin/bash
set -e

echo "==> Installing Flux 2 (FP8) for ComfyUI"

# Create target directory
mkdir -p ~/ComfyUI/models/unet
cd ~/ComfyUI/models/unet

# Check if already downloaded
if [ -f "flux2-fp8.safetensors" ]; then
    echo "Flux 2 (FP8) already installed"
    exit 0
fi

echo "==> Downloading Flux 2 FP8 model with parallel downloads..."
huggingface-cli download black-forest-labs/FLUX.2-fp8 flux2-fp8.safetensors --local-dir . --max-workers 8

echo "==> Flux 2 (FP8) installed successfully"
echo "Model location: ~/ComfyUI/models/unet/flux2-fp8.safetensors"
`,
	"cogvideox-1.5": `#!/bin/bash
set -e
echo "==> Installing CogVideoX 1.5 5B"
pip3 install --upgrade diffusers transformers accelerate
python3 -c "from huggingface_hub import snapshot_download; snapshot_download('THUDM/CogVideoX1.5-5B', local_dir='$HOME/models/cogvideox-1.5')"
echo "==> CogVideoX 1.5 5B installed successfully"
`,
	"cogvideox-i2v": `#!/bin/bash
set -e
echo "==> Installing CogVideoX 1.5 I2V"
pip3 install --upgrade diffusers transformers accelerate
python3 -c "from huggingface_hub import snapshot_download; snapshot_download('THUDM/CogVideoX1.5-5B-I2V', local_dir='$HOME/models/cogvideox-i2v')"
echo "==> CogVideoX 1.5 I2V installed successfully"
`,
	"hunyuan-video": `#!/bin/bash
set -e
echo "==> Installing HunyuanVideo"
pip3 install --upgrade diffusers transformers accelerate
python3 -c "from huggingface_hub import snapshot_download; snapshot_download('tencent/HunyuanVideo', local_dir='$HOME/models/hunyuan-video')"
echo "==> HunyuanVideo installed successfully"
`,
	"pyramid-flow": `#!/bin/bash
set -e
echo "==> Installing Pyramid Flow"
pip3 install --upgrade diffusers transformers accelerate
python3 -c "from huggingface_hub import snapshot_download; snapshot_download('rain1011/pyramid-flow-miniflux', local_dir='$HOME/models/pyramid-flow')"
echo "==> Pyramid Flow installed successfully"
`,
	"svd-xt": `#!/bin/bash
set -e
echo "==> Installing SVD-XT 1.1 for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download stabilityai/stable-video-diffusion-img2vid-xt-1-1 --local-dir svd-xt --max-workers 8
echo "==> SVD-XT 1.1 installed successfully"
`,
	"i2v-adapter": `#!/bin/bash
set -e
echo "==> Installing I2V-Adapter"
mkdir -p ~/ComfyUI/custom_nodes
cd ~/ComfyUI/custom_nodes
if [ ! -d "I2V-Adapter" ]; then
    git clone https://github.com/KlingTeam/I2V-Adapter.git
fi
cd I2V-Adapter && pip3 install -r requirements.txt
echo "==> I2V-Adapter installed successfully"
`,
	"sd3.5-large": `#!/bin/bash
set -e
echo "==> Installing Stable Diffusion 3.5 Large for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download stabilityai/stable-diffusion-3.5-large --local-dir sd3.5-large --max-workers 8
echo "==> SD 3.5 Large installed successfully"
`,
	"sd3.5-large-turbo": `#!/bin/bash
set -e
echo "==> Installing SD 3.5 Large Turbo for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download stabilityai/stable-diffusion-3.5-large-turbo --local-dir sd3.5-large-turbo --max-workers 8
echo "==> SD 3.5 Large Turbo installed successfully"
`,
	"sd3.5-medium": `#!/bin/bash
set -e
echo "==> Installing Stable Diffusion 3.5 Medium for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download stabilityai/stable-diffusion-3.5-medium --local-dir sd3.5-medium --max-workers 8
echo "==> SD 3.5 Medium installed successfully"
`,
	"sdxl-turbo": `#!/bin/bash
set -e
echo "==> Installing SDXL Turbo for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download stabilityai/sdxl-turbo --local-dir sdxl-turbo --max-workers 8
echo "==> SDXL Turbo installed successfully"
`,
	"sdxl-lightning": `#!/bin/bash
set -e
echo "==> Installing SDXL Lightning for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download ByteDance/SDXL-Lightning --local-dir sdxl-lightning --max-workers 8
echo "==> SDXL Lightning installed successfully"
`,
	"playground-v2.5": `#!/bin/bash
set -e
echo "==> Installing Playground v2.5 for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download playgroundai/playground-v2.5-1024px-aesthetic --local-dir playground-v2.5 --max-workers 8
echo "==> Playground v2.5 installed successfully"
`,
	"pixart-sigma": `#!/bin/bash
set -e
echo "==> Installing PixArt-Σ for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download PixArt-alpha/PixArt-Sigma-XL-2-1024-MS --local-dir pixart-sigma --max-workers 8
echo "==> PixArt-Σ installed successfully"
`,
	"kandinsky-3": `#!/bin/bash
set -e
echo "==> Installing Kandinsky 3"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download ai-forever/Kandinsky3.1 --local-dir kandinsky-3 --max-workers 8
echo "==> Kandinsky 3 installed successfully"
`,
	"kolors": `#!/bin/bash
set -e
echo "==> Installing Kolors for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download Kwai-Kolors/Kolors --local-dir kolors --max-workers 8
echo "==> Kolors installed successfully"
`,
	"real-esrgan": `#!/bin/bash
set -e
echo "==> Installing Real-ESRGAN for ComfyUI"
mkdir -p ~/ComfyUI/models/upscale_models
cd ~/ComfyUI/models/upscale_models
wget -nc https://github.com/xinntao/Real-ESRGAN/releases/download/v0.1.0/RealESRGAN_x4plus.pth || true
wget -nc https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.2.4/RealESRGAN_x4plus_anime_6B.pth || true
echo "==> Real-ESRGAN installed successfully"
`,
	"gfpgan": `#!/bin/bash
set -e
echo "==> Installing GFPGAN for ComfyUI"
mkdir -p ~/ComfyUI/models/facerestore_models
cd ~/ComfyUI/models/facerestore_models
wget -nc https://github.com/TencentARC/GFPGAN/releases/download/v1.3.0/GFPGANv1.4.pth || true
echo "==> GFPGAN installed successfully"
`,
	"aurasr": `#!/bin/bash
set -e
echo "==> Installing AuraSR for ComfyUI"
mkdir -p ~/ComfyUI/models/upscale_models
cd ~/ComfyUI/models/upscale_models
huggingface-cli download fal/AuraSR --local-dir aurasr --max-workers 8
echo "==> AuraSR installed successfully"
`,
	"supir": `#!/bin/bash
set -e
echo "==> Installing SUPIR for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download Kijai/SUPIR_pruned --local-dir supir --max-workers 8
echo "==> SUPIR installed successfully"
`,
	"rife": `#!/bin/bash
set -e
echo "==> Installing RIFE for ComfyUI"
mkdir -p ~/ComfyUI/custom_nodes
cd ~/ComfyUI/custom_nodes
if [ ! -d "ComfyUI-Frame-Interpolation" ]; then
    git clone https://github.com/Fannovel16/ComfyUI-Frame-Interpolation.git
fi
cd ComfyUI-Frame-Interpolation && pip3 install -r requirements.txt
python3 install.py
echo "==> RIFE installed successfully"
`,
	"film": `#!/bin/bash
set -e
echo "==> Installing FILM for ComfyUI"
mkdir -p ~/ComfyUI/custom_nodes
cd ~/ComfyUI/custom_nodes
if [ ! -d "ComfyUI-Frame-Interpolation" ]; then
    git clone https://github.com/Fannovel16/ComfyUI-Frame-Interpolation.git
fi
cd ComfyUI-Frame-Interpolation && pip3 install -r requirements.txt
python3 install.py
echo "==> FILM installed successfully"
`,
	"sd-inpainting": `#!/bin/bash
set -e
echo "==> Installing SD 1.5 Inpainting for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download runwayml/stable-diffusion-inpainting --local-dir sd-inpainting --max-workers 8
echo "==> SD 1.5 Inpainting installed successfully"
`,
	"sdxl-inpainting": `#!/bin/bash
set -e
echo "==> Installing SDXL Inpainting for ComfyUI"
mkdir -p ~/ComfyUI/models/checkpoints
cd ~/ComfyUI/models/checkpoints
huggingface-cli download diffusers/stable-diffusion-xl-1.0-inpainting-0.1 --local-dir sdxl-inpainting --max-workers 8
echo "==> SDXL Inpainting installed successfully"
`,
	"controlnet-canny": `#!/bin/bash
set -e
echo "==> Installing ControlNet Canny for ComfyUI"
mkdir -p ~/ComfyUI/models/controlnet
cd ~/ComfyUI/models/controlnet
huggingface-cli download lllyasviel/sd-controlnet-canny --local-dir controlnet-canny --max-workers 8
echo "==> ControlNet Canny installed successfully"
`,
	"controlnet-depth": `#!/bin/bash
set -e
echo "==> Installing ControlNet Depth for ComfyUI"
mkdir -p ~/ComfyUI/models/controlnet
cd ~/ComfyUI/models/controlnet
huggingface-cli download lllyasviel/sd-controlnet-depth --local-dir controlnet-depth --max-workers 8
echo "==> ControlNet Depth installed successfully"
`,
	"controlnet-openpose": `#!/bin/bash
set -e
echo "==> Installing ControlNet OpenPose for ComfyUI"
mkdir -p ~/ComfyUI/models/controlnet
cd ~/ComfyUI/models/controlnet
huggingface-cli download lllyasviel/sd-controlnet-openpose --local-dir controlnet-openpose --max-workers 8
echo "==> ControlNet OpenPose installed successfully"
`,
	"ip-adapter": `#!/bin/bash
set -e
echo "==> Installing IP-Adapter for ComfyUI"
mkdir -p ~/ComfyUI/models/ipadapter
cd ~/ComfyUI/models/ipadapter
huggingface-cli download h94/IP-Adapter --local-dir ip-adapter --max-workers 8
echo "==> IP-Adapter installed successfully"
`,
	"ip-adapter-faceid": `#!/bin/bash
set -e
echo "==> Installing IP-Adapter FaceID for ComfyUI"
mkdir -p ~/ComfyUI/models/ipadapter
cd ~/ComfyUI/models/ipadapter
huggingface-cli download h94/IP-Adapter-FaceID --local-dir ip-adapter-faceid --max-workers 8
echo "==> IP-Adapter FaceID installed successfully"
`,
	"instantid": `#!/bin/bash
set -e
echo "==> Installing InstantID for ComfyUI"
mkdir -p ~/ComfyUI/models/instantid
cd ~/ComfyUI/models/instantid
huggingface-cli download InstantX/InstantID --local-dir instantid --max-workers 8
echo "==> InstantID installed successfully"
`,
}

