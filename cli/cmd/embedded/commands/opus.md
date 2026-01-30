opus - Project Genesis Documentation Suite Generator

Usage: Initialize new projects with comprehensive structured documentation based on application descriptions.

**Parallelized Project Initialization Protocol v2.0:**

This protocol uses sequentially-parallelized execution with agent pools, consensus mechanisms, confidence scoring, iterative refinement, and multi-layer validation. Each parallel phase spawns Task tool agents that work independently and synthesize results through dedicated merge agents.

---

## Quality Levels

```
┌─────────────┬─────────────┬─────────────┬─────────────────────────────────────┐
│ Level       │ Agents      │ Phases      │ Features                            │
├─────────────┼─────────────┼─────────────┼─────────────────────────────────────┤
│ standard    │ 18          │ 12 (skip 4) │ Single-pass, no consensus,          │
│             │             │             │ skip adversarial + meta-validation  │
├─────────────┼─────────────┼─────────────┼─────────────────────────────────────┤
│ high        │ 35          │ 15 (all)    │ Full protocol, single-pass,         │
│ (default)   │             │             │ all validation layers               │
├─────────────┼─────────────┼─────────────┼─────────────────────────────────────┤
│ maximum     │ 70+         │ 15 + loops  │ Consensus generation (2-3x agents), │
│             │             │             │ iterative refinement until clean,   │
│             │             │             │ speculative execution enabled       │
└─────────────┴─────────────┴─────────────┴─────────────────────────────────────┘
```

**Default:** `high` - Balances thoroughness with execution time.

---

## Domain-Specific Specialist Pools

Detected automatically or specified via `--domain [type]`:

```
┌─────────────────┬────────────────────────────────────────────────────────────┐
│ Domain          │ Additional Specialists Added to Pool                       │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ web             │ Frontend Architect, API Designer, Auth Specialist,         │
│                 │ Performance Optimizer, SEO/Accessibility Expert            │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ mobile          │ Platform Specialist (iOS/Android), Offline-First Expert,   │
│                 │ Push Notification Designer, App Store Guidelines Reviewer  │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ cli             │ UX Writing Specialist, Shell Compatibility Checker,        │
│                 │ Flag/Argument Designer, Man Page Specialist                │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ data            │ Pipeline Architect, Model Card Writer, Ethics Reviewer,    │
│                 │ Data Governance Specialist, Reproducibility Auditor        │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ api             │ OpenAPI Specialist, Versioning Strategist, Rate Limit      │
│                 │ Designer, SDK Generation Advisor, Webhook Architect        │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ enterprise      │ Compliance Specialist, Audit Trail Designer, Multi-tenant  │
│                 │ Architect, SLA Definer, Disaster Recovery Planner          │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ game            │ Game Loop Documenter, Asset Pipeline Specialist,           │
│                 │ Multiplayer Architect, Platform Cert Reviewer              │
├─────────────────┼────────────────────────────────────────────────────────────┤
│ auto            │ Analyze description → Select most relevant pool            │
│ (default)       │ Can combine multiple pools if hybrid project detected      │
└─────────────────┴────────────────────────────────────────────────────────────┘
```

---

## Execution Architecture Overview

```
Phase 1:  Setup ─────────────────────────────────────────────────── [Sequential]
    │
    ▼
Phase 2:  Domain Detection ──┬── Pattern Matcher ───┐
                             ├── Keyword Analyzer ──┼──→ Domain Profile ─ [Parallel]
                             └── Complexity Scorer ─┘
    │
    ▼
Phase 3:  Analysis Pool ─────┬── Technical Analyst ─────┐
                             ├── Business Analyst ──────┤
                             ├── UX Analyst ────────────┤
                             ├── Risk Analyst ──────────┼──→ Analysis ──── [Parallel]
                             ├── [Domain Specialist 1] ─┤    Context
                             ├── [Domain Specialist 2] ─┤
                             └── [Domain Specialist N] ─┘
    │
    ▼
Phase 4:  Generation Pool ───┬── README Specialist ─────┐
          (with confidence)  ├── PURPOSE Specialist ────┤
                             ├── INTENT Specialist ─────┤
                             ├── CONCEPTS Specialist ───┼──→ 7 Drafts ──── [Parallel]
                             ├── METHODS Specialist ────┤    + Confidence
                             ├── CLAUDE Specialist ─────┤    Scores
                             ├── SPEC Specialist ───────┤
                             └── [Domain Specialists] ──┘
    │
    ├── [maximum quality: duplicate pool for consensus] ──→ Consensus Merger
    │
    ▼
Phase 5:  Alignment Pool ────┬── PURPOSE↔INTENT ────────┐
                             ├── CONCEPTS↔SPEC ─────────┤
                             ├── README↔ALL ────────────┼──→ Alignment ─── [Parallel]
                             ├── METHODS↔SPEC ──────────┤    Report
                             └── Cross-Reference Index ─┘
    │
    ▼
Phase 6:  Deduplication Pool ┬── Content Fingerprinter ─┐
                             ├── Semantic Similarity ───┼──→ Dedup ──────── [Parallel]
                             ├── Reference Consolidator ┘    Recommendations
                             └── Merge Strategist ──────┘
    │
    ▼
Phase 7:  Validation Pool ───┬── Syntax Validator ──────┐
                             ├── Coherence Auditor ─────┤
                             ├── Completeness Scanner ──┼──→ Validation ── [Parallel]
                             ├── Tone Harmonizer ───────┤    Report
                             └── Actionability Assessor ┘
    │
    ▼
Phase 8:  Reading Level Pool ┬── Flesch-Kincaid Scorer ─┐
                             ├── Jargon Density Meter ──┼──→ Readability ─ [Parallel]
                             ├── Audience Matcher ──────┤    Report
                             └── Complexity Balancer ───┘
    │
    ▼
Phase 9:  Evidence Pool ─────┬── Claim Extractor ───────┐
                             ├── Evidence Linker ───────┼──→ Evidence ──── [Parallel]
                             ├── Caveat Flagger ────────┤    Report
                             └── Citation Formatter ────┘
    │
    ▼
Phase 10: Adversarial Pool ──┬── Devil's Advocate ──────┐
                             ├── Naive User Simulator ──┼──→ Adversarial ─ [Parallel]
                             └── Maintainer Perspective ┘    Report
    │
    ▼
Phase 11: Meta-Validation ───┬── Validator Checker ─────┐
          Pool               ├── Coverage Auditor ──────┼──→ Meta ──────── [Parallel]
                             ├── Consistency Verifier ──┤    Report
                             └── Blind Spot Detector ───┘
    │
    ▼
Phase 12: Iterative Loop ────── Critical issues? ───────────────────────── [Conditional]
                             │
                   ┌─── Yes ─┴─ No ───┐
                   ▼                   ▼
           Return to Phase 4     Continue
           (max 3 iterations)
    │
    ▼
Phase 13: Synthesis ─────────┬── Report Aggregator ─────┐
                             ├── Fix Prioritizer ───────┼──→ Unified ───── [Parallel]
                             ├── Conflict Resolver ─────┤    Action Plan
                             └── Change Orchestrator ───┘
    │
    ▼
Phase 14: Resolution & File Creation ───────────────────────────────────── [Sequential]
    │
    ▼
Phase 15: Completion Report ────────────────────────────────────────────── [Sequential]
```

---

## Phase 1: Directory Validation & Setup (Sequential)

🎯 **Pre-flight Checks**
- Validate target directory is empty or obtain user confirmation
- Create project workspace environment
- Initialize shared context store for agent communication
- Establish file creation permissions
- Parse CLI flags and set quality level
- Initialize speculative execution queue (if maximum quality)

**Output:** `workspace_context` object with configuration and permissions

---

## Phase 2: Domain Detection & Pool Selection (Parallel Pool)

Launch 3 detection agents via Task tool, working independently:

**Agent Pool:**

```yaml
Task: Domain Pattern Matcher
Subagent: general-purpose
Prompt: |
  Analyze this project description for domain indicators:
  "{description}"

  Check for patterns indicating: web, mobile, cli, data, api, enterprise, game

  Return JSON:
  {
    "detected_domains": ["domain1", "domain2"],
    "confidence": {"domain1": 0.9, "domain2": 0.7},
    "indicators_found": ["keyword1", "pattern1"]
  }
```

```yaml
Task: Technical Keyword Analyzer
Subagent: general-purpose
Prompt: |
  Extract technical keywords and frameworks from:
  "{description}"

  Categorize by domain affinity.

  Return JSON:
  {
    "keywords": ["react", "postgres", "REST"],
    "framework_signals": {"web": 3, "api": 2, "mobile": 0},
    "technology_stack_hints": []
  }
```

```yaml
Task: Complexity & Scope Scorer
Subagent: general-purpose
Prompt: |
  Assess project complexity from:
  "{description}"

  Return JSON:
  {
    "complexity_score": 7,  // 1-10
    "scope_indicators": ["multi-user", "real-time", "distributed"],
    "recommended_depth": "comprehensive",  // minimal|standard|comprehensive
    "hybrid_project": true,
    "primary_domain": "web",
    "secondary_domains": ["api", "data"]
  }
```

**Synthesis Agent:**

```yaml
Task: Domain Profile Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize domain detection results from 3 analyzers:

  Pattern Matcher: {pattern_result}
  Keyword Analyzer: {keyword_result}
  Complexity Scorer: {complexity_result}

  Resolve conflicts using confidence weighting.
  Select specialist pool configuration.

  Return JSON:
  {
    "final_domain": "web",
    "secondary_domains": ["api"],
    "specialist_pool": ["Frontend Architect", "API Designer", "Auth Specialist"],
    "complexity_level": "comprehensive",
    "pool_justification": "..."
  }
```

**Output:** `domain_profile` with selected specialist pool

---

## Phase 3: Multi-Perspective Analysis (Parallel Pool)

Launch base analysts + domain specialists simultaneously via Task tool:

**Base Agent Pool (always active):**

```yaml
Task: Technical Analyst
Subagent: general-purpose
Prompt: |
  Context: {workspace_context}
  Domain Profile: {domain_profile}
  Description: "{description}"

  Analyze for:
  - Architecture patterns and system design
  - Technology stack implications
  - Integration requirements and dependencies
  - Scalability and performance considerations
  - Infrastructure and deployment needs

  Include confidence score (0-1) for each finding.
  Flag areas needing evidence or verification.

  Return JSON:
  {
    "findings": [...],
    "confidence_scores": {"architecture": 0.85, ...},
    "evidence_needed": ["scalability claim requires benchmarks"],
    "cross_references": ["relates to Risk Analyst: security integration"]
  }
```

```yaml
Task: Business Analyst
Subagent: general-purpose
Prompt: |
  [Similar structure with business-focused analysis]

  Analyze for:
  - Value propositions and market positioning
  - Target audience and user segments
  - Success metrics and KPIs
  - Competitive differentiation
  - Revenue model considerations

  Include confidence scores and evidence flags.
  Note cross-references to other analysts.
```

```yaml
Task: UX Analyst
Subagent: general-purpose
Prompt: |
  [Similar structure with UX-focused analysis]

  Analyze for:
  - User journey mapping
  - Accessibility requirements
  - Information architecture
  - Onboarding and learning curve
  - Feedback mechanisms
```

```yaml
Task: Risk Analyst
Subagent: general-purpose
Prompt: |
  [Similar structure with risk-focused analysis]

  Analyze for:
  - Security vulnerabilities and threats
  - Compliance and regulatory requirements
  - Technical debt risks
  - Failure modes and recovery
  - Data privacy concerns
```

**Domain Specialist Pool (activated based on domain_profile):**

```yaml
# Example: Web domain specialists
Task: Frontend Architect Analyst
Subagent: general-purpose
Prompt: |
  Domain expertise: Frontend architecture, component design, state management

  Analyze "{description}" for frontend-specific considerations:
  - Component architecture patterns
  - State management requirements
  - Build and bundle optimization needs
  - Browser compatibility scope
  - Progressive enhancement opportunities

  Cross-reference with: Technical Analyst (architecture), UX Analyst (interactions)
```

```yaml
Task: API Designer Analyst
Subagent: general-purpose
Prompt: |
  Domain expertise: API design, REST/GraphQL patterns, versioning

  Analyze for API-specific considerations:
  - Endpoint structure and naming
  - Authentication/authorization patterns
  - Rate limiting and quotas
  - Versioning strategy
  - Documentation requirements (OpenAPI)

  Cross-reference with: Technical Analyst (integration), Risk Analyst (security)
```

**Analysis Synthesis Agent:**

```yaml
Task: Analysis Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize all analyst outputs into unified context:

  Technical: {technical_analysis}
  Business: {business_analysis}
  UX: {ux_analysis}
  Risk: {risk_analysis}
  Domain Specialists: {domain_analyses}

  Tasks:
  1. Merge findings, resolving conflicts by confidence score
  2. Build cross-reference index linking related findings
  3. Aggregate evidence requirements
  4. Calculate composite confidence per topic area
  5. Identify gaps where no analyst provided coverage

  Return JSON:
  {
    "unified_analysis": {...},
    "cross_reference_index": {...},
    "evidence_requirements": [...],
    "confidence_matrix": {...},
    "coverage_gaps": [...],
    "conflict_resolutions": [...]
  }
```

**Output:** `analysis_context` with unified findings and cross-references

---

## Phase 4: Specialist Document Generation (Parallel Pool)

Launch 7 document specialists + domain enhancers via Task tool.
Each agent outputs content + confidence scores + evidence flags.

**At maximum quality:** Duplicate each specialist (A/B/C variants), then run Consensus Merger.

**Base Generation Pool:**

```yaml
Task: README Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: Developer onboarding, quick-start optimization
  Analysis Context: {analysis_context}
  Domain Profile: {domain_profile}

  Generate README.md optimized for:
  - Immediate understanding (< 30 seconds scan)
  - Copy-paste-ready commands
  - Clear navigation to detailed docs
  - Badge placement and visual hierarchy
  - Contributing invitation

  For each section, provide:
  - Content
  - Confidence score (0-1)
  - Evidence flags (claims needing support)
  - Cross-references to other docs

  Return JSON:
  {
    "document": "# Project Name\n...",
    "section_confidence": {
      "overview": 0.95,
      "installation": 0.88,
      "quick_start": 0.72  // flagged for review
    },
    "evidence_flags": [
      {"section": "performance", "claim": "handles 10k requests/sec", "needs": "benchmark"}
    ],
    "cross_references": [
      {"section": "architecture", "references": "SPECIFICATION.md#system-design"}
    ]
  }
```

```yaml
Task: PURPOSE Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: Mission articulation, strategic positioning
  Analysis Context: {analysis_context}

  Generate PURPOSE.md with:
  - Compelling problem statement
  - Clear solution articulation
  - Measurable objectives
  - Stakeholder value mapping
  - Vision and roadmap context

  Include confidence scores and evidence requirements per section.
  Note which claims require external validation.
```

```yaml
Task: INTENT Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: User goal modeling, acceptance criteria
  Analysis Context: {analysis_context}

  Generate INTENT.md with:
  - Primary/secondary user personas
  - Goal-oriented use cases
  - Feature requirements (prioritized)
  - Acceptance criteria per use case
  - Edge cases and boundaries

  Include confidence scores. Flag personas needing user research validation.
```

```yaml
Task: CONCEPTS Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: Pedagogical structure, terminology precision
  Analysis Context: {analysis_context}

  Generate CONCEPTS.md with:
  - Progressive complexity introduction
  - Precise terminology definitions
  - Concept relationship mapping
  - Analogy bridges for complex ideas
  - Prerequisites and learning paths

  Flag any terms that may have ambiguous definitions across the industry.
```

```yaml
Task: METHODS Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: Implementation strategy, operational procedures
  Analysis Context: {analysis_context}

  Generate METHODS.md with:
  - Methodology justification
  - Implementation approach
  - Testing strategy and quality gates
  - Deployment patterns
  - Operational runbooks

  Flag recommendations that depend on team size/skill assumptions.
```

```yaml
Task: CLAUDE Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: AI collaboration, agent orchestration
  Analysis Context: {analysis_context}

  Generate CLAUDE.md with:
  - AI assistant role definitions
  - Code review protocols
  - Context preservation strategies
  - Tool/MCP integration patterns
  - Quality automation guidelines

  Ensure compatibility with current Claude Code capabilities.
```

```yaml
Task: SPECIFICATION Specialist
Subagent: general-purpose
Prompt: |
  Domain expertise: Technical precision, API contracts
  Analysis Context: {analysis_context}

  Generate SPECIFICATION.md with:
  - System architecture diagrams (mermaid)
  - API specifications
  - Data models and schemas
  - Security specifications
  - Performance requirements and SLAs

  Every metric must have evidence flag if not derived from requirements.
```

**Domain Enhancement Agents (parallel with base):**

```yaml
Task: Domain Document Enhancer
Subagent: general-purpose
Prompt: |
  Domain: {domain_profile.final_domain}
  Specialist Pool: {domain_profile.specialist_pool}
  Base Documents: {base_documents}

  Review all 7 documents for domain-specific enhancements:
  - Add domain-specific sections where missing
  - Enhance existing sections with domain expertise
  - Add domain-specific warnings and best practices
  - Include domain-standard tooling recommendations

  Return enhancement patches for each document.
```

**Consensus Mechanism (maximum quality only):**

```yaml
Task: Document Consensus Merger
Subagent: general-purpose
Prompt: |
  Merge 3 variants of each document into consensus version:

  README variants: {readme_a}, {readme_b}, {readme_c}
  [... for each document ...]

  Consensus strategy:
  1. Identify common content (keep)
  2. Identify unique valuable additions (merge)
  3. Identify conflicts (resolve by confidence score)
  4. Identify weak sections (flag for review)

  Return:
  {
    "consensus_documents": {...},
    "merge_decisions": [...],
    "confidence_improvements": {...},
    "remaining_conflicts": [...]
  }
```

**Output:** 7 draft documents with confidence metadata and cross-references

---

## Phase 5: Cross-Document Alignment Verification (Parallel Pool)

```yaml
Task: PURPOSE ↔ INTENT Alignment Checker
Subagent: general-purpose
Prompt: |
  Documents: {purpose_md}, {intent_md}

  Verify bidirectional alignment:
  - Every objective in PURPOSE has use cases in INTENT
  - Every use case in INTENT serves an objective in PURPOSE
  - Success metrics align with acceptance criteria
  - No orphaned objectives or disconnected intents

  Return:
  {
    "alignment_score": 0.92,
    "misalignments": [...],
    "orphaned_objectives": [...],
    "disconnected_intents": [...],
    "fix_recommendations": [...]
  }
```

```yaml
Task: CONCEPTS ↔ SPECIFICATION Alignment Checker
Subagent: general-purpose
Prompt: |
  Verify terminology consistency:
  - All terms in SPECIFICATION defined in CONCEPTS
  - No undefined jargon
  - Consistent abstraction levels
  - Architectural concepts match technical specs
```

```yaml
Task: README ↔ ALL Alignment Checker
Subagent: general-purpose
Prompt: |
  Verify README accurately represents all documents:
  - Summary accuracy
  - Navigation link validity
  - Feature highlights match INTENT
  - Quick-start matches METHODS
```

```yaml
Task: METHODS ↔ SPECIFICATION Alignment Checker
Subagent: general-purpose
Prompt: |
  Verify implementation feasibility:
  - Methods support all specified requirements
  - Testing covers specifications
  - Deployment supports architecture
  - No implementation gaps
```

```yaml
Task: Cross-Reference Index Builder
Subagent: general-purpose
Prompt: |
  Build comprehensive cross-reference index:

  All documents: {all_documents}

  Create:
  {
    "term_index": {"term": [{"doc": "CONCEPTS.md", "line": 45}, ...]},
    "concept_graph": {"concept": ["related1", "related2"]},
    "link_map": {"source": "target"},
    "orphaned_references": [...],
    "missing_links": [...]
  }
```

**Alignment Synthesis Agent:**

```yaml
Task: Alignment Report Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize all alignment checker outputs:

  Checkers: {all_alignment_results}
  Cross-Reference Index: {cross_ref_index}

  Produce unified alignment report with:
  - Overall alignment score
  - Critical misalignments (must fix)
  - Warnings (should fix)
  - Suggestions (nice to fix)
  - Auto-fixable issues vs manual review needed
```

**Output:** `alignment_report` with prioritized issues

---

## Phase 6: Semantic Deduplication (Parallel Pool)

```yaml
Task: Content Fingerprinter
Subagent: general-purpose
Prompt: |
  Generate semantic fingerprints for all content blocks:

  Documents: {all_documents}

  For each paragraph/section:
  - Generate semantic hash
  - Extract key concepts
  - Identify information type (definition, example, instruction, etc.)

  Return fingerprint database for similarity matching.
```

```yaml
Task: Semantic Similarity Detector
Subagent: general-purpose
Prompt: |
  Using fingerprints: {fingerprint_db}

  Find semantically similar content across documents:
  - Near-duplicate explanations
  - Redundant examples
  - Repeated definitions
  - Overlapping instructions

  Threshold: 0.75 similarity = flag for review

  Return:
  {
    "duplicates": [
      {"locations": ["CONCEPTS.md:45", "SPEC.md:120"], "similarity": 0.89}
    ],
    "near_duplicates": [...],
    "intentional_repetition": [...]  // cross-doc reinforcement
  }
```

```yaml
Task: Reference Consolidator
Subagent: general-purpose
Prompt: |
  Given duplicate analysis: {similarity_results}

  Recommend consolidation strategy:
  - Which location should be canonical?
  - What cross-references to add?
  - Which duplicates are intentional (keep)?
  - Which can be safely deduplicated?

  Return actionable consolidation plan.
```

```yaml
Task: Merge Strategist
Subagent: general-purpose
Prompt: |
  Create merge execution plan:

  Consolidation recommendations: {consolidation_plan}

  For each deduplication:
  - Determine canonical location
  - Draft replacement cross-reference text
  - Identify ripple effects on other references
  - Ensure no information loss

  Return executable merge plan with rollback capability.
```

**Deduplication Synthesis Agent:**

```yaml
Task: Deduplication Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize deduplication analysis:

  Fingerprints: {fingerprints}
  Similarities: {similarities}
  Consolidation Plan: {consolidation}
  Merge Strategy: {merge_plan}

  Produce:
  {
    "dedup_actions": [...],
    "space_savings_estimate": "~15% content reduction",
    "clarity_improvements": [...],
    "references_to_add": [...],
    "content_to_remove": [...]
  }
```

**Output:** `dedup_report` with consolidation actions

---

## Phase 7: Multi-Dimensional Validation (Parallel Pool)

```yaml
Task: Syntax Validator
Subagent: general-purpose
Prompt: |
  Validate all 7 documents for syntax correctness:

  Documents: {all_documents}

  Check:
  - Markdown syntax (headers, lists, code blocks, tables)
  - Link validity (internal and external format)
  - Code block language tags
  - Consistent heading hierarchy
  - Proper escaping of special characters

  Return:
  {
    "syntax_issues": [
      {"file": "README.md", "line": 45, "issue": "unclosed code block", "fix": "add ```"}
    ],
    "auto_fixable": [...],
    "manual_review": [...]
  }
```

```yaml
Task: Coherence Auditor
Subagent: general-purpose
Prompt: |
  Audit logical coherence across all documents:

  Check:
  - Logical flow within each document
  - Argument consistency
  - No contradictions across documents
  - Appropriate detail levels
  - Clear transitions

  Flag any logical leaps or unsupported conclusions.
```

```yaml
Task: Completeness Scanner
Subagent: general-purpose
Prompt: |
  Scan for completeness against templates:

  Check:
  - All required sections present
  - No placeholder text (TODO, TBD, etc.)
  - Sufficient depth per section
  - Examples where needed
  - Proper conclusions

  Compare against domain-specific requirements from {domain_profile}.
```

```yaml
Task: Tone Harmonizer
Subagent: general-purpose
Prompt: |
  Analyze tone consistency:

  Check:
  - Consistent voice across documents
  - Appropriate formality level
  - No jarring style shifts
  - Consistent person (we/you/they)
  - Professional but approachable

  Provide tone adjustment recommendations with examples.
```

```yaml
Task: Actionability Assessor
Subagent: general-purpose
Prompt: |
  Assess actionability of all guidance:

  Check:
  - Clear next steps provided
  - Commands are copy-paste ready
  - Links are actionable
  - Decision points have clear guidance
  - No vague recommendations

  Flag every instance of "consider", "might", "could" without concrete guidance.
```

**Validation Synthesis Agent:**

```yaml
Task: Validation Report Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize all validation results:

  Results: {all_validation_results}

  Produce:
  {
    "total_issues": 47,
    "by_severity": {"critical": 2, "warning": 15, "suggestion": 30},
    "by_file": {...},
    "by_dimension": {...},
    "quality_scores": {
      "syntax": 0.98,
      "coherence": 0.91,
      "completeness": 0.87,
      "tone": 0.94,
      "actionability": 0.82
    },
    "priority_fix_list": [...]
  }
```

**Output:** `validation_report` with quality scores

---

## Phase 8: Reading Level Analysis (Parallel Pool)

```yaml
Task: Flesch-Kincaid Scorer
Subagent: general-purpose
Prompt: |
  Calculate readability metrics for all documents:

  Documents: {all_documents}

  Compute per document and per section:
  - Flesch-Kincaid Grade Level
  - Flesch Reading Ease
  - Average sentence length
  - Average syllables per word

  Return detailed readability map.
```

```yaml
Task: Jargon Density Analyzer
Subagent: general-purpose
Prompt: |
  Analyze technical jargon density:

  Documents: {all_documents}
  Defined terms: {concepts_glossary}

  For each section:
  - Count jargon terms (defined vs undefined)
  - Calculate jargon density ratio
  - Identify jargon clusters (too many terms too fast)
  - Flag undefined technical terms

  Return jargon heat map.
```

```yaml
Task: Audience Matcher
Subagent: general-purpose
Prompt: |
  Match content complexity to declared audience:

  Documents: {all_documents}
  Declared audience: {intent_personas}

  For each persona:
  - Assess if content matches expected knowledge level
  - Identify sections too advanced
  - Identify sections too basic
  - Recommend adjustments

  Return audience fit analysis.
```

```yaml
Task: Complexity Balancer
Subagent: general-purpose
Prompt: |
  Analyze complexity distribution:

  Readability scores: {readability}
  Jargon density: {jargon}
  Audience fit: {audience}

  Identify:
  - Complexity spikes (sudden difficulty increases)
  - Complexity valleys (unnecessary simplification)
  - Progressive disclosure opportunities
  - Restructuring recommendations

  Return complexity balance recommendations.
```

**Reading Level Synthesis Agent:**

```yaml
Task: Reading Level Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize readability analysis:

  Produce:
  {
    "overall_grade_level": 10.5,
    "by_document": {...},
    "audience_fit_score": 0.85,
    "jargon_issues": [...],
    "complexity_spikes": [...],
    "simplification_opportunities": [...],
    "restructuring_recommendations": [...]
  }
```

**Output:** `readability_report` with audience fit analysis

---

## Phase 9: Evidence Verification (Parallel Pool)

```yaml
Task: Claim Extractor
Subagent: general-purpose
Prompt: |
  Extract all claims from documentation:

  Documents: {all_documents}

  Identify and categorize:
  - Performance claims ("handles 10k req/sec")
  - Capability claims ("supports all major browsers")
  - Comparison claims ("faster than X")
  - Future claims ("will support Y")
  - Assumption claims ("users typically...")

  Return:
  {
    "claims": [
      {
        "text": "handles 10k requests per second",
        "location": "README.md:45",
        "type": "performance",
        "verifiable": true,
        "evidence_required": "benchmark results"
      }
    ]
  }
```

```yaml
Task: Evidence Linker
Subagent: general-purpose
Prompt: |
  For each extracted claim: {claims}

  Attempt to link to evidence:
  - Check if evidence exists in documentation
  - Check if evidence is referenced
  - Check if evidence is external (needs link)
  - Check if claim is self-evident (no evidence needed)

  Return evidence linkage map with gaps.
```

```yaml
Task: Caveat Flagger
Subagent: general-purpose
Prompt: |
  Review claims lacking evidence:

  Unlinked claims: {unlinked_claims}

  For each:
  - Can it be softened with caveats?
  - Should it be removed?
  - Does it need a "verify" note?
  - Is it blocking (must have evidence)?

  Return caveat recommendations.
```

```yaml
Task: Citation Formatter
Subagent: general-purpose
Prompt: |
  Format evidence and citations consistently:

  Evidence links: {evidence_map}

  Generate:
  - Inline citation format
  - Reference section entries
  - "Evidence needed" markers
  - External link formatting

  Return formatted citation additions.
```

**Evidence Synthesis Agent:**

```yaml
Task: Evidence Report Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize evidence verification:

  Produce:
  {
    "total_claims": 45,
    "evidenced": 32,
    "needs_evidence": 8,
    "needs_caveat": 5,
    "claim_map": {...},
    "citations_to_add": [...],
    "caveats_to_add": [...],
    "claims_to_remove": [...]
  }
```

**Output:** `evidence_report` with citation requirements

---

## Phase 10: Adversarial Review (Parallel Pool)

```yaml
Task: Devil's Advocate
Subagent: general-purpose
Prompt: |
  Challenge documentation adversarially:

  Documents: {all_documents}
  Claims: {evidence_report.claims}

  Attack vectors:
  - Unstated assumptions that could fail
  - Logical gaps in arguments
  - Feasibility of proposed approaches
  - Missing failure modes
  - Overconfident claims
  - Edge cases not addressed
  - Security vulnerabilities in recommendations

  Be aggressive. Find weaknesses.

  Return:
  {
    "challenges": [
      {
        "target": "METHODS.md:scaling-strategy",
        "attack": "Assumes horizontal scaling without addressing state management",
        "severity": "high",
        "strengthening": "Add state management strategy section"
      }
    ]
  }
```

```yaml
Task: Naive User Simulator
Subagent: general-purpose
Prompt: |
  Simulate a newcomer attempting to use documentation:

  Documents: {all_documents}

  Walkthrough:
  1. Start at README
  2. Attempt to understand project in 60 seconds
  3. Follow quick-start without prior knowledge
  4. Try to find specific information
  5. Attempt to contribute

  Document every point of confusion:
  - Jargon before definition
  - Missing prerequisites
  - Unclear instructions
  - Dead ends
  - Assumed knowledge

  Return confusion journey with specific failure points.
```

```yaml
Task: Maintainer Perspective Auditor
Subagent: general-purpose
Prompt: |
  Evaluate long-term maintainability:

  Documents: {all_documents}

  Assess:
  - Content likely to become stale
  - Version-specific content (dates, version numbers)
  - External dependencies that may change
  - Update burden for common changes
  - Documentation-as-code practices
  - Single points of failure in docs

  Return sustainability assessment with decay risk scores.
```

**Adversarial Synthesis Agent:**

```yaml
Task: Adversarial Report Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize adversarial findings:

  Devil's Advocate: {devils_advocate}
  Naive User: {naive_user}
  Maintainer: {maintainer}

  Produce:
  {
    "critical_weaknesses": [...],
    "user_experience_issues": [...],
    "sustainability_risks": [...],
    "improvement_backlog": [...],
    "documentation_health_score": 0.78
  }
```

**Output:** `adversarial_report` with prioritized weaknesses

---

## Phase 11: Meta-Validation (Parallel Pool)

```yaml
Task: Validator Checker
Subagent: general-purpose
Prompt: |
  Audit the validators' work:

  Validation Report: {validation_report}
  Documents: {all_documents}

  Check:
  - Did validators miss obvious issues?
  - Are validator findings consistent with each other?
  - Did any validator make errors?
  - Are severity ratings appropriate?

  Return meta-validation findings.
```

```yaml
Task: Coverage Auditor
Subagent: general-purpose
Prompt: |
  Audit validation coverage:

  All reports: {all_reports}
  Documents: {all_documents}

  Check:
  - Every section validated by at least 2 dimensions
  - No blind spots in coverage
  - Edge content (footers, badges) checked
  - Code examples validated
  - Links tested

  Return coverage gaps.
```

```yaml
Task: Consistency Verifier
Subagent: general-purpose
Prompt: |
  Verify consistency across all reports:

  All reports: {all_reports}

  Check:
  - No conflicting recommendations
  - Severity ratings consistent
  - Priorities aligned
  - No duplicate issues reported differently

  Return consistency issues.
```

```yaml
Task: Blind Spot Detector
Subagent: general-purpose
Prompt: |
  Detect systematic blind spots:

  All reports: {all_reports}
  Domain Profile: {domain_profile}

  Check for:
  - Domain-specific concerns not addressed
  - Common documentation anti-patterns not checked
  - Accessibility not validated
  - Internationalization readiness not checked
  - License/legal not verified

  Return blind spot analysis.
```

**Meta-Validation Synthesis Agent:**

```yaml
Task: Meta-Validation Synthesizer
Subagent: general-purpose
Prompt: |
  Synthesize meta-validation:

  Produce:
  {
    "validator_accuracy": 0.94,
    "coverage_score": 0.89,
    "consistency_score": 0.96,
    "blind_spots_found": [...],
    "validation_quality": "high",
    "additional_issues": [...]
  }
```

**Output:** `meta_validation_report`

---

## Phase 12: Iterative Refinement Loop (Conditional)

**Trigger Condition:**
```
IF (adversarial_report.critical_weaknesses.length > 0
    OR validation_report.critical_issues.length > 0
    OR meta_validation.additional_issues.severity == "critical")
    AND iteration_count < 3
THEN return to Phase 4 with constraints
ELSE continue to Phase 13
```

**Loop Execution:**

```yaml
Task: Refinement Constraint Builder
Subagent: general-purpose
Prompt: |
  Build constraints for refinement iteration:

  Critical issues: {critical_issues}
  Iteration: {iteration_count}
  Previous attempts: {previous_fixes}

  Create constraint set:
  - Specific sections to regenerate
  - Issues that must be resolved
  - Patterns to avoid
  - Successful patterns to maintain

  Return refinement constraints for Phase 4 re-run.
```

**Loop continues until:**
- No critical issues remain, OR
- 3 iterations completed (diminishing returns threshold)

**Output:** Loop status and iteration metadata

---

## Phase 13: Synthesis & Resolution (Parallel Pool)

```yaml
Task: Report Aggregator
Subagent: general-purpose
Prompt: |
  Aggregate all reports into unified view:

  Reports:
  - Alignment: {alignment_report}
  - Deduplication: {dedup_report}
  - Validation: {validation_report}
  - Readability: {readability_report}
  - Evidence: {evidence_report}
  - Adversarial: {adversarial_report}
  - Meta-validation: {meta_validation_report}

  Create unified issue database with:
  - Deduplication (same issue from multiple reports)
  - Cross-referencing
  - Complete severity mapping
  - Source attribution
```

```yaml
Task: Fix Prioritizer
Subagent: general-purpose
Prompt: |
  Prioritize fixes from aggregated issues:

  Issues: {aggregated_issues}

  Prioritization criteria:
  1. Severity (critical → suggestion)
  2. User impact (blocks usage → nice to have)
  3. Fix complexity (quick fix → major rewrite)
  4. Dependencies (unblocks other fixes)

  Return prioritized fix queue.
```

```yaml
Task: Conflict Resolver
Subagent: general-purpose
Prompt: |
  Resolve conflicting recommendations:

  All recommendations: {all_recommendations}

  For each conflict:
  - Identify conflicting advice
  - Analyze root cause
  - Determine correct resolution
  - Document reasoning

  Return conflict resolutions.
```

```yaml
Task: Change Orchestrator
Subagent: general-purpose
Prompt: |
  Orchestrate change execution plan:

  Prioritized fixes: {fix_queue}
  Conflicts resolved: {resolutions}
  Documents: {all_documents}

  Create execution plan:
  - Ordered change list
  - Dependencies between changes
  - Rollback checkpoints
  - Verification steps per change

  Return executable change plan.
```

**Synthesis Agent:**

```yaml
Task: Final Synthesis
Subagent: general-purpose
Prompt: |
  Create final unified action plan:

  Aggregated: {aggregated}
  Prioritized: {prioritized}
  Resolved: {resolved}
  Orchestrated: {orchestrated}

  Produce:
  {
    "total_changes": 67,
    "execution_order": [...],
    "estimated_improvements": {
      "quality_score": "+15%",
      "readability": "+8%",
      "consistency": "+12%"
    },
    "change_plan": [...]
  }
```

**Output:** `unified_action_plan`

---

## Phase 14: Resolution & File Creation (Sequential)

🔧 **Execute Change Plan**

```
For each change in unified_action_plan.execution_order:
  1. Read affected section
  2. Apply change
  3. Verify change doesn't introduce new issues
  4. Record change for audit trail
  5. Checkpoint if major change
```

**Change Categories Applied:**
- Alignment synchronizations
- Deduplication consolidations
- Syntax corrections
- Coherence improvements
- Completeness additions
- Tone adjustments
- Actionability enhancements
- Readability improvements
- Evidence citations
- Caveat insertions
- Adversarial strengthening
- Cross-reference additions

📁 **File Creation**
- Write all 7 documentation files
- Set appropriate permissions
- Create supporting directories
- Generate cross-reference index file (optional)

**Output:** Complete documentation suite on disk

---

## Phase 15: Completion Report & Quality Summary (Sequential)

📊 **Generate Comprehensive Report**

```
════════════════════════════════════════════════════════════════════════════════
                              OPUS COMPLETION REPORT v2.0
════════════════════════════════════════════════════════════════════════════════

📁 Project: [project-name]
📍 Location: [target-directory]
⏱️  Generated: [timestamp]
🎚️  Quality Level: [standard|high|maximum]
🔄 Iterations: [count]

────────────────────────────────────────────────────────────────────────────────
                              DOMAIN ANALYSIS
────────────────────────────────────────────────────────────────────────────────

Primary Domain:    [web|mobile|cli|data|api|enterprise|game]
Secondary Domains: [list]
Specialist Pool:   [specialists activated]
Complexity Level:  [minimal|standard|comprehensive]

────────────────────────────────────────────────────────────────────────────────
                              AGENT EXECUTION SUMMARY
────────────────────────────────────────────────────────────────────────────────

Total Agents Spawned:    [count]
├── Analysis Pool:       [count] agents
├── Generation Pool:     [count] agents
├── Alignment Pool:      [count] agents
├── Deduplication Pool:  [count] agents
├── Validation Pool:     [count] agents
├── Readability Pool:    [count] agents
├── Evidence Pool:       [count] agents
├── Adversarial Pool:    [count] agents
├── Meta-Validation:     [count] agents
└── Synthesis Pool:      [count] agents

Consensus Merges:        [count] (maximum quality only)
Cross-References Built:  [count]

────────────────────────────────────────────────────────────────────────────────
                              QUALITY METRICS
────────────────────────────────────────────────────────────────────────────────

Analysis Depth:
  ├── Technical:    ████████████████████ [confidence]
  ├── Business:     ████████████████████ [confidence]
  ├── UX:           ████████████████████ [confidence]
  ├── Risk:         ████████████████████ [confidence]
  └── Domain:       ████████████████████ [confidence]

Document Quality:
  ├── README.md:        ████████████████░░░░ [score]% (confidence: [c])
  ├── PURPOSE.md:       ██████████████████░░ [score]% (confidence: [c])
  ├── INTENT.md:        █████████████████░░░ [score]% (confidence: [c])
  ├── CONCEPTS.md:      ██████████████████░░ [score]% (confidence: [c])
  ├── METHODS.md:       ████████████████░░░░ [score]% (confidence: [c])
  ├── CLAUDE.md:        █████████████████░░░ [score]% (confidence: [c])
  └── SPECIFICATION.md: ████████████████░░░░ [score]% (confidence: [c])

Validation Scores:
  ├── Syntax:           ████████████████████ [score]%
  ├── Coherence:        ██████████████████░░ [score]%
  ├── Completeness:     █████████████████░░░ [score]%
  ├── Tone:             ██████████████████░░ [score]%
  └── Actionability:    ████████████████░░░░ [score]%

Alignment Scores:
  ├── PURPOSE ↔ INTENT:   ████████████████████ [score]%
  ├── CONCEPTS ↔ SPEC:    ██████████████████░░ [score]%
  ├── README ↔ ALL:       █████████████████░░░ [score]%
  └── METHODS ↔ SPEC:     ██████████████████░░ [score]%

Readability Analysis:
  ├── Flesch-Kincaid Grade: [grade]
  ├── Jargon Density:       [low|medium|high]
  ├── Audience Fit:         ████████████████░░░░ [score]%
  └── Complexity Balance:   [balanced|spiky|flat]

Evidence Verification:
  ├── Total Claims:         [count]
  ├── Evidenced:            [count] ([percent]%)
  ├── Caveats Added:        [count]
  └── Needs User Verify:    [count]

Adversarial Review:
  ├── Challenges Found:     [count]
  ├── Challenges Resolved:  [count]
  ├── User Confusions Fixed:[count]
  └── Sustainability Score: ████████████████░░░░ [score]%

Meta-Validation:
  ├── Validator Accuracy:   [score]%
  ├── Coverage Score:       [score]%
  └── Blind Spots Found:    [count] (addressed: [count])

────────────────────────────────────────────────────────────────────────────────
                              DEDUPLICATION SUMMARY
────────────────────────────────────────────────────────────────────────────────

Content Analyzed:        [word count] words
Duplicates Found:        [count]
Duplicates Consolidated: [count]
Space Savings:           ~[percent]%
Cross-References Added:  [count]

────────────────────────────────────────────────────────────────────────────────
                              FILES CREATED
────────────────────────────────────────────────────────────────────────────────

  ✅ README.md           [size] - Project overview and setup
  ✅ PURPOSE.md          [size] - Mission and objectives
  ✅ INTENT.md           [size] - Goals and use cases
  ✅ CONCEPTS.md         [size] - Key concepts and terminology
  ✅ METHODS.md          [size] - Implementation approaches
  ✅ CLAUDE.md           [size] - AI collaboration guidelines
  ✅ SPECIFICATION.md    [size] - Technical requirements

Total Documentation:     [total size]

────────────────────────────────────────────────────────────────────────────────
                              CONFIDENCE REPORT
────────────────────────────────────────────────────────────────────────────────

High Confidence Sections (>0.9):
  [list of sections]

Medium Confidence Sections (0.7-0.9):
  [list of sections with recommendations]

Low Confidence Sections (<0.7):
  ⚠️  [section] - Recommend user review: [reason]
  ⚠️  [section] - Needs domain expert verification: [reason]

────────────────────────────────────────────────────────────────────────────────
                              EVIDENCE REQUIREMENTS
────────────────────────────────────────────────────────────────────────────────

Claims Requiring User Verification:
  📋 [claim] in [file:line] - Needs: [evidence type]
  📋 [claim] in [file:line] - Needs: [evidence type]

Caveats Added (review for accuracy):
  ⚡ [location]: "[caveat text]"

────────────────────────────────────────────────────────────────────────────────
                              NEXT STEPS
────────────────────────────────────────────────────────────────────────────────

Immediate:
  1. Review low-confidence sections flagged above
  2. Verify evidence requirements with actual data
  3. Review caveats for appropriate tone

Recommended:
  4. Add project-specific examples and code samples
  5. Initialize version control: git init && git add .
  6. Run /fabricate to generate implementation from docs

Optional:
  7. Run /reopus to enhance existing documentation
  8. Configure CI to validate docs on change

════════════════════════════════════════════════════════════════════════════════
                         OPUS v2.0 - Parallelized Quality
════════════════════════════════════════════════════════════════════════════════
```

---

## Integration Patterns

⚡ **Command Invocation**
```bash
# Standard usage (high quality, auto domain detection)
opus "Create a web application for task management with real-time collaboration"

# Specify quality level
opus --quality standard "Quick prototype for demo"
opus --quality maximum "Production API for financial services"

# Specify domain explicitly
opus --domain mobile "Cross-platform fitness tracking app"
opus --domain "web,api" "Full-stack e-commerce platform"

# Other options
opus --dry-run "Preview analysis without file creation"
opus --verbose "Show all agent outputs"
opus --interactive "Approval gates between phases"
opus --require-evidence "Strict evidence requirements"
opus --skip-adversarial "Skip adversarial review phase"
```

🔧 **CLI Options**
```
--quality [standard|high|maximum]  Quality level (default: high)
--domain [type,type,...]          Domain hint(s) for specialist selection
--dry-run                         Preview without file creation
--verbose                         Display all agent outputs
--interactive                     Approval gates between phases
--require-evidence                Strict mode: block on unverified claims
--skip-adversarial               Skip adversarial review (faster)
--max-iterations [n]             Max refinement iterations (default: 3)
--confidence-threshold [0-1]     Min confidence to pass (default: 0.7)
```

---

## Speculative Execution (Maximum Quality)

At maximum quality level, speculative execution is enabled:

```
Phase N running...
  └── Speculatively start Phase N+1 with predicted outputs
      ├── If Phase N output matches prediction → keep speculative work
      └── If Phase N output differs → discard and restart Phase N+1

Benefits:
- Reduces total execution time by ~20%
- Particularly effective for validation phases
- Automatically disabled if speculation accuracy < 80%
```

---

## Quality Guarantees

| Dimension | Guarantee | Achieved Through |
|-----------|-----------|------------------|
| **Depth** | Multi-perspective analysis | 4+ parallel analyst agents |
| **Domain Expertise** | Specialist optimization | Domain-specific agent pools |
| **Confidence** | Uncertainty quantification | Per-section confidence scores |
| **Consistency** | Cross-document alignment | 5 alignment checker agents |
| **Accuracy** | Evidence verification | Claim extraction + linking |
| **Correctness** | Multi-dimensional validation | 5 validation agents |
| **Readability** | Audience-appropriate | 4 readability agents |
| **Robustness** | Adversarial stress-testing | 3 adversarial reviewers |
| **Meta-Quality** | Validation of validators | 4 meta-validation agents |
| **Refinement** | Iterative improvement | Conditional refinement loops |
| **Synthesis** | Unified resolution | 4 synthesis agents |
| **Deduplication** | No redundancy | 4 deduplication agents |

---

## Artistic Foundation

Named "opus" after musical compositions, representing the creation of a complete, structured work that harmoniously combines multiple elements into a cohesive artistic and functional project foundation.

The parallelized architecture mirrors a symphony orchestra:
- **Analysis Pool** = Strings section (foundational harmony)
- **Generation Pool** = Woodwinds (melodic themes)
- **Validation Pool** = Brass (powerful verification)
- **Adversarial Pool** = Percussion (stress testing the rhythm)
- **Synthesis Pool** = Conductor (unified direction)

Multiple specialist sections play simultaneously under coordinated direction, with iterative rehearsal (refinement loops) until the performance reaches excellence.

---

## Example Output Structure

```
project-directory/
├── README.md           # Project overview and setup
├── PURPOSE.md          # Mission and objectives
├── INTENT.md           # Goals and use cases
├── CONCEPTS.md         # Key concepts and terminology
├── METHODS.md          # Implementation approaches
├── CLAUDE.md           # AI collaboration guidelines
├── SPECIFICATION.md    # Technical requirements
└── .opus/              # (optional) Generation metadata
    ├── confidence.json # Section confidence scores
    ├── evidence.json   # Claim/evidence mapping
    └── cross-ref.json  # Cross-reference index
```

---

## Target Applications

- New project initialization with comprehensive documentation foundation
- High-stakes projects requiring evidence-backed documentation
- Regulated industries needing audit-ready documentation
- Open source projects requiring contributor-friendly docs
- Enterprise projects with multiple stakeholder audiences
- API-first products requiring precise specifications
- ML/AI projects needing ethics and model documentation
