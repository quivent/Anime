# Tools Command

Lists all installed Claude Code tools with detailed information including function tools, slash commands, and performance metrics.

**Features:**
🔧 **Function Tools** - Lists all custom tools in ~/.claude/tools/ with metadata
📋 **Slash Commands** - Shows available commands from ~/.claude/commands/
⚡ **Performance Stats** - Tool usage analytics and performance metrics  
📊 **Integration Status** - Validation of tool installation and functionality
🎯 **Quick Access** - Direct links to tool documentation and usage examples

**Usage Examples:**
- List all tools: `/tools`
- Filter by category: `/tools --category analysis`
- Show performance: `/tools --performance`
- Detailed view: `/tools --detailed`
- Export list: `/tools --format json`

**Output Options:**
- Default: Summary table with key information
- Detailed: Complete tool specifications and capabilities
- Performance: Usage statistics and performance metrics
- JSON: Machine-readable format for integration

**Tool Categories:**
- Analysis: Document comparison, code quality, performance monitoring
- Development: Dependency management, testing, refactoring
- Content: Documentation, templates, conversion utilities
- Utilities: File management, system optimization, workflow automation

Target: $ARGUMENTS

Scans and displays comprehensive information about all installed Claude Code tools including function tools, slash commands, performance data, and integration status.