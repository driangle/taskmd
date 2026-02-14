---
name: release
description: Create a new release by bumping versions, tagging, pushing, and generating release notes. Use when the user wants to release a new version.
allowed-tools: Bash, Read, Edit, Grep, Glob
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

3. **Validate the working directory is clean**:
   - Run `git status --porcelain` — if there are uncommitted changes, stop and tell the user to commit or stash first.
   - Run `git fetch origin`.
   - Verify the current branch exists on the remote.
   - Verify local and remote are in sync (not ahead, behind, or diverged).

4. **Check the tag doesn't already exist** (locally and on remote):
   ```bash
   git tag -l "v<version>"
   git ls-remote --tags origin "refs/tags/v<version>"
   ```

5. **If `--dry-run`**, stop here and report that validation passed.

6. **Update all version references** in the project (3 files):
   - `apps/cli/internal/cli/root.go` — update the `Version` variable (e.g., `Version   = "X.Y.Z"`)
   - `package.json` — update the `"version"` field
   - `apps/web/package.json` — update the `"version"` field
   Use the `Edit` tool to make these changes.

7. **Commit the version changes** with the standardized message:
   ```
   chore: bump version to X.Y.Z

   Prepare for release vX.Y.Z

   Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
   ```
   Stage only the 3 modified files — do not use `git add -A`.

8. **Create an annotated git tag** with the standard message:
   ```
   Release vX.Y.Z

   This release includes pre-built binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64/Intel, arm64/Apple Silicon)
   - Windows (amd64)

   All binaries include the embedded web dashboard.
   ```

9. **Generate release notes** from the commit history since the last tag:
   ```bash
   git log $(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD --pretty=format:"- %s" --no-merges
   ```
   Present the release notes to the user before pushing.

10. **If `--no-push`**, stop here and report what was created locally.

11. **Push the release commit and tag** (ask the user for confirmation first):
    ```bash
    git push origin <current-branch>
    git push origin "vX.Y.Z"
    ```

12. **Report success** with the release tag and a link to the GitHub releases page. Note that the GitHub Actions release workflow will build and publish the binaries automatically.

### Error Handling

- Fail fast on any error. Do not continue if a step fails.
- If a step fails after version files were modified but before pushing, offer to roll back:
  - Delete the local tag: `git tag -d vX.Y.Z`
  - Reset the version commit: `git reset --soft HEAD~1`
- Always provide clear, actionable error messages.
