//go:build ignore

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
	default: // "full" — strict superset of standard + minimal
		return mk(
			// T2V 14B dual-expert
			[2]string{"models/diffusion_models/wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors", "14B T2V high-noise"},
			[2]string{"models/diffusion_models/wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors", "14B T2V low-noise"},
			// I2V 14B dual-expert
			[2]string{"models/diffusion_models/wan2.2_i2v_high_noise_14B_fp8_scaled.safetensors", "14B I2V high-noise"},
			[2]string{"models/diffusion_models/wan2.2_i2v_low_noise_14B_fp8_scaled.safetensors", "14B I2V low-noise"},
			// 5B TI2V
			[2]string{"models/diffusion_models/wan2.2_ti2v_5B_fp16.safetensors", "5B TI2V"},
			// 4-step LoRAs
			[2]string{"models/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_high_noise.safetensors", "4-step high-noise LoRA"},
			[2]string{"models/loras/wan2.2_t2v_lightx2v_4steps_lora_v1.1_low_noise.safetensors", "4-step low-noise LoRA"},
			// Shared
			[2]string{"models/text_encoders/umt5_xxl_fp8_e4m3fn_scaled.safetensors", "umt5_xxl encoder"},
			[2]string{"models/vae/wan_2.1_vae.safetensors", "wan 2.1 VAE"},
			[2]string{"models/vae/wan2.2_vae.safetensors", "wan 2.2 VAE"},
		)
	}
}

// phase is one bootstrap step.
type phase struct {
	id        string                 // package id in installer.Scripts (or "")
	name      string                 // human label
	check     func() (bool, string)  // returns (satisfied, detail)
	skipMsg   string                 // shown when phase is skipped because already done
	heavyGate bool                   // true -> require --yes for unattended install
	heavyNote string                 // shown alongside the gate
	custom    func(*setupOpts) error // override for non-installer phases (e.g. start ComfyUI)
	estimate  time.Duration          // estimated wall-clock time for this phase
}

// wanModelFiles enumerates every file the Wan 2.2 workflows need.
// Used by both the check function and the post-download summary.
var wanModelFiles = []struct{ rel, label string }{
	{filepath.Join("models", "diffusion_models", "wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors"), "t2v high-noise 14B"},
	{filepath.Join("models", "diffusion_models", "wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors"), "t2v low-noise 14B"},
	{filepath.Join("models", "diffusion_models", "wan2.2_i2v_high_noise_14B_fp8_scaled.safetensors"), "i2v high-noise 14B"},
	{filepath.Join("models", "diffusion_models", "wan2.2_i2v_low_noise_14B_fp8_scaled.safetensors"), "i2v low-noise 14B"},
	{filepath.Join("models", "diffusion_models", "wan2.2_ti2v_5B_fp16.safetensors"), "ti2v 5B"},
	{filepath.Join("models", "text_encoders", "umt5_xxl_fp8_e4m3fn_scaled.safetensors"), "umt5_xxl encoder"},
	{filepath.Join("models", "vae", "wan_2.1_vae.safetensors"), "wan 2.1 VAE"},
	{filepath.Join("models", "vae", "wan2.2_vae.safetensors"), "wan 2.2 VAE"},
	{filepath.Join("models", "loras", "wan2.2_t2v_lightx2v_4steps_lora_v1.1_high_noise.safetensors"), "lightx2v high-noise LoRA"},
	{filepath.Join("models", "loras", "wan2.2_t2v_lightx2v_4steps_lora_v1.1_low_noise.safetensors"), "lightx2v low-noise LoRA"},
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
			skipMsg:  "already installed",
			estimate: 15 * time.Minute,
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
			skipMsg:  "torch cu13.0 + sageattention present",
			estimate: 5 * time.Minute,
		},
		{
			id:   "wannodes",
			name: "Wan custom-node stack",
			check: func() (bool, string) {
				return fileExists(join("ComfyUI", "custom_nodes", "ComfyUI-WanVideoWrapper", ".git"))
			},
			skipMsg:  "WanVideoWrapper present",
			estimate: 3 * time.Minute,
		},
		{
			id:   "wanmodels",
			name: "Wan 2.2 model set (~85GB)",
			// Verify ALL files the Wan 2.2 workflows need -- not just a
			// subset. A partial download (killed after the first few files)
			// used to pass the old 4-file check, then the next render would
			// fail with "missing VAE" or "node not found". Each file must
			// exist AND have size > 0 (a zero-byte file means the download
			// was interrupted mid-write).
			check: func() (bool, string) {
				var missing []string
				for _, f := range wanModelFiles {
					p := join("ComfyUI", f.rel)
					info, err := os.Stat(p)
					if err != nil || info.Size() == 0 {
						missing = append(missing, f.label)
					}
				}
				if len(missing) > 0 {
					return false, fmt.Sprintf("missing %d/%d: %s", len(missing), len(wanModelFiles), strings.Join(missing, ", "))
				}
				return true, fmt.Sprintf("all %d model files present and non-empty", len(wanModelFiles))
			},
			skipMsg:   "all required model files present",
			heavyGate: true,
			heavyNote: "downloads ~85GB of Wan 2.2 weights from HuggingFace",
			estimate:  45 * time.Minute,
		},
		{
			id:   "comfort",
			name: "Comfort studio (web UI)",
			check: func() (bool, string) {
				return fileExists(join("Comfort", "comfort-ui", "dist", "index.html"))
			},
			skipMsg:  "dist/index.html present",
			estimate: 8 * time.Minute,
		},
		{
			id:   "domain",
			name: "Domain + SSL (comfort.producer.cafe)",
			check: func() (bool, string) {
				// The domain install script writes to sites-available/comfort
				// and symlinks to sites-enabled/comfort
				if exists("/etc/nginx/sites-enabled/comfort") {
					return true, "https://comfort.producer.cafe"
				}
				return false, "not configured"
			},
			skipMsg: "https://comfort.producer.cafe",
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
			skipMsg:  "running",
			custom:   ensureComfyServer,
			estimate: 2 * time.Minute,
		},
	}
}

// countWanModelFilesPresent returns how many of the expected model files
// exist and are non-empty. Used for post-download summary.
func countWanModelFilesPresent() (present int, totalBytes int64) {
	home, _ := os.UserHomeDir()
	for _, f := range wanModelFiles {
		p := filepath.Join(home, "ComfyUI", f.rel)
		info, err := os.Stat(p)
		if err == nil && info.Size() > 0 {
			present++
			totalBytes += info.Size()
		}
	}
	return
}

// durationStr converts a time.Duration to the compact "Xm Ys" string
// using the shared formatDuration(seconds int) from common.go.
func durationStr(d time.Duration) string {
	return formatDuration(int(d.Round(time.Second).Seconds()))
}

// ensureComfyStudioReady walks each bootstrap phase. Returns nil only when every
// phase is satisfied at the end (so the caller can proceed to serve the studio).
func ensureComfyStudioReady(opts *setupOpts) error {
	phases := wanStudioPhases()
	total := len(phases)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// -- First pass: probe every phase to detect first-run vs resume vs ready --
	type probeResult struct {
		ok     bool
		detail string
	}
	probes := make([]probeResult, total)
	allSatisfied := true
	firstUnsatisfied := -1
	satisfiedCount := 0
	for i, ph := range phases {
		ok, detail := ph.check()
		probes[i] = probeResult{ok, detail}
		if ok {
			satisfiedCount++
		} else {
			allSatisfied = false
			if firstUnsatisfied < 0 {
				firstUnsatisfied = i
			}
		}
	}

	// -- Determine run mode --
	isFirstRun := satisfiedCount == 0
	isResuming := !allSatisfied && satisfiedCount > 0 && !opts.checkOnly
	bootstrapStart := time.Now()

	fmt.Fprintln(w)
	if allSatisfied {
		// Subsequent run: everything is ready. Terse output.
		fmt.Fprintln(w, theme.GlowStyle.Render("Wan studio - ready"))
		fmt.Fprintln(w)
		for i, ph := range phases {
			prefix := theme.DimTextStyle.Render(fmt.Sprintf("[%d/%d]", i+1, total))
			label := theme.HighlightStyle.Render(fmt.Sprintf("%-32s", ph.name))
			fmt.Fprintf(w, "  %s %s %s\n", prefix, label,
				theme.DimTextStyle.Render("ok"))
		}
	} else if isFirstRun {
		// First run: educational header.
		fmt.Fprintln(w, theme.GlowStyle.Render("Wan studio - first-time setup"))
		fmt.Fprintln(w)
		fmt.Fprintln(w, theme.DimTextStyle.Render("  This will install the full Wan 2.2 video generation stack:"))
		fmt.Fprintln(w, theme.DimTextStyle.Render("  ComfyUI, PyTorch cu130, Kijai custom nodes, ~85GB of models,"))
		fmt.Fprintln(w, theme.DimTextStyle.Render("  and the Comfort studio UI. Total estimated time: ~76 minutes."))
		fmt.Fprintln(w)
	} else if isResuming {
		// Resume: tell the user exactly where we pick up.
		fmt.Fprintln(w, theme.GlowStyle.Render("Wan studio - resuming setup"))
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("Resuming from phase %d (%s)...",
				firstUnsatisfied+1, phases[firstUnsatisfied].name)))
		fmt.Fprintf(w, "  %s\n",
			theme.DimTextStyle.Render(fmt.Sprintf("Phases 1-%d already complete, skipping ahead.",
				firstUnsatisfied)))
		fmt.Fprintln(w)
	} else {
		// check-only mode with missing phases
		fmt.Fprintln(w, theme.GlowStyle.Render("Wan studio - environment check"))
		fmt.Fprintln(w)
	}

	if allSatisfied {
		// Everything is already installed. Skip the per-phase walk entirely
		// and jump to post-bootstrap.
		goto postBootstrap
	}

	// -- Per-phase walk --
	for i, ph := range phases {
		phaseNum := i + 1
		prefix := fmt.Sprintf("[%d/%d]", phaseNum, total)
		label := theme.HighlightStyle.Render(fmt.Sprintf("%-32s", ph.name))

		ok := probes[i].ok
		detail := probes[i].detail

		if ok {
			// Phase already satisfied -- show it, don't silently skip.
			fmt.Fprintf(w, "  %s %s %s  %s\n",
				theme.DimTextStyle.Render(prefix),
				label,
				theme.SuccessStyle.Render("(already installed)"),
				theme.DimTextStyle.Render(ph.skipMsg))
			continue
		}

		// Phase not satisfied.
		fmt.Fprintf(w, "  %s %s %s\n",
			theme.DimTextStyle.Render(prefix),
			label,
			theme.WarningStyle.Render(detail))

		if opts.checkOnly {
			continue
		}
		if opts.skipInstall {
			return fmt.Errorf("phase %q not satisfied and --skip-install was given", ph.name)
		}

		if ph.id == "wanmodels" && opts.skipModels {
			fmt.Fprintf(w, "         %s\n",
				theme.DimTextStyle.Render("(skipped -- --skip-models)"))
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

		// Pre-install messaging for wanmodels: tell the user what they're in for.
		if ph.id == "wanmodels" {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "         %s\n",
				theme.InfoStyle.Render("Downloading Wan 2.2 models (~85GB)."))
			fmt.Fprintf(w, "         %s\n",
				theme.DimTextStyle.Render("This will take 30-60 minutes depending on bandwidth."))
			fmt.Fprintf(w, "         %s\n",
				theme.DimTextStyle.Render("10 files: 5 diffusion models, 1 text encoder, 2 VAEs, 2 LoRAs."))
			fmt.Fprintln(w)
		}

		estStr := ""
		if ph.estimate > 0 {
			estStr = " (est. " + durationStr(ph.estimate) + ")"
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "         %s\n",
			theme.InfoStyle.Render("Installing "+ph.name+estStr+"..."))
		fmt.Fprintln(w)
		w.Flush()

		phaseStart := time.Now()
		var err error
		if ph.custom != nil {
			err = ph.custom(opts)
		} else if ph.id != "" {
			err = runInstallScript(ph.id)
		}
		phaseElapsed := time.Since(phaseStart)

		if err != nil {
			// FAIL
			fmt.Fprintln(w)
			fmt.Fprintf(w, "  %s %s %s  %s\n",
				theme.DimTextStyle.Render(prefix),
				label,
				theme.WarningStyle.Render("FAIL"),
				theme.DimTextStyle.Render(durationStr(phaseElapsed)))
			totalElapsed := time.Since(bootstrapStart)
			fmt.Fprintf(w, "\n  %s\n",
				theme.DimTextStyle.Render(fmt.Sprintf("Total elapsed: %s", durationStr(totalElapsed))))
			return fmt.Errorf("phase %q failed: %w", ph.name, err)
		}

		// Re-check after install -- fail loudly if it didn't work, since the
		// next phase might silently depend on this one.
		if ok, detail := ph.check(); !ok {
			return fmt.Errorf("phase %q completed but check still fails: %s", ph.name, detail)
		}

		// Post-install messaging for wanmodels: summarize what we got.
		if ph.id == "wanmodels" {
			present, totalBytes := countWanModelFilesPresent()
			fmt.Fprintln(w)
			fmt.Fprintf(w, "         %s\n",
				theme.SuccessStyle.Render(fmt.Sprintf("Downloaded %d/%d model files (%dGB total).",
					present, len(wanModelFiles), totalBytes/(1024*1024*1024))))
		}

		// DONE
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s %s %s  %s\n",
			theme.DimTextStyle.Render(prefix),
			label,
			theme.SuccessStyle.Render("DONE"),
			theme.DimTextStyle.Render(durationStr(phaseElapsed)))
	}

	// Total elapsed for the entire bootstrap.
	if !opts.checkOnly {
		totalElapsed := time.Since(bootstrapStart)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s\n",
			theme.DimTextStyle.Render(fmt.Sprintf("Total elapsed: %s", durationStr(totalElapsed))))
	}

	if opts.checkOnly {
		// Re-tally so the user sees a clean summary at the end.
		fmt.Fprintln(w)
		fmt.Fprintln(w, theme.DimTextStyle.Render("  (--check-only: not installing)"))
	}

postBootstrap:

	// Post-bootstrap: deliver the default workflow JSON if missing.
	ensureDefaultWorkflow(w)

	// Post-bootstrap: warn (don't fail) if ComfyUI isn't reachable yet.
	// The "ComfyUI server" phase above starts it, but a --check-only or
	// --skip-install run may legitimately leave it down.
	comfyURL := "http://127.0.0.1:8188"
	if !comfyServerReachable() {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s  %s\n", theme.SymbolWarning,
			theme.WarningStyle.Render("ComfyUI is not reachable at "+comfyURL+"/system_stats"))
		fmt.Fprintf(w, "       %s\n",
			theme.DimTextStyle.Render("Start it with: screen -dmS comfyui bash -c 'cd ~/ComfyUI && ./venv/bin/python main.py --listen --use-sage-attention'"))
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
		return "    (no log at " + logFile + " yet -- attach to the live session: screen -r comfyui)"
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

	// Tuning preview — show the env + flags we'll pass to ComfyUI when it
	// starts. Computing this here is cheap (just inspects gpu info + venv
	// glob) and helps users diagnose perf surprises before they hit them.
	tuning := AutoTuneComfyUI(g, detectSageInstalled())
	flagSummary := "(default)"
	if len(tuning.Flags) > 0 {
		flagSummary = strings.Join(tuning.Flags, " ")
	}
	row("Tuning", theme.SuccessStyle.Render(flagSummary), "comfyui flags · WAN_COMFY_FLAGS to override")
	for _, n := range tuning.Notes {
		fmt.Fprintf(w, "  %s %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "")),
			theme.DimTextStyle.Render("· "+n))
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

// ensureDefaultWorkflow copies the embedded workflow JSON into the ComfyUI
// workflows directory if it doesn't already exist there. This is a best-effort
// convenience -- failure only prints a note, never blocks bootstrap.
func ensureDefaultWorkflow(w *bufio.Writer) {
	home, _ := os.UserHomeDir()
	const wfName = "Wan2.2_14B_T2V_NoLoRA_MaxQuality.json"
	dest := filepath.Join(home, "ComfyUI", "user", "default", "workflows", wfName)

	if exists(dest) {
		return // already in place
	}

	// Try to read from the embedded workflows FS (compiled into the binary).
	data, err := embeddedWorkflows.ReadFile("embedded/workflows/" + wfName)
	if err != nil {
		// Embedded asset not available (e.g., dev build without the file).
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s  %s\n", theme.SymbolWarning,
			theme.WarningStyle.Render("Default workflow not found: "+wfName))
		fmt.Fprintf(w, "       %s\n",
			theme.DimTextStyle.Render("Place it at: "+dest))
		fmt.Fprintf(w, "       %s\n",
			theme.DimTextStyle.Render("Or copy from the anime repo: cli/cmd/embedded/workflows/"+wfName))
		return
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s  %s\n", theme.SymbolWarning,
			theme.WarningStyle.Render("Could not create workflow directory: "+err.Error()))
		return
	}

	if err := os.WriteFile(dest, data, 0o644); err != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s  %s\n", theme.SymbolWarning,
			theme.WarningStyle.Render("Failed to write workflow: "+err.Error()))
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s  %s  %s\n", theme.SymbolSuccess,
		theme.HighlightStyle.Render(fmt.Sprintf("%-32s", "Default workflow")),
		theme.DimTextStyle.Render("delivered -> "+dest))
}
