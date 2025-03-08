package gowebapp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs HTTP requests and responses using the provided logger.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture request body
			var requestBodyBytes []byte
			if r.Body != nil {
				requestBodyBytes, _ = io.ReadAll(r.Body)
				// Restore body so it can be read again downstream
				r.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
			}

			// Try to decode request body as JSON; if invalid, keep raw string
			var requestBodyMap map[string]interface{}
			if len(requestBodyBytes) > 0 {
				if err := json.Unmarshal(requestBodyBytes, &requestBodyMap); err != nil {
					requestBodyMap = map[string]interface{}{"raw": string(requestBodyBytes)}
				}
			} else {
				requestBodyMap = nil
			}

			// Wrap response writer to capture status and body
			lrw := newLoggingResponseWriter(w)

			// Measure processing time
			startTime := time.Now()

			// Call next handler in chain
			next.ServeHTTP(lrw, r)

			// Compute request duration
			duration := time.Since(startTime)

			// Capture response body
			responseBody := lrw.GetBody()

			// Try to decode response body as JSON; if invalid, keep raw string
			var responseBodyMap map[string]interface{}
			if len(responseBody) > 0 {
				if err := json.Unmarshal(responseBody, &responseBodyMap); err != nil {
					responseBodyMap = map[string]interface{}{"raw": string(responseBody)}
				}
			} else {
				responseBodyMap = nil
			}

			// Log combined request and response in a single entry
			logger.Info(context.Background(), "http_request",
				golog.WithField("request", map[string]interface{}{
					"method":  r.Method,
					"path":    r.URL.Path,
					"query":   r.URL.Query(),
					"headers": r.Header,
					"body":    requestBodyMap, // decoded map when JSON, raw otherwise
				}),
				golog.WithField("response", map[string]interface{}{
					"status":   lrw.statusCode,
					"headers":  lrw.Header(),
					"body":     responseBodyMap,         // decoded map when JSON, raw otherwise
					"duration": duration.Milliseconds(), // duration in milliseconds
				}),
			)
		})
	}
}

// LoggingResponseWriter wraps http.ResponseWriter to capture status code and response body.
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

// newLoggingResponseWriter creates a new LoggingResponseWriter.
func newLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code before writing the header.
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body before writing it through.
func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	// capture response body
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

// GetBody returns the captured response body as bytes.
func (lrw *LoggingResponseWriter) GetBody() []byte {
	return lrw.body.Bytes()
}
