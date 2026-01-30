Generate optimized command specifications from natural language prompts using linguistic inference and zero-redundancy naming protocols.

Usage: Transform descriptive prompts into structured command protocols with intelligent name generation, comprehensive functionality design, and deployment-ready specifications.

**PROMPT PRIORITIZATION MANDATE**: When a prompt is provided, it takes ABSOLUTE PRIORITY over all local context analysis. The compose command MUST focus exclusively on the user prompt requirements before considering any environmental factors.

**MANDATORY TODOLIST SEQUENTIAL EXECUTION**: All compose operations MUST follow the strict sequential todolist execution protocol as enforced in audit.md, with phase-by-phase validation and completion tracking.

**Sequential Command Composition Protocol:**

🎯 **Phase 1: Prompt Analysis & Intent Extraction** [TODOLIST REQUIRED]
- MANDATORY: Create sequential todolist for all compose phases
- Parse natural language prompt for core operational intent (PROMPT PRIORITY ONLY)
- Extract primary action verbs and domain context from user prompt
- Identify explicit naming conventions when provided in prompt
- Classify command complexity and scope requirements based on prompt specifications

🔤 **Phase 2: Linguistic Inference Engine** [TODOLIST CHECKPOINT]
- Mark Phase 1 as completed in todolist before proceeding
- Generate concise command names prioritizing single words (from prompt only)
- Eliminate linguistic redundancy and verbose terminology
- Apply semantic compression while preserving prompt intent
- Validate pronounceability and memorability factors

📝 **Phase 3: Command Architecture Design** [TODOLIST CHECKPOINT]
- Mark Phase 2 as completed in todolist before proceeding
- Structure comprehensive functionality specifications based on prompt requirements
- Design parameter sets and option configurations aligned with prompt goals
- Define input validation and error handling protocols
- Establish success criteria and performance metrics derived from prompt

🔧 **Phase 4: Implementation Protocol Generation** [TODOLIST CHECKPOINT]
- Mark Phase 3 as completed in todolist before proceeding
- Create detailed execution workflows and process sequences
- Specify tool integrations and dependency requirements for prompt fulfillment
- Generate validation checkpoints and quality gates per audit.md protocol
- Design automation patterns and optimization strategies

📊 **Phase 5: Documentation Framework Creation** [TODOLIST CHECKPOINT]
- Mark Phase 4 as completed in todolist before proceeding
- Generate complete command documentation structure
- Create usage examples and implementation guides reflecting prompt intent
- Specify output formats and reporting mechanisms
- Establish integration patterns and deployment protocols

🚀 **Phase 6: Command Deployment Specification** [TODOLIST CHECKPOINT]
- Mark Phase 5 as completed in todolist before proceeding
- Generate deployment-ready command file structure in `~/.claude/commands/` directory
- Create CLI integration specifications and help systems
- Design namespace management and conflict resolution
- Establish versioning and update protocols
- Ensure directory creation and write permissions validation
- Mark all phases as completed in final todolist update

**Linguistic Inference Rules:**

**Name Generation Priority:**
1. **Explicit Naming** - Use exact name if prompt begins with command name
2. **Single Word Preference** - Prioritize concise single-word commands
3. **Semantic Compression** - Eliminate redundant linguistic elements
4. **Domain Specificity** - Preserve technical precision and clarity
5. **Pronunciation Optimization** - Ensure verbal communication clarity

**Zero Redundancy Protocol:**
- Remove filler words, articles, and unnecessary modifiers
- Compress compound concepts into essential semantic units
- Eliminate verbose descriptions in favor of precise terminology
- Prioritize action-oriented naming over descriptive phrases

**Linguistic Analysis Patterns:**

📋 **Action Extraction**
- Primary verbs: analyze, deploy, optimize, validate, generate
- Secondary actions: monitor, integrate, configure, enhance
- Compound operations: merge, synchronize, coordinate, orchestrate
- Meta-operations: compose, transform, synthesize, architect

🎨 **Domain Classification**
- Development: build, test, debug, deploy, integrate
- Analysis: audit, analyze, validate, assess, measure
- Operations: monitor, manage, orchestrate, coordinate
- Creation: generate, compose, create, synthesize, architect

**Command Architecture Templates:**

🔧 **Simple Operations** (1-3 phases)
- Single-purpose commands with focused functionality
- Minimal parameter sets and straightforward execution
- Direct output formats and basic integration patterns

📊 **Complex Protocols** (4-8 phases)
- Multi-phase sequential execution frameworks
- Comprehensive parameter configurations and validation
- Advanced integration patterns and automation capabilities

🌐 **System Integration** (Variable phases)
- Cross-system coordination and orchestration protocols
- Dynamic phase generation based on system complexity
- Adaptive execution patterns and intelligent routing


**Composition Categories:**

⚡ **Operational Commands**
- System management, monitoring, and maintenance protocols
- Performance optimization and resource management
- Deployment automation and infrastructure coordination

🔍 **Analysis Commands** 
- Data analysis, validation, and quality assessment protocols
- System auditing, security analysis, and compliance checking
- Performance benchmarking and optimization identification

🏗️ **Creation Commands**
- Content generation, system building, and architecture design
- Template creation, documentation generation, and specification building
- Integration pattern development and framework establishment

🔄 **Coordination Commands**
- Multi-system orchestration and workflow management
- Agent coordination, task distribution, and process synchronization
- Cross-platform integration and communication protocols

**Output Specifications:**

📄 **Command File Structure**
```markdown
[Generated Command Name] - [Semantic Description]

Usage: [Concise operational description]

**[Sequential Protocol Framework]:**

🎯 **Phase 1: [Core Operation]**
- [Detailed phase specifications]
- [Validation criteria and success metrics]

[Additional phases as needed]

**[Integration Patterns]:**
- [System integration specifications]
- [Automation protocols and triggers]

**Target**: $ARGUMENTS
```

🔧 **CLI Integration**
- Command registration with help system integration
- Parameter validation and argument parsing protocols
- Error handling and user feedback mechanisms
- Performance monitoring and usage analytics

📊 **Deployment Package**
- Complete command specification document written to `~/.claude/commands/[command_name].md`
- Integration scripts and configuration templates  
- Validation tests and quality assurance protocols
- Documentation and usage examples
- Automatic directory creation if `~/.claude/commands/` doesn't exist

**Quality Standards:**

✅ **Semantic Precision** - Target high intent preservation in name compression
✅ **Linguistic Efficiency** - Minimal redundancy with optimal clarity balance
✅ **Deployment Readiness** - Comprehensive implementation specifications
✅ **Integration Compatibility** - Standardized system integration patterns
✅ **Performance Optimization** - Designed for efficient execution and resource usage

**Naming Algorithm:**

```
IF prompt.startsWith(explicit_name):
    command_name = extract_explicit_name(prompt)
ELSE:
    primary_action = extract_primary_verb(prompt)
    domain_context = identify_domain(prompt)
    command_name = compress_semantic_units(primary_action, domain_context)
    
command_name = apply_zero_redundancy(command_name)
command_name = validate_pronunciation(command_name)
command_name = ensure_uniqueness(command_name)
```

**TODOLIST EXECUTION ENFORCEMENT:**

**MANDATORY PROTOCOL COMPLIANCE:**
1. **TodoWrite REQUIRED**: All compose operations MUST begin with creating a comprehensive todolist covering all 6 phases
2. **Sequential Phase Execution**: Each phase must be marked as "in_progress" before execution and "completed" upon validation
3. **Checkpoint Validation**: No phase progression without completing previous phase todolist validation
4. **Audit.md Compliance**: All validation checkpoints follow the comprehensive accuracy validation framework
5. **Final Verification**: Complete todolist validation before command deployment

**PROMPT ABSOLUTE PRIORITY PROTOCOL:**
- User prompt specifications override ALL local context analysis
- Environmental factors considered ONLY after prompt requirements are fully addressed
- Command generation focuses exclusively on prompt intent before system integration
- Local system compatibility verified AFTER prompt requirements are implemented

**Composition Target**: $ARGUMENTS

**Deployment Protocol**:
1. Generate optimized command name using linguistic inference (PROMPT PRIORITY ONLY)
2. Create comprehensive command specification with MANDATORY sequential todolist execution and validation checkpoints per audit.md protocol
3. Write command file to `~/.claude/commands/[command_name].md`
4. Validate file creation and accessibility
5. Report deployment location and integration status

The compose protocol will analyze the provided prompt with ABSOLUTE PRIORITY over local context, generate an optimized command name using linguistic inference, and create a comprehensive command specification with MANDATORY sequential todolist execution, validation checkpoints per audit.md protocol, and deployment-ready documentation following established command framework standards, automatically deploying to the standardized command directory structure.

