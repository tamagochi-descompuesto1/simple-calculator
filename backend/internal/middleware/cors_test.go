package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testOrigin = "http://localhost:5173"

// trackedHandler is a handler that records whether it was called, to check that
// a middleware did or did not pass the request on.
type trackedHandler struct {
	called bool
}

func (h *trackedHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true
	_, _ = io.WriteString(w, "from handler")
}

func newPreflight() *http.Request {
	req := httptest.NewRequest(http.MethodOptions, "/add", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	return req
}

func TestCORS_Preflight(t *testing.T) {
	next := &trackedHandler{}
	rec := httptest.NewRecorder()

	CORS(testOrigin)(next).ServeHTTP(rec, newPreflight())

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %v, want %v", rec.Code, http.StatusNoContent)
	}

	if next.called {
		t.Error("preflight reached the next handler, want it answered by CORS")
	}

	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}

	wantHeaders := map[string]string{
		"Access-Control-Allow-Origin":  testOrigin,
		"Access-Control-Allow-Methods": "POST",
		"Access-Control-Allow-Headers": "Content-Type",
	}
	for name, want := range wantHeaders {
		if got := rec.Header().Get(name); !strings.Contains(got, want) {
			t.Errorf("%s = %q, want it to contain %q", name, got, want)
		}
	}
}

func TestCORS_ActualRequest(t *testing.T) {
	next := &trackedHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/add", nil)
	req.Header.Set("Origin", testOrigin)

	CORS(testOrigin)(next).ServeHTTP(rec, req)

	if !next.called {
		t.Error("request did not reach the next handler")
	}

	if got := rec.Body.String(); got != "from handler" {
		t.Errorf("body = %q, want %q", got, "from handler")
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}

	if got := rec.Header().Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("Vary = %q, want it to contain %q", got, "Origin")
	}
}

// TestCORS_PlainOptions checks that an OPTIONS request that is not a preflight
// (no Access-Control-Request-Method) goes on to the next handler.
func TestCORS_PlainOptions(t *testing.T) {
	next := &trackedHandler{}
	req := httptest.NewRequest(http.MethodOptions, "/add", nil)

	CORS(testOrigin)(next).ServeHTTP(httptest.NewRecorder(), req)

	if !next.called {
		t.Error("plain OPTIONS request did not reach the next handler")
	}
}

// TestCORS_UsesConfiguredOrigin checks that the allowed origin comes from the
// argument and is not a fixed value.
func TestCORS_UsesConfiguredOrigin(t *testing.T) {
	tests := []string{testOrigin, "https://calculator.example.com"}

	for _, origin := range tests {
		t.Run(origin, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)

			CORS(origin)(&trackedHandler{}).ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, origin)
			}
		})
	}
}
