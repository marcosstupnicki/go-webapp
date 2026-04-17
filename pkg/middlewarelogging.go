package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with full body, headers, and duration.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read and restore request body.
			var requestBodyMap map[string]interface{}
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(bodyBytes) > 0 {
					if err := json.Unmarshal(bodyBytes, &requestBodyMap); err != nil {
						requestBodyMap = map[string]interface{}{"raw": string(bodyBytes)}
					}
				}
			}

			lrw := newLoggingResponseWriter(w)

			start := time.Now()
			next.ServeHTTP(lrw, r)
			duration := time.Since(start)

			logger.Info(r.Context(), "http_request",
				golog.Field("request", map[string]interface{}{
					"body":    requestBodyMap,
					"headers": r.Header,
					"method":  r.Method,
					"path":    r.URL.Path,
					"query":   r.URL.Query(),
				}),
				golog.Field("response", map[string]interface{}{
					"body":     lrw.GetBodyMap(),
					"duration": duration.Milliseconds(),
					"headers":  lrw.Header(),
					"status":   lrw.statusCode,
				}),
			)
		})
	}
}

// LoggingResponseWriter captures the status code and body written by
// downstream handlers.
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func (lrw *LoggingResponseWriter) GetBody() []byte {
	return lrw.body.Bytes()
}

func (lrw *LoggingResponseWriter) GetBodyMap() map[string]interface{} {
	responseBody := lrw.GetBody()
	if len(responseBody) == 0 {
		return nil
	}
	var responseBodyMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseBodyMap); err != nil {
		return map[string]interface{}{"raw": string(responseBody)}
	}
	return responseBodyMap
}
