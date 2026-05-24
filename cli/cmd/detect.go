package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/launch"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var detectServer string

var detectCmd = &cobra.Command{
	Use:   "detect [path]",
	Short: "Detect what a project is and offer to set it up",
	Long: `Scan a directory, detect the project type and framework,
then offer actionable next steps — including setting it up as a live site.

Examples:
  anime detect                        # Detect current directory
  anime detect ./myapp                # Detect a specific path
  anime detect /home/ubuntu/api -s wings  # Detect on remote server`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDetect,
}

func init() {
	detectCmd.Flags().StringVarP(&detectServer, "server", "s", "", "Detect on remote server")
	rootCmd.AddCommand(detectCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	fmt.Println(theme.RenderBanner("🔍 DETECT 🔍"))
	fmt.Println()

	// Remote detection — set manageServer so runManageCmd works
	if detectServer != "" {
		manageServer = detectServer
		return runDetectRemote(path)
	}

	// Local detection
	return runDetectLocal(path)
}

func runDetectLocal(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not found: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", absPath)
	}

	fmt.Printf("  %s %s\n\n",
		theme.DimTextStyle.Render("Scanning:"),
		theme.HighlightStyle.Render(absPath))

	result := launch.AnalyzeProject(absPath)
	dirName := filepath.Base(absPath)

	if result.Project == nil || result.Project.Type == launch.ProjectUnknown {
		printUnknownProject()
		return nil
	}

	printDetectionResult(result, dirName)
	offerActions(result, absPath, dirName)
	return nil
}

func runDetectRemote(path string) error {
	fmt.Printf("  %s %s on %s\n\n",
		theme.DimTextStyle.Render("Scanning:"),
		theme.HighlightStyle.Render(path),
		theme.InfoStyle.Render(detectServer))

	// Validate path is shell-safe
	if err := validate.ShellSafe(path); err != nil {
		return fmt.Errorf("unsafe path: %w", err)
	}

	// Run detection heuristics over SSH
	script := fmt.Sprintf(`#!/bin/bash
cd "%s" 2>/dev/null || { echo "PATH_NOT_FOUND"; exit 1; }
echo "DIR_NAME=$(basename "$(pwd)")"

# Detect project type
[ -f package.json ] && echo "HAS_PACKAGE_JSON=1"
[ -f go.mod ] && echo "HAS_GO_MOD=1"
[ -f requirements.txt ] && echo "HAS_REQUIREMENTS_TXT=1"
[ -f pyproject.toml ] && echo "HAS_PYPROJECT=1"
[ -f Cargo.toml ] && echo "HAS_CARGO=1"
[ -f Dockerfile ] && echo "HAS_DOCKER=1"
[ -f index.html ] && echo "HAS_INDEX_HTML=1"
[ -f Makefile ] && echo "HAS_MAKEFILE=1"

# Framework detection
{ [ -f next.config.js ] || [ -f next.config.mjs ] || [ -f next.config.ts ]; } && echo "FRAMEWORK=nextjs"
{ [ -f nuxt.config.js ] || [ -f nuxt.config.ts ]; } && echo "FRAMEWORK=nuxt"
{ [ -f vite.config.js ] || [ -f vite.config.ts ]; } && echo "FRAMEWORK=vite"
[ -f astro.config.mjs ] && echo "FRAMEWORK=astro"
grep -q '"express"' package.json 2>/dev/null && echo "FRAMEWORK=express"
grep -q '"fastify"' package.json 2>/dev/null && echo "FRAMEWORK=fastify"
grep -q 'fastapi' requirements.txt 2>/dev/null && echo "FRAMEWORK=fastapi"
grep -q 'django' requirements.txt 2>/dev/null && echo "FRAMEWORK=django"
grep -q 'flask' requirements.txt 2>/dev/null && echo "FRAMEWORK=flask"

# Port detection
grep -oP '"start":\s*"[^"]*--port\s+\K\d+' package.json 2>/dev/null && echo "PORT_FROM_PKG=1"
grep -oP 'PORT\s*[:=]\s*\K\d+' .env 2>/dev/null | head -1

# Package manager
[ -f yarn.lock ] && echo "PKG_MGR=yarn"
[ -f pnpm-lock.yaml ] && echo "PKG_MGR=pnpm"
[ -f bun.lockb ] && echo "PKG_MGR=bun"
[ -f package-lock.json ] && echo "PKG_MGR=npm"

# Check if already running
ss -tlnp 2>/dev/null | grep -oP ':\K\d+(?=\s)' | sort -un | head -5 | while read p; do echo "LISTENING_PORT=$p"; done
`, path)

	output, err := runManageCmd(script)
	if err != nil {
		if strings.Contains(output, "PATH_NOT_FOUND") {
			return fmt.Errorf("path not found on %s: %s", detectServer, path)
		}
		return fmt.Errorf("detection failed: %w", err)
	}

	// Parse results
	vars := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	dirName := vars["DIR_NAME"]
	if dirName == "" {
		dirName = filepath.Base(path)
	}

	// Determine type
	projectType := "unknown"
	typeEmoji := "📦"
	if vars["HAS_PACKAGE_JSON"] == "1" {
		projectType = "Node.js"
		typeEmoji = "🟢"
	} else if vars["HAS_GO_MOD"] == "1" {
		projectType = "Go"
		typeEmoji = "🔵"
	} else if vars["HAS_REQUIREMENTS_TXT"] == "1" || vars["HAS_PYPROJECT"] == "1" {
		projectType = "Python"
		typeEmoji = "🐍"
	} else if vars["HAS_CARGO"] == "1" {
		projectType = "Rust"
		typeEmoji = "🦀"
	} else if vars["HAS_DOCKER"] == "1" {
		projectType = "Docker"
		typeEmoji = "🐳"
	} else if vars["HAS_INDEX_HTML"] == "1" {
		projectType = "Static Site"
		typeEmoji = "📄"
	}

	if projectType == "unknown" {
		printUnknownProject()
		return nil
	}

	fmt.Printf("  %s %s %s\n", typeEmoji,
		theme.HighlightStyle.Render("Type:"),
		theme.SuccessStyle.Render(projectType))

	if fw := vars["FRAMEWORK"]; fw != "" {
		fmt.Printf("  🏗️  %s %s\n",
			theme.HighlightStyle.Render("Framework:"),
			theme.SuccessStyle.Render(fw))
	}

	if pm := vars["PKG_MGR"]; pm != "" {
		fmt.Printf("  📦 %s %s\n",
			theme.HighlightStyle.Render("Pkg mgr:"),
			theme.DimTextStyle.Render(pm))
	}

	if vars["HAS_DOCKER"] == "1" && projectType != "Docker" {
		fmt.Printf("  🐳 %s %s\n",
			theme.HighlightStyle.Render("Docker:"),
			theme.DimTextStyle.Render("Dockerfile found"))
	}

	// Show listening ports
	var ports []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "LISTENING_PORT=") {
			port := strings.TrimPrefix(strings.TrimSpace(line), "LISTENING_PORT=")
			ports = append(ports, port)
		}
	}
	if len(ports) > 0 {
		fmt.Printf("  🌐 %s %s\n",
			theme.HighlightStyle.Render("Listening:"),
			theme.DimTextStyle.Render("ports "+strings.Join(ports, ", ")))
	}

	fmt.Println()

	// Offer remote actions
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.InfoStyle.Render("  What do you want to do?"))
	fmt.Println()

	actions := []struct {
		label string
		desc  string
	}{
		{"Set up as a site", "nginx + domain + SSL"},
		{"Start/restart the app", "run in background"},
		{"View logs", "check what's running"},
		{"Nothing", "just wanted to know what it is"},
	}

	for i, a := range actions {
		fmt.Printf("    %s %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d", i+1)),
			theme.InfoStyle.Render(a.label),
			theme.DimTextStyle.Render(a.desc))
	}
	fmt.Println()
	fmt.Print("  Choice [4]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return detectSetupSite(reader, path, dirName, projectType, vars)
	case "2":
		return detectStartApp(reader, path, dirName, projectType, vars)
	case "3":
		return detectViewLogs(path)
	default:
		fmt.Println()
		return nil
	}
}

func detectSetupSite(reader *bufio.Reader, remotePath, dirName, projectType string, vars map[string]string) error {
	fmt.Println()
	fmt.Print("  Domain: ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain required")
	}
	if err := validate.Domain(domain); err != nil {
		return err
	}

	// Determine config based on project type
	var nginxConfig string

	switch projectType {
	case "Static Site":
		fmt.Printf("  %s Static site — serving files directly\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess))

		nginxConfig = fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    root %s;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }
}`, domain, remotePath)

	default:
		// App server — reverse proxy
		port := "3000"
		if fw := vars["FRAMEWORK"]; fw != "" {
			switch fw {
			case "fastapi":
				port = "8000"
			case "django":
				port = "8000"
			case "flask":
				port = "5000"
			case "nextjs":
				port = "3000"
			case "express", "fastify":
				port = "3000"
			}
		}

		fmt.Printf("  Port [%s]: ", port)
		portInput, _ := reader.ReadString('\n')
		portInput = strings.TrimSpace(portInput)
		if portInput != "" {
			port = portInput
		}
		if err := validate.Port(port); err != nil {
			return err
		}

		fmt.Printf("  %s Reverse proxy to :%s\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess), port)

		nginxConfig = fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}`, domain, port)
	}

	fmt.Print("  Set up SSL? [Y/n]: ")
	sslChoice, _ := reader.ReadString('\n')
	wantSSL := strings.TrimSpace(strings.ToLower(sslChoice)) != "n"

	fmt.Println()
	fmt.Printf("  %s Writing nginx config for %s...\n", theme.SymbolLoading, domain)

	script := fmt.Sprintf(`cat > /tmp/nginx_%s << 'NGINXEOF'
%s
NGINXEOF
sudo mv /tmp/nginx_%s /etc/nginx/sites-available/%s
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx
echo "done"`, domain, nginxConfig, domain, domain, domain, domain)

	output, err := runManageCmd(script)
	if err != nil {
		return fmt.Errorf("nginx setup failed: %s", output)
	}
	fmt.Printf("  %s Site configured\n", theme.SuccessStyle.Render(theme.SymbolSuccess))

	if wantSSL {
		fmt.Printf("  %s Requesting SSL...\n", theme.SymbolLoading)
		sslScript := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, domain)
		output, err = runManageCmd(sslScript)
		if err != nil {
			fmt.Printf("  %s SSL failed — site is live on HTTP\n", theme.WarningStyle.Render("⚠️"))
			fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Run: sudo certbot --nginx -d "+domain))
		} else {
			fmt.Printf("  %s SSL installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}
	}

	fmt.Println()
	fmt.Printf("  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("✨ %s is live!", domain)))
	fmt.Println()
	return nil
}

func detectStartApp(reader *bufio.Reader, remotePath, dirName, projectType string, vars map[string]string) error {
	fmt.Println()

	// Build a start command based on type
	var startCmd string
	switch projectType {
	case "Node.js":
		pm := vars["PKG_MGR"]
		if pm == "" {
			pm = "npm"
		}
		fw := vars["FRAMEWORK"]
		if fw == "nextjs" {
			startCmd = fmt.Sprintf("cd %s && %s run build && %s start", remotePath, pm, pm)
		} else {
			startCmd = fmt.Sprintf("cd %s && %s start", remotePath, pm)
		}
	case "Python":
		fw := vars["FRAMEWORK"]
		switch fw {
		case "fastapi":
			startCmd = fmt.Sprintf("cd %s && uvicorn main:app --host 0.0.0.0 --port 8000", remotePath)
		case "django":
			startCmd = fmt.Sprintf("cd %s && python3 manage.py runserver 0.0.0.0:8000", remotePath)
		case "flask":
			startCmd = fmt.Sprintf("cd %s && flask run --host 0.0.0.0", remotePath)
		default:
			startCmd = fmt.Sprintf("cd %s && python3 main.py", remotePath)
		}
	case "Go":
		startCmd = fmt.Sprintf("cd %s && go build -o server . && ./server", remotePath)
	case "Rust":
		startCmd = fmt.Sprintf("cd %s && cargo build --release && ./target/release/%s", remotePath, dirName)
	case "Docker":
		startCmd = fmt.Sprintf("cd %s && docker compose up -d 2>/dev/null || docker build -t %s . && docker run -d --name %s %s", remotePath, dirName, dirName, dirName)
	default:
		fmt.Println(theme.WarningStyle.Render("  Don't know how to start this project type"))
		return nil
	}

	fmt.Printf("  Command: %s\n", theme.DimTextStyle.Render(startCmd))
	fmt.Print("  Run in screen session? [Y/n]: ")
	screenChoice, _ := reader.ReadString('\n')
	useScreen := strings.TrimSpace(strings.ToLower(screenChoice)) != "n"

	if useScreen {
		startCmd = fmt.Sprintf("screen -dmS %s bash -c '%s'", dirName, startCmd)
	}

	fmt.Printf("\n  %s Starting %s...\n", theme.SymbolLoading, dirName)
	output, err := runManageCmd(startCmd)
	if err != nil {
		fmt.Printf("  %s %s\n", theme.ErrorStyle.Render(theme.SymbolError), theme.DimTextStyle.Render(output))
		return fmt.Errorf("failed to start")
	}

	if useScreen {
		fmt.Printf("  %s Running in screen session '%s'\n", theme.SuccessStyle.Render(theme.SymbolSuccess), dirName)
		fmt.Printf("  %s Attach: screen -r %s\n", theme.DimTextStyle.Render("  "), dirName)
	} else {
		fmt.Printf("  %s Started\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}
	fmt.Println()
	return nil
}

func detectViewLogs(remotePath string) error {
	fmt.Println()
	script := fmt.Sprintf(`
echo "=== Screen sessions ==="
screen -ls 2>/dev/null || echo "(no screen sessions)"
echo ""
echo "=== Processes in %s ==="
ps aux | grep "%s" | grep -v grep || echo "(nothing running)"
echo ""
echo "=== Recent logs ==="
ls -t %s/*.log %s/logs/*.log 2>/dev/null | head -3 | while read f; do
    echo "--- $f ---"
    tail -5 "$f"
done
`, remotePath, remotePath, remotePath, remotePath)

	output, err := runManageCmd(script)
	if err != nil {
		// Non-fatal — show what we got
	}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(line, "===") {
			fmt.Printf("  %s\n", theme.InfoStyle.Render(line))
		} else if strings.HasPrefix(line, "---") {
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(line))
		} else {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
	}
	fmt.Println()
	return nil
}

func printUnknownProject() {
	fmt.Println(theme.WarningStyle.Render("  Could not identify project type"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Looked for: package.json, go.mod, requirements.txt, pyproject.toml, Cargo.toml, Dockerfile"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Suggestions:"))
	fmt.Printf("    %s  %s\n", theme.HighlightStyle.Render("anime init"), theme.DimTextStyle.Render("Generate config from scratch"))
	fmt.Printf("    %s  %s\n", theme.HighlightStyle.Render("anime unpack"), theme.DimTextStyle.Render("Bootstrap server with dev tools"))
	fmt.Println()
}

func printDetectionResult(result *launch.DetectionResult, dirName string) {
	p := result.Project

	typeEmoji := "📦"
	typeLabel := string(p.Type)
	switch p.Type {
	case launch.ProjectNodeJS:
		typeEmoji = "🟢"
		typeLabel = "Node.js"
	case launch.ProjectPython:
		typeEmoji = "🐍"
		typeLabel = "Python"
	case launch.ProjectGo:
		typeEmoji = "🔵"
		typeLabel = "Go"
	case launch.ProjectRust:
		typeEmoji = "🦀"
		typeLabel = "Rust"
	case launch.ProjectDocker:
		typeEmoji = "🐳"
		typeLabel = "Docker"
	case launch.ProjectStatic:
		typeEmoji = "📄"
		typeLabel = "Static Site"
	}

	fmt.Printf("  %s %s %s\n", typeEmoji,
		theme.HighlightStyle.Render("Type:"),
		theme.SuccessStyle.Render(typeLabel))

	if p.Framework != "" {
		fmt.Printf("  🏗️  %s %s\n",
			theme.HighlightStyle.Render("Framework:"),
			theme.SuccessStyle.Render(p.Framework))
	}
	if p.EntryPoint != "" {
		fmt.Printf("  📄 %s %s\n",
			theme.HighlightStyle.Render("Entry:"),
			theme.DimTextStyle.Render(p.EntryPoint))
	}
	if p.RunCommand != "" {
		fmt.Printf("  ▶️  %s %s\n",
			theme.HighlightStyle.Render("Run:"),
			theme.DimTextStyle.Render(p.RunCommand))
	}
	if p.Port > 0 {
		fmt.Printf("  🌐 %s %s\n",
			theme.HighlightStyle.Render("Port:"),
			theme.DimTextStyle.Render(fmt.Sprintf("%d", p.Port)))
	}
	if p.PackageManager != "" {
		fmt.Printf("  📦 %s %s\n",
			theme.HighlightStyle.Render("Pkg mgr:"),
			theme.DimTextStyle.Render(p.PackageManager))
	}
	if p.HasDocker {
		fmt.Printf("  🐳 %s %s\n",
			theme.HighlightStyle.Render("Docker:"),
			theme.DimTextStyle.Render("Dockerfile found"))
	}
	if result.Database != nil && result.Database.NeedsDatabase {
		dbType := string(result.Database.PrimaryType)
		if dbType == "" || dbType == "unknown" {
			dbType = "detected"
		}
		fmt.Printf("  💾 %s %s\n",
			theme.HighlightStyle.Render("Database:"),
			theme.DimTextStyle.Render(dbType))
	}
	if len(result.Issues) > 0 {
		fmt.Println()
		for _, issue := range result.Issues {
			icon := "⚠️"
			if issue.Severity == "error" {
				icon = "❌"
			}
			fmt.Printf("    %s %s\n", icon, theme.DimTextStyle.Render(issue.Message))
		}
	}
}

func offerActions(result *launch.DetectionResult, absPath, dirName string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  What do you want to do?"))
	fmt.Println()
	fmt.Printf("    %s Ship to a server\n", theme.HighlightStyle.Render("1"))
	fmt.Printf("    %s Generate deployment config (anime.yaml)\n", theme.HighlightStyle.Render("2"))
	fmt.Printf("    %s Nothing\n", theme.HighlightStyle.Render("3"))
	fmt.Println()
	fmt.Print("  Choice [3]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("  Server alias: ")
		server, _ := reader.ReadString('\n')
		server = strings.TrimSpace(server)
		if server == "" {
			return
		}
		fmt.Println()
		fmt.Printf("  %s Run: %s\n\n",
			theme.InfoStyle.Render("→"),
			theme.HighlightStyle.Render(fmt.Sprintf("anime ship %s %s", absPath, server)))

		// Actually run it
		shipArgs := []string{"ship", absPath, server}
		shipCmd := exec.Command(os.Args[0], shipArgs...)
		shipCmd.Stdout = os.Stdout
		shipCmd.Stderr = os.Stderr
		shipCmd.Stdin = os.Stdin
		shipCmd.Run()

		fmt.Println()
		fmt.Printf("  %s Then on the server: %s\n",
			theme.InfoStyle.Render("→"),
			theme.HighlightStyle.Render(fmt.Sprintf("anime detect ~/%s -s %s", dirName, server)))

	case "2":
		initArgs := []string{"init", absPath}
		initCmd := exec.Command(os.Args[0], initArgs...)
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		initCmd.Stdin = os.Stdin
		initCmd.Run()
	}

	fmt.Println()
}
