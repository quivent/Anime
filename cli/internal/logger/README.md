# Structured Logging Infrastructure

This package provides a production-ready structured logging system for the anime CLI using Go 1.21+ `log/slog`.

## Features

- **Structured logging** with key-value pairs
- **Multiple output formats**: Text (for terminal) and JSON (for log aggregation)
- **Log levels**: Debug, Info, Warn, Error
- **Contextual loggers**: Add server, module, and operation context
- **Global configuration**: Controlled via command-line flags
- **Thread-safe**: Safe for concurrent use
- **Zero dependencies**: Uses only standard library `log/slog`

## Quick Start

### For Users

The logger is automatically initialized when using the anime CLI. Control logging with flags:

```bash
# Enable debug logging
anime deploy my-server --debug

# Set specific log level
anime push lambda --log-level info

# Log to file (JSON format)
anime deploy my-server --log-file /tmp/anime.log --log-level debug
```

### For Developers

```go
import "github.com/joshkornreich/anime/internal/logger"

// Simple logging
logger.Info("Operation started", "module", "pytorch")
logger.Error("Connection failed", "error", err)

// Contextual logging
modLogger := logger.WithModule("pytorch")
modLogger.Info("Installing", "estimated_time_minutes", 8)

// Multiple context layers
serverLogger := logger.WithServer("lambda-1")
opLogger := serverLogger.With("operation", "deploy")
opLogger.Info("Deployment started")
```

## Architecture

### Components

1. **Global Logger**: `logger.Logger` - The singleton logger instance
2. **Initialization**: `logger.Init()` - Sets up the logger with level and output
3. **Convenience Functions**: `Debug()`, `Info()`, `Warn()`, `Error()` - Simple logging
4. **Context Helpers**: `WithServer()`, `WithModule()`, `WithOperation()` - Add context

### Output Formats

**Text Output (Terminal)**:
```
time=2025-12-13T10:30:45.123Z level=INFO msg="Module installed" module=pytorch duration_ms=480000
```

**JSON Output (File)**:
```json
{"time":"2025-12-13T10:30:45.123Z","level":"INFO","msg":"Module installed","module":"pytorch","duration_ms":480000}
```

### Log Levels

- **Debug** (`--log-level debug` or `--debug`): Verbose diagnostic information
- **Info** (`--log-level info`): General informational messages
- **Warn** (`--log-level warn`): Warning messages
- **Error** (`--log-level error`, default): Error messages only

## Design Principles

### 1. Separation of Concerns

**Keep user-facing TUI output separate from operational logging:**

```go
// ✅ DO: TUI output for users
fmt.Println(theme.SuccessStyle.Render("✨ Deployment complete!"))
fmt.Printf("Target: %s\n", theme.HighlightStyle.Render(target))

// ✅ DO: Structured logging for operators/debugging
logger.Info("Deployment completed",
    "target", target,
    "duration_ms", duration,
    "modules_installed", len(modules),
)
```

```go
// ❌ DON'T: Mix concerns
logger.Info("✨ Deployment complete!") // This should be fmt.Print
logger.Info("Target: " + target)       // This should be structured
```

### 2. Structured Fields

Always use key-value pairs for structured data:

```go
// ✅ DO: Structured fields
logger.Info("Module installed",
    "module", "pytorch",
    "duration_ms", 480000,
    "size_mb", 8192,
)

// ❌ DON'T: String concatenation
logger.Info(fmt.Sprintf("Module %s installed in %dms", module, duration))
```

### 3. Contextual Logging

Use contextual loggers to avoid repeating fields:

```go
// ✅ DO: Create contextual logger
modLogger := logger.WithModule("pytorch")
modLogger.Info("Starting installation")
modLogger.Debug("Downloading dependencies")
modLogger.Info("Installation complete")

// ❌ DON'T: Repeat context in every call
logger.Info("Starting installation", "module", "pytorch")
logger.Debug("Downloading dependencies", "module", "pytorch")
logger.Info("Installation complete", "module", "pytorch")
```

### 4. Error Logging

Always log errors with context before returning:

```go
// ✅ DO: Log error with context
if err := installModule(modID); err != nil {
    logger.Error("Module installation failed",
        "module", modID,
        "error", err,
    )
    return fmt.Errorf("failed to install %s: %w", modID, err)
}

// ❌ DON'T: Return error without logging
if err := installModule(modID); err != nil {
    return err
}
```

## API Reference

### Initialization

```go
func Init(level slog.Level, outputFile string) error
```

Initializes the global logger with the specified level and output file. If `outputFile` is empty, logs to stderr in text format. If specified, logs to file in JSON format.

### Convenience Functions

```go
func Debug(msg string, args ...any)
func Info(msg string, args ...any)
func Warn(msg string, args ...any)
func Error(msg string, args ...any)
```

Log messages at the specified level with optional key-value pairs.

**Example:**
```go
logger.Info("Server connected", "host", "192.168.1.100", "port", 22)
```

### Context Helpers

```go
func WithServer(name string) *slog.Logger
func WithModule(id string) *slog.Logger
func WithOperation(op string) *slog.Logger
func WithContext(fields map[string]any) *slog.Logger
```

Create contextual loggers with predefined fields.

**Example:**
```go
serverLogger := logger.WithServer("lambda-1")
serverLogger.Info("Connection established")
// Output: ... msg="Connection established" server=lambda-1
```

### Context Integration

```go
func FromContext(ctx context.Context) *slog.Logger
func ToContext(ctx context.Context, logger *slog.Logger) context.Context
```

Store and retrieve loggers from `context.Context`.

**Example:**
```go
ctx := logger.ToContext(ctx, logger.WithServer("lambda-1"))
log := logger.FromContext(ctx)
log.Info("Processing request")
```

### Utility Functions

```go
func IsDebugEnabled() bool
```

Returns true if debug logging is enabled.

**Example:**
```go
if logger.IsDebugEnabled() {
    // Expensive debug operation
    details := collectDetailedMetrics()
    logger.Debug("Detailed metrics", "data", details)
}
```

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./internal/logger -v

# Run specific test
go test ./internal/logger -run TestInit -v

# Run with coverage
go test ./internal/logger -cover
```

### Testing with Logging

```go
func TestMyFunction(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    logger.Logger = slog.New(handler)

    // Run your code
    MyFunction()

    // Verify logs
    output := buf.String()
    if !strings.Contains(output, "expected message") {
        t.Errorf("Expected log not found: %s", output)
    }
}
```

## Usage Examples

See [USAGE.md](USAGE.md) for comprehensive usage guide.

See [EXAMPLES.md](EXAMPLES.md) for file-specific implementation examples.

## Migration Guide

### Step 1: Identify Logging Opportunities

Look for:
- `fmt.Print*` calls used for debugging/operational info
- Error returns without logging
- Complex operations without progress tracking

### Step 2: Add Import

```go
import "github.com/joshkornreich/anime/internal/logger"
```

### Step 3: Replace Debug Print Statements

```go
// Before
fmt.Printf("DEBUG: Connecting to %s\n", host)

// After
logger.Debug("Connecting to server", "host", host)
```

### Step 4: Add Operational Logging

```go
// Add to start of operation
logger.Info("Starting deployment",
    "server", serverName,
    "modules", modules,
)

// Add to error paths
if err != nil {
    logger.Error("Deployment failed",
        "server", serverName,
        "error", err,
    )
    return err
}

// Add to completion
logger.Info("Deployment complete",
    "server", serverName,
    "duration_ms", duration,
)
```

### Step 5: Use Contextual Loggers

```go
// Create once per operation/server/module
deployLogger := logger.WithOperation("deploy")

// Use throughout the operation
deployLogger.Info("Starting")
deployLogger.Debug("Loading config")
deployLogger.Info("Complete")
```

## Performance Considerations

1. **Lazy Evaluation**: Expensive operations should check if the level is enabled:
   ```go
   if logger.IsDebugEnabled() {
       details := expensiveOperation()
       logger.Debug("Details", "data", details)
   }
   ```

2. **Context Reuse**: Create contextual loggers once and reuse:
   ```go
   modLogger := logger.WithModule("pytorch")  // Create once
   modLogger.Info("Step 1")                   // Reuse
   modLogger.Info("Step 2")                   // Reuse
   ```

3. **Structured Fields**: Prefer structured fields over string concatenation:
   ```go
   // Faster - no string allocation
   logger.Info("Installed", "module", mod, "time", dur)

   // Slower - creates temporary strings
   logger.Info(fmt.Sprintf("Installed %s in %v", mod, dur))
   ```

## Log Analysis

### Using jq for JSON Logs

```bash
# Pretty-print
jq '.' /tmp/anime.log

# Filter by level
jq 'select(.level == "ERROR")' /tmp/anime.log

# Filter by module
jq 'select(.module == "pytorch")' /tmp/anime.log

# Extract specific fields
jq '{time, module, msg, duration_ms}' /tmp/anime.log

# Count errors by module
jq 'select(.level == "ERROR") | .module' /tmp/anime.log | sort | uniq -c

# Find slow operations
jq 'select(.duration_ms > 10000) | {module, duration_ms, msg}' /tmp/anime.log
```

### Log Aggregation

JSON logs can be easily ingested into log aggregation systems:

- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **Splunk**
- **CloudWatch Logs**
- **Datadog**

## Troubleshooting

### Logs Not Appearing

1. Check log level: `--log-level debug`
2. Check if logger is initialized (automatic in CLI)
3. Verify output destination (stderr by default)

### Too Many Logs

1. Increase log level: `--log-level warn` or `--log-level error`
2. Review debug statements and remove unnecessary ones

### Performance Impact

1. Use `IsDebugEnabled()` for expensive operations
2. Avoid logging in tight loops
3. Use appropriate log levels

## Best Practices Summary

1. ✅ **DO** use structured key-value pairs
2. ✅ **DO** log errors with context before returning
3. ✅ **DO** use contextual loggers for related operations
4. ✅ **DO** keep TUI output separate from operational logs
5. ✅ **DO** include timing information for long operations
6. ✅ **DO** use appropriate log levels
7. ❌ **DON'T** log sensitive data (API keys, passwords)
8. ❌ **DON'T** replace user-facing output with logs
9. ❌ **DON'T** use string concatenation for structured data
10. ❌ **DON'T** log in tight loops without throttling

## License

Part of the anime CLI project.
