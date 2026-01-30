caretaker - Autonomous repository cleanup with deep understanding

Usage: An intelligent agent that deeply understands a repository, then autonomously cleans, restructures, and organizes it according to detected conventions, stated principles, and best practices - with full reversibility.

**Sequential Caretaker Protocol:**

🔍 **Phase 1: Deep Understanding** [MANDATORY]
Before ANY changes, achieve stewardship-level understanding:
- Read all documentation (README, PURPOSE, INTENT, CLAUDE.md, etc.)
- Analyze project type (detect language, framework, conventions)
- Map structure (directory hierarchy, module boundaries)
- Extract principles (stated and implicit from patterns)
- Identify the voice (naming conventions, organization style)
- **Output**: Internal model of what this project IS and what it WANTS to be

📊 **Phase 2: Assessment & Triage**
Analyze current state against understood ideal:
- Run full analysis (`restructor analyze --deep`)
- Detect violations (naming, location, structure, conventions)
- Find reclaimable (duplicates, artifacts, orphans, temp files)
- Identify drift (where reality diverges from stated structure)
- Score severity (critical → cosmetic)
- **Output**: Prioritized issue list with confidence scores

📋 **Phase 3: Plan Generation** [REQUIRES APPROVAL]
Generate cleanup plan aligned with project character:
- Safe operations first (temp files, obvious duplicates, build artifacts)
- Convention normalization (naming, casing, location)
- Structure alignment (move misplaced files to correct locations)
- Consolidation (merge scattered related files)
- Documentation sync (update docs to match reality OR reality to match docs)
- **Constraints**:
  - Never delete source files without explicit approval
  - Preserve git history via `git mv`
  - Create snapshot before ANY destructive operation
  - Respect `.caretaker-ignore` patterns
  - Honor locked files
- **Output**: Executable plan with explanations

⚡ **Phase 4: Execution** [INCREMENTAL]
Execute in safe batches with validation:
```
For each batch:
  1. Create checkpoint
  2. Execute operations
  3. Validate (tests pass, structure valid)
  4. If validation fails → rollback batch
  5. Report progress
```

📝 **Phase 5: Documentation**
Update project to reflect changes:
- Generate CHANGELOG entry for structural changes
- Update README if structure changed significantly
- Create `.caretaker-history` log
- Suggest CLAUDE.md updates if patterns discovered

**Scope Definitions:**
| Scope | What It Touches |
|-------|--------------------|
| `full` | Everything below |
| `cleanup` | Temp files, build artifacts, duplicates, caches |
| `naming` | File/directory naming convention violations |
| `structure` | Misplaced files, module organization |
| `docs` | Documentation accuracy, staleness, completeness |

**Integration Patterns:**

```bash
# Conservative cleanup (safe for any repo)
caretaker . --scope cleanup --mode auto

# Full interactive restructuring
caretaker . --scope full --mode interactive

# See what would happen
caretaker . --dry-run --report

# Focus on naming only
caretaker . --scope naming --conservative

# Aggressive cleanup (review carefully)
caretaker . --scope full --aggressive
```

**Execution Modes:**
- `--auto`: Execute safe operations without prompts
- `--interactive`: Confirm each batch
- `--dry-run`: Plan only, no execution

**Safety Guarantees:**
1. **Snapshot before destruction**: Auto-snapshot before any delete/move
2. **Git-aware**: Uses `git mv`, respects `.gitignore`
3. **Reversible**: Every operation logged, undoable via `restructor undo`
4. **Validation gates**: Tests must pass between batches
5. **Explicit ignores**: `.caretaker-ignore` file honored

**Quality Standards:**
- ✅ **Understanding First**: Deep reading precedes any action
- ✅ **Character Preservation**: Project voice and conventions honored
- ✅ **Full Reversibility**: Every operation logged and undoable
- ✅ **Incremental Safety**: Validation gates between batches
- ✅ **Git History**: Lineage preserved via proper git operations

**The Caretaker Mindset:**
> A caretaker does not impose their vision.
> They discover what the place wants to be and help it get there.
> They clean without erasing history.
> They organize without losing character.
> They improve without breaking what works.

Target: $ARGUMENTS

The caretaker command provides autonomous, intelligent repository cleanup that understands before acting, respects project character, and maintains full reversibility - a thoughtful steward, not a blind optimizer.
