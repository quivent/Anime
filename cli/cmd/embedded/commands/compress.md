# /compress - Find Minimum Description Length

*"It is impossible to compress data such that the average code rate is less than the Shannon entropy of the source, without losing information."* — Source Coding Theorem

---

## PURPOSE

Find the shortest possible description of a system that preserves all essential information. This is not about saving bytes - it's about finding the true structure.

---

## THE INSIGHT

Every system has a minimum description length - the shortest program that produces it. This is related to Kolmogorov complexity.

When your description is longer than necessary:
- You haven't found the right abstraction
- There's redundancy you haven't factored out
- You're encoding noise as if it were signal

When your description approaches minimum:
- The structure becomes clear
- Patterns become visible
- The essential logic stands alone

---

## PROTOCOL

**Arguments**: $ARGUMENTS (code, system, or concept to compress)

### Phase 1: Identify Redundancy

What repeats?
- Literal repetition (copy-paste code)
- Structural repetition (same pattern, different data)
- Conceptual repetition (same idea, different expression)

### Phase 2: Factor Out Patterns

For each redundancy:
- Can it become a function? (behavioral pattern)
- Can it become a data structure? (structural pattern)
- Can it become a convention? (implicit pattern)
- Can it become a generator? (meta-pattern)

### Phase 3: Find the Generating Function

What's the shortest program that produces this system?

Often a complex system is the output of a simple rule applied repeatedly. Find that rule.

### Phase 4: Verify Losslessness

After compression:
- Can you reconstruct the original exactly?
- Have you lost any essential information?
- Is the compressed form actually simpler to understand?

---

## OUTPUT FORMAT

```
COMPRESSION ANALYSIS: [target]

CURRENT SIZE: [lines/files/complexity measure]

REDUNDANCIES IDENTIFIED:
- [pattern 1]: occurs [N] times, [X] lines each
- [pattern 2]: occurs [M] times, [Y] lines each
...

COMPRESSION OPPORTUNITIES:

1. [redundancy] → [compressed form]
   Savings: [amount]

2. [redundancy] → [compressed form]
   Savings: [amount]
...

GENERATING FUNCTION:
[The simplest rule/program that produces this system]

THEORETICAL MINIMUM: [estimate]
ACHIEVABLE MINIMUM: [practical estimate]

COMPRESSED DESCRIPTION:
[The shortest accurate description of what this system does]
```

---

## EXAMPLES

### Compressing a CRUD API

**Before**: 4 files × 5 endpoints × 50 lines = 1000 lines
**Pattern**: Every endpoint follows the same structure
**Generating function**: `for each entity: generate(CRUD_template, entity_schema)`
**Compressed**: 1 template + 4 schemas = ~100 lines
**Compression ratio**: 10:1

### Compressing a State Machine

**Before**: Giant switch statement with 50 cases
**Pattern**: States, transitions, actions
**Generating function**: `state_machine(states, transitions, actions)`
**Compressed**: Declarative state/transition table
**Compression ratio**: 5:1

### Compressing Documentation

**Before**: 100 pages of prose
**Pattern**: Same information repeated at different abstraction levels
**Generating function**: One source of truth + views
**Compressed**: Core spec + generated docs
**Compression ratio**: 4:1

---

## LIMITS

My source coding theorem proves: you cannot compress below the entropy of the source without losing information.

But most systems are far above their entropy. The redundancy is there - you just haven't found it yet.

*Compression is not about making things shorter. It's about finding the true structure that was always there.*

---

## THE TEST

If your compressed description is harder to understand than the original, you've done it wrong.

True compression reveals structure. It makes things clearer, not more cryptic.

*The minimum description length is often the most illuminating.*
