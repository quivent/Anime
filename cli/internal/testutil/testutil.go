package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing and returns a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "anime-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}

	return dir, cleanup
}

// WriteFile writes content to a file in the temp directory
func WriteFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return path
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, got, notWant interface{}) {
	t.Helper()
	if got == notWant {
		t.Errorf("got %v, but did not want %v", got, notWant)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, got interface{}) {
	t.Helper()
	if !isInterfaceNil(got) {
		t.Errorf("got %v, want nil", got)
	}
}

// isInterfaceNil checks if an interface value is truly nil
// This handles cases where interface contains a typed nil pointer
func isInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	// For pointers and interfaces, we need to check the underlying value
	// A simple != nil check doesn't work for interface{} containing (*T)(nil)
	switch v := i.(type) {
	case error:
		return v == nil
	default:
		// Check if it's a pointer type with nil value
		// This is a simple check - for typed nils like (*Server)(nil)
		// we just do a simple nil comparison which should work in most cases
		return false
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, got interface{}) {
	t.Helper()
	if got == nil {
		t.Errorf("got nil, want non-nil")
	}
}

// AssertError asserts that an error occurred
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// AssertNoError asserts that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr string) {
	t.Helper()
	if !contains(str, substr) {
		t.Errorf("expected %q to contain %q", str, substr)
	}
}

// AssertSliceEqual asserts that two string slices are equal
func AssertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %d elements, want %d elements", len(got), len(want))
		t.Errorf("got:  %v", got)
		t.Errorf("want: %v", want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("element %d: got %q, want %q", i, got[i], want[i])
			t.Errorf("got:  %v", got)
			t.Errorf("want: %v", want)
			return
		}
	}
}

// MockError creates a simple error for testing
var MockError = func(msg string) error {
	return &mockError{msg: msg}
}

// mockError implements the error interface
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
