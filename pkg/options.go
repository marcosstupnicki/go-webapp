package gowebapp

// webAppConfig holds configurable settings for the WebApp.
type webAppConfig struct {
	corsEnabled     bool
	corsOrigins     []string
	securityHeaders bool
}

// defaultConfig returns sensible defaults.
func defaultConfig() webAppConfig {
	return webAppConfig{}
}

// Option is a functional option for configuring the WebApp.
type Option func(*webAppConfig)

// WithCORS enables CORS with the specified allowed origins.
// Pass []string{"*"} to allow all origins.
func WithCORS(origins []string) Option {
	return func(c *webAppConfig) {
		c.corsEnabled = true
		c.corsOrigins = origins
	}
}

// WithSecurityHeaders enables standard security response headers.
func WithSecurityHeaders() Option {
	return func(c *webAppConfig) {
		c.securityHeaders = true
	}
}
