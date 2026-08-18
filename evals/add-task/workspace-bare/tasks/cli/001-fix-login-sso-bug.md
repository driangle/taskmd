---
id: "001"
title: Fix login SSO bug
status: in-progress
priority: high
effort: medium
type: bug
tags: [auth, urgent]
created: 2026-03-01
---

# Fix login SSO bug

## Objective

Fix the login flow when users authenticate via SSO. Currently, the callback handler does not properly validate the SAML response, causing intermittent login failures for enterprise users.

## Tasks

- [x] Reproduce the bug with a test SSO provider
- [ ] Add SAML response validation in the callback handler
- [ ] Handle expired SSO sessions gracefully
- [ ] Add error logging for failed SSO attempts

## Acceptance Criteria

- SSO login works reliably with SAML 2.0 providers
- Expired sessions show a user-friendly error message
- Failed SSO attempts are logged with correlation IDs
