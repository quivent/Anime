package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS management utilities",
	Long:  "Manage DNS records for your domains",
}

var vercelCmd = &cobra.Command{
	Use:   "vercel [domain] [ip]",
	Short: "Manage Vercel DNS records",
	Long: `Manage DNS records for domains using Vercel DNS.

Usage:
  anime dns vercel domains        - List all domains
  anime dns vercel DOMAIN IP      - Set A records for domain to IP`,
	Run: runVercel,
}

var namecheapCmd = &cobra.Command{
	Use:   "namecheap [domain] [ip]",
	Short: "Manage Namecheap DNS records",
	Long: `Manage DNS records for domains using the Namecheap API.

Usage:
  anime dns namecheap domains        - List all domains
  anime dns namecheap DOMAIN IP      - Set A records for domain to IP

Configuration:
  Add namecheap credentials to ~/.dns-config.json:
  {
    "namecheap": {
      "apiUser": "your-username",
      "apiKey": "your-api-key",
      "username": "your-username"
    }
  }

Prerequisites:
  - Enable API access in your Namecheap account
  - Whitelist your IP address in the Namecheap dashboard`,
	Run: runNamecheap,
}

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare [domain] [ip]",
	Short: "Manage Cloudflare DNS records",
	Long: `Manage DNS records for domains using the Cloudflare API.

Usage:
  anime dns cloudflare domains        - List all zones
  anime dns cloudflare DOMAIN IP      - Set A records for domain to IP

Configuration:
  Add cloudflare credentials to ~/.dns-config.json:
  {
    "cloudflare": {
      "apiToken": "your-api-token"
    }
  }

Prerequisites:
  - Create an API token at https://dash.cloudflare.com/profile/api-tokens
  - Token needs Zone:Read and DNS:Edit permissions`,
	Run: runCloudflare,
}

var dnsListCmd = &cobra.Command{
	Use:   "list [ip]",
	Short: "List all domains pointing to an IP across all providers",
	Long: `List domains pointing to a specific IP address across all configured DNS providers.

Usage:
  anime dns list           - List domains pointing to your current public IP
  anime dns list IP        - List domains pointing to the specified IP`,
	Run: runDNSList,
}

func init() {
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.AddCommand(vercelCmd)
	dnsCmd.AddCommand(namecheapCmd)
	dnsCmd.AddCommand(cloudflareCmd)
	dnsCmd.AddCommand(dnsListCmd)
}

func runVercel(cmd *cobra.Command, args []string) {
	// Check if vercel CLI is installed
	if _, err := exec.LookPath("vercel"); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ Vercel CLI not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📦 Installation Options:"))
		fmt.Println()

		// Check which package managers are available
		installMethods := []struct {
			cmd     string
			check   string
			install string
			desc    string
		}{
			{"npm", "npm", "npm i -g vercel", "Node Package Manager"},
			{"pnpm", "pnpm", "pnpm add -g vercel", "Fast Node Package Manager"},
			{"yarn", "yarn", "yarn global add vercel", "Yarn Package Manager"},
			{"brew", "brew", "brew install vercel-cli", "Homebrew (macOS/Linux)"},
		}

		available := []string{}
		for _, method := range installMethods {
			if _, err := exec.LookPath(method.check); err == nil {
				fmt.Printf("  %s %s\n",
					theme.SuccessStyle.Render("✓"),
					theme.HighlightStyle.Render(method.install))
				fmt.Printf("    %s\n",
					theme.DimTextStyle.Render(method.desc))
				available = append(available, method.install)
			}
		}

		if len(available) == 0 {
			fmt.Println(theme.WarningStyle.Render("  ⚠ No package managers found"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  Install Node.js first:"))
			fmt.Println(theme.DimTextStyle.Render("    • macOS/Linux: https://nodejs.org"))
			fmt.Println(theme.DimTextStyle.Render("    • Or use Homebrew: brew install node"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("  Or download Vercel CLI directly:"))
			fmt.Println(theme.DimTextStyle.Render("    • https://vercel.com/download"))
		}

		fmt.Println()
		return
	}

	// If no args or "domains" subcommand, list domains
	if len(args) == 0 || (len(args) == 1 && args[0] == "domains") {
		listVercelDomains()
		return
	}

	// Set A records for domain
	if len(args) != 2 {
		fmt.Println(theme.ErrorStyle.Render("✗ Invalid arguments"))
		fmt.Println(theme.DimTextStyle.Render("  Usage: anime dns vercel DOMAIN IP"))
		return
	}

	domain := args[0]
	ip := args[1]
	setVercelARecords(domain, ip)
}

func listVercelDomains() {
	fmt.Println(theme.RenderBanner("⚡ VERCEL DOMAINS ⚡"))
	fmt.Println()

	// Run vercel dns ls to get all DNS records
	output, err := exec.Command("vercel", "dns", "ls").CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ Failed to list domains"))
		fmt.Println(theme.DimTextStyle.Render("  " + string(output)))
		return
	}

	// Parse and display domains
	lines := strings.Split(string(output), "\n")
	domains := make(map[string]bool)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Extract domain from the line
			domain := fields[1]
			if domain != "Domain" && domain != "" {
				domains[domain] = true
			}
		}
	}

	if len(domains) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠ No domains found"))
		fmt.Println(theme.DimTextStyle.Render("  Add a domain at: https://vercel.com/dashboard/domains"))
		fmt.Println()
		return
	}

	fmt.Println(theme.GlowStyle.Render("🌸 Your Domains:"))
	fmt.Println()

	i := 1
	for domain := range domains {
		fmt.Printf("  %s %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d.", i)),
			theme.SuccessStyle.Render(domain))
		i++
	}
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("✨ Quick Setup:"))
	fmt.Printf("  %s\n",
		theme.DimTextStyle.Render("anime dns vercel DOMAIN IP - Set A records"))
	fmt.Println()
}

// --- Namecheap DNS Provider ---

type namecheapConfig struct {
	APIUser  string `json:"apiUser"`
	APIKey   string `json:"apiKey"`
	Username string `json:"username"`
}

type cloudflareConfig struct {
	APIToken string `json:"apiToken"`
}

type dnsConfig struct {
	Token      string           `json:"token"`
	TeamID     string           `json:"teamId"`
	Namecheap  namecheapConfig  `json:"namecheap"`
	Cloudflare cloudflareConfig `json:"cloudflare"`
}

// Namecheap XML response types
type namecheapAPIResponse struct {
	XMLName xml.Name `xml:"ApiResponse"`
	Status  string   `xml:"Status,attr"`
	Errors  struct {
		Error []struct {
			Number  string `xml:"Number,attr"`
			Message string `xml:",chardata"`
		} `xml:"Error"`
	} `xml:"Errors"`
	CommandResponse struct {
		DomainGetList struct {
			Domains []struct {
				Name string `xml:"Name,attr"`
			} `xml:"Domain"`
		} `xml:"DomainGetListResult"`
		DomainDNSGetHosts struct {
			Hosts []struct {
				HostID     string `xml:"HostId,attr"`
				Name       string `xml:"Name,attr"`
				Type       string `xml:"Type,attr"`
				Address    string `xml:"Address,attr"`
				MXPref     string `xml:"MXPref,attr"`
				TTL        string `xml:"TTL,attr"`
				IsActive   string `xml:"IsActive,attr"`
				IsDDNSEnabled string `xml:"IsDDNSEnabled,attr"`
			} `xml:"host"`
		} `xml:"DomainDNSGetHostsResult"`
		DomainDNSSetHosts struct {
			IsSuccess string `xml:"IsSuccess,attr"`
		} `xml:"DomainDNSSetHostsResult"`
	} `xml:"CommandResponse"`
}

func loadDNSConfig() (*dnsConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".dns-config.json"))
	if err != nil {
		return nil, fmt.Errorf("cannot read ~/.dns-config.json: %w", err)
	}
	var cfg dnsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid ~/.dns-config.json: %w", err)
	}
	return &cfg, nil
}

func getPublicIPForDNS() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("failed to detect public IP: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IP response: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

func splitDomain(domain string) (sld, tld string) {
	// Handle multi-part TLDs like .co.uk, .com.au
	multiPartTLDs := map[string]bool{
		"co.uk": true, "org.uk": true, "me.uk": true, "net.uk": true,
		"com.au": true, "net.au": true, "org.au": true,
		"co.nz": true, "net.nz": true, "org.nz": true,
		"co.za": true, "co.in": true, "co.jp": true,
		"com.br": true, "com.mx": true, "com.cn": true,
	}
	parts := strings.Split(domain, ".")
	if len(parts) >= 3 {
		candidate := strings.Join(parts[len(parts)-2:], ".")
		if multiPartTLDs[candidate] {
			return strings.Join(parts[:len(parts)-2], "."), candidate
		}
	}
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	}
	return domain, ""
}

func namecheapAPICall(cfg *namecheapConfig, clientIP, command string, extraParams url.Values) (*namecheapAPIResponse, error) {
	params := url.Values{}
	params.Set("ApiUser", cfg.APIUser)
	params.Set("ApiKey", cfg.APIKey)
	params.Set("UserName", cfg.Username)
	params.Set("ClientIp", clientIP)
	params.Set("Command", command)

	for k, v := range extraParams {
		for _, val := range v {
			params.Add(k, val)
		}
	}

	apiURL := "https://api.namecheap.com/xml.response?" + params.Encode()
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResp namecheapAPIResponse
	if err := xml.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if apiResp.Status == "ERROR" {
		msgs := []string{}
		for _, e := range apiResp.Errors.Error {
			msgs = append(msgs, e.Message)
		}
		return &apiResp, fmt.Errorf("API error: %s", strings.Join(msgs, "; "))
	}

	return &apiResp, nil
}

func runNamecheap(cmd *cobra.Command, args []string) {
	cfg, err := loadDNSConfig()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	if cfg.Namecheap.APIKey == "" || cfg.Namecheap.APIUser == "" {
		fmt.Println(theme.ErrorStyle.Render("✗ Namecheap credentials not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Add to ~/.dns-config.json:"))
		fmt.Println(theme.DimTextStyle.Render(`  {
    "namecheap": {
      "apiUser": "your-username",
      "apiKey": "your-api-key",
      "username": "your-username"
    }
  }`))
		fmt.Println()
		return
	}

	if len(args) == 0 || (len(args) == 1 && args[0] == "domains") {
		listNamecheapDomains(&cfg.Namecheap)
		return
	}

	if len(args) != 2 {
		fmt.Println(theme.ErrorStyle.Render("✗ Invalid arguments"))
		fmt.Println(theme.DimTextStyle.Render("  Usage: anime dns namecheap DOMAIN IP"))
		return
	}

	domain := args[0]
	ip := args[1]
	setNamecheapARecords(&cfg.Namecheap, domain, ip)
}

func listNamecheapDomains(cfg *namecheapConfig) {
	fmt.Println(theme.RenderBanner("🌐 NAMECHEAP DOMAINS 🌐"))
	fmt.Println()

	clientIP, err := getPublicIPForDNS()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	params := url.Values{}
	params.Set("PageSize", "100")
	apiResp, err := namecheapAPICall(cfg, clientIP, "namecheap.domains.getList", params)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	domains := apiResp.CommandResponse.DomainGetList.Domains
	if len(domains) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠ No domains found"))
		fmt.Println(theme.DimTextStyle.Render("  Manage domains at: https://ap.www.namecheap.com/Domains/DomainList"))
		fmt.Println()
		return
	}

	fmt.Println(theme.GlowStyle.Render("🌸 Your Domains:"))
	fmt.Println()

	for i, d := range domains {
		fmt.Printf("  %s %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d.", i+1)),
			theme.SuccessStyle.Render(d.Name))
	}
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("✨ Quick Setup:"))
	fmt.Printf("  %s\n",
		theme.DimTextStyle.Render("anime dns namecheap DOMAIN IP - Set A records"))
	fmt.Println()
}

func setNamecheapARecords(cfg *namecheapConfig, domain, ip string) {
	fmt.Println(theme.RenderBanner("🌐 DNS CONFIGURATION 🌐"))
	fmt.Println()

	fmt.Printf("%s Configuring DNS for %s → %s\n",
		theme.SymbolSparkle,
		theme.HighlightStyle.Render(domain),
		theme.InfoStyle.Render(ip))
	fmt.Println()

	clientIP, err := getPublicIPForDNS()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	sld, tld := splitDomain(domain)
	if tld == "" {
		fmt.Println(theme.ErrorStyle.Render("✗ Invalid domain format"))
		return
	}

	// Step 0: Ensure domain is using Namecheap BasicDNS
	fmt.Printf("  %s Activating Namecheap DNS...\n", theme.SymbolBolt)
	setDefaultParams := url.Values{}
	setDefaultParams.Set("SLD", sld)
	setDefaultParams.Set("TLD", tld)
	_, err = namecheapAPICall(cfg, clientIP, "namecheap.domains.dns.setDefault", setDefaultParams)
	if err != nil {
		// Log but continue - might already be set or might still work
		fmt.Printf("    %s %s (continuing anyway)\n", theme.WarningStyle.Render("⚠"), theme.DimTextStyle.Render(err.Error()))
	} else {
		fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("✓"), theme.DimTextStyle.Render("Namecheap BasicDNS active"))
	}

	// Step 1: Get existing host records
	fmt.Printf("  %s Fetching existing records...\n", theme.SymbolBolt)
	getParams := url.Values{}
	getParams.Set("SLD", sld)
	getParams.Set("TLD", tld)

	apiResp, err := namecheapAPICall(cfg, clientIP, "namecheap.domains.dns.getHosts", getParams)
	if err != nil {
		fmt.Printf("    %s %s\n", theme.ErrorStyle.Render("✗"), theme.ErrorStyle.Render(err.Error()))
		return
	}

	// Step 2: Build new host records list, preserving existing non-A records
	// and existing A records that aren't @ or *
	setParams := url.Values{}
	setParams.Set("SLD", sld)
	setParams.Set("TLD", tld)

	idx := 1
	existingHosts := apiResp.CommandResponse.DomainDNSGetHosts.Hosts
	for _, h := range existingHosts {
		// Skip A records for @ and * — we'll replace those
		if h.Type == "A" && (h.Name == "@" || h.Name == "*") {
			continue
		}
		n := strconv.Itoa(idx)
		setParams.Set("HostName"+n, h.Name)
		setParams.Set("RecordType"+n, h.Type)
		setParams.Set("Address"+n, h.Address)
		setParams.Set("TTL"+n, h.TTL)
		if h.MXPref != "" && h.MXPref != "0" {
			setParams.Set("MXPref"+n, h.MXPref)
		}
		idx++
	}

	// Step 3: Add our A records for @ and *
	records := []struct {
		name string
		desc string
	}{
		{"@", "Root domain"},
		{"*", "Wildcard subdomain"},
	}

	for _, rec := range records {
		n := strconv.Itoa(idx)
		setParams.Set("HostName"+n, rec.name)
		setParams.Set("RecordType"+n, "A")
		setParams.Set("Address"+n, ip)
		setParams.Set("TTL"+n, "60")
		idx++

		fmt.Printf("  %s Setting %s record (%s)...\n",
			theme.SymbolBolt,
			theme.HighlightStyle.Render(rec.name),
			theme.DimTextStyle.Render(rec.desc))
	}

	// Step 4: Apply all records
	fmt.Printf("  %s Applying DNS records...\n", theme.SymbolBolt)
	_, err = namecheapAPICall(cfg, clientIP, "namecheap.domains.dns.setHosts", setParams)
	if err != nil {
		fmt.Printf("    %s %s\n", theme.ErrorStyle.Render("✗"), theme.ErrorStyle.Render(err.Error()))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Check that your IP is whitelisted in the Namecheap dashboard"))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ DNS configuration complete!"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📊 Summary:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Records set: %d (preserved %d existing)", 2, idx-3)))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("TTL: 60 seconds"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("Propagation: ~1-5 minutes"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🌐 Your domain is now pointing to:"))
	fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(domain), theme.InfoStyle.Render(ip))
	fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render("*."+domain), theme.InfoStyle.Render(ip))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("✨ Verify DNS:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("dig "+domain))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("dig *."+domain))
	fmt.Println()
}

// --- Cloudflare DNS Provider ---

type cfAPIResponse struct {
	Success bool            `json:"success"`
	Errors  []cfAPIError    `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type cfAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cfDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func cloudflareAPI(token, method, path string, body io.Reader) (*cfAPIResponse, error) {
	req, err := http.NewRequest(method, "https://api.cloudflare.com/client/v4"+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResp cfAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResp.Success {
		msgs := []string{}
		for _, e := range apiResp.Errors {
			msgs = append(msgs, e.Message)
		}
		return &apiResp, fmt.Errorf("API error: %s", strings.Join(msgs, "; "))
	}

	return &apiResp, nil
}

func runCloudflare(cmd *cobra.Command, args []string) {
	cfg, err := loadDNSConfig()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	if cfg.Cloudflare.APIToken == "" {
		fmt.Println(theme.ErrorStyle.Render("✗ Cloudflare credentials not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Add to ~/.dns-config.json:"))
		fmt.Println(theme.DimTextStyle.Render(`  {
    "cloudflare": {
      "apiToken": "your-api-token"
    }
  }`))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Create a token at:"))
		fmt.Println(theme.DimTextStyle.Render("  https://dash.cloudflare.com/profile/api-tokens"))
		fmt.Println(theme.DimTextStyle.Render("  Permissions: Zone:Read, DNS:Edit"))
		fmt.Println()
		return
	}

	if len(args) == 0 || (len(args) == 1 && args[0] == "domains") {
		listCloudflareDomains(cfg.Cloudflare.APIToken)
		return
	}

	if len(args) != 2 {
		fmt.Println(theme.ErrorStyle.Render("✗ Invalid arguments"))
		fmt.Println(theme.DimTextStyle.Render("  Usage: anime dns cloudflare DOMAIN IP"))
		return
	}

	domain := args[0]
	ip := args[1]
	setCloudflareARecords(cfg.Cloudflare.APIToken, domain, ip)
}

func listCloudflareDomains(token string) {
	fmt.Println(theme.RenderBanner("☁ CLOUDFLARE DOMAINS ☁"))
	fmt.Println()

	apiResp, err := cloudflareAPI(token, "GET", "/zones?per_page=50&status=active", nil)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
		return
	}

	var zones []cfZone
	if err := json.Unmarshal(apiResp.Result, &zones); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗ Failed to parse zones"))
		return
	}

	if len(zones) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠ No domains found"))
		fmt.Println(theme.DimTextStyle.Render("  Add a domain at: https://dash.cloudflare.com"))
		fmt.Println()
		return
	}

	fmt.Println(theme.GlowStyle.Render("🌸 Your Domains:"))
	fmt.Println()

	for i, z := range zones {
		fmt.Printf("  %s %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%d.", i+1)),
			theme.SuccessStyle.Render(z.Name))
	}
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("✨ Quick Setup:"))
	fmt.Printf("  %s\n",
		theme.DimTextStyle.Render("anime dns cloudflare DOMAIN IP - Set A records"))
	fmt.Println()
}

func cloudflareGetZoneID(token, domain string) (string, error) {
	apiResp, err := cloudflareAPI(token, "GET", "/zones?name="+url.QueryEscape(domain), nil)
	if err != nil {
		return "", err
	}

	var zones []cfZone
	if err := json.Unmarshal(apiResp.Result, &zones); err != nil {
		return "", fmt.Errorf("failed to parse zones: %w", err)
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("zone not found for %s", domain)
	}

	return zones[0].ID, nil
}

func cloudflareUpsertARecord(token, zoneID, name, ip string) error {
	// Check if record exists
	path := fmt.Sprintf("/zones/%s/dns_records?type=A&name=%s", zoneID, url.QueryEscape(name))
	apiResp, err := cloudflareAPI(token, "GET", path, nil)
	if err != nil {
		return err
	}

	var records []cfDNSRecord
	if err := json.Unmarshal(apiResp.Result, &records); err != nil {
		return fmt.Errorf("failed to parse records: %w", err)
	}

	payload := fmt.Sprintf(`{"type":"A","name":"%s","content":"%s","ttl":60,"proxied":false}`, name, ip)

	if len(records) > 0 {
		// Update existing record
		updatePath := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, records[0].ID)
		_, err = cloudflareAPI(token, "PUT", updatePath, strings.NewReader(payload))
	} else {
		// Create new record
		createPath := fmt.Sprintf("/zones/%s/dns_records", zoneID)
		_, err = cloudflareAPI(token, "POST", createPath, strings.NewReader(payload))
	}

	return err
}

func setCloudflareARecords(token, domain, ip string) {
	fmt.Println(theme.RenderBanner("☁ DNS CONFIGURATION ☁"))
	fmt.Println()

	fmt.Printf("%s Configuring DNS for %s → %s\n",
		theme.SymbolSparkle,
		theme.HighlightStyle.Render(domain),
		theme.InfoStyle.Render(ip))
	fmt.Println()

	// Step 1: Get zone ID
	fmt.Printf("  %s Looking up zone...\n", theme.SymbolBolt)
	zoneID, err := cloudflareGetZoneID(token, domain)
	if err != nil {
		fmt.Printf("    %s %s\n", theme.ErrorStyle.Render("✗"), theme.ErrorStyle.Render(err.Error()))
		return
	}

	// Step 2: Upsert A records
	records := []struct {
		name string
		fqdn string
		desc string
	}{
		{"@", domain, "Root domain"},
		{"*", "*." + domain, "Wildcard subdomain"},
	}

	success := 0
	failed := 0

	for _, rec := range records {
		fmt.Printf("  %s Setting %s record (%s)...\n",
			theme.SymbolBolt,
			theme.HighlightStyle.Render(rec.name),
			theme.DimTextStyle.Render(rec.desc))

		if err := cloudflareUpsertARecord(token, zoneID, rec.fqdn, ip); err != nil {
			fmt.Printf("    %s %s\n", theme.ErrorStyle.Render("✗"), theme.ErrorStyle.Render(err.Error()))
			failed++
		} else {
			fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("✓"), theme.SuccessStyle.Render("Configured"))
			success++
		}
	}

	fmt.Println()

	if failed == 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ DNS configuration complete!"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📊 Summary:"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Records set: %d", success)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("TTL: 60 seconds"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Proxied: No (DNS only)"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Propagation: ~1-2 minutes"))
		fmt.Println()

		fmt.Println(theme.InfoStyle.Render("🌐 Your domain is now pointing to:"))
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(domain), theme.InfoStyle.Render(ip))
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render("*."+domain), theme.InfoStyle.Render(ip))
		fmt.Println()
	} else {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠ Completed with %d error(s)", failed)))
		fmt.Println(theme.DimTextStyle.Render("  Check your Cloudflare API token permissions"))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("✨ Verify DNS:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("dig "+domain))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("dig *."+domain))
	fmt.Println()
}

func setVercelARecords(domain, ip string) {
	fmt.Println(theme.RenderBanner("⚡ DNS CONFIGURATION ⚡"))
	fmt.Println()

	fmt.Printf("%s Configuring DNS for %s → %s\n",
		theme.SymbolSparkle,
		theme.HighlightStyle.Render(domain),
		theme.InfoStyle.Render(ip))
	fmt.Println()

	// Records to set: @ (root) and * (wildcard)
	records := []struct {
		name string
		desc string
	}{
		{"@", "Root domain"},
		{"*", "Wildcard subdomain"},
	}

	success := 0
	failed := 0

	for _, rec := range records {
		fmt.Printf("  %s Setting %s record (%s)...\n",
			theme.SymbolBolt,
			theme.HighlightStyle.Render(rec.name),
			theme.DimTextStyle.Render(rec.desc))

		// First, try to remove existing A records for this subdomain
		// We'll ignore errors as the record might not exist
		listCmd := exec.Command("vercel", "dns", "ls", domain)
		output, _ := listCmd.CombinedOutput()

		// Parse output to find existing A records for this subdomain
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				recordID := fields[0]
				recordName := fields[2]
				recordType := fields[3]

				// Check if this is an A record for our subdomain
				if recordType == "A" && recordName == rec.name {
					fmt.Printf("    %s Removing old record %s\n",
						theme.DimTextStyle.Render("↳"),
						theme.DimTextStyle.Render(recordID))
					exec.Command("vercel", "dns", "rm", recordID).Run()
				}
			}
		}

		// Add new A record with lowest TTL (60 seconds)
		addCmd := exec.Command("vercel", "dns", "add", domain, rec.name, "A", ip)
		output, err := addCmd.CombinedOutput()

		if err != nil {
			fmt.Printf("    %s %s\n",
				theme.ErrorStyle.Render("✗"),
				theme.ErrorStyle.Render("Failed"))
			if len(output) > 0 {
				fmt.Printf("      %s\n",
					theme.DimTextStyle.Render(strings.TrimSpace(string(output))))
			}
			failed++
		} else {
			fmt.Printf("    %s %s\n",
				theme.SuccessStyle.Render("✓"),
				theme.SuccessStyle.Render("Configured"))
			success++
		}
	}

	fmt.Println()

	if failed == 0 {
		fmt.Println(theme.SuccessStyle.Render("✓ DNS configuration complete!"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📊 Summary:"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Records added: %d", success)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("TTL: 60 seconds (lowest)"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Propagation: ~1-2 minutes"))
		fmt.Println()

		fmt.Println(theme.InfoStyle.Render("🌐 Your domain is now pointing to:"))
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(domain), theme.InfoStyle.Render(ip))
		fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render("*."+domain), theme.InfoStyle.Render(ip))
		fmt.Println()
	} else {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠ Completed with %d error(s)", failed)))
		fmt.Println(theme.DimTextStyle.Render("  Check your Vercel authentication and domain ownership"))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("✨ Verify DNS:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("dig "+domain))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render("vercel dns ls "+domain))
	fmt.Println()
}

// --- DNS List Command ---

func runDNSList(cmd *cobra.Command, args []string) {
	var targetIP string

	if len(args) >= 1 {
		targetIP = args[0]
	} else {
		ip, err := getPublicIPForDNS()
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("✗ " + err.Error()))
			return
		}
		targetIP = ip
	}

	fmt.Println(theme.RenderBanner("🌐 DNS INVENTORY 🌐"))
	fmt.Println()
	fmt.Printf("%s Scanning all providers for domains → %s\n",
		theme.SymbolSparkle,
		theme.HighlightStyle.Render(targetIP))
	fmt.Println()

	cfg, _ := loadDNSConfig()

	type domainResult struct {
		domain   string
		provider string
		match    bool // A record points to targetIP
	}

	var results []domainResult

	// --- Namecheap ---
	if cfg != nil && cfg.Namecheap.APIKey != "" {
		fmt.Printf("  %s Checking Namecheap...\n", theme.SymbolBolt)
		clientIP, err := getPublicIPForDNS()
		if err == nil {
			params := url.Values{}
			params.Set("PageSize", "100")
			apiResp, err := namecheapAPICall(&cfg.Namecheap, clientIP, "namecheap.domains.getList", params)
			if err == nil {
				for _, d := range apiResp.CommandResponse.DomainGetList.Domains {
					ips, err := net.LookupHost(d.Name)
					match := false
					if err == nil {
						for _, ip := range ips {
							if ip == targetIP {
								match = true
								break
							}
						}
					}
					results = append(results, domainResult{domain: d.Name, provider: "Namecheap", match: match})
				}
			}
		}
	}

	// --- Cloudflare ---
	if cfg != nil && cfg.Cloudflare.APIToken != "" {
		fmt.Printf("  %s Checking Cloudflare...\n", theme.SymbolBolt)
		apiResp, err := cloudflareAPI(cfg.Cloudflare.APIToken, "GET", "/zones?per_page=50&status=active", nil)
		if err == nil {
			var zones []cfZone
			if err := json.Unmarshal(apiResp.Result, &zones); err == nil {
				for _, z := range zones {
					// Skip if already found via Namecheap
					found := false
					for _, r := range results {
						if r.domain == z.Name {
							found = true
							break
						}
					}
					if found {
						continue
					}
					ips, err := net.LookupHost(z.Name)
					match := false
					if err == nil {
						for _, ip := range ips {
							if ip == targetIP {
								match = true
								break
							}
						}
					}
					results = append(results, domainResult{domain: z.Name, provider: "Cloudflare", match: match})
				}
			}
		}
	}

	fmt.Println()

	// Display results
	pointed := []domainResult{}
	notPointed := []domainResult{}
	for _, r := range results {
		if r.match {
			pointed = append(pointed, r)
		} else {
			notPointed = append(notPointed, r)
		}
	}

	if len(pointed) > 0 {
		fmt.Println(theme.GlowStyle.Render(fmt.Sprintf("🌸 Domains pointing to %s:", targetIP)))
		fmt.Println()
		for i, r := range pointed {
			fmt.Printf("  %s %s %s\n",
				theme.HighlightStyle.Render(fmt.Sprintf("%d.", i+1)),
				theme.SuccessStyle.Render(r.domain),
				theme.DimTextStyle.Render("("+r.provider+")"))
		}
		fmt.Println()
	}

	if len(notPointed) > 0 {
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("Other domains (not pointing to %s):", targetIP)))
		fmt.Println()
		for i, r := range notPointed {
			fmt.Printf("  %s %s %s\n",
				theme.DimTextStyle.Render(fmt.Sprintf("%d.", i+1)),
				theme.DimTextStyle.Render(r.domain),
				theme.DimTextStyle.Render("("+r.provider+")"))
		}
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("📊 Summary:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Total domains: %d", len(results))))
	fmt.Printf("  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("Pointing here: %d", len(pointed))))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("Elsewhere: %d", len(notPointed))))
	fmt.Println()
}
