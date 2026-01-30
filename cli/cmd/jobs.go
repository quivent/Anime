package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	jobsServer string
	jobsAll    bool
)

var jobsCmd = &cobra.Command{
	Use:   "jobs [server]",
	Short: "Show running jobs and services",
	Long: `Display running jobs, services, and background processes.

Shows:
  • HTTP servers (anime serve)
  • Installation processes
  • Running services (ollama, comfyui)
  • Model downloads
  • Background tasks

Examples:
  anime jobs                    # Show jobs on local machine
  anime jobs lambda             # Show jobs on lambda server
  anime jobs 192.168.1.100      # Show jobs on specific server
  anime jobs --all              # Show all processes, not just anime-related`,
	RunE: runJobs,
}

func init() {
	jobsCmd.Flags().BoolVarP(&jobsAll, "all", "a", false, "Show all processes, not just anime-related")
	rootCmd.AddCommand(jobsCmd)
}

func runJobs(cmd *cobra.Command, args []string) error {
	// If no server specified, show local jobs
	if len(args) == 0 {
		return runJobsLocal()
	}

	server := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get target
	var target string
	if strings.Contains(server, "@") {
		target = server
	} else if strings.Contains(server, ".") {
		target = "ubuntu@" + server
	} else {
		// Try alias first
		target = cfg.GetAlias(server)
		if target == "" {
			// Try server name
			if s, err := cfg.GetServer(server); err == nil {
				target = fmt.Sprintf("%s@%s", s.User, s.Host)
			}
		}
	}

	if target == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Server not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Configure a server first:"))
		fmt.Println(theme.HighlightStyle.Render("  anime push lambda"))
		fmt.Println()
		return fmt.Errorf("server not configured")
	}

	// Parse target to get user and host
	parts := strings.Split(target, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", target)
	}
	user := parts[0]
	host := parts[1]

	fmt.Println()
	fmt.Println(theme.RenderBanner("⚙️  ANIME JOBS ⚙️"))
	fmt.Println()
	fmt.Printf("  Server: %s\n", theme.HighlightStyle.Render(target))
	fmt.Println()

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer sshClient.Close()

	// Check for running jobs
	jobs := []Job{}

	// 1. Check for serve processes
	serveJobs, err := checkServeJobs(sshClient)
	if err == nil {
		jobs = append(jobs, serveJobs...)
	}

	// 2. Check for ollama
	ollamaJobs, err := checkOllamaJobs(sshClient)
	if err == nil {
		jobs = append(jobs, ollamaJobs...)
	}

	// 3. Check for ComfyUI
	comfyJobs, err := checkComfyUIJobs(sshClient)
	if err == nil {
		jobs = append(jobs, comfyJobs...)
	}

	// 4. Check for installations/downloads
	installJobs, err := checkInstallationJobs(sshClient)
	if err == nil {
		jobs = append(jobs, installJobs...)
	}

	// Display results
	if len(jobs) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No running jobs found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Start a job with:"))
		fmt.Println(theme.HighlightStyle.Render("    anime serve ~/outputs"))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📋 Running Jobs"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, job := range jobs {
		displayJob(job)
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Total: %d running job(s)", len(jobs))))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	return nil
}

type Job struct {
	Type    string // serve, ollama, comfyui, install, download
	Name    string
	PID     string
	Port    string
	URL     string
	Status  string
	Details string
}

func checkServeJobs(client *ssh.Client) ([]Job, error) {
	jobs := []Job{}

	// Check for serve.pid
	output, err := client.RunCommand("if [ -f ~/serve.pid ]; then cat ~/serve.pid; fi")
	if err != nil || strings.TrimSpace(output) == "" {
		return jobs, nil
	}

	pid := strings.TrimSpace(output)

	// Check if process is still running
	checkCmd := fmt.Sprintf("ps -p %s -o pid,args --no-headers 2>/dev/null", pid)
	psOutput, err := client.RunCommand(checkCmd)
	if err != nil || strings.TrimSpace(psOutput) == "" {
		return jobs, nil
	}

	// Extract port from command line
	port := ""
	if strings.Contains(psOutput, "http.server") {
		parts := strings.Fields(psOutput)
		for i, part := range parts {
			if part == "http.server" && i+1 < len(parts) {
				port = parts[i+1]
				break
			}
		}
	}

	// Get host from client
	host := client.Host()

	job := Job{
		Type:   "serve",
		Name:   "HTTP Server",
		PID:    pid,
		Port:   port,
		Status: "running",
	}

	if port != "" {
		job.URL = fmt.Sprintf("http://%s:%s", host, port)
		job.Details = fmt.Sprintf("Serving on port %s", port)
	}

	jobs = append(jobs, job)
	return jobs, nil
}

func checkOllamaJobs(client *ssh.Client) ([]Job, error) {
	jobs := []Job{}

	// Check if ollama is running
	output, err := client.RunCommand("pgrep -f 'ollama serve' 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return jobs, nil
	}

	pid := strings.TrimSpace(strings.Split(output, "\n")[0])

	job := Job{
		Type:    "ollama",
		Name:    "Ollama Server",
		PID:     pid,
		Port:    "11434",
		Status:  "running",
		URL:     fmt.Sprintf("http://%s:11434", client.Host()),
		Details: "LLM inference server",
	}

	jobs = append(jobs, job)
	return jobs, nil
}

func checkComfyUIJobs(client *ssh.Client) ([]Job, error) {
	jobs := []Job{}

	// Check if ComfyUI is running
	output, err := client.RunCommand("pgrep -f 'python.*main.py.*ComfyUI' 2>/dev/null")
	if err != nil || strings.TrimSpace(output) == "" {
		return jobs, nil
	}

	pid := strings.TrimSpace(strings.Split(output, "\n")[0])

	job := Job{
		Type:    "comfyui",
		Name:    "ComfyUI",
		PID:     pid,
		Port:    "8188",
		Status:  "running",
		URL:     fmt.Sprintf("http://%s:8188", client.Host()),
		Details: "Stable Diffusion UI",
	}

	jobs = append(jobs, job)
	return jobs, nil
}

func checkInstallationJobs(client *ssh.Client) ([]Job, error) {
	jobs := []Job{}

	// Check for apt/dpkg processes
	output, err := client.RunCommand("pgrep -af 'apt-get|dpkg|pip install|git clone|wget|curl.*download' 2>/dev/null | head -5")
	if err != nil || strings.TrimSpace(output) == "" {
		return jobs, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid := fields[0]
		cmdLine := strings.Join(fields[1:], " ")

		// Truncate long command lines
		if len(cmdLine) > 60 {
			cmdLine = cmdLine[:57] + "..."
		}

		jobType := "install"
		jobName := "Installation"

		if strings.Contains(cmdLine, "pip install") {
			jobName = "Python Package Install"
		} else if strings.Contains(cmdLine, "git clone") {
			jobName = "Git Clone"
			jobType = "download"
		} else if strings.Contains(cmdLine, "wget") || strings.Contains(cmdLine, "curl") {
			jobName = "Download"
			jobType = "download"
		} else if strings.Contains(cmdLine, "apt-get") {
			jobName = "System Package Install"
		}

		job := Job{
			Type:    jobType,
			Name:    jobName,
			PID:     pid,
			Status:  "running",
			Details: cmdLine,
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func displayJob(job Job) {
	// Icon based on type
	icon := "⚙️"
	switch job.Type {
	case "serve":
		icon = "🌐"
	case "ollama":
		icon = "🤖"
	case "comfyui":
		icon = "🎨"
	case "download":
		icon = "📥"
	case "install":
		icon = "📦"
	}

	fmt.Printf("  %s %s\n", icon, theme.SuccessStyle.Render(job.Name))
	fmt.Printf("     PID:    %s\n", theme.HighlightStyle.Render(job.PID))

	if job.Port != "" {
		fmt.Printf("     Port:   %s\n", theme.HighlightStyle.Render(job.Port))
	}

	if job.URL != "" {
		fmt.Printf("     URL:    %s\n", theme.HighlightStyle.Render(job.URL))
	}

	if job.Details != "" {
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(job.Details))
	}

	fmt.Println()
}

// runJobsLocal shows jobs on the local machine (used when SSH fails or for localhost)
func runJobsLocal() error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚙️  LOCAL JOBS ⚙️"))
	fmt.Println()

	jobs := []Job{}

	// Check for workflow/animation job PID files
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		// Look for workflow PID files: ~/workflow-*.pid
		pattern := filepath.Join(homeDir, "workflow-*.pid")
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, pidFile := range matches {
				job := checkLocalJobFile(pidFile, "workflow")
				if job != nil {
					jobs = append(jobs, *job)
				}
			}
		}

		// Look for animation PID files: ~/animation-*.pid
		pattern = filepath.Join(homeDir, "animation-*.pid")
		matches, err = filepath.Glob(pattern)
		if err == nil {
			for _, pidFile := range matches {
				job := checkLocalJobFile(pidFile, "animation")
				if job != nil {
					jobs = append(jobs, *job)
				}
			}
		}
	}

	// Check for serve jobs (HTTP server)
	if homeDir != "" {
		serveFile := filepath.Join(homeDir, "serve.pid")
		job := checkLocalJobFile(serveFile, "serve")
		if job != nil {
			jobs = append(jobs, *job)
		}
	}

	if len(jobs) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No running jobs found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Start jobs with:"))
		fmt.Println(theme.HighlightStyle.Render("    anime animate <collection>"))
		fmt.Println(theme.HighlightStyle.Render("    anime workflow <collection> <workflow>"))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.InfoStyle.Render("📋 Running Jobs:"))
	fmt.Println()

	for _, job := range jobs {
		displayJob(job)
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Total: %d running job(s)", len(jobs))))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	return nil
}

// checkLocalJobFile checks if a PID file exists and the process is still running
func checkLocalJobFile(pidFile string, jobType string) *Job {
	// Check if file exists
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return nil
	}

	pid := strings.TrimSpace(string(data))
	if pid == "" {
		return nil
	}

	// Check if process is still running
	cmd := exec.Command("ps", "-p", pid, "-o", "pid,command")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		// Process not running, clean up PID file
		os.Remove(pidFile)
		return nil
	}

	// Parse the job ID from filename
	baseName := filepath.Base(pidFile)
	jobID := strings.TrimSuffix(baseName, ".pid")

	// Get log file path if it exists
	logFile := strings.TrimSuffix(pidFile, ".pid") + ".log"
	logExists := false
	if _, err := os.Stat(logFile); err == nil {
		logExists = true
	}

	job := &Job{
		Type:   jobType,
		PID:    pid,
		Status: "running",
	}

	switch jobType {
	case "workflow":
		job.Name = fmt.Sprintf("Workflow: %s", strings.TrimPrefix(jobID, "workflow-"))
		if logExists {
			job.Details = fmt.Sprintf("Log: tail -f %s", logFile)
		}
	case "animation":
		job.Name = fmt.Sprintf("Animation: %s", strings.TrimPrefix(jobID, "animation-"))
		if logExists {
			job.Details = fmt.Sprintf("Log: tail -f %s", logFile)
		}
	case "serve":
		job.Name = "HTTP Server"
		// Try to extract port from command line
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			if strings.Contains(lines[1], "http.server") {
				fields := strings.Fields(lines[1])
				for i, field := range fields {
					if field == "http.server" && i+1 < len(fields) {
						job.Port = fields[i+1]
						job.URL = fmt.Sprintf("http://localhost:%s", fields[i+1])
						break
					}
				}
			}
		}
	}

	return job
}
