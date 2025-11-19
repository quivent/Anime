package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	installNewCmd = &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install packages on Lambda server",
		Long:  "Install one or more packages with automatic dependency resolution",
		Run:   runInstallNew,
	}

	installRemote      bool
	installServer  string
	dryRun      bool
	skipConfirm bool
	phased      bool
)

func init() {
	rootCmd.AddCommand(installNewCmd)
	installNewCmd.Flags().BoolVarP(&installRemote, "remote", "r", false, "Install on remote server via SSH")
	installNewCmd.Flags().StringVarP(&installServer, "server", "s", "", "Server name (required for remote install)")
	installNewCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without doing it")
	installNewCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	installNewCmd.Flags().BoolVar(&phased, "phased", false, "Install in phases with confirmation between each")
}

func runInstallNew(cmd *cobra.Command, args []string) {
	// Show anime-themed help if no packages specified
	if len(args) == 0 {
		showInstallHelp()
		return
	}

	packageIDs := args

	// Resolve dependencies
	resolved, err := installer.ResolveDependencies(packageIDs)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error: " + err.Error()))
		os.Exit(1)
	}

	// Display anime-styled installation plan
	fmt.Println(theme.RenderBanner("⚡ INSTALLATION QUEST ⚡"))
	fmt.Println()

	var totalTime time.Duration
	for i, pkg := range resolved {
		marker := theme.SymbolBranch
		if i == len(resolved)-1 {
			marker = theme.SymbolLastBranch
		}

		// Determine if requested or dependency
		badge := ""
		badgeStyle := theme.DimTextStyle
		if contains(packageIDs, pkg.ID) {
			badge = " " + theme.SymbolStar + " SELECTED"
			badgeStyle = theme.SuccessStyle
		} else {
			badge = " " + theme.SymbolBolt + " auto-included"
			badgeStyle = theme.DimTextStyle
		}

		// Category emoji
		emoji := theme.SymbolSakura
		switch pkg.Category {
		case "Foundation":
			emoji = "🏗️"
		case "ML Framework":
			emoji = "🤖"
		case "LLM Runtime":
			emoji = "🔮"
		case "Models":
			emoji = "⭐"
		case "Application":
			emoji = "🎯"
		}

		fmt.Printf("%s %s %s%s\n",
			theme.InfoStyle.Render(marker),
			emoji,
			theme.HighlightStyle.Render(pkg.Name),
			badgeStyle.Render(badge))
		fmt.Printf("    %s\n", theme.SecondaryTextStyle.Render(pkg.Description))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(fmt.Sprintf("⏱️  %s  |  💾 %s", pkg.EstimatedTime, pkg.Size)))
		fmt.Println()

		totalTime += pkg.EstimatedTime
	}

	// Summary box
	summary := fmt.Sprintf("📦 Total: %d packages  |  ⏱️  Est. time: %s",
		len(resolved), totalTime)
	fmt.Println(theme.InfoStyle.Render(summary))
	fmt.Println()

	if dryRun {
		fmt.Println(theme.WarningStyle.Render("✨ Dry run complete - no changes made"))
		return
	}

	// Confirm
	if !skipConfirm {
		fmt.Print("Proceed with installation? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Installation cancelled")
			return
		}
	}

	// Execute installation
	if installRemote {
		runRemoteInstall(resolved)
	} else {
		runLocalInstall(resolved)
	}
}

func runLocalInstall(packages []*installer.Package) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 DEPLOYING PACKAGES 🚀"))
	fmt.Println()

	for i, pkg := range packages {
		if phased && i > 0 {
			fmt.Printf("\n%s\n",
				theme.WarningStyle.Render("⏸️  Press Enter to continue, or Ctrl+C to abort..."))
			bufio.NewReader(os.Stdin).ReadString('\n')
		}

		// Progress indicator
		fmt.Printf("[%s] %s %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%d/%d", i+1, len(packages))),
			theme.SymbolLoading,
			theme.HighlightStyle.Render("Installing "+pkg.Name))

		script, exists := installer.GetScript(pkg.ID)
		if !exists {
			fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Script not found"))
			continue
		}

		// Execute script
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  %s %s\n",
				theme.ErrorStyle.Render(theme.SymbolError+" FAILED"),
				theme.DimTextStyle.Render(fmt.Sprintf("(after %s)", elapsed.Round(time.Second))))
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("💥 Installation quest failed! Check errors above."))
			os.Exit(1)
		}

		// Success with elapsed time
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess+" COMPLETE"),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s)", elapsed.Round(time.Second))))
		fmt.Println()
	}

	// Victory banner
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println(theme.SuccessStyle.Render("  ✨ QUEST COMPLETE! All packages deployed! ✨"))
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
}

func runRemoteInstall(packages []*installer.Package) {
	if installServer == "" {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error: --server flag required"))
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error loading config: " + err.Error()))
		os.Exit(1)
	}

	server, err := cfg.GetServer(installServer)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " " + err.Error()))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🌐 Establishing connection to %s...", installServer)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("   Host: %s@%s", server.User, server.Host)))

	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Connection failed: " + err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Connected!"))
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 REMOTE DEPLOYMENT 🚀"))
	fmt.Println()

	for i, pkg := range packages {
		if phased && i > 0 {
			fmt.Printf("\n%s\n",
				theme.WarningStyle.Render("⏸️  Press Enter to continue, or Ctrl+C to abort..."))
			bufio.NewReader(os.Stdin).ReadString('\n')
		}

		fmt.Printf("[%s] %s %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%d/%d", i+1, len(packages))),
			theme.SymbolLoading,
			theme.HighlightStyle.Render("Remote: "+pkg.Name))

		script, exists := installer.GetScript(pkg.ID)
		if !exists {
			fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Script not found"))
			continue
		}

		start := time.Now()
		output, err := client.RunCommand(script)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " FAILED"))
			fmt.Println(theme.DimTextStyle.Render(output))
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("💥 Remote deployment failed!"))
			os.Exit(1)
		}

		// Show output (dimmed)
		if len(output) > 0 {
			fmt.Println(theme.DimTextStyle.Render("  " + strings.ReplaceAll(output, "\n", "\n  ")))
		}
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess+" COMPLETE"),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s)", elapsed.Round(time.Second))))
		fmt.Println()
	}

	// Victory banner
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ Remote quest complete on %s! ✨", installServer)))
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func showInstallHelp() {
	fmt.Println(theme.RenderBanner("⚡ ANIME INSTALL ⚡"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🌸 Choose your adventure:"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime packages"))
	fmt.Println(theme.DimTextStyle.Render("    Browse all available packages"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime interactive"))
	fmt.Println(theme.DimTextStyle.Render("    Select packages with beautiful TUI"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime install <package-id> [package-id...]"))
	fmt.Println(theme.DimTextStyle.Render("    Install specific packages"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  anime install core                    # Install core system"))
	fmt.Println(theme.DimTextStyle.Render("  anime install core pytorch ollama     # Install multiple packages"))
	fmt.Println(theme.DimTextStyle.Render("  anime install --dry-run core          # Preview installation"))
	fmt.Println(theme.DimTextStyle.Render("  anime install --phased core pytorch   # Install with confirmations"))
	fmt.Println(theme.DimTextStyle.Render("  anime install --remote -s lambda core # Install on remote server"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("📦 Available packages:"))
	packages := installer.GetPackages()
	for _, pkg := range packages {
		fmt.Printf("  %s %s\n",
			theme.WarningStyle.Render("•"),
			theme.DimTextStyle.Render(pkg.ID+" - "+pkg.Name))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("Run 'anime packages' for detailed package information"))
}
