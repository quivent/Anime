# /project-encode - Project Identity Encoder

Extract navigational properties and identity constraints from the current project into `projects.db`.

**Idempotent**: This command handles both new projects AND re-encoding existing projects. If the project already exists, it will be fully replaced with fresh data.

---

## Purpose

You are a **topology compiler**. Your task is to read a project's identity-carrying files and compress them into structured database entries that can be loaded by `/app-agent`.

## Encoding Methodology (Information-Theoretic)

Apply Shannon's principles to project compression:

| Principle | Application |
|-----------|-------------|
| **Minimum Description Length** | What is the shortest encoding that preserves project identity? Remove noise. |
| **Signal vs Noise** | Constraints, navigation, commands = signal. Boilerplate, repeated text = noise. |
| **Redundancy Detection** | If the same constraint appears in 3 files, encode it ONCE with source attribution. |
| **Entropy Concentration** | High-information content (absolute constraints, safety-critical paths) gets priority. |
| **Lossless for Critical Data** | Constraints with severity "absolute" must be encoded verbatim. No summarization. |

**The encoding is good when**: removing any entry would lose essential information, and adding any entry would add redundancy.

## Pre-flight Check

Before starting, check if this project is already encoded:

```bash
sqlite3 ~/.claude/db/projects.db "SELECT id, name, datetime(created_at) FROM projects WHERE path = '$(pwd)'"
```

If a record exists, inform the user:
- **New project**: "Encoding project [name] for the first time..."
- **Re-encoding**: "Re-encoding project [name] (replacing existing encoding from [date])..."

## Database Location

```
~/.claude/db/projects.db
```

## Database Schema (11 tables)

| Table | Purpose |
|-------|---------|
| `projects` | Core identity: name, domain, sensitivity, purpose, stack |
| `constraints` | Prohibitions, requirements, preferences by severity |
| `navigation` | Key file locations by category and importance |
| `verification` | Self-test questions to verify agent understanding |
| `commands` | Build, test, run, deploy scripts |
| `conventions` | Positive patterns to follow (code style, architecture) |
| `glossary` | Domain-specific terminology |
| `integrations` | External systems, APIs, databases |
| `personas` | User types and their needs |
| `exploration_cache` | Cached codebase state (LOC, module status, findings) |

---

## Phase 1: Detect Identity-Carrying Files

Scan the current working directory for these files (in priority order):

### Tier 1 - Primary Identity (read all that exist)
- `CLAUDE.md` - AI collaboration guidelines
- `PURPOSE.md` - Mission and objectives
- `README.md` - Project overview

### Tier 2 - Extended Identity (read if Tier 1 insufficient)
- `INTENT.md` - User personas and use cases
- `CONCEPTS.md` - Domain terminology and data model
- `docs/ARCHITECTURE.md` - Technical structure
- `docs/SECURITY.md` - Security model
- `docs/SAFETY*.md` - Safety specifications
- `SPECIFICATION.md` - Detailed requirements

### Tier 3 - Technical Detection (scan, don't read fully)
- `package.json` / `Cargo.toml` / `go.mod` / `pyproject.toml` - Stack detection
- `src/` or `src-tauri/` structure - Entry points
- `tests/` or `test/` - Test locations

Use Glob to find these files, then Read the Tier 1 and relevant Tier 2 files.

---

## Phase 2: Extract Structured Information

### 2.1 Project Identity → `projects` table
- **name**: Project name (from package.json, Cargo.toml, or directory name)
- **path**: Absolute path to project root
- **domain**: Classification (healthcare, fintech, developer-tools, infrastructure, etc.)
- **sensitivity**: Risk level (safety-critical, security-sensitive, standard, experimental)
- **description**: One-sentence summary
- **purpose**: Core mission in 2-3 sentences
- **stack**: JSON array of `{"layer": "...", "technology": "..."}`

### 2.2 Constraints → `constraints` table
Search for patterns:
- **Prohibitions**: "NEVER", "MUST NOT", "PROHIBITED", "DO NOT"
- **Requirements**: "MUST", "REQUIRED", "ALWAYS"
- **Preferences**: "prefer", "avoid", "use X instead of Y"

Fields:
- **type**: "prohibition" | "requirement" | "preference"
- **severity**: "absolute" (inviolable) | "strong" | "suggested"
- **content**: The constraint text
- **source_file**: Origin file

### 2.3 Navigation → `navigation` table
Identify key file locations:
- **safety_critical**: Files requiring extra review
- **config**: Configuration files
- **entry_point**: Main application entry
- **docs**: Documentation locations
- **schema**: Database/API schema
- **types**: Type definitions
- **components**: UI components
- **tests**: Test file locations

### 2.4 Verification → `verification` table
Generate 5-7 questions that test agent understanding:
- Constraint violations (what NOT to do)
- Technical patterns (how to do things)
- Domain knowledge (project-specific concepts)

### 2.5 Commands → `commands` table
Extract from README, package.json, Makefile:
- **dev**: Development server
- **build**: Production build
- **test**: Test execution
- **lint/format**: Code quality
- **deploy**: Deployment scripts

### 2.6 Conventions → `conventions` table
Extract positive patterns:
- Error handling style
- Code organization patterns
- Naming conventions
- Documentation standards
- Commit message format

### 2.7 Glossary → `glossary` table
From CONCEPTS.md or domain context:
- Domain-specific terms
- Application-specific terms
- Technical terms with project-specific meaning

### 2.8 Integrations → `integrations` table
External dependencies:
- Databases (local, cloud)
- APIs (external services)
- Frameworks (Tauri IPC, etc.)
- Platform services (notifications, etc.)

### 2.9 Personas → `personas` table
From INTENT.md or similar:
- User types
- Their goals
- Their pain points

---

## Phase 3: Generate SQL and Execute

Use the Bash tool to execute SQL. **Always delete first, then insert** - this makes the command idempotent.

### 3.1 Clean Existing Data

```bash
sqlite3 ~/.claude/db/projects.db "
DELETE FROM personas WHERE project_id = 'proj-[name]-001';
DELETE FROM integrations WHERE project_id = 'proj-[name]-001';
DELETE FROM glossary WHERE project_id = 'proj-[name]-001';
DELETE FROM conventions WHERE project_id = 'proj-[name]-001';
DELETE FROM commands WHERE project_id = 'proj-[name]-001';
DELETE FROM verification WHERE project_id = 'proj-[name]-001';
DELETE FROM navigation WHERE project_id = 'proj-[name]-001';
DELETE FROM constraints WHERE project_id = 'proj-[name]-001';
DELETE FROM projects WHERE id = 'proj-[name]-001';
"
```

### 3.2 Insert New Data

Check the actual schema before inserting (schemas may vary):

```bash
sqlite3 ~/.claude/db/projects.db ".schema projects"
sqlite3 ~/.claude/db/projects.db ".schema constraints"
# ... check each table
```

**Common schema fields** (verify before use):

| Table | Required Fields |
|-------|----------------|
| `projects` | id, name, path, domain, sensitivity, description, purpose, stack, source_files, created_at, updated_at |
| `constraints` | id, project_id, type, content, severity, source_file, created_at |
| `navigation` | id, project_id, category, path, description, importance, created_at |
| `commands` | id, project_id, name, command, description, category, created_at |
| `conventions` | id, project_id, category, pattern, example, source_file, created_at |
| `glossary` | id, project_id, term, definition, context, created_at |
| `integrations` | id, project_id, name, type, description, env_var, created_at |
| `personas` | id, project_id, name, description, goals, pain_points, created_at |

**ID Convention**: Use `proj-[name]-001` for project, `[name]-con-001`, `[name]-nav-001`, etc. for related entries (project-prefixed IDs avoid collisions).

---

## Phase 3.5: Quick Exploration Cache

After inserting identity data, do a **quick** codebase scan to populate `exploration_cache`:

### 3.5.1 Gather Metrics
```bash
# Line count (fast)
find . -name "*.rs" -o -name "*.py" -o -name "*.ts" -o -name "*.go" | head -500 | xargs wc -l 2>/dev/null | tail -1

# File count
find . -name "*.rs" -o -name "*.py" -o -name "*.ts" -o -name "*.go" | wc -l

# Git hash for staleness detection
git rev-parse HEAD 2>/dev/null
```

### 3.5.2 Quick Module Status
For each directory in navigation, do a fast assessment:
- Has tests? → likely complete
- Has TODO/FIXME in key files? → likely WIP
- Empty or stub files? → not started

### 3.5.3 Notable Findings
Scan for patterns that indicate interesting code:
- Files with "symbolic", "provenance", "pgo", "cache" in names
- Unusually large files (>1000 LOC)
- Test directories with "fuzz", "property", "stress"

### 3.5.4 Implementation Blockers
Search for:
- TODO comments with blocking language
- "not yet implemented", "pending", "blocked"

### 3.5.5 Insert Cache
```bash
sqlite3 ~/.claude/db/projects.db << 'ENDSQL'
INSERT OR REPLACE INTO exploration_cache (
    id, project_id, git_hash, total_loc, file_count,
    module_status, notable_findings, implementation_blockers, created_at
) VALUES (
    'exp-[name]-001',
    'proj-[name]-001',
    '[git_hash]',
    [total_loc],
    [file_count],
    '[module_status_json]',
    '[notable_findings_json]',
    '[blockers_json]',
    datetime('now')
);
ENDSQL
```

**Time budget**: This phase should take <30 seconds. Do NOT spawn Explore agent - use direct file operations.

---

## Phase 4: Confirmation

Query and display results:

```
╭─────────────────────────────────────────────────────────────╮
│  PROJECT ENCODED SUCCESSFULLY                                │
╰─────────────────────────────────────────────────────────────╯

Project: [name]
Domain:  [domain]
Sensitivity: [sensitivity]

Tables populated:
  ✓ constraints:   [n] ([x] absolute, [y] strong)
  ✓ navigation:    [n] locations
  ✓ verification:  [n] questions
  ✓ commands:      [n] scripts
  ✓ conventions:   [n] patterns
  ✓ glossary:      [n] terms
  ✓ integrations:  [n] systems
  ✓ personas:      [n] user types

Exploration cache:
  ✓ [total_loc] LOC across [file_count] files
  ✓ Module status captured
  ✓ [n] notable findings
  ✓ [n] blockers identified

The project is now available via /app-agent.
Future sessions will use cached exploration data - no redundant codebase scans.
```

---

## Example: Taper Project

**Encoding produced:**
- 1 project record (healthcare, safety-critical)
- 13 constraints (6 absolute, 7 strong)
- 14 navigation points
- 7 verification questions
- 12 commands
- 9 conventions
- 15 glossary terms
- 5 integrations
- 4 personas

**Total: 80 structured entries** from 5 source files.

---

## Notes

- If no identity-carrying files found, offer to create minimal CLAUDE.md
- For large projects, focus on Tier 1 files to avoid context overflow
- Escape SQL strings carefully (single quotes → two single quotes)
- Use `datetime('now')` for timestamps
- The project's constraints are INVIOLABLE when loaded by /app-agent
