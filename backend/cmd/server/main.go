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
)

// shutdownTimeout is how long in-flight requests get to finish once the server
// is told to stop.
const shutdownTimeout = 10 * time.Second

// getenv returns the value of the environment variable key, or fallback if the
// variable is not set or is empty.
func getenv(key, fallback string) string {
	v := os.Getenv(key)

	if v == "" {
		return fallback
	}

	return v
}

// initializeServer starts the server and sends the error that stops it to
// serverChannel. ListenAndServe blocks until the server stops, so this function
// is meant to run in its own goroutine.
func initializeServer(server *http.Server, serverChannel chan error) {
	err := server.ListenAndServe()

	serverChannel <- err
}

// main runs the server and turns a failure into a non-zero exit code. It is
// kept this small on purpose: os.Exit skips deferred calls, so all the work
// (and its defers) lives in run, which returns before main exits.
func main() {
	if err := run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// run configures and starts the HTTP server, then blocks until the server fails
// or a stop signal (Ctrl+C, or SIGTERM from `docker stop`) arrives. On a signal
// it shuts the server down gracefully and returns nil. All routing lives in
// handler.NewRouter; this function only deals with configuration and lifecycle.
func run() error {
	port := getenv("PORT", "8080")

	// ctx is cancelled when the process receives one of the stop signals.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Buffer of 1 so the server goroutine can always deliver its error and exit,
	// even if nobody is reading anymore.
	serverChannel := make(chan error, 1)

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
	go initializeServer(server, serverChannel)

	// Wait for whichever happens first: the server fails, or we are told to stop.
	select {
	case err := <-serverChannel:
		// The server stopped on its own (for example, the port is already in use).
		return err
	case <-ctx.Done():
		// Stop listening for signals so a second Ctrl+C kills the process at once.
		stop()

		// Stop accepting new connections and give in-flight requests time to finish.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down server: %w", err)
		}

		slog.Info("server stopped")
		return nil
	}
}
