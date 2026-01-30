# 🎌 anime - Quick Start

No more shell scripts! Use this beautiful Go CLI instead.

## Install

```bash
cd /Users/joshkornreich/lambda

# Build and install
make install

# Or just build
make build
./anime --help
```

## Usage Flow

### 1. Configure Servers (Interactive TUI)

```bash
anime config
```

**What you'll see:**

```
🎌 anime - Lambda Configuration

▶ Manage Servers
  Configure Modules
  API Keys
  Save & Exit

↑/↓: navigate • enter: select • q: quit
```

**Navigate through:**
1. **Manage Servers** → Add your Lambda machines
   - Name: `lambda-gh200-1`
   - Host: `192.168.1.100`
   - User: `ubuntu`
   - SSH Key: `~/.ssh/lambda_key.pem`
   - Cost/hr: `20.0`

2. **Configure Modules** → Select server, then check modules you want:
   ```
   ▶ ☑ Core System (5m) - CUDA, Python, Node.js, Docker
     ☑ PyTorch + AI Libraries (2m) - PyTorch, Transformers, Diffusers, xformers
     ☑ Ollama Server (1m) - Ollama LLM server (no models)
     ☐ Small Models (7B) (8m) - Mistral, Llama 3.3 8B, Qwen3 8B
     ☐ Medium Models (14-34B) (25m) - Qwen3 14B, Qwen3 32B, Mixtral 8x7B, DeepSeek Coder
     ☐ Large Models (70B+) (60m) - Llama 3.3 70B, Qwen3 235B MoE
     ☐ ComfyUI (2m) - Stable Diffusion UI with Manager
     ☑ Claude Code CLI (1m) - Anthropic Claude Code CLI

   Estimated cost: $3.00 @ $20/hr

   ↑/↓: navigate • space: toggle • enter: save • esc: cancel
   ```

3. **API Keys** (Optional) → Add API keys
   - Anthropic
   - OpenAI
   - HuggingFace
   - Lambda Labs

4. **Save & Exit**

### 2. Deploy to Server

```bash
anime deploy lambda-gh200-1
```

**You'll see live progress:**

```
🎌 anime - Installing on lambda-gh200-1

System Information:
  OS: Ubuntu 22.04.3 LTS
  GPU: NVIDIA GH200 480GB
  Free Disk: 1.2T | Free RAM: 450G

Installation Progress:

  ✓ Core System - Complete
  ▶ PyTorch + AI Libraries - Installing
  ⏳ Ollama Server - Pending
  ⏳ Claude Code CLI - Pending

Recent output:
  Installing PyTorch with CUDA support...
  Successfully installed torch-2.2.0+cu124

Elapsed: 7m 23s | Cost: $2.46

Press Ctrl+C to cancel
```

### 3. Check Server Status

```bash
anime status lambda-gh200-1
```

**Output:**

```
Connecting to lambda-gh200-1...

System Information:
  OS: Ubuntu 22.04.3 LTS
  Architecture: aarch64
  Kernel: 5.15.0-91-generic
  GPU: NVIDIA GH200 480GB
  Free Disk: 1.2T
  Free Memory: 450G

Installed Components:
  Python: ✓ 3.10.12
  Node.js: ✓ v20.11.0
  Docker: ✓ Docker version 25.0.3
  NVIDIA: ✓ NVIDIA-SMI 550.54.15
  CUDA: ✓ release 12.4, V12.4.131
  PyTorch: ✓ 2.2.0+cu124
  Ollama: ✓ ollama version 0.1.26
  ComfyUI: ✗ Not installed
  Claude Code: ✓ 1.0.0

Ollama Models:
NAME                    ID              SIZE      MODIFIED
llama3.3:8b            abc123          4.7 GB    2 hours ago
mistral:latest         def456          4.1 GB    2 hours ago
qwen3:8b              ghi789          5.2 GB    2 hours ago
```

### 4. List All Servers

```bash
anime list
```

**Output:**

```
Configured servers:

  lambda-gh200-1 (ubuntu@192.168.1.100)
    Cost: $20.00/hr
    Modules: 4 configured
    Estimated deployment: $3.00

  lambda-gh200-2 (ubuntu@192.168.1.101)
    Cost: $18.00/hr
    Modules: 7 configured
    Estimated deployment: $12.50
```

## Example Workflows

### Minimal Setup (Testing, ~$3)

```bash
anime config
# Select: Core + PyTorch
# Time: 7 minutes

anime deploy test-server
```

### Standard Setup (Production, ~$8)

```bash
anime config
# Select: Core + PyTorch + Ollama + Small Models
# Time: 16 minutes

anime deploy prod-server
```

### Full Setup (Everything, ~$25)

```bash
anime config
# Select: All modules including Large Models
# Time: 60+ minutes

anime deploy full-server
```

## Tips

1. **Start Small** - Select Core + PyTorch first ($3, 7 min)
2. **Test Connection** - Run `anime status SERVER` before deploying
3. **Cost Aware** - Check estimated cost in TUI before saving
4. **Multiple Servers** - Configure different profiles for different purposes
5. **Watch Progress** - The TUI shows real-time progress and costs

## Keyboard Shortcuts

### Main Menu
- `j/k` or `↑/↓` - Navigate
- `Enter` - Select
- `q` - Quit

### Server/Module Lists
- `Space` - Toggle (module selection)
- `d` - Delete (server list)
- `Esc` - Go back

### Forms
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `Enter` - Submit
- `Esc` - Cancel

## Configuration File

After using `anime config`, your config is saved to:

```
~/.config/anime/config.yaml
```

You can edit it manually if needed:

```yaml
servers:
  - name: lambda-gh200-1
    host: 192.168.1.100
    user: ubuntu
    ssh_key: ~/.ssh/lambda_key.pem
    cost_per_hour: 20.0
    modules:
      - core
      - pytorch
      - ollama
      - models-small
      - claude
```

## Troubleshooting

### Can't connect to server

```bash
# Test SSH manually
ssh -i ~/.ssh/lambda_key.pem ubuntu@192.168.1.100

# Check permissions
chmod 600 ~/.ssh/lambda_key.pem
```

### Module installation failed

```bash
# Check server status
anime status lambda-gh200-1

# SSH in and check logs
ssh ubuntu@YOUR_IP
tail -f /tmp/anime-install-*.sh
```

### TUI not displaying correctly

```bash
# Ensure terminal supports colors
echo $TERM  # Should be xterm-256color or similar

# Try different terminal
# iTerm2, Warp, or modern terminal recommended
```

## Build from Source

```bash
cd /Users/joshkornreich/lambda

# Get dependencies
go mod download

# Build
go build -o anime

# Run
./anime config
```

## Comparison: Before & After

### Before (Shell Scripts)

```bash
# Copy script
cat setup-core.sh | pbcopy

# SSH in
ssh ubuntu@192.168.1.100

# Paste and pray
bash
# (paste)

# Wait 60 minutes
# Hope nothing breaks
# No idea how much it costs
# Can't cancel safely
```

### After (anime CLI)

```bash
# Configure once
anime config
  # Beautiful TUI
  # See costs before committing
  # Save multiple servers

# Deploy anytime
anime deploy lambda-gh200-1
  # Real-time progress
  # Live cost tracking
  # Cancel anytime (Ctrl+C)

# Check status
anime status lambda-gh200-1
  # See what's installed
  # Check GPU status
  # List models
```

## Next Steps

1. **Build and install**: `make install`
2. **Configure your first server**: `anime config`
3. **Deploy**: `anime deploy SERVER_NAME`
4. **Check status**: `anime status SERVER_NAME`

Enjoy! No more shell scripts 🎉
