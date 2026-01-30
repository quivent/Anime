package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	lambdaParallel bool
	lambdaJobs     int
)

var lambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "Lambda server operations",
	Long:  "Commands for managing Lambda servers and installing modules.",
	Run:   runLambdaHelp,
}

var lambdaDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Show default packages for new Lambda GPU instances",
	Long:  "Display recommended default packages to install on a fresh Lambda GPU instance",
	Run:   runLambdaDefaults,
}

var lambdaInstallCmd = &cobra.Command{
	Use:   "install [modules...]",
	Short: "Install modules on lambda server",
	Long: `Install one or more modules on the configured lambda server.

Modules are installed with dependencies resolved automatically.
Use --parallel to enable concurrent installation of independent modules.

Examples:
  anime lambda install python            # Install single module
  anime lambda install pytorch ollama    # Install multiple modules
  anime lambda install --parallel        # Install all configured modules in parallel
  anime lambda install python --jobs 8   # Use 8 parallel jobs for compilation`,
	RunE: runLambdaInstall,
}

var lambdaLaunchCmd = &cobra.Command{
	Use:   "launch <config>",
	Short: "Launch predefined configurations on Lambda server",
	Long: `Launch web servers and services with predefined configurations.

Available configurations:
  comfyui    - Launch ComfyUI web interface (port 8188)
  ollama     - Start Ollama server (port 11434)
  jupyter    - Launch JupyterLab (random port)

Examples:
  anime lambda launch comfyui            # Launch ComfyUI web server
  anime lambda launch ollama             # Start Ollama server
  anime lambda launch jupyter            # Start JupyterLab`,
	RunE: runLambdaLaunch,
}

func init() {
	lambdaInstallCmd.Flags().BoolVarP(&lambdaParallel, "parallel", "p", true, "Install independent modules in parallel (default true)")
	lambdaInstallCmd.Flags().IntVarP(&lambdaJobs, "jobs", "j", 0, "Number of parallel compilation jobs (0 = auto-detect all cores)")

	lambdaCmd.AddCommand(lambdaDefaultsCmd)
	lambdaCmd.AddCommand(lambdaInstallCmd)
	lambdaCmd.AddCommand(lambdaLaunchCmd)
	rootCmd.AddCommand(lambdaCmd)
}

func runLambdaHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ LAMBDA GPU MANAGEMENT ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Quick-start guide for Lambda GPU instances"))
	fmt.Println()

	// Check if lambda server is configured
	cfg, err := config.Load()
	lambdaConfigured := false
	var lambdaTarget string
	if err == nil {
		lambdaTarget = cfg.GetAlias("lambda")
		if lambdaTarget != "" {
			lambdaConfigured = true
		} else if _, err := cfg.GetServer("lambda"); err == nil {
			lambdaConfigured = true
		}
	}

	// Show configuration status
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📡 Server Status"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	if lambdaConfigured {
		fmt.Printf("  Lambda Server:  %s\n", theme.SuccessStyle.Render("✓ Configured"))
		if lambdaTarget != "" {
			fmt.Printf("  Target:         %s\n", theme.HighlightStyle.Render(lambdaTarget))
		}
	} else {
		fmt.Printf("  Lambda Server:  %s\n", theme.WarningStyle.Render("◯ Not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 To configure your Lambda server:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime set lambda <server-ip>"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Example:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime set lambda 209.20.159.132"))
	}
	fmt.Println()

	// Available commands
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime lambda defaults", "View recommended default packages with installation status"},
		{"anime lambda install <package>", "Install packages on your Lambda server"},
		{"anime lambda install --parallel", "Install all configured modules in parallel"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	// Quick start workflow
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Typical Workflow"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	steps := []struct {
		num  string
		cmd  string
		desc string
	}{
		{"1", "anime set lambda <ip>", "Configure your Lambda server"},
		{"2", "anime lambda defaults", "View recommended packages"},
		{"3", "anime install core python pytorch...", "Install packages"},
		{"4", "anime packages", "Browse all available packages"},
	}

	for _, s := range steps {
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(s.num+"."),
			theme.HighlightStyle.Render(s.cmd))
		fmt.Printf("     %s\n", theme.DimTextStyle.Render(s.desc))
		fmt.Println()
	}

	// Next step
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Next Step"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if lambdaConfigured {
		fmt.Println(theme.GlowStyle.Render("  View recommended packages for your Lambda server:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime lambda defaults"))
	} else {
		fmt.Println(theme.GlowStyle.Render("  Configure your Lambda server to get started:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime set lambda <your-server-ip>"))
	}
	fmt.Println()
}

func runLambdaDefaults(cmd *cobra.Command, args []string) {
	// Anime-style banner
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ LAMBDA GPU DEFAULTS ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📦 Recommended starter pack for new Lambda GPU instances"))
	fmt.Println(theme.DimTextStyle.Render("   A curated selection to get you up and running with AI development"))
	fmt.Println()

	// Define default packages for a new Lambda GPU instance
	defaultPackages := []string{
		"core",
		"python",
		"pytorch",
		"ollama",
		"models-small",
		"claude",
	}

	packages := installer.GetPackages()

	// Check installation status if lambda server is configured
	installedPackages := make(map[string]bool)
	cfg, err := config.Load()
	connectedToLambda := false
	if err == nil {
		// Try to get lambda server
		lambdaTarget := cfg.GetAlias("lambda")
		_, hasLambdaServer := cfg.GetServer("lambda")
		if lambdaTarget != "" || hasLambdaServer == nil {
			var user, host string
			if lambdaTarget != "" {
				if strings.Contains(lambdaTarget, "@") {
					parts := strings.SplitN(lambdaTarget, "@", 2)
					user = parts[0]
					host = parts[1]
				} else {
					user = "ubuntu"
					host = lambdaTarget
				}
			} else if server, err := cfg.GetServer("lambda"); err == nil {
				user = server.User
				host = server.Host
			}

			// Try to check installation status
			if host != "" {
				sshClient, err := ssh.NewClient(host, user, "")
				if err == nil {
					defer sshClient.Close()
					connectedToLambda = true
					// Check each package
					for _, pkgID := range defaultPackages {
						if checkPackageInstalled(sshClient, pkgID) {
							installedPackages[pkgID] = true
						}
					}
				}
			}
		}
	}

	// Group packages by category for clearer display
	type categoryGroup struct {
		emoji    string
		title    string
		packages []string
	}

	categories := []categoryGroup{
		{"🏗️", "Foundation", []string{"core", "python"}},
		{"🤖", "ML Framework", []string{"pytorch"}},
		{"🔮", "LLM Runtime", []string{"ollama"}},
		{"⭐", "Models", []string{"models-small"}},
		{"🎯", "Developer Tools", []string{"claude"}},
	}

	// Count installed vs needed
	totalInstalled := 0
	totalNeeded := len(defaultPackages)

	// Display packages by category
	for _, cat := range categories {
		fmt.Printf("  %s %s\n", cat.emoji, theme.SuccessStyle.Render(cat.title))
		fmt.Println(theme.DimTextStyle.Render("  " + strings.Repeat("─", 60)))

		for _, pkgID := range cat.packages {
			pkg, exists := packages[pkgID]
			if !exists {
				continue
			}

			// Check if installed
			isInstalled := installedPackages[pkgID]
			if isInstalled {
				totalInstalled++
			}

			// Status badge
			var statusBadge string
			if connectedToLambda {
				if isInstalled {
					statusBadge = theme.SuccessStyle.Render(" ✓ INSTALLED")
				} else {
					statusBadge = theme.WarningStyle.Render(" ◯ NOT INSTALLED")
				}
			}

			// Package name with ID and status
			fmt.Printf("  %s %s %s%s\n",
				theme.SymbolSparkle,
				theme.HighlightStyle.Render(pkg.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", pkg.ID)),
				statusBadge)

			// Description with better formatting
			fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(pkg.Description))

			// Metadata in a clean row
			if !isInstalled || !connectedToLambda {
				fmt.Printf("    %s  %s\n",
					theme.DimTextStyle.Render("⏱️  "+pkg.EstimatedTime.String()),
					theme.DimTextStyle.Render("💾 "+pkg.Size))
			} else {
				fmt.Printf("    %s\n",
					theme.DimTextStyle.Render("Already installed on your Lambda server"))
			}
			fmt.Println()
		}
	}

	// Installation status summary
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Installation Status"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	if connectedToLambda {
		fmt.Printf("  Installed:        %s / %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d", totalInstalled)),
			theme.HighlightStyle.Render(fmt.Sprintf("%d", totalNeeded)))

		if totalInstalled < totalNeeded {
			fmt.Printf("  Remaining time:   %s\n", theme.HighlightStyle.Render("~55 minutes"))
			fmt.Printf("  Remaining space:  %s\n", theme.HighlightStyle.Render("~28GB"))
		} else {
			fmt.Printf("  %s\n", theme.SuccessStyle.Render("✨ All default packages are installed!"))
		}
	} else {
		fmt.Printf("  Total packages:   %s\n", theme.HighlightStyle.Render("6"))
		fmt.Printf("  Estimated time:   %s\n", theme.HighlightStyle.Render("~55 minutes"))
		fmt.Printf("  Total disk space: %s\n", theme.HighlightStyle.Render("~28GB"))
		fmt.Printf("\n  %s\n", theme.DimTextStyle.Render("💡 Configure Lambda server to see installation status"))
	}
	fmt.Println()

	// Show installation commands
	if !connectedToLambda || totalInstalled < totalNeeded {
		// Build install command with only missing packages
		var missingPackages []string
		for _, pkgID := range defaultPackages {
			if !installedPackages[pkgID] {
				missingPackages = append(missingPackages, pkgID)
			}
		}

		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(theme.InfoStyle.Render("✨ Installation Commands"))
		fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()

		if connectedToLambda && len(missingPackages) > 0 && len(missingPackages) < len(defaultPackages) {
			fmt.Println(theme.GlowStyle.Render("  Install Missing Packages Only"))
			installCmd := "anime install " + strings.Join(missingPackages, " ")
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ "+installCmd))
			fmt.Println()
		}

		fmt.Println(theme.GlowStyle.Render("  Install All Default Packages"))
		fullInstallCmd := "anime install " + strings.Join(defaultPackages, " ")
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ "+fullInstallCmd))
		fmt.Println()

		fmt.Println(theme.GlowStyle.Render("  Interactive Selection"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime interactive"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  (Choose packages with a beautiful TUI interface)"))
		fmt.Println()
	}

	// Additional helpful commands
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Additional Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime packages"),
		theme.DimTextStyle.Render("Browse all available packages"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime packages --tree"),
		theme.DimTextStyle.Render("View dependency tree"))
	fmt.Printf("  %s - %s\n",
		theme.HighlightStyle.Render("anime install --help"),
		theme.DimTextStyle.Render("See all installation options"))
	fmt.Println()
}

func runLambdaInstall(cmd *cobra.Command, args []string) error {
	// Load config to get lambda server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get lambda alias or find a server named lambda
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		// Try to find a server named "lambda"
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	if lambdaTarget == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Lambda server not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Set it up with:"))
		fmt.Println(theme.HighlightStyle.Render("  anime set lambda <server-ip>"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Example:"))
		fmt.Println(theme.HighlightStyle.Render("  anime set lambda 209.20.159.132"))
		fmt.Println()
		return fmt.Errorf("lambda server not configured")
	}

	// Determine which modules to install
	var modules []string
	if len(args) > 0 {
		modules = args
	} else {
		// If no modules specified, look for configured modules on lambda server
		if server, err := cfg.GetServer("lambda"); err == nil && len(server.Modules) > 0 {
			modules = server.Modules
		} else {
			return fmt.Errorf("no modules specified. Usage: anime lambda install <module> [module...]")
		}
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Installing modules on Lambda server"))
	fmt.Println()
	fmt.Printf("  Target:   %s\n", theme.HighlightStyle.Render(lambdaTarget))
	fmt.Printf("  Modules:  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%v", modules)))

	modeStr := "Sequential"
	if lambdaParallel {
		modeStr = "Parallel (auto-optimized)"
	}
	fmt.Printf("  Mode:     %s\n", theme.SuccessStyle.Render(modeStr))

	jobsStr := "Auto-detect all cores"
	if lambdaJobs > 0 {
		jobsStr = fmt.Sprintf("%d cores", lambdaJobs)
	}
	fmt.Printf("  Jobs:     %s\n", theme.HighlightStyle.Render(jobsStr))
	fmt.Println()

	// Parse target to get user and host
	var user, host string
	if strings.Contains(lambdaTarget, "@") {
		parts := strings.SplitN(lambdaTarget, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = lambdaTarget
	}

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect to lambda server: %w", err)
	}
	defer sshClient.Close()

	// Create installer
	inst := installer.New(sshClient)

	// Set parallel mode and jobs if specified
	if lambdaParallel {
		inst.SetParallel(true)
	}
	if lambdaJobs > 0 {
		inst.SetJobs(lambdaJobs)
	}

	// Create server config for TUI
	server := &config.Server{
		Name:    "lambda",
		Host:    host,
		User:    user,
		Modules: modules,
	}

	// Use TUI for installation
	m := tui.NewInstallModel(server)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}

func runLambdaLaunch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		showLaunchHelp()
		return nil
	}

	configName := args[0]

	// Load config to get lambda server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get lambda target
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	if lambdaTarget == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Lambda server not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Set it up with:"))
		fmt.Println(theme.HighlightStyle.Render("  anime set lambda <server-ip>"))
		fmt.Println()
		return fmt.Errorf("lambda server not configured")
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
	fmt.Println(theme.RenderBanner("🚀 LAUNCHING " + strings.ToUpper(configName) + " 🚀"))
	fmt.Println()

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect to lambda server: %w", err)
	}
	defer sshClient.Close()

	// Get launch command based on config
	launchCmd := getLaunchCommand(configName)
	if launchCmd == "" {
		return fmt.Errorf("unknown configuration: %s", configName)
	}

	fmt.Printf("  Server:   %s\n", theme.HighlightStyle.Render(lambdaTarget))
	fmt.Printf("  Config:   %s\n", theme.HighlightStyle.Render(configName))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Starting service..."))
	fmt.Println()

	// Execute launch command
	output, err := sshClient.RunCommand(launchCmd)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("  ❌ Failed to launch"))
		fmt.Println(theme.DimTextStyle.Render(output))
		return fmt.Errorf("launch failed: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ Service launched successfully!"))
	fmt.Println()

	// Show access information
	showAccessInfo(configName, host)

	return nil
}

func getLaunchCommand(configName string) string {
	commands := map[string]string{
		"comfyui": `cd ~/ComfyUI && nohup python main.py --listen 0.0.0.0 --port 8188 > ~/comfyui.log 2>&1 & echo $! > ~/comfyui.pid && echo "ComfyUI started on port 8188"`,
		"ollama":  `sudo systemctl start ollama && echo "Ollama server started on port 11434"`,
		"jupyter": `cd ~ && PORT=$((10000 + RANDOM % 10000)) && nohup jupyter lab --ip=0.0.0.0 --port=$PORT --no-browser --allow-root > ~/jupyter.log 2>&1 & echo $! > ~/jupyter.pid && echo "JupyterLab started on port $PORT" && cat ~/jupyter.log | grep "http.*token"`,
	}
	return commands[configName]
}

func showAccessInfo(configName string, host string) {
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🌐 Access Information"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	switch configName {
	case "comfyui":
		fmt.Printf("  Web Interface:  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("http://%s:8188", host)))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  💡 Tip: Open this URL in your browser to access ComfyUI"))
		fmt.Println(theme.DimTextStyle.Render("      You can generate images and videos from the web interface"))

	case "ollama":
		fmt.Printf("  API Endpoint:   %s\n", theme.HighlightStyle.Render(fmt.Sprintf("http://%s:11434", host)))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  💡 Tip: Use this endpoint with Ollama client or API"))
		fmt.Println(theme.DimTextStyle.Render("      Example: curl http://"+host+":11434/api/generate -d '{...}'"))

	case "jupyter":
		fmt.Println(theme.DimTextStyle.Render("  💡 Check ~/jupyter.log for the access token and port"))
		fmt.Println(theme.DimTextStyle.Render("      ssh to the server and run: cat ~/jupyter.log | grep token"))
	}
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🎯 Management Commands"))
	fmt.Printf("  Stop service:   %s\n", theme.HighlightStyle.Render(fmt.Sprintf("kill $(cat ~/%s.pid)", configName)))
	fmt.Printf("  View logs:      %s\n", theme.HighlightStyle.Render(fmt.Sprintf("tail -f ~/%s.log", configName)))
	fmt.Println()
}

func showLaunchHelp() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ LAMBDA LAUNCH ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🚀 Launch predefined configurations on your Lambda server"))
	fmt.Println()

	configs := []struct {
		name string
		port string
		desc string
	}{
		{"comfyui", "8188", "ComfyUI web interface for image/video generation"},
		{"ollama", "11434", "Ollama LLM server API"},
		{"jupyter", "random", "JupyterLab notebook environment"},
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Available Configurations"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for _, c := range configs {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.name))
		fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(c.desc))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render("Port: "+c.port))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Examples"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime lambda launch comfyui"))
	fmt.Println(theme.DimTextStyle.Render("    Launch ComfyUI web server"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime lambda launch ollama"))
	fmt.Println(theme.DimTextStyle.Render("    Start Ollama LLM server"))
	fmt.Println()
}
