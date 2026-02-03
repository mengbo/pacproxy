package upstream

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
)

// SOCKS5Connector connects through a SOCKS5 proxy.
type SOCKS5Connector struct {
	dialer proxy.Dialer
}

// NewSOCKS5Connector creates a new SOCKS5 proxy connector.
func NewSOCKS5Connector(proxyHost string, proxyPort int, username, password string) (*SOCKS5Connector, error) {
	proxyURL := &url.URL{
		Scheme: "socks5",
		Host:   fmt.Sprintf("%s:%d", proxyHost, proxyPort),
	}

	var auth *proxy.Auth
	if username != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("creating SOCKS5 dialer: %w", err)
	}

	// If authentication is provided, we need to use a different approach
	if auth != nil {
		dialer, err = proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("creating SOCKS5 dialer with auth: %w", err)
		}
	}

	return &SOCKS5Connector{
		dialer: dialer,
	}, nil
}

// Connect establishes a connection through the SOCKS5 proxy.
func (s *SOCKS5Connector) Connect(ctx context.Context, address string) (net.Conn, error) {
	// The proxy.Dialer doesn't support context, so we use a workaround
	// Create a channel to handle the dial
	type result struct {
		conn net.Conn
		err  error
	}

	resultCh := make(chan result, 1)

	go func() {
		conn, err := s.dialer.Dial("tcp", address)
		resultCh <- result{conn, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-resultCh:
		return r.conn, r.err
	}
}

// Name returns the name of the connector.
func (s *SOCKS5Connector) Name() string {
	return "SOCKS5"
}
