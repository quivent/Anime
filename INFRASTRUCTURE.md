# Infrastructure Reference

## What This Is

**anime** is a Go CLI tool (`github.com/joshkornreich/anime`) that started as a Lambda GH200 GPU instance manager and has grown into a general-purpose infrastructure management tool for the Quivent/Influx ecosystem. It manages remote servers, deploys services, and runs agents.

---

## The Server: flux (`88.216.198.144`)

A dedicated bare-metal server (not Lambda), running Ubuntu 24.04.

| Resource | Value |
|---|---|
| RAM | 62 GB |
| Disk | 792 GB NVMe (`/dev/nvme0n1p2`, 879 GB total, 44 GB used) |
| OS | Ubuntu 24.04 |
| SSH user | `flux` (sudo password: configured separately) |

### Services Running on flux

| Service | Status | Notes |
|---|---|---|
| **nginx** | active | Reverse proxy for all domains |
| **postgresql@16-main** | active | Database for Mattermost |
| **mattermost** | active | Team chat server |
| **tuwunel** | active | Matrix homeserver (Conduit fork) |

### Domains (all on the same TLS cert: `flux.family` SAN)

| Domain | Points To | Notes |
|---|---|---|
| `flux.family` | nginx → static | Serves Matrix `.well-known` delegation |
| `matrix.flux.family` | nginx → Tuwunel (port TBD) | Matrix homeserver for `@user:flux.family` IDs |
| `chat.flux.family` | nginx → Mattermost `:8065` | Team chat, HTTPS, WebSocket-aware |

---

## Mattermost

**Version:** 9.11.0  
**URL:** `https://chat.flux.family`  
**Local port:** `8065` (not exposed directly; nginx proxies)  
**Install path:** `/opt/mattermost`  
**Database:** PostgreSQL 16, DB name `mattermost`, user `mattermost`  
**Systemd unit:** `mattermost.service`  
**Logs:** `journalctl -u mattermost` or `/opt/mattermost/logs/`

**Admin credentials** are stored in `~/.matrix/config.yaml` on the local machine.

### CLI connection (`anime matrix`)

The `anime matrix` subcommand manages this Mattermost instance:

```
anime matrix status                    # server health, users, teams
anime matrix connect --url <url>       # authenticate and save token
anime matrix users list                # list all users
anime matrix users add <name>          # create user
anime matrix channels list             # list channels
anime matrix channels create <name>    # create channel
anime matrix agents spawn <name>       # spawn a Claude Code bot user
anime matrix watch <channel-id>        # live-tail a channel
anime matrix history <channel-id>      # read message history
anime matrix send <channel-id> "msg"   # send a message
```

Config stored at `~/.matrix/config.yaml`.

---

## The CLI: `anime`

**Repo:** `github.com/quivent/Anime` (origin), `github.com/Influx-Designs/anime` (influx)  
**Language:** Go 1.25, Cobra + Bubble Tea  
**Binary:** built locally, installed to `~/bin/anime`  
**Module:** `github.com/joshkornreich/anime`

### Major command groups

| Command | Purpose |
|---|---|
| `anime matrix` | Mattermost management (users, channels, agents, watch) |
| `anime deploy` | SSH deployment to remote servers |
| `anime server` | Manage configured Lambda/remote servers |
| `anime wan` | Wan video generation stack (ComfyUI) |
| `anime comfyui` | ComfyUI node and model management |
| `anime vllm` | vLLM inference server management |
| `anime ollama` | Ollama model management |
| `anime inference` | Inference routing |
| `anime lambda` | Lambda Labs API integration |
| `anime install` | Package/module installation on servers |
| `anime git` / `anime gh` | Git and GitHub shortcuts |
| `anime db` | Database operations |
| `anime dns` | DNS management |
| `anime service` | systemd service management |
| `anime cron` | Cron job management |
| `anime backup` | Backup operations |
| `anime logs` | Log tailing |
| `anime ssh` | SSH shortcuts |
| `anime sync` / `anime rsync` | File sync |
| `anime config` | CLI configuration TUI |

### Internal packages

| Package | Purpose |
|---|---|
| `internal/term` | Aurum design system — stdlib-only ANSI, no external deps |
| `internal/theme` | Design layer — wraps `term`, exposes `Ok/Fail/Warn/Info/Section/KV/Rule/NewTable` |
| `internal/mmapi` | Mattermost REST API v4 client |
| `internal/mmcfg` | Mattermost config (`~/.matrix/config.yaml`) |
| `internal/installer` | SSH-based module installer |
| `internal/ssh` | SSH client |
| `internal/tui` | Bubble Tea TUI screens |
| `internal/vfs` | Embedded virtual filesystem |

---

## Design System: Aurum

Palette: molten gold over obsidian. Stdlib-only ANSI (no Charmbracelet dependency in the rendering path).

| Color | Hex | Use |
|---|---|---|
| Gold | `#D9B45A` | Headings, highlights, names |
| GoldBright | `#F6DF9A` | Emphasis |
| GoldDeep | `#A6802F` | Subdued gold |
| Cyan | `#41E0D0` | Structure, links, info |
| Jade | `#4ADE80` | Success, online, accepted |
| Loss | `#FF5C5C` | Error, dead, rejected |
| Ink | `#D8D5CC` | Body text |
| InkMuted | `#9A958B` | Secondary text |
| InkFaint | `#635F58` | Timestamps, metadata |

**Usage from any cmd file:**
```go
import "github.com/joshkornreich/anime/internal/theme"

theme.Ok("server ready")
theme.Fail("connection refused")
theme.Warn("partial result")
theme.Info("connecting to " + theme.Cyan.S(url))
theme.Section("USERS")
theme.KV("server", url)
theme.Rule()
tbl := theme.NewTable("name", "status", "pid")
tbl.Row(theme.Bold(theme.Gold.S(name)), theme.Jade.S("running"), pid)
fmt.Print(tbl.Render())
```

---

## Matrix / Tuwunel (legacy, still running)

The `matrix.flux.family` domain still points to a running Tuwunel (Matrix/Conduit fork) homeserver. This was the previous chat infrastructure before the Mattermost pivot. It is still active but no longer managed by the `anime matrix` subcommand (which now targets Mattermost exclusively).

User IDs on the old system were `@name:flux.family`.

---

## Local Config Files

| File | Purpose |
|---|---|
| `~/.matrix/config.yaml` | Mattermost server URL, token, team ID |
| `~/.matrix/logs/` | Agent log files |
| `~/.matrix/runners/` | Agent bash runner scripts |
| `~/.config/anime/` | Main CLI config (servers, API keys) |
