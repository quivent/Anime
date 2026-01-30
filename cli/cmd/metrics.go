package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show GPU usage, runtime, and cost metrics",
	Long: `Display comprehensive metrics for your Lambda GPU instance:
  • GPU utilization and memory usage
  • Instance uptime and runtime
  • Cost tracking based on instance pricing
  • Output files generated (images, videos, models)

Perfect for tracking your AI workloads and costs.`,
	RunE: runMetrics,
}

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show resource usage, runtime, and cost",
	Long: `Display comprehensive usage metrics for your Lambda GPU instance:
  • GPU utilization and memory usage
  • Instance uptime and runtime
  • Cost tracking based on instance pricing
  • Output files generated (images, videos, models)

Perfect for tracking your AI workloads and costs.`,
	RunE: runUsage,
}

var (
	localMode bool
)

func init() {
	usageCmd.Flags().BoolVarP(&localMode, "local", "l", false, "Run locally on this server")
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(usageCmd)
}

type Metrics struct {
	GPUs          []GPUMetric
	UptimeSeconds int64
	InstanceCost  float64
	Outputs       OutputMetrics
}

type GPUMetric struct {
	Index       int
	Name        string
	Utilization int
	MemoryUsed  string
	MemoryTotal string
	Temperature int
	PowerDraw   string
}

type OutputMetrics struct {
	Images      int
	Videos      int
	Models      int
	TotalSizeGB float64
}

func runMetrics(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		// Continue with default config if load fails
		cfg = &config.Config{}
	}

	// Check if we're running on a GPU server (detect local execution)
	executor := &LocalExecutor{}
	if isRunningOnGPUServer(executor) {
		// Running locally on the GPU server
		fmt.Println()
		fmt.Println(theme.RenderBanner("📊 LOCAL METRICS 📊"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Collecting metrics from local GPU server..."))
		fmt.Println()

		// Get metrics locally
		metrics, err := collectMetricsWithExecutor(executor, cfg)
		if err != nil {
			return fmt.Errorf("failed to collect metrics: %w", err)
		}

		// Get hostname
		hostname := "localhost"
		if hostnameOut, err := executor.RunCommand("hostname"); err == nil {
			hostname = strings.TrimSpace(hostnameOut)
		}

		// Display metrics
		displayMetrics(metrics, hostname)
		return nil
	}

	// Not running locally, try to connect to remote server
	// Get lambda target
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	// If no remote lambda server configured, show error
	if lambdaTarget == "" {
		fmt.Println()
		fmt.Println(theme.RenderBanner("📊 METRICS 📊"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  No GPU server detected locally and no remote server configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 To configure a remote Lambda server:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime config"))
		fmt.Println()
		return nil
	}

	// Parse target
	var user, host string
	if strings.Contains(lambdaTarget, "@") {
		parts := strings.SplitN(lambdaTarget, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = lambdaTarget
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("📊 LAMBDA METRICS 📊"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Fetching real-time metrics from "+host+"..."))
	fmt.Println()

	// Connect to server
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer sshClient.Close()

	// Get metrics
	metrics, err := collectMetrics(sshClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Display metrics
	displayMetrics(metrics, host)

	return nil
}

// isRunningOnGPUServer checks if we're running on a server with GPUs
func isRunningOnGPUServer(executor CommandExecutor) bool {
	// Check for nvidia-smi (primary indicator of GPU server)
	_, err := executor.RunCommand("command -v nvidia-smi")
	return err == nil
}

// CommandExecutor interface for running commands locally or remotely
type CommandExecutor interface {
	RunCommand(cmd string) (string, error)
}

// LocalExecutor runs commands on the local machine
type LocalExecutor struct{}

func (e *LocalExecutor) RunCommand(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// SSHExecutor wraps ssh.Client to implement CommandExecutor
type SSHExecutor struct {
	client *ssh.Client
}

func (e *SSHExecutor) RunCommand(cmd string) (string, error) {
	return e.client.RunCommand(cmd)
}

func runUsage(cmd *cobra.Command, args []string) error {
	// Use local executor
	executor := &LocalExecutor{}

	// Check if we're on a GPU server
	if !isRunningOnGPUServer(executor) {
		fmt.Println()
		fmt.Println(theme.RenderBanner("📊 USAGE METRICS 📊"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  No GPUs detected on this machine"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 This command must be run on a GPU server"))
		fmt.Println(theme.DimTextStyle.Render("    To monitor a remote server, use: anime metrics"))
		fmt.Println()
		return nil
	}

	// Running on GPU server - show local metrics
	fmt.Println()
	fmt.Println(theme.RenderBanner("📊 USAGE METRICS 📊"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Collecting local GPU metrics..."))
	fmt.Println()

	// Load config for cost info
	cfg, err := config.Load()
	if err != nil {
		// Continue with default cost if config fails
		cfg = &config.Config{}
	}

	// Get metrics
	metrics, err := collectMetricsWithExecutor(executor, cfg)
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Get hostname
	hostname := "localhost"
	if hostnameOut, err := executor.RunCommand("hostname"); err == nil {
		hostname = strings.TrimSpace(hostnameOut)
	}

	// Display metrics
	displayMetrics(metrics, hostname)

	return nil
}

// detectInstanceCost auto-detects hourly cost based on GPU type and count
// Based on Lambda Labs on-demand pricing (as of 2024)
func detectInstanceCost(gpus []GPUMetric) float64 {
	if len(gpus) == 0 {
		return 0.00
	}

	gpuModel := strings.ToUpper(gpus[0].Name)
	gpuCount := len(gpus)

	// Lambda Labs GPU pricing (per GPU per hour)
	var pricePerGPU float64

	switch {
	// NVIDIA H200 - Latest flagship
	case strings.Contains(gpuModel, "H200"):
		pricePerGPU = 4.50 // H200 141GB SXM

	// NVIDIA GH200 - Grace Hopper Superchip
	case strings.Contains(gpuModel, "GH200"):
		pricePerGPU = 3.50 // GH200 96GB

	// NVIDIA H100 - Flagship datacenter GPU
	case strings.Contains(gpuModel, "H100") && strings.Contains(gpuModel, "SXM"):
		pricePerGPU = 2.49 // H100 80GB SXM5
	case strings.Contains(gpuModel, "H100") && strings.Contains(gpuModel, "PCIE"):
		pricePerGPU = 2.29 // H100 80GB PCIe
	case strings.Contains(gpuModel, "H100"):
		pricePerGPU = 2.49 // Default H100 pricing

	// NVIDIA A100 - Previous generation flagship
	case strings.Contains(gpuModel, "A100") && strings.Contains(gpuModel, "SXM4") && strings.Contains(gpuModel, "80GB"):
		pricePerGPU = 1.29 // A100 80GB SXM4
	case strings.Contains(gpuModel, "A100") && strings.Contains(gpuModel, "SXM4"):
		pricePerGPU = 1.10 // A100 40GB SXM4
	case strings.Contains(gpuModel, "A100") && strings.Contains(gpuModel, "PCIE"):
		pricePerGPU = 0.80 // A100 40GB PCIe
	case strings.Contains(gpuModel, "A100"):
		pricePerGPU = 1.10 // Default A100 pricing

	// NVIDIA L40S - Ada Lovelace datacenter
	case strings.Contains(gpuModel, "L40S"):
		pricePerGPU = 1.50 // L40S 48GB

	// NVIDIA L40 - Ada Lovelace datacenter
	case strings.Contains(gpuModel, "L40"):
		pricePerGPU = 1.29 // L40 48GB

	// NVIDIA A10 - Ampere inference GPU
	case strings.Contains(gpuModel, "A10G"):
		pricePerGPU = 0.60 // A10G 24GB
	case strings.Contains(gpuModel, "A10"):
		pricePerGPU = 0.60 // A10 24GB

	// NVIDIA RTX 6000 Ada - Workstation GPU
	case strings.Contains(gpuModel, "RTX 6000 ADA") || strings.Contains(gpuModel, "RTX6000 ADA"):
		pricePerGPU = 0.80 // RTX 6000 Ada 48GB

	// NVIDIA RTX A6000 - Previous gen workstation
	case strings.Contains(gpuModel, "RTX A6000") || strings.Contains(gpuModel, "A6000"):
		pricePerGPU = 0.50 // RTX A6000 48GB

	// NVIDIA RTX A4000 - Mid-range workstation
	case strings.Contains(gpuModel, "RTX A4000") || strings.Contains(gpuModel, "A4000"):
		pricePerGPU = 0.20 // RTX A4000 16GB

	// NVIDIA V100 - Older datacenter GPU
	case strings.Contains(gpuModel, "V100"):
		pricePerGPU = 0.80 // V100 16GB/32GB

	// NVIDIA Tesla T4 - Turing inference GPU
	case strings.Contains(gpuModel, "T4"):
		pricePerGPU = 0.50 // T4 16GB

	// Consumer/Gaming GPUs (if detected)
	case strings.Contains(gpuModel, "RTX 4090"):
		pricePerGPU = 0.50 // RTX 4090 24GB
	case strings.Contains(gpuModel, "RTX 3090"):
		pricePerGPU = 0.40 // RTX 3090 24GB

	default:
		// Unknown GPU - try to make educated guess based on VRAM
		if len(gpus) > 0 {
			var memTotal int
			fmt.Sscanf(gpus[0].MemoryTotal, "%d", &memTotal)

			// Rough estimate based on memory size
			if memTotal >= 80000 { // 80GB+
				pricePerGPU = 2.00 // High-end datacenter
			} else if memTotal >= 40000 { // 40GB+
				pricePerGPU = 1.00 // Mid-range datacenter
			} else if memTotal >= 24000 { // 24GB+
				pricePerGPU = 0.60 // Entry datacenter/workstation
			} else {
				pricePerGPU = 0.30 // Consumer/small workstation
			}
		} else {
			return 0.00
		}
	}

	// Calculate total cost for multi-GPU instances
	totalCost := pricePerGPU * float64(gpuCount)

	return totalCost
}

func collectMetrics(client *ssh.Client, cfg *config.Config) (*Metrics, error) {
	executor := &SSHExecutor{client: client}
	return collectMetricsWithExecutor(executor, cfg)
}

func collectMetricsWithExecutor(executor CommandExecutor, cfg *config.Config) (*Metrics, error) {
	metrics := &Metrics{}

	// Get GPU metrics
	gpuCmd := `nvidia-smi --query-gpu=index,name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits`
	gpuOutput, err := executor.RunCommand(gpuCmd)
	if err != nil {
		// nvidia-smi failed - probably no GPUs
		return metrics, nil
	}

	lines := strings.Split(strings.TrimSpace(gpuOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) >= 7 {
			var gpu GPUMetric
			fmt.Sscanf(fields[0], "%d", &gpu.Index)
			gpu.Name = strings.TrimSpace(fields[1])
			fmt.Sscanf(fields[2], "%d", &gpu.Utilization)
			gpu.MemoryUsed = strings.TrimSpace(fields[3]) + " MB"
			gpu.MemoryTotal = strings.TrimSpace(fields[4]) + " MB"
			fmt.Sscanf(fields[5], "%d", &gpu.Temperature)
			gpu.PowerDraw = strings.TrimSpace(fields[6]) + " W"
			metrics.GPUs = append(metrics.GPUs, gpu)
		}
	}

	// Get uptime
	uptimeCmd := `cat /proc/uptime | awk '{print $1}'`
	uptimeOutput, err := executor.RunCommand(uptimeCmd)
	if err == nil {
		var uptime float64
		fmt.Sscanf(strings.TrimSpace(uptimeOutput), "%f", &uptime)
		metrics.UptimeSeconds = int64(uptime)
	}

	// Always auto-detect cost based on GPU model and count
	// This ensures we get accurate pricing even if config is missing or outdated
	metrics.InstanceCost = detectInstanceCost(metrics.GPUs)

	// Allow config override if explicitly set (and higher than auto-detected)
	if server, err := cfg.GetServer("lambda"); err == nil && server.CostPerHour > metrics.InstanceCost {
		metrics.InstanceCost = server.CostPerHour
	}

	// Count output files
	metrics.Outputs = countOutputsWithExecutor(executor)

	return metrics, nil
}

func countOutputs(client *ssh.Client) OutputMetrics {
	executor := &SSHExecutor{client: client}
	return countOutputsWithExecutor(executor)
}

func countOutputsWithExecutor(executor CommandExecutor) OutputMetrics {
	outputs := OutputMetrics{}

	// Count images in ComfyUI output
	imageCmd := `find ~/ComfyUI/output -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" \) 2>/dev/null | wc -l`
	if imgOut, err := executor.RunCommand(imageCmd); err == nil {
		fmt.Sscanf(strings.TrimSpace(imgOut), "%d", &outputs.Images)
	}

	// Count videos
	videoCmd := `find ~ -type f \( -name "*.mp4" -o -name "*.avi" -o -name "*.mov" \) -path "*/output/*" 2>/dev/null | wc -l`
	if vidOut, err := executor.RunCommand(videoCmd); err == nil {
		fmt.Sscanf(strings.TrimSpace(vidOut), "%d", &outputs.Videos)
	}

	// Count model files
	modelCmd := `find ~ -type f \( -name "*.safetensors" -o -name "*.ckpt" -o -name "*.pth" \) -path "*/models/*" 2>/dev/null | wc -l`
	if modOut, err := executor.RunCommand(modelCmd); err == nil {
		fmt.Sscanf(strings.TrimSpace(modOut), "%d", &outputs.Models)
	}

	// Get total size
	sizeCmd := `du -sb ~/ComfyUI/output 2>/dev/null | awk '{print $1}'`
	if sizeOut, err := executor.RunCommand(sizeCmd); err == nil {
		var sizeBytes float64
		fmt.Sscanf(strings.TrimSpace(sizeOut), "%f", &sizeBytes)
		outputs.TotalSizeGB = sizeBytes / (1024 * 1024 * 1024)
	}

	return outputs
}

func displayMetrics(m *Metrics, host string) {
	// Header
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🖥️  Instance Overview"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  Host:          %s\n", theme.HighlightStyle.Render(host))

	// Display GPU configuration summary
	if len(m.GPUs) > 0 {
		gpuName := m.GPUs[0].Name
		// Simplify GPU name for display
		gpuName = strings.ReplaceAll(gpuName, "NVIDIA ", "")
		gpuName = strings.ReplaceAll(gpuName, " PCIe", "")
		fmt.Printf("  GPUs:          %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%dx %s", len(m.GPUs), gpuName)))
	}

	// Display cost rate with indication if auto-detected or configured
	if m.InstanceCost > 0 {
		// Calculate per-GPU cost for multi-GPU systems
		perGPUCost := m.InstanceCost
		if len(m.GPUs) > 1 {
			perGPUCost = m.InstanceCost / float64(len(m.GPUs))
			fmt.Printf("  Cost Rate:     %s %s\n",
				theme.HighlightStyle.Render(fmt.Sprintf("$%.2f/hour", m.InstanceCost)),
				theme.DimTextStyle.Render(fmt.Sprintf("($%.2f per GPU × %d GPUs)", perGPUCost, len(m.GPUs))))
		} else {
			fmt.Printf("  Cost Rate:     %s %s\n",
				theme.HighlightStyle.Render(fmt.Sprintf("$%.2f/hour", m.InstanceCost)),
				theme.DimTextStyle.Render("(auto-detected)"))
		}
	} else {
		fmt.Printf("  Cost Rate:     %s\n",
			theme.WarningStyle.Render("$0.00/hour - GPU type not recognized"))
		if len(m.GPUs) > 0 {
			fmt.Printf("                 %s\n",
				theme.DimTextStyle.Render(fmt.Sprintf("(Detected: %s - please report this)", m.GPUs[0].Name)))
		}
	}
	fmt.Println()

	// Uptime and cost
	hours := float64(m.UptimeSeconds) / 3600.0
	totalCost := hours * m.InstanceCost

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("⏱️  Runtime & Cost"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Format uptime
	uptimeDuration := time.Duration(m.UptimeSeconds) * time.Second
	days := int(uptimeDuration.Hours() / 24)
	hours24 := int(uptimeDuration.Hours()) % 24
	minutes := int(uptimeDuration.Minutes()) % 60

	uptimeStr := ""
	if days > 0 {
		uptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours24, minutes)
	} else if hours24 > 0 {
		uptimeStr = fmt.Sprintf("%dh %dm", hours24, minutes)
	} else {
		uptimeStr = fmt.Sprintf("%dm", minutes)
	}

	fmt.Printf("  Uptime:        %s\n", theme.HighlightStyle.Render(uptimeStr))
	fmt.Printf("  Runtime Hours: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.2f hours", hours)))
	fmt.Printf("  Total Cost:    %s\n", theme.SuccessStyle.Render(fmt.Sprintf("$%.2f", totalCost)))
	fmt.Println()

	// GPU metrics
	if len(m.GPUs) > 0 {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("🎮 GPU Metrics"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		for _, gpu := range m.GPUs {
			fmt.Printf("  GPU %d: %s\n", gpu.Index, theme.HighlightStyle.Render(gpu.Name))
			fmt.Println()

			// Utilization bar
			utilBar := createBar(gpu.Utilization, 100, 30)
			utilColor := theme.SuccessStyle
			if gpu.Utilization > 80 {
				utilColor = theme.WarningStyle
			} else if gpu.Utilization < 20 {
				utilColor = theme.DimTextStyle
			}
			fmt.Printf("    Utilization:  %s %s\n",
				utilBar,
				utilColor.Render(fmt.Sprintf("%d%%", gpu.Utilization)))

			// Memory bar
			var memUsed, memTotal int
			fmt.Sscanf(gpu.MemoryUsed, "%d", &memUsed)
			fmt.Sscanf(gpu.MemoryTotal, "%d", &memTotal)
			memPercent := 0
			if memTotal > 0 {
				memPercent = (memUsed * 100) / memTotal
			}
			memBar := createBar(memPercent, 100, 30)
			memColor := theme.InfoStyle
			if memPercent > 90 {
				memColor = theme.ErrorStyle
			}
			fmt.Printf("    Memory:       %s %s\n",
				memBar,
				memColor.Render(fmt.Sprintf("%s / %s (%d%%)", gpu.MemoryUsed, gpu.MemoryTotal, memPercent)))

			// Temperature
			tempColor := theme.SuccessStyle
			if gpu.Temperature > 80 {
				tempColor = theme.ErrorStyle
			} else if gpu.Temperature > 70 {
				tempColor = theme.WarningStyle
			}
			fmt.Printf("    Temperature:  %s\n", tempColor.Render(fmt.Sprintf("%d°C", gpu.Temperature)))

			// Power
			fmt.Printf("    Power Draw:   %s\n", theme.DimTextStyle.Render(gpu.PowerDraw))
			fmt.Println()
		}
	}

	// Output metrics
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎨 Production Output"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  Images:        %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d files", m.Outputs.Images)))
	fmt.Printf("  Videos:        %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d files", m.Outputs.Videos)))
	fmt.Printf("  Models:        %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d files", m.Outputs.Models)))
	fmt.Printf("  Total Size:    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.2f GB", m.Outputs.TotalSizeGB)))
	fmt.Println()

	// Efficiency metrics
	if hours > 0 {
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("⚡ Efficiency"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		totalOutputs := m.Outputs.Images + m.Outputs.Videos
		if totalOutputs > 0 {
			costPerOutput := totalCost / float64(totalOutputs)
			outputsPerHour := float64(totalOutputs) / hours

			fmt.Printf("  Output/Hour:   %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.1f files", outputsPerHour)))
			fmt.Printf("  Cost/Output:   %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$%.3f", costPerOutput)))
		} else {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render("No outputs generated yet"))
		}
		fmt.Println()
	}

	// Footer
	fmt.Println(theme.DimTextStyle.Render("  💡 Tip: Metrics update in real-time. Run 'anime metrics' anytime!"))
	fmt.Println()
}

func createBar(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}

	filled := (value * width) / max
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	// Color based on percentage
	percent := (value * 100) / max
	if percent > 80 {
		return theme.WarningStyle.Render(bar)
	} else if percent > 50 {
		return theme.SuccessStyle.Render(bar)
	} else {
		return theme.DimTextStyle.Render(bar)
	}
}
