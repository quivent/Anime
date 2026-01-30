declaude - Reality-Based Assessment of Claude Code Output Claims

Usage: Assess Claude Code output for overaggerated claims and unhealthy optimistic assertions, bringing actual accomplishments back to reality through rigorous multi-perspective analysis.

**Rigorous Reality Assessment Protocol:**

🎯 **Phase 1: Context Extraction & Claim Identification**
- Extract Claude Code output from specified context source
- Identify specific claims, assertions, and stated accomplishments
- Categorize claims by type: technical capabilities, performance metrics, completion status, future projections
- Flag potentially exaggerated language patterns and superlative usage
- Create structured claim inventory for systematic analysis

🔍 **Phase 2: Multi-Perspective Calibration Analysis**
- Deploy calibrator-analyst agent for systematic bias detection
- Execute iterative analysis across optimistic→neutral→pessimistic spectrum
- Apply systematic doubt and evidence validation protocols
- Cross-reference claims against demonstrable evidence and concrete outputs
- Identify gaps between aspirational statements and actual deliverables

⚖️ **Phase 3: Evidence Validation & Reality Gap Detection**
- Verify each claim against concrete evidence and measurable outcomes
- Assess feasibility of stated capabilities within given constraints
- Identify instances of wishful thinking, unsubstantiated projections, or inflated descriptions
- Document specific examples of overconfidence bias and unrealistic timeline estimates
- Calculate reality gap percentages between claims and demonstrable results

🔧 **Phase 4: Accuracy Correction & Realistic Reframing**
- Reframe exaggerated claims with evidence-based language
- Provide realistic assessments of actual accomplishments achieved
- Distinguish between completed work, work-in-progress, and aspirational goals
- Replace superlative language with measured, factual descriptions
- Generate accuracy-corrected versions of original claims

📊 **Phase 5: Comprehensive Reality Assessment Report**
- Generate detailed analysis report with claim-by-claim evaluation
- Provide overall accuracy score and reliability assessment
- Include specific examples of corrected language and realistic reframing
- Offer recommendations for improving claim accuracy in future outputs
- Create actionable feedback for reducing optimistic bias patterns

**Parameters:**

- `[context_source]` - Source of Claude Code output to analyze (required)
- `--severity [surface|deep|comprehensive]` - Analysis depth and thoroughness level
- `--format [report|json|summary]` - Output format for analysis results
- `--calibration [conservative|balanced|aggressive]` - Bias detection sensitivity
- `--focus [claims|performance|timelines|capabilities]` - Specific analysis focus area

**Usage Examples:**

```bash
# Basic reality assessment of recent session output
/declaude session_transcript.md

# Comprehensive analysis with detailed reporting
/declaude project_status.md --severity comprehensive --format report

# Focus on performance claims with conservative calibration
/declaude benchmark_results.md --focus performance --calibration conservative

# Quick summary assessment of capability statements
/declaude feature_list.md --severity surface --format summary
```

**Integration Patterns:**

🔗 **Automated Quality Gates**
- Integrate with project review workflows for automated claim validation
- Trigger analysis on documentation updates and status reports
- Generate reality-check alerts for high-confidence assertions requiring evidence

🎯 **Continuous Accuracy Improvement**
- Track accuracy trends and bias patterns over time
- Generate recommendations for improving claim precision
- Establish baselines for realistic assessment language and evidence standards

**Output Formats:**

📋 **Standard Report Format**
```
REALITY ASSESSMENT REPORT
========================

Overall Accuracy Score: X/100
Reality Gap Index: X%

CLAIM ANALYSIS:
- Original: [Exaggerated claim]
- Reality: [Evidence-based correction]
- Gap: [Specific overstatement identified]

PATTERN ANALYSIS:
- Optimistic bias frequency: X instances
- Unsubstantiated claims: X instances
- Evidence gaps: X critical areas

RECOMMENDATIONS:
- [Specific improvements for accuracy]
- [Language adjustments needed]
- [Evidence requirements for future claims]
```

📊 **JSON Output Format**
```json
{
  "accuracy_score": 0-100,
  "reality_gap_index": "percentage",
  "claim_analysis": [
    {
      "original": "claim text",
      "corrected": "reality-based version",
      "gap_type": "overstatement|timeline|capability",
      "evidence_level": "none|weak|moderate|strong"
    }
  ],
  "bias_patterns": {
    "optimistic_frequency": "count",
    "superlative_usage": "count",
    "unsubstantiated_assertions": "count"
  },
  "recommendations": ["improvement suggestions"]
}
```

**Quality Standards:**

✅ **Evidence-Based Analysis** - All assessments grounded in demonstrable facts and measurable outcomes
✅ **Systematic Bias Detection** - Multi-perspective calibration prevents overcorrection in either direction
✅ **Actionable Feedback** - Specific, implementable recommendations for improving claim accuracy
✅ **Measurable Results** - Quantified accuracy scores and reality gap metrics for objective evaluation
✅ **Pattern Recognition** - Identification of recurring bias patterns for systematic improvement

**Tool Integration:**

- **calibrator-analyst**: Core analysis engine for multi-perspective assessment
- **Task**: Agent coordination and workflow management
- **Read/Grep**: Context extraction and claim identification
- **Write**: Report generation and output formatting

**Deployment Status:** ✅ Active
**Location:** `~/.claude/commands/declaude.md`
**Integration:** CLI-ready with SlashCommand support via `/declaude`