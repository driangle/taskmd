---
id: "043"
title: "Create user guides for CLI and Web, update README"
status: completed
priority: high
effort: medium
dependencies: []
tags:
  - documentation
  - user-experience
  - cli
  - web
  - mvp
created: 2026-02-08
---

# Create User Guides for CLI and Web, Update README

## Objective

Create comprehensive user guides for taskmd that cover both CLI and web interface usage, and update the README.md to serve as the primary entry point for new users.

## Context

Currently, the project lacks:
- A main README.md in the project root
- User-focused documentation for CLI usage
- User-focused documentation for web interface usage
- Quick start guides for different user personas

This task aims to create accessible, practical documentation that helps users get started quickly and understand both the CLI and web interfaces.

## Tasks

### README.md (Project Root)
- [x] Create `/Users/driangle/workplace/gg/md-task-tracker/README.md` with:
  - [x] Project overview and key features
  - [x] Quick start section (30-second setup)
  - [x] Installation instructions (CLI and web)
  - [x] Links to detailed user guides
  - [x] Link to TASKMD_SPEC.md for task format reference
  - [x] Contributing guidelines (reference CLAUDE.md for developers)
  - [x] License information

### CLI User Guide
- [x] Create `docs/guides/cli-guide.md` with:
  - [x] Installation (homebrew, go install, binary download)
  - [x] Basic concepts (tasks, statuses, dependencies)
  - [x] Common workflows:
    - [x] Creating and organizing tasks
    - [x] Listing and filtering tasks
    - [x] Validating task files
    - [x] Visualizing dependencies with graph
    - [x] Finding next task to work on
    - [x] Exporting tasks
  - [x] Command reference (organized by use case, not alphabetically):
    - `list` - View and filter tasks
    - `validate` - Check task file correctness
    - `graph` - Visualize task dependencies
    - `next` - Find available tasks
    - `stats` - Project statistics
    - `show` - View task details
    - `export` - Export to other formats
    - `web` - Start web interface
  - [x] Tips and best practices
  - [x] Troubleshooting common issues
  - [x] Configuration (`~/.taskmd/config.yaml`)

### Web User Guide
- [x] Create `docs/guides/web-guide.md` with:
  - [x] Starting the web server (`taskmd web`)
  - [x] Navigating the interface
  - [x] Task list view:
    - [x] Sorting and filtering
    - [x] Status updates
    - [x] Bulk operations
  - [x] Board view (kanban):
    - [x] Drag and drop
    - [x] Status columns
    - [x] Grouping options
  - [x] Graph view:
    - [x] Dependency visualization
    - [x] Interactive navigation
  - [x] Task detail view:
    - [x] Viewing task information
    - [x] Inline editing (if implemented)
    - [x] Related tasks
  - [x] Project switching
  - [x] Dark mode and preferences
  - [x] Keyboard shortcuts
  - [x] Live reload functionality

### Quick Start Guides
- [x] Create `docs/guides/quickstart.md` with:
  - [x] 5-minute getting started for CLI users
  - [x] 5-minute getting started for web users
  - [x] First task workflow (create, validate, complete)

## Acceptance Criteria

- [x] README.md exists and provides clear project overview
- [x] README.md has installation instructions for both CLI and web
- [x] cli-guide.md covers all major commands with practical examples
- [x] web-guide.md covers all web interface features with screenshots (if possible)
- [x] quickstart.md gets users productive in under 5 minutes
- [x] All guides use consistent formatting and terminology
- [x] Guides reference each other appropriately
- [x] All documentation is tested by following steps exactly as written
- [x] Examples use realistic task scenarios (not foo/bar)
- [x] Links between documents work correctly

## Implementation Notes

### Documentation Structure
```
/
├── README.md (main entry point)
├── docs/
│   ├── guides/
│   │   ├── quickstart.md
│   │   ├── cli-guide.md
│   │   └── web-guide.md
│   ├── taskmd_specification.md (existing - task format reference)
│   └── templates/
│       └── CLAUDE.md (existing - for AI assistants)
└── CLAUDE.md (developer guide - existing)
```

### Style Guidelines
- Use active voice and imperative mood
- Include practical examples for every feature
- Start each guide with "What you'll learn" section
- Use callouts for tips, warnings, and notes
- Keep examples realistic (use actual task scenarios)
- Include terminal/UI output examples
- Test all commands and instructions

### Target Audiences
1. **CLI users**: Developers comfortable with terminal, want automation and scripting
2. **Web users**: Team members who prefer visual interfaces, less technical
3. **New users**: Need quick wins and clear next steps

### Screenshots (Optional Enhancement)
- If feasible, add screenshots to web guide showing:
  - Main task list view
  - Board/kanban view
  - Graph visualization
  - Task detail view

## Testing Checklist

- [x] Fresh install following README instructions works
- [x] CLI guide commands all execute successfully
- [x] Web guide steps match actual interface
- [x] Links between documents work
- [x] No broken internal references
- [x] Examples use task IDs/formats that validate
- [x] Terminology consistent with TASKMD_SPEC.md

## Related Tasks

- Task 036: Generate CLAUDE.md template (completed) - provides AI assistant documentation
- Task 025 (archived): CLI polish & error handling - mentions README creation

## References

- `docs/TASKMD_SPEC.md` - Task format specification
- `CLAUDE.md` - Developer guidelines
- Existing CLI help text in command definitions
- Web interface at `apps/web/src/`
