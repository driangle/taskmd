#!/usr/bin/env bash
#
# Decide the next sdk/go version and bump the pin — the unattended entry point
# used by CI on main. Wraps bump-sdk-pin.sh, which does the actual work.
#
# Why this exists
# ---------------
# apps/cli and sdk/go are separate modules, and `go install ...@latest` reads
# apps/cli/go.mod rather than go.work. A stale pin on main therefore breaks
# external installs immediately, not at release time (issue #8, PR #9).
#
# Rather than leaving main red until someone bumps the pin by hand, CI heals it:
# if sdk/go changed since its last tag and the pin is behind, this tags the next
# version, repoints the pin, and commits.
#
# Choosing the version
# --------------------
# Defaults to a PATCH bump, which is correct for the common case (additive or
# fix-only SDK changes). Pre-1.0, a BREAKING change must be a minor bump instead
# — a machine cannot detect that, so you flag it in the commit that makes the
# break by including this line anywhere in the message:
#
#     sdk-bump: minor
#
# Any commit touching sdk/go since the last tag carrying that marker promotes the
# whole batch to a minor bump. Get this wrong and the mistake is permanent: module
# versions are immutable once the proxy has fetched them.
#
# Usage:
#   ./scripts/auto-bump-sdk-pin.sh             # bump if needed, else no-op
#   ./scripts/auto-bump-sdk-pin.sh --dry-run   # report what it would do

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

cd "$PROJECT_ROOT"

# Nothing to do unless the pin is actually stale. check-sdk-pin --strict exits 0
# when the pin is current, which is the overwhelmingly common case.
if ./scripts/check-sdk-pin.sh --strict >/dev/null 2>&1; then
    echo "auto-bump-sdk-pin: pin is current, nothing to do"
    exit 0
fi

LAST_TAG="$(git tag -l 'sdk/go/v*' --sort=-v:refname | head -n 1)"

if [[ -z "$LAST_TAG" ]]; then
    NEXT="0.1.0"
    LEVEL="initial"
else
    CURRENT="${LAST_TAG#sdk/go/v}"
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

    # A breaking change cannot be inferred from a diff, so it is opt-in via a
    # marker in the commit message of whatever made the break.
    if git log "$LAST_TAG"..HEAD -- sdk/go | grep -qiE '^[[:space:]]*sdk-bump:[[:space:]]*minor'; then
        NEXT="$MAJOR.$((MINOR + 1)).0"
        LEVEL="minor (sdk-bump: minor marker found)"
    else
        NEXT="$MAJOR.$MINOR.$((PATCH + 1))"
        LEVEL="patch"
    fi
fi

echo "auto-bump-sdk-pin: pin is stale; ${LAST_TAG:-no previous tag} -> sdk/go/v$NEXT ($LEVEL)"

if [[ "$DRY_RUN" == "true" ]]; then
    exec ./scripts/bump-sdk-pin.sh "$NEXT" --dry-run
fi

exec ./scripts/bump-sdk-pin.sh "$NEXT"
