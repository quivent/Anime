package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	startServer string
	startLocal  bool
)

var startCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start services on your Lambda server (interactive)",
	Long: `Start common services with automatic port forwarding.

Services:
  comfyui     - Stable Diffusion UI (port 8188)
  ollama      - LLM inference server (port 11434)
  jupyter     - Jupyter notebook (port 8888)
  serve       - HTTP file server (custom port)

Examples:
  anime start              # Interactive menu
  anime start comfyui      # Start ComfyUI with port forwarding
  anime start ollama       # Start Ollama server
  anime start jupyter      # Start Jupyter notebook`,
	RunE: runStart,
}

func init() {
	startCmd.Flags().StringVarP(&startServer, "server", "s", "lambda", "Server to use")
	startCmd.Flags().BoolVarP(&startLocal, "local", "l", false, "Start locally instead of on server")
	rootCmd.AddCommand(startCmd)

	// anime serve llama → delegates to start llama
	serveCmd.AddCommand(&cobra.Command{
		Use:   "llama",
		Short: "Start Llama 3.3 70B + 1B spec decode on vLLM",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd, []string{"llama"})
		},
	})
}

func runStart(cmd *cobra.Command, args []string) error {
	// If no service specified, show interactive menu
	if len(args) == 0 {
		return showStartMenu()
	}

	service := strings.ToLower(args[0])

	// Check if we should run locally
	if startLocal {
		return runStartLocal(service)
	}

	// Get server target
	target, host, user, keyPath, err := getServerTarget(startServer)
	if err != nil {
		// If no server configured, run locally
		fmt.Println(theme.WarningStyle.Render("⚠️  No remote server configured, running locally..."))
		fmt.Println()
		return runStartLocal(service)
	}

	// Check if we're already on the target server (localhost)
	if isLocalhost(host) {
		fmt.Println(theme.InfoStyle.Render("✨ You're already on the server, running locally..."))
		fmt.Println()
		return runStartLocal(service)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 ANIME START 🚀"))
	fmt.Println()
	fmt.Printf("  Server:  %s\n", theme.HighlightStyle.Render(target))
	fmt.Printf("  Service: %s\n", theme.HighlightStyle.Render(service))
	fmt.Println()

	// Create SSH client with key from config
	sshClient, err := ssh.NewClient(host, user, keyPath)
	if err != nil {
		// If SSH fails, try running locally
		fmt.Println(theme.WarningStyle.Render("⚠️  Can't connect to server, running locally..."))
		fmt.Println()
		return runStartLocal(service)
	}
	defer sshClient.Close()

	// Start the requested service
	switch service {
	case "comfyui", "ui", "comfy":
		return startComfyUI(sshClient, host)
	case "ollama", "llm":
		return startOllama(sshClient, host)
	case "jupyter", "notebook", "lab":
		return startJupyter(sshClient, host)
	case "serve", "http":
		return startHTTPServer(sshClient, host)
	case "llama", "vllm":
		return startLlama(sshClient, host)
	default:
		return fmt.Errorf("unknown service: %s\n\nRun 'anime start' for options", service)
	}
}

func showStartMenu() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎯 WHAT DO YOU WANT TO DO? 🎯"))
	fmt.Println()

	options := []struct {
		cmd  string
		desc string
		icon string
	}{
		{"anime start comfyui", "Generate images with Stable Diffusion", "🎨"},
		{"anime start ollama", "Run local LLMs (Llama, Mistral, etc)", "🤖"},
		{"anime start jupyter", "Code in Jupyter notebooks", "📓"},
		{"anime start serve", "Share files over HTTP", "🌐"},
		{"anime jobs", "See what's currently running", "⚙️"},
		{"anime ssh", "SSH into your server", "💻"},
	}

	for _, opt := range options {
		fmt.Printf("  %s  %s\n", opt.icon, theme.SuccessStyle.Render(opt.cmd))
		fmt.Printf("      %s\n", theme.DimTextStyle.Render(opt.desc))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("💡 Quick shortcuts:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime ui      → anime start comfyui"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime llm     → anime start ollama"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime notebook → anime start jupyter"))
	fmt.Println()

	return nil
}

func startComfyUI(client *ssh.Client, host string) error {
	fmt.Println(theme.InfoStyle.Render("🎨 Starting ComfyUI..."))
	fmt.Println()

	// Check if already running
	output, _ := client.RunCommand("pgrep -f 'python.*main.py.*ComfyUI'")
	if strings.TrimSpace(output) != "" {
		fmt.Println(theme.WarningStyle.Render("⚠️  ComfyUI is already running!"))
		fmt.Println()
		return setupPortForwarding(host, "8188", "ComfyUI", "http://localhost:8188")
	}

	// Check if ComfyUI is installed
	exists, _ := client.RunCommand("test -d ~/ComfyUI && echo yes")
	if strings.TrimSpace(exists) != "yes" {
		fmt.Println(theme.ErrorStyle.Render("❌ ComfyUI not installed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Install it first:"))
		fmt.Println(theme.HighlightStyle.Render("  anime install comfyui"))
		fmt.Println()
		return fmt.Errorf("ComfyUI not installed")
	}

	// Start ComfyUI in background
	startCmd := `cd ~/ComfyUI && nohup python main.py --listen 0.0.0.0 --port 8188 > ~/comfyui.log 2>&1 & echo $! > ~/comfyui.pid`
	_, err := client.RunCommand(startCmd)
	if err != nil {
		return fmt.Errorf("failed to start ComfyUI: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ ComfyUI started"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Waiting for server to be ready..."))

	// Wait for server to start
	for i := 0; i < 10; i++ {
		output, _ := client.RunCommand("curl -s http://localhost:8188 > /dev/null && echo ready")
		if strings.Contains(output, "ready") {
			break
		}
		fmt.Print(".")
		exec.Command("sleep", "1").Run()
	}
	fmt.Println()
	fmt.Println()

	return setupPortForwarding(host, "8188", "ComfyUI", "http://localhost:8188")
}

func startOllama(client *ssh.Client, host string) error {
	fmt.Println(theme.InfoStyle.Render("🤖 Starting Ollama..."))
	fmt.Println()

	// Check if already running
	output, _ := client.RunCommand("pgrep -f 'ollama serve'")
	if strings.TrimSpace(output) != "" {
		fmt.Println(theme.WarningStyle.Render("⚠️  Ollama is already running!"))
		fmt.Println()
		return setupPortForwarding(host, "11434", "Ollama", "http://localhost:11434")
	}

	// Check if Ollama is installed
	exists, _ := client.RunCommand("which ollama")
	if strings.TrimSpace(exists) == "" {
		fmt.Println(theme.ErrorStyle.Render("❌ Ollama not installed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Install it first:"))
		fmt.Println(theme.HighlightStyle.Render("  anime install ollama"))
		fmt.Println()
		return fmt.Errorf("Ollama not installed")
	}

	// Start Ollama in background
	startCmd := `nohup ollama serve > ~/ollama.log 2>&1 & echo $! > ~/ollama.pid`
	_, err := client.RunCommand(startCmd)
	if err != nil {
		return fmt.Errorf("failed to start Ollama: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Ollama started"))
	fmt.Println()

	return setupPortForwarding(host, "11434", "Ollama", "http://localhost:11434")
}

func startJupyter(client *ssh.Client, host string) error {
	fmt.Println(theme.InfoStyle.Render("📓 Starting Jupyter..."))
	fmt.Println()

	// Check if already running
	output, _ := client.RunCommand("pgrep -f 'jupyter'")
	if strings.TrimSpace(output) != "" {
		fmt.Println(theme.WarningStyle.Render("⚠️  Jupyter is already running!"))
		fmt.Println()
		return setupPortForwarding(host, "8888", "Jupyter", "http://localhost:8888")
	}

	// Start Jupyter in background
	startCmd := `nohup jupyter lab --ip=0.0.0.0 --port=8888 --no-browser --allow-root > ~/jupyter.log 2>&1 & echo $! > ~/jupyter.pid`
	_, err := client.RunCommand(startCmd)
	if err != nil {
		return fmt.Errorf("failed to start Jupyter: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ Jupyter started"))
	fmt.Println()

	// Get the token
	fmt.Println(theme.DimTextStyle.Render("  Waiting for token..."))
	var token string
	for i := 0; i < 10; i++ {
		output, _ := client.RunCommand("grep -oP 'token=\\K[a-f0-9]+' ~/jupyter.log | head -1")
		token = strings.TrimSpace(output)
		if token != "" {
			break
		}
		exec.Command("sleep", "1").Run()
	}

	url := "http://localhost:8888"
	if token != "" {
		url = fmt.Sprintf("http://localhost:8888/?token=%s", token)
	}

	return setupPortForwarding(host, "8888", "Jupyter", url)
}

func startHTTPServer(client *ssh.Client, host string) error {
	fmt.Println(theme.InfoStyle.Render("🌐 Starting HTTP Server..."))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Use 'anime serve <path>' for more control"))
	fmt.Println()
	return fmt.Errorf("please use 'anime serve <path>' to start HTTP server")
}

func setupPortForwarding(host, port, serviceName, url string) error {
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🌐 %s is ready!", serviceName)))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("📡 Setting up port forwarding..."))
	fmt.Println()

	// Start SSH port forwarding
	sshCmd := exec.Command("ssh", "-N", "-L", fmt.Sprintf("%s:localhost:%s", port, port), "ubuntu@"+host)

	if err := sshCmd.Start(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Port forwarding failed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Manual forwarding:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("ssh -L %s:localhost:%s ubuntu@%s", port, port, host)))
		fmt.Println()
		return err
	}

	// Wait a moment for port forwarding to establish
	exec.Command("sleep", "2").Run()

	fmt.Println(theme.SuccessStyle.Render("✓ Port forwarding active"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Access Information"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(url))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  ✨ Opening in your browser in 3 seconds..."))
	fmt.Println()

	// Open browser after a delay
	go func() {
		exec.Command("sleep", "3").Run()
		openBrowser(url)
	}()

	fmt.Println(theme.InfoStyle.Render("⚙️  Port forwarding is active"))
	fmt.Println(theme.DimTextStyle.Render("   Press Ctrl+C to stop"))
	fmt.Println()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println()
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🛑 Stopping port forwarding..."))
	sshCmd.Process.Kill()
	fmt.Println(theme.SuccessStyle.Render("✓ Stopped"))
	fmt.Println()

	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch {
	case exec.Command("which", "open").Run() == nil:
		cmd = exec.Command("open", url)
	case exec.Command("which", "xdg-open").Run() == nil:
		cmd = exec.Command("xdg-open", url)
	case exec.Command("which", "wsl-open").Run() == nil:
		cmd = exec.Command("wsl-open", url)
	default:
		return
	}

	cmd.Run()
}

func getServerTarget(server string) (target, host, user, keyPath string, err error) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to load config: %w", err)
	}

	// Parse server argument
	if strings.Contains(server, "@") {
		target = server
		parts := strings.Split(target, "@")
		user = parts[0]
		host = parts[1]
		keyPath = "" // Use default keys
	} else if strings.Contains(server, ".") {
		user = "ubuntu"
		host = server
		target = user + "@" + host
		keyPath = "" // Use default keys
	} else {
		// Try alias first
		target = cfg.GetAlias(server)
		if target == "" {
			// Try server name
			if s, err := cfg.GetServer(server); err == nil {
				target = fmt.Sprintf("%s@%s", s.User, s.Host)
				keyPath = s.SSHKey // Get SSH key from config
			}
		}

		if target == "" {
			return "", "", "", "", fmt.Errorf("server not found: %s\n\nConfigure it with: anime set %s <ip>", server, server)
		}

		parts := strings.Split(target, "@")
		if len(parts) == 2 {
			user = parts[0]
			host = parts[1]
		}
	}

	return target, host, user, keyPath, nil
}

func startLlama(client *ssh.Client, host string) error {
	fmt.Println(theme.InfoStyle.Render("🦙 Starting Llama 3.3 70B + 1B spec decode on vLLM..."))
	fmt.Println()

	// Check if already running
	output, _ := client.RunCommand("curl -s http://localhost:8000/health")
	if strings.TrimSpace(output) != "" {
		fmt.Println(theme.SuccessStyle.Render("✓ vLLM Llama server already running on :8000"))
		fmt.Println()
		return setupPortForwarding(host, "8000", "vLLM Llama", "http://localhost:8000/v1")
	}

	// Check if vLLM is installed
	exists, _ := client.RunCommand("python3 -c 'import vllm' 2>/dev/null && echo yes")
	if strings.TrimSpace(exists) != "yes" {
		fmt.Println(theme.ErrorStyle.Render("❌ vLLM not installed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Install it first:"))
		fmt.Println(theme.HighlightStyle.Render("  anime install llama"))
		return fmt.Errorf("vLLM not installed")
	}

	// Detect GPU count for tensor parallel
	gpuCount, _ := client.RunCommand("nvidia-smi --query-gpu=name --format=csv,noheader | wc -l")
	tp := strings.TrimSpace(gpuCount)
	if tp == "" || tp == "0" {
		tp = "1"
	}

	// Start vLLM with speculative decoding
	startScript := fmt.Sprintf(`screen -dmS vllm-llama bash -c '
python3 -m vllm.entrypoints.openai.api_server \
    --model meta-llama/Llama-3.3-70B-Instruct \
    --speculative-model meta-llama/Llama-3.2-1B-Instruct \
    --num-speculative-tokens 5 \
    --use-v2-block-manager \
    --tensor-parallel-size %s \
    --host 0.0.0.0 \
    --port 8000 \
    --trust-remote-code \
    2>&1 | tee /tmp/vllm-llama.log
'`, tp)

	_, err := client.RunCommand(startScript)
	if err != nil {
		return fmt.Errorf("failed to start vLLM: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ vLLM Llama server starting (loading 70B model)"))
	fmt.Println(theme.DimTextStyle.Render("  Waiting for server to be ready..."))

	// Wait for server
	for i := 0; i < 120; i++ {
		output, _ := client.RunCommand("curl -s http://localhost:8000/health > /dev/null && echo ready")
		if strings.Contains(output, "ready") {
			break
		}
		fmt.Print(".")
		exec.Command("sleep", "5").Run()
	}
	fmt.Println()

	// Verify
	output, _ = client.RunCommand("curl -s http://localhost:8000/health")
	if strings.TrimSpace(output) == "" {
		fmt.Println(theme.ErrorStyle.Render("❌ Server did not start"))
		fmt.Println(theme.DimTextStyle.Render("  Check logs: ssh into server, then: screen -r vllm-llama"))
		return fmt.Errorf("vLLM server failed to start")
	}

	fmt.Println(theme.SuccessStyle.Render("✓ vLLM Llama server ready"))
	fmt.Println()
	return setupPortForwarding(host, "8000", "vLLM Llama", "http://localhost:8000/v1")
}

func startLlamaLocal() error {
	fmt.Println(theme.InfoStyle.Render("🦙 Starting Llama 3.3 70B + 1B spec decode on vLLM..."))
	fmt.Println()

	// Check if already running
	checkCmd := exec.Command("curl", "-s", "http://localhost:8000/health")
	if out, err := checkCmd.Output(); err == nil && len(out) > 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ vLLM Llama server already running on :8000"))
		fmt.Println(theme.HighlightStyle.Render("  API: http://localhost:8000/v1"))
		return nil
	}

	// Get GPU count
	gpuCmd := exec.Command("bash", "-c", "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l")
	gpuOut, _ := gpuCmd.Output()
	tp := strings.TrimSpace(string(gpuOut))
	if tp == "" || tp == "0" {
		tp = "1"
	}

	// Start in screen
	script := fmt.Sprintf(`screen -dmS vllm-llama bash -c '
python3 -m vllm.entrypoints.openai.api_server \
    --model meta-llama/Llama-3.3-70B-Instruct \
    --speculative-model meta-llama/Llama-3.2-1B-Instruct \
    --num-speculative-tokens 5 \
    --use-v2-block-manager \
    --tensor-parallel-size %s \
    --host 0.0.0.0 \
    --port 8000 \
    --trust-remote-code \
    2>&1 | tee /tmp/vllm-llama.log
'`, tp)

	cmd := exec.Command("bash", "-c", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start vLLM: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("✓ vLLM Llama server starting"))
	fmt.Println(theme.DimTextStyle.Render("  Loading 70B model — this takes a few minutes"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Monitor:  screen -r vllm-llama"))
	fmt.Println(theme.InfoStyle.Render("  Logs:     tail -f /tmp/vllm-llama.log"))
	fmt.Println(theme.InfoStyle.Render("  API:      http://localhost:8000/v1"))
	fmt.Println(theme.InfoStyle.Render("  Stop:     screen -S vllm-llama -X quit"))
	return nil
}

func isLocalhost(host string) bool {
	// Check common localhost identifiers
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	// Check if it's our current hostname
	currentHost, err := os.Hostname()
	if err == nil && currentHost == host {
		return true
	}

	return false
}

func runStartLocal(service string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 STARTING LOCALLY 🚀"))
	fmt.Println()
	fmt.Printf("  Service: %s\n", theme.HighlightStyle.Render(service))
	fmt.Println()

	// Map to anime run command
	switch service {
	case "comfyui", "ui", "comfy":
		return runService(nil, []string{"comfyui"})
	case "ollama", "llm":
		return runService(nil, []string{"ollama"})
	case "jupyter", "notebook", "lab":
		return runService(nil, []string{"jupyter"})
	case "tensorboard", "tb":
		return runService(nil, []string{"tensorboard"})
	case "llama", "vllm":
		return startLlamaLocal()
	default:
		return fmt.Errorf("unknown service: %s\n\nRun 'anime start' for options", service)
	}
}
