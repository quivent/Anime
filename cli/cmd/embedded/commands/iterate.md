iterate - Intelligent iterative refinement with audit-validated completion

Usage: Transform prompts through iterative refinement cycles until completion criteria are met and validated through audit checkpoints.

**Sequential Iterative Refinement Protocol:**

🎯 **Phase 1: Prompt Preparation & Context Analysis**
- Parse and validate input prompt specifications with content sanitization
- Establish baseline metrics and success criteria with threshold configuration
- Initialize iteration tracking system with progress monitoring
- Configure refinement strategy parameters based on prompt complexity analysis

🔄 **Phase 2: Iterative Refinement Engine**
- Execute prompt processing with intelligent feedback loops and adaptive learning
- Apply refinement strategies (incremental/aggressive/adaptive) based on iteration analysis
- Monitor convergence patterns and quality improvements with real-time metrics
- Implement adaptive strategy adjustment based on progress patterns and optimization opportunities

📊 **Phase 3: Completion Assessment**
- Evaluate explicit completion criteria matching with pattern recognition
- Perform implicit pattern recognition for completion signals and quality gates
- Execute comprehensive quality analysis and metric validation with threshold enforcement
- Generate completion confidence scoring with multi-dimensional assessment frameworks

✅ **Phase 4: Audit Validation Gateway**
- Execute /audit command with current iteration state and comprehensive analysis
- Validate audit results against configured thresholds with pass/fail determination
- Implement intelligent retry logic for failed audit scenarios with optimization
- Generate final validation report with actionable insights and improvement recommendations

📈 **Phase 5: Results Compilation & Reporting**
- Compile iteration history and convergence analysis with performance metrics
- Generate comprehensive completion report with quality assessments and insights
- Archive final validated state for future reference and learning integration
- Provide actionable insights and improvement recommendations for optimization

**Integration Patterns:**

🔧 **CLI Integration**
- Standard argument parsing with comprehensive help system integration
- Parameter validation with bounds checking and error handling protocols
- Progress indicators with real-time iteration tracking and visual feedback
- Result caching with intelligent optimization and performance enhancement

📊 **Automation Protocols**
- Automatic task delegation for complex prompts requiring specialized expertise
- Seamless /audit command integration with result validation and caching
- Agent coordination through Task tool for domain-specific refinement requirements
- Performance monitoring with metrics collection and optimization tracking

🎨 **Output Format**
```
Iteration #X/Y - [Strategy: {refinement_strategy}]
Progress: [████████░░] 80% - Convergence detected
Quality Metrics: {current_score}/1.0 (Target: {audit_threshold})
Completion Status: {criteria_analysis}

Final Result: ✅ COMPLETED - Audit Score: 0.92/0.90
Total Iterations: 8/15
Convergence Pattern: Exponential improvement
Time Efficiency: 65% faster than baseline
```

**Usage Examples:**

```bash
# Basic iteration with default parameters
/iterate "Optimize database query performance"

# Advanced iteration with custom criteria and thresholds
/iterate "Design microservices architecture" --max_iterations 15 --audit_threshold 0.90

# Specific completion criteria with explicit validation requirements
/iterate "Implement user authentication" --completion_criteria "security_validated,tests_passing,documentation_complete"

# Adaptive refinement strategy with intelligent optimization
/iterate "Refactor legacy codebase" --refinement_strategy adaptive --max_iterations 20
```

**Parameters:**

- `prompt` (required): Target prompt for iterative refinement and optimization
- `max_iterations` (optional): Safety limit for iteration cycles (default: 10, range: 1-50)
- `completion_criteria` (optional): Explicit completion conditions (comma-separated)
- `audit_threshold` (optional): Minimum audit score for completion (default: 0.85, range: 0.0-1.0)
- `refinement_strategy` (optional): Iteration approach (incremental/aggressive/adaptive, default: adaptive)

**Quality Standards:**

✅ **Completion Detection Accuracy** - Target ≥95% accuracy in completion identification
✅ **Audit Pass Rate** - ≥85% audit success rate on final iteration with quality validation
✅ **Iteration Efficiency** - Optimize for early termination when completion criteria met
✅ **Zero False Positives** - Eliminate false completion scenarios through robust validation
✅ **Performance Optimization** - Maximum efficiency with intelligent convergence detection

**Error Handling:**

🚨 **Graceful Degradation** - Intelligent handling of iteration limit exceeded scenarios
🔄 **Intelligent Retry** - Automatic retry logic for temporary audit failures with optimization
💬 **User Feedback** - Clear guidance on completion criteria ambiguity with actionable suggestions
📝 **Comprehensive Logging** - Detailed logging for debugging, optimization, and continuous improvement

**Deployment Configuration:**

- **Command Registration**: Automatic integration with Claude Code help system
- **Namespace Management**: Conflict resolution with existing command infrastructure
- **Versioning Protocol**: Update management with backward compatibility
- **Performance Monitoring**: Usage analytics and optimization tracking with metrics collection

**Target Completion Criteria:**
- Audit validation pass with score ≥ threshold
- Explicit criteria satisfaction (when specified)
- Implicit completion pattern recognition
- Quality gate validation across all assessment dimensions
- User satisfaction and actionable result delivery