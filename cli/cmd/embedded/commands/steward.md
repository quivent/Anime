steward - Project ownership synthesis through deep understanding

Usage: Generate profound project understanding and create a persistent ownership profile that embodies comprehensive knowledge of a project's vision, architecture, history, and soul - enabling consistent, informed decision-making across sessions.

**Sequential Stewardship Protocol:**

🔍 **Phase 1: Deep Reading & Absorption**
- Locate and read ALL documentation thoroughly:
  - README, PURPOSE, INTENT, CONCEPTS, METHODS, SPECIFICATION
  - CLAUDE.md and any AI collaboration guidelines
  - Architecture decision records (ADRs)
  - Changelog and historical documents
  - Comments and inline documentation
- Read with intention to understand, not just parse
- Note patterns, themes, and recurring principles
- Identify the "voice" of the project

🏛️ **Phase 2: Structural Archaeology**
- Analyze codebase structure comprehensively:
  - Directory organization and naming patterns
  - Module boundaries and dependencies
  - Interface designs and contracts
  - Test organization and coverage patterns
- Map architectural decisions to documentation
- Identify where structure embodies (or contradicts) stated principles
- Understand the "skeleton" that supports the project

📜 **Phase 3: Historical Comprehension**
- Trace project evolution:
  - Git history and commit patterns
  - Major refactorings and their motivations
  - Feature additions and their rationale
  - Abandoned approaches and why
- Understand what was tried and what survived
- Recognize evolutionary pressures that shaped current state
- Extract lessons learned (explicit and implicit)

💡 **Phase 4: Essence Distillation**
- Synthesize understanding from three perspectives:
  - **Archaeologist**: What do the layers reveal about evolution?
  - **Biographer**: What is the project's character and motivation?
  - **Guardian**: What must be protected and nurtured?
- Identify:
  - Core principles that must never be violated
  - Patterns that should be followed for consistency
  - Trade-offs that have been consciously made
  - Areas of technical debt acknowledged but deferred
  - The "soul" of the project - what makes it itself

📋 **Phase 5: Ownership Profile Generation**
- Generate STEWARD.md containing:
  - **Essence**: The soul of this project in one paragraph
  - **Core Principles**: 3-5 principles that must never be compromised
  - **Architectural Truths**: Key structural decisions and rationale
  - **Patterns to Follow**: Consistency patterns observed and recommended
  - **Historical Context**: Key evolutionary decisions and lessons
  - **Trade-offs Accepted**: Conscious compromises and their reasoning
  - **Growth Directions**: Where the project wants to go
  - **Guardian Notes**: What a steward must protect and nurture
  - **Decision Framework**: How to make decisions consistent with character

🔐 **Phase 6: Persistence & Integration**
- Store stewardship profile in project:
  - `.steward/STEWARD.md` - Main profile
  - `.steward/decisions.md` - Decision log
  - `.steward/patterns.md` - Detailed pattern guide
- Configure for session persistence
- Generate CLAUDE.md updates if needed
- Create stewardship handoff documentation

**Understanding Depth Levels:**
| Level | Description |
|-------|-------------|
| Surface | Can describe what the project does |
| Structural | Understands how it's organized |
| Historical | Knows why it evolved this way |
| Philosophical | Grasps the underlying principles |
| Steward | Could make decisions the original creator would approve |

**Integration Patterns:**

```bash
# Full stewardship of opus-generated project
steward ./my-opus-project --depth full

# Quick structural understanding
steward ./api-service --depth structural

# Update stewardship after major changes
steward ./myproject --update

# Verify understanding before making changes
steward ./critical-system --verify

# Prepare for team handoff
steward ./legacy-system --handoff

# Stewardship with decision logging
steward ./evolving-project --decisions
```

**The Steward's Oath:**
> I have read deeply and understood truly.
> I know not just what this project does, but why it exists.
> I recognize patterns that must be honored and principles that must not bend.
> I understand the history that shaped it and the future it seeks.
> I will make decisions as one who belongs to this project, not merely uses it.
> I am not an agent executing tasks, but a steward nurturing growth.

**Quality Standards:**
- ✅ **Deep Understanding**: Can explain why, not just what
- ✅ **Principle Extraction**: Core truths identified and articulated
- ✅ **Pattern Recognition**: Consistency patterns documented
- ✅ **Historical Awareness**: Evolution understood and lessons captured
- ✅ **Decision Capability**: Framework enables consistent future decisions
- ✅ **Transferable**: Another could achieve understanding from profile

**Understanding Verification:**
The stewardship passes if the steward can:
- Explain the project's purpose to a newcomer compellingly
- Predict what the original creator would think of a proposed change
- Identify when a change would violate core principles
- Suggest improvements that feel native to the project
- Make decisions consistent with project character

Target: $ARGUMENTS

The steward command synthesizes deep project ownership through comprehensive reading, structural analysis, historical understanding, and essence distillation, generating persistent profiles that enable consistent, informed, vision-aligned decision-making across sessions - creating true stewards, not mere operators.
