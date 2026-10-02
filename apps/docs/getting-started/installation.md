# Installation

## Homebrew (macOS and Linux)

The recommended way to install taskmd:

```bash
# Add the tap
brew tap driangle/tap

# Install taskmd
brew install taskmd

# Verify installation
taskmd --version
```

## Pre-built Binaries

Download from the [GitHub releases page](https://github.com/driangle/taskmd/releases):

```bash
# Archives contain a single binary named taskmd-<os>-<arch>

# macOS / Linux (example: Apple Silicon)
tar -xzf taskmd-v*-darwin-arm64.tar.gz
sudo install -m755 taskmd-darwin-arm64 /usr/local/bin/taskmd

# Windows: unzip, rename taskmd-windows-<arch>.exe to taskmd.exe,
# and add its folder to your PATH
```

Available platforms:
- **Linux**: AMD64, ARM64
- **macOS**: AMD64 (Intel), ARM64 (Apple Silicon)
- **Windows**: AMD64, ARM64

## Docker

The Docker image includes the full CLI and the embedded web dashboard. No Go or Node.js required.

**Web dashboard** (default):
```bash
docker run --rm -p 8080:8080 -v ./tasks:/tasks:ro ghcr.io/driangle/taskmd
```

Open `http://localhost:8080` to view the dashboard.

**CLI commands:**
```bash
docker run --rm -v ./tasks:/tasks ghcr.io/driangle/taskmd taskmd list
docker run --rm -v ./tasks:/tasks ghcr.io/driangle/taskmd taskmd next
```

**Custom port:**
```bash
docker run --rm -p 3000:3000 -e TASKMD_WEB_PORT=3000 -v ./tasks:/tasks:ro ghcr.io/driangle/taskmd
```

**With docker-compose:**
```yaml
services:
  taskmd:
    image: ghcr.io/driangle/taskmd
    ports:
      - "8080:8080"
    volumes:
      - ./tasks:/tasks:ro
```

## Install with Go

Requires Go 1.25 or later:

```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

::: warning CLI only
`go install` builds the CLI without the web dashboard — the built frontend
assets are not part of the Go module, so `taskmd web start` will report
"No web UI embedded in this build". Use a [pre-built binary](#pre-built-binaries),
[Homebrew](#homebrew-macos-and-linux), or [Docker](#docker) if you want the web dashboard.
:::

Make sure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Build from Source

```bash
# Clone repository
git clone https://github.com/driangle/taskmd.git
cd taskmd/apps/cli

# Build CLI only
make build

# Build with embedded web interface
make build-full

# Binary will be at ./taskmd
```

## Verify Installation

```bash
taskmd --version
taskmd --help
```

## Next Steps

- [Quick Start](/getting-started/) - Create your first project
- [Core Concepts](/getting-started/concepts) - Understand the fundamentals
