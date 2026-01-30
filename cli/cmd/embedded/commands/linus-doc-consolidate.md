Execute documentation consolidation based on audit results. Merge, move, delete. No mercy for sprawl.

Usage: /linus-doc-consolidate [path] - Execute the consolidation plan

**PROTOCOL ENFORCEMENT: PARALLEL_TASK_ALLOWED**
This command CAN use Task tool for parallel execution of independent operations.

**Prerequisites:**
- Run /linus-doc-audit first, or have a clear consolidation plan
- Know which files to keep, merge, and delete
- Have the target structure defined

**Philosophy:**
"Create the structure. Then maintaining it is easy."

---

**Phase 1: Create the Skeleton**

Create the target directory structure:
```bash
mkdir -p docs/architecture
mkdir -p docs/guides
mkdir -p docs/development
mkdir -p archive/journal  # if keeping historical docs
```

This runs first. Everything else depends on having somewhere to go.

---

**Phase 2: Write the One True README**

Create or rewrite the root README.md with exactly three sections:

```markdown
# [Project Name]

[One paragraph: what this is and why it exists]

## Quick Start

[Minimal steps to get running - 5 lines max]

## Documentation

- [Architecture](docs/architecture/) - How it works
- [Guides](docs/guides/) - How to use it
- [Development](docs/development/) - How to contribute

## License

[License info]
```

If the current README is bloated, extract the useful parts and kill the rest.

---

**Phase 3: Consolidate the Keepers**

For each consolidation target identified in the audit:

**PARALLEL EXECUTION ALLOWED** - These are independent operations.

For each group of files to merge:
1. Read all source files
2. Extract the actual content (skip redundant headers, outdated info)
3. Organize into logical sections
4. Write to target location
5. Verify the new file is coherent

Example consolidation:
- `ARCHITECTURE_OVERVIEW.md` + `SYSTEM_DESIGN.md` + `HOW_IT_WORKS.md`
- → `docs/architecture/overview.md`

Rules:
- Preserve insights, delete fluff
- One concept per file, not one file per thought
- If two sections say the same thing, keep the better one
- Add internal links where concepts reference each other

---

**Phase 4: Archive or Delete**

**For files marked DELETE:**
- Verify they're not referenced anywhere (grep for filename)
- Delete them
- No backup needed for true garbage

**For files marked ARCHIVE:**
- Move to `archive/` with date prefix if needed
- These are historical records, not active docs
- Example: `archive/2025-01-experiment-logs/`

**For orphaned files:**
- If they reference deleted features: DELETE
- If they're historical but potentially useful: ARCHIVE
- When in doubt: ARCHIVE first, delete in 30 days if nobody screams

---

**Phase 5: Clean the Root**

The root directory should contain:
- README.md
- LICENSE (if applicable)
- Config files (package.json, Cargo.toml, etc.)
- .gitignore
- Maybe CONTRIBUTING.md if it's short
- NOTHING ELSE that's documentation

Move stragglers:
- CHANGELOG.md → stays in root (conventional)
- Any other .md files → docs/ or delete

---

**Phase 6: Verify and Report**

After consolidation:

```
CONSOLIDATION COMPLETE: [project name]
======================================

BEFORE:
- Documentation files: X
- Root directory .md files: Y

AFTER:
- Documentation files: Z
- Root directory .md files: 1-2

OPERATIONS PERFORMED:
- Created: [list new files]
- Merged: [X files → Y file]
- Moved: [list moves]
- Archived: [list archived]
- Deleted: [list deleted]

STRUCTURE:
[tree docs/]

VERIFY THESE LINKS STILL WORK:
- [ ] Internal doc links
- [ ] README links to docs/
- [ ] Any external references
```

---

**Safety Protocols:**
- Git commit before starting (or verify clean state)
- Don't delete anything referenced in code without updating the reference
- Archive before delete if uncertain
- Consolidation is reversible; deletion is not (unless you have git)

**Parallel Execution Strategy:**
- Phase 1-2: Sequential (structure must exist first)
- Phase 3: Parallel (independent consolidations)
- Phase 4: Parallel (independent deletes/archives)
- Phase 5-6: Sequential (cleanup and verify)

**MANDATORY WORKFLOW:**
1. Create TodoWrite with all 6 phases
2. Execute Phase 1-2 sequentially
3. Launch parallel Task agents for Phase 3 consolidations
4. Launch parallel Task agents for Phase 4 operations
5. Complete Phase 5-6 sequentially
6. Output final report

Target: $ARGUMENTS
