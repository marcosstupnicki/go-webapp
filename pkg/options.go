package gowebapp

import (
	"net/http"
	"time"

	golog "github.com/marcosstupnicki/go-log"
)

// webAppConfig holds all configurable settings for the WebApp.
type webAppConfig struct {
	// Logger
	logger   *golog.Logger
	loggerFn func() golog.Logger // fallback: create default logger

	// CORS
	corsEnabled bool
	corsOrigins []string

	// Timeouts
	readTimeout  time.Duration
	writeTimeout time.Duration

	// Health check
	healthPath string

	// Custom handlers
	notFoundHandler         http.HandlerFunc
	methodNotAllowedHandler http.HandlerFunc

	// Logging middleware
	maxBodyLogSize   int
	skipRequestBody  bool
	skipResponseBody bool
}

// defaultConfig returns the default configuration.
func defaultConfig() webAppConfig {
	return webAppConfig{
		healthPath:     "/healthz",
		readTimeout:    30 * time.Second,
		writeTimeout:   30 * time.Second,
		maxBodyLogSize: 4096, // 4KB
	}
}

// Option is a functional option for configuring the WebApp.
type Option func(*webAppConfig)

// WithLogger sets the logger for the WebApp. The webapp middleware
// will automatically call golog.Enrich to add request_id, method,
// and path to the context for every request.
//
//	logger := golog.New("local")
//	app := gowebapp.New("local", "8080", gowebapp.WithLogger(logger))
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

// WithTimeout sets both read and write timeouts for the HTTP server.
func WithTimeout(d time.Duration) Option {
	return func(c *webAppConfig) {
		c.readTimeout = d
		c.writeTimeout = d
	}
}

// WithHealthPath sets the health check endpoint path.
// Default is "/healthz".
func WithHealthPath(path string) Option {
	return func(c *webAppConfig) {
		c.healthPath = path
	}
}

// WithNotFoundHandler sets a custom handler for 404 responses.
func WithNotFoundHandler(h http.HandlerFunc) Option {
	return func(c *webAppConfig) {
		c.notFoundHandler = h
	}
}

// WithMethodNotAllowedHandler sets a custom handler for 405 responses.
func WithMethodNotAllowedHandler(h http.HandlerFunc) Option {
	return func(c *webAppConfig) {
		c.methodNotAllowedHandler = h
	}
}

// WithMaxBodyLogSize sets the maximum body size (in bytes) that the
// logging middleware will capture. Bodies larger than this are truncated
// in logs. Default is 4096 (4KB).
func WithMaxBodyLogSize(size int) Option {
	return func(c *webAppConfig) {
		c.maxBodyLogSize = size
	}
}

// WithSkipBodyLogging disables request and/or response body logging.
func WithSkipBodyLogging(skipRequest, skipResponse bool) Option {
	return func(c *webAppConfig) {
		c.skipRequestBody = skipRequest
		c.skipResponseBody = skipResponse
	}
}
