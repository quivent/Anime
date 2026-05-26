package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
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

		// Sanitize name: lowercase, no spaces
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

		ch, err := client.CreateChannel(teamID, slug, name, channelType, mxChannelPurpose)
		if err != nil {
			return err
		}

		// Invite users
		for _, username := range mxChannelInvite {
			u, err := client.GetUserByUsername(strings.TrimPrefix(username, "@"))
			if err != nil {
				fmt.Printf("  %s %s\n", theme.SymbolWarning,
					theme.WarningStyle.Render("User not found: "+username))
				continue
			}
			if err := client.AddChannelMember(ch.ID, u.ID); err != nil {
				fmt.Printf("  %s %s\n", theme.SymbolWarning,
					theme.WarningStyle.Render("Invite "+username+": "+err.Error()))
			} else {
				fmt.Printf("  %s %s\n", theme.SymbolSuccess,
					theme.SuccessStyle.Render("Invited "+username))
			}
		}

		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Channel created"))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("ID:"), theme.DimTextStyle.Render(ch.ID))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Name:"), theme.DimTextStyle.Render(ch.DisplayName))
		visibility := "public"
		if mxChannelPrivate {
			visibility = "private"
		}
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Type:"), theme.DimTextStyle.Render(visibility))
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
		fmt.Println()
		fmt.Println(theme.RenderBanner("CHANNELS"))
		fmt.Println()
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
		fmt.Printf("  Total: %s\n\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(channels))))
		for _, ch := range channels {
			typeStr := ""
			switch ch.Type {
			case "P":
				typeStr = theme.DimTextStyle.Render("[private]")
			case "D":
				typeStr = theme.DimTextStyle.Render("[direct]")
			}
			fmt.Printf("  %s %-24s %s %s\n",
				theme.SymbolStar,
				theme.HighlightStyle.Render(ch.DisplayName),
				theme.DimTextStyle.Render(fmt.Sprintf("%d members", ch.MemberCount)),
				typeStr)
			if ch.Purpose != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(ch.Purpose))
			}
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(ch.ID))
		}
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
		fmt.Printf("  %s %s %s\n", theme.SymbolSuccess,
			theme.SuccessStyle.Render("Joined"), theme.DimTextStyle.Render(args[0]))
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
		fmt.Printf("  %s %s %s\n", theme.SymbolSuccess,
			theme.SuccessStyle.Render("Left"), theme.DimTextStyle.Render(args[0]))
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
		fmt.Printf("  %s %s\n", theme.SymbolSuccess,
			theme.SuccessStyle.Render("Invited @"+u.Username))
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
