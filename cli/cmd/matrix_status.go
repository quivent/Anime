package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/synapse"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Check server health and running services",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX STATUS"))
		fmt.Println()

		cfg, _ := matrixcfg.Load()

		fmt.Println(matrixSeparator())
		fmt.Println(theme.InfoStyle.Render("  Homeserver"))
		fmt.Println(matrixSeparator())
		fmt.Println()
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("URL:"), theme.DimTextStyle.Render(cfg.Homeserver.URL))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Domain:"), theme.DimTextStyle.Render(cfg.Homeserver.Domain))
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.DimTextStyle.Render(cfg.Homeserver.AdminUser))

		if synapse.IsHealthy(cfg.Homeserver.URL) {
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Status:"), theme.SuccessStyle.Render("ONLINE"))
			client := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
			if ver, err := client.ServerVersion(); err == nil {
				fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Version:"), theme.DimTextStyle.Render(ver))
			}
			if cfg.Homeserver.AdminToken != "" {
				admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
				if users, err := admin.ListUsers(0, 1); err == nil {
					fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Users:"), theme.DimTextStyle.Render(fmt.Sprintf("%d", users.Total)))
				}
				if rooms, err := admin.ListRooms(0, 1); err == nil {
					fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Rooms:"), theme.DimTextStyle.Render(fmt.Sprintf("%d", rooms.Total)))
				}
			}
		} else {
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Status:"), theme.ErrorStyle.Render("OFFLINE"))
		}
		fmt.Println()

		if len(cfg.Agents) > 0 {
			fmt.Println(matrixSeparator())
			fmt.Println(theme.InfoStyle.Render("  Agents"))
			fmt.Println(matrixSeparator())
			fmt.Println()
			for _, a := range cfg.Agents {
				alive := matrixIsAlive(a.PID)
				st := a.Status
				if !alive && st == "running" {
					st = "dead"
				}
				stStr := theme.SuccessStyle.Render(st)
				if st != "running" {
					stStr = theme.ErrorStyle.Render(st)
				}
				fmt.Printf("  %s %-16s %s  %s\n", theme.SymbolStar, theme.HighlightStyle.Render(a.Name), stStr, theme.DimTextStyle.Render(a.UserID))
			}
			fmt.Println()
		}

		if len(cfg.Daemons) > 0 {
			fmt.Println(matrixSeparator())
			fmt.Println(theme.InfoStyle.Render("  Daemons"))
			fmt.Println(matrixSeparator())
			fmt.Println()
			for _, d := range cfg.Daemons {
				alive := matrixIsAlive(d.PID)
				st := d.Status
				if !alive && st == "running" {
					st = "dead"
				}
				stStr := theme.SuccessStyle.Render(st)
				if st != "running" {
					stStr = theme.ErrorStyle.Render(st)
				}
				fmt.Printf("  %s %-16s %s  PID %d\n", theme.SymbolBolt, theme.HighlightStyle.Render(d.Name), stStr, d.PID)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	matrixCmd.AddCommand(matrixStatusCmd)
}
