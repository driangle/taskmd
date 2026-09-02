# list-tasks suite

See [../CLAUDE.md](../CLAUDE.md) for the rules that apply to every suite.

## Read-only skill: graded from output, not the filesystem

`list-tasks` mutates nothing, so there is no created file to inspect. `check_output` pipes
the agent's final text to a stdlib-only Go grader on **stdin**
(`workspace/.verify/output.go`), which asserts the exact expected ID set **two-sided** —
expected IDs present, competing ones absent. A presence-only check would pass an agent that
dumps every task.

Every eval also runs `no-mutation` (the fixture is untouched) and `tool_not_used`.
`Write`/`Edit` are deliberately reachable so `no-mutation` can actually fail.

## After touching a grader

```bash
cd workspace/.verify && GOWORK=off go test ./...   # both directions, zero tokens
cp -R workspace/.verify/. workspace-bare/.verify/  # the two must stay identical
```

The tests pin every bug found so far, using the verbatim agent text that broke each one. Add
a case rather than only fixing the regex.

## Two grader rules that are easy to undo

- **Structure decides what counts, not keywords.** When an answer has two or more list items
  (lines starting with an ID), only those count and prose is commentary. Agents routinely
  close a correct answer by naming what they excluded — *"the other two tasks (006 export
  CSV, 004 update README) are in the Polish phase, not MVP"* — and that sentence contains no
  cue word. Do not "fix" this with a blocklist of words like "excluded"; that was tried and
  it loses.
- **Match IDs and title tokens, never phrasing or layout.** Grading the plugin skill's
  pretty-printed table would measure formatting compliance and unfairly fail `no-skill`.

## Other

- Fixture is **six** tasks (`001`–`006`). Graders assume six.
- Keep `run:` in the `/bin/sh -c '…'` form — it is correct regardless of how a given skival
  build resolves a relative `run` string.
- `list-scope-filter` means the task **group** (`tasks/cli/`), not taskmd's `--scope` flag;
  the fixture configures no `scopes:`.
