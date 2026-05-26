package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore Synapse data",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var mxBackupOutput string

var matrixBackupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup of the Synapse data directory",
	Example: `  anime matrix backup create
  anime matrix backup create -o ~/backups/matrix-2026-05-26.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		dataDir := cfg.Synapse.DataDir
		if dataDir == "" {
			return fmt.Errorf("no data directory configured")
		}
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			return fmt.Errorf("data directory does not exist: %s", dataDir)
		}

		output := mxBackupOutput
		if output == "" {
			output = fmt.Sprintf("matrix-backup-%s.tar.gz", time.Now().Format("2006-01-02-150405"))
		}

		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Backing up "+dataDir+"..."))
		if err := matrixRunBash(fmt.Sprintf("tar -czf %s -C %s .", output, dataDir)); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		info, _ := os.Stat(output)
		size := "unknown"
		if info != nil {
			size = fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024))
		}
		fmt.Printf("  %s %s %s (%s)\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Backup created:"), theme.HighlightStyle.Render(output), size)
		return nil
	},
}

var matrixBackupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore Synapse data from a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		dataDir := cfg.Synapse.DataDir
		if dataDir == "" {
			return fmt.Errorf("no data directory configured")
		}

		backupFile := args[0]
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			return fmt.Errorf("backup file not found: %s", backupFile)
		}

		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Restoring to "+dataDir+"..."))
		os.MkdirAll(dataDir, 0755)
		if err := matrixRunBash(fmt.Sprintf("tar -xzf %s -C %s", backupFile, dataDir)); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Restored"))
		fmt.Println(theme.DimTextStyle.Render("  Restart Synapse: anime matrix setup restart"))
		return nil
	},
}

func init() {
	matrixBackupCreateCmd.Flags().StringVarP(&mxBackupOutput, "output", "o", "", "Output file path")
	matrixBackupCmd.AddCommand(matrixBackupCreateCmd)
	matrixBackupCmd.AddCommand(matrixBackupRestoreCmd)
	matrixCmd.AddCommand(matrixBackupCmd)
}
