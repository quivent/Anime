package cmd

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve <path>",
	Short: "Serve content on public IP at a random obscure port",
	Long: `Start an HTTP server to serve content from a directory on the Lambda server.

The server will be accessible on the public IP at a random high port (10000-65535)
for security through obscurity.

Examples:
  anime serve ~/outputs              # Serve outputs directory
  anime serve ~/ComfyUI/output       # Serve ComfyUI outputs
  anime serve . --port 12345         # Serve current dir on specific port`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "Specific port to use (default: random 10000-65535)")
}

func runServe(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		showServeHelp()
		return nil
	}

	servePath := args[0]

	// Load config to get lambda server
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get lambda target
	lambdaTarget := cfg.GetAlias("lambda")
	if lambdaTarget == "" {
		if server, err := cfg.GetServer("lambda"); err == nil {
			lambdaTarget = fmt.Sprintf("%s@%s", server.User, server.Host)
		}
	}

	if lambdaTarget == "" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Lambda server not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Set it up with:"))
		fmt.Println(theme.HighlightStyle.Render("  anime set lambda <server-ip>"))
		fmt.Println()
		return fmt.Errorf("lambda server not configured")
	}

	// Parse target
	var user, host string
	if strings.Contains(lambdaTarget, "@") {
		parts := strings.SplitN(lambdaTarget, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = lambdaTarget
	}

	// Generate random port if not specified
	if servePort == 0 {
		servePort = 10000 + rand.Intn(55535)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("🌐 ANIME SERVE 🌐"))
	fmt.Println()

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return fmt.Errorf("failed to connect to lambda server: %w", err)
	}
	defer sshClient.Close()

	fmt.Printf("  Server:       %s\n", theme.HighlightStyle.Render(lambdaTarget))
	fmt.Printf("  Serving:      %s\n", theme.HighlightStyle.Render(servePath))
	fmt.Printf("  Port:         %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", servePort)))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Starting HTTP server..."))
	fmt.Println()

	// Start Python HTTP server
	serveCmd := fmt.Sprintf(
		`cd %s && nohup python3 -m http.server %d --bind 0.0.0.0 > ~/serve.log 2>&1 & echo $! > ~/serve.pid && echo "Server started on port %d"`,
		servePath, servePort, servePort)

	output, err := sshClient.RunCommand(serveCmd)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("  ❌ Failed to start server"))
		fmt.Println(theme.DimTextStyle.Render(output))
		return fmt.Errorf("serve failed: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ Server started successfully!"))
	fmt.Println()

	// Show access information
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🌐 Access Information"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	url := fmt.Sprintf("http://%s:%d", host, servePort)
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(url))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  💡 Tip: This server is publicly accessible on the internet"))
	fmt.Println(theme.DimTextStyle.Render("      The random port provides security through obscurity"))
	fmt.Println(theme.DimTextStyle.Render("      Share this URL to allow others to download files"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🎯 Management Commands"))
	fmt.Printf("  Stop server:    %s\n", theme.HighlightStyle.Render("kill $(cat ~/serve.pid)"))
	fmt.Printf("  View logs:      %s\n", theme.HighlightStyle.Render("tail -f ~/serve.log"))
	fmt.Printf("  Check status:   %s\n", theme.HighlightStyle.Render("ps aux | grep http.server"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("⚠️  Security Notes"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("  • Files are publicly accessible to anyone with the URL"))
	fmt.Println(theme.WarningStyle.Render("  • Consider using this only for temporary sharing"))
	fmt.Println(theme.WarningStyle.Render("  • Stop the server when done to close the port"))
	fmt.Println()

	return nil
}

func showServeHelp() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ ANIME SERVE ⚡"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🌐 Serve content on your Lambda server's public IP"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📖 Description"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Start an HTTP server to serve files from a directory."))
	fmt.Println(theme.DimTextStyle.Render("  Uses a random high port (10000-65535) for security."))
	fmt.Println(theme.DimTextStyle.Render("  Perfect for sharing generated images, videos, or datasets."))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Examples"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	examples := []struct {
		cmd  string
		desc string
	}{
		{"anime serve ~/ComfyUI/output", "Serve ComfyUI generated images/videos"},
		{"anime serve ~/outputs", "Serve your outputs directory"},
		{"anime serve . --port 12345", "Serve current directory on specific port"},
		{"anime serve ~/datasets --port 20000", "Serve datasets on port 20000"},
	}

	for _, ex := range examples {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ "+ex.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(ex.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎯 Common Use Cases"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s %s\n",
		"📹",
		theme.DimTextStyle.Render("Share generated videos from Mochi or CogVideoX"))
	fmt.Printf("  %s %s\n",
		"🖼️ ",
		theme.DimTextStyle.Render("Share images from ComfyUI or Stable Diffusion"))
	fmt.Printf("  %s %s\n",
		"📊",
		theme.DimTextStyle.Render("Share training checkpoints or model weights"))
	fmt.Printf("  %s %s\n",
		"💾",
		theme.DimTextStyle.Render("Share datasets for collaboration"))
	fmt.Println()
}
