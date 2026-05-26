package cmd

import (
	"fmt"
	"time"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxUserPassword    string
	mxUserEmail       string
	mxUserAdmin       bool
	mxRevokeAdmin     bool
	mxUserSearch      string
)

var matrixUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "u"},
	Short:   "Manage Mattermost user accounts",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixUsersAddCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Create a new user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		cfg, _ := mmcfg.Load()
		if mxUserPassword == "" {
			mxUserPassword = matrixGeneratePassword(16)
		}
		email := mxUserEmail
		if email == "" {
			email = username + "@" + "chat.local"
		}
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Creating @"+username+"..."))
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		u, err := client.CreateUser(username, email, mxUserPassword)
		if err != nil {
			return err
		}
		if mxUserAdmin {
			_ = client.SetAdmin(u.ID, true)
		}
		// Add to default team
		if cfg.Server.TeamID != "" {
			_ = client.AddTeamMember(cfg.Server.TeamID, u.ID)
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Created"))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Username:"), theme.DimTextStyle.Render("@"+u.Username))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Email:"), theme.DimTextStyle.Render(u.Email))
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
		cfg, _ := mmcfg.Load()
		fmt.Println()
		fmt.Println(theme.RenderBanner("USERS"))
		fmt.Println()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		if mxUserSearch != "" {
			found, err := client.SearchUsers(mxUserSearch, 100)
			if err != nil {
				return err
			}
			fmt.Printf("  Results: %s\n\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(found))))
			for _, u := range found {
				badge := ""
				if u.IsAdmin() {
					badge = theme.WarningStyle.Render(" [admin]")
				}
				if u.IsDeleted() {
					badge += theme.ErrorStyle.Render(" (deactivated)")
				}
				created := ""
				if u.CreateAt > 0 {
					created = time.Unix(u.CreateAt/1000, 0).Format("2006-01-02")
				}
				fmt.Printf("  %s %-28s %s%s  %s\n",
					theme.SymbolStar,
					theme.HighlightStyle.Render("@"+u.Username),
					theme.DimTextStyle.Render(u.Email),
					badge,
					theme.DimTextStyle.Render(created))
			}
		} else {
			all, err := client.ListUsers(0, 200)
			if err != nil {
				return err
			}
			fmt.Printf("  Total: %s\n\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(all))))
			for _, u := range all {
				badge := ""
				if u.IsAdmin() {
					badge = theme.WarningStyle.Render(" [admin]")
				}
				if u.IsDeleted() {
					badge += theme.ErrorStyle.Render(" (deactivated)")
				}
				created := ""
				if u.CreateAt > 0 {
					created = time.Unix(u.CreateAt/1000, 0).Format("2006-01-02")
				}
				fmt.Printf("  %s %-28s %s%s  %s\n",
					theme.SymbolStar,
					theme.HighlightStyle.Render("@"+u.Username),
					theme.DimTextStyle.Render(u.Email),
					badge,
					theme.DimTextStyle.Render(created))
			}
		}
		fmt.Println()
		return nil
	},
}

var matrixUsersRemoveCmd = &cobra.Command{
	Use:     "remove <username>",
	Aliases: []string{"rm", "deactivate"},
	Short:   "Deactivate a user",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		u, err := client.GetUserByUsername(args[0])
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if err := client.DeactivateUser(u.ID); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Deactivated @"+u.Username))
		return nil
	},
}

var matrixUsersAdminCmd = &cobra.Command{
	Use:   "admin <username>",
	Short: "Grant or revoke admin privileges",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		u, err := client.GetUserByUsername(args[0])
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		grant := !mxRevokeAdmin
		if err := client.SetAdmin(u.ID, grant); err != nil {
			return err
		}
		verb := "Granted admin to"
		if !grant {
			verb = "Revoked admin from"
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render(verb+" @"+u.Username))
		return nil
	},
}

var matrixUsersResetCmd = &cobra.Command{
	Use:   "reset-password <username>",
	Short: "Reset a user's password",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		if mxUserPassword == "" {
			mxUserPassword = matrixGeneratePassword(16)
		}
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		u, err := client.GetUserByUsername(args[0])
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if err := client.ResetPassword(u.ID, mxUserPassword); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Password reset for @"+u.Username))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("New Password:"), theme.DimTextStyle.Render(mxUserPassword))
		return nil
	},
}

func init() {
	matrixUsersAddCmd.Flags().StringVarP(&mxUserPassword, "password", "p", "", "Password (generated if empty)")
	matrixUsersAddCmd.Flags().StringVarP(&mxUserEmail, "email", "e", "", "Email address")
	matrixUsersAddCmd.Flags().BoolVarP(&mxUserAdmin, "admin", "a", false, "Make admin")
	matrixUsersListCmd.Flags().StringVarP(&mxUserSearch, "search", "s", "", "Search users by name or email")
	matrixUsersAdminCmd.Flags().BoolVar(&mxRevokeAdmin, "revoke", false, "Revoke admin")
	matrixUsersResetCmd.Flags().StringVarP(&mxUserPassword, "password", "p", "", "New password (generated if empty)")

	matrixUsersCmd.AddCommand(matrixUsersAddCmd)
	matrixUsersCmd.AddCommand(matrixUsersListCmd)
	matrixUsersCmd.AddCommand(matrixUsersRemoveCmd)
	matrixUsersCmd.AddCommand(matrixUsersAdminCmd)
	matrixUsersCmd.AddCommand(matrixUsersResetCmd)
	matrixCmd.AddCommand(matrixUsersCmd)
}
