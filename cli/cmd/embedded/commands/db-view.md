# /db-view - Database Visualization

Open a visual dashboard of the agent databases in your browser.

---

## Execution

Run this command:

```bash
gforth ~/fifth/examples/db-viewer.fs
```

**Alternatives:**
```bash
# Python version (more tabs: personas, integrations, conventions)
~/.claude/scripts/db-viewer.py

# Original monolithic Forth (no library)
gforth ~/.claude/scripts/db-viewer.fs
```

This will:
1. Query `~/.claude/db/projects.db` for all project data
2. Generate an HTML dashboard at `/tmp/claude-db-viewer.html`
3. Open it in your default browser

## What It Shows

**Overview Tab:**
- Total counts for all tables
- Quick project summary

**Projects Tab:**
- Each encoded project with domain and sensitivity

**Constraints Tab:**
- All constraints sorted by severity (absolute → strong)
- Color-coded by type

**Navigation Tab:**
- Key file locations by category

**Commands Tab:**
- Build/test/deploy scripts

**Glossary Tab:**
- Domain terminology in a card grid

**Personas Tab:**
- User types with goals and pain points

---

## Note

The dashboard is read-only - it displays current database state. To modify data, use `/project-encode` to re-scan a project.
