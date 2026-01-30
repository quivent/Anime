Design hybrid architectures that combine LLM fluency with formal reasoning reliability.

Usage: Structure a solution using the LLM sandwich pattern - language models for interface, formal reasoning for computation.

**The Sandwich Principle:** LLMs are skilled writers and storytellers, not reasoning experts. They don't understand the logic behind language. Wrap them around a formal core that actually computes - deterministically, reliably, explicably.

---

## Why Hybrid Architecture Matters

The current AI landscape confuses fluency with understanding:
- LLMs produce text that *sounds* like reasoning
- They're often wrong, or right for wrong reasons
- They can't explain why they said what they said
- They're black boxes when transparency is required

The solution isn't to abandon LLMs. It's to use them for what they're good at:
- Natural language interface
- Broad knowledge access
- Fluent communication

And wrap them around formal reasoning that provides:
- Deterministic computation
- Reliable inference
- Explainable results
- Causal understanding

---

## Execution Architecture

```
Phase 1:  Problem Analysis ────────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Layer Design ───────────┬── Interface Layer Designer ──────┐
          (Parallel)              ├── Reasoning Core Designer ───────┼──→ Architecture
                                  ├── Integration Designer ──────────┤
                                  └── Verification Designer ─────────┘
    │
    ▼
Phase 3:  Flow Specification ──────────────────────────────────────── [Sequential]
    │
    ▼
Phase 4:  Reliability Analysis ────────────────────────────────────── [Sequential]
```

---

## Phase 1: Problem Analysis (Sequential)

```yaml
Task: Problem Analyzer
Subagent: general-purpose
Prompt: |
  Problem: $ARGUMENTS

  Analyze for hybrid architecture suitability:

  {
    "problem": "what needs to be solved",
    "stakes": "low / medium / high / critical",
    "explainability_required": true/false,
    "reliability_required": "approximate OK / must be correct / zero tolerance",
    "current_approach": "how is this done now",
    "failure_modes": "what goes wrong with pure LLM approach"
  }

  Key questions:
  1. Does this require deterministic computation?
  2. Does someone need to understand WHY the answer emerged?
  3. Can we afford statistical betting, or do we need causal certainty?
  4. Is fluency enough, or do we need formal correctness?

  Be honest. Not everything needs hybrid architecture.
  But high-stakes applications almost always do.
```

---

## Phase 2: Layer Design (Parallel Pool)

```yaml
Task: Interface Layer Designer
Subagent: general-purpose
Prompt: |
  Problem: {problem_analysis}

  Design the LLM interface layers:

  {
    "input_layer": {
      "function": "translate human language to formal representation",
      "llm_role": "understanding user intent",
      "output_format": "structured query / formal specification",
      "ambiguity_handling": "how to resolve unclear inputs"
    },
    "output_layer": {
      "function": "translate formal results to human language",
      "llm_role": "natural language generation",
      "input_format": "structured result from reasoning core",
      "explanation_generation": "how to explain the reasoning"
    }
  }

  The LLM layers are the bread of the sandwich.
  They handle fluency. They don't handle logic.
```

```yaml
Task: Reasoning Core Designer
Subagent: general-purpose
Prompt: |
  Problem: {problem_analysis}

  Design the formal reasoning core:

  {
    "core_function": "what computation happens here",
    "reasoning_type": "deductive / inductive / abductive / causal / mathematical",
    "representation": "how knowledge is represented formally",
    "inference_mechanism": "how conclusions are derived",
    "determinism": "is output deterministic given input",
    "explainability": "can each step be traced and justified"
  }

  Requirements:
  - Must be formally specifiable
  - Must produce consistent results
  - Must be able to show its work
  - Must NOT rely on pattern matching for critical decisions

  This is the meat of the sandwich.
  This is where reliability comes from.
```

```yaml
Task: Integration Designer
Subagent: general-purpose
Prompt: |
  Interface layers: {interface_design}
  Reasoning core: {core_design}

  Design the integration points:

  {
    "input_translation": {
      "from": "natural language",
      "to": "formal representation",
      "validation": "how to verify translation correctness",
      "fallback": "what happens if translation fails"
    },
    "output_translation": {
      "from": "formal result",
      "to": "natural language",
      "fidelity": "how to ensure explanation matches reasoning",
      "confidence_communication": "how to express uncertainty"
    },
    "error_handling": {
      "translation_errors": "...",
      "reasoning_errors": "...",
      "integration_errors": "..."
    }
  }

  The integration is where systems fail.
  Design for failure. Make failures visible.
```

```yaml
Task: Verification Designer
Subagent: general-purpose
Prompt: |
  Full architecture: {all_layers}

  Design verification mechanisms:

  {
    "input_verification": {
      "did_llm_understand_correctly": "how to check",
      "is_formal_representation_valid": "validation rules"
    },
    "reasoning_verification": {
      "is_inference_correct": "formal proof / test cases",
      "are_assumptions_valid": "assumption checking"
    },
    "output_verification": {
      "does_explanation_match_reasoning": "consistency check",
      "is_natural_language_faithful": "fidelity check"
    },
    "end_to_end": {
      "roundtrip_test": "input → formal → result → explanation → verify",
      "adversarial_test": "inputs designed to break the system"
    }
  }

  Watson had confidence scores everywhere.
  Your system needs verification everywhere.
```

---

## Phase 3: Flow Specification (Sequential)

```yaml
Task: Flow Specifier
Subagent: general-purpose
Prompt: |
  Full architecture: {all_designs}

  Specify the complete data flow:

  ```
  USER INPUT (natural language)
       │
       ▼
  ┌─────────────────────────────────────────────┐
  │  LLM INPUT LAYER                            │
  │  - Parse intent                             │
  │  - Resolve ambiguities                      │
  │  - Translate to formal representation       │
  │  - Output: structured query                 │
  └─────────────────────────────────────────────┘
       │
       ▼
  ┌─────────────────────────────────────────────┐
  │  FORMAL REASONING CORE                      │
  │  - Validate input                           │
  │  - Execute deterministic inference          │
  │  - Generate proof trace / explanation data  │
  │  - Output: structured result + reasoning    │
  └─────────────────────────────────────────────┘
       │
       ▼
  ┌─────────────────────────────────────────────┐
  │  LLM OUTPUT LAYER                           │
  │  - Translate result to natural language     │
  │  - Generate human-readable explanation      │
  │  - Communicate confidence appropriately     │
  │  - Output: natural language response        │
  └─────────────────────────────────────────────┘
       │
       ▼
  USER OUTPUT (natural language + explanation)
  ```

  For each transition, specify:
  - Exact data format
  - Validation performed
  - Error handling
  - Logging / traceability
```

---

## Phase 4: Reliability Analysis (Sequential)

```yaml
Task: Reliability Analyzer
Subagent: general-purpose
Prompt: |
  Complete architecture: {flow_specification}

  Analyze reliability:

  {
    "single_points_of_failure": [
      {
        "location": "where",
        "failure_mode": "how it fails",
        "consequence": "what happens",
        "mitigation": "how to address"
      }
    ],
    "llm_failure_modes": [
      "hallucination in input translation",
      "unfaithful output explanation",
      "confidence miscommunication"
    ],
    "formal_failure_modes": [
      "invalid input from LLM layer",
      "incomplete knowledge base",
      "edge case in inference rules"
    ],
    "overall_reliability": {
      "deterministic_portion": "X% of the pipeline",
      "statistical_portion": "Y% of the pipeline",
      "critical_path_reliability": "assessment"
    }
  }

  The goal: formal core handles all critical decisions.
  LLM layers handle only interface / fluency.
  If LLM fails, system degrades gracefully, doesn't give wrong answers.
```

---

## Output Format

```
================================================================================
                    FERRUCCI HYBRID ARCHITECTURE
================================================================================

PROBLEM: {problem_description}

STAKES: [{LOW / MEDIUM / HIGH / CRITICAL}]
EXPLAINABILITY REQUIRED: [{YES / NO}]
HYBRID APPROPRIATE: [{YES / NO / OVERKILL}]

--------------------------------------------------------------------------------
                         ARCHITECTURE OVERVIEW
--------------------------------------------------------------------------------

┌─────────────────────────────────────────────────────────────────────────────┐
│                           LLM INPUT LAYER                                   │
│  Role: Natural language understanding, intent extraction                    │
│  Output: {formal_representation_type}                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        FORMAL REASONING CORE                                │
│  Type: {reasoning_type}                                                     │
│  Representation: {knowledge_representation}                                 │
│  Inference: {inference_mechanism}                                           │
│  Deterministic: {YES/NO}                                                    │
│  Explainable: {YES/NO}                                                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          LLM OUTPUT LAYER                                   │
│  Role: Result translation, explanation generation                           │
│  Input: {structured_result_type}                                            │
└─────────────────────────────────────────────────────────────────────────────┘

--------------------------------------------------------------------------------
                         INTEGRATION POINTS
--------------------------------------------------------------------------------

INPUT TRANSLATION:
  From: Natural language
  To: {formal_format}
  Validation: {validation_method}
  Failure handling: {fallback}

OUTPUT TRANSLATION:
  From: {formal_result}
  To: Natural language
  Fidelity check: {method}
  Confidence: {how_communicated}

--------------------------------------------------------------------------------
                         RELIABILITY ANALYSIS
--------------------------------------------------------------------------------

CRITICAL PATH:
  □ All critical decisions made in formal core: {YES/NO}
  □ LLM failures degrade gracefully: {YES/NO}
  □ Every inference step traceable: {YES/NO}
  □ Confidence calibrated and communicated: {YES/NO}

SINGLE POINTS OF FAILURE:
  ⚠️ {failure_point_1}: {mitigation}
  ⚠️ {failure_point_2}: {mitigation}

OVERALL RELIABILITY ASSESSMENT:
  Deterministic portion: {X%}
  Statistical portion: {Y%}
  Critical path reliability: [{HIGH / MEDIUM / LOW}]

--------------------------------------------------------------------------------
                         VERIFICATION DESIGN
--------------------------------------------------------------------------------

INPUT VERIFICATION:
  □ {check_1}
  □ {check_2}

REASONING VERIFICATION:
  □ {check_1}
  □ {check_2}

OUTPUT VERIFICATION:
  □ {check_1}
  □ {check_2}

--------------------------------------------------------------------------------
                         WHY THIS ARCHITECTURE
--------------------------------------------------------------------------------

FLUENCY FROM: LLM layers (input understanding, output generation)
RELIABILITY FROM: Formal reasoning core (deterministic, traceable)
EXPLANATION FROM: Reasoning trace → LLM translation

The user gets natural language interaction.
The system gets formal correctness.
The explanation shows the work.

================================================================================
      "LLMs are skilled writers and storytellers, not reasoning experts.
       Wrap them around a formal core that actually computes."
                                — Dave Ferrucci
================================================================================
```

---

## Arguments

$ARGUMENTS - Problem or system to design with hybrid architecture

---

## Ferrucci Principles Embodied

1. **Fluency ≠ Understanding**: LLMs talk beautifully; that's not reasoning
2. **Formal core for reliability**: Deterministic computation for critical decisions
3. **Show your work**: Every inference traceable and explainable
4. **Graceful degradation**: LLM failures shouldn't produce wrong answers
5. **The specific case**: Formal reasoning handles this case, not statistical betting

---

## When to Use This

Use `/ferrucci-sandwich` when:
- Stakes are high (medicine, law, finance, safety)
- Explainability is required
- Statistical betting is insufficient
- You need to answer "why?"
- Current pure-LLM approach has unacceptable failure modes

Don't use when:
- Approximate answers are fine
- Fluency is the only requirement
- Cost of formal reasoning exceeds benefit
- The problem doesn't require causal understanding
