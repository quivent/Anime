---
description: Execute autonomous gap resolution using coordinated multi-agent system
argument-hint: [path] [--max-iterations N] [--agents research,learner,solver,protocol] [--quality-threshold 0.9]
allowed-tools: Task(morchestrator), Bash(morchestrator orchestrate:*), Write, Edit, Read
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to perform autonomous gap resolution with multi-agent coordination and quality validation.

**Target Path:** ${1:-.}
**Resolution Options:** $ARGUMENTS

Execute comprehensive gap resolution including:
- Intelligent gap-to-agent assignment based on capabilities and success rates
- Multi-agent coordination (research, learner, solver, protocol-designer agents)
- Load balancing and task distribution optimization
- Real-time progress monitoring and validation
- Quality assurance with 90% accuracy and 95% rigor enforcement
- Iterative improvement cycles with intelligent stopping conditions

The morchestrator will coordinate specialized agents to resolve detected gaps, validate solutions against quality standards, and provide detailed progress reporting throughout the resolution process.