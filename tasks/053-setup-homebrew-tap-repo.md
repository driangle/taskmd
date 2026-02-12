---
id: "053"
title: "Set up homebrew-tap repository"
status: completed
priority: critical
effort: small
dependencies: []
tags:
  - homebrew
  - infrastructure
  - setup
  - mvp
created: 2026-02-12
---

# Set up homebrew-tap repository

## Objective

Create and configure the `homebrew-tap` repository that will host the Homebrew formula for taskmd. This is a prerequisite for enabling Homebrew installation.

## Context

Homebrew taps are Git repositories that contain formula files. The naming convention is `homebrew-<tap-name>`, and users install from it with `brew tap owner/tap-name`.

This task covers the **one-time setup** of the tap repository. Once set up, the main taskmd release workflow will automatically update the formula on each release.

## Tasks

### 1. Create the Repository

- [ ] Create new GitHub repository: `homebrew-tap`
  - Owner: `driangle`
  - Full name: `driangle/homebrew-tap`
  - Public visibility (required for Homebrew)
  - Add description: "Homebrew tap for taskmd - markdown-based task management"
  - Initialize with README

### 2. Set Up Repository Structure

- [ ] Create `Formula/` directory
- [ ] Create initial `Formula/taskmd.rb` file:

```ruby
class Taskmd < Formula
  desc "Markdown-based task management CLI and web dashboard"
  homepage "https://github.com/driangle/taskmd"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/driangle/taskmd/releases/download/v0.1.0/taskmd-v0.1.0-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/driangle/taskmd/releases/download/v0.1.0/taskmd-v0.1.0-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/driangle/taskmd/releases/download/v0.1.0/taskmd-v0.1.0-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/driangle/taskmd/releases/download/v0.1.0/taskmd-v0.1.0-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "taskmd-darwin-arm64" => "taskmd" if OS.mac? && Hardware::CPU.arm?
    bin.install "taskmd-darwin-amd64" => "taskmd" if OS.mac? && Hardware::CPU.intel?
    bin.install "taskmd-linux-arm64" => "taskmd" if OS.linux? && Hardware::CPU.arm?
    bin.install "taskmd-linux-amd64" => "taskmd" if OS.linux? && Hardware::CPU.intel?
  end

  test do
    system "#{bin}/taskmd", "--version"
    assert_match version.to_s, shell_output("#{bin}/taskmd --version")
  end
end
```

- [ ] Create `README.md`:

```markdown
# Homebrew Tap for taskmd

Homebrew formulae for [taskmd](https://github.com/driangle/taskmd) - a markdown-based task management system.

## Installation

```bash
# Add the tap
brew tap driangle/tap

# Install taskmd
brew install taskmd

# Verify installation
taskmd --version
```

## Upgrading

```bash
brew update
brew upgrade taskmd
```

## Uninstalling

```bash
brew uninstall taskmd
brew untap driangle/tap
```

## Formula Updates

The formula in this repository is automatically updated by the taskmd release workflow when a new version is published.

## Issues

For issues with taskmd itself, please report them at: https://github.com/driangle/taskmd/issues

For issues with the Homebrew formula, please report them here.
```

### 3. Configure GitHub Actions Access

The main taskmd repository needs permission to push formula updates to this tap repository.

**Option A: Personal Access Token (Recommended)**

- [ ] Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
- [ ] Generate new token with permissions:
  - `repo` (full control of private repositories) - needed to push to homebrew-tap
- [ ] Copy the token
- [ ] Go to taskmd repository → Settings → Secrets and variables → Actions
- [ ] Add new repository secret:
  - Name: `HOMEBREW_TAP_TOKEN`
  - Value: [paste the token]

**Option B: Deploy Key (Alternative)**

- [ ] Generate SSH key pair: `ssh-keygen -t ed25519 -C "homebrew-tap-deploy"`
- [ ] Add public key to homebrew-tap repository as deploy key (with write access)
- [ ] Add private key to taskmd repository as secret: `HOMEBREW_TAP_DEPLOY_KEY`

### 4. Test the Setup

- [ ] Clone the tap locally: `brew tap driangle/tap`
- [ ] Verify the tap is added: `brew tap`
- [ ] Try installing (will fail if no release yet): `brew install taskmd`
- [ ] Untap for now: `brew untap driangle/tap`

## Acceptance Criteria

- ✅ Repository `driangle/homebrew-tap` exists and is public
- ✅ Repository has `Formula/taskmd.rb` with initial placeholder formula
- ✅ Repository has helpful README with installation instructions
- ✅ GitHub token or deploy key is configured in taskmd repository secrets
- ✅ Local test shows tap can be added successfully

## Next Steps

After completing this task:
1. Update task 045 to `in-progress`
2. The main taskmd release workflow will be updated to:
   - Generate the formula with real version and checksums
   - Push updates to this tap repository
3. Test the full release cycle with a new version tag

## Notes

- The formula will be auto-generated on each release, so manual edits will be overwritten
- The tap repository should remain simple - just Formula/ directory and README
- Once stable, consider submitting to Homebrew core for wider distribution

## References

- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
