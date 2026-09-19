# Worklog: 01m2xmkd6 — Fix Docker image broken by loopback web bind default

## 2026-09-19 — Started

Picked up the task in worktree `.claude/worktrees/01m2xmkd6` (branch
`task/01m2xmkd6`). Confirmed both `Dockerfile` and `Dockerfile.release` end in
`CMD ["taskmd", "web", "start"]`, which since 1d21674 binds 127.0.0.1 inside
the container's network namespace.

## 2026-09-19 — Decision: ENV, not CMD

Chose `ENV TASKMD_WEB_HOST=0.0.0.0` over baking `--host 0.0.0.0` into `CMD`.
The docs contain CMD overrides (`web start --port 3000`, `web start
--readonly`); with the flag in `CMD` those overrides would silently drop the
wildcard bind and re-break the image. The env var survives any override.
`TASKMD_WEB_HOST` works without new code via viper's `SetEnvPrefix("TASKMD")`
+ `AutomaticEnv` (apps/cli/internal/cli/root.go:158-160), and a bound-but-
unchanged pflag default loses to env in viper's precedence.

## 2026-09-19 — CI smoke test

The "Test Docker image" step had no `shell: bash`, so it ran under Actions'
default `bash -e {0}` (no pipefail) and `curl -sf ... | head -c 500` took its
status from `head`. Added `shell: bash`.

## 2026-09-19 — Empirical verification

Cross-compiled linux/arm64 and built both images locally.

- Fixed `Dockerfile.release`: logs `Bound to 0.0.0.0`, `curl
  http://localhost:PORT/api/tasks` returns the task JSON (exit 0).
- Full `Dockerfile` (web assets embedded): same, plus `/` returns 200.
- Documented examples all verified: `--port 3000` CMD override,
  `-e TASKMD_WEB_PORT=3000`, `web start --readonly`, `taskmd list`.
- Loopback-bound control image (ENV line stripped): the exact CI script run
  under `bash -e` exits **0** (the old false green); under `bash -eo pipefail`
  it exits **56**. That is the acceptance criterion about demonstrating the
  smoke test now fails against a deliberately loopback-bound image.

Cleaned up all verification containers and images.

## 2026-09-19 — Regression guards

Added `apps/cli/internal/cli/docker_bind_test.go`: asserts both Dockerfiles
set `ENV TASKMD_WEB_HOST=0.0.0.0` and that the Docker smoke-test step declares
`shell: bash` (comment lines stripped so the explanatory comment cannot
satisfy the guard). Both confirmed to fail when the corresponding fix is
reverted.

## 2026-09-19 — Docs

Documented `TASKMD_WEB_HOST` and `TASKMD_WEB_PORT` in the Environment
Variables section of `apps/docs/reference/configuration.md` as a table with
the config-key mapping, plus a Docker tip. Added a matching tip to the Docker
Compose section of `apps/docs/guide/web.md`.

Left the pre-existing precedence-table inaccuracy on that page alone — noted
in the task as out of scope.
