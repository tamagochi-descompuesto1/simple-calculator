package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/calculator"
)

// silenceLogs discards log output for the duration of the test, so expected
// error logs do not look like failures, and restores the previous logger after.
func silenceLogs(t *testing.T) {
	t.Helper()

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   any
		want   string
	}{
		{name: "int success", status: http.StatusOK, body: successBody{Result: 30}, want: `{"result":30}`},
		{name: "float success", status: http.StatusOK, body: successBody{Result: 1.3993}, want: `{"result":1.3993}`},
		// Regression: a result of 0 must not be dropped from the body (no omitempty).
		{name: "zero success", status: http.StatusOK, body: successBody{Result: 0}, want: `{"result":0}`},
		{name: "negative success", status: http.StatusOK, body: successBody{Result: -5}, want: `{"result":-5}`},
		{name: "domain error", status: http.StatusUnprocessableEntity, body: errorBody{Error: "division by zero"}, want: `{"error":"division by zero"}`},
		{name: "bad request error", status: http.StatusBadRequest, body: errorBody{Error: "invalid JSON body"}, want: `{"error":"invalid JSON body"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeJSON(rec, tc.status, tc.body)
			got := strings.TrimSpace(rec.Body.String())

			if rec.Code != tc.status {
				t.Errorf("writeJSON(%v, %v) status = %v, want %v", tc.status, tc.body, rec.Code, tc.status)
			}

			if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("writeJSON(%v, %v) Content-Type = %q, want %q", tc.status, tc.body, contentType, "application/json")
			}

			if got != tc.want {
				t.Errorf("writeJSON(%v, %v) body = %v, want %v", tc.status, tc.body, got, tc.want)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		msg    string
		want   string
	}{
		{name: "bad request", status: http.StatusBadRequest, msg: "invalid JSON body", want: `{"error":"invalid JSON body"}`},
		{name: "domain error", status: http.StatusUnprocessableEntity, msg: "division by zero", want: `{"error":"division by zero"}`},
		{name: "internal error", status: http.StatusInternalServerError, msg: "internal error", want: `{"error":"internal error"}`},
		// The encoder must escape quotes; building the JSON by hand would break here.
		{name: "message with quotes", status: http.StatusBadRequest, msg: `field "a" is required`, want: `{"error":"field \"a\" is required"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeError(rec, tc.status, tc.msg)
			got := strings.TrimSpace(rec.Body.String())

			if rec.Code != tc.status {
				t.Errorf("writeError(%v, %q) status = %v, want %v", tc.status, tc.msg, rec.Code, tc.status)
			}

			if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("writeError(%v, %q) Content-Type = %q, want %q", tc.status, tc.msg, contentType, "application/json")
			}

			if got != tc.want {
				t.Errorf("writeError(%v, %q) body = %v, want %v", tc.status, tc.msg, got, tc.want)
			}
		})
	}
}

// TestWriteJSON_EncodeFailure covers the case where the value cannot be
// serialized. NaN is not valid JSON, which is what calculator.checkResult
// protects against. The status line is already sent when encoding fails, so
// the client sees the requested status with an empty body.
func TestWriteJSON_EncodeFailure(t *testing.T) {
	silenceLogs(t)

	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, successBody{Result: math.NaN()})

	if rec.Code != http.StatusOK {
		t.Errorf("writeJSON(NaN) status = %v, want %v", rec.Code, http.StatusOK)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("writeJSON(NaN) body = %q, want empty", rec.Body.String())
	}
}

func TestWriteOperationError(t *testing.T) {
	silenceLogs(t)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{name: "division by zero", err: calculator.ErrDivisionByZero, wantStatus: http.StatusUnprocessableEntity, wantMsg: calculator.ErrDivisionByZero.Error()},
		{name: "negative sqrt", err: calculator.ErrNegativeSqrt, wantStatus: http.StatusUnprocessableEntity, wantMsg: calculator.ErrNegativeSqrt.Error()},
		{name: "overflow", err: calculator.ErrOverflow, wantStatus: http.StatusUnprocessableEntity, wantMsg: calculator.ErrOverflow.Error()},
		{name: "undefined", err: calculator.ErrUndefined, wantStatus: http.StatusUnprocessableEntity, wantMsg: calculator.ErrUndefined.Error()},
		// errors.Is must see through wrapping; comparing with == would fall to the 500 here.
		{name: "wrapped sentinel", err: fmt.Errorf("divide failed: %w", calculator.ErrDivisionByZero), wantStatus: http.StatusUnprocessableEntity, wantMsg: "divide failed: division by zero"},
		// The original message ("boom") must never reach the client.
		{name: "unknown error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantMsg: "internal error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeOperationError(rec, tc.err)

			if rec.Code != tc.wantStatus {
				t.Errorf("writeOperationError(%v) status = %v, want %v", tc.err, rec.Code, tc.wantStatus)
			}

			var got errorBody
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("writeOperationError(%v) body is not valid JSON: %v", tc.err, err)
			}
			if got.Error != tc.wantMsg {
				t.Errorf("writeOperationError(%v) message = %q, want %q", tc.err, got.Error, tc.wantMsg)
			}
		})
	}
}

// TestOperationErrorsTable guards the table itself: every row must carry a real
// error and a status in the 4xx range, so a typo cannot silently turn a domain
// error into a 500 or a success.
func TestOperationErrorsTable(t *testing.T) {
	for _, row := range operationErrors {
		if row.err == nil {
			t.Error("operationErrors has a row with a nil error")
		}
		if row.status < 400 || row.status > 499 {
			t.Errorf("operationErrors row for %v has status %v, want a 4xx code", row.err, row.status)
		}
	}
}

// serve sends a POST request with the given body to h and returns the response.
func serve(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func TestOperate(t *testing.T) {
	silenceLogs(t)

	failing := func(a, b float64) (float64, error) { return 0, errors.New("boom") }

	tests := []struct {
		name       string
		op         BinaryOp
		body       string
		wantStatus int
		wantBody   string
	}{
		// Successful requests.
		{name: "success", op: calculator.Add, body: `{"a":6,"b":3}`, wantStatus: http.StatusOK, wantBody: `{"result":9}`},
		{name: "decimal success", op: calculator.Add, body: `{"a":0.5,"b":0.25}`, wantStatus: http.StatusOK, wantBody: `{"result":0.75}`},
		{name: "result of zero", op: calculator.Add, body: `{"a":5,"b":-5}`, wantStatus: http.StatusOK, wantBody: `{"result":0}`},
		{name: "operand zero is not missing", op: calculator.Add, body: `{"a":5,"b":0}`, wantStatus: http.StatusOK, wantBody: `{"result":5}`},

		// The operation fails.
		{name: "domain error", op: calculator.Divide, body: `{"a":5,"b":0}`, wantStatus: http.StatusUnprocessableEntity, wantBody: `{"error":"division by zero"}`},
		{name: "unexpected error", op: failing, body: `{"a":1,"b":2}`, wantStatus: http.StatusInternalServerError, wantBody: `{"error":"internal error"}`},

		// Malformed requests never reach the operation.
		{name: "empty body", op: calculator.Add, body: ``, wantStatus: http.StatusBadRequest, wantBody: `{"error":"request body is empty"}`},
		{name: "invalid json", op: calculator.Add, body: `{"a":`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"invalid JSON body"}`},
		{name: "not json at all", op: calculator.Add, body: `hello`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"invalid JSON body"}`},
		{name: "body is an array", op: calculator.Add, body: `[1,2]`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"request body must be a JSON object"}`},
		{name: "body is null", op: calculator.Add, body: `null`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: a"}`},
		{name: "missing a", op: calculator.Add, body: `{"b":3}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: a"}`},
		{name: "missing b", op: calculator.Add, body: `{"a":3}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: b"}`},
		{name: "unknown field", op: calculator.Add, body: `{"a":1,"b":2,"c":3}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"unknown field \"c\""}`},
		// encoding/json matches field names ignoring case; this documents that behavior.
		{name: "field names ignore case", op: calculator.Add, body: `{"A":1,"B":2}`, wantStatus: http.StatusOK, wantBody: `{"result":3}`},
		{name: "string instead of number", op: calculator.Add, body: `{"a":"hello","b":2}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"field a must be a valid number"}`},
		{name: "number as string", op: calculator.Add, body: `{"a":"5","b":2}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"field a must be a valid number"}`},
		{name: "number out of range", op: calculator.Add, body: `{"a":1e999,"b":1}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"field a must be a valid number"}`},
		{name: "data after the object", op: calculator.Add, body: `{"a":1,"b":2} garbage`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"request body must contain a single JSON object"}`},
		{name: "two objects", op: calculator.Add, body: `{"a":1,"b":2}{"a":3,"b":4}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"request body must contain a single JSON object"}`},
		{name: "body too large", op: calculator.Add, body: strings.Repeat(" ", maxBodyBytes+1) + `{"a":1,"b":2}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"request body too large"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(Operate(tc.op), tc.body)
			got := strings.TrimSpace(rec.Body.String())

			if rec.Code != tc.wantStatus {
				t.Errorf("Operate body %q status = %v, want %v", abbreviate(tc.body), rec.Code, tc.wantStatus)
			}

			if got != tc.wantBody {
				t.Errorf("Operate body %q response = %v, want %v", abbreviate(tc.body), got, tc.wantBody)
			}
		})
	}
}

func TestOperateUnary(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "success", body: `{"a":9}`, wantStatus: http.StatusOK, wantBody: `{"result":3}`},
		{name: "result of zero", body: `{"a":0}`, wantStatus: http.StatusOK, wantBody: `{"result":0}`},
		{name: "domain error", body: `{"a":-1}`, wantStatus: http.StatusUnprocessableEntity, wantBody: `{"error":"square root of a negative number"}`},
		{name: "missing a", body: `{}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing field: a"}`},
		{name: "second operand is rejected", body: `{"a":4,"b":1}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"unknown field \"b\""}`},
		{name: "invalid json", body: `{"a":`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"invalid JSON body"}`},
		{name: "string instead of number", body: `{"a":"nine"}`, wantStatus: http.StatusBadRequest, wantBody: `{"error":"field a must be a valid number"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(OperateUnary(calculator.Sqrt), tc.body)
			got := strings.TrimSpace(rec.Body.String())

			if rec.Code != tc.wantStatus {
				t.Errorf("OperateUnary body %q status = %v, want %v", tc.body, rec.Code, tc.wantStatus)
			}

			if got != tc.wantBody {
				t.Errorf("OperateUnary body %q response = %v, want %v", tc.body, got, tc.wantBody)
			}
		})
	}
}

// TestOperate_UsesOperands checks that a and b reach the operation in the right
// order, which a commutative operation like Add cannot show.
func TestOperate_UsesOperands(t *testing.T) {
	rec := serve(Operate(calculator.Subtract), `{"a":10,"b":3}`)

	if got := strings.TrimSpace(rec.Body.String()); got != `{"result":7}` {
		t.Errorf("Operate(Subtract) with a=10, b=3 response = %v, want %v", got, `{"result":7}`)
	}
}

// abbreviate shortens very long test bodies so failure messages stay readable.
func abbreviate(s string) string {
	if len(s) > 40 {
		return s[:40] + "..."
	}

	return s
}
