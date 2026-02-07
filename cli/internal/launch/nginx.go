package launch

import (
	"bytes"
	"fmt"
	"text/template"
)

// NginxConfig holds parameters for nginx config generation
type NginxConfig struct {
	Domain    string
	Port      int
	AppName   string
	AuthType  string // "oauth2", "basic", "none"
}

const nginxBaseTemplate = `server {
    listen 80;
    server_name {{.Domain}};

    client_max_body_size 100M;
{{.AuthBlock}}
    location / {
{{.LocationAuthBlock}}        proxy_pass http://127.0.0.1:{{.Port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }
}
`

type nginxTemplateData struct {
	Domain             string
	Port               int
	AuthBlock          string
	LocationAuthBlock  string
}

// GenerateNginxConfig returns nginx site configuration
func GenerateNginxConfig(cfg NginxConfig) (string, error) {
	data := nginxTemplateData{
		Domain: cfg.Domain,
		Port:   cfg.Port,
	}

	switch cfg.AuthType {
	case "oauth2":
		data.AuthBlock = oauth2AuthBlock()
		data.LocationAuthBlock = oauth2LocationBlock()
	case "basic":
		data.AuthBlock = ""
		data.LocationAuthBlock = basicAuthLocationBlock(cfg.AppName)
	default:
		data.AuthBlock = ""
		data.LocationAuthBlock = ""
	}

	tmpl, err := template.New("nginx").Parse(nginxBaseTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse nginx template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute nginx template: %w", err)
	}

	return buf.String(), nil
}

func oauth2AuthBlock() string {
	return `
    # OAuth2 proxy endpoints
    location /oauth2/ {
        proxy_pass http://127.0.0.1:4180;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /oauth2/callback {
        proxy_pass http://127.0.0.1:4180;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
`
}

func oauth2LocationBlock() string {
	return `        auth_request /oauth2/auth;
        error_page 401 = /oauth2/sign_in;
        auth_request_set $user $upstream_http_x_auth_request_user;
        auth_request_set $email $upstream_http_x_auth_request_email;
        proxy_set_header X-User $user;
        proxy_set_header X-Email $email;
`
}

func basicAuthLocationBlock(appName string) string {
	return fmt.Sprintf(`        auth_basic "Restricted";
        auth_basic_user_file /etc/nginx/.htpasswd-%s;
`, appName)
}

// InstallNginxConfig writes the config, creates symlink, tests, and reloads
func InstallNginxConfig(appName, content, password string, runner CommandRunner) error {
	sitesAvailable := fmt.Sprintf("/etc/nginx/sites-available/%s", appName)
	sitesEnabled := fmt.Sprintf("/etc/nginx/sites-enabled/%s", appName)

	// Write config file
	writeCmd := fmt.Sprintf("cat > %s << 'NGINXEOF'\n%s\nNGINXEOF", sitesAvailable, content)
	if _, err := runner.RunSudo(writeCmd, password); err != nil {
		return fmt.Errorf("failed to write nginx config: %w", err)
	}

	// Create symlink
	if _, err := runner.RunSudo(fmt.Sprintf("ln -sf %s %s", sitesAvailable, sitesEnabled), password); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Test config
	out, err := runner.RunSudo("nginx -t 2>&1", password)
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", out)
	}

	// Reload
	if _, err := runner.RunSudo("systemctl reload nginx", password); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// SetupSSL runs certbot for the domain
func SetupSSL(domain, email, password string, runner CommandRunner) error {
	cmd := fmt.Sprintf("certbot --nginx -d %s --non-interactive --agree-tos -m %s", domain, email)
	out, err := runner.RunSudo(cmd, password)
	if err != nil {
		return fmt.Errorf("certbot failed: %s: %w", out, err)
	}
	return nil
}

// ExpandedNginxConfig holds full auth configuration for nginx
type ExpandedNginxConfig struct {
	Domain   string
	Port     int
	AppName  string
	Auth     *AuthConfig
}

// ExpandedNginxBlocks contains all generated nginx blocks
type ExpandedNginxBlocks struct {
	HTTPContext     string // rate limit zones (goes in http context)
	ServerLocations string // auth service locations
	LocationPre     string // IP access, rate limiting
	LocationAuth    string // auth_request, auth_basic
	LocationPost    string // security headers
}

// GenerateExpandedNginxConfig generates nginx config with full auth support
func GenerateExpandedNginxConfig(cfg ExpandedNginxConfig) (string, string, error) {
	blocks := GenerateNginxAuthBlocks(cfg.AppName, cfg.Auth)

	// Build template data
	data := struct {
		Domain          string
		Port            int
		ServerLocations string
		LocationBlocks  string
	}{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		ServerLocations: blocks.ServerLocations,
		LocationBlocks:  blocks.LocationPre + blocks.LocationAuth + blocks.LocationPost,
	}

	// Extended template
	tmpl := `server {
    listen 80;
    server_name {{.Domain}};

    client_max_body_size 100M;
{{.ServerLocations}}
    location / {
{{.LocationBlocks}}        proxy_pass http://127.0.0.1:{{.Port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }
}
`

	t, err := template.New("nginx").Parse(tmpl)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), blocks.HTTPContext, nil
}

// GenerateNginxAuthBlocks generates all auth-related nginx blocks
func GenerateNginxAuthBlocks(appName string, auth *AuthConfig) *ExpandedNginxBlocks {
	blocks := &ExpandedNginxBlocks{}

	if auth == nil {
		return blocks
	}

	// Rate limiting (HTTP context)
	if auth.RateLimit != nil && auth.RateLimit.RequestsPerSec > 0 {
		blocks.HTTPContext = GenerateRateLimitZone(auth.RateLimit, appName)
	}

	// IP access control (location pre)
	if auth.IPAccess != nil && len(auth.IPAccess.CIDRs) > 0 {
		blocks.LocationPre += GenerateIPAccessBlock(auth.IPAccess)
	}

	// Rate limiting directive (location pre)
	if auth.RateLimit != nil && auth.RateLimit.RequestsPerSec > 0 {
		blocks.LocationPre += GenerateRateLimitDirective(auth.RateLimit, appName)
	}

	// OAuth2 (any provider)
	if auth.HasOAuth2() && auth.OAuth2 != nil {
		oauthBlocks := GenerateOAuth2NginxBlocks(4180)
		blocks.ServerLocations += oauthBlocks.AuthEndpoints
		blocks.LocationAuth += oauthBlocks.AuthRequest
	}

	// Web3/SIWE
	if auth.HasMethod(AuthMethodWeb3SIWE) && auth.Web3 != nil {
		port := auth.Web3.ServicePort
		if port == 0 {
			port = 4181
		}
		siweBlocks := GenerateSIWENginxBlocks(port)
		blocks.ServerLocations += siweBlocks.ServiceLocation
		blocks.LocationAuth += siweBlocks.AuthRequest
	}

	// Basic auth
	if auth.HasMethod(AuthMethodBasic) {
		blocks.LocationAuth += basicAuthLocationBlock(appName)
	}

	// API key (uses auth_request to validation service)
	if auth.HasMethod(AuthMethodAPIKey) && auth.APIKey != nil {
		blocks.ServerLocations += GenerateAPIKeyNginxAuthBlock(4182, auth.APIKey.HeaderName)
		blocks.LocationAuth += GenerateAPIKeyLocationBlock()
	}

	// Security headers
	if auth.Security != nil {
		blocks.LocationPost += generateSecurityHeaders(auth.Security)
	}

	return blocks
}

// generateSecurityHeaders generates nginx security header directives
func generateSecurityHeaders(cfg *SecurityHeadersConfig) string {
	if cfg == nil {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("        # Security Headers\n")

	if cfg.HSTS {
		maxAge := cfg.HSTSMaxAge
		if maxAge == 0 {
			maxAge = 31536000 // 1 year
		}
		subs := ""
		if cfg.HSTSIncludeSubs {
			subs = "; includeSubDomains"
		}
		buf.WriteString(fmt.Sprintf("        add_header Strict-Transport-Security \"max-age=%d%s\" always;\n", maxAge, subs))
	}

	if cfg.CSP != "" {
		buf.WriteString(fmt.Sprintf("        add_header Content-Security-Policy \"%s\" always;\n", cfg.CSP))
	}

	if cfg.XFrameOptions != "" {
		buf.WriteString(fmt.Sprintf("        add_header X-Frame-Options \"%s\" always;\n", cfg.XFrameOptions))
	}

	if cfg.XContentType {
		buf.WriteString("        add_header X-Content-Type-Options \"nosniff\" always;\n")
	}

	if cfg.ReferrerPolicy != "" {
		buf.WriteString(fmt.Sprintf("        add_header Referrer-Policy \"%s\" always;\n", cfg.ReferrerPolicy))
	}

	return buf.String()
}

// InstallRateLimitZone installs rate limit zone in nginx http context
func InstallRateLimitZone(appName, zoneConfig, password string, runner CommandRunner) error {
	if zoneConfig == "" {
		return nil
	}

	// Write to a separate rate limit config file
	configPath := fmt.Sprintf("/etc/nginx/conf.d/%s-ratelimit.conf", appName)
	writeCmd := fmt.Sprintf("cat > %s << 'EOF'\n%sEOF", configPath, zoneConfig)

	if _, err := runner.RunSudo(writeCmd, password); err != nil {
		return fmt.Errorf("failed to write rate limit config: %w", err)
	}

	return nil
}

// RemoveRateLimitZone removes rate limit zone config
func RemoveRateLimitZone(appName, password string, runner CommandRunner) error {
	configPath := fmt.Sprintf("/etc/nginx/conf.d/%s-ratelimit.conf", appName)
	_, _ = runner.RunSudo(fmt.Sprintf("rm -f %s", configPath), password)
	return nil
}
