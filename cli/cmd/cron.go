package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var cronServer string

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Manage cron jobs",
	Run:   runCronHelp,
}

var cronListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List cron jobs",
	RunE:    runCronList,
}

var cronAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a cron job (interactive)",
	Long: `Interactive wizard to add a cron job.

Examples:
  anime cron add
  anime cron add -s wings`,
	RunE: runCronAdd,
}

var cronRemoveCmd = &cobra.Command{
	Use:   "remove <line-number>",
	Short: "Remove a cron job by line number",
	Long: `Remove a cron job. Use 'anime cron list' first to see line numbers.

Examples:
  anime cron remove 3
  anime cron remove 1 -s wings`,
	Args: cobra.ExactArgs(1),
	RunE: runCronRemove,
}

func init() {
	for _, c := range []*cobra.Command{cronListCmd, cronAddCmd, cronRemoveCmd} {
		c.Flags().StringVarP(&cronServer, "server", "s", "", "Remote server")
	}
	cronCmd.AddCommand(cronListCmd)
	cronCmd.AddCommand(cronAddCmd)
	cronCmd.AddCommand(cronRemoveCmd)
	rootCmd.AddCommand(cronCmd)
}

func runCronHelp(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("⏰ CRON ⏰"))
	fmt.Println()
	cmds := []struct{ c, d string }{
		{"anime cron list", "List all cron jobs"},
		{"anime cron add", "Add a cron job (interactive)"},
		{"anime cron remove <n>", "Remove cron job by line number"},
	}
	for _, c := range cmds {
		fmt.Printf("  %s\n    %s\n\n", theme.HighlightStyle.Render(c.c), theme.DimTextStyle.Render(c.d))
	}
}

func runCronList(cmd *cobra.Command, args []string) error {
	output, err := runOnServer(cronServer, "crontab -l 2>/dev/null || echo '(no crontab)'")
	if err != nil && output == "" {
		return fmt.Errorf("failed: %w", err)
	}

	fmt.Println()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lineNum := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		} else {
			lineNum++
			fmt.Printf("  %s  %s\n",
				theme.InfoStyle.Render(fmt.Sprintf("[%d]", lineNum)),
				theme.HighlightStyle.Render(line))
		}
	}
	fmt.Println()
	return nil
}

func runCronAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.RenderBanner("⏰ ADD CRON JOB ⏰"))
	fmt.Println()

	// Schedule presets
	fmt.Println(theme.InfoStyle.Render("  1. Schedule"))
	fmt.Println()
	presets := []struct{ label, cron string }{
		{"Every minute", "* * * * *"},
		{"Every 5 minutes", "*/5 * * * *"},
		{"Every hour", "0 * * * *"},
		{"Every day at midnight", "0 0 * * *"},
		{"Every day at 6am", "0 6 * * *"},
		{"Every Sunday at midnight", "0 0 * * 0"},
		{"Custom (enter your own)", ""},
	}
	for i, p := range presets {
		if p.cron != "" {
			fmt.Printf("    %s %s  %s\n",
				theme.HighlightStyle.Render(fmt.Sprintf("%d", i+1)),
				theme.InfoStyle.Render(p.label),
				theme.DimTextStyle.Render(p.cron))
		} else {
			fmt.Printf("    %s %s\n",
				theme.HighlightStyle.Render(fmt.Sprintf("%d", i+1)),
				theme.InfoStyle.Render(p.label))
		}
	}
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	schedChoice, _ := reader.ReadString('\n')
	schedChoice = strings.TrimSpace(schedChoice)
	if schedChoice == "" {
		schedChoice = "1"
	}

	var schedule string
	idx := 0
	fmt.Sscanf(schedChoice, "%d", &idx)
	if idx >= 1 && idx <= len(presets) && presets[idx-1].cron != "" {
		schedule = presets[idx-1].cron
	} else {
		fmt.Print("  Cron expression (e.g. */5 * * * *): ")
		schedule, _ = reader.ReadString('\n')
		schedule = strings.TrimSpace(schedule)
	}

	if schedule == "" {
		return fmt.Errorf("schedule required")
	}

	fmt.Println()
	fmt.Print("  2. Command to run: ")
	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("command required")
	}

	if err := validate.ShellSafe(command); err != nil {
		return fmt.Errorf("unsafe command: %w", err)
	}

	cronLine := fmt.Sprintf("%s %s", schedule, command)
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Job:"), theme.HighlightStyle.Render(cronLine))
	fmt.Print("  Add this job? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) == "n" {
		fmt.Println("  Cancelled")
		return nil
	}

	script := fmt.Sprintf(`(crontab -l 2>/dev/null; echo '%s') | crontab - && echo "Job added"`, cronLine)
	output, err := runOnServer(cronServer, script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}

func runCronRemove(cmd *cobra.Command, args []string) error {
	lineNum := args[0]
	if err := validate.ShellSafe(lineNum); err != nil {
		return err
	}

	// Get current crontab, remove the nth non-comment line
	script := fmt.Sprintf(`
TEMP=$(mktemp)
crontab -l 2>/dev/null > "$TEMP"
LINE=0
FOUND=""
while IFS= read -r row; do
    case "$row" in \#*|"") ;; *)
        LINE=$((LINE+1))
        if [ "$LINE" = "%s" ]; then
            FOUND="$row"
            continue
        fi
    ;; esac
    echo "$row"
done < "$TEMP" | crontab -
rm "$TEMP"
if [ -n "$FOUND" ]; then
    echo "Removed: $FOUND"
else
    echo "Line %s not found"
fi`, lineNum, lineNum)

	output, err := runOnServer(cronServer, script)
	if err != nil {
		return fmt.Errorf("failed: %s", output)
	}
	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	return nil
}
