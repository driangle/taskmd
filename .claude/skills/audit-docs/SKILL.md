---
name: audit-docs
description: Audit documentation coverage across CLI commands, web features, and configuration. Builds the CLI, discovers all commands/flags, checks web pages, and cross-references against docs/ and apps/docs/.
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Task
---

# Audit Documentation

Discover all CLI commands, web features, and configuration options, then cross-reference against the documentation in `docs/` and `apps/docs/` to identify gaps, outdated content, and missing sections.

## Instructions

Arguments in `$ARGUMENTS` are optional flags:
- `--fix` — after reporting gaps, update the documentation files to fill them
- `--cli-only` — only audit CLI command documentation
- `--web-only` — only audit web feature documentation
- `--verbose` — include per-flag detail in the report

If no arguments are provided, run a full audit and produce a report (no fixes).

### Phase 1: Build the CLI

1. Build the development binary so we have the latest commands:
   ```bash
   cd apps/cli && make install-dev
   ```
2. Verify the binary works:
   ```bash
   taskmd-dev --version
   ```

### Phase 2: Discover CLI commands and flags

1. **Get the top-level help**: Run `taskmd-dev --help` and parse the list of available commands.
2. **For each command**, run `taskmd-dev <command> --help` to capture:
   - Usage string
   - Description (Short + Long)
   - All flags (name, type, default, description)
   - Subcommands (e.g., `web start`)
3. For commands with subcommands, recurse: run `taskmd-dev <command> <subcommand> --help`.
4. Also capture global flags from `taskmd-dev --help`.
5. Store all of this as a structured inventory for comparison.

### Phase 3: Discover web features

1. Read the web app source to identify all pages/views:
   - Glob `apps/web/src/pages/*.tsx` for page components
   - Glob `apps/web/src/components/*/` for feature areas
2. Read `apps/web/src/App.tsx` to extract route definitions (URL paths).
3. Read `apps/cli/internal/web/*.go` to identify API endpoints served by the backend.
4. Compile a list of:
   - **Pages**: name, URL path, key features
   - **API endpoints**: method, path, description

### Phase 4: Discover configuration options

1. Read `apps/cli/internal/cli/root.go` to extract global flags and config loading.
2. Read any config-related files (grep for `.taskmd.yaml` or `viper` usage) to find all supported config keys.
3. Check `docs/.taskmd.yaml.example` for documented config options.

### Phase 5: Cross-reference against documentation

For each discovered item, check whether it is documented in both locations:

#### CLI Commands → Documentation

For each command discovered in Phase 2:

1. **`apps/docs/guide/cli.md`** (VitePress site):
   - Is the command listed in the Quick Reference table?
   - Does it have its own section with usage examples?
   - Are all flags documented?
   - Are examples accurate (flag names, defaults match reality)?

2. **`docs/guides/cli-guide.md`** (standalone docs):
   - Same checks as above.

#### Web Features → Documentation

For each page/API endpoint discovered in Phase 3:

1. **`apps/docs/guide/web.md`** (VitePress site):
   - Is the page/view listed?
   - Are key features described?
   - Are API endpoints documented?

2. **`docs/guides/web-guide.md`** (standalone docs):
   - Same checks.

#### Configuration → Documentation

For each config key discovered in Phase 4:

1. **`apps/docs/reference/configuration.md`**:
   - Is the option listed in the Supported Options table?
   - Is there an example?

#### Specification → Documentation

1. Compare `docs/taskmd_specification.md` (authoritative spec) with `apps/docs/reference/specification.md` (VitePress version).
   - Check version numbers match.
   - Check all frontmatter fields are listed in both.
   - Check enum values (status, priority, effort) match.

#### VitePress Navigation

1. Read `apps/docs/.vitepress/config.ts` and verify:
   - All docs pages listed in the sidebar actually exist as files.
   - No orphan pages exist that aren't linked from the sidebar.

### Phase 6: Generate report

Produce a structured markdown report with these sections:

```
## Documentation Audit Report

### Summary
- Total CLI commands: X (documented: Y, missing: Z)
- Total CLI flags: X (documented: Y, missing: Z)
- Total web pages: X (documented: Y, missing: Z)
- Total API endpoints: X (documented: Y, missing: Z)
- Total config options: X (documented: Y, missing: Z)

### CLI Command Gaps
For each undocumented or partially documented command:
- Command name
- What's missing (entire section, specific flags, examples)
- Which doc file(s) need updating

### Web Feature Gaps
For each undocumented page or endpoint:
- Feature name
- What's missing
- Which doc file(s) need updating

### Configuration Gaps
For each undocumented config option:
- Option name
- Which doc file(s) need updating

### Stale/Inaccurate Documentation
Items where docs don't match the actual CLI behavior:
- Flag name mismatches
- Wrong defaults
- Missing or renamed commands
- Version mismatches between spec files

### VitePress Navigation Issues
- Broken sidebar links
- Orphan pages
```

Print the report to stdout.

### Phase 7: Fix gaps (only if `--fix` was passed)

If `$ARGUMENTS` contains `--fix`:

1. For each gap identified in Phase 6, update the relevant documentation file:
   - Add missing command sections to `apps/docs/guide/cli.md`
   - Add missing flag tables
   - Add missing web feature descriptions to `apps/docs/guide/web.md`
   - Add missing config options to `apps/docs/reference/configuration.md`
   - Sync specification versions
2. Follow the existing documentation style and formatting conventions in each file.
3. After making changes, list all modified files.

If `--fix` was NOT passed, end with: "Run `/audit-docs --fix` to automatically update documentation."

### Error Handling

- If `make install-dev` fails, report the build error and stop.
- If a command's `--help` output can't be parsed, note it in the report and continue.
- If a documentation file doesn't exist, note it as a gap.
