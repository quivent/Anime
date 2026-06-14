---
description: Automated task-driven comparison — run the same resource-heavy workflow on two parity programs and compare
argument-hint: <program-A> <program-B> --task "workflow description" [--record cpu,gpu,ram,disk] [--output report.md]
---

# BENCH-ARENA — Automated workflow comparison for parity programs

Two programs that do the same thing. One automated workflow. Full resource recording on both. Verdict with evidence.

This is NOT just "run both and compare timing" (that's `/bench-compare`). This automates a REAL workflow — opening files, performing operations, exporting results — and records CPU/GPU/RAM/disk the entire time.

**Input:** $ARGUMENTS

## WHEN TO USE

- Comparing two video editors rendering the same project
- Comparing two compilers building the same codebase  
- Comparing two databases running the same query suite
- Comparing two ML frameworks training the same model
- Comparing two browsers loading the same pages
- Any "which tool should we use?" question with resource implications

## EXECUTION

### Step 1: Define the workflow task

The `--task` is a description of what both programs should do. The agent translates this into executable steps for each program.

```
--task "Open a 4K video file, apply color correction, add a title overlay, export as H.264 1080p"
--task "Build the project from clean, run test suite, generate coverage report"  
--task "Import 1M rows from CSV, run 10 analytical queries, export results"
```

The agent must figure out HOW to automate this for each program:
- CLI flags and arguments
- Scripting APIs (AppleScript, Python bindings, etc.)
- Input files that need to be prepared
- Output validation (did both produce correct results?)

### Step 2: Prepare identical inputs

Both programs receive exactly the same input:
- Same source files
- Same configuration (where applicable)
- Same output format requirements

Document what was prepared:
```
INPUT: 4K video file (test-input.mov, 120s, 3.2GB)
CONFIG: H.264 export, 1080p, CRF 23
OUTPUT: exported-A.mp4, exported-B.mp4
```

### Step 3: Run Program A with full monitoring

```bash
# Launch bench-monitor on program A
/bench-monitor "<program-A-workflow-command>" --duration until-complete --interval 1s --gpu --csv arena-A.csv
```

Capture:
- CPU/GPU/RAM/disk over time (via bench-monitor)
- Wall clock for entire workflow
- Output file (for correctness check)

### Step 4: Quiesce system, then run Program B

```bash
# Wait for system to return to idle
sleep 10
# Verify idle
top -l 1 | head -5

# Run program B with same monitoring
/bench-monitor "<program-B-workflow-command>" --duration until-complete --interval 1s --gpu --csv arena-B.csv
```

### Step 5: Validate outputs

Before comparing resources, verify both programs produced CORRECT output:
- Do both outputs exist?
- Are they the expected format/size?
- Are they equivalent? (diff, visual comparison, checksum where applicable)

If one program produced wrong output, the comparison is invalid for that program.

### Step 6: Summary Report

```markdown
## Arena Report: <A> vs <B>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM> / <GPU>
**Task:** <workflow description>
**Input:** <input description and size>

### Output Validation
| Program | Output | Size | Correct? |
|---------|--------|------|----------|
| A | exported-A.mp4 | 245 MB | YES |
| B | exported-B.mp4 | 238 MB | YES |

### Head-to-Head Resource Comparison

| Metric | A | B | Winner | Margin |
|--------|---|---|--------|--------|
| **Wall clock** | 45s | 62s | A | 27% faster |
| **Peak CPU** | 780% | 420% | B | A used 1.9x more CPU |
| **Mean CPU** | 450% | 380% | B | A used 1.2x more CPU |
| **Peak RAM** | 2.1 GB | 3.4 GB | A | B used 1.6x more RAM |
| **Mean RAM** | 1.8 GB | 2.9 GB | A | B used 1.6x more RAM |
| **Peak GPU** | 95% | 45% | — | A is GPU-heavy, B is CPU-heavy |
| **Peak VRAM** | 1.2 GB | 0.3 GB | B | A used 4x more VRAM |
| **Disk read** | 3.2 GB | 3.2 GB | TIE | Same input |
| **Disk write** | 0.5 GB | 0.8 GB | A | B wrote 1.6x more |
| **Total energy** | ~12 Wh | ~15 Wh | A | estimated from CPU+GPU |

### Resource Profiles (ASCII, overlaid)

```
CPU % over time:
 800|A:****
 600|A:    ****        B:
 400|A:        ****  B:****
 200|        B:****B:      ****
   0+--+--+--+--+--+--+--+--+--
    0  5  10 15 20 25 30 35 40s
    A=solid  B=dashed

RAM (GB) over time:
 3.5|              B:*****
 3.0|           B:**     **
 2.5|        B:**         **
 2.0|A:*****B:*              *
 1.5|A:    *****
 1.0|
   0+--+--+--+--+--+--+--+--+--
    0  5  10 15 20 25 30 35 40s
```

### Strategy Comparison
| Aspect | A | B |
|--------|---|---|
| Parallelism | Uses all 10 cores aggressively | Uses 4-6 cores conservatively |
| GPU usage | Offloads decode + encode to GPU | CPU-only pipeline |
| Memory strategy | Low peak, streams data | Loads into RAM, processes in bulk |
| Disk strategy | Minimal temp files | Heavy temp file usage |

### Verdict

**Speed winner:** A (27% faster)
**Efficiency winner:** B (lower CPU, lower VRAM, lower total energy)
**RAM winner:** A (38% less RAM)

**Overall:** [Depends on constraint — if time matters, A wins. If running alongside other work, B's lower resource footprint is better.]

### Trade-off Matrix
| If you care about... | Use |
|---------------------|-----|
| Fastest completion | A |
| Lowest memory footprint | A |
| Lowest CPU impact (background task) | B |
| GPU-free operation | B |
| Lowest total energy | A |

### Raw Data
- A samples: arena-A.csv (N rows)
- B samples: arena-B.csv (N rows)
[CSV data or path to files]
```

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--task "..."` | required | The workflow to automate on both programs |
| `--record METRICS` | cpu,ram,gpu,disk | Which resources to monitor |
| `--output FILE` | stdout | Write report to file |
| `--input FILE` | auto | Shared input file for both programs |
| `--validate` | true | Check output correctness before comparing |
| `--interval Ns` | 1s | Sampling frequency during monitoring |
