# Releasing

How to create a new release of taskmd.

## Overview

Releases are driven by `scripts/release.sh`. The script bumps versions, tags the SDK and plugins when they changed, pushes a `vX.Y.Z` tag, and then the **Release** GitHub Actions workflow takes over. The workflow:

1. Runs the gating jobs: lint and test the CLI, test the SDK, check the web frontend and VS Code extension, build the docs, and build the Docker image
2. Builds the web frontend and embeds it into the Go binary
3. Cross-compiles binaries for every supported platform, compresses them, and builds the MCPB bundles
4. Generates SHA256 checksums and creates the GitHub release with all artifacts attached
5. Updates the Homebrew formula in `driangle/homebrew-tap`
6. Builds and pushes the Docker image to GHCR

Do not create the tag by hand. A bare `git tag` skips the version bump, the `sdk/go` tag and pin, the plugin version gate, and the release notes.

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

### 2. Preview

From a clean, pushed `main` (the script refuses to run with uncommitted changes or unpushed commits), run a dry run first. It validates the version, and reports every SDK or plugin version the real run will demand, without changing anything:

```bash
./scripts/release.sh --dry-run 0.8.1
```

### 3. Release

Write the release notes to a file, then run the script:

```bash
./scripts/release.sh 0.8.1 --notes-file notes.md
```

The script, in order: validates the tree, updates the version in `apps/cli/internal/cli/root.go` and the `package.json` files, bumps any plugin manifests you passed versions for, tags and pushes `sdk/go` if it changed and repoints the `apps/cli/go.mod` pin at that tag, verifies the CLI builds with `GOWORK=off`, commits, creates the annotated `vX.Y.Z` tag, pushes, waits for the workflow, and applies the release notes to the GitHub release.

Two things can stop it:

- **`sdk/go` changed since its last tag.** Pass `--sdk-version X.Y.Z`. Pre-1.0, a breaking API change is a minor bump and anything additive or a fix is a patch bump.
- **A plugin directory changed since the last release tag.** Pass `--plugin-taskmd-version`, `--plugin-lite-version`, or `--plugin-mcp-version` for each one. The dry run lists every missing bump at once. See [ADR 0003](https://github.com/driangle/taskmd/blob/main/docs/adr/0003-plugin-versioning-policy.md) for how to size them.

```bash
./scripts/release.sh 0.8.1 --notes-file notes.md \
  --sdk-version 0.4.10 \
  --plugin-mcp-version 1.1.2
```

Other flags: `--no-push` creates the tag locally without pushing, for testing; `--skip-checks` bypasses the clean-tree and branch checks and should not be needed for a normal release.

### 4. Verify

The script waits for the workflow and prints the release URL. Check the **Releases** page for:
- `taskmd-v0.8.1-linux-amd64.tar.gz`
- `taskmd-v0.8.1-linux-arm64.tar.gz`
- `taskmd-v0.8.1-darwin-amd64.tar.gz`
- `taskmd-v0.8.1-darwin-arm64.tar.gz`
- `taskmd-v0.8.1-windows-amd64.zip`
- `taskmd-v0.8.1-windows-arm64.zip`
- `taskmd-v0.8.1-*.mcpb` (6 MCPB bundles, one per platform/arch)
- `checksums.txt`

Then confirm the Homebrew formula and the `ghcr.io/driangle/taskmd` image carry the new version.

## Version Information

Each binary includes embedded version information:

```bash
./taskmd --version
# Shows: version number, git commit SHA, build date
```

## Versioning

Four things version independently, each on its own line:

| What | Tag or file | Rule |
|------|-------------|------|
| CLI / repo | `vX.Y.Z` | Pre-1.0. Every release. |
| `sdk/go` | `sdk/go/vX.Y.Z` | Pre-1.0: breaking change is a minor bump, otherwise patch. Tagged only when the module changed. Module versions are immutable; never retag. |
| `taskmd`, `taskmd-lite` plugins | `<plugin>/.claude-plugin/plugin.json` | Pre-1.0, bumped when the directory changed. |
| `taskmd-mcp` plugin | `claude-code-plugin-mcp/.claude-plugin/plugin.json` | Stable `1.x`: a changed tool signature is a major bump. |

Pre-release suffixes such as `0.9.0-rc.1` are accepted by the script. Full rules are in [ADR 0003](https://github.com/driangle/taskmd/blob/main/docs/adr/0003-plugin-versioning-policy.md) and the "Versioning" section of `AGENTS.md`.

## Release Checklist

- [ ] All tests pass (`make check`)
- [ ] Documentation is up to date
- [ ] `./scripts/release.sh --dry-run X.Y.Z` passes and you have a version for every SDK or plugin bump it demands
- [ ] `./scripts/release.sh X.Y.Z --notes-file notes.md ...` completes
- [ ] Release workflow completes successfully
- [ ] Docs site redeploys automatically (the script's `package.json` bump on `main` triggers it)
- [ ] All platform binaries are attached
- [ ] All MCPB bundles are attached (6 total)
- [ ] Checksums file is included
- [ ] Release notes are accurate

## Troubleshooting

### Workflow Fails

Check the **Actions** tab for error logs. Common issues:
- Web build failures: check `apps/web/package.json` dependencies
- Go build failures: check `apps/cli/go.mod` and imports
- Permission errors: verify the workflow has `contents: write` (release) and `packages: write` (GHCR image) permissions
- Homebrew step fails: the `HOMEBREW_TAP_TOKEN` secret must be able to push to `driangle/homebrew-tap`

### Missing Artifacts

If binaries are missing from the release:

1. Check the **Compress binaries** step completed successfully
2. Verify the file paths in the **Create Release** step match the generated files

### Re-running a Release

If the workflow failed after the tag was pushed:

1. Delete the GitHub release and the remote tag
2. Delete the local tag: `git tag -d v0.8.1`
3. Fix the cause, then run `./scripts/release.sh` again with the same version

Do not reuse an `sdk/go` tag that was already pushed; the Go module proxy may have cached it. Pick the next patch version instead.

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
VERSION="0.8.1"
GIT_COMMIT=$(git rev-parse HEAD)
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-X 'github.com/driangle/taskmd/apps/cli/internal/cli.Version=${VERSION}' \
         -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitCommit=${GIT_COMMIT}' \
         -X 'github.com/driangle/taskmd/apps/cli/internal/cli.BuildDate=${BUILD_DATE}'"

GOOS=linux GOARCH=amd64 go build -tags embed_web -ldflags="$LDFLAGS" -o taskmd-linux-amd64 ./cmd/taskmd
# ... repeat for other platforms
```

Using the automated workflow is strongly recommended for consistency and reproducibility.
