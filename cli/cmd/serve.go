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
	"github.com/joshkornreich/anime/internal/stack"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var serveLogLines int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve and manage web applications",
	Long: `Full-stack app server: detect, deploy, proxy, secure, and manage.

SUBCOMMANDS:
  setup       Interactive wizard to serve a web app
  status      Show status of served apps
  stop        Stop a served app
  logs        View logs for a served app
  list        List all served apps

EXAMPLES:
  anime serve                          # Start setup wizard
  anime serve setup ./myapp            # Serve app at path
  anime serve status                   # Show all running apps
  anime serve stop myapp               # Stop an app
  anime serve logs myapp -n 100        # View last 100 lines`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServeSetup(cmd, args)
	},
}

var serveSetupCmd = &cobra.Command{
	Use:   "setup [path]",
	Short: "Interactive wizard to serve a web application",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServeSetup,
}

var serveStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of served applications",
	RunE:  runServeStatus,
}

var serveStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a served application",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeStop,
}

var serveLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show logs for a served application",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeLogs,
}

var serveListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all served applications",
	RunE:    runServeList,
}

var serveRemoveCmd = &cobra.Command{
	Use:   "remove <app>",
	Short: "Stop and completely remove a deployed app",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeRemove,
}

var serveStackCmd = &cobra.Command{
	Use:   "stack [path]",
	Short: "Deploy a multi-service stack from anime.yaml",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServeStack,
}

var serveEditCmd = &cobra.Command{
	Use:   "edit <app>",
	Short: "Interactively reconfigure a deployed app",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeEdit,
}

func init() {
	serveCmd.AddCommand(serveSetupCmd)
	serveCmd.AddCommand(serveStatusCmd)
	serveCmd.AddCommand(serveStopCmd)
	serveCmd.AddCommand(serveLogsCmd)
	serveCmd.AddCommand(serveListCmd)
	serveCmd.AddCommand(serveRemoveCmd)
	serveCmd.AddCommand(serveStackCmd)
	serveCmd.AddCommand(serveEditCmd)

	serveLogsCmd.Flags().IntVarP(&serveLogLines, "lines", "n", 50, "Number of log lines")

	rootCmd.AddCommand(serveCmd)
}

// ── Setup Wizard ──────────────────────────────────────────────────────

func runServeSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check for anime.yaml stack config
	if configPath, err := stack.FindStackConfig(absPath); err == nil {
		fmt.Println()
		fmt.Printf("  %s Found anime.yaml at %s\n", theme.InfoStyle.Render("*"), theme.DimTextStyle.Render(configPath))
		if promptUserYesNo(reader, "  Deploy as stack", true) {
			// Delegate to stack deployment
			return runServeStack(cmd, args)
		}
		fmt.Println(theme.DimTextStyle.Render("  Continuing with single-service wizard..."))
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("SERVE"))
	fmt.Println()

	// ── Step 1: Detect project type ────────────────────────────────
	printServeStep(1, 12, "Detecting project")

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
	printServeStep(2, 12, "Run command")

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Suggested:"), theme.InfoStyle.Render(detected.RunCommand))
	fmt.Print(theme.HighlightStyle.Render("  Command (enter to accept) ▶ "))
	runCmdInput, _ := reader.ReadString('\n')
	runCmdInput = strings.TrimSpace(runCmdInput)
	if runCmdInput != "" {
		detected.RunCommand = runCmdInput
	}
	fmt.Println()

	// ── Step 3: Port ───────────────────────────────────────────────
	printServeStep(3, 12, "Application port")

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

	// ── Step 4: Database detection ────────────────────────────────
	printServeStep(4, 12, "Database detection")

	dbInfo := analysis.Database
	dbProvisionLocal := false   // true if user wants us to provision Postgres
	dbConnectionURL := ""       // user-provided connection string
	dbRunMigrations := false    // whether to run migrations after deploy

	if dbInfo != nil && dbInfo.Detected {
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("Database:"), theme.HighlightStyle.Render(string(dbInfo.PrimaryType)))
		if len(dbInfo.Tools) > 0 {
			toolNames := make([]string, len(dbInfo.Tools))
			for i, t := range dbInfo.Tools {
				toolNames[i] = t.Name
			}
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Tools:"), strings.Join(toolNames, ", "))
		}
		if len(dbInfo.Tables) > 0 {
			display := dbInfo.Tables
			if len(display) > 8 {
				display = display[:8]
			}
			fmt.Printf("  %s %s", theme.DimTextStyle.Render("Tables:"), strings.Join(display, ", "))
			if len(dbInfo.Tables) > 8 {
				fmt.Printf(" (+%d more)", len(dbInfo.Tables)-8)
			}
			fmt.Println()
		}
		if dbInfo.HasMigrations {
			fmt.Printf("  %s %s (%s)\n", theme.DimTextStyle.Render("Migrations:"), dbInfo.MigrationTool, dbInfo.MigrationCmd)
		}
		if len(dbInfo.Queries) > 0 {
			fmt.Printf("  %s %d patterns across source\n", theme.DimTextStyle.Render("Queries:"), len(dbInfo.Queries))
		}
		if len(dbInfo.EnvVars) > 0 {
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Env vars:"), strings.Join(dbInfo.EnvVars, ", "))
		}
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No database usage detected"))
	}
	fmt.Println()

	// ── Step 5: Database provisioning ─────────────────────────────
	printServeStep(5, 12, "Database setup")

	if dbInfo != nil && dbInfo.NeedsDatabase {
		// Check if DATABASE_URL already in .env
		hasDBURL := false
		if analysis.EnvFile != "" {
			envData, err := launch.ParseEnvFile(analysis.EnvFile)
			if err == nil {
				if _, ok := envData["DATABASE_URL"]; ok {
					hasDBURL = true
				}
			}
		}

		if hasDBURL {
			fmt.Println(theme.DimTextStyle.Render("  DATABASE_URL found in .env — using existing"))
		} else if dbInfo.PrimaryType == launch.DBPostgres {
			fmt.Println("  1. Provision local Postgres (recommended)")
			fmt.Println("  2. Provide connection string")
			fmt.Println("  3. Skip database setup")
			dbChoice := promptUserChoice(reader, "  Select", []string{"1", "2", "3"})

			switch dbChoice {
			case "1":
				dbProvisionLocal = true
				fmt.Println(theme.DimTextStyle.Render("  Postgres will be provisioned on target server"))
			case "2":
				fmt.Print(theme.HighlightStyle.Render("  DATABASE_URL ▶ "))
				urlInput, _ := reader.ReadString('\n')
				dbConnectionURL = strings.TrimSpace(urlInput)
			case "3":
				fmt.Println(theme.DimTextStyle.Render("  Skipping database setup"))
			}
		} else {
			fmt.Println("  1. Provide connection string")
			fmt.Println("  2. Skip database setup")
			dbChoice := promptUserChoice(reader, "  Select", []string{"1", "2"})

			if dbChoice == "1" {
				fmt.Print(theme.HighlightStyle.Render("  DATABASE_URL ▶ "))
				urlInput, _ := reader.ReadString('\n')
				dbConnectionURL = strings.TrimSpace(urlInput)
			} else {
				fmt.Println(theme.DimTextStyle.Render("  Skipping database setup"))
			}
		}
	} else {
		fmt.Println(theme.DimTextStyle.Render("  No database provisioning needed"))
	}
	fmt.Println()

	// ── Step 6: Target server ──────────────────────────────────────
	printServeStep(6, 12, "Target server")

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

	// ── Step 7: Sync project (remote only) ─────────────────────────
	printServeStep(7, 12, "Project sync")

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

	// ── Step 8: App name + Domain ──────────────────────────────────
	printServeStep(8, 12, "Domain")

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

	// ── Step 9: Nginx reverse proxy ────────────────────────────────
	printServeStep(9, 12, "Nginx reverse proxy")

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

	// ── Step 10: SSL with Let's Encrypt ────────────────────────────
	printServeStep(10, 12, "SSL certificate")

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

	// ── Step 11: Authentication ────────────────────────────────────
	printServeStep(11, 12, "Authentication")

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

	// ── Step 12: Systemd service ───────────────────────────────────
	printServeStep(12, 12, "Systemd service")

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

	// Database provisioning (now that runner is available)
	if dbProvisionLocal {
		dbName := strings.ReplaceAll(appName, "-", "_") + "_db"
		dbUser := strings.ReplaceAll(appName, "-", "_")
		dbPassword := launch.GenerateRandomPassword(16)

		fmt.Println(theme.DimTextStyle.Render("  Provisioning Postgres..."))
		if err := launch.ProvisionPostgres(dbName, dbUser, dbPassword, sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  Postgres provisioning failed: " + err.Error()))
		} else {
			dbConnectionURL = fmt.Sprintf("postgresql://%s:%s@localhost:5432/%s", dbUser, dbPassword, dbName)
			fmt.Println(theme.SuccessStyle.Render("  Postgres provisioned"))
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Database:"), dbName)
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("User:"), dbUser)
		}
	}

	// Inject DATABASE_URL if we have one
	if dbConnectionURL != "" {
		envVars["DATABASE_URL"] = dbConnectionURL
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
		Description: fmt.Sprintf("anime serve: %s", appName),
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

	// ── Run database migrations ─────────────────────────────────────
	if dbInfo != nil && dbInfo.HasMigrations && dbConnectionURL != "" {
		fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Migrations:"), dbInfo.MigrationCmd)
		if promptUserYesNo(reader, "  Run migrations now", true) {
			migCmd := fmt.Sprintf("cd %s && DATABASE_URL='%s' %s", remotePath, dbConnectionURL, dbInfo.MigrationCmd)
			fmt.Println(theme.DimTextStyle.Render("  Running migrations..."))
			out, err := runner.Run(migCmd)
			if err != nil {
				fmt.Println(theme.ErrorStyle.Render("  Migration failed: " + err.Error()))
				if out != "" {
					fmt.Println(theme.DimTextStyle.Render("  " + out))
				}
				dbRunMigrations = false
			} else {
				fmt.Println(theme.SuccessStyle.Render("  Migrations applied"))
				dbRunMigrations = true
			}
		}
		fmt.Println()
	}

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
	if dbInfo != nil && dbInfo.Detected {
		app.DatabaseType = string(dbInfo.PrimaryType)
	}
	if dbProvisionLocal {
		app.DatabaseLocal = true
		app.DatabaseName = strings.ReplaceAll(appName, "-", "_") + "_db"
		app.DatabaseUser = strings.ReplaceAll(appName, "-", "_")
	}
	if dbRunMigrations {
		app.MigrationsRun = true
	}
	cfg.AddLaunchedApp(app)
	if err := cfg.Save(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Failed to save config: " + err.Error()))
	}

	// ── Summary ────────────────────────────────────────────────────
	fmt.Println(theme.RenderBanner("SERVING"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("App:"), theme.HighlightStyle.Render(appName))
	fmt.Printf("  %s %s (%s)\n", theme.DimTextStyle.Render("Type:"), string(detected.Type), detected.Framework)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Command:"), detected.RunCommand)
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Port:"), detected.Port)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), serverName)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Domain:"), domain)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("SSL:"), boolToYesNo(enableSSL))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Auth:"), authType)
	if dbInfo != nil && dbInfo.Detected {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Database:"), string(dbInfo.PrimaryType))
		if dbProvisionLocal {
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("DB name:"), strings.ReplaceAll(appName, "-", "_")+"_db")
		}
		if dbRunMigrations {
			fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Migrations:"), "applied")
		}
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Service:"), serviceName)
	fmt.Println()

	protocol := "http"
	if enableSSL {
		protocol = "https"
	}
	fmt.Printf("  %s %s://%s\n", theme.SuccessStyle.Render("URL:"), protocol, domain)
	fmt.Println()
	fmt.Printf("  %s anime serve status\n", theme.DimTextStyle.Render("Check:"))
	fmt.Printf("  %s anime serve logs %s\n", theme.DimTextStyle.Render("Logs:"), appName)
	fmt.Printf("  %s anime serve stop %s\n", theme.DimTextStyle.Render("Stop:"), appName)
	fmt.Println()

	return nil
}

// ── Subcommands ───────────────────────────────────────────────────────

func runServeStatus(cmd *cobra.Command, args []string) error {
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
	fmt.Println(theme.RenderBanner("SERVE STATUS"))
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

func runServeStop(cmd *cobra.Command, args []string) error {
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

func runServeLogs(cmd *cobra.Command, args []string) error {
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

	logs, err := launch.GetServiceLogs(app.ServiceName, serveLogLines, runner)
	if err != nil {
		return err
	}

	fmt.Println(logs)
	return nil
}

func runServeList(cmd *cobra.Command, args []string) error {
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
	fmt.Println(theme.RenderBanner("SERVED APPS"))
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

func runServeRemove(cmd *cobra.Command, args []string) error {
	appName := args[0]

	// Load config and get app
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return err
	}

	// Confirm with user
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	if !promptUserYesNo(reader, "  This will stop the app and remove all server configuration. Continue", false) {
		fmt.Println(theme.DimTextStyle.Render("  Aborted"))
		fmt.Println()
		return nil
	}

	// Get runner for the app's server
	runner, err := getRunnerForApp(*app)
	if err != nil {
		return err
	}

	// Prompt for sudo password
	fmt.Print(theme.HighlightStyle.Render("  Sudo password ▶ "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Println()
	fmt.Printf("  Removing %s...\n", appName)

	// Stop the service (handle both static and process apps)
	if app.ServiceName == "" {
		// For static apps, nginx config removal will be handled below
		fmt.Println(theme.DimTextStyle.Render("  Static app - no process to stop"))
	} else {
		// Stop the systemd service
		fmt.Println(theme.DimTextStyle.Render("  Stopping service..."))
		if err := launch.StopService(app.ServiceName, password, runner); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: failed to stop service: " + err.Error()))
		}
	}

	// Also stop OAuth2 proxy if applicable
	if app.AuthType == "oauth2" {
		oauthService := "anime-" + appName + "-oauth2"
		fmt.Println(theme.DimTextStyle.Render("  Stopping OAuth2 proxy..."))
		launch.StopService(oauthService, password, runner)
	}

	// Remove nginx config
	fmt.Println(theme.DimTextStyle.Render("  Removing nginx config..."))
	sitesEnabled := fmt.Sprintf("/etc/nginx/sites-enabled/%s", appName)
	sitesAvailable := fmt.Sprintf("/etc/nginx/sites-available/%s", appName)

	if _, err := runner.RunSudo(fmt.Sprintf("rm -f %s", sitesEnabled), password); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: failed to remove sites-enabled: " + err.Error()))
	}
	if _, err := runner.RunSudo(fmt.Sprintf("rm -f %s", sitesAvailable), password); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: failed to remove sites-available: " + err.Error()))
	}

	// Reload nginx
	if _, err := runner.RunSudo("systemctl reload nginx", password); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: nginx reload failed: " + err.Error()))
	}

	// Remove systemd service (for non-static apps)
	if app.ServiceName != "" {
		fmt.Println(theme.DimTextStyle.Render("  Removing systemd service..."))
		serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", app.ServiceName)

		// Disable the service
		if _, err := runner.RunSudo(fmt.Sprintf("systemctl disable %s 2>/dev/null || true", app.ServiceName), password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: failed to disable service: " + err.Error()))
		}

		// Remove the service file
		if _, err := runner.RunSudo(fmt.Sprintf("rm -f %s", serviceFile), password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: failed to remove service file: " + err.Error()))
		}

		// Reload systemd
		if _, err := runner.RunSudo("systemctl daemon-reload", password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: systemctl daemon-reload failed: " + err.Error()))
		}
	}

	// Remove OAuth2 proxy service if it was enabled
	if app.AuthType == "oauth2" {
		fmt.Println(theme.DimTextStyle.Render("  Removing OAuth2 proxy service..."))
		oauthService := "anime-" + appName + "-oauth2"
		oauthServiceFile := fmt.Sprintf("/etc/systemd/system/%s.service", oauthService)

		// Disable the service
		if _, err := runner.RunSudo(fmt.Sprintf("systemctl disable %s 2>/dev/null || true", oauthService), password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: failed to disable OAuth2 service: " + err.Error()))
		}

		// Remove the service file
		if _, err := runner.RunSudo(fmt.Sprintf("rm -f %s", oauthServiceFile), password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: failed to remove OAuth2 service file: " + err.Error()))
		}

		// Reload systemd
		if _, err := runner.RunSudo("systemctl daemon-reload", password); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Warning: systemctl daemon-reload failed: " + err.Error()))
		}
	}

	// Remove rate limit config if exists
	rateLimitConf := fmt.Sprintf("/etc/nginx/conf.d/%s-ratelimit.conf", appName)
	fmt.Println(theme.DimTextStyle.Render("  Removing rate limit config..."))
	if _, err := runner.RunSudo(fmt.Sprintf("rm -f %s", rateLimitConf), password); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: failed to remove rate limit config: " + err.Error()))
	}

	// Remove from config
	cfg.RemoveLaunchedApp(appName)

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Failed to save config: " + err.Error()))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  App removed successfully"))
	fmt.Println()

	return nil
}

func runServeStack(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Get path from args (default to current dir)
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("STACK DEPLOY"))
	fmt.Println()

	// Find anime.yaml
	printServeStep(1, 6, "Finding stack config")
	configPath, err := stack.FindStackConfig(absPath)
	if err != nil {
		return fmt.Errorf("anime.yaml not found: %w", err)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("Found:"), theme.DimTextStyle.Render(configPath))
	fmt.Println()

	// Load and parse config
	printServeStep(2, 6, "Loading stack config")
	stackCfg, err := stack.LoadStackConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("Stack:"), theme.HighlightStyle.Render(stackCfg.Name))
	fmt.Println()

	// Validate config
	printServeStep(3, 6, "Validating config")
	if err := stack.ValidateStackConfig(stackCfg); err != nil {
		return fmt.Errorf("invalid stack config: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("  Config is valid"))
	fmt.Println()

	// Show stack summary
	printServeStep(4, 6, "Stack summary")
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Name:"), theme.HighlightStyle.Render(stackCfg.Name))
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Services:"), len(stackCfg.Services))

	// List services with their dependencies
	for name, svc := range stackCfg.Services {
		svcType := "app"
		if svc.Type == "postgres" {
			svcType = "database"
		}
		fmt.Printf("    - %s (%s)", theme.HighlightStyle.Render(name), svcType)
		if len(svc.DependsOn) > 0 {
			fmt.Printf(" -> depends on: %s", strings.Join(svc.DependsOn, ", "))
		}
		fmt.Println()
	}

	if stackCfg.Routing != nil && stackCfg.Routing.Domain != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Domain:"), stackCfg.Routing.Domain)
	}
	if stackCfg.SSL {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("SSL:"), "enabled")
	}
	fmt.Println()

	// Server selection
	printServeStep(5, 6, "Target server")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var serverChoices []string

	// Add "local" as first option
	serverChoices = append(serverChoices, "1")
	fmt.Println("  1. Local (this machine)")

	// Add configured servers
	for i, srv := range cfg.Servers {
		num := strconv.Itoa(i + 2)
		serverChoices = append(serverChoices, num)
		fmt.Printf("  %s. %s (%s@%s)\n", num, srv.Name, srv.User, srv.Host)
	}

	// Add "other" option for manual entry
	otherNum := strconv.Itoa(len(cfg.Servers) + 2)
	serverChoices = append(serverChoices, otherNum)
	fmt.Printf("  %s. Other (enter manually)\n", otherNum)

	serverChoice := promptUserChoice(reader, "  Select server", serverChoices)

	var runner launch.CommandRunner
	var serverName string
	var sshUser string

	choiceIdx, _ := strconv.Atoi(serverChoice)
	if choiceIdx == 1 {
		// Local
		runner = launch.NewLocalRunner()
		serverName = "local"
		sshUser = runner.User()
	} else if choiceIdx <= len(cfg.Servers)+1 {
		// Configured server
		srv := cfg.Servers[choiceIdx-2]
		sshClient, err := ssh.NewClient(srv.Host, srv.User, "")
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", srv.Name, err)
		}
		runner = launch.NewRemoteRunner(sshClient, srv.User)
		serverName = srv.Name
		sshUser = srv.User
	} else {
		// Manual entry
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

	// Prompt for sudo password
	fmt.Print(theme.HighlightStyle.Render("  Sudo password ▶ "))
	sudoPassword, _ := reader.ReadString('\n')
	sudoPassword = strings.TrimSpace(sudoPassword)
	fmt.Println()

	// Deploy stack
	printServeStep(6, 6, "Deploying stack")
	fmt.Println(theme.DimTextStyle.Render("  This may take a few minutes..."))
	fmt.Println()

	if err := stack.DeployStack(stackCfg, serverName, sudoPassword, runner); err != nil {
		return fmt.Errorf("stack deployment failed: %w", err)
	}

	// Save stack info to config
	for name, svc := range stackCfg.Services {
		// Skip database services for LaunchedApp tracking
		if svc.Type == "postgres" {
			continue
		}

		app := config.LaunchedApp{
			Name:        stackCfg.Name + "-" + name,
			Path:        absPath,
			ProjectType: "stack",
			RunCommand:  svc.Start,
			Port:        svc.Port,
			Domain:      svc.Domain,
			Server:      serverName,
			ServiceName: launch.ServiceName(stackCfg.Name + "-" + name),
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		cfg.AddLaunchedApp(app)
	}

	if err := cfg.Save(); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Warning: failed to save config: " + err.Error()))
	}

	// Success summary
	fmt.Println()
	fmt.Println(theme.RenderBanner("STACK DEPLOYED"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Stack:"), theme.HighlightStyle.Render(stackCfg.Name))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), serverName)
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("User:"), sshUser)
	fmt.Println()

	// Show URLs for each service
	fmt.Println(theme.InfoStyle.Render("  Services:"))
	protocol := "http"
	if stackCfg.SSL {
		protocol = "https"
	}
	for name, svc := range stackCfg.Services {
		if svc.Type == "postgres" {
			fmt.Printf("    %s %s (database)\n", theme.SuccessStyle.Render("*"), name)
		} else if svc.Domain != "" {
			fmt.Printf("    %s %s -> %s://%s\n", theme.SuccessStyle.Render("*"), name, protocol, svc.Domain)
		} else if svc.Port > 0 {
			fmt.Printf("    %s %s -> port %d\n", theme.SuccessStyle.Render("*"), name, svc.Port)
		} else {
			fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("*"), name)
		}
	}
	fmt.Println()

	fmt.Printf("  %s anime serve status\n", theme.DimTextStyle.Render("Check:"))
	fmt.Printf("  %s anime serve stop %s-<service>\n", theme.DimTextStyle.Render("Stop:"), stackCfg.Name)
	fmt.Println()

	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────

func printServeStep(current, total int, title string) {
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

func runServeEdit(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return fmt.Errorf("app not found: %s", appName)
	}

	runner, err := getRunnerForApp(*app)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(theme.RenderBanner("EDIT: " + appName))
	fmt.Println()

	// Track changes
	var changes []string
	needsNginxReload := false
	needsServiceRestart := false

	// Show current configuration
	fmt.Println(theme.InfoStyle.Render("  Current configuration:"))
	protocol := "http"
	if app.SSLEnabled {
		protocol = "https"
	}
	fmt.Printf("    %s %s://%s\n", theme.DimTextStyle.Render("URL:"), protocol, app.Domain)
	if app.ServiceName != "" {
		fmt.Printf("    %s %d\n", theme.DimTextStyle.Render("Port:"), app.Port)
	}
	fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("Auth:"), app.AuthType)
	fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("SSL:"), boolToYesNo(app.SSLEnabled))
	fmt.Println()

	// Domain change
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Current domain:"), theme.HighlightStyle.Render(app.Domain))
	if promptUserYesNo(reader, "  Change domain", false) {
		fmt.Print(theme.HighlightStyle.Render("  New domain ▶ "))
		newDomain, _ := reader.ReadString('\n')
		newDomain = strings.TrimSpace(newDomain)
		if newDomain != "" && newDomain != app.Domain {
			app.Domain = newDomain
			changes = append(changes, fmt.Sprintf("Domain: %s", newDomain))
			needsNginxReload = true
		}
	}
	fmt.Println()

	// Port change (only for apps with services)
	if app.ServiceName != "" {
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Current port:"), app.Port)
		if promptUserYesNo(reader, "  Change port", false) {
			fmt.Print(theme.HighlightStyle.Render("  New port ▶ "))
			portInput, _ := reader.ReadString('\n')
			portInput = strings.TrimSpace(portInput)
			if portInput != "" {
				if newPort, err := strconv.Atoi(portInput); err == nil && newPort != app.Port {
					app.Port = newPort
					changes = append(changes, fmt.Sprintf("Port: %d", newPort))
					needsNginxReload = true
					needsServiceRestart = true
				}
			}
		}
		fmt.Println()
	}

	// Auth type change
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Current auth:"), app.AuthType)
	if promptUserYesNo(reader, "  Change authentication", false) {
		fmt.Println("    1. Google OAuth")
		fmt.Println("    2. HTTP Basic Auth")
		fmt.Println("    3. No authentication")
		authChoice := promptUserChoice(reader, "    Select", []string{"1", "2", "3"})

		newAuthType := app.AuthType
		switch authChoice {
		case "1":
			newAuthType = "oauth2"
		case "2":
			newAuthType = "basic"
		case "3":
			newAuthType = "none"
		}

		if newAuthType != app.AuthType {
			app.AuthType = newAuthType
			changes = append(changes, fmt.Sprintf("Auth: %s", newAuthType))
			needsNginxReload = true
		}
	}
	fmt.Println()

	// SSL change
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Current SSL:"), boolToYesNo(app.SSLEnabled))
	if promptUserYesNo(reader, "  Toggle SSL", false) {
		app.SSLEnabled = !app.SSLEnabled
		changes = append(changes, fmt.Sprintf("SSL: %s", boolToYesNo(app.SSLEnabled)))
		needsNginxReload = true
	}
	fmt.Println()

	// If no changes, exit early
	if len(changes) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No changes made"))
		fmt.Println()
		return nil
	}

	// Show summary and confirm
	fmt.Println(theme.InfoStyle.Render("  Changes to apply:"))
	for _, change := range changes {
		fmt.Printf("    - %s\n", change)
	}
	fmt.Println()

	if !promptUserYesNo(reader, "  Apply changes", true) {
		fmt.Println(theme.DimTextStyle.Render("  Aborted"))
		return nil
	}

	// Get sudo password if needed
	var sudoPassword string
	if needsNginxReload || needsServiceRestart {
		fmt.Print(theme.HighlightStyle.Render("  Sudo password ▶ "))
		sudoPassword, _ = reader.ReadString('\n')
		sudoPassword = strings.TrimSpace(sudoPassword)
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Applying changes..."))

	// Regenerate nginx config if needed
	if needsNginxReload {
		nginxCfg := launch.NginxConfig{
			Domain:   app.Domain,
			Port:     app.Port,
			AppName:  app.Name,
			AuthType: app.AuthType,
		}
		nginxContent, err := launch.GenerateNginxConfig(nginxCfg)
		if err != nil {
			return fmt.Errorf("failed to generate nginx config: %w", err)
		}
		if err := launch.InstallNginxConfig(app.Name, nginxContent, sudoPassword, runner); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  Nginx update failed: " + err.Error()))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  Nginx config updated"))
		}
	}

	// Restart service if needed (port change affects systemd environment)
	if needsServiceRestart && app.ServiceName != "" {
		// Use systemctl restart via RunSudo
		if _, err := runner.RunSudo(fmt.Sprintf("systemctl restart %s", app.ServiceName), sudoPassword); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Service restart failed: " + err.Error()))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  Service restarted"))
		}
	}

	// Save updated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Summary
	fmt.Println()
	fmt.Println(theme.RenderBanner("UPDATED"))
	fmt.Println()
	for _, change := range changes {
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("*"), change)
	}
	fmt.Println()

	newProtocol := "http"
	if app.SSLEnabled {
		newProtocol = "https"
	}
	fmt.Printf("  %s %s://%s\n", theme.SuccessStyle.Render("URL:"), newProtocol, app.Domain)
	fmt.Println()

	return nil
}
