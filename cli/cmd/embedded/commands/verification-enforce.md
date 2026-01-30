Execute mandatory verification enforcement for agent claims to prevent false success assertions and ensure accountability.

Usage: Verify agent claims with concrete evidence before allowing completion assertions.

**Verification Enforcement Protocol:**

🔍 **Phase 1: Claim Verification**
- Agent claim interception and analysis
- Evidence requirement identification and validation
- Automated verification protocol execution
- Accuracy, rigor, and completeness assessment

🎯 **Phase 2: Evidence Collection**
- File existence and integrity verification
- Compilation and build success validation
- Test execution and result analysis
- Functionality demonstration requirements

⚖️ **Phase 3: Accountability Assessment**
- Performance score impact calculation
- Penalty application for false claims
- Trust score adjustment and tracking
- Agent performance history updates

🚫 **Phase 4: False Claim Prevention**
- Automatic blocking of unverified claims
- Graduated penalty enforcement system
- Agent extinction protocol for persistent violations
- Integration with existing performance tracking

**Verification Levels:**
- **Basic**: File existence and syntax validation
- **Standard**: + Compilation and basic functionality tests
- **Comprehensive**: + Full test suite and integration verification

**Evidence Types:**
- `file_exists`: Verify claimed files actually exist
- `compilation_success`: Validate build/compile success
- `test_execution`: Run and verify test results
- `functionality_demo`: Demonstrate working functionality
- `integration_success`: Verify system integration

**Quality Thresholds:**
- **Accuracy Target**: 90% (must be met)
- **Rigor Standard**: 95% (must be met)
- **Implementation Completeness**: 85% (minimum required)
- **Maximum False Claims**: 3 (before major penalties)

**Penalty Structure:**
```
FIRST_FALSE_CLAIM:        -10 points + warning
SECOND_FALSE_CLAIM:       -20 points + performance review
THIRD_FALSE_CLAIM:        -30 points + restricted access
ONGOING_FALSE_CLAIMS:     Progressive penalties → extinction
```

**Integration Commands:**
- `verification-enforce verify AGENT CLAIM_TYPE DESCRIPTION EVIDENCE_ARRAY`
- `verification-enforce status` - Check system status
- `verification-enforce install` - Install git hooks
- `verification-enforce report AGENT` - Generate accountability report

**Automated Hooks:**
- **Pre-commit**: Verify file integrity and syntax
- **Post-commit**: Update accountability metrics
- **Command wrapper**: Intercept completion claims

**Success Metrics:**
- ✅ **Verification Rate**: Percentage of claims successfully verified
- ✅ **Accuracy Score**: Evidence quality and correctness rating
- ✅ **False Claim Reduction**: Decrease in unverified assertions
- ✅ **Trust Score**: Agent reliability and accountability rating
- ✅ **System Integrity**: Prevention of segmentation faults and failures

**Configuration Options:**
- `--agent [BOB|ALICE]`: Target agent for verification
- `--level [basic|standard|comprehensive]`: Verification thoroughness
- `--evidence [type1,type2,...]`: Required evidence types
- `--auto-penalize`: Enable automatic penalty application
- `--install-hooks`: Setup git verification hooks

Target agent: $ARGUMENTS

The verification enforcement protocol ensures agent accountability through mandatory evidence collection and prevents false success claims that lead to system failures.