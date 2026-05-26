package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var mxWatchAllRooms bool

var matrixWatchCmd = &cobra.Command{
	Use:   "watch [room-id]",
	Short: "Tail live messages from a room (or all rooms)",
	Long: `Stream messages in real-time using Matrix /sync long-polling.
Press Ctrl-C to stop.`,
	Example: `  anime matrix watch '!abc:localhost'
  anime matrix watch --all
  anime matrix watch '#general:localhost'`,
	RunE: runMatrixWatch,
}

func init() {
	matrixWatchCmd.Flags().BoolVar(&mxWatchAllRooms, "all", false, "Watch all joined rooms")
	matrixCmd.AddCommand(matrixWatchCmd)
}

func runMatrixWatch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !mxWatchAllRooms {
		return fmt.Errorf("specify a room ID or use --all")
	}

	cfg, _ := matrixcfg.Load()
	client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)

	filterRoom := ""
	if len(args) > 0 {
		filterRoom = args[0]
		// Resolve alias if needed
		if strings.HasPrefix(filterRoom, "#") {
			resolved, err := client.ResolveAlias(filterRoom)
			if err != nil {
				return fmt.Errorf("cannot resolve alias %s: %w", filterRoom, err)
			}
			filterRoom = resolved
		}
	}

	fmt.Println()
	if filterRoom != "" {
		fmt.Printf("  %s %s %s\n", theme.SymbolBolt, theme.InfoStyle.Render("Watching"), theme.HighlightStyle.Render(filterRoom))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolBolt, theme.InfoStyle.Render("Watching all rooms"))
	}
	fmt.Println(theme.DimTextStyle.Render("  Ctrl-C to stop"))
	fmt.Println()

	// Initial sync to get the since token (skip old messages)
	syncResp, err := client.Sync("", 0)
	if err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}
	since := syncResp.NextBatch

	// Handle Ctrl-C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	for {
		select {
		case <-sigCh:
			fmt.Println()
			fmt.Printf("  %s\n", theme.DimTextStyle.Render("Stopped"))
			return nil
		default:
		}

		resp, err := client.Sync(since, 30000)
		if err != nil {
			fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("sync error: "+err.Error()))
			time.Sleep(2 * time.Second)
			continue
		}
		since = resp.NextBatch

		for roomID, room := range resp.Rooms.Join {
			if filterRoom != "" && roomID != filterRoom {
				continue
			}
			for _, ev := range room.Timeline.Events {
				if ev.Type != "m.room.message" {
					continue
				}
				body, _ := ev.Content["body"].(string)
				if body == "" {
					continue
				}
				ts := time.UnixMilli(ev.OriginTS).Format("15:04:05")
				sender := ev.Sender
				// Shorten sender for display
				if idx := strings.Index(sender, ":"); idx > 0 {
					sender = sender[:idx]
				}

				roomLabel := ""
				if mxWatchAllRooms {
					short := roomID
					if len(short) > 12 {
						short = short[:12] + "..."
					}
					roomLabel = theme.DimTextStyle.Render("["+short+"] ")
				}

				fmt.Printf("  %s %s%s  %s\n",
					theme.DimTextStyle.Render(ts),
					roomLabel,
					theme.HighlightStyle.Render(sender),
					body)
			}
		}
	}
}
