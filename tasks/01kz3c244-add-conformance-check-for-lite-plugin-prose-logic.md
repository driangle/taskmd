---
id: "01kz3c244"
title: "Add conformance check for lite plugin prose logic vs CLI"
status: pending
priority: medium
effort: medium
type: improvement
phase: critical-feedback
dependencies: ["01kz3c242"]
tags: [plugins, testing, correctness]
created_at: 2026-08-03
---

# Add conformance check for lite plugin prose logic vs CLI

## Objective

`claude-code-plugin-lite` reimplements core CLI behavior as English prose so it can run
without the Go binary: ID generation (sequential / prefixed / random / ULID), `.taskmd.yaml`
parsing, slug generation, and the full frontmatter template are all re-expressed as
natural-language instructions across 13 `SKILL.md` files. The CLI's core algorithms now exist
twice — once in Go, once in prose — with **no test** guarding the prose copy. Unlike the spec
(which has drift detection), the lite skills can silently diverge from real CLI behavior, and
every CLI behavior change must be manually re-expressed in 13 files.

## Tasks

- [ ] Inventory which CLI behaviors the lite skills reimplement (ID strategies, slug rules,
      frontmatter template, validation rules)
- [ ] Define a small set of golden fixtures (inputs → expected file/frontmatter) shared by both
- [ ] Add a check that the lite skills' documented rules match CLI output on those fixtures
      (e.g. assert generated frontmatter/slug/ID-shape equivalence)
- [ ] Wire the check into CI so lite/CLI drift fails the build
- [ ] Document the "single source of truth is the CLI" contract in the lite plugin README

## Acceptance Criteria

- A CI check fails when the lite skills' documented behavior diverges from the CLI
- The set of duplicated behaviors is explicitly enumerated and covered by fixtures
- Contributors are told, in-repo, that the CLI is authoritative and the lite prose must follow
