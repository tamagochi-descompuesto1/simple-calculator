// Package handler exposes the calculator operations over HTTP. It checks that
// requests are well formed and turns results and errors into JSON responses.
// The math itself, and its rules, live in the calculator package.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxBodyBytes is the largest request body we accept (1 MB). A request only
// needs a few bytes, so anything bigger is a mistake or an attack.
const maxBodyBytes = 1 << 20

// BinaryOp is any operation that takes two operands.
type BinaryOp func(a, b float64) (float64, error)

// UnaryOp is any operation that takes one operand.
type UnaryOp func(a float64) (float64, error)

// binaryRequest is the JSON body of an operation with two operands.
// The fields are pointers so a missing field (nil) can be told apart
// from a field sent as 0.
type binaryRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

// unaryRequest is the JSON body of an operation with one operand.
type unaryRequest struct {
	A *float64 `json:"a"`
}

// Operate returns a handler that reads a and b from the request body, applies
// op to them and writes the result as JSON.
func Operate(op BinaryOp) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req binaryRequest
		if err := decodeBody(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.A == nil {
			writeError(w, http.StatusBadRequest, "missing field: a")
			return
		}
		if req.B == nil {
			writeError(w, http.StatusBadRequest, "missing field: b")
			return
		}

		result, err := op(*req.A, *req.B)
		if err != nil {
			writeOperationError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, successBody{Result: result})
	}
}

// OperateUnary is like [Operate], for operations with a single operand.
func OperateUnary(op UnaryOp) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req unaryRequest
		if err := decodeBody(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.A == nil {
			writeError(w, http.StatusBadRequest, "missing field: a")
			return
		}

		result, err := op(*req.A)
		if err != nil {
			writeOperationError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, successBody{Result: result})
	}
}

// decodeBody reads the JSON body of r into dst.
//
// The returned error is always safe to show to the client: its message says
// what is wrong with the request (empty body, invalid JSON, unknown field,
// wrong type, too large, or extra data after the JSON object).
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return describeDecodeError(err)
	}

	// Decode reads only the first JSON value, so `{"a":1,"b":2} garbage` would
	// be accepted. Reading again must hit the end of the body (io.EOF).
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

// describeDecodeError turns an error from the JSON decoder into a short message
// for the client. The decoder's own messages mention Go internals (struct
// names, types), which the client should not see.
func describeDecodeError(err error) error {
	var typeErr *json.UnmarshalTypeError
	var tooLarge *http.MaxBytesError

	switch {
	case errors.Is(err, io.EOF):
		return errors.New("request body is empty")
	case errors.As(err, &tooLarge):
		return errors.New("request body too large")
	case errors.As(err, &typeErr):
		if typeErr.Field == "" {
			return errors.New("request body must be a JSON object")
		}
		return fmt.Errorf("field %s must be a valid number", typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		// The decoder has no dedicated error type for this case, only the text.
		return errors.New(strings.TrimPrefix(err.Error(), "json: "))
	default:
		return errors.New("invalid JSON body")
	}
}
