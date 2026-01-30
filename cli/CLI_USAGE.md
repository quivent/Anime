# anime - Simple CLI Usage

**No TUI needed!** Use simple commands instead.

## Add a Server

```bash
anime add --name lambda-1 --host 192.168.1.100 --user ubuntu --key ~/.ssh/lambda.pem --cost 20
```

**Short version:**
```bash
anime add -n lambda-1 -H 192.168.1.100 -u ubuntu -k ~/.ssh/lambda.pem -c 20
```

**Minimal (uses defaults):**
```bash
anime add -n lambda-1 -H 192.168.1.100
# Defaults: user=ubuntu, key=~/.ssh/id_rsa, cost=20.0
```

## Set Modules (No TUI!)

```bash
# See all available modules
anime list-modules

# Set modules via CLI
anime set-modules lambda-1 core pytorch ollama models-small

# Minimal setup
anime set-modules lambda-1 core pytorch

# Full setup
anime set-modules lambda-1 core pytorch ollama models-large comfyui claude
```

## List Everything

```bash
# List servers
anime list

# List available modules
anime list-modules
```

## Deploy

```bash
anime deploy lambda-1
```

## Check Status

```bash
anime status lambda-1
```

## Remove Server

```bash
anime remove lambda-1
# or
anime rm lambda-1
```

## Complete Example

```bash
# 1. Add server
anime add -n my-server -H 192.168.1.100

# 2. Set modules
anime set-modules my-server core pytorch ollama models-small

# 3. Deploy
anime deploy my-server

# 4. Check status
anime status my-server
```

## Quick Combos

### Minimal ($3, 7min)
```bash
anime add -n test -H 192.168.1.100
anime set-modules test core pytorch
anime deploy test
```

### Standard ($8, 16min)
```bash
anime add -n prod -H 192.168.1.101
anime set-modules prod core pytorch ollama models-small
anime deploy prod
```

### Full ($25, 60min)
```bash
anime add -n full -H 192.168.1.102
anime set-modules full core pytorch ollama models-large comfyui claude
anime deploy full
```

## Available Modules

- **core** - CUDA, Python, Node.js, Docker (5m)
- **pytorch** - PyTorch + AI libraries (2m)
- **ollama** - Ollama server (1m)
- **models-small** - 7B models (8m)
- **models-medium** - 14-34B models (25m)
- **models-large** - 70B+ models (40m)
- **comfyui** - ComfyUI (2m)
- **claude** - Claude Code CLI (1m)

## Interactive Module Picker (Optional)

If you want a SIMPLE interactive picker:

```bash
anime modules lambda-1
```

This opens a simplified UI where you just type numbers:
```
Select modules for: lambda-1

Available modules:

  [ ] 1. Core System (5m) - CUDA, Python, Node.js, Docker
  [ ] 2. PyTorch + AI Libraries (2m) - PyTorch, Transformers, Diffusers, xformers
  [ ] 3. Ollama Server (1m) - Ollama LLM server (no models)
  [ ] 4. Small Models (7B) (8m) - Mistral, Llama 3.3 8B, Qwen3 8B
  [ ] 5. Medium Models (14-34B) (25m) - Qwen3 14B, Qwen3 32B, Mixtral 8x7B, DeepSeek Coder
  [ ] 6. Large Models (70B+) (60m) - Llama 3.3 70B, Qwen3 235B MoE
  [ ] 7. ComfyUI (2m) - Stable Diffusion UI with Manager
  [ ] 8. Claude Code CLI (1m) - Anthropic Claude Code CLI

Type numbers separated by spaces (e.g., '1 2 3 8'), then press Enter:
>
```

Just type: `1 2 3 8` and press Enter. Done!

## Help

```bash
anime --help              # Main help
anime add --help          # Add server help
anime set-modules --help  # Set modules help
anime deploy --help       # Deploy help
```

## All Commands

```bash
anime add              # Add server
anime remove           # Remove server
anime list             # List servers
anime list-modules     # List available modules
anime set-modules      # Set modules (CLI)
anime modules          # Simple module picker (TUI)
anime deploy           # Deploy to server
anime status           # Check server status
anime config           # Full TUI (if you want it)
```

**Pro tip:** You can do EVERYTHING without the TUI using simple CLI commands!
