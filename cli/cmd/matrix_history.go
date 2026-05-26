package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var mxHistoryLimit int

var matrixHistoryCmd = &cobra.Command{
	Use:     "history <room-id>",
	Aliases: []string{"hist", "messages", "msgs"},
	Short:   "Read message history from a room",
	Example: `  anime matrix history '!abc:localhost'
  anime matrix history '#general:localhost' --limit 50
  anime matrix history '!abc:localhost' -n 10`,
	Args: cobra.ExactArgs(1),
	RunE: runMatrixHistory,
}

func init() {
	matrixHistoryCmd.Flags().IntVarP(&mxHistoryLimit, "limit", "n", 25, "Number of messages to fetch")
	matrixCmd.AddCommand(matrixHistoryCmd)
}

func runMatrixHistory(cmd *cobra.Command, args []string) error {
	cfg, _ := matrixcfg.Load()
	client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)

	roomID := args[0]
	// Resolve alias
	if strings.HasPrefix(roomID, "#") {
		resolved, err := client.ResolveAlias(roomID)
		if err != nil {
			return fmt.Errorf("cannot resolve %s: %w", roomID, err)
		}
		roomID = resolved
	}

	fmt.Println()
	fmt.Printf("  %s %s %s\n", theme.SymbolStar, theme.InfoStyle.Render("History for"), theme.HighlightStyle.Render(roomID))
	fmt.Println()

	resp, err := client.RoomMessages(roomID, "", "b", mxHistoryLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	if len(resp.Chunk) == 0 {
		fmt.Println(theme.DimTextStyle.Render("  (no messages)"))
		fmt.Println()
		return nil
	}

	// Messages come in reverse order (newest first) with dir=b, so reverse
	msgs := resp.Chunk
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	for _, ev := range msgs {
		if ev.Type != "m.room.message" {
			continue
		}
		body, _ := ev.Content["body"].(string)
		msgtype, _ := ev.Content["msgtype"].(string)
		if body == "" {
			continue
		}

		ts := time.UnixMilli(ev.OriginTS).Format("2006-01-02 15:04:05")
		sender := ev.Sender
		if idx := strings.Index(sender, ":"); idx > 0 {
			sender = sender[:idx]
		}

		prefix := ""
		if msgtype == "m.image" {
			prefix = "[image] "
		} else if msgtype == "m.file" {
			prefix = "[file] "
		}

		fmt.Printf("  %s  %s  %s%s\n",
			theme.DimTextStyle.Render(ts),
			theme.HighlightStyle.Render(fmt.Sprintf("%-16s", sender)),
			prefix,
			body)
	}
	fmt.Println()
	return nil
}
