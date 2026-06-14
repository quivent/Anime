---
description: Run a safeguard loop (L16-L18) — regression detection, debt tracking, convergence termination
argument-hint: <L16|L17|L18|REGRESS|DEBTWATCH|ARBITER> <target> [--flags]
---

Run a safeguard loop protocol.

**Input:** $ARGUMENTS

| Codename | # | Protocol | When to Use |
|----------|---|----------|-------------|
| REGRESS | L16 | Regression Detection | Recent improvements may have broken existing behavior |
| DEBTWATCH | L17 | Technical Debt Tracking | Debt may be accumulating faster than it's being repaid |
| ARBITER | L18 | Convergence Termination | Multiple loops running, need to terminate converged ones |

These are adversarial loops — they protect the system from its own improvement process.

Read the full specs from `~/protocols/loops/06-safeguard-loops.md`.
Read the universal loop contract from `~/protocols/loops/INDEX.md`.
Execute the selected protocol with DORO depth-2.

**Every iteration must produce a REFLECTION ARTIFACT. No exceptions.**
