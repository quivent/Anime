package launch

import (
	"fmt"
	"strings"
)

// OAuth2Provider represents supported OAuth2 providers
type OAuth2Provider string

const (
	OAuth2ProviderGoogle OAuth2Provider = "google"
	OAuth2ProviderGitHub OAuth2Provider = "github"
	OAuth2ProviderOIDC   OAuth2Provider = "oidc"
)

// OAuth2ProxyConfig holds all configuration for oauth2-proxy
type OAuth2ProxyConfig struct {
	AppName      string
	Domain       string
	User         string
	Provider     OAuth2Provider
	ClientID     string
	ClientSecret string
	CookieSecret string

	// Google-specific
	EmailDomain string

	// GitHub-specific
	Org  string
	Team string

	// OIDC-specific
	IssuerURL string
	Scopes    []string

	// Session
	CookieTTL    string
	CookieSecure bool
	Port         int
}

// GenerateOAuth2ProxyArgs generates command line arguments for oauth2-proxy
func GenerateOAuth2ProxyArgs(cfg *OAuth2ProxyConfig) []string {
	port := cfg.Port
	if port == 0 {
		port = 4180
	}

	args := []string{
		fmt.Sprintf("--provider=%s", cfg.Provider),
		fmt.Sprintf("--client-id=%s", cfg.ClientID),
		fmt.Sprintf("--client-secret=%s", cfg.ClientSecret),
		fmt.Sprintf("--cookie-secret=%s", cfg.CookieSecret),
		fmt.Sprintf("--http-address=0.0.0.0:%d", port),
		fmt.Sprintf("--redirect-url=https://%s/oauth2/callback", cfg.Domain),
		"--cookie-secure=true",
		"--reverse-proxy=true",
		"--set-xauthrequest=true",
	}

	switch cfg.Provider {
	case OAuth2ProviderGoogle:
		if cfg.EmailDomain != "" {
			args = append(args, fmt.Sprintf("--email-domain=%s", cfg.EmailDomain))
		}

	case OAuth2ProviderGitHub:
		if cfg.Org != "" {
			args = append(args, fmt.Sprintf("--github-org=%s", cfg.Org))
		}
		if cfg.Team != "" {
			args = append(args, fmt.Sprintf("--github-team=%s", cfg.Team))
		}

	case OAuth2ProviderOIDC:
		if cfg.IssuerURL != "" {
			args = append(args, fmt.Sprintf("--oidc-issuer-url=%s", cfg.IssuerURL))
		}
		if len(cfg.Scopes) > 0 {
			args = append(args, fmt.Sprintf("--scope=%s", strings.Join(cfg.Scopes, " ")))
		} else {
			args = append(args, "--scope=openid email profile")
		}
	}

	if cfg.CookieTTL != "" {
		args = append(args, fmt.Sprintf("--cookie-expire=%s", cfg.CookieTTL))
	}

	return args
}

// GenerateOAuth2ProxySystemdUnit generates a systemd unit file for oauth2-proxy
func GenerateOAuth2ProxySystemdUnit(cfg *OAuth2ProxyConfig) string {
	args := GenerateOAuth2ProxyArgs(cfg)
	argsStr := strings.Join(args, " \\\n    ")

	return fmt.Sprintf(`[Unit]
Description=OAuth2 Proxy for %s
After=network-online.target

[Service]
ExecStart=/usr/local/bin/oauth2-proxy \
    %s
User=%s
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
`, cfg.AppName, argsStr, cfg.User)
}

// OAuth2NginxBlocks generates nginx configuration blocks for OAuth2
type OAuth2NginxBlocks struct {
	// AuthEndpoints - location blocks for /oauth2/ endpoints
	AuthEndpoints string
	// AuthRequest - auth_request directive for protected locations
	AuthRequest string
}

// GenerateOAuth2NginxBlocks generates nginx blocks for OAuth2 integration
func GenerateOAuth2NginxBlocks(port int) *OAuth2NginxBlocks {
	if port == 0 {
		port = 4180
	}

	authEndpoints := fmt.Sprintf(`
    # OAuth2 proxy endpoints
    location /oauth2/ {
        proxy_pass http://127.0.0.1:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location = /oauth2/auth {
        proxy_pass http://127.0.0.1:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Content-Length "";
        proxy_pass_request_body off;
    }

    location /oauth2/callback {
        proxy_pass http://127.0.0.1:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
`, port, port, port)

	authRequest := `        auth_request /oauth2/auth;
        error_page 401 = /oauth2/sign_in;
        auth_request_set $user $upstream_http_x_auth_request_user;
        auth_request_set $email $upstream_http_x_auth_request_email;
        auth_request_set $groups $upstream_http_x_auth_request_groups;
        proxy_set_header X-User $user;
        proxy_set_header X-Email $email;
        proxy_set_header X-Groups $groups;
`

	return &OAuth2NginxBlocks{
		AuthEndpoints: authEndpoints,
		AuthRequest:   authRequest,
	}
}

// OAuth2ProviderInfo contains information about an OAuth2 provider
type OAuth2ProviderInfo struct {
	Name           string
	AuthURL        string
	TokenURL       string
	UserInfoURL    string
	ScopesRequired []string
	SetupURL       string
}

// OAuth2Providers contains info about supported providers
var OAuth2Providers = map[OAuth2Provider]OAuth2ProviderInfo{
	OAuth2ProviderGoogle: {
		Name:           "Google",
		AuthURL:        "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:       "https://oauth2.googleapis.com/token",
		UserInfoURL:    "https://openidconnect.googleapis.com/v1/userinfo",
		ScopesRequired: []string{"openid", "email", "profile"},
		SetupURL:       "https://console.cloud.google.com/apis/credentials",
	},
	OAuth2ProviderGitHub: {
		Name:           "GitHub",
		AuthURL:        "https://github.com/login/oauth/authorize",
		TokenURL:       "https://github.com/login/oauth/access_token",
		UserInfoURL:    "https://api.github.com/user",
		ScopesRequired: []string{"user:email", "read:org"},
		SetupURL:       "https://github.com/settings/developers",
	},
	OAuth2ProviderOIDC: {
		Name:           "OpenID Connect",
		ScopesRequired: []string{"openid", "email", "profile"},
	},
}

// GetOAuth2ProviderInfo returns info about a provider
func GetOAuth2ProviderInfo(provider OAuth2Provider) *OAuth2ProviderInfo {
	if info, ok := OAuth2Providers[provider]; ok {
		return &info
	}
	return nil
}

// ValidateOAuth2Config validates an OAuth2 configuration
func ValidateOAuth2Config(cfg *OAuth2Config) error {
	if cfg == nil {
		return fmt.Errorf("OAuth2 config is nil")
	}
	if cfg.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}
	if cfg.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	switch OAuth2Provider(cfg.Provider) {
	case OAuth2ProviderGoogle:
		// Google requires email domain
	case OAuth2ProviderGitHub:
		// GitHub optionally requires org
	case OAuth2ProviderOIDC:
		if cfg.IssuerURL == "" {
			return fmt.Errorf("issuer URL is required for OIDC provider")
		}
	default:
		return fmt.Errorf("unknown provider: %s", cfg.Provider)
	}

	return nil
}
