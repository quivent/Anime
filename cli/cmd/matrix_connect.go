package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/synapse"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxConnURL      string
	mxConnDomain   string
	mxConnUser     string
	mxConnPassword string
	mxConnToken    string
)

var matrixConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an existing Matrix homeserver",
	Long: `Connect to a running Matrix homeserver by URL.
Authenticates, verifies the connection, and saves to ~/.matrix/config.yaml.`,
	Example: `  anime matrix connect --url http://localhost:8008 --user admin --password s3cret
  anime matrix connect --url https://matrix.example.com --token syt_...
  anime matrix connect --url http://192.168.1.50:8008 -u admin -p admin`,
	RunE: runMatrixConnect,
}

func init() {
	matrixConnectCmd.Flags().StringVar(&mxConnURL, "url", "http://localhost:8008", "Homeserver URL")
	matrixConnectCmd.Flags().StringVar(&mxConnDomain, "domain", "", "Server domain (auto-detected if not set)")
	matrixConnectCmd.Flags().StringVarP(&mxConnUser, "user", "u", "", "Admin username")
	matrixConnectCmd.Flags().StringVarP(&mxConnPassword, "password", "p", "", "Admin password")
	matrixConnectCmd.Flags().StringVarP(&mxConnToken, "token", "t", "", "Access token (skip login)")

	matrixCmd.AddCommand(matrixConnectCmd)
}

func runMatrixConnect(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CONNECT"))
	fmt.Println()

	if mxConnToken == "" && (mxConnUser == "" || mxConnPassword == "") {
		fmt.Println(theme.ErrorStyle.Render("  Provide --token or --user + --password"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime matrix connect --url http://host:8008 --user admin --password pass"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime matrix connect --url http://host:8008 --token syt_..."))
		fmt.Println()
		return fmt.Errorf("missing credentials")
	}

	// Test connectivity
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Testing "+mxConnURL+"..."))
	if !synapse.IsHealthy(mxConnURL) {
		client := matrixapi.NewClient(mxConnURL, "")
		if _, err := client.ServerVersion(); err != nil {
			fmt.Printf("  %s %s\n", theme.SymbolError, theme.ErrorStyle.Render("Cannot reach server"))
			return fmt.Errorf("unreachable: %w", err)
		}
	}
	fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Reachable"))

	// Authenticate
	token := mxConnToken
	if token == "" {
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Logging in as "+mxConnUser+"..."))
		client := matrixapi.NewClient(mxConnURL, "")
		var err error
		token, err = client.Login(mxConnUser, mxConnPassword)
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Authenticated"))
	}

	// Identify
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Verifying identity..."))
	client := matrixapi.NewClient(mxConnURL, token)
	userID, err := client.WhoAmI()
	if err != nil {
		return err
	}
	fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Logged in as"), theme.HighlightStyle.Render(userID))

	// Auto-detect domain
	domain := mxConnDomain
	if domain == "" {
		_, domain = matrixSplitUserID(userID)
		if domain == "" {
			domain = "localhost"
		}
	}

	// Version
	if ver, err := client.ServerVersion(); err == nil {
		fmt.Printf("  %s %s %s\n", theme.SymbolInfo, theme.DimTextStyle.Render("Server:"), theme.DimTextStyle.Render(ver))
	}

	// Admin access
	admin := matrixapi.NewAdminClient(mxConnURL, token, domain)
	hasAdmin := false
	if users, err := admin.ListUsers(0, 1); err == nil {
		hasAdmin = true
		fmt.Printf("  %s %s (%d users)\n", theme.SymbolShield, theme.SuccessStyle.Render("Admin API accessible"), users.Total)
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("No admin API access"))
	}

	// Save
	cfg, _ := matrixcfg.Load()
	cfg.Homeserver = matrixcfg.HomeserverConfig{
		URL: mxConnURL, Domain: domain, AdminToken: token, AdminUser: userID,
	}
	cfg.Save()

	fmt.Println()
	fmt.Println(matrixSeparator())
	fmt.Println(theme.SuccessStyle.Render("  Connected!"))
	fmt.Println(matrixSeparator())
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Server:"), theme.InfoStyle.Render(mxConnURL))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Domain:"), theme.DimTextStyle.Render(domain))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("User:"), theme.DimTextStyle.Render(userID))
	if hasAdmin {
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.SuccessStyle.Render("yes"))
	}
	fmt.Println()
	return nil
}
