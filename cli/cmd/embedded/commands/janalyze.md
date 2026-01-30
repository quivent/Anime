Execute the Janitor protocol in planning mode for comprehensive filesystem organization and cleanup analysis with 100% data integrity protection.

Usage: Perform systematic filesystem analysis and generate detailed cleanup plans without executing any destructive operations.

**PROTOCOL ENFORCEMENT: SEQUENTIAL_TODO_REQUIRED**
This command MUST use TodoWrite tool with sequential execution. DO NOT use Task tool parallelization.

**Janitor Protocol Planning Framework:**

🔍 **Phase 1: Filesystem Topology Analysis**
- Analyze directory structure and file distribution patterns
- Map filesystem hierarchy and identify organizational inefficiencies
- Detect duplicate files using content hashing and size analysis
- Assess cross-reference integrity and symbolic link validity

📊 **Phase 2: Intelligent File Classification**
- **Project Files**: Active development directories with build artifacts
- **Archive Files**: Historical data requiring preservation with compression
- **Temporary Files**: Safe-to-remove cache, logs, and build outputs
- **Duplicate Files**: Identical content with different paths or names
- **Orphaned Files**: Broken links, missing dependencies, abandoned resources

🔄 **Phase 3: Hierarchical Clustering Optimization**
- Apply performance-weighted hierarchical clustering algorithms
- Group related files by content similarity and access patterns
- Optimize directory structure for logical organization
- Plan migration paths for improved filesystem efficiency

📋 **Phase 4: Safety Analysis & Risk Assessment**
- Identify high-value files requiring absolute protection
- Analyze potential data loss scenarios and mitigation strategies
- Generate comprehensive backup recommendations
- Create rollback procedures for all proposed changes

🎯 **Phase 5: Cleanup Plan Generation**
- **Safe Removal Plan**: Files confirmed safe for deletion
- **Organization Plan**: Directory restructuring and file migration
- **Deduplication Plan**: Duplicate resolution with primary file selection
- **Archive Plan**: Compression and long-term storage optimization
- **Monitoring Plan**: Automated maintenance and circuit breaker protocols

**Planning Mode Safety Protocols:**
1. **Read-Only Analysis** - No filesystem modifications during planning
2. **Multiple Validation Passes** - Cross-verify all deletion candidates
3. **Backup Requirement Assessment** - Identify critical data protection needs
4. **Circuit Breaker Integration** - Automatic safety halt conditions
5. **Audit Trail Generation** - Complete operation logging and tracking
6. **Rollback Plan Creation** - Detailed recovery procedures for all changes
7. **Risk Categorization** - Low/Medium/High risk operation classification

**Advanced Safety Features:**
- **Content Fingerprinting**: SHA-256 hashing for duplicate detection
- **Reference Tracking**: Symbolic link and dependency mapping
- **Access Pattern Analysis**: Recent usage detection and preservation
- **Version Control Integration**: Git repository awareness and protection
- **Configuration File Detection**: Critical system file identification

**Output Format:**
- 📊 **Filesystem Analysis Report** - Current state and efficiency metrics
- 🗂️ **Organization Plan** - Detailed restructuring recommendations
- 🧹 **Cleanup Plan** - Safe removal candidates with justification
- 📦 **Deduplication Plan** - Duplicate file resolution strategy
- 🔒 **Safety Assessment** - Risk analysis and protection requirements
- 📈 **Efficiency Projections** - Space savings and performance improvements
- ⚡ **Automation Recommendations** - Future maintenance protocols

**Protocol Options:**
- `--analysis-only`: Generate comprehensive filesystem analysis report
- `--conservative`: Apply strictest safety criteria for recommendations
- `--aggressive`: More extensive cleanup with calculated risk assessment
- `--preserve-all`: Maximum data protection with minimal changes
- `--dry-run-preview`: Show exactly what would be executed without planning mode

**Safety Guarantees:**
- ✅ **Zero Data Loss Risk** - Planning mode never modifies files
- ✅ **Multiple Validation Layers** - Cross-verification of all operations
- ✅ **Comprehensive Backup Plans** - Protection strategy for all changes
- ✅ **Circuit Breaker Protection** - Automatic safety halt mechanisms
- ✅ **Audit Trail Completeness** - Full operation logging and tracking

**MANDATORY TODO WORKFLOW:**
When executing this command, Claude MUST:
1. Create TodoWrite with all 5 phases as separate tasks
2. Mark each phase as "in_progress" before starting
3. Complete each phase before moving to the next
4. Mark each phase as "completed" immediately after finishing
5. Never use Task tool for parallel execution

Target filesystem: $ARGUMENTS

The janitize protocol will generate a comprehensive filesystem optimization plan with absolute data integrity protection, providing detailed analysis and recommendations without executing any potentially destructive operations.