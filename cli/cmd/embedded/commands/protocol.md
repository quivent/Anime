---
description: Route to a protocol from the compendium by name
argument-hint: <protocol-name> [task-description] [--flags]
---

Route to a protocol from the protocols compendium. This is the master entry point.

**Input:** $ARGUMENTS

## Available Protocols
- **AOM** — Allostatic Orbit Maintenance (Friston, Layer 1)
- **ASS** — Adaptive Satisficing Search (Simon, Layer 1)
- **CAAR** — Capacity-Aware Adaptive Reliability (Shannon, Layer 2)
- **COHD** — Cost-Optimal Hierarchical Dispatch (Empirical, Layer 3)
- **CPRD** — Context-Preserving Recursive Delegation (Empirical, Layer 3)
- **DORO** — Depth-Optimal Recursive Orchestration (Empirical, Layer 3)
- **ECD** — Entropic Context Distillation (Shannon, Layer 2)
- **ERO** — Emergent Recursive Orchestration (Empirical, Layer 3)
- **EVIPLEX** — EVIPLEX (Ferrucci, Layer 4)
- **FTRO** — Fault-Tolerant Recursive Orchestration (Empirical, Layer 3)
- **KRA** — K-Line Resonance Architecture (Minsky, Layer 3)
- **NDAA** — Near-Decomposable Administrative Architecture (Simon, Layer 3)
- **NEE** — Negative Expertise Engine (Minsky, Layer 3)
- **OBEP** — Oracle-Bounded Execution Protocol (Turing, Layer 2)
- **PBAC** — Paxos-BFT Agent Consensus (Lamport, Layer 3)
- **PCT** — Perturbative Creation Test (Feynman, Layer 1)
- **PEP** — Priority Executive Protocol (Hamilton, Layer 4)
- **RIP** — Reflective Incompleteness Protocol (Turing, Layer 2)
- **SPA** — Stored-Program Agent (Von Neumann, Layer 2)
- **SPO** — Stationary Phase Orchestration (Feynman, Layer 3)
- **TRIBUNALNET** — TRIBUNALNET (Ferrucci, Layer 3)
- **TSCO** — Tri-Stance Cognitive Orchestration (Dennett, Layer 4)
- **TVM** — Temporal Verification Mesh (Lamport, Layer 3)
- **UCS** — Universal Constructor Swarm (Von Neumann, Layer 3)
- **VNC** — Variational Niche Construction (Friston, Layer 1)
- **five-terrain** — Five-Terrain Doctrine (Sun Tzu, Layer 4)
- **governor** — The Governor (Wiener, Layer 1)
- **morphogenesis** — Morphogenesis (Kay, Layer 2)
- **predictor** — The Predictor (Wiener, Layer 1)
- **slipnet** — Slipnet Dispatch (Hofstadter, Layer 1)
- **tangled-eval** — Tangled Eval (Hofstadter, Layer 1)
- **water-doctrine** — Water Doctrine (Sun Tzu, Layer 3)

### Wealth Protocols
- **w/administrative-architecture** — Administrative Architecture (Priority 3, signal -> execute)
- **w/apollo-pattern** — The Apollo Pattern (Priority 4, approach -> negotiate -> execute)
- **w/bandwidth-arbitrage** — Bandwidth Arbitrage (Priority 2, approach -> negotiate -> execute)
- **w/bottega** — The Bottega (Priority 5, execute -> compound)
- **w/bottega-model** — The Bottega Model (Priority 6, execute -> compound)
- **w/compilation-encryption** — Compilation-as-Encryption (Priority 4, signal -> execute -> compound)
- **w/compilation-monopoly** — Compilation Monopoly (Priority 2, signal -> approach -> negotiate -> execute -> compound)
- **w/convergent** — Convergent Portfolio (Priority 1, ALL)
- **w/crisis-convergence** — Crisis Convergence (Priority 2, signal)
- **w/crypto-identity** — Continuous Crypto Identity (Priority 5, execute -> compound)
- **w/decidable-compiler** — Decidable Compiler (Priority 2, signal -> approach -> execute)
- **w/deterministic-standard** — Deterministic Standard (Priority 3, approach -> execute -> compound)
- **w/energy-arbitrage** — Energy Arbitrage (Priority 3, signal -> approach -> negotiate)
- **w/entropy-fortress** — Entropy Fortress (Priority 3, signal -> approach -> negotiate -> execute -> compound)
- **w/ephemeralization** — Ephemeralization Engine (Priority 5, signal -> execute)
- **w/foundational-primitive** — Foundational Primitive License (Priority 4, negotiate -> execute -> compound)
- **w/incommensurability-arbitrage** — Incommensurability Arbitrage (Priority 3, approach -> negotiate -> execute)
- **w/inference-licensing** — Inference Infrastructure Licensing (Priority 3, approach -> negotiate -> execute)
- **w/instrument** — The Instrument (Priority 4, execute)
- **w/mission-critical** — Mission-Critical Certification (Priority 2, signal -> approach -> execute -> compound)
- **w/modern-medicis** — Modern Medicis (Priority 4, approach -> negotiate)
- **w/notebook-strategy** — The Notebook Strategy (Priority 3, signal)
- **w/safety-arbitrage** — Safety Certification Arbitrage (Priority 3, approach -> negotiate)
- **w/satisficing-architecture** — Satisficing Architecture (Priority 1, signal)
- **w/sep-dual** — SEP + Application Dual Strategy (Priority 5, execute -> compound)
- **w/stackelberg** — Stackelberg Cascade (Priority 3, negotiate -> execute -> compound)
- **w/sword-shield** — The Sword and The Shield (Priority 3, signal -> approach -> negotiate -> execute)
- **w/tensegrity** — Tensegrity Portfolio (Priority 2, signal)
- **w/territorial-monopoly** — Territorial Monopoly (Priority 4, execute -> compound)
- **w/trojan** — The Trojan (Priority 3, signal -> execute -> compound)
- **w/unity-monetization** — Unity Monetization (Priority 4, signal -> negotiate)
- **w/vickrey-auction** — Vickrey Auction (Priority 4, negotiate -> execute)

## Routing

Parse the first argument as the protocol name. Run `proto run <name> <remaining-args>` to execute.
If the protocol name is ambiguous, list the closest matches and ask for clarification.
