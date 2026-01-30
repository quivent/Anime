Isolate component contributions through systematic removal.

Usage: Remove components one at a time to understand what's actually doing the work.

**The Ablation Principle:** If you can't remove a component and see the system degrade, that component isn't contributing. Real contributions are measurable by their absence.

---

## Why Ablation Matters

Systems accrete complexity. Features get added. Nobody removes anything. Eventually you have 100 components and no idea which ones matter.

Ablation is brutal simplification:
- Remove component → measure impact
- No impact → component is dead weight
- Big impact → component is critical
- Small impact → component is nice-to-have

This command structures systematic ablation studies.

---

## Execution Architecture

```
Phase 1:  Component Inventory ──────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Dependency Analysis ─────────────────────────────────────── [Sequential]
    │
    ▼
Phase 3:  Ablation Design ─────────┬── Single Ablation Designer ─────┐
          (Parallel)               ├── Cumulative Ablation Designer ─┼──→ Ablation Plan
                                   └── Minimal Core Identifier ──────┘
    │
    ▼
Phase 4:  Ablation Execution ──────┬── Ablation 1 ───────────────────┐
          (Parallel where possible)├── Ablation 2 ───────────────────┼──→ Results
                                   └── Ablation N ───────────────────┘
    │
    ▼
Phase 5:  Contribution Analysis ────────────────────────────────────── [Sequential]
```

---

## Phase 1: Component Inventory (Sequential)

```yaml
Task: Component Cataloger
Subagent: general-purpose
Prompt: |
  System to ablate: $ARGUMENTS

  Create complete component inventory:

  {
    "components": [
      {
        "id": "C1",
        "name": "descriptive name",
        "description": "what it does",
        "type": "data / algorithm / feature / preprocessing / postprocessing",
        "removable": true/false,
        "removal_complexity": "trivial / moderate / difficult"
      }
    ],
    "total_components": N,
    "architecture_summary": "high-level system description"
  }

  Include everything. Even components you think are essential.
  Assumptions about essentiality should be tested, not assumed.
```

---

## Phase 2: Dependency Analysis (Sequential)

```yaml
Task: Dependency Mapper
Subagent: general-purpose
Prompt: |
  Components: {component_inventory}

  Map dependencies:

  {
    "dependency_graph": {
      "C1": ["C3", "C5"],  // C1 depends on C3 and C5
      "C2": [],            // C2 has no dependencies
      ...
    },
    "removal_order": ["C7", "C4", "C2", ...],  // Safe removal order
    "bundled_components": [
      ["C1", "C3"],  // Must be removed together
      ...
    ],
    "truly_independent": ["C2", "C6", ...]  // Can remove in isolation
  }

  Identify which components can be ablated independently.
```

---

## Phase 3: Ablation Design (Parallel Pool)

```yaml
Task: Single Ablation Designer
Subagent: general-purpose
Prompt: |
  Components: {component_inventory}
  Dependencies: {dependency_map}

  Design single-component ablation experiments:

  For each removable component:
  {
    "ablation_id": "ABL-C1",
    "component_removed": "C1",
    "co-removed": ["C3"],  // Due to dependencies
    "control_condition": "Full system",
    "expected_impact": "prediction before running",
    "measurement_metric": "how to measure impact",
    "success_criteria": "what constitutes meaningful degradation"
  }

  Be explicit about what "impact" means for each ablation.
```

```yaml
Task: Cumulative Ablation Designer
Subagent: general-purpose
Prompt: |
  Components: {component_inventory}
  Dependencies: {dependency_map}

  Design cumulative ablation series:

  Start with least expected impact, progressively remove more:

  {
    "ablation_series": [
      {"step": 1, "remove": ["C7"], "cumulative_removed": ["C7"]},
      {"step": 2, "remove": ["C4"], "cumulative_removed": ["C7", "C4"]},
      ...
    ],
    "stopping_criterion": "when to stop removing",
    "expected_degradation_curve": "prediction of how performance drops"
  }

  This reveals which components provide marginal vs. essential value.
```

```yaml
Task: Minimal Core Identifier
Subagent: general-purpose
Prompt: |
  Components: {component_inventory}
  Dependencies: {dependency_map}

  Hypothesize minimal viable core:

  {
    "minimal_core": ["C1", "C5"],
    "rationale": "why these are truly essential",
    "expected_performance": "what minimal core achieves",
    "ablation_target": "remove everything EXCEPT minimal core"
  }

  This tests the hypothesis: "Everything else is optimization, not foundation."
```

---

## Phase 4: Ablation Execution (Parallel where possible)

```yaml
Task: Ablation Executor
Subagent: general-purpose
Prompt: |
  Ablation: {ablation_spec}
  System: {system_reference}

  Execute ablation:

  1. Create ablated version of system
  2. Run evaluation benchmark
  3. Compare to control (full system)

  {
    "ablation_id": "ABL-C1",
    "component_removed": "C1",
    "control_performance": 0.XX,
    "ablated_performance": 0.XX,
    "delta": -0.XX,
    "relative_impact": "X% degradation",
    "failure_modes": "any new failure patterns",
    "surprising_findings": "unexpected behavior"
  }

  Be precise about measurements. Vague impressions aren't ablation.
```

---

## Phase 5: Contribution Analysis (Sequential)

```yaml
Task: Contribution Analyzer
Subagent: general-purpose
Prompt: |
  All ablation results: {ablation_results}

  Analyze component contributions:

  {
    "contribution_ranking": [
      {"component": "C1", "contribution": "high", "evidence": "15% degradation"},
      {"component": "C4", "contribution": "none", "evidence": "0% degradation"},
      ...
    ],
    "essential_components": ["C1", "C5"],
    "dead_weight": ["C4", "C7", "C9"],
    "marginal_value": ["C2", "C3"],
    "minimal_viable_system": {
      "components": [...],
      "performance_retained": "X%",
      "complexity_reduction": "Y%"
    },
    "recommendations": "what to keep, remove, investigate"
  }
```

---

## Output Format

```
================================================================================
                    FERRUCCI ABLATION STUDY
================================================================================

SYSTEM: {system_description}
COMPONENTS ANALYZED: [N]

--------------------------------------------------------------------------------
                         ABLATION RESULTS
--------------------------------------------------------------------------------

Component         Removal Impact    Contribution    Verdict
-------------------------------------------------------------------------------
C1 [name]         -15%              CRITICAL        KEEP
C2 [name]         -3%               MARGINAL        EVALUATE
C3 [name]         -0.1%             NEGLIGIBLE      REMOVE
C4 [name]         +2%               HARMFUL         REMOVE (improves!)
C5 [name]         -22%              CRITICAL        KEEP
...

--------------------------------------------------------------------------------
                      CONTRIBUTION DISTRIBUTION
--------------------------------------------------------------------------------

CRITICAL (>10% impact):    [N] components
MODERATE (5-10% impact):   [N] components
MARGINAL (1-5% impact):    [N] components
NEGLIGIBLE (<1% impact):   [N] components
HARMFUL (negative):        [N] components  <-- These are actively hurting

--------------------------------------------------------------------------------
                      CUMULATIVE ABLATION
--------------------------------------------------------------------------------

Components    Performance    Δ from Full    Components Remaining
Removed       (% of full)
-------------------------------------------------------------------------------
0             100%           -              [all]
1             99%            -1%            [list]
2             97%            -3%            [list]
3             95%            -5%            [list]
...
N-2           78%            -22%           [minimal core]
N-1           23%            -77%           [single component]
N             0%             -100%          [empty]

Knee of the curve: [X] components removed (after this, degradation accelerates)

--------------------------------------------------------------------------------
                      MINIMAL VIABLE SYSTEM
--------------------------------------------------------------------------------

HYPOTHESIS: Only [C1, C5, C8] are truly necessary.

RESULT:
  • Full system: 100% performance with [N] components
  • Minimal core: [X]% performance with [3] components
  • Complexity reduction: [Y]%

CONCLUSION: [Assessment of whether minimal core is viable]

--------------------------------------------------------------------------------
                      RECOMMENDATIONS
--------------------------------------------------------------------------------

REMOVE (no impact):
  • C3: [0.1% impact, not worth complexity]
  • C4: [Actively harmful, remove immediately]
  • C7: [Never triggered in evaluation]

INVESTIGATE (small impact, may be contextual):
  • C2: [3% impact, but may matter in specific cases]

KEEP (verified critical):
  • C1: [15% impact, core algorithm]
  • C5: [22% impact, essential preprocessing]

SIMPLIFIED ARCHITECTURE:
  [Diagram of minimal viable system]

================================================================================
      "If you can't measure what a component contributes by removing it,
       you don't understand your system."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - System, model, or pipeline to ablate

---

## Ferrucci Principles Embodied

1. **Measurement over intuition**: Don't assume - remove and measure
2. **Complexity is debt**: Every component should earn its place
3. **Architecture clarity**: Ablation reveals true structure
4. **Honest assessment**: Some components hurt performance
5. **Minimal viable system**: Find the core, remove the rest

