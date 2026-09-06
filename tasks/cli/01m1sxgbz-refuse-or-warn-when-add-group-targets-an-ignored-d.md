---
title: "Refuse or warn when add --group targets an ignored directory"
id: "01m1sxgbz"
status: completed
priority: medium
type: bug
tags: ["add", "config", "scanner"]
created: "2026-09-05"
effort: small
completed_at: 2026-09-05
---

# Refuse or warn when add --group targets an ignored directory

## Objective

`taskmd add --group <dir>` will happily create a task in a directory the scanner is configured to
ignore. It prints `Created task <id>: <path>` and exits 0, the file lands on disk, and the task is
then invisible to `list`, `get`, `board`, `stats`, `next` and `validate`. Nothing warns, at
creation or afterwards.

Creation succeeding while the result is unreachable is the defect. A user has no signal that
anything is wrong until they go looking for the task and cannot find it.

## Steps to Reproduce

```bash
mkdir repro && cd repro && mkdir tasks
printf 'dir: ./tasks\nid:\n  strategy: ulid\n  length: 9\nignore:\n  - content\n' > .taskmd.yaml

taskmd add "Probe task" --group content
taskmd list
taskmd validate
```

## Expected Behavior

`add` refuses, naming the ignored directory and the config key responsible:

```
Error: --group "content" resolves to tasks/content, which is excluded by the
'ignore' key in .taskmd.yaml. The task would not be visible to list, get or
validate. Remove it from 'ignore', or choose another group.
```

A warn-and-continue variant would also be acceptable, as long as the user is told before they walk
away believing the task exists.

## Actual Behavior

```
$ taskmd add "Probe task" --group content
Created task 01m1rkjtk: .../repro/tasks/content/01m1rkjtk-probe-task.md

$ taskmd list
No tasks found

$ taskmd validate
✓ All 0 task(s) are valid
```

The file is on disk and well-formed. Every read path silently skips it. `validate` reporting
`All 0 task(s) are valid` is the most misleading of the three — it is the command a user runs
specifically to find out whether their tasks are in order.

## Where it goes wrong

`runAdd` (`apps/cli/internal/cli/add.go:116-118`) joins the group onto the scan dir without ever
consulting the ignore list, even though `flags` is already in hand from line 107:

```go
outputDir := scanDir
if addGroup != "" {
    outputDir = filepath.Join(scanDir, addGroup)
}
```

`flags.IgnoreDirs` comes from `viper.GetStringSlice("ignore")` (`root.go:184`) and reaches the
scanner via `newTaskScanner` (`scan.go:15`). So the write path and the read path disagree, and
only the read path knows about `ignore`.

Worth noting while fixing: `ignore` entries are **bare directory names**, not paths.
`NewScanner` folds them into a set and `shouldSkipDir` tests the basename
(`sdk/go/scanner/scanner.go:37-48,150`), so an entry matches a directory of that name at *any*
depth under `dir`. The check in `add` has to match that rule, not do a prefix comparison against
the configured value.

## Related confusion this causes

Encountered in a downstream project whose `.taskmd.yaml` read:

```yaml
dir: ./tasks
ignore:
  - content
```

The intent was to keep the scanner out of the repository's own `content/` directory. But `dir`
already scopes scanning to `./tasks`, so the entry could only ever match `tasks/content/` — which
existed, and held 275 task files (120 of them pending). They were absent from every taskmd view
for months. It surfaced only when `taskmd add --group content` reported success and `taskmd get`
then said "task not found".

Two things made this hard to see, and both are addressable:

1. `add` accepted the ignored group without comment (this task).
2. Nothing ever reports that files were skipped, so the gap between "files on disk" and "tasks
   scanned" is invisible. `validate --verbose`, or `stats`, could say *"skipped 275 files in 4
   ignored directories"*. Consider filing that separately if it is out of scope here.

Documenting that `ignore` matches bare directory names relative to `dir` — and that it therefore
cannot be used to exclude paths outside `dir` — would also help. `docs/taskmd_specification.md`
is the place.

## Tasks

- [x] Decide refuse vs. warn for `add --group <ignored>`; refusing is preferred, since the task is
      unusable either way
- [x] Implement the check in `runAdd`, matching the scanner's basename semantics rather than
      comparing paths
- [x] Check the other write paths for the same gap — anything that takes a destination directory
      (`rm`, `archive`, `import`) has the same disagreement available to it
- [x] Add a CLI test: ignored group refused, non-ignored group unaffected, and a nested group
      whose *basename* is ignored (e.g. `--group a/content`) also refused
- [x] Document in `docs/taskmd_specification.md` that `ignore` entries are directory names matched
      at any depth under `dir`, not paths — then `make sync-spec`

## Acceptance Criteria

- `taskmd add --group <ignored>` does not silently produce an unreachable task
- The message names both the offending directory and the `ignore` key, so the user can act on it
- A group whose basename is ignored at any depth is treated the same as a top-level one
- The spec states the matching semantics of `ignore`

## Environment

- OS: macOS (Darwin 23.2.0)
- Version: taskmd 0.5.0, commit `87ba411b6489921ce6ae1c636ef65232506a1fb5`, built 2026-09-01
