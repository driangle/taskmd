---
id: "035"
title: "report command - Comprehensive report generation"
status: pending
priority: low
effort: large
dependencies: ["020", "021", "022", "023"]
tags:
  - cli
  - go
  - commands
  - reports
  - mvp
created: 2026-02-08
---

# Report Command - Comprehensive Report Generation

## Objective

Implement the `report` command to generate a single comprehensive report artifact combining summary statistics, task groupings, and dependency graphs.

## Tasks

- [ ] Create `internal/cli/report.go` for report command
- [ ] Support output formats:
  - `md` (default) - Rich markdown report
  - `html` - HTML report with styling
  - `json` - Structured JSON report
- [ ] Include sections:
  - Project summary (stats overview)
  - Tasks grouped by status
  - Critical path analysis
  - Blocked tasks list
  - Dependency graph (optional)
- [ ] Implement `--include-graph` flag to embed dependency visualization
- [ ] Implement `--group-by <field>` for main grouping
- [ ] Implement `--out <file>` to write to file
- [ ] HTML format should include CSS styling and be self-contained
- [ ] Markdown format should be well-structured with headers and sections

## Acceptance Criteria

- `taskmd report` generates comprehensive markdown report
- Report includes stats, grouped tasks, and analysis
- `--format html` produces styled HTML
- `--include-graph` embeds dependency graph
- `--group-by priority` changes grouping strategy
- HTML output is self-contained and viewable in browser
- Works with stdin and explicit file paths

## Examples

```bash
taskmd report > report.md
taskmd report --format html --include-graph --out report.html
taskmd report --group-by status --format json
cat tasks.md | taskmd report --stdin
```
