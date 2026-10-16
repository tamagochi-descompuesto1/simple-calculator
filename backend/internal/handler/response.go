package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/calculator"
)

// successBody is the JSON body of a successful operation: {"result": <number>}.
// Result must not use omitempty, otherwise a result of 0 would be dropped.
type successBody struct {
	Result float64 `json:"result"`
}

// errorBody is the JSON body of any failed request: {"error": "<message>"}.
type errorBody struct {
	Error string `json:"error"`
}

// healthBody is the JSON body of the health check: {"status": "ok"}.
type healthBody struct {
	Status string `json:"status"`
}

// writeJSON writes v as a JSON response with the given HTTP status code.
// The Content-Type header must be set before WriteHeader, because headers
// are ignored once the status line has been sent.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// The status is already sent at this point, so an encoding failure can
	// only be logged; it cannot be reported to the client.
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error serializing JSON", "error", err)
	}
}

// writeError writes msg as a JSON error response with the given HTTP status code.
// It is the only place that knows the {"error": "<message>"} shape, so changing
// the error format only requires touching this function.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// operationErrors maps every error the calculator can return to the HTTP status
// that reports it. To expose a new calculator error, add a row here; an error
// missing from this table is reported as a 500 and logged.
//
// The first matching row wins, so if one sentinel ever wraps another, the more
// specific one must come first.
var operationErrors = []struct {
	err    error
	status int
}{
	{calculator.ErrDivisionByZero, http.StatusUnprocessableEntity},
	{calculator.ErrNegativeSqrt, http.StatusUnprocessableEntity},
	{calculator.ErrOverflow, http.StatusUnprocessableEntity},
	{calculator.ErrUndefined, http.StatusUnprocessableEntity},
}

// writeOperationError reports an error returned by a calculator operation.
// Errors listed in [operationErrors] are sent to the client with their own
// message. Anything else is unexpected: the real error is only logged, and the
// client receives a generic message so no internal details leak.
func writeOperationError(w http.ResponseWriter, err error) {
	for _, known := range operationErrors {
		if errors.Is(err, known.err) {
			writeError(w, known.status, err.Error())
			return
		}
	}

	slog.Error("unexpected operation error", "error", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
