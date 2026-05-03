Deploy the socratic-kv-signals llama.cpp fork to a GH200 and connect it to the socratic-tuner.

## Pipeline

Execute these steps in order. All commands are non-interactive and agent-safe.

### Step 1: Deploy llama.cpp fork to GH200

For a **fresh GH200** (no llama.cpp, no model):
```bash
~/socratic-tuner/tools/gh200-deploy.py 192.222.51.141 --init --json 2>/dev/null
```

For an **incremental deploy** (code changes only):
```bash
~/socratic-tuner/tools/gh200-deploy.py 192.222.51.141 --json 2>/dev/null
```

The `--init` mode will:
1. SSH to the GH200
2. Generate ed25519 deploy key on remote (skip if exists)
3. Add deploy key to `quivent/llama.cpp` via local `gh` CLI
4. Configure SSH for github.com on remote
5. `git clone --branch socratic-kv-signals` (skip if repo exists)
6. Full CUDA build
7. Download Llama 3.3 70B Q4_K_M (~40GB, skip if exists)
8. Start server with `--flash-attn off` + verify 9 signal types + write manifest

The default (no `--init`) mode will:
1. `git pull origin socratic-kv-signals` on remote
2. Incremental `cmake --build` (~15-30s)
3. Stop + start server
4. Verify + write manifest

If `--json` output reports `"status": "ok"` and `"signal_count": 9`, deployment is complete.

### Step 2: Read the manifest

```bash
cat ~/.conduct/deployments/active.json
```

The manifest contains `endpoint`, `signals.types`, and `signals.request_flags` — everything needed to call the server.

### Step 3: Connect the socratic-tuner (if running)

The socratic-tuner Tauri app reads the manifest via its "Discover" button, which calls `tt_import_conduct_deployment`. This auto-sets `server_url` and enables Full signal extraction preset.

If the Tauri app is not running, the manifest is still the contract. Any tool that reads `~/.conduct/deployments/active.json` knows the endpoint and capabilities.

### Step 4: Verify with a direct API call

```bash
ENDPOINT=$(python3 -c "import json; print(json.load(open('$HOME/.conduct/deployments/active.json'))['endpoint'])")
curl -s "$ENDPOINT/v1/chat/completions" \
  -H 'Content-Type: application/json' \
  -d '{"model":"llama","messages":[{"role":"user","content":"Hello"}],"max_tokens":16,"return_kv_cache":true,"kv_layer_step":8}' \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print('Response:', d['choices'][0]['message']['content'][:100]); print('KV layers:', len(d.get('internals',{}).get('layer_stats',[])))"
```

## Quick commands

| Action | Command |
|--------|---------|
| Fresh setup | `~/socratic-tuner/tools/gh200-deploy.py <host> --init` |
| Fresh + spec decode | `~/socratic-tuner/tools/gh200-deploy.py <host> --init --spec-decode` |
| Incremental deploy | `~/socratic-tuner/tools/gh200-deploy.py <host>` |
| Deploy + spec decode | `~/socratic-tuner/tools/gh200-deploy.py <host> --spec-decode` |
| Status check | `~/socratic-tuner/tools/gh200-deploy.py <host> --status` |
| Signal test | `~/socratic-tuner/tools/gh200-deploy.py <host> --verify` |
| Restart (no build) | `~/socratic-tuner/tools/gh200-deploy.py <host> --restart` |
| Restart + spec decode | `~/socratic-tuner/tools/gh200-deploy.py <host> --restart --spec-decode` |
| Stop server | `~/socratic-tuner/tools/gh200-deploy.py <host> --stop` |
| Tail logs | `~/socratic-tuner/tools/gh200-deploy.py <host> --logs` |

## Environment overrides

```bash
GH200_HOST=10.0.0.5 GH200_PORT=8000 ~/socratic-tuner/tools/gh200-deploy.py --json
```

$ARGUMENTS
