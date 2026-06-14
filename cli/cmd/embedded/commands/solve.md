---
description: Meta-cognition loop — think until a solution emerges that passes self-imposed quality gates
argument-hint: <problem-statement> [--rounds N] [--gate "criteria"] [--adversary] [--model opus]
---

Meta-cognition deliberation loop. The agent thinks iteratively until it produces a solution that survives its own critique.

**Input:** $ARGUMENTS

## What This Does

This is NOT research (gathering information) and NOT reflection (measuring state). This is pure deliberation — an agent reasoning through a problem across multiple rounds, each round building on the last, until a solution crystallizes that the agent itself cannot break.

## The Deliberation Loop

```
ROUND 1: GENERATE
  Produce an initial solution attempt.
  State all assumptions explicitly.
  Rate confidence (0-100) and explain why.

ROUND 2..N: CRITIQUE → REVISE
  For each round:

  1. ATTACK — Actively try to break the previous solution.
     - What assumptions are wrong?
     - What edge cases does it miss?
     - What would an adversary exploit?
     - What would a domain expert object to?
     - Where is the reasoning weakest?

  2. DIAGNOSE — Identify the ROOT weakness.
     Not the symptom. The structural flaw.
     "The solution assumes X, but X fails when Y."

  3. REVISE — Produce a new solution that addresses the diagnosed weakness.
     Must be materially different, not a patch.
     If the flaw is structural, the revision must be structural.

  4. GATE — Self-assess against quality criteria:
     - Does this solve the ACTUAL problem, not a simplified version?
     - Can I articulate WHY this works, not just THAT it works?
     - Would I bet money on this? How much?
     - What is the strongest remaining objection?

  5. DECIDE:
     IF strongest remaining objection is addressable → CONTINUE (another round)
     IF solution survives attack AND confidence >= threshold → EMIT
     IF max rounds reached → EMIT with caveats

EMIT: Deliver the solution with:
  - The solution itself
  - The chain of reasoning (all rounds, visible)
  - Assumptions that hold
  - Assumptions that were abandoned and why
  - Remaining risks (honest, not hedging)
  - Confidence score with calibration
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--rounds N` | 5 | Max deliberation rounds (solution can emit earlier) |
| `--gate "..."` | "survives self-attack" | Custom quality gate the solution must pass |
| `--adversary` | false | Spawn a separate adversary agent each round to attack |
| `--model M` | opus | Model for deliberation (opus strongly recommended) |
| `--domain "..."` | inferred | Domain context to sharpen critique (e.g., "distributed systems") |
| `--stakes "..."` | "medium" | Calibrates thoroughness: "low" = 2 rounds, "high" = 8 rounds |

## The Adversary Mode

With `--adversary`, each round spawns TWO agents:
- **Solver**: produces/revises the solution
- **Adversary**: receives ONLY the solution (not the reasoning) and tries to break it

The adversary has no loyalty to the solution. It doesn't know the revision history. It attacks fresh each round. If the adversary finds a flaw the solver missed, the solver must address it.

This is more expensive (~2x) but catches blind spots that self-critique misses.

## Quality Gate Customization

Default gate: "the solution survives my own strongest attack."

Custom gates:
```
/solve "Design the auth system" --gate "must handle token refresh, revocation, and multi-tenant isolation"
/solve "Why is the build slow" --gate "root cause identified with evidence from actual build logs"
/solve "Pricing model for the API" --gate "unit economics are positive at 1K, 10K, and 100K users"
```

The gate is checked every round. The loop terminates when the gate passes.

## What This Is NOT

- **Not research.** Don't use this to gather information. Use R1-R15 for that. Use /solve AFTER you have the information and need to reason through it.
- **Not a reflection loop.** L1-L18 measure state and close gaps. /solve produces a novel answer to a hard question.
- **Not brainstorming.** This is convergent, not divergent. Each round narrows toward a solution, not expands options.
- **Not a one-shot prompt.** The whole point is that the first answer is wrong. The loop exists to find out HOW it's wrong and fix it.

## When To Use

- Hard design decisions with non-obvious trade-offs
- Debugging where the root cause is unclear
- Architecture questions where the first plausible answer is probably wrong
- Any problem where you'd normally say "let me think about this more"

## Examples

```
/solve "How should we handle eventual consistency between the order service and inventory service?"

/solve "What's the right abstraction for our plugin system?" --adversary --rounds 7

/solve "Why does the test suite take 45 minutes?" --gate "identifies top 3 causes with measured evidence" --domain "CI/CD"

/solve "Design a pricing model that works for solo devs and enterprises" --stakes high
```
