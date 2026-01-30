package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var ollamaCmd = &cobra.Command{
	Use:                "ollama [args]",
	Short:              "Run ollama commands",
	Long:               `Alias for the ollama command. Passes all arguments directly to ollama.`,
	DisableFlagParsing: true, // Pass all flags directly to ollama
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if this is the start subcommand
		if len(args) > 0 && args[0] == "start" {
			// Let the start subcommand handle this
			return nil
		}

		// Find ollama in PATH
		ollamaPath, err := exec.LookPath("ollama")
		if err != nil {
			return fmt.Errorf("ollama not found in PATH. Please install ollama first")
		}

		// Create the command with all arguments
		ollamaExec := exec.Command(ollamaPath, args...)

		// Connect stdin, stdout, stderr to allow interactive usage
		ollamaExec.Stdin = os.Stdin
		ollamaExec.Stdout = os.Stdout
		ollamaExec.Stderr = os.Stderr

		// Run the command
		if err := ollamaExec.Run(); err != nil {
			// If ollama exits with non-zero, preserve that exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return err
		}

		return nil
	},
}

var ollamaStartCmd = &cobra.Command{
	Use:   "start <model-name>",
	Short: "Start an Ollama model in the background",
	Long:  `Starts an Ollama model running in the background using nohup.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Model name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime ollama start <model-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime ollama start llama3.3:8b"))
			fmt.Println(theme.DimTextStyle.Render("  anime ollama start mistral"))
			fmt.Println(theme.DimTextStyle.Render("  anime ollama start qwen3:14b"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  ollama list           # List downloaded models"))
			fmt.Println(theme.DimTextStyle.Render("  anime query <model>   # Query a model interactively"))
			fmt.Println()
			return fmt.Errorf("start requires a model name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		// Find ollama in PATH
		ollamaPath, err := exec.LookPath("ollama")
		if err != nil {
			return fmt.Errorf("ollama not found in PATH. Please install ollama first")
		}

		// Create a background command using nohup
		// This will run: nohup ollama run <model> > /tmp/ollama-<model>.log 2>&1 &
		logFile := fmt.Sprintf("/tmp/ollama-%s.log", modelName)

		// Use bash to run the command in the background with nohup
		bashCmd := fmt.Sprintf("nohup %s run %s > %s 2>&1 &", ollamaPath, modelName, logFile)
		ollamaExec := exec.Command("bash", "-c", bashCmd)

		// Start the command
		if err := ollamaExec.Start(); err != nil {
			return fmt.Errorf("failed to start ollama: %w", err)
		}

		fmt.Printf("✓ Started Ollama model '%s' in the background\n", modelName)
		fmt.Printf("  Log file: %s\n", logFile)
		fmt.Printf("  To view logs: tail -f %s\n", logFile)
		fmt.Printf("  To stop: pkill -f 'ollama run %s'\n", modelName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ollamaCmd)
	ollamaCmd.AddCommand(ollamaStartCmd)
}
