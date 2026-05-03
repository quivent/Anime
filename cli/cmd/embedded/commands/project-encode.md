# /project-encode — Project Identity Encoder

Compress a project's identity into one JSON file that `/app-agent` loads at activation.

**Design principle**: minimum description length. Identity is ~2 KB of essential signal: who you are, what you must not break, how to build, what the stack is. Everything else is reconstructible on demand from the codebase.

**Target runtime**: under 60 seconds. One probe, one parallel read batch, one extraction pass, one write.

---

## Output

Single file: `~/.claude/encodings/[path-with-dashes]/core.json`

The path encoding replaces `/` with `-` and drops the leading `-`.
Example: `/Users/joshkornreich/DAW` → `~/.claude/encodings/-Users-joshkornreich-DAW/core.json`

Create the directory if it does not exist. Do not write `stamp.json`, `map.json`, `meta.json`, or per-section files — `/app-agent` falls back gracefully when those are absent and reads `git_hash` / `encoded_at` from `core.json` directly.

---

## Pipeline

### Step 1 — Probe (single parallel batch)

Issue these in one message:

```bash
# Existence check for all candidate identity files at once
ls CLAUDE.md PURPOSE.md README.md INTENT.md \
   pyproject.toml package.json Cargo.toml go.mod \
   docs/ARCHITECTURE.md docs/SAFETY.md docs/SECURITY.md 2>/dev/null

# Current commit
git rev-parse HEAD 2>/dev/null
```

### Step 2 — Read everything that exists, in parallel

One message containing parallel `Read` calls for every identity file the probe confirmed exists. Do NOT serialize these reads.

Identity files (priority order, read all that exist):
- **Tier 1**: `CLAUDE.md`, `PURPOSE.md`, `README.md`
- **Tier 2**: `INTENT.md`, `docs/ARCHITECTURE.md`, `docs/SAFETY.md`, `docs/SECURITY.md`
- **Manifest** (one of): `pyproject.toml`, `package.json`, `Cargo.toml`, `go.mod`

If none of Tier 1 exists, the project has no encoded identity. Offer to create a minimal `CLAUDE.md` and stop.

### Step 3 — Extract in one pass

Produce the entire JSON in a single reasoning step. Do not loop section-by-section. The schema is small enough to fill in one shot.

```json
{
  "identity": {
    "name": "...",
    "path": "/abs/path",
    "domain": "developer-tools | healthcare | fintech | infrastructure | research | ...",
    "sensitivity": "safety-critical | security-sensitive | standard | experimental",
    "purpose": "1-3 sentences"
  },
  "stack": ["language: rust", "framework: tauri", "..."],
  "constraints": [
    {
      "severity": "absolute | strong | preference",
      "type": "prohibition | requirement",
      "content": "verbatim from source for absolute; paraphrased for strong/preference",
      "source_file": "CLAUDE.md"
    }
  ],
  "commands": [
    {"name": "build", "command": "...", "category": "build | test | run | deploy"}
  ],
  "agent": {
    "name": "wavesmith",
    "activation_command": "/wavesmith",
    "profile": "~/.agents/wavesmith.md"
  },
  "git_hash": "...",
  "encoded_at": "2026-04-29T20:30:00-04:00"
}
```

**What goes in, what stays out:**

| Include | Skip |
|---|---|
| Constraints with NEVER / MUST / DO NOT — verbatim if `absolute` | Conventions, naming style, code patterns (derive on demand) |
| Build / test / run / deploy commands | Glossary terms (read CONCEPTS.md when asked) |
| Stack as flat string list | Personas, user types (read INTENT.md when asked) |
| `agent` block only if a `/persona` slash command is wired up | Verification questions, archetype facets, exploration metrics |

**Constraint extraction rules:**
- `absolute`: contains NEVER, MUST NOT, PROHIBITED, "inviolable" → preserve verbatim
- `strong`: contains MUST, REQUIRED, ALWAYS → may paraphrase, keep meaning intact
- `preference`: "prefer", "avoid", "use X instead of Y" → paraphrase freely
- Deduplicate: if the same rule appears in 3 files, encode once with the highest-severity source

**Stack:** one line per layer. `"language: python"`, `"framework: fastapi"`, `"db: postgres"`. Do not nest.

**Commands:** extract from `package.json` `scripts`, `pyproject.toml` `[project.scripts]`, `Makefile` targets, README usage blocks. Cap at ~10 entries — the most-used ones.

**Agent block:** include only if you find a clear `/persona-command` reference in CLAUDE.md or the project hosts a domain-expert persona file. Otherwise omit the key entirely.

### Step 4 — Write

One `Write` call to `core.json`. Pretty-print with sorted keys for diff stability.

If a prior `core.json` exists at the path, overwrite it. Do not byte-compare or skip — the extraction itself is cheap enough that the optimization is not worth the spec complexity.

If a legacy `stamp.json`, `map.json`, `meta.json`, `context.json`, or per-section file (`identity.json`, `commands.json`, etc.) sits in the encoding directory from a prior encoder version, leave them alone. `/app-agent` ignores them when `core.json` is present and authoritative. (A user can `rm` the directory before re-encoding for a clean slate.)

### Step 5 — Confirm

```
Encoded [name] → ~/.claude/encodings/[path-with-dashes]/core.json
  domain:      [domain]
  sensitivity: [sensitivity]
  constraints: [n] ([k] absolute)
  commands:    [n]
  stack:       [n] layers
  agent:       [name | none]
```

That's it. Five lines, no banner art.

---

## Notes

- Re-encoding is a full re-run. No diffing, no caching. If it gets slow enough to matter again, add caching then — not preemptively.
- This command does not touch `projects.db`. The single `core.json` is the sole encoding artifact.
- Tier 2 files (ARCHITECTURE, SAFETY, SECURITY, INTENT) inform extraction but do not get their own sections. Their constraints flow into `constraints[]`; their stack hints flow into `stack`.
- For very large CLAUDE.md / PURPOSE.md (>50 KB), Read with a `limit` of 800 lines — identity content is almost always in the first portion.
- The `/app-agent` command consumes `core.json` and works without `stamp.json` (it reads `git_hash` and `encoded_at` inline from core when stamp is absent). This is the intended steady state.
