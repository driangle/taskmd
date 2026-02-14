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

::: tip
Only project-level settings are supported in config files. Per-invocation preferences like `format`, `verbose`, and `quiet` are intentionally CLI-only.
:::

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

## Shell Aliases

For quick access, add aliases to your shell config:

```bash
# ~/.bashrc or ~/.zshrc
alias tm='taskmd --dir ./tasks'
alias tmw='taskmd web start --port 8080 --open'
alias tnext='taskmd next --limit 3'
alias thigh='taskmd list --filter priority=high --filter status=pending'
```
