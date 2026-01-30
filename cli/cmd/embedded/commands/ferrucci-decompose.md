Break complex questions into atomic, testable sub-questions.

Usage: Decompose a big question until each piece can be independently validated.

**The Decomposition Principle:** Watson didn't answer questions directly. It decomposed them into hundreds of sub-questions, answered each, then combined. Complex questions hide simpler ones.

---

## Why Decomposition Matters

Big questions can't be directly tested. "Does the system work?" is unanswerable. But:
- "Does component X produce valid output?" - testable
- "Does component Y handle edge case Z?" - testable
- "Do components X and Y integrate correctly?" - testable

Decomposition transforms vague questions into rigorous tests.

---

## Execution Architecture

```
Phase 1:  Question Formalization ───────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Multi-Perspective Decomposition ─┬── Structural Decomposer ──┐
          (Parallel)                       ├── Causal Decomposer ───────┼──→ Sub-Question Tree
                                           ├── Temporal Decomposer ─────┤
                                           └── Conditional Decomposer ──┘
    │
    ▼
Phase 3:  Sub-Question Validation ──────────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Integration Mapping ──────────────────────────────────────── [Sequential]
```

---

## Phase 1: Question Formalization (Sequential)

```yaml
Task: Question Formalizer
Subagent: general-purpose
Prompt: |
  Question: $ARGUMENTS

  Formalize the question:

  {
    "original_question": "verbatim",
    "question_type": "existence / comparison / causation / mechanism / evaluation",
    "implicit_scope": "what's assumed",
    "ambiguities": ["terms or concepts that need clarification"],
    "answerable_form": "rewritten to be directly answerable",
    "success_criteria": "what would constitute a complete answer"
  }

  Clarify before decomposing. Decomposing an ambiguous question produces ambiguous sub-questions.
```

---

## Phase 2: Multi-Perspective Decomposition (Parallel Pool)

```yaml
Task: Structural Decomposer
Subagent: general-purpose
Prompt: |
  Question: {formalized_question}

  Decompose by STRUCTURE (components/parts):

  {
    "decomposition_type": "structural",
    "sub_questions": [
      {
        "id": "S1",
        "sub_question": "Does [component 1] work?",
        "parent": "root",
        "atomic": true/false,
        "testable": true/false,
        "test_method": "how to answer this"
      },
      {
        "id": "S2",
        "sub_question": "Does [component 2] work?",
        "parent": "root",
        "atomic": true/false,
        ...
      }
    ],
    "integration_question": "Do the components work together?"
  }

  If a sub-question isn't atomic, decompose further.
```

```yaml
Task: Causal Decomposer
Subagent: general-purpose
Prompt: |
  Question: {formalized_question}

  Decompose by CAUSATION (cause/effect chain):

  {
    "decomposition_type": "causal",
    "sub_questions": [
      {
        "id": "C1",
        "sub_question": "Does [cause] actually occur?",
        "causal_role": "cause / mechanism / effect",
        "parent": "root",
        "atomic": true/false,
        "testable": true/false
      },
      {
        "id": "C2",
        "sub_question": "Does [cause] lead to [intermediate]?",
        "causal_role": "mechanism",
        ...
      },
      {
        "id": "C3",
        "sub_question": "Does [intermediate] lead to [effect]?",
        "causal_role": "mechanism",
        ...
      }
    ],
    "causal_chain": ["C1", "C2", "C3"]
  }

  Trace the causal path. Each link must be validated.
```

```yaml
Task: Temporal Decomposer
Subagent: general-purpose
Prompt: |
  Question: {formalized_question}

  Decompose by TIME (sequence/phases):

  {
    "decomposition_type": "temporal",
    "sub_questions": [
      {
        "id": "T1",
        "sub_question": "Does [phase 1] happen correctly?",
        "temporal_position": "initial",
        "prerequisites": [],
        "atomic": true/false,
        "testable": true/false
      },
      {
        "id": "T2",
        "sub_question": "Does [phase 2] follow correctly?",
        "temporal_position": "middle",
        "prerequisites": ["T1"],
        ...
      }
    ],
    "sequence": ["T1", "T2", "T3", ...]
  }

  Questions about processes have temporal structure.
```

```yaml
Task: Conditional Decomposer
Subagent: general-purpose
Prompt: |
  Question: {formalized_question}

  Decompose by CONDITIONS (if/then scenarios):

  {
    "decomposition_type": "conditional",
    "sub_questions": [
      {
        "id": "IF1",
        "sub_question": "Given [condition A], does [outcome] hold?",
        "condition": "A",
        "context_specific": true,
        "atomic": true/false,
        "testable": true/false
      },
      {
        "id": "IF2",
        "sub_question": "Given [condition B], does [outcome] hold?",
        "condition": "B",
        ...
      }
    ],
    "condition_space": "What conditions need to be tested",
    "universal_claim": "Does it hold across ALL conditions?"
  }

  General claims hide conditional exceptions.
```

---

## Phase 3: Sub-Question Validation (Sequential)

```yaml
Task: Sub-Question Validator
Subagent: general-purpose
Prompt: |
  All sub-questions from all decomposers:
  {structural_subs}, {causal_subs}, {temporal_subs}, {conditional_subs}

  Validate and merge:

  1. **Redundancy check**: Which sub-questions overlap?
  2. **Completeness check**: Do sub-questions fully cover the original?
  3. **Atomicity check**: Are all marked "atomic" truly atomic?
  4. **Testability check**: Can each be independently validated?

  {
    "merged_sub_questions": [
      {
        "id": "Q1",
        "sub_question": "...",
        "sources": ["S1", "C1"],  // Which decomposers identified this
        "final_atomic": true,
        "final_testable": true,
        "priority": "critical / important / supporting"
      }
    ],
    "coverage_assessment": "Does this set fully answer the original?",
    "gaps": "Any aspects not covered by sub-questions"
  }
```

---

## Phase 4: Integration Mapping (Sequential)

```yaml
Task: Integration Mapper
Subagent: general-purpose
Prompt: |
  Validated sub-questions: {validated_subs}
  Original question: {formalized_question}

  Map how sub-question answers combine:

  {
    "answer_integration": {
      "rule": "ALL / ANY / WEIGHTED / CONDITIONAL",
      "explanation": "How sub-answers combine to answer the original"
    },
    "dependencies": {
      "Q1": [],  // Q1 can be answered independently
      "Q2": ["Q1"],  // Q2 depends on Q1's answer
      ...
    },
    "evaluation_order": ["Q1", "Q3", "Q2", "Q4", ...],
    "early_termination": {
      "if_Q1_false": "Original question answered NO",
      ...
    }
  }

  Define exactly how to go from sub-answers to final answer.
```

---

## Output Format

```
================================================================================
                    FERRUCCI DECOMPOSITION
================================================================================

ORIGINAL QUESTION:
{original_question}

FORMALIZED AS:
{answerable_form}

--------------------------------------------------------------------------------
                         SUB-QUESTION TREE
--------------------------------------------------------------------------------

ROOT: {original_question}
│
├─ Q1 [critical]: {sub_question_1}
│  ├─ Source: Structural (S1), Causal (C1)
│  ├─ Atomic: Yes
│  ├─ Testable: Yes
│  └─ Test: {how_to_test}
│
├─ Q2 [critical]: {sub_question_2}
│  ├─ Source: Causal (C2)
│  ├─ Atomic: Yes
│  ├─ Testable: Yes
│  ├─ Depends on: Q1
│  └─ Test: {how_to_test}
│
├─ Q3 [important]: {sub_question_3}
│  ├─ Source: Temporal (T1), Conditional (IF1)
│  ├─ Atomic: No → Further decomposed:
│  │  ├─ Q3.1: {atomic_sub}
│  │  └─ Q3.2: {atomic_sub}
│  └─ Test: {how_to_test}
│
└─ Q4 [supporting]: {sub_question_4}
   └─ ...

--------------------------------------------------------------------------------
                         DECOMPOSITION VIEWS
--------------------------------------------------------------------------------

STRUCTURAL VIEW:
  Component A → Q1
  Component B → Q2
  Integration → Q5

CAUSAL VIEW:
  Cause → Q1
  Mechanism → Q2, Q3
  Effect → Q4

TEMPORAL VIEW:
  Phase 1 → Q1
  Phase 2 → Q2
  Phase 3 → Q3, Q4

CONDITIONAL VIEW:
  If condition A: Q1 + Q2 sufficient
  If condition B: Q3 also required

--------------------------------------------------------------------------------
                         ANSWER INTEGRATION
--------------------------------------------------------------------------------

To answer: "{original_question}"

INTEGRATION RULE: {ALL/ANY/WEIGHTED/CONDITIONAL}

If Q1 = YES and Q2 = YES and Q3 = YES:
  → Answer is YES with confidence based on Q4

If Q1 = NO:
  → Answer is NO (early termination)

If Q1 = YES but Q2 = NO:
  → Answer is PARTIAL, specifically: [interpretation]

--------------------------------------------------------------------------------
                         EVALUATION PLAN
--------------------------------------------------------------------------------

PHASE 1 (can parallelize):
  □ Q1: {test method}
  □ Q4: {test method}

PHASE 2 (depends on Phase 1):
  □ Q2: {test method} [requires Q1]

PHASE 3 (depends on Phase 2):
  □ Q3.1: {test method}
  □ Q3.2: {test method}

PHASE 4 (integration):
  □ Combine answers using integration rule
  □ Compute final confidence

--------------------------------------------------------------------------------
                      COMPLETENESS CHECK
--------------------------------------------------------------------------------

Coverage: [X%] of original question scope

COVERED:
  ✓ {aspect 1}
  ✓ {aspect 2}

NOT COVERED (gaps):
  ✗ {aspect that wasn't decomposed}

ASSUMPTION (not tested, taken as given):
  • {implicit assumption}

================================================================================
      "A question you can't decompose is a question you don't understand."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - The complex question to decompose into atomic sub-questions

---

## Ferrucci Principles Embodied

1. **Atomic testability**: Break down until each piece is independently verifiable
2. **Multiple perspectives**: Structural, causal, temporal, conditional views
3. **Integration clarity**: Know exactly how sub-answers combine
4. **Early termination**: Some sub-answers short-circuit the whole question
5. **Completeness awareness**: Know what you're covering and what you're not

