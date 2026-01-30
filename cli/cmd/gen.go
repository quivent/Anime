package cmd

import (
	"fmt"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/spf13/cobra"
)

var (
	genName    string
	genHost    string
	genUser    string
	genKey     string
	genCost    float64
	genModules []string
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate raw bash commands to run on server",
	Long:  `Prints the actual bash commands to copy/paste into your SSH session. No anime needed on server.`,
	Example: `  anime gen
  anime gen -m core,pytorch,ollama,models-large`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// DEFAULT: core, pytorch, ollama, models-small
		if len(genModules) == 0 {
			genModules = []string{"core", "pytorch", "ollama", "models-small"}
		}

		// Resolve dependencies
		moduleIDs := genModules
		resolved := resolveDeps(moduleIDs)

		fmt.Println("# Copy and paste these commands into your SSH session:")
		fmt.Println("# ssh ubuntu@YOUR_SERVER_IP")
		fmt.Println()

		totalMinutes := 0
		for _, modID := range resolved {
			script, ok := installer.GetScript(modID)
			if !ok {
				continue
			}

			// Find module info
			var modInfo *config.Module
			for _, m := range config.AvailableModules {
				if m.ID == modID {
					modInfo = &m
					break
				}
			}

			if modInfo != nil {
				totalMinutes += modInfo.TimeMinutes
				fmt.Printf("# Installing: %s (%d minutes)\n", modInfo.Name, modInfo.TimeMinutes)
			}

			// Print the script directly
			fmt.Println(script)
			fmt.Println()
		}

		fmt.Printf("# Total estimated time: %d minutes\n", totalMinutes)
		fmt.Printf("# Estimated cost: $%.2f @ $20/hr\n", float64(totalMinutes)/60.0*20.0)

		return nil
	},
}

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Show pre-made bash command templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("# SSH into your server first:")
		fmt.Println("# ssh ubuntu@YOUR_SERVER_IP")
		fmt.Println()
		fmt.Println("# Then copy/paste one of these:")
		fmt.Println()

		templates := map[string][]string{
			"Minimal (Core + PyTorch, ~$3, 7min)": {"core", "pytorch"},
			"Standard (Core + PyTorch + Ollama + Small Models, ~$8, 16min)": {"core", "pytorch", "ollama", "models-small"},
			"Full (Everything, ~$25, 60min)":                                {"core", "pytorch", "ollama", "models-large", "comfyui", "claude"},
			"Just Core (CUDA, Python, Node, Docker, ~$2, 5min)":             {"core"},
			"LLM Only (Core + Ollama + Large Models, ~$15, 46min)":          {"core", "ollama", "models-large"},
		}

		for name, mods := range templates {
			fmt.Printf("## %s\n", name)
			fmt.Printf("anime gen -m %s\n\n", strings.Join(mods, ","))
		}

		fmt.Println("## Available modules:")
		for _, mod := range config.AvailableModules {
			fmt.Printf("  %s - %s (%dm)\n", mod.ID, mod.Name, mod.TimeMinutes)
		}
		fmt.Println()
		fmt.Println("## Usage:")
		fmt.Println("anime gen              # Default: core,pytorch,ollama,models-small")
		fmt.Println("anime gen -m core,pytorch,ollama")
		fmt.Println()
		fmt.Println("This will print the actual bash commands to run on your server.")
		fmt.Println("No need for anime on the server!")

		return nil
	},
}

func resolveDeps(moduleIDs []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	var addDeps func(string)
	addDeps = func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true

		for _, mod := range config.AvailableModules {
			if mod.ID == id {
				for _, dep := range mod.Dependencies {
					addDeps(dep)
				}
				result = append(result, id)
				break
			}
		}
	}

	for _, id := range moduleIDs {
		addDeps(id)
	}

	return result
}

func init() {
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(templatesCmd)

	genCmd.Flags().StringSliceVarP(&genModules, "modules", "m", []string{}, "Comma-separated module list (default: core,pytorch,ollama,models-small)")
}
