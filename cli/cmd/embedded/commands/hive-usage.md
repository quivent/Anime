# /hive-usage — show swarm-task dispatch instructions

Present the beekeeper with a concise, live-status guide to the hive's
swarm-task system. This is reference, not cognition — keep the output
faithful to the source, skip any ceremony, and do NOT propose work.

## What to do

1. Call `swarm_health()` to get the live antennae status.
2. Render the output below, substituting the live status at the top.
3. Include the two inspection commands at the bottom.
4. Stop. Do not offer follow-up actions. Do not propose work.

If the `hive-swarm` MCP tools are not loaded in this session (fresh
session needed), say so briefly at the top and still show the rest.

---

## Output template (render verbatim, adapting only the status line)

```
⚇  Swarm dispatch — hive-swarm MCP + apps/hive-wails Swarm tab

Antennae: <from swarm_health: reachable/down + endpoint + models>

Four tools (Claude Code + app, same backend):

  swarm_task(description, prompt, count=1, synthesize=False)
      Claude Task equivalent. N parallel Qwen workers. synthesize=True
      adds one consolidation pass (the "1 to integrate" allowance).

  swarm_scout(question, count=3)
      Structured scouts — each returns {finding, confidence,
      dance_vector{direction, distance, quality}, failure_mode}.

  swarm_think(question)
      One-shot fastest path. No parallelism, no synthesis.

  swarm_health()
      Ping /v1/models. Returns {endpoint, reachable, models[]}.

Routing rule:

  Tier 3/4 → swarm_*   (file search, pattern extraction, routine
                        summarization, light audits, classification)
  Tier 1/2 → Task      (frontier judgment, charter-grade decisions,
                        multi-file code synthesis)

Where it lives:

  MCP server            ~/.claude/mcp-servers/hive-swarm/server.py
  Registration          ~/.claude/config.json (mcpServers.hive-swarm)
  Per-Task hook         ~/.claude/hooks/task-swarm-nudge.sh
  Session policy hook   ~/.claude/hooks/swarm-policy-reminder.sh
  Audit log             ~/.claude/audit/task-calls.jsonl
  App surface           apps/hive-wails — click the Swarm tab
  Run artifacts         ~/hive/.swarm/swarm-runs/<kind>-<desc>-<ts>/
  Operator guide        HIVE.md  (full detail, troubleshooting, tuning)
  Agent policy          CLAUDE.md § "Swarm dispatch — routing policy"

Inspect usage:

  tail -f ~/.claude/audit/task-calls.jsonl          # every Task call, live
  ls -lt ~/hive/.swarm/swarm-runs/ | head           # recent swarm runs
```

## Hard constraints on this command

- No proposals. No "want me to…?". No greeting.
- Do not dispatch a swarm call to demonstrate usage. Only `swarm_health`.
- If the user wants more detail, point them at `HIVE.md`.
