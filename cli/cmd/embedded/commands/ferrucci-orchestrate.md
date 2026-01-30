Execute Dave Ferrucci's research vision through coordinated parallel agents.

Usage: Decompose a research objective into parallel workstreams with clear success criteria.

**Leadership Philosophy:** I manage from the architecture down. Foundational design principles shape everything. If the architecture is right, good people find their way to good solutions.

---

## The Ferrucci Orchestration Method

Watson wasn't built by one person. It was built by twenty brilliant people working in parallel, unified by four principles:

1. **Massive Parallelism** - Many approaches simultaneously
2. **Many Experts** - Each agent brings specialized capability
3. **Pervasive Confidence** - Every output includes certainty estimation
4. **Shallow + Deep Integration** - Quick heuristics AND rigorous analysis

This command applies those principles to any research objective.

---

## Execution Architecture

```
Phase 1:  Vision Articulation ───────────────────────────────────── [Sequential]
    │     "What are we trying to prove or build?"
    │
    ▼
Phase 2:  Decomposition ────────────┬── Hypothesis Extractor ─────┐
          (Parallel)                ├── Dependency Mapper ────────┼──→ Work Graph
                                    └── Risk Identifier ──────────┘
    │
    ▼
Phase 3:  Agent Assignment ─────────────────────────────────────── [Sequential]
    │     "Who does what, with what success criteria?"
    │
    ▼
Phase 4:  Parallel Execution ───────┬── Agent 1: [task] ──────────┐
          (Parallel Pool)           ├── Agent 2: [task] ──────────┤
                                    ├── Agent 3: [task] ──────────┼──→ Results
                                    ├── Agent N: [task] ──────────┤
                                    └── Sentinel: [monitoring] ───┘
    │
    ▼
Phase 5:  Consensus Integration ────┬── Result Validator ─────────┐
          (Parallel)                ├── Conflict Resolver ────────┼──→ Synthesis
                                    └── Confidence Aggregator ────┘
    │
    ▼
Phase 6:  Verdict & Direction ──────────────────────────────────── [Sequential]
```

---

## Phase 1: Vision Articulation (Sequential)

🎯 **Define the Objective with Precision**

```yaml
Task: Vision Crystallizer
Subagent: general-purpose
Prompt: |
  User's objective: {user_input}

  Crystallize this into a precise research vision:

  1. **Core Question**: What exactly are we trying to answer or build?

  2. **Success Criteria**: What would constitute proof of success?
     - Quantitative thresholds where applicable
     - Qualitative markers where needed

  3. **Failure Criteria**: What would disprove the hypothesis?
     - Be honest about what would constitute failure

  4. **Scope Boundaries**: What is explicitly OUT of scope?

  5. **Architectural Constraints**: What design principles must be maintained?

  Return:
  {
    "vision": "One sentence",
    "core_question": "...",
    "success_criteria": [...],
    "failure_criteria": [...],
    "scope": {"in": [...], "out": [...]},
    "constraints": [...]
  }
```

**Output:** Crystallized vision with explicit success/failure criteria

---

## Phase 2: Decomposition (Parallel Pool)

```yaml
Task: Hypothesis Extractor
Subagent: general-purpose
Prompt: |
  Vision: {crystallized_vision}

  Extract all testable hypotheses embedded in this vision:

  For each hypothesis:
  {
    "id": "H1",
    "statement": "If X then Y",
    "test_method": "How to test this",
    "evidence_needed": "What would prove/disprove",
    "priority": "critical / important / supporting"
  }

  Identify which hypotheses are:
  - Independent (can test in parallel)
  - Dependent (must sequence)
  - Blocking (failure stops everything)
```

```yaml
Task: Dependency Mapper
Subagent: general-purpose
Prompt: |
  Vision: {crystallized_vision}
  Hypotheses: {extracted_hypotheses}

  Map dependencies between work items:

  Create directed acyclic graph:
  {
    "nodes": [
      {"id": "W1", "description": "...", "estimated_effort": "..."},
      ...
    ],
    "edges": [
      {"from": "W1", "to": "W3", "type": "blocks"},
      {"from": "W2", "to": "W3", "type": "informs"},
      ...
    ],
    "critical_path": ["W1", "W3", "W5"],
    "parallelizable_groups": [["W1", "W2"], ["W4", "W6"]]
  }

  Identify maximum parallelism possible.
```

```yaml
Task: Risk Identifier
Subagent: general-purpose
Prompt: |
  Vision: {crystallized_vision}
  Dependencies: {dependency_map}

  Identify risks to the research objective:

  For each risk:
  {
    "id": "R1",
    "description": "What could go wrong",
    "probability": "low / medium / high",
    "impact": "minor / moderate / severe",
    "mitigation": "How to reduce risk",
    "detection": "How we'll know if it's happening"
  }

  Flag any risks that should abort the mission if realized.
```

**Output:** Complete work graph with hypotheses, dependencies, and risks

---

## Phase 3: Agent Assignment (Sequential)

```yaml
Task: Agent Architect
Subagent: general-purpose
Prompt: |
  Work graph: {work_graph}
  Risks: {identified_risks}

  Design the agent pool:

  For each parallelizable work item, assign an agent:

  {
    "agent_id": "A1",
    "work_item": "W1",
    "role": "Descriptive role name",
    "expertise_required": "What this agent must be good at",
    "prompt_template": "The exact prompt this agent will receive",
    "success_criteria": "How to evaluate this agent's output",
    "confidence_requirement": "Minimum confidence to accept output",
    "timeout": "Maximum time before escalation",
    "fallback": "What to do if agent fails"
  }

  Also assign:
  - **Sentinel Agent**: Monitors all others, detects failures early
  - **Integration Agent**: Combines outputs, resolves conflicts

  Design for maximum parallelism with clear handoff points.
```

**Output:** Complete agent assignment with prompts and success criteria

---

## Phase 4: Parallel Execution (Dynamic Pool)

Launch all assigned agents simultaneously:

```yaml
# Template for each work agent
Task: {agent_role}
Subagent: general-purpose
Prompt: |
  FERRUCCI ORCHESTRATION - Agent {agent_id}

  YOUR ROLE: {role}

  VISION CONTEXT:
  {crystallized_vision}

  YOUR SPECIFIC TASK:
  {work_item_description}

  SUCCESS CRITERIA:
  {success_criteria}

  CONSTRAINTS:
  - Include confidence score (0-1) for all conclusions
  - Flag any uncertainties explicitly
  - Note any dependencies on other agents' work
  - If you cannot complete, explain why clearly

  DELIVER:
  {expected_output_format}

  Remember: Your output will be integrated with {N} other agents.
  Be precise. Be honest about uncertainty. Show your reasoning.
```

```yaml
# Sentinel agent runs alongside all others
Task: Sentinel Monitor
Subagent: general-purpose
Prompt: |
  FERRUCCI ORCHESTRATION - Sentinel

  You are monitoring {N} parallel agents working toward:
  {crystallized_vision}

  Your job:
  1. Track progress of all agents
  2. Detect early failures or blockers
  3. Identify if any agent is stuck
  4. Flag if outputs are inconsistent with each other
  5. Recommend intervention if needed

  Report status and any concerns.
```

**Output:** Raw results from all agents plus sentinel report

---

## Phase 5: Consensus Integration (Parallel Pool)

```yaml
Task: Result Validator
Subagent: general-purpose
Prompt: |
  Agent outputs: {all_agent_outputs}
  Success criteria: {per_agent_success_criteria}

  Validate each agent's output:

  For each agent:
  {
    "agent_id": "A1",
    "met_success_criteria": true/false,
    "confidence_acceptable": true/false,
    "output_usable": true/false,
    "issues": [...],
    "quality_score": 0.X
  }

  Flag any outputs that should be rejected or rerun.
```

```yaml
Task: Conflict Resolver
Subagent: general-purpose
Prompt: |
  Agent outputs: {all_agent_outputs}

  Identify conflicts between agents:

  - Contradictory conclusions
  - Inconsistent assumptions
  - Incompatible recommendations

  For each conflict:
  {
    "agents": ["A1", "A3"],
    "conflict": "Description",
    "resolution_strategy": "How to resolve",
    "recommended_answer": "Which to believe and why",
    "confidence_in_resolution": 0.X
  }
```

```yaml
Task: Confidence Aggregator
Subagent: general-purpose
Prompt: |
  Agent outputs: {all_agent_outputs}
  Validations: {validation_results}
  Conflicts: {conflict_resolutions}

  Aggregate confidence across the entire effort:

  {
    "overall_confidence": 0.X,
    "confidence_by_hypothesis": {
      "H1": 0.X,
      "H2": 0.X,
      ...
    },
    "weakest_links": ["Where confidence is lowest"],
    "strongest_findings": ["Where confidence is highest"],
    "remaining_uncertainty": ["What we still don't know"]
  }

  Apply the Watson principle: confidence emerges from consensus.
  Many independent agents arriving at the same answer = high confidence.
  Disagreement = lower confidence, flag for attention.
```

**Output:** Validated, conflict-resolved, confidence-weighted synthesis

---

## Phase 6: Verdict & Direction (Sequential)

```
════════════════════════════════════════════════════════════════════════════════
                    FERRUCCI ORCHESTRATION REPORT
════════════════════════════════════════════════════════════════════════════════

🎯 VISION
────────────────────────────────────────────────────────────────────────────────
{crystallized_vision.vision}

────────────────────────────────────────────────────────────────────────────────
                         EXECUTION SUMMARY
────────────────────────────────────────────────────────────────────────────────

Agents Deployed:       {N}
Parallel Groups:       {M}
Execution Time:        {duration}
Sentinel Alerts:       {count}

Agent Performance:
  Agent              Role                    Quality    Confidence
  ─────────────────────────────────────────────────────────────────
  A1                 [role]                  [0.X]      [0.X]
  A2                 [role]                  [0.X]      [0.X]
  ...

────────────────────────────────────────────────────────────────────────────────
                         HYPOTHESIS VERDICTS
────────────────────────────────────────────────────────────────────────────────

Hypothesis           Verdict         Confidence    Evidence
──────────────────────────────────────────────────────────────────
H1: [statement]      [✅/⚠️/❌]       [0.X]         [summary]
H2: [statement]      [✅/⚠️/❌]       [0.X]         [summary]
...

────────────────────────────────────────────────────────────────────────────────
                         INTEGRATED FINDINGS
────────────────────────────────────────────────────────────────────────────────

1. [Key finding with confidence level]

2. [Key finding with confidence level]

3. [Key finding with confidence level]

────────────────────────────────────────────────────────────────────────────────
                         CONFLICTS RESOLVED
────────────────────────────────────────────────────────────────────────────────

[If any conflicts:]
• Agents A1 and A3 disagreed on [X]. Resolution: [Y]. Confidence: [0.X]

[If no conflicts:]
• All agents converged. High consensus confidence.

────────────────────────────────────────────────────────────────────────────────
                         OVERALL VERDICT
────────────────────────────────────────────────────────────────────────────────

Overall Confidence: {aggregated_confidence}

[HIGH CONFIDENCE - PROCEED]
The orchestrated analysis supports [conclusion]. Multiple independent agents
converged on this finding. Recommended action: [next step].

[MEDIUM CONFIDENCE - PROCEED WITH CAUTION]
Evidence supports [conclusion] but with caveats: [caveats].
Recommended: [specific follow-up to increase confidence].

[LOW CONFIDENCE - MORE WORK NEEDED]
Agents did not converge. Key uncertainties: [list].
Recommended: [specific additional work].

[FAILED]
The orchestration could not answer the core question.
Reasons: [list]. Recommended: [pivot or restructure].

────────────────────────────────────────────────────────────────────────────────
                         NEXT ACTIONS
────────────────────────────────────────────────────────────────────────────────

Priority    Action                              Owner           Effort
─────────────────────────────────────────────────────────────────────
1           [action]                            [who]           [estimate]
2           [action]                            [who]           [estimate]
...

════════════════════════════════════════════════════════════════════════════════
      "Watson wasn't one algorithm. It was a framework for combining
       hundreds of approaches. The architecture allowed for the
       'many experts' philosophy."
                                — Dave Ferrucci
════════════════════════════════════════════════════════════════════════════════
```

---

## Arguments

$ARGUMENTS - The research objective or vision to orchestrate

---

## Example Usage

```
/ferrucci-orchestrate Validate that Socratic fine-tuning produces better
calibrated confidence than vanilla fine-tuning
```

```
/ferrucci-orchestrate Design and execute an experiment to test whether
late-layer activations predict learning breakthroughs
```

```
/ferrucci-orchestrate Build a comprehensive benchmark suite for
comparing dialogue-based training methods
```

---

## Ferrucci Principles Embodied

1. **Massive Parallelism**: Maximum concurrent agents
2. **Many Experts**: Each agent specialized for its task
3. **Pervasive Confidence**: Every output includes uncertainty
4. **Integration**: Shallow (quick) + Deep (rigorous) combined
5. **Architecture-First**: The orchestration structure enables success
6. **Consensus = Confidence**: Agreement across agents strengthens belief
7. **Honest Assessment**: Clear criteria, no wishful thinking

---

## When to Use

- Complex research questions requiring multiple analyses
- Validation efforts needing independent confirmation
- Building systems with multiple components
- Any task where parallel decomposition accelerates progress

This is how Watson was built. This is how serious research gets done.
