---
description: Regression benchmark — compare current build against a previous build or git ref to catch performance regressions
argument-hint: <command> --baseline <git-ref|profile-file> [--threshold 10%] [--focus time|mem|all]
---

# BENCH-REGRESS — Catch performance regressions in your own code

Compare the current version of your program against a previous version. Find regressions before they ship.

**Input:** $ARGUMENTS

## EXECUTION

### Step 1: Establish baseline

If `--baseline` is a git ref (commit hash, tag, branch):
```bash
# Build the baseline version
git stash  # save current state
git checkout <ref>
<build command>
# Run bench-profile on baseline
/bench-profile <command> --iterations 5
# Save results
git checkout -  # return to current
git stash pop
<build command>  # rebuild current
```

If `--baseline` is a profile file: read the saved profile data.

### Step 2: Profile current version

Run `/bench-profile` on the current build with the same workload and iterations as baseline.

### Step 3: Compare

For each metric, compute:
- **Delta:** current - baseline
- **Delta %:** (current - baseline) / baseline × 100
- **Verdict:** FASTER / SLOWER / SAME (within noise threshold)

A metric is a **regression** if it's slower by more than `--threshold` (default: 10%).
A metric is an **improvement** if it's faster by more than the threshold.
Within threshold = SAME.

### Step 4: Summary Report

```markdown
## Regression Report: <command>

**Date:** <timestamp>
**Current:** <git ref or "working tree">
**Baseline:** <git ref or profile file>
**Threshold:** ±X%

### Regression Check
| Metric | Baseline | Current | Delta | Delta % | Verdict |
|--------|----------|---------|-------|---------|---------|
| Wall clock | 1.23s | 1.45s | +0.22s | +17.9% | REGRESSION |
| User CPU | 1.10s | 1.30s | +0.20s | +18.2% | REGRESSION |
| Sys CPU | 0.13s | 0.15s | +0.02s | +15.4% | REGRESSION |
| Peak RSS | 45 MB | 44 MB | -1 MB | -2.2% | SAME |
| Page faults | 1200 | 1180 | -20 | -1.7% | SAME |

### Overall Verdict
**REGRESSION DETECTED** / **NO REGRESSION** / **IMPROVED**

### Regressions Found
1. Wall clock +17.9% — exceeds 10% threshold
2. User CPU +18.2% — exceeds 10% threshold

### Likely Cause
[If git ref was used: diff the two refs and identify what changed]
```bash
git diff <baseline>..<current> --stat
```
[List files changed and correlate with regression]

### Improvements Found
[Any metrics that improved beyond threshold]

### Recommendation
- **BLOCK:** regression exceeds 2x threshold — investigate before merging
- **WARN:** regression exceeds threshold — review and justify
- **PASS:** no regressions detected

### Raw Data
[baseline profile output]
[current profile output]
```

## FLAGS

| Flag | Default | Description |
|------|---------|-------------|
| `--baseline REF` | required | Git ref or saved profile file to compare against |
| `--threshold N%` | 10% | Regression threshold — deltas within this are SAME |
| `--focus F` | all | Focus: time, mem, or all |
| `--block-on-regress` | false | Exit with error code if regression found (for CI) |
