Replay the previous sequence of slash commands that were executed in this session.

Usage: After running a series of commands (e.g., `/audit`, `/fix`, `/audit`), use `/repeat` to replay that exact sequence again.

**Repeat Protocol:**

1. Look back through this conversation for all slash commands that were executed
2. Collect them in order of execution
3. Replay each one sequentially, waiting for completion before proceeding to the next
4. Skip `/repeat` itself to avoid infinite loops

**Behavior:**
- If no previous commands were found in the conversation, report: "No commands to repeat."
- Commands are replayed with their original arguments
- Each replayed command runs to full completion before the next begins

**Example:**
If the previous session contained:
```
/audit README.md
/fix the broken links
/audit README.md
```

Then `/repeat` would execute those three commands again in order.

Replay the previous command sequence now.
