---
title: "Model task access intent and generate queue-ready metadata"
id: "01ky5jabs"
status: pending
priority: medium
effort: large
type: feature
tags: ["task-metadata", "safe-queue", "concurrency", "skills"]
touches:
  - "plugin/skills"
  - "cli"
context:
  - "claude-code-plugin/skills/add-task/SKILL.md"
  - "claude-code-plugin/skills/safe-queue/SKILL.md"
  - "claude-code-plugin/skills/safe-queue/scripts/safe_queue.py"
  - "docs/taskmd_specification.md"
verify:
  - type: bash
    run: "cd apps/cli && go test ./..."
  - type: bash
    run: "cd apps/cli && make e2e"
  - type: assert
    check: "The add-task skill produces explicit, queue-ready touches, context, verify, and resources metadata from a sufficiently detailed task request."
created: "2026-07-22"
---

# Model task access intent and generate queue-ready metadata

## Objective

Make newly created tasks queue-ready by default. Taskmd and its add-task
agent workflow must distinguish files or scopes that may be modified from
read-only context and verification-only references, and must represent
exclusive operational resources explicitly.

The workflow should infer this metadata conservatively from a sufficiently
detailed request, validate it, and surface ambiguity instead of silently
creating tasks whose concurrency safety cannot be assessed.

## Tasks

- [ ] Add `resources` as a first-class optional task field in the canonical
      specification, Go model, parser, serializers, CLI output, validation,
      templates, MCP surfaces, and editor integrations.
- [ ] Define `touches` as mutable scope intent, `context` as read-only
      references, `verify` as completion checks, and `resources` as exclusive
      operational dependencies.
- [ ] Preserve the distinction between absent `resources` and explicit
      `resources: []`, so schedulers can reject missing operational evidence
      while accepting an explicit no-exclusive-resources assertion.
- [ ] Update `taskmd add` or the add-task skill to populate `touches`,
      `context`, `verify`, and `resources` from the task request.
- [ ] Resolve mutable scopes through `.taskmd.yaml`; permit precise literal
      paths where supported and reject ambiguous prose as mutable scope
      evidence.
- [ ] Ensure optional creation targets and missing read-only context files do
      not incorrectly make mutable scope evidence stale.
- [ ] Require the task creator to surface unresolved access or resource
      ambiguity rather than guessing that concurrent work is safe.
- [ ] Add validation and actionable diagnostics for missing, unknown, broad,
      or unresolvable mutable scopes and absent resource declarations.
- [ ] Update Safe Queue to consume the first-class metadata rather than
      reparsing custom frontmatter.
- [ ] Add unit and end-to-end coverage for inferred metadata, explicit empty
      resources, exclusive-resource conflicts, missing context files, creation
      targets, and ambiguous task descriptions.
- [ ] Update task templates, plugin documentation, specification copies, and
      examples.

## Acceptance Criteria

- New tasks created from sufficiently detailed requests contain accurate
  `touches`, `context`, `verify`, and explicit `resources` metadata without a
  second manual cleanup pass.
- `touches` contains only scopes or paths the task may mutate; files that are
  merely read or tested are represented by `context` or `verify`.
- `resources: []` is distinguishable from an omitted resources declaration.
- Ambiguous access intent produces an actionable warning or clarification
  requirement instead of optimistic scheduling metadata.
- Safe Queue can assess generated tasks without custom frontmatter parsing or
  conflating missing context with stale mutable scope.
- The canonical specification and generated copies remain synchronized.
- CLI unit tests and end-to-end tests pass.
