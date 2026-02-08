---
id: "044"
title: "Add GitHub Actions workflow to build and attach release artifacts"
status: pending
priority: medium
effort: medium
dependencies: []
tags:
  - ci
  - infrastructure
  - cli
created: 2026-02-08
---

# Add GitHub Actions Workflow to Build and Attach Release Artifacts

## Objective

Create a GitHub Actions workflow that automatically builds `taskmd` binaries for multiple platforms and attaches them to the GitHub release when a new tag is pushed.

## Context

Currently there is no automated release process. Users must build from source to get the CLI. A release workflow triggered by tag pushes will produce pre-built binaries that get attached to the corresponding GitHub release, making installation straightforward.

## Tasks

- [ ] Create `.github/workflows/release.yml`
- [ ] Trigger on tag push (`v*` pattern)
- [ ] Build `taskmd` binaries for target platforms:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- [ ] Use `embed_web` build tag for full builds (includes web dashboard)
- [ ] Build the web frontend (`pnpm install && pnpm build`) and copy dist to `apps/cli/internal/web/static/dist/` before Go compilation
- [ ] Name artifacts clearly (e.g., `taskmd-linux-amd64`, `taskmd-darwin-arm64.tar.gz`)
- [ ] Compress binaries (tar.gz for Linux/macOS, zip for Windows)
- [ ] Attach compressed artifacts to the GitHub release
- [ ] Include version info via ldflags (`-X main.Version`, `-X main.GitCommit`, `-X main.BuildDate`)
- [ ] Generate checksums file (SHA256) for all artifacts

## Acceptance Criteria

- Workflow only runs on tag pushes matching `v*`
- Binaries are built with `embed_web` tag (web dashboard included)
- All five platform targets produce working binaries
- Artifacts are attached to the GitHub release automatically
- Version, git commit, and build date are embedded in the binary
- SHA256 checksums file is included in the release
- Workflow does not run on regular branch pushes or PRs

## Example Workflow Structure

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
      # Build web frontend
      # Cross-compile Go binaries
      # Create GitHub release with artifacts
```

## References

- `apps/cli/Makefile` - existing build commands
- `apps/cli/cmd/taskmd/main.go` - version ldflags
- `apps/web/` - web frontend build
