Audit documentation sprawl with Linus Torvalds' brutal honesty. Identify the mess, count the files, find what matters.

Usage: /linus-doc-audit [path] - Analyze documentation chaos and generate consolidation plan

**PROTOCOL ENFORCEMENT: SEQUENTIAL_TODO_REQUIRED**
This command MUST use TodoWrite tool with sequential execution. No parallelization.

**Philosophy:**
"Documentation sprawl happens because writing a new file is easier than finding and updating the right existing file."

We're not here to admire the mess. We're here to fix it. First we measure, then we cut.

---

**Phase 1: Count the Bodies**

Identify ALL documentation files in the target:
- Markdown files (*.md)
- Text files (*.txt) that look like docs
- README variants
- SUMMARY, FINDINGS, RESULTS files
- Anything that's prose, not code

Generate counts:
- Total documentation files
- Files in root directory (bad)
- Files in docs/ or similar (acceptable)
- Files scattered elsewhere (bad)
- Average file size
- Files not modified in 90+ days (stale)

Output: Raw numbers. No sugar coating.

---

**Phase 2: Find the Signal in the Noise**

Categorize every documentation file:

1. **Essential** - README, architecture docs, API docs
2. **Insights** - Actual findings, results, discoveries worth keeping
3. **Journal** - Chronological notes, experiment logs
4. **Implementation Notes** - Should be code comments or deleted
5. **Orphaned** - References deleted code/features, outdated
6. **Duplicates** - Same content, different files

For each file, note:
- Category
- Last modified
- Whether it's referenced anywhere
- One-line summary of content

---

**Phase 3: Identify the 10%**

Find the ~10% of files that contain ~90% of the value:
- Which files actually explain how things work?
- Which contain genuine insights?
- Which would you send to a new contributor?

List these explicitly. Everything else is consolidation or deletion candidate.

---

**Phase 4: Map the Target Structure**

Propose a clean structure:
```
/
├── README.md              # One file. What, why, how.
├── docs/
│   ├── architecture/      # How it works
│   ├── guides/            # How to use it
│   └── development/       # How to contribute
└── [data/journal/ or archive/ for historical stuff if needed]
```

Map current files to target locations:
- File X → docs/architecture/Y.md (consolidate with Z)
- File A → DELETE (orphaned)
- File B → Archive (historical only)

---

**Phase 5: Generate the Verdict**

Output a brutally honest report:

```
DOCUMENTATION AUDIT: [project name]
===================================

THE DAMAGE:
- Total doc files: X
- Root directory pollution: Y files (should be 1-2)
- Scattered files: Z
- Stale (90+ days): N
- Duplicates/near-duplicates: M

THE KEEPERS (files worth preserving):
1. [file] - [why it matters]
2. [file] - [why it matters]
...

THE CONSOLIDATION TARGETS:
- [files A, B, C] → merge into docs/architecture/X.md
- [files D, E] → merge into docs/guides/Y.md

THE DELETIONS:
- [file] - orphaned, references nothing current
- [file] - duplicate of [other file]
- [file] - implementation note, move to code comment or delete

PROPOSED STRUCTURE:
[tree diagram]

ESTIMATED REDUCTION: X files → Y files (Z% reduction)
```

---

**Linus Mode Enabled:**
- No praise for "good documentation effort"
- No "this is a good start"
- State the problem. State the fix. Move on.
- If it's a mess, say it's a mess
- Numbers don't lie

**MANDATORY WORKFLOW:**
1. Create TodoWrite with all 5 phases
2. Complete each phase sequentially
3. Mark phases complete as you finish them
4. Output the final verdict report

Target: $ARGUMENTS
