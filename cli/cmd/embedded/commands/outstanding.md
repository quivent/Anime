---
description: Find outstanding work — known roadmap items + discover untracked problems
allowed-tools: Bash, Read, Glob, Grep
---

# /outstanding — Outstanding Work Discovery

Find everything that needs doing. Known roadmap items **and** problems nobody has tracked yet.

Read-only. Does not run benchmarks or tests. Does not write to DB. Uses latest stored data + benchmark CSV.

---

## Phase 1: Load Known State

Single batch query:

```bash
CWD=$(pwd)
sqlite3 -separator '	' ~/.claude/db/projects.db << ENDSQL
.headers off

SELECT '=== ROADMAP ===' as tag;
SELECT id, category, title, priority, status, perf, ops, impact, blocked_by
  FROM roadmap WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND status != 'done'
  ORDER BY priority ASC;

SELECT '=== METRICS ===' as tag;
SELECT ms.category, ms.name, ms.field, ms.value, ms.collected_at
  FROM metric_snapshots ms
  WHERE ms.project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND ms.collected_at = (
      SELECT MAX(ms2.collected_at) FROM metric_snapshots ms2
      WHERE ms2.source_id = ms.source_id
    );

SELECT '=== BASELINES ===' as tag;
SELECT b.field, b.value FROM baselines b
  WHERE b.project_id = (SELECT id FROM projects WHERE path = '$CWD');

ENDSQL
```

Parse the tagged output into roadmap items, current metrics, and baseline values.

## Phase 2: Load Latest Benchmark Detail

Query the database for per-benchmark results from the latest run:

```sql
SELECT '=== BENCH_RUN ===' as tag;
SELECT id, timestamp, total, passed, rfail, grfail, runtime_ratio, sixth_wins
  FROM benchmark_runs
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY timestamp DESC LIMIT 1;

SELECT '=== BENCH_RESULTS ===' as tag;
SELECT name, status, compile_sixth_ms, compile_gcc_ms, run_sixth_ms, run_gcc_ms, ratio
  FROM benchmark_results
  WHERE run_id = (SELECT id FROM benchmark_runs
                  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
                  ORDER BY timestamp DESC LIMIT 1)
  ORDER BY name;
```

Add these queries to the Phase 1 batch query (single DB open).

Partition into:
- **passing**: status = PASS, has a ratio
- **failing**: status = RFAIL/CFAIL/GRFAIL/GCFAIL

If `benchmark_runs` is empty, skip Phases 4a-4d and note "No benchmark data — run benchmarks and persist to DB first."

**Fallback**: If the DB has no benchmark_results but a CSV exists on disk (`bench/results/*/results.csv`), read the CSV and note "WARNING: benchmark data not in DB — run /status to persist."

## Phase 3: Known Work

Display open roadmap items grouped by category. For each item, cross-reference against benchmark data:

- If the item's `impact` field mentions specific benchmark names, show their current ratios from the CSV
- If the item describes a pattern (e.g., "BEGIN loops", "recursive calls", "stack depth"), scan passing benchmarks for matches and show the worst ratios as evidence
- Show the item's expected perf improvement alongside actual current ratios

Format:

```
── Known Work (22 items) ──────────────

optimization #1: Variable-depth tracking (reg-depth)
  Expected: fixes 81 crashes
  Evidence: 80 RFAIL benchmarks, 9 depth* benchmarks at 73x-2509x

optimization #3: Inlining small words
  Expected: 2-5x on call-heavy benchmarks
  Evidence: permute 6.6x, linsearch 6.5x, strength2 5.8x, mutrec2 9.0x, mutrec3 8.3x

...
```

## Phase 4: Discovery

Actively look for problems not tracked in the roadmap.

### 4a. Unmapped Failures

For each failing benchmark (RFAIL, CFAIL, GRFAIL):
1. Check if ANY roadmap item mentions it by name or pattern
2. If no roadmap item covers it, it's unmapped

To determine likely cause, read the benchmark's `.fs` source file (first 20 lines) and look for:
- Deep stack operations (PICK 3+, ROLL, deep nesting)
- Array/memory operations (CELLS +, ALLOT, @, !)
- Specific Forth words that are known-broken

Group unmapped failures by likely cause. Report each group.

### 4b. Unmapped Performance Outliers

For each passing benchmark with ratio > 10x:
1. Check if ANY roadmap item mentions it or its pattern
2. If not covered, it's an unmapped performance problem

To diagnose, read the benchmark source (first 30 lines) and identify the dominant pattern:
- Loop type (DO, BEGIN, WHILE, UNTIL)
- Stack depth (how many items manipulated)
- Call pattern (recursive, mutual recursion, leaf functions)
- Arithmetic pattern (multiply, divide, modulo)

### 4c. Regressions Since Baseline

Compare current metric snapshots against baseline values:

| Check | Regression if... |
|-------|-----------------|
| pass count | current < baseline |
| wrong count | current > baseline |
| rfail count | current > baseline |
| runtime_ratio | current > baseline (higher = slower) |
| selfhost | current != PASS |

Report any regressions found. If none, report "None detected."

### 4d. Benchmark Pattern Clustering

Analyze all 240 benchmarks for structural patterns:

1. **Name-based clusters**: Group by name prefix (depth*, loop*, pick*, rot*, swap*, unroll*, ploop*, doloop*, inline*, etc.). For each cluster, report: count, pass/fail split, median ratio, worst ratio.

2. **Ratio tiers**: How many benchmarks fall in each band?
   - < 1x (Sixth faster): count
   - 1-2x: count
   - 2-5x: count
   - 5-10x: count
   - 10-100x: count
   - 100x+: count
   - FAIL: count

3. **Anomalies**: Benchmarks where Sixth is significantly faster than GCC (ratio < 0.5x). These might indicate measurement error or GCC pessimization worth understanding.

### 4e. Code-Level Signals

Quick, targeted scans — not deep exploration:

1. **TODOs/FIXMEs**: `grep -rn 'TODO\|FIXME\|HACK\|XXX' src/ --include='*.fs'` — count and list unique entries
2. **WORK.md drift**: Count items in WORK.md, compare against roadmap count in DB. If WORK.md has items not in DB, list them.
3. **Dead files**: Check if files mentioned in roadmap cleanup items still exist (e.g., opt-vect.fs)

## Phase 5: Output

```
OUTSTANDING WORK
Date: [today]   Branch: [branch] @ [hash]
Data: benchmarks from [CSV timestamp], metrics from [snapshot timestamp]

── Known ([n] roadmap items) ──────────
  [Phase 3 output — items with evidence]

── Discovered ─────────────────────────

  Unmapped Failures ([n]):
    [group]: [benchmarks] — likely cause: [pattern]
    ...

  Unmapped Slowness ([n] benchmarks > 10x, untracked):
    [benchmark] [ratio]x — [diagnosis]
    ...

  Regressions:
    [metric]: [baseline] → [current]
    [Or: "None detected"]

  Clusters:
    [prefix]: [n] total, [pass]/[fail], median [x]x, worst [y]x
    ...

  Ratio Distribution:
    < 1x (Sixth wins): [n]
    1-2x:   [n]
    2-5x:   [n]
    5-10x:  [n]
    10-100x: [n]
    100x+:  [n]
    FAIL:   [n]

  Code Signals:
    [n] TODOs in compiler source
    WORK.md: [SYNC | DRIFT: +n items not in DB]
    Dead files: [list or "none"]

── What To Do Next ────────────────────
  [Phase 6 output — max 5 concrete actions]

── Summary ────────────────────────────
  Known:      [n] roadmap items
  Discovered: [n] untracked problems
  Total:      [n] outstanding
```

## Phase 6: Recommended Actions

After the summary, produce a concrete **"What to do next"** section. This is the most important part — it turns analysis into action.

### Selection criteria

Rank actions by:
1. **Unblocks the most other work** (e.g., a crash fix that enables 80 benchmarks to run)
2. **Highest measured impact** (worst current ratios × number of benchmarks affected)
3. **Lowest effort** (single file change vs. multi-file refactor)
4. **Prerequisite chains** (if B is blocked by A, recommend A first)

### Format

```
── What To Do Next ────────────────────

  1. IMPLEMENT #[n]: [title]
     Why: [evidence — what it fixes/improves, how many benchmarks, measured ratios]
     Effort: [estimate — which files, how many words/ops to change]
     Unblocks: [what becomes possible after this ships]

  2. IMPLEMENT #[n]: [title]
     Why: [evidence]
     Effort: [estimate]
     Unblocks: [what becomes possible]

  3. [CLOSE/ENCODE/FIX/INVESTIGATE] #[n]: [title]
     Why: [evidence]

  Housekeeping:
    - [any DB sync, dead item closure, WORK.md updates needed]
```

### Rules for recommendations

- Maximum 5 recommended actions. Prioritize ruthlessly.
- Every recommendation must say **IMPLEMENT**, **CLOSE**, **ENCODE**, **FIX**, or **INVESTIGATE** — never "ship" (nothing is implemented until code is written).
- For IMPLEMENT items: name the specific source files and Forth words that need changing.
- For CLOSE items: explain why the item is done or irrelevant.
- Never recommend re-running benchmarks or tests as a top action — that's verification, not work.
- If the #1 recommendation is the same as last time, note how long it's been the top priority. Stale top priorities indicate a blocked or underestimated task.

## Rules

- Read-only. Never write to DB, never modify files.
- Use latest available data. Don't re-run expensive commands.
- Every discovered item must cite evidence (benchmark name + ratio, failure count, code location).
- Don't diagnose what you can't see. If benchmark source isn't readable, say so.
- Group related problems. 80 individual RFAIL listings is noise; "80 RFAILs, 60 share root cause X" is signal.
- Be honest about coverage gaps. If the analysis can't determine a root cause, say "unknown — needs investigation."
