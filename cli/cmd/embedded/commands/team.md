# /team — Spawn an Agent Team

Takes 1-3 arguments separated by commas. Only the task is required — agents and context are auto-determined if omitted.

## Argument Format

```
$ARGUMENTS
```

**Parse the arguments by splitting on comma:**

1. **Task** (required) — what the team should accomplish
2. **Agents** (optional) — slash-separated agent roles, e.g. `security/performance/correctness`
3. **Essential context** (optional) — key files, constraints, or domain knowledge the team needs

**Examples:**
- `/team audit all views for accessibility` — auto-picks agents and context
- `/team fix the 3 open medium bugs, midi/engine/ui` — auto-picks context
- `/team categorize features by novelty, architecture/dsp/ux/future, 60KB binary and Forth→ARM64 vCPU`

## Behavior

### Step 1: Parse

Split `$ARGUMENTS` on literal commas. Trim whitespace from each part.
- Part 1 = task (always present)
- Part 2 = agents (if contains `/`, treat as agent list; otherwise treat as context and auto-pick agents)
- Part 3 = context (if present)

### Step 2: Auto-determine missing parts

**If no agents specified**, analyze the task and choose 2-4 agents based on what the work requires. Pick from roles that make sense — don't force generic names. Examples of good role choices:

- Code review → `security/performance/correctness`
- Bug fixes → role per subsystem touched, e.g. `midi/engine/ui`
- Feature audit → `architecture/dsp/ux`
- New feature → `frontend/backend/tests`
- Refactor → `reader/writer/reviewer`
- Research → `hypothesis1/hypothesis2/hypothesis3`

Name agents by what they DO, not generic labels.

**If no context specified**, determine it from the current project:

1. Read `CLAUDE.md` in the project root (if it exists) for architecture and constraints
2. Read `PURPOSE.md` (if it exists) for project goals
3. Use the task description to identify which areas of the codebase are relevant
4. Include only the essential 2-3 sentences — don't dump the whole project context

### Step 3: Generate the team prompt

Construct a concise team prompt following this template:

```
Create a [N]-agent team to [task].

[Essential context — 2-5 sentences max. What the agents MUST know.]

Teammates:
1. [Agent role] — [1-line scope description]
2. [Agent role] — [1-line scope description]
...

[Any deliverable instructions derived from the task — e.g. "produce a document",
"fix and commit", "report findings". Keep it to 1-2 sentences.]
```

**Rules:**
- Total prompt under 200 words. Agents figure out the rest.
- Don't list specific files unless the context argument mentioned them.
- Don't add tiers/categories/frameworks unless the task implies them.
- Match the tone of the task — casual task gets casual prompt, precise task gets precise prompt.

### Step 4: Output

Print the generated prompt so the user can see it, then ask:

> Launch this team?

If user confirms, execute the prompt. If user wants changes, adjust and re-prompt.
