#!/bin/sh
#
# Check that apps/cli/go.mod's sdk/go pin is not behind the in-repo SDK.
#
# Why this exists
# ---------------
# apps/cli and sdk/go are separate Go modules. go.work makes local builds and CI
# use the in-repo sdk/go, which hides a stale pin in apps/cli/go.mod. External
# consumers have no workspace:
#
#     go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
#
# resolves the pin from go.mod instead. If the CLI uses SDK symbols added after
# the pinned commit, that install fails to compile. This has happened before —
# see issue #8 and PR #9, which fixed the pin by hand but never added a guard.
#
# There are no `apps/cli/vX.Y.Z` tags, so `@latest` resolves to main HEAD. Drift
# therefore breaks users as soon as it lands on main, not at release time.
#
# What it checks
# --------------
# Whether sdk/go has changed since the commit (or tag) apps/cli/go.mod pins.
# This is a pure git comparison: no network, and it does not require the SDK
# change to have been pushed.
#
# Modes
# -----
#   --staged  (pre-commit) Tolerates the commit that *introduces* SDK changes,
#             because the pin cannot reference a commit that is not pushed yet.
#             Blocks any later commit that leaves the pin stale, so the bump
#             cannot be forgotten — which is exactly how #8 slipped through.
#
#   --strict  (CI on main and release tags) Any drift is an error. At that point
#             the SDK commit exists and is pushed, so the pin can and must
#             reference it.
#
# Exit codes: 0 ok (or skipped), 1 stale pin.

set -e

MODE="${1:---staged}"
REPO_ROOT="$(git rev-parse --show-toplevel)"
GO_MOD="$REPO_ROOT/apps/cli/go.mod"
SDK_MODULE="github.com/driangle/taskmd/sdk/go"
SDK_DIR="sdk/go"
CLI_PACKAGE="github.com/driangle/taskmd/apps/cli/cmd/taskmd"

# The bump command shown in every failure message. Releases go through
# scripts/release.sh, which tags sdk/go and repoints the pin as part of the
# release; bump-sdk-pin.sh does the same job between releases.
BUMP_CMD="make bump-sdk-pin VERSION=<version>   # tags sdk/go, pushes it, repoints the pin, commits"

if [ ! -f "$GO_MOD" ]; then
    echo "check-sdk-pin: $GO_MOD not found, skipping"
    exit 0
fi

# Extract the pinned version, e.g. "v0.0.0-20260423161219-b2f7c29aba77" or "v0.4.0".
PINNED="$(awk -v mod="$SDK_MODULE" '$1 == mod { print $2; exit }' "$GO_MOD")"
if [ -z "$PINNED" ]; then
    echo "check-sdk-pin: no $SDK_MODULE requirement in go.mod, skipping"
    exit 0
fi

# Resolve the pin to a git ref. Both pin forms are supported:
#   released version  vX.Y.Z                       -> the sdk/go/vX.Y.Z tag (preferred)
#   pseudo-version    vX.Y.Z-<timestamp>-<12-hex>  -> the commit hash
case "$PINNED" in
    *-*-*)
        REF="${PINNED##*-}"
        PIN_KIND="pseudo-version"
        ;;
    *)
        REF="$SDK_DIR/$PINNED"
        PIN_KIND="released version"
        ;;
esac

if ! git -C "$REPO_ROOT" rev-parse --verify --quiet "${REF}^{commit}" >/dev/null 2>&1; then
    # Shallow clone, or a tag that was never created. Can't judge; don't block.
    echo "check-sdk-pin: pinned $PIN_KIND $PINNED ($REF) not found locally, skipping"
    exit 0
fi

# Has the SDK moved since the pin?
if git -C "$REPO_ROOT" diff --quiet "$REF" HEAD -- "$SDK_DIR"; then
    COMMITTED_DRIFT=0
else
    COMMITTED_DRIFT=1
fi

# Are SDK changes part of what is being committed right now?
if [ "$MODE" = "--staged" ] &&
    git -C "$REPO_ROOT" diff --cached --name-only | grep -q "^$SDK_DIR/"; then
    STAGED_SDK=1
else
    STAGED_SDK=0
fi

if [ "$COMMITTED_DRIFT" -eq 0 ] && [ "$STAGED_SDK" -eq 0 ]; then
    exit 0
fi

if [ "$STAGED_SDK" -eq 1 ]; then
    # The introducing commit. Blocking here would deadlock: the pin has to name
    # a commit that does not exist yet.
    echo "check-sdk-pin: NOTE — this commit changes $SDK_DIR/."
    echo
    echo "  apps/cli/go.mod still pins $PINNED."
    echo
    echo "  Usually you can ignore this: once the commit is on main, CI tags the next"
    echo "  $SDK_DIR version and pushes the pin bump for you. Just 'git pull' afterwards,"
    echo "  or your next commit will be blocked by a pin that is stale only locally."
    echo
    echo "  Breaking API change? Say so in the commit message, or it ships as a patch:"
    echo
    echo "      sdk-bump: minor"
    echo
    echo "  To bump now instead of waiting for CI, from a clean tree:"
    echo
    echo "      $BUMP_CMD"
    exit 0
fi

# Drift is already committed and this change is unrelated: stop here.
echo "check-sdk-pin: FAIL — apps/cli/go.mod pins a stale $SDK_MODULE version."
echo
echo "  pinned:  $PINNED ($REF)"
echo "  but $SDK_DIR/ has changed since that commit:"
echo
git -C "$REPO_ROOT" diff --stat "$REF" HEAD -- "$SDK_DIR" | sed 's/^/      /'
echo
echo "  go.work hides this locally, but external installs read go.mod:"
echo "      go install $CLI_PACKAGE@latest   # would fail to compile"
echo
echo "  If this drift is already on main, CI has probably bumped the pin already —"
echo "  'git pull' may be all you need."
echo
echo "  Otherwise, fix it with one command, from a clean tree:"
echo
echo "      $BUMP_CMD"
echo
echo "  See issue #8 / PR #9 for the previous occurrence."
exit 1
