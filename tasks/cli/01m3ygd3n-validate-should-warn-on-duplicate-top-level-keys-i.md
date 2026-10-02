---
title: "validate should warn on duplicate top-level keys in .taskmd.yaml"
id: "01m3ygd3n"
status: pending
priority: medium
type: bug
tags: ["validate", "config"]
created: "2026-10-02"
effort: small
---

# validate should warn on duplicate top-level keys in .taskmd.yaml

## Steps to Reproduce

1. Create a `.taskmd.yaml` containing the same top-level key twice, for example:

   ```yaml
   id:
     strategy: sequential
   id:
     strategy: ulid
   ```

2. Run `taskmd validate`.
3. Run `taskmd next-id`.

## Expected Behavior

`validate` reports a warning naming the duplicated key (`id`), so the user learns
that part of the config is being ignored. Ideally the config loader rejects or
warns on the file as well.

## Actual Behavior

`validate` prints `All N task(s) are valid`. The rest of the file loads normally
(`task-dir` is honoured), but the duplicated section is dropped and `next-id`
returns sequential `001` regardless of key order. The user has no signal that
their `id` setting was ignored.

Found while fixing the docs: the configuration reference used to show all four
ID strategies in one YAML block with `id:` repeated four times (fixed in e6a94bb).

## Tasks

- [ ] Add a duplicate-top-level-key check to the `.taskmd.yaml` checks run by `validate`
- [ ] Emit a warning that names the duplicated key
- [ ] Add a test with a duplicated `id:` key
- [ ] Decide whether the config loader should also warn at load time

## Acceptance Criteria

- [ ] `taskmd validate` warns on a `.taskmd.yaml` with a duplicated top-level key
- [ ] The warning names the key
- [ ] A regression test covers the duplicated `id:` case

## Environment

- OS: macOS
- Version: taskmd 0.8.0
