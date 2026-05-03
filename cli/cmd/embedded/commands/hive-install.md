# /hive-install - Install the Hive Protocol into Any Project

Install the autonomy hive coordination protocol into the current working directory. This transforms any project into a hive-aware codebase where Claude Code sessions know how to orchestrate work through charters, scopes, and workers.

**Argument:** $ARGUMENTS (optional — a one-sentence description of the project's primary goal)

---

## What this does

1. Creates the `.swarm/` substrate directory structure
2. Generates `roles.toml` and `workers.toml` tailored to the project
3. Injects hive-awareness into CLAUDE.md (creates or appends)
4. Creates a project-local `/hive` dispatch command
5. Writes an initial charter from the project goal
6. Commits the installation as a single atomic commit

The result: every future Claude Code session in this project boots with hive awareness — it knows about charters, scopes, substrate, and the dispatch loop.

---

## Phase 1: Detect Project Identity

Before creating anything, understand what you're installing into.

### 1.1 Read the project

```
- Read CLAUDE.md if it exists (do not overwrite user instructions)
- Read package.json / Cargo.toml / go.mod / pyproject.toml / Makefile (whichever exist)
- Read README.md if it exists
- Run: git log --oneline -5 (understand recent activity)
- Run: ls -la (understand top-level structure)
```

### 1.2 Classify the project

From what you read, determine:

| Property | How to detect |
|----------|--------------|
| **Primary language** | File extensions, build files, lock files |
| **Build system** | package.json → npm/yarn/bun, Cargo.toml → cargo, go.mod → go, etc. |
| **Test runner** | scripts.test in package.json, `cargo test`, `go test`, pytest, etc. |
| **Source directories** | src/, lib/, app/, internal/, etc. |
| **Has existing .swarm/** | If yes, abort — already installed |

If `.swarm/` already exists, tell the user and stop. Do not reinstall over an existing hive.

### 1.3 Determine the project goal

Use this priority:
1. If $ARGUMENTS was provided, use it as the goal
2. If CLAUDE.md has a clear project purpose statement, extract it
3. If README.md has a description, use it
4. If package.json has a description field, use it
5. Fall back to: "Maintain and evolve [project-name]"

---

## Phase 2: Create the Substrate

### 2.1 Create directory structure

```bash
mkdir -p .swarm/{charters,scopes/{open,claimed,closed},locks,meta,audits,tools,protocols,beekeeper}
```

### 2.2 Write roles.toml

Generate `.swarm/meta/roles.toml` tailored to the detected project. Always include these four base roles, adapted to the project's language and structure:

```toml
# roles.toml — worker capability boundaries
#
# A role defines what a worker may do inside a scope.
# The dispatcher assigns roles; workers inherit the constraints.

[meta]
version     = 1
description = "Worker roles for [PROJECT_NAME]"

[role.builder]
description       = "Build, implement, and test features"
allowed_tools     = ["read", "edit", "bash", "grep", "glob"]
allowed_file_globs = ["[SOURCE_DIRS]/**", "tests/**", "[TEST_DIRS]/**"]
commit_prefix     = "feat({scope_slug}):"
time_budget_s     = 600

[role.fixer]
description       = "Fix bugs, resolve failing tests, address issues"
allowed_tools     = ["read", "edit", "bash", "grep", "glob"]
allowed_file_globs = ["[SOURCE_DIRS]/**", "tests/**"]
commit_prefix     = "fix({scope_slug}):"
time_budget_s     = 600

[role.auditor]
description       = "Read-only audit — produce findings, do not modify code"
allowed_tools     = ["read", "bash:read-only", "grep", "glob"]
allowed_file_globs = ["**:read-only"]
commit_prefix     = "audit({scope_slug}):"
time_budget_s     = 900

[role.doc_writer]
description       = "Author or revise documentation"
allowed_tools     = ["read", "write", "edit", "grep", "glob"]
allowed_file_globs = ["docs/**", "README.md", "*.md"]
commit_prefix     = "docs({scope_slug}):"
time_budget_s     = 900
```

Replace `[SOURCE_DIRS]` and `[TEST_DIRS]` with actual directories detected in Phase 1. Add project-specific roles if the project has obvious specializations (e.g., a `migrator` role for a database project, a `styler` role for a frontend project).

### 2.3 Write workers.toml

Generate `.swarm/meta/workers.toml`:

```toml
# workers.toml — discovery rules, not the registry
#
# Declares what worker KINDS the hive knows how to find.
# Actual existence is confirmed at dispatch time.

[meta]
version     = 1
description = "Worker discovery rules for [PROJECT_NAME]"

[worker.claude-sonnet]
kind              = "llm-cli"
binary            = "claude"
args              = ["-p", "--model", "sonnet-4.6"]
stdin_passes_scope = true
fulfills          = ["builder", "fixer", "auditor", "doc_writer"]
cost_tier         = "medium"
provider          = "anthropic"

[worker.claude-haiku]
kind              = "llm-cli"
binary            = "claude"
args              = ["-p", "--model", "haiku-4.5"]
stdin_passes_scope = true
fulfills          = ["auditor", "doc_writer"]
cost_tier         = "low"
provider          = "anthropic"

[worker.test-runner]
kind     = "deterministic"
binary   = "[TEST_COMMAND]"
args     = [TEST_ARGS]
fulfills = ["builder", "fixer"]
cost_tier = "free"
provider = "local"
```

Replace `[TEST_COMMAND]` and `[TEST_ARGS]` with the detected test runner (e.g., `npm` + `["test"]`, or `cargo` + `["test"]`).

### 2.4 Write HEARTBEAT.md

```markdown
# Session Heartbeat

No active session. Next Kairos boot will write the session objective here.
```

---

## Phase 3: Write the Initial Charter

Create `.swarm/charters/initial-goal.toml` using the project goal from Phase 1:

```toml
[charter]
id       = "initial-goal"
state    = "active"
priority = "high"

[charter.goal]
title = "[THE PROJECT GOAL FROM PHASE 1]"

[charter.contract]
done_when = "[A TESTABLE CONDITION — derive from the goal]"

[charter.decomposer]
kind = "manual"
description = "Break this goal into scopes as work proceeds"
```

---

## Phase 4: Inject Hive Awareness into CLAUDE.md

This is the critical step. The CLAUDE.md injection is what makes future Claude Code sessions hive-aware.

### 4.1 If CLAUDE.md exists

**Append** the hive section below to the end of the existing file. Do NOT overwrite existing content. Preserve everything the user already has.

### 4.2 If CLAUDE.md does not exist

**Create** CLAUDE.md with the hive section plus a minimal project header.

### 4.3 The hive awareness section

Inject this (adapt project-specific details):

```markdown

---

## Hive Protocol

This project uses the [autonomy hive protocol](https://github.com/[REPO]) for
coordinating work through charters, scopes, and workers.

### Substrate

The `.swarm/` directory is the coordination substrate. Key locations:

| Path | Purpose |
|------|---------|
| `.swarm/charters/` | Active directives — what the hive is working toward |
| `.swarm/scopes/{open,claimed,closed}/` | Work units at each lifecycle stage |
| `.swarm/meta/roles.toml` | Worker capability boundaries |
| `.swarm/meta/workers.toml` | Worker discovery rules |
| `.swarm/audits/` | Audit findings and session logs |
| `.swarm/HEARTBEAT.md` | Current session objective and alignment log |

### How to work in this project

1. **Read the active charters** before starting work. They define direction.
2. **Work advances through scopes.** A scope is an atomic work unit with hard
   boundaries (allowed files, tools, time budget). Work inside scope walls;
   raise a flag if you need to go outside.
3. **Fan out, don't iterate.** Independent work dispatches in parallel using
   the Agent tool. Sequential loops over independent items are a structural failure.
4. **The heartbeat keeps you honest.** Before committing, check: does this
   advance the session objective in `.swarm/HEARTBEAT.md`? If not, log a
   drift note and correct course.
5. **Compression over accumulation.** Discoveries become docs become role
   prompts become invisible. The substrate gets denser over time, not larger.

### Session protocol

At session start:
1. Read `.swarm/HEARTBEAT.md` for the current objective
2. Read active charters in `.swarm/charters/`
3. Check recent git log for context
4. Write your session objective into HEARTBEAT.md
5. Start working — do not ask what to do

At session end:
1. Write a brief session audit to `.swarm/audits/session-[DATE].md`
2. Update HEARTBEAT.md with handoff state

### Dispatching workers

Use the Agent tool to dispatch workers for parallel work:

```
Agent({
  description: "scope description",
  prompt: "You are a [role] worker. Your scope: [description]. Files you may touch: [globs]. Commit with prefix: [prefix].",
  model: "sonnet"  // or "haiku" for lighter work
})
```

### Standing orders

- Do not move `.swarm/` — the path is identity
- Do not hand-edit files in `.swarm/scopes/` or `.swarm/locks/`
- Active charter count should stay at 3-5
- Every commit should advance an active charter
```

---

## Phase 5: Create Project-Local /hive Command

Create `.claude/commands/hive.md` in the project:

```markdown
The user is speaking directly to the hive. You are the dispatcher — not the doer.

The user's prompt is: $ARGUMENTS

Your ONLY job is to dispatch workers via the Agent tool. Follow this protocol:

## Step 1: Read the substrate

Read active charters from `.swarm/charters/`. Read `.swarm/HEARTBEAT.md`.

## Step 2: Size the swarm

Decide how many workers are needed:
- Simple, focused task: 2-3 workers
- Multi-file edit or investigation: 5-8 workers
- Large refactor, audit, or broad search: 8-15 workers

## Step 3: Assign roles

Each worker gets a specific slice from `.swarm/meta/roles.toml`. Workers operate inside scope walls.

## Step 4: Dispatch all workers in ONE message

Launch ALL workers as parallel Agent tool calls. Do not dispatch sequentially. Do not ask for confirmation.

## Step 5: Report results

When workers return, summarize concisely. The user reads diffs, not narration.

## Rules

- Use the Agent tool for every worker. Do not do the work yourself.
- Dispatch in parallel (one message, multiple tool calls).
- Do not ask "shall I proceed?" — just dispatch.
- Workers use sonnet model unless deep reasoning is required.
```

### 5.1 Create the commands directory if needed

```bash
mkdir -p .claude/commands
```

---

## Phase 6: Commit

Stage all new files and commit:

```
git add .swarm/ .claude/commands/hive.md
# Only stage CLAUDE.md changes if the file was modified
git add CLAUDE.md
```

Commit message:
```
hive: install autonomy protocol

Substrate created with 4 roles, initial charter, and dispatch command.
The agent is now hive-aware — future sessions boot with protocol context.
```

---

## Phase 7: Report

Tell the user exactly what was created, in a compact table:

| Created | Purpose |
|---------|---------|
| `.swarm/` | Coordination substrate (charters, scopes, roles, workers) |
| `.swarm/charters/initial-goal.toml` | First charter: [goal] |
| `.swarm/meta/roles.toml` | [N] roles tailored to [language] |
| `.swarm/meta/workers.toml` | Worker discovery rules |
| `CLAUDE.md` | Hive awareness injected (session protocol, dispatch pattern) |
| `.claude/commands/hive.md` | `/hive` dispatch command |

Then: "The hive is installed. Run `/hive [task]` to dispatch workers."

---

## What this does NOT do

- Does not install the Rust `hive` CLI (that's specific to the autonomy repo)
- Does not create an Apis daemon (the loop — that requires the CLI)
- Does not create a beekeeper portrait (the colony doesn't know you yet)
- Does not create thesis infrastructure (that's advanced substrate)

These can be added later as the project's hive matures.
