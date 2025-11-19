# anime CLI - Architecture

## Overview

`anime` is a Go-based CLI tool for managing Lambda Labs GH200 GPU instances. It replaces fragile shell scripts with a robust, type-safe, interactive TUI experience.

## Design Principles

1. **No Shell Scripts** - All logic in Go, scripts embedded as strings
2. **Interactive** - Beautiful TUI using Bubble Tea framework
3. **Cost Aware** - Show estimated costs before deployment
4. **Modular** - Install only what you need
5. **Real-time Feedback** - Live progress and output streaming
6. **Safe** - Graceful error handling, cancellable operations

## Architecture Diagram

```
┌─────────────────────────────────────────────┐
│              anime CLI                       │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │         Commands (Cobra)               │ │
│  │  • config  • deploy  • status  • list  │ │
│  └────────────────────────────────────────┘ │
│                    │                         │
│       ┌────────────┼────────────┐           │
│       │            │            │           │
│  ┌────▼────┐  ┌───▼────┐  ┌───▼─────┐     │
│  │   TUI   │  │ Config │  │   SSH   │     │
│  │(Bubble) │  │ (YAML) │  │ Client  │     │
│  └─────────┘  └────────┘  └─────────┘     │
│                                │            │
│                         ┌──────▼────────┐  │
│                         │   Installer   │  │
│                         │  • Scripts    │  │
│                         │  • Progress   │  │
│                         └───────────────┘  │
└────────────────────────────────────────────┘
                           │
                           │ SSH
                           ▼
              ┌────────────────────────┐
              │  Lambda GH200 Server   │
              │  • Ubuntu 22.04        │
              │  • NVIDIA GH200 GPU    │
              └────────────────────────┘
```

## Directory Structure

```
anime/
├── main.go                     # Entry point
├── cmd/                        # Commands
│   ├── root.go                # Root command + CLI setup
│   ├── config.go              # Interactive config TUI
│   ├── deploy.go              # Deploy with progress
│   ├── install.go             # Install anime + list servers
│   └── status.go              # Check server status
├── internal/
│   ├── config/                # Configuration
│   │   └── config.go          # Server/module config, YAML I/O
│   ├── installer/             # Installation engine
│   │   ├── installer.go       # SSH deployment, progress tracking
│   │   └── scripts.go         # Embedded bash scripts
│   ├── ssh/                   # SSH client
│   │   └── client.go          # SSH operations, file upload
│   └── tui/                   # Terminal UI
│       ├── config.go          # Multi-screen config TUI
│       └── install.go         # Real-time install progress
├── go.mod                     # Dependencies
├── Makefile                   # Build automation
├── README.md                  # Full documentation
├── QUICKSTART.md              # Quick start guide
└── ARCHITECTURE.md            # This file
```

## Component Details

### 1. Commands (`cmd/`)

Built with [Cobra](https://github.com/spf13/cobra):

- **`anime config`** - Interactive TUI for configuration
- **`anime deploy SERVER`** - Deploy modules to server
- **`anime status SERVER`** - Check server status
- **`anime list`** - List all configured servers

Each command is self-contained in its own file.

### 2. Configuration (`internal/config/`)

**File:** `~/.config/anime/config.yaml`

**Structure:**
```go
type Config struct {
    Servers []Server  // List of Lambda servers
    APIKeys APIKeys   // API keys for various services
}

type Server struct {
    Name        string   // e.g., "lambda-gh200-1"
    Host        string   // IP or hostname
    User        string   // SSH user (usually "ubuntu")
    SSHKey      string   // Path to private key
    CostPerHour float64  // For cost estimation
    Modules     []string // Selected module IDs
}
```

**Modules:**
- Defined in `AvailableModules` slice
- Each has: ID, Name, Description, TimeMinutes, Dependencies
- Auto-resolves dependencies during deployment

### 3. TUI (`internal/tui/`)

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss):

**Config TUI (`config.go`):**
- Multi-screen interface (menu, server list, server edit, module select, API keys)
- Keyboard navigation (vim keys + arrows)
- Real-time cost estimation
- Form inputs with validation

**Install TUI (`install.go`):**
- Real-time progress updates
- Live output streaming
- Cost tracking (elapsed time × cost/hr)
- System info display
- Graceful error handling

### 4. SSH Client (`internal/ssh/`)

Built with [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh):

**Capabilities:**
- Key-based authentication
- Command execution
- Output streaming (for progress)
- File upload (SCP-like)
- Connection management

**Methods:**
```go
client.RunCommand(cmd string) (string, error)
client.RunCommandWithProgress(cmd, progressChan)
client.UploadString(content, remotePath)
client.MakeExecutable(path)
client.FileExists(path) bool
```

### 5. Installer (`internal/installer/`)

**Installation Flow:**

1. **Resolve Dependencies**
   - Takes selected modules
   - Recursively adds dependencies
   - Returns ordered list (dependencies first)

2. **For Each Module:**
   - Get embedded bash script
   - Upload to `/tmp/anime-install-{module}.sh`
   - Make executable
   - Run with progress streaming
   - Clean up script

3. **Progress Updates:**
   ```go
   type ProgressUpdate struct {
       Module  string  // Current module ID
       Status  string  // "Starting", "Installing", "Complete", "Failed"
       Output  string  // Command output
       Error   error   // If failed
       Done    bool    // True when complete
   }
   ```

**Embedded Scripts:**
- All bash scripts stored in `Scripts` map
- No external files needed
- Scripts are embedded at compile time
- Delivered to server via SSH

## Data Flow

### Configuration Flow

```
User Input (TUI)
    │
    ▼
ConfigModel (Bubble Tea)
    │
    ▼
Config struct
    │
    ▼
YAML serialization
    │
    ▼
~/.config/anime/config.yaml
```

### Deployment Flow

```
User runs: anime deploy SERVER
    │
    ▼
Load config from YAML
    │
    ▼
Resolve module dependencies
    │
    ▼
Connect to server via SSH
    │
    ▼
Get system info
    │
    ▼
For each module:
    Upload script → Execute → Stream output
    │
    ▼
Update TUI with progress
    │
    ▼
Installation complete
```

## Key Technologies

| Component | Technology | Purpose |
|-----------|-----------|---------|
| CLI Framework | [Cobra](https://github.com/spf13/cobra) | Command structure |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Interactive UI |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| Forms | [Bubbles](https://github.com/charmbracelet/bubbles) | Text inputs |
| SSH | [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) | Remote execution |
| Config | [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) | YAML parsing |

## Error Handling

### Strategy

1. **Graceful Degradation**
   - Connection failures → show error, don't crash
   - Script failures → mark module failed, continue or abort
   - Config errors → create default, warn user

2. **User Feedback**
   - All errors shown in TUI
   - Color coding (red for errors)
   - Helpful error messages

3. **Cleanup**
   - Always close SSH connections
   - Remove temporary scripts from server
   - Save partial progress in config

### Example Error Flow

```go
// In installer.go
if err := i.installModule(modID); err != nil {
    i.sendProgress(modID, "Failed", "", err, true)
    return fmt.Errorf("failed to install %s: %w", modID, err)
}
```

```go
// In TUI
case progressMsg:
    if msg.Error != nil {
        m.err = msg.Error  // Display in red
    }
```

## Cost Estimation

**Formula:**
```
cost = (total_minutes / 60) * cost_per_hour
```

**Features:**
- Pre-deployment estimation (in config TUI)
- Real-time tracking (during deployment)
- Module + dependency time calculation
- Displayed before user commits

**Example:**
```
Core (5m) + PyTorch (2m) + Ollama (1m) = 8 minutes
8/60 * $20/hr = $2.67
```

## Security Considerations

1. **SSH Keys**
   - Private keys never transmitted
   - Proper file permissions checked (600)
   - Keys stored locally only

2. **API Keys**
   - Stored in `~/.config/anime/config.yaml` (0600 perms)
   - Masked in TUI (shown as `•••`)
   - Never logged or displayed

3. **Remote Execution**
   - Scripts uploaded to `/tmp`, deleted after
   - No persistent modifications without user consent
   - All commands visible in TUI output

4. **Host Key Verification**
   - Currently: `InsecureIgnoreHostKey` (TODO)
   - Future: Proper known_hosts management

## Performance

### Speed Optimizations

1. **Parallel Downloads**
   - Ollama models pulled concurrently (`&` in bash)
   - Multiple small operations batched

2. **Dependency Resolution**
   - Single pass algorithm
   - Modules installed in optimal order

3. **Progress Streaming**
   - Buffered channels (100 capacity)
   - Throttled updates (500ms intervals)

### Resource Usage

- **Memory:** ~20MB (TUI + SSH connection)
- **Network:** Minimal (only SSH traffic)
- **Disk:** <10MB binary, <1KB config

## Testing Strategy

### Manual Testing

```bash
# Test TUI navigation
anime config
# (navigate through all screens)

# Test deployment (dry run concept)
# Future: anime deploy --dry-run SERVER

# Test status
anime status SERVER
```

### Future: Automated Testing

```go
// Test config loading
func TestLoadConfig(t *testing.T) { ... }

// Test dependency resolution
func TestResolveDependencies(t *testing.T) { ... }

// Test SSH connection
func TestSSHConnection(t *testing.T) { ... }
```

## Extension Points

### Adding New Modules

1. Add to `AvailableModules` in `config.go`:
```go
{
    ID:          "new-module",
    Name:        "New Module",
    Description: "What it does",
    TimeMinutes: 5,
    Dependencies: []string{"core"},
    Script:      "new-module",
}
```

2. Add script to `Scripts` in `scripts.go`:
```go
"new-module": `#!/bin/bash
set -e
echo "Installing new module..."
# installation commands
`,
```

### Adding New Commands

1. Create `cmd/newcmd.go`:
```go
var newCmd = &cobra.Command{
    Use:   "new",
    Short: "New command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // implementation
    },
}
```

2. Register in `cmd/root.go`:
```go
func init() {
    rootCmd.AddCommand(newCmd)
}
```

## Comparison: Shell vs Go

| Aspect | Shell Scripts | anime CLI |
|--------|--------------|-----------|
| Type Safety | ❌ None | ✅ Full Go typing |
| Error Handling | ❌ Basic | ✅ Comprehensive |
| User Interface | ❌ Text output | ✅ Interactive TUI |
| Progress Tracking | ❌ Echo statements | ✅ Real-time UI |
| Configuration | ❌ Hardcoded | ✅ YAML + TUI |
| Cost Estimation | ❌ Manual | ✅ Automatic |
| Multiple Servers | ❌ Multiple scripts | ✅ Single config |
| Maintainability | ❌ Low | ✅ High |
| Testing | ❌ Difficult | ✅ Standard Go tests |

## Future Enhancements

1. **Better Host Key Verification**
   - Store/verify SSH host keys
   - Warn on MITM attacks

2. **Parallel Deployments**
   - Deploy to multiple servers simultaneously
   - Progress dashboard

3. **Rollback Support**
   - Track installation state
   - Rollback on failure

4. **Template Support**
   - Save module combinations as templates
   - Quick deploy from templates

5. **Web Dashboard**
   - Optional web UI (Bubble Tea + SSH)
   - Remote management

6. **CI/CD Integration**
   - Non-interactive mode
   - JSON output for scripting

7. **Cloud Provider Integration**
   - Spawn Lambda instances via API
   - Auto-configure SSH access

## Conclusion

`anime` replaces fragile shell scripts with a robust, type-safe, interactive CLI. The architecture prioritizes:
- **User Experience** - Beautiful TUI, real-time feedback
- **Reliability** - Proper error handling, graceful failures
- **Maintainability** - Clean Go code, modular design
- **Extensibility** - Easy to add modules and commands

No more shell scripts! 🎉
