.PHONY: install-dev install-dev-full check check-lite test lint check-file-length check-sdk-pin bump-sdk-pin sync-spec docker-build docker-run

# Install development binary (delegates to apps/cli)
install-dev:
	$(MAKE) -C apps/cli install-dev

# Install development binary with embedded web assets
install-dev-full:
	$(MAKE) -C apps/cli install-dev-full

# Quick check: compile and lint all projects (no tests)
check-lite: check-file-length
	cd apps/cli && go build ./...
	$(MAKE) -C apps/cli lint
	cd sdk/go && go build ./...
	cd apps/web && pnpm run typeCheck
	cd apps/web && pnpm run lint
	cd apps/vscode && pnpm run lint

# Warn about Go files over 300 lines (non-blocking guardrail; excludes tests)
check-file-length:
	./scripts/check-go-file-length.sh 300

# Fail if apps/cli/go.mod's sdk/go pin is behind the in-repo SDK. go.work hides
# this locally, but external `go install` reads go.mod. See issue #8 / PR #9.
check-sdk-pin:
	./scripts/check-sdk-pin.sh --strict

# Tag sdk/go and repoint the pin at it, in one step. Use between releases, when
# an SDK change has landed but the next release is not imminent.
#   make bump-sdk-pin VERSION=0.4.1
bump-sdk-pin:
	@test -n "$(VERSION)" || { echo "usage: make bump-sdk-pin VERSION=0.4.1"; exit 1; }
	./scripts/bump-sdk-pin.sh $(VERSION)

# Run all checks (compile, lint, tests for all projects + docs build)
check: check-lite
	cd apps/cli && go test ./...
	cd sdk/go && go test ./...
	cd apps/web && npx vitest run
	cd apps/vscode && pnpm test
	cd apps/docs && pnpm build

# Run tests only
test:
	$(MAKE) -C apps/cli test
	cd sdk/go && go test ./...

# Run linter only
lint:
	$(MAKE) -C apps/cli lint

# Sync spec copies from docs/taskmd_specification.md
sync-spec:
	$(MAKE) -C apps/cli sync-spec

# Build Docker image
docker-build:
	docker build -t taskmd:local .

# Run Docker container (mount ./tasks as read-only)
docker-run: docker-build
	docker run --rm -p 8080:8080 -v ./tasks:/tasks:ro taskmd:local
