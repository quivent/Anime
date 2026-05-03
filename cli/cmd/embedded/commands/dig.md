# /dig — MLX Optimization Loop

You are entering the MLX Threadhouse at `~/mlx-th`. Read `~/mlx-th/README.md` first to understand why you are here.

## Context

This slash command is a self-referencing optimization loop. You investigate, implement, benchmark, document, and then invoke `/dig` again. The loop does not stop.

## Workspace

- `~/mlx-th/README.md` — Mission, identity, sentence
- `~/mlx-th/LOG.md` — Running optimization log
- `~/mlx-th/BENCHMARKS.md` — All measurements
- `~/mlx-th/agents/` — Agent definitions (warden, profiler, kernel-engineer, benchmarker, auditor)
- `~/mlx-th/logs/` — Per-cycle investigation logs
- `~/mlx-th/benchmarks/` — Raw benchmark data
- `~/mlx-th/patches/` — Optimization patches

## The Cycle

Execute these steps in order. Do not skip steps. Do not skip benchmarks.

### Step 1: Orient

Read `~/mlx-th/LOG.md` and `~/mlx-th/BENCHMARKS.md`. Know the current state. What cycle is this? What's the current best tok/s? What was tried before?

### Step 2: Investigate (Profiler Agent)

Launch a subagent as the Profiler (`~/mlx-th/agents/profiler.md`). Tasks:
- Check hardware specs if not already documented
- Profile MLX's Metal kernel dispatch for the benchmark model
- Search for Apple Metal optimization guides, M4 GPU architecture docs
- Search for recent MLX GitHub issues, PRs, and discussions about performance
- Search for llama.cpp Metal kernel optimizations that MLX hasn't adopted
- Read the MLX source code for the quantized matmul kernel
- Identify the top 3 time-consuming operations

Write findings to `~/mlx-th/logs/profile-cycle-N.md`

### Step 3: Ideate (Kernel Engineer Agent)

Based on profiler findings, produce **5-15 optimization possibilities**. For each:
- Name and one-line description
- Which component it targets (kernel, memory, dispatch, algorithm)
- Estimated difficulty (trivial / moderate / hard)
- Rationale — WHY this should help, grounded in profiler data
- Risk — what could go wrong

Document in `~/mlx-th/logs/ideation-cycle-N.md`

### Step 4: Cross-Reference (Auditor Agent)

Launch the Auditor to review ideations:
- Does each rationale hold up against measured data?
- Has this been tried before? Check prior cycles.
- Are the difficulty estimates realistic?
- Flag any speculation.

Append audit to `~/mlx-th/logs/ideation-cycle-N.md`

### Step 5: Select and Implement

Pick the top 1-3 optimizations that passed audit. For each:
- Implement as a patch, script, or configuration change
- Save to `~/mlx-th/patches/cycle-N-optimization-name/`
- Document exactly what was changed

### Step 6: Benchmark (Benchmarker Agent)

Run the standard benchmark protocol:
1. Baseline (unmodified MLX): 3 runs, median
2. With optimization: 3 runs, median
3. Record both to `~/mlx-th/benchmarks/cycle-N/`
4. Update `~/mlx-th/BENCHMARKS.md`

### Step 7: Document

Append cycle results to `~/mlx-th/LOG.md`:
- What was investigated
- What was tried
- What the numbers say
- What opened up for next cycle

### Step 8: Warden Check

Invoke the Warden (`~/mlx-th/agents/warden.md`):
- Is the cycle complete?
- Are all benchmarks recorded?
- Any regressions?

### Step 9: Dig

Say: "Cycle N complete. Invoking /dig."

Then invoke `/dig` to begin the next cycle.

## Rules

- No claims without benchmarks
- No benchmarks without protocol
- No optimizations without profiler data
- No skipping the auditor
- The loop does not stop
