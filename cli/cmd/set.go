package cmd

import (
	"fmt"
	"net"
	"os/exec"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <alias> [target]",
	Short: "Set an alias for a server",
	Long: `Create or update an alias for a server target.

This allows you to use short names instead of typing full server addresses.

When run on the server itself without a target, it will auto-detect the server's IP.

Examples:
  anime set lambda                          # Auto-detect IP (when run on server)
  anime set lambda 209.20.159.132           # Create alias 'lambda' -> '209.20.159.132'
  anime set lambda ubuntu@209.20.159.132    # Create alias with user
  anime push lambda                          # Use the alias
`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runSet,
}

func showSetUsage() {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("❌ Missing required arguments"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
	fmt.Println(theme.HighlightStyle.Render("  anime set <alias> [target]"))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
	fmt.Println(theme.DimTextStyle.Render("  anime set lambda 209.20.159.132           # Create alias"))
	fmt.Println(theme.DimTextStyle.Render("  anime set lambda ubuntu@209.20.159.132    # Alias with user"))
	fmt.Println(theme.DimTextStyle.Render("  anime set lambda                          # Auto-detect IP (on server)"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Options:"))
	fmt.Println(theme.DimTextStyle.Render("  anime set --list                          # List all aliases"))
	fmt.Println(theme.DimTextStyle.Render("  anime set --delete lambda                 # Delete an alias"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
	fmt.Println(theme.DimTextStyle.Render("  anime push <alias>                        # Push to aliased server"))
	fmt.Println(theme.DimTextStyle.Render("  anime list                                # List all servers"))
	fmt.Println()
}

var (
	deleteAlias bool
	listAliases bool
)

func init() {
	setCmd.Flags().BoolVarP(&deleteAlias, "delete", "d", false, "Delete an alias")
	setCmd.Flags().BoolVarP(&listAliases, "list", "l", false, "List all aliases")
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// List aliases
	if listAliases {
		return showAliases(cfg)
	}

	// Delete alias
	if deleteAlias {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Delete requires an alias name"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime set --delete <alias>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Example:"))
			fmt.Println(theme.DimTextStyle.Render("  anime set --delete lambda"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Tip:"))
			fmt.Println(theme.DimTextStyle.Render("  List aliases with: anime set --list"))
			fmt.Println()
			return fmt.Errorf("delete requires exactly one argument: alias name")
		}
		return deleteAliasCmd(cfg, args[0])
	}

	// Set alias
	if len(args) == 0 {
		showSetUsage()
		return fmt.Errorf("set requires at least one argument: alias name")
	}

	alias := args[0]
	var target string

	if len(args) == 2 {
		target = args[1]
	} else if len(args) == 1 {
		// Only one arg provided - try to auto-detect
		detectedIP, err := detectServerIP()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w\n\nUsage: anime set %s <server-ip>", err, alias)
		}
		target = detectedIP
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("🔍 Auto-detected server IP: " + detectedIP))
	} else {
		showSetUsage()
		return fmt.Errorf("set requires one or two arguments: alias [target]")
	}

	// Check if alias already exists
	existing := cfg.GetAlias(alias)
	isUpdate := existing != ""

	// Set the alias
	cfg.SetAlias(alias, target)

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Println()
	if isUpdate {
		fmt.Println(theme.SuccessStyle.Render("✓ Alias updated"))
		fmt.Println()
		fmt.Printf("  %s: %s → %s\n",
			theme.HighlightStyle.Render(alias),
			theme.DimTextStyle.Render(existing),
			theme.InfoStyle.Render(target))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ Alias created"))
		fmt.Println()
		fmt.Printf("  %s → %s\n",
			theme.HighlightStyle.Render(alias),
			theme.InfoStyle.Render(target))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Usage: ") + theme.HighlightStyle.Render("anime push "+alias))
	fmt.Println()

	return nil
}

func showAliases(cfg *config.Config) error {
	aliases := cfg.ListAliases()

	if len(aliases) == 0 {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("No aliases configured"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Create one with: ") + theme.HighlightStyle.Render("anime set <alias> <target>"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📋 Configured Aliases"))
	fmt.Println()

	// Sort aliases by name for consistent output
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	// Find max alias length for alignment
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Print each alias with source indicator
	embeddedCount := 0
	for _, name := range names {
		padding := strings.Repeat(" ", maxLen-len(name))
		source := ""
		if cfg.IsEmbeddedAlias(name) {
			// Check if runtime config overrides the embedded default
			if cfg.Aliases != nil {
				if _, overridden := cfg.Aliases[name]; overridden {
					source = theme.DimTextStyle.Render(" (override)")
				} else {
					source = theme.DimTextStyle.Render(" (embedded)")
					embeddedCount++
				}
			} else {
				source = theme.DimTextStyle.Render(" (embedded)")
				embeddedCount++
			}
		}
		fmt.Printf("  %s%s  →  %s%s\n",
			theme.HighlightStyle.Render(name),
			padding,
			theme.InfoStyle.Render(aliases[name]),
			source)
	}

	fmt.Println()
	if embeddedCount > 0 {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Total: %d alias(es) (%d embedded)", len(aliases), embeddedCount)))
	} else {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Total: %d alias(es)", len(aliases))))
	}
	fmt.Println()

	return nil
}

func deleteAliasCmd(cfg *config.Config, alias string) error {
	// Get the target before deleting to show in confirmation
	target := cfg.GetAlias(alias)
	if target == "" {
		return fmt.Errorf("alias '%s' not found", alias)
	}

	// Delete the alias
	if err := cfg.DeleteAlias(alias); err != nil {
		return err
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Alias deleted"))
	fmt.Println()
	fmt.Printf("  %s → %s\n",
		theme.HighlightStyle.Render(alias),
		theme.DimTextStyle.Render(target))
	fmt.Println()

	return nil
}

// detectServerIP attempts to detect the server's public IP address
func detectServerIP() (string, error) {
	// Try multiple methods to get the server's public IP

	// Method 1: Use curl to get public IP from ipify
	cmd := exec.Command("curl", "-s", "https://api.ipify.org")
	if output, err := cmd.Output(); err == nil {
		ip := strings.TrimSpace(string(output))
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	// Method 2: Try ifconfig.me
	cmd = exec.Command("curl", "-s", "https://ifconfig.me")
	if output, err := cmd.Output(); err == nil {
		ip := strings.TrimSpace(string(output))
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	// Method 3: Get local network interface IP (fallback)
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not detect server IP - please provide it manually")
}
