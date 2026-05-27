package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	t "github.com/joshkornreich/anime/internal/term"
	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/spf13/cobra"
)

var (
	mxAgentModel    string
	mxAgentChannels []string
	mxAgentChannel  string
	mxAgentPassword string
	mxAgentPrompt   string
)

var matrixAgentsCmd = &cobra.Command{
	Use:     "agents",
	Aliases: []string{"agent", "a"},
	Short:   "Manage Claude Code agents as Mattermost bot users",
	Run:     func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var matrixAgentsSpawnCmd = &cobra.Command{
	Use:   "spawn <name>",
	Short: "Spawn a new Claude Code agent bot",
	Example: `  anime matrix agents spawn helper --channel <channel-id>
  anime matrix agents spawn coder --channels <id1>,<id2>
  anime matrix agents spawn reviewer --channel <id> --prompt "You review code"`,
	Args: cobra.ExactArgs(1),
	RunE: runMatrixAgentsSpawn,
}

var matrixAgentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		t.Section("AGENTS")
		if len(cfg.Agents) == 0 {
			t.Info("no agents")
			fmt.Println("  " + t.Dim("anime matrix agents spawn <name> --channel <channel-id>"))
			fmt.Println()
			return nil
		}
		tbl := t.NewTable("name", "user", "model", "status", "pid")
		for _, a := range cfg.Agents {
			alive := matrixIsAlive(a.PID)
			st := a.Status
			if !alive && st == "running" {
				st = "dead"
			}
			stCell := t.Jade.S(st)
			if st != "running" {
				stCell = t.Loss.S(st)
			}
			tbl.Row(
				t.Bold(t.Gold.S(a.Name)),
				t.Dim(a.UserID),
				t.Dim(a.Model),
				stCell,
				fmt.Sprintf("%d", a.PID),
			)
		}
		fmt.Print(tbl.Render())
		fmt.Println()
		return nil
	},
}

var matrixAgentsStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
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
		t.Ok("stopped " + args[0])
		return nil
	},
}

var matrixAgentsRestartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "Restart an agent (stop + respawn)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		agent := cfg.GetAgent(args[0])
		if agent == nil {
			return fmt.Errorf("agent %q not found", args[0])
		}
		if agent.PID > 0 {
			syscall.Kill(-agent.PID, syscall.SIGTERM)
		}
		mxAgentModel = agent.Model
		mxAgentChannels = agent.Channels
		mxAgentChannel = ""
		mxAgentPassword = ""
		return runMatrixAgentsSpawn(cmd, args)
	},
}

var matrixAgentsLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show agent logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
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
	matrixAgentsSpawnCmd.Flags().StringVarP(&mxAgentChannel, "channel", "c", "", "Channel ID to join")
	matrixAgentsSpawnCmd.Flags().StringSliceVar(&mxAgentChannels, "channels", nil, "Multiple channel IDs")
	matrixAgentsSpawnCmd.Flags().StringVar(&mxAgentPassword, "password", "", "Agent account password")
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
	cfg, _ := mmcfg.Load()

	channels := mxAgentChannels
	if mxAgentChannel != "" {
		channels = append(channels, mxAgentChannel)
	}
	if mxAgentPassword == "" {
		mxAgentPassword = matrixGeneratePassword(24)
	}

	t.Section("SPAWN AGENT")

	client := mmClient(cfg.Server.URL, cfg.Server.Token)

	agentUsername := "agent-" + name
	agentEmail := agentUsername + "@chat.local"
	var agentUserID string

	u, err := client.GetUserByUsername(agentUsername)
	if err != nil {
		u, err = client.CreateUser(agentUsername, agentEmail, mxAgentPassword)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		t.Ok("user created: @" + agentUsername)
	} else {
		t.Info("user exists, reusing @" + agentUsername)
		if err := client.ResetPassword(u.ID, mxAgentPassword); err != nil {
			return fmt.Errorf("reset password: %w", err)
		}
	}
	agentUserID = u.ID

	if cfg.Server.TeamID != "" {
		_ = client.AddTeamMember(cfg.Server.TeamID, agentUserID)
	}

	t.Info("authenticating agent…")
	agentToken, err := mmClient(cfg.Server.URL, "").Login(agentUsername, mxAgentPassword)
	if err != nil {
		return fmt.Errorf("agent login: %w", err)
	}
	t.Ok("authenticated")

	agentClient := mmClient(cfg.Server.URL, agentToken)
	for _, chID := range channels {
		if err := agentClient.AddChannelMember(chID, agentUserID); err != nil {
			t.Warn("join " + chID + ": " + err.Error())
		} else {
			t.Ok("joined " + t.Dim(chID))
		}
	}

	logDir := filepath.Join(mmcfg.Dir(), "logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "agent-"+name+".log")

	runnerScript := mmAgentRunner(name, cfg.Server.URL, agentToken, agentUserID, mxAgentModel, mxAgentPrompt, channels)
	runnerDir := filepath.Join(mmcfg.Dir(), "runners")
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
	t.Ok(fmt.Sprintf("daemon started  PID %d", pid))

	cfg.AddAgent(mmcfg.AgentConfig{
		Name: name, UserID: agentUserID, Token: agentToken,
		Channels: channels, Model: mxAgentModel, Status: "running",
		PID: pid, LogFile: logFile,
	})
	cfg.AddDaemon(mmcfg.DaemonConfig{
		Name: "agent-" + name, PID: pid, Status: "running",
		StartedAt: time.Now().Format(time.RFC3339), Type: "agent", LogFile: logFile,
	})
	cfg.Save()

	fmt.Println()
	t.KV("agent", name)
	t.KV("user", "@"+agentUsername)
	t.KV("model", mxAgentModel)
	t.KV("logs", logFile)
	fmt.Println()
	return nil
}

func mmAgentRunner(name, mmURL, token, botUserID, model, systemPrompt string, channels []string) string {
	prompt := systemPrompt
	if prompt == "" {
		prompt = fmt.Sprintf("You are %s, a helpful AI assistant in a team chat. Be concise and helpful.", name)
	}
	escapedPrompt := strings.ReplaceAll(prompt, "'", "'\\''")
	channelList := strings.Join(channels, " ")

	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail
MM_URL="%s"
TOKEN="%s"
BOT_USER_ID="%s"
MODEL="%s"
AGENT_NAME="%s"
CHANNELS=(%s)
SYSTEM_PROMPT='%s'
BACKOFF=1
MAX_BACKOFF=60

log() { echo "[$(date '+%%Y-%%m-%%d %%H:%%M:%%S')] [$AGENT_NAME] $1"; }
command -v curl >/dev/null || { log "curl not found"; exit 1; }
command -v jq >/dev/null || { log "jq not found"; exit 1; }
command -v claude >/dev/null || { log "claude not found"; exit 1; }

send_post() {
    local channel_id="$1" message="$2"
    curl -sf -X POST \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$(jq -n --arg c "$channel_id" --arg m "$message" '{channel_id:$c,message:$m}')" \
        "$MM_URL/api/v4/posts" >/dev/null 2>&1
}

log "Starting. Channels: ${CHANNELS[*]:-none}"
SINCE=$(date +%%s%%3N)

while true; do
    for channel_id in "${CHANNELS[@]:-}"; do
        [ -z "$channel_id" ] && continue
        RESP=$(curl -sf \
            -H "Authorization: Bearer $TOKEN" \
            "$MM_URL/api/v4/channels/$channel_id/posts?since=$SINCE&per_page=50" 2>/dev/null) || {
            log "Poll failed for $channel_id, backoff ${BACKOFF}s"
            sleep "$BACKOFF"
            BACKOFF=$(( BACKOFF * 2 > MAX_BACKOFF ? MAX_BACKOFF : BACKOFF * 2 ))
            continue
        }
        BACKOFF=1

        echo "$RESP" | jq -r --arg bot "$BOT_USER_ID" '
            .order[] as $id |
            .posts[$id] |
            select(.user_id != $bot) |
            select(.type == "") |
            select(.message | length > 0) |
            [.channel_id, .user_id, .message] | @tsv
        ' 2>/dev/null | while IFS=$'\t' read -r ch_id user_id message; do
            log "MSG from $user_id in $ch_id: ${message:0:80}"
            response=$(claude -p \
                --model "$MODEL" \
                --system-prompt "$SYSTEM_PROMPT" \
                --no-session-persistence \
                "$message" 2>/dev/null) || response="Sorry, I encountered an error."
            [ -z "$response" ] && response="(empty response)"
            send_post "$ch_id" "$response"
            log "REPLIED in $ch_id (${#response} chars)"
        done
    done

    SINCE=$(date +%%s%%3N)
    sleep 2
done
`, mmURL, token, botUserID, model, name, channelList, escapedPrompt)
}
