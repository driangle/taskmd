---
title: "Support renaming a task via set --title"
id: "01kzrxxj9"
status: completed
priority: medium
type: feature
tags: ["cli", "set"]
created: "2026-08-11"
effort: medium
touches: ["cli/set"]
completed_at: 2026-08-11
---

# Support renaming a task via set --title

## Objective

Renaming a task today means hand-editing the `title:` frontmatter field, the `# Heading`
in the body, and the filename slug — three places that silently drift apart. Add a
`--title` flag to `taskmd set` that does all three in one operation.

Requested in [issue #16](https://github.com/driangle/taskmd/issues/16).

Scope: core (mutating a markdown task file — no external systems involved).

## Design

```bash
taskmd set 01kzrxxj9 --title "New name"              # frontmatter + H1, file untouched
taskmd set 01kzrxxj9 --title "New name" --rename     # also rename file to <id>-<new-slug>.md
taskmd set 01kzrxxj9 --title "New name" --dry-run    # preview, including the rename path
```

Decisions to lock in during implementation:

- **File rename is opt-in** (`--rename`), not automatic: the path may be referenced by
  git history, links, or bookmarks, so a title tweak should not move the file by
  surprise. `--rename` without `--title` is an error.
- **New filename** = `<id>-<slug.Slugify(newTitle)>.md` in the same directory, matching
  what `taskmd add` generates. Refuse the rename if the destination already exists.
- **H1 rewrite** targets only the first `# ` heading in the body, and only when it
  matches the old title. If it doesn't match, leave the body alone and note it in the
  output so a custom heading isn't clobbered.
- **Worklog files** (`.worklogs/<ID>.md`) are keyed by ID, not by slug, so a rename
  needs no worklog handling — confirm this while implementing.

`taskfile.UpdateRequest` already carries a `Title *string` field that
`UpdateTaskFile` writes, so the frontmatter half needs no SDK change. The H1 rewrite and
the rename are new.

## Tasks

- [x] Add `--title` and `--rename` flags to `setCmd`, wire `--title` into
      `buildSetRequest` / `hasUpdates` / `buildChangeLog`
- [x] Reject `--rename` without `--title` with a clear error
- [x] Rewrite the leading `# <old title>` heading in the body when it matches the old title
- [x] Implement the file rename: compute `<id>-<slug>.md`, refuse to overwrite an
      existing file, and rename after the content write succeeds
- [x] Show the rename (old path -> new path) in the `set` confirmation output and honor
      `--dry-run` (preview only, no writes)
- [x] Unit tests in `internal/cli/set_rename_test.go` and `sdk/go/taskfile/heading_test.go`:
      title-only, title+rename, H1 matching and non-matching bodies, destination-exists
      conflict, `--rename` without `--title`, dry-run leaves disk untouched
- [x] E2E test covering `set --title --rename` end to end
- [x] Update `set` command help/examples, `apps/docs` CLI reference, and the
      `update-task` plugin skill
- [x] Run `make check` and `make e2e`

Deliberately **not** done: exposing `title` on the MCP `set` tool. That tool's schema is
a narrower surface than the CLI on purpose — it already omits `type`, `parent`, `phase`,
`touches`, and `dependencies` — and giving an agent-facing tool the power to move files
on disk deserves its own decision rather than riding along here.

## Acceptance Criteria

- `taskmd set <id> --title "New name"` updates the `title:` frontmatter field and leaves
  the file path unchanged.
- The body's leading `# <old title>` heading becomes `# New name` when it matched the old
  title; a body whose heading differs is left untouched.
- `--rename` additionally renames the file to `<id>-<new-slug>.md` in the same directory,
  and errors instead of overwriting if that path already exists.
- `--rename` without `--title` fails with an actionable error message.
- `--dry-run` prints the title change and the prospective new path without writing or
  renaming anything.
- The task is still resolvable by ID after a rename, and `taskmd validate` passes.
- Tests cover the cases above and `make check` / `make e2e` are green.
