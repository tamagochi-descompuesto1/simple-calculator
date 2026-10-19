# Simple Calculator

A full-stack calculator: a React + TypeScript frontend that sends each operation
to a Go REST API. It supports addition, subtraction, multiplication, division,
power, square root and percentage, validates input on both sides, and works on
phones and in light and dark themes.

## Quick start

With [Docker](https://docs.docker.com/get-docker/) (Compose v2) you do not need
Go or Node installed:

```bash
docker compose up --build
```

Then open <http://localhost:3000>. Stop it with `Ctrl+C` and
`docker compose down`.

If port 3000 or 8080 is already in use, pick others (see
[`.env.example`](.env.example)):

```bash
FRONTEND_PORT=3001 BACKEND_PORT=8081 docker compose up --build
```

### Without Docker

Each part has its own instructions and needs Go 1.27 or Node 24 respectively:

- Backend: [`backend/README.md`](backend/README.md) (`go run ./cmd/server`, port 8080)
- Frontend: [`frontend/README.md`](frontend/README.md) (`npm run dev`, port 5173)

## How it works

```
 Browser ──── loads the page ────►  frontend   nginx serving static files  :3000
    │
    └──── POST /add, /divide … ──►  backend    Go REST API (JSON)           :8080
```

The browser loads the page from the frontend and then calls the API directly;
the frontend container never talks to the backend. For that reason the API
address is a build-time setting of the frontend (`VITE_API_URL`), and the
backend has to allow the frontend's origin (`CORS_ALLOWED_ORIGIN`). Compose sets
both consistently from the two ports.

The API has one `POST` endpoint per operation (`/add`, `/subtract`, `/multiply`,
`/divide`, `/power`, `/percentage`, `/sqrt`) plus `GET /health`. Request and
response examples are in [`backend/README.md`](backend/README.md#api).

## Repository layout

```
.
├── backend/             Go API: calculator rules, HTTP handlers, middleware, Dockerfile
├── frontend/            React app: form, theme, API client, Dockerfile
├── docker-compose.yml   runs both services
├── .github/workflows/   CI
└── .env.example         configuration reference
```

## Design decisions

These were taken at the start of the project and refined while building it. Each
part's README goes into more detail about its own:
[backend](backend/README.md#design-decisions) and
[frontend](frontend/README.md#design-decisions).

The guiding principle was **correctness, clarity and maintainability before any
extra feature**: small pieces with one job, rules written once and tested where
they live, and nothing built that the brief does not ask for (no expression
parser, no history, no deployment).

### Technology choices

| Choice                                                                     | Why                                                                                                                                                                                                                                | Considered instead                                                                   |
| -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| **Go** for the backend                                                     | Preferred by the brief. Static typing and explicit error values suit a small API with strict rules, and it compiles to one static binary, so the Docker image is about 23 MB.                                                      |                                                                                      |
| **`net/http` router, no framework**                                        | Since Go 1.22 the standard router understands patterns such as `POST /add` and answers `405` for the wrong method by itself. Middleware are plain `func(http.Handler) http.Handler`, so a framework adds nothing for eight routes. | `chi`, kept as the way out if routing ever grows                                     |
| **Standard library for the rest** (`encoding/json`, `log/slog`, `testing`) | No third-party dependencies, so nothing to audit or keep updated.                                                                                                                                                                  | `testify`: table-driven `testing` reads well enough at this size                     |
| **React + TypeScript + Vite** for the frontend                             | Preferred by the brief. Types let the list of operations define the valid operation ids, so a typo does not compile. Vite is fast and is the current way to start a React app.                                                     | Create React App, which is deprecated                                                |
| **Tailwind CSS v4 + CSS variables**                                        | A polished, themeable interface without a heavy UI library. Colors, shadows and fonts are defined once as tokens, and the light and dark themes come from the same variables.                                                      | A component library (more weight, less control)                                      |
| **Fonts bundled with `@fontsource`**                                       | No dependency on an external font service; works offline and inside Docker.                                                                                                                                                        | A Google Fonts link                                                                  |
| **Vitest + Testing Library**                                               | Vitest shares Vite's configuration. Testing Library queries by role and label, the way a user finds things, which also keeps accessibility honest.                                                                                 | Jest                                                                                 |
| **ESLint + Prettier**                                                      | Standard linting for React and TypeScript, and a single formatter so style is never debated.                                                                                                                                       | Oxlint: faster, but younger and with fewer rules; speed does not matter at this size |
| **GitHub Actions**                                                         | Runs where the code lives. Three parallel jobs (backend, frontend, docker) keep feedback fast. Language and tool versions are read from `go.mod` and `.nvmrc` or pinned.                                                           |                                                                                      |
| **Docker + Compose**                                                       | One command runs everything on a machine with neither Go nor Node, which is what makes the project reproducible for a reviewer. Multi-stage builds keep the build tools out of the final images.                                   | Documenting only a manual setup                                                      |
| **A monorepo**                                                             | One repository, one CI and one Compose file; a change that touches the API and the interface is a single pull request.                                                                                                             | Two repositories                                                                     |

### Backend

- **Layers with one-way dependencies.** `calculator` (pure functions, no HTTP)
  ← `handler` (JSON, status codes, routes) ← `cmd/server` (configuration,
  middleware, lifecycle). The math rules are testable without HTTP, and the HTTP
  layer can change without touching them.
- **One endpoint per operation.** `POST /divide` maps to `calculator.Divide`,
  which is easy to read, document and test. Rejected: a generic
  `{"operation": "add"}` endpoint (it needs a dispatcher and the route stops
  documenting itself) and an endpoint that takes expressions like `2+3*5`
  (it means writing a parser the brief does not ask for).
- **A handler factory.** `Operate(op)` and `OperateUnary(op)` build the handler
  for any operation of the right shape, so decoding, validation and responses are
  written once. The operation is passed in, so tests can supply a fake one, which
  is the only practical way to test the unexpected-error path.
- **Domain rules live in `calculator`, not in the handler.** Division by zero, a
  negative square root, `0` to a negative power and results that cannot be
  represented are errors of the operation itself. The handler only checks that
  the request is well formed.
- **Errors are values, with a status policy.** Named errors (`ErrDivisionByZero`
  and others) are compared with `errors.Is` and mapped by one table to `422`: the
  request is valid but cannot be computed. `400` is for malformed requests. Anything
  unexpected is `500` with a generic message and the real error only in the log,
  so internals never reach the client. Exposing a new error is one table row.
- **Floating-point safety.** Go floats never fail; they silently produce `NaN` or
  `±Inf`, which JSON cannot carry. Every result goes through one check that turns
  them into errors. Percentage is computed as `a * (b / 100)` because
  `(a * b) / 100` can overflow for values whose result still fits.
- **Strict request decoding.** Operands are `*float64`, so a missing field is
  distinguishable from `0` (otherwise `{"a": 5}` sent to `/divide` would report a
  division by zero). Unknown fields, data after the JSON object and bodies over
  1 MB are rejected, and the decoder's messages, which mention Go internals, are
  replaced by short ones.
- **Middleware as small wrappers.** Panic recovery, request logging and CORS, in
  the order `Logging → Recover → CORS → router`: the `500` produced by recovery
  gets logged, and browser preflight requests are answered before the router would
  reject them with `405`.
- **CORS in the backend, not a development proxy.** One environment variable
  gives the same behavior in local development and in Compose, where the page and
  the API are on different origins.
- **Operability.** Server timeouts, graceful shutdown (so `docker stop` does not
  cut requests in flight), configuration through environment variables, and
  `GET /health` for health checks.

### Frontend

- **A structured form, not a free-text expression.** You type a number, pick an
  operation and type a number. That avoids parsing syntax and means every request
  is something the API already understands. The order reads like written
  arithmetic, and fields are named by meaning where the order matters (base and
  exponent, amount and percent).
- **One list drives the interface.** The keypad, the form labels and the API
  client types all come from `operations.ts`, so adding an operation starts there.
- **Numbers stay as text while typing.** States such as `5.` or `-` are normal
  while typing and are not numbers yet. They are parsed on every render instead of
  stored twice, only plain decimals are accepted, and values that overflow are
  rejected.
- **Validation in two places.** The form gives instant feedback; the backend
  remains the source of truth and its messages are shown too.
- **Network access in one small module.** Components can be tested with it
  replaced, and failures become one error type whose message is safe to show.
- **Results are rounded for display only.** Fifteen significant digits hide
  floating-point noise (`0.1 + 0.2` shows `0.3`) without cutting off exact
  integers. The API always returns the exact value.
- **Errors appear inline,** next to the field or in the result panel, instead of
  in a toast: a calculator has one place the user is looking at.
- **Accessibility from the start.** The keypad is a labelled radio group, errors
  are linked to their fields, and results and errors are announced.
- **Themes from one set of variables.** The page follows the system theme, a
  choice is saved only when the user clicks the toggle, and a small inline script
  applies it before the first paint to avoid a flash.

### Build, test and delivery

- **Configuration through environment variables,** documented in
  [`.env.example`](.env.example). One caveat: the frontend's `VITE_API_URL` is
  embedded when it is built, so in Docker it is a build argument, and it must be an
  address the **browser** can reach (not a name only containers know).
- **Compose keeps the settings consistent.** Changing `FRONTEND_PORT` or
  `BACKEND_PORT` changes the published ports, the API address and the allowed CORS
  origin together, so one cannot be left out of sync with the others.
- **Tests at every level.** Table-driven unit tests where the rules live,
  `httptest` tests of the handlers and of the router with real requests, a test of
  the whole middleware chain, component tests with the API mocked, and a CI smoke
  test that starts the containers and runs a real calculation including the CORS
  preflight. Important tests were also verified by breaking the code on purpose.
- **Small, safe containers.** Multi-stage builds, non-root users, health checks,
  pinned base images, and the exec form of `ENTRYPOINT` so the shutdown signal
  reaches the server.
- **Quality gates.** Lint, format check, type-check, tests with coverage and the
  Docker smoke test run on every pull request, and `main` only accepts changes
  through pull requests.
- **No continuous deployment.** There is no target environment and the project is
  meant to be run locally. The CI proves that the images build and work together,
  which is what a reviewer needs.

## Testing

```bash
cd backend && go test ./... -race -cover      # backend
cd frontend && npm install && npm test        # frontend
```

| Part                    | Coverage                                                                           |
| ----------------------- | ---------------------------------------------------------------------------------- |
| Backend `calculator`    | 100%                                                                               |
| Backend `handler`       | 100%                                                                               |
| Backend `middleware`    | 96.7%                                                                              |
| Backend `cmd/server`    | 16.1% (the HTTP chain is tested; process startup and shutdown are checked by hand) |
| Frontend (all of `src`) | 100% of statements, functions and lines; 97% of branches                           |

Tests are table-driven in the backend (`go test`) and use Vitest with Testing
Library in the frontend. Coverage reports are generated by the CI and attached to
each run as artifacts; the commands to produce them locally are in each part's
README.

## Continuous integration

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs three jobs in
parallel on every pull request:

| Job        | What it checks                                                                                                                                |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `backend`  | `golangci-lint` (includes formatting and `go vet`), tests with `-race`, coverage report                                                       |
| `frontend` | ESLint, Prettier, type-check and build, tests with coverage                                                                                   |
| `docker`   | Builds both images with Compose, starts them, and checks that the page loads, the API answers a real calculation and CORS allows the frontend |

`main` is protected: changes go through pull requests with these checks passing.

## Known limitations

- There are no browser-level end-to-end tests. The Compose job checks the
  running containers over HTTP, and the interface is covered by component tests.
- `404` and `405` answers come from Go's router and are plain text, not JSON.
- The images are built in CI but not published to a registry, and there is no
  deployment: it is meant to be run locally.
- The theme follows the system only when the page loads.

## How this was built

I built this with Claude Code as a pair-programming partner. The prompts I used
are in [`PROMPTS.md`](PROMPTS.md).
