package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/protocol"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	protocolDryRun    bool
	protocolVerify    bool
	protocolServer    string
	protocolLogFile   string
	protocolSkipPhases []string
	protocolOnlyPhases []string
)

var protocolCmd = &cobra.Command{
	Use:   "protocol <protocol-name>",
	Short: "Run pre-defined setup protocols",
	Long: `Execute multi-phase setup protocols for common deployment scenarios.

Protocols are pre-defined sequences of installation and configuration steps
that automate complex deployments like LLM servers, GPU clusters, and more.

Available protocols:
  coverage    Deploy DeepSeek V3.2-Exp on 8×B200 cluster

Examples:
  anime protocol list                    # List available protocols
  anime protocol coverage --dry-run      # Preview the coverage protocol
  anime protocol coverage --verify       # Run with verification after each phase
  anime protocol coverage --server lambda-1  # Run on remote server
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// No args - just show protocols list
			return runProtocolList(cmd, args)
		}
		return runProtocol(cmd, args)
	},
}

var protocolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available protocols",
	Long:  "Display all registered protocols with their descriptions and requirements",
	RunE:  runProtocolList,
}

func init() {
	// Protocol flags
	protocolCmd.Flags().BoolVar(&protocolDryRun, "dry-run", false, "Preview protocol without executing")
	protocolCmd.Flags().BoolVar(&protocolVerify, "verify", false, "Run verification after each phase")
	protocolCmd.Flags().StringVar(&protocolServer, "server", "", "Run on remote server (name or alias)")
	protocolCmd.Flags().StringVar(&protocolLogFile, "log", "", "Log output to file")
	protocolCmd.Flags().StringSliceVar(&protocolSkipPhases, "skip-phases", []string{}, "Phases to skip")
	protocolCmd.Flags().StringSliceVar(&protocolOnlyPhases, "only-phases", []string{}, "Only run these phases")

	// Add subcommands
	protocolCmd.AddCommand(protocolListCmd)

	// Register with root
	rootCmd.AddCommand(protocolCmd)
}

func runProtocol(cmd *cobra.Command, args []string) error {
	protocolName := args[0]

	// Get the protocol registry
	registry := protocol.GetGlobalRegistry()

	// Get the protocol
	proto, err := registry.Get(protocolName)
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Protocol not found: %s", protocolName)))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📋 Available protocols:"))

		protocols := registry.List()
		if len(protocols) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  No protocols registered"))
		} else {
			for _, p := range protocols {
				fmt.Printf("  %s - %s\n",
					theme.HighlightStyle.Render(p.Name),
					theme.DimTextStyle.Render(p.Description))
			}
		}
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Use: anime protocol list    # For more details"))
		fmt.Println()
		return err
	}

	// Build execution options
	options := protocol.ExecutionOptions{
		DryRun:       protocolDryRun,
		Verify:       protocolVerify,
		LogFile:      protocolLogFile,
		StopOnError:  true,
		SkipPhases:   protocolSkipPhases,
		OnlyPhases:   protocolOnlyPhases,
		AutoContinue: true,
	}

	// Handle remote execution
	if protocolServer != "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Try to get server (handles both names and aliases)
		server, err := cfg.GetServer(protocolServer)
		if err != nil {
			// Try as alias
			target := cfg.GetAlias(protocolServer)
			if target == "" {
				return fmt.Errorf("server not found: %s", protocolServer)
			}

			// Parse target
			parts := strings.Split(target, "@")
			if len(parts) == 2 {
				options.SSHUser = parts[0]
				options.SSHHost = parts[1]
			} else {
				options.SSHHost = target
				options.SSHUser = "ubuntu" // default
			}
		} else {
			options.SSHHost = server.Host
			options.SSHUser = server.User
			options.SSHKey = server.SSHKey
		}

		options.Server = protocolServer

		// Mark all commands as remote
		for _, phase := range proto.Phases {
			for i := range phase.Commands {
				phase.Commands[i].Remote = true
			}
		}
	}

	// Create executor
	executor, err := protocol.NewExecutor(options)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer executor.Close()

	// Execute protocol
	result, err := executor.Execute(proto)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Exit with error if protocol failed
	if !result.Success {
		os.Exit(1)
	}

	return nil
}

func runProtocolList(cmd *cobra.Command, args []string) error {
	registry := protocol.GetGlobalRegistry()
	protocols := registry.List()

	fmt.Println()
	fmt.Println(theme.RenderBanner("🎴 AVAILABLE PROTOCOLS 🎴"))
	fmt.Println()

	if len(protocols) == 0 {
		fmt.Println(theme.DimTextStyle.Render("No protocols registered"))
		fmt.Println()
		return nil
	}

	// Group by category
	categories := registry.Categories()

	if len(categories) == 0 {
		// No categories, just list all
		for _, p := range protocols {
			printProtocolInfo(p)
		}
	} else {
		// List by category
		for _, category := range categories {
			categoryProtos := registry.ListByCategory(category)
			if len(categoryProtos) == 0 {
				continue
			}

			fmt.Println(theme.GetCategoryStyle(category).Render(fmt.Sprintf("📂 %s", category)))
			fmt.Println()

			for _, p := range categoryProtos {
				printProtocolInfo(p)
			}
		}

		// List uncategorized
		uncategorized := []*protocol.Protocol{}
		for _, p := range protocols {
			if p.Category == "" {
				uncategorized = append(uncategorized, p)
			}
		}

		if len(uncategorized) > 0 {
			fmt.Println(theme.HeaderStyle.Render("📂 Other"))
			fmt.Println()

			for _, p := range uncategorized {
				printProtocolInfo(p)
			}
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("💡 Usage:"))
	fmt.Println(theme.DimTextStyle.Render("  anime protocol <name> --dry-run    # Preview a protocol"))
	fmt.Println(theme.DimTextStyle.Render("  anime protocol <name>              # Execute a protocol"))
	fmt.Println()

	return nil
}

func printProtocolInfo(p *protocol.Protocol) {
	// Protocol name and version
	fmt.Printf("  %s %s\n",
		theme.HighlightStyle.Render(p.Name),
		theme.DimTextStyle.Render("v"+p.Version))

	// Description
	fmt.Printf("    %s\n", theme.DimTextStyle.Render(p.Description))

	// Requirements
	if p.Requirements.GPUs > 0 {
		fmt.Printf("    %s %d × GPU (%dGB each)\n",
			theme.DimTextStyle.Render(theme.SymbolBolt),
			p.Requirements.GPUs,
			p.Requirements.GPUMemoryGB)
	}

	if p.Requirements.CUDA != "" {
		fmt.Printf("    %s CUDA: %s\n",
			theme.DimTextStyle.Render(theme.SymbolConfig),
			p.Requirements.CUDA)
	}

	if p.Requirements.Python != "" {
		fmt.Printf("    %s Python: %s\n",
			theme.DimTextStyle.Render(theme.SymbolConfig),
			p.Requirements.Python)
	}

	// Phases
	fmt.Printf("    %s %d phases\n",
		theme.DimTextStyle.Render(theme.SymbolModule),
		len(p.Phases))

	// Phase names
	fmt.Print("    ")
	phaseNames := []string{}
	for _, phase := range p.Phases {
		phaseNames = append(phaseNames, phase.Name)
	}
	fmt.Println(theme.DimTextStyle.Render("    → " + strings.Join(phaseNames, " → ")))

	fmt.Println()
}
