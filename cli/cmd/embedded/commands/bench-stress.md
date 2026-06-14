---
description: Stress test — push a program to its breaking point and document exactly where it fails
argument-hint: <command-or-service> [--mode connections|memory|throughput|all] [--ceiling 10000]
---

# BENCH-STRESS — Find the breaking point

Push a program until it breaks. Document exactly where, why, and how it fails. The deliverable is a failure map.

**Input:** $ARGUMENTS

## STRESS MODES

### connections — concurrent load
Ramp concurrent requests/connections from 1 to `--ceiling`:
```
1 → 5 → 10 → 25 → 50 → 100 → 250 → 500 → 1000 → 2500 → 5000 → 10000
```
At each level: measure response time (p50, p95, p99), error rate, throughput. Stop when error rate > 50% or program crashes.

### memory — allocation pressure
Feed progressively larger inputs or trigger memory-intensive operations. Monitor RSS at 1s intervals. Stop at OOM kill or system swap thrashing.

### throughput — sustained high load
Maintain maximum request rate for increasing durations: 10s, 30s, 60s, 5m, 15m. Watch for throughput decay, increasing latency, or resource exhaustion.

### all — run all three sequentially

## EXECUTION

### Step 1: Baseline
Run at minimal load. Record: response time, throughput, RSS, CPU. This is the "healthy" reference.

### Step 2: Ramp
Increase load in the selected mode. At each level:
```bash
# For connection stress (example with curl/wrk/hey):
hey -n 1000 -c $LEVEL http://localhost:PORT/endpoint

# For memory stress:
<command> < <generated-input-at-size-N>

# For throughput stress:
hey -z ${DURATION}s -c 50 http://localhost:PORT/endpoint
```

Record at each level:
- Response time: p50, p95, p99
- Error rate (%)
- Throughput (req/s or items/s)
- Peak RSS
- CPU utilization

### Step 3: Identify the breaking point

The breaking point is where ONE of:
- Error rate exceeds 5% (degraded) or 50% (broken)
- p99 latency exceeds 10x baseline
- Program crashes or is OOM killed
- Throughput drops below 50% of peak

### Step 4: Summary Report

```markdown
## Stress Report: <command>

**Date:** <timestamp>
**System:** <OS> / <CPU> / <RAM>
**Mode:** connections / memory / throughput / all

### Baseline (healthy)
| Metric | Value |
|--------|-------|
| Response time (p50) | Xms |
| Throughput | X req/s |
| Peak RSS | X MB |
| CPU | X% |

### Stress Ramp
| Level | p50 | p95 | p99 | Errors | Throughput | RSS | CPU |
|-------|-----|-----|-----|--------|------------|-----|-----|
| 1     | ... | ... | ... | 0%     | ...        | ... | ... |
| 10    | ... | ... | ... | 0%     | ...        | ... | ... |
| 100   | ... | ... | ... | 2%     | ...        | ... | ... |
| 500   | ... | ... | ... | 15%    | ...        | ... | ... |
| 1000  | ... | ... | ... | CRASH  | —          | OOM | — |

### Breaking Point
- **Degradation begins:** level N (error rate > 5%)
- **Failure point:** level M (crash / OOM / error > 50%)
- **Failure mode:** [OOM / timeout / connection refused / crash]
- **Last stable level:** N-1

### Capacity Verdict
- **Safe operating capacity:** up to N (with headroom)
- **Maximum burst capacity:** M (will degrade)
- **Will fail at:** M+1

### Failure Analysis
[What resource was exhausted? Memory? File descriptors? CPU? Thread pool?]
[Is the failure graceful (error response) or catastrophic (crash)?]

### Raw Data
[all outputs from each level]
```
