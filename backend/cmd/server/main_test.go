package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testOrigin = "http://localhost:5173"

// captureLogs sends log output to the returned buffer for the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &buf
}

// serve sends one request through newHandler.
func serve(origin string, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	newHandler(origin).ServeHTTP(rec, req)

	return rec
}

func post(path, body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
}

// TestNewHandler_Preflight proves CORS sits outside the router, which would reply 405.
func TestNewHandler_Preflight(t *testing.T) {
	captureLogs(t)

	req := httptest.NewRequest(http.MethodOptions, "/add", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	rec := serve(testOrigin, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %v, want %v", rec.Code, http.StatusNoContent)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}
}

// TestNewHandler_Responses checks the API through the whole chain. Errors must
// carry the CORS header too, or the browser hides them from the frontend.
func TestNewHandler_Responses(t *testing.T) {
	captureLogs(t)

	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   string
	}{
		{name: "success", req: post("/add", `{"a":6,"b":3}`), wantStatus: http.StatusOK, wantBody: `{"result":9}`},
		{name: "health", req: httptest.NewRequest(http.MethodGet, "/health", nil), wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},
		{name: "domain error", req: post("/divide", `{"a":5,"b":0}`), wantStatus: http.StatusUnprocessableEntity, wantBody: `{"error":"division by zero"}`},
		{name: "malformed request", req: post("/add", `{"a":1}`), wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: b"}`},
		{name: "wrong method", req: httptest.NewRequest(http.MethodGet, "/add", nil), wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(testOrigin, tc.req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tc.wantStatus)
			}

			if tc.wantBody != "" {
				if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
					t.Errorf("body = %v, want %v", got, tc.wantBody)
				}
			}

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
			}
		})
	}
}

// TestNewHandler_UsesConfiguredOrigin checks that the origin argument is the one sent.
func TestNewHandler_UsesConfiguredOrigin(t *testing.T) {
	captureLogs(t)

	const origin = "https://calculator.example.com"
	rec := serve(origin, httptest.NewRequest(http.MethodGet, "/health", nil))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

// TestNewHandler_LogsRequests checks that Logging is part of the chain.
func TestNewHandler_LogsRequests(t *testing.T) {
	logs := captureLogs(t)

	serve(testOrigin, post("/divide", `{"a":5,"b":0}`))

	for _, want := range []string{"method=POST", "path=/divide", "status=422"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("log %q does not contain %q", logs.String(), want)
		}
	}
}
