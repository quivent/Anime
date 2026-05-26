package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxRoomTopic  string
	mxRoomInvite []string
	mxRoomDirect bool
)

var matrixRoomsCmd = &cobra.Command{
	Use:     "rooms",
	Aliases: []string{"room", "r"},
	Short:   "Manage Matrix rooms",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixRoomsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new room",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, _ := matrixcfg.Load()
		invite := make([]string, len(mxRoomInvite))
		for i, u := range mxRoomInvite {
			if !strings.HasPrefix(u, "@") {
				invite[i] = fmt.Sprintf("@%s:%s", u, cfg.Homeserver.Domain)
			} else {
				invite[i] = u
			}
		}
		client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
		roomID, err := client.CreateRoom(name, mxRoomTopic, invite, mxRoomDirect)
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Room created"))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Room ID:"), theme.DimTextStyle.Render(roomID))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Name:"), theme.DimTextStyle.Render(name))
		fmt.Println()
		return nil
	},
}

var matrixRoomsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all rooms",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX ROOMS"))
		fmt.Println()
		admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
		rooms, err := admin.ListRooms(0, 100)
		if err != nil {
			return err
		}
		fmt.Printf("  Total: %s\n\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", rooms.Total)))
		for _, r := range rooms.Rooms {
			name := r.Name
			if name == "" {
				name = "(unnamed)"
			}
			fmt.Printf("  %s %-24s %s\n", theme.SymbolStar, theme.HighlightStyle.Render(name), theme.DimTextStyle.Render(fmt.Sprintf("%d members", r.NumMembers)))
			if r.Topic != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(r.Topic))
			}
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(r.RoomID))
		}
		fmt.Println()
		return nil
	},
}

var matrixRoomsJoinCmd = &cobra.Command{
	Use:   "join <room-id>",
	Short: "Join a room",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
		roomID, err := client.JoinRoom(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Joined"), theme.DimTextStyle.Render(roomID))
		return nil
	},
}

var matrixRoomsLeaveCmd = &cobra.Command{
	Use:   "leave <room-id>",
	Short: "Leave a room",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
		return client.LeaveRoom(args[0])
	},
}

var matrixRoomsInviteCmd = &cobra.Command{
	Use:   "invite <room-id> <user>",
	Short: "Invite a user to a room",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		uid := args[1]
		if !strings.HasPrefix(uid, "@") {
			uid = fmt.Sprintf("@%s:%s", uid, cfg.Homeserver.Domain)
		}
		client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
		if err := client.InviteUser(args[0], uid); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Invited "+uid))
		return nil
	},
}

func init() {
	matrixRoomsCreateCmd.Flags().StringVarP(&mxRoomTopic, "topic", "t", "", "Room topic")
	matrixRoomsCreateCmd.Flags().StringSliceVarP(&mxRoomInvite, "invite", "i", nil, "Users to invite")
	matrixRoomsCreateCmd.Flags().BoolVar(&mxRoomDirect, "direct", false, "Direct message room")

	matrixRoomsCmd.AddCommand(matrixRoomsCreateCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsListCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsJoinCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsLeaveCmd)
	matrixRoomsCmd.AddCommand(matrixRoomsInviteCmd)
	matrixCmd.AddCommand(matrixRoomsCmd)
}
