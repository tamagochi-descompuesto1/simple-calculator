// Command server runs the calculator HTTP API.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/handler"
	"github.com/tamagochi-descompuesto1/simple-calculator/backend/internal/middleware"
)

// shutdownTimeout is how long in-flight requests get to finish on shutdown.
const shutdownTimeout = 10 * time.Second

// getenv returns the variable key, or fallback if it is unset or empty.
func getenv(key, fallback string) string {
	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	return v
}

// initializeServer runs the server and sends the error that stops it to
// serverChannel. ListenAndServe blocks, so call this in a goroutine.
func initializeServer(server *http.Server, serverChannel chan error) {
	err := server.ListenAndServe()

	serverChannel <- err
}

// newHandler wraps the router in the middleware. It is a function so tests run
// the same chain as the server.
//
// A request goes through Logging, Recover, CORS and then the router. Logging is
// outermost so it records the 500 that Recover produces, and CORS answers
// preflight requests before the router rejects them with 405.
func newHandler(allowedOrigin string) http.Handler {
	h := handler.NewRouter()
	h = middleware.CORS(allowedOrigin)(h)
	h = middleware.Recover(h)
	h = middleware.Logging(h)

	return h
}

// main exits with a non-zero code on failure. The work is in run because
// os.Exit skips deferred calls.
func main() {
	if err := run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// run starts the server and blocks until it fails or a stop signal arrives,
// then shuts down gracefully.
func run() error {
	port := getenv("PORT", "8080")

	// The frontend's origin. The default is the Vite development server.
	allowedOrigin := getenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	// Cancelled on Ctrl+C or SIGTERM (docker stop).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Buffered so the server goroutine can always deliver its error and exit.
	serverChannel := make(chan error, 1)

	server := &http.Server{
		// No host: listen on all interfaces, which a container needs.
		Addr:    ":" + port,
		Handler: newHandler(allowedOrigin),

		// Timeouts so slow clients cannot hold connections open.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	slog.Info("server starting", "port", port, "cors_origin", allowedOrigin)
	go initializeServer(server, serverChannel)

	select {
	case err := <-serverChannel:
		// The server failed on its own, e.g. the port is already in use.
		return err
	case <-ctx.Done():
		// Restore default signal handling so a second Ctrl+C exits at once.
		stop()

		// Stop accepting connections and let in-flight requests finish.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down server: %w", err)
		}

		slog.Info("server stopped")
		return nil
	}
}
