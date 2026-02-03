package pac

import (
	"testing"
)

func TestNewEngine(t *testing.T) {
	script := `
function FindProxyForURL(url, host) {
    if (host == "localhost" || host == "127.0.0.1") {
        return "DIRECT";
    }
    return "PROXY proxy.example.com:8080";
}
`

	engine, err := NewEngine(script)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
}

func TestFindProxyForURL(t *testing.T) {
	script := `
function FindProxyForURL(url, host) {
    if (host == "localhost" || host == "127.0.0.1") {
        return "DIRECT";
    }
    if (shExpMatch(host, "*.internal.com")) {
        return "PROXY internal-proxy:8080";
    }
    return "PROXY proxy.example.com:8080; DIRECT";
}
`

	engine, err := NewEngine(script)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name     string
		url      string
		host     string
		wantType ProxyType
	}{
		{
			name:     "localhost",
			url:      "http://localhost/path",
			host:     "localhost",
			wantType: Direct,
		},
		{
			name:     "external",
			url:      "http://example.com/path",
			host:     "example.com",
			wantType: HTTP,
		},
		{
			name:     "internal",
			url:      "http://app.internal.com/path",
			host:     "app.internal.com",
			wantType: HTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxies, err := engine.FindProxyForURL(tt.url, tt.host)
			if err != nil {
				t.Errorf("FindProxyForURL() error = %v", err)
				return
			}
			if len(proxies) == 0 {
				t.Error("FindProxyForURL() returned no proxies")
				return
			}
			if proxies[0].Type != tt.wantType {
				t.Errorf("FindProxyForURL() first proxy type = %v, want %v", proxies[0].Type, tt.wantType)
			}
		})
	}
}

func TestPACHelpers(t *testing.T) {
	script := `
function FindProxyForURL(url, host) {
    return "DIRECT";
}
`

	engine, err := NewEngine(script)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name:     "isPlainHostName true",
			fn:       func() interface{} { return engine.isPlainHostName("localhost") },
			expected: true,
		},
		{
			name:     "isPlainHostName false",
			fn:       func() interface{} { return engine.isPlainHostName("example.com") },
			expected: false,
		},
		{
			name:     "dnsDomainIs true",
			fn:       func() interface{} { return engine.dnsDomainIs("www.example.com", "example.com") },
			expected: true,
		},
		{
			name:     "dnsDomainIs false",
			fn:       func() interface{} { return engine.dnsDomainIs("www.example.com", "other.com") },
			expected: false,
		},
		{
			name:     "dnsDomainLevels",
			fn:       func() interface{} { return engine.dnsDomainLevels("www.example.com") },
			expected: 3,
		},
		{
			name:     "shExpMatch true",
			fn:       func() interface{} { return engine.shExpMatch("www.example.com", "*.example.com") },
			expected: true,
		},
		{
			name:     "shExpMatch false",
			fn:       func() interface{} { return engine.shExpMatch("www.other.com", "*.example.com") },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
