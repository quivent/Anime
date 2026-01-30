---
description: Check MLX model server status, start/stop server, and manage loaded models
argument-hint: [status|start|stop|restart] [--model llama-70b] [--quant 4bit] [--no-draft]
allowed-tools: Bash, Read
model: claude-sonnet-4-20250514
---

Manage the MLX model server for local LLM inference.

**Action:** ${1:-status}
**Options:** $ARGUMENTS

## Commands

### Status (default)
Check if the model server is running and what's loaded:
```bash
# Check if server is responding
curl -s http://127.0.0.1:8000/v1/models 2>/dev/null || echo "Server not running"

# Check for running processes
ps aux | grep -E "mlx_lm|eigen loader" | grep -v grep
```

### Start
Start the MLX model server with speculative decoding:
```bash
cd /Users/joshkornreich/Eigen && ./eigen loader serve
```

Default configuration:
- Model: Llama 3.3 70B 4-bit (`mlx-community/Llama-3.3-70B-Instruct-4bit`)
- Draft model: Llama 3.2 1B 4-bit (for speculative decoding)
- Endpoint: http://127.0.0.1:8000/v1/chat/completions

### Stop
Stop the running model server:
```bash
pkill -f "mlx_lm.server" || pkill -f "eigen loader serve"
```

### Restart
Stop then start the server.

## Options
- `--model`: Model to load (llama-70b, llama-8b, qwen-72b, mistral-large)
- `--quant`: Quantization level (3bit, 4bit, 8bit, fp16)
- `--no-draft`: Disable speculative decoding

## Quick Test
After starting, test with:
```bash
curl -s http://127.0.0.1:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"Hello"}],"stream":false}' | jq -r '.choices[0].message.content'
```

Execute the requested action and report results.
