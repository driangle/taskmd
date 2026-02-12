---
id: "054"
title: "Add FAQ.md documentation file"
status: pending
priority: medium
effort: small
dependencies: []
tags:
  - documentation
  - faq
  - user-experience
created: 2026-02-12
---

# Add FAQ.md Documentation File

## Objective

Create a comprehensive FAQ.md file in the `docs/` directory to answer common questions about taskmd, helping users quickly find solutions to typical issues and understand key concepts.

## Context

Users often have similar questions when starting with a new tool. An FAQ document helps reduce support burden and improves the onboarding experience by providing quick answers to common questions about installation, usage, task format, and troubleshooting.

## Tasks

- [ ] Create `docs/FAQ.md` file with proper structure
- [ ] Add installation and setup questions
  - How do I install taskmd?
  - What are the system requirements?
  - How do I verify my installation?
  - Can I install without Homebrew?
- [ ] Add task format questions
  - What is the task file format?
  - What frontmatter fields are required?
  - How do I create my first task?
  - What are valid status values?
  - How do priorities work?
- [ ] Add CLI usage questions
  - How do I list all tasks?
  - How do I filter tasks by status?
  - What output formats are supported?
  - How do I visualize task dependencies?
  - Can I export task data?
- [ ] Add web UI questions
  - How do I start the web interface?
  - What port does the web server use?
  - Does the web UI auto-refresh?
  - Can I use the web UI in production?
- [ ] Add dependency management questions
  - How do I add task dependencies?
  - What happens if dependencies are circular?
  - Can I see a dependency graph?
  - How do dependencies affect task status?
- [ ] Add troubleshooting questions
  - Why won't my task file parse?
  - Why is my task not showing up?
  - How do I fix validation errors?
  - Why is the graph command failing?
- [ ] Add integration questions
  - Can I use taskmd with Git?
  - How do I integrate with CI/CD?
  - Can I use taskmd in a monorepo?
  - Does taskmd work with other task systems?
- [ ] Add best practices questions
  - How should I organize task files?
  - What's a good task naming convention?
  - How granular should tasks be?
  - When should I use subtasks vs separate tasks?
- [ ] Link FAQ from main README.md
- [ ] Review for clarity and completeness

## Acceptance Criteria

- FAQ.md file exists in `docs/` directory
- Contains at least 20 common questions with clear answers
- Questions are organized into logical sections
- Answers are concise but complete (2-4 sentences each)
- Includes code examples where relevant
- Uses proper markdown formatting
- Links to relevant documentation sections
- Referenced from README.md
- No duplicate questions

## Suggested Structure

```markdown
# Frequently Asked Questions (FAQ)

## Installation & Setup
Q: ...
A: ...

## Task Format
Q: ...
A: ...

## CLI Usage
Q: ...
A: ...

## Web UI
Q: ...
A: ...

## Dependencies
Q: ...
A: ...

## Troubleshooting
Q: ...
A: ...

## Integration
Q: ...
A: ...

## Best Practices
Q: ...
A: ...
```

## Example Questions to Include

**Installation:**
- Q: How do I install taskmd on macOS?
  - A: Install via Homebrew: `brew install driangle/tap/taskmd`. Alternatively, download from GitHub releases or install via `go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest`.

**Task Format:**
- Q: What's the minimum required frontmatter for a task?
  - A: Only `id` and `title` are required. Example: `id: "001"` and `title: "My task"`. All other fields (status, priority, effort, etc.) are optional.

**CLI Usage:**
- Q: How do I see only pending tasks?
  - A: Use `taskmd list --status pending` to filter by status. You can also use `--exclude-status completed` to hide completed tasks.

**Troubleshooting:**
- Q: Why is my YAML frontmatter failing to parse?
  - A: Check that strings with special characters are quoted, arrays use proper YAML syntax (`["item1", "item2"]` or bulleted list format), and dates follow ISO 8601 format (`YYYY-MM-DD`).

## References

- Task 043: User guides and README (related)
- Task 046: Documentation site (FAQ will be included)
- `docs/taskmd_specification.md`: Task format reference

## Success Metrics

- Users can find answers to common questions without filing issues
- Reduced number of support questions on recurring topics
- Improved onboarding experience for new users
