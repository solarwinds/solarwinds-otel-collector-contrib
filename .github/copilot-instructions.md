# REPOSITORY INSTRUCTIONS

## Repository Overview

This is a public repository for SolarWinds OTel Collector components.
It has a private counterpart repository for private components.

Components from both repositories are part of the SolarWinds OTel Collector builds.

## Mandatory Development Workflow

**Initialize all work by stating: "FOLLOWING REPOSITORY INSTRUCTIONS..."**

### Command Reference
- Execute `make help | grep -v "install"` to see how to build, format, lint, test, release and do other available operations
- Consult README.md for additional guidance

### Required Development Steps

**Step 1: Requirements Analysis**
- Analyze current codebase state relevant to your task
- Document strengths, weaknesses, and technical debt
- Map key components and their interactions
- Create ASCII diagrams for current architecture and proposed changes
- Output findings under "ANALYSIS" header

**Step 2: Architecture Planning**
- Design comprehensive, numbered implementation plan
- Define separation of concerns and component boundaries  
- Specify interfaces, data structures, and interaction patterns
- Identify architectural weaknesses and iterate solutions
- Document final proposal under "PLANNED ARCHITECTURE" header

**Step 3: Implementation**
- Follow existing codebase patterns and conventions
- Write focused, single-responsibility functions
- Ensure code testability through proper abstractions

**Step 4: Testing Strategy**
- Run existing test suites
- Create temporary validation tests as needed
- Implement comprehensive test coverage for new features

**Step 5: Code Review Preparation**
- Conduct critical self-review of all changes
- Verify correctness and adherence to best practices
- Address any obvious issues or technical debt

**Step 6: Quality Assurance**
- Execute lint and format tools, use `make help` to see all available commands
- Resolve all quality violations before submission

**Step 7: Cleanup and Documentation**
- Remove temporary files and dead code
- Update README.md with essential information only
- Update component README.md files when appropriate
- Update vNext section in CHANGELOG.md when appropriate (not necessary for chores, but do it for new features, bugfixes, breaking changes and other significant changes)
- Ensure documentation accuracy and completeness

### Code Quality Standards
- Avoid self-evident comments; document complex logic and architectural decisions
- Maintain consistency with existing codebase patterns, look around for examples before implementing changes
- Prioritize readability and maintainability over cleverness
