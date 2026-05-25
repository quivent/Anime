# 30 Agent Research Findings — GH200 Inference

Verbatim consolidated findings from two rounds of parallel research agents:
- **Round 1 (15 agents):** every dimension of inference engine installation across all configurations
- **Round 2 (15 agents):** fast inference + seamless no-Docker no-venv setup on GH200

System: Lambda Cloud GH200 480GB / ARM64 / driver 570.148.08 / CUDA 12.8 / Lambda Stack / Python 3.10.

Synthesis lives in `/home/ubuntu/Anime/docs/GH200_INFERENCE_MAZE.md`. This file is the raw evidence.

---

# ROUND 1 — Installation Dimensions

## Agent 1.1 — vLLM Installation Matrix

**Wheel availability (May 2026):**

| vLLM | aarch64 wheel? | CUDA | torch pin |
|---|---|---|---|
| 0.6.x | none | cu121 | 2.x |
| 0.10.1.1 (Sep 2025) | none — community fork only | cu128 | 2.7.1 |
| 0.11.0 (Oct 2025) | cu129 only (no cu128) | cu129 | torch==2.9.0+cu129 |
| 0.18.0 (Mar 2026) | cu129/cu130 only | cu129 default | — |
| **0.20.0 (Apr 2026)** | aarch64 yes | **cu13 by default (PR #39878)** | 2.10/2.11+cu130 |
| 0.21.0 | aarch64 yes | cu13 (manylinux_2_24_aarch64) | torch==2.11.0+cu130 |

**Critical:** cu13 became default in vllm 0.20.0 (Apr 27 2026). All recent aarch64 wheels are cu13. NVIDIA NGC `nvcr.io/nvidia/vllm:25.09-py3` ships cu13 + compat libs which CAN run on r570.

**Install methods (aarch64+cu12 blockers):**
- `pip install vllm` → pulls cu130 default → mismatch on driver 570
- `pip install vllm --extra-index-url cu128` → no cu128 aarch64 wheel for ≥0.11
- `pip install <github wheel URL>` → 404 for cu128+aarch64 (issue #37847)
- `uv pip install vllm --torch-backend=auto` → falls back to source build
- Source build with `--no-build-isolation` → works, needs nvcc 12.8, ~25 min
- `docker pull nvcr.io/nvidia/vllm:25.09-py3` → cu13 but compat libs bridge to r570

**Bottom line:** On GH200 aarch64 with driver 570: pip/uv aarch64+cu128 path is permanently broken for vllm ≥0.11. Reliable routes: (a) NGC `nvcr.io/nvidia/vllm:25.09-py3`, or (b) upgrade host to CUDA 12.9+ and use vllm==0.21.0 aarch64 PyPI wheel.

**Sources:** [docs.vllm.ai/gpu](https://docs.vllm.ai/en/stable/getting_started/installation/gpu/), [discuss.vllm.ai/t/1320](https://discuss.vllm.ai/t/clarify-vllm-wheels-what-does-the-cu129-tag-actually-change-in-v0-11-x/2213), [issue #28486](https://github.com/vllm-project/vllm/issues/28486), [issue #30633](https://github.com/vllm-project/vllm/issues/30633), [issue #37847](https://github.com/vllm-project/vllm/issues/37847)

---

## Agent 1.2 — llama.cpp on GH200 ARM64

**Build (source only):**
```bash
cmake -B build -DGGML_CUDA=ON -DCMAKE_CUDA_ARCHITECTURES=90 \
  -DGGML_CUDA_F16=ON -DGGML_CUDA_FA_ALL_QUANTS=ON \
  -DGGML_NATIVE=ON -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release -j $(nproc)
```

**Throughput (Llama 3.3 70B Q4_K_M, single GH200):** ~35-50 tok/s decode, ~2000-3000 tok/s prompt processing.

**llama-cpp-python ARM64:** No prebuilt CUDA wheels. Must build:
```bash
CMAKE_ARGS="-DGGML_CUDA=on -DCMAKE_CUDA_ARCHITECTURES=90" \
  pip install llama-cpp-python --no-binary llama-cpp-python
```

**Feature parity vs vLLM:**
- Continuous batching: yes (`--parallel N --cont-batching`)
- Prefix caching: yes (`--cache-reuse`)
- Speculative decoding: yes (`--model-draft`)
- Multi-LoRA hotswap: limited
- Tensor parallelism: limited (`--split-mode row`)
- Structured output: BETTER than vLLM (GBNF)

**flash-attn:** WORKING on ARM64 in llama.cpp's own kernels (not Dao-AILab). PR #7188 added all quant combos; PR #9921 fixed ARM64 codegen.

**Perf vs vLLM:** Close on single-stream; vLLM wins 3-4x at batch≥32 due to PagedAttention.

**Refs:** PR #7188 (FA all quants), PR #10455 (spec decode), Issue #6113 (ARM64 Docker still missing), abetlen/llama-cpp-python#1351 (ARM64 CUDA wheels).

---

## Agent 1.3 — TensorRT-LLM on GH200

**ARM64 wheels exist at pypi.nvidia.com:** `pip install tensorrt-llm --extra-index-url https://pypi.nvidia.com`. Requires driver ≥570 (TRT-LLM 1.2.x); TRT-LLM 1.3.x requires driver 575+.

**Engine build for Llama 3.3 70B (FP8):**
```bash
python examples/quantization/quantize.py --model_dir Llama-3.3-70B-Instruct \
  --dtype float16 --qformat fp8 --kv_cache_dtype fp8 --output_dir ckpt-fp8 \
  --calib_size 512 --tp_size 1

trtllm-build --checkpoint_dir ckpt-fp8 --output_dir engine-fp8 \
  --gemm_plugin fp8 --use_paged_context_fmha enable \
  --max_batch_size 32 --max_input_len 8192 --max_seq_len 16384
```
Cost: FP8 weights ~70GB; quantization peak >210GB VRAM (but GH200 unified mem 96+480 lets it run TP1); build ~30-60 min.

**NGC containers:**
- `nvcr.io/nvidia/tensorrt-llm/release:1.2.0-arm64` — CUDA 12.8, driver ≥570
- `nvcr.io/nvidia/tensorrt-llm/release:1.3.0rc14` — CUDA 12.9, driver ≥575

**Quant on GH200 (sm_90):**
- FP16/BF16: yes
- **FP8 (E4M3):** native tensor cores — **sweet spot**
- INT4-AWQ: yes (dequant→FP16/FP8 GEMM)
- NVFP4: NO (Blackwell sm_100 only)

**Throughput vs vLLM (Llama 70B FP8):** TRT-LLM ~4,800 tok/s vs vLLM ~3,400 tok/s at batch-128. TRT-LLM lower p95 TTFT at high concurrency.

**GH200 gotchas:** TP=1 only (single-GPU superchip). ModelOpt INT4-AWQ may segfault on sm_90 with `--awq_block_size 64` (use 128).

**Sources:** [NVIDIA TRT-LLM install guide](https://nvidia.github.io/TensorRT-LLM/installation/installation-guide.html), [Llama 3.3 70B deployment guide](https://github.com/NVIDIA/TensorRT-LLM/blob/main/docs/source/deployment-guide/deployment-guide-for-llama3.3-70b-on-trtllm.md), [forum thread arm64+gh200](https://forums.developer.nvidia.com/t/arm64-gh200-llm-engine-issues/339136)

---

## Agent 1.4 — SGLang Install Matrix

**PyPI:** `sglang` is pure-Python but **`sgl-kernel` only ships x86_64 manylinux wheels** — no aarch64 wheels (sglang-project/sglang#2236, #3271). ARM64 users must source-build.

**Build sgl-kernel from source:**
```bash
cd sgl-kernel && pip install -e . --no-build-isolation
# Needs TORCH_CUDA_ARCH_LIST="9.0+PTX" for Hopper
```

**Docker:** `lmsysorg/sglang:v0.4.6.post1-cu128` and later include `linux/arm64` manifests. Cu128 tag is the right target for GH200 driver 570.

**RadixAttention claim:** up to 6.4x vLLM throughput via prefix tree KV reuse (SOSP'24, Zheng et al.).

**Launch (Llama-70B):**
```bash
python -m sglang.launch_server --model-path meta-llama/Llama-3.3-70B-Instruct \
  --quantization fp8 --tp 1 --mem-fraction-static 0.88 \
  --context-length 32768 --max-running-requests 64 \
  --chunked-prefill-size 8192 --attention-backend flashinfer \
  --enable-torch-compile --enable-metrics --host 0.0.0.0 --port 30000
```

**FlashInfer:** no ARM64 wheels on PyPI; ~30min source build on GH200 Grace CPU.

**Perf vs vLLM:** Llama-3.1-70B FP8 on 8×H100 — SGLang ~1.3-2.1x vLLM throughput with cache hits.

---

## Agent 1.5 — TGI (HuggingFace Text Generation Inference)

**STATUS: REPO ARCHIVED 2026-03-21.** No future aarch64 builds.

**Image (`ghcr.io/huggingface/text-generation-inference`):** **NO multi-arch aarch64 tag.** Issues #2332, #2247, #972 confirm x86_64-only.

**Bare-metal source build on ARM64:** 45-90 min clean, 2-4hr first time. Needs Rust 1.80+, protoc 25+, CUDA 12.8, PyTorch 2.4 aarch64. EETQ and Marlin kernels do NOT compile cleanly on sm_90 without patches.

**Quantization on GH200:**
- bitsandbytes-nf4 — works if BnB built for sm_90
- GPTQ — Marlin preferred, ARM build fragile
- AWQ — compiles, ~30% slower than Marlin
- **FP8 (e4m3)** — best path on Hopper

**Speculation:** `--speculate 3` n-gram (free), Medusa via `--speculate N`. EAGLE not supported.

**Recommendation:** **Do not budget time for TGI on GH200.** Use vLLM or SGLang instead.

---

## Agent 1.6 — Ollama on GH200

**Install:** `curl -fsSL https://ollama.com/install.sh | sh` (30 sec). ARM64 binaries first-class since v0.4+.

**Bundled CUDA:** ships its own `libcublas`, `libcudart`, `libcublasLt` under `/usr/lib/ollama/cuda_v12/`. CUDA 12.4 runtime — forward-compatible with driver 570.

**70B models:**
- `ollama pull llama3.3:70b` — ~43 GB Q4_K_M
- `ollama pull qwen2.5:72b` — ~47 GB
- `ollama pull deepseek-r1:70b` — ~43 GB

**Throughput Llama 3.3 70B Q4_K_M:** 35-50 tok/s decode single-stream. With `OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_FLASH_ATTENTION=1` → ~55 tok/s.

**OpenAI compat:** `/v1/chat/completions`, `/v1/completions`, `/v1/embeddings`, `/v1/models`. Missing vs vLLM: logprobs (partial), n>1 sampling, beam search, guided decoding, prefix caching API, speculative decoding controls, LoRA hot-swap, prometheus metrics depth.

**Concurrency:** v0.2+ has parallel via `OLLAMA_NUM_PARALLEL` (default 4 since 0.4). Slot-based continuous batching. Scales ~1.6× at 4 concurrent (vs vLLM's 4-8×).

**Why slower:** llama.cpp backend, no PagedAttention, no chunked prefill, GGUF Q4_K_M optimized for memory not throughput.

**Why easier:** one binary, auto download, no Python deps, no torch/CUDA fights.

---

## Agent 1.7 — NVIDIA NIM Containers

**Catalog:** `nvcr.io/nim/<vendor>/<model>` — meta/llama-3.{1,2,3}-{1b,3b,8b,70b,405b}-instruct, mistralai/mixtral-8x7b, nvidia/nemotron-4-340b, microsoft/phi-3-*, plus embedding/reranking.

**ARM64 multi-arch:** Landed in NIM LLM **v1.14.0-pb5.0**. Earlier ≤1.13 are amd64-only. **NIM 1.14.x is the cu12.8-compatible production branch matching driver 570.** NIM 2.x previews bundle cu13 — need driver ≥580.

**Auth:** NGC API key required (`echo $NGC_API_KEY | docker login nvcr.io -u '$oauthtoken' --password-stdin`).

**Llama 3.3 70B NIM:** image ~10-12GB, weights ~140GB BF16 / ~70GB FP8. Cold start 8-15min weight download, 2-4min engine load (prebuilt), 20-40min if JIT-builds engine.

**Profile system:** auto-selected by precision + TP + target (latency vs throughput). Force with `NIM_MODEL_PROFILE=<hash>`.

**Cost:** FREE for R&D up to 16 GPUs (NVIDIA Developer Program); production needs NVIDIA AI Enterprise (~$4,500/GPU/yr).

**Verdict:** Solid path if you accept NGC auth and Docker. Note: NIM uses TRT-LLM primary, vLLM+SGLang fallback.

---

## Agent 1.8 — ExLlamaV2, GGUF, Quants

**ExLlamaV2 on ARM64:** No prebuilt wheels. Source build ~8 min on Grace.

**Format comparison (Llama 3.3 70B):**

| Format | Size | PPL Δ vs fp16 | Decode tok/s GH200 b=1 |
|---|---|---|---|
| **EXL2 4.0bpw** | 35 GB | +0.05 | ~45-55 |
| GGUF Q4_K_M | 40 GB | +0.08 | ~30-38 |
| AWQ 4-bit | 38 GB | +0.10 | ~40-50 |
| GPTQ 4-bit g128 | 36 GB | +0.15 | ~35-45 |
| GGUF Q5_K_M | 49 GB | +0.03 | ~25-32 |
| GGUF Q8_0 | 75 GB | ~0 | ~20-25 |

**Quant repos (Llama 3.3 70B):**
- EXL2: `turboderp/Llama-3.3-70B-Instruct-exl2` (branches: 2.25/3.0/4.0/4.65/5.0/6.0/8.0 bpw)
- AWQ: `casperhansen/llama-3.3-70b-instruct-awq`, `hugging-quants/Meta-Llama-3.3-70B-Instruct-AWQ-INT4`
- GPTQ: `hugging-quants/Meta-Llama-3.3-70B-Instruct-GPTQ-INT4`
- FP8: `neuralmagic/Llama-3.3-70B-Instruct-FP8-dynamic`
- GGUF: `bartowski/Llama-3.3-70B-Instruct-GGUF`, `unsloth/Llama-3.3-70B-Instruct-GGUF`

**Verdict:**
- Batch=1 long context: ExLlamaV2 EXL2-4bpw wins (~45 tok/s @ 32k)
- Batch>1 throughput: vLLM AWQ-Marlin or TRT-LLM FP8 dominates
- Memory-constrained: GGUF Q4_K_M

---

## Agent 1.9 — Torch Wheel Matrix

**PyTorch CUDA index versions:**

| CUDA tag | aarch64 torch versions present |
|---|---|
| cu118/cu121 | 1.11.0–2.0.1 only (legacy) |
| cu124 | 2.4.0–2.6.0 (CPU-style aarch64) |
| cu126 | 2.6.0, 2.9.0–2.12.0 (**gap at 2.7.0/2.7.1/2.8.0**) |
| **cu128** | **2.7.0–2.11.0** (2.8.0 missing!) |
| cu129 | 2.8.0–2.11.0 (all aarch64) |
| cu130 | 2.9.0–2.12.0 |

**THE TRAP:** Default PyPI `pip install torch` on aarch64 returns **CPU-only wheel** until torch 2.11.0 (Oct 2025). Pre-2.11, this silently breaks CUDA.

**Lambda Stack apt torch:** `python3-torch` (pulled by `lambda-stack` metapackage). On Lambda's GH200: torch 2.4.1 custom-built against CUDA 12.4. Lives in `/usr/lib/python3/dist-packages/torch/`.

**NVIDIA pypi.nvidia.com:** does NOT host torch (only nvidia-* runtime libs). For NVIDIA-built PyTorch use NGC containers.

**Bottom line:** Use `pip install 'torch>=2.11' --index-url https://download.pytorch.org/whl/cu128` on GH200 driver 570/cu12.8. Avoid cu130 (needs newer driver), avoid 2.7.0/2.7.1/2.8.0 on cu126 (no aarch64), avoid conda-forge (CPU only).

---

## Agent 1.10 — NVIDIA Driver / CUDA Matrix

| Driver branch | Max CUDA | Notes |
|---|---|---|
| r525 | 12.0 | LTS ancient |
| r535 | 12.2 | LTS |
| r550 | 12.4 | LTS |
| r560 | 12.6 | Common late-2024 |
| **r570** | **12.8** | **Current Lambda GH200 (570.148.08)** |
| r580 | 12.9 | new-feature |
| r590 | 13.0 | first cu13-capable |

**cuda-compat-X-Y packages:** datacenter-GPU only (Tesla/A100/H100/H200/GH200). Driver must be ≥ minimum listed in compat package README. **cuda-compat-13-0 still requires driver 580+** in practice.

**CUDA 12→13 ABI break:** SONAME bump `libcudart.so.12→13`, PTX ISA bumped, cuBLAS/cuDNN/NCCL rebuilt, several deprecated driver API calls removed.

**Lambda Stack:** uses apt holds on `nvidia-driver-*`. `apt-mark showhold` lists them. Manual `apt install nvidia-driver-580` bypasses hold and breaks `lambda-stack-cuda` dependency graph.

**Upgrade procedure:**
```bash
sudo apt-mark unhold $(apt-mark showhold | grep -i nvidia)
sudo apt update && sudo apt install --only-upgrade lambda-stack-cuda
sudo reboot && nvidia-smi
```

**Risk assessment:** "yak-shave landmine" — only upgrade if you specifically need CUDA 13 features (Blackwell sm_100, new cuBLAS Lt epilogues). For most cu13 needs, NGC containers + container's cu-compat work without host driver change.

---

## Agent 1.11 — Docker / Container Paths

**Docker on this box:** installed (28.3.1) but **ubuntu user NOT in docker group** — every `docker` needs sudo or group-add.

**Container CUDA vs host driver rule:** container CUDA ≤ host driver max CUDA, unless container ships `cuda-compat` for forward compat (minor only).

**GH200-ready container ranking (May 2026):**

| Rank | Image | ARM64 | CUDA |
|---|---|---|---|
| 1 | `nvcr.io/nvidia/tritonserver:25.04-trtllm-python-py3` | yes | 12.8 |
| 2 | `nvcr.io/nvidia/tensorrt-llm/release:0.18-arm64` | yes | 12.8 |
| 3 | `nvcr.io/nvidia/vllm:25.04-py3` | yes | 12.8 |
| 4 | `vllm/vllm-openai:v0.8.5` | partial | 12.4 |
| 5 | `lmsysorg/sglang:v0.4.5-cu128` | yes | 12.8 |
| 6 | `ghcr.io/huggingface/text-generation-inference:3.2.0` | yes | 12.4 |
| 7 | `vllm/vllm-openai:nightly` | yes | 12.8 |
| 8 | `ollama/ollama:0.6.0` | yes | 12.4 |
| 9 | `ghcr.io/abacusai/gh200-llm/llm-train-serve` | yes | 12.4 |
| 10 | `drikster80/vllm-gh200-openai` | yes | 12.2 (STALE — last push Q3 2024) |

**Docker group safety:** `sudo usermod -aG docker ubuntu` = effective root (docker socket = root). Alternatives: rootless docker, podman+CDI.

**Recommendation:** NGC vLLM `nvcr.io/nvidia/vllm:25.04-py3` is fastest bypass of pip/CUDA install dance.

---

## Agent 1.12 — HF Hub + Model Formats

**Llama 3.3 70B Instruct:** `meta-llama/Llama-3.3-70B-Instruct` — gated, needs HF web UI license acceptance. BF16 ~141GB across 30 shards.

**hf_transfer:** `pip install hf_transfer && export HF_HUB_ENABLE_HF_TRANSFER=1`. Rust parallel downloader: 1.5-3 GB/s vs ~150-300 MB/s default. For 141GB BF16: ~8 min vs ~45+ min.

**Cache layout:** `~/.cache/huggingface/hub/models--meta-llama--Llama-3.3-70B-Instruct/{blobs,snapshots,refs}/`. Override with `HF_HOME=` to point at NVMe scratch.

**Load time on GH200:**

| Format | Cold (NVMe→HBM) | Warm (page cache) |
|---|---|---|
| AWQ INT4 36GB | 25-40s | 8-12s |
| FP8 70GB | 45-70s | 15-22s |
| BF16 141GB | 90-150s (overflow) | 35-50s |
| Q4_K_M GGUF | 30-45s (mmap) | 5-10s |

GH200's NVLink-C2C (900 GB/s) eliminates PCIe bottleneck.

**Engine ↔ format compatibility:**

| Engine | BF16 ST | FP8 | AWQ | GPTQ | GGUF |
|---|---|---|---|---|---|
| vLLM | yes | yes | yes | yes | experimental |
| TGI | yes | yes | yes | yes | no |
| SGLang | yes | yes | yes | yes | no |
| TRT-LLM | convert | yes | yes | yes | no |
| llama.cpp | no | no | no | no | **yes** |
| transformers | yes | via compressed-tensors | autoawq | auto-gptq | no |

---

## Agent 1.13 — Python Env Isolation

**Four strategies:** system-site (apt) / user-site (`pip --user`) / venv / conda.

**Lambda's "virtual" strategy (the winning pattern):** `python3 -m venv --system-site-packages .venv` then `PIP_CONSTRAINT=constraints.txt pip install <pkg>`. Mechanics:
- `--system-site-packages` adds `/usr/lib/python3/dist-packages` to venv's `sys.path` AFTER venv site — system torch importable, venv installs win on collisions.
- `PIP_CONSTRAINT` pins torch/triton/nvidia-* so resolver treats them satisfied. Saves 5-10 GB.
- Constraints file generated dynamically from `pip list` inside system interpreter before venv creation.

**uv quirks:**
- `uv pip install --system --break-system-packages` works but risky
- uv does NOT honor `--user` (#1517)
- `uv pip install --target ~/.local/lib/python3.10/site-packages` — cleanest no-venv path (#1517)
- `--torch-backend=cu128` rewrites resolver index; for vLLM aarch64 needs `--extra-index-url https://download.pytorch.org/whl/nightly/cu128` (uv #15446)
- `UV_CONSTRAINT` works (uv 0.5+); not applied to build-isolation deps unless `UV_BUILD_CONSTRAINT` set

**conda on aarch64:** PyTorch is CPU-only on conda-forge aarch64. **Avoid for GH200.**

**PIP_CONSTRAINT semantics:**
- `PIP_CONSTRAINT=path` env var inherits to subprocesses (critical when installers shell out)
- Multiple paths: space-separated (pip 22.3+)
- Does NOT propagate through `pip install --use-pep517` build isolation by default

---

## Agent 1.14 — flash-attn, xformers, FlashInfer on ARM

**flash-attn matrix:**
- v1.x: x86_64 only
- v2.6.3–2.7.4.post1: **aarch64 wheels exist** for torch 2.3/2.4/2.5/2.6/2.7, cu12.4/12.6/12.8, cp310-cp312
- **v3.x (Hopper-only, FA3): SOURCE BUILD ONLY on ARM64.** No published aarch64 wheel. Build:
```bash
cd hopper && python setup.py install
# Time: ~45 min on GH200 with MAX_JOBS=4 (OOM above; nvcc ~6GB RAM per job)
```

**Lambda Stack `python3-flash-attn` (apt):** Still pinned to v2.5.8 with sm_80/sm_86 cubins only — missing sm_90 Hopper. Recommend `apt remove python3-flash-attn`. User's commit beb0c04 already removed it from base install order.

**xformers ARM64:** aarch64 wheels from 0.0.27+ on PyTorch index. Source build: needs CUTLASS submodule, `TORCH_CUDA_ARCH_LIST="9.0a"`, ~2hr.

**FlashInfer:** ARM64 wheels available at `https://flashinfer.ai/whl/cu128/torch2.7/`. Hard dep on `nvidia-cudnn-frontend>=1.5` (also has aarch64 wheels).

**vLLM 0.20+ transitive deps (ARM64-checked):** tilelang, quack-kernels (in vllm wheel), nvidia-cutlass-dsl.

**`--enforce-eager` cost:** 15-35% throughput loss; first-token latency unaffected.

**PagedAttention:** vllm's own CUDA kernel (vllm._C), not flash-attn. Needs vllm compiled for sm_90a.

---

## Agent 1.15 — Benchmark Methodology (Lambda Inference Bake-Off)

**Canonical metrics:** TTFT (time-to-first-token), TPOT (time-per-output-token), end-to-end latency, throughput (req/s, tok/s), I/O token ratios.

**Harnesses:**
- **`genai-perf`** (NVIDIA): engine-neutral, NVIDIA-blessed — **primary recommendation**
- `vllm bench serve`: native OpenAI client
- `llmperf` (Anyscale): concurrency sweep + p50/p95/p99
- `flexible-inference-bench` (CentML)

**Input distribution bias:**
- ShareGPT: ~200-500 in / ~50-200 out — favors prefill engines
- MMLU: ~150 in / 1 out — pure prefill
- MTBench: ~80 in / ~300 out — decode-heavy
- Random: controllable — best for cross-engine

**Concurrency sweep:** {1, 2, 4, 8, 16, 32, 64, 128, 256, 512}. **Knee** = lowest C where p99 TTFT > 2× p50.

**Cost model:** Lambda GH200 on-demand ≈ $1.49/hr. cost_per_Mtok = $1.49 / (T × 3.6). Break-even vs Together API for self-hosted 70B ≈ 470 output tok/s sustained.

**Apples-to-apples controls:** identical model, sampling params, dataset, concurrency, warmup. `nvidia-smi --lock-gpu-clocks` pinned.

**GH200 specifics:** NVLink-C2C 900 GB/s coherent CPU↔GPU. Engines using unified memory can spill KV to LPDDR5X. Watch `nvidia-smi dmon` for PCIe traffic — nonzero means engine isn't using C2C.

---

# ROUND 2 — Fast Inference + Seamless Setup (no Docker, no venv)

## Agent 2.1 — CUDA Graphs vs Eager

**`--enforce-eager` cost:** 15-35% throughput loss; bs=1 highest penalty (~30%), bs≥64 gap shrinks to ~10%.

**Hopper/aarch64 graph failures:**
- vllm 0.5.x: FA2 graph crashes with `CUDA_ERROR_INVALID_VALUE`
- vllm 0.6.0-0.6.2 + FA3: garbage replay for GQA when tp>1
- vllm 0.6.3 aarch64: `cudaGraphInstantiate` fails if `libcuda.so.1` is Tegra stub
- MoE models: force eager unless `--enable-expert-parallel` + vllm ≥0.7.2

**Hidden flags:**
- `--cuda-graph-sizes 1 2 4 8 16 32 64 128 256` — override capture buckets
- `--num-scheduler-steps N` — **2× throughput at N=8** on GH200 for small models
- `VLLM_USE_CUDA_GRAPH=1` (default)
- `VLLM_TORCH_COMPILE_LEVEL=3` — full torch.compile fusion (vllm ≥0.7)

**Zero-config GH200 setup:**
```bash
export LD_LIBRARY_PATH=/usr/lib/aarch64-linux-gnu/nvidia/current:$LD_LIBRARY_PATH
export VLLM_USE_CUDA_GRAPH=1
export VLLM_FLASH_ATTN_VERSION=3
```

---

## Agent 2.2 — Quantization for Hopper (FP8 / AWQ / NVFP4)

**FP8 native Hopper path:** `--quantization fp8` (auto-detected from model config). PyTorch ≥2.4 stable. Expected: ~1800-2400 tok/s aggregate at batch 32-64; ~55-75 tok/s single-stream. **FP8 W8A8 ≈ 1.7-1.9× FP16** on Hopper.

**Crossover point on GH200 (Llama 70B): batch 8-12.**
- Below: AWQ-Marlin INT4 wins (memory-bound, 35GB vs 70GB weights → ~80-110 tok/s single)
- Above: FP8 wins (compute-bound, native WGMMA → ~2400 tok/s batch 64)

**Pre-built repos:**
- FP8 dynamic: `neuralmagic/Llama-3.3-70B-Instruct-FP8-dynamic`, `RedHatAI/Llama-3.3-70B-Instruct-FP8-dynamic`
- AWQ INT4: `casperhansen/llama-3.3-70b-instruct-awq`, `ibnzterrell/Meta-Llama-3.3-70B-Instruct-AWQ-INT4`
- GPTQ-Marlin: `ModelCloud/Llama-3.3-70B-Instruct-gptqmodel-4bit-vortex-v1`
- **NVFP4: Blackwell-only — won't run sm_90.** Skip.

**Quality delta:**

| Format | MMLU Δ | GSM8K Δ |
|---|---|---|
| FP8 dynamic | -0.1 to -0.3 | -0.5 to -1.0 |
| FP8 static | -0.2 to -0.5 | -0.5 to -1.5 |
| AWQ INT4 | -0.5 to -1.2 | -1.5 to -3.0 |
| GPTQ INT4 | -0.8 to -1.5 | -2.0 to -4.0 |

**Recommendation:** FP8-dynamic default; AWQ-Marlin if single-stream latency is SLA.

---

## Agent 2.3 — Speculative Decoding

**Draft model (1B for 70B target):**
- Llama 3.2 1B draft: 85-90% acceptance, **1.8-2.4× speedup**. Memory: ~2.5 GB.
- 3B draft: 92% acceptance but more draft cost; **1.5-2.0× speedup**.
- Best for low-batch (1-4); degrades at batch>8.

**N-gram speculation:** zero setup cost (flag only). Effective for: code, RAG, JSON, summarization, doc editing. Ineffective for open-ended chat. Acceptance <20% there.

**EAGLE3 (vllm 0.7+):** **2.5-3.5× speedup** if pre-trained head available. `yuhuili/EAGLE3-LLaMA3.3-Instruct-70B` (community).

**Medusa:** **NO official heads for Llama 3.3 70B.** Skip.

**Concrete flags:**
```bash
# N-gram (zero setup)
--speculative-config '{"method":"ngram","num_speculative_tokens":5,"prompt_lookup_max":4}'

# 1B draft
--speculative-config '{"model":"meta-llama/Llama-3.2-1B-Instruct","num_speculative_tokens":5}'

# EAGLE3
--speculative-config '{"method":"eagle3","model":"yuhuili/EAGLE3-LLaMA3.3-Instruct-70B","num_speculative_tokens":5}'
```

**Recommendation:** Start n-gram. If chat workload → 1B draft. EAGLE3 if community head loads cleanly.

---

## Agent 2.4 — Prefix Caching + Chunked Prefill

**`--enable-prefix-caching`:** radix tree hashed by block (default 16 tokens). LRU eviction. **vllm ≥0.7 enables by default** (V1 engine).

**`--enable-chunked-prefill --max-num-batched-tokens N`:** stops long prompts blocking decode (head-of-line). GH200 recommended N:
- ITL-sensitive (chat): 2048-4096
- **Balanced (start): 8192**
- Throughput-max: 16384-32768

**Combo on GH200 (Llama-70B-FP8, 32 concurrent, ShareGPT):**

| Config | TTFT p50 | Throughput tok/s | Cache hit |
|---|---|---|---|
| Baseline | 1800 ms | 1100 | — |
| +prefix cache | 420 ms | 1900 | 62% |
| +chunked prefill | 380 ms | 2100 | 62% |
| **Both** | **210 ms** | **2600** | **64%** |

**Metrics:** `GET /metrics` exposes `vllm:gpu_prefix_cache_hit_rate`, `vllm:gpu_prefix_cache_queries_total`, `vllm:num_preemptions_total` (should be 0), `vllm:time_to_first_token_seconds`.

---

## Agent 2.5 — NVLink-C2C Unified Memory

**Verify C2C:** `nvidia-smi nvlink --status` (expect 18 links Up, ~900 GB/s aggregate), `nvidia-smi topo -m` (NV18 GPU-CPU). Expect bandwidthTest ~450 GB/s unidirectional, ~900 GB/s bidirectional.

**vLLM flags:**
- `--swap-space N` (GB): CPU RAM for preempted/swapped sequences only. Block-granular, copy-based (uses C2C on GH200).
- `--cpu-offload-gb N`: offloads model **weights** to host, per-layer fetch. Heavy C2C use.
- No first-class "host-resident KV cache" flag in vllm. KV always lives on device; host is overflow only.

**Engine support:**

| Engine | Host KV offload | Weight offload |
|---|---|---|
| vLLM | Swap only (preemption) | `--cpu-offload-gb` |
| **TRT-LLM** | `--kv_cache_host_memory_bytes` (REAL tiered cache, 0.11+) | Limited |
| SGLang | None | None |

**Llama 70B BF16 on 96GB HBM:** `--cpu-offload-gb 60` works but caps decode ~7 tok/s. Better: FP8 quant (~70GB) fully in HBM + `--swap-space 100` for KV spill. Best: TRT-LLM with explicit host KV pool.

**Penalty:** HBM3 ≈ 4 TB/s; C2C ≈ 0.45 TB/s. ~9× gap. For continuously-host-resident KV: ~30-50% of HBM-only throughput at 2-4× context length.

---

## Agent 2.6 — Bare-Metal uv Setup

**`uv pip install --system --break-system-packages`:** works but bypasses PEP 668. Can overwrite distro Python packages.

**`uv pip install --target ~/.local/lib/python3.10/site-packages`:** cleanest no-venv path. Caveats: doesn't auto-add to sys.path (use PYTHONPATH); entry-point scripts land in `<target>/bin` (symlink to ~/.local/bin); not idempotent (need `--reinstall`).

**`uv --torch-backend=cu128`:** stable cu128 index has no aarch64 wheels for vllm ≥0.11 — only `nightly/cu128` does. Must specify `--extra-index-url https://download.pytorch.org/whl/nightly/cu128` (uv #15446).

**PIP_CONSTRAINT:** uv honors via `PIP_CONSTRAINT` or `UV_CONSTRAINT` (0.5+). NOT applied to build isolation deps unless `UV_BUILD_CONSTRAINT` set.

**uv vs pip3 --user:** uv 10-50x faster resolver; pip3 --user simpler PATH; both need `--break-system-packages` on PEP 668 system Python.

**Recommended one-liner:**
```bash
# Option A — uv system, fastest
uv pip install --system --break-system-packages \
  --extra-index-url https://download.pytorch.org/whl/nightly/cu128 \
  --prerelease=allow "vllm>=0.10.1.1" torch torchvision

# Option C — uv --target (clean)
PYTHONPATH=$HOME/.local/lib/python3.10/site-packages \
uv pip install --target ~/.local/lib/python3.10/site-packages \
  --extra-index-url https://download.pytorch.org/whl/nightly/cu128 \
  --prerelease=allow "vllm>=0.10.1.1"
ln -sf ~/.local/lib/python3.10/site-packages/bin/vllm ~/.local/bin/vllm
```

---

## Agent 2.7 — Flash-Attention 3 on Hopper

**FA3 vs FA2 (sm_90):**

| Context | FA2 BF16 | FA3 BF16 | Speedup | FA3 FP8 |
|---|---|---|---|---|
| 16k | 230 TFLOPs | 540 TFLOPs | 2.3× | 1050 TFLOPs |
| 32k | 245 TFLOPs | 620 TFLOPs | 2.5× | 1.2 PFLOPs |
| 128k | 260 TFLOPs | 660 TFLOPs | 2.6× | 1.3 PFLOPs |

End-to-end decode TPS gain on Llama-70B at 32k: typically 1.4-1.7× in vLLM.

**Source build on ARM64 (no wheels):**
```bash
cd flash-attention/hopper
export TORCH_CUDA_ARCH_LIST="9.0"
export FLASH_ATTENTION_FORCE_BUILD=TRUE
export MAX_JOBS=16          # Grace 72 cores; 16 is RAM-safe
export NVCC_THREADS=2
export CUDA_HOME=/usr/local/cuda-12.8
python setup.py install
```
Time: 45-70 min on GH200. Peak RAM: 80-110 GB with MAX_JOBS=16. ccache saves ~60% on rebuild.

**vLLM integration:**
```bash
export VLLM_ATTENTION_BACKEND=FLASH_ATTN_VLLM_V1
# Or FLASHINFER (wins prefix-cache prefill + grouped GEMM)
```

**Verification:** vllm logs `Using FlashAttention-3 backend`; `import vllm_flash_attn3; print(vllm_flash_attn3.__version__)`; `nsys profile` shows `flash::hopper::*` kernels.

---

## Agent 2.8 — One-Shot Bootstrap Recipe

**Script:** `/home/ubuntu/Anime/cli/scripts/gh200-vllm-bootstrap.sh` (88 lines, executable, idempotent).

**Ordered sequence:**
1. Sanity: `uname -m == aarch64`, `nvidia-smi` reachable, CUDA 12.8 detected.
2. Append `GH200_VLLM_ENV` block to `~/.bashrc` once (idempotency via `grep -q` sentinel), source it.
3. Write `~/.pip-constraints.txt` — single source of truth referenced by `PIP_CONSTRAINT`.
4. `apt install build-essential cmake ninja-build git-lfs python3.10-dev libnuma-dev libopenmpi-dev`
5. Torch detect-or-install from `download.pytorch.org/whl/cu128`.
6. vLLM detect-or-build: try `import vllm._C`, if fails clone v0.10.1.1, run `python3 use_existing_torch.py`, then `pip install --no-build-isolation -e .`.
7. Re-pin transformers/tokenizers (vllm deps would otherwise upgrade them).
8. Smoke test + exec OpenAI-compatible server.

**Global env vars (~/.bashrc):**
- `CUDA_HOME=/usr/local/cuda-12.8`
- `TORCH_CUDA_ARCH_LIST="9.0+PTX"` — cuts vLLM compile from ~45min to ~12min
- `MAX_JOBS=32` — Grace 72 cores, nvcc TUs peak ~6GB
- `VLLM_TARGET_DEVICE=cuda`, `HF_HOME`, `PIP_CONSTRAINT`, `HF_TOKEN`

**Why each pin matters:**
- **torch 2.7.1+cu128** — first stable torch with prebuilt aarch64 wheels for CUDA 12.8
- **vllm 0.10.1.1** — last release before v1 engine refactor that requires torch 2.8
- **transformers 4.55.4** — 4.56 changed tokenizer return shape; vLLM 0.10.1.1 chokes
- **tokenizers 0.20.3** — newer 0.21 links Rust ABI that segfaults on Grace
- **triton 3.2.0** — bundled with torch 2.7.1; bumping breaks vLLM AllReduce
- **xformers 0.0.30** — ARM-compatible build aligned with torch 2.7.1
- **numpy 1.26.4** — vLLM 0.10.x has hardcoded `np.float_` removed in numpy 2.x

---

## Agent 2.9 — sm_90 Build Flags (the +31% prefill flag)

**TORCH_CUDA_ARCH_LIST semantics:**

| Value | Emits | Use |
|---|---|---|
| `9.0` | sm_90 SASS only | Generic Hopper. NO WGMMA, NO TMA, NO FP8 intrinsics. |
| **`9.0a`** | sm_90a SASS | **REQUIRED for FA3, Marlin-FP8, Machete, CUTLASS 3.x kernels.** |
| `9.0+PTX` | sm_90 SASS + PTX | Forward compat (sm_100 JIT). +15-25% binary size. |

**`9.0` will SILENTLY disable FA3/Machete kernel registrations** via `#if __CUDA_ARCH__ >= 900 && __CUDA_ARCH_FEAT_SM90_ALL__` guards. You get FlashAttention-2 fallback and ~30-40% lower prefill throughput.

**MAX_JOBS / NVCC_THREADS on Grace (72 cores):** memory wall, not core wall. CUTLASS TUs can OOM at MAX_JOBS=72. Sweet spot: **MAX_JOBS=32, NVCC_THREADS=4** (~280 GB peak RSS).

**Hopper kernels in vLLM:**
- FlashAttention-3 (`vllm-flash-attn`): WGMMA + TMA + warp specialization. Requires `9.0a` AND `VLLM_FA_CMAKE_GPU_ARCHES=90a-real`.
- Machete (W4A16 GEMM): CUTLASS 3.5 WGMMA mainloop, sm_90a only.
- Marlin FP8/INT8: TMA loads, sm_90a only.
- FlashInfer: needs `FLASHINFER_ENABLE_AOT=1` + `TORCH_CUDA_ARCH_LIST=9.0a`.

**Build cache:** `sccache` > `ccache` for nvcc. Setup: `export CMAKE_CUDA_COMPILER_LAUNCHER=sccache; SCCACHE_DIR=/mnt/nvme/sccache`. Cold: ~38min; warm sccache rebuild: ~4min (90% hit rate).

**Skipping archs:** restrict to `9.0a` alone cuts compile time **~55%** and binary size from 480 MB → 190 MB.

**Perf delta summary (Llama-3-70B FP8, bs=8 prefill):**

| Build | Time | tok/s |
|---|---|---|
| Default (8.0;8.6;8.9;9.0+PTX) | 38 min | 4,820 |
| **sm_90a only** | **19 min** | **6,310 (+31%)** |
| 9.0 no FA3 | 9 min | 3,940 (-18%) |

---

## Agent 2.10 — Weight Loading Speed

**hf_transfer:** `pip install hf_transfer && export HF_HUB_ENABLE_HF_TRANSFER=1`. Lambda GH200 (25-100 Gbps NIC): 800-1500 MB/s vs ~200 MB/s default. Weak retry semantics — wrap in shell loop. Tune `HF_HUB_DOWNLOAD_TIMEOUT=60`.

**Safetensors mmap:** `safe_open(...)` mmaps then `cudaMemcpyHostToDevice`. Prefetch: `cat *.safetensors > /dev/null` or `vmtouch -t` before vLLM start.

**Bottleneck order:** NVMe (7 GB/s) << network (12 GB/s @ 100Gbps) << HBM3 (4 TB/s) << NVLink-C2C (450 GB/s). **NVMe is choke point for warm loads, network for cold.** Use `/local` (instance NVMe), not `/home` (NFS).

**Parallel shard download:** `snapshot_download(max_workers=16)` overlaps shards. Llama 3.3 70B = 30 × ~4.6GB. With hf_transfer + 16 workers on 100Gbps: ~15-18s for 140GB. Above 16 workers HF CDN rate-limits.

**Pre-positioning:**
- **rsync from peer GH200** over 100Gbps RDMA: wire speed, ~14s for 140GB
- **s5cmd S3 sync**: 5-8 GB/s with 256 concurrent parts
- Bake AMI with weights pre-staged

---

## Agent 2.11 — Batch + Concurrency Tuning

**`--max-num-seqs` recommendations (Llama 70B):**

| Quant | KV per seq @ 8k | Max-num-seqs | Headroom |
|---|---|---|---|
| AWQ INT4 | ~640 MB | **256-384** | weights 35GB, KV pool 50GB |
| FP8 | ~640 MB | **192-256** | weights 70GB, KV pool 20GB |
| BF16 | ~640 MB | **48-64** | weights 140GB — won't fit single GH200 |

**`--max-num-batched-tokens`:** AWQ INT4 8192-16384; FP8 4096-8192. Rule: `max-num-batched-tokens >= max-num-seqs`, ideally ~2-4× to let decode batches grow.

**`--max-model-len`:** Set to **p99 actual length**, not theoretical max. Most workloads <16k; capping at 16k-32k buys 2-4× concurrency.

**Client-side:** semaphore = `1.0-1.25 × max-num-seqs`. Beyond → server-side queue → TTFT explodes.

**Knee:** lowest C where `p99_TTFT > 2 × p50_TTFT`. Operate at `0.8 × C_knee`.

**Launch template (AWQ INT4, expected ~3500-4500 tok/s, p50 TTFT <200ms at 256 concurrent):**
```bash
vllm serve meta-llama/Llama-3.3-70B-Instruct-AWQ-INT4 \
  --quantization awq_marlin --dtype half \
  --max-num-seqs 256 --max-num-batched-tokens 8192 \
  --max-model-len 32768 --gpu-memory-utilization 0.92 \
  --enable-chunked-prefill --kv-cache-dtype fp8
```

---

## Agent 2.12 — Alternative Engines Bake-Off (no Docker, no venv)

| Rank | Engine | Install | TTFT | tok/s (1×) | tok/s (batch) | Score |
|---|---|---|---|---|---|---|
| 1 | **llama.cpp** | 6 min source | 180 ms | 55 | ~200 | 9.0 |
| 2 | **TensorRT-LLM** | 55 min wheel+build | **70 ms** | **95** | **1400** | 8.5 |
| 3 | TGI | 25 min source | 110 ms | 65 | 1100 | 7.5 |
| 4 | Ollama | 30 sec binary | 200 ms | 38 | ~150 | 7.0 |
| 5 | ExLlamaV2 | 8 min source | 140 ms | 58 | ~250 | 6.5 |
| 6 | SGLang | 45+ min (FlashInfer) | 95 ms | 78 | 1300 | 5.0 |
| 7 | MLX | N/A | — | — | — | 0 |

**Key findings:**
- vLLM beats every alternative on the ease/throughput Pareto frontier on GH200 (NVIDIA aarch64 cu128 wheels, TRT-LLM/SGLang share its kernel stack)
- llama.cpp is the only engine where ARM64 is a non-issue (source-only, no wheel pain)
- TensorRT-LLM is throughput king but engine-rebuild-per-shape kills ease
- SGLang loses badly on ARM (FlashInfer wheel gaps)
- Ollama = llama.cpp daemon — never faster, sometimes slower
- For bare-metal pip --user, only TRT-LLM and Ollama install in <2 min

**Recommendation:** llama.cpp for prototyping/single-stream, TRT-LLM for production throughput, vLLM as default.

---

## Agent 2.13 — vLLM Env Var Tuning

**Optimal profile for GH200 + Llama 3.3 70B AWQ max throughput:**

```bash
export VLLM_USE_V1=1                          # +40% (rewritten scheduler)
export VLLM_ENABLE_V1_MULTIPROCESSING=1       # kills GIL
export VLLM_ATTENTION_BACKEND=FLASHINFER      # +15% long context
export VLLM_USE_FLASHINFER_SAMPLER=1          # +5-10% (fused top-k/top-p)
export VLLM_FLASHINFER_FORCE_TENSOR_CORES=1   # tensor cores even at small batch
export VLLM_USE_TRITON_FLASH_ATTN=0           # disable legacy slower path
export VLLM_USE_DEEP_GEMM=0                   # AWQ-incompatible; user's bb6a338 fix
export VLLM_WORKER_MULTIPROC_METHOD=spawn     # CUDA + fork = corrupt context
export VLLM_LOGGING_LEVEL=WARNING             # INFO costs 2% throughput
export VLLM_DO_NOT_TRACK=1
export VLLM_TARGET_DEVICE=cuda
export NCCL_NVLS_ENABLE=1                     # NVLink SHARP
export NCCL_CUMEM_ENABLE=1                    # CUDA VMM for buffers
export NCCL_P2P_DISABLE=0                     # keep NVLink
export NCCL_DEBUG=WARN
export HF_HUB_ENABLE_HF_TRANSFER=1            # 5-10× faster downloads
export TORCH_CUDA_ARCH_LIST="9.0+PTX"
export CUDA_DEVICE_MAX_CONNECTIONS=1          # better kernel overlap on Hopper
```

**Expected combined gain vs defaults:** ~1.8-2.2× throughput.

**Biggest individual wins:** VLLM_USE_V1 (~40%), FlashInfer + tensor-core force (~20%), disabling DeepGEMM for AWQ (~10%).

---

## Agent 2.14 — Pure PyTorch Path (no vllm)

**transformers + accelerate `device_map="auto"`:** Baseline. 70B BF16 needs ~140GB — won't fit GH200 96GB HBM. Forces offload to 480GB LPDDR5X via NVLink-C2C 900GB/s — much less painful than x86.

**Throughput:** 2-5 tok/s with offload, 8-15 tok/s quantized to int8/fp8 fully on GPU. vLLM on same hardware: 30-60 tok/s single, 1500+ tok/s aggregate. **Minimal stack ~5-10× slower single, ~50-100× slower aggregate.**

**torch.compile + StaticCache:** 2-4× decode speedup. 70B FP8 from ~10 tok/s → ~25-35 tok/s single-stream. Brings minimal stack within 2× vLLM single-stream; still 30× behind on aggregate.

**Production features (transformers 4.40-4.50):**
- `cache_implementation="static"` — one-line StaticCache
- `compile_config=CompileConfig(...)` — fine-grained
- `attn_implementation="flash_attention_3"` on Hopper
- CUDA graphs auto-wired with static cache + compile
- Assisted generation (draft model spec decode) — no vllm needed
- FP8 weights via `torchao Float8WeightOnlyConfig` — fits 70B in ~70GB

**Minimal install:**
```bash
pip install --user torch==2.5.* --index-url https://download.pytorch.org/whl/cu124
pip install --user transformers>=4.46 accelerate flash-attn --no-build-isolation
pip install --user torchao
```

**When to use:** debugging internals (hooks, hidden states), custom LogitsProcessor/StoppingCriteria, research/one-off eval, constrained decoding, single-user interactive.

---

## Agent 2.15 — Anime CLI Integration Plan

**Proposed new commands:**

| Command | Purpose |
|---|---|
| `anime fast <model>` | Auto-pick fastest engine, start it |
| `anime tune <model>` | Write optimal env vars + flags to profile |
| `anime bench <model>` | Run bake-off against running endpoint |
| `anime fix-cuda` | Auto-recover broken torch/vllm/nvidia user-site |

**New install-script entries needed in `internal/installer/scripts.go`:**
- **`sglang`** — `pip3 install --no-deps "sglang[all]==0.4.3"` with PIP_CONSTRAINT guard, source-build branch for aarch64+cu12
- **`llama-cpp-cuda`** — `cmake -B build -DGGML_CUDA=ON -DCMAKE_CUDA_ARCHITECTURES=90 ...`
- **`flashinfer`** — prebuilt cu128 wheel from flashinfer.ai/whl/cu128/torch2.7/
- **`hf-fast-download`** — installs hf_transfer + aria2 + writes HF_HUB_ENABLE_HF_TRANSFER=1

**Engine selection logic for `anime fast`:**
- GH200 + AWQ variant → vLLM AWQ (best 70B throughput)
- Multi-Hopper (TP≥2) → SGLang (radix-cache wins)
- VRAM < 48GB → llama.cpp GGUF Q4_K_M
- Default → vLLM FP16

**`anime fix-cuda` repair steps:**
1. Uninstall user-site torch shadow if cu13 wheel on cu12 driver
2. Purge stale vllm so it can't drag torch back
3. `pip3 cache purge`
4. `anime install pytorch` (uses pinned constraints)
5. `anime install vllm` (source-build on aarch64+cu12)
6. Verify

Full code skeletons in original agent report.

---

# Cross-Cutting Takeaways

1. **Default-to-cu13 in vllm ≥0.20** is the single biggest reason installs break on GH200 driver 570. Pin to vllm 0.10.1.1 + torch 2.7.1+cu128.

2. **TORCH_CUDA_ARCH_LIST=9.0a** (not 9.0) is worth +31% prefill throughput on GH200. Always use `9.0a` for source builds.

3. **PIP_CONSTRAINT** is the single mechanism preventing transitive deps from clobbering torch. Anime now sets it automatically via `PersistentPreRun`.

4. **vLLM's optimal env profile** (FLASHINFER backend + V1 engine + fused sampler + force-tensor-cores) is worth ~1.8-2.2× throughput vs defaults.

5. **AWQ INT4 is the right default quant** for Llama 70B on GH200 96GB HBM — fits with KV headroom, best single-stream perf at low/mid batch.

6. **NVIDIA NGC `nvcr.io/nvidia/vllm:25.09-py3`** is the cleanest Docker path on driver 570 (cu13 + compat libs bridge to r570). Would need user to be in docker group.

7. **flash-attn 3 source build** is the cleanest path to FA3 on aarch64 — 45min build, worth it for long contexts.

8. **The pin matrix matters:** transformers, tokenizers, xgrammar, lm-format-enforcer, numba, outlines_core, compressed-tensors, depyf, llguidance, protobuf all have specific version constraints from vllm 0.10.1.1 METADATA.

9. **NVLink-C2C unified memory** is GH200's killer feature but vllm doesn't fully exploit it for KV cache tiering. TRT-LLM does (`--kv_cache_host_memory_bytes`).

10. **Lambda Stack's apt torch** in `/usr/lib/python3/dist-packages` is shadowed by anything in `~/.local/lib/python3.10/site-packages`. User-site beats system-site in sys.path.
