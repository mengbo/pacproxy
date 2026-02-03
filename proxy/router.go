// Package proxy provides the HTTP proxy server implementation.
package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mengbo/pacproxy/pac"
	"github.com/mengbo/pacproxy/upstream"
)

// Router handles routing requests based on PAC rules.
type Router struct {
	engine     *pac.Engine
	fetcher    *pac.Fetcher
	pacURL     string
	mu         sync.RWMutex
	connectors map[string]func(host string, port int) (upstream.Connector, error)
}

// NewRouter creates a new proxy router.
func NewRouter(pacURL string, fetcher *pac.Fetcher) (*Router, error) {
	r := &Router{
		pacURL:     pacURL,
		fetcher:    fetcher,
		connectors: make(map[string]func(host string, port int) (upstream.Connector, error)),
	}

	// Register connector factories
	r.connectors["DIRECT"] = func(host string, port int) (upstream.Connector, error) {
		return upstream.NewDirectConnector(), nil
	}
	r.connectors["PROXY"] = func(host string, port int) (upstream.Connector, error) {
		return upstream.NewHTTPConnector(host, port, "", ""), nil
	}
	r.connectors["SOCKS5"] = func(host string, port int) (upstream.Connector, error) {
		return upstream.NewSOCKS5Connector(host, port, "", "")
	}
	r.connectors["SOCKS"] = r.connectors["SOCKS5"] // Alias

	// Initial PAC fetch
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.refreshPAC(ctx); err != nil {
		return nil, fmt.Errorf("initial PAC fetch: %w", err)
	}

	return r, nil
}

// Route determines the upstream connector for the given URL.
func (r *Router) Route(ctx context.Context, rawURL string, host string) (upstream.Connector, error) {
	// Refresh PAC if needed (in background)
	go r.refreshPACIfNeeded()

	r.mu.RLock()
	engine := r.engine
	r.mu.RUnlock()

	if engine == nil {
		return nil, fmt.Errorf("PAC engine not initialized")
	}

	// Call FindProxyForURL
	proxies, err := engine.FindProxyForURLContext(ctx, rawURL, host)
	if err != nil {
		return nil, fmt.Errorf("PAC resolution: %w", err)
	}

	if len(proxies) > 0 {
		slog.Debug("PAC result", "proxy", proxies[0].String(), "total", len(proxies))
	}

	// Try each proxy in order (fallback support)
	for _, proxy := range proxies {
		connector, err := r.createConnector(proxy)
		if err != nil {
			slog.Warn("Failed to create connector",
				"proxy", proxy.String(),
				"error", err)
			continue
		}
		return connector, nil
	}

	return nil, fmt.Errorf("no valid proxy found")
}

// createConnector creates a connector for the given proxy.
func (r *Router) createConnector(proxy *pac.Proxy) (upstream.Connector, error) {
	factory, ok := r.connectors[proxy.Type.String()]
	if !ok {
		return nil, fmt.Errorf("unknown proxy type: %s", proxy.Type.String())
	}

	return factory(proxy.Host, proxy.Port)
}

// refreshPACIfNeeded refreshes the PAC file if the cache has expired.
func (r *Router) refreshPACIfNeeded() {
	// The fetcher handles caching, so just try to fetch
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.refreshPAC(ctx); err != nil {
		slog.Warn("Failed to refresh PAC", "error", err)
	}
}

// refreshPAC fetches and parses the PAC file.
func (r *Router) refreshPAC(ctx context.Context) error {
	content, err := r.fetcher.Fetch(ctx, r.pacURL)
	if err != nil {
		return err
	}

	engine, err := pac.NewEngine(content)
	if err != nil {
		return fmt.Errorf("parsing PAC: %w", err)
	}

	r.mu.Lock()
	r.engine = engine
	r.mu.Unlock()

	slog.Info("PAC file refreshed successfully")
	return nil
}
