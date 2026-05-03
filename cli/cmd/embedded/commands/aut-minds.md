---
description: List brilliant minds available as worker subagents
---

Show all 85+ brilliant mind identities that can be auto-attached to spawned workers:

```bash
~/sixth/packages/aut/bin/aut minds
```

These are at `~/benchmarks/Work/brilliant_minds/minds/` — each has an IDENTITY.md that gets prepended to a worker's system prompt when relevant keywords appear in the task description.

Auto-detection examples:
- prompt mentions "compiler" → chris_lattner
- prompt mentions "forth" → chuck_moore
- prompt mentions "benchmark" → brendan_gregg
- prompt mentions "kernel" → linus_torvalds
- prompt mentions "neural" → geoffrey_hinton

Override with `aut grow --mind <name> "task"` or disable with `--no-mind`.
