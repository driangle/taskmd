# Configuration

taskmd supports `.taskmd.yaml` configuration files for setting default options.

## Config File

Create a `.taskmd.yaml` file in your project root or home directory:

```yaml
# .taskmd.yaml

# Default directory to search for task files
dir: ./tasks

# Web server configuration
web:
  # Default port for the web dashboard
  port: 8080

  # Automatically open browser when starting the web server
  auto_open_browser: false
```

## Config File Locations

Config files are loaded in this order (highest precedence first):

1. **Project-level**: `./.taskmd.yaml` - project-specific settings
2. **Global**: `~/.taskmd.yaml` - user-wide defaults
3. **Custom**: `--config path/to/config.yaml` - explicit path
4. **Built-in defaults** - fallback values

Command-line flags always override config file values.

## Supported Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dir` | string | `.` | Default task directory |
| `web.port` | integer | `8080` | Web server port |
| `web.auto_open_browser` | boolean | `false` | Auto-open browser on `web start` |
| `scopes` | map | — | Scope-to-path mappings for the `touches` field ([details](#scopes-configuration)) |

::: tip
Only project-level settings are supported in config files. Per-invocation preferences like `format`, `verbose`, and `quiet` are intentionally CLI-only.
:::

## Scopes Configuration {#scopes-configuration}

The `scopes` key defines concrete mappings for the abstract scope identifiers used by the [`touches`](/reference/specification#frontmatter-schema) frontmatter field. The `tracks` command uses `touches` to detect spatial overlap between tasks and assign them to parallel work tracks — tasks sharing a scope are placed in separate tracks to avoid merge conflicts.

```yaml
# .taskmd.yaml
scopes:
  cli/graph:
    description: "Graph visualization and dependency rendering"
    paths:
      - "apps/cli/internal/graph/"
      - "apps/cli/internal/cli/graph.go"
  cli/output:
    paths:
      - "apps/cli/internal/cli/format.go"
```

Each scope entry has the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `description` | No | Human-readable explanation of what the scope covers. Included in validation error messages. |
| `paths` | No | List of file or directory paths that the scope maps to. |

**Behavior:**

- When scopes are configured, any `touches` value in a task that does not match a configured scope produces a warning.
- When no scopes config exists, all `touches` values are accepted silently.

## Usage Examples

### Project Setup

```bash
# Create project config
cat > .taskmd.yaml <<EOF
dir: ./tasks
web:
  port: 3000
  auto_open_browser: true
EOF

# Now these commands use config defaults
taskmd list              # Uses ./tasks directory
taskmd web start         # Uses port 3000 and opens browser

# CLI flags still override config
taskmd list --dir ./other-tasks
taskmd web start --port 8080
```

### Global Defaults

Create `~/.taskmd.yaml` for defaults that apply to all projects:

```yaml
web:
  port: 3000
  auto_open_browser: true
```

## Environment Variables

taskmd supports environment variables with the `TASKMD_` prefix:

```bash
export TASKMD_DIR=./tasks
export TASKMD_VERBOSE=true
```

**Precedence** (highest to lowest):
1. Command-line flags
2. Project-level `.taskmd.yaml`
3. Global `~/.taskmd.yaml`
4. Environment variables
5. Built-in defaults

## Sync Configuration {#sync-configuration}

The `sync` command reads its configuration from the `sync` section of `.taskmd.yaml`. Each source defines where to fetch tasks from, how to map fields, and where to write files.

```yaml
# .taskmd.yaml
dir: ./tasks

sync:
  sources:
    - name: github
      project: "owner/repo"
      token_env: GITHUB_TOKEN       # Environment variable holding the API token
      output_dir: ./tasks/synced     # Where to write synced task files
      field_map:
        status:
          open: pending
          closed: completed
        priority:
          urgent: critical
          high: high
          medium: medium
          low: low
        labels_to_tags: true         # Convert issue labels to task tags
        assignee_to_owner: true      # Map assignee to owner field
      filters:
        state: open                  # Only sync open issues
```

**Source fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique name for this source |
| `project` | No | Project identifier (e.g., `owner/repo` for GitHub) |
| `base_url` | No | Custom API base URL |
| `token_env` | No | Environment variable name for API token |
| `user_env` | No | Environment variable name for username |
| `output_dir` | Yes | Directory where synced task files are written |
| `field_map` | No | How to map external fields to taskmd frontmatter |
| `filters` | No | Source-specific filters (e.g., `state: open`) |

**Field mapping (`field_map`):**

| Sub-field | Type | Description |
|-----------|------|-------------|
| `status` | `map[string]string` | Map external status values to taskmd statuses |
| `priority` | `map[string]string` | Map external priority values to taskmd priorities |
| `labels_to_tags` | `bool` | Convert external labels/categories to task tags |
| `assignee_to_owner` | `bool` | Map external assignee to the `owner` field |

## Shell Aliases

For quick access, add aliases to your shell config:

```bash
# ~/.bashrc or ~/.zshrc
alias tm='taskmd --dir ./tasks'
alias tmw='taskmd web start --port 8080 --open'
alias tnext='taskmd next --limit 3'
alias thigh='taskmd list --filter priority=high --filter status=pending'
```
