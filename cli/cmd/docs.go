package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var docsSection string

var docsCmd = &cobra.Command{
	Use:   "docs [section]",
	Short: "Display comprehensive CLI documentation",
	Long: `Display comprehensive documentation for the Anime CLI.

Sections:
  overview     - Introduction and core concepts
  installer    - Software installation system
  source       - Source control commands
  packages     - Package management
  server       - Server management
  llm          - LLM and AI commands
  config       - Configuration options
  all          - Complete documentation

Examples:
  anime docs              # Show overview
  anime docs installer    # Installer documentation
  anime docs source       # Source control documentation
  anime docs all          # Complete documentation`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func runDocs(cmd *cobra.Command, args []string) {
	section := "overview"
	if len(args) == 1 {
		section = strings.ToLower(args[0])
	}

	var content string
	switch section {
	case "overview":
		content = docsOverview()
	case "installer", "install":
		content = docsInstaller()
	case "source", "src":
		content = docsSource()
	case "packages", "pkg":
		content = docsPackages()
	case "server", "servers":
		content = docsServer()
	case "llm", "ai":
		content = docsLLM()
	case "config", "configuration":
		content = docsConfig()
	case "all", "full":
		content = docsAll()
	default:
		fmt.Printf("\n%s Unknown section: %s\n\n", theme.ErrorStyle.Render("Error:"), section)
		fmt.Println("Available sections: overview, installer, source, packages, server, llm, config, all")
		return
	}

	// Try to render with glow if available
	if renderWithGlow(content) {
		return
	}

	// Fallback to plain output
	fmt.Println(content)
}

func renderWithGlow(content string) bool {
	// Check if glow is available
	_, err := exec.LookPath("glow")
	if err != nil {
		return false
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "anime-docs-*.md")
	if err != nil {
		return false
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		return false
	}
	tmpFile.Close()

	// Run glow
	glowCmd := exec.Command("glow", "-p", tmpFile.Name())
	glowCmd.Stdout = os.Stdout
	glowCmd.Stderr = os.Stderr
	glowCmd.Stdin = os.Stdin

	if err := glowCmd.Run(); err != nil {
		return false
	}

	return true
}

func docsOverview() string {
	return `# Anime CLI

> Advanced CLI for Lambda GH200 deployment, source control, and package management.

## Core Systems

Anime CLI provides three integrated systems:

| System | Command | Purpose |
|--------|---------|---------|
| **Installer** | ` + "`anime install`" + ` | Deploy software modules to GH200 instances |
| **Source** | ` + "`anime source`" + ` | Rsync-based code synchronization |
| **Packages** | ` + "`anime pkg`" + ` | Publish and manage reusable packages |

## Quick Start

` + "```bash" + `
# View this help
anime docs

# Browse available software
anime packages

# Install software
anime install python pytorch claude

# Push code to remote
anime source push

# Publish a package
anime pkg publish
` + "```" + `

## Getting Help

` + "```bash" + `
anime docs [section]     # Detailed documentation
anime examples [command] # Quick usage examples
anime reference          # Interactive explorer
anime tree               # Full command tree
anime <cmd> --help       # Command-specific help
` + "```" + `

## Documentation Sections

- ` + "`anime docs installer`" + ` - Software installation
- ` + "`anime docs source`" + ` - Source control
- ` + "`anime docs packages`" + ` - Package management
- ` + "`anime docs server`" + ` - Server management
- ` + "`anime docs llm`" + ` - AI/LLM commands
- ` + "`anime docs config`" + ` - Configuration
- ` + "`anime docs all`" + ` - Complete documentation
`
}

func docsInstaller() string {
	return `# Installer System

> Deploy software modules to Lambda GH200 instances with automatic dependency resolution.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime packages`" + ` | Browse available packages |
| ` + "`anime interactive`" + ` | Interactive package selector |
| ` + "`anime install <pkg>`" + ` | Install packages |

## Package Categories

### Foundation
` + "```bash" + `
anime install core       # gcc, git, curl, wget, python3, pkg-config
anime install python     # Python 3.11+, pip, venv, numpy, pandas
anime install nodejs     # Node.js 20 LTS
anime install go         # Go programming language
` + "```" + `

### GPU & ML
` + "```bash" + `
anime install nvidia     # NVIDIA drivers & CUDA 12.4
anime install pytorch    # PyTorch with CUDA support
` + "```" + `

### LLM Runtime
` + "```bash" + `
anime install ollama     # Ollama for local LLMs
anime install vllm       # High-throughput LLM serving
anime install sglang     # Fast LLM inference
` + "```" + `

### LLM Models
` + "```bash" + `
anime install llama3              # Meta's Llama 3
anime install qwen2               # Alibaba's Qwen 2
anime install deepseek-coder      # DeepSeek Coder
anime install command-r-7b        # Cohere Command-R
` + "```" + `

### Applications
` + "```bash" + `
anime install claude     # Claude Code CLI (Anthropic)
anime install comfyui    # ComfyUI with manager
anime install docker     # Docker containers
` + "```" + `

### Image Generation
` + "```bash" + `
anime install flux-dev       # Flux.1 Dev
anime install flux-schnell   # Flux.1 Schnell (fast)
anime install sdxl           # Stable Diffusion XL
anime install sd15           # Stable Diffusion 1.5
` + "```" + `

### Video Generation
` + "```bash" + `
anime install wan2       # Wan2.1 video generation
anime install hunyuan    # Hunyuan video model
` + "```" + `

## Options

` + "```bash" + `
anime install -y <pkg>           # Skip confirmation
anime install --dry-run <pkg>    # Preview installation
anime install -r -s alice <pkg>  # Install on remote server
anime install --phased <pkg>     # Install with phase confirmations
` + "```" + `

## Dependency Resolution

Packages automatically install their dependencies:

` + "```bash" + `
# This will also install: core → python → pytorch → nvidia → comfyui
anime install flux-dev
` + "```" + `
`
}

func docsSource() string {
	return `# Source Control System

> Rsync-based code synchronization with remote servers.

## Overview

Source control syncs code to ` + "`alice:~/cpm/anime`" + ` by default.
Uses rsync over SSH with embedded keys for secure transfer.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime source push`" + ` | Push local changes to remote |
| ` + "`anime source pull`" + ` | Pull remote changes to local |
| ` + "`anime source clone`" + ` | Clone remote repo into new folder |
| ` + "`anime source status`" + ` | Show sync status / diff |
| ` + "`anime source sync`" + ` | Bidirectional sync |
| ` + "`anime source link`" + ` | Link directory to remote path |
| ` + "`anime source init`" + ` | Initialize and push new repo |
| ` + "`anime source list`" + ` | List remote repositories |
| ` + "`anime source tree`" + ` | Tree view of remote repos |
| ` + "`anime source history`" + ` | Show push/pull history |
| ` + "`anime source rename`" + ` | Rename/move remote repo |
| ` + "`anime source delete`" + ` | Delete remote repo |

## Workflow Examples

### New Project
` + "```bash" + `
# Initialize and push
cd myproject
anime source init myproject

# Future pushes
anime source push
` + "```" + `

### Existing Project
` + "```bash" + `
# Clone from remote
anime source clone org/project
cd project

# Make changes and push
anime source push
` + "```" + `

### Team Sync
` + "```bash" + `
# Check what's different
anime source status

# Bidirectional sync (newer files win)
anime source sync

# Pull only
anime source pull
` + "```" + `

## Link System

Link your directory to skip specifying path each time:

` + "```bash" + `
# Link once
anime source link org/myproject

# Now these work without arguments
anime source push
anime source pull
anime source status
anime source sync
` + "```" + `

## Flags

| Flag | Description |
|------|-------------|
| ` + "`-s, --server`" + ` | Override default server (alice) |
| ` + "`-n, --dry-run`" + ` | Preview without making changes |
| ` + "`-f, --force`" + ` | Force overwrite (clone only) |

## Excluded Files

These are automatically excluded from sync:
- ` + "`.git`" + `, ` + "`node_modules`" + `, ` + "`cpm_modules`" + `
- ` + "`__pycache__`" + `, ` + "`*.pyc`" + `
- ` + "`.env`" + `, ` + "`venv`" + `, ` + "`.venv`" + `
- ` + "`.cpm-link`" + `, ` + "`.cpm-installed.json`" + `
`
}

func docsPackages() string {
	return `# Package Management System

> Publish and manage reusable packages with versioning.

## Overview

Packages are stored at ` + "`alice:~/cpm/packages`" + ` with versioning support.
Each package has a ` + "`cpm.json`" + ` manifest file.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime pkg init`" + ` | Create new cpm.json |
| ` + "`anime pkg publish`" + ` | Publish to registry |
| ` + "`anime pkg republish`" + ` | Update existing version |
| ` + "`anime pkg install`" + ` | Install a package |
| ` + "`anime pkg uninstall`" + ` | Remove installed package |
| ` + "`anime pkg search`" + ` | Search registry |
| ` + "`anime pkg info`" + ` | Show package details |
| ` + "`anime pkg versions`" + ` | List available versions |
| ` + "`anime pkg update`" + ` | Update installed packages |
| ` + "`anime pkg list`" + ` | List installed packages |

## Package Manifest (cpm.json)

` + "```json" + `
{
  "name": "mypackage",
  "version": "1.0.0",
  "description": "My awesome package",
  "author": "Your Name",
  "keywords": ["utility", "tools"],
  "license": "MIT",
  "repository": "https://github.com/you/mypackage"
}
` + "```" + `

## Publishing Workflow

` + "```bash" + `
# Create package manifest
anime pkg init mypackage

# Edit cpm.json with your details
# ... make your changes ...

# Publish version 1.0.0
anime pkg publish

# Update version in cpm.json to 1.0.1
# Publish new version
anime pkg publish

# Or update existing version in place
anime pkg republish
` + "```" + `

## Installing Packages

` + "```bash" + `
# Install latest version
anime pkg install mypackage

# Install specific version
anime pkg install mypackage@1.0.0

# Install globally
anime pkg install -g mypackage

# Force reinstall
anime pkg install -f mypackage
` + "```" + `

## Installation Paths

| Scope | Path |
|-------|------|
| Local | ` + "`./cpm_modules/<package>`" + ` |
| Global | ` + "`~/.cpm/packages/<package>`" + ` |

## Flags

| Flag | Description |
|------|-------------|
| ` + "`-s, --server`" + ` | Override default server |
| ` + "`-n, --dry-run`" + ` | Preview without changes |
| ` + "`-g, --global`" + ` | Use global packages |
| ` + "`-f, --force`" + ` | Force overwrite |
`
}

func docsServer() string {
	return `# Server Management

> Configure and manage Lambda GH200 instances.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime add <name> <ip>`" + ` | Add a server |
| ` + "`anime set <name> [ip]`" + ` | Set/update server (auto-detect IP) |
| ` + "`anime status`" + ` | Show server status |
| ` + "`anime remove <name>`" + ` | Remove a server |
| ` + "`anime deploy`" + ` | Deploy to server |
| ` + "`anime ssh <name>`" + ` | SSH into server |

## Adding Servers

` + "```bash" + `
# Add with explicit IP
anime add alice 192.168.1.100

# Set with auto-detection
anime set alice

# Update existing server
anime set alice 192.168.1.101
` + "```" + `

## Server Status

` + "```bash" + `
# Show all servers
anime status

# Check specific server
anime status alice
` + "```" + `

## SSH Access

` + "```bash" + `
# SSH into a server
anime ssh alice

# Run command on server
anime ssh alice "nvidia-smi"
` + "```" + `

## Deployment

` + "```bash" + `
# Deploy to configured server
anime deploy

# Deploy with specific config
anime deploy --config myconfig.yaml
` + "```" + `
`
}

func docsLLM() string {
	return `# LLM & AI Commands

> Query language models and use AI-powered features.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime query <model> \"prompt\"`" + ` | Query Ollama models |
| ` + "`anime prompt \"natural language\"`" + ` | AI-interpreted commands |
| ` + "`anime models`" + ` | List downloaded models |

## Querying Models

` + "```bash" + `
# Query a specific model
anime query llama3 "Explain quantum computing"

# Query with context
anime query deepseek-coder "Write a Python function to sort a list"

# List available models
anime models
` + "```" + `

## Natural Language Commands

` + "```bash" + `
# Let AI interpret your intent
anime prompt "install pytorch and cuda"
anime prompt "show me the server status"
anime prompt "list all downloaded models"
` + "```" + `

## Installing LLM Models

` + "```bash" + `
# Install Ollama runtime
anime install ollama

# Install specific models
anime install llama3
anime install qwen2
anime install deepseek-coder
anime install command-r-7b
` + "```" + `

## Model Categories

### General Purpose
- ` + "`llama3`" + ` - Meta's Llama 3 (8B/70B)
- ` + "`qwen2`" + ` - Alibaba's Qwen 2

### Code
- ` + "`deepseek-coder`" + ` - Code-specialized model
- ` + "`codellama`" + ` - Meta's Code Llama

### Efficient
- ` + "`command-r-7b`" + ` - Cohere's efficient model
- ` + "`phi3`" + ` - Microsoft's Phi-3
`
}

func docsConfig() string {
	return `# Configuration

> Configure Anime CLI settings and preferences.

## Commands

| Command | Description |
|---------|-------------|
| ` + "`anime config`" + ` | Show current configuration |
| ` + "`anime config set <key> <value>`" + ` | Set configuration value |
| ` + "`anime config get <key>`" + ` | Get configuration value |

## Configuration File

Located at ` + "`~/.anime/config.yaml`" + `

` + "```yaml" + `
# Server configuration
servers:
  alice:
    host: 192.168.1.100
    user: ubuntu
  bob:
    host: 192.168.1.101
    user: ubuntu

# Default settings
defaults:
  server: alice

# Module configuration
modules:
  - core
  - python
  - pytorch
` + "```" + `

## Environment Variables

| Variable | Description |
|----------|-------------|
| ` + "`ANIME_SERVER`" + ` | Default server name |
| ` + "`ANIME_CONFIG`" + ` | Config file path |

## SSH Keys

Anime CLI uses embedded SSH keys for authentication.
Keys are stored securely and used automatically.

## Diagnostics

` + "```bash" + `
# Run diagnostics
anime doctor

# Check installation
anime doctor --check

# View logs
anime logs
` + "```" + `
`
}

func docsAll() string {
	return docsOverview() + "\n---\n\n" +
		docsInstaller() + "\n---\n\n" +
		docsSource() + "\n---\n\n" +
		docsPackages() + "\n---\n\n" +
		docsServer() + "\n---\n\n" +
		docsLLM() + "\n---\n\n" +
		docsConfig()
}
