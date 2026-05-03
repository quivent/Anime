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

# ============================================================
# Detect environment type
# ============================================================
IS_LAMBDA_STACK=false
IS_GPU_BASE=false

if dpkg -l | grep -q "python3-torch-cuda" || dpkg -l | grep -q "python3-tensorflow"; then
    IS_LAMBDA_STACK=true
    echo "==> Lambda Stack detected - will clean up conflicts"
else
    IS_GPU_BASE=true
    echo "==> GPU Base / Clean environment detected"
fi

# ============================================================
# Pick the right PyTorch wheel index for THIS host.
# ARM64 + CUDA-13 driver (GH200/H200) → cu130 nightly directly,
# avoiding the wasteful "install cu128 then have wantorch swap it" flow.
# Everything else stays on cu128 stable.
# ============================================================
ARCH=$(uname -m)
TORCH_INDEX="https://download.pytorch.org/whl/cu128"
TORCH_FLAVOR="cu128 stable"
PIP_PRE_FLAG=""
DRIVER_MAJOR=""
if command -v nvidia-smi >/dev/null 2>&1; then
    DRIVER_MAJOR=$(nvidia-smi --query-gpu=driver_version --format=csv,noheader 2>/dev/null | head -1 | cut -d. -f1 || echo "")
    CUDA_DRV=$(nvidia-smi --query-gpu=cuda_version --format=csv,noheader 2>/dev/null | head -1 || echo "")
fi
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    # GH200/H200 — prefer cu130 nightly so the comfy_kitchen.cuda backend lights up.
    if [ "${CUDA_DRV%%.*}" = "13" ] || [ -n "${FORCE_CU130:-}" ]; then
        TORCH_INDEX="https://download.pytorch.org/whl/nightly/cu130"
        TORCH_FLAVOR="cu130 nightly (ARM64 + CUDA13 detected)"
        PIP_PRE_FLAG="--pre"
    fi
fi
echo "==> torch flavor: $TORCH_FLAVOR"

# ============================================================
# GPU BASE FAST PATH - No TensorFlow to remove
# ============================================================
if [ "$IS_GPU_BASE" = true ]; then
    if python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
        TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)" 2>/dev/null)
        TORCH_CUDA=$(python3 -c "import torch; print(torch.version.cuda or '')" 2>/dev/null)
        echo "==> PyTorch $TORCH_VERSION (cuda=$TORCH_CUDA) already installed"
    else
        echo "==> Installing PyTorch ($TORCH_FLAVOR)..."
        pip3 install $PIP_PRE_FLAG torch torchvision torchaudio --index-url "$TORCH_INDEX"
    fi

    echo "==> Installing AI libraries..."
    pip3 install transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python

    python3 -c "import torch; print(f'PyTorch {torch.__version__} | CUDA: {torch.cuda.is_available()}')"
    echo "==> PyTorch installed successfully (GPU Base fast path)"
    exit 0
fi

# ============================================================
# LAMBDA STACK PATH - Requires TensorFlow cleanup
# ============================================================
remove_tensorflow() {
    echo "==> [Pre-flight] Removing TensorFlow to prevent dependency conflicts..."

    # Remove all TensorFlow packages via pip (user installs)
    pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu tensorflow-intel tensorflow-macos \
        tensorflow-io tensorflow-io-gcs-filesystem tf-keras keras 2>/dev/null || true

    # Remove system-wide pip installs
    sudo pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu 2>/dev/null || true

    # Remove apt-installed tensorflow (common on Lambda/Ubuntu)
    sudo apt-get remove -y python3-tensorflow 2>/dev/null || true

    # Remove broken ml_dtypes compiled against old numpy
    pip3 uninstall -y ml_dtypes 2>/dev/null || true
    sudo pip3 uninstall -y ml_dtypes 2>/dev/null || true

    # Set environment variables to prevent TensorFlow from being loaded
    export TRANSFORMERS_NO_TF=1
    export USE_TF=0
    export USE_TORCH=1
    export TF_CPP_MIN_LOG_LEVEL=3

    # Persist these settings (only add if not already present)
    if ! grep -q "TRANSFORMERS_NO_TF" ~/.bashrc 2>/dev/null; then
        cat >> ~/.bashrc << 'TFENV'

# Disable TensorFlow to prevent NumPy conflicts with PyTorch/vLLM
export TRANSFORMERS_NO_TF=1
export USE_TF=0
export USE_TORCH=1
export TF_CPP_MIN_LOG_LEVEL=3
TFENV
    fi

    echo "    ✓ TensorFlow removed"
}

# Execute TensorFlow removal for Lambda Stack
remove_tensorflow

# ============================================================
# Install PyTorch (Lambda Stack path)
# ============================================================
if python3 -c "import torch" 2>/dev/null; then
    TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)" 2>/dev/null)
    echo "==> PyTorch $TORCH_VERSION already installed"
    echo "==> Installing AI libraries (preserving torch version)..."
    pip3 install --no-deps accelerate
    pip3 install --upgrade-strategy only-if-needed \
        transformers diffusers safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python
else
    echo "==> Installing PyTorch ($TORCH_FLAVOR)..."
    pip3 install $PIP_PRE_FLAG torch torchvision torchaudio --index-url "$TORCH_INDEX"
    pip3 install transformers diffusers accelerate safetensors bitsandbytes \
        numpy scipy pandas matplotlib pillow opencv-python
fi

# Install fresh ml_dtypes compatible with current NumPy
pip3 install --upgrade ml_dtypes 2>/dev/null || true

python3 -c "import torch; print(f'PyTorch {torch.__version__} | CUDA: {torch.cuda.is_available()}')"
`,

	"flash-attn": `#!/bin/bash
set -e
echo "==> Installing Flash Attention"

# Verify PyTorch is installed with CUDA
if ! python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
    echo "Error: PyTorch with CUDA not found. Please install 'pytorch' package first: anime install pytorch"
    exit 1
fi

# Check if flash-attn is already installed
if python3 -c "import flash_attn" 2>/dev/null; then
    FA_VERSION=$(python3 -c "import flash_attn; print(flash_attn.__version__)" 2>/dev/null || echo "unknown")
    echo "==> Flash Attention $FA_VERSION already installed"
    exit 0
fi

echo "==> Installing Flash Attention (this may take 10+ minutes to compile)..."
echo "==> Note: Compilation requires significant CPU and memory resources"

# Install flash-attn with --no-build-isolation to use existing torch/cuda
pip3 install flash-attn --no-build-isolation

# Verify installation
echo "==> Verifying Flash Attention installation..."
python3 -c "import flash_attn; print(f'Flash Attention {flash_attn.__version__} installed successfully')"

echo "==> Flash Attention installed successfully"
echo ""
echo "Usage: from flash_attn import flash_attn_func, flash_attn_qkvpacked_func"
echo "Documentation: https://github.com/Dao-AILab/flash-attention"
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

# ============================================================
# STEP 1: Detect environment type
# ============================================================
echo "==> [1/6] Detecting environment..."

ARCH=$(uname -m)
IS_LAMBDA_STACK=false
IS_GPU_BASE=false
TORCH_VERSION=""
HAS_SYSTEM_TORCH=false
HAS_SYSTEM_TF=false

# Check for Lambda Stack indicators
if dpkg -l | grep -q "python3-torch-cuda"; then
    HAS_SYSTEM_TORCH=true
fi
if dpkg -l | grep -q "python3-tensorflow"; then
    HAS_SYSTEM_TF=true
fi

# Determine environment type
if [ "$HAS_SYSTEM_TORCH" = true ] || [ "$HAS_SYSTEM_TF" = true ]; then
    IS_LAMBDA_STACK=true
    echo "    ⚠ Lambda Stack detected (system ML packages present)"
    echo "    → Will clean up conflicting packages"
else
    IS_GPU_BASE=true
    echo "    ✓ GPU Base / Clean environment detected"
    echo "    → Using streamlined install path"
fi

# ============================================================
# GPU BASE FAST PATH - No conflicts to clean up
# ============================================================
if [ "$IS_GPU_BASE" = true ]; then
    echo "==> [2/6] GPU Base: Checking PyTorch..."

    # Install PyTorch if not present
    if ! python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
        echo "    → Installing PyTorch with CUDA 12.8..."
        pip3 install torch --index-url https://download.pytorch.org/whl/cu128
    fi
    TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__)" 2>/dev/null)
    echo "    ✓ PyTorch $TORCH_VERSION with CUDA ready"

    echo "==> [3/6] GPU Base: Skipping cleanup (no conflicts)"
    echo "==> [4/6] GPU Base: Skipping NumPy fix (no system constraints)"

    echo "==> [5/6] GPU Base: Installing vLLM..."
    pip3 install vllm

    echo "==> [6/6] Verifying installation..."
    if python3 -c "from vllm import LLM; print('vLLM OK')" 2>/dev/null; then
        VLLM_VERSION=$(python3 -c "import vllm; print(vllm.__version__)" 2>/dev/null)
        echo "    ✓ vLLM $VLLM_VERSION installed successfully"
        python3 -c "import torch; print(f'    ✓ PyTorch {torch.__version__} | CUDA: {torch.cuda.is_available()}')"
        echo ""
        echo "==> vLLM installed successfully (GPU Base fast path)"
        exit 0
    else
        echo "    ✗ vLLM verification failed"
        echo "    Run: anime vllm doctor --fix"
        exit 1
    fi
fi

# ============================================================
# LAMBDA STACK PATH - Requires cleanup
# ============================================================
echo "==> [1/6] Lambda Stack: Continuing with cleanup path..."

# Check PyTorch CUDA
if python3 -c "import torch; assert torch.cuda.is_available()" 2>/dev/null; then
    TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__.split('+')[0])" 2>/dev/null)
    echo "    ✓ PyTorch $TORCH_VERSION with CUDA available"
else
    echo "    ✗ PyTorch with CUDA not found"
    if [ "$HAS_SYSTEM_TORCH" = true ]; then
        echo ""
        echo "ERROR: System PyTorch CUDA not working."
        echo "Try: pip3 uninstall torch && python3 -c 'import torch; print(torch.cuda.is_available())'"
    else
        echo ""
        echo "ERROR: Please install PyTorch first: anime install pytorch"
    fi
    exit 1
fi

# ============================================================
# STEP 2: Check if vLLM is already installed and working
# ============================================================
echo "==> [2/6] Checking existing vLLM installation..."
if python3 -c "from vllm import LLM" 2>/dev/null; then
    VLLM_VERSION=$(python3 -c "import vllm; print(vllm.__version__)" 2>/dev/null || echo "unknown")
    echo "    ✓ vLLM $VLLM_VERSION already installed and working"
    exit 0
fi
echo "    → vLLM not found, proceeding with installation"

# ============================================================
# STEP 3: Remove conflicting packages
# ============================================================
echo "==> [3/6] Cleaning up conflicting packages..."

# Remove TensorFlow (causes NumPy conflicts with vLLM/transformers)
pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu tensorflow-intel \
    tensorflow-io tf-keras keras 2>/dev/null || true
sudo apt-get remove -y python3-tensorflow 2>/dev/null || true
# Force remove stale TensorFlow directories (apt leaves these behind)
sudo rm -rf /usr/lib/python3/dist-packages/tensorflow* 2>/dev/null || true

# Remove pip-installed torch if system torch is available (prevents override)
if [ "$HAS_SYSTEM_TORCH" = true ]; then
    echo "    → Removing pip torch to use system torch-cuda..."
    pip3 uninstall -y torch torchvision torchaudio 2>/dev/null || true
fi

# Set TensorFlow prevention env vars
export TRANSFORMERS_NO_TF=1
export USE_TF=0
export USE_TORCH=1

if ! grep -q "TRANSFORMERS_NO_TF" ~/.bashrc 2>/dev/null; then
    echo -e "\n# Disable TensorFlow\nexport TRANSFORMERS_NO_TF=1\nexport USE_TF=0\nexport USE_TORCH=1" >> ~/.bashrc
fi
echo "    ✓ Conflicting packages removed"

# ============================================================
# STEP 4: NumPy check (vLLM 0.13.0+ works with NumPy 2.x)
# ============================================================
echo "==> [4/6] Checking NumPy..."
# Note: vLLM 0.13.0+ works with NumPy 2.x, no downgrade needed
echo "    ✓ NumPy OK (vLLM 0.13.0+ compatible with NumPy 2.x)"

# ============================================================
# STEP 5: Install vLLM (version depends on environment)
# ============================================================
echo "==> [5/6] Installing vLLM..."

# Re-check torch version after cleanup
TORCH_VERSION=$(python3 -c "import torch; print(torch.__version__.split('+')[0])" 2>/dev/null || echo "2.9.0")
TORCH_MAJOR_MINOR=$(echo $TORCH_VERSION | cut -d. -f1,2)

echo "    → PyTorch version: $TORCH_VERSION"
echo "    → Architecture: $ARCH"

if [ "$ARCH" = "aarch64" ]; then
    # ARM64/GH200: Use cu128 index for CUDA-enabled PyTorch
    echo "    → ARM64: Installing vLLM + PyTorch 2.9.1+cu128..."

    pip3 install --upgrade pip

    # Install vLLM first (pulls its deps including CPU torch)
    pip3 install vllm

    # Then reinstall torch with CUDA from cu128 wheel index (overwrites CPU torch)
    echo "    → Upgrading to CUDA-enabled PyTorch..."
    pip3 install torch==2.9.1 --index-url https://download.pytorch.org/whl/cu128

    echo "    ✓ vLLM installed with torch+cu128"
else
    # x86_64: use latest vLLM
    echo "    → Using latest vLLM"
    pip3 install vllm
fi
echo "    ✓ vLLM installed"

# ============================================================
# STEP 6: Verify installation with error recovery
# ============================================================
echo "==> [6/6] Verifying vLLM installation..."

# First try: direct import
if python3 -c "from vllm import LLM, SamplingParams; print('vLLM OK')" 2>/dev/null; then
    echo "    ✓ vLLM verification passed"
else
    echo "    ⚠ Initial verification failed, attempting recovery..."

    # Recovery: Check if TensorFlow snuck back in
    if python3 -c "import tensorflow" 2>/dev/null; then
        echo "    → TensorFlow detected again, removing..."
        pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu 2>/dev/null || true
    fi

    # Recovery: Try with explicit environment
    if TRANSFORMERS_NO_TF=1 USE_TF=0 python3 -c "from vllm import LLM, SamplingParams; print('vLLM OK')" 2>/dev/null; then
        echo "    ✓ vLLM works with TF disabled"
    else
        # Last resort: Check the actual error
        echo "    → Diagnosing issue..."
        python3 -c "
import sys
try:
    from vllm import LLM
    print('vLLM imported successfully')
except ImportError as e:
    print(f'Import error: {e}')
    if 'numpy' in str(e).lower():
        print('DIAGNOSIS: NumPy version conflict')
        sys.exit(2)
    elif 'tensorflow' in str(e).lower():
        print('DIAGNOSIS: TensorFlow conflict')
        sys.exit(3)
    else:
        print(f'DIAGNOSIS: Unknown - {e}')
        sys.exit(1)
except Exception as e:
    print(f'Error: {e}')
    sys.exit(1)
" 2>&1 || {
            EXIT_CODE=$?
            if [ $EXIT_CODE -eq 2 ]; then
                echo "    → NumPy conflict detected, reinstalling ml_dtypes..."
                # vLLM 0.13+ supports NumPy 2.x - don't downgrade, fix ml_dtypes instead
                pip3 install --upgrade --force-reinstall ml_dtypes 2>/dev/null || true
            elif [ $EXIT_CODE -eq 3 ]; then
                echo "    → TensorFlow still causing issues, removing via pip..."
                # Safe removal via pip instead of rm -rf
                pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu tensorflow-intel tensorflow-macos 2>/dev/null || true
                pip3 uninstall -y tensorflow-io tensorflow-io-gcs-filesystem tf-keras keras 2>/dev/null || true
                # Also try system pip
                sudo pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu 2>/dev/null || true
                # Remove apt package if present
                sudo apt-get remove -y python3-tensorflow 2>/dev/null || true
            fi

            # Final verification
            if ! python3 -c "from vllm import LLM; print('vLLM OK')" 2>/dev/null; then
                echo "    ✗ vLLM installation failed after recovery attempts"
                echo ""
                echo "Manual fix options:"
                echo "  1. sudo apt-get remove python3-tensorflow"
                echo "  2. pip3 uninstall tensorflow tensorflow-cpu tensorflow-gpu"
                echo "  3. Create a fresh venv: python3 -m venv ~/vllm-env && source ~/vllm-env/bin/activate && pip install vllm"
                exit 1
            fi
        }
    fi
fi

python3 -c "import torch; print(f'torch {torch.__version__} | CUDA: {torch.cuda.is_available()}')"

echo ""
echo "==> vLLM installed successfully"
echo ""
echo "Usage:"
echo "  vllm serve <model>                    # Start OpenAI-compatible server"
echo "  python -m vllm.entrypoints.openai.api_server --model <model>"
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

	"comfy-cli": `#!/bin/bash
set -e
echo "==> Installing comfy-cli (ComfyUI Management CLI)"

# Check if comfy-cli is already installed
if command -v comfy &> /dev/null; then
    COMFY_VERSION=$(comfy --version 2>/dev/null || echo "unknown")
    echo "comfy-cli already installed: $COMFY_VERSION"
    exit 0
fi

# Ensure pipx is installed for isolated Python app installation
if ! command -v pipx &> /dev/null; then
    echo "==> Installing pipx for isolated Python app management..."
    pip3 install --user pipx
    python3 -m pipx ensurepath
    # Add to current session
    export PATH="$PATH:$HOME/.local/bin"
fi

# Install comfy-cli using pipx (creates isolated venv automatically)
echo "==> Installing comfy-cli via pipx (isolated environment)..."
pipx install comfy-cli

# Verify installation
echo "==> Verifying comfy-cli installation..."
if command -v comfy &> /dev/null; then
    comfy --version
    echo "==> comfy-cli installed successfully!"
else
    # Try with explicit path
    if [ -x "$HOME/.local/bin/comfy" ]; then
        echo "==> comfy-cli installed to ~/.local/bin/comfy"
        echo "==> Add ~/.local/bin to your PATH if not already present"
        $HOME/.local/bin/comfy --version
    else
        echo "Error: comfy-cli installation failed"
        exit 1
    fi
fi

echo ""
echo "Quick start commands:"
echo "  comfy install              - Install ComfyUI"
echo "  comfy launch               - Start ComfyUI server"
echo "  comfy node list            - List custom nodes"
echo "  comfy model list           - List installed models"
echo "  comfy --help               - Show all commands"
`,

	"comfyui": `#!/bin/bash
set -euo pipefail
echo "==> Installing ComfyUI (manual venv install, host-aware torch)"

COMFYUI_DIR="$HOME/ComfyUI"
VENV="$COMFYUI_DIR/venv"

if [ -f "$COMFYUI_DIR/main.py" ] && [ -x "$VENV/bin/python" ]; then
    echo "==> ComfyUI already at $COMFYUI_DIR (venv at $VENV)"
    exit 0
fi

# ─── apt deps the rest of this script depends on ──────────────────
# A fresh Lambda / GH200 / H100 instance may be missing git, python3-venv,
# build-essential, or screen — apt-install them up front so cloning the
# repo, creating the venv, building wheels, and (later) running ComfyUI in
# a screen session all just work. We intentionally keep this list small.
need_apt=()
command -v git           >/dev/null 2>&1 || need_apt+=(git)
command -v screen        >/dev/null 2>&1 || need_apt+=(screen)
python3 -c 'import venv' >/dev/null 2>&1 || need_apt+=(python3-venv)
dpkg -s build-essential  >/dev/null 2>&1 || need_apt+=(build-essential)
if [ "${#need_apt[@]}" -gt 0 ]; then
    echo "==> apt-installing: ${need_apt[*]}"
    if [ "$(id -u)" -eq 0 ]; then
        DEBIAN_FRONTEND=noninteractive apt-get update -y
        DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${need_apt[@]}"
    else
        sudo -n true 2>/dev/null || { echo "ERROR: need sudo to apt-install: ${need_apt[*]}"; exit 1; }
        sudo DEBIAN_FRONTEND=noninteractive apt-get update -y
        sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${need_apt[@]}"
    fi
fi

# ─── pick torch wheel index from driver-supported CUDA (matches wantorch) ───
# Same logic as the wantorch phase: install the newest pytorch.org index
# that fits within the driver's max-CUDA. Avoids "install cu128 then have
# wantorch immediately swap to cu130" wasted-bandwidth path on H100/GH200.
ARCH=$(uname -m)
CUDA_DRV=""
if command -v nvidia-smi >/dev/null 2>&1; then
    CUDA_DRV=$(nvidia-smi --query-gpu=cuda_version --format=csv,noheader 2>/dev/null | head -1 | tr -d ' ' || true)
fi
DRV_MAJ="${CUDA_DRV%%.*}"
DRV_MIN=$(echo "${CUDA_DRV#*.}" | cut -d. -f1)
[ -z "$DRV_MAJ" ] && DRV_MAJ=12
[ -z "$DRV_MIN" ] && DRV_MIN=8
TORCH_INDEX="https://download.pytorch.org/whl/cu128"
PIP_PRE=""
TORCH_FLAVOR="cu128 stable"
case "$DRV_MAJ" in
    13) TORCH_INDEX="https://download.pytorch.org/whl/nightly/cu130"; PIP_PRE="--pre"; TORCH_FLAVOR="cu130 nightly" ;;
    12) if [ "$DRV_MIN" -ge 8 ]; then TORCH_FLAVOR="cu128 stable";
        elif [ "$DRV_MIN" -ge 4 ]; then TORCH_INDEX="https://download.pytorch.org/whl/cu124"; TORCH_FLAVOR="cu124 stable";
        else TORCH_INDEX="https://download.pytorch.org/whl/cu121"; TORCH_FLAVOR="cu121 stable"; fi ;;
    11) TORCH_INDEX="https://download.pytorch.org/whl/cu118"; TORCH_FLAVOR="cu118 stable" ;;
esac
[ -n "${FORCE_CU130:-}" ] && { TORCH_INDEX="https://download.pytorch.org/whl/nightly/cu130"; PIP_PRE="--pre"; TORCH_FLAVOR="cu130 nightly (forced)"; }
echo "==> torch flavor: $TORCH_FLAVOR  (driver_cuda=${CUDA_DRV:-unknown}, arch=$ARCH)"

# ─── clone ComfyUI if missing ───
if [ ! -f "$COMFYUI_DIR/main.py" ]; then
    echo "==> Cloning ComfyUI..."
    git clone --depth 1 https://github.com/comfyanonymous/ComfyUI.git "$COMFYUI_DIR"
fi

# ─── create venv at $COMFYUI_DIR/venv (canonical path used by wantorch) ───
if [ ! -x "$VENV/bin/python" ]; then
    echo "==> Creating venv at $VENV..."
    python3 -m venv "$VENV"
fi

"$VENV/bin/pip" install --upgrade -q pip wheel setuptools

echo "==> Installing torch ($TORCH_FLAVOR) into venv..."
"$VENV/bin/pip" install $PIP_PRE torch torchvision torchaudio --index-url "$TORCH_INDEX"

echo "==> Installing ComfyUI requirements..."
"$VENV/bin/pip" install -r "$COMFYUI_DIR/requirements.txt"

# ─── install ComfyUI Manager (so users can browse/install nodes from UI) ───
NODES="$COMFYUI_DIR/custom_nodes"
mkdir -p "$NODES"
if [ ! -d "$NODES/ComfyUI-Manager" ]; then
    echo "==> Installing ComfyUI-Manager..."
    git clone --depth 1 https://github.com/Comfy-Org/ComfyUI-Manager.git "$NODES/ComfyUI-Manager"
fi

# ─── verify ───
"$VENV/bin/python" -c "import torch; print(f'torch={torch.__version__} cuda={torch.version.cuda} device={torch.cuda.get_device_name(0) if torch.cuda.is_available() else \"cpu\"}')"

# ─── launch script ───
cat > "$COMFYUI_DIR/launch.sh" <<'LAUNCH_EOF'
#!/bin/bash
cd "$HOME/ComfyUI"
exec ./venv/bin/python main.py "$@"
LAUNCH_EOF
chmod +x "$COMFYUI_DIR/launch.sh"

echo ""
echo "==> ComfyUI installed at $COMFYUI_DIR (venv: $VENV)"
echo "==> Start with: screen -dmS comfyui bash -c 'cd $COMFYUI_DIR && ./venv/bin/python main.py --listen'"
echo "    or: $COMFYUI_DIR/launch.sh --listen"
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

# Detect architecture
ARCH=$(dpkg --print-architecture)
echo "==> Detected architecture: $ARCH"

# Validate architecture is supported
if [ "$ARCH" != "amd64" ] && [ "$ARCH" != "arm64" ]; then
    echo "Error: Unsupported architecture: $ARCH"
    echo "NVIDIA CUDA only supports amd64 (x86_64) and arm64 (aarch64)"
    exit 1
fi

echo "==> Downloading CUDA keyring for $ARCH..."
wget -q "https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/${ARCH}/cuda-keyring_1.1-1_all.deb" -O /tmp/cuda-keyring.deb
wait_for_dpkg
sudo dpkg -i /tmp/cuda-keyring.deb
rm -f /tmp/cuda-keyring.deb
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

echo "==> Installing Go (latest stable)"

GO_ARCH="amd64"
if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
    GO_ARCH="arm64"
fi

GO_OS="linux"
if [ "$(uname -s)" = "Darwin" ]; then
    GO_OS="darwin"
fi

# Skip if already installed and recent (1.23+)
if command -v go &>/dev/null; then
    CUR=$(go version | sed 's/.*go1\.\([0-9]*\).*/\1/')
    if [ "${CUR:-0}" -ge 23 ]; then
        echo "  ✓ $(go version) — up to date"
        exit 0
    fi
    echo "  → Upgrading from go1.${CUR}"
fi

# Fetch latest stable
GO_VERSION=$(curl -sL 'https://go.dev/VERSION?m=text' | head -1)
echo "  → ${GO_VERSION} ${GO_OS}/${GO_ARCH}"

curl -sLo /tmp/go.tar.gz "https://go.dev/dl/${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm -f /tmp/go.tar.gz

# PATH
for rc in ~/.bashrc ~/.profile; do
    if ! grep -q "/usr/local/go/bin" "$rc" 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> "$rc"
    fi
done
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

echo "  ✓ $(/usr/local/go/bin/go version)"
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

	"gh": `#!/bin/bash
set -e
echo "==> Installing GitHub CLI (gh)"

if command -v gh &> /dev/null; then
    echo "GitHub CLI $(gh --version | head -1) already installed"
    exit 0
fi

# Wait for dpkg lock
wait_for_dpkg() {
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        echo "Waiting for other package managers to finish..."
        sleep 5
    done
}

echo "==> Adding GitHub CLI repository"
wait_for_dpkg

# Install prerequisites
sudo apt install -y curl

# Add GitHub CLI repository
(type -p wget >/dev/null || (sudo apt update && sudo apt-get install wget -y)) \
    && sudo mkdir -p -m 755 /etc/apt/keyrings \
    && out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    && cat $out | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
    && sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null

echo "==> Installing gh package"
wait_for_dpkg
sudo apt update
sudo apt install gh -y

echo "==> GitHub CLI installed successfully"
gh --version

echo ""
echo "To authenticate, run: gh auth login"
`,

	"make": `#!/bin/bash
set -e
echo "==> Installing Make & Build Tools"

# Wait for dpkg lock
wait_for_dpkg() {
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        echo "Waiting for other package managers to finish..."
        sleep 5
    done
}

echo "==> Installing GNU Make, autotools, and build utilities"
wait_for_dpkg
sudo apt update
sudo apt install -y make automake autoconf libtool cmake ninja-build pkg-config

echo "==> Make & Build Tools installed successfully"
make --version
cmake --version
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
pip install --upgrade diffusers transformers accelerate
python -c "from huggingface_hub import snapshot_download; snapshot_download('THUDM/CogVideoX1.5-5B', local_dir='$HOME/models/cogvideox-1.5')"
echo "==> CogVideoX 1.5 5B installed successfully"
`,
	"cogvideox-i2v": `#!/bin/bash
set -e
echo "==> Installing CogVideoX 1.5 I2V"
pip install --upgrade diffusers transformers accelerate
python -c "from huggingface_hub import snapshot_download; snapshot_download('THUDM/CogVideoX1.5-5B-I2V', local_dir='$HOME/models/cogvideox-i2v')"
echo "==> CogVideoX 1.5 I2V installed successfully"
`,
	"hunyuan-video": `#!/bin/bash
set -e
echo "==> Installing HunyuanVideo"
pip install --upgrade diffusers transformers accelerate
python -c "from huggingface_hub import snapshot_download; snapshot_download('tencent/HunyuanVideo', local_dir='$HOME/models/hunyuan-video')"
echo "==> HunyuanVideo installed successfully"
`,
	"pyramid-flow": `#!/bin/bash
set -e
echo "==> Installing Pyramid Flow"
pip install --upgrade diffusers transformers accelerate
python -c "from huggingface_hub import snapshot_download; snapshot_download('rain1011/pyramid-flow-miniflux', local_dir='$HOME/models/pyramid-flow')"
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
cd I2V-Adapter && pip install -r requirements.txt
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
cd ComfyUI-Frame-Interpolation && pip install -r requirements.txt
python install.py
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
cd ComfyUI-Frame-Interpolation && pip install -r requirements.txt
python install.py
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

	// =========================================================================
	// GH200 + Wan 2.2 stack (added — captures tuned May 2026 workflow)
	// =========================================================================

	"wantorch": `#!/bin/bash
set -euo pipefail
echo "==> Installing PyTorch + sage attention into ComfyUI venv (driver-aware)"

VENV="$HOME/ComfyUI/venv"
if [ ! -x "$VENV/bin/python" ]; then
    echo "ERROR: $VENV/bin/python not found. Install 'comfyui' first."
    exit 1
fi

ARCH=$(uname -m)

# ─── pick torch wheel index from driver-supported CUDA ──────────────
# We do NOT pin a single CUDA version. Wan 2.2 + sageattention work on
# any modern PyTorch (>=2.4) with CUDA support; the only constraint is
# that the wheel must match what the host's NVIDIA driver supports.
# nvidia-smi reports the maximum CUDA the driver can run; we pick the
# newest pytorch.org index that fits within it.
DRIVER_CUDA=""
if command -v nvidia-smi >/dev/null 2>&1; then
    DRIVER_CUDA=$(nvidia-smi --query-gpu=cuda_version --format=csv,noheader 2>/dev/null | head -1 | tr -d ' ' || true)
fi
DRIVER_MAJOR="${DRIVER_CUDA%%.*}"
DRIVER_MINOR=$(echo "${DRIVER_CUDA#*.}" | cut -d. -f1)
[ -z "$DRIVER_MAJOR" ] && DRIVER_MAJOR=12
[ -z "$DRIVER_MINOR" ] && DRIVER_MINOR=8

TORCH_INDEX=""
PIP_PRE=""
TORCH_LABEL=""
case "$DRIVER_MAJOR" in
    13)
        TORCH_INDEX="https://download.pytorch.org/whl/nightly/cu130"
        PIP_PRE="--pre"
        TORCH_LABEL="cu130 nightly (driver supports CUDA 13.x)"
        ;;
    12)
        if [ "$DRIVER_MINOR" -ge 8 ]; then
            TORCH_INDEX="https://download.pytorch.org/whl/cu128"
            TORCH_LABEL="cu128 stable (driver supports CUDA 12.8+)"
        elif [ "$DRIVER_MINOR" -ge 4 ]; then
            TORCH_INDEX="https://download.pytorch.org/whl/cu124"
            TORCH_LABEL="cu124 stable (driver supports CUDA 12.4-12.7)"
        else
            TORCH_INDEX="https://download.pytorch.org/whl/cu121"
            TORCH_LABEL="cu121 stable (driver supports CUDA 12.1-12.3)"
        fi
        ;;
    11)
        TORCH_INDEX="https://download.pytorch.org/whl/cu118"
        TORCH_LABEL="cu118 stable (driver supports CUDA 11.x)"
        ;;
    *)
        TORCH_INDEX="https://download.pytorch.org/whl/cu128"
        TORCH_LABEL="cu128 stable (no driver detected; safe default)"
        ;;
esac
echo "==> arch=$ARCH  driver_cuda=${DRIVER_CUDA:-unknown}  → $TORCH_LABEL"

# Optional override: anyone wanting a specific index can set in env.
if [ -n "${WAN_TORCH_INDEX:-}" ]; then
    TORCH_INDEX="$WAN_TORCH_INDEX"
    PIP_PRE="${WAN_TORCH_PRE:-}"
    TORCH_LABEL="user-pinned via WAN_TORCH_INDEX=$TORCH_INDEX"
    echo "==> override: $TORCH_LABEL"
fi

# ─── skip swap if torch + sage already work ─────────────────────────
WORKING=$("$VENV/bin/python" - <<'PY' 2>/dev/null || true
try:
    import torch, sageattention  # noqa: F401
    if torch.cuda.is_available():
        print(f"ok cu{torch.version.cuda}")
except Exception:
    pass
PY
)
if [ -n "$WORKING" ] && [ -z "${WAN_TORCH_FORCE:-}" ]; then
    echo "==> $WORKING — skipping torch swap (set WAN_TORCH_FORCE=1 to override)"
    "$VENV/bin/pip" show hf_transfer >/dev/null 2>&1 || "$VENV/bin/pip" install -q hf_transfer
    "$VENV/bin/python" -c "import torch; print(f'torch={torch.__version__}  cuda={torch.version.cuda}  device={torch.cuda.get_device_name(0)}')"
    exit 0
fi

echo "==> Installing torch from $TORCH_INDEX (drops in ~3GB of wheels)"
"$VENV/bin/pip" install $PIP_PRE --upgrade --force-reinstall \
    torch torchvision torchaudio \
    --index-url "$TORCH_INDEX"

echo "==> Installing hf_transfer (fast HF downloads) and sageattention"
"$VENV/bin/pip" install -q hf_transfer sageattention

# Verify
"$VENV/bin/python" - <<'PY'
import torch
assert torch.cuda.is_available(), "CUDA not available after install"
print(f"torch={torch.__version__}  cuda={torch.version.cuda}  device={torch.cuda.get_device_name(0)}")
try:
    from importlib.metadata import version as _v
    print(f"sageattention={_v('sageattention')}")
except Exception:
    import sageattention as _s  # noqa: F401
    print("sageattention=installed")
PY

echo ""
echo "==> Done. Start ComfyUI with sage attention engaged:"
echo "    cd ~/ComfyUI && ./venv/bin/python main.py --listen --use-sage-attention"
`,

	"wannodes": `#!/bin/bash
set -euo pipefail
echo "==> Installing Kijai's Wan custom-node stack into ComfyUI"

VENV="$HOME/ComfyUI/venv"
NODES="$HOME/ComfyUI/custom_nodes"

if [ ! -d "$HOME/ComfyUI" ]; then
    echo "ERROR: ~/ComfyUI not found. Install 'comfyui' first."
    exit 1
fi
mkdir -p "$NODES"

clone_or_pull() {
    local url=$1 name=$2
    if [ -d "$NODES/$name/.git" ]; then
        echo "==> $name already cloned, pulling latest"
        git -C "$NODES/$name" pull --ff-only
    else
        echo "==> Cloning $name"
        git clone --depth 1 "$url" "$NODES/$name"
    fi
}

clone_or_pull https://github.com/kijai/ComfyUI-WanVideoWrapper.git ComfyUI-WanVideoWrapper
clone_or_pull https://github.com/kijai/ComfyUI-KJNodes.git           ComfyUI-KJNodes
clone_or_pull https://github.com/Comfy-Org/ComfyUI-Manager.git       ComfyUI-Manager

echo "==> Installing custom-node Python deps (skipping numpy/torchcodec pins to avoid downgrades)"
"$VENV/bin/pip" install -q \
    ftfy "accelerate>=1.2.1" einops "diffusers>=0.33.0" "peft>=0.17.0" \
    "sentencepiece>=0.2.0" protobuf pyloudnorm "gguf>=0.17.1" \
    opencv-python-headless scipy color-matcher matplotlib mss \
    GitPython PyGithub typer rich toml uv chardet "transformers>=4.50.3"

# Drop in the no-LoRA max-quality T2V workflow if we have it embedded
WF_DIR="$HOME/ComfyUI/user/default/workflows"
mkdir -p "$WF_DIR"
WF_FILE="$WF_DIR/Wan2.2_14B_T2V_NoLoRA_MaxQuality.json"
if [ ! -f "$WF_FILE" ]; then
    cat >"$WF_FILE" <<'WFEOF'
{"placeholder":true,"note":"Run gh200-wan-full to install the full workflow JSON, or copy from anime/cli/embedded/workflows/."}
WFEOF
fi

echo "==> Custom-node stack installed. Restart ComfyUI to load."
`,

	"wanmodels": `#!/bin/bash
set -euo pipefail

# WAN_INSTALL_LEVEL controls how much we pull. Set by anime wan studio
# --minimal | --standard | --full (default: full). Higher levels are supersets.
LEVEL="${WAN_INSTALL_LEVEL:-full}"
case "$LEVEL" in
    minimal)  echo "==> Downloading Wan 2.2 minimal set: 5B TI2V + encoder + VAE  (~20GB, fits 12GB VRAM)" ;;
    standard) echo "==> Downloading Wan 2.2 standard set: 14B T2V dual-expert + 4-step LoRAs + encoder + VAE  (~35GB, fits 24GB VRAM)" ;;
    full)     echo "==> Downloading Wan 2.2 full set: T2V+I2V dual-expert + 5B + LoRAs + encoders + VAEs  (~85GB, fits 48GB+ VRAM)" ;;
    *)        echo "ERROR: unknown WAN_INSTALL_LEVEL=$LEVEL (expected: minimal|standard|full)"; exit 1 ;;
esac

VENV="$HOME/ComfyUI/venv"
MODELS="$HOME/ComfyUI/models"

if [ ! -d "$HOME/ComfyUI" ]; then
    echo "ERROR: ~/ComfyUI not found. Install 'comfyui' first."
    exit 1
fi
mkdir -p "$MODELS/diffusion_models" "$MODELS/text_encoders" "$MODELS/vae" "$MODELS/loras"

# hf_transfer dramatically speeds up large downloads; install if missing
"$VENV/bin/pip" show hf_transfer >/dev/null 2>&1 || "$VENV/bin/pip" install -q hf_transfer

export HF_HUB_ENABLE_HF_TRANSFER=1
# Pass HF_TOKEN through if we have one (embedded or env). Auth avoids
# rate limits and makes downloads more reliable.
if [ -z "${HF_TOKEN:-}" ] && command -v anime >/dev/null 2>&1; then
    EMBEDDED_HF=$(anime embed token list 2>/dev/null | grep -i 'hf:' | awk '{print $2}' || true)
    [ -n "$EMBEDDED_HF" ] && export HF_TOKEN="$EMBEDDED_HF"
fi
export HF_TOKEN="${HF_TOKEN:-}"

# Each (repo, file_in_repo, dest_subdir, levels) — flatten to dest_subdir/<basename>.
# 'levels' is a comma-separated list of which install levels include this file.
"$VENV/bin/python" - <<'PY'
import os, shutil, sys, time
from pathlib import Path
from huggingface_hub import hf_hub_download
from huggingface_hub.utils import GatedRepoError, RepositoryNotFoundError, HfHubHTTPError

LEVEL = os.environ.get("WAN_INSTALL_LEVEL", "full")
ROOT = Path(os.path.expanduser("~/ComfyUI/models"))
TOKEN = os.environ.get("HF_TOKEN") or None

# (repo, file_in_repo, dest_subdir, set_of_levels_that_need_it)
ITEMS = [
    # Wan 2.2 dual-expert T2V (standard + full)
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/diffusion_models/wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors", "diffusion_models", {"standard", "full"}),
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/diffusion_models/wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors",  "diffusion_models", {"standard", "full"}),
    # Wan 2.2 dual-expert I2V (full only — saves ~30GB on standard)
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/diffusion_models/wan2.2_i2v_high_noise_14B_fp8_scaled.safetensors", "diffusion_models", {"full"}),
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/diffusion_models/wan2.2_i2v_low_noise_14B_fp8_scaled.safetensors",  "diffusion_models", {"full"}),
    # Wan 2.2 TI2V 5B (minimal + full — small/fast model for low-VRAM hosts)
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/diffusion_models/wan2.2_ti2v_5B_fp16.safetensors", "diffusion_models", {"minimal", "full"}),
    # 4-step lightx2v LoRAs — required for 14B "fast" preset (standard + full)
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_high_noise.safetensors", "loras", {"standard", "full"}),
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_low_noise.safetensors",  "loras", {"standard", "full"}),
    # Text encoder — every level needs it
    ("Comfy-Org/Wan_2.1_ComfyUI_repackaged", "split_files/text_encoders/umt5_xxl_fp8_e4m3fn_scaled.safetensors", "text_encoders", {"minimal", "standard", "full"}),
    # VAEs: 14B uses wan_2.1; 5B uses wan2.2.
    ("Comfy-Org/Wan_2.1_ComfyUI_repackaged", "split_files/vae/wan_2.1_vae.safetensors", "vae", {"standard", "full"}),
    ("Comfy-Org/Wan_2.2_ComfyUI_Repackaged", "split_files/vae/wan2.2_vae.safetensors", "vae", {"minimal", "full"}),
]

selected = [(r, f, s) for (r, f, s, levels) in ITEMS if LEVEL in levels]
print(f"==> level={LEVEL}  pulling {len(selected)} files (skips already-on-disk)\n", flush=True)

for repo, fname, sub in selected:
    dest_dir = ROOT / sub
    dest_dir.mkdir(parents=True, exist_ok=True)
    target = dest_dir / Path(fname).name
    if target.exists() and target.stat().st_size > 0:
        print(f"SKIP  {target.name} ({target.stat().st_size/1024/1024:.0f}MB)")
        continue
    print(f"PULL  {target.name}", flush=True)
    t0 = time.time()
    try:
        p = hf_hub_download(repo_id=repo, filename=fname, local_dir=str(dest_dir), token=TOKEN)
    except GatedRepoError:
        print(f"ERROR: {repo} is gated. Run: huggingface-cli login  (or export HF_TOKEN=hf_...)", file=sys.stderr)
        sys.exit(2)
    except RepositoryNotFoundError:
        print(f"ERROR: {repo} not found (was it renamed?)", file=sys.stderr)
        sys.exit(2)
    except HfHubHTTPError as e:
        if "401" in str(e) or "403" in str(e):
            print(f"ERROR: HF auth required for {repo}/{fname}. Run: huggingface-cli login  (or export HF_TOKEN=hf_...)", file=sys.stderr)
        else:
            print(f"ERROR: HF download failed for {repo}/{fname}: {e}", file=sys.stderr)
        sys.exit(2)
    src = Path(p)
    if src != target:
        shutil.move(str(src), str(target))
        # cleanup nested split_files dir if empty
        nest = dest_dir / "split_files"
        if nest.exists():
            shutil.rmtree(nest, ignore_errors=True)
    sz = target.stat().st_size / (1024**3)
    print(f"  done  {sz:.1f}GB in {time.time()-t0:.1f}s")

print("ALL_DONE")
PY

echo ""
echo "==> Wan 2.2 $LEVEL set landed in ~/ComfyUI/models/. Restart ComfyUI to see them in node dropdowns."
`,

	"wan": `#!/bin/bash
set -euo pipefail
echo "==> GH200 + Wan 2.2 full setup — meta-package"
echo ""
echo "Dependencies (pytorch-gh200, comfyui-wan-stack, wan2.2-full) install first."
echo "This step just verifies the stack and writes the no-LoRA max-quality workflow."
echo ""

VENV="$HOME/ComfyUI/venv"

# Sanity check
"$VENV/bin/python" - <<'PY'
import torch
print(f"torch          {torch.__version__}  cuda={torch.version.cuda}")
try:
    from importlib.metadata import version as _v
    print(f"sage attention {_v('sageattention')}")
except Exception:
    import sageattention as _s  # noqa: F401
    print("sage attention installed")
print(f"device         {torch.cuda.get_device_name(0)}")
PY

# Drop the canonical no-LoRA max-quality workflow into user workflows
WF_DIR="$HOME/ComfyUI/user/default/workflows"
mkdir -p "$WF_DIR"
WF_FILE="$WF_DIR/Wan2.2_14B_T2V_NoLoRA_MaxQuality.json"
if [ -f "$WF_FILE" ] && [ "$(wc -c <"$WF_FILE")" -gt 1000 ]; then
    echo "==> workflow already present at $WF_FILE"
else
    echo "==> Note: full workflow JSON ships in anime/cli/embedded/workflows/."
    echo "    Copy it manually if not auto-installed:"
    echo "    cp \$ANIME_REPO/cli/embedded/workflows/Wan2.2_14B_T2V_NoLoRA_MaxQuality.json $WF_FILE"
fi

echo ""

# Verify Comfort is running
COMFORT_UI="$HOME/Comfort/comfort-ui"
if [ -d "$COMFORT_UI/dist" ]; then
    if screen -list 2>/dev/null | grep -q comfort; then
        echo "==> Comfort running on :3000"
    else
        echo "==> Starting Comfort on :3000..."
        screen -dmS comfort bash -c "cd $COMFORT_UI && npx serve dist -l 3000"
    fi
else
    echo "==> Comfort not found — run: anime install comfort"
fi
echo ""
`,

	"comfort": `#!/bin/bash
set -euo pipefail
echo "==> Installing Comfort — Wan T2V Atelier UI"

# ─── ensure node + npm ────────────────────────────────────────────
# Lambda Stack / GH200 / fresh Ubuntu images do not ship a Node
# runtime. Rather than fail loud and force the user to discover the
# 'nodejs' package, we install Node 20 LTS via NodeSource here. The
# studio bootstrap is meant to be a single command end-to-end.
if ! command -v node >/dev/null 2>&1 || ! command -v npm >/dev/null 2>&1; then
    echo "==> node/npm missing — installing Node 20 LTS via NodeSource"
    if [ "$(id -u)" -eq 0 ]; then
        SUDO=""
    else
        sudo -n true 2>/dev/null || { echo "ERROR: need sudo to install nodejs"; exit 1; }
        SUDO="sudo"
    fi
    $SUDO DEBIAN_FRONTEND=noninteractive apt-get update -y
    $SUDO DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends curl ca-certificates
    curl -fsSL https://deb.nodesource.com/setup_20.x | $SUDO -E bash -
    $SUDO DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends nodejs
    echo "==> node $(node --version) / npm $(npm --version) installed"
fi

REPO="$HOME/Comfort"
UI="$REPO/comfort-ui"
GH_REPO="quivent/comfort"

# ─── clone the Comfort repo ───────────────────────────────────────
# The repo is currently PRIVATE under quivent/, so a bare HTTPS clone
# fails (and on a TTY git prompts for username/password forever — the
# "Password authentication is not supported" cul-de-sac). We try
# multiple credential paths in order and never let git prompt.
clone_comfort() {
    local target="$1"
    # GIT_TERMINAL_PROMPT=0 makes git fail fast instead of asking for
    # credentials interactively; if our auto-detection is wrong we want
    # a clean error, not a hang.
    export GIT_TERMINAL_PROMPT=0

    # 1. SSH key with github.com access (most common dev-box setup).
    #    Probe with -T to confirm the key authenticates, then clone.
    if ssh -T -o BatchMode=yes -o ConnectTimeout=5 -o StrictHostKeyChecking=accept-new git@github.com 2>&1 | grep -q "successfully authenticated"; then
        echo "==> trying SSH (your ~/.ssh key authenticates against github.com)"
        if GIT_SSH_COMMAND="ssh -o BatchMode=yes" git clone --depth 1 "git@github.com:${GH_REPO}.git" "$target" 2>&1; then
            return 0
        fi
        echo "    SSH clone failed (key may not have access to ${GH_REPO})"
    fi

    # 2. gh CLI is logged in (also common — Lambda images often have it).
    if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
        local user="$(gh api user --jq .login 2>/dev/null || echo authenticated)"
        echo "==> trying gh CLI (logged in as $user)"
        if gh repo clone "$GH_REPO" "$target" -- --depth 1 2>&1; then
            [ -d "$target/.git" ] && return 0
        fi
    fi

    # 3. HTTPS with a personal access token from env (CI-friendly).
    local tok="${GH_TOKEN:-${GITHUB_TOKEN:-}}"
    if [ -n "$tok" ]; then
        echo "==> trying HTTPS with GH_TOKEN/GITHUB_TOKEN"
        if git clone --depth 1 "https://x-access-token:${tok}@github.com/${GH_REPO}.git" "$target" 2>&1; then
            return 0
        fi
    fi

    # 4. Anonymous HTTPS — only succeeds if the repo becomes public.
    echo "==> trying anonymous HTTPS (works only if ${GH_REPO} is public)"
    if git clone --depth 1 "https://github.com/${GH_REPO}.git" "$target" 2>&1; then
        return 0
    fi

    cat >&2 <<HELP

ERROR: could not clone github.com/${GH_REPO} via any of:
       SSH key, gh CLI, GH_TOKEN env, anonymous HTTPS.

       quivent/comfort is private. Easiest fix is one command:

           anime gh login

       That walks you through gh auth (web flow), generates an SSH
       key if you don't have one, uploads it to GitHub, and verifies
       access. Then re-run:

           anime wan studio --yes

       Alternatives if you'd rather not use anime gh login:
         • Run gh auth login yourself, then add SSH key:
             gh auth login --git-protocol ssh --web
             gh ssh-key add ~/.ssh/id_ed25519.pub --title "\$(hostname)"
         • Export a PAT with repo scope:
             export GH_TOKEN=ghp_...

HELP
    return 1
}

# Clone or update the Comfort repo
if [ -d "$REPO/.git" ]; then
    echo "==> $REPO already cloned, pulling latest"
    GIT_TERMINAL_PROMPT=0 git -C "$REPO" pull --ff-only 2>&1 || echo "WARN: pull failed, continuing with current tree"
else
    echo "==> Cloning github.com/${GH_REPO} to $REPO"
    clone_comfort "$REPO" || exit 1
fi

if [ ! -d "$UI" ]; then
    echo "ERROR: $UI missing — repo layout changed?"
    exit 1
fi

cd "$UI"

# Install + build the UI. Prefer ci (lockfile-respecting) when a lock is present.
if [ -f package-lock.json ]; then
    echo "==> npm ci"
    npm ci --no-audit --no-fund
else
    echo "==> npm install"
    npm install --no-audit --no-fund
fi

echo "==> npm run build"
npm run build

if [ ! -d dist ]; then
    echo "ERROR: build did not produce dist/ — check build output above"
    exit 1
fi

# Drop a tiny launch hint at ~/.anime/comfort-path so the Go side can find it
mkdir -p "$HOME/.anime"
echo "$UI" > "$HOME/.anime/comfort-path"

echo ""
echo "==> Comfort installed at $UI"
echo "==> dist/ ready ($(du -sh dist | cut -f1))"
echo ""

# Start Comfort
echo "==> Starting Comfort on :3000..."
if screen -list 2>/dev/null | grep -q comfort; then
    screen -S comfort -X quit 2>/dev/null || true
fi
screen -dmS comfort bash -c "cd $UI && npx serve dist -l 3000"
echo "    Comfort running on :3000 (screen -r comfort)"
`,

	"domain": `#!/bin/bash
set -euo pipefail

DOMAIN="comfort.producer.cafe"

echo "==> Domain + SSL setup for $DOMAIN"
echo ""

# Step 1: Detect public IP
echo "==> Detecting public IP..."
MY_IP=$(curl -s --max-time 10 ifconfig.me || true)
if [ -z "$MY_IP" ]; then
    MY_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || true)
fi
if [ -z "$MY_IP" ]; then
    echo "ERROR: could not detect public IP"
    echo "  Run manually: anime dns point $DOMAIN <your-ip> --ssl"
    exit 1
fi
echo "    Public IP: $MY_IP"
echo ""

# Step 2: Verify Comfort studio is reachable
echo "==> Checking Comfort studio..."
if curl -s --max-time 5 http://127.0.0.1:5180 >/dev/null 2>&1; then
    echo "    Comfort responding on :5180"
elif curl -s --max-time 5 http://127.0.0.1:5173 >/dev/null 2>&1; then
    echo "    Comfort dev responding on :5173"
else
    echo "    WARNING: Comfort not responding on :5180 or :5173"
    echo "    Start it first: anime wan studio --public --port 5180"
fi
echo ""

# Step 3: Check dns config exists
echo "==> Checking DNS credentials..."
if [ ! -f "$HOME/.dns-config.json" ]; then
    echo "ERROR: ~/.dns-config.json not found"
    echo ""
    echo "  Create it with your Vercel API token:"
    echo "    echo '{\"token\": \"your-vercel-token\", \"teamId\": \"your-team-id\"}' > ~/.dns-config.json"
    echo ""
    echo "  Get a token at: https://vercel.com/account/tokens"
    echo "  Then retry: anime install domain"
    exit 1
fi
echo "    Credentials found"
echo ""

# Step 4: Point DNS (no SSL yet)
echo "==> Pointing $DOMAIN → $MY_IP..."
echo ""
anime dns point "$DOMAIN" "$MY_IP"
echo ""

# Step 5: Install nginx + certbot locally (we're already on the server)
echo "==> Installing nginx + certbot..."
if [ "$(id -u)" -eq 0 ]; then SUDO=""; else SUDO="sudo"; fi
$SUDO apt-get update -y -qq
$SUDO apt-get install -y -qq nginx certbot python3-certbot-nginx

echo "==> Configuring nginx reverse proxy for $DOMAIN → Comfort dev server..."

# Detect Comfort dev port: Vite defaults to 5173, studio defaults to 5180
COMFORT_PORT=5173
if curl -s --max-time 2 http://127.0.0.1:5180 >/dev/null 2>&1; then
    COMFORT_PORT=5180
fi

cat <<NGINX | $SUDO tee /etc/nginx/sites-available/comfort >/dev/null
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://127.0.0.1:$COMFORT_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
NGINX

$SUDO ln -sf /etc/nginx/sites-available/comfort /etc/nginx/sites-enabled/comfort
$SUDO rm -f /etc/nginx/sites-enabled/default
$SUDO nginx -t && $SUDO systemctl reload nginx
echo "    nginx configured → :$COMFORT_PORT"

echo "==> Requesting Let's Encrypt certificate..."
$SUDO certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --register-unsafely-without-email --redirect
echo "    SSL certificate issued"

$SUDO systemctl enable certbot.timer 2>/dev/null || true

echo ""
echo "============================================"
echo ""
echo "  https://$DOMAIN"
echo ""
echo "============================================"
`,
}

