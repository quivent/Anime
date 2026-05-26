package cmd

import (
	"fmt"
	"time"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var mxHistoryLimit int

var matrixHistoryCmd = &cobra.Command{
	Use:     "history <channel-id>",
	Aliases: []string{"hist", "messages", "msgs"},
	Short:   "Read message history from a channel",
	Example: `  anime matrix history <channel-id>
  anime matrix history <channel-id> --limit 50`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		channelID := args[0]

		fmt.Println()
		if ch, err := client.GetChannel(channelID); err == nil {
			fmt.Printf("  %s %s\n", theme.InfoStyle.Render("History:"),
				theme.HighlightStyle.Render("#"+ch.DisplayName))
		} else {
			fmt.Printf("  %s %s\n", theme.InfoStyle.Render("History:"),
				theme.HighlightStyle.Render(channelID))
		}
		fmt.Println()

		pl, err := client.GetChannelPostsPage(channelID, 0, mxHistoryLimit)
		if err != nil {
			return fmt.Errorf("failed to fetch messages: %w", err)
		}

		if len(pl.OrderArr) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  (no messages)"))
			fmt.Println()
			return nil
		}

		// Cache user names
		userNames := map[string]string{}
		resolveUser := func(userID string) string {
			if name, ok := userNames[userID]; ok {
				return name
			}
			if u, err := client.GetUser(userID); err == nil {
				userNames[userID] = u.Username
				return u.Username
			}
			return userID
		}

		// Display chronologically (reverse the newest-first order array)
		for i := len(pl.OrderArr) - 1; i >= 0; i-- {
			p := pl.Posts[pl.OrderArr[i]]
			if p.Type != "" { // skip system messages
				continue
			}
			if p.Message == "" {
				continue
			}
			ts := time.Unix(p.CreateAt/1000, 0).Format("2006-01-02 15:04:05")
			username := resolveUser(p.UserID)

			fmt.Printf("  %s  %s  %s\n",
				theme.DimTextStyle.Render(ts),
				theme.HighlightStyle.Render(fmt.Sprintf("%-20s", "@"+username)),
				p.Message)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	matrixHistoryCmd.Flags().IntVarP(&mxHistoryLimit, "limit", "n", 25, "Number of messages to fetch")
	matrixCmd.AddCommand(matrixHistoryCmd)
}
