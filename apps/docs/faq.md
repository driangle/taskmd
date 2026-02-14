# FAQ

Quick answers to common questions about taskmd.

## Installation & Setup

### How do I install taskmd?

The easiest way is via Homebrew:

```bash
brew tap driangle/tap && brew install taskmd
```

Alternatively, download pre-built binaries from the [releases page](https://github.com/driangle/taskmd/releases), install with Go (`go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest`), or [build from source](/getting-started/installation).

### What are the system requirements?

Pre-built binaries work on macOS, Linux, and Windows with no additional dependencies. Building from source requires Go 1.22+. The web interface works in any modern browser.

### How do I verify my installation?

Run `taskmd --version` to see version information. Run `taskmd --help` to see available commands.

### How do I set up agent configuration files?

Run `taskmd agents init` to create configuration files for AI coding assistants. Use `--gemini` for Gemini or `--codex` for Codex. These files help AI assistants understand your taskmd workflow.

## Task Format

### What is the task file format?

Tasks are markdown files with YAML frontmatter. The frontmatter contains structured metadata (id, title, status, etc.) enclosed in `---` delimiters, followed by a markdown body for descriptions. See the [Task Specification](/reference/specification) for complete details.

### What frontmatter fields are required?

Only `id` (unique identifier) and `title` (brief description) are required. `status` is recommended. All other fields are optional.

### How do I create my first task?

Create a file like `tasks/001-my-first-task.md`:

```markdown
---
id: "001"
title: "My first task"
status: pending
---

# My First Task

Description of what needs to be done.
```

Then run `taskmd list tasks/` to see your task. See the [Quick Start](/getting-started/) for a full walkthrough.

### What are valid status values?

`pending` (not started), `in-progress` (actively working), `completed` (finished), `blocked` (cannot proceed), and `cancelled` (will not be completed).

### What's the difference between effort and priority?

**Priority** indicates importance (how critical the task is). **Effort** indicates complexity and time. A task can be high priority but small effort (urgent bug fix), or low priority but large effort (nice-to-have feature).

## CLI Usage

### How do I list all tasks?

```bash
taskmd list tasks/
```

Use `--format json` or `--format yaml` for machine-readable output.

### How do I filter tasks?

```bash
# By status
taskmd list --filter status=pending

# By priority
taskmd list --filter priority=high

# Multiple filters (AND logic)
taskmd list --filter status=pending --filter priority=high
```

### How do I find the next task to work on?

```bash
taskmd next tasks/
```

This provides intelligent suggestions based on priorities, dependencies, and status.

### How do I visualize dependencies?

```bash
# Text-based graph
taskmd graph tasks/ --format ascii

# Mermaid diagram
taskmd graph tasks/ --format mermaid
```

The [web interface](/guide/web) also provides an interactive graph view.

## Web UI

### How do I start the web interface?

```bash
taskmd web start --dir tasks/ --open
```

### Does the web UI auto-refresh?

Yes. It watches for file changes and updates automatically via Server-Sent Events.

### What port does the web server use?

Default is 8080. Change it with `--port`: `taskmd web start --port 3000`.

## Dependencies

### How do I add task dependencies?

Add a `dependencies` array to the frontmatter:

```yaml
dependencies:
  - "001"
  - "015"
```

Dependencies indicate which tasks must be completed before this one can start.

### What happens with circular dependencies?

The `validate` command detects circular dependencies and reports an error. Resolve by restructuring your tasks to break the cycle.

### How do dependencies affect task recommendations?

The `next` command only suggests tasks whose dependencies are all completed. It also prioritizes tasks that unblock the most other work.

## Configuration

### Can I set default options?

Yes. Create a `.taskmd.yaml` file in your project root or home directory. See [Configuration](/reference/configuration) for details.

### Where should I put my config file?

- **Project-level**: `.taskmd.yaml` in your project root (takes precedence)
- **Global**: `~/.taskmd.yaml` for user-wide defaults

### What can I configure?

```yaml
dir: ./tasks                    # Default task directory
web:
  port: 8080                   # Web server port
  auto_open_browser: true      # Auto-open browser
```

## Troubleshooting

### Why won't my task file parse?

Common issues: missing `---` delimiters, invalid YAML syntax, incorrect indentation, or missing required fields. Run `taskmd validate tasks/` to see specific errors.

### Why is my task not showing up?

Check that: the file ends with `.md`, frontmatter has valid YAML, required fields (id, title) are present, and you're scanning the correct directory.

### Why is the graph command failing?

Usually caused by circular dependencies or missing task references. Run `taskmd validate tasks/` first to identify issues.

## Best Practices

### How should I organize task files?

For small projects (< 50 tasks), a flat `tasks/` directory works well. For larger projects, use subdirectories by feature area (`tasks/cli/`, `tasks/web/`). Use `NNN-descriptive-title.md` naming.

### How granular should tasks be?

Aim for tasks completable in hours to a few days. Use subtasks (markdown checkboxes) for finer-grained tracking within a task. Create separate tasks when items need individual dependencies or tracking.

### Should I delete completed tasks?

Keep them for historical reference. Use `--exclude-status completed` to hide them from views, or move them to an `archive/` subdirectory.

## Need More Help?

- [Quick Start](/getting-started/) - Hands-on tutorial
- [CLI Guide](/guide/cli) - Full command reference
- [Task Specification](/reference/specification) - Complete format details
- [GitHub Issues](https://github.com/driangle/taskmd/issues) - Report bugs or request features
