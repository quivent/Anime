---
description: Build a feature with hard evidence at every step — nothing advances without proof it works
argument-hint: <feature-description> [--spec file] [--strict]
---

# FORGE — Evidence-Gated Feature Development

Build a feature. Every step requires hard evidence before advancing. No step is "done" until proven.

**Input:** $ARGUMENTS

---

## THE RULE

**Nothing advances without proof.**

Not "I wrote the code" — proof it compiles.
Not "it compiles" — proof it runs.
Not "it runs" — proof it produces correct output.
Not "it produces output" — proof the output matches the spec.
Not "it matches the spec" — proof it handles edge cases.
Not "it handles edge cases" — proof it doesn't break existing functionality.

Each gate produces an EVIDENCE ARTIFACT. The artifact is a file, a screenshot, a test output, or a command transcript. Not a claim. Not a summary. The artifact itself.

---

## PHASE 1: SPEC — Define what "done" looks like BEFORE writing code

Before touching any code, produce a specification that defines success in testable terms.

```
FEATURE SPEC: {feature name}

MUST (each item becomes a gate):
  1. [Specific, testable behavior]
  2. [Specific, testable behavior]
  ...

MUST NOT:
  1. [Specific failure condition]
  2. [Specific failure condition]
  ...

EDGE CASES:
  1. [Input/condition that could break it]
  2. [Input/condition that could break it]
  ...

INTEGRATION:
  1. [Existing feature that must still work]
  2. [Existing feature that must still work]
  ...
```

If `--spec file` is provided, read the spec from that file instead of generating one.

**GATE 0 EVIDENCE:** The spec itself, written to `{cwd}/forge-spec.md`. No code is written until this exists.

---

## PHASE 2: TEST FIRST — Write the tests before the implementation

For every MUST item and EDGE CASE in the spec, write a test. The tests MUST FAIL at this point (there's no implementation yet). If any test passes before implementation, the test is wrong — it's not testing what it claims.

**GATE 1 EVIDENCE:**
- Test file(s) exist
- Test execution output showing ALL tests FAIL
- Paste the actual output — not "tests fail as expected"

```
## Gate 1 Evidence — Tests Written, All Failing
$ [test command]
[actual output showing failures]
Tests: 0 passed, N failed
```

If tests pass before implementation, STOP. The tests are testing nothing. Rewrite them.

---

## PHASE 3: IMPLEMENT — Write the minimum code to pass tests

Implement the feature. Write the minimum code needed to make the tests pass. Not the "nice" version — the version that proves the concept works.

**GATE 2 EVIDENCE:**
- Test execution output showing ALL tests PASS
- Build output showing clean compilation
- Paste actual output

```
## Gate 2 Evidence — Implementation Passes Tests
$ [build command]
[actual build output — no errors, no warnings]

$ [test command]
[actual test output — all pass]
Tests: N passed, 0 failed
```

If any test still fails, you are not past this gate. Fix the implementation or fix the test (with documented justification for why the test was wrong).

---

## PHASE 4: EDGE CASES — Prove it handles the hard inputs

For every EDGE CASE in the spec, execute the implementation with that input and capture the result.

**GATE 3 EVIDENCE:**
- For each edge case: the exact input, the exact output, and whether it matches expected behavior
- Any edge case that produces wrong output is a bug — fix it before proceeding

```
## Gate 3 Evidence — Edge Cases
Edge case 1: [input] → [output] ✓ matches expected
Edge case 2: [input] → [output] ✓ matches expected
Edge case 3: [input] → [output] ✗ BUG — expected [X], got [Y]
  → Fixed in [file:line], re-run: [output] ✓
```

---

## PHASE 5: INTEGRATION — Prove nothing is broken

Run the FULL existing test suite. Not just the new tests — everything. Every existing test must still pass. This is where regressions are caught.

**GATE 4 EVIDENCE:**
- Full test suite output
- Diff showing: previous pass count vs current pass count
- Zero regressions

```
## Gate 4 Evidence — Integration
$ [full test suite command]
[actual output]
Previously passing: N tests
Currently passing: N + M tests (M new)
Regressions: 0
```

If any existing test breaks, you are not past this gate. Fix the regression. Do not skip the test. Do not mark it as "expected failure."

---

## PHASE 6: PROVE THE MUST NOTs

For every MUST NOT in the spec, attempt to trigger the failure condition and prove it doesn't happen.

**GATE 5 EVIDENCE:**
- For each MUST NOT: the attempt to trigger it and proof it was prevented

```
## Gate 5 Evidence — MUST NOT Verification
MUST NOT 1: "Must not crash on empty input"
  → Sent empty input: [output showing graceful handling] ✓
MUST NOT 2: "Must not allow unauthenticated access"
  → Sent request without auth: [output showing 401] ✓
```

---

## PHASE 7: SHIP — Produce the deliverable

Only after all 6 gates pass, produce the final deliverable:

```
## Forge Report: {feature name}

**Spec:** forge-spec.md
**Status:** ALL GATES PASSED

### Gate Summary
| Gate | Name | Evidence | Verdict |
|------|------|----------|---------|
| 0 | Spec written | forge-spec.md | PASS |
| 1 | Tests fail pre-impl | [test output] | PASS |
| 2 | Tests pass post-impl | [test output] | PASS |
| 3 | Edge cases handled | [edge case log] | PASS |
| 4 | Integration clean | [full suite output] | PASS |
| 5 | MUST NOTs verified | [verification log] | PASS |

### Files Changed
[list of files with brief description of changes]

### Evidence Archive
[all gate evidence, preserved in forge-evidence.md]
```

---

## OPERATIONAL RULES

1. **Evidence is output, not narration.** "The tests pass" is not evidence. The test output IS evidence. Paste it. Every gate requires pasted output from actual execution.

2. **Gates are sequential.** You cannot skip to Phase 5 because "I'm confident Phase 3 works." Run Phase 3. Paste the evidence. Then proceed.

3. **Failed gates block.** If a gate fails, you fix the issue and re-run the gate. You do not proceed to the next gate with a known failure.

4. **The spec is immutable during implementation.** If you realize the spec is wrong during Phase 3-6, you STOP, update the spec (Gate 0 again), rewrite affected tests (Gate 1 again), and re-run from the beginning. The spec is not adjusted to match the implementation. The implementation is adjusted to match the spec.

5. **No untested code ships.** Every public function, every error path, every branch must be exercised by a test that was written BEFORE the implementation existed.

6. **Evidence is persisted.** All gate evidence is written to `forge-evidence.md` in the working directory. This file is the proof that the feature was built correctly.
