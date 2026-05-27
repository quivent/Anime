package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmapi"
	"github.com/joshkornreich/anime/internal/mmcfg"
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
	t.Section("MATTERMOST")
	fmt.Println()
	fmt.Println("  " + t.Cyan.S("Mattermost Team Chat Management"))
	fmt.Println("  " + t.Dim("Setup / Users / Channels / Agents / Daemons"))
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
	fmt.Println("  " + t.Bold(t.Gold.S("Quick Start")))
	fmt.Println()
	for _, q := range quick {
		fmt.Printf("  %-46s%s\n",
			t.Gold.S(q.cmd),
			t.Dim(q.desc))
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
		t.Section("CONFIG")
		t.KV("file", mmcfg.Path())
		fmt.Println()
		data, _ := yaml.Marshal(cfg)
		fmt.Println(t.Dim(string(data)))
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
		t.Ok(t.Bold(t.Gold.S(args[0])) + " = " + t.Dim(args[1]))
		return nil
	},
}

var matrixConfigInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a fresh configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(mmcfg.Path()); err == nil {
			t.Warn("config already exists: " + mmcfg.Path())
			return nil
		}
		cfg := &mmcfg.Config{
			Server: mmcfg.ServerConfig{URL: "http://localhost:8065"},
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		t.Ok("config initialized: " + mmcfg.Path())
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
		t.Ok("sent  " + t.Dim(post.ID))
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
