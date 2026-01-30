package launch

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type               string // "oauth2", "basic", "none"
	GoogleClientID     string
	GoogleClientSecret string
	CookieSecret       string
	EmailDomain        string // e.g., "*" for any Google account
	Username           string // basic auth
	Password           string // basic auth
}

// GenerateCookieSecret creates a random 32-byte cookie secret
func GenerateCookieSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// InstallOAuth2Proxy installs the oauth2-proxy binary if not present
func InstallOAuth2Proxy(password string, runner CommandRunner) error {
	// Check if already installed
	out, _ := runner.Run("which oauth2-proxy 2>/dev/null")
	if out != "" {
		return nil
	}

	installCmd := `cd /tmp && \
curl -sL https://github.com/oauth2-proxy/oauth2-proxy/releases/download/v7.6.0/oauth2-proxy-v7.6.0.linux-amd64.tar.gz | tar xz && \
mv oauth2-proxy-v7.6.0.linux-amd64/oauth2-proxy /usr/local/bin/ && \
chmod +x /usr/local/bin/oauth2-proxy && \
rm -rf oauth2-proxy-v7.6.0.linux-amd64`

	if _, err := runner.RunSudo(installCmd, password); err != nil {
		return fmt.Errorf("failed to install oauth2-proxy: %w", err)
	}
	return nil
}

const oauth2ProxyTemplate = `[Unit]
Description=OAuth2 Proxy for %s
After=network-online.target

[Service]
ExecStart=/usr/local/bin/oauth2-proxy \
    --provider=google \
    --client-id=%s \
    --client-secret=%s \
    --cookie-secret=%s \
    --email-domain=%s \
    --http-address=0.0.0.0:4180 \
    --redirect-url=https://%s/oauth2/callback \
    --cookie-secure=true \
    --reverse-proxy=true \
    --set-xauthrequest=true
User=%s
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
`

// GenerateOAuth2ProxyUnit returns a systemd unit for oauth2-proxy
func GenerateOAuth2ProxyUnit(appName, domain, user string, auth AuthConfig) string {
	return fmt.Sprintf(oauth2ProxyTemplate,
		appName,
		auth.GoogleClientID,
		auth.GoogleClientSecret,
		auth.CookieSecret,
		auth.EmailDomain,
		domain,
		user,
	)
}

// InstallOAuth2ProxyService installs and starts the OAuth2 proxy systemd service
func InstallOAuth2ProxyService(appName, content, password string, runner CommandRunner) error {
	serviceName := "anime-" + appName + "-oauth2"
	return InstallSystemdUnit(serviceName, content, password, runner)
}

// CreateHtpasswd creates an htpasswd file for basic auth
func CreateHtpasswd(appName, username, password, sudoPassword string, runner CommandRunner) error {
	htpasswdFile := fmt.Sprintf("/etc/nginx/.htpasswd-%s", appName)

	// Try htpasswd command first, fall back to openssl
	cmd := fmt.Sprintf(
		`which htpasswd >/dev/null 2>&1 && htpasswd -cb %s '%s' '%s' || echo '%s:'$(openssl passwd -apr1 '%s') > %s`,
		htpasswdFile, username, password,
		username, password, htpasswdFile,
	)
	if _, err := runner.RunSudo(cmd, sudoPassword); err != nil {
		return fmt.Errorf("failed to create htpasswd: %w", err)
	}
	return nil
}
