---
description: Run a completeness loop (L1-L3) — feature gaps, API surface, integration wiring
argument-hint: <L1|L2|L3|GAPCLOSE|SURFACE|NEXUS> <target> [--flags]
---

Run a completeness loop protocol.

**Input:** $ARGUMENTS

| Codename | # | Protocol | When to Use |
|----------|---|----------|-------------|
| GAPCLOSE | L1 | Feature Completeness Audit | Spec exists, implementation has gaps |
| SURFACE | L2 | API Surface Completeness | API contract not fully implemented/documented/tested |
| NEXUS | L3 | Integration Completeness | Components exist but integration points are disconnected |

Read the full specs from `~/protocols/loops/01-completeness-loops.md`.
Read the universal loop contract from `~/protocols/loops/INDEX.md`.
Execute the selected protocol with DORO depth-2.

**Every iteration must produce a REFLECTION ARTIFACT. No exceptions.**
