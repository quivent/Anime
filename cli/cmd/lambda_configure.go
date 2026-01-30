package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var lambdaConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Lambda Cloud API key and discover servers",
	Long: `Set your Lambda Cloud API key and automatically discover running instances.

This command will:
  1. Prompt for your Lambda Cloud API key
  2. Save it securely in your anime config
  3. Query Lambda Cloud API for running instances
  4. Auto-add discovered instances to your server list

Get your API key from: https://cloud.lambdalabs.com/api-keys`,
	RunE: runLambdaConfigure,
}

func init() {
	lambdaCmd.AddCommand(lambdaConfigureCmd)
}

func runLambdaConfigure(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ LAMBDA CLOUD CONFIGURATION ⚡"))
	fmt.Println()

	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Check if API key already exists
	if cfg.APIKeys.LambdaLabs != "" {
		fmt.Println(theme.InfoStyle.Render("🔑 Lambda API key is already configured"))
		fmt.Println()
		fmt.Print(theme.HighlightStyle.Render("Do you want to update it? (y/N): "))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			// Skip to discovery with existing key
			return discoverInstances(cfg, cfg.APIKeys.LambdaLabs)
		}
		fmt.Println()
	}

	// Prompt for API key
	fmt.Println(theme.InfoStyle.Render("📝 Lambda Cloud API Configuration"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Get your API key from:"))
	fmt.Println(theme.HighlightStyle.Render("  https://cloud.lambdalabs.com/api-keys"))
	fmt.Println()
	fmt.Print(theme.GlowStyle.Render("Enter your Lambda API key: "))

	reader := bufio.NewReader(os.Stdin)
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Save API key
	cfg.APIKeys.LambdaLabs = apiKey
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ API key saved successfully"))
	fmt.Println()

	// Discover instances
	return discoverInstances(cfg, apiKey)
}

func discoverInstances(cfg *config.Config, apiKey string) error {
	fmt.Println(theme.InfoStyle.Render("🔍 Discovering Lambda Cloud instances..."))
	fmt.Println()

	// Query Lambda Cloud API
	req, err := http.NewRequest("GET", "https://cloud.lambdalabs.com/api/v1/instances", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query Lambda API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Lambda API returned status %d (check your API key)", resp.StatusCode)
	}

	var instancesResp LambdaInstancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&instancesResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(instancesResp.Data) == 0 {
		fmt.Println(theme.WarningStyle.Render("⚠️  No running instances found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Launch an instance at:"))
		fmt.Println(theme.HighlightStyle.Render("  https://cloud.lambdalabs.com/instances"))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Found %d running instance(s)", len(instancesResp.Data))))
	fmt.Println()

	// Display instances
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📡 Running Instances"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	addedCount := 0
	updatedCount := 0

	for _, instance := range instancesResp.Data {
		if instance.Status != "active" {
			continue
		}

		// Use instance name or generate one from ID
		serverName := instance.Name
		if serverName == "" {
			serverName = "lambda-" + instance.ID[:8]
		}

		// Calculate cost per hour
		costPerHour := instance.Price / 100.0

		// Display instance info
		fmt.Printf("  🖥️  %s\n",
			theme.HighlightStyle.Render(serverName))
		fmt.Printf("    IP:       %s\n", theme.InfoStyle.Render(instance.IP))
		fmt.Printf("    Type:     %s\n", theme.DimTextStyle.Render(instance.InstanceType))
		fmt.Printf("    Region:   %s\n", theme.DimTextStyle.Render(instance.Region))
		fmt.Printf("    Cost:     %s\n", theme.DimTextStyle.Render(fmt.Sprintf("$%.2f/hr", costPerHour)))

		// Check if server already exists
		existingServer, err := cfg.GetServer(serverName)
		if err == nil {
			// Server exists, update it
			existingServer.Host = instance.IP
			existingServer.CostPerHour = costPerHour
			cfg.UpdateServer(serverName, *existingServer)
			fmt.Printf("    Status:   %s\n", theme.SuccessStyle.Render("✓ Updated"))
			updatedCount++
		} else {
			// Add new server
			server := config.Server{
				Name:        serverName,
				Host:        instance.IP,
				User:        "ubuntu",
				SSHKey:      "",
				CostPerHour: costPerHour,
				Modules:     []string{},
			}
			cfg.AddServer(server)
			fmt.Printf("    Status:   %s\n", theme.SuccessStyle.Render("✓ Added"))
			addedCount++
		}
		fmt.Println()
	}

	// Save updated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Summary
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Summary"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  New servers added:      %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", addedCount)))
	fmt.Printf("  Existing servers updated: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", updatedCount)))
	fmt.Println()

	// Next steps
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Next Steps"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime list"))
	fmt.Println(theme.DimTextStyle.Render("    View all configured servers"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime ssh <server-name>"))
	fmt.Println(theme.DimTextStyle.Render("    Connect to a server"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <server-name> <packages>"))
	fmt.Println(theme.DimTextStyle.Render("    Install packages on a server"))
	fmt.Println()

	return nil
}
