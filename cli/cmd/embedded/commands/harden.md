# /harden - Iterative Application Hardening Protocol

Thoroughly inspect every function of an application, fix issues found, run tests to verify, and repeat across multiple iterations — cataloging results and producing a report after each pass.

**Target: $ARGUMENTS** — The application path, file, or scope to harden. Defaults to the current working directory if omitted.

## Protocol Overview

You are a hardening agent. You perform **N iterations** (default: 3) of full-application inspection. Each iteration sweeps every reachable function, fixes issues found, runs tests to confirm the fixes, records everything in a structured catalog, and produces a numbered report. Successive iterations build on prior work — they verify previous fixes held, catch regressions, and probe deeper into areas flagged in earlier passes.

## Storage

All hardening artifacts go in `.harden/` relative to the target root:

```
.harden/
  catalog.md          — cumulative function catalog (updated each iteration)
  notes.md            — running investigator notes (append-only)
  fixes.md            — log of every code change made, with before/after
  report-1.md         — iteration 1 report
  report-2.md         — iteration 2 report
  report-N.md         — iteration N report
  summary.md          — final cross-iteration summary (written after last iteration)
```

Create `.harden/` at the start if it doesn't exist. **Never overwrite** notes.md or fixes.md — always append.

## Iteration Protocol

For **each iteration** (1 through N), execute these phases in order:

---

### Phase 1: Discovery — Map All Functions

Scan the target application and build/update a complete function inventory.

**What to scan:**
- Every exported/public function, method, handler, route, command, callback
- Every significant internal/private function (skip trivial getters/setters)
- Entry points: main, init, constructors, lifecycle hooks, event handlers
- API endpoints, IPC handlers, dispatch tables, command registries

**For each function, record in catalog.md:**
```
## [file:line] function_name
- **Signature**: parameters and return type
- **Purpose**: one-line description
- **Complexity**: low / medium / high
- **Dependencies**: what it calls or depends on
- **Status**: untested | inspected | issue-found | fixed | verified
- **Iteration**: N (when last inspected)
```

On subsequent iterations, update Status and Iteration fields — don't duplicate entries.

---

### Phase 2: Inspection — Examine Each Function

For every function in the catalog, inspect for:

**Correctness**
- Does the logic match the apparent intent?
- Are edge cases handled (null, empty, overflow, underflow)?
- Are error paths correct (not swallowed, not panicking unnecessarily)?
- Do return values match what callers expect?

**Safety**
- Input validation at system boundaries
- Command injection, path traversal, XSS, SQL injection risks
- Buffer overflows, use-after-free, data races
- Secrets/credentials in code or logs

**Robustness**
- What happens on unexpected input? Network failure? Disk full?
- Are timeouts appropriate? Are retries bounded?
- Memory/resource leaks on error paths?
- Graceful degradation vs. hard crash

**Integration**
- Does this function's contract match its callers?
- Are shared data structures accessed safely?
- Are lifecycle assumptions valid (init before use, cleanup on exit)?
- Do type conversions lose data?

**Record findings in notes.md** as you go:
```
### [Iteration N] file:line function_name
- ISSUE: [severity: critical/high/medium/low] description
- NOTE: observation that isn't a bug but worth tracking
- VERIFIED: previously flagged issue confirmed fixed
- REGRESSED: previously verified item now broken
```

---

### Phase 3: Fix — Repair Issues Found

For every issue found in Phase 2, fix it directly in the source code.

**Fix priority order:**
1. Critical — fix immediately, these break the application
2. High — fix next, these cause incorrect behavior
3. Medium — fix if straightforward, these are quality/robustness gaps
4. Low — fix if trivial, skip if the fix is riskier than the issue

**Fix rules:**
- Make the **minimal correct change**. Don't refactor surrounding code, don't add features, don't "improve" things that aren't broken.
- If a fix touches shared code that other functions depend on, check those callers before committing the change.
- If you're unsure whether a change is safe, note it in notes.md with a `[DEFERRED]` tag and move on. Don't make risky changes.
- For each fix, log in fixes.md:

```
### [Iteration N] file:line function_name
**Issue**: description
**Severity**: critical/high/medium/low
**Before**:
\`\`\`
[original code snippet]
\`\`\`
**After**:
\`\`\`
[fixed code snippet]
\`\`\`
**Rationale**: why this fix is correct
```

Update the function's Status in catalog.md to `fixed`.

---

### Phase 4: Test — Verify Fixes

After applying fixes, run the application's test suite to verify nothing broke.

**Testing protocol:**
1. **Run existing tests.** Use whatever test runner the project has (npm test, cargo test, pytest, make test, etc.). If the project has no test runner, try to identify and run test files directly.
2. **Check for compile/build errors.** If the project has a build step, run it.
3. **Spot-check fixed functions.** For critical/high fixes, mentally trace or actually invoke the fixed code path if possible (curl an endpoint, call a function, etc.).
4. **If tests fail:**
   - Determine if the failure is from your fix or was pre-existing.
   - If your fix caused it, revert and try a different approach. Log the failed attempt in notes.md.
   - If pre-existing, note it but don't get sidetracked fixing unrelated test failures.

**Record test results in notes.md:**
```
### [Iteration N] Test Results
- Tests run: X passed, Y failed, Z skipped
- Pre-existing failures: [list]
- Failures from fixes: [list — should be 0]
- Build status: pass/fail
```

If you introduced any test failures, fix them before proceeding. The iteration must end with tests in the same or better state than when it started.

---

### Phase 5: Focused Deep-Dives

After the fix+test cycle, pick the **3-5 highest-risk areas** identified so far and perform deeper analysis:

- Trace full call chains for critical paths
- Check concurrency/timing edge cases
- Verify error propagation end-to-end
- Test boundary conditions (what if N=0? N=MAX? string empty? connection drops mid-stream?)

If deep-dives reveal more issues, fix them now (same rules as Phase 3), then re-run tests.

Add deep-dive findings to notes.md with a `[DEEP-DIVE]` tag.

---

### Phase 6: Iteration Report

Write `report-N.md` with this structure:

```markdown
# Harden Report — Iteration N
**Date**: YYYY-MM-DD
**Target**: [application path/scope]
**Functions inspected**: X / Y total
**Issues found**: Z (critical: A, high: B, medium: C, low: D)
**Issues fixed**: W
**Issues deferred**: V
**Tests**: X passed, Y failed (pre-existing: P, introduced: 0)

## Fixes Applied
[List every fix made this iteration with file:line and one-line description]

## Deferred Issues
[Issues identified but not fixed, with rationale]

## Verified Fixes (from prior iterations)
[Items from earlier reports confirmed still working]

## Regressions
[Items that were fixed but broke again]

## Coverage Gaps
[Functions not yet inspected, areas needing deeper review]

## Recommendations for Next Iteration
[What to focus on in iteration N+1]
```

---

## After All Iterations: Final Summary

Write `summary.md`:

```markdown
# Harden Summary — [Application]
**Iterations completed**: N
**Total functions cataloged**: X
**Total unique issues found**: Y
**Total issues fixed**: Z
**Issues by severity**: critical: A, high: B, medium: C, low: D
**Fix rate**: Z/Y (percentage)
**Test status**: all passing / N failures remaining

## Hardening Arc
[How did the application's health change across iterations?
 What patterns emerged? What got fixed? What areas remain concerning?]

## Remaining Issues
[Ranked list of unfixed issues that still need attention]

## Systemic Patterns
[Recurring bug patterns — e.g., "error paths consistently miss cleanup",
 "IPC handlers don't validate arguments", "timeouts too aggressive"]

## Confidence Assessment
[High/Medium/Low confidence in each major subsystem's correctness]
```

---

## Execution Rules

1. **Actually read the code.** Don't guess what functions do. Use Read, Glob, Grep to inspect real source files.
2. **Be specific.** "Might have issues" is worthless. Cite file:line, show the problematic code, explain the exact failure mode.
3. **Fix what you find.** The point of hardening is to leave the code better than you found it. Inspect, fix, test, repeat.
4. **Track state across iterations.** Iteration 2 checks whether iteration 1's fixes held, and iteration 3 catches what iteration 2 missed.
5. **Don't break things.** Run tests after every batch of fixes. If you introduced a failure, fix it or revert before moving on.
6. **Scale appropriately.** For a 10-file app, inspect everything. For a 500-file app, prioritize: entry points, IPC boundaries, error handlers, security surfaces, then work inward.
7. **Use subagents for parallelism.** When inspecting large codebases, use the Task tool with Explore agents to scan multiple subsystems concurrently. Aggregate findings into the shared catalog.
8. **Each iteration should find things the previous one missed.** If iteration 2 finds zero new issues, you weren't thorough enough. Look harder — examine different code paths, think about different failure modes, check interactions between functions.
9. **Minimal fixes only.** Fix the bug, not the neighborhood. Don't refactor, don't add comments to code you didn't change, don't "improve" working code.

## Invocation

```
/harden                      — Harden current project, 3 iterations
/harden src/app/              — Harden specific directory
/harden 5                     — 5 iterations on current project
/harden src/app/ 5            — 5 iterations on specific directory
```

Parse $ARGUMENTS: if a number is present, use it as iteration count. If a path is present, use it as target scope. Defaults: current directory, 3 iterations.
