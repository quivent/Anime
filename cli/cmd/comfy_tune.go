package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/gpu"
)

// ComfyTuning is the bundle of CLI flags + environment variables we pass
// to ComfyUI's main.py to optimize for the detected host. Computed once
// at server start by AutoTuneComfyUI().
type ComfyTuning struct {
	Flags []string          // appended after `main.py --listen` (e.g. `--use-sage-attention`)
	Env   map[string]string // env vars set in the launch shell (e.g. PYTORCH_CUDA_ALLOC_CONF)
	Notes []string          // human-readable explanation of why each knob was chosen
}

// AutoTuneComfyUI returns the tuning bundle for a given host configuration.
// Heuristics:
//
//	Universal env (always set, no downsides on any modern CUDA):
//	  PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True  — prevents
//	      fragmentation OOMs on borderline-VRAM workflows
//	  NVIDIA_TF32_OVERRIDE=1                            — TF32 on Ampere+
//	  CUDA_MODULE_LOADING=LAZY                          — faster cold start
//	  PYTHONUNBUFFERED=1                                — clean tee streaming
//
//	Conditional flags (only when they help):
//	  --use-sage-attention   when sageInstalled (caller checks the venv)
//	  --reserve-vram 4       on ≥80GB GPUs (GH200/H200) — leaves room
//	                         for OS + any sidecar processes
//	  --lowvram              when per-GPU VRAM <12GB; ComfyUI usually
//	                         auto-detects but we belt-and-brace
//
//	Deliberately NOT forced:
//	  --highvram             ComfyUI auto-picks; forcing can OOM on
//	                         workflows that swap experts (Wan 2.2 dual)
//	  --fast                 Hopper fp8 matmul; version-dependent flag
//	                         that's easy to break across ComfyUI updates
//
// User overrides via the WAN_COMFY_FLAGS env var (space-separated tokens
// appended after the auto-tuned ones — last wins for repeated flags).
func AutoTuneComfyUI(g *gpu.SystemInfo, sageInstalled bool) ComfyTuning {
	t := ComfyTuning{
		Env: map[string]string{
			"PYTORCH_CUDA_ALLOC_CONF": "expandable_segments:True",
			"NVIDIA_TF32_OVERRIDE":    "1",
			"CUDA_MODULE_LOADING":     "LAZY",
			"PYTHONUNBUFFERED":        "1",
		},
	}
	t.Notes = append(t.Notes,
		"PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True (anti-frag)",
		"NVIDIA_TF32_OVERRIDE=1 (TF32 on Ampere+)",
		"CUDA_MODULE_LOADING=LAZY (faster cold start)")

	if !g.Available || len(g.GPUs) == 0 {
		t.Flags = append(t.Flags, "--cpu")
		t.Notes = append(t.Notes, "--cpu (no GPU detected — will be very slow)")
		return t
	}

	if sageInstalled {
		t.Flags = append(t.Flags, "--use-sage-attention")
		t.Notes = append(t.Notes, "--use-sage-attention (~30% faster on Wan 2.2)")
	}

	perGPU := g.GPUs[0].VRAM
	switch {
	case perGPU >= 80:
		// Big-iron path (H100 80GB / H200 / GH200 96GB+). ComfyUI's
		// default mode is fine but we leave 4GB for the OS + screen
		// session + any monitoring sidecars so a heavy workflow
		// doesn't fight nvtop/ssh-tunnel for the last byte.
		t.Flags = append(t.Flags, "--reserve-vram", "4")
		t.Notes = append(t.Notes, fmt.Sprintf("--reserve-vram 4 (≥80GB per-GPU: %dGB)", perGPU))
	case perGPU < 12:
		// Tight box. ComfyUI usually auto-picks --lowvram here but
		// some images report VRAM oddly through nvidia-smi (containers,
		// MIG slices); make it explicit to avoid swap thrash.
		t.Flags = append(t.Flags, "--lowvram")
		t.Notes = append(t.Notes, fmt.Sprintf("--lowvram (only %dGB per-GPU)", perGPU))
	default:
		t.Notes = append(t.Notes, fmt.Sprintf("default VRAM mode (%dGB per-GPU; ComfyUI auto-picks)", perGPU))
	}

	if extra := os.Getenv("WAN_COMFY_FLAGS"); extra != "" {
		t.Flags = append(t.Flags, strings.Fields(extra)...)
		t.Notes = append(t.Notes, "+ user override via $WAN_COMFY_FLAGS: "+extra)
	}

	return t
}

// EnvLines renders the env map as a stable, sorted slice of `KEY=value`
// strings — convenient to dump for diagnostics or pass to exec.Command.
func (t ComfyTuning) EnvLines() []string {
	out := make([]string, 0, len(t.Env))
	for k, v := range t.Env {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

// detectSageInstalled checks the ComfyUI venv site-packages for the
// sageattention package. Same glob the server-launch path uses; kept here
// so AutoTune callers don't have to recompute it.
func detectSageInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	matches, _ := filepath.Glob(filepath.Join(home, "ComfyUI", "venv", "lib", "python*", "site-packages", "sageattention"))
	return len(matches) > 0
}
