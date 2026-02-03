// Package upstream provides upstream proxy connectors.
package upstream

import (
	"context"
	"net"
)

// Connector is the interface for establishing connections through upstream proxies.
type Connector interface {
	// Connect establishes a connection to the target address through the upstream proxy.
	Connect(ctx context.Context, address string) (net.Conn, error)
	// Name returns the name of the connector.
	Name() string
}
