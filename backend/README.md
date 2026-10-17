# Calculator API (backend)

A small REST API, written in Go, that performs arithmetic operations. It uses only
the standard library (`net/http`, `encoding/json`, `log/slog`), with no third-party
dependencies.

## Quick start

With Go (version in [`go.mod`](go.mod)):

```bash
cd backend
go run ./cmd/server          # listens on :8080
curl localhost:8080/health   # {"status":"ok"}
```

With Docker (no Go needed):

```bash
docker build -t calculator-backend ./backend
docker run --rm -p 8080:8080 calculator-backend
```

## Configuration

Configuration comes from environment variables. See [`.env.example`](../.env.example).

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Port the server listens on |
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5173` | Origin allowed to call the API from a browser (the frontend) |

## API

Every operation is its own `POST` endpoint. Requests and responses are JSON.

| Endpoint | Body | Result |
|---|---|---|
| `POST /add` | `{"a": n, "b": n}` | `a + b` |
| `POST /subtract` | `{"a": n, "b": n}` | `a - b` |
| `POST /multiply` | `{"a": n, "b": n}` | `a * b` |
| `POST /divide` | `{"a": n, "b": n}` | `a / b` |
| `POST /power` | `{"a": n, "b": n}` | `a` raised to the power `b` |
| `POST /percentage` | `{"a": n, "b": n}` | `b` percent of `a`, i.e. `a * b / 100` |
| `POST /sqrt` | `{"a": n}` | square root of `a` |
| `GET /health` | none | `{"status":"ok"}` |

A successful response is `200` with `{"result": <number>}`. Any failure is
`{"error": "<message>"}` with one of these status codes:

| Status | When | Examples |
|---|---|---|
| `400 Bad Request` | The request is malformed | invalid JSON, empty body, missing field, unknown field, wrong type, extra data after the JSON object, body over 1 MB |
| `422 Unprocessable Entity` | The request is well formed, but the operation is not possible | division by zero, square root of a negative number, result too large, result not a real number |
| `500 Internal Server Error` | Unexpected failure. The message is always generic; the real error is only logged | |
| `404` / `405` | Unknown route / wrong method. Produced by the router, as plain text | |

### Examples

```console
$ curl -s -X POST localhost:8080/add -d '{"a": 6, "b": 3}'
{"result":9}

$ curl -s -X POST localhost:8080/percentage -d '{"a": 200, "b": 15}'
{"result":30}

$ curl -s -X POST localhost:8080/sqrt -d '{"a": 16}'
{"result":4}
```

Domain errors (`422`):

```console
$ curl -s -X POST localhost:8080/divide -d '{"a": 5, "b": 0}'
{"error":"division by zero"}

$ curl -s -X POST localhost:8080/sqrt -d '{"a": -4}'
{"error":"square root of a negative number"}

$ curl -s -X POST localhost:8080/power -d '{"a": 10, "b": 400}'
{"error":"result is too large to represent"}

$ curl -s -X POST localhost:8080/power -d '{"a": -8, "b": 0.5}'
{"error":"result is not a real number"}
```

Malformed requests (`400`):

```console
$ curl -s -X POST localhost:8080/add -d '{"a": 1}'
{"error":"missing field: b"}

$ curl -s -X POST localhost:8080/add -d '{"a": 1, "b": 2, "c": 3}'
{"error":"unknown field \"c\""}

$ curl -s -X POST localhost:8080/add -d '{"a": "one", "b": 2}'
{"error":"field a must be a valid number"}
```

## Project layout

```
backend/
├── cmd/server/          entry point: configuration, middleware chain, graceful shutdown
├── internal/calculator/ the operations and their math rules; knows nothing about HTTP
├── internal/handler/    HTTP layer: request parsing, status codes, JSON responses, routes
└── internal/middleware/ panic recovery, request logging, CORS
```

Dependencies only point one way: `cmd/server` → `handler` → `calculator`. The
calculator can be used and tested without any HTTP code.

## Design decisions

**Two layers of validation.** The `handler` checks the *shape* of a request (valid
JSON, fields present, right types). The `calculator` enforces the *math* (division
by zero, negative square root, unrepresentable results). Keeping the rules in the
calculator means they are unit-testable without HTTP, and the generic handler does
not need to know which operation it is running.

**One endpoint per operation.** `POST /divide` maps to `calculator.Divide`, which is
easy to read, document and test. A generic `{"operation": "add"}` endpoint, or an
expression parser, would add complexity the task does not call for.

**Errors are values.** Every operation returns `(float64, error)`, and the calculator
exposes sentinel errors (`ErrDivisionByZero`, `ErrNegativeSqrt`, `ErrOverflow`,
`ErrUndefined`). The handler recognizes them with `errors.Is`, so renaming a message
never breaks the mapping. A single table in `handler/response.go` maps each sentinel
to its HTTP status; an error that is not in the table becomes a `500` with a generic
message, so internals never leak to clients.

**No `NaN` or `Inf` in responses.** Go floating-point math never fails; it silently
yields `NaN` or `±Inf`, which JSON cannot represent. Every result goes through one
check that turns them into `ErrUndefined` / `ErrOverflow`, so the server never
returns a broken body. A result that is too *small* (such as `10^-400`) is not an
error: it rounds to `0`.

**Missing fields are detected.** The request fields are `*float64`, so a missing
field (`nil`) is distinguishable from a field sent as `0`. Without that,
`{"a": 5}` to `/divide` would report "division by zero" instead of "missing field: b".
Unknown fields are rejected, so a typo such as `"b2"` gives a clear error.

**`400` vs `422`.** A request that cannot be understood is `400`. A request that is
understood but cannot be carried out, such as dividing by zero, is `422`.

**Percentage.** `/percentage` returns `b` percent of `a`. It is computed as
`a * (b / 100)` rather than `(a * b) / 100`, which would overflow on large values
whose final result still fits.

**Middleware order.** `Logging → Recover → CORS → router`. Logging is outermost so it
also records the `500` produced by recovery; CORS sits next to the router so it can
answer browser preflight requests (`OPTIONS`), which the router would otherwise
reject with `405`.

**Production basics.** The server sets read, write and idle timeouts, limits the
request body to 1 MB, and shuts down gracefully on `SIGINT` / `SIGTERM`, finishing
in-flight requests. The Docker image is multi-stage (about 23 MB), runs as a
non-root user, has a health check, and uses the exec form of `ENTRYPOINT` so that
`docker stop` delivers the signal straight to the server.

## Testing

```bash
cd backend
go test ./...                                     # run all tests
go test ./... -race -coverprofile=coverage.out    # what CI runs
go tool cover -func=coverage.out                  # coverage per function
go tool cover -html=coverage.out                  # coverage report in the browser
```

Tests are table-driven and live next to the code (`*_test.go`). Coverage:

| Package | Coverage |
|---|---|
| `internal/calculator` | 100% |
| `internal/handler` | 100% |
| `internal/middleware` | 96.7% |
| `cmd/server` | `newHandler` is tested end to end; the process startup and shutdown code in `run` is not (see below) |

The `handler` tests include end-to-end tests that send real HTTP requests through the
router, and `cmd/server` tests the full middleware chain.

CI ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml)) runs the linter and the
tests on every pull request and publishes the coverage report as a build artifact.

### Linting

```bash
cd backend
golangci-lint run        # configuration: .golangci.yml
```

## Known limitations

- `404` and `405` responses come from Go's router and are plain text, not JSON. The
  frontend never calls unknown routes, so this was left as is.
- JSON field names are matched case-insensitively (standard `encoding/json`
  behavior), so `{"A": 1, "B": 2}` is accepted.
- `run` in `cmd/server` (listening, signals, shutdown) is verified by hand, not by
  automated tests, since it starts a real server and waits for signals.
