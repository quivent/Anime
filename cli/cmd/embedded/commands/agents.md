---
description: Manage and coordinate the multi-agent system for development tasks
argument-hint: [action] [agent-type] [--list] [--status] [--assign task-id] [--performance]
allowed-tools: Task(morchestrator), Bash(morchestrator:*), Read
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to manage, coordinate, and monitor the specialized multi-agent system.

**Action:** ${1:-status}
**Agent/Options:** $ARGUMENTS

Available agent management operations:
- **List agents**: Show all registered agents (research, learner, solver, protocol-designer, wikipedia-research)
- **Agent status**: View individual agent performance, success rates, and current tasks
- **Task assignment**: Manually assign specific tasks or gaps to capable agents
- **Load balancing**: Optimize task distribution across agents based on availability
- **Performance analytics**: Analyze agent effectiveness and success rate metrics
- **Coordination protocols**: Monitor agent communication and collaboration patterns

The morchestrator will provide comprehensive agent ecosystem management with real-time monitoring, intelligent task distribution, and performance optimization for maximum development efficiency.