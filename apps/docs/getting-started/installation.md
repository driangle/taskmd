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
# Download the archive for your platform
# Extract it
tar -xzf taskmd-v*.tar.gz  # macOS/Linux
# or unzip for Windows

# Move to a directory in your PATH
sudo mv taskmd /usr/local/bin/  # macOS/Linux
```

Available platforms:
- **Linux**: AMD64, ARM64
- **macOS**: AMD64 (Intel), ARM64 (Apple Silicon)
- **Windows**: AMD64

## Install with Go

Requires Go 1.22 or later:

```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

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
