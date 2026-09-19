---
title: "Prune stale references from the audit-docs skill"
id: "01m2x95dh"
status: pending
priority: medium
type: chore
tags: ["skills", "docs"]
created: "2026-09-19"
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

Both must be edited together, or they drift. Worth deciding whether one should
be generated from the other (or symlinked) rather than hand-maintained — but
that decision is broader than this skill and may belong in its own task.

## Tasks

- [ ] Remove the `docs/guides/cli-guide.md` sub-step from Phase 5 ("CLI
      Commands -> Documentation") in both copies.
- [ ] Remove the `docs/guides/web-guide.md` sub-step from Phase 5 ("Web
      Features -> Documentation") in both copies.
- [ ] Reword the surrounding text that assumes two documentation trees, so the
      remaining checks read as the single source they now are
      (`apps/docs/**`).
- [ ] Re-read Phase 7 (`--fix`), which references the same removed files when
      describing what to update.
- [ ] Audit the rest of SKILL.md for other paths that no longer resolve —
      verify each referenced path exists before leaving it in.
- [ ] Confirm the two copies remain byte-identical after the edit
      (`diff .claude/skills/audit-docs/SKILL.md .agents/skills/audit-docs/SKILL.md`).

## Acceptance Criteria

- [ ] No path referenced in either copy of SKILL.md fails to resolve in the
      repository.
- [ ] `docs/guides/` appears nowhere in either copy.
- [ ] `diff` between the two copies reports no differences.
- [ ] A `/audit-docs` run completes without reporting a missing-file gap for
      any path the skill itself named.
