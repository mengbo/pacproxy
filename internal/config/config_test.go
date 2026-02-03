package config

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				ListenAddr: "127.0.0.1:8080",
				PACURL:     "http://example.com/proxy.pac",
				LogLevel:   "info",
			},
			wantErr: false,
		},
		{
			name: "missing listen address",
			cfg: &Config{
				ListenAddr: "",
				PACURL:     "http://example.com/proxy.pac",
				LogLevel:   "info",
			},
			wantErr: true,
		},
		{
			name: "missing PAC URL",
			cfg: &Config{
				ListenAddr: "127.0.0.1:8080",
				PACURL:     "",
				LogLevel:   "info",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				ListenAddr: "127.0.0.1:8080",
				PACURL:     "http://example.com/proxy.pac",
				LogLevel:   "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid debug log level",
			cfg: &Config{
				ListenAddr: "127.0.0.1:8080",
				PACURL:     "http://example.com/proxy.pac",
				LogLevel:   "debug",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
