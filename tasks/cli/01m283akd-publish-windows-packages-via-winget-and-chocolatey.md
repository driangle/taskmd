---
id: "01m283akd"
title: "Publish Windows packages via winget and Chocolatey"
status: pending
priority: medium
effort: medium
phase: windows-support
dependencies: []
tags: ["distribution", "windows", "packaging"]
created_at: 2026-09-11
---

# Publish Windows packages via winget and Chocolatey

## Objective

Give Windows users a package-manager install path, so they get the release
binary (with the embedded web dashboard) instead of falling back to
`go install`, which ships the CLI only.

Requested in [issue #18](https://github.com/driangle/taskmd/issues/18): the
reporter installed with Go on Windows and hit "No web UI embedded in this
build" — a winget or Chocolatey package would have steered them to the full
binary.

Release builds already produce `taskmd-windows-amd64.exe` and
`taskmd-windows-arm64.exe` with `-tags embed_web`
(`.github/workflows/release.yml`), so this is a distribution task only — no
build changes needed.

## Tasks

- [ ] Evaluate winget vs Chocolatey ordering — winget is manifest-based
      against release URLs (no hosted package, lower maintenance) and is a
      good first target; Chocolatey requires a maintained package and
      moderation review
- [ ] Submit a winget manifest to `microsoft/winget-pkgs` for the amd64 and
      arm64 release binaries
- [ ] Automate the manifest update in `release.yml` (e.g. `wingetcreate` or
      an action that PRs the new version on each release)
- [ ] Decide whether Chocolatey is worth the ongoing maintenance; if yes,
      publish and automate it, if no, record the decision here
- [ ] Document the new install path(s) in `README.md`,
      `apps/docs/getting-started/installation.md`, and `apps/docs/faq.md`
- [ ] Comment on and close issue #18's packaging ask once an option ships

## Acceptance Criteria

- `winget install taskmd` (or the chosen package id) installs a binary where
  `taskmd web start` serves the embedded dashboard
- New releases update the package without manual steps (or the manual step is
  documented in the release checklist)
- Installation docs list the Windows package manager option
