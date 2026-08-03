#!/bin/sh
#
# Warn about Go source files that exceed a maximum line count.
#
# This is a soft guardrail toward the project's "small, focused files"
# philosophy (see CLAUDE.md). It is intentionally NON-BLOCKING: it prints
# warnings but always exits 0, so it never fails a commit or CI run.
#
# Test files (*_test.go) are excluded, mirroring the funlen/gocyclo exemption
# for tests in apps/cli/.golangci.yml -- table-driven tests legitimately grow
# large.
#
# Usage: scripts/check-go-file-length.sh [max_lines]
#   max_lines defaults to 300.

set -eu

MAX_LINES="${1:-300}"

# Collect tracked, non-test .go files. Fall back to `find` outside a git repo.
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    files=$(git ls-files '*.go' | grep -v '_test\.go$' || true)
else
    files=$(find . -name '*.go' ! -name '*_test.go' -type f | sed 's|^\./||')
fi

count=0
for f in $files; do
    [ -f "$f" ] || continue
    n=$(wc -l < "$f" | tr -d ' ')
    if [ "$n" -gt "$MAX_LINES" ]; then
        if [ "$count" -eq 0 ]; then
            echo "WARN: Go files over ${MAX_LINES} lines (consider splitting into focused files):"
        fi
        printf '  %5s  %s\n' "$n" "$f"
        count=$((count + 1))
    fi
done

if [ "$count" -gt 0 ]; then
    echo "WARN: ${count} file(s) exceed ${MAX_LINES} lines. This is a warning only; commit not blocked."
fi

exit 0
