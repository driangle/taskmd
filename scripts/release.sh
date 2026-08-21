#!/usr/bin/env bash

set -Eeuo pipefail  # -E so the ERR trap is inherited by functions

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
DRY_RUN=false
SKIP_CHECKS=false
NO_PUSH=false
VERSION=""
NOTES_FILE=""
SDK_VERSION=""
PLUGIN_TASKMD_VERSION=""
PLUGIN_LITE_VERSION=""
PLUGIN_MCP_VERSION=""

# Help message
usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS] VERSION

Create a new release of taskmd with automated version updates and GitHub release.

ARGUMENTS:
    VERSION     Version number in semver format (e.g., 0.0.1, 1.2.3, 2.0.0-beta.1)
                Can be prefixed with 'v' or not (both v0.0.1 and 0.0.1 are valid)

OPTIONS:
    -h, --help          Show this help message
    -d, --dry-run       Run without making any changes (validation only)
    -n, --no-push       Create tag locally but don't push (for testing)
    --notes-file FILE   Path to a file containing release notes (required for GitHub release)
    --sdk-version VER   Version for the sdk/go module, tagged as sdk/go/vVER.
                        Required when sdk/go has changed since its last tag.
                        Omit when sdk/go is unchanged (no SDK tag is created).
    --plugin-taskmd-version VER
                        New version for the 'taskmd' plugin manifest.
    --plugin-lite-version VER
                        New version for the 'taskmd-lite' plugin manifest.
    --plugin-mcp-version VER
                        New version for the 'taskmd-mcp' plugin manifest.
                        Each is required when that plugin's directory changed
                        since the last release tag, and ignored otherwise.
    --skip-checks       Skip git status and branch checks (use with caution)

EXAMPLES:
    $(basename "$0") 0.0.1 --notes-file notes.md    # Create release with notes
    $(basename "$0") v1.2.3 --notes-file notes.md    # Create release v1.2.3
    $(basename "$0") --dry-run 0.0.2                  # Test release process without changes
    $(basename "$0") --no-push 0.0.1                  # Create tag locally only
    $(basename "$0") 0.3.1 --sdk-version 0.4.1 --notes-file notes.md
                                                      # CLI v0.3.1 + sdk/go/v0.4.1
    $(basename "$0") 0.3.1 --plugin-mcp-version 1.1.0 --notes-file notes.md
                                                      # CLI v0.3.1 + taskmd-mcp 1.1.0

PROCESS:
     1. Validate git repository state (clean working directory)
     2. Validate version format (semantic versioning)
     3. Update version in package.json files
     4. Bump changed plugin manifests (see the plugin version flags)
     5. Tag and push sdk/go (only when it changed; see --sdk-version)
     6. Point apps/cli/go.mod at the SDK version and verify the external build
     7. Commit version changes
     8. Create annotated git tag
     9. Push changes and tag to GitHub
    10. Monitor GitHub Actions release workflow
    11. Apply release notes to the GitHub release
    12. Report success with release URL

VERSIONING:
    apps/cli and sdk/go are separate Go modules and version independently.
    The CLI is tagged vX.Y.Z; the SDK is tagged sdk/go/vX.Y.Z and only when
    its code changed. apps/cli/go.mod pins a released SDK version, so external
    `go install` keeps working. See "The sdk/go pin" in AGENTS.md.

    The three marketplace plugins version independently too, on their own
    semver lines, with no tags of their own. A plugin is bumped only when its
    directory changed; the repo version is never propagated into a plugin
    manifest. If a plugin changed and no version flag was given, this script
    fails rather than shipping a modified plugin under its old version.
    See docs/adr/0003-plugin-versioning-policy.md.

REQUIREMENTS:
    - git (with GitHub remote configured)
    - gh CLI (GitHub CLI) - for monitoring workflow
    - jq (for JSON processing)

EOF
    exit 0
}

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

log_step() {
    echo -e "\n${BLUE}▶${NC} ${BLUE}$1${NC}"
}

# Error handler
error_exit() {
    log_error "$1"
    exit 1
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -n|--no-push)
                NO_PUSH=true
                shift
                ;;
            --notes-file)
                if [[ -z "${2:-}" ]]; then
                    error_exit "--notes-file requires a file path argument"
                fi
                NOTES_FILE="$2"
                shift 2
                ;;
            --sdk-version)
                if [[ -z "${2:-}" ]]; then
                    error_exit "--sdk-version requires a version argument"
                fi
                SDK_VERSION="$2"
                shift 2
                ;;
            --plugin-taskmd-version)
                if [[ -z "${2:-}" ]]; then
                    error_exit "--plugin-taskmd-version requires a version argument"
                fi
                PLUGIN_TASKMD_VERSION="$2"
                shift 2
                ;;
            --plugin-lite-version)
                if [[ -z "${2:-}" ]]; then
                    error_exit "--plugin-lite-version requires a version argument"
                fi
                PLUGIN_LITE_VERSION="$2"
                shift 2
                ;;
            --plugin-mcp-version)
                if [[ -z "${2:-}" ]]; then
                    error_exit "--plugin-mcp-version requires a version argument"
                fi
                PLUGIN_MCP_VERSION="$2"
                shift 2
                ;;
            --skip-checks)
                SKIP_CHECKS=true
                shift
                ;;
            -*)
                error_exit "Unknown option: $1. Use --help for usage information."
                ;;
            *)
                if [[ -z "$VERSION" ]]; then
                    VERSION="$1"
                else
                    error_exit "Multiple version arguments provided. Use --help for usage information."
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$VERSION" ]]; then
        error_exit "Version argument is required. Use --help for usage information."
    fi
}

# Validate version format
validate_version() {
    local version="$1"

    # Remove 'v' prefix if present
    version="${version#v}"

    # Semantic versioning regex
    local semver_regex='^([0-9]+)\.([0-9]+)\.([0-9]+)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$'

    if [[ ! "$version" =~ $semver_regex ]]; then
        error_exit "Invalid version format: $version. Must follow semantic versioning (e.g., 1.2.3, 1.0.0-beta.1)"
    fi

    log_success "Version format valid: $version" >&2
    echo "$version"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites"

    # Check git
    if ! command -v git &> /dev/null; then
        error_exit "git is not installed"
    fi
    log_success "git is installed"

    # Check gh CLI
    if ! command -v gh &> /dev/null; then
        log_warning "gh CLI is not installed - workflow monitoring will be skipped"
        log_info "Install with: brew install gh (macOS) or see https://cli.github.com/"
    else
        log_success "gh CLI is installed"
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        log_warning "jq is not installed - JSON processing may be limited"
        log_info "Install with: brew install jq (macOS)"
    else
        log_success "jq is installed"
    fi

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error_exit "Not in a git repository"
    fi
    log_success "In git repository"

    # Check if GitHub remote is configured
    if ! git remote get-url origin &> /dev/null; then
        error_exit "No 'origin' remote configured"
    fi
    log_success "GitHub remote configured"
}

# Check git status
check_git_status() {
    if [[ "$SKIP_CHECKS" == "true" ]]; then
        log_warning "Skipping git status checks (--skip-checks enabled)"
        return
    fi

    log_step "Checking git repository status"

    # Check for uncommitted changes
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Working directory has uncommitted changes:"
        git status --short
        error_exit "Commit or stash changes before releasing"
    fi
    log_success "Working directory is clean"

    # Fetch latest from remote
    log_info "Fetching latest from remote..."
    git fetch origin

    # Check if current branch exists on remote
    local current_branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)

    if ! git rev-parse --verify "origin/$current_branch" &> /dev/null; then
        error_exit "Current branch '$current_branch' doesn't exist on remote. Push it first."
    fi
    log_success "Current branch exists on remote"

    # Check if we're behind remote
    local local_commit remote_commit
    local_commit=$(git rev-parse HEAD)
    remote_commit=$(git rev-parse "origin/$current_branch")

    if [[ "$local_commit" != "$remote_commit" ]]; then
        log_error "Local branch is not in sync with remote"

        # Check if we're behind
        if git merge-base --is-ancestor HEAD "origin/$current_branch"; then
            error_exit "Local branch is behind remote. Pull latest changes first."
        fi

        # Check if we're ahead
        if git merge-base --is-ancestor "origin/$current_branch" HEAD; then
            error_exit "Local branch is ahead of remote. Push changes first."
        fi

        error_exit "Local and remote branches have diverged. Resolve conflicts first."
    fi
    log_success "Local branch is in sync with remote"
}

# Check if tag already exists
check_tag_exists() {
    local tag="$1"

    log_step "Checking if tag exists"

    if git rev-parse "$tag" &> /dev/null; then
        error_exit "Tag $tag already exists locally"
    fi

    if git ls-remote --tags origin | grep -q "refs/tags/$tag"; then
        error_exit "Tag $tag already exists on remote"
    fi

    log_success "Tag $tag does not exist"
}

# Update version references across the project
update_versions() {
    local version="$1"

    log_step "Updating version references"

    # Update Go version constant in root.go
    local root_go="$PROJECT_ROOT/apps/cli/internal/cli/root.go"
    if [[ -f "$root_go" ]]; then
        sed -i '' "s/Version   = \"[^\"]*\"/Version   = \"$version\"/" "$root_go"
        log_success "Updated $root_go"
    fi

    # Update root package.json
    local root_pkg="$PROJECT_ROOT/package.json"
    if [[ -f "$root_pkg" ]]; then
        if command -v jq &> /dev/null; then
            local tmp_file
            tmp_file=$(mktemp)
            jq --arg ver "$version" '.version = $ver' "$root_pkg" > "$tmp_file"
            mv "$tmp_file" "$root_pkg"
            log_success "Updated $root_pkg"
        else
            log_warning "jq not installed, skipping $root_pkg"
        fi
    fi

    # Update apps/web/package.json
    local web_pkg="$PROJECT_ROOT/apps/web/package.json"
    if [[ -f "$web_pkg" ]]; then
        if command -v jq &> /dev/null; then
            local tmp_file
            tmp_file=$(mktemp)
            jq --arg ver "$version" '.version = $ver' "$web_pkg" > "$tmp_file"
            mv "$tmp_file" "$web_pkg"
            log_success "Updated $web_pkg"
        else
            log_warning "jq not installed, skipping $web_pkg"
        fi
    fi

    # Update apps/vscode/package.json
    local vscode_pkg="$PROJECT_ROOT/apps/vscode/package.json"
    if [[ -f "$vscode_pkg" ]]; then
        if command -v jq &> /dev/null; then
            local tmp_file
            tmp_file=$(mktemp)
            jq --arg ver "$version" '.version = $ver' "$vscode_pkg" > "$tmp_file"
            mv "$tmp_file" "$vscode_pkg"
            log_success "Updated $vscode_pkg"
        else
            log_warning "jq not installed, skipping $vscode_pkg"
        fi
    fi

    # Plugin manifests are deliberately NOT updated here: the three marketplace
    # plugins version independently of the repo. See release_plugins().
}

# --- Plugin versioning -------------------------------------------------------
#
# The three marketplace plugins (taskmd, taskmd-lite, taskmd-mcp) are on their
# own semver lines and are NOT bumped by a repo release. A plugin moves only
# when its own directory moved, so its version number keeps its meaning for
# whoever installed it. See docs/adr/0003-plugin-versioning-policy.md.
#
# Plugins have no tags of their own, so "released state" is the last repo tag:
# that is the commit the marketplace served at the previous release.

# Return the most recent repo release tag (vX.Y.Z), skipping module tags like
# sdk/go/vX.Y.Z, or empty if the repo was never released.
latest_release_tag() {
    git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true
}

# Read the current version out of a plugin manifest.
plugin_manifest_version() {
    jq -r '.version // ""' "$1"
}

# Fail unless $2 is a strictly higher semver than $1.
#
# Plugin versions are what consumers have cached, so a repeated or lowered
# version silently ships changed code under a version someone already has.
require_version_increase() {
    local current="$1" next="$2" plugin="$3"

    if [[ "$current" == "$next" ]]; then
        error_exit "$plugin is already at $current. Pick a higher version — a plugin's changes cannot ship under a version consumers already have."
    fi

    local lowest
    lowest=$(printf '%s\n%s\n' "$current" "$next" | sort -V | head -n 1)
    if [[ "$lowest" != "$current" ]]; then
        error_exit "$plugin version $next is lower than the current $current. Plugin versions only move forward."
    fi
}

# Report whether a plugin's directory changed since the last release.
plugin_changed_since_release() {
    local dir="$1"
    local last_tag
    last_tag=$(latest_release_tag)

    if [[ -z "$last_tag" ]]; then
        return 0   # never released: treat as changed
    fi

    ! git diff --quiet "$last_tag" HEAD -- "$dir"
}

# Validate one plugin's version situation without touching anything.
#
# Mirrors release_sdk_module(): unchanged means nothing to do, changed without a
# version means stop and ask. Runs as a pre-flight check so --dry-run catches a
# missing bump before anything is written.
#
# Returns 1 for a missing bump rather than exiting, so one run can report every
# plugin that needs a version instead of failing on them one at a time.
check_plugin_version() {
    local plugin="$1" dir="$2" flag="$3" requested="$4"

    local manifest="$PROJECT_ROOT/$dir/.claude-plugin/plugin.json"
    if [[ ! -f "$manifest" ]]; then
        log_warning "$plugin: no manifest at $manifest, skipping"
        return
    fi

    local last_tag
    last_tag=$(latest_release_tag)

    if ! plugin_changed_since_release "$dir"; then
        log_success "$plugin unchanged since $last_tag, no bump needed"
        if [[ -n "$requested" ]]; then
            log_warning "$flag $requested given but $dir has not changed; it will be ignored"
        fi
        return
    fi

    if [[ -z "$requested" ]]; then
        log_error "$plugin has changed since ${last_tag:-its last release}:"
        git diff --stat "${last_tag:-HEAD}" HEAD -- "$dir" | tail -n 5
        log_error ""
        log_error "Pass $flag X.Y.Z to release it. See docs/adr/0003-plugin-versioning-policy.md"
        log_error "for what a patch, minor, and major bump mean for this plugin."
        return 1
    fi

    if ! command -v jq &> /dev/null; then
        error_exit "jq is required to bump $plugin. Install it with: brew install jq"
    fi

    local next current
    next=$(validate_version "$requested")
    current=$(plugin_manifest_version "$manifest")
    require_version_increase "$current" "$next" "$plugin"

    log_info "$plugin will be bumped: $current -> $next"
}

# Write a plugin's new version into its manifest. Assumes check_plugin_version()
# already accepted it.
apply_plugin_version() {
    local plugin="$1" dir="$2" requested="$3"

    local manifest="$PROJECT_ROOT/$dir/.claude-plugin/plugin.json"

    if [[ ! -f "$manifest" || -z "$requested" ]] || ! plugin_changed_since_release "$dir"; then
        return
    fi

    local next
    next=$(validate_version "$requested")

    local tmp_file
    tmp_file=$(mktemp)
    jq --arg ver "$next" '.version = $ver' "$manifest" > "$tmp_file"
    mv "$tmp_file" "$manifest"
    log_success "Updated $manifest to $next"
}

# Guard the "plugin.json is the single source of truth" rule.
#
# A version mirrored into marketplace.json is a second copy with nothing keeping
# it honest, and duplicated version numbers drift. Keep them out entirely.
check_marketplace_versions() {
    local marketplace="$PROJECT_ROOT/.claude-plugin/marketplace.json"

    if [[ ! -f "$marketplace" ]] || ! command -v jq &> /dev/null; then
        return
    fi

    if jq -e '[.plugins[]? | has("version")] | any' "$marketplace" > /dev/null; then
        log_error "$marketplace has version fields on its plugin entries."
        log_error "Plugin versions live only in <plugin>/.claude-plugin/plugin.json."
        log_error "See docs/adr/0003-plugin-versioning-policy.md"
        exit 1
    fi

    log_success "marketplace.json carries no duplicate versions"
}

# Pre-flight: every changed plugin must have a valid, higher version to ship under.
check_plugin_versions() {
    log_step "Checking plugin versions"

    check_marketplace_versions

    # Report every plugin missing a bump in one pass, so a release that touched
    # all three does not have to be re-run once per plugin.
    local missing=0

    check_plugin_version "taskmd" "claude-code-plugin" \
        "--plugin-taskmd-version" "$PLUGIN_TASKMD_VERSION" || missing=1
    check_plugin_version "taskmd-lite" "claude-code-plugin-lite" \
        "--plugin-lite-version" "$PLUGIN_LITE_VERSION" || missing=1
    check_plugin_version "taskmd-mcp" "claude-code-plugin-mcp" \
        "--plugin-mcp-version" "$PLUGIN_MCP_VERSION" || missing=1

    if [[ "$missing" == "1" ]]; then
        error_exit "One or more changed plugins need a version. See the errors above."
    fi
}

# Write the accepted versions into the manifests of every changed plugin.
update_plugin_versions() {
    log_step "Updating plugin manifests"

    apply_plugin_version "taskmd" "claude-code-plugin" "$PLUGIN_TASKMD_VERSION"
    apply_plugin_version "taskmd-lite" "claude-code-plugin-lite" "$PLUGIN_LITE_VERSION"
    apply_plugin_version "taskmd-mcp" "claude-code-plugin-mcp" "$PLUGIN_MCP_VERSION"
}

# Return the most recent sdk/go tag, or empty if the module was never tagged.
latest_sdk_tag() {
    git tag -l 'sdk/go/v*' --sort=-v:refname | head -n 1
}

# Report whether sdk/go changed since its last tag.
sdk_changed_since_last_tag() {
    local last_tag
    last_tag=$(latest_sdk_tag)

    if [[ -z "$last_tag" ]]; then
        return 0   # never tagged: treat as changed
    fi

    ! git diff --quiet "$last_tag" HEAD -- sdk/go
}

# Tag and push sdk/go when it changed.
#
# apps/cli and sdk/go are separate modules with independent versions. The SDK is
# only tagged when its code moved, so its version numbers keep their semver
# meaning for anyone importing the library.
#
# The tag has to exist on the remote before apps/cli can pin it, which is why
# this runs — and pushes — before the pin bump and the CLI tag.
release_sdk_module() {
    local last_tag
    last_tag=$(latest_sdk_tag)

    if ! sdk_changed_since_last_tag; then
        log_success "sdk/go unchanged since ${last_tag:-(untagged)}, no SDK release needed"
        if [[ -n "$SDK_VERSION" ]]; then
            log_warning "--sdk-version $SDK_VERSION given but sdk/go has not changed; ignoring"
            SDK_VERSION=""
        fi
        return
    fi

    if [[ -z "$SDK_VERSION" ]]; then
        log_error "sdk/go has changed since ${last_tag:-its last release}:"
        git diff --stat "${last_tag:-HEAD}" HEAD -- sdk/go | tail -n 5
        log_error ""
        log_error "Pass --sdk-version X.Y.Z to release it. Pre-1.0, a breaking API"
        log_error "change is a minor bump; additions and fixes are a patch bump."
        exit 1
    fi

    local sdk_clean="${SDK_VERSION#v}"
    local sdk_tag="sdk/go/v$sdk_clean"

    if git rev-parse --verify --quiet "refs/tags/$sdk_tag" >/dev/null; then
        error_exit "Tag $sdk_tag already exists. Module versions are immutable once published; pick a new version."
    fi

    log_step "Tagging sdk/go as $sdk_tag"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[dry-run] would create and push $sdk_tag"
        return
    fi

    git tag -a "$sdk_tag" -m "sdk/go v$sdk_clean

Release of the Go SDK module (github.com/driangle/taskmd/sdk/go)."
    log_success "Created tag $sdk_tag"

    if [[ "$NO_PUSH" == "true" ]]; then
        log_warning "--no-push: $sdk_tag not pushed, so the go.mod pin cannot resolve it yet"
        return
    fi

    git push origin "$sdk_tag"
    log_success "Pushed tag $sdk_tag"
}

# Point apps/cli/go.mod at the released SDK version.
#
# go.work makes in-repo builds use the local sdk/go, hiding a stale pin. External
# `go install` reads go.mod instead, so a stale pin ships a CLI that will not
# compile for anyone outside the repo (issue #8, PR #9).
sync_sdk_pin() {
    log_step "Syncing sdk/go pin in apps/cli/go.mod"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[dry-run] would pin sdk/go to ${SDK_VERSION:-its current version}"
        return
    fi

    if [[ -n "$SDK_VERSION" ]]; then
        local sdk_clean="${SDK_VERSION#v}"
        (
            cd apps/cli
            go get "github.com/driangle/taskmd/sdk/go@v$sdk_clean"
            go mod tidy
        )
        log_success "Pinned sdk/go to v$sdk_clean"
    else
        log_success "sdk/go unchanged, pin left as is"
    fi
}

# Verify the CLI builds the way an external `go install` would: without the
# workspace, resolving sdk/go from the pin instead of the in-repo copy.
verify_external_build() {
    log_step "Verifying CLI builds without go.work (simulates go install)"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[dry-run] skipping external build check"
        return
    fi

    if (cd apps/cli && GOWORK=off go build -o /dev/null ./cmd/taskmd); then
        log_success "External build OK"
    else
        log_error "CLI does not build with GOWORK=off — the sdk/go pin is stale."
        log_error "External 'go install' would fail for this release."
        exit 1
    fi
}

# Commit version changes
commit_version_changes() {
    local version="$1"

    log_step "Committing version changes"

    git add package.json apps/web/package.json apps/vscode/package.json apps/cli/internal/cli/root.go apps/cli/go.mod apps/cli/go.sum 2>/dev/null || true
    git add claude-code-plugin/.claude-plugin/plugin.json claude-code-plugin-lite/.claude-plugin/plugin.json claude-code-plugin-mcp/.claude-plugin/plugin.json 2>/dev/null || true

    if [[ -z $(git diff --cached --name-only) ]]; then
        log_warning "No version changes to commit"
        return
    fi

    local commit_msg="chore: bump version to $version

Prepare for release v$version

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

    git commit -m "$commit_msg"
    log_success "Committed version changes"
}

# Create git tag
create_git_tag() {
    local version="$1"
    local tag="v$version"

    log_step "Creating git tag $tag"

    local tag_msg="Release $tag

This release includes pre-built binaries for:
- Linux (amd64, arm64)
- macOS (amd64/Intel, arm64/Apple Silicon)
- Windows (amd64)

All binaries include the embedded web dashboard.

MCPB bundles (one-click MCP server install) are included for all platforms."

    git tag -a "$tag" -m "$tag_msg"
    log_success "Created tag $tag"
}

# Apply release notes to the GitHub release created by CI
update_release_notes() {
    local version="$1"
    local tag="v$version"
    local notes_file="$2"

    if ! command -v gh &> /dev/null; then
        log_warning "gh CLI not available - cannot update release notes"
        log_info "Update release notes manually at the GitHub releases page"
        return 0
    fi

    log_step "Updating release notes"

    # CI creates the release in a late job step, so it may not exist the instant
    # the workflow reports success.
    local attempt
    for attempt in $(seq 1 12); do
        gh release view "$tag" &> /dev/null && break
        sleep 5
    done

    if gh release edit "$tag" --notes-file "$notes_file"; then
        log_success "Release notes applied to $tag"
        return 0
    fi

    log_error "Failed to apply release notes to $tag"
    log_info "Apply them manually with:"
    log_info "  gh release edit $tag --notes-file $notes_file"
    return 1
}

# Push changes
push_changes() {
    local version="$1"
    local tag="v$version"

    log_step "Pushing changes to GitHub"

    local current_branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)

    # Push branch
    log_info "Pushing branch $current_branch..."
    git push origin "$current_branch"
    log_success "Pushed branch $current_branch"

    # Push tag
    log_info "Pushing tag $tag..."
    git push origin "$tag"
    log_success "Pushed tag $tag"
}

# Monitor GitHub Actions workflow
monitor_workflow() {
    local version="$1"
    local tag="v$version"

    if ! command -v gh &> /dev/null; then
        log_warning "gh CLI not available - cannot monitor workflow"
        log_info "Check workflow status at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
        log_info "Release artifacts will be available once the workflow completes"
        return 0
    fi

    log_step "Monitoring GitHub Actions workflow"

    log_info "Waiting for workflow to start..."

    # Find the run triggered by THIS tag. Taking the newest run unconditionally
    # races with GitHub registering it, and silently latches onto the previous
    # release's run - which, being already complete, reports instant success.
    local sha
    sha=$(git rev-parse "$tag^{commit}")

    local workflow_id=""
    local attempt
    for attempt in $(seq 1 30); do
        workflow_id=$(gh run list --workflow=release.yml --limit=20 \
            --json databaseId,headSha,headBranch \
            --jq "[.[] | select(.headBranch == \"$tag\" or .headSha == \"$sha\")][0].databaseId" \
            2>/dev/null || echo "")
        [[ -n "$workflow_id" && "$workflow_id" != "null" ]] && break
        workflow_id=""
        sleep 5
    done

    if [[ -z "$workflow_id" ]]; then
        log_warning "Timed out waiting for a workflow run for $tag"
        log_info "Check manually at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
        log_info "Release artifacts will be available once the workflow completes"
        return 0
    fi

    log_info "Workflow started (ID: $workflow_id)"
    log_info "Watching workflow progress..."

    # Watch the workflow
    if gh run watch "$workflow_id" --exit-status; then
        log_success "Workflow completed successfully!"
        return 0
    else
        log_error "Workflow failed or was cancelled"
        local workflow_url
        workflow_url=$(gh run view "$workflow_id" --json url --jq '.url' 2>/dev/null || echo "https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions")
        log_error "Check details at: $workflow_url"
        return 1
    fi
}

# Get release URL
get_release_url() {
    local version="$1"
    local tag="v$version"

    if ! command -v gh &> /dev/null; then
        local repo_url
        repo_url=$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')
        echo "https://github.com/$repo_url/releases/tag/$tag"
        return
    fi

    gh release view "$tag" --json url --jq '.url' 2>/dev/null || {
        local repo_url
        repo_url=$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')
        echo "https://github.com/$repo_url/releases/tag/$tag"
    }
}

# Version to roll back to if a release step fails; empty disables rollback
# (set once the tag has been pushed and rolling back is no longer appropriate).
ROLLBACK_VERSION=""

# ERR trap installed around the release steps in main()
on_release_error() {
    local exit_code=$?
    trap - ERR
    log_error "Release step failed (exit $exit_code)"
    if [[ -n "$ROLLBACK_VERSION" ]]; then
        rollback "$ROLLBACK_VERSION"
    fi
    error_exit "Release failed"
}

# Rollback on failure
rollback() {
    local version="$1"
    local tag="v$version"

    log_warning "Release failed. Rolling back changes..."

    # Delete local tag if it exists
    if git rev-parse "$tag" &> /dev/null; then
        git tag -d "$tag" 2>/dev/null || true
        log_info "Deleted local tag $tag"
    fi

    # Try to delete remote tag if it was pushed
    if [[ "$NO_PUSH" == "false" ]] && git ls-remote --tags origin | grep -q "refs/tags/$tag"; then
        log_warning "Tag was pushed to remote. You may want to delete it manually:"
        log_info "  git push origin :refs/tags/$tag"
    fi

    # Reset version commit if it was made
    if git log -1 --pretty=%B | grep -q "chore: bump version to $version"; then
        git reset --soft HEAD~1
        log_info "Reset version commit (changes are staged)"
    fi
}

# Main release process
main() {
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║     TaskMD Release Script              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}\n"

    parse_args "$@"

    # Normalize version (remove 'v' prefix)
    local clean_version
    clean_version=$(validate_version "$VERSION")
    local tag="v$clean_version"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "DRY RUN MODE - No changes will be made"
    fi

    log_info "Release version: $clean_version"
    log_info "Git tag: $tag"
    echo ""

    # Change to project root
    cd "$PROJECT_ROOT"

    # Pre-flight checks
    check_prerequisites
    check_git_status
    check_tag_exists "$tag"
    check_plugin_versions

    if [[ "$DRY_RUN" == "true" ]]; then
        log_success "\nDry run completed successfully!"
        log_info "Run without --dry-run to create the release"
        exit 0
    fi

    # Confirm with user
    echo ""
    read -p "$(echo -e ${YELLOW}Create release $tag? [y/N]:${NC} )" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 0
    fi

    # Validate notes file if pushing (required for GitHub release)
    if [[ "$NO_PUSH" == "false" && -z "$NOTES_FILE" ]]; then
        error_exit "Release notes are required. Provide --notes-file <path> with polished release notes."
    fi

    if [[ -n "$NOTES_FILE" && ! -f "$NOTES_FILE" ]]; then
        error_exit "Notes file not found: $NOTES_FILE"
    fi

    # Perform release.
    #
    # These steps run under a plain ERR trap rather than `{ ... } || ...`:
    # placing a group (or a function call) in a tested context makes bash ignore
    # `set -e` for everything inside it, so a mid-block failure would neither
    # abort nor trip the rollback - it would just carry on and tag anyway.
    local workflow_failed=false
    local notes_failed=false

    ROLLBACK_VERSION="$clean_version"
    trap on_release_error ERR

    update_versions "$clean_version"
    update_plugin_versions
    release_sdk_module
    sync_sdk_pin
    verify_external_build
    commit_version_changes "$clean_version"
    create_git_tag "$clean_version"

    if [[ "$NO_PUSH" == "false" ]]; then
        push_changes "$clean_version"

        # The tag is public from here on, so rolling back is no longer the right
        # move - report failures instead.
        trap - ERR
        ROLLBACK_VERSION=""

        # Monitor workflow
        if ! monitor_workflow "$clean_version"; then
            workflow_failed=true
        fi

        # Apply release notes to the CI-created release
        if [[ "$workflow_failed" == "false" ]]; then
            if ! update_release_notes "$clean_version" "$NOTES_FILE"; then
                notes_failed=true
            fi
        fi
    else
        trap - ERR
        ROLLBACK_VERSION=""
        log_warning "Skipping push (--no-push enabled)"
        log_info "To push manually: git push origin $(git rev-parse --abbrev-ref HEAD) && git push origin $tag"
    fi

    # Report results
    echo ""
    if [[ "$notes_failed" == "true" ]]; then
        echo -e "${YELLOW}╔════════════════════════════════════════╗${NC}"
        echo -e "${YELLOW}║  Released, but Notes Not Applied!     ║${NC}"
        echo -e "${YELLOW}╚════════════════════════════════════════╝${NC}\n"

        log_warning "Version: $clean_version"
        log_warning "Tag: $tag (pushed, workflow succeeded)"
        log_error "Release notes were NOT applied - the release body is empty"
        log_info "Apply them with:"
        log_info "  gh release edit $tag --notes-file $NOTES_FILE"

        exit 1
    elif [[ "$workflow_failed" == "true" ]]; then
        echo -e "${YELLOW}╔════════════════════════════════════════╗${NC}"
        echo -e "${YELLOW}║   Tag Pushed but Workflow Failed!     ║${NC}"
        echo -e "${YELLOW}╚════════════════════════════════════════╝${NC}\n"

        log_warning "Version: $clean_version"
        log_warning "Tag: $tag (pushed successfully)"
        log_error "GitHub Actions workflow failed"

        local repo_url
        repo_url=$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')
        log_info "Check workflow at: https://github.com/$repo_url/actions"
        log_info "Fix the issue and re-run the workflow, or delete the tag to retry:"
        log_info "  git tag -d $tag"
        log_info "  git push origin :refs/tags/$tag"

        exit 1
    else
        echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║     Release Created Successfully!     ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════╝${NC}\n"

        log_success "Version: $clean_version"
        log_success "Tag: $tag"

        if [[ "$NO_PUSH" == "false" ]]; then
            local release_url
            release_url=$(get_release_url "$clean_version")
            log_success "Release URL: $release_url"

            echo ""
            log_info "Release artifacts available at:"
            log_info "  • taskmd-$tag-linux-amd64.tar.gz"
            log_info "  • taskmd-$tag-linux-arm64.tar.gz"
            log_info "  • taskmd-$tag-darwin-amd64.tar.gz"
            log_info "  • taskmd-$tag-darwin-arm64.tar.gz"
            log_info "  • taskmd-$tag-windows-amd64.zip"
            log_info "  • taskmd-v${clean_version}-darwin-arm64.mcpb"
            log_info "  • taskmd-v${clean_version}-darwin-amd64.mcpb"
            log_info "  • taskmd-v${clean_version}-linux-amd64.mcpb"
            log_info "  • taskmd-v${clean_version}-linux-arm64.mcpb"
            log_info "  • taskmd-v${clean_version}-windows-amd64.mcpb"
            log_info "  • taskmd-v${clean_version}-windows-arm64.mcpb"
            log_info "  • checksums.txt"
        fi

        echo ""
    fi
}

# Run main
main "$@"
