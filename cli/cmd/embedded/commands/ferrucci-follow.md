Test if a system can answer the follow-up question, not just the initial query.

Usage: Stress-test claims by asking "why?" and "how do you know?" recursively.

**The Follow-Up Principle:** Watson wasn't just answering Jeopardy! questions. It was designed to handle the implicit follow-up: "Why is that the answer?"

---

## Why Follow-Ups Matter

Most systems optimize for the first answer. But understanding means surviving interrogation. If you claim X, you should be able to explain:
- Why X and not Y?
- What evidence supports X?
- Under what conditions would X be wrong?
- How does X connect to Z?

This command stress-tests claims through recursive follow-up.

---

## Execution Architecture

```
Phase 1:  Claim Extraction ─────────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Follow-Up Generation ────┬── Why Generator ────────────────┐
          (Parallel)               ├── How Generator ────────────────┼──→ Question Bank
                                   ├── What-If Generator ────────────┤
                                   └── Connection Generator ─────────┘
    │
    ▼
Phase 3:  Recursive Interrogation ──────────────────────────────────── [Sequential]
    │     (Up to N levels deep)
    ▼
Phase 4:  Coherence Assessment ─────────────────────────────────────── [Sequential]
```

---

## Phase 1: Claim Extraction (Sequential)

```yaml
Task: Claim Extractor
Subagent: general-purpose
Prompt: |
  Source: $ARGUMENTS

  Extract all claims that could be interrogated:

  For each claim:
  {
    "id": "C1",
    "claim": "The exact statement",
    "type": "factual / causal / predictive / evaluative",
    "implicit_assumptions": ["what must be true for this claim"],
    "stated_confidence": "explicit or implied"
  }

  Focus on claims that SHOULD survive follow-up questioning.
  Ignore hedged speculation (those correctly signal uncertainty).
```

---

## Phase 2: Follow-Up Generation (Parallel Pool)

```yaml
Task: Why Generator
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}

  For each claim, generate "why" follow-ups:

  {
    "claim_id": "C1",
    "why_questions": [
      "Why is this true rather than [alternative]?",
      "Why should we believe this?",
      "Why does this matter?"
    ]
  }

  Generate questions a skeptical expert would ask.
```

```yaml
Task: How Generator
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}

  For each claim, generate "how" follow-ups:

  {
    "claim_id": "C1",
    "how_questions": [
      "How do you know this?",
      "How was this measured/tested?",
      "How confident are you?"
    ]
  }

  Focus on methodology and evidence.
```

```yaml
Task: What-If Generator
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}

  For each claim, generate counterfactual challenges:

  {
    "claim_id": "C1",
    "what_if_questions": [
      "What if [assumption] is wrong?",
      "What would change your mind?",
      "What's the strongest argument against this?"
    ]
  }

  Test the boundaries of the claim.
```

```yaml
Task: Connection Generator
Subagent: general-purpose
Prompt: |
  Claims: {extracted_claims}

  For each claim, generate connection questions:

  {
    "claim_id": "C1",
    "connection_questions": [
      "How does this relate to [other claim]?",
      "Does this contradict [known fact]?",
      "What does this imply for [domain]?"
    ]
  }

  Test integration with broader knowledge.
```

---

## Phase 3: Recursive Interrogation (Sequential, Iterative)

```yaml
Task: Recursive Interrogator
Subagent: general-purpose
Prompt: |
  Claim: {claim}
  Question bank: {all_follow_up_questions}

  Conduct recursive interrogation:

  ROUND 1:
  - Ask the first follow-up question
  - Evaluate the response
  - Generate level-2 follow-ups based on that response

  ROUND 2:
  - Ask level-2 follow-up
  - Evaluate response
  - Generate level-3 follow-ups

  ROUND 3:
  - Ask level-3 follow-up
  - Evaluate response
  - Note where reasoning bottoms out

  Track:
  {
    "claim_id": "C1",
    "interrogation_depth": N,
    "where_it_broke": "level and question",
    "failure_mode": "circular / undefined / contradictory / admitted_uncertainty",
    "survival_score": 0.X
  }

  A claim that survives 3 levels of "why" deserves confidence.
  A claim that breaks at level 1 shouldn't have been asserted.
```

---

## Phase 4: Coherence Assessment (Sequential)

```
================================================================================
                    FERRUCCI FOLLOW-UP SURVIVAL REPORT
================================================================================

CLAIMS ANALYZED: [N]

--------------------------------------------------------------------------------
                         SURVIVAL SUMMARY
--------------------------------------------------------------------------------

CLAIM               DEPTH SURVIVED    FAILURE MODE         VERDICT
-------------------------------------------------------------------------------
C1: [claim]         3/3               N/A                  ROBUST
C2: [claim]         2/3               Circular @ L3        MODERATE
C3: [claim]         1/3               Undefined @ L2       WEAK
C4: [claim]         0/3               Contradictory @ L1   FRAGILE

--------------------------------------------------------------------------------
                      INTERROGATION DETAILS
--------------------------------------------------------------------------------

CLAIM: [C1]
├── Q1: "Why is this true?"
│   └── A1: [response] - ADEQUATE
│       ├── Q1.1: "How do you know?"
│       │   └── A1.1: [response] - ADEQUATE
│       │       ├── Q1.1.1: "What if that's wrong?"
│       │       │   └── A1.1.1: [response] - ADEQUATE
│       │       └── VERDICT: Survived interrogation
...

--------------------------------------------------------------------------------
                      FAILURE ANALYSIS
--------------------------------------------------------------------------------

Common Failure Modes:

CIRCULAR (claim restates itself):
  • [example from analysis]

UNDEFINED (term lacks clear meaning):
  • [example from analysis]

CONTRADICTORY (conflicts with other claims):
  • [example from analysis]

ADMITTED UNCERTAINTY (correctly signals limit):
  • [example from analysis] - This is GOOD. Honest uncertainty.

--------------------------------------------------------------------------------
                      RECOMMENDATIONS
--------------------------------------------------------------------------------

STRENGTHEN:
• [Claim X] - Add evidence for the level-2 response
• [Claim Y] - Define [term] before asserting

SOFTEN:
• [Claim Z] - Change from assertion to hypothesis

REMOVE:
• [Claim W] - Cannot survive basic interrogation

================================================================================
      "If you can't answer the follow-up question, you didn't really
       understand the first one."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - Document, claim, or system output to interrogate

---

## Ferrucci Principles Embodied

1. **Depth over breadth**: Better to deeply interrogate few claims than superficially check many
2. **Recursive accountability**: Every answer generates new questions
3. **Failure mode taxonomy**: Different failures need different responses
4. **Honest uncertainty**: Admitting "I don't know" at the right depth is success
5. **The Watson test**: Can it explain WHY it gave that answer?

