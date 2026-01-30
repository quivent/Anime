package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/embeddb"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the embedded database",
	Long: `Manage the database embedded directly in the anime binary.

The embedded database stores key-value pairs inside the binary itself.
This allows the binary to carry its own configuration without external files.

Storage is compressed and fits in a 64KB reserved block within the binary.`,
	RunE: runDBInfo,
}

var dbGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get a value from the embedded database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBGet,
}

var dbSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a value in the embedded database",
	Args:  cobra.ExactArgs(2),
	RunE:  runDBSet,
}

var dbDeleteCmd = &cobra.Command{
	Use:   "delete KEY",
	Short: "Delete a key from the embedded database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBDelete,
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entries in the embedded database",
	RunE:  runDBList,
}

var dbExportCmd = &cobra.Command{
	Use:   "export [FILE]",
	Short: "Export the embedded database to JSON",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDBExport,
}

var dbImportCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Import data into the embedded database from JSON",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBImport,
}

var dbClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all data from the embedded database",
	RunE:  runDBClear,
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbGetCmd)
	dbCmd.AddCommand(dbSetCmd)
	dbCmd.AddCommand(dbDeleteCmd)
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbExportCmd)
	dbCmd.AddCommand(dbImportCmd)
	dbCmd.AddCommand(dbClearCmd)
}

func runDBInfo(_ *cobra.Command, _ []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	info, err := db.Info()
	if err != nil {
		return fmt.Errorf("failed to get database info: %w", err)
	}

	fmt.Println(theme.InfoStyle.Render("Embedded Database"))
	fmt.Println()

	// Storage bar
	barWidth := 40
	usedWidth := (info.UsagePercent * barWidth) / 100
	if usedWidth == 0 && info.UsedBytes > 0 {
		usedWidth = 1
	}
	bar := strings.Repeat("█", usedWidth) + strings.Repeat("░", barWidth-usedWidth)

	fmt.Printf("  Storage:   [%s] %d%%\n", theme.HighlightStyle.Render(bar), info.UsagePercent)
	fmt.Printf("  Reserved:  %s\n", theme.DimTextStyle.Render(formatBytes(int64(info.ReservedBytes))))
	fmt.Printf("  Used:      %s\n", theme.HighlightStyle.Render(formatBytes(int64(info.UsedBytes))))
	fmt.Printf("  Free:      %s\n", theme.SuccessStyle.Render(formatBytes(int64(info.FreeBytes))))
	fmt.Printf("  Entries:   %d\n", info.Entries)
	fmt.Printf("  Binary:    %s\n", theme.DimTextStyle.Render(db.BinaryPath()))

	if info.UseLegacy {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("  ⚠ Using legacy append mode (rebuild binary to use reserved space)"))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Use 'anime db list' to see all entries"))
	fmt.Println(theme.DimTextStyle.Render("  Use 'anime db set KEY VALUE' to store data"))

	return nil
}

func runDBGet(_ *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	key := args[0]
	value := db.GetString(key)

	if value == "" {
		// Check if it's in other storage types
		if v := db.GetAlias(key); v != "" {
			fmt.Println(v)
			return nil
		}
		if v := db.GetSetting(key); v != "" {
			fmt.Println(v)
			return nil
		}
		return fmt.Errorf("key not found: %s", key)
	}

	fmt.Println(value)
	return nil
}

func runDBSet(_ *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	key := args[0]
	value := args[1]

	db.SetString(key, value)

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("%s %s = %s\n", theme.SuccessStyle.Render("✓"), key, value)
	return nil
}

func runDBDelete(_ *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	key := args[0]
	db.Delete(key)

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("%s deleted %s\n", theme.SuccessStyle.Render("✓"), key)
	return nil
}

func runDBList(_ *cobra.Command, _ []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	stats := db.Stats()
	hasData := false

	// KV entries
	keys := db.Keys()
	if len(keys) > 0 {
		hasData = true
		fmt.Println(theme.InfoStyle.Render("Key-Value Entries"))
		for _, k := range keys {
			v := db.GetString(k)
			if len(v) > 60 {
				v = v[:60] + "..."
			}
			fmt.Printf("  %s = %s\n", theme.HighlightStyle.Render(k), v)
		}
		fmt.Println()
	}

	// Aliases
	aliases := db.ListAliases()
	if len(aliases) > 0 {
		hasData = true
		fmt.Println(theme.InfoStyle.Render("Aliases"))
		for k, v := range aliases {
			fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(k), v)
		}
		fmt.Println()
	}

	// Shell aliases
	shellAliases := db.ListShellAliases()
	if len(shellAliases) > 0 {
		hasData = true
		fmt.Println(theme.InfoStyle.Render("Shell Aliases"))
		for k, v := range shellAliases {
			fmt.Printf("  %s → %s\n", theme.HighlightStyle.Render(k), v)
		}
		fmt.Println()
	}

	// Settings
	if stats["settings"] > 0 {
		hasData = true
		fmt.Println(theme.InfoStyle.Render("Settings"))
		data, _ := db.Export()
		var exported map[string]interface{}
		json.Unmarshal(data, &exported)
		if settings, ok := exported["t"].(map[string]interface{}); ok {
			for k, v := range settings {
				fmt.Printf("  %s = %v\n", theme.HighlightStyle.Render(k), v)
			}
		}
		fmt.Println()
	}

	if !hasData {
		fmt.Println(theme.DimTextStyle.Render("No data in embedded database"))
		fmt.Println()
		fmt.Println("  Use 'anime db set KEY VALUE' to store data")
	}

	return nil
}

func runDBExport(_ *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	data, err := db.Export()
	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	if len(args) > 0 {
		if err := os.WriteFile(args[0], data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("%s exported to %s\n", theme.SuccessStyle.Render("✓"), args[0])
	} else {
		fmt.Println(string(data))
	}

	return nil
}

func runDBImport(_ *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := db.Import(data); err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("%s imported from %s\n", theme.SuccessStyle.Render("✓"), args[0])
	return nil
}

func runDBClear(_ *cobra.Command, _ []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to open embedded database: %w", err)
	}

	db.Clear()

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Printf("%s database cleared\n", theme.SuccessStyle.Render("✓"))
	return nil
}
