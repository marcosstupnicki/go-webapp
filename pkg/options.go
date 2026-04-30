package gowebapp

import (
	golog "github.com/marcosstupnicki/go-log"
	"go.uber.org/zap/zapcore"
)

const defaultHTTPLogBodyLimitBytes int64 = 4 << 10

// webAppConfig holds configurable settings for the WebApp.
type webAppConfig struct {
	corsEnabled        bool
	corsOrigins        []string
	corsAllowedHeaders []string
	securityHeaders    map[string]string
	httpLogging        HTTPLoggingConfig
	realIPEnabled      bool
	loggerOptions      []golog.LogOption
}

// HTTPLoggingConfig controls the built-in access log middleware.
type HTTPLoggingConfig struct {
	Enabled                bool
	IncludeRequestHeaders  bool
	IncludeRequestQuery    bool
	IncludeRequestBody     bool
	IncludeResponseHeaders bool
	IncludeResponseBody    bool
	MaxBodyBytes           int64
	RedactedHeaders        []string
	RedactedQueryParams    []string
}

// defaultConfig returns sensible defaults.
func defaultConfig() webAppConfig {
	return webAppConfig{
		corsAllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		httpLogging:        DefaultHTTPLoggingConfig(),
		realIPEnabled:      true,
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

// WithHTTPLogging configures the built-in access log middleware.
func WithHTTPLogging(config HTTPLoggingConfig) Option {
	return func(c *webAppConfig) {
		c.httpLogging = normalizeHTTPLoggingConfig(config)
	}
}

// WithoutHTTPLogging disables the built-in access log middleware.
func WithoutHTTPLogging() Option {
	return func(c *webAppConfig) {
		c.httpLogging = HTTPLoggingConfig{}
	}
}

// WithRealIP controls whether chi's RealIP middleware is installed.
//
// RealIP trusts forwarding headers such as X-Forwarded-For. Disable it when
// the app can receive direct untrusted traffic and let a trusted edge proxy
// normalize the remote address instead.
func WithRealIP(enabled bool) Option {
	return func(c *webAppConfig) {
		c.realIPEnabled = enabled
	}
}

// WithLogLevel sets the minimum log level for the default go-log logger.
func WithLogLevel(level zapcore.Level) Option {
	return func(c *webAppConfig) {
		c.loggerOptions = append(c.loggerOptions, golog.WithLevel(level))
	}
}

// DefaultHTTPLoggingConfig returns a conservative access-log configuration.
// Request and response bodies, headers, and query params are opt-in because
// they often contain secrets or personal data.
func DefaultHTTPLoggingConfig() HTTPLoggingConfig {
	return HTTPLoggingConfig{
		Enabled:      true,
		MaxBodyBytes: defaultHTTPLogBodyLimitBytes,
		RedactedHeaders: []string{
			"Authorization",
			"Cookie",
			"Proxy-Authorization",
			"Set-Cookie",
			"X-Api-Key",
			"X-Auth-Token",
			"X-Service-Key",
		},
		RedactedQueryParams: []string{
			"access_token",
			"api_key",
			"client_secret",
			"code",
			"password",
			"secret",
			"token",
		},
	}
}

// VerboseHTTPLoggingConfig includes bodies, headers, and query params with
// redaction and bounded body capture. It is useful for local debugging.
func VerboseHTTPLoggingConfig() HTTPLoggingConfig {
	cfg := DefaultHTTPLoggingConfig()
	cfg.IncludeRequestHeaders = true
	cfg.IncludeRequestQuery = true
	cfg.IncludeRequestBody = true
	cfg.IncludeResponseHeaders = true
	cfg.IncludeResponseBody = true
	cfg.MaxBodyBytes = 64 << 10
	return cfg
}

func normalizeHTTPLoggingConfig(config HTTPLoggingConfig) HTTPLoggingConfig {
	if !config.Enabled {
		return HTTPLoggingConfig{}
	}
	if config.MaxBodyBytes <= 0 {
		config.MaxBodyBytes = defaultHTTPLogBodyLimitBytes
	}
	if config.RedactedHeaders == nil {
		config.RedactedHeaders = DefaultHTTPLoggingConfig().RedactedHeaders
	} else {
		config.RedactedHeaders = cloneStringSlice(config.RedactedHeaders)
	}
	if config.RedactedQueryParams == nil {
		config.RedactedQueryParams = DefaultHTTPLoggingConfig().RedactedQueryParams
	} else {
		config.RedactedQueryParams = cloneStringSlice(config.RedactedQueryParams)
	}
	return config
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
