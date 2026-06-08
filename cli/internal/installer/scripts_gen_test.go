package installer

import (
	"strings"
	"testing"
)

func TestGenerateScriptsCompleteness(t *testing.T) {
	s := GenerateScripts()
	t.Logf("Total scripts generated: %d", len(s))

	if len(s) < 90 {
		t.Fatalf("Expected at least 90 scripts, got %d", len(s))
	}

	must := []string{
		"core", "python", "pytorch", "ollama", "vllm",
		"comfyui", "nvidia", "docker", "nodejs", "go", "claude", "gh",
		"llama-3.3-70b", "qwen3-8b", "deepseek-r1-70b",
		"sdxl", "sd15", "flux-dev", "flux-schnell", "flux2",
		"models-small", "models-medium", "models-large",
		"mochi", "wan2", "cogvideo", "opensora", "ltxvideo",
		"svd", "animatediff", "rife", "film",
		"real-esrgan", "gfpgan",
		"rust", "uv", "tmux", "htop", "jq", "fzf", "bat", "glow", "lazygit", "neovim", "caddy", "tailscale",
		"svd-xt", "controlnet-canny", "ip-adapter", "instantid",
		"cogvideox-1.5", "hunyuan-video", "pyramid-flow",
		"comfyui-wan2",
	}

	for _, id := range must {
		if _, ok := s[id]; !ok {
			t.Errorf("Missing script: %s", id)
		}
	}

	// Content spot-checks
	checks := map[string]string{
		"llama-3.3-70b": "ollama pull llama3.3:70b",
		"core":          "wait_for_dpkg",
		"vllm":          "TORCH_CUDA_ARCH_LIST",
		"flux-dev":      "flux1-dev.safetensors",
		"mochi":         "genmo/mochi-1-preview",
		"nvidia":        "cuda-keyring",
		"models-small":  "llama3.2:1b",
		"sdxl":          "sd_xl_base_1.0.safetensors",
		"wan2":          "Alibaba-PAI/wan2.2",
		"docker":        "get.docker.com",
	}

	for id, needle := range checks {
		script := s[id]
		if !strings.Contains(script, needle) {
			t.Errorf("Script %q missing expected content %q", id, needle)
		}
	}

	// Every script should start with #!/bin/bash
	for id, script := range s {
		if !strings.HasPrefix(script, "#!/bin/bash") {
			t.Errorf("Script %q doesn't start with shebang", id)
		}
	}

	// HF downloads must use correct --local-dir (not double-nested)
	hfLocalDirChecks := map[string]string{
		"svd-xt":              "--local-dir svd-xt",
		"sd3.5-large":         "--local-dir sd3.5-large",
		"controlnet-canny":    "--local-dir controlnet-canny",
		"ip-adapter":          "--local-dir ip-adapter",
		"instantid":           "--local-dir instantid",
	}
	for id, needle := range hfLocalDirChecks {
		if !strings.Contains(s[id], needle) {
			t.Errorf("Script %q missing expected --local-dir: %q", id, needle)
		}
	}

	// No script should be empty
	for id, script := range s {
		if len(strings.TrimSpace(script)) < 20 {
			t.Errorf("Script %q is suspiciously short (%d bytes)", id, len(script))
		}
	}

	// No format artifacts
	for id, script := range s {
		if strings.Contains(script, "%!") || strings.Contains(script, "(MISSING)") || strings.Contains(script, "(EXTRA)") {
			t.Errorf("Script %q contains Go format artifact", id)
		}
	}

	// Every script should contain set -e
	for id, script := range s {
		if !strings.Contains(script, "set -e") {
			t.Errorf("Script %q missing 'set -e'", id)
		}
	}
}

func TestGeneratorEdgeCases(t *testing.T) {
	s := GenerateScripts()

	// opensora uses editable install
	if !strings.Contains(s["opensora"], "pip3 install -e .") {
		t.Error("opensora missing editable pip install")
	}

	// rife/film run install.py
	for _, id := range []string{"rife", "film"} {
		if !strings.Contains(s[id], "install.py") {
			t.Errorf("%s missing install.py", id)
		}
		if !strings.Contains(s[id], "requirements.txt") {
			t.Errorf("%s missing requirements.txt install", id)
		}
	}

	// svd/animatediff should NOT run requirements.txt (originals didn't)
	for _, id := range []string{"svd", "animatediff"} {
		if strings.Contains(s[id], "requirements.txt") {
			t.Errorf("%s should not install requirements.txt (not in original)", id)
		}
	}

	// real-esrgan downloads both models
	if !strings.Contains(s["real-esrgan"], "anime_6B") {
		t.Error("real-esrgan missing anime_6B model")
	}

	// rust installs cargo tools
	if !strings.Contains(s["rust"], "cargo install") {
		t.Error("rust missing cargo tool installs")
	}
}
