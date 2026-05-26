package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxAgentModel    string
	mxAgentRoom     string
	mxAgentRooms    []string
	mxAgentPassword string
	mxAgentPrompt   string
)

var matrixAgentsCmd = &cobra.Command{
	Use:     "agents",
	Aliases: []string{"agent", "a"},
	Short:   "Manage Claude Code agents as Matrix bot users",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixAgentsSpawnCmd = &cobra.Command{
	Use:   "spawn <name>",
	Short: "Spawn a new Claude Code agent",
	Example: `  anime matrix agents spawn helper --room '!abc:localhost'
  anime matrix agents spawn coder --rooms general,dev
  anime matrix agents spawn reviewer --room '!abc:localhost' --prompt "You review code"`,
	Args: cobra.ExactArgs(1),
	RunE: runMatrixAgentsSpawn,
}

var matrixAgentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		fmt.Println()
		fmt.Println(theme.RenderBanner("MATRIX AGENTS"))
		fmt.Println()
		if len(cfg.Agents) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  No agents"))
			fmt.Printf("  %s\n\n", theme.HighlightStyle.Render("anime matrix agents spawn <name> --room <room>"))
			return nil
		}
		for _, a := range cfg.Agents {
			alive := matrixIsAlive(a.PID)
			st := a.Status
			if !alive && st == "running" {
				st = "dead"
			}
			stStr := theme.SuccessStyle.Render(st)
			if st != "running" {
				stStr = theme.ErrorStyle.Render(st)
			}
			fmt.Printf("  %s %s\n", theme.SymbolStar, theme.HighlightStyle.Render(a.Name))
			fmt.Printf("    User: %s  Model: %s  %s  PID %d\n",
				theme.DimTextStyle.Render(a.UserID), theme.DimTextStyle.Render(a.Model), stStr, a.PID)
			if len(a.Rooms) > 0 {
				fmt.Printf("    Rooms: %s\n", theme.DimTextStyle.Render(strings.Join(a.Rooms, ", ")))
			}
			fmt.Println()
		}
		return nil
	},
}

var matrixAgentsStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		agent := cfg.GetAgent(args[0])
		if agent == nil {
			return fmt.Errorf("agent %q not found", args[0])
		}
		if agent.PID > 0 {
			syscall.Kill(-agent.PID, syscall.SIGTERM)
		}
		agent.Status = "stopped"
		agent.PID = 0
		cfg.RemoveDaemon("agent-" + args[0])
		cfg.Save()
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Stopped "+args[0]))
		return nil
	},
}

var matrixAgentsRestartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "Restart an agent (stop + respawn)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		agent := cfg.GetAgent(args[0])
		if agent == nil {
			return fmt.Errorf("agent %q not found", args[0])
		}
		// Stop
		if agent.PID > 0 {
			syscall.Kill(-agent.PID, syscall.SIGTERM)
		}
		// Re-spawn with saved config
		mxAgentModel = agent.Model
		mxAgentRooms = agent.Rooms
		mxAgentRoom = ""
		mxAgentPassword = "" // will re-login with saved token
		return runMatrixAgentsSpawn(cmd, args)
	},
}

var matrixAgentsLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show agent logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		agent := cfg.GetAgent(args[0])
		if agent == nil {
			return fmt.Errorf("agent %q not found", args[0])
		}
		if agent.LogFile == "" {
			return fmt.Errorf("no log file")
		}
		return matrixRunBash(fmt.Sprintf("tail -100 %s", agent.LogFile))
	},
}

func init() {
	matrixAgentsSpawnCmd.Flags().StringVarP(&mxAgentModel, "model", "m", "claude-sonnet-4-20250514", "Claude model")
	matrixAgentsSpawnCmd.Flags().StringVarP(&mxAgentRoom, "room", "r", "", "Room to join")
	matrixAgentsSpawnCmd.Flags().StringSliceVar(&mxAgentRooms, "rooms", nil, "Multiple rooms")
	matrixAgentsSpawnCmd.Flags().StringVar(&mxAgentPassword, "password", "", "Agent password")
	matrixAgentsSpawnCmd.Flags().StringVar(&mxAgentPrompt, "prompt", "", "System prompt")

	matrixAgentsCmd.AddCommand(matrixAgentsSpawnCmd)
	matrixAgentsCmd.AddCommand(matrixAgentsListCmd)
	matrixAgentsCmd.AddCommand(matrixAgentsStopCmd)
	matrixAgentsCmd.AddCommand(matrixAgentsRestartCmd)
	matrixAgentsCmd.AddCommand(matrixAgentsLogsCmd)
	matrixCmd.AddCommand(matrixAgentsCmd)
}

func runMatrixAgentsSpawn(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, _ := matrixcfg.Load()

	rooms := mxAgentRooms
	if mxAgentRoom != "" {
		rooms = append(rooms, mxAgentRoom)
	}
	if mxAgentPassword == "" {
		mxAgentPassword = matrixGeneratePassword(24)
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("SPAWN AGENT"))
	fmt.Println()

	// Create Matrix user
	agentUsername := "agent-" + name
	admin := matrixapi.NewAdminClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken, cfg.Homeserver.Domain)
	if err := admin.CreateUser(agentUsername, mxAgentPassword, fmt.Sprintf("[Bot] %s", name), false); err != nil {
		if !strings.Contains(err.Error(), "User ID already taken") {
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolInfo, theme.DimTextStyle.Render("User exists, reusing"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("User created: "+agentUsername))
	}

	// Login
	agentClient := matrixapi.NewClient(cfg.Homeserver.URL, "")
	token, err := agentClient.Login(agentUsername, mxAgentPassword)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	agentClient.AccessToken = token
	fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Authenticated"))

	// Join rooms
	userID := fmt.Sprintf("@%s:%s", agentUsername, cfg.Homeserver.Domain)
	adminClient := matrixapi.NewClient(cfg.Homeserver.URL, cfg.Homeserver.AdminToken)
	for _, room := range rooms {
		adminClient.InviteUser(room, userID)
		if _, err := agentClient.JoinRoom(room); err != nil {
			fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Join "+room+": "+err.Error()))
		} else {
			fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Joined "+room))
		}
	}

	// Spawn daemon
	logDir := filepath.Join(matrixcfg.Dir(), "logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "agent-"+name+".log")

	runnerScript := matrixAgentRunner(name, cfg.Homeserver.URL, token, mxAgentModel, mxAgentPrompt)
	runnerDir := filepath.Join(matrixcfg.Dir(), "runners")
	os.MkdirAll(runnerDir, 0755)
	runnerPath := filepath.Join(runnerDir, "agent-"+name+".sh")
	os.WriteFile(runnerPath, []byte(runnerScript), 0755)

	logF, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	proc := exec.Command("bash", runnerPath)
	proc.Stdout = logF
	proc.Stderr = logF
	proc.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := proc.Start(); err != nil {
		logF.Close()
		return err
	}
	logF.Close()

	pid := proc.Process.Pid
	fmt.Printf("  %s %s PID %d\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Daemon started"), pid)

	// Save
	cfg.AddAgent(matrixcfg.AgentConfig{
		Name: name, UserID: userID, AccessToken: token,
		Rooms: rooms, Model: mxAgentModel, Status: "running",
		PID: pid, LogFile: logFile,
	})
	cfg.AddDaemon(matrixcfg.DaemonConfig{
		Name: "agent-" + name, PID: pid, Status: "running",
		StartedAt: time.Now().Format(time.RFC3339), Type: "agent", LogFile: logFile,
	})
	cfg.Save()

	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Agent:"), theme.InfoStyle.Render(name))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("User:"), theme.DimTextStyle.Render(userID))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Model:"), theme.DimTextStyle.Render(mxAgentModel))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Logs:"), theme.DimTextStyle.Render(logFile))
	fmt.Println()
	return nil
}

func matrixAgentRunner(name, homeserverURL, token, model, systemPrompt string) string {
	prompt := systemPrompt
	if prompt == "" {
		prompt = fmt.Sprintf("You are %s, a helpful AI assistant in a Matrix chat. Be concise.", name)
	}
	escapedPrompt := strings.ReplaceAll(prompt, "'", "'\\''")

	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail
HOMESERVER="%s"
TOKEN="%s"
MODEL="%s"
AGENT_NAME="%s"
SYSTEM_PROMPT='%s'
BACKOFF=1
MAX_BACKOFF=60

log() { echo "[$(date '+%%Y-%%m-%%d %%H:%%M:%%S')] [$AGENT_NAME] $1"; }
command -v curl >/dev/null || { log "curl not found"; exit 1; }
command -v jq >/dev/null || { log "jq not found"; exit 1; }
command -v claude >/dev/null || { log "claude not found"; exit 1; }

send_message() {
    local room_id="$1" body="$2"
    curl -sf -X PUT \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$(jq -n --arg b "$body" '{msgtype:"m.text", body:$b}')" \
        "$HOMESERVER/_matrix/client/v3/rooms/$room_id/send/m.room.message/$(date +%%s%%N)" \
        >/dev/null 2>&1
}

log "Starting"
SINCE=$(curl -sf -H "Authorization: Bearer $TOKEN" \
    "$HOMESERVER/_matrix/client/v3/sync?timeout=0" \
    | jq -r '.next_batch // empty' 2>/dev/null) || true
log "Initial sync (since=${SINCE:-none})"

while true; do
    SYNC_URL="$HOMESERVER/_matrix/client/v3/sync?timeout=30000"
    [ -n "${SINCE:-}" ] && SYNC_URL="$SYNC_URL&since=$SINCE"
    RESP=$(curl -sf -H "Authorization: Bearer $TOKEN" "$SYNC_URL" 2>/dev/null) || {
        log "Sync failed, backoff ${BACKOFF}s"
        sleep "$BACKOFF"
        BACKOFF=$(( BACKOFF * 2 > MAX_BACKOFF ? MAX_BACKOFF : BACKOFF * 2 ))
        continue
    }
    BACKOFF=1
    NEW_SINCE=$(echo "$RESP" | jq -r '.next_batch // empty' 2>/dev/null)
    [ -n "$NEW_SINCE" ] && SINCE="$NEW_SINCE"

    echo "$RESP" | jq -r '
        .rooms.join // {} | to_entries[] |
        .key as $room |
        .value.timeline.events[]? |
        select(.type == "m.room.message") |
        select(.sender | test("agent-'"$AGENT_NAME"'") | not) |
        select(.content.body // "" | length > 0) |
        [$room, .sender, .content.body] | @tsv
    ' 2>/dev/null | while IFS=$'\t' read -r room_id sender body; do
        log "MSG from $sender in $room_id: ${body:0:80}"
        response=$(claude -p \
            --model "$MODEL" \
            --system-prompt "$SYSTEM_PROMPT" \
            --no-session-persistence \
            "$body" 2>/dev/null) || response="Sorry, I encountered an error."
        [ -z "$response" ] && response="(empty response)"
        send_message "$room_id" "$response"
        log "REPLIED in $room_id (${#response} chars)"
    done
    sleep 1
done
`, homeserverURL, token, model, name, escapedPrompt)
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
