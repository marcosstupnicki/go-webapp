package gowebapp

// webAppConfig holds configurable settings for the WebApp.
type webAppConfig struct {
	corsEnabled        bool
	corsOrigins        []string
	corsAllowedHeaders []string
	securityHeaders    map[string]string
}

// defaultConfig returns sensible defaults.
func defaultConfig() webAppConfig {
	return webAppConfig{
		corsAllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
	}
}

// Option is a functional option for configuring the WebApp.
type Option func(*webAppConfig)

// WithCORS enables CORS with the specified allowed origins.
// Pass []string{"*"} to allow all origins.
func WithCORS(origins []string) Option {
	return func(c *webAppConfig) {
		c.corsEnabled = true
		c.corsOrigins = cloneStringSlice(origins)
	}
}

// WithCORSAllowedHeaders configures the CORS request headers accepted by the
// app when CORS is enabled.
func WithCORSAllowedHeaders(headers []string) Option {
	return func(c *webAppConfig) {
		c.corsAllowedHeaders = cloneStringSlice(headers)
	}
}

// WithSecurityHeaders enables standard security response headers.
func WithSecurityHeaders(headers map[string]string) Option {
	return func(c *webAppConfig) {
		c.securityHeaders = cloneHeaderMap(headers)
	}
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func cloneHeaderMap(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		out[key] = value
	}
	return out
}
