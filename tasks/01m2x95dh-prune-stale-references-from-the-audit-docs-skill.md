---
title: "Prune stale references from the audit-docs skill"
id: "01m2x95dh"
status: completed
priority: medium
type: chore
tags: ["skills", "docs"]
created: "2026-09-19"
completed_at: 2026-09-19
---

# Prune stale references from the audit-docs skill

## Description

The `audit-docs` skill tells the agent to cross-reference two files that no
longer exist. Phase 5 instructs:

- "**`docs/guides/cli-guide.md`** (standalone docs)" — SKILL.md line 78
- "**`docs/guides/web-guide.md`** (standalone docs)" — SKILL.md line 90

`docs/guides/` is not in the repository at all. Every audit run spends effort
checking for files that were removed, and the instruction to check "both
locations" implies a second documentation tree that no longer exists — which
can push an audit toward reporting phantom gaps or recreating dead files.

Found while running `/audit-docs` against the web host-bind change
(branch `feat/web-host-bind`).

## Scope note

The skill is duplicated, byte-identical, in two places:

- `.claude/skills/audit-docs/SKILL.md`
- `.agents/skills/audit-docs/SKILL.md`

**Resolved while working this task:** they are not two copies.
`.claude/skills/audit-docs` is a git-tracked **symlink** to
`../../.agents/skills/audit-docs`, so both paths are the same file and cannot
drift. `diff` reports no differences for that reason. The open question about
generating or symlinking one from the other is therefore already answered, and
no follow-up task was filed.

## Tasks

- [x] Remove the `docs/guides/cli-guide.md` sub-step from Phase 5 ("CLI
      Commands -> Documentation") in both copies.
- [x] Remove the `docs/guides/web-guide.md` sub-step from Phase 5 ("Web
      Features -> Documentation") in both copies.
- [x] Reword the surrounding text that assumes two documentation trees, so the
      remaining checks read as the single source they now are
      (`apps/docs/**`).
- [x] Re-read Phase 7 (`--fix`), which references the same removed files when
      describing what to update. — Verified: Phase 7 named only `apps/docs/...`
      paths and needed no change.
- [x] Audit the rest of SKILL.md for other paths that no longer resolve —
      verify each referenced path exists before leaving it in.
- [x] Confirm the two copies remain byte-identical after the edit
      (`diff .claude/skills/audit-docs/SKILL.md .agents/skills/audit-docs/SKILL.md`).

## Acceptance Criteria

- [x] No path referenced in either copy of SKILL.md fails to resolve in the
      repository.
- [x] `docs/guides/` appears nowhere in either copy.
- [x] `diff` between the two copies reports no differences.
- [x] A `/audit-docs` run completes without reporting a missing-file gap for
      any path the skill itself named.
