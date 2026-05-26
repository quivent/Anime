package cmd

import (
	"fmt"
	"time"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxUserPassword    string
	mxUserDisplayName string
	mxUserAdmin       bool
	mxRevokeAdmin     bool
)

var matrixUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "u"},
	Short:   "Manage Matrix user accounts",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixUsersAddCmd = &cobra.Command{
	Use:  "add <username>",
	Short: "Create a new user",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		cfg, _ := matrixcfg.Load()
		if mxUserPassword == "" {
			mxUserPassword = matrixGeneratePassword(16)
		}
		dn := mxUserDisplayName
		if dn == "" {
			dn = username
		}
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Creating @"+username+":"+cfg.Homeserver.Domain+"..."))
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		if err := admin.CreateUser(username, mxUserPassword, dn, mxUserAdmin); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Created"))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("User ID:"), theme.DimTextStyle.Render(fmt.Sprintf("@%s:%s", username, cfg.Homeserver.Domain)))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Password:"), theme.DimTextStyle.Render(mxUserPassword))
		fmt.Println()
		return nil
	},
}

var matrixUsersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX USERS"))
		fmt.Println()
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		users, err := admin.ListUsers(0, 100)
		if err != nil {
			return err
		}
		fmt.Printf("  Total: %s\n\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", users.Total)))
		for _, u := range users.Users {
			badge := ""
			if u.Admin == 1 {
				badge = theme.WarningStyle.Render(" [admin]")
			}
			if u.Deactivated == 1 {
				badge += theme.ErrorStyle.Render(" (deactivated)")
			}
			created := ""
			if u.CreationTS > 0 {
				created = time.Unix(u.CreationTS/1000, 0).Format("2006-01-02")
			}
			fmt.Printf("  %s %-30s %s%s  %s\n", theme.SymbolStar, theme.HighlightStyle.Render(u.Name), theme.DimTextStyle.Render(u.DisplayName), badge, theme.DimTextStyle.Render(created))
		}
		fmt.Println()
		return nil
	},
}

var matrixUsersRemoveCmd = &cobra.Command{
	Use:     "remove <username>",
	Aliases: []string{"rm"},
	Short:   "Deactivate a user",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		uid := args[0]
		if uid[0] != '@' {
			uid = fmt.Sprintf("@%s:%s", uid, cfg.Homeserver.Domain)
		}
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		if err := admin.DeactivateUser(uid); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Deactivated "+uid))
		return nil
	},
}

var matrixUsersAdminCmd = &cobra.Command{
	Use:   "admin <username>",
	Short: "Grant/revoke admin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		uid := args[0]
		if uid[0] != '@' {
			uid = fmt.Sprintf("@%s:%s", uid, cfg.Homeserver.Domain)
		}
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		if err := admin.SetAdmin(uid, !mxRevokeAdmin); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Done"))
		return nil
	},
}

var matrixUsersResetCmd = &cobra.Command{
	Use:   "reset-password <username>",
	Short: "Reset a user's password",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		uid := args[0]
		if uid[0] != '@' {
			uid = fmt.Sprintf("@%s:%s", uid, cfg.Homeserver.Domain)
		}
		if mxUserPassword == "" {
			mxUserPassword = matrixGeneratePassword(16)
		}
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		if err := admin.ResetPassword(uid, mxUserPassword); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Password reset"))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("New Password:"), theme.DimTextStyle.Render(mxUserPassword))
		return nil
	},
}

func init() {
	matrixUsersAddCmd.Flags().StringVarP(&mxUserPassword, "password", "p", "", "Password (generated if empty)")
	matrixUsersAddCmd.Flags().StringVarP(&mxUserDisplayName, "display-name", "n", "", "Display name")
	matrixUsersAddCmd.Flags().BoolVarP(&mxUserAdmin, "admin", "a", false, "Make admin")
	matrixUsersAdminCmd.Flags().BoolVar(&mxRevokeAdmin, "revoke", false, "Revoke admin")
	matrixUsersResetCmd.Flags().StringVarP(&mxUserPassword, "password", "p", "", "New password")

	matrixUsersCmd.AddCommand(matrixUsersAddCmd)
	matrixUsersCmd.AddCommand(matrixUsersListCmd)
	matrixUsersCmd.AddCommand(matrixUsersRemoveCmd)
	matrixUsersCmd.AddCommand(matrixUsersAdminCmd)
	matrixUsersCmd.AddCommand(matrixUsersResetCmd)
	matrixCmd.AddCommand(matrixUsersCmd)
}
