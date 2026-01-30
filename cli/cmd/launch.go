package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/launch"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var launchLogLines int

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch and manage web applications",
	Long: `Full-stack app launcher: detect, deploy, proxy, secure, and manage.

SUBCOMMANDS:
  setup       Interactive wizard to launch a web app
  status      Show status of launched apps
  stop        Stop a launched app
  logs        View logs for a launched app
  list        List all launched apps

EXAMPLES:
  anime launch                          # Start setup wizard
  anime launch setup ./myapp            # Launch app at path
  anime launch status                   # Show all running apps
  anime launch stop myapp               # Stop an app
  anime launch logs myapp -n 100        # View last 100 lines`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLaunchSetup(cmd, args)
	},
}

var launchSetupCmd = &cobra.Command{
	Use:   "setup [path]",
	Short: "Interactive wizard to launch a web application",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLaunchSetup,
}

var launchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of launched applications",
	RunE:  runLaunchStatus,
}

var launchStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a launched application",
	Args:  cobra.ExactArgs(1),
	RunE:  runLaunchStop,
}

var launchLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show logs for a launched application",
	Args:  cobra.ExactArgs(1),
	RunE:  runLaunchLogs,
}

var launchListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all launched applications",
	RunE:    runLaunchList,
}

func init() {
	launchCmd.AddCommand(launchSetupCmd)
	launchCmd.AddCommand(launchStatusCmd)
	launchCmd.AddCommand(launchStopCmd)
	launchCmd.AddCommand(launchLogsCmd)
	launchCmd.AddCommand(launchListCmd)

	launchLogsCmd.Flags().IntVarP(&launchLogLines, "lines", "n", 50, "Number of log lines")

	rootCmd.AddCommand(launchCmd)
}

// ── Setup Wizard ──────────────────────────────────────────────────────

func runLaunchSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("LAUNCH"))
	fmt.Println()

	// ── Step 1: Detect project type ────────────────────────────────
	printLaunchStep(1, 10, "Detecting project")

	analysis := launch.AnalyzeProject(absPath)
	detected := analysis.Project

	if detected.Type == launch.ProjectUnknown {
		fmt.Println(theme.ErrorStyle.Render("  No project detected"))
		fmt.Println()
		for _, issue := range analysis.Issues {
			fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("✗"), issue.Message)
			if issue.Detail != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(issue.Detail))
			}
			if issue.Fix != "" {
				fmt.Printf("    %s %s\n", theme.HighlightStyle.Render("Fix:"), issue.Fix)
			}
		}
		fmt.Println()
		return fmt.Errorf("cannot launch: no project found at %s", absPath)
	}

	// Show what was detected
	fmt.Printf("  %s %s (%s)\n", theme.SuccessStyle.Render("Detected:"), theme.HighlightStyle.Render(string(detected.Type)), detected.Framework)
	if detected.EntryPoint != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Entry:"), detected.EntryPoint)
	}
	if detected.PackageManager != "" && detected.PackageManager != "npm" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Package manager:"), detected.PackageManager)
	}
	if detected.ProcfileUsed {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Procfile:"), "using web process")
	}
	if detected.Type == launch.ProjectStatic || detected.Type == launch.ProjectBuild {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Note:"), "no server found — generated serving layer")
	}

	// Show issues and bail on errors
	// First pass: separate errors, warnings, and prompts
	hasErrors := false
	var prompts []launch.ProjectIssue
	if !analysis.Clean {
		fmt.Println()
		for _, issue := range analysis.Issues {
			if issue.Severity == "prompt" {
				prompts = append(prompts, issue)
				continue
			}
			if issue.Severity == "error" {
				fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("✗"), issue.Message)
				hasErrors = true
			} else {
				fmt.Printf("  %s %s\n", theme.WarningStyle.Render("!"), issue.Message)
			}
			if issue.Detail != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(issue.Detail))
			}
			if issue.Fix != "" {
				fmt.Printf("    %s %s\n", theme.HighlightStyle.Render("Fix:"), issue.Fix)
			}
		}
		if hasErrors {
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("  Fix the issues above and try again."))
			fmt.Println()
			return fmt.Errorf("project has issues that need to be resolved first")
		}
		// Warnings only — ask to continue
		if len(prompts) == 0 {
			fmt.Println()
			if !promptUserYesNo(reader, "  Continue with warnings", false) {
				return fmt.Errorf("aborted")
			}
		}
	}

	// Handle interactive prompts (ambiguous entry points, multiple binaries, etc.)
	for _, issue := range prompts {
		if len(issue.Choices) == 0 {
			continue
		}
		fmt.Printf("\n  %s %s\n", theme.WarningStyle.Render("?"), issue.Message)
		if issue.Detail != "" {
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(issue.Detail))
		}
		for i, c := range issue.Choices {
			fmt.Printf("    %d. %s\n", i+1, c)
		}
		choiceNums := make([]string, len(issue.Choices))
		for i := range issue.Choices {
			choiceNums[i] = strconv.Itoa(i + 1)
		}
		pick := promptUserChoice(reader, "    Select", choiceNums)
		idx, _ := strconv.Atoi(pick)
		if idx >= 1 && idx <= len(issue.Choices) {
			chosen := issue.Choices[idx-1]
			switch issue.ChoiceKey {
			case "binary":
				detected.RunCommand = "go run ./" + chosen
			case "entry_point":
				detected.EntryPoint = chosen
				detected.RunCommand = "python " + chosen
			case "run_command":
				detected.RunCommand = chosen
			}
		}
	}

	// If there were only warnings (no errors, but warnings shown), ask to continue
	if !analysis.Clean && !hasErrors && len(prompts) > 0 {
		fmt.Println()
		if !promptUserYesNo(reader, "  Continue", true) {
			return fmt.Errorf("aborted")
		}
	}
	fmt.Println()

	// ── Step 2: Run command ────────────────────────────────────────
	printLaunchStep(2, 10, "Run command")

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Suggested:"), theme.InfoStyle.Render(detected.RunCommand))
	fmt.Print(theme.HighlightStyle.Render("  Command (enter to accept) ▶ "))
	runCmdInput, _ := reader.ReadString('\n')
	runCmdInput = strings.TrimSpace(runCmdInput)
	if runCmdInput != "" {
		detected.RunCommand = runCmdInput
	}
	fmt.Println()

	// ── Step 3: Port ───────────────────────────────────────────────
	printLaunchStep(3, 10, "Application port")

	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Detected:"), detected.Port)
	fmt.Print(theme.HighlightStyle.Render("  Port (enter to accept) ▶ "))
	portInput, _ := reader.ReadString('\n')
	portInput = strings.TrimSpace(portInput)
	if portInput != "" {
		if p, err := strconv.Atoi(portInput); err == nil {
			detected.Port = p
		}
	}
	fmt.Println()

	// ── Step 4: Target server ──────────────────────────────────────
	printLaunchStep(4, 10, "Target server")

	fmt.Println("  1. Local (this machine)")
	fmt.Println("  2. Remote server")
	targetChoice := promptUserChoice(reader, "  Select target", []string{"1", "2"})

	var runner launch.CommandRunner
	var serverName string
	var sshUser string

	if targetChoice == "1" {
		runner = launch.NewLocalRunner()
		serverName = "local"
		sshUser = runner.User()
	} else {
		fmt.Print(theme.HighlightStyle.Render("  Server (alias or user@host) ▶ "))
		serverInput, _ := reader.ReadString('\n')
		serverInput = strings.TrimSpace(serverInput)

		target, err := parseServerTarget(serverInput)
		if err != nil {
			return fmt.Errorf("failed to resolve server: %w", err)
		}

		parts := strings.SplitN(target, "@", 2)
		user := parts[0]
		host := target
		if len(parts) == 2 {
			host = parts[1]
		}

		sshClient, err := ssh.NewClient(host, user, "")
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", target, err)
		}
		runner = launch.NewRemoteRunner(sshClient, user)
		serverName = serverInput
		sshUser = user
	}
	fmt.Println()

	// ── Step 5: Sync project (remote only) ─────────────────────────
	printLaunchStep(5, 10, "Project sync")

	remotePath := absPath
	if serverName != "local" {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Syncing project to remote..."))

		// Determine remote path
		remoteBase := "~/apps/" + filepath.Base(absPath)
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Remote path:"), theme.InfoStyle.Render(remoteBase))
		fmt.Print(theme.HighlightStyle.Render("  Path (enter to accept) ▶ "))
		pathInput, _ := reader.ReadString('\n')
		pathInput = strings.TrimSpace(pathInput)
		if pathInput != "" {
			remoteBase = pathInput
		}
		remotePath = remoteBase

		// Create dir and rsync
		runner.Run(fmt.Sprintf("mkdir -p %s", remotePath))

		target, _ := parseServerTarget(serverName)
		sshOpts := "ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10"

		// Build smart rsync excludes based on project type
		excludeArgs := ""
		for _, exc := range launch.RsyncExcludes(detected.Type) {
			excludeArgs += fmt.Sprintf(" --exclude='%s'", exc)
		}
		rsyncCmd := fmt.Sprintf("rsync -avz --progress%s -e '%s' %s/ %s:%s/", excludeArgs, sshOpts, absPath, target, remotePath)
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Running rsync..."))

		localRunner := launch.NewLocalRunner()
		out, err := localRunner.Run(rsyncCmd)
		if err != nil {
			return fmt.Errorf("rsync failed: %s: %w", out, err)
		}
		fmt.Println(theme.SuccessStyle.Render("  Synced"))
	} else {
		fmt.Println(theme.DimTextStyle.Render("  Local deployment, no sync needed"))
	}
	fmt.Println()

	// ── Step 6: App name + Domain ──────────────────────────────────
	printLaunchStep(6, 10, "Domain")

	defaultName := filepath.Base(absPath)
	fmt.Print(theme.HighlightStyle.Render(fmt.Sprintf("  App name (%s) ▶ ", defaultName)))
	nameInput, _ := reader.ReadString('\n')
	nameInput = strings.TrimSpace(nameInput)
	appName := defaultName
	if nameInput != "" {
		appName = nameInput
	}
	// Sanitize: lowercase, replace spaces/underscores with dashes
	appName = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(appName, " ", "-"), "_", "-"))

	fmt.Print(theme.HighlightStyle.Render("  Domain (e.g., myapp.example.com) ▶ "))
	domainInput, _ := reader.ReadString('\n')
	domain := strings.TrimSpace(domainInput)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	fmt.Println()

	// ── Step 7: Nginx reverse proxy ────────────────────────────────
	printLaunchStep(7, 10, "Nginx reverse proxy")

	fmt.Print(theme.HighlightStyle.Render("  Sudo password ▶ "))
	sudoPassword, _ := reader.ReadString('\n')
	sudoPassword = strings.TrimSpace(sudoPassword)

	// Ensure nginx is installed
	out, _ := runner.Run("which nginx 2>/dev/null")
	if out == "" {
		fmt.Println(theme.DimTextStyle.Render("  Installing nginx..."))
		if _, err := runner.RunSudo("apt-get update -qq && apt-get install -y -qq nginx", sudoPassword); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  Failed to install nginx: " + err.Error()))
			fmt.Println(theme.DimTextStyle.Render("  Continuing without nginx..."))
		}
	}

	// We'll set auth type after step 9, generate nginx config then
	fmt.Println(theme.DimTextStyle.Render("  Nginx config will be written after auth setup"))
	fmt.Println()

	// ── Step 8: SSL with Let's Encrypt ─────────────────────────────
	printLaunchStep(8, 10, "SSL certificate")

	enableSSL := promptUserYesNo(reader, "  Enable SSL with Let's Encrypt", true)
	var sslEmail string
	if enableSSL {
		// Ensure certbot is installed
		out, _ := runner.Run("which certbot 2>/dev/null")
		if out == "" {
			fmt.Println(theme.DimTextStyle.Render("  Installing certbot..."))
			runner.RunSudo("apt-get install -y -qq certbot python3-certbot-nginx", sudoPassword)
		}

		fmt.Print(theme.HighlightStyle.Render("  Email for Let's Encrypt ▶ "))
		emailInput, _ := reader.ReadString('\n')
		sslEmail = strings.TrimSpace(emailInput)
	}
	fmt.Println()

	// ── Step 9: Authentication ─────────────────────────────────────
	printLaunchStep(9, 10, "Authentication")

	fmt.Println("  1. Google OAuth (recommended)")
	fmt.Println("  2. HTTP Basic Auth")
	fmt.Println("  3. No authentication")
	authChoice := promptUserChoice(reader, "  Select auth", []string{"1", "2", "3"})

	authType := "none"
	var authCfg launch.AuthConfig

	switch authChoice {
	case "1":
		authType = "oauth2"
		fmt.Print(theme.HighlightStyle.Render("  Google Client ID ▶ "))
		clientID, _ := reader.ReadString('\n')
		fmt.Print(theme.HighlightStyle.Render("  Google Client Secret ▶ "))
		clientSecret, _ := reader.ReadString('\n')
		fmt.Print(theme.HighlightStyle.Render("  Allowed email domain (* for any) ▶ "))
		emailDomain, _ := reader.ReadString('\n')

		cookieSecret, err := launch.GenerateCookieSecret()
		if err != nil {
			return fmt.Errorf("failed to generate cookie secret: %w", err)
		}

		authCfg = launch.AuthConfig{
			Type:               "oauth2",
			GoogleClientID:     strings.TrimSpace(string(clientID)),
			GoogleClientSecret: strings.TrimSpace(string(clientSecret)),
			CookieSecret:       cookieSecret,
			EmailDomain:        strings.TrimSpace(string(emailDomain)),
		}

		// Install oauth2-proxy
		fmt.Println(theme.DimTextStyle.Render("  Installing oauth2-proxy..."))
		if err := launch.InstallOAuth2Proxy(sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  Failed to install oauth2-proxy: " + err.Error()))
			fmt.Println(theme.DimTextStyle.Render("  Falling back to basic auth"))
			authType = "basic"
			authChoice = "2" // fall through
		}

	case "2":
		authType = "basic"
	}

	if authType == "basic" {
		fmt.Print(theme.HighlightStyle.Render("  Username ▶ "))
		username, _ := reader.ReadString('\n')
		fmt.Print(theme.HighlightStyle.Render("  Password ▶ "))
		password, _ := reader.ReadString('\n')

		authCfg = launch.AuthConfig{
			Type:     "basic",
			Username: strings.TrimSpace(string(username)),
			Password: strings.TrimSpace(string(password)),
		}

		// Install apache2-utils for htpasswd
		runner.RunSudo("apt-get install -y -qq apache2-utils 2>/dev/null || true", sudoPassword)

		fmt.Println(theme.DimTextStyle.Render("  Creating htpasswd file..."))
		if err := launch.CreateHtpasswd(appName, authCfg.Username, authCfg.Password, sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  Failed: " + err.Error()))
		}
	}
	fmt.Println()

	// Now write nginx config with auth
	fmt.Println(theme.DimTextStyle.Render("  Writing nginx config..."))
	nginxCfg := launch.NginxConfig{
		Domain:   domain,
		Port:     detected.Port,
		AppName:  appName,
		AuthType: authType,
	}
	nginxContent, err := launch.GenerateNginxConfig(nginxCfg)
	if err != nil {
		return fmt.Errorf("failed to generate nginx config: %w", err)
	}
	if err := launch.InstallNginxConfig(appName, nginxContent, sudoPassword, runner); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Nginx setup failed: " + err.Error()))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Nginx configured"))
	}

	// SSL (must come after nginx config is in place)
	if enableSSL && sslEmail != "" {
		fmt.Println(theme.DimTextStyle.Render("  Running certbot..."))
		if err := launch.SetupSSL(domain, sslEmail, sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  SSL failed: " + err.Error()))
			enableSSL = false
		} else {
			fmt.Println(theme.SuccessStyle.Render("  SSL certificate installed"))
		}
	}

	// OAuth2 proxy systemd service
	if authType == "oauth2" {
		fmt.Println(theme.DimTextStyle.Render("  Creating OAuth2 proxy service..."))
		oauthUnit := launch.GenerateOAuth2ProxyUnit(appName, domain, sshUser, authCfg)
		if err := launch.InstallOAuth2ProxyService(appName, oauthUnit, sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  OAuth2 proxy service failed: " + err.Error()))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  OAuth2 proxy running"))
		}
	}

	// ── Step 10: Systemd service ───────────────────────────────────
	printLaunchStep(10, 10, "Systemd service")

	serviceName := launch.ServiceName(appName)

	// Determine the full exec command
	execStart := detected.RunCommand
	// For node/python, use full path
	if detected.Type == launch.ProjectNodeJS {
		// Use bash -c for complex commands
		execStart = fmt.Sprintf("/bin/bash -c 'cd %s && %s'", remotePath, detected.RunCommand)
	} else if detected.Type == launch.ProjectPython {
		execStart = fmt.Sprintf("/bin/bash -c 'cd %s && %s'", remotePath, detected.RunCommand)
	} else if detected.Type == launch.ProjectDocker {
		execStart = fmt.Sprintf("/bin/bash -c 'cd %s && %s'", remotePath, detected.RunCommand)
	} else {
		execStart = fmt.Sprintf("/bin/bash -c 'cd %s && %s'", remotePath, detected.RunCommand)
	}

	envVars := map[string]string{
		"NODE_ENV": "production",
	}

	// Import .env variables if found
	if analysis.EnvFile != "" {
		envData, err := launch.ParseEnvFile(analysis.EnvFile)
		if err == nil && len(envData) > 0 {
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("  Found .env variables:"))
			for k, v := range envData {
				fmt.Printf("    %s=%s\n", theme.HighlightStyle.Render(k), theme.DimTextStyle.Render(launch.MaskEnvValue(v)))
			}
			if promptUserYesNo(reader, "  Import into systemd service", true) {
				for k, v := range envData {
					envVars[k] = v
				}
				fmt.Println(theme.SuccessStyle.Render("  Imported"))
			}
		}
	}

	sysCfg := launch.SystemdConfig{
		Name:        serviceName,
		Description: fmt.Sprintf("anime launch: %s", appName),
		ExecStart:   execStart,
		WorkingDir:  remotePath,
		User:        sshUser,
		Port:        detected.Port,
		Environment: envVars,
	}

	unitContent, err := launch.GenerateSystemdUnit(sysCfg)
	if err != nil {
		return fmt.Errorf("failed to generate systemd unit: %w", err)
	}

	fmt.Println(theme.DimTextStyle.Render("  Creating service..."))
	if err := launch.InstallSystemdUnit(serviceName, unitContent, sudoPassword, runner); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Service creation failed: " + err.Error()))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  Service started"))
	}
	fmt.Println()

	// ── Save to config ─────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app := config.LaunchedApp{
		Name:           appName,
		Path:           absPath,
		ProjectType:    string(detected.Type),
		RunCommand:     detected.RunCommand,
		Port:           detected.Port,
		Domain:         domain,
		Server:         serverName,
		RemotePath:     remotePath,
		ServiceName:    serviceName,
		AuthType:       authType,
		SSLEnabled:     enableSSL,
		PackageManager: detected.PackageManager,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}
	cfg.AddLaunchedApp(app)
	if err := cfg.Save(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Failed to save config: " + err.Error()))
	}

	// ── Summary ────────────────────────────────────────────────────
	fmt.Println(theme.RenderBanner("LAUNCHED"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("App:"), theme.HighlightStyle.Render(appName))
	fmt.Printf("  %s %s (%s)\n", theme.DimTextStyle.Render("Type:"), string(detected.Type), detected.Framework)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Command:"), detected.RunCommand)
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Port:"), detected.Port)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), serverName)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Domain:"), domain)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("SSL:"), boolToYesNo(enableSSL))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Auth:"), authType)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Service:"), serviceName)
	fmt.Println()

	protocol := "http"
	if enableSSL {
		protocol = "https"
	}
	fmt.Printf("  %s %s://%s\n", theme.SuccessStyle.Render("URL:"), protocol, domain)
	fmt.Println()
	fmt.Printf("  %s anime launch status\n", theme.DimTextStyle.Render("Check:"))
	fmt.Printf("  %s anime launch logs %s\n", theme.DimTextStyle.Render("Logs:"), appName)
	fmt.Printf("  %s anime launch stop %s\n", theme.DimTextStyle.Render("Stop:"), appName)
	fmt.Println()

	return nil
}

// ── Subcommands ───────────────────────────────────────────────────────

func runLaunchStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.LaunchedApps) == 0 {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  No launched apps"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("LAUNCH STATUS"))
	fmt.Println()

	for _, app := range cfg.LaunchedApps {
		runner, err := getRunnerForApp(app)
		if err != nil {
			fmt.Printf("  %s  %s  %s\n",
				theme.ErrorStyle.Render("?"),
				theme.HighlightStyle.Render(app.Name),
				theme.DimTextStyle.Render("(cannot connect)"))
			continue
		}

		status, _ := launch.GetServiceStatus(app.ServiceName, runner)

		statusStyle := theme.ErrorStyle
		if status == "active" {
			statusStyle = theme.SuccessStyle
		}

		fmt.Printf("  %s  %s  %s  %s\n",
			statusStyle.Render(status),
			theme.HighlightStyle.Render(app.Name),
			theme.DimTextStyle.Render(app.Domain),
			theme.DimTextStyle.Render("("+app.Server+")"))
	}
	fmt.Println()

	return nil
}

func runLaunchStop(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return err
	}

	runner, err := getRunnerForApp(*app)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(theme.HighlightStyle.Render("  Sudo password ▶ "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Println()
	fmt.Printf("  Stopping %s...\n", appName)

	if err := launch.StopService(app.ServiceName, password, runner); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Also stop OAuth2 proxy if applicable
	if app.AuthType == "oauth2" {
		launch.StopService("anime-"+appName+"-oauth2", password, runner)
	}

	fmt.Println(theme.SuccessStyle.Render("  Stopped"))
	fmt.Println()

	return nil
}

func runLaunchLogs(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return err
	}

	runner, err := getRunnerForApp(*app)
	if err != nil {
		return err
	}

	logs, err := launch.GetServiceLogs(app.ServiceName, launchLogLines, runner)
	if err != nil {
		return err
	}

	fmt.Println(logs)
	return nil
}

func runLaunchList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.LaunchedApps) == 0 {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  No launched apps"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("LAUNCHED APPS"))
	fmt.Println()

	for _, app := range cfg.LaunchedApps {
		protocol := "http"
		if app.SSLEnabled {
			protocol = "https"
		}
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(app.Name))
		fmt.Printf("    %s %s://%s\n", theme.DimTextStyle.Render("URL:"), protocol, app.Domain)
		fmt.Printf("    %s %s  %s %d  %s %s  %s %s\n",
			theme.DimTextStyle.Render("Type:"), app.ProjectType,
			theme.DimTextStyle.Render("Port:"), app.Port,
			theme.DimTextStyle.Render("Auth:"), app.AuthType,
			theme.DimTextStyle.Render("Server:"), app.Server)
		fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("Service:"), app.ServiceName)
		fmt.Println()
	}

	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────

func printLaunchStep(current, total int, title string) {
	fmt.Printf("  %s %s\n",
		theme.InfoStyle.Render(fmt.Sprintf("[%d/%d]", current, total)),
		theme.HighlightStyle.Render(title))
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func getRunnerForApp(app config.LaunchedApp) (launch.CommandRunner, error) {
	if app.Server == "" || app.Server == "local" {
		return launch.NewLocalRunner(), nil
	}

	target, err := parseServerTarget(app.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server %s: %w", app.Server, err)
	}

	parts := strings.SplitN(target, "@", 2)
	user := parts[0]
	host := target
	if len(parts) == 2 {
		host = parts[1]
	}

	client, err := ssh.NewClient(host, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", app.Server, err)
	}

	return launch.NewRemoteRunner(client, user), nil
}
