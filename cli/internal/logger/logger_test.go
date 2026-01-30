package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		outputFile string
		wantErr    bool
	}{
		{
			name:       "init with debug level to stderr",
			level:      slog.LevelDebug,
			outputFile: "",
			wantErr:    false,
		},
		{
			name:       "init with info level to stderr",
			level:      slog.LevelInfo,
			outputFile: "",
			wantErr:    false,
		},
		{
			name:       "init with file output",
			level:      slog.LevelInfo,
			outputFile: filepath.Join(t.TempDir(), "test.log"),
			wantErr:    false,
		},
		{
			name:       "init with nested directory",
			level:      slog.LevelInfo,
			outputFile: filepath.Join(t.TempDir(), "logs", "nested", "test.log"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.level, tt.outputFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && Logger == nil {
				t.Error("Logger should not be nil after successful Init()")
			}

			// Cleanup
			if tt.outputFile != "" && !tt.wantErr {
				os.Remove(tt.outputFile)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	Logger = slog.New(handler)

	tests := []struct {
		name     string
		logFunc  func(string, ...any)
		message  string
		wantLog  bool
		minLevel slog.Level
	}{
		{
			name:     "debug message",
			logFunc:  Debug,
			message:  "debug test",
			wantLog:  true,
			minLevel: slog.LevelDebug,
		},
		{
			name:     "info message",
			logFunc:  Info,
			message:  "info test",
			wantLog:  true,
			minLevel: slog.LevelInfo,
		},
		{
			name:     "warn message",
			logFunc:  Warn,
			message:  "warn test",
			wantLog:  true,
			minLevel: slog.LevelWarn,
		},
		{
			name:     "error message",
			logFunc:  Error,
			message:  "error test",
			wantLog:  true,
			minLevel: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message)

			output := buf.String()
			if tt.wantLog && !strings.Contains(output, tt.message) {
				t.Errorf("Expected log message %q not found in output: %s", tt.message, output)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	// Create temp file for JSON output
	tmpFile := filepath.Join(t.TempDir(), "test.log")
	err := Init(slog.LevelInfo, tmpFile)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Log some messages
	Info("test message", "key", "value", "count", 42)

	// Read the log file
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]any
	if err := json.Unmarshal(data, &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify fields
	if msg, ok := logEntry["msg"].(string); !ok || msg != "test message" {
		t.Errorf("Expected msg='test message', got %v", logEntry["msg"])
	}

	if level, ok := logEntry["level"].(string); !ok || level != "INFO" {
		t.Errorf("Expected level='INFO', got %v", logEntry["level"])
	}

	if key, ok := logEntry["key"].(string); !ok || key != "value" {
		t.Errorf("Expected key='value', got %v", logEntry["key"])
	}

	if count, ok := logEntry["count"].(float64); !ok || count != 42 {
		t.Errorf("Expected count=42, got %v", logEntry["count"])
	}
}

func TestWithServer(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	serverLogger := WithServer("test-server")
	if serverLogger == nil {
		t.Fatal("WithServer() returned nil")
	}

	serverLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "server=test-server") {
		t.Errorf("Expected server context in output: %s", output)
	}
}

func TestWithModule(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	moduleLogger := WithModule("test-module")
	if moduleLogger == nil {
		t.Fatal("WithModule() returned nil")
	}

	moduleLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "module=test-module") {
		t.Errorf("Expected module context in output: %s", output)
	}
}

func TestWithOperation(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	opLogger := WithOperation("deploy")
	if opLogger == nil {
		t.Fatal("WithOperation() returned nil")
	}

	opLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "operation=deploy") {
		t.Errorf("Expected operation context in output: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	fields := map[string]any{
		"server":    "test-server",
		"module":    "pytorch",
		"operation": "install",
		"user":      "ubuntu",
	}

	ctxLogger := WithContext(fields)
	if ctxLogger == nil {
		t.Fatal("WithContext() returned nil")
	}

	ctxLogger.Info("test message")

	output := buf.String()
	for key, value := range fields {
		expected := key + "=" + value.(string)
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %q in output: %s", expected, output)
		}
	}
}

func TestContextIntegration(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	// Create context with logger
	ctx := context.Background()
	ctxLogger := WithServer("ctx-server")
	ctx = ToContext(ctx, ctxLogger)

	// Retrieve logger from context
	retrievedLogger := FromContext(ctx)
	if retrievedLogger == nil {
		t.Fatal("FromContext() returned nil")
	}

	retrievedLogger.Info("context test")

	output := buf.String()
	if !strings.Contains(output, "server=ctx-server") {
		t.Errorf("Expected server context in output: %s", output)
	}
}

func TestIsDebugEnabled(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected bool
	}{
		{
			name:     "debug level enabled",
			level:    slog.LevelDebug,
			expected: true,
		},
		{
			name:     "info level disables debug",
			level:    slog.LevelInfo,
			expected: false,
		},
		{
			name:     "warn level disables debug",
			level:    slog.LevelWarn,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: tt.level,
			})
			Logger = slog.New(handler)

			result := IsDebugEnabled()
			if result != tt.expected {
				t.Errorf("IsDebugEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNilLogger(t *testing.T) {
	// Test that functions handle nil logger gracefully
	Logger = nil

	// These should not panic
	Debug("test")
	Info("test")
	Warn("test")
	Error("test")

	if WithServer("test") != nil {
		t.Error("WithServer() should return nil when Logger is nil")
	}

	if WithModule("test") != nil {
		t.Error("WithModule() should return nil when Logger is nil")
	}

	if WithOperation("test") != nil {
		t.Error("WithOperation() should return nil when Logger is nil")
	}

	if IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return false when Logger is nil")
	}
}

func TestMultipleContextLayers(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)

	// Chain multiple context layers
	serverLogger := WithServer("my-server")
	moduleLogger := serverLogger.With("module", "pytorch")
	opLogger := moduleLogger.With("operation", "install")

	opLogger.Info("multi-context test")

	output := buf.String()
	expectedParts := []string{
		"server=my-server",
		"module=pytorch",
		"operation=install",
		"multi-context test",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %q in output: %s", expected, output)
		}
	}
}
