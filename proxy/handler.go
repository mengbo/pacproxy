package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

// Handler handles HTTP proxy requests.
type Handler struct {
	router *Router
}

// NewHandler creates a new HTTP handler.
func NewHandler(router *Router) *Handler {
	return &Handler{
		router: router,
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle CONNECT method for HTTPS
	if r.Method == http.MethodConnect {
		h.handleConnect(w, r)
		return
	}

	// Handle regular HTTP requests
	h.handleHTTP(w, r)
}

// handleHTTP handles regular HTTP requests.
func (h *Handler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "http://" + r.Host + targetURL
	}

	hostWithoutPort, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		hostWithoutPort = r.Host
	}

	connector, err := h.router.Route(ctx, targetURL, hostWithoutPort)
	if err != nil {
		slog.Error("Routing failed", "error", err, "url", targetURL)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
		return
	}

	slog.Debug("Routing request",
		"method", r.Method,
		"url", targetURL,
		"upstream", connector.Name())

	// Connect to the target
	conn, err := connector.Connect(ctx, r.Host)
	if err != nil {
		slog.Error("Connection failed", "error", err, "host", r.Host)
		http.Error(w, "Connection Failed", http.StatusBadGateway)
		return
	}
	defer conn.Close()

	// Write the request to the target
	if err := r.Write(conn); err != nil {
		slog.Error("Failed to write request", "error", err)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
		return
	}

	// Read the response
	resp, err := http.ReadResponse(
		bufio.NewReader(conn),
		r,
	)
	if err != nil {
		slog.Error("Failed to read response", "error", err)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	if _, err := io.Copy(w, resp.Body); err != nil {
		slog.Error("Failed to copy response body", "error", err)
	}
}

// handleConnect handles HTTPS CONNECT requests.
func (h *Handler) handleConnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	host := r.Host

	hostWithoutPort, _, err := net.SplitHostPort(host)
	if err != nil {
		hostWithoutPort = host
	}

	// Route the request
	connector, err := h.router.Route(ctx, "https://"+host, hostWithoutPort)
	if err != nil {
		slog.Error("Routing failed", "error", err, "host", host)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
		return
	}

	slog.Debug("Routing CONNECT",
		"host", host,
		"upstream", connector.Name())

	// Connect to the target
	targetConn, err := connector.Connect(ctx, host)
	if err != nil {
		slog.Error("Connection failed", "error", err, "host", host)
		http.Error(w, "Connection Failed", http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		slog.Error("Failed to hijack connection", "error", err)
		http.Error(w, "Proxy Error", http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection established
	fmt.Fprintf(clientConn, "HTTP/%d.%d 200 Connection established\r\n\r\n",
		r.ProtoMajor, r.ProtoMinor)

	// Bidirectional copy
	go func() {
		io.Copy(targetConn, clientConn)
		targetConn.Close()
	}()

	io.Copy(clientConn, targetConn)
}
