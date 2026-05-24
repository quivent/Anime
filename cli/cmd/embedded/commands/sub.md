# /sub — Fast Subagent Dispatch

Parse `$ARGUMENTS` for an optional model name, optional count, and the task.

## Argument parsing

Format: `[model] [count] TASK`

- **model**: `opus`, `sonnet`, or `haiku` (case-insensitive). Default: `opus`.
- **count**: integer 1-50. Default: `1`.
- **TASK**: everything remaining after model/count are consumed.

The first word is checked: if it matches a model name, consume it. The next word is checked: if it's a number, consume it as count. Everything else is the task.

Examples:
- `/sub find all TODO comments` → 1 Opus agent, task = "find all TODO comments"
- `/sub sonnet find all TODO comments` → 1 Sonnet agent, task = "find all TODO comments"
- `/sub sonnet 10 find all TODO comments` → 10 Sonnet agents, task = "find all TODO comments"
- `/sub haiku 5 search for unused imports` → 5 Haiku agents
- `/sub 3 check these files` → 3 Opus agents (count without model)

## Execution

**Prioritize speed. Dispatch immediately. No preamble, no planning, no confirmation.**

Use the `Agent` tool with `subagent_type: "general-purpose"`.

- Set `model` to the parsed model (omit if opus, since it's the default).
- If count is 1: launch one Agent with the full task as the prompt.
- If count > 1: launch ALL agents in a **single message** (parallel tool calls). Give each agent the same task prompt. Add `"You are agent N of M. Focus on different areas than the others."` to help them spread coverage.
- Set a clear, short `description` derived from the first few words of the task.

When agents return, synthesize their results into a concise briefing. If multiple agents, deduplicate and merge findings.
