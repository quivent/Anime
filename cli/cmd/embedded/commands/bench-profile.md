---
description: Profile a single program's resource consumption — CPU, memory, disk, network, timing — with hard numbers
argument-hint: <command-or-binary> [--iterations N] [--duration Ns] [--focus cpu|mem|io|all]
---

# BENCH-PROFILE — Measure a program's resource consumption

Profile one program. Produce hard numbers. No estimates, no "should be around X." Measured values only.

**Input:** $ARGUMENTS

---

## WHAT YOU MEASURE

Run the target program and capture ALL of the following. Every metric is a number from an actual execution, not a prediction.

### Timing
```bash
# Wall clock, user CPU, system CPU
/usr/bin/time -l <command>          # macOS
/usr/bin/time -v <command>          # Linux
```
Capture: wall clock (s), user time (s), sys time (s), CPU utilization (%)

### Memory
- Peak RSS (resident set size) — maximum physical memory used
- Virtual memory size
- Memory growth pattern: sample RSS at 1s intervals during execution
- Heap allocations if available (leaks, malloc history)

### CPU
- User vs system time ratio
- Which cores used (if measurable)
- CPU profile if available: `sample <pid>` (macOS) or `perf record` (Linux)
- Hot functions (top 10 by CPU time)

### Disk I/O
- Bytes read / written
- File descriptors opened
- Temporary files created and sizes

### Network (if applicable)
- Connections opened
- Bytes sent / received
- DNS lookups

### Process
- Exit code
- Threads spawned (peak count)
- Child processes spawned

---

## EXECUTION PROTOCOL

### Step 1: Baseline system state
Before running the target, capture system state:
```bash
vm_stat                    # macOS memory pressure
sysctl hw.memsize          # total RAM
sysctl hw.ncpu             # CPU count
df -h                      # disk space
```

### Step 2: Run with instrumentation
Run the target program with timing and resource measurement. Run it `--iterations` times (default: 3). Discard the first run (warmup). Report median of remaining runs.

```bash
# macOS
/usr/bin/time -l <command> 2>&1

# If more detail needed:
dtrace -n 'syscall:::entry /pid == $target/ { @[probefunc] = count(); }' -c "<command>"

# For memory over time:
while kill -0 $PID 2>/dev/null; do ps -o rss= -p $PID; sleep 1; done
```

### Step 3: Produce the profile report

```markdown
## Profile Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM> / <Disk>
**Iterations:** N (median of N-1 after warmup discard)

### Timing
| Metric | Value |
|--------|-------|
| Wall clock | X.XXs |
| User CPU | X.XXs |
| System CPU | X.XXs |
| CPU utilization | XX% |

### Memory
| Metric | Value |
|--------|-------|
| Peak RSS | XX MB |
| Virtual size | XX MB |
| Page faults | XX |
| Memory growth | [flat / linear / spike at Xs] |

### CPU
| Metric | Value |
|--------|-------|
| User/sys ratio | X.X |
| Top function | name (XX% of CPU time) |

### Disk I/O
| Metric | Value |
|--------|-------|
| Bytes read | XX MB |
| Bytes written | XX MB |
| Files opened | XX |

### Process
| Metric | Value |
|--------|-------|
| Exit code | X |
| Peak threads | X |
| Child processes | X |

### Raw Output
<paste the actual /usr/bin/time -l output here>
```

### Step 4: Identify anomalies
Flag anything unusual:
- Peak RSS > 10x input size
- CPU utilization < 25% (I/O bound) or > 200% (parallel)
- Wall clock >> user+sys time (blocking on I/O or locks)
- Memory growth is not flat (possible leak)

---

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--iterations N` | 3 | Number of runs (first discarded as warmup) |
| `--duration Ns` | none | For long-running programs: measure for N seconds then kill |
| `--focus F` | all | Focus on: cpu, mem, io, or all |
| `--compare FILE` | none | Compare against a previous profile report |

---

## RULES

1. **Every number comes from execution.** No estimates. No "approximately." Run it and report what happened.
2. **Multiple iterations.** One run is noise. Three runs with median is signal.
3. **Raw output preserved.** The actual tool output is pasted in the report, not summarized.
4. **System state documented.** The profile is meaningless without knowing the machine it ran on.
