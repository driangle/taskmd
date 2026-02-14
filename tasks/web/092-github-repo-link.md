---
id: "092"
title: "Add GitHub repository link to web interface"
status: completed
priority: low
effort: small
tags:
  - mvp
created: 2026-02-14
---

# Add GitHub Repository Link to Web Interface

## Objective

Add a link to the taskmd GitHub repository somewhere accessible in the web interface (e.g., in the sidebar footer, header, or an about/help section). This gives users easy access to documentation, issues, and source code.

## Tasks

- [x] Decide on placement for the GitHub link (sidebar footer, header icon, etc.)
- [x] Add a GitHub icon/link component pointing to the repository URL
- [x] Make the repository URL configurable or derive it from build-time config
- [x] Style the link to fit naturally within the existing UI
- [x] Ensure the link opens in a new tab
- [x] Verify appearance in both light and dark modes

## Acceptance Criteria

- A GitHub link is visible and accessible in the web interface
- Clicking the link opens the taskmd GitHub repository in a new tab
- The link placement feels natural and non-intrusive
- Works correctly in both light and dark modes
