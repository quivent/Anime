package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap <server-name>",
	Short: "Deploy anime CLI to a remote server",
	Long:  "Build and deploy the anime CLI to a Lambda server for remote management",
	Args:  cobra.ExactArgs(1),
	Run:   runBootstrap,
}

func init() {
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
		os.Exit(1)
	}
	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Build complete"))
	fmt.Println()

	// Step 2: Copy to server
	fmt.Println(theme.InfoStyle.Render("[2/4] " + theme.SymbolDeploy + " Deploying to server..."))

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
		os.Exit(1)
	}
	fmt.Println(theme.SuccessStyle.Render("  " + theme.SymbolSuccess + " Uploaded"))
	fmt.Println()

	// Step 3: Install on server
	fmt.Println(theme.InfoStyle.Render("[3/4] " + theme.SymbolConfig + " Installing on server..."))

	installScript := `
mkdir -p ~/.local/bin
mv /tmp/anime ~/.local/bin/
chmod +x ~/.local/bin/anime
if ! grep -q '.local/bin' ~/.bashrc; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
fi
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
