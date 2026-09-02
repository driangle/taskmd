#!/usr/bin/env bash
#
# Run a skival suite and stamp the run with the commit it measured.
#
#   evals/run-eval.sh <suite-dir> [extra skival args...]
#   evals/run-eval.sh list-tasks
#   evals/run-eval.sh list-tasks --samples 1        # smoke run
#
# Every run lands in its own <suite-dir>/results/<timestamp>/ and carries a
# snapshot.json recording the commit, the tool versions and the wall clock.
#
# Why this exists: a skill eval measures the skill files *as of some commit*. A
# report that does not name that commit cannot be compared against a later one,
# and "did the fix work?" becomes unanswerable. Editing a skill and re-running is
# the whole point of the harness, so the provenance has to be automatic.
#
# results/ is gitignored and regenerable. The durable record is the committed
# report under <suite-dir>/reports/<date>-<commit>.md.

set -euo pipefail

SUITE_DIR="${1:-}"
if [[ -z "$SUITE_DIR" ]]; then
  echo "usage: $0 <suite-dir> [skival args...]" >&2
  exit 2
fi
shift

REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
cd "$REPO_ROOT/evals/$SUITE_DIR"

[[ -f suite.yaml ]] || { echo "no suite.yaml in evals/$SUITE_DIR" >&2; exit 2; }

COMMIT="$(git -C "$REPO_ROOT" rev-parse HEAD)"
SHORT="$(git -C "$REPO_ROOT" rev-parse --short HEAD)"
# Repo-wide, deliberately: the skills under test live outside the suite directory,
# so a check scoped to the suite would call a run "clean" while measuring edited
# skill files. results/ is gitignored and so never shows up here.
DIRTY=false
if [[ -n "$(git -C "$REPO_ROOT" status --porcelain)" ]]; then
  DIRTY=true
  cat >&2 <<WARN

  ⚠  Working tree is dirty. This run will be recorded against $SHORT, but that
     commit is not what is being measured. Commit first for a citable result.

WARN
fi

echo "suite:  evals/$SUITE_DIR"
echo "commit: $SHORT$([[ $DIRTY == true ]] && echo ' (dirty)')"
echo

skival validate suite.yaml >/dev/null || { echo "suite.yaml failed validation" >&2; exit 1; }

STARTED="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
set +e
skival run suite.yaml --results-dir ./results "$@"
STATUS=$?
set -e
FINISHED="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# skival creates its own timestamped directory under --results-dir.
RUN_DIR="$(ls -dt ./results/*/ 2>/dev/null | head -1 || true)"
if [[ -z "$RUN_DIR" ]]; then
  echo "no results directory was produced" >&2
  exit "$STATUS"
fi

cat > "${RUN_DIR}snapshot.json" <<JSON
{
  "suite": "$SUITE_DIR",
  "commit": "$COMMIT",
  "commit_short": "$SHORT",
  "dirty": $DIRTY,
  "started_at": "$STARTED",
  "finished_at": "$FINISHED",
  "skival_args": "$*",
  "taskmd_version": "$(taskmd --version 2>/dev/null | head -1 | tr -d '\n')",
  "taskmd_path": "$(command -v taskmd)",
  "skival_version": "$(skival --version 2>/dev/null | head -1 | tr -d '\n')"
}
JSON

echo
echo "run dir:  ${RUN_DIR}"
echo "snapshot: ${RUN_DIR}snapshot.json (commit $SHORT)"
echo
echo "Commit the write-up as evals/$SUITE_DIR/reports/$(date -u +%Y-%m-%d)-$SHORT.md"
exit "$STATUS"
