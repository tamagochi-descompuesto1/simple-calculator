# Calculator (frontend)

A single-page calculator built with React and TypeScript. It sends each
operation to the [backend API](../backend/README.md) and shows the result, with
light and dark themes and a layout that works on phones.

## Quick start

Requires Node.js 24 (the version in [`.nvmrc`](.nvmrc); `nvm use` picks it up).

```bash
cd frontend
cp .env.example .env     # optional, see Configuration
npm install
npm run dev              # http://localhost:5173
```

The app needs the backend running to calculate anything (by default at
`http://localhost:8080`; see the backend README).

## Configuration

| Variable       | Default                 | Meaning                                              |
| -------------- | ----------------------- | ---------------------------------------------------- |
| `VITE_API_URL` | `http://localhost:8080` | Address of the backend API, without a trailing slash |

Vite embeds `VITE_` variables **at build time**, so changing the value needs a
new build (or a restart of `npm run dev`). The backend must also allow this
app's origin through its `CORS_ALLOWED_ORIGIN` setting, which defaults to
`http://localhost:5173`.

## Scripts

| Command                 | What it does                                               |
| ----------------------- | ---------------------------------------------------------- |
| `npm run dev`           | Development server with hot reload                         |
| `npm run build`         | Type-checks the code and builds the app into `dist/`       |
| `npm run preview`       | Serves the production build locally                        |
| `npm test`              | Runs the tests once                                        |
| `npm run test:watch`    | Runs the tests again on every change                       |
| `npm run test:coverage` | Runs the tests and writes a coverage report to `coverage/` |
| `npm run lint`          | ESLint                                                     |
| `npm run format`        | Formats the code with Prettier                             |
| `npm run format:check`  | Fails if any file is not formatted (used by CI)            |

## Using the calculator

1. Type the first number.
2. Pick an operation from the keypad: add, subtract, multiply, divide, power,
   square root or percentage.
3. Type the second number (square root takes only one).
4. Press **Calculate**, or press Enter in a field.

The button stays disabled until the numbers are valid, an invalid number shows
an error under its field, and errors from the backend (such as dividing by
zero) appear in the result panel. Power names its fields _Base_ and _Exponent_,
and percentage _Amount_ and _Percent_ (15 percent of 200 is 30).

## Project layout

```
src/
├── api/
│   ├── operations.ts     the list of operations: single source for the UI and the client
│   └── apiClient.ts      fetch calls to the backend, and the ApiError type
├── components/
│   ├── Calculator.tsx    the form: state, validation, submit, result
│   ├── OperationPicker.tsx  the keypad of operations (a radio group)
│   ├── NumberField.tsx   a labelled text input with an inline error
│   ├── Window.tsx        the retro window frame
│   └── ThemeToggle.tsx   the light/dark button
├── hooks/
│   └── useTheme.ts       theme state, saved choice and system preference
├── lib/
│   └── numbers.ts        parsing of typed numbers and formatting of results
├── test/setup.ts         test environment setup
├── App.tsx               puts the pieces together
└── index.css             design tokens and Tailwind configuration
```

## Design decisions

**A structured form, not a free-text expression.** The user picks an operation
and fills in numbers instead of typing `2+3*5`. That avoids writing an
expression parser, and every request is something the API already understands.

**One list drives the interface.** `operations.ts` holds each operation's id
(which is also its backend route), name, symbol, whether it takes one operand,
and the labels of its fields. The keypad, the form and the API client are all
built from it, so adding an operation starts with one new entry there.

**Numbers stay text while typing.** Inputs are plain text with a decimal
keyboard on phones, because states such as `5.` or `-` are normal while typing
and are not numbers yet. `parseNumber` accepts plain decimals only (no
exponents, hex or thousands separators) and rejects values that overflow to
`Infinity`.

**Validation in two places.** The form checks the numbers right away so the
user gets feedback without waiting; the backend remains the source of truth
and its errors are shown as well.

**A small, mockable API client.** All network access is in `apiClient.ts`.
Failures become an `ApiError` whose message is safe to show: the backend's own
message, a generic one for unreadable responses, or "Cannot reach the server".
Components are tested with that module mocked.

**Results are rounded for display only.** `formatResult` keeps 15 significant
digits, which hides floating-point noise (`0.1 + 0.2` shows `0.3`) without
cutting off exact integers.

**Theme.** Colors are CSS variables defined once with `light-dark()`, so the
page follows the system theme by default and `data-theme` on `<html>` forces
one. A choice is saved only when the user clicks the toggle, so people who never
choose keep following their system. A short inline script in `index.html`
applies the saved theme before the first paint to avoid a flash of the wrong
one.

**Accessibility.** The keypad is a labelled radio group (arrow keys work),
every field has a label, errors are linked to their field, and results and
errors are announced (`role="status"` and `role="alert"`). Decorative symbols
and icons are hidden from screen readers.

**Styling.** Tailwind CSS v4 with design tokens in `index.css`. Fonts
(Space Grotesk, Inter and JetBrains Mono) are bundled with the app through
`@fontsource` packages, so it does not depend on an external font service.

## Testing

```bash
npm test                  # run once
npm run test:coverage     # with a coverage report in coverage/
```

Tests use [Vitest](https://vitest.dev) and
[Testing Library](https://testing-library.com/docs/react-testing-library/intro/),
with `user-event` to interact the way a person would. They cover the parsing and
formatting helpers, the API client (with `fetch` replaced), the theme hook, each
component, and the full form flow with the API client mocked.

| Area                         | Coverage       |
| ---------------------------- | -------------- |
| Statements, functions, lines | 100%           |
| Branches                     | 97%; see below |

The one uncovered branch is the default API address in `apiClient.ts`: which side
of `VITE_API_URL ?? 'http://localhost:8080'` runs depends on whether a local
`.env` exists, so a single run can only exercise one of them.

CI ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml)) runs the linter,
the format check, the build and the tests with coverage on every pull request,
and publishes the coverage report as a build artifact.

## Known limitations

- The theme follows the system only when the page loads; changing the system
  theme while the page is open has no effect until a reload.
- Responsive layout (the window width and the keypad columns) is not covered by
  automated tests, since the test environment has no real layout.
- Results with more than 15 significant digits are rounded when displayed.
