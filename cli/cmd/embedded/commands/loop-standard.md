---
description: Run a standard enforcement loop (L7-L9) — design conformance, quality elevation, perf baselines
argument-hint: <L7|L8|L9|CONFORM|ASCEND|BASELINE> <target> [--flags]
---

Run a standard enforcement loop protocol.

**Input:** $ARGUMENTS

| Codename | # | Protocol | When to Use |
|----------|---|----------|-------------|
| CONFORM | L7 | Design Standard Enforcement | Codebase drifting from design standard |
| ASCEND | L8 | Progressive Quality Elevation | Code quality below target across multiple dimensions |
| BASELINE | L9 | Performance Baseline Maintenance | Performance may have regressed, baselines need ratification |

Read the full specs from `~/protocols/loops/03-standard-loops.md`.
Read the universal loop contract from `~/protocols/loops/INDEX.md`.
Execute the selected protocol with DORO depth-2.

**Every iteration must produce a REFLECTION ARTIFACT. No exceptions.**
