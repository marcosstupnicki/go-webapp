package gowebapp

import (
	golog "github.com/marcosstupnicki/go-log"
)

// webAppConfig holds configurable settings for the WebApp.
type webAppConfig struct {
	logger      *golog.Logger
	corsEnabled bool
	corsOrigins []string
}

// defaultConfig returns sensible defaults.
func defaultConfig() webAppConfig {
	return webAppConfig{}
}

// Option is a functional option for configuring the WebApp.
type Option func(*webAppConfig)

// WithLogger sets an explicit logger. When omitted, New() creates one
// automatically from the environment string.
func WithLogger(l golog.Logger) Option {
	return func(c *webAppConfig) {
		c.logger = &l
	}
}

// WithCORS enables CORS with the specified allowed origins.
// Pass []string{"*"} to allow all origins.
func WithCORS(origins []string) Option {
	return func(c *webAppConfig) {
		c.corsEnabled = true
		c.corsOrigins = origins
	}
}
