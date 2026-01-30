# Structured Logging Usage Guide

This guide demonstrates how to use the structured logging infrastructure in the anime CLI.

## Quick Start

The logger is automatically initialized when the CLI starts. Users can control logging via flags:

```bash
# Enable debug logging
anime deploy my-server --debug

# Set specific log level
anime push lambda --log-level info

# Log to file
anime deploy my-server --log-file /tmp/anime.log --log-level debug

# Log to file with JSON format (better for log aggregation)
anime deploy my-server --log-file /var/log/anime/deploy.log --log-level info
```

## Basic Logging

### Simple Messages

```go
package mypackage

import "github.com/joshkornreich/anime/internal/logger"

func MyFunction() {
    logger.Debug("Starting operation")
    logger.Info("Processing started")
    logger.Warn("Configuration not optimized")
    logger.Error("Failed to connect", "error", err)
}
```

### With Structured Fields

```go
logger.Info("Module installation started",
    "module_id", "pytorch",
    "estimated_time_minutes", 8,
    "dependencies", []string{"core", "cuda"},
)
```

## Contextual Logging

### Server Context

```go
// Create a server-specific logger
serverLogger := logger.WithServer("lambda-1")
serverLogger.Info("Connecting to server")
serverLogger.Debug("SSH key loaded", "key_path", "/home/user/.ssh/id_rsa")
```

### Module Context

```go
// Create a module-specific logger
modLogger := logger.WithModule("pytorch")
modLogger.Info("Installing module",
    "name", "PyTorch + AI Libraries",
    "estimated_time_minutes", 2,
)
modLogger.Debug("Script loaded", "script_id", "pytorch")
```

### Operation Context

```go
// Create an operation-specific logger
opLogger := logger.WithOperation("deploy")
opLogger.Info("Deployment started", "target", "lambda-1")
opLogger.Error("Deployment failed", "error", err)
```

### Multiple Context Layers

```go
// Chain multiple context layers for rich logging
serverLogger := logger.WithServer("my-server")
moduleLogger := serverLogger.With("module", "pytorch")
opLogger := moduleLogger.With("operation", "install")

opLogger.Info("installation step completed",
    "step", "dependencies",
    "duration_ms", 1234,
)
```

### Custom Context

```go
// Create logger with multiple custom fields
fields := map[string]any{
    "server":    "lambda-1",
    "module":    "pytorch",
    "operation": "install",
    "user":      "ubuntu",
}

contextLogger := logger.WithContext(fields)
contextLogger.Info("Complex operation started")
```

## Pattern: Keep fmt.Print for User-Facing TUI, Use logger.* for Operational Info

### DON'T ❌

```go
// Don't replace TUI output with logs
logger.Info("✨ Push complete!")  // This should be fmt.Print
logger.Info("Target: lambda-1")   // This should be fmt.Print
```

### DO ✅

```go
// Keep TUI output with fmt.Print
fmt.Println(theme.SuccessStyle.Render("✨ Push complete!"))
fmt.Printf("Target: %s\n", theme.HighlightStyle.Render(target))

// Use logger for operational/debug info
logger.Info("Push operation completed",
    "target", target,
    "version", version,
    "duration_ms", duration,
    "bytes_transferred", bytes,
)
```

## File-Specific Examples

### internal/installer/installer.go

```go
package installer

import "github.com/joshkornreich/anime/internal/logger"

func (i *Installer) Install(modules []string) error {
    // Log installation start with all context
    logger.Info("Starting module installation",
        "module_count", len(modules),
        "modules", modules,
        "parallel", i.parallel,
        "jobs", i.jobs,
        "server", i.serverName,
    )

    // Keep fmt.Print for TUI progress display
    i.sendProgress("", "Starting installation",
        fmt.Sprintf("Installing %d modules", len(modules)), nil, false)

    // ... installation logic ...

    // Log success with metrics
    logger.Info("All modules installed successfully",
        "module_count", len(modules),
        "duration_ms", duration,
        "server", i.serverName,
    )

    return nil
}

func (i *Installer) installModule(modID string) error {
    // Create module-specific logger
    modLogger := logger.WithModule(modID)

    modLogger.Info("Installing module",
        "name", module.Name,
        "estimated_time_minutes", module.TimeMinutes,
    )

    modLogger.Debug("Uploading installation script",
        "remote_path", remotePath)

    // Keep TUI progress updates
    i.sendProgress(modID, "Starting",
        fmt.Sprintf("Installing %s", module.Name), nil, false)

    // Log errors with context
    if err := i.client.UploadString(script, remotePath); err != nil {
        modLogger.Error("Failed to upload script", "error", err)
        return fmt.Errorf("failed to upload script: %w", err)
    }

    modLogger.Info("Module installation complete")
    return nil
}
```

### internal/ssh/client.go

```go
package ssh

import "github.com/joshkornreich/anime/internal/logger"

func NewClient(host, user, keyPath string) (*Client, error) {
    logger.Debug("Creating SSH client",
        "host", host,
        "user", user,
        "key_path", keyPath,
    )

    // Try SSH agent first
    if agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        logger.Debug("Using SSH agent authentication")
        agentClient := agent.NewClient(agentConn)
        authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
    } else {
        logger.Debug("SSH agent not available", "error", err)
    }

    client, err := ssh.Dial("tcp", host, config)
    if err != nil {
        logger.Error("SSH connection failed",
            "host", host,
            "user", user,
            "error", err,
        )
        return nil, fmt.Errorf("failed to dial: %w", err)
    }

    logger.Info("SSH connection established",
        "host", host,
        "user", user,
    )

    return &Client{client: client, config: config, host: host}, nil
}
```

### internal/config/config.go

```go
package config

import "github.com/joshkornreich/anime/internal/logger"

func Load() (*Config, error) {
    path, err := GetConfigPath()
    if err != nil {
        logger.Error("Failed to get config path", "error", err)
        return nil, err
    }

    logger.Debug("Loading configuration", "path", path)

    // Create default config if doesn't exist
    if _, err := os.Stat(path); os.IsNotExist(err) {
        logger.Info("Config file not found, creating default",
            "path", path,
        )
        return &Config{...}, nil
    }

    logger.Info("Configuration loaded successfully",
        "path", path,
        "server_count", len(cfg.Servers),
    )

    return &cfg, nil
}

func (c *Config) Save() error {
    path, err := GetConfigPath()
    if err != nil {
        logger.Error("Failed to get config path", "error", err)
        return err
    }

    logger.Debug("Saving configuration",
        "path", path,
        "server_count", len(c.Servers),
    )

    if err := os.WriteFile(path, data, 0600); err != nil {
        logger.Error("Failed to write config file",
            "path", path,
            "error", err,
        )
        return err
    }

    logger.Info("Configuration saved successfully", "path", path)
    return nil
}
```

### cmd/deploy.go

```go
package cmd

import "github.com/joshkornreich/anime/internal/logger"

func runDeploy(cmd *cobra.Command, args []string) error {
    serverName := args[0]

    // Create operation-specific logger
    deployLogger := logger.WithOperation("deploy")
    deployLogger.Info("Starting deployment",
        "server", serverName,
        "modules", server.Modules,
    )

    // Keep fmt.Print for TUI banner
    fmt.Println(theme.InfoStyle.Render("🚀 Deploying to " + serverName))

    // Log detailed progress
    deployLogger.Debug("Creating installer instance")
    installer := installer.New(client)

    if err := installer.Install(server.Modules); err != nil {
        deployLogger.Error("Deployment failed",
            "server", serverName,
            "error", err,
        )
        return err
    }

    deployLogger.Info("Deployment completed successfully",
        "server", serverName,
        "duration_ms", duration,
    )

    // Keep TUI success message
    fmt.Println(theme.SuccessStyle.Render("✨ Deployment complete!"))

    return nil
}
```

### cmd/push.go

```go
package cmd

import "github.com/joshkornreich/anime/internal/logger"

func runPush(cmd *cobra.Command, args []string) error {
    target := parseServerTarget(args[0])

    pushLogger := logger.WithOperation("push")
    pushLogger.Info("Starting push operation",
        "target", target,
        "arch", pushArch,
        "include_source", pushIncludeSource,
    )

    // Keep TUI progress indicators
    fmt.Print(theme.DimTextStyle.Render("▶ Building binary... "))

    // Log build details
    pushLogger.Debug("Building Linux binary",
        "arch", pushArch,
        "ldflags", ldflags,
    )

    binaryPath, version, err := buildLinuxBinary()
    if err != nil {
        pushLogger.Error("Build failed", "error", err)
        fmt.Println(theme.ErrorStyle.Render("✗"))
        return err
    }

    pushLogger.Info("Binary built successfully",
        "version", version,
        "arch", pushArch,
        "size_bytes", binarySize,
    )

    // Keep TUI success
    fmt.Println(theme.SuccessStyle.Render("✓"))

    return nil
}
```

## Testing with Logging

When writing tests, you can verify log output:

```go
func TestWithLogging(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    logger.Logger = slog.New(handler)

    // Run your code
    DoSomething()

    // Verify logs
    output := buf.String()
    if !strings.Contains(output, "expected log message") {
        t.Errorf("Expected log not found: %s", output)
    }
}
```

## Debugging Tips

### Enable Debug Logging

```bash
# Enable debug logging for troubleshooting
anime deploy my-server --debug

# Or with explicit level
anime deploy my-server --log-level debug
```

### Log to File for Analysis

```bash
# Log to file for later analysis
anime deploy my-server --log-file /tmp/deploy.log --log-level debug

# View logs in real-time
tail -f /tmp/deploy.log
```

### JSON Logs for Log Aggregation

When logging to a file, JSON format is used automatically, making it easy to parse with tools like `jq`:

```bash
# Pretty-print JSON logs
jq '.' /tmp/deploy.log

# Filter by level
jq 'select(.level == "ERROR")' /tmp/deploy.log

# Filter by module
jq 'select(.module == "pytorch")' /tmp/deploy.log

# Extract specific fields
jq '{time, module, msg}' /tmp/deploy.log
```

## Best Practices

1. **Separate Concerns**: Use `fmt.Print` for user-facing TUI output, `logger.*` for operational/debug info
2. **Add Context**: Use `WithServer`, `WithModule`, `WithOperation` to add context
3. **Structured Fields**: Always use key-value pairs for structured data
4. **Log Errors**: Always log errors with context before returning them
5. **Debug Appropriately**: Use debug level for verbose information that helps troubleshooting
6. **Measure Performance**: Log durations, counts, and sizes for performance analysis
7. **Avoid Secrets**: Never log sensitive data (API keys, passwords, etc.)

## Migration Strategy

When updating existing code:

1. Identify `fmt.Print` calls that are for debugging/operational info
2. Replace with appropriate `logger.*` calls
3. Add structured fields for context
4. Keep user-facing TUI output as `fmt.Print`
5. Test with `--debug` flag to verify logging
