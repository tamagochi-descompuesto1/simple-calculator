// Package middleware holds the HTTP middleware that wraps the router: panic
// recovery, request logging and CORS.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// internalErrorBody is the fixed 500 body. The handler package's helpers are
// private, so it is written here.
const internalErrorBody = `{"error":"internal error"}`

// Recover turns a panic in next into a JSON 500 and logs it with its stack
// trace. If next had already started the response, the status can no longer
// change.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
