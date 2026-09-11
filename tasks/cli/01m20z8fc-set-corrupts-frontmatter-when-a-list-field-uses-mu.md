---
id: "01m20z8fc"
title: "set corrupts frontmatter when a list field uses multi-line flow style"
status: completed
priority: high
type: bug
effort: small
dependencies: []
tags: ["set", "frontmatter", "yaml"]
touches: ["cli/set"]
created_at: 2026-09-08
completed_at: 2026-09-11
---

# set corrupts frontmatter when a list field uses multi-line flow style

## Objective

`taskmd set --depends-on` (and `--add-touches` / `--add-pr`, which share the
same writer) produces **invalid YAML** when the field it is updating is written
as a *multi-line* flow sequence — the shape Prettier emits whenever the
single-line form exceeds the print width:

```yaml
dependencies:
  ["01m1yxnq6", "01m1y0xgd", "01m1ymgp7", "01m1yp22d"]
```

It exits 0, prints a normal `Updated task …` summary, and leaves a damaged file.
There are three outcomes, depending on the operation:

1. **Invalid YAML** — the common case. taskmd then drops the task silently:
   `get` cannot find it, `list` omits it, and `validate` reports `✓ All N
   task(s) are valid` for the reduced N. Nothing says a file was skipped.
2. **Valid YAML, values lost** — the `--add-*` flags write a list missing the
   entries that were already there, while the summary claims they were kept.
3. **Valid YAML, a neighbouring field silently corrupted** — the key-removing
   paths leave an orphan that folds into the scalar above it, e.g.
   `status: 'pending ["aaa000001"]'`.

Each is covered below.

Writing success while the file becomes unreadable is the defect. The single-line
form `dependencies: ["a", "b"]` is handled correctly, so whether a project hits
this depends only on how long its lists have grown.

## Steps to Reproduce

```bash
mkdir repro && cd repro && git init -q . && mkdir tasks
printf 'dir: ./tasks\nid:\n  strategy: ulid\n  length: 9\n' > .taskmd.yaml

for id in dep000001 dep000002; do
  printf -- '---\nid: "%s"\ntitle: "Stub"\nstatus: pending\ncreated_at: 2026-09-08\n---\n\n# Stub\n' "$id" \
    > "tasks/$id-stub.md"
done

cat > tasks/tgt000001-target.md <<'YAML'
---
id: "tgt000001"
title: "Target"
status: pending
dependencies:
  ["dep000001"]
created_at: 2026-09-08
---

# Target
YAML

taskmd set tgt000001 --depends-on dep000001,dep000002; echo "exit=$?"
cat tasks/tgt000001-target.md
taskmd get tgt000001
taskmd list
taskmd validate
```

## Expected Behavior

The field is rewritten in place, as it already is for the single-line form:

```yaml
dependencies: ["dep000001", "dep000002"]
```

If a shape cannot be rewritten safely, `set` should refuse and say so, rather
than write a file it cannot read back.

## Actual Behavior

```
$ taskmd set tgt000001 --depends-on dep000001,dep000002
Updated task tgt000001 (Target):
  dependencies: [dep000001] -> [dep000001, dep000002]
exit=0
```

A block sequence is inserted **before** the original flow node, which is left
behind as an orphan under the same key:

```yaml
dependencies:
  - dep000001
  - dep000002
  ["dep000001"]
```

PyYAML on that frontmatter: `while scanning a simple key`. And taskmd stops
seeing the task at all:

```
$ taskmd get tgt000001
No exact match found for "tgt000001". Did you mean:

  1. dep000001: Stub (67% match) [dep000001-stub.md]

$ taskmd list          # only the two stubs
$ taskmd validate
✓ All 2 task(s) are valid
```

`validate` is the most misleading of the three: it is the command run to find
out whether the tasks are in order, and it reports a clean bill of health for a
set that has silently shrunk. (Same failure mode as `01m1sxgbz`, different
cause.)

## Second symptom: existing values are dropped

For the `--add-*` flags the multi-line flow node is also invisible to the code
that computes the new list, so previously-set values are lost — while the
summary line claims they were kept:

```
$ taskmd set aaa000003 --add-touches web
Updated task aaa000003 (Repro):
  touches: [cli] -> [cli, web]        # <- claims both
```

```yaml
touches:
  - web                                # <- only the added one
  ["cli"]
```

The summary is right (it comes from a real YAML parse); the written block list
is not. `--depends-on` does not lose values, because the flag supplies the whole
list.

## Third symptom: the file parses, and a neighbouring field is silently wrong

The two paths that **remove** the key — `--depends-on ""` (clear) and
`--remove-*` when the result is empty — delete the `dependencies:` /`touches:`
line but still leave the orphan behind. With no key above it, the orphan is
indented under whatever scalar precedes it, and YAML folds it into that scalar
as a multi-line plain value. The file parses cleanly, the task stays visible,
and a core field is quietly corrupted:

```yaml
# before
status: pending
dependencies:
  ["aaa000001"]

# after `taskmd set tgt000007 --depends-on ""`
status: pending
  ["aaa000001"]
```

```
$ taskmd get tgt000007
Status: pending ["aaa000001"]
```

PyYAML agrees: `status` is now the string `'pending ["aaa000001"]'`, and
`dependencies` is gone. Here `validate` *does* flag it (`1 error(s)`), because
the value no longer matches the status enum — but that is luck. Had the orphan
landed under a free-text field such as `title`, nothing would have objected at
all.

This is the more dangerous of the two outcomes. A file that fails to parse
disappears loudly enough to notice; a file that parses with the wrong `status`
keeps working and reports the wrong thing.

## Where it goes wrong

`applyListFieldUpdates` (`sdk/go/taskfile/taskfile.go:395-455`) decides between
its inline and multiline branches by testing **the key's own line** for a
bracket:

```go
// Inline format
if strings.Contains(lines[lineIdx], "[") {
    ...
    lines[lineIdx] = FormatInlineList(fieldName, newValues)
    return lines, closeIdx
}

// Multiline format
removeStart := lineIdx + 1
removeEnd := removeStart
for removeEnd < closeIdx && strings.HasPrefix(strings.TrimSpace(lines[removeEnd]), "- ") {
    removeEnd++
}
```

In a multi-line flow sequence the key line is bare `dependencies:`, so the test
at :416 fails and control reaches the block-list branch. That branch only
consumes following lines beginning with `- ` (:430); the flow lines begin with
`[` and `"`, so `removeEnd == removeStart`, **nothing is removed**, and the new
`  - value` lines are spliced in at :446-451 ahead of the surviving flow node.

The fix has to recognise a flow sequence that opens on a line after its key, and
consume through the matching `]` — a `[` on the key line is not the same
question as "is this value a flow sequence". Worth noting the `]` may be several
lines down and entries may carry a trailing comma, which is what Prettier emits.

All three list fields go through this function — `pr` (:114), `touches` (:121),
`dependencies` (:126) — so `--add-pr`, `--remove-pr`, `--add-touches`,
`--remove-touches` and `--depends-on` are all affected.

**Correction (triage):** `tags` *is* affected too. It uses a different function,
`applyTagUpdates` (`sdk/go/taskfile/taskfile.go:322-365`), but that function has
the identical defect — the same `strings.Contains(lines[tagsLineIdx], "[")` key-line
test at :339 and the same `- `-only consume loop at :347. Verified: `set --add-tag`
on a multi-line flow `tags` produces the same orphaned-flow-node corruption. The fix
must cover both functions, or better, extract the shared shape-detection logic once
and route `tags` and the three list fields through it.

## How it was found

In a downstream project (`pinspot`) whose `tasks/` are Prettier-formatted, every
`dependencies` list past ~4 ids is in the multi-line flow form. A `set
… --depends-on` there corrupted the target task's frontmatter; it was caught by
eye in `git diff` before being committed, so the broken state never landed. Had
it been committed, the task would simply have disappeared from every taskmd view
with no error to explain it.

The invocation was a **no-op** in content — the id being added was already in
the list, and the summary printed identical before and after — and it rewrote
the field anyway.

## Tasks

- [x] Make `applyListFieldUpdates` detect a flow sequence whose `[` opens on a
      line after the key, and replace it through its matching `]`, including the
      multi-line and trailing-comma forms Prettier emits
- [x] Fix the companion read path so a multi-line flow value is visible to the
      `--add-*` / `--remove-*` merge, and stop the summary line claiming values
      the file did not receive
- [x] Add the table-driven cases below to `sdk/go/taskfile/taskfile_test.go`
      (landed in a new `sdk/go/taskfile/listfields_test.go`)
- [x] Add the CLI-level case below to the `set` tests
- [x] Consider a guard in the writer: re-parse the frontmatter after writing and
      fail loudly rather than leave an unreadable file. That would have turned
      this into an error message instead of a vanished task
- [ ] Consider having `validate` report files it skipped, so a task that becomes
      unparseable is visible rather than merely absent (see `01m1sxgbz`, which
      raised the same gap from the `add --group` side — file separately if out
      of scope here) — **deferred to `01m1sxgbz`**, which already covers this gap

## Resolution

Fixed by extracting the shape handling into `sdk/go/taskfile/listfields.go`:
`findListField` measures the *full extent* of a field's value rather than assuming
it occupies one line, and `scanFlow` walks a flow sequence from its opening `[` to
the matching `]` across lines, ignoring brackets inside quotes. The four entry
points (`parseCurrentListField`, `parseCurrentTags`, `applyListFieldUpdates`,
`applyTagUpdates`) all route through it, so the read and write paths agree and the
duplication that let `tags` regress independently is gone.

`UpdateTaskFile` now re-parses the frontmatter it is about to write and refuses
rather than leaving an unreadable file.

## Test cases

Existing coverage stops short of this: `TestUpdateTaskFile_MultilineTags`
(`taskfile_test.go:166`) covers the *block* form, for `tags` — which does not go
through `applyListFieldUpdates` at all. No test exercises a **flow sequence
whose `[` opens on a line after the key**, for any field, which is why this
shipped.

Note that `tags` not going through `applyListFieldUpdates` does **not** make it
safe — `applyTagUpdates` carries the same defect independently (see the triage
correction above), so it needs the same shapes covered. Case 11 below is the
verified repro.

**Invariant for every case below** — assert all three, not just the text:

1. the frontmatter re-parses as YAML;
2. the parsed value equals the expected list;
3. no line of the old value survives (the bug leaves an orphan that a
   `strings.Contains` on the *new* value would happily miss).

Assertion 1 is the one that matters. A test that only greps for `"b"` passes
against the corrupted output today.

### `applyListFieldUpdates` — `sdk/go/taskfile`

Run each shape against `dependencies`, `touches` and `pr`; all three share the
writer, and only `dependencies` was hit in the wild.

| #   | Input value                             | Update                     | Expected                          | Today      |
| --- | --------------------------------------- | -------------------------- | --------------------------------- | ---------- |
| 1   | `dependencies: ["a"]`                   | replace with `a,b`         | `dependencies: ["a", "b"]`        | passes     |
| 2   | `dependencies:`⏎`  ["a"]`               | replace with `a,b`         | `dependencies: ["a", "b"]`        | **fails**  |
| 3   | `dependencies:`⏎`  [`⏎`    "a",`⏎`  ]`  | replace with `a,b`         | `dependencies: ["a", "b"]`        | **fails**  |
| 4   | `dependencies:`⏎`  - a`                 | replace with `a,b`         | block form kept, both values      | passes     |
| 5   | `dependencies: []`                      | replace with `a`           | `dependencies: ["a"]`             | passes     |
| 6   | key absent                              | replace with `a`           | key inserted before the `---`     | passes     |
| 7   | `dependencies:`⏎`  ["a"]`               | replace with `[]` (clear)  | key removed, **no orphan left**   | **fails**³ |
| 8   | `touches:`⏎`  ["cli"]`                  | `AddTouches: ["web"]`      | `touches: ["cli", "web"]`         | **fails**¹ |
| 9   | `touches:`⏎`  ["cli", "web"]`           | `RemTouches: ["cli"]`      | `touches: ["web"]`                | **fails**³ |
| 10  | `pr:`⏎`  ["u1"]`                        | `AddPRs: ["u2"]`           | `pr: ["u1", "u2"]`                | **fails**  |
| 11  | `tags:`⏎`  ["alpha"]`                   | `AddTags: ["beta"]`        | `tags: ["alpha", "beta"]`         | **fails**² |

² Case 11 covers `applyTagUpdates`, not `applyListFieldUpdates`. Verified in
triage: `set <id> --add-tag beta` writes `tags:`⏎`  - beta`⏎`  ["alpha"]` and the
task drops out of `get`/`list`/`validate`, exactly as for the other three fields.
Whatever shape-detection helper the fix introduces should be shared with this
function rather than duplicated.

¹ Case 8 is the value-loss symptom: today it writes `touches:`⏎`  - web`⏎
`  ["cli"]` — `cli` is gone from the block list, and the CLI summary still
prints `[cli] -> [cli, web]`.

³ Cases 7 and 9 fail in the **third** way described below — the file still
parses, so assertion 1 alone does not catch them. Assertion 3 is what does.

Every verdict in the **Today** column was executed against `0.6.0` / `a03dc9f`,
not predicted. Cases 7 and 9 in particular did *not* behave the way I first
guessed, which is why the third symptom below exists.

Case 3 is the one that matters in practice: it is what Prettier emits, trailing
comma included, and it is the shape every affected project is already carrying.
Case 7 guards the clear path, which has the same orphan hazard with no new lines
to hide it.

### CLI — `set`

Against a fixture whose `dependencies` is in shape 3:

```
taskmd set <id> --depends-on a,b   # exits 0
taskmd get <id>                     # still finds the task
taskmd list                         # still lists it
taskmd validate                     # still counts it
```

Today `set` exits 0 and the other three quietly stop seeing the task. Asserting
on `get` after the write is what turns this from a formatting nit into a
regression test for the actual damage.

## Acceptance Criteria

- Every case above passes, including the YAML-parse assertion
- `taskmd set --depends-on` on a multi-line flow list produces valid YAML the
  parser reads back, with the requested values
- `--add-touches` / `--add-pr` on a multi-line flow list keep the values already
  there, and the summary matches what was written
- Single-line flow and block-sequence shapes are unchanged
- Clearing a multi-line flow list removes the whole value, leaving no orphan for
  a neighbouring scalar to absorb
- No `set` invocation can leave a task file that taskmd itself cannot read, or
  one it reads back with a field it did not write
- `taskmd validate` passes

## Environment

- OS: macOS (Darwin 25.6.0)
- Reproduced on taskmd **0.6.0**, commit `a03dc9f2aa1ac5e04c5d3d6845e92fc0de876425`
  — this repo's current `main`, so the line numbers above are against HEAD
- Originally hit on taskmd 0.5.0, commit `87ba411b6489921ce6ae1c636ef65232506a1fb5`
