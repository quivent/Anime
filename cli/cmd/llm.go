package cmd

import (
	"github.com/spf13/cobra"
)

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Quick shortcut for 'anime start ollama'",
	Long:  `Start Ollama LLM server with automatic port forwarding. Shortcut for 'anime start ollama'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call start command with ollama
		return runStart(cmd, []string{"ollama"})
	},
}

func init() {
	llmCmd.Flags().StringVarP(&startServer, "server", "s", "lambda", "Server to use")
	rootCmd.AddCommand(llmCmd)
}
