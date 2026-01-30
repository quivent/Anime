# Topologist - Repository Topology & Initialization Protocol

Usage: Systematically analyze current directory topology, initialize git repository if needed, create private GitHub remote with SSH configuration, and generate intelligent staging plan with batched commits.

**Repository Topology Analysis & GitHub Integration Protocol:**

🎯 **Phase 1: Repository Topology Detection**
- Analyze `.git` directory existence and integrity validation
- Scan directory structure for file type classification and project detection
- Evaluate existing remote configurations and repository health status
- **Validation Criteria**: Repository status accurately identified with comprehensive topology mapping

🔧 **Phase 2: Conditional Git Initialization** 
- Initialize git repository with modern branch naming conventions
- Configure user settings and repository-specific git configuration
- Establish initial commit structure and branch protection protocols
- **Validation Criteria**: Git repository properly initialized with secure configuration standards

🌐 **Phase 3: GitHub Remote Integration**
- Extract current directory name for repository naming consistency
- Parse command arguments for organization specification (format: "on ORG/REPO" or "--org ORG")
- Determine target organization (command-specified, parameter-specified, or gh CLI default account)
- Create private GitHub repository with proper permissions and settings
- Configure SSH remote origin with authentication key management
- **Validation Criteria**: Remote repository created and SSH origin configured successfully

📝 **Phase 4: Essential File Generation**
- Generate context-aware .gitignore based on detected file types and frameworks
- Create comprehensive README.md with project structure documentation (conditional)
- Validate file permissions, encoding, and accessibility across platforms
- **Validation Criteria**: Essential files created with appropriate content and formatting

📊 **Phase 5: Intelligent Staging Strategy**
- Categorize all files by type, modification status, and logical grouping
- Generate batched commit strategy with descriptive messages and proper staging
- Create executable staging plan with review capabilities and rollback options
- **Validation Criteria**: Comprehensive staging plan generated with intelligent file grouping

**Integration Patterns:**
- GitHub CLI (`gh`) integration for repository management and authentication
- SSH key management for secure remote access and configuration
- File type detection for intelligent .gitignore generation and project analysis
- Git workflow automation with safety validation and error handling

**Parameters:**
- `--org [organization]`: Specify target GitHub organization (default: personal account)
- `--private`: Create private repository (default: true)
- `--dry-run`: Preview all actions without execution for safety validation
- `--force`: Override existing configurations and safety prompts
- `--template [type]`: Specify .gitignore template (auto-detected if not specified)
- `--no-readme`: Skip README.md creation even if absent

**Usage Examples:**

```bash
# Basic repository initialization in current directory
topologist

# Initialize with specific organization (parameter format)
topologist --org mycompany

# Initialize with specific organization (command format)
topologist on mycompany/project-name

# Preview actions without execution
topologist --dry-run

# Force override existing configurations
topologist --force --org enterprise

# Skip README creation
topologist --no-readme --template node
```

**Output Format:**
- Repository topology analysis report
- GitHub repository creation confirmation with SSH clone URL
- Generated .gitignore and README.md file confirmations
- Executable staging plan with batched commit commands
- Validation status for each phase with success/failure indicators

**Error Handling:**
- Git installation and configuration validation
- GitHub CLI authentication and permissions verification  
- SSH key availability and GitHub account access validation
- File system permissions and directory access checks
- Network connectivity and GitHub API availability validation

**Integration with CLI System:**
- Command available via `/topologist` slash command
- Automatic help generation with parameter descriptions
- Tab completion for organization names and template types
- Progress indicators for long-running operations
- Comprehensive error reporting with actionable resolution steps

**Version**: 1.0.0  
**Dependencies**: git, gh CLI, ssh-keygen, file type detection utilities  
**Deployment Status**: Ready for immediate integration and execution