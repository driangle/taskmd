.PHONY: build run clean install test lint lint-fix

# Build the binary
build:
	go build -o taskmd ./cmd/taskmd

# Run the application
run:
	go run ./cmd/taskmd

# Clean build artifacts
clean:
	rm -f taskmd

# Install the binary to $GOPATH/bin
install:
	go install ./cmd/taskmd

# Run tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run

# Run linter with auto-fix
lint-fix:
	golangci-lint run --fix

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o taskmd-darwin-amd64 ./cmd/taskmd
	GOOS=darwin GOARCH=arm64 go build -o taskmd-darwin-arm64 ./cmd/taskmd
	GOOS=linux GOARCH=amd64 go build -o taskmd-linux-amd64 ./cmd/taskmd
	GOOS=windows GOARCH=amd64 go build -o taskmd-windows-amd64.exe ./cmd/taskmd
