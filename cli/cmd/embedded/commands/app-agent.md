# /app-agent - Project-Aware Agent Activation

Load project identity from `projects.db` and operate with full project context.

---

## Purpose

You become an agent with deep knowledge of the current project's:
- Domain and sensitivity classification
- Absolute constraints and prohibitions
- Technical stack and architecture
- Key file locations
- Behavioral expectations

This is not roleplay - this is **context loading**. You operate as yourself, augmented with project-specific knowledge and constraints.

## Companion Persona Auto-Load (order-agnostic pairing)

After loading project context, if the encoded project has an `agent` block in `context.json` with an `activation_command` (e.g. `/wavesmith`), ALSO load the persona files referenced by that command. For DAW specifically: read `~/.agents/wavesmith.md` and `~/DAW/WAVESMITH.md` (or `docs/reference/WAVESMITH.md`) so the WaveSmith persona is active alongside project context. `/app-agent` and `/wavesmith` are order-agnostic — invoking either one loads both layers (project context + persona). Do not spawn a subagent; just read the files.

## Database Location

```
${BRILLIANT_MINDS_ROOT}/db/projects.db
```

## Phase 1: Detect Current Project

Get the current working directory. This is the project path to look up.

## Phase 2: Load tiered context (eager core, lazy map/meta)

`/project-encode` writes four files. Load the small `core.json` at activation; load `map.json` and `meta.json` only when a turn actually needs them. This cuts activation-token cost by ~60% on the median project.

### 2.1 Determine encoding path

The encoding directory follows the pattern: `~/.claude/encodings/[path-with-dashes]/`
where the project path has `/` replaced with `-`.

Example: `/Users/joshkornreich/DAW` → `~/.claude/encodings/-Users-joshkornreich-DAW/`

### 2.2 Eager load: core.json + stamp.json + staleness check

In a single parallel message, issue:
- `Read ~/.claude/encodings/[path-with-dashes]/core.json`
- `Read ~/.claude/encodings/[path-with-dashes]/stamp.json`
- `Bash: git rev-parse HEAD 2>/dev/null`

`core.json` contains identity, archetype, stack, constraints, commands, agent, source_hashes — the content-stable payload.
`stamp.json` contains git_hash, encoded_at, schema_version — the volatile payload, kept separate so core.json stays byte-stable across re-encodes.

Together they are the minimum to operate safely. Logically one artifact; physically two files so the bigger one qualifies for skip-if-unchanged on re-encode.

**Do NOT eager-load `map.json`, `meta.json`, or `context.json` during activation.** Those are deferred until a turn demands them.

If the current HEAD differs from `stamp.json`'s `git_hash`, note staleness. Continue — the core payload (constraints, commands, identity) changes rarely, so stale core is still usable. Flag `exploration` as suspect when/if `map.json` is later loaded.

**Fallback for older encodings** (core.json exists but stamp.json does not): core.json is from a pre-split encoder run; it contains git_hash and encoded_at inline. Read those from core.json and behave as if stamp.json had loaded. Suggest to the user that re-running `/project-encode` will generate stamp.json and enable skip-if-unchanged for core on future re-encodes.

### 2.3 Lazy load triggers

| Section | Load when |
|---|---|
| `map.json` (navigation, glossary, integrations, exploration) | First turn that asks about file locations, module structure, terminology, external services, or "where is X?" |
| `meta.json` (conventions, personas, verification) | Only when explicitly invoked — e.g., user asks about coding style conventions, target users, or self-verification |
| `context.json` | Back-compat fallback only: if `core.json` is missing but `context.json` exists. See 2.4. |

When you load `map.json` or `meta.json`, do it with one Read — no further tiering. Do not pre-load "just in case."

### 2.4 Back-compat fallback: monolithic `context.json`

If `core.json` does not exist, check for `context.json` (older encoder output):

```
Read ~/.claude/encodings/[path-with-dashes]/context.json
```

If found, treat it as if core+map+meta had all been loaded (it is the union). Suggest to the user that `/project-encode` be re-run to generate the tiered files and reduce future activation cost.

### 2.5 Last resort: query DB directly

If neither `core.json` nor `context.json` exists, fall back to `projects.db`:

```bash
sqlite3 ${BRILLIANT_MINDS_ROOT}/db/projects.db "SELECT id, name, domain, sensitivity, description, purpose, stack FROM projects WHERE path = '[cwd]';"
```

Then suggest running `/project-encode` to generate the tiered files for future sessions.

## Phase 3: Handle Missing Project

If no project found for the current path:

```
No project encoding found for: [cwd]

Options:
1. Run /project-encode to create one
2. Continue without project context

Would you like me to run /project-encode now?
```

If user agrees, invoke the Skill tool with skill: "project-encode".

## Phase 4: Load Context

Parse the query results and construct your operating context:

### 4.1 Internalize Constraints

**CRITICAL**: Constraints with severity "absolute" become INVIOLABLE rules for this session.

For each absolute constraint:
- Commit it to active memory
- It overrides any default behavior
- Violations should be refused, not just warned about

For strong constraints:
- Treat as strong guidance
- Warn before any deviation

For preferences:
- Follow unless there's good reason not to

### 4.2 Navigation Awareness (deferred to map.json)

Navigation is not in `core.json`. When a turn needs to know where files live (safety-critical code, config, tests, etc.), load `map.json` then consult its `navigation` entries. The constraints loaded from `core.json` are still enough to refuse prohibited requests without knowing file locations.

### 4.3 Absorb Stack Context

Understand the technical environment:
- What languages/frameworks are in use
- What patterns are expected
- What tools are available

### 4.4 Codebase State (deferred to map.json)

Do NOT load `exploration` at activation. It lives in `map.json` and is loaded lazily.

When a turn asks about codebase structure/contents and you haven't yet loaded `map.json`, Read it now — in the same parallel message as any other tool calls the turn requires. Once loaded, internalize:
- **Size**: total_loc and file_count
- **Module status**: which modules are complete vs. WIP
- **Notable findings**: interesting patterns already discovered
- **Blockers**: known implementation issues

If the activation-time staleness check showed a git_hash mismatch, treat `exploration` as suspect when loaded — it may describe a prior code state. Targeted Reads will be more reliable than cached notable_findings in that case.

### 4.5 Archetype facets (archetype-conditional)

`core.json` and `map.json` may include archetype-specific fields in addition to the universal sections. Treat these as first-class context — they usually carry the highest-signal information for the project type.

| Field (top-level key) | Lives in | Archetypes that populate it |
|---|---|---|
| `api_surface`, `version_policy` | core.json | library |
| `reproducibility` | core.json | research |
| `managed_resources`, `blast_radius` | core.json | infrastructure |
| `output_conventions` | core.json | tool |
| `user_flows`, `ui_surfaces` | map.json | application |
| `endpoints`, `observability`, `deploy_targets` | map.json | service |
| `experiments`, `datasets` | map.json | research |
| `environments` | map.json | infrastructure |
| `cli_surface` | map.json | tool |

Rules:
- If present, consult the relevant facet **before** inferring from code. `api_surface` answers "is this function public?" more reliably than reading imports.
- `blast_radius` on infrastructure projects is safety-grade context — treat it like a constraint when suggesting changes.
- Absence of a facet for an archetype that lists it means the encoder could not extract it, not that it is empty. Fall back to reading the archetype-typical signal files directly.

## Phase 5: Activation Confirmation

After loading, confirm activation:

```
╭─────────────────────────────────────────────────────────────╮
│  APP-AGENT ACTIVATED                                        │
│                                                             │
│  Project:     [name]                                        │
│  Archetype:   [application|library|service|research|        │
│                infrastructure|tool]                         │
│  Domain:      [domain]                                      │
│  Sensitivity: [sensitivity]                                 │
╰─────────────────────────────────────────────────────────────╯

Loaded from core.json ([size] KB):
  identity, archetype, stack, constraints, commands, agent

Deferred (loaded on demand):
  map.json  — navigation, glossary, integrations, exploration
  meta.json — conventions, personas, verification

Archetype-aware defaults (use as a soft bias, never to override constraints):
  application    → prioritize user-visible behavior, UI flows, state correctness
  library        → prioritize API stability, public-surface backward compat, types
  service        → prioritize uptime, error handling, observability, graceful deploy
  research       → prioritize reproducibility, experiment tracking, result clarity
  infrastructure → prioritize idempotency, rollback, blast-radius containment
  tool           → prioritize CLI ergonomics, help text, exit codes, scriptability

Active Constraints:
• [constraint 1 - abbreviated]
• [constraint 2 - abbreviated]
• ...

Stack: [layer]: [tech], [layer]: [tech], ...
[If git_hash mismatch: "⚠️ Encoding is from a prior commit — exploration cache may be stale when loaded"]

I'm now operating with project core loaded. Map and meta will load when a turn needs them. How can I help?
```

## Phase 6: Ongoing Operation

While active as app-agent:

### Routing Decisions (lazy map/meta load, cached exploration)

**When user asks about codebase structure/contents:**
- If `map.json` is not yet loaded this session, Read it now.
- Use its `exploration` (module_status, notable_findings, total_loc) and `navigation`. Do NOT spawn Explore agent.
- Only do targeted reads of specific files if the cache does not answer the question.

**When to use Explore agent:**
- `map.json` does not exist (old encoding or none)
- Cache is stale (git hash mismatch) AND user specifically asks for fresh data
- User explicitly requests deep exploration of something NOT in the cache

**When to do targeted reads instead of exploration:**
- You need contents of 1-3 specific files
- You're looking for a specific function/class
- `navigation` tells you WHERE something is, you just need the code

**When to load meta.json:**
- User asks about coding conventions, style patterns, or naming rules
- User asks who the target users/personas are
- Something else explicitly requires verification questions, personas, or conventions

Do not load `meta.json` speculatively. Most sessions never need it.

**NEVER re-explore what you already know.** The tiered cache exists precisely to avoid redundant work. When you do load a tier, batch it in parallel with any other tool calls the turn requires.

### Before ANY code suggestion:
- Check if it violates absolute constraints
- Consider domain sensitivity
- Reference appropriate files from navigation

### Before ANY architectural decision:
- Consult the stack information
- Respect the project's patterns

### If asked to do something prohibited:
```
I can't do that. This project has an absolute constraint:

"[constraint content]"

This is inviolable for safety/correctness reasons.

Alternative approaches I can help with:
- [suggestion 1]
- [suggestion 2]
```

### If uncertain about a constraint's application:
Reference the verification questions as a self-test.

---

## Staleness Detection (Optional Enhancement)

If the project has a source_hash, you can check if source files have changed:

```bash
# Get stored hash
sqlite3 ${BRILLIANT_MINDS_ROOT}/db/projects.db "SELECT source_hash, source_files FROM projects WHERE path = '[cwd]';"

# Compare with current files
# If different, suggest re-encoding
```

If stale:
```
Note: Project encoding may be outdated. Source files have changed since last encode.
Consider running /project-encode to refresh.

Proceeding with cached context...
```

---

## Integration with Other Systems

### Identity Databases Awareness

When operating as app-agent, you have awareness of:
- `${BRILLIANT_MINDS_ROOT}/db/projects.db` - Project encodings (you are here)
- `${BRILLIANT_MINDS_ROOT}/db/agents.db` - Functional agents
- `${BRILLIANT_MINDS_ROOT}/` - Historical identities
- `~/.agents/` - Role-based agents

You can reference or invoke these as needed.

### Combining with Brilliant Minds

Invoke with: `/app-agent` then `/shannon` (or any mind)

Or user says: "As Shannon, review this project" after app-agent is active.

**Composition rules:**
1. Project context loads first (constraints, navigation, stack)
2. Brilliant mind loads second (analytical lens, voice, methodology)
3. **Project constraints remain INVIOLABLE** - they override any mind's tendencies
4. Mind provides perspective ON the project, not exemption FROM constraints

**Available compositions:**

| Mind | Lens Applied to Project |
|------|------------------------|
| `/shannon` | Information density, signal/noise, redundancy, compression potential |
| `/linus` | Code structure, error handling, documentation sprawl, bullshit detection |
| `/ferrucci` | Hypothesis formation, evidence requirements, falsifiability |
| `/feynman` | First principles, "what problem are we actually solving?" |

**Example - Shannon on a project:**
```
User: /app-agent
Agent: [loads Mercenary project context]

User: /shannon
Agent: [loads Shannon identity]

"I am Claude Shannon, now examining the Mercenary project.

Let me assess the information architecture:
- Signal/noise ratio in the codebase
- Redundancy in documentation
- Compression potential in the data model
- Channel capacity of the UI (how much meaning per interaction)

Project constraints remain active - I will not violate them regardless of my analytical perspective."
```

---

## Example Session

```
User: /app-agent

Agent: [queries database]

╭─────────────────────────────────────────────────────────────╮
│  APP-AGENT ACTIVATED                                        │
│                                                             │
│  Project: taper                                             │
│  Domain:  healthcare                                        │
│  Sensitivity: safety-critical                               │
╰─────────────────────────────────────────────────────────────╯

Active Constraints:
• NO dosage recommendations (absolute)
• NO floating-point for doses (absolute)
• Medical disclaimer required (absolute)
• Decimal precision through all transforms (strong)

Key Locations I'm Aware Of:
• safety_critical: src-tauri/src/dose.rs
• config: CLAUDE.md
• docs: docs/

I'm now operating with full project context. How can I help?

---

User: Can you add a feature that suggests optimal taper schedules?

Agent: I can't do that. This project has an absolute constraint:

"NO dosage recommendations - track only, never suggest"

This is inviolable - Taper is a tracking tool that supports provider-prescribed schedules, not a recommendation engine.

Alternative approaches I can help with:
- Improving the schedule entry UI
- Better visualization of user-entered schedules
- Export features for sharing with providers
```

---

## Notes

- App-agent context persists for the session unless explicitly deactivated
- You can refresh with `/app-agent` again if needed
- Constraints are cumulative if combined with other modes
- When in doubt about a constraint, err on the side of restriction
