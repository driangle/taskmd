# Releasing

How to create a new release of taskmd.

## Overview

The release process is automated via GitHub Actions. When you push a version tag, the workflow will:

1. Build the web frontend (Vite + React SPA)
2. Embed the web assets into the Go binary
3. Cross-compile binaries for multiple platforms
4. Compress the binaries
5. Generate SHA256 checksums
6. Create a GitHub release with all artifacts attached

## Supported Platforms

| Platform | Architecture |
|----------|-------------|
| Linux | AMD64, ARM64 |
| macOS | AMD64 (Intel), ARM64 (Apple Silicon) |
| Windows | AMD64, ARM64 |

All binaries include the embedded web dashboard.

Each release also produces **MCPB bundles** (`.mcpb` files) for all 6 platform/architecture combinations. MCPB bundles enable one-click MCP server installation in clients that support the format.

## Creating a Release

### 1. Prepare

Ensure all changes are committed and tests pass:

```bash
cd apps/cli
make check  # Runs tests and linting
```

### 2. Tag and Push

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to trigger the release workflow
git push origin v1.0.0
```

### 3. Monitor

Go to the **Actions** tab in your GitHub repository and watch the **Release** workflow. It typically takes 3-5 minutes.

### 4. Verify

Check the **Releases** page for:
- `taskmd-v1.0.0-linux-amd64.tar.gz`
- `taskmd-v1.0.0-linux-arm64.tar.gz`
- `taskmd-v1.0.0-darwin-amd64.tar.gz`
- `taskmd-v1.0.0-darwin-arm64.tar.gz`
- `taskmd-v1.0.0-windows-amd64.zip`
- `taskmd-v1.0.0-windows-arm64.zip`
- `taskmd-v1.0.0-*.mcpb` (6 MCPB bundles, one per platform/arch)
- `checksums.txt`

## Version Information

Each binary includes embedded version information:

```bash
./taskmd --version
# Shows: version number, git commit SHA, build date
```

## Semantic Versioning

Follow [semver.org](https://semver.org/):

- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features, backward compatible
- **PATCH** (v1.0.1): Bug fixes, backward compatible

Pre-release suffixes: `v1.0.0-alpha.1`, `v1.0.0-beta.1`, `v1.0.0-rc.1`

## Release Checklist

- [ ] All tests pass (`make check`)
- [ ] Documentation is up to date
- [ ] Version tag follows semantic versioning
- [ ] Tag is pushed to GitHub
- [ ] Release workflow completes successfully
- [ ] Docs site redeploys automatically (triggered by `package.json` version bump on `main`)
- [ ] All platform binaries are attached
- [ ] All MCPB bundles are attached (6 total)
- [ ] Checksums file is included
- [ ] Release notes are accurate

## Troubleshooting

### Workflow Fails

Check the **Actions** tab for error logs. Common issues:
- Web build failures: check `apps/web/package.json` dependencies
- Go build failures: check `apps/cli/go.mod` and imports
- Permission errors: verify the workflow has `contents: write` permission

### Missing Artifacts

If binaries are missing from the release:

1. Check the **Compress binaries** step completed successfully
2. Verify the file paths in the **Create Release** step match the generated files

### Re-running a Release

1. Delete the existing release and tag from GitHub
2. Delete the local tag: `git tag -d v1.0.0`
3. Create and push the tag again

## Manual Release (Not Recommended)

If you need to build releases manually:

```bash
# Build web frontend
cd apps/web
pnpm install
pnpm build

# Copy to CLI
cd ../cli
mkdir -p internal/web/static
cp -r ../web/dist internal/web/static/dist

# Build for all platforms
VERSION="1.0.0"
GIT_COMMIT=$(git rev-parse HEAD)
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-X 'github.com/driangle/taskmd/apps/cli/internal/cli.Version=${VERSION}' \
         -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitCommit=${GIT_COMMIT}' \
         -X 'github.com/driangle/taskmd/apps/cli/internal/cli.BuildDate=${BUILD_DATE}'"

GOOS=linux GOARCH=amd64 go build -tags embed_web -ldflags="$LDFLAGS" -o taskmd-linux-amd64 ./cmd/taskmd
# ... repeat for other platforms
```

Using the automated workflow is strongly recommended for consistency and reproducibility.
