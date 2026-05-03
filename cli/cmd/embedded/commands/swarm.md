# /swarm — 24-agent Qwen swarm

**Argument:** `$ARGUMENTS` (required — the prompt/question to swarm on)

## Endpoint

```bash
ENDPOINT="${QWEN_ENDPOINT:-http://192.222.51.2:8001}"
MODEL="qwen3.5-27b"
```

Check `~/hive/.swarm/ENDPOINTS.toml` if `$QWEN_ENDPOINT` is unset. Never hardcode the IP. Never use Anthropic APIs for swarm agents.

## Protocol

Fire **exactly 24** parallel curl requests in a **single Bash call** using `&` and `wait`. Do NOT use the Task tool.

```bash
mkdir -p .swarm/responses .swarm/findings .swarm/signals

fire() {
  local id=$(printf '%03d' $1) angle="$2"
  local payload=$(jq -n --arg c "You are swarm agent $1/24. Angle: $angle

PROBLEM: $ARGUMENTS

Give ONE specific, actionable finding. Be concrete. Under 400 words." \
    '{"model":"qwen3.5-27b","messages":[{"role":"user","content":$c}],"temperature":0.75,"max_tokens":1024,"chat_template_kwargs":{"enable_thinking":false}}')
  curl -s "$ENDPOINT/v1/chat/completions" -H "Content-Type: application/json" -d "$payload" \
    > ".swarm/responses/agent-${id}.json" && touch ".swarm/signals/DONE-agent-${id}" &
}

# 6 groups × 4 agents — adapt angles to the actual problem:
# A (1-4):  Execution / parse path
# B (5-8):  Root cause hypotheses
# C (9-12): Language/compiler traps
# D (13-16): Data flow analysis
# E (17-20): Edge cases
# F (21-24): Fix proposals / alternatives

fire 1 "Trace the execution path" && fire 2 "..." && ... && fire 24 "..."
wait
```

**Before dispatch:** read relevant source files first; embed actual code in prompts.

## Phases

1. **DISPATCH** — fire all 24, write responses to `.swarm/responses/agent-NNN.json`
2. **SYNTHESIZE** — read all responses, write `.swarm/synthesis.md` (themes, contradictions, ranked findings)
3. **REPORT** — write `.swarm/report.md`; act on the top finding immediately

Log all actions to `.swarm/swarm.log`. Maintain `.swarm/progress.json`.

## Rules

1. Exactly 24 agents. Always. Single Bash `&`/`wait` call.
2. All traffic to `$ENDPOINT` (qwen3.5-27b). No other endpoint.
3. Log everything to `.swarm/swarm.log`.
4. `$ARGUMENTS` drives the prompt — if empty, ask the user.
5. Act on results — implement the top fix after synthesizing.

---

Execute now. Prompt: `$ARGUMENTS`
