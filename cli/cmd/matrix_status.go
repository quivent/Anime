package cmd

import (
	"fmt"
	"strings"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var matrixStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show Mattermost server health and stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()

		t.Section("STATUS")

		if cfg.Server.URL == "" {
			t.Warn("not connected — run: anime matrix connect")
			fmt.Println()
			return nil
		}

		t.KV("server", t.Cyan.S(cfg.Server.URL))
		if cfg.Server.Username != "" {
			t.KV("user", "@"+cfg.Server.Username)
		}
		if cfg.Server.TeamName != "" {
			t.KV("team", cfg.Server.TeamName)
		}
		fmt.Println()

		client := mmClient(cfg.Server.URL, cfg.Server.Token)

		t.Info("checking server…")
		ver, err := client.ServerVersion()
		if err != nil {
			t.Fail("unreachable: " + err.Error())
			fmt.Println()
			return nil
		}
		t.Ok("online  " + t.Dim("v"+ver))

		if users, err := client.ListUsers(0, 200); err == nil {
			active := 0
			for _, u := range users {
				if !u.IsDeleted() {
					active++
				}
			}
			t.KV("users", fmt.Sprintf("%d active", active))
		}

		if teams, err := client.GetTeams(0, 50); err == nil {
			names := make([]string, len(teams))
			for i, tm := range teams {
				names[i] = tm.DisplayName
			}
			t.KV("teams", strings.Join(names, ", "))
		}

		if cfg.Server.TeamID != "" {
			if channels, err := client.GetTeamChannels(cfg.Server.TeamID, 0, 200); err == nil {
				t.KV("channels", fmt.Sprintf("%d", len(channels)))
			}
		}

		if len(cfg.Agents) > 0 {
			fmt.Println()
			t.Rule()
			fmt.Println("  " + t.Bold(t.Cyan.S("Agents")))
			fmt.Println()
			tbl := t.NewTable("name", "status", "pid", "model")
			for _, a := range cfg.Agents {
				alive := matrixIsAlive(a.PID)
				st := a.Status
				if !alive && st == "running" {
					st = "dead"
				}
				stCell := t.Jade.S(st)
				if st != "running" {
					stCell = t.Loss.S(st)
				}
				tbl.Row(t.Bold(t.Gold.S(a.Name)), stCell, fmt.Sprintf("%d", a.PID), t.Dim(a.Model))
			}
			fmt.Print(tbl.Render())
		}

		if len(cfg.Daemons) > 0 {
			fmt.Println()
			t.Rule()
			fmt.Println("  " + t.Bold(t.Cyan.S("Daemons")))
			fmt.Println()
			tbl := t.NewTable("name", "type", "status", "pid")
			for _, d := range cfg.Daemons {
				alive := matrixIsAlive(d.PID)
				st := d.Status
				if !alive && st == "running" {
					st = "dead"
				}
				stCell := t.Jade.S(st)
				if st == "paused" {
					stCell = t.Gold.S(st)
				} else if st != "running" {
					stCell = t.Loss.S(st)
				}
				tbl.Row(t.Bold(t.Gold.S(d.Name)), t.Dim(d.Type), stCell, fmt.Sprintf("%d", d.PID))
			}
			fmt.Print(tbl.Render())
		}

		fmt.Println()
		return nil
	},
}

func init() {
	matrixCmd.AddCommand(matrixStatusCmd)
}
