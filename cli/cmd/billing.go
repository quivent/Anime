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

var billingCmd = &cobra.Command{
	Use:   "billing",
	Short: "Check Lambda Cloud billing and pricing",
	Long: `Query your Lambda Cloud account for:
  • Current billing period costs
  • Running instances and their hourly rates
  • GPU pricing lookup
  • Cost estimates

Requires LAMBDA_API_KEY environment variable.

Examples:
  anime billing              # Show current billing info
  anime billing --pricing    # Show GPU pricing table`,
	RunE: runBilling,
}

var (
	showPricing bool
)

func init() {
	billingCmd.Flags().BoolVarP(&showPricing, "pricing", "p", false, "Show GPU pricing table")
	rootCmd.AddCommand(billingCmd)
}

type LambdaInstance struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	InstanceType string  `json:"instance_type_name"`
	Region       string  `json:"region_name"`
	Status       string  `json:"status"`
	IP           string  `json:"ip"`
	Hostname     string  `json:"hostname"`
	Price        float64 `json:"price_cents_per_hour"`
}

type LambdaInstancesResponse struct {
	Data []LambdaInstance `json:"data"`
}

func runBilling(cmd *cobra.Command, args []string) error {
	apiKey := os.Getenv("LAMBDA_API_KEY")
	if apiKey == "" {
		fmt.Println()
		fmt.Println(theme.RenderBanner("💰 BILLING"))
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("⚠️  LAMBDA_API_KEY not set"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Get your API key from:"))
		fmt.Println(theme.HighlightStyle.Render("  https://cloud.lambdalabs.com/api-keys"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Then set it:"))
		fmt.Println(theme.HighlightStyle.Render("  export LAMBDA_API_KEY=your-key-here"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Or add to ~/.bashrc for persistence"))
		fmt.Println()
		return nil
	}

	if showPricing {
		return showGPUPricing(apiKey)
	}

	return showCurrentBilling(apiKey)
}

func showCurrentBilling(apiKey string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("💰 LAMBDA CLOUD BILLING"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Fetching running instances..."))
	fmt.Println()

	// Query Lambda Cloud API for running instances
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

	var instancesResp LambdaInstancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&instancesResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(instancesResp.Data) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  No running instances"))
		fmt.Println()
		return nil
	}

	// Display instances
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🖥️  Running Instances"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	totalCostPerHour := 0.0
	for _, instance := range instancesResp.Data {
		if instance.Status != "active" {
			continue
		}

		pricePerHour := float64(instance.Price) / 100.0
		totalCostPerHour += pricePerHour

		fmt.Printf("  %s\n", theme.HighlightStyle.Render(instance.Name))
		fmt.Printf("    Type:   %s\n", theme.InfoStyle.Render(instance.InstanceType))
		fmt.Printf("    IP:     %s\n", theme.DimTextStyle.Render(instance.IP))
		fmt.Printf("    Region: %s\n", theme.DimTextStyle.Render(instance.Region))
		fmt.Printf("    Cost:   %s\n", theme.SuccessStyle.Render(fmt.Sprintf("$%.2f/hour", pricePerHour)))
		fmt.Println()
	}

	// Show totals
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💵 Cost Summary"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  Total Cost:  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$%.2f/hour", totalCostPerHour)))
	fmt.Printf("  Daily:       %s\n", theme.DimTextStyle.Render(fmt.Sprintf("$%.2f (24 hours)", totalCostPerHour*24)))
	fmt.Printf("  Weekly:      %s\n", theme.DimTextStyle.Render(fmt.Sprintf("$%.2f (7 days)", totalCostPerHour*24*7)))
	fmt.Printf("  Monthly:     %s\n", theme.DimTextStyle.Render(fmt.Sprintf("$%.2f (30 days)", totalCostPerHour*24*30)))
	fmt.Println()

	// Try to get session cost from server if we can connect
	cfg, err := config.Load()
	if err == nil {
		target := cfg.GetAlias("lambda")
		if target != "" {
			parts := strings.Split(target, "@")
			if len(parts) == 2 {
				user, host := parts[0], parts[1]
				if sshClient, err := ssh.NewClient(host, user, ""); err == nil {
					defer sshClient.Close()

					// Get uptime
					uptimeOutput, err := sshClient.RunCommand("cat /proc/uptime | awk '{print $1}'")
					if err == nil {
						var uptime float64
						fmt.Sscanf(strings.TrimSpace(uptimeOutput), "%f", &uptime)
						hours := uptime / 3600.0
						sessionCost := hours * totalCostPerHour

						fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
						fmt.Println(theme.InfoStyle.Render("⏱️  Current Session"))
						fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
						fmt.Println()
						fmt.Printf("  Uptime:      %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.2f hours", hours)))
						fmt.Printf("  Session Cost: %s\n", theme.WarningStyle.Render(fmt.Sprintf("$%.2f", sessionCost)))
						fmt.Println()
					}
				}
			}
		}
	}

	return nil
}

func showGPUPricing(apiKey string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("💰 GPU PRICING"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Fetching current Lambda Cloud pricing..."))
	fmt.Println()

	// Query Lambda Cloud API for instance types
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://cloud.lambdalabs.com/api/v1/instance-types", nil)
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected API response format")
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎮 Available GPU Instances"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	for instanceType, details := range data {
		detailsMap, ok := details.(map[string]interface{})
		if !ok {
			continue
		}

		description, _ := detailsMap["instance_type"].(map[string]interface{})
		if description == nil {
			continue
		}

		name, _ := description["name"].(string)
		priceCents, _ := description["price_cents_per_hour"].(float64)

		if name != "" && priceCents > 0 {
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(name))
			fmt.Printf("    ID:    %s\n", theme.DimTextStyle.Render(instanceType))
			fmt.Printf("    Cost:  %s\n", theme.SuccessStyle.Render(fmt.Sprintf("$%.2f/hour", priceCents/100.0)))
			fmt.Println()
		}
	}

	return nil
}
