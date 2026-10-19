package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/calculator"
)

// successBody is {"result": n}. No omitempty: a result of 0 must still appear.
type successBody struct {
	Result float64 `json:"result"`
}

// errorBody is {"error": "message"}.
type errorBody struct {
	Error string `json:"error"`
}

// healthBody is {"status": "ok"}.
type healthBody struct {
	Status string `json:"status"`
}

// writeJSON writes v as JSON with the given status. The headers must be set
// before WriteHeader, or they are ignored.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// The status is already sent, so an encoding failure can only be logged.
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error serializing JSON", "error", err)
	}
}

// writeError writes {"error": msg}. It is the only place that knows that shape.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// operationErrors maps each calculator error to its HTTP status. Add a row to
// expose a new error; unlisted ones become a 500. The first match wins.
var operationErrors = []struct {
	err    error
	status int
}{
	{calculator.ErrDivisionByZero, http.StatusUnprocessableEntity},
	{calculator.ErrNegativeSqrt, http.StatusUnprocessableEntity},
	{calculator.ErrOverflow, http.StatusUnprocessableEntity},
	{calculator.ErrUndefined, http.StatusUnprocessableEntity},
}

// writeOperationError answers with the error's own message if it is listed in
// [operationErrors]. Otherwise it logs the error and sends a generic 500, so
// internals do not leak.
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
