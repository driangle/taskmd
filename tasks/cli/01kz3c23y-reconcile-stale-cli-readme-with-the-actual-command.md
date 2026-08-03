---
id: "01kz3c23y"
title: "Reconcile stale CLI README with the actual command CLI"
status: pending
priority: high
effort: small
type: docs
phase: critical-feedback
dependencies: []
tags: [docs, cli, cleanup]
created_at: 2026-08-03
---

# Reconcile stale CLI README with the actual command CLI

## Objective

The CLI `README.md` describes the product as an **"Interactive TUI built with Bubble Tea"**
using **Glamour** and **goldmark** — dependencies that are not in `go.mod`. The product
pivoted from a TUI to a non-interactive command CLI (cobra) and the docs were never
reconciled. This is a "code rewritten, docs not updated" tell and it misleads new
contributors on the first file they read.

## Tasks

- [ ] Rewrite the CLI `README.md` intro to describe the actual cobra-based command CLI
- [ ] Remove references to Bubble Tea, Glamour, goldmark, and any "interactive TUI" framing
- [ ] Verify every dependency mentioned in the README actually appears in `go.mod`
- [ ] Ensure command examples in the README match real command names and flags
- [ ] Cross-check against `apps/docs` so the README and docs site agree

## Acceptance Criteria

- The CLI README accurately describes a command CLI, not a TUI
- No dependency is named in the README that is absent from `go.mod`
- All command/flag examples in the README run against the current binary
