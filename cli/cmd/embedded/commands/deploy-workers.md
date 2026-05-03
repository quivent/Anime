Deploy the autonomous worker protocol to the current working directory.

## What This Does

Creates a self-contained autonomous worker system in the current directory with:
- `daemon.py` — perpetual background worker (30-min cycles via launchd)
- `tasks.py` — domain-specific task definitions
- `populate-tasks.py` — auto-generates tasks from gaps in progress
- `MEMORY.md` — session memory (read every cycle)
- `PROGRESS.md` — task queue (read/written every cycle)
- `LAW.md` — the eleven laws governing all work
- `INTEGRATED.md` — tracks what's built vs designed vs shipped
- `workers-state.json` — active worker tracking
- `update-workers.sh` — add/done/list/reset workers
- `start.sh` / `stop.sh` / `status.sh` — daemon control
- `config/daemon.json` — cycle interval, API model, thresholds
- Menu bar app with tier/worker/task display
- launchd plist for auto-start

## How to Use

1. Run `/deploy-workers` in any repository
2. Edit `MEMORY.md` with the project's mission
3. Edit `PROGRESS.md` with initial tasks
4. Edit `tasks.py` to define domain-specific task functions
5. Run `./start.sh` to begin autonomous operation
6. Run `/autonomous` to orchestrate workers manually

## Implementation

Read the template system at `~/waveworkers-system/` and adapt it:
- Copy all infrastructure files (daemon, tasks, scripts, law)
- Generate project-specific MEMORY.md from the current repo's README, CLAUDE.md, or purpose docs
- Generate initial PROGRESS.md by scanning for TODOs, FIXMEs, open issues
- Set up launchd with a unique identifier based on the directory name
- Build a menu bar app with the project name

The daemon name, PID file, and launchd plist must be unique per deployment to avoid conflicts with other running daemons.

**Do NOT copy waveworkers-specific content** (P-functions, vocal synthesis, ForGE plans). Copy only the infrastructure. The content comes from the target repo.
