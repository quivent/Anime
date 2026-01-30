Load a Collaborative Intelligence agent into the current Claude Code session with full context injection.

Usage: /agent-load <agent-name> [mode]

This command uses the AgentContext tool to inject comprehensive agent intelligence including:
- Core memory files (MEMORY.md, ContinuousLearning.md, etc.)
- BRAIN architecture references and synthesis documents  
- Recent working sessions and collaboration history
- Agent metadata and behavioral patterns

**Agent Name**: $1 (required - name of agent to load, e.g., "Athena", "Topologist", "Manager")
**Mode**: $2 (optional - "switch" for behavioral takeover, "reference" for info-only, default is behavioral integration)

Examples:
- `/agent-load Athena` - Load Athena with behavioral integration
- `/agent-load Topologist switch` - Complete context switch to Topologist  
- `/agent-load Manager reference` - Load Manager as reference only

The tool will inject the agent's complete intelligence and request behavioral integration unless specified otherwise.