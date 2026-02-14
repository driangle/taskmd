# Quick Start Guide

Get up and running with taskmd in under 5 minutes.

## What You'll Learn

- Install taskmd
- Create your first tasks
- Use basic CLI commands
- Launch the web interface

## Prerequisites

- Go 1.22+ (if building from source)
- A terminal
- A text editor

## For CLI Users (5 minutes)

### Step 1: Install taskmd

**Option A: Build from source**
```bash
git clone https://github.com/yourusername/md-task-tracker.git
cd md-task-tracker/apps/cli
make build-full
# Add bin/taskmd to your PATH or use ./bin/taskmd
```

**Option B: Install with Go**
```bash
go install github.com/yourusername/md-task-tracker/cmd/taskmd@latest
```

### Step 2: Create a Project Directory

```bash
mkdir -p my-project/tasks
cd my-project
```

### Step 3: Create Your First Task

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

### Step 4: Validate Your Task

```bash
taskmd validate tasks/
```

Expected output:
```
✓ All tasks valid
Found 1 task(s)
```

### Step 5: List Your Tasks

```bash
taskmd list tasks/
```

You should see your task displayed in a table format.

### Step 6: Check Project Statistics

```bash
taskmd stats tasks/
```

See metrics like total tasks, completion rate, and status breakdown.

### Step 7: Create a Second Task with Dependency

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

### Step 8: Visualize Dependencies

```bash
taskmd graph tasks/ --format ascii
```

You'll see a dependency graph showing that task 002 depends on 001.

### Step 9: Find Next Task

```bash
taskmd next tasks/
```

taskmd will recommend task 001 since it has no dependencies and is pending.

🎉 **Congratulations!** You've mastered the CLI basics.

## For Web Users (5 minutes)

### Step 1: Install and Set Up

Follow steps 1-3 from the CLI guide above to install taskmd and create some tasks.

### Step 2: Start the Web Server

```bash
cd my-project
taskmd web start --open
```

This will:
- Start the server on http://localhost:8080
- Open your browser automatically
- Begin watching for file changes

### Step 3: Explore the Task List

The default view shows all tasks in a sortable table:

- **Click column headers** to sort
- **Use the search box** to filter
- **Click task IDs** to view details
- **Filter by status** using the dropdown

### Step 4: Try the Board View

Click **"Board"** in the navigation:

- Tasks organized by status (columns)
- Drag and drop to change status
- Group by priority, tags, or effort
- Visual overview of progress

### Step 5: Visualize Dependencies

Click **"Graph"** in the navigation:

- Interactive dependency graph
- Click nodes to see task details
- Pan and zoom to navigate
- Identify blockers visually

### Step 6: View Statistics

Click **"Stats"** in the navigation:

- Completion metrics
- Status breakdown
- Priority distribution
- Tag analysis

### Step 7: Edit Task Files

Make changes to task files in your editor. The web interface updates automatically via live reload!

Try editing `tasks/001-setup-project.md`:
```markdown
status: in-progress
```

Watch the web interface update immediately.

### Step 8: Use Keyboard Shortcuts

In the task list:
- `j/k` - Navigate up/down
- `Enter` - Open task details
- `/` - Focus search
- `Esc` - Clear filters

🎉 **You're ready to manage tasks visually!**

## First Real Workflow

Now try a complete workflow:

### 1. Plan Your Work

Create a few interconnected tasks:

```bash
# Task 001: Foundation
# (already created)

# Task 002: Feature A (depends on 001)
# Task 003: Feature B (depends on 001)
# Task 004: Integration (depends on 002, 003)
```

### 2. Find Next Task

```bash
taskmd next tasks/
```

Should recommend task 001 (no dependencies).

### 3. Start Working

Update `tasks/001-setup-project.md`:
```markdown
status: in-progress
```

### 4. Track Progress

Check off subtasks as you complete them:
```markdown
- [x] Create tasks directory
- [x] Add .gitignore
- [x] Initialize git repository
```

### 5. Complete Task

When done:
```markdown
status: completed
```

### 6. Find Next Task

```bash
taskmd next tasks/
```

Now recommends task 002 or 003 (dependencies satisfied).

### 7. Monitor Progress

```bash
taskmd stats tasks/
```

See your completion rate increase!

## Common Patterns

### Weekly Planning

```bash
# See all pending tasks
taskmd list tasks/ --status pending

# Check what's blocked
taskmd graph tasks/ --format ascii --exclude-status completed

# Find next priorities
taskmd list tasks/ --priority high --status pending
```

### Project Overview

```bash
# Quick stats
taskmd stats tasks/

# Visual board
taskmd web start --open
# (then navigate to Board view)
```

### Daily Workflow

```bash
# Morning: What should I work on?
taskmd next tasks/

# During work: Validate changes
taskmd validate tasks/

# End of day: Check progress
taskmd stats tasks/
```

## Next Steps

- **Learn CLI commands**: [CLI User Guide](cli-guide.md)
- **Master the web interface**: [Web User Guide](web-guide.md)
- **Understand task format**: [Task Specification](../taskmd_specification.md)
- **Set up your editor**: Add syntax highlighting for YAML frontmatter

## Tips for Success

1. **Keep tasks atomic**: One clear objective per task
2. **Use dependencies**: Model task relationships explicitly
3. **Update status regularly**: Keep your board current
4. **Tag consistently**: Use a common set of tags across tasks
5. **Validate often**: Run `taskmd validate` before committing
6. **Review statistics**: Use stats to identify bottlenecks

## Troubleshooting

### "No tasks found"

- Check that your tasks directory exists
- Ensure files have `.md` extension
- Verify YAML frontmatter format
- Run `taskmd validate` to check for errors

### "Invalid task format"

- Check YAML frontmatter is properly formatted
- Ensure required fields are present: id, title, status
- Verify status is one of: pending, in-progress, completed, blocked
- Run `taskmd validate` for specific error messages

### Web server won't start

- Check if port 8080 is already in use: `lsof -i :8080`
- Try a different port: `taskmd web start --port 3000`
- Check verbose output: `taskmd web start --verbose`

### Changes not showing in web interface

- Verify live reload is working (check browser console)
- Try refreshing the page manually
- Check file permissions
- Ensure files are saved properly

## Getting Help

- Run `taskmd --help` for command list
- Run `taskmd [command] --help` for command details
- Check [CLI Guide](cli-guide.md) for detailed examples
- Check [Web Guide](web-guide.md) for web features
- Report issues on GitHub

---

**Ready to dive deeper?** Continue to the [CLI User Guide](cli-guide.md) or [Web User Guide](web-guide.md).
