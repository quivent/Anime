package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/synapse"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ─── root ───────────────────────────────────────────────────────────────────

var matrixCmd = &cobra.Command{
	Use:   "matrix",
	Short: "Matrix/Element homeserver management",
	Long: `Full-featured Matrix homeserver management — setup, connect, users, rooms,
agents, daemons, messaging.`,
	Run: matrixWelcome,
}

func matrixWelcome(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("MATRIX"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Matrix/Element Homeserver Management"))
	fmt.Println(theme.DimTextStyle.Render("  Setup / Users / Rooms / Agents / Daemons"))
	fmt.Println()

	quick := []struct{ cmd, desc string }{
		{"anime matrix setup", "Deploy a native Synapse homeserver"},
		{"anime matrix connect --url <url>", "Connect to an existing server"},
		{"anime matrix status", "Check server health"},
		{"anime matrix users add <name>", "Create a user"},
		{"anime matrix rooms create <name>", "Create a room"},
		{"anime matrix agents spawn <name>", "Spawn a Claude Code bot"},
		{"anime matrix send <room> \"msg\"", "Send a message"},
	}
	fmt.Println(theme.GlowStyle.Render("Quick Start:"))
	fmt.Println()
	for _, q := range quick {
		fmt.Printf("  %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-40s", q.cmd)),
			theme.DimTextStyle.Render(q.desc))
	}
	fmt.Println()
}

// ─── config ─────────────────────────────────────────────────────────────────

var matrixConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Matrix CLI configuration",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := matrixcfg.Load()
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX CONFIG"))
		fmt.Println()
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("File:"), theme.DimTextStyle.Render(matrixcfg.Path()))
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
		cfg, err := matrixcfg.Load()
		if err != nil {
			return err
		}
		switch args[0] {
		case "homeserver.url":
			cfg.Homeserver.URL = args[1]
		case "homeserver.domain":
			cfg.Homeserver.Domain = args[1]
		case "homeserver.admin_token":
			cfg.Homeserver.AdminToken = args[1]
		case "homeserver.admin_user":
			cfg.Homeserver.AdminUser = args[1]
		case "synapse.data_dir":
			cfg.Synapse.DataDir = args[1]
		default:
			return fmt.Errorf("unknown key: %s", args[0])
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("  %s %s = %s\n", theme.SymbolSuccess, theme.HighlightStyle.Render(args[0]), theme.DimTextStyle.Render(args[1]))
		return nil
	},
}

var matrixConfigInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a fresh configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(matrixcfg.Path()); err == nil {
			fmt.Printf("  %s\n", theme.WarningStyle.Render("Config already exists: "+matrixcfg.Path()))
			return nil
		}
		cfg := &matrixcfg.Config{
			Homeserver: matrixcfg.HomeserverConfig{URL: "http://localhost:8008", Domain: "localhost"},
			Synapse:    matrixcfg.SynapseConfig{DataDir: matrixcfg.Dir() + "/data"},
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Config initialized: "+matrixcfg.Path()))
		return nil
	},
}

// ─── send ───────────────────────────────────────────────────────────────────

var matrixSendAsUser string

var matrixSendCmd = &cobra.Command{
	Use:   "send <room-id> <message>",
	Short: "Send a message to a room",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		roomID := args[0]
		message := strings.Join(args[1:], " ")
		cfg, err := matrixcfg.Load()
		if err != nil {
			return err
		}
		token := cfg.Homeserver.AdminToken
		sender := cfg.Homeserver.AdminUser
		if matrixSendAsUser != "" {
			agent := cfg.GetAgent(matrixSendAsUser)
			if agent == nil {
				return fmt.Errorf("agent %q not found", matrixSendAsUser)
			}
			token = agent.AccessToken
			sender = agent.UserID
		}
		client := matrixapi.NewClient(cfg.Homeserver.URL, token)
		eventID, err := client.SendMessage(roomID, message)
		if err != nil {
			return err
		}
		fmt.Printf("  %s %s  %s -> %s  %s\n",
			theme.SymbolStar, theme.SuccessStyle.Render("Sent"),
			theme.DimTextStyle.Render(sender), theme.DimTextStyle.Render(roomID),
			theme.DimTextStyle.Render(eventID))
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

func matrixSplitUserID(userID string) (string, string) {
	if len(userID) < 2 || userID[0] != '@' {
		return userID, ""
	}
	userID = userID[1:]
	for i, c := range userID {
		if c == ':' {
			return userID[:i], userID[i+1:]
		}
	}
	return userID, ""
}

func matrixSeparator() string {
	return theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// keep imports used
var _ = synapse.IsHealthy
var _ = matrixapi.NewClient

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
