package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show Mattermost server health and stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()

		fmt.Println()
		fmt.Println(theme.RenderBanner("STATUS"))
		fmt.Println()

		if cfg.Server.URL == "" {
			fmt.Println(theme.WarningStyle.Render("  Not connected. Run: anime matrix connect"))
			fmt.Println()
			return nil
		}

		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Server:"), theme.InfoStyle.Render(cfg.Server.URL))
		if cfg.Server.Username != "" {
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("User:"), theme.DimTextStyle.Render("@"+cfg.Server.Username))
		}
		if cfg.Server.TeamName != "" {
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Team:"), theme.DimTextStyle.Render(cfg.Server.TeamName))
		}
		fmt.Println()

		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		// Health check
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Checking server..."))
		ver, err := client.ServerVersion()
		if err != nil {
			fmt.Printf("  %s %s\n", theme.SymbolError, theme.ErrorStyle.Render("Unreachable: "+err.Error()))
			fmt.Println()
			return nil
		}
		fmt.Printf("  %s %s %s\n", theme.SymbolSuccess,
			theme.SuccessStyle.Render("Online"), theme.DimTextStyle.Render("v"+ver))

		// Users
		users, err := client.ListUsers(0, 200)
		if err == nil {
			active := 0
			for _, u := range users {
				if !u.IsDeleted() {
					active++
				}
			}
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Users:"),
				theme.DimTextStyle.Render(fmt.Sprintf("%d active", active)))
		}

		// Teams
		teams, err := client.GetTeams(0, 50)
		if err == nil {
			names := make([]string, len(teams))
			for i, t := range teams {
				names[i] = t.DisplayName
			}
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Teams:"),
				theme.DimTextStyle.Render(strings.Join(names, ", ")))
		}

		// Channels
		if cfg.Server.TeamID != "" {
			channels, err := client.GetTeamChannels(cfg.Server.TeamID, 0, 200)
			if err == nil {
				fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Channels:"),
					theme.DimTextStyle.Render(fmt.Sprintf("%d", len(channels))))
			}
		}

		// Agents
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
				fmt.Printf("  %s %-20s %s  PID %d  %s\n",
					theme.SymbolBolt,
					theme.HighlightStyle.Render(a.Name),
					stStr, a.PID,
					theme.DimTextStyle.Render(a.Model))
			}
			fmt.Println()
		}

		// Daemons
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
				fmt.Printf("  %s %-20s %s  PID %d\n",
					theme.SymbolBolt, theme.HighlightStyle.Render(d.Name), stStr, d.PID)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	matrixCmd.AddCommand(matrixStatusCmd)
}
