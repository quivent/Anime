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

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/vercel"
	"github.com/spf13/cobra"
)

var (
	dnsPointSSL    bool
	dnsPointServer string
)

var dnsPointCmd = &cobra.Command{
	Use:   "point <domain> <ip>",
	Short: "Point domain to IP via Vercel API + auto-SSL",
	Long: `Point a domain to an IP address using the Vercel DNS API.

Uses the Vercel token from anime config (set with 'anime dns auth <token>').
No vercel CLI needed.

Examples:
  anime dns point mydomain.com 192.168.1.100
  anime dns point sub.mydomain.com 192.168.1.100
  anime dns point sub.mydomain.com 192.168.1.100 --ssl
  anime dns point sub.mydomain.com 192.168.1.100 --ssl --server lambda`,
	Args: cobra.ExactArgs(2),
	RunE: runDNSPoint,
}

var dnsAuthCmd = &cobra.Command{
	Use:   "auth <vercel-token> [team-id]",
	Short: "Store Vercel API token for DNS management",
	Long: `Store your Vercel API token so anime can manage DNS records.

Get a token at: https://vercel.com/account/tokens

Examples:
  anime dns auth vck_xxxxx
  anime dns auth vck_xxxxx team_yyyyy`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runDNSAuth,
}

func init() {
	dnsPointCmd.Flags().BoolVar(&dnsPointSSL, "ssl", false, "Set up Let's Encrypt SSL + nginx reverse proxy on target")
	dnsPointCmd.Flags().StringVarP(&dnsPointServer, "server", "s", "", "Server alias for SSH (default: use IP directly)")
	dnsCmd.AddCommand(dnsPointCmd)
	dnsCmd.AddCommand(dnsAuthCmd)
}

func runDNSAuth(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.APIKeys.Vercel = args[0]
	if len(args) > 1 {
		cfg.APIKeys.VercelTeamID = args[1]
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Vercel token saved"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  You can now run:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime dns point <domain> <ip>"))
	fmt.Println()
	return nil
}

func runDNSPoint(cmd *cobra.Command, args []string) error {
	domain := args[0]
	ip := args[1]

	// Load token: env > config.yaml > embedded
	token := os.Getenv("VERCEL_TOKEN")
	teamID := os.Getenv("VERCEL_TEAM_ID")

	if token == "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		token = cfg.APIKeys.Vercel
		if teamID == "" {
			teamID = cfg.APIKeys.VercelTeamID
		}
	}
	if token == "" {
		token = vercel.GetToken()
	}
	if teamID == "" {
		teamID = vercel.GetTeamID()
	}

	if token == "" {
		fmt.Println(theme.ErrorStyle.Render("✗ No Vercel token configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Set one with:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime dns auth <vercel-token>"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Get a token at: https://vercel.com/account/tokens"))
		fmt.Println()
		return fmt.Errorf("missing vercel token")
	}

	fmt.Println()
	fmt.Printf("  %s  Point %s → %s\n",
		theme.InfoStyle.Render("Step 1"),
		theme.HighlightStyle.Render(domain),
		theme.InfoStyle.Render(ip))
	fmt.Printf("  %s\n",
		theme.DimTextStyle.Render("         Setting A record via Vercel API"))
	fmt.Println()

	// Parse domain into zone + subdomain
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid domain: %s", domain)
	}
	zone := strings.Join(parts[len(parts)-2:], ".")
	subdomain := ""
	if len(parts) > 2 {
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}

	// Remove existing A records for this subdomain
	if err := vercelRemoveARecords(token, teamID, zone, subdomain); err != nil {
		fmt.Printf("  │ %s (may not exist yet)\n", theme.DimTextStyle.Render(err.Error()))
	}

	// Create A record
	if err := vercelCreateARecord(token, teamID, zone, subdomain, ip); err != nil {
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

	if !dnsPointSSL {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Add --ssl to set up Let's Encrypt + nginx"))
		fmt.Println()
		return nil
	}

	// SSL setup on remote server
	target := ip
	if dnsPointServer != "" {
		target = dnsPointServer
	}

	fmt.Printf("  %s  Installing certbot + nginx on %s\n",
		theme.InfoStyle.Render("Step 2"),
		theme.HighlightStyle.Render(target))
	fmt.Println()

	sslScript := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

echo "  │ Installing nginx + certbot..."
if [ "$(id -u)" -eq 0 ]; then SUDO=""; else SUDO="sudo"; fi
$SUDO apt-get update -y -qq
$SUDO apt-get install -y -qq nginx certbot python3-certbot-nginx

echo "  │ Configuring nginx reverse proxy for %s → :3000..."
cat <<'NGINX' | $SUDO tee /etc/nginx/sites-available/anime-proxy >/dev/null
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

$SUDO ln -sf /etc/nginx/sites-available/anime-proxy /etc/nginx/sites-enabled/anime-proxy
$SUDO rm -f /etc/nginx/sites-enabled/default
$SUDO nginx -t && $SUDO systemctl reload nginx

echo "  │ Requesting Let's Encrypt certificate..."
$SUDO certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email --redirect

echo "  │ Enabling auto-renewal..."
$SUDO systemctl enable certbot.timer 2>/dev/null || true

echo "  │ Done — https://%s is live"
`, domain, domain, domain, domain)

	sshArgs := dnsPointSSHArgs(target, sslScript)
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		fmt.Printf("  %s  SSL setup failed: %s\n", theme.ErrorStyle.Render("✗"), err.Error())
		fmt.Println(theme.DimTextStyle.Render("    DNS is pointed, but SSL needs manual setup"))
		return err
	}

	fmt.Println()
	fmt.Printf("  %s  %s is live\n",
		theme.SuccessStyle.Render("✓"),
		theme.SuccessStyle.Render("https://"+domain))
	fmt.Println()

	return nil
}

// dnsPointSSHArgs builds SSH args for running a script on a remote server.
func dnsPointSSHArgs(target, script string) []string {
	// Resolve server alias if needed
	if !strings.Contains(target, "@") && !strings.Contains(target, ".") {
		cfg, err := config.Load()
		if err == nil {
			if resolved := cfg.GetAlias(target); resolved != "" {
				target = resolved
			}
		}
	}

	return []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		target,
		script,
	}
}

// --- Vercel API helpers ---

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
	if teamID != "" {
		path += "?teamId=" + teamID
	}

	data, err := vercelAPI(token, "GET", path, nil)
	if err != nil {
		return err
	}

	var resp vercelDNSListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, rec := range resp.Records {
		if rec.Type == "A" && rec.Name == subdomain {
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
	if teamID != "" {
		path += "?teamId=" + teamID
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
