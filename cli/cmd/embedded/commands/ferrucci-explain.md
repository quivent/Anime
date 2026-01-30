Generate multi-level explanations that can survive different audiences.

Usage: Explain at multiple depths - if you can't explain it simply, you don't understand it deeply.

**The Explanation Principle:** Watson could trace every answer back to its sources. Understanding isn't just getting the right answer - it's being able to explain WHY at any level of detail.

---

## Why Multi-Level Explanation Matters

Real understanding scales. You should be able to explain:
- To an expert: Full technical depth, no hand-waving
- To a peer: Core ideas with relevant detail
- To a novice: Essential concepts, accessible language
- To yourself: The actual mechanism, no bullshit

If any level fails, your understanding has a gap.

---

## Execution Architecture

```
Phase 1:  Subject Analysis ─────────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Multi-Level Generation ──────┬── Expert Explainer ──────────┐
          (Parallel)                   ├── Peer Explainer ────────────┼──→ Explanation Set
                                       ├── Novice Explainer ──────────┤
                                       └── Core Mechanism Extractor ──┘
    │
    ▼
Phase 3:  Consistency Audit ────────────────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Gap Identification ───────────────────────────────────────── [Sequential]
```

---

## Phase 1: Subject Analysis (Sequential)

```yaml
Task: Subject Analyzer
Subagent: general-purpose
Prompt: |
  Subject to explain: $ARGUMENTS

  Analyze the explanation target:

  {
    "subject": "what needs explaining",
    "type": "concept / mechanism / result / decision / system",
    "key_components": ["the pieces that make it up"],
    "key_relationships": ["how components connect"],
    "prerequisites": ["what someone needs to know first"],
    "common_misconceptions": ["what people often get wrong"],
    "core_insight": "the one thing that makes it click"
  }

  Understand before explaining. Most bad explanations come from incomplete understanding.
```

---

## Phase 2: Multi-Level Generation (Parallel Pool)

```yaml
Task: Expert Explainer
Subagent: general-purpose
Prompt: |
  Subject: {subject_analysis}

  Generate expert-level explanation:

  AUDIENCE: Technical expert in the field
  GOAL: Full depth, maximum precision, no simplification

  {
    "level": "expert",
    "explanation": "The full technical explanation",
    "technical_terms_used": ["list of jargon"],
    "assumptions_about_audience": ["what they already know"],
    "citations_needed": ["what should be referenced"],
    "nuances_included": ["subtle points experts would want"],
    "limitations_acknowledged": ["what's not covered"]
  }

  Don't dumb it down. Experts want precision.
```

```yaml
Task: Peer Explainer
Subagent: general-purpose
Prompt: |
  Subject: {subject_analysis}

  Generate peer-level explanation:

  AUDIENCE: Intelligent colleague in adjacent field
  GOAL: Core ideas with enough context to build on

  {
    "level": "peer",
    "explanation": "Clear explanation with context",
    "technical_terms_used": ["only essential jargon, defined when used"],
    "analogies": ["comparisons to familiar concepts"],
    "why_it_matters": "significance in broader context",
    "key_takeaways": ["the 3-5 things they should remember"]
  }

  Balance depth and accessibility.
```

```yaml
Task: Novice Explainer
Subagent: general-purpose
Prompt: |
  Subject: {subject_analysis}

  Generate novice-level explanation:

  AUDIENCE: Intelligent person with no background
  GOAL: Genuine understanding, not just words

  {
    "level": "novice",
    "explanation": "Simple, clear explanation",
    "no_jargon": true,
    "analogies": ["everyday comparisons"],
    "build_up": ["step by step from basics"],
    "check_understanding": ["questions they should be able to answer"],
    "common_confusions_addressed": ["what usually trips people up"]
  }

  Simple isn't dumbed down. Simple is hard. Simple means you understand deeply.
```

```yaml
Task: Core Mechanism Extractor
Subagent: general-purpose
Prompt: |
  Subject: {subject_analysis}

  Extract the core mechanism:

  AUDIENCE: Yourself
  GOAL: The irreducible essence. No bullshit. No hand-waving.

  {
    "level": "core",
    "mechanism": "The actual thing that makes it work",
    "in_one_sentence": "...",
    "what_actually_happens": "step by step, no abstractions",
    "why_this_and_not_alternatives": "what makes this approach work",
    "the_key_insight": "the one thing someone must grasp"
  }

  If you can't explain the mechanism, you're just reciting.
  This is the test of whether YOU understand.
```

---

## Phase 3: Consistency Audit (Sequential)

```yaml
Task: Consistency Auditor
Subagent: general-purpose
Prompt: |
  All explanations: {expert}, {peer}, {novice}, {core}

  Audit consistency across levels:

  {
    "consistency_checks": [
      {
        "aspect": "core claim",
        "expert_says": "...",
        "peer_says": "...",
        "novice_says": "...",
        "core_says": "...",
        "consistent": true/false,
        "discrepancy": "if inconsistent, what differs"
      }
    ],
    "simplification_validity": [
      {
        "simplification": "what was simplified for novice",
        "technically_accurate": true/false,
        "acceptably_imprecise": true/false,
        "misleading": true/false
      }
    ],
    "overall_consistency": 0.X
  }

  Good explanations simplify without lying.
  Catch anywhere the levels contradict.
```

---

## Phase 4: Gap Identification (Sequential)

```yaml
Task: Gap Identifier
Subagent: general-purpose
Prompt: |
  All explanations: {expert}, {peer}, {novice}, {core}
  Consistency audit: {audit}

  Identify understanding gaps:

  {
    "gaps_revealed": [
      {
        "gap": "what couldn't be explained clearly",
        "appeared_at_level": "which explanation broke down",
        "implication": "what this suggests about understanding",
        "resolution": "how to fix the gap"
      }
    ],
    "hand_waving_detected": [
      {
        "location": "where in the explanation",
        "the_hand_wave": "what was glossed over",
        "actually_needed": "what should have been explained"
      }
    ],
    "understanding_score": 0.X,
    "weakest_area": "where understanding is shakiest"
  }

  The inability to explain reveals the gaps in understanding.
```

---

## Output Format

```
================================================================================
                    FERRUCCI MULTI-LEVEL EXPLANATION
================================================================================

SUBJECT: {subject}

--------------------------------------------------------------------------------
                         EXPERT LEVEL
--------------------------------------------------------------------------------
Audience: Technical expert in the field

{expert_explanation}

Technical terms: {list}
References needed: {list}
Nuances: {list}

--------------------------------------------------------------------------------
                         PEER LEVEL
--------------------------------------------------------------------------------
Audience: Intelligent colleague in adjacent field

{peer_explanation}

Key analogies: {list}
Key takeaways:
  1. {takeaway}
  2. {takeaway}
  3. {takeaway}

Why it matters: {significance}

--------------------------------------------------------------------------------
                         NOVICE LEVEL
--------------------------------------------------------------------------------
Audience: Intelligent person with no background

{novice_explanation}

Build-up:
  Step 1: {foundation}
  Step 2: {build}
  Step 3: {complete}

Check your understanding:
  • Can you answer: {question}?
  • Can you answer: {question}?

--------------------------------------------------------------------------------
                         CORE MECHANISM
--------------------------------------------------------------------------------
The irreducible essence:

{mechanism_explanation}

In one sentence: {one_liner}

The key insight: {key_insight}

--------------------------------------------------------------------------------
                      CONSISTENCY AUDIT
--------------------------------------------------------------------------------

                Expert    Peer      Novice    Core      Consistent?
Claim 1         {text}    {text}    {text}    {text}    ✓/✗
Claim 2         {text}    {text}    {text}    {text}    ✓/✗
...

Simplifications checked:
  ✓ {simplification} - Accurate enough
  ✗ {simplification} - Misleading, needs revision

Overall consistency: [0.X]

--------------------------------------------------------------------------------
                      UNDERSTANDING GAPS
--------------------------------------------------------------------------------

GAPS DETECTED:
  ⚠️ {gap}: Couldn't explain {what} at {level}
     → Implication: {what this means}
     → To fix: {what to study}

HAND-WAVING DETECTED:
  ⚠️ {location}: Glossed over {what}
     → Actually needed: {real explanation}

UNDERSTANDING SCORE: [X/10]

Weakest area: {where understanding is shakiest}

--------------------------------------------------------------------------------
                      THE FEYNMAN TEST
--------------------------------------------------------------------------------

"If you can't explain it simply, you don't understand it well enough."

{assessment of whether all levels succeeded}

================================================================================
      "Watson could trace every answer back to its sources. That's not
       decoration. That's the difference between answering and understanding."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - Subject, concept, or finding to explain at multiple levels

---

## Ferrucci Principles Embodied

1. **Transparency at all levels**: Understanding means being able to show work at any depth
2. **Consistency across simplification**: Simple explanations shouldn't contradict expert ones
3. **Gap detection**: Inability to explain reveals incomplete understanding
4. **The mechanism**: Beyond words to actual causal understanding
5. **Feynman's test**: Simple explanation as proof of deep understanding

