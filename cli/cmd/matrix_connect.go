package cmd

import (
	"fmt"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
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
	Example: `  anime matrix connect --url http://localhost:8065 --user admin --password secret
  anime matrix connect --url https://chat.example.com --token <personal-access-token>`,
	RunE: runMatrixConnect,
}

func init() {
	matrixConnectCmd.Flags().StringVar(&mxConnURL, "url", "http://localhost:8065", "Mattermost server URL")
	matrixConnectCmd.Flags().StringVarP(&mxConnUser, "user", "u", "", "Username or email")
	matrixConnectCmd.Flags().StringVarP(&mxConnPassword, "password", "p", "", "Password")
	matrixConnectCmd.Flags().StringVarP(&mxConnToken, "token", "T", "", "Personal access token")
	matrixCmd.AddCommand(matrixConnectCmd)
}

func runMatrixConnect(cmd *cobra.Command, args []string) error {
	t.Section("CONNECT")

	if mxConnToken == "" && (mxConnUser == "" || mxConnPassword == "") {
		t.Fail("provide --token or --user + --password")
		fmt.Println()
		fmt.Println("  " + t.Gold.S("anime matrix connect --url http://host:8065 --user admin --password pass"))
		fmt.Println("  " + t.Gold.S("anime matrix connect --url http://host:8065 --token <token>"))
		fmt.Println()
		return fmt.Errorf("missing credentials")
	}

	client := mmClient(mxConnURL, "")

	t.Info("testing " + mxConnURL + "…")
	ver, err := client.ServerVersion()
	if err != nil {
		t.Fail("cannot reach server: " + err.Error())
		return fmt.Errorf("unreachable: %w", err)
	}
	t.Ok("reachable  " + t.Dim("v"+ver))

	token := mxConnToken
	if token == "" {
		t.Info("logging in as " + mxConnUser + "…")
		token, err = client.Login(mxConnUser, mxConnPassword)
		if err != nil {
			return err
		}
		t.Ok("authenticated")
	}

	client = mmClient(mxConnURL, token)
	t.Info("verifying identity…")
	me, err := client.GetMe()
	if err != nil {
		return err
	}
	t.Ok("logged in as " + t.Bold(t.Gold.S("@"+me.Username)))

	teamID, teamName := "", ""
	teams, err := client.GetTeams(0, 10)
	if err == nil && len(teams) > 0 {
		teamID = teams[0].ID
		teamName = teams[0].Name
		t.Info(t.Dim("team: " + teams[0].DisplayName))
	}
	if me.IsAdmin() {
		t.Info(t.Gold.S("admin"))
	}

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
	t.Rule()
	t.KV("server", t.Cyan.S(mxConnURL))
	t.KV("user", "@"+me.Username)
	if teamName != "" {
		t.KV("team", teamName)
	}
	t.Ok("connected")
	fmt.Println()
	return nil
}
