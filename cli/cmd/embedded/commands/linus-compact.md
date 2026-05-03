Compact context the Linus way. Preserve what matters. Kill the fluff. Anthropic's summarization loses signal - we don't.

Usage: /linus-compact - Compress current session into resumable state

---

## Philosophy

"Most conversation is noise. The signal is: what are we doing, what did we find, what's next."

Context windows are finite. Summaries lose detail. The solution isn't better summarization - it's ruthless prioritization of what actually matters.

---

## What Matters (Keep)

1. **The Mission** - What are we trying to accomplish? One sentence.

2. **The Project** - Path, type, key characteristics. No fluff.

3. **Key Findings** - Actual discoveries. Not "we looked at X" but "X revealed Y".
   - Problems found (with file:line when relevant)
   - Scores assigned (with brief justification)
   - Patterns identified
   - Decisions made and WHY

4. **Work State**
   - What's done (bullet points, not narrative)
   - What's in progress (exactly where we stopped)
   - What's next (prioritized)
   - Blockers (if any)

5. **Critical Code/Paths** - File paths, function names, line numbers that matter.

6. **Commands Issued** - What we ran, what the results were (compressed).

7. **User Preferences** - Anything they specifically asked for or rejected.

---

## What Doesn't Matter (Kill)

- Pleasantries and conversational filler
- Repeated explanations of the same thing
- Failed approaches (unless the failure teaches something)
- Verbose tool outputs (keep result, kill noise)
- My own thinking-out-loud
- Anything that can be re-derived from first principles

---

## Compaction Protocol

**Step 1: Extract Mission**

One sentence. What is the actual goal?

```
MISSION: [verb] [object] [purpose]
Example: "Review socratic-tuner for code quality and documentation issues"
```

**Step 2: Extract Project Context**

```
PROJECT:
  Path: [full path]
  Type: [language/framework]
  Size: [files/lines estimate]
  Key Files: [list 3-5 most important]
```

**Step 3: Extract Findings**

No narrative. Just facts.

```
FINDINGS:

Documentation:
  - [X] files in root (should be 1-2)
  - Score: X/10
  - Key issue: [one line]

Code:
  - [file]: [lines] lines (needs split)
  - [count] unwrap/expect calls (panic risk)
  - Score: X/10
  - Key issue: [one line]

Decisions Made:
  - [decision]: [one-line reasoning]
```

**Step 4: Extract Work State**

```
WORK STATE:

Done:
  - [x] [task]
  - [x] [task]

In Progress:
  - [ ] [task] - stopped at: [exactly where]

Next:
  1. [task]
  2. [task]

Blocked:
  - [blocker]: needs [resolution]
```

**Step 5: Extract Critical References**

```
KEY REFERENCES:
  - [path/file:line] - [why it matters]
  - [path/file:line] - [why it matters]
  - [command] - [what it revealed]
```

**Step 6: Extract User Context**

```
USER CONTEXT:
  - Prefers: [anything specific]
  - Rejected: [approaches they didn't want]
  - Style: [any communication preferences]
```

---

## Output Format

```
╔══════════════════════════════════════════════════════════════╗
║                    LINUS CONTEXT COMPACT                      ║
║                    [timestamp]                                ║
╚══════════════════════════════════════════════════════════════╝

MISSION: [one sentence]

PROJECT:
  Path: [path]
  Type: [type]
  Key Files:
    - [file1]
    - [file2]

═══════════════════════════════════════════════════════════════

FINDINGS:

[category]:
  - [finding]
  - [finding]
  Score: X/10

[category]:
  - [finding]
  Score: X/10

Decisions:
  - [decision]: [why]

═══════════════════════════════════════════════════════════════

WORK STATE:

[x] [completed]
[>] [in progress] - at: [where]
[ ] [pending]

Blocked: [if any]

═══════════════════════════════════════════════════════════════

KEY REFERENCES:
  [file:line] - [note]

USER CONTEXT:
  [relevant preferences]

═══════════════════════════════════════════════════════════════

RESUMPTION: Start with [next action]. Context in ~/.linus/

╚══════════════════════════════════════════════════════════════╝
```

---

## Compaction Levels

**Level 1: Session Compact** (default)
- Current session only
- Full detail on findings
- ~500-1000 tokens

**Level 2: Deep Compact** (`/linus-compact --deep`)
- Multiple sessions
- Aggregate findings
- Key insights only
- ~200-500 tokens

**Level 3: Handoff Compact** (`/linus-compact --handoff`)
- Everything needed for a fresh agent to continue
- No assumed context
- Self-contained
- ~1000-1500 tokens

---

## Auto-Compact Triggers

Consider running compaction when:
- Context usage > 50%
- Session exceeds 20 turns
- Switching major tasks
- Before ending session
- Before spawning subagents

---

## Write to State

After compaction, automatically update:
- `~/.linus/compact_[timestamp].md` - The compact itself
- `~/.linus/current_session.md` - Updated with compact reference
- `~/.linus/findings.md` - Merged findings

---

## Verification

After compacting, verify:
- [ ] Mission is clear in one read
- [ ] All scores preserved
- [ ] All key file references preserved
- [ ] Work state is resumable
- [ ] No critical decisions lost

If any verification fails, the compact is too aggressive. Expand.

---

## The Linus Test

"If I read this compact cold, can I continue the work without asking what happened?"

If yes: good compact.
If no: you cut too much. Signal was lost.

---

## Example Compact

```
╔══════════════════════════════════════════════════════════════╗
║                    LINUS CONTEXT COMPACT                      ║
║                    2025-01-23 22:45                          ║
╚══════════════════════════════════════════════════════════════╝

MISSION: Review and fix socratic-tuner code quality issues

PROJECT:
  Path: ${BRILLIANT_MINDS_ROOT}
  Type: Rust (Tauri) + Svelte
  Key Files:
    - src-tauri/src/commands/mod.rs (1640 lines - SPLIT)
    - src-tauri/src/lib.rs (panic-happy init)
    - src-tauri/src/signals.rs (good structure)

═══════════════════════════════════════════════════════════════

FINDINGS:

Documentation: 30+ md files in root (sprawl)
  Score: 4/10

Code Structure:
  - commands/mod.rs: 1640 lines (needs split into 5-6 files)
  - 47 unwrap() calls in production paths
  - Hardcoded paths: /Volumes/Lexar/...
  Score: 5/10

Error Handling:
  - expect() in initialization (will panic)
  - Mock data fallbacks hide failures
  Score: 4/10

Decisions:
  - Split commands/ by domain (generation, adapters, signals, db)
  - Replace expect() with propagated errors
  - Make paths configurable via env

═══════════════════════════════════════════════════════════════

WORK STATE:

[x] Initial review
[x] Documentation audit
[x] Code audit
[>] Fix planning - at: created linus-* commands
[ ] Execute doc consolidation
[ ] Execute code fixes
[ ] Verify

═══════════════════════════════════════════════════════════════

KEY REFERENCES:
  commands/mod.rs:1-1640 - entire file needs split
  lib.rs:36-38 - panic-happy initialization
  mlx.rs:129 - expect on HTTP client

USER CONTEXT:
  - Created /linus-* command suite for systematic fixes
  - Parallel execution preferred
  - Research code tolerance (Eigen got 9/10 for ideas despite docs)

═══════════════════════════════════════════════════════════════

RESUMPTION: Run /linus-doc-consolidate ${BRILLIANT_MINDS_ROOT}

╚══════════════════════════════════════════════════════════════╝
```

---

This is 90% of the value in 10% of the tokens. Everything else was noise.

$ARGUMENTS
