---
description: Execute specific phases of the 8-phase MORCHESTRATED_COMMUNICATION_PROTOCOL
argument-hint: [phase-number|phase-name] [path] [--validate] [--quality-check] [--detailed]
allowed-tools: Task(morchestrator), Bash(morchestrator orchestrate:*), Write, Edit, Read
model: claude-sonnet-4-20250514
---

Use the morchestrator agent to execute specific phases of the comprehensive 8-phase development protocol.

**Phase:** ${1:-all}
**Target Path:** ${2:-.}
**Protocol Options:** $ARGUMENTS

Available protocol phases:
1. **Requirements Analysis & Decomposition** - Extract explicit/implicit requirements, constraints
2. **Architecture Design & Component Structure** - Architectural patterns, component relationships
3. **Technology Stack Selection** - Technology-requirement matching, context consideration
4. **Development Environment Setup & Project Scaffolding** - Project structure, dependencies
5. **Core Implementation** - Incremental development, separation of concerns
6. **Testing & Quality Assurance** - Comprehensive testing strategy, quality checks
7. **Documentation & User Guides** - Complete documentation suite creation
8. **Build, Package, and Deploy** - Multi-platform distribution and deployment

The morchestrator will execute the specified phase(s) with systematic methodology, quality gates at each step, and comprehensive validation ensuring 90% accuracy and 95% rigor standards throughout the development process.