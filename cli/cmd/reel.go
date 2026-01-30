package cmd

import (
	"github.com/joshkornreich/anime/internal/reel"
	"github.com/spf13/cobra"
)

var reelCmd = &cobra.Command{
	Use:   "reel [subcommand] [args]",
	Short: "SkyReels video generation",
	Long: `SkyReels Video Generation - Subcommand-based interface for SkyReels-V2.

Configure video generation parameters step by step, then execute.
Session state is preserved between commands.

Configuration Subcommands:
  prompt       Set the generation prompt
  frames       Set frame count (97, 4s, 8s, etc)
  resolution   Set resolution (540p, 720p)
  steps        Set inference steps
  guidance     Set CFG guidance scale
  seed         Set random seed
  output       Set output directory
  model        Select model variant
  image        Set input image (for image-to-video)
  fps          Set frames per second
  script       Select generation script

Optimization Subcommands:
  usp          Toggle multi-GPU parallelism
  offload      Toggle CPU offloading
  teacache     Toggle/configure TeaCache acceleration

Execution Subcommands:
  run          Execute video generation
  dry          Preview without executing
  show         Show current configuration
  reset        Reset to defaults

Examples:
  # Quick generation workflow
  anime reel prompt "A serene lake at sunset with mountains"
  anime reel frames 4s
  anime reel resolution 720p
  anime reel run

  # One-liner with chaining
  anime reel prompt "Ocean waves" && anime reel run

  # Fast preview with TeaCache
  anime reel teacache 0.3
  anime reel run

  # Check current config
  anime reel show`,
	Run: func(cmd *cobra.Command, args []string) {
		reel.Handle(args)
	},
	// Allow arbitrary args to be passed through to subcommand handler
	DisableFlagParsing: true,
}

func init() {
	rootCmd.AddCommand(reelCmd)
}
