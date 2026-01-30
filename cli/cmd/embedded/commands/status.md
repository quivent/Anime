---
description: Monitor progress and status of development orchestration
argument-hint: [path] [--detailed] [--agents] [--phases] [--format table|json|yaml]
allowed-tools: Task(morchestrator), Bash(morchestrator status:*), Read
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to provide comprehensive progress monitoring and status reporting for ongoing development orchestration.

**Target Path:** ${1:-.}
**Status Options:** $ARGUMENTS

Generate detailed status reports including:
- Phase completion status across all 8 development phases
- Real-time gap count monitoring (active, resolved, failed)
- Agent activity dashboard with task allocation and completion rates
- Iteration progress tracking with effectiveness metrics
- Quality metrics monitoring (accuracy, rigor, completeness)
- Estimated completion time calculation
- Performance optimization insights and recommendations

The morchestrator will analyze current progress, agent performance, and provide actionable insights for improving development efficiency and quality outcomes.