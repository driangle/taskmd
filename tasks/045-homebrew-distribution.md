---
id: "045"
title: "Publish taskmd via Homebrew"
status: in-progress
priority: critical
effort: medium
dependencies: ["052"]
tags:
  - distribution
  - homebrew
  - packaging
  - infrastructure
  - mvp
created: 2026-02-08
---

# Publish taskmd via Homebrew

## Objective

Enable users to install taskmd via Homebrew on macOS and Linux, providing a standard installation method for the CLI tool alongside the existing GitHub releases.

## Context

Currently, taskmd is distributed via GitHub releases with pre-built binaries for multiple platforms. While this works, Homebrew is a preferred installation method for many macOS and Linux users as it handles:

- Installation and PATH setup automatically
- Version management and upgrades
- Dependency management
- Uninstallation

Homebrew supports two distribution methods:
1. **Homebrew Core** - Central repository (requires approval, strict guidelines)
2. **Homebrew Tap** - Personal/organization repository (more flexible, faster iteration)

For this task, we'll start with a Homebrew tap to maintain control and iterate quickly.

## Tasks

- [ ] Create a Homebrew tap repository (`homebrew-tap` or similar)
- [ ] Create Homebrew formula (`taskmd.rb`)
  - Define download URLs for binaries
  - Include SHA256 checksums
  - Set up installation paths and symlinks
  - Add test blocks to verify installation
- [ ] Update release workflow to generate formula automatically
  - Extract version, commit, and checksums
  - Generate/update formula file
  - Commit formula to tap repository
- [ ] Test formula installation locally
  - Test on macOS (Intel and Apple Silicon)
  - Test on Linux (AMD64)
  - Verify `taskmd --version` works
  - Verify `taskmd` commands function correctly
- [ ] Document Homebrew installation in README
  - Add installation instructions
  - Include tap command: `brew tap <org>/tap`
  - Include install command: `brew install taskmd`
  - Document upgrade process: `brew upgrade taskmd`
- [ ] Test the full release cycle
  - Create a test release tag
  - Verify formula is auto-generated
  - Test installation from tap
- [ ] (Optional) Add formula audit checks
  - Ensure formula follows Homebrew style guidelines
  - Run `brew audit --strict taskmd`

## Acceptance Criteria

- Homebrew tap repository is created and configured
- Formula file (`taskmd.rb`) is complete and functional
- Users can install with `brew tap <org>/tap && brew install taskmd`
- Installation includes the full binary with embedded web UI
- Formula is automatically updated on each release
- Documentation includes clear Homebrew installation instructions
- Formula passes `brew audit` checks (no critical warnings)
- Tested successfully on macOS and Linux

## Implementation Notes

### Homebrew Formula Structure

Basic structure for `taskmd.rb`:

```ruby
class Taskmd < Formula
  desc "Markdown-based task management CLI and web dashboard"
  homepage "https://github.com/<org>/md-task-tracker"
  version "1.0.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/<org>/md-task-tracker/releases/download/v1.0.0/taskmd-v1.0.0-darwin-arm64.tar.gz"
    sha256 "<checksum>"
  elsif OS.mac?
    url "https://github.com/<org>/md-task-tracker/releases/download/v1.0.0/taskmd-v1.0.0-darwin-amd64.tar.gz"
    sha256 "<checksum>"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/<org>/md-task-tracker/releases/download/v1.0.0/taskmd-v1.0.0-linux-arm64.tar.gz"
    sha256 "<checksum>"
  else
    url "https://github.com/<org>/md-task-tracker/releases/download/v1.0.0/taskmd-v1.0.0-linux-amd64.tar.gz"
    sha256 "<checksum>"
  end

  def install
    bin.install "taskmd"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/taskmd --version")
  end
end
```

### Auto-generation in Release Workflow

Add a step to `.github/workflows/release.yml` to:
1. Download `checksums.txt`
2. Parse SHA256 values for each platform
3. Generate `taskmd.rb` with current version and checksums
4. Commit and push to tap repository

### Tap Repository Setup

1. Create new repo: `homebrew-tap` (or `homebrew-taskmd`)
2. Structure: `Formula/taskmd.rb`
3. Enable GitHub Actions for auto-updates
4. Set up appropriate permissions for formula updates

## Testing

Local testing steps:

```bash
# Add tap locally
brew tap <org>/tap

# Install
brew install taskmd

# Verify installation
which taskmd
taskmd --version
taskmd list --help

# Test web server
taskmd web --port 3000

# Uninstall
brew uninstall taskmd
brew untap <org>/tap
```

## References

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Tap Documentation](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Formula Style Guide](https://docs.brew.sh/Formula-Cookbook#style-guide)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)

## Future Enhancements

After the tap is established and stable:
- Consider submitting to Homebrew Core for wider distribution
- Add Homebrew analytics tracking (opt-in)
- Create cask formula if a GUI version is developed
- Add to additional package managers (apt, yum, Scoop for Windows)
