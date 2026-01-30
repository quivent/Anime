package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	bootstrapDelta bool
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap <server-name>",
	Short: "Deploy anime CLI to a remote server",
	Long:  `Build and deploy the anime CLI to a Lambda server for remote management.

Use --delta for incremental updates (rsync) - only transfers changed bytes.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Server name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime bootstrap <server-name>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime bootstrap lambda-1"))
			fmt.Println(theme.DimTextStyle.Render("  anime bootstrap my-server"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 What it does:"))
			fmt.Println(theme.DimTextStyle.Render("  Builds anime CLI and deploys it to your remote server"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Related Commands:"))
			fmt.Println(theme.DimTextStyle.Render("  anime push   # Deploy current directory to server"))
			fmt.Println(theme.DimTextStyle.Render("  anime list   # List all servers"))
			fmt.Println()
			return fmt.Errorf("bootstrap requires a server name")
		}
		return nil
	},
	Run: runBootstrap,
}

func init() {
	bootstrapCmd.Flags().BoolVar(&bootstrapDelta, "delta", false, "Use rsync for incremental transfer (faster updates)")
	rootCmd.AddCommand(bootstrapCmd)
}

func runBootstrap(cmd *cobra.Command, args []string) {
	serverName := args[0]

	fmt.Println(theme.RenderBanner("🚀 ANIME BOOTSTRAP 🚀"))
	fmt.Println()

	// Load config to get server details
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error loading config: " + err.Error()))
		os.Exit(1)
	}

	server, err := cfg.GetServer(serverName)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Server not found: " + serverName))
		fmt.Println(theme.DimTextStyle.Render("\nAvailable servers:"))
		for _, s := range cfg.Servers {
			fmt.Println(theme.DimTextStyle.Render("  • " + s.Name))
		}
		os.Exit(1)
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📡 Target: %s@%s", server.User, server.Host)))
	fmt.Println()

	// Step 1: Build for ARM64 (Lambda GH200)
	fmt.Println(theme.InfoStyle.Render("[1/4] " + theme.SymbolBuild + " Building anime for ARM64/Linux..."))

	buildCmd := exec.Command("go", "build", "-o", "/tmp/anime-arm64")
	buildCmd.Dir = "/Users/joshkornreich/lambda"
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm64")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Build failed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Troubleshooting:"))
		fmt.Println(theme.DimTextStyle.Render("  • Ensure Go is installed: go version"))
		fmt.Println(theme.DimTextStyle.Render("  • Check source exists: ls /Users/joshkornreich/lambda"))
		fmt.Println(theme.DimTextStyle.Render("  • Try manual build: cd /Users/joshkornreich/lambda && go build"))
		fmt.Println()
		os.Exit(1)
	}
	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Build complete"))
	fmt.Println()

	// Step 2: Copy to server
	if bootstrapDelta {
		fmt.Println(theme.InfoStyle.Render("[2/4] " + theme.SymbolDeploy + " Deploying to server (delta)..."))

		// Use rsync for incremental transfer - only sends changed bytes
		rsyncArgs := []string{
			"-avz", "--progress", "--partial",
			"-e", fmt.Sprintf("ssh -i %s -o BatchMode=yes -o StrictHostKeyChecking=no", server.SSHKey),
			"/tmp/anime-arm64",
			fmt.Sprintf("%s@%s:/tmp/anime", server.User, server.Host),
		}
		rsyncCmd := exec.Command("rsync", rsyncArgs...)
		rsyncCmd.Stdout = os.Stdout
		rsyncCmd.Stderr = os.Stderr

		if err := rsyncCmd.Run(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Delta deploy failed"))
			fmt.Println(theme.DimTextStyle.Render("  Falling back to full copy..."))
			// Fall through to scp
		} else {
			fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Uploaded (delta)"))
			goto installStep
		}
	}

	fmt.Println(theme.InfoStyle.Render("[2/4] " + theme.SymbolDeploy + " Deploying to server..."))
	{
		scpArgs := []string{
			"-i", server.SSHKey,
			"-o", "BatchMode=yes",
			"-o", "StrictHostKeyChecking=no",
			"/tmp/anime-arm64",
			fmt.Sprintf("%s@%s:/tmp/anime", server.User, server.Host),
		}
		scpCmd := exec.Command("scp", scpArgs...)
		scpCmd.Stdout = os.Stdout
		scpCmd.Stderr = os.Stderr

		if err := scpCmd.Run(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Deploy failed"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Troubleshooting:"))
			fmt.Println(theme.DimTextStyle.Render("  • Check SSH key permissions: chmod 600 " + server.SSHKey))
			fmt.Println(theme.DimTextStyle.Render("  • Verify server is reachable: ssh -i " + server.SSHKey + " " + server.User + "@" + server.Host))
			fmt.Println(theme.DimTextStyle.Render("  • Try manual copy: scp /tmp/anime-arm64 " + server.User + "@" + server.Host + ":/tmp/anime"))
			fmt.Println()
			os.Exit(1)
		}
		fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Uploaded"))
	}

installStep:
	fmt.Println()

	// Step 3: Install on server
	fmt.Println(theme.InfoStyle.Render("[3/4] " + theme.SymbolConfig + " Installing on server..."))

	installScript := `
mkdir -p ~/.local/bin
mv /tmp/anime ~/.local/bin/
chmod +x ~/.local/bin/anime

# Function to add PATH if not already present
add_to_path() {
    local rc_file="$1"
    local path_line='export PATH="$HOME/.local/bin:$PATH"'

    touch "$rc_file"

    if ! grep -q "\.local/bin" "$rc_file" 2>/dev/null; then
        echo "" >> "$rc_file"
        echo "# Added by anime bootstrap" >> "$rc_file"
        echo "$path_line" >> "$rc_file"
    fi
}

# Add to appropriate shell config files
if [ -f "$HOME/.bashrc" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
    add_to_path "$HOME/.bashrc"
fi

if [ -f "$HOME/.zshrc" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
    add_to_path "$HOME/.zshrc"
fi

add_to_path "$HOME/.profile"

# Add system-wide PATH config
if [ -w /etc/profile.d ] 2>/dev/null || sudo -n true 2>/dev/null; then
    sudo tee /etc/profile.d/anime-path.sh > /dev/null <<'PATHEOF'
# Added by anime bootstrap - ensures ~/.local/bin is in PATH for all users
if [ -d "$HOME/.local/bin" ]; then
    case ":$PATH:" in
        *":$HOME/.local/bin:"*) ;;
        *) export PATH="$HOME/.local/bin:$PATH" ;;
    esac
fi
PATHEOF
    sudo chmod +x /etc/profile.d/anime-path.sh
fi

# Export for current session
export PATH="$HOME/.local/bin:$PATH"
`

	sshArgs := []string{
		"-i", server.SSHKey,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", server.User, server.Host),
		installScript,
	}
	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  " + theme.SymbolError + " Installation failed"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Troubleshooting:"))
		fmt.Println(theme.DimTextStyle.Render("  • The binary was uploaded but installation script failed"))
		fmt.Println(theme.DimTextStyle.Render("  • Try manual install:"))
		fmt.Println(theme.DimTextStyle.Render("    ssh " + server.User + "@" + server.Host + " 'mv /tmp/anime ~/.local/bin/'"))
		fmt.Println(theme.DimTextStyle.Render("    ssh " + server.User + "@" + server.Host + " 'chmod +x ~/.local/bin/anime'"))
		fmt.Println()
		os.Exit(1)
	}
	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Installed"))
	fmt.Println()

	// Step 4: Verify
	fmt.Println(theme.InfoStyle.Render("[4/4] " + theme.SymbolSparkle + " Verifying installation..."))

	verifyArgs := []string{
		"-i", server.SSHKey,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", server.User, server.Host),
		"~/.local/bin/anime tree | head -5",
	}
	verifyCmd := exec.Command("ssh", verifyArgs...)
	verifyCmd.Stdout = os.Stdout
	verifyCmd.Stderr = os.Stderr

	if err := verifyCmd.Run(); err != nil {
		fmt.Println(theme.WarningStyle.Render("  " + theme.SymbolWarning + " Verification had issues (may still work)"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Verified"))
	}
	fmt.Println()

	// Cleanup
	os.Remove("/tmp/anime-arm64")

	// Success banner
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println(theme.SuccessStyle.Render("  ✨ Bootstrap complete! ✨"))
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  SSH to server and run: anime"))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Or try: ssh %s@%s 'anime tree'", server.User, server.Host)))
	fmt.Println()
}
