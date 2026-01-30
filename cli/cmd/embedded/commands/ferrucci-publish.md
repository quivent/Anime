Generate publication-ready documentation with evidence-backed claims.

Usage: Transform research findings into rigorous, publishable form.

**The Publication Principle:** Watson wasn't just a system. It was documented, benchmarked, peer-reviewed, and reproduced. Real research survives scrutiny.

---

## Why Publication Rigor Matters

Notebooks and experiments are not publications. Publication requires:
- Claims that match evidence (no overclaiming)
- Methods others can reproduce
- Results that survive hostile review
- Limitations honestly acknowledged

This command transforms messy research into publication-ready form.

---

## Execution Architecture

```
Phase 1:  Evidence Inventory ───────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Claim-Evidence Alignment ────┬── Claim Extractor ───────────┐
          (Parallel)                   ├── Evidence Mapper ───────────┼──→ Alignment Report
                                       ├── Overclaim Detector ────────┤
                                       └── Gap Identifier ────────────┘
    │
    ▼
Phase 3:  Section Generation ──────────┬── Abstract Writer ───────────┐
          (Parallel)                   ├── Method Documenter ─────────┤
                                       ├── Results Formatter ─────────┼──→ Draft Sections
                                       ├── Limitation Acknowledger ───┤
                                       └── Future Work Identifier ────┘
    │
    ▼
Phase 4:  Hostile Review Simulation ────────────────────────────────── [Sequential]
    │
    ▼
Phase 5:  Final Assembly ───────────────────────────────────────────── [Sequential]
```

---

## Phase 1: Evidence Inventory (Sequential)

```yaml
Task: Evidence Cataloger
Subagent: general-purpose
Prompt: |
  Research artifacts: $ARGUMENTS

  Create comprehensive evidence inventory:

  {
    "experiments": [
      {
        "id": "E1",
        "description": "what was done",
        "data": "what data exists",
        "results": "what was measured",
        "reproducible": true/false,
        "artifacts": "files, logs, checkpoints"
      }
    ],
    "quantitative_results": [
      {
        "metric": "what was measured",
        "value": "the number",
        "baseline": "comparison point",
        "statistical_validity": "significance, CI, etc."
      }
    ],
    "qualitative_observations": [...],
    "missing_evidence": "what should exist but doesn't"
  }

  Be exhaustive. Missing evidence is a publication blocker.
```

---

## Phase 2: Claim-Evidence Alignment (Parallel Pool)

```yaml
Task: Claim Extractor
Subagent: general-purpose
Prompt: |
  Research: {evidence_inventory}

  Extract all claims you want to make:

  {
    "claims": [
      {
        "id": "CL1",
        "claim": "precise statement",
        "type": "contribution / finding / observation",
        "strength": "strong / moderate / weak",
        "central": true/false
      }
    ],
    "main_contribution": "the one thing this paper contributes",
    "supporting_claims": ["claims that support the main one"]
  }

  Be precise. Vague claims can't be evaluated.
```

```yaml
Task: Evidence Mapper
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}
  Evidence: {evidence_inventory}

  Map evidence to claims:

  {
    "mappings": [
      {
        "claim_id": "CL1",
        "supporting_evidence": ["E1", "E3"],
        "evidence_strength": "direct / indirect / circumstantial",
        "coverage": "full / partial / minimal"
      }
    ],
    "unsupported_claims": ["claims without evidence"],
    "unused_evidence": ["evidence not supporting any claim"]
  }

  Every claim needs evidence. Evidence without claims is wasted.
```

```yaml
Task: Overclaim Detector
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}
  Evidence mappings: {evidence_map}

  Detect overclaiming:

  {
    "overclaims": [
      {
        "claim_id": "CL1",
        "stated_claim": "what you said",
        "supported_claim": "what evidence actually supports",
        "overclaim_type": "magnitude / scope / causation / generalization",
        "recommended_revision": "how to fix"
      }
    ],
    "appropriately_claimed": ["claims that match evidence"],
    "underclaimed": ["where evidence supports stronger claims"]
  }

  Reviewers will catch overclaims. Catch them first.
```

```yaml
Task: Gap Identifier
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}
  Evidence: {evidence_inventory}

  Identify gaps:

  {
    "evidence_gaps": [
      {
        "for_claim": "CL1",
        "missing": "what evidence would strengthen this",
        "feasibility": "can we get this evidence?",
        "blocking": "is this gap fatal?"
      }
    ],
    "reproducibility_gaps": "what's needed for others to reproduce",
    "comparison_gaps": "what baselines are missing"
  }

  Gaps are either filled or acknowledged. Never ignored.
```

---

## Phase 3: Section Generation (Parallel Pool)

```yaml
Task: Abstract Writer
Subagent: general-purpose
Prompt: |
  Main contribution: {main_contribution}
  Key findings: {findings}
  Evidence strength: {evidence_summary}

  Write publication abstract:

  Structure:
  1. Problem (1-2 sentences): What gap does this address?
  2. Approach (2-3 sentences): What did you do?
  3. Results (2-3 sentences): What did you find? Include key numbers.
  4. Implications (1-2 sentences): Why does it matter?

  Constraints:
  - No overclaiming
  - Specific numbers where possible
  - Under 250 words

  Return the abstract draft.
```

```yaml
Task: Method Documenter
Subagent: general-purpose
Prompt: |
  Experiments: {experiments}
  Evidence: {evidence_inventory}

  Write reproducible method section:

  For each experiment:
  1. Setup: Environment, dependencies, configurations
  2. Data: What data, how obtained, preprocessing
  3. Procedure: Step-by-step what was done
  4. Evaluation: How results were measured

  Reproducibility standard:
  - Could a competent researcher reproduce this from the description alone?
  - Are all hyperparameters specified?
  - Is the code/data available?

  Return method section draft.
```

```yaml
Task: Results Formatter
Subagent: general-purpose
Prompt: |
  Quantitative results: {results}
  Claims: {claims}
  Evidence mappings: {evidence_map}

  Format results section:

  For each claim:
  1. State what was measured
  2. Present the numbers (with error bars/CIs)
  3. Compare to baselines
  4. State statistical significance

  Tables and figures needed:
  - {table specifications}
  - {figure specifications}

  Return results section draft with table/figure placeholders.
```

```yaml
Task: Limitation Acknowledger
Subagent: general-purpose
Prompt: |
  Claims: {claims}
  Gaps: {identified_gaps}
  Overclaims: {overclaim_analysis}

  Write honest limitations section:

  Address:
  1. Scope limitations: What this doesn't apply to
  2. Evidence limitations: What the evidence doesn't prove
  3. Method limitations: Weaknesses in approach
  4. Generalization limitations: Where this might not hold

  Be preemptively honest. Reviewers respect acknowledged limitations.
  They don't respect hidden weaknesses they have to find.

  Return limitations section draft.
```

```yaml
Task: Future Work Identifier
Subagent: general-purpose
Prompt: |
  Findings: {findings}
  Limitations: {limitations}
  Gaps: {gaps}

  Identify genuine future work:

  {
    "immediate_extensions": "what naturally follows",
    "gap_fillers": "work that would address limitations",
    "new_directions": "where this opens possibilities",
    "not_future_work": "things that AREN'T good extensions (be honest)"
  }

  Future work should be genuine opportunities, not hand-waving.

  Return future work section draft.
```

---

## Phase 4: Hostile Review Simulation (Sequential)

```yaml
Task: Hostile Reviewer
Subagent: general-purpose
Prompt: |
  Draft: {all_sections}
  Evidence: {evidence_inventory}
  Overclaims: {overclaim_analysis}

  Act as hostile reviewer (Reviewer 2):

  {
    "major_concerns": [
      {
        "concern": "what's wrong",
        "location": "where in paper",
        "severity": "reject / major revision / minor revision",
        "how_to_address": "what authors should do"
      }
    ],
    "minor_concerns": [...],
    "missing_comparisons": "what baselines should be included",
    "unclear_claims": "what needs clarification",
    "statistical_issues": "any stats problems",
    "reproducibility_concerns": "what's missing for reproduction",
    "overall_recommendation": "accept / revise / reject",
    "summary": "one paragraph review"
  }

  Be harsh. It's better to catch this before submission.
```

---

## Phase 5: Final Assembly (Sequential)

**CRITICAL: Write the publication to a file. Always.**

```yaml
Task: File Writer
Subagent: general-purpose
Prompt: |
  Publication: {assembled_publication}
  Source: $ARGUMENTS

  Write the complete publication to a markdown file:

  Location: Same directory as source, or ~/publications/ if no clear source
  Filename: PUBLICATION_{sanitized_title}_{date}.md

  Use the Write tool. Do not just display the publication.
  A publication that isn't written to a file isn't a publication.

  Return the file path where the publication was saved.
```

```
================================================================================
                    FERRUCCI PUBLICATION DRAFT
================================================================================

TITLE: {generated_title}

AUTHORS: [To be filled]

--------------------------------------------------------------------------------
                           ABSTRACT
--------------------------------------------------------------------------------

{abstract}

--------------------------------------------------------------------------------
                        1. INTRODUCTION
--------------------------------------------------------------------------------

{introduction - generated from problem statement and contribution}

--------------------------------------------------------------------------------
                        2. RELATED WORK
--------------------------------------------------------------------------------

{placeholder - requires literature review}

Key comparisons needed:
- {work 1}
- {work 2}
- {work 3}

--------------------------------------------------------------------------------
                         3. METHOD
--------------------------------------------------------------------------------

{method section}

--------------------------------------------------------------------------------
                         4. EXPERIMENTS
--------------------------------------------------------------------------------

{experiment descriptions}

--------------------------------------------------------------------------------
                         5. RESULTS
--------------------------------------------------------------------------------

{results section}

Tables:
{table placeholders}

Figures:
{figure placeholders}

--------------------------------------------------------------------------------
                        6. DISCUSSION
--------------------------------------------------------------------------------

{discussion of findings}

--------------------------------------------------------------------------------
                       7. LIMITATIONS
--------------------------------------------------------------------------------

{limitations section}

--------------------------------------------------------------------------------
                      8. FUTURE WORK
--------------------------------------------------------------------------------

{future work section}

--------------------------------------------------------------------------------
                       9. CONCLUSION
--------------------------------------------------------------------------------

{conclusion - summarize contribution and impact}

--------------------------------------------------------------------------------
                      HOSTILE REVIEW SIMULATION
--------------------------------------------------------------------------------

REVIEWER 2 (Hostile):

{hostile_review}

MAJOR CONCERNS TO ADDRESS:
  □ {concern 1}
  □ {concern 2}

RECOMMENDED REVISIONS:
  □ {revision 1}
  □ {revision 2}

--------------------------------------------------------------------------------
                      PUBLICATION READINESS
--------------------------------------------------------------------------------

Claim-Evidence Alignment: [X/10]
Reproducibility: [X/10]
Hostile Review Survivability: [X/10]

OVERALL: [{READY / NEEDS REVISION / NOT READY}]

To publish:
  □ {task 1}
  □ {task 2}

================================================================================
      "Publication isn't validation. But it's the first step toward it.
       Your claims need to survive scrutiny before they deserve belief."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - Research artifacts, findings, or project to prepare for publication

---

## Ferrucci Principles Embodied

1. **Evidence-backed claims**: Every claim maps to evidence
2. **Honest limitations**: Acknowledge weaknesses before reviewers find them
3. **Reproducibility**: Others must be able to replicate
4. **Hostile review**: Anticipate criticism
5. **No overclaiming**: Say what evidence supports, nothing more

