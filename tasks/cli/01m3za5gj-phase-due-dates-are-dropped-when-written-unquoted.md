---
title: "Phase due dates are dropped when written unquoted in .taskmd.yaml"
id: "01m3za5gj"
status: pending
priority: medium
type: bug
tags: ["config", "phases"]
created: "2026-10-02"
effort: small
---

# Phase due dates are dropped when written unquoted in .taskmd.yaml

## Steps to Reproduce

1. In `.taskmd.yaml`, declare a phase with an unquoted date, exactly as the
   configuration reference and `docs/.taskmd.yaml.example` show it:

   ```yaml
   phases:
     - id: v1
       name: "v1"
       due: 2026-12-01
   ```

2. Run `taskmd phases`.
3. Change the line to `due: "2026-12-01"` and run it again.

## Expected Behavior

Both spellings show `2026-12-01` in the Due column. YAML parses an unquoted
`YYYY-MM-DD` as a timestamp, and `model.FlexibleTime` already handles the
`!!timestamp` tag for task frontmatter, so config should accept it too.

## Actual Behavior

Unquoted: the Due column shows `-` and `--format json` has no `due`.
Quoted: the date is shown.

`validate` does not warn in either case.

## Cause

`apps/cli/internal/cli/validate.go:370` reads the phase map with
`m["due"].(string)`. Viper decodes the unquoted date into a `time.Time`, so the
type assertion fails and the field is silently skipped. The same map-walking
pattern exists in `apps/cli/internal/cli/web.go:151`; check whether the web
Phases view and `/api/board?groupBy=phase` lose the date the same way.

Found while auditing the docs (session of 2026-10-02): every documented example
uses the unquoted form, so following the docs produces a phase with no due date.

## Tasks

- [ ] Accept both `string` and `time.Time` for `due` when parsing phase config (one helper, used by every phase loader)
- [ ] Add a test with an unquoted date in `.taskmd.yaml`
- [ ] Check the web phase loader for the same assertion
- [ ] Decide whether `validate` should warn on a `due` value it cannot parse

## Acceptance Criteria

- [ ] `taskmd phases` shows the due date for an unquoted `due: 2026-12-01`
- [ ] A regression test covers the unquoted form
- [ ] The documented examples work without adding quotes

## Environment

- OS: macOS
- Version: taskmd 0.8.0
