package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Global logger instance
var Logger *slog.Logger

// ContextKey is used for storing logger context in context.Context
type ContextKey string

const (
	ServerKey    ContextKey = "server"
	ModuleKey    ContextKey = "module"
	OperationKey ContextKey = "operation"
)

// Init initializes the global logger with the specified level and output
func Init(level slog.Level, outputFile string) error {
	var writer io.Writer = os.Stderr

	// If output file is specified, create/open it
	if outputFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		writer = file
	}

	// Determine handler based on output type
	var handler slog.Handler
	if outputFile != "" {
		// Use JSON handler for file output (better for log aggregation)
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Customize attribute names if needed
				return a
			},
		})
	} else {
		// Use text handler for terminal output (more readable)
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	}

	Logger = slog.New(handler)
	return nil
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if Logger != nil {
		Logger.Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	if Logger != nil {
		Logger.Info(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if Logger != nil {
		Logger.Warn(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if Logger != nil {
		Logger.Error(msg, args...)
	}
}

// WithServer adds server context to the logger
func WithServer(name string) *slog.Logger {
	if Logger == nil {
		return nil
	}
	return Logger.With("server", name)
}

// WithModule adds module context to the logger
func WithModule(id string) *slog.Logger {
	if Logger == nil {
		return nil
	}
	return Logger.With("module", id)
}

// WithOperation adds operation context to the logger
func WithOperation(op string) *slog.Logger {
	if Logger == nil {
		return nil
	}
	return Logger.With("operation", op)
}

// WithContext creates a logger with multiple context fields
func WithContext(fields map[string]any) *slog.Logger {
	if Logger == nil {
		return nil
	}

	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return Logger.With(args...)
}

// FromContext extracts logger from context, falls back to global logger
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return Logger
	}

	// Try to extract logger from context
	if ctxLogger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return ctxLogger
	}

	return Logger
}

// ToContext adds logger to context
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

// GetLevel returns the current log level as a string
func GetLevel() string {
	if Logger == nil {
		return "info"
	}

	// Try to determine level from handler
	// This is a simplified approach - in production you might want to store the level
	return "info"
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	if Logger == nil {
		return false
	}
	return Logger.Enabled(context.Background(), slog.LevelDebug)
}
