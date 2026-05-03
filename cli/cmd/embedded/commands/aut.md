---
description: Run the Sixth autonomy CLI (aut status, tasks, workers, daemon, etc.)
argument-hint: <subcommand> [args...]
---

# aut — Sixth Autonomy CLI

Run the native Forth autonomy CLI at `~/sixth/packages/aut/bin/aut`.

## Steps

1. Run the command:
```bash
~/sixth/packages/aut/bin/aut $ARGUMENTS
```

2. If the binary doesn't exist or is stale, rebuild first:
```bash
cd ~/sixth/packages/aut && ../../compiler/bin/s3 aut.fs bin/aut && codesign --force --sign - bin/aut
```

3. Report the output to the user.

## Available subcommands
- `status` — daemon status, tier, cycle count
- `tasks` — list/add/show tasks
- `workers` — list/kill/spawn workers
- `daemon` — start/stop/restart daemon
- `monitor` — live TUI dashboard
- `catalog` — list packages, commands, agents
- `grow` — spawn new capabilities via Claude Code
- `health` — deep health check
- `backup` — snapshot/restore state
- `tail` — tail logs
- `bench` — run benchmarks
- `shell` — interactive REPL
- `bootstrap` — first-run setup
- `help` — list all commands
