package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var envServer string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage .env files on servers",
	Run:   runEnvHelp,
}

var envListCmd = &cobra.Command{
	Use:     "list <path>",
	Aliases: []string{"ls"},
	Short:   "List variables in a .env file",
	Long: `Examples:
  anime env list /home/ubuntu/myapp/.env
  anime env ls ~/api/.env -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runEnvList,
}

var envGetCmd = &cobra.Command{
	Use:   "get <path> <key>",
	Short: "Get a variable from a .env file",
	Long: `Examples:
  anime env get ~/api/.env DATABASE_URL
  anime env get ~/api/.env SECRET_KEY -s wings`,
	Args: cobra.ExactArgs(2),
	RunE: runEnvGet,
}

var envSetCmd = &cobra.Command{
	Use:   "set <path> <key> <value>",
	Short: "Set a variable in a .env file",
	Long: `Examples:
  anime env set ~/api/.env PORT 8080
  anime env set ~/api/.env DATABASE_URL "postgres://..." -s wings`,
	Args: cobra.ExactArgs(3),
	RunE: runEnvSet,
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove <path> <key>",
	Short: "Remove a variable from a .env file",
	Args:  cobra.ExactArgs(2),
	RunE:  runEnvRemove,
}

func init() {
	for _, c := range []*cobra.Command{envListCmd, envGetCmd, envSetCmd, envRemoveCmd} {
		c.Flags().StringVarP(&envServer, "server", "s", "", "Remote server")
	}
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envRemoveCmd)
	rootCmd.AddCommand(envCmd)
}

func runEnvHelp(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("📋 ENV 📋"))
	fmt.Println()
	cmds := []struct{ c, d string }{
		{"anime env list <path>", "List all variables"},
		{"anime env get <path> <key>", "Get a variable"},
		{"anime env set <path> <key> <value>", "Set a variable"},
		{"anime env remove <path> <key>", "Remove a variable"},
	}
	for _, c := range cmds {
		fmt.Printf("  %s\n    %s\n\n", theme.HighlightStyle.Render(c.c), theme.DimTextStyle.Render(c.d))
	}
	fmt.Println(theme.DimTextStyle.Render("  All commands support --server/-s for remote"))
	fmt.Println()
}

func runEnvList(cmd *cobra.Command, args []string) error {
	path := args[0]
	if err := validate.ShellSafe(path); err != nil {
		return err
	}
	script := fmt.Sprintf(`cat "%s" 2>/dev/null || echo "File not found: %s"`, path, path)
	output, err := runOnServer(envServer, script)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	fmt.Printf("  %s\n\n", theme.HighlightStyle.Render(path))
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		} else if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			fmt.Printf("  %s=%s\n", theme.HighlightStyle.Render(parts[0]), theme.DimTextStyle.Render(parts[1]))
		} else {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
	}
	fmt.Println()
	return nil
}

func runEnvGet(cmd *cobra.Command, args []string) error {
	path, key := args[0], args[1]
	if err := validate.ShellSafe(path); err != nil {
		return err
	}
	if err := validate.ShellSafe(key); err != nil {
		return err
	}
	script := fmt.Sprintf(`grep "^%s=" "%s" 2>/dev/null | head -1 | cut -d= -f2-`, key, path)
	output, err := runOnServer(envServer, script)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	value := strings.TrimSpace(output)
	if value == "" {
		fmt.Printf("  %s not found in %s\n", theme.WarningStyle.Render(key), path)
	} else {
		fmt.Printf("  %s=%s\n", theme.HighlightStyle.Render(key), theme.DimTextStyle.Render(value))
	}
	return nil
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	path, key, value := args[0], args[1], args[2]
	if err := validate.ShellSafe(path); err != nil {
		return err
	}
	if err := validate.ShellSafe(key); err != nil {
		return err
	}
	// Value can contain special chars but we use single quotes in shell
	script := fmt.Sprintf(`
touch "%s"
if grep -q "^%s=" "%s" 2>/dev/null; then
    sed -i "s|^%s=.*|%s=%s|" "%s"
    echo "Updated %s"
else
    echo '%s=%s' >> "%s"
    echo "Added %s"
fi`, path, key, path, key, key, value, path, key, key, value, path, key)
	output, err := runOnServer(envServer, script)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}

func runEnvRemove(cmd *cobra.Command, args []string) error {
	path, key := args[0], args[1]
	if err := validate.ShellSafe(path); err != nil {
		return err
	}
	if err := validate.ShellSafe(key); err != nil {
		return err
	}
	script := fmt.Sprintf(`sed -i "/^%s=/d" "%s" && echo "Removed %s"`, key, path, key)
	output, err := runOnServer(envServer, script)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}
