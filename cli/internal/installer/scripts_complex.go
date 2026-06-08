package installer

// complexScripts contains installation scripts that have enough conditional
// logic, error handling, or CUDA-aware behavior that converting them to a
// Go struct generator would lose clarity without gaining safety.
//
// The shared bash fragment `bashWaitForDpkg` (defined in scriptgen.go) is
// prepended to scripts that need it, eliminating the 4x copy-paste.
var complexScripts = map[string]string{

	"core": bashPreamble + bashWaitForDpkg + bashFixBrokenPackages + `
fix_broken_packages

echo "==> Installing Core System (Essential build tools only)"
wait_for_dpkg
sudo apt update
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

PIP_EXTRA=""
if pip3 install --help 2>&1 | grep -q "break-system-packages"; then
    PIP_EXTRA="--break-system-packages"
fi

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

    if ! python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
        echo "==> ERROR: PyTorch $TORCH_VERSION is installed but CUDA is unavailable."
        echo "==> A prior pip install likely replaced torch with a wheel built for the wrong CUDA."
        echo "==> Recovery:"
        echo "==>   pip uninstall -y torch torchvision torchaudio xformers"
        echo "==>   anime install pytorch"
        exit 1
    fi

    echo "==> PyTorch $TORCH_VERSION already installed (CUDA OK)"

    CONSTRAINT_FILE="$HOME/.config/anime/torch-constraints.txt"
    mkdir -p "$(dirname "$CONSTRAINT_FILE")"
    {
        echo "torch==$TORCH_VERSION"
        echo "torchvision"
        echo "torchaudio"
    } > "$CONSTRAINT_FILE"
    export PIP_CONSTRAINT="$CONSTRAINT_FILE"
    echo "==> Pinned torch via $CONSTRAINT_FILE"

    pip3 install --upgrade-strategy only-if-needed --no-deps xformers
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python

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

NEEDS_SOURCE_BUILD=false
VLLM_PIN=""
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    TORCH_CUDA_MAJOR=$(echo "$TORCH_CUDA" | cut -d. -f1)
    if [ "$TORCH_CUDA_MAJOR" = "12" ]; then
        NEEDS_SOURCE_BUILD=true
        VLLM_PIN="==0.10.1.1"
        echo "==> aarch64 + CUDA 12 detected — source-building vllm 0.10.1.1 (last cu12-compatible release)"

        TORCH_MAJOR_MINOR=$(echo "$TORCH_VERSION" | cut -d. -f1,2)
        if [ "$TORCH_MAJOR_MINOR" != "2.7" ]; then
            echo "==> Installing cu128 torch==2.7.1 for vllm 0.10.x compatibility..."
            pip3 install --upgrade-strategy only-if-needed \
                --index-url https://download.pytorch.org/whl/cu128 \
                torch==2.7.1 torchvision==0.22.1 torchaudio==2.7.1
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

if ! python3 -c "import torch; assert torch.cuda.is_available(), 'replaced'" 2>/dev/null; then
    echo "==> FATAL: torch was replaced despite PIP_CONSTRAINT. CUDA broken."
    echo "==> Recovery:"
    echo "==>   rm -rf ~/.local/lib/python3.10/site-packages/{torch*,torchvision*,torchaudio*,nvidia*,triton*,cuda_*,xformers*,vllm*}"
    echo "==>   anime install pytorch && anime install vllm"
    exit 1
fi

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

	"comfyui": `#!/bin/bash
set -e
echo "==> Installing ComfyUI"

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
grep -v "^torch" "$COMFYUI_DIR/requirements.txt" > /tmp/comfyui-requirements-filtered.txt || true

if [ -s /tmp/comfyui-requirements-filtered.txt ]; then
    pip3 install -r /tmp/comfyui-requirements-filtered.txt --upgrade-strategy only-if-needed
fi
rm -f /tmp/comfyui-requirements-filtered.txt

echo "Installing ComfyUI Manager..."
git clone https://github.com/ltdrdata/ComfyUI-Manager.git "$COMFYUI_DIR/custom_nodes/ComfyUI-Manager"

echo "==> ComfyUI installed successfully"
echo "==> PyTorch and CUDA installation preserved"
`,

	"nvidia": bashPreamble + bashWaitForDpkg + `
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

	"docker": bashPreamble + bashWaitForDpkg + `
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

	"nodejs": bashPreamble + bashWaitForDpkg + `
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

GO_VERSION="1.23.5"
GO_ARCH="amd64"

if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
    GO_ARCH="arm64"
fi

echo "==> Downloading Go $GO_VERSION for $GO_ARCH"
wget -q https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz -O /tmp/go.tar.gz

echo "==> Installing Go to /usr/local"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

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

if command -v claude &> /dev/null; then
    echo "Claude Code already installed: $(which claude)"
    exit 0
fi

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

if [ -z "$NPM" ]; then
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"
    command -v npm &> /dev/null && NPM="npm"
fi

if [ -z "$NPM" ]; then
    NPM=$(ls -1 $HOME/.nvm/versions/node/*/bin/npm $HOME/.local/share/nvm/*/bin/npm 2>/dev/null | tail -1)
fi

if [ -z "$NPM" ]; then
    echo "Error: npm not found. Install nodejs first: anime install nodejs"
    exit 1
fi

echo "==> Using npm at: $NPM"

$NPM install -g @anthropic-ai/claude-code 2>/dev/null || sudo $NPM install -g @anthropic-ai/claude-code

echo "==> Claude Code installed successfully"
command -v claude && echo "==> Verified: $(which claude)" || echo "Note: Restart your shell to pick up claude in PATH"
`,

	"gh": `#!/bin/bash
set -e
echo "==> Installing GitHub CLI"
if command -v gh &> /dev/null; then
    echo "GitHub CLI $(gh --version | head -1) already installed"
    exit 0
fi

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

	"comfyui-wan2": `#!/bin/bash
set -e
echo "==> Installing ComfyUI Wan2 Wrapper"

if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

if [ ! -d "$HOME/video-models/wan2" ]; then
    echo "Error: Wan2.2 not found. Please install 'wan2' package first: anime install wan2"
    exit 1
fi

cd ~/ComfyUI/custom_nodes

if [ -d "ComfyUI-WanWrapper" ]; then
    echo "ComfyUI Wan2 Wrapper already installed"
    exit 0
fi

echo "==> Cloning ComfyUI-WanWrapper..."
git clone https://github.com/kijai/ComfyUI-WanWrapper

cd ComfyUI-WanWrapper

if [ -f "requirements.txt" ]; then
    echo "==> Installing wrapper requirements (excluding torch/cuda to avoid conflicts)..."
    grep -v -E "^torch|^nvidia-|^triton" requirements.txt > /tmp/comfyui-wan2-requirements-filtered.txt || true
    if [ -s /tmp/comfyui-wan2-requirements-filtered.txt ]; then
        pip3 install -r /tmp/comfyui-wan2-requirements-filtered.txt --upgrade-strategy only-if-needed
    else
        echo "==> No additional requirements needed (torch/cuda already installed)"
    fi
fi

if [ ! -L "models" ] && [ -d "$HOME/video-models/wan2" ]; then
    echo "==> Linking Wan2 models to ComfyUI..."
    ln -s "$HOME/video-models/wan2/checkpoints" models
fi

echo "==> ComfyUI Wan2 Wrapper installed successfully"
echo "Location: ~/ComfyUI/custom_nodes/ComfyUI-WanWrapper"
echo "Restart ComfyUI to load the new nodes"
`,
}
