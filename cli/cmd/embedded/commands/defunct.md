# /defunct - Document Status Checker

Check if a document is defunct (outdated/stale) before citing as fact.

## Instructions

1. Parse the document path from $ARGUMENTS

2. Check THREE sources for defunct status:

   **Source 1: DEFUNCT_INDEX.md**
   - Read `identity/repository/DEFUNCT_INDEX.md`
   - Search for the document path in the table
   - If found → DEFUNCT

   **Source 2: Document Frontmatter**
   - Read the first 20 lines of the target document
   - Check for `status: DEFUNCT` in frontmatter
   - If found → DEFUNCT

   **Source 3: CLAUDE.md Warnings**
   - Check if path matches any entry in "What NOT to Trust" section
   - If found → STALE/DEFUNCT

3. Report status with full details

## Response Format

### If DEFUNCT:
```
DEFUNCT: [path]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: DEFUNCT
Defunct Date: [date if available]
Reason: [reason]
Superseded By: [replacement path or "none"]
Source: [which check found it]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ACTION: Do not cite as current fact.
        May reference as "historical approach" or "previously attempted".
```

### If ACTIVE:
```
ACTIVE: [path]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: ACTIVE
Checks Passed:
  ✓ Not in DEFUNCT_INDEX.md
  ✓ No defunct frontmatter
  ✓ Not in CLAUDE.md warnings
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ACTION: Safe to cite as current fact.
```

### If FILE NOT FOUND:
```
UNKNOWN: [path]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: FILE NOT FOUND
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ACTION: Verify path. Document may have been moved or deleted.
```

## Examples

```
/defunct docs/08-protocols/SOME_PROTOCOL.md
→ DEFUNCT (entire directory is stale)

/defunct PROJECT_STATE.md
→ ACTIVE (authoritative current state)

/defunct docs/04-research/PROPER_TOPOLOGIST_PROTOCOL.md
→ Check index and frontmatter...
```

## System Reference

Full defunct system documented in:
- `identity/repository/DEFUNCT_SYSTEM.md` - System specification
- `identity/repository/DEFUNCT_INDEX.md` - Central registry

$ARGUMENTS
