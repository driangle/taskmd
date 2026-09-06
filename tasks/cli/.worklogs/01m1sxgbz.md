## 2026-09-05T20:16:18Z

Fixed: add --group now refuses when the destination resolves under a directory the scanner skips.

**Decision:** refuse, not warn. A warning still leaves a file on disk that no read command can reach, and the task is unusable either way — better to fail before the file exists.

**Where the check lives:** rather than re-implementing the ignore rule in the CLI, added `Scanner.SkippedSegment(relPath)` to `sdk/go/scanner`. It walks the path segments through the scanner's own `shouldSkipDirectory`, so the write path and the read path can no longer disagree about what "ignored" means. Covers configured `ignore` entries, the default skip dirs, and dot-prefixed names alike, and matches by basename at any depth — so `--group a/content` is refused exactly as `--group content` is.

`apps/cli/internal/cli/add_group.go` renders the message, naming the directory, the reason (the `ignore` key when the user put it there, "always skipped" otherwise) and a remedy. add.go was already 372 lines, so this went in its own file.

**Other write paths:** checked. Only `add` takes a destination directory — `rm` and `archive` operate on tasks found by scanning (an ignored task is unreachable to them by construction), the MCP add tool has no group parameter, and there is no `import` command. No gap to close there.

**Spec:** added an "Ignored Directories" section under File Organization documenting that `ignore` entries are bare directory names matched at any depth under `dir`, that they therefore cannot exclude anything outside `dir`, and the always-skipped set. Ran `make sync-spec`.

**Tests:** `internal/cli/add_group_test.go` (ignored refused, nested basename refused, non-ignored unaffected, default skip dir, hidden dir), an e2e reproducing the exact bug report through a real .taskmd.yaml, and a table test for `SkippedSegment` in the SDK. Full `go test ./...`, `make e2e` and `make lint` pass.

**Not done (out of scope, worth filing separately):** nothing reports how many files were skipped, so the gap between files-on-disk and tasks-scanned is still invisible to `validate --verbose` / `stats` — item 2 of "Related confusion this causes" in the task.
