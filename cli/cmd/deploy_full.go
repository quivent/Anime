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

var (
	deployFullDomain string
	deployFullPort   string
	deployFullNoSSL  bool
)

var deployFullCmd = &cobra.Command{
	Use:   "deploy <path> <server>",
	Short: "Ship, detect, install deps, start, and configure nginx + SSL",
	Long: `Full deployment pipeline in one command:

  1. Ship files to server
  2. Detect project type
  3. Install dependencies
  4. Start the app
  5. Configure nginx reverse proxy
  6. Set up SSL via Let's Encrypt

Examples:
  anime deploy ./myapp wings --domain api.example.com
  anime deploy ./site wings --domain site.example.com --port 8080
  anime deploy ./static wings --domain docs.example.com --no-ssl`,
	Args: cobra.ExactArgs(2),
	RunE: runDeployFull,
}

func init() {
	deployFullCmd.Flags().StringVarP(&deployFullDomain, "domain", "d", "", "Domain name for nginx/SSL")
	deployFullCmd.Flags().StringVarP(&deployFullPort, "port", "p", "", "App port (auto-detected if not set)")
	deployFullCmd.Flags().BoolVar(&deployFullNoSSL, "no-ssl", false, "Skip SSL setup")
	rootCmd.AddCommand(deployFullCmd)
}

func runDeployFull(cmd *cobra.Command, args []string) error {
	source := args[0]
	server := args[1]
	reader := bufio.NewReader(os.Stdin)

	absPath, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	dirName := filepath.Base(absPath)

	fmt.Println(theme.RenderBanner("🚀 FULL DEPLOY 🚀"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.HighlightStyle.Render(absPath))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(server))

	// Step 1: Detect project locally
	fmt.Println()
	fmt.Printf("  %s Detecting project...\n", theme.SymbolLoading)
	result := launch.AnalyzeProject(absPath)
	if result.Project == nil || result.Project.Type == launch.ProjectUnknown {
		fmt.Printf("  %s Could not detect project type\n", theme.WarningStyle.Render("⚠️"))
		fmt.Println(theme.DimTextStyle.Render("    Continuing with file transfer only..."))
	} else {
		p := result.Project
		fmt.Printf("  %s %s", theme.SuccessStyle.Render(theme.SymbolSuccess), string(p.Type))
		if p.Framework != "" {
			fmt.Printf(" (%s)", p.Framework)
		}
		if p.Port > 0 {
			fmt.Printf(" on port %d", p.Port)
		}
		fmt.Println()
	}

	// Determine port
	port := deployFullPort
	if port == "" && result.Project != nil && result.Project.Port > 0 {
		port = fmt.Sprintf("%d", result.Project.Port)
	}
	if port == "" {
		port = "3000"
	}
	if err := validate.Port(port); err != nil {
		return err
	}

	// Get domain if not provided
	domain := deployFullDomain
	if domain == "" {
		fmt.Print("  Domain (or empty to skip nginx): ")
		domain, _ = reader.ReadString('\n')
		domain = strings.TrimSpace(domain)
	}
	if domain != "" {
		if err := validate.Domain(domain); err != nil {
			return err
		}
	}

	fmt.Println()

	// Step 2: Ship files
	fmt.Printf("  %s Shipping to %s...\n", theme.SymbolLoading, server)
	shipArgs := []string{"ship", absPath, server}
	shipCmd := exec.Command(os.Args[0], shipArgs...)
	shipCmd.Stdout = os.Stdout
	shipCmd.Stderr = os.Stderr
	if err := shipCmd.Run(); err != nil {
		return fmt.Errorf("ship failed: %w", err)
	}

	remotePath := fmt.Sprintf("~/%s", dirName)

	// Step 3: Install deps on remote
	fmt.Printf("\n  %s Installing dependencies...\n", theme.SymbolLoading)
	manageServer = server

	var installScript string
	if result.Project != nil {
		switch result.Project.Type {
		case launch.ProjectNodeJS:
			pm := "npm"
			if result.Project.PackageManager != "" {
				pm = result.Project.PackageManager
			}
			installScript = fmt.Sprintf("cd %s && %s install --production 2>&1 | tail -5", remotePath, pm)
		case launch.ProjectPython:
			installScript = fmt.Sprintf("cd %s && pip3 install -r requirements.txt 2>&1 | tail -5", remotePath)
		case launch.ProjectGo:
			installScript = fmt.Sprintf("cd %s && go build -o server . 2>&1 | tail -5", remotePath)
		case launch.ProjectRust:
			installScript = fmt.Sprintf("cd %s && cargo build --release 2>&1 | tail -5", remotePath)
		}
	}

	if installScript != "" {
		output, err := runOnServer(server, installScript)
		if err != nil {
			fmt.Printf("  %s Dependency install had issues:\n", theme.WarningStyle.Render("⚠️"))
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(output)))
		} else {
			fmt.Printf("  %s Dependencies installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}
	}

	// Step 4: Start the app
	fmt.Printf("  %s Starting app...\n", theme.SymbolLoading)
	var startCmd string
	if result.Project != nil {
		switch result.Project.Type {
		case launch.ProjectNodeJS:
			pm := "npm"
			if result.Project.PackageManager != "" {
				pm = result.Project.PackageManager
			}
			startCmd = fmt.Sprintf("cd %s && %s start", remotePath, pm)
		case launch.ProjectPython:
			fw := result.Project.Framework
			switch fw {
			case "fastapi":
				startCmd = fmt.Sprintf("cd %s && uvicorn main:app --host 0.0.0.0 --port %s", remotePath, port)
			case "django":
				startCmd = fmt.Sprintf("cd %s && python3 manage.py runserver 0.0.0.0:%s", remotePath, port)
			case "flask":
				startCmd = fmt.Sprintf("cd %s && flask run --host 0.0.0.0 --port %s", remotePath, port)
			default:
				startCmd = fmt.Sprintf("cd %s && python3 main.py", remotePath)
			}
		case launch.ProjectGo:
			startCmd = fmt.Sprintf("cd %s && ./server", remotePath)
		case launch.ProjectRust:
			startCmd = fmt.Sprintf("cd %s && ./target/release/%s", remotePath, dirName)
		case launch.ProjectStatic:
			// No app to start for static sites
		}
	}

	if startCmd != "" {
		screenScript := fmt.Sprintf("screen -dmS %s bash -c '%s' && echo 'Started in screen: %s'", dirName, startCmd, dirName)
		output, err := runOnServer(server, screenScript)
		if err != nil {
			fmt.Printf("  %s Start had issues: %s\n", theme.WarningStyle.Render("⚠️"), strings.TrimSpace(output))
		} else {
			fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
		}
	}

	// Step 5: Nginx + SSL
	if domain != "" {
		fmt.Printf("  %s Configuring nginx for %s...\n", theme.SymbolLoading, domain)

		var nginxConfig string
		if result.Project != nil && result.Project.Type == launch.ProjectStatic {
			nginxConfig = fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    root %s;
    index index.html;
    location / { try_files $uri $uri/ =404; }
}`, domain, remotePath)
		} else {
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

		nginxScript := fmt.Sprintf(`cat > /tmp/nginx_%s << 'NGINXEOF'
%s
NGINXEOF
sudo mv /tmp/nginx_%s /etc/nginx/sites-available/%s
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx && echo "nginx configured"`,
			domain, nginxConfig, domain, domain, domain, domain)

		output, err := runOnServer(server, nginxScript)
		if err != nil {
			fmt.Printf("  %s Nginx config failed: %s\n", theme.ErrorStyle.Render(theme.SymbolError), strings.TrimSpace(output))
		} else {
			fmt.Printf("  %s Nginx configured\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}

		// SSL
		if !deployFullNoSSL {
			fmt.Printf("  %s Setting up SSL...\n", theme.SymbolLoading)
			sslScript := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, domain)
			output, err = runOnServer(server, sslScript)
			if err != nil {
				fmt.Printf("  %s SSL failed (site live on HTTP): %s\n",
					theme.WarningStyle.Render("⚠️"),
					theme.DimTextStyle.Render("run: sudo certbot --nginx -d "+domain))
			} else {
				fmt.Printf("  %s SSL installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
			}
		}
	}

	// Done
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	if domain != "" {
		protocol := "https"
		if deployFullNoSSL {
			protocol = "http"
		}
		fmt.Printf("  %s %s://%s\n",
			theme.SuccessStyle.Render("✨ DEPLOYED!"),
			protocol, domain)
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✨ DEPLOYED!"))
	}
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	return nil
}
