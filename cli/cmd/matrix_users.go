package cmd

import (
	"fmt"
	"time"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var (
	mxUserPassword string
	mxUserEmail    string
	mxUserAdmin    bool
	mxRevokeAdmin  bool
	mxUserSearch   string
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
			email = username + "@chat.local"
		}
		t.Info("creating @" + username + "…")
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		u, err := client.CreateUser(username, email, mxUserPassword)
		if err != nil {
			return err
		}
		if mxUserAdmin {
			_ = client.SetAdmin(u.ID, true)
		}
		if cfg.Server.TeamID != "" {
			_ = client.AddTeamMember(cfg.Server.TeamID, u.ID)
		}
		t.Ok("created")
		t.KV("username", "@"+u.Username)
		t.KV("email", u.Email)
		t.KV("password", mxUserPassword)
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
		t.Section("USERS")
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		if mxUserSearch != "" {
			found, err := client.SearchUsers(mxUserSearch, 100)
			if err != nil {
				return err
			}
			fmt.Printf("  %s results\n\n", t.Bold(t.Gold.S(fmt.Sprintf("%d", len(found)))))
			tbl := t.NewTable("username", "email", "role", "created")
			for _, u := range found {
				role := ""
				if u.IsAdmin() {
					role = t.Gold.S("admin")
				}
				if u.IsDeleted() {
					if role != "" {
						role += " "
					}
					role += t.Loss.S("deactivated")
				}
				created := ""
				if u.CreateAt > 0 {
					created = t.Dim(time.Unix(u.CreateAt/1000, 0).Format("2006-01-02"))
				}
				tbl.Row(t.Bold(t.Gold.S("@"+u.Username)), t.Dim(u.Email), role, created)
			}
			fmt.Print(tbl.Render())
		} else {
			all, err := client.ListUsers(0, 200)
			if err != nil {
				return err
			}
			fmt.Printf("  %s total\n\n", t.Bold(t.Gold.S(fmt.Sprintf("%d", len(all)))))
			tbl := t.NewTable("username", "email", "role", "created")
			for _, u := range all {
				role := ""
				if u.IsAdmin() {
					role = t.Gold.S("admin")
				}
				if u.IsDeleted() {
					if role != "" {
						role += " "
					}
					role += t.Loss.S("deactivated")
				}
				created := ""
				if u.CreateAt > 0 {
					created = t.Dim(time.Unix(u.CreateAt/1000, 0).Format("2006-01-02"))
				}
				tbl.Row(t.Bold(t.Gold.S("@"+u.Username)), t.Dim(u.Email), role, created)
			}
			fmt.Print(tbl.Render())
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
		t.Ok("deactivated @" + u.Username)
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
		if grant {
			t.Ok("granted admin to @" + u.Username)
		} else {
			t.Ok("revoked admin from @" + u.Username)
		}
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
		t.Ok("password reset for @" + u.Username)
		t.KV("new password", mxUserPassword)
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
