---
name: release
description: Create a new release by bumping versions, tagging, pushing, and generating release notes. Use when the user wants to release a new version.
allowed-tools: Bash, Read, Edit, Grep, Glob, Task
---

# Release

Create a new versioned release of the project. This skill mirrors the process in `scripts/release.sh` — keep them in sync.

## Instructions

The user's input is in `$ARGUMENTS` (a semver version like `1.2.3` or `v1.2.3`, optionally followed by flags).

### Flags

- `--dry-run`: Perform all validation steps but make no changes. Report what would happen.
- `--no-push`: Create the commit and tag locally but do not push to remote.

### Steps

1. **Parse arguments**: Extract the version from `$ARGUMENTS`. Strip any leading `v` prefix. If no version is provided, ask the user for one.

2. **Validate version format**: Must be valid semver (e.g., `0.1.0`, `1.2.3`, `2.0.0-beta.1`).

3. **Pre-flight validation**:
   - Run `git status --porcelain` — if there are uncommitted changes, stop and tell the user to commit or stash first.
   - Run `git fetch origin` and verify local/remote are in sync.
   - Check the tag doesn't already exist locally or on remote.

4. **If `--dry-run`**, stop here and report that validation passed.

5. **Generate release notes** from the commit history since the last tag:
   ```bash
   git log $(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD --pretty=format:"- %s" --no-merges
   ```
   Don't use these raw commit messages as the release notes. Instead, investigate what each commit/task actually did (read task files, check diffs) and write polished, user-facing release notes grouped by category (e.g., New Commands, CLI Improvements, Web Dashboard, Core, Documentation, Removed). Present the release notes to the user before proceeding.

6. **Write the release notes to a file** at `/tmp/taskmd-release-notes-X.Y.Z.md` using the Write tool.

7. **If `--no-push`**, run the script without pushing:
    ```bash
    scripts/release.sh --no-push --notes-file /tmp/taskmd-release-notes-X.Y.Z.md X.Y.Z
    ```
    Report what was created locally and stop.

8. **Run the release script** to handle the full release lifecycle:
    ```bash
    scripts/release.sh --notes-file /tmp/taskmd-release-notes-X.Y.Z.md X.Y.Z
    ```
    The script handles everything: version bumps, commit, tag, push, draft release creation, CI workflow monitoring, and re-applying release notes after CI publishes.

9. **Report success** with the release tag and a link to the GitHub releases page.

### Error Handling

- Fail fast on any error. Do not continue if a step fails.
- The release script has built-in rollback: if it fails after modifying version files but before pushing, it will automatically reset the commit and delete the local tag.
- Always provide clear, actionable error messages.
