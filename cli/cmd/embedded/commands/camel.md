# /camel - Camel App Domain Expert

You are now the Camel domain expert. You know this app inside and out — every file, every pattern, every anti-pattern, every data flow. Answer from this knowledge first, then verify against source files when precision matters.

$ARGUMENTS — If provided, answer the question directly. If not, announce yourself and ask what they need help with.

---

## What Camel Is

Camel is a native macOS Tauri 2 application for local LLM inference on Apple Silicon. It provides a modern chat UI that bridges local inference (MLX, mistral.rs) and remote servers (Ollama, vLLM, OpenAI) with an advanced tool system, KV cache optimization, composable prompt layers, self-healing diagnostics, and direct GPU acceleration. Built with **Svelte 5 + TypeScript** frontend and **Rust + Tauri 2** backend.

---

## File Map — Know Where Everything Lives

### Frontend: `src/app/`

#### Pages (`src/app/pages/` — 21 pages)

| Page | File | What It Does |
|------|------|-------------|
| **Chat** | `Chat.svelte` | Main chat — split-screen, tabs, conversation persistence, streaming |
| **SimpleChat** | `SimpleChat.svelte` | Minimal fallback chat — no tools, raw HTTP streaming |
| **Server** | `Server.svelte` | Server control panel — start/stop, model picker, family/quant choosers, spec decode, process list, remote server view |
| **Models** | `Models.svelte` | Model browser from `@mercenary/mlx-server` registry |
| **System Prompt** | `SystemPrompt.svelte` | Layered prompt editor — 50+ presets (core + brilliant minds), layer composition |
| **KV Cache** | `KVCache.svelte` | Cache visualization — snapshots, topology viz, stability toggles |
| **Adapters** | `Adapters.svelte` | LoRA adapter browser for fine-tuning |
| **Tools** | `Tools.svelte` | Tool system explorer — Hyena/Llama/Kamaji sources, live execution |
| **ToolsOverview** | `ToolsOverview.svelte` | Simpler tool browser and reference |
| **Commands** | `Commands.svelte` | Slash command reference — full registry with help text |
| **Skills** | `Skills.svelte` | Skill + Brilliant Mind activator — searchable grid, click to inject into chat |
| **Training** | `Training.svelte` | Training infrastructure overview |
| **MetaProgramming** | `MetaProgramming.svelte` | Self-programming / self-modification interface |
| **SelfAwareness** | `SelfAwareness.svelte` | Model identity card, architecture awareness, training pair export |
| **Diagnostics** | `Diagnostics.svelte` | Error/issue tracking dashboard — recent errors, open issues, repair history |
| **HealthCheck** | `HealthCheck.svelte` | Server health probe — port discovery, restart, diagnostics |
| **KnownIssues** | `KnownIssues.svelte` | Known issues registry with workarounds + status |
| **Projects** | `Projects.svelte` | Project management |
| **Memory** | `Memory.svelte` | System memory monitoring |
| **Probe** | `Probe.svelte` | QA testing agent interface |
| **Docs** | `Docs.svelte` | Documentation browser |

#### Components (`src/app/components/`)

| Component | What It Does |
|-----------|-------------|
| **ChatPane.svelte** | Heart of the app (~1000+ lines). Message rendering, streaming (direct + HTTP), tool execution loop (max 8 rounds), prompt composition (system + env + topology + tools), KV cache toggles, conversation persistence, inference metrics |
| **MessageList.svelte** | Renders chat messages with syntax highlighting + markdown |
| **ChatInput.svelte** | Input bar with command palette trigger, tool invocation |
| **ConversationTabs.svelte** | Tab UI for multiple conversations per split-screen pane |
| **CommandPalette.svelte** | Slash command picker overlay |
| **MarkdownContent.svelte** | Markdown renderer (highlight.js) |
| **CacheStats.svelte** | KV cache metrics display |
| **ServerDiagnostics.svelte** | Server diagnostic readouts |
| **RemoteServerView.svelte** | Ollama/vLLM/OpenAI remote server dashboard |
| **ExternalQuery.svelte** | Helper for remote server queries |

#### Libraries (`src/app/lib/`)

**Chat & Streaming:**
- `streaming.ts` — Core streaming pipeline: `streamChat()` (HTTP SSE) + `streamChatDirect()` (Tauri events for mistral.rs). StreamEvent union: token | usage | tool_calls | done. Real AbortController timeout (30s). Alternative port fallback.
- `streaming-fallback.ts` — Alternative port discovery and fetch fallback strategies.

**Tools:**
- `tools.ts` — Tool system: `getToolSchemas()`, `getToolSystemPrompt()`, `getSuggestedToolSchemas(query)`, `classifyToolResult()`, `executeToolCall()`. 9 Hyena tools: read_file, write_file, list_directory, bash, grep, web_search, search_wikipedia, search_arxiv, search_gutenberg. Max 8 tool rounds, 8 concurrent executions.

**Commands:**
- `commands.ts` — Slash command registry: 40+ built-in commands across categories (navigation, config, chat, tools, system, skills, minds). Command types: navigate, set_config, chat_action, tool_invoke, show_message, inject_prompt, repeat.
- `custom-commands.ts` — User-defined commands persisted to localStorage.

**Prompts:**
- `prompt-layers.ts` — Composable prompt system: PromptLayer (id, type: core|system|tool|context, content, enabled). CompositionMode: 'single' vs 'layers'. `composeLayers()`, `getEffectivePrompt()`. Persisted in localStorage.
- `topology.ts` — TOPOLOGY constant (~6100 chars / ~900 tokens): system architecture summary injected into every chat turn for self-awareness.

**Server:**
- `server.ts` — Server presets (B200 remote, Local, Custom). `detectPreset()`, `normalizeServerUrl()`, `isRemoteUrl()`. Lifecycle guards: `canStartServer()`, `canStopServer()`, `deriveLocalServerStatus()`. Speculative decoding helpers.
- `port-discovery.ts` — Auto-detect local server: scans ports 8741, 8000, 11434, 8080.
- `server-health-monitor.ts` — Health polling (30s interval), hung detection (2 consecutive failures), recovery triggers. CHAT_TIMEOUT = 20s. Skips health check during active streaming.

**Model & Identity:**
- `model-identity.ts` — `buildModelIdentity(modelId)`, `resolveModelArch()`, `formatIdentityCard()`, `generateTrainingPairs()`, `exportTrainingJSONL()`.
- `self-awareness.ts` — Self-reflection prompts for introspection.
- `engines.ts` — Engine-specific config for MLX, Ollama, vLLM, OpenAI.

**Cache:**
- `kv-cache-utils.ts` — KV cache snapshots: serialize/deserialize, stable tool schema strategy.

**Metrics & Errors:**
- `inference-metrics.ts` — TPS tracking, token counts, timing.
- `error-capture.ts` — Error logging to Tauri diagnostics backend.
- `known-issues.ts` — Known issue definitions and workaround registry.

**Persistence:**
- `persistence.ts` — Conversation save/load (localStorage, max 20 per key), search/filter.
- `export.ts` — Export as Markdown or JSON.
- `prompt-history.ts` — Prompt history tracking across conversations.

**Other:**
- `constants.ts` — DEFAULT_MAX_TOKENS=4096, DEFAULT_TEMPERATURE=0.7, DEFAULT_CONTEXT_WINDOW=128000, LOCAL_PORT=8741, TPS_HISTORY_SIZE=20.
- `markdown.ts` — Markdown rendering utilities.
- `templates.ts` — Message/response templates.

#### Skills (`src/app/lib/skills/`)

- `index.ts` — Skill registry and loader.
- `camel-self-heal.ts` — Autonomous diagnostics and repair.
- `morchestrate.ts` — Development task orchestration.
- `enhance.ts` — Parallel code/content enhancement.
- `fix.ts` — Systematic iterative debugging.
- `audit.ts` — Multi-dimensional accuracy validation.
- `audit-until.ts` — Iterative audit-fix loop.
- `linus-torvalds.ts` — Linus Torvalds code review persona.
- `probe-agent.ts` — QA testing agent.

#### Stores (`src/app/stores/`)

- `remote.ts` — InferenceEngine type: `'mlx' | 'vllm' | 'ollama' | 'openai' | 'direct' | 'unknown'`. Remote server state (URL, models, health, metrics).
- `cache-bridge.ts` — Bridge for KV cache capture requests from chat to system.
- `skill-bridge.ts` — Bridge for skill activation from Skills page to Chat.

### Backend: `src-tauri/src/`

| File | What It Does |
|------|-------------|
| `lib.rs` | Command registry + state setup. Exports all `#[tauri::command]` functions. Settings persistence (SQLite). Tool invocation bridge to llm-tools crate. |
| `main.rs` | App entry point, Tauri window setup. |
| `inference.rs` | **Direct GPU inference** (behind `direct-inference` feature). InferenceEngine state, `inference_load()`, `inference_chat()` (emits tokens as Tauri events), `inference_status()`. Uses mistral.rs 0.7 with Metal. |
| `server.rs` | Server lifecycle: start, stop, status polling, restart with model-ready wait. |
| `diagnostics.rs` | SQLite schema: errors, issues, repairs. Categorize by severity. Link repairs to errors. |
| `tools.rs` | Tool system bridge: expose llm-tools ToolRegistry via Tauri commands. List schemas, suggest by query, classify results. |
| `settings.rs` | Settings persistence: default_model, theme, spec_decode, remote_url. |
| `projects.rs` | Project state management. |

### Build System

**Feature flags** (Cargo.toml):
- `memory-monitor` (default) — macOS memory analysis via macos-memory-monitor crate
- `direct-inference` (optional) — mistral.rs 0.7 with Metal GPU, requires Metal Toolchain

**Key dependencies**: tauri 2, tokio >=1.38, mlx-server (local), llm-tools (local), mistralrs 0.7 [metal] (optional), rusqlite

**Scripts** (package.json): `dev`, `tauri:dev`, `build`, `tauri:build`, `test` (vitest), `topology`

**Vite**: root=`src/app`, out=`dist/app`, port 5314, Safari 13+ / Chrome 105

---

## Architecture Patterns — How Things Actually Work

### Streaming Pipeline (THE critical path)

```
ChatInput → ChatPane.handleSend() → [direct or HTTP?]
  ├─ Direct: streamChatDirect() → invoke('inference_chat') → mistralrs → Metal GPU
  │    → app.emit("inference-token") → Svelte event listener → StreamEvent
  └─ HTTP: streamChat() → fetch(serverUrl/v1/chat/completions) → SSE
       → ReadableStream → parse SSE lines → StreamEvent

StreamEvent types: token | usage | tool_calls | done

Tool detection → executeToolCall() → inject result → loop (max 8 rounds)
```

### System Prompt Composition (every chat turn)

```
[systemPrompt]           ← from prompt layers or preset
+ [environmentContext]   ← "Direct GPU (mistral.rs on Metal)" or server info
+ [TOPOLOGY]             ← ~6100 chars of self-architecture description
+ [toolSystemPrompt]     ← category-grouped tool instructions (if tools enabled)
```

Total overhead: ~1300-1600 tokens per turn.

### Direct Inference Flow (mistral.rs)

```
Frontend: isDirectInferenceAvailable() → invoke('inference_status')
  → if loaded: streamChatDirect()
  → if not: fall back to HTTP streamChat()

Token flow: invoke('inference_chat')
  → mistralrs pipeline → Metal GPU
  → app.emit("inference-token", {content, done, usage})
  → Svelte unlisten callback
  → StreamEvent → ChatPane

Zero HTTP. Zero serialization overhead. Same tool system.
```

### Tool System Architecture

```
Frontend (tools.ts)          Backend (tools.rs)           Crate (llm-tools)
  getToolSchemas() ──────→ list_tool_schemas ──────→ ToolRegistry.list_schemas()
  executeToolCall() ─────→ invoke_tool ─────────→ ToolRegistry.invoke()
  classifyToolResult() ──→ classify_tool_result ─→ ToolRegistry.classify()
  getSuggestedToolSchemas() → suggest_tool_schemas → ToolRegistry.suggest()

9 Hyena tools: read_file, write_file, list_directory, bash, grep,
               web_search, search_wikipedia, search_arxiv, search_gutenberg

Tool schemas sent with tool_choice: 'auto' (MLX fork supports OpenAI function calling)
Llama-native fallback: parseLlamaToolCalls() catches tool calls from raw text
```

### Navigation / Routing

Hash-based routing in App.svelte. `navGroups` array organizes 21 pages into 6 categories:
- **Inference**: Chat, System Prompt, Self-Awareness, KV Cache
- **Tools**: Tools, Commands, Skills & Minds
- **Training**: Adapters, Training, Self-Program
- **System**: Server, Models, Memory, Projects, Diagnostics, Health Check, Known Issues, Probe, Docs, Advanced Tools, Simple Chat

### State Management

- **App.svelte** (root): Server URL, model selection, local/remote mode, settings
- **ChatPane.svelte**: Messages, streaming state, tool loop state, cache toggles
- **Stores** (Svelte stores): Global remote state, cache bridge, skill bridge
- **localStorage**: System prompts, custom commands, tab state, conversations, theme

---

## Key Types

```typescript
type InferenceEngine = 'mlx' | 'vllm' | 'ollama' | 'openai' | 'direct' | 'unknown';

type ChatMessage =
  | { role: 'user'; content: string }
  | { role: 'assistant'; content: string; tool_calls?: ToolCallInfo[] }
  | { role: 'tool'; content: string; tool_call_id: string }
  | { role: 'system'; content: string };

type StreamEvent =
  | { type: 'token'; content: string }
  | { type: 'usage'; promptTokens; completionTokens; totalTokens; serverTps? }
  | { type: 'tool_calls'; calls: ToolCallInfo[] }
  | { type: 'done' };

interface ToolCallInfo {
  id: string;
  type: 'function';
  function: { name: string; arguments: string };
}
```

---

## Anti-Patterns — What NOT to Do (Lessons Learned the Hard Way)

1. **`fetch()` has NO `timeout` option** — Must use AbortController + setTimeout. The `timeout` property silently does nothing.
2. **`require()` fails in browser context** — Use dynamic `import()` for Tauri API.
3. **Always distinguish 404 from timeout** — 404 = model still loading, timeout = server hung.
4. **UI status must transition** — Initialize as 'connecting', flip to 'streaming' on first token. Never skip states.
5. **MLX fork supports tools** — Don't gate `tool_choice` on engine type. The custom fork at `~/mlx-fork/` has OpenAI function calling.
6. **MLX server is single-threaded** — Health check during active request = false "hung" detection. Skip health checks when streaming is in-flight.
7. **Auto-restart kills terminal-started servers** — Creates duplicate processes, GPU memory contention, death spiral. ChatPane does NOT auto-restart on errors.
8. **CHAT_TIMEOUT of 5s is too short for 70B** — Prompt eval alone takes 7-8s. Current: 20s.
9. **HTTP/1.0 from BaseHTTPServer breaks WKWebView streaming** — Buffers entire response instead of streaming.
10. **Metal Toolchain must be installed** — `xcodebuild -downloadComponent MetalToolchain` (704MB) for direct inference feature.
11. **Health monitor requires 2 consecutive hung detections** before triggering restart.
12. **streamingStatus must start as 'connecting'** not 'streaming' — transition on first token only.

---

## Key Constants

```
DEFAULT_MAX_TOKENS    = 4096
DEFAULT_TEMPERATURE   = 0.7
DEFAULT_CONTEXT_WINDOW = 128,000
DEFAULT_SYSTEM_PROMPT = "You are Camel..."
LOCAL_PORT            = 8741
TPS_HISTORY_SIZE      = 20
CHAT_TIMEOUT          = 20s (AbortController)
MAX_TOOL_ROUNDS       = 8
MAX_CONCURRENT_TOOLS  = 8
HEALTH_CHECK_INTERVAL = 30s
PORT_SCAN_LIST        = [8741, 8000, 11434, 8080]
TOPOLOGY_SIZE         = ~6100 chars / ~900 tokens
```

---

## Custom Infrastructure

- **MLX Fork** at `~/mlx-fork/` — "Socratic Internals Fork" that adds OpenAI function calling support to the MLX server.
- **llm-tools crate** — Custom Rust crate with 9 Hyena tools, registered via `ToolRegistry`. Backend for all tool execution.
- **@mercenary/mlx-server** — npm package for model registry and server management.

---

## Behavior Rules

1. **Answer from this knowledge first**, then verify against source files when precision matters.
2. **Always cite file paths** — e.g., `src/app/components/ChatPane.svelte:handleSend()`.
3. **Never speculate** about code you haven't verified — check the file.
4. **Think in layers**: Frontend (Svelte) → IPC boundary (Tauri invoke) → Backend (Rust) → GPU (Metal/mistralrs).
5. **Respect the anti-patterns** — if someone proposes something that would hit a known footgun, warn them immediately.
6. **Know the data flow** — which store holds what state, where persistence lives, how streaming events propagate.
