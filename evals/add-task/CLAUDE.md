# add-task suite

See [../CLAUDE.md](../CLAUDE.md) for the rules that apply to every suite.

## `workspace/` is frozen — do not add tasks to it

It carries **five** tasks (`001`–`005`), not the shared fixture's six. Two things depend on
that exactly:

- `workspace/.verify/taskmd.go` hardcodes `baselineIDs` as `001`–`005`, and every grader
  identifies the agent's new task as "the one with an unknown ID". A sixth fixture task
  makes every check fail.
- [REPORT.md](REPORT.md)'s committed baseline was measured against this fixture. Changing it
  invalidates those numbers.

Grow [`../fixtures/`](../fixtures) instead — that is what it exists for.

## Grading

This is a **write** skill, so correctness is graded from the filesystem: `check` steps run
`workspace/.verify` against the task file the agent created. Contrast
[`../list-tasks`](../list-tasks), which grades reported output.

`workspace-bare/.verify/` is a byte-identical copy. Change one, copy to the other.
