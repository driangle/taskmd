# Stage 1: Build web frontend
FROM node:20-alpine AS web-builder

RUN corepack enable && corepack prepare pnpm@9 --activate

WORKDIR /build

# Copy workspace config for pnpm
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/web/package.json apps/web/

# Install dependencies
RUN NODE_ENV=development pnpm install --frozen-lockfile --filter @taskmd/web...

# Copy web source and build
COPY apps/web/ apps/web/
RUN pnpm --filter @taskmd/web build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS go-builder

WORKDIR /build

# Copy Go workspace files and module definitions for dependency caching
COPY go.work go.work.sum ./
COPY apps/cli/go.mod apps/cli/go.sum apps/cli/
COPY sdk/go/go.mod sdk/go/go.sum sdk/go/

RUN cd apps/cli && go mod download

# Copy web dist from stage 1 into the embed location
COPY --from=web-builder /build/apps/web/dist /build/apps/cli/internal/web/static/dist

# Copy Go source
COPY apps/cli/ apps/cli/
COPY sdk/go/ sdk/go/

# Build args for version info
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN cd apps/cli && CGO_ENABLED=0 go build \
    -tags embed_web \
    -ldflags="-s -w \
      -X 'github.com/driangle/taskmd/apps/cli/internal/cli.Version=${VERSION}' \
      -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitCommit=${GIT_COMMIT}' \
      -X 'github.com/driangle/taskmd/apps/cli/internal/cli.BuildDate=${BUILD_DATE}'" \
    -o /taskmd ./cmd/taskmd

# Stage 3: Minimal runtime image
FROM alpine:3.20

ARG VERSION=dev

LABEL org.opencontainers.image.title="taskmd" \
      org.opencontainers.image.description="Markdown-based task management CLI and web dashboard" \
      org.opencontainers.image.url="https://github.com/driangle/taskmd" \
      org.opencontainers.image.source="https://github.com/driangle/taskmd" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}"

RUN apk add --no-cache ca-certificates tini

# Run as non-root user
RUN adduser -D -h /home/taskmd taskmd
USER taskmd

COPY --from=go-builder /taskmd /usr/local/bin/taskmd

WORKDIR /tasks

EXPOSE 8080

# Bind the web server to every interface inside the container.
# `taskmd web start` defaults to 127.0.0.1, which is correct on a host but
# unreachable through Docker's published-port DNAT: the container's network
# namespace is already the isolation boundary and `-p` is the operator's
# explicit opt-in. Set as ENV rather than baked into CMD so that documented
# CMD overrides (e.g. `web start --port 3000`) keep working unchanged.
ENV TASKMD_WEB_HOST=0.0.0.0

ENTRYPOINT ["tini", "--"]
CMD ["taskmd", "web", "start"]
