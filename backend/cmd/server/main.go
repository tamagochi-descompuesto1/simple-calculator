package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/handler"
)

// getenv returns the value of the environment variable key, or fallback if the
// variable is not set or is empty.
func getenv(key, fallback string) string {
	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	return v
}

// main configures and starts the HTTP server. All routing lives in
// handler.NewRouter; this function only deals with configuration and startup.
func main() {
	port := getenv("PORT", "8080")

	server := &http.Server{
		// ":8080" (no host) listens on every network interface, which is
		// required for the server to be reachable from outside a container.
		Addr:    ":" + port,
		Handler: handler.NewRouter(),

		// Timeouts stop slow or stalled clients from holding connections open
		// forever. Requests and responses here are tiny, so these are generous.
		ReadHeaderTimeout: 5 * time.Second,   // time to receive the request headers
		ReadTimeout:       10 * time.Second,  // time to receive the whole request
		WriteTimeout:      10 * time.Second,  // time to send the response
		IdleTimeout:       120 * time.Second, // how long an idle keep-alive connection stays open
		MaxHeaderBytes:    1 << 20,           // 1 MB cap on the request headers
	}

	slog.Info("server starting", "port", port)
	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "error", err)
	}
}
