package middleware

import "net/http"

// CORS lets a browser page served from allowedOrigin call the API. It answers
// preflight (OPTIONS) requests itself with 204, since the router has no OPTIONS
// routes and would reply 405.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", allowedOrigin)
			// Tell caches the response depends on the request's Origin.
			header.Add("Vary", "Origin")

			if isPreflight(r) {
				header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				header.Set("Access-Control-Allow-Headers", "Content-Type")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isPreflight reports whether r is a CORS preflight: an OPTIONS request that
// names the method it wants in Access-Control-Request-Method.
func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
