#!/usr/bin/env bash
# gh200-vllm-bootstrap.sh — blank Lambda GH200 → vLLM + Llama 70B in <30 min.
# Idempotent. No Docker. No venv. ARM64 + driver 570 / CUDA 12.8.
set -euo pipefail
log(){ printf '\033[36m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
die(){ printf '\033[31m[FATAL]\033[0m %s\n' "$*" >&2; exit 1; }

# ---------- 0. Sanity ----------
[[ "$(uname -m)" == "aarch64" ]] || die "expected aarch64 (GH200)"
nvidia-smi >/dev/null 2>&1 || die "nvidia driver missing"
nvidia-smi | grep -q "CUDA Version: 12.8" || log "WARN: CUDA 12.8 not reported; continuing"

# ---------- 1. Global env (write once into ~/.bashrc) ----------
BRC="$HOME/.bashrc"
grep -q "GH200_VLLM_ENV" "$BRC" || cat >>"$BRC" <<'EOF'
# === GH200_VLLM_ENV ===
export CUDA_HOME=/usr/local/cuda-12.8
export PATH=$CUDA_HOME/bin:$PATH
export LD_LIBRARY_PATH=$CUDA_HOME/lib64:${LD_LIBRARY_PATH:-}
export TORCH_CUDA_ARCH_LIST="9.0+PTX"        # Hopper sm_90 only — halves build time
export MAX_JOBS=32                           # Grace has 72 cores; 32 keeps RAM safe
export VLLM_TARGET_DEVICE=cuda
export HF_HOME=/home/ubuntu/.cache/huggingface
export PIP_CONSTRAINT=/home/ubuntu/.pip-constraints.txt
# export HF_TOKEN=hf_xxx  # user must fill in
EOF
# shellcheck disable=SC1090
source "$BRC"
[[ -n "${HF_TOKEN:-}" ]] || log "WARN: HF_TOKEN unset — gated models will 401"

# ---------- 2. Constraints file — single source of truth for pins ----------
cat >"$PIP_CONSTRAINT" <<'EOF'
torch==2.7.1+cu128
torchvision==0.22.1+cu128
transformers==4.55.4
tokenizers==0.20.3
numpy==1.26.4
triton==3.2.0
xformers==0.0.30
EOF

# ---------- 3. APT layer (idempotent — apt-get is a no-op if already current) ----------
log "apt layer"
sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
  build-essential cmake ninja-build git git-lfs python3.10-dev python3-pip \
  libnuma-dev pkg-config libopenmpi-dev || die "apt failed"
git lfs install --skip-repo >/dev/null

# ---------- 4. Torch: detect-or-install (ARM wheels live on download.pytorch.org/whl/cu128) ----------
TORCH_OK=$(python3 -c "import torch,sys; print(torch.__version__==\"2.7.1+cu128\" and torch.cuda.is_available())" 2>/dev/null || echo False)
if [[ "$TORCH_OK" != "True" ]]; then
  log "installing torch 2.7.1+cu128 (ARM wheel)"
  pip install --upgrade pip wheel setuptools
  pip install --index-url https://download.pytorch.org/whl/cu128 \
    torch==2.7.1+cu128 torchvision==0.22.1+cu128 || die "torch install failed"
  python3 -c "import torch; assert torch.cuda.is_available()" || die "torch can't see CUDA"
fi

# ---------- 5. vLLM: try wheel, else source-build (no prebuilt ARM wheel exists for 0.10.1.1) ----------
VLLM_OK=$(python3 -c "import vllm._C, vllm; print(vllm.__version__)" 2>/dev/null || echo MISSING)
if [[ "$VLLM_OK" != "0.10.1.1" ]]; then
  log "vllm missing or broken ($VLLM_OK) — source build"
  pip uninstall -y vllm vllm-flash-attn 2>/dev/null || true
  SRC="$HOME/src/vllm"
  [[ -d "$SRC" ]] || git clone --depth 1 --branch v0.10.1.1 https://github.com/vllm-project/vllm "$SRC"
  cd "$SRC"
  pip install -r requirements/build.txt
  pip install -r requirements/cuda.txt
  # use_existing_torch keeps our pinned 2.7.1+cu128 from being overwritten
  python3 use_existing_torch.py
  pip install --no-build-isolation -e . || die "vllm build failed — check ninja log"
  python3 -c "import vllm._C" || die "vllm._C still missing — likely TORCH_CUDA_ARCH_LIST mismatch"
fi

# ---------- 6. Transformers pin (vllm 0.10.1.1 breaks on >=4.56 tokenizer API) ----------
pip install --upgrade "transformers==4.55.4" "tokenizers==0.20.3"

# ---------- 7. Smoke test + launch ----------
log "smoke test"
python3 -c "from vllm import LLM; print('vllm import OK')" || die "vllm import broken"

MODEL="${MODEL:-meta-llama/Meta-Llama-3.1-70B-Instruct}"
log "launching vLLM server: $MODEL"
exec python3 -m vllm.entrypoints.openai.api_server \
  --model "$MODEL" --tensor-parallel-size 1 \
  --max-model-len 8192 --gpu-memory-utilization 0.92 \
  --host 0.0.0.0 --port 8000
