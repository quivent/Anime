---
description: "Load Lithos language understanding — primitives, font tables, kernel bypass"
---

# ENFORCEMENT — LITHOS MODE ACTIVE

When this skill is loaded, the following rules are **INVIOLABLE**:

## ALLOWED — write, edit, execute ONLY these file types:
- `.ls` — Lithos source (UTF-8 primitives and compositions)
- `.lsb` — Lithos binary (1-byte-per-primitive canonical form)
- `.lt` — Lithos font table (glyph → blob mapping)
- `.li` — Lithos intermediate (compiler output)
- `.lion` — Lithos cargo (weight cartography)
- Font table compilation scripts (`build-*.sh`, `Makefile` targets for lithos)
- Shell commands that invoke lithos tools (`lithos`, `dispatch`, `forge`, `lint-ls.sh`)

## FORBIDDEN — do NOT write, edit, or execute:
- `.air` — Apple IR (output of the font table, not your concern)
- `.metal` — Metal shading language (the font table produces this, not you)
- `.metallib` — Metal library archive (build artifact)
- `.cu` / `.cuda` — CUDA C++ (the font table produces PTX, not you)
- `.ptx` — PTX assembly (font table output)
- `.m` / `.swift` — Host language files
- Any GPU shading language directly

**You write lithos. The font table outputs machine code. If you find yourself
writing C, CUDA, Metal, or any host language — STOP. You are doing it wrong.
Express the computation in `.ls` and let the compiler + font table handle the rest.**

---

# Lithos — Language Identity

A program is a sequence of symbols. Each symbol maps to a machine code blob
in a font table. The compiler concatenates blobs. The output runs on silicon.

There are no functions, no variables, no strings, no numbers, no reserved
words, no constants. There are only symbols, a small set of types, and
compositions of symbols.

Two facts govern everything:

1. **The font table IS the compiler.** A symbol enters. A blob exits. That's
   compilation. Parsing, optimization, register allocation, linking — ceremony.
2. **The byte IS the dispatch.** Input value indexes the font table. The blob
   at that index runs. No dispatcher, no event loop, no switch statement.

## The 24 Primitives

### Arithmetic (rank by symbol count)

| Scalar | Vector | Matrix | Tensor | Op |
|--------|--------|--------|--------|----|
| `*` | `**` | `***` | `****` | Multiply |
| `+` | `++` | `+++` | `++++` | Add |
| `-` | `--` | `---` | `----` | Subtract |
| `/` | `//` | `///` | `////` | Divide |

### Reductions

| Symbol | Meaning |
|--------|---------|
| `Σ` | Sum reduction |
| `△` | Max reduction |
| `▽` | Min reduction |
| `#` | Index prefix (`# △ x` = argmax) |

### Rank-changing

| Symbol | Meaning |
|--------|---------|
| `·` | Inner product — contracts rank. `· u v` = scalar, `· A v` = vec, `· A B` = mat |
| `⊗` | Outer product — expands rank. `⊗ u v` = mat |

### Memory and Registers

| Symbol | Meaning |
|--------|---------|
| `→` | Load from memory (width ladder: `→` byte, `→→` 64-bit, `→→→` 128-bit, `→→→→` NEON) |
| `←` | Store to memory (same width ladder) |
| `↑` | Read register |
| `↓` | Write register |

### Scalar Math (SFU)

| Symbol | Meaning |
|--------|---------|
| `⅟` | Reciprocal (1/x) |
| `√` | Square root |
| `log₂` | Log base 2 |
| `∿` | Sine |
| `∾` | Cosine |
| `^` | Power (prefix: `^ base exp`) |
| `ln` | Natural log |

### Control

| Symbol | Meaning |
|--------|---------|
| `?` | Conditional |
| `↻` | Loop |

That is the whole language. Primitive symbols. Everything else is a composition.

## The Compiler

`compiler/lithos.s` — 42KB, 1351 lines of ARM64 assembly. Self-contained.
Four targets via font table selection:

```
./compiler/lithos input.ls output.bin              # ARM64 (default)
./compiler/lithos input.ls output.wasm --wasm      # WebAssembly
./compiler/lithos input.ls output.archive --agx    # Apple GPU raw bytes
./compiler/lithos input.ls output.ll --air         # LLVM IR -> metallib -> M4 GPU
```

Core loop: read one `.lsb` opcode byte → index active font table (X3 register) →
copy blob to output → bump register allocator → patch register slots.

Produces a 69KB binary. GCC is 15M lines.

Source formats:
- `.ls` — UTF-8 text, human-editable (Unicode codepoints for primitives)
- `.lsb` — canonical 1-byte-per-primitive binary (direct array index into blob table)
- Lossless interconversion between the two

## Font Tables

A font table maps symbols to machine code blobs. One table per target.

### ARM64 Font (`arch/arm64/arm64-font.s`, 39KB)

256-entry dispatch table. 24 core primitives + 128 composites.
NEON vector ops (FMLA, FMLS, ADDV, SMAXV, FADDP).
SME2 streaming (0x4A-0x4E): FMOPA matmul (174 GFLOPS FP32, 1554 GFLOPS FP16).
Field arithmetic: x25519 GF(2^255-19) — fe_add, fe_sub, fe_sq, fe_mul.

### ARM64 N-gram Table (`arch/arm64/arm64-ngram2.s`, 60KB)

65,536-byte (256x256) 2-gram fusion lookup. Compile-time adjacent-pair
detection emits a single composite blob instead of two sequential ops.
128 discovered patterns: scalar MADD, NEON load+arithmetic, vector fused ops.
Hardware `√⅟` → FRSQRTE (one instruction replaces two).

### GPU AIR Font Atlas (`arch/gpu/air-fragments/`, 76 files)

46 `.air` blobs + 27 `.metal` sources → `lithos-font.metallib` (339KB).
Built via `build-air-font.sh` (`xcrun metal -c` → `xcrun metallib`).
Verified by `verify-air-font.py` (64 functions, 27 required primitives).

Opcode map (`air-font-table.tsv`):
- **0x20-0x2F:** arithmetic (scalar/vector/matrix/tensor `* + - /`)
- **0x40-0x45:** reductions (`Σ △ ▽` via simd_sum/simd_max/simd_min)
- **0x60-0x65:** SFU (`⅟ √ ∿ ∾ ^`)
- **0xA0-0xB5:** fused n-grams (`σ*` silu, `η` rmsnorm, `·q4` matvec_q4, attention, rope, qkv_fused, ffn_fused)

### DSP Font (`arch/arm64/arm64-dsp-font.s`, 21KB)

Specialized font for audio synthesis (Quantum DAW).

### GLSL / WASM Fonts (`packages/paint/wasm-font.mjs`, `wgsl-font.mjs`)

Map primitives to GLSL intrinsics / WASM bytecodes for the renderer.

## Cargo Format (.lion)

Binary weight file. 80-byte header + raw tensor cargo. Zero JSON at runtime.

**Header (80 bytes, all u64 LE):**
- Magic + version (8B)
- `schema_id` (1=Q4_K_M, 2=MLX 4-bit)
- `n_layers, d_model, head_dim, n_kv_heads, d_ff, vocab` (6 x 8B)
- SHA-256 hash truncated to 16B

**Llama 3.3 70B Q4:** 80 layers, d=8192, head_dim=128, n_kv=8, d_ff=28672, vocab=128256, ~42 GiB.

Cargo is raw tensor bytes in canonical schema order, 64-byte aligned. No
per-tensor headers. Kernel compiles with literal byte offsets — no runtime
name lookup. One mmap call replaces the entire safetensors/torch/vLLM stack.

## GPU Kernels (Production)

### uber.metal (Ubershader v3)

Single PSO, 7 chained operations per token:
1. K1 fused: RMSNorm + Q/K/V triple Q4 matvec (TG=32)
2. ATTN_ROPE: online softmax with per-head RoPE (TG=32, eliminates 4 barriers)
3. Q4MV_RES: Q4 matvec + residual (TG=256)
4. K2: FFN gate/up Q4 matvec (TG=32)
5. HEAD_RMSN: final layer norm (TG=256)
6. HEAD_LOGITS: vocabulary logits (TG=32)
7. ARGMAX_EMBED: token selection via simd reduction tree (TG=256)

### persistent_megakernel.metal

Daemon kernel. 20 TGs x 1024 threads. Ring-buffer I/O. Atomic counter
cross-TG sync (monotonic epochs). Never returns — polls input, runs
forward pass, writes output.

### weight_stationary.metal

Carver Mead + Carmack architecture: weights as wiring, tokens as signals.
Weights stay L2-resident across tokens. 13,300 tok/s compute-bound target.

All kernels use Q4 MLX format (uint32 packed nibbles + per-group FP16
scales/biases). No standalone dequant — fused inline in the matmul.

## Runtime

### lithos-runtime.m (1222 lines, canonical)

Loads 25 kernel functions from `lithos-font.metallib`. Two execution paths:
- **ICB path** (default): Pre-records 276 dispatches into MTLIndirectCommandBuffer;
  ONE commit+wait per token eliminates CPU-GPU round-trips
- **Per-dispatch** (debug): 116 commit+wait cycles, ~3.5x slower

Manages: working buffers (x, xn, q, k, v, attn, o, gate, up, down, logits),
KV cache (2x MAX_SEQ x DKV), weight file mmapped at ~42 GiB for 70B.

Canonical verification: BOS(128000) -> 16309 (correct).

### Decode Harnesses (`decode/`, 39 files)

**1B:** `infer-fused-atlas.m` (AIR font atlas, 81 dispatches/token)
**70B:** `lithos-infer-70b.m` (uint64_t offsets, HEAD_DIM=128 2-pass attention)
**Probes:** 7 isolated kernel tests (embed, lm_head, rope, softmax_kv, q4mv, rmsn, argmax)
**AIR proofs:** 20 LLVM IR intermediate validation files

## Measured Performance

### 1B (Llama 3.2 1B Q4, M4 Max)

| Config | tok/s | Notes |
|--------|-------|-------|
| ubershader ICB | **494.7** | Fastest. Single ICB, Q4 embedded FP16 |
| megakernel 40 TG | **245** | One dispatch, 4.07ms best, 225/225 correct |
| FP16 embedded | **149.8** | Pre-dequanted FP16, no Q4 ALU |
| FP16 .lion | **127.7** | FP16 weights via .lion |
| AIR font atlas ICB | **82** | 276 dispatches, 15/15 correct |
| vs MLX baseline | **62.6 vs 61.0** | 1.03x win at same sha |

### 70B (Llama 3.3 70B Q4, M4 Max)

| Config | tok/s | Notes |
|--------|-------|-------|
| lithos-uber-70b best | **10.524** | Production ubershader |
| lithos-uber-70b w/ prefill | **8.864** | 1117ms prefill overhead |
| ObjC scaffolding baseline | **7.91** | Pre-Lithos |
| Python scaffolding baseline | **8.9** | Pre-Lithos |
| Bandwidth ceiling | ~14 | 42 GiB / 3 GB/s |
| Weight-stationary theoretical | **13,300** | Compute-bound target |

### Portable Kernel Benchmarks (Lithos vs GCC -O2)

| Kernel | Ratio | Notes |
|--------|-------|-------|
| Softmax attention | **4.514x** | Schraudolph fast-exp trick |
| CSS token resolution | **1.396x** | |
| Raycast | **1.119x** | |
| UTF8 decode | **1.005x** | Parity |
| **Average across 630 benchmarks** | **1.51x** | 555 GFLOPS matrix ops |

## Compositions (NOT Part of the Language)

| Alias | Definition |
|-------|------------|
| `σ` | `* -1  ^ e  + 1  ⅟` (sigmoid) |
| `amplify` | `σ *` (SiLU/swish) |
| `η` | `** Σ / D  √  ⅟  **  ** w` (RMSNorm) |
| `ς` | `^ e  + 1  ln` (softplus) |
| `‖` | `· √` (L2 magnitude) |
| `direction` | `· √ ⅟ **` (unit vector) |

## .ls Programs

| Program | Primitives | Purpose |
|---------|-----------|---------|
| `decode/megakernel-v0.ls` | `+` | Proof-of-origin, byte-dispatch evidence |
| `decode/matmul_tiny.ls` | `*` `→` `←` | Scalar a*b, 52 AGX bytes |
| `decode/rmsnorm_tiny.ls` | `**` `Σ` `/` | RMS normalization, simd_sum verified |
| `decode/embed_lookup.ls` | `→` `*` `+` `←` | Token-to-embedding gather |
| `decode/residual_add.ls` | `++` `→` `←` | Elementwise vector add |
| `decode/lm_head_argmax.ls` | `·` `△` `#` | Final projection + argmax |
| `decode/decode_token.ls` | `-` `?` `↻` `←` `↑` | Persistent daemon loop |

## WRONG vs RIGHT

```
WRONG: sigmoid x         RIGHT: σ
WRONG: gate x            RIGHT: σ
WRONG: silu x            RIGHT: amplify
WRONG: rmsnorm x w D     RIGHT: η
WRONG: normalize x w D   RIGHT: η
WRONG: softplus x        RIGHT: ς
WRONG: matvec A v        RIGHT: · A v
WRONG: reduce_sum x      RIGHT: Σ
WRONG: e^ x              RIGHT: ^ e x
WRONG: 1/√ x             RIGHT: √ ⅟
WRONG: dup / acc / swap   RIGHT: (never — Forth is not Lithos)
WRONG: * ↑ 0             RIGHT: * (register index is cargo, not source)
```

### Absolute prohibitions

- **NO C** for bootstrap, lexer, parser, emitter, or any tool processing `.ls` files.
- **NO Python, NO Sixth, NO Objective-C** for the same.
- **NO numbers in source.** Numbers are data. Data is cargo (`.lion`) or font table.
- **NO English function names.** If it isn't one of the primitives or a named composition, it doesn't exist.
- **NO inventing.** If it isn't in the primitive table, it's not Lithos.
- **NO Forth patterns.** `dup`, `swap`, `acc`, `variable`, `does>` — Forth is not Lithos.

### Naming in tools

Even in host-language code, name glyph-referentially:
- **Right:** `T_STAR`, `T_SIGMA`, `T_ARROW_RIGHT`
- **Wrong:** `TOK_MUL`, `TOK_SUM`, `TOK_LOAD`

## Packages

### fterm (audio synthesis — terminal UI deleted)

Audio engine active: `audio-dylib.s` (452 lines, CoreAudio glue + builtin
sine), `audio-render.c` (121 lines, AudioUnit setup), `poly-sine.s` (138 lines,
4-voice NEON polyphonic sine). Quadrature oscillator proven at 440Hz.
756-byte synthesis engine as ARM64 blobs. 578x headroom at 48kHz on M4.

### paint (GLSL/WASM rendering)

Server-side SDF emitter specializes GLSL shaders by camera/frustum.
Compile-time culling: GPU never sees code for invisible objects.
WASM-native WebGL2 runtime (~1800 lines hand-assembled WASM). Zero JS per frame.
Three scenes: Virgo, Taurus, Capricorn. 128MB -> 31MB heap (Three.js eliminated).

## Domain Examples

### Inference

```
· ** Σ
```

Inner product, elementwise multiply, sum reduction. Core of every transformer
layer. Font table blobs handle Q4 dequant fused in the matmul.

### Audio (`packages/fterm/audio-blobs.s`)

```
∿ ← → ? ← ↻
```

Sine oscillator generates sample, store to buffer, load state, branch on
envelope stage, store level, loop per sample. 8 blobs, 756 bytes.

### Renderer

```
*
```

Multiply. Scale factor times geometry. The font table blob knows the camera
matrix, the projection, the viewport. The program is one symbol.

## The Voice

**On font tables:** Every layer of computing is a font table. The compiler
didn't use font tables as a trick. Font tables are what compilers always were.

**On simplification:** If the program is longer than a few symbols, you're
encoding data as code. Find the data. Move it out. A terminal is three symbols.
An inference engine is three symbols.

**On velocity:** Build the thing. Then build it better. Then build it again.
Don't plan the perfect version. Build the wrong one fast.

**On what Lithos is:** Not a programming language. An observation about
computation: every transform is a table lookup. Input, table, output.

## Project Status

- **L1 (Correctness):** CLOSED (13/13 checkpoints)
- **L2 (Velocity):** ACTIVE — gate target 16 tok/s on 70B, 6 phases, 18 checkpoints
- **L3-L10:** LOCKED (gated behind L2)
- **Mission:** Llama 3.3 70B Q4 on M4 Max. Target: the moon.
- **Current 70B:** 10.5 tok/s (ubershader), 8 tok/s (scaffolding baseline)

## Build

```bash
make                    # compiler + bench harnesses (~30s)
make app                # Swift IDE (LithosCode.app)
make air                # GPU font atlas (lithos-font.metallib)
make bench-all          # run 619 benchmarks
```

Requires: Xcode CLT (`as -march=armv9-a+sme`, `ld`, `xcrun`), Rust, Python 3.

## Quick Reference

| Item | Location |
|------|----------|
| Language spec | `docs/language/CANONICAL-SYNTAX.md` |
| Compiler | `compiler/lithos.s` (42KB, 1351 lines ARM64) |
| ARM64 font | `arch/arm64/arm64-font.s` (39KB, 256 entries) |
| N-gram fusion | `arch/arm64/arm64-ngram2.s` (60KB, 65536-byte table) |
| DSP font | `arch/arm64/arm64-dsp-font.s` (21KB) |
| Compositions | `arch/arm64/lithos-compositions.s` (73KB) |
| GPU font atlas | `arch/gpu/air-fragments/` (76 files) |
| Metallib | `arch/gpu/lithos-font.metallib` (339KB) |
| Font table map | `arch/gpu/air-font-table.tsv` |
| Atlas builder | `arch/gpu/build-air-font.sh` |
| Ubershader | `host/tools/agx/kernels/uber.metal` |
| Megakernel | `runtime/persistent_megakernel.metal` |
| Weight-stationary | `runtime/weight_stationary.metal` |
| Runtime | `runtime/lithos-runtime.m` (1222 lines) |
| 70B harness | `decode/lithos-infer-70b.m` |
| 1B fused atlas | `decode/infer-fused-atlas.m` |
| Decode programs | `decode/*.ls` (7 kernel programs) |
| Probes | `decode/probes/*.ls` (7 isolated tests) |
| AIR proofs | `decode/air-proofs/*.ll` (20 LLVM IR files) |
| Cargo format | `.lion` files (80-byte header + raw tensors) |
| Lion converter | `host/tools/lion/lion-convert.fs` |
| Bench results | `bench/packages/results.jsonl` (597 records) |
| Bench harness | `host/tools/bench.py` |
| Lint | `tools/lint-ls.py` (validates .ls primitives) |
| Port index | `PORT-INDEX.ls` (926 lines, status + blockers) |
| Binary ports | `bench/lithos/*.lsb` (254 ported benchmarks) |
| CPU library | `cpu/*.lsb` (16 modules) |
| IDE (Swift) | `host/swift/lithos-code/` (4668 lines, 18 files) |
| IDE (Rust TUI) | `host/tools/lithos-code/tui/` (3092 lines) |
| Paint renderer | `packages/paint/` (GLSL/WASM SDF) |
| Audio engine | `packages/fterm/audio-dylib.s` |
| Pipeline doc | `docs/PIPELINE.md` |
| Runtime docs | `docs/runtime/` (19 docs) |
| AGX docs | `docs/apple/` (phase-1 DAG, assembly audit) |
| Compiler design | `docs/compiler/design/byte-dispatch.md` |
| Archive | `archive/` (quarantined: Python, Sixth, ObjC proofs) |
| Automation | `.overnight/` (10k tok/s campaign, 10 pending tasks) |
| Project rules | `CLAUDE.md` at repo root |
