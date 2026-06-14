---
description: Compare two programs head-to-head — same workload, same machine, measured resource consumption with verdict
argument-hint: <program-A> <program-B> [--workload "command args"] [--iterations N] [--focus cpu|mem|io|all]
---

# BENCH-COMPARE — Head-to-head program comparison

Run two programs under identical conditions. Measure everything. Produce a winner with evidence.

**Input:** $ARGUMENTS

---

## METHODOLOGY

The comparison is only valid if both programs run:
- On the same machine
- Under the same load conditions
- With the same input/workload
- In the same session (no reboot between)
- With the system quiesced (no background jobs skewing results)

### Step 1: Define the workload

If `--workload` is provided, use it. Otherwise, determine a fair workload:
- If both programs do the same thing (two web servers, two compilers): use the same input
- If they do different things: this is the wrong tool. Use `/bench-profile` on each separately.

Document the workload explicitly:
```
WORKLOAD: <exact command or input description>
FAIRNESS CHECK: both programs receive identical input? YES/NO
```

### Step 2: Quiesce the system

Before benchmarking:
```bash
# Check for heavy background processes
top -l 1 -n 5 | head -15
# Check memory pressure
vm_stat
# Check disk activity
iostat -c 3
```

If the system is not idle, note it. If it's heavily loaded, wait or warn.

### Step 3: Interleaved execution

Do NOT run all of A then all of B. Interleave to cancel out system drift:

```
Run order: A B B A A B  (3 iterations each, interleaved)
```

For each run, capture via `/usr/bin/time -l`:
- Wall clock
- User CPU
- System CPU  
- Peak RSS
- Page faults
- Involuntary context switches

### Step 4: Statistical comparison

For each metric, compute:
- Median of A runs
- Median of B runs
- Ratio: A/B (>1 means A is worse, <1 means A is better)
- Difference: |A - B|
- Significance: if the ranges overlap, the difference is NOT significant

```markdown
## Comparison Report: <A> vs <B>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM>
**Workload:** <description>
**Iterations:** N per program (interleaved)

### Head-to-Head

| Metric | A (median) | B (median) | Ratio A/B | Winner | Significant? |
|--------|-----------|-----------|-----------|--------|-------------|
| Wall clock | X.XXs | X.XXs | X.XX | A/B | YES/NO |
| User CPU | X.XXs | X.XXs | X.XX | A/B | YES/NO |
| Sys CPU | X.XXs | X.XXs | X.XX | A/B | YES/NO |
| Peak RSS | XX MB | XX MB | X.XX | A/B | YES/NO |
| Page faults | XX | XX | X.XX | A/B | YES/NO |
| Context switches | XX | XX | X.XX | A/B | YES/NO |

### Verdict

**Overall winner:** A / B / DRAW
**By what margin:** XX% faster, XX% less memory, etc.
**Confidence:** HIGH (no overlap in ranges) / LOW (ranges overlap)

### Trade-offs
[Does A win on speed but lose on memory? Document the trade-off.]

### All Runs (raw data)

| Run | Program | Wall | User | Sys | RSS | Faults |
|-----|---------|------|------|-----|-----|--------|
| 1 | A | ... | ... | ... | ... | ... |
| 2 | B | ... | ... | ... | ... | ... |
| 3 | B | ... | ... | ... | ... | ... |
| 4 | A | ... | ... | ... | ... | ... |
| 5 | A | ... | ... | ... | ... | ... |
| 6 | B | ... | ... | ... | ... | ... |
```

### Step 5: Deeper analysis (if requested or if results are close)

If the difference is < 10% on the primary metric:
- Increase iterations to 10
- Profile both with CPU sampling to find WHERE time is spent differently
- Check memory allocation patterns (one might GC more)
- Check I/O patterns (one might buffer differently)

---

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--workload "..."` | inferred | The command/input both programs process |
| `--iterations N` | 3 | Runs per program (interleaved, so 2N total runs) |
| `--focus F` | all | Focus on: cpu, mem, io, speed, or all |
| `--deep` | false | Run deeper analysis (10 iterations + CPU profiling) |

---

## RULES

1. **Same machine, same session.** Never compare runs from different machines or different days.
2. **Interleaved execution.** A-B-B-A-A-B, not A-A-A-B-B-B. System state drifts.
3. **Median, not mean.** Outliers happen. Median is robust.
4. **Significance check.** If A's slowest run is faster than B's fastest run, the difference is real. If ranges overlap, it might be noise.
5. **Raw data preserved.** Every individual run is in the report. The reader can verify.
6. **Trade-offs documented.** "A is 2x faster but uses 3x memory" is not "A wins." It's a trade-off. Report it as one.
