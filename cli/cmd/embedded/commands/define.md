# define - System Functionality Analysis and Documentation Generator

Comprehensive system functionality analysis through documentation and code examination with structured markdown output generation.

Usage: Analyze target systems, correlate documentation with implementation, and generate complete functionality definitions with feature mapping and architectural insights.

**Sequential Functionality Analysis Protocol:**

🎯 **Phase 1: Discovery & Inventory**
- Enumerate all source files, documentation, and configuration files
- Identify primary programming languages and frameworks used
- Catalog build systems, dependencies, and external integrations
- Generate comprehensive file structure mapping and classification

🔍 **Phase 2: Code Analysis & Pattern Recognition**
- Parse source code for function definitions, class structures, and APIs
- Extract comments, docstrings, and inline documentation content
- Identify design patterns, architectural decisions, and code organization
- Analyze dependencies and inter-module relationships across codebase

📚 **Phase 3: Documentation Correlation**
- Cross-reference code functionality with existing documentation
- Identify documentation gaps, inconsistencies, and outdated information
- Extract business logic and domain-specific functionality descriptions
- Map user-facing features to underlying implementation components

⚙️ **Phase 4: Functionality Classification**
- Categorize features by domain (UI, API, data processing, etc.)
- Classify functionality by user roles and access patterns
- Identify core vs. auxiliary features and system capabilities
- Generate functionality hierarchy and dependency mapping structure

📝 **Phase 5: Structured Documentation Generation**
- Generate comprehensive functionality overview with system architecture
- Create feature-by-feature detailed descriptions with implementation details
- Document API endpoints, CLI commands, and user interface components
- Include usage examples, integration patterns, and configuration options

🔄 **Phase 6: Validation & Quality Assurance**
- Verify completeness of functionality coverage against source code
- Validate accuracy of descriptions against actual implementation behavior
- Ensure markdown formatting consistency and structural organization
- Generate summary statistics and completeness metrics for documentation

**Command Parameters:**

```bash
define [target_path] [--output filename] [--depth level] [--format type]
```

**Parameter Specifications:**
- `target_path`: Directory or project to analyze (default: current directory)
- `--output`: Output filename for generated documentation (default: functionality.md)
- `--depth`: Analysis depth level [surface|detailed|comprehensive] (default: detailed)
- `--format`: Output format specification [markdown|json|structured] (default: markdown)

**Integration Patterns:**

📋 **Automated Workflow Integration**
- CI/CD pipeline integration for documentation generation on code changes
- Git hook integration for automatic functionality updates on commits
- Development workflow integration with documentation synchronization
- Quality gate integration with completeness validation requirements

🔧 **Tool Ecosystem Integration**
- IDE plugin support for real-time functionality documentation
- API documentation generation with OpenAPI/Swagger compatibility
- Code comment synchronization with functionality descriptions
- Dependency tracking integration with package managers and build systems

📊 **Analysis Output Formats**

**Markdown Documentation Structure:**
```markdown
# System Functionality Overview
## Core Components
## Feature Catalog
## API Reference
## CLI Commands
## Configuration Options
## Integration Capabilities
## Usage Examples
## Architecture Insights
```

**Success Metrics:**

✅ **Functionality Coverage**: 95%+ of implemented features documented  
✅ **Accuracy Validation**: Code-documentation correlation verification  
✅ **Structural Consistency**: Standardized markdown formatting and organization  
✅ **Completeness Analysis**: Comprehensive feature inventory with classification  
✅ **Integration Readiness**: Deployment-ready documentation for immediate use  

**Quality Standards:**

🎯 **Semantic Precision** - Accurate functionality descriptions matching implementation  
📚 **Documentation Completeness** - Comprehensive coverage of all system capabilities  
🔄 **Synchronization Accuracy** - Real-time correlation between code and documentation  
⚡ **Performance Efficiency** - Optimized analysis algorithms for large codebases  
🏗️ **Architectural Clarity** - Clear system structure and component relationship mapping  

**Target**: functionality.md - Analyzes the documentation and code and defines the functionality of the system

The define command provides comprehensive system functionality analysis, generating structured documentation that bridges the gap between implementation and user understanding through systematic code analysis, documentation correlation, and feature mapping with deployment-ready output specifications.