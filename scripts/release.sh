#!/usr/bin/env bash

set -euo pipefail

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
    --skip-checks       Skip git status and branch checks (use with caution)

EXAMPLES:
    $(basename "$0") 0.0.1 --notes-file notes.md    # Create release with notes
    $(basename "$0") v1.2.3 --notes-file notes.md    # Create release v1.2.3
    $(basename "$0") --dry-run 0.0.2                  # Test release process without changes
    $(basename "$0") --no-push 0.0.1                  # Create tag locally only

PROCESS:
    1. Validate git repository state (clean working directory)
    2. Validate version format (semantic versioning)
    3. Update version in package.json files
    4. Commit version changes
    5. Create annotated git tag
    6. Push changes and tag to GitHub
    7. Monitor GitHub Actions release workflow
    8. Apply release notes to the GitHub release
    9. Report success with release URL

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

    # Update claude-code-plugin/.claude-plugin/plugin.json
    local plugin_json="$PROJECT_ROOT/claude-code-plugin/.claude-plugin/plugin.json"
    if [[ -f "$plugin_json" ]]; then
        if command -v jq &> /dev/null; then
            local tmp_file
            tmp_file=$(mktemp)
            jq --arg ver "$version" '.version = $ver' "$plugin_json" > "$tmp_file"
            mv "$tmp_file" "$plugin_json"
            log_success "Updated $plugin_json"
        else
            log_warning "jq not installed, skipping $plugin_json"
        fi
    fi
}

# Sync apps/cli's sdk/go pin to the current SDK state.
#
# apps/cli and sdk/go are separate modules. go.work makes in-repo builds use the
# local SDK, which hides a stale pin — but `go install .../cmd/taskmd@...` reads
# apps/cli/go.mod and fails to compile against an old SDK. That shipped once
# already (issue #8, fixed by hand in PR #9); this makes it automatic.
sync_sdk_pin() {
    log_step "Syncing sdk/go pin in apps/cli/go.mod"

    (
        cd apps/cli
        go get github.com/driangle/taskmd/sdk/go@HEAD
        go mod tidy
    )

    if [[ -z $(git status --porcelain apps/cli/go.mod apps/cli/go.sum) ]]; then
        log_success "sdk/go pin already current"
        return
    fi

    log_success "Bumped sdk/go pin"
}

# Verify the CLI builds the way an external `go install` would: without the
# workspace, resolving sdk/go from the pin instead of the in-repo copy.
verify_external_build() {
    log_step "Verifying CLI builds without go.work (simulates go install)"

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

    git add package.json apps/web/package.json apps/vscode/package.json apps/cli/internal/cli/root.go claude-code-plugin/.claude-plugin/plugin.json apps/cli/go.mod apps/cli/go.sum 2>/dev/null || true

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

    gh release edit "$tag" --notes-file "$notes_file"
    log_success "Release notes applied to $tag"
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
    sleep 5

    # Get the latest workflow run for the release workflow
    local workflow_id
    workflow_id=$(gh run list --workflow=release.yml --limit=1 --json databaseId --jq '.[0].databaseId' 2>/dev/null || echo "")

    if [[ -z "$workflow_id" ]]; then
        log_warning "Could not find workflow run"
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

    # Perform release
    local release_failed=false
    local workflow_failed=false

    {
        update_versions "$clean_version"
        sync_sdk_pin
        verify_external_build
        commit_version_changes "$clean_version"
        create_git_tag "$clean_version"

        if [[ "$NO_PUSH" == "false" ]]; then
            push_changes "$clean_version"

            # Monitor workflow
            if ! monitor_workflow "$clean_version"; then
                workflow_failed=true
            fi

            # Apply release notes to the CI-created release
            if [[ "$workflow_failed" == "false" ]]; then
                update_release_notes "$clean_version" "$NOTES_FILE"
            fi
        else
            log_warning "Skipping push (--no-push enabled)"
            log_info "To push manually: git push origin $(git rev-parse --abbrev-ref HEAD) && git push origin $tag"
        fi
    } || {
        release_failed=true
    }

    if [[ "$release_failed" == "true" ]]; then
        rollback "$clean_version"
        error_exit "Release failed"
    fi

    # Report results
    echo ""
    if [[ "$workflow_failed" == "true" ]]; then
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
