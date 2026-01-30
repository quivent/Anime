// +build ignore

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joshkornreich/anime/internal/logger"
)

func main() {
	fmt.Println("=== Structured Logging Demo ===")
	fmt.Println()

	// Demo 1: Text output to stderr (development)
	fmt.Println("1. Text logging to stderr (development mode):")
	logger.Init(slog.LevelDebug, "")

	logger.Debug("This is a debug message", "key", "value")
	logger.Info("Application started", "version", "1.0.0")
	logger.Warn("Configuration not optimized", "setting", "cache_size")
	logger.Error("Failed to connect", "error", "connection refused")

	fmt.Println()
	time.Sleep(100 * time.Millisecond)

	// Demo 2: JSON output to file (production)
	fmt.Println("2. JSON logging to file (production mode):")
	tmpFile := "/tmp/anime-demo.log"
	logger.Init(slog.LevelInfo, tmpFile)

	logger.Info("Module installation started",
		"module", "pytorch",
		"estimated_time_minutes", 8,
		"dependencies", []string{"core", "cuda"},
	)

	logger.Warn("Slow download detected",
		"module", "pytorch",
		"speed_mbps", 2.5,
	)

	logger.Info("Module installation complete",
		"module", "pytorch",
		"duration_ms", 480000,
	)

	fmt.Println("Logs written to:", tmpFile)
	fmt.Println()

	// Show the JSON log content
	data, _ := os.ReadFile(tmpFile)
	fmt.Println("Log file content:")
	fmt.Println(string(data))

	// Demo 3: Contextual logging
	fmt.Println("3. Contextual logging (with server/module/operation context):")
	logger.Init(slog.LevelDebug, "")

	// Server context
	serverLogger := logger.WithServer("lambda-1")
	serverLogger.Info("Connecting to server")

	// Module context
	modLogger := logger.WithModule("pytorch")
	modLogger.Info("Installing module", "name", "PyTorch + AI Libraries")

	// Operation context
	opLogger := logger.WithOperation("deploy")
	opLogger.Info("Deployment started", "target", "lambda-1")

	// Multiple context layers
	multiLogger := logger.WithServer("lambda-1")
	multiLogger = multiLogger.With("module", "pytorch")
	multiLogger = multiLogger.With("operation", "install")
	multiLogger.Info("Complex operation",
		"step", "dependencies",
		"duration_ms", 1234,
	)

	fmt.Println()

	// Demo 4: Custom context
	fmt.Println("4. Custom context fields:")
	logger.Init(slog.LevelInfo, "")

	fields := map[string]any{
		"server":    "lambda-1",
		"module":    "pytorch",
		"operation": "install",
		"user":      "ubuntu",
	}

	contextLogger := logger.WithContext(fields)
	contextLogger.Info("Operation started with full context")

	fmt.Println()

	// Demo 5: Debug level filtering
	fmt.Println("5. Log level filtering (set to INFO, debug messages hidden):")
	logger.Init(slog.LevelInfo, "")

	logger.Debug("This debug message will NOT appear")
	logger.Info("This info message WILL appear")
	logger.Warn("This warning message WILL appear")

	fmt.Println()

	// Cleanup
	os.Remove(tmpFile)

	fmt.Println("=== Demo Complete ===")
}
