package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	dnsPointSSL    bool
	dnsPointServer string
)

var dnsPointCmd = &cobra.Command{
	Use:   "point <domain> <ip>",
	Short: "Point domain to IP via Vercel API + auto-SSL",
	Long: `Point a domain to an IP address using the Vercel DNS API, then
optionally set up Let's Encrypt SSL with nginx reverse proxy on the server.

Uses the Vercel token from ~/.dns-config.json (no vercel CLI needed).

Examples:
  anime dns point comfort.mydomain.com 192.168.1.100
  anime dns point comfort.mydomain.com 192.168.1.100 --ssl
  anime dns point comfort.mydomain.com 192.168.1.100 --ssl --server captain`,
	Args: cobra.ExactArgs(2),
	RunE: runDNSPoint,
}

func init() {
	dnsPointCmd.Flags().BoolVar(&dnsPointSSL, "ssl", false, "Set up Let's Encrypt SSL + nginx reverse proxy on target")
	dnsPointCmd.Flags().StringVarP(&dnsPointServer, "server", "s", "", "Server alias for SSH (default: use IP directly)")
	dnsCmd.AddCommand(dnsPointCmd)
}

func runDNSPoint(cmd *cobra.Command, args []string) error {
	domain := args[0]
	ip := args[1]

	// Load Vercel token
	cfg, err := loadDNSConfig()
	if err != nil {
		return fmt.Errorf("load dns config: %w", err)
	}
	if cfg.Token == "" {
		fmt.Println(theme.ErrorStyle.Render("✗ No Vercel token in ~/.dns-config.json"))
		fmt.Println(theme.DimTextStyle.Render("  Set \"token\": \"vck_...\" in ~/.dns-config.json"))
		return fmt.Errorf("missing vercel token")
	}

	fmt.Println()
	fmt.Printf("  %s  Point %s → %s\n",
		theme.InfoStyle.Render("Step 1/3"),
		theme.HighlightStyle.Render(domain),
		theme.InfoStyle.Render(ip))
	fmt.Printf("  %s  %s\n",
		theme.DimTextStyle.Render("       "),
		theme.DimTextStyle.Render("Setting A record via Vercel API"))
	fmt.Println()

	// Parse domain into zone + subdomain
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid domain: %s", domain)
	}
	// Zone is last two parts (e.g., mydomain.com)
	zone := strings.Join(parts[len(parts)-2:], ".")
	subdomain := ""
	if len(parts) > 2 {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}

	// Remove existing A records for this subdomain
	if err := vercelRemoveARecords(cfg.Token, cfg.TeamID, zone, subdomain); err != nil {
		fmt.Printf("  │ %s (may not exist yet)\n", theme.DimTextStyle.Render(err.Error()))
	}

	// Create A record
	if err := vercelCreateARecord(cfg.Token, cfg.TeamID, zone, subdomain, ip); err != nil {
		fmt.Printf("  %s  %s\n", theme.ErrorStyle.Render("✗"), err.Error())
		return err
	}

	displayDomain := domain
	if subdomain == "" {
		displayDomain = zone
	}
	fmt.Printf("  %s  %s → %s\n",
		theme.SuccessStyle.Render("✓"),
		theme.SuccessStyle.Render(displayDomain),
		ip)
	fmt.Println()

	// Step 2: SSL setup
	if !dnsPointSSL {
		fmt.Printf("  %s  Skipped (use --ssl to enable)\n",
			theme.DimTextStyle.Render("Step 2/3"))
		fmt.Printf("  %s  Skipped\n",
			theme.DimTextStyle.Render("Step 3/3"))
		fmt.Println()
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render("Done:"),
			theme.DimTextStyle.Render("DNS pointed, no SSL"))
		fmt.Println()
		return nil
	}

	target := ip
	if dnsPointServer != "" {
		target = dnsPointServer
	}

	fmt.Printf("  %s  Installing certbot + nginx on %s\n",
		theme.InfoStyle.Render("Step 2/3"),
		theme.HighlightStyle.Render(target))
	fmt.Println()

	// Install certbot + nginx and get cert
	sslScript := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

echo "  │ Installing nginx + certbot..."
if [ "$(id -u)" -eq 0 ]; then SUDO=""; else SUDO="sudo"; fi
$SUDO apt-get update -y -qq
$SUDO apt-get install -y -qq nginx certbot python3-certbot-nginx

echo "  │ Configuring nginx reverse proxy for %s → :3000..."
cat <<'NGINX' | $SUDO tee /etc/nginx/sites-available/comfort >/dev/null
server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGINX

$SUDO ln -sf /etc/nginx/sites-available/comfort /etc/nginx/sites-enabled/comfort
$SUDO rm -f /etc/nginx/sites-enabled/default
$SUDO nginx -t && $SUDO systemctl reload nginx

echo "  │ Requesting Let's Encrypt certificate..."
$SUDO certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email --redirect

echo "  │ Enabling auto-renewal..."
$SUDO systemctl enable certbot.timer 2>/dev/null || true

echo "  │ Done — https://%s is live"
`, domain, domain, domain, domain)

	sshTarget := target
	if !strings.Contains(target, "@") && target != ip {
		// It's a server alias, resolve via anime config
		sshTarget = target
	}

	sshArgs := buildSSHArgs(sshTarget, sslScript)
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		fmt.Printf("  %s  SSL setup failed: %s\n", theme.ErrorStyle.Render("✗"), err.Error())
		fmt.Println(theme.DimTextStyle.Render("    DNS is pointed, but SSL needs manual setup"))
		return err
	}

	fmt.Println()
	fmt.Printf("  %s  SSL certificate issued\n",
		theme.SuccessStyle.Render("Step 2/3"))
	fmt.Println()
	fmt.Printf("  %s  Verifying...\n",
		theme.InfoStyle.Render("Step 3/3"))

	fmt.Println()
	fmt.Printf("  %s  %s is live\n",
		theme.SuccessStyle.Render("✓"),
		theme.SuccessStyle.Render("https://"+domain))
	fmt.Println()

	return nil
}

// --- Vercel API helpers (no CLI needed) ---

type vercelDNSRecord struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type vercelDNSListResponse struct {
	Records []vercelDNSRecord `json:"records"`
}

func vercelAPI(token, method, path string, body io.Reader) ([]byte, error) {
	url := "https://api.vercel.com" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return data, fmt.Errorf("vercel API %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func vercelRemoveARecords(token, teamID, zone, subdomain string) error {
	path := fmt.Sprintf("/v4/domains/%s/records", zone)
	// Try without teamId first (personal tokens), fall back to teamId
	if teamID != "" {
		_, err := vercelAPI(token, "GET", path, nil)
		if err != nil {
			path += "?teamId=" + teamID
		}
	}

	data, err := vercelAPI(token, "GET", path, nil)
	if err != nil {
		return err
	}

	var resp vercelDNSListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	name := subdomain
	if name == "" {
		name = ""
	}

	for _, rec := range resp.Records {
		if rec.Type == "A" && rec.Name == name {
			delPath := fmt.Sprintf("/v2/domains/%s/records/%s", zone, rec.ID)
			if teamID != "" {
				delPath += "?teamId=" + teamID
			}
			fmt.Printf("  │ Removing old A record (%s → %s)\n",
				theme.DimTextStyle.Render(rec.Name),
				theme.DimTextStyle.Render(rec.Value))
			vercelAPI(token, "DELETE", delPath, nil)
		}
	}
	return nil
}

func vercelCreateARecord(token, teamID, zone, subdomain, ip string) error {
	path := fmt.Sprintf("/v2/domains/%s/records", zone)
	// Try without teamId first; add only if personal scope fails
	if teamID != "" {
		testPath := fmt.Sprintf("/v4/domains/%s/records", zone)
		_, err := vercelAPI(token, "GET", testPath, nil)
		if err != nil {
			path += "?teamId=" + teamID
		}
	}

	payload := map[string]interface{}{
		"name":  subdomain,
		"type":  "A",
		"value": ip,
		"ttl":   60,
	}
	body, _ := json.Marshal(payload)

	_, err := vercelAPI(token, "POST", path, bytes.NewReader(body))
	return err
}

