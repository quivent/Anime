package cmd

import (
	"fmt"
	"strings"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var (
	mxChannelPurpose string
	mxChannelPrivate bool
	mxChannelInvite  []string
)

var matrixRoomsCmd = &cobra.Command{
	Use:     "channels",
	Aliases: []string{"channel", "rooms", "room", "r"},
	Short:   "Manage Mattermost channels",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixRoomsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		teamID := cfg.Server.TeamID
		if teamID == "" {
			teams, err := client.GetTeams(0, 1)
			if err != nil || len(teams) == 0 {
				return fmt.Errorf("no team configured — run anime matrix connect first")
			}
			teamID = teams[0].ID
		}

		channelType := "O"
		if mxChannelPrivate {
			channelType = "P"
		}
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		ch, err := client.CreateChannel(teamID, slug, name, channelType, mxChannelPurpose)
		if err != nil {
			return err
		}

		for _, username := range mxChannelInvite {
			u, err := client.GetUserByUsername(strings.TrimPrefix(username, "@"))
			if err != nil {
				t.Warn("user not found: " + username)
				continue
			}
			if err := client.AddChannelMember(ch.ID, u.ID); err != nil {
				t.Warn("invite " + username + ": " + err.Error())
			} else {
				t.Ok("invited @" + u.Username)
			}
		}

		t.Ok("channel created")
		t.KV("id", ch.ID)
		t.KV("name", ch.DisplayName)
		visibility := "public"
		if mxChannelPrivate {
			visibility = "private"
		}
		t.KV("type", visibility)
		fmt.Println()
		return nil
	},
}

var matrixRoomsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		t.Section("CHANNELS")
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		teamID := cfg.Server.TeamID
		if teamID == "" {
			teams, err := client.GetTeams(0, 1)
			if err != nil || len(teams) == 0 {
				return fmt.Errorf("no team configured — run anime matrix connect first")
			}
			teamID = teams[0].ID
		}

		channels, err := client.GetTeamChannels(teamID, 0, 200)
		if err != nil {
			return err
		}

		fmt.Printf("  %s channels\n\n", t.Bold(t.Gold.S(fmt.Sprintf("%d", len(channels)))))
		tbl := t.NewTable("name", "id", "members", "type")
		for _, ch := range channels {
			vis := t.Dim("public")
			switch ch.Type {
			case "P":
				vis = t.Dim("private")
			case "D":
				vis = t.Dim("direct")
			}
			tbl.Row(
				t.Bold(t.Gold.S(ch.DisplayName)),
				t.Dim(ch.ID),
				fmt.Sprintf("%d", ch.MemberCount),
				vis,
			)
		}
		fmt.Print(tbl.Render())
		fmt.Println()
		return nil
	},
}

var matrixRoomsJoinCmd = &cobra.Command{
	Use:   "join <channel-id>",
	Short: "Join a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		me, err := client.GetMe()
		if err != nil {
			return err
		}
		if err := client.AddChannelMember(args[0], me.ID); err != nil {
			return err
		}
		t.Ok("joined " + t.Dim(args[0]))
		return nil
	},
}

var matrixRoomsLeaveCmd = &cobra.Command{
	Use:   "leave <channel-id>",
	Short: "Leave a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		me, err := client.GetMe()
		if err != nil {
			return err
		}
		if err := client.RemoveChannelMember(args[0], me.ID); err != nil {
			return err
		}
		t.Ok("left " + t.Dim(args[0]))
		return nil
	},
}

var matrixRoomsInviteCmd = &cobra.Command{
	Use:   "invite <channel-id> <username>",
	Short: "Invite a user to a channel",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)
		username := strings.TrimPrefix(args[1], "@")
		u, err := client.GetUserByUsername(username)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if err := client.AddChannelMember(args[0], u.ID); err != nil {
			return err
		}
		t.Ok("invited @" + u.Username)
		return nil
	},
}

func init() {
	matrixRoomsCreateCmd.Flags().StringVarP(&mxChannelPurpose, "purpose", "t", "", "Channel purpose/description")
	matrixRoomsCreateCmd.Flags().BoolVar(&mxChannelPrivate, "private", false, "Create a private channel")
	matrixRoomsCreateCmd.Flags().StringSliceVarP(&mxChannelInvite, "invite", "i", nil, "Users to invite")

	matrixRoomsCmd.AddCommand(matrixRoomsCreateCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsListCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsJoinCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsLeaveCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsInviteCmd)
	matrixCmd.AddCommand(matrixRoomsCmd)
}
