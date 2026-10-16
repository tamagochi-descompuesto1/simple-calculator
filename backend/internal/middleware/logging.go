package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to remember the status code the
// handler writes. The ResponseWriter interface has no way to read it back, and
// the logging middleware needs it once the handler is done.
//
// It embeds http.ResponseWriter, so it gets all of its methods for free; only
// WriteHeader is redefined, to record the status before passing it on.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader records the status code and forwards it. Only the first call is
// recorded, because net/http ignores any later one: the first status is the
// one the client actually receives.
func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}

	s.ResponseWriter.WriteHeader(code)
}

// Logging writes one log line per request with its method, path, response
// status and duration. The query string is left out on purpose, since it may
// carry data that should not end up in logs.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// A handler that writes a body without calling WriteHeader gets a 200
		// from net/http, so 200 is the status to assume until told otherwise.
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}
