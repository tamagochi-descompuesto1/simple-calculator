// Package handler exposes the calculator over HTTP: it validates requests and
// writes JSON responses. The math rules live in the calculator package.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxBodyBytes caps the request body at 1 MB.
const maxBodyBytes = 1 << 20

// BinaryOp is an operation on two operands.
type BinaryOp func(a, b float64) (float64, error)

// UnaryOp is an operation on one operand.
type UnaryOp func(a float64) (float64, error)

// binaryRequest uses pointers to tell a missing field (nil) from 0.
type binaryRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

type unaryRequest struct {
	A *float64 `json:"a"`
}

// Operate returns a handler that applies op to the a and b of the request body.
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

// OperateUnary is like [Operate], for one operand.
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

// decodeBody reads the JSON body into dst. Its errors are safe to show the client.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return describeDecodeError(err)
	}

	// Decode reads one value, so a second read must hit EOF or there is trailing data.
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

// describeDecodeError replaces the decoder's messages, which expose Go
// internals, with short ones.
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
		// The decoder has no error type for this case, only text.
		return errors.New(strings.TrimPrefix(err.Error(), "json: "))
	default:
		return errors.New("invalid JSON body")
	}
}
