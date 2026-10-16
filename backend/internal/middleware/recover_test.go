package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// silenceLogs discards log output for the duration of the test, so expected
// error logs do not look like failures, and restores the previous logger after.
func silenceLogs(t *testing.T) {
	t.Helper()

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
}

func TestRecover_Panic(t *testing.T) {
	silenceLogs(t)

	tests := []struct {
		name       string
		panicValue any
	}{
		{name: "string panic", panicValue: "boom"},
		{name: "error panic", panicValue: io.ErrUnexpectedEOF},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tc.panicValue)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/add", nil)

			// If Recover fails to catch the panic, it reaches the test and fails it.
			Recover(panicking).ServeHTTP(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %v, want %v", rec.Code, http.StatusInternalServerError)
			}

			if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			if got := strings.TrimSpace(rec.Body.String()); got != internalErrorBody {
				t.Errorf("body = %v, want %v", got, internalErrorBody)
			}
		})
	}
}

func TestRecover_NoPanic(t *testing.T) {
	healthy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "kept")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "hello")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Recover(healthy).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %v, want %v", rec.Code, http.StatusCreated)
	}

	if got := rec.Header().Get("X-Test"); got != "kept" {
		t.Errorf("X-Test header = %q, want %q", got, "kept")
	}

	if got := rec.Body.String(); got != "hello" {
		t.Errorf("body = %q, want %q", got, "hello")
	}
}

// TestRecover_DoesNotLeakPanicValue checks that the panic message never reaches
// the client, since it may contain internal details.
func TestRecover_DoesNotLeakPanicValue(t *testing.T) {
	silenceLogs(t)

	secret := "database password is hunter2"
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(secret)
	})

	rec := httptest.NewRecorder()
	Recover(panicking).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if strings.Contains(rec.Body.String(), secret) {
		t.Errorf("response body %q leaks the panic value", rec.Body.String())
	}
}
