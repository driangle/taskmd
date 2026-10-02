# Best Practices and Workflows

Recommended patterns for organizing tasks, structuring projects, and getting the most out of taskmd. Take what fits your project.

## Task File Organization

### Flat vs. Grouped Directories

Start flat, group when it gets noisy. A project with 10-20 tasks works fine with everything in `tasks/`:

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
│   └── 011-api-endpoints.md
├── frontend/
│   └── 020-component-library.md
└── infra/
    └── 030-ci-pipeline.md
```

The `group` field is derived from the directory name, so `taskmd list --filter group=backend` works without any frontmatter. Premature grouping adds friction while the shape of the project is still emerging, so let the structure follow the work.

### Naming Conventions

Use the `NNN-descriptive-slug.md` pattern that `taskmd add` produces:

- **Zero-pad IDs** so files sort naturally: `001`, `042`, not `1`, `42`
- **Keep slugs short** but meaningful: `012-auth-middleware.md`, not `012-implement-jwt-based-authentication-middleware-for-api.md`
- **Use hyphens**, not underscores or spaces

Consistent names make tasks findable with `ls`, `grep`, and tab completion, and the slug gives context without opening the file. If groups keep independent ID sequences, or several people create tasks on separate branches, switch to the `prefixed`, `random`, or `ulid` [ID strategy](/reference/configuration#id-strategy-configuration) to avoid collisions.

### Writing Good Tasks

A well-written task has three sections: an **Objective** (what and why, one or two sentences), a **Tasks** checklist of concrete steps, and **Acceptance Criteria** that define done unambiguously. The objective gives context to whoever picks the task up, human or AI; the checklist makes progress visible; the criteria are what an assistant will check its own work against.

```markdown
---
id: "042"
title: "Add rate limiting to API"
status: pending
priority: high
effort: medium
type: feature
tags: ["api", "security"]
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

**Size tasks for one work session.** A task with more than about eight subtasks, or one that takes days, hides progress and turns into a graveyard of half-checked items. Split it into focused tasks with clear deliverables and use `dependencies` to sequence them; the `/taskmd:split-task` skill in the Claude Code plugin does this for you.

## Workflows

### Solo Developer

The default `solo` workflow:

1. Run `taskmd next` to pick a task
2. Mark it `in-progress` with `taskmd set <ID> --status in-progress`
3. Do the work, checking off subtasks as you go
4. Mark it `completed` with `taskmd set <ID> --done`

- **Use `taskmd next --quick-wins`** when you have limited time; it surfaces small, high-impact tasks.
- **Don't over-specify dependencies.** Working alone, you serialize naturally. Add a dependency only when order truly matters.
- **Use tags for planning horizons**, such as `mvp`, `v2`, `nice-to-have`, or [phases](#phases) once you have dates.
- **Keep the backlog groomed.** Cancel tasks you will never do; a long list of stale `pending` tasks makes `next` noisy and stats misleading. `taskmd list --filter status=pending --sort created_at` finds the oldest.

### Team Workflow with PR Review

For teams where work goes through code review, set `workflow: pr-review` in `.taskmd.yaml`:

1. Pick a task from `taskmd next` and mark it `in-progress`
2. Create a feature branch and do the work
3. Open a PR and attach it: `taskmd set <ID> --done --add-pr https://github.com/org/repo/pull/123`. In this workflow `--done` moves the task to `in-review`, not `completed`
4. When the PR merges, mark the task `completed`, or let the [`taskmd-complete` action](#complete-tasks-when-prs-merge) do it

- **Use `owner` to assign tasks** and `taskmd list --filter owner=alice` to see who has what.
- **Use `touches` for scope awareness.** When two people work on tasks that touch the same files, `tracks` warns about the overlap.
- **Agree on a tag vocabulary** such as `frontend`, `backend`, `api`, `docs`. `frontend`, `front-end`, `fe`, and `ui` used interchangeably make filtering useless. Audit with `taskmd list --format json | jq '.[].tags[]' | sort -u`.
- **Review task changes in PRs.** Task files live in the repo, so status updates show up in diffs and reviewers can check they match the code.

### AI-Assisted Workflow

With AI coding assistants, taskmd is the coordination layer between you and the assistant.

With the Claude Code plugin:

```
/taskmd:next-task           # pick the best task
/taskmd:do-task 042         # read the task, plan, and execute
/taskmd:complete-task 042   # mark it done after verification
```

With any assistant that has file access, point it at `tasks/` and ask it to work on the highest-priority pending task, mark it in-progress, complete the subtasks, and mark it done.

- **Write detailed acceptance criteria.** Assistants verify their own work against them; vague criteria produce vague results.
- **Use the `verify` field** for automated checks the assistant can run:

  ```yaml
  verify:
    - type: bash
      run: "go test ./..."
      dir: "apps/cli"
    - type: assert
      check: "New endpoint returns paginated results"
  ```

- **Use `context` to point at relevant files** so the assistant does not have to search:

  ```yaml
  context:
    - "apps/cli/internal/api/handlers.go"
    - "docs/api-design.md"
  ```

- **Keep tasks small enough for one session.** "Build the entire authentication system" produces worse results than "design the auth flow", "implement JWT signing", and "add login endpoint".

## Dependencies

Add a dependency when one task **cannot start** until another is **completed**: schema before API, API before UI, infrastructure before features.

```yaml
# 012-api-endpoints.md
dependencies: ["010"]  # needs the database schema from task 010
```

Do not add dependencies for:

- **Preference ordering.** "I'd like to do A before B" is a priority, not a dependency.
- **Same-area work.** Two frontend tasks do not necessarily depend on each other; use `touches` to flag overlap.
- **Completed tasks.** The dependency is already satisfied and only adds noise.

Over-linking creates artificial bottlenecks: `next` only recommends tasks whose dependencies are done, so if everything depends on everything, nothing is recommended. If you could realistically finish B without A, they are related, not dependent, and tags express that.

Check the graph periodically:

```bash
taskmd graph --format ascii
```

Look for long chains that could run in parallel, one task blocking five others that might need splitting, and disconnected clusters that might want a shared root.

## Scopes and Touches

The `touches` field declares which code areas a task modifies. `tracks` uses it to detect overlap and group tasks into parallel work streams, which matters for teams and for running several AI sessions at once.

```yaml
# In a task
touches:
  - cli/graph
  - cli/output
```

Define scopes in `.taskmd.yaml` to map the names to paths:

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
```

- **Name scopes by area, not by task.** `cli/graph` is reusable; `graph-refactor` is not.
- **Keep them granular enough to detect conflicts.** A scope covering the whole CLI tells you nothing; `cli/graph` and `cli/output` do.
- **Don't scope everything.** Only areas where parallel work actually conflicts. A shared utility everyone imports does not need a scope.
- **Add descriptions**; they appear in validation warnings.

Then `taskmd tracks` groups actionable tasks into non-conflicting streams, one per developer or AI session.

## Phases

Phases group tasks by time-based targets: releases (`v0.2`, `v1.0`), sprints (`sprint-3`, `2026-Q2`), or stages (`alpha`, `beta`, `launch`). Tags answer "what is this about"; phases answer "when does this need to be done". If everything is "someday", tags are enough; once you have a release date, phases make priorities concrete.

Define them in `.taskmd.yaml`, **in the order you will work them**:

```yaml
phases:
  - id: "v0.2"
    name: "Core CLI"
    description: "Core CLI features"
    due: 2026-04-01
  - id: "v0.3"
    name: "Web dashboard"
    due: 2026-06-01
```

The order matters. `taskmd next` finishes earlier phases before recommending later ones, so a critical task in `v0.3` waits behind actionable `v0.2` work. Unphased tasks, such as an urgent bug, compete in the current phase rather than queueing behind everything; set `next.unphased: last` if unphased means unscheduled in your project. `taskmd next --strict-phases=false` gives a one-off view that treats phase as a score bonus only.

With phases configured, `validate` warns when a task references an undefined phase. Assign and view:

```bash
taskmd set 042 --phase v0.2
taskmd add "New feature" --phase v0.2
taskmd list --phase v0.2
taskmd next --phase v0.2
taskmd board --group-by phase
taskmd stats --group-by phase
taskmd phases                      # progress per phase
```

Keep phase ids short; they appear in every table and board.

## CI/CD Integration

### Pre-commit Validation

Catching a duplicate ID or broken dependency at commit time is cheaper than finding it in CI:

```bash
# .githooks/pre-commit
#!/bin/sh
taskmd validate || { echo "Task validation failed."; exit 1; }
```

```bash
git config core.hooksPath .githooks
```

### CI Validation

For changes that bypass local hooks:

```yaml
# GitHub Actions
- name: Validate tasks
  run: |
    brew install driangle/tap/taskmd
    taskmd validate --strict
```

`--strict` adds warnings for tasks with no `status`, `priority`, `effort`, `group`, `tags`, or body, and exits with code 2 when any warning is present, so it fails the job. Two consequences: a flat `tasks/` directory warns "no group" for every task, so either use groups or run plain `validate` in CI; and decide as a team which of those fields you actually require before turning strict mode on.

### Complete Tasks When PRs Merge

The repo ships a composite action, `.github/actions/taskmd-complete`, that marks a task `completed` when a PR referencing it merges. It reads the task ID from the PR body or branch name (`task: 042` by default; the pattern is configurable) and commits the status change. `.github/workflows/taskmd-complete-example.yml` is a ready-to-copy workflow. It pairs naturally with the `pr-review` workflow, where tasks sit in `in-review` until merge.

## Worklogs

Worklogs are most valuable when they capture **decisions** and **context**, not just status. Write an entry when starting a task (approach and plan), making a key decision (options considered and why), hitting a blocker (what you tried), and finishing (what was done, what is open). Later, "why did we use token bucket instead of sliding window?" has an answer.

```markdown
## 2026-02-15T10:30:00Z

Started implementing rate limiting.

**Approach:** Token bucket with Redis for distributed limiting. Considered
sliding window but token bucket handles burst traffic better for our case.

**Completed:**
- [x] Added rate limit middleware
- [x] Configured per-route limits

**Open questions:**
- Should we expose rate limit config via API?
```

Worklogs are opt-in. Enable them with `worklogs: true` in `.taskmd.yaml`; until then `taskmd worklog --add` refuses to write and agents skip worklog steps. Existing worklogs are always readable.

## Example Setup

A team project with grouped tasks, scopes for parallel work, PR review, and worklogs:

```
platform/
├── .taskmd.yaml
├── tasks/
│   ├── api/
│   │   ├── .worklogs/
│   │   ├── 010-auth.md
│   │   └── 011-users-endpoint.md
│   ├── web/
│   │   └── 020-dashboard.md
│   └── infra/
│       └── 030-ci-pipeline.md
├── CLAUDE.md          # AI assistant instructions (taskmd init --claude)
└── .mcp.json          # MCP server config, if using taskmd-mcp
```

```yaml
# .taskmd.yaml
task-dir: ./tasks
workflow: pr-review
worklogs: true
scopes:
  api/auth:
    description: "Authentication handlers"
    paths: ["apps/api/auth/"]
  web/dashboard:
    paths: ["apps/web/src/dashboard/"]
```

A solo project needs none of this beyond `task-dir`: a flat `tasks/` directory, no scopes, no worklogs. Add each piece when its absence starts to hurt.

## What's Next?

- **[CLI Guide](/guide/cli)** -- full command reference
- **[Configuration](/reference/configuration)** -- all config options
- **[Task Specification](/reference/specification)** -- complete format reference
- **[Claude Code Plugin](/guide/claude-code-plugin)** -- AI assistant integration
