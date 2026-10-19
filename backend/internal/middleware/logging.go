package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder remembers the status code, which a ResponseWriter cannot
// report back.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader records the first status only; net/http ignores later calls.
func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}

	s.ResponseWriter.WriteHeader(code)
}

// Logging logs the method, path, status and duration of each request. The query
// string is left out because it may carry sensitive data.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// net/http answers 200 when the handler never calls WriteHeader.
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
