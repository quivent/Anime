package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
)

func TestBashSyntax(t *testing.T) {
	scripts := GenerateScripts()
	t.Logf("Checking bash syntax for %d scripts", len(scripts))

	// Sort keys for deterministic output
	ids := make([]string, 0, len(scripts))
	for id := range scripts {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	tmpDir := t.TempDir()
	var failed []string

	for _, id := range ids {
		script := scripts[id]
		path := filepath.Join(tmpDir, id+".sh")

		if err := os.WriteFile(path, []byte(script), 0644); err != nil {
			t.Fatalf("Failed to write temp file for %q: %v", id, err)
		}

		cmd := exec.Command("bash", "-n", path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("SYNTAX ERROR in script %q:\n%s", id, string(output))
			failed = append(failed, id)
		} else {
			t.Logf("OK: %s", id)
		}
	}

	if len(failed) > 0 {
		t.Logf("\n--- SUMMARY: %d/%d scripts failed syntax check ---", len(failed), len(scripts))
		for _, id := range failed {
			t.Logf("  FAIL: %s", id)
		}
	} else {
		t.Logf("\n--- All %d scripts passed bash syntax check ---", len(scripts))
	}
}
