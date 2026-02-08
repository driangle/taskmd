# Release Process

This document describes how to create a new release of `taskmd`.

## Overview

The release process is automated via GitHub Actions. When you push a version tag, the workflow will:

1. Build the web frontend (Vite + React SPA)
2. Embed the web assets into the Go binary
3. Cross-compile binaries for multiple platforms
4. Compress the binaries
5. Generate SHA256 checksums
6. Create a GitHub release with all artifacts attached

## Supported Platforms

The release workflow builds binaries for:

- **Linux**: AMD64, ARM64
- **macOS**: AMD64 (Intel), ARM64 (Apple Silicon)
- **Windows**: AMD64

All binaries include the embedded web dashboard.

## Creating a Release

### 1. Prepare the Release

Ensure all changes are committed and tests pass:

```bash
cd apps/cli
make check  # Runs tests and linting
```

### 2. Create and Push a Version Tag

```bash
# Create an annotated tag (recommended)
git tag -a v1.0.0 -m "Release v1.0.0"

# Or create a lightweight tag
git tag v1.0.0

# Push the tag to trigger the release workflow
git push origin v1.0.0
```

### 3. Monitor the Workflow

1. Go to the **Actions** tab in your GitHub repository
2. Watch the **Release** workflow run
3. It typically takes 3-5 minutes to complete

### 4. Verify the Release

Once the workflow completes:

1. Go to the **Releases** page in your GitHub repository
2. Verify the new release is published
3. Check that all 5 binary archives are attached:
   - `taskmd-v1.0.0-linux-amd64.tar.gz`
   - `taskmd-v1.0.0-linux-arm64.tar.gz`
   - `taskmd-v1.0.0-darwin-amd64.tar.gz`
   - `taskmd-v1.0.0-darwin-arm64.tar.gz`
   - `taskmd-v1.0.0-windows-amd64.zip`
   - `checksums.txt`

## Version Information

Each binary includes embedded version information that can be viewed with:

```bash
./taskmd --version
```

This displays:
- Version number (from the git tag)
- Git commit SHA
- Build date

## Troubleshooting

### Workflow Fails

If the release workflow fails:

1. Check the **Actions** tab for error logs
2. Common issues:
   - Web build failures: Check `apps/web/package.json` dependencies
   - Go build failures: Check `apps/cli/go.mod` and imports
   - Permission errors: Verify the workflow has `contents: write` permission

### Missing Artifacts

If binaries are missing from the release:

1. Check the **Compress binaries** step completed successfully
2. Verify the file paths in the **Create Release** step match the generated files

### Re-running a Release

To recreate a release:

1. Delete the existing release and tag from GitHub
2. Delete the local tag: `git tag -d v1.0.0`
3. Create and push the tag again

## Release Checklist

- [ ] All tests pass (`make check`)
- [ ] Documentation is up to date
- [ ] CHANGELOG is updated (if applicable)
- [ ] Version tag follows semantic versioning (vMAJOR.MINOR.PATCH)
- [ ] Tag is pushed to GitHub
- [ ] Release workflow completes successfully
- [ ] All platform binaries are attached to the release
- [ ] Checksums file is included
- [ ] Release notes are accurate

## Semantic Versioning

Follow [Semantic Versioning](https://semver.org/) for version numbers:

- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features, backward compatible
- **PATCH** (v1.0.1): Bug fixes, backward compatible

## Pre-releases

For pre-release versions, use suffixes:

- Alpha: `v1.0.0-alpha.1`
- Beta: `v1.0.0-beta.1`
- Release candidate: `v1.0.0-rc.1`

The workflow will mark releases with these suffixes as "pre-release" automatically.

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

LDFLAGS="-X 'github.com/driangle/md-task-tracker/apps/cli/internal/cli.Version=${VERSION}' \
         -X 'github.com/driangle/md-task-tracker/apps/cli/internal/cli.GitCommit=${GIT_COMMIT}' \
         -X 'github.com/driangle/md-task-tracker/apps/cli/internal/cli.BuildDate=${BUILD_DATE}'"

GOOS=linux GOARCH=amd64 go build -tags embed_web -ldflags="$LDFLAGS" -o taskmd-linux-amd64 ./cmd/taskmd
# ... repeat for other platforms
```

However, using the automated workflow is strongly recommended for consistency and reproducibility.
