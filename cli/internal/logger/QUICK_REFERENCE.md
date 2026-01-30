# Structured Logging Quick Reference

## Import

```go
import "github.com/joshkornreich/anime/internal/logger"
```

## Basic Logging

```go
logger.Debug("message", "key", value)
logger.Info("message", "key", value)
logger.Warn("message", "key", value)
logger.Error("message", "key", value)
```

## Contextual Logging

```go
// Server context
serverLog := logger.WithServer("lambda-1")

// Module context
moduleLog := logger.WithModule("pytorch")

// Operation context
opLog := logger.WithOperation("deploy")

// Custom context
contextLog := logger.WithContext(map[string]any{
    "server": "lambda-1",
    "module": "pytorch",
})
```

## Common Patterns

### Start of Operation

```go
logger.Info("Starting operation",
    "operation", "deploy",
    "server", serverName,
    "modules", modules,
)
```

### Error Handling

```go
if err != nil {
    logger.Error("Operation failed",
        "operation", "deploy",
        "error", err,
    )
    return fmt.Errorf("failed: %w", err)
}
```

### Completion

```go
logger.Info("Operation complete",
    "operation", "deploy",
    "duration_ms", time.Since(start).Milliseconds(),
)
```

### Module Installation

```go
modLog := logger.WithModule(modID)
modLog.Info("Installing module",
    "name", module.Name,
    "estimated_time_minutes", module.TimeMinutes,
)
```

### File Operations

```go
logger.Debug("Reading file", "path", configPath)
logger.Info("Configuration loaded",
    "path", configPath,
    "size_bytes", len(data),
)
```

### Network Operations

```go
logger.Debug("Connecting to server",
    "host", host,
    "port", port,
)
logger.Info("Connection established",
    "host", host,
    "connection_time_ms", duration,
)
```

## CLI Usage

```bash
# Enable debug logging
anime command --debug

# Set log level
anime command --log-level info

# Log to file
anime command --log-file /tmp/anime.log

# Combined
anime command --log-file /tmp/anime.log --debug
```

## DO / DON'T

### ✅ DO

```go
// Structured fields
logger.Info("Installed", "module", "pytorch", "duration_ms", 480000)

// Log errors with context
logger.Error("Failed to connect", "host", host, "error", err)

// Create contextual loggers
modLog := logger.WithModule("pytorch")
modLog.Info("Starting")

// Keep TUI separate
fmt.Println(theme.SuccessStyle.Render("✨ Success!"))
logger.Info("Operation successful", "duration_ms", duration)
```

### ❌ DON'T

```go
// String concatenation
logger.Info(fmt.Sprintf("Installed %s in %dms", mod, dur))

// Mix TUI and logs
logger.Info("✨ Success!")  // Should be fmt.Print

// Repeat context
logger.Info("Step 1", "module", "pytorch")
logger.Info("Step 2", "module", "pytorch")  // Use WithModule instead
```

## Log Analysis

```bash
# Pretty-print JSON logs
jq '.' /tmp/anime.log

# Filter by level
jq 'select(.level == "ERROR")' /tmp/anime.log

# Filter by field
jq 'select(.module == "pytorch")' /tmp/anime.log

# Extract fields
jq '{time, module, msg}' /tmp/anime.log

# Find slow operations
jq 'select(.duration_ms > 10000)' /tmp/anime.log
```

## Testing

```go
func TestWithLogging(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    logger.Logger = slog.New(handler)

    // Run code
    MyFunction()

    // Verify logs
    output := buf.String()
    if !strings.Contains(output, "expected") {
        t.Error("Expected log not found")
    }
}
```

## Common Field Names

Use consistent names for common fields:

- `server`: Server name
- `host`: Server hostname/IP
- `module`: Module ID
- `operation`: Operation name (deploy, push, install)
- `error`: Error value
- `duration_ms`: Duration in milliseconds
- `size_bytes`: Size in bytes
- `path`: File path
- `port`: Network port
- `user`: Username
- `count`: Count of items
- `version`: Version string

## Cheat Sheet

| Task | Code |
|------|------|
| Import | `import "github.com/joshkornreich/anime/internal/logger"` |
| Info log | `logger.Info("msg", "key", val)` |
| Error log | `logger.Error("msg", "error", err)` |
| Debug log | `logger.Debug("msg", "key", val)` |
| Server context | `logger.WithServer(name)` |
| Module context | `logger.WithModule(id)` |
| Operation context | `logger.WithOperation(op)` |
| Check debug | `logger.IsDebugEnabled()` |

## Quick Migration

1. Add import: `import "github.com/joshkornreich/anime/internal/logger"`
2. Find debug `fmt.Print`: Replace with `logger.Debug()`
3. Add error logging: Before each `return err`
4. Add operation logging: Start, errors, completion
5. Test: `go run . --debug`
