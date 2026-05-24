package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

var (
	unpackMinimal bool
	unpackFull    bool
	unpackRemote  bool
	unpackServer  string
	unpackDryRun  bool
	unpackYes     bool
	unpackUser    string
	unpackRepos   []string
)

var unpackCmd = &cobra.Command{
	Use:   "unpack",
	Short: "Bootstrap a server with dev essentials (no GPU)",
	Long: `Unpack bootstraps a server with everything needed for development — no GPU required.
For GPU/inference setup, use 'anime wizard' instead.

Two tiers:

  minimal   core + python + uv                         (~6 min, ~1GB)
  default   core + python + uv + nodejs + go + rust +  (~8 min, ~2GB)
            gh + claude
            (parallelized — independent packages install simultaneously)

Use --user to create a system user and --repo to clone GitHub repos.

Examples:
  anime unpack                                  # Default server setup
  anime unpack --minimal                        # Just core + python + uv
  anime unpack --user deploy                    # Create 'deploy' user too
  anime unpack --repo org/repo --repo org/repo2 # Clone repos after setup
  anime unpack -r -s mybox --user app           # Remote server bootstrap
  anime unpack --dry-run                        # Preview what would happen`,
	Run: runUnpack,
}

func init() {
	unpackCmd.Flags().BoolVar(&unpackMinimal, "minimal", false, "Install only core + python + uv")
	// --full is accepted but identical to default (all server essentials)
	unpackCmd.Flags().BoolVar(&unpackFull, "full", false, "Full server dev environment (same as default)")
	unpackCmd.Flags().BoolVarP(&unpackRemote, "remote", "r", false, "Install on remote server via SSH")
	unpackCmd.Flags().StringVarP(&unpackServer, "server", "s", "", "Server name for remote install")
	unpackCmd.Flags().BoolVar(&unpackDryRun, "dry-run", false, "Show what would be installed without doing it")
	unpackCmd.Flags().BoolVarP(&unpackYes, "yes", "y", false, "Skip confirmation prompt")
	unpackCmd.Flags().StringVar(&unpackUser, "user", "", "Create a system user with sudo access")
	unpackCmd.Flags().StringSliceVar(&unpackRepos, "repo", nil, "GitHub repos to clone after setup (e.g. org/repo)")
	rootCmd.AddCommand(unpackCmd)
}

func getUnpackPackages() []string {
	if unpackMinimal {
		return []string{"core", "python", "uv"}
	}
	// default and full are the same — everything
	return []string{
		"core", "python", "uv", "nodejs", "go", "rust",
		"gh", "claude",
	}
}

func unpackTierName() string {
	if unpackMinimal {
		return "MINIMAL"
	}
	return "SERVER"
}

func runUnpack(cmd *cobra.Command, args []string) {
	tierName := unpackTierName()

	// Validate inputs
	if unpackUser != "" {
		if err := validate.Username(unpackUser); err != nil {
			fmt.Println(theme.ErrorStyle.Render("❌ " + err.Error()))
			os.Exit(1)
		}
	}
	for _, repo := range unpackRepos {
		if err := validate.RepoSlug(repo); err != nil {
			fmt.Println(theme.ErrorStyle.Render("❌ " + err.Error()))
			os.Exit(1)
		}
	}

	fmt.Println(theme.RenderBanner("📦 UNPACK — " + tierName + " 📦"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🌸 Server bootstrap — everything to run a dev server"))
	fmt.Println(theme.DimTextStyle.Render("   Skips packages already installed. Parallelizes where possible."))
	fmt.Println(theme.DimTextStyle.Render("   For GPU/inference setup, use: anime wizard"))
	fmt.Println()

	packageIDs := getUnpackPackages()

	// Resolve dependencies
	resolved, err := installer.ResolveDependencies(packageIDs)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error resolving dependencies: " + err.Error()))
		os.Exit(1)
	}

	// Determine which are already installed
	var sshClient *ssh.Client
	if unpackRemote {
		sshClient = unpackConnectRemote()
		if sshClient == nil {
			return
		}
		defer sshClient.Close()
	}

	var toInstall []*installer.Package
	var alreadyInstalledNames []string
	var alreadyInstalledIDs []string

	for _, pkg := range resolved {
		isInstalled := false
		if unpackRemote && sshClient != nil {
			isInstalled = checkPackageInstalled(sshClient, pkg.ID)
		} else {
			isInstalled = checkPackageInstalledLocal(pkg.ID)
		}

		if isInstalled {
			alreadyInstalledNames = append(alreadyInstalledNames, pkg.Name)
			alreadyInstalledIDs = append(alreadyInstalledIDs, pkg.ID)
		} else {
			toInstall = append(toInstall, pkg)
		}
	}

	// Show what's already installed
	if len(alreadyInstalledNames) > 0 {
		fmt.Println(theme.SuccessStyle.Render("  Already installed (skipping):"))
		for _, name := range alreadyInstalledNames {
			fmt.Printf("    %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), theme.DimTextStyle.Render(name))
		}
		fmt.Println()
	}

	// Show extra steps
	hasExtras := unpackUser != "" || len(unpackRepos) > 0
	if hasExtras {
		fmt.Println(theme.InfoStyle.Render("  Post-install steps:"))
		if unpackUser != "" {
			fmt.Printf("    %s Create user %s with sudo access\n",
				theme.SymbolBolt, theme.HighlightStyle.Render(unpackUser))
		}
		fmt.Printf("    %s Generate SSH key pair\n", theme.SymbolBolt)
		for _, repo := range unpackRepos {
			fmt.Printf("    %s Clone %s\n",
				theme.SymbolBolt, theme.HighlightStyle.Render(repo))
		}
		fmt.Println()
	}

	// Nothing to do?
	if len(toInstall) == 0 && !hasExtras {
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		fmt.Println(theme.SuccessStyle.Render("  ✨ Everything is already unpacked! ✨"))
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		fmt.Println()
		return
	}

	// Show install plan
	var totalTime time.Duration
	for i, pkg := range toInstall {
		marker := theme.SymbolBranch
		if i == len(toInstall)-1 {
			marker = theme.SymbolLastBranch
		}

		badge := ""
		badgeStyle := theme.DimTextStyle
		if contains(packageIDs, pkg.ID) {
			badge = " " + theme.SymbolStar + " SELECTED"
			badgeStyle = theme.SuccessStyle
		} else {
			badge = " " + theme.SymbolBolt + " auto-included"
			badgeStyle = theme.DimTextStyle
		}

		emoji := theme.SymbolSakura
		switch pkg.Category {
		case "Foundation":
			emoji = "🏗️"
		case "Application":
			emoji = "🎯"
		case "Runtime":
			emoji = "📦"
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

	// Show wave plan
	waves := buildWaves(toInstall, alreadyInstalledIDs)
	if len(waves) > 1 || (len(waves) == 1 && len(waves[0]) > 1) {
		fmt.Println(theme.InfoStyle.Render("⚡ Parallel execution plan:"))
		for i, wave := range waves {
			names := make([]string, len(wave))
			for j, p := range wave {
				names[j] = p.ID
			}
			if len(wave) > 1 {
				fmt.Printf("    Wave %d: %s %s\n", i+1,
					theme.HighlightStyle.Render(strings.Join(names, ", ")),
					theme.DimTextStyle.Render("(parallel)"))
			} else {
				fmt.Printf("    Wave %d: %s\n", i+1,
					theme.HighlightStyle.Render(names[0]))
			}
		}
		fmt.Println()
	}

	summary := fmt.Sprintf("📦 %d packages in %d waves  |  ⏱️  Est. wall time: ~%d min",
		len(toInstall), len(waves), estimateWallMinutes(waves))
	fmt.Println(theme.InfoStyle.Render(summary))
	fmt.Println()

	if unpackDryRun {
		fmt.Println(theme.WarningStyle.Render("✨ Dry run complete — no changes made"))
		return
	}

	// Confirm
	if !unpackYes {
		fmt.Print("Proceed with unpack? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "n" || response == "no" {
			fmt.Println("Unpack cancelled")
			return
		}
	}

	// Execute package installs
	if len(toInstall) > 0 {
		if unpackRemote && sshClient != nil {
			runUnpackRemote(sshClient, toInstall, waves)
		} else {
			runUnpackLocal(toInstall, waves)
		}
	}

	// Post-install: create user, generate SSH keys, clone repos
	runUnpackExtras(sshClient)
}

// estimateWallMinutes calculates wall-clock time with parallelization
func estimateWallMinutes(waves [][]*installer.Package) int {
	total := time.Duration(0)
	for _, wave := range waves {
		longest := time.Duration(0)
		for _, pkg := range wave {
			if pkg.EstimatedTime > longest {
				longest = pkg.EstimatedTime
			}
		}
		total += longest
	}
	mins := int(total.Minutes())
	if mins < 1 {
		mins = 1
	}
	return mins
}

func unpackConnectRemote() *ssh.Client {
	if unpackServer == "" {
		cfg, err := config.Load()
		if err == nil {
			lambdaTarget := cfg.GetAlias("lambda")
			if lambdaTarget != "" {
				unpackServer = "lambda"
			}
		}
		if unpackServer == "" {
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required for remote unpack"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime unpack --remote --server <server-name>"))
			fmt.Println(theme.DimTextStyle.Render("  anime unpack -r -s lambda"))
			fmt.Println()
			return nil
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error loading config: " + err.Error()))
		return nil
	}

	var user, host, sshKey string

	target := cfg.GetAlias(unpackServer)
	if target != "" {
		if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			user = parts[0]
			host = parts[1]
		} else {
			user = "ubuntu"
			host = target
		}
	} else {
		server, err := cfg.GetServer(unpackServer)
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render("❌ Server not found: " + unpackServer))
			return nil
		}
		user = server.User
		host = server.Host
		sshKey = server.SSHKey
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("🌐 Connecting to %s (%s@%s)...", unpackServer, user, host)))

	client, err := ssh.NewClient(host, user, sshKey)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Connection failed: " + err.Error()))
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Connected!"))
	fmt.Println()
	return client
}

// buildWaves groups packages into parallel waves based on dependencies.
// alreadyInstalled are package IDs that were filtered out (already present).
func buildWaves(packages []*installer.Package, alreadyInstalled []string) [][]*installer.Package {
	installed := make(map[string]bool)
	// Pre-seed with packages that are already installed so deps are satisfied
	for _, id := range alreadyInstalled {
		installed[id] = true
	}
	remaining := make([]*installer.Package, len(packages))
	copy(remaining, packages)

	var waves [][]*installer.Package
	for len(remaining) > 0 {
		var wave []*installer.Package
		var next []*installer.Package
		for _, pkg := range remaining {
			ready := true
			for _, dep := range pkg.Dependencies {
				if !installed[dep] {
					ready = false
					break
				}
			}
			if ready {
				wave = append(wave, pkg)
			} else {
				next = append(next, pkg)
			}
		}
		if len(wave) == 0 {
			wave = next
			next = nil
		}
		waves = append(waves, wave)
		for _, pkg := range wave {
			installed[pkg.ID] = true
		}
		remaining = next
	}
	return waves
}

type installResult struct {
	pkg     *installer.Package
	err     error
	output  string
	elapsed time.Duration
}

func runUnpackLocal(packages []*installer.Package, waves [][]*installer.Package) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 UNPACKING 🚀"))
	fmt.Println()

	totalStart := time.Now()
	failed := 0

	// Local installs run sequentially — parallel apt/dpkg causes lock conflicts
	for i, pkg := range packages {
		fmt.Printf("[%s] %s %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%d/%d", i+1, len(packages))),
			theme.SymbolLoading,
			theme.HighlightStyle.Render("Installing "+pkg.Name))

		script, exists := installer.GetScript(pkg.ID)
		if !exists {
			fmt.Printf("  %s Script not found for %s\n",
				theme.ErrorStyle.Render(theme.SymbolError), pkg.ID)
			failed++
			fmt.Println()
			continue
		}

		cmd := exec.Command("bash", "-c", script)
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout

		start := time.Now()
		if err := cmd.Start(); err != nil {
			fmt.Printf("  %s Failed to start: %s\n",
				theme.ErrorStyle.Render(theme.SymbolError), err)
			failed++
			fmt.Println()
			continue
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "==>") {
				fmt.Printf("  %s\n", theme.InfoStyle.Render(line))
			} else if line != "" {
				fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
			}
		}

		err := cmd.Wait()
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  %s %s %s\n",
				theme.ErrorStyle.Render(theme.SymbolError+" FAILED"),
				theme.HighlightStyle.Render(pkg.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", elapsed.Round(time.Second))))
			fmt.Printf("  %s %s\n\n",
				theme.ErrorStyle.Render("  Reason:"),
				theme.DimTextStyle.Render(err.Error()))
			failed++
		} else {
			fmt.Printf("  %s %s %s\n\n",
				theme.SuccessStyle.Render(theme.SymbolSuccess+" COMPLETE"),
				theme.HighlightStyle.Render(pkg.Name),
				theme.DimTextStyle.Render(fmt.Sprintf("(%s)", elapsed.Round(time.Second))))
		}
	}

	totalElapsed := time.Since(totalStart)
	unpackVictory(len(packages), failed, totalElapsed)
}

func runUnpackRemote(client *ssh.Client, packages []*installer.Package, waves [][]*installer.Package) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🚀 REMOTE UNPACK 🚀"))
	fmt.Println()

	totalStart := time.Now()
	completed := 0
	failed := 0

	for waveIdx, wave := range waves {
		if len(wave) == 1 {
			fmt.Printf("%s %s\n",
				theme.InfoStyle.Render(fmt.Sprintf("  Wave %d/%d", waveIdx+1, len(waves))),
				theme.HighlightStyle.Render(wave[0].Name))
		} else {
			names := make([]string, len(wave))
			for i, p := range wave {
				names[i] = p.Name
			}
			fmt.Printf("%s %s %s\n",
				theme.InfoStyle.Render(fmt.Sprintf("  Wave %d/%d", waveIdx+1, len(waves))),
				theme.HighlightStyle.Render(strings.Join(names, " + ")),
				theme.DimTextStyle.Render("(parallel)"))
		}
		fmt.Println()

		for _, pkg := range wave {
			fmt.Printf("  %s %s\n",
				theme.SymbolLoading,
				theme.DimTextStyle.Render("Starting "+pkg.Name+"..."))
		}
		fmt.Println()

		results := make([]installResult, len(wave))
		var wg sync.WaitGroup

		for i, pkg := range wave {
			wg.Add(1)
			go func(idx int, p *installer.Package) {
				defer wg.Done()
				script, exists := installer.GetScript(p.ID)
				if !exists {
					results[idx] = installResult{pkg: p, err: fmt.Errorf("script not found")}
					return
				}
				start := time.Now()
				output, err := client.RunCommand(script)
				results[idx] = installResult{pkg: p, err: err, output: output, elapsed: time.Since(start)}
			}(i, pkg)
		}

		// Ticker for long-running waves
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			elapsed := 0
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					elapsed += 10
					fmt.Printf("  %s %s\n",
						theme.DimTextStyle.Render("⏳"),
						theme.DimTextStyle.Render(fmt.Sprintf("  %ds elapsed, installing...", elapsed)))
				}
			}
		}()

		wg.Wait()
		close(done)

		for _, res := range results {
			completed++
			if res.err != nil {
				failed++
				fmt.Printf("  %s %s %s\n",
					theme.ErrorStyle.Render(theme.SymbolError+" FAILED"),
					theme.HighlightStyle.Render(res.pkg.Name),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s)", res.elapsed.Round(time.Second))))
				if len(res.output) > 0 {
					lines := strings.Split(strings.TrimSpace(res.output), "\n")
					start := 0
					if len(lines) > 10 {
						start = len(lines) - 10
						fmt.Printf("  %s\n", theme.DimTextStyle.Render("  ... (showing last 10 lines)"))
					}
					for _, line := range lines[start:] {
						fmt.Printf("  %s\n", theme.DimTextStyle.Render("  "+line))
					}
				}
				fmt.Println()
			} else {
				fmt.Printf("  %s %s %s\n",
					theme.SuccessStyle.Render(theme.SymbolSuccess+" COMPLETE"),
					theme.HighlightStyle.Render(res.pkg.Name),
					theme.DimTextStyle.Render(fmt.Sprintf("(%s)", res.elapsed.Round(time.Second))))
			}
		}
		fmt.Println()
	}

	totalElapsed := time.Since(totalStart)
	unpackVictory(len(packages), failed, totalElapsed)
}

func runUnpackExtras(sshClient *ssh.Client) {
	hasExtras := unpackUser != "" || len(unpackRepos) > 0
	if !hasExtras {
		return
	}

	fmt.Println(theme.RenderBanner("🔧 POST-INSTALL 🔧"))
	fmt.Println()

	runCmd := func(description, script string) error {
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.HighlightStyle.Render(description))
		if unpackRemote && sshClient != nil {
			output, err := sshClient.RunCommand(script)
			if err != nil {
				fmt.Printf("  %s %s\n", theme.ErrorStyle.Render(theme.SymbolError+" FAILED"), theme.DimTextStyle.Render(output))
				return err
			}
			if len(output) > 0 {
				for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
					fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
				}
			}
		} else {
			cmd := exec.Command("bash", "-c", script)
			stdout, _ := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout
			if err := cmd.Start(); err != nil {
				fmt.Printf("  %s\n", theme.ErrorStyle.Render(theme.SymbolError+" FAILED: "+err.Error()))
				return err
			}
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Printf("  %s\n", theme.DimTextStyle.Render(scanner.Text()))
			}
			if err := cmd.Wait(); err != nil {
				fmt.Printf("  %s\n", theme.ErrorStyle.Render(theme.SymbolError+" FAILED"))
				return err
			}
		}
		fmt.Printf("  %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess+" Done"))
		fmt.Println()
		return nil
	}

	// Create user
	if unpackUser != "" {
		script := fmt.Sprintf(`#!/bin/bash
set -e
if id "%s" &>/dev/null; then
    echo "User %s already exists"
else
    sudo useradd -m -s /bin/bash "%s"
    sudo usermod -aG sudo "%s"
    echo "Created user %s with sudo access"
fi
`, unpackUser, unpackUser, unpackUser, unpackUser, unpackUser)
		if err := runCmd(fmt.Sprintf("Creating user '%s'", unpackUser), script); err != nil {
			fmt.Println(theme.WarningStyle.Render("  Continuing despite user creation failure..."))
		}
	}

	// Generate SSH keys
	sshKeyScript := `#!/bin/bash
set -e
if [ -f ~/.ssh/id_ed25519 ]; then
    echo "SSH key already exists: ~/.ssh/id_ed25519"
    echo "Public key:"
    cat ~/.ssh/id_ed25519.pub
else
    ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -N "" -C "$(whoami)@$(hostname)"
    echo "SSH key generated: ~/.ssh/id_ed25519"
    echo "Public key:"
    cat ~/.ssh/id_ed25519.pub
fi
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_ed25519
chmod 644 ~/.ssh/id_ed25519.pub
`
	if err := runCmd("Generating SSH key pair (ed25519)", sshKeyScript); err != nil {
		fmt.Println(theme.WarningStyle.Render("  Continuing despite SSH key generation failure..."))
	}

	// Clone repos
	for _, repo := range unpackRepos {
		repoURL := repo
		if !strings.Contains(repo, "://") && !strings.HasPrefix(repo, "git@") {
			repoURL = "https://github.com/" + repo + ".git"
		}
		script := fmt.Sprintf(`#!/bin/bash
set -e
REPO_NAME=$(basename "%s" .git)
if [ -d "$HOME/$REPO_NAME" ]; then
    echo "$REPO_NAME already cloned in ~/$REPO_NAME"
else
    git clone "%s" "$HOME/$REPO_NAME"
    echo "Cloned to ~/$REPO_NAME"
fi
`, repoURL, repoURL)
		if err := runCmd(fmt.Sprintf("Cloning %s", repo), script); err != nil {
			fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  Failed to clone %s, continuing...", repo)))
		}
	}
}

func unpackVictory(total, failures int, elapsed time.Duration) {
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	if failures > 0 {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  ⚡ UNPACKED %d/%d packages in %s (%d failed)",
			total-failures, total, elapsed.Round(time.Second), failures)))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ UNPACKED! %d packages in %s ✨",
			total, elapsed.Round(time.Second))))
	}
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🎯 What's next:"))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("anime wizard"), theme.DimTextStyle.Render("GPU/inference setup wizard"))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("anime packages"), theme.DimTextStyle.Render("Browse all available packages"))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("anime install <pkg>"), theme.DimTextStyle.Render("Install additional packages"))
	fmt.Println()
}
