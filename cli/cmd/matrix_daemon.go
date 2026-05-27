package cmd

import (
	"fmt"
	"syscall"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var matrixDaemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"daemons", "d"},
	Short:   "Manage background daemon processes",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixDaemonListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all daemons",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		t.Section("DAEMONS")
		if len(cfg.Daemons) == 0 {
			t.Info("no daemons")
			fmt.Println()
			return nil
		}
		tbl := t.NewTable("name", "type", "status", "pid", "started")
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
			tbl.Row(
				t.Bold(t.Gold.S(d.Name)),
				t.Dim(d.Type),
				stCell,
				fmt.Sprintf("%d", d.PID),
				t.Dim(d.StartedAt),
			)
		}
		fmt.Print(tbl.Render())
		fmt.Println()
		return nil
	},
}

var matrixDaemonStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a daemon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(-d.PID, syscall.SIGTERM)
				}
				cfg.Daemons[i].Status = "stopped"
				cfg.Save()
				t.Ok("stopped " + args[0])
				return nil
			}
		}
		return fmt.Errorf("daemon %q not found", args[0])
	},
}

var matrixDaemonStopAllCmd = &cobra.Command{
	Use:   "stop-all",
	Short: "Stop all daemons",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		count := 0
		for i, d := range cfg.Daemons {
			if d.Status == "running" || d.Status == "paused" {
				if d.PID > 0 {
					syscall.Kill(-d.PID, syscall.SIGTERM)
				}
				cfg.Daemons[i].Status = "stopped"
				count++
			}
		}
		cfg.Save()
		t.Ok(fmt.Sprintf("stopped %d daemons", count))
		return nil
	},
}

var matrixDaemonPauseCmd = &cobra.Command{
	Use:   "pause <name>",
	Short: "Pause a daemon (SIGSTOP)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(d.PID, syscall.SIGSTOP)
				}
				cfg.Daemons[i].Status = "paused"
				cfg.Save()
				t.Ok("paused " + args[0])
				return nil
			}
		}
		return fmt.Errorf("daemon %q not found", args[0])
	},
}

var matrixDaemonResumeCmd = &cobra.Command{
	Use:   "resume <name>",
	Short: "Resume a paused daemon (SIGCONT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(d.PID, syscall.SIGCONT)
				}
				cfg.Daemons[i].Status = "running"
				cfg.Save()
				t.Ok("resumed " + args[0])
				return nil
			}
		}
		return fmt.Errorf("daemon %q not found", args[0])
	},
}

var matrixDaemonLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show daemon logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		for _, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.LogFile == "" {
					return fmt.Errorf("no log file")
				}
				return matrixRunBash(fmt.Sprintf("tail -100 %s", d.LogFile))
			}
		}
		return fmt.Errorf("daemon %q not found", args[0])
	},
}

var matrixDaemonCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove dead daemon entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		var alive []mmcfg.DaemonConfig
		cleaned := 0
		for _, d := range cfg.Daemons {
			if d.Status == "stopped" || !matrixIsAlive(d.PID) {
				cleaned++
			} else {
				alive = append(alive, d)
			}
		}
		cfg.Daemons = alive
		cfg.Save()
		t.Ok(fmt.Sprintf("cleaned %d dead daemons", cleaned))
		return nil
	},
}

func init() {
	matrixDaemonCmd.AddCommand(matrixDaemonListCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonStopCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonStopAllCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonPauseCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonResumeCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonLogsCmd)
	matrixDaemonCmd.AddCommand(matrixDaemonCleanCmd)
	matrixCmd.AddCommand(matrixDaemonCmd)
}
