# Build stage: Go toolchain, discarded after the build.
FROM golang:1.27.2-alpine AS build
WORKDIR /src

# Dependency files first, so the download layer is cached. go.sum is optional
# while there are no dependencies.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gives a static binary that runs in a minimal image. -s -w
# strips debug information.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

# Run stage: only the binary.
FROM alpine:3.24

RUN adduser -D -u 10001 app
COPY --from=build /server /server
USER app

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:${PORT}/health || exit 1

# Exec form: `docker stop` sends SIGTERM straight to the server, so the graceful
# shutdown runs. The shell form would put `sh` in between.
ENTRYPOINT ["/server"]
