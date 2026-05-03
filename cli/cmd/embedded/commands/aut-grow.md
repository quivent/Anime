---
description: Self-grow — spawn Claude Code workers (with auto-detected brilliant mind subagents)
argument-hint: [--workers N] [--mind <name>] [--no-mind] <description>
---

Deploy autonomous workers to build something new:

```bash
~/sixth/packages/aut/bin/aut grow $ARGUMENTS
```

## Auto-mind detection
By default, `grow` scans the task description for keywords and auto-attaches a relevant brilliant mind's IDENTITY.md as the worker's system prompt:

- "compiler" → chris_lattner
- "forth" → chuck_moore
- "benchmark" / "perform" → brendan_gregg
- "kernel" / "git" → linus_torvalds
- "neural" / "deep learning" → geoffrey_hinton
- "graphics" / "game engine" → john_carmack
- "first principles" → richard_feynman
- ...and 20+ more keyword routes

## Flags
- `--workers N` — spawn N parallel workers (default 1)
- `--mind <name>` — explicit mind override (e.g. `--mind alan_kay`)
- `--no-mind` — disable auto-detection, spawn vanilla worker

## Examples
```
aut grow add a JSON parser to lib/                       # auto-detects nothing → vanilla
aut grow optimize the cache layout                       # → mike_acton
aut grow --workers 3 refactor the compiler frontend      # → 3× chris_lattner
aut grow --mind chuck_moore rewrite the dispatch table   # explicit
aut grow --no-mind quick fix to the readme               # vanilla
```

List all available minds: `aut minds`

