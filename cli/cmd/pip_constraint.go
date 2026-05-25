package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensurePipConstraint pins torch via PIP_CONSTRAINT so any pip command spawned
// by anime (install scripts, vllm doctor fixes, etc.) cannot silently swap
// torch onto a wheel built for the wrong CUDA version — the GH200 trap
// documented in lambda/cli/cmd/inference.go:258.
//
// Runs as rootCmd.PersistentPreRunE, so the env var is set before any
// subcommand executes. Subprocesses inherit os.Environ() and therefore the
// constraint, no per-call wiring needed.
//
// No-op when torch is not installed (initial bootstrap path).
func ensurePipConstraint() {
	torchVer, err := exec.Command("python3", "-c", "import torch; print(torch.__version__)").Output()
	if err != nil {
		return
	}
	ver := strings.TrimSpace(string(torchVer))
	if ver == "" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".config", "anime")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "torch-constraints.txt")

	content := fmt.Sprintf("torch==%s\ntorchvision\ntorchaudio\n", ver)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return
	}

	os.Setenv("PIP_CONSTRAINT", path)
}
