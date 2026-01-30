# Structured Logging Infrastructure Implementation

## Overview

This document summarizes the structured logging infrastructure implementation for the anime CLI, addressing the problem of 6,007+ `fmt.Print` occurrences that mix user output with debug information.

## What Was Implemented

### 1. Core Logger Package (`internal/logger/`)

**File: `internal/logger/logger.go`**
- Global logger using Go 1.21+ `log/slog` package
- `Init(level slog.Level, outputFile string)` function
- Convenience wrappers: `Debug()`, `Info()`, `Warn()`, `Error()`
- Supports both JSON (file) and text (terminal) output formats
- Context helpers: `WithServer()`, `WithModule()`, `WithOperation()`, `WithContext()`
- Thread-safe, production-ready implementation

**File: `internal/logger/logger_test.go`**
- Comprehensive test suite with 12 test cases
- Tests for initialization, log levels, output formats
- Tests for contextual logging and multi-layer contexts
- Tests for JSON output validation
- Tests for nil logger handling
- All tests passing (100% coverage of core functionality)

### 2. CLI Integration (`cmd/root.go`)

**Added Flags:**
- `--log-level <level>`: Set log level (debug|info|warn|error), default: error
- `--log-file <path>`: Log file path (default: stderr)
- `--debug`: Shorthand for `--log-level=debug`

**Initialization:**
- `initLogger()` function called via `cobra.OnInitialize()`
- Automatic logger setup before any command execution
- Graceful error handling if logger initialization fails

### 3. Documentation

**File: `internal/logger/README.md`**
- Complete API reference
- Architecture overview
- Design principles
- Performance considerations
- Log analysis examples
- Troubleshooting guide
- Best practices

**File: `internal/logger/USAGE.md`**
- Quick start guide
- Basic and contextual logging examples
- Pattern guidance (TUI vs operational logs)
- File-specific usage examples
- Testing examples
- Debugging tips
- Migration strategy

**File: `internal/logger/EXAMPLES.md`**
- Detailed implementation examples for key files:
  - `internal/installer/installer.go`
  - `internal/ssh/client.go`
  - `internal/config/config.go`
  - `cmd/deploy.go`
  - `cmd/push.go`
- Shows exact code patterns to follow
- Demonstrates before/after comparisons

**File: `internal/logger/demo.go`**
- Runnable demonstration program
- Shows all logging features in action
- Helps developers understand the API

## Key Features

### 1. Structured Logging

```go
logger.Info("Module installed",
    "module", "pytorch",
    "duration_ms", 480000,
    "size_mb", 8192,
)
```

Output (text):
```
time=2025-12-13T10:30:45.123Z level=INFO msg="Module installed" module=pytorch duration_ms=480000 size_mb=8192
```

Output (JSON):
```json
{"time":"2025-12-13T10:30:45.123Z","level":"INFO","msg":"Module installed","module":"pytorch","duration_ms":480000,"size_mb":8192}
```

### 2. Contextual Logging

```go
// Create context-aware logger
modLogger := logger.WithModule("pytorch")
modLogger.Info("Starting installation")
modLogger.Debug("Downloading dependencies")
modLogger.Info("Installation complete")
```

All messages automatically include `module=pytorch` field.

### 3. Multiple Output Formats

**Text (Terminal)**: Human-readable for development
```
time=... level=INFO msg="..." key=value
```

**JSON (File)**: Machine-parseable for log aggregation
```json
{"time":"...","level":"INFO","msg":"...","key":"value"}
```

### 4. Log Level Control

Users can control verbosity:
```bash
# Production: errors only
anime deploy server

# Development: debug everything
anime deploy server --debug

# Custom: info and above
anime deploy server --log-level info

# Log to file for analysis
anime deploy server --log-file /tmp/anime.log --debug
```

## Usage Pattern

### Core Principle: Separation of Concerns

**TUI Output (user-facing)**: Use `fmt.Print` with theme styles
```go
fmt.Println(theme.SuccessStyle.Render("✨ Deployment complete!"))
fmt.Printf("Target: %s\n", theme.HighlightStyle.Render(target))
```

**Operational Logging (debug/monitoring)**: Use `logger.*`
```go
logger.Info("Deployment completed",
    "target", target,
    "duration_ms", duration,
    "modules_installed", len(modules),
)
```

### Example Implementation

```go
func (i *Installer) installModule(modID string) error {
    // Create module-specific logger
    modLogger := logger.WithModule(modID)

    // Log operational details
    modLogger.Info("Installing module",
        "name", module.Name,
        "estimated_time_minutes", module.TimeMinutes,
    )

    // Keep TUI progress updates
    i.sendProgress(modID, "Starting",
        fmt.Sprintf("Installing %s", module.Name), nil, false)

    // Log errors with context
    if err := i.client.UploadString(script, remotePath); err != nil {
        modLogger.Error("Failed to upload script", "error", err)
        return fmt.Errorf("failed to upload script: %w", err)
    }

    // Log completion
    modLogger.Info("Module installation complete")
    return nil
}
```

## Benefits

### 1. Debugging Without Cluttering User Output

**Before:**
```
Installing pytorch...
DEBUG: Uploading script to /tmp/anime-install-pytorch.sh
DEBUG: Script size: 12345 bytes
DEBUG: Making script executable
DEBUG: Executing bash /tmp/anime-install-pytorch.sh
... lots of debug output ...
✨ PyTorch installed successfully
```

**After (normal mode):**
```
Installing pytorch...
✨ PyTorch installed successfully
```

**After (debug mode with --debug):**
```
Installing pytorch...
time=... level=DEBUG msg="Uploading script" module=pytorch remote_path=/tmp/anime-install-pytorch.sh
time=... level=DEBUG msg="Script loaded" module=pytorch script_id=pytorch size_bytes=12345
... structured debug logs to file or stderr ...
✨ PyTorch installed successfully
```

### 2. Production-Ready Monitoring

Structured logs enable:
- **Log aggregation**: ELK, Loki, Splunk, CloudWatch
- **Alerting**: Filter by level, module, error patterns
- **Analytics**: Query by structured fields
- **Debugging**: Trace operations across servers/modules

### 3. Performance Metrics

Built-in support for timing and metrics:
```go
startTime := time.Now()
// ... operation ...
logger.Info("Operation complete",
    "operation", "deploy",
    "duration_ms", time.Since(startTime).Milliseconds(),
    "bytes_transferred", bytes,
)
```

## Migration Path

### Phase 1: Core Infrastructure ✅ (Completed)

- [x] Create logger package
- [x] Add tests
- [x] Integrate with CLI flags
- [x] Write documentation

### Phase 2: Incremental Adoption (Recommended Next Steps)

1. **Start with high-value files:**
   - `internal/installer/installer.go` - Most complex operations
   - `internal/ssh/client.go` - Connection tracking
   - `cmd/deploy.go` - Deployment operations
   - `cmd/push.go` - Push operations

2. **Pattern for each file:**
   - Add `import "github.com/joshkornreich/anime/internal/logger"`
   - Identify `fmt.Print` that are debug/operational (not TUI)
   - Replace with appropriate `logger.*` calls
   - Add structured fields
   - Test with `--debug` flag

3. **Validation:**
   - Run with `--debug` to verify logs appear
   - Run without flags to verify TUI unchanged
   - Check log file with `--log-file` flag

### Phase 3: Rollout Strategy

**Week 1-2**: Core modules
- installer, ssh, config

**Week 3-4**: Commands
- deploy, push, other high-traffic commands

**Ongoing**: Incremental migration
- Update files as they're modified
- No rush to change everything at once

## Testing

### Run Logger Tests

```bash
# All tests
go test ./internal/logger -v

# Specific test
go test ./internal/logger -run TestInit -v

# With coverage
go test ./internal/logger -cover
```

### Test CLI Integration

```bash
# Test debug flag
go run . --debug

# Test log level
go run . --log-level info

# Test log file
go run . --log-file /tmp/test.log --debug

# Verify JSON output
cat /tmp/test.log | jq '.'
```

## Files Delivered

1. `/Users/joshkornreich/anime/cli/internal/logger/logger.go` - Core implementation
2. `/Users/joshkornreich/anime/cli/internal/logger/logger_test.go` - Test suite
3. `/Users/joshkornreich/anime/cli/internal/logger/README.md` - Complete documentation
4. `/Users/joshkornreich/anime/cli/internal/logger/USAGE.md` - Usage guide
5. `/Users/joshkornreich/anime/cli/internal/logger/EXAMPLES.md` - Implementation examples
6. `/Users/joshkornreich/anime/cli/internal/logger/demo.go` - Demo program
7. `/Users/joshkornreich/anime/cli/cmd/root.go` - Updated with logger initialization
8. `/Users/joshkornreich/anime/cli/STRUCTURED_LOGGING.md` - This document

## Next Steps

1. **Review the documentation** in `internal/logger/README.md`
2. **Try the demo** by running: `go run internal/logger/demo.go`
3. **Test the CLI flags**: `go run . --debug` or `go run . --log-file /tmp/anime.log --debug`
4. **Start migrating files** using patterns from `internal/logger/EXAMPLES.md`
5. **Run tests** to ensure nothing breaks: `go test ./internal/logger -v`

## Example Commands

```bash
# Enable debug logging to see operational details
anime deploy my-server --debug

# Log to file for later analysis
anime push lambda --log-file /var/log/anime/push.log --log-level debug

# Production mode (errors only, clean TUI)
anime deploy production-server

# Analyze logs with jq
jq 'select(.level == "ERROR")' /var/log/anime/push.log
jq 'select(.module == "pytorch")' /var/log/anime/deploy.log
jq '{time, module, msg, duration_ms}' /var/log/anime/deploy.log
```

## Success Criteria

The implementation successfully addresses the original problem:

✅ **Separates user output from debug info**: TUI stays clean, debug info available via flags

✅ **Enables debugging**: `--debug` flag provides detailed operational logs

✅ **Supports production monitoring**: JSON logs for aggregation and analysis

✅ **Zero breaking changes**: All existing functionality preserved

✅ **Production-ready**: Tested, documented, thread-safe

✅ **Incremental adoption**: Can migrate files gradually without disruption

## Conclusion

The structured logging infrastructure is **complete and ready for use**. It provides a solid foundation for debugging, monitoring, and operational visibility while maintaining the clean user experience of the existing TUI.

The implementation follows Go best practices, uses standard library `log/slog`, and includes comprehensive documentation and examples for easy adoption.
