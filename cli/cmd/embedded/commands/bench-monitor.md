---
description: Live resource monitor — record CPU, GPU, RAM, disk, network over time with time-series output and charts
argument-hint: <command-or-pid> [--duration 60s] [--interval 1s] [--gpu] [--output report.md]
---

# BENCH-MONITOR — Continuous resource recording over time

Attach to a running process or launch a command and record CPU, GPU, RAM, disk I/O, and network at regular intervals. Produces a time-series dataset and summary report with ASCII charts.

**Input:** $ARGUMENTS

## WHAT IT RECORDS

Every `--interval` (default: 1s), capture:

| Resource | Metric | How |
|----------|--------|-----|
| **CPU** | % utilization, user/sys split | `ps -o %cpu=,time= -p $PID` |
| **RAM** | RSS (MB), virtual (MB) | `ps -o rss=,vsz= -p $PID` |
| **GPU** | utilization %, VRAM used | `sudo powermetrics --samplers gpu_power -i 1000 -n 1` (macOS) or `nvidia-smi --query-gpu=utilization.gpu,memory.used --format=csv,noheader` |
| **Disk** | bytes read, bytes written | `iotop` or `/proc/$PID/io` (Linux) / `fs_usage -w -f filesys -p $PID` (macOS) |
| **Network** | bytes sent, bytes received | `nettop -p $PID -J bytes_in,bytes_out -l 1` (macOS) or `/proc/$PID/net/dev` |
| **Threads** | count | `ps -M -p $PID | wc -l` |
| **FDs** | open file descriptors | `lsof -p $PID 2>/dev/null | wc -l` |

## EXECUTION

### Step 1: Attach or launch

If input is a PID: attach to existing process.
If input is a command: launch it, capture PID.

```bash
# Launch and capture
$COMMAND &
PID=$!

# Or attach to existing
PID=<given>
```

### Step 2: Sample loop

```bash
START=$(date +%s)
echo "time,cpu,rss_mb,vsz_mb,gpu_pct,gpu_vram_mb,disk_r_mb,disk_w_mb,net_in_mb,net_out_mb,threads,fds" > samples.csv

while kill -0 $PID 2>/dev/null; do
    ELAPSED=$(($(date +%s) - START))
    [ $ELAPSED -ge $DURATION ] && break

    CPU=$(ps -o %cpu= -p $PID | tr -d ' ')
    RSS=$(ps -o rss= -p $PID | awk '{printf "%.1f", $1/1024}')
    VSZ=$(ps -o vsz= -p $PID | awk '{printf "%.1f", $1/1024}')
    # GPU, disk, net sampling per platform
    
    echo "$ELAPSED,$CPU,$RSS,$VSZ,..." >> samples.csv
    sleep $INTERVAL
done
```

### Step 3: Summary Report

```markdown
## Resource Monitor Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM> / <GPU>
**Duration:** Xs actual (Xs requested)
**Samples:** N at Xs intervals
**Process:** <command> (PID: XXXX)

### Resource Summary
| Resource | Min | Max | Mean | Final | Trend |
|----------|-----|-----|------|-------|-------|
| CPU % | X | X | X | X | stable/growing/spiky |
| RSS (MB) | X | X | X | X | stable/growing/leak |
| Virtual (MB) | X | X | X | X | stable/growing |
| GPU % | X | X | X | X | stable/burst/sustained |
| GPU VRAM (MB) | X | X | X | X | stable/growing |
| Disk read (MB) | X | X | X | X | burst/sustained |
| Disk write (MB) | X | X | X | X | burst/sustained |
| Net in (MB) | X | X | X | X | burst/sustained |
| Net out (MB) | X | X | X | X | burst/sustained |
| Threads | X | X | X | X | stable/growing |
| Open FDs | X | X | X | X | stable/growing |

### Time Series Charts (ASCII)

```
CPU %:
 100|    *
  75|   * **    *
  50| **    ****  ***
  25|*              ****
   0+--+--+--+--+--+--+--
    0  10 20 30 40 50 60s

RSS (MB):
 200|                 ***
 150|           *****
 100|     ******
  50|*****
   0+--+--+--+--+--+--+--
    0  10 20 30 40 50 60s

GPU %:
 100|
  75|  ***   ***
  50| *   * *   *
  25|*     *     ********
   0+--+--+--+--+--+--+--
    0  10 20 30 40 50 60s
```

### Peak Events
- **CPU peak:** X% at T=Xs [what was happening]
- **RAM peak:** X MB at T=Xs
- **GPU peak:** X% at T=Xs

### Anomalies
- [Memory leak: RSS grew Xmb/min without plateau]
- [CPU spike: 100% for Xs at T=Xs]
- [GPU idle: 0% utilization despite GPU-capable workload]
- [FD leak: open descriptors growing without bound]

### Raw CSV
[Full samples.csv data pasted or path to file]
```

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--duration Ns` | 60s | How long to record |
| `--interval Ns` | 1s | Sampling frequency |
| `--gpu` | auto-detect | Include GPU metrics (requires powermetrics or nvidia-smi) |
| `--output FILE` | stdout | Write report to file |
| `--csv FILE` | none | Also save raw CSV to this path |
| `--no-charts` | false | Skip ASCII charts, just produce tables |
