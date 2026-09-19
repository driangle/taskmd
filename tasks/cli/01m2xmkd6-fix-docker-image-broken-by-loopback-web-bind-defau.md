---
title: "Fix Docker image broken by loopback web bind default"
id: "01m2xmkd6"
status: pending
priority: high
type: bug
tags: ["docker", "web", "ci"]
created: "2026-09-19"
---

# Fix Docker image broken by loopback web bind default

`taskmd web start` now defaults to binding `127.0.0.1` (issue #23, branch
`feat/web-host-bind`). Both Dockerfiles run that command as their default
`CMD`, so the server binds loopback *inside the container's network
namespace* and Docker's published-port DNAT to the container IP is refused.
Every documented `docker run -p` invocation is broken.

The CI Docker smoke test does not catch this, so the regression would ship
green.

## Steps to Reproduce

1. Build the image from `Dockerfile.release` with a binary built from
   `feat/web-host-bind`.
2. `docker run -d -p 8080:8080 -v "$(pwd)/tasks:/tasks:ro" <image>`
3. `curl -sf http://localhost:8080/api/tasks`

## Expected Behavior

The dashboard is reachable on the published port, as documented in
`README.md`, `apps/docs/getting-started/installation.md` and
`apps/docs/guide/web.md`.

## Actual Behavior

`curl` fails with exit 56 (connection reset). Container logs show:

```
taskmd web server running at http://127.0.0.1:8080
```

Verified empirically: `--host 0.0.0.0` and `-e TASKMD_WEB_HOST=0.0.0.0` both
restore reachability in the same image.

## Root Cause

A container's network namespace is itself the isolation boundary, and `-p` is
the operator's explicit opt-in to exposure. Wildcard is therefore the correct
bind *inside* a container, even though loopback is the correct default on a
host. The image must opt back in; the fix does not belong in the docs.

## Tasks

- [ ] Add an explicit wildcard bind to `Dockerfile` (line ~76) and
      `Dockerfile.release` (line ~29) — either `CMD ["taskmd", "web", "start",
      "--host", "0.0.0.0"]` or `ENV TASKMD_WEB_HOST=0.0.0.0`. Prefer the `ENV`
      form if `CMD` overrides in the docs should keep working without the flag.
- [ ] Fix the CI assertion in `.github/workflows/ci.yml` (~line 262): the
      "Test Docker image" step runs under Actions' default `bash -e {0}` with
      no `pipefail`, so `curl -sf ... | head -c 500` always exits 0. Add
      `shell: bash` to the step so the pipeline status is honored.
- [ ] Confirm the smoke test actually fails against an unfixed image before
      landing the Dockerfile change (guard against re-introducing false
      confidence).
- [ ] Document `TASKMD_WEB_HOST` in the Environment Variables section of
      `apps/docs/reference/configuration.md`; it is the natural fix for Docker
      Compose users and is currently undocumented (as is `TASKMD_WEB_PORT`,
      which `installation.md` already uses).
- [ ] Re-check the Docker Compose example in `apps/docs/guide/web.md` (~line
      498) once the image is fixed.

## Acceptance Criteria

- [ ] `docker run -d -p 8080:8080 -v ./tasks:/tasks:ro <image>` serves
      `GET /api/tasks` on the host with no extra flags.
- [ ] Both `Dockerfile` and `Dockerfile.release` produce a reachable server.
- [ ] The CI Docker smoke test fails (non-zero exit) when the server is
      unreachable, demonstrated against a deliberately loopback-bound image.
- [ ] Every `docker run` example in `README.md`,
      `apps/docs/getting-started/installation.md` and `apps/docs/guide/web.md`
      works as written.
- [ ] `TASKMD_WEB_HOST` appears in the configuration reference.

## Notes

The precedence table in the Environment Variables section of
`apps/docs/reference/configuration.md` lists env vars *below* both config
files, which contradicts viper's actual order (flag -> env -> config).
Pre-existing and out of scope, but it is on the page a Compose user would
read while fixing this.
