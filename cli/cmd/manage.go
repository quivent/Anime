package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var manageServer string

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage sites, domains, nginx, and SSL",
	Long:  "Server management commands for sites, domains, nginx configs, and SSL certificates.",
	Run:   runManageHelp,
}

// ── sites ──

var manageSitesCmd = &cobra.Command{
	Use:     "sites",
	Aliases: []string{"ls"},
	Short:   "List all nginx sites",
	RunE:    runManageSites,
}

var manageSiteAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a new site with nginx + SSL wizard",
	Long: `Interactive wizard to set up a new site:

  1. Domain name
  2. Upstream type (static, reverse proxy, redirect)
  3. Upstream port or path
  4. SSL via Let's Encrypt (certbot)
  5. Enable site

Examples:
  anime manage add api.example.com
  anime manage add api.example.com -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runManageSiteAdd,
}

var manageSiteRemoveCmd = &cobra.Command{
	Use:   "remove <domain>",
	Short: "Remove a site and its nginx config",
	Args:  cobra.ExactArgs(1),
	RunE:  runManageSiteRemove,
}

var manageSiteEnableCmd = &cobra.Command{
	Use:   "enable <domain>",
	Short: "Enable a disabled site",
	Args:  cobra.ExactArgs(1),
	RunE:  runManageSiteEnable,
}

var manageSiteDisableCmd = &cobra.Command{
	Use:   "disable <domain>",
	Short: "Disable a site without removing it",
	Args:  cobra.ExactArgs(1),
	RunE:  runManageSiteDisable,
}

// ── ssl ──

var manageSSLCmd = &cobra.Command{
	Use:   "ssl <domain>",
	Short: "Set up SSL certificate via Let's Encrypt",
	Long: `Request and install an SSL certificate using certbot.

Examples:
  anime manage ssl api.example.com
  anime manage ssl api.example.com -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runManageSSL,
}

// ── nginx ──

var manageNginxCmd = &cobra.Command{
	Use:   "nginx <action>",
	Short: "Nginx management (reload, restart, test, status, logs)",
	Long: `Manage the nginx service.

Actions:
  reload    - Reload nginx config (graceful)
  restart   - Restart nginx service
  test      - Test nginx configuration
  status    - Show nginx status
  logs      - Tail nginx error log

Examples:
  anime manage nginx reload
  anime manage nginx test -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runManageNginx,
}

func init() {
	// Add server flag to all subcommands
	for _, cmd := range []*cobra.Command{
		manageSitesCmd, manageSiteAddCmd, manageSiteRemoveCmd,
		manageSiteEnableCmd, manageSiteDisableCmd,
		manageSSLCmd, manageNginxCmd,
	} {
		cmd.Flags().StringVarP(&manageServer, "server", "s", "", "Run on remote server")
	}

	manageCmd.AddCommand(manageSitesCmd)
	manageCmd.AddCommand(manageSiteAddCmd)
	manageCmd.AddCommand(manageSiteRemoveCmd)
	manageCmd.AddCommand(manageSiteEnableCmd)
	manageCmd.AddCommand(manageSiteDisableCmd)
	manageCmd.AddCommand(manageSSLCmd)
	manageCmd.AddCommand(manageNginxCmd)

	rootCmd.AddCommand(manageCmd)
}

func runManageCmd(script string) (string, error) {
	if manageServer == "" {
		cmd := exec.Command("bash", "-c", script)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	var user, host, sshKey string
	target := cfg.GetAlias(manageServer)
	if target != "" {
		if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			user = parts[0]
			host = parts[1]
		} else {
			user = "ubuntu"
			host = target
		}
	} else {
		server, err := cfg.GetServer(manageServer)
		if err != nil {
			return "", fmt.Errorf("server not found: %s", manageServer)
		}
		user = server.User
		host = server.Host
		sshKey = server.SSHKey
	}

	client, err := ssh.NewClient(host, user, sshKey)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	return client.RunCommand(script)
}

func runManageHelp(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("🌐 MANAGE 🌐"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Server management — sites, domains, nginx, SSL"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime manage sites", "List all nginx sites"},
		{"anime manage add <domain>", "Add new site (interactive wizard)"},
		{"anime manage remove <domain>", "Remove a site"},
		{"anime manage enable <domain>", "Enable a disabled site"},
		{"anime manage disable <domain>", "Disable a site"},
		{"anime manage ssl <domain>", "Set up Let's Encrypt SSL"},
		{"anime manage nginx reload", "Reload nginx config"},
		{"anime manage nginx test", "Test nginx configuration"},
		{"anime manage nginx status", "Show nginx status"},
		{"anime manage nginx logs", "Tail nginx error log"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n\n", theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println(theme.DimTextStyle.Render("  All commands support --server/-s for remote execution"))
	fmt.Println()
}

func runManageSites(cmd *cobra.Command, args []string) error {
	fmt.Println(theme.RenderBanner("🌐 SITES 🌐"))
	fmt.Println()

	script := `
echo "=== ENABLED ==="
ls -1 /etc/nginx/sites-enabled/ 2>/dev/null | grep -v default || echo "(none)"
echo ""
echo "=== AVAILABLE ==="
ls -1 /etc/nginx/sites-available/ 2>/dev/null | grep -v default || echo "(none)"
echo ""
echo "=== DOMAINS ==="
grep -rh 'server_name' /etc/nginx/sites-enabled/ 2>/dev/null | sed 's/.*server_name //;s/;//' | sort -u || echo "(none)"
`

	output, err := runManageCmd(script)
	if err != nil {
		return fmt.Errorf("failed to list sites: %w", err)
	}

	// Parse and display
	sections := strings.Split(output, "===")
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		lines := strings.SplitN(section, "\n", 2)
		if len(lines) < 2 {
			continue
		}
		header := strings.TrimSpace(lines[0])
		body := strings.TrimSpace(lines[1])

		switch header {
		case "ENABLED":
			fmt.Println(theme.SuccessStyle.Render("  Active sites:"))
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && line != "(none)" {
					fmt.Printf("    %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), theme.HighlightStyle.Render(line))
				}
			}
			if body == "(none)" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render("No active sites"))
			}
			fmt.Println()
		case "AVAILABLE":
			fmt.Println(theme.InfoStyle.Render("  Available (including disabled):"))
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && line != "(none)" {
					fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("•"), theme.DimTextStyle.Render(line))
				}
			}
			fmt.Println()
		case "DOMAINS":
			fmt.Println(theme.InfoStyle.Render("  Domains served:"))
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && line != "(none)" && line != "_" {
					fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("🌐"), theme.HighlightStyle.Render(line))
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func runManageSiteAdd(cmd *cobra.Command, args []string) error {
	domain := args[0]
	if err := validate.Domain(domain); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.RenderBanner("🌐 ADD SITE 🌐"))
	fmt.Println()
	fmt.Printf("  Domain: %s\n\n", theme.HighlightStyle.Render(domain))

	// Step 1: Site type
	fmt.Println(theme.InfoStyle.Render("  1. Site type"))
	fmt.Println()
	fmt.Printf("    %s Reverse proxy  %s\n", theme.HighlightStyle.Render("1"), theme.DimTextStyle.Render("(default — forward to localhost port)"))
	fmt.Printf("    %s Static files   %s\n", theme.HighlightStyle.Render("2"), theme.DimTextStyle.Render("(serve from directory)"))
	fmt.Printf("    %s Redirect       %s\n", theme.HighlightStyle.Render("3"), theme.DimTextStyle.Render("(301 redirect to another URL)"))
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	typeChoice, _ := reader.ReadString('\n')
	typeChoice = strings.TrimSpace(typeChoice)
	if typeChoice == "" {
		typeChoice = "1"
	}

	var nginxConfig string

	switch typeChoice {
	case "1": // Reverse proxy
		fmt.Print("  Upstream port [3000]: ")
		portInput, _ := reader.ReadString('\n')
		port := strings.TrimSpace(portInput)
		if port == "" {
			port = "3000"
		}
		if err := validate.Port(port); err != nil {
			return err
		}
		fmt.Printf("  %s Proxy to localhost:%s\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess), port)

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

	case "2": // Static
		fmt.Print("  Document root [/var/www/" + domain + "]: ")
		rootInput, _ := reader.ReadString('\n')
		docRoot := strings.TrimSpace(rootInput)
		if docRoot == "" {
			docRoot = "/var/www/" + domain
		}
		fmt.Printf("  %s Serving from %s\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess), docRoot)

		nginxConfig = fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    root %s;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }
}`, domain, docRoot)

	case "3": // Redirect
		fmt.Print("  Redirect to URL: ")
		targetInput, _ := reader.ReadString('\n')
		targetURL := strings.TrimSpace(targetInput)
		if targetURL == "" {
			return fmt.Errorf("redirect URL required")
		}
		fmt.Printf("  %s Redirect to %s\n\n", theme.SuccessStyle.Render(theme.SymbolSuccess), targetURL)

		nginxConfig = fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    return 301 %s$request_uri;
}`, domain, targetURL)
	}

	// Step 2: SSL
	fmt.Println(theme.InfoStyle.Render("  2. SSL certificate"))
	fmt.Print("  Set up Let's Encrypt SSL? [Y/n]: ")
	sslChoice, _ := reader.ReadString('\n')
	sslChoice = strings.TrimSpace(strings.ToLower(sslChoice))
	wantSSL := sslChoice != "n" && sslChoice != "no"

	// Confirmation
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  Domain:  %s\n", theme.HighlightStyle.Render(domain))
	fmt.Printf("  SSL:     %v\n", wantSSL)
	fmt.Println(theme.SuccessStyle.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Print("  Create site? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm == "n" || confirm == "no" {
		fmt.Println("  Cancelled")
		return nil
	}

	// Execute
	fmt.Println()

	// Write nginx config
	fmt.Printf("  %s Writing nginx config...\n", theme.SymbolLoading)
	writeScript := fmt.Sprintf(`cat > /tmp/nginx_%s << 'NGINXEOF'
%s
NGINXEOF
sudo mv /tmp/nginx_%s /etc/nginx/sites-available/%s
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
echo "Config written"`, domain, nginxConfig, domain, domain, domain, domain)

	// Create doc root for static sites
	if typeChoice == "2" {
		writeScript += fmt.Sprintf(`
sudo mkdir -p /var/www/%s
sudo chown -R www-data:www-data /var/www/%s`, domain, domain)
	}

	output, err := runManageCmd(writeScript)
	if err != nil {
		fmt.Printf("  %s %s\n", theme.ErrorStyle.Render(theme.SymbolError), theme.DimTextStyle.Render(output))
		return fmt.Errorf("failed to write config: %w", err)
	}
	fmt.Printf("  %s Config written\n", theme.SuccessStyle.Render(theme.SymbolSuccess))

	// Test nginx
	fmt.Printf("  %s Testing nginx config...\n", theme.SymbolLoading)
	output, err = runManageCmd("sudo nginx -t 2>&1")
	if err != nil {
		fmt.Printf("  %s Nginx config test failed:\n", theme.ErrorStyle.Render(theme.SymbolError))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(output))
		return fmt.Errorf("nginx config invalid")
	}
	fmt.Printf("  %s Config valid\n", theme.SuccessStyle.Render(theme.SymbolSuccess))

	// Reload nginx
	fmt.Printf("  %s Reloading nginx...\n", theme.SymbolLoading)
	_, err = runManageCmd("sudo systemctl reload nginx")
	if err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}
	fmt.Printf("  %s Nginx reloaded\n", theme.SuccessStyle.Render(theme.SymbolSuccess))

	// SSL
	if wantSSL {
		fmt.Printf("  %s Requesting SSL certificate...\n", theme.SymbolLoading)
		sslScript := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, domain)
		output, err = runManageCmd(sslScript)
		if err != nil {
			fmt.Printf("  %s SSL setup failed (site is live on HTTP):\n", theme.WarningStyle.Render("⚠️"))
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(output)))
			fmt.Println()
			fmt.Printf("  %s Run manually: %s\n",
				theme.DimTextStyle.Render("Fix:"),
				theme.HighlightStyle.Render(fmt.Sprintf("sudo certbot --nginx -d %s", domain)))
		} else {
			fmt.Printf("  %s SSL certificate installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ Site %s is live!", domain)))
	fmt.Println()
	return nil
}

func runManageSiteRemove(cmd *cobra.Command, args []string) error {
	domain := args[0]
	if err := validate.Domain(domain); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("  Remove site %s? [y/N]: ", theme.HighlightStyle.Render(domain))
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("  Cancelled")
		return nil
	}

	script := fmt.Sprintf(`
sudo rm -f /etc/nginx/sites-enabled/%s
sudo rm -f /etc/nginx/sites-available/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx
echo "Removed %s"`, domain, domain, domain)

	output, err := runManageCmd(script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s removed\n", theme.SuccessStyle.Render(theme.SymbolSuccess), domain)
	return nil
}

func runManageSiteEnable(cmd *cobra.Command, args []string) error {
	domain := args[0]
	if err := validate.Domain(domain); err != nil {
		return err
	}
	script := fmt.Sprintf(`
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx
echo "Enabled %s"`, domain, domain, domain)

	output, err := runManageCmd(script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s enabled\n", theme.SuccessStyle.Render(theme.SymbolSuccess), domain)
	return nil
}

func runManageSiteDisable(cmd *cobra.Command, args []string) error {
	domain := args[0]
	if err := validate.Domain(domain); err != nil {
		return err
	}
	script := fmt.Sprintf(`
sudo rm -f /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx
echo "Disabled %s"`, domain, domain)

	output, err := runManageCmd(script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s disabled\n", theme.SuccessStyle.Render(theme.SymbolSuccess), domain)
	return nil
}

func runManageSSL(cmd *cobra.Command, args []string) error {
	domain := args[0]
	if err := validate.Domain(domain); err != nil {
		return err
	}

	fmt.Printf("  %s Requesting SSL for %s...\n",
		theme.SymbolLoading, theme.HighlightStyle.Render(domain))

	script := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, domain)
	output, err := runManageCmd(script)
	if err != nil {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(output)))
		return fmt.Errorf("certbot failed: %w", err)
	}

	fmt.Printf("  %s SSL certificate installed for %s\n",
		theme.SuccessStyle.Render(theme.SymbolSuccess), domain)
	return nil
}

func runManageNginx(cmd *cobra.Command, args []string) error {
	action := args[0]

	var script string
	switch action {
	case "reload":
		script = "sudo nginx -t 2>&1 && sudo systemctl reload nginx && echo 'Nginx reloaded'"
	case "restart":
		script = "sudo systemctl restart nginx && echo 'Nginx restarted'"
	case "test":
		script = "sudo nginx -t 2>&1"
	case "status":
		script = "sudo systemctl status nginx --no-pager 2>&1"
	case "logs":
		script = "sudo tail -50 /var/log/nginx/error.log 2>/dev/null || echo 'No error log found'"
	default:
		return fmt.Errorf("unknown action: %s (try: reload, restart, test, status, logs)", action)
	}

	fmt.Printf("  %s nginx %s...\n", theme.SymbolLoading, action)
	output, err := runManageCmd(script)
	if output != "" {
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
	}
	if err != nil && action != "status" {
		return fmt.Errorf("failed: %w", err)
	}
	return nil
}
