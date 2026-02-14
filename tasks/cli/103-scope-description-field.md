---
id: "103"
title: "Add description field to scopes in .taskmd.yaml"
status: pending
priority: low
effort: small
tags:
  - cli
  - config
  - mvp
created: 2026-02-14
---

# Add Description Field to Scopes in .taskmd.yaml

## Objective

Support an optional `description` field on each scope entry in `.taskmd.yaml` to improve human readability. Currently scopes only define `paths`, making it hard to understand at a glance what a scope represents. A `description` field gives users a plain-language explanation of each scope's purpose.

Example:

```yaml
scopes:
  cli/graph:
    description: "Graph visualization and dependency analysis"
    paths:
      - "apps/cli/internal/graph/"
      - "apps/cli/internal/cli/graph.go"
```

## Tasks

- [ ] Update the config struct to include an optional `Description` field on scope entries
- [ ] Update the config parser to read and preserve the `description` field
- [ ] Include the description in `taskmd validate` output when warning about unknown `touches` values
- [ ] Update the specification (`docs/taskmd_specification.md`) to document the `description` field
- [ ] Add tests for parsing scopes with and without descriptions

## Acceptance Criteria

- Scopes with a `description` field are parsed correctly
- Scopes without a `description` field continue to work as before (backward compatible)
- The description is preserved during config read/write operations
- Validate warnings referencing scopes include the description when available
- Specification is updated to document the new field
