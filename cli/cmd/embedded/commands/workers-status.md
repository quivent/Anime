Show the current status of all autonomous worker systems. Read the state files and report.

## Execute These Commands

1. Show waveworkers daemon status:
```bash
~/waveworkers/status.sh
```

2. Show active workers:
```bash
cat ~/waveworkers/workers-state.json
```

3. Show task progress:
```bash
echo "Waveworkers:" && grep -c '^\- \[x\]' ~/waveworkers/PROGRESS.md && echo "done" && grep -c '^\- \[ \]' ~/waveworkers/PROGRESS.md && echo "open"
```

4. Show optimization status:
```bash
~/optimization/status.sh 2>/dev/null || echo "Optimization daemon not running"
echo "Optimization:" && grep -c '^\- \[x\]' ~/optimization/PROGRESS.md 2>/dev/null && echo "done" && grep -c '^\- \[ \]' ~/optimization/PROGRESS.md 2>/dev/null && echo "open"
```

5. Show integration status summary:
```bash
echo "=== Integration ===" && grep -c 'SHIPPED' ~/waveworkers/INTEGRATED.md && echo "shipped" && grep -c 'TESTED' ~/waveworkers/INTEGRATED.md && echo "tested" && grep -c 'INTEGRATED' ~/waveworkers/INTEGRATED.md && echo "integrated" && grep -c 'DESIGNED' ~/waveworkers/INTEGRATED.md && echo "designed"
```

6. Check if dashboard is running:
```bash
curl -s http://localhost:8767/api/status > /dev/null 2>&1 && echo "Dashboard: http://localhost:8767" || echo "Dashboard: not running (~/waveworkers/dashboard/start-dashboard.sh)"
```

## Report Format

Present a clean summary table:

| System | Status | Progress | Workers |
|--------|--------|----------|---------|
| waveworkers | tier/running | done/total | active count |
| optimization | tier/running | done/total | — |
| dashboard | running/stopped | — | — |
| tunnel | running/stopped | — | — |

Then list active workers by name and task.
Then list next 5 unchecked tasks from PROGRESS.md.

## Load Balancing

If active workers < 3 and unchecked tasks > 0, suggest deploying more workers. If active workers > 8, suggest waiting for some to complete. The target is 3-5 concurrent workers for optimal throughput without context thrashing.
