#!/usr/bin/env bash
#
# Tag sdk/go and repoint apps/cli/go.mod at the new version, in one command.
#
# Why this exists
# ---------------
# apps/cli and sdk/go are separate modules. go.work hides a stale pin locally,
# but external `go install ...@latest` reads apps/cli/go.mod, so drift on main
# ships a CLI that will not compile for anyone outside the repo (issue #8/PR #9).
#
# scripts/release.sh already handles this at release time. This script covers the
# gap *between* releases: you land an SDK change, and until the pin is bumped the
# pre-commit hook blocks your next commit and CI's sdk-pin job fails on main.
#
# The ordering is the fiddly part, and the reason this is a script rather than a
# line in the docs: `go get` resolves versions through the module proxy, so the
# tag has to be pushed before the pin can reference it. Pushing a tag also pushes
# the commit objects it points at, so tag-then-push-then-pin works from a branch
# that has not been pushed yet. Get the order wrong and `go get` fails with an
# unhelpful "unknown revision".
#
# Usage:
#   ./scripts/bump-sdk-pin.sh 0.4.1            # tag sdk/go/v0.4.1, pin, commit
#   ./scripts/bump-sdk-pin.sh 0.4.1 --dry-run  # show what would happen
#
# Pre-1.0 versioning: a breaking API change is a MINOR bump (v0.4.0 -> v0.5.0);
# additive or fix-only changes are a PATCH bump (v0.4.0 -> v0.4.1).

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SDK_MODULE="github.com/driangle/taskmd/sdk/go"

DRY_RUN=false
VERSION=""

log_info()    { echo -e "${BLUE}==>${NC} $1"; }
log_success() { echo -e "${GREEN}✓${NC} $1"; }
log_warning() { echo -e "${YELLOW}!${NC} $1"; }
log_error()   { echo -e "${RED}✗${NC} $1" >&2; }
error_exit()  { log_error "$1"; exit 1; }

usage() {
    cat << EOF
Usage: $(basename "$0") VERSION [--dry-run]

Tag sdk/go as sdk/go/vVERSION, push the tag, repoint apps/cli/go.mod at it,
verify the CLI still builds without the workspace, and commit the pin bump.

ARGUMENTS:
    VERSION     SDK version, with or without a leading 'v' (e.g. 0.4.1)

OPTIONS:
    --dry-run   Print the steps without tagging, pushing, or committing
    -h, --help  Show this help

Pre-1.0: breaking API change = minor bump, additive/fix = patch bump.
Module versions are immutable once pushed — pick the next number, never retag.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        -h|--help) usage; exit 0 ;;
        -*) error_exit "Unknown option: $1" ;;
        *)
            [[ -n "$VERSION" ]] && error_exit "Unexpected argument: $1"
            VERSION="$1"; shift ;;
    esac
done

[[ -z "$VERSION" ]] && { usage; exit 1; }

cd "$PROJECT_ROOT"

CLEAN_VERSION="${VERSION#v}"
if ! [[ "$CLEAN_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
    error_exit "Invalid version: $VERSION (expected semver, e.g. 0.4.1)"
fi
SDK_TAG="sdk/go/v$CLEAN_VERSION"

# --- Preflight -------------------------------------------------------------

# A dirty tree would let the tag point at a commit that does not contain the SDK
# change being released, or sweep unrelated edits into the pin-bump commit.
if [[ -n "$(git status --porcelain)" ]]; then
    error_exit "Working tree is not clean. Commit or stash your changes first."
fi

if git rev-parse --verify --quiet "refs/tags/$SDK_TAG" >/dev/null; then
    error_exit "Tag $SDK_TAG already exists. Module versions are immutable; pick a new version."
fi

LAST_TAG="$(git tag -l 'sdk/go/v*' --sort=-v:refname | head -n 1)"
if [[ -n "$LAST_TAG" ]] && git diff --quiet "$LAST_TAG" HEAD -- sdk/go; then
    log_warning "sdk/go is unchanged since $LAST_TAG — there is nothing to release."
    exit 0
fi

log_info "Releasing sdk/go as $SDK_TAG (previous: ${LAST_TAG:-none})"
git diff --stat "${LAST_TAG:-HEAD}" HEAD -- sdk/go | tail -n 5

if [[ "$DRY_RUN" == "true" ]]; then
    echo
    log_info "[dry-run] would run:"
    echo "    git tag -a $SDK_TAG -m 'sdk/go v$CLEAN_VERSION'"
    echo "    git push origin $SDK_TAG"
    echo "    (cd apps/cli && go get $SDK_MODULE@v$CLEAN_VERSION && go mod tidy)"
    echo "    (cd apps/cli && GOWORK=off go build ./cmd/taskmd)"
    echo "    git commit apps/cli/go.mod apps/cli/go.sum -m 'chore(sdk): pin sdk/go v$CLEAN_VERSION'"
    exit 0
fi

# --- Tag and push ----------------------------------------------------------

# The tag must be pushed before `go get` can resolve it: the module proxy fetches
# from the remote, not from the local clone. Pushing a tag also pushes the commit
# objects it references, so this works even when the branch itself is unpushed.
log_info "Tagging $SDK_TAG"
git tag -a "$SDK_TAG" -m "sdk/go v$CLEAN_VERSION

Release of the Go SDK module ($SDK_MODULE)."
log_success "Created tag $SDK_TAG"

log_info "Pushing $SDK_TAG"
if ! git push origin "$SDK_TAG"; then
    git tag -d "$SDK_TAG" >/dev/null
    error_exit "Failed to push $SDK_TAG (local tag removed, nothing else changed)."
fi
log_success "Pushed tag $SDK_TAG"

# --- Repoint the pin -------------------------------------------------------

log_info "Pinning apps/cli/go.mod to $SDK_MODULE@v$CLEAN_VERSION"
(
    cd apps/cli
    go get "$SDK_MODULE@v$CLEAN_VERSION"
    go mod tidy
)
log_success "Pin updated"

# The pin only matters for consumers outside the workspace, so verify the way
# they build: GOWORK=off resolves sdk/go from the pin, not from the repo.
log_info "Verifying the CLI builds without go.work (simulates go install)"
if ! (cd apps/cli && GOWORK=off go build -o /dev/null ./cmd/taskmd); then
    log_error "CLI does not build with GOWORK=off — external 'go install' would fail."
    log_error "The tag is already pushed and cannot be retagged. Fix the code and"
    log_error "release the next version, e.g. $(basename "$0") <next-version>."
    exit 1
fi
log_success "External build OK"

# --- Commit ----------------------------------------------------------------

if [[ -z "$(git status --porcelain apps/cli/go.mod apps/cli/go.sum)" ]]; then
    log_success "Pin was already up to date, nothing to commit"
    exit 0
fi

git commit -q apps/cli/go.mod apps/cli/go.sum -m "chore(sdk): pin sdk/go v$CLEAN_VERSION

Repoints apps/cli/go.mod at $SDK_TAG so external
'go install $SDK_MODULE/...@latest' resolves an SDK that contains the
symbols the CLI uses. See 'The sdk/go pin' in AGENTS.md."
log_success "Committed the pin bump"

echo
log_success "Done. Push the branch to land it:  git push"
