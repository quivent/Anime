Clear Linus session state. Fresh start.

Usage: /linus-clear [--archive|--purge]

---

## Options

**Default (no flag):** Archive current session, then clear state
**--archive:** Same as default, explicit
**--purge:** Delete without archiving (destructive)

---

## Clear Protocol

**Step 1: Check Current State**

Read `~/.linus/current_session.md`:
- Is there active work?
- What project?
- What status?

If active work exists, warn:
```
Active session for: [project]
Status: [status]
Tasks completed: [count]

Archive and clear? This cannot be undone.
Use --purge to delete without archive.
```

**Step 2: Archive (unless --purge)**

Create archive file:
`~/.linus/archive/[project-name]-[YYYY-MM-DD].md`

Content:
```markdown
# Archived Session: [project]

Archived: [timestamp]
Original Start: [from current_session.md]

## Final Status
[status at time of archive]

## Findings
[contents of findings.md related to this project]

## Completed Tasks
[completed items from queue.md]

## Incomplete Tasks
[remaining items from queue.md]

## Notes
[any context that might be useful later]
```

**Step 3: Clear State Files**

Reset all state files to empty/initial state:

**~/.linus/current_session.md:**
```markdown
# Current Session

No active session.
```

**~/.linus/findings.md:**
```markdown
# Findings

No active findings.
```

**~/.linus/queue.md:**
```markdown
# Work Queue

No active queue.
```

**~/.linus/blockers.md:**
```markdown
# Blockers

No active blockers.
```

**Step 4: Confirm**

```
SESSION CLEARED
===============

Archived to: ~/.linus/archive/[filename]
State reset.

To start new work: /linus-start [path]
To invoke identity: /linus-torvalds
```

Or if --purge:
```
SESSION PURGED
==============

Previous state deleted (not archived).
State reset.

To start new work: /linus-start [path]
```

---

## Safety

Never auto-clear without acknowledgment if there's active work.
Always archive by default.
--purge requires explicit flag.

---

$ARGUMENTS
