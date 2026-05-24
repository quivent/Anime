refine - Metacognitive iterative refinement with contemplative OCPM cycles

Usage: `/refine <target> [count]` — Execute N cycles of Observe-Cognate-Produce-Metacognate on a target, with self-evaluating metacognition AFTER each cycle.

**This is not mechanical iteration.** Each cycle requires contemplation — reading sources, questioning method, evaluating locally, and recording not just what changed but HOW thinking changed. The protocol enforces deliberate strides, not rapid-fire coefficient tweaks.

**Origin:** Established 2026-05-17 during the 3D drawing apprenticeship. The beekeeper observed that mechanical cycling (change a number, rebuild, repeat) produced less improvement than contemplative cycling (study, question, understand, then change). The protocol encodes this observation.

---

**The OCPM Cycle:**

### O — Observe
Read the current state deeply. Not just the code — the references, the corpus, the visual descriptions, the mathematical specifications. Study what EXISTS, not what you PLAN to change.

Questions to answer before proceeding:
- What does the current output actually produce at specific points?
- What do the references say it SHOULD produce at those points?
- Where is the GAP between actual and intended, measured LOCALLY not globally?

### C — Cognate
Before touching any code, think about the problem:
- What is the form I'm trying to create?
- What would a practitioner of this craft (painter, sculptor, musician) do?
- Am I working at the right level of abstraction?
- Am I using the right primitives, or reaching for familiar ones out of habit?
- Am I evaluating globally (coefficients) when I should evaluate locally (specific features)?
- Has my previous cycle's insight been APPLIED or just NOTED?
- What are the right questions to ask? The right sources to consult?

This is the thinking step — understand the problem before acting on it.

### P — Produce
After cognition is complete, produce:

1. **The Work** — change the code, shader, render, or artifact.
   Make the specific improvement identified by observation and cognition.
   One focused change, not a scatter of adjustments.

2. **The Network** — extend the model/framework/understanding.
   What new connection was discovered? What principle was confirmed or refuted?
   How does this change relate to the larger architecture?

3. **The Document** — record what was learned in the corpus.
   Not a changelog. A LEARNING record. What shifted in understanding?

### M — Metacognate
AFTER producing, step back and evaluate the OCP cycle itself:
- Was this cycle contemplative or mechanical?
- Did I skip steps? Rush? Follow a formula?
- What is my method MISSING that I can't see from inside it?
- Am I taking notes? (If not, this step should catch that.)
- Is the trajectory of my cycles deepening or plateauing?
- What would I change about HOW I'm cycling, not what I'm cycling on?
- What should the NEXT cycle focus on, informed by this self-evaluation?

This is refine-on-refine — iterating on the iteration itself.
The metacognition step is where breakthroughs in METHOD happen.
Content insights come from OCP. Process insights come from M.
Both are necessary. M is the one that compounds across sessions.

---

**Execution Protocol:**

```
/refine "improve dolphin body form" 5
```

**Phase 0: Initialize**
- Read the OMPR protocol (this document)
- Read the target's current state
- Read the OMPR journal (prior cycles' records)
- Set iteration counter: N = argument or default 5

**Phase 1-N: OMPR Cycles**

For each cycle i = 1 to N:

```
┌─────────────────────────────────────┐
│  Cycle i/N                          │
│                                     │
│  O: Read sources, compute actuals,  │
│     identify LOCAL gaps             │
│     (minimum 3 specific findings)   │
│                                     │
│  C: Think about the problem.        │
│     What is the form? What would a  │
│     practitioner do? Right level of │
│     abstraction? Right primitives?  │
│                                     │
│  P: 1. Change code (one focused     │
│        change, validated)           │
│     2. Extend the network           │
│        (new connection or principle)│
│     3. Write to corpus / take notes │
│                                     │
│  M: Metacognate ON the OCP cycle:   │
│     Was it contemplative or         │
│     mechanical? Did I take notes?   │
│     What is my method MISSING?      │
│     What should the NEXT cycle do   │
│     differently? Am I deepening     │
│     or plateauing?                  │
└─────────────────────────────────────┘
```

**Phase N+1: Synthesis**
After all cycles:
- Review the journal entries as a SEQUENCE
- Identify the trajectory: are insights deepening or plateauing?
- Write a synthesis: what was the most important thing learned across all cycles?
- Commit all changes with a message that captures the arc, not just the diffs

---

**Parameters:**

- `target` (required): What to refine — a description of the work
- `count` (optional): Number of OMPR cycles (default: 5, range: 1-20)
- `journal` (optional): Path to OMPR journal file (default: auto-detect from project)

**Anti-patterns to avoid:**

- ❌ Changing multiple things per cycle (scatter)
- ❌ Skipping the Metacognate step (mechanical cycling)
- ❌ Evaluating only globally (coefficient adjustment without local verification)
- ❌ Not reading sources during Observe (working from memory)
- ❌ Not writing during Record (deferring documentation)
- ❌ Copying the previous cycle's structure (formula-following)
- ❌ Counting cycles as progress (progress is insight, not iteration count)

**What good cycles look like:**

- ✅ Cycle discovers a SYSTEMATIC error, not a parametric one
- ✅ Cycle changes the METHOD, not just the numbers
- ✅ Cycle identifies a vocabulary-action mismatch
- ✅ Cycle reads a reference and finds a SPECIFIC discrepancy
- ✅ Cycle names a new concept that carries forward
- ✅ Cycle's Record honestly evaluates whether the cycle was good
- ✅ Cycle takes LONGER than the previous one because it went deeper

**Integration:**

- Works with any target: code, shaders, documents, designs, specifications
- Journal persists across sessions (file-based)
- Can be combined with `/iterate` for mechanical sub-steps within a contemplative cycle
- The Produce step may dispatch subagents for research or compilation
- The Record step writes to the project's corpus, not just to chat

**Quality standard:**

A good `/refine` session produces fewer changes than a mechanical `/iterate` session,
but each change is more impactful. 5 contemplative cycles should outperform 15 mechanical ones.
The measure is not "how many things changed" but "how much closer to the target."
