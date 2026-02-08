.PHONY: build run clean install install-full test lint lint-fix web-build web-embed build-full

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

# Run tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run

# Run linter with auto-fix
lint-fix:
	golangci-lint run --fix

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

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o taskmd-darwin-amd64 ./cmd/taskmd
	GOOS=darwin GOARCH=arm64 go build -o taskmd-darwin-arm64 ./cmd/taskmd
	GOOS=linux GOARCH=amd64 go build -o taskmd-linux-amd64 ./cmd/taskmd
	GOOS=windows GOARCH=amd64 go build -o taskmd-windows-amd64.exe ./cmd/taskmd
