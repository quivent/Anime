package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	serverTarget string
	serverLocal  bool
)

var serverCmd = &cobra.Command{
	Use:   "server [target]",
	Short: "Show server information and GPU details",
	Long: `Display detailed information about the server including GPU count, types,
total RAM, uptime, and system specifications.

By default connects to the 'lambda' server alias. Use --local for local machine.

Examples:
  anime server                    # Show info for lambda (default)
  anime server production         # Show info for production server
  anime server ubuntu@10.0.0.5    # Show info for specific host
  anime server --local            # Show local machine info`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServerInfo,
}

var serverMonitorCmd = &cobra.Command{
	Use:   "monitor [target]",
	Short: "Open GPU monitoring TUI",
	Long: `Open an interactive TUI that displays real-time GPU metrics including:
- GPU utilization and inference load
- Memory usage per GPU
- Temperature/heat levels
- Power consumption

All GPUs are displayed side by side for easy comparison.

Examples:
  anime server monitor                    # Monitor lambda (default)
  anime server monitor production         # Monitor production server
  anime server monitor ubuntu@10.0.0.5    # Monitor specific host
  anime server monitor --local            # Monitor local machine`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServerMonitor,
}

func init() {
	// Server command flags
	serverCmd.PersistentFlags().BoolVar(&serverLocal, "local", false, "Use local machine instead of remote server")

	// Add subcommands
	serverCmd.AddCommand(serverMonitorCmd)

	// Register with root
	rootCmd.AddCommand(serverCmd)
}

func runServerInfo(cmd *cobra.Command, args []string) error {
	target := "lambda"
	if len(args) > 0 {
		target = args[0]
	}

	if serverLocal {
		return showLocalServerInfo()
	}

	return showRemoteServerInfo(target)
}

// resolveServerTargetInfo resolves server target to host and user
func resolveServerTargetInfo(cfg *config.Config, target string) (host, user string) {
	// Default user
	user = "ubuntu"

	// Check if it's an alias
	if alias := cfg.GetAlias(target); alias != "" {
		if strings.Contains(alias, "@") {
			parts := strings.SplitN(alias, "@", 2)
			user = parts[0]
			host = parts[1]
		} else {
			host = alias
		}
		return
	}

	// Check if it's a server config
	if server, err := cfg.GetServer(target); err == nil {
		host = server.Host
		user = server.User
		return
	}

	// Check if it looks like user@host
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		user = parts[0]
		host = parts[1]
		return
	}

	// Assume it's just a host
	host = target
	return
}

func showLocalServerInfo() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("SERVER INFO"))
	fmt.Println()

	hostname, _ := os.Hostname()
	fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Hostname:"), theme.HighlightStyle.Render(hostname))
	fmt.Printf("  %s %s/%s\n", theme.GlowStyle.Render("Platform:"), theme.InfoStyle.Render(runtime.GOOS), theme.InfoStyle.Render(runtime.GOARCH))
	fmt.Println()

	// GPU Information
	fmt.Println(theme.SuccessStyle.Render("  GPU INFORMATION"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	gpuInfo := getLocalGPUInfo()
	if gpuInfo == "" {
		fmt.Println(theme.DimTextStyle.Render("  No NVIDIA GPUs detected"))
	} else {
		fmt.Print(gpuInfo)
	}
	fmt.Println()

	// Memory Information
	fmt.Println(theme.SuccessStyle.Render("  SYSTEM MEMORY"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))
	memInfo := getLocalMemoryInfo()
	fmt.Print(memInfo)
	fmt.Println()

	// CPU Information
	fmt.Println(theme.SuccessStyle.Render("  CPU INFORMATION"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))
	cpuInfo := getLocalCPUInfo()
	fmt.Print(cpuInfo)
	fmt.Println()

	// Uptime
	fmt.Println(theme.SuccessStyle.Render("  UPTIME"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))
	uptime := getLocalUptime()
	fmt.Printf("  %s\n", theme.InfoStyle.Render(uptime))
	fmt.Println()

	return nil
}

func showRemoteServerInfo(target string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target
	host, user := resolveServerTargetInfo(cfg, target)

	fmt.Println()
	fmt.Println(theme.RenderBanner("SERVER INFO"))
	fmt.Println()
	fmt.Printf("  %s %s@%s\n", theme.GlowStyle.Render("Connecting to:"), theme.HighlightStyle.Render(user), theme.HighlightStyle.Render(host))
	fmt.Println()

	// Create SSH client
	client, err := ssh.NewClientWithOptions(host, user, "", GetSSHClientOptions())
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Get hostname
	hostname, _ := client.RunCommand("hostname")
	hostname = strings.TrimSpace(hostname)
	fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Hostname:"), theme.HighlightStyle.Render(hostname))

	// Get OS info
	osInfo, _ := client.RunCommand("cat /etc/os-release | grep PRETTY_NAME | cut -d'\"' -f2")
	osInfo = strings.TrimSpace(osInfo)
	if osInfo != "" {
		fmt.Printf("  %s %s\n", theme.GlowStyle.Render("OS:"), theme.InfoStyle.Render(osInfo))
	}

	// Get architecture
	arch, _ := client.RunCommand("uname -m")
	arch = strings.TrimSpace(arch)
	fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Architecture:"), theme.InfoStyle.Render(arch))
	fmt.Println()

	// GPU Information
	fmt.Println(theme.SuccessStyle.Render("  GPU INFORMATION"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	gpuOutput, err := client.RunCommand("nvidia-smi --query-gpu=index,name,memory.total,driver_version --format=csv,noheader")
	if err != nil {
		fmt.Println(theme.DimTextStyle.Render("  No NVIDIA GPUs detected or nvidia-smi not available"))
	} else {
		lines := strings.Split(strings.TrimSpace(gpuOutput), "\n")
		fmt.Printf("  %s %s\n\n", theme.GlowStyle.Render("GPU Count:"), theme.HighlightStyle.Render(fmt.Sprintf("%d", len(lines))))

		for _, line := range lines {
			fields := strings.Split(line, ", ")
			if len(fields) >= 4 {
				idx := strings.TrimSpace(fields[0])
				name := strings.TrimSpace(fields[1])
				memory := strings.TrimSpace(fields[2])
				driver := strings.TrimSpace(fields[3])

				fmt.Printf("  %s %s\n", theme.HighlightStyle.Render(fmt.Sprintf("GPU %s:", idx)), theme.SuccessStyle.Render(name))
				fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("Memory:"), theme.InfoStyle.Render(memory))
				fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("Driver:"), theme.InfoStyle.Render(driver))
				fmt.Println()
			}
		}
	}

	// Memory Information
	fmt.Println(theme.SuccessStyle.Render("  SYSTEM MEMORY"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	memOutput, _ := client.RunCommand("free -h | grep Mem")
	if memOutput != "" {
		fields := strings.Fields(memOutput)
		if len(fields) >= 3 {
			fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Total:"), theme.HighlightStyle.Render(fields[1]))
			fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Used:"), theme.InfoStyle.Render(fields[2]))
			if len(fields) >= 4 {
				fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Free:"), theme.SuccessStyle.Render(fields[3]))
			}
		}
	}
	fmt.Println()

	// CPU Information
	fmt.Println(theme.SuccessStyle.Render("  CPU INFORMATION"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	cpuModel, _ := client.RunCommand("lscpu | grep 'Model name' | cut -d':' -f2 | xargs")
	cpuModel = strings.TrimSpace(cpuModel)
	if cpuModel != "" {
		fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Model:"), theme.HighlightStyle.Render(cpuModel))
	}

	cpuCores, _ := client.RunCommand("nproc")
	cpuCores = strings.TrimSpace(cpuCores)
	if cpuCores != "" {
		fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Cores:"), theme.InfoStyle.Render(cpuCores))
	}
	fmt.Println()

	// Uptime
	fmt.Println(theme.SuccessStyle.Render("  UPTIME"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	uptime, _ := client.RunCommand("uptime -p 2>/dev/null || uptime")
	uptime = strings.TrimSpace(uptime)
	fmt.Printf("  %s\n", theme.InfoStyle.Render(uptime))
	fmt.Println()

	// Disk Space
	fmt.Println(theme.SuccessStyle.Render("  DISK SPACE"))
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────────────────"))

	diskOutput, _ := client.RunCommand("df -h / | tail -1")
	if diskOutput != "" {
		fields := strings.Fields(diskOutput)
		if len(fields) >= 5 {
			fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Total:"), theme.HighlightStyle.Render(fields[1]))
			fmt.Printf("  %s %s\n", theme.GlowStyle.Render("Used:"), theme.InfoStyle.Render(fields[2]))
			fmt.Printf("  %s %s (%s)\n", theme.GlowStyle.Render("Available:"), theme.SuccessStyle.Render(fields[3]), theme.DimTextStyle.Render(fields[4]+" used"))
		}
	}
	fmt.Println()

	// Quick tip
	fmt.Println(theme.DimTextStyle.Render("  Tip: Use 'anime server monitor' for real-time GPU monitoring"))
	fmt.Println()

	return nil
}

func runServerMonitor(cmd *cobra.Command, args []string) error {
	target := "lambda"
	if len(args) > 0 {
		target = args[0]
	}

	if serverLocal {
		return tui.RunServerMonitor("", "", true)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target
	host, user := resolveServerTargetInfo(cfg, target)

	return tui.RunServerMonitor(host, user, false)
}

// Local system info helpers
func getLocalGPUInfo() string {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var result strings.Builder
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	result.WriteString(fmt.Sprintf("  %s %s\n\n", theme.GlowStyle.Render("GPU Count:"), theme.HighlightStyle.Render(fmt.Sprintf("%d", len(lines)))))

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) >= 4 {
			idx := strings.TrimSpace(fields[0])
			name := strings.TrimSpace(fields[1])
			memory := strings.TrimSpace(fields[2])
			driver := strings.TrimSpace(fields[3])

			result.WriteString(fmt.Sprintf("  %s %s\n", theme.HighlightStyle.Render(fmt.Sprintf("GPU %s:", idx)), theme.SuccessStyle.Render(name)))
			result.WriteString(fmt.Sprintf("    %s %s\n", theme.DimTextStyle.Render("Memory:"), theme.InfoStyle.Render(memory)))
			result.WriteString(fmt.Sprintf("    %s %s\n", theme.DimTextStyle.Render("Driver:"), theme.InfoStyle.Render(driver)))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func getLocalMemoryInfo() string {
	var result strings.Builder

	if runtime.GOOS == "linux" {
		cmd := exec.Command("sh", "-c", "free -h | grep Mem")
		output, err := cmd.Output()
		if err == nil {
			fields := strings.Fields(string(output))
			if len(fields) >= 3 {
				result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Total:"), theme.HighlightStyle.Render(fields[1])))
				result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Used:"), theme.InfoStyle.Render(fields[2])))
				if len(fields) >= 4 {
					result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Free:"), theme.SuccessStyle.Render(fields[3])))
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", "sysctl -n hw.memsize")
		output, err := cmd.Output()
		if err == nil {
			totalBytes := parseServerInt(strings.TrimSpace(string(output)))
			totalGB := totalBytes / (1024 * 1024 * 1024)
			result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Total:"), theme.HighlightStyle.Render(fmt.Sprintf("%d GB", totalGB))))
		}
	}

	return result.String()
}

func getLocalCPUInfo() string {
	var result strings.Builder

	if runtime.GOOS == "linux" {
		cmd := exec.Command("sh", "-c", "lscpu | grep 'Model name' | cut -d':' -f2 | xargs")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Model:"), theme.HighlightStyle.Render(strings.TrimSpace(string(output)))))
		}
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Model:"), theme.HighlightStyle.Render(strings.TrimSpace(string(output)))))
		}
	}

	result.WriteString(fmt.Sprintf("  %s %s\n", theme.GlowStyle.Render("Cores:"), theme.InfoStyle.Render(fmt.Sprintf("%d", runtime.NumCPU()))))

	return result.String()
}

func getLocalUptime() string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("uptime", "-p")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", "uptime | awk -F'up ' '{print $2}' | awk -F',' '{print $1, $2}'")
		output, err := cmd.Output()
		if err == nil {
			return "up " + strings.TrimSpace(string(output))
		}
	}
	return "unknown"
}

func parseServerInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
