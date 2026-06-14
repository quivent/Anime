package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var comfyuiCmd = &cobra.Command{
	Use:   "comfyui <start|stop|status|logs>",
	Short: "Manage ComfyUI server",
	Args:  cobra.ExactArgs(1),
	RunE:  runComfyUICommand,
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
	fmt.Println(theme.InfoStyle.Render("🚀 Starting ComfyUI in background..."))

	// Pick the venv python (where torch cu130 + sageattention live) when it
	// exists; fall back to system python3 only if the venv is missing — that
	// case will probably crash on `import torch`, but at least we don't
	// silently mask the user's broken install behind a system python that
	// happens to have a different (wrong) torch.
	home, _ := os.UserHomeDir()
	venvPy := filepath.Join(home, "ComfyUI", "venv", "bin", "python")
	pyCmd := "python3"
	sageFlag := ""
	if _, err := os.Stat(venvPy); err == nil {
		pyCmd = "./venv/bin/python"
		// Only enable --use-sage-attention if the package is actually present
		// in the venv. The studio bootstrap installs it via the wantorch
		// phase, but a user running `anime comfyui start` directly (without
		// going through wan studio) may have a venv with plain torch only —
		// passing the flag in that case makes ComfyUI refuse to start.
		// Glob covers python3.10/3.11/3.12 site-packages dirs.
		sageGlob := filepath.Join(home, "ComfyUI", "venv", "lib", "python*", "site-packages", "sageattention")
		if matches, _ := filepath.Glob(sageGlob); len(matches) > 0 {
			sageFlag = " --use-sage-attention"
		}
	}
	// Ensure ~/.anime exists before tee writes into it. tee creates the file
	// but not the parent dir, so a fresh box that hits `anime comfyui start`
	// before any other CLI command would otherwise fail the log pipe.
	animeDir := filepath.Join(home, ".anime")
	if err := os.MkdirAll(animeDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", animeDir, err)
	}
	logPath := filepath.Join(animeDir, "comfyui.log")
	launch := fmt.Sprintf("cd ~/ComfyUI && exec %s main.py --listen%s 2>&1 | tee -a %s", pyCmd, sageFlag, logPath)
	cmd := exec.Command("screen", "-dmS", "comfyui", "bash", "-c", launch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start ComfyUI: %w", err)
	}

	// Wait for the server to actually become reachable before claiming success.
	// ComfyUI takes 5-15s on a warm box, 60-90s on first boot with model imports.
	fmt.Print(theme.DimTextStyle.Render("  Waiting for ComfyUI to become reachable"))
	reachable := false
	for i := 0; i < 30; i++ {
		fmt.Print(".")
		time.Sleep(2 * time.Second)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://127.0.0.1:8188/system_stats")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				reachable = true
				break
			}
		}
	}
	fmt.Println()
	fmt.Println()

	publicIP := getPublicIPForComfyUI()
	if reachable {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ ComfyUI started and reachable at http://%s:8188", publicIP)))
	} else {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠ ComfyUI screen session launched but not yet reachable at http://%s:8188", publicIP)))
		fmt.Println(theme.DimTextStyle.Render("  First boot can take 60-90s for model imports."))
		fmt.Println(theme.DimTextStyle.Render("  Check progress: anime comfyui logs"))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("View logs:    anime comfyui logs"))
	fmt.Println(theme.DimTextStyle.Render("Check status: anime comfyui status"))
	fmt.Println(theme.DimTextStyle.Render("Stop server:  anime comfyui stop"))
	fmt.Println()

	return nil
}

func stopComfyUIServer() error {
	fmt.Println(theme.InfoStyle.Render("🛑 Stopping ComfyUI..."))

	cmd := exec.Command("screen", "-S", "comfyui", "-X", "quit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop ComfyUI: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ ComfyUI stopped"))
	return nil
}

func statusComfyUIServer() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 COMFYUI STATUS"))
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
	fmt.Println(theme.InfoStyle.Render("📋 ComfyUI logs (Ctrl+C to exit)"))
	fmt.Println()

	cmd := exec.Command("screen", "-r", "comfyui")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	fmt.Println(theme.DimTextStyle.Render("To view logs, run:"))
	fmt.Println(theme.HighlightStyle.Render("  screen -r comfyui"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("(Press Ctrl+A then D to detach)"))

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

	// Fallback: try to get public IP from external service
	cmd = exec.Command("curl", "-s", "ifconfig.me")
	if output, err := cmd.Output(); err == nil {
		ip := strings.TrimSpace(string(output))
		if ip != "" {
			return ip
		}
	}

	// Last resort: return localhost
	return "127.0.0.1"
}
