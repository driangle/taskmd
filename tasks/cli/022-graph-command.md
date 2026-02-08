---
id: "022"
title: "graph command - Export dependency graph"
status: pending
priority: high
effort: large
dependencies: ["017", "019"]
tags:
  - cli
  - go
  - commands
  - visualization
created: 2026-02-08
---

# Graph Command - Export Dependency Graph

## Objective

Implement the `graph` command to export task dependency graphs in various formats for visualization and analysis.

## Tasks

- [ ] Create `internal/cli/graph.go` for graph command
- [ ] Create `internal/graph/` package for graph generation
- [ ] Support output formats:
  - `dot` - Graphviz DOT format
  - `mermaid` - Mermaid diagram syntax
  - `ascii` - ASCII art tree
  - `json` - JSON graph structure
- [ ] Implement `--root <task-id>` to start from specific task
- [ ] Implement `--focus <task-id>` to highlight specific task
- [ ] Implement `--upstream` to show only dependencies (ancestors)
- [ ] Implement `--downstream` to show only dependents (descendants)
- [ ] Implement `--out <file>` to write to file
- [ ] Default: full graph in mermaid format
- [ ] Handle cyclic dependencies gracefully

## Acceptance Criteria

- `taskmd graph` outputs full dependency graph in mermaid format
- `--format dot` produces valid Graphviz DOT
- `--format ascii` shows tree structure in terminal
- `--root T1 --downstream` shows only tasks depending on T1
- `--focus T2` highlights T2 in the graph
- Circular dependencies are detected and displayed
- Works with stdin and explicit file paths

## Examples

```bash
taskmd graph > deps.mmd
taskmd graph --format dot | dot -Tpng > graph.png
taskmd graph --format ascii
taskmd graph --root T1 --downstream
cat tasks.md | taskmd graph --stdin --format mermaid
```
