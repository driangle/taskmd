---
name: release
description: Create a new release by bumping versions, tagging, pushing, and generating release notes. Use when the user wants to release a new version.
metadata:
  internal: true
---

# Release

Create a new versioned release of the project. This skill mirrors the process in `scripts/release.sh` — keep them in sync.

## Instructions

The user's input is in `$ARGUMENTS` (a semver version like `1.2.3` or `v1.2.3`, optionally followed by flags).

### Flags

- `--dry-run`: Perform all validation steps but make no changes. Report what would happen.
- `--no-push`: Create the commit and tag locally but do not push to remote.
- `--plugin-taskmd-version VER` / `--plugin-lite-version VER` / `--plugin-mcp-version VER`:
  New version for that plugin's manifest. Required when that plugin's directory changed
  since the last release tag (see step 4 below); ignored otherwise.

### Steps

1. **Parse arguments**: Extract the version from `$ARGUMENTS`. Strip any leading `v` prefix. If no version is provided, ask the user for one.

2. **Validate version format**: Must be valid semver (e.g., `0.1.0`, `1.2.3`, `2.0.0-beta.1`).

3. **Pre-flight validation**:
   - Run `git status --porcelain` — if there are uncommitted changes, stop and tell the user to commit or stash first.
   - Run `git fetch origin` and verify local/remote are in sync.
   - Check the tag doesn't already exist locally or on remote.

4. **Check whether any plugin needs a version bump.** The three marketplace plugins are on
   independent semver lines and are *not* bumped by a repo release — the script fails the
   release if a plugin's directory changed and no version was given for it.

   ```bash
   LAST=$(git describe --tags --abbrev=0 --match 'v[0-9]*')
   for d in claude-code-plugin claude-code-plugin-lite claude-code-plugin-mcp; do
     git diff --quiet "$LAST" HEAD -- "$d" || echo "CHANGED: $d"
   done
   ```

   For each changed plugin, read the diff, decide the bump against the guarantees in
   [`docs/adr/0003-plugin-versioning-policy.md`](../../../docs/adr/0003-plugin-versioning-policy.md),
   and confirm the proposed versions with the user alongside the release notes. Note that
   `taskmd` and `taskmd-lite` are pre-1.0 (a renamed or removed skill is a *minor* bump)
   while `taskmd-mcp` is `1.x` and stable (a changed tool signature is a *major* bump).
   Pass the results as the `--plugin-*-version` flags in steps 7-8.

5. **If `--dry-run`**, stop here and report that validation passed.

6. **Generate release notes** from the commit history since the last tag:

   ```bash
   git log $(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD --pretty=format:"- %s" --no-merges
   ```

   Don't use these raw commit messages as the release notes. Instead, investigate what each commit/task actually did (read task files, check diffs) and write polished, user-facing release notes grouped by category (e.g., New Commands, CLI Improvements, Web Dashboard, Core, Documentation, Removed). Present the release notes to the user before proceeding.

7. **Write the release notes to a file** at `/tmp/taskmd-release-notes-X.Y.Z.md` using the Write tool.

8. **If `--no-push`**, run the script without pushing:

   ```bash
   echo "y" | scripts/release.sh --no-push --notes-file /tmp/taskmd-release-notes-X.Y.Z.md X.Y.Z
   ```

   Report what was created locally and stop.

9. **Run the release script** to handle the full release lifecycle:

   ```bash
   echo "y" | scripts/release.sh --notes-file /tmp/taskmd-release-notes-X.Y.Z.md X.Y.Z
   ```

   Append a `--plugin-*-version` flag for each plugin that step 4 found changed, e.g.
   `--plugin-mcp-version 1.1.0`. Omit them entirely when no plugin changed.

   The script has an interactive confirmation prompt — piping `echo "y"` auto-confirms it. The user already approved the release when they confirmed the release notes, so no second confirmation is needed. The script handles everything: version bumps, commit, tag, push, CI workflow monitoring, and applying release notes after CI creates the release.

10. **Report success** with the release tag and a link to the GitHub releases page.

### Error Handling

- Fail fast on any error. Do not continue if a step fails.
- The release script has built-in rollback: if it fails after modifying version files but before pushing, it will automatically reset the commit and delete the local tag.
- Always provide clear, actionable error messages.
