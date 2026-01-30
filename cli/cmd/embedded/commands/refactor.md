refactor - Intelligent structural transformation with behavioral preservation

Usage: Transform project structure while preserving behavior through systematic decomposition, recomposition, and architectural realignment with full reversibility.

**CRITICAL SAFETY PROTOCOL**: This command operates in PREVIEW MODE by default. No changes are applied without explicit user approval at each phase gate.

---

## Sequential Refactoring Protocol

### Phase 1: Structural Analysis & Complexity Assessment

**1.1 Codebase Reconnaissance**
- Parse existing project structure and generate dependency graph
- Identify architectural patterns currently in use (or misuse)
- Map behavioral boundaries and interface contracts
- Create comprehensive structure snapshot for rollback

**1.2 Complexity & Risk Assessment**
Evaluate the following dimensions to determine refactoring aggressiveness:

| Dimension | Low Risk | Medium Risk | High Risk |
|-----------|----------|-------------|-----------|
| **Test Coverage** | >80% | 40-80% | <40% |
| **Dependency Coupling** | Loose | Moderate | Tight |
| **File Size** | <300 LOC avg | 300-1000 LOC | >1000 LOC |
| **Cyclomatic Complexity** | <10 avg | 10-20 | >20 |
| **External Integrations** | None/Few | Some | Many critical |
| **Documentation** | Comprehensive | Partial | Sparse |

**1.3 Aggressiveness Calibration**
Based on assessment, automatically calibrate approach:

- **Conservative** (High Risk Score): Single-file changes, extensive validation between each, preserve all existing patterns
- **Moderate** (Medium Risk Score): Small batches of related changes, validation per batch
- **Aggressive** (Low Risk Score): Larger structural changes, batch validation

> **GATE 1**: Present complexity assessment and recommended aggressiveness level. Ask user to confirm or override before proceeding.

---

### Phase 2: Target Architecture Design

**2.1 Multi-Perspective Synthesis**
- **Architect View**: Optimal module boundaries and clean interfaces
- **Pragmatist View**: Minimal change paths to achieve goals
- **Future Developer View**: Discoverability and navigability
- **Risk-Aware View**: Changes ordered by safety, not convenience

**2.2 Migration Roadmap Generation**
- Define target directory hierarchy and module organization
- Generate dependency-aware transformation ordering
- Identify high-risk transformations requiring extra validation
- Create rollback checkpoints for each transformation batch

**2.3 Change Impact Analysis**
For each proposed transformation, document:
- Files affected (direct and indirect)
- Import/export changes required
- Potential behavioral impacts
- Rollback complexity

> **GATE 2**: Present complete refactoring plan with all proposed changes. User must explicitly approve the plan before any execution begins.

---

### Phase 3: Behavioral Preservation Setup

**3.1 Contract Identification**
- Identify all external interfaces and public contracts
- Document implicit contracts (naming conventions, file locations consumers depend on)
- Flag any contracts that cannot be preserved

**3.2 Validation Infrastructure**
- Create/verify behavioral test suite capturing current functionality
- Establish validation gates for each transformation phase
- Set up continuous verification during transformation
- Define success criteria for each change

> **GATE 3**: Confirm validation infrastructure is ready. If test coverage is insufficient, ask user whether to proceed with higher risk or add tests first.

---

### Phase 4: Proposal Audit & Iterative Execution

**CRITICAL: All changes are PROPOSED first, then AUDITED, then executed only with explicit approval.**

**4.1 Change Proposal Generation**
For each transformation batch, generate detailed proposal:

```
PROPOSED CHANGE #N of M
========================
Type: [Extract|Move|Rename|Merge|Split]
Risk Level: [Low|Medium|High]
Files Affected: [list]

BEFORE:
  [current structure/code snippet]

AFTER:
  [proposed structure/code snippet]

Impact Analysis:
  - Imports to update: [count]
  - Tests affected: [count]
  - External contracts: [preserved|modified|broken]

Rollback: [automatic|manual steps required]
```

**4.2 Proposal Audit Checklist**
Before approving any change, verify:
- [ ] Change aligns with stated refactoring goals
- [ ] Behavioral preservation is achievable
- [ ] Rollback path is clear
- [ ] No unintended side effects identified
- [ ] Naming conventions followed
- [ ] Import paths will resolve correctly

**4.3 Iterative Execution Loop**
```
FOR each transformation_batch:
    1. PRESENT proposed changes with full diff preview
    2. ASK user: "Apply this batch? [Yes/Skip/Modify/Abort]"
    3. IF approved:
        a. Execute transformation
        b. Run validation suite
        c. IF validation fails:
            - Present failure details
            - ASK: "Rollback this change? [Yes/No]"
        d. IF validation passes:
            - Commit checkpoint
            - Report success
    4. IF skipped: Continue to next batch
    5. IF abort: Stop all execution, preserve current state
```

> **GATE 4**: Each batch requires explicit approval. User can approve, skip, modify, or abort at any point.

---

### Phase 5: Verification & Completion

**5.1 Final Validation**
- Execute full behavioral test suite
- Verify all import paths resolve
- Confirm no orphaned files or dead code introduced
- Validate dependency graph improvements

**5.2 Change Report Generation**
```
REFACTORING COMPLETE
====================
Changes Applied: N of M proposed
Changes Skipped: X
Changes Rolled Back: Y

Before/After Metrics:
  - Coupling: [before] → [after]
  - Cohesion: [before] → [after]
  - Avg File Size: [before] → [after]
  - Test Coverage: [before] → [after]

Rollback Archive: [location]
```

> **GATE 5**: Present final report. Confirm user is satisfied or offer full rollback.

---

## Command Options

```bash
# Analyze only - no changes proposed
/refactor ./src --analyze-only

# Preview mode (DEFAULT) - propose changes, require approval for each
/refactor ./src --strategy decompose

# Batch approval - group related changes, approve per batch
/refactor ./src --strategy normalize --batch-size 5

# Set aggressiveness manually (override auto-calibration)
/refactor ./src --aggressiveness conservative

# Audit existing refactor plan
/refactor --audit ./refactor-plan.json
```

## Transformation Strategies

| Strategy | Description | Default Aggressiveness |
|----------|-------------|----------------------|
| `decompose` | Break monolith into modules by domain/layer/feature | Moderate |
| `consolidate` | Merge scattered fragments into cohesive units | Conservative |
| `normalize` | Apply naming conventions consistently | Aggressive |
| `restructure` | Comprehensive architectural realignment | Conservative |
| `extract` | Pull specific functionality into separate module | Moderate |

## Quality Standards

- **Behavioral Preservation**: All tests must pass post-transformation
- **History Preservation**: Git lineage remains intact
- **Full Reversibility**: Complete rollback capability at any point
- **Convention Compliance**: 100% naming consistency
- **Audit Trail**: Every change documented and approved

## Safety Guarantees

1. **No automatic execution** - Every change requires explicit approval
2. **Incremental application** - Changes applied one batch at a time
3. **Continuous validation** - Tests run after each change
4. **Instant rollback** - Any change can be reverted immediately
5. **Complexity-aware** - Aggressiveness calibrated to codebase risk
6. **Full audit trail** - Every proposal documented before execution

---

Target: $ARGUMENTS

The refactor command delivers intelligent structural transformation guided by architectural wisdom, complexity-aware risk management, and mandatory human oversight at every decision point.
