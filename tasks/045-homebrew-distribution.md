---
id: "045"
title: "Publish taskmd via Homebrew"
status: completed
priority: critical
effort: medium
dependencies: ["052", "053"]
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

- [x] Create a Homebrew tap repository (`homebrew-tap` or similar)
- [x] Create Homebrew formula (`taskmd.rb`)
  - Define download URLs for binaries
  - Include SHA256 checksums
  - Set up installation paths and symlinks
  - Add test blocks to verify installation
- [x] Update release workflow to generate formula automatically
  - Extract version, commit, and checksums
  - Generate/update formula file
  - Commit formula to tap repository
- [x] Test formula installation locally
  - Test on macOS (Intel and Apple Silicon)
  - Test on Linux (AMD64)
  - Verify `taskmd --version` works
  - Verify `taskmd` commands function correctly
- [x] Document Homebrew installation in README
  - Add installation instructions
  - Include tap command: `brew tap driangle/tap`
  - Include install command: `brew install taskmd`
  - Document upgrade process: `brew upgrade taskmd`
- [x] Test the full release cycle
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

## Testing the Release Cycle

### Step 1: Create a Test Release

```bash
# Make sure all changes are committed
git add .
git commit -m "feat: add Homebrew distribution support"

# Create and push a version tag (e.g., v0.1.0)
git tag v0.1.0
git push origin v0.1.0
```

### Step 2: Monitor the Release Workflow

1. Go to https://github.com/driangle/taskmd/actions
2. Watch the "Release" workflow run
3. Verify all steps complete successfully, including "Update Homebrew Formula"
4. Check that the release is created at https://github.com/driangle/taskmd/releases

### Step 3: Verify the Formula Update

1. Go to https://github.com/driangle/homebrew-tap
2. Check that `Formula/taskmd.rb` was updated with the new version
3. Verify the checksums are real (not placeholders)
4. Check the commit message shows the correct version

### Step 4: Test Installation Locally

```bash
# If you already have the tap, update it
brew update

# If not, add the tap
brew tap driangle/tap

# Install taskmd
brew install taskmd

# Verify installation
taskmd --version  # Should show v0.1.0
which taskmd      # Should show /usr/local/bin/taskmd or similar

# Test basic commands
taskmd list tasks/
taskmd stats tasks/
taskmd web start --help

# Test the embedded web UI works
taskmd web start --port 3000
# Open http://localhost:3000 and verify it loads
```

### Step 5: Test Upgrade Path

```bash
# Create another release (e.g., v0.1.1)
git tag v0.1.1
git push origin v0.1.1

# Wait for workflow to complete, then upgrade
brew update
brew upgrade taskmd

# Verify new version
taskmd --version  # Should show v0.1.1
```

### Step 6: Optional - Audit the Formula

```bash
# Check formula follows Homebrew guidelines
brew audit --strict taskmd

# Test installation from scratch
brew uninstall taskmd
brew untap driangle/tap
brew tap driangle/tap
brew install taskmd
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
