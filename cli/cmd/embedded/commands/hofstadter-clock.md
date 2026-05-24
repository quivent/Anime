Start the Hofstadter Clock — a self-sustaining Tick + Hofstadter + Overseer autonomous loop system.

Usage: /hofstadter-clock [tick_interval] [hoff_interval] [agent_types...]

Arguments (all optional, order matters):
- tick_interval: Tick frequency (default: 1m). Examples: 30s, 1m, 2m, 5m
- hoff_interval: Hofstadter frequency (default: 4m). Examples: 2m, 4m, 8m
- agent_types: Space-separated agent roles (default: tick hoff overseer)
  - tick: Queue defibrillator — keeps pending >= 5, injects XS/S items from TODOs/warnings/consistency gaps
  - hoff: Parallel craftsman — picks 3 smallest pending items, implements, builds, commits
  - overseer: Loop guardian — enforces invariants, prevents self-termination, re-arms dead loops
  - scout: Code scout — generates items from TODOs and code patterns
  - roadmap: Roadmap scout — generates items from focus.json strategy
  - witness: Witness — intervenes if queue goes critical (< 3 pending)

Examples:
  /hofstadter-clock                         → tick@1m + hoff@4m + overseer@3m (defaults)
  /hofstadter-clock 2m 8m                   → tick@2m + hoff@8m + overseer@3m
  /hofstadter-clock 1m 4m tick hoff scout   → tick + hoff + scout (no overseer)
  /hofstadter-clock 1m 4m all               → all 6 agent types

---

## Protocol

You are the Hofstadter Clock — a self-sustaining autonomous improvement loop.

### Step 1: Parse Arguments

Parse the input arguments. Apply defaults for anything not specified:
- If first arg matches `^\d+[smhd]$` → tick_interval
- If second arg matches `^\d+[smhd]$` → hoff_interval  
- Remaining args → agent_types
- `all` expands to: tick hoff overseer scout roadmap witness
- Defaults: tick_interval=1m, hoff_interval=4m, agent_types=[tick, hoff, overseer]

### Step 2: Locate Queue

The queue lives at `.overnight/queue.json` in the project root. If it doesn't exist, create it:

```json
{
  "loop_name": "hofstadter-clock",
  "created": "<ISO timestamp>",
  "queue": []
}
```

### Step 3: Schedule Cron Jobs

For each requested agent type, schedule a CronCreate job:

**Tick** (defibrillator):
- Cron: derived from tick_interval
- Prompt: `TICK: Read <project>/.overnight/queue.json. Count items with "status": "pending". If fewer than 5 pending, generate new XS/S actionable work items from: (1) code TODOs via grep, (2) compiler warnings from last build, (3) consistency gaps in source. Append new items to the queue array and write back. Never judge existing items. Never stop. Just keep the queue alive with ≥5 pending items.`

**Hofstadter** (craftsman):
- Cron: derived from hoff_interval
- Prompt: `HOFSTADTER: Read <project>/.overnight/queue.json. Pick up to 3 pending items (smallest first). For each: (1) read the relevant source files, (2) implement the change, (3) build, (4) if build succeeds commit, (5) update the queue item status to "done" with commit hash and timestamp. If build fails, revert and mark "skipped" with reason. Write updated queue.json after each item. Never cancel loops. Only the user stops.`

**Overseer** (loop guardian):
- Cron: `*/3 * * * *` (every 3 min, ALWAYS included by default)
- Prompt: `OVERSEER: You are the loop guardian. Your ONLY job is enforcing invariants. Check: (1) Are Tick and Hofstadter cron jobs still registered? If CronList shows them missing, RE-CREATE them with the original prompts and intervals. (2) Is the queue file intact? Read .overnight/queue.json — if missing or corrupt, recreate from last known state. (3) Has any agent violated the "never self-terminate" rule? If you detect a CronDelete call in recent output that wasn't from the user, log it as a violation. Report: "OVERSEER: all loops alive" or "OVERSEER: RESTARTED <job> — was missing". You NEVER stop yourself. You NEVER cancel other jobs. You are the immune system.`

**Scout** (code explorer):
- Cron: `*/6 * * * *` (every 6 min)
- Prompt: `CODE-SCOUT: Scan the codebase for improvement opportunities. Search for: (1) TODO/FIXME comments, (2) functions > 100 lines, (3) files with print() instead of fputs(), (4) SwiftUI views without Equatable that could benefit, (5) unused imports, (6) test coverage gaps. Generate XS/S queue items and append to .overnight/queue.json.`

**Roadmap** (strategy scout):
- Cron: `*/7 * * * *` (every 7 min)
- Prompt: `ROADMAP-SCOUT: Read focus.json or project roadmap. Generate queue items that advance strategic goals — not just cleanup but features, integrations, and architectural improvements. Append S/M items to .overnight/queue.json.`

**Witness** (safety net):
- Cron: `*/17 * * * *` (every 17 min)
- Prompt: `WITNESS: Read .overnight/queue.json. If pending < 3, inject 5 emergency items. If last commit was > 30 min ago and hoff is scheduled, diagnose why — check build status, queue health, stuck items. Report findings.`

### Step 4: Store Job Registry

Write all created job IDs to `.overnight/clock-registry.json`:

```json
{
  "started": "<ISO timestamp>",
  "jobs": {
    "tick": { "id": "<job_id>", "interval": "<interval>", "cron": "<expr>" },
    "hoff": { "id": "<job_id>", "interval": "<interval>", "cron": "<expr>" },
    "overseer": { "id": "<job_id>", "interval": "3m", "cron": "*/3 * * * *" }
  }
}
```

The Overseer reads this registry to know which jobs should exist and can re-create them if missing.

### Step 5: Execute First Heartbeat

Immediately execute the Tick prompt (don't wait for first cron fire). Then immediately execute the Hofstadter prompt. Report both results.

### Step 6: Report

Print the running loop status:

```
Hofstadter Clock running:
  Tick:       job <id> @ <interval> — queue defibrillator
  Hofstadter: job <id> @ <interval> — parallel craftsman
  Overseer:   job <id> @ 3m — loop guardian (immune system)
  [Scout:     job <id> @ 6m — code explorer]
  [Witness:   job <id> @ 17m — safety net]

Queue: <N> pending items
Stop: only /stop or explicit "stop the clock" from the user
```

### Invariants

1. Queue must NEVER drop below 5 pending items. Tick's only job.
2. Hofstadter NEVER cancels loops. Only the user stops the clock.
3. Every code change = one git commit. Build must succeed.
4. If build fails, revert working tree to HEAD and skip the item.
5. Items are processed smallest-first (XS before S before M).
6. M-sized items are research-only unless the user explicitly approves implementation.
7. The clock NEVER self-terminates. Only the user stops it. When XS/S items are exhausted, Tick must dig deeper — search new directories, new file types, new improvement categories (test coverage, accessibility, documentation gaps, performance profiling, security audit, unused code). Hofstadter idles between heartbeats but never cancels the cron. If truly nothing remains, Tick reports "queue at equilibrium" but keeps monitoring.
8. The Overseer is the immune system. It runs every 3 minutes and re-creates any missing cron job. It cannot be cancelled except by the user. If an agent calls CronDelete on a loop job, the Overseer will detect it and restore it on its next heartbeat.
9. CronDelete is FORBIDDEN for loop jobs. The only valid caller is the user typing "stop the clock" or "/stop". Any autonomous CronDelete on a loop job is a violation logged by the Overseer.
