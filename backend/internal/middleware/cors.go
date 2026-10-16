package middleware

import "net/http"

// CORS returns a middleware that lets a browser page served from allowedOrigin
// call the API. Browsers block requests to a different origin (for example a
// frontend on localhost:5173 calling the API on localhost:8080) unless the
// server allows it with these headers.
//
// Before a POST with a JSON body, the browser first sends a "preflight":
// an OPTIONS request asking whether the real request is allowed. The router has
// no OPTIONS routes and would answer 405, so CORS answers the preflight itself
// with 204 and does not pass it on.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", allowedOrigin)
			// The answer depends on the Origin request header, so caches must
			// not reuse it for requests with a different one.
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

// isPreflight reports whether r is a CORS preflight request: an OPTIONS request
// that announces, in Access-Control-Request-Method, the method it wants to use.
// A plain OPTIONS request without that header is not a preflight.
func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
