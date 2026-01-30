# refabricate - Massively Parallel Iterative Project Regeneration

Transform updated opus documentation into evolved implementations through multi-dimensional parallel analysis, intelligent delta computation, and coordinated regeneration.

Usage: `/refabricate [project-path]` - Execute parallel refabrication using updated opus documentation to extend, modify, or rebuild existing applications.

**Prerequisites:** An existing project with opus documentation suite. The command analyzes changes and orchestrates massively parallel regeneration.

---

## Parallel Refabrication Architecture

```
                         ┌──────────────────────┐
                         │   /refabricate       │
                         │   <project-path>     │
                         └──────────┬───────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐               ┌────▼────┐               ┌────▼────┐
    │ Current │               │  Opus   │               │ History │
    │ State   │               │  Docs   │               │ Tracker │
    │ Analyzer│               │ Parser  │               │ Agent   │
    └────┬────┘               └────┬────┘               └────┬────┘
         │                          │                          │
         └──────────────────────────┼──────────────────────────┘
                                    ▼
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │ Struct  │ │ API     │ │ Logic   │ │ Config  │ │ Test    │
    │ Delta   │ │ Delta   │ │ Delta   │ │ Delta   │ │ Delta   │
    └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │           │           │
         └───────────┴───────────┼───────────┴───────────┘
                                 ▼
                    ┌───────────────────────────────┐
                    │    Mode Selection Engine      │
                    │  extend|modify|rebuild|section│
                    └───────────────┬───────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │ Module  │ │ Module  │ │ Module  │ │ Module  │ │ Module  │
    │ Regen 1 │ │ Regen 2 │ │ Regen 3 │ │ Regen N │ │ Regen M │
    └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │           │           │
         └───────────┴───────────┼───────────┴───────────┘
                                 ▼
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │ Integ   │ │ Compat  │ │ Preserve│ │ Merge   │ │ Conflict│
    │ Verify  │ │ Check   │ │ Verify  │ │ Handler │ │ Resolver│
    └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │           │           │
         └───────────┴───────────┼───────────┴───────────┘
                                 ▼
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │ Unit    │ │ Integ   │ │ Regress │ │ Spec    │ │ Preserve│
    │ Tester  │ │ Tester  │ │ Tester  │ │Validate │ │ Validate│
    └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │           │           │
         └───────────┴───────────┼───────────┴───────────┘
                                 ▼
                    ┌───────────────────────────────┐
                    │   Master Synthesis Agent      │
                    │   Final Integration & Report  │
                    └───────────────────────────────┘
```

---

## Sequential Protocol with Massive Parallelism

### Phase 1: Parallel State Analysis

**Objective:** Simultaneously analyze current project state, updated opus docs, and change history.

**Parallel Execution Protocol:**

```
# Spawn parallel analyzers for comprehensive state understanding
SPAWN Task(subagent_type: "general-purpose", prompt: current_state_analysis_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: opus_documentation_parse_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: change_history_analysis_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: dependency_graph_analysis_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: custom_code_detection_prompt(...))
```

**Agent Distribution:**
| Agent | Focus | Extraction |
|-------|-------|------------|
| State-Analyzer | Current Implementation | File structure, modules, interfaces, patterns |
| Opus-Parser | Updated Documentation | All 7 opus docs, requirements, specifications |
| History-Tracker | Change Timeline | Previous versions, evolution patterns, decisions |
| Dependency-Mapper | Integration Points | Internal/external deps, coupling analysis |
| Custom-Detector | Preserved Code | Custom implementations, manual modifications |

**Agent Instructions Template:**
```markdown
You are analyzing {focus_area} for project refabrication.

**Project Path:** {project_path}
**Analysis Focus:** {focus_area}

**Extract and Structure:**
1. Current state inventory with metadata
2. Component boundaries and interfaces
3. Integration points and dependencies
4. Quality metrics and patterns
5. Areas requiring preservation

**Output Format (JSON):**
{
  "focus": "{focus_area}",
  "inventory": {...},
  "interfaces": [...],
  "dependencies": [...],
  "preservation_candidates": [...],
  "quality_metrics": {...}
}
```

**Synthesis Checkpoint:**
- Merge all 5 analysis outputs into `project_state` object
- Build unified component map
- Identify preservation requirements
- Establish baseline for delta computation

---

### Phase 2: Parallel Delta Computation

**Objective:** Simultaneously compute changes across all project dimensions.

**Parallel Execution Protocol:**

```
# Spawn parallel delta analyzers for each dimension
SPAWN Task(subagent_type: "general-purpose", prompt: structural_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: api_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: logic_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: config_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: test_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: documentation_delta_prompt(project_state))
SPAWN Task(subagent_type: "general-purpose", prompt: security_delta_prompt(project_state))
```

**Agent Distribution:**
| Agent | Dimension | Delta Scope |
|-------|-----------|-------------|
| Delta-Structure | Architecture | Directories, modules, file organization |
| Delta-API | Interfaces | Endpoints, contracts, schemas |
| Delta-Logic | Business Logic | Algorithms, workflows, rules |
| Delta-Config | Configuration | Settings, environment, feature flags |
| Delta-Test | Testing | Test cases, coverage, validation |
| Delta-Docs | Documentation | README, inline docs, API docs |
| Delta-Security | Security | Auth, permissions, vulnerabilities |

**Agent Instructions Template:**
```markdown
You are computing {dimension} delta for refabrication.

**Current State:** {project_state.current[dimension]}
**Updated Opus:** {project_state.opus[dimension]}

**Compute:**
1. Additions: New elements in opus not in current
2. Modifications: Changed elements requiring updates
3. Deletions: Elements in current not in updated opus
4. Conflicts: Incompatible changes requiring resolution

**Output Format (JSON):**
{
  "dimension": "{dimension}",
  "additions": [...],
  "modifications": [...],
  "deletions": [...],
  "conflicts": [...],
  "impact_score": 0-100,
  "recommended_mode": "extend|modify|rebuild|section"
}
```

**Synthesis Checkpoint:**
- Aggregate all delta outputs into `change_manifest`
- Compute overall impact score
- Generate mode recommendation matrix
- Identify critical path changes

---

### Phase 3: Parallel Mode Strategy Analysis

**Objective:** Simultaneously analyze feasibility and implications for each refabrication mode.

**Parallel Execution Protocol:**

```
# Spawn parallel strategists for each mode
SPAWN Task(subagent_type: "general-purpose", prompt: extend_mode_analysis_prompt(change_manifest))
SPAWN Task(subagent_type: "general-purpose", prompt: modify_mode_analysis_prompt(change_manifest))
SPAWN Task(subagent_type: "general-purpose", prompt: rebuild_mode_analysis_prompt(change_manifest))
SPAWN Task(subagent_type: "general-purpose", prompt: section_mode_analysis_prompt(change_manifest))
```

**Agent Distribution:**
| Agent | Mode | Analysis Focus |
|-------|------|----------------|
| Strategist-Extend | EXTEND | Additive feasibility, isolation requirements |
| Strategist-Modify | MODIFY | Surgical precision, interface stability |
| Strategist-Rebuild | REBUILD | Clean slate benefits, preservation costs |
| Strategist-Section | SECTION | Module boundaries, cascade effects |

**Agent Instructions Template:**
```markdown
You are analyzing {mode} mode feasibility for refabrication.

**Change Manifest:** {change_manifest}
**Preservation Requirements:** {preservation_requirements}

**Analyze:**
1. Mode applicability given detected changes
2. Risk assessment and mitigation strategies
3. Effort estimation and resource requirements
4. Preservation compatibility
5. Rollback complexity

**Output Format (JSON):**
{
  "mode": "{mode}",
  "feasibility_score": 0-100,
  "risk_assessment": {...},
  "effort_estimate": "low|medium|high",
  "preservation_conflicts": [...],
  "execution_plan": [...],
  "recommendation": "recommended|viable|not_recommended"
}
```

**Mode Selection:**
- Present all 4 mode analyses to user
- Highlight recommended mode with rationale
- Allow override with informed consent
- Generate execution plan for selected mode

---

### Phase 4: Parallel Preservation Analysis

**Objective:** Simultaneously analyze and prepare all preservation targets.

**Parallel Execution Protocol:**

```
# Spawn parallel preservation analyzers
SPAWN Task(subagent_type: "general-purpose", prompt: custom_code_preservation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: data_preservation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: config_preservation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: test_preservation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: integration_preservation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: documentation_preservation_prompt(...))
```

**Agent Distribution:**
| Agent | Target | Preservation Scope |
|-------|--------|-------------------|
| Preserve-Custom | Custom Code | Manual implementations, overrides |
| Preserve-Data | Data Assets | Databases, files, state |
| Preserve-Config | Configuration | Env vars, settings, secrets |
| Preserve-Test | Test Assets | Custom tests, fixtures, mocks |
| Preserve-Integration | Integrations | External APIs, third-party |
| Preserve-Docs | Documentation | Custom docs, comments |

**Agent Instructions Template:**
```markdown
You are analyzing {target} preservation for refabrication.

**Project State:** {project_state}
**Selected Mode:** {selected_mode}
**Change Manifest:** {change_manifest}

**Analyze:**
1. Elements requiring preservation
2. Preservation method (copy, reference, merge)
3. Integration points with regenerated code
4. Validation requirements post-refabrication
5. Conflict resolution strategies

**Output Format (JSON):**
{
  "target": "{target}",
  "elements": [
    {
      "path": "file/path",
      "preservation_method": "copy|reference|merge",
      "integration_points": [...],
      "validation_checks": [...]
    }
  ],
  "conflicts": [...],
  "resolution_strategies": [...]
}
```

**Synthesis Checkpoint:**
- Merge all preservation analyses into `preservation_plan`
- Create backup manifest
- Establish restoration checkpoints
- Validate no conflicts between preservation targets

---

### Phase 5: Massive Parallel Regeneration

**Objective:** Simultaneously regenerate all affected components based on selected mode.

**Component Extraction:**
```
# Extract components requiring regeneration from change_manifest
components = extract_regeneration_targets(change_manifest, selected_mode)
# Varies by mode: EXTEND (new only), MODIFY (changed), REBUILD (all), SECTION (targeted)
```

**Parallel Execution Protocol:**

```
# Spawn regeneration agent for EACH component simultaneously
FOR EACH component IN components:
    SPAWN Task(
        subagent_type: "general-purpose",
        prompt: component_regeneration_prompt(component, preservation_plan),
        run_in_background: false
    )
```

**Agent Distribution (Dynamic):**
| Agent | Component | Regeneration Scope |
|-------|-----------|-------------------|
| Regen-Core | core/ | Business logic, domain models |
| Regen-API | api/ | Endpoints, handlers, routing |
| Regen-Models | models/ | Data models, schemas |
| Regen-Services | services/ | Business services |
| Regen-Auth | auth/ | Authentication, authorization |
| Regen-Config | config/ | Configuration management |
| Regen-N | {component_n}/ | Additional components |

**Regeneration Agent Instructions Template:**
```markdown
You are regenerating the {component.name} component for refabrication.

**Mode:** {selected_mode}
**Component Specification:** {component.spec}
**Updated Opus Requirements:** {opus_requirements}
**Preservation Plan:** {preservation_plan.for_component}

**Your Task:**
1. Generate updated implementation per new opus specs
2. Integrate with preserved elements where specified
3. Maintain interface compatibility with unchanged components
4. Include comprehensive error handling
5. Create/update corresponding tests

**Mode-Specific Behavior:**
- EXTEND: Add new functionality only, no modification to existing
- MODIFY: Surgical updates to existing implementation
- REBUILD: Fresh implementation from opus specs
- SECTION: Complete component replacement

**Output Format (JSON):**
{
  "component": "{component.name}",
  "files": [
    {
      "path": "relative/path",
      "content": "implementation",
      "action": "create|modify|preserve",
      "merge_with": "preserved_file_path" // if applicable
    }
  ],
  "preserved_integrations": [...],
  "interface_changes": [...],
  "migration_notes": [...]
}
```

**Batch Synthesis - Integration Coherence:**
```
# Group regenerated components for integration verification
batches = partition(regenerated_components, batch_size=4)

FOR EACH batch IN batches:
    SPAWN Task(
        subagent_type: "general-purpose",
        prompt: integration_coherence_prompt(batch, preservation_plan)
    )
```

---

### Phase 6: Parallel Integration Verification

**Objective:** Simultaneously verify integration between preserved and regenerated components.

**Parallel Execution Protocol:**

```
# Spawn integration verification agents
SPAWN Task(subagent_type: "general-purpose", prompt: interface_compatibility_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: dependency_resolution_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: preservation_integrity_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: merge_conflict_resolution_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: migration_path_validation_prompt(...))
```

**Agent Distribution:**
| Agent | Verification | Scope |
|-------|-------------|-------|
| Verify-Interface | Interface Compatibility | APIs, contracts, signatures |
| Verify-Deps | Dependency Resolution | Import/export, circular deps |
| Verify-Preserve | Preservation Integrity | Custom code, data, config |
| Verify-Merge | Merge Conflicts | Overlapping changes |
| Verify-Migration | Migration Paths | State transitions, data migration |

**Agent Instructions Template:**
```markdown
You are verifying {verification_type} for refabrication.

**Regenerated Components:** {regenerated_components}
**Preserved Elements:** {preserved_elements}
**Original Interfaces:** {original_interfaces}

**Verify:**
1. All integration points maintain compatibility
2. No broken references or missing dependencies
3. Preserved elements integrate correctly with new code
4. Merge operations produce valid results
5. Migration paths are complete and reversible

**Output Format (JSON):**
{
  "verification_type": "{verification_type}",
  "status": "pass|fail|warning",
  "issues": [
    {
      "severity": "critical|warning|info",
      "location": "component.file:line",
      "issue": "description",
      "resolution": "fix suggestion"
    }
  ],
  "required_fixes": [...],
  "integration_report": {...}
}
```

**Conflict Resolution Protocol:**
- Automatic resolution for non-overlapping changes
- Parallel conflict resolver agents for complex merges
- User escalation for unresolvable conflicts

---

### Phase 7: Massive Parallel Validation

**Objective:** Simultaneously execute comprehensive validation across all dimensions.

**Parallel Execution Protocol:**

```
# Spawn validation agents for each quality dimension
SPAWN Task(subagent_type: "general-purpose", prompt: unit_test_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: integration_test_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: regression_test_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: spec_compliance_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: preservation_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: performance_validation_prompt(...))
SPAWN Task(subagent_type: "general-purpose", prompt: security_validation_prompt(...))
```

**Agent Distribution:**
| Agent | Validation | Scope |
|-------|-----------|-------|
| Validate-Unit | Unit Tests | All component tests |
| Validate-Integration | Integration Tests | Cross-component |
| Validate-Regression | Regression Tests | Unchanged behavior |
| Validate-Spec | Specification | Opus compliance |
| Validate-Preserve | Preservation | Custom code integrity |
| Validate-Perf | Performance | No degradation |
| Validate-Security | Security | No new vulnerabilities |

**Agent Instructions Template:**
```markdown
You are validating {validation_dimension} post-refabrication.

**Project State:** {refabricated_state}
**Original State:** {original_state}
**Change Manifest:** {change_manifest}
**Preservation Plan:** {preservation_plan}

**Validate:**
1. Execute {validation_dimension} checks
2. Compare against baseline where applicable
3. Identify regressions or degradations
4. Verify preserved elements remain functional
5. Generate detailed validation report

**Output Format (JSON):**
{
  "dimension": "{validation_dimension}",
  "status": "pass|fail|warning",
  "score": 0-100,
  "baseline_comparison": {
    "before": {...},
    "after": {...},
    "delta": {...}
  },
  "findings": [...],
  "regressions": [...],
  "recommendations": [...]
}
```

---

### Phase 8: Master Synthesis & Deployment

**Objective:** Aggregate all parallel outputs and finalize refabrication.

**Execution Protocol:**

```
SPAWN Task(
    subagent_type: "general-purpose",
    prompt: master_synthesis_prompt(all_phase_outputs)
)
```

**Master Synthesizer Instructions:**
```markdown
You are the master synthesizer completing project refabrication.

**All Phase Outputs:**
- Phase 1 State Analysis: {state_outputs}
- Phase 2 Delta Computation: {delta_outputs}
- Phase 3 Mode Strategy: {strategy_outputs}
- Phase 4 Preservation: {preservation_outputs}
- Phase 5 Regeneration: {regeneration_outputs}
- Phase 6 Integration: {integration_outputs}
- Phase 7 Validation: {validation_outputs}

**Your Objectives:**
1. Apply any remaining integration fixes
2. Finalize all file changes
3. Ensure rollback checkpoint is complete
4. Generate comprehensive refabrication report
5. Provide migration guidance if needed

**Output Format (JSON):**
{
  "final_files": [...],
  "final_adjustments": [...],
  "refabrication_statistics": {
    "mode": "extend|modify|rebuild|section",
    "components_analyzed": N,
    "components_regenerated": N,
    "components_preserved": N,
    "files_created": N,
    "files_modified": N,
    "files_deleted": N,
    "spec_compliance": "%",
    "test_coverage": "%"
  },
  "rollback_available": true,
  "rollback_checkpoint": "path",
  "migration_notes": [...],
  "refabrication_summary": "..."
}
```

---

## Command Invocation

```bash
# Auto-detect mode based on changes
/refabricate ./MyProject

# Explicit mode selection
/refabricate ./MyProject --mode extend
/refabricate ./MyProject --mode modify
/refabricate ./MyProject --mode rebuild
/refabricate ./MyProject --mode section --targets "auth,api"

# With options
/refabricate ./MyProject --mode modify --preserve-custom --dry-run
/refabricate ./MyProject --mode rebuild --preserve-data --preserve-config
/refabricate ./MyProject --parallelism maximum --validation comprehensive
```

## Parameters & Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `project-path` | path | `.` | Target project directory |
| `--mode` | enum | auto | Mode (extend\|modify\|rebuild\|section) |
| `--targets` | list | - | Components for section mode |
| `--preserve-custom` | flag | true | Preserve custom implementations |
| `--preserve-data` | flag | true | Preserve data directories |
| `--preserve-config` | flag | true | Preserve configuration |
| `--preserve-tests` | flag | true | Preserve custom tests |
| `--parallelism` | enum | high | Agent parallelism level |
| `--validation` | enum | standard | Validation depth |
| `--dry-run` | flag | false | Preview without execution |
| `--interactive` | flag | false | Step-by-step approval |
| `--verbose` | flag | false | Detailed output |

## Parallelism Levels

| Level | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 | Phase 6 | Phase 7 |
|-------|---------|---------|---------|---------|---------|---------|---------|
| low | 3 | 3 | 2 | 3 | 4 | 3 | 4 |
| medium | 4 | 5 | 4 | 5 | 8 | 4 | 6 |
| high | 5 | 7 | 4 | 6 | 12 | 5 | 7 |
| maximum | 5 | 7 | 4 | 6 | unlimited | 5 | 7 |

---

## Operational Modes

### ⚡ EXTEND Mode - Additive Enhancement
- Parallel analyzers identify net-new additions only
- Regeneration agents create new components without touching existing
- Integration verification ensures no conflicts with existing code
- Ideal for: Feature additions, new endpoints, capability expansion

### 🔧 MODIFY Mode - Surgical Integration
- Delta analyzers identify precise change boundaries
- Regeneration agents perform targeted updates
- Preservation agents protect unchanged code
- Ideal for: Spec refinements, interface updates, behavior changes

### 🔨 REBUILD Mode - Complete Regeneration
- Full state analysis for comprehensive understanding
- All components regenerated from updated opus
- Preservation agents protect data/config only
- Ideal for: Major rewrites, architecture changes, clean slate

### 📦 SECTION Mode - Targeted Reconstruction
- Section-scoped delta computation
- Targeted component regeneration
- Interface stability verification for boundaries
- Ideal for: Module rewrites, subsystem updates

---

## Quality Assurance

**Per-Phase Validation:**
- ✅ Phase 1: State analysis complete, all sources accessible
- ✅ Phase 2: Delta computed across all dimensions
- ✅ Phase 3: Mode selected with execution plan
- ✅ Phase 4: Preservation plan validated
- ✅ Phase 5: All components regenerated
- ✅ Phase 6: Integration verified, conflicts resolved
- ✅ Phase 7: All validation dimensions pass
- ✅ Phase 8: Master synthesis complete

**Rollback Protocol:**
- Automatic checkpoint before Phase 5
- Incremental rollback capability
- Full restoration with single command
- Change audit trail preserved

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Phase 1 (State) | <30s | 5 parallel analyzers |
| Phase 2 (Delta) | <20s | 7 parallel delta agents |
| Phase 3 (Strategy) | <15s | 4 parallel strategists |
| Phase 4 (Preserve) | <20s | 6 parallel preserve agents |
| Phase 5 (Regen) | <3min | N parallel regenerators |
| Phase 6 (Integrate) | <30s | 5 parallel verifiers |
| Phase 7 (Validate) | <1min | 7 parallel validators |
| Total Refabrication | <6min | Standard complexity |

---

## Integration with Opus Ecosystem

**Complementary Commands:**
- `/opus` → Generates initial documentation suite
- `/fabricate` → Initial project implementation
- `/reopus` → Enhances documentation suite
- `/refabricate` → Updates implementation from evolved docs

**Workflow:**
```
Initial:  /opus → /fabricate → project

Evolution: /reopus "enhancement" → updated docs
           /refabricate → evolved project

Iteration: Repeat as specifications evolve
```

---

## Output & Reporting

```
╔══════════════════════════════════════════════════════════════════╗
║                    REFABRICATION SUMMARY                          ║
╠══════════════════════════════════════════════════════════════════╣
║ Mode: MODIFY                    Parallelism: HIGH                 ║
║ Phases Completed: 8/8           Duration: 4m 32s                  ║
╠══════════════════════════════════════════════════════════════════╣
║                      COMPONENT ANALYSIS                           ║
╠══════════════════════════════════════════════════════════════════╣
║ Components Analyzed:     24     │ Parallel Agents Used:    47     ║
║ Components Regenerated:   8     │ Delta Dimensions:         7     ║
║ Components Preserved:    16     │ Preservation Targets:     6     ║
║ New Components Added:     3     │ Validation Dimensions:    7     ║
╠══════════════════════════════════════════════════════════════════╣
║                      QUALITY METRICS                              ║
╠══════════════════════════════════════════════════════════════════╣
║ Specification Compliance:  97%  │ Test Coverage:           84%    ║
║ Integration Score:         94%  │ Preservation Integrity:  100%   ║
║ Regression Tests:        PASS   │ Security Scan:          PASS    ║
╠══════════════════════════════════════════════════════════════════╣
║ Rollback Available: YES         │ Checkpoint: .refab/checkpoint1  ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## Error Handling

| Error | Phase | Resolution |
|-------|-------|------------|
| `STATE_ANALYSIS_FAILED` | 1 | Check project structure |
| `OPUS_PARSE_FAILED` | 1 | Validate opus documents |
| `DELTA_CONFLICT` | 2 | Review conflicting changes |
| `MODE_INFEASIBLE` | 3 | Select alternative mode |
| `PRESERVATION_CONFLICT` | 4 | Resolve preservation overlaps |
| `REGENERATION_FAILED` | 5 | Retry component agent |
| `INTEGRATION_FAILURE` | 6 | Manual conflict resolution |
| `VALIDATION_FAILURE` | 7 | Review and fix findings |
| `SYNTHESIS_FAILURE` | 8 | Manual intervention |

---

## Implementation Checklist

When executing this command:

- [ ] Spawn Phase 1 parallel state analyzers (5 agents)
- [ ] Synthesize into project_state object
- [ ] Spawn Phase 2 parallel delta agents (7 agents)
- [ ] Generate change_manifest with impact scores
- [ ] Spawn Phase 3 parallel mode strategists (4 agents)
- [ ] Present mode options, confirm selection
- [ ] Spawn Phase 4 parallel preservation agents (6 agents)
- [ ] Generate preservation_plan and backups
- [ ] Spawn Phase 5 parallel regeneration agents (N agents)
- [ ] Execute integration synthesis batches
- [ ] Spawn Phase 6 parallel integration verifiers (5 agents)
- [ ] Resolve all conflicts
- [ ] Spawn Phase 7 parallel validators (7 agents)
- [ ] Execute master synthesis
- [ ] Generate refabrication report
- [ ] Confirm rollback capability

---

The refabricate command delivers massively parallel iterative project evolution, transforming updated opus documentation into evolved implementations with unprecedented speed, comprehensive preservation, and guaranteed rollback capability.
