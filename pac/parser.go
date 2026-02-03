package pac

import (
	"fmt"
	"net/url"
	"strings"
)

// ProxyType represents the type of proxy.
type ProxyType int

const (
	// Direct means no proxy, connect directly.
	Direct ProxyType = iota
	// HTTP represents an HTTP proxy.
	HTTP
	// SOCKS5 represents a SOCKS5 proxy.
	SOCKS5
	// SOCKS4 represents a SOCKS4 proxy.
	SOCKS4
)

// String returns the string representation of the proxy type.
func (p ProxyType) String() string {
	switch p {
	case Direct:
		return "DIRECT"
	case HTTP:
		return "PROXY"
	case SOCKS5:
		return "SOCKS5"
	case SOCKS4:
		return "SOCKS4"
	default:
		return "UNKNOWN"
	}
}

// Proxy represents a single proxy directive.
type Proxy struct {
	Type ProxyType
	Host string
	Port int
}

// String returns the string representation of the proxy.
func (p *Proxy) String() string {
	if p.Type == Direct {
		return "DIRECT"
	}
	return fmt.Sprintf("%s %s:%d", p.Type.String(), p.Host, p.Port)
}

// ParseProxyString parses a PAC proxy directive string.
// It handles formats like:
//   - "DIRECT"
//   - "PROXY host:port"
//   - "SOCKS5 host:port"
//   - "SOCKS host:port" (treated as SOCKS5)
//   - "SOCKS4 host:port"
func ParseProxyString(s string) (*Proxy, error) {
	s = strings.TrimSpace(s)

	if strings.ToUpper(s) == "DIRECT" {
		return &Proxy{Type: Direct}, nil
	}

	parts := strings.Fields(s)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid proxy directive: %s", s)
	}

	var proxyType ProxyType
	switch strings.ToUpper(parts[0]) {
	case "PROXY":
		proxyType = HTTP
	case "SOCKS", "SOCKS5":
		proxyType = SOCKS5
	case "SOCKS4":
		proxyType = SOCKS4
	default:
		return nil, fmt.Errorf("unknown proxy type: %s", parts[0])
	}

	// Parse host:port
	host, port, err := parseHostPort(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parsing host:port: %w", err)
	}

	return &Proxy{
		Type: proxyType,
		Host: host,
		Port: port,
	}, nil
}

// ParseProxyList parses a semicolon-separated list of proxy directives.
// It handles fallback chains like:
//   - "PROXY proxy1:8080; PROXY proxy2:8080; DIRECT"
func ParseProxyList(s string) ([]*Proxy, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []*Proxy{{Type: Direct}}, nil
	}

	directives := strings.Split(s, ";")
	proxies := make([]*Proxy, 0, len(directives))

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		proxy, err := ParseProxyString(directive)
		if err != nil {
			return nil, fmt.Errorf("parsing directive %q: %w", directive, err)
		}
		proxies = append(proxies, proxy)
	}

	if len(proxies) == 0 {
		return []*Proxy{{Type: Direct}}, nil
	}

	return proxies, nil
}

// parseHostPort parses a host:port string.
func parseHostPort(s string) (string, int, error) {
	// Handle IPv6 addresses in brackets
	if strings.HasPrefix(s, "[") {
		idx := strings.LastIndex(s, "]")
		if idx == -1 {
			return "", 0, fmt.Errorf("invalid IPv6 address: %s", s)
		}
		host := s[1:idx]
		portStr := ""
		if idx+1 < len(s) && s[idx+1] == ':' {
			portStr = s[idx+2:]
		}
		if portStr == "" {
			return host, 80, nil // Default port
		}
		var port int
		_, err := fmt.Sscanf(portStr, "%d", &port)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %s", portStr)
		}
		return host, port, nil
	}

	// Handle host:port or host
	u, err := url.Parse("http://" + s)
	if err != nil {
		return "", 0, err
	}

	host := u.Hostname()
	portStr := u.Port()

	if portStr == "" {
		return host, 80, nil // Default port
	}

	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %s", portStr)
	}

	return host, port, nil
}
