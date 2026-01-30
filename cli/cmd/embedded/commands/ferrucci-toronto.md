Confidence calibration audit - identify claims that should print "?????" instead of asserting.

Usage: Audit outputs for proper uncertainty signaling. Named after Watson's "What is Toronto?????" answer.

**The Toronto Principle:** Five question marks meant Watson knew it didn't know. That's not a bug - that's the system being honest.

---

## Why Toronto Matters

Watson wagered $947 on Toronto because it had only 14% confidence. The question marks were a feature, not a failure. Most systems today assert with false confidence. This command audits for proper uncertainty signaling.

**Ferrucci's design:** "No component commits to a single answer. All produce features and associated confidence scores."

---

## Execution Architecture

```
Phase 1:  Assertion Inventory ──────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Evidence Binding ─────────┬── Evidence Linker ──────────┐
          (Parallel)                ├── Reversibility Checker ────┼──→ Evidence Map
                                    └── Confidence Estimator ─────┘
    │
    ▼
Phase 3:  Threshold Application ────────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Calibrated Rewrite ───────────────────────────────────── [Sequential]
```

---

## Phase 1: Assertion Inventory (Sequential)

```yaml
Task: Assertion Extractor
Subagent: general-purpose
Prompt: |
  Target document/output: $ARGUMENTS

  Extract every assertion:

  For each statement, categorize:

  1. **Definitive assertions** ("X is Y", "The system does Z")
     - Stated as fact
     - No hedging language

  2. **Hedged assertions** ("X appears to be Y", "likely", "probably")
     - Acknowledges uncertainty
     - But still makes a claim

  3. **Questions/speculation** ("Could X be Y?", "We hypothesize")
     - Not asserting
     - Proper epistemic status

  Return:
  {
    "assertions": [
      {
        "id": "A1",
        "text": "exact quote",
        "location": "line/section",
        "type": "definitive / hedged / speculation",
        "implicit_confidence": "what confidence this IMPLIES"
      }
    ]
  }

  Focus on definitive assertions - these are the claims that need evidence.
```

---

## Phase 2: Evidence Binding (Parallel Pool)

```yaml
Task: Evidence Linker
Subagent: general-purpose
Prompt: |
  Assertions: {assertion_inventory}
  Context: {full_document}

  For each assertion, find supporting evidence:

  {
    "assertion_id": "A1",
    "evidence": [
      {
        "type": "experimental / citation / logical / none",
        "source": "where the evidence comes from",
        "strength": "strong / moderate / weak / none",
        "description": "what the evidence actually shows"
      }
    ],
    "evidence_gap": "what evidence is MISSING"
  }

  Be strict. "It seems obvious" is not evidence.
  "Prior work suggests" needs a citation.
  "Our experiments show" needs specific results.
```

```yaml
Task: Reversibility Checker
Subagent: general-purpose
Prompt: |
  Assertions: {assertion_inventory}

  For each assertion:

  {
    "assertion_id": "A1",
    "reversible": true/false,
    "reversal_condition": "what new evidence would change this",
    "reversal_impact": "if reversed, what else breaks"
  }

  Irreversible assertions (can't be updated with new evidence) should be rare.
  Flag any that seem inappropriately certain.
```

```yaml
Task: Confidence Estimator
Subagent: general-purpose
Prompt: |
  Assertions: {assertion_inventory}
  Evidence: {evidence_map}

  Estimate actual confidence for each assertion:

  Apply Watson-style confidence estimation:
  - Multiple independent evidence sources = higher confidence
  - Single source = lower confidence
  - No evidence = very low confidence
  - Contradictory evidence = uncertainty flag

  {
    "assertion_id": "A1",
    "estimated_confidence": 0.XX,
    "confidence_breakdown": {
      "evidence_strength": 0.X,
      "source_independence": 0.X,
      "reversibility_penalty": 0.X
    },
    "confidence_rationale": "why this confidence level"
  }
```

---

## Phase 3: Threshold Application (Sequential)

```yaml
Task: Toronto Thresholder
Subagent: general-purpose
Prompt: |
  Assertions with confidence: {confidence_estimates}

  Apply Toronto thresholds:

  CLAIM TYPE              THRESHOLD    BELOW THRESHOLD ACTION
  ──────────────────────────────────────────────────────────────
  System capability       0.90         Add "?????" / strong hedge
  Experimental result     0.80         Add uncertainty interval
  Methodology choice      0.70         Acknowledge alternatives
  Future prediction       0.60         Frame as hypothesis
  Design decision         0.50         Note as choice, not fact

  For each assertion:
  {
    "assertion_id": "A1",
    "current_confidence": 0.XX,
    "required_threshold": 0.XX,
    "passes_threshold": true/false,
    "recommended_action": "none / hedge / question_marks / remove"
  }

  Watson wagered $947 on Toronto at 14% confidence.
  The question marks were appropriate signaling.
  What wagers are being made in this document?
```

---

## Phase 4: Calibrated Rewrite (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI TORONTO AUDIT
                   Confidence Calibration Report
════════════════════════════════════════════════════════════════════════════════

📊 AUDIT SUMMARY
────────────────────────────────────────────────────────────────────────────────

Total Assertions Analyzed:    [N]
├── Passing Threshold:        [N] ([%])
├── Below Threshold:          [N] ([%])  ← Need attention
└── Severely Miscalibrated:   [N] ([%])  ← Critical

────────────────────────────────────────────────────────────────────────────────
                      ASSERTIONS NEEDING CALIBRATION
────────────────────────────────────────────────────────────────────────────────

❌ HIGH RISK (confidence < 0.50, stated as fact):

  Location: [line/section]
  Original: "[exact assertion]"
  Confidence: [XX]%
  Evidence: [summary]

  RECOMMENDED: [specific rewrite with appropriate uncertainty]

⚠️  MEDIUM RISK (confidence 0.50-0.80, understated uncertainty):

  Location: [line/section]
  Original: "[assertion]"
  Confidence: [XX]%

  RECOMMENDED: [rewrite]

────────────────────────────────────────────────────────────────────────────────
                         THE TORONTO TEST
────────────────────────────────────────────────────────────────────────────────

If this were Jeopardy!, here's what Watson would wager:

Assertion              Confidence    Wager (of $1000)    Signal
──────────────────────────────────────────────────────────────────
[A1]                   95%           $950                Assert confidently
[A2]                   72%           $500                Hedge appropriately
[A3]                   34%           $100                Add "?????"
[A4]                   14%           $47 (Toronto)       Strong uncertainty

────────────────────────────────────────────────────────────────────────────────
                      CALIBRATED VERSION
────────────────────────────────────────────────────────────────────────────────

[Full rewrite of document with appropriate confidence markers]

Changes made:
• [N] assertions softened with uncertainty language
• [N] assertions annotated with confidence percentages
• [N] "?????" markers added for low-confidence claims
• [N] assertions unchanged (appropriately confident)

────────────────────────────────────────────────────────────────────────────────
                      CONFIDENCE DISTRIBUTION
────────────────────────────────────────────────────────────────────────────────

0-20%   ████░░░░░░░░░░░░░░░░  [N] assertions  ← Should rarely assert
20-40%  ██████░░░░░░░░░░░░░░  [N] assertions  ← Heavy hedging needed
40-60%  ████████░░░░░░░░░░░░  [N] assertions  ← Moderate hedging
60-80%  ██████████░░░░░░░░░░  [N] assertions  ← Light hedging
80-100% ████████████████████  [N] assertions  ← Can assert confidently

Overall calibration score: [X/10]

════════════════════════════════════════════════════════════════════════════════
      "What is Toronto?????"

       The question marks weren't a bug. They were Watson being honest
       about what it didn't know.
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - Document, output, or claim to audit for confidence calibration

---

## Ferrucci Principles Embodied

1. **Pervasive confidence estimation**: Every claim has uncertainty
2. **Honest signaling**: Low confidence should be visible, not hidden
3. **The wager test**: Would you bet money on this? How much?
4. **Calibration over assertion**: Better to be uncertain than wrong
5. **Toronto as feature**: "?????" is appropriate when you don't know
