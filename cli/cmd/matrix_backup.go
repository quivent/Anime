package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var matrixBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore Mattermost data",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var mxBackupOutput string
var mxBackupDBOnly bool

var matrixBackupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Backup Mattermost data + database",
	Example: `  anime matrix backup create
  anime matrix backup create -o ~/backups/mm-2026.tar.gz
  anime matrix backup create --db-only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()

		output := mxBackupOutput
		if output == "" {
			output = fmt.Sprintf("mattermost-backup-%s.tar.gz", time.Now().Format("2006-01-02-150405"))
		}

		if !mxBackupDBOnly {
			// Archive the install data directory
			dataDir := cfg.Install.DataDir
			if dataDir == "" {
				return fmt.Errorf("no install directory configured — run anime matrix setup first")
			}
			if _, err := os.Stat(dataDir); os.IsNotExist(err) {
				return fmt.Errorf("install directory does not exist: %s", dataDir)
			}

			fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Backing up "+dataDir+"..."))
			if err := matrixRunBash(fmt.Sprintf("tar -czf %s --exclude='%s/bin' --exclude='%s/logs' -C %s .", output, dataDir, dataDir, dataDir)); err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
		}

		// Database dump via pg_dump if available
		dbOut := output + ".sql"
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Dumping database..."))
		dbScript := fmt.Sprintf(`pg_dump -U mattermost mattermost > %s 2>/dev/null && echo ok || echo skip`, dbOut)
		_ = matrixRunBash(dbScript)

		if !mxBackupDBOnly {
			info, _ := os.Stat(output)
			size := "unknown"
			if info != nil {
				size = fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024))
			}
			fmt.Printf("  %s %s %s (%s)\n",
				theme.SymbolSuccess,
				theme.SuccessStyle.Render("Backup created:"),
				theme.HighlightStyle.Render(output),
				size)
		}
		if _, err := os.Stat(dbOut); err == nil {
			fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("DB dump: "+dbOut))
		}
		return nil
	},
}

var matrixBackupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore Mattermost data from a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		dataDir := cfg.Install.DataDir
		if dataDir == "" {
			return fmt.Errorf("no install directory configured")
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

		// Restore DB if SQL file exists
		dbFile := backupFile + ".sql"
		if _, err := os.Stat(dbFile); err == nil {
			fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Restoring database..."))
			_ = matrixRunBash(fmt.Sprintf("psql -U mattermost mattermost < %s 2>/dev/null", dbFile))
		}

		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Restored"))
		fmt.Println(theme.DimTextStyle.Render("  Restart server: anime matrix setup restart"))
		return nil
	},
}

func init() {
	matrixBackupCreateCmd.Flags().StringVarP(&mxBackupOutput, "output", "o", "", "Output file path")
	matrixBackupCreateCmd.Flags().BoolVar(&mxBackupDBOnly, "db-only", false, "Only dump the database")
	matrixBackupCmd.AddCommand(matrixBackupCreateCmd)
	matrixBackupCmd.AddCommand(matrixBackupRestoreCmd)
	matrixCmd.AddCommand(matrixBackupCmd)
}
