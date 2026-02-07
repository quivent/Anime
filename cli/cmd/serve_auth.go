package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/launch"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// Flags for serve auth setup
var (
	authMethod          string
	authClientID        string
	authClientSecret    string
	authEmailDomain     string
	authGitHubOrg       string
	authGitHubTeam      string
	authOIDCIssuer      string
	authWeb3            bool
	authChainID         int
	authAllowedAddrs    []string
	authIPAllow         []string
	authIPDeny          []string
	authRateLimit       int
	authRateBurst       int
	authAPIKey          bool
	authAPIKeyHeader    string
	authHSTS            bool
	authInteractive     bool
	authRotateCookies   bool
	authRotateAPIKeys   bool
	authRotateKeyID     string
	authRotateAll       bool
)

var serveAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication for served applications",
	Long: `Configure, inspect, and manage authentication layers for served apps.

SUBCOMMANDS:
  status <name>    Show current auth configuration
  setup <name>     Configure or reconfigure authentication
  rotate <name>    Rotate secrets and keys
  disable <name>   Remove authentication

AUTH METHODS:
  oauth2-google    Google OAuth2 with email domain filtering
  oauth2-github    GitHub OAuth2 with org/team requirements
  oauth2-oidc      Generic OpenID Connect provider
  web3-siwe        Sign-In With Ethereum (Web3 wallet)
  basic            HTTP Basic Authentication
  api-key          API key header authentication

LAYERED SECURITY:
  --ip-allow       IP whitelist (CIDR notation)
  --ip-deny        IP blacklist (CIDR notation)
  --rate-limit     Rate limiting (requests/sec)
  --hsts           HTTP Strict Transport Security

EXAMPLES:
  anime serve auth status myapp
  anime serve auth setup myapp --method oauth2-google --client-id XXX --client-secret YYY
  anime serve auth setup myapp --method oauth2-github --github-org myorg
  anime serve auth setup myapp --web3 --chain-id 1
  anime serve auth setup myapp --ip-allow 10.0.0.0/8 --ip-allow 192.168.0.0/16
  anime serve auth setup myapp --rate-limit 100
  anime serve auth rotate myapp --cookies
  anime serve auth rotate myapp --api-keys
  anime serve auth disable myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var serveAuthStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show authentication configuration for an app (or all apps)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServeAuthStatus,
}

var serveAuthTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available authentication methods",
	RunE:  runServeAuthTypes,
}

var serveAuthSetupCmd = &cobra.Command{
	Use:   "setup <name>",
	Short: "Configure authentication for an app",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeAuthSetup,
}

var serveAuthRotateCmd = &cobra.Command{
	Use:   "rotate <name>",
	Short: "Rotate authentication secrets",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeAuthRotate,
}

var serveAuthDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Remove authentication from an app",
	Args:  cobra.ExactArgs(1),
	RunE:  runServeAuthDisable,
}

func init() {
	// Setup flags
	serveAuthSetupCmd.Flags().StringVar(&authMethod, "method", "", "Auth method: oauth2-google, oauth2-github, oauth2-oidc, web3-siwe, basic, api-key")
	serveAuthSetupCmd.Flags().StringVar(&authClientID, "client-id", "", "OAuth2 client ID")
	serveAuthSetupCmd.Flags().StringVar(&authClientSecret, "client-secret", "", "OAuth2 client secret")
	serveAuthSetupCmd.Flags().StringVar(&authEmailDomain, "email-domain", "*", "Allowed email domain (Google OAuth)")
	serveAuthSetupCmd.Flags().StringVar(&authGitHubOrg, "github-org", "", "Required GitHub organization")
	serveAuthSetupCmd.Flags().StringVar(&authGitHubTeam, "github-team", "", "Required GitHub team")
	serveAuthSetupCmd.Flags().StringVar(&authOIDCIssuer, "oidc-issuer", "", "OIDC issuer URL")
	serveAuthSetupCmd.Flags().BoolVar(&authWeb3, "web3", false, "Enable Web3/SIWE authentication")
	serveAuthSetupCmd.Flags().IntVar(&authChainID, "chain-id", 1, "Ethereum chain ID (1=mainnet)")
	serveAuthSetupCmd.Flags().StringSliceVar(&authAllowedAddrs, "allowed-addresses", nil, "Allowed wallet addresses")
	serveAuthSetupCmd.Flags().StringSliceVar(&authIPAllow, "ip-allow", nil, "Allow CIDRs (whitelist)")
	serveAuthSetupCmd.Flags().StringSliceVar(&authIPDeny, "ip-deny", nil, "Deny CIDRs (blacklist)")
	serveAuthSetupCmd.Flags().IntVar(&authRateLimit, "rate-limit", 0, "Requests per second per IP")
	serveAuthSetupCmd.Flags().IntVar(&authRateBurst, "rate-burst", 50, "Burst size for rate limiting")
	serveAuthSetupCmd.Flags().BoolVar(&authAPIKey, "api-key", false, "Enable API key authentication")
	serveAuthSetupCmd.Flags().StringVar(&authAPIKeyHeader, "api-key-header", "X-API-Key", "API key header name")
	serveAuthSetupCmd.Flags().BoolVar(&authHSTS, "hsts", true, "Enable HSTS header")
	serveAuthSetupCmd.Flags().BoolVarP(&authInteractive, "interactive", "i", false, "Use interactive wizard")

	// Rotate flags
	serveAuthRotateCmd.Flags().BoolVar(&authRotateCookies, "cookies", false, "Rotate cookie secrets")
	serveAuthRotateCmd.Flags().BoolVar(&authRotateAPIKeys, "api-keys", false, "Rotate API keys")
	serveAuthRotateCmd.Flags().StringVar(&authRotateKeyID, "key-id", "", "Specific API key ID to rotate")
	serveAuthRotateCmd.Flags().BoolVar(&authRotateAll, "all", false, "Rotate all secrets")

	// Register subcommands
	serveAuthCmd.AddCommand(serveAuthStatusCmd)
	serveAuthCmd.AddCommand(serveAuthSetupCmd)
	serveAuthCmd.AddCommand(serveAuthRotateCmd)
	serveAuthCmd.AddCommand(serveAuthDisableCmd)
	serveAuthCmd.AddCommand(serveAuthTypesCmd)

	// Register auth under serve
	serveCmd.AddCommand(serveAuthCmd)
}

// ── Types ──────────────────────────────────────────────────────────────

func runServeAuthTypes(cmd *cobra.Command, args []string) error {
	theme.RenderBanner("Authentication Methods")

	fmt.Println(theme.InfoStyle.Render("\n  Primary Authentication:"))
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("oauth2-google") + "   Google OAuth2")
	fmt.Println("                    Email domain filtering (e.g., @company.com)")
	fmt.Println("                    Setup: console.cloud.google.com/apis/credentials")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("oauth2-github") + "   GitHub OAuth2")
	fmt.Println("                    Organization/team requirements")
	fmt.Println("                    Setup: github.com/settings/developers")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("oauth2-oidc") + "     Generic OpenID Connect")
	fmt.Println("                    Works with Okta, Auth0, Keycloak, etc.")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("web3-siwe") + "       Sign-In With Ethereum")
	fmt.Println("                    Wallet-based auth (MetaMask, WalletConnect)")
	fmt.Println("                    Optional address whitelist")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("basic") + "           HTTP Basic Authentication")
	fmt.Println("                    Simple username/password")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("api-key") + "         API Key Authentication")
	fmt.Println("                    Header-based key validation")
	fmt.Println("                    Auto-generated secure keys")

	fmt.Println(theme.InfoStyle.Render("\n  Security Layers (combinable):"))
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("--ip-allow") + "      IP Whitelist (CIDR notation)")
	fmt.Println("                    Example: --ip-allow 10.0.0.0/8")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("--ip-deny") + "       IP Blacklist (CIDR notation)")
	fmt.Println("                    Example: --ip-deny 1.2.3.4/32")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("--rate-limit") + "    Rate Limiting")
	fmt.Println("                    Requests per second per IP")
	fmt.Println("                    Example: --rate-limit 100 --rate-burst 200")
	fmt.Println()
	fmt.Println("    " + theme.HighlightStyle.Render("--hsts") + "          Strict Transport Security")
	fmt.Println("                    Forces HTTPS (enabled by default)")

	fmt.Println(theme.InfoStyle.Render("\n  Security Headers (auto-enabled):"))
	fmt.Println()
	fmt.Println("    • HSTS (Strict-Transport-Security)")
	fmt.Println("    • X-Frame-Options (clickjacking protection)")
	fmt.Println("    • X-Content-Type-Options (MIME sniffing prevention)")
	fmt.Println("    • Referrer-Policy")

	fmt.Println(theme.DimTextStyle.Render("\n  Combine methods: anime serve auth setup myapp --method oauth2-github --ip-allow 10.0.0.0/8 --rate-limit 100"))

	return nil
}

// ── Status ─────────────────────────────────────────────────────────────

func runServeAuthStatusAll(cfg *config.Config) error {
	theme.RenderBanner("Auth Status: All Apps")

	if len(cfg.LaunchedApps) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No served apps found"))
		return nil
	}

	for _, app := range cfg.LaunchedApps {
		auth := app.GetEffectiveAuth()

		// App name with status indicator
		var status string
		if auth.IsEmpty() {
			status = theme.WarningStyle.Render("NO AUTH")
		} else {
			status = theme.SuccessStyle.Render("PROTECTED")
		}

		fmt.Printf("\n  %s [%s]\n", theme.HighlightStyle.Render(app.Name), status)
		fmt.Printf("    Domain: %s\n", app.Domain)

		if !auth.IsEmpty() {
			// Show methods
			methods := strings.Join(auth.Methods, ", ")
			fmt.Printf("    Methods: %s\n", methods)

			// Show layers
			layers := []string{}
			if auth.IPAccess != nil {
				layers = append(layers, fmt.Sprintf("IP %s (%d CIDRs)", auth.IPAccess.Mode, len(auth.IPAccess.CIDRs)))
			}
			if auth.RateLimit != nil {
				layers = append(layers, fmt.Sprintf("Rate limit %d/s", auth.RateLimit.RequestsPerSec))
			}
			if auth.APIKey != nil {
				layers = append(layers, fmt.Sprintf("%d API keys", len(auth.APIKey.Keys)))
			}
			if auth.Security != nil && auth.Security.HSTS {
				layers = append(layers, "HSTS")
			}
			if len(layers) > 0 {
				fmt.Printf("    Layers: %s\n", strings.Join(layers, ", "))
			}
		}
	}

	fmt.Println()
	return nil
}

func runServeAuthStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no app specified, show all apps
	if len(args) == 0 {
		return runServeAuthStatusAll(cfg)
	}

	appName := args[0]
	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return fmt.Errorf("app not found: %s", appName)
	}

	theme.RenderBanner(fmt.Sprintf("Auth Status: %s", appName))

	auth := app.GetEffectiveAuth()
	if auth.IsEmpty() {
		fmt.Println(theme.WarningStyle.Render("  No authentication configured"))
		fmt.Println(theme.DimTextStyle.Render("  Run: anime serve auth setup " + appName))
		return nil
	}

	// Display auth methods
	fmt.Println(theme.InfoStyle.Render("  Methods:"))
	for _, method := range auth.Methods {
		fmt.Printf("    • %s\n", method)
	}

	// OAuth2 details
	if auth.OAuth2 != nil {
		fmt.Println(theme.InfoStyle.Render("\n  OAuth2:"))
		fmt.Printf("    Provider: %s\n", auth.OAuth2.Provider)
		if auth.OAuth2.EmailDomain != "" {
			fmt.Printf("    Email Domain: %s\n", auth.OAuth2.EmailDomain)
		}
		if auth.OAuth2.Org != "" {
			fmt.Printf("    GitHub Org: %s\n", auth.OAuth2.Org)
		}
		if auth.OAuth2.Team != "" {
			fmt.Printf("    GitHub Team: %s\n", auth.OAuth2.Team)
		}
	}

	// Web3 details
	if auth.Web3 != nil {
		fmt.Println(theme.InfoStyle.Render("\n  Web3/SIWE:"))
		fmt.Printf("    Chain ID: %d\n", auth.Web3.ChainID)
		fmt.Printf("    Service Port: %d\n", auth.Web3.ServicePort)
		if len(auth.Web3.AllowedAddresses) > 0 {
			fmt.Printf("    Allowed Addresses: %d configured\n", len(auth.Web3.AllowedAddresses))
		}
	}

	// IP access
	if auth.IPAccess != nil {
		fmt.Println(theme.InfoStyle.Render("\n  IP Access Control:"))
		fmt.Printf("    Mode: %s\n", auth.IPAccess.Mode)
		fmt.Printf("    CIDRs: %s\n", strings.Join(auth.IPAccess.CIDRs, ", "))
	}

	// Rate limiting
	if auth.RateLimit != nil {
		fmt.Println(theme.InfoStyle.Render("\n  Rate Limiting:"))
		fmt.Printf("    Requests/sec: %d\n", auth.RateLimit.RequestsPerSec)
		fmt.Printf("    Burst: %d\n", auth.RateLimit.BurstSize)
	}

	// API keys
	if auth.APIKey != nil {
		fmt.Println(theme.InfoStyle.Render("\n  API Keys:"))
		fmt.Printf("    Header: %s\n", auth.APIKey.HeaderName)
		fmt.Printf("    Keys: %d configured\n", len(auth.APIKey.Keys))
		for _, key := range auth.APIKey.Keys {
			expires := "never"
			if key.ExpiresAt != "" {
				expires = key.ExpiresAt
			}
			fmt.Printf("      • %s (%s) - expires: %s\n", key.Name, key.ID, expires)
		}
	}

	// Security headers
	if auth.Security != nil {
		fmt.Println(theme.InfoStyle.Render("\n  Security Headers:"))
		if auth.Security.HSTS {
			fmt.Printf("    HSTS: enabled (max-age: %d)\n", auth.Security.HSTSMaxAge)
		}
		if auth.Security.CSP != "" {
			fmt.Printf("    CSP: %s\n", auth.Security.CSP)
		}
		if auth.Security.XFrameOptions != "" {
			fmt.Printf("    X-Frame-Options: %s\n", auth.Security.XFrameOptions)
		}
	}

	return nil
}

// ── Setup ──────────────────────────────────────────────────────────────

func runServeAuthSetup(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return fmt.Errorf("app not found: %s", appName)
	}

	theme.RenderBanner(fmt.Sprintf("Auth Setup: %s", appName))

	// Initialize auth config if nil
	if app.Auth == nil {
		app.Auth = &config.AppAuthConfig{
			Methods: []string{},
		}
	}

	// Interactive mode
	if authInteractive || (authMethod == "" && !authWeb3 && len(authIPAllow) == 0 && len(authIPDeny) == 0 && authRateLimit == 0 && !authAPIKey) {
		return runServeAuthSetupInteractive(app, cfg)
	}

	// Flag-based setup
	return runServeAuthSetupFlags(app, cfg)
}

func runServeAuthSetupInteractive(app *config.LaunchedApp, cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.InfoStyle.Render("  Select authentication method(s):"))
	fmt.Println("  1. Google OAuth")
	fmt.Println("  2. GitHub OAuth")
	fmt.Println("  3. Generic OIDC")
	fmt.Println("  4. Web3/SIWE (Ethereum wallet)")
	fmt.Println("  5. HTTP Basic Auth")
	fmt.Println("  6. API Key")
	fmt.Println("  7. None (public access)")
	fmt.Print(theme.HighlightStyle.Render("  Select (1-7) ▶ "))

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return setupOAuth2Google(reader, app, cfg)
	case "2":
		return setupOAuth2GitHub(reader, app, cfg)
	case "3":
		return setupOAuth2OIDC(reader, app, cfg)
	case "4":
		return setupWeb3SIWE(reader, app, cfg)
	case "5":
		return setupBasicAuth(reader, app, cfg)
	case "6":
		return setupAPIKey(reader, app, cfg)
	case "7":
		app.Auth.Methods = []string{"none"}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(theme.SuccessStyle.Render("  Authentication disabled"))
		return nil
	default:
		return fmt.Errorf("invalid selection: %s", choice)
	}
}

func runServeAuthSetupFlags(app *config.LaunchedApp, cfg *config.Config) error {
	modified := false

	// OAuth2 setup
	if authMethod != "" {
		switch authMethod {
		case "oauth2-google":
			if authClientID == "" || authClientSecret == "" {
				return fmt.Errorf("--client-id and --client-secret required for oauth2-google")
			}
			cookieSecret, _ := launch.GenerateCookieSecret()
			app.Auth.OAuth2 = &config.AppOAuth2Config{
				Provider:     "google",
				ClientID:     authClientID,
				ClientSecret: authClientSecret,
				CookieSecret: cookieSecret,
				EmailDomain:  authEmailDomain,
			}
			app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-google")
			modified = true

		case "oauth2-github":
			if authClientID == "" || authClientSecret == "" {
				return fmt.Errorf("--client-id and --client-secret required for oauth2-github")
			}
			cookieSecret, _ := launch.GenerateCookieSecret()
			app.Auth.OAuth2 = &config.AppOAuth2Config{
				Provider:     "github",
				ClientID:     authClientID,
				ClientSecret: authClientSecret,
				CookieSecret: cookieSecret,
				Org:          authGitHubOrg,
				Team:         authGitHubTeam,
			}
			app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-github")
			modified = true

		case "oauth2-oidc":
			if authClientID == "" || authClientSecret == "" || authOIDCIssuer == "" {
				return fmt.Errorf("--client-id, --client-secret, and --oidc-issuer required for oauth2-oidc")
			}
			cookieSecret, _ := launch.GenerateCookieSecret()
			app.Auth.OAuth2 = &config.AppOAuth2Config{
				Provider:     "oidc",
				ClientID:     authClientID,
				ClientSecret: authClientSecret,
				CookieSecret: cookieSecret,
				IssuerURL:    authOIDCIssuer,
			}
			app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-oidc")
			modified = true

		case "basic":
			return fmt.Errorf("basic auth requires interactive mode (-i) for password input")

		case "api-key":
			authAPIKey = true
		}
	}

	// Web3/SIWE
	if authWeb3 {
		app.Auth.Web3 = &config.AppWeb3Config{
			ChainID:          authChainID,
			AllowedAddresses: authAllowedAddrs,
			ServicePort:      4181,
			NonceExpiry:      "5m",
			SessionTTL:       "24h",
		}
		app.Auth.Methods = appendMethod(app.Auth.Methods, "web3-siwe")
		modified = true
	}

	// IP access control
	if len(authIPAllow) > 0 {
		app.Auth.IPAccess = &config.AppIPAccessConfig{
			Mode:  "allow",
			CIDRs: authIPAllow,
		}
		modified = true
	} else if len(authIPDeny) > 0 {
		app.Auth.IPAccess = &config.AppIPAccessConfig{
			Mode:  "deny",
			CIDRs: authIPDeny,
		}
		modified = true
	}

	// Rate limiting
	if authRateLimit > 0 {
		app.Auth.RateLimit = &config.AppRateLimitConfig{
			RequestsPerSec: authRateLimit,
			BurstSize:      authRateBurst,
			ZoneSize:       "10m",
		}
		modified = true
	}

	// API key
	if authAPIKey {
		if app.Auth.APIKey == nil {
			app.Auth.APIKey = &config.AppAPIKeyConfig{
				HeaderName: authAPIKeyHeader,
				Keys:       []config.AppAPIKey{},
			}
		}
		// Generate initial key
		key, plaintext, err := launch.NewAPIKey("default", nil, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}
		app.Auth.APIKey.Keys = append(app.Auth.APIKey.Keys, config.AppAPIKey{
			ID:        key.ID,
			Name:      key.Name,
			KeyHash:   key.KeyHash,
			CreatedAt: key.CreatedAt,
		})
		app.Auth.Methods = appendMethod(app.Auth.Methods, "api-key")
		modified = true

		fmt.Println(theme.SuccessStyle.Render("  API Key generated:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(plaintext))
		fmt.Println(theme.WarningStyle.Render("  Save this key - it cannot be retrieved later!"))
	}

	// Security headers
	if authHSTS {
		if app.Auth.Security == nil {
			app.Auth.Security = &config.AppSecurityConfig{}
		}
		app.Auth.Security.HSTS = true
		app.Auth.Security.HSTSMaxAge = 31536000
		app.Auth.Security.XContentType = true
		app.Auth.Security.XFrameOptions = "SAMEORIGIN"
		app.Auth.Security.ReferrerPolicy = "strict-origin-when-cross-origin"
		modified = true
	}

	if !modified {
		return fmt.Errorf("no auth options specified, use -i for interactive mode")
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("  Auth configuration updated"))
	fmt.Println(theme.DimTextStyle.Render("  Note: Restart the app or reload nginx to apply changes"))

	return nil
}

// ── Interactive Setup Helpers ──────────────────────────────────────────

func setupOAuth2Google(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  Google Client ID ▶ "))
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print(theme.HighlightStyle.Render("  Google Client Secret ▶ "))
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	fmt.Print(theme.HighlightStyle.Render("  Allowed email domain (* for any) ▶ "))
	emailDomain, _ := reader.ReadString('\n')
	emailDomain = strings.TrimSpace(emailDomain)
	if emailDomain == "" {
		emailDomain = "*"
	}

	cookieSecret, _ := launch.GenerateCookieSecret()
	app.Auth.OAuth2 = &config.AppOAuth2Config{
		Provider:     "google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CookieSecret: cookieSecret,
		EmailDomain:  emailDomain,
	}
	app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-google")

	return saveAndApplyAuth(app, cfg)
}

func setupOAuth2GitHub(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  GitHub Client ID ▶ "))
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print(theme.HighlightStyle.Render("  GitHub Client Secret ▶ "))
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	fmt.Print(theme.HighlightStyle.Render("  Required GitHub Org (optional) ▶ "))
	org, _ := reader.ReadString('\n')
	org = strings.TrimSpace(org)

	fmt.Print(theme.HighlightStyle.Render("  Required GitHub Team (optional) ▶ "))
	team, _ := reader.ReadString('\n')
	team = strings.TrimSpace(team)

	cookieSecret, _ := launch.GenerateCookieSecret()
	app.Auth.OAuth2 = &config.AppOAuth2Config{
		Provider:     "github",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CookieSecret: cookieSecret,
		Org:          org,
		Team:         team,
	}
	app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-github")

	return saveAndApplyAuth(app, cfg)
}

func setupOAuth2OIDC(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  OIDC Issuer URL ▶ "))
	issuer, _ := reader.ReadString('\n')
	issuer = strings.TrimSpace(issuer)

	fmt.Print(theme.HighlightStyle.Render("  Client ID ▶ "))
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print(theme.HighlightStyle.Render("  Client Secret ▶ "))
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	cookieSecret, _ := launch.GenerateCookieSecret()
	app.Auth.OAuth2 = &config.AppOAuth2Config{
		Provider:     "oidc",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CookieSecret: cookieSecret,
		IssuerURL:    issuer,
		Scopes:       []string{"openid", "email", "profile"},
	}
	app.Auth.Methods = appendMethod(app.Auth.Methods, "oauth2-oidc")

	return saveAndApplyAuth(app, cfg)
}

func setupWeb3SIWE(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  Chain ID (1=mainnet, 5=goerli) ▶ "))
	chainIDStr, _ := reader.ReadString('\n')
	chainIDStr = strings.TrimSpace(chainIDStr)
	chainID := 1
	if chainIDStr != "" {
		fmt.Sscanf(chainIDStr, "%d", &chainID)
	}

	fmt.Print(theme.HighlightStyle.Render("  Allowed wallet addresses (comma-separated, empty for any) ▶ "))
	addrsStr, _ := reader.ReadString('\n')
	addrsStr = strings.TrimSpace(addrsStr)
	var allowedAddrs []string
	if addrsStr != "" {
		for _, addr := range strings.Split(addrsStr, ",") {
			allowedAddrs = append(allowedAddrs, strings.TrimSpace(addr))
		}
	}

	app.Auth.Web3 = &config.AppWeb3Config{
		ChainID:          chainID,
		AllowedAddresses: allowedAddrs,
		ServicePort:      4181,
		NonceExpiry:      "5m",
		SessionTTL:       "24h",
	}
	app.Auth.Methods = appendMethod(app.Auth.Methods, "web3-siwe")

	return saveAndApplyAuth(app, cfg)
}

func setupBasicAuth(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  Username ▶ "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print(theme.HighlightStyle.Render("  Password ▶ "))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	app.Auth.Basic = &config.AppBasicAuthConfig{
		Username:     username,
		PasswordHash: launch.HashAPIKey(password), // Using same hash function
		Realm:        app.Name,
	}
	app.Auth.Methods = appendMethod(app.Auth.Methods, "basic")

	return saveAndApplyAuth(app, cfg)
}

func setupAPIKey(reader *bufio.Reader, app *config.LaunchedApp, cfg *config.Config) error {
	fmt.Print(theme.HighlightStyle.Render("  API key header name [X-API-Key] ▶ "))
	headerName, _ := reader.ReadString('\n')
	headerName = strings.TrimSpace(headerName)
	if headerName == "" {
		headerName = "X-API-Key"
	}

	fmt.Print(theme.HighlightStyle.Render("  Key name/description ▶ "))
	keyName, _ := reader.ReadString('\n')
	keyName = strings.TrimSpace(keyName)
	if keyName == "" {
		keyName = "default"
	}

	key, plaintext, err := launch.NewAPIKey(keyName, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to generate API key: %w", err)
	}

	if app.Auth.APIKey == nil {
		app.Auth.APIKey = &config.AppAPIKeyConfig{
			HeaderName: headerName,
			Keys:       []config.AppAPIKey{},
		}
	}

	app.Auth.APIKey.Keys = append(app.Auth.APIKey.Keys, config.AppAPIKey{
		ID:        key.ID,
		Name:      key.Name,
		KeyHash:   key.KeyHash,
		CreatedAt: key.CreatedAt,
	})
	app.Auth.Methods = appendMethod(app.Auth.Methods, "api-key")

	fmt.Println(theme.SuccessStyle.Render("\n  API Key generated:"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render(plaintext))
	fmt.Println(theme.WarningStyle.Render("  Save this key - it cannot be retrieved later!"))

	return saveAndApplyAuth(app, cfg)
}

func saveAndApplyAuth(app *config.LaunchedApp, cfg *config.Config) error {
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("\n  Auth configuration saved"))
	fmt.Println(theme.DimTextStyle.Render("  Run 'anime serve auth apply " + app.Name + "' to deploy changes"))
	return nil
}

// ── Rotate ─────────────────────────────────────────────────────────────

func runServeAuthRotate(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return fmt.Errorf("app not found: %s", appName)
	}

	if app.Auth == nil || app.Auth.IsEmpty() {
		return fmt.Errorf("no authentication configured for %s", appName)
	}

	theme.RenderBanner(fmt.Sprintf("Rotate Secrets: %s", appName))

	rotated := false

	// Rotate cookie secret
	if authRotateCookies || authRotateAll {
		if app.Auth.OAuth2 != nil {
			newSecret, err := launch.GenerateCookieSecret()
			if err != nil {
				return fmt.Errorf("failed to generate cookie secret: %w", err)
			}
			app.Auth.OAuth2.CookieSecret = newSecret
			fmt.Println(theme.SuccessStyle.Render("  Cookie secret rotated"))
			rotated = true
		}
	}

	// Rotate API keys
	if authRotateAPIKeys || authRotateAll {
		if app.Auth.APIKey != nil && len(app.Auth.APIKey.Keys) > 0 {
			for i := range app.Auth.APIKey.Keys {
				if authRotateKeyID != "" && app.Auth.APIKey.Keys[i].ID != authRotateKeyID {
					continue
				}
				key, plaintext, err := launch.NewAPIKey(app.Auth.APIKey.Keys[i].Name, nil, 0, 0)
				if err != nil {
					return fmt.Errorf("failed to generate API key: %w", err)
				}
				app.Auth.APIKey.Keys[i].KeyHash = key.KeyHash
				app.Auth.APIKey.Keys[i].CreatedAt = key.CreatedAt

				fmt.Printf(theme.SuccessStyle.Render("  API Key '%s' rotated:\n"), app.Auth.APIKey.Keys[i].Name)
				fmt.Printf("    %s\n", theme.HighlightStyle.Render(plaintext))
				rotated = true
			}
		}
	}

	if !rotated {
		return fmt.Errorf("nothing to rotate, use --cookies, --api-keys, or --all")
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(theme.WarningStyle.Render("\n  Restart services to apply new secrets"))
	return nil
}

// ── Disable ────────────────────────────────────────────────────────────

func runServeAuthDisable(cmd *cobra.Command, args []string) error {
	appName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app, err := cfg.GetLaunchedApp(appName)
	if err != nil {
		return fmt.Errorf("app not found: %s", appName)
	}

	theme.RenderBanner(fmt.Sprintf("Disable Auth: %s", appName))

	// Clear auth config
	app.Auth = &config.AppAuthConfig{
		Methods: []string{"none"},
	}
	app.AuthType = "" // Clear legacy field too

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("  Authentication disabled"))
	fmt.Println(theme.DimTextStyle.Render("  Reload nginx to apply changes"))

	return nil
}

// ── Helpers ────────────────────────────────────────────────────────────

func appendMethod(methods []string, method string) []string {
	// Remove "none" if adding a real method
	if method != "none" {
		var filtered []string
		for _, m := range methods {
			if m != "none" {
				filtered = append(filtered, m)
			}
		}
		methods = filtered
	}
	// Check if already exists
	for _, m := range methods {
		if m == method {
			return methods
		}
	}
	return append(methods, method)
}
