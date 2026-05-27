package cmd

import (
	"fmt"
	"time"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
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

		label := channelID
		if ch, err := client.GetChannel(channelID); err == nil {
			label = "#" + ch.DisplayName
		}
		t.Section(label)

		pl, err := client.GetChannelPostsPage(channelID, 0, mxHistoryLimit)
		if err != nil {
			return fmt.Errorf("failed to fetch messages: %w", err)
		}

		if len(pl.OrderArr) == 0 {
			t.Info("no messages")
			fmt.Println()
			return nil
		}

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

		for i := len(pl.OrderArr) - 1; i >= 0; i-- {
			p := pl.Posts[pl.OrderArr[i]]
			if p.Type != "" || p.Message == "" {
				continue
			}
			ts := time.Unix(p.CreateAt/1000, 0).Format("2006-01-02 15:04:05")
			username := resolveUser(p.UserID)
			fmt.Printf("  %s  %-22s  %s\n",
				t.Dim(ts),
				t.Bold(t.Gold.S("@"+username)),
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
