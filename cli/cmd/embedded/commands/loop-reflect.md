---
description: Run an iterative reflection loop protocol (L1-L18) — completeness, parity, standards, progress, meta-evolution, safeguard
argument-hint: <protocol-or-description> [target] [--flags]
---

Run a protocol from the Iterative Reflection Loop suite.

**Input:** $ARGUMENTS

## Routing

| Codename | # | Protocol | Module |
|----------|---|----------|--------|
| GAPCLOSE | L1 | Feature Completeness Audit | Completeness |
| SURFACE | L2 | API Surface Completeness | Completeness |
| NEXUS | L3 | Integration Completeness | Completeness |
| PARITY | L4 | Feature Parity Convergence | Parity |
| DOCWATCH | L5 | Documentation-Implementation Parity | Parity |
| TESTWATCH | L6 | Test Coverage Convergence | Parity |
| CONFORM | L7 | Design Standard Enforcement | Standards |
| ASCEND | L8 | Progressive Quality Elevation | Standards |
| BASELINE | L9 | Performance Baseline Maintenance | Standards |
| MERIDIAN | L10 | Project Fulfillment Tracking | Progress |
| APEX | L11 | Milestone Convergence | Progress |
| PERIMETER | L12 | Scope Integrity | Progress |
| METACYCLE | L13 | Evolution Quality Reflection | Meta-Evolution |
| CADENCE | L14 | Enhancement Velocity | Meta-Evolution |
| MORPHOGEN | L15 | Capability Growth Reflection | Meta-Evolution |
| REGRESS | L16 | Regression Detection | Safeguard |
| DEBTWATCH | L17 | Technical Debt Tracking | Safeguard |
| ARBITER | L18 | Convergence Termination | Safeguard |

## Quick Decision

- **Drive implementation to spec?** L1 GAPCLOSE
- **Ensure API surface is complete?** L2 SURFACE
- **Wire all integration points?** L3 NEXUS
- **Bring platforms to parity?** L4 PARITY
- **Keep docs in sync?** L5 DOCWATCH
- **Drive test coverage?** L6 TESTWATCH
- **Enforce design standard?** L7 CONFORM
- **Raise code quality?** L8 ASCEND
- **Maintain perf baselines?** L9 BASELINE
- **Track project toward ship?** L10 MERIDIAN
- **Hit a deadline?** L11 APEX
- **Prevent scope creep?** L12 PERIMETER
- **Improve dev process?** L13 METACYCLE
- **Balance speed and quality?** L14 CADENCE
- **Check growth direction?** L15 MORPHOGEN
- **Catch regressions?** L16 REGRESS
- **Track tech debt?** L17 DEBTWATCH
- **Stop loops that converged?** L18 ARBITER

Read the INDEX at `~/protocols/loops/INDEX.md` for the universal loop contract and composition patterns.
Read the protocol spec from the appropriate module file in `~/protocols/loops/`.
Execute the protocol on the given target using the DORO depth-2 implementation spec.

**Every iteration must produce a REFLECTION ARTIFACT. No exceptions.**
