---
title: "Report malformed YAML frontmatter in task-like files instead of silently skipping"
id: "01m28s976"
status: completed
priority: high
type: bug
tags: ["validate", "scanner", "parser"]
created: "2026-09-11"
completed_at: 2026-09-11
---

# Report malformed YAML frontmatter in task-like files instead of silently skipping

GitHub issue: https://github.com/driangle/taskmd/issues/19

## Steps to Reproduce

1. Create a task file with a YAML syntax error in its frontmatter, e.g. mixing flow and block list styles:

   ```markdown
   ---
   id: "003"
   title: "Lorem"
   status: pending
   priority: medium
   dependencies: [ ]
   tags: [
      - "backlog"
   ]
   created_at: 2026-09-11
   ---

   # Lorem
   ```

2. Run `taskmd validate` (or `list`, or open the web UI).

## Expected Behavior

`validate` reports the file as invalid (with the YAML parse error and file path) and exits non-zero. Other surfaces (list/web/MCP) at least have the error available in scan results.

## Actual Behavior

The scanner silently skips the file (`sdk/go/scanner/scanner.go`, ~line 109: parse errors are dropped unless `--verbose`). `validate` reports "all N tasks are valid" with the broken task missing from the count, and the task vanishes from the web UI with no feedback. The parser (`sdk/go/parser/markdown.go`) already returns a proper `ParseError` — the scanner discards it because it cannot distinguish "not a task file" from "broken task file".

## Design Constraints

Not every `.md` file with a frontmatter block is a task (docs pages, blog posts, Jekyll/Hugo content can live under a scanned directory). The fix must not make `validate` fail on foreign markdown. Agreed approach: **key sniffing** —

- A YAML parse failure is escalated to a `ScanError` **only if** the raw frontmatter text contains task-signature keys (line-anchored `id:`, `status:`, `priority:`, or `dependencies:`).
- Files whose frontmatter is valid YAML but which lack `id`/`title` remain silently skipped (unchanged behavior — that is how non-task docs with frontmatter are already handled).
- Files with malformed frontmatter and no task-signature keys remain silently skipped.
- The sniff is deliberately loose (regex over raw text, since the YAML failed to parse); document it in code and in the spec as a behavioral contract.

## Tasks

- [x] Parser (`sdk/go/parser`): `ParseError` gained a `Frontmatter []byte` field (set when frontmatter was present but invalid, including unclosed blocks) and a `MalformedTaskFrontmatter()` method.
- [x] Add a task-signature key sniffer: `parser.HasTaskSignature` (line-anchored match for `id:`, `status:`, `priority:`, `dependencies:`) in `sdk/go/parser/sniff.go`.
- [x] Scanner (`sdk/go/scanner`): `Scan` escalates sniffed parse failures into `result.Errors` instead of silently skipping. `ScanArchive` was deliberately left unchanged — its signature has no error channel and adding one is a breaking SDK change; archived broken files still surface indirectly as missing-dependency errors.
- [x] `validate` (`apps/cli/internal/cli/validate.go`): scan errors merge into the validation result as errors (named file + YAML error, exit 1) via `addScanErrors`.
- [x] Other consumers: `reportScanErrors` (used by list/board/graph/next/etc.) now warns on stderr without requiring `--verbose` (`--quiet` suppresses); web/MCP go through the same scanner and tolerate the errors.
- [x] Tests: scanner tests for reported/skipped/unclosed cases, parser sniffer tests, e2e tests reproducing issue #19 (`validate` fails, foreign markdown passes, `list` warns).
- [x] Documented the sniffing contract in `docs/taskmd_specification.md` (Validation section) and ran `make sync-spec`.
- [x] SDK change is additive → patch bump; CI auto-heals the pin (no manual bump needed unless releasing).

## Acceptance Criteria

- Running `taskmd validate` on a directory containing the issue #19 example file exits non-zero and names the file with a YAML parse error.
- A non-task markdown file with malformed frontmatter and no task-signature keys does not fail validation and produces no error.
- A markdown file with valid-YAML frontmatter but no `id`/`title` is still silently skipped (no regression).
- Existing scanner/validate tests pass; new tests cover the three cases above.

## Environment

- OS: any
- Version: 0.6.1 (reported against main)
