---
description: Register and verify project requirements that must hold across changes
---

# /enforce - Requirement Enforcement Protocol

**Arguments:** $ARGUMENTS

## Behavior

### If arguments are provided: REGISTER a new enforcement

1. Read the enforcements file at `~/.claude/projects/{project}/enforcements.json` (create if missing)
2. Parse the argument as a natural language requirement
3. For each requirement, determine:
   - `id`: Short kebab-case identifier (e.g., `tray-icon-persists`)
   - `requirement`: The full requirement text as stated by the user
   - `check`: A concrete verification strategy — what files to read, what patterns to grep for, what commands to run
   - `files`: List of files relevant to this enforcement
   - `registered`: ISO timestamp
4. Append to the enforcements array and save
5. Immediately verify the new enforcement passes
6. Report: registered + current status (PASS/FAIL)

### If no arguments provided: CHECK all enforcements

1. Read `~/.claude/projects/{project}/enforcements.json`
2. For EACH enforcement, perform its verification:
   - Read the relevant files
   - Check for the required patterns, logic, or behavior
   - Run any specified commands (compile checks, grep, etc.)
3. Report results as a table:

```
ENFORCEMENT STATUS
─────────────────────────────────────
[PASS] tray-icon-persists    Tray icon stays when window closed
[PASS] dual-port-probe       server_status checks both 8741 and 8000
[FAIL] chat-uses-correct-port  Chat requests use detected endpoint
       ↳ client.ts:135 still hardcodes port 8000
─────────────────────────────────────
Result: 2/3 passing
```

4. For any FAIL: explain what's wrong and where, with file:line references
5. If ALL pass: confirm all enforcements hold

### Special subcommands

- `/enforce list` — Show all registered enforcements without checking
- `/enforce remove <id>` — Remove an enforcement by ID
- `/enforce clear` — Remove all enforcements (ask confirmation first)

## Verification strategies

Use the lightest verification that catches regressions:

| Type | Method |
|------|--------|
| **Pattern exists** | Grep for required code pattern in specified file |
| **Pattern absent** | Grep confirms dangerous pattern is NOT present |
| **File exists** | Glob for required file |
| **Compiles** | Run `cargo check` or equivalent |
| **Logic check** | Read file, verify control flow matches requirement |
| **Config check** | Read config file, verify setting present |

## Storage format

```json
{
  "enforcements": [
    {
      "id": "tray-icon-persists",
      "requirement": "App icon on menu bar must persist when window is closed",
      "check": {
        "type": "pattern_exists",
        "file": "src-tauri/src/lib.rs",
        "pattern": "CloseRequested.*prevent_close|prevent_close.*CloseRequested",
        "description": "Window close handler prevents default close and hides instead"
      },
      "files": ["src-tauri/src/lib.rs"],
      "registered": "2026-02-14T12:00:00Z"
    }
  ]
}
```

## Rules

- Enforcements persist across sessions via the JSON file
- Each enforcement must have a concrete, automatable check — no subjective criteria
- When registering, immediately verify it passes (catch bad registrations early)
- When checking, read actual code — don't trust cached state
- Keep enforcement IDs short and descriptive
- The project path for the JSON file is derived from the current working directory
