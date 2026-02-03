package pac

import (
	"testing"
)

func TestParseProxyString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ProxyType
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "DIRECT",
			input:    "DIRECT",
			wantType: Direct,
			wantHost: "",
			wantPort: 0,
			wantErr:  false,
		},
		{
			name:     "PROXY with host:port",
			input:    "PROXY proxy.example.com:8080",
			wantType: HTTP,
			wantHost: "proxy.example.com",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "SOCKS5",
			input:    "SOCKS5 socks.example.com:1080",
			wantType: SOCKS5,
			wantHost: "socks.example.com",
			wantPort: 1080,
			wantErr:  false,
		},
		{
			name:     "SOCKS alias",
			input:    "SOCKS socks.example.com:1080",
			wantType: SOCKS5,
			wantHost: "socks.example.com",
			wantPort: 1080,
			wantErr:  false,
		},
		{
			name:     "lowercase",
			input:    "proxy proxy.example.com:8080",
			wantType: HTTP,
			wantHost: "proxy.example.com",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "invalid directive",
			input:    "INVALID proxy.example.com:8080",
			wantType: 0,
			wantHost: "",
			wantPort: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProxyString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("ParseProxyString() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Host != tt.wantHost {
				t.Errorf("ParseProxyString() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Port != tt.wantPort {
				t.Errorf("ParseProxyString() Port = %v, want %v", got.Port, tt.wantPort)
			}
		})
	}
}

func TestParseProxyList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single DIRECT",
			input:   "DIRECT",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "single PROXY",
			input:   "PROXY proxy.example.com:8080",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "fallback chain",
			input:   "PROXY proxy1:8080; PROXY proxy2:8080; DIRECT",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "with spaces",
			input:   " PROXY proxy1:8080 ; DIRECT ",
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProxyList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxyList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("ParseProxyList() returned %d proxies, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestProxyString(t *testing.T) {
	tests := []struct {
		name  string
		proxy *Proxy
		want  string
	}{
		{
			name:  "DIRECT",
			proxy: &Proxy{Type: Direct},
			want:  "DIRECT",
		},
		{
			name:  "HTTP",
			proxy: &Proxy{Type: HTTP, Host: "proxy.example.com", Port: 8080},
			want:  "PROXY proxy.example.com:8080",
		},
		{
			name:  "SOCKS5",
			proxy: &Proxy{Type: SOCKS5, Host: "socks.example.com", Port: 1080},
			want:  "SOCKS5 socks.example.com:1080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.proxy.String()
			if got != tt.want {
				t.Errorf("Proxy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
