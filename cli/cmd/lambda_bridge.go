package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/spf13/cobra"
)

var forceBootstrap bool

var lambdaBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Deploy the lambda CLI to the remote server",
	Long: `Upload the cross-compiled lambda CLI to the configured Lambda server.
Detects remote architecture (arm64/amd64) automatically.

Examples:
  anime lambda bootstrap          # Deploy lambda CLI
  anime lambda bootstrap --force  # Overwrite existing`,
	RunE: runLambdaBootstrap,
}

var lambdaRunCmd = &cobra.Command{
	Use:                "run [command] [args...]",
	Short:              "Run a lambda CLI command on the remote server (flag passthrough)",
	DisableFlagParsing: true,
	RunE:               runLambdaRun,
}

func init() {
	lambdaBootstrapCmd.Flags().BoolVarP(&forceBootstrap, "force", "f", false, "Overwrite existing binary")
	lambdaCmd.AddCommand(lambdaBootstrapCmd)
	lambdaCmd.AddCommand(lambdaRunCmd)
}

// resolveLambdaTarget returns SSH user and host for the configured lambda server.
func resolveLambdaTarget() (string, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}
	target := cfg.GetAlias("lambda")
	if target == "" {
		if srv, err := cfg.GetServer("lambda"); err == nil {
			target = fmt.Sprintf("%s@%s", srv.User, srv.Host)
		}
	}
	if target == "" {
		return "", "", fmt.Errorf("lambda server not configured — run: anime set lambda <ip>")
	}
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		return parts[0], parts[1], nil
	}
	return "ubuntu", target, nil
}

func runLambdaBootstrap(cmd *cobra.Command, args []string) error {
	user, host, err := resolveLambdaTarget()
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("SSH failed: %w", err)
	}
	defer client.Close()

	// Detect remote architecture
	archOut, err := client.RunCommand("uname -m")
	if err != nil {
		return fmt.Errorf("arch detection failed: %w", err)
	}
	arch := strings.TrimSpace(archOut)

	var binaryName, goarch string
	switch arch {
	case "aarch64":
		binaryName = "lambda-linux-arm64"
		goarch = "arm64"
	case "x86_64":
		binaryName = "lambda-linux-amd64"
		goarch = "amd64"
	default:
		return fmt.Errorf("unsupported arch: %s", arch)
	}

	// Locate local binary
	home, _ := os.UserHomeDir()
	localPath := filepath.Join(home, "lambda", "cli", binaryName)
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("not found: %s\n  build: cd ~/lambda/cli && GOOS=linux GOARCH=%s go build -o %s .", localPath, goarch, binaryName)
	}

	// Resolve remote home for absolute paths
	remoteHome, err := client.RunCommand("echo $HOME")
	if err != nil {
		return fmt.Errorf("cannot resolve remote $HOME: %w", err)
	}
	remoteBin := strings.TrimSpace(remoteHome) + "/.local/bin"
	remotePath := remoteBin + "/lambda"

	// Check existing deployment
	if !forceBootstrap {
		checkOut, _ := client.RunCommand("test -f " + remotePath + " && echo exists")
		if strings.TrimSpace(checkOut) == "exists" {
			fmt.Printf("  lambda already at %s:%s — use --force to overwrite\n", host, remotePath)
			return nil
		}
	}

	client.RunCommand("mkdir -p " + remoteBin)

	sizeMB := float64(info.Size()) / (1024 * 1024)
	fmt.Printf("  %s (%.1fMB) → %s@%s:%s\n", binaryName, sizeMB, user, host, remotePath)

	if err := client.UploadFile(localPath, remotePath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	if err := client.MakeExecutable(remotePath); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	// Ensure ~/.local/bin is on PATH
	client.RunCommand(`grep -q '\.local/bin' ~/.bashrc 2>/dev/null || echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc`)

	// Verify
	out, _ := client.RunCommand(remotePath + " version 2>/dev/null || echo deployed")
	fmt.Printf("  ✓ %s\n", strings.TrimSpace(out))
	return nil
}

func runLambdaRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: anime lambda run <command> [args...]\n  example: anime lambda run specs --json")
	}
	return runLambdaRemote(args)
}

// runLambdaOrProxy shows help with no args, proxies unrecognized subcommands to the remote lambda CLI.
func runLambdaOrProxy(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		runLambdaHelp(cmd, args)
		return nil
	}
	return runLambdaRemote(args)
}

// runLambdaRemote executes a command on the remote lambda CLI via SSH.
func runLambdaRemote(args []string) error {
	user, host, err := resolveLambdaTarget()
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("SSH to %s failed: %w", host, err)
	}
	defer client.Close()

	remoteCmd := "$HOME/.local/bin/lambda " + strings.Join(args, " ")
	out, err := client.RunCommand(remoteCmd)
	if out != "" {
		fmt.Print(out)
	}
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(out, "not found") {
			return fmt.Errorf("lambda CLI not deployed — run: anime lambda bootstrap")
		}
		return err
	}
	return nil
}
