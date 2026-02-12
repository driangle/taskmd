---
id: "056"
title: "Implement .taskmd.yaml configuration file support"
status: pending
priority: medium
effort: small
dependencies: []
tags:
  - cli
  - configuration
  - enhancement
created: 2026-02-12
---

# Implement .taskmd.yaml Configuration File Support

## Objective

Add proper support for `.taskmd.yaml` configuration files to allow users to set default options without repeating command-line flags. The infrastructure (Viper) is already in place but not fully functional.

## Context

The CLI currently has Viper set up and binds flags, but the code reads flag variables directly instead of from Viper, so config file values are ignored. We need to:

1. Fix `GetGlobalFlags()` to read from Viper instead of flag variables
2. Add support for project-level `.taskmd.yaml` (not just `~/.taskmd.yaml`)
3. Support only essential config options: `dir`, `web.port`, and `web.open`

**Current Code Location:** `apps/cli/internal/cli/root.go`

## Tasks

- [ ] Update `initConfig()` to search for config in both home directory and current/project directory
  - Add current directory to config search path
  - Ensure project-level config takes precedence over global config
- [ ] Modify `GetGlobalFlags()` to read from Viper with flag values as overrides
  - Use `viper.GetString("dir")` instead of reading `dir` variable directly
  - Use `viper.GetBool("verbose")` instead of reading `verbose` variable directly
  - Ensure command-line flags override config file values (current behavior preserved)
- [ ] Add web server config support
  - Read `web.port` from config if not specified via flag
  - Read `web.open` from config if not specified via flag
  - Update `webStartCmd` flag initialization to use viper defaults
- [ ] Add tests for config file loading
  - Test global config (`~/.taskmd.yaml`)
  - Test project config (`./.taskmd.yaml`)
  - Test precedence: CLI flags > project config > global config > defaults
  - Test web config options
- [ ] Update documentation
  - Update `docs/FAQ.md` to reflect working config support
  - Update `README.md` with correct config examples
  - Add config file example to `docs/guides/cli-guide.md`
- [ ] Create example config file in `docs/` directory

## Implementation Details

### Supported Config Options

Only these options need config file support:

```yaml
# .taskmd.yaml
dir: ./tasks              # Default task directory
web:
  port: 8080             # Default web server port
  open: true             # Auto-open browser on web start
```

### Config Precedence (highest to lowest)

1. Command-line flags (explicit user intent)
2. Project-level `.taskmd.yaml` (project-specific defaults)
3. Global `~/.taskmd.yaml` (user-wide defaults)
4. Built-in defaults (fallback)

### Code Changes in root.go

**Update initConfig():**
```go
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        // Add current directory (project-level config)
        viper.AddConfigPath(".")

        // Add home directory (global config)
        home, err := os.UserHomeDir()
        if err == nil {
            viper.AddConfigPath(home)
        }

        viper.SetConfigType("yaml")
        viper.SetConfigName(".taskmd")
    }

    viper.SetEnvPrefix("TASKMD")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err == nil && verbose {
        fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    }
}
```

**Update GetGlobalFlags():**
```go
func GetGlobalFlags() GlobalFlags {
    return GlobalFlags{
        Stdin:   stdin,
        Format:  format,
        Quiet:   quiet,
        Verbose: viper.GetBool("verbose"),
        Dir:     viper.GetString("dir"),
    }
}
```

**Update web flag defaults:**
```go
func init() {
    rootCmd.AddCommand(webCmd)
    webCmd.AddCommand(webStartCmd)

    // Set defaults from viper if available
    defaultPort := viper.GetInt("web.port")
    if defaultPort == 0 {
        defaultPort = 8080
    }

    webStartCmd.Flags().IntVar(&webPort, "port", defaultPort, "server port")
    webStartCmd.Flags().BoolVar(&webDev, "dev", false, "enable dev mode")
    webStartCmd.Flags().BoolVar(&webOpen, "open", viper.GetBool("web.open"), "open browser on start")
}
```

## Acceptance Criteria

- [ ] Config files are loaded from both `~/.taskmd.yaml` and `./.taskmd.yaml`
- [ ] Project-level config takes precedence over global config
- [ ] Command-line flags override config file values
- [ ] `dir` option works from config file
- [ ] `web.port` option works from config file
- [ ] `web.open` option works from config file
- [ ] Tests verify config loading and precedence
- [ ] Documentation updated to reflect working config support
- [ ] Example config file provided in `docs/` directory

## Testing

### Manual Testing

1. Create `~/.taskmd.yaml` with `dir: ./my-tasks`
2. Run `taskmd list` and verify it uses `./my-tasks`
3. Run `taskmd list --dir ./other` and verify flag overrides config
4. Create `./.taskmd.yaml` with `dir: ./local-tasks`
5. Verify project config takes precedence over global config
6. Test web config: set `web.port: 3000` and verify `taskmd web start` uses port 3000

### Automated Testing

Create `internal/cli/config_test.go` with tests for:
- Config file loading from home directory
- Config file loading from current directory
- Precedence order (flags > project > global > defaults)
- Web server config options

## References

- Current implementation: `apps/cli/internal/cli/root.go`
- Current implementation: `apps/cli/internal/cli/web.go`
- Viper documentation: https://github.com/spf13/viper
- Task 054: FAQ documentation (needs updating after this is done)

## Notes

- We're deliberately limiting config options to `dir`, `web.port`, and `web.open` only
- Other flags like `format`, `verbose`, `quiet` should remain CLI-only for simplicity
- This keeps config files focused on project-specific settings, not per-invocation preferences
