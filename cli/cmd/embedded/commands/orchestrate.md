---
description: Execute comprehensive development orchestration using the Morchestrator agent
argument-hint: [path] [--auto-resolve] [--max-iterations N] [--prompt "custom guidance"]
allowed-tools: Task(morchestrator), Bash(morchestrator orchestrate:*), Read, Write, Edit
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to execute the MORCHESTRATED_COMMUNICATION_PROTOCOL on the specified path with autonomous gap detection and resolution.

**Target Path:** ${1:-.}
**Options:** $ARGUMENTS

Execute comprehensive development orchestration including:
- 8-phase development protocol implementation
- Advanced gap detection using pattern recognition
- Multi-agent coordination for resolution
- Quality assurance with 90% accuracy and 95% rigor standards
- Real-time progress monitoring and reporting

The morchestrator will systematically analyze the codebase, detect implementation gaps, coordinate specialized agents for resolution, and execute iterative improvement cycles until completion or max iterations reached.