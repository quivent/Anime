---
description: Scaling benchmark — how does performance change as input grows 1x, 10x, 100x, 1000x
argument-hint: <command> --input-gen "generator" [--steps 1,10,100,1000] [--focus time|mem|all]
---

# BENCH-SCALE — Scaling behavior under growing input

Measure how a program behaves as its workload increases by orders of magnitude. Answers: is it O(n)? O(n²)? Does it fall off a cliff?

**Input:** $ARGUMENTS

## EXECUTION

### Step 1: Define the input scale ladder

Either use `--steps` (default: 1,10,100,1000) or let the command determine sensible sizes based on the program.

For each step, the `--input-gen` command produces input at that scale:
```bash
# Example: scaling a JSON parser
--input-gen "python3 -c \"import json; print(json.dumps({'x':1}*SIZE))\""
# SIZE is substituted with each step value
```

### Step 2: Run at each scale point

For each scale step, run `/bench-profile` internally (3 iterations, median):
- Wall clock, user CPU, sys CPU
- Peak RSS
- Disk I/O

### Step 3: Compute scaling characteristics

```
| Scale | Wall (s) | RSS (MB) | Wall Ratio | RSS Ratio | Wall/n |
|-------|----------|----------|------------|-----------|--------|
| 1x    | 0.05     | 12       | —          | —         | 0.050  |
| 10x   | 0.48     | 15       | 9.6x       | 1.25x     | 0.048  |
| 100x  | 4.9      | 45       | 10.2x      | 3.0x      | 0.049  |
| 1000x | 502      | 3200     | 102.4x     | 71.1x     | 0.502  |
```

**Wall Ratio**: time at this step / time at previous step. Linear = matches scale factor. Superlinear = exceeds it.
**Wall/n**: time normalized by input size. Constant = O(n). Growing = worse than linear.

### Step 4: Classify scaling behavior

For each metric:
- **Linear O(n)**: ratio tracks scale factor, wall/n is constant
- **Linearithmic O(n log n)**: ratio slightly exceeds scale factor
- **Quadratic O(n²)**: ratio is ~square of scale factor
- **Cliff**: works fine up to N, then OOM/timeout/crash

### Step 5: Summary Report

```markdown
## Scaling Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM>
**Scale steps:** [list]

### Scaling Table
[table from Step 3]

### Scaling Classification
- **Time complexity:** O(n) / O(n log n) / O(n²) / cliff at Nx
- **Memory complexity:** O(1) / O(n) / O(n²) / cliff at Nx

### Scaling Chart (ASCII)
[ASCII plot of wall clock vs input size]

### Anomalies
- [Any cliff points, discontinuities, or unexpected behavior]

### Practical Limits
- **Comfortable operating range:** up to Nx input
- **Degraded but functional:** Nx to Mx
- **Will fail:** above Mx (OOM / timeout / crash)

### Raw Data
[all individual run outputs]
```
