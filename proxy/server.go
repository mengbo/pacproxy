package proxy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is the HTTP proxy server.
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	handler    *Handler
}

// NewServer creates a new proxy server.
func NewServer(listenAddr string, handler *Handler) (*Server, error) {
	return &Server{
		httpServer: &http.Server{
			Addr:        listenAddr,
			Handler:     handler,
			IdleTimeout: 120 * time.Second,
		},
		handler: handler,
	}, nil
}

// Start starts the proxy server.
func (s *Server) Start() error {
	// Create listener
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}
	s.listener = listener

	slog.Info("Starting pacproxy", "address", s.httpServer.Addr)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the proxy server.
func (s *Server) Stop() error {
	slog.Info("Stopping pacproxy")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// WaitForShutdown blocks until a shutdown signal is received.
func (s *Server) WaitForShutdown() {
	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh

	slog.Info("Shutdown signal received")
}

// Addr returns the server address.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.httpServer.Addr
}

// PrintUsageHelp prints instructions for configuring client tools.
func (s *Server) PrintUsageHelp() {
	addr := s.Addr()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "pacproxy is running on %s\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Configure your client tools with:\n")
	fmt.Fprintf(os.Stderr, "  export HTTP_PROXY=http://%s\n", addr)
	fmt.Fprintf(os.Stderr, "  export HTTPS_PROXY=http://%s\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Or use the proxy for a single command:\n")
	fmt.Fprintf(os.Stderr, "  HTTP_PROXY=http://%s curl https://example.com\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")
	fmt.Fprintf(os.Stderr, "\n")
}
