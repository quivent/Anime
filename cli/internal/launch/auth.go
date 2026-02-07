package launch

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// AuthMethod represents a type of authentication
type AuthMethod string

const (
	AuthMethodNone         AuthMethod = "none"
	AuthMethodBasic        AuthMethod = "basic"
	AuthMethodOAuth2Google AuthMethod = "oauth2-google"
	AuthMethodOAuth2GitHub AuthMethod = "oauth2-github"
	AuthMethodOAuth2OIDC   AuthMethod = "oauth2-oidc"
	AuthMethodWeb3SIWE     AuthMethod = "web3-siwe"
	AuthMethodAPIKey       AuthMethod = "api-key"
)

// AuthConfig holds complete authentication configuration
// Supports multiple auth methods that can be combined
type AuthConfig struct {
	// Methods enabled for this app (can combine multiple)
	Methods []AuthMethod `yaml:"methods,omitempty"`

	// OAuth2 configuration (Google, GitHub, OIDC)
	OAuth2 *OAuth2Config `yaml:"oauth2,omitempty"`

	// Basic auth configuration
	Basic *BasicAuthConfig `yaml:"basic,omitempty"`

	// Web3/SIWE configuration
	Web3 *Web3Config `yaml:"web3,omitempty"`

	// API key configuration
	APIKey *APIKeyConfig `yaml:"api_key,omitempty"`

	// IP access control
	IPAccess *IPAccessConfig `yaml:"ip_access,omitempty"`

	// Rate limiting
	RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty"`

	// Security headers
	Security *SecurityHeadersConfig `yaml:"security,omitempty"`

	// Legacy fields for backwards compatibility
	Type               string `yaml:"type,omitempty"` // DEPRECATED: use Methods
	GoogleClientID     string `yaml:"-"`              // DEPRECATED: use OAuth2
	GoogleClientSecret string `yaml:"-"`              // DEPRECATED: use OAuth2
	CookieSecret       string `yaml:"-"`              // DEPRECATED: use OAuth2
	EmailDomain        string `yaml:"-"`              // DEPRECATED: use OAuth2
	Username           string `yaml:"-"`              // DEPRECATED: use Basic
	Password           string `yaml:"-"`              // DEPRECATED: use Basic
}

// OAuth2Config holds OAuth2 provider settings
type OAuth2Config struct {
	Provider     string `yaml:"provider"`                // google, github, oidc
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	CookieSecret string `yaml:"cookie_secret"`

	// Google-specific
	EmailDomain string `yaml:"email_domain,omitempty"` // * for any

	// GitHub-specific
	Org  string `yaml:"org,omitempty"`  // require org membership
	Team string `yaml:"team,omitempty"` // require team membership

	// Generic OIDC
	IssuerURL string   `yaml:"issuer_url,omitempty"`
	Scopes    []string `yaml:"scopes,omitempty"`

	// Session settings
	CookieTTL    string `yaml:"cookie_ttl,omitempty"` // e.g., "168h" for 7 days
	CookieSecure bool   `yaml:"cookie_secure"`
	CookieDomain string `yaml:"cookie_domain,omitempty"`
}

// BasicAuthConfig holds basic auth settings
type BasicAuthConfig struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"` // stored hash, not plaintext
	Realm        string `yaml:"realm,omitempty"`
}

// Web3Config holds SIWE (Sign-In With Ethereum) settings
type Web3Config struct {
	ChainID          int      `yaml:"chain_id"`                     // 1 for mainnet
	AllowedAddresses []string `yaml:"allowed_addresses,omitempty"`  // wallet whitelist
	RequireENS       bool     `yaml:"require_ens,omitempty"`        // require ENS name
	NonceExpiry      string   `yaml:"nonce_expiry,omitempty"`       // e.g., "5m"
	ServicePort      int      `yaml:"service_port"`                 // SIWE verifier port (default 4181)
	SessionTTL       string   `yaml:"session_ttl,omitempty"`        // session duration
}

// APIKeyConfig holds API key authentication settings
type APIKeyConfig struct {
	HeaderName string   `yaml:"header_name"` // e.g., "X-API-Key"
	Keys       []APIKey `yaml:"keys"`
}

// APIKey represents a single API key
type APIKey struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`                 // human-readable
	KeyHash   string   `yaml:"key_hash"`             // SHA-256 hash of key
	Scopes    []string `yaml:"scopes,omitempty"`     // permission scopes
	RateLimit int      `yaml:"rate_limit,omitempty"` // requests per minute
	ExpiresAt string   `yaml:"expires_at,omitempty"`
	CreatedAt string   `yaml:"created_at"`
	LastUsed  string   `yaml:"last_used,omitempty"`
}

// IPAccessConfig holds IP whitelist/blacklist settings
type IPAccessConfig struct {
	Mode  string   `yaml:"mode"`  // "allow" or "deny"
	CIDRs []string `yaml:"cidrs"` // e.g., ["10.0.0.0/8", "192.168.1.0/24"]
}

// RateLimitConfig holds nginx rate limiting settings
type RateLimitConfig struct {
	RequestsPerSec int    `yaml:"requests_per_sec"` // per IP
	BurstSize      int    `yaml:"burst_size"`
	ZoneSize       string `yaml:"zone_size"` // e.g., "10m"
}

// SecurityHeadersConfig holds HTTP security header settings
type SecurityHeadersConfig struct {
	HSTS            bool   `yaml:"hsts"`
	HSTSMaxAge      int    `yaml:"hsts_max_age,omitempty"`       // seconds, default 31536000
	HSTSIncludeSubs bool   `yaml:"hsts_include_subs,omitempty"`
	CSP             string `yaml:"csp,omitempty"`
	XFrameOptions   string `yaml:"x_frame_options,omitempty"`    // DENY, SAMEORIGIN
	XContentType    bool   `yaml:"x_content_type_options"`       // nosniff
	ReferrerPolicy  string `yaml:"referrer_policy,omitempty"`
}

// HasMethod checks if a specific auth method is enabled
func (c *AuthConfig) HasMethod(method AuthMethod) bool {
	for _, m := range c.Methods {
		if m == method {
			return true
		}
	}
	return false
}

// HasOAuth2 returns true if any OAuth2 method is enabled
func (c *AuthConfig) HasOAuth2() bool {
	return c.HasMethod(AuthMethodOAuth2Google) ||
		c.HasMethod(AuthMethodOAuth2GitHub) ||
		c.HasMethod(AuthMethodOAuth2OIDC)
}

// IsEmpty returns true if no authentication is configured
func (c *AuthConfig) IsEmpty() bool {
	if c == nil {
		return true
	}
	return len(c.Methods) == 0 || (len(c.Methods) == 1 && c.Methods[0] == AuthMethodNone)
}

// GetPrimaryMethod returns the first/primary auth method
func (c *AuthConfig) GetPrimaryMethod() AuthMethod {
	if len(c.Methods) == 0 {
		return AuthMethodNone
	}
	return c.Methods[0]
}

// GenerateAPIKeyID creates a unique API key ID
func GenerateAPIKeyID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "ak_" + hex.EncodeToString(b)
}

// GenerateAPIKey creates a new API key (returns plaintext, store hash)
func GenerateAPIKey() (plaintext string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// NewAPIKey creates a new APIKey entry
func NewAPIKey(name string, scopes []string, rateLimit int, expiresIn time.Duration) (*APIKey, string, error) {
	plaintext, err := GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}

	key := &APIKey{
		ID:        GenerateAPIKeyID(),
		Name:      name,
		KeyHash:   HashAPIKey(plaintext),
		Scopes:    scopes,
		RateLimit: rateLimit,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if expiresIn > 0 {
		key.ExpiresAt = time.Now().Add(expiresIn).UTC().Format(time.RFC3339)
	}

	return key, plaintext, nil
}

// HashAPIKey creates a SHA-256 hash of an API key
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// VerifyAPIKey checks if a plaintext key matches a stored hash
func VerifyAPIKey(plaintext, hash string) bool {
	return HashAPIKey(plaintext) == hash
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
