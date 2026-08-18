---
id: "003"
title: Patch XSS vulnerability in comments
status: pending
priority: critical
effort: small
type: bug
tags: [security, urgent]
created: 2026-03-03
---

# Patch XSS vulnerability in comments

## Objective

Fix a reflected XSS vulnerability in the comments section. User input is not properly sanitized before rendering in the DOM.

## Tasks

- [ ] Sanitize all user input in comment rendering
- [ ] Add CSP headers to prevent inline script execution
- [ ] Write regression tests for XSS vectors

## Acceptance Criteria

- No XSS payloads execute when submitted as comments
- CSP headers are set on all responses
- Regression tests cover OWASP XSS cheat sheet vectors
