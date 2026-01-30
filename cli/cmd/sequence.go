package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var seqModules []string

var sequenceCmd = &cobra.Command{
	Use:   "sequence",
	Short: "List commands in order - no scripts, no comments",
	Long:  `Just the command sequence. 1, 2, 3, 4, 5... Done.`,
	Example: `  anime sequence
  anime sequence -m core,pytorch,ollama,models-large`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// DEFAULT: core, pytorch, ollama, models-small
		if len(seqModules) == 0 {
			seqModules = []string{"core", "pytorch", "ollama", "models-small"}
		}

		// Resolve dependencies
		resolved := resolveDeps(seqModules)

		for _, modID := range resolved {
			script, ok := installer.GetScript(modID)
			if !ok {
				continue
			}

			// Parse script to extract actual commands
			lines := strings.Split(script, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)

				// Skip empty, comments, shebang, set -e
				if line == "" ||
				   strings.HasPrefix(line, "#") ||
				   strings.HasPrefix(line, "set -e") ||
				   line == "fi" ||
				   line == "EOF" ||
				   strings.HasPrefix(line, "EOF") {
					continue
				}

				// Skip if/then/else structure markers
				if strings.HasPrefix(line, "if ") ||
				   strings.HasPrefix(line, "then") ||
				   strings.HasPrefix(line, "else") ||
				   line == "fi" {
					continue
				}

				// Print command
				fmt.Println(line)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sequenceCmd)
	sequenceCmd.Flags().StringSliceVarP(&seqModules, "modules", "m", []string{}, "Comma-separated module list (default: core,pytorch,ollama,models-small)")
}
