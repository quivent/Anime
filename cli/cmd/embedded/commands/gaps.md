---
description: Detect and analyze implementation gaps using advanced pattern recognition
argument-hint: [path] [--type all|todo|logic|test|doc] [--severity critical|high|medium|low] [--format table|json]
allowed-tools: Task(morchestrator), Bash(morchestrator orchestrate:*), Grep, Read
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to perform comprehensive gap detection and analysis on the specified codebase.

**Target Path:** ${1:-.}
**Analysis Options:** $ARGUMENTS

Perform systematic gap detection including:
- TODO pattern scanning (TODO(human), TODO(auto-detect), implementation markers)
- Logic gap analysis (unimplemented functions, incomplete error handling)
- Testing coverage assessment (missing test coverage, validation needs)
- Documentation validation (accuracy, completeness, consistency checks)
- Integration gap detection (missing connections, incomplete integrations)

The morchestrator will classify gaps by severity (Critical, High, Medium, Low) and type (Todo, Logic, Validation, Integration, Test, Doc) with detailed analysis and recommendations for resolution.