# plan - Comprehensive Execution Planning and Implementation Orchestration

Generate comprehensive execution planning documents for autonomous implementation with iterative progress tracking, validation protocols, and intelligent agent team coordination.

Usage: Transform implementation requirements into structured execution plans with phase groupings, task breakdown, agent team selection, and context integration for autonomous development coordination.

**Sequential Planning Protocol:**

🎯 **Phase 1: Requirement Analysis & Context Integration**
- Parse natural language prompts and extract core implementation requirements
- Analyze prior context from files, directories, and project metadata  
- Identify technical dependencies, constraints, and success criteria
- Generate requirement completeness assessment with coverage validation

🧠 **Phase 2: Project Classification & Complexity Assessment**
- Classify project type using codebase analysis and requirement patterns
- Assess implementation complexity using multi-dimensional scoring
- Identify critical path dependencies and potential bottlenecks
- Generate timeline estimation with confidence intervals

👥 **Phase 3: Intelligent Agent Team Selection**
- Analyze requirement-to-skill mapping for optimal team composition
- Select agent combinations based on proven collaboration patterns
- Validate team expertise alignment with project technical requirements
- Generate team justification with capability matrix

📋 **Phase 4: Phase Design & Dependency Mapping**
- Create sequential implementation phases with clear boundaries
- Map inter-phase dependencies and validation checkpoints
- Design parallel execution opportunities and resource optimization
- Establish phase-gate criteria and success validation protocols

✅ **Phase 5: Task Breakdown & Granular Planning**
- Generate trackable task items with measurable acceptance criteria
- Create task dependency chains and execution sequence optimization
- Establish validation protocols and quality assurance checkpoints
- Design progress tracking framework with milestone identification

📖 **Phase 6: Context Linking & Reference Integration**
- Create comprehensive context reference mapping for incoming agents
- Generate links to relevant documentation, specifications, and resources
- Establish knowledge transfer protocols and onboarding procedures
- Design context preservation and update synchronization

📝 **Phase 7: Documentation Generation & Plan Compilation**
- Compile comprehensive plan.md with all planning components
- Generate executive summary and implementation overview
- Create detailed phase breakdowns with task specifications
- Establish progress tracking integration and monitoring protocols

🚀 **Phase 8: Deployment Preparation & Autonomous Readiness**
- Validate plan completeness and implementation readiness
- Establish autonomous execution protocols and error handling
- Generate handoff documentation for assigned agent teams
- Create execution monitoring and course correction procedures

**Command Parameters:**

```bash
plan [prompt] [options]

Options:
  --prompt <text>        Primary implementation prompt (required)
  --context <path>       Prior context file or directory analysis
  --agents <team>        Pre-select agent team configuration
                         Options: web-team, cli-team, data-team, security-team, auto
  --phases <number>      Specify phase count (default: auto-detect based on complexity)
  --output <filename>    Custom plan filename (default: plan.md)
  --template <type>      Planning template selection
                         Options: standard, agile, waterfall, research, prototype
  --validation          Enable comprehensive validation protocols
  --timeline <weeks>     Target timeline for implementation completion
  --priority <level>     Implementation priority: low, medium, high, critical
```

**Agent Team Templates:**

🌐 **Web Development Team**
- **Primary**: reactor (React specialist), ui (interface design), webarchitect (system design)
- **Supporting**: performance-optimization-agent, tester, debugger
- **Use Case**: Frontend applications, web services, UI/UX projects

⚡ **CLI Development Team**  
- **Primary**: clia (CLI specialist), developer, basher (shell scripting)
- **Supporting**: tester, documenter, performance-optimization-agent
- **Use Case**: Command-line tools, automation scripts, system utilities

📊 **Data & Analytics Team**
- **Primary**: data-pipeline-orchestrator, business-intelligence-analyst, analyst
- **Supporting**: performance-optimization-agent, database, researcher
- **Use Case**: Data processing, analytics platforms, business intelligence

🔐 **Security & Infrastructure Team**
- **Primary**: cryptographer, infrastructurer, networker, sysadmin
- **Supporting**: performance-optimization-agent, inspector, fixer
- **Use Case**: Security implementations, infrastructure, DevOps projects

🏗️ **Architecture & Systems Team**
- **Primary**: architect, engineer, c-systems-architect, technicalarchitect
- **Supporting**: performance-optimization-agent, tester, inspector
- **Use Case**: System architecture, performance optimization, enterprise systems

**Planning Templates:**

📋 **Standard Template** - Balanced approach with comprehensive coverage
📈 **Agile Template** - Sprint-based planning with iterative development
🏗️ **Waterfall Template** - Sequential phases with formal validation gates
🔬 **Research Template** - Investigation-focused with hypothesis validation
⚡ **Prototype Template** - Rapid development with proof-of-concept focus

**Output Format:**

The generated plan.md includes:

```markdown
# [Project Name] - Implementation Plan

## Executive Summary
- Project overview and objectives
- Implementation timeline and milestones
- Resource requirements and team composition
- Success criteria and validation protocols

## Requirements Analysis
- Functional requirements with acceptance criteria
- Technical constraints and dependencies
- Risk assessment and mitigation strategies
- Context references and supporting documentation

## Agent Team Assignment
- Selected agents with expertise justification
- Team coordination protocols and communication
- Responsibility matrix and accountability framework
- Escalation procedures and conflict resolution

## Implementation Phases
### Phase 1: [Phase Name]
- **Objectives**: Clear phase goals and deliverables
- **Tasks**: Granular task breakdown with owners
- **Dependencies**: Prerequisites and blocking conditions
- **Validation**: Success criteria and quality gates
- **Timeline**: Duration estimates and milestone dates

[Additional phases as generated]

## Progress Tracking Framework
- TodoWrite integration for milestone tracking
- Progress reporting and status updates
- Course correction protocols and adaptation procedures
- Success metrics and performance indicators

## Context References
- [Generated links to relevant documentation]
- [Project specifications and requirements]
- [Technical references and external resources]
- [Team coordination and communication channels]
```

**Integration Patterns:**

🔗 **Context Analysis Pipeline**
```
Input → Read/Glob/Grep → Analysis → Pattern Recognition → Reference Generation
```

🧠 **Agent Selection Algorithm**
```
Requirements → Skill Mapping → Team Templates → Expertise Validation → Team Selection
```

📊 **Progress Synchronization**
```
Plan Generation → TodoWrite Integration → Task Assignment → Progress Monitoring → Plan Updates
```

**Quality Assurance:**

✅ **Planning Completeness** - 90%+ requirement coverage validation
✅ **Agent Expertise Alignment** - Skill-requirement matching verification  
✅ **Phase Dependency Validation** - Critical path analysis and optimization
✅ **Task Granularity Assessment** - Measurable acceptance criteria verification
✅ **Context Integration Accuracy** - Reference linking and accessibility validation

**Usage Examples:**

```bash
# Basic project planning
plan "Implement user authentication system with OAuth2 integration"

# Advanced planning with context analysis
plan --prompt "Add real-time chat feature to existing application" \
     --context ./docs \
     --agents web-team \
     --template agile

# Enterprise project with comprehensive validation
plan "Migrate legacy database to microservices architecture" \
     --output enterprise_migration_plan.md \
     --template waterfall \
     --validation \
     --timeline 12 \
     --priority critical

# Research project planning
plan --prompt "Analyze performance bottlenecks in distributed system" \
     --agents data-team \
     --template research \
     --phases 6

# CLI tool development
plan "Create command-line interface for API management" \
     --agents cli-team \
     --template prototype \
     --output cli_development_plan.md
```

**Success Metrics:**

- 📊 **Plan Completeness**: 95%+ requirement coverage with validation
- 🎯 **Team Alignment**: Optimal agent-requirement matching and expertise validation
- ⚡ **Implementation Readiness**: Autonomous execution capability with minimal intervention
- 📈 **Progress Trackability**: Comprehensive milestone and task completion monitoring
- 🔄 **Adaptability**: Plan modification and course correction capability

**Command Architecture:**

The plan command operates as a **Complex Protocol** with 8 sequential phases, comprehensive parameter configurations, and advanced integration patterns. It provides autonomous implementation orchestration through intelligent agent team coordination and structured execution planning.

**Dependencies:**
- Read, Glob, Grep tools for context analysis
- Write, Edit tools for plan generation  
- Task tool for agent coordination
- TodoWrite for progress tracking integration
- File system access for context reference creation

**Output Location:** Generated plans are written to the current directory as `plan.md` (default) or custom filename specified via `--output` parameter.

---

**Target**: plan.md - Comprehensive execution planning document for autonomous implementation with iterative progress tracking, auditing, validation, phase groupings, trackable task items, and pre-selected agent teams with context linking for incoming agents.

The plan command transforms implementation requirements into deployment-ready execution strategies through intelligent analysis, optimal team selection, and structured autonomous implementation coordination.