# /lithos-bench — Lithos ARM64 Primitive Optimization Protocol

Activate the ARM64 blob optimization loop. Beat GCC -O2 on every primitive and n-gram. Record every change.

---

## FAST LOAD

Working directory: `/Users/joshkornreich/lithos/arch/arm64/`

Files that matter:
- `lithos-bench.s` — THE blobs. Every primitive lives here.
- `bench-primitives.c` — The harness. ONLY benchmark tool.
- `bench-full-lithos-fns.s` — Real-program Lithos functions (fibonacci, factorial, gcd, etc.)
- `bench/three-way/*/run` — Three-way benchmarks: C/GCC vs Sixth vs Lithos

Tracking:
- `/Users/joshkornreich/.supervision/incoming.md` — mandatory completion reports
- `bench/three-way/` — three-way suite (fibonacci, factorial, collatz, gcd, dotprod, sha256, sha512, aes, sum_squares)

---

## PROTOCOL: PRIMITIVE LOOP

**Measure → Identify → Disassemble GCC → Fix → Measure → Record**

### Step 1: Baseline

```bash
cd /Users/joshkornreich/lithos/arch/arm64
gcc -O2 -o bench-primitives bench-primitives.c lithos-bench.s -lm
./bench-primitives
```

Identify all rows where `ratio > 1.10` and parity = `no`.

### Step 2: Identify the blob

```bash
grep -n "^_lithos_PRIMNAME" lithos-bench.s
```

Read the blob (usually 10-30 lines).

### Step 3: Disassemble GCC's baseline

Generate GCC's version of the same operation for comparison:

```bash
cat << 'EOF' > /tmp/prim.c
static void g_op(float*a,float*b,float*d,long c) { for(long i=0;i<c;i++)d[i]=a[i]+b[i]; }
EOF
gcc -O2 -S -o /tmp/prim.s /tmp/prim.c && cat /tmp/prim.s
```

Key patterns GCC uses on ARM64 that often beat naive ld1/st1:
- `ldp q0, q1, [x0, #-32]; ldp q2, q3, [x0], #64` — pre-offset avoids post-inc anti-deps
- `stp q0, q1` — store pair, better bandwidth than `st1`
- 2-chain MUL unroll — MADD hides 3-cycle latency for multiply-accumulate
- `UDIV + MSUB` — remainder in one step (a mod b = a - (a/b)*b)
- `CBZ` after `MSUB` on the remainder (not before UDIV) — skips 2 MOVs on final GCD step

### Step 4: Fix

Edit `lithos-bench.s`. Common fixes:
- Replace `ld1 {v0-v3}` + `st1` with `ldp q / stp q` using pre-offset pointers
- Add 2× unroll with register interleaving
- Move loop-exit check to after the compute (not before) to skip redundant MOVs

Rebuild and rerun immediately:
```bash
gcc -O2 -o bench-primitives bench-primitives.c lithos-bench.s -lm && ./bench-primitives
```

### Step 5: Record

Every change that moves a primitive from `no` to `YES` gets reported:

```bash
~/.supervision/report "[<ISO timestamp>] lithos/arm64 DONE: <prim> fixed — <ratio before> → <ratio after>"
```

Example:
```bash
~/.supervision/report "[2026-05-06T18:30:00] lithos/arm64 DONE: tadd (++++) — 1.19x no → 0.97x YES (pre-offset ldp/stp)"
```

---

## PROTOCOL: N-GRAM LOOP

**Analyze corpus → Identify top sequences → Create three-way benchmark → Optimize function → Record**

### Step 1: Corpus analysis

```bash
python3 /tmp/ngrams.py   # script lives at this path from the last session
```

Or re-derive from: `/Users/joshkornreich/lithos/bench/lithos/*.ls` (254 programs).

Key findings (as of 2026-05-06):
- Corpus: 190 trivial (`main ⇌ + - ? * /`) + 64 interesting programs
- Top pure arithmetic bigrams: `* +` (8 files → MADD), `/ -` (16 → UDIV+MSUB), `→ -` (13)
- Top trigrams: `→ - ?` (15), `/ - ?` (11), `- ? +` (8)
- Key: `?` is a call boundary — optimize within contiguous arithmetic runs

### Step 2: Three-way benchmark

Each n-gram pattern maps to a benchmark in `bench/three-way/`:

| n-gram | benchmark | notes |
|--------|-----------|-------|
| `+ ?` | fibonacci | pure recursion |
| `* + ?` | factorial | MUL+ADD+recurse |
| `→ - ? * + + - ?` | collatz | complex |
| `* +` | dotprod | MADD fused |
| `/ - ?` | gcd | UDIV+MSUB+recurse |
| `→ → → →` | (future) | quad LDP |
| `← ←` | (future) | STP |

Create new suite: `bench/three-way/<name>/` with `harness.c` + `run` (executable).

harness.c requirements:
- Use `volatile` inputs to prevent constant-folding
- `__attribute__((noinline))` on GCC baseline functions
- Use `HZ = 24000000.0` (mach_absolute_time: 3/125 GHz on M4)
- Output lines: `C/GCC -O2: X.X Mops/s`, `Lithos HW: Y.Y Mops/s`, `Sixth: Z.Z Mops/s`

### Step 3: Optimize the function

Edit `bench-full-lithos-fns.s`. The function signature from bench/three-way/harness.c's `extern` declaration tells you the ABI.

Reference implementations already in `bench-full-lithos-fns.s`:
- `_lithos_fibonacci` — `+ ?` (simple recursion)
- `_lithos_factorial` — 2-chain MUL unroll
- `_lithos_collatz_sum` — branch-predicted step
- `_lithos_gcd_sum` — CBZ-after-MSUB exit
- `_lithos_dotprod` — 4-accumulator MADD

### Step 4: Record

```bash
~/.supervision/report "[<ISO ts>] lithos/arm64 DONE: <ngram> three-way — C=Xops/s Lithos=Yops/s (Z×)"
```

---

## INVIOLABLE RULES

1. **ARM64 only**. No C includes, no libm fallbacks, no intrinsics headers in blobs.
2. **bench-primitives is the only primitive benchmark**. One run per change.
3. **Blobs in `lithos-bench.s` are source** — grep before claiming anything doesn't exist.
4. **Every completed change is reported** to `~/.supervision/incoming.md` via the `report` command.
5. **No documentation** unless explicitly asked. No `.md` files beyond this one.

---

## CURRENT STATE (as of 2026-05-06)

```
Primitives (bench-primitives):
  All YES except run-to-run noise (borderline primitives flip ≤ 3% between runs).
  ++++  (tadd) — FIXED 2026-05-06: pre-offset ldp/stp — was 1.15-1.19x, now 0.97x

Three-way benchmarks:
  fibonacci   — Lithos competitive
  factorial   — 2-chain MUL unroll (0.80→1.06x)
  collatz     — Lithos faster than GCC (1.23×) — branch prediction beats GCC's branchless csinc
  gcd         — 0.95× (CBZ-after-MSUB, FIXED 2026-05-06)
  dotprod     — MADD 4-acc
  sha256/sha512/aes — crypto
  sum_squares — GCC vectorizes; Lithos scalar 0.42×

N-gram analysis complete. Top unbuilt three-way target: →→→→ (quad LDP, life.ls pattern).
```

---

## ACTIVATION

When `/lithos-bench` is invoked:

1. Read `lithos-bench.s` header and scan for blob list.
2. Run `bench-primitives` and identify failures.
3. If no failures: run all three-way benchmarks and report ratios.
4. Announce state and propose the single highest-value next fix.

Format:
```
LITHOS-BENCH active.

Primitives: N/M YES — failing: [list]
Three-way:  fibonacci=X×, factorial=X×, gcd=X×, ...
Next: [single most valuable action]
Moving.
```
