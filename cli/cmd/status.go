package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [server-name]",
	Short: "Check status of a Lambda server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// FIRST: Check if we're already on a GPU server
		localExecutor := &LocalExecutor{}
		if isRunningOnGPUServer(localExecutor) {
			return runLocalStatus()
		}

		cfg, err := config.Load()
		if err != nil {
			// If config doesn't exist, try to auto-detect
			cfg = &config.Config{}
		}

		// Default to lambda if no argument provided
		serverName := "lambda"
		if len(args) > 0 {
			serverName = args[0]
		}

		// Try to get server from config
		server, err := cfg.GetServer(serverName)
		if err != nil {
			// Server not found in config - try to use it as an alias
			if alias := cfg.GetAlias(serverName); alias != "" {
				// Parse alias target
				var user, host string
				if strings.Contains(alias, "@") {
					parts := strings.SplitN(alias, "@", 2)
					user = parts[0]
					host = parts[1]
				} else {
					user = "ubuntu"
					host = alias
				}
				server = &config.Server{
					Name:        serverName,
					Host:        host,
					User:        user,
					SSHKey:      "",
					CostPerHour: 0, // Will be auto-detected
				}
			} else {
				// Server not found - run locally
				return runLocalStatus()
			}
		}

		// Check if server is localhost/local - just run locally
		if isLocalServer(server.Host) {
			return runLocalStatus()
		}

		fmt.Println()
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🔌 Connecting to %s...", server.Name)))

		client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
		if err != nil {
			// Connection failed - might be running locally
			return runLocalStatus()
		}
		defer client.Close()

		inst := installer.New(client)

		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("💻 System Information:"))
		fmt.Println()
		info, err := inst.GetSystemInfo()
		if err != nil {
			return err
		}

		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("OS:"), theme.InfoStyle.Render(info["os"]))
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Architecture:"), theme.InfoStyle.Render(info["arch"]))
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Kernel:"), theme.InfoStyle.Render(info["kernel"]))
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("GPU:"), theme.HighlightStyle.Render(info["gpu"]))
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Free Disk:"), theme.SuccessStyle.Render(info["disk_free"]))
		fmt.Printf("  %s  %s\n", theme.DimTextStyle.Render("Free Memory:"), theme.SuccessStyle.Render(info["mem_free"]))

		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("📦 Installed Components:"))
		fmt.Println()

		// Check for various components
		checks := map[string]string{
			"Python":     "python3 --version 2>/dev/null",
			"Node.js":    "node --version 2>/dev/null",
			"Docker":     "docker --version 2>/dev/null",
			"NVIDIA":     "nvidia-smi --version 2>/dev/null | head -1",
			"CUDA":       "nvcc --version 2>/dev/null | grep release",
			"PyTorch":    "python3 -c 'import torch; print(torch.__version__)' 2>/dev/null",
			"Ollama":     "ollama --version 2>/dev/null",
			"ComfyUI":    "[ -d ~/ComfyUI ] && echo 'Installed' || echo 'Not found'",
			"Claude Code": "claude-code --version 2>/dev/null",
		}

		for name, cmd := range checks {
			output, err := client.RunCommand(cmd)
			output = strings.TrimSpace(output)
			if err != nil || output == "Not found" || output == "" {
				fmt.Printf("  %s %s %s\n",
					theme.DimTextStyle.Render(name+":"),
					theme.WarningStyle.Render("✗"),
					theme.WarningStyle.Render("Not installed"))
			} else {
				fmt.Printf("  %s %s %s\n",
					theme.DimTextStyle.Render(name+":"),
					theme.SuccessStyle.Render("✓"),
					theme.InfoStyle.Render(output))
			}
		}

		// List Ollama models if installed
		output, err := client.RunCommand("ollama list 2>/dev/null")
		if err == nil && output != "" {
			fmt.Println()
			fmt.Println(theme.GlowStyle.Render("🤖 Ollama Models:"))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render(output))
		}

		// Next steps
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.GlowStyle.Render("💡 What to do next:"))
		fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime metrics"))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render("View GPU metrics and cost tracking"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages status"))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render("Check installation status of all packages"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime workstation"))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render("Launch interactive monitoring dashboard"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <package-id>"))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render("Install additional packages"))
		fmt.Println()

		return nil
	},
}

func runLocalStatus() error {
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Running on local GPU server"))
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("💻 System Information:"))
	fmt.Println()

	// Run local commands to get system info
	checks := map[string]struct {
		cmd  string
		desc string
	}{
		"OS":           {"lsb_release -d 2>/dev/null | cut -f2- || uname -s", "OS"},
		"Architecture": {"uname -m", "Architecture"},
		"Kernel":       {"uname -r", "Kernel"},
		"GPU":          {"nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | head -1", "GPU"},
		"Free Disk":    {"df -h / | awk 'NR==2 {print $4}'", "Free Disk"},
		"Free Memory":  {"free -h | awk 'NR==2 {print $4}'", "Free Memory"},
	}

	for name, check := range checks {
		output, err := runLocalCommand(check.cmd)
		if err != nil || output == "" {
			fmt.Printf("  %s  %s\n",
				theme.DimTextStyle.Render(name+":"),
				theme.WarningStyle.Render("N/A"))
		} else {
			value := strings.TrimSpace(output)
			// Highlight GPU in particular
			style := theme.InfoStyle
			if name == "GPU" {
				style = theme.HighlightStyle
			} else if name == "Free Disk" || name == "Free Memory" {
				style = theme.SuccessStyle
			}
			fmt.Printf("  %s  %s\n",
				theme.DimTextStyle.Render(name+":"),
				style.Render(value))
		}
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("📦 Installed Components:"))
	fmt.Println()

	// Check for various components locally
	componentChecks := map[string]string{
		"Python":      "python3 --version 2>/dev/null",
		"Node.js":     "node --version 2>/dev/null",
		"Docker":      "docker --version 2>/dev/null",
		"NVIDIA":      "nvidia-smi --version 2>/dev/null | head -1",
		"CUDA":        "nvcc --version 2>/dev/null | grep release",
		"PyTorch":     "python3 -c 'import torch; print(torch.__version__)' 2>/dev/null",
		"Ollama":      "ollama --version 2>/dev/null",
		"ComfyUI":     "[ -d ~/ComfyUI ] && echo 'Installed' || echo 'Not found'",
		"Claude Code": "claude-code --version 2>/dev/null",
	}

	for name, cmdStr := range componentChecks {
		output, err := runLocalCommand(cmdStr)
		output = strings.TrimSpace(output)
		if err != nil || output == "Not found" || output == "" {
			fmt.Printf("  %s %s %s\n",
				theme.DimTextStyle.Render(name+":"),
				theme.WarningStyle.Render("✗"),
				theme.WarningStyle.Render("Not installed"))
		} else {
			fmt.Printf("  %s %s %s\n",
				theme.DimTextStyle.Render(name+":"),
				theme.SuccessStyle.Render("✓"),
				theme.InfoStyle.Render(output))
		}
	}

	// List Ollama models if installed
	output, err := runLocalCommand("ollama list 2>/dev/null")
	if err == nil && output != "" {
		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("🤖 Ollama Models:"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render(output))
	}

	// Next steps
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.GlowStyle.Render("💡 What to do next:"))
	fmt.Println(theme.InfoStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime metrics"))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render("View GPU metrics and cost tracking"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages status"))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render("Check installation status of all packages"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime workstation"))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render("Launch interactive monitoring dashboard"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <package-id>"))
	fmt.Printf("    %s\n", theme.DimTextStyle.Render("Install additional packages"))
	fmt.Println()

	return nil
}

func runLocalCommand(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func isLocalServer(host string) bool {
	// Check if host is localhost, 127.0.0.1, or the current hostname
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "local" {
		return true
	}

	// Get current hostname and compare
	cmd := exec.Command("hostname")
	output, err := cmd.Output()
	if err == nil {
		currentHost := strings.TrimSpace(string(output))

		// Normalize both for comparison (replace periods and dashes with same char)
		normalizeHost := func(h string) string {
			h = strings.ReplaceAll(h, ".", "-")
			h = strings.ReplaceAll(h, "_", "-")
			return strings.ToLower(h)
		}

		if normalizeHost(host) == normalizeHost(currentHost) {
			return true
		}

		// Also check if they contain each other
		if strings.Contains(normalizeHost(currentHost), normalizeHost(host)) ||
		   strings.Contains(normalizeHost(host), normalizeHost(currentHost)) {
			return true
		}
	}

	// Also check local IP addresses
	cmd = exec.Command("hostname", "-I")
	output, err = cmd.Output()
	if err == nil {
		ips := strings.Fields(string(output))
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			// Normalize both for comparison
			normalizeHost := func(h string) string {
				h = strings.ReplaceAll(h, ".", "-")
				h = strings.ReplaceAll(h, "_", "-")
				return strings.ToLower(h)
			}
			// Compare both with dots and with dashes
			if ip == host || normalizeHost(ip) == normalizeHost(host) {
				return true
			}
		}
	}

	return false
}

