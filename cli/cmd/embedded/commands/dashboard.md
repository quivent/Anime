---
description: Project dashboard — roadmap, metrics, progress vs baseline, audit — from projects.db
allowed-tools: Bash
---

# /dashboard — Project Health Dashboard

Read-only. Shows roadmap, stored metrics, **progress vs baseline**, and **auto-audit** from `projects.db`. No commands executed — this is instant.

For **fresh** data, use `/status`. For **encoding structure**, use `/app-topology`.

---

## Execution

**One Bash call. One database open.**

```bash
CWD=$(pwd)
HASH=$(git rev-parse HEAD 2>/dev/null || echo "no-git")
sqlite3 -separator '	' ~/.claude/db/projects.db << ENDSQL
.headers off

SELECT '=== PROJECT ===' as tag;
SELECT name, domain, sensitivity FROM projects WHERE path = '$CWD';

SELECT '=== METRICS ===' as tag;
SELECT ms.category, ms.name, ms.field, ms.value, ms.collected_at, ms.git_hash
  FROM metric_snapshots ms
  WHERE ms.project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND ms.collected_at = (
      SELECT MAX(ms2.collected_at) FROM metric_snapshots ms2
      WHERE ms2.source_id = ms.source_id
    )
  ORDER BY ms.category, ms.name, ms.field;

SELECT '=== BASELINES ===' as tag;
SELECT b.name, b.field, b.value, b.git_hash, b.created_at
  FROM baselines b
  WHERE b.project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY b.field;

SELECT '=== ROADMAP ===' as tag;
SELECT category, title, priority, status, perf, ops, impact, blocked_by
  FROM roadmap
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY priority ASC;

SELECT '=== ROADMAP_COUNTS ===' as tag;
SELECT
  COUNT(*) as total,
  SUM(CASE WHEN status = 'todo' THEN 1 ELSE 0 END) as todo,
  SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
  SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) as done,
  SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END) as blocked
FROM roadmap
WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

SELECT '=== COMPLETED ===' as tag;
SELECT r.title, r.completed_at, r.snapshot_name,
       b.field, b.value
  FROM roadmap r
  LEFT JOIN baselines b ON b.project_id = r.project_id AND b.name = r.snapshot_name
  WHERE r.project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND r.status = 'done'
  ORDER BY r.completed_at DESC, b.field;

SELECT '=== ARCH_CONSTRAINTS ===' as tag;
SELECT content, severity
  FROM constraints
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND type = 'architecture'
  ORDER BY CASE severity WHEN 'absolute' THEN 1 WHEN 'strong' THEN 2 ELSE 3 END;

SELECT '=== ROOT_CAUSES ===' as tag;
SELECT title, description, perf
  FROM roadmap
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND status != 'done'
    AND description IS NOT NULL AND description != ''
  ORDER BY priority ASC
  LIMIT 10;

SELECT '=== HISTORY ===' as tag;
SELECT title, perf, ops, impact, completed_at
  FROM roadmap
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
    AND status = 'done'
  ORDER BY priority ASC;

SELECT '=== BENCH_RUN ===' as tag;
SELECT id, timestamp, arch, compiler, git_branch, git_hash,
       total, passed, cfail, gcfail, rfail, grfail, sixth_wins,
       runtime_ratio, sum_compile_sixth_ms, sum_compile_gcc_ms,
       sum_run_sixth_ms, sum_run_gcc_ms, wall_time_ms
  FROM benchmark_runs
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY timestamp DESC LIMIT 1;

SELECT '=== BENCH_RESULTS ===' as tag;
SELECT name, status, compile_sixth_ms, compile_gcc_ms,
       run_sixth_ms, run_gcc_ms, ratio
  FROM benchmark_results
  WHERE run_id = (SELECT id FROM benchmark_runs
                  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
                  ORDER BY timestamp DESC LIMIT 1)
  ORDER BY
    CASE WHEN status = 'PASS' AND ratio IS NOT NULL THEN 0 ELSE 1 END,
    CASE WHEN status = 'PASS' AND ratio IS NOT NULL THEN -ratio ELSE 0 END,
    name;

SELECT '=== RATIO_DIST ===' as tag;
SELECT
  SUM(CASE WHEN ratio < 1.0 THEN 1 ELSE 0 END),
  SUM(CASE WHEN ratio >= 1.0 AND ratio < 2.0 THEN 1 ELSE 0 END),
  SUM(CASE WHEN ratio >= 2.0 AND ratio < 5.0 THEN 1 ELSE 0 END),
  SUM(CASE WHEN ratio >= 5.0 AND ratio < 10.0 THEN 1 ELSE 0 END),
  SUM(CASE WHEN ratio >= 10.0 AND ratio < 100.0 THEN 1 ELSE 0 END),
  SUM(CASE WHEN ratio >= 100.0 THEN 1 ELSE 0 END)
  FROM benchmark_results
  WHERE run_id = (SELECT id FROM benchmark_runs
                  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
                  ORDER BY timestamp DESC LIMIT 1)
    AND status = 'PASS' AND ratio IS NOT NULL;

SELECT '=== CACHE ===' as tag;
SELECT git_hash, total_loc, file_count
  FROM exploration_cache
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

ENDSQL
```

## Output Format

Parse the tagged output. Compare current metrics against baselines to compute deltas.

### Progress Computation

For each metric field that has both a current snapshot and a baseline value:

| Field | Progress Formula |
|-------|-----------------|
| `pass` | `(current - baseline) / (baseline_total - baseline_pass) * 100` = % of remaining failures fixed |
| `wrong` | `baseline - current` = wrongs eliminated (show as negative = good) |
| `runtime_ratio` | `(baseline - current) / baseline * 100` = % faster (lower ratio = better) |
| `total` | `current - baseline` = new tests added |
| `status` (selfhost) | `0 = PASS`, `nonzero = FAIL` — binary |

**Roadmap completion**: `done / total * 100` = % of work items completed.

### Display

```
╭─────────────────────────────────────────────────────────────╮
│  DASHBOARD                                                    │
│  [name] — baseline: [baseline_name] @ [baseline_hash_short]  │
╰─────────────────────────────────────────────────────────────╯

── Progress vs Baseline ───────────────
  Roadmap:    [done]/[total] complete ([pct]%)
  Tests:      [base_pass]/[base_total] → [cur_pass]/[cur_total]  (+[delta] passing)
  Wrong:      [base_wrong] → [cur_wrong]  ([delta])
  Ratio:      [base_ratio]x → [cur_ratio]x GCC  ([pct]% faster)
  Selfhost:   [PASS|FAIL]
  [If no baseline: "No baseline set — run /project-encode to create one"]
  [If no benchmark data yet: "Benchmarks: awaiting first collection"]

── Tests ──────────────────────────────
  PASS: [pass]/[total]   WRONG: [wrong]   CFAIL: [cfail]   RFAIL: [rfail]

── Benchmarks ─────────────────────────
  Passed: [passed]/[total]   Runtime: [runtime_ratio]x GCC
  Sixth wins: [sixth_wins]   Wall time: [wall_time]s
  Compile: [sum_compile_sixth]ms sixth vs [sum_compile_gcc]ms GCC ([compile_ratio]x faster)
  [If no snapshots: "No data collected — run /status to collect"]

  Ratio Distribution:
    < 1x (Sixth wins): [n]
    1-2x:   [n]
    2-5x:   [n]
    5-10x:  [n]
    10-100x: [n]
    100x+:  [n]
    FAIL:   [n]

── All Benchmarks (sorted by ratio, slowest first) ──
  Render the full BENCH_RESULTS table from the DB query. All 240 benchmarks.
  PASS benchmarks first sorted by ratio descending, then FAILs alphabetically.

  | # | Benchmark | Status | Compile (6th) | Compile (GCC) | Runtime (6th) | Runtime (GCC) | Ratio |
  |:--|:----------|:-------|:--------------|:--------------|:--------------|:--------------|:------|
  | 1 | depth7 | PASS | 13ms | 53ms | 2509ms | 1ms | 2509.0x |
  | 2 | depth3 | PASS | 23ms | 95ms | 1264ms | 1ms | 1264.0x |
  ...
  | 159 | alias1 | PASS | 19ms | 120ms | 1ms | 7ms | 0.1x |
  | 160 | ack | RFAIL | 18ms | 125ms | — | — | — |
  ...

── Verification ───────────────────────
  Selfhost: [PASS|FAIL] — [message]

── Collected ──────────────────────────
  Timestamp: [collected_at]
  Commit:    [git_hash_short]
  Current:   [HEAD_short]
  [If hashes match: "FRESH"]
  [If hashes differ: "STALE — data is from [n] commits ago"]

── Roadmap ([n] remaining) ───────────────
  Render as aligned markdown table. Only show items with status != 'done'.
  Group by category. Use status icons: ⬜ = todo, 🔄 = in_progress, ✅ = done.
  Perf column: multiplier (2x), instruction savings (-2ins), or — if N/A.

  | St | #  | Title                                | Perf   | Ops                    |
  |:---|:---|:-------------------------------------|:-------|:-----------------------|
  | ⬜ | 1  | Variable-depth tracking (reg-depth)  | —      | fixes 81 crashes       |
  | ⬜ | 2  | CSEL for simple IF patterns          | 3-10x  | saves 3-5 instr/branch |
  | 🔄 | 3  | Inlining small words                 | 2-5x   | saves BL+RET/call      |
  ...

  All columns left-aligned and padded consistently.
  [If blocked_by: append "(blocked by #N)" to Ops column]

  Summary: [x] ⬜ [y] 🔄 [z] blocked

── Root Cause Analysis ──────────────────
  [Only show if ROOT_CAUSES query returned rows]
  Show top worst performers → root cause → which roadmap item fixes it.

  | Benchmark   | Ratio | Root Cause                    | Fix     |
  |:------------|:------|:------------------------------|:--------|
  | permute     | 14.0x | deep recursion, no inlining   | #3      |
  | branch2     | 10.0x | branch vs CSEL, 50% mispred   | #2      |
  | linsearch   | 8.0x  | 100K function calls            | #3      |
  ...

  Extract from roadmap descriptions: parse benchmark names and ratios where present.

── Architecture Constraints ─────────────
  [Only show if ARCH_CONSTRAINTS query returned rows]
  Hard-learned rules from past regressions:

  • [content — abbreviated to one line]
  ...

── Optimization History ─────────────────
  [Only show if HISTORY query returned rows]
  What has been done, with measured impact:

  | St | Title                          | Perf   | Ops           | Impact                     |
  |:---|:-------------------------------|:-------|:--------------|:---------------------------|
  | ✅ | TOS caching (X19)              | —      | -2ins/op      | memory → 1 register        |
  | ✅ | NOS caching (X21)              | 1.14x  | -1mem/binop   | 4.66x → 4.08x GCC (12.5%) |
  | ✅ | DO...LOOP register caching     | 1.39x  | -3ins/iter    | 8.98x → 6.47x GCC (28%)   |
  ...

── Metric Sources ([n]) ───────────────
  [category]/[name]: [command]
    Fields: [field1] ([type]), [field2] ([type]), ...
```

## Baseline Management

The `baselines` table stores reference values for progress comparison.

**Set automatically**: `/project-encode` captures baseline from initial metric collection.

**Reset manually**: To re-baseline at the current state:
```sql
sqlite3 ~/.claude/db/projects.db << 'SQL'
DELETE FROM baselines WHERE project_id = (SELECT id FROM projects WHERE path = '$(pwd)');
INSERT INTO baselines (id, project_id, name, source_id, field, value, git_hash, created_at)
SELECT 'base-' || ms.field, ms.project_id, 'manual-reset', ms.source_id, ms.field, ms.value, ms.git_hash, datetime('now')
FROM metric_snapshots ms
WHERE ms.project_id = (SELECT id FROM projects WHERE path = '$(pwd)')
  AND ms.collected_at = (SELECT MAX(ms2.collected_at) FROM metric_snapshots ms2 WHERE ms2.source_id = ms.source_id);
SQL
```

## If No Project Found

```
No project encoding found for: [cwd]
Run /project-encode first.
```

## Auto-Audit

After rendering the dashboard, audit both WORK.md and DASHBOARD.md for drift. This runs automatically every time — no separate command needed.

### Audit WORK.md (source document)

Read `WORK.md` from the project root. Compare its content against the DB encoding:

1. **Roadmap count**: Count work items in WORK.md (lines matching task/item patterns in roadmap sections). Compare against `SELECT COUNT(*) FROM roadmap WHERE project_id = ...`. If counts differ, report drift.
2. **Constraint count**: Count architecture constraint bullets in WORK.md. Compare against `SELECT COUNT(*) FROM constraints WHERE type = 'architecture' AND project_id = ...`.
3. **History count**: Count completed items in WORK.md. Compare against `SELECT COUNT(*) FROM roadmap WHERE status = 'done' AND project_id = ...`.

Display:
```
── Audit: WORK.md ─────────────────────
  Roadmap items:  WORK=[n]  DB=[n]  [SYNC | DRIFT: +n/-n]
  Constraints:    WORK=[n]  DB=[n]  [SYNC | DRIFT]
  History:        WORK=[n]  DB=[n]  [SYNC | DRIFT]
  [If any DRIFT: "Re-encode with /project-encode to sync"]
  [If WORK.md missing: "No WORK.md found — roadmap lives only in DB"]
```

### Audit DASHBOARD.md (generated output)

Check the existing `DASHBOARD.md` in the project root:

1. **Exists?** If not, note "will be created".
2. **Timestamp**: Parse `<!-- Generated by /dashboard at [datetime] -->` header. Compare age.
3. **Git hash**: If the dashboard contains a commit hash, compare against current HEAD.

Display:
```
── Audit: DASHBOARD.md ────────────────
  Status:    [CURRENT | STALE | MISSING]
  Generated: [datetime] ([age] ago)
  Commit:    [hash_short] vs HEAD [hash_short]
  [If STALE or MISSING: "Regenerating now..."]
```

The dashboard is **always regenerated** at the end of `/dashboard`, so this audit just reports the state of the *previous* generation for awareness.

### Audit in DASHBOARD.md output

Include the audit results in the written `DASHBOARD.md` file as the final section, so offline readers can see sync status.

## File Output

After rendering the dashboard to the terminal, **also write it to `DASHBOARD.md`** in the project root. This lets the user read progress outside of Claude Code.

Use the Write tool to write the rendered markdown to `[project_root]/DASHBOARD.md`. The file should contain the full dashboard output as valid markdown — same content displayed in the terminal but formatted for reading in any editor/viewer.

Include a header with generation timestamp:
```markdown
<!-- Generated by /dashboard at [datetime] — do not edit manually -->
```

This file should be in `.gitignore` (add it if not already present) — it's ephemeral output, not source.

## HTML Output

After writing DASHBOARD.md, also generate `dashboard.html` in the project root. This is a standalone HTML file (no external dependencies) that presents the same data in a browser-friendly format.

**Important:** This is `dashboard.html`, NOT `perf.html`. Do not touch or overwrite `perf.html`.

Read the template at `~/.claude/templates/dashboard.html`. Fill `{{PLACEHOLDER}}` tokens with DB data, then write to `[project_root]/dashboard.html`.

**Template:** `~/.claude/templates/dashboard.html`
**Output:** `[project_root]/dashboard.html`

### Placeholders

| Placeholder | Source |
|-------------|--------|
| `{{PROJECT_NAME}}` | `projects.name` |
| `{{DATE}}` | current date |
| `{{GIT_BRANCH}}`, `{{GIT_HASH}}` | `benchmark_runs` latest row |
| `{{BENCH_TIMESTAMP}}` | `benchmark_runs.timestamp` |
| `{{METRICS_TIMESTAMP}}` | `metric_snapshots.collected_at` |
| `{{ROADMAP_DONE}}`, `{{ROADMAP_TOTAL}}`, `{{ROADMAP_PCT}}` | `roadmap` counts |
| `{{ROADMAP_OPEN}}` | count where status != 'done' |
| `{{TEST_PASS}}`, `{{TEST_TOTAL}}`, `{{TEST_WRONG}}`, `{{TEST_CFAIL}}`, `{{TEST_RFAIL}}` | `metric_snapshots` where category='test' |
| `{{TESTS_CLASS}}` | 'good' if pass/total > 0.95, 'warn' if > 0.80, else 'bad' |
| `{{BENCH_PASSED}}`, `{{BENCH_TOTAL}}`, `{{BENCH_RFAIL}}`, `{{BENCH_GRFAIL}}` | `benchmark_runs` latest |
| `{{BENCH_CLASS}}` | 'good' if passed/total > 0.80, 'warn' if > 0.50, else 'bad' |
| `{{RUNTIME_RATIO}}` | `benchmark_runs.runtime_ratio` |
| `{{RATIO_CLASS}}` | 'good' if < 2, 'warn' if < 10, else 'bad' |
| `{{SIXTH_WINS}}` | `benchmark_runs.sixth_wins` |
| `{{WALL_TIME}}` | `benchmark_runs.wall_time_ms / 1000` (seconds) |
| `{{COMPILE_RATIO}}` | `sum_compile_gcc_ms / sum_compile_sixth_ms` |
| `{{SUM_COMPILE_SIXTH}}`, `{{SUM_COMPILE_GCC}}` | `benchmark_runs` sums |
| `{{SELFHOST_STATUS}}` | 'PASS' or 'FAIL' from metric_snapshots |
| `{{SELFHOST_CLASS}}` | 'good' if PASS, else 'bad' |
| `{{SELFHOST_MESSAGE}}` | metric_snapshots verification message |
| `{{DIST_WINS}}` ... `{{DIST_FAIL}}` | RATIO_DIST query results |
| `{{BENCH_ROWS}}` | Generated HTML `<tr>` rows for all 240 benchmarks |
| `{{ROADMAP_ROWS}}` | Generated HTML `<tr>` rows for open roadmap items |
| `{{HISTORY_ROWS}}` | Generated HTML `<tr>` rows for completed items |
| `{{HISTORY_COUNT}}` | count of done items |
| `{{CONSTRAINT_ITEMS}}` | Generated HTML `<div class="constraint">` items |
| `{{GENERATED_AT}}` | ISO datetime |

### Bench row generation

For each benchmark result, generate a `<tr>`:

```html
<tr>
  <td class="num" data-sort="[n]">[n]</td>
  <td>[name]</td>
  <td><span class="badge badge-[status_lower]">[status]</span></td>
  <td class="num">[compile_sixth]ms</td>
  <td class="num">[compile_gcc]ms</td>
  <td class="num">[run_sixth]ms</td>
  <td class="num">[run_gcc]ms</td>
  <td class="num [ratio_class]" data-sort="[ratio_or_99999]">[ratio]x</td>
</tr>
```

Ratio class: `ratio-green` (< 1), `ratio-ok` (1-5), `ratio-yellow` (5-10), `ratio-orange` (10-100), `ratio-red` (100+).
For failures: runtime columns show `—`, ratio shows `—`, data-sort=99999.

### Roadmap row generation

```html
<tr>
  <td><span class="badge badge-[status]">[icon]</span></td>
  <td class="num">[priority]</td>
  <td>[title]</td>
  <td>[category]</td>
  <td>[perf]</td>
  <td>[ops]</td>
</tr>
```

Status icons: todo=`&#x2B1C;`, in_progress=`&#x1F504;`, blocked=`&#x1F6AB;`, research=`&#x1F52C;`

## Data Source

**All dashboard data comes from the database.** No CSV files are read.

| Data | DB Table | Query |
|------|----------|-------|
| Project info | `projects` | name, domain, sensitivity |
| Test metrics | `metric_snapshots` | category = 'test' |
| Verification | `metric_snapshots` | category = 'verification' |
| Benchmark aggregates | `benchmark_runs` | latest run by timestamp |
| Per-benchmark detail | `benchmark_results` | all rows for latest run_id |
| Ratio distribution | `benchmark_results` | CASE/SUM aggregation |
| Baselines | `baselines` | all fields for project |
| Roadmap | `roadmap` | all items, grouped by category |
| Constraints | `constraints` | type = 'architecture' |
| History | `roadmap` | status = 'done' |

If `benchmark_runs` is empty, show "No benchmark data — run benchmarks and persist to DB first."

## Notes

- Read-only on the database. It changes nothing there, runs nothing.
- Instant — just a database query.
- Progress section only appears when baselines exist.
- **WORK.md** = source document (human-authored roadmap, architecture, root causes). Encoded into DB by `/project-encode`.
- **DASHBOARD.md** = generated output (metrics, progress, audit). Written by `/dashboard`. In `.gitignore`.
- **dashboard.html** = generated HTML output. Written by `/dashboard`. In `.gitignore`.
- Auto-audit runs every time — compares WORK.md against DB, flags drift.
- Use `/status` to collect fresh metrics and update snapshots.
- Use `/project-encode` to re-discover sources and re-collect.
- Roadmap items come from WORK.md at encode time. Re-encode to sync drift.
