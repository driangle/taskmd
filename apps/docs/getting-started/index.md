# Quick Start

Install taskmd, create a project, and add your first task. The [Tutorial](/getting-started/tutorial) continues from here with dependencies, recommendations, validation, and AI assistant setup.

## 1. Install

```bash
brew tap driangle/tap
brew install taskmd
taskmd --version
```

Other options, including pre-built binaries for Linux, macOS, and Windows, Docker, and `go install`, are on the [Installation](./installation) page.

## 2. Initialize a project

```bash
cd my-project
taskmd init
```

`init` creates a `tasks/` directory, a `.taskmd.yaml` config, task templates, the format specification, and an agent configuration file for Claude Code, Codex, or Gemini if you want one. Run it non-interactively with `taskmd init --task-dir ./tasks --claude`.

## 3. Add a task and list it

```bash
taskmd add "Set up project repository" --priority high
taskmd list
```

The task is a markdown file under `tasks/` with YAML frontmatter. Open it in your editor to fill in the objective, subtasks, and acceptance criteria.

## 4. Open the dashboard

```bash
taskmd web start --open
```

![Tasks view on first launch](/images/web/tasks.view.png)

## Next

- [Tutorial](./tutorial): dependencies, the dependency graph, `next`, validation, and AI assistants
- [Core Concepts](./concepts): statuses, priorities, dependencies, and file organization
- [CLI Guide](/guide/cli): full command reference
