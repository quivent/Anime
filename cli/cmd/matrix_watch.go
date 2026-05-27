package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var mxWatchAll bool

var matrixWatchCmd = &cobra.Command{
	Use:   "watch <channel-id>",
	Short: "Live-tail messages from a channel",
	Example: `  anime matrix watch <channel-id>
  anime matrix watch --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		var channelIDs []string
		if mxWatchAll {
			if cfg.Server.TeamID == "" {
				return fmt.Errorf("no team configured — run anime matrix connect first")
			}
			me, err := client.GetMe()
			if err != nil {
				return err
			}
			channels, err := client.GetUserChannels(cfg.Server.TeamID, me.ID)
			if err != nil {
				return err
			}
			for _, ch := range channels {
				if ch.Type != "D" {
					channelIDs = append(channelIDs, ch.ID)
				}
			}
			t.Info(fmt.Sprintf("watching %s", t.Bold(t.Gold.S(fmt.Sprintf("%d channels", len(channelIDs))))))
		} else {
			if len(args) == 0 {
				return fmt.Errorf("specify a channel ID or use --all")
			}
			channelIDs = []string{args[0]}
			if ch, err := client.GetChannel(args[0]); err == nil {
				t.Info("watching " + t.Bold(t.Gold.S("#"+ch.DisplayName)))
			}
		}

		if len(channelIDs) == 0 {
			return fmt.Errorf("no channels to watch")
		}

		chanNames := map[string]string{}
		for _, id := range channelIDs {
			if ch, err := client.GetChannel(id); err == nil {
				chanNames[id] = ch.DisplayName
			} else {
				chanNames[id] = id
			}
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

		fmt.Println("  " + t.Dim("Ctrl-C to stop"))
		fmt.Println()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		since := time.Now().UnixMilli()

		for {
			select {
			case <-sig:
				fmt.Println()
				t.Info("stopped")
				return nil
			default:
			}

			for _, chID := range channelIDs {
				pl, err := client.GetChannelPosts(chID, since, 50)
				if err != nil {
					continue
				}
				for i := len(pl.OrderArr) - 1; i >= 0; i-- {
					p := pl.Posts[pl.OrderArr[i]]
					if p.Type != "" || p.CreateAt <= since {
						continue
					}
					chanName := chanNames[p.ChannelID]
					username := resolveUser(p.UserID)
					ts := time.Unix(p.CreateAt/1000, 0).Format("15:04:05")
					msg := strings.ReplaceAll(p.Message, "\n", " ")
					if len(msg) > 120 {
						msg = msg[:117] + "…"
					}
					fmt.Printf("  %s %s %s: %s\n",
						t.Dim("["+ts+"]"),
						t.Bold(t.Gold.S("#"+chanName)),
						t.Cyan.S("@"+username),
						msg)
				}
			}

			since = time.Now().UnixMilli()
			time.Sleep(2 * time.Second)
		}
	},
}

func init() {
	matrixWatchCmd.Flags().BoolVar(&mxWatchAll, "all", false, "Watch all joined channels")
	matrixCmd.AddCommand(matrixWatchCmd)
}
