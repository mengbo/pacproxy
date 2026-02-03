package pac

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/dop251/goja"
)

// Engine executes PAC (Proxy Auto-Config) JavaScript files.
type Engine struct {
	vm       *goja.Runtime
	mu       sync.Mutex
	resolver *net.Resolver
}

// NewEngine creates a new PAC engine with the given PAC script.
func NewEngine(script string) (*Engine, error) {
	e := &Engine{
		vm:       goja.New(),
		resolver: net.DefaultResolver,
	}

	// Register PAC helper functions
	if err := e.registerHelpers(); err != nil {
		return nil, fmt.Errorf("registering PAC helpers: %w", err)
	}

	// Execute the PAC script
	if _, err := e.vm.RunString(script); err != nil {
		return nil, fmt.Errorf("executing PAC script: %w", err)
	}

	return e, nil
}

// FindProxyForURL executes the FindProxyForURL function in the PAC script.
// It returns a list of proxy directives.
func (e *Engine) FindProxyForURL(rawURL string, host string) ([]*Proxy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Call FindProxyForURL
	fn, ok := goja.AssertFunction(e.vm.Get("FindProxyForURL"))
	if !ok {
		return nil, fmt.Errorf("FindProxyForURL function not found in PAC script")
	}

	result, err := fn(goja.Undefined(), e.vm.ToValue(rawURL), e.vm.ToValue(host))
	if err != nil {
		return nil, fmt.Errorf("executing FindProxyForURL: %w", err)
	}

	proxyStr := result.String()
	return ParseProxyList(proxyStr)
}

// FindProxyForURLContext executes FindProxyForURL with context support.
func (e *Engine) FindProxyForURLContext(ctx context.Context, rawURL string, host string) ([]*Proxy, error) {
	// For now, just call the non-context version
	// In the future, this could support cancellation
	return e.FindProxyForURL(rawURL, host)
}

// registerHelpers registers all PAC helper functions in the JavaScript runtime.
func (e *Engine) registerHelpers() error {
	helpers := map[string]interface{}{
		"isPlainHostName":     e.isPlainHostName,
		"dnsDomainIs":         e.dnsDomainIs,
		"localHostOrDomainIs": e.localHostOrDomainIs,
		"isResolvable":        e.isResolvable,
		"isInNet":             e.isInNet,
		"dnsResolve":          e.dnsResolve,
		"myIpAddress":         e.myIpAddress,
		"dnsDomainLevels":     e.dnsDomainLevels,
		"shExpMatch":          e.shExpMatch,
		"weekdayRange":        e.weekdayRange,
		"dateRange":           e.dateRange,
		"timeRange":           e.timeRange,
	}

	for name, fn := range helpers {
		if err := e.vm.Set(name, fn); err != nil {
			return fmt.Errorf("setting %s: %w", name, err)
		}
	}

	return nil
}

// isPlainHostName checks if the host name contains no dots.
func (e *Engine) isPlainHostName(host string) bool {
	return !strings.Contains(host, ".")
}

// dnsDomainIs checks if the host is in the given domain.
func (e *Engine) dnsDomainIs(host, domain string) bool {
	if !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	return strings.HasSuffix(strings.ToLower(host), strings.ToLower(domain))
}

// localHostOrDomainIs checks if the host matches the given domain or is a local host.
func (e *Engine) localHostOrDomainIs(host, hostdom string) bool {
	if host == hostdom {
		return true
	}
	if !strings.HasPrefix(hostdom, ".") {
		hostdom = "." + hostdom
	}
	return strings.HasSuffix(strings.ToLower(host), strings.ToLower(hostdom))
}

// isResolvable checks if the host can be resolved to an IP address.
func (e *Engine) isResolvable(host string) bool {
	_, err := e.resolver.LookupHost(context.Background(), host)
	return err == nil
}

// isInNet checks if the IP address is in the given network.
func (e *Engine) isInNet(ipaddr, pattern, maskstr string) bool {
	ip := net.ParseIP(ipaddr)
	if ip == nil {
		return false
	}

	patternIP := net.ParseIP(pattern)
	if patternIP == nil {
		return false
	}

	maskIP := net.ParseIP(maskstr)
	if maskIP == nil {
		return false
	}

	// Convert to 4-byte representation for IPv4
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
		patternIP = patternIP.To4()
		maskIP = maskIP.To4()
	}

	if patternIP == nil || maskIP == nil {
		return false
	}

	for i := range ip {
		if (ip[i] & maskIP[i]) != (patternIP[i] & maskIP[i]) {
			return false
		}
	}
	return true
}

// dnsResolve resolves the host to an IP address.
func (e *Engine) dnsResolve(host string) string {
	addrs, err := e.resolver.LookupHost(context.Background(), host)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0]
}

// myIpAddress returns the local machine's IP address.
func (e *Engine) myIpAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// dnsDomainLevels returns the number of domain levels in the host.
func (e *Engine) dnsDomainLevels(host string) int {
	parts := strings.Split(host, ".")
	count := 0
	for _, part := range parts {
		if part != "" {
			count++
		}
	}
	return count
}

// shExpMatch checks if the string matches the given shell expression pattern.
func (e *Engine) shExpMatch(str, pattern string) bool {
	// Convert shell pattern to regex
	regex := shellPatternToRegex(pattern)
	matched, _ := regexp.MatchString(regex, str)
	return matched
}

// weekdayRange is a stub for the PAC function (not commonly used).
func (e *Engine) weekdayRange(args ...interface{}) bool {
	// Simplified implementation - always returns true
	return true
}

// dateRange is a stub for the PAC function (not commonly used).
func (e *Engine) dateRange(args ...interface{}) bool {
	// Simplified implementation - always returns true
	return true
}

// timeRange is a stub for the PAC function (not commonly used).
func (e *Engine) timeRange(args ...interface{}) bool {
	// Simplified implementation - always returns true
	return true
}

// shellPatternToRegex converts a shell glob pattern to a regex pattern.
func shellPatternToRegex(pattern string) string {
	var result strings.Builder
	result.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch c {
		case '*':
			result.WriteString(".*")
		case '?':
			result.WriteString(".")
		case '.':
			result.WriteString(`\.`) // Escape dot
		case '\\':
			if i+1 < len(pattern) {
				result.WriteString(regexp.QuoteMeta(string(pattern[i+1])))
				i++
			}
		default:
			result.WriteString(regexp.QuoteMeta(string(c)))
		}
	}

	result.WriteString("$")
	return result.String()
}

// SetResolver sets a custom DNS resolver for the engine.
func (e *Engine) SetResolver(resolver *net.Resolver) {
	e.resolver = resolver
}
