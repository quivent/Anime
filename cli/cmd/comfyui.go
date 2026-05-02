package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/gpu"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var comfyuiCmd = &cobra.Command{
	Use:   "comfyui <start|stop|status|logs>",
	Short: "Manage the ComfyUI render engine (used by anime wan studio)",
	Long: `Manage the ComfyUI render engine — the Python service on :8188 that
actually executes Wan 2.2 workflows. The Comfort web studio (anime wan studio)
talks to this engine over /api and /ws; you usually don't run anime comfyui
directly unless you're debugging.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("Action required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime comfyui <action>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Available Actions:"))
			fmt.Println(theme.DimTextStyle.Render("  start   - Start render engine in background"))
			fmt.Println(theme.DimTextStyle.Render("  stop    - Stop render engine"))
			fmt.Println(theme.DimTextStyle.Render("  status  - Show engine status and health"))
			fmt.Println(theme.DimTextStyle.Render("  logs    - View engine logs"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Examples:"))
			fmt.Println(theme.HighlightStyle.Render("  anime comfyui start"))
			fmt.Println(theme.HighlightStyle.Render("  anime comfyui status"))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("Most users want: anime wan studio  (the Comfort web UI)"))
			fmt.Println()
			return fmt.Errorf("comfyui requires an action (start|stop|status|logs)")
		}
		return nil
	},
	RunE: runComfyUICommand,
}

func init() {
	rootCmd.AddCommand(comfyuiCmd)
}

func runComfyUICommand(cmd *cobra.Command, args []string) error {
	action := args[0]

	switch action {
	case "start":
		return startComfyUIServer()
	case "stop":
		return stopComfyUIServer()
	case "status":
		return statusComfyUIServer()
	case "logs":
		return logsComfyUIServer()
	default:
		return fmt.Errorf("unknown action: %s (use: start|stop|status|logs)", action)
	}
}

func startComfyUIServer() error {
	fmt.Println(theme.InfoStyle.Render("🚀 Starting render engine (ComfyUI) in background..."))

	// Pick the venv python (where torch cu130 + sageattention live) when it
	// exists; fall back to system python3 only if the venv is missing — that
	// case will probably crash on `import torch`, but at least we don't
	// silently mask the user's broken install behind a system python that
	// happens to have a different (wrong) torch.
	home, _ := os.UserHomeDir()
	venvPy := filepath.Join(home, "ComfyUI", "venv", "bin", "python")
	pyCmd := "python3"
	sageInstalled := false
	if _, err := os.Stat(venvPy); err == nil {
		pyCmd = "./venv/bin/python"
		sageInstalled = detectSageInstalled()
	}
	// AutoTune adds the right env + flags for this host (sage backend if
	// installed, --reserve-vram on big iron, --lowvram on tight boxes,
	// PYTORCH_CUDA_ALLOC_CONF / TF32 / lazy CUDA module loading).
	tuning := AutoTuneComfyUI(gpu.GetSystemInfo(), sageInstalled)

	// Ensure ~/.anime exists before tee writes into it. tee creates the file
	// but not the parent dir, so a fresh box that hits `anime comfyui start`
	// before any other CLI command would otherwise fail the log pipe.
	animeDir := filepath.Join(home, ".anime")
	_ = os.MkdirAll(animeDir, 0o755)
	logPath := filepath.Join(animeDir, "comfyui.log")

	// Build a tiny launch script: env vars exported, then exec the python.
	// We export inside the screen-launched bash so the env actually applies
	// to ComfyUI rather than just our orchestrator.
	var sb strings.Builder
	sb.WriteString("cd ~/ComfyUI && ")
	for _, kv := range tuning.EnvLines() {
		sb.WriteString("export ")
		sb.WriteString(kv)
		sb.WriteString(" && ")
	}
	sb.WriteString("exec ")
	sb.WriteString(pyCmd)
	sb.WriteString(" main.py --listen")
	for _, f := range tuning.Flags {
		sb.WriteString(" ")
		sb.WriteString(f)
	}
	sb.WriteString(" 2>&1 | tee -a ")
	sb.WriteString(logPath)
	launch := sb.String()

	// Show the user what we just decided. The studio's --check view dumps
	// the same info — this is the inline "I'm starting it now" version.
	fmt.Println(theme.DimTextStyle.Render("  Tuning:"))
	for _, n := range tuning.Notes {
		fmt.Println(theme.DimTextStyle.Render("    · " + n))
	}

	cmd := exec.Command("screen", "-dmS", "comfyui", "bash", "-c", launch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start render engine: %w", err)
	}

	// Wait a moment and check if it started
	fmt.Print(theme.DimTextStyle.Render("  Waiting for startup"))
	for i := 0; i < 5; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
	fmt.Println()

	publicIP := getPublicIPForComfyUI()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Render engine reachable at http://%s:8188", publicIP)))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("View logs:    anime comfyui logs"))
	fmt.Println(theme.DimTextStyle.Render("Check status: anime comfyui status"))
	fmt.Println(theme.DimTextStyle.Render("Stop engine:  anime comfyui stop"))
	fmt.Println()

	return nil
}

func stopComfyUIServer() error {
	fmt.Println(theme.InfoStyle.Render("🛑 Stopping render engine..."))

	cmd := exec.Command("screen", "-S", "comfyui", "-X", "quit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop render engine: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Render engine stopped"))
	return nil
}

func statusComfyUIServer() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 RENDER ENGINE STATUS"))
	fmt.Println()

	// 1. Check if ComfyUI directory exists
	homeDir, _ := os.UserHomeDir()
	comfyPath := filepath.Join(homeDir, "ComfyUI")

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Installation"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if _, err := os.Stat(comfyPath); os.IsNotExist(err) {
		fmt.Printf("  Path: %s\n", theme.WarningStyle.Render(comfyPath))
		fmt.Printf("  Status: %s\n", theme.WarningStyle.Render("❌ Not installed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Install with: anime packages install comfyui"))
		return nil
	}

	fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(comfyPath))
	fmt.Printf("  Status: %s\n", theme.SuccessStyle.Render("✓ Installed"))

	// Check for main.py
	mainPy := filepath.Join(comfyPath, "main.py")
	if _, err := os.Stat(mainPy); os.IsNotExist(err) {
		fmt.Printf("  Main file: %s\n", theme.WarningStyle.Render("❌ main.py not found"))
	} else {
		fmt.Printf("  Main file: %s\n", theme.SuccessStyle.Render("✓ main.py found"))
	}
	fmt.Println()

	// 2. Check if screen session exists
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🖥️  Process Status"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	screenCmd := exec.Command("screen", "-ls")
	screenOutput, _ := screenCmd.CombinedOutput()
	hasScreen := strings.Contains(string(screenOutput), "comfyui")

	if hasScreen {
		fmt.Printf("  Screen session: %s\n", theme.SuccessStyle.Render("✓ comfyui session active"))
	} else {
		fmt.Printf("  Screen session: %s\n", theme.WarningStyle.Render("❌ No comfyui session"))
	}

	// 3. Check if python process is running
	psCmd := exec.Command("pgrep", "-f", "ComfyUI.*main.py")
	psOutput, _ := psCmd.Output()
	hasPythonProcess := len(strings.TrimSpace(string(psOutput))) > 0

	if hasPythonProcess {
		pids := strings.TrimSpace(string(psOutput))
		fmt.Printf("  Python process: %s (PID: %s)\n", theme.SuccessStyle.Render("✓ Running"), theme.DimTextStyle.Render(pids))
	} else {
		fmt.Printf("  Python process: %s\n", theme.WarningStyle.Render("❌ Not running"))
	}

	// 4. Check if port 8188 is listening
	portCmd := exec.Command("bash", "-c", "netstat -tuln 2>/dev/null | grep :8188 || ss -tuln 2>/dev/null | grep :8188")
	portOutput, _ := portCmd.Output()
	hasPort := len(portOutput) > 0

	if hasPort {
		fmt.Printf("  Port 8188: %s\n", theme.SuccessStyle.Render("✓ Listening"))
	} else {
		fmt.Printf("  Port 8188: %s\n", theme.WarningStyle.Render("❌ Not listening"))
	}
	fmt.Println()

	// 5. Try to hit the API
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🌐 API Health"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	publicIP := getPublicIPForComfyUI()
	localURL := "http://127.0.0.1:8188"
	publicURL := fmt.Sprintf("http://%s:8188", publicIP)

	client := &http.Client{Timeout: 2 * time.Second}

	// Try local endpoint
	resp, err := client.Get(localURL)
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		fmt.Printf("  Local endpoint: %s\n", theme.SuccessStyle.Render(fmt.Sprintf("✓ %s responding", localURL)))
		fmt.Printf("  Public URL: %s\n", theme.HighlightStyle.Render(publicURL))
	} else if err != nil {
		fmt.Printf("  Local endpoint: %s\n", theme.WarningStyle.Render(fmt.Sprintf("❌ %s not responding", localURL)))
		fmt.Printf("  Error: %s\n", theme.DimTextStyle.Render(err.Error()))
	}
	fmt.Println()

	// 6. Show recent logs/errors if available
	if hasScreen {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("📋 Recent Logs"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		// Try to get screenlog if it exists
		screenlogPath := filepath.Join(homeDir, "screenlog.0")
		if _, err := os.Stat(screenlogPath); err == nil {
			tailCmd := exec.Command("tail", "-20", screenlogPath)
			if output, err := tailCmd.Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if line != "" {
						if strings.Contains(strings.ToLower(line), "error") {
							fmt.Println(theme.WarningStyle.Render("  " + line))
						} else {
							fmt.Println(theme.DimTextStyle.Render("  " + line))
						}
					}
				}
			}
		} else {
			fmt.Println(theme.DimTextStyle.Render("  View logs with: anime comfyui logs"))
			fmt.Println(theme.DimTextStyle.Render("  Or attach to screen: screen -r comfyui"))
		}
		fmt.Println()
	}

	// 7. GPU status if running
	if hasPythonProcess {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("🎮 GPU Usage"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		// Try nvidia-smi to see if ComfyUI is using GPU
		nvidiaSmiCmd := exec.Command("nvidia-smi", "--query-compute-apps=pid,process_name,used_memory", "--format=csv,noheader")
		if output, err := nvidiaSmiCmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			foundGPUUsage := false
			for _, line := range lines {
				if strings.Contains(line, "python") {
					fmt.Println(theme.SuccessStyle.Render("  ✓ Using GPU: " + strings.TrimSpace(line)))
					foundGPUUsage = true
				}
			}
			if !foundGPUUsage {
				fmt.Println(theme.WarningStyle.Render("  ⚠️  No GPU usage detected"))
			}
		}
		fmt.Println()
	}

	// 8. Show actions
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 Actions"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if !hasPythonProcess {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui start"))
		fmt.Println(theme.DimTextStyle.Render("    Start ComfyUI server"))
		fmt.Println()
	} else {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui logs"))
		fmt.Println(theme.DimTextStyle.Render("    View server logs"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime comfyui stop"))
		fmt.Println(theme.DimTextStyle.Render("    Stop ComfyUI server"))
		fmt.Println()
		if hasPort {
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("Open in browser:"))
			fmt.Printf("    %s\n", theme.GlowStyle.Render(publicURL))
			fmt.Println()
		}
	}

	return nil
}

func logsComfyUIServer() error {
	fmt.Println(theme.InfoStyle.Render("📋 Render engine logs (Ctrl+C to exit)"))
	fmt.Println()

	home, _ := os.UserHomeDir()
	logFile := filepath.Join(home, ".anime", "comfyui.log")
	if _, err := os.Stat(logFile); err == nil {
		cmd := exec.Command("tail", "-n", "200", "-F", logFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println(theme.WarningStyle.Render("No log file at " + logFile))
	fmt.Println(theme.DimTextStyle.Render("ComfyUI may not have been started by this CLI yet."))
	fmt.Println(theme.DimTextStyle.Render("Attach to the running screen session instead:"))
	fmt.Println(theme.HighlightStyle.Render("  screen -r comfyui   (Ctrl+A D to detach)"))
	return nil
}

func getPublicIPForComfyUI() string {
	// Try to get hostname first (e.g., 209-20-159-132)
	cmd := exec.Command("hostname")
	if output, err := cmd.Output(); err == nil {
		hostname := strings.TrimSpace(string(output))
		// If hostname looks like an IP with dashes, convert to dots
		if strings.Count(hostname, "-") >= 3 {
			parts := strings.Split(hostname, "-")
			if len(parts) >= 4 {
				// Check if looks like IP format
				allNumeric := true
				for _, part := range parts[:4] {
					for _, ch := range part {
						if ch < '0' || ch > '9' {
							allNumeric = false
							break
						}
					}
				}
				if allNumeric {
					return strings.Join(parts[:4], ".")
				}
			}
		}
		return hostname
	}

	// Fallback: try to get public IP from external service. Hard-cap the curl
	// at 3s so a flaky network can't stall studio bootstrap.
	cmd = exec.Command("curl", "-s", "--max-time", "3", "ifconfig.me")
	if output, err := cmd.Output(); err == nil {
		ip := strings.TrimSpace(string(output))
		if ip != "" {
			return ip
		}
	}

	// Last resort: return localhost
	return "127.0.0.1"
}
