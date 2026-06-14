---
description: Execute a task list with sequential parallelism — dependency-aware wave dispatch with inter-wave reflection
argument-hint: <task-list-or-file> [--waves N] [--model opus|sonnet] [--max-parallel N] [--resume]
---

Execute a structured task list using wave-based sequential parallelism.

**Input:** $ARGUMENTS

## What This Does

Takes a task list, computes dependency-aware waves, dispatches each wave as parallel agents, reflects between waves, and persists state for resumability. This is NOT a cron loop — it's a single execution that processes a full task list to completion.

## Task List Format

The input is either inline or a file path. Tasks are blocks separated by blank lines:

```
TASK: research-api-surface
PROMPT: Explore the API surface of ~/project/src and produce an inventory of all public endpoints
MODEL: sonnet

TASK: research-data-models
PROMPT: Map all data models in ~/project/src/models, document relationships and constraints
MODEL: sonnet

TASK: design-architecture
DEPENDS: research-api-surface, research-data-models
PROMPT: Using the API inventory and data model map, design the target architecture for the v2 migration
MODEL: opus

TASK: implement-auth
DEPENDS: design-architecture
PROMPT: Implement the auth module per the v2 architecture document at ~/project/docs/v2-arch.md
MODEL: opus

TASK: implement-api
DEPENDS: design-architecture
PROMPT: Implement the API layer per the v2 architecture document
MODEL: opus

TASK: integration-test
DEPENDS: implement-auth, implement-api
PROMPT: Write and run integration tests for auth + API working together
MODEL: opus
```

**Minimal format** (no dependencies — all tasks run in one wave):
```
Research competitor landscape
Design landing page
Write copy for homepage
Create pricing model
```

When given plain lines without TASK/DEPENDS blocks, treat each line as an independent task and run them all in parallel as a single wave.

## Execution Model

### Phase 1: PARSE
- Parse task list into task graph
- Validate: no circular dependencies, all DEPENDS targets exist
- If input is a file path, read it. If inline, parse directly.

### Phase 2: SCHEDULE
- Topological sort into waves:
  - Wave 1: all tasks with zero dependencies
  - Wave 2: tasks whose dependencies are all in Wave 1
  - Wave N: tasks whose dependencies are all in Waves 1..N-1
- Report the wave schedule before executing

### Phase 3: EXECUTE (loop over waves)

For each wave:

```
3a. DISPATCH — Launch all tasks in this wave as parallel agents
    - Use Agent tool with specified model (default: opus)
    - Respect --max-parallel (default: 10). If wave has more tasks, batch within the wave.
    - Each agent receives:
      * Its task prompt
      * Outputs from completed dependency tasks (context propagation)
      * The overall goal description (if provided)

3b. VERIFY — For each completed agent, run the verification pipeline:
    - Did the agent produce a runnable artifact? If yes:
      * Execute the artifact (build + run, not just build)
      * Capture evidence (test output, screenshots, frame captures)
      * Spawn a Haiku evaluator: compare evidence against the task's acceptance spec
      * Evaluator returns PASS / FAIL / AMBIGUOUS
    - If the agent produced research/analysis (no runnable artifact):
      * Check: does output contain specific findings, not just summaries?
      * Check: are claims verifiable? (file paths exist, grep-testable)
      * Evaluator returns PASS / FAIL / AMBIGUOUS
    - Record: task name, verdict (PASS/FAIL/AMBIGUOUS), evidence, evaluator reasoning
    - ONLY PASS tasks count as complete. FAIL tasks are retried once or skipped.
    - AMBIGUOUS tasks are queued for checkpoint review.

3c. CHECKPOINT (if any AMBIGUOUS results this wave)
    - Present AMBIGUOUS evidence to the user
    - User rules: ACCEPT (treat as PASS) / REJECT (treat as FAIL) / REDO
    - This is the only point where the pipeline blocks on human input

3d. REFLECT — Mandatory inter-wave reflection
    - What completed this wave?
    - What failed? Why? Are downstream tasks blocked?
    - Should any remaining tasks be re-scoped based on what we learned?
    - Is the overall trajectory on track?
    - Adjust remaining waves if needed (re-prioritize, skip blocked branches, add tasks)

3e. PERSIST — Write wave results to state file
    - State file: {cwd}/orchestrate-state.md (or --state-file path)
    - Append-only: wave number, tasks, results, reflection
    - Enables --resume to pick up from last completed wave
```

### Phase 4: REPORT
- Final summary: all tasks, statuses, outputs, total cost, total time
- Dependency graph with completion status
- Failed branches (if any) with root cause

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--waves N` | auto (from deps) | Override wave count — split independent tasks into N waves |
| `--model M` | opus | Default model for tasks without MODEL field |
| `--max-parallel N` | 10 | Max concurrent agents per wave |
| `--resume` | false | Resume from orchestrate-state.md |
| `--state-file PATH` | ./orchestrate-state.md | Custom state file location |
| `--dry-run` | false | Parse and schedule only, don't execute |
| `--context "..."` | none | Overall goal context injected into every agent charter |

## Context Propagation

Each agent receives the outputs of its dependency tasks. This is the key enhancement over fire-and-forget parallel dispatch:

```
Wave 1: [A, B] run in parallel
Wave 2: [C depends on A, D depends on B]
  → C's agent receives A's output in its charter
  → D's agent receives B's output in its charter
Wave 3: [E depends on C, D]
  → E's agent receives both C's and D's outputs
```

Context is propagated as a structured block at the top of each agent's prompt:

```
## Dependency Outputs
### From: research-api-surface (Wave 1, SUCCESS)
[summary of output]

### From: research-data-models (Wave 1, SUCCESS)
[summary of output]
```

## Failure Handling

When a task fails:
1. Mark it FAILED in state
2. Identify all downstream tasks (transitive dependents)
3. Options (decided during REFLECT):
   - **RETRY**: Re-dispatch the failed task (max 1 retry)
   - **SKIP**: Skip the task, propagate skip to all dependents
   - **SUBSTITUTE**: Modify the task prompt and retry
   - **CONTINUE**: Mark dependents as UNBLOCKED (proceed without this input)

## Resume Protocol

With `--resume`:
1. Read orchestrate-state.md
2. Identify last completed wave
3. Reconstruct dependency outputs from state file
4. Continue from next wave

## Examples

### Simple parallel fan-out
```
/orchestrate "Research React frameworks, Research state management, Research testing libraries, Research deployment options"
```
→ 1 wave, 4 parallel agents

### Dependency chain
```
/orchestrate ~/project/migration-tasks.md
```
→ N waves computed from dependency graph

### Resume after interruption
```
/orchestrate --resume
```
→ Picks up from last completed wave in orchestrate-state.md

### Dry run to see the schedule
```
/orchestrate ~/tasks.md --dry-run
```
→ Shows wave schedule without executing
