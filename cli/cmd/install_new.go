package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/hf"
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
	installAuto bool
	installFrom int
)

func init() {
	rootCmd.AddCommand(installNewCmd)
	installNewCmd.Flags().BoolVarP(&installRemote, "remote", "r", false, "Install on remote server via SSH")
	installNewCmd.Flags().StringVarP(&installServer, "server", "s", "", "Server name (required for remote install)")
	installNewCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without doing it")
	installNewCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	installNewCmd.Flags().BoolVar(&phased, "phased", false, "Install one at a time with confirmation between each")
	installNewCmd.Flags().BoolVar(&installAuto, "auto", false, "Run all steps without pausing (non-interactive)")
	installNewCmd.Flags().IntVar(&installFrom, "from", 0, "Resume from step N (skip already-completed packages)")
}

func runInstallNew(cmd *cobra.Command, args []string) {
	// Show anime-themed help if no packages specified
	if len(args) == 0 {
		showInstallHelp()
		return
	}

	packageIDs := args

	// Default to step-by-step for heavy bundles unless --auto is passed
	if !installAuto && !cmd.Flags().Changed("phased") {
		for _, id := range packageIDs {
			if id == "wan" || id == "comfort" || id == "llama" {
				phased = true
				break
			}
		}
	}
	if installAuto {
		phased = false
	}

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
		fmt.Print("Proceed with installation? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "n" || response == "no" {
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

// installFailure tracks a failed package installation
type installFailure struct {
	pkg     *installer.Package
	err     error
	elapsed time.Duration
}

func runLocalInstall(packages []*installer.Package) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 DEPLOYING PACKAGES 🚀"))
	fmt.Println()

	total := len(packages)
	fmt.Printf("  %s %s\n",
		theme.InfoStyle.Render("Plan:"),
		theme.DimTextStyle.Render(fmt.Sprintf("%d packages to install", total)))
	fmt.Println()

	// Show what's coming
	for i, pkg := range packages {
		bullet := "○"
		if i == 0 {
			bullet = "●"
		}
		fmt.Printf("  %s %s %s  %s\n",
			theme.DimTextStyle.Render(fmt.Sprintf("%s %d.", bullet, i+1)),
			theme.HighlightStyle.Render(pkg.Name),
			theme.DimTextStyle.Render(pkg.Size),
			theme.DimTextStyle.Render("~"+pkg.EstimatedTime.String()))
	}
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────"))
	fmt.Println()

	var failures []installFailure
	var successes []*installer.Package
	totalStart := time.Now()

	// --from: skip already-completed steps
	startIdx := 0
	if installFrom > 0 {
		startIdx = installFrom - 1 // user-facing is 1-indexed
		if startIdx >= total {
			startIdx = 0
		}
		if startIdx > 0 {
			fmt.Printf("  %s Resuming from step %d (%s), skipping %d completed\n\n",
				theme.InfoStyle.Render("↳"),
				startIdx+1,
				packages[startIdx].Name,
				startIdx)
		}
	}

	for i, pkg := range packages {
		if i < startIdx {
			continue
		}
		remaining := total - i
		// Phased mode: pause and confirm
		if phased && i > 0 {
			fmt.Println()
			fmt.Printf("  %s  %s\n",
				theme.WarningStyle.Render("⏸"),
				theme.DimTextStyle.Render(fmt.Sprintf("%d of %d complete, %d remaining", i, total, remaining)))
			if len(failures) > 0 {
				fmt.Printf("  %s  %s\n",
					theme.ErrorStyle.Render("!"),
					theme.ErrorStyle.Render(fmt.Sprintf("%d failed so far", len(failures))))
			}
			fmt.Println()
			fmt.Printf("  %s ", theme.InfoStyle.Render("Next:"))
			fmt.Printf("%s — %s\n", theme.HighlightStyle.Render(pkg.Name), pkg.Description)
			fmt.Printf("  %s %s, estimated %s\n",
				theme.DimTextStyle.Render("     "),
				theme.DimTextStyle.Render(pkg.Size),
				theme.DimTextStyle.Render(pkg.EstimatedTime.String()))
			fmt.Println()
			fmt.Printf("  %s", theme.DimTextStyle.Render("Press Enter to install, or Ctrl+C to stop → "))
			bufio.NewReader(os.Stdin).ReadString('\n')
		}

		// Step header
		fmt.Printf("  %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("Step %d/%d", i+1, total)),
			theme.HighlightStyle.Render(pkg.Name))
		fmt.Printf("  %s  %s\n",
			theme.DimTextStyle.Render("       "),
			theme.DimTextStyle.Render(pkg.Description))
		fmt.Println()

		script, exists := installer.GetScript(pkg.ID)
		if !exists {
			fmt.Println(theme.ErrorStyle.Render("  ✗ Script not found for " + pkg.ID))
			failures = append(failures, installFailure{pkg: pkg, err: fmt.Errorf("script not found")})
			continue
		}

		// Execute script with indented output
		cmd := exec.Command("bash", "-c", script)
		cmd.Env = os.Environ()
		// Inject embedded HF token if not already set
		if os.Getenv("HF_TOKEN") == "" && hf.EmbeddedToken != "" {
			cmd.Env = append(cmd.Env, "HF_TOKEN="+hf.EmbeddedToken)
		}
		cmd.Stdout = newPrefixWriter(os.Stdout, "  │ ")
		cmd.Stderr = newPrefixWriter(os.Stderr, "  │ ")

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

		if err != nil {
			fmt.Println()
			fmt.Printf("  %s %s  %s\n",
				theme.ErrorStyle.Render("✗"),
				theme.ErrorStyle.Render(pkg.Name+" FAILED"),
				theme.DimTextStyle.Render(fmt.Sprintf("after %s", formatElapsed(elapsed))))
			fmt.Printf("  %s  %s\n",
				theme.DimTextStyle.Render("  "),
				theme.DimTextStyle.Render(err.Error()))
			fmt.Printf("  %s  %s\n",
				theme.DimTextStyle.Render("  "),
				theme.InfoStyle.Render(fmt.Sprintf("Resume from here: anime install --from %d --phased <packages...>", i+1)))
			fmt.Println()
			failures = append(failures, installFailure{pkg: pkg, err: err, elapsed: elapsed})
			continue
		}

		// Success
		fmt.Println()
		fmt.Printf("  %s %s  %s\n",
			theme.SuccessStyle.Render("✓"),
			theme.SuccessStyle.Render(pkg.Name+" done"),
			theme.DimTextStyle.Render(formatElapsed(elapsed)))
		fmt.Println()
		successes = append(successes, pkg)
	}

	// Final summary
	totalElapsed := time.Since(totalStart)
	fmt.Println(theme.DimTextStyle.Render("  ─────────────────────────────────────────"))
	fmt.Println()
	fmt.Printf("  %s  %s installed, %s failed, %s total\n",
		theme.InfoStyle.Render("Done:"),
		theme.SuccessStyle.Render(fmt.Sprintf("%d", len(successes))),
		theme.ErrorStyle.Render(fmt.Sprintf("%d", len(failures))),
		theme.DimTextStyle.Render(formatElapsed(totalElapsed)))
	fmt.Println()

	showInstallSummary(successes, failures, "")
}

func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// prefixWriter wraps an io.Writer and prepends a prefix to each line.
// This keeps script output visually nested under the install step.
type prefixWriter struct {
	w      *os.File
	prefix string
	atBOL  bool
}

func newPrefixWriter(w *os.File, prefix string) *prefixWriter {
	return &prefixWriter{w: w, prefix: prefix, atBOL: true}
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	written := 0
	for _, b := range p {
		if pw.atBOL {
			pw.w.WriteString(pw.prefix)
			pw.atBOL = false
		}
		n, err := pw.w.Write([]byte{b})
		written += n
		if err != nil {
			return written, err
		}
		if b == '\n' {
			pw.atBOL = true
		}
	}
	return len(p), nil
}

func runRemoteInstall(packages []*installer.Package) {
	if installServer == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Server name required for remote install"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
		fmt.Println(theme.HighlightStyle.Render("  anime install --remote --server <server-name> <packages...>"))
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
		fmt.Println(theme.DimTextStyle.Render("  anime install --remote -s lambda core pytorch"))
		fmt.Println(theme.DimTextStyle.Render("  anime install -r -s my-server ollama"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
		fmt.Println(theme.DimTextStyle.Render("  anime list                 # List available servers"))
		fmt.Println(theme.DimTextStyle.Render("  anime add <server-name>    # Add a new server"))
		fmt.Println()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error loading config: " + err.Error()))
		os.Exit(1)
	}

	server, err := cfg.GetServer(installServer)
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Server not found: " + installServer))
		fmt.Println()
		if len(cfg.Servers) > 0 {
			fmt.Println(theme.InfoStyle.Render("📋 Available servers:"))
			for _, s := range cfg.Servers {
				fmt.Println(theme.DimTextStyle.Render("  • " + s.Name))
			}
			fmt.Println()
		}
		fmt.Println(theme.InfoStyle.Render("💡 Options:"))
		fmt.Println(theme.DimTextStyle.Render("  anime add <server-name>    # Add a new server"))
		fmt.Println(theme.DimTextStyle.Render("  anime list                 # List all servers"))
		fmt.Println()
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

	var failures []installFailure
	var successes []*installer.Package

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
			failures = append(failures, installFailure{pkg: pkg, err: fmt.Errorf("script not found")})
			continue
		}

		start := time.Now()
		output, err := client.RunCommand(script)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  %s %s\n",
				theme.ErrorStyle.Render(theme.SymbolError+" FAILED"),
				theme.DimTextStyle.Render(fmt.Sprintf("(after %s)", elapsed.Round(time.Second))))
			fmt.Println()

			// Show error output briefly
			if len(output) > 0 {
				lines := strings.Split(output, "\n")
				if len(lines) > 5 {
					lines = lines[len(lines)-5:]
				}
				fmt.Println(theme.DimTextStyle.Render("  " + strings.Join(lines, "\n  ")))
				fmt.Println()
			}

			failures = append(failures, installFailure{pkg: pkg, err: err, elapsed: elapsed})
			// Continue with next package instead of exiting
			continue
		}

		// Show output (dimmed)
		if len(output) > 0 {
			fmt.Println(theme.DimTextStyle.Render("  " + strings.ReplaceAll(output, "\n", "\n  ")))
		}
		fmt.Printf("  %s %s\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess+" COMPLETE"),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s)", elapsed.Round(time.Second))))
		fmt.Println()
		successes = append(successes, pkg)
	}

	// Show summary
	showInstallSummary(successes, failures, installServer)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// showInstallSummary displays the final installation summary
func showInstallSummary(successes []*installer.Package, failures []installFailure, serverName string) {
	fmt.Println()

	// All succeeded
	if len(failures) == 0 {
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		if serverName != "" {
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ QUEST COMPLETE on %s! ✨", serverName)))
		} else {
			fmt.Println(theme.SuccessStyle.Render("  ✨ QUEST COMPLETE! All packages deployed! ✨"))
		}
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		fmt.Println()
		return
	}

	// Some or all failed
	fmt.Println(theme.WarningStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println(theme.WarningStyle.Render("  📊 INSTALLATION SUMMARY"))
	fmt.Println(theme.WarningStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()

	// Show successes
	if len(successes) > 0 {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Successfully installed (%d):", len(successes))))
		for _, pkg := range successes {
			fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("•"), theme.DimTextStyle.Render(pkg.Name))
		}
		fmt.Println()
	}

	// Show failures
	fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("✗ Failed to install (%d):", len(failures))))
	for _, f := range failures {
		fmt.Printf("  %s %s", theme.ErrorStyle.Render("•"), theme.HighlightStyle.Render(f.pkg.Name))
		if f.elapsed > 0 {
			fmt.Printf(" %s", theme.DimTextStyle.Render(fmt.Sprintf("(after %s)", f.elapsed.Round(time.Second))))
		}
		fmt.Println()
	}
	fmt.Println()

	// Help section
	fmt.Println(theme.InfoStyle.Render("💡 To retry failed packages:"))
	failedIDs := make([]string, len(failures))
	for i, f := range failures {
		failedIDs[i] = f.pkg.ID
	}
	retryCmd := "anime install " + strings.Join(failedIDs, " ")
	if serverName != "" {
		retryCmd = fmt.Sprintf("anime install --remote -s %s %s", serverName, strings.Join(failedIDs, " "))
	}
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(retryCmd))
	fmt.Println()

	fmt.Println(theme.DimTextStyle.Render("💡 For troubleshooting: anime doctor"))
	fmt.Println()

	// Exit with error code if there were failures
	os.Exit(1)
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
