package handler

import (
	"net/http"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/calculator"
)

// NewRouter returns the API routes. The router itself answers unknown routes
// and wrong methods (404 and 405), in plain text.
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

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
}
