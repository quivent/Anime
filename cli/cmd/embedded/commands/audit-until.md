Iterative audit-fix loop: audit the target, fix all issues found, re-audit, and repeat until the audit returns clean.

Usage: Provide a target (file, document, or codebase area) and this protocol will autonomously cycle through audit and fix phases until zero issues remain.

**Iterative Audit-Fix Protocol:**

## Loop Structure

Repeat the following cycle until the audit phase finds zero issues:

### Phase 1: Audit

Apply the full multi-dimensional validation framework:

🔍 **System Functionality Verification**
- Validate claims against actual system state
- Cross-reference file system, configurations, and code

📊 **Mathematical/Statistical Validation**
- Verify all calculations and numeric claims
- Check formula accuracy and metric derivations

🌐 **External Fact Verification**
- Cross-check factual claims against reliable sources
- Validate technical specifications and standards

📋 **Internal Consistency Audit**
- Identify contradictions and broken references
- Verify terminology consistency and logical flow

🎯 **Evidence-Based Assessment**
- Check that all claims are supported with evidence
- Identify unsupported assertions

Produce a numbered list of issues found, each classified as:
- ❌ **Invalid** - Demonstrably incorrect, must be fixed
- ⚠️ **Questionable** - Likely wrong or unsupported, should be fixed
- 💡 **Minor** - Cosmetic or style issue, fix if straightforward

### Phase 2: Fix

For each issue found in the audit:
1. Identify the exact location (file, line, section)
2. Determine the correct value, statement, or structure
3. Apply the fix directly
4. Record what was changed and why

### Phase 3: Re-Audit

Run the full audit again on the now-modified target.
- If new issues are found, return to Phase 2
- If zero issues are found, proceed to Phase 4

### Phase 4: Completion

When the audit returns clean:
1. Summarize all changes made across all iterations
2. Report the number of audit-fix cycles required
3. Confirm the target is now fully validated
4. Declare: "Audit clean. All issues resolved."

**Constraints:**
- Maximum 10 audit-fix cycles to prevent infinite loops
- If the same issue persists after 3 fix attempts, flag it as requiring human intervention
- Each cycle must reduce the total issue count or halt with explanation
- Never introduce new issues while fixing existing ones

**Stop Condition:**
The loop terminates when:
- The audit phase finds **zero** issues (success), OR
- The maximum cycle count (10) is reached (partial success), OR
- An issue cannot be resolved after 3 attempts (escalation needed)

Audit-until target: $ARGUMENTS
