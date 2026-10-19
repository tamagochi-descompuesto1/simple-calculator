# Prompts and working method

This project was built in a pair-programming style with Claude Code (Anthropic's
coding assistant) inside VS Code. This file shows how it was used: the initial
prompt, the way of working, a selection of representative prompts, and a
summary of who wrote what.

The session was held in Spanish, so the prompts below are **translated into
English**. Most of the roughly 130
prompts of the session were short and conversational (for example "now let's continue with
the middleware" or "can you check this file?"), so listing all of them would not
say much.

## The initial prompt (translated from Spanish)

The first prompt set the context and the rules for the whole session. In short:
a technical assignment for a full-stack calculator (React + TypeScript frontend,
Go REST backend), five working days, about two to four hours of expected effort;
a monorepo layout and design decisions taken after requirements revission (one endpoint
per operation, layered backend, structured form instead of free-text
expressions, Vitest and Go's `testing`, GitHub Actions, multi-stage Docker); the
state of the work so far; and the role asked of the assistant: a senior
pair-programming guide who explains concepts as they appear, proposes design
and asks questions step by step, and prioritizes correctness, clarity and
maintainability over extra features.

It is reproduced below in full:

```text
I'm doing a technical test. I have 5 business days; it should take about 2-4 hours. I'm going to use Claude Code to build the project with me.

## Assignment goal

Build a full-stack calculator application with a React frontend and a backend microservice. The frontend should consume the backend API to perform basic and advanced arithmetic operations. Focus on clean design, maintainable code, and testable architecture.

Functional requirements:
- Operations: Addition, Subtraction, Multiplication, Division. Optional: Exponentiation, Square Root, Percentage
- Frontend (React): intuitive UI, input validation and error handling, responsive (basic mobile support)
- Backend (REST API): endpoints for the operations, validate input and edge cases (division by zero, invalid data), respond in JSON

Non-functional:
- Clean, readable, idiomatic code (frontend and backend)
- Unit tests covering key functionality in both layers
- Documentation: setup instructions, API usage, design rationale
- Optional: Dockerfile for full-stack deployment

Constraints:
- Frontend: React with TypeScript preferred
- Backend: Go preferred

Deliverables:
- Git repo with frontend + backend
- README with setup, API examples, design decisions
- Unit tests + coverage report
- Optional: Dockerfile to run frontend+backend together
- Share any prompts used to create the project

## Decisions taken

**Repo structure (monorepo):**

calculator-app/
├── backend/
│ ├── cmd/server/main.go # entry point, starts the server
│ ├── internal/handler/ # HTTP handlers (parsing, status codes)
│ ├── internal/calculator/ # pure operation logic (knows nothing about HTTP)
│ ├── internal/middleware/ # logging, recover, CORS
│ ├── go.mod / go.sum
├── frontend/ # React + TypeScript (Vite)
├── .github/workflows/ # CI with GitHub Actions
├── docker-compose.yml
└── README.md


**Backend:**
- Go with the native `net/http` router (Go 1.22+, it already supports patterns like `"POST /add"` without needing a framework). Possibly migrate to `chi` later if chained middleware is needed, but start without it.
- I already have Go installed (1.27.2) and ran `go mod init github.com/tamagochi-descompuesto1/simple-calculator/backend` inside `backend/`.
- Layered architecture: `internal/calculator` = pure functions testable without HTTP (e.g. `func Add(a, b float64) float64`, `func Divide(a, b float64) (float64, error)`). `internal/handler` = parses the request JSON, calls `calculator`, formats the response and status codes.
- **API design decided**: one endpoint PER operation (RESTful), NOT a generic endpoint with `{"operation": "add", ...}`, and NO support for expressions like `numbers[]+operations[]` (out of scope, it's expression parsing, not what the brief asks for).
  - `POST /add` `{a, b}` → `{result}`
  - `POST /subtract` `{a, b}`
  - `POST /multiply` `{a, b}`
  - `POST /divide` `{a, b}` (reject b=0 with 400 and an error JSON `{"error": "division by zero"}`)
  - `POST /power` `{a, b}` (exponentiation)
  - `POST /sqrt` `{a}` (unary, reject negatives)
  - `POST /percentage` `{a, b}` → interpreted as "b% of a" = `(a*b)/100`
  - `GET /health`
- Generic handler pattern like `handler.Operate(calculator.Add)` so the parsing/response boilerplate is not repeated on every route.
- Validation in two layers: client (immediate feedback) + backend (the real source of truth, never trust the client alone).
- Tests: Go convention, NO separate `tests/` folder — they go next to the code with the `_test.go` suffix in the same package (e.g. `operations.go` + `operations_test.go` inside `internal/calculator`). Use native `testing` (not yet decided whether to add `testify` for more readable asserts). Run with `go test ./... -cover`.

**Frontend:**
- React + TypeScript with Vite (not Create React App).
- UI design: NOT a free-text input like a real calculator (`2+3*5`), to avoid having to parse/validate syntax. Instead: structured state with a `<select>`/buttons to choose the operation, input `a` always visible, input `b` hidden/disabled if the operation is unary (sqrt).
- A separate `apiClient.ts` that encapsulates the `fetch` calls to the backend, so it can be mocked in tests.
- Input validation (they are strings while the user types, not a number directly): an `isValidNumber`-style function, "Calculate" button disabled if invalid, inline error messages, and also show the errors the backend returns.
- Responsive: Flexbox/Grid + a basic media query for mobile. No heavy UI library (consider Tailwind only if it helps look polished quickly).
- Tests: Vitest + React Testing Library (the standard with Vite).

**CI/CD:**
- GitHub Actions with two parallel jobs: backend (`go test ./... -cover`, `golangci-lint run`) and frontend (`npm ci`, `npm run lint`, `npm run test -- --coverage`). Optional: a job to build the Docker images.

**Docker:**
- Multi-stage Dockerfile for the backend (build with `golang:1.27-alpine`, runtime on `alpine:latest`) so the evaluator does not need Go installed or to worry about versions — this is a design decision to document in the README.
- I'm on Linux Mint, which is why Docker matters for guaranteeing reproducibility in the evaluation.

## Where I left off

I already have:
- Go 1.27.2 installed and working
- `backend/` with `go.mod` initialized
- `internal/calculator/` with simple functions for the operations.

## What I need you to do

I want you to pick this up from here, acting as my senior pair programmer. Help me:
1. Refine `internal/calculator` with its pure functions and tests
3. Refine `internal/handler` with the generic `Operate()` pattern
4. Wire everything together in `main.go`
5. Set up the frontend with Vite + TS
6. Build everything with tests, CI, Docker and the README at the end

Keep proposing and asking about any design matter; I prefer to validate step by step. Write the least amount of code, leave that to me; just serve as a guide so the project meets what matters most, which is prioritizing the correctness, clarity and maintainability of the project over any extra feature.
```

## How we worked

After the initial prompt, the work followed the same loop for every piece of the
project (a package, a component, a CI job):

1. **The author says what comes next, or what they are unsure about.** Usually in
   one or two lines.
2. **The assistant proposes a design and asks for decisions**, giving a
   recommendation and the trade-off of each option, and explains any new concept in plain terms.
3. **The author decides.** The decisions were theirs and several changed the
   proposal (for example the error-to-status table, Tailwind and a retro theme
   taken from their portfolio, keeping extra features out).
4. **The code is written and run.** In some cases by the author alone, following the
   guidance; other times the assistant also wrote parts when asked (see below).
5. **The assistant reviews.** It reads the files, runs the formatter, linter,
   type-checker and tests, and reports what is wrong, ordered by importance.
   Several real bugs were found this way (a missing `return` in a component, a
   regular expression with a look-alike Unicode character, a server that exited
   silently with code 0 when the port was busy, a missing `.nvmrc` in the CI).
6. **Tests are checked, not just run.** A test that has never failed is not
   proven, so important tests were verified by breaking the code on purpose and
   confirming that the right test fails. The CI jobs were verified the same way.
7. **Small steps.** Each piece lived in its own short-lived branch and pull
   request, with CI checks that must pass before merging into `main`.

Ground rules the author set along the way:

- Most of the time, the assistant was a guide and should not write the code.
  When necessary the author explicitly asked it to
  implement a piece of code or to write its tests.
- Explanations had to be concrete and direct.
  When a reply was judged too abstract, the style changed to explain step by step.
- Correctness, maintainability and cleanliness came before any extra feature.
- Documentation and the trimming of code comments were left for the end.

## Representative prompts (translated from Spanish)

Each one comes with a line of context.

**A learning question. Go isn't the author's main programming language, so some prompts asked specifications about the language.**

```text
Is there a way to do docstrings in Go? What format is used? Do the sentinels also have a format, or the package comment?
```

**A design discussion: the author reacts to a proposal, extends it with their own idea, and records numbered decisions.**

```text
Ok, I made the corrections you pointed out and you're completely right about what you told me. But then we have to extend the list of errors, not only to ErrDivisionByZero and ErrNegativeSqrt but also to an ErrNotFinite, right? To raise an error on any math.IsNaN and math.IsInf. But how do we map this to an error that is easy for the user to understand? Telling them the result is not finite is confusing; I think something like "The result cannot be represented" or something like that.
Answering to your questions:
1. Completely agree
2. Let's go with A
3. I prefer it in the calculator
4. Let's stay with testing, if it gets bigger we'll use testify
I just fixed the package into a single operations.go but I need you to help me checking the validations and the error types
```

**Pushing back on an implementation. This idea replaced a `switch` with a table that maps each error to its HTTP status.**

```text
In this case for the switch cases, wouldn't it be better to have a set type of the sentinels, and if any error is in those sentinels then we throw the custom writeError? And then for maintenance we would only add another error to this set of sentinels, or we could make another set of sentinels that correspond to another type of HTTP code
```

**Steering how the assistant works: a request for simpler explanations, and permission to write code.**

```text
Sometimes I don't understand your way of explaining things, it's very ambiguous. I need you to go one step further, as if I had just started learning. Help me implementing it now please, write the tests, and then we'll continue with the main.go file
```

**Scope control. The author repeatedly kept extra features out.**

```text
Answering your questions:
1. Keypad, it's equally intuitive, right? Because it's one of the things they asked for
2. @fontsource is better
3. Nothing extra, what we have is enough, let's prioritize correctness, maintainability and cleanliness before any extra feature
Let's start the feat/frontend-theme branch
```

**A usability doubt from the author, which led to reordering the form (number, operation, number) and naming the fields by meaning.**

```text
Hey, I have doubts about the design, I don't know how intuitive it is, maybe what if we add an (i) or (?) icon with a small dropdown window that gives some info, it's that sometimes I feel it's more intuitive number, operation, number, rather than operation, number, number
```

**A question about how to organize the documentation, and whether to trim code comments.**

```text
I want to ask you something about the documentation. I know you already did the backend one, I just kept wondering what would be best. Do we put a docs/ folder with detailed documentation? do we leave everything in the back and front READMEs? What do we do with the root README? I was also thinking that if we already have everything in the README, maybe it's time to lighten the code comments, I feel they're a bit loaded.
```

**A process question about protecting the `main` branch in GitHub.**

```text
Hey, I'm setting up the ruleset now, do I add actors that can bypass? It comes with the block force pushes and restrict deletions rules, do I leave them? What settings do I put in require status checks to pass?
```

## Who wrote what

This is an honest summary of who wrote what code. In every case the author reviewed it, ran it and made the design decisions.

| Area                 | Written by the author (with guidance and review)                                                                     | Written by the assistant (at the author's request)                                                        |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Backend `calculator` | The operations and most test cases                                                                                   | Corrections to the tests, and the percentage tests                                                        |
| Backend `handler`    | Response helpers, the first version of the error mapper, the router, `Operate`, `OperateUnary` and request decoding, | The error table and the handler and router tests                                                          |
| Backend server       | `main.go`, graceful shutdown, recover, logging and CORS middleware                                                   | The refactor into `run`, the recover, logging and CORS middleware tests                                   |
| Backend tooling      | The CI workflow and the backend Dockerfile                                                                           | `golangci-lint` configuration and the code fixes it asked for                                             |
| Frontend base        | Project setup, Tailwind, Vitest, Prettier, design tokens                                                             | -                                                                                                         |
| Frontend logic       | API client, operations list, number parsing, theme hook, and their tests                                             | Corrections to the API client and to some tests                                                           |
| Frontend components  | `Window` and `ThemeToggle` (and the `Window` tests), `App`, `Calculator`, `NumberField`, `OperationPicker`           | `ThemeToggle` tests, corrections to `App`, `Calculator`, `NumberField`, `OperationPicker` and their tests |
| Docker and CI        | The frontend and Docker CI jobs, with review, frontend Dockerfile                                                    | Fixes to the frontend Dockerfile, `docker-compose.yml`, and fixes to the CI test script                   |
| Documentation        |                                                                                                                      | The three READMEs and this file, from the author's decisions                                              |

The assistant also ran the tests, linters and builds, and ran the containers
locally to verify its own changes. It did not deploy anything.
