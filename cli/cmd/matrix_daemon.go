package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
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
		cfg, _ := matrixcfg.Load()
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX DAEMONS"))
		fmt.Println()
		if len(cfg.Daemons) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  No daemons"))
			fmt.Println()
			return nil
		}
		for _, d := range cfg.Daemons {
			alive := matrixIsAlive(d.PID)
			st := d.Status
			if !alive && st == "running" {
				st = "dead"
			}
			stStr := theme.SuccessStyle.Render(st)
			if st == "paused" {
				stStr = theme.WarningStyle.Render(st)
			} else if st == "dead" || st == "stopped" {
				stStr = theme.ErrorStyle.Render(st)
			}
			fmt.Printf("  %s %-20s %s  PID %-8d  %s\n",
				theme.SymbolBolt, theme.HighlightStyle.Render(d.Name), stStr, d.PID, theme.DimTextStyle.Render(d.Type))
			if d.StartedAt != "" {
				fmt.Printf("    Started: %s\n", theme.DimTextStyle.Render(d.StartedAt))
			}
		}
		fmt.Println()
		return nil
	},
}

var matrixDaemonStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a daemon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(-d.PID, syscall.SIGTERM)
				}
				cfg.Daemons[i].Status = "stopped"
				cfg.Save()
				fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Stopped "+args[0]))
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
		cfg, _ := matrixcfg.Load()
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
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render(fmt.Sprintf("Stopped %d daemons", count)))
		return nil
	},
}

var matrixDaemonPauseCmd = &cobra.Command{
	Use:   "pause <name>",
	Short: "Pause a daemon (SIGSTOP)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(d.PID, syscall.SIGSTOP)
				}
				cfg.Daemons[i].Status = "paused"
				cfg.Save()
				fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Paused "+args[0]))
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
		cfg, _ := matrixcfg.Load()
		for i, d := range cfg.Daemons {
			if d.Name == args[0] {
				if d.PID > 0 {
					syscall.Kill(d.PID, syscall.SIGCONT)
				}
				cfg.Daemons[i].Status = "running"
				cfg.Save()
				fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Resumed "+args[0]))
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
		cfg, _ := matrixcfg.Load()
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
		cfg, _ := matrixcfg.Load()
		var alive []matrixcfg.DaemonConfig
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
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render(fmt.Sprintf("Cleaned %d dead daemons", cleaned)))
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

// Ensure os import is used
var _ = os.Getpid
