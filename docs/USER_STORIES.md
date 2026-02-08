# User Stories

This document defines user stories for the md-task-tracker project. Stories are organized by persona and feature area.

## Personas

| Persona | Description |
|---------|-------------|
| **Developer** | An individual developer managing tasks for a personal or small-team project |
| **Tech Lead** | Someone overseeing project progress, reviewing dependencies, and planning work |
| **CI/CD System** | An automated pipeline consuming task data for reporting or gating |

---

## Task Authoring

**US-001** As a Developer, I want to create tasks as plain markdown files with YAML frontmatter so that I can use my preferred text editor and version control workflow.

**US-002** As a Developer, I want tasks to have a minimal required schema (id, title, status) so that I can create tasks quickly without filling in unnecessary fields.

**US-003** As a Developer, I want optional fields (priority, effort, tags, dependencies) so that I can add detail incrementally as a task matures.

**US-004** As a Developer, I want to write freeform markdown in the task body (objectives, subtasks, acceptance criteria) so that I can capture rich context alongside the structured metadata.

**US-005** As a Developer, I want to track granular progress with markdown checklists (`- [ ]` / `- [x]`) inside a task file so that I can see sub-step completion without creating separate tasks.

---

## Task Organization

**US-010** As a Developer, I want to organize tasks into subdirectories (e.g., `tasks/cli/`, `tasks/web/`) so that related tasks are grouped by feature area.

**US-011** As a Developer, I want the group field to be automatically derived from the parent directory when I don't set it explicitly so that organization requires no extra effort.

**US-012** As a Developer, I want to tag tasks with labels so that I can categorize and filter across groups.

**US-013** As a Developer, I want to archive completed tasks into an `archive/` subdirectory so that the active task list stays focused.

---

## Dependencies

**US-020** As a Developer, I want to declare dependencies between tasks by ID so that the order of work is explicit.

**US-021** As a Tech Lead, I want to visualize the dependency graph so that I can understand the critical path and parallelizable work.

**US-022** As a Tech Lead, I want to identify the critical path (longest dependency chain) so that I can focus effort on what actually determines the timeline.

**US-023** As a Developer, I want circular and missing dependencies detected automatically so that I don't silently introduce broken references.

**US-024** As a Tech Lead, I want to view upstream (ancestors) or downstream (dependents) of a specific task so that I can assess the impact of delays or scope changes.

---

## Listing and Filtering

**US-030** As a Developer, I want to list all tasks in a table with key columns (id, title, status, priority) so that I get a quick overview.

**US-031** As a Developer, I want to filter tasks by status, priority, effort, group, or tag so that I can focus on what's relevant right now.

**US-032** As a Developer, I want to sort tasks by any field so that I can find the highest-priority or most-recently-created work.

**US-033** As a Developer, I want to choose which columns appear in the table so that I see only what I care about.

**US-034** As a Developer, I want to combine multiple filters (AND logic) so that I can ask precise questions like "all high-priority pending tasks in the cli group."

---

## Board View

**US-040** As a Developer, I want to see tasks in a kanban-style board grouped by status so that I can see work distribution at a glance.

**US-041** As a Tech Lead, I want to group the board by priority, effort, group, or tag so that I can slice the backlog in different ways.

---

## Validation

**US-050** As a Developer, I want to validate my task files against the spec so that I catch formatting errors before they cause problems.

**US-051** As a Developer, I want strict mode to warn about missing optional fields so that I can keep task quality high when I choose to.

**US-052** As a CI/CD System, I want validation to return structured JSON and meaningful exit codes so that I can gate merges on task integrity.

---

## Statistics and Reporting

**US-060** As a Tech Lead, I want to see aggregate project stats (total tasks, status breakdown, priority breakdown, effort breakdown) so that I can assess project health.

**US-061** As a Tech Lead, I want to know how many tasks are blocked and the average dependency depth so that I can spot structural problems.

**US-062** As a Tech Lead, I want stats in JSON format so that I can feed them into dashboards or reports.

---

## Snapshots and Automation

**US-070** As a CI/CD System, I want a machine-readable snapshot of all tasks so that I can track project state over time.

**US-071** As a CI/CD System, I want derived fields (blocked status, topological order, critical path membership) included in the snapshot so that I don't need to recompute graph analysis.

**US-072** As a Tech Lead, I want a core-only snapshot (id, title, dependencies) so that I can produce lightweight diffs between sprints.

**US-073** As a CI/CD System, I want snapshot output in JSON or YAML so that I can integrate with existing tooling.

---

## Graph Export

**US-080** As a Tech Lead, I want to export the dependency graph as a Mermaid diagram so that I can embed it in markdown docs or PRs.

**US-081** As a Tech Lead, I want to export the graph in Graphviz DOT format so that I can render high-quality images.

**US-082** As a Developer, I want an ASCII graph in the terminal so that I can see dependencies without leaving the command line.

**US-083** As a Tech Lead, I want to exclude completed tasks from the graph by default so that the visualization stays relevant.

**US-084** As a Tech Lead, I want to focus the graph on a specific task and its neighborhood so that I can zoom into one area of the project.

---

## Output Flexibility

**US-090** As a Developer, I want every command to support table, JSON, and YAML output so that I can choose human-readable or machine-readable depending on context.

**US-091** As a Developer, I want to write output to a file with `--out` so that I can save reports without shell redirection.

**US-092** As a Developer, I want to pipe task data via stdin so that I can compose commands in a Unix pipeline.

---

## Web UI

**US-100** As a Developer, I want a browser-based interface for viewing and managing tasks so that I don't have to use the terminal for everything.

**US-101** As a Developer, I want to register project folders and switch between them so that I can manage multiple projects from one UI.

**US-102** As a Developer, I want a sortable, filterable task table in the browser so that I get the same power as the CLI in a visual format.

**US-103** As a Developer, I want to edit task status and priority inline in the table so that I can update tasks without opening a file.

**US-104** As a Developer, I want a task creation dialog so that I can add new tasks without writing frontmatter by hand.

**US-105** As a Developer, I want the web UI to read and write the same markdown files as the CLI so that both tools stay in sync with no extra steps.

---

## General

**US-110** As a Developer, I want all task data stored as plain files in my repo so that tasks are versioned, diffable, and portable with no external service.

**US-111** As a Developer, I want no database or server dependency so that I can start tracking tasks by just creating a `tasks/` directory.

**US-112** As a Developer, I want unknown frontmatter fields preserved on read/write so that I can extend the format for my own needs without losing data.
