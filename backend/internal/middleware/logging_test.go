package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureLogs sends log output to the returned buffer for the duration of the
// test, and restores the previous logger after.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &buf
}

func TestLogging(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		target  string
		handler http.HandlerFunc
		want    []string
	}{
		{
			name:   "explicit status",
			method: http.MethodPost,
			target: "/divide",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnprocessableEntity)
			},
			want: []string{"msg=request", "method=POST", "path=/divide", "status=422", "duration="},
		},
		{
			name:   "body without explicit status defaults to 200",
			method: http.MethodGet,
			target: "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "ok")
			},
			want: []string{"method=GET", "path=/health", "status=200"},
		},
		{
			name:   "first status wins",
			method: http.MethodPost,
			target: "/add",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: []string{"status=404"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logs := captureLogs(t)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.target, nil)
			Logging(tc.handler).ServeHTTP(rec, req)

			for _, want := range tc.want {
				if !strings.Contains(logs.String(), want) {
					t.Errorf("log line %q does not contain %q", logs.String(), want)
				}
			}
		})
	}
}

// TestLogging_OmitsQueryString checks that query parameters, which may carry
// sensitive data, are not written to the log.
func TestLogging_OmitsQueryString(t *testing.T) {
	logs := captureLogs(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodGet, "/health?token=secret", nil)
	Logging(handler).ServeHTTP(httptest.NewRecorder(), req)

	if strings.Contains(logs.String(), "secret") {
		t.Errorf("log line %q contains the query string", logs.String())
	}
}

// TestLogging_PassesResponseThrough checks that logging does not alter what
// the client receives.
func TestLogging_PassesResponseThrough(t *testing.T) {
	captureLogs(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "kept")
		w.WriteHeader(http.StatusTeapot)
		io.WriteString(w, "tea")
	})

	rec := httptest.NewRecorder()
	Logging(handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %v, want %v", rec.Code, http.StatusTeapot)
	}

	if got := rec.Header().Get("X-Test"); got != "kept" {
		t.Errorf("X-Test header = %q, want %q", got, "kept")
	}

	if got := rec.Body.String(); got != "tea" {
		t.Errorf("body = %q, want %q", got, "tea")
	}
}
