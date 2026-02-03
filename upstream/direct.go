package upstream

import (
	"context"
	"net"
)

// DirectConnector establishes direct connections without any proxy.
type DirectConnector struct{}

// NewDirectConnector creates a new direct connector.
func NewDirectConnector() *DirectConnector {
	return &DirectConnector{}
}

// Connect establishes a direct TCP connection to the target address.
func (d *DirectConnector) Connect(ctx context.Context, address string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, "tcp", address)
}

// Name returns the name of the connector.
func (d *DirectConnector) Name() string {
	return "DIRECT"
}
