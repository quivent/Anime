package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
)

// setupOpts configures bootstrap behaviour for `anime wan studio`.
type setupOpts struct {
	checkOnly    bool   // print status, install nothing
	skipInstall  bool   // launch what's there, never install
	yes          bool   // skip large-download confirmation prompts
	skipModels   bool   // never run wanmodels (it's the ~80GB phase)
	installLevel string // minimal | standard | full — controls wanmodels download size
}

// recommendedLevel maps detected VRAM to the smallest install level that
// still gives the user a working preset for their box. Used to auto-select
// when the user doesn't pass --minimal / --standard / --full.
//
//	>=48GB → full     (everything: T2V dual + I2V dual + 5B + LoRAs ≈ 85GB)
//	>=24GB → standard (T2V dual 14B + 4-step LoRAs ≈ 35GB; covers
//	                   t2v-14b-dual-fast preset on 4090/L40S)
//	 <24GB → minimal  (5B TI2V only ≈ 20GB; covers ti2v-5b preset on 3090)
func recommendedLevel(vramGB int) string {
	switch {
	case vramGB >= 48:
		return "full"
	case vramGB >= 24:
		return "standard"
	default:
		return "minimal"
	}
}

// modelsRequiredForLevel returns the (relativePath, label) pairs that the
// wanmodels phase check must verify for a given install level. Each level
// is a strict superset of the smaller ones: "full" has everything,
// "standard" has the 14B fast pipeline, "minimal" has just the 5B path.
func modelsRequiredForLevel(level string) []struct{ rel, label string } {
	mk := func(pairs ...[2]string) []struct{ rel, label string } {
		out := make([]struct{ rel, label string }, len(pairs))
		for i, p := range pairs {
			out[i] = struct{ rel, label string }{p[0], p[1]}
		}
		return out
	}
	switch level {
	case "minimal":
		return mk(
			[2]string{"models/diffusion_models/wan2.2_ti2v_5B_fp16.safetensors", "5B TI2V"},
			[2]string{"models/text_encoders/umt5_xxl_fp8_e4m3fn_scaled.safetensors", "umt5_xxl encoder"},
			[2]string{"models/vae/wan2.2_vae.safetensors", "wan 2.2 VAE"},
		)
	case "standard":
		return mk(
			[2]string{"models/diffusion_models/wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors", "14B T2V high-noise"},
			[2]string{"models/diffusion_models/wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors", "14B T2V low-noise"},
			[2]string{"models/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_high_noise.safetensors", "4-step high-noise LoRA"},
			[2]string{"models/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_low_noise.safetensors", "4-step low-noise LoRA"},
			[2]string{"models/text_encoders/umt5_xxl_fp8_e4m3fn_scaled.safetensors", "umt5_xxl encoder"},
			[2]string{"models/vae/wan_2.1_vae.safetensors", "wan 2.1 VAE"},
		)
	default: // "full"
		return mk(
			[2]string{"models/diffusion_models/wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors", "14B T2V high-noise"},
			[2]string{"models/diffusion_models/wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors", "14B T2V low-noise"},
			[2]string{"models/text_encoders/umt5_xxl_fp8_e4m3fn_scaled.safetensors", "umt5_xxl encoder"},
			[2]string{"models/vae/wan_2.1_vae.safetensors", "wan 2.1 VAE"},
		)
	}
}

// phase is one bootstrap step.
type phase struct {
	id        string                  // package id in installer.Scripts (or "")
	name      string                  // human label
	check     func() (bool, string)   // returns (satisfied, detail)
	skipMsg   string                  // shown when phase is skipped because already done
	heavyGate bool                    // true → require --yes for unattended install
	heavyNote string                  // shown alongside the gate
	custom    func(*setupOpts) error  // override for non-installer phases (e.g. start ComfyUI)
}

func wanStudioPhases(level string) []phase {
	home, _ := os.UserHomeDir()
	join := func(parts ...string) string { return filepath.Join(append([]string{home}, parts...)...) }

	fileExists := func(p string) (bool, string) {
		if exists(p) {
			return true, p
		}
		return false, "missing: " + p
	}

	// Size + name vary per install level so the user sees what they're
	// actually about to download (and what's already on disk).
	modelsName := "Wan 2.2 model set (full · ~85GB)"
	modelsHeavy := "downloads ~85GB of Wan 2.2 weights from HuggingFace"
	switch level {
	case "minimal":
		modelsName = "Wan 2.2 model set (minimal · ~20GB)"
		modelsHeavy = "downloads ~20GB (5B TI2V + encoder + VAE)"
	case "standard":
		modelsName = "Wan 2.2 model set (standard · ~35GB)"
		modelsHeavy = "downloads ~35GB (14B T2V dual + 4-step LoRAs + encoder + VAE)"
	}

	return []phase{
		{
			id:   "comfyui",
			name: "Render engine (ComfyUI)",
			check: func() (bool, string) {
				return fileExists(join("ComfyUI", "main.py"))
			},
			skipMsg: "installed",
		},
		{
			id:   "wantorch",
			name: "PyTorch + sage attention",
			// Accept any CUDA torch that imports cleanly with sageattention
			// available and torch.cuda.is_available() == True. The wantorch
			// install script picks the wheel index from the host's driver
			// (cu130 / cu128 / cu124 / cu121 / cu118), so we pin behavior on
			// "is it actually working" rather than a specific cu version.
			check: func() (bool, string) {
				py := join("ComfyUI", "venv", "bin", "python")
				if !exists(py) {
					return false, "ComfyUI venv not built yet"
				}
				// Probe in Python; print a single OK/FAIL line so the wrapper
				// can show it inline without smearing a traceback.
				probe := `try:
    import torch, sageattention
    if not torch.cuda.is_available():
        print("FAIL: torch.cuda.is_available() == False")
    else:
        print("OK " + torch.version.cuda)
except ImportError as e:
    print("FAIL: " + str(e))`
				out, _ := exec.Command(py, "-c", probe).CombinedOutput()
				line := strings.TrimSpace(string(out))
				// Take just the last non-empty line in case Python printed
				// any warnings before our marker.
				if lines := strings.Split(line, "\n"); len(lines) > 0 {
					line = strings.TrimSpace(lines[len(lines)-1])
				}
				if strings.HasPrefix(line, "OK ") {
					return true, "torch cu" + strings.TrimPrefix(line, "OK ") + " + sageattention"
				}
				if strings.HasPrefix(line, "FAIL:") {
					return false, strings.TrimSpace(strings.TrimPrefix(line, "FAIL:"))
				}
				return false, "torch/sageattention probe produced no output"
			},
			skipMsg: "working torch + sageattention",
		},
		{
			id:   "wannodes",
			name: "Wan custom-node stack",
			check: func() (bool, string) {
				return fileExists(join("ComfyUI", "custom_nodes", "ComfyUI-WanVideoWrapper", ".git"))
			},
			skipMsg: "Kijai WanVideoWrapper present",
		},
		{
			id:   "wanmodels",
			name: modelsName,
			// Required files vary by install level — check exactly the set
			// the user asked for. A single-file check is a lie because if
			// the download dies after the first file lands, re-running the
			// studio would think the phase is satisfied and the next render
			// would fail with "missing VAE".
			check: func() (bool, string) {
				required := modelsRequiredForLevel(level)
				var missing []string
				for _, f := range required {
					if !exists(join("ComfyUI", f.rel)) {
						missing = append(missing, f.label)
					}
				}
				if len(missing) > 0 {
					return false, "missing: " + strings.Join(missing, ", ")
				}
				return true, fmt.Sprintf("%d files for level=%s on disk", len(required), level)
			},
			skipMsg:   "all required model files present",
			heavyGate: true,
			heavyNote: modelsHeavy,
		},
		{
			id:   "comfort",
			name: "Comfort studio (web UI)",
			check: func() (bool, string) {
				return fileExists(join("Comfort", "comfort-ui", "dist", "index.html"))
			},
			skipMsg: "dist/index.html present",
		},
		{
			id:   "", // not an installer phase
			name: "Render engine running",
			check: func() (bool, string) {
				if comfyServerReachable() {
					return true, "responding at http://127.0.0.1:8188"
				}
				return false, "not running"
			},
			skipMsg: "responding",
			custom:  ensureComfyServer,
		},
	}
}

// ensureComfyStudioReady walks each bootstrap phase. Returns nil only when every
// phase is satisfied at the end (so the caller can proceed to serve the studio).
func ensureComfyStudioReady(opts *setupOpts) error {
	// If the user didn't pick a level explicitly, auto-pick from VRAM. This
	// means a fresh `anime wan studio --yes` on a 24GB box pulls 35GB
	// instead of 85GB by default.
	if opts.installLevel == "" {
		opts.installLevel = recommendedLevel(gpu.GetTotalVRAM())
	}
	// Propagate to the wanmodels install script (it reads WAN_INSTALL_LEVEL).
	os.Setenv("WAN_INSTALL_LEVEL", opts.installLevel)

	phases := wanStudioPhases(opts.installLevel)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	fmt.Fprintln(w)
	fmt.Fprintln(w, theme.GlowStyle.Render("🌀 Wan studio · environment check"))
	fmt.Fprintln(w)
	printHostSummary(w, opts.installLevel)
	fmt.Fprintln(w)

	for _, ph := range phases {
		ok, detail := ph.check()
		label := theme.HighlightStyle.Render(fmt.Sprintf("%-32s", ph.name))
		if ok {
			fmt.Fprintf(w, "  %s %s  %s\n", theme.SymbolSuccess, label,
				theme.DimTextStyle.Render(ph.skipMsg+" — "+detail))
			continue
		}
		fmt.Fprintf(w, "  %s %s  %s\n", theme.SymbolWarning, label, theme.WarningStyle.Render(detail))

		if opts.checkOnly {
			continue
		}
		if opts.skipInstall {
			return fmt.Errorf("phase %q not satisfied and --skip-install was given", ph.name)
		}

		if ph.id == "wanmodels" && opts.skipModels {
			fmt.Fprintln(w, "    "+theme.DimTextStyle.Render("(skipped — --skip-models)"))
			continue
		}

		if ph.heavyGate && !opts.yes {
			w.Flush()
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("  This phase " + ph.heavyNote + "."))
			fmt.Print(theme.HighlightStyle.Render("  Continue? [y/N] "))
			var ans string
			fmt.Scanln(&ans)
			if !strings.EqualFold(strings.TrimSpace(ans), "y") &&
				!strings.EqualFold(strings.TrimSpace(ans), "yes") {
				return fmt.Errorf("aborted at phase %q (re-run with --yes to skip the prompt)", ph.name)
			}
		}

		fmt.Fprintln(w)
		fmt.Fprintln(w, "    "+theme.InfoStyle.Render("→ installing "+ph.name))
		fmt.Fprintln(w)
		w.Flush()

		var err error
		if ph.custom != nil {
			err = ph.custom(opts)
		} else if ph.id != "" {
			err = runInstallScript(ph.id)
		}
		if err != nil {
			return fmt.Errorf("phase %q failed: %w", ph.name, err)
		}

		// Re-check after install — fail loudly if it didn't work, since the
		// next phase might silently depend on this one.
		if ok, detail := ph.check(); !ok {
			return fmt.Errorf("phase %q completed but check still fails: %s", ph.name, detail)
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s %s  %s\n", theme.SymbolSuccess,
			theme.HighlightStyle.Render(fmt.Sprintf("%-32s", ph.name)),
			theme.SuccessStyle.Render("ready"))
	}

	if opts.checkOnly {
		// Re-tally so the user sees a clean summary at the end.
		fmt.Fprintln(w)
		fmt.Fprintln(w, theme.DimTextStyle.Render("  (--check-only: not installing)"))
	}
	return nil
}

// runInstallScript fetches the bash script for a package id and runs it locally,
// streaming stdout/stderr. The caller is responsible for printing a heading.
func runInstallScript(id string) error {
	script, ok := installer.GetScript(id)
	if !ok {
		return fmt.Errorf("no install script registered for %q", id)
	}
	c := exec.Command("bash", "-c", script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	return c.Run()
}

// ensureComfyServer starts the render engine (ComfyUI) in a screen session if
// it isn't already reachable, then waits for the HTTP API to come up. First-
// boot import of WanVideoWrapper + KJNodes + sageattention can take 60-90s
// on a cold venv, so we give it 150s and dump the tail of the log on timeout
// to surface why.
func ensureComfyServer(opts *setupOpts) error {
	if comfyServerReachable() {
		return nil
	}
	if err := startComfyUIServer(); err != nil {
		return err
	}
	const waitSeconds = 150
	start := time.Now()
	lastTick := time.Now()
	for time.Since(start) < waitSeconds*time.Second {
		time.Sleep(1 * time.Second)
		if comfyServerReachable() {
			return nil
		}
		// Light, dot-only progress every 5s so the user knows we're waiting,
		// without flooding stdout.
		if time.Since(lastTick) > 5*time.Second {
			fmt.Print(".")
			lastTick = time.Now()
		}
	}
	fmt.Println()
	return fmt.Errorf("render engine did not become reachable on :8188 within %ds.\n%s\n%s",
		waitSeconds,
		"  Last 30 lines of the engine log:",
		tailComfyLog(30))
}

// tailComfyLog returns the last `n` lines of ~/.anime/comfyui.log with each
// line indented, so we can dump it directly inside an error string.
func tailComfyLog(n int) string {
	home, _ := os.UserHomeDir()
	logFile := filepath.Join(home, ".anime", "comfyui.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "    (no log at " + logFile + " yet — attach to the live session: screen -r comfyui)"
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	for i, l := range lines {
		lines[i] = "    " + l
	}
	return strings.Join(lines, "\n")
}

// recommendedPreset returns the Wan preset that best fits the host's VRAM:
//
//	>=48GB → t2v-14b-dual-maxq (4090 dual / H100 / GH200 — full quality)
//	>=24GB → t2v-14b-dual-fast (4090 / A100 40 / L40S — 4-step LoRA)
//	>=12GB → ti2v-5b           (3090 / 4080 / A40 24 — 5B model)
//	 <12GB → "" (unsupported)
//
// Used by the studio --check output so users know which preset to pick.
func recommendedPreset(vramGB int) (preset, why string) {
	switch {
	case vramGB >= 48:
		return "t2v-14b-dual-maxq", "≥48GB VRAM → full quality 14B at 1280x720"
	case vramGB >= 24:
		return "t2v-14b-dual-fast", "≥24GB VRAM → 14B + 4-step LoRA at 832x480 (default)"
	case vramGB >= 12:
		return "ti2v-5b", "≥12GB VRAM → 5B TI2V at 832x480"
	case vramGB > 0:
		return "", fmt.Sprintf("only %dGB VRAM detected — Wan 2.2 needs ≥12GB", vramGB)
	default:
		return "t2v-14b-dual-fast", "no GPU detected on this host (default preset assumes ≥24GB on the render box)"
	}
}

// printHostSummary prints a one-block GPU/VRAM/driver/disk preamble before the
// phase checks. It's the first thing the user sees on `--check`, so they
// immediately know what their box is and which preset/install level fits.
func printHostSummary(w *bufio.Writer, level string) {
	g := gpu.GetSystemInfo()
	row := func(label, val, hint string) {
		fmt.Fprintf(w, "  %s %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-12s", label)),
			val,
			theme.DimTextStyle.Render(hint))
	}

	if !g.Available {
		row("GPU", theme.WarningStyle.Render("none detected"),
			"nvidia-smi not found — install nvidia drivers (anime install nvidia)")
	} else {
		gpuLine := fmt.Sprintf("%dx %s", g.Count, g.GPUs[0].Name)
		if g.Count == 1 {
			gpuLine = g.GPUs[0].Name
		}
		row("GPU", theme.SuccessStyle.Render(gpuLine), "")
		row("VRAM", theme.PrimaryTextStyle.Render(fmt.Sprintf("%d GB", g.TotalVRAM)),
			fmt.Sprintf("%d MiB total across %d GPU(s)", g.GPUs[0].VRAMMiB*g.Count, g.Count))
		drvLine := g.DriverVersion
		if g.CUDAVersion != "" {
			drvLine += "  (CUDA " + g.CUDAVersion + ")"
		}
		row("Driver", theme.PrimaryTextStyle.Render(drvLine),
			"determines which torch wheel index wantorch picks")
	}

	row("Arch", runtime.GOARCH, runtime.GOOS+" — wheels match this architecture")

	// Free disk on $HOME (where ~/ComfyUI/models lives).
	home, _ := os.UserHomeDir()
	if home != "" {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(home, &stat); err == nil {
			freeGB := int(stat.Bavail) * int(stat.Bsize) / (1 << 30)
			diskHint := home
			diskStyle := theme.PrimaryTextStyle
			if freeGB < 100 {
				diskStyle = theme.WarningStyle
				diskHint = home + "  (Wan 2.2 model set is ~85GB)"
			}
			row("Free disk", diskStyle.Render(fmt.Sprintf("%d GB", freeGB)), diskHint)
		}
	}

	// Recommended preset based on detected VRAM.
	preset, why := recommendedPreset(g.TotalVRAM)
	if preset == "" {
		row("Preset", theme.WarningStyle.Render("(none fits this VRAM)"), why)
	} else {
		row("Preset", theme.SuccessStyle.Render(preset), why)
	}

	// Selected install level (set by --minimal/--standard/--full or auto).
	levelHint := ""
	switch level {
	case "minimal":
		levelHint = "5B model only · ~20GB · ~5min"
	case "standard":
		levelHint = "14B T2V dual + 4-step LoRAs · ~35GB · ~10min"
	case "full":
		levelHint = "everything (T2V+I2V dual + 5B + LoRAs) · ~85GB · ~22min"
	}
	row("Install", theme.SuccessStyle.Render(level), levelHint)

	// HF auth status — important because future Wan repos may go gated.
	if os.Getenv("HF_TOKEN") != "" {
		row("HF auth", theme.SuccessStyle.Render("HF_TOKEN set"), "model downloads will use this token")
	} else if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".cache", "huggingface", "token")); err == nil {
		row("HF auth", theme.SuccessStyle.Render("huggingface-cli login cached"), "")
	} else {
		row("HF auth", theme.DimTextStyle.Render("anonymous"), "ok for public repos · run: huggingface-cli login")
	}
}

func comfyServerReachable() bool {
	cli := &http.Client{Timeout: 2 * time.Second}
	resp, err := cli.Get("http://127.0.0.1:8188/system_stats")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == 200
}
