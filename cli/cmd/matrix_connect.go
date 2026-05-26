package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxConnURL      string
	mxConnUser     string
	mxConnPassword string
	mxConnToken    string
)

var matrixConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a Mattermost server",
	Long: `Authenticate to a running Mattermost server and save credentials to ~/.matrix/config.yaml.
Supports login (username + password) or direct token.`,
	Example: `  anime matrix connect --url http://localhost:8065 --user admin --password secret
  anime matrix connect --url https://chat.example.com --token <personal-access-token>`,
	RunE: runMatrixConnect,
}

func init() {
	matrixConnectCmd.Flags().StringVar(&mxConnURL, "url", "http://localhost:8065", "Mattermost server URL")
	matrixConnectCmd.Flags().StringVarP(&mxConnUser, "user", "u", "", "Username or email")
	matrixConnectCmd.Flags().StringVarP(&mxConnPassword, "password", "p", "", "Password")
	matrixConnectCmd.Flags().StringVarP(&mxConnToken, "token", "t", "", "Personal access token")

	matrixCmd.AddCommand(matrixConnectCmd)
}

func runMatrixConnect(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("CONNECT"))
	fmt.Println()

	if mxConnToken == "" && (mxConnUser == "" || mxConnPassword == "") {
		fmt.Println(theme.ErrorStyle.Render("  Provide --token or --user + --password"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime matrix connect --url http://host:8065 --user admin --password pass"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime matrix connect --url http://host:8065 --token <token>"))
		fmt.Println()
		return fmt.Errorf("missing credentials")
	}

	client := mmClient(mxConnURL, "")

	// Test connectivity
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Testing "+mxConnURL+"..."))
	ver, err := client.ServerVersion()
	if err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolError, theme.ErrorStyle.Render("Cannot reach server: "+err.Error()))
		return fmt.Errorf("unreachable: %w", err)
	}
	fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Reachable"), theme.DimTextStyle.Render("v"+ver))

	// Authenticate
	token := mxConnToken
	if token == "" {
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Logging in as "+mxConnUser+"..."))
		token, err = client.Login(mxConnUser, mxConnPassword)
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Authenticated"))
	}

	// Verify identity
	client = mmClient(mxConnURL, token)
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Verifying identity..."))
	me, err := client.GetMe()
	if err != nil {
		return err
	}
	fmt.Printf("  %s %s %s\n", theme.SymbolSuccess,
		theme.SuccessStyle.Render("Logged in as"),
		theme.HighlightStyle.Render("@"+me.Username))

	// Get teams
	teamID := ""
	teamName := ""
	teams, err := client.GetTeams(0, 10)
	if err == nil && len(teams) > 0 {
		teamID = teams[0].ID
		teamName = teams[0].Name
		fmt.Printf("  %s %s\n", theme.SymbolInfo,
			theme.DimTextStyle.Render("Team: "+teams[0].DisplayName))
	}

	if me.IsAdmin() {
		fmt.Printf("  %s %s\n", theme.SymbolShield, theme.SuccessStyle.Render("Admin"))
	}

	// Save
	cfg, _ := mmcfg.Load()
	cfg.Server = mmcfg.ServerConfig{
		URL:      mxConnURL,
		Token:    token,
		Username: me.Username,
		TeamID:   teamID,
		TeamName: teamName,
	}
	cfg.Save()

	fmt.Println()
	fmt.Println(matrixSeparator())
	fmt.Println(theme.SuccessStyle.Render("  Connected!"))
	fmt.Println(matrixSeparator())
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Server:"), theme.InfoStyle.Render(mxConnURL))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("User:"), theme.DimTextStyle.Render("@"+me.Username))
	if teamName != "" {
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Team:"), theme.DimTextStyle.Render(teamName))
	}
	fmt.Println()
	return nil
}
