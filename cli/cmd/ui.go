package cmd

import (
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Quick shortcut for 'anime start comfyui'",
	Long:  `Start ComfyUI with automatic port forwarding. Shortcut for 'anime start comfyui'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call start command with comfyui
		return runStart(cmd, []string{"comfyui"})
	},
}

func init() {
	uiCmd.Flags().StringVarP(&startServer, "server", "s", "lambda", "Server to use")
	rootCmd.AddCommand(uiCmd)
}
