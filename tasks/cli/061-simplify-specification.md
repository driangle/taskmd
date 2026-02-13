---
id: "061"
title: "Simplify task specification document"
status: pending
priority: medium
effort: medium
dependencies: ["001"]
tags:
  - documentation
  - spec
  - ux
  - mvp
created: 2026-02-12
---

# Simplify Task Specification Document

## Objective

Simplify and clarify the taskmd specification to make it easier to understand, implement, and follow. Remove unnecessary complexity, consolidate redundant sections, and improve overall readability.

## Tasks

- [ ] Review current specification document (`docs/TASKMD_SPEC.md` or `tasks/001-taskmd-specification.md`)
- [ ] Identify areas of complexity or redundancy
- [ ] Simplify frontmatter schema:
  - Mark truly optional fields clearly
  - Remove rarely-used or unnecessary fields
  - Consolidate similar fields
  - Clarify required vs optional fields
- [ ] Reduce number of examples (keep only the most illustrative ones)
- [ ] Consolidate sections that cover similar topics
- [ ] Simplify language and remove verbose explanations
- [ ] Create a "quick reference" section at the top
- [ ] Move advanced/edge-case content to appendix or separate document
- [ ] Update any code that references removed/changed fields
- [ ] Update validation logic to match simplified spec
- [ ] Update tests to reflect simplified spec
- [ ] Review with fresh eyes (or ask for feedback)

## Acceptance Criteria

- Specification is 30-50% shorter than current version
- Core concepts are explained in plain, concise language
- Quick reference section provides at-a-glance guidance
- Examples are clear and cover common use cases
- No essential information is lost in simplification
- Code and validation logic align with simplified spec
- All tests pass with updated spec

## Implementation Notes

### Principles for Simplification

1. **Remove the unnecessary**: Cut fields, examples, or sections that don't add value
2. **Consolidate the similar**: Merge redundant sections or explanations
3. **Clarify the essential**: Make required fields and core concepts crystal clear
4. **Progressive disclosure**: Put basics first, advanced details later
5. **Show, don't tell**: Use examples instead of lengthy explanations

### Areas to Consider

Potential simplifications:
- **Frontmatter fields**: Do we need all of them? Can some be optional or removed?
- **Status values**: Are all status values necessary?
- **Dependencies syntax**: Can it be simpler?
- **Tags format**: Is the current format the simplest?
- **Date formats**: Do we support too many formats?
- **Priority/effort levels**: Could these be simplified?

### What to Keep

Don't sacrifice:
- Core functionality
- Backward compatibility (or document breaking changes)
- Clarity for beginners
- Flexibility for power users (but maybe move to "advanced" section)

### Structure Suggestion

```markdown
# Taskmd Specification

## Quick Reference
[One-page overview with minimal example]

## Core Concepts
[Brief explanation of tasks, frontmatter, status]

## Frontmatter Schema
[Required and optional fields, clearly marked]

## Common Examples
[2-3 examples covering 80% of use cases]

## Advanced Usage (Optional)
[Edge cases, complex scenarios, rarely-used features]
```

## Questions to Answer

- Which frontmatter fields are truly essential?
- Can we reduce the number of valid status values?
- Do we need multiple date formats or just one?
- Can dependencies be expressed more simply?
- Are there any fields that are never or rarely used?

## Examples

### Before (Complex)
```yaml
---
id: "042"
title: "Complex task"
status: in-progress
priority: high
effort: medium
owner: john
assignee: john
created: 2026-01-15T10:30:00Z
updated: 2026-01-20T14:22:00Z
due: 2026-02-01
started: 2026-01-16T09:00:00Z
completed: null
milestone: v1.0
area: backend
component: api
version: 1.0.0
dependencies: ["041", "039"]
blocks: ["043"]
tags:
  - cli
  - urgent
  - bug
labels:
  - high-priority
category: bug
type: task
---
```

### After (Simplified)
```yaml
---
id: "042"
title: "Complex task"
status: in-progress
priority: high
dependencies: ["041", "039"]
tags:
  - cli
  - bug
created: 2026-01-15
---
```

## Success Metrics

- Time to understand spec reduced by 50%
- New contributors can start writing tasks in <5 minutes
- Fewer questions about "what fields do I need?"
- Validation errors are clearer
- Spec document is easier to maintain

## References

- Current specification: `docs/TASKMD_SPEC.md` or `tasks/001-taskmd-specification.md`
- Parser implementation: `internal/parser/`
- Validation logic: `internal/validator/`
