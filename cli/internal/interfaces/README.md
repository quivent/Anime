# Anime CLI Interface Abstraction Layer

This package defines core interface abstractions for the anime CLI components, enabling loose coupling, dependency injection, and improved testability.

## Overview

The interfaces package provides clean abstractions for:

- **SSHClient**: Remote command execution and file operations
- **Installer**: Module installation with dependency resolution
- **SourceController**: Source code synchronization operations
- **PackageManager**: Package registry and distribution management

## Interfaces

### SSHClient

Handles secure SSH operations including command execution and file transfers.

```go
type SSHClient interface {
    RunCommand(cmd string) (string, error)
    RunCommandWithProgress(cmd string, progress chan<- string) error
    UploadString(content, path string) error
    MakeExecutable(path string) error
    Close() error
}
```

**Implementation**: `internal/ssh.Client`

**Example**:
```go
import (
    "github.com/joshkornreich/anime/internal/interfaces"
    "github.com/joshkornreich/anime/internal/ssh"
)

func processWithSSH(client interfaces.SSHClient) error {
    output, err := client.RunCommand("uname -a")
    if err != nil {
        return err
    }
    fmt.Println(output)
    return nil
}

// Usage
client, err := ssh.NewClient("alice", "user", "~/.ssh/id_rsa")
if err != nil {
    return err
}
defer client.Close()

processWithSSH(client)
```

### Installer

Manages module installation with dependency resolution and parallel execution.

```go
type Installer interface {
    Install(modules []string) error
    GetProgressChannel() <-chan ProgressUpdate
    SetParallel(parallel bool)
    SetJobs(jobs int)
    TestConnection() error
}
```

**Implementation**: `internal/installer.Installer`

**Example**:
```go
import (
    "github.com/joshkornreich/anime/internal/interfaces"
    "github.com/joshkornreich/anime/internal/installer"
    "github.com/joshkornreich/anime/internal/ssh"
)

func installModules(inst interfaces.Installer, modules []string) error {
    // Configure parallel installation
    inst.SetParallel(true)
    inst.SetJobs(4)

    // Monitor progress
    go func() {
        for update := range inst.GetProgressChannel() {
            fmt.Printf("[%s] %s: %s\n", update.Module, update.Status, update.Output)
            if update.Error != nil {
                fmt.Printf("Error: %v\n", update.Error)
            }
        }
    }()

    return inst.Install(modules)
}

// Usage
client, _ := ssh.NewClient("alice", "user", "~/.ssh/id_rsa")
inst := installer.New(client)
installModules(inst, []string{"pytorch", "comfyui"})
```

### SourceController

Provides rsync-based source control operations for code synchronization.

```go
type SourceController interface {
    Push() error
    Pull() error
    Status() error
    Link(remotePath string) error
    Unlink() error
}
```

**Implementation**: `internal/source.Controller`

**Example**:
```go
import (
    "github.com/joshkornreich/anime/internal/interfaces"
    "github.com/joshkornreich/anime/internal/source"
)

func syncCode(sc interfaces.SourceController) error {
    // Link to remote path
    if err := sc.Link("myproject"); err != nil {
        return err
    }

    // Check status
    if err := sc.Status(); err != nil {
        return err
    }

    // Push changes
    return sc.Push()
}

// Usage
config := &source.Config{
    Server:  "alice",
    KeyPath: "~/.ssh/id_rsa",
}
controller := source.NewController("alice", config)
syncCode(controller)
```

### PackageManager

Manages package installation, publication, and distribution.

```go
type PackageManager interface {
    Install(name string, global bool, force bool) error
    Uninstall(name string, global bool) error
    Publish() error
    Search(query string) error
    List(global bool) error
}
```

**Implementation**: `internal/pkg.Manager`

**Example**:
```go
import (
    "github.com/joshkornreich/anime/internal/interfaces"
    "github.com/joshkornreich/anime/internal/pkg"
)

func managePackages(pm interfaces.PackageManager) error {
    // Install a package
    if err := pm.Install("mypackage@1.0.0", false, false); err != nil {
        return err
    }

    // List installed packages
    if err := pm.List(false); err != nil {
        return err
    }

    return nil
}

// Usage
config := &pkg.Config{
    Server:  "alice",
    KeyPath: "~/.ssh/id_rsa",
}
manager := pkg.NewManager("alice", config)
managePackages(manager)
```

## Benefits

### Testability

Interfaces enable easy mocking for unit tests:

```go
type MockSSHClient struct {
    CommandOutput string
    CommandError  error
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
    return m.CommandOutput, m.CommandError
}

// ... implement other methods

func TestMyFunction(t *testing.T) {
    mock := &MockSSHClient{
        CommandOutput: "test output",
    }

    result := processWithSSH(mock)
    // assertions...
}
```

### Dependency Injection

Components can accept interfaces rather than concrete types:

```go
type DeploymentService struct {
    ssh       interfaces.SSHClient
    installer interfaces.Installer
}

func NewDeploymentService(ssh interfaces.SSHClient, inst interfaces.Installer) *DeploymentService {
    return &DeploymentService{
        ssh:       ssh,
        installer: inst,
    }
}
```

### Loose Coupling

Implementations can be swapped without affecting consumers:

```go
// Production
sshClient := ssh.NewClient("alice", "user", "~/.ssh/id_rsa")

// Testing
sshClient := &MockSSHClient{...}

// Same function works with both
processWithSSH(sshClient)
```

## Implementation Notes

### Interface Satisfaction Checks

All concrete implementations include compile-time checks to ensure interface satisfaction:

```go
// In internal/ssh/client.go
var _ interfaces.SSHClient = (*Client)(nil)

// In internal/installer/installer.go
var _ interfaces.Installer = (*Installer)(nil)

// In internal/source/controller.go
var _ interfaces.SourceController = (*Controller)(nil)

// In internal/pkg/manager.go
var _ interfaces.PackageManager = (*Manager)(nil)
```

These checks fail at compile-time if the implementation doesn't satisfy the interface, catching errors early.

### Type Aliases

Some packages use type aliases to maintain backward compatibility while satisfying interfaces:

```go
// In internal/installer/installer.go
type ProgressUpdate = interfaces.ProgressUpdate
```

This allows existing code using `installer.ProgressUpdate` to continue working while ensuring the installer satisfies the `interfaces.Installer` interface.

## Design Principles

1. **Single Responsibility**: Each interface has a focused purpose
2. **Interface Segregation**: Interfaces are minimal and specific
3. **Dependency Inversion**: Depend on abstractions, not concretions
4. **Open/Closed**: Open for extension, closed for modification

## Future Enhancements

Potential areas for expansion:

- Additional interfaces for configuration management
- Monitoring and logging interfaces
- Event notification interfaces
- Plugin/extension interfaces
