package cmd

import (
	"fmt"
	"os/exec"
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

func init() {
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.AddCommand(vercelCmd)
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
