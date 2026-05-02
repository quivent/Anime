package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
)

// setupOpts configures bootstrap behaviour for `anime wan studio`.
type setupOpts struct {
	checkOnly   bool // print status, install nothing
	skipInstall bool // launch what's there, never install
	yes         bool // skip large-download confirmation prompts
	skipModels  bool // never run wanmodels (it's the ~80GB phase)
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

func wanStudioPhases() []phase {
	home, _ := os.UserHomeDir()
	join := func(parts ...string) string { return filepath.Join(append([]string{home}, parts...)...) }

	fileExists := func(p string) (bool, string) {
		if exists(p) {
			return true, p
		}
		return false, "missing: " + p
	}

	return []phase{
		{
			id:   "comfyui",
			name: "ComfyUI",
			check: func() (bool, string) {
				return fileExists(join("ComfyUI", "main.py"))
			},
			skipMsg: "already installed",
		},
		{
			id:   "wantorch",
			name: "PyTorch cu130 + sage attention",
			check: func() (bool, string) {
				py := join("ComfyUI", "venv", "bin", "python")
				if !exists(py) {
					return false, "ComfyUI venv not built yet"
				}
				out, err := exec.Command(py, "-c",
					"import torch, sageattention; print(torch.version.cuda)").CombinedOutput()
				if err != nil {
					return false, "torch/sageattention not importable"
				}
				cuda := strings.TrimSpace(string(out))
				if cuda != "13.0" {
					return false, "torch cuda=" + cuda + " (want 13.0)"
				}
				return true, "torch cu" + cuda + " + sageattention"
			},
			skipMsg: "torch cu13.0 + sageattention present",
		},
		{
			id:   "wannodes",
			name: "Kijai Wan custom-node stack",
			check: func() (bool, string) {
				return fileExists(join("ComfyUI", "custom_nodes", "ComfyUI-WanVideoWrapper", ".git"))
			},
			skipMsg: "WanVideoWrapper present",
		},
		{
			id:   "wanmodels",
			name: "Wan 2.2 model set (~85GB)",
			// A single-file check is a lie: if the download dies after the
			// first file lands, re-running the studio thinks the phase is
			// satisfied and the next render fails with "node not found" or
			// "missing VAE". Verify the four files the default workflow
			// actually loads (high noise + low noise + text encoder + VAE).
			check: func() (bool, string) {
				required := []struct{ rel, label string }{
					{filepath.Join("models", "diffusion_models", "wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors"), "high-noise 14B"},
					{filepath.Join("models", "diffusion_models", "wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors"), "low-noise 14B"},
					{filepath.Join("models", "text_encoders", "umt5_xxl_fp8_e4m3fn_scaled.safetensors"), "umt5_xxl encoder"},
					{filepath.Join("models", "vae", "wan_2.1_vae.safetensors"), "wan 2.1 VAE"},
				}
				var missing []string
				for _, f := range required {
					if !exists(join("ComfyUI", f.rel)) {
						missing = append(missing, f.label)
					}
				}
				if len(missing) > 0 {
					return false, "missing: " + strings.Join(missing, ", ")
				}
				return true, "high+low 14B + encoder + VAE on disk"
			},
			skipMsg:   "all required model files present",
			heavyGate: true,
			heavyNote: "downloads ~85GB of Wan 2.2 weights from HuggingFace",
		},
		{
			id:   "comfort",
			name: "Comfort studio UI",
			check: func() (bool, string) {
				return fileExists(join("Comfort", "comfort-ui", "dist", "index.html"))
			},
			skipMsg: "dist/index.html present",
		},
		{
			id:   "", // not an installer phase
			name: "ComfyUI server",
			check: func() (bool, string) {
				if comfyServerReachable() {
					return true, "responding at http://127.0.0.1:8188"
				}
				return false, "not running"
			},
			skipMsg: "running",
			custom:  ensureComfyServer,
		},
	}
}

// ensureComfyStudioReady walks each bootstrap phase. Returns nil only when every
// phase is satisfied at the end (so the caller can proceed to serve the studio).
func ensureComfyStudioReady(opts *setupOpts) error {
	phases := wanStudioPhases()
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	fmt.Fprintln(w)
	fmt.Fprintln(w, theme.GlowStyle.Render("🌀 Wan studio · environment check"))
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

// ensureComfyServer starts ComfyUI in a screen session if it isn't already
// reachable, then waits for the HTTP API to come up. First-boot import of
// WanVideoWrapper + KJNodes + sageattention can take 60-90s on a cold venv,
// so we give it 150s and dump the tail of the log on timeout to surface why.
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
	return fmt.Errorf("ComfyUI did not become reachable on :8188 within %ds.\n%s\n%s",
		waitSeconds,
		"  Last 30 lines of the log:",
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

func comfyServerReachable() bool {
	cli := &http.Client{Timeout: 2 * time.Second}
	resp, err := cli.Get("http://127.0.0.1:8188/system_stats")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == 200
}
