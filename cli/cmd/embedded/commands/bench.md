bench - Comprehensive Benchmark Analysis and Comparison

Usage: Produces comprehensive benchmark analysis, documented in markdown, between two provided projects using dynamically determined metrics with tabular comparisons and difference highlighting.

**Sequential Benchmark Protocol Framework:**

🎯 **Phase 1: Project Discovery & Classification**
- Analyze project structure and technology stack identification
- Determine project type classification (web app, CLI, library, mobile, etc.)
- Validate project similarity and comparability assessment
- Extract metadata from configuration files (package.json, Cargo.toml, etc.)
- Establish baseline compatibility for meaningful comparison

🧮 **Phase 2: Dynamic Metrics Determination**
- Execute automated project analysis to identify applicable benchmarks
- Select relevant metric categories based on project characteristics:
  - **Code Quality**: Lines of code, cyclomatic complexity, maintainability index
  - **Performance**: Build times, test execution times, bundle/binary sizes
  - **Architecture**: Dependency counts, coupling metrics, module cohesion
  - **Documentation**: Comment density, README quality, API documentation coverage
  - **Testing**: Test coverage percentages, test-to-code ratios, test complexity
  - **Security**: Vulnerability scans, dependency security scoring, audit results
- Configure metric collection tools specific to detected technologies
- Establish measurement baselines and normalization factors

📊 **Phase 3: Comprehensive Data Collection**
- Execute parallel metric collection across both projects
- Gather quantitative measurements using appropriate toolchains:
  - **Static Analysis**: cloc, sonarqube, language-specific linters
  - **Performance Benchmarks**: hyperfine, criterion, lighthouse, custom timing
  - **Security Audits**: npm audit, cargo audit, dependency vulnerability scanners
  - **Test Coverage**: jest, pytest-cov, tarpaulin, coverage-specific tools
- Collect qualitative assessments for subjective metrics
- Handle measurement errors and provide fallback analysis

🔍 **Phase 4: Comparative Analysis Engine**
- Normalize metrics across different measurement scales and units
- Calculate percentage differences with statistical significance testing
- Identify areas of significant variation (>10% difference thresholds)
- Generate performance indicators and trend analysis
- Create categorical rankings and weighted scoring systems
- Detect outliers and anomalous measurements for validation

📋 **Phase 5: Intelligent Report Generation**
- Generate structured markdown document with executive summary
- Create comprehensive tabular comparisons with visual indicators:
  - 🟢 Superior performance (>10% better)
  - 🟡 Comparable performance (±10% range)
  - 🔴 Inferior performance (>10% worse)
- Include percentage differences with directional indicators
- Add detailed analysis sections with actionable insights
- Generate recommendations based on comparative findings

📈 **Phase 6: Output Optimization & Validation**
- Format report with consistent markdown structure and styling
- Validate all measurements and calculations for accuracy
- Generate executive summary with key performance differentiators
- Include methodology notes and measurement limitations
- Export to specified format (markdown/json/html) with proper encoding

**Command Syntax:**
```bash
bench <project1_path> <project2_path> [options]
```

**Parameters:**
- `project1_path`: Absolute or relative path to first project directory (required)
- `project2_path`: Absolute or relative path to second project directory (required)
- `--output/-o <file>`: Output file path (default: benchmark_analysis.md)
- `--format/-f <type>`: Output format - markdown/json/html (default: markdown)  
- `--metrics/-m <config>`: Custom metrics configuration file path
- `--verbose/-v`: Enable verbose output with detailed analysis steps
- `--exclude/-e <patterns>`: Exclude specific metrics or file patterns (comma-separated)
- `--timeout/-t <seconds>`: Maximum execution time per benchmark (default: 300)
- `--parallel/-p`: Enable parallel metric collection (default: sequential)

**Output Report Structure:**
```markdown
# Benchmark Analysis Report

**Generated**: [ISO timestamp]
**Projects**: [project1_name] vs [project2_name]  
**Analysis Duration**: [execution_time]
**Metrics Analyzed**: [metric_count]

## Executive Summary
[Key findings, recommendations, and overall comparison verdict]

## Project Comparison Overview
### Project A: [name] | [technology_stack] | [size_metrics]
### Project B: [name] | [technology_stack] | [size_metrics]

## Comprehensive Benchmark Results

### Code Quality Metrics
| Metric | Project A | Project B | Difference | Performance |
|--------|-----------|-----------|------------|-------------|
| Lines of Code | [value] | [value] | [±%] | [🟢🟡🔴] |
| Cyclomatic Complexity | [value] | [value] | [±%] | [🟢🟡🔴] |
| Maintainability Index | [value] | [value] | [±%] | [🟢🟡🔴] |

### Performance Benchmarks  
| Metric | Project A | Project B | Difference | Performance |
|--------|-----------|-----------|------------|-------------|
| Build Time (ms) | [value] | [value] | [±%] | [🟢🟡🔴] |
| Bundle Size (KB) | [value] | [value] | [±%] | [🟢🟡🔴] |
| Test Execution (ms) | [value] | [value] | [±%] | [🟢🟡🔴] |

### Architecture Analysis
| Metric | Project A | Project B | Difference | Performance |
|--------|-----------|-----------|------------|-------------|
| Dependencies | [value] | [value] | [±%] | [🟢🟡🔴] |
| Coupling Index | [value] | [value] | [±%] | [🟢🟡🔴] |
| Module Cohesion | [value] | [value] | [±%] | [🟢🟡🔴] |

## Detailed Category Analysis
[Comprehensive analysis for each metric category with insights]

## Actionable Recommendations
[Specific, implementable recommendations based on findings]
```

**Integration Patterns:**
- **CI/CD Integration**: Compatible with GitHub Actions, Jenkins, GitLab CI
- **Automation Triggers**: Git hooks, scheduled analysis, PR-based comparisons
- **Tool Chain Integration**: Seamless integration with existing development toolchains
- **Report Distribution**: Email notifications, Slack integration, dashboard publishing

**Error Handling:**
- Graceful handling of missing dependencies or analysis tools
- Fallback metrics when primary analysis fails
- Clear error reporting with suggested remediation steps
- Partial analysis completion with documented limitations

**Performance Optimization:**
- Parallel execution of independent metric collection
- Incremental analysis for large projects
- Caching of expensive computations between runs
- Resource-aware execution with configurable limits

**Quality Assurance:**
- Input validation for project paths and configuration
- Measurement accuracy verification and statistical validation
- Cross-platform compatibility (macOS, Linux, Windows)
- Comprehensive test coverage for metric calculations

**Usage Examples:**

**Basic Comparison:**
```bash
bench ./project-a ./project-b
```

**Advanced Analysis with Custom Output:**
```bash
bench ~/repos/app-v1 ~/repos/app-v2 --output detailed_analysis.md --verbose --format markdown
```

**Selective Metrics with Exclusions:**
```bash
bench ./legacy ./modern --exclude "test_coverage,security" --metrics ./custom_metrics.json
```

**CI/CD Integration:**
```bash
bench $PR_BASE_DIR $PR_HEAD_DIR --format json --output benchmark_results.json --timeout 600
```

This command provides comprehensive project comparison capabilities with intelligent metric selection, detailed tabular analysis, and actionable insights for development teams seeking data-driven project evaluation and optimization guidance.