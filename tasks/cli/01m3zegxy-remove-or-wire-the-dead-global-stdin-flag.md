---
title: "Remove or wire the dead global --stdin flag"
id: "01m3zegxy"
status: pending
priority: low
type: bug
tags: ["cli", "flags"]
created: "2026-10-02"
effort: small
---

# Remove or wire the dead global --stdin flag

## Steps to Reproduce

1. In a project with valid tasks, pipe an invalid task to validate:

   ```bash
   printf -- '---\nid: "x"\n---\nno title\n' | taskmd validate --stdin
   ```

## Expected Behavior

Either the piped task is validated (and fails for the missing title), or the
flag does not exist.

## Actual Behavior

`✓ All 3 task(s) are valid`: the flag is accepted and ignored, and the task
directory is scanned as usual. `--stdin` is bound as a persistent flag in
`apps/cli/internal/cli/root.go` and shown in `taskmd --help`, but nothing reads
it; `InputResolver` in `input.go` has no callers.

The docs no longer advertise the flag (removed from the CLI guide's Global
Flags table on 2026-10-02), but the root help and the `snapshot` help example
still do.

## Tasks

- [ ] Decide: remove the flag and `InputResolver`, or wire `--stdin` into the commands it makes sense for (`validate`, `snapshot`)
- [ ] Update root help and the `snapshot` help example accordingly
- [ ] Add a test for whichever behaviour is chosen

## Acceptance Criteria

- [ ] `taskmd --help` does not list a flag that does nothing
- [ ] If kept, `taskmd validate --stdin < file` validates the piped content

## Environment

- OS: macOS
- Version: taskmd 0.8.0
