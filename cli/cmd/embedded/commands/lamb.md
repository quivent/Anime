# /lamb - Llama Agent Multi-Brain

Parallel swarm queries to Llama 3.3 70B via vLLM, synthesized locally.

## Endpoint

```
URL:   http://132.145.129.171:8000/v1/chat/completions
Model: meta-llama/Llama-3.3-70B-Instruct
```

## Usage

```
/lamb <task or question>
/lamb 5 <task>              # explicit swarm size (default: 10)
/lamb raw <prompt>          # single query, no decomposition
/lamb chain <task>          # sequential: each query sees prior results
```

## Protocol

You are the LAMB coordinator. You decompose tasks, dispatch parallel queries to a remote Llama 70B instance, collect results, and synthesize them.

### Step 1: Parse Arguments

From `$ARGUMENTS`:
- If first token is a number N (2-20), use N as swarm size, rest is the task
- If first token is `raw`, send a single direct query with no decomposition
- If first token is `chain`, run sequential chain mode (see below)
- Otherwise, default swarm size = 10, entire argument is the task

### Step 2: Decompose (Swarm Mode)

Analyze the task and decompose into N **distinct, non-overlapping angles**. Each sub-query should:
- Attack a different facet of the problem
- Be self-contained (no dependency on other sub-queries)
- Include enough context for Llama to produce a useful standalone answer

Write decomposition to the user before dispatching:

```
=== LAMB SWARM ===
Task: [original task]
Swarm size: N
Queries:
  1. [angle/sub-question 1]
  2. [angle/sub-question 2]
  ...
Dispatching...
```

### Step 3: Dispatch Parallel Queries

For each sub-query, spawn a **parallel Task agent** (subagent_type: "Bash"). Each agent runs:

```bash
curl -s --max-time 120 http://132.145.129.171:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "messages": [
      {"role": "system", "content": "You are a precise, thorough analyst. Answer the question directly and completely. Be specific, cite reasoning, avoid filler."},
      {"role": "user", "content": "<THE SUB-QUERY>"}
    ],
    "temperature": 0.7,
    "max_tokens": 4096
  }'
```

**IMPORTANT:**
- All N curl commands MUST be dispatched in parallel (separate Task agents in a single message)
- Escape all JSON properly - use single quotes for the outer -d argument, double quotes inside
- If a query contains quotes or special characters, escape them for the JSON payload
- Each Task agent should parse the JSON response and return ONLY the `choices[0].message.content` text. Use: `curl ... | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['choices'][0]['message']['content'])"`

### Step 4: Collect & Synthesize

Once all agents return, present results and synthesize:

```
=== LAMB RESULTS ===

--- Query 1: [angle] ---
[Llama's response]

--- Query 2: [angle] ---
[Llama's response]

...

=== SYNTHESIS ===
[Your analysis combining all Llama responses:
 - Where they agree (high confidence)
 - Where they diverge (flag for investigation)
 - What's missing (gaps none of them covered)
 - Actionable conclusions
 - Anything the responses got wrong that you can correct]
```

The synthesis is the most important part. You (Claude) have stronger reasoning than the Llama responses - your job is to **upgrade** their combined output, not just concatenate it.

### Step 5: Follow-Up

After synthesis, ask: "Want me to drill deeper on any of these angles, or send a follow-up swarm?"

---

## Chain Mode

When `chain` is the first argument:

1. Send initial query to Llama
2. Analyze the response
3. Formulate a follow-up query that builds on what Llama said (filling gaps, pushing deeper)
4. Send follow-up to Llama
5. Repeat for 3-5 rounds or until diminishing returns
6. Synthesize the full chain into a final output

Each query in the chain includes the prior exchange as context:
```json
{
  "messages": [
    {"role": "system", "content": "You are a precise analyst. Build on the prior discussion."},
    {"role": "user", "content": "[original question]"},
    {"role": "assistant", "content": "[llama's prior response]"},
    {"role": "user", "content": "[follow-up question from Claude]"}
  ]
}
```

---

## Error Handling

- **Timeout (>120s):** Report which query timed out, synthesize from the ones that returned
- **Connection refused:** Report endpoint down, do not retry indefinitely
- **Malformed response:** Show raw response, skip that query in synthesis
- **Empty response:** Note it, proceed with others

---

## Notes

- Temperature 0.7 is the default. For factual/precise tasks, the coordinator may lower to 0.3. For creative/divergent tasks, raise to 0.9.
- Max tokens 4096 is the default. For short answers, reduce to save time.
- The coordinator (Claude) always has final say on synthesis quality. Llama provides raw material; Claude refines it.

---

$ARGUMENTS
