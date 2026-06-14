---
description: Cold vs warm start benchmark — measure first-run penalty from cache misses, JIT, lazy init, disk reads
argument-hint: <command> [--iterations 10] [--flush-method cache|reboot|purge]
---

# BENCH-COLD — Cold start vs warm performance

Measure the difference between first invocation (cold: empty caches, no JIT, no mmap'd pages) and subsequent invocations (warm: hot caches, compiled code, resident pages). Answers: how bad is the cold start penalty?

**Input:** $ARGUMENTS

## EXECUTION

### Step 1: Cold start measurement

Flush caches before each cold run:
```bash
# macOS — purge disk cache
sudo purge

# Linux — drop page cache
sync; echo 3 | sudo tee /proc/sys/vm/drop_caches

# Alternative if no root: use a different binary path each time (defeats filesystem cache)
cp <binary> /tmp/cold_run_$RANDOM && /usr/bin/time -l /tmp/cold_run_$RANDOM
```

Run `--iterations` cold starts (default: 5). Each preceded by cache flush. Record:
- Wall clock
- User CPU
- Sys CPU
- Peak RSS
- Page faults (CRITICAL — this is the cold start signal)
- Involuntary context switches

### Step 2: Warm start measurement

Run the program `--iterations` times WITHOUT flushing between runs. Discard first run. Record same metrics.

### Step 3: Transition measurement

Run 10 consecutive invocations and track metrics for EACH individually. This shows the warmup curve — how many runs until performance stabilizes.

```
Run 1 (cold): 2.3s — 15000 page faults
Run 2: 0.8s — 200 page faults
Run 3: 0.5s — 50 page faults
Run 4: 0.45s — 12 page faults
Run 5: 0.44s — 10 page faults  ← stabilized
```

### Step 4: Summary Report

```markdown
## Cold Start Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM>
**Cache flush method:** purge / drop_caches / copy

### Cold vs Warm
| Metric | Cold (median) | Warm (median) | Penalty | Ratio |
|--------|--------------|--------------|---------|-------|
| Wall clock | 2.30s | 0.44s | +1.86s | 5.2x |
| User CPU | 0.80s | 0.40s | +0.40s | 2.0x |
| Sys CPU | 1.20s | 0.03s | +1.17s | 40x |
| Peak RSS | 120 MB | 45 MB | +75 MB | 2.7x |
| Page faults | 15000 | 10 | +14990 | 1500x |
| Context sw. | 450 | 12 | +438 | 37.5x |

### Cold Start Penalty
- **Wall clock penalty:** +X.XXs (X.Xx slower)
- **Dominated by:** page faults / JIT / lazy init / disk reads
- **Sys CPU ratio:** Xx (high = kernel time loading pages from disk)
- **Page fault ratio:** Xx (high = lots of cold memory access)

### Warmup Curve
| Run | Wall (s) | Faults | Status |
|-----|----------|--------|--------|
| 1 | 2.30 | 15000 | COLD |
| 2 | 0.80 | 200 | WARMING |
| 3 | 0.50 | 50 | WARMING |
| 4 | 0.45 | 12 | WARM |
| 5 | 0.44 | 10 | WARM |

**Runs to warm:** 4 invocations until stable performance

### Where Cold Time Goes
- **Page faults × page size:** ~X MB loaded from disk
- **Sys CPU dominance:** kernel spent Xs loading pages vs Xs in warm
- **If JIT (Java, Node, etc.):** first-run compilation overhead

### Practical Impact
- **Single invocation use (CLI tools):** cold start IS the performance
- **Server/daemon:** cold start matters once at boot, then irrelevant
- **Lambda/serverless:** cold start is critical — every new instance pays it
- **Recommendation:** [prewarming strategy if applicable]

### Raw Data
[all individual run outputs for cold, warm, and transition]
```

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--iterations N` | 5 | Runs per category (cold and warm) |
| `--flush-method M` | purge (macOS) / drop_caches (Linux) | How to flush caches before cold runs |
| `--transition N` | 10 | Consecutive runs for warmup curve |
| `--no-sudo` | false | Skip cache flush (uses binary-copy trick instead) |
