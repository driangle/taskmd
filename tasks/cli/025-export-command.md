---
id: "025"
title: "export command - Multi-artifact export"
status: pending
priority: low
effort: medium
dependencies: ["021", "022", "023", "024"]
tags:
  - cli
  - go
  - commands
  - export
created: 2026-02-08
---

# Export Command - Multi-Artifact Export

## Objective

Implement the `export` command to generate multiple output artifacts at once for comprehensive project documentation and CI/CD integration.

## Tasks

- [ ] Create `internal/cli/export.go` for export command
- [ ] Implement `--formats` flag accepting comma-separated list:
  - `json` - JSON snapshot
  - `md` - Markdown board/report
  - `mermaid` - Mermaid graph
  - `dot` - Graphviz DOT
  - `html` - HTML report
- [ ] Implement `--out <dir>` to specify output directory (required)
- [ ] Implement `--group-by <field>` applied to grouped outputs
- [ ] Generate multiple files based on formats:
  - `snapshot.json`, `snapshot.yaml`
  - `board.md`, `report.md`
  - `graph.mmd`, `graph.dot`
  - `report.html`
- [ ] Create output directory if it doesn't exist
- [ ] Include manifest file listing all generated artifacts

## Acceptance Criteria

- `taskmd export --out ./build --formats json,md` generates multiple files
- Output directory is created if missing
- Each format produces its corresponding file
- Manifest file lists all generated artifacts with metadata
- `--group-by` applies to grouped outputs
- Clear error if `--out` is missing
- Works with stdin and explicit file paths

## Examples

```bash
taskmd export --out ./build --formats json,mermaid,md
taskmd export --out ./artifacts --formats json,html,dot --group-by status
cat tasks.md | taskmd export --stdin --out ./output --formats json,md
```
