---
description: Sustained load benchmark — run for extended duration to find memory leaks, GC pauses, and throughput decay
argument-hint: <command-or-service> [--duration 30m] [--sample-interval 5s] [--workload "command"]
---

# BENCH-SUSTAINED — Long-running stability test

Run a program under steady load for an extended period. Catch problems that only appear over time: memory leaks, GC pause accumulation, throughput decay, file descriptor exhaustion, log file growth.

**Input:** $ARGUMENTS

## EXECUTION

### Step 1: Establish steady-state workload

Define a workload that represents normal operation:
```bash
# For a server: steady request rate
while true; do curl -s http://localhost:PORT/endpoint > /dev/null; sleep 0.1; done

# For a CLI tool: repeated invocation
while true; do <command> < input.txt > /dev/null; done

# For a long-running process: just let it run
<command> &
```

### Step 2: Sample at regular intervals

Every `--sample-interval` (default: 5s) for `--duration` (default: 30m), capture:

```bash
# Memory
ps -o rss=,vsz= -p $PID

# Open file descriptors
lsof -p $PID 2>/dev/null | wc -l

# CPU
ps -o %cpu= -p $PID

# For servers: response time + throughput snapshot
curl -w "%{time_total}" -s http://localhost:PORT/endpoint > /dev/null

# Disk usage of working directory / logs
du -sh <relevant-dirs>
```

Record all samples in a time series.

### Step 3: Analyze trends

For each metric across the time series:
- **Stable:** flat line (within 5% of initial value)
- **Growing:** monotonically increasing (leak candidate)
- **Decaying:** monotonically decreasing (throughput decay)
- **Spiky:** periodic jumps (GC pauses, batch processing)
- **Cliff:** sudden change after N minutes (delayed failure)

### Step 4: Summary Report

```markdown
## Sustained Load Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM>
**Duration:** Xm
**Sample interval:** Xs
**Total samples:** N

### Trend Summary
| Metric | Start | End | Trend | Verdict |
|--------|-------|-----|-------|---------|
| RSS | 45 MB | 48 MB | +6.7% | STABLE |
| Virtual | 120 MB | 310 MB | +158% | GROWING — possible leak |
| Open FDs | 24 | 24 | 0% | STABLE |
| CPU | 12% | 14% | +16.7% | STABLE |
| Response time | 5ms | 8ms | +60% | GROWING — latency creep |
| Throughput | 200/s | 180/s | -10% | DECAYING |
| Log size | 1 MB | 45 MB | +4400% | GROWING — unbounded logs |

### Time Series (ASCII)
```
RSS (MB):
  50 |                              ___---
  45 |____---___---___---___---___--
  40 |
     +----+----+----+----+----+----
     0    5m   10m  15m  20m  25m  30m
```

### Issues Found

#### Memory Leak (if detected)
- RSS grew from X to Y over Z minutes
- Growth rate: ~N MB/minute
- Projected OOM in: ~M hours at this rate

#### Throughput Decay (if detected)
- Started at X req/s, ended at Y req/s
- Decay rate: ~N%/minute

#### GC Pauses (if detected)
- N pauses detected (response time spikes > 10x median)
- Average pause duration: Xms
- Longest pause: Xms
- Frequency: every ~N seconds

#### Resource Accumulation (if detected)
- File descriptors: growing / stable
- Log files: growing / stable / rotated
- Temp files: accumulating / cleaned

### Stability Verdict
- **STABLE:** all metrics within 10% of starting values — safe for long-running deployment
- **DEGRADING:** one or more metrics trending wrong — will fail eventually
- **LEAKING:** memory growth detected — will OOM in estimated ~N hours
- **DECAYING:** throughput declining — will become unusable in ~N hours

### Recommendations
[Specific actions based on findings]

### Raw Time Series Data
[all sample data points, CSV format]
```
