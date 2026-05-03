# /kernighan-audit

Read the code cold. Report what Kernighan would fix. Change nothing.

Scope: $ARGUMENTS (file, glob, directory; no arg = project source).

## Method

Read each file as if you didn't write it. No diffs, no history, no context about prior passes. Just the code as it is now. The telephone test: read it aloud. Note everything that makes you stumble.

## P0: Measure

Count chars in scope: `find <scope> -name '*.rs' -o -name '*.svelte' -o -name '*.ts' | xargs wc -c`

## P1: Scan

Read every file top to bottom. For each, tally:

| Category | What to count |
|----------|---------------|
| **Comment noise** | Restatement docs, banners, echoes, changelog, commented-out code, field docs restating name |
| **Comment signal** | *Why* comments, SAFETY, real TODOs, algorithm explanations — these stay |
| **Air** | Blank line clusters, multi-line expressions that fit one line, single-use variables naming nothing |
| **Name bloat** | Redundant prefixes/suffixes, `get_` getters, long names in short scopes, inconsistent conventions |
| **Verbose idioms** | `if let Some/else`, `match bool`, repeated `.map_err(format!())`, nested iterator chains |
| **Telephone failures** | Lines you'd have to re-read to explain aloud |

## P2: Report

```
FILES: [n] scanned, [n] chars total

Per-file summary (worst offenders first):
  [filename]: [n] comment noise, [n] air, [n] name bloat, [n] verbose idioms
  [filename]: ...

Totals:
  Comment noise:    [n] instances, est. -[n] chars
  Air:              [n] instances, est. -[n] chars
  Name bloat:       [n] instances, est. -[n] chars
  Verbose idioms:   [n] instances, est. -[n] chars
  TOTAL:            est. -[n] chars (-[x.x]%)

Signal preserved:   [n] *why* comments, [n] SAFETY blocks, [n] algorithm docs

Top 5 worst lines (file:line — what's wrong):
  1. ...
  2. ...
  3. ...
  4. ...
  5. ...
```

## Rules

- **Change nothing.** This is assessment, not surgery.
- Estimate char savings honestly. Round down, not up.
- If a file reads clean, say so. Not everything needs compression.
- Flag anything where compression would hurt clarity — that's signal too.
