package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with request details, response metadata, and duration.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody any
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				if len(bodyBytes) > 0 {
					var parsed map[string]any
					if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
						requestBody = parsed
					} else {
						requestBody = map[string]any{"raw": string(bodyBytes)}
					}
				}
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			logger.Info(r.Context(), "http_request",
				golog.Field("request", map[string]any{
					"body":    requestBody,
					"headers": r.Header,
					"method":  r.Method,
					"path":    r.URL.Path,
					"query":   r.URL.Query(),
				}),
				golog.Field("response", map[string]any{
					"headers":     ww.Header(),
					"status":      status,
					"bytes":       ww.BytesWritten(),
					"duration_ms": duration.Milliseconds(),
				}),
			)
		})
	}
}
