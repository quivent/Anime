Design experiments that can definitively falsify a hypothesis, not just confirm it.

Usage: Force rigorous experiment design with explicit failure criteria before running anything.

**Core Principle:** A theory that can't be falsified isn't useful. Design to prove yourself wrong.

---

## The Falsification Imperative

Most experiments are designed to confirm what the researcher already believes. This command inverts that: design the experiment to DISPROVE your hypothesis. If it survives, your confidence is earned. If it fails, you've learned something real.

---

## Execution Architecture

```
Phase 1:  Claim Formalization ──────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Decomposition ────────────┬── Sub-Hypothesis Extractor ─┐
          (Parallel)                ├── Dependency Mapper ────────┼──→ Hypothesis Tree
                                    └── Assumption Surfacer ──────┘
    │
    ▼
Phase 3:  Falsification Design ─────┬── Kill Criteria Designer ───┐
          (Parallel)                ├── Control Condition Builder ┼──→ Test Protocol
                                    └── Confound Identifier ──────┘
    │
    ▼
Phase 4:  Evidence Integration ─────────────────────────────────── [Sequential]
    │
    ▼
Phase 5:  Protocol Document ────────────────────────────────────── [Sequential]
```

---

## Phase 1: Claim Formalization (Sequential)

```yaml
Task: Claim Formalizer
Subagent: general-purpose
Prompt: |
  User's hypothesis: $ARGUMENTS

  Formalize this into a testable claim:

  1. **State the claim precisely**
     - Observable: What would you see if true?
     - Measurable: How would you quantify it?
     - Time-bound: Over what period?

  2. **Identify implicit assumptions**
     - What must be true for this hypothesis to be meaningful?
     - What are you taking for granted?

  3. **State the null hypothesis**
     - What is the default assumption if your claim is wrong?
     - This is what you're trying to reject.

  Return:
  {
    "claim": "Precise statement",
    "observable_prediction": "What you'd see if true",
    "measurement_method": "How to quantify",
    "implicit_assumptions": [...],
    "null_hypothesis": "Default if claim is false"
  }
```

---

## Phase 2: Hypothesis Decomposition (Parallel Pool)

```yaml
Task: Sub-Hypothesis Extractor
Subagent: general-purpose
Prompt: |
  Formalized claim: {claim}

  Break into testable sub-hypotheses:

  Complex claims often hide multiple sub-claims. Extract them:

  For each sub-hypothesis:
  {
    "id": "H1.1",
    "statement": "...",
    "parent": "H1",
    "testable_independently": true/false,
    "priority": "blocking / important / supporting"
  }

  Identify:
  - Which sub-hypotheses MUST be true for the main claim?
  - Which are nice-to-have supporting evidence?
  - What's the minimum set that would validate the claim?
```

```yaml
Task: Dependency Mapper
Subagent: general-purpose
Prompt: |
  Sub-hypotheses: {sub_hypotheses}

  Map dependencies:

  - If H1.1 fails, does H1.2 become meaningless?
  - Which must be tested first?
  - Are any redundant (same test answers both)?

  Return dependency DAG:
  {
    "nodes": [...],
    "edges": [{"from": "H1.1", "to": "H1.3", "type": "requires"}],
    "critical_path": ["H1.1", "H1.2", "H1.5"],
    "test_order": [...]
  }
```

```yaml
Task: Assumption Surfacer
Subagent: general-purpose
Prompt: |
  Formalized claim: {claim}
  Sub-hypotheses: {sub_hypotheses}

  Surface hidden assumptions:

  For each assumption:
  {
    "assumption": "What's being assumed",
    "if_violated": "What happens to the hypothesis",
    "testable": true/false,
    "risk_level": "low / medium / high"
  }

  Be adversarial. Find the assumptions the researcher doesn't realize they're making.
```

---

## Phase 3: Falsification Design (Parallel Pool)

```yaml
Task: Kill Criteria Designer
Subagent: general-purpose
Prompt: |
  Hypothesis tree: {hypothesis_tree}

  For EACH sub-hypothesis, define kill criteria:

  {
    "hypothesis_id": "H1.1",
    "falsification_observation": "What you'd see if FALSE",
    "threshold": "Specific value that disproves (e.g., p > 0.05, effect < 0.1)",
    "severity": "STOP_ALL / REVISE / NOTE_AND_CONTINUE",
    "minimum_n": "Sample size needed for confident rejection"
  }

  Be precise. Vague kill criteria allow motivated reasoning.

  For the main hypothesis, define the "Toronto threshold":
  - Below what confidence should you print "?????" and not assert?
```

```yaml
Task: Control Condition Builder
Subagent: general-purpose
Prompt: |
  Hypothesis: {formalized_claim}

  Design control conditions:

  1. **Baseline control** (no intervention)
     - What does the system do without your manipulation?

  2. **Negative control** (inverted hypothesis)
     - If you do the opposite, do you get the opposite result?

  3. **Random/noise control**
     - If you randomize the key variable, does the effect disappear?

  4. **Alternative explanation control**
     - For each confound, a condition that tests if IT explains the result

  Return structured control condition specifications.
```

```yaml
Task: Confound Identifier
Subagent: general-purpose
Prompt: |
  Hypothesis: {formalized_claim}
  Experimental setup: {controls}

  Identify confounds:

  A confound is any variable OTHER than your hypothesis that could explain results.

  For each confound:
  {
    "variable": "What it is",
    "how_it_confounds": "How it could produce false positive",
    "isolation_strategy": "How to control for it",
    "residual_risk": "Risk even after control"
  }

  Be paranoid. What would a hostile reviewer point to?
```

---

## Phase 4: Evidence Integration Plan (Sequential)

```yaml
Task: Integration Planner
Subagent: general-purpose
Prompt: |
  All sub-hypotheses: {hypothesis_tree}
  Kill criteria: {kill_criteria}
  Controls: {control_conditions}

  Design evidence integration:

  1. **How will multiple evidence sources combine?**
     - If 3/5 sub-hypotheses pass, what's the verdict?
     - Apply Watson principle: more independent sources agreeing = higher confidence

  2. **Confidence thresholds for conclusions**
     - What overall confidence required to claim success?
     - What confidence for "promising but more work needed"?
     - What confidence for "hypothesis rejected"?

  3. **Pre-registration**
     - Lock in analysis plan BEFORE seeing results
     - Prevent p-hacking and post-hoc rationalization

  Return integration specification.
```

---

## Phase 5: Protocol Document (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI HYPOTHESIS PROTOCOL
════════════════════════════════════════════════════════════════════════════════

📋 HYPOTHESIS
────────────────────────────────────────────────────────────────────────────────
{formalized_claim}

NULL HYPOTHESIS: {null_hypothesis}

────────────────────────────────────────────────────────────────────────────────
                         SUB-HYPOTHESES
────────────────────────────────────────────────────────────────────────────────

ID        Statement                           Priority      Kill Threshold
──────────────────────────────────────────────────────────────────────────
H1.1      [statement]                         [priority]    [threshold]
H1.2      [statement]                         [priority]    [threshold]
...

Test Order: H1.1 → H1.2 → H1.3 (based on dependencies)

────────────────────────────────────────────────────────────────────────────────
                         CONTROL CONDITIONS
────────────────────────────────────────────────────────────────────────────────

Condition          Purpose                    Expected Result
──────────────────────────────────────────────────────────────────────────
Baseline           No intervention            [expected]
Negative           Inverted manipulation      [expected]
Random             Noise control              [expected]
[Confound X]       Rule out alternative       [expected]

────────────────────────────────────────────────────────────────────────────────
                         KILL CRITERIA
────────────────────────────────────────────────────────────────────────────────

⚠️  IF [observation], THEN [action]:

• If p > 0.05 on H1.1: STOP - main hypothesis unsupported
• If effect size < 0.2: REVISE - effect too small for practical significance
• If random control matches treatment: STOP - no real effect

TORONTO THRESHOLD: Below [X]% confidence, report "?????" not assertion.

────────────────────────────────────────────────────────────────────────────────
                         HIDDEN ASSUMPTIONS
────────────────────────────────────────────────────────────────────────────────

⚠️  This experiment assumes:
• [assumption 1] - If violated: [consequence]
• [assumption 2] - If violated: [consequence]

────────────────────────────────────────────────────────────────────────────────
                         EVIDENCE INTEGRATION
────────────────────────────────────────────────────────────────────────────────

SUCCESS requires: [criteria]
PARTIAL requires: [criteria]
FAILURE indicated by: [criteria]

Pre-registered analysis plan: [locked before data collection]

════════════════════════════════════════════════════════════════════════════════
       "Design the experiment to prove yourself wrong. If it survives,
        your confidence is earned."
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - The hypothesis to formalize and design falsification tests for

---

## Ferrucci Principles Embodied

1. **Falsification-first**: Design to disprove, not confirm
2. **Explicit kill criteria**: Know exactly what would stop you
3. **Multiple independent tests**: Consensus across sub-hypotheses
4. **Assumption surfacing**: Make hidden beliefs explicit
5. **Pre-registration**: Lock analysis before seeing data
6. **The Toronto threshold**: Know when to print "?????"
