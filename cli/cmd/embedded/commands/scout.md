---
description: Send 15 scout subagents to explore the repository for rapid context acquisition
allowed-tools: Agent, Bash, Read, Glob, Grep
user-intent: exploration
---

# SCOUT — Scalable Repository Reconnaissance

Launch Explore subagents in parallel to map the current repository. Default: **5**. Accepts `5`, `10`, or `15`. Add `opus` for Opus-tier scouts.

## Argument Parsing

Parse `$ARGUMENTS` for:
- A number: `5`, `10`, or `15` (default: `5` if omitted)
- The word `opus` (case-insensitive) — if present, use `model: "opus"` on every Agent call
- An optional **focus prompt** — any remaining text after the number and model keyword. When present, ALL scouts receive this as additional context appended to their individual mission: "Additional focus: {prompt}". This steers every scout toward the user's area of interest while still covering their assigned domain.

Examples:
- `/scout` → 5 scouts, default model, general exploration
- `/scout 10` → 10 scouts, default model
- `/scout 15` → 15 scouts, default model
- `/scout opus` → 5 scouts, Opus model
- `/scout 15 opus` → 15 scouts, Opus model
- `/scout 10 find all dead code` → 10 scouts, each focused on dead code within their domain
- `/scout 15 opus how does the compiler handle composition expansion` → 15 Opus scouts, all oriented toward composition expansion
- `/scout security vulnerabilities` → 5 scouts, each examining security within their domain

## Execution

Use the Agent tool with `subagent_type: "Explore"`. Launch ALL scouts for the chosen tier in a **single message** (parallel). Each scout should report in under 300 words.

If `opus` was specified, add `model: "opus"` to every Agent call.

---

## Tier 1 — Always dispatched (5 scouts)

**Scout 1 — Project Skeleton**
Map the top-level directory structure. What are the main directories? What's the overall architecture pattern (monorepo, single app, library, etc.)? List the top 2 levels of directories.

**Scout 2 — Build & Config**
Find all build configuration: package.json, Cargo.toml, Makefile, tsconfig, webpack/vite/rollup configs, CMakeLists. Report the tech stack, build commands, and toolchain.

**Scout 3 — Entry Points**
Find the main entry points: index files, main files, app files, server entry. Trace the boot sequence — what loads first and what does it initialize?

**Scout 4 — Documentation**
Find README, CLAUDE.md, docs directories, PURPOSE.md, ARCHITECTURE.md, or any documentation files. Summarize what the project says about itself.

**Scout 5 — Core Logic**
Find the main business logic — the code that does the actual work (not glue, not config, not UI). What are the core algorithms or computations?

---

## Tier 2 — Added at 10+ (scouts 6-10)

**Scout 6 — Recent Git Activity**
Run `git log --oneline -30` and `git diff --stat HEAD~10..HEAD`. What areas are actively being worked on? What's the recent focus?

**Scout 7 — Dependencies**
Examine dependency files (package.json dependencies, requirements.txt, Cargo.toml deps, go.mod). What are the key libraries? Any notable or unusual dependencies?

**Scout 8 — API Surface**
Find HTTP routes, API endpoints, exported functions, public interfaces. What does this project expose to the outside world?

**Scout 9 — Types & Schemas**
Find type definitions, interfaces, schemas, models, database migrations. What are the key data structures?

**Scout 10 — Test Infrastructure**
Find test files and test configuration. What test framework is used? What's the test structure? How do you run tests? Report coverage if configured.

---

## Tier 3 — Added at 15 (scouts 11-15)

**Scout 11 — State & Data Flow**
Find state management patterns: stores, context providers, global state, databases, caches. How does data flow through the system?

**Scout 12 — UI & Rendering**
Find UI components, templates, shaders, rendering code, HTML files. What's the presentation layer? What rendering approach is used?

**Scout 13 — Scripts & Tooling**
Find scripts/ directories, CI/CD configs (.github/workflows, Dockerfile), dev tools, code generation. What automation exists?

**Scout 14 — Error Handling & Logging**
Search for error handling patterns, logging setup, monitoring, error boundaries. How does the project handle failures?

**Scout 15 — Secrets & Environment**
Find .env.example, config files, environment variable references. What external services or configuration does the project need? (Do NOT read actual .env files — only examples and references.)

---

## Synthesis

After all scouts return, produce a structured briefing:

```
## Repository Briefing: [project name]

**Stack**: [languages, frameworks, key deps]
**Architecture**: [pattern in 1 sentence]
**Entry point**: [where execution begins]
**Build**: [how to build/run]
**Test**: [how to test]

### Key Areas
[3-5 bullet points on the most important code areas]

### Active Work
[What's being worked on based on git history]

### Notable
[Anything unusual, impressive, or important to know]
```

Keep the final briefing under 500 words. Dense signal, no filler.
