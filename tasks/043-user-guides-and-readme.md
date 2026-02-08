---
id: "043"
title: "Create user guides for CLI and Web, update README"
status: pending
priority: high
effort: medium
dependencies: []
tags:
  - documentation
  - user-experience
  - cli
  - web
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
- [ ] Create `/Users/driangle/workplace/gg/md-task-tracker/README.md` with:
  - [ ] Project overview and key features
  - [ ] Quick start section (30-second setup)
  - [ ] Installation instructions (CLI and web)
  - [ ] Links to detailed user guides
  - [ ] Link to TASKMD_SPEC.md for task format reference
  - [ ] Contributing guidelines (reference CLAUDE.md for developers)
  - [ ] License information

### CLI User Guide
- [ ] Create `docs/guides/cli-guide.md` with:
  - [ ] Installation (homebrew, go install, binary download)
  - [ ] Basic concepts (tasks, statuses, dependencies)
  - [ ] Common workflows:
    - [ ] Creating and organizing tasks
    - [ ] Listing and filtering tasks
    - [ ] Validating task files
    - [ ] Visualizing dependencies with graph
    - [ ] Finding next task to work on
    - [ ] Exporting tasks
  - [ ] Command reference (organized by use case, not alphabetically):
    - `list` - View and filter tasks
    - `validate` - Check task file correctness
    - `graph` - Visualize task dependencies
    - `next` - Find available tasks
    - `stats` - Project statistics
    - `inspect` - View task details
    - `export` - Export to other formats
    - `web` - Start web interface
  - [ ] Tips and best practices
  - [ ] Troubleshooting common issues
  - [ ] Configuration (`~/.taskmd/config.yaml`)

### Web User Guide
- [ ] Create `docs/guides/web-guide.md` with:
  - [ ] Starting the web server (`taskmd web`)
  - [ ] Navigating the interface
  - [ ] Task list view:
    - [ ] Sorting and filtering
    - [ ] Status updates
    - [ ] Bulk operations
  - [ ] Board view (kanban):
    - [ ] Drag and drop
    - [ ] Status columns
    - [ ] Grouping options
  - [ ] Graph view:
    - [ ] Dependency visualization
    - [ ] Interactive navigation
  - [ ] Task detail view:
    - [ ] Viewing task information
    - [ ] Inline editing (if implemented)
    - [ ] Related tasks
  - [ ] Project switching
  - [ ] Dark mode and preferences
  - [ ] Keyboard shortcuts
  - [ ] Live reload functionality

### Quick Start Guides
- [ ] Create `docs/guides/quickstart.md` with:
  - [ ] 5-minute getting started for CLI users
  - [ ] 5-minute getting started for web users
  - [ ] First task workflow (create, validate, complete)

## Acceptance Criteria

- [ ] README.md exists and provides clear project overview
- [ ] README.md has installation instructions for both CLI and web
- [ ] cli-guide.md covers all major commands with practical examples
- [ ] web-guide.md covers all web interface features with screenshots (if possible)
- [ ] quickstart.md gets users productive in under 5 minutes
- [ ] All guides use consistent formatting and terminology
- [ ] Guides reference each other appropriately
- [ ] All documentation is tested by following steps exactly as written
- [ ] Examples use realistic task scenarios (not foo/bar)
- [ ] Links between documents work correctly

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

- [ ] Fresh install following README instructions works
- [ ] CLI guide commands all execute successfully
- [ ] Web guide steps match actual interface
- [ ] Links between documents work
- [ ] No broken internal references
- [ ] Examples use task IDs/formats that validate
- [ ] Terminology consistent with TASKMD_SPEC.md

## Related Tasks

- Task 036: Generate CLAUDE.md template (completed) - provides AI assistant documentation
- Task 025 (archived): CLI polish & error handling - mentions README creation

## References

- `docs/TASKMD_SPEC.md` - Task format specification
- `CLAUDE.md` - Developer guidelines
- Existing CLI help text in command definitions
- Web interface at `apps/web/src/`
