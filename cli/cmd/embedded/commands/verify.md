---
description: Verify an artifact against acceptance criteria — execute it, capture evidence, evaluate independently
argument-hint: <artifact-path-or-description> [--spec "criteria"] [--type code|ui|gpu|doc] [--strict]
---

Verify that an artifact actually works. Not "does it compile" — does it DO what it claims.

**Input:** $ARGUMENTS

## The Verification Pipeline

This command implements the SPEC → EXECUTE → EVALUATE chain. It is the antidote to report-as-verification.

### Step 1: IDENTIFY the artifact and its type

Determine what we're verifying and how to run it:

| Type | Artifact | How to Execute | Evidence Captured |
|------|----------|---------------|-------------------|
| **code** | Source file, module, service | Build + run + test suite | Test output, exit codes, stdout/stderr |
| **ui** | Web page, app screen, component | Launch + screenshot | Screenshots of key states |
| **gpu** | Lithos composition, shader, Metal | Compile + render frame | Frame capture, GPU timings |
| **doc** | Markdown, API docs, README | Render + check links + run examples | Rendered output, link check results, example execution |
| **api** | Endpoint, service, CLI tool | Call with test inputs | Response payloads, status codes, timing |

If `--type` is not specified, infer from the file extension and context.

### Step 2: ESTABLISH acceptance criteria

If `--spec` is provided, use it directly.

If not, generate acceptance criteria by reading the artifact and asking:
- What does this claim to do? (extract from comments, docstrings, README)
- What are the minimum behavioral expectations?
- What would "broken" look like?

Produce a structured spec:
```
ACCEPTANCE SPEC for {artifact}:
  MUST: [things that must be true for PASS]
  MUST NOT: [things that indicate failure]
  EVIDENCE REQUIRED: [what to capture]
```

### Step 3: EXECUTE the artifact

Run it. Not read it — RUN it.

**For code:**
```
1. Build (compile/install deps)
2. Run test suite if it exists
3. If no tests: execute the entry point with reasonable inputs
4. Capture: stdout, stderr, exit code, timing
```

**For UI:**
```
1. Launch the dev server or open the file
2. Navigate to key screens/states
3. Screenshot each state
4. Check for: blank screens, error overlays, missing elements
```

**For GPU/Lithos:**
```
1. Compile the composition
2. Render a frame (or N frames for animation)
3. Capture the frame(s)
4. Check for: solid black, solid white, NaN artifacts, missing geometry
```

**For docs:**
```
1. Render markdown to HTML (or check in browser)
2. Verify all internal links resolve
3. Run every code example in the doc
4. Check for: broken links, failing examples, stale references
```

### Step 4: EVALUATE — independent judgment

Spawn a separate evaluation (do NOT self-evaluate from the same context that built it).

The evaluator receives:
- The acceptance spec (from Step 2)
- The evidence artifacts (from Step 3)
- NOT the builder's explanation or reasoning

The evaluator returns one of:
- **PASS**: All MUST criteria met, no MUST NOT violations, evidence is clean
- **FAIL**: One or more MUST criteria not met, with specific failure identified
- **AMBIGUOUS**: Evidence is inconclusive — needs human eyes

### Step 5: REPORT with evidence

```
## Verification Report: {artifact}

**Verdict:** PASS | FAIL | AMBIGUOUS
**Type:** {artifact type}
**Executed:** {timestamp}

### Acceptance Spec
{the criteria}

### Evidence
{test output / screenshots / frame captures / link check results}

### Evaluation
{evaluator's reasoning}

### Failures (if any)
{specific failures with evidence}

### Action Required (if FAIL or AMBIGUOUS)
{what needs to happen next}
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--spec "..."` | auto-generated | Explicit acceptance criteria |
| `--type T` | inferred | Artifact type: code, ui, gpu, doc, api |
| `--strict` | false | AMBIGUOUS is treated as FAIL |
| `--fix` | false | If FAIL, attempt to fix and re-verify (max 2 attempts) |
| `--compare PATH` | none | Golden reference to diff against |

## Examples

```
/verify src/auth/login.ts --spec "login succeeds with valid creds, fails with invalid, rate-limits after 5 attempts"

/verify ~/project/index.html --type ui

/verify ~/lithos/compositions/galaxy.ls --type gpu --spec "renders non-black frame with >100 visible stars"

/verify README.md --type doc --strict

/verify src/api/routes.ts --type api --spec "GET /health returns 200, POST /users validates email format"
```

## Integration with /orchestrate

When used inside `/orchestrate`, verification happens automatically after each BUILD step. The orchestrator calls `/verify` on every artifact before marking the task as complete. Only PASS artifacts count toward completion.
