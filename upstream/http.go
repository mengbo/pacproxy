package upstream

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// HTTPConnector connects through an HTTP proxy.
type HTTPConnector struct {
	proxyURL *url.URL
	username string
	password string
}

// NewHTTPConnector creates a new HTTP proxy connector.
func NewHTTPConnector(proxyHost string, proxyPort int, username, password string) *HTTPConnector {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", proxyHost, proxyPort),
	}
	return &HTTPConnector{
		proxyURL: proxyURL,
		username: username,
		password: password,
	}
}

// Connect establishes a connection through the HTTP proxy using CONNECT method.
func (h *HTTPConnector) Connect(ctx context.Context, address string) (net.Conn, error) {
	// Connect to the proxy server
	var dialer net.Dialer
	proxyConn, err := dialer.DialContext(ctx, "tcp", h.proxyURL.Host)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %w", err)
	}

	// Send CONNECT request
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: address},
		Host:   address,
		Header: make(http.Header),
	}

	// Add authentication if provided
	if h.username != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(h.username + ":" + h.password))
		req.Header.Set("Proxy-Authorization", "Basic "+auth)
	}

	// Write the request
	if err := req.Write(proxyConn); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("writing CONNECT request: %w", err)
	}

	// Read the response
	br := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("reading CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		proxyConn.Close()
		return nil, fmt.Errorf("CONNECT failed with status: %s", resp.Status)
	}

	return proxyConn, nil
}

// Name returns the name of the connector.
func (h *HTTPConnector) Name() string {
	return "HTTP"
}
