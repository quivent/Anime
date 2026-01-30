package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEmbeddedSourceExists verifies that source code is embedded in the binary.
// This test MUST pass for the build to be considered valid for distribution.
func TestEmbeddedSourceExists(t *testing.T) {
	if !HasEmbeddedSource() {
		t.Fatal("CRITICAL: No embedded source code found in binary. " +
			"Build with 'make build' to embed source. " +
			"This is required for 'anime extract' and 'anime develop' to work on deployed systems.")
	}
}

// TestEmbeddedSourceSize verifies the embedded source has reasonable size
func TestEmbeddedSourceSize(t *testing.T) {
	size := GetEmbeddedSourceSize()
	if size == 0 {
		t.Fatal("Embedded source size is 0 - source not properly embedded")
	}

	// Source should be at least 100KB (sanity check)
	minSize := int64(100 * 1024)
	if size < minSize {
		t.Errorf("Embedded source size %d bytes is suspiciously small (expected > %d bytes)", size, minSize)
	}

	// Log the size for visibility
	t.Logf("Embedded source size: %d bytes (%.2f MB)", size, float64(size)/(1024*1024))
}

// TestBuildDirSet verifies that BuildDir is set during proper builds.
// Note: During 'go test', ldflags are NOT applied so BuildDir will always be empty.
// This test documents the expected behavior rather than enforcing it.
func TestBuildDirSet(t *testing.T) {
	// Note: BuildDir is set via ldflags during 'make build', not during 'go test'
	// So this test can only verify the variable exists and log its value
	if BuildDir == "" {
		t.Log("BuildDir is empty (expected during 'go test' since ldflags aren't applied)")
		t.Log("When built with 'make build', BuildDir will contain the source directory path")
	} else {
		t.Logf("BuildDir: %s", BuildDir)
	}
}

// TestExtractEmbeddedSource verifies that embedded source can be extracted
func TestExtractEmbeddedSource(t *testing.T) {
	if !HasEmbeddedSource() {
		t.Skip("No embedded source available - skipping extraction test")
	}

	// Create a temporary directory for extraction
	tmpDir, err := os.MkdirTemp("", "anime-src-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract the source
	extractedPath, err := ExtractEmbeddedSource(tmpDir)
	if err != nil {
		t.Fatalf("Failed to extract embedded source: %v", err)
	}

	// Verify go.mod exists in extracted source
	goModPath := filepath.Join(extractedPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod not found in extracted source at %s", goModPath)
	}

	// Verify main.go exists
	mainGoPath := filepath.Join(extractedPath, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("main.go not found in extracted source at %s", mainGoPath)
	}

	// Verify cmd directory exists
	cmdDir := filepath.Join(extractedPath, "cmd")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		t.Errorf("cmd/ directory not found in extracted source at %s", cmdDir)
	}

	// Verify Makefile exists
	makefilePath := filepath.Join(extractedPath, "Makefile")
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		t.Errorf("Makefile not found in extracted source at %s", makefilePath)
	}

	t.Logf("Successfully extracted source to: %s", extractedPath)
}

// TestSourceCacheDir verifies the cache directory path is valid
func TestSourceCacheDir(t *testing.T) {
	cacheDir := GetSourceCacheDir()
	if cacheDir == "" {
		t.Error("GetSourceCacheDir returned empty string")
	}

	// Should contain anime-cli-source
	if !filepath.IsAbs(cacheDir) || filepath.Base(cacheDir) != "anime-cli-source" {
		t.Errorf("Unexpected cache directory format: %s", cacheDir)
	}

	t.Logf("Cache directory: %s", cacheDir)
}
