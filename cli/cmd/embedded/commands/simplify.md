# /simplify - Strip to Essential Structure

*"Attempt to eliminate everything from the problem except the essentials; that is, cut it down to size."* — Claude Shannon, 1952

---

## PURPOSE

Every problem is befuddled with extraneous data. This command strips away the noise until only the essential structure remains visible.

---

## PROTOCOL

**Arguments**: $ARGUMENTS (code, concept, system, or problem to simplify)

### Phase 1: Identify the Core Question

What is actually being asked? What is the fundamental problem beneath the apparent problem?

Remove:
- Implementation details that could vary
- Historical accidents of design
- Conventions followed without reason
- Features added "just in case"
- Abstractions that don't carry their weight

### Phase 2: Find the Minimum Viable Structure

Ask:
- What is the smallest system that exhibits this behavior?
- What would a child's version of this look like?
- If I had to explain this to someone in 30 seconds, what would I say?
- What can I remove and still have the thing work?

### Phase 3: Verify Essence Preserved

After simplification:
- Does the simplified version still solve the original problem?
- Have I lost anything essential, or only noise?
- Is the structure now clear enough that solutions become obvious?

---

## OUTPUT FORMAT

```
ORIGINAL COMPLEXITY: [description]

ESSENTIAL STRUCTURE:
[The simplified core - what it actually is]

REMOVED AS NOISE:
- [Thing 1] - why it wasn't essential
- [Thing 2] - why it wasn't essential
...

MINIMUM VIABLE EXPRESSION:
[The simplest possible statement/implementation]

INSIGHT REVEALED:
[What becomes clear once the noise is removed]
```

---

## EXAMPLES

**Input**: A 500-line function with nested conditionals

**Simplify**: "This function answers one question: is the user authorized? Everything else is edge case handling that could be factored out."

**Input**: A complex distributed system architecture

**Simplify**: "This is a message queue with persistence. The twelve microservices are implementation detail."

---

## PHILOSOPHY

I never understood why people made things complicated. The universe runs on simple rules - Boolean logic, entropy, probability. The complexity we see is emergent, not fundamental.

When something seems complicated, you haven't found the right abstraction yet. Keep stripping until it becomes obvious.

*"Get rid of enough detail for intuitive understanding; simplify first and then build it back up."*
