---
description: Run a parity loop (L4-L6) — feature parity, doc sync, test coverage convergence
argument-hint: <L4|L5|L6|PARITY|DOCWATCH|TESTWATCH> <target> [--flags]
---

Run a parity loop protocol.

**Input:** $ARGUMENTS

| Codename | # | Protocol | When to Use |
|----------|---|----------|-------------|
| PARITY | L4 | Feature Parity Convergence | Parallel implementations have drifted apart |
| DOCWATCH | L5 | Documentation-Implementation Parity | Docs out of sync with code |
| TESTWATCH | L6 | Test Coverage Convergence | Critical paths untested, coverage is superficial |

Read the full specs from `~/protocols/loops/02-parity-loops.md`.
Read the universal loop contract from `~/protocols/loops/INDEX.md`.
Execute the selected protocol with DORO depth-2.

**Every iteration must produce a REFLECTION ARTIFACT. No exceptions.**
