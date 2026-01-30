# Try It Right Now! 🎌

## 1. Build and Install (30 seconds)

```bash
cd /Users/joshkornreich/lambda
make install
```

You should see:
```
go build -o anime main.go
sudo mv anime /usr/local/bin/anime
✓ anime installed to /usr/local/bin/anime
```

## 2. Try the Interactive TUI (No Server Required!)

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

### Try These Actions:

#### Add a Test Server
1. Press `Enter` on "Manage Servers"
2. Navigate down to "▶ Add new server" and press `Enter`
3. Fill in test data:
   - **Name:** Type `my-test-server`
   - Press `Tab`
   - **Host:** Type `192.168.1.100` (fake is fine)
   - Press `Tab`
   - **User:** Type `ubuntu`
   - Press `Tab`
   - **SSH Key:** Type `~/.ssh/id_rsa`
   - Press `Tab`
   - **Cost/hr:** Type `20.0`
   - Press `Enter` to save

#### Select Modules (See Live Cost Estimates!)
1. You should be back at the server list
2. Navigate to your server and press `Enter`
3. You'll see the module selection screen:

```
Select modules for my-test-server:

▶ ☐ Core System (5m) - CUDA, Python, Node.js, Docker
  ☐ PyTorch + AI Libraries (2m) - PyTorch, Transformers, Diffusers, xformers
  ☐ Ollama Server (1m) - Ollama LLM server (no models)
  ☐ Small Models (7B) (8m) - Mistral, Llama 3.3 8B, Qwen3 8B
  ☐ Medium Models (14-34B) (25m) - Qwen3 14B, Qwen3 32B, Mixtral 8x7B, DeepSeek Coder
  ☐ Large Models (70B+) (60m) - Llama 3.3 70B, Qwen3 235B MoE
  ☐ ComfyUI (2m) - Stable Diffusion UI with Manager
  ☐ Claude Code CLI (1m) - Anthropic Claude Code CLI

Estimated cost: $0.00 @ $20/hr

↑/↓: navigate • space: toggle • enter: save • esc: cancel
```

4. Press `Space` to check "Core System" → Watch cost update to **$1.67**
5. Press `Down` and `Space` to check "PyTorch" → Watch cost update to **$2.33**
6. Press `Down` and `Space` to check "Ollama" → Watch cost update to **$2.67**
7. Press `Down` and `Space` to check "Small Models" → Watch cost update to **$5.33**
8. Press `Space` again to uncheck → Watch cost go down
9. Play with toggling modules and watching the cost!

#### Add API Keys (Optional)
1. Press `Esc` to go back
2. Press `Esc` again to reach main menu
3. Navigate to "API Keys" and press `Enter`
4. Try typing an API key (it will show as `•••`)
5. Press `Tab` to move between fields

#### Save and Exit
1. Press `Esc` until you're at the main menu
2. Navigate to "Save & Exit" and press `Enter`

## 3. See Your Configuration

```bash
anime list
```

You'll see:
```
Configured servers:

  my-test-server (ubuntu@192.168.1.100)
    Cost: $20.00/hr
    Modules: 3 configured
    Estimated deployment: $2.67
```

## 4. View Your Config File

```bash
cat ~/.config/anime/config.yaml
```

You'll see:
```yaml
servers:
- name: my-test-server
  host: 192.168.1.100
  user: ubuntu
  ssh_key: ~/.ssh/id_rsa
  cost_per_hour: 20
  modules:
  - core
  - pytorch
  - ollama
api_keys:
  anthropic: ""
  openai: ""
  huggingface: ""
  lambda_labs: ""
```

## 5. Test Help Command

```bash
anime --help
```

```bash
anime deploy --help
```

```bash
anime status --help
```

## 6. When You Have a Real Server

### Configure It

```bash
anime config
```

1. Add your real server:
   - Name: `lambda-gh200-1`
   - Host: `<your-actual-IP>`
   - User: `ubuntu`
   - SSH Key: `~/.ssh/lambda_key.pem`
   - Cost/hr: `20.0`

2. Select modules (start small!):
   - ✅ Core System
   - ✅ PyTorch
   - Maybe ✅ Ollama
   - Maybe ✅ Claude Code

3. Save & Exit

### Deploy It

```bash
anime deploy lambda-gh200-1
```

You'll see:
```
🎌 anime - Installing on lambda-gh200-1

Connecting to lambda-gh200-1...

System Information:
  OS: Ubuntu 22.04.3 LTS
  Architecture: aarch64
  GPU: NVIDIA GH200 480GB
  Free Disk: 1.2T
  Free Memory: 450G

Installation Progress:

  ▶ Core System - Starting
  ⏳ PyTorch + AI Libraries - Pending
  ⏳ Ollama Server - Pending
  ⏳ Claude Code CLI - Pending

Recent output:
  ==> Installing Core System
  ==> Updating system packages

Elapsed: 0m 23s | Cost: $0.08

Press Ctrl+C to cancel
```

### Check Status

```bash
anime status lambda-gh200-1
```

You'll see:
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
```

## 7. Play with the TUI

The TUI is beautiful and interactive. Try:

### Navigation
- `j/k` (vim keys) or arrow keys
- `Space` to toggle checkboxes
- `Tab` to move between form fields
- `Enter` to confirm
- `Esc` to go back
- `q` to quit (from main menu)

### Watch Costs Update Live
- Toggle modules on/off
- Watch the "Estimated cost" line update instantly
- Dependencies are auto-calculated!

### Multiple Servers
- Add multiple test servers
- Give them different configurations
- Switch between them easily

### API Keys
- Enter API keys (they show as `•••`)
- Secure storage in `~/.config/anime/config.yaml` (0600 permissions)

## 8. Test Without a Server

You can use `anime config` and `anime list` without any server at all!

Just skip the `anime deploy` and `anime status` commands until you have a real Lambda instance.

## What to Notice

### Beautiful TUI
- Colors! (green for selected, blue for info, red for errors)
- Icons! (▶ for current, ☑/☐ for checkboxes, ✓/✗ for status)
- Smooth navigation
- Clear instructions at the bottom

### Real-Time Cost Calculation
- Toggles modules: cost updates instantly
- Includes dependencies automatically
- Shows both time and cost

### Type Safety
- All Go code (no bash fragility)
- Proper error handling
- Can't make syntax errors in config

### Configuration Management
- YAML storage
- Multiple server profiles
- Easy editing (TUI or file)

## Comparison: Before vs After

### Before (Shell Scripts)
```bash
# Need to edit shell script variables
# Copy/paste into SSH session
# Watch text scroll by
# No idea how much time/cost remains
# One server = one script
# Error = start over
```

### After (anime CLI)
```bash
# Interactive TUI configuration
# One command deployment
# Beautiful progress UI
# Live time and cost tracking
# Multiple servers in one config
# Errors shown gracefully
```

## Next Steps

1. ✅ Try `anime config` right now (no server needed!)
2. ✅ Play with the TUI interface
3. ✅ Watch costs update as you toggle modules
4. ✅ Add multiple test servers
5. ⏳ When you get a real Lambda server, deploy for real!

## Help

```bash
anime --help              # Main help
anime config --help       # Config help
anime deploy --help       # Deploy help
anime status --help       # Status help
anime list --help         # List help
```

## Documentation

- **START_HERE.md** - Overview and quick start
- **QUICKSTART.md** - Detailed walkthrough
- **README.md** - Full documentation
- **ARCHITECTURE.md** - Technical details
- **SUMMARY.md** - Project summary

---

**Go ahead, try it!**

```bash
anime config
```

No shell scripts. Just a beautiful TUI. 🎌
