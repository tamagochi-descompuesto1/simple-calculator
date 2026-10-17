# Stage 1: build the server binary. This image has the Go toolchain and is
# discarded afterwards; only the compiled binary moves on to the final image.
FROM golang:1.27.2-alpine AS build
WORKDIR /src

# Copy only the dependency files first, so Docker can reuse the cached download
# step on every build where they have not changed. go.sum is optional (the glob
# matches nothing while the project has no dependencies).
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produces a self-contained binary that does not need system
# libraries, so it can run in a different, minimal image. -s -w strips debug
# information to make it smaller.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: the image that actually runs. No Go, no source code, just the binary.
FROM alpine:3.24

# Run as an unprivileged user instead of root.
RUN adduser -D -u 10001 app
COPY --from=build /server /server
USER app

ENV PORT=8080
EXPOSE 8080

# Docker calls /health periodically and reports the container as healthy or not.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:${PORT}/health || exit 1

# Exec form (a JSON list) makes the server the main process, so `docker stop`
# delivers SIGTERM straight to it and the graceful shutdown runs. The shell form
# would put `sh` in between, which does not forward the signal.
ENTRYPOINT ["/server"]
