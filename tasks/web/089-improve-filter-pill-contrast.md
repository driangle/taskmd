---
id: "089"
title: "Improve filter pill contrast for active vs inactive states"
status: completed
priority: medium
effort: small
tags:
  - mvp
created: 2026-02-14
---

# Improve Filter Pill Contrast for Active vs Inactive States

## Objective

The gray pill-style filters used across the web UI (task list, board, etc.) have insufficient contrast between their active (selected) and inactive (unselected) states. Update the styling so users can clearly distinguish which filters are currently applied at a glance.

## Tasks

- [ ] Audit all filter pill components across web pages to identify shared styles
- [ ] Update active/selected pill styling with higher contrast (e.g., stronger background color, bolder text, or distinct border)
- [ ] Ensure inactive/unselected pills remain visually subdued but still readable
- [ ] Verify contrast meets WCAG AA accessibility guidelines
- [ ] Test across both light and dark modes
- [ ] Confirm consistent styling on all pages that use filter pills

## Acceptance Criteria

- Active filter pills are clearly distinguishable from inactive ones without squinting
- The contrast difference is obvious and immediate
- Styling is consistent across all pages that use filter pills
- Works correctly in both light and dark modes
- No regression in existing filter functionality
