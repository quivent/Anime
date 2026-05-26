package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/joshkornreich/anime/internal/mmapi"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ─── root ───────────────────────────────────────────────────────────────────

var matrixCmd = &cobra.Command{
	Use:   "matrix",
	Short: "Mattermost team chat management",
	Long: `Full-featured Mattermost management — setup, connect, users, channels,
agents, daemons, messaging.`,
	Run: matrixWelcome,
}

func matrixWelcome(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("MATTERMOST"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Mattermost Team Chat Management"))
	fmt.Println(theme.DimTextStyle.Render("  Setup / Users / Channels / Agents / Daemons"))
	fmt.Println()

	quick := []struct{ cmd, desc string }{
		{"anime matrix setup", "Deploy a Mattermost server"},
		{"anime matrix connect --url <url>", "Connect to an existing server"},
		{"anime matrix status", "Check server health"},
		{"anime matrix users add <name>", "Create a user"},
		{"anime matrix channels create <name>", "Create a channel"},
		{"anime matrix agents spawn <name>", "Spawn a Claude Code bot"},
		{"anime matrix send <channel> \"msg\"", "Send a message"},
		{"anime matrix watch <channel>", "Live-tail channel messages"},
	}
	fmt.Println(theme.GlowStyle.Render("Quick Start:"))
	fmt.Println()
	for _, q := range quick {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-44s", q.cmd)),
			theme.DimTextStyle.Render(q.desc))
	}
	fmt.Println()
}

// ─── config ─────────────────────────────────────────────────────────────────

var matrixConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := mmcfg.Load()
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(theme.RenderBanner("CONFIG"))
		fmt.Println()
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("File:"), theme.DimTextStyle.Render(mmcfg.Path()))
		fmt.Println()
		data, _ := yaml.Marshal(cfg)
		fmt.Println(theme.DimTextStyle.Render(string(data)))
		return nil
	},
}

var matrixConfigSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := mmcfg.Load()
		if err != nil {
			return err
		}
		switch args[0] {
		case "server.url":
			cfg.Server.URL = args[1]
		case "server.token":
			cfg.Server.Token = args[1]
		case "server.username":
			cfg.Server.Username = args[1]
		case "server.team_id":
			cfg.Server.TeamID = args[1]
		case "server.team_name":
			cfg.Server.TeamName = args[1]
		case "install.data_dir":
			cfg.Install.DataDir = args[1]
		default:
			return fmt.Errorf("unknown key: %s", args[0])
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("  %s %s = %s\n", theme.SymbolSuccess,
			theme.HighlightStyle.Render(args[0]), theme.DimTextStyle.Render(args[1]))
		return nil
	},
}

var matrixConfigInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a fresh configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(mmcfg.Path()); err == nil {
			fmt.Printf("  %s\n", theme.WarningStyle.Render("Config already exists: "+mmcfg.Path()))
			return nil
		}
		cfg := &mmcfg.Config{
			Server: mmcfg.ServerConfig{URL: "http://localhost:8065"},
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess,
			theme.SuccessStyle.Render("Config initialized: "+mmcfg.Path()))
		return nil
	},
}

// ─── send ───────────────────────────────────────────────────────────────────

var matrixSendAsUser string

var matrixSendCmd = &cobra.Command{
	Use:   "send <channel-id> <message>",
	Short: "Send a message to a channel",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelID := args[0]
		message := strings.Join(args[1:], " ")
		cfg, err := mmcfg.Load()
		if err != nil {
			return err
		}
		token := cfg.Server.Token
		if matrixSendAsUser != "" {
			agent := cfg.GetAgent(matrixSendAsUser)
			if agent == nil {
				return fmt.Errorf("agent %q not found", matrixSendAsUser)
			}
			token = agent.Token
		}
		client := mmClient(cfg.Server.URL, token)
		post, err := client.CreatePost(channelID, message, nil)
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s  %s\n",
			theme.SymbolStar, theme.SuccessStyle.Render("Sent"),
			theme.DimTextStyle.Render(post.ID))
		return nil
	},
}

// ─── helpers ────────────────────────────────────────────────────────────────

func matrixGeneratePassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "changeme"
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i, v := range b {
		result[i] = charset[int(v)%len(charset)]
	}
	return string(result)
}

func matrixRunBash(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func matrixSeparator() string {
	return theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func matrixIsAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func mmClient(url, token string) *mmapi.Client {
	return mmapi.NewClient(url, token)
}

// ─── init ───────────────────────────────────────────────────────────────────

func init() {
	matrixSendCmd.Flags().StringVar(&matrixSendAsUser, "as", "", "Send as a specific agent")

	matrixConfigCmd.AddCommand(matrixConfigShowCmd)
	matrixConfigCmd.AddCommand(matrixConfigSetCmd)
	matrixConfigCmd.AddCommand(matrixConfigInitCmd)

	matrixCmd.AddCommand(matrixConfigCmd)
	matrixCmd.AddCommand(matrixSendCmd)

	rootCmd.AddCommand(matrixCmd)
}
