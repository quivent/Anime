---
description: Fresh project status — run tests, benchmarks, persist snapshots to DB
allowed-tools: Bash, Read, Glob, Grep, Task
---

# /status — Project Status Check

Run a **fresh** status check of the current project. Execute real commands, report results, and **persist snapshots to `projects.db`** so `/dashboard` has current data.

## What to Report

### 1. Tests (run fresh)

Look for test infrastructure and run it. Common patterns:
- `./tests/test`, `npm test`, `cargo test`, `pytest`, `make test`, etc.
- Report: total pass/fail/skip counts, any WRONG results
- If tests produce too much output, summarize the final counts

### 2. Build / Verification (run fresh)

Check that the project builds and any integrity checks pass:
- Self-hosting verification, type checking, lint, compilation, etc.
- Report: pass/fail with brief detail on failures

### 3. Benchmarks (if applicable, run fresh)

If the project has benchmarks, run them:
- Report: key aggregate metrics, comparison to baseline if available
- Keep it to headline numbers — don't dump raw tables

### 4. Known Work Remaining

Query roadmap from `projects.db` (don't re-read WORK.md):
```sql
SELECT category, title, priority, status, impact
  FROM roadmap WHERE project_id = (SELECT id FROM projects WHERE path = '$(pwd)')
  ORDER BY priority ASC;
```

## Persist Snapshots

After collecting fresh results, **update `metric_snapshots`** in `projects.db` so the data is available to `/dashboard`.

### Procedure

1. Query `metric_sources` for this project to get source IDs, fields, and patterns
2. For each command you already ran, apply the stored regex patterns to extract field values
3. Insert new snapshot rows:

```sql
INSERT INTO metric_snapshots (id, project_id, source_id, category, name, field, value, collected_at, git_hash)
VALUES ('[name]-snap-[timestamp]-[field]', '[project_id]', '[source_id]', '[cat]', '[name]', '[field]', '[value]', datetime('now'), '[git_hash]'),
       ...;
```

4. Prune old snapshots (keep last 10 per source):

```sql
DELETE FROM metric_snapshots WHERE id IN (
  SELECT id FROM metric_snapshots ms
  WHERE ms.project_id = '[project_id]'
    AND ms.source_id = '[source_id]'
  ORDER BY collected_at DESC
  LIMIT -1 OFFSET 10
);
```

### Per-Item Snapshots

When a roadmap item is marked as `done` (status changes), capture a tagged snapshot linking the metric state to that completion event. This enables per-item before/after comparison.

After changing a roadmap item's status to `done`:
```sql
INSERT INTO baselines (id, project_id, name, source_id, field, value, git_hash, created_at)
SELECT '[name]-done-[roadmap_id]-' || ms.field, ms.project_id, 'done:[roadmap_title]', ms.source_id, ms.field, ms.value, ms.git_hash, datetime('now')
FROM metric_snapshots ms
WHERE ms.project_id = '[project_id]'
  AND ms.collected_at = (
    SELECT MAX(ms2.collected_at) FROM metric_snapshots ms2
    WHERE ms2.source_id = ms.source_id
  );
```

This creates a named baseline like `done:Variable-depth tracking` so you can see what the metrics were when each item was completed.

## Output Format

```
PROJECT STATUS: [project name]
Date: [today]
Branch: [current branch] @ [short hash]

── Tests ──────────────────────────────
[results]

── Build / Verification ───────────────
[results]

── Benchmarks ─────────────────────────
[results or "N/A"]

── Progress vs Baseline ───────────────
  Roadmap:  [done]/[total] complete ([pct]%)
  Tests:    [base_pass] → [cur_pass]  (+[delta] passing)
  Wrong:    [base_wrong] → [cur_wrong]
  Ratio:    [base_ratio]x → [cur_ratio]x GCC
  [Show only if baselines exist in DB]

── Remaining Work ─────────────────────
[from roadmap table, concise list]

── Snapshots ──────────────────────────
  Updated [n] metric snapshots in projects.db
```

## Rules

- Run everything fresh. No cached data, no stale numbers.
- Report raw results. Do not editorialize or suggest fixes.
- If a check fails, report the failure — do not try to fix it.
- If a check takes too long (>2 min), report timeout and move on.
- Keep the report concise. Signal, not noise.
- Always persist snapshots — this is what feeds `/dashboard`.
