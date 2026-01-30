# Design - Modular Templated Design System

Transform prompts into applicable modular, templated designs that facilitate fast iterations across different project formats.

Usage: Apply intelligent template selection and composition for rapid design implementation with structured output generation.

**Sequential Modular Design Protocol:**

🎯 **Phase 1: Project Context Analysis**
- Analyze current directory structure for project type indicators
- Scan for package.json, Cargo.toml, README.md, or other metadata files
- Identify existing design patterns and architectural conventions
- Determine optimal template categories (CLI, web, system, documentation)

🔍 **Phase 2: Template Selection Engine**
- Match prompt requirements to available template libraries
- Access ~/Documents/Projects/Templates/ for standardized templates
- Apply intelligent template filtering based on project context
- Prioritize modular components for maximum flexibility

🧩 **Phase 3: Modular Component Assembly**
- Dynamically combine template pieces based on prompt specifications
- Apply zero-redundancy principles to eliminate template overlap
- Ensure component compatibility and seamless integration
- Optimize for fast iteration and modification cycles

📝 **Phase 4: Structured Design Generation**
- Create comprehensive design.md with modular template application
- Generate structured markdown with clear section hierarchies
- Include implementation guidelines and rapid modification protocols
- Provide fast iteration checkpoints and validation criteria

🔄 **Phase 5: Fast Iteration Pipeline**
- Enable rapid design modifications without full regeneration
- Implement hot-reload capabilities for template adjustments
- Create feedback loops for design refinement and optimization
- Support incremental updates and component swapping

✅ **Phase 6: Implementation & Binary Update**
- Implement design changes directly into project codebase
- Automatically rebuild binaries when CLI projects are detected
- Install updated binaries to ~/.local/bin/ for global access
- Validate implementation through functional testing
- Generate deployment-ready documentation and implementation guides
- Establish maintenance protocols for design evolution

**Template Categories:**

🖥️ **CLI Templates**
- Go + Cobra command-line interface frameworks
- C-based system utilities and performance tools
- Terminal user interface (TUI) and interactive command systems
- Binary deployment and installation protocols
- **AUTOMATIC BINARY REBUILDS**: Detects CLI projects and rebuilds/installs binaries
- **GLOBAL INSTALLATION**: Updates ~/.local/bin/ automatically after implementation

🌐 **Web Application Templates**
- React component architecture and state management
- Full-stack application design with API integration
- Responsive design systems and UI component libraries
- Progressive web application (PWA) specifications

🔧 **System Architecture Templates**
- Microservices architecture and distributed system design
- Database schema design and data pipeline architectures
- Infrastructure as code and deployment automation
- Performance monitoring and observability frameworks

📚 **Documentation Templates**
- Technical specification and API documentation
- User guides and tutorial creation frameworks
- Project organization and README standardization
- Knowledge base and learning resource structures

**Fast Iteration Features:**

⚡ **Rapid Modification**
- Component-level updates without full regeneration
- Template hotswapping for design experimentation
- Incremental validation and error checking
- Real-time preview and feedback systems

🔄 **Template Composition**
- Mix-and-match modular components
- Cross-category template integration
- Custom template creation from existing components
- Template versioning and rollback capabilities

🎨 **Design Customization**
- Project-specific template adaptation
- Brand and style guide integration
- Custom naming conventions and architectural patterns
- Flexible output format configuration

**Integration Patterns:**

🤖 **Agent Coordination**
- Templatist agent for comprehensive template creation
- Pattern-Recognition-Primer for advanced pattern analysis
- CherryPicker for selective template component extraction
- Architect for structural framework design validation

🔧 **Tool Integration**
- File system analysis for project format detection
- Template library access and management
- Version control integration for design tracking
- Automated testing and validation protocols
- **BINARY BUILD AUTOMATION**: Automatic detection of CLI projects requiring compilation
- **INSTALLATION PIPELINE**: Streamlined binary deployment to global PATH locations

📊 **Output Specifications**
- Structured design.md file generation
- Implementation roadmap and timeline creation
- Resource requirements and dependency specifications
- Success metrics and validation criteria

**Usage Examples:**

```bash
# Basic design generation from prompt
design "Create a CLI tool for data analysis with visualization"

# Web application design with specific framework
design "Build a React dashboard for project metrics" --template=web

# System architecture design
design "Design microservices architecture for e-commerce" --template=system

# Documentation framework
design "Create technical documentation system" --template=docs
```

**Template Integration:**
- Access to ~/Documents/Projects/Templates/cli-dashboard-template/
- Integration with existing CLI template extraction systems
- Connection to Collaborative Intelligence agent frameworks
- Compatibility with established project organization patterns

**Quality Standards:**
✅ **Template Coherence** - Ensure logical template component integration
✅ **Fast Iteration** - Sub-second modification and preview capabilities  
✅ **Project Alignment** - Match templates to detected project formats
✅ **Modular Design** - Component-based architecture for maximum flexibility
✅ **Documentation Quality** - Clear implementation guides and specifications

**Performance Metrics:**
- Template selection accuracy: >90% project format matching
- Generation speed: <2 seconds for standard templates
- Iteration speed: <500ms for component modifications
- User satisfaction: >95% design applicability rating
- Implementation success: >85% direct deployment viability

**CRITICAL IMPLEMENTATION PROTOCOL:**

🔨 **CLI Project Detection & Auto-Build**
When design command detects CLI projects (presence of main.go, Makefile, Cargo.toml, etc.):

1. **IMPLEMENT DESIGN CHANGES**: Apply visual and structural improvements to codebase
2. **REBUILD BINARY**: Execute appropriate build commands (go build, make, cargo build)
3. **INSTALL GLOBALLY**: Copy binary to ~/.local/bin/ with proper permissions
4. **VALIDATE INSTALLATION**: Test binary functionality and global accessibility
5. **REPORT STATUS**: Confirm successful implementation and installation

**Auto-Build Triggers:**
- Go projects: Detects go.mod → runs `go build` → installs to ~/.local/bin/
- C projects: Detects Makefile → runs `make build` → runs `make install`
- Rust projects: Detects Cargo.toml → runs `cargo build --release` → installs binary
- Python CLIs: Detects setup.py → runs `pip install -e .` for editable install

**Implementation Commands:**
```bash
# Go CLI projects
go build -ldflags="-s -w" -o [binary_name] .
chmod +x [binary_name]
cp [binary_name] ~/.local/bin/

# C projects with Makefile
make clean && make build && make install

# Rust projects
cargo build --release
cp target/release/[binary_name] ~/.local/bin/
```

**Command Classification**: Implementation Command - Design generation WITH automatic binary deployment
**Complexity Level**: Advanced Protocol (6 phases + build automation) - Full implementation pipeline
**Integration Scope**: System Integration - Cross-template coordination with automated binary management