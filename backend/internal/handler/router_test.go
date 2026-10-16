package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// send makes a real HTTP request to the test server and returns the response
// together with its body as text.
func send(t *testing.T, srv *httptest.Server, method, path, body string) (*http.Response, string) {
	t.Helper()

	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("building %s %s request: %v", method, path, err)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("sending %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response of %s %s: %v", method, path, err)
	}

	return resp, strings.TrimSpace(string(data))
}

// TestRouter sends real requests through the router. Each operation route gets
// operands for which every other operation would give a different result, so a
// route wired to the wrong operation makes the test fail.
func TestRouter(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	t.Cleanup(srv.Close)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "add", method: http.MethodPost, path: "/add", body: `{"a":8,"b":2}`, wantStatus: http.StatusOK, wantBody: `{"result":10}`},
		{name: "subtract", method: http.MethodPost, path: "/subtract", body: `{"a":8,"b":2}`, wantStatus: http.StatusOK, wantBody: `{"result":6}`},
		{name: "multiply", method: http.MethodPost, path: "/multiply", body: `{"a":8,"b":2}`, wantStatus: http.StatusOK, wantBody: `{"result":16}`},
		{name: "divide", method: http.MethodPost, path: "/divide", body: `{"a":8,"b":2}`, wantStatus: http.StatusOK, wantBody: `{"result":4}`},
		{name: "power", method: http.MethodPost, path: "/power", body: `{"a":8,"b":2}`, wantStatus: http.StatusOK, wantBody: `{"result":64}`},
		{name: "percentage", method: http.MethodPost, path: "/percentage", body: `{"a":200,"b":15}`, wantStatus: http.StatusOK, wantBody: `{"result":30}`},
		{name: "sqrt", method: http.MethodPost, path: "/sqrt", body: `{"a":16}`, wantStatus: http.StatusOK, wantBody: `{"result":4}`},
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},

		// Errors travel the whole way back as JSON.
		{name: "domain error", method: http.MethodPost, path: "/divide", body: `{"a":5,"b":0}`, wantStatus: http.StatusUnprocessableEntity, wantBody: `{"error":"division by zero"}`},
		{name: "unary domain error", method: http.MethodPost, path: "/sqrt", body: `{"a":-4}`, wantStatus: http.StatusUnprocessableEntity, wantBody: `{"error":"square root of a negative number"}`},
		{name: "malformed request", method: http.MethodPost, path: "/add", body: `{"a":1}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: b"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, got := send(t, srv, tc.method, tc.path, tc.body)

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("%s %s status = %v, want %v", tc.method, tc.path, resp.StatusCode, tc.wantStatus)
			}

			if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("%s %s Content-Type = %q, want %q", tc.method, tc.path, contentType, "application/json")
			}

			if got != tc.wantBody {
				t.Errorf("%s %s body = %v, want %v", tc.method, tc.path, got, tc.wantBody)
			}
		})
	}
}

// TestRouter_UnmatchedRequests covers requests the router rejects by itself.
// Only the status is checked: those responses are plain text, not JSON.
func TestRouter_UnmatchedRequests(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	t.Cleanup(srv.Close)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "unknown route", method: http.MethodPost, path: "/modulo", wantStatus: http.StatusNotFound},
		{name: "root", method: http.MethodGet, path: "/", wantStatus: http.StatusNotFound},
		{name: "get on an operation", method: http.MethodGet, path: "/add", wantStatus: http.StatusMethodNotAllowed},
		{name: "put on an operation", method: http.MethodPut, path: "/divide", wantStatus: http.StatusMethodNotAllowed},
		{name: "post on health", method: http.MethodPost, path: "/health", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := send(t, srv, tc.method, tc.path, "")

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("%s %s status = %v, want %v", tc.method, tc.path, resp.StatusCode, tc.wantStatus)
			}
		})
	}
}
