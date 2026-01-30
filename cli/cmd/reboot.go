package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot [server]",
	Short: "Reboot remote server",
	Long: `Reboot a remote GPU server (requires sudo permissions).

This is useful for:
  • Fixing NVIDIA driver/library version mismatches
  • Applying kernel updates
  • Clearing stuck processes
  • Resetting GPU states

Examples:
  anime reboot              # Reboot configured lambda server
  anime reboot lambda       # Reboot lambda server by alias
  anime reboot 209.20.159.132  # Reboot by IP`,
	RunE: runReboot,
}

func init() {
	rootCmd.AddCommand(rebootCmd)
}

func runReboot(cmd *cobra.Command, args []string) error {
	apiKey := os.Getenv("LAMBDA_API_KEY")

	// If no API key, try direct SSH method
	if apiKey == "" {
		return rebootViaSSH(args)
	}

	// Use Lambda Cloud API (preferred method)
	return rebootViaAPI(apiKey, args)
}

func rebootViaAPI(apiKey string, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔄 SERVER REBOOT"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Fetching running instances..."))
	fmt.Println()

	// Get instances from Lambda Cloud API
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://cloud.lambdalabs.com/api/v1/instances", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query Lambda API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var instancesResp struct {
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			IP       string `json:"ip"`
			Hostname string `json:"hostname"`
			Status   string `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&instancesResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(instancesResp.Data) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠️  No running instances found"))
		fmt.Println()
		return nil
	}

	// Find instance to reboot
	var instanceID, instanceName string
	if len(args) > 0 {
		searchTerm := args[0]
		// Search by name, IP, or hostname
		for _, inst := range instancesResp.Data {
			if inst.Name == searchTerm || inst.IP == searchTerm || strings.Contains(inst.Hostname, searchTerm) {
				instanceID = inst.ID
				instanceName = inst.Name
				break
			}
		}
		if instanceID == "" {
			fmt.Println(theme.ErrorStyle.Render("❌ Instance not found: " + searchTerm))
			fmt.Println()
			return fmt.Errorf("instance not found")
		}
	} else {
		// Default to first instance
		instanceID = instancesResp.Data[0].ID
		instanceName = instancesResp.Data[0].Name
	}

	fmt.Printf("  Instance: %s\n", theme.HighlightStyle.Render(instanceName))
	fmt.Printf("  ID:       %s\n", theme.DimTextStyle.Render(instanceID))
	fmt.Println()

	// Confirm
	fmt.Print(theme.WarningStyle.Render("⚠️  This will reboot the instance. Continue? (Y/n): "))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "n" || response == "no" {
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("❌ Cancelled"))
		fmt.Println()
		return nil
	}

	// Trigger reboot via API
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Triggering reboot via Lambda Cloud API..."))

	rebootReq, err := http.NewRequest("POST", fmt.Sprintf("https://cloud.lambdalabs.com/api/v1/instances/%s/restart", instanceID), nil)
	if err != nil {
		return fmt.Errorf("failed to create reboot request: %w", err)
	}

	rebootReq.Header.Set("Authorization", "Bearer "+apiKey)
	rebootReq.Header.Set("Content-Type", "application/json")

	rebootResp, err := client.Do(rebootReq)
	if err != nil {
		return fmt.Errorf("failed to send reboot request: %w", err)
	}
	defer rebootResp.Body.Close()

	if rebootResp.StatusCode != 200 {
		body, _ := io.ReadAll(rebootResp.Body)
		return fmt.Errorf("API error (status %d): %s", rebootResp.StatusCode, string(body))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Reboot initiated successfully"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Instance is rebooting..."))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The instance will be back online in ~30-60 seconds"))
	fmt.Println(theme.DimTextStyle.Render("  Wait a moment, then test with:"))
	fmt.Println(theme.HighlightStyle.Render("    anime usage"))
	fmt.Println()

	return nil
}

func rebootViaSSH(args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	// Get target server
	var target string
	if len(args) > 0 {
		server := args[0]
		if strings.Contains(server, "@") {
			target = server
		} else if strings.Contains(server, ".") {
			target = "ubuntu@" + server
		} else {
			target = cfg.GetAlias(server)
			if target == "" {
				if s, err := cfg.GetServer(server); err == nil {
					target = fmt.Sprintf("%s@%s", s.User, s.Host)
				}
			}
		}
	} else {
		target = cfg.GetAlias("lambda")
		if target == "" {
			if s, err := cfg.GetServer("lambda"); err == nil {
				target = fmt.Sprintf("%s@%s", s.User, s.Host)
			}
		}
	}

	if target == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ No server specified and LAMBDA_API_KEY not set"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Option 1: Set Lambda API key (recommended):"))
		fmt.Println(theme.HighlightStyle.Render("  export LAMBDA_API_KEY=your-key"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Option 2: Specify server:"))
		fmt.Println(theme.HighlightStyle.Render("  anime reboot lambda"))
		fmt.Println()
		return fmt.Errorf("no server specified")
	}

	parts := strings.Split(target, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", target)
	}
	user, host := parts[0], parts[1]

	fmt.Println()
	fmt.Println(theme.RenderBanner("🔄 SERVER REBOOT"))
	fmt.Println()
	fmt.Printf("  Target: %s\n", theme.HighlightStyle.Render(target))
	fmt.Println()
	fmt.Print(theme.WarningStyle.Render("⚠️  This will reboot the server. Continue? (Y/n): "))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "n" || response == "no" {
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("❌ Cancelled"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Connecting to server..."))
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer sshClient.Close()

	fmt.Println(theme.InfoStyle.Render("Sending reboot command..."))
	sshClient.RunCommand("sudo reboot || reboot")

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Reboot command sent"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Server is rebooting..."))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The server will be back online in ~30-60 seconds"))
	fmt.Println(theme.DimTextStyle.Render("  Wait a moment, then test with:"))
	fmt.Println(theme.HighlightStyle.Render("    anime usage"))
	fmt.Println()

	return nil
}
