# syntax=docker/dockerfile:1
#
# Multi-stage, non-root, static build. Based on private-infra-kit's
# docker/Dockerfile.go-distroless.

# ── build stage ──────────────────────────────────────────────────────────────
FROM golang:1.26 AS build
WORKDIR /src
ENV GOTOOLCHAIN=auto CGO_ENABLED=0

# Dependencies first so a source-only change doesn't re-download the module cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/curly ./cmd

# ── runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian13:nonroot

# The old label pointed at github.com/supercakecrumb/curly-bot, a repo this
# image has never been built from — which breaks GHCR's "source" link.
LABEL org.opencontainers.image.source=https://github.com/supercakecrumb/curly-notification-telegram-bot
LABEL org.opencontainers.image.description="curly — self-hosted notification relay: HTTP in, Telegram out"

WORKDIR /app
COPY --from=build /out/curly /app/curly

# Mandatory per the wiki 'deployment' guideline. Exec form: distroless has no
# shell, so the JSON array is the only form that works here.
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD ["/app/curly", "--healthcheck"]

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/curly"]
