package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/embeddb"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync server IPs from Lambda Cloud API",
	Long: `Query Lambda Cloud API and update all server IPs, aliases, and SSH config.

This command:
  1. Queries Lambda API for all running instances
  2. Updates server IPs in anime config
  3. Updates aliases to point to new IPs
  4. Updates ~/.ssh/config host entries
  5. Removes stale known_hosts entries for changed IPs
  6. Registers the anime embedded key on any new instances

Run this after reprovisioning any Lambda instance.`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	apiKey := cfg.APIKeys.LambdaLabs
	if apiKey == "" {
		apiKey = os.Getenv("LAMBDA_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("no Lambda API key configured. Run: anime lambda configure")
	}

	fmt.Println(theme.InfoStyle.Render("🔄 Syncing server IPs from Lambda Cloud..."))
	fmt.Println()

	instances, err := fetchLambdaInstances(apiKey)
	if err != nil {
		return fmt.Errorf("Lambda API error: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠ No running instances found"))
		return nil
	}

	changed := 0
	registered := 0

	for _, instance := range instances {
		if instance.Status != "active" || instance.IP == "" {
			continue
		}

		name := instance.Name
		if name == "" {
			name = "lambda-" + instance.ID[:8]
		}

		// Check if server exists and IP changed
		srv, err := cfg.GetServer(name)
		if err == nil && srv.Host != instance.IP {
			oldIP := srv.Host
			srv.Host = instance.IP
			cfg.UpdateServer(name, *srv)

			// Update alias too
			cfg.SetAlias(strings.ToLower(name), "ubuntu@"+instance.IP)

			// Remove stale known_hosts
			removeKnownHost(oldIP)
			removeKnownHost(instance.IP)

			// Update SSH config
			updateSSHConfigHost(name, instance.IP)

			fmt.Printf("  %s  %s → %s\n",
				theme.HighlightStyle.Render(name),
				theme.DimTextStyle.Render(oldIP),
				theme.SuccessStyle.Render(instance.IP))
			changed++
		} else if err != nil {
			// New server, add it
			newSrv := config.Server{
				Name:        name,
				Host:        instance.IP,
				User:        "ubuntu",
				CostPerHour: instance.Price / 100.0,
			}
			cfg.AddServer(newSrv)
			cfg.SetAlias(strings.ToLower(name), "ubuntu@"+instance.IP)
			fmt.Printf("  %s  %s (new)\n",
				theme.HighlightStyle.Render(name),
				theme.SuccessStyle.Render(instance.IP))
			changed++
		} else {
			fmt.Printf("  %s  %s (unchanged)\n",
				theme.DimTextStyle.Render(name),
				theme.DimTextStyle.Render(instance.IP))
		}

		// Register embedded key on instance
		if registerEmbeddedKeyOnHost("ubuntu@" + instance.IP) {
			registered++
		}
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	if changed > 0 {
		fmt.Printf(theme.SuccessStyle.Render("✓ Updated %d server(s)")+" ", changed)
	}
	if registered > 0 {
		fmt.Printf(theme.SuccessStyle.Render("✓ Registered key on %d server(s)"), registered)
	}
	if changed == 0 && registered == 0 {
		fmt.Print(theme.SuccessStyle.Render("✓ All servers up to date"))
	}
	fmt.Println()

	return nil
}

// fetchLambdaInstances queries Lambda API for running instances
func fetchLambdaInstances(apiKey string) ([]LambdaInstance, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://cloud.lambdalabs.com/api/v1/instances", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("API key expired or invalid (status %d). Get a new one at https://cloud.lambdalabs.com/api-keys", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result LambdaInstancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// resolveIPFromLambdaAPI tries to look up a server's current IP from Lambda API
// Returns the IP if found, empty string otherwise. Does not print anything.
func resolveIPFromLambdaAPI(serverName string) string {
	cfg, err := config.Load()
	if err != nil {
		return ""
	}

	apiKey := cfg.APIKeys.LambdaLabs
	if apiKey == "" {
		apiKey = os.Getenv("LAMBDA_API_KEY")
	}
	if apiKey == "" {
		return ""
	}

	instances, err := fetchLambdaInstances(apiKey)
	if err != nil {
		return ""
	}

	nameLower := strings.ToLower(serverName)
	for _, inst := range instances {
		if inst.Status == "active" && inst.IP != "" {
			instName := strings.ToLower(inst.Name)
			if instName == nameLower {
				return inst.IP
			}
		}
	}
	return ""
}

// removeKnownHost removes a host from known_hosts
func removeKnownHost(host string) {
	if host == "" {
		return
	}
	exec.Command("ssh-keygen", "-R", host).Run()
}

// updateSSHConfigHost updates a host entry in ~/.ssh/config
func updateSSHConfigHost(name, newIP string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	sshConfig := filepath.Join(home, ".ssh", "config")
	data, err := os.ReadFile(sshConfig)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	nameLower := strings.ToLower(name)
	inBlock := false
	modified := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(trimmed, "host ") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && fields[1] == nameLower {
				inBlock = true
			} else {
				inBlock = false
			}
		}
		if inBlock && strings.Contains(strings.ToLower(strings.TrimSpace(line)), "hostname") {
			// Preserve indentation
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = indent + "HostName " + newIP
			modified = true
			inBlock = false
		}
	}

	if modified {
		os.WriteFile(sshConfig, []byte(strings.Join(lines, "\n")), 0600)
	}
}

// registerEmbeddedKeyOnHost tries to register the anime embedded key on a host.
// First tests if the embedded key already works. If not, tries system keys to inject it.
// Returns true if key was newly registered.
func registerEmbeddedKeyOnHost(target string) bool {
	// Test if embedded key already works
	if keyPath, cleanup, err := GetEmbeddedSSHKeyPath(); err == nil {
		defer cleanup()
		testCmd := exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			"-o", "BatchMode=yes",
			"-i", keyPath,
			target, "echo ok")
		if testCmd.Run() == nil {
			return false // already works
		}

		// Embedded key doesn't work - try to register it via system keys
		pubKey := getEmbeddedPubKey()
		if pubKey == "" {
			return false
		}

		// Try SSH with no specific key (use agent + default keys)
		injectCmd := exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
			"-o", "BatchMode=yes",
			target,
			fmt.Sprintf("mkdir -p ~/.ssh && grep -qF '%s' ~/.ssh/authorized_keys 2>/dev/null || echo '%s' >> ~/.ssh/authorized_keys", pubKey, pubKey))
		if injectCmd.Run() == nil {
			return true
		}
	}
	return false
}

// getEmbeddedPubKey returns the embedded public key from the embeddb
func getEmbeddedPubKey() string {
	db, err := embeddb.DB()
	if err != nil {
		return ""
	}
	data := db.Get(keyPublic)
	if data == nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
