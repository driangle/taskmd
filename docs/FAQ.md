# Frequently Asked Questions (FAQ)

Quick answers to common questions about taskmd.

## Installation & Setup

### Q: How do I install taskmd?

**A:** The easiest way is via Homebrew: `brew tap driangle/tap && brew install taskmd`. Alternatively, download pre-built binaries from the [releases page](https://github.com/driangle/taskmd/releases), install with Go (`go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest`), or build from source using `make build-full`.

### Q: What are the system requirements?

**A:** taskmd requires Go 1.22+ if building from source. The pre-built binaries work on macOS, Linux, and Windows with no additional dependencies. For the web interface, any modern browser (Chrome, Firefox, Safari, Edge) will work.

### Q: How do I verify my installation?

**A:** Run `taskmd --version` to confirm the installation. You should see version information displayed. You can also run `taskmd --help` to see available commands and verify everything is working correctly.

### Q: Can I install without Homebrew?

**A:** Yes! Download pre-built binaries from the [GitHub releases page](https://github.com/driangle/taskmd/releases), use `go install`, or clone the repository and run `make build-full` in the `apps/cli` directory. See the [README installation section](../README.md#installation) for detailed instructions.

### Q: How do I set up agent configuration files?

**A:** Run `taskmd agents init` to create configuration files for AI coding assistants. By default, this creates `CLAUDE.md` with taskmd documentation for Claude Code. Use `--gemini` to create `GEMINI.md` for Gemini, or `--codex` to create `AGENTS.md` for Codex. You can create multiple configs at once: `taskmd agents init --claude --gemini`. These files help AI assistants understand your taskmd workflow and provide better assistance.

## Task Format

### Q: What is the task file format?

**A:** Tasks are markdown files with YAML frontmatter. The frontmatter contains structured metadata (id, title, status, etc.) enclosed in `---` delimiters, followed by a markdown body for detailed descriptions. See the [Task Format Specification](taskmd_specification.md) for complete details.

### Q: What frontmatter fields are required?

**A:** Only three fields are required: `id` (unique task identifier), `title` (brief description), and `status` (current state). All other fields like priority, effort, dependencies, and tags are optional but recommended for better organization.

### Q: How do I create my first task?

**A:** Create a file like `tasks/001-my-first-task.md` with this minimal content:
```markdown
---
id: "001"
title: "My first task"
status: pending
---

# My First Task

Description of what needs to be done.
```
Then run `taskmd list tasks/` to see your task.

### Q: What are valid status values?

**A:** Valid statuses are `pending` (not started), `in-progress` (actively working), `completed` (finished), `blocked` (cannot proceed), and `cancelled` (will not be completed). The typical flow is: pending → in-progress → completed.

### Q: How do priorities work?

**A:** Priorities are optional and include `low` (nice to have), `medium` (standard work), `high` (important for success), and `critical` (urgent). Use them to filter and sort tasks. If not specified, tasks default to `medium` priority.

### Q: What's the difference between effort and priority?

**A:** Priority indicates importance (how critical the task is), while effort indicates complexity and time investment. A task can be high priority but small effort (quick bug fix), or low priority but large effort (nice-to-have feature). They're independent dimensions for planning work.

## CLI Usage

### Q: How do I list all tasks?

**A:** Run `taskmd list <directory>` where `<directory>` is your tasks folder. For example: `taskmd list tasks/` or `taskmd list .` if you're in the tasks directory. Use `--format json` or `--format yaml` for machine-readable output.

### Q: How do I filter tasks by status?

**A:** Use the `--status` flag to show only specific statuses: `taskmd list --status pending`. Use `--exclude-status` to hide certain statuses: `taskmd list --exclude-status completed,cancelled`. You can specify multiple values separated by commas.

### Q: What output formats are supported?

**A:** Most commands support `table` (default, human-readable), `json` (machine-readable, perfect for scripts), and `yaml` (human and machine-readable). Some commands like `graph` also support `ascii` (text-based visualization) and `mermaid` (for documentation).

### Q: How do I visualize task dependencies?

**A:** Use the `graph` command: `taskmd graph tasks/ --format ascii` for a text-based graph, or `--format mermaid` to generate Mermaid.js syntax for embedding in documentation. Add `--exclude-status completed` to focus on active tasks.

### Q: Can I export task data?

**A:** Yes! Use `--format json` or `--format yaml` with any command and redirect to a file: `taskmd list tasks/ --format json > tasks.json`. The `snapshot` command provides timestamped exports: `taskmd snapshot tasks/ --output snapshot.json`.

### Q: How do I find the next task to work on?

**A:** Run `taskmd next tasks/` to get intelligent suggestions based on priorities, dependencies, and status. It shows tasks that are ready to start (no blocking dependencies) and prioritizes high-priority work.

## Web UI

### Q: How do I start the web interface?

**A:** Run `taskmd web start --dir tasks/ --open` to start the server and automatically open your browser. If you don't want auto-open, omit the `--open` flag and manually navigate to `http://localhost:8080`.

### Q: What port does the web server use?

**A:** The default port is 8080. You can change it with `--port`: `taskmd web start --port 3000`. If port 8080 is in use, the server will automatically try the next available port.

### Q: Does the web UI auto-refresh?

**A:** Yes! The web UI watches for file changes and automatically refreshes when task files are modified. This makes it great for seeing real-time updates as you work on tasks in your editor.

### Q: Can I use the web UI in production?

**A:** The web UI is designed for local development and personal use. While it's suitable for small teams on trusted networks, it lacks authentication and advanced security features needed for internet-facing deployments. Consider it a development tool rather than a production application.

## Dependencies

### Q: How do I add task dependencies?

**A:** Add a `dependencies` array to the frontmatter with task IDs:
```yaml
dependencies:
  - "001"
  - "015"
```
Always use the task ID (from the `id` field), not the filename. Dependencies indicate which tasks must be completed before this one can start.

### Q: What happens if dependencies are circular?

**A:** The `validate` command will detect circular dependencies and report an error. For example, if task A depends on B, and B depends on A, validation fails. Circular dependencies indicate a design problem that should be resolved by restructuring your tasks.

### Q: Can I see a dependency graph?

**A:** Yes! Use `taskmd graph tasks/ --format ascii` for a text visualization, or `--format mermaid` for a diagram you can render in documentation. The web UI also shows an interactive graph view under the "Graph" tab.

### Q: How do dependencies affect task status?

**A:** Dependencies are informational and help with planning. The `next` command automatically considers dependencies when suggesting tasks, showing only tasks whose dependencies are completed. However, taskmd doesn't automatically change status based on dependencies—you manage that manually.

## Troubleshooting

### Q: Why won't my task file parse?

**A:** Common issues include: missing `---` delimiters around frontmatter, invalid YAML syntax (unquoted strings with special characters), incorrect indentation, or missing required fields (id, title, status). Run `taskmd validate tasks/` to see specific parsing errors with line numbers.

### Q: Why is my task not showing up?

**A:** Check that: (1) the file ends with `.md`, (2) frontmatter has valid YAML syntax, (3) required fields (id, title, status) are present, (4) you're scanning the correct directory. Run `taskmd validate tasks/` to identify issues.

### Q: How do I fix validation errors?

**A:** Run `taskmd validate tasks/ --verbose` to see detailed error messages. Common fixes: quote strings with colons or special characters, use proper YAML array syntax for tags and dependencies, ensure dates follow `YYYY-MM-DD` format, and verify all dependency IDs reference existing tasks.

### Q: Why is the graph command failing?

**A:** Graph generation fails if there are circular dependencies or missing task references. Run `taskmd validate tasks/` first to identify dependency issues. Also ensure all tasks referenced in `dependencies` arrays actually exist in your task files.

## Integration

### Q: Can I use taskmd with Git?

**A:** Absolutely! Task files are plain markdown, perfect for version control. Commit task files to your repository, and changes are tracked like any other code. Many users keep tasks in the same repo as their code for seamless project management.

### Q: How do I integrate with CI/CD?

**A:** Use `taskmd validate tasks/` in your CI pipeline to ensure task files are valid. Use `taskmd list --format json` to export task data for processing in scripts. You can fail builds if validation fails or if critical tasks are incomplete.

### Q: Can I use taskmd in a monorepo?

**A:** Yes! Each project can have its own `tasks/` directory, or you can have a shared tasks directory at the root. Use subdirectories to organize tasks by project: `tasks/frontend/`, `tasks/backend/`, etc. The `group` field also helps categorize tasks logically.

### Q: Does taskmd work with other task systems?

**A:** Since taskmd uses open formats (markdown and YAML), you can easily migrate to/from other systems. Export to JSON (`taskmd list --format json`) for integration with other tools. The simple format makes it easy to write converters or use task data in custom workflows.

## Best Practices

### Q: How should I organize task files?

**A:** For small projects (< 50 tasks), a flat `tasks/` directory works well. For larger projects, use subdirectories by feature area (`tasks/cli/`, `tasks/web/`) or by milestone. Use consistent naming like `NNN-descriptive-title.md` for easy sorting.

### Q: What's a good task naming convention?

**A:** Use `NNN-descriptive-title.md` where `NNN` is the zero-padded ID (001, 042, 137) and the title is lowercase with hyphens. Examples: `001-project-setup.md`, `042-user-authentication.md`. This keeps files sorted by ID while remaining human-readable.

### Q: How granular should tasks be?

**A:** Tasks should be completable in hours to a few days. Too small ("rename variable") makes overhead costly. Too large ("build entire app") makes tracking progress difficult. Aim for focused, single-objective tasks. Use subtasks in the markdown body for finer-grained tracking within a task.

### Q: When should I use subtasks vs separate tasks?

**A:** Use subtasks (markdown checkboxes `- [ ]`) for steps within a single cohesive task that don't need individual tracking or dependencies. Create separate tasks when work items can be done by different people, have different priorities, or need to be referenced as dependencies by other tasks.

### Q: Should I delete completed tasks?

**A:** Keep completed tasks for historical reference. They document what was done and help with project retrospectives. If your task list gets cluttered, use `--exclude-status completed` when listing, or move completed tasks to an `archive/` subdirectory. Completed tasks also help track velocity and effort estimates.

### Q: How do I handle blocked tasks?

**A:** Set status to `blocked` and document the blocker in the task body. If the blocker is another task, add it as a dependency. If it's external (waiting for approval, third-party API, etc.), describe it clearly. Use `taskmd list --status blocked` to review blockers regularly.

## Configuration

### Q: Can I set default options?

**A:** Yes! Create a `.taskmd.yaml` file in your project directory or home directory. Supported options include `dir` (default task directory), `web.port` (web server port), and `web.auto_open_browser` (auto-open browser on web start). See the [example config file](/.taskmd.yaml.example) for a complete template. Command-line flags always override config file values.

### Q: Where should I put my config file?

**A:** You have two options:
1. **Project-level**: Create `.taskmd.yaml` in your project root for project-specific defaults
2. **Global**: Create `~/.taskmd.yaml` in your home directory for user-wide defaults

Project-level config takes precedence over global config, and command-line flags override both.

### Q: What can I configure in .taskmd.yaml?

**A:** The config file supports three options:
```yaml
dir: ./tasks                    # Default task directory
web:
  port: 8080                   # Default web server port
  auto_open_browser: true      # Auto-open browser on web start
```
Other flags like `format`, `verbose`, and `quiet` are intentionally CLI-only to keep config files focused on project settings.

### Q: How do I change the default task directory?

**A:** Set `dir: ./my-tasks` in your `.taskmd.yaml` file, or use the `--dir` flag: `taskmd list --dir my-tasks/` (or `-d` for short). You can also use environment variables: `export TASKMD_DIR=./my-tasks`. Precedence: CLI flags > project config > global config > env vars > defaults.

## Need More Help?

- **Guides**: Check the [guides directory](guides/) for comprehensive tutorials
- **Specification**: See [taskmd_specification.md](taskmd_specification.md) for complete format details
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/driangle/taskmd/issues)
- **Quick Start**: See [Quick Start Guide](guides/quickstart.md) for a hands-on tutorial
