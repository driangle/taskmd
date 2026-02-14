---
id: "091"
title: "Display taskmd version in web header"
status: in-progress
priority: low
effort: small
tags:
  - mvp
created: 2026-02-14
---

# Display taskmd Version in Web Header

## Objective

Show the current version of the taskmd program in the top-left corner of the web interface, next to the main logo. This gives users a quick way to confirm which version they are running.

## Tasks

- [ ] Expose the taskmd version from the Go backend via an API endpoint or embed it in the initial page data
- [ ] Display the version string next to the logo in the web header/sidebar
- [ ] Style the version label to be subtle and non-intrusive (e.g., smaller font, muted color)
- [ ] Ensure the version updates correctly across builds (uses the same version injected via ldflags)
- [ ] Write tests for the version endpoint or data injection

## Acceptance Criteria

- The current taskmd version is visible in the top-left of the web UI next to the logo
- The version matches the output of `taskmd --version`
- The version label is styled subtly and does not disrupt the header layout
- Works correctly in both light and dark modes
