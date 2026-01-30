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
