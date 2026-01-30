package protocol

import "time"

// NewCoverageProtocol creates the DeepSeek V3.2-Exp coverage protocol
// for 8×B200 GPU cluster deployment - handles blank Ubuntu system
func NewCoverageProtocol() *Protocol {
	return &Protocol{
		Name:        "coverage",
		Description: "Deploy DeepSeek V3.2-Exp on 8×B200 cluster with vLLM",
		Category:    "LLM",
		Version:     "1.0.0",
		Requirements: Requirements{
			GPUs:        8,
			GPUMemoryGB: 192, // B200 has 192GB HBM3e
			SystemMemGB: 512,
			DiskSpaceGB: 1000,
			CUDA:        "12.4+",
			Python:      "3.11+",
			OS:          []string{"linux"},
			Arch:        []string{"arm64", "amd64"},
		},
		Phases: []*Phase{
			// Phase 1: System Dependencies - Install everything from scratch
			{
				Name:        "System Dependencies",
				Description: "Install CUDA toolkit, Python 3.11, and core dependencies",
				Status:      StatusPending,
				Commands: []Command{
					{
						Description: "Update apt package lists",
						Command:     "apt-get",
						Args:        []string{"update"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install essential build tools",
						Command:     "apt-get",
						Args:        []string{"install", "-y", "build-essential", "wget", "curl", "git", "software-properties-common", "lsb-release", "ninja-build"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install Python 3.11 (add PPA if needed)",
						Command:     "bash",
						Args: []string{"-c", `
if python3.11 --version >/dev/null 2>&1; then
  echo "Python 3.11 already installed: $(python3.11 --version)"
else
  echo "Adding deadsnakes PPA for Python 3.11..."
  add-apt-repository -y ppa:deadsnakes/ppa
  apt-get update
  apt-get install -y python3.11 python3.11-venv python3.11-dev
fi
`},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install NVIDIA CUDA keyring (auto-detect arch)",
						Command:     "bash",
						Args: []string{"-c", `
ARCH=$(dpkg --print-architecture)
DISTRO=$(lsb_release -cs 2>/dev/null || echo "jammy")
if [ "$DISTRO" = "noble" ]; then DISTRO="ubuntu2404"; else DISTRO="ubuntu2204"; fi
if [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
  CUDA_ARCH="sbsa"
else
  CUDA_ARCH="x86_64"
fi
echo "Detected: $DISTRO / $CUDA_ARCH"
wget -q "https://developer.download.nvidia.com/compute/cuda/repos/${DISTRO}/${CUDA_ARCH}/cuda-keyring_1.1-1_all.deb" -O /tmp/cuda-keyring.deb && \
dpkg -i /tmp/cuda-keyring.deb && \
rm /tmp/cuda-keyring.deb && \
apt-get update
`},
						Sudo:        true,
						IgnoreError: false, // This must succeed for CUDA to install
					},
					{
						Description: "Install CUDA toolkit 12.4",
						Command:     "apt-get",
						Args:        []string{"install", "-y", "cuda-toolkit-12-4"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Add CUDA to PATH",
						Command:     "bash",
						Args: []string{"-c", `
echo 'export PATH=/usr/local/cuda-12.4/bin:$PATH' >> /etc/profile.d/cuda.sh && \
echo 'export LD_LIBRARY_PATH=/usr/local/cuda-12.4/lib64:$LD_LIBRARY_PATH' >> /etc/profile.d/cuda.sh && \
chmod +x /etc/profile.d/cuda.sh
`},
						Sudo:        true,
						IgnoreError: true,
					},
					{
						Description: "Create Python virtual environment",
						Command:     "bash",
						Args: []string{"-c", `
if [ -f /opt/vllm-env/bin/python ]; then
  echo "Virtual environment already exists at /opt/vllm-env"
else
  python3.11 -m venv /opt/vllm-env
  echo "Created virtual environment at /opt/vllm-env"
fi
`},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Upgrade pip in virtual environment",
						Command:     "/opt/vllm-env/bin/pip",
						Args:        []string{"install", "--upgrade", "pip", "setuptools", "wheel"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install PyTorch with CUDA 12.4 support",
						Command:     "/opt/vllm-env/bin/pip",
						Args: []string{
							"install",
							"torch",
							"torchvision",
							"torchaudio",
							"--index-url",
							"https://download.pytorch.org/whl/cu124",
						},
						Sudo:        true,
						IgnoreError: false,
					},
				},
				Verification: &Verification{
					Description: "Verify PyTorch CUDA support",
					Command:     "/opt/vllm-env/bin/python",
					Args:        []string{"-c", "import torch; print('CUDA:', torch.cuda.is_available())"},
					Timeout:     30 * time.Second,
				},
			},

			// Phase 2: vLLM Installation
			{
				Name:         "vLLM Installation",
				Description:  "Install vLLM and required dependencies",
				Status:       StatusPending,
				Dependencies: []string{"System Dependencies"},
				Commands: []Command{
					{
						Description: "Install vLLM",
						Command:     "/opt/vllm-env/bin/pip",
						Args:        []string{"install", "vllm>=0.6.0"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install flash-attention (pre-built or skip)",
						Command:     "bash",
						Args: []string{"-c", `
# flash-attn is OPTIONAL - vLLM works fine without it (uses PagedAttention)
# Try pre-built wheel with 60s timeout, skip if not available
export MAX_JOBS=$(nproc)
timeout 60 /opt/vllm-env/bin/pip install flash-attn --no-build-isolation --prefer-binary 2>&1 || {
  echo "⚠ flash-attn pre-built not available, skipping (vLLM will use built-in attention)"
  exit 0
}
`},
						Sudo:        true,
						IgnoreError: true, // Optional - vLLM works without it
					},
					{
						Description: "Install transformers and safetensors",
						Command:     "/opt/vllm-env/bin/pip",
						Args:        []string{"install", "transformers>=4.44.0", "safetensors"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install additional dependencies",
						Command:     "/opt/vllm-env/bin/pip",
						Args:        []string{"install", "accelerate", "pydantic", "fastapi", "uvicorn", "requests"},
						Sudo:        true,
						IgnoreError: false,
					},
				},
				Verification: &Verification{
					Description: "Verify vLLM installation",
					Command:     "/opt/vllm-env/bin/python",
					Args:        []string{"-c", "import vllm; print('vLLM version:', vllm.__version__)"},
					Timeout:     30 * time.Second,
				},
			},

			// Phase 3: Model Download
			{
				Name:         "Model Download",
				Description:  "Download DeepSeek-V3 from HuggingFace",
				Status:       StatusPending,
				Dependencies: []string{"vLLM Installation"},
				Commands: []Command{
					{
						Description: "Create model directory",
						Command:     "mkdir",
						Args:        []string{"-p", "/models/deepseek-v3"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Set permissions on model directory",
						Command:     "chmod",
						Args:        []string{"755", "/models"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Install huggingface-hub CLI with hf_transfer",
						Command:     "/opt/vllm-env/bin/pip",
						Args:        []string{"install", "huggingface-hub[cli]", "hf_transfer"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Download DeepSeek-V3 model weights (this takes a while)",
						Command:     "bash",
						Args: []string{"-c", `
export HF_HUB_ENABLE_HF_TRANSFER=1
/opt/vllm-env/bin/huggingface-cli download deepseek-ai/DeepSeek-V3 \
  --local-dir /models/deepseek-v3 \
  --local-dir-use-symlinks False
`},
						Sudo:        true,
						IgnoreError: false,
					},
				},
				Verification: &Verification{
					Description: "Verify model files exist",
					Command:     "test",
					Args:        []string{"-f", "/models/deepseek-v3/config.json"},
					Timeout:     10 * time.Second,
				},
			},

			// Phase 4: vLLM Server Launch
			{
				Name:         "vLLM Server Launch",
				Description:  "Launch vLLM server with optimal B200 configuration",
				Status:       StatusPending,
				Dependencies: []string{"Model Download"},
				Commands: []Command{
					{
						Description: "Create vLLM systemd service file",
						Command:     "bash",
						Args:        []string{"-c", "cat > /etc/systemd/system/vllm-deepseek.service << 'EOF'\n" + vllmServiceFileContent() + "EOF"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Reload systemd daemon",
						Command:     "systemctl",
						Args:        []string{"daemon-reload"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Enable vLLM service",
						Command:     "systemctl",
						Args:        []string{"enable", "vllm-deepseek.service"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Start vLLM service",
						Command:     "systemctl",
						Args:        []string{"start", "vllm-deepseek.service"},
						Sudo:        true,
						IgnoreError: false,
					},
					{
						Description: "Wait for vLLM to be ready (polls health endpoint)",
						Command:     "bash",
						Args: []string{"-c", `
echo "Waiting for vLLM to start (this may take a few minutes for model loading)..."
for i in $(seq 1 60); do
  if curl -sf http://localhost:8000/health >/dev/null 2>&1; then
    echo "vLLM is ready!"
    exit 0
  fi
  echo "  Attempt $i/60 - waiting 10s..."
  sleep 10
done
echo "Timeout waiting for vLLM"
exit 1
`},
						IgnoreError: false,
					},
				},
				Verification: &Verification{
					Description: "Check vLLM service status",
					Command:     "systemctl",
					Args:        []string{"is-active", "vllm-deepseek.service"},
					ExpectedOut: "active",
					Timeout:     10 * time.Second,
				},
			},

			// Phase 5: Verification
			{
				Name:         "Verification",
				Description:  "Verify vLLM server is running and responding",
				Status:       StatusPending,
				Dependencies: []string{"vLLM Server Launch"},
				Commands: []Command{
					{
						Description: "Check health endpoint",
						Command:     "curl",
						Args:        []string{"-sf", "http://localhost:8000/health"},
						IgnoreError: false,
					},
					{
						Description: "Get model info",
						Command:     "curl",
						Args:        []string{"-sf", "http://localhost:8000/v1/models"},
						IgnoreError: false,
					},
					{
						Description: "Test inference with simple prompt",
						Command:     "curl",
						Args: []string{
							"-sf", "-X", "POST",
							"http://localhost:8000/v1/completions",
							"-H", "Content-Type: application/json",
							"-d", `{"model": "/models/deepseek-v3", "prompt": "Hello", "max_tokens": 5}`,
						},
						IgnoreError: false,
					},
					{
						Description: "Check GPU memory usage",
						Command:     "nvidia-smi",
						Args:        []string{"--query-gpu=name,memory.used,memory.total", "--format=csv,noheader"},
						IgnoreError: false,
					},
				},
				Verification: &Verification{
					Description: "Verify health endpoint returns 200",
					Command:     "curl",
					Args:        []string{"-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:8000/health"},
					ExpectedOut: "200",
					Timeout:     10 * time.Second,
				},
			},

			// Phase 6: Optional Prefix Cache Warmup
			{
				Name:        "Prefix Cache Warmup",
				Description: "Pre-cache common prefixes for faster inference (optional)",
				Status:      StatusPending,
				Commands: []Command{
					{
						Description: "Create warmup script",
						Command:     "bash",
						Args:        []string{"-c", "cat > /opt/vllm-env/warmup_cache.py << 'EOF'\n" + warmupScriptContent() + "EOF"},
						Sudo:        true,
						IgnoreError: true,
					},
					{
						Description: "Run warmup script",
						Command:     "/opt/vllm-env/bin/python",
						Args:        []string{"/opt/vllm-env/warmup_cache.py"},
						Sudo:        true,
						IgnoreError: true,
					},
				},
			},
		},
	}
}

// vllmServiceFileContent returns the systemd service file content for vLLM
func vllmServiceFileContent() string {
	return `[Unit]
Description=vLLM Server for DeepSeek-V3
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vllm-env
Environment="PATH=/usr/local/cuda-12.4/bin:/opt/vllm-env/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin"
Environment="LD_LIBRARY_PATH=/usr/local/cuda-12.4/lib64"
ExecStart=/opt/vllm-env/bin/python -m vllm.entrypoints.openai.api_server \
    --model /models/deepseek-v3 \
    --tensor-parallel-size 8 \
    --max-model-len 131072 \
    --gpu-memory-utilization 0.95 \
    --enable-prefix-caching \
    --enable-chunked-prefill \
    --max-num-seqs 256 \
    --trust-remote-code \
    --dtype float16 \
    --quantization fp8 \
    --port 8000 \
    --host 0.0.0.0
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`
}

// warmupScriptContent returns the Python script for cache warmup
func warmupScriptContent() string {
	return `#!/usr/bin/env python3
"""Warmup script for vLLM prefix caching."""
import requests

ENDPOINT = "http://localhost:8000/v1/completions"
PREFIXES = ["INT. ", "EXT. ", "FADE IN:", "The ", "Once upon"]

def warmup():
    print("Warming up prefix cache...")
    for p in PREFIXES:
        try:
            r = requests.post(ENDPOINT, json={
                "model": "/models/deepseek-v3",
                "prompt": p,
                "max_tokens": 1,
                "temperature": 0
            }, timeout=30)
            print(f"  ✓ {p!r}")
        except Exception as e:
            print(f"  ✗ {p!r}: {e}")
    print("Done!")

if __name__ == "__main__":
    warmup()
`
}
