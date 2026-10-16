package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
