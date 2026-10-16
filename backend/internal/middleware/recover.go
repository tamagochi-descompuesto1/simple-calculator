// Package middleware holds the HTTP middleware that wraps the API router:
// panic recovery, request logging and CORS.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// internalErrorBody is the fixed JSON sent when a handler panics. It is written
// here instead of reusing the handler package's helpers, which are private.
const internalErrorBody = `{"error":"internal error"}`

// Recover turns a panic in next into a 500 JSON response, so the client always
// gets an answer instead of a dropped connection. The panic and its stack trace
// are logged; nothing about them is sent to the client.
//
// If next already wrote part of the response before panicking, the status line
// is already sent and cannot be changed; the 500 only takes effect otherwise.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A deferred function still runs when next panics. recover() returns the
		// value passed to panic, or nil if there was no panic, and it only works
		// when called directly from a deferred function.
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered",
					"panic", recovered,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(internalErrorBody)); err != nil {
					slog.Error("error writing panic response", "error", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
