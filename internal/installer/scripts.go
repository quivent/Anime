package installer

// Embedded installation scripts
var Scripts = map[string]string{
	"core": `#!/bin/bash
set -e
echo "==> Installing Core System"
sudo apt update && sudo apt upgrade -y
sudo apt install -y build-essential git curl wget vim htop tmux cmake pkg-config \
    libssl-dev libffi-dev python3 python3-pip python3-venv python3-dev

# NVIDIA/CUDA
if ! command -v nvidia-smi &> /dev/null; then
    echo "==> Installing NVIDIA drivers and CUDA"
    wget -q https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/arm64/cuda-keyring_1.1-1_all.deb -O /tmp/cuda-keyring.deb
    sudo dpkg -i /tmp/cuda-keyring.deb
    sudo apt update
    sudo apt install -y cuda-toolkit-12-4 nvidia-driver-550
fi

# Node.js
if ! command -v node &> /dev/null; then
    echo "==> Installing Node.js"
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt install -y nodejs
    sudo npm install -g yarn pnpm typescript
fi

# Docker
if ! command -v docker &> /dev/null; then
    echo "==> Installing Docker"
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
fi

echo "==> Core system installed successfully"
`,

	"python": `#!/bin/bash
set -e
echo "==> Setting up Python environment"
pip3 install --upgrade pip setuptools wheel
pip3 install numpy scipy pandas matplotlib pillow
echo "==> Python environment ready"
python3 --version
pip3 --version
`,

	"pytorch": `#!/bin/bash
set -e
echo "==> Installing PyTorch and AI libraries"
pip3 install --upgrade pip
pip3 install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu124
pip3 install transformers diffusers accelerate safetensors xformers bitsandbytes
pip3 install numpy scipy pandas matplotlib pillow opencv-python
echo "==> PyTorch installed successfully"
python3 -c "import torch; print(f'CUDA available: {torch.cuda.is_available()}')"
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
COMFYUI_DIR="$HOME/ComfyUI"
if [ -d "$COMFYUI_DIR" ]; then
    echo "ComfyUI already exists"
    exit 0
fi

git clone https://github.com/comfyanonymous/ComfyUI.git "$COMFYUI_DIR"
pip3 install -r "$COMFYUI_DIR/requirements.txt"
git clone https://github.com/ltdrdata/ComfyUI-Manager.git "$COMFYUI_DIR/custom_nodes/ComfyUI-Manager"
echo "==> ComfyUI installed successfully"
`,

	"claude": `#!/bin/bash
set -e
echo "==> Installing Claude Code CLI"
if command -v claude-code &> /dev/null; then
    echo "Claude Code already installed"
    exit 0
fi
sudo npm install -g @anthropic-ai/claude-code
echo "==> Claude Code installed successfully"
claude-code --version
`,
}

