Adversarial stress-testing to find breaking points.

Usage: Push claims, systems, or findings until they fail. Document exactly where and how.

**The Stress Principle:** Watson was tested against the best Jeopardy! players in the world. Not friendly benchmarks. Adversarial conditions reveal true capability.

---

## Why Stress Testing Matters

Systems fail at the edges. Claims fail under scrutiny. The question isn't whether something works - it's where it stops working.

Stress testing finds:
- Edge cases that break assumptions
- Adversarial inputs that exploit weaknesses
- Load conditions that expose limits
- Distribution shifts that reveal overfitting

This command structures adversarial validation.

---

## Execution Architecture

```
Phase 1:  Target Specification ─────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Attack Surface Mapping ──┬── Edge Case Finder ─────────────┐
          (Parallel)               ├── Adversarial Input Designer ───┼──→ Attack Plan
                                   ├── Distribution Shift Designer ──┤
                                   └── Load/Scale Tester ────────────┘
    │
    ▼
Phase 3:  Stress Execution ────────┬── Edge Case Battery ────────────┐
          (Parallel)               ├── Adversarial Battery ──────────┼──→ Failure Map
                                   ├── Distribution Battery ─────────┤
                                   └── Scale Battery ────────────────┘
    │
    ▼
Phase 4:  Breaking Point Analysis ──────────────────────────────────── [Sequential]
```

---

## Phase 1: Target Specification (Sequential)

```yaml
Task: Target Analyzer
Subagent: general-purpose
Prompt: |
  Target: $ARGUMENTS

  Specify the stress test target:

  {
    "target": "precise description",
    "type": "claim / system / model / method",
    "stated_capability": "what it claims to do",
    "operating_envelope": "normal conditions it's designed for",
    "expected_weak_points": "hypothesized vulnerabilities",
    "critical_failure_threshold": "what constitutes unacceptable failure"
  }

  Be honest about what success means. Moving goalposts isn't stress testing.
```

---

## Phase 2: Attack Surface Mapping (Parallel Pool)

```yaml
Task: Edge Case Finder
Subagent: general-purpose
Prompt: |
  Target: {target_spec}

  Identify edge cases:

  {
    "edge_cases": [
      {
        "id": "EC1",
        "description": "The edge condition",
        "why_edge": "What makes this non-typical",
        "expected_behavior": "What should happen",
        "predicted_failure_mode": "How it might break",
        "severity_if_fails": "low / medium / high / critical"
      }
    ],
    "coverage": "aspects of the system these edge cases probe"
  }

  Think like a hostile user. What would break this?
```

```yaml
Task: Adversarial Input Designer
Subagent: general-purpose
Prompt: |
  Target: {target_spec}

  Design adversarial inputs:

  {
    "adversarial_inputs": [
      {
        "id": "ADV1",
        "input": "The adversarial input",
        "attack_vector": "What weakness this exploits",
        "expected_vulnerability": "What should fail",
        "detection_difficulty": "How hard to catch this",
        "real_world_likelihood": "Could this happen naturally?"
      }
    ]
  }

  Include:
  - Malformed inputs
  - Boundary values
  - Contradictory instructions
  - Inputs designed to confuse
  - Real-world messy data
```

```yaml
Task: Distribution Shift Designer
Subagent: general-purpose
Prompt: |
  Target: {target_spec}

  Design distribution shift tests:

  {
    "distribution_shifts": [
      {
        "id": "DS1",
        "shift_type": "temporal / demographic / domain / language / format",
        "training_distribution": "What it was trained/designed for",
        "test_distribution": "What we're testing",
        "expected_degradation": "How much performance loss",
        "acceptable_threshold": "How much degradation is OK"
      }
    ]
  }

  Test: Does this work ONLY on data similar to training, or does it generalize?
```

```yaml
Task: Load Scale Tester
Subagent: general-purpose
Prompt: |
  Target: {target_spec}

  Design load and scale tests:

  {
    "scale_tests": [
      {
        "id": "SC1",
        "dimension": "What we're scaling (size, speed, volume)",
        "normal_operating_point": "Typical usage",
        "stress_point": "Elevated usage",
        "breaking_point_hypothesis": "Where we expect failure",
        "failure_mode_hypothesis": "How we expect it to fail"
      }
    ]
  }

  Find the limits. Every system has them.
```

---

## Phase 3: Stress Execution (Parallel Batteries)

```yaml
Task: Edge Case Battery
Subagent: general-purpose
Prompt: |
  Edge cases: {edge_cases}
  Target: {target_spec}

  Execute edge case battery:

  For each edge case:
  {
    "edge_case_id": "EC1",
    "result": "pass / fail / partial",
    "actual_behavior": "What happened",
    "expected_behavior": "What should have happened",
    "failure_severity": "if failed, how bad",
    "reproducible": true/false,
    "notes": "observations"
  }

  Document failures precisely. Reproducibility matters.
```

```yaml
Task: Adversarial Battery
Subagent: general-purpose
Prompt: |
  Adversarial inputs: {adversarial_inputs}
  Target: {target_spec}

  Execute adversarial battery:

  For each adversarial input:
  {
    "adversarial_id": "ADV1",
    "result": "defended / exploited / partial",
    "actual_response": "What the system did",
    "exploitation_severity": "if exploited, impact",
    "fix_difficulty": "trivial / moderate / difficult / architectural",
    "notes": "observations"
  }
```

```yaml
Task: Distribution Battery
Subagent: general-purpose
Prompt: |
  Distribution shifts: {distribution_shifts}
  Target: {target_spec}

  Execute distribution shift battery:

  For each shift:
  {
    "shift_id": "DS1",
    "baseline_performance": "on training distribution",
    "shifted_performance": "on test distribution",
    "degradation": "% change",
    "acceptable": true/false,
    "failure_pattern": "what specifically fails",
    "notes": "observations"
  }
```

```yaml
Task: Scale Battery
Subagent: general-purpose
Prompt: |
  Scale tests: {scale_tests}
  Target: {target_spec}

  Execute scale battery:

  For each scale test:
  {
    "scale_id": "SC1",
    "tested_points": [
      {"level": "1x", "performance": "...", "stable": true},
      {"level": "2x", "performance": "...", "stable": true},
      {"level": "5x", "performance": "...", "stable": false}
    ],
    "breaking_point": "where it failed",
    "failure_mode": "how it failed",
    "graceful_degradation": true/false
  }
```

---

## Phase 4: Breaking Point Analysis (Sequential)

```
================================================================================
                    FERRUCCI STRESS TEST REPORT
================================================================================

TARGET: {target_description}

--------------------------------------------------------------------------------
                         STRESS TEST SUMMARY
--------------------------------------------------------------------------------

                          Tests    Passed    Failed    Pass Rate
Edge Cases                [N]      [n]       [n]       [X%]
Adversarial Inputs        [N]      [n]       [n]       [X%]
Distribution Shifts       [N]      [n]       [n]       [X%]
Scale Tests               [N]      [n]       [n]       [X%]

OVERALL ROBUSTNESS SCORE: [X/10]

--------------------------------------------------------------------------------
                         BREAKING POINTS IDENTIFIED
--------------------------------------------------------------------------------

CRITICAL FAILURES (must fix):

  [BP1] Edge case EC3: [description]
    • Triggered by: [input/condition]
    • Failure mode: [what happened]
    • Impact: [consequence]
    • Fix complexity: [estimate]

  [BP2] Adversarial ADV2: [description]
    • Attack vector: [how exploited]
    • Exploitation: [what attacker gains]
    • Defense recommendation: [how to fix]

MODERATE FAILURES (should fix):

  [BP3] Distribution shift DS1: [description]
    • Performance drop: [X%]
    • Acceptable threshold: [Y%]
    • Mitigation: [options]

MINOR FAILURES (nice to fix):

  [BP4] Scale test SC2: [description]
    • Breaking point: [where]
    • Workaround: [available?]

--------------------------------------------------------------------------------
                         FAILURE MODE TAXONOMY
--------------------------------------------------------------------------------

GRACEFUL DEGRADATION (good):
  • [System slowed but continued working at SC1]
  • [Gave uncertain answer instead of wrong answer at DS2]

SILENT FAILURE (dangerous):
  • [Produced confident wrong answer at ADV1]
  • [No indication of out-of-distribution at DS3]

CATASTROPHIC FAILURE (critical):
  • [Complete crash at EC5]
  • [Infinite loop at ADV4]

--------------------------------------------------------------------------------
                         ROBUSTNESS ENVELOPE
--------------------------------------------------------------------------------

VERIFIED SAFE OPERATING CONDITIONS:
  • Input types: [what works]
  • Scale: [up to X]
  • Distribution: [similar to training]

VERIFIED UNSAFE CONDITIONS:
  • [Condition 1]: fails with [consequence]
  • [Condition 2]: fails with [consequence]

UNKNOWN (not tested):
  • [Conditions that weren't stress tested]

--------------------------------------------------------------------------------
                         RECOMMENDATIONS
--------------------------------------------------------------------------------

IMMEDIATE (blocking deployment):
  □ Fix [BP1] - critical vulnerability
  □ Fix [BP2] - exploitable weakness

SHORT-TERM (before scaling):
  □ Address [BP3] - distribution shift handling
  □ Add graceful degradation for [SC2]

LONG-TERM (architectural):
  □ Redesign [component] to handle [edge case class]
  □ Add monitoring for [failure mode]

--------------------------------------------------------------------------------
                      CONFIDENCE CALIBRATION
--------------------------------------------------------------------------------

Based on stress testing:

STATED CAPABILITY: "{original_claim}"

ACTUAL CAPABILITY: "{revised_claim_based_on_stress_testing}"

CONFIDENCE ADJUSTMENT:
  • Before stress testing: [stated confidence]
  • After stress testing: [calibrated confidence]
  • Reason: [what we learned]

================================================================================
      "Stress testing isn't about making yourself feel good.
       It's about finding out how your system actually fails."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - System, claim, or finding to stress test

---

## Ferrucci Principles Embodied

1. **Adversarial validation**: Test against hostile conditions, not friendly ones
2. **Failure mode mapping**: Know HOW things fail, not just IF
3. **Honest capability claims**: Adjust claims based on evidence
4. **Operational envelope**: Know where the system works and where it doesn't
5. **Graceful degradation**: Failing safely is better than failing silently

