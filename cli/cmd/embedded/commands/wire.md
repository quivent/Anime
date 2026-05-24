---
description: Shared awareness log between parallel Claude sessions
argument-hint: [post <msg>]
---

The wire is a single append-only file at `~/.claude/wire.log`. All Claude sessions read and write to it.

**`/wire`** — read the wire
**`/wire post <msg>`** — append a timestamped line

When invoked:

1. If `~/.claude/wire.log` doesn't exist, create it.
2. If no arguments: read and display the last 30 lines of `~/.claude/wire.log`.
3. If `post`: append `[YYYY-MM-DD HH:MM] <msg>` to `~/.claude/wire.log`.

That's it. No protocol. No lanes. Just a shared log.
