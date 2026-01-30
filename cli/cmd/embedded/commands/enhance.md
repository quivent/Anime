# enhance - Massively Parallel Enhancement Engine

Discover, propose, and implement intuitive enhancements that produce the desired result through multi-dimensional parallel analysis and intelligent recommendation synthesis.

Usage: `/enhance [target] [desired-outcome]` - Analyze target system and propose enhancements that intuitively achieve the specified outcome.

**Philosophy:** Enhancement should be intuitive. You describe what you want, and the system proposes concrete enhancements that produce that result. No arbitrary thresholds—just intelligent analysis and actionable proposals.

---

## Quality Levels

```
┌─────────────────┬─────────────────┬─────────────┬─────────────────────────────────────────────┐
│ Level           │ Task Instances  │ Phases      │ Features                                    │
├─────────────────┼─────────────────┼─────────────┼─────────────────────────────────────────────┤
│ quick           │ 12              │ 4           │ Fast discovery, top proposals only,         │
│                 │                 │             │ skip deep verification                      │
├─────────────────┼─────────────────┼─────────────┼─────────────────────────────────────────────┤
│ standard        │ 26              │ 6           │ Full scout coverage, proposal synthesis,    │
│ (default)       │                 │             │ implementation + basic verification         │
├─────────────────┼─────────────────┼─────────────┼─────────────────────────────────────────────┤
│ comprehensive   │ 45+             │ 8           │ Deep analysis, all scout dimensions,        │
│                 │                 │             │ multi-pass verification, evolution planning │
├─────────────────┼─────────────────┼─────────────┼─────────────────────────────────────────────┤
│ maximum         │ 70+             │ 8 + loops   │ Consensus proposals (2-3x scouts),          │
│                 │                 │             │ iterative refinement, outcome optimization  │
└─────────────────┴─────────────────┴─────────────┴─────────────────────────────────────────────┘
```

**Default:** `standard` - Balances discovery depth with execution speed.

---

## Enhancement Dimension Specialists

Automatically activated based on desired outcome, or specify via `--focus [dimensions]`:

```
┌─────────────────┬────────────────────────────────────────────────────────────────────────┐
│ Dimension       │ Task Instances Added to Scout Pool                                     │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ feature         │ Capability Analyst, Feature Designer, Integration Planner,             │
│                 │ API Extender, Workflow Optimizer, Automation Specialist                │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ performance     │ Bottleneck Hunter, Cache Strategist, Query Optimizer,                  │
│                 │ Memory Analyst, Concurrency Expert, Load Balancer                      │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ ux              │ Journey Mapper, Friction Finder, Accessibility Auditor,                │
│                 │ Onboarding Specialist, Error Message Humanizer, Flow Optimizer         │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ security        │ Vulnerability Scanner, Auth Strengthener, Input Validator,             │
│                 │ Secrets Auditor, Dependency Checker, Compliance Mapper                 │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ architecture    │ Pattern Recognizer, Coupling Analyzer, Scalability Planner,            │
│                 │ Debt Identifier, Abstraction Reviewer, Module Boundary Expert          │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ integration     │ API Connector, Webhook Designer, Event System Planner,                 │
│                 │ Third-Party Specialist, Data Sync Architect, Protocol Adapter          │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ simplification  │ Complexity Reducer, Dead Code Hunter, Abstraction Flattener,           │
│                 │ Configuration Consolidator, Dependency Pruner, Over-Engineering Finder │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ auto            │ Analyze desired outcome → Select most relevant dimensions              │
│ (default)       │ Can combine multiple dimensions if outcome spans concerns              │
└─────────────────┴────────────────────────────────────────────────────────────────────────┘
```

---

## Execution Architecture Overview

```
Phase 1:  Understanding ────────┬── State Mapper ─────────────┐
          Pool                  ├── Outcome Parser ───────────┤
                                ├── Gap Analyzer ─────────────┼──→ Enhancement ─ [Parallel]
                                ├── Constraint Scout ─────────┤    Context
                                └── Context Extractor ────────┘
    │
    ▼
Phase 2:  Dimension Detection ──┬── Outcome Classifier ───────┐
                                ├── Keyword Analyzer ─────────┼──→ Dimension ─── [Parallel]
                                └── Priority Scorer ──────────┘    Profile
    │
    ▼
Phase 3:  Scout Pool ───────────┬── Feature Scout ────────────┐
          (7 dimensions)        ├── Performance Scout ────────┤
                                ├── UX Scout ─────────────────┤
                                ├── Security Scout ───────────┼──→ Enhancement ─ [Parallel]
                                ├── Architecture Scout ───────┤    Opportunities
                                ├── Integration Scout ────────┤
                                ├── Simplification Scout ─────┤
                                └── [Dimension Specialists] ──┘
    │
    ├── [maximum quality: duplicate pool for consensus] ──→ Consensus Merger
    │
    ▼
Phase 4:  Proposal Synthesis ───┬── Quick Wins Synthesizer ───┐
          Pool                  ├── Core Enhancement Synth ───┤
                                ├── Strategic Synth ──────────┼──→ Ranked ────── [Parallel]
                                ├── Synergy Finder ───────────┤    Proposals
                                └── Intuition Validator ──────┘
    │
    ▼
Phase 5:  User Presentation ────── Present Proposals ─────────────────────────── [Interactive]
                                │
                      ┌─── Approve ─┴─ Modify ─┴─ Reject ───┐
                      ▼              ▼              ▼
                 Proceed      Re-synthesize      Stop
    │
    ▼
Phase 6:  Implementation Pool ──┬── Enhancement Task 1 ──────┐
          (N parallel Tasks)    ├── Enhancement Task 2 ──────┤
                                ├── Enhancement Task 3 ──────┼──→ Implemented ─ [Parallel]
                                ├── Enhancement Task N ──────┤    Changes
                                └── Coherence Checker ───────┘
    │
    ▼
Phase 7:  Verification Pool ────┬── Integration Verifier ────┐
                                ├── Quality Verifier ────────┤
                                ├── Performance Verifier ────┼──→ Verification ─ [Parallel]
                                ├── Regression Verifier ─────┤    Report
                                └── Outcome Verifier ────────┘
    │
    ▼
Phase 8:  Outcome Synthesis ────┬── Achievement Scorer ──────┐
                                ├── Gap Analyzer ────────────┼──→ Final ──────── [Parallel]
                                ├── Evolution Planner ───────┤    Report
                                └── Next Steps Generator ────┘
```

---

## Phase 1: Parallel Understanding

🎯 **Objective:** Simultaneously understand current state, desired outcome, and the gap between them.

**Task Tool Instance Configuration:**

```yaml
┌─────────────────────┬────────────────────────────────────────────────────────────────┐
│ Task Instance       │ Understanding Mission                                          │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ State-Mapper        │ Deep analysis of what exists now:                              │
│                     │ • File structure and component inventory                       │
│                     │ • Current capabilities and limitations                         │
│                     │ • Technical debt and pain points                               │
│                     │ • Usage patterns and hot paths                                 │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Outcome-Parser      │ Understanding what success looks like:                         │
│                     │ • Explicit goals stated in desired outcome                     │
│                     │ • Implicit requirements (unstated but needed)                  │
│                     │ • Success indicators and verification criteria                 │
│                     │ • User intent behind the request                               │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Gap-Analyzer        │ Identifying the delta:                                         │
│                     │ • What's missing to achieve outcome                            │
│                     │ • What needs to change                                         │
│                     │ • Obstacles and blockers                                       │
│                     │ • Shortest path from here to there                             │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Constraint-Scout    │ Discovering boundaries:                                        │
│                     │ • Technical limitations                                        │
│                     │ • Resource constraints                                         │
│                     │ • Dependency requirements                                      │
│                     │ • Compatibility requirements                                   │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Context-Extractor   │ Environmental understanding:                                   │
│                     │ • How the system is used                                       │
│                     │ • Who the users are                                            │
│                     │ • Integration context                                          │
│                     │ • Historical decisions and rationale                           │
└─────────────────────┴────────────────────────────────────────────────────────────────┘
```

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 5 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Map current state",                                                     │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: Enhancement analysis for {target}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: State Mapper                                                                  │
│                                                                                         │
│     Perform deep analysis of what exists now:                                           │
│     1. Analyze file structure and component inventory                                   │
│     2. Document current capabilities and limitations                                    │
│     3. Identify technical debt and pain points                                          │
│     4. Map usage patterns and hot paths                                                 │
│                                                                                         │
│     OUTPUT FORMAT:                                                                      │
│     - current_state: { files: [], capabilities: [], limitations: [], debt: [] }        │
│     - hot_paths: list of frequently used code paths                                     │
│     - pain_points: list of identified issues                                            │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Parse desired outcome",                                                 │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: Enhancement analysis for {target}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Outcome Parser                                                                │
│                                                                                         │
│     Understand what success looks like:                                                 │
│     1. Extract explicit goals from desired outcome statement                            │
│     2. Infer implicit requirements (unstated but necessary)                             │
│     3. Define success indicators and verification criteria                              │
│     4. Interpret user intent behind the request                                         │
│                                                                                         │
│     OUTPUT FORMAT:                                                                      │
│     - explicit_goals: list of stated objectives                                         │
│     - implicit_requirements: list of inferred needs                                     │
│     - success_criteria: measurable indicators                                           │
│     - user_intent: interpreted meaning                                                  │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Analyze enhancement gap",                                               │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: Enhancement analysis for {target}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Gap Analyzer                                                                  │
│                                                                                         │
│     Identify the delta between current and desired state:                               │
│     1. What's missing to achieve the outcome                                            │
│     2. What existing things need to change                                              │
│     3. Obstacles and blockers preventing success                                        │
│     4. Shortest path from current state to outcome                                      │
│                                                                                         │
│     OUTPUT FORMAT:                                                                      │
│     - missing: list of absent capabilities                                              │
│     - changes_needed: list of modifications required                                    │
│     - blockers: list of obstacles                                                       │
│     - shortest_path: ordered steps to outcome                                           │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout constraints",                                                     │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: Enhancement analysis for {target}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Constraint Scout                                                              │
│                                                                                         │
│     Discover boundaries and limitations:                                                │
│     1. Technical limitations of current architecture                                    │
│     2. Resource constraints (compute, memory, storage)                                  │
│     3. Dependency requirements and version constraints                                  │
│     4. Compatibility requirements with existing systems                                 │
│                                                                                         │
│     OUTPUT FORMAT:                                                                      │
│     - technical_limits: list of hard constraints                                        │
│     - resource_constraints: available resources                                         │
│     - dependencies: required packages/versions                                          │
│     - compatibility: integration requirements                                           │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Extract context",                                                       │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: Enhancement analysis for {target}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Context Extractor                                                             │
│                                                                                         │
│     Understand the environment:                                                         │
│     1. How the system is used in practice                                               │
│     2. Who the users are and their needs                                                │
│     3. Integration context with other systems                                           │
│     4. Historical decisions and their rationale                                         │
│                                                                                         │
│     OUTPUT FORMAT:                                                                      │
│     - usage_patterns: how system is used                                                │
│     - user_profiles: who uses it and why                                                │
│     - integrations: connected systems                                                   │
│     - historical_context: past decisions and reasons                                    │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → MERGE INTO enhancement_context                                              │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

**Output:** `enhancement_context` - Unified understanding enabling intuitive proposals

---

## Phase 2: Dimension Detection & Priority Scoring

🔍 **Objective:** Determine which enhancement dimensions are most relevant to the desired outcome.

**Detection Task Tool Pool:**

```yaml
┌─────────────────────┬────────────────────────────────────────────────────────────────┐
│ Task Instance       │ Detection Focus                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Outcome-Classifier  │ Map desired outcome to enhancement dimensions:                 │
│                     │ "faster" → performance | "easier" → ux | "safer" → security    │
│                     │ "cleaner" → architecture + simplification                      │
│                     │ "more capable" → feature + integration                         │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Keyword-Analyzer    │ Extract dimension signals from outcome description:            │
│                     │ • Performance words: fast, quick, responsive, efficient        │
│                     │ • UX words: easy, intuitive, simple, friendly, clear           │
│                     │ • Security words: safe, secure, protected, private             │
│                     │ • Architecture words: clean, organized, maintainable           │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│ Priority-Scorer     │ Rank dimensions by outcome relevance:                          │
│                     │ • Primary: Direct outcome contributors                         │
│                     │ • Secondary: Supporting dimensions                             │
│                     │ • Tertiary: Nice-to-have if discovered                         │
└─────────────────────┴────────────────────────────────────────────────────────────────┘
```

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 3 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Classify outcome dimensions",                                           │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Outcome Classifier                                                            │
│                                                                                         │
│     Map the desired outcome to enhancement dimensions:                                  │
│     - feature: capability additions, new functionality                                  │
│     - performance: speed, efficiency, resource usage                                    │
│     - ux: user experience, ease of use, intuitiveness                                   │
│     - security: protection, safety, privacy                                             │
│     - architecture: structure, maintainability, patterns                                │
│     - integration: connections, APIs, external systems                                  │
│     - simplification: reduction, cleanup, streamlining                                  │
│                                                                                         │
│     OUTPUT: dimension_classification with weights 0.0-1.0                               │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Analyze outcome keywords",                                              │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Keyword Analyzer                                                              │
│                                                                                         │
│     Extract dimension signals from outcome description:                                 │
│     - Performance: fast, quick, responsive, efficient, optimized, speedy                │
│     - UX: easy, intuitive, simple, friendly, clear, accessible, usable                  │
│     - Security: safe, secure, protected, private, authenticated, encrypted             │
│     - Architecture: clean, organized, maintainable, modular, structured                 │
│     - Feature: add, new, capability, function, support, enable                          │
│     - Integration: connect, sync, import, export, API, webhook                          │
│     - Simplification: remove, reduce, simplify, eliminate, consolidate                  │
│                                                                                         │
│     OUTPUT: keyword_signals with detected words and confidence                          │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Score dimension priority",                                              │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Priority Scorer                                                               │
│                                                                                         │
│     Rank dimensions by outcome relevance:                                               │
│     - Primary (weight 1.0): Direct outcome contributors                                 │
│     - Secondary (weight 0.5): Supporting dimensions                                     │
│     - Tertiary (weight 0.2): Nice-to-have if discovered                                 │
│                                                                                         │
│     OUTPUT: priority_ranking with:                                                      │
│     - primary_dimensions: list                                                          │
│     - secondary_dimensions: list                                                        │
│     - tertiary_dimensions: list                                                         │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → MERGE INTO dimension_profile                                                │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

**Dimension Priority Matrix:**

```
                    Outcome Keywords
                    │
    ┌───────────────┼───────────────────────────────────────────────────────┐
    │               │ faster  easier  safer  cleaner  capable  reliable    │
    ├───────────────┼───────────────────────────────────────────────────────┤
    │ feature       │   ○       ○       ○       ○        ●        ○        │
    │ performance   │   ●       ○       ○       ○        ○        ◐        │
    │ ux            │   ◐       ●       ○       ◐        ○        ○        │
    │ security      │   ○       ○       ●       ○        ○        ◐        │
    │ architecture  │   ○       ◐       ○       ●        ◐        ●        │
    │ integration   │   ○       ○       ○       ○        ●        ○        │
    │ simplification│   ◐       ●       ○       ●        ○        ◐        │
    └───────────────┴───────────────────────────────────────────────────────┘

    ● = Primary (weight 1.0)    ◐ = Secondary (weight 0.5)    ○ = Tertiary (weight 0.2)
```

---

## Phase 3: Parallel Enhancement Scouting

🔬 **Objective:** Simultaneously discover enhancement opportunities across all relevant dimensions.

**Scout Task Tool Pool:**

```yaml
┌─────────────────────┬────────────────────────────────────────────────────────────────┐
│ Task Instance       │ Discovery Mission                                              │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Feature-Scout       │ Hunt for capability enhancements:                              │
│ 🎯                  │ • New features that directly produce the outcome              │
│                     │ • Missing functionality users would expect                     │
│                     │ • Automation opportunities                                     │
│                     │ • Workflow improvements                                        │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Performance-Scout   │ Hunt for speed and efficiency:                                 │
│ ⚡                  │ • Bottlenecks causing slowness                                 │
│                     │ • Caching opportunities                                        │
│                     │ • Query optimization targets                                   │
│                     │ • Resource usage improvements                                  │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ UX-Scout            │ Hunt for experience improvements:                              │
│ 👤                  │ • Friction points in user journeys                            │
│                     │ • Confusing interfaces or flows                                │
│                     │ • Missing feedback or guidance                                 │
│                     │ • Accessibility gaps                                           │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Security-Scout      │ Hunt for protection improvements:                              │
│ 🛡️                  │ • Vulnerability patterns                                       │
│                     │ • Authentication/authorization gaps                            │
│                     │ • Input validation weaknesses                                  │
│                     │ • Dependency risks                                             │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Architecture-Scout  │ Hunt for structural improvements:                              │
│ 🏗️                  │ • Code organization opportunities                              │
│                     │ • Coupling/cohesion issues                                     │
│                     │ • Scalability blockers                                         │
│                     │ • Pattern improvements                                         │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Integration-Scout   │ Hunt for connection opportunities:                             │
│ 🔌                  │ • External service integrations                                │
│                     │ • API enhancements                                             │
│                     │ • Event/webhook opportunities                                  │
│                     │ • Data synchronization improvements                            │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Simplification-     │ Hunt for things to REMOVE or REDUCE:                           │
│ Scout 🧹            │ • Over-engineered solutions                                    │
│                     │ • Dead code and unused features                                │
│                     │ • Unnecessary complexity                                       │
│                     │ • Redundant abstractions                                       │
│                     │                                                                │
└─────────────────────┴────────────────────────────────────────────────────────────────┘
```

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 7 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout feature enhancements",                                            │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Feature Scout 🎯                                                              │
│                                                                                         │
│     Hunt for capability enhancements in {target}:                                       │
│     1. New features that directly produce the desired outcome                           │
│     2. Missing functionality users would expect                                         │
│     3. Automation opportunities                                                         │
│     4. Workflow improvements                                                            │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name: short descriptive name                                                      │
│     - what: concrete description                                                        │
│     - why: how it contributes to outcome                                                │
│     - how: implementation approach                                                      │
│     - impact: low/medium/high                                                           │
│     - effort: low/medium/high                                                           │
│     - intuitive_score: 0-100 (how obvious/natural this feels)                           │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout performance enhancements",                                        │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Performance Scout ⚡                                                          │
│                                                                                         │
│     Hunt for speed and efficiency improvements in {target}:                             │
│     1. Bottlenecks causing slowness                                                     │
│     2. Caching opportunities                                                            │
│     3. Query optimization targets                                                       │
│     4. Resource usage improvements                                                      │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - performance_metric: what to measure                                               │
│     - expected_improvement: quantified if possible                                      │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout UX enhancements",                                                 │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: UX Scout 👤                                                                   │
│                                                                                         │
│     Hunt for experience improvements in {target}:                                       │
│     1. Friction points in user journeys                                                 │
│     2. Confusing interfaces or flows                                                    │
│     3. Missing feedback or guidance                                                     │
│     4. Accessibility gaps                                                               │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - user_benefit: direct improvement to user experience                               │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout security enhancements",                                           │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Security Scout 🛡️                                                             │
│                                                                                         │
│     Hunt for protection improvements in {target}:                                       │
│     1. Vulnerability patterns                                                           │
│     2. Authentication/authorization gaps                                                │
│     3. Input validation weaknesses                                                      │
│     4. Dependency risks                                                                 │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - risk_addressed: what threat is mitigated                                          │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout architecture enhancements",                                       │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Architecture Scout 🏗️                                                         │
│                                                                                         │
│     Hunt for structural improvements in {target}:                                       │
│     1. Code organization opportunities                                                  │
│     2. Coupling/cohesion issues                                                         │
│     3. Scalability blockers                                                             │
│     4. Pattern improvements                                                             │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - maintainability_gain: long-term benefit                                           │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout integration enhancements",                                        │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Integration Scout 🔌                                                          │
│                                                                                         │
│     Hunt for connection opportunities in {target}:                                      │
│     1. External service integrations                                                    │
│     2. API enhancements                                                                 │
│     3. Event/webhook opportunities                                                      │
│     4. Data synchronization improvements                                                │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - integration_benefit: what becomes possible                                        │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Scout simplification enhancements",                                     │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Simplification Scout 🧹                                                       │
│                                                                                         │
│     Hunt for things to REMOVE or REDUCE in {target}:                                    │
│     1. Over-engineered solutions                                                        │
│     2. Dead code and unused features                                                    │
│     3. Unnecessary complexity                                                           │
│     4. Redundant abstractions                                                           │
│                                                                                         │
│     For each enhancement found, provide:                                                │
│     - name, what, why, how, impact, effort, intuitive_score                             │
│     - lines_removed: approximate reduction                                              │
│     - complexity_reduction: what becomes simpler                                        │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → COLLECT INTO enhancement_opportunities                                      │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

**Scout Output Schema:**

```json
{
  "dimension": "performance",
  "enhancements": [
    {
      "name": "Query Result Caching",
      "what": "Add Redis cache for frequently-accessed queries",
      "why": "Eliminates 60% of database hits, directly speeds up responses",
      "how": "Implement cache-aside pattern with 5-minute TTL",
      "impact": "high",
      "effort": "medium",
      "outcome_contribution": "Reduces avg response time by ~200ms",
      "intuitive_score": 92,
      "dependencies": ["redis connection"],
      "verification": "Response time < 100ms for cached queries"
    }
  ],
  "dimension_summary": "3 high-impact performance opportunities found"
}
```

---

## Phase 4: Proposal Synthesis

💡 **Objective:** Synthesize scout findings into coherent, actionable, prioritized proposals.

**Synthesis Task Tool Pool:**

```yaml
┌─────────────────────┬────────────────────────────────────────────────────────────────┐
│ Task Instance       │ Synthesis Focus                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Quick-Wins          │ Find immediately actionable enhancements:                      │
│ Synthesizer ⚡      │ • Low effort, high intuitive value                            │
│                     │ • Can be done independently                                    │
│                     │ • Visible improvement quickly                                  │
│                     │ • Low risk of side effects                                     │
│                     │                                                                │
│                     │ Criteria: effort ≤ medium AND impact ≥ medium                 │
│                     │           AND intuitive_score ≥ 80                             │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Core Enhancement    │ Find enhancements that directly produce the outcome:           │
│ Synthesizer 🎯      │ • High outcome contribution                                   │
│                     │ • Worth the effort investment                                  │
│                     │ • Forms the backbone of improvement                            │
│                     │ • May require coordination                                     │
│                     │                                                                │
│                     │ Criteria: outcome_contribution = high                          │
│                     │           AND aligns with primary dimensions                   │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Strategic           │ Find foundational improvements:                                │
│ Synthesizer 🏛️      │ • Enable future enhancements                                  │
│                     │ • Reduce long-term complexity                                  │
│                     │ • Build sustainable foundation                                 │
│                     │ • Pay off over time                                            │
│                     │                                                                │
│                     │ Criteria: enables future improvements                          │
│                     │           OR reduces maintenance burden                        │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Synergy Finder 🔗   │ Find enhancements that amplify each other:                     │
│                     │ • Combinations with multiplicative effect                      │
│                     │ • Shared dependencies (implement together)                     │
│                     │ • Sequential enablement (A enables B)                          │
│                     │ • Parallel independence (can be done together)                 │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Intuition           │ Validate proposals feel right:                                 │
│ Validator 🧠        │ • Would a user naturally expect this?                         │
│                     │ • Is the path straightforward?                                 │
│                     │ • Does it feel like the "obvious" solution?                    │
│                     │ • Will success be apparent?                                    │
│                     │                                                                │
└─────────────────────┴────────────────────────────────────────────────────────────────┘
```

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 5 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Synthesize quick wins",                                                 │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     OPPORTUNITIES: {enhancement_opportunities}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Quick Wins Synthesizer ⚡                                                     │
│                                                                                         │
│     Find immediately actionable enhancements:                                           │
│     - Low effort, high intuitive value                                                  │
│     - Can be done independently                                                         │
│     - Visible improvement quickly                                                       │
│     - Low risk of side effects                                                          │
│                                                                                         │
│     Filter criteria:                                                                    │
│     - effort ≤ medium AND impact ≥ medium AND intuitive_score ≥ 80                     │
│                                                                                         │
│     OUTPUT: quick_wins list with priority order                                         │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Synthesize core enhancements",                                          │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     OPPORTUNITIES: {enhancement_opportunities}                                          │
│     DIMENSION PROFILE: {dimension_profile}                                              │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Core Enhancement Synthesizer 🎯                                               │
│                                                                                         │
│     Find enhancements that directly produce the outcome:                                │
│     - High outcome contribution                                                         │
│     - Worth the effort investment                                                       │
│     - Forms the backbone of improvement                                                 │
│     - May require coordination                                                          │
│                                                                                         │
│     Filter criteria:                                                                    │
│     - outcome_contribution = high AND aligns with primary dimensions                    │
│                                                                                         │
│     OUTPUT: core_enhancements list with dependencies                                    │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Synthesize strategic enhancements",                                     │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     OPPORTUNITIES: {enhancement_opportunities}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Strategic Synthesizer 🏛️                                                      │
│                                                                                         │
│     Find foundational improvements:                                                     │
│     - Enable future enhancements                                                        │
│     - Reduce long-term complexity                                                       │
│     - Build sustainable foundation                                                      │
│     - Pay off over time                                                                 │
│                                                                                         │
│     Filter criteria:                                                                    │
│     - enables future improvements OR reduces maintenance burden                         │
│                                                                                         │
│     OUTPUT: strategic_enhancements with future_value rationale                          │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Find enhancement synergies",                                            │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     OPPORTUNITIES: {enhancement_opportunities}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Synergy Finder 🔗                                                             │
│                                                                                         │
│     Find enhancements that amplify each other:                                          │
│     - Combinations with multiplicative effect                                           │
│     - Shared dependencies (implement together)                                          │
│     - Sequential enablement (A enables B)                                               │
│     - Parallel independence (can be done together)                                      │
│                                                                                         │
│     OUTPUT: synergies list with:                                                        │
│     - combination: [enhancement_ids]                                                    │
│     - synergy_type: multiplicative/shared/sequential/parallel                           │
│     - combined_benefit: what the combination achieves                                   │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Validate proposal intuition",                                           │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     OPPORTUNITIES: {enhancement_opportunities}                                          │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Intuition Validator 🧠                                                        │
│                                                                                         │
│     Validate that proposals feel right:                                                 │
│     - Would a user naturally expect this?                                               │
│     - Is the path straightforward?                                                      │
│     - Does it feel like the "obvious" solution?                                         │
│     - Will success be apparent?                                                         │
│                                                                                         │
│     For each enhancement, score:                                                        │
│     - intuitive_fit: 0-100                                                              │
│     - obvious_factor: does it feel like "of course, that's right"                       │
│     - clarity_score: how clear is the improvement                                       │
│                                                                                         │
│     Flag any that feel forced, over-engineered, or non-obvious                          │
│                                                                                         │
│     OUTPUT: validated_proposals with intuition scores                                   │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → MERGE INTO ranked_proposals                                                 │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

**Proposal Priority Matrix:**

```
                                        EFFORT
                        ┌─────────────────────────────────────┐
                        │    Low      Medium      High        │
              ┌─────────┼─────────────────────────────────────┤
              │         │                                     │
              │  High   │  ⭐ QUICK    🎯 CORE    🏛️ STRATEGIC │
              │         │    WIN                              │
    IMPACT    │─────────┼─────────────────────────────────────┤
              │         │                                     │
              │  Medium │  ⭐ QUICK    📋 BACKLOG  📋 BACKLOG  │
              │         │    WIN                              │
              │─────────┼─────────────────────────────────────┤
              │         │                                     │
              │  Low    │  📋 BACKLOG  ❌ SKIP     ❌ SKIP     │
              │         │                                     │
              └─────────┴─────────────────────────────────────┘
```

---

## Phase 5: User Presentation

📋 **Objective:** Present synthesized proposals for user selection.

**Presentation Format:**

```
╔══════════════════════════════════════════════════════════════════════════════════╗
║                           ENHANCEMENT PROPOSALS                                   ║
║                                                                                   ║
║  Target: ./myproject                                                              ║
║  Desired Outcome: "faster response times"                                         ║
║  Dimensions Activated: performance (primary), architecture (secondary)            ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  ⭐ QUICK WINS (Do These First)                                                   ║
║  ─────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  [1] ⚡ Query Result Caching                                                      ║
║      │                                                                            ║
║      ├── What: Add Redis cache for frequently-accessed queries                    ║
║      ├── Why:  Eliminates 60% of database hits                                    ║
║      ├── Impact: HIGH    Effort: MEDIUM    Intuitive: 92/100                      ║
║      └── Outcome: Reduces avg response time by ~200ms                             ║
║                                                                                   ║
║  [2] ⚡ Response Compression                                                      ║
║      │                                                                            ║
║      ├── What: Enable gzip compression for API responses                          ║
║      ├── Why:  Reduces payload size by 70%                                        ║
║      ├── Impact: MEDIUM  Effort: LOW       Intuitive: 95/100                      ║
║      └── Outcome: Faster transfer, especially on slow connections                 ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  🎯 CORE ENHANCEMENTS (These Produce The Outcome)                                 ║
║  ─────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  [3] 🎯 Database Query Optimization                                               ║
║      │                                                                            ║
║      ├── What: Add indexes, optimize N+1 queries, batch operations                ║
║      ├── Why:  Addresses root cause of slow responses                             ║
║      ├── Impact: HIGH    Effort: MEDIUM    Intuitive: 88/100                      ║
║      └── Outcome: 3x improvement in database operation speed                      ║
║                                                                                   ║
║  [4] 🎯 Async Background Processing                                               ║
║      │                                                                            ║
║      ├── What: Move non-critical work to background jobs                          ║
║      ├── Why:  User doesn't wait for operations they don't need                   ║
║      ├── Impact: HIGH    Effort: MEDIUM    Intuitive: 85/100                      ║
║      └── Outcome: Perceived response time cut in half                             ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  🏛️ STRATEGIC (Foundation For Future)                                            ║
║  ─────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  [5] 🏛️ Connection Pooling Optimization                                          ║
║      │                                                                            ║
║      ├── What: Implement proper connection pooling and management                 ║
║      ├── Why:  Enables scaling and prevents connection exhaustion                 ║
║      ├── Impact: MEDIUM  Effort: MEDIUM    Intuitive: 75/100                      ║
║      └── Outcome: Consistent performance under load                               ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  🔗 SYNERGIES                                                                     ║
║  ─────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  • [1] + [3] together = 80% reduction (multiplicative effect)                     ║
║  • [4] enables future horizontal scaling                                          ║
║  • [1], [2] can be implemented in parallel (no dependencies)                      ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  SELECT PROPOSALS: Enter numbers (e.g., "1,2,3" or "all" or "quick-wins")         ║
║                                                                                   ║
╚══════════════════════════════════════════════════════════════════════════════════╝
```

---

## Phase 6: Massive Parallel Implementation

🔧 **Objective:** Simultaneously implement all approved enhancements.

**Implementation Task Tool Pool:**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         PARALLEL IMPLEMENTATION                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  FOR EACH enhancement IN approved_enhancements:                                  │
│                                                                                  │
│      Task(                                                                       │
│        description: "Implement {enhancement.name}",                              │
│        subagent_type: "general-purpose",                                         │
│        prompt: """                                                               │
│          CONTEXT: {enhancement_context}                                          │
│          TARGET: {target}                                                        │
│          ROLE: Enhancement Implementer                                           │
│                                                                                  │
│          ENHANCEMENT TO IMPLEMENT:                                               │
│          - Name: {enhancement.name}                                              │
│          - What: {enhancement.what}                                              │
│          - How: {enhancement.how}                                                │
│          - Expected Outcome: {enhancement.outcome_contribution}                  │
│                                                                                  │
│          IMPLEMENTATION PRINCIPLES:                                              │
│          1. SIMPLEST SOLUTION - Don't over-engineer                              │
│          2. OUTCOME FOCUSED - Every line serves the result                       │
│          3. CLEAN INTEGRATION - Work with existing patterns                      │
│          4. VERIFIABLE SUCCESS - Include verification steps                      │
│                                                                                  │
│          OUTPUT:                                                                 │
│          - files_modified: list of changed files                                 │
│          - changes_summary: brief description                                    │
│          - verification_steps: how to confirm it works                           │
│        """                                                                       │
│      )                                                                           │
│                                                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Task      │  │   Task      │  │   Task      │  │   Task      │             │
│  │ Instance 1  │  │ Instance 2  │  │ Instance 3  │  │ Instance N  │             │
│  │             │  │             │  │             │  │             │             │
│  │  Implement  │  │  Implement  │  │  Implement  │  │  Implement  │   ···       │
│  │   Caching   │  │ Compression │  │   Query     │  │   Async     │             │
│  │             │  │             │  │    Opt      │  │  Processing │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                │                    │
│         └────────────────┴────────────────┴────────────────┘                    │
│                                    │                                             │
│                                    ▼                                             │
│                        ┌─────────────────────┐                                   │
│      Task(             │  Coherence Checker  │                                   │
│        description:    │                     │                                   │
│          "Check        │  • No conflicts     │                                   │
│           coherence",  │  • Clean integration│                                   │
│        subagent_type:  │  • Consistent style │                                   │
│          "general-     └─────────────────────┘                                   │
│           purpose",                                                              │
│        prompt: "..."                                                             │
│      )                                                                           │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Coherence Checker Task:**

```
Task(
  description: "Check implementation coherence",
  subagent_type: "general-purpose",
  prompt: """
    CONTEXT: {enhancement_context}
    IMPLEMENTATIONS: {all_implementation_results}
    ROLE: Coherence Checker

    Verify all implementations work together:
    1. No conflicts between changes (same files, conflicting patterns)
    2. Clean import/export relationships
    3. Consistent coding style across enhancements
    4. No circular dependencies introduced
    5. All integrations properly connected

    OUTPUT:
    - coherence_status: pass/fail
    - conflicts: list of any conflicts found
    - resolution_steps: how to fix conflicts if any
  """
)
```

**Implementation Principles:**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        IMPLEMENTATION PRINCIPLES                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  1. SIMPLEST SOLUTION                                                            │
│     ├── Don't over-engineer                                                      │
│     ├── Prefer straightforward approaches                                        │
│     └── If it feels complex, find a simpler way                                  │
│                                                                                  │
│  2. OUTCOME FOCUSED                                                              │
│     ├── Every line serves the desired result                                     │
│     ├── Remove anything that doesn't contribute                                  │
│     └── Verify the enhancement produces its stated effect                        │
│                                                                                  │
│  3. CLEAN INTEGRATION                                                            │
│     ├── Work with existing patterns                                              │
│     ├── Minimal disruption to working code                                       │
│     └── Easy to understand what changed and why                                  │
│                                                                                  │
│  4. VERIFIABLE SUCCESS                                                           │
│     ├── Include verification steps                                               │
│     ├── Make success obvious                                                     │
│     └── Provide evidence the outcome is achieved                                 │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 7: Parallel Verification

✅ **Objective:** Simultaneously verify all enhancements produce the desired outcome.

**Verification Task Tool Pool:**

```yaml
┌─────────────────────┬────────────────────────────────────────────────────────────────┐
│ Task Instance       │ Verification Question                                          │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Integration         │ Do the enhancements work together?                             │
│ Verifier 🔗         │ • No conflicts between changes                                │
│                     │ • Clean import/export relationships                            │
│                     │ • Consistent patterns across enhancements                      │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Quality             │ Is the implementation solid?                                   │
│ Verifier ✨         │ • Code quality meets standards                                │
│                     │ • Error handling comprehensive                                 │
│                     │ • Edge cases covered                                           │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Performance         │ Any degradation introduced?                                    │
│ Verifier ⚡         │ • No new bottlenecks                                          │
│                     │ • Resource usage acceptable                                    │
│                     │ • Meets performance targets                                    │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ Regression          │ Did we break anything?                                         │
│ Verifier 🛡️         │ • Existing functionality preserved                            │
│                     │ • Tests still pass                                             │
│                     │ • No unintended side effects                                   │
│                     │                                                                │
├─────────────────────┼────────────────────────────────────────────────────────────────┤
│                     │                                                                │
│ OUTCOME             │ ★ DID WE PRODUCE THE DESIRED RESULT? ★                        │
│ VERIFIER 🎯         │                                                               │
│                     │ This is the most important verification:                       │
│ (CRITICAL)          │                                                                │
│                     │ • Does the system now achieve the stated outcome?              │
│                     │ • Is the improvement measurable and obvious?                   │
│                     │ • Would the user say "yes, this is what I wanted"?             │
│                     │ • What evidence proves outcome achievement?                    │
│                     │                                                                │
└─────────────────────┴────────────────────────────────────────────────────────────────┘
```

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 5 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Verify integration",                                                    │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     ROLE: Integration Verifier 🔗                                                       │
│                                                                                         │
│     Verify: Do the enhancements work together?                                          │
│     - Check for conflicts between changes                                               │
│     - Verify clean import/export relationships                                          │
│     - Confirm consistent patterns across enhancements                                   │
│                                                                                         │
│     OUTPUT: integration_status with pass/fail and details                               │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Verify quality",                                                        │
│   subagent_type: "Explore",                                                             │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     ROLE: Quality Verifier ✨                                                           │
│                                                                                         │
│     Verify: Is the implementation solid?                                                │
│     - Code quality meets standards                                                      │
│     - Error handling is comprehensive                                                   │
│     - Edge cases are covered                                                            │
│                                                                                         │
│     OUTPUT: quality_status with pass/fail and issues                                    │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Verify performance",                                                    │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     ROLE: Performance Verifier ⚡                                                       │
│                                                                                         │
│     Verify: Any degradation introduced?                                                 │
│     - No new bottlenecks                                                                │
│     - Resource usage acceptable                                                         │
│     - Meets performance targets                                                         │
│                                                                                         │
│     OUTPUT: performance_status with metrics                                             │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Verify no regressions",                                                 │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     ROLE: Regression Verifier 🛡️                                                        │
│                                                                                         │
│     Verify: Did we break anything?                                                      │
│     - Existing functionality preserved                                                  │
│     - Tests still pass (run if available)                                               │
│     - No unintended side effects                                                        │
│                                                                                         │
│     OUTPUT: regression_status with test results                                         │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Verify outcome achieved",                                               │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Outcome Verifier 🎯 (CRITICAL)                                                │
│                                                                                         │
│     ★ THE MOST IMPORTANT VERIFICATION ★                                                 │
│                                                                                         │
│     Verify: Did we produce the desired result?                                          │
│     - Does the system now achieve the stated outcome?                                   │
│     - Is the improvement measurable and obvious?                                        │
│     - Would the user say "yes, this is what I wanted"?                                  │
│     - What evidence proves outcome achievement?                                         │
│                                                                                         │
│     OUTPUT:                                                                             │
│     - outcome_achieved: true/false                                                      │
│     - achievement_percentage: 0-100                                                     │
│     - evidence: concrete proof of improvement                                           │
│     - before_after: measurable comparison                                               │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → MERGE INTO verification_report                                              │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 8: Outcome Synthesis & Evolution

📊 **Objective:** Report on outcome achievement and propose natural evolution.

**Parallel Task Tool Execution:**

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ SPAWN ALL IN PARALLEL (single message with 4 Task tool calls)                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│ Task(                                                                                   │
│   description: "Score achievement",                                                     │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     VERIFICATION: {verification_report}                                                 │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Achievement Scorer                                                            │
│                                                                                         │
│     Quantify the outcome achievement:                                                   │
│     - Overall achievement percentage                                                    │
│     - Per-enhancement contribution                                                      │
│     - Before/after metrics comparison                                                   │
│     - User satisfaction prediction                                                      │
│                                                                                         │
│     OUTPUT: achievement_score with detailed breakdown                                   │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Analyze remaining gaps",                                                │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     VERIFICATION: {verification_report}                                                 │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Gap Analyzer                                                                  │
│                                                                                         │
│     Identify what's still missing:                                                      │
│     - Remaining gaps to 100% achievement                                                │
│     - Unaddressed aspects of the outcome                                                │
│     - Areas that could be further improved                                              │
│                                                                                         │
│     OUTPUT: remaining_gaps with priority                                                │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Plan evolution path",                                                   │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     IMPLEMENTATIONS: {implementation_results}                                           │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Evolution Planner                                                             │
│                                                                                         │
│     What naturally comes next:                                                          │
│     - Enhancements enabled by current work                                              │
│     - Logical progressions from here                                                    │
│     - Foundation that's now in place                                                    │
│                                                                                         │
│     OUTPUT: evolution_path with natural next steps                                      │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
│ Task(                                                                                   │
│   description: "Generate next steps",                                                   │
│   subagent_type: "general-purpose",                                                     │
│   prompt: """                                                                           │
│     CONTEXT: {enhancement_context}                                                      │
│     VERIFICATION: {verification_report}                                                 │
│     DESIRED OUTCOME: {outcome}                                                          │
│     ROLE: Next Steps Generator                                                          │
│                                                                                         │
│     Recommend actionable next steps:                                                    │
│     - Immediate actions (verify in production)                                          │
│     - Short-term improvements (quick wins discovered)                                   │
│     - Medium-term evolution (build on foundation)                                       │
│                                                                                         │
│     OUTPUT: next_steps with priority and rationale                                      │
│   """                                                                                   │
│ )                                                                                       │
│                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ AWAIT ALL → MERGE INTO final_report                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

**Final Report Format:**

```
╔══════════════════════════════════════════════════════════════════════════════════╗
║                           ENHANCEMENT REPORT                                      ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  Target: ./myproject                                                              ║
║  Desired Outcome: "faster response times"                                         ║
║  Enhancement Duration: 4m 32s                                                     ║
║  Parallel Task Instances Used: 28                                                 ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║                        ★ OUTCOME ACHIEVEMENT ★                                    ║
║                                                                                   ║
║  ┌─────────────────────────────────────────────────────────────────────────────┐ ║
║  │                                                                              │ ║
║  │   BEFORE                          AFTER                                      │ ║
║  │   ══════                          ═════                                      │ ║
║  │                                                                              │ ║
║  │   Avg Response: 450ms      →      Avg Response: 95ms                         │ ║
║  │   P95 Response: 1200ms     →      P95 Response: 180ms                        │ ║
║  │   DB Queries/req: 12       →      DB Queries/req: 3                          │ ║
║  │   Cache Hit Rate: 0%       →      Cache Hit Rate: 78%                        │ ║
║  │                                                                              │ ║
║  │   ════════════════════════════════════════════════════════════════════════  │ ║
║  │                                                                              │ ║
║  │   RESULT: ✅ 79% FASTER - OUTCOME ACHIEVED                                   │ ║
║  │                                                                              │ ║
║  └─────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  ENHANCEMENTS APPLIED                                                             ║
║  ────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  ✅ Query Result Caching                                                          ║
║     ├── Status: Implemented                                                       ║
║     ├── Impact: Reduced DB load by 60%                                            ║
║     └── Files: src/cache/redis.ts, src/api/middleware.ts                          ║
║                                                                                   ║
║  ✅ Response Compression                                                          ║
║     ├── Status: Implemented                                                       ║
║     ├── Impact: Payload size reduced 70%                                          ║
║     └── Files: src/server/compression.ts                                          ║
║                                                                                   ║
║  ✅ Database Query Optimization                                                   ║
║     ├── Status: Implemented                                                       ║
║     ├── Impact: Query time reduced 75%                                            ║
║     └── Files: src/db/queries.ts, migrations/add_indexes.sql                      ║
║                                                                                   ║
║  ✅ Async Background Processing                                                   ║
║     ├── Status: Implemented                                                       ║
║     ├── Impact: Non-blocking response path                                        ║
║     └── Files: src/jobs/processor.ts, src/api/handlers.ts                         ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  VERIFICATION RESULTS                                                             ║
║  ────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  ┌────────────────────┬──────────┬──────────────────────────────────────────┐   ║
║  │ Dimension          │ Status   │ Notes                                    │   ║
║  ├────────────────────┼──────────┼──────────────────────────────────────────┤   ║
║  │ Integration        │ ✅ PASS  │ All enhancements work together           │   ║
║  │ Quality            │ ✅ PASS  │ Code meets standards                     │   ║
║  │ Performance        │ ✅ PASS  │ No degradation, targets exceeded         │   ║
║  │ Regression         │ ✅ PASS  │ All existing tests pass                  │   ║
║  │ Outcome            │ ✅ PASS  │ 79% faster - outcome achieved            │   ║
║  └────────────────────┴──────────┴──────────────────────────────────────────┘   ║
║                                                                                   ║
╠══════════════════════════════════════════════════════════════════════════════════╣
║                                                                                   ║
║  NATURAL EVOLUTION                                                                ║
║  ────────────────────────────────────────────────────────────────────────────── ║
║                                                                                   ║
║  What comes next intuitively:                                                     ║
║                                                                                   ║
║  1. Predictive Prefetching                                                        ║
║     └── Anticipate user requests, pre-warm cache                                  ║
║     └── Why natural: Builds on caching foundation                                 ║
║                                                                                   ║
║  2. Request Coalescing                                                            ║
║     └── Deduplicate simultaneous identical requests                               ║
║     └── Why natural: Complements async processing                                 ║
║                                                                                   ║
║  3. Edge Caching                                                                  ║
║     └── Move cache closer to users geographically                                 ║
║     └── Why natural: Next level of caching strategy                               ║
║                                                                                   ║
╚══════════════════════════════════════════════════════════════════════════════════╝
```

---

## Command Invocation

```bash
# Basic - describe what you want
/enhance ./myproject "faster response times"
/enhance ./api "easier to use"
/enhance ./app "more intuitive navigation"
/enhance ./service "handle more traffic"
/enhance ./codebase "cleaner and more maintainable"

# With dimension focus
/enhance ./myproject "better" --focus performance,ux
/enhance ./api "more secure" --focus security
/enhance ./app "simpler" --focus simplification,architecture

# Quality levels
/enhance ./project "faster" --quality quick           # Fast discovery
/enhance ./project "faster" --quality standard        # Default
/enhance ./project "faster" --quality comprehensive   # Deep analysis
/enhance ./project "faster" --quality maximum         # Consensus + iteration

# Exploration mode (no specific outcome)
/enhance ./codebase --explore                         # "What could be better?"

# Control modes
/enhance ./app "cleaner" --propose-only               # Just show proposals
/enhance ./app "cleaner" --auto-approve               # Implement all
/enhance ./app "cleaner" --interactive                # Step by step
```

---

## Parameters & Options

```
┌─────────────────────┬─────────────┬─────────────┬──────────────────────────────────┐
│ Parameter           │ Type        │ Default     │ Description                      │
├─────────────────────┼─────────────┼─────────────┼──────────────────────────────────┤
│ target              │ path        │ .           │ What to enhance                  │
│ outcome             │ string      │ -           │ Desired result in plain language │
├─────────────────────┼─────────────┼─────────────┼──────────────────────────────────┤
│ --focus             │ list        │ auto        │ Dimensions to emphasize          │
│ --quality           │ enum        │ standard    │ Analysis depth                   │
│ --explore           │ flag        │ false       │ Discovery mode                   │
├─────────────────────┼─────────────┼─────────────┼──────────────────────────────────┤
│ --propose-only      │ flag        │ false       │ Show proposals only              │
│ --auto-approve      │ flag        │ false       │ Implement all proposals          │
│ --interactive       │ flag        │ false       │ Step-by-step approval            │
├─────────────────────┼─────────────┼─────────────┼──────────────────────────────────┤
│ --verbose           │ flag        │ false       │ Detailed output                  │
│ --dry-run           │ flag        │ false       │ Preview without changes          │
└─────────────────────┴─────────────┴─────────────┴──────────────────────────────────┘
```

---

## Intuitive Enhancement Philosophy

### What Makes an Enhancement Intuitive?

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                  │
│   INTUITIVE ENHANCEMENTS                                                         │
│   ══════════════════════                                                         │
│                                                                                  │
│   1. OBVIOUS VALUE                                                               │
│      You immediately see why it helps.                                           │
│      No explanation needed.                                                      │
│                                                                                  │
│   2. NATURAL FIT                                                                 │
│      It feels like it belongs.                                                   │
│      Like it should have always been there.                                      │
│                                                                                  │
│   3. SIMPLE PATH                                                                 │
│      Implementation isn't convoluted.                                            │
│      Straightforward to add.                                                     │
│                                                                                  │
│   4. CLEAR RESULT                                                                │
│      You can tell when it's working.                                             │
│      Success is obvious.                                                         │
│                                                                                  │
│   5. USER-CENTRIC                                                                │
│      Serves actual needs.                                                        │
│      Makes the user's life better.                                               │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Enhancement Quality Hierarchy

```
                         ┌───────────────────────────────┐
                         │                               │
                         │      ★ INTUITIVE ★            │
                         │                               │
                         │  Feels obvious in retrospect  │
                         │  "Of course, that's perfect"  │
                         │                               │
                         ├───────────────────────────────┤
                         │                               │
                         │        LOGICAL                │
                         │                               │
                         │  Makes sense when explained   │
                         │  "Ah, I see why"              │
                         │                               │
                         ├───────────────────────────────┤
                         │                               │
                         │        TECHNICAL              │
                         │                               │
                         │  Requires expertise           │
                         │  "Trust me, it's better"      │
                         │                               │
                         └───────────────────────────────┘

                         PREFER: intuitive > logical > technical
```

### The Enhancement Question

For every proposed enhancement, ask:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                  │
│   "Does this directly produce the desired outcome                                │
│    in a way that feels natural?"                                                 │
│                                                                                  │
│   ┌──────────────────────────────────────────────────────────────────────────┐  │
│   │                                                                           │  │
│   │   YES  →  Propose it                                                      │  │
│   │                                                                           │  │
│   │   NO   →  Find a better path                                              │  │
│   │                                                                           │  │
│   └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Integration with Opus Ecosystem

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           OPUS ECOSYSTEM                                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│   /opus         Generate documentation suite (define what should exist)          │
│       │                                                                          │
│       ▼                                                                          │
│   /fabricate    Create implementation (build what's defined)                     │
│       │                                                                          │
│       ├──────────────────────────────────────────────────────┐                   │
│       │                                                      │                   │
│       ▼                                                      ▼                   │
│   /enhance      Improve existing system              /reopus   Evolve docs       │
│       │         (make it better)                         │     (update definition)│
│       │                                                  │                       │
│       │                                                  ▼                       │
│       │                                          /refabricate                    │
│       │                                              │                           │
│       │                                              │  Regenerate from          │
│       │                                              │  evolved docs             │
│       │                                              │                           │
│       └──────────────────────────────────────────────┘                           │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### When to Use Enhance vs Refabricate

```
┌───────────────────────────────────────────┬──────────────────────────────────────┐
│ Use /enhance when...                      │ Use /refabricate when...             │
├───────────────────────────────────────────┼──────────────────────────────────────┤
│ "I want this to be faster"                │ "The spec changed, update the code"  │
│ "Make this easier to use"                 │ "Add this specific feature from opus"│
│ "Clean up this code"                      │ "Rebuild this module per new spec"   │
│ "Add a feature that makes sense"          │ "Sync implementation with docs"      │
│ "What could be improved here?"            │ "Apply documentation changes"        │
└───────────────────────────────────────────┴──────────────────────────────────────┘
```

---

## Performance Targets

```
┌─────────────────────────────┬─────────────┬───────────────────────────────────────┐
│ Phase                       │ Target      │ Task Instance Count                   │
├─────────────────────────────┼─────────────┼───────────────────────────────────────┤
│ Phase 1: Understanding      │ < 30s       │ 5 parallel Task instances             │
│ Phase 2: Dimension Detection│ < 10s       │ 3 parallel Task instances             │
│ Phase 3: Scouting           │ < 45s       │ 7 parallel Task instances             │
│ Phase 4: Proposal Synthesis │ < 20s       │ 5 parallel Task instances             │
│ Phase 5: User Presentation  │ Interactive │ -                                     │
│ Phase 6: Implementation     │ < 3min      │ N parallel Task instances             │
│ Phase 7: Verification       │ < 30s       │ 5 parallel Task instances             │
│ Phase 8: Outcome Synthesis  │ < 15s       │ 4 parallel Task instances             │
├─────────────────────────────┼─────────────┼───────────────────────────────────────┤
│ TOTAL                       │ < 5min      │ 26+ Task instances (standard quality) │
├─────────────────────────────┼─────────────┼───────────────────────────────────────┤
│ Outcome Achievement Target  │ > 85%       │ User satisfaction with result         │
└─────────────────────────────┴─────────────┴───────────────────────────────────────┘
```

---

## Implementation Checklist

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        EXECUTION CHECKLIST                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  [ ] Phase 1: Spawn parallel understanding Task instances (5 instances)          │
│      [ ] Task(description: "Map current state", subagent_type: "Explore")        │
│      [ ] Task(description: "Parse desired outcome", subagent_type: "general-purpose")
│      [ ] Task(description: "Analyze enhancement gap", subagent_type: "general-purpose")
│      [ ] Task(description: "Scout constraints", subagent_type: "Explore")        │
│      [ ] Task(description: "Extract context", subagent_type: "Explore")          │
│      [ ] → Synthesize into enhancement_context                                   │
│                                                                                  │
│  [ ] Phase 2: Spawn dimension detection Task instances (3 instances)             │
│      [ ] Task(description: "Classify outcome dimensions", subagent_type: "general-purpose")
│      [ ] Task(description: "Analyze outcome keywords", subagent_type: "general-purpose")
│      [ ] Task(description: "Score dimension priority", subagent_type: "general-purpose")
│      [ ] → Generate dimension_profile                                            │
│                                                                                  │
│  [ ] Phase 3: Spawn parallel scout Task instances (7 instances)                  │
│      [ ] Task(description: "Scout feature enhancements", subagent_type: "Explore")
│      [ ] Task(description: "Scout performance enhancements", subagent_type: "Explore")
│      [ ] Task(description: "Scout UX enhancements", subagent_type: "Explore")    │
│      [ ] Task(description: "Scout security enhancements", subagent_type: "Explore")
│      [ ] Task(description: "Scout architecture enhancements", subagent_type: "Explore")
│      [ ] Task(description: "Scout integration enhancements", subagent_type: "Explore")
│      [ ] Task(description: "Scout simplification enhancements", subagent_type: "Explore")
│      [ ] → Collect enhancement opportunities                                     │
│                                                                                  │
│  [ ] Phase 4: Spawn parallel synthesizer Task instances (5 instances)            │
│      [ ] Task(description: "Synthesize quick wins", subagent_type: "general-purpose")
│      [ ] Task(description: "Synthesize core enhancements", subagent_type: "general-purpose")
│      [ ] Task(description: "Synthesize strategic enhancements", subagent_type: "general-purpose")
│      [ ] Task(description: "Find enhancement synergies", subagent_type: "general-purpose")
│      [ ] Task(description: "Validate proposal intuition", subagent_type: "general-purpose")
│      [ ] → Generate ranked proposals                                             │
│                                                                                  │
│  [ ] Phase 5: Present proposals to user                                          │
│      [ ] Display Quick Wins                                                      │
│      [ ] Display Core Enhancements                                               │
│      [ ] Display Strategic Enhancements                                          │
│      [ ] Display Synergies                                                       │
│      [ ] → Await user selection                                                  │
│                                                                                  │
│  [ ] Phase 6: Spawn parallel implementer Task instances (N instances)            │
│      [ ] Task(description: "Implement {name}", subagent_type: "general-purpose") │
│      [ ] ... for each approved enhancement                                       │
│      [ ] Task(description: "Check implementation coherence", subagent_type: "general-purpose")
│      [ ] → Apply implementations                                                 │
│                                                                                  │
│  [ ] Phase 7: Spawn parallel verifier Task instances (5 instances)               │
│      [ ] Task(description: "Verify integration", subagent_type: "Explore")       │
│      [ ] Task(description: "Verify quality", subagent_type: "Explore")           │
│      [ ] Task(description: "Verify performance", subagent_type: "general-purpose")
│      [ ] Task(description: "Verify no regressions", subagent_type: "general-purpose")
│      [ ] Task(description: "Verify outcome achieved", subagent_type: "general-purpose") ★
│      [ ] → Generate verification report                                          │
│                                                                                  │
│  [ ] Phase 8: Spawn outcome synthesis Task instances (4 instances)               │
│      [ ] Task(description: "Score achievement", subagent_type: "general-purpose")│
│      [ ] Task(description: "Analyze remaining gaps", subagent_type: "general-purpose")
│      [ ] Task(description: "Plan evolution path", subagent_type: "general-purpose")
│      [ ] Task(description: "Generate next steps", subagent_type: "general-purpose")
│      [ ] → Generate final report                                                 │
│                                                                                  │
│  [ ] Present outcome achievement and natural evolution                           │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

The enhance command delivers intuitive system improvement through parallel discovery and implementation, focusing on producing the desired outcome rather than arbitrary thresholds—enhancements that just make sense.
