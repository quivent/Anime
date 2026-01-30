package cmd

import (
	"github.com/spf13/cobra"
)

var notebookCmd = &cobra.Command{
	Use:   "notebook",
	Short: "Quick shortcut for 'anime start jupyter'",
	Long:  `Start Jupyter notebook with automatic port forwarding. Shortcut for 'anime start jupyter'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call start command with jupyter
		return runStart(cmd, []string{"jupyter"})
	},
}

func init() {
	notebookCmd.Flags().StringVarP(&startServer, "server", "s", "lambda", "Server to use")
	rootCmd.AddCommand(notebookCmd)
}
