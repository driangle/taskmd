# Best Practices and Workflows

This guide covers recommended patterns for organizing tasks, structuring projects, and getting the most out of taskmd. Each recommendation includes a rationale so you can decide what fits your workflow.

## Task File Organization

### Flat vs. Grouped Directories

**Start flat, group when it gets noisy.**

A new project with 10-20 tasks works fine with everything in `tasks/`:

```
tasks/
├── 001-project-setup.md
├── 002-user-auth.md
├── 003-api-endpoints.md
└── 004-deploy-pipeline.md
```

Once you have 30+ tasks or distinct work areas, introduce subdirectories:

```
tasks/
├── 001-specification.md
├── backend/
│   ├── 010-database-schema.md
│   ├── 011-api-endpoints.md
│   └── 012-auth-middleware.md
├── frontend/
│   ├── 020-component-library.md
│   └── 021-dashboard-layout.md
└── infra/
    ├── 030-ci-pipeline.md
    └── 031-staging-deploy.md
```

**Why?** Groups reduce visual noise and make `taskmd list tasks/backend` useful for focusing on one area. But premature grouping adds friction when you're still figuring out the shape of the project. Let the structure emerge from the work.

**Tip:** The `group` field is derived from the directory name automatically. You don't need to set it in frontmatter unless you want to override it.

### Naming Conventions

Use the `NNN-descriptive-slug.md` pattern:

- **Zero-pad IDs** so files sort naturally: `001`, `042`, not `1`, `42`
- **Keep slugs short** but meaningful: `012-auth-middleware.md`, not `012-implement-jwt-based-authentication-middleware-for-api.md`
- **Use hyphens**, not underscores or spaces

When groups have independent ID sequences, prefix with the group name:

```
tasks/
├── cli/
│   ├── 015-graph-command.md     # ID: "015"
│   └── 016-validate-command.md  # ID: "016"
└── web/
    ├── 010-board-view.md        # ID: "010"
    └── 011-graph-view.md        # ID: "011"
```

**Why?** Consistent naming makes tasks easy to find with `ls`, `grep`, and tab completion. The slug in the filename gives you context without opening the file.

### Writing Good Tasks

A well-written task has three sections:

1. **Objective** -- what and why (1-2 sentences)
2. **Tasks** -- a checklist of concrete steps
3. **Acceptance Criteria** -- how to know it's done

```markdown
---
id: "042"
title: "Add rate limiting to API"
status: pending
priority: high
effort: medium
type: feature
tags:
  - api
  - security
---

# Add Rate Limiting to API

## Objective

Protect the API from abuse by adding rate limiting to public endpoints.
Without this, a single client can exhaust server resources.

## Tasks

- [ ] Choose a rate limiting strategy (token bucket vs sliding window)
- [ ] Add rate limiting middleware
- [ ] Configure per-endpoint limits
- [ ] Add rate limit headers to responses
- [ ] Write integration tests

## Acceptance Criteria

- Public endpoints return 429 when rate limit is exceeded
- Rate limit headers (X-RateLimit-*) are present on all responses
- Authenticated endpoints have higher limits than anonymous ones
```

**Why?** The objective gives context to anyone (human or AI) picking up the task. The checklist makes progress visible. The acceptance criteria define "done" unambiguously.

::: tip
Keep tasks focused on a single deliverable. If a task has more than 8-10 subtasks, it's probably two tasks. Use `dependencies` to link them.
:::

## Workflows

### Solo Developer

For a single developer working on a project, the default `solo` workflow is straightforward:

```yaml
# .taskmd.yaml
workflow: solo
```

**Typical flow:**

1. Run `taskmd next` to pick a task
2. Mark it `in-progress` with `taskmd set <ID> --status in-progress`
3. Do the work, checking off subtasks as you go
4. Mark it `completed` with `taskmd set <ID> --done`

**Tips for solo workflows:**

- **Use `taskmd next --quick-wins`** when you have limited time. It surfaces small, high-impact tasks.
- **Don't over-specify dependencies.** If you're the only one working, you naturally serialize work. Add dependencies only when the order truly matters (e.g., "build the schema before building the API").
- **Use tags for planning horizons.** Tags like `mvp`, `v2`, `nice-to-have` help you focus on what matters now.
- **Keep the backlog groomed.** Cancel tasks you'll never do. A long list of stale tasks makes `next` less useful.

### Team Workflow with PR Review

For teams where work goes through code review, use the `pr-review` workflow:

```yaml
# .taskmd.yaml
workflow: pr-review
```

**Typical flow:**

1. Pick a task from `taskmd next`
2. Create a feature branch
3. Mark the task `in-progress`
4. Do the work
5. Open a PR and associate it: `taskmd set <ID> --done --add-pr https://github.com/org/repo/pull/123`
6. The task moves to `in-review` (not `completed`)
7. After the PR is merged, mark it `completed`

**Tips for team workflows:**

- **Use `owner` to assign tasks.** Filter with `taskmd list --filter owner=alice` to see who's working on what.
- **Use `touches` for scope awareness.** When two people work on tasks that touch the same files, the `tracks` command warns you about potential merge conflicts.
- **Establish a tag vocabulary.** Agree on a consistent set of tags across the team (e.g., `frontend`, `backend`, `api`, `docs`). Inconsistent tags make filtering useless.
- **Review tasks in PRs.** Since task files live in the repo, changes to task status show up in pull requests. Reviewers can verify that task updates match the code changes.

### AI-Assisted Workflow

When working with AI coding assistants, taskmd becomes the coordination layer between you and the AI:

**With Claude Code (plugin):**

```
/taskmd:next-task           # AI picks the best task
/taskmd:do-task 042         # AI reads the task, plans, and executes
/taskmd:complete-task 042   # AI marks it done after verification
```

**With any AI assistant (direct file access):**

Point the AI at your task files and ask it to work through them:

> "Read the tasks in `tasks/` and work on the highest-priority pending task. Mark it in-progress, complete the subtasks, and mark it done."

**Tips for AI workflows:**

- **Write detailed acceptance criteria.** AI assistants use these to verify their own work. Vague criteria lead to vague results.
- **Use the `verify` field** to define automated checks. The AI can run these to confirm it did the work correctly:

  ```yaml
  verify:
    - type: bash
      run: "go test ./..."
      dir: "apps/cli"
    - type: assert
      check: "New endpoint returns paginated results"
  ```

- **Use `context` to point the AI at relevant files.** This saves it from searching the codebase:

  ```yaml
  context:
    - "apps/cli/internal/api/handlers.go"
    - "docs/api-design.md"
  ```

- **Keep tasks small enough for a single session.** AI assistants work best with focused, well-scoped tasks. A task that says "build the entire authentication system" will produce worse results than three tasks that say "design the auth flow", "implement JWT signing", and "add login endpoint".

## Dependencies

### When to Use Dependencies

Add a dependency when one task **cannot start** until another is **completed**. Common cases:

- **Schema before API**: You can't build endpoints without the data model
- **API before UI**: The frontend needs backend endpoints to call
- **Setup before everything**: Infrastructure tasks often block feature work

```yaml
# 012-api-endpoints.md
dependencies: ["010"]  # Needs the database schema from task 010
```

### When Not to Use Dependencies

Don't add dependencies for:

- **Preference ordering** -- "I'd like to do A before B" is not a dependency. Use priority instead.
- **Same-area work** -- Two frontend tasks don't necessarily depend on each other. Use `touches` to flag overlap instead.
- **Completed tasks** -- Dependencies on already-completed tasks add noise without value. If the task is done, the dependency is already satisfied.

**Why?** Over-linking creates artificial bottlenecks. The `next` command uses dependencies to filter recommendations -- if everything depends on everything, nothing gets recommended.

::: tip
A good rule of thumb: if you could realistically complete task B without task A being done, they're not dependent. They might be *related*, but that's what tags are for.
:::

### Keeping the Graph Clean

Visualize your dependency graph periodically:

```bash
taskmd graph --format ascii --exclude-status completed
```

Look for:

- **Long chains** (A → B → C → D → E) -- can any steps be parallelized?
- **Fan-in bottlenecks** -- one task blocking five others might need to be split
- **Orphan clusters** -- disconnected groups of tasks might need a shared root

## Scopes and Touches

### What They're For

The `touches` field declares which code areas a task modifies. The `tracks` command uses this to detect overlap and organize tasks into parallel work streams.

```yaml
# Task that modifies the graph module
touches:
  - cli/graph
  - cli/output
```

### Setting Up Scopes

Define scopes in `.taskmd.yaml` to map abstract names to real paths:

```yaml
scopes:
  cli/graph:
    description: "Graph visualization and rendering"
    paths:
      - "apps/cli/internal/graph/"
      - "apps/cli/internal/cli/graph.go"
  cli/output:
    paths:
      - "apps/cli/internal/cli/format.go"
  web/board:
    paths:
      - "apps/web/src/components/board/"
```

**Best practices:**

- **Name scopes by area, not by task.** `cli/graph` is reusable; `graph-refactor` is not.
- **Keep scopes granular enough to be useful.** A scope called `cli` that covers the entire CLI codebase doesn't help detect conflicts. `cli/graph` and `cli/output` do.
- **Don't scope everything.** Only define scopes for areas where parallel work actually causes conflicts. A shared utility file that everyone imports doesn't need a scope.
- **Add descriptions** to make validation warnings helpful:

  ```yaml
  scopes:
    cli/graph:
      description: "Graph visualization and dependency rendering"
  ```

### Using Tracks for Parallel Work

Once tasks have `touches` annotations, the `tracks` command groups them into non-conflicting work streams:

```bash
taskmd tracks
```

This is especially useful for teams or when running multiple AI assistants in parallel. Each track can be assigned to a different developer or AI session without risking merge conflicts.

## Phases

### What Phases Are For

Phases group tasks by time-based targets — releases, sprints, project stages, or deadlines. Unlike tags (which categorize by topic), phases answer "when does this need to be done?"

Common uses:
- **Releases**: `v0.2`, `v1.0`
- **Sprints**: `sprint-3`, `2026-Q2`
- **Project stages**: `alpha`, `beta`, `launch`

### Configuring Phases

Define phases in `.taskmd.yaml` to enable validation and due-date-aware ranking:

```yaml
# .taskmd.yaml
phases:
  - name: "v0.2"
    description: "Core CLI features"
    due: 2026-04-01
  - name: "v0.3"
    description: "Web dashboard"
    due: 2026-06-01
```

When phases are configured, `taskmd validate` warns if a task references an undefined phase. Without configuration, all phase values are accepted silently.

### Assigning Tasks to Phases

Set a task's phase in frontmatter or via the CLI:

```yaml
# In the task file
phase: "v0.2"
```

```bash
# Via CLI
taskmd set 042 --phase v0.2
taskmd add --title "New feature" --phase v0.2
```

### Filtering and Viewing by Phase

```bash
# List tasks in a phase
taskmd list --phase v0.2

# Next task for a phase
taskmd next --phase v0.2

# Board grouped by phase
taskmd board --group-by phase

# Stats grouped by phase
taskmd stats --group-by phase
```

### Tips

- **Start with phases when you have deadlines.** If everything is "someday," tags work fine. Once you have a release date, phases make priorities concrete.
- **Keep phase names short.** They appear in tables and boards. `v0.2` is better than `Version 0.2 - Core CLI Features Release`.
- **Add `due` dates** to get phase-aware ranking in `taskmd next`. Tasks in phases with closer due dates score higher.
- **Review phase progress** with `taskmd board --group-by phase` to see which phases are on track.

## CI/CD Integration

### Pre-commit Validation

Add `taskmd validate` to your pre-commit hooks to catch issues before they're committed:

```bash
# .githooks/pre-commit
#!/bin/sh
taskmd validate
if [ $? -ne 0 ]; then
  echo "Task validation failed. Fix errors before committing."
  exit 1
fi
```

Enable the hooks directory:

```bash
git config core.hooksPath .githooks
```

**Why?** Catching invalid task files at commit time is cheaper than discovering them in CI or production. Common catches include duplicate IDs, missing required fields, and broken dependency references.

### CI Pipeline Validation

Add validation to your CI pipeline for tasks that bypass local hooks:

```yaml
# GitHub Actions example
- name: Validate tasks
  run: |
    brew install driangle/tap/taskmd
    taskmd validate --strict
```

### Validation in Pull Requests

Use strict validation to catch warnings as well as errors:

```bash
taskmd validate --strict
```

Strict mode catches additional issues like:
- Tasks without a `created` date
- Missing markdown body
- Non-standard file naming

## Worklogs

### When to Write Entries

Worklogs are most valuable when they capture **decisions** and **context**, not just status updates.

Write a worklog entry when:

- **Starting a task** -- note your approach and initial plan
- **Making a key decision** -- document the options you considered and why you chose one
- **Hitting a blocker** -- describe the problem and what you've tried
- **Finishing a task** -- summarize what was done and any open items

### What to Include

Good worklog entries have:

```markdown
## 2026-02-15T10:30:00Z

Started implementing rate limiting.

**Approach:** Using token bucket algorithm with Redis for distributed
rate limiting. Considered sliding window but token bucket handles
burst traffic better for our use case.

**Completed:**
- [x] Added rate limit middleware
- [x] Configured per-route limits

**Open questions:**
- Should we expose rate limit config via API?
```

**Why?** Worklogs create an audit trail that's invaluable for debugging, onboarding, and understanding past decisions. When someone asks "why did we use token bucket instead of sliding window?", the answer is in the worklog.

### When to Enable Worklogs

Worklogs are opt-in — an absent key means disabled. For team projects, or whenever you want an
audit trail of decisions, turn them on:

```yaml
# .taskmd.yaml
worklogs: true
```

Until you do, `taskmd worklog --add` refuses to write and explains why, and agents skip worklog
steps. Only writing is gated: existing worklogs are always readable regardless of this setting.

## Common Pitfalls

### Over-engineering task structure

**Problem:** Creating deeply nested directories, complex tagging schemes, and dense dependency graphs before you have more than a handful of tasks.

**Fix:** Start simple. A flat `tasks/` directory with 5-10 tasks needs no groups, no scopes, and minimal dependencies. Add structure when the simplicity starts hurting.

### Stale backlogs

**Problem:** A long list of `pending` tasks that nobody intends to work on. This makes `taskmd next` recommendations noisy and project stats misleading.

**Fix:** Regularly review pending tasks. Cancel anything you won't do in the foreseeable future. A smaller, accurate backlog is more useful than a comprehensive, stale one.

```bash
# Find old pending tasks
taskmd list --filter status=pending --sort created
```

### Dependencies as wish lists

**Problem:** Adding dependencies that express preference rather than real blockers. "I'd like to do the API before the docs" becomes a dependency that blocks docs unnecessarily.

**Fix:** Only add a dependency when task B literally cannot be completed without task A's output. Use priority and effort to express ordering preferences instead.

### Huge tasks

**Problem:** Tasks with 15+ subtasks that take days to complete. Progress is invisible, and the task becomes a graveyard of half-checked items.

**Fix:** Split large tasks into 2-4 focused tasks with clear deliverables. Use `dependencies` to sequence them. Each task should be completable in a single work session.

### Inconsistent tags

**Problem:** Using `frontend`, `front-end`, `fe`, and `ui` interchangeably. Filtering by any single tag misses related tasks.

**Fix:** Establish a tag vocabulary early and document it. Use `taskmd list --format json | jq '.[].tags[]' | sort -u` to audit your current tags.

### Ignoring validation

**Problem:** Task files with typos in status values, broken dependency references, or duplicate IDs. These cause silent failures in `next`, `graph`, and `tracks`.

**Fix:** Run `taskmd validate` before committing. Better yet, add it to your [pre-commit hook](#pre-commit-validation).

## Project Organization Examples

### Small Personal Project

```
my-app/
├── .taskmd.yaml
├── tasks/
│   ├── 001-setup.md
│   ├── 002-core-feature.md
│   ├── 003-polish.md
│   └── 004-deploy.md
└── src/
```

```yaml
# .taskmd.yaml
dir: ./tasks
worklogs: false
```

Simple, flat, no groups. Worklogs off because it's just you.

### Medium Team Project

```
platform/
├── .taskmd.yaml
├── tasks/
│   ├── api/
│   │   ├── 010-auth.md
│   │   ├── 011-users-endpoint.md
│   │   └── 012-payments.md
│   ├── web/
│   │   ├── 020-dashboard.md
│   │   └── 021-settings-page.md
│   └── infra/
│       ├── 030-ci-pipeline.md
│       └── 031-monitoring.md
├── apps/
└── docs/
```

```yaml
# .taskmd.yaml
dir: ./tasks
workflow: pr-review
scopes:
  api/auth:
    paths: ["apps/api/auth/"]
  api/users:
    paths: ["apps/api/users/"]
  web/dashboard:
    paths: ["apps/web/src/dashboard/"]
```

Grouped by area, scopes defined for parallel work, PR-review workflow for code review.

### Large Project with AI Assistants

```
enterprise/
├── .taskmd.yaml
├── tasks/
│   ├── 001-specification.md
│   ├── backend/
│   │   ├── .worklogs/
│   │   │   ├── 010.md
│   │   │   └── 011.md
│   │   ├── 010-data-model.md
│   │   └── 011-api-layer.md
│   ├── frontend/
│   │   ├── .worklogs/
│   │   ├── 020-component-lib.md
│   │   └── 021-app-shell.md
│   └── devops/
│       ├── 030-terraform.md
│       └── 031-monitoring.md
├── CLAUDE.md          # AI assistant instructions
└── .mcp.json          # MCP server config
```

```yaml
# .taskmd.yaml
dir: ./tasks
workflow: pr-review
worklogs: true

scopes:
  backend/models:
    description: "Data models and database schema"
    paths: ["apps/backend/models/"]
  backend/api:
    description: "REST API handlers and routing"
    paths: ["apps/backend/api/"]
  frontend/components:
    paths: ["apps/frontend/src/components/"]
```

Full setup with worklogs for audit trail, scopes for parallel AI sessions, and PR-review workflow for team coordination.

## What's Next?

- **[CLI Guide](/guide/cli)** -- full command reference
- **[Configuration](/reference/configuration)** -- all config options
- **[Task Specification](/reference/specification)** -- complete format reference
- **[Claude Code Plugin](/guide/claude-code-plugin)** -- AI assistant integration
