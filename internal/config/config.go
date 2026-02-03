// Package config provides configuration structures.
package config

import (
	"fmt"
	"net/url"
)

// Config holds the application configuration.
type Config struct {
	ListenAddr string
	PACURL     string
	LogLevel   string
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen address is required")
	}

	if c.PACURL == "" {
		return fmt.Errorf("PAC URL is required")
	}

	// Validate PAC URL
	if _, err := url.Parse(c.PACURL); err != nil {
		return fmt.Errorf("invalid PAC URL: %w", err)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}
