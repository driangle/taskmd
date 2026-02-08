---
id: "001"
title: "taskmd specification - Define markdown task format conventions"
status: pending
priority: high
effort: medium
dependencies: []
tags:
  - documentation
  - specification
  - core
created: 2026-02-08
---

# taskmd Specification

## Objective

Create a comprehensive specification document that defines the conventions, formats, and best practices for taskmd markdown files. This specification will serve as the canonical reference for how tasks should be structured, organized, and formatted in this project.

## Tasks

- [ ] Create `TASKMD_SPEC.md` or `docs/taskmd-specification.md` specification document
- [ ] Define frontmatter schema:
  - Required fields (id, title, status)
  - Optional fields (priority, effort, dependencies, tags, group, created, description)
  - Field types and valid values
  - Enum definitions (Status, Priority, Effort)
- [ ] Document markdown file naming conventions:
  - File naming patterns (e.g., `001-task-name.md`, `task-name.md`)
  - Best practices for file organization
- [ ] Define directory structure conventions:
  - How groups are derived from directory structure
  - Frontmatter `group` field vs. directory-based groups
  - Recommended directory organization patterns
- [ ] Document dependency reference format:
  - How to reference other tasks (by ID)
  - Circular dependency handling
  - Missing dependency behavior
- [ ] Define status lifecycle:
  - Valid status values: pending, in-progress, completed, blocked
  - Status transitions and meanings
  - When to use each status
- [ ] Define priority levels:
  - Valid values: low, medium, high, critical
  - Guidance on when to use each level
- [ ] Define effort estimates:
  - Valid values: small, medium, large
  - Guidelines for estimating effort
- [ ] Document tag conventions:
  - Tag naming patterns
  - Common/standard tags
  - Best practices for tag usage
- [ ] Include examples:
  - Minimal valid task
  - Fully-specified task
  - Task with dependencies
  - Task with complex markdown body
- [ ] Document best practices:
  - Task granularity guidelines
  - Dependency management
  - Status updates workflow
  - File organization strategies

## Acceptance Criteria

- Specification document is complete and comprehensive
- All frontmatter fields are documented with types and valid values
- Examples are clear and demonstrate key concepts
- Naming conventions are defined
- Directory organization patterns are explained
- Status, priority, and effort enums are fully documented
- Document includes both minimal and complex examples
- Best practices section provides actionable guidance

## Deliverable

A markdown specification document (`TASKMD_SPEC.md` or similar) that can be used as:
- Reference documentation for users creating tasks
- Validation specification for the CLI `validate` command
- Onboarding documentation for new contributors
- Canonical source of truth for task format

## Notes

This specification should be:
- Clear and concise
- Easy to reference
- Include practical examples
- Define validation rules that the CLI can implement
- Extensible for future enhancements

The specification will inform the implementation of the `validate` command (task 019) and serve as documentation for all users of the taskmd system.
