.PHONY: build run clean install install-full install-dev install-dev-full test lint lint-fix tidy check web-build web-embed build-full man

# Build the binary
build:
	go build -o taskmd ./cmd/taskmd

# Run the application
run:
	go run ./cmd/taskmd

# Clean build artifacts
clean:
	rm -f taskmd
	rm -rf internal/web/static/dist

# Install the binary to $GOPATH/bin
install:
	go install ./cmd/taskmd

# Install with embedded web assets
install-full: web-embed
	go install -tags embed_web ./cmd/taskmd

# ldflags for injecting git info into dev builds
DEV_COMMIT := $(shell git rev-parse --short HEAD)
DEV_DIRTY  := $(shell git diff --quiet 2>/dev/null && echo false || echo true)
DEV_LDFLAGS := -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitCommit=$(DEV_COMMIT)' \
               -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitDirty=$(DEV_DIRTY)'

# Install as taskmd-dev (keeps Homebrew version available)
install-dev:
	go build -ldflags "$(DEV_LDFLAGS)" -o ~/bin/taskmd-dev ./cmd/taskmd
	@echo "Installed development binary to ~/bin/taskmd-dev"
	@echo "Usage: taskmd-dev [command]"

# Install as taskmd-dev with embedded web assets
install-dev-full: web-embed
	go build -tags embed_web -ldflags "$(DEV_LDFLAGS)" -o ~/bin/taskmd-dev ./cmd/taskmd
	@echo "Installed development binary (with web) to ~/bin/taskmd-dev"
	@echo "Usage: taskmd-dev [command]"

# Run tests
test:
	go test ./...

GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo $(shell go env GOPATH)/bin/golangci-lint)

# Run linter
lint:
	$(GOLANGCI_LINT) run

# Run linter with auto-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix

# Run go mod tidy to clean up dependencies
tidy:
	go mod tidy

# Run all checks (test, lint, vet)
check: test lint
	go vet ./...

# Build the web SPA
web-build:
	cd ../../apps/web && pnpm install && pnpm build

# Copy built SPA into Go-embeddable location
web-embed: web-build
	rm -rf internal/web/static/dist
	mkdir -p internal/web/static
	cp -r ../../apps/web/dist internal/web/static/dist

# Build Go binary with embedded web assets
build-full: web-embed
	go build -tags embed_web -o taskmd ./cmd/taskmd

# Generate man pages
man: build
	mkdir -p man
	./taskmd man man/
	@echo "Man pages generated in man/"

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o taskmd-darwin-amd64 ./cmd/taskmd
	GOOS=darwin GOARCH=arm64 go build -o taskmd-darwin-arm64 ./cmd/taskmd
	GOOS=linux GOARCH=amd64 go build -o taskmd-linux-amd64 ./cmd/taskmd
	GOOS=windows GOARCH=amd64 go build -o taskmd-windows-amd64.exe ./cmd/taskmd
