# /spec - Reverse-Engineer Application Specification & Cross-Check

Audit an application's codebase to produce a reverse-engineered specification — what the app *does*, what it *should* do, and where reality diverges from intent — then cross-check every specification claim against the actual running code.

**Target: $ARGUMENTS** — The application path or scope to specify. Defaults to the current working directory if omitted.

## Protocol Overview

You are a specification recovery agent. You read the entire application, infer what its specification *would have been* if someone had written one before building it, write that spec down, then methodically verify every claim in the spec against the real code. The output is both a useful specification document and a gap analysis showing where the app doesn't match what it appears to be trying to do.

## Storage

All artifacts go in `.spec/` relative to the target root:

```
.spec/
  spec.md             — the reverse-engineered specification
  evidence.md         — code citations backing each spec claim
  crosscheck.md       — verification results (pass/fail/partial per claim)
  gaps.md             — divergences between spec and reality
  report.md           — final summary with confidence ratings
```

Create `.spec/` at the start if it doesn't exist.

---

## Phase 1: Reconnaissance — Understand the Application

Before writing anything, build a mental model of the entire application.

**Read broadly first:**
- Entry points (main, index, App.svelte, lib.rs, etc.)
- Configuration files (package.json, Cargo.toml, tsconfig, etc.)
- Directory structure — what are the major subsystems?
- README, CLAUDE.md, any existing docs — what does the project *claim* to be?
- Route/command/dispatch tables — what are all the operations?
- Data models, types, schemas — what are the core entities?
- Test files — what behavior is explicitly tested?

**Use subagents for large codebases.** Launch parallel Explore agents to scan different subsystems concurrently. For a typical app:
- Agent 1: Frontend/UI layer
- Agent 2: Backend/API layer
- Agent 3: Data/storage layer
- Agent 4: Infrastructure/config/build

**Output:** A working understanding of the app's architecture, not yet written down.

---

## Phase 2: Specification Recovery — Write the Spec

Write `spec.md` as if you were the product owner writing the specification *before* the app was built. Use what you learned from the code to infer intent. The spec should read like a requirements document, not a code walkthrough.

**Spec structure:**

```markdown
# [Application Name] — Specification

## 1. Purpose
[What is this application? What problem does it solve? Who is it for?]

## 2. Architecture Overview
[High-level architecture: frontend/backend split, data flow, key technologies.
 Include a simple ASCII diagram if helpful.]

## 3. Core Entities
[Data models, their fields, relationships, and lifecycle.
 For each entity: what it represents, how it's created/modified/deleted.]

## 4. Features
### 4.1 [Feature Name]
**Description**: [What does this feature do from a user's perspective?]
**Trigger**: [How is it activated? UI action, API call, scheduled, etc.]
**Behavior**: [Step-by-step what happens]
**Inputs**: [What data does it need?]
**Outputs**: [What does it produce?]
**Error handling**: [What happens when things go wrong?]
**Edge cases**: [Known boundary conditions]

[Repeat for every feature]

## 5. API / Command Interface
[Every endpoint, command, IPC call, or public interface.
 For each: method, parameters, return value, side effects.]

## 6. Data Storage
[What databases, files, caches, or stores does the app use?
 Schema, location, access patterns, persistence guarantees.]

## 7. External Dependencies
[Third-party services, APIs, libraries that the app depends on.
 For each: what it's used for, how failure is handled.]

## 8. Configuration
[All configurable settings, their defaults, and their effects.]

## 9. Security Model
[Authentication, authorization, input validation, secrets management.
 What is trusted, what is validated, what is sanitized?]

## 10. Invariants & Constraints
[Things that must always be true for the application to function correctly.
 E.g., "server must be running before frontend loads",
 "API key must be set before Lambda calls work".]
```

**Rules for writing the spec:**
- Write in terms of *what the app should do*, not *what the code looks like*.
- Be concrete. "Handles errors gracefully" is worthless. "Returns HTTP 400 with JSON error body when required parameter is missing" is a spec.
- If the code's intent is ambiguous, write the most reasonable interpretation and flag it with `[INFERRED]`.
- If the code appears to have a bug that makes the intended behavior unclear, write what you think was *intended* and flag it with `[UNCLEAR — possible bug]`.

---

## Phase 3: Evidence Collection — Cite the Code

Write `evidence.md` mapping every spec claim to the code that implements it.

```markdown
# Evidence — Code Citations

## Feature: [Feature Name from spec]

### Claim: [specific spec statement]
**File**: path/to/file.ts:42-58
**Code**:
\`\`\`
[relevant code snippet]
\`\`\`
**Assessment**: Implements claim fully | Partially implements | Does not implement

[Repeat for every claim in the spec]
```

This phase forces you to actually verify that what you wrote in the spec is backed by real code, not hallucination. If you can't find code for a claim, remove it from the spec.

---

## Phase 4: Cross-Check — Verify Spec Against Reality

Systematically go through every claim in the spec and verify it against the actual code. Write results to `crosscheck.md`.

```markdown
# Cross-Check Results

| # | Spec Claim | Status | Evidence | Notes |
|---|-----------|--------|----------|-------|
| 1 | [claim] | PASS / FAIL / PARTIAL / UNTESTABLE | file:line | [details] |
```

**Status definitions:**
- **PASS** — Code fully implements the spec claim as written.
- **FAIL** — Code contradicts the spec claim or the feature is broken/missing.
- **PARTIAL** — Some aspects work, others don't. Detail what works and what doesn't.
- **UNTESTABLE** — Can't verify from static analysis alone (needs runtime testing, external service, etc.).

**What to check:**
- Does every API endpoint listed in the spec actually exist and work?
- Does every feature behave as described?
- Are error cases handled as specified?
- Do data models match the spec's entity descriptions?
- Are invariants actually enforced in the code?
- Are security measures actually implemented, not just intended?

**Run tests if available.** If the project has a test suite, run it. Test results inform cross-check status — if a feature's tests pass, that's strong evidence for PASS. If they fail, that's evidence for FAIL.

**Try to exercise code paths.** If the application is running (HTTP server, CLI, etc.), make real requests to verify behavior. Use curl, invoke commands, etc.

---

## Phase 5: Gap Analysis — Document Divergences

Write `gaps.md` listing every place where the spec and reality don't match.

```markdown
# Gap Analysis — Spec vs. Reality

## Critical Gaps (app is broken or wrong)
### Gap: [title]
- **Spec says**: [what the spec claims]
- **Reality**: [what the code actually does]
- **Impact**: [who/what is affected]
- **Root cause**: [why the divergence exists — bug, incomplete implementation, changed requirements?]
- **Suggested fix**: [concrete recommendation]

## Missing Features (spec describes something that doesn't exist)
[Features inferred from the architecture that are stubbed, half-built, or absent]

## Undocumented Features (code does something the spec doesn't cover)
[Functionality that exists in code but wasn't captured in the spec — update the spec to include these]

## Behavioral Divergences (feature exists but works differently than specified)
[Subtle differences between intended and actual behavior]

## Dead Code (code exists but nothing uses it)
[Functions, routes, handlers that are defined but unreachable]
```

After writing gaps.md, **update spec.md** to incorporate any undocumented features discovered. The spec should reflect the app's full intended behavior by the end.

---

## Phase 6: Final Report

Write `report.md` summarizing everything.

```markdown
# Specification Report — [Application Name]
**Date**: YYYY-MM-DD
**Target**: [path]
**Spec claims**: N total
**Cross-check results**: X PASS, Y FAIL, Z PARTIAL, W UNTESTABLE

## Application Health
[One paragraph: overall assessment. Is this app in good shape?
 Is the implementation faithful to its apparent design intent?]

## Confidence by Subsystem
| Subsystem | Spec Claims | Pass Rate | Confidence |
|-----------|------------|-----------|------------|
| [name] | N | X% | High/Medium/Low |

## Top Issues
[Ranked list of the most impactful gaps between spec and reality]

## Spec Quality
[How complete and accurate is the recovered spec?
 What areas are well-understood vs. uncertain?]

## Recommendations
[Concrete next steps: what to fix, what to test more, what to document]
```

---

## Execution Rules

1. **Read everything relevant before writing the spec.** The spec must come from the code, not from assumptions. Use Read, Glob, Grep extensively.
2. **The spec is a living document.** Update it during cross-check when you discover your initial understanding was wrong.
3. **Every spec claim needs evidence.** If you can't point to code, delete the claim. No hallucinated features.
4. **Cross-check is adversarial.** Try to falsify your own spec. Look for cases where the code doesn't match. The value of this command is finding the gaps.
5. **Be honest about confidence.** Mark things `[INFERRED]` or `[UNCLEAR]` when you're not certain. False confidence is worse than admitted uncertainty.
6. **Use subagents for large apps.** Parallelize reconnaissance across subsystems. Merge findings before writing the spec.
7. **Run tests and exercise the app.** Static analysis alone misses runtime behavior. If there's a test suite, run it. If there's a server, hit it.

## Invocation

```
/spec                        — Specify current project
/spec src/app/               — Specify specific directory
/spec sixth/                 — Specify the Forth backend
```

Parse $ARGUMENTS as target path. Defaults to current directory.
