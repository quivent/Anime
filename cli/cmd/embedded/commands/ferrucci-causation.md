Distinguish correlation from causation with rigorous intervention analysis.

Usage: Audit claims for causal validity. Does X actually cause Y, or do they just co-occur?

**The Causation Principle:** Statistics tell you patterns exist. They don't tell you why. The doctors gave me population statistics. I asked about THIS patient. That's the difference between correlation and causation.

---

## Why Causation Auditing Matters

Most "findings" are correlations dressed as causation:
- "Users who do X have better outcomes" (correlation)
- "Doing X causes better outcomes" (causal claim - needs intervention evidence)

This command forces rigor:
- What's the claimed causal relationship?
- What's the actual evidence?
- Could a confound explain this?
- What intervention would test causation?

---

## Execution Architecture

```
Phase 1:  Causal Claim Extraction ──────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Evidence Analysis ───────────┬── Correlation Assessor ──────┐
          (Parallel)                   ├── Confound Finder ───────────┼──→ Evidence Map
                                       ├── Mechanism Tracer ──────────┤
                                       └── Intervention Designer ─────┘
    │
    ▼
Phase 3:  Causal Inference Audit ───────────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Verdict & Downgrade ──────────────────────────────────────── [Sequential]
```

---

## Phase 1: Causal Claim Extraction (Sequential)

```yaml
Task: Causal Claim Extractor
Subagent: general-purpose
Prompt: |
  Source: $ARGUMENTS

  Extract all causal claims (explicit or implicit):

  {
    "causal_claims": [
      {
        "id": "CC1",
        "stated_cause": "X",
        "stated_effect": "Y",
        "claim_text": "exact quote or paraphrase",
        "claim_strength": "causes / leads to / associated with / predicts",
        "implicit": true/false,
        "location": "where in document"
      }
    ]
  }

  Be aggressive. Claims like "X improves Y" imply causation even if not stated.
  "Associated with" often sneaks causal implications.
```

---

## Phase 2: Evidence Analysis (Parallel Pool)

```yaml
Task: Correlation Assessor
Subagent: general-purpose
Prompt: |
  Causal claims: {causal_claims}

  For each claim, assess the correlational evidence:

  {
    "claim_id": "CC1",
    "correlation_evidence": {
      "type": "observational / experimental / quasi-experimental",
      "sample_size": "N",
      "effect_size": "magnitude",
      "statistical_significance": "p-value or CI",
      "replication": "has this been replicated?"
    },
    "correlation_strength": "strong / moderate / weak / none",
    "correlation_limitations": "what this evidence can't show"
  }

  Correlation is necessary but not sufficient for causation.
```

```yaml
Task: Confound Finder
Subagent: general-purpose
Prompt: |
  Causal claims: {causal_claims}

  For each claim, identify potential confounds:

  {
    "claim_id": "CC1",
    "confounds": [
      {
        "variable": "Z",
        "how_it_confounds": "Z could cause both X and Y",
        "plausibility": "high / medium / low",
        "controlled_for": true/false,
        "control_method": "how it was controlled (if at all)"
      }
    ],
    "selection_bias": "could the sample be biased?",
    "reverse_causation": "could Y actually cause X?",
    "spurious_correlation": "could this be coincidence?"
  }

  Be paranoid. What would make this correlation spurious?
```

```yaml
Task: Mechanism Tracer
Subagent: general-purpose
Prompt: |
  Causal claims: {causal_claims}

  For each claim, trace the proposed mechanism:

  {
    "claim_id": "CC1",
    "proposed_mechanism": "How does X cause Y?",
    "mechanism_steps": [
      "X leads to A",
      "A leads to B",
      "B leads to Y"
    ],
    "each_step_evidenced": true/false,
    "mechanism_plausibility": "high / medium / low / unknown",
    "alternative_mechanisms": ["other ways X could lead to Y"],
    "mechanism_gaps": ["steps without evidence"]
  }

  Correlation without mechanism is suspicious.
  Mechanism without correlation is hypothesis.
  Both together strengthens causal inference.
```

```yaml
Task: Intervention Designer
Subagent: general-purpose
Prompt: |
  Causal claims: {causal_claims}

  For each claim, design an intervention test:

  {
    "claim_id": "CC1",
    "ideal_intervention": {
      "manipulation": "what we would change",
      "control": "what we would compare against",
      "predicted_effect": "what we'd see if causal",
      "predicted_null": "what we'd see if not causal"
    },
    "intervention_feasibility": "possible / difficult / impossible",
    "ethical_constraints": "any ethical issues with intervention",
    "natural_experiments": "any natural interventions we could use",
    "existing_intervention_evidence": "has anyone done this?"
  }

  The gold standard for causation is intervention.
  If you can't intervene, what's the next best evidence?
```

---

## Phase 3: Causal Inference Audit (Sequential)

```yaml
Task: Causal Inference Auditor
Subagent: general-purpose
Prompt: |
  All evidence: {correlation}, {confounds}, {mechanism}, {intervention}

  Apply causal inference criteria:

  For each claim, assess against Bradford Hill criteria:

  {
    "claim_id": "CC1",
    "bradford_hill": {
      "strength": {"score": 0-2, "evidence": "..."},
      "consistency": {"score": 0-2, "evidence": "..."},
      "specificity": {"score": 0-2, "evidence": "..."},
      "temporality": {"score": 0-2, "evidence": "..."},
      "dose_response": {"score": 0-2, "evidence": "..."},
      "plausibility": {"score": 0-2, "evidence": "..."},
      "coherence": {"score": 0-2, "evidence": "..."},
      "experiment": {"score": 0-2, "evidence": "..."},
      "analogy": {"score": 0-2, "evidence": "..."}
    },
    "total_score": "X/18",
    "causal_verdict": "established / probable / possible / unlikely / no_evidence",
    "key_weakness": "what would most strengthen the causal case"
  }

  Be rigorous. Most claims don't survive this audit.
```

---

## Phase 4: Verdict & Downgrade (Sequential)

```
================================================================================
                    FERRUCCI CAUSATION AUDIT
================================================================================

SOURCE: {document/claim analyzed}

--------------------------------------------------------------------------------
                      CAUSAL CLAIMS IDENTIFIED
--------------------------------------------------------------------------------

CC1: "[X] causes [Y]"
  Location: {where}
  Strength: {causes / leads to / associated with}

CC2: "[A] leads to [B]"
  Location: {where}
  Strength: {strength}

...

--------------------------------------------------------------------------------
                      EVIDENCE ASSESSMENT
--------------------------------------------------------------------------------

CLAIM: CC1 - "[X] causes [Y]"

CORRELATION EVIDENCE:
  Type: {observational / experimental}
  Sample: N = {size}
  Effect: {magnitude}
  Significance: {p-value}
  Replication: {yes/no}

  Correlation strength: [{STRONG / MODERATE / WEAK}]

CONFOUND ANALYSIS:
  Potential confounds identified: [N]

  ⚠️  {Confound Z}: Could explain relationship
      Controlled for: {yes/no}
      Plausibility: {high/medium/low}

  Selection bias: {assessment}
  Reverse causation: {assessment}

MECHANISM:
  Proposed: X → A → B → Y
  Step evidence:
    X → A: {evidenced / hypothesized / unknown}
    A → B: {evidenced / hypothesized / unknown}
    B → Y: {evidenced / hypothesized / unknown}

  Mechanism strength: [{STRONG / MODERATE / WEAK / NONE}]

INTERVENTION EVIDENCE:
  Experimental tests: {exist / don't exist}
  Natural experiments: {available / not available}
  Intervention result: {supports / contradicts / inconclusive / none}

--------------------------------------------------------------------------------
                      BRADFORD HILL ASSESSMENT
--------------------------------------------------------------------------------

Criterion          Score (0-2)    Evidence
───────────────────────────────────────────────────────
Strength           [X]            {brief evidence}
Consistency        [X]            {brief evidence}
Specificity        [X]            {brief evidence}
Temporality        [X]            {brief evidence}
Dose-Response      [X]            {brief evidence}
Plausibility       [X]            {brief evidence}
Coherence          [X]            {brief evidence}
Experiment         [X]            {brief evidence}
Analogy            [X]            {brief evidence}
───────────────────────────────────────────────────────
TOTAL              [X/18]

--------------------------------------------------------------------------------
                         CAUSAL VERDICT
--------------------------------------------------------------------------------

CLAIM: "[X] causes [Y]"

VERDICT: [{ESTABLISHED / PROBABLE / POSSIBLE / UNLIKELY / NO EVIDENCE}]

EVIDENCE SUPPORTS:
  ✓ {what the evidence does show}
  ✓ {what the evidence does show}

EVIDENCE DOES NOT SUPPORT:
  ✗ {causal claim that's not supported}
  ✗ {causal claim that's not supported}

RECOMMENDED LANGUAGE DOWNGRADE:

  ORIGINAL: "[X] causes [Y]"

  SHOULD BE: "[X] is associated with [Y], though causation
              has not been established because [reason]"

  OR: "[X] may contribute to [Y], but confounds including
       [Z] have not been ruled out"

--------------------------------------------------------------------------------
                      WHAT WOULD ESTABLISH CAUSATION
--------------------------------------------------------------------------------

To move from [{current verdict}] to [established causation]:

1. {specific evidence needed}
2. {specific evidence needed}
3. {specific evidence needed}

Feasibility: {assessment of whether this is possible}

--------------------------------------------------------------------------------
                      SUMMARY
--------------------------------------------------------------------------------

Claims audited: [N]
  Established causation: [n]
  Probable causation: [n]
  Possible causation: [n]
  Unlikely causation: [n]
  No evidence: [n]

OVERALL: [Assessment of causal rigor in the source]

================================================================================
      "The doctors gave me statistics about populations. I asked about
       THIS patient. That's the difference between pattern-matching
       and understanding."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - Document, claim, or finding to audit for causal validity

---

## Ferrucci Principles Embodied

1. **Statistics vs. causation**: Patterns aren't explanations
2. **Mechanism matters**: How does X lead to Y?
3. **Confound paranoia**: What else could explain this?
4. **Intervention standard**: Can you manipulate and observe?
5. **Honest downgrade**: Most causal claims should be weakened

