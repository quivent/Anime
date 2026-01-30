Validate findings through multiple independent analyses that should converge.

Usage: Apply the Watson principle - confidence emerges from consensus across independent approaches.

**The Triangulation Principle:** Watson didn't trust any single algorithm. It ran 100+ different techniques and looked for convergence. Agreement across independent methods = high confidence.

---

## Why Triangulation Matters

A single analysis can be wrong. Two analyses might share biases. But when multiple INDEPENDENT approaches arrive at the same answer, confidence compounds:

- Different algorithms, same answer = robust finding
- Different data subsets, same pattern = generalizable
- Different analysts, same conclusion = not observer bias

This command structures multi-method validation.

---

## Execution Architecture

```
Phase 1:  Finding Specification ────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Method Design ───────────┬── Approach A Designer ──────────┐
          (Parallel)               ├── Approach B Designer ──────────┼──→ Method Pool
                                   ├── Approach C Designer ──────────┤
                                   └── Independence Auditor ─────────┘
    │
    ▼
Phase 3:  Parallel Execution ──────┬── Approach A Executor ──────────┐
          (Parallel)               ├── Approach B Executor ──────────┼──→ Results
                                   └── Approach C Executor ──────────┘
    │
    ▼
Phase 4:  Convergence Analysis ─────────────────────────────────────── [Sequential]
```

---

## Phase 1: Finding Specification (Sequential)

```yaml
Task: Finding Formalizer
Subagent: general-purpose
Prompt: |
  Finding to validate: $ARGUMENTS

  Formalize what we're trying to triangulate:

  {
    "finding": "Precise statement of the claim",
    "type": "existence / magnitude / relationship / causation",
    "observable": "What we should see if true",
    "not_observable": "What we should see if false",
    "confidence_threshold": "What convergence level would be convincing"
  }

  Be precise. Vague findings can't be triangulated.
```

---

## Phase 2: Method Design (Parallel Pool)

```yaml
Task: Approach A Designer
Subagent: general-purpose
Prompt: |
  Finding: {formalized_finding}

  Design the FIRST independent approach:

  {
    "approach_id": "A",
    "method_type": "e.g., statistical, qualitative, experimental",
    "data_source": "What data this approach uses",
    "algorithm": "How the analysis works",
    "expected_output": "What form the result takes",
    "independence_from_B": "Why this is independent from other approaches",
    "potential_biases": "Known limitations of this method"
  }

  This approach should be able to stand alone.
```

```yaml
Task: Approach B Designer
Subagent: general-purpose
Prompt: |
  Finding: {formalized_finding}
  Approach A: {approach_a}

  Design SECOND independent approach that differs from A:

  {
    "approach_id": "B",
    "method_type": "DIFFERENT from A",
    "data_source": "DIFFERENT or independent subset",
    "algorithm": "DIFFERENT methodology",
    "expected_output": "Should be comparable to A",
    "independence_from_A": "Explicit statement of independence",
    "potential_biases": "Known limitations, should be DIFFERENT from A's"
  }

  If A's biases could produce false positive, B's biases should not.
```

```yaml
Task: Approach C Designer
Subagent: general-purpose
Prompt: |
  Finding: {formalized_finding}
  Approaches A, B: {approach_a}, {approach_b}

  Design THIRD independent approach:

  {
    "approach_id": "C",
    "method_type": "DIFFERENT from A and B",
    "data_source": "DIFFERENT or independent",
    "algorithm": "DIFFERENT methodology",
    "expected_output": "Comparable",
    "independence_from_A_B": "How this differs from both",
    "potential_biases": "DIFFERENT from A and B"
  }

  Three independent approaches with different biases = strong triangulation.
```

```yaml
Task: Independence Auditor
Subagent: general-purpose
Prompt: |
  All approaches: {approach_a}, {approach_b}, {approach_c}

  Audit independence:

  {
    "A_B_independence": {
      "score": 0.X,
      "shared_assumptions": [...],
      "shared_data": [...],
      "potential_correlation": "description"
    },
    "A_C_independence": {...},
    "B_C_independence": {...},
    "overall_independence": 0.X,
    "recommendation": "Proceed / Need more diverse approach"
  }

  Be critical. Apparent independence often hides shared assumptions.
```

---

## Phase 3: Parallel Execution (Parallel Pool)

```yaml
Task: Approach A Executor
Subagent: general-purpose
Prompt: |
  Finding: {formalized_finding}
  Method: {approach_a}

  Execute Approach A:

  1. Apply the specified methodology
  2. Generate results in expected format
  3. Estimate confidence in THIS approach's conclusion

  {
    "approach_id": "A",
    "finding_supported": true/false,
    "magnitude": "if applicable",
    "confidence": 0.X,
    "key_evidence": "what drove the conclusion",
    "caveats": "limitations of this analysis",
    "raw_output": "detailed results"
  }

  Be honest about what this approach can and cannot show.
```

```yaml
Task: Approach B Executor
Subagent: general-purpose
Prompt: |
  [Same structure as A, for approach B]
```

```yaml
Task: Approach C Executor
Subagent: general-purpose
Prompt: |
  [Same structure as A, for approach C]
```

---

## Phase 4: Convergence Analysis (Sequential)

```yaml
Task: Convergence Analyzer
Subagent: general-purpose
Prompt: |
  Results: {approach_a_results}, {approach_b_results}, {approach_c_results}
  Independence: {independence_audit}

  Analyze convergence:

  {
    "finding_supported_by": ["A", "B", "C"] or subset,
    "convergence_pattern": "full / partial / divergent",
    "agreement_score": 0.X,
    "weighted_confidence": 0.X,
    "discrepancies": [
      {
        "between": ["A", "B"],
        "nature": "what differs",
        "resolution": "which to believe and why"
      }
    ],
    "triangulated_conclusion": "final verdict with confidence"
  }

  Apply Watson principle:
  - All agree = high confidence
  - 2/3 agree = moderate confidence, investigate discrepancy
  - All disagree = low confidence, finding not validated
```

---

## Output Format

```
================================================================================
                    FERRUCCI TRIANGULATION REPORT
================================================================================

FINDING: {formalized_finding}

--------------------------------------------------------------------------------
                         APPROACH SUMMARY
--------------------------------------------------------------------------------

Approach    Method              Data Source         Independence
-------------------------------------------------------------------------------
A           [method]            [source]            High / Medium / Low
B           [method]            [source]            High / Medium / Low
C           [method]            [source]            High / Medium / Low

Independence Matrix:
         A      B      C
    A    -      0.X    0.X
    B    0.X    -      0.X
    C    0.X    0.X    -

Overall Independence Score: [0.X]

--------------------------------------------------------------------------------
                         INDIVIDUAL RESULTS
--------------------------------------------------------------------------------

APPROACH A: [method]
  Finding Supported: [Yes/No]
  Confidence: [0.X]
  Key Evidence: [summary]
  Caveats: [limitations]

APPROACH B: [method]
  Finding Supported: [Yes/No]
  Confidence: [0.X]
  Key Evidence: [summary]
  Caveats: [limitations]

APPROACH C: [method]
  Finding Supported: [Yes/No]
  Confidence: [0.X]
  Key Evidence: [summary]
  Caveats: [limitations]

--------------------------------------------------------------------------------
                         CONVERGENCE ANALYSIS
--------------------------------------------------------------------------------

PATTERN: [Full Convergence / Partial Convergence / Divergence]

[If Full Convergence:]
All three independent approaches support the finding.
Combined confidence: [0.X] (weighted by independence and individual confidence)
This is a ROBUST finding.

[If Partial Convergence:]
Approaches [X] and [Y] agree, but [Z] differs.
Discrepancy analysis:
  • [Z]'s result may be due to [explanation]
  • Recommendation: [how to resolve]
Combined confidence: [0.X]
This finding is PROBABLE but needs [additional validation].

[If Divergence:]
Approaches produced inconsistent results.
Possible explanations:
  • Finding is not real
  • Approaches measure different things
  • Hidden variable affects some approaches
Combined confidence: [0.X]
This finding is NOT VALIDATED.

--------------------------------------------------------------------------------
                      TRIANGULATED VERDICT
--------------------------------------------------------------------------------

FINDING: {original_finding}

VERDICT: [VALIDATED / PROBABLY TRUE / UNCERTAIN / NOT SUPPORTED]

CONFIDENCE: [0.X]

EVIDENCE QUALITY:
  ███████████░░░░░░░░░ [X/10]

RECOMMENDATION:
  [Specific next step based on verdict]

================================================================================
      "Watson ran hundreds of algorithms. When they agreed, we had
       confidence. When they disagreed, we had a research problem."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - The finding or claim to triangulate across multiple independent analyses

---

## Ferrucci Principles Embodied

1. **Many Experts**: Multiple approaches, each with specialized strengths
2. **Pervasive Confidence**: Each approach contributes calibrated uncertainty
3. **Independence Matters**: Correlated errors don't triangulate
4. **Convergence = Confidence**: Agreement across methods strengthens belief
5. **Honest Divergence**: When approaches disagree, admit uncertainty

