package handler

import (
	"net/http"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/calculator"
)

// NewRouter returns the HTTP handler for the whole API: one POST route per
// calculator operation, plus GET /health. Requests to unknown routes or with
// the wrong method are answered by the router itself (404 and 405, plain text).
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /add", Operate(calculator.Add))
	mux.Handle("POST /subtract", Operate(calculator.Subtract))
	mux.Handle("POST /multiply", Operate(calculator.Multiply))
	mux.Handle("POST /divide", Operate(calculator.Divide))
	mux.Handle("POST /power", Operate(calculator.Power))
	mux.Handle("POST /sqrt", OperateUnary(calculator.Sqrt))
	mux.Handle("POST /percentage", Operate(calculator.Percentage))

	mux.HandleFunc("GET /health", health)

	return mux
}

// health reports that the server is up. It takes no input and has no state, so
// it is a plain handler function instead of one built by a constructor.
func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
}
