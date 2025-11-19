# 🎌 anime - Lambda GH200 Management CLI

A beautiful Go CLI for managing Lambda Labs GH200 instances. No more shell scripts!

## Features

- 🎨 **Beautiful TUI** - Interactive terminal UI using Bubble Tea
- 🚀 **Easy Configuration** - Configure servers, modules, and API keys visually
- 💰 **Cost Estimation** - See estimated costs before deployment
- 📊 **Real-time Progress** - Watch installation progress live
- 🔌 **SSH Management** - Automatic SSH connection and script deployment
- 📦 **Modular Installation** - Install only what you need

## Installation

```bash
cd /Users/joshkornreich/lambda
go build -o anime
sudo mv anime /usr/local/bin/

# Or install directly
go install
```

## Quick Start

### 1. Configure Your Servers

```bash
anime config
```

This opens an interactive TUI where you can:
- Add/edit Lambda servers
- Select installation modules
- Configure API keys (Anthropic, OpenAI, HuggingFace, Lambda Labs)
- See cost estimates

### 2. Deploy to a Server

```bash
anime deploy lambda-gh200-1
```

Watch the installation progress in real-time with:
- Module-by-module status
- Live output streaming
- Real-time cost tracking
- Beautiful progress indicators

### 3. Check Server Status

```bash
anime status lambda-gh200-1
```

See:
- System information
- Installed components
- GPU status
- Available models

### 4. List All Servers

```bash
anime list
```

## Available Modules

| Module | Time | Description |
|--------|------|-------------|
| **Core System** | 5 min | CUDA 12.4, Python, Node.js, Docker |
| **PyTorch** | 2 min | PyTorch, Transformers, Diffusers, xformers |
| **Ollama** | 1 min | Ollama LLM server (no models) |
| **Small Models** | 8 min | Mistral, Llama 3.3 8B, Qwen 2.5 7B |
| **Medium Models** | 25 min | Qwen 2.5 14B, Mixtral, DeepSeek Coder |
| **Large Models** | 40 min | Llama 3.3 70B, Qwen 2.5 72B |
| **ComfyUI** | 2 min | Stable Diffusion UI with Manager |
| **Claude Code** | 1 min | Anthropic Claude Code CLI |

## Configuration

Config is stored in `~/.config/anime/config.yaml`:

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

api_keys:
  anthropic: sk-ant-...
  openai: sk-...
  huggingface: hf_...
  lambda_labs: lambda_...
```

## Usage Examples

### Basic Setup (Minimal Cost)

```bash
# 1. Configure server
anime config
  # Add server with basic info
  # Select: Core + PyTorch (~$3 total)

# 2. Deploy
anime deploy my-server

# 3. Check status
anime status my-server
```

### Production Setup

```bash
# 1. Configure with all modules
anime config
  # Select: Core + PyTorch + Ollama + Small Models (~$8 total)

# 2. Deploy and watch progress
anime deploy production-server
```

### Multiple Servers

```bash
# Configure different servers for different purposes
anime config
  # Server 1: "dev" - Core + PyTorch
  # Server 2: "llm" - Core + Ollama + Large Models
  # Server 3: "imaging" - Core + PyTorch + ComfyUI

# Deploy to specific server
anime deploy llm
anime deploy imaging
```

## TUI Navigation

### Main Menu
- `↑/↓` or `j/k` - Navigate
- `Enter` - Select
- `q` - Quit

### Server List
- `↑/↓` - Navigate
- `Enter` - Configure modules
- `d` - Delete server
- `Esc` - Back to menu

### Module Selection
- `↑/↓` - Navigate
- `Space` - Toggle module
- `Enter` - Save selection
- `Esc` - Cancel

### Form Inputs
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `Enter` - Save
- `Esc` - Cancel

## Cost Estimation

The CLI automatically calculates estimated costs based on:
- Selected modules and their installation time
- Module dependencies (auto-included)
- Server's cost per hour

Example output:
```
Estimated cost: $8.50 @ $20/hr
  Core System: 5 min
  PyTorch: 2 min
  Ollama: 1 min
  Small Models: 8 min
  ---
  Total: ~17 minutes
```

## Architecture

```
anime/
├── cmd/                  # CLI commands
│   ├── root.go          # Root command
│   ├── config.go        # Config TUI command
│   ├── deploy.go        # Deploy command
│   ├── install.go       # Install/list commands
│   └── status.go        # Status command
├── internal/
│   ├── config/          # Configuration management
│   │   └── config.go    # Config struct and modules
│   ├── installer/       # Installation logic
│   │   ├── installer.go # SSH deployment
│   │   └── scripts.go   # Embedded bash scripts
│   ├── ssh/             # SSH client
│   │   └── client.go    # SSH operations
│   └── tui/             # Terminal UI
│       ├── config.go    # Config TUI
│       └── install.go   # Install progress TUI
└── main.go              # Entry point
```

## Development

```bash
# Run without installing
go run main.go config

# Build
go build -o anime

# Install
go install

# Run tests
go test ./...

# Format
go fmt ./...
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/crypto/ssh` - SSH client
- `gopkg.in/yaml.v3` - YAML config

## Troubleshooting

### Connection Issues

```bash
# Test SSH manually
ssh -i ~/.ssh/lambda_key.pem ubuntu@192.168.1.100

# Check config
cat ~/.config/anime/config.yaml
```

### Installation Failures

```bash
# Check status to see what's installed
anime status my-server

# Check server logs
ssh ubuntu@YOUR_IP
journalctl -u ollama -f
cat /tmp/anime-install-*.sh
```

### Permission Issues

```bash
# Ensure SSH key has correct permissions
chmod 600 ~/.ssh/lambda_key.pem

# Ensure user has sudo access
ssh ubuntu@YOUR_IP "sudo -v"
```

## Comparison with Shell Scripts

| Feature | anime CLI | Shell Scripts |
|---------|-----------|---------------|
| Configuration | Interactive TUI | Manual editing |
| Cost Estimation | Built-in | Manual calculation |
| Progress Tracking | Real-time UI | Text output |
| Error Handling | Graceful | Exit on error |
| Module Selection | Visual checkboxes | Comment/uncomment |
| Multiple Servers | Easy switching | Multiple files |
| API Key Management | Encrypted storage | Plain text files |

## License

MIT

## Contributing

PRs welcome! Please ensure:
- Code is formatted (`go fmt`)
- Tests pass (`go test ./...`)
- TUI flows work correctly
