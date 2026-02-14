# Quick Start

Get up and running with taskmd in under 5 minutes.

## What You'll Learn

- Install taskmd
- Create your first tasks
- Use basic CLI commands
- Launch the web interface

## Prerequisites

- A terminal
- A text editor
- Go 1.22+ (only if building from source)

## Step 1: Install taskmd

The easiest way is via Homebrew:

```bash
brew tap driangle/tap
brew install taskmd
taskmd --version
```

See [Installation](./installation) for all options including pre-built binaries, Go install, and building from source.

## Step 2: Create a Project

```bash
mkdir -p my-project/tasks
cd my-project
```

## Step 3: Create Your First Task

Create `tasks/001-setup-project.md`:

```markdown
---
id: "001"
title: "Set up project repository"
status: pending
priority: high
effort: small
tags:
  - setup
created: 2026-02-09
---

# Set Up Project Repository

## Objective
Initialize the project repository with basic structure.

## Tasks
- [x] Create tasks directory
- [ ] Add .gitignore
- [ ] Initialize git repository
- [ ] Create README

## Acceptance Criteria
- Repository is initialized
- Basic structure is in place
```

## Step 4: Validate Your Task

```bash
taskmd validate tasks/
```

Expected output:
```
✓ All tasks valid
Found 1 task(s)
```

## Step 5: List Your Tasks

```bash
taskmd list tasks/
```

You should see your task displayed in a table format.

## Step 6: Create a Second Task with Dependency

Create `tasks/002-write-docs.md`:

```markdown
---
id: "002"
title: "Write project documentation"
status: pending
priority: medium
effort: medium
dependencies:
  - "001"
tags:
  - documentation
created: 2026-02-09
---

# Write Project Documentation

## Objective
Create comprehensive documentation for the project.

## Tasks
- [ ] Write README
- [ ] Add usage examples
- [ ] Document API

## Acceptance Criteria
- README is complete
- Examples are clear and tested
```

## Step 7: Visualize Dependencies

```bash
taskmd graph tasks/ --format ascii
```

You'll see a dependency graph showing that task 002 depends on 001.

## Step 8: Find Next Task

```bash
taskmd next tasks/
```

taskmd will recommend task 001 since it has no dependencies and is pending.

## Step 9: Launch the Web Interface

```bash
taskmd web start --open
```

This opens a web dashboard with task list, board, graph, and statistics views.

## Step 10: Try the Interactive TUI

```bash
taskmd tui tasks/
```

Navigate with arrow keys, press `/` to search, `q` to quit.

## Next Steps

- [Installation](./installation) - All installation methods
- [Core Concepts](./concepts) - Understand tasks, statuses, and dependencies
- [CLI Guide](/guide/cli) - Full command reference
- [Web Interface](/guide/web) - Web dashboard walkthrough
