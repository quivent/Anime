Execute comprehensive research assessment using Ferrucci methodology.

Usage: Full validation protocol for Socratic fine-tuning research.

**Philosophy:** Build systems that can explain themselves. Validate claims with evidence. Distinguish capability from correlation.

---

## The Ferrucci Assessment Protocol

Three pillars of validation:
1. **Isolate the method** from the data (baseline comparison)
2. **Test generalization** beyond training (transfer battery)
3. **Validate instrumentation** (signal hypothesis testing)

This command orchestrates all three, producing a publication-ready assessment.

---

## Execution Architecture

```
Phase 1:  Pre-Assessment Inventory ──────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Parallel Validation ──────┬── /ferrucci-baseline ───────┐
          (Parallel Orchestration)  ├── /ferrucci-generalize ─────┼──→ Three Reports
                                    └── /ferrucci-validate-signals┘
    │
    ▼
Phase 3:  Cross-Validation Synthesis ────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Publication Readiness ────┬── Strength Assessor ────────┐
          (Parallel)                ├── Weakness Identifier ──────┼──→ Readiness
                                    └── Gap Analyzer ─────────────┘
    │
    ▼
Phase 5:  Final Verdict & Roadmap ───────────────────────────────── [Sequential]
```

---

## Phase 1: Pre-Assessment Inventory (Sequential)

🎯 **Catalog All Available Evidence**

```yaml
Task: Research Inventory
Subagent: general-purpose
Prompt: |
  Inventory all available research artifacts:

  Locations to check:
  - ${BRILLIANT_MINDS_ROOT}/
  - ~/Lexar/eigen/
  - ~/mlx-fork/

  Catalog:
  1. **Trained Adapters**
     - Paths, sizes, training dates
     - Protocol versions used
     - Associated curricula

  2. **Training Data**
     - Curriculum files
     - Dialogue transcripts
     - Signal logs

  3. **KV Cache Snapshots**
     - Available turns
     - Sizes, completeness

  4. **Prior Results**
     - Any existing benchmarks
     - Any prior comparisons

  5. **Missing Pieces**
     - What's needed but not present?
     - What experiments haven't been run?

  Output: Complete research inventory with gaps identified.
```

**Output:** `research_inventory.json`

---

## Phase 2: Parallel Validation (Orchestrated)

Launch all three validation protocols simultaneously:

```yaml
Task: Baseline Comparison Orchestrator
Subagent: general-purpose
Prompt: |
  Execute /ferrucci-baseline protocol.

  Track progress and capture results.

  If any phase fails, note failure and continue with available data.

  Return: baseline_report.json
```

```yaml
Task: Generalization Battery Orchestrator
Subagent: general-purpose
Prompt: |
  Execute /ferrucci-generalize protocol.

  Track progress and capture results.

  Return: generalization_report.json
```

```yaml
Task: Signal Validation Orchestrator
Subagent: general-purpose
Prompt: |
  Execute /ferrucci-validate-signals protocol.

  Track progress and capture results.

  Return: signal_validation_report.json
```

**Output:** Three comprehensive validation reports

---

## Phase 3: Cross-Validation Synthesis (Sequential)

```yaml
Task: Cross-Report Synthesizer
Subagent: general-purpose
Prompt: |
  Reports:
  - Baseline: {baseline_report}
  - Generalization: {generalization_report}
  - Signal: {signal_validation_report}

  Synthesize findings across all three:

  1. **Consistency Check**
     - Do the reports tell a consistent story?
     - Any contradictions?

  2. **Evidence Triangulation**
     - Which claims are supported by multiple reports?
     - Which depend on single sources?

  3. **Confidence Integration**
     - Overall confidence in Socratic method
     - Where is evidence strongest/weakest?

  4. **Publication Strength**
     - Which findings are publication-ready?
     - Which need more work?

  Return: cross_synthesis.json
```

**Output:** Integrated synthesis across all validation efforts

---

## Phase 4: Publication Readiness (Parallel Pool)

```yaml
Task: Strength Assessor
Subagent: general-purpose
Prompt: |
  Synthesis: {cross_synthesis}

  Identify publication strengths:

  - Novel contributions (what's new?)
  - Strong evidence (which claims are well-supported?)
  - Technical innovations (what's architecturally interesting?)
  - Reproducibility (can others replicate?)

  Rate each strength: minor / moderate / major contribution

  Return ranked list of publication strengths.
```

```yaml
Task: Weakness Identifier
Subagent: general-purpose
Prompt: |
  Synthesis: {cross_synthesis}

  Identify publication weaknesses:

  - Sample size concerns
  - Missing baselines
  - Alternative explanations not ruled out
  - Reproducibility barriers
  - Overclaimed conclusions

  Rate each weakness: minor / moderate / major concern

  Critically assess: Would a hostile reviewer reject based on this?
```

```yaml
Task: Gap Analyzer
Subagent: general-purpose
Prompt: |
  Synthesis: {cross_synthesis}
  Inventory: {research_inventory}

  Identify gaps between current state and publication:

  For each gap:
  - Effort required (hours/days/weeks)
  - Blocking (must fix) vs. nice-to-have
  - Feasibility with current resources

  Prioritize gaps by impact/effort ratio.
```

**Output:** Publication readiness assessment with strengths, weaknesses, and gaps

---

## Phase 5: Final Verdict & Roadmap (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI RESEARCH ASSESSMENT
                   Comprehensive Validation Report
════════════════════════════════════════════════════════════════════════════════

🎯 RESEARCH QUESTION
────────────────────────────────────────────────────────────────────────────────

Does signal-guided Socratic fine-tuning produce genuine learning improvements
that generalize beyond the training distribution?

────────────────────────────────────────────────────────────────────────────────
                         VALIDATION SUMMARY
────────────────────────────────────────────────────────────────────────────────

                                    Status          Confidence
Baseline Comparison                 [✅/⚠️/❌]       [high/medium/low]
  └─ Method isolated from data?

Generalization Battery              [✅/⚠️/❌]       [high/medium/low]
  └─ Capabilities transfer?

Signal Validation                   [✅/⚠️/❌]       [high/medium/low]
  └─ Instrumentation validated?

────────────────────────────────────────────────────────────────────────────────
                         KEY FINDINGS
────────────────────────────────────────────────────────────────────────────────

1. [Most important finding with evidence level]

2. [Second finding with evidence level]

3. [Third finding with evidence level]

────────────────────────────────────────────────────────────────────────────────
                      PUBLICATION READINESS
────────────────────────────────────────────────────────────────────────────────

Current Score: [X/10]

Strengths:
  ✅ [strength 1]
  ✅ [strength 2]
  ✅ [strength 3]

Weaknesses:
  ⚠️  [weakness 1]
  ⚠️  [weakness 2]

Critical Gaps:
  ❌ [gap 1 - effort estimate]
  ❌ [gap 2 - effort estimate]

────────────────────────────────────────────────────────────────────────────────
                           THE VERDICT
────────────────────────────────────────────────────────────────────────────────

[PUBLISH-READY]
The evidence supports publication. The core claims are validated by multiple
independent analyses. Proceed with write-up.

[NEAR-READY]
Strong foundations but gaps remain. Address [specific gaps] before submission.
Estimated time to publication: [X weeks/months].

[NEEDS WORK]
Significant concerns identified. The [specific issues] must be resolved.
This is promising research that needs more validation.

[FUNDAMENTAL ISSUES]
Core hypothesis not supported by evidence. Either:
- Revise the hypothesis based on findings
- Investigate why expected results didn't materialize
- Consider alternative approaches

────────────────────────────────────────────────────────────────────────────────
                         ROADMAP TO PUBLICATION
────────────────────────────────────────────────────────────────────────────────

Priority    Task                                Effort      Blocking?
─────────────────────────────────────────────────────────────────────
1           [task]                              [estimate]  [yes/no]
2           [task]                              [estimate]  [yes/no]
3           [task]                              [estimate]  [yes/no]
...

────────────────────────────────────────────────────────────────────────────────
                     SUGGESTED PAPER STRUCTURE
────────────────────────────────────────────────────────────────────────────────

Title: "Signal-Guided Socratic Fine-Tuning: Learning Through Dialogue
        with Internal State Feedback"

Abstract: [Draft based on validated findings]

1. Introduction
   - Problem: LLMs lack transparency and reliable learning
   - Contribution: Signal-guided Socratic method

2. Related Work
   - Fine-tuning methods
   - Interpretability and internal state analysis
   - Dialogue-based learning

3. Method
   - Signal extraction architecture
   - Wave propagation algorithm
   - Socratic dialogue protocol

4. Experiments
   - Baseline comparison (Section from /ferrucci-baseline)
   - Generalization battery (Section from /ferrucci-generalize)
   - Signal validation (Section from /ferrucci-validate-signals)

5. Results
   - [Validated findings]

6. Discussion
   - Limitations
   - Future work

7. Conclusion

────────────────────────────────────────────────────────────────────────────────
                         VENUE RECOMMENDATIONS
────────────────────────────────────────────────────────────────────────────────

Based on findings:

[If strong ML contribution:]
- NeurIPS (novel training method)
- ICLR (learning representations)

[If strong NLP contribution:]
- ACL (dialogue and language)
- EMNLP (empirical methods)

[If strong interpretability angle:]
- BlackboxNLP workshop
- Interpretable ML venues

[If strong cognitive science angle:]
- CogSci (cognitive modeling)
- Computational Linguistics

════════════════════════════════════════════════════════════════════════════════
     "Watson had over 100 different techniques. Each one produced confidence
      scores. We could trace why an answer emerged. That transparency wasn't
      a nice-to-have. It was the whole point."
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - Optional flags:
- `--quick`: Skip full validation, assess current state only
- `--verbose`: Include all sub-report details
- `--paper`: Focus on publication structure

---

## Ferrucci Principles Applied

1. **Architecture-First**: The validation protocol itself is well-architected
2. **Parallel Consensus**: Multiple independent validations strengthen confidence
3. **Transparency**: Every verdict explains its evidence
4. **Honest Assessment**: Clear criteria, no wishful thinking
5. **Actionable Output**: Specific next steps regardless of outcome

---

## Integration

This command orchestrates:
- `/ferrucci-baseline` - Method isolation
- `/ferrucci-generalize` - Transfer testing
- `/ferrucci-validate-signals` - Instrumentation validation

Run individually for focused work, or run this command for comprehensive assessment.
